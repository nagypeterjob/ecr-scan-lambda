# ecr-scan-lambda
Lambdas which does ECR scan and sends results to slack

## Getting started

The present repository does a little bit of SlackOps by reporting basic daily ECR vulnerablity scan reports to an arbitrary Slack channel. 
The serverless deployment consists of tho AWS Lambda functions:
- `ecr-scan-lambda` for enabling *ScanOnPush* parameter on each repository and running scans (There is a one scan / image / day limit by AWS)
- `ecr-report-lambda` for sending collected vulnerablity report to your Slack channel

Both functions are timed & run by Cloudwatch events. Can be configured in *serverless.yml*

### Prerequisits
1. Get a Slack application [token](https://api.slack.com/start/building)
2. Read through the environment variables for bot functions
3. Use serverless.yml to deploy functions to your AWS environment (or integrate it to your CI/CD pipeline)

In order to work properly, the functions need the following AWS policies:
```
- Effect: "Allow"
  Action:
    - ecr:GetAuthorizationToken
    - ecr:DescribeRepositories
    - ecr:ListImages
    - ecr:DescribeImages
    - ecr:DescribeImageScanFindings
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

## Environment variables

### For ecr-scan-lambda
- *ENV* - Lambda function environment, *Required*
- *AWS_REGION* - AWS region to deploy functions to, *Required*
- *MINIMUM_SEVERITY* - The minimum severity level which should be reported, `Default: HIGH`, *Optional* 
- *SLACK_TOKEN* - Slack API Token, *Required*
- *SLACK_CHANNEL* - Slack channel name to report to (with #prefix), *Required*, *Example*: #ecr-scan
- *ECR_ID* - If you want to use other ECR than the default, *Optional*
- *EMOJI_CRITICAL* - Override default emoji for this severity level,  `Default: :no_entry:`, *Optional*
- *EMOJI_HIGH* - Override default emoji for this severity level,  `Default: :warning:`, *Optional*
- *EMOJI_MEDIUM* - Override default emoji for this severity level,  `Default: :pill:`, *Optional*
- *EMOJI_LOW* - Override default emoji for this severity level,  `Default: :rain_cloud:`, *Optional*
- *EMOJI_INFORMATIONAL* - Override default emoji for this severity level,  `Default: :information_source:`, *Optional*
- *EMOJI_UNDEFINED* - Override default emoji for this severity level,  `Default: :question:`, *Optional*

### For ecr-report-lambda
- *ENV* - Lambda function environment, *Required*
- *AWS_REGION* - AWS region to deploy functions to, *Required*
- *ECR_ID* - If you want to use other ECR than the default, *Optional*

## Known problems waiting for improvement
- There are code duplications in report/report.go and scan/scan.go. Functions like `func (a *app) listRepositories(maxRepos int) (*ecr.DescribeRepositoriesOutput, error)` could be lifted to `internal` and reused in both handlers.
- `PutImageScanningConfiguration` and `StartImageScan` in scan/scan.go can run independently for each repository. The code could leverage goroutines and run concurrently.
- Same thing with `DescribeImageScanFindings` in report/report.go. The code could leverage goroutines and run concurrently.
- Mocks and some tests could be definitely improved. More tests should be added.
- The report funcion works well until 1000 repositories. The function currently doesn't implement paging. Paging for `DescribeRepositories` should be implemented.

## Issues