BIN := txtdirect
MAINTAINER := okkurlabs
VERSION := 0.4.0
IMAGE := $(MAINTAINER)/$(BIN):$(VERSION)

BUILD_GOOS := $(if $(GOOS),$(GOOS),linux)
BUILD_GOARCH := $(if $(GOARCH),$(GOARCH),amd64)

# Repo's root import path (under GOPATH).
PKG := github.com/txtdirect/txtdirect
CONTAINER ?= $(BIN)

.DEFAULT_GOAL := build

recipe:
	if [ -d caddy-copy ]; then cd caddy-copy && git checkout . && git pull && cd ..; else git clone https://github.com/mholt/caddy caddy-copy; fi
	find caddy-copy/caddyhttp/httpserver -name 'plugin.go' -type f -exec sed -i -e "s/gopkg/txtdirect/" -- {} +
	find caddy-copy/caddy/caddymain -name 'run.go' -type f -exec sed -i -e "s/\/\/ This is where other plugins get plugged in (imported)/_ \"github.com\/txtdirect\/txtdirect\/caddy\"/" -- {} +
	find caddy-copy/caddy/caddymain -name 'run.go' -type f -exec sed -i -e '/_ "github.com\/txtdirect\/txtdirect\/caddy"/a _ "github.com\/miekg\/caddy-prometheus"' -- {} +
	find caddy-copy/caddy/caddymain -name 'run.go' -type f -exec sed -i -e '/_ "github.com\/txtdirect\/txtdirect\/caddy"/a _ "github.com\/captncraig\/caddy-realip"' -- {} +
	find caddy-copy/caddy/caddymain -name 'run.go' -type f -exec sed -i -e 's/var EnableTelemetry = true/var EnableTelemetry = false/' -- {} +
	cd caddy-copy/caddy && \
	CGO_ENABLED=0 GOARCH=$(BUILD_GOARCH) GOOS=$(BUILD_GOOS) go build -ldflags="-s -w"
	mv caddy-copy/caddy/caddy ./$(BIN)

dependencies:
	go get github.com/mholt/caddy/caddy
	go get github.com/caddyserver/builds
	go get github.com/miekg/caddy-prometheus
	go get github.com/captncraig/caddy-realip
	go get gopkg.in/natefinch/lumberjack.v2
	go get github.com/miekg/dns
	go get github.com/gomods/athens/...
	rm -rf $(GOPATH)/src/github.com/gomods/athens/vendor/github.com/spf13/afero
	go get github.com/spf13/afero
	go get github.com/prometheus/client_golang/...

build: dependencies recipe

test: dependencies
	go test -v `go list ./... | grep -v caddy-copy`

image-build: docker-build
	docker build -t $(IMAGE) .

docker-run: image-build
	docker run --name $(CONTAINER) $(IMAGE)

docker-test:
	docker run -v $(shell pwd):/go/src/github.com/txtdirect/txtdirect golang:1.11-alpine /bin/sh -c "cd /go/src/github.com/txtdirect/txtdirect && apk add git gcc musl-dev make && GOROOT=\"/usr/local/go\" make test"

docker-build:
	docker run -v $(shell pwd):/go/src/github.com/txtdirect/txtdirect golang:1.11-alpine /bin/sh -c "cd /go/src/github.com/txtdirect/txtdirect && apk add git gcc musl-dev make && make build && rm -rf caddy-copy"

version:
	@echo $(VERSION)
