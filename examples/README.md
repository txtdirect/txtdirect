<!--
Copyright 2017 - The TXTDIRECT Authors

This work is licensed under a Creative Commons Attribution-ShareAlike 4.0 International License;
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    https://creativecommons.org/licenses/by-sa/4.0/legalcode
Unless required by applicable law or agreed to in writing, documentation
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->

# Configuration
**Redirect all requests to "www"-subdomain:**  
*example.com -> www.example.com*
```
txtdirect {
  enable www
}
```

**Redirect to host provided in TXT record:**  
*Default: Return 404 on empty record*
```
txtdirect {
  enable host
}
```

**Redirect to host provided in TXT record:**  
*Default: Redirect to "www"-subdomain on empty record*
```
txtdirect {
  enable host www
}
```

**Redirect to host provided in TXT record:**  
*Default: Redirect to "www"-subdomain on empty record*
```
txtdirect {
  redirect https://example.com
}
```

**Enable everything except "www"-subdomain redirection:**  
```
txtdirect {
    disable www
}
```

**Enable go meta/vanity redirects:**  
*pkg.example.com -> github.com/some/pkg.git*
```
txtdirect {
  enable gometa
}
```

# TXT record
"txtdirect.example.com" is your hosted TXTDIRECT instance and is usually provided as CNAME.

For more details take a look at our [TXT-record specification](/docs/README.md#specification).

## Host
*example.com -> about.example.com 301*
```
example.com                   3600 IN A      127.0.0.1
_redirect.example.com         3600 IN TXT    "v=txtv0;to=https://about.example.com;type=host"
```

*www.example.com -> about.example.com 301*
```
www.example.com               3600 IN CNAME  txtdirect.example.com.
_redirect.www.example.com     3600 IN TXT    "v=txtv0;to=https://about.example.com;type=host;code=301"
```

*www.example.com -> about.example.com 302*
```
www.example.com               3600 IN CNAME  txtdirect.example.com.
_redirect.www.example.        3600 IN TXT    "v=txtv0;to=https://about.example.com;type=host;code=302"
```

*gophers.example.com -> example.com/gophers*
```
gophers.example               3600 IN CNAME  txtdirect.example.com.
_redirect.gophers.example.com 3600 IN TXT    "v=txtv0;to=https://example.com/gophers;type=host"
```

## Gometa
*pkg.example.com -> github.com/some/repo*
```
pkg.example.com               3600 IN CNAME  txtdirect.example.com.
_redirect.pkg.example.com     3600 IN TXT    "v=txtv0;to=https://github.com/some/repo;type=gometa"
```

---

We are happy to get new contributions.

See how you can contribute with our [contribution guide](/CONTRIBUTING.md).
