package bluetooth

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/OhohLeo/hifi-baby/audio"
	"github.com/OhohLeo/hifi-baby/settings"
	"github.com/OhohLeo/hifi-baby/sql"
)

// Server represents the BLE GATT server
type Server struct {
	config         Config
	adapter        *Adapter
	handler        *CharacteristicHandler
	serviceManager *ServiceManager
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewServer creates a new BLE GATT server
func NewServer(
	audio *audio.Audio,
	settings *settings.Settings,
	database *sql.Database,
	config Config,
) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize adapter
	adapter, err := NewAdapter(ctx, config)
	if err != nil {
		cancel()
		return nil, NewBLEError("SERVER_INIT_FAILED", "Failed to create adapter", err)
	}

	// Initialize characteristic handler
	handler := NewCharacteristicHandler(ctx, audio, settings, database, config)

	// Initialize service manager
	serviceManager := NewServiceManager(adapter.GetAdapter(), handler)

	server := &Server{
		config:         config,
		adapter:        adapter,
		handler:        handler,
		serviceManager: serviceManager,
		ctx:            ctx,
		cancel:         cancel,
	}

	return server, nil
}

// Run starts the BLE GATT server
func (s *Server) Run() error {
	if !s.config.Enabled {
		log.Info().Msg("Bluetooth server disabled in configuration")
		return nil
	}

	log.Info().Msg("Starting Bluetooth GATT server")

	// Initialize adapter
	if err := s.adapter.Initialize(); err != nil {
		return NewBLEError("ADAPTER_INIT_FAILED", "Failed to initialize adapter", err)
	}

	// Register services
	serviceUUIDs, err := s.serviceManager.RegisterServices()
	if err != nil {
		return NewBLEError("SERVICE_REGISTRATION_FAILED", "Failed to register services", err)
	}

	log.Info().Int("services", len(serviceUUIDs)).Msg("GATT services registered")

	// Start player state notifications
	s.startPlayerStateNotifications()

	// Start advertising
	if err := s.adapter.StartAdvertising(serviceUUIDs); err != nil {
		return NewBLEError("ADVERTISING_FAILED", "Failed to start advertising", err)
	}

	log.Info().
		Str("device_name", s.config.DeviceName).
		Bool("discoverable", s.config.Discoverable).
		Msg("Bluetooth GATT server running")

	// Keep the server running
	<-s.ctx.Done()

	return nil
}

// startPlayerStateNotifications starts periodic player state notifications
func (s *Server) startPlayerStateNotifications() {
	// Start a goroutine to periodically update readable characteristics
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				s.serviceManager.UpdateReadableCharacteristics()
			}
		}
	}()

	log.Info().Msg("Characteristic value updates enabled (2s interval)")
}

// Close shuts down the BLE GATT server
func (s *Server) Close() error {
	log.Info().Msg("Shutting down Bluetooth GATT server")

	s.cancel()

	if s.handler != nil {
		s.handler.Close()
	}

	if s.adapter != nil {
		s.adapter.Close()
	}

	return nil
}
