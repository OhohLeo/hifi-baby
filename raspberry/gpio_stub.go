//go:build !rpi

package raspberry

import (
	"errors"

	"github.com/rs/zerolog/log"
)

const (
	StopMusic   = "stop"
	ChangeMusic = "change"
)

// Gpio is a stub implementation for non-Raspberry Pi platforms.
type Gpio struct{}

// NewGpio creates a new Gpio instance.
func NewGpio(chip string, offset int) *Gpio {
	log.Warn().Msg("GPIO is not available on this platform.")
	return &Gpio{}
}

// Listen is a no-op method for non-Raspberry Pi platforms.
func (g *Gpio) Listen(musicControl chan string) error {
	log.Info().Msg("GPIO listener is not started (not a Raspberry Pi platform).")
	return errors.New("GPIO not supported on this platform")
}

// IsGpioEnabled returns false for non-Raspberry Pi platforms.
func IsGpioEnabled() bool {
	return false
}
