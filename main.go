package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync/atomic"

	"github.com/go-yaml/yaml"
	log "github.com/sirupsen/logrus"
)

var (
	port   = flag.Int("port", 8080, "port to run go-balance on")
	config Config
)

// Service is the service to be load balanced
type Service struct {
	// service name
	Name string `yaml:"name"`
	// protocol://ip1:port1 protocol://ip2:port2 ...
	Replicas []string `yaml:"replicas"`
}

// Config is the configuration for the load balancer
type Config struct {
	Services []Service `yaml:"services"`

	// load balancing strategy
	Strategy string `yaml:"strategy"`
}

// Server is an instance of a running server
type Server struct {
	url   *url.URL
	proxy *httputil.ReverseProxy
}

// NewServer returns a Server instance
func NewServer(url *url.URL) *Server {
	return &Server{
		url:   url,
		proxy: httputil.NewSingleHostReverseProxy(url),
	}
}

// implement for round-robin first
// abstract later
type ServerList struct {
	Servers []*Server

	// the current server to forward the request to
	current uint32
}

// this should be concurrent safe
func (sl *ServerList) Next() uint32 {
	n := atomic.AddUint32(&sl.current, 1)
	lenS := uint32(len(sl.Servers))
	if n > lenS {
		n -= lenS
	}

	return n
}

// GoBalance is the load balancer
type GoBalance struct {
	Config     *Config
	ServerList *ServerList
}

func (gb GoBalance) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: support per service forwarding
	srv := gb.getNextServer()
	srv.proxy.ServeHTTP(w, r)
}

func (gb *GoBalance) initServers() {
	gb.ServerList.Servers = make([]*Server, 0)

	for _, s := range gb.Config.Services {
		for _, r := range s.Replicas {
			if url, err := url.Parse(r); err != nil {
				log.Panicf("initServers err: %s\n", err)
			} else {
				srv := NewServer(url)
				gb.ServerList.Servers = append(gb.ServerList.Servers, srv)
			}
		}
	}
}

func (gb *GoBalance) getNextServer() *Server {
	next := gb.ServerList.Next()
	srv := gb.ServerList.Servers[next]

	return srv
}

// NewGoBalance returns an instance of GoBalance
func NewGoBalance(config *Config) *GoBalance {
	gb := GoBalance{
		Config: config,
	}
	gb.initServers()

	return &gb
}

func InitConfig() {
	if f, err := os.ReadFile("./config.yaml"); err != nil {
		log.Panicf("InitConfig err: %s\n", err)
	} else {
		uErr := yaml.Unmarshal(f, &config)
		if uErr != nil {
			log.Panicf("InitConfig err: %s\n", uErr)
		}
	}
}

func main() {
	flag.Parse()

	InitConfig()
	gb := NewGoBalance(&config)

	strPort := fmt.Sprint(*port)
	srv := http.Server{
		Addr:    ":" + strPort,
		Handler: gb,
	}

	log.Printf("Starting GoBalance on port: %s\n", strPort)
	srv.ListenAndServe()
}
