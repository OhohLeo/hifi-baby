package app

import (
	"github.com/Netflix/go-env"
)

type Config struct {
	ServerURL   string `env:"SERVER_URL,default=localhost:3000"`
	StoragePath string `env:"STORAGE_PATH,default=tracks"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	_, err := env.UnmarshalFromEnviron(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
