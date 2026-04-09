#!/bin/bash

set -euo pipefail

# Constants
readonly TOOL_NAME="fnm"
# Installer script SHA provided via INSTALLER_SHA env var from Taskfile
# To update: change the SHA in runtime.Taskfile.yml RUNTIME_INSTALLER_SHA.fnm
readonly INSTALL_SCRIPT_SHA="${INSTALLER_SHA:?INSTALLER_SHA env var is required — set in Taskfile}"
[[ "${INSTALL_SCRIPT_SHA}" =~ ^[0-9a-fA-F]{40}$ ]] || { echo "X Error: INSTALLER_SHA must be a full 40-character hexadecimal commit SHA" >&2; exit 1; }
readonly INSTALL_SCRIPT_URL="https://raw.githubusercontent.com/Schniz/fnm/${INSTALL_SCRIPT_SHA}/.ci/install.sh"
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
  INSTALLER_SHA=<sha> $0                         # Install latest
  INSTALLER_SHA=<sha> $0 1.37.0                  # Install 1.37.0
  INSTALLER_SHA=<sha> $0 1.37.0 ~/.local/fnm     # Install 1.37.0 to custom location

Note: Normally invoked via Taskfile (e.g., task setup:fnm), which sets INSTALLER_SHA automatically.
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

# Download installation script to temp file (avoid piping curl to shell)
tempDir="$(mktemp -d)" || die "Failed to create temp directory"
INSTALL_SCRIPT="${tempDir}/install.sh"
log "Downloading official installation script (pinned to ${INSTALL_SCRIPT_SHA})"
if ! curl -fsSL "${INSTALL_SCRIPT_URL}" -o "${INSTALL_SCRIPT}"; then
  die "Failed to download installation script. Check network connection."
fi
chmod +x "${INSTALL_SCRIPT}"

# Execute downloaded script
log "Executing installation script"
if ! /bin/bash "${INSTALL_SCRIPT}" --skip-shell --install-dir "${INSTALL_DIR}" --release "${VERSION}"; then
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
