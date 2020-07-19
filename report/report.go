package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/sns"
	api "github.com/nagypeterjob/ecr-scan-lambda/pkg/api"
	exp "github.com/nagypeterjob/ecr-scan-lambda/pkg/exporters"
	"github.com/nagypeterjob/ecr-scan-lambda/pkg/logger"
)

type app struct {
	api             *api.ECRService
	env             string
	exporters       []exp.Exporter
	logger          *logger.Logger
	minimumSeverity string
	numWorkers      int
	region          string
}

func initExporters(config config, logger *logger.Logger) ([]exp.Exporter, error) {
	var exporters []exp.Exporter

	logger.Infof("Exporters enabled: %s", config.exporters)

	enabledExporters := strings.Split(config.exporters, ",")

	for _, e := range enabledExporters {

		if e == "log" {
			logger.Debug("Initializing log exporter...")
			logexp := exp.NewLogExporter(e)
			exporters = append(exporters, logexp)
		}

		if e == "slack" {
			logger.Debug("Initializing slack exporter...")
			slack := exp.NewSlackExporter(e, config.slack.token, config.slack.channel)
			exporters = append(exporters, slack)
		}

		if e == "sns" {
			logger.Debug("Initializing sns exporter...")
			sess, err := session.NewSession(&aws.Config{Region: &config.region})
			if err != nil {
				return nil, err
			}
			client := sns.New(sess)

			service := api.NewSNSService(client)

			sns := exp.NewSNSExporter(e, service, config.sns.topicARN)
			exporters = append(exporters, sns)
		}

		if e == "mailgun" {
			logger.Debug("Initializing Mailgun exporter...")
			mg := exp.NewMailgunExporter(e, config.mailgun.recipients, config.mailgun.from, config.mailgun.apiKey)
			exporters = append(exporters, mg)
		}
	}
	return exporters, nil
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

	// Scan repositories then filter them based on provided severity level
	filtered, failed := a.api.GatherVulnerabilities(ctx, repositories, a.minimumSeverity, a.numWorkers)

	// Format and send vulnerability reports to each enabled exporters
	for _, e := range a.exporters {
		send, err := e.Format(filtered, failed)
		if err != nil {
			return errorResponse(err)
		}

		if err = send(); err != nil {
			return errorResponse(err)
		}

		a.logger.Infof("%s exporter has sucessfully sent message", e.Name())
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}
}

func errorResponse(err error) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}
}

// Handler glues the lambda logic together
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	err := printVersion()
	if err != nil {
		return errorResponse(err), err
	}

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

	exporters, err := initExporters(config, logger)
	if err != nil {
		return errorResponse(err), err
	}

	app := app{
		api:             api.NewECRService(config.ecrID, config.region, config.imageTag, logger, ecr.New(sess)),
		env:             config.env,
		exporters:       exporters,
		logger:          logger,
		minimumSeverity: config.minimumSeverity,
		numWorkers:      int(nw),
		region:          config.region,
	}
	return app.Handle(request), nil
}

func main() {
	lambda.Start(Handler)
}
