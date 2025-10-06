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

package e2etests

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/Azure/mpf/pkg/domain"
	"github.com/Azure/mpf/pkg/infrastructure/ARMTemplateShared"
	"github.com/Azure/mpf/pkg/infrastructure/authorizationCheckers/ARMTemplateDeployment"

	// "github.com/Azure/mpf/pkg/infrastructure/authorizationCheckers/ARMTemplateWhatIf"
	mpfSharedUtils "github.com/Azure/mpf/pkg/infrastructure/mpfSharedUtils"
	resourceGroupManager "github.com/Azure/mpf/pkg/infrastructure/resourceGroupManager"
	sproleassignmentmanager "github.com/Azure/mpf/pkg/infrastructure/spRoleAssignmentManager"
	"github.com/Azure/mpf/pkg/usecase"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type MpfCLIArgs struct {
	SubscriptionID       string
	ResourceGroupNamePfx string
	DeploymentNamePfx    string
	SPClientID           string
	SPObjectID           string
	SPClientSecret       string
	TenantID             string
	TemplateFilePath     string
	ParametersFilePath   string
	Location             string
	MPFMode              string
	ShowDetailedOutput   bool
	JSONOutput           bool
}

func getMPFConfig(mpfArgs MpfCLIArgs) domain.MPFConfig {
	mpfConfig := domain.MPFConfig{
		SubscriptionID: mpfArgs.SubscriptionID,
		TenantID:       mpfArgs.TenantID,
	}
	mpfRole := &domain.Role{}
	mpfRG := &domain.ResourceGroup{}
	mpfSP := &domain.ServicePrincipal{}

	roleDefUUID, _ := uuid.NewRandom()
	mpfRole.RoleDefinitionID = roleDefUUID.String()
	mpfRole.RoleDefinitionName = fmt.Sprintf("tmp-rol-%s", mpfSharedUtils.GenerateRandomString(7))
	mpfRole.RoleDefinitionResourceID = fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Authorization/roleDefinitions/%s", mpfArgs.SubscriptionID, mpfRole.RoleDefinitionID)
	log.Infoln("roleDefinitionResourceID:", mpfRole.RoleDefinitionResourceID)
	mpfRG.ResourceGroupName = fmt.Sprintf("%s-%s", mpfArgs.ResourceGroupNamePfx, mpfSharedUtils.GenerateRandomString(7))
	mpfRG.ResourceGroupResourceID = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", mpfArgs.SubscriptionID, mpfRG.ResourceGroupName)
	mpfRG.Location = mpfArgs.Location
	mpfSP.SPObjectID = mpfArgs.SPObjectID
	mpfSP.SPClientID = mpfArgs.SPClientID
	mpfSP.SPClientSecret = mpfArgs.SPClientSecret

	mpfConfig.Role = *mpfRole
	mpfConfig.ResourceGroup = *mpfRG
	mpfConfig.SP = *mpfSP
	return mpfConfig
}

func getTestingMPFArgs() (MpfCLIArgs, error) {

	subscriptionID := os.Getenv("MPF_SUBSCRIPTIONID")
	servicePrincipalClientID := os.Getenv("MPF_SPCLIENTID")
	servicePrincipalObjectID := os.Getenv("MPF_SPOBJECTID")
	servicePrincipalClientSecret := os.Getenv("MPF_SPCLIENTSECRET")
	tenantID := os.Getenv("MPF_TENANTID")
	resourceGroupNamePfx := "e2eTest"
	deploymentNamePfx := "e2eTest"
	location := "eastus"

	if subscriptionID == "" || servicePrincipalClientID == "" || servicePrincipalObjectID == "" || servicePrincipalClientSecret == "" || tenantID == "" {
		return MpfCLIArgs{}, errors.New("required environment variables not set")
	}

	return MpfCLIArgs{
		SubscriptionID:       subscriptionID,
		ResourceGroupNamePfx: resourceGroupNamePfx,
		DeploymentNamePfx:    deploymentNamePfx,
		SPClientID:           servicePrincipalClientID,
		SPObjectID:           servicePrincipalObjectID,
		SPClientSecret:       servicePrincipalClientSecret,
		TenantID:             tenantID,
		Location:             location,
	}, nil

}

// func TestARMTemplatMultiResourceTemplate(t *testing.T) {

