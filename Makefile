.PHONY: help
help:
	@cat Makefile | grep '# `' | grep -v '@cat Makefile'

# `make build`
.PHONY: build
build:
	CGO_ENABLED=0 go build -o ./build/swarm-curl        -mod vendor -trimpath -ldflags '-s -w' ./cmd/swarm-curl/
	CGO_ENABLED=0 go build -o ./build/swarm-curl-daemon -mod vendor -trimpath -ldflags '-s -w' ./cmd/swarm-curl-daemon/

# `make release-swarm-curl-daemon`
.PHONY: release-swarm-curl-daemon
release-swarm-curl-daemon:
	docker buildx build \
		--push \
		--platform linux/amd64,linux/arm64 \
		--tag "$${DOCKER_RELEASE_USERNAME:?}/swarm-curl-daemon:$${DOCKER_RELEASE_TAG:?}" --file docker.swarm-curl-daemon.dockerfile .

# `make clean`
.PHONY: clean
clean:
	rm -rf ./build/
	cd ./cmd/swarm-curl-daemon/ && make clean
