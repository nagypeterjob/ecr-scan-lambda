# ecr-scan-lambda
Lambdas which does ECR scan and sends results to slack

[![Go Report Card](https://goreportcard.com/badge/github.com/nagypeterjob/ecr-scan-lambda)](https://goreportcard.com/report/github.com/nagypeterjob/ecr-scan-lambda)
![](https://github.com/nagypeterjob/ecr-scan-lambda/workflows/Go%20tests/badge.svg?branch=master)


## Getting started

The present repository does a little bit of SlackOps by reporting basic daily ECR vulnerablity scan reports to an arbitrary Slack channel. 
The serverless deployment consists of tho AWS Lambda functions:
- `ecr-scan-lambda` for enabling *ScanOnPush* parameter on each repository and running scans (There is a one scan / image / day limit by AWS)
- `ecr-report-lambda` for sending collected vulnerablity report to your Slack channel

Both functions are timed & run by Cloudwatch events. Can be configured in **serverless.yml**

### Prerequisites
1. The report function returns scan restults for the `latest` version of each image. Make sure to have `latest` version tag for your images beside the semantically tagged ones. 
2. Get a Slack application [token](https://api.slack.com/start/building)
  * Create a new Application (bot)
  * Choose the channel the bot will post messages to
  * Set oauth scope **channel:write** (maybe redeploy the application to the workspace)
  * Grab the Bot Oauth Access Token
  * Invite the bot to the selected slack channel (@BotName, then `Invite Bot`)
3. Read through the environment variables for bot functions
4. Use **serverless.yml** to deploy functions to your AWS environment (or integrate it to your CI/CD pipeline)

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
```
The proper role and policies are created by the serverless framework.

### How to compile code
1. compile:
`make build`
2. test:
`make test`
3. lint: 
`make link`

```bash
NOTE: make build compiles both functions.
```

### How to deploy functions

Deploy with minimum configuration:
```bash
REGION=us-east-1 serverless deploy --stage production
```

```bash
NOTE: the Serverless framework will create a Cloudformation deployment.
```

## Environment variables

### For ecr-scan-lambda
- **ENV** - Lambda function environment, **Required**
- **REGION** - AWS region of the setup, **Required**
- **MINIMUM_SEVERITY** - The minimum severity level which should be reported, `Default: HIGH`, **Optional**
- **SLACK_TOKEN** - Slack API Token, **Required**
- **SLACK_CHANNEL** - Slack channel name to report to (with #prefix), **Required**, *Example*: #ecr-scan
- **ECR_ID** - If you want to use other ECR than the default, **Optional**
- **EMOJI_CRITICAL** - Override default emoji for this severity level,  `Default: :no_entry:`, **Optional**
- **EMOJI_HIGH** - Override default emoji for this severity level,  `Default: :warning:`, **Optional**
- **EMOJI_MEDIUM** - Override default emoji for this severity level,  `Default: :pill:`, **Optional**
- **EMOJI_LOW** - Override default emoji for this severity level,  `Default: :rain_cloud:`, **Optional**
- **EMOJI_INFORMATIONAL** - Override default emoji for this severity level,  `Default: :information_source:`, **Optional**
- **EMOJI_UNDEFINED** - Override default emoji for this severity level,  `Default: :question:`, **Optional**

### For ecr-report-lambda
- **ENV** - Lambda function environment, **Required**
- **REGION** - AWS region of the setup, **Required**
- **ECR_ID** - If you want to use other ECR than the default, **Optional**

## Known problems waiting for improvement
- `PutImageScanningConfiguration` and `StartImageScan` in scan/scan.go can run independently for each repository. The code could leverage goroutines and run concurrently.
- Same thing with `DescribeImageScanFindings` in report/report.go. The code could leverage goroutines and run concurrently.
- Mocks and some tests could be definitely improved. More tests should be added.
- The report funcion works well until 1000 repositories. The function currently doesn't implement paging. Paging for `DescribeRepositories` should be implemented.

## Issues
If stumble upon errors or just need a feature request, please open an issue. PRs are welcome.
