package txtdirect

import (
	"fmt"
	"log"
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
			},
			nil,
		},
		{
			"v=txtv0;to=https://example.com/",
			record{
				Version: "txtv0",
				To:      "https://example.com/",
				Code:    301,
			},
			nil,
		},
		{
			"v=txtv0;https://example.com/;code=302",
			record{
				Version: "txtv0",
				To:      "https://example.com/",
				Code:    302,
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
	}
}

func TestHandleDefault(t *testing.T) {
	testURL := "https://%d._td.txtdirect.org"
	dnsURL := "_redirect.%d._td.txtdirect.org"

	for i := 0; ; i++ {
		_, err := net.LookupTXT(fmt.Sprintf(dnsURL, i))
		if err != nil {
			break
		}
		req, _ := http.NewRequest("GET", testURL, nil)
		rec := httptest.NewRecorder()
		err = Redirect(rec, req)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
	}
}

func TestHandleSuccess(t *testing.T) {
	testURL := "https://%d._ths.txtdirect.org"
	dnsURL := "_redirect.%d._ths.txtdirect.org"

	for i := 0; ; i++ {
		_, err := net.LookupTXT(fmt.Sprintf(dnsURL, i))
		if err != nil {
			break
		}
		req, _ := http.NewRequest("GET", fmt.Sprintf(testURL, i), nil)
		rec := httptest.NewRecorder()
		err = Redirect(rec, req)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
	}
}

func TestHandleFailure(t *testing.T) {
	testURL := "https://%d._thf.txtdirect.org"
	dnsURL := "_redirect.%d._thf.txtdirect.org"

	for i := 0; ; i++ {
		_, err := net.LookupTXT(fmt.Sprintf(dnsURL, i))
		if err != nil {
			log.Print(err, i)
			break
		}
		req, _ := http.NewRequest("GET", fmt.Sprintf(testURL, i), nil)
		rec := httptest.NewRecorder()
		err = Redirect(rec, req)
		if err == nil {
			t.Errorf("Expected error, got nil)")
		}
	}
}
