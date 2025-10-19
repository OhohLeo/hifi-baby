# HiFi Baby Bluetooth Low Energy (BLE) Server

This package implements a complete BLE GATT server that mirrors all HTTP server functionality, enabling Bluetooth clients to control the HiFi Baby audio player.

## Features

All HTTP API endpoints are available via BLE GATT characteristics:

### Audio Control
- Play/pause/resume/stop tracks
- Set playback position
- Volume control (increase, decrease, mute)
- Get player state (with notifications)
- List all tracks

### Track Management
- Upload audio files (chunked transfer)
- Delete tracks
- Upload progress tracking

### Statistics
- Get listened tracks history
- Get most listened tracks

### Settings
- Read current settings
- Update settings

## Architecture

### Components

- **adapter.go**: Bluetooth adapter initialization and management
- **services.go**: GATT service and characteristic registration
- **characteristics.go**: Business logic handlers for all operations
- **protocol.go**: JSON message structures for BLE communication
- **file_transfer.go**: Chunked file upload management
- **notifications.go**: Real-time notification system
- **server.go**: Main BLE server orchestration
- **errors.go**: BLE-specific error types

### GATT Services

#### 1. Audio Control Service (`6ba1b900-0001-1000-8000-00805f9b34fb`)
- **Track Control** (`6ba1b901-...`): Write - Play/pause/resume/stop commands
- **Track Position** (`6ba1b902-...`): Read/Write - Get/set playback position
- **Volume Control** (`6ba1b903-...`): Write - Volume commands
- **Player State** (`6ba1b904-...`): Read/Notify - Current player state
- **Track List** (`6ba1b905-...`): Read - List of all tracks

#### 2. Track Management Service (`6ba1b900-0002-1000-8000-00805f9b34fb`)
- **Upload Control** (`6ba1b906-...`): Write - Init/finalize/cancel uploads
- **Upload Data** (`6ba1b907-...`): Write - Binary chunk data
- **Upload Status** (`6ba1b908-...`): Read/Notify - Upload progress
- **Track Delete** (`6ba1b909-...`): Write - Delete tracks

#### 3. Statistics Service (`6ba1b900-0003-1000-8000-00805f9b34fb`)
- **Listened Tracks** (`6ba1b90a-...`): Read/Write - Query listened tracks
- **Most Listened** (`6ba1b90b-...`): Read/Write - Query most listened

#### 4. Settings Service (`6ba1b900-0004-1000-8000-00805f9b34fb`)
- **Settings Read** (`6ba1b90c-...`): Read - Get current settings
- **Settings Write** (`6ba1b90d-...`): Write - Update settings

## Configuration

Set these environment variables to configure the BLE server:

```bash
# Enable/disable BLE server
BLE_ENABLED=true

# Bluetooth device name (visible during scanning)
BLE_DEVICE_NAME="HiFi Baby"

# Enable discoverable mode
BLE_DISCOVERABLE=true

# Maximum transmission unit (bytes)
BLE_MTU=512

# File transfer chunk size (bytes)
BLE_CHUNK_SIZE=480

# Storage path for uploaded files
BLE_STORAGE_PATH="tracks"
```

## Protocol

### Command/Response Format

All characteristics use JSON encoding:

**Generic Command:**
```json
{
  "command": "string",
  "params": {...}
}
```

**Generic Response:**
```json
{
  "success": boolean,
  "data": object | null,
  "error": string | null
}
```

### Example Commands

**Play a Track:**
```json
{
  "command": "play",
  "track_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Set Volume (Mute):**
```json
{
  "command": "mute",
  "enable": true
}
```

**Set Playback Position:**
```json
{
  "position": 120.5
}
```

### File Upload Protocol

1. **Initiate Upload** (Write to Upload Control):
```json
{
  "command": "init",
  "filename": "track.mp3",
  "size": 5242880,
  "checksum": "sha256-hash"
}
```

2. **Upload Chunks** (Write to Upload Data):
Binary format:
```
[0-3]   uint32 session_id_hash
[4-7]   uint32 chunk_index
[8-11]  uint32 chunk_size
[12-end] binary data
```

3. **Finalize Upload** (Write to Upload Control):
```json
{
  "command": "finalize",
  "session_id": "uuid"
}
```

## Client Usage

### Python Client

A complete Python client using Bleak is provided in `/scripts/bluetooth_client.py`.

**Install dependencies:**
```bash
cd scripts
pip3 install -r requirements.txt
```

**Examples:**

```bash
# List all tracks
./bluetooth_client.py tracks

# Get player state
./bluetooth_client.py state

# Play a track
./bluetooth_client.py play <track-id>

# Pause/resume/stop playback
./bluetooth_client.py pause
./bluetooth_client.py resume
./bluetooth_client.py stop

# Volume control
./bluetooth_client.py volume-up
./bluetooth_client.py volume-down
./bluetooth_client.py mute true

# Upload an audio file
./bluetooth_client.py upload /path/to/track.mp3

# Delete a track
./bluetooth_client.py delete <track-id>

# Get statistics
./bluetooth_client.py listened "2025-01-01T00:00:00Z"
./bluetooth_client.py most-listened "2025-01-01T00:00:00Z" 10

# Get settings
./bluetooth_client.py settings

# Listen to real-time player state updates
./bluetooth_client.py listen
```

## Integration

The BLE server runs alongside the HTTP server:

```go
import "github.com/OhohLeo/hifi-baby/bluetooth"

