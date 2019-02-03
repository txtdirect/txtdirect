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
	"fmt"
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
	"_redirect.pkg.metapath.e2e.test.":        "v=txtv0;to=https://github.com/okkur/reposeed-server;type=gometa",
	"_redirect.second.pkg.metapath.e2e.test.": "v=txtv0;to=https://github.com/okkur/reposeed;type=gometa",
	// type=""
	"_redirect.about.test.": "v=txtv0;to=https://about.txtdirect.org",
	"_redirect.pkg.test.":   "v=txtv0;to=https://pkg.txtdirect.org;type=gometa",
}

// Testing DNS server port
const port = 6000

// Initialize dns server instance
var server = &dns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}

func TestMain(m *testing.M) {
	go RunDNSServer()
	os.Exit(m.Run())
}

func TestParse(t *testing.T) {
	tests := []struct {
		txtRecord string
		expected  record
		err       error
	}{
		{
			"v=txtv0;to=https://example.com/;code=302",
			record{
				Version: "txtv0",
				To:      "https://example.com/",
				Code:    302,
				Type:    "host",
			},
			nil,
		},
		{
			"v=txtv0;to=https://example.com/",
			record{
				Version: "txtv0",
				To:      "https://example.com/",
				Code:    302,
				Type:    "host",
			},
			nil,
		},
		{
			"v=txtv0;to=https://example.com/;code=302",
			record{
				Version: "txtv0",
				To:      "https://example.com/",
				Code:    302,
				Type:    "host",
			},
			nil,
		},
		{
			"v=txtv0;to=https://example.com/;code=302;vcs=hg;type=gometa",
			record{
				Version: "txtv0",
				To:      "https://example.com/",
				Code:    302,
				Vcs:     "hg",
				Type:    "gometa",
			},
			nil,
		},
		{
			"v=txtv0;to=https://example.com/;code=302;type=gometa;vcs=git",
			record{
				Version: "txtv0",
				To:      "https://example.com/",
				Code:    302,
				Vcs:     "git",
				Type:    "gometa",
			},
			nil,
		},
		{
			"v=txtv0;to=https://example.com/;code=test",
			record{},
			fmt.Errorf("could not parse status code"),
		},
		{
			"v=txtv1;to=https://example.com/;code=test",
			record{},
			fmt.Errorf("unhandled version 'txtv1'"),
		},
		{
			"v=txtv0;https://example.com/",
			record{},
			fmt.Errorf("arbitrary data not allowed"),
		},
		{
			"v=txtv0;to=https://example.com/caddy;type=path;code=302",
			record{
				Version: "txtv0",
				To:      "https://example.com/caddy",
				Type:    "path",
				Code:    302,
			},
			nil,
		},
		{
			"v=txtv0;to=https://example.com/;key=value",
			record{
				Version: "txtv0",
				To:      "https://example.com/",
				Code:    302,
				Type:    "host",
			},
			nil,
		},
		{
			"v=txtv0;to={?url}",
			record{
				Version: "txtv0",
				To:      "https://example.com/testing",
				Code:    302,
				Type:    "host",
			},
			nil,
		},
		{
			"v=txtv0;to={?url};from={method}",
			record{
				Version: "txtv0",
				To:      "https://example.com/testing",
				Code:    302,
				Type:    "host",
				From:    "GET",
			},
			nil,
		},
	}

	for i, test := range tests {
		r := record{}
		c := Config{
			Enable: []string{test.expected.Type},
		}
		req, _ := http.NewRequest("GET", "http://example.com?url=https://example.com/testing", nil)
		err := r.Parse(test.txtRecord, req, c)

		if err != nil {
			if test.err == nil || !strings.HasPrefix(err.Error(), test.err.Error()) {
				t.Errorf("Test %d: Unexpected error: %s", i, err)
			}
			continue
		}
		if err == nil && test.err != nil {
			t.Errorf("Test %d: Expected error, got nil", i)
			continue
		}

		if got, want := r.Version, test.expected.Version; got != want {
			t.Errorf("Test %d: Expected Version to be '%s', got '%s'", i, want, got)
		}
		if got, want := r.To, test.expected.To; got != want {
			t.Errorf("Test %d: Expected To to be '%s', got '%s'", i, want, got)
		}
		if got, want := r.Code, test.expected.Code; got != want {
			t.Errorf("Test %d: Expected Code to be '%d', got '%d'", i, want, got)
		}
		if got, want := r.Type, test.expected.Type; got != want {
			t.Errorf("Test %d: Expected Type to be '%s', got '%s'", i, want, got)
		}
		if got, want := r.Vcs, test.expected.Vcs; got != want {
			t.Errorf("Test %d: Expected Vcs to be '%s', got '%s'", i, want, got)
		}
	}
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
		txt      string
		expected string
		enable   []string
	}{
		{
			"https://host.e2e.test",
			txts["_redirect.host.e2e.test"],
			"https://plain.host.test",
			[]string{"host"},
		},
		{
			"https://nocode.host.e2e.test",
			txts["_redirect.nocode.host.e2e.test."],
			"https://nocode.host.test",
			[]string{"host"},
		},
		{
			"https://noversion.host.e2e.test",
			txts["_redirect.noversion.host.e2e.test."],
			"https://noversion.host.test",
			[]string{"host"},
		},
		{
			"https://noto.host.e2e.test",
			txts["_redirect.noto.host.e2e.test."],
			"",
			[]string{"host"},
		},
		{
			"https://path.e2e.test/",
			txts["_redirect.path.e2e.test."],
			"https://root.fallback.test",
			[]string{"path", "host"},
		},
		{
			"https://path.e2e.test/nocode",
			txts["_redirect.nocode.path.e2e.test."],
			"https://nocode.fallback.path.test",
			[]string{"path", "host"},
		},
		{
			"https://path.e2e.test/noversion",
			txts["_redirect.noversion.path.e2e.test."],
			"https://fallback.path.test",
			[]string{"path", "host"},
		},
		{
			"https://path.e2e.test/noto",
			txts["_redirect.noto.path.e2e.test."],
			"",
			[]string{"path", "host"},
		},
		{
			"https://path.e2e.test/noroot",
			txts["_redirect.noroot.path.e2e.test."],
			"https://fallback.path.test",
			[]string{"path", "host"},
		},
		{
			"https://pkg.txtdirect.test",
			txts["_redirect.pkg.txtdirect.test."],
			"https://github.com/txtdirect/txtdirect",
			[]string{"gometa"},
		},
		{
			"https://metapath.e2e.test/pkg",
			txts["_redirect.metapath.e2e.test."],
			"https://github.com/okkur/reposeed-server",
			[]string{"gometa", "path"},
		},
		{
			"https://metapath.e2e.test/pkg/second",
			txts["_redirect.metapath.e2e.test."],
			"https://github.com/okkur/reposeed",
			[]string{"gometa", "path"},
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
