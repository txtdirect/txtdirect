/*
Copyright 2017 - The TXTDirect Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package txtdirect

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/caddyserver/caddy/caddyhttp/header"
	"github.com/caddyserver/caddy/caddyhttp/httpserver"
	"github.com/caddyserver/caddy/caddyhttp/proxy"
	"github.com/miekg/dns"
)

// Testing TXT records
var txts = map[string]string{
	// type=host
	"_redirect.host.e2e.test.":           "v=txtv0;to=https://plain.host.test;type=host;code=302",
	"_redirect.nocode.host.e2e.test.":    "v=txtv0;to=https://nocode.host.test;type=host",
	"_redirect.noversion.host.e2e.test.": "to=https://noversion.host.test;type=host",
	"_redirect.noto.host.e2e.test.":      "v=txtv0;type=host",
	// type=path
	"_redirect.path.e2e.test.":           "v=txtv0;to=https://fallback.path.test;root=https://root.fallback.test;type=path",
	"_redirect.nocode.path.e2e.test.":    "v=txtv0;to=https://nocode.fallback.path.test;type=host",
	"_redirect.noversion.path.e2e.test.": "to=https://noversion.fallback.path.test;type=path",
	"_redirect.noto.path.e2e.test.":      "v=txtv0;type=path",
	"_redirect.noroot.path.e2e.test.":    "v=txtv0;to=https://noroot.fallback.path.test;type=path;code=302",
	"_redirect.metapath.e2e.test.":       "v=txtv0;type=path",
	// type=gometa
	"_redirect.pkg.txtdirect.test.":           "v=txtv0;to=https://github.com/txtdirect/txtdirect;type=gometa;vcs=git",
	"_redirect.pkgweb.metapath.e2e.test.":     "v=txtv0;to=https://github.com/txtdirect/txtdirect;type=gometa;website=https://godoc.org/go.txtdirect.org/txtdirect",
	"_redirect.pkg.metapath.e2e.test.":        "v=txtv0;to=https://github.com/okkur/reposeed-server;type=gometa",
	"_redirect.second.pkg.metapath.e2e.test.": "v=txtv0;to=https://github.com/okkur/reposeed;type=gometa",
	// type=""
	"_redirect.about.test.": "v=txtv0;to=https://about.txtdirect.org",
	"_redirect.pkg.test.":   "v=txtv0;to=https://pkg.txtdirect.org;type=gometa",

	//
	//	Fallback records
	//

	// type=host
	"_redirect.fallbackhost.test.": "v=txtv0;to=https://{label3};type=host;code=302",

	// type=path
	"_redirect.fallbackpath.test.":        "v=txtv0;type=path;code=302;to=https://to.works.fine.test",
	"_redirect.to.fallbackpath.test.":     "v=txtv0;type=path;to=https://to.works.fine.test;code=302",
	"_redirect.refrom.fallbackpath.test.": "v=txtv0;type=path;re=exist;from=$1$2;to=https://to.works.fine.test;code=302",
	"_redirect.lenfrom.test.":             "v=txtv0;type=path;from=$1$2$3;to=https://lenfrom.fallback.test;code=302",

	// type=dockerv2
	"_redirect.fallbackdockerv2.test.":       "v=txtv0;to=https://docker.to.test/;root=https://docker.root.test;type=dockerv2",
	"_redirect.wrong.fallbackdockerv2.test.": "v=txtv0;to=:/wrong.uri/;type=dockerv2;website=https://gcr.io/",

	// type=gometa
	"_redirect.fallbackgometa.test.":          "v=txtv0;type=path;to=https://gometa.path.to.test",
	"_redirect.pathto.fallbackgometa.test.":   "v=txtv0;to=wrong:/url.test;type=gometa",
	"_redirect.redirect.fallbackgometa.test.": "v=txtv0;to=https://github.com/okkur/reposeed-server/;type=gometa",
}

// Testing DNS server port
const port = 6000

// Initialize dns server instance
var server = &dns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}

func TestMain(m *testing.M) {
	go RunDNSServer()
	os.Exit(m.Run())
}

func TestRedirectBlacklist(t *testing.T) {
	config := Config{
		Enable: []string{"path"},
	}
	req := httptest.NewRequest("GET", "https://txtdirect.com/favicon.ico", nil)
	w := httptest.NewRecorder()

	err := Redirect(w, req, config)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}

func Test_query(t *testing.T) {
	tests := []struct {
		zone string
		txt  string
	}{
		{
			"_redirect.about.test.",
			txts["_redirect.about.test."],
		},
		{
			"_redirect.pkg.test.",
			txts["_redirect.pkg.test."],
		},
	}
	for _, test := range tests {
		ctx := context.Background()
		c := Config{
			Resolver: "127.0.0.1:" + strconv.Itoa(port),
		}
		resp, err := query(test.zone, ctx, c)
		if err != nil {
			t.Fatal(err)
		}
		if resp[0] != txts[test.zone] {
			t.Fatalf("Expected %s, got %s", txts[test.zone], resp[0])
		}
	}
}

func parseDNSQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeTXT:
			log.Printf("Query for %s\n", q.Name)
			m.Answer = append(m.Answer, &dns.TXT{
				Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60},
				Txt: []string{txts[q.Name]},
			})
		}
	}
}

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		parseDNSQuery(m)
	}

	w.WriteMsg(m)
}

func RunDNSServer() {
	dns.HandleFunc("test.", handleDNSRequest)
	err := server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Printf("Failed to start server: %s\n ", err.Error())
	}
}

func TestRedirectE2e(t *testing.T) {
	tests := []struct {
		url      string
		expected string
		enable   []string
		referer  bool
	}{
		{
			url:      "https://host.e2e.test",
			expected: "https://plain.host.test",
			enable:   []string{"host"},
		},
		{
			url:      "https://nocode.host.e2e.test",
			expected: "https://nocode.host.test",
			enable:   []string{"host"},
		},
		{
			url:      "https://noversion.host.e2e.test",
			expected: "https://noversion.host.test",
			enable:   []string{"host"},
		},
		{
			url:      "https://noto.host.e2e.test",
			expected: "",
			enable:   []string{"host"},
		},
		{
			url:      "https://path.e2e.test/",
			expected: "https://root.fallback.test",
			enable:   []string{"path", "host"},
		},
		{
			url:      "https://path.e2e.test/nocode",
			expected: "https://nocode.fallback.path.test",
			enable:   []string{"path", "host"},
		},
		{
			url:      "https://path.e2e.test/noversion",
			expected: "https://noversion.fallback.path.test",
			enable:   []string{"path", "host"},
		},
		{
			url:      "https://path.e2e.test/noto",
			expected: "",
			enable:   []string{"path", "host"},
		},
		{
			url:      "https://path.e2e.test/noroot",
			expected: "https://noroot.fallback.path.test",
			enable:   []string{"path", "host"},
		},
		{
			url:      "https://pkg.txtdirect.test?go-get=1",
			expected: "https://github.com/txtdirect/txtdirect",
			enable:   []string{"gometa"},
		},
		{
			url:      "https://metapath.e2e.test/pkg?go-get=1",
			expected: "https://github.com/okkur/reposeed-server",
			enable:   []string{"gometa", "path"},
		},
		{
			url:      "https://metapath.e2e.test/pkg/second?go-get=1",
			expected: "https://github.com/okkur/reposeed",
			enable:   []string{"gometa", "path"},
		},
		{
			url:      "https://127.0.0.1/test",
			expected: "404",
			enable:   []string{"host"},
		},
		{
			url:      "https://192.168.1.2",
			expected: "404",
			enable:   []string{"host"},
		},
		{
			url:      "https://2001:db8:1234:0000:0000:0000:0000:0000",
			expected: "404",
			enable:   []string{"host"},
		},
		{
			url:      "https://2001:db8:1234::/48",
			expected: "404",
			enable:   []string{"host"},
		},
		{
			url:      "https://host.e2e.test",
			expected: "https://plain.host.test",
			enable:   []string{"host"},
			referer:  true,
		},
	}
	for _, test := range tests {
		req := httptest.NewRequest("GET", test.url, nil)
		resp := httptest.NewRecorder()
		c := Config{
			Resolver: "127.0.0.1:" + strconv.Itoa(port),
			Enable:   test.enable,
		}
		if err := Redirect(resp, req, c); err != nil {
			t.Errorf("Unexpected error occured: %s", err.Error())
		}
		if !strings.Contains(resp.Body.String(), test.expected) {
			t.Errorf("Expected %s to be in \"%s\"", test.expected, resp.Body.String())
		}
		if test.referer && resp.Header().Get("Referer") != req.Host {
			t.Errorf("Expected %s referer but got \"%s\"", req.Host, resp.Header().Get("Referer"))
		}
	}
}

func TestConfigE2e(t *testing.T) {
	tests := []struct {
		url    string
		txt    string
		enable []string
	}{
		{
			"https://e2e.txtdirect",
			txts["_redirect.path.txtdirect."],
			[]string{},
		},
		{
			"https://path.txtdirect/test",
			txts["_redirect.path.e2e.txtdirect."],
			[]string{"host"},
		},
		{
			"https://gometa.txtdirect",
			txts["_redirect.gometa.txtdirect."],
			[]string{"host"},
		},
	}
	for _, test := range tests {
		req := httptest.NewRequest("GET", test.url, nil)
		resp := httptest.NewRecorder()
		c := Config{
			Resolver: "127.0.0.1:" + strconv.Itoa(port),
			Redirect: "https://txtdirect.org",
		}
		Redirect(resp, req, c)
		if resp.Header().Get("Location") != c.Redirect {
			t.Errorf("Request didn't redirect to the specified URI after failure")
		}
	}
}

func Test_isIP(t *testing.T) {
	tests := []struct {
		host     string
		expected bool
	}{
		{
			"https://example.test",
			false,
		},
		{
			"http://example.test",
			false,
		},
		{
			"http://192.168.test.subdomain.test",
			false,
		},
		{
			"192.168.1.1",
			true,
		},
		{
			"https://122.221.122.221",
			true,
		},
		{
			"FE80:0000:0000:0000:0202:B3FF:FE1E:8329",
			true,
		},
		{
			"FE80::0202:B3FF:FE1E:8329",
			true,
		},
	}
	for _, test := range tests {
		if result := isIP(test.host); result != test.expected {
			t.Errorf("%s is an IP not a domain", test.host)
		}
	}
}

func Test_customResolver(t *testing.T) {
	tests := []struct {
		config Config
	}{
		{
			Config{
				Resolver: "127.0.0.1",
			},
		},
		{
			Config{
				Resolver: "8.8.8.8",
			},
		},
	}
	for _, test := range tests {
		resolver := customResolver(test.config)
		if resolver.PreferGo != true {
			t.Errorf("Expected PreferGo option to be enabled in the returned resolver")
		}
	}
}

func Test_contains(t *testing.T) {
	tests := []struct {
		array    []string
		word     string
		expected bool
	}{
		{
			[]string{"test", "txtdirect"},
			"test",
			true,
		},
		{
			[]string{"test", "txtdirect", "contains"},
			"txtdirect",
			true,
		},
		{
			[]string{"test", "txtdirect", "random"},
			"contains",
			false,
		},
	}
	for _, test := range tests {
		if result := contains(test.array, test.word); result != test.expected {
			t.Errorf("Expected %t but got %t.\nArray: %v \nWord: %v", test.expected, result, test.array, test.word)
		}
	}
}

func Test_getBaseTarget(t *testing.T) {
	tests := []struct {
		record record
		reqURL string
		url    string
		status int
	}{
		{
			record{
				To:   "https://example.test",
				Code: 200,
			},
			"https://nowhere.test",
			"https://example.test",
			200,
		},
		{
			record{
				To:   "https://{host}/{method}",
				Code: 200,
			},
			"https://somewhere.test",
			"https://somewhere.test/GET",
			200,
		},
		{
			record{
				To:   "https://testing.test{path}",
				Code: 301,
			},
			"https://example.test/testing/path",
			"https://testing.test/testing/path",
			301,
		},
	}
	for _, test := range tests {
		req := httptest.NewRequest("GET", test.reqURL, nil)
		to, status, err := getBaseTarget(test.record, req)
		if err != nil {
			t.Errorf("Expected the err to be nil but got %s", err)
		}
		if to != test.url {
			t.Errorf("Expected %s but got %s", test.url, to)
		}
		if err != nil {
			t.Errorf("Expected %d but got %d", test.status, status)
		}
	}
}

// Note: ServerHeader isn't a function, this test is for checking
// response's Server header.
func TestServerHeaderE2E(t *testing.T) {
	tests := []struct {
		url          string
		enable       []string
		headerPlugin bool
		proxyPlugin  bool
		expected     string
	}{
		{
			"https://host.e2e.test",
			[]string{"host"},
			false,
			false,
			"TXTDirect",
		},
		{
			"https://host.e2e.test",
			[]string{"host"},
			true,
			false,
			"Testing-TXTDirect",
		},
		{
			"https://host.e2e.test",
			[]string{"host"},
			false,
			true,
			"Testing-TXTDirect",
		},
	}
	for _, test := range tests {
		req := httptest.NewRequest("GET", test.url, nil)
		resp := httptest.NewRecorder()
		c := Config{
			Resolver: "127.0.0.1:" + strconv.Itoa(port),
			Enable:   test.enable,
		}
		err := Redirect(resp, req, c)
		if err != nil {
			t.Errorf("Unexpected Error: %s", err.Error())
		}

		// Use Caddy's header plugin to replace the header
		if test.headerPlugin {
			s := header.Headers{
				Next: httpserver.HandlerFunc(func(w http.ResponseWriter, r *http.Request) (int, error) {
					w.WriteHeader(http.StatusOK)
					return 0, nil
				}),
				Rules: []header.Rule{
					{Path: "/", Headers: http.Header{
						"Server": []string{test.expected},
					}},
				},
			}
			_, err := s.ServeHTTP(resp, req)
			if err != nil {
				t.Errorf("Couldn't replace the header using caddy's header plugin: %s", err.Error())
			}
		}

		if test.proxyPlugin {
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Server", test.expected)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Hello, client"))
			}))
			defer backend.Close()

			// Setup the fake upsteam
			uri, _ := url.Parse(backend.URL)
			u := fakeUpstream{
				name:          backend.URL,
				from:          "/",
				timeout:       proxyTimeout,
				fallbackDelay: fallbackDelay,
				host: &proxy.UpstreamHost{
					Name:         backend.URL,
					ReverseProxy: proxy.NewSingleHostReverseProxy(uri, "", http.DefaultMaxIdleConnsPerHost, proxyTimeout, fallbackDelay),
				},
			}

			p := &proxy.Proxy{
				Next:      httpserver.EmptyNext, // prevents panic in some cases when test fails
				Upstreams: []proxy.Upstream{&u},
			}
			p.ServeHTTP(resp, req)
		}

		if !contains(resp.Header()["Server"], test.expected) {
			t.Errorf("Expected \"Server\" header to be %s but it's %s", test.expected, resp.Header().Get("Server"))
		}
	}
}

// Setup fakeUpstream type and methods
type fakeUpstream struct {
	name          string
	host          *proxy.UpstreamHost
	from          string
	without       string
	timeout       time.Duration
	fallbackDelay time.Duration
}

func (u *fakeUpstream) AllowedPath(requestPath string) bool { return true }
func (u *fakeUpstream) GetFallbackDelay() time.Duration     { return 300 * time.Millisecond }
func (u *fakeUpstream) GetTryDuration() time.Duration       { return 1 * time.Second }
func (u *fakeUpstream) GetTryInterval() time.Duration       { return 250 * time.Millisecond }
func (u *fakeUpstream) GetTimeout() time.Duration           { return u.timeout }
func (u *fakeUpstream) GetHostCount() int                   { return 1 }
func (u *fakeUpstream) Stop() error                         { return nil }
func (u *fakeUpstream) From() string                        { return u.from }
func (u *fakeUpstream) Select(r *http.Request) *proxy.UpstreamHost {
	if u.host == nil {
		uri, err := url.Parse(u.name)
		if err != nil {
			log.Fatalf("Unable to url.Parse %s: %v", u.name, err)
		}
		u.host = &proxy.UpstreamHost{
			Name:         u.name,
			ReverseProxy: proxy.NewSingleHostReverseProxy(uri, u.without, http.DefaultMaxIdleConnsPerHost, u.GetTimeout(), u.GetFallbackDelay()),
		}
	}
	return u.host
}
