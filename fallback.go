package txtdirect

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Fallback struct {
	rw      http.ResponseWriter
	request *http.Request
	config  Config

	records    []record
	pathRecord record
	lastRecord record

	fallbackType string
	code         int
}

// fallback redirects the request to the given fallback address
// and if it's not provided it will check txtdirect config for
// default fallback address
func fallback(w http.ResponseWriter, r *http.Request, fallbackType string, code int, c Config) {
	if code == http.StatusMovedPermanently {
		w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d", status301CacheAge))
	}
	w.Header().Add("Status-Code", strconv.Itoa(code))

	f := Fallback{
		rw:           w,
		request:      r,
		config:       c,
		fallbackType: fallbackType,
		code:         code,
	}

	if fallbackType != "global" {
		// Fetch records from request's context and set the []record type on them
		f.fetchRecords()

		// Redirect to first record's `to=` field
		if (fallbackType == "to" && f.lastRecord.To != "") && f.validateURIs() {
			http.Redirect(w, r, f.lastRecord.To, code)
			f.countFallback(f.lastRecord.Type)
		}

		// Redirect to first record's `website=` field
		if fallbackType == "website" && f.lastRecord.Website != "" {
			http.Redirect(w, r, f.lastRecord.Website, code)
			f.countFallback(f.lastRecord.Type)
		}

		// Redirect to first record's `root=` field
		if fallbackType == "root" && f.lastRecord.Root != "" {
			http.Redirect(w, r, f.lastRecord.Root, code)
			f.countFallback(f.lastRecord.Type)
		}

		// Redirect to path record's `website=` field
		if fallbackType == "website" && f.pathRecord.Website != "" {
			http.Redirect(w, r, f.pathRecord.Website, code)
			f.countFallback(f.pathRecord.Type)
		}

		// Redirect to path record's `root=` field
		if fallbackType == "root" && f.pathRecord.Root != "" {
			http.Redirect(w, r, f.pathRecord.Root, code)
			f.countFallback(f.pathRecord.Type)
		}

		// Redirect to path record's `to=` field
		if f.pathRecord.To != "" {
			http.Redirect(w, r, f.pathRecord.To, code)
			f.countFallback(f.pathRecord.Type)
		}

		// If non of the above cases applied on the record, jump into global redirects
		f.globalFallbacks(f.lastRecord.Type)
		log.Printf("[txtdirect]: %s > %s", r.Host+r.URL.Path, w.Header().Get("Location"))
		return
	}

	f.globalFallbacks("")

	log.Printf("[txtdirect]: %s > %s", r.Host+r.URL.Path, w.Header().Get("Location"))
}

func (f *Fallback) countFallback(recType string) {
	if f.config.Prometheus.Enable {
		FallbacksCount.WithLabelValues(f.request.Host, recType, f.fallbackType).Add(1)
		RequestsByStatus.WithLabelValues(f.request.URL.Host, string(f.code)).Add(1)
	}
}

func (f *Fallback) globalFallbacks(recordType string) {
	if contains(f.config.Enable, "www") {
		s := strings.Join([]string{defaultProtocol, "://", defaultSub, ".", f.request.URL.Host}, "")

		http.Redirect(f.rw, f.request, s, f.code)

		f.countFallback(recordType)
	} else if f.config.Redirect != "" {
		f.rw.Header().Set("Status-Code", strconv.Itoa(http.StatusMovedPermanently))

		http.Redirect(f.rw, f.request, f.config.Redirect, http.StatusMovedPermanently)

		f.code = http.StatusMovedPermanently

		f.countFallback(recordType)
	} else {
		http.NotFound(f.rw, f.request)
	}
}

func (f *Fallback) fetchRecords() {
	f.records = f.request.Context().Value("records").([]record)
	// Note: This condition should get changed when we support more record aggregations.
	if len(f.records) >= 2 {
		f.pathRecord = f.records[len(f.records)-2]
	}
	f.lastRecord = f.records[len(f.records)-1]
}

func (f *Fallback) validateURIs() bool {
	if _, err := url.Parse(f.lastRecord.To); err != nil {
		return false
	}
	if _, err := url.Parse(f.pathRecord.To); err != nil {
		return false
	}
	return true
}
