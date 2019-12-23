package internal

import (
	"fmt"
	"os"
)

type Config struct {
	Region          string
	MinimumSeverity string
	Env             string
	EcrID           string
	SlackToken      string
	SlackChannel    string
	EmojiMap        map[string]string
}

func populateEmojiValue(key string, fallback string) string {
	value := os.Getenv(fmt.Sprintf("EMOJI_%s", key))
	if len(value) == 0 {
		return fallback
	}
	return value
}

func InitConfig() Config {
	minSev := os.Getenv("MINIMUM_SEVERITY")
	if len(minSev) == 0 {
		minSev = "HIGH"
	}

	emojiMap := map[string]string{}
	defaultEmojis := []string{
		":no_entry:",
		":warning:",
		":pill:",
		":rain_cloud:",
		":information_source:",
		":question:",
	}

	for i, key := range SeverityList {
		emojiMap[key] = populateEmojiValue(key, defaultEmojis[i])
	}

	return Config{
		Region:          os.Getenv("AWS_REGION"),
		MinimumSeverity: minSev,
		Env:             os.Getenv("ENV"),
		EcrID:           os.Getenv("ECR_ID"),
		SlackToken:      os.Getenv("SLACK_TOKEN"),
		SlackChannel:    os.Getenv("SLACK_CHANNEL"),
		EmojiMap:        emojiMap,
	}
}
