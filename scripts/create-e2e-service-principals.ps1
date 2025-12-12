<#
.SYNOPSIS
Creates three Azure service principals for MPF E2E tests and optionally configures GitHub repository secrets.

.DESCRIPTION
PowerShell 7+ (pwsh) cross-platform equivalent of scripts/create-e2e-service-principals.sh.
Intentionally kept simple: directly invokes az/gh and stores outputs in variables.
#>

[CmdletBinding()]
param(
    [switch] $Yes,
    [switch] $AddToGitHub,
    [Alias('NoAddToGitHub')]
    [switch] $DoNotAddToGitHub,
    [string] $CredentialsFile = 'e2e-service-principals-credentials.json'
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Info([string]$m) { Write-Host "[INFO] $m" -ForegroundColor Cyan }
function Ok([string]$m) { Write-Host "[SUCCESS] $m" -ForegroundColor Green }
function Warn([string]$m) { Write-Host "[WARNING] $m" -ForegroundColor Yellow }

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
    $confirm = Read-Host 'Do you want to create service principals in this subscription? (y/N)'
    if ($confirm -notin @('y', 'Y')) {
        Info 'Operation cancelled by user'
        exit 0
    }
}

$shouldAddToGitHub = $false
if ($DoNotAddToGitHub) {
    $shouldAddToGitHub = $false
}
elseif ($AddToGitHub -or $Yes) {
    $shouldAddToGitHub = $true
}
else {
    $ans = Read-Host 'Do you want to automatically add secrets to GitHub repository? (Y/n)'
    $shouldAddToGitHub = ($ans -notin @('n', 'N'))
}

function New-Sp([string]$Name, [string]$Label) {
    Info "Creating service principal: $Label"

    $spJson = & az ad sp create-for-rbac --name $Name --skip-assignment --only-show-errors --output json
    if ($LASTEXITCODE -ne 0) { throw "Failed to create service principal: $Name" }
    $sp = $spJson | ConvertFrom-Json

    $objectId = (& az ad sp show --id $sp.appId --query id --output tsv --only-show-errors).Trim()
    if ($LASTEXITCODE -ne 0 -or -not $objectId) { throw "Failed to retrieve service principal object id for: $Name" }

    Ok "Created service principal: $Label"
    Info "  App ID: $($sp.appId)"
    Info "  Object ID: $objectId"
    Warn "  Client Secret: [HIDDEN - will be shown in summary]"

    return [pscustomobject]@{
        ClientId     = $sp.appId
        ClientSecret = $sp.password
        ObjectId     = $objectId
    }
}

Write-Host ""
Info 'Creating three service principals for E2E tests...'
Write-Host ""

$terraformSp = New-Sp 'mpf-terraform-e2e-sp' 'MPF Terraform E2E'
Write-Host ""
$armBicepSp = New-Sp 'mpf-arm-bicep-e2e-sp' 'MPF ARM/Bicep E2E'
Write-Host ""
$windowsSp = New-Sp 'mpf-windows-e2e-sp' 'MPF Windows E2E'
Write-Host ""

Ok 'All service principals created successfully!'
Write-Host ""

Info "Saving credentials to $CredentialsFile"
$payload = [ordered]@{
    terraform_e2e   = [ordered]@{ client_id = $terraformSp.ClientId; client_secret = $terraformSp.ClientSecret; object_id = $terraformSp.ObjectId }
    arm_bicep_e2e   = [ordered]@{ client_id = $armBicepSp.ClientId; client_secret = $armBicepSp.ClientSecret; object_id = $armBicepSp.ObjectId }
    windows_e2e     = [ordered]@{ client_id = $windowsSp.ClientId; client_secret = $windowsSp.ClientSecret; object_id = $windowsSp.ObjectId }
    tenant_id       = $acct.tenantId
    subscription_id = $acct.id
}
$payload | ConvertTo-Json -Depth 5 | Set-Content -Path $CredentialsFile -Encoding UTF8
Ok "Credentials saved to $CredentialsFile"
Warn 'IMPORTANT: Store this file securely and delete it after configuring GitHub secrets!'
Write-Host ""

