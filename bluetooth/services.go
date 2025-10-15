package bluetooth

import (
	"github.com/rs/zerolog/log"
	"tinygo.org/x/bluetooth"
)

// Service and Characteristic UUID Constants
const (
	// Service UUIDs
	AudioControlServiceUUID    = "6ba1b900-0001-1000-8000-00805f9b34fb"
	TrackManagementServiceUUID = "6ba1b900-0002-1000-8000-00805f9b34fb"
	StatisticsServiceUUID      = "6ba1b900-0003-1000-8000-00805f9b34fb"
	SettingsServiceUUID        = "6ba1b900-0004-1000-8000-00805f9b34fb"

	// Audio Control Service Characteristics
	TrackControlCharUUID  = "6ba1b901-0001-1000-8000-00805f9b34fb"
	TrackPositionCharUUID = "6ba1b902-0001-1000-8000-00805f9b34fb"
	VolumeControlCharUUID = "6ba1b903-0001-1000-8000-00805f9b34fb"
	PlayerStateCharUUID   = "6ba1b904-0001-1000-8000-00805f9b34fb"
	TrackListCharUUID     = "6ba1b905-0001-1000-8000-00805f9b34fb"

	// Track Management Service Characteristics
	TrackUploadControlCharUUID = "6ba1b906-0002-1000-8000-00805f9b34fb"
	TrackUploadDataCharUUID    = "6ba1b907-0002-1000-8000-00805f9b34fb"
	TrackUploadStatusCharUUID  = "6ba1b908-0002-1000-8000-00805f9b34fb"
	TrackDeleteCharUUID        = "6ba1b909-0002-1000-8000-00805f9b34fb"

	// Statistics Service Characteristics
	ListenedTracksCharUUID = "6ba1b90a-0003-1000-8000-00805f9b34fb"
	MostListenedCharUUID   = "6ba1b90b-0003-1000-8000-00805f9b34fb"

	// Settings Service Characteristics
	SettingsReadCharUUID  = "6ba1b90c-0004-1000-8000-00805f9b34fb"
	SettingsWriteCharUUID = "6ba1b90d-0004-1000-8000-00805f9b34fb"
)

// ServiceManager manages GATT service registration and lifecycle
type ServiceManager struct {
	adapter         *bluetooth.Adapter
	handler         *CharacteristicHandler
	characteristics map[string]*bluetooth.Characteristic
}

// NewServiceManager creates a new service manager
func NewServiceManager(adapter *bluetooth.Adapter, handler *CharacteristicHandler) *ServiceManager {
	return &ServiceManager{
		adapter:         adapter,
		handler:         handler,
		characteristics: make(map[string]*bluetooth.Characteristic),
	}
}

// RegisterServices registers all GATT services and characteristics
func (sm *ServiceManager) RegisterServices() ([]bluetooth.UUID, error) {
	var serviceUUIDs []bluetooth.UUID

	// Register Audio Control Service
	audioServiceUUID, err := sm.registerAudioControlService()
	if err != nil {
		return nil, err
	}
	serviceUUIDs = append(serviceUUIDs, audioServiceUUID)

	// Register Track Management Service
	trackServiceUUID, err := sm.registerTrackManagementService()
	if err != nil {
		return nil, err
	}
	serviceUUIDs = append(serviceUUIDs, trackServiceUUID)

	// Register Statistics Service
	statsServiceUUID, err := sm.registerStatisticsService()
	if err != nil {
		return nil, err
	}
	serviceUUIDs = append(serviceUUIDs, statsServiceUUID)

	// Register Settings Service
	settingsServiceUUID, err := sm.registerSettingsService()
	if err != nil {
		return nil, err
	}
	serviceUUIDs = append(serviceUUIDs, settingsServiceUUID)

	log.Info().Int("services", len(serviceUUIDs)).Msg("All GATT services registered")
	return serviceUUIDs, nil
}

