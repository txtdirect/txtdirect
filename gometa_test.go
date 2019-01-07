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
<html>
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
<html>
<head>
<meta name="go-import" content="empty.com/test git ">

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
<html>
<head>
<meta name="go-import" content="root.com git redirect.com/my-root-package">

</head>
</html>`,
		},
		{
			host: "root.com",
			path: "/",
			record: record{
				Vcs: "git",
				To:  "github.com/txtdirect/txtdirect",
			},
			expected: `<!DOCTYPE html>
<html>
<head>
<meta name="go-import" content="root.com git github.com/txtdirect/txtdirect">
<meta name="go-source" content="root.com _ github.com/txtdirect/txtdirect/tree/v2{/dir} github.com/txtdirect/txtdirect/blob/v2{/dir}/{file}#L{line}">
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

func TestInternalFolderInPath(t *testing.T) {
	rec := httptest.NewRecorder()
	err := gometa(rec, record{}, "example.com", "/test/internal")
	if err == nil {
		t.Errorf("Expected to get an error when '/internal' folder included in path")
	}
}
