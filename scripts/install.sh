#!/bin/sh
# Install script for arq — https://github.com/orangekame3/arq
# Usage: curl -fsSL https://raw.githubusercontent.com/orangekame3/arq/main/scripts/install.sh | sh
set -eu

REPO="orangekame3/arq"
INSTALL_DIR="${ARQ_INSTALL_DIR:-${HOME}/.local/bin}"

info() { printf '\033[1;34m%s\033[0m\n' "$*"; }
error() { printf '\033[1;31merror: %s\033[0m\n' "$*" >&2; exit 1; }

# Detect OS
case "$(uname -s)" in
  Linux*)  OS="Linux" ;;
  Darwin*) OS="Darwin" ;;
  *)       error "unsupported OS: $(uname -s)" ;;
esac

# Detect architecture
case "$(uname -m)" in
  x86_64|amd64) ARCH="x86_64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)             error "unsupported architecture: $(uname -m)" ;;
esac

# Resolve latest version from GitHub API
info "Fetching latest release..."
if command -v curl >/dev/null 2>&1; then
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed 's/.*"v\(.*\)".*/\1/')
elif command -v wget >/dev/null 2>&1; then
  VERSION=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed 's/.*"v\(.*\)".*/\1/')
else
  error "curl or wget is required"
fi

[ -z "$VERSION" ] && error "failed to determine latest version"

# Build download URL
ARCHIVE="arq_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/v${VERSION}/${ARCHIVE}"

info "Installing arq v${VERSION} (${OS}/${ARCH})..."

# Download and extract
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$URL" -o "${TMPDIR}/${ARCHIVE}"
else
  wget -q "$URL" -O "${TMPDIR}/${ARCHIVE}"
fi

tar -xzf "${TMPDIR}/${ARCHIVE}" -C "$TMPDIR"

# Install binary
mkdir -p "$INSTALL_DIR"
install -m 755 "${TMPDIR}/arq" "${INSTALL_DIR}/arq"

info "Installed arq to ${INSTALL_DIR}/arq"

# Verify
if "${INSTALL_DIR}/arq" --version >/dev/null 2>&1; then
  info "$(${INSTALL_DIR}/arq --version)"
else
  error "installation verification failed"
fi

# PATH hint
case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    info ""
    info "Add ${INSTALL_DIR} to your PATH:"
    info "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    ;;
esac
