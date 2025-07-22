## MPF Display Options

### Deployment Mode for ARM and Bicep

ARM and Bicep templates now use **Full Deployment mode** exclusively with Incremental deployment mode, which actually deploys resources to determine permissions and then cleans them up automatically. This provides comprehensive and accurate permission detection but takes significantly longer than the previous what-if mode - expect execution times of several minutes to longer depending on template complexity and the resources being deployed.

**Note**: The previous what-if analysis mode (which completed in ~90 seconds) has been deprecated due to incomplete permission detection in some scenarios, particularly for complex resource configurations and certain Azure services.

### View Terraform logs

In the case of Terraform, using --verbose or --debug will also show the Terraform logs.

### View Detailed Output

The utility by default prints the overall required permissions. It is also possible to view additional details of the permissions at an individual resource level by using the **showDetailedOutput** flag. The following is a sample of detailed output:

```shell
# Detailed Text output
export MPF_SUBSCRIPTIONID=YOUR_SUBSCRIPTION_ID
export MPF_TENANTID=YOUR_TENANT_ID
export MPF_SPCLIENTID=YOUR_SP_CLIENT_ID
export MPF_SPCLIENTSECRET=YOUR_SP_CLIENT_SECRET
export MPF_SPOBJECTID=YOUR_SP_OBJECT_ID

$ ./azmpf arm --templateFilePath ./samples/templates/aks-private-subnet.json --parametersFilePath ./samples/templates/aks-private-subnet-parameters.json --showDetailedOutput

------------------------------------------------------------------------------------------------------------------------------------------
Permissions Required:
------------------------------------------------------------------------------------------------------------------------------------------
Microsoft.ContainerService/managedClusters/read
Microsoft.ContainerService/managedClusters/write
Microsoft.Network/virtualNetworks/read
Microsoft.Network/virtualNetworks/subnets/read
Microsoft.Network/virtualNetworks/subnets/write
Microsoft.Network/virtualNetworks/write
Microsoft.Resources/deployments/read
Microsoft.Resources/deployments/write
------------------------------------------------------------------------------------------------------------------------------------------


Breakdown of permissions by different resource types:

Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-9uOhnnS/providers/Microsoft.Network/virtualNetworks/azmpfakstestvnet: 
Microsoft.Network/virtualNetworks/read
Microsoft.Network/virtualNetworks/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-9uOhnnS/providers/Microsoft.Network/virtualNetworks/azmpfakstestvnet/subnets/azmpfakstestsubnet: 
Microsoft.Network/virtualNetworks/subnets/read
Microsoft.Network/virtualNetworks/subnets/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-9uOhnnS/providers/Microsoft.ContainerService/managedClusters/azmpfakstestcluster: 
Microsoft.ContainerService/managedClusters/read
Microsoft.ContainerService/managedClusters/write
--------------


### JSON Output which by default shows the details as well

It is possible to also get the JSON output, which by default shows details as well. The following is a sample of JSON output:

```shell
export MPF_SUBSCRIPTIONID=YOUR_SUBSCRIPTION_ID
export MPF_TENANTID=YOUR_TENANT_ID
export MPF_SPCLIENTID=YOUR_SP_CLIENT_ID
export MPF_SPCLIENTSECRET=YOUR_SP_CLIENT_SECRET
export MPF_SPOBJECTID=YOUR_SP_OBJECT_ID

$ ./azmpf arm --templateFilePath ./samples/templates/aks-private-subnet.json --parametersFilePath ./samples/templates/aks-private-subnet-parameters.json --jsonOutput

