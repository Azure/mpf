#!/bin/bash

set -euo pipefail

# Constants
readonly TOOL_NAME="az"
readonly KEYRING_URL="https://packages.microsoft.com/keys/microsoft.asc"
readonly KEYRING_PATH="/etc/apt/keyrings/microsoft.gpg"
readonly SOURCES_LIST="/etc/apt/sources.list.d/azure-cli.sources"

# Configuration (can be overridden by env)
VERSION="${1:-${VERSION:-latest}}"

log() {
  echo "-> $*" >&2
}

die() {
  echo "X Error: $*" >&2
  exit "${2:-1}"
}

usage() {
  cat <<EOF
Usage: sudo $0 [VERSION]

Positional arguments:
  VERSION           Version to install (default: latest from apt repo)

Environment variables:
  VERSION           Desired version (default: latest)

Note: This script must be run with sudo or as root to install Azure CLI system-wide via apt.

Examples:
  sudo $0              # Install latest
  sudo $0 2.51.0       # Install specific version
  VERSION=2.51.0 $0    # Install via env
EOF
}

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

check_sudo

log "Installing ${TOOL_NAME} (${VERSION}) via apt repository"

# Check dependencies
log "Checking dependencies"
for dep in curl gpg lsb_release; do
  command -v "${dep}" >/dev/null 2>&1 || die "Missing required dependency: ${dep}"
done

# Install apt transport dependencies
log "Installing apt dependencies"
apt-get update || die "Failed to update apt cache"
apt-get install -y apt-transport-https ca-certificates gnupg || die "Failed to install apt dependencies"

# Create keyrings directory
log "Setting up Microsoft GPG key"
# shellcheck disable=SC2174
mkdir -p -m 755 /etc/apt/keyrings || die "Failed to create keyrings directory"

# Download and install GPG key
if ! curl -sLS "${KEYRING_URL}" | gpg --batch --yes --dearmor -o "${KEYRING_PATH}"; then
  die "Failed to download and install GPG keyring"
fi
chmod go+r "${KEYRING_PATH}" || die "Failed to set keyring permissions"

# Add apt repository using DEB822 format
log "Adding Azure CLI apt repository"
arch="$(dpkg --print-architecture)" || die "Failed to detect architecture"
codename="$(lsb_release -cs)" || die "Failed to detect distribution codename"

cat <<EOF | tee "${SOURCES_LIST}" >/dev/null || die "Failed to add apt repository"
Types: deb
URIs: https://packages.microsoft.com/repos/azure-cli/
Suites: ${codename}
Components: main
Architectures: ${arch}
Signed-by: ${KEYRING_PATH}
EOF

# Update apt cache and install
log "Updating apt cache"
apt-get update || die "Failed to update apt cache"

log "Installing ${TOOL_NAME}"
if [[ "${VERSION}" == "latest" ]]; then
  apt-get install -y azure-cli || die "Failed to install ${TOOL_NAME}"
else
  # Install specific version (format: azure-cli=version-1~codename)
  apt-get install -y "azure-cli=${VERSION}-1~${codename}" || die "Failed to install ${TOOL_NAME} version ${VERSION}"
fi

log "✓ Successfully installed ${TOOL_NAME}"

# Verify installation
"${TOOL_NAME}" --version || die "Installed binary failed to run"
