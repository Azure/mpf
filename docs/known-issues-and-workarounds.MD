# Known Issues and Workarounds

## Bicep

### Parameter File Format

Both ARM-style JSON parameters files and Bicep-native `.bicepparam` files are supported for Bicep executions. When a `.bicepparam` file is provided via `--parametersFilePath`, MPF automatically compiles it to ARM JSON format using `bicep build-params` before deployment. See the [quickstart guide](installation-and-quickstart.md) for examples of both formats.

## Terraform

### Remote Backend Access Denied

When using a remote backend (e.g., Azure Storage for Terraform state), MPF removes all existing role assignments from the service principal before starting analysis. This can cause Terraform to fail with a 403 error when trying to access the backend.

**Workaround**: Use the `--initialPermissions` flag to specify the permissions required to access the remote backend. These permissions will be added to the custom role before Terraform runs.

Example for Azure Storage backend:

```bash
azmpf terraform \
  --initialPermissions "Microsoft.Storage/storageAccounts/read,Microsoft.Storage/storageAccounts/listKeys/action,Microsoft.Storage/storageAccounts/blobServices/containers/read,Microsoft.Storage/storageAccounts/blobServices/containers/blobs/read,Microsoft.Storage/storageAccounts/blobServices/containers/blobs/write" \
  --tfPath $(which terraform) \
  --workingDir ./my-terraform \
  # ... other required flags
```

Or create a JSON file with the permissions and reference it:

```bash
azmpf terraform \
  --initialPermissions @backend-permissions.json \
  # ... other flags
```

See [Initial Permissions](./commandline-flags-and-env-variables.md#initial-permissions) for more details on the JSON file format.

### Token Expiry

If your default Azure credentials token issued for the utility expires before the utility completes the execution, that execution will fail. When this happens, the utility saves the permissions inferred up to that point in the Terraform module directory, and these are automatically added next time the utility executes for the same Terraform module directory.

### Terraform azurerm provider crash

The azurerm provider version < 4.2.0 can crash when using resources like `azurerm_application_insights` without all permissions to create the resource. Details can be found in the [GitHub issue](https://github.com/hashicorp/terraform-provider-azurerm/issues/27961#issuecomment-2467155082). If this issue is encountered, it is recommended to upgrade your azurerm provider version to 4.2.0 or higher. It must be noted that there are some Azure Verified Modules which still use older versions of the Azure provider.

### Existing Resource / Import Errors

Creation of certain resources, like the application insights resource, also involves the creation of current billing features resource, as described in the [GitHub issue](https://github.com/hashicorp/terraform-provider-azurerm/issues/27961#issuecomment-2467392936). This means that if the identity used by Terraform has permissions to create the application insights resource but not the current billing features resource, the application insights resource will be created in Azure, but the Terraform apply will fail. This means that the creation of the application insights resource will not be tracked by Terraform and the state file will be out of sync with the actual resources in Azure. When the utility adds the required permission and executes Terraform apply again, Terraform will give an error that the resource already exists in Azure and that this can be resolved by importing the resource into the Terraform state file. As a workaround, the utility automatically appends [missing required permissions](../pkg/domain/appendPermissionsForSpecialCases.go#L12-19) when ```Microsoft.Insights/components/read``` or ```Microsoft.Insights/components/write``` permissions are detected as missing permissions.
When ```Microsoft.Insights/components/read``` is detected ```Microsoft.Insights/components/currentbillingfeatures/read``` and ```Microsoft.AlertsManagement/smartDetectorAlertRules/read``` permissions are appended. Similarly, when ```Microsoft.Insights/components/write``` is detected ```Microsoft.Insights/components/currentbillingfeatures/write``` and ```Microsoft.AlertsManagement/smartDetectorAlertRules/write``` permissions are appended.

### Billing Features Payload Error

This issue is also related to the [GitHub Issue](https://github.com/hashicorp/terraform-provider-azurerm/issues/27961#issuecomment-2520407658). The utility retries the request to work around this issue.

### Retrieving Static Website Properties Error

When analyzing a private Storage Account, Terraform may fail with an error like:

```text
Error: retrieving static website properties for Storage Account (Subscription: "SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS"
Resource Group Name: "rg-example"
Storage Account Name: "saexample"): executing request: unexpected status 404 (404 The specified resource does not exist.) with ResourceNotFound: The specified resource does not exist.
```

Setting and retrieving static website properties uses the Storage [data plane API](https://github.com/hashicorp/terraform-provider-azurerm/issues/24982#issuecomment-1972246655), which is served from the storage account's own endpoint. That endpoint is subject to the same network access restrictions as the account itself, so when the account has a private endpoint or its network rules deny public access, the request is only reachable from a network location that can resolve and reach the private endpoint.

Because this is a network reachability failure rather than a control-plane authorization failure, the returned error carries no missing-permission details, so the utility cannot infer or append any RBAC permission to resolve it.

**Workarounds**:

- Run the utility from a network location that can reach the storage account's private endpoint, such as a virtual machine or self-hosted runner joined to the account's virtual network (or a peered/linked network) with private DNS resolution configured for the storage account's `web` sub-resource.
- Alternatively, if the goal is only to determine the minimum permissions, temporarily allow public network access on the storage account (or remove the private endpoint / network rules) in the Terraform configuration during analysis, then restore the private configuration afterwards.

### Authorization_RequestDenied Error

Currently if you attempt perform actions like adding an Azure AD group via terraform, a Authorization_RequestDenied Error is received.

Sample Error:

```text
Error: Creating group "Group-name-axtwb"

  with ...._ds_group[0],
  on ....../rbac.tf line 3, in resource "azuread_group" "res_ds_group":
   3: resource "azuread_group" "res_ds_group" {

GroupsClient.BaseClient.Post(): unexpected status 403 with OData error:
Authorization_RequestDenied: Insufficient privileges to complete the ...
```

From the [terraform docs](https://registry.terraform.io/providers/hashicorp/azuread/latest/docs/guides/service_principal_configuration), adding these permissions may require global admin privilege or admin consent. For this reason the utility cannot automatically add any permissions to the MPF SP to get around this error.

A potential workaround is to disable the Azure AD resource creation which caused this error in the terraform code, and then re-execute the utility.

## Common

### Execution Time Reduction Workarounds

For ARM and Bicep, autoAddReadPermissionForEachWrite is set to true by default. This means that if a write permission is detected, the utility will automatically add the corresponding read permission.

For Terraform autoAddDeletePermissionForEachWrite is set to true by default. This means that if a write permission is detected, the utility will automatically add the corresponding delete permission.
