# AI Agent Instructions for Azure MPF

This document provides instructions for AI agents (such as GitHub Copilot CLI) to analyze Azure Infrastructure as Code (IaC) repositories using the MPF (Minimum Permissions Finder) utility.

## Overview

When asked to analyze Azure IaC for minimum permissions, follow this process:

1. Identify the IaC type (ARM, Bicep, or Terraform)
2. Execute the Azure MPF utility to determine minimum required permissions
3. List the minimum permissions needed for deployment
4. Suggest Azure built-in RBAC roles that cover those permissions

## Prerequisites

### Required Tools

- **Azure CLI** (`az`): Must be logged in with an account that has the following permissions:
  - `Microsoft.Authorization/roleDefinitions/read|write|delete`
  - `Microsoft.Authorization/roleAssignments/read|write|delete`
  - `Microsoft.Resources/subscriptions/resourcegroups/read|write|delete`

- **MPF Utility** (`azmpf`): Download from [Azure/mpf releases](https://github.com/Azure/mpf/releases)

- **For Bicep**: Bicep CLI installed (`bicep` or `az bicep`)
- **For Terraform**: Terraform CLI installed (`terraform`)

### Environment Variables (Required)

```bash
export MPF_SUBSCRIPTIONID="<your-subscription-id>"
export MPF_TENANTID="<your-tenant-id>"
export MPF_SPCLIENTID="<service-principal-client-id>"
export MPF_SPCLIENTSECRET="<service-principal-secret>"
export MPF_SPOBJECTID="<service-principal-object-id>"
```

### Service Principal Setup

Create a dedicated Service Principal for MPF analysis (it should have NO roles assigned - MPF manages roles dynamically):

```bash
MPF_SP=$(az ad sp create-for-rbac --name "MPF-Analyzer-SP" --skip-assignment)
export MPF_SPCLIENTID=$(echo $MPF_SP | jq -r .appId)
export MPF_SPCLIENTSECRET=$(echo $MPF_SP | jq -r .password)
export MPF_SPOBJECTID=$(az ad sp show --id $MPF_SPCLIENTID --query id -o tsv)
export MPF_TENANTID=$(az account show --query tenantId -o tsv)
export MPF_SUBSCRIPTIONID=$(az account show --query id -o tsv)
```

---

## Step 1: Identify IaC Type

### Detection Patterns

| IaC Type   | File Patterns                                        | Identifying Markers                              |
|------------|------------------------------------------------------|--------------------------------------------------|
| **ARM**    | `*.json` (with `$schema` containing `deploymentTemplate`) | `"$schema": "https://schema.management.azure.com/schemas/...deploymentTemplate.json"` |
| **Bicep**  | `*.bicep`                                           | Bicep syntax: `resource`, `param`, `var`, `output` keywords |
| **Terraform** | `*.tf`, `*.tf.json`                              | `terraform {}` block, `provider "azurerm" {}` |

### Detection Commands

```bash
# Check for ARM templates
find . -name "*.json" -exec grep -l 'deploymentTemplate.json' {} \; 2>/dev/null

# Check for Bicep files
find . -name "*.bicep" 2>/dev/null

# Check for Terraform files
find . -name "*.tf" -o -name "*.tf.json" 2>/dev/null
```

---

## Step 2: Execute MPF Analysis

### ARM Template Analysis

```bash
azmpf arm \
  --templateFilePath ./path/to/template.json \
  --parametersFilePath ./path/to/parameters.json \
  --jsonOutput \
  --verbose
```

### Bicep Analysis

```bash
export MPF_BICEPEXECPATH=$(which bicep)

azmpf bicep \
  --bicepFilePath ./path/to/main.bicep \
  --parametersFilePath ./path/to/parameters.json \
  --jsonOutput \
  --verbose
```

### Terraform Analysis

```bash
export MPF_TFPATH=$(which terraform)

# Ensure terraform is initialized
cd ./terraform-module-dir
terraform init

azmpf terraform \
  --workingDir $(pwd) \
  --varFilePath ./terraform.tfvars \
  --jsonOutput \
  --verbose
```

### Common Optional Flags

- `--location <region>` - Azure region (default: eastus2)
- `--resourceGroupNamePfx <prefix>` - Resource group prefix (default: testdeployrg)
- `--showDetailedOutput` - Shows per-resource permissions

---

## Step 3: Parse MPF Output

MPF outputs permissions in JSON format when using `--jsonOutput`:

```json
{
  "permissions": [
    "Microsoft.Storage/storageAccounts/read",
    "Microsoft.Storage/storageAccounts/write",
    "Microsoft.Resources/deployments/read",
    "Microsoft.Resources/deployments/write"
  ]
}
```

---

## Step 4: Suggest Azure Built-in Roles

Match discovered permissions against Azure built-in roles:

| Resource Provider | Suggested Role | Role ID |
|-------------------|----------------|---------|
| Microsoft.Compute | Virtual Machine Contributor | `9980e02c-c2be-4d73-94e8-173b1dc7cf3c` |
| Microsoft.Network | Network Contributor | `4d97b98b-1d4f-4787-a291-c67834d212e7` |
| Microsoft.Storage | Storage Account Contributor | `17d1049b-9a84-46fb-8f53-869881c3d3ab` |
| Microsoft.ContainerService | Azure Kubernetes Service Contributor | `ed7f3fbd-7b88-4dd4-9017-9adb7ce333f8` |
| Microsoft.ContainerRegistry | AcrPush | `8311e382-0749-4cb8-b61a-304f252e45ec` |
| Microsoft.KeyVault | Key Vault Contributor | `f25e0fa2-a7c8-4377-a976-54943a77a395` |
| Microsoft.Web | Website Contributor | `de139f84-1756-47ae-9be6-808fbbe84772` |
| Microsoft.Sql | SQL DB Contributor | `9b7fa17d-e63e-47b0-bb0a-15c516ac86ec` |
| Microsoft.Authorization | User Access Administrator | `18d7d88d-d35e-4fb5-a5c3-7773c20a72d9` |

### Generate Custom Role Definition

For least-privilege access, generate a custom role:

```json
{
  "Name": "Custom Deployment Role",
  "Description": "Minimum permissions for IaC deployment",
  "Actions": ["<permissions-from-mpf-output>"],
  "NotActions": [],
  "AssignableScopes": ["/subscriptions/<subscription-id>"]
}
```

---

## Cleanup

After analysis, delete the Service Principal:

```bash
az ad sp delete --id "$MPF_SPCLIENTID"
```

---

## Important Notes

1. **Use in Dev/Test Only**: MPF actually deploys resources to determine permissions. Always run in a development or test environment.

2. **Resource Cleanup**: MPF automatically cleans up resources it creates, but verify cleanup completed successfully.

3. **Execution Time**: Analysis can take several minutes depending on template complexity.

4. **Terraform State**: For Terraform, MPF may import existing resources into state. Use isolated environments.

## Additional Resources

- [Azure MPF GitHub Repository](https://github.com/Azure/mpf)
- [Azure Built-in Roles Reference](https://learn.microsoft.com/azure/role-based-access-control/built-in-roles)
- [Azure Permissions Reference](https://azure.permissions.cloud/)
