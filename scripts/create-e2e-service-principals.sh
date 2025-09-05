#!/bin/bash

# Script to create service principals for E2E test parallel execution
# This script creates three separate service principals for Terraform, ARM/Bicep, and Windows E2E tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if Azure CLI and GitHub CLI are installed and logged in
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if Azure CLI is installed
    if ! command -v az &> /dev/null; then
        print_error "Azure CLI is not installed. Please install it first: https://docs.microsoft.com/en-us/cli/azure/install-azure-cli"
        exit 1
    fi
    
    # Check if logged in to Azure
    if ! az account show &> /dev/null; then
        print_error "Not logged in to Azure. Please run 'az login' first."
        exit 1
    fi
    
    # Check if GitHub CLI is installed
    if ! command -v gh &> /dev/null; then
        print_error "GitHub CLI is not installed. Please install it first: https://cli.github.com/"
        exit 1
    fi
    
    # Check if logged in to GitHub
    if ! gh auth status &> /dev/null; then
        print_error "Not logged in to GitHub. Please run 'gh auth login' first."
        exit 1
    fi
    
    # Check if we're in a GitHub repository
    if ! gh repo view &> /dev/null; then
        print_error "Not in a GitHub repository or no remote configured."
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Function to create a service principal
create_service_principal() {
    local sp_name=$1
    local sp_display_name=$2
    
    print_status "Creating service principal: $sp_display_name"
    
    # Create the service principal
    local sp_output=$(az ad sp create-for-rbac \
        --name "$sp_name" \
        --skip-assignment \
        --output json)
    
    if [ $? -eq 0 ]; then
        print_success "Created service principal: $sp_display_name"
        
        # Extract information
        local app_id=$(echo "$sp_output" | jq -r '.appId')
        local password=$(echo "$sp_output" | jq -r '.password')
        local tenant=$(echo "$sp_output" | jq -r '.tenant')
        
        # Get the object ID
        local object_id=$(az ad sp show --id "$app_id" --query "id" --output tsv)
        
        # Store in associative array (using global variables for simplicity)
        case "$sp_name" in
            "mpf-terraform-e2e-sp")
                TERRAFORM_CLIENT_ID="$app_id"
                TERRAFORM_CLIENT_SECRET="$password"
                TERRAFORM_OBJECT_ID="$object_id"
                ;;
            "mpf-arm-bicep-e2e-sp")
                ARM_BICEP_CLIENT_ID="$app_id"
                ARM_BICEP_CLIENT_SECRET="$password"
                ARM_BICEP_OBJECT_ID="$object_id"
                ;;
            "mpf-windows-e2e-sp")
                WINDOWS_CLIENT_ID="$app_id"
                WINDOWS_CLIENT_SECRET="$password"
                WINDOWS_OBJECT_ID="$object_id"
                ;;
        esac
        
        print_status "  App ID: $app_id"
        print_status "  Object ID: $object_id"
        print_warning "  Client Secret: [HIDDEN - will be shown in summary]"
        
    else
        print_error "Failed to create service principal: $sp_display_name"
        exit 1
    fi
}

# Function to add secrets to GitHub repository
add_github_secrets() {
    print_status "Adding secrets to GitHub repository..."
    
    local repo_name=$(gh repo view --json nameWithOwner --jq '.nameWithOwner')
    print_status "Repository: $repo_name"
    # When multiple remotes exist (e.g., origin + upstream), gh secret set can fail with
    # "multiple remotes detected". We explicitly pass -R to avoid ambiguity.
    # Additionally, we export GH_REPO so any subsequent gh commands inherit the context.
    export GH_REPO="$repo_name"
    
    # Add Terraform E2E secrets
    print_status "Adding Terraform E2E secrets..."
    echo "$TERRAFORM_CLIENT_ID" | gh secret set MPF_TERRAFORM_SPCLIENTID -R "$repo_name"
    echo "$TERRAFORM_CLIENT_SECRET" | gh secret set MPF_TERRAFORM_SPCLIENTSECRET -R "$repo_name"
    echo "$TERRAFORM_OBJECT_ID" | gh secret set MPF_TERRAFORM_SPOBJECTID -R "$repo_name"
    
    # Add ARM/Bicep E2E secrets
    print_status "Adding ARM/Bicep E2E secrets..."
    echo "$ARM_BICEP_CLIENT_ID" | gh secret set MPF_ARM_BICEP_SPCLIENTID -R "$repo_name"
    echo "$ARM_BICEP_CLIENT_SECRET" | gh secret set MPF_ARM_BICEP_SPCLIENTSECRET -R "$repo_name"
    echo "$ARM_BICEP_OBJECT_ID" | gh secret set MPF_ARM_BICEP_SPOBJECTID -R "$repo_name"
    
    # Add Windows E2E secrets
    print_status "Adding Windows E2E secrets..."
    echo "$WINDOWS_CLIENT_ID" | gh secret set MPF_WINDOWS_SPCLIENTID -R "$repo_name"
    echo "$WINDOWS_CLIENT_SECRET" | gh secret set MPF_WINDOWS_SPCLIENTSECRET -R "$repo_name"
    echo "$WINDOWS_OBJECT_ID" | gh secret set MPF_WINDOWS_SPOBJECTID -R "$repo_name"
    
    print_success "All secrets added to GitHub repository!"
}

