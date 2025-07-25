## MPF Display Options

### Deployment Mode for ARM and Bicep

ARM and Bicep templates now use **Full Deployment mode** exclusively with Incremental deployment mode, which actually deploys resources to determine permissions and then cleans them up automatically. This provides comprehensive and accurate permission detection but takes significantly longer than the previous what-if mode - expect execution times of several minutes to longer depending on template complexity and the resources being deployed.

**Recommendation**: Always use the `--verbose` flag with ARM and Bicep commands to see progress during long-running deployments.

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

$ ./azmpf arm --templateFilePath ./samples/templates/aks-private-subnet.json --parametersFilePath ./samples/templates/aks-private-subnet-parameters.json --showDetailedOutput --verbose
```

Or using PowerShell on Windows:

```powershell
# Detailed Text output
$env:MPF_SUBSCRIPTIONID = "YOUR_SUBSCRIPTION_ID"
$env:MPF_TENANTID = "YOUR_TENANT_ID"
$env:MPF_SPCLIENTID = "YOUR_SP_CLIENT_ID"
$env:MPF_SPCLIENTSECRET = "YOUR_SP_CLIENT_SECRET"
$env:MPF_SPOBJECTID = "YOUR_SP_OBJECT_ID"

.\azmpf.exe arm --templateFilePath .\samples\templates\aks-private-subnet.json --parametersFilePath .\samples\templates\aks-private-subnet-parameters.json --showDetailedOutput --verbose
```

Output:

```
INFO[0000] Executing MPF for ARM                        
INFO[0000] TemplateFilePath: ./samples/templates/aks-private-subnet.json
INFO[0000] ParametersFilePath: ./samples/templates/aks-private-subnet-parameters.json 
INFO[0000] Location: eastus2                            
INFO[0001] Creating Resource Group: testdeployrg-OJ2zCNA  
INFO[0007] Resource Group: testdeployrg-OJ2zCNA created successfully  
INFO[0008] Deleted all existing role assignments for service principal  
INFO[0008] Initializing Custom Role                     
INFO[0014] Custom role initialized successfully         
INFO[0014] Assigning new custom role to service principal 
INFO[0018] New Custom Role assigned to service principal successfully 
INFO[0023] Iteration Number: 0                          
INFO[0023] Successfully Parsed Deployment Authorization Error 
INFO[0023] Adding mising scopes/permissions to final result map... 
INFO[0023] Adding permission/scope to role...........   
INFO[0027] Permission/scope added to role successfully  
INFO[0082] Iteration Number: 1                          
INFO[0082] Successfully Parsed Deployment Authorization Error 
INFO[0082] Adding mising scopes/permissions to final result map... 
INFO[0082] Adding permission/scope to role...........   
INFO[0085] Permission/scope added to role successfully  
INFO[0560] Iteration Number: 2                          
INFO[0560] Authorization Successful                     
INFO[0560] Cleaning up resources...                     
INFO[0560] *************************                    
INFO[0573] Role definition deleted successfully         
INFO[0577] Resource group deletion initiated successfully... 
------------------------------------------------------------------------------------------------------------------------------------------
Permissions Required:
------------------------------------------------------------------------------------------------------------------------------------------
Microsoft.ContainerService/managedClusters/read
Microsoft.ContainerService/managedClusters/write
Microsoft.Network/virtualNetworks/read
Microsoft.Network/virtualNetworks/subnets/join/action
Microsoft.Network/virtualNetworks/subnets/read
Microsoft.Network/virtualNetworks/subnets/write
Microsoft.Network/virtualNetworks/write
Microsoft.Resources/deployments/read
Microsoft.Resources/deployments/write
------------------------------------------------------------------------------------------------------------------------------------------


Break down of permissions by different resource types:

Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-OJ2zCNA/providers/Microsoft.Network/virtualNetworks/azmpfakstestvnet: 
Microsoft.Network/virtualNetworks/read
Microsoft.Network/virtualNetworks/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-OJ2zCNA/providers/Microsoft.Network/virtualNetworks/azmpfakstestvnet/subnets/azmpfakstestsubnet: 
Microsoft.Network/virtualNetworks/subnets/join/action
Microsoft.Network/virtualNetworks/subnets/read
Microsoft.Network/virtualNetworks/subnets/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-OJ2zCNA/providers/Microsoft.ContainerService/managedClusters/azmpfakstestcluster: 
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

$ ./azmpf arm --templateFilePath ./samples/templates/aks-private-subnet.json --parametersFilePath ./samples/templates/aks-private-subnet-parameters.json --outputFormat json --verbose
```

Or using PowerShell on Windows:

```powershell
$env:MPF_SUBSCRIPTIONID = "YOUR_SUBSCRIPTION_ID"
$env:MPF_TENANTID = "YOUR_TENANT_ID"
$env:MPF_SPCLIENTID = "YOUR_SP_CLIENT_ID"
$env:MPF_SPCLIENTSECRET = "YOUR_SP_CLIENT_SECRET"
$env:MPF_SPOBJECTID = "YOUR_SP_OBJECT_ID"

.\azmpf.exe arm --templateFilePath .\samples\templates\aks-private-subnet.json --parametersFilePath .\samples\templates\aks-private-subnet-parameters.json --outputFormat json --verbose
```

Output:

```
INFO[0000] Executing MPF for ARM                        
INFO[0000] TemplateFilePath: ./samples/templates/aks-private-subnet.json
INFO[0000] ParametersFilePath: ./samples/templates/aks-private-subnet-parameters.json 
INFO[0000] Location: eastus2                            
INFO[0001] Creating Resource Group: testdeployrg-Uqc4z3E  
INFO[0007] Resource Group: testdeployrg-Uqc4z3E created successfully  
INFO[0007] Deleted all existing role assignments for service principal  
INFO[0007] Initializing Custom Role                     
INFO[0014] Custom role initialized successfully         
INFO[0014] Assigning new custom role to service principal 
INFO[0018] New Custom Role assigned to service principal successfully 
INFO[0024] Iteration Number: 0                          
INFO[0024] Successfully Parsed Deployment Authorization Error 
INFO[0024] Adding mising scopes/permissions to final result map... 
INFO[0024] Adding permission/scope to role...........   
INFO[0027] Permission/scope added to role successfully  
INFO[0145] Iteration Number: 1                          
INFO[0145] Successfully Parsed Deployment Authorization Error 
INFO[0145] Adding mising scopes/permissions to final result map... 
INFO[0145] Adding permission/scope to role...........   
INFO[0148] Permission/scope added to role successfully  
INFO[0434] Iteration Number: 2                          
INFO[0434] Authorization Successful                     
INFO[0434] Cleaning up resources...                     
INFO[0434] *************************                    
INFO[0447] Role definition deleted successfully         
INFO[0452] Resource group deletion initiated successfully... 
{
  "permissions": [
    "Microsoft.ContainerService/managedClusters/read",
    "Microsoft.ContainerService/managedClusters/write",
    "Microsoft.Network/virtualNetworks/read",
    "Microsoft.Network/virtualNetworks/subnets/join/action",
    "Microsoft.Network/virtualNetworks/subnets/read",
    "Microsoft.Network/virtualNetworks/subnets/write",
    "Microsoft.Network/virtualNetworks/write",
    "Microsoft.Resources/deployments/read",
    "Microsoft.Resources/deployments/write"
  ],
  "permissionsByResourceScope": [
    {
      "resourceScope": "/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-Uqc4z3E/providers/Microsoft.Network/virtualNetworks/azmpfakstestvnet",
      "permissions": [
        "Microsoft.Network/virtualNetworks/read",
        "Microsoft.Network/virtualNetworks/write"
      ]
    },
    {
      "resourceScope": "/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-Uqc4z3E/providers/Microsoft.Network/virtualNetworks/azmpfakstestvnet/subnets/azmpfakstestsubnet",
      "permissions": [
        "Microsoft.Network/virtualNetworks/subnets/join/action",
        "Microsoft.Network/virtualNetworks/subnets/read",
        "Microsoft.Network/virtualNetworks/subnets/write"
      ]
    },
    {
      "resourceScope": "/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-Uqc4z3E/providers/Microsoft.ContainerService/managedClusters/azmpfakstestcluster",
      "permissions": [
        "Microsoft.ContainerService/managedClusters/read",
        "Microsoft.ContainerService/managedClusters/write"
      ]
    }
  ]
}

```

### Viewing info, warn, or debug level logs

By default, the log level is error. More verbose logs can be viewed by setting the LOG_LEVEL environment variable to info, warn, or debug. Additionally, the global flag --verbose can be used to view info level logging and --debug can be used to view debug level logging. The following is a sample of default logging (error level only):

```shell
export MPF_SUBSCRIPTIONID=YOUR_SUBSCRIPTION_ID
export MPF_TENANTID=YOUR_TENANT_ID
export MPF_SPCLIENTID=YOUR_SP_CLIENT_ID
export MPF_SPCLIENTSECRET=YOUR_SP_CLIENT_SECRET
export MPF_SPOBJECTID=YOUR_SP_OBJECT_ID

$ ./azmpf arm --templateFilePath ./samples/templates/aks-private-subnet.json --parametersFilePath ./samples/templates/aks-private-subnet-parameters.json
```

Or using PowerShell on Windows:

```powershell
$env:MPF_SUBSCRIPTIONID = "YOUR_SUBSCRIPTION_ID"
$env:MPF_TENANTID = "YOUR_TENANT_ID"
$env:MPF_SPCLIENTID = "YOUR_SP_CLIENT_ID"
$env:MPF_SPCLIENTSECRET = "YOUR_SP_CLIENT_SECRET"
$env:MPF_SPOBJECTID = "YOUR_SP_OBJECT_ID"

.\azmpf.exe arm --templateFilePath .\samples\templates\aks-private-subnet.json --parametersFilePath .\samples\templates\aks-private-subnet-parameters.json
```

Output:

```
Permissions Required:
------------------------------------------------------------------------------------------------------------------------------------------
Microsoft.ContainerService/managedClusters/read
Microsoft.ContainerService/managedClusters/write
Microsoft.Network/virtualNetworks/read
Microsoft.Network/virtualNetworks/subnets/join/action
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
```

Or using PowerShell on Windows:

```powershell
$env:MPF_SUBSCRIPTIONID = "YOUR_SUBSCRIPTION_ID"
$env:MPF_TENANTID = "YOUR_TENANT_ID"
$env:MPF_SPCLIENTID = "YOUR_SP_CLIENT_ID"
$env:MPF_SPCLIENTSECRET = "YOUR_SP_CLIENT_SECRET"
$env:MPF_SPOBJECTID = "YOUR_SP_OBJECT_ID"

.\azmpf.exe arm --templateFilePath .\samples\templates\multi-resource-template.json --parametersFilePath .\samples\templates\multi-resource-parameters.json --showDetailedOutput
```

Output:

```
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
Microsoft.OperationalInsights/workspaces/listKeys/action
Microsoft.OperationalInsights/workspaces/read
Microsoft.OperationalInsights/workspaces/sharedKeys/action
Microsoft.OperationalInsights/workspaces/write
Microsoft.OperationsManagement/solutions/read
Microsoft.OperationsManagement/solutions/write
Microsoft.Resources/deployments/read
Microsoft.Resources/deployments/write
Microsoft.Resources/subscriptions/resourceGroups/read
Microsoft.Storage/storageAccounts/read
Microsoft.Storage/storageAccounts/write
------------------------------------------------------------------------------------------------------------------------------------------


Break down of permissions by different resource types:

Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-6Xr5Jew/providers/Microsoft.Network/publicIPAddresses/aks-v7crf54zl65ugBastionPublicIp: 
Microsoft.Network/publicIPAddresses/read
Microsoft.Network/publicIPAddresses/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-6Xr5Jew/providers/Microsoft.Network/networkSecurityGroups/VmSubnetNsg/providers/Microsoft.Insights/diagnosticSettings/default: 
Microsoft.Insights/diagnosticSettings/read
Microsoft.Insights/diagnosticSettings/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-6Xr5Jew/providers/Microsoft.Insights/activityLogAlerts/AllAzureAdvisorAlert: 
Microsoft.Insights/activityLogAlerts/read
Microsoft.Insights/activityLogAlerts/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-6Xr5Jew/providers/Microsoft.Network/privateDnsZones/privatelink.azurecr.io/virtualNetworkLinks/link_to_aks-v7crf54zl65ugvnet: 
Microsoft.Network/privateDnsZones/virtualNetworkLinks/read
Microsoft.Network/privateDnsZones/virtualNetworkLinks/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-6Xr5Jew/providers/Microsoft.Network/privateDnsZones/privatelink.blob.core.windows.net/virtualNetworkLinks/link_to_aks-v7crf54zl65ugvnet: 
Microsoft.Network/privateDnsZones/virtualNetworkLinks/read
Microsoft.Network/privateDnsZones/virtualNetworkLinks/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-6Xr5Jew: 
Microsoft.Resources/subscriptions/resourceGroups/read
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-6Xr5Jew/providers/Microsoft.OperationsManagement/solutions/ContainerInsights(mantmplawsp131): 
Microsoft.OperationsManagement/solutions/read
Microsoft.OperationsManagement/solutions/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-6Xr5Jew/providers/Microsoft.Authorization/roleAssignments/4ca965ad-af42-5996-80a4-6d589a2cae77: 
Microsoft.Authorization/roleAssignments/read
Microsoft.Authorization/roleAssignments/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-6Xr5Jew/providers/Microsoft.Network/bastionHosts/aks-v7crf54zl65ugBastion: 
Microsoft.Network/bastionHosts/read
Microsoft.Network/bastionHosts/write
--------------


Permissions required for /subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-6Xr5Jew/providers/Microsoft.Compute/virtualMachines/TestVm: 
Microsoft.Compute/virtualMachines/read
Microsoft.Compute/virtualMachines/write

(showing only first few resource breakdowns - there are many more resources in this complex template)

```