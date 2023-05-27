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
	port = flag.Int("port", 8080, "port to run go-balance on")
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
	Config *Config
}

func (gb GoBalance) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	srv := serverList.Servers[serverList.current]
	serverList.Next()
	srv.proxy.ServeHTTP(w, r)
}

// NewGoBalance returns an instance of GoBalance
func NewGoBalance(config *Config) GoBalance {
	return GoBalance{
		Config: config,
	}
}

var (
	config     Config
	serverList ServerList
)

func InitServers(gb *GoBalance) {
	for _, s := range gb.Config.Services {
		for _, r := range s.Replicas {
			if url, err := url.Parse(r); err != nil {
				log.Panicf("InitServers err: %s\n", err)
			} else {
				srv := NewServer(url)
				serverList.Servers = append(serverList.Servers, srv)
			}
		}
	}
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
	InitServers(&gb)

	strPort := fmt.Sprint(*port)
	srv := http.Server{
		Addr:    ":" + strPort,
		Handler: gb,
	}

	log.Printf("Starting GoBalance on port: %s\n", strPort)
	srv.ListenAndServe()
}
