package caddy

import (
	"testing"

	"github.com/mholt/caddy"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input     string
		shouldErr bool
		expected  []string
	}{
		{
			`
			txtdirect {
				wrong keyword
			}
			`,
			true,
			[]string{},
		},
		{
			`
			txtdirect {
				enable
			}
			`,
			true,
			[]string{},
		},
		{
			`
			txtdirect {
				enable this
				disable that
			}
			`,
			true,
			[]string{},
		},
		{
			`txtdirect`,
			false,
			[]string{"host", "gometa"},
		},
		{
			`
			txtdirect {
				enable host
			}
			`,
			false,
			[]string{"host"},
		},
		{
			`
			txtdirect {
				disable host
			}
			`,
			false,
			[]string{"gometa"},
		},
	}

	for i, test := range tests {
		c := caddy.NewTestController("http", test.input)
		_, err := parse(c)
		if !test.shouldErr && err != nil {
			t.Errorf("Test %d: Unexpected error %s", i, err)
			continue
		}
		if test.shouldErr && err == nil {
			t.Errorf("Test %d: Expected error", i)
		}
	}
}
