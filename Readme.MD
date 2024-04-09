# az-mpf utility (Azure Deployment Minimum Permissions Finder)

This utility finds the minimum permissions required for a given Azure deployment. This can help when you need to figure out the details of what permissions a service principal or managed identity will need to deploy a given ARM template, bicep file or Terrform module. Similarly when assigning a Service Principal / Managed Identity to an Azure Policy Assignment, this utility can help you figure out the minimum permissions required by the Service Principal / Managed Identity to enforce/remediate the policy. It is recommended that the utility is used in a development or test environment to find the minimum permissions required.

## How It works

The overview of how this utility works is as follows:

![Overview](docs/images/overview.png)

Following is the detailed flow of how this utility works:
* The key parameters the utility needs are the **Service Principal details** (Client ID, Secret and Object ID) and details needed for the specific deployment provider: 
  * ARM: ARM template file and parameters file needed
  * Terraform: Terraform module directory and variables file needed
* The utility **removes any existing Role Assignments for provided Service Principal**
* A Custom Role (with no assigned permissions) is created
* The Service Principal (SP) is assigned the new custom role
* For the above steps the utitity uses the **default Azure CLI credentials** which needs to have permissions to create custom role, and role assignments. The details of the permissions required by the default Azure CLI credentials are provided in the [Permissions required by default Azure CLI credentials](#permissions-required-by-default-azure-cli-credentials) section.
* For the following sub steps the **Service Principal Credentials** are used. These sub steps are re-tried till the deployment succeeds
  * Depending on the provider (ARM, Bicep, or Terraform) a deployment is tried
  * If the Service Principal does not have sufficient permissions an authorization error is returned by the deployment. If Authorization errors have occured, they are parsed to fetch the missing scopes and permissions. The [authorizationErrorParser Tests](./pkg/domain/authorizationErrorParser_test.go) provides details of the different kinds of Authorization errors typically received.
  * The missing permissions are added to the custom Role. This sub step uses **default Azure CLI credentials** to update role permissions.
* Once no authorization error is received, the utility prints the permissions assigned to the Service Principal
* The required permissions are displayed based on the display options. These options can be used to view the resource wise break up of permissions and also to export the result in JSON format
* All resources created are cleaned up by the utility including the Role Assignments and Custom Role.

## Supported Deployment Providers

* Azure **ARM** Template: The ARM endpoints of both modes described below return multiple authorization errors at a time and as a result the final output is usually displayed in under 90 seconds, even for complex templates.
  * Whatif mode: This this the default ARM mode. It uses the ARM whatif endpoint to get the authorization errors and find the minimum permissions required for a deployment, without actually create any resources.
  * FullDeployment mode: This mode has been deprecated and is no longer supported.
* **Bicep**: The Bicep mode is similar to the ARM whatif mode, and uses the ARM whatif endpoint to get the authorization errors and find the minimum permissions required for a deployment, without actually create any resources. Internally the utility converts the bicep file to ARM template and then uses the ARM whatif endpoint.
* **Terraform**: The Terraform mode finds the minimum permissions required for a deployment, by getting the authorization errors from the Terraform plan/apply, and destroy commands. All resources are cleaned up by the utility. Since terraform calls the ARM apis for one resource at a time, the authorization errors are not received in bulk, and as a result it can take quite long to get the final result. The overall time is the time taken to run the terraform plan/apply/destroy commands, plus the overhead of getting and parsing the authorization errors a few times.

## Flags and Environment Variables

The commands can be used with flags or environment variables. The following are the flags and environment variables supported by the utility:

### Global Flags (Common to all providers)

| Flag               | Environment Variable | Required / Optional | Description                                                                                                      |
| ------------------ | -------------------- | ------------------ | ---------------------------------------------------------------------------------------------------------------- |
| subscriptionID     | MPF_SUBSCRIPTIONID   | Required           |                                                                                                                  |
| tenantID           | MPF_TENANTID         | Required           |                                                                                                                  |
| spClientID         | MPF_SPCLIENTID       | Required           |                                                                                                                  |
| spObjectID         | MPF_SPOBJECTID       | Required           | Note this is the SP Object id and is different from the Client ID                                                |
| spClientSecret     | MPF_SPCLIENTSECRET   | Required           |                                                                                                                  |
| showDetailedOutput | MPF_SHOWDETAILEDOUTPUT | Optional       | If set to true, the output shows details of permissions resource wise as well. This is not needed if --jsonOutput is specified |
| jsonOutput         | MPF_JSONOUTPUT       | Optional           | If set to true, the detailed output is printed in JSON format                                                            |
| verbose            | MPF_VERBOSE          | Optional           | If set to true, verbose output with informational messages is displayed                                          |
| debug              | MPF_DEBUG            | Optional           | If set to true, output with detailed debug messages is displayed. The debug messages may contain sensitive tokens |

### ARM Flags

| Flag                 | Environment Variable     | Required / Optional | Description                                                                                                      |
| -------------------- | ------------------------ | ------------------ | ---------------------------------------------------------------------------------------------------------------- |
| templateFilePath     | MPF_TEMPLATEFILEPATH     | Required           | ARM template file with path                                                                                       |
| parametersFilePath   | MPF_PARAMETERSFILEPATH   | Required           | ARM template parameters file with path                                                                            |
| resourceGroupNamePfx | MPF_RESOURCEGROUPNAMEPFX | Optional           | Prefix for the resource group name. If not provided, default prefix is testdeployrg. For ARM deployments this temporary resource group is created                                |
| deploymentNamePfx    | MPF_DEPLOYMENTNAMEPFX    | Optional           | Prefix for the deployment name. If not provided, default prefix is testDeploy. For ARM deployments this temporary deployment is created                                |
| location             | MPF_LOCATION             | Optional           | Location for the resource group. If not provided, default location is eastus                                      |

### Bicep Flags

| Flag                 | Environment Variable     | Required / Optional | Description                                                                                                      |
| -------------------- | ------------------------ | ------------------ | ---------------------------------------------------------------------------------------------------------------- |
| bicepFilePath        | MPF_BICEPFILEPATH        | Required           | Bicep file with path                                                                                              |
| parametersFilePath   | MPF_PARAMETERSFILEPATH   | Required           | Bicep parameters file with path                                                                                    |
| bicepExecPath        | MPF_BICEPEXECPATH        | Required           | Path to the Bicep executable                                                                                       |
| resourceGroupNamePfx | MPF_RESOURCEGROUPNAMEPFX | Optional           | Prefix for the resource group name. If not provided, default prefix is testdeployrg. For Bicep deployments this temporary resource group is created                                |
| deploymentNamePfx    | MPF_DEPLOYMENTNAMEPFX    | Optional           | Prefix for the deployment name. If not provided, default prefix is testDeploy. For Bicep deployments this temporary deployment is created                                |
| location             | MPF_LOCATION             | Optional           | Location for the resource group. If not provided, default location is eastus                                      |

### Terraform Flags

| Flag                 | Environment Variable     | Required / Optional | Description                                                                                                      |
| -------------------- | ------------------------ | ------------------ | ---------------------------------------------------------------------------------------------------------------- |
| tfPath               | MPF_TFPATH               | Required           | Path to the Terraform executable                                                                             |
| workingDir           | MPF_WORKINGDIR           | Required           | Path to the Terraform module directory                                                                             |
| varFilePath          | MPF_VARFILEPATH          | Optional           | Path to the Terraform variables file                                                                             |
| importExistingResourcesToState | MPF_IMPORTEXISTINGRESOURCESTOSTATE | Optional | Default Value is true. This is required for some scenarios as described in the [Known Issues - Import Errors](./docs/known-issues-and-workarounds.MD#existing-resource--import-errors) |
| targetModule        | MPF_TARGETMODULE         | Optional           | Target module to be used for the Terraform deployment |                                                                             |


## Installation

You can download the latest version for your platform from the [releases](https://github.com/maniSbindra/az-mpf/releases/) link.

For example, to download the latest version for Windows:

```shell
# Please change version in the URL to the latest version
curl -LO https://github.com/maniSbindra/az-mpf/releases/download/v0.11.0/az-mpf_0.11.0_windows_amd64.tar.gz
tar -xzf az-mpf_0.11.0_windows_amd64.tar.gz
mv az-mpf_0.11.0_windows_amd64 az-mpf.exe
chmod +x ./az-mpf.exe
```

And for Mac Arm64:
  
```shell
# Please change version in the URL to the latest version
curl -LO https://github.com/maniSbindra/az-mpf/releases/download/v0.11.0/az-mpf_0.11.0_darwin_arm64.tar.gz
tar -xzf az-mpf_0.11.0_darwin_arm64.tar.gz
mv az-mpf-darwin-arm64 az-mpf
chmod +x ./az-mpf
```

## Verify Release Binaries using SLSA Verifier

The release binaries are signed using the SLSA (Supply Chain Levels for Software Artifacts) framework. You can verify the release binaries using the [SLSA Verifier](https://github.com/slsa-framework/slsa-verifier). The following steps show how to verify the release binary for darwin/arm64 against release 0.11.0, using the SLSA Verifier:

Note:! There was an issue with provenance generation of this release. This will be addressed as a part of [issue #52](https://github.com/maniSbindra/az-mpf/issues/52)

```shell
# download arm64 darwin release binary and the multiple.intoto.jsonl file
curl -LO https://github.com/maniSbindra/az-mpf/releases/download/v0.11.0/az-mpf_0.11.0_darwin_arm64.tar.gz
curl -LO https://github.com/maniSbindra/az-mpf/releases/download/v0.11.0/multiple.intoto.jsonl

# modify path to the slsa-verifier binary as per your installation and verify the release binary
$ /PATH_TO_VERIFER/slsa-verifier verify-artifact az-mpf_0.11.0_darwin_arm64.tar.gz   --provenance-path multiple.intoto.jsonl   --source-uri github.com/maniSbindra/az-mpf  --source-tag v0.11.0

Verified signature against tlog entry index 76002620 at URL: https://rekor.sigstore.dev/api/v1/log/entries/24296fb24b8ad77ae67470bb23a7a809a475c46078739a61725936641b76508ac420e02e217475de
Verified build using builder "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@refs/tags/v1.9.0" at commit 648fe797c5b5253360ae1c6e60bd9521544209bc
Verifying artifact az-mpf_0.11.0_darwin_arm64.tar.gz: PASSED

PASSED: Verified SLSA provenance
```


## Usage Details

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

Is is also possible to additionally view detailed resource level permissions required as shown in the [display options](docs/display-options.MD) documents.

The blog post [Figuring out the Minimum Permissions Required to Deploy an Azure ARM Template](https://medium.com/microsoftazure/figuring-out-the-minimum-permissions-required-to-deploy-an-azure-arm-template-d1c1e74092fa) provides a more contextual usage scenario for az-mpf.


## Display Options

To view details of display options the utility provides please refer to the [display options](docs/display-options.MD) document.

## Building Locally

You can also build locally by cloning this repo and running `make build`.

## Testing Locally

### Unit Tests

To run the unit tests, run `make test`.

### End to End ARM and Bicep Tests

To run the end to end tests for ARM and Bicep, you need to have the following environment variables set, and then execute `make test-e2e`:

```shell
export MPF_SUBSCRIPTIONID=YOUR_SUBSCRIPTION_ID
export MPF_TENANTID=YOUR_TENANT_ID
export MPF_SPCLIENTID=YOUR_SP_CLIENT_ID
export MPF_SPCLIENTSECRET=YOUR_SP_CLIENT_SECRET
export MPF_SPOBJECTID=YOUR_SP_OBJECT_ID
export MPF_BICEPEXECPATH="/opt/homebrew/bin/bicep" # Path to the Bicep executable

make test-e2e
```

### End to End Terraform Tests

The Terraform end to end tests, can take a long time to execute, depending on the resources being created. To run the end to end tests for Terraform, you need to have the following environment variables set, and then execute `make test-e2e-terraform`:

```shell
export MPF_SUBSCRIPTIONID=YOUR_SUBSCRIPTION_ID
export MPF_TENANTID=YOUR_TENANT_ID
export MPF_SPCLIENTID=YOUR_SP_CLIENT_ID
export MPF_SPCLIENTSECRET=YOUR_SP_CLIENT_SECRET
export MPF_SPOBJECTID=YOUR_SP_OBJECT_ID
export MPF_TFPATH=$(which terraform) # Path to the Terraform executable

make test-e2e-terraform
```

## Permissions required by default Azure CLI credentials

The default Azure CLI credentials used by the utility need to have the following permissions:

* "Microsoft.Authorization/roleDefinitions/read"
* "Microsoft.Authorization/roleDefinitions/write"
* "Microsoft.Authorization/roleDefinitions/delete"
* "Microsoft.Authorization/roleAssignments/read"
* "Microsoft.Authorization/roleAssignments/write"
* "Microsoft.Authorization/roleAssignments/delete"
* "Microsoft.Resources/subscriptions/resourcegroups/delete"
* "Microsoft.Resources/subscriptions/resourcegroups/read"
* "Microsoft.Resources/subscriptions/resourcegroups/write"

## Known Issues and Workarounds

The [Known Issues and Workarounds](docs/known-issues-and-workarounds.MD) document provides details on the known issues and workarounds for the utility.

## Debugging Locally and Contributing

The [debugging locally](./docs/debugging-locally.MD) file provides details on how to debug locally.



## License

This project is under an [MIT License](LICENSE).

## Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

## Trademarks

This project may contain trademarks or logos for projects, products, or services. Authorized use of Microsoft 
trademarks or logos is subject to and must follow 
[Microsoft's Trademark & Brand Guidelines](https://www.microsoft.com/en-us/legal/intellectualproperty/trademarks/usage/general).
Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship.
Any use of third-party trademarks or logos are subject to those third-party's policies.