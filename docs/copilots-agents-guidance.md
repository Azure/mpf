# Copilots and Agents Integration Guide for MPF

This guide provides comprehensive information for integrating the Azure Deployment Minimum Permissions Finder (MPF) tool into automated workflows, AI agents, and copilot systems.

## Overview

MPF is designed to be automation-friendly and can be seamlessly integrated into CI/CD pipelines, AI agent workflows, and automated deployment processes. The tool provides structured JSON output that can be easily parsed and consumed programmatically.

### Why Use MPF in Automated Workflows?

- **Security Best Practices**: Automatically determine the least-privilege permissions needed for Azure deployments
- **DevSecOps Integration**: Embed security permission analysis into your deployment pipelines
- **Compliance Automation**: Generate permission reports for audit and compliance purposes
- **Dynamic IAM**: Automatically create and manage service principals with minimal required permissions

## Integration Patterns

### 1. Command-Line Integration

MPF can be invoked directly from any automation system that can execute shell commands:

```bash
# ARM Template deployment analysis
azmpf arm \
  --subscriptionID "12345678-1234-1234-1234-123456789012" \
  --tenantID "87654321-4321-4321-4321-210987654321" \
  --spClientID "abcd1234-5678-90ef-ghij-klmnopqrstuv" \
  --spObjectID "wxyz9876-5432-10fe-dcba-vutsrqponmlk" \
  --spClientSecret "your-secret-here" \
  --armTemplateFilePath "./template.json" \
  --parameterFilePath "./parameters.json" \
  --resourceGroupName "rg-test" \
  --jsonOutput

# Terraform module analysis
azmpf terraform \
  --subscriptionID "12345678-1234-1234-1234-123456789012" \
  --tenantID "87654321-4321-4321-4321-210987654321" \
  --spClientID "abcd1234-5678-90ef-ghij-klmnopqrstuv" \
  --spObjectID "wxyz9876-5432-10fe-dcba-vutsrqponmlk" \
  --spClientSecret "your-secret-here" \
  --executablePath "/usr/local/bin/terraform" \
  --workingDir "./terraform-config" \
  --varFilePath "./terraform.tfvars" \
  --jsonOutput

# Bicep deployment analysis
azmpf bicep \
  --subscriptionID "12345678-1234-1234-1234-123456789012" \
  --tenantID "87654321-4321-4321-4321-210987654321" \
  --spClientID "abcd1234-5678-90ef-ghij-klmnopqrstuv" \
  --spObjectID "wxyz9876-5432-10fe-dcba-vutsrqponmlk" \
  --spClientSecret "your-secret-here" \
  --bicepFilePath "./main.bicep" \
  --parameterFilePath "./parameters.json" \
  --resourceGroupName "rg-test" \
  --bicepExecPath "/usr/local/bin/bicep" \
  --jsonOutput
```

### 2. Prerequisites: Service Principal Setup

Before integrating MPF into your automated workflows, you need to create a service principal in your Azure Active Directory tenant. The service principal requires **no roles assigned to it** initially, as MPF will dynamically manage role assignments during execution and remove them when complete.

#### Creating a Service Principal

You can create a service principal using the Azure CLI or the Azure portal. Here's how to create one using the Azure CLI:

```bash
# Login to Azure
az login

# Create service principal with no role assignments
MPF_SP=$(az ad sp create-for-rbac --name "MPF_SP" --skip-assignment)

# Extract the required values
MPF_SPCLIENTID=$(echo $MPF_SP | jq -r .appId)
MPF_SPCLIENTSECRET=$(echo $MPF_SP | jq -r .password)
MPF_SPOBJECTID=$(az ad sp show --id $MPF_SPCLIENTID --query id -o tsv)

# Display the values for use in automation
echo "Service Principal created successfully!"
echo "Client ID: $MPF_SPCLIENTID"
echo "Object ID: $MPF_SPOBJECTID"
# Note: Never display secrets in logs or console output for security
echo "Client Secret: [HIDDEN - Store securely for automation use]"
```

