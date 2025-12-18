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
$GitHubRepo = 'ghalint'
$ToolName = 'ghalint'

function Write-Log {
    param([Parameter(Mandatory)] [string]$Message)
    Write-Host "-> $Message"
}

function Show-Usage {
    @"
Usage: install_ghalint.ps1 [VERSION] [INSTALL_DIR]

Positional arguments:
  VERSION           Version to install (default: latest)
  INSTALL_DIR       Custom install directory

Environment variables:
  VERSION           Desired version (default: latest)
  INSTALL_DIR       Install directory override
  GITHUB_TOKEN      GitHub token for API authentication

Examples:
  ./install_ghalint.ps1                      # Install latest
  ./install_ghalint.ps1 1.5.4                # Install 1.5.4
  ./install_ghalint.ps1 1.5.4 "$HOME\.local\bin"   # Install 1.5.4 to a custom dir
"@
}

# Help
if ($args -contains '-h' -or $args -contains '--help') {
    Show-Usage
    exit 0
}

# Normalize version
if ([string]::IsNullOrWhiteSpace($Version)) {
    $Version = 'latest'
}

$versionTag = if ($Version -eq 'latest') {
    'latest'
}
elseif ($Version -match '^v') {
    $Version
}
else {
    "v$Version"
}

# Determine install directory (simple, user-first default)
if ([string]::IsNullOrWhiteSpace($InstallDir)) {
    $InstallDir = Join-Path $HOME '.local\bin'
}

if (-not (Test-Path -LiteralPath $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# Detect architecture
$arch = switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture) {
    'X64' { 'amd64' }
    'Arm64' { 'arm64' }
    default { throw "Unsupported architecture: $([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture)" }
}

Write-Log "Installing $ToolName ($versionTag) to $InstallDir"

# GitHub API
$headers = @{ 'Accept' = 'application/vnd.github+json' }
if (-not [string]::IsNullOrWhiteSpace($env:GITHUB_TOKEN)) {
    $headers['Authorization'] = "Bearer $($env:GITHUB_TOKEN)"
}

$apiUrl = if ($versionTag -eq 'latest') {
    "https://api.github.com/repos/$GitHubOwner/$GitHubRepo/releases/latest"
}
else {
    "https://api.github.com/repos/$GitHubOwner/$GitHubRepo/releases/tags/$versionTag"
}

$tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString('n'))
New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

try {
    Write-Log 'Fetching release information'
    $release = Invoke-RestMethod -Headers $headers -Uri $apiUrl

    # Asset pattern (example): ghalint_1.5.4_windows_amd64.zip
    $verNoV = if ($versionTag -eq 'latest') {
        # use tag_name to compute asset name
        ($release.tag_name -replace '^v', '')
    }
    else {
        ($versionTag -replace '^v', '')
    }

    $assetName = "${ToolName}_${verNoV}_windows_${arch}.zip"
    $asset = $release.assets | Where-Object { $_.name -eq $assetName } | Select-Object -First 1
    if (-not $asset) {
        throw "No asset found for architecture $arch (expected $assetName)"
    }

    $zipPath = Join-Path $tempDir $assetName
    Write-Log "Downloading $($asset.browser_download_url)"
    Invoke-WebRequest -Headers $headers -Uri $asset.browser_download_url -OutFile $zipPath

    Write-Log 'Extracting archive'
    Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force

    $exeSource = Join-Path $tempDir "$ToolName.exe"
    if (-not (Test-Path -LiteralPath $exeSource)) {
        $found = Get-ChildItem -Path $tempDir -Recurse -Filter "$ToolName.exe" -File -ErrorAction SilentlyContinue | Select-Object -First 1
        if ($found) {
            $exeSource = $found.FullName
        }
    }

    if (-not (Test-Path -LiteralPath $exeSource)) {
        throw "Binary not found in archive ($ToolName.exe)"
    }

    $exeDest = Join-Path $InstallDir "$ToolName.exe"

    Write-Log 'Installing binary'
    Copy-Item -LiteralPath $exeSource -Destination $exeDest -Force

    Write-Log "✓ Successfully installed $ToolName to $exeDest"

    & $exeDest --version | Out-Null
}
finally {
    if (Test-Path -LiteralPath $tempDir) {
        Remove-Item -LiteralPath $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}
