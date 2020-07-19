package exporters

import (
	"encoding/json"
	"strconv"

	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/nagypeterjob/ecr-scan-lambda/pkg/api"
	"github.com/nagypeterjob/ecr-scan-lambda/pkg/severity"
)

// SNSExporter publishes message to SNS topic as json
type SNSExporter struct {
	client   *api.SNSService
	name     string
	topicARN string
}

type jsonData struct {
	Head           string       `json:"head"`
	Vulnerablities []repository `json:"vulnerablities"`
	Failed         []string     `json:"failed"`
	Default        string       `json:"default"`
}

type repository struct {
	Name     string         `json:"name"`
	Link     string         `json:"link"`
	Findings []vulnerablity `json:"findings"`
}

type vulnerablity struct {
	Severity string `json:"severity"`
	Count    string `json:"count"`
}

// NewSNSExporter .
func NewSNSExporter(name string, client *api.SNSService, topicARN string) *SNSExporter {
	return &SNSExporter{
		client:   client,
		name:     name,
		topicARN: topicARN,
	}
}

// Name .
func (s SNSExporter) Name() string {
	return s.name
}

// Format clousure formats scan results and returns a function that sends report on invocation
func (s SNSExporter) Format(filtered []*api.RepositoryInfo, failed []*api.RepositoryInfo) (func() error, error) {
	js := jsonData{
		Head:           reportHeadText,
		Vulnerablities: s.format(filtered),
		Failed:         s.formatFailed(failed),
	}

	bytes, err := marshal(js)
	if err != nil {
		return nil, err
	}

	msg := string(bytes)

	return func() error {
		input := sns.PublishInput{
			Message:  &msg,
			TopicArn: &s.topicARN,
		}

		if _, err = s.client.Publish(&input); err != nil {
			return err
		}

		return nil
	}, nil
}

func marshal(data jsonData) ([]byte, error) {
	return json.Marshal(data)
}

func (s SNSExporter) format(repositories []*api.RepositoryInfo) []repository {
	var ret []repository
	for _, r := range repositories {
		repo := repository{
			Name: r.Name,
			Link: r.Link,
		}

		for _, key := range severity.SeverityList {
			if val, ok := r.Severity.Count[key]; ok {
				repo.Findings = append(repo.Findings, vulnerablity{
					Severity: key,
					Count:    strconv.FormatInt(*val, 10),
				})
			}
		}
		ret = append(ret, repo)
	}
	return ret
}

func (s SNSExporter) formatFailed(repositories []*api.RepositoryInfo) []string {
	var ret []string
	for _, r := range repositories {
		ret = append(ret, r.Name)
	}
	return ret
}
