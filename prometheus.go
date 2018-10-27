package txtdirect

import (
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Prometheus contains Prometheus's configuration
type Prometheus struct {
	Enable   bool
	Address  string
	Path     string
	Hostname string
	once     sync.Once
	next     httpserver.Handler
	handler  http.Handler
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
	prometheusAddr string = "localhost:9191"
	prometheusPath string = "/metrics"
)

func NewPrometheus(addr, path string) *Prometheus {
	if addr == "" || path == "" {
		addr = prometheusAddr
		path = prometheusPath
	}
	p := &Prometheus{
		Address: addr,
		Path:    path,
	}
	return p
}

func (p *Prometheus) start() error {
	p.once.Do(func() {

		prometheus.MustRegister(RequestsCount)
		prometheus.MustRegister(RequestsByStatus)

		if p.Address != "" {
			http.Handle(p.Path, p.handler)
			go func() {
				err := http.ListenAndServe(p.Address, nil)
				if err != nil {
					log.Printf("[ERROR] Starting handler: %v", err)
				}
			}()
		}
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
		return httpserver.HandlerFunc(func(w http.ResponseWriter, r *http.Request) (int, error) {
			if r.URL.Path == p.Path {
				p.handler.ServeHTTP(w, r)
				return 0, nil
			}
			return next.ServeHTTP(w, r)
		})
	})
	cfg.AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		p.next = next
		return p.next
	})
}