// Initialize
bleServer, err := bluetooth.NewServer(audioInstance, settings, database, config.Bluetooth)
if err != nil {
    return err
}

// Run in background
go func() {
    if err := bleServer.Run(); err != nil {
        log.Error().Err(err).Msg("BLE server error")
    }
}()
```

## Technical Details

### Concurrency

- All handlers are goroutine-safe with mutex protection
- File uploads use session-based management
- Characteristic values update every 2 seconds for read operations
- Independent write handlers for each characteristic

### Error Handling

- Structured error types in `errors.go`
- JSON error responses with error codes
- Graceful handling of invalid commands/sessions
- Checksum verification for file uploads

### Notifications

- Player state updates pushed to subscribed clients
- Upload progress notifications during file transfers
- Automatic characteristic value updates

## Platform Requirements

### Linux (Raspberry Pi)

- BlueZ Bluetooth stack (usually pre-installed)
- Bluetooth adapter (built-in or USB)

**Check Bluetooth status:**
```bash
hciconfig
bluetoothctl show
```

### Permissions and Startup

The application needs Bluetooth capabilities and the adapter must be in discoverable mode.

**Recommended: Use the startup script**
```bash
./scripts/start_bluetooth.sh
```

This script automatically:
- Configures the Bluetooth adapter (discoverable, pairable)
- Sets the required capabilities
- Starts the application

**Manual startup:**
```bash
# 1. Configure Bluetooth adapter (REQUIRED before starting app)
bluetoothctl power on
bluetoothctl discoverable on
bluetoothctl pairable on

# 2. Set capabilities
sudo setcap 'cap_net_raw,cap_net_admin+eip' ./hifi-baby

# 3. Run the application
./hifi-baby
```

**Why this is needed:** TinyGo's Bluetooth library on Linux requires the adapter to be in discoverable mode before the application starts advertising. If you start the app before enabling discoverable mode, advertising will fail with "Failed to register advertisement".

## Testing

### Manual Testing

1. **Start the server:**
```bash
./hifi-baby
```

2. **Scan for the device:**
```bash
bluetoothctl
scan on
# Look for "HiFi Baby"
```

3. **Use the Python client:**
```bash
cd scripts
./bluetooth_client.py state
```

### Troubleshooting

**Advertising fails on startup (most common issue):**

This error typically occurs: `bluetooth: could not start advertisement: Failed to register advertisement`

**Root cause:** The Bluetooth adapter is not in discoverable mode when the application tries to start advertising.

**Solution 1 - Use the startup script (recommended):**
```bash
./scripts/start_bluetooth.sh
```

**Solution 2 - Manual fix:**
```bash
# 1. Check current adapter status
bluetoothctl show | grep -E "(Discoverable|Pairable)"

# 2. If Discoverable: no or Pairable: no, enable them:
bluetoothctl power on
bluetoothctl discoverable on
bluetoothctl pairable on

# 3. Verify it's enabled
bluetoothctl show | grep -E "(Discoverable|Pairable)"
# Should show: Discoverable: yes and Pairable: yes

# 4. Restart your application
./hifi-baby
```

**If still failing, clean up BlueZ:**
```bash
# Stop any other BLE applications
sudo killall hifi-baby 2>/dev/null

# Restart BlueZ to clear old advertisements
sudo systemctl restart bluetooth
sleep 2

# Configure adapter
bluetoothctl power on
bluetoothctl discoverable on
bluetoothctl pairable on

# Clean up cached data (if needed)
sudo rm -rf /var/lib/bluetooth/*/cache
sudo systemctl restart bluetooth
sleep 2

# Run with startup script
./scripts/start_bluetooth.sh
```

**Device not found:**
- Check `BLE_ENABLED=true` in configuration
- Verify Bluetooth adapter is powered: `hciconfig hci0 up`
- Check BlueZ is running: `systemctl status bluetooth`
- Ensure device name matches: default is "HiFi Baby", check `BLE_DEVICE_NAME`
- Wait a few seconds after starting the server before scanning

**Connection fails:**
- Ensure device is discoverable
- Check permissions (capabilities or root)
- Verify no other process is using Bluetooth
- Try: `sudo systemctl restart bluetooth`

**Characteristics not readable:**
- Wait 2 seconds after connection (value updates every 2s)
- Try reading after performing a write operation

**Client can't find device:**
- Make sure you're not using quotes in the device name: use `--device-name Hifi` not `--device-name '"Hifi'`
- Check the actual device name in logs or bluetoothctl: `bluetoothctl devices`
- Try scanning without specifying device name to see all devices

## Performance

- **MTU**: 512 bytes (configurable)
- **Chunk Size**: 480 bytes for file transfers
- **Update Interval**: 2 seconds for readable characteristics
- **Concurrent Connections**: Supports multiple BLE clients

## Security Considerations

- **Pairing**: Device is pairable by default
- **Encryption**: BlueZ handles link-layer encryption automatically
- **Access Control**: No authentication layer (local network security)
- **File Validation**: SHA256 checksum verification for uploads

## Future Enhancements

Potential improvements:
- Authentication/authorization layer
- Encrypted characteristics for sensitive data
- Connection-specific notifications (per-client state)
- Batch operations for multiple tracks
- Resume incomplete uploads
- Extended MTU negotiation

## License

Same as parent project.
