package api

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
)

type mockSNSService struct {
	snsiface.SNSAPI
}

func (s mockSNSService) Publish(input *sns.PublishInput) (*sns.PublishOutput, error) {
	return &sns.PublishOutput{
		MessageId: aws.String("uniqueId"),
	}, nil
}

func TestPublish(t *testing.T) {
	svc := mockSNSService{}

	output, err := svc.Publish(&sns.PublishInput{
		MessageStructure: aws.String("json"),
		Message:          aws.String("{ \"default\": \"empty\"  }"),
		TopicArn:         aws.String("dummy:arm:1234/topic"),
	})

	if err != nil {
		t.Fatalf("Error publishing to topic: %s", err.Error())
	}
	if output.MessageId == nil {
		t.Fatalf("MessageID is empty: %s", err.Error())
	}

}
