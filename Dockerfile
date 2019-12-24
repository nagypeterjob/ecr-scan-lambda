  
FROM golang:1.21 as build-stage

RUN mkdir -p /go/src/github.com/ecr-scan-lambda/
WORKDIR  /go/src/github.com/ecr-scan-lambda/
RUN go get y
ADD . ./
RUN make lint
RUN make test
RUN make build-linux

FROM node:10-alpine
WORKDIR /app
COPY --from=build-stage /go/src/github.com/ecr-scan-lambda /app/
RUN npm install -g serverless
RUN apk --no-cache add coreutils 
ADD . ./
CMD [ "serverless", "--noDeploy", "--stage", "dev" ]