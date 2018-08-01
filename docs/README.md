<!--
Copyright 2017 - The TXTDirect Authors

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

# Documentation

For configuration examples for the caddy plugin look at our [examples section](/examples/README.md#configuration)

# Specification
*The implementation and specification are a work in progress*

## TXT record
### Basic requirements
* TXT record key-value pairs must be delimited by semicolons ";"
* Key and value must be encoded using utf8
* Punycode for domain names should be used
* Ordering key-value pairs can be done

### Location
* TXT record must be accessible under the subdomain "\_redirect"

### URLs
* All URLs must be encoded
* ";" must be escaped as "%3B"

Examples:  
    ; -> %3B  
    ? -> %3F  
    = -> %3D  
`https://example.com/page/about=us -> https://example.com/page/about%3Dus`

External Links:  
    https://en.wikipedia.org/wiki/Percent-encoding  
    https://tools.ietf.org/html/rfc3986#page-11

### type=host
*v*
* Mandatory
* Permitted values: "txtv0"

*to*
* Recommended
* Default: Fallbacks such as `www` or `redirect` config
* Permitted values: "absolute/relative URL"

*code*
* Optional
* Default: "301"
* Permitted values: "301", "302"

### type=path
*v*
* Mandatory
* Permitted values: "txtv0"

*to*
* Optional
* Default: Fallbacks such as `www` or `redirect` config
* General path fallback

*root*
* Optional
* Default: Fallbacks such as `www` or `redirect` config
* Root path fallback

*from*
* Permitted values: "simplified regex"

*re*
* Permitted values: "regex"

Wildcards for catch all records can be used by providing "\_" as subdomain.  
Wildcards must be subdomains under a specific domain.
  `_redirect._.test` <-- is allowed
  `_redirect.test._` <-- is not allowed
  
### type=gometa
*v*
* Mandatory
* Permitted values: "txtv0"

*to*
* Recommended
* Default: Fallbacks such as `www` or `redirect` config
* Permitted values: "absolute URL for repository"

<!--
The specifics especially concerning the dep registry idea need to be fleshed out.

## type=dep
*v*
* Mandatory
* Possible values: "txtv0"

*to*
* Recommended
* Default: Last plain value "v=txtv0;to=github.com/user/package" == "v=txtv0;example.com/user/package"
* Possible values: "absolute URL pointing to the package root"

### type=dockerv2
*v*
* Mandatory
* Possible values: "txtv0"

*to*
* Recommended
* Default: Last plain value "v=txtv0;to=example.com" == "v=txtv0;example.com"
* Possible values: "absolute/relative URL"
-->

[Full list of TXT record examples](/examples/README.md#txt-record)

---

Start your first contribution with some documentation.

See how you can contribute with our [contribution guide](/CONTRIBUTING.md).
