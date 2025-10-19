#!/bin/bash
# Start script for HiFi Baby with Bluetooth support
# This script ensures the Bluetooth adapter is properly configured before starting the app

set -e

echo "Configuring Bluetooth adapter..."
bluetoothctl power on
bluetoothctl discoverable on
bluetoothctl pairable on

echo ""
echo "Bluetooth adapter status:"
bluetoothctl show | grep -E "(Name|Alias|Powered|Discoverable|Pairable)" | head -10

echo ""
echo "Starting HiFi Baby..."
cd "$(dirname "$0")/.."

# Check if binary exists
if [ ! -f "./hifi-baby" ]; then
    echo "Binary not found. Building..."
    go build
fi

# Check capabilities
if ! getcap ./hifi-baby | grep -q "cap_net_raw,cap_net_admin"; then
    echo "Setting Bluetooth capabilities..."
    sudo setcap 'cap_net_raw,cap_net_admin+eip' ./hifi-baby
fi

echo ""
echo "Launching HiFi Baby..."
./hifi-baby
