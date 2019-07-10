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
	"io"
	"net/http"
	"sync"
)

var bufferPool = sync.Pool{New: createBuffer}

func createBuffer() interface{} {
	return make([]byte, 0, 32*1024)
}

var skipHeaders = map[string]struct{}{
	"Content-Type":        {},
	"Content-Disposition": {},
	"Accept-Ranges":       {},
	"Set-Cookie":          {},
	"Cache-Control":       {},
	"Expires":             {},
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		if _, ok := dst[k]; ok {
			if _, shouldSkip := skipHeaders[k]; shouldSkip {
				continue
			}
			if k != "Server" {
				dst.Del(k)
			}
		}
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func pooledIoCopy(dst io.Writer, src io.Reader) {
	buf := bufferPool.Get().([]byte)
	defer bufferPool.Put(buf)

	bufCap := cap(buf)
	io.CopyBuffer(dst, src, buf[0:bufCap:bufCap])
}
