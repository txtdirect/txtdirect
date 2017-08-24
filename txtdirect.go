package txtdirect

import (
	"fmt"
	"strconv"
	"strings"
)

type record struct {
	Version string
	To      string
	Code    int
}

func (r record) Parse(str string) (record, error) {
	s := strings.Split(str, ";")
	for _, l := range s {
		switch {
		case strings.HasPrefix(l, "v="):
			l = strings.TrimPrefix(l, "v=")
			r.Version = l
		case strings.HasPrefix(l, "to="):
			l = strings.TrimPrefix(l, "to=")
			r.To = l
		case strings.HasPrefix(l, "code="):
			l = strings.TrimPrefix(l, "code=")
			i, err := strconv.Atoi(l)
			if err != nil {
				return r, fmt.Errorf("could not parse status code: %s", err)
			}
			r.Code = i
		}
	}
	return r, nil
}
