package main

import (
	"os"

	"errors"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	r53 "github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/ssm"
)

const (
	defaultEnvProjectID = "PROJECT_ID"

	eventSource = "aws.events"

	stateRunning = "RUNNING"

	defaultTTL      = 0
	defaultWeight   = 1
	defaultPriority = 1
)

var (
	errNoProjectID = errors.New("no ProjectID present")
	errNoEvent     = errors.New("no CloudWatch event")
	errNoChange    = errors.New("no records to change")
)

var (
	sess   = session.New()
	ecsSvc = ecs.New(sess)
	ec2Svc = ec2.New(sess)
	r53Svc = r53.New(sess)
)

func handler(req events.CloudWatchEvent) error {
	var err error

	if req.Source != eventSource {
		return errNoEvent
	}

	projectID, ok := os.LookupEnv(defaultEnvProjectID)
	if !ok {
		return errNoProjectID
	}

	lambdaFunc := new(Func)
	lambdaFunc.ProjectID = projectID
	lambdaFunc.SSM = ssm.New(sess)

	if err = lambdaFunc.init(); err != nil {
		return err
	}

	return lambdaFunc.registerServices()
}

func main() {
	lambda.Start(handler)
}
