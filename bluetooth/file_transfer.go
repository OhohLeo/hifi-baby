package bluetooth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// UploadSession represents an active file upload session
type UploadSession struct {
	ID           uuid.UUID
	Filename     string
	TotalSize    int64
	Checksum     string
	ChunkSize    int
	ReceivedSize int64
	FilePath     string
	TempFile     *os.File
	StartTime    time.Time
	LastUpdate   time.Time
	mu           sync.RWMutex
}

// FileTransferManager manages file upload sessions
type FileTransferManager struct {
	sessions    map[uuid.UUID]*UploadSession
	storagePath string
	chunkSize   int
	mu          sync.RWMutex
}

// NewFileTransferManager creates a new file transfer manager
func NewFileTransferManager(storagePath string, chunkSize int) *FileTransferManager {
	return &FileTransferManager{
		sessions:    make(map[uuid.UUID]*UploadSession),
		storagePath: storagePath,
		chunkSize:   chunkSize,
	}
}

// InitiateUpload creates a new upload session
func (ftm *FileTransferManager) InitiateUpload(filename string, size int64, checksum string) (*UploadSession, error) {
	ftm.mu.Lock()
	defer ftm.mu.Unlock()

	sessionID := uuid.New()
	tempPath := filepath.Join(ftm.storagePath, fmt.Sprintf(".upload_%s", sessionID.String()))

	// Ensure storage directory exists
	if err := os.MkdirAll(ftm.storagePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	tempFile, err := os.Create(tempPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	session := &UploadSession{
		ID:           sessionID,
		Filename:     filename,
		TotalSize:    size,
		Checksum:     checksum,
		ChunkSize:    ftm.chunkSize,
		ReceivedSize: 0,
		FilePath:     tempPath,
		TempFile:     tempFile,
		StartTime:    time.Now(),
		LastUpdate:   time.Now(),
	}

	ftm.sessions[sessionID] = session

	log.Info().
		Str("session_id", sessionID.String()).
		Str("filename", filename).
		Int64("size", size).
		Msg("Upload session initiated")

	return session, nil
}

// WriteChunk writes a chunk of data to the upload session
func (ftm *FileTransferManager) WriteChunk(sessionID uuid.UUID, chunkIndex uint32, data []byte) (int64, error) {
	ftm.mu.RLock()
	session, exists := ftm.sessions[sessionID]
	ftm.mu.RUnlock()

	if !exists {
		return 0, ErrInvalidSession
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	// Write chunk to file
	n, err := session.TempFile.Write(data)
	if err != nil {
		return 0, fmt.Errorf("failed to write chunk: %w", err)
	}

	session.ReceivedSize += int64(n)
	session.LastUpdate = time.Now()

	log.Debug().
		Str("session_id", sessionID.String()).
		Uint32("chunk_index", chunkIndex).
		Int("chunk_size", n).
		Int64("received", session.ReceivedSize).
		Int64("total", session.TotalSize).
		Msg("Chunk written")

	return session.ReceivedSize, nil
}

// FinalizeUpload completes the upload and moves the file to final location
func (ftm *FileTransferManager) FinalizeUpload(sessionID uuid.UUID) (string, error) {
	ftm.mu.Lock()
	defer ftm.mu.Unlock()

	session, exists := ftm.sessions[sessionID]
	if !exists {
		return "", ErrInvalidSession
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	// Close temp file
	if err := session.TempFile.Close(); err != nil {
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	// Verify checksum if provided
	if session.Checksum != "" {
		if err := ftm.verifyChecksum(session.FilePath, session.Checksum); err != nil {
			os.Remove(session.FilePath)
			return "", err
		}
	}

	// Move file to final location
	finalPath := filepath.Join(ftm.storagePath, session.Filename)
	if err := os.Rename(session.FilePath, finalPath); err != nil {
		return "", fmt.Errorf("failed to move file: %w", err)
	}

	duration := time.Since(session.StartTime)
	log.Info().
		Str("session_id", sessionID.String()).
		Str("filename", session.Filename).
		Int64("size", session.ReceivedSize).
		Dur("duration", duration).
		Msg("Upload finalized successfully")

	delete(ftm.sessions, sessionID)
	return finalPath, nil
}

// CancelUpload cancels an upload session and cleans up
func (ftm *FileTransferManager) CancelUpload(sessionID uuid.UUID) error {
	ftm.mu.Lock()
	defer ftm.mu.Unlock()

	session, exists := ftm.sessions[sessionID]
	if !exists {
		return ErrInvalidSession
	}

	session.TempFile.Close()
	os.Remove(session.FilePath)
	delete(ftm.sessions, sessionID)

	log.Info().
		Str("session_id", sessionID.String()).
		Msg("Upload session canceled")

	return nil
}

// GetSession retrieves an upload session
func (ftm *FileTransferManager) GetSession(sessionID uuid.UUID) (*UploadSession, bool) {
	ftm.mu.RLock()
	defer ftm.mu.RUnlock()

	session, exists := ftm.sessions[sessionID]
	return session, exists
}

// verifyChecksum verifies the SHA256 checksum of a file
func (ftm *FileTransferManager) verifyChecksum(filePath string, expectedChecksum string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	actualChecksum := hex.EncodeToString(hash.Sum(nil))
	if actualChecksum != expectedChecksum {
		return NewBLEError(
			"CHECKSUM_MISMATCH",
			fmt.Sprintf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum),
			nil,
		)
	}

	return nil
}

// GetProgress returns the current upload progress
func (s *UploadSession) GetProgress() UploadProgressResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	percent := 0.0
	if s.TotalSize > 0 {
		percent = float64(s.ReceivedSize) / float64(s.TotalSize) * 100.0
	}

	return UploadProgressResponse{
		Received: s.ReceivedSize,
		Total:    s.TotalSize,
		Percent:  percent,
	}
}
