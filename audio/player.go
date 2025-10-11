package audio

import (
	"time"
)

// PlayerState represents the current state of an audio track playback.
type PlayerState struct {
	CurrentTrack    *Track        `json:"currentTrack"` // CurrentTrack is the audio track currently being played.
	IsPlaying       bool          `json:"isPlaying"`    // IsPlaying indicates whether the track playback is active.
	IsMuted         bool          `json:"isMuted"`      // IsMuted indicates whether the sound is muted.
	SecondsPosition int           `json:"position"`     // SecondsPosition is the current playback position in seconds in the track.
	position        time.Duration // Position is the current playback position in the track.
}

// InitializeTrack initializes the current track and resets the elapsed and total time.
func (ps *PlayerState) InitializeTrack(track *Track, elapsedTime, duration time.Duration) {
	ps.CurrentTrack = track
	ps.IsPlaying = true
	ps.SecondsPosition = 0
	ps.position = 0
}

// UpdatePosition updates the current playback position.
func (ps *PlayerState) UpdatePosition(position time.Duration) {
	ps.SecondsPosition = int(position.Seconds())
	ps.position = position
}

// StopTrack stops the playback and resets the track and time information.
func (ps *PlayerState) StopTrack() {
	ps.CurrentTrack = nil
	ps.IsPlaying = false
	ps.SecondsPosition = 0
	ps.position = 0
}