{
  "/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X44": [
    "Microsoft.ContainerService/managedClusters/read",
    "Microsoft.ContainerService/managedClusters/write",
    "Microsoft.Network/virtualNetworks/read",
    "Microsoft.Network/virtualNetworks/subnets/read",
    "Microsoft.Network/virtualNetworks/subnets/write",
    "Microsoft.Network/virtualNetworks/write",
    "Microsoft.Resources/deployments/read",
    "Microsoft.Resources/deployments/write"
  ],
  "/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X44/providers/Microsoft.ContainerService/managedClusters/azmpfakstestcluster": [
    "Microsoft.ContainerService/managedClusters/read",
    "Microsoft.ContainerService/managedClusters/write"
  ],
  "/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X44/providers/Microsoft.Network/virtualNetworks/azmpfakstestvnet": [
    "Microsoft.Network/virtualNetworks/read",
    "Microsoft.Network/virtualNetworks/write"
  ],
  "/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X44/providers/Microsoft.Network/virtualNetworks/azmpfakstestvnet/subnets/azmpfakstestsubnet": [
    "Microsoft.Network/virtualNetworks/subnets/read",
    "Microsoft.Network/virtualNetworks/subnets/write"
  ]
}

```

### Viewing info, warn, or debug level logs

By default, the log level is error. More verbose logs can be viewed by setting the LOG_LEVEL environment variable to info, warn, or debug. Additionally, the global flag --verbose can be used to view info level logging and --debug can be used to view debug level logging. The following is a sample of info level logging:

```shell
export MPF_SUBSCRIPTIONID=YOUR_SUBSCRIPTION_ID
export MPF_TENANTID=YOUR_TENANT_ID
export MPF_SPCLIENTID=YOUR_SP_CLIENT_ID
export MPF_SPCLIENTSECRET=YOUR_SP_CLIENT_SECRET
export MPF_SPOBJECTID=YOUR_SP_OBJECT_ID

$ ./azmpf arm --templateFilePath ./samples/templates/aks-private-subnet.json --parametersFilePath ./samples/templates/aks-private-subnet-parameters.json --verbose

INFO[0000] Executing MPF for ARM                        
INFO[0000] TemplateFilePath: ./samples/templates/aks-private-subnet.json 
INFO[0000] ParametersFilePath: ./samples/templates/aks-private-subnet-parameters.json 
INFO[0000] roleDefinitionResourceID: /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/providers/Microsoft.Authorization/roleDefinitions/7c24bf65-d277-42ea-9082-fc3c71e463a7 
INFO[0012] Creating Resource Group: testdeployrg-dSwhpJw  
INFO[0017] Resource Group: testdeployrg-dSwhpJw created successfully  
INFO[0019] Deleted all existing role assignments for service principal  
INFO[0019] Initializing Custom Role                     
INFO[0022] Custom role initialized successfully         
INFO[0022] Assigning new custom role to service principal 
INFO[0027] New Custom Role assigned to service principal successfully 
INFO[0027] Adding initial permissions to requiredPermissions map 
INFO[0034] Whatif Results Response Body is empty, retrying in a bit... 
INFO[0036] Whatif Results Response Body is empty, retrying in a bit... 
INFO[0037] Whatif Results Response Body is empty, retrying in a bit... 
INFO[0039] Whatif Results Response Received..           
INFO[0039] Successfully Parsed Deployment Authorization Error 
INFO[0039] Adding mising scopes/permissions to final result map... 
INFO[0039] Adding permission/scope to role...........   
INFO[0042] Permission/scope added to role successfully  
INFO[0051] Whatif Results Response Body is empty, retrying in a bit... 
INFO[0053] Whatif Results Response Body is empty, retrying in a bit... 
INFO[0054] Whatif Results Response Body is empty, retrying in a bit... 
INFO[0057] Whatif Results Response Body is empty, retrying in a bit... 
INFO[0058] Whatif Results Response Body is empty, retrying in a bit... 
INFO[0059] Whatif Results Response Body is empty, retrying in a bit... 
INFO[0060] Whatif Results Response Body is empty, retrying in a bit... 
INFO[0062] Whatif Results Response Received..           
WARN[0062] Non Authorizaton error occured: {"status":"Failed","error":{"code":"InvalidTemplateDeployment","message":"The template deployment 'testDeploy-kyjUxeB' is not valid according to the validation procedure. The tracking id is '6c41b65d-4606-400d-80d1-0ae76bb8b2b3'. See inner errors for details.","details":[{"code":"InvalidParameter","message":"Preflight validation check for resource(s) for container service azmpfakstestcluster in resource group testdeployrg-dSwhpJw failed. Message: Required parameter servicePrincipalProfile is missing (null).. Details: "}]}} 
INFO[0062] Authorization Successful                     
INFO[0062] Cleaning up resources...                     
INFO[0062] *************************                    
INFO[0062] No additional cleanup needed in WhatIf mode  
INFO[0062] *************************                    
INFO[0071] Role definition deleted successfully         
INFO[0075] Resource group deletion initiated successfully... 
------------------------------------------------------------------------------------------------------------------------------------------
Permissions Required:
------------------------------------------------------------------------------------------------------------------------------------------
Microsoft.ContainerService/managedClusters/read
Microsoft.ContainerService/managedClusters/write
Microsoft.Network/virtualNetworks/read
Microsoft.Network/virtualNetworks/subnets/read
Microsoft.Network/virtualNetworks/subnets/write
Microsoft.Network/virtualNetworks/write
Microsoft.Resources/deployments/read
Microsoft.Resources/deployments/write
------------------------------------------------------------------------------------------------------------------------------------------

