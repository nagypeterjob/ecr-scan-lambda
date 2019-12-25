package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/nagypeterjob/ecr-scan-lambda/internal"
)

type app struct {
	env           string
	region        string
	ecrRegistryID string
	ecrService    *internal.ECRService
}

func errorResponse(err error) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}
}

func (a *app) Handle(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	list, err := a.ecrService.ListRepositories(1000)
	if err != nil {
		return errorResponse(err)
	}
	for _, repo := range list.Repositories {
		_, err = a.ecrService.PutImageScanningConfiguration(repo)
		if err != nil {
			fmt.Println(fmt.Sprintf("Could't set image scaning configuration for repository: %s, error: %s", *repo.RepositoryName, err.Error()))
		}

		_, err := a.ecrService.StartImageScan(repo.RepositoryName)
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
	app := app{
		env:           os.Getenv("ENV"),
		region:        region,
		ecrRegistryID: os.Getenv("ECR_ID"),
		ecrService:    internal.NewECRService(region, ecr.New(sess)),
	}
	return app.Handle(request), nil
}

func main() {
	lambda.Start(Handler)
}
