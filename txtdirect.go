package txtdirect

import (
	"fmt"
	"log"
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
			if r.Version != "txtv0" {
				return r, fmt.Errorf("unhandled version '%s'", r.Version)
			}
			log.Print("WARN: txtv0 is not suitable for production")

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