```

### Complex ARM template sample

Let us look at the detailed output of a more complex ARM template. The following is the command used to run the sample:

```shell
export MPF_SUBSCRIPTIONID=YOUR_SUBSCRIPTION_ID
export MPF_TENANTID=YOUR_TENANT_ID
export MPF_SPCLIENTID=YOUR_SP_CLIENT_ID
export MPF_SPCLIENTSECRET=YOUR_SP_CLIENT_SECRET
export MPF_SPOBJECTID=YOUR_SP_OBJECT_ID

$ ./azmpf arm --templateFilePath ./samples/templates/multi-resource-template.json --parametersFilePath ./samples/templates/multi-resource-parameters.json --showDetailedOutput 

------------------------------------------------------------------------------------------------------------------------------------------
Permissions Required:
------------------------------------------------------------------------------------------------------------------------------------------
Microsoft.Authorization/roleAssignments/read
Microsoft.Authorization/roleAssignments/write
Microsoft.Compute/virtualMachines/extensions/read
Microsoft.Compute/virtualMachines/extensions/write
Microsoft.Compute/virtualMachines/read
Microsoft.Compute/virtualMachines/write
Microsoft.ContainerRegistry/registries/read
Microsoft.ContainerRegistry/registries/write
Microsoft.ContainerService/managedClusters/read
Microsoft.ContainerService/managedClusters/write
Microsoft.Insights/actionGroups/read
Microsoft.Insights/actionGroups/write
Microsoft.Insights/activityLogAlerts/read
Microsoft.Insights/activityLogAlerts/write
Microsoft.Insights/diagnosticSettings/read
Microsoft.Insights/diagnosticSettings/write
Microsoft.KeyVault/vaults/read
Microsoft.KeyVault/vaults/write
Microsoft.ManagedIdentity/userAssignedIdentities/read
Microsoft.ManagedIdentity/userAssignedIdentities/write
Microsoft.Network/ApplicationGatewayWebApplicationFirewallPolicies/read
Microsoft.Network/ApplicationGatewayWebApplicationFirewallPolicies/write
Microsoft.Network/applicationGateways/read
Microsoft.Network/applicationGateways/write
Microsoft.Network/bastionHosts/read
Microsoft.Network/bastionHosts/write
Microsoft.Network/natGateways/read
Microsoft.Network/natGateways/write
Microsoft.Network/networkInterfaces/read
Microsoft.Network/networkInterfaces/write
Microsoft.Network/networkSecurityGroups/read
Microsoft.Network/networkSecurityGroups/write
Microsoft.Network/privateDnsZones/read
Microsoft.Network/privateDnsZones/virtualNetworkLinks/read
Microsoft.Network/privateDnsZones/virtualNetworkLinks/write
Microsoft.Network/privateDnsZones/write
Microsoft.Network/privateEndpoints/privateDnsZoneGroups/read
Microsoft.Network/privateEndpoints/privateDnsZoneGroups/write
Microsoft.Network/privateEndpoints/read
Microsoft.Network/privateEndpoints/write
Microsoft.Network/publicIPAddresses/read
Microsoft.Network/publicIPAddresses/write
Microsoft.Network/publicIPPrefixes/read
Microsoft.Network/publicIPPrefixes/write
Microsoft.Network/virtualNetworks/read
Microsoft.Network/virtualNetworks/write
Microsoft.OperationalInsights/workspaces/read
Microsoft.OperationalInsights/workspaces/write
Microsoft.OperationsManagement/solutions/read
Microsoft.OperationsManagement/solutions/write
Microsoft.Resources/deployments/read
Microsoft.Resources/deployments/write
Microsoft.Storage/storageAccounts/read
Microsoft.Storage/storageAccounts/write
------------------------------------------------------------------------------------------------------------------------------------------