// 	mpfArgs, err := getTestingMPFArgs()
// 	if err != nil {
// 		t.Skip("required environment variables not set, skipping end to end test")
// 	}
// 	mpfArgs.TemplateFilePath = "../samples/templates/multi-resource-template.json"
// 	mpfArgs.ParametersFilePath = "../samples/templates/multi-resource-parameters.json"

// 	ctx := t.Context()

// 	mpfConfig := getMPFConfig(mpfArgs)

// 	deploymentName := fmt.Sprintf("%s-%s", mpfArgs.DeploymentNamePfx, mpfSharedUtils.GenerateRandomString(7))
// 	armConfig := &ARMTemplateShared.ArmTemplateAdditionalConfig{
// 		TemplateFilePath:   mpfArgs.TemplateFilePath,
// 		ParametersFilePath: mpfArgs.ParametersFilePath,
// 		DeploymentName:     deploymentName,
// 	}

// 	// azAPIClient := azureAPI.NewAzureAPIClients(mpfArgs.SubscriptionID)
// 	var rgManager usecase.ResourceGroupManager
// 	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
// 	rgManager = resourceGroupManager.NewResourceGroupManager(mpfArgs.SubscriptionID)
// 	spRoleAssignmentManager = sproleassignmentmanager.NewSPRoleAssignmentManager(mpfArgs.SubscriptionID)

// 	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
// 	var mpfService *usecase.MPFService

// // 	deploymentAuthorizationCheckerCleaner = ARMTemplateWhatIf.NewARMTemplateWhatIfAuthorizationChecker(mpfArgs.SubscriptionID, *armConfig)
// 	initialPermissionsToAdd := []string{"Microsoft.Resources/deployments/*", "Microsoft.Resources/subscriptions/operationresults/read"}
// 	permissionsToAddToResult := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
// 	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, true, false, true, false)

// 	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	//check if mpfResult.RequiredPermissions is not empty and has 54 permissions for scope ResourceGroupResourceID
// 	// Microsoft.Authorization/roleAssignments/read
// 	// Microsoft.Authorization/roleAssignments/write
// 	// Microsoft.Compute/virtualMachines/extensions/read
// 	// Microsoft.Compute/virtualMachines/extensions/write
// 	// Microsoft.Compute/virtualMachines/read
// 	// Microsoft.Compute/virtualMachines/write
// 	// Microsoft.ContainerRegistry/registries/read
// 	// Microsoft.ContainerRegistry/registries/write
// 	// Microsoft.ContainerService/managedClusters/read
// 	// Microsoft.ContainerService/managedClusters/write
// 	// Microsoft.Insights/actionGroups/read
// 	// Microsoft.Insights/actionGroups/write
// 	// Microsoft.Insights/activityLogAlerts/read
// 	// Microsoft.Insights/activityLogAlerts/write
// 	// Microsoft.Insights/diagnosticSettings/read
// 	// Microsoft.Insights/diagnosticSettings/write
// 	// Microsoft.KeyVault/vaults/read
// 	// Microsoft.KeyVault/vaults/write
// 	// Microsoft.ManagedIdentity/userAssignedIdentities/read
// 	// Microsoft.ManagedIdentity/userAssignedIdentities/write
// 	// Microsoft.Network/ApplicationGatewayWebApplicationFirewallPolicies/read
// 	// Microsoft.Network/ApplicationGatewayWebApplicationFirewallPolicies/write
// 	// Microsoft.Network/applicationGateways/read
// 	// Microsoft.Network/applicationGateways/write
// 	// Microsoft.Network/bastionHosts/read
// 	// Microsoft.Network/bastionHosts/write
// 	// Microsoft.Network/natGateways/read
// 	// Microsoft.Network/natGateways/write
// 	// Microsoft.Network/networkInterfaces/read
// 	// Microsoft.Network/networkInterfaces/write
// 	// Microsoft.Network/networkSecurityGroups/read
// 	// Microsoft.Network/networkSecurityGroups/write
// 	// Microsoft.Network/privateDnsZones/read
// 	// Microsoft.Network/privateDnsZones/virtualNetworkLinks/read
// 	// Microsoft.Network/privateDnsZones/virtualNetworkLinks/write
// 	// Microsoft.Network/privateDnsZones/write
// 	// Microsoft.Network/privateEndpoints/privateDnsZoneGroups/read
// 	// Microsoft.Network/privateEndpoints/privateDnsZoneGroups/write
// 	// Microsoft.Network/privateEndpoints/read
// 	// Microsoft.Network/privateEndpoints/write
// 	// Microsoft.Network/publicIPAddresses/read
// 	// Microsoft.Network/publicIPAddresses/write
// 	// Microsoft.Network/publicIPPrefixes/read
// 	// Microsoft.Network/publicIPPrefixes/write
// 	// Microsoft.Network/virtualNetworks/read
// 	// Microsoft.Network/virtualNetworks/write
// 	// Microsoft.OperationalInsights/workspaces/read
// 	// Microsoft.OperationalInsights/workspaces/write
// 	// Microsoft.OperationsManagement/solutions/read
// 	// Microsoft.OperationsManagement/solutions/write
// 	// Microsoft.Resources/deployments/read
// 	// Microsoft.Resources/deployments/write
// 	// Microsoft.Storage/storageAccounts/read
// 	// Microsoft.Storage/storageAccounts/write
// 	assert.NotEmpty(t, mpfResult.RequiredPermissions)
// 	assert.Equal(t, 54, len(mpfResult.RequiredPermissions[mpfConfig.SubscriptionID]))
// }

