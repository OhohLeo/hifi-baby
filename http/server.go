package http

import (
	"encoding/json"
	"net/http" // Ensure os is imported
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/OhohLeo/hifi-baby/audio"
	"github.com/OhohLeo/hifi-baby/settings"
	"github.com/OhohLeo/hifi-baby/sql"
)

type Config struct {
	ServerURL    string `env:"SERVER_URL,default=localhost:3000"`
	ServerUIPath string `env:"SERVER_UI_PATH,default=dist"`
}

type Server struct {
	audio     *audio.Audio
	router    *chi.Mux
	serverURL string // Use ServerURL to start the server
	config    Config
	settings  *settings.Settings
	database  *sql.Database
}

// NewServer creates a new Server instance with routes configured for audio management.
func NewServer(
	audio *audio.Audio,
	config Config,
	settings *settings.Settings,
	database *sql.Database,
) *Server {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},                            // Accept requests from all origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"}, // Specify allowed methods
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value of the Access-Control-Max-Age header.
	}))

	// Serve static files from 'dist' directory
	FileServer(r, "/", config.ServerUIPath)

	server := &Server{
		audio:     audio,
		router:    r,
		serverURL: config.ServerURL,
		config:    config,
		settings:  settings,
		database:  database,
	}

	r.Route("/audio", func(r chi.Router) {
		r.Post("/", server.addTrack)                              // Add a track
		r.Delete("/{trackID}", server.removeTrack)                // Remove a track
		r.Post("/play/{trackID}", server.playTrack)               // Play a track
		r.Post("/pause", server.pauseTrack)                       // Pause the current track
		r.Post("/resume", server.resumeTrack)                     // Resume the current track
		r.Post("/stop", server.stopTrack)                         // Stop the current track
		r.Get("/tracks", server.listTracks)                       // List all tracks
		r.Get("/tracks/listened", server.listenedTracks)          // List all listened tracks
		r.Get("/tracks/most-listened", server.mostListenedTracks) // Get the most listened tracks
		r.Get("/state", server.currentPlayerState)                // Get the current player state
		r.Post("/volume/up", server.increaseVolume)               // Increase volume
		r.Post("/volume/down", server.decreaseVolume)             // Decrease volume
		r.Post("/volume/mute", server.muteVolume)                 // Mute volume
	})

	r.Get("/settings", server.getSettings)
	r.Put("/settings", server.updateSettings)

	return server
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root string) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.FileServer(http.Dir(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = chi.URLParam(r, "*")
		fs.ServeHTTP(w, r)
	})
}

// Example of a method to start the HTTP server using the Router in Server
func (s *Server) Run() error {
	log.Info().Msgf("Starting server on %s", s.serverURL)
	return http.ListenAndServe(s.serverURL, s.router) // Use ServerURL to start the server
}

func (s *Server) addTrack(w http.ResponseWriter, r *http.Request) {
	// Handle file upload
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file upload: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Add the track to the audio manager
	track, err := s.audio.AddTrack(file, header)
	if err != nil {
		http.Error(w, "Failed to add the track", http.StatusInternalServerError)
		return
	}
	// Respond to the client
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(track)
}

func (s *Server) removeTrack(w http.ResponseWriter, r *http.Request) {
	trackID, err := uuid.Parse(chi.URLParam(r, "trackID"))
	if err != nil {
		http.Error(w, "Invalid track id", http.StatusBadRequest)
		return
	}
	err = s.audio.RemoveTrack(trackID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) playTrack(w http.ResponseWriter, r *http.Request) {
	trackID, err := uuid.Parse(chi.URLParam(r, "trackID"))
	if err != nil {
		http.Error(w, "Invalid track id", http.StatusBadRequest)
		return
	}

	s.audio.PlayTrack(trackID)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) pauseTrack(w http.ResponseWriter, r *http.Request) {
	s.audio.Pause()
	w.WriteHeader(http.StatusOK)
}

func (s *Server) resumeTrack(w http.ResponseWriter, r *http.Request) {
	s.audio.Resume()
	w.WriteHeader(http.StatusOK)
}

func (s *Server) stopTrack(w http.ResponseWriter, r *http.Request) {
	s.audio.Stop()
	w.WriteHeader(http.StatusOK)
}

func (s *Server) listTracks(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(s.audio.Tracks())
}

func (s *Server) listenedTracks(w http.ResponseWriter, r *http.Request) {
	since := r.URL.Query().Get("since")
	sinceTime, err := time.Parse(time.RFC3339, since)
	if err != nil {
		http.Error(w, "Invalid since time", http.StatusBadRequest)
		return
	}

	listenedTracks, err := s.database.ListenedTracks(sinceTime)
	if err != nil {
		http.Error(w, "Failed to get listened tracks", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(listenedTracks)
}

func (s *Server) mostListenedTracks(w http.ResponseWriter, r *http.Request) {
	since := r.URL.Query().Get("since")
	sinceTime, err := time.Parse(time.RFC3339, since)
	if err != nil {
		http.Error(w, "Invalid since time", http.StatusBadRequest)
		return
	}

	topNb := r.URL.Query().Get("topNb")
	topNbInt, err := strconv.Atoi(topNb)
	if err != nil {
		http.Error(w, "Invalid topNb", http.StatusBadRequest)
		return
	}

	mostListenedTracks, err := s.database.MostListenedTracks(sinceTime, topNbInt)
	if err != nil {
		http.Error(w, "Failed to get listened tracks", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(mostListenedTracks)
}

func (s *Server) currentPlayerState(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(s.audio.GetPlayerState())
}

func (s *Server) increaseVolume(w http.ResponseWriter, r *http.Request) {
	s.audio.IncreaseVolume()
	w.WriteHeader(http.StatusOK)
}

func (s *Server) decreaseVolume(w http.ResponseWriter, r *http.Request) {
	s.audio.DecreaseVolume()
	w.WriteHeader(http.StatusOK)
}

func (s *Server) muteVolume(w http.ResponseWriter, r *http.Request) {
	enableParam := r.URL.Query().Get("enable")
	if enableParam != "true" && enableParam != "false" {
		http.Error(w, "Query parameter 'enable' must be 'true' or 'false'", http.StatusBadRequest)
		return
	}
	s.audio.Mute(enableParam == "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) getSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Encode the settings to the response
	if err := json.NewEncoder(w).Encode(s.settings); err != nil {
		http.Error(w, "Failed to encode settings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) updateSettings(w http.ResponseWriter, r *http.Request) {
	var newSettings settings.Settings

	if err := json.NewDecoder(r.Body).Decode(&newSettings); err != nil {
		http.Error(w, "Failed to decode settings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Encode the updated configuration back to JSON
	if err := s.settings.Update(&newSettings); err != nil {
		http.Error(w, "Failed to update settings: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
