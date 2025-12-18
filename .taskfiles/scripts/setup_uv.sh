#!/bin/bash

set -euo pipefail

# Constants
readonly TOOL_NAME="uv"
readonly INSTALL_SCRIPT_URL="https://astral.sh/uv/install.sh"

# Configuration (can be overridden by env)
VERSION="${1:-${VERSION:-latest}}"
INSTALL_DIR="${2:-${INSTALL_DIR:-}}"

# Logging helper
log() {
  echo "-> $*" >&2
}

# Error handling helper
die() {
  echo "X Error: $*" >&2
  exit "${2:-1}"
}

# Help message
usage() {
  cat <<EOF
Usage: $0 [VERSION] [INSTALL_DIR]

Positional arguments:
  VERSION           Version to install (default: latest)
  INSTALL_DIR       Custom install directory

Environment variables:
  VERSION           Desired version (default: latest)
  INSTALL_DIR       Install directory override
  GITHUB_TOKEN      GitHub token for API authentication

Examples:
  $0                      # Install latest
  $0 0.5.0                # Install 0.5.0
  $0 0.5.0 ~/.local/bin   # Install 0.5.0 to custom location
  VERSION=0.5.0 $0        # Install 0.5.0 via env
EOF
}

# Show help if requested
if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

# Normalize VERSION: empty/whitespace -> "latest", strip "v" prefix if present
if [[ -z "${VERSION//[[:space:]]/}" ]]; then
  VERSION="latest"
elif [[ "${VERSION}" != "latest" ]]; then
  # uv installer expects version without 'v' prefix
  VERSION="${VERSION#v}"
fi

# Determine install directory
if [[ -z "${INSTALL_DIR}" ]]; then
  if [[ "${EUID}" -eq 0 ]]; then
    INSTALL_DIR="/usr/local/bin"
  else
    INSTALL_DIR="${HOME}/.local/bin"
  fi
fi

# Check dependencies
command -v curl >/dev/null 2>&1 || die "Missing required dependency: curl"

# Create install directory if it doesn't exist
if [[ ! -d "${INSTALL_DIR}" ]]; then
  mkdir -p "${INSTALL_DIR}" || die "Cannot create install directory ${INSTALL_DIR}"
fi

log "Installing ${TOOL_NAME} (${VERSION}) to ${INSTALL_DIR}"

# GitHub API authentication
if [[ -n "${GITHUB_TOKEN:-}" ]]; then
  ghAuthHeader=(-H "Authorization: Bearer ${GITHUB_TOKEN}")
else
  ghAuthHeader=()
fi

log "INSTALL_DIR=${INSTALL_DIR}"
log "VERSION=${VERSION}"

# Execute remote installation script
log "Fetching and executing official installation script"
if ! curl "${ghAuthHeader[@]}" -fsSL --proto '=https' --tlsv1.3 "${INSTALL_SCRIPT_URL}" | env UV_INSTALL_DIR="${INSTALL_DIR}" UV_VERSION="${VERSION}" UV_GITHUB_TOKEN="${GITHUB_TOKEN:-}" sh; then
  die "Installation failed. Check version or network connection."
fi

log "✓ Successfully installed ${TOOL_NAME} to ${INSTALL_DIR}/uv"

# Run tool version to verify
"${INSTALL_DIR}/uv" --version || die "Installed binary failed to run (${INSTALL_DIR}/uv)"
