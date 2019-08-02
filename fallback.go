package txtdirect

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// fallback redirects the request to the given fallback address
// and if it's not provided it will check txtdirect config for
// default fallback address
func fallback(w http.ResponseWriter, r *http.Request, fallback, recordType, fallbackType string, code int, c Config) {
	if code == http.StatusMovedPermanently {
		w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d", status301CacheAge))
	}
	w.Header().Add("Status-Code", strconv.Itoa(code))

	if fallbackType != "global" {
		// Fetch records from request's context and set the []record type on them
		records := r.Context().Value("records").([]record)

		// Redirect to first record's `to=` field
		if fallbackType == "to" && records[0].To != "" {
			http.Redirect(w, r, records[0].To, code)
			if c.Prometheus.Enable {
				countFallback(r, records[0].Type, fallbackType, code)
			}
		}

		// Redirect to first record's `website=` field
		if fallbackType == "website" && records[0].Website != "" {
			http.Redirect(w, r, records[0].Website, code)
			if c.Prometheus.Enable {
				countFallback(r, records[0].Type, fallbackType, code)
			}
		}

		// Special case when path is used in fetching the final record
		var pathRecord record
		if len(records) >= 2 {
			pathRecord = records[len(records)-2]
		}

		if fallbackType == "root" && pathRecord.Root != "" {
			http.Redirect(w, r, pathRecord.Root, code)
			if c.Prometheus.Enable {
				countFallback(r, records[len(records)-1].Type, fallbackType, code)
			}
		}

		if pathRecord.To != "" {
			http.Redirect(w, r, pathRecord.To, code)
			if c.Prometheus.Enable {
				countFallback(r, records[len(records)-1].Type, fallbackType, code)
			}
		}

		// If non of the above cases applied on the record, jump into global redirects
		globalFallbacks(w, r, c, code, records[len(records)-1].Type)
		log.Printf("[txtdirect]: %s > %s", r.Host+r.URL.Path, w.Header().Get("Location"))
		return
	}

	globalFallbacks(w, r, c, code, "")

	log.Printf("[txtdirect]: %s > %s", r.Host+r.URL.Path, w.Header().Get("Location"))
}

func addRecordToContext(r *http.Request, rec record) *http.Request {
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

func countFallback(r *http.Request, recType, fallbackType string, code int) {
	FallbacksCount.WithLabelValues(r.Host, recType, fallbackType).Add(1)
	RequestsByStatus.WithLabelValues(r.URL.Host, string(code)).Add(1)
}

func globalFallbacks(w http.ResponseWriter, r *http.Request, c Config, code int, recordType string) {
	if contains(c.Enable, "www") {
		s := strings.Join([]string{defaultProtocol, "://", defaultSub, ".", r.URL.Host}, "")
		http.Redirect(w, r, s, code)
		if c.Prometheus.Enable {
			countFallback(r, recordType, "subdomain", code)
		}
	} else if c.Redirect != "" {
		w.Header().Set("Status-Code", strconv.Itoa(http.StatusMovedPermanently))

		http.Redirect(w, r, c.Redirect, http.StatusMovedPermanently)

		if c.Prometheus.Enable {
			countFallback(r, recordType, "redirect", http.StatusMovedPermanently)
		}
	} else {
		http.NotFound(w, r)
	}
}