func TestARMTemplatMultiResourceTemplateFullDeployment(t *testing.T) {

	mpfArgs, err := getTestingMPFArgs()
	if err != nil {
		t.Skip("required environment variables not set, skipping end to end test")
	}
	mpfArgs.TemplateFilePath = "../samples/templates/multi-resource-template.json"
	mpfArgs.ParametersFilePath = "../samples/templates/multi-resource-parameters.json"

	ctx := t.Context()

	mpfConfig := getMPFConfig(mpfArgs)

	deploymentName := fmt.Sprintf("%s-%s", mpfArgs.DeploymentNamePfx, mpfSharedUtils.GenerateRandomString(7))
	armConfig := &ARMTemplateShared.ArmTemplateAdditionalConfig{
		TemplateFilePath:   mpfArgs.TemplateFilePath,
		ParametersFilePath: mpfArgs.ParametersFilePath,
		DeploymentName:     deploymentName,
	}

	// azAPIClient := azureAPI.NewAzureAPIClients(mpfArgs.SubscriptionID)
	var rgManager usecase.ResourceGroupManager
	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
	rgManager = resourceGroupManager.NewResourceGroupManager(mpfArgs.SubscriptionID)
	spRoleAssignmentManager = sproleassignmentmanager.NewSPRoleAssignmentManager(mpfArgs.SubscriptionID)

	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
	var mpfService *usecase.MPFService

	deploymentAuthorizationCheckerCleaner = ARMTemplateDeployment.NewARMTemplateDeploymentAuthorizationChecker(mpfArgs.SubscriptionID, *armConfig)
	initialPermissionsToAdd := []string{"Microsoft.Resources/deployments/*", "Microsoft.Resources/subscriptions/operationresults/read"}
	permissionsToAddToResult := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, true, false, true, false)

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		t.Error(err)
	}

	//check if mpfResult.RequiredPermissions is not empty and has 57 permissions
	// Microsoft.Authorization/roleAssignments/read
	// Microsoft.Authorization/roleAssignments/write
	// Microsoft.Compute/virtualMachines/extensions/read
	// Microsoft.Compute/virtualMachines/extensions/write
	// Microsoft.Compute/virtualMachines/read
	// Microsoft.Compute/virtualMachines/write
	// Microsoft.ContainerRegistry/registries/read
	// Microsoft.ContainerRegistry/registries/write
	// Microsoft.ContainerService/managedClusters/read
	// Microsoft.ContainerService/managedClusters/write
	// Microsoft.Insights/actionGroups/read
	// Microsoft.Insights/actionGroups/write
	// Microsoft.Insights/activityLogAlerts/read
	// Microsoft.Insights/activityLogAlerts/write
	// Microsoft.Insights/diagnosticSettings/read
	// Microsoft.Insights/diagnosticSettings/write
	// Microsoft.KeyVault/vaults/read
	// Microsoft.KeyVault/vaults/write
	// Microsoft.ManagedIdentity/userAssignedIdentities/read
	// Microsoft.ManagedIdentity/userAssignedIdentities/write
	// Microsoft.Network/ApplicationGatewayWebApplicationFirewallPolicies/read
	// Microsoft.Network/ApplicationGatewayWebApplicationFirewallPolicies/write
	// Microsoft.Network/applicationGateways/read
	// Microsoft.Network/applicationGateways/write
	// Microsoft.Network/bastionHosts/read
	// Microsoft.Network/bastionHosts/write
	// Microsoft.Network/natGateways/read
	// Microsoft.Network/natGateways/write
	// Microsoft.Network/networkInterfaces/read
	// Microsoft.Network/networkInterfaces/write
	// Microsoft.Network/networkSecurityGroups/read
	// Microsoft.Network/networkSecurityGroups/write
	// Microsoft.Network/privateDnsZones/read
	// Microsoft.Network/privateDnsZones/virtualNetworkLinks/read
	// Microsoft.Network/privateDnsZones/virtualNetworkLinks/write
	// Microsoft.Network/privateDnsZones/write
	// Microsoft.Network/privateEndpoints/privateDnsZoneGroups/read
	// Microsoft.Network/privateEndpoints/privateDnsZoneGroups/write
	// Microsoft.Network/privateEndpoints/read
	// Microsoft.Network/privateEndpoints/write
	// Microsoft.Network/publicIPAddresses/read
	// Microsoft.Network/publicIPAddresses/write
	// Microsoft.Network/publicIPPrefixes/read
	// Microsoft.Network/publicIPPrefixes/write
	// Microsoft.Network/virtualNetworks/read
	// Microsoft.Network/virtualNetworks/write
	// Microsoft.OperationalInsights/workspaces/listKeys/action
	// Microsoft.OperationalInsights/workspaces/read
	// Microsoft.OperationalInsights/workspaces/sharedKeys/action
	// Microsoft.OperationalInsights/workspaces/write
	// Microsoft.OperationsManagement/solutions/read
	// Microsoft.OperationsManagement/solutions/write
	// Microsoft.Resources/deployments/read
	// Microsoft.Resources/deployments/write
	// Microsoft.Resources/subscriptions/resourceGroups/read
	// Microsoft.Storage/storageAccounts/read
	// Microsoft.Storage/storageAccounts/write
	assert.NotEmpty(t, mpfResult.RequiredPermissions)
	assert.Equal(t, 57, len(mpfResult.RequiredPermissions[mpfConfig.SubscriptionID]))
}

