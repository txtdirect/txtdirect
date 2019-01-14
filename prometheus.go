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
		Name:      "redirect_count_total",
		Help:      "Total requests per host",
	}, []string{"host"})

	RequestsByStatus = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "txtdirect",
		Name:      "redirect_status_count_total",
		Help:      "Total returned statuses per host",
	}, []string{"host", "status"})

	RequestsCountBasedOnType = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "txtdirect",
		Name:      "redirect_type_count_total",
		Help:      "Total requests for each host based on type",
	}, []string{"host", "type"})

	FallbacksCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "txtdirect",
		Name:      "fallback_type_count_total",
		Help:      "Total fallbacks triggered for each type",
	}, []string{"host", "type"})

	once sync.Once
)

const (
	shutdownTimeout time.Duration = time.Second * 5
	// prometheusAddr is the address the where the metrics are exported by default.
	prometheusAddr string = "localhost:9183"
	prometheusPath string = "/metrics"
)

func NewPrometheus(addr, path string) *Prometheus {
	if addr == "" {
		addr = prometheusAddr
	}
	if path == "" {
		path = prometheusPath
	}
	p := &Prometheus{
		Path:    path,
		Address: addr,
	}
	return p
}

func (p *Prometheus) start() error {
	p.once.Do(func() {
		prometheus.MustRegister(RequestsCount)
		prometheus.MustRegister(RequestsByStatus)
		http.Handle(p.Path, p.handler)
		go func() {
			err := http.ListenAndServe(p.Address, nil)
			if err != nil {
				log.Printf("<%s> [txtdirect]: Couldn't start http handler for prometheus metrics. %s", time.Now().Format(logFormat), err.Error())
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
