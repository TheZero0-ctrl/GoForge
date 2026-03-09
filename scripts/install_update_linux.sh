#!/bin/bash

set -euo pipefail

# allow specifying different destination directory
DIR="${DIR:-$HOME/.local/bin}"
REPO="${REPO:-TheZero0-ctrl/GoForge}"

# map architecture variations to release binaries
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH (supported: amd64, arm64)" >&2
    exit 1
    ;;
esac

# prepare release version and download URLs
if [[ -n "${VERSION:-}" ]]; then
  TAG="$VERSION"
else
  TAG="$(curl -L -s -H 'Accept: application/json' "https://github.com/${REPO}/releases/latest" | sed -e 's/.*"tag_name":"\([^"]*\)".*/\1/')"
fi

FILE="goforge_${TAG#v}_linux_${ARCH}.tar.gz"
BASE_URL="https://github.com/${REPO}/releases/download/${TAG}"
ARCHIVE_URL="${BASE_URL}/${FILE}"
CHECKSUMS_URL="${BASE_URL}/checksums.txt"

# install/update the local binary
curl -fL -o "$FILE" "$ARCHIVE_URL"
curl -fL -o checksums.txt "$CHECKSUMS_URL"
grep " $FILE$" checksums.txt | sha256sum -c -
tar -xzf "$FILE" goforge
install -Dm 755 goforge "$DIR/goforge"
rm -f goforge "$FILE" checksums.txt

echo "Installed goforge ${TAG} to ${DIR}/goforge"
