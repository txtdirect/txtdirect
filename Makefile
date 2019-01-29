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
	git clone https://github.com/mholt/caddy caddy-copy
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
	go get github.com/spf13/afero
	go get github.com/prometheus/client_golang/...

build: build-container run-container get-binary

build-container:
	docker build -t $(IMAGE) .

run-container:
	docker run --name $(CONTAINER) $(IMAGE) -d

get-binary:
	docker cp $(CONTAINER):/go/src/$(PKG)/$(BIN) .