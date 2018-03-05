package main

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

// func (m *mockEscClient) DescribeContainerInstances(input *ecs.DescribeContainerInstancesInput) (*ecs.DescribeContainerInstancesOutput, error) {
// 	output := &ecs.DescribeContainerInstancesOutput{
// 		ContainerInstances: []*ecs.ContainerInstance{},
// 	}

// 	return output, nil
// }

func TestHandler(t *testing.T) {
	// Setup Test
	mockEvent := &events.CloudWatchEvent{}

	handler(*mockEvent)
}
