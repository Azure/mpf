#!/bin/bash

# Script to create service principals for E2E test isolation.
#
# This script creates dedicated service principals per scenario (terraform/arm/bicep) and per OS (linux/windows)
# so that concurrent E2E runs do not interfere with each other.

set -euo pipefail

ECOSYSTEMS=("terraform" "arm" "bicep")
OSES=("linux" "windows")

declare -A CLIENT_ID
declare -A CLIENT_SECRET
declare -A OBJECT_ID

YES=false
ADD_TO_GITHUB=""
OUTPUT_FILE="e2e-service-principals.json"

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

usage() {
  cat <<EOF
Usage: $(basename "$0") [options]

Creates (or reuses) Entra ID app/service principals for MPF E2E tests and optionally writes GitHub repo secrets.

Options:
  -y, --yes                Skip prompts (non-interactive). Defaults to NOT adding secrets to GitHub unless --add-to-github is set.
      --add-to-github      Automatically add secrets to GitHub repository.
  -o, --output-file PATH   Output JSON file path (default: e2e-service-principals.json)
  -h, --help               Show this help.
EOF
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -y|--yes)
        YES=true
        shift
        ;;
      --add-to-github)
        ADD_TO_GITHUB=true
        shift
        ;;
      -o|--output-file)
        OUTPUT_FILE="$2"
        shift 2
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      *)
        print_error "Unknown argument: $1"
        usage
        exit 2
        ;;
    esac
  done

  if [[ "${YES}" == true && -z "${ADD_TO_GITHUB}" ]]; then
    ADD_TO_GITHUB=false
  fi

  if [[ -z "${ADD_TO_GITHUB}" ]]; then
    ADD_TO_GITHUB=""
  fi
}

title_case() {
  local s="$1"
  if [[ -z "${s}" ]]; then
    echo ""
    return
  fi
  echo "${s:0:1}" | tr '[:lower:]' '[:upper:]' | tr -d '\n'
  echo "${s:1}"
}

# Function to check if Azure CLI and GitHub CLI are installed and logged in
check_prerequisites() {
  print_status "Checking prerequisites..."

  # Check if Azure CLI is installed
  if ! command -v az &>/dev/null; then
    print_error "Azure CLI is not installed. Please install it first: https://docs.microsoft.com/en-us/cli/azure/install-azure-cli"
    exit 1
  fi

  # Check if jq is installed (required for JSON parsing)
  if ! command -v jq &>/dev/null; then
    print_error "jq is not installed. Please install it first: https://jqlang.github.io/jq/"
    exit 1
  fi

  # Check if logged in to Azure
  if ! az account show &>/dev/null; then
    print_error "Not logged in to Azure. Please run 'az login' first."
    exit 1
  fi

  # Check if GitHub CLI is installed
  if ! command -v gh &>/dev/null; then
    print_error "GitHub CLI is not installed. Please install it first: https://cli.github.com/"
    exit 1
  fi

  # Check if logged in to GitHub
  if ! gh auth status &>/dev/null; then
    print_error "Not logged in to GitHub. Please run 'gh auth login' first."
    exit 1
  fi

  # Check if we're in a GitHub repository
  if ! gh repo view &>/dev/null; then
    print_error "Not in a GitHub repository or no remote configured."
    exit 1
  fi

  print_success "Prerequisites check passed"
}

