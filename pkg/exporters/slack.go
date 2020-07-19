package exporters

import (
	"bytes"
	"fmt"
	"time"

	"github.com/nagypeterjob/ecr-scan-lambda/pkg/api"
	"github.com/nagypeterjob/ecr-scan-lambda/pkg/severity"
	"github.com/nlopes/slack"
)

// SlackService data structure for storing slack client related data
type SlackService struct {
	client  *slack.Client
	channel string
	name    string
}

// NewSlackExporter populates a new SlackService instance
func NewSlackExporter(name string, token string, channel string) *SlackService {
	return &SlackService{
		client:  slack.New(token),
		channel: channel,
		name:    name,
	}
}

// Name .
func (s SlackService) Name() string {
	return s.name
}

// Format clousure formats scan results and returns a function that sends report on invocation
func (s SlackService) Format(filtered []*api.RepositoryInfo, failed []*api.RepositoryInfo) (func() error, error) {
	var failedMsgBuffer bytes.Buffer
	if len(failed) > 0 {
		failedMsgBuffer.WriteString(boldn(reportFailedHeadText))

		for _, f := range failed {
			failedMsgBuffer.WriteString(f.Name + "\n")
		}
	}

	failedMsg := failedMsgBuffer.String()

	// Send publishes message to provided slack channel
	return func() error {
		err := s.PostStandaloneMessage(bold(reportHeadText))
		if err != nil {
			return err
		}

		if len(filtered) == 0 {
			return s.PostStandaloneMessage(reportClean)
		}

		for _, r := range filtered {
			blockParts := s.BuildMessageBlock(r)
			channelID, timestamp, err := s.PostMessage(blockParts...)
			if err != nil {
				return err
			}
			fmt.Printf("Message successfully sent to channel %s at %s\n", channelID, timestamp)
		}

		if len(failedMsg) != 0 {
			err := s.PostStandaloneMessage(failedMsg)
			if err != nil {
				return err
			}
		}

		return nil
	}, nil
}

// BuildMessageBlock constructs severity related message body
func (s *SlackService) BuildMessageBlock(r *api.RepositoryInfo) []slack.Block {
	headerSection := s.GenerateTextBlock(fmt.Sprintf("Vulnerabilities found in *%s*:", r.Name))
	linkSection := s.GenerateTextBlock(fmt.Sprintf("View detailed scan results <%s| on ECR console>", r.Link))

	var buffer bytes.Buffer
	for _, key := range severity.SeverityList {
		if val, ok := r.Severity.Count[key]; ok {
			buffer.WriteString(fmt.Sprintf("%s *%d*\n", key, *val))
		}
	}
	severitySection := s.GenerateTextBlock(buffer.String())

	return []slack.Block{
		headerSection,
		severitySection,
		linkSection,
		slack.NewDividerBlock(),
	}
}

// GenerateTextBlock returns a slack SectionBlock for text input
func (s *SlackService) GenerateTextBlock(text string) slack.Block {
	textBlock := slack.NewTextBlockObject("mrkdwn", text, false, false)
	return slack.NewSectionBlock(textBlock, nil, nil)
}

// BlockMessage returns a slack BlockMessage for multiple Block inputs
func (s *SlackService) BlockMessage(blocks ...slack.Block) slack.Message {
	return slack.NewBlockMessage(blocks...)
}

// PostMessage sends provided slack MessageBlocks to the given slack channel
func (s *SlackService) PostMessage(blocks ...slack.Block) (string, string, error) {
	// Wait one second so posting doesn't exceed Slack's rate limit
	time.Sleep(1 * time.Second)
	return s.client.PostMessage(s.channel, slack.MsgOptionBlocks(blocks...))
}

// PostStandaloneMessage generates slack SectionBlock for provided text and sends it to the given slack channel
func (s *SlackService) PostStandaloneMessage(message string) error {
	blockParts := []slack.Block{s.GenerateTextBlock(message)}
	_, _, err := s.PostMessage(blockParts...)
	return err
}

func bold(message string) string {
	return fmt.Sprintf("*%s*", message)
}

// bold formatting with new line in the end
func boldn(message string) string {
	return fmt.Sprintf("*%s*\n", message)
}
