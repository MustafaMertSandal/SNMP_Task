package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
}

type ServerConfig struct {
	Addr string `yaml:"addr"` //yaml içindeki anahtarda "addr" kısmını buraya doldurur.
}

func Load(path string) (Config, error) {
	var cfg Config

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	// Default
	if cfg.Server.Addr == "" {
		cfg.Server.Addr = ":8080"
	}

	fmt.Printf("Config loaded from %s\n", cfg.Server.Addr)

	return cfg, nil
}
