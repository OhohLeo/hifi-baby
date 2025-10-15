package bluetooth

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"tinygo.org/x/bluetooth"

	"github.com/OhohLeo/hifi-baby/audio"
	"github.com/OhohLeo/hifi-baby/settings"
	"github.com/OhohLeo/hifi-baby/sql"
)

// CharacteristicHandler manages all GATT characteristic operations
type CharacteristicHandler struct {
	audio        *audio.Audio
	settings     *settings.Settings
	database     *sql.Database
	ftManager    *FileTransferManager
	notifyMgr    *NotificationManager
	config       Config
	mu           sync.RWMutex
	ctx          context.Context

	// State for query-based characteristics
	listenedTracksQuery StatisticsQuery
	mostListenedQuery   StatisticsQuery
}

// NewCharacteristicHandler creates a new characteristic handler
func NewCharacteristicHandler(
	ctx context.Context,
	audio *audio.Audio,
	settings *settings.Settings,
	database *sql.Database,
	config Config,
) *CharacteristicHandler {
	return &CharacteristicHandler{
		audio:     audio,
		settings:  settings,
		database:  database,
		ftManager: NewFileTransferManager(config.StoragePath, config.ChunkSize),
		notifyMgr: NewNotificationManager(ctx),
		config:    config,
		ctx:       ctx,
	}
}

// Track Control Characteristic Handlers

func (ch *CharacteristicHandler) HandleTrackControlWrite(connection bluetooth.Connection, offset int, value []byte) {
	var cmd TrackControlCommand
	if err := DecodeCommand(value, &cmd); err != nil {
		log.Error().Err(err).Msg("Failed to decode track control command")
		return
	}

	switch cmd.Command {
	case "play":
		if cmd.TrackID == nil {
			log.Error().Msg("Track ID required for play command")
			return
		}
		ch.audio.PlayTrack(*cmd.TrackID)
		log.Info().Str("track_id", cmd.TrackID.String()).Msg("Playing track via BLE")

	case "pause":
		ch.audio.Pause()
		log.Info().Msg("Paused playback via BLE")

	case "resume":
		ch.audio.Resume()
		log.Info().Msg("Resumed playback via BLE")

	case "stop":
		ch.audio.Stop()
		log.Info().Msg("Stopped playback via BLE")

	default:
		log.Error().Str("command", cmd.Command).Msg("Unknown track control command")
	}
}

// Track Position Characteristic Handlers

func (ch *CharacteristicHandler) HandleTrackPositionRead(connection bluetooth.Connection, offset int) ([]byte, error) {
	state := ch.audio.GetPlayerState()
	data := PositionResponse{
		Position: state.Position.Seconds(),
	}
	return EncodeResponse(true, data, nil), nil
}

func (ch *CharacteristicHandler) HandleTrackPositionWrite(connection bluetooth.Connection, offset int, value []byte) {
	var cmd PositionCommand
	if err := DecodeCommand(value, &cmd); err != nil {
		log.Error().Err(err).Msg("Failed to decode position command")
		return
	}

	position := time.Duration(cmd.Position) * time.Second
	if err := ch.audio.SetPosition(position); err != nil {
		log.Error().Err(err).Float64("position", cmd.Position).Msg("Failed to set position")
		return
	}

	log.Info().Float64("position", cmd.Position).Msg("Set playback position via BLE")
}

// Volume Control Characteristic Handler

func (ch *CharacteristicHandler) HandleVolumeControlWrite(connection bluetooth.Connection, offset int, value []byte) {
	var cmd VolumeCommand
	if err := DecodeCommand(value, &cmd); err != nil {
		log.Error().Err(err).Msg("Failed to decode volume command")
		return
	}

	switch cmd.Command {
	case "increase":
		ch.audio.IncreaseVolume()
		log.Info().Msg("Increased volume via BLE")

	case "decrease":
		ch.audio.DecreaseVolume()
		log.Info().Msg("Decreased volume via BLE")

	case "mute":
		if cmd.Enable == nil {
			log.Error().Msg("Enable parameter required for mute command")
			return
		}
		ch.audio.Mute(*cmd.Enable)
		log.Info().Bool("enabled", *cmd.Enable).Msg("Set mute via BLE")

	default:
		log.Error().Str("command", cmd.Command).Msg("Unknown volume command")
	}
}

