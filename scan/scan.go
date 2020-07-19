package main

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	api "github.com/nagypeterjob/ecr-scan-lambda/pkg/api"
	"github.com/nagypeterjob/ecr-scan-lambda/pkg/logger"
)

type app struct {
	api        *api.ECRService
	env        string
	imageTag   string
	logger     *logger.Logger
	numWorkers int
	region     string
}

func errorResponse(err error) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}
}

func (a *app) StartScan(in chan *api.ScanningResult, cancelFunc context.CancelFunc) chan error {
	errc := make(chan error, 1)
	var wg sync.WaitGroup
	for i := 0; i < a.numWorkers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for r := range in {
				if r.Err != nil {
					cancelFunc()
					errc <- r.Err
					return
				}

				_, err := a.api.StartImageScan(r.Output.RepositoryName)
				a.logger.Infof("StartImageScan - (goroutine #%d) \n", i)
				if err != nil {
					if aerr, ok := err.(awserr.Error); ok {
						switch aerr.Code() {
						case ecr.ErrCodeLimitExceededException:
							a.logger.Errorf("Todays scan limit exceeded for %s repository.\n", *r.Output.RepositoryName)
						case ecr.ErrCodeImageNotFoundException:
							a.logger.Errorf("Image with tag %s not found in repository %s.\n", a.imageTag, *r.Output.RepositoryName)
						default:
							a.logger.Errorf("Error when scanning repository: %s, error: %s \n", *r.Output.RepositoryName, err.Error())
							cancelFunc()
							return
						}
					}
				}
			}
		}(i)
	}
	wg.Wait()
	close(errc)
	return errc
}

func (a *app) Handle(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	// Load all ecr repositories into a channel
	repositories, describeError := a.api.DescribeRepositoriesPages(ctx)

	if err := <-describeError; err != nil {
		cancelFunc()
		fmt.Println(err)
	}

	// Enable image scanning ability on repositories
	scanningResult := a.api.GenImageScanningConfiguration(ctx, repositories, a.numWorkers)

	// Scan image with provided tag for each repositories
	scanError := a.StartScan(scanningResult, cancelFunc)

	if err := <-scanError; err != nil {
		cancelFunc()
		fmt.Println(err)
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}
}

// Handler glues the lambda logic together
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	config, err := initConfig()
	if err != nil {
		return errorResponse(err), err
	}

	sess, err := session.NewSession(&aws.Config{Region: &config.region})
	if err != nil {
		return errorResponse(err), err
	}

	nw, err := strconv.ParseInt(config.numWorkers, 10, 64)
	if err != nil {
		return errorResponse(err), err
	}

	logger, err := logger.NewLogger(config.logLevel)
	if err != nil {
		return errorResponse(err), err
	}

	app := app{
		api:        api.NewECRService(config.ecrID, config.region, config.imageTag, logger, ecr.New(sess)),
		env:        config.env,
		imageTag:   config.imageTag,
		logger:     logger,
		numWorkers: int(nw),
		region:     config.region,
	}
	return app.Handle(request), nil
}

func main() {
	lambda.Start(Handler)
}