// func TestARMTemplatAksPrivateSubnetTemplate(t *testing.T) {

// 	mpfArgs, err := getTestingMPFArgs()
// 	if err != nil {
// 		t.Skip("required environment variables not set, skipping end to end test")
// 	}
// 	mpfArgs.TemplateFilePath = "../samples/templates/aks-private-subnet.json"
// 	mpfArgs.ParametersFilePath = "../samples/templates/aks-private-subnet-parameters.json"

// 	ctx := t.Context()

// 	mpfConfig := getMPFConfig(mpfArgs)

// 	deploymentName := fmt.Sprintf("%s-%s", mpfArgs.DeploymentNamePfx, mpfSharedUtils.GenerateRandomString(7))
// 	armConfig := &ARMTemplateShared.ArmTemplateAdditionalConfig{
// 		TemplateFilePath:   mpfArgs.TemplateFilePath,
// 		ParametersFilePath: mpfArgs.ParametersFilePath,
// 		DeploymentName:     deploymentName,
// 	}

// 	// azAPIClient := azureAPI.NewAzureAPIClients(mpfArgs.SubscriptionID)
// 	var rgManager usecase.ResourceGroupManager
// 	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
// 	rgManager = resourceGroupManager.NewResourceGroupManager(mpfArgs.SubscriptionID)
// 	spRoleAssignmentManager = sproleassignmentmanager.NewSPRoleAssignmentManager(mpfArgs.SubscriptionID)

// 	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
// 	var mpfService *usecase.MPFService

// 	deploymentAuthorizationCheckerCleaner = ARMTemplateWhatIf.NewARMTemplateWhatIfAuthorizationChecker(mpfArgs.SubscriptionID, *armConfig)
// 	initialPermissionsToAdd := []string{"Microsoft.Resources/deployments/*", "Microsoft.Resources/subscriptions/operationresults/read"}
// 	permissionsToAddToResult := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
// 	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, true, false, true, false)

