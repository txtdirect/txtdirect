BIN := txtdirect
MAINTAINER := okkur
VERSION := 0.4.0
IMAGE := $(MAINTAINER)/$(BIN):$(VERSION)

BUILD_GOOS := $(if $(GOOS),$(GOOS),linux)
BUILD_GOARCH := $(if $(GOARCH),$(GOARCH),amd64)

CONTAINER ?= $(BIN)

.DEFAULT_GOAL := build

recipe:
	find caddy-copy/caddyhttp/httpserver -name 'plugin.go' -type f -exec sed -i -e "s/gopkg/txtdirect/" -- {} +
	find caddy-copy/caddy/caddymain -name 'run.go' -type f -exec sed -i -e "s/\/\/ This is where other plugins get plugged in (imported)/_ \"github.com\/miekg\/caddy-prometheus\"/" -- {} +
	find caddy-copy/caddy/caddymain -name 'run.go' -type f -exec sed -i -e '/_ "github.com\/miekg\/caddy-prometheus"/a _ "github.com\/txtdirect\/txtdirect\/caddy"' -- {} +
	find caddy-copy/caddy/caddymain -name 'run.go' -type f -exec sed -i -e '/_ "github.com\/miekg\/caddy-prometheus"/a _ "github.com\/SchumacherFM\/mailout"' -- {} +
	find caddy-copy/caddy/caddymain -name 'run.go' -type f -exec sed -i -e '/_ "github.com\/miekg\/caddy-prometheus"/a _ "github.com\/captncraig\/caddy-realip"' -- {} +
	find caddy-copy/caddy/caddymain -name 'run.go' -type f -exec sed -i -e 's/var EnableTelemetry = true/var EnableTelemetry = false/' -- {} +
	cd caddy-copy/caddy && \
	GO111MODULE=on CGO_ENABLED=0 GOARCH=$(BUILD_GOARCH) GOOS=$(BUILD_GOOS) go build -ldflags="-s -w"
	if [ -f ./$(BIN) ]; then rm ./$(BIN); fi
	mv caddy-copy/caddy/caddy ./$(BIN)

dependencies:
	if [ -d caddy-copy ]; then cd caddy-copy && git checkout . && git pull && cd ..; else git clone https://github.com/mholt/caddy caddy-copy; fi
	cd caddy-copy && \
	if [ -f go.mod ]; then echo "already initialized go modules"; else go mod init; fi && \
	if ! grep -q "txtdirect => ../" go.mod; then \
		echo -e "\nreplace github.com/txtdirect/txtdirect => ../" >> go.mod && \
		echo "replace github.com/mholt/caddy => ../caddy-copy" >> go.mod; \
	fi && \
	GO111MODULE=on go get github.com/lucas-clemente/quic-go@master && \
	GO111MODULE=on go get github.com/russross/blackfriday@master && \
	GO111MODULE=on go get github.com/txtdirect/txtdirect@master && \
	GO111MODULE=on go get github.com/SchumacherFM/mailout && \
	GO111MODULE=on go get github.com/captncraig/caddy-realip && \
	GO111MODULE=on go get github.com/miekg/caddy-prometheus && \
	cd ..

build: dependencies recipe

test: dependencies
	GO111MODULE=on go test -v `go list ./... | grep -v caddy-copy`

image-build: docker-build
	docker build -t $(IMAGE) .

docker-run: image-build
	docker run --name $(CONTAINER) $(IMAGE)

docker-test:
	docker run --network=host -v $(shell pwd):/source -v $(GOPATH)/pkg/mod:/go/pkg/mod golang:1.12-alpine /bin/sh -c "cd /source && apk add git gcc musl-dev make && GOROOT=\"/usr/local/go\" make test"

docker-build:
	docker run --network=host -v $(shell pwd):/source -v $(GOPATH)/pkg/mod:/go/pkg/mod golang:1.12-alpine /bin/sh -c "cd /source && apk add git gcc musl-dev make && make build && rm -rf caddy-copy"

version:
	@echo $(VERSION)
