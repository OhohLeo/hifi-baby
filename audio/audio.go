package audio

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/speaker"
	"github.com/rs/zerolog/log"
)

type Config struct {
	StoragePath string `env:"STORAGE_PATH,default=tracks"`
}

type StoredConfig struct {
	BaseVolume    float64 `json:"base_volume"`
	DefaultVolume float64 `json:"default_volume"`
	MinVolume     float64 `json:"min_volume"`
	MaxVolume     float64 `json:"max_volume"`
	VolumeStep    float64 `json:"volume_step"`
	SilentEnabled bool    `json:"silent_enabled"`
}

type Capabilities interface {
	AddListenedTrack(track *Track, when time.Time, during int64) error
}

// Audio manages a list of audio tracks, playback state, volume control, and storage path.
type Audio struct {
	tracks       []*Track        // tracks holds a slice of all available tracks.
	activeStream *beep.Ctrl      // ctrlStream controls the pause and resume of the active stream
	volume       *effects.Volume // volume controls the volume of the playback.
	storagePath  string          // storagePath is the base path where audio files are stored.
	playRequests chan int        // playRequests is a channel for play requests
	stopChan     chan bool       // stopChan is a channel to signal stop
	playerState  PlayerState     // playerState holds the current state of the audio player.
	storedConfig StoredConfig    // storedConfig holds the stored configuration for the audio player.
	capabilities Capabilities
}

// NewAudio creates a new Audio instance with a given list of track paths and a storage path.
func NewAudio(
	config Config,
	storedConfig StoredConfig,
	capabilities Capabilities,
) (*Audio, error) {
	storagePath := config.StoragePath
	audio := &Audio{
		tracks: []*Track{},
		volume: &effects.Volume{
			Base:   storedConfig.BaseVolume,
			Volume: storedConfig.DefaultVolume,
			Silent: storedConfig.SilentEnabled,
		},
		storagePath:  storagePath,
		playRequests: make(chan int),
		stopChan:     make(chan bool),
		storedConfig: storedConfig,
		capabilities: capabilities,
	}

	// Ensure the directory exists or create it
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		if err := os.MkdirAll(storagePath, 0755); err != nil {
			return nil, err
		}
	}

	// Lire tous les fichiers .mp3 et .wav dans le répertoire storagePath
	err := filepath.Walk(storagePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if isSupportedFormat(ext) {
			if _, err := audio.addTrack(path); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Initialisation du haut-parleur avec le format décodé
	sampleRate := beep.SampleRate(44100)
	if err := speaker.Init(sampleRate, sampleRate.N(time.Second/5)); err != nil {
		return nil, fmt.Errorf("speaker issue : %v", err)
	}

	return audio, nil
}

// checkTrackIndex checks if the provided index is within the bounds of the track list.
func (a *Audio) checkTrackIndex(index int) error {
	if index < 0 || index >= len(a.tracks) {
		return errors.New("index out of range")
	}
	return nil
}

// AddTrack appends a new track to the audio manager, determining its index based on the current list size.
// It returns the newly created track and any error encountered.
func (a *Audio) AddTrack(file multipart.File, header *multipart.FileHeader) (*Track, error) {
	fullPath := filepath.Join(a.storagePath, header.Filename)

	// Vérifier l'extension du fichier
	ext := filepath.Ext(header.Filename)
	if !isSupportedFormat(ext) {
		return nil, fmt.Errorf("unsupported file type '%s'", ext)
	}

	// Construire le chemin complet où le fichier sera sauvegardé
	filePath := path.Join(a.storagePath, header.Filename)

	// Créer le fichier sur le disque
	out, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to create the file '%s'", filePath)
	}
	defer out.Close()

	// Copier le contenu du fichier téléchargé dans le nouveau fichier
	if _, err := io.Copy(out, file); err != nil {
		return nil, fmt.Errorf("failed to save the file '%s'", filePath)
	}

	return a.addTrack(fullPath)
}

func (a *Audio) addTrack(path string) (*Track, error) {
	newTrack, err := NewTrack(path, len(a.tracks))
	if err != nil {
		return nil, err
	}
	a.tracks = append(a.tracks, newTrack)
	return newTrack, nil
}

// RemoveTrack removes a track from the list by index and handles playback and file deletion.
func (a *Audio) RemoveTrack(index int) error {
	if err := a.checkTrackIndex(index); err != nil {
		return err
	}

	trackToRemove := a.tracks[index]

	// Stop playback if the track to be removed is currently playing.
	if a.activeStream != nil && trackToRemove == a.playerState.CurrentTrack {
		a.Stop()
	}

	// Delete the track file from the filesystem.
	if err := trackToRemove.Delete(); err != nil {
		return err
	}

	// Remove the track from the list.
	a.tracks = append(a.tracks[:index], a.tracks[index+1:]...)
	return nil
}

