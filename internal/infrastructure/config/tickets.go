package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// TicketsConfig represents the tickets configuration
type TicketsConfig struct {
	Tickets []string `yaml:"tickets"`
}

// LoadTickets loads tickets from a YAML file
func LoadTickets(path string) (*TicketsConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config TicketsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
