[CmdletBinding()]
param(
    [Parameter(Position = 0)]
    [string]$Version = $env:VERSION,

    [Parameter(Position = 1)]
    [string]$InstallDir = $env:INSTALL_DIR
)

$ErrorActionPreference = 'Stop'

# Constants
$GitHubOwner = 'suzuki-shunsuke'
$GitHubRepo = 'pinact'
$ToolName = 'pinact'

function Write-Log {
    param([Parameter(Mandatory)] [string]$Message)
    Write-Host "-> $Message"
}

function Die {
    param(
        [Parameter(Mandatory)] [string]$Message,
        [int]$ExitCode = 1
    )
    Write-Error "X Error: $Message"
    exit $ExitCode
}

function Show-Usage {
    @"
Usage: install_pinact.ps1 [VERSION] [INSTALL_DIR]

Positional arguments:
  VERSION           Version to install (default: latest)
  INSTALL_DIR       Custom install directory

Environment variables:
  VERSION           Desired version (default: latest)
  INSTALL_DIR       Install directory override
  GITHUB_TOKEN      GitHub token for API authentication

Examples:
  ./install_pinact.ps1                      # Install latest
  ./install_pinact.ps1 3.6.0                # Install 3.6.0
  ./install_pinact.ps1 3.6.0 "$HOME\.local\bin"   # Install 3.6.0 to a custom dir
"@
}

# Help
if ($args -contains '-h' -or $args -contains '--help' -or $PSBoundParameters.ContainsKey('Help')) {
    Show-Usage
    exit 0
}

# Normalize version
if ([string]::IsNullOrWhiteSpace($Version)) {
    $Version = 'latest'
}
elseif ($Version -match '^[0-9]') {
    $Version = "v$Version"
}

# Determine install directory (simple, user-first default)
if ([string]::IsNullOrWhiteSpace($InstallDir)) {
    $InstallDir = Join-Path $HOME '.local\bin'
}

# Ensure install directory exists
if (-not (Test-Path -LiteralPath $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# Detect architecture
$arch = switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture) {
    'X64' { 'amd64' }
    'Arm64' { 'arm64' }
    default { Die "Unsupported architecture: $([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture)" }
}

Write-Log "Installing $ToolName ($Version) to $InstallDir"

# GitHub API
$headers = @{ 'Accept' = 'application/vnd.github+json' }
if (-not [string]::IsNullOrWhiteSpace($env:GITHUB_TOKEN)) {
    $headers['Authorization'] = "Bearer $($env:GITHUB_TOKEN)"
}

$apiUrl = if ($Version -eq 'latest') {
    "https://api.github.com/repos/$GitHubOwner/$GitHubRepo/releases/latest"
}
else {
    "https://api.github.com/repos/$GitHubOwner/$GitHubRepo/releases/tags/$Version"
}

$tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString('n'))
New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

try {
    Write-Log 'Fetching release information'
    $release = Invoke-RestMethod -Headers $headers -Uri $apiUrl

    $assetName = "${ToolName}_windows_${arch}.zip"
    $asset = $release.assets | Where-Object { $_.name -eq $assetName } | Select-Object -First 1
    if (-not $asset) {
        Die "No asset found for architecture $arch (expected $assetName)"
    }

    $zipPath = Join-Path $tempDir $assetName
    Write-Log "Downloading $($asset.browser_download_url)"
    Invoke-WebRequest -Headers $headers -Uri $asset.browser_download_url -OutFile $zipPath

    Write-Log 'Extracting archive'
    Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force

    $exeSource = Join-Path $tempDir "$ToolName.exe"
    if (-not (Test-Path -LiteralPath $exeSource)) {
        # Fallback: search anywhere in extracted tree
        $found = Get-ChildItem -Path $tempDir -Recurse -Filter "$ToolName.exe" -File -ErrorAction SilentlyContinue | Select-Object -First 1
        if ($found) {
            $exeSource = $found.FullName
        }
    }

    if (-not (Test-Path -LiteralPath $exeSource)) {
        Die "Binary not found in archive ($ToolName.exe)"
    }

    $exeDest = Join-Path $InstallDir "$ToolName.exe"

    Write-Log 'Installing binary'
    Copy-Item -LiteralPath $exeSource -Destination $exeDest -Force

    Write-Log "✓ Successfully installed $ToolName to $exeDest"

    # Verify
    & $exeDest version | Out-Null
}
finally {
    if (Test-Path -LiteralPath $tempDir) {
        Remove-Item -LiteralPath $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}
