#!/usr/bin/env bash

set -e

# Get to workdir
cd "$(realpath "$(dirname "$(realpath "${BASH_SOURCE[0]}")")")"

mkdir -p ./build/

# Get git branch name and calculate version
version=$(git rev-parse --abbrev-ref HEAD | sed -E 's#^(develop|release)/##' | sed -E 's#^(main|master)$#latest#')

# Get git branch name and determine release type (develop or release)
release_type=$(git rev-parse --abbrev-ref HEAD | grep -q '^release/' && echo 'release' || echo 'develop')

# Get git commit time
commit_time=$(git log -1 --format=%cd --date=format:'%Y%m%d%H%M%S')

# Directory name: swarm-curl-daemon_${version}_${release_type}_${commit_time}
if [ "$release_type" = "release" ]; then
  package_name="swarm-curl-daemon_${version}"
else
  package_name="swarm-curl-daemon_${version}_${release_type}_${commit_time}"
fi
echo "package_name: ${package_name}"

# Create build/${package_name}
rm -rf   "./build/"
mkdir -p "./build/${package_name}/"

# Copy files
rsync -avz ./installer/ "./build/${package_name}/"

# Package
(
  cd ./build/
  tar -cvf   "${package_name}.tar" "${package_name}/"
  xz -v -T 0 "${package_name}.tar"
  md5sum     "${package_name}.tar.xz" | tee "${package_name}.tar.xz.md5sum"
  sha256sum  "${package_name}.tar.xz" | tee "${package_name}.tar.xz.sha256sum"
)