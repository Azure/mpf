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
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/manisbindra/az-mpf/pkg/infrastructure/ARMTemplateShared"
	"github.com/manisbindra/az-mpf/pkg/infrastructure/authorizationCheckers/ARMTemplateWhatIf"
	mpfSharedUtils "github.com/manisbindra/az-mpf/pkg/infrastructure/mpfSharedUtils"
	resourceGroupManager "github.com/manisbindra/az-mpf/pkg/infrastructure/resourceGroupManager"
	sproleassignmentmanager "github.com/manisbindra/az-mpf/pkg/infrastructure/spRoleAssignmentManager"
	"github.com/manisbindra/az-mpf/pkg/usecase"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func checkBicepTestEnvVars() bool {
	return os.Getenv("MPF_BICEPEXECPATH") == ""
}

func TestBicepAks(t *testing.T) {

	mpfArgs, err := getTestingMPFArgs()
	if err != nil {
		t.Skip("required environment variables not set, skipping end to end test")
	}

	if checkBicepTestEnvVars() {
		t.Skip("required environment variables not set, skipping end to end test")
	}

	bicepExecPath := os.Getenv("MPF_BICEPEXECPATH")
	bicepFilePath := "../samples/bicep/aks-private-subnet.bicep"
	parametersFilePath := "../samples/bicep/aks-private-subnet-params.json"

	bicepFilePath, _ = getAbsolutePath(bicepFilePath)
	parametersFilePath, _ = getAbsolutePath(parametersFilePath)

	armTemplatePath := strings.TrimSuffix(bicepFilePath, ".bicep") + ".json"
	bicepCmd := exec.Command(bicepExecPath, "build", bicepFilePath, "--outfile", armTemplatePath)
	bicepCmd.Dir = filepath.Dir(bicepFilePath)

	_, err = bicepCmd.CombinedOutput()
	if err != nil {
		log.Error(err)
		t.Error(err)
	}
	// defer os.Remove(armTemplatePath)

	ctx := context.Background()

	mpfConfig := getMPFConfig(mpfArgs)

	deploymentName := fmt.Sprintf("%s-%s", mpfArgs.DeploymentNamePfx, mpfSharedUtils.GenerateRandomString(7))
	armConfig := &ARMTemplateShared.ArmTemplateAdditionalConfig{
		TemplateFilePath:   armTemplatePath,
		ParametersFilePath: parametersFilePath,
		DeploymentName:     deploymentName,
	}

	// azAPIClient := azureAPI.NewAzureAPIClients(mpfArgs.SubscriptionID)
	var rgManager usecase.ResourceGroupManager
	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
	rgManager = resourceGroupManager.NewResourceGroupManager(mpfArgs.SubscriptionID)
	spRoleAssignmentManager = sproleassignmentmanager.NewSPRoleAssignmentManager(mpfArgs.SubscriptionID)

	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
	var mpfService *usecase.MPFService

	deploymentAuthorizationCheckerCleaner = ARMTemplateWhatIf.NewARMTemplateWhatIfAuthorizationChecker(mpfArgs.SubscriptionID, *armConfig)
	initialPermissionsToAdd := []string{"Microsoft.Resources/deployments/*", "Microsoft.Resources/subscriptions/operationresults/read"}
	permissionsToAddToResult := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, true, false, true)

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		t.Error(err)
	}

	//check if mpfResult.RequiredPermissions is not empty and has 8	 permissions for scope ResourceGroupResourceID
	// Microsoft.ContainerService/managedClusters/read
	// Microsoft.ContainerService/managedClusters/write
	// Microsoft.Network/virtualNetworks/read
	// Microsoft.Network/virtualNetworks/subnets/read
	// Microsoft.Network/virtualNetworks/subnets/write
	// Microsoft.Network/virtualNetworks/write
	// Microsoft.Resources/deployments/read
	// Microsoft.Resources/deployments/write
	assert.NotEmpty(t, mpfResult.RequiredPermissions)
	assert.Equal(t, 8, len(mpfResult.RequiredPermissions[mpfConfig.ResourceGroup.ResourceGroupResourceID]))
}

func getAbsolutePath(path string) (string, error) {
	absPath := path
	if !filepath.IsAbs(path) {

		absWorkingDir, err := os.Getwd()
		if err != nil {
			return "", err
		}
		absPath = absWorkingDir + "/" + absPath
	}
	return absPath, nil
}
