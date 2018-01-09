BIN=txtdirect

all:
	go get github.com/mholt/caddy/caddy
	go get github.com/caddyserver/builds
	go get github.com/miekg/caddy-prometheus
	go get github.com/captncraig/caddy-realip
	go get -d -u
	rm -rf caddy-copy
	git clone https://github.com/mholt/caddy caddy-copy
	find caddy-copy/caddyhttp/httpserver -name 'plugin.go' -type f -exec sed -i -e "s/gopkg/txtdirect/g" -- {} +
	find caddy-copy/caddy/caddymain -name 'run.go' -type f -exec sed -i -e "s/\/\/ This is where other plugins get plugged in (imported)/_ \"github.com\/txtdirect\/txtdirect\/caddy\"/g" -- {} +
	find caddy-copy/caddy/caddymain -name 'run.go' -type f -exec sed -i -e '/_ "github.com\/txtdirect\/txtdirect\/caddy"/a _ "github.com\/miekg\/caddy-prometheus"' -- {} +
	find caddy-copy/caddy/caddymain -name 'run.go' -type f -exec sed -i -e '/_ "github.com\/txtdirect\/txtdirect\/caddy"/a _ "github.com\/captncraig\/caddy-realip"' -- {} +
	cd caddy-copy/caddy && \
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w"
	mv caddy-copy/caddy/caddy ./$(BIN)

docker: all
	docker build -t seetheprogress/txtdirect:dev-$$(git rev-parse HEAD | cut -c1-6) .
