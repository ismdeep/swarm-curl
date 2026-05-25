BINDIR ?= /usr/local/bin
BUILD_DIR ?= build

# `make help`
.PHONY: help
help:
	@cat Makefile | grep '# `' | grep -v '@cat Makefile'

# `make build`
.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -o $(BUILD_DIR)/swarm-curl_linux_amd64         -mod vendor -trimpath -ldflags '-s -w' ./cmd/swarm-curl/
	CGO_ENABLED=0 GOOS=linux  GOARCH=arm64 go build -o $(BUILD_DIR)/swarm-curl_linux_arm64         -mod vendor -trimpath -ldflags '-s -w' ./cmd/swarm-curl/
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/swarm-curl_darwin_amd64        -mod vendor -trimpath -ldflags '-s -w' ./cmd/swarm-curl/
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/swarm-curl_darwin_arm64        -mod vendor -trimpath -ldflags '-s -w' ./cmd/swarm-curl/
	CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -o $(BUILD_DIR)/swarm-curl-daemon_linux_amd64  -mod vendor -trimpath -ldflags '-s -w' ./cmd/swarm-curl-daemon/
	CGO_ENABLED=0 GOOS=linux  GOARCH=arm64 go build -o $(BUILD_DIR)/swarm-curl-daemon_linux_arm64  -mod vendor -trimpath -ldflags '-s -w' ./cmd/swarm-curl-daemon/
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/swarm-curl-daemon_darwin_amd64 -mod vendor -trimpath -ldflags '-s -w' ./cmd/swarm-curl-daemon/
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/swarm-curl-daemon_darwin_arm64 -mod vendor -trimpath -ldflags '-s -w' ./cmd/swarm-curl-daemon/

# `make install`
.PHONY: install
install:
	CGO_ENABLED=0 go build -o $(BINDIR)/swarm-curl        -mod vendor -trimpath -ldflags '-s -w' ./cmd/swarm-curl/
	CGO_ENABLED=0 go build -o $(BINDIR)/swarm-curl-daemon -mod vendor -trimpath -ldflags '-s -w' ./cmd/swarm-curl-daemon/

# `make clean`
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
