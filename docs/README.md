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

## Specification
### General
* The TXT record must consist of multiple sections, delimited by semicolons(;)
* Each section is a key=value pair
* The encoding is utf8.
* Punycode for domain names is permitted, but not required.
* The ordering of the tuples (a key-value pair) is arbitary, and not mandated.

### Keys
*v=*:
* Mandatory
* Possible values: "txtv0"

*to=*:
* Recommended
* Default: Last plain value "v=txtv0;to=example.com" == "v=txtv0;example.com"
* Possible values: "absolute URLs", "relative URLs"

*code=*:
* Optional
* Default: "301"
* Possible values: "301", "302"

*type=*:
* Mandatory
* Possible values: "host", "gometa"

---

Start your first contribution with some documentation.

See how you can contribute with our [contribution guide](/CONTRIBUTING.md).
