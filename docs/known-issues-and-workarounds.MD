# Known Issues and Workarounds

## Bicep

### Parameter File Format

Currently only ARM type [parameters files](https://github.com/Azure/mpf/blob/main/samples/bicep/aks-private-subnet-invalid-params.json) are supported even for bicep executions. Bicep type parameters files are currently not supported and result in an error. A new feature request has been created to track this issue [issue #12](https://github.com/Azure/mpf/issues/12).

## Terraform

### Token Expiry

If your default Azure credentials token issued for the utility expires before the utility completes the execution, that execution will fail. When this happens, the utility saves the permissions inferred up to that point in the Terraform module directory, and these are automatically added next time the utility executes for the same Terraform module directory.

### Terraform azurerm provider crash

The azurerm provider version < 4.2.0 can crash when using resources like `azurerm_application_insights` without all permissions to create the resource. Details can be found in the [GitHub issue](https://github.com/hashicorp/terraform-provider-azurerm/issues/27961#issuecomment-2467155082). If this issue is encountered, it is recommended to upgrade your azurerm provider version to 4.2.0 or higher. It must be noted that there are some Azure Verified Modules which still use older versions of the Azure provider.

### Existing Resource / Import Errors

Creation of certain resources, like the application insights resource, also involves the creation of current billing features resource, as described in the [GitHub issue](https://github.com/hashicorp/terraform-provider-azurerm/issues/27961#issuecomment-2467392936). This means that if the identity used by Terraform has permissions to create the application insights resource but not the current billing features resource, the application insights resource will be created in Azure, but the Terraform apply will fail. This means that the creation of the application insights resource will not be tracked by Terraform and the state file will be out of sync with the actual resources in Azure. When the utility adds the required permission and executes Terraform apply again, Terraform will give an error that the resource already exists in Azure and that this can be resolved by importing the resource into the Terraform state file. As a workaround, the utility automatically appends [missing required permissions](../pkg/domain/appendPermissionsForSpecialCases.go#L12-19) when ```Microsoft.Insights/components/read``` or ```Microsoft.Insights/components/write``` permissions are detected as missing permissions.
When ```Microsoft.Insights/components/read``` is detected ```Microsoft.Insights/components/currentbillingfeatures/read``` and ```Microsoft.AlertsManagement/smartDetectorAlertRules/read``` permissions are appended. Similarly, when ```Microsoft.Insights/components/write``` is detected ```Microsoft.Insights/components/currentbillingfeatures/write``` and ```Microsoft.AlertsManagement/smartDetectorAlertRules/write``` permissions are appended.


### Billing Features Payload Error

This issue is also related to the [GitHub Issue](https://github.com/hashicorp/terraform-provider-azurerm/issues/27961#issuecomment-2520407658). The utility retries the request to work around this issue.

### Authorization_RequestDenied Error

Currently if you attempt perform actions like adding an Azure AD group via terraform, a Authorization_RequestDenied Error is received.

Sample Error:

```
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