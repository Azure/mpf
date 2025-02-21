# MPF Installation and Quickstart

## Installation

You can download the latest version for your platform from the [releases](https://github.com/maniSbindra/az-mpf/releases/) link.

For example, to download the latest version for Windows:

```shell
# Please change the version in the URL to the latest version
curl -LO https://github.com/maniSbindra/az-mpf/releases/download/v0.11.0/az-mpf_0.11.0_windows_amd64.tar.gz
tar -xzf az-mpf_0.11.0_windows_amd64.tar.gz
mv az-mpf_0.11.0_windows_amd64 az-mpf.exe
chmod +x ./az-mpf.exe
```

And for Mac Arm64:
  
```shell
# Please change the version in the URL to the latest version
curl -LO https://github.com/maniSbindra/az-mpf/releases/download/v0.11.0/az-mpf_0.11.0_darwin_arm64.tar.gz
tar -xzf az-mpf_0.11.0_darwin_arm64.tar.gz
mv az-mpf-darwin-arm64 az-mpf
chmod +x ./az-mpf
```

## Quickstart / Usage

### ARM

```shell
export MPF_SUBSCRIPTIONID=YOUR_SUBSCRIPTION_ID
export MPF_TENANTID=YOUR_TENANT_ID
export MPF_SPCLIENTID=YOUR_SP_CLIENT_ID
export MPF_SPCLIENTSECRET=YOUR_SP_CLIENT_SECRET
export MPF_SPOBJECTID=YOUR_SP_OBJECT_ID

$ ./az-mpf arm --templateFilePath ./samples/templates/aks-private-subnet.json --parametersFilePath ./samples/templates/aks-private-subnet-parameters.json

------------------------------------------------------------------------------------------------------------------------------------------
Permissions Required:
------------------------------------------------------------------------------------------------------------------------------------------
Microsoft.ContainerService/managedClusters/read
Microsoft.ContainerService/managedClusters/write
Microsoft.Network/virtualNetworks/read
Microsoft.Network/virtualNetworks/subnets/read
Microsoft.Network/virtualNetworks/subnets/write
Microsoft.Network/virtualNetworks/write
Microsoft.Resources/deployments/read
Microsoft.Resources/deployments/write

```

### Bicep

```shell
export MPF_SUBSCRIPTIONID=YOUR_SUBSCRIPTION_ID
export MPF_TENANTID=YOUR_TENANT_ID
export MPF_SPCLIENTID=YOUR_SP_CLIENT_ID
export MPF_SPCLIENTSECRET=YOUR_SP_CLIENT_SECRET
export MPF_SPOBJECTID=YOUR_SP_OBJECT_ID
export MPF_BICEPEXECPATH="/opt/homebrew/bin/bicep" # Path to the Bicep executable

$ ./az-mpf bicep --bicepFilePath ./samples/bicep/aks-private-subnet.bicep --parametersFilePath ./samples/bicep/aks-private-subnet-params.json

------------------------------------------------------------------------------------------------------------------------------------------
Permissions Required:
------------------------------------------------------------------------------------------------------------------------------------------
Microsoft.ContainerService/managedClusters/read
Microsoft.ContainerService/managedClusters/write
Microsoft.Network/virtualNetworks/read
Microsoft.Network/virtualNetworks/subnets/read
Microsoft.Network/virtualNetworks/subnets/write
Microsoft.Network/virtualNetworks/write
Microsoft.Resources/deployments/read
Microsoft.Resources/deployments/write
------------------------------------------------------------------------------------------------------------------------------------------

```

### Terraform

```shell
export MPF_SUBSCRIPTIONID=YOUR_SUBSCRIPTION_ID
export MPF_TENANTID=YOUR_TENANT_ID
export MPF_SPCLIENTID=YOUR_SP_CLIENT_ID
export MPF_SPCLIENTSECRET=YOUR_SP_CLIENT_SECRET
export MPF_SPOBJECTID=YOUR_SP_OBJECT_ID
export MPF_TFPATH=TERRAFORM_EXECUTABLE_PATH

# pushd .
# cd ./samples/terraform/aci/
# $MPF_TFPATH init
# popd

$ ./az-mpf terraform --workingDir `pwd`/samples/terraform/aci --varFilePath `pwd`/samples/terraform/aci/dev.vars.tfvars --debug
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

It is also possible to additionally view detailed resource-level permissions required as shown in the [display options](docs/display-options.MD) document.

The blog post [Figuring out the Minimum Permissions Required to Deploy an Azure ARM Template](https://medium.com/microsoftazure/figuring-out-the-minimum-permissions-required-to-deploy-an-azure-arm-template-d1c1e74092fa) provides a more contextual usage scenario for az-mpf.