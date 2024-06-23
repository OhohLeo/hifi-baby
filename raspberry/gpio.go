package raspberry

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/warthog618/go-gpiocdev"
)

type Gpio struct {
	chip           string
	offset         int
	line           *gpiocdev.Line
	lastEventsTime []time.Time
}

// NewGpio creates a new Gpio instance
func NewGpio(chip string, offset int) *Gpio {
	return &Gpio{
		chip:   chip,
		offset: offset,
	}
}

const (
	StopMusic   = "stop"
	ChangeMusic = "change"
)

func (g *Gpio) Listen(musicControl chan string) error {
	log.Info().Msg("Listening to GPIO events")

	var err error
	g.line, err = gpiocdev.RequestLine(
		g.chip,
		g.offset,
		gpiocdev.WithBothEdges,
		gpiocdev.WithPullUp,
		gpiocdev.WithDebounce(50*time.Millisecond),
		gpiocdev.WithEventHandler(
			func(evt gpiocdev.LineEvent) {
				log.Debug().Msg("GPIO event detected")
				currentTime := time.Now()
				g.lastEventsTime = append(g.lastEventsTime, currentTime)

				// Drop all last events that occurred more than 1 second ago
				oneSecondsAgo := time.Now().Add(-1 * time.Second)
				var recentEvents []time.Time
				for _, eventTime := range g.lastEventsTime {
					if eventTime.After(oneSecondsAgo) {
						recentEvents = append(recentEvents, eventTime)
					}
				}

				g.lastEventsTime = recentEvents

				if len(recentEvents) == 0 {
					return
				}
				// Determine the action based on number of recent events
				if len(recentEvents) >= 2 {
					log.Info().Msg("Stopping music")
					musicControl <- StopMusic
					g.lastEventsTime = []time.Time{}
				} else {
					log.Info().Msg("Changing music")
					musicControl <- ChangeMusic
				}
			},
		),
	)

	return err
}
