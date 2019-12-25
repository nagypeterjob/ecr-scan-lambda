package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/nagypeterjob/ecr-scan-lambda/internal"
)

type app struct {
	env             string
	region          string
	minimumSeverity string
	ecrService      *internal.ECRService
	slackService    *internal.SlackService
}

func (a *app) Handle(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	list, err := a.ecrService.ListRepositories(1000)
	if err != nil {
		return errorResponse(err)
	}

	filtered, scanErrors := a.ecrService.GetFilteredFindings(list, a.region, a.minimumSeverity)

	headerMsg := fmt.Sprintf("*Scan results on %s*", time.Now().Format("2006 Jan 02"))
	err = a.slackService.PostStandaloneMessage(headerMsg)
	if err != nil {
		return errorResponse(err)
	}

	for _, r := range filtered {
		blockParts := a.slackService.BuildMessageBlock(r)

		channelID, timestamp, err := a.slackService.PostMessage(blockParts...)
		if err != nil {
			return errorResponse(err)
		}
		fmt.Printf("Message successfully sent to channel %s at %s\n", channelID, timestamp)
	}

	if len(scanErrors) != 0 {
		errorMsg := fmt.Sprintf(":x: *Failed to get scan results from the following repos:* :x:")
		err = a.slackService.PostStandaloneMessage(errorMsg)
		if err != nil {
			return errorResponse(err)
		}

		var failedRepos string
		for _, failed := range scanErrors {
			failedRepos += failed.RepositoryName + "\n"
		}
		err = a.slackService.PostStandaloneMessage(failedRepos)
		if err != nil {
			return errorResponse(err)
		}
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}
}

func errorResponse(err error) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}
}

// Handler glues the lambda logic together
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	err := internal.PrintVersion()
	if err != nil {
		return errorResponse(err), nil
	}

	config := internal.InitConfig()
	sess, err := session.NewSession(&aws.Config{Region: &config.Region})
	if err != nil {
		return errorResponse(err), nil
	}

	app := app{
		env:             config.Env,
		region:          config.Region,
		minimumSeverity: config.MinimumSeverity,
		ecrService:      internal.NewECRService(config.EcrID, ecr.New(sess)),
		slackService: internal.NewSlackService(
			config.SlackToken,
			config.SlackChannel,
			config.EmojiMap,
		),
	}
	return app.Handle(request), nil
}

func main() {
	lambda.Start(Handler)
}