// Player State Characteristic Handler

func (ch *CharacteristicHandler) HandlePlayerStateRead(connection bluetooth.Connection, offset int) ([]byte, error) {
	state := ch.audio.GetPlayerState()

	resp := PlayerStateResponse{
		Track:     state.CurrentTrack,
		IsPlaying: state.IsPlaying,
		IsMuted:   state.IsMuted,
		Position:  state.Position.Seconds(),
		Duration:  state.Duration.Seconds(),
	}

	return EncodeResponse(true, resp, nil), nil
}

// Track List Characteristic Handler

func (ch *CharacteristicHandler) HandleTrackListRead(connection bluetooth.Connection, offset int) ([]byte, error) {
	tracks := ch.audio.Tracks()

	resp := TrackListResponse{
		Tracks: tracks,
		Total:  len(tracks),
		Offset: 0,
		Limit:  len(tracks),
	}

	return EncodeResponse(true, resp, nil), nil
}

// Upload Control Characteristic Handler

func (ch *CharacteristicHandler) HandleUploadControlWrite(connection bluetooth.Connection, offset int, value []byte) {
	var cmd UploadControlCommand
	if err := DecodeCommand(value, &cmd); err != nil {
		log.Error().Err(err).Msg("Failed to decode upload control command")
		return
	}

	switch cmd.Command {
	case "init":
		session, err := ch.ftManager.InitiateUpload(cmd.Filename, cmd.Size, cmd.Checksum)
		if err != nil {
			log.Error().Err(err).Str("filename", cmd.Filename).Msg("Failed to initiate upload")
			return
		}

		resp := UploadInitResponse{
			SessionID: session.ID.String(),
			ChunkSize: session.ChunkSize,
			MaxChunks: int(session.TotalSize/int64(session.ChunkSize)) + 1,
		}

		respData := EncodeResponse(true, resp, nil)
		log.Info().
			Str("session_id", session.ID.String()).
			Str("filename", cmd.Filename).
			Int64("size", cmd.Size).
			Msg("Upload session initiated")

		// Send response (this is a simplified approach; proper response handling depends on TinyGo implementation)
		log.Debug().Bytes("response", respData).Msg("Upload init response")

	case "finalize":
		sessionID, err := uuid.Parse(cmd.SessionID)
		if err != nil {
			log.Error().Err(err).Str("session_id", cmd.SessionID).Msg("Invalid session ID")
			return
		}

		filePath, err := ch.ftManager.FinalizeUpload(sessionID)
		if err != nil {
			log.Error().Err(err).Str("session_id", cmd.SessionID).Msg("Failed to finalize upload")
			return
		}

		// Add track to audio manager
		track, err := ch.audio.AddTrackFromPath(filePath)
		if err != nil {
			log.Error().Err(err).Str("file_path", filePath).Msg("Failed to add track")
			return
		}

		resp := UploadFinalizeResponse{
			Track: track,
		}

		respData := EncodeResponse(true, resp, nil)
		log.Info().
			Str("session_id", cmd.SessionID).
			Str("track_id", track.ID.String()).
			Str("filename", track.Name).
			Msg("Upload finalized successfully")

		log.Debug().Bytes("response", respData).Msg("Upload finalize response")

	case "cancel":
		sessionID, err := uuid.Parse(cmd.SessionID)
		if err != nil {
			log.Error().Err(err).Str("session_id", cmd.SessionID).Msg("Invalid session ID")
			return
		}

		if err := ch.ftManager.CancelUpload(sessionID); err != nil {
			log.Error().Err(err).Str("session_id", cmd.SessionID).Msg("Failed to cancel upload")
			return
		}

		log.Info().Str("session_id", cmd.SessionID).Msg("Upload canceled")

	default:
		log.Error().Str("command", cmd.Command).Msg("Unknown upload control command")
	}
}

