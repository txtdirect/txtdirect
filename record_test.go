/*
Copyright 2019 - The TXTDirect Authors
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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseRecord(t *testing.T) {
	tests := []struct {
		txtRecord string
		expected  record
		err       error
		status    int
	}{
		{
			txtRecord: "v=txtv0;to=https://example.com/;code=302",
			expected: record{
				Version: "txtv0",
				To:      "https://example.com/",
				Code:    302,
				Type:    "host",
			},
			err: nil,
		},
		{
			txtRecord: "v=txtv0;to=https://example.com/",
			expected: record{
				Version: "txtv0",
				To:      "https://example.com/",
				Code:    302,
				Type:    "host",
			},
			err: nil,
		},
		{
			txtRecord: "v=txtv0;to=https://example.com/;code=302",
			expected: record{
				Version: "txtv0",
				To:      "https://example.com/",
				Code:    302,
				Type:    "host",
			},
			err: nil,
		},
		{
			txtRecord: "v=txtv0;to=https://example.com/;code=302;vcs=hg;type=gometa",
			expected: record{
				Version: "txtv0",
				To:      "https://example.com/",
				Code:    302,
				Vcs:     "hg",
				Type:    "gometa",
			},
			err: nil,
		},
		{
			txtRecord: "v=txtv0;to=https://example.com/;code=302;type=gometa;vcs=git",
			expected: record{
				Version: "txtv0",
				To:      "https://example.com/",
				Code:    302,
				Vcs:     "git",
				Type:    "gometa",
			},
			err: nil,
		},
		{
			txtRecord: "v=txtv0;to=https://example.com/;code=test",
			expected:  record{},
			err:       fmt.Errorf("could not parse status code"),
		},
		{
			txtRecord: "v=txtv1;to=https://example.com/;code=test",
			expected:  record{},
			err:       fmt.Errorf("unhandled version 'txtv1'"),
		},
		{
			txtRecord: "v=txtv0;https://example.com/",
			expected:  record{},
			err:       fmt.Errorf("arbitrary data not allowed"),
		},
		{
			txtRecord: "v=txtv0;to=https://example.com/caddy;type=path;code=302",
			expected: record{
				Version: "txtv0",
				To:      "https://example.com/caddy",
				Type:    "path",
				Code:    302,
			},
			err: nil,
		},
		{
			txtRecord: "v=txtv0;to=https://example.com/;key=value",
			expected: record{
				Version: "txtv0",
				To:      "https://example.com/",
				Code:    302,
				Type:    "host",
			},
			err: nil,
		},
		{
			txtRecord: "v=txtv0;to={?url}",
			expected: record{
				Version: "txtv0",
				To:      "https://example.com/testing",
				Code:    302,
				Type:    "host",
			},
			err: nil,
		},
		{
			txtRecord: "v=txtv0;to={?url};from={method}",
			expected: record{
				Version: "txtv0",
				To:      "https://example.com/testing",
				Code:    302,
				Type:    "host",
				From:    "GET",
			},
			err: nil,
		},
		{
			txtRecord: "v=txtv0;ref=true;code=302",
			expected: record{
				Version: "txtv0",
				Type:    "host",
				Code:    302,
				Ref:     true,
			},
			status: 404,
		},
		{
			txtRecord: "v=txtv0;ref=false;code=302",
			expected: record{
				Version: "txtv0",
				Type:    "host",
				Code:    302,
				Ref:     false,
			},
			status: 404,
		},
		{
			txtRecord: "v=txtv0;type=path;code=302;>Header-1=HeaderValue",
			expected: record{
				Version: "txtv0",
				Type:    "path",
				Code:    302,
				Ref:     false,
				Headers: map[string]string{"Header-1": "HeaderValue"},
			},
		},
		{
			txtRecord: "v=txtv0;type=path;code=302;>Header-1=HeaderValue;>Header-2=HeaderValue",
			expected: record{
				Version: "txtv0",
				Type:    "path",
				Code:    302,
				Ref:     false,
				Headers: map[string]string{
					"Header-1": "HeaderValue",
					"Header-2": "HeaderValue",
				},
			},
		},
		{
			// Commonly used http headers used for better real world testing.
			// Taken from: https://www.whitehatsec.com/blog/list-of-http-response-headers/
			txtRecord: `v=txtv0;type=path;code=302;
						>Access-Control-Allow-Credentials=true;
						>Access-Control-Allow-Headers=X-PINGOTHER;
						>Access-Control-Allow-Methods=PUT%2C%20DELETE%2C%20XMODIFY;
						>Access-Control-Allow-Origin=http%3A%2F%2Fexample.org;
						>Access-Control-Expose-Headers=X-My-Custom-Header%2C%20X-Another-Custom-Header;
						>Access-Control-Max-Age=2520;
						>Accept-Ranges=bytes;
						>Age=12;
						>Allow=GET%2C%20HEAD%2C%20POST%2C%20OPTIONS;
						>Alternate-Protocol=443%3Anpn-spdy%2F2%2C443%3Anpn-spdy%2F2;
						>Cache-Control=private%2C%20no-cache%2C%20must-revalidate;
						>Client-Date=Tue%2C%2027%20Jan%202009%2018%3A17%3A30%20GMT;
						>Client-Peer=123.123.123.123%3A80;
						>Client-Response-Num=1;
						>Connection=Keep-Alive;
						>Content-Disposition=attachment%3B%20filename%3D%22example.exe%22;
						>Content-Encoding=gzip;
						>Content-Language=en;
						>Content-Length=1329;
						>Content-Location=%2Findex.htm;
						>Content-MD5=Q2hlY2sgSW50ZWdyaXR5IQ%3D%3D;
						>Content-Range=bytes%2021010-47021%2F47022;
						>Content-Security-Policy=default-src%20%E2%80%98self%E2%80%99;
						>Content-Security-Policy-Report-Only=default-src%20%E2%80%98self%E2%80%99%3B%20%E2%80%A6%3B%20report-uri%20%2Fcsp_report_parser%3B;
						>Content-Type=text%2Fhtml;
						>Date=Fri%2C%2022%20Jan%202010%2004%3A00%3A00%20GMT;
						>ETag=737060cd8c284d8af7ad3082f209582d;
						>Expires=Mon%2C%2026%20Jul%201997%2005%3A00%3A00%20GMT;
						>HTTP=%2F1.1%20401%20Unauthorized;
						>Keep-Alive=timeout%3D3%2C%20max%3D87;
						>Last-Modified=Tue%2C%2015%20Nov%201994%2012%3A45%3A26%20%2B0000;
						>Link=%3Chttp%3A%2F%2Fwww.example.com%2F%3E%3B%20rel%3D%22cononical%22;
						>Location=http%3A%2F%2Fwww.example.com%2F;
						>P3P=policyref%3D%22http%3A%2F%2Fwww.example.com%2Fw3c%2Fp3p.xml%22%2C%20CP%3D%22NOI%20DSP%20COR%20ADMa%20OUR%20NOR%20STA%22;
						>Pragma=no-cache;
						>Proxy-Authenticate=Basic;
						>Proxy-Connection=Keep-Alive;
						>Refresh=5%3B%20url%3Dhttp%3A%2F%2Fwww.example.com%2F;
						>Retry-After=120;
						>Server=Apache;
						>Set-Cookie=test%3D1%3B%20domain%3Dexample.com%3B%20path%3D%2F%3B%20expires%3DTue%2C%2001-Oct-2013%2019%3A16%3A48%20GMT;
						>Status=200%20OK;
						>Strict-Transport-Security=max-age%3D16070400%3B%20includeSubDomains%3B%20preload;
						>Timing-Allow-Origin=www.example.com;
						>Trailer=Max-Forwards;
						>Transfer-Encoding=chunked;
						>Upgrade=HTTP%2F2.0%2C%20SHTTP%2F1.3%2C%20IRC%2F6.9%2C%20RTA%2Fx11;
						>Vary=%2A;
						>Via=1.0%20fred%2C%201.1%20example.com%20%28Apache%2F1.1%29;
						>Warning=Warning%3A%20199%20Miscellaneous%20warning;
						>WWW-Authenticate=Basic;
						>X-Aspnet-Version=2.0.50727;
						>X-Content-Type-Options=nosniff;
						>X-Frame-Options=deny;
						>X-Permitted-Cross-Domain-Policies=master-only;
						>X-Pingback=http%3A%2F%2Fwww.example.com%2Fpingback%2Fxmlrpc;
						>X-Powered-By=PHP%2F5.4.0;
						>X-Robots-Tag=noindex%2Cnofollow;
						>X-UA-Compatible=Chome%3D1;
						>X-XSS-Protection=1%3B%20mode%3Dblock`,
			expected: record{
				Version: "txtv0",
				Type:    "path",
				Code:    302,
				Ref:     false,
				Headers: map[string]string{
					"Access-Control-Allow-Credentials":    "true",
					"Access-Control-Allow-Headers":        "X-PINGOTHER",
					"Access-Control-Allow-Methods":        "PUT, DELETE, XMODIFY",
					"Access-Control-Allow-Origin":         "http://example.org",
					"Access-Control-Expose-Headers":       "X-My-Custom-Header, X-Another-Custom-Header",
					"Access-Control-Max-Age":              "2520",
					"Accept-Ranges":                       "bytes",
					"Age":                                 "12",
					"Allow":                               "GET, HEAD, POST, OPTIONS",
					"Alternate-Protocol":                  "443:npn-spdy/2,443:npn-spdy/2",
					"Cache-Control":                       "private, no-cache, must-revalidate",
					"Client-Date":                         "Tue, 27 Jan 2009 18:17:30 GMT",
					"Client-Peer":                         "123.123.123.123:80",
					"Client-Response-Num":                 "1",
					"Connection":                          "Keep-Alive",
					"Content-Disposition":                 "attachment; filename=\"example.exe\"",
					"Content-Encoding":                    "gzip",
					"Content-Language":                    "en",
					"Content-Length":                      "1329",
					"Content-Location":                    "/index.htm",
					"Content-MD5":                         "Q2hlY2sgSW50ZWdyaXR5IQ==",
					"Content-Range":                       "bytes 21010-47021/47022",
					"Content-Security-Policy":             "default-src ‘self’",
					"Content-Security-Policy-Report-Only": "default-src ‘self’; …; report-uri /csp_report_parser;",
					"Content-Type":                        "text/html",
					"Date":                                "Fri, 22 Jan 2010 04:00:00 GMT",
					"ETag":                                "737060cd8c284d8af7ad3082f209582d",
					"Expires":                             "Mon, 26 Jul 1997 05:00:00 GMT",
					"HTTP":                                "/1.1 401 Unauthorized",
					"Keep-Alive":                          "timeout=3, max=87",
					"Last-Modified":                       "Tue, 15 Nov 1994 12:45:26 +0000",
					"Link":                                "<http://www.example.com/>; rel=\"cononical\"",
					"Location":                            "http://www.example.com/",
					"P3P":                                 "policyref=\"http://www.example.com/w3c/p3p.xml\", CP=\"NOI DSP COR ADMa OUR NOR STA\"",
					"Pragma":                              "no-cache",
					"Proxy-Authenticate":                  "Basic",
					"Proxy-Connection":                    "Keep-Alive",
					"Refresh":                             "5; url=http://www.example.com/",
					"Retry-After":                         "120",
					"Server":                              "Apache",
					"Set-Cookie":                          "test=1; domain=example.com; path=/; expires=Tue, 01-Oct-2013 19:16:48 GMT",
					"Status":                              "200 OK",
					"Strict-Transport-Security":           "max-age=16070400; includeSubDomains; preload",
					"Timing-Allow-Origin":                 "www.example.com",
					"Trailer":                             "Max-Forwards",
					"Transfer-Encoding":                   "chunked",
					"Upgrade":                             "HTTP/2.0, SHTTP/1.3, IRC/6.9, RTA/x11",
					"Vary":                                "*",
					"Via":                                 "1.0 fred, 1.1 example.com (Apache/1.1)",
					"Warning":                             "Warning: 199 Miscellaneous warning",
					"WWW-Authenticate":                    "Basic",
					"X-Aspnet-Version":                    "2.0.50727",
					"X-Content-Type-Options":              "nosniff",
					"X-Frame-Options":                     "deny",
					"X-Permitted-Cross-Domain-Policies":   "master-only",
					"X-Pingback":                          "http://www.example.com/pingback/xmlrpc",
					"X-Powered-By":                        "PHP/5.4.0",
					"X-Robots-Tag":                        "noindex,nofollow",
					"X-UA-Compatible":                     "Chome=1",
					"X-XSS-Protection":                    "1; mode=block",
				},
			},
		},
	}

	for i, test := range tests {
		c := Config{
			Enable: []string{test.expected.Type},
		}
		req, _ := http.NewRequest("GET", "http://example.com?url=https://example.com/testing", nil)
		w := httptest.NewRecorder()
		r, err := ParseRecord(test.txtRecord, w, req, c)

		if err != nil {
			if test.err == nil || !strings.HasPrefix(err.Error(), test.err.Error()) {
				t.Errorf("Test %d: Unexpected error: %s", i, err)
			}
			continue
		}

		if test.status != 0 {
			if w.Result().StatusCode != test.status {
				t.Errorf("Test %d: Expected status code %d, got %d", i, test.status, w.Result().StatusCode)
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

		if len(r.Headers) != len(test.expected.Headers) {
			t.Errorf("Test %d: Expected %d headers, got '%d'", i, len(r.Headers), len(test.expected.Headers))
		}

		for header, val := range r.Headers {
			if test.expected.Headers[header] != val {
				t.Errorf("Test %d: Expected %s Header to be '%s', got '%s'",
					i, header, test.expected.Headers[header], val)
			}
		}
	}
}