#### Alternative: Using Azure Portal

1. Navigate to **Azure Active Directory** > **App registrations**
2. Click **New registration**
3. Provide a name (e.g., "MPF-ServicePrincipal")
4. Select **Accounts in this organizational directory only**
5. Click **Register**
6. Note the **Application (client) ID** and **Directory (tenant) ID**
7. Go to **Certificates & secrets** > **New client secret**
8. Create a secret and save the **Value** immediately
9. Go to **Overview** and note the **Object ID**

#### Security Considerations for Service Principals

- **Store credentials securely**: Use Azure Key Vault, GitHub Secrets, or Azure DevOps Variable Groups
- **Rotate secrets regularly**: Set up automated secret rotation
- **Monitor usage**: Enable auditing for the service principal
- **Principle of least privilege**: MPF handles this automatically by assigning minimal permissions

### 3. Environment Variable Configuration

For better security and automation, use environment variables instead of command-line parameters:

```bash
export MPF_SUBSCRIPTIONID="12345678-1234-1234-1234-123456789012"
export MPF_TENANTID="87654321-4321-4321-4321-210987654321"
export MPF_SPCLIENTID="abcd1234-5678-90ef-ghij-klmnopqrstuv"
export MPF_SPOBJECTID="wxyz9876-5432-10fe-dcba-vutsrqponmlk"
export MPF_SPCLIENTSECRET="your-secret-here"

# Then run MPF with minimal parameters
azmpf arm --armTemplateFilePath "./template.json" --parameterFilePath "./parameters.json" --resourceGroupName "rg-test" --jsonOutput
```

## Output Consumption

### JSON Output Format

When using the `--jsonOutput` flag, MPF provides structured output that can be easily parsed:

```json
{
  "subscriptionId": "12345678-1234-1234-1234-123456789012",
  "deploymentProvider": "ARM",
  "templateFilePath": "./template.json",
  "parameterFilePath": "./parameters.json",
  "resourceGroupName": "rg-test",
  "executionTime": "2m30s",
  "totalPermissions": 15,
  "permissions": [
    {
      "scope": "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/rg-test",
      "actions": [
        "Microsoft.Resources/subscriptions/resourcegroups/read",
        "Microsoft.Resources/subscriptions/resourcegroups/write"
      ]
    },
    {
      "scope": "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/rg-test/providers/Microsoft.Storage/storageAccounts/mystorageaccount",
      "actions": [
        "Microsoft.Storage/storageAccounts/read",
        "Microsoft.Storage/storageAccounts/write",
        "Microsoft.Storage/storageAccounts/delete"
      ]
    }
  ]
}
```

### Parsing JSON Output

Here are examples of how to parse and use the JSON output in different environments:

#### Python Example

```python
import json
import subprocess

def run_mpf_analysis(template_path, parameters_path, rg_name):
    """Run MPF analysis and return parsed results."""
    cmd = [
        "azmpf", "arm",
        "--armTemplateFilePath", template_path,
        "--parameterFilePath", parameters_path,
        "--resourceGroupName", rg_name,
        "--jsonOutput"
    ]
    
    result = subprocess.run(cmd, capture_output=True, text=True)
    
    if result.returncode != 0:
        raise Exception(f"MPF failed: {result.stderr}")
    
    return json.loads(result.stdout)

def create_custom_role_definition(mpf_result):
    """Create Azure custom role definition from MPF results."""
    permissions = []
    for perm in mpf_result["permissions"]:
        permissions.extend(perm["actions"])
    
    role_definition = {
        "Name": f"CustomRole-{mpf_result['resourceGroupName']}",
        "Description": f"Custom role for {mpf_result['templateFilePath']} deployment",
        "Actions": list(set(permissions)),  # Remove duplicates
        "NotActions": [],
        "AssignableScopes": [f"/subscriptions/{mpf_result['subscriptionId']}"]
    }
    
    return role_definition
```

