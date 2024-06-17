package app

import (
	"github.com/OhohLeo/hifi-baby/audio"
	"github.com/OhohLeo/hifi-baby/http"
)

type App struct {
	Server *http.Server
	Audio  *audio.Audio
}

// NewApp creates a new application instance with initialized components.
func NewApp(cfg *Config) (*App, error) {
	audioInstance, err := audio.NewAudio(cfg.StoragePath)
	if err != nil {
		return nil, err
	}

	server := http.NewServer(audioInstance, cfg.ServerURL)

	// Create the application instance with initialized components
	app := &App{
		Server: server,
		Audio:  audioInstance,
	}

	return app, nil
}

// Run starts the HTTP server and the audio manager.
func (app *App) Run() error {
	// Start the audio management in a goroutine to run it concurrently
	go app.Audio.Run()

	// Start the HTTP server using the Run() method of Server
	return app.Server.Run()
}
