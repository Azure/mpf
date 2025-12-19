Param(
    [string]$SubscriptionId = $env:SUBSCRIPTION_ID,
    [string]$RolePrefix = $(if ($env:TMP_ROLE_PREFIX) { $env:TMP_ROLE_PREFIX } else { 'tmp-rol-' })
)

$ErrorActionPreference = 'Stop'

if (-not (Get-Command az -ErrorAction SilentlyContinue)) {
    Write-Error 'Azure CLI (az) is required'
    exit 1
}

if ([string]::IsNullOrWhiteSpace($SubscriptionId)) {
    try {
        $SubscriptionId = az account show --query id -o tsv 2>$null
    }
    catch {
        $SubscriptionId = ''
    }
}

if ([string]::IsNullOrWhiteSpace($SubscriptionId)) {
    Write-Error 'Subscription not provided and no default found. Set SUBSCRIPTION_ID or login.'
    exit 1
}

az account set --subscription $SubscriptionId | Out-Null

$query = "[?starts_with(roleName, '$RolePrefix')].{Id:id,Name:name,RoleName:roleName}"
$defs = az role definition list --custom-role-only true --subscription $SubscriptionId --query $query -o json | ConvertFrom-Json

if (-not $defs -or $defs.Count -eq 0) {
    Write-Host 'No tmp-rol* custom roles found'
    exit 0
}

foreach ($def in $defs) {
    $roleId = $def.Id
    $roleName = $def.Name
    $roleDisplay = $def.RoleName

    Write-Host "Processing role $roleDisplay ($roleName)"

    $assignments = az role assignment list --all --subscription $SubscriptionId --role $roleId --query '[].id' -o tsv
    if ([string]::IsNullOrWhiteSpace($assignments)) {
        Write-Host "No assignments for $roleDisplay"
    }
    else {
        $assignments -split "`n" | Where-Object { -not [string]::IsNullOrWhiteSpace($_) } | ForEach-Object {
            Write-Host "Deleting assignment $_"
            az role assignment delete --ids $_
        }
    }

    Write-Host "Deleting role definition $roleName"
    az role definition delete --name $roleName --subscription $SubscriptionId
    Write-Host "Done $roleDisplay"
    Write-Host
}
