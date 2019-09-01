/*
Copyright 2017 - The TXTDirect Authors
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
	"log"
	"os"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/caddyserver/caddy"
)

func TestCaddyParse(t *testing.T) {
	tests := []struct {
		input     string
		shouldErr bool
		expected  Config
	}{
		{
			`
			txtdirect {
				wrong keyword
			}
			`,
			true,
			Config{},
		},
		{
			`
			txtdirect {
				enable
			}
			`,
			true,
			Config{},
		},
		{
			`
			txtdirect {
				disable
			}
			`,
			true,
			Config{},
		},
		{
			`
			txtdirect {
				enable this
				disable that
			}
			`,
			true,
			Config{},
		},
		{
			`
			txtdirect {
				disable this
				enable that
			}
			`,
			true,
			Config{},
		},
		{
			`
			txtdirect {
				redirect
			}
			`,
			true,
			Config{},
		},
		{
			`txtdirect`,
			false,
			Config{
				Enable: allOptions,
			},
		},
		{
			`
			txtdirect {
				enable host
			}
			`,
			false,
			Config{
				Enable: []string{"host"},
			},
		},
		{
			`
			txtdirect {
				disable host
			}
			`,
			false,
			Config{
				Enable: []string{"path", "gometa", "www"},
			},
		},
		{
			`
			txtdirect {
				redirect https://example.com
			}
			`,
			false,
			Config{
				Redirect: "https://example.com",
				Enable:   allOptions,
			},
		},
		{
			`
			txtdirect {
				enable host
				redirect https://example.com
			}
			`,
			false,
			Config{
				Redirect: "https://example.com",
				Enable:   []string{"host"},
			},
		},
		{
			`
			txtdirect {
				enable path
				redirect https://example.com
			}
			`,
			false,
			Config{
				Redirect: "https://example.com",
				Enable:   []string{"path"},
			},
		},
		{
			`
			txtdirect {
				enable host
				redirect https://example.com
				prometheus
			}
			`,
			false,
			Config{
				Redirect: "https://example.com",
				Enable:   []string{"host"},
				Prometheus: Prometheus{
					Enable:  true,
					Address: "localhost:9183",
					Path:    "/metrics",
				},
			},
		},
		{
			`
			txtdirect {
				enable host
				redirect https://example.com
				prometheus {
					address localhost:6666
					path /metrics
				}
			}
			`,
			false,
			Config{
				Redirect: "https://example.com",
				Enable:   []string{"host"},
				Prometheus: Prometheus{
					Enable:  true,
					Address: "localhost:6666",
					Path:    "/metrics",
				},
			},
		},
		{
			`
			txtdirect {
				enable host gomods
				redirect https://example.com
				gomods
				resolver 127.0.0.1
			}
			`,
			false,
			Config{
				Redirect: "https://example.com",
				Enable:   []string{"host", "gomods"},
				Gomods: Gomods{
					Enable:   true,
					GoBinary: os.Getenv("GOROOT") + "/bin/go",
					Workers:  1,
				},
				Resolver: "127.0.0.1",
			},
		},
		{
			`
			txtdirect {
				enable host gomods
				redirect https://example.com
				gomods {
					gobinary /my/go/binary
					cache
				}
				resolver 127.0.0.1
			}
			`,
			false,
			Config{
				Redirect: "https://example.com",
				Enable:   []string{"host", "gomods"},
				Resolver: "127.0.0.1",
				Gomods: Gomods{
					Enable:   true,
					GoBinary: "/my/go/binary",
					Workers:  1,
					Cache: Cache{
						Enable: true,
						Type:   "tmp",
						Path:   "/tmp/txtdirect/gomods",
					},
				},
			},
		},
		{
			`
			txtdirect {
				enable host gomods
				redirect https://example.com
				gomods {
					gobinary /my/go/binary
					cache {
						type local
						path /my/cache/path
					}
				}
				resolver 127.0.0.1
			}
			`,
			false,
			Config{
				Redirect: "https://example.com",
				Enable:   []string{"host", "gomods"},
				Resolver: "127.0.0.1",
				Gomods: Gomods{
					Enable:   true,
					GoBinary: "/my/go/binary",
					Workers:  1,
					Cache: Cache{
						Enable: true,
						Type:   "local",
						Path:   "/my/cache/path",
					},
				},
			},
		},
		{
			`
			txtdirect {
				enable host gomods
				redirect https://example.com
				gomods {
					gobinary /my/go/binary
					cache {
						type local
						path /my/cache/path
					}
					workers 5
				}
				resolver 127.0.0.1
			}
			`,
			false,
			Config{
				Redirect: "https://example.com",
				Enable:   []string{"host", "gomods"},
				Resolver: "127.0.0.1",
				Gomods: Gomods{
					Enable:   true,
					GoBinary: "/my/go/binary",
					Cache: Cache{
						Enable: true,
						Type:   "local",
						Path:   "/my/cache/path",
					},
					Workers: 5,
				},
			},
		},
		{
			`
			txtdirect {
				enable host gomods
				logfile stderr
				resolver 127.0.0.1
			}
			`,
			false,
			Config{
				Enable:    []string{"host", "gomods"},
				Resolver:  "127.0.0.1",
				LogOutput: "stderr",
			},
		},
		{
			`
			txtdirect {
				enable host gomods
				logfile
				resolver 127.0.0.1
			}
			`,
			false,
			Config{
				Enable:    []string{"host", "gomods"},
				Resolver:  "127.0.0.1",
				LogOutput: "stdout",
			},
		},
		{
			`
			txtdirect {
				enable host
				redirect https://example.com
				qr {
					size 256
					background "#ffffff"
					foreground "#000000"
				}
				resolver 127.0.0.1
			}
			`,
			false,
			Config{
				Redirect: "https://example.com",
				Enable:   []string{"host"},
				Resolver: "127.0.0.1",
				Qr: Qr{
					Size:            256,
					BackgroundColor: "ffffffff",
					ForegroundColor: "000000ff",
				},
			},
		},
		{
			`
			txtdirect {
				enable host
				redirect https://example.com
				qr {
					size 256
					background "#ffffff"
					foreground "#000000"
					recovery_level 1
				}
				resolver 127.0.0.1
			}
			`,
			false,
			Config{
				Redirect: "https://example.com",
				Enable:   []string{"host"},
				Resolver: "127.0.0.1",
				Qr: Qr{
					Size:            256,
					BackgroundColor: "ffffffff",
					ForegroundColor: "000000ff",
					RecoveryLevel:   1,
				},
			},
		},
		{
			`
			txtdirect {
				enable host
				redirect https://example.com
				resolver 127.0.0.1
				qr
			}
			`,
			false,
			Config{
				Redirect: "https://example.com",
				Enable:   []string{"host"},
				Resolver: "127.0.0.1",
				Qr: Qr{
					Size:            256,
					BackgroundColor: "ffffffff",
					ForegroundColor: "00000000",
				},
			},
		},
	}

	for i, test := range tests {
		log.Println(log.Flags())
		c := caddy.NewTestController("http", test.input)
		conf, err := parse(c)
		if !test.shouldErr && err != nil {
			t.Errorf("Test %d: Unexpected error %s", i, err)
			continue
		}
		if test.shouldErr {
			if err == nil {
				t.Errorf("Test %d: Expected error", i)
			}
			continue
		}

		// Check configs for each enabled type
		for _, e := range conf.Enable {
			switch e {
			case "gomods":
				// Fs field gets filled by default when parsing the config
				test.expected.Gomods.Fs = conf.Gomods.Fs
				// Set the default cache path for expected config if cache type is tmp
				if conf.Gomods.Cache.Type == "tmp" {
					test.expected.Gomods.Cache.Path = afero.GetTempDir(test.expected.Gomods.Fs, "")
				}

				if conf.Gomods != test.expected.Gomods {
					t.Errorf("Expected %+v for gomods config got %+v", test.expected.Gomods, conf.Gomods)
				}
			case "qr":
				if conf.Qr != test.expected.Qr {
					t.Errorf("Expected %+v for qr config got %+v", test.expected.Qr, conf.Qr)
				}
			}
		}

		if test.expected.Prometheus.Enable == true {
			if conf.Prometheus != test.expected.Prometheus {
				t.Errorf("Expected %+v for prometheus config got %+v", test.expected.Prometheus, conf.Prometheus)
			}
		}

		if test.expected.Resolver != conf.Resolver {
			t.Errorf("Expected resolver to be %s, but got %s", test.expected.Resolver, conf.Resolver)
		}

		if test.expected.LogOutput != conf.LogOutput {
			t.Errorf("Expected log output to be %s, but got %s", test.expected.LogOutput, conf.LogOutput)
		}

		if !identical(conf.Enable, test.expected.Enable) {
			options := fmt.Sprintf("[ %s ]", strings.Join(conf.Enable, ", "))
			expected := fmt.Sprintf("[ %s ]", strings.Join(test.expected.Enable, ", "))
			t.Errorf("Test %d: Expected options %s, got %s", i, expected, options)
		}
	}
}

func identical(s1, s2 []string) bool {
	if s1 == nil {
		if s2 == nil {
			return true
		}
		return false
	}
	if s2 == nil {
		return false
	}

	if len(s1) != len(s2) {
		return false
	}

	for i := range s1 {
		found := false
		for j := range s2 {
			if s1[i] == s2[j] {
				found = true
			}
		}

		if !found {
			return false
		}
	}
	return true
}
