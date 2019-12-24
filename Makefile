COMMIT_HASH=$$(git rev-list -1 HEAD)
TAG_VERSION=$$(git tag --sort=committerdate | tail -1)

.PHONY: test
test:
	go test -count=1 -v $(shell go list ./... | grep -v /vendor/ | grep -v /metrics/cmd)

.PHONY: lint
lint:
	golint -set_exit_status `go list ./...`

.PHONY: build
build:
	GOOS=darwin go build -o="bin/handler" -ldflags="\
	-X 'github.com/nagypeterjob/ecr-scan-lambda/internal/version.commitHash=$(COMMIT_HASH)' \
	-X 'github.com/nagypeterjob/ecr-scan-lambda/internal/version.version=$(TAG_VERSION)' \
	-X 'github.com/nagypeterjob/ecr-scan-lambda/internal/version.date=$$(date)' \
	-X 'github.com/nagypeterjob/ecr-scan-lambda/internal/version.author=Peter Nagy<peter.nagy@recart.com>'" handler.go

.PHONY: build-linux 
build-linux:
	GOOS=linux go build -o="bin/handler" -ldflags="\
	-X 'github.com/nagypeterjob/ecr-scan-lambda/internal/version.xommitHash=$(COMMIT_HASH)' \
	-X 'github.com/nagypeterjob/ecr-scan-lambda/internal/version.version=$(TAG_VERSION)' \
	-X 'github.com/nagypeterjob/ecr-scan-lambda/internal/version.date=$$(date)' \
	-X 'github.com/nagypeterjob/ecr-scan-lambda/internal/version.author=Peter Nagy<peter.nagy@recart.com>'" handler.go
	
vendor:
	go mod vendor