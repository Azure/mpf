# MPF (Azure Deployment Minimum Permissions Finder) - Copilot Instructions

## Repository Overview

**MPF** is a Go-based CLI utility that finds the minimum Azure RBAC permissions required for deploying ARM templates, Bicep files, or Terraform modules. It works by iteratively attempting deployments with a test Service Principal, parsing authorization errors, and incrementally adding permissions until the deployment succeeds. The tool supports ARM (Azure Resource Manager), Bicep, and Terraform deployment providers.

## High-Level Repository Details

- **Language**: Go 1.25.5
- **Repository Size**: ~17MB, 54 Go files, ~2,300+ lines of production Go code
- **Project Type**: Command-line utility (CLI)
- **Binary Name**: `azmpf`
- **Main Package**: `./cmd`
- **Architecture**: Clean architecture with domain, infrastructure, usecase, and presentation layers
- **Build Tools**: Make, Task (go-task), GoReleaser
- **Testing**: Unit tests, E2E tests (ARM, Bicep, Terraform), CLI tests
- **Dependencies**: Azure SDK for Go, Cobra (CLI framework), Terraform exec, Viper (config)

## Building and Testing

### Prerequisites

- Go 1.25.5 (specified in go.mod)
- For Task-based builds: Install Task from https://taskfile.dev
- For E2E tests: Azure CLI, Bicep CLI, Terraform, Service Principal credentials

### Build Commands

**Using Make (always works, no additional dependencies):**

```bash
# Clean build artifacts
make clean
# Note: make clean has a bug - it fails with "refusing to remove '.' or '..' directory"
# This error is harmless and can be ignored. The build still works.

# Build the binary (outputs ./azmpf)
make build

# Build all platforms (darwin-arm64, darwin-amd64, linux-amd64, windows-amd64)
make build-all
```

**Using Task (preferred in CI, requires task to be installed):**

```bash
# Download dependencies
task deps:download

# Build using goreleaser (outputs to bin/<os>-<arch>/azmpf)
task build:mpf

# Build all platforms with goreleaser snapshot
task build:all
```

**Direct Go commands:**

```bash
# Simple build
go build -o azmpf ./cmd

# Build all packages
go build -v ./...
```

### Testing Commands

**Unit Tests:**

```bash
# Using Make (recommended - matches what's tested in CI)
make test

# Using Task (requires test tools installation first)
task test:tools  # Install gotestsum, gocover-cobertura
task testunit
```

The unit tests cover these packages:
- `./pkg/domain`
- `./pkg/infrastructure/ARMTemplateShared`
- `./pkg/infrastructure/mpfSharedUtils`
- `./pkg/infrastructure/authorizationCheckers/terraform`

**E2E Tests (require Azure credentials):**

Set these environment variables first:
```bash
export MPF_SUBSCRIPTIONID=<your-subscription-id>
export MPF_TENANTID=<your-tenant-id>
export MPF_SPCLIENTID=<service-principal-client-id>
export MPF_SPCLIENTSECRET=<service-principal-secret>
export MPF_SPOBJECTID=<service-principal-object-id>
```

For Bicep tests, also set:
```bash
export MPF_BICEPEXECPATH="/usr/local/bin/bicep"
```

For Terraform tests, also set:
```bash
export MPF_TFPATH=$(which terraform)
```

Then run:
```bash
# ARM E2E tests (timeout: 45m)
make test-e2e-arm
# or: task teste2e:arm

# Bicep E2E tests (timeout: 45m)
make test-e2e-bicep
# or: task teste2e:bicep

# Terraform E2E tests (timeout: 45m, can take very long)
make test-e2e-terraform
# or: task teste2e:terraform
```

**CLI Tests:**

```bash
task testcli:arm      # Tests ARM template processing
task testcli:bicep    # Tests Bicep file processing
task testcli:terraform # Tests Terraform module processing
```

### Linting

**Using Task:**

```bash
# Install linters
task install:govulncheck
task install:golangci-lint
task install:markdownlint

# Run Go linters
task lint:go
task golangci-lint
task govulncheck

# Run markdown linters
task lint:md

# Run all linters
task lint  # Also runs terraform and file linters
```

**Using golangci-lint directly:**

```bash
golangci-lint run --fix ./...
```

## Project Structure and Architecture

### Directory Layout

```
mpf/
├── cmd/                    # CLI entry point and command definitions
│   ├── main.go            # Main entry point with version info
│   ├── rootCmd.go         # Root command setup
│   ├── armCmd.go          # ARM template command
│   ├── bicepCmd.go        # Bicep command
│   └── terraformCmd.go    # Terraform command
├── pkg/
│   ├── domain/            # Core business logic, error parsers, models
│   ├── infrastructure/    # External integrations (Azure API, ARM, Terraform)
│   │   ├── ARMTemplateShared/
│   │   ├── authorizationCheckers/  # ARM, Bicep, Terraform checkers
│   │   ├── azureAPI/
│   │   ├── mpfSharedUtils/
│   │   ├── resourceGroupManager/
│   │   └── spRoleAssignmentManager/
│   ├── presentation/      # Output formatting
│   └── usecase/          # Application orchestration
├── e2eTests/             # End-to-end tests for ARM, Bicep, Terraform
├── samples/              # Sample templates for testing
│   ├── bicep/
│   ├── templates/        # ARM templates
│   └── terraform/
├── docs/                 # Documentation
├── scripts/              # Helper scripts (e.g., create-e2e-service-principals.sh)
├── .github/
│   ├── workflows/        # CI/CD pipelines
│   └── linters/          # Linter configurations
├── Makefile              # Make-based build automation
├── Taskfile.yml          # Task-based build automation (preferred in CI)
├── go.mod / go.sum       # Go dependencies
└── .goreleaser.yml       # Release configuration
```

