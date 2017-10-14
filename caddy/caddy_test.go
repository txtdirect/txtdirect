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

package caddy

import (
	"fmt"
	"strings"
	"testing"

	"github.com/txtdirect/txtdirect"

	"github.com/mholt/caddy"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input     string
		shouldErr bool
		expected  txtdirect.Config
	}{
		{
			`
			txtdirect {
				wrong keyword
			}
			`,
			true,
			txtdirect.Config{},
		},
		{
			`
			txtdirect {
				enable
			}
			`,
			true,
			txtdirect.Config{},
		},
		{
			`
			txtdirect {
				disable
			}
			`,
			true,
			txtdirect.Config{},
		},
		{
			`
			txtdirect {
				enable this
				disable that
			}
			`,
			true,
			txtdirect.Config{},
		},
		{
			`
			txtdirect {
				disable this
				enable that
			}
			`,
			true,
			txtdirect.Config{},
		},
		{
			`
			txtdirect {
				redirect
			}
			`,
			true,
			txtdirect.Config{},
		},
		{
			`txtdirect`,
			false,
			txtdirect.Config{
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
			txtdirect.Config{
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
			txtdirect.Config{
				Enable: []string{"gometa", "www"},
			},
		},
		{
			`
			txtdirect {
				redirect https://example.com
			}
			`,
			false,
			txtdirect.Config{
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
			txtdirect.Config{
				Redirect: "https://example.com",
				Enable:   []string{"host"},
			},
		},
	}

	for i, test := range tests {
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
