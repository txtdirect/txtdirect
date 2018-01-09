BIN=txtdirect

all:
	go get github.com/mholt/caddy/caddy
	go get github.com/caddyserver/builds
	go get github.com/miekg/caddy-prometheus
	go get github.com/captncraig/caddy-realip
	go get -d -u
	find $$GOPATH/src/github.com/mholt/caddy/caddyhttp/httpserver -name 'plugin.go' -type f -exec sed -i -e "s/gopkg/txtdirect/g" -- {} +
	find $$GOPATH/src/github.com/mholt/caddy/caddy/caddymain -name 'run.go' -type f -exec sed -i -e "s/\/\/ This is where other plugins get plugged in (imported)/_ \"github.com\/txtdirect\/txtdirect\/caddy\"/g" -- {} +
	find $$GOPATH/src/github.com/mholt/caddy/caddy/caddymain -name 'run.go' -type f -exec sed -i -e '/_ "github.com\/txtdirect\/txtdirect\/caddy"/a _ "github.com\/miekg\/caddy-prometheus"' -- {} +
	find $$GOPATH/src/github.com/mholt/caddy/caddy/caddymain -name 'run.go' -type f -exec sed -i -e '/_ "github.com\/txtdirect\/txtdirect\/caddy"/a _ "github.com\/captncraig\/caddy-realip"' -- {} +
	cd $$GOPATH/src/github.com/mholt/caddy/caddy && \
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w"
	mv $$GOPATH/src/github.com/mholt/caddy/caddy/caddy ./$(BIN)

docker: all
	docker build -t seetheprogress/txtdirect:dev-$$(git rev-parse HEAD | cut -c1-6) .
