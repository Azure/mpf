#!/bin/bash

set -euo pipefail

# Constants
readonly GITHUB_OWNER="suzuki-shunsuke"
readonly GITHUB_REPO="ghalint"
readonly TOOL_NAME="ghalint"

# Configuration (can be overridden by env)
VERSION="${1:-${VERSION:-latest}}"
INSTALL_DIR="${2:-${INSTALL_DIR:-}}"

tempDir=""

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
  $0 1.2.3                # Install 1.2.3
  $0 1.2.3 ~/.local/bin   # Install 1.2.3 to ~/.local/bin
  VERSION=v1.2.3 $0       # Install 1.2.3 via env
EOF
}

cleanup() {
  if [[ -n "${tempDir}" && -d "${tempDir}" ]]; then
    rm -rf "${tempDir}"
  fi
}
trap cleanup EXIT INT TERM

# Show help if requested
if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

# Normalize VERSION: empty/whitespace -> "latest", numeric -> add "v" prefix
if [[ -z "${VERSION//[[:space:]]/}" ]]; then
  VERSION="latest"
elif [[ "${VERSION}" != "latest" && "${VERSION}" =~ ^[0-9] ]]; then
  VERSION="v${VERSION}"
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
for dep in curl jq tar; do
  command -v "${dep}" >/dev/null 2>&1 || die "Missing required dependency: ${dep}"
done

# Create install directory if it doesn't exist
if [[ ! -d "${INSTALL_DIR}" ]]; then
  mkdir -p "${INSTALL_DIR}" || die "Cannot create install directory ${INSTALL_DIR}"
fi

# Detect architecture
archRaw="$(uname -m)"
case "${archRaw}" in
  x86_64 | amd64) arch="amd64" ;;
  arm64 | aarch64) arch="arm64" ;;
  # armv7l | armv6l) arch="arm" ;;
  # i386 | i686) arch="386" ;;
  *) die "Unsupported architecture: ${archRaw}" ;;
esac

log "Installing ${TOOL_NAME} (${VERSION}) to ${INSTALL_DIR}"

# GitHub API authentication
ghAuthHeader=(-H "Accept: application/vnd.github+json")
if [[ -n "${GITHUB_TOKEN:-}" ]]; then
  ghAuthHeader+=(-H "Authorization: Bearer ${GITHUB_TOKEN}")
fi

# Create temp directory
tempDir="$(mktemp -d)" || die "Failed to create temp directory"

# Fetch release info and download URL
log "Fetching release information"
if [[ "${VERSION}" == "latest" ]]; then
  apiUrl="https://api.github.com/repos/${GITHUB_OWNER}/${GITHUB_REPO}/releases/latest"
else
  apiUrl="https://api.github.com/repos/${GITHUB_OWNER}/${GITHUB_REPO}/releases/tags/${VERSION}"
fi

releaseJson="${tempDir}/release.json"
if ! curl "${ghAuthHeader[@]}" -fsSL --proto '=https' --tlsv1.3 "${apiUrl}" -o "${releaseJson}"; then
  die "Failed to fetch release information. Check version or network connection."
fi

# Extract download URL
downloadUrl="$(jq -r --arg arch "${arch}" \
  '.assets[] | select(.browser_download_url | endswith("_linux_\($arch).tar.gz")) | .browser_download_url' \
  "${releaseJson}")"

[[ -n "${downloadUrl}" ]] || die "No asset found for architecture ${arch}"

log "Downloading ${downloadUrl}"
archivePath="${tempDir}/${TOOL_NAME}.tar.gz"
curl "${ghAuthHeader[@]}" -fsSL --proto '=https' --tlsv1.3 "${downloadUrl}" -o "${archivePath}" || die "Download failed"

# Extract binary
log "Extracting archive"
tar -xf "${archivePath}" -C "${tempDir}" "${TOOL_NAME}" || die "Extraction failed"
[[ -f "${tempDir}/${TOOL_NAME}" ]] || die "Binary not found in archive"

# Install binary
log "Installing binary"
mv "${tempDir}/${TOOL_NAME}" "${INSTALL_DIR}/${TOOL_NAME}" || die "Failed to install binary"
chmod 0755 "${INSTALL_DIR}/${TOOL_NAME}" || die "Failed to set permissions"

log "✓ Successfully installed ${TOOL_NAME} to ${INSTALL_DIR}/${TOOL_NAME}"

# Run tool version to verify
"${INSTALL_DIR}/${TOOL_NAME}" --version || die "Installed failed to run (${INSTALL_DIR}/${TOOL_NAME})"
