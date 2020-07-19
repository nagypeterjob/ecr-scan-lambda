package main

import (
	"fmt"
	"os"
)

// Config stores lambda configuration
type config struct {
	region          string
	minimumSeverity string
	env             string
	ecrID           string
	imageTag        string
	exporters       string
	logLevel        string
	numWorkers      string

	slack   slackConfig
	sns     snsConfig
	mailgun mailgunConfig
}

type slackConfig struct {
	token   string
	channel string
}

type snsConfig struct {
	topicARN string
}

type mailgunConfig struct {
	apiKey     string
	from       string
	recipients string
}

func retrive(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}

func required(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return value, fmt.Errorf("Required environtment variable %s is not set", key)
	}
	return value, nil
}

// initConfig populates lambda configuration
func initConfig() (config, error) {

	env, err := required("ENV")
	if err != nil {
		return config{}, err
	}

	region, err := required("REGION")
	if err != nil {
		return config{}, err
	}

	return config{
		env:             env,
		region:          region,
		ecrID:           retrive("ECR_ID", ""),
		exporters:       retrive("EXPORTERS", "log"),
		imageTag:        retrive("IMAGE_TAG", "latest"),
		logLevel:        retrive("LOG_LEVEL", "INFO"),
		numWorkers:      retrive("NUM_WORKERS", "10"),
		minimumSeverity: retrive("MINIMUM_SEVERITY", "CRITICAL"),
		mailgun: mailgunConfig{
			apiKey:     retrive("MAILGUN_API_KEY", ""),
			from:       retrive("MAILGUN_FROM", ""),
			recipients: retrive("MAILGUN_RECIPIENTS", ""),
		},
		slack: slackConfig{
			token:   retrive("SLACK_TOKEN", ""),
			channel: retrive("SLACK_CHANNEL", ""),
		},

		sns: snsConfig{
			topicARN: retrive("SNS_TOPIC_ARN", ""),
		},
	}, nil
}
