#!/bin/bash

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <user@host>"
    exit 1
fi

TARGET=$1
TARGET_DIR="/home/hifi-baby"

echo "Deploying to $TARGET"

# Transfer files
scp -r hifi-baby settings.json ui/dist deployment/hifi-baby.service deployment/hifi-baby.default "$TARGET:$TARGET_DIR/"

echo "Deployment complete."
echo "On the target machine, you may need to run:"
echo "sudo cp $TARGET_DIR/hifi-baby.service /etc/systemd/system/"
echo "sudo systemctl daemon-reload"
echo "sudo systemctl enable hifi-baby.service"
echo "sudo systemctl start hifi-baby.service"
