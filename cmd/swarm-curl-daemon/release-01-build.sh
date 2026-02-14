#!/usr/bin/env bash

set -e

# Get to workdir
cd "$(realpath "$(dirname "$(realpath "${BASH_SOURCE[0]}")")")"

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./installer/binary/amd64/swarm-curl-daemon -mod vendor -trimpath -ldflags '-s -w' .
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o ./installer/binary/arm64/swarm-curl-daemon -mod vendor -trimpath -ldflags '-s -w' .