Breakdown of permissions by different resource types:

Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Network/privateEndpoints/AcrPrivateEndpoint: 
Microsoft.Network/privateEndpoints/read
Microsoft.Network/privateEndpoints/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Compute/virtualMachines/TestVm/extensions/LogAnalytics: 
Microsoft.Compute/virtualMachines/extensions/read
Microsoft.Compute/virtualMachines/extensions/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Network/privateEndpoints/KeyVaultPrivateEndpoint/privateDnsZoneGroups/vaultPrivateDnsZoneGroup: 
Microsoft.Network/privateEndpoints/privateDnsZoneGroups/read
Microsoft.Network/privateEndpoints/privateDnsZoneGroups/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.OperationalInsights/workspaces/mantmplawsp131: 
Microsoft.OperationalInsights/workspaces/read
Microsoft.OperationalInsights/workspaces/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Network/privateDnsZones/privatelink.azurecr.io: 
Microsoft.Network/privateDnsZones/read
Microsoft.Network/privateDnsZones/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Authorization/roleAssignments/d65bd81b-db67-5930-a2dc-4f7255ed2015: 
Microsoft.Authorization/roleAssignments/read
Microsoft.Authorization/roleAssignments/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.ContainerService/managedClusters/aks-gqrwcl54p5g24: 
Microsoft.ContainerService/managedClusters/read
Microsoft.ContainerService/managedClusters/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.ManagedIdentity/userAssignedIdentities/appgw-gqrwcl54p5g24ManagedIdentity: 
Microsoft.ManagedIdentity/userAssignedIdentities/read
Microsoft.ManagedIdentity/userAssignedIdentities/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Network/privateDnsZones/privatelink.blob.core.windows.net: 
Microsoft.Network/privateDnsZones/read
Microsoft.Network/privateDnsZones/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Network/networkInterfaces/TestVmNic: 
Microsoft.Network/networkInterfaces/read
Microsoft.Network/networkInterfaces/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.ContainerRegistry/registries/acrgqrwcl54p5g24: 
Microsoft.ContainerRegistry/registries/read
Microsoft.ContainerRegistry/registries/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Authorization/roleAssignments/2c38f3d4-5171-5d2f-9cf8-47864a926b0a: 
Microsoft.Authorization/roleAssignments/read
Microsoft.Authorization/roleAssignments/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Network/privateEndpoints/AcrPrivateEndpoint/privateDnsZoneGroups/registryPrivateDnsZoneGroup: 
Microsoft.Network/privateEndpoints/privateDnsZoneGroups/read
Microsoft.Network/privateEndpoints/privateDnsZoneGroups/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.ManagedIdentity/userAssignedIdentities/aks-gqrwcl54p5g24ManagedIdentity: 
Microsoft.ManagedIdentity/userAssignedIdentities/read
Microsoft.ManagedIdentity/userAssignedIdentities/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.KeyVault/vaults/keyvault-gqrwcl54p5g24: 
Microsoft.KeyVault/vaults/read
Microsoft.KeyVault/vaults/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Network/privateDnsZones/privatelink.azurecr.io/virtualNetworkLinks/link_to_aks-gqrwcl54p5g24vnet: 
Microsoft.Network/privateDnsZones/virtualNetworkLinks/read
Microsoft.Network/privateDnsZones/virtualNetworkLinks/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Network/ApplicationGatewayWebApplicationFirewallPolicies/appgw-gqrwcl54p5g24WafPolicy: 
Microsoft.Network/ApplicationGatewayWebApplicationFirewallPolicies/read
Microsoft.Network/ApplicationGatewayWebApplicationFirewallPolicies/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Network/publicIPPrefixes/aks-gqrwcl54p5g24PublicIpPrefix: 
Microsoft.Network/publicIPPrefixes/read
Microsoft.Network/publicIPPrefixes/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Authorization/roleAssignments/97184c34-a124-5589-9ab2-f38c59602a61: 
Microsoft.Authorization/roleAssignments/read
Microsoft.Authorization/roleAssignments/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Network/bastionHosts/aks-gqrwcl54p5g24Bastion/providers/Microsoft.Insights/diagnosticSettings/default: 
Microsoft.Insights/diagnosticSettings/read
Microsoft.Insights/diagnosticSettings/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Network/privateDnsZones/privatelink.vault.azure.net: 
Microsoft.Network/privateDnsZones/read
Microsoft.Network/privateDnsZones/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Network/virtualNetworks/aks-gqrwcl54p5g24Vnet: 
Microsoft.Network/virtualNetworks/read
Microsoft.Network/virtualNetworks/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.OperationsManagement/solutions/ContainerInsights(mantmplawsp131): 
Microsoft.OperationsManagement/solutions/read
Microsoft.OperationsManagement/solutions/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Storage/storageAccounts/bootgqrwcl54p5g24: 
Microsoft.Storage/storageAccounts/read
Microsoft.Storage/storageAccounts/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Network/applicationGateways/appgw-gqrwcl54p5g24: 
Microsoft.Network/applicationGateways/read
Microsoft.Network/applicationGateways/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Network/privateEndpoints/BlobStorageAccountPrivateEndpoint/privateDnsZoneGroups/blobPrivateDnsZoneGroup: 
Microsoft.Network/privateEndpoints/privateDnsZoneGroups/read
Microsoft.Network/privateEndpoints/privateDnsZoneGroups/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Network/privateEndpoints/BlobStorageAccountPrivateEndpoint: 
Microsoft.Network/privateEndpoints/read
Microsoft.Network/privateEndpoints/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Network/publicIPAddresses/aks-gqrwcl54p5g24BastionPublicIp: 
Microsoft.Network/publicIPAddresses/read
Microsoft.Network/publicIPAddresses/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Network/privateEndpoints/KeyVaultPrivateEndpoint: 
Microsoft.Network/privateEndpoints/read
Microsoft.Network/privateEndpoints/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Network/privateDnsZones/privatelink.blob.core.windows.net/virtualNetworkLinks/link_to_aks-gqrwcl54p5g24vnet: 
Microsoft.Network/privateDnsZones/virtualNetworkLinks/read
Microsoft.Network/privateDnsZones/virtualNetworkLinks/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Network/privateDnsZones/privatelink.vault.azure.net/virtualNetworkLinks/link_to_aks-gqrwcl54p5g24vnet: 
Microsoft.Network/privateDnsZones/virtualNetworkLinks/read
Microsoft.Network/privateDnsZones/virtualNetworkLinks/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Network/natGateways/aks-gqrwcl54p5g24NatGateway: 
Microsoft.Network/natGateways/read
Microsoft.Network/natGateways/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Insights/actionGroups/emailActionGroupName: 
Microsoft.Insights/actionGroups/read
Microsoft.Insights/actionGroups/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Network/bastionHosts/aks-gqrwcl54p5g24Bastion: 
Microsoft.Network/bastionHosts/read
Microsoft.Network/bastionHosts/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Network/networkSecurityGroups/VmSubnetNsg/providers/Microsoft.Insights/diagnosticSettings/default: 
Microsoft.Insights/diagnosticSettings/read
Microsoft.Insights/diagnosticSettings/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Network/networkSecurityGroups/VmSubnetNsg: 
Microsoft.Network/networkSecurityGroups/read
Microsoft.Network/networkSecurityGroups/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Insights/activityLogAlerts/AllAzureAdvisorAlert: 
Microsoft.Insights/activityLogAlerts/read
Microsoft.Insights/activityLogAlerts/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Compute/virtualMachines/TestVm: 
Microsoft.Compute/virtualMachines/read
Microsoft.Compute/virtualMachines/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.ManagedIdentity/userAssignedIdentities/aks-gqrwcl54p5g24AadPodManagedIdentity: 
Microsoft.ManagedIdentity/userAssignedIdentities/read
Microsoft.ManagedIdentity/userAssignedIdentities/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-nib4c9B/providers/Microsoft.Compute/virtualMachines/TestVm/extensions/DependencyAgent: 
Microsoft.Compute/virtualMachines/extensions/read
Microsoft.Compute/virtualMachines/extensions/write
--------------
```

### Subscription scoped ARM deployment

The --subscriptionScoped flag indicates whether the deployment is scoped to a subscription. If set, the deployment will target a subscription else it is resource group scoped deployment. The following is a sample of subscription scoped ARM deployment:

```shell
export MPF_SUBSCRIPTIONID=YOUR_SUBSCRIPTION_ID
export MPF_TENANTID=YOUR_TENANT_ID
export MPF_SPCLIENTID=YOUR_SP_CLIENT_ID
export MPF_SPCLIENTSECRET=YOUR_SP_CLIENT_SECRET
export MPF_SPOBJECTID=YOUR_SP_OBJECT_ID
export MPF_BICEPEXECPATH="/opt/homebrew/bin/bicep" # Path to the Bicep executable

$ ./azmpf arm --templateFilePath ./samples/templates/subscription-scope-create-rg.json --parametersFilePath ./samples/templates/subscription-scope-create-rg-params.json --subscriptionScoped --location eastus2
------------------------------------------------------------------------------------------------------------------------------------------
Permissions Required:
------------------------------------------------------------------------------------------------------------------------------------------
Microsoft.Resources/deployments/read
Microsoft.Resources/deployments/write
Microsoft.Resources/subscriptions/resourceGroups/read
Microsoft.Resources/subscriptions/resourceGroups/write
------------------------------------------------------------------------------------------------------------------------------------------
```

### Subscription scoped Bicep deployment

The --subscriptionScoped flag indicates whether the deployment is scoped to a subscription. If set, the deployment will target a subscription else it is resource group scoped deployment. The following is a sample of subscription scoped Bicep deployment:

```shell
export MPF_SUBSCRIPTIONID=YOUR_SUBSCRIPTION_ID
export MPF_TENANTID=YOUR_TENANT_ID
export MPF_SPCLIENTID=YOUR_SP_CLIENT_ID
export MPF_SPCLIENTSECRET=YOUR_SP_CLIENT_SECRET
export MPF_SPOBJECTID=YOUR_SP_OBJECT_ID
export MPF_BICEPEXECPATH="/opt/homebrew/bin/bicep" # Path to the Bicep executable

$ ./azmpf bicep --bicepFilePath ./samples/bicep/subscription-scope-create-rg.bicep --parametersFilePath ./samples/bicep/subscription-scope-create-rg-params.json  --subscriptionScoped --location eastus2
------------------------------------------------------------------------------------------------------------------------------------------
Permissions Required:
------------------------------------------------------------------------------------------------------------------------------------------
Microsoft.Resources/deployments/read
Microsoft.Resources/deployments/write
Microsoft.Resources/subscriptions/resourceGroups/read
Microsoft.Resources/subscriptions/resourceGroups/write
------------------------------------------------------------------------------------------------------------------------------------------
```

