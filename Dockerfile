FROM alpine:3.9

RUN apk --no-cache add ca-certificates tor

ADD txtdirect /txtdirect

CMD ["/txtdirect"]