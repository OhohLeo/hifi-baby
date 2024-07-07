package sql

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/OhohLeo/hifi-baby/audio"
)

// Config represents the database configuration.
type Config struct {
	Path    string        `env:"DATABASE_PATH,default=./hifi-baby.db"`
	Timeout time.Duration `env:"DATABASE_TIMEOUT,default=10s"`
}

// Database handles the database.
type Database struct {
	orm *gorm.DB
}

// NewDatabase creates a new database.
func NewDatabase(cfg Config) (*Database, error) {
	orm, err := gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gorm: %w", err)
	}

	if err := orm.AutoMigrate(&ListenedTrack{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &Database{orm: orm}, nil
}

// ListenedTrack represents a track that has been listened to.
type ListenedTrack struct {
	gorm.Model `json:"-"`

	TrackName string    `json:"track_name"`
	At        time.Time `json:"at"`
	During    int64     `json:"during"`
}

// AddListenedTrack adds a listened track to the database.
func (db *Database) AddListenedTrack(track *audio.Track, when time.Time, during int64) error {
	return db.orm.Create(&ListenedTrack{
		TrackName: track.Name,
		At:        when,
		During:    during,
	}).Error
}

// ListenedTracks gets all listened tracks since the given time.
func (db *Database) ListenedTracks(since time.Time) ([]*ListenedTrack, error) {
	var tracks []*ListenedTrack
	query := db.orm.Model(&ListenedTrack{}).
		Where("at > ?", since).
		Order("at DESC").
		Find(&tracks)
	if err := query.Error; err != nil {
		return nil, fmt.Errorf("failed to get listened tracks: %w", err)
	}
	return tracks, nil
}

// MostListenedTrack represents a track that has been listened to.
type MostListenedTrack struct {
	TrackName string `json:"track_name"`
	Since     string `json:"since"`
	During    int64  `json:"during"`
	Count     int    `json:"count"`
}

// MostListenedTracks gets the most listened tracks since the given time.
func (db *Database) MostListenedTracks(since time.Time, topNb int) ([]*MostListenedTrack, error) {
	var mostListenedTracks []*MostListenedTrack
	query := db.orm.Model(&ListenedTrack{}).
		Select("track_name, min(at) as since, sum(during) as during, count(track_name) as count").
		Where("at > ?", since).
		Group("track_name").
		Order("during DESC").
		Order("since DESC").
		Limit(topNb).
		Find(&mostListenedTracks)
	if err := query.Error; err != nil {
		return nil, fmt.Errorf("failed to get most listened tracks: %w", err)
	}

	return mostListenedTracks, nil
}
