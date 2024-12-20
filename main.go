package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/OhohLeo/hifi-baby/app"
	"github.com/OhohLeo/hifi-baby/settings"
)

func main() {
	cfg, err := app.NewConfig()
	if err != nil {
		log.Fatal().Msgf("Erreur lors de l'initialisation de la configuration : %v", err)
	}

	settings, err := settings.NewSettings(cfg.SettingsPath)
	if err != nil {
		log.Fatal().Msgf("Erreur lors de l'initialisation du stockage : %v", err)
	}

	level, errLevel := zerolog.ParseLevel(cfg.LogLevel)
	if errLevel != nil {
		log.Fatal().Msgf("Erreur lors de la définition du niveau de log %q : %v", cfg.LogLevel, errLevel)
	}
	zerolog.SetGlobalLevel(level)

	app, errApp := app.NewApp(cfg, settings)
	if errApp != nil {
		log.Fatal().Msgf("Erreur lors de l'initialisation de l'application : %v", errApp)
	}

	if err := app.Run(); err != nil {
		log.Fatal().Msgf("Erreur lors de l'exécution de l'application : %v", err)
	}
}
