<a href='https://about.txtdirect.org'><img src='https://raw.githubusercontent.com/txtdirect/txtdirect/master/media/logo.svg?sanitize=true' width='500'/></a>

Convenient redirects based on DNS TXT-records

 [![state](https://img.shields.io/badge/state-beta-blue.svg)]() [![release](https://img.shields.io/github/release/txtdirect/txtdirect.svg)](https://txtdirect.org/releases/) [![license](https://img.shields.io/github/license/txtdirect/txtdirect.svg)](LICENSE) [![Build Status](https://travis-ci.org/txtdirect/txtdirect.svg?branch=master)](https://travis-ci.org/txtdirect/txtdirect)

**NOTE: This is a beta release, we do not consider it completely production ready yet. Use at your own risk.**

# TXTDirect
Convenient and minimalistic DNS based redirects, while controlling your data with your own DNS records.

## Using TXTDirect
Take a look at our full [documentation](https://txtdirect.org/docs/).

## Support
For detailed information on support options see our [support guide](/SUPPORT.md).

## Helping out
Best place to start is our [contribution guide](/CONTRIBUTING.md).

----

*Code is licensed under the [Apache License, Version 2.0](/LICENSE).*  
*Documentation/examples are licensed under [Creative Commons BY-SA 4.0](/docs/LICENSE).*  
*Illustrations, trademarks and third-party resources are owned by their respective party and are subject to different licensing.*

*The TXTDIRECT logo was created by [Florin Luca](https://99designs.com/profiles/florinluca)*

---

Copyright 2017 - The TXTDirect Authors

---

# How-to: build and run mock env
```
## binary build w/o docker
make build
./txtdirect run -c ./config/caddy/Caddyfile

## full environment mock w/ docker & docker compose
make image-build
docker compose up  # terminal-1
docker compose exec -it txtdirect sh # terminal-2
```

see the following config:

* [caddy (is webserver hosting txtdirect)](/config/caddy)
    * hosted on `port 8080`
* [bind9 (is mock DNS nameserver)](/config/bind)
    * mocked under the `example.lan` domain
    * named.conf - holds the [txtdirect redirect rules](https://about.txtdirect.org/docs/examples/host/) 
* [docker compose (describes the mock env)](/docker-compose.yaml)
    * `dns` directive points us to the mock bind server
    * `extra_hosts` acts as the initial lookup pushing toward txtdirect

```
/ # curl -L -v test-google.example.lan:8080
* Host test-google.example.lan:8080 was resolved.
* IPv6: (none)
* IPv4: 10.55.12.6
*   Trying 10.55.12.6:8080...
* Connected to test-google.example.lan (10.55.12.6) port 8080
> GET / HTTP/1.1
> Host: test-google.example.lan:8080
> User-Agent: curl/8.7.1
> Accept: */*
>
* Request completely sent off
< HTTP/1.1 301 Moved Permanently
< Cache-Control: max-age=604800
< Content-Type: text/html; charset=utf-8
< Location: https://www.google.com
< Server: TXTDirect
< Status-Code: 301
< Date: Fri, 24 May 2024 20:04:14 GMT
< Content-Length: 57
<
* Ignoring the response-body
* Connection #0 to host test-google.example.lan left intact
* Clear auth, redirects to port from 8080 to 443
* Issue another request to this URL: 'https://www.google.com/'
* Host www.google.com:443 was resolved.
* IPv6: 2607:f8b0:400a:804::2004
* IPv4: 142.250.217.100
*   Trying 142.250.217.100:443...
* Connected to www.google.com (142.250.217.100) port 443
* ALPN: curl offers h2,http/1.1
* TLSv1.3 (OUT), TLS handshake, Client hello (1):
*  CAfile: /etc/ssl/certs/ca-certificates.crt
*  CApath: /etc/ssl/certs
^C
/ #
```
