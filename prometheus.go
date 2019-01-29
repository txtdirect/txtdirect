package txtdirect

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Prometheus contains Prometheus's configuration
type Prometheus struct {
	Enable  bool
	Address string
	Path    string

	once    sync.Once
	next    httpserver.Handler
	handler http.Handler
}

var (
	RequestsCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "txtdirect",
		Name:      "total_requests_per_host",
		Help:      "Total requests per host",
	}, []string{"host"})

	RequestsByStatus = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "txtdirect",
		Name:      "total_returned_statuses_per_host",
		Help:      "Total returned statuses per host",
	}, []string{"host", "status"})

	once sync.Once
)

const (
	shutdownTimeout time.Duration = time.Second * 5
	// prometheusAddr is the address the where the metrics are exported by default.
	prometheusAddr string = "localhost:9183"
	prometheusPath string = "/metrics"
)

// SetDefaults sets the default values for prometheus config
// if the fields are empty
func (p *Prometheus) SetDefaults() {
	if p.Address == "" {
		p.Address = prometheusAddr
	}
	if p.Path == "" {
		p.Path = prometheusPath
	}
}

func (p *Prometheus) start() error {
	p.once.Do(func() {
		prometheus.MustRegister(RequestsCount)
		prometheus.MustRegister(RequestsByStatus)
		http.Handle(p.Path, p.handler)
		go func() {
			err := http.ListenAndServe(p.Address, nil)
			if err != nil {
				log.Printf("[txtdirect]: Couldn't start http handler for prometheus metrics. %s", err.Error())
			}
		}()
	})
	return nil
}

func (p *Prometheus) Setup(c *caddy.Controller) {
	p.handler = promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
		ErrorHandling: promhttp.HTTPErrorOnError,
		ErrorLog:      log.New(os.Stderr, "", log.LstdFlags),
	})

	once.Do(func() {
		c.OnStartup(p.start)
	})

	cfg := httpserver.GetConfig(c)
	cfg.AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		p.next = next
		return p
	})
}

func (p *Prometheus) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	next := p.next

	rw := httpserver.NewResponseRecorder(w)

	status, err := next.ServeHTTP(rw, r)

	return status, err
}

// ParsePrometheus parses the txtdirect config for Prometheus
func (p *Prometheus) ParsePrometheus(c *caddy.Controller, key, value string) error {
	switch key {
	case "enable":
		value, err := strconv.ParseBool(value)
		if err != nil {
			return c.ArgErr()
		}
		p.Enable = value
	case "address":
		// TODO: validate the given address
		p.Address = value
	case "path":
		p.Path = value
	default:
		return c.ArgErr() // unhandled option for prometheus
	}
	return nil
}
