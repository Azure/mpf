# MPF Installation and Quickstart

## Installation

You can download the latest version for your platform from the [releases](https://github.com/Azure/mpf/releases) link.

For example, to download the latest version for Linux/amd64:

```shell
# Please change the version in the URL to the latest version
curl -LO https://github.com/Azure/mpf/releases/download/v0.16.0/azmpf_0.16.0_linux_amd64.zip
unzip azmpf_0.16.0_linux_amd64.zip
mv azmpf_v0.16.0 azmpf
chmod +x ./azmpf
```

For Mac Arm64:

```shell
# Please change the version in the URL to the latest version
curl -LO https://github.com/Azure/mpf/releases/download/v0.16.0/azmpf_0.16.0_darwin_arm64.zip
unzip azmpf_0.16.0_darwin_arm64.zip
mv azmpf_v0.16.0 azmpf
chmod +x ./azmpf
```

And for Windows:

```powershell
# Please change the version in the URL to the latest version
Invoke-WebRequest -Uri "https://github.com/Azure/mpf/releases/download/v0.16.0/azmpf_0.16.0_windows_amd64.zip" -OutFile "azmpf_0.16.0_windows_amd64.zip"
Expand-Archive -Path "azmpf_0.16.0_windows_amd64.zip" -DestinationPath "."
Rename-Item -Path "azmpf_v0.16.0.exe" -NewName "azmpf.exe"
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

It is also possible to additionally view detailed resource-level permissions required as shown in the [display options](./display-options.MD) document.

The blog post [Figuring out the Minimum Permissions Required to Deploy an Azure ARM Template](https://medium.com/microsoftazure/figuring-out-the-minimum-permissions-required-to-deploy-an-azure-arm-template-d1c1e74092fa) provides a more contextual usage scenario for azmpf.
