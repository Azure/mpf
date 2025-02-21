# MPF command line flags and environment variables

## Global Flags (Common to all providers)

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

When used for Terraform, the verbose and debug flags show detailed logs from Terraform.

## ARM Flags

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

## Terraform Flags

| Flag                 | Environment Variable     | Required / Optional | Description                                                                                                      |
| -------------------- | ------------------------ | ------------------ | ---------------------------------------------------------------------------------------------------------------- |
| tfPath               | MPF_TFPATH               | Required           | Path to the Terraform executable                                                                             |
| workingDir           | MPF_WORKINGDIR           | Required           | Path to the Terraform module directory                                                                             |
| varFilePath          | MPF_VARFILEPATH          | Optional           | Path to the Terraform variables file                                                                             |
| importExistingResourcesToState | MPF_IMPORTEXISTINGRESOURCESTOSTATE | Optional | Default Value is true. This is required for some scenarios as described in the [Known Issues - Import Errors](./docs/known-issues-and-workarounds.MD#existing-resource--import-errors) |
| targetModule        | MPF_TARGETMODULE         | Optional           | Target module to be used for the Terraform deployment |                                                                             |

