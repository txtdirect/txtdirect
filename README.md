<a href='https://about.txtdirect.org'><img src='https://github.com/txtdirect/txtdirect/blob/master/media/logo.svg' width='500'/></a>

DNS TXT-record based redirects

 [![state](https://img.shields.io/badge/state-unstable-red.svg)]() [![release](https://img.shields.io/github/release/txtdirect/txtdirect.svg)](https://github.com/txtdirect/txtdirect/releases) [![license](https://img.shields.io/github/license/txtdirect/txtdirect.svg)](LICENSE) [![Build Status](https://travis-ci.org/txtdirect/txtdirect.svg?branch=master)](https://travis-ci.org/txtdirect/txtdirect) [![Go Report Card](https://goreportcard.com/badge/github.com/txtdirect/txtdirect)](https://goreportcard.com/report/github.com/txtdirect/txtdirect)

**NOTE: This is a work-in-progress, we do not consider it production ready. Use at your own risk.**

# TXTDIRECT
Convenient and minimalistic DNS based redirects

## Using TXTDIRECT
Take a look at our full [documentation](/docs).

## Examples

Redirect requests from example.com to www.example.com:

```
txtdirect {
    enable www
}
```

Redirect to host provided in TXT record if one is found, otherwise 404:

```
txtdirect {
    enable host
}
```

Redirect to host provided in TXT record if one is found, otherwise redirect to www.example.com:

```
txtdirect {
    enable host www
}
```

Enable go package vanity URLs:

```
txtdirect {
    enable gometa
}
```

Redirect to example.com if no TXT record is found for the request:

```
txtdirect {
    redirect https://example.com
}
```

Enable everything except redirection to www.example.com from example.com:

```
txtdirect {
    disable www
}
```

## Support
For detailed information on support options see our [support guide](/SUPPORT.md).

## Helping out
Best place to start is our [contribution guide](/CONTRIBUTING.md).

----

*Code is licensed under the [Apache License, Version 2.0](/LICENSE).*  
*Documentation/examples are licensed under [Creative Commons BY-SA 4.0](/docs/LICENSE).*  
*Illustrations, trademarks and third-party resources are owned by their respective party and are subject to different licensing.*

*The TXTdirect logo was created by [Florin Luca](https://99designs.com/profiles/florinluca)*

---

Copyright 2017 - The TXTDIRECT Authors
