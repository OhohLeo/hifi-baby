package bluetooth

// Config holds Bluetooth Low Energy server configuration
type Config struct {
	Enabled      bool   `env:"BLE_ENABLED,default=true"`
	DeviceName   string `env:"BLE_DEVICE_NAME,default=HiFi Baby"`
	Discoverable bool   `env:"BLE_DISCOVERABLE,default=true"`
	MTU          uint16 `env:"BLE_MTU,default=512"`        // Maximum transmission unit
	ChunkSize    int    `env:"BLE_CHUNK_SIZE,default=480"` // File transfer chunk size (MTU - overhead)
	StoragePath  string `env:"BLE_STORAGE_PATH,default=tracks"`
}

// DefaultConfig returns the default Bluetooth configuration
func DefaultConfig() Config {
	return Config{
		Enabled:      true,
		DeviceName:   "HiFi Baby",
		Discoverable: true,
		MTU:          512,
		ChunkSize:    480,
		StoragePath:  "tracks",
	}
}
