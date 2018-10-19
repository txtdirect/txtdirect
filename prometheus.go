package txtdirect

// Prometheus contains Prometheus's configuration
type Prometheus struct {
	Enable   bool
	Address  string
	Path     string
	Serve    string
	Hostname string
}