create_service_principal() {
  local ecosystem="$1"
  local os="$2"

  local sp_name
  sp_name="mpf-${ecosystem}-${os}-e2e-sp"
  local display_name
  display_name="MPF $(title_case "${ecosystem}") $(title_case "${os}") E2E"

  print_status "Creating service principal: ${display_name}"

  # NOTE: `az ad sp create-for-rbac --skip-assignment` is deprecated.
  # Create an Entra ID app + service principal + client secret explicitly, without RBAC assignments.

  local app_id
  app_id=$(az ad app list --filter "displayName eq '${sp_name}'" --query '[0].appId' --output tsv 2>/dev/null || true)
  if [[ -z "${app_id}" ]]; then
    local app_output
    app_output=$(az ad app create --display-name "${sp_name}" --output json)
    app_id=$(echo "$app_output" | jq -r '.appId')
    if [[ -z "${app_id}" || "${app_id}" == "null" ]]; then
      print_error "Failed to create app registration for: ${sp_name}"
      exit 1
    fi
  else
    print_status "Reusing existing app registration: ${sp_name}"
  fi

  # Ensure a service principal exists for this app id (no-op if it already exists).
  az ad sp create --id "${app_id}" --output none 2>/dev/null || true

  local object_id
  object_id=""
  # Entra can be eventually-consistent; retry fetching object_id a few times.
  for _ in {1..20}; do
    object_id=$(az ad sp show --id "${app_id}" --query "id" --output tsv 2>/dev/null || true)
    if [[ -n "${object_id}" ]]; then
      break
    fi
    sleep 2
  done
  if [[ -z "${object_id}" ]]; then
    print_error "Failed to retrieve service principal object id for appId: ${app_id}"
    exit 1
  fi

  local cred_output
  # Reset password (rotate secret) for this app.
  cred_output=$(az ad app credential reset --id "${app_id}" --display-name "mpf-e2e" --output json)

  local password
  password=$(echo "$cred_output" | jq -r '.password')
  if [[ -z "${password}" || "${password}" == "null" ]]; then
    print_error "Failed to create client secret for appId: ${app_id}"
    exit 1
  fi

  local key
  key="${ecosystem}_${os}"
  CLIENT_ID["${key}"]="${app_id}"
  CLIENT_SECRET["${key}"]="${password}"
  OBJECT_ID["${key}"]="${object_id}"

  print_success "Created service principal: ${display_name}"
  print_status "  App ID: ${app_id}"
  print_status "  Object ID: ${object_id}"
  print_warning "  Client Secret: [HIDDEN - will be written to file / GitHub secrets]"
}

# Function to add secrets to GitHub repository
add_github_secrets() {
  print_status "Adding secrets to GitHub repository..."

  local repo_name
  repo_name=$(gh repo view --json nameWithOwner --jq '.nameWithOwner')
  print_status "Repository: ${repo_name}"
  # When multiple remotes exist (e.g., origin + upstream), gh secret set can fail with
  # "multiple remotes detected". We explicitly pass -R to avoid ambiguity.
  # Additionally, we export GH_REPO so any subsequent gh commands inherit the context.
  export GH_REPO="${repo_name}"

  for ecosystem in "${ECOSYSTEMS[@]}"; do
    for os in "${OSES[@]}"; do
      local key
      key="${ecosystem}_${os}"
      local prefix
      prefix="MPF_${ecosystem^^}_${os^^}"

      print_status "Setting secrets for ${ecosystem} ${os}..."
      echo "${CLIENT_ID[$key]}" | gh secret set "${prefix}_SPCLIENTID" -R "${repo_name}"
      echo "${CLIENT_SECRET[$key]}" | gh secret set "${prefix}_SPCLIENTSECRET" -R "${repo_name}"
      echo "${OBJECT_ID[$key]}" | gh secret set "${prefix}_SPOBJECTID" -R "${repo_name}"
    done
  done

  print_status "Setting shared secrets (tenant/subscription)..."
  echo "${TENANT_ID}" | gh secret set MPF_TENANTID -R "${repo_name}"
  echo "${SUBSCRIPTION_ID}" | gh secret set MPF_SUBSCRIPTIONID -R "${repo_name}"

  print_success "All secrets added to GitHub repository!"
}

