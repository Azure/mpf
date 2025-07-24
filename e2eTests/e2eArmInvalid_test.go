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
	"testing"

	"github.com/Azure/mpf/pkg/infrastructure/ARMTemplateShared"
	"github.com/Azure/mpf/pkg/infrastructure/authorizationCheckers/ARMTemplateDeployment"
	mpfSharedUtils "github.com/Azure/mpf/pkg/infrastructure/mpfSharedUtils"
	resourceGroupManager "github.com/Azure/mpf/pkg/infrastructure/resourceGroupManager"
	sproleassignmentmanager "github.com/Azure/mpf/pkg/infrastructure/spRoleAssignmentManager"
	"github.com/Azure/mpf/pkg/usecase"
	"github.com/stretchr/testify/assert"
)

// func TestARMTemplatWhatIfInvalidParams(t *testing.T) {

// 	mpfArgs, err := getTestingMPFArgs()
// 	if err != nil {
// 		t.Skip("required environment variables not set, skipping end to end test")
// 	}
// 	mpfArgs.TemplateFilePath = "../samples/templates/aks.json"
// 	mpfArgs.ParametersFilePath = "../samples/templates/aks-invalid-params.json"

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
// 	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, true, false, true)

// 	_, err = mpfService.GetMinimumPermissionsRequired()
// 	assert.Error(t, err)
// 	// assert that error is InvalidTemplate error
// 	// assert.Equal(t, "InvalidTemplate", err.Error())
// 	if !errors.Is(err, ARMTemplateShared.ErrInvalidTemplate) {
// 		t.Errorf("Error is not of type InvalidTemplate")
// 	}

// }

// func TestARMTemplatWhatIfInvalidTemplate(t *testing.T) {

// 	mpfArgs, err := getTestingMPFArgs()
// 	if err != nil {
// 		t.Skip("required environment variables not set, skipping end to end test")
// 	}
// 	mpfArgs.TemplateFilePath = "../samples/templates/aks-invalid-template.json"
// 	mpfArgs.ParametersFilePath = "../samples/templates/aks-parameters.json"

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
// 	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, true, false, true)

// 	_, err = mpfService.GetMinimumPermissionsRequired()
// 	assert.Error(t, err)
// 	// assert that error is InvalidTemplate error
// 	// assert.Equal(t, "InvalidTemplate", err.Error())
// 	if !errors.Is(err, ARMTemplateShared.ErrInvalidTemplate) {
// 		t.Errorf("Error is not of type InvalidTemplate")
// 	}

// }

func TestARMTemplatDeploymentInvalidParams(t *testing.T) {

	mpfArgs, err := getTestingMPFArgs()
	if err != nil {
		t.Skip("required environment variables not set, skipping end to end test")
	}
	mpfArgs.TemplateFilePath = "../samples/templates/aks.json"
	mpfArgs.ParametersFilePath = "../samples/templates/aks-invalid-params.json"

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
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, true, false, true)

	_, err = mpfService.GetMinimumPermissionsRequired()
	assert.Error(t, err)
	// assert that error is InvalidTemplate error
	if !errors.Is(err, ARMTemplateShared.ErrInvalidTemplate) {
		t.Errorf("Error is not of type InvalidTemplate")
	}

}

func TestARMTemplatDeploymentInvalidTemplate(t *testing.T) {

	mpfArgs, err := getTestingMPFArgs()
	if err != nil {
		t.Skip("required environment variables not set, skipping end to end test")
	}
	mpfArgs.TemplateFilePath = "../samples/templates/aks-invalid-template.json"
	mpfArgs.ParametersFilePath = "../samples/templates/aks-parameters.json"

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
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, true, false, true)

	_, err = mpfService.GetMinimumPermissionsRequired()
	assert.Error(t, err)
	// assert that error is InvalidTemplate error
	if !errors.Is(err, ARMTemplateShared.ErrInvalidTemplate) {
		t.Errorf("Error is not of type InvalidTemplate")
	}

}

func TestARMTemplatDeploymentBlankTemplateAndParams(t *testing.T) {

	mpfArgs, err := getTestingMPFArgs()
	if err != nil {
		t.Skip("required environment variables not set, skipping end to end test")
	}
	mpfArgs.TemplateFilePath = "../samples/templates/blank-template.json"
	mpfArgs.ParametersFilePath = "../samples/templates/blank-params.json"

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
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, true, false, true)

	_, err = mpfService.GetMinimumPermissionsRequired()
	assert.Error(t, err)
	// assert that error is InvalidTemplate error
	if !errors.Is(err, ARMTemplateShared.ErrInvalidTemplate) {
		t.Errorf("Error is not of type InvalidTemplate")
	}

}
