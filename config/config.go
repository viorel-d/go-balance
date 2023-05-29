package config

import (
	"io"
	"io/ioutil"

	"github.com/go-yaml/yaml"
	"github.com/viorel-d/go-balance/pkg"
)

// Config is the configuration for the load balancer
type Config struct {
	// load balancing strategy
	Strategy string `yaml:"strategy"`

	// services to be load balanced
	Services []pkg.Service `yaml:"services"`
}

// Get returns a reference to Config if reader content is yaml compatible
func Get(reader io.Reader) (*Config, error) {
	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	if err := yaml.Unmarshal(buf, config); err != nil {
		return nil, err
	}

	return config, nil
}