// Upload Data Characteristic Handler

func (ch *CharacteristicHandler) HandleUploadDataWrite(connection bluetooth.Connection, offset int, value []byte) {
	if len(value) < 12 {
		log.Error().Int("length", len(value)).Msg("Invalid chunk format: too short")
		return
	}

	// Parse binary chunk format:
	// [0-3]   uint32 session_id_hash
	// [4-7]   uint32 chunk_index
	// [8-11]  uint32 chunk_size
	// [12-end] binary data
	sessionIDHash := binary.BigEndian.Uint32(value[0:4])
	chunkIndex := binary.BigEndian.Uint32(value[4:8])
	chunkSize := binary.BigEndian.Uint32(value[8:12])
	chunkData := value[12:]

	// Find session by hash (simplified - in production, maintain a hash-to-UUID map)
	var sessionID uuid.UUID
	ch.ftManager.mu.RLock()
	for id := range ch.ftManager.sessions {
		// Use first 4 bytes of UUID as hash
		idBytes, _ := id.MarshalBinary()
		if binary.BigEndian.Uint32(idBytes[0:4]) == sessionIDHash {
			sessionID = id
			break
		}
	}
	ch.ftManager.mu.RUnlock()

	if sessionID == uuid.Nil {
		log.Error().Uint32("hash", sessionIDHash).Msg("Session not found for hash")
		return
	}

	if int(chunkSize) != len(chunkData) {
		log.Warn().
			Uint32("expected", chunkSize).
			Int("actual", len(chunkData)).
			Msg("Chunk size mismatch")
	}

	received, err := ch.ftManager.WriteChunk(sessionID, chunkIndex, chunkData)
	if err != nil {
		log.Error().
			Err(err).
			Str("session_id", sessionID.String()).
			Uint32("chunk_index", chunkIndex).
			Msg("Failed to write chunk")
		return
	}

	session, _ := ch.ftManager.GetSession(sessionID)
	progress := session.GetProgress()

	log.Debug().
		Str("session_id", sessionID.String()).
		Uint32("chunk_index", chunkIndex).
		Int64("received", received).
		Float64("percent", progress.Percent).
		Msg("Chunk received")
}

// Upload Status Characteristic Handler

func (ch *CharacteristicHandler) HandleUploadStatusRead(connection bluetooth.Connection, offset int) ([]byte, error) {
	// Return status of active upload sessions
	ch.ftManager.mu.RLock()
	defer ch.ftManager.mu.RUnlock()

	// For simplicity, return the first active session if any
	for _, session := range ch.ftManager.sessions {
		progress := session.GetProgress()
		return EncodeResponse(true, progress, nil), nil
	}

	// No active sessions
	return EncodeResponse(true, nil, nil), nil
}

// Track Delete Characteristic Handler

func (ch *CharacteristicHandler) HandleTrackDeleteWrite(connection bluetooth.Connection, offset int, value []byte) {
	var cmd TrackDeleteCommand
	if err := DecodeCommand(value, &cmd); err != nil {
		log.Error().Err(err).Msg("Failed to decode track delete command")
		return
	}

	trackID, err := uuid.Parse(cmd.TrackID)
	if err != nil {
		log.Error().Err(err).Str("track_id", cmd.TrackID).Msg("Invalid track ID")
		return
	}

	if err := ch.audio.RemoveTrack(trackID); err != nil {
		log.Error().Err(err).Str("track_id", cmd.TrackID).Msg("Failed to remove track")
		return
	}

	log.Info().Str("track_id", cmd.TrackID).Msg("Track deleted via BLE")
}

// Listened Tracks Characteristic Handlers

