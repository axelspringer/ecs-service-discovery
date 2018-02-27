package main

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	r53 "github.com/aws/aws-sdk-go/service/route53"
)

type srvRecord struct {
	ClusterName *string
	ServiceName string
	Container   *ecs.Container
	Zone        *r53.HostedZone
	IP          string
	Priority    int
	Weight      int
	R53Zone     string
	VpcID       string
}

func (r *srvRecord) name() string {
	return strings.Join([]string{r.ServiceName, aws.StringValue(r.ClusterName), r.R53Zone}, ".")
}

func (r *srvRecord) value(hostPort string) string {
	return strings.Join([]string{strconv.Itoa(r.Priority), strconv.Itoa(r.Weight), hostPort, r.IP}, " ")
}