// 	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	//check if mpfResult.RequiredPermissions is not empty and has 8	 permissions for scope ResourceGroupResourceID
// 	// Microsoft.ContainerService/managedClusters/read
// 	// Microsoft.ContainerService/managedClusters/write
// 	// Microsoft.Network/virtualNetworks/read
// 	// Microsoft.Network/virtualNetworks/subnets/read
// 	// Microsoft.Network/virtualNetworks/subnets/write
// 	// Microsoft.Network/virtualNetworks/write
// 	// Microsoft.Resources/deployments/read
// 	// Microsoft.Resources/deployments/write
// 	assert.NotEmpty(t, mpfResult.RequiredPermissions)
// 	assert.Equal(t, 8, len(mpfResult.RequiredPermissions[mpfConfig.SubscriptionID]))
// }

// func TestARMTemplatAIHub(t *testing.T) {

// 	mpfArgs, err := getTestingMPFArgs()
// 	if err != nil {
// 		t.Skip("required environment variables not set, skipping end to end test")
// 	}
// 	mpfArgs.TemplateFilePath = "../samples/templates/full-deployment-additional/ai-hub.json"
// 	mpfArgs.ParametersFilePath = "../samples/templates/full-deployment-additional/ai-hub.parameters.json"

// 	ctx := t.Context()

// 	mpfConfig := getMPFConfig(mpfArgs)

// 	deploymentName := fmt.Sprintf("%s-%s", mpfArgs.DeploymentNamePfx, mpfSharedUtils.GenerateRandomString(7))
// 	armConfig := &ARMTemplateShared.ArmTemplateAdditionalConfig{
// 		TemplateFilePath:   mpfArgs.TemplateFilePath,
// 		ParametersFilePath: mpfArgs.ParametersFilePath,
// 		DeploymentName:     deploymentName,
// 	}

// 	// azAPIClient := azureAPI.NewAzureAPIClients(mpfArgs.SubscriptionID)
// 	var rgManager usecase.ResourceGroupManager
// 	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
// 	rgManager = resourceGroupManager.NewResourceGroupManager(mpfArgs.SubscriptionID)
// 	spRoleAssignmentManager = sproleassignmentmanager.NewSPRoleAssignmentManager(mpfArgs.SubscriptionID)

// 	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
// 	var mpfService *usecase.MPFService

// 	deploymentAuthorizationCheckerCleaner = ARMTemplateWhatIf.NewARMTemplateWhatIfAuthorizationChecker(mpfArgs.SubscriptionID, *armConfig)
// 	initialPermissionsToAdd := []string{"Microsoft.Resources/deployments/*", "Microsoft.Resources/subscriptions/operationresults/read"}
// 	permissionsToAddToResult := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
// 	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, true, false, true, false)

// 	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	//for certain resources the whatif checker will not be able to get the permissions

// 	//"Microsoft.CognitiveServices/accounts/read"
// 	//"Microsoft.CognitiveServices/accounts/write"
// 	//"Microsoft.MachineLearningServices/workspaces/hubs/read" // Not detected by whatif
// 	//"Microsoft.MachineLearningServices/workspaces/hubs/write" // Not detected by whatif
// 	//"Microsoft.MachineLearningServices/workspaces/read"
// 	//"Microsoft.MachineLearningServices/workspaces/write"
// 	//"Microsoft.Resources/deployments/read"
// 	//"Microsoft.Resources/deployments/write"

// 	assert.NotEmpty(t, mpfResult.RequiredPermissions)
// 	assert.Equal(t, 6, len(mpfResult.RequiredPermissions[mpfConfig.SubscriptionID]))
// }

