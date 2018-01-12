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

# Documentation

For configuration examples for the caddy plugin look at our [examples section](/examples/README.md#configuration)

# Specification
*implementation does not yet implement the full specification*

## TXT record
### Basic requirements
* The TXT record must consist of multiple sections, delimited by semicolons ";"
* Each section is a key=value pair
* The key and value encoding is utf8
* Punycode for domain names is permitted, but not required
* The ordering of the tuples (a key-value pair) is arbitary

### Location
The TXT record must be accessible under the subdomain "\_redirect".

### type=host
*v*
* Mandatory
* Possible values: "txtv0"

*to*
* Recommended
* Default: Last plain value "v=txtv0;to=example.com" == "v=txtv0;example.com"
* Possible values: "absolute/relative URL"

*code*
* Optional
* Default: "301"
* Possible values: "301", "302"

### type=path
*v*
* Mandatory
* Possible values: "txtv0"

*to*
* Optional
* Website address to redirect to for non special (dep, docker) users

*from*
* Possible values: "absolute/relative URL + simplified regex"

*re*
* Possible values: "absolute/relative URL + simplified regex"

Path forces the implementation to look up a new record.
Each match from `from` or `re` is added as subdomain to be looked up:
`from=/$1/$2` -> \_redirect.<2>.<1>

Wildcards for catch all records are also possible, by providing "\_" as subdomain.
The wildcards need to be subdomains under specific domains if used.
`_redirect._.test` <-- is allowed
`_redirect.test._` <-- is not allowed
  
### type=gometa
*v*
* Mandatory
* Possible values: "txtv0"

*to*
* Recommended
* Default: Last plain value "v=txtv0;to=example.com" == "v=txtv0;example.com"
* Possible values: "absolute URL being the repository route"

The specifics especially concerning the dep registry idea need to be fleshed out.

### type=dockerv2
*v*
* Mandatory
* Possible values: "txtv0"

*to*
* Recommended
* Default: Last plain value "v=txtv0;to=example.com" == "v=txtv0;example.com"
* Possible values: "absolute/relative URL"

[Full list of TXT record examples](/examples/README.md#txt-record)

---

Start your first contribution with some documentation.

See how you can contribute with our [contribution guide](/CONTRIBUTING.md).
