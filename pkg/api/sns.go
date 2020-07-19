package api

import (
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
)

// SNSService implements SNS API
type SNSService struct {
	client snsiface.SNSAPI
}

// NewSNSService .
func NewSNSService(client snsiface.SNSAPI) *SNSService {
	return &SNSService{
		client: client,
	}
}

// Publish passes PublishInput to upstream AWS SDK SNS Publish
func (s *SNSService) Publish(input *sns.PublishInput) (*sns.PublishOutput, error) {
	return s.client.Publish(input)
}
