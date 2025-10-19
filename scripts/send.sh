#!/bin/bash

set -e

# --- Argument Parsing ---
if [ -z "$1" ]; then
    echo "Usage: $0 <target> [action]"
    echo "Supported targets: rpi, rpi0"
    echo "Supported actions: build, deploy, configure"
    exit 1
fi

TARGET_PLATFORM=$1
ACTION=${2:-build} # Default to 'build' if no action is specified

TARGET_USER="hifi-baby"
TARGET_DIR="/home/hifi-baby"

# --- Function Definitions ---

build_binary() {
    echo "--- Building for $TARGET_PLATFORM... ---"
    ./scripts/build.sh "$TARGET_PLATFORM"
}

deploy_files() {
    echo "--- Transferring files to $TARGET_USER@$TARGET_HOST... ---"
    FILES_TO_TRANSFER="hifi-baby settings.json deployment/hifi-baby.service deployment/hifi-baby.default"
    if [ -d "ui/dist" ]; then
        FILES_TO_TRANSFER="$FILES_TO_TRANSFER ui/dist"
    else
        echo "Warning: ui/dist directory not found, skipping."
    fi
    scp -r $FILES_TO_TRANSFER "$TARGET_USER@$TARGET_HOST:$TARGET_DIR/"
}

configure_service() {
    echo "--- Installing and starting the service on $TARGET_HOST... ---"
    if [ "$TARGET_PLATFORM" == "rpi0" ]; then
        echo "Configuring Bluetooth for rpi0..."
        ssh "$TARGET_USER@$TARGET_HOST" << EOF
            set -e
            sudo hciconfig hci0 up
            sudo bluetoothctl power on
            sudo bluetoothctl discoverable on
            sudo bluetoothctl pairable on
            sudo setcap 'cap_net_raw,cap_net_admin+eip' "$TARGET_DIR/hifi-baby"
            
            sudo cp "$TARGET_DIR/hifi-baby.service" /etc/systemd/system/
            sudo cp "$TARGET_DIR/hifi-baby.default" /etc/default/hifi-baby
            sudo systemctl daemon-reload
            sudo systemctl enable hifi-baby.service
            sudo systemctl restart hifi-baby.service
EOF
    else
        ssh "$TARGET_USER@$TARGET_HOST" << EOF
            set -e
            sudo cp "$TARGET_DIR/hifi-baby.service" /etc/systemd/system/
            sudo cp "$TARGET_DIR/hifi-baby.default" /etc/default/hifi-baby
            sudo systemctl daemon-reload
            sudo systemctl enable hifi-baby.service
            sudo systemctl restart hifi-baby.service
EOF
    fi

    echo "--- Checking service status on $TARGET_HOST... ---"
    ssh "$TARGET_USER@$TARGET_HOST" "sudo systemctl status hifi-baby.service"
}

# --- Target Host Resolution ---
case "$TARGET_PLATFORM" in
    "rpi")
        TARGET_HOST="hifi-baby.local"
        ;;
    "rpi0")
        TARGET_HOST="hifi-baby-zero.local"
        ;;
    *)
        echo "Unknown target: $TARGET_PLATFORM"
        echo "Supported targets: rpi, rpi0"
        exit 1
        ;;
esac

echo "Deploying to $TARGET_PLATFORM at $TARGET_HOST with action: $ACTION"

# --- Execution Flow ---
case "$ACTION" in
    "build")
        build_binary
        deploy_files
        configure_service
        ;;
    "deploy")
        deploy_files
        configure_service
        ;;
    "configure")
        configure_service
        ;;
    *)
        echo "Unknown action: $ACTION"
        echo "Supported actions: build, deploy, configure"
        exit 1
        ;;
esac

echo "Deployment complete."
