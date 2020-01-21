package variables

import "time"

const (
	Basezone          = "_redirect"
	DefaultSub        = "www"
	DefaultProtocol   = "https"
	ProxyKeepalive    = 30
	FallbackDelay     = 300 * time.Millisecond
	ProxyTimeout      = 30 * time.Second
	Status301CacheAge = 604800
	// Testing DNS server port
	Port = 6000
)
