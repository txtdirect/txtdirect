FROM alpine:3.6
RUN apk --no-cache add ca-certificates
ADD txtdirect /caddy
CMD ["/caddy"]