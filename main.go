package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"errors"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	r53 "github.com/aws/aws-sdk-go/service/route53"
)

const (
	eventSource = "aws.events"

	stateRunning = "RUNNING"

	defaultTTL      = 0
	defaultWeight   = 1
	defaultPriority = 1
)

var (
	r53Zone    = os.Getenv("ROUTE53_ZONE")
	r53ZoneID  = os.Getenv("ROUTE53_ZONE_ID")
	ecsCluster = os.Getenv("ECS_CLUSTER")
)

var (
	errNoEvent  = errors.New("no CloudWatch event")
	errNoChange = errors.New("no records to change")
)

var (
	ecsSvc = ecs.New(session.New())
	ec2Svc = ec2.New(session.New())
	r53Svc = r53.New(session.New())
)

func handler(req events.CloudWatchEvent) error {
	var err error

	if req.Source != eventSource {
		return errNoEvent
	}

	if err = registerServices(); err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return err
	}

	return err
}

func registerServices() error {
	var err error

	batchChanges := make([]*r53.Change, 0)

	serviceArns := make([]*string, 0)
	serviceArns, err = listServiceArns(serviceArns, nil)

	services := make([]*ecs.Service, 0)
	services, err = describeServices(serviceArns, services, 0)
	if err != nil {
		return err
	}

	resourceRecords := make([]*r53.ResourceRecordSet, 0)
	resourceRecords, err = listResourceRecords(resourceRecords, nil, nil)
	if err != nil {
		return err
	}

	deleteResources := make([]*r53.ResourceRecordSet, 0)

	for _, resourceRecord := range resourceRecords {
		if aws.StringValue(resourceRecord.Type) == r53.RRTypeSrv {
			shouldDelete := true
			for _, service := range services {
				dnsName := strings.Join([]string{aws.StringValue(service.ServiceName), aws.StringValue(&ecsCluster), r53Zone}, ".")
				if aws.StringValue(resourceRecord.Name) == dnsName {
					shouldDelete = false
				}
			}
			if shouldDelete {
				deleteResources = append(deleteResources, resourceRecord)
			}
		}
	}

	if len(deleteResources) > 0 {
		changes := make([]*r53.Change, 0)
		for _, deleteResource := range deleteResources {
			change := &r53.Change{
				Action:            aws.String(r53.ChangeActionDelete),
				ResourceRecordSet: deleteResource,
			}
			changes = append(changes, change)
		}

		params := &r53.ChangeResourceRecordSetsInput{
			ChangeBatch: &r53.ChangeBatch{
				Changes: changes,
			},
			HostedZoneId: aws.String(r53ZoneID),
		}

		_, err = r53Svc.ChangeResourceRecordSets(params)

		if err != nil {
			return err
		}
	}

	for _, service := range services {
		dnsName := strings.Join([]string{aws.StringValue(service.ServiceName), aws.StringValue(&ecsCluster), r53Zone}, ".")

		taskArns := make([]*string, 0)
		taskArns, err = listTasksArns(service.ServiceName, taskArns, nil)
		if err != nil {
			return err
		}

		tasks := make([]*ecs.Task, 0)
		tasks, err = describeTasks(taskArns, tasks, 0)
		if err != nil {
			return err
		}

		change, err := createSRVChangeRecord(dnsName, service.ServiceName, tasks)
		if err != nil {
			return err
		}

		batchChanges = append(batchChanges, change)
	}

	params := &r53.ChangeResourceRecordSetsInput{
		ChangeBatch: &r53.ChangeBatch{
			Changes: batchChanges,
			Comment: aws.String("ECS-Service-Discovery"),
		},
		HostedZoneId: aws.String(r53ZoneID),
	}

	_, err = r53Svc.ChangeResourceRecordSets(params)

	return err
}

func taskChange(task *ecs.Task) ([]*r53.ResourceRecord, error) {
	var err error

	containerInstances, err := describeContainerInstances(task.ClusterArn, task.ContainerInstanceArn)
	if err != nil || len(containerInstances) == 0 {
		return nil, err
	}

	containerInstance := containerInstances[0]

	instances, err := describeEc2Instances(containerInstance.Ec2InstanceId)
	if err != nil || len(instances) == 0 {
		return nil, err
	}
	instance := instances[0]
	ip := aws.StringValue(instance.PrivateIpAddress)
	changeRecords := make([]*r53.ResourceRecord, 0)

	for _, container := range task.Containers {
		for _, binding := range container.NetworkBindings {
			record := &r53.ResourceRecord{
				Value: aws.String(strings.Join([]string{strconv.Itoa(defaultPriority), strconv.Itoa(defaultWeight), strconv.FormatInt(*binding.HostPort, 10), ip}, " ")),
			}
			changeRecords = append(changeRecords, record)
		}
	}

	return changeRecords, nil
}

func createSRVChangeRecord(dnsName string, serviceName *string, tasks []*ecs.Task) (*r53.Change, error) {
	resRecords := make([]*r53.ResourceRecord, 0)

	for _, task := range tasks {
		taskRecords, err := taskChange(task)
		if err != nil {
			return nil, err
		}
		resRecords = append(resRecords, taskRecords...)
	}

	return &r53.Change{
		Action: aws.String(r53.ChangeActionUpsert),
		ResourceRecordSet: &r53.ResourceRecordSet{
			Name: aws.String(dnsName),
			// It creates an A record with the IP of the host running the agent
			Type:            aws.String(r53.RRTypeSrv),
			ResourceRecords: resRecords,
			SetIdentifier:   serviceName,
			// TTL=0 to avoid DNS caches
			TTL:    aws.Int64(defaultTTL),
			Weight: aws.Int64(defaultWeight),
			// MultiValueAnswer: aws.Bool(defaultMultiValueAnswer),
		},
	}, nil
}

func describeTasks(taskArns []*string, tasks []*ecs.Task, n int) ([]*ecs.Task, error) {
	pages := (len(tasks) / 100) - 1
	params := &ecs.DescribeTasksInput{
		Cluster: aws.String(ecsCluster),
		Tasks:   taskArns[(n * 100):int(math.Min(float64(100), float64(len(taskArns)-(n*100))))],
	}
	taskDesc, err := ecsSvc.DescribeTasks(params)
	if err != nil {
		return tasks, err
	}
	tasks = append(tasks, taskDesc.Tasks...)

	n++

	if n <= pages {
		describeTasks(taskArns, tasks, n)
	}

	return tasks, nil
}

func listTasksArns(service *string, taskArns []*string, nextToken *string) ([]*string, error) {
	var err error
	params := &ecs.ListTasksInput{
		Cluster:       aws.String(ecsCluster),
		DesiredStatus: aws.String(stateRunning),
		ServiceName:   service,
		NextToken:     nextToken,
	}

	tasks, err := ecsSvc.ListTasks(params)
	if err != nil {
		return taskArns, nil
	}
	taskArns = append(taskArns, tasks.TaskArns...)

	if tasks.NextToken != nil {
		listTasksArns(service, taskArns, nextToken)
	}

	return taskArns, nil
}

func describeServices(serviceArns []*string, services []*ecs.Service, n int) ([]*ecs.Service, error) {
	pages := (len(services) / 10) - 1
	params := &ecs.DescribeServicesInput{
		Cluster:  aws.String(ecsCluster),
		Services: serviceArns[(n * 10):int(math.Min(float64(10), float64(len(serviceArns)-(n*10))))],
	}
	serviceDesc, err := ecsSvc.DescribeServices(params)
	if err != nil {
		return services, err
	}
	services = append(services, serviceDesc.Services...)

	n++

	if n <= pages {
		describeServices(serviceArns, services, n)
	}

	return services, nil
}

func listResourceRecords(rRecords []*r53.ResourceRecordSet, startRecordName *string, startRecordType *string) ([]*r53.ResourceRecordSet, error) {
	var err error
	params := &r53.ListResourceRecordSetsInput{
		HostedZoneId:    aws.String(r53ZoneID),
		StartRecordType: startRecordType,
	}

	if startRecordName != nil && aws.StringValue(startRecordName) != "" {
		params.StartRecordName = startRecordName
	}

	if startRecordType != nil && aws.StringValue(startRecordType) != "" {
		params.StartRecordType = startRecordType
	}

	records, err := r53Svc.ListResourceRecordSets(params)
	if err != nil {
		return rRecords, err
	}

	rRecords = append(rRecords, records.ResourceRecordSets...)

	if *records.IsTruncated {
		listResourceRecords(rRecords, records.NextRecordName, records.NextRecordType)
	}

	return rRecords, nil
}

func listServiceArns(serviceArns []*string, nextToken *string) ([]*string, error) {
	var err error
	params := &ecs.ListServicesInput{
		Cluster:   aws.String(ecsCluster),
		NextToken: nextToken,
	}

	services, err := ecsSvc.ListServices(params)
	if err != nil {
		return serviceArns, nil
	}
	serviceArns = append(serviceArns, services.ServiceArns...)

	if services.NextToken != nil {
		listServiceArns(serviceArns, nextToken)
	}

	return serviceArns, nil
}

func describeEc2Instances(instance *string) ([]*ec2.Instance, error) {
	var err error
	ec2Instances := make([]*ec2.Instance, 0)
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{instance},
	}
	result, err := ec2Svc.DescribeInstances(input) // parse Reservations
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances { // loop in loop, sorry
			ec2Instances = append(ec2Instances, instance)
		}
	}

	return ec2Instances, err
}

func describeContainerInstances(clusterArn *string, instanceArn *string) ([]*ecs.ContainerInstance, error) {
	params := &ecs.DescribeContainerInstancesInput{
		Cluster:            clusterArn,
		ContainerInstances: []*string{instanceArn},
	}
	containers, err := ecsSvc.DescribeContainerInstances(params)

	return containers.ContainerInstances, err
}

func main() {
	lambda.Start(handler)
}
