FROM docker.io/library/alpine:3.20

RUN apk --no-cache add ca-certificates

ADD txtdirect /txtdirect

ENTRYPOINT ["/txtdirect"]
