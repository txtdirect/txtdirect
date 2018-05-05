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
	"net/http/httptest"
	"testing"
)

func TestRedirectPath(t *testing.T) {
	req := httptest.NewRequest("GET", "https://example.com/caddy", nil)
	w := httptest.NewRecorder()
	r := record{
		Version: "txtv0",
		To:      "https://example.com/caddy",
		Type:    "path",
	}
	redirectPath(w, req, r, req.Host, req.URL.Path)
	if w.Code != 302 {
		t.Errorf("Expected %d, got %d", 302, w.Code)
	}
}