func (ch *CharacteristicHandler) HandleListenedTracksWrite(connection bluetooth.Connection, offset int, value []byte) {
	if err := DecodeCommand(value, &ch.listenedTracksQuery); err != nil {
		log.Error().Err(err).Msg("Failed to decode listened tracks query")
		return
	}

	log.Debug().Str("since", ch.listenedTracksQuery.Since).Msg("Listened tracks query set")
}

func (ch *CharacteristicHandler) HandleListenedTracksRead(connection bluetooth.Connection, offset int) ([]byte, error) {
	sinceTime, err := time.Parse(time.RFC3339, ch.listenedTracksQuery.Since)
	if err != nil {
		return EncodeResponse(false, nil, fmt.Errorf("invalid since time format")), nil
	}

	dbTracks, err := ch.database.ListenedTracks(sinceTime)
	if err != nil {
		return EncodeResponse(false, nil, err), nil
	}

	// Convert database tracks to protocol format
	tracks := make([]ListenedTrack, len(dbTracks))
	for i, t := range dbTracks {
		tracks[i] = ListenedTrack{
			TrackName: t.TrackName,
			At:        t.At,
			During:    t.During,
		}
	}

	resp := ListenedTracksResponse{
		Tracks: tracks,
	}

	return EncodeResponse(true, resp, nil), nil
}

// Most Listened Tracks Characteristic Handlers

func (ch *CharacteristicHandler) HandleMostListenedWrite(connection bluetooth.Connection, offset int, value []byte) {
	if err := DecodeCommand(value, &ch.mostListenedQuery); err != nil {
		log.Error().Err(err).Msg("Failed to decode most listened query")
		return
	}

	log.Debug().
		Str("since", ch.mostListenedQuery.Since).
		Interface("top_nb", ch.mostListenedQuery.TopNb).
		Msg("Most listened query set")
}

func (ch *CharacteristicHandler) HandleMostListenedRead(connection bluetooth.Connection, offset int) ([]byte, error) {
	sinceTime, err := time.Parse(time.RFC3339, ch.mostListenedQuery.Since)
	if err != nil {
		return EncodeResponse(false, nil, fmt.Errorf("invalid since time format")), nil
	}

	topNb := 10 // default
	if ch.mostListenedQuery.TopNb != nil {
		topNb = *ch.mostListenedQuery.TopNb
	}

	dbTracks, err := ch.database.MostListenedTracks(sinceTime, topNb)
	if err != nil {
		return EncodeResponse(false, nil, err), nil
	}

	// Convert database tracks to protocol format
	tracks := make([]MostListenedTrack, len(dbTracks))
	for i, t := range dbTracks {
		// Parse the Since field from string to time.Time
		sinceTime, _ := time.Parse(time.RFC3339, t.Since)
		tracks[i] = MostListenedTrack{
			TrackName: t.TrackName,
			Since:     sinceTime,
			During:    t.During,
			Count:     t.Count,
		}
	}

	resp := MostListenedTracksResponse{
		Tracks: tracks,
	}

	return EncodeResponse(true, resp, nil), nil
}

// Settings Characteristic Handlers

func (ch *CharacteristicHandler) HandleSettingsRead(connection bluetooth.Connection, offset int) ([]byte, error) {
	return EncodeResponse(true, ch.settings, nil), nil
}

func (ch *CharacteristicHandler) HandleSettingsWrite(connection bluetooth.Connection, offset int, value []byte) {
	var newSettings settings.Settings
	if err := json.Unmarshal(value, &newSettings); err != nil {
		log.Error().Err(err).Msg("Failed to decode settings")
		return
	}

	if err := ch.settings.Update(&newSettings); err != nil {
		log.Error().Err(err).Msg("Failed to update settings")
		return
	}

	log.Info().Msg("Settings updated via BLE")
}

// GetNotificationManager returns the notification manager for setting up notifications
func (ch *CharacteristicHandler) GetNotificationManager() *NotificationManager {
	return ch.notifyMgr
}

// Close cleans up the characteristic handler
func (ch *CharacteristicHandler) Close() {
	ch.notifyMgr.Close()
}
