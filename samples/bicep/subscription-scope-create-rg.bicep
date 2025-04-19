targetScope = 'subscription'

@description('Generates a stable unique string for the resource group name')
var randomString = uniqueString(subscription().id, 'resourceGroupNameSeed')

@description('Resource group name')
var resourceGroupName = 'rg-${take(randomString, 13)}'

@description('location of the resource group')
var location = 'eastus2'

// Module to deploy the resource group at the subscription scope
resource coreResourceGroup 'Microsoft.Resources/resourceGroups@2021-04-01' = {
    name: resourceGroupName
    location: location
  }
