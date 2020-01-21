package record

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"go.txtdirect.org/txtdirect/config"
	"go.txtdirect.org/txtdirect/plugins/prometheus"
	"go.txtdirect.org/txtdirect/variables"
)

// Fallback keeps the data necessary for the fallback flow
type FallbackType struct {
	rw      http.ResponseWriter
	request *http.Request
	config  config.Config

	records    []Record
	pathRecord Record
	lastRecord Record

	fallbackType string
	code         int

	// Which record to use for fallback. (last record or path record)
	last        bool
	aggregation bool
}

// Fallback redirects the request to the given fallback address
// and if it's not provided it will check txtdirect config for
// default fallback address
func Fallback(w http.ResponseWriter, r *http.Request, fallbackType string, code int, c config.Config) {
	if code == http.StatusMovedPermanently {
		w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d", variables.Status301CacheAge))
	}
	w.Header().Add("Status-Code", strconv.Itoa(code))

	f := FallbackType{
		rw:           w,
		request:      r,
		config:       c,
		fallbackType: fallbackType,
		code:         code,
	}

	if fallbackType != "global" {
		// Fetch records from request's context and set the []record type on them
		f.fetchRecords()

		if !f.lastRecordFallback() {
			if !f.pathFallback() {
				// If non of the above cases applied on the record, jump into global redirects
				f.globalFallbacks(f.lastRecord.Type)
			}
		}
		log.Printf("[txtdirect]: %s > %s", r.Host+r.URL.Path, w.Header().Get("Location"))
		return
	}

	f.globalFallbacks("")

	log.Printf("[txtdirect]: %s > %s", r.Host+r.URL.Path, w.Header().Get("Location"))
}

func (f *FallbackType) countFallback(recType string) {
	if f.config.Prometheus.Enable {
		prometheus.FallbacksCount.WithLabelValues(f.request.Host, recType, f.fallbackType).Add(1)
		prometheus.RequestsByStatus.WithLabelValues(f.request.URL.Host, strconv.Itoa(f.code)).Add(1)
	}
}

func (f *FallbackType) globalFallbacks(recordType string) {
	if contains(f.config.Enable, "www") {
		s := strings.Join([]string{variables.DefaultProtocol, "://", variables.DefaultSub, ".", f.request.URL.Host}, "")

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

func (f *FallbackType) fetchRecords() {
	f.records = f.request.Context().Value("records").([]Record)
	// Note: This condition should get changed when we support more record aggregations.
	if len(f.records) >= 2 {
		f.pathRecord = f.records[len(f.records)-2]
	}
	f.lastRecord = f.records[len(f.records)-1]
}

// Checks the last record's `to=`, `website=`, and `root=` field to fallback
// Returns false if it can't find an endpoint to fallback to
func (f *FallbackType) lastRecordFallback() bool {
	// Redirect to first record's `to=` field
	if f.fallbackType == "to" && f.lastRecord.To != "" {
		http.Redirect(f.rw, f.request, f.lastRecord.To, f.code)
		f.countFallback(f.lastRecord.Type)
		return true
	}

	// Redirect to first record's `website=` field
	if f.fallbackType == "website" && f.lastRecord.Website != "" {
		http.Redirect(f.rw, f.request, f.lastRecord.Website, f.code)
		f.countFallback(f.lastRecord.Type)
		return true
	}

	// Redirect to first record's `root=` field
	if f.fallbackType == "root" && f.lastRecord.Root != "" {
		http.Redirect(f.rw, f.request, f.lastRecord.Root, f.code)
		f.countFallback(f.lastRecord.Type)
		return true
	}
	return false
}

// Checks the path record's `to=`, `website=`, and `root=` field to fallback
// Returns false if it can't find an endpoint to fallback to
func (f *FallbackType) pathFallback() bool {
	// Redirect to path record's `website=` field
	if f.fallbackType == "website" && f.pathRecord.Website != "" {
		http.Redirect(f.rw, f.request, f.pathRecord.Website, f.code)
		f.countFallback(f.pathRecord.Type)
		return true
	}

	// Redirect to path record's `root=` field
	if f.fallbackType == "root" && f.pathRecord.Root != "" {
		http.Redirect(f.rw, f.request, f.pathRecord.Root, f.code)
		f.countFallback(f.pathRecord.Type)
		return true
	}

	// Redirect to path record's `to=` field
	if f.pathRecord.To != "" {
		http.Redirect(f.rw, f.request, f.pathRecord.To, f.code)
		f.countFallback(f.pathRecord.Type)
		return true
	}
	return false
}
