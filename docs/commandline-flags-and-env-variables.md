# MPF command line flags and environment variables

**Note**: Environment variables can be set using bash/shell syntax (e.g., `export MPF_SUBSCRIPTIONID=value`) on Linux/macOS, or using PowerShell syntax (e.g., `$env:MPF_SUBSCRIPTIONID = "value"`) on Windows.

## Global Flags (Common to all providers)

| Flag               | Environment Variable   | Required / Optional | Description                                                                                                                       |
|--------------------|------------------------|---------------------|-----------------------------------------------------------------------------------------------------------------------------------|
| subscriptionID     | MPF_SUBSCRIPTIONID     | Required            |                                                                                                                                   |
| tenantID           | MPF_TENANTID           | Required            |                                                                                                                                   |
| spClientID         | MPF_SPCLIENTID         | Required            |                                                                                                                                   |
| spObjectID         | MPF_SPOBJECTID         | Required            | Note this is the SP Object id and is different from the Client ID                                                                 |
| spClientSecret     | MPF_SPCLIENTSECRET     | Required            |                                                                                                                                   |
| showDetailedOutput | MPF_SHOWDETAILEDOUTPUT | Optional            | If set to true, the output shows details of permissions resource wise as well. This is not needed if --jsonOutput is specified    |
| jsonOutput         | MPF_JSONOUTPUT         | Optional            | If set to true, the detailed output is printed in JSON format                                                                     |
| verbose            | MPF_VERBOSE            | Optional            | If set to true, verbose output with informational messages is displayed                                                           |
| debug              | MPF_DEBUG              | Optional            | If set to true, output with detailed debug messages is displayed. The debug messages may contain sensitive tokens                 |
| initialPermissions | MPF_INITIALPERMISSIONS | Optional            | Initial permissions to seed the custom role with before MPF analysis. See [Initial Permissions](#initial-permissions) for details |

When used for Terraform, the verbose and debug flags show detailed logs from Terraform.

## ARM Flags

| Flag                 | Environment Variable     | Required / Optional | Description                                                                                                                                       |
|----------------------|--------------------------|---------------------|---------------------------------------------------------------------------------------------------------------------------------------------------|
| templateFilePath     | MPF_TEMPLATEFILEPATH     | Required            | ARM template file with path                                                                                                                       |
| parametersFilePath   | MPF_PARAMETERSFILEPATH   | Required            | ARM template parameters file with path                                                                                                            |
| resourceGroupNamePfx | MPF_RESOURCEGROUPNAMEPFX | Optional            | Prefix for the resource group name. If not provided, default prefix is testdeployrg. For ARM deployments this temporary resource group is created |
| deploymentNamePfx    | MPF_DEPLOYMENTNAMEPFX    | Optional            | Prefix for the deployment name. If not provided, default prefix is testDeploy. For ARM deployments this temporary deployment is created           |
| location             | MPF_LOCATION             | Optional            | Location for the resource group. If not provided, default location is eastus2                                                                     |

### Bicep Flags

| Flag                 | Environment Variable     | Required / Optional | Description                                                                                                                                         |
|----------------------|--------------------------|---------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| bicepFilePath        | MPF_BICEPFILEPATH        | Required            | Bicep file with path                                                                                                                                |
| parametersFilePath   | MPF_PARAMETERSFILEPATH   | Required            | Bicep parameters file with path                                                                                                                     |
| bicepExecPath        | MPF_BICEPEXECPATH        | Required            | Path to the Bicep executable                                                                                                                        |
| resourceGroupNamePfx | MPF_RESOURCEGROUPNAMEPFX | Optional            | Prefix for the resource group name. If not provided, default prefix is testdeployrg. For Bicep deployments this temporary resource group is created |
| deploymentNamePfx    | MPF_DEPLOYMENTNAMEPFX    | Optional            | Prefix for the deployment name. If not provided, default prefix is testDeploy. For Bicep deployments this temporary deployment is created           |
| location             | MPF_LOCATION             | Optional            | Location for the resource group. If not provided, default location is eastus2                                                                       |

## Terraform Flags

| Flag                           | Environment Variable               | Required / Optional | Description                                                                                                                                                                       |
|--------------------------------|------------------------------------|---------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| tfPath                         | MPF_TFPATH                         | Required            | Path to the Terraform executable                                                                                                                                                  |
| workingDir                     | MPF_WORKINGDIR                     | Required            | Path to the Terraform module directory                                                                                                                                            |
| varFilePath                    | MPF_VARFILEPATH                    | Optional            | Path to the Terraform variables file                                                                                                                                              |
| importExistingResourcesToState | MPF_IMPORTEXISTINGRESOURCESTOSTATE | Optional            | Default Value is true. This is required for some scenarios as described in the [Known Issues - Import Errors](./known-issues-and-workarounds.MD#existing-resource--import-errors) |
| targetModule                   | MPF_TARGETMODULE                   | Optional            | Target module to be used for the Terraform deployment                                                                                                                             |

## Initial Permissions

The `--initialPermissions` flag allows you to specify permissions that should be added to the custom role before MPF starts its analysis. This is particularly useful when:

- Using **Terraform with a remote backend** (e.g., Azure Storage) that requires permissions to access the state store
- You want to **reduce MPF execution time** by seeding known permissions upfront
- Your deployment has **prerequisites** that need specific permissions before the main deployment can proceed

### Usage

The flag accepts two formats:

#### 1. Comma-separated list

```bash
azmpf terraform \
  --initialPermissions "Microsoft.Storage/storageAccounts/read,Microsoft.Storage/storageAccounts/listKeys/action" \
  --workingDir ./my-terraform \
  # ... other flags
```

#### 2. JSON file reference (prefix with @)

```bash
azmpf terraform \
  --initialPermissions @backend-permissions.json \
  --workingDir ./my-terraform \
  # ... other flags
```

The JSON file must have the following format:

```json
{
  "RequiredPermissions": {
    "": [
      "Microsoft.Storage/storageAccounts/read",
      "Microsoft.Storage/storageAccounts/listKeys/action",
      "Microsoft.Storage/storageAccounts/blobServices/containers/read",
      "Microsoft.Storage/storageAccounts/blobServices/containers/blobs/read",
      "Microsoft.Storage/storageAccounts/blobServices/containers/blobs/write"
    ]
  }
}
```

### Example: Terraform Remote Backend

When using Azure Storage as a Terraform remote backend, the service principal needs permissions to access the storage account. Create a file called `backend-permissions.json`:

```json
{
  "RequiredPermissions": {
    "": [
      "Microsoft.Storage/storageAccounts/read",
      "Microsoft.Storage/storageAccounts/listKeys/action",
      "Microsoft.Storage/storageAccounts/blobServices/containers/read",
      "Microsoft.Storage/storageAccounts/blobServices/containers/blobs/read",
      "Microsoft.Storage/storageAccounts/blobServices/containers/blobs/write"
    ]
  }
}
```

Then run MPF with:

```bash
azmpf terraform \
  --initialPermissions @backend-permissions.json \
  --tfPath $(which terraform) \
  --workingDir ./my-terraform \
  # ... other required flags
```

### Example: Bicep with Pre-existing Storage Backend (Comma-separated)

When deploying Bicep templates that depend on pre-existing Azure Storage (for configuration, state, or secrets), you can use comma-separated permissions to speed up analysis:

```bash
export MPF_SUBSCRIPTIONID="YOUR_SUBSCRIPTION_ID"
export MPF_TENANTID="YOUR_TENANT_ID"
export MPF_SPCLIENTID="YOUR_SP_CLIENT_ID"
export MPF_SPCLIENTSECRET="YOUR_SP_CLIENT_SECRET"
export MPF_SPOBJECTID="YOUR_SP_OBJECT_ID"
export MPF_BICEPEXECPATH=$(which bicep)

azmpf bicep \
  --initialPermissions "Microsoft.Storage/storageAccounts/read,Microsoft.Storage/storageAccounts/listKeys/action,Microsoft.Storage/storageAccounts/blobServices/containers/read" \
  --bicepFilePath ./samples/bicep/aks-private-subnet.bicep \
  --parametersFilePath ./samples/bicep/aks-private-subnet-params.json \
  --verbose
```

Or using PowerShell on Windows:

```powershell
$env:MPF_SUBSCRIPTIONID = "YOUR_SUBSCRIPTION_ID"
$env:MPF_TENANTID = "YOUR_TENANT_ID"
$env:MPF_SPCLIENTID = "YOUR_SP_CLIENT_ID"
$env:MPF_SPCLIENTSECRET = "YOUR_SP_CLIENT_SECRET"
$env:MPF_SPOBJECTID = "YOUR_SP_OBJECT_ID"
$env:MPF_BICEPEXECPATH = "C:\Program Files\Azure Bicep CLI\bicep.exe"

.\azmpf.exe bicep `
  --initialPermissions "Microsoft.Storage/storageAccounts/read,Microsoft.Storage/storageAccounts/listKeys/action,Microsoft.Storage/storageAccounts/blobServices/containers/read" `
  --bicepFilePath .\samples\bicep\aks-private-subnet.bicep `
  --parametersFilePath .\samples\bicep\aks-private-subnet-params.json `
  --verbose
```

### Example: Bicep with JSON File Format

For complex deployments with many pre-requisite permissions, using a JSON file is cleaner and more maintainable. Create a file called `bicep-backend-permissions.json`:

```json
{
  "RequiredPermissions": {
    "": [
      "Microsoft.Storage/storageAccounts/read",
      "Microsoft.Storage/storageAccounts/listKeys/action",
      "Microsoft.Storage/storageAccounts/blobServices/containers/read",
      "Microsoft.Storage/storageAccounts/blobServices/containers/blobs/read",
      "Microsoft.Storage/storageAccounts/blobServices/containers/blobs/write"
    ]
  }
}
```

Then run MPF with:

```bash
export MPF_SUBSCRIPTIONID="YOUR_SUBSCRIPTION_ID"
export MPF_TENANTID="YOUR_TENANT_ID"
export MPF_SPCLIENTID="YOUR_SP_CLIENT_ID"
export MPF_SPCLIENTSECRET="YOUR_SP_CLIENT_SECRET"
export MPF_SPOBJECTID="YOUR_SP_OBJECT_ID"
export MPF_BICEPEXECPATH=$(which bicep)

azmpf bicep \
  --initialPermissions @bicep-backend-permissions.json \
  --bicepFilePath ./samples/bicep/aks-private-subnet.bicep \
  --parametersFilePath ./samples/bicep/aks-private-subnet-params.json \
  --verbose
```

Or using PowerShell on Windows:

```powershell
$env:MPF_SUBSCRIPTIONID = "YOUR_SUBSCRIPTION_ID"
$env:MPF_TENANTID = "YOUR_TENANT_ID"
$env:MPF_SPCLIENTID = "YOUR_SP_CLIENT_ID"
$env:MPF_SPCLIENTSECRET = "YOUR_SP_CLIENT_SECRET"
$env:MPF_SPOBJECTID = "YOUR_SP_OBJECT_ID"
$env:MPF_BICEPEXECPATH = "C:\Program Files\Azure Bicep CLI\bicep.exe"

.\azmpf.exe bicep `
  --initialPermissions @bicep-backend-permissions.json `
  --bicepFilePath .\samples\bicep\aks-private-subnet.bicep `
  --parametersFilePath .\samples\bicep\aks-private-subnet-params.json `
  --verbose
```

**Benefits of using `--initialPermissions`:**

- Reduces MPF execution time by avoiding permission discovery iterations for known prerequisites
- Cleaner separation of concerns: pre-requisite permissions vs. deployment permissions
- JSON file format provides a reusable, version-controllable configuration
