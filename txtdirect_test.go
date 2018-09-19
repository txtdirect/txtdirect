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
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/miekg/dns"
)

// Testing TXT records
var txts = map[string]string{
	"_redirect.about.txtdirect.": "v=txtv0;to=https://about.txtdirect.org",
	"_redirect.pkg.txtdirect.":   "v=txtv0;to=https://pkg.txtdirect.org;type=gometa",
}

// Testing DNS server port
const port = 6000

// Initialize dns server instance
var server = &dns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}

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
				Code:    301,
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
				Code:    301,
				Type:    "host",
			},
			nil,
		},
		{
			"v=txtv0;to={?url}",
			record{
				Version: "txtv0",
				To:      "https://example.com/testing",
				Code:    301,
				Type:    "host",
			},
			nil,
		},
		{
			"v=txtv0;to={?url};from={method}",
			record{
				Version: "txtv0",
				To:      "https://example.com/testing",
				Code:    301,
				Type:    "host",
				From:    "GET",
			},
			nil,
		},
	}

	for i, test := range tests {
		r := record{}
		req, _ := http.NewRequest("GET", "http://example.com?url=https://example.com/testing", nil)
		err := r.Parse(test.txtRecord, req)

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

/*
DNS TXT records currently registered at _td.test.txtdirect.org available in:
https://raw.githubusercontent.com/txtdirect/_test-records/master/test.txtdirect.org
*/
func TestRedirectDefault(t *testing.T) {
	testURL := "https://%d._td.test.txtdirect.org"
	dnsURL := "_redirect.%d._td.test.txtdirect.org"

	config := Config{
		Enable: []string{"host"},
	}

	for i := 0; ; i++ {
		_, err := net.LookupTXT(fmt.Sprintf(dnsURL, i))
		if err != nil {
			break
		}
		req, _ := http.NewRequest("GET", fmt.Sprintf(testURL, i), nil)
		rec := httptest.NewRecorder()
		err = Redirect(rec, req, config)
		if err != nil {
			t.Errorf("test %d: Unexpected error: %s", i, err)
		}
	}
}

/*
DNS TXT records currently registered at _ths.test.txtdirect.org available in:
https://raw.githubusercontent.com/txtdirect/_test-records/master/test.txtdirect.org
*/
func TestRedirectSuccess(t *testing.T) {
	testURL := "https://%d._ths.test.txtdirect.org"
	dnsURL := "_redirect.%d._ths.test.txtdirect.org"

	config := Config{
		Enable: []string{"host", "gometa"},
	}

	for i := 0; ; i++ {
		_, err := net.LookupTXT(fmt.Sprintf(dnsURL, i))
		if err != nil {
			break
		}
		req, _ := http.NewRequest("GET", fmt.Sprintf(testURL, i), nil)
		rec := httptest.NewRecorder()
		err = Redirect(rec, req, config)
		if err != nil {
			t.Errorf("test %d: Unexpected error: %s", i, err)
		}
	}
}

/*
DNS TXT records currently registered at _thf.test.txtdirect.org available in:
https://raw.githubusercontent.com/txtdirect/_test-records/master/test.txtdirect.org
*/
func TestRedirectFailure(t *testing.T) {
	testURL := "https://%d._thf.test.txtdirect.org"
	dnsURL := "_redirect.%d._thf.test.txtdirect.org"

	config := Config{
		Enable: []string{"host"},
	}

	for i := 0; ; i++ {
		_, err := net.LookupTXT(fmt.Sprintf(dnsURL, i))
		if err != nil {
			break
		}
		req, _ := http.NewRequest("GET", fmt.Sprintf(testURL, i), nil)
		rec := httptest.NewRecorder()
		err = Redirect(rec, req, config)
		if err == nil {
			t.Errorf("test %d: Expected error, got nil", i)
		}
	}
}

/*
DNS TXT records currently registered at _ths.test.txtdirect.org available in:
https://raw.githubusercontent.com/txtdirect/_test-records/master/test.txtdirect.org
*/
func TestPathBasedRoutingRedirect(t *testing.T) {
	config := Config{
		Enable: []string{"path"},
	}
	req := httptest.NewRequest("GET", "https://pkg.txtdirect.com/caddy/v1/", nil)
	w := httptest.NewRecorder()

	err := Redirect(w, req, config)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
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

func TestParsePlaceholders(t *testing.T) {
	tests := []struct {
		url         string
		placeholder string
		expected    string
	}{
		{
			"example.com{uri}",
			"/?test=test",
			"example.com/?test=test",
		},
		{
			"example.com{uri}/{~test}",
			"/?test=test",
			"example.com/?test=test/test",
		},
		{
			"example.com{uri}/{>Test}",
			"/?test=test",
			"example.com/?test=test/test-header",
		},
		{
			"example.com{uri}/{?test}",
			"/?test=test",
			"example.com/?test=test/test",
		},
	}
	for _, test := range tests {
		req := httptest.NewRequest("GET", "https://example.com"+test.placeholder, nil)
		req.AddCookie(&http.Cookie{Name: "test", Value: "test"})
		req.Header.Add("Test", "test-header")
		result := parsePlaceholders(test.url, req)
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func Test_query(t *testing.T) {
	tests := []struct {
		zone string
		txt  string
	}{
		{
			"_redirect.about.txtdirect.",
			txts["_redirect.about.txtdirect."],
		},
		{
			"_redirect.pkg.txtdirect.",
			txts["_redirect.pkg.txtdirect."],
		},
	}
	go RunDNSServer()
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
	dns.HandleFunc("txtdirect.", handleDNSRequest)
	err := server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}
}
