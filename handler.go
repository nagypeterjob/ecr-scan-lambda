package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"github.com/nagypeterjob/ecr-scan-lambda/internal"
)

type app struct {
	env             string
	region          string
	minimumSeverity string
	ecrRegistryID   string
	ecrService      ecriface.ECRAPI
	slackService    *internal.SlackService
}

func (a *app) listRepositories(maxRepos int) (*ecr.DescribeRepositoriesOutput, error) {
	mr := int64(maxRepos)
	input := ecr.DescribeRepositoriesInput{
		MaxResults: &mr,
	}

	if len(a.ecrRegistryID) != 0 {
		input.RegistryId = aws.String(a.ecrRegistryID)
	}

	return a.ecrService.DescribeRepositories(&input)
}

func (a *app) describeImageScanFindings(input *ecr.DescribeImageScanFindingsInput) (*ecr.DescribeImageScanFindingsOutput, error) {
	return a.ecrService.DescribeImageScanFindings(input)
}

func (a *app) GetFindings(r *ecr.DescribeRepositoriesOutput) ([]ecr.DescribeImageScanFindingsOutput, []internal.ScanErrors) {
	var findings []ecr.DescribeImageScanFindingsOutput
	var failed []internal.ScanErrors

	for _, repo := range r.Repositories {
		describeInput := ecr.DescribeImageScanFindingsInput{
			ImageId: &ecr.ImageIdentifier{
				ImageTag: aws.String("latest"),
			},
			RepositoryName: repo.RepositoryName,
		}

		if len(a.ecrRegistryID) != 0 {
			describeInput.RegistryId = aws.String(a.ecrRegistryID)
		}

		finding, err := a.describeImageScanFindings(&describeInput)
		if err != nil {
			failed = append(failed, internal.ScanErrors{RepositoryName: *repo.RepositoryName})
		} else {
			findings = append(findings, *finding)
		}
	}
	return findings, failed
}

func (a *app) filterBySeverity(findings []ecr.DescribeImageScanFindingsOutput) []internal.Repository {
	var filtered []internal.Repository
	for _, finding := range findings {
		if finding.ImageScanFindings != nil && len(finding.ImageScanFindings.FindingSeverityCounts) != 0 {
			r := internal.Repository{
				Name: *finding.RepositoryName,
				Severity: internal.Severity{
					Count: finding.ImageScanFindings.FindingSeverityCounts,
					Link:  fmt.Sprintf("https://console.aws.amazon.com/ecr/repositories/%s/image/%s/scan-results?region=%s", *finding.RepositoryName, *finding.ImageId.ImageDigest, a.region),
				},
			}
			if r.Severity.CalculateScore() >= internal.SeverityTable[a.minimumSeverity] {
				filtered = append(filtered, r)
			}
		}
	}
	return filtered
}

func (a *app) Handle(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	list, err := a.listRepositories(1000)
	if err != nil {
		return errorResponse(err)
	}

	findings, scanErrors := a.GetFindings(list)

	filtered := a.filterBySeverity(findings)

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
		errorMsg := fmt.Sprintf(":x: *Failed get scan results from the following repos:* :x:")
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

// Handler is glues the lambda logic together
func Handler(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	err := internal.PrintVersion()
	if err != nil {
		return errorResponse(err)
	}

	config := internal.InitConfig()
	sess, err := session.NewSession(&aws.Config{Region: &config.Region})
	if err != nil {
		return errorResponse(err)
	}
	svc := ecr.New(sess)

	app := app{
		env:             config.Env,
		region:          config.Region,
		minimumSeverity: config.MinimumSeverity,
		ecrService:      svc,
		ecrRegistryID:   config.EcrID,
		slackService: internal.NewSlackService(
			config.SlackToken,
			config.SlackChannel,
			config.EmojiMap,
		),
	}
	return app.Handle(request)
}

func main() {
	lambda.Start(Handler)
}
