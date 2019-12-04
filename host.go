package txtdirect

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// Host keeps data for "host" type requests
type Host struct {
	rw  http.ResponseWriter
	req *http.Request
	c   Config
	rec record
}

// NewHost returns a fresh instance of Host struct
func NewHost(w http.ResponseWriter, r *http.Request, rec record, c Config) *Host {
	return &Host{
		rw:  w,
		req: r,
		rec: rec,
		c:   c,
	}
}

// Redirect redirects the request to the endpoint defined in the record
func (h *Host) Redirect() error {
	to, code, err := getBaseTarget(h.rec, h.req)
	if err != nil {
		log.Print("Fallback is triggered because an error has occurred: ", err)
		fallback(h.rw, h.req, "to", code, h.c)
		return nil
	}
	log.Printf("[txtdirect]: %s > %s", h.req.Host+h.req.URL.Path, to)
	if code == http.StatusMovedPermanently {
		h.rw.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d", status301CacheAge))
	}
	h.rw.Header().Add("Status-Code", strconv.Itoa(code))
	http.Redirect(h.rw, h.req, to, code)
	if h.c.Prometheus.Enable {
		RequestsByStatus.WithLabelValues(h.req.Host, strconv.Itoa(code)).Add(1)
	}
	return nil
}
