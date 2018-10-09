# Local TXTDirect Configuration
## Linux
* Install CoreDNS
    ```
    go get github.com/coredns/coredns
    ```
* Add testing domain to /etc/hosts e.g. 127.0.0.1 example.test
* Create CoreDNS config file named 'Corefile'
    ```
    .:5353 {
      file /path/to/example.test example.test
      forward . 8.8.8.8
      errors stdout
      log
    }
    ```
* Create DNS record file named `example.test`
    ```
    @                  3600 IN SOA      ns.example.com domains.example.com. (
                                          2010101010   ; serial
                                          5m           ; refresh
                                          5m           ; retry
                                          1w           ; expire
                                          12h    )     ; minimum
    @                  86400 IN NS      ns.example.com.
    @                  86400 IN NS      ns.example.com.

    @                     60 IN A       127.0.0.1
    _redirect             60 IN TXT     "v=txtv0;to=full-wildcard.worked.example.test;root=http//root.worked.example.test;type=path;code=302"
    _redirect.two.one     60 IN TXT     "v=txtv0;to=two-one.worked.example.test;type=host;code=302"
    redirect..one         60 IN TXT     "v=txtv0;to=two-wildcard.worked.example.test;type=host;code=302"
    redirect.._           60 IN TXT     "v=txtv0;to=two-one-wildcard.worked.example.test;type=host;code=302"
    _redirect.nohost      60 IN TXT     "v=txtv0;to=nohost.worked.example.test;code=302"
    _redirect.host        60 IN TXT     "v=txtv0;host.worked.example.test;type=host;code=302"
    ```
* Create a caddyfile named `caddy.test`
    ```
    example.test:8080 {
        tls off
        txtdirect {
            enable path host
            resolver 127.0.0.1:5353
            logfile stdout
        }
        log / stdout "{remote} - [{when_iso}] \"{method} {uri} {proto}\" {status} {size} {latency}"
        errors stdout
    }
    ```
* Navigate to the directory in terminal where the Corefile was created and start CoreDNS with the Corefile
    ```
    coredns -conf Corefile
    ```
*  Navigate to the local TXTDirect repository directory in terminal and rebuild TXTDirect
    ```
    make build
    ```
    *Note: if the above command does not pull in all the changes and create a TXTDirect binary file, use `make travis-build` instead.
* Start the caddyfile
    ```
    ./txtdirect -conf /<directory>/caddy.test
    ```

## Mac
The instructions for configuring and running TXTDirect on a Mac are the same as Linux with the below exceptions.

* Since the sed command is not available on Mac, a workaround is to install `gnu-sed`

    ```
    brew install gnu-sed
    ```
* Replace the contents of the Makefile in the local TXTDirect repository with the contents of this snippet
    
```
BIN=txtdirect
TAG=$(if $(TRAVIS_TAG),$(TRAVIS_TAG),dev)
COMMIT=$(if $(TRAVIS_COMMIT),$(TRAVIS_COMMIT),$(shell git rev-parse HEAD))
BUILD_REF=$(shell echo $(COMMIT) | cut -c1-6)

build: fetch-dependencies
	rm -rf caddy-copy
	git clone https://github.com/mholt/caddy caddy-copy
	find caddy-copy/caddyhttp/httpserver -name 'plugin.go' -type f -exec gsed -i -e "s/gopkg/txtdirect/" -- {} +
	find caddy-copy/caddy/caddymain -name 'run.go' -type f -exec gsed -i -e "s/\/\/ This is where other plugins get plugged in (imported)/_ \"github.com\/txtdirect\/txtdirect\/caddy\"/" -- {} +
	find caddy-copy/caddy/caddymain -name 'run.go' -type f -exec gsed -i -e '/_ "github.com\/txtdirect\/txtdirect\/caddy"/a _ "github.com\/miekg\/caddy-prometheus"' -- {} +
	find caddy-copy/caddy/caddymain -name 'run.go' -type f -exec gsed -i -e '/_ "github.com\/txtdirect\/txtdirect\/caddy"/a _ "github.com\/captncraig\/caddy-realip"' -- {} +
	find caddy-copy/caddy/caddymain -name 'run.go' -type f -exec gsed -i -e 's/const enableTelemetry = true/const enableTelemetry = false/' -- {} +
	cd caddy-copy/caddy && \
	CGO_ENABLED=0 GOOS=darwin go build -ldflags="-s -w"
	mv caddy-copy/caddy/caddy ./$(BIN)

travis-build: fetch-dependencies
	cd $$GOPATH/src/github.com/mholt/caddy && \
	find caddyhttp/httpserver -name 'plugin.go' -type f -exec gsed -i -e "s/gopkg/txtdirect/" -- {} + && \
	find caddy/caddymain -name 'run.go' -type f -exec gsed -i -e "s/\/\/ This is where other plugins get plugged in (imported)/_ \"github.com\/txtdirect\/txtdirect\/caddy\"/" -- {} + && \
	find caddy/caddymain -name 'run.go' -type f -exec gsed -i -e '/_ "github.com\/txtdirect\/txtdirect\/caddy"/a _ "github.com\/miekg\/caddy-prometheus"' -- {} + && \
	find caddy/caddymain -name 'run.go' -type f -exec gsed -i -e '/_ "github.com\/txtdirect\/txtdirect\/caddy"/a _ "github.com\/captncraig\/caddy-realip"' -- {} + && \
	find caddy/caddymain -name 'run.go' -type f -exec gsed -i -e 's/const enableTelemetry = true/const enableTelemetry = false/' -- {} + && \
	cd caddy && \
	CGO_ENABLED=0 GOOS=darwin go build -ldflags="-s -w"
	mv $$GOPATH/src/github.com/mholt/caddy/caddy/caddy txtdirect

fetch-dependencies:
	go get github.com/mholt/caddy/caddy
	go get github.com/caddyserver/builds
	go get github.com/miekg/caddy-prometheus
	go get github.com/captncraig/caddy-realip

docker:
	docker build -t seetheprogress/txtdirect:$(TAG)-$(BUILD_REF) .

docker-push:
	docker push seetheprogress/txtdirect:$(TAG)-$(BUILD_REF)

.PHONY: clean
clean:
	rm -rf caddy-copy/
	rm $(BIN)
```
---

Start your first contribution with some documentation.

See how you can contribute with our [contribution guide](/CONTRIBUTING.md).