package main

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestHandler(t *testing.T) {
	// Setup Test
	mockEvent := &events.CloudWatchEvent{}

	handler(*mockEvent)
}