if ($shouldAddToGitHub) {
    Info 'Adding secrets to GitHub repository...'
    $repoName = (& gh repo view --json nameWithOwner --jq '.nameWithOwner').Trim()
    if ($LASTEXITCODE -ne 0 -or -not $repoName) { throw 'Failed to determine repo name (owner/name) via gh repo view.' }
    Info "Repository: $repoName"

    # When multiple remotes exist (e.g., origin + upstream), `gh secret set` can fail with
    # "multiple remotes detected". We explicitly pass -R to avoid ambiguity.
    # Additionally, we set GH_REPO so any subsequent gh commands inherit the context.
    $env:GH_REPO = $repoName

    & gh secret set MPF_TERRAFORM_SPCLIENTID --body $terraformSp.ClientId -R $repoName | Out-Null
    & gh secret set MPF_TERRAFORM_SPCLIENTSECRET --body $terraformSp.ClientSecret -R $repoName | Out-Null
    & gh secret set MPF_TERRAFORM_SPOBJECTID --body $terraformSp.ObjectId -R $repoName | Out-Null

    & gh secret set MPF_ARM_BICEP_SPCLIENTID --body $armBicepSp.ClientId -R $repoName | Out-Null
    & gh secret set MPF_ARM_BICEP_SPCLIENTSECRET --body $armBicepSp.ClientSecret -R $repoName | Out-Null
    & gh secret set MPF_ARM_BICEP_SPOBJECTID --body $armBicepSp.ObjectId -R $repoName | Out-Null

    & gh secret set MPF_WINDOWS_SPCLIENTID --body $windowsSp.ClientId -R $repoName | Out-Null
    & gh secret set MPF_WINDOWS_SPCLIENTSECRET --body $windowsSp.ClientSecret -R $repoName | Out-Null
    & gh secret set MPF_WINDOWS_SPOBJECTID --body $windowsSp.ObjectId -R $repoName | Out-Null

    Ok 'All secrets added to GitHub repository!'
    Write-Host ""
    Info 'You can verify the secrets were added by running:'
    Write-Host 'gh secret list' -ForegroundColor Cyan
}
else {
    Info 'GitHub Repository Secrets Configuration'
    Write-Host ''
    Write-Host 'Terraform E2E Secrets:' -ForegroundColor Yellow
    Write-Host "MPF_TERRAFORM_SPCLIENTID = $($terraformSp.ClientId)"
    Write-Host "MPF_TERRAFORM_SPCLIENTSECRET = $($terraformSp.ClientSecret)"
    Write-Host "MPF_TERRAFORM_SPOBJECTID = $($terraformSp.ObjectId)"
    Write-Host ''
    Write-Host 'ARM/Bicep E2E Secrets:' -ForegroundColor Yellow
    Write-Host "MPF_ARM_BICEP_SPCLIENTID = $($armBicepSp.ClientId)"
    Write-Host "MPF_ARM_BICEP_SPCLIENTSECRET = $($armBicepSp.ClientSecret)"
    Write-Host "MPF_ARM_BICEP_SPOBJECTID = $($armBicepSp.ObjectId)"
    Write-Host ''
    Write-Host 'Windows E2E Secrets:' -ForegroundColor Yellow
    Write-Host "MPF_WINDOWS_SPCLIENTID = $($windowsSp.ClientId)"
    Write-Host "MPF_WINDOWS_SPCLIENTSECRET = $($windowsSp.ClientSecret)"
    Write-Host "MPF_WINDOWS_SPOBJECTID = $($windowsSp.ObjectId)"
    Write-Host ''
}

Write-Host "========================================" -ForegroundColor Green
Write-Host ' Setup Complete!' -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ''
Info 'Next steps:'
if ($shouldAddToGitHub) {
    Write-Host '1. ✅ Secrets have been added to GitHub repository'
    Write-Host '2. Test the updated workflows'
    Write-Host "3. Delete the credentials file: $CredentialsFile"
    Write-Host '4. Verify secrets with: gh secret list'
}
else {
    Write-Host '1. Add the secrets above to your GitHub repository'
    Write-Host '2. Test the updated workflows'
    Write-Host "3. Delete the credentials file: $CredentialsFile"
}
Write-Host ''
Warn 'Remember: These service principals have no permissions by default.'
Warn 'The MPF utility will manage all required permissions dynamically.'
