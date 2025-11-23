package config

import (
	"os"
	"time"

	"github.com/lezzercringe/avito-test-assignment/internal/platform/postgres"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Postgres        postgres.Config `yaml:"postgres"`
	ShutdownTimeout time.Duration   `yaml:"shutdown_timeout"`
	ServeAddr       string          `yaml:"serve_addr"`
}

func Load(path string, cfg *Config) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return yaml.NewDecoder(f).Decode(cfg)
}
