
param clusterName string
param vnetName string
param subnetName string
param vnetAddressPrefix string = '10.13.0.0/16'
param subnetAddressPrefix string = '10.13.2.0/24'
param servicePrincipalClientId string = ''

resource vnet 'Microsoft.Network/virtualNetworks@2023-05-01' = {
  name: vnetName
  location: resourceGroup().location
  properties: {
    addressSpace: {
      addressPrefixes: [
        vnetAddressPrefix
      ]
    }
  }
}

resource subnet 'Microsoft.Network/virtualNetworks/subnets@2023-05-01' = {
  name: '${vnetName}/${subnetName}'
  location: resourceGroup().location
  dependsOn: [vnet]
  properties: {
    addressPrefix: subnetAddressPrefix
  }
}

resource aksCluster 'Microsoft.ContainerService/managedClusters@2023-05-01' = {
  name: clusterName
  location: resourceGroup().location
  dependsOn: [subnet]
  identity: {
    type: 'SystemAssigned'
  }
  properties: {
    kubernetesVersion: '1.30.0'
    dnsPrefix: clusterName
    servicePrincipalProfile: {}
    agentPoolProfiles: [
      {
        name: 'agentpool'
        count: 1
        vmSize: 'Standard_D2s_v3'
        osType: 'Linux'
        osDiskSizeGB: 30
        vnetSubnetID: subnet.id
        mode: 'System'
      }
    ]
    networkProfile: {
      networkPlugin: 'azure'
      loadBalancerSku: 'standard'
      networkPolicy: 'azure'
    }
  }
}
