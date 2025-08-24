# GitHub Copilot Instructions for MPF Development

This file provides context and instructions for GitHub Copilot to assist developers working on the Azure Deployment Minimum Permissions Finder (MPF) codebase.

## Project Overview

MPF is a CLI tool that analyzes Azure deployments (ARM templates, Bicep, Terraform) to determine the minimum required permissions for deployment. The tool uses a "deploy and discover" approach where it temporarily deploys resources with a service principal to capture the exact permissions needed.

### Core Purpose
- **Security**: Determine least-privilege permissions for Azure deployments
- **Automation**: Enable seamless integration into CI/CD pipelines and IaC workflows
- **Compliance**: Generate audit-ready permission reports for governance

## Architecture & Patterns

### Clean Architecture Structure
```
cmd/                    # CLI commands and entry points (Cobra)
├── main.go            # Application entry point
├── rootCmd.go         # Root command configuration
├── armCmd.go          # ARM template command
├── bicepCmd.go        # Bicep command  
└── terraformCmd.go    # Terraform command

pkg/
├── domain/            # Business logic and domain models
├── infrastructure/    # External dependencies (Azure SDK, file system)
├── presentation/      # Output formatting and display
└── usecase/          # Application use cases and orchestration
```

### Key Go Patterns Used
- **Dependency Injection**: Services are injected rather than instantiated
- **Error Wrapping**: Use `fmt.Errorf("context: %w", err)` for error chains
- **Interface Segregation**: Small, focused interfaces for testability
- **Builder Pattern**: Configuration builders for complex objects

## Azure Domain Knowledge

### Service Principal Authentication
```go
// Standard service principal configuration
type ServicePrincipal struct {
    ClientID     string // Application ID from Azure AD
    ClientSecret string // Secret value (never log this)
    ObjectID     string // Service principal object ID
    TenantID     string // Azure AD tenant ID
}
```

### Required Service Principal Setup
MPF requires a service principal with:
- **No initial role assignments** (MPF manages roles dynamically)
- **Application permissions** in Azure AD
- **Access to target subscription** for role assignment during execution

#### Creating Service Principal for Development/Testing
```bash
# Login to Azure
az login

# Create service principal with no role assignments
MPF_SP=$(az ad sp create-for-rbac --name "MPF-Dev-SP" --skip-assignment)

# Extract required values
export MPF_SPCLIENTID=$(echo $MPF_SP | jq -r .appId)
export MPF_SPCLIENTSECRET=$(echo $MPF_SP | jq -r .password)  
export MPF_SPOBJECTID=$(az ad sp show --id $MPF_SPCLIENTID --query id -o tsv)
export MPF_SUBSCRIPTIONID="your-subscription-id"
export MPF_TENANTID="your-tenant-id"
```

### Azure Resource Deployment Modes
- **ARM Templates**: Use deployment mode "Complete" for accurate permission detection
- **Bicep**: Compiled to ARM, follows same deployment patterns
- **Terraform**: Uses Azure Provider, requires plan/apply cycle simulation

### Permission Analysis Patterns
```go
// Common permission structures
type Permission struct {
    Action string `json:"action"`           // e.g., "Microsoft.Resources/deployments/write"
    Scope  string `json:"scope"`            // e.g., "/subscriptions/{id}/resourceGroups/{rg}"
    Level  string `json:"level"`            // subscription, resourceGroup, resource
}

// Permission grouping by scope
type PermissionsByScope map[string][]Permission
```

## Technology Stack & Dependencies

### Core Dependencies
- **CLI Framework**: `github.com/spf13/cobra` - command structure
- **Configuration**: `github.com/spf13/viper` - environment variables and config files
- **Azure SDK**: `github.com/Azure/azure-sdk-for-go` - Azure API interactions
- **Logging**: `github.com/sirupsen/logrus` - structured logging
- **Testing**: `github.com/stretchr/testify` - assertions and mocks
- **Terraform**: `github.com/hashicorp/terraform-exec` - Terraform execution

### Environment Variable Conventions
```go
// All environment variables use MPF_ prefix
const (
    EnvPrefix = "MPF"
    EnvSubscriptionID = "MPF_SUBSCRIPTIONID"
    EnvTenantID = "MPF_TENANTID"
    EnvSPClientID = "MPF_SPCLIENTID"
    EnvSPObjectID = "MPF_SPOBJECTID"
    EnvSPClientSecret = "MPF_SPCLIENTSECRET"
)
```

## Testing Conventions

### Unit Test Patterns
```go
func TestFunctionName(t *testing.T) {
    // Arrange
    input := "test input"
    expected := "expected output"
    
    // Act
    result, err := FunctionToTest(input)
    
    // Assert
    assert.Nil(t, err)
    assert.Equal(t, expected, result)
}
```

### Mock Patterns
```go
// Use testify/mock for complex dependencies
type MockAzureClient struct {
    mock.Mock
}

func (m *MockAzureClient) CreateDeployment(ctx context.Context, params DeploymentParams) error {
    args := m.Called(ctx, params)
    return args.Error(0)
}
```

