<#
.SYNOPSIS
Creates Azure service principals for MPF E2E tests (ecosystem × OS) and optionally configures GitHub repository secrets.

.DESCRIPTION
PowerShell 7+ (pwsh) cross-platform equivalent of scripts/create-e2e-service-principals.sh.
Uses a small in-script matrix (ecosystem × OS) to create/reuse principals dynamically.
#>

[CmdletBinding()]
param(
    [switch] $Yes,
    [switch] $AddToGitHub,
    [string] $OutputFile = 'e2e-service-principals.json'
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Info([string]$m) { Write-Host "[INFO] $m" -ForegroundColor Cyan }
function Ok([string]$m) { Write-Host "[SUCCESS] $m" -ForegroundColor Green }
function Warn([string]$m) { Write-Host "[WARNING] $m" -ForegroundColor Yellow }

function To-TitleCase([string]$s) {
    if (-not $s) { return $s }
    return (Get-Culture).TextInfo.ToTitleCase($s)
}

function Get-SpName([string]$ecosystem, [string]$os) {
    return "mpf-$ecosystem-$os-e2e-sp"
}

function Get-Label([string]$ecosystem, [string]$os) {
    return "MPF $(To-TitleCase $ecosystem) $(To-TitleCase $os) E2E"
}

function Get-SecretPrefix([string]$ecosystem, [string]$os) {
    return "MPF_$($ecosystem.ToUpperInvariant())_$($os.ToUpperInvariant())"
}

if (-not (Get-Command az -ErrorAction SilentlyContinue)) { throw 'az not found. Install Azure CLI: https://learn.microsoft.com/cli/azure/install-azure-cli' }
if (-not (Get-Command gh -ErrorAction SilentlyContinue)) { throw 'gh not found. Install GitHub CLI: https://cli.github.com/' }

Write-Host "========================================" -ForegroundColor Green
Write-Host " MPF E2E Service Principals Creator" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

Info 'Checking prerequisites...'

# JSON parsing is resilient as long as we only parse stdout. Warnings go to stderr.
$acctJson = & az account show --only-show-errors --output json
if ($LASTEXITCODE -ne 0) { throw 'Not logged in to Azure. Run: az login' }
$acct = $acctJson | ConvertFrom-Json

& gh auth status | Out-Null
if ($LASTEXITCODE -ne 0) { throw 'Not logged in to GitHub. Run: gh auth login' }

& gh repo view | Out-Null
if ($LASTEXITCODE -ne 0) { throw 'Not in a GitHub repository (or no remote configured).' }

Ok 'Prerequisites check passed'
Write-Host ""

Info 'Current Azure Context:'
Info "  Subscription: $($acct.name) ($($acct.id))"
Info "  Tenant: $($acct.tenantId)"
Write-Host ""

if (-not $Yes) {
    $confirm = Read-Host 'Do you want to create service principals in this subscription? (Y/n)'
    if ($confirm -in @('n', 'N')) {
        Info 'Operation cancelled by user'
        exit 0
    }
}

$shouldAddToGitHub = $false
if ($AddToGitHub) {
    $shouldAddToGitHub = $true
}
elseif (-not $Yes) {
    $ans = Read-Host 'Do you want to automatically add secrets to GitHub repository? (y/N)'
    $shouldAddToGitHub = ($ans -in @('y', 'Y'))
}

function New-Sp([string]$Name, [string]$Label) {
    Info "Creating service principal: $Label"

    # NOTE: `az ad sp create-for-rbac --skip-assignment` is deprecated.
    # We create an Entra ID app + service principal + client secret explicitly, without RBAC assignments.

    $appIdRaw = & az ad app list --filter "displayName eq '$Name'" --query "[0].appId" --output tsv --only-show-errors 2>$null
    $appId = if ($null -eq $appIdRaw) { '' } else { $appIdRaw.ToString().Trim() }
    if ($LASTEXITCODE -ne 0) { throw "Failed to query existing app registration: $Name" }

    if (-not $appId) {
        $appJson = & az ad app create --display-name $Name --only-show-errors --output json
        if ($LASTEXITCODE -ne 0) { throw "Failed to create app registration: $Name" }
        $app = $appJson | ConvertFrom-Json
        if (-not $app.appId) { throw "Failed to retrieve appId for app registration: $Name" }
        $appId = $app.appId
    }
    else {
        Info "Reusing existing app registration: $Name"
    }

    # Ensure a service principal exists for this appId.
    & az ad sp create --id $appId --only-show-errors --output json 2>$null | Out-Null

    # Entra can be eventually-consistent; retry fetching objectId a few times.
    $objectId = ''
    for ($i = 0; $i -lt 20; $i++) {
        $objectIdRaw = & az ad sp show --id $appId --query id --output tsv --only-show-errors 2>$null
        $objectId = if ($null -eq $objectIdRaw) { '' } else { $objectIdRaw.ToString().Trim() }
        if ($LASTEXITCODE -eq 0 -and $objectId) { break }
        Start-Sleep -Seconds 2
    }
    if (-not $objectId) { throw "Failed to retrieve service principal object id for appId: $appId" }

    # Reset password (rotate secret) for this app.
    $credJson = & az ad app credential reset --id $appId --display-name 'mpf-e2e' --only-show-errors --output json
    if ($LASTEXITCODE -ne 0) { throw "Failed to reset client secret for appId: $appId" }
    if (-not ([string]$credJson).Trim()) { throw "Credential reset returned empty output for appId: $appId" }

    try {
        $cred = $credJson | ConvertFrom-Json
    }
    catch {
        throw "Failed to parse credential reset output for appId: $appId. Raw output: $credJson"
    }

    if (-not $cred -or -not $cred.password) { throw "Failed to retrieve client secret for appId: $appId" }

    Ok "Created service principal: $Label"
    Info "  App ID: $appId"
    Info "  Object ID: $objectId"
    Warn "  Client Secret: [HIDDEN - will be shown in summary]"

    return [pscustomobject]@{
        ClientId     = $appId
        ClientSecret = $cred.password
        ObjectId     = $objectId
    }
}

function Set-GitHubSecrets {
    param(
        [Parameter(Mandatory)]
        [string] $RepoName,
        [Parameter(Mandatory)]
        [string] $Prefix,
        [Parameter(Mandatory)]
        [pscustomobject] $Sp
    )

    & gh secret set "${Prefix}_SPCLIENTID" --body $Sp.ClientId -R $RepoName | Out-Null
    & gh secret set "${Prefix}_SPCLIENTSECRET" --body $Sp.ClientSecret -R $RepoName | Out-Null
    & gh secret set "${Prefix}_SPOBJECTID" --body $Sp.ObjectId -R $RepoName | Out-Null
}

function Set-GitHubSharedSecrets {
    param(
        [Parameter(Mandatory)]
        [string] $RepoName,
        [Parameter(Mandatory)]
        [string] $TenantId,
        [Parameter(Mandatory)]
        [string] $SubscriptionId
    )

    & gh secret set 'MPF_TENANTID' --body $TenantId -R $RepoName | Out-Null
    & gh secret set 'MPF_SUBSCRIPTIONID' --body $SubscriptionId -R $RepoName | Out-Null
}

Write-Host ""

$ecosystems = @('terraform', 'arm', 'bicep')
$oses = @('linux', 'windows')
$matrix = foreach ($ecosystem in $ecosystems) {
    foreach ($os in $oses) {
        [pscustomobject]@{ ecosystem = $ecosystem; os = $os }
    }
}

Info "Creating service principals for E2E tests ($($matrix.Count) total)..."
Write-Host ""

$servicePrincipals = @{}
foreach ($m in $matrix) {
    $spName = Get-SpName -ecosystem $m.ecosystem -os $m.os
    $label = Get-Label -ecosystem $m.ecosystem -os $m.os
    $key = "$($m.ecosystem)_$($m.os)"
    $servicePrincipals[$key] = New-Sp $spName $label
    Write-Host ""
}

Ok 'All service principals created successfully!'
Write-Host ""

Info "Saving credentials to $($OutputFile)"
$payload = [ordered]@{}
foreach ($m in $matrix) {
    $key = "$($m.ecosystem)_$($m.os)"
    $sp = $servicePrincipals[$key]
    $payload[$key] = [ordered]@{ client_id = $sp.ClientId; client_secret = $sp.ClientSecret; object_id = $sp.ObjectId }
}
$payload['tenant_id'] = $acct.tenantId
$payload['subscription_id'] = $acct.id
$payload | ConvertTo-Json -Depth 5 | Set-Content -Path $OutputFile -Encoding UTF8
Ok "Credentials saved to $($OutputFile)"
Warn 'IMPORTANT: Store this file securely and delete it after configuring GitHub secrets!'
Write-Host ""

if ($shouldAddToGitHub) {
    Info 'Adding secrets to GitHub repository...'
    $repoNameRaw = & gh repo view --json nameWithOwner --jq '.nameWithOwner'
    $repoName = if ($null -eq $repoNameRaw) { '' } else { $repoNameRaw.ToString().Trim() }
    if ($LASTEXITCODE -ne 0 -or -not $repoName) { throw 'Failed to determine repo name (owner/name) via gh repo view.' }
    Info "Repository: $repoName"

    # When multiple remotes exist (e.g., origin + upstream), `gh secret set` can fail with
    # "multiple remotes detected". We explicitly pass -R to avoid ambiguity.
    # Additionally, we set GH_REPO so any subsequent gh commands inherit the context.
    $env:GH_REPO = $repoName

    foreach ($m in $matrix) {
        $key = "$($m.ecosystem)_$($m.os)"
        $prefix = Get-SecretPrefix -ecosystem $m.ecosystem -os $m.os
        Set-GitHubSecrets -RepoName $repoName -Prefix $prefix -Sp $servicePrincipals[$key]
    }

    Set-GitHubSharedSecrets -RepoName $repoName -TenantId $acct.tenantId -SubscriptionId $acct.id

    Ok 'All secrets added to GitHub repository!'
    Write-Host ""
    Info 'You can verify the secrets were added by running:'
    Write-Host 'gh secret list' -ForegroundColor Cyan
}
else {
    Info 'GitHub Repository Secrets Configuration'
    Write-Host ''
    Write-Host ''

    foreach ($m in $matrix) {
        $key = "$($m.ecosystem)_$($m.os)"
        $prefix = Get-SecretPrefix -ecosystem $m.ecosystem -os $m.os
        $sp = $servicePrincipals[$key]
        Write-Host "${prefix}_SPCLIENTID = $($sp.ClientId)"
        Write-Host "${prefix}_SPCLIENTSECRET = $($sp.ClientSecret)"
        Write-Host "${prefix}_SPOBJECTID = $($sp.ObjectId)"
        Write-Host ''
    }

    Write-Host ''
    Write-Host "MPF_TENANTID = $($acct.tenantId)"
    Write-Host "MPF_SUBSCRIPTIONID = $($acct.id)"
}

Write-Host "========================================" -ForegroundColor Green
Write-Host ' Setup Complete!' -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ''
Info 'Next steps:'
if ($shouldAddToGitHub) {
    Write-Host '1. ✅ Secrets have been added to GitHub repository'
    Write-Host '2. Test the updated workflows'
    Write-Host "3. Delete the credentials file: $($OutputFile)"
    Write-Host '4. Verify secrets with: gh secret list'
}
else {
    Write-Host '1. Add the secrets above to your GitHub repository'
    Write-Host '2. Test the updated workflows'
    Write-Host "3. Delete the credentials file: $($OutputFile)"
}
Write-Host ''
Warn 'Remember: These service principals have no permissions by default.'
Warn 'The MPF utility will manage all required permissions dynamically.'
