package main

import (
	"fmt"
	"os"
)

// Config stores lambda configuration
type config struct {
	env        string
	ecrID      string
	imageTag   string
	logLevel   string
	numWorkers string
	region     string
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
		env:        env,
		region:     region,
		ecrID:      retrive("ECR_ID", ""),
		imageTag:   retrive("IMAGE_TAG", "latest"),
		logLevel:   retrive("LOG_LEVEL", "INFO"),
		numWorkers: retrive("NUM_WORKERS", "2"),
	}, nil
}
