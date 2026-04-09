#!/bin/bash

set -euo pipefail

# Constants
readonly TOOL_NAME="azd"
# Installer script SHA provided via INSTALLER_SHA env var from Taskfile
# To update: change the SHA in azure.Taskfile.yml AZURE_INSTALLER_SHA.azd
readonly INSTALL_SCRIPT_SHA="${INSTALLER_SHA:?INSTALLER_SHA env var is required — set in Taskfile}"
[[ "${INSTALL_SCRIPT_SHA}" =~ ^[0-9a-fA-F]{40}$ ]] || { echo "X Error: INSTALLER_SHA must be a full 40-character hexadecimal commit SHA" >&2; exit 1; }
readonly INSTALL_SCRIPT_URL="https://raw.githubusercontent.com/Azure/azure-dev/${INSTALL_SCRIPT_SHA}/cli/installer/install-azd.sh"
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

Environment variables (required):
  INSTALLER_SHA     Full 40-char hex commit SHA for the installer script (set by Taskfile)

Environment variables (optional):
  VERSION           Desired version (default: latest)
  INSTALL_DIR       Install directory override
  GITHUB_TOKEN      GitHub token for API authentication

Examples:
  INSTALLER_SHA=<sha> $0                      # Install latest
  INSTALLER_SHA=<sha> $0 1.2.3                # Install 1.2.3
  INSTALLER_SHA=<sha> $0 1.2.3 ~/.local/bin   # Install 1.2.3 to ~/.local/bin

Note: Normally invoked via Taskfile (e.g., task install:azd), which sets INSTALLER_SHA automatically.
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

# Download installation script to temp file (avoid piping curl to shell)
tempDir="$(mktemp -d)" || die "Failed to create temp directory"
INSTALL_SCRIPT="${tempDir}/install-azd.sh"
log "Downloading official installation script (pinned to ${INSTALL_SCRIPT_SHA})"
if ! curl -fsSL "${INSTALL_SCRIPT_URL}" -o "${INSTALL_SCRIPT}"; then
  die "Failed to download installation script. Check network connection."
fi
chmod +x "${INSTALL_SCRIPT}"

# Execute downloaded script
log "Executing installation script"
if ! /bin/bash "${INSTALL_SCRIPT}" --version "${VERSION}" --install-folder "${INSTALL_DIR}" --symlink-folder "${INSTALL_DIR}"; then
  die "Installation failed. Check version or network connection."
fi

log "✓ Successfully installed ${TOOL_NAME} to ${INSTALL_DIR}/${TOOL_NAME}"

# Run tool version to verify
"${INSTALL_DIR}/${TOOL_NAME}" version || die "Installed binary failed to run (${INSTALL_DIR}/${TOOL_NAME})"
