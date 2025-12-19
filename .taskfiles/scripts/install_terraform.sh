#!/bin/bash

set -euo pipefail

# Constants
readonly TOOL_NAME="terraform"
readonly KEYRING_URL="https://apt.releases.hashicorp.com/gpg"
readonly KEYRING_PATH="/usr/share/keyrings/hashicorp-archive-keyring.gpg"
readonly SOURCES_LIST="/etc/apt/sources.list.d/hashicorp.list"

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

Note: This script must be run with sudo or as root to install Terraform system-wide via apt.

Examples:
  sudo $0              # Install latest
  sudo $0 1.6.0        # Install specific version
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

check_sudo

log "Installing ${TOOL_NAME} (${VERSION}) via apt repository"

# Check dependencies
log "Checking dependencies"
for dep in wget gpg lsb_release; do
  if ! command -v "${dep}" >/dev/null 2>&1; then
    die "Missing required dependency: ${dep}"
  fi
done

# Download and install GPG key
log "Setting up HashiCorp GPG key"
if ! wget -O - "${KEYRING_URL}" | gpg --batch --yes --dearmor -o "${KEYRING_PATH}"; then
  die "Failed to download and install GPG keyring"
fi

# Add apt repository
log "Adding HashiCorp apt repository"
arch="$(dpkg --print-architecture)" || die "Failed to detect architecture"
codename="$(lsb_release -cs)" || die "Failed to detect distribution codename"

echo "deb [arch=${arch} signed-by=${KEYRING_PATH}] https://apt.releases.hashicorp.com ${codename} main" \
  | tee "${SOURCES_LIST}" >/dev/null || die "Failed to add apt repository"

# Update apt cache and install
log "Updating apt cache"
apt-get update || die "Failed to update apt cache"

log "Installing ${TOOL_NAME}"
if [[ "${VERSION}" == "latest" ]]; then
  apt-get install -y terraform || die "Failed to install ${TOOL_NAME}"
else
  # Try to install specific version (format: terraform=version)
  apt-get install -y "terraform=${VERSION}" || die "Failed to install ${TOOL_NAME} version ${VERSION}"
fi

log "✓ Successfully installed ${TOOL_NAME}"

# Verify installation
"${TOOL_NAME}" --version || die "Installed binary failed to run"