#### PowerShell Example

```powershell
function Invoke-MPFAnalysis {
    param(
        [string]$TemplatePath,
        [string]$ParametersPath,
        [string]$ResourceGroupName
    )
    
    $result = azmpf arm --armTemplateFilePath $TemplatePath --parameterFilePath $ParametersPath --resourceGroupName $ResourceGroupName --jsonOutput
    
    if ($LASTEXITCODE -ne 0) {
        throw "MPF analysis failed"
    }
    
    return $result | ConvertFrom-Json
}

function New-CustomRoleFromMPF {
    param([object]$MPFResult)
    
    $actions = $MPFResult.permissions | ForEach-Object { $_.actions } | Sort-Object -Unique
    
    $roleDefinition = @{
        Name = "CustomRole-$($MPFResult.resourceGroupName)"
        Description = "Custom role for $($MPFResult.templateFilePath) deployment"
        Actions = $actions
        NotActions = @()
        AssignableScopes = @("/subscriptions/$($MPFResult.subscriptionId)")
    }
    
    return $roleDefinition
}
```

#### Bash/Shell Example

```bash
#!/bin/bash

run_mpf_analysis() {
    local template_path="$1"
    local parameters_path="$2"
    local rg_name="$3"
    
    azmpf arm \
        --armTemplateFilePath "$template_path" \
        --parameterFilePath "$parameters_path" \
        --resourceGroupName "$rg_name" \
        --jsonOutput > mpf_result.json
    
    if [ $? -ne 0 ]; then
        echo "MPF analysis failed" >&2
        exit 1
    fi
    
    echo "mpf_result.json"
}

extract_permissions() {
    local result_file="$1"
    jq -r '.permissions[].actions[]' "$result_file" | sort | uniq
}

create_role_assignment() {
    local result_file="$1"
    local sp_object_id="$2"
    
    local subscription_id=$(jq -r '.subscriptionId' "$result_file")
    local permissions=$(extract_permissions "$result_file")
    
    # Create custom role definition
    echo "Creating custom role with permissions: $permissions"
    # Implementation would go here
}
```

## CI/CD Integration Examples

### GitHub Actions

```yaml
name: Azure Deployment with MPF Analysis

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  analyze-permissions:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    
    - name: Download MPF
      run: |
        curl -L -o azmpf https://github.com/Azure/mpf/releases/latest/download/azmpf-linux-amd64
        chmod +x azmpf
        sudo mv azmpf /usr/local/bin/
    
    - name: Run MPF Analysis
      env:
        MPF_SUBSCRIPTIONID: ${{ secrets.AZURE_SUBSCRIPTION_ID }}
        MPF_TENANTID: ${{ secrets.AZURE_TENANT_ID }}
        MPF_SPCLIENTID: ${{ secrets.AZURE_CLIENT_ID }}
        MPF_SPOBJECTID: ${{ secrets.AZURE_SP_OBJECT_ID }}
        MPF_SPCLIENTSECRET: ${{ secrets.AZURE_CLIENT_SECRET }}
      run: |
        azmpf arm \
          --armTemplateFilePath ./infrastructure/template.json \
          --parameterFilePath ./infrastructure/parameters.json \
          --resourceGroupName "rg-${{ github.run_id }}" \
          --jsonOutput > permissions.json
    
    - name: Upload Permissions Report
      uses: actions/upload-artifact@v2
      with:
        name: permissions-report
        path: permissions.json
    
    - name: Create Custom Role
      run: |
        python scripts/create_custom_role.py permissions.json
```

### Azure DevOps

