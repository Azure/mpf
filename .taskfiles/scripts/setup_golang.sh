#!/bin/bash

set -euo pipefail

# Constants
readonly TOOL_NAME="go"

# Configuration (can be overridden by env)
VERSION="${1:-${VERSION:-latest}}"
INSTALL_DIR="${2:-${INSTALL_DIR:-}}"

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
Usage: $0 [VERSION] [INSTALL_DIR]

Positional arguments:
  VERSION           Version to install (default: latest)
  INSTALL_DIR       Custom install directory

Environment variables:
  VERSION           Desired version (default: latest)
  INSTALL_DIR       Install directory override

Examples:
  sudo $0                         # Install latest to /usr/local
  sudo $0 1.21.0                  # Install 1.21.0 to /usr/local
  $0 1.21.0 ~/.local/go           # Install 1.21.0 to ~/.local/go
  VERSION=1.21.0 $0               # Install 1.21.0 via env
EOF
}

cleanup() {
  if [[ -n "${tempFile}" && -f "${tempFile}" ]]; then
    rm -f "${tempFile}"
  fi
}
trap cleanup EXIT INT TERM

# Show help if requested
if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

# Normalize VERSION: empty/whitespace -> "latest"
if [[ -z "${VERSION//[[:space:]]/}" ]]; then
  VERSION="latest"
fi

# Check dependencies
for dep in curl jq wget; do
  command -v "${dep}" >/dev/null 2>&1 || die "Missing required dependency: ${dep}"
done

# Determine Go version to install
if [[ "${VERSION}" != "latest" ]]; then
  # Strip 'go' prefix if present and add it back
  VERSION="${VERSION#go}"
  goVersion="go${VERSION}"
else
  log "Fetching latest Go version"
  goVersion=$(curl -fsSL --proto '=https' --tlsv1.3 "https://go.dev/dl/?mode=json" | jq -r '.[0].version')
fi

# Detect architecture
archRaw="$(uname -m)"
case "${archRaw}" in
  x86_64 | amd64) arch="amd64" ;;
  arm64 | aarch64) arch="arm64" ;;
  armv7l | armv6l) arch="armv6l" ;;
  i386 | i686) arch="386" ;;
  *) die "Unsupported architecture: ${archRaw}" ;;
esac

goFile="${goVersion}.linux-${arch}.tar.gz"

# Determine install directory and installation mode
if [[ -z "${INSTALL_DIR}" ]]; then
  if [[ "${EUID}" -eq 0 ]]; then
    INSTALL_DIR="/usr/local"
    installMode="system"
  else
    INSTALL_DIR="${HOME}/.local"
    installMode="user"
  fi
else
  # Custom install dir - determine mode based on permissions
  if [[ "${EUID}" -eq 0 ]]; then
    installMode="system"
  else
    installMode="user"
  fi
fi

goInstallPath="${INSTALL_DIR}/go"

# Create install directory if it doesn't exist
if [[ ! -d "${INSTALL_DIR}" ]]; then
  mkdir -p "${INSTALL_DIR}" || die "Cannot create install directory ${INSTALL_DIR}"
fi

log "Installing ${TOOL_NAME} (${goVersion}) for ${arch} to ${goInstallPath}"

# Check if Go is already installed and matches the desired version
if [[ -x "${goInstallPath}/bin/go" ]]; then
  installedVersion=$("${goInstallPath}/bin/go" version | awk '{print $3}')
  if [[ "${installedVersion}" == "${goVersion}" ]]; then
    log "✓ Go ${goVersion} is already installed. Skipping installation."
    exit 0
  else
    log "Go ${installedVersion} is installed, but ${goVersion} is required. Updating..."
  fi
fi

# Download Go
log "Downloading ${goFile}"
tempFile="${goFile}"
if ! wget -q "https://golang.org/dl/${goFile}"; then
  die "Failed to download Go ${goVersion}"
fi

# Remove existing installation and extract new one
log "Extracting Go to ${INSTALL_DIR}"
rm -rf "${goInstallPath}"
tar -C "${INSTALL_DIR}" -xzf "${goFile}" || die "Failed to extract Go archive"

log "✓ Successfully installed ${TOOL_NAME} to ${goInstallPath}"

# Configure environment based on installation mode
if [[ "${installMode}" == "system" ]]; then
  # System-wide installation - create profile.d script
  if [[ ! -f /etc/profile.d/golang.sh ]]; then
    log "Setting up system-wide environment variables"
    # shellcheck disable=SC2016
    echo "export PATH=\"\$PATH:${goInstallPath}/bin\"" | tee /etc/profile.d/golang.sh >/dev/null
  fi
else
  # User installation - add to shell configs
  log "Setting up user environment variables"

  # Function to add environment variables to shell config
  add_to_shell() {
    local shell_config="$1"

    if [[ -f "${shell_config}" ]]; then
      # Array of environment variables to add
      # shellcheck disable=SC2016
      local env_vars=(
        "export PATH=\$PATH:${goInstallPath}/bin"
        'export GOPATH=$HOME/go'
        'export PATH=$PATH:$GOPATH/bin'
      )

      # Add each variable if not already present
      for str in "${env_vars[@]}"; do
        if ! grep -qF "${str}" "${shell_config}"; then
          echo "${str}" >>"${shell_config}"
        fi
      done
    fi
  }

  # Configure both bash and zsh if present
  add_to_shell ~/.bashrc
  add_to_shell ~/.zshrc

  log "Note: Restart your shell or run 'source ~/.bashrc' to use Go"
fi

# Verify installation
"${goInstallPath}/bin/go" version || die "Installed binary failed to run"
