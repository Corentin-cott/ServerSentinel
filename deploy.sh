#!/bin/bash
set -e
set -o pipefail

echo "Starting deployment..."

USER="serveursentinel"
WORKDIR="/opt/serveursentinel"
BUILD_CACHE="$WORKDIR/.cache/go-build"
MOD_CACHE="$WORKDIR/.cache/go-mod"
BRANCH_USED="#27/support-docker"

cd "$WORKDIR" || { echo "Error: Cannot access $WORKDIR"; exit 1; }

echo "Fetching latest code from main branch..."
sudo -u $USER git fetch origin
sudo -u $USER git checkout "$BRANCH_USED"
sudo -u $USER git pull origin "$BRANCH_USED"

echo "Ensuring Go cache directories exist..."
sudo -u $USER mkdir -p "$BUILD_CACHE"
sudo -u $USER mkdir -p "$MOD_CACHE"

echo "Building serversentinel daemon with custom caches..."
sudo -u $USER env GOCACHE="$BUILD_CACHE" GOMODCACHE="$MOD_CACHE" go build -o bin/serversentinel ./cmd/daemon

echo "Restarting serversentinel service..."
systemctl restart serveursentinel

echo "Deployment completed successfully!"