// ============================================================================
// Simple Azure Storage Account Bicep Template
// ============================================================================
// This template demonstrates a basic, best-practice storage account deployment.
// It's designed to be easy to understand while showing security-first principles.

// Parameters: Values that users provide at deployment time
// ============================================================================

// Required: The name of the storage account (must be globally unique, 3-24 chars, lowercase)
param storageAccountName string

// Optional: The Azure region where the storage account will be deployed
// Default: Uses the same location as the resource group where this is deployed
param location string = resourceGroup().location

// Optional: The storage account SKU (performance/redundancy level)
// Default: Standard_LRS = Locally Redundant Storage (least expensive)
// Other options: Standard_GRS, Standard_RAGRS, Premium_LRS, etc.
param storageAccountType string = 'Standard_LRS'

// Resources: The actual Azure resources to create
// ============================================================================

// Create an Azure Storage Account
// Using API version 2023-01-01 for the latest stable features
resource storageAccount 'Microsoft.Storage/storageAccounts@2023-01-01' = {
  name: storageAccountName
  location: location
  
  // kind: StorageV2 supports all modern storage features (blobs, queues, tables, files)
  kind: 'StorageV2'
  
  // sku: Defines performance tier and redundancy level
  sku: {
    name: storageAccountType
  }
  
  // properties: Configure storage account behavior and security
  properties: {
    // accessTier: Hot storage provides low latency (good for frequently accessed data)
    // Alternative: Cool tier for infrequent access (lower cost, higher access charges)
    accessTier: 'Hot'
    
    // Security: Disable public blob access to prevent accidental data exposure
    // If you need public access, change to true and configure specific containers
    allowBlobPublicAccess: false
    
    // Security: Require TLS 1.2 or higher for all connections (enforces encryption)
    // Prevents connections from older clients using weaker TLS versions
    minimumTlsVersion: 'TLS1_2'
  }
}

// Outputs: Values returned after successful deployment
// ============================================================================

// Return the full resource ID so other templates or scripts can reference this storage account
output storageAccountId string = storageAccount.id

// Return just the storage account name for convenient reference
output storageAccountName string = storageAccount.name
