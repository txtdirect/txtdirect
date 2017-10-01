package txtdirect

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
				Vcs:     "git",
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
				Vcs:     "git",
				Type:    "host",
			},
			nil,
		},
		{
			"v=txtv0;https://example.com/;code=302",
			record{
				Version: "txtv0",
				To:      "https://example.com/",
				Code:    302,
				Vcs:     "git",
				Type:    "host",
			},
			nil,
		},
		{
			"v=txtv0;https://example.com/;code=302;vcs=hg",
			record{
				Version: "txtv0",
				To:      "https://example.com/",
				Code:    302,
				Vcs:     "hg",
				Type:    "host",
			},
			nil,
		},
		{
			"v=txtv0;https://example.com/;code=302;type=gometa",
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
			"v=txtv0;https://example.com/;https://google.com;code=test",
			record{},
			fmt.Errorf("multiple values without keys"),
		},
		{
			"v=txtv1;to=https://example.com/;code=test",
			record{},
			fmt.Errorf("unhandled version 'txtv1'"),
		},
	}

	for i, test := range tests {
		r := record{}
		err := r.Parse(test.txtRecord)

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

	for i := 0; ; i++ {
		_, err := net.LookupTXT(fmt.Sprintf(dnsURL, i))
		if err != nil {
			break
		}
		req, _ := http.NewRequest("GET", fmt.Sprintf(testURL, i), nil)
		rec := httptest.NewRecorder()
		err = Redirect(rec, req, []string{"host"})
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
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

	for i := 0; ; i++ {
		_, err := net.LookupTXT(fmt.Sprintf(dnsURL, i))
		if err != nil {
			break
		}
		req, _ := http.NewRequest("GET", fmt.Sprintf(testURL, i), nil)
		rec := httptest.NewRecorder()
		err = Redirect(rec, req, []string{"host", "gometa"})
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
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

	for i := 0; ; i++ {
		_, err := net.LookupTXT(fmt.Sprintf(dnsURL, i))
		if err != nil {
			break
		}
		req, _ := http.NewRequest("GET", fmt.Sprintf(testURL, i), nil)
		rec := httptest.NewRecorder()
		err = Redirect(rec, req, []string{"host"})
		if err == nil {
			t.Errorf("Expected error, got nil)")
		}
	}
}
