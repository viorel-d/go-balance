package pkg

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

// Service is the service to be load balanced
type Service struct {
	// service name
	Name string `yaml:"name"`
	// protocol://ip1:port1 protocol://ip2:port2 ...
	Replicas []string `yaml:"replicas"`
}

// Server is an instance of a running server
type Server struct {
	Url   *url.URL
	Proxy *httputil.ReverseProxy
}

func (s *Server) Forward(w http.ResponseWriter, r *http.Request) {
	s.Proxy.ServeHTTP(w, r)
}

// NewServer returns a Server instance
func NewServer(url *url.URL) *Server {
	return &Server{
		Url:   url,
		Proxy: httputil.NewSingleHostReverseProxy(url),
	}
}

// TODO: implement for round robin first
// abstract later
type ServerList struct {
	Servers []*Server

	// the current server to forward the request to
	current uint32
}

// Next atomically increments current server index in range of [0, len(Servers) -1]
func (sl *ServerList) Next() uint32 {
	n := atomic.AddUint32(&sl.current, 1)
	lenS := uint32(len(sl.Servers))
	if n > lenS {
		n -= lenS
	}

	return n
}

// GetNextServer returns the next server instance from servers list
func (sl *ServerList) GetNextServer() *Server {
	next := sl.Next()
	srv := sl.Servers[next]

	return srv
}
