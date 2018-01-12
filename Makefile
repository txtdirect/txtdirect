BIN=txtdirect
TAG=$(if $(TRAVIS_TAG),$(TRAVIS_TAG),dev)
COMMIT=$(if $(TRAVIS_COMMIT),$(TRAVIS_COMMIT),$(shell git rev-parse HEAD))
BUILD_REF=$(shell echo $(COMMIT) | cut -c1-6)

build: fetch-dependencies
	rm -rf caddy-copy
	git clone https://github.com/mholt/caddy caddy-copy
	find caddy-copy/caddyhttp/httpserver -name 'plugin.go' -type f -exec sed -i -e "s/gopkg/txtdirect/g" -- {} +
	find caddy-copy/caddy/caddymain -name 'run.go' -type f -exec sed -i -e "s/\/\/ This is where other plugins get plugged in (imported)/_ \"github.com\/txtdirect\/txtdirect\/caddy\"/g" -- {} +
	find caddy-copy/caddy/caddymain -name 'run.go' -type f -exec sed -i -e '/_ "github.com\/txtdirect\/txtdirect\/caddy"/a _ "github.com\/miekg\/caddy-prometheus"' -- {} +
	find caddy-copy/caddy/caddymain -name 'run.go' -type f -exec sed -i -e '/_ "github.com\/txtdirect\/txtdirect\/caddy"/a _ "github.com\/captncraig\/caddy-realip"' -- {} +
	cd caddy-copy/caddy && \
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w"
	mv caddy-copy/caddy/caddy ./$(BIN)

travis-build: fetch-dependencies
	cd $$GOPATH/src/github.com/mholt/caddy && \
	find caddyhttp/httpserver -name 'plugin.go' -type f -exec sed -i -e "s/gopkg/txtdirect/g" -- {} + && \
	find caddy/caddymain -name 'run.go' -type f -exec sed -i -e "s/\/\/ This is where other plugins get plugged in (imported)/_ \"github.com\/txtdirect\/txtdirect\/caddy\"/g" -- {} + && \
	find caddy/caddymain -name 'run.go' -type f -exec sed -i -e '/_ "github.com\/txtdirect\/txtdirect\/caddy"/a _ "github.com\/miekg\/caddy-prometheus"' -- {} + && \
	find caddy/caddymain -name 'run.go' -type f -exec sed -i -e '/_ "github.com\/txtdirect\/txtdirect\/caddy"/a _ "github.com\/captncraig\/caddy-realip"' -- {} + && \
	cd caddy && \
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w"
	mv $$GOPATH/src/github.com/mholt/caddy/caddy/caddy txtdirect

fetch-dependencies:
	go get github.com/mholt/caddy/caddy
	go get github.com/caddyserver/builds
	go get github.com/miekg/caddy-prometheus
	go get github.com/captncraig/caddy-realip
	go get -d -u

docker:
	docker build -t seetheprogress/txtdirect:$(TAG)-$(BUILD_REF) .

docker-push:
	docker push seetheprogress/txtdirect:$(TAG)-$(BUILD_REF)
