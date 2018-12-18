package txtdirect

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus contains Prometheus's configuration
type Prometheus struct {
	Enable  bool
	Address string
	Path    string
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
		Path:    path,
		Address: addr,
	}
	return p
}

func (p *Prometheus) Start() {
	prometheus.Register(RequestsCount)
	prometheus.Register(RequestsByStatus)
}
