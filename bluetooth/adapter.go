package bluetooth

import (
	"context"

	"github.com/rs/zerolog/log"
	"tinygo.org/x/bluetooth"
)

// Adapter manages the Bluetooth adapter for BLE operations
type Adapter struct {
	adapter       *bluetooth.Adapter
	config        Config
	ctx           context.Context
	cancel        context.CancelFunc
	advertisement *bluetooth.Advertisement // Keep advertisement reference to prevent GC
}

// NewAdapter creates a new Bluetooth adapter instance
func NewAdapter(ctx context.Context, config Config) (*Adapter, error) {
	adapter := bluetooth.DefaultAdapter

	ctx, cancel := context.WithCancel(ctx)

	a := &Adapter{
		adapter: adapter,
		config:  config,
		ctx:     ctx,
		cancel:  cancel,
	}

	return a, nil
}

// Initialize powers on the adapter and configures it
func (a *Adapter) Initialize() error {
	log.Info().Msg("Initializing Bluetooth adapter")

	// Enable the adapter
	if err := a.adapter.Enable(); err != nil {
		return NewBLEError("ADAPTER_ENABLE_FAILED", "Failed to enable Bluetooth adapter", err)
	}

	log.Info().
		Str("device_name", a.config.DeviceName).
		Bool("discoverable", a.config.Discoverable).
		Uint16("mtu", a.config.MTU).
		Msg("Bluetooth adapter initialized successfully")

	return nil
}

// GetAdapter returns the underlying Bluetooth adapter
func (a *Adapter) GetAdapter() *bluetooth.Adapter {
	return a.adapter
}

// StartAdvertising begins BLE advertising with the device name
func (a *Adapter) StartAdvertising(services []bluetooth.UUID) error {
	log.Info().Msg("Starting BLE advertising")

	adv := a.adapter.DefaultAdvertisement()

	// Store the advertisement reference to prevent garbage collection
	a.advertisement = adv

	// Absolute minimal configuration - only device name, no services
	// This tests if the issue is with service UUID advertising
	advOptions := bluetooth.AdvertisementOptions{
		LocalName: a.config.DeviceName,
	}

	// Configure advertising
	if err := adv.Configure(advOptions); err != nil {
		log.Error().Err(err).Msg("Failed to configure advertising")
		return NewBLEError("ADVERTISING_CONFIG_FAILED", "Failed to configure advertising", err)
	}

	log.Info().
		Str("device_name", a.config.DeviceName).
		Int("services", len(services)).
		Msg("Advertising configured, attempting to start")

	// Start advertising
	log.Info().Msg("Calling adv.Start()...")
	if err := adv.Start(); err != nil {
		log.Error().Err(err).
			Msg("Failed to start advertising. Common causes: " +
				"1) Another BLE app is running " +
				"2) BlueZ needs cleanup (sudo systemctl restart bluetooth) " +
				"3) Adapter not discoverable (bluetoothctl: discoverable on) " +
				"4) Missing permissions (use sudo setcap or run as root)")
		return NewBLEError("ADVERTISING_START_FAILED", "Failed to start advertising", err)
	}

	log.Info().
		Str("device_name", a.config.DeviceName).
		Int("services", len(services)).
		Msg("BLE advertising started successfully - adv.Start() returned nil")

	// Try to get adapter address to verify it's working
	addr, err := a.adapter.Address()
	if err != nil {
		log.Warn().Err(err).Msg("Could not get adapter address")
	} else {
		log.Info().
			Str("address", addr.String()).
			Msg("Bluetooth adapter address")
	}

	return nil
}

// Close shuts down the adapter
func (a *Adapter) Close() error {
	log.Info().Msg("Closing Bluetooth adapter")
	a.cancel()
	return nil
}

// Context returns the adapter's context
func (a *Adapter) Context() context.Context {
	return a.ctx
}
