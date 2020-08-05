FROM golang:1.14-alpine

RUN apk --no-cache add git

WORKDIR /e2e

RUN GO111MODULE=on go get -u github.com/google/go-containerregistry/cmd/crane@34fb8ff

CMD [ "go" ]
