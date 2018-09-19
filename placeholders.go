package txtdirect

import (
	"net/http"
	"net/url"
	"path"
	"strings"
)

func parsePlaceholders(input string, r *http.Request) string {
	placeholders := PlaceholderRegex.FindAllStringSubmatch(input, -1)
	for _, placeholder := range placeholders {
		switch placeholder[0] {
		case "{uri}":
			input = strings.Replace(input, "{uri}", r.URL.RequestURI(), -1)
		case "{dir}":
			dir, _ := path.Split(r.URL.Path)
			input = strings.Replace(input, "{dir}", dir, -1)
		case "{file}":
			_, file := path.Split(r.URL.Path)
			input = strings.Replace(input, "{file}", file, -1)
		case "{fragment}":
			input = strings.Replace(input, "{fragment}", r.URL.Fragment, -1)
		case "{host}":
			input = strings.Replace(input, "{host}", r.URL.Host, -1)
		case "{hostonly}":
			input = strings.Replace(input, "{hostonly}", r.URL.Hostname(), -1)
		case "{method}":
			input = strings.Replace(input, "{method}", r.Method, -1)
		case "{path}":
			input = strings.Replace(input, "{path}", r.URL.Path, -1)
		case "{path_escaped}":
			input = strings.Replace(input, "{path_escaped}", url.QueryEscape(r.URL.Path), -1)
		case "{port}":
			input = strings.Replace(input, "{port}", r.URL.Port(), -1)
		case "{query}":
			input = strings.Replace(input, "{query}", r.URL.RawQuery, -1)
		case "{query_escaped}":
			input = strings.Replace(input, "{query_escaped}", url.QueryEscape(r.URL.RawQuery), -1)
		case "{uri_escaped}":
			input = strings.Replace(input, "{uri_escaped}", url.QueryEscape(r.URL.RequestURI()), -1)
		case "{user}":
			user, _, ok := r.BasicAuth()
			if !ok {
				input = strings.Replace(input, "{user}", "", -1)
			}
			input = strings.Replace(input, "{user}", user, -1)
		}
		if placeholder[0][1] == '>' {
			want := placeholder[0][2 : len(placeholder[0])-1]
			for key, values := range r.Header {
				// Header placeholders (case-insensitive)
				if strings.EqualFold(key, want) {
					input = strings.Replace(input, placeholder[0], strings.Join(values, ","), -1)
				}
			}
		}
		if placeholder[0][1] == '~' {
			name := placeholder[0][2 : len(placeholder[0])-1]
			if cookie, err := r.Cookie(name); err == nil {
				input = strings.Replace(input, placeholder[0], cookie.Value, -1)
			}
		}
		if placeholder[0][1] == '?' {
			query := r.URL.Query()
			name := placeholder[0][2 : len(placeholder[0])-1]
			input = strings.Replace(input, placeholder[0], query.Get(name), -1)
		}
	}
	return input
}