```yaml
trigger:
- main

pool:
  vmImage: 'ubuntu-latest'

steps:
- task: Bash@3
  displayName: 'Download MPF'
  inputs:
    targetType: 'inline'
    script: |
      curl -L -o azmpf https://github.com/Azure/mpf/releases/latest/download/azmpf-linux-amd64
      chmod +x azmpf
      sudo mv azmpf /usr/local/bin/

- task: Bash@3
  displayName: 'Run MPF Analysis'
  env:
    MPF_SUBSCRIPTIONID: $(AZURE_SUBSCRIPTION_ID)
    MPF_TENANTID: $(AZURE_TENANT_ID)
    MPF_SPCLIENTID: $(AZURE_CLIENT_ID)
    MPF_SPOBJECTID: $(AZURE_SP_OBJECT_ID)
    MPF_SPCLIENTSECRET: $(AZURE_CLIENT_SECRET)
  inputs:
    targetType: 'inline'
    script: |
      azmpf arm \
        --armTemplateFilePath ./infrastructure/template.json \
        --parameterFilePath ./infrastructure/parameters.json \
        --resourceGroupName "rg-$(Build.BuildId)" \
        --jsonOutput > $(Agent.TempDirectory)/permissions.json

- task: PublishBuildArtifacts@1
  displayName: 'Publish Permissions Report'
  inputs:
    pathToPublish: '$(Agent.TempDirectory)/permissions.json'
    artifactName: 'permissions-report'
```

## Error Handling

### Common Error Scenarios

1. **Authentication Failures**
   - Verify service principal credentials
   - Ensure service principal has sufficient permissions for MPF operations
   - Check tenant ID and subscription ID accuracy

2. **Resource Creation Failures**
   - Verify resource names are unique
   - Check Azure quota limits
   - Ensure target resource group exists or can be created

3. **Template/Module Validation Errors**
   - Validate ARM/Bicep template syntax
   - Verify Terraform module configuration
   - Check parameter file format and values

### Error Handling Best Practices

```python
import subprocess
import json
import logging

def run_mpf_with_retry(template_path, parameters_path, rg_name, max_retries=3):
    """Run MPF with retry logic for transient failures."""
    for attempt in range(max_retries):
        try:
            result = subprocess.run([
                "azmpf", "arm",
                "--armTemplateFilePath", template_path,
                "--parameterFilePath", parameters_path,
                "--resourceGroupName", f"{rg_name}-{attempt}",
                "--jsonOutput"
            ], capture_output=True, text=True, timeout=1800)  # 30 minute timeout
            
            if result.returncode == 0:
                return json.loads(result.stdout)
            else:
                logging.warning(f"Attempt {attempt + 1} failed: {result.stderr}")
                if attempt == max_retries - 1:
                    raise Exception(f"MPF failed after {max_retries} attempts: {result.stderr}")
        
        except subprocess.TimeoutExpired:
            logging.warning(f"Attempt {attempt + 1} timed out")
            if attempt == max_retries - 1:
                raise Exception("MPF timed out after maximum retries")
        
        except json.JSONDecodeError as e:
            logging.warning(f"Invalid JSON output on attempt {attempt + 1}: {e}")
            if attempt == max_retries - 1:
                raise Exception("Failed to parse MPF JSON output")
```

## Best Practices for Automation

### 1. Security Considerations

- **Use managed identities** when running in Azure environments
- **Store secrets securely** using key vaults or secret management systems
- **Rotate service principal credentials** regularly
- **Use separate service principals** for different environments (dev/test/prod)

### 2. Performance Optimization

- **Use parallel execution** for multiple template analyses
- **Cache results** for identical templates/parameters combinations
- **Set appropriate timeouts** (ARM/Bicep: 30-45 minutes, Terraform: 60+ minutes)
- **Clean up resources** promptly to avoid quota issues

### 3. Monitoring and Logging

- **Log execution times** for performance monitoring
- **Track permission changes** over time
- **Monitor failure rates** and common error patterns
- **Set up alerts** for unexpected failures

### 4. Resource Management

