#!/bin/bash

set -e

echo "Initializing HiFi Baby for Linux"

# Create user and directories
if ! id "hifi-baby" &>/dev/null; then
    echo "Creating hifi-baby user"
    sudo useradd -r -s /bin/false hifi-baby
fi

echo "Creating directories"
sudo mkdir -p /opt/hifi-baby/tracks
sudo mkdir -p /opt/hifi-baby/dist
sudo chown -R hifi-baby:hifi-baby /opt/hifi-baby

# Install service file
echo "Installing systemd service"
sudo cp deployment/hifi-baby.service /etc/systemd/system/
sudo cp deployment/hifi-baby.default /etc/default/hifi-baby

# Update paths in default file for linux
# Note: This uses /opt/hifi-baby as the base directory
# You may need to adjust this
echo "Configuring environment file"
cat << EOF | sudo tee /etc/default/hifi-baby
SERVER_URL=0.0.0.0:3000
SETTINGS_PATH=/opt/hifi-baby/settings.json
STORAGE_PATH=/opt/hifi-baby/tracks
SERVER_UI_PATH=/opt/hifi-baby/dist
DATABASE_PATH=/opt/hifi-baby/hifi-baby.db
DATABASE_TIMEOUT=10s
EOF


echo "Reloading systemd"
sudo systemctl daemon-reload

echo "Initialization complete."
echo "To finish setup:"
echo "1. Copy the 'hifi-baby' binary to /opt/hifi-baby/"
echo "2. Copy 'settings.json' to /opt/hifi-baby/"
echo "3. Copy the UI 'dist' folder to /opt/hifi-baby/"
echo "4. Run 'sudo systemctl enable hifi-baby.service'"
echo "5. Run 'sudo systemctl start hifi-baby.service'"
