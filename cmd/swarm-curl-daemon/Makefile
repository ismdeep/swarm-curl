.PHONY: help
help:
	@cat Makefile | grep '# `' | grep -v '@cat Makefile'

# `make build`
.PHONY: build
build:
	bash release-01-build.sh

# `make pack`
.PHONY: pack
pack:
	bash release-02-pack.sh

# `make clean`
.PHONY: clean
clean:
	rm -rf ./installer/binary/
	rm -rf ./build/
