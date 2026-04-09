#!/bin/bash

set -euo pipefail

# Constants
readonly GITHUB_OWNER="rhysd"
readonly GITHUB_REPO="actionlint"
readonly TOOL_NAME="actionlint"
# Installer script SHA provided via INSTALLER_SHA env var from Taskfile
# To update: change the SHA in github.Taskfile.yml GITHUB_INSTALLER_SHA.actionlint
readonly INSTALL_SCRIPT_SHA="${INSTALLER_SHA:?INSTALLER_SHA env var is required — set in Taskfile}"
readonly INSTALL_SCRIPT_URL="https://raw.githubusercontent.com/${GITHUB_OWNER}/${GITHUB_REPO}/${INSTALL_SCRIPT_SHA}/scripts/download-actionlint.bash"

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

cleanup() {
  if [[ -n "${tempDir}" && -d "${tempDir}" ]]; then
    rm -rf "${tempDir}"
  fi
}
trap cleanup EXIT INT TERM

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

# Download installation script to temp file (avoid piping curl to shell)
tempDir="$(mktemp -d)" || die "Failed to create temp directory"
INSTALL_SCRIPT="${tempDir}/download-actionlint.bash"
log "Downloading official installation script (pinned to ${INSTALL_SCRIPT_SHA})"
if ! curl "${ghAuthHeader[@]}" -fsSL "${INSTALL_SCRIPT_URL}" -o "${INSTALL_SCRIPT}"; then
  die "Failed to download installation script. Check network connection."
fi
chmod +x "${INSTALL_SCRIPT}"

# Execute downloaded script
log "Executing installation script"
if ! /bin/bash "${INSTALL_SCRIPT}" "${VERSION}" "${INSTALL_DIR}"; then
  die "Installation failed. Check version or network connection."
fi

log "✓ Successfully installed ${TOOL_NAME} to ${INSTALL_DIR}/${TOOL_NAME}"

# Run tool version to verify
"${INSTALL_DIR}/${TOOL_NAME}" -version || die "Installed binary failed to run (${INSTALL_DIR}/${TOOL_NAME})"