```bash
# Example: Automated cleanup script
#!/bin/bash

cleanup_mpf_resources() {
    local subscription_id="$1"
    local prefix="mpf-temp"
    
    # Clean up resource groups created by MPF
    az group list --query "[?starts_with(name, '$prefix')].name" -o tsv | \
    while read rg_name; do
        echo "Cleaning up resource group: $rg_name"
        az group delete --name "$rg_name" --yes --no-wait
    done
    
    # Clean up custom roles created by MPF
    az role definition list --custom-role-only --query "[?starts_with(roleName, 'MPF-Custom-Role')].name" -o tsv | \
    while read role_name; do
        echo "Cleaning up custom role: $role_name"
        az role definition delete --name "$role_name"
    done
}
```

## Advanced Integration Scenarios

### 1. Policy-as-Code Integration

```python
def integrate_with_policy_framework(mpf_result, policy_framework):
    """Integrate MPF results with policy-as-code frameworks."""
    permissions = extract_permissions(mpf_result)
    
    if policy_framework == "open-policy-agent":
        return create_opa_policy(permissions)
    elif policy_framework == "azure-policy":
        return create_azure_policy(permissions)
    else:
        raise ValueError(f"Unsupported policy framework: {policy_framework}")

def create_opa_policy(permissions):
    """Create OPA policy from MPF permissions."""
    policy = {
        "package": "azure.rbac",
        "default": {"allow": False},
        "allow": {
            "input.action": {"$in": permissions}
        }
    }
    return policy
```

### 2. Infrastructure-as-Code Integration

```python
def generate_terraform_rbac(mpf_result):
    """Generate Terraform RBAC configuration from MPF results."""
    permissions = extract_permissions(mpf_result)
    
    terraform_config = f"""
resource "azurerm_role_definition" "custom_role" {{
  name        = "CustomRole-{mpf_result['resourceGroupName']}"
  scope       = "/subscriptions/{mpf_result['subscriptionId']}"
  description = "Custom role for {mpf_result['templateFilePath']} deployment"

  permissions {{
    actions = [
      {','.join([f'"{perm}"' for perm in permissions])}
    ]
    not_actions = []
  }}

  assignable_scopes = [
    "/subscriptions/{mpf_result['subscriptionId']}"
  ]
}}

resource "azurerm_role_assignment" "custom_assignment" {{
  scope              = "/subscriptions/{mpf_result['subscriptionId']}"
  role_definition_id = azurerm_role_definition.custom_role.role_definition_resource_id
  principal_id       = var.service_principal_object_id
}}
"""
    return terraform_config
```

## Troubleshooting

### Common Issues and Solutions

1. **"Service Principal not found" errors**
   - Verify the service principal exists in the correct tenant
   - Check that the object ID is correct (not the application ID)

2. **"Insufficient permissions" errors**
   - Ensure your default Azure CLI credentials have the required permissions
   - Review the [permissions requirements](../Readme.MD#permissions-required-by-default-azure-cli-credentials)

3. **Timeout errors during long-running deployments**
   - Increase timeout values for complex templates
   - Consider breaking large templates into smaller modules

4. **JSON parsing errors**
   - Check for mixed output (stdout/stderr)
   - Use `--debug` flag to identify output issues
   - Redirect stderr to separate from JSON output

### Debug Mode

Use the `--debug` flag for detailed execution information:

```bash
azmpf arm \
  --armTemplateFilePath "./template.json" \
  --parameterFilePath "./parameters.json" \
  --resourceGroupName "rg-debug" \
  --debug \
  --jsonOutput 2>debug.log 1>result.json
```

## Additional Resources

- [MPF Installation and Quickstart](./installation-and-quickstart.md)
- [Command Line Flags and Environment Variables](./commandline-flags-and-env-variables.md)
- [Display Options](./display-options.MD)
- [Known Issues and Workarounds](./known-issues-and-workarounds.MD)
- [MPF Design Documentation](./mpf-design.md)

## Contributing

If you have suggestions for improving this automation guidance or want to share additional integration patterns, please contribute to the [MPF repository](https://github.com/Azure/mpf).
