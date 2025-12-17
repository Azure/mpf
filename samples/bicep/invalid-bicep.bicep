param location string
param clusterName string = 'myAKSCluster'
param vnetName string = 'myVNet'
param subnetName string = 'mySubnet'

resource vnet 'Microsoft.Network/virtualNetworks@2021-02-01' = {
  // names and locations are incorrect, this file should result in invalidTemplate error
  names: vnetName
  locations: location
  properties: {
    addressSpace: {
      addressPrefixes: [
        '10.0.0.0/16'
      ]
    }
  }
}

resource subnet 'Microsoft.Network/virtualNetworks/subnets@2021-02-01' = {
  parent: vnet
  name: subnetName
  properties: {
    addressPrefix: '10.0.0.0/24'
    privateEndpointNetworkPolicies: 'Disabled'
    privateLinkServiceNetworkPolicies: 'Disabled'
  }
}

resource aksCluster 'Microsoft.ContainerService/managedClusters@2021-07-01' = {
  name: clusterName
  location: location
  properties: {
    kubernetesVersion: '1.21.2'
    dnsPrefix: clusterName
    enableRBAC: true
    networkProfile: {
      networkPlugin: 'azure'
      networkMode: 'private'
      loadBalancerSku: 'standard'
      networkPolicy: 'calico'
      podCidr: '10.244.0.0/16'
      serviceCidr: '10.245.0.0/16'
      dockerBridgeCidr: '172.17.0.1/16'
      outboundType: 'loadBalancer'
      loadBalancerProfile: {
        managedOutboundIPs: {
          count: 1
        }
      }
    }
    agentPoolProfiles: [
      {
        name: 'agentpool'
        count: 1
        vmSize: 'Standard_D2s_v6'
        osType: 'Linux'
        osDiskSizeGB: 30
        vnetSubnetID: subnet.id
      }
    ]
  }
}
