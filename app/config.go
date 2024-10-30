package app

import (
	"github.com/Netflix/go-env"

	"github.com/OhohLeo/hifi-baby/audio"
	"github.com/OhohLeo/hifi-baby/http"
	"github.com/OhohLeo/hifi-baby/sql"
)

type Config struct {
	Audio    audio.Config
	Database sql.Config
	Server   http.Config

	LogLevel     string `env:"LOG_LEVEL,default=info"`
	SettingsPath string `env:"SETTINGS_PATH,default=settings.json"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	_, err := env.UnmarshalFromEnviron(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
