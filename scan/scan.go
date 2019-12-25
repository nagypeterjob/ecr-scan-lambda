package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
)

type app struct {
	env           string
	region        string
	ecrRegistryID string
	ecrService    ecriface.ECRAPI
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

func errorResponse(err error) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}
}

func (a *app) Handle(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	list, err := a.listRepositories(1000)
	if err != nil {
		return errorResponse(err)
	}
	for _, repo := range list.Repositories {
		scanConfigInput := ecr.PutImageScanningConfigurationInput{
			RepositoryName: repo.RepositoryName,
			ImageScanningConfiguration: &ecr.ImageScanningConfiguration{
				ScanOnPush: aws.Bool(true),
			},
		}
		if len(a.ecrRegistryID) != 0 {
			scanConfigInput.RegistryId = aws.String(a.ecrRegistryID)
		}

		_, err := a.ecrService.PutImageScanningConfiguration(&scanConfigInput)
		if err != nil {
			fmt.Println(fmt.Sprintf("Could't set image scaning configuration for repository: %s, error: %s", *repo.RepositoryName, err.Error()))
		}
		startImageScanInput := ecr.StartImageScanInput{
			ImageId: &ecr.ImageIdentifier{
				ImageTag: aws.String("latest"),
			},
			RepositoryName: repo.RepositoryName,
		}
		_, err = a.ecrService.StartImageScan(&startImageScanInput)
		if err != nil {
			fmt.Println(fmt.Sprintf("Image scan today was already done for repository: %s, error: %s", *repo.RepositoryName, err.Error()))
		}
	}
	return events.APIGatewayProxyResponse{}
}

// Handler glues the lambda logic together
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	region := os.Getenv("REGION")
	sess, err := session.NewSession(&aws.Config{Region: &region})
	if err != nil {
		return errorResponse(err), nil
	}
	svc := ecr.New(sess)
	app := app{
		env:           os.Getenv("ENV"),
		region:        region,
		ecrRegistryID: os.Getenv("ECR_ID"),
		ecrService:    svc,
	}
	return app.Handle(request), nil
}

func main() {
	lambda.Start(Handler)
}
