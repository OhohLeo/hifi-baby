package app

import (
	"github.com/Netflix/go-env"

	"github.com/OhohLeo/hifi-baby/audio"
	"github.com/OhohLeo/hifi-baby/http"
)

type Config struct {
	Server http.Config
	Audio  audio.Config

	LogLevel         string `env:"LOG_LEVEL,default=info"`
	StoredConfigPath string `env:"STORED_CONFIG_PATH,default=stored_config.json"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	_, err := env.UnmarshalFromEnviron(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
