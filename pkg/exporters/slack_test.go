package exporters

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/nagypeterjob/ecr-scan-lambda/pkg/api"
	"github.com/nagypeterjob/ecr-scan-lambda/pkg/severity"
	"github.com/nlopes/slack"
)

var slackService = NewSlackExporter("slack", "", "")

func blockToJSON(block interface{}) (string, error) {
	b, err := json.MarshalIndent(block, "", "    ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func TestGenerateTextBlock(t *testing.T) {
	cases := []struct {
		block    slack.Block
		expected string
	}{
		{
			block: slackService.GenerateTextBlock("first test"),
			expected: `{
    "type": "section",
    "text": {
        "type": "mrkdwn",
        "text": "first test"
    }
}`,
		},
		{
			block: slackService.GenerateTextBlock(":fire: second test"),
			expected: `{
    "type": "section",
    "text": {
        "type": "mrkdwn",
        "text": ":fire: second test"
    }
}`,
		},
	}
	for i, c := range cases {
		parsed, err := blockToJSON(c.block)
		if err != nil {
			t.Fatalf("TestGenerateTextBlock json parse has failed")
		}
		if parsed != c.expected {
			t.Fatalf("[%d] values are not equal, wanting: %s, got: %s", i, c.expected, parsed)
		}
	}
}

func TestBuildMessageBlock(t *testing.T) {
	cases := []struct {
		repository *api.RepositoryInfo
		expected   int
	}{
		{
			repository: &api.RepositoryInfo{
				Name: "TestRepository",
				Severity: severity.Matrix{
					Count: map[string]*int64{
						"CRITICAL":      aws.Int64(1),
						"HIGH":          aws.Int64(2),
						"MEDIUM":        aws.Int64(3),
						"LOW":           aws.Int64(4),
						"INFORMATIONAL": aws.Int64(5),
						"UNDEFINED":     aws.Int64(6),
					},
				},
			},
			expected: 4,
		},
	}

	for i, c := range cases {
		blocks := slackService.BuildMessageBlock(c.repository)

		if len(blocks) != c.expected {
			t.Fatalf("[%d] values are not equal, wanting: %d, got: %d", i, c.expected, len(blocks))
		}
	}
}

func TestBlockMessage(t *testing.T) {
	cases := []struct {
		repository *api.RepositoryInfo
		expected   string
	}{
		{
			repository: &api.RepositoryInfo{
				Name: "TestRepository/TestRepo1",
				Link: "https://console.aws.amazon.com/ecr/repositories/TestRepo/Test1/image/xxxyyyzzzddd/scan-results?region=us-east-1",
				Severity: severity.Matrix{
					Count: map[string]*int64{
						"CRITICAL":      aws.Int64(1),
						"HIGH":          aws.Int64(2),
						"MEDIUM":        aws.Int64(3),
						"LOW":           aws.Int64(4),
						"INFORMATIONAL": aws.Int64(5),
						"UNDEFINED":     aws.Int64(6),
					},
				},
			},
			expected: `{
    "replace_original": false,
    "delete_original": false,
    "blocks": [
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "Vulnerabilities found in *TestRepository/TestRepo1*:"
            }
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "CRITICAL *1*\nHIGH *2*\nMEDIUM *3*\nLOW *4*\nINFORMATIONAL *5*\nUNDEFINED *6*\n"
            }
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "View detailed scan results \u003chttps://console.aws.amazon.com/ecr/repositories/TestRepo/Test1/image/xxxyyyzzzddd/scan-results?region=us-east-1| on ECR console\u003e"
            }
        },
        {
            "type": "divider"
        }
    ]
}`,
		},
	}
	for i, c := range cases {
		blocks := slackService.BuildMessageBlock(c.repository)
		block := slackService.BlockMessage(blocks...)

		parsed, err := blockToJSON(block)
		if err != nil {
			t.Fatalf("TestBlockMessage json parse has failed")
		}
		if parsed != c.expected {
			t.Fatalf("[%d] values are not equal, wanting: %s, got: %s", i, c.expected, parsed)
		}
	}
}
