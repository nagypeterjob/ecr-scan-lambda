package internal

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/nlopes/slack"
)

var slackService = NewSlackService("", "", map[string]string{
	"CRITICAL":      ":no_entry:",
	"HIGH":          ":warning:",
	"MEDIUM":        ":pill:",
	"LOW":           ":rain_cloud:",
	"INFORMATIONAL": ":information_source:",
	"UNDEFINED":     ":question:",
})

func blockToJson(block interface{}) (string, error) {
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
		parsed, err := blockToJson(c.block)
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
		repository Repository
		expected   int
	}{
		{
			repository: Repository{
				Name: "TestRepository",
				Severity: Severity{
					Count: map[string]*int64{
						"CRITICAL":      aws.Int64(1),
						"HIGH":          aws.Int64(2),
						"MEDIUM":        aws.Int64(3),
						"LOW":           aws.Int64(4),
						"INFORMATIONAL": aws.Int64(5),
						"UNDEFINED":     aws.Int64(6),
					},
					Link: "https://console.aws.amazon.com/ecr/repositories/testrepository/image/sha256:262000a32cb2e7cbd397f7bc0feb2c495f2f5fc59d6a06ca3078eec3dcd4084b/scan-results?region=us-region-1",
				},
			},
			expected: 9,
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
		repository Repository
		expected   string
	}{
		{
			repository: Repository{
				Name: "TestRepository",
				Severity: Severity{
					Count: map[string]*int64{
						"CRITICAL":      aws.Int64(1),
						"HIGH":          aws.Int64(2),
						"MEDIUM":        aws.Int64(3),
						"LOW":           aws.Int64(4),
						"INFORMATIONAL": aws.Int64(5),
						"UNDEFINED":     aws.Int64(6),
					},
					Link: "https://console.aws.amazon.com/ecr/repositories/testrepository/image/sha256:262000a32cb2e7cbd397f7bc0feb2c495f2f5fc59d6a06ca3078eec3dcd4084b/scan-results?region=us-region-1",
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
                "text": "Vulnerabilities found in *TestRepository*:"
            }
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": ":no_entry: CRITICAL *1* "
            }
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": ":warning: HIGH *2* "
            }
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": ":pill: MEDIUM *3* "
            }
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": ":rain_cloud: LOW *4* "
            }
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": ":information_source: INFORMATIONAL *5* "
            }
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": ":question: UNDEFINED *6* "
            }
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "View detailed scan results \u003chttps://console.aws.amazon.com/ecr/repositories/testrepository/image/sha256:262000a32cb2e7cbd397f7bc0feb2c495f2f5fc59d6a06ca3078eec3dcd4084b/scan-results?region=us-region-1| on ECR console\u003e"
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
		msg := slackService.BlockMessage(blocks...)

		parsed, err := blockToJson(msg)
		if err != nil {
			t.Fatalf("TestBlockMessage json parse has failed")
		}
		if parsed != c.expected {
			t.Fatalf("[%d] values are not equal, wanting: %s, got: %s", i, c.expected, parsed)
		}
	}
}
