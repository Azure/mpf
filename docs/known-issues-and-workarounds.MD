# Known Issues and Workarounds

## Terraform

### Token Expiry

If your default Azure credentials token issued for the utility expires before the utility completes the execution, that execution will fail. When this happens the utility saves the permissions inferred upto that point in the terraform module directory, and these are automatically added next time the utility executes for the same terraform module directory.

### Terraform azurerm provider crash

The azurerm provider version < 4.2.0 can crash when using resources like `azurerm_application_insights` without all permissions to create the resource. Details can be found in the [GitHub issue](https://github.com/hashicorp/terraform-provider-azurerm/issues/27961#issuecomment-2467155082). If this issue is encountered it is recommended to upgrade your azurerm provider version to 4.2.0 or higher. It must be noted that there are some Azure Verified Modules which still use older versions of the azure provider.

### Existing Resource / Import Errors

Creation of certain resources, like the application insights resource, also involves the creation of current billing features resource, as described in the [GitHub issue](https://github.com/hashicorp/terraform-provider-azurerm/issues/27961#issuecomment-2467392936). This means that if the identity used by Terraform has permissions to create the application insights resource but not the current billing features resource, the application insights resource will be created in Azure, but the Terraform apply will fail. This means that the creation of the application insights resource will not be tracked by Terraform and the state file will be out of sync with the actual resources in Azure. When the utility adds the required permission and executes Terraform apply again, Terraform will give an error that the resource already exists in Azure and that this can be resolved by importing the resource into the Terraform state file. In this scenario, by default, the MPF utility will import such resources into the Terraform state file and proceed further. It needs to be noted that in the cleanup phase the imported resources will also be deleted.

### Billing Features Payload Error

This issue is also related to the [Github Issue](https://github.com/hashicorp/terraform-provider-azurerm/issues/27961#issuecomment-2520407658). The utility retries the request to workaround this issue.
