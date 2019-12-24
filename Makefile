COMMIT_HASH=$$(git rev-list -1 HEAD)
TAG_VERSION=$$(git tag --sort=committerdate | tail -1)

.PHONY: test
test:
	go test -count=1 -v $(shell go list ./... | grep -v /vendor/ | grep -v /metrics/cmd)

.PHONY: build
build:
	GOOS=darwin go build -o="bin/handler" -ldflags="\
	-X 'github.com/nagypeterjob/ecr-vuln-alert-lambda/internal/version.CommitHash=$(COMMIT_HASH)' \
	-X 'github.com/nagypeterjob/ecr-vuln-alert-lambda/internal/version.Version=$(TAG_VERSION)' \
	-X 'github.com/nagypeterjob/ecr-vuln-alert-lambda/internal/version.Date=$$(date)' \
	-X 'github.com/nagypeterjob/ecr-vuln-alert-lambda/internal/version.Author=Peter Nagy<peter.nagy@recart.com>'" handler.go

.PHONY: build-linux 
build-linux:
	GOOS=linux go build -o="bin/handler" -ldflags="\
	-X 'github.com/nagypeterjob/ecr-vuln-alert-lambda/internal/version.CommitHash=$(COMMIT_HASH)' \
	-X 'github.com/nagypeterjob/ecr-vuln-alert-lambda/internal/version.Version=$(TAG_VERSION)' \
	-X 'github.com/nagypeterjob/ecr-vuln-alert-lambda/internal/version.Date=$$(date)' \
	-X 'github.com/nagypeterjob/ecr-vuln-alert-lambda/internal/version.Author=Peter Nagy<peter.nagy@recart.com>'" handler.go
	
vendor:
	go mod vendor