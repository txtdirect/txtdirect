<a href='https://about.txtdirect.org'><img src='https://github.com/txtdirect/txtdirect/blob/master/media/logo.svg' width='500'/></a>

DNS TXT-record based redirects

 [![state](https://img.shields.io/badge/state-unstable-red.svg)]() [![release](https://img.shields.io/github/release/txtdirect/txtdirect.svg)](https://github.com/txtdirect/txtdirect/releases) [![license](https://img.shields.io/github/license/txtdirect/txtdirect.svg)](LICENSE) [![Build Status](https://travis-ci.org/txtdirect/txtdirect.svg?branch=master)](https://travis-ci.org/txtdirect/txtdirect) [![Go Report Card](https://goreportcard.com/badge/github.com/txtdirect/txtdirect)](https://goreportcard.com/report/github.com/txtdirect/txtdirect)

**NOTE: This is a work-in-progress, we do not consider it production ready. Use at your own risk.**

# TXTDIRECT
Convenient and minimalistic DNS based redirects

## Using TXTDIRECT
**Redirect to host provided in TXT record:**  
*Default: Redirect to "www"-subdomain on empty record*
```
txtdirect {
  enable host www
}
```
*www.example.com -> about.example.com 301*
```
www.example.com               3600 IN CNAME  txtdirect.example.com.
_redirect.www.example.com     3600 IN TXT    "v=txtv0;to=https://about.example.com;type=host;code=301"
```

Further examples:  
[Configuration](/examples/README.md#configuration)  
[TXT-records](/examples/README.md#txt-record)  

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

Copyright 2017 - The TXTDIRECT Authors
