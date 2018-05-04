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
	"net/http"
	"strings"
)

func redirectPath(w http.ResponseWriter, req *http.Request, r record, host, path string) error {
	pathSlice := strings.Split(path, "/")
	url := basezone + "." + pathSlice[1] + "." + host
	http.Redirect(w, req, url, 302)
	return nil
}
