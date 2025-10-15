package bluetooth

import "fmt"

// BLEError represents a Bluetooth-specific error
type BLEError struct {
	Code    string
	Message string
	Cause   error
}

// Error implements the error interface
func (e *BLEError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap implements error unwrapping
func (e *BLEError) Unwrap() error {
	return e.Cause
}

// NewBLEError creates a new BLE error with an optional cause
func NewBLEError(code, message string, cause error) *BLEError {
	return &BLEError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Predefined error types
var (
	ErrInvalidCommand     = &BLEError{Code: "INVALID_COMMAND", Message: "Invalid command format"}
	ErrTrackNotFound      = &BLEError{Code: "TRACK_NOT_FOUND", Message: "Track not found"}
	ErrNoTrackPlaying     = &BLEError{Code: "NO_TRACK_PLAYING", Message: "No track currently playing"}
	ErrUploadInProgress   = &BLEError{Code: "UPLOAD_IN_PROGRESS", Message: "Upload already in progress"}
	ErrInvalidSession     = &BLEError{Code: "INVALID_SESSION", Message: "Invalid upload session"}
	ErrFileUploadFailed   = &BLEError{Code: "FILE_UPLOAD_FAILED", Message: "File upload failed"}
	ErrAdapterNotReady    = &BLEError{Code: "ADAPTER_NOT_READY", Message: "Bluetooth adapter not ready"}
	ErrInvalidPosition    = &BLEError{Code: "INVALID_POSITION", Message: "Invalid playback position"}
	ErrInvalidTrackID     = &BLEError{Code: "INVALID_TRACK_ID", Message: "Invalid track ID format"}
	ErrChecksumMismatch   = &BLEError{Code: "CHECKSUM_MISMATCH", Message: "File checksum verification failed"}
	ErrInvalidTimeFormat  = &BLEError{Code: "INVALID_TIME_FORMAT", Message: "Invalid time format (expected RFC3339)"}
)
