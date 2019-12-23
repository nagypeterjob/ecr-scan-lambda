package internal

import (
	"fmt"

	"github.com/nlopes/slack"
)

type SlackService struct {
	Client   *slack.Client
	Channel  string
	EmojiMap map[string]string
}

func NewSlackService(token string, channel string, emojiMap map[string]string) *SlackService {
	return &SlackService{
		Client:   slack.New(token),
		EmojiMap: emojiMap,
		Channel:  channel,
	}
}

func (svc *SlackService) BuildMessageBlock(r Repository) []slack.Block {
	headerText := fmt.Sprintf("Vulnerabilities found in *%s*:", r.Name)
	headerSection := svc.GenerateTextBlock(headerText)

	linkText := fmt.Sprintf("View detailed scan results <%s| on ECR console>", r.Severity.Link)
	linkSection := svc.GenerateTextBlock(linkText)

	sevCount := r.Severity.Count

	result := []slack.Block{
		headerSection,
	}

	for _, key := range SeverityList {
		if val, ok := sevCount[key]; ok {
			msg := fmt.Sprintf("%s %s *%d* ", svc.EmojiMap[key], key, *val)
			result = append(result, svc.GenerateTextBlock(msg))
		}
	}

	result = append(result, linkSection)
	result = append(result, slack.NewDividerBlock())
	return result
}

func (svc *SlackService) GenerateTextBlock(text string) slack.Block {
	textBlock := slack.NewTextBlockObject("mrkdwn", text, false, false)
	return slack.NewSectionBlock(textBlock, nil, nil)
}

func (svc *SlackService) BlockMessage(blocks ...slack.Block) slack.Message {
	return slack.NewBlockMessage(blocks...)
}

func (svc *SlackService) PostMessage(blocks ...slack.Block) (string, string, error) {
	return svc.Client.PostMessage(svc.Channel, slack.MsgOptionBlocks(blocks...))
}

func (svc *SlackService) PostStandaloneMessage(message string) error {
	blockParts := []slack.Block{svc.GenerateTextBlock(message)}
	_, _, err := svc.PostMessage(blockParts...)
	return err
}
