COMMIT_HASH=$$(git rev-list -1 HEAD)
TAG_VERSION=$$(git tag --sort=committerdate | tail -1)

.PHONY: test
test:
	go test -v $(shell go list ./... | grep -v /vendor/ | grep -v /metrics/cmd)

.PHONY: build
build:
	GOOS=darwin go build -o="bin/handler" handler.go

.PHONY: build-linux 
build-linux:
	GOOS=linux go build -o="bin/handler" handler.go
	
vendor:
	go mod vendor