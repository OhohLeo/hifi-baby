# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Hifi Baby is a musical children's Hi-Fi system designed for Raspberry Pi. It randomly plays music tracks on demand when a physical button is pressed. The system manages audio playback, GPIO button controls, and provides a REST API for track management.

## Build and Run Commands

```bash
# Build the application
go build

# Run the application
./hifi-baby

# Build and run in one step
go build && ./hifi-baby
```

## Environment Variables

| Module | Variable | Description | Default |
|--------|----------|-------------|---------|
| General | LOG_LEVEL | Log level (debug, info, warn, error) | info |
| General | SETTINGS_PATH | Path to settings configuration file | settings.json |
| Audio | STORAGE_PATH | Path to audio tracks storage directory | tracks |
| Server | SERVER_URL | HTTP server listen address | localhost:3000 |
| Server | SERVER_UI_PATH | Path to UI static files | dist |
| Database | DATABASE_PATH | SQLite database file path | ./hifi-baby.db |
| DATABASE_TIMEOUT | Database operation timeout | 10s |

## Architecture

### Component Initialization Flow

The application follows a specific initialization order in `main.go`:
1. **Config**: Environment variables are loaded via `app.NewConfig()`
2. **Settings**: JSON settings file is loaded via `settings.NewSettings()`
3. **Database**: SQLite database is initialized via `sql.NewDatabase()`
4. **Audio**: Audio system is initialized with config, settings, and database
5. **Server**: HTTP server is created with audio, config, settings, and database
6. **GPIO**: Raspberry Pi GPIO handler is initialized for button control
7. **App.Run()**: Starts three concurrent goroutines (GPIO listener, audio manager, HTTP server)

### Core Components

**app/**: Application orchestration
- `app.go`: Coordinates Audio, HTTP Server, GPIO, and Database components
- `config.go`: Environment variable configuration loading using go-env
- Three concurrent goroutines run: GPIO listener, audio player, HTTP server

**audio/**: Audio playback system
- `audio.go`: Main audio manager using gopxl/beep library for playback
- `track.go`: Track metadata and file operations
- `player.go`: Player state management
- Supports MP3, WAV, FLAC, OGG formats
- Channel-based playback control (`playRequests` and `stopChan` channels)
- Volume control with effects (mute, increase/decrease)
- Implements Capabilities interface to report listened tracks to database

**http/**: REST API server
- `server.go`: Chi router-based HTTP server with CORS enabled
- REST API endpoints under `/audio/*` for track management and playback control
- Static file serving for UI (from `dist/` directory)
- Settings management endpoints at `/settings`
- Postman collection available at `doc/HifiBaby.postman_collection.json`

**raspberry/**: GPIO control
- `gpio.go`: Physical button input handling via warthog618/go-gpiocdev
- Default configuration: gpiochip0, pin 16
- Button press detection: single press = change track, double press within 1 second = stop
- Sends actions through channel to app layer

**sql/**: Database layer
- `database.go`: GORM-based SQLite persistence
- Tracks listening history (track name, timestamp, duration)
- Provides analytics: `ListenedTracks()` and `MostListenedTracks()`

**settings/**: Configuration persistence
- `settings.go`: JSON-based settings file management
- Audio settings: volume controls (base, default, min, max, step), silent mode
- Settings are loaded at startup and can be updated via API

### Key Design Patterns

**Dependency Injection**: Components are wired together through constructors, with Database implementing the `Capabilities` interface required by Audio for tracking listened tracks.

**Channel-based Communication**: GPIO events flow through a string channel to the app layer. Audio playback requests use UUID channels, and stop signals use boolean channels.

**Speaker Library Locking**: The gopxl/beep speaker requires explicit `speaker.Lock()` and `speaker.Unlock()` calls when modifying playback state (pause, volume changes).

## System Requirements

```bash
# Install SQLite3 (required for database)
sudo apt install sqlite3
```

## API Structure

The REST API follows these conventions:
- Base path: `/audio`
- Track operations: POST `/audio/`, DELETE `/audio/{trackID}`, GET `/audio/tracks`
- Playback controls: POST `/audio/play/{trackID}`, `/audio/pause`, `/audio/resume`, `/audio/stop`
- Volume controls: POST `/audio/volume/up`, `/audio/volume/down`, `/audio/volume/mute?enable=true`
- Analytics: GET `/audio/tracks/listened?since={RFC3339}`, `/audio/tracks/most-listened?since={RFC3339}&topNb={N}`
- Player state: GET `/audio/state`
- Settings: GET/PUT `/settings`

## Important Constraints

- The application is designed for single-user, embedded Raspberry Pi usage
- Audio playback is sequential (one track at a time)
- GPIO is hardcoded to gpiochip0, pin 16 in `app.go:41`
- No tests currently exist in the codebase
- When changing code, always check with `go build`