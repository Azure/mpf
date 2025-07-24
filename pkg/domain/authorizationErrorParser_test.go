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

// test parseMultiAuthorizationFailedErrors
import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNonAuthorizationError(t *testing.T) {
	nonAuthorizationError := "{\"error\":{\"code\":\"InvalidTemplateDeployment\",\"message\":\"The template deployment failed with error: 'The resource with id: '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.Network/applicationGateways/appgw-zcrtnp4zt4k44' failed validation with message: 'Application Gateway sku tier is not allowed for this subscription.'.'.\"}}"
	spm, err := GetScopePermissionsFromAuthError(nonAuthorizationError)
	assert.NotNil(t, err)
	// assert.NotNil(t, spm)
	assert.Equal(t, len(spm), 0)
}

func TestBlankDeploymentError(t *testing.T) {
	spm, err := GetScopePermissionsFromAuthError("")
	assert.NotNil(t, err)
	assert.Nil(t, spm)
}

func TestSingleAuthorizationFailedError(t *testing.T) {
	singleAuthorizationFailedError := "{\"error\":{\"code\":\"AuthorizationFailed\",\"message\":\"The client 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' with object id 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' does not have authorization to perform action 'Microsoft.Resources/deployments/write' over scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourcegroups/testdeployrg/providers/Microsoft.Resources/deployments/testDeploy-f6RkAT3' or the scope is invalid. If access was recently granted, please refresh your credentials.\"}}"
	spm, err := GetScopePermissionsFromAuthError(singleAuthorizationFailedError)
	l := len(spm)
	assert.Nil(t, err)
	assert.NotNil(t, spm)
	assert.GreaterOrEqual(t, l, 1)

	// Assert first and last values in the map
	firstMatch := spm["/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourcegroups/testdeployrg/providers/Microsoft.Resources/deployments/testDeploy-f6RkAT3"]
	assert.Equal(t, "Microsoft.Resources/deployments/write", firstMatch[0])
}

func TestSingleAuthorizationSpaceFailedError(t *testing.T) {
	singleAuthorizationSpaceFailedError := "{\"error\":{\"code\":\"InvalidTemplateDeployment\",\"message\":\"The template deployment failed with error: 'Authorization failed for template resource 'appgw-zcrtnp4zt4k44WafPolicy' of type 'Microsoft.Network/ApplicationGatewayWebApplicationFirewallPolicies'. The client 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' with object id 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' does not have permission to perform action 'Microsoft.Network/ApplicationGatewayWebApplicationFirewallPolicies/write' at scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.Network/ApplicationGatewayWebApplicationFirewallPolicies/appgw-zcrtnp4zt4k44WafPolicy'.'.\"}}"
	spm, err := GetScopePermissionsFromAuthError(singleAuthorizationSpaceFailedError)
	l := len(spm)
	assert.Nil(t, err)
	assert.NotNil(t, spm)
	assert.GreaterOrEqual(t, l, 1)

	// Assert first and last values in the map
	firstMatch := spm["/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.Network/ApplicationGatewayWebApplicationFirewallPolicies/appgw-zcrtnp4zt4k44WafPolicy"]
	assert.Equal(t, "Microsoft.Network/ApplicationGatewayWebApplicationFirewallPolicies/write", firstMatch[0])
}

func TestMultiAuthorizationSpaceFailedErrors(t *testing.T) {
	multiLineError := "{\"error\":{\"code\":\"InvalidTemplateDeployment\",\"message\":\"Deployment failed with multiple errors: 'Authorization failed for template resource 'aks-zcrtnp4zt4k44PublicIpPrefix' of type 'Microsoft.Network/publicIPPrefixes'. The client 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' with object id 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' does not have permission to perform action 'Microsoft.Network/publicIPPrefixes/write' at scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.Network/publicIPPrefixes/aks-zcrtnp4zt4k44PublicIpPrefix'.:Authorization failed for template resource 'aks-zcrtnp4zt4k44NatGateway' of type 'Microsoft.Network/natGateways'. The client 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' with object id 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' does not have permission to perform action 'Microsoft.Network/natGateways/write' at scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.Network/natGateways/aks-zcrtnp4zt4k44NatGateway'.:Authorization failed for template resource 'aks-zcrtnp4zt4k44BastionPublicIp' of type 'Microsoft.Network/publicIPAddresses'. The client 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' with object id 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' does not have permission to perform action 'Microsoft.Network/publicIPAddresses/write' at scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.Network/publicIPAddresses/aks-zcrtnp4zt4k44BastionPublicIp'.:Authorization failed for template resource 'aks-zcrtnp4zt4k44Bastion' of type 'Microsoft.Network/bastionHosts'. The client 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' with object id 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' does not have permission to perform action 'Microsoft.Network/bastionHosts/write' at scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.Network/bastionHosts/aks-zcrtnp4zt4k44Bastion'.:Authorization failed for template resource 'default' of type 'Microsoft.Insights/diagnosticSettings'. The client 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' with object id 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' does not have permission to perform action 'Microsoft.Insights/diagnosticSettings/write' at scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.Network/bastionHosts/aks-zcrtnp4zt4k44Bastion/providers/Microsoft.Insights/diagnosticSettings/default'.:Authorization failed for template resource 'bootzcrtnp4zt4k44' of type 'Microsoft.Storage/storageAccounts'. The client 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' with object id 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' does not have permission to perform action 'Microsoft.Storage/storageAccounts/write' at scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.Storage/storageAccounts/bootzcrtnp4zt4k44'.:Authorization failed for template resource 'TestVmNic' of type 'Microsoft.Network/networkInterfaces'. The client 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' with object id 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' does not have permission to perform action 'Microsoft.Network/networkInterfaces/write' at scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.Network/networkInterfaces/TestVmNic'.:Authorization failed for template resource 'TestVm' of type 'Microsoft.Compute/virtualMachines'. The client 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' with object id 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' does not have permission to perform action 'Microsoft.Compute/virtualMachines/write' at scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.Compute/virtualMachines/TestVm'.:Authorization failed for template resource 'TestVm/LogAnalytics' of type 'Microsoft.Compute/virtualMachines/extensions'. The client 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' with object id 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' does not have permission to perform action 'Microsoft.Compute/virtualMachines/extensions/write' at scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.Compute/virtualMachines/TestVm/extensions/LogAnalytics'.:Authorization failed for template resource 'TestVm/DependencyAgent' of type 'Microsoft.Compute/virtualMachines/extensions'. The client 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' with object id 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' does not have permission to perform action 'Microsoft.Compute/virtualMachines/extensions/write' at scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.Compute/virtualMachines/TestVm/extensions/DependencyAgent'.:Authorization failed for template resource 'VmSubnetNsg' of type 'Microsoft.Network/networkSecurityGroups'. The client 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' with object id 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' does not have permission to perform action 'Microsoft.Network/networkSecurityGroups/write' at scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.Network/networkSecurityGroups/VmSubnetNsg'.:Authorization failed for template resource 'default' of type 'Microsoft.Insights/diagnosticSettings'. The client 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' with object id 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' does not have permission to perform action 'Microsoft.Insights/diagnosticSettings/write' at scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.Network/networkSecurityGroups/VmSubnetNsg/providers/Microsoft.Insights/diagnosticSettings/default'.:Authorization failed for template resource 'aks-zcrtnp4zt4k44Vnet' of type 'Microsoft.Network/virtualNetworks'. The client 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' with object id 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' does not have permission to perform action 'Microsoft.Network/virtualNetworks/write' at scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.Network/virtualNetworks/aks-zcrtnp4zt4k44Vnet'.:Authorization failed for template resource 'aks-zcrtnp4zt4k44ManagedIdentity' of type 'Microsoft.ManagedIdentity/userAssignedIdentities'. The client 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' with object id 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' does not have permission to perform action 'Microsoft.ManagedIdentity/userAssignedIdentities/write' at scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/aks-zcrtnp4zt4k44ManagedIdentity'.:Authorization failed for template resource 'appgw-zcrtnp4zt4k44ManagedIdentity' of type 'Microsoft.ManagedIdentity/userAssignedIdentities'. The client 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' with object id 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' does not have permission to perform action 'Microsoft.ManagedIdentity/userAssignedIdentities/write' at scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/appgw-zcrtnp4zt4k44ManagedIdentity'.:Authorization failed for template resource 'aks-zcrtnp4zt4k44AadPodManagedIdentity' of type 'Microsoft.ManagedIdentity/userAssignedIdentities'. The client 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' with object id 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' does not have permission to perform action 'Microsoft.ManagedIdentity/userAssignedIdentities/write' at scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/aks-zcrtnp4zt4k44AadPodManagedIdentity'.:Authorization failed for template resource '2b59f3a8-2ca1-5443-8b8d-9967389389ef' of type 'Microsoft.Authorization/roleAssignments'. The client 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' with object id 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' does not have permission to perform action 'Microsoft.Authorization/roleAssignments/write' at scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.Authorization/roleAssignments/2b59f3a8-2ca1-5443-8b8d-9967389389ef'.:Authorization failed for template resource 'f349d09e-d18e-5e03-84c5-fa8c85263431' of type 'Microsoft.Authorization/roleAssignments'. The client 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' with object id 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' does not have permission to perform action 'Microsoft.Authorization/roleAssignments/write' at scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.Authorization/roleAssignments/f349d09e-d18e-5e03-84c5-fa8c85263431'.:Authorization failed for template resource 'b592ef3b-748a-5026-ae8a-1f9be12a59e7' of type 'Microsoft.Authorization/roleAssignments'. The client 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' with object id 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' does not have permission to perform action 'Microsoft.Authorization/roleAssignments/write' at scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.Authorization/roleAssignments/b592ef3b-748a-5026-ae8a-1f9be12a59e7'.:Authorization failed for template resource 'keyvault-zcrtnp4zt4k44' of type 'Microsoft.KeyVault/vaults'. The client 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' with object id 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' does not have permission to perform action 'Microsoft.KeyVault/vaults/write' at scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.KeyVault/vaults/keyvault-zcrtnp4zt4k44'.'\"}}"
	spm, err := GetScopePermissionsFromAuthError(multiLineError)
	l := len(spm)
	assert.Nil(t, err)
	assert.NotNil(t, spm)
	assert.GreaterOrEqual(t, l, 20)

	// Assert first and last values in the map
	firstMatch := spm["/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.Network/publicIPPrefixes/aks-zcrtnp4zt4k44PublicIpPrefix"]
	assert.Equal(t, "Microsoft.Network/publicIPPrefixes/write", firstMatch[0])

	lastMatch := spm["/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg/providers/Microsoft.KeyVault/vaults/keyvault-zcrtnp4zt4k44"]
	assert.Equal(t, "Microsoft.KeyVault/vaults/write", lastMatch[0])

}
