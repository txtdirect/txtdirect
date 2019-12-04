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
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type record struct {
	Version string
	To      string
	Code    int
	Type    string
	Vcs     string
	Website string
	From    string
	Root    string
	Re      string
	Ref     bool
}

// getRecord uses the given host to find a TXT record
// and then parses the txt record and returns a TXTDirect record
// struct instance. It returns an error when it can't find any txt
// records or if the TXT record is not standard.
func getRecord(host string, c Config, w http.ResponseWriter, r *http.Request) (record, error) {
	txts, err := query(host, r.Context(), c)
	if err != nil {
		log.Printf("Initial DNS query failed: %s", err)
	}
	// if error present or record empty, jump into wildcards
	if err != nil || txts[0] == "" {
		hostSlice := strings.Split(host, ".")
		hostSlice[0] = "_"
		host = strings.Join(hostSlice, ".")
		txts, err = query(host, r.Context(), c)
		if err != nil {
			log.Printf("Wildcard DNS query failed: %s", err.Error())
			return record{}, err
		}
	}

	if len(txts) != 1 {
		return record{}, fmt.Errorf("could not parse TXT record with %d records", len(txts))
	}

	var rec record
	if rec, err = ParseRecord(txts[0], w, r, c); err != nil {
		return rec, fmt.Errorf("could not parse record: %s", err)
	}

	r = rec.addToContext(r)

	return rec, nil
}

// Parse takes a string containing the DNS TXT record and returns
// a TXTDirect record struct instance.
// It will return an error if the DNS TXT record is not standard or
// if the record type is not enabled in the TXTDirect's config.
func ParseRecord(str string, w http.ResponseWriter, req *http.Request, c Config) (record, error) {
	r := record{}
	s := strings.Split(str, ";")
	for _, l := range s {
		switch {
		case strings.HasPrefix(l, "code="):
			l = strings.TrimPrefix(l, "code=")
			i, err := strconv.Atoi(l)
			if err != nil {
				return record{}, fmt.Errorf("could not parse status code: %s", err)
			}
			r.Code = i

		case strings.HasPrefix(l, "from="):
			l = strings.TrimPrefix(l, "from=")
			l, err := parsePlaceholders(l, req, []string{})
			if err != nil {
				return record{}, err
			}
			r.From = l

		case strings.HasPrefix(l, "re="):
			l = strings.TrimPrefix(l, "re=")
			r.Re = l

		case strings.HasPrefix(l, "ref="):
			l, err := strconv.ParseBool(strings.TrimPrefix(l, "ref="))
			if err != nil {
				fallback(w, req, "global", http.StatusMovedPermanently, c)
				return record{}, err
			}
			r.Ref = l

		case strings.HasPrefix(l, "root="):
			l = strings.TrimPrefix(l, "root=")
			l = ParseURI(l, w, req, c)
			r.Root = l

		case strings.HasPrefix(l, "to="):
			l = strings.TrimPrefix(l, "to=")
			l, err := parsePlaceholders(l, req, []string{})
			if err != nil {
				return record{}, err
			}
			l = ParseURI(l, w, req, c)
			r.To = l

		case strings.HasPrefix(l, "type="):
			l = strings.TrimPrefix(l, "type=")
			r.Type = l

		case strings.HasPrefix(l, "v="):
			l = strings.TrimPrefix(l, "v=")
			r.Version = l
			if r.Version != "txtv0" {
				return record{}, fmt.Errorf("unhandled version '%s'", r.Version)
			}
			log.Print("WARN: txtv0 is not suitable for production")

		case strings.HasPrefix(l, "vcs="):
			l = strings.TrimPrefix(l, "vcs=")
			r.Vcs = l

		case strings.HasPrefix(l, "website="):
			l = strings.TrimPrefix(l, "website=")
			l = ParseURI(l, w, req, c)
			r.Website = l

		default:
			tuple := strings.Split(l, "=")
			if len(tuple) != 2 {
				return record{}, fmt.Errorf("arbitrary data not allowed")
			}
			continue
		}
		if len(l) > 255 {
			return record{}, fmt.Errorf("TXT record cannot exceed the maximum of 255 characters")
		}
		if r.Type == "dockerv2" && r.To == "" {
			return record{}, fmt.Errorf("[txtdirect]: to= field is required in dockerv2 type")
		}
	}

	if r.Code == 0 {
		r.Code = http.StatusFound
	}

	if r.Type == "" {
		r.Type = "host"
	}

	if r.Type == "host" && r.To == "" {
		fallback(w, req, "global", http.StatusMovedPermanently, c)
		return nil
	}

	if !contains(c.Enable, r.Type) {
		return record{}, fmt.Errorf("%s type is not enabled in configuration", r.Type)
	}

	return r, nil
}

// Adds the given record to the request's context with "records" key.
func (rec record) addToContext(r *http.Request) *http.Request {
	// Fetch fallback config from context and add the record to it
	recordsContext := r.Context().Value("records")

	// Create a new records field in the context if it doesn't exist
	if recordsContext == nil {
		return r.WithContext(context.WithValue(r.Context(), "records", []record{rec}))
	}

	records := append(recordsContext.([]record), rec)

	// Replace the fallback config instance inside the request's context
	return r.WithContext(context.WithValue(r.Context(), "records", records))
}

// ParseURI parses the given URI and triggers fallback if the URI isn't valid
func ParseURI(uri string, w http.ResponseWriter, r *http.Request, c Config) string {
	url, err := url.Parse(uri)
	if err != nil {
		fallback(w, r, "global", http.StatusMovedPermanently, c)
		return ""
	}
	return url.String()
}
