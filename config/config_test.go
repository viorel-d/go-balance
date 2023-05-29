package config

import (
	"bytes"
	"testing"

	"github.com/go-yaml/yaml"
)

type TestService struct {
	Name     string   `yaml:"name"`
	Replicas []string `yaml:"replicas"`
}

type TestConfig struct {
	Strategy string        `yaml:"strategy"`
	Services []TestService `yaml:"services"`
}

var ec = TestConfig{
	Strategy: "RoundRobin",
	Services: []TestService{{
		Name: "test",
		Replicas: []string{
			"http://127.0.0.1:8081",
			"http://127.0.0.1:8082",
		},
	}},
}

func GetTestConfig() (*Config, error) {
	f, _ := yaml.Marshal(ec)
	reader := bytes.NewBuffer(f)

	return Get(reader)
}

func TestGet(t *testing.T) {
	_, err := GetTestConfig()
	if err != nil {
		t.Errorf("expected nil error, got %s\n", err)
	}
}

func TestConfigStrategy(t *testing.T) {
	c, _ := GetTestConfig()

	if c.Strategy != ec.Strategy {
		t.Errorf("expected strategy %s, got %s\n", ec.Strategy, c.Strategy)
	}
}

func TestConfigServices(t *testing.T) {
	c, _ := GetTestConfig()

	if len(c.Services) != len(ec.Services) {
		t.Errorf("expected services len %v, got %v\n", len(ec.Services), len(c.Services))
	}

	if c.Services[0].Name != ec.Services[0].Name {
		t.Errorf("expected service[0] name %s, got %s\n", ec.Services[0].Name, c.Services[0].Name)
	}

	if len(c.Services[0].Replicas) != len(ec.Services[0].Replicas) {
		t.Errorf("expected service[0] replicas len %v, got %v\n", len(ec.Services[0].Replicas), len(c.Services[0].Replicas))
	}

	for i, r := range c.Services[0].Replicas {
		if r != ec.Services[0].Replicas[i] {
			t.Errorf("expected service[0] replica[%v] %s, got %s", i, ec.Services[0].Replicas[i], r)
			break
		}
	}
}
