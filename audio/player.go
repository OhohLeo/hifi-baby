package audio

import (
	"encoding/json"
	"time"
)

// PlayerState represents the current state of an audio track playback.
type PlayerState struct {
	CurrentTrack *Track        `json:"track"`
	IsPlaying    bool          `json:"isPlaying"`
	IsMuted      bool          `json:"isMuted"`
	Position     time.Duration `json:"position"`
	Duration     time.Duration `json:"duration"`
}

// MarshalJSON is a custom marshaler for PlayerState to convert durations to seconds.
func (ps *PlayerState) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		CurrentTrack *Track  `json:"track"`
		IsPlaying    bool    `json:"isPlaying"`
		IsMuted      bool    `json:"isMuted"`
		Position     float64 `json:"position"`
		Duration     float64 `json:"duration"`
	}{
		CurrentTrack: ps.CurrentTrack,
		IsPlaying:    ps.IsPlaying,
		IsMuted:      ps.IsMuted,
		Position:     ps.Position.Seconds(),
		Duration:     ps.Duration.Seconds(),
	})
}

// InitializeTrack sets up the player state for a new track.
func (ps *PlayerState) InitializeTrack(track *Track, elapsedTime, duration time.Duration) {
	ps.CurrentTrack = track
	ps.IsPlaying = true
	ps.Position = elapsedTime
	ps.Duration = duration
}

// UpdatePosition updates the playback position of the current track.
func (ps *PlayerState) UpdatePosition(position time.Duration) {
	ps.Position = position
}

// StopTrack resets the player state when a track stops playing.
func (ps *PlayerState) StopTrack() {
	ps.CurrentTrack = nil
	ps.IsPlaying = false
	ps.Position = 0
	ps.Duration = 0
}