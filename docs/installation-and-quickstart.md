# MPF Installation and Quickstart

## Installation

You can download the latest version for your platform from the [releases](https://github.com/Azure/mpf/releases) link.

For example, to download the latest version for Linux/amd64:

```shell
# Please change the version in the URL to the latest version
curl -LO https://github.com/Azure/mpf/releases/download/v0.17.0/azmpf_linux_amd64.tar.gz
tar -xzf azmpf_linux_amd64.tar.gz
chmod +x ./azmpf
```

For Mac Arm64:

```shell
# Please change the version in the URL to the latest version
curl -LO https://github.com/Azure/mpf/releases/download/v0.17.0/azmpf_darwin_arm64.tar.gz
tar -xzf azmpf_darwin_arm64.tar.gz
chmod +x ./azmpf
```

And for Windows:

```powershell
# Please change the version in the URL to the latest version
Invoke-WebRequest -Uri "https://github.com/Azure/mpf/releases/download/v0.17.0/azmpf_windows_amd64.zip" -OutFile "azmpf_windows_amd64.zip"
Expand-Archive -Path "azmpf_windows_amd64.zip" -DestinationPath "."
```

## Creating a service principal for MPF

To use MPF, you need to create a service principal in your Azure Active Directory tenant. You can create a service principal using the Azure CLI or the Azure portal. The service principal needs no roles assigned to it, as the MPF utility will as it is remove any assigned roles each time it executes.

Here is an example of how to create a service principal using the Azure CLI on Linux/macOS:

```shell
# az login

MPF_SP=$(az ad sp create-for-rbac --name "MPF_SP" --skip-assignment)
MPF_SPCLIENTID=$(echo $MPF_SP | jq -r .appId)
MPF_SPCLIENTSECRET=$(echo $MPF_SP | jq -r .password)
MPF_SPOBJECTID=$(az ad sp show --id $MPF_SPCLIENTID --query id -o tsv)
```

And here is the equivalent example using Azure CLI on Windows PowerShell:

```powershell
# az login

$MPF_SP = az ad sp create-for-rbac --name "MPF_SP" --skip-assignment | ConvertFrom-Json
$env:MPF_SPCLIENTID = $MPF_SP.appId
$env:MPF_SPCLIENTSECRET = $MPF_SP.password
$env:MPF_SPOBJECTID = (az ad sp show --id $MPF_SP.appId --query id -o tsv)
```

## Quickstart / Usage

**Important**: ARM and Bicep deployments now use **Full Deployment mode** exclusively with Incremental deployment mode, which creates and deploys resources (then cleans them up automatically) to determine the required permissions. This provides the most accurate permission detection but takes longer than the previous what-if mode - expect execution times of several minutes to longer depending on template complexity and the resources being deployed. The previous what-if analysis mode (which completed in ~90 seconds) has been deprecated due to incomplete permission detection in some scenarios.

**Recommendation**: Always use the `--verbose` flag with ARM and Bicep commands to see progress during long-running deployments.

### ARM

```shell
export MPF_SUBSCRIPTIONID="YOUR_SUBSCRIPTION_ID"
export MPF_TENANTID="YOUR_TENANT_ID"
export MPF_SPCLIENTID="YOUR_SP_CLIENT_ID"
export MPF_SPCLIENTSECRET="YOUR_SP_CLIENT_SECRET"
export MPF_SPOBJECTID="YOUR_SP_OBJECT_ID"

$ ./azmpf arm --templateFilePath ./samples/templates/aks-private-subnet.json --parametersFilePath ./samples/templates/aks-private-subnet-parameters.json --verbose
```

Or using PowerShell on Windows:

```powershell
$env:MPF_SUBSCRIPTIONID = "YOUR_SUBSCRIPTION_ID"
$env:MPF_TENANTID = "YOUR_TENANT_ID"
$env:MPF_SPCLIENTID = "YOUR_SP_CLIENT_ID"
$env:MPF_SPCLIENTSECRET = "YOUR_SP_CLIENT_SECRET"
$env:MPF_SPOBJECTID = "YOUR_SP_OBJECT_ID"

.\azmpf.exe arm --templateFilePath .\samples\templates\aks-private-subnet.json --parametersFilePath .\samples\templates\aks-private-subnet-parameters.json --verbose
```

Output:

```text
INFO[0000] Executing MPF for ARM
INFO[0000] TemplateFilePath: ./samples/templates/aks-private-subnet.json
INFO[0000] ParametersFilePath: ./samples/templates/aks-private-subnet-parameters.json
INFO[0000] Location: eastus2
INFO[0001] Creating Resource Group: testdeployrg-OJ2zCNA
INFO[0007] Resource Group: testdeployrg-OJ2zCNA created successfully
INFO[0008] Deleted all existing role assignments for service principal
INFO[0008] Initializing Custom Role
INFO[0014] Custom role initialized successfully
INFO[0014] Assigning new custom role to service principal
INFO[0018] New Custom Role assigned to service principal successfully
INFO[0023] Iteration Number: 0
INFO[0023] Successfully Parsed Deployment Authorization Error
INFO[0023] Adding mising scopes/permissions to final result map...
INFO[0023] Adding permission/scope to role...........
INFO[0027] Permission/scope added to role successfully
INFO[0082] Iteration Number: 1
INFO[0082] Successfully Parsed Deployment Authorization Error
INFO[0082] Adding mising scopes/permissions to final result map...
INFO[0082] Adding permission/scope to role...........
INFO[0085] Permission/scope added to role successfully
INFO[0560] Iteration Number: 2
INFO[0560] Authorization Successful
INFO[0560] Cleaning up resources...
INFO[0560] *************************
INFO[0573] Role definition deleted successfully
INFO[0577] Resource group deletion initiated successfully...
------------------------------------------------------------------------------------------------------------------------------------------
Permissions Required:
------------------------------------------------------------------------------------------------------------------------------------------
Microsoft.ContainerService/managedClusters/read
Microsoft.ContainerService/managedClusters/write
Microsoft.Network/virtualNetworks/read
Microsoft.Network/virtualNetworks/subnets/join/action
Microsoft.Network/virtualNetworks/subnets/read
Microsoft.Network/virtualNetworks/subnets/write
Microsoft.Network/virtualNetworks/write
Microsoft.Resources/deployments/read
Microsoft.Resources/deployments/write
------------------------------------------------------------------------------------------------------------------------------------------

```

#### ARM with JSON Output

To get the output in JSON format (which includes per-resource permission details by default), use the `--jsonOutput` flag:

```shell
export MPF_SUBSCRIPTIONID="YOUR_SUBSCRIPTION_ID"
export MPF_TENANTID="YOUR_TENANT_ID"
export MPF_SPCLIENTID="YOUR_SP_CLIENT_ID"
export MPF_SPCLIENTSECRET="YOUR_SP_CLIENT_SECRET"
export MPF_SPOBJECTID="YOUR_SP_OBJECT_ID"

$ ./azmpf arm --templateFilePath ./samples/templates/aks-private-subnet.json --parametersFilePath ./samples/templates/aks-private-subnet-parameters.json --jsonOutput --verbose
```

Or using PowerShell on Windows:

```powershell
$env:MPF_SUBSCRIPTIONID = "YOUR_SUBSCRIPTION_ID"
$env:MPF_TENANTID = "YOUR_TENANT_ID"
$env:MPF_SPCLIENTID = "YOUR_SP_CLIENT_ID"
$env:MPF_SPCLIENTSECRET = "YOUR_SP_CLIENT_SECRET"
$env:MPF_SPOBJECTID = "YOUR_SP_OBJECT_ID"

.\azmpf.exe arm --templateFilePath .\samples\templates\aks-private-subnet.json --parametersFilePath .\samples\templates\aks-private-subnet-parameters.json --jsonOutput --verbose
```

Output (verbose INFO lines omitted for brevity):

```json
{
  "SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS": [
    "Microsoft.ContainerService/managedClusters/read",
    "Microsoft.ContainerService/managedClusters/write",
    "Microsoft.Network/virtualNetworks/read",
    "Microsoft.Network/virtualNetworks/subnets/join/action",
    "Microsoft.Network/virtualNetworks/subnets/read",
    "Microsoft.Network/virtualNetworks/subnets/write",
    "Microsoft.Network/virtualNetworks/write",
    "Microsoft.Resources/deployments/read",
    "Microsoft.Resources/deployments/write"
  ],
  "/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-OJ2zCNA/providers/Microsoft.ContainerService/managedClusters/azmpfakstestcluster": [
    "Microsoft.ContainerService/managedClusters/read",
    "Microsoft.ContainerService/managedClusters/write"
  ],
  "/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-OJ2zCNA/providers/Microsoft.Network/virtualNetworks/azmpfakstestvnet": [
    "Microsoft.Network/virtualNetworks/read",
    "Microsoft.Network/virtualNetworks/write"
  ],
  "/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-OJ2zCNA/providers/Microsoft.Network/virtualNetworks/azmpfakstestvnet/subnets/azmpfakstestsubnet": [
    "Microsoft.Network/virtualNetworks/subnets/join/action",
    "Microsoft.Network/virtualNetworks/subnets/read",
    "Microsoft.Network/virtualNetworks/subnets/write"
  ]
}
```

The JSON output is a map where the subscription ID key contains all aggregate permissions, and each resource scope key contains permissions specific to that resource. The `--showDetailedOutput` flag is not needed with `--jsonOutput` (they are mutually exclusive). For more display options, see [display options](display-options.MD).

#### ARM with Initial Permissions

The `--initialPermissions` flag allows you to seed known permissions before MPF starts its analysis. This can reduce execution time by avoiding extra permission-discovery iterations. It accepts either a comma-separated list or a JSON file reference (prefixed with `@`).

##### Comma-separated format

```shell
export MPF_SUBSCRIPTIONID="YOUR_SUBSCRIPTION_ID"
export MPF_TENANTID="YOUR_TENANT_ID"
export MPF_SPCLIENTID="YOUR_SP_CLIENT_ID"
export MPF_SPCLIENTSECRET="YOUR_SP_CLIENT_SECRET"
export MPF_SPOBJECTID="YOUR_SP_OBJECT_ID"

$ ./azmpf arm --templateFilePath ./samples/templates/aks-private-subnet.json --parametersFilePath ./samples/templates/aks-private-subnet-parameters.json \
  --initialPermissions "Microsoft.Network/virtualNetworks/read,Microsoft.Network/virtualNetworks/write,Microsoft.Network/virtualNetworks/subnets/read,Microsoft.Network/virtualNetworks/subnets/write" \
  --verbose
```

Or using PowerShell on Windows:

```powershell
$env:MPF_SUBSCRIPTIONID = "YOUR_SUBSCRIPTION_ID"
$env:MPF_TENANTID = "YOUR_TENANT_ID"
$env:MPF_SPCLIENTID = "YOUR_SP_CLIENT_ID"
$env:MPF_SPCLIENTSECRET = "YOUR_SP_CLIENT_SECRET"
$env:MPF_SPOBJECTID = "YOUR_SP_OBJECT_ID"

.\azmpf.exe arm --templateFilePath .\samples\templates\aks-private-subnet.json --parametersFilePath .\samples\templates\aks-private-subnet-parameters.json `
  --initialPermissions "Microsoft.Network/virtualNetworks/read,Microsoft.Network/virtualNetworks/write,Microsoft.Network/virtualNetworks/subnets/read,Microsoft.Network/virtualNetworks/subnets/write" `
  --verbose
```

##### JSON file format

For many permissions, a JSON file is cleaner. Create a file (e.g., `arm-initial-permissions.json`):

```json
{
  "RequiredPermissions": {
    "": [
      "Microsoft.Network/virtualNetworks/read",
      "Microsoft.Network/virtualNetworks/write",
      "Microsoft.Network/virtualNetworks/subnets/read",
      "Microsoft.Network/virtualNetworks/subnets/write",
      "Microsoft.Network/virtualNetworks/subnets/join/action"
    ]
  }
}
```

Then reference it with the `@` prefix:

```shell
export MPF_SUBSCRIPTIONID="YOUR_SUBSCRIPTION_ID"
export MPF_TENANTID="YOUR_TENANT_ID"
export MPF_SPCLIENTID="YOUR_SP_CLIENT_ID"
export MPF_SPCLIENTSECRET="YOUR_SP_CLIENT_SECRET"
export MPF_SPOBJECTID="YOUR_SP_OBJECT_ID"

$ ./azmpf arm --templateFilePath ./samples/templates/aks-private-subnet.json --parametersFilePath ./samples/templates/aks-private-subnet-parameters.json \
  --initialPermissions @arm-initial-permissions.json \
  --verbose
```

Or using PowerShell on Windows:

```powershell
$env:MPF_SUBSCRIPTIONID = "YOUR_SUBSCRIPTION_ID"
$env:MPF_TENANTID = "YOUR_TENANT_ID"
$env:MPF_SPCLIENTID = "YOUR_SP_CLIENT_ID"
$env:MPF_SPCLIENTSECRET = "YOUR_SP_CLIENT_SECRET"
$env:MPF_SPOBJECTID = "YOUR_SP_OBJECT_ID"

.\azmpf.exe arm --templateFilePath .\samples\templates\aks-private-subnet.json --parametersFilePath .\samples\templates\aks-private-subnet-parameters.json `
  --initialPermissions @arm-initial-permissions.json `
  --verbose
```

For full details on the `--initialPermissions` flag, see [Initial Permissions](commandline-flags-and-env-variables.md#initial-permissions).

### Bicep

```shell
export MPF_SUBSCRIPTIONID="YOUR_SUBSCRIPTION_ID"
export MPF_TENANTID="YOUR_TENANT_ID"
export MPF_SPCLIENTID="YOUR_SP_CLIENT_ID"
export MPF_SPCLIENTSECRET="YOUR_SP_CLIENT_SECRET"
export MPF_SPOBJECTID="YOUR_SP_OBJECT_ID"
export MPF_BICEPEXECPATH="/usr/local/bin/bicep" # Path to the Bicep executable

$ ./azmpf bicep --bicepFilePath ./samples/bicep/aks-private-subnet.bicep --parametersFilePath ./samples/bicep/aks-private-subnet-params.json --verbose
```

Or using PowerShell on Windows:

```powershell
$env:MPF_SUBSCRIPTIONID = "YOUR_SUBSCRIPTION_ID"
$env:MPF_TENANTID = "YOUR_TENANT_ID"
$env:MPF_SPCLIENTID = "YOUR_SP_CLIENT_ID"
$env:MPF_SPCLIENTSECRET = "YOUR_SP_CLIENT_SECRET"
$env:MPF_SPOBJECTID = "YOUR_SP_OBJECT_ID"
$env:MPF_BICEPEXECPATH = "C:\Program Files\Azure Bicep CLI\bicep.exe" # Path to the Bicep executable

.\azmpf.exe bicep --bicepFilePath .\samples\bicep\aks-private-subnet.bicep --parametersFilePath .\samples\bicep\aks-private-subnet-params.json --verbose
```

Output:

```text
INFO[0000] BicepFilePath: ./samples/bicep/aks-private-subnet.bicep
INFO[0000] ParametersFilePath: ./samples/bicep/aks-private-subnet-params.json
INFO[0000] Location: eastus2
INFO[0001] Successfully built ./samples/bicep/aks-private-subnet.bicep to ./samples/bicep/aks-private-subnet.json
INFO[0001] Creating Resource Group: testdeployrg-Qz5bP5s
INFO[0007] Resource Group: testdeployrg-Qz5bP5s created successfully
INFO[0007] Deleted all existing role assignments for service principal
INFO[0007] Initializing Custom Role
INFO[0014] Custom role initialized successfully
INFO[0014] Assigning new custom role to service principal
INFO[0018] New Custom Role assigned to service principal successfully
INFO[0024] Iteration Number: 0
INFO[0024] Successfully Parsed Deployment Authorization Error
INFO[0024] Adding mising scopes/permissions to final result map...
INFO[0024] Adding permission/scope to role...........
INFO[0028] Permission/scope added to role successfully
INFO[0175] Iteration Number: 1
INFO[0175] Successfully Parsed Deployment Authorization Error
INFO[0175] Adding mising scopes/permissions to final result map...
INFO[0175] Adding permission/scope to role...........
INFO[0178] Permission/scope added to role successfully
INFO[0675] Iteration Number: 2
INFO[0675] Authorization Successful
INFO[0675] Cleaning up resources...
INFO[0675] *************************
INFO[0688] Role definition deleted successfully
INFO[0693] Resource group deletion initiated successfully...
------------------------------------------------------------------------------------------------------------------------------------------
Permissions Required:
------------------------------------------------------------------------------------------------------------------------------------------
Microsoft.ContainerService/managedClusters/read
Microsoft.ContainerService/managedClusters/write
Microsoft.Network/virtualNetworks/read
Microsoft.Network/virtualNetworks/subnets/join/action
Microsoft.Network/virtualNetworks/subnets/read
Microsoft.Network/virtualNetworks/subnets/write
Microsoft.Network/virtualNetworks/write
Microsoft.Resources/deployments/read
Microsoft.Resources/deployments/write
------------------------------------------------------------------------------------------------------------------------------------------

```

### Bicep with JSON Output

You can also get JSON output from Bicep deployments, which is useful for programmatic processing and automation:

```bash
export MPF_SUBSCRIPTIONID="YOUR_SUBSCRIPTION_ID"
export MPF_TENANTID="YOUR_TENANT_ID"
export MPF_SPCLIENTID="YOUR_SP_CLIENT_ID"
export MPF_SPCLIENTSECRET="YOUR_SP_CLIENT_SECRET"
export MPF_SPOBJECTID="YOUR_SP_OBJECT_ID"
export MPF_BICEPEXECPATH=$(which bicep)

./azmpf bicep --bicepFilePath ./samples/bicep/storage-account-simple.bicep --parametersFilePath ./samples/bicep/storage-account-simple-params.json --jsonOutput --verbose
```

Or using PowerShell on Windows:

```powershell
$env:MPF_SUBSCRIPTIONID = "YOUR_SUBSCRIPTION_ID"
$env:MPF_TENANTID = "YOUR_TENANT_ID"
$env:MPF_SPCLIENTID = "YOUR_SP_CLIENT_ID"
$env:MPF_SPCLIENTSECRET = "YOUR_SP_CLIENT_SECRET"
$env:MPF_SPOBJECTID = "YOUR_SP_OBJECT_ID"
$env:MPF_BICEPEXECPATH = (Get-Command bicep).Source # Dynamically resolves to the Bicep executable path, works across different installation locations

.\azmpf.exe bicep --bicepFilePath .\samples\bicep\storage-account-simple.bicep --parametersFilePath .\samples\bicep\storage-account-simple-params.json --jsonOutput --verbose
```

### Terraform

```shell
export MPF_SUBSCRIPTIONID="YOUR_SUBSCRIPTION_ID"
export MPF_TENANTID="YOUR_TENANT_ID"
export MPF_SPCLIENTID="YOUR_SP_CLIENT_ID"
export MPF_SPCLIENTSECRET="YOUR_SP_CLIENT_SECRET"
export MPF_SPOBJECTID="YOUR_SP_OBJECT_ID"
export MPF_TFPATH="TERRAFORM_EXECUTABLE_PATH"

# pushd .
# cd ./samples/terraform/aci/
# $MPF_TFPATH init
# popd

$ ./azmpf terraform --workingDir `pwd`/samples/terraform/aci --varFilePath `pwd`/samples/terraform/aci/dev.vars.tfvars --debug
```

Or using PowerShell on Windows:

```powershell
$env:MPF_SUBSCRIPTIONID = "YOUR_SUBSCRIPTION_ID"
$env:MPF_TENANTID = "YOUR_TENANT_ID"
$env:MPF_SPCLIENTID = "YOUR_SP_CLIENT_ID"
$env:MPF_SPCLIENTSECRET = "YOUR_SP_CLIENT_SECRET"
$env:MPF_SPOBJECTID = "YOUR_SP_OBJECT_ID"
$env:MPF_TFPATH = "C:\Program Files\Terraform\terraform.exe" # Path to the Terraform executable

# Push-Location
# Set-Location .\samples\terraform\aci\
# & $env:MPF_TFPATH init
# Pop-Location

.\azmpf.exe terraform --workingDir "$PWD\samples\terraform\aci" --varFilePath "$PWD\samples\terraform\aci\dev.vars.tfvars" --debug
```

Output:

```text
.
# debug information
.
------------------------------------------------------------------------------------------------------------------------------------------
Permissions Required:
------------------------------------------------------------------------------------------------------------------------------------------
Microsoft.ContainerInstance/containerGroups/delete
Microsoft.ContainerInstance/containerGroups/read
Microsoft.ContainerInstance/containerGroups/write
Microsoft.Resources/deployments/read
Microsoft.Resources/deployments/write
Microsoft.Resources/subscriptions/resourcegroups/delete
Microsoft.Resources/subscriptions/resourcegroups/read
Microsoft.Resources/subscriptions/resourcegroups/write
------------------------------------------------------------------------------------------------------------------------------------------

```

#### Terraform with JSON Output

To get the output in JSON format, use the `--jsonOutput` flag:

```shell
export MPF_SUBSCRIPTIONID="YOUR_SUBSCRIPTION_ID"
export MPF_TENANTID="YOUR_TENANT_ID"
export MPF_SPCLIENTID="YOUR_SP_CLIENT_ID"
export MPF_SPCLIENTSECRET="YOUR_SP_CLIENT_SECRET"
export MPF_SPOBJECTID="YOUR_SP_OBJECT_ID"
export MPF_TFPATH="TERRAFORM_EXECUTABLE_PATH"

$ ./azmpf terraform --workingDir `pwd`/samples/terraform/aci --varFilePath `pwd`/samples/terraform/aci/dev.vars.tfvars --jsonOutput --verbose
```

Or using PowerShell on Windows:

```powershell
$env:MPF_SUBSCRIPTIONID = "YOUR_SUBSCRIPTION_ID"
$env:MPF_TENANTID = "YOUR_TENANT_ID"
$env:MPF_SPCLIENTID = "YOUR_SP_CLIENT_ID"
$env:MPF_SPCLIENTSECRET = "YOUR_SP_CLIENT_SECRET"
$env:MPF_SPOBJECTID = "YOUR_SP_OBJECT_ID"
$env:MPF_TFPATH = "C:\Program Files\Terraform\terraform.exe"

.\azmpf.exe terraform --workingDir "$PWD\samples\terraform\aci" --varFilePath "$PWD\samples\terraform\aci\dev.vars.tfvars" --jsonOutput --verbose
```

Output (verbose INFO lines omitted for brevity):

```json
{
  "SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS": [
    "Microsoft.ContainerInstance/containerGroups/delete",
    "Microsoft.ContainerInstance/containerGroups/read",
    "Microsoft.ContainerInstance/containerGroups/write",
    "Microsoft.Resources/deployments/read",
    "Microsoft.Resources/deployments/write",
    "Microsoft.Resources/subscriptions/resourcegroups/delete",
    "Microsoft.Resources/subscriptions/resourcegroups/read",
    "Microsoft.Resources/subscriptions/resourcegroups/write"
  ]
}
```

The JSON output is a map where the subscription ID key contains all aggregate permissions, and each resource scope key contains permissions specific to that resource. The `--showDetailedOutput` flag is not needed with `--jsonOutput` (they are mutually exclusive). For more display options, see [display options](display-options.MD).

#### Terraform with Module Targeting

When working with Terraform configurations that contain multiple modules, you can use the `--targetModule` flag to scope MPF analysis to a specific module. This is useful for large configurations where you only need permissions for a subset of resources.

The following example targets only the `module.law` module within the `module-test-with-targetting` sample:

```shell
export MPF_SUBSCRIPTIONID="YOUR_SUBSCRIPTION_ID"
export MPF_TENANTID="YOUR_TENANT_ID"
export MPF_SPCLIENTID="YOUR_SP_CLIENT_ID"
export MPF_SPCLIENTSECRET="YOUR_SP_CLIENT_SECRET"
export MPF_SPOBJECTID="YOUR_SP_OBJECT_ID"
export MPF_TFPATH="TERRAFORM_EXECUTABLE_PATH"

# Ensure terraform is initialized first
# pushd ./samples/terraform/module-test-with-targetting && $MPF_TFPATH init && popd

$ ./azmpf terraform --workingDir `pwd`/samples/terraform/module-test-with-targetting --varFilePath `pwd`/samples/terraform/module-test-with-targetting/terraform.tfvars --targetModule "module.law" --verbose
```

Or using PowerShell on Windows:

```powershell
$env:MPF_SUBSCRIPTIONID = "YOUR_SUBSCRIPTION_ID"
$env:MPF_TENANTID = "YOUR_TENANT_ID"
$env:MPF_SPCLIENTID = "YOUR_SP_CLIENT_ID"
$env:MPF_SPCLIENTSECRET = "YOUR_SP_CLIENT_SECRET"
$env:MPF_SPOBJECTID = "YOUR_SP_OBJECT_ID"
$env:MPF_TFPATH = "C:\Program Files\Terraform\terraform.exe"

# Ensure terraform is initialized first
# Push-Location .\samples\terraform\module-test-with-targetting\; & $env:MPF_TFPATH init; Pop-Location

.\azmpf.exe terraform --workingDir "$PWD\samples\terraform\module-test-with-targetting" --varFilePath "$PWD\samples\terraform\module-test-with-targetting\terraform.tfvars" --targetModule "module.law" --verbose
```

> **Note**: The `--targetModule` value uses Terraform's module address syntax (e.g., `module.law`). Only the targeted module's resources will be deployed, and only the permissions required for that module will be reported.

#### Terraform with Initial Permissions

The `--initialPermissions` flag is especially useful for Terraform when using a **remote backend** (e.g., Azure Storage for state). MPF removes all existing role assignments from the service principal before analysis, which can cause Terraform to fail with a 403 error when accessing the backend state store (see [#172](https://github.com/Azure/mpf/issues/172)).

##### Using a JSON file for remote backend permissions

Create a file called `backend-permissions.json` (a sample is provided at `samples/terraform/backend-permissions.json`):

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

Then reference it with the `@` prefix:

```shell
export MPF_SUBSCRIPTIONID="YOUR_SUBSCRIPTION_ID"
export MPF_TENANTID="YOUR_TENANT_ID"
export MPF_SPCLIENTID="YOUR_SP_CLIENT_ID"
export MPF_SPCLIENTSECRET="YOUR_SP_CLIENT_SECRET"
export MPF_SPOBJECTID="YOUR_SP_OBJECT_ID"
export MPF_TFPATH=$(which terraform)

$ ./azmpf terraform --workingDir `pwd`/samples/terraform/aci --varFilePath `pwd`/samples/terraform/aci/dev.vars.tfvars \
  --initialPermissions @backend-permissions.json \
  --verbose
```

Or using PowerShell on Windows:

```powershell
$env:MPF_SUBSCRIPTIONID = "YOUR_SUBSCRIPTION_ID"
$env:MPF_TENANTID = "YOUR_TENANT_ID"
$env:MPF_SPCLIENTID = "YOUR_SP_CLIENT_ID"
$env:MPF_SPCLIENTSECRET = "YOUR_SP_CLIENT_SECRET"
$env:MPF_SPOBJECTID = "YOUR_SP_OBJECT_ID"
$env:MPF_TFPATH = "C:\Program Files\Terraform\terraform.exe"

.\azmpf.exe terraform --workingDir "$PWD\samples\terraform\aci" --varFilePath "$PWD\samples\terraform\aci\dev.vars.tfvars" `
  --initialPermissions @backend-permissions.json `
  --verbose
```

##### Using comma-separated permissions

For fewer permissions, you can specify them inline:

```shell
$ ./azmpf terraform --workingDir `pwd`/samples/terraform/aci --varFilePath `pwd`/samples/terraform/aci/dev.vars.tfvars \
  --initialPermissions "Microsoft.Storage/storageAccounts/read,Microsoft.Storage/storageAccounts/listKeys/action" \
  --verbose
```

Or using PowerShell on Windows:

```powershell
.\azmpf.exe terraform --workingDir "$PWD\samples\terraform\aci" --varFilePath "$PWD\samples\terraform\aci\dev.vars.tfvars" `
  --initialPermissions "Microsoft.Storage/storageAccounts/read,Microsoft.Storage/storageAccounts/listKeys/action" `
  --verbose
```

For full details on the `--initialPermissions` flag, see [Initial Permissions](commandline-flags-and-env-variables.md#initial-permissions). For more context on the remote backend issue, see [Known Issues - Remote Backend Access Denied](known-issues-and-workarounds.MD#remote-backend-access-denied).

It is also possible to additionally view detailed resource-level permissions required as shown in the [display options](./display-options.MD) document.

The blog post [Figuring out the Minimum Permissions Required to Deploy an Azure ARM Template](https://medium.com/microsoftazure/figuring-out-the-minimum-permissions-required-to-deploy-an-azure-arm-template-d1c1e74092fa) provides a more contextual usage scenario for azmpf.
