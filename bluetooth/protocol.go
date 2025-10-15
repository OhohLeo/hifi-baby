package bluetooth

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/OhohLeo/hifi-baby/audio"
)

// Response is the generic response wrapper for all BLE operations
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// TrackControlCommand represents playback control commands
type TrackControlCommand struct {
	Command string     `json:"command"` // play, pause, resume, stop
	TrackID *uuid.UUID `json:"track_id,omitempty"`
}

// PositionCommand represents position get/set operations
type PositionCommand struct {
	Position float64 `json:"position"` // seconds
}

// PositionResponse contains current playback position
type PositionResponse struct {
	Position float64 `json:"position"` // seconds
}

// VolumeCommand represents volume control operations
type VolumeCommand struct {
	Command string `json:"command"` // increase, decrease, mute
	Enable  *bool  `json:"enable,omitempty"`
}

// PlayerStateResponse contains the current player state
type PlayerStateResponse struct {
	Track     *audio.Track `json:"track"`
	IsPlaying bool         `json:"isPlaying"`
	IsMuted   bool         `json:"isMuted"`
	Position  float64      `json:"position"`  // seconds
	Duration  float64      `json:"duration"`  // seconds
}

// TrackListResponse contains a list of all tracks
type TrackListResponse struct {
	Tracks []*audio.Track `json:"tracks"`
	Total  int            `json:"total"`
	Offset int            `json:"offset"`
	Limit  int            `json:"limit"`
}

// UploadControlCommand manages file upload lifecycle
type UploadControlCommand struct {
	Command   string `json:"command"` // init, finalize, cancel
	Filename  string `json:"filename,omitempty"`
	Size      int64  `json:"size,omitempty"`
	Checksum  string `json:"checksum,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// UploadInitResponse contains upload session details
type UploadInitResponse struct {
	SessionID string `json:"session_id"`
	ChunkSize int    `json:"chunk_size"`
	MaxChunks int    `json:"max_chunks"`
}

// UploadProgressResponse tracks upload progress
type UploadProgressResponse struct {
	Received int64   `json:"received"`
	Total    int64   `json:"total"`
	Percent  float64 `json:"percent"`
}

// UploadFinalizeResponse contains the created track info
type UploadFinalizeResponse struct {
	Track *audio.Track `json:"track"`
}

// TrackDeleteCommand deletes a track by ID
type TrackDeleteCommand struct {
	TrackID string `json:"track_id"`
}

// StatisticsQuery contains parameters for statistics queries
type StatisticsQuery struct {
	Since string `json:"since"` // RFC3339 format
	TopNb *int   `json:"top_nb,omitempty"`
}

// ListenedTrack represents a single listening event
type ListenedTrack struct {
	TrackName string    `json:"track_name"`
	At        time.Time `json:"at"`
	During    int64     `json:"during"` // seconds
}

// ListenedTracksResponse contains listened track history
type ListenedTracksResponse struct {
	Tracks []ListenedTrack `json:"tracks"`
}

// MostListenedTrack represents aggregate listening statistics
type MostListenedTrack struct {
	TrackName string    `json:"track_name"`
	Since     time.Time `json:"since"`
	During    int64     `json:"during"` // total seconds
	Count     int       `json:"count"`
}

// MostListenedTracksResponse contains most played tracks
type MostListenedTracksResponse struct {
	Tracks []MostListenedTrack `json:"tracks"`
}

// EncodeResponse creates a JSON response with success/error information
func EncodeResponse(success bool, data interface{}, err error) []byte {
	resp := Response{
		Success: success,
		Data:    data,
	}
	if err != nil {
		resp.Error = err.Error()
	}

	bytes, _ := json.Marshal(resp)
	return bytes
}

// DecodeCommand unmarshals JSON command data into target structure
func DecodeCommand(data []byte, target interface{}) error {
	return json.Unmarshal(data, target)
}
