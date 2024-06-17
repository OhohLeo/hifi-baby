package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"

	"github.com/OhohLeo/hifi-baby/audio"
)

type Server struct {
	audio     *audio.Audio
	router    *chi.Mux
	serverURL string // Use ServerURL to start the server
}

// NewServer creates a new Server instance with routes configured for audio management.
func NewServer(audio *audio.Audio, serverURL string) *Server {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	server := &Server{
		audio:     audio,
		router:    r,
		serverURL: serverURL,
	}

	r.Route("/audio", func(r chi.Router) {
		r.Put("/", server.addTrack)                    // Add a track
		r.Delete("/{trackIndex}", server.removeTrack)  // Remove a track
		r.Post("/play/{trackIndex}", server.playTrack) // Play a track
		r.Post("/pause", server.pauseTrack)            // Pause the current track
		r.Post("/resume", server.resumeTrack)          // Resume the current track
		r.Post("/stop", server.stopTrack)              // Stop the current track
		r.Get("/tracks", server.listTracks)            // List all tracks
		r.Get("/current", server.currentTrack)         // Get the current track
		r.Post("/volume/up", server.increaseVolume)    // Increase volume
		r.Post("/volume/down", server.decreaseVolume)  // Decrease volume
	})

	return server
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
	trackIndex, err := strconv.Atoi(chi.URLParam(r, "trackIndex"))
	if err != nil {
		http.Error(w, "Invalid track index", http.StatusBadRequest)
		return
	}
	err = s.audio.RemoveTrack(trackIndex)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) playTrack(w http.ResponseWriter, r *http.Request) {
	trackIndex, err := strconv.Atoi(chi.URLParam(r, "trackIndex"))
	if err != nil {
		http.Error(w, "Invalid track index", http.StatusBadRequest)
		return
	}

	s.audio.PlayTrack(trackIndex)
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

func (s *Server) currentTrack(w http.ResponseWriter, r *http.Request) {
	currentTrack := s.audio.GetCurrentTrack()
	if currentTrack != nil {
		json.NewEncoder(w).Encode(currentTrack)
	} else {
		http.Error(w, "No track is currently playing", http.StatusNotFound)
	}
}

func (s *Server) increaseVolume(w http.ResponseWriter, r *http.Request) {
	s.audio.IncreaseVolume()
	w.WriteHeader(http.StatusOK)
}

func (s *Server) decreaseVolume(w http.ResponseWriter, r *http.Request) {
	s.audio.DecreaseVolume()
	w.WriteHeader(http.StatusOK)
}
