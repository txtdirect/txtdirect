<!--
Copyright 2019 - The TXTDirect Authors

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

# Local TXTDirect Configuration

## Linux

- Install CoreDNS
  ```
  $ go get github.com/coredns/coredns
  ```
- Add testing domain to /etc/hosts e.g. 127.0.0.1 example.test
- Create CoreDNS config file named 'Corefile'
  ```
  .:5353 {
    file /path/to/example.test example.test
    forward . 8.8.8.8
    errors stdout
    log
  }
  ```
- Create DNS record file named `example.test`

  ```
  @                  3600 IN SOA      ns.example.test. domains.example.test. (
                                        2017101010   ; serial
                                        5m           ; refresh
                                        5m           ; retry
                                        1w           ; expire
                                        12h    )     ; minimum

  @                  86400 IN NS      ns.example.test.
  @                  86400 IN NS      ns.example.test.

  @                     60 IN A       127.0.0.1

  _redirect             60 IN TXT     "v=txtv0;to=http://full-wildcard.worked.example.test;root=http://root.worked.example.test;type=path;code=302"

  _redirect.two.one     60 IN TXT     "v=txtv0;to=http://google.com;type=proxy;code=302"

  _redirect._.one       60 IN TXT     "v=txtv0;to=http://two-wildcard.worked.example.test;type=host;code=302"

  _redirect._._         60 IN TXT     "v=txtv0;to=http://two-one-wildcard.worked.example.test;type=host;code=302"

  nohost 60 IN A       127.0.0.1
  _redirect.nohost      60 IN TXT     "v=txtv0;to=http://nohost.worked.example.test;code=302"

  host 60 IN A       127.0.0.1
  _redirect.host        60 IN TXT     "v=txtv0;to=http://host.worked.example.test;type=host;code=302"

  kubernetes 60 IN A       127.0.0.1
  _redirect.kubernetes  60 IN TXT     "v=txtv0;to=http://worked.example.test/{label1};type=host;code=302"
  ```

- Create a caddyfile named `caddy.test`
  ```
  {
      order txtdirect last
  }

  :8080

  tls internal

  txtdirect {
      enable path host
      resolver 127.0.0.1:5353
      logfile stdout
  }

  log
  ```
- Navigate to the directory in terminal where the Corefile was created and start CoreDNS with the Corefile
  ```
  $ coredns -conf Corefile
  ```
- Navigate to the local TXTDirect repository directory in terminal and rebuild TXTDirect.  
  You can build the project on your local machine or use docker.
  ```bash
  $ make build
  // OR
  $ make docker-build
  ```
- Start the caddyfile
  ```
  $ ./txtdirect run --config /<directory>/caddy.test --adapter caddyfile
  ```

---

Start your first contribution with some documentation.

See how you can contribute with our [contribution guide](/CONTRIBUTING.md).
