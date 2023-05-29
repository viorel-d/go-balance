package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
	"github.com/viorel-d/go-balance/config"
	"github.com/viorel-d/go-balance/pkg"
)

var (
	port       = flag.Int("port", 8080, "port to run go-balance on")
	configPath = flag.String("config", "./config.yaml", "Path of the config file")
)

// GoBalance is the load balancer
type GoBalance struct {
	Config     *config.Config
	ServerList *pkg.ServerList
}

func (gb GoBalance) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: support per service forwarding
	// TODO: add request logging
	// TODO: add error handling
	srv := gb.ServerList.GetNextServer()
	srv.Forward(w, r)
}

func (gb *GoBalance) initServers() {
	gb.ServerList.Servers = make([]*pkg.Server, 0)

	for _, s := range gb.Config.Services {
		for _, r := range s.Replicas {
			if url, err := url.Parse(r); err != nil {
				log.Panicf("initServers err: %s\n", err)
			} else {
				srv := pkg.NewServer(url)
				gb.ServerList.Servers = append(gb.ServerList.Servers, srv)
			}
		}
	}
}

// NewGoBalance returns an instance of GoBalance
func NewGoBalance(config *config.Config) *GoBalance {
	gb := GoBalance{
		Config: config,
	}
	gb.initServers()

	return &gb
}

func main() {
	flag.Parse()

	configFile, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Panicln(err)
	}
	configBuf := bytes.NewBuffer(configFile)
	config, err := config.Get(configBuf)
	if err != nil {
		log.Panicln(err)
	}

	strPort := fmt.Sprint(*port)
	srv := http.Server{
		Addr:    ":" + strPort,
		Handler: NewGoBalance(config),
	}

	log.Printf("Starting GoBalance on port: %s\n", strPort)
	srv.ListenAndServe()
}
