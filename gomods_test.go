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
	"net/http/httptest"
	"os"
	"testing"
)

func Test_gomods(t *testing.T) {
	tests := []struct {
		host     string
		path     string
		expected string
	}{
		{
			path: "/github.com/okkur/reposeed-server/@v/list",
		},
		{
			path: "/github.com/okkur/reposeed-server/@v/v0.1.0.info",
		},
		{
			path: "/github.com/okkur/reposeed-server/@v/v0.1.0.mod",
		},
		{
			path: "/github.com/okkur/reposeed-server/@v/v0.1.0.zip",
		},
	}
	for _, test := range tests {
		if err := os.MkdirAll("/tmp/gomods", os.ModePerm); err != nil {
			t.Fatal("Couldn't create storage directory (/tmp/gomods)")
		}
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", fmt.Sprintf("https://example.com%s", test.path), nil)
		c := Config{
			Gomods: Gomods{
				Enable:   true,
				Workers:  2,
				GoBinary: os.Getenv("GOROOT") + "/bin/go",
				Cache: Cache{
					Type: "local",
					Path: "/tmp/gomods",
				},
			},
		}
		c.Gomods.SetDefaults()
		err := gomods(w, r, test.path, c)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
	}
}
