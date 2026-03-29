package audio

import (
	"fmt"

	"github.com/gen2brain/malgo"
	"github.com/rs/zerolog/log"
)

type AudioSink struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func ListAudioSinks() ([]AudioSink, error) {
	context, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		log.Info().Msg(message)
	})
	if err != nil {
		return nil, err
	}
	defer context.Uninit()
	defer context.Free()

	devices, err := context.Devices(malgo.Playback)
	if err != nil {
		return nil, err
	}

	audioSinks := make([]AudioSink, len(devices))
	for i, deviceInfo := range devices {
		audioSinks[i] = AudioSink{
			Name:        deviceInfo.Name(),
			Description: deviceInfo.Name(),
		}
	}

	return audioSinks, nil
}

func SetDefaultSink(sinkName string) error {
	// malgo does not have a concept of a system-wide default sink.
	// The sink is selected when initializing a device.
	// This function is now a placeholder.
	log.Info().Msgf("SetDefaultSink is a placeholder with malgo. Sink is selected on device initialization.")
	return nil
}

func GetDefaultSink() (string, error) {
	// malgo does not have a concept of a system-wide default sink.
	// This function is now a placeholder.
	return "", fmt.Errorf("GetDefaultSink is not supported by malgo")
}
