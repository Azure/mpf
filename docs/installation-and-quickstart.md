# MPF Installation and Quickstart

## Installation

You can download the latest version for your platform from the [releases](https://github.com/Azure/mpf/releases) link.

For example, to download the latest version for Linux/amd64:

```shell
# Please change the version in the URL to the latest version
curl -LO https://github.com/Azure/mpf/releases/download/v0.13.0/azmpf_0.13.0_linux_amd64.zip
unzip azmpf_0.13.0_linux_amd64.zip
mv azmpf_v0.13.0 azmpf
chmod +x ./azmpf
```

And for Mac Arm64:
  
```shell
# Please change the version in the URL to the latest version
curl -LO https://github.com/Azure/mpf/releases/download/v0.13.0/azmpf_0.13.0_darwin_arm64.zip
unzip azmpf_0.13.0_darwin_arm64.zip
mv azmpf_v0.13.0 azmpf
chmod +x ./azmpf
```

## Creating a service principal for MPF

To use MPF, you need to create a service principal in your Azure Active Directory tenant. You can create a service principal using the Azure CLI or the Azure portal. The service principal needs no roles assigned to it, as the MPF utility will as it is remove any assigned roles each time it executes.
Here is an example of how to create a service principal using the Azure CLI:

```shell
# az login

MPF_SP=$(az ad sp create-for-rbac --name "MPF_SP" --skip-assignment)
MPF_SPCLIENTID=$(echo $MPF_SP | jq -r .appId)
MPF_SPCLIENTSECRET=$(echo $MPF_SP | jq -r .password)
MPF_SPOBJECTID=$(az ad sp show --id $MPF_SPCLIENTID --query id -o tsv)
```

## Quickstart / Usage

### ARM

```shell
export MPF_SUBSCRIPTIONID=YOUR_SUBSCRIPTION_ID
export MPF_TENANTID=YOUR_TENANT_ID
export MPF_SPCLIENTID=YOUR_SP_CLIENT_ID
export MPF_SPCLIENTSECRET=YOUR_SP_CLIENT_SECRET
export MPF_SPOBJECTID=YOUR_SP_OBJECT_ID

$ ./azmpf arm --templateFilePath ./samples/templates/aks-private-subnet.json --parametersFilePath ./samples/templates/aks-private-subnet-parameters.json

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

$ ./azmpf bicep --bicepFilePath ./samples/bicep/aks-private-subnet.bicep --parametersFilePath ./samples/bicep/aks-private-subnet-params.json

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

$ ./azmpf terraform --workingDir `pwd`/samples/terraform/aci --varFilePath `pwd`/samples/terraform/aci/dev.vars.tfvars --debug
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
