package audio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/vorbis"
	"github.com/gopxl/beep/wav"
)

var namespace = uuid.MustParse("a1536713-cf3b-4c19-8ffb-f2f48d99a4e4")
var supportedFormats = map[string]struct{}{
	".flac": {},
	".ogg":  {},
	".mp3":  {},
	".wav":  {},
}

// isSupportedFormat checks if the file extension is supported for audio tracks.
func isSupportedFormat(ext string) bool {
	_, supported := supportedFormats[strings.ToLower(ext)]
	return supported
}

// Track represents an individual audio track, including its file path, format, and index in the track list.
type Track struct {
	ID     uuid.UUID `json:"id"`     // Is an unique identifier depending on file path to the audio track.
	Path   string    `json:"path"`   // Path is the file path to the audio track.
	Format string    `json:"format"` // Format is the audio format of the track (e.g., mp3, wav).
	Name   string    `json:"name"`   // Name is the file name of the audio track.
}

// NewTrack creates a new Track instance from a given file path.
// It checks if the file exists and determines its format.
func NewTrack(path string) (*Track, error) {
	// Check if the file exists at the specified path.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", path)
	}

	// Extract the file extension and convert it to lower case.
	ext := strings.ToLower(filepath.Ext(path))
	format := ""
	switch ext {
	case ".flac":
		format = "flac"
	case ".ogg":
		format = "ogg"
	case ".mp3":
		format = "mp3"
	case ".wav":
		format = "wav"
	default:
		// Return an error if the file format is not supported.
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	// Extract id & file name from the path.
	id := uuid.NewMD5(namespace, []byte(path))
	name := filepath.Base(path)

	// Return a new Track instance with the determined path, format, index, and name.
	return &Track{ID: id, Path: path, Format: format, Name: name}, nil
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
	case "flac":
		return flac.Decode(f)
	case "ogg":
		return vorbis.Decode(f)
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
