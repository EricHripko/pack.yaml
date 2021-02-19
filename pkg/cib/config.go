package cib

import (
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

// Config that drives that image creation. Typically stored in pack.yaml file.
type Config struct {
	// Whether debugging capabilities should be included in the image.
	Debug bool
	// Entrypoint for the resulting image.
	Entrypoint []string
	// Command for the resulting image.
	Command []string
	// User to be used in the resulting image.
	User string
	// Other configuration fields. Typically used by plugins for additional
	// settings.
	Other map[string]interface{} `mapstructure:",remain"`
}

// NewConfig returns an instance of configuration with pre-populated defaults.
func NewConfig() *Config {
	return &Config{
		Debug:      true,
		Entrypoint: []string{},
		Command:    []string{},
		User:       "nobody",
		Other:      make(map[string]interface{}),
	}
}

// ReadConfig parses the configuration provided into a structured format.
func ReadConfig(data []byte) (*Config, error) {
	// Decode YAML
	m := make(map[string]interface{})
	if err := yaml.Unmarshal(data, m); err != nil {
		return nil, err
	}

	// Map
	config := NewConfig()
	err := mapstructure.Decode(m, config)
	return config, err
}
