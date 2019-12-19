FROM golang:1.13-alpine

RUN apk --no-cache add git

WORKDIR /e2e

RUN GO111MODULE=on go get -u github.com/google/go-containerregistry/cmd/crane

CMD [ "go" ]
