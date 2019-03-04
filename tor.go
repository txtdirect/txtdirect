package txtdirect

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/mholt/caddy"
	cproxy "github.com/mholt/caddy/caddyhttp/proxy"
	"golang.org/x/net/proxy"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// DefaultOnionServicePort is the port used to serve the onion service on
const DefaultOnionServicePort = 4242

// Tor type config struct
type Tor struct {
	Enable bool
	// Socks5 proxy port
	Port      int
	DataDir   string
	Torrc     string
	DebugMode bool
	LogFile   string

	instance        *tor.Tor
	contextCanceler context.CancelFunc
	onion           *tor.OnionService
}

type TorResponse struct {
	headers    http.Header
	body       []byte
	bodyReader bytes.Buffer
	bodyWriter bytes.Buffer
	status     int
}

// TODO: Discuss these values
const (
	torProxyKeepalive = 30000000
	torFallbackDelay  = 30000000 * time.Millisecond
	torProxyTimeout   = 30000000 * time.Second
)

var bufferPool = sync.Pool{New: createBuffer}

func createBuffer() interface{} {
	return make([]byte, 0, 32*1024)
}

func (t *Tor) Start(c *caddy.Controller) {
	var debugger io.Writer
	if t.DebugMode {
		if t.LogFile != "" {
			debugger = &lumberjack.Logger{
				Filename:   t.LogFile,
				MaxSize:    100,
				MaxAge:     14,
				MaxBackups: 10,
			}
		}
		debugger = os.Stdout
	}

	torInstance, err := tor.Start(nil, &tor.StartConf{
		NoAutoSocksPort: true,
		ExtraArgs:       []string{"--SocksPort", strconv.Itoa(t.Port)},
		TempDataDirBase: t.DataDir,
		TorrcFile:       t.Torrc,
		DebugWriter:     debugger,
	})
	if err != nil {
		log.Panicf("Unable to start Tor: %v", err)
	}

	listenCtx := context.Background()

	onion, err := torInstance.Listen(listenCtx, &tor.ListenConf{LocalPort: 8868, RemotePorts: []int{80}})
	if err != nil {
		log.Panicf("Unable to start onion service: %v", err)
	}

	t.onion = onion
	t.instance = torInstance
}

// Stop stops the tor instance, context listener and the onion service
func (t *Tor) Stop() error {
	if err := t.instance.Close(); err != nil {
		return fmt.Errorf("[txtdirect]: Couldn't close the tor instance. %s", err.Error())
	}
	t.onion.Close()
	return nil
}

// Proxy redirects the request to the local onion serivce and the actual proxying
// happens inside onion service's http handler
func (t *Tor) Proxy(w http.ResponseWriter, r *http.Request, rec record, c Config) error {
	u, err := url.Parse(rec.To)
	if err != nil {
		return err
	}
	u.Path = r.URL.Path

	// Create a socks5 dialer
	dialer, err := proxy.SOCKS5("tcp", fmt.Sprintf("127.0.0.1:%d", t.Port), nil, proxy.Direct)
	if err != nil {
		log.Fatal(err)
	}

	reverseProxy := cproxy.NewSingleHostReverseProxy(u, "", torProxyKeepalive, torProxyTimeout, torFallbackDelay)
	reverseProxy.Transport = &http.Transport{
		Dial: dialer.Dial,
	}

	tmpResposne := TorResponse{headers: make(http.Header)}
	if err := reverseProxy.ServeHTTP(&tmpResposne, r, nil); err != nil {
		return fmt.Errorf("[txtdirect]: Coudln't proxy the request to the background onion service. %s", err.Error())
	}

	// Decompress the body based on "Content-Encoding" header and write to a writer buffer
	if err := tmpResposne.WriteBody(); err != nil {
		return fmt.Errorf("[txtdirect]: Couldn't write the response body: %s", err.Error())
	}

	// Replace the URL hosts with the request's host
	if err := tmpResposne.ReplaceBody(u.Scheme, u.Host, r.Host); err != nil {
		return fmt.Errorf("[txtdirect]: Couldn't replace urls inside the response body: %s", err.Error())
	}

	copyHeader(w.Header(), tmpResposne.Header())

	// Write the final response from the temporary ResponseWriter to the main ResponseWriter
	if _, err := w.Write(tmpResposne.Body()); err != nil {
		return fmt.Errorf("[txtdirect]: Couldn't write the temporary response to main response body: %s", err.Error())
	}

	return nil
}

