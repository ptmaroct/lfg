#!/usr/bin/env bash
# lfg installer — fetches the latest GitHub release tarball, extracts
# the binary, drops it in $PREFIX/bin (default /usr/local/bin, falls
# back to ~/.local/bin when /usr/local isn't writable).
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/ptmaroct/lfg/main/install.sh | sh
#
# Env overrides:
#   LFG_VERSION  — pin to a specific tag (default: latest)
#   LFG_PREFIX   — install root (default: /usr/local or ~/.local)

set -euo pipefail

OWNER="ptmaroct"
REPO="lfg"
LFG_VERSION="${LFG_VERSION:-latest}"

err()  { printf '\033[31merror:\033[0m %s\n' "$*" >&2; exit 1; }
info() { printf '\033[36m==>\033[0m %s\n' "$*"; }

require() {
  command -v "$1" >/dev/null 2>&1 || err "$1 is required but not installed"
}

require curl
require tar
require uname

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) err "unsupported architecture: $ARCH" ;;
esac

case "$OS" in
  darwin|linux) ;;
  *) err "unsupported OS: $OS (lfg supports macOS and Linux)" ;;
esac

# Resolve version → release tag.
if [ "$LFG_VERSION" = "latest" ]; then
  info "resolving latest release..."
  LFG_VERSION="$(curl -fsSL "https://api.github.com/repos/$OWNER/$REPO/releases/latest" \
    | grep -E '"tag_name"' | head -1 | sed -E 's/.*"([^"]+)".*/\1/')"
  [ -n "$LFG_VERSION" ] || err "could not resolve latest version"
fi
info "installing lfg $LFG_VERSION ($OS/$ARCH)"

# Pick install prefix.
if [ -z "${LFG_PREFIX:-}" ]; then
  if [ -w /usr/local/bin ] 2>/dev/null || { [ -d /usr/local/bin ] && [ "$(id -u)" = 0 ]; }; then
    LFG_PREFIX="/usr/local"
  else
    LFG_PREFIX="$HOME/.local"
    mkdir -p "$LFG_PREFIX/bin"
  fi
fi
DEST="$LFG_PREFIX/bin/lfg"

# Download + extract.
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
VER_NUM="${LFG_VERSION#v}"
TARBALL="lfg_${VER_NUM}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$OWNER/$REPO/releases/download/$LFG_VERSION/$TARBALL"

info "downloading $URL"
curl -fsSL -o "$TMP/$TARBALL" "$URL" || err "download failed"
tar -xzf "$TMP/$TARBALL" -C "$TMP"

mv "$TMP/lfg" "$DEST"
chmod +x "$DEST"

info "installed to $DEST"
case ":$PATH:" in
  *":$LFG_PREFIX/bin:"*) ;;
  *) info "note: $LFG_PREFIX/bin is not on your PATH; add it to your shell rc" ;;
esac
"$DEST" version || true