# Function to display GitHub secrets configuration
display_github_secrets() {
  print_status "GitHub Repository Secrets Configuration"
  echo ""
  echo "Add the following secrets to your GitHub repository:"
  echo ""

  for ecosystem in "${ECOSYSTEMS[@]}"; do
    for os in "${OSES[@]}"; do
      local key
      key="${ecosystem}_${os}"
      local prefix
      prefix="MPF_${ecosystem^^}_${os^^}"

      # Mask client secret to avoid leaking it via logs
      local client_secret="${CLIENT_SECRET[$key]}"
      local client_secret_display
      if [[ ${#client_secret} -gt 8 ]]; then
        client_secret_display="${client_secret:0:4}...${client_secret: -4}"
      else
        client_secret_display="<redacted>"
      fi

      echo "${prefix}_SPCLIENTID = ${CLIENT_ID[$key]}"
      echo "${prefix}_SPCLIENTSECRET = ${client_secret_display} (full value stored securely in ${OUTPUT_FILE})"
      echo "${prefix}_SPOBJECTID = ${OBJECT_ID[$key]}"
      echo ""
    done
  done

  echo ""
  echo "MPF_TENANTID = ${TENANT_ID}"
  echo "MPF_SUBSCRIPTIONID = ${SUBSCRIPTION_ID}"
}

# Function to save credentials to a file
save_credentials() {
  print_status "Saving credentials to ${OUTPUT_FILE}"

  {
    for ecosystem in "${ECOSYSTEMS[@]}"; do
      for os in "${OSES[@]}"; do
        local key
        key="${ecosystem}_${os}"
        printf "%s\t%s\t%s\t%s\n" \
          "${key}" \
          "${CLIENT_ID[$key]}" \
          "${CLIENT_SECRET[$key]}" \
          "${OBJECT_ID[$key]}"
      done
    done
  } | jq -Rn --arg tenant "${TENANT_ID}" --arg sub "${SUBSCRIPTION_ID}" '
    reduce inputs as $line ({};
      ($line | split("\t")) as [$k, $id, $sec, $oid]
      | .[$k] = { client_id: $id, client_secret: $sec, object_id: $oid }
    )
    | .tenant_id = $tenant
    | .subscription_id = $sub
  ' >"${OUTPUT_FILE}"

  print_success "Credentials saved to ${OUTPUT_FILE}"
  print_warning "IMPORTANT: Store this file securely and delete it after configuring GitHub secrets!"
}

# Main execution
main() {
  parse_args "$@"

  echo -e "${GREEN}========================================${NC}"
  echo -e "${GREEN} MPF E2E Service Principals Creator     ${NC}"
  echo -e "${GREEN}========================================${NC}"
  echo ""

  # Check prerequisites early for parity with PowerShell script.
  check_prerequisites
  echo ""

  # Get current subscription info
  local subscription_name
  subscription_name=$(az account show --query name --output tsv)
  local subscription_id
  subscription_id=$(az account show --query id --output tsv)
  local tenant_id
  tenant_id=$(az account show --query tenantId --output tsv)

  # Export shared values for other functions (GitHub secret setting + display).
  TENANT_ID="${tenant_id}"
  SUBSCRIPTION_ID="${subscription_id}"

  print_status "Current Azure Context:"
  print_status "  Subscription: ${subscription_name} (${subscription_id})"
  print_status "  Tenant: ${tenant_id}"
  echo ""

  if [[ "${YES}" != true ]]; then
    # Confirm with user
    read -p "Do you want to create service principals in this subscription? (Y/n): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Nn]$ ]]; then
      print_status "Operation cancelled by user"
      exit 0
    fi

    # Ask about GitHub secrets if not decided by args
    if [[ -z "${ADD_TO_GITHUB}" ]]; then
      echo ""
      read -p "Do you want to automatically add secrets to GitHub repository? (y/N): " -n 1 -r
      echo ""
      if [[ $REPLY =~ ^[Yy]$ ]]; then
        ADD_TO_GITHUB=true
        print_status "Will automatically add secrets to GitHub repository"
      else
        ADD_TO_GITHUB=false
        print_status "Will not add secrets to GitHub (you can add them manually later)"
      fi
    fi
  fi

  if [[ -z "${ADD_TO_GITHUB}" ]]; then
    ADD_TO_GITHUB=false
  fi

  # Create service principals
  print_status "Creating service principals for E2E tests (${#ECOSYSTEMS[@]} ecosystems × ${#OSES[@]} OSes)..."
  echo ""

  for ecosystem in "${ECOSYSTEMS[@]}"; do
    for os in "${OSES[@]}"; do
      create_service_principal "${ecosystem}" "${os}"
      echo ""
    done
  done

  # Display results
  print_success "All service principals created successfully!"
  echo ""

  save_credentials
  echo ""

  # Add to GitHub if requested
  if [[ "${ADD_TO_GITHUB}" = true ]]; then
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
  if [[ "${ADD_TO_GITHUB}" = true ]]; then
    echo "1. ✅ Secrets have been added to GitHub repository"
    echo "2. Test the updated workflows"
    echo "3. Delete the credentials file: ${OUTPUT_FILE}"
    echo "4. Verify secrets with: gh secret list"
  else
    echo "1. Add the secrets above to your GitHub repository"
    echo "2. Test the updated workflows"
    echo "3. Delete the credentials file: ${OUTPUT_FILE}"
  fi
  echo ""
  print_warning "Remember: These service principals have no permissions by default."
  print_warning "The MPF utility will manage all required permissions dynamically."
}

# Run the main function
main "$@"