func (sm *ServiceManager) registerAudioControlService() (bluetooth.UUID, error) {
	serviceUUID, err := bluetooth.ParseUUID(AudioControlServiceUUID)
	if err != nil {
		return bluetooth.UUID{}, err
	}

	// Track Control Characteristic (Write)
	trackControlUUID, _ := bluetooth.ParseUUID(TrackControlCharUUID)
	var trackControlChar bluetooth.Characteristic

	// Track Position Characteristic (Read, Write)
	trackPositionUUID, _ := bluetooth.ParseUUID(TrackPositionCharUUID)
	var trackPositionChar bluetooth.Characteristic

	// Volume Control Characteristic (Write)
	volumeControlUUID, _ := bluetooth.ParseUUID(VolumeControlCharUUID)
	var volumeControlChar bluetooth.Characteristic

	// Player State Characteristic (Read, Notify)
	playerStateUUID, _ := bluetooth.ParseUUID(PlayerStateCharUUID)
	var playerStateChar bluetooth.Characteristic

	// Track List Characteristic (Read)
	trackListUUID, _ := bluetooth.ParseUUID(TrackListCharUUID)
	var trackListChar bluetooth.Characteristic

	// Add service with characteristics
	err = sm.adapter.AddService(&bluetooth.Service{
		UUID: serviceUUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				Handle: &trackControlChar,
				UUID:   trackControlUUID,
				Flags:  bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					sm.handler.HandleTrackControlWrite(client, offset, value)
				},
			},
			{
				Handle: &trackPositionChar,
				UUID:   trackPositionUUID,
				Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					sm.handler.HandleTrackPositionWrite(client, offset, value)
					// Update the value for future reads
					data, _ := sm.handler.HandleTrackPositionRead(client, offset)
					trackPositionChar.Write(data)
				},
			},
			{
				Handle: &volumeControlChar,
				UUID:   volumeControlUUID,
				Flags:  bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					sm.handler.HandleVolumeControlWrite(client, offset, value)
				},
			},
			{
				Handle: &playerStateChar,
				UUID:   playerStateUUID,
				Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicNotifyPermission,
			},
			{
				Handle: &trackListChar,
				UUID:   trackListUUID,
				Flags:  bluetooth.CharacteristicReadPermission,
			},
		},
	})

	if err != nil {
		return bluetooth.UUID{}, err
	}

	// Store characteristics for later access
	sm.characteristics[PlayerStateCharUUID] = &playerStateChar
	sm.characteristics[TrackListCharUUID] = &trackListChar
	sm.characteristics[TrackPositionCharUUID] = &trackPositionChar

	log.Info().Msg("Audio Control Service registered")
	return serviceUUID, nil
}

func (sm *ServiceManager) registerTrackManagementService() (bluetooth.UUID, error) {
	serviceUUID, err := bluetooth.ParseUUID(TrackManagementServiceUUID)
	if err != nil {
		return bluetooth.UUID{}, err
	}

	// Track Upload Control Characteristic (Write)
	uploadControlUUID, _ := bluetooth.ParseUUID(TrackUploadControlCharUUID)
	var uploadControlChar bluetooth.Characteristic

	// Track Upload Data Characteristic (Write)
	uploadDataUUID, _ := bluetooth.ParseUUID(TrackUploadDataCharUUID)
	var uploadDataChar bluetooth.Characteristic

	// Track Upload Status Characteristic (Read, Notify)
	uploadStatusUUID, _ := bluetooth.ParseUUID(TrackUploadStatusCharUUID)
	var uploadStatusChar bluetooth.Characteristic

	// Track Delete Characteristic (Write)
	deleteUUID, _ := bluetooth.ParseUUID(TrackDeleteCharUUID)
	var deleteChar bluetooth.Characteristic

	// Add service with characteristics
	err = sm.adapter.AddService(&bluetooth.Service{
		UUID: serviceUUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				Handle: &uploadControlChar,
				UUID:   uploadControlUUID,
				Flags:  bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					sm.handler.HandleUploadControlWrite(client, offset, value)
				},
			},
			{
				Handle: &uploadDataChar,
				UUID:   uploadDataUUID,
				Flags:  bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					sm.handler.HandleUploadDataWrite(client, offset, value)
				},
			},
			{
				Handle: &uploadStatusChar,
				UUID:   uploadStatusUUID,
				Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicNotifyPermission,
			},
			{
				Handle: &deleteChar,
				UUID:   deleteUUID,
				Flags:  bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					sm.handler.HandleTrackDeleteWrite(client, offset, value)
				},
			},
		},
	})

	if err != nil {
		return bluetooth.UUID{}, err
	}

	sm.characteristics[TrackUploadStatusCharUUID] = &uploadStatusChar

	log.Info().Msg("Track Management Service registered")
	return serviceUUID, nil
}

