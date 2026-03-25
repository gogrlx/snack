#!/bin/sh
# Install snack - a unified CLI for system package managers
# Usage: curl -sSfL https://raw.githubusercontent.com/gogrlx/snack/main/install.sh | sh
set -e

REPO="gogrlx/snack"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and arch
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    armv*)   ARCH="arm" ;;
esac

# macOS universal binary
if [ "$OS" = "darwin" ]; then
    ARCH="universal"
fi

echo "Detected: ${OS}/${ARCH}"

# Get latest release tag
TAG="$(curl -sSf "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')"
VERSION="${TAG#v}"

if [ -z "$VERSION" ]; then
    echo "Error: could not determine latest version" >&2
    exit 1
fi

echo "Installing snack ${VERSION}..."

TARBALL="snack-${VERSION}-${OS}-${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${TAG}/${TARBALL}"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

curl -sSfL "$URL" -o "${TMP}/${TARBALL}"
tar xzf "${TMP}/${TARBALL}" -C "$TMP"

if [ -w "$INSTALL_DIR" ]; then
    mv "${TMP}/snack" "${INSTALL_DIR}/snack"
else
    echo "Installing to ${INSTALL_DIR} (requires sudo)..."
    sudo mv "${TMP}/snack" "${INSTALL_DIR}/snack"
fi

chmod +x "${INSTALL_DIR}/snack"
echo "snack ${VERSION} installed to ${INSTALL_DIR}/snack"