// GetPlayerState returns the current state of the audio player.
func (a *Audio) GetPlayerState() PlayerState {
	return a.playerState
}

// Tracks returns a slice of all available tracks.
func (a *Audio) Tracks() []*Track {
	return a.tracks
}

// PlayRandomTrack selects a random track and plays it.
func (a *Audio) PlayRandomTrack() {
	randomIndex := rand.Intn(len(a.tracks))
	a.PlayTrack(randomIndex)
}

// Play a specific track from the track list based on the index.
func (a *Audio) PlayTrack(index int) {
	// Stop the currently playing track if it exists
	if a.activeStream != nil && !a.activeStream.Paused {
		a.Stop()
	}

	log.Info().Msgf("Playing track %d: %s", index, a.tracks[index].Path)

	// Send the play request for the new track
	a.playRequests <- index
}

// Pause the currently playing track if it is not already paused.
func (a *Audio) Pause() {
	if a.activeStream == nil || a.activeStream.Paused {
		return
	}
	speaker.Lock()
	defer speaker.Unlock()
	a.activeStream.Paused = true
	a.playerState.IsPlaying = false
}

// Resume the playback of the currently paused track if it is paused.
func (a *Audio) Resume() {
	if a.activeStream == nil || !a.activeStream.Paused {
		return
	}
	speaker.Lock()
	defer speaker.Unlock()
	a.activeStream.Paused = false
	a.playerState.IsPlaying = true
}

// IncreaseVolume increases the audio volume.
func (a *Audio) IncreaseVolume() {
	speaker.Lock()
	defer speaker.Unlock()

	a.volume.Volume += a.storedConfig.VolumeStep
	if a.volume.Volume > a.storedConfig.MaxVolume {
		a.volume.Volume = a.storedConfig.MaxVolume
	}
}

// DecreaseVolume decreases the audio volume.
func (a *Audio) DecreaseVolume() {
	speaker.Lock()
	defer speaker.Unlock()

	a.volume.Volume -= a.storedConfig.VolumeStep
	if a.volume.Volume < a.storedConfig.MinVolume {
		a.volume.Volume = a.storedConfig.MinVolume
	}
}

// Mute mutes the currently playing audio.
func (a *Audio) Mute(enable bool) {
	speaker.Lock()
	defer speaker.Unlock()

	a.volume.Silent = enable
	a.playerState.IsMuted = enable
}

func (a *Audio) Run() {
	log.Info().Msg("Audio manager started")
	for index := range a.playRequests {
		if err := a.playTrack(index); err != nil {
			log.Error().Msgf("Error playing track: %v", err)
			continue
		}
	}
}

func (a *Audio) playTrack(index int) error {
	if index < 0 || index >= len(a.tracks) {
		return fmt.Errorf("index out of range")
	}
	track := a.tracks[index]

	// Open the track file
	file, err := track.Open()
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Decode the opened file
	streamer, format, err := track.Decode(file)
	if err != nil {
		return fmt.Errorf("error during decoding: %v", err)
	}
	defer streamer.Close()

	a.activeStream = &beep.Ctrl{Streamer: streamer, Paused: false}
	a.volume.Streamer = a.activeStream
	startTime := time.Now() // Start time of the track
	a.playerState.InitializeTrack(
		track,
		format.SampleRate.D(streamer.Position()).Round(time.Second),
		format.SampleRate.D(streamer.Len()).Round(time.Second),
	)
	defer func() {
		a.activeStream = nil
		a.volume.Streamer = nil
		a.playerState.StopTrack()
		duration := int64(time.Since(startTime).Seconds())

		if err := a.capabilities.AddListenedTrack(track, startTime, duration); err != nil {
			log.Error().Msgf("Error adding listened track: %v", err)
		}
	}()

	log.Info().Msgf("Playing track: %s\n", track.Path)

	speaker.Play(beep.Seq(a.volume, beep.Callback(func() {
		a.stopChan <- true
	})))
	defer speaker.Clear()

	for {
		select {
		case <-a.stopChan:
			a.playerState.StopTrack()
			log.Info().Msgf("Stopped playing track: %s\n", track.Path)
			return nil
		case <-time.After(time.Second):
			speaker.Lock()
			if posStreamer, ok := streamer.(interface{ Position() int }); ok {
				elapsedTime := format.SampleRate.D(posStreamer.Position()).Round(time.Second)
				log.Info().Msgf("Position: %s", elapsedTime)
			} else {
				log.Error().Msg("The streamer does not support the Position method")
			}
			speaker.Unlock()
		}
	}
}

// Stop any currently playing track and resets playback state.
func (a *Audio) Stop() {
	if a.activeStream != nil {
		a.stopChan <- true // Send a signal to stop playback
	}
}

// Close the playRequests channel
func (a *Audio) Close() {
	close(a.playRequests)
	close(a.stopChan)
}