### Test File Organization
- Co-locate tests with source: `file.go` → `file_test.go`
- Use descriptive test names: `TestCompletePartialActionResourceGroup`
- Group related tests in same file
- Use table-driven tests for multiple scenarios

## Build & Development Commands

```bash
# Build binary
make build

# Run tests
make test

# Run end-to-end tests (requires Azure credentials)
make test-e2e-arm
make test-e2e-bicep  
make test-e2e-terraform

# Clean build artifacts
make clean

# Cross-platform builds
make build-all
```

## Code Style & Conventions

### File Headers
All Go files should include the MIT license header:
```go
//     MIT License
//
//     Copyright (c) Microsoft Corporation.
//     [full license text...]

package packagename
```

### Error Handling
```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to deploy ARM template: %w", err)
}

// Use domain-specific error types when appropriate
type DeploymentError struct {
    TemplateFile string
    Cause        error
}

func (e DeploymentError) Error() string {
    return fmt.Sprintf("deployment failed for template %s: %v", e.TemplateFile, e.Cause)
}
```

### Logging Conventions
```go
import log "github.com/sirupsen/logrus"

// Use structured logging
log.WithFields(log.Fields{
    "subscriptionID": config.SubscriptionID,
    "resourceGroup": config.ResourceGroupName,
    "template": config.TemplateFile,
}).Info("Starting ARM template deployment")

// Log levels
log.Debug("Detailed debugging information")
log.Info("General information")
log.Warn("Warning conditions")
log.Error("Error conditions")
```

## Integration Patterns for IaC

### Command-Line Integration
```go
// Standard command pattern for IaC tools
func ExecuteMPF(deploymentType string, config MPFConfig) (*PermissionResult, error) {
    cmd := exec.Command("azmpf", deploymentType,
        "--subscriptionID", config.SubscriptionID,
        "--tenantID", config.TenantID,
        "--spClientID", config.SPClientID,
        "--spObjectID", config.SPObjectID,
        "--spClientSecret", config.SPClientSecret,
        "--jsonOutput")
    
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("mpf execution failed: %w", err)
    }
    
    var result PermissionResult
    if err := json.Unmarshal(output, &result); err != nil {
        return nil, fmt.Errorf("failed to parse MPF output: %w", err)
    }
    
    return &result, nil
}
```

### JSON Output Structure
```go
// Expected JSON output format for automation
type MPFResult struct {
    Permissions []Permission `json:"permissions"`
    Summary     Summary      `json:"summary"`
    Metadata    Metadata     `json:"metadata"`
}

type Summary struct {
    TotalPermissions int    `json:"totalPermissions"`
    UniqueActions    int    `json:"uniqueActions"`
    ScopeTypes       []string `json:"scopeTypes"`
}
```

## Security Considerations

### Secret Management
```go
// Never log sensitive values
log.WithFields(log.Fields{
    "clientID": config.SPClientID,
    "objectID": config.SPObjectID,
    // DO NOT LOG: clientSecret, accessTokens
}).Debug("Service principal configuration")

// Use environment variables for secrets in automation
func LoadConfigFromEnv() (*MPFConfig, error) {
    config := &MPFConfig{
        SubscriptionID: os.Getenv("MPF_SUBSCRIPTIONID"),
        TenantID:       os.Getenv("MPF_TENANTID"),
        SPClientID:     os.Getenv("MPF_SPCLIENTID"),
        SPObjectID:     os.Getenv("MPF_SPOBJECTID"),
        SPClientSecret: os.Getenv("MPF_SPCLIENTSECRET"),
    }
    
    return config, validateConfig(config)
}
```

### Temporary Role Assignment
```go
// MPF pattern: Assign minimal roles temporarily, then cleanup
func (a *AuthorizationChecker) AnalyzeWithTemporaryRoles(ctx context.Context) error {
    // 1. Deploy with minimal permissions
    // 2. Capture permission errors
    // 3. Grant required permissions incrementally  
    // 4. Re-deploy to verify
    // 5. ALWAYS cleanup role assignments
    defer a.cleanupRoleAssignments(ctx)
    
    return a.executeAnalysis(ctx)
}
```

## Common Development Tasks

### Adding New IaC Provider Support
1. Create new command in `cmd/` (follow existing patterns)
2. Implement provider-specific logic in `pkg/infrastructure/`
3. Add domain models in `pkg/domain/` if needed
4. Create use case orchestration in `pkg/usecase/`
5. Add comprehensive tests including e2e scenarios

### Extending Permission Analysis
1. Add new error parsers in `pkg/domain/` for specific Azure error patterns
2. Implement permission completion logic for partial actions
3. Add scope-level permission grouping
4. Update JSON output format if needed

### Debugging Permission Issues
```go
// Enable debug logging for detailed tracing
export LOG_LEVEL=debug

// Common debugging patterns
log.WithFields(log.Fields{
    "action": permission.Action,
    "scope": permission.Scope, 
    "errorCode": azureError.Code,
}).Debug("Permission analysis step")
```

This guidance should help GitHub Copilot provide contextually appropriate suggestions for MPF development work, including proper Azure integration patterns, security considerations, and testing approaches.