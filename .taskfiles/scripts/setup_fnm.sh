#!/bin/bash

set -euo pipefail

# Constants
readonly TOOL_NAME="fnm"
readonly INSTALL_SCRIPT_URL="https://fnm.vercel.app/install"

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
  $0                         # Install latest
  $0 1.37.0                  # Install 1.37.0
  $0 1.37.0 ~/.local/fnm     # Install 1.37.0 to custom location
  VERSION=1.37.0 $0          # Install 1.37.0 via env
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
  INSTALL_DIR="${HOME}/.local/share/fnm"
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

# Execute remote installation script
log "Fetching and executing official installation script"
if ! curl "${ghAuthHeader[@]}" -fsSL --proto '=https' --tlsv1.3 "${INSTALL_SCRIPT_URL}" | bash -s -- --skip-shell --install-dir "${INSTALL_DIR}" --release "${VERSION}"; then
  die "Installation failed. Check version or network connection."
fi

log "✓ Successfully installed ${TOOL_NAME} to ${INSTALL_DIR}"

# Run tool version to verify
"${INSTALL_DIR}/fnm" --version || die "Installed binary failed to run (${INSTALL_DIR}/fnm)"

# Create symlink in ~/.local/bin
SYMLINK_DIR="${HOME}/.local/bin"
SYMLINK_PATH="${SYMLINK_DIR}/${TOOL_NAME}"

if [[ ! -d "${SYMLINK_DIR}" ]]; then
  mkdir -p "${SYMLINK_DIR}" || die "Cannot create symlink directory ${SYMLINK_DIR}"
fi

if [[ -L "${SYMLINK_PATH}" ]]; then
  log "Removing existing symlink at ${SYMLINK_PATH}"
  rm -f "${SYMLINK_PATH}"
elif [[ -e "${SYMLINK_PATH}" ]]; then
  die "Cannot create symlink: ${SYMLINK_PATH} already exists and is not a symlink"
fi

ln -s "${INSTALL_DIR}/${TOOL_NAME}" "${SYMLINK_PATH}" || die "Failed to create symlink ${SYMLINK_PATH}"
log "✓ Created symlink: ${SYMLINK_PATH} -> ${INSTALL_DIR}/${TOOL_NAME}"

# Function to add environment variables to shell config
add_to_shell() {
  local shell_config="$1"
  local shell_type="$2"

  if [[ -f "${shell_config}" ]]; then
    # Array of environment variables to add
    local env_vars=(
      "export PATH=\$PATH:${INSTALL_DIR}"
    )

    # Add each variable if not already present
    for str in "${env_vars[@]}"; do
      if ! grep -qF "${str}" "${shell_config}"; then
        echo "${str}" >>"${shell_config}"
      fi
    done

    # Add fnm env command if not already present
    if ! grep -q "fnm env" "${shell_config}"; then
      echo "eval \"\$(fnm env --use-on-cd --shell ${shell_type})\"" >>"${shell_config}"
    fi
  fi
}

# Configure both bash and zsh if present
add_to_shell ~/.bashrc bash
add_to_shell ~/.zshrc zsh
add_to_shell ~/.config/fish/config.fish fish

# Verify installation
"${INSTALL_DIR}/${TOOL_NAME}" --version || die "Installed binary failed to run"

# Since this script runs in bash, only evaluate fnm env for bash
log "Setting up environment for current session"
eval "$("${INSTALL_DIR}/${TOOL_NAME}" env --use-on-cd --shell bash)"
