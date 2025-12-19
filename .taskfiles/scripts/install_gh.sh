#!/bin/bash

set -euo pipefail

# Constants
readonly TOOL_NAME="gh"
readonly KEYRING_URL="https://cli.github.com/packages/githubcli-archive-keyring.gpg"
readonly KEYRING_PATH="/etc/apt/keyrings/githubcli-archive-keyring.gpg"
readonly SOURCES_LIST="/etc/apt/sources.list.d/github-cli.list"

# Configuration (can be overridden by env)
VERSION="${1:-${VERSION:-latest}}"

tempFile=""

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
Usage: sudo $0 [VERSION]

Positional arguments:
  VERSION           Version to install (default: latest from apt repo)

Environment variables:
  VERSION           Desired version (default: latest)

Note: This script must be run with sudo or as root to install GitHub CLI system-wide via apt.

Examples:
  sudo $0             # Install latest
  sudo $0 2.40.0      # Install specific version
  VERSION=2.40.0 $0   # Install ia env
EOF
}

cleanup() {
  if [[ -n "${tempFile}" && -f "${tempFile}" ]]; then
    rm -f "${tempFile}"
  fi
}
trap cleanup EXIT INT TERM

check_sudo() {
  if [[ "${EUID}" -ne 0 ]]; then
    die "This script must be run as root or with sudo"
  fi
}

# Show help if requested
if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

# Normalize VERSION: empty/whitespace -> "latest"
if [[ -z "${VERSION//[[:space:]]/}" ]]; then
  VERSION="latest"
fi

# Main execution
check_sudo

log "Installing ${TOOL_NAME} (${VERSION}) via apt repository"

# Check dependencies and install wget if needed
log "Checking dependencies"
if ! type -p wget >/dev/null; then
  log "Installing wget"
  apt-get update || die "Failed to update apt cache"
  apt-get install wget -y || die "Failed to install wget"
fi

# Create keyrings directory
log "Setting up apt keyring"
# shellcheck disable=SC2174
mkdir -p -m 755 /etc/apt/keyrings || die "Failed to create keyrings directory"

# Download and install GPG key
tempFile="$(mktemp)" || die "Failed to create temp file"
if ! wget -nv -O"${tempFile}" "${KEYRING_URL}"; then
  die "Failed to download GPG keyring"
fi

# shellcheck disable=SC2002
cat "${tempFile}" | tee "${KEYRING_PATH}" >/dev/null || die "Failed to install GPG keyring"
chmod go+r "${KEYRING_PATH}" || die "Failed to set keyring permissions"

# Add apt repository
log "Adding GitHub CLI apt repository"
# shellcheck disable=SC2174
mkdir -p -m 755 /etc/apt/sources.list.d || die "Failed to create sources.list.d directory"

arch="$(dpkg --print-architecture)" || die "Failed to detect architecture"
echo "deb [arch=${arch} signed-by=${KEYRING_PATH}] https://cli.github.com/packages stable main" \
  | tee "${SOURCES_LIST}" >/dev/null || die "Failed to add apt repository"

# Update apt cache
log "Updating apt cache"
apt-get update || die "Failed to update apt cache"

# Install gh
log "Installing ${TOOL_NAME}"
if [[ "${VERSION}" == "latest" ]]; then
  apt-get install gh -y || die "Failed to install ${TOOL_NAME}"
else
  # Try to install specific version (format: gh=version)
  apt-get install "gh=${VERSION}" -y || die "Failed to install ${TOOL_NAME} version ${VERSION}"
fi

log "✓ Successfully installed ${TOOL_NAME}"

# Verify installation
"${TOOL_NAME}" --version || die "Installed binary failed to run"
