package audio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/wav"
)

var supportedFormats = map[string]struct{}{
	".mp3": {},
	".wav": {},
}

// isSupportedFormat checks if the file extension is supported for audio tracks.
func isSupportedFormat(ext string) bool {
	_, supported := supportedFormats[strings.ToLower(ext)]
	return supported
}

// Track represents an individual audio track, including its file path, format, and index in the track list.
type Track struct {
	Path   string `json:"path"`   // Path is the file path to the audio track.
	Format string `json:"format"` // Format is the audio format of the track (e.g., mp3, wav).
	Index  int    `json:"index"`  // Index is the position of the track in the track list.
	Name   string `json:"name"`   // Name is the file name of the audio track.
}

// NewTrack creates a new Track instance from a given file path and index.
// It checks if the file exists and determines its format.
func NewTrack(path string, index int) (*Track, error) {
	// Check if the file exists at the specified path.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", path)
	}

	// Extract the file extension and convert it to lower case.
	ext := strings.ToLower(filepath.Ext(path))
	format := ""
	switch ext {
	case ".mp3":
		format = "mp3"
	case ".wav":
		format = "wav"
	default:
		// Return an error if the file format is not supported.
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	// Extract the file name from the path.
	name := filepath.Base(path)

	// Return a new Track instance with the determined path, format, index, and name.
	return &Track{Path: path, Format: format, Index: index, Name: name}, nil
}

// Open opens the track file.
func (t *Track) Open() (*os.File, error) {
	f, err := os.Open(t.Path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Decode decodes the opened file into a beep.Streamer and beep.Format.
// It takes an opened file as input and uses the track's format to determine how to decode it.
func (t *Track) Decode(f *os.File) (beep.StreamSeekCloser, beep.Format, error) {
	if f == nil {
		return nil, beep.Format{}, fmt.Errorf("file is not open")
	}

	switch t.Format {
	case "mp3":
		return mp3.Decode(f)
	case "wav":
		return wav.Decode(f)
	default:
		return nil, beep.Format{}, fmt.Errorf("unsupported format: %s", t.Format)
	}
}

// Delete removes the audio file associated with the track from the filesystem.
func (t *Track) Delete() error {
	return os.Remove(t.Path)
}
