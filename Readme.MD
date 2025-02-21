# MPF utility (Azure Deployment Minimum Permissions Finder)

This utility finds the minimum permissions required for a given Azure deployment. This can help when you need to figure out the details of what permissions a service principal or managed identity will need to deploy a given ARM template, bicep file or Terraform module. Similarly when assigning a Service Principal / Managed Identity to an Azure Policy Assignment, this utility can help you figure out the minimum permissions required by the Service Principal / Managed Identity to enforce/remediate the policy. It is recommended that the utility is used in a development or test environment to find the minimum permissions required.

## How It works

![MPF Flow](./docs/images/mpf-flow.svg)

The overview of how this utility works is as follows:

* The key parameters the utility needs are the **Service Principal details** (Client ID, Secret and Object ID) and details needed for the specific deployment provider: 
  * ARM: ARM template file and parameters file needed
  * Terraform: Terraform module directory and variables file needed
* The utility **removes any existing Role Assignments for provided Service Principal**
* A Custom Role (with no assigned permissions) is created
* The Service Principal (SP) is assigned the new custom role
* For the above steps the utitity uses the **default Azure CLI credentials** which needs to have permissions to create custom role, and role assignments. The details of the permissions required by the default Azure CLI credentials are provided in the [Permissions required by default Azure CLI credentials](#permissions-required-by-default-azure-cli-credentials) section.
* These sub steps are re-tried till the deployment succeeds
  * Depending on the provider (ARM, Bicep, or Terraform) a deployment is tried
  * If the Service Principal does not have sufficient permissions an authorization error is returned by the deployment. If Authorization errors have occured, they are parsed to fetch the missing scopes and permissions. The [authorizationErrorParser Tests](./pkg/domain/authorizationErrorParser_test.go) provides details of the different kinds of Authorization errors typically received.
  * The missing permissions are added to the custom Role.
  * The missing permissions are added to the result
* Once no authorization error is received, the utility prints the permissions assigned to the Service Principal
* The required permissions are displayed based on the display options. These options can be used to view the resource wise break up of permissions and also to export the result in JSON format
* All resources created are cleaned up by the utility including the Role Assignments and Custom Role.

## MPF Design

The [MPF Design](docs/mpf-design.md) document provides details on the design including the packages and abstractions.

## Supported Deployment Providers

* Azure **ARM** Template: The ARM endpoints of both modes described below return multiple authorization errors at a time and as a result the final output is usually displayed in under 90 seconds, even for complex templates.
  * Whatif mode: This this the default ARM mode. It uses the ARM whatif endpoint to get the authorization errors and find the minimum permissions required for a deployment, without actually create any resources.
  * FullDeployment mode: This mode has been deprecated and is no longer supported.
* **Bicep**: The Bicep mode is similar to the ARM whatif mode, and uses the ARM whatif endpoint to get the authorization errors and find the minimum permissions required for a deployment, without actually create any resources. Internally the utility converts the bicep file to ARM template and then uses the ARM whatif endpoint.
* **Terraform**: The Terraform mode finds the minimum permissions required for a deployment, by getting the authorization errors from the Terraform plan/apply, and destroy commands. All resources are cleaned up by the utility. Since terraform calls the ARM apis for one resource at a time, the authorization errors are not received in bulk, and as a result it can take quite long to get the final result. The overall time is the time taken to run the terraform plan/apply/destroy commands, plus the overhead of getting and parsing the authorization errors a few times.

## Flags and Environment Variables

The commands can be used with flags or environment variables. For details on the flags and environment variables, please refer to the [command line flags and environment variables](docs/commandline-flags-and-env-variables.MD) document.

## Installation

For installation instructions, please refer to the [installation and quickstart](docs/installation-and-quickstart.md) document.


## Usage Details

For usage details, please refer to the [quickstart / usage details](docs/installation-and-quickstart.md#quickstart--usage) document.


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