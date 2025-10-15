package app

import (
	"github.com/rs/zerolog/log"

	"github.com/OhohLeo/hifi-baby/audio"
	"github.com/OhohLeo/hifi-baby/bluetooth"
	"github.com/OhohLeo/hifi-baby/http"
	"github.com/OhohLeo/hifi-baby/raspberry"
	"github.com/OhohLeo/hifi-baby/settings"
	"github.com/OhohLeo/hifi-baby/sql"
)

type App struct {
	Server          *http.Server
	BluetoothServer *bluetooth.Server
	Audio           *audio.Audio
	Gpio            *raspberry.Gpio
	Database        *sql.Database
}

// NewApp creates a new application instance with initialized components.
func NewApp(cfg *Config, settings *settings.Settings) (*App, error) {
	database, err := sql.NewDatabase(cfg.Database)
	if err != nil {
		return nil, err
	}

	audioInstance, err := audio.NewAudio(
		cfg.Audio,
		settings.Audio,
		database,
	)
	if err != nil {
		return nil, err
	}

	server := http.NewServer(audioInstance, cfg.Server, settings, database)

	// Initialize Bluetooth server
	bleServer, err := bluetooth.NewServer(audioInstance, settings, database, cfg.Bluetooth)
	if err != nil {
		return nil, err
	}

	app := &App{
		Server:          server,
		BluetoothServer: bleServer,
		Audio:           audioInstance,
		Gpio:            raspberry.NewGpio("gpiochip0", 16),
		Database:        database,
	}

	return app, nil
}

func (app *App) listenToGpio() {
	actions := make(chan string)
	defer close(actions)

	go func() {
		err := app.Gpio.Listen(actions)
		if err != nil {
			log.Error().Err(err).Msg("Error listening to GPIO events")
		}
	}()

	for action := range actions {
		log.Info().Msgf("Action: %v", action)
		switch action {
		case raspberry.StopMusic:
			app.Audio.Stop()
		case raspberry.ChangeMusic:
			app.Audio.PlayRandomTrack()
		}

	}
}

// Run starts the HTTP server, Bluetooth server, and the audio manager.
func (app *App) Run() error {
	// Start listening to GPIO events
	go app.listenToGpio()

	// Start the audio management in a goroutine to run it concurrently
	go app.Audio.Run()

	// Start the Bluetooth server in a goroutine
	go func() {
		if err := app.BluetoothServer.Run(); err != nil {
			log.Error().Err(err).Msg("Bluetooth server error")
		}
	}()

	// Start the HTTP server using the Run() method of Server (blocking)
	return app.Server.Run()
}