### Key Files

- **go.mod**: Declares Go 1.25.5, Azure SDK dependencies
- **Makefile**: Build, test, clean commands (has known issue with clean target)
- **Taskfile.yml**: More comprehensive build automation with linting, testing
- **.editorconfig**: Code style (tabs for .go files, 2 spaces for yaml/json)
- **.gitignore**: Excludes azmpf binaries, .terraform/, coverage files, bin/, dist/

### Configuration Files

- **Linting**: `.github/linters/.markdownlint-cli2.yaml`, `.github/linters/.lychee.toml`
- **Editor**: `.editorconfig` (tabs for Go, 2 spaces for YAML/JSON/etc)
- **DevContainer**: `.devcontainer/devcontainer.json` (Ubuntu 24.04, Go, Azure CLI, Bicep, Terraform, Task)
- **Go modules**: `go.mod` uses Go 1.25.5 (auto-downloads on first use)

## CI/CD Workflows

### Main CI Pipeline: `mpf-ci.yml`

Triggered on: Pull requests to main, workflow_dispatch

Jobs:
1. **lintGo**: Runs on ubuntu-24.04
   - Setup Go, Task
   - `task deps`
   - `task install:govulncheck && task govulncheck` (warnings only, doesn't fail)
   - golangci-lint (only new issues)

2. **linkMarkdowns**: Runs on ubuntu-24.04
   - `task install:markdownlint`
   - `task lint:md`

3. **build**: Runs on ubuntu-24.04
   - `task deps:download`
   - `task build:mpf`

4. **test**: Runs on ubuntu-24.04
   - `task deps:download`
   - `task test:tools`
   - `task testunit`
   - Uploads test results and coverage

### E2E Pipelines

- **arm-bicep-e2e-v2.yaml**: Scheduled nightly (21:00 UTC), merge groups, workflow_dispatch
  - Runs ARM and Bicep E2E tests + CLI tests
  - Requires Azure federated credentials
  - Timeout: 45 minutes

- **terraform-e2e-v2.yaml**: Similar to ARM/Bicep but for Terraform
  - Can take very long due to Terraform resource-by-resource API calls

- **windows-e2e-PR-v2.yaml**: Windows-specific E2E tests

- **az-mpf-ci.yaml**: Legacy CI (push to any branch)
  - Uses `go get ./...`, `go build -v ./...`, `make test`

## Known Issues and Workarounds

### Build Issues

1. **Makefile clean target error**:
   - **Issue**: `make clean` fails with "refusing to remove '.' or '..' directory"
   - **Cause**: Line 56 in Makefile has `@rm -rf $(OUTPUT_DIR)` where `OUTPUT_DIR = .`
   - **Workaround**: Ignore the error; it's harmless. The build still succeeds.
   - **Fix**: Don't use `make clean` or manually fix the Makefile to not delete `.`

2. **Task not installed**:
   - **Issue**: `task` commands fail if Task isn't installed
   - **Workaround**: Use `make` commands instead, or install Task via https://taskfile.dev
   - CI workflows install Task using `arduino/setup-task` action

3. **Go version auto-download**:
   - Go 1.25.5 will auto-download if not present (takes ~30 seconds first time)

### Test Issues

1. **E2E tests require credentials**: E2E tests will fail without proper Azure Service Principal credentials set in environment variables.

2. **E2E timeout**: E2E tests have 45-minute timeout; Terraform tests can approach this limit.

### Code Style

- Go files use **tabs** for indentation (.editorconfig enforces this)
- YAML/JSON files use **2 spaces**
- All files should have final newline and no trailing whitespace

## Validation Steps

Before submitting changes:

1. **Build validation**:
   ```bash
   make build
   ./azmpf --help  # Verify binary works
   ```

2. **Unit tests** (required):
   ```bash
   make test
   ```

3. **Linting** (if changing Go code):
   ```bash
   golangci-lint run ./...
   ```

4. **Linting** (if changing Markdown):
   ```bash
   task install:markdownlint
   task lint:md
   ```

5. **Check git status**: Ensure no unintended files are staged (e.g., azmpf binary should be gitignored)

## Trust These Instructions

**These instructions have been validated by running the commands and inspecting the repository structure.** If something doesn't work as described, check for:
- Missing environment variables (for E2E tests)
- Task not installed (use Make commands instead)
- The known `make clean` issue (safe to ignore)

Only search for additional information if these instructions are incomplete or you encounter errors not mentioned here.