// ParseTor parses the txtdirect config for Tor proxy
func (t *Tor) ParseTor(c *caddy.Controller) error {
	switch c.Val() {
	case "port":
		value, err := strconv.Atoi(c.RemainingArgs()[0])
		if err != nil {
			return fmt.Errorf("The given value for port field is not standard. It should be an integer")
		}
		t.Port = value

	case "datadir":
		t.DataDir = c.RemainingArgs()[0]

	case "torrc":
		t.Torrc = c.RemainingArgs()[0]

	case "debug_mode":
		value, err := strconv.ParseBool(c.RemainingArgs()[0])
		if err != nil {
			return fmt.Errorf("The given value for debug_mode field is not standard. It should be a boolean")
		}
		t.DebugMode = value

	case "logfile":
		t.LogFile = c.RemainingArgs()[0]

	default:
		return c.ArgErr() // unhandled option for tor
	}
	return nil
}

// SetDefaults sets the default values for prometheus config
// if the fields are empty
func (t *Tor) SetDefaults() {
	if t.Port == 0 {
		t.Port = DefaultOnionServicePort
	}
}

// Header returns response headers
func (t *TorResponse) Header() http.Header {
	return t.headers
}

func (t *TorResponse) Write(body []byte) (int, error) {
	reader := bytes.NewReader(body)
	pooledIoCopy(&t.bodyReader, reader)
	t.body = body
	return len(body), nil
}

// Body returns response's body. This method should only get called after WriteBody()
func (t *TorResponse) Body() []byte {
	return t.bodyWriter.Bytes()
}

// WriteHeader Writes the given status code to response
func (t *TorResponse) WriteHeader(status int) {
	t.status = status
}

func (t *TorResponse) ReplaceBody(scheme, to, host string) error {
	replacedBody := bytes.Replace(t.bodyWriter.Bytes(), []byte(scheme+"://"+to), []byte(scheme+"://"+host), -1)
	t.bodyWriter.Reset()
	if _, err := t.bodyWriter.Write(replacedBody); err != nil {
		return err
	}
	return nil
}

func (t *TorResponse) WriteBody() error {
	switch t.Header().Get("Content-Encoding") {
	case "gzip":
		reader, err := gzip.NewReader(&t.bodyReader)
		if err != nil {
			return err
		}
		defer reader.Close()
		_, err = io.Copy(&t.bodyWriter, reader)
		if err != nil {
			return err
		}
		t.Header().Del("Content-Encoding")
	default:
		_, err := io.Copy(&t.bodyWriter, &t.bodyReader)
		if err != nil {
			return err
		}
	}
	return nil
}

var skipHeaders = map[string]struct{}{
	"Content-Type":        {},
	"Content-Disposition": {},
	"Accept-Ranges":       {},
	"Set-Cookie":          {},
	"Cache-Control":       {},
	"Expires":             {},
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		if _, ok := dst[k]; ok {
			if _, shouldSkip := skipHeaders[k]; shouldSkip {
				continue
			}
			if k != "Server" {
				dst.Del(k)
			}
		}
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func pooledIoCopy(dst io.Writer, src io.Reader) {
	buf := bufferPool.Get().([]byte)
	defer bufferPool.Put(buf)

	bufCap := cap(buf)
	io.CopyBuffer(dst, src, buf[0:bufCap:bufCap])
}
