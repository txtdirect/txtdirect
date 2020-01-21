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

	"go.txtdirect.org/txtdirect/config"
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
	}

	for i, test := range tests {
		c := config.Config{
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