# Function to display GitHub secrets configuration
display_github_secrets() {
    print_status "GitHub Repository Secrets Configuration"
    echo ""
    echo "Add the following secrets to your GitHub repository:"
    echo ""
    
    echo -e "${YELLOW}Terraform E2E Secrets:${NC}"
    echo "MPF_TERRAFORM_SPCLIENTID = $TERRAFORM_CLIENT_ID"
    echo "MPF_TERRAFORM_SPCLIENTSECRET = $TERRAFORM_CLIENT_SECRET"
    echo "MPF_TERRAFORM_SPOBJECTID = $TERRAFORM_OBJECT_ID"
    echo ""
    
    echo -e "${YELLOW}ARM/Bicep E2E Secrets:${NC}"
    echo "MPF_ARM_BICEP_SPCLIENTID = $ARM_BICEP_CLIENT_ID"
    echo "MPF_ARM_BICEP_SPCLIENTSECRET = $ARM_BICEP_CLIENT_SECRET"
    echo "MPF_ARM_BICEP_SPOBJECTID = $ARM_BICEP_OBJECT_ID"
    echo ""
    
    echo -e "${YELLOW}Windows E2E Secrets:${NC}"
    echo "MPF_WINDOWS_SPCLIENTID = $WINDOWS_CLIENT_ID"
    echo "MPF_WINDOWS_SPCLIENTSECRET = $WINDOWS_CLIENT_SECRET"
    echo "MPF_WINDOWS_SPOBJECTID = $WINDOWS_OBJECT_ID"
    echo ""
}

# Function to save credentials to a file
save_credentials() {
    local creds_file="e2e-service-principals-credentials.json"
    
    print_status "Saving credentials to $creds_file"
    
    cat > "$creds_file" << EOF
{
  "terraform_e2e": {
    "client_id": "$TERRAFORM_CLIENT_ID",
    "client_secret": "$TERRAFORM_CLIENT_SECRET",
    "object_id": "$TERRAFORM_OBJECT_ID"
  },
  "arm_bicep_e2e": {
    "client_id": "$ARM_BICEP_CLIENT_ID",
    "client_secret": "$ARM_BICEP_CLIENT_SECRET",
    "object_id": "$ARM_BICEP_OBJECT_ID"
  },
  "windows_e2e": {
    "client_id": "$WINDOWS_CLIENT_ID",
    "client_secret": "$WINDOWS_CLIENT_SECRET",
    "object_id": "$WINDOWS_OBJECT_ID"
  },
  "tenant_id": "$(az account show --query tenantId --output tsv)",
  "subscription_id": "$(az account show --query id --output tsv)"
}
EOF

    print_success "Credentials saved to $creds_file"
    print_warning "IMPORTANT: Store this file securely and delete it after configuring GitHub secrets!"
}

# Main execution
main() {
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN} MPF E2E Service Principals Creator    ${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    
    # Get current subscription info
    local subscription_name=$(az account show --query name --output tsv)
    local subscription_id=$(az account show --query id --output tsv)
    local tenant_id=$(az account show --query tenantId --output tsv)
    
    print_status "Current Azure Context:"
    print_status "  Subscription: $subscription_name ($subscription_id)"
    print_status "  Tenant: $tenant_id"
    echo ""
    
    # Confirm with user
    read -p "Do you want to create service principals in this subscription? (y/N): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_status "Operation cancelled by user"
        exit 0
    fi
    
    # Ask about GitHub secrets
    echo ""
    read -p "Do you want to automatically add secrets to GitHub repository? (Y/n): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Nn]$ ]]; then
        ADD_TO_GITHUB=false
        print_status "Will not add secrets to GitHub (you can add them manually later)"
    else
        ADD_TO_GITHUB=true
        print_status "Will automatically add secrets to GitHub repository"
    fi
    
    # Check prerequisites
    check_prerequisites
    echo ""
    
    # Create service principals
    print_status "Creating three service principals for E2E tests..."
    echo ""
    
    create_service_principal "mpf-terraform-e2e-sp" "MPF Terraform E2E"
    echo ""
    
    create_service_principal "mpf-arm-bicep-e2e-sp" "MPF ARM/Bicep E2E"
    echo ""
    
    create_service_principal "mpf-windows-e2e-sp" "MPF Windows E2E"
    echo ""
    
    # Display results
    print_success "All service principals created successfully!"
    echo ""
    
    # Save credentials
    save_credentials
    echo ""
    
    # Add to GitHub if requested
    if [ "$ADD_TO_GITHUB" = true ]; then
        add_github_secrets
        echo ""
        print_success "Secrets successfully added to GitHub repository!"
        echo ""
        print_status "You can verify the secrets were added by running:"
        print_status "gh secret list"
    else
        # Display GitHub secrets configuration for manual setup
        display_github_secrets
    fi
    
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN} Setup Complete!                       ${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    print_status "Next steps:"
    if [ "$ADD_TO_GITHUB" = true ]; then
        echo "1. ✅ Secrets have been added to GitHub repository"
        echo "2. Test the updated workflows"
        echo "3. Delete the credentials file: e2e-service-principals-credentials.json"
        echo "4. Verify secrets with: gh secret list"
    else
        echo "1. Add the secrets above to your GitHub repository"
        echo "2. Test the updated workflows"
        echo "3. Delete the credentials file: e2e-service-principals-credentials.json"
    fi
    echo ""
    print_warning "Remember: These service principals have no permissions by default."
    print_warning "The MPF utility will manage all required permissions dynamically."
}

# Run the main function
main "$@"
