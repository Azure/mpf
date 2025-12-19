#!/bin/bash

set -euo pipefail

# Constants
readonly TOOL_NAME="pwsh"
readonly KEYRING_URL_BASE="https://packages.microsoft.com/config"
readonly PACKAGES_DEB="packages-microsoft-prod.deb"

# Configuration (can be overridden by env)
VERSION="${1:-${VERSION:-latest}}"

tempFile=""

log() {
  echo "-> $*" >&2
}

die() {
  echo "X Error: $*" >&2
  exit "${2:-1}"
}
usage() {
  cat <<EOF
Usage: $0 [VERSION]

Positional arguments:
  VERSION           Version to install (default: latest from apt repo)

Environment variables:
  VERSION           Desired version (default: latest)

Notes:
  - Linux: installs PowerShell from packages.microsoft.com (preferred method in official docs). Must be run with sudo/root.
  - macOS: installs PowerShell using Homebrew (preferred method in official docs).

Examples:
  sudo $0              # Linux: install latest
  $0                   # macOS: install latest
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

os="$(uname -s | tr '[:upper:]' '[:lower:]')"

if [[ "${os}" == "darwin" ]]; then
  log "Installing ${TOOL_NAME} (${VERSION}) using Homebrew"

  command -v brew >/dev/null 2>&1 || die "brew is not installed. Install it from https://brew.sh/"

  if [[ "${VERSION}" != "latest" ]]; then
    log "Note: version pinning isn't supported by this Homebrew installer path; installing latest stable."
  fi

  brew install --cask powershell || die "Failed to install PowerShell via Homebrew"

  log "✓ Successfully installed ${TOOL_NAME}"
  pwsh --version >/dev/null 2>&1 || die "Installed binary failed to run"
  exit 0
fi

if [[ "${os}" != "linux" ]]; then
  die "Unsupported OS: ${os}"
fi

check_sudo

if [[ "${VERSION}" != "latest" ]]; then
  die "This installer uses the Microsoft package repository method and installs the latest available version. For specific versions, use the official direct-download method."
fi

log "Installing ${TOOL_NAME} (${VERSION}) via Microsoft package repository"

log "Checking dependencies"
for dep in apt-get wget dpkg; do
  command -v "${dep}" >/dev/null 2>&1 || die "Missing required dependency: ${dep}"
done

log "Detecting distribution"
[[ -f /etc/os-release ]] || die "Missing /etc/os-release"
# shellcheck disable=SC1091
source /etc/os-release

case "${ID:-}" in
  debian)
    distro="debian"
    ;;
  ubuntu)
    distro="ubuntu"
    ;;
  *)
    die "Unsupported Linux distribution: ${ID:-unknown}"
    ;;
esac

[[ -n "${VERSION_ID:-}" ]] || die "Unable to determine VERSION_ID from /etc/os-release"

log "Updating package lists"
apt-get update || die "Failed to update apt cache"

log "Installing prerequisites"
apt-get install -y wget || die "Failed to install prerequisites"

log "Setting up Microsoft repository"
tempFile="${PACKAGES_DEB}"
repoUrl="${KEYRING_URL_BASE}/${distro}/${VERSION_ID}/${PACKAGES_DEB}"

if ! wget -q "${repoUrl}"; then
  die "Failed to download Microsoft repository configuration: ${repoUrl}"
fi

dpkg -i "${PACKAGES_DEB}" || die "Failed to register Microsoft repository keys"
rm -f "${PACKAGES_DEB}" || true
tempFile=""

log "Updating package lists (post-repo)"
apt-get update || die "Failed to update apt cache"

log "Installing PowerShell"
apt-get install -y powershell || die "Failed to install ${TOOL_NAME}"

log "✓ Successfully installed ${TOOL_NAME}"
"${TOOL_NAME}" -Version >/dev/null 2>&1 || die "Installed binary failed to run"