func (sm *ServiceManager) registerStatisticsService() (bluetooth.UUID, error) {
	serviceUUID, err := bluetooth.ParseUUID(StatisticsServiceUUID)
	if err != nil {
		return bluetooth.UUID{}, err
	}

	// Listened Tracks Characteristic (Read, Write)
	listenedTracksUUID, _ := bluetooth.ParseUUID(ListenedTracksCharUUID)
	var listenedTracksChar bluetooth.Characteristic

	// Most Listened Tracks Characteristic (Read, Write)
	mostListenedUUID, _ := bluetooth.ParseUUID(MostListenedCharUUID)
	var mostListenedChar bluetooth.Characteristic

	// Add service with characteristics
	err = sm.adapter.AddService(&bluetooth.Service{
		UUID: serviceUUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				Handle: &listenedTracksChar,
				UUID:   listenedTracksUUID,
				Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					sm.handler.HandleListenedTracksWrite(client, offset, value)
					// Update the value for future reads
					data, _ := sm.handler.HandleListenedTracksRead(client, offset)
					listenedTracksChar.Write(data)
				},
			},
			{
				Handle: &mostListenedChar,
				UUID:   mostListenedUUID,
				Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					sm.handler.HandleMostListenedWrite(client, offset, value)
					// Update the value for future reads
					data, _ := sm.handler.HandleMostListenedRead(client, offset)
					mostListenedChar.Write(data)
				},
			},
		},
	})

	if err != nil {
		return bluetooth.UUID{}, err
	}

	sm.characteristics[ListenedTracksCharUUID] = &listenedTracksChar
	sm.characteristics[MostListenedCharUUID] = &mostListenedChar

	log.Info().Msg("Statistics Service registered")
	return serviceUUID, nil
}

func (sm *ServiceManager) registerSettingsService() (bluetooth.UUID, error) {
	serviceUUID, err := bluetooth.ParseUUID(SettingsServiceUUID)
	if err != nil {
		return bluetooth.UUID{}, err
	}

	// Settings Read Characteristic (Read)
	settingsReadUUID, _ := bluetooth.ParseUUID(SettingsReadCharUUID)
	var settingsReadChar bluetooth.Characteristic

	// Settings Write Characteristic (Write)
	settingsWriteUUID, _ := bluetooth.ParseUUID(SettingsWriteCharUUID)
	var settingsWriteChar bluetooth.Characteristic

	// Add service with characteristics
	err = sm.adapter.AddService(&bluetooth.Service{
		UUID: serviceUUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				Handle: &settingsReadChar,
				UUID:   settingsReadUUID,
				Flags:  bluetooth.CharacteristicReadPermission,
			},
			{
				Handle: &settingsWriteChar,
				UUID:   settingsWriteUUID,
				Flags:  bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					sm.handler.HandleSettingsWrite(client, offset, value)
					// Update settings read characteristic
					data, _ := sm.handler.HandleSettingsRead(client, offset)
					settingsReadChar.Write(data)
				},
			},
		},
	})

	if err != nil {
		return bluetooth.UUID{}, err
	}

	sm.characteristics[SettingsReadCharUUID] = &settingsReadChar

	log.Info().Msg("Settings Service registered")
	return serviceUUID, nil
}

// GetCharacteristic returns a registered characteristic by UUID
func (sm *ServiceManager) GetCharacteristic(uuid string) *bluetooth.Characteristic {
	return sm.characteristics[uuid]
}

// UpdateReadableCharacteristics updates all readable characteristics with current values
func (sm *ServiceManager) UpdateReadableCharacteristics() {
	// This method can be called periodically to update readable characteristic values
	// Use Connection(0) as dummy connection for internal updates

	// Update player state
	if char := sm.characteristics[PlayerStateCharUUID]; char != nil {
		data, _ := sm.handler.HandlePlayerStateRead(bluetooth.Connection(0), 0)
		char.Write(data)
	}

	// Update track list
	if char := sm.characteristics[TrackListCharUUID]; char != nil {
		data, _ := sm.handler.HandleTrackListRead(bluetooth.Connection(0), 0)
		char.Write(data)
	}

	// Update settings
	if char := sm.characteristics[SettingsReadCharUUID]; char != nil {
		data, _ := sm.handler.HandleSettingsRead(bluetooth.Connection(0), 0)
		char.Write(data)
	}
}