func TestARMTemplatAksPrivateSubnetTemplateFullDeployment(t *testing.T) {

	mpfArgs, err := getTestingMPFArgs()
	if err != nil {
		t.Skip("required environment variables not set, skipping end to end test")
	}
	mpfArgs.TemplateFilePath = "../samples/templates/aks-private-subnet.json"
	mpfArgs.ParametersFilePath = "../samples/templates/aks-private-subnet-parameters.json"

	ctx := t.Context()

	mpfConfig := getMPFConfig(mpfArgs)

	deploymentName := fmt.Sprintf("%s-%s", mpfArgs.DeploymentNamePfx, mpfSharedUtils.GenerateRandomString(7))
	armConfig := &ARMTemplateShared.ArmTemplateAdditionalConfig{
		TemplateFilePath:   mpfArgs.TemplateFilePath,
		ParametersFilePath: mpfArgs.ParametersFilePath,
		DeploymentName:     deploymentName,
	}

	var rgManager usecase.ResourceGroupManager
	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
	rgManager = resourceGroupManager.NewResourceGroupManager(mpfArgs.SubscriptionID)
	spRoleAssignmentManager = sproleassignmentmanager.NewSPRoleAssignmentManager(mpfArgs.SubscriptionID)

	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
	var mpfService *usecase.MPFService

	deploymentAuthorizationCheckerCleaner = ARMTemplateDeployment.NewARMTemplateDeploymentAuthorizationChecker(mpfArgs.SubscriptionID, *armConfig)
	initialPermissionsToAdd := []string{"Microsoft.Resources/deployments/*", "Microsoft.Resources/subscriptions/operationresults/read"}
	permissionsToAddToResult := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, true, false, true, false)

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		t.Error(err)
	}

	//check if mpfResult.RequiredPermissions is not empty and has 8 permissions for scope ResourceGroupResourceID
	// Microsoft.ContainerService/managedClusters/read
	// Microsoft.ContainerService/managedClusters/write
	// Microsoft.Network/virtualNetworks/read
	// Microsoft.Network/virtualNetworks/subnets/join/action
	// Microsoft.Network/virtualNetworks/subnets/read
	// Microsoft.Network/virtualNetworks/subnets/write
	// Microsoft.Network/virtualNetworks/write
	// Microsoft.Resources/deployments/read
	// Microsoft.Resources/deployments/write
	assert.NotEmpty(t, mpfResult.RequiredPermissions)
	assert.Equal(t, 9, len(mpfResult.RequiredPermissions[mpfConfig.SubscriptionID]))
}

// func TestARMTemplatAIHubFullDeployment(t *testing.T) {

// 	mpfArgs, err := getTestingMPFArgs()
// 	if err != nil {
// 		t.Skip("required environment variables not set, skipping end to end test")
// 	}
// 	mpfArgs.TemplateFilePath = "../samples/templates/full-deployment-additional/ai-hub.json"
// 	mpfArgs.ParametersFilePath = "../samples/templates/full-deployment-additional/ai-hub.parameters.json"

// 	ctx := t.Context()

// 	mpfConfig := getMPFConfig(mpfArgs)

// 	deploymentName := fmt.Sprintf("%s-%s", mpfArgs.DeploymentNamePfx, mpfSharedUtils.GenerateRandomString(7))
// 	armConfig := &ARMTemplateShared.ArmTemplateAdditionalConfig{
// 		TemplateFilePath:   mpfArgs.TemplateFilePath,
// 		ParametersFilePath: mpfArgs.ParametersFilePath,
// 		DeploymentName:     deploymentName,
// 	}

// 	var rgManager usecase.ResourceGroupManager
// 	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
// 	rgManager = resourceGroupManager.NewResourceGroupManager(mpfArgs.SubscriptionID)
// 	spRoleAssignmentManager = sproleassignmentmanager.NewSPRoleAssignmentManager(mpfArgs.SubscriptionID)

// 	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
// 	var mpfService *usecase.MPFService

// 	deploymentAuthorizationCheckerCleaner = ARMTemplateDeployment.NewARMTemplateDeploymentAuthorizationChecker(mpfArgs.SubscriptionID, *armConfig)
// 	initialPermissionsToAdd := []string{"Microsoft.Resources/deployments/*", "Microsoft.Resources/subscriptions/operationresults/read"}
// 	permissionsToAddToResult := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
// 	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, true, false, true, false)

// 	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	//full deployment mode will capture more permissions than whatif
// 	//including the hub-specific permissions that whatif can't detect
// 	//"Microsoft.CognitiveServices/accounts/read"
// 	//"Microsoft.CognitiveServices/accounts/write"
// 	//"Microsoft.MachineLearningServices/workspaces/hubs/read"
// 	//"Microsoft.MachineLearningServices/workspaces/hubs/write"
// 	//"Microsoft.MachineLearningServices/workspaces/read"
// 	//"Microsoft.MachineLearningServices/workspaces/write"
// 	//"Microsoft.Resources/deployments/read"
// 	//"Microsoft.Resources/deployments/write"

// 	assert.NotEmpty(t, mpfResult.RequiredPermissions)
// 	fmt.Printf("Required Permissions: %v\n", mpfResult.RequiredPermissions[mpfConfig.SubscriptionID])
// 	assert.GreaterOrEqual(t, len(mpfResult.RequiredPermissions[mpfConfig.SubscriptionID]), 8)
// }
