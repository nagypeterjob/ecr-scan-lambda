# ecr-scan-lambda
Lambdas which trigger ECR vulnerability scan and sends reports to multiple selected destinations

[![Go Report Card](https://goreportcard.com/badge/github.com/nagypeterjob/ecr-scan-lambda)](https://goreportcard.com/report/github.com/nagypeterjob/ecr-scan-lambda)
![](https://github.com/nagypeterjob/ecr-scan-lambda/workflows/Go%20tests/badge.svg?branch=master)


## Changelog
[Read Changelog!](https://github.com/nagypeterjob/ecr-scan-lambda/blob/master/CHANGELOG.md)

## Getting started

The serverless deployment has two AWS Lambda functions:
- `ecr-scan-lambda` for enabling *ScanOnPush* parameter on each repository and triggering scans (There is a one scan / image / day limit by AWS)
- `ecr-report-lambda` for sending cumulated vulnerablity report to selected destinations

Both functions are triggered by Cloudwatch events. Can be configured via **serverless.yml**

### Prerequisites
1. It is considered to be a best practice to push a container image to a repository with multiple tags. Tags could be:
    1. The semantic version of the release, or a commit hash (use this to deploy your application)
    2. A "static" tag which always points to the latest image e.g.: `latest` (use this for vulnerability scans)
2. It is recommended to set the `IMAGE_TAG` environment variable to your "static" tag. 
3. See the list of [available](#environment-variables) environment variables for the functions
4. Install [Serverless framework](https://www.serverless.com/framework/docs/getting-started/) on your local machine
5. Use **serverless.yml** to deploy functions to your AWS environment (or integrate it to your CI/CD pipeline)

In order to work properly, the functions need the following AWS policies:
```
- Effect: "Allow"
  Action:
    - ecr:GetAuthorizationToken
    - ecr:DescribeRepositories
    - ecr:ListImages
    - ecr:DescribeImages
    - ecr:DescribeImageScanFindings
    - ecr:StartImageScan
    - ecr:PutImageScanningConfiguration
    - logs:PutLogEvents
    - logs:CreateLogGroup
    - logs:CreateLogStream
  Resource: "*"
  
  # Only if SNS exporter is used
- Effect: "Allow"
  Action:
    - sns:Publish
  Resources: "arn:aws:sns:${env:AWS_REGION}:*:${opt:sns-topic}"
```
The proper role and policies are created by the serverless framework during deployment.

### How to compile code
1. compile:
    1. `make build` GOOS flag will be set dinamically (darwin or linux). e.g.: running command on osx will build osx executable.
    2. `make build-linux` GOOS target will be linux, appropriate for Lambda
2. test:
`make test`
3. lint: 
`make lint`

```text
NOTE: make build compiles both functions.
```

### How to deploy functions

Deploy with minimum configuration:
```bash
$ make build-linux
$ AWS_REGION=us-east-1 serverless deploy --stage production
```

```text
NOTE: the Serverless framework will create a Cloudformation deployment.
```

#### Deploy without bulding the project
- Install [Serverless framework](https://www.serverless.com/framework/docs/getting-started/) on your local machine
- Navigate to the latest [release](https://github.com/nagypeterjob/ecr-scan-lambda/releases) and download `deployment.zip`
- Unzip `deployment.zip` and place the two binaries in a directory called `bin`
```
root/
├── bin/
│   ├── report-linux
│   └── scan-linux
└── serverless.yml
```
Then run:
```bash
$ AWS_REGION=us-east-1 serverless deploy --stage production
```

## Exporters

There are multiple exporters `ecr-report-lambda` can work with. If there is not a suitable one already, feel free to contribute one by implementing the [exporter interface](https://github.com/nagypeterjob/ecr-scan-lambda/blob/master/pkg/exporters/exporter.go)!

To enable any exporter, set `EXPORTERS` environment variable (see [section](#environment-variables))

### Log

The default exporter to use is *Log*. The exporter does nothing else but prints the vulnerability report to stdout so it appears in logs. It is just an example implementation of the exporter interface and also comes handy when debugging.

### Mailgun

Reports can be sent via Mailgun to arbitrary recipients in the same plaintext format that Log exporter uses. Configure exporter by sertting `MAILGUN_API_KEY`, `MAILGUN_FROM` and `MAILGUN_RECIPIENTS` environment variables.

### Slack

Post vulnerability reports to a selected Slack channel with Slack exporter.

Get a Slack application [token](https://api.slack.com/start/building)
  * Create a new Application (bot)
  * Choose the channel the bot will post messages to
  * Set oauth scope **channel:write** (you may have to redeploy the application to the workspace)
  * Grab the Bot Oauth Access Token
  * Set the `SLACK_TOKEN` and `SLACK_CHANNEL` environment variables
  * Invite the bot to the selected slack channel (@BotName, then `Invite Bot`)

### SNS

SNS exporter enables sending vulnerability reports to an arbitrary sns topic. Start using the exporter by setting the `SNS_TOPIC_ARN` environment variable.

To deploy function using SNS, uncomment the sns role in **serverless.yml** under `roleStatements` key and run:
```bash
AWS_REGION=us-east-1 serverless deploy --stage production --sns-topic <TOPIC_NAME>
```

## Environment variables

### For ecr-scan-lambda
- **ENV** - Lambda function environment, **Required**
- **REGION** - AWS region where the function is executed, **Required**
- **ECR_ID** - Override the default ECR registry belonging to the account **Optional** (*Default:* ``)
- **IMAGE_TAG** - Override the container image tag being scanned  **Optional** (*Default:* `latest`)
- **LOG_LEVEL** - Function log level **Optional** (*Default:* `INFO`)
- **NUM_WORKERS** - Number of goroutines spawned **Optional** (*Default:* `2`)

### For ecr-report-lambda
- **ENV** - Lambda function environment, **Required**
- **REGION** - AWS region where the function is executed, **Required**
- **ECR_ID** - Override the default ECR registry belonging to the account **Optional** (*Default:* ``)
- **EXPORTERS** - Comma separated, smallcaps list of exporters to enable **Optional** (*Default:* `log`), *Example*: logs,mailgun,slack
- **IMAGE_TAG** - Override the container image tag being scanned  **Optional** (*Default:* `latest`)
- **LOG_LEVEL** - Function log level **Optional** (*Default:* `INFO`)
- **NUM_WORKERS** - Number of goroutines spawned **Optional** (*Default:* `2`)
- **MAILGUN_API_KEY** - Mailgun API KEY (Only relevant when Mailgun is enabled via `EXPORTERS`)
- **MAILGUN_FROM** -  Mailgun sender email address (Only relevant when Mailgun is enabled via `EXPORTERS`)
- **MAILGUN_RECIPIENTS** - Comma separated list of email addresses to send report to (Only relevant when Mailgun is enabled via `EXPORTERS`), *Example*: example@recart.com,example2@recart.com
- **MINIMUM_SEVERITY** - The minimum severity level which should be reported **Optional** (*Default*: `CRITICAL`) 
- **SLACK_TOKEN** - Slack API Token (Only relevant when Slack is enabled via `EXPORTERS`)
- **SLACK_CHANNEL** - Slack channel name to report to (with **#** prefix) (Only relevant when Slack is enabled via `EXPORTERS`), *Example*: #ecr-scan
- **SNS_TOPIC_ARN** - SNS topic to publish report to. (Only relevant when SNS is enabled via `EXPORTERS`)


## Screenshots
![alt text](https://github.com/nagypeterjob/ecr-scan-lambda/blob/master/screenshots/slack_exporter.png "Slack exporter")

## Price
According to [dashbird](https://dashbird.io/lambda-cost-calculator/) calculator, the default deployment will cost virtually nothing.

## Improvement ideas
- Mocks and some tests could be definitely improved. More tests should be added.
- Format Mailgun (or any email exporter) message as HTML

## Issues
If stumble upon errors or just need a feature request, please open an issue. PRs are welcome.
