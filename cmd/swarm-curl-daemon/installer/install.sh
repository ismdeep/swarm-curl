#!/usr/bin/env bash

set -e

# Get to workdir
cd "$(realpath "$(dirname "$(realpath "${BASH_SOURCE[0]}")")")"

echo "Installing swarm-curl-daemon..."

# Check root permission
if [ "$EUID" -ne 0 ]; then
  echo "[ERROR] please run as root." && false
fi

# Initialize addr to 0.0.0.0:38080 if not exists in both /etc/swarm-curl-daemon/addr and ${HOME}/.swarm-curl-daemon/addr
if [ ! -f /etc/swarm-curl-daemon/addr ] && [ ! -f "${HOME}/.swarm-curl-daemon/addr" ]; then
    mkdir -p /etc/swarm-curl-daemon
    echo "0.0.0.0:38080" > /etc/swarm-curl-daemon/addr
fi

# Initialize token with random uuid if not exists in both /etc/swarm-curl-daemon/token and ${HOME}/.swarm-curl-daemon/token
if [ ! -f /etc/swarm-curl-daemon/token ] && [ ! -f "${HOME}/.swarm-curl-daemon/token" ]; then
    mkdir -p /etc/swarm-curl-daemon
    uuidgen > /etc/swarm-curl-daemon/token
fi

# Stop swarm-curl-daemon service
systemctl disable --now swarm-curl-daemon.service || true

# Copy binary based on architecture
case $(uname -m) in
x86_64|amd64)
  cp ./binary/amd64/swarm-curl-daemon /usr/local/bin/swarm-curl-daemon
  ;;
aarch64|arm64)
  cp ./binary/arm64/swarm-curl-daemon /usr/local/bin/swarm-curl-daemon
  ;;
*)
  echo "[ERROR] unsupported platform: $(uname -m)" && false
  ;;
esac

# Copy swarm-curl-daemon.service file
cp swarm-curl-daemon.service /etc/systemd/system/swarm-curl-daemon.service

# Start swarm-curl-daemon.service
systemctl daemon-reload
systemctl enable --now swarm-curl-daemon.service

echo "Install completed."
