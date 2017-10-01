package txtdirect

import (
	"io/ioutil"
	"net/http/httptest"
	"testing"
)

func TestGometa(t *testing.T) {
	tests := []struct {
		host     string
		path     string
		record   record
		expected string
	}{
		{
			host: "example.com",
			path: "/test",
			record: record{
				Vcs: "git",
				To:  "redirect.com/my-go-pkg",
			},
			expected: `<!DOCTYPE html>
<head>
<meta name="go-import" content="example.com/test git redirect.com/my-go-pkg">
</head>
</html>`,
		},
		{
			host:   "empty.com",
			path:   "/test",
			record: record{},
			expected: `<!DOCTYPE html>
<head>
<meta name="go-import" content="empty.com/test  ">
</head>
</html>`,
		},
		{
			host: "root.com",
			path: "/",
			record: record{
				Vcs: "git",
				To:  "redirect.com/my-root-package",
			},
			expected: `<!DOCTYPE html>
<head>
<meta name="go-import" content="root.com git redirect.com/my-root-package">
</head>
</html>`,
		},
	}

	for i, test := range tests {
		rec := httptest.NewRecorder()
		err := gometa(rec, test.record, test.host, test.path)
		if err != nil {
			t.Errorf("Test %d: Unexpected error: %s", i, err)
			continue
		}
		txt, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			t.Errorf("Test %d: Unexpected error: %s", i, err)
			continue
		}
		if got, want := string(txt), test.expected; got != want {
			t.Errorf("Test %d:\nExpected\n%s\nto be:\n%s", i, got, want)
		}
	}
}
