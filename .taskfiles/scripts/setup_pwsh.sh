#!/bin/bash

set -euo pipefail

# Constants
readonly TOOL_NAME="pwsh"
readonly KEYRING_URL_BASE="https://packages.microsoft.com/config"
readonly PACKAGES_DEB="packages-microsoft-prod.deb"

# Configuration (can be overridden by env)
VERSION="${1:-${VERSION:-latest}}"

tempFile=""
tempDir=""

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
  PWSH_SHA256_LINUX_X64    Expected SHA256 for pinned linux-x64 archive
  PWSH_SHA256_LINUX_ARM64  Expected SHA256 for pinned linux-arm64 archive
  PWSH_SHA256_OSX_X64      Expected SHA256 for pinned osx-x64 archive
  PWSH_SHA256_OSX_ARM64    Expected SHA256 for pinned osx-arm64 archive

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
  if [[ -n "${tempDir}" && -d "${tempDir}" ]]; then
    rm -rf "${tempDir}"
  fi
}
trap cleanup EXIT INT TERM

check_sudo() {
  if [[ "${EUID}" -ne 0 ]]; then
    die "This script must be run as root or with sudo"
  fi
}

verify_sha256() {
  local filePath="$1"
  local expectedSha="$2"
  local actualSha

  [[ -n "${expectedSha}" ]] || die "Missing expected SHA256 for pinned PowerShell archive"

  if command -v sha256sum >/dev/null 2>&1; then
    actualSha="$(sha256sum "${filePath}" | awk '{print $1}')"
  elif command -v shasum >/dev/null 2>&1; then
    actualSha="$(shasum -a 256 "${filePath}" | awk '{print $1}')"
  else
    die "Missing required dependency: sha256sum or shasum"
  fi

  [[ "${actualSha}" == "${expectedSha}" ]] || die "SHA256 mismatch for ${filePath}"
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
version="${VERSION#v}"

if [[ "${os}" == "darwin" ]]; then
  if [[ "${VERSION}" == "latest" ]]; then
    log "Installing ${TOOL_NAME} (${VERSION}) using Homebrew"
    command -v brew >/dev/null 2>&1 || die "brew is not installed. Install it from https://brew.sh/"
    brew install --cask powershell || die "Failed to install PowerShell via Homebrew"
    log "✓ Successfully installed ${TOOL_NAME}"
    pwsh --version >/dev/null 2>&1 || die "Installed binary failed to run"
    exit 0
  fi

  log "Installing ${TOOL_NAME} (${version}) from pinned release archive"

  for dep in curl tar; do
    command -v "${dep}" >/dev/null 2>&1 || die "Missing required dependency: ${dep}"
  done

  arch="$(uname -m)"
  case "${arch}" in
    x86_64 | amd64)
      asset_arch="x64"
      expectedSha="${PWSH_SHA256_OSX_X64:-}"
      ;;
    arm64 | aarch64)
      asset_arch="arm64"
      expectedSha="${PWSH_SHA256_OSX_ARM64:-}"
      ;;
    *) die "Unsupported architecture: ${arch}" ;;
  esac

  tempDir="$(mktemp -d)" || die "Failed to create temporary directory"
  installDir="${HOME}/.local/powershell/${version}"
  binDir="${HOME}/.local/bin"
  archivePath="${tempDir}/powershell.tar.gz"
  downloadUrl="https://github.com/PowerShell/PowerShell/releases/download/v${version}/powershell-${version}-osx-${asset_arch}.tar.gz"

  mkdir -p "${installDir}" "${binDir}" || die "Failed to create install directories"
  curl -fsSL "${downloadUrl}" -o "${archivePath}" || die "Failed to download PowerShell ${version}"
  verify_sha256 "${archivePath}" "${expectedSha}"
  tar -xzf "${archivePath}" -C "${installDir}" || die "Failed to extract PowerShell ${version}"
  chmod +x "${installDir}/pwsh" || die "Failed to make pwsh executable"
  ln -sf "${installDir}/pwsh" "${binDir}/pwsh" || die "Failed to link pwsh"

  log "✓ Successfully installed ${TOOL_NAME}"
  "${binDir}/pwsh" --version >/dev/null 2>&1 || die "Installed binary failed to run"
  exit 0
fi

if [[ "${os}" != "linux" ]]; then
  die "Unsupported OS: ${os}"
fi

arch="$(uname -m)"

if [[ "${VERSION}" == "latest" && ("${arch}" == "aarch64" || "${arch}" == "arm64") ]]; then
  log "⚠ PowerShell is not available via apt for ARM64 architecture."
  log "Installing PowerShell via .NET global tool instead..."

  # Install .NET SDK if not present
  if ! command -v dotnet >/dev/null 2>&1; then
    check_sudo
    log "Installing .NET SDK..."
    apt-get update
    apt-get install -y dotnet-sdk-8.0 || die "Failed to install .NET SDK"
  fi

  # Install PowerShell as a .NET global tool
  dotnet tool install --global PowerShell || log "PowerShell may already be installed"

  # Add to PATH if needed
  export PATH="$PATH:$HOME/.dotnet/tools"
  if command -v pwsh >/dev/null 2>&1; then
    log "✓ Successfully installed ${TOOL_NAME} via .NET global tool"
    exit 0
  else
    log "⚠ PowerShell installation skipped on ARM64 - not critical for development"
    exit 0
  fi
fi

check_sudo

if [[ "${VERSION}" != "latest" ]]; then
  for dep in curl tar; do
    command -v "${dep}" >/dev/null 2>&1 || die "Missing required dependency: ${dep}"
  done

  case "${arch}" in
    x86_64 | amd64)
      asset_arch="x64"
      expectedSha="${PWSH_SHA256_LINUX_X64:-}"
      ;;
    arm64 | aarch64)
      asset_arch="arm64"
      expectedSha="${PWSH_SHA256_LINUX_ARM64:-}"
      ;;
    *) die "Unsupported architecture: ${arch}" ;;
  esac

  tempDir="$(mktemp -d)" || die "Failed to create temporary directory"
  installDir="/opt/microsoft/powershell/${version}"
  archivePath="${tempDir}/powershell.tar.gz"
  downloadUrl="https://github.com/PowerShell/PowerShell/releases/download/v${version}/powershell-${version}-linux-${asset_arch}.tar.gz"

  log "Installing ${TOOL_NAME} (${version}) from pinned release archive"
  mkdir -p "${installDir}" || die "Failed to create install directory ${installDir}"
  curl -fsSL "${downloadUrl}" -o "${archivePath}" || die "Failed to download PowerShell ${version}"
  verify_sha256 "${archivePath}" "${expectedSha}"
  tar -xzf "${archivePath}" -C "${installDir}" || die "Failed to extract PowerShell ${version}"
  chmod +x "${installDir}/pwsh" || die "Failed to make pwsh executable"
  ln -sf "${installDir}/pwsh" /usr/bin/pwsh || die "Failed to link pwsh"

  log "✓ Successfully installed ${TOOL_NAME}"
  "${TOOL_NAME}" -Version >/dev/null 2>&1 || die "Installed binary failed to run"
  exit 0
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
