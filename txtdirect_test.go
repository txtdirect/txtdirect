/*
Copyright 2017 - The TXTdirect Authors
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
	"os"
	"strconv"
	"strings"
	"testing"

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

	// type=path
	"_redirect.fallbackpath.test.":             "v=txtv0;type=path",
	"_redirect.withoutroot.fallbackpath.test.": "v=txtv0;type=path",

	// type=dockerv2
	"_redirect.fallbackdockerv2.test.":         "v=txtv0;type=path",
	"_redirect.correct.fallbackdockerv2.test.": "v=txtv0;to=https://gcr.io/;type=dockerv2",
	"_redirect.wrong.fallbackdockerv2.test.":   "v=txtv0;to=://gcr.io/;type=dockerv2",

	// type=gometa
	"_redirect.fallbackgometa.test.":          "v=txtv0;type=path",
	"_redirect.website.fallbackgometa.test.":  "v=txtv0;to=https://github.com/okkur/reposeed-server/;website=https://about.okkur.io/;type=gometa",
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
	}{
		{
			"https://host.e2e.test",
			"https://plain.host.test",
			[]string{"host"},
		},
		{
			"https://nocode.host.e2e.test",
			"https://nocode.host.test",
			[]string{"host"},
		},
		{
			"https://noversion.host.e2e.test",
			"https://noversion.host.test",
			[]string{"host"},
		},
		{
			"https://noto.host.e2e.test",
			"",
			[]string{"host"},
		},
		{
			"https://path.e2e.test/",
			"https://root.fallback.test",
			[]string{"path", "host"},
		},
		{
			"https://path.e2e.test/nocode",
			"https://nocode.fallback.path.test",
			[]string{"path", "host"},
		},
		{
			"https://path.e2e.test/noversion",
			"https://fallback.path.test",
			[]string{"path", "host"},
		},
		{
			"https://path.e2e.test/noto",
			"",
			[]string{"path", "host"},
		},
		{
			"https://path.e2e.test/noroot",
			"https://fallback.path.test",
			[]string{"path", "host"},
		},
		{
			"https://pkg.txtdirect.test?go-get=1",
			"https://github.com/txtdirect/txtdirect",
			[]string{"gometa"},
		},
		{
			"https://metapath.e2e.test/pkg?go-get=1",
			"https://github.com/okkur/reposeed-server",
			[]string{"gometa", "path"},
		},
		{
			"https://metapath.e2e.test/pkg/second?go-get=1",
			"https://github.com/okkur/reposeed",
			[]string{"gometa", "path"},
		},
		{
			"https://127.0.0.1/test",
			"404",
			[]string{"host"},
		},
		{
			"https://192.168.1.2",
			"404",
			[]string{"host"},
		},
		{
			"https://2001:db8:1234:0000:0000:0000:0000:0000",
			"404",
			[]string{"host"},
		},
		{
			"https://2001:db8:1234::/48",
			"404",
			[]string{"host"},
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
			t.Fatalf("Unexpected error occured: %s", err.Error())
		}
		if !strings.Contains(resp.Body.String(), test.expected) {
			t.Fatalf("Expected %s to be in \"%s\"", test.expected, resp.Body.String())
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
			Enable:   test.enable,
		}
		err := Redirect(resp, req, c)
		if err == nil && !strings.Contains(err.Error(), "option disabled") {
			t.Fatalf("required option is not enabled, but there is no error returned")
		}
	}
}

func Test_fallback(t *testing.T) {
	tests := []struct {
		url      string
		code     int
		redirect string
	}{
		{
			"https://goto.fallback.test",
			301,
			"",
		},
		{
			"",
			403,
			"https://goto.redirect.test",
		},
		{
			"https://goto.fallback.test",
			404,
			"https://dontgoto.redirect.test",
		},
	}
	for _, test := range tests {
		req := httptest.NewRequest("GET", "https://testing.test", nil)
		resp := httptest.NewRecorder()
		c := Config{
			Redirect: test.redirect,
			Enable:   []string{"www"},
		}
		fallback(resp, req, test.url, "test", test.code, c)
		if resp.Code != test.code {
			t.Errorf("Response's status code (%d) doesn't match with expected status code (%d).", resp.Code, test.code)
		}
	}
}

func TestFallbackE2e(t *testing.T) {
	tests := []struct {
		url         string
		enable      []string
		fallbackURL string
		redirect    string
		headers     http.Header
	}{
		{
			"https://fallbackpath.test/withoutroot/",
			[]string{"www", "path"},
			"",
			"http://fallback.test",
			http.Header{},
		},
		{
			"https://fallbackpath.test/nosubdomain",
			[]string{"www", "path"},
			"",
			"http://fallback.test",
			http.Header{},
		},
		{
			"https://fallbackpath.test/",
			[]string{"www", "path"},
			"",
			"http://fallback.test",
			http.Header{},
		},
		{
			"https://fallbackdockerv2.test/correct",
			[]string{"www", "dockerv2", "path"},
			"https://gcr.io/",
			"",
			http.Header{"User-Agent": []string{"Docker-Server"}},
		},
		{
			"https://fallbackdockerv2.test/wrong",
			[]string{"www", "dockerv2", "path"},
			"https://gcr.io/",
			"",
			http.Header{"User-Agent": []string{"Docker-Client"}},
		},
		{
			"https://fallbackdockerv2.test/correct",
			[]string{"www", "dockerv2", "path"},
			"https://gcr.io/",
			"",
			http.Header{"User-Agent": []string{"Docker-Client"}},
		},
		{
			"https://fallbackgometa.test/website",
			[]string{"www", "host", "path"},
			"https://about.okkur.io/",
			"",
			http.Header{},
		},
		{
			"https://fallbackgometa.test/redirect",
			[]string{"www", "host", "path"},
			"",
			"https://about.okkur.io/",
			http.Header{},
		},
	}
	for _, test := range tests {
		req := httptest.NewRequest("GET", test.url, nil)
		req.Header = test.headers
		resp := httptest.NewRecorder()
		c := Config{
			Resolver: "127.0.0.1:" + strconv.Itoa(port),
			Enable:   test.enable,
			Redirect: test.redirect,
		}
		err := Redirect(resp, req, c)
		if resp.Result().Header.Get("Location") != test.redirect && resp.Result().Header.Get("Location") != test.fallbackURL {
			t.Errorf("Expected %s got %s", test.redirect, resp.Result().Header.Get("Location"))
		}
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
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
