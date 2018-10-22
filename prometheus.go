package txtdirect

import (
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus contains Prometheus's configuration
type Prometheus struct {
	Enable   bool
	Address  string
	Path     string
	Serve    string
	Hostname string
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
		Help:      "Total requests per host",
	}, []string{"host", "status"})
)

func InitPrometheus() {
	err := prometheus.Register(RequestsCount)
	err = prometheus.Register(RequestsByStatus)
	if err != nil {
		log.Printf("<%s> [txtdirect]: Prometheus registration error: %s", time.Now().Format(logFormat), err.Error())
	}
}
