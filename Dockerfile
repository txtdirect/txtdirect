FROM golang:1.11-alpine

RUN apk --no-cache add ca-certificates make git && \
  mkdir -p $GOPATH/src/github.com/txtdirect/txtdirect

WORKDIR $GOPATH/src/github.com/txtdirect/txtdirect

COPY . .

RUN make dependencies

RUN make recipe

CMD ["./txtdirect"]