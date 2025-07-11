//     MIT License
//
//     Copyright (c) Microsoft Corporation.
//
//     Permission is hereby granted, free of charge, to any person obtaining a copy
//     of this software and associated documentation files (the "Software"), to deal
//     in the Software without restriction, including without limitation the rights
//     to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//     copies of the Software, and to permit persons to whom the Software is
//     furnished to do so, subject to the following conditions:
//
//     The above copyright notice and this permission notice shall be included in all
//     copies or substantial portions of the Software.
//
//     THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//     IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//     FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//     AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//     LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//     OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//     SOFTWARE

package domain

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSingleLinkedAuthorizationFailedError(t *testing.T) {
	singleLinkedAuthorizationFailedError := "error: LinkedAuthorizationFailed: The client 'a31fc7f1-1349-4b3c-af16-60422be430cc' with object id 'a31fc7f1-1349-4b3c-af16-60422be430cc' has permission to perform action 'Microsoft.ContainerService/managedClusters/write' on scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/az-mpf-tf-test-rg/providers/Microsoft.ContainerService/managedClusters/aks-32a70ccbb3247e2b'; however, it does not have permission to perform action(s) 'Microsoft.Network/virtualNetworks/subnets/join/action' on the linked scope(s) '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/az-mpf-tf-test-rg/providers/Microsoft.Network/virtualNetworks/vnet-32a70ccbb3247e2b/subnets/subnet-32a70ccbb3247e2b' (respectively) or the linked scope(s) are invalid."
	spm, err := GetScopePermissionsFromAuthError(singleLinkedAuthorizationFailedError)
	l := len(spm)
	assert.Nil(t, err)
	assert.NotNil(t, spm)
	assert.GreaterOrEqual(t, l, 1)

	// Assert values in the map
	firstMatch := spm["/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/az-mpf-tf-test-rg/providers/Microsoft.Network/virtualNetworks/vnet-32a70ccbb3247e2b/subnets/subnet-32a70ccbb3247e2b"]
	assert.Equal(t, "Microsoft.Network/virtualNetworks/subnets/join/action", firstMatch[0])
}

func TestLinkedAuthorizationFailedErrors(t *testing.T) {
	LinkedAuthorizationFailedErrorMsg := `GET https://management.azure.com/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourcegroups/e2eTest-zSUDZd0/providers/Microsoft.Resources/deployments/e2eTest-MfwksQa/operationStatuses/08584489259249625918
--------------------------------------------------------------------------------
RESPONSE 200: 200 OK
ERROR CODE: DeploymentFailed
--------------------------------------------------------------------------------
{
  "status": "Failed",
  "error": {
    "code": "DeploymentFailed",
    "message": "At least one resource deployment operation failed. Please list deployment operations for details. Please see https://aka.ms/arm-deployment-operations for usage details.",
    "details": [
      {
        "code": "Forbidden",
        "message": "{\r
  \"error\": {\r
    \"code\": \"LinkedAuthorizationFailed\",\r
    \"message\": \"The client 'ddfcf162-2cf2-40cf-bd4a-49a63e248436' with object id 'ddfcf162-2cf2-40cf-bd4a-49a63e248436' has permission to perform action 'Microsoft.Insights/diagnosticSettings/write' on scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourcegroups/e2eTest-zSUDZd0/providers/Microsoft.KeyVault/vaults/keyvault-efuk33krrmm6q/providers/Microsoft.Insights/diagnosticSettings/default'; however, it does not have permission to perform action(s) 'Microsoft.OperationalInsights/workspaces/sharedKeys/action' on the linked scope(s) '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/e2eTest-zSUDZd0/providers/Microsoft.OperationalInsights/workspaces/mantmplawsp131' (respectively) or the linked scope(s) are invalid.\"\r
  }\r
}"
      },
      {
        "code": "Forbidden",
        "message": "{\r
  \"error\": {\r
    \"code\": \"LinkedAuthorizationFailed\",\r
    \"message\": \"The client 'ddfcf162-2cf2-40cf-bd4a-49a63e248436' with object id 'ddfcf162-2cf2-40cf-bd4a-49a63e248436' has permission to perform action 'Microsoft.Insights/diagnosticSettings/write' on scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourcegroups/e2eTest-zSUDZd0/providers/Microsoft.Network/networkSecurityGroups/VmSubnetNsg/providers/Microsoft.Insights/diagnosticSettings/default'; however, it does not have permission to perform action(s) 'Microsoft.OperationalInsights/workspaces/sharedKeys/action' on the linked scope(s) '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/e2eTest-zSUDZd0/providers/Microsoft.OperationalInsights/workspaces/mantmplawsp131' (respectively) or the linked scope(s) are invalid.\"\r
  }\r
}"
      },
      {
        "code": "Forbidden",
        "message": "{\r
  \"error\": {\r
    \"code\": \"LinkedAuthorizationFailed\",\r
    \"message\": \"The client 'ddfcf162-2cf2-40cf-bd4a-49a63e248436' with object id 'ddfcf162-2cf2-40cf-bd4a-49a63e248436' has permission to perform action 'Microsoft.Insights/diagnosticSettings/write' on scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourcegroups/e2eTest-zSUDZd0/providers/Microsoft.ContainerRegistry/registries/acrefuk33krrmm6q/providers/Microsoft.Insights/diagnosticSettings/default'; however, it does not have permission to perform action(s) 'Microsoft.OperationalInsights/workspaces/sharedKeys/action' on the linked scope(s) '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/e2eTest-zSUDZd0/providers/Microsoft.OperationalInsights/workspaces/mantmplawsp131' (respectively) or the linked scope(s) are invalid.\"\r
  }\r
}"
      },
      {
        "code": "Forbidden",
        "message": "{\r
  \"error\": {\r
    \"code\": \"AuthorizationFailed\",\r
    \"message\": \"The client 'ddfcf162-2cf2-40cf-bd4a-49a63e248436' with object id 'ddfcf162-2cf2-40cf-bd4a-49a63e248436' does not have authorization to perform action 'Microsoft.OperationalInsights/workspaces/listKeys/action' over scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourcegroups/e2eTest-zSUDZd0/providers/Microsoft.OperationalInsights/workspaces/mantmplawsp131' or the scope is invalid. If access was recently granted, please refresh your credentials.\"\r
  }\r
}"
      },
      {
        "code": "Forbidden",
        "message": "{\r
  \"error\": {\r
    \"code\": \"LinkedAuthorizationFailed\",\r
    \"message\": \"The client 'ddfcf162-2cf2-40cf-bd4a-49a63e248436' with object id 'ddfcf162-2cf2-40cf-bd4a-49a63e248436' has permission to perform action 'Microsoft.Insights/activityLogAlerts/write' on scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourcegroups/e2eTest-zSUDZd0/providers/Microsoft.Insights/activityLogAlerts/AllAzureAdvisorAlert'; however, it does not have permission to perform action(s) '/read' on the linked scope(s) '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/e2eTest-zSUDZd0' (respectively) or the linked scope(s) are invalid.\"\r
  }\r
}"
      },
      {
        "code": "BadRequest",
        "message": "{\r
  \"error\": {\r
    \"code\": \"IPv4StandardSkuPublicIpCountLimitReached\",\r
    \"message\": \"Cannot create more than 10 IPv4 Standard SKU public IP addresses for this subscription in this region.\",\r
    \"details\": []\r
  }\r
}"
      }
    ]
  }
}
--------------------------------------------------------------------------------
"
}`
	spm, err := GetScopePermissionsFromAuthError(LinkedAuthorizationFailedErrorMsg)
	l := len(spm)
	assert.Nil(t, err)
	assert.NotNil(t, spm)
	fmt.Printf("spm: %v\n", spm)
	assert.Equal(t, l, 3)
}
