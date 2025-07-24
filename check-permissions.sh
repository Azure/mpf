#!/bin/bash

# MIT License
# 
# Copyright (c) Microsoft Corporation.
# 
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
# 
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
# 
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE

# Script to check if the current user has the required permissions to run MPF
# This script verifies that the current Azure CLI user has all the permissions
# required by MPF to create custom roles and role assignments.

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Required permissions for MPF
REQUIRED_PERMISSIONS=(
    "Microsoft.Authorization/roleDefinitions/read"
    "Microsoft.Authorization/roleDefinitions/write"
    "Microsoft.Authorization/roleDefinitions/delete"
    "Microsoft.Authorization/roleAssignments/read"
    "Microsoft.Authorization/roleAssignments/write"
    "Microsoft.Authorization/roleAssignments/delete"
    "Microsoft.Resources/subscriptions/resourcegroups/delete"
    "Microsoft.Resources/subscriptions/resourcegroups/read"
    "Microsoft.Resources/subscriptions/resourcegroups/write"
)

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to check if Azure CLI is installed and user is logged in
check_azure_cli() {
    print_status $BLUE "Checking Azure CLI setup..."
    
    if ! command -v az &> /dev/null; then
        print_status $RED "❌ Azure CLI is not installed. Please install Azure CLI and try again."
        echo "   Installation instructions: https://docs.microsoft.com/en-us/cli/azure/install-azure-cli"
        exit 1
    fi
    
    if ! az account show &> /dev/null; then
        print_status $RED "❌ You are not logged in to Azure CLI. Please run 'az login' and try again."
        exit 1
    fi
    
    print_status $GREEN "✅ Azure CLI is installed and you are logged in."
}

# Function to get current subscription info
get_subscription_info() {
    local subscription_info
    subscription_info=$(az account show --query '{subscriptionId:id, subscriptionName:name, tenantId:tenantId}' -o json 2>/dev/null)
    
    if [ $? -ne 0 ]; then
        print_status $RED "❌ Failed to get current subscription information."
        exit 1
    fi
    
    echo "$subscription_info"
}

# Function to check permissions
check_permissions() {
    local subscription_id=$1
    local missing_permissions=()
    local has_all_permissions=true
    
    print_status $BLUE "Checking permissions for subscription: $subscription_id"
    echo
    
    # Get current user's role assignments at subscription scope
    local role_assignments
    role_assignments=$(az role assignment list --scope "/subscriptions/$subscription_id" --include-inherited --query '[].roleDefinitionName' -o tsv 2>/dev/null)
    
    if [ $? -ne 0 ]; then
        print_status $RED "❌ Failed to get role assignments. Please check your Azure CLI permissions."
        return 1
    fi
    
    # Check if user has Owner or Contributor + User Access Administrator roles
    if echo "$role_assignments" | grep -q "Owner"; then
        print_status $GREEN "✅ You have the 'Owner' role - all permissions are available!"
        return 0
    fi
    
    local has_contributor=false
    local has_user_access_admin=false
    
    if echo "$role_assignments" | grep -q "Contributor"; then
        has_contributor=true
    fi
    
    if echo "$role_assignments" | grep -q "User Access Administrator"; then
        has_user_access_admin=true
    fi
    
    if [ "$has_contributor" = "true" ] && [ "$has_user_access_admin" = "true" ]; then
        print_status $GREEN "✅ You have 'Contributor' and 'User Access Administrator' roles - all permissions are available!"
        return 0
    fi
    
    # If not Owner or Contributor+UAA, do detailed permission check
    print_status $YELLOW "Performing detailed permission check..."
    echo
    
    for permission in "${REQUIRED_PERMISSIONS[@]}"; do
        printf "%-60s" "Checking: $permission"
        
        # Try to perform a test operation to check the permission
        local has_permission=false
        
        case "$permission" in
            "Microsoft.Authorization/roleDefinitions/read")
                az role definition list --name "Reader" --scope "/subscriptions/$subscription_id" &>/dev/null && has_permission=true
                ;;
            "Microsoft.Authorization/roleDefinitions/write"|"Microsoft.Authorization/roleDefinitions/delete")
                # These are harder to test without actually creating/deleting, so we check role assignments
                if [ "$has_user_access_admin" = "true" ] || echo "$role_assignments" | grep -Eq "(Owner|User Access Administrator)"; then
                    has_permission=true
                fi
                ;;
            "Microsoft.Authorization/roleAssignments/read")
                az role assignment list --scope "/subscriptions/$subscription_id" --query '[0].id' -o tsv &>/dev/null && has_permission=true
                ;;
            "Microsoft.Authorization/roleAssignments/write"|"Microsoft.Authorization/roleAssignments/delete")
                if [ "$has_user_access_admin" = "true" ] || echo "$role_assignments" | grep -Eq "(Owner|User Access Administrator)"; then
                    has_permission=true
                fi
                ;;
            "Microsoft.Resources/subscriptions/resourcegroups/read")
                az group list --query '[0].name' -o tsv &>/dev/null && has_permission=true
                ;;
            "Microsoft.Resources/subscriptions/resourcegroups/write"|"Microsoft.Resources/subscriptions/resourcegroups/delete")
                if [ "$has_contributor" = "true" ] || echo "$role_assignments" | grep -Eq "(Owner|Contributor)"; then
                    has_permission=true
                fi
                ;;
        esac
        
        if [ "$has_permission" = "true" ]; then
            print_status $GREEN "✅"
        else
            print_status $RED "❌"
            missing_permissions+=("$permission")
            has_all_permissions=false
        fi
    done
    
    echo
    
    if [ "$has_all_permissions" = "true" ]; then
        print_status $GREEN "🎉 All required permissions are available! You can run MPF."
    else
        print_status $RED "❌ Missing permissions detected. You may encounter issues running MPF."
        echo
        print_status $YELLOW "Missing permissions:"
        for permission in "${missing_permissions[@]}"; do
            echo "  - $permission"
        done
        echo
        print_status $YELLOW "To resolve this, you need one of the following role combinations:"
        echo "  - Owner (recommended)"
        echo "  - Contributor + User Access Administrator"
        echo "  - Custom role with all the above permissions"
        echo
        print_status $YELLOW "Your current roles:"
        if [ -n "$role_assignments" ]; then
            echo "$role_assignments" | sed 's/^/  - /'
        else
            echo "  - No roles found"
        fi
        echo
        print_status $YELLOW "Contact your Azure administrator to assign the appropriate roles."
        return 1
    fi
}

# Function to show help
show_help() {
    echo "MPF Permissions Checker"
    echo
    echo "This script checks if the current Azure CLI user has all the permissions"
    echo "required to run the MPF (Minimum Permissions Finder) utility."
    echo
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Options:"
    echo "  -s, --subscription-id ID    Check permissions for specific subscription ID"
    echo "  -h, --help                  Show this help message"
    echo
    echo "If no subscription ID is provided, the script will use the current"
    echo "default subscription from Azure CLI."
    echo
    echo "Required permissions:"
    for permission in "${REQUIRED_PERMISSIONS[@]}"; do
        echo "  - $permission"
    done
}

# Main function
main() {
    local subscription_id=""
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -s|--subscription-id)
                subscription_id="$2"
                shift 2
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                print_status $RED "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    print_status $BLUE "MPF Permissions Checker"
    print_status $BLUE "======================="
    echo
    
    # Check Azure CLI setup
    check_azure_cli
    echo
    
    # Get subscription info
    if [ -z "$subscription_id" ]; then
        local sub_info
        sub_info=$(get_subscription_info)
        subscription_id=$(echo "$sub_info" | jq -r '.subscriptionId')
        local subscription_name
        subscription_name=$(echo "$sub_info" | jq -r '.subscriptionName')
        local tenant_id
        tenant_id=$(echo "$sub_info" | jq -r '.tenantId')
        
        print_status $BLUE "Using current subscription:"
        echo "  Name: $subscription_name"
        echo "  ID: $subscription_id"
        echo "  Tenant: $tenant_id"
        echo
    else
        print_status $BLUE "Using specified subscription: $subscription_id"
        echo
    fi
    
    # Check permissions
    if check_permissions "$subscription_id"; then
        exit 0
    else
        exit 1
    fi
}

# Run main function with all arguments
main "$@"