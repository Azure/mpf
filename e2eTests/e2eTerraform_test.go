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
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/Azure/mpf/pkg/infrastructure/authorizationCheckers/terraform"
	rgm "github.com/Azure/mpf/pkg/infrastructure/resourceGroupManager"
	spram "github.com/Azure/mpf/pkg/infrastructure/spRoleAssignmentManager"
	"github.com/Azure/mpf/pkg/usecase"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestTerraformACI(t *testing.T) {

	mpfArgs, err := getTestingMPFArgs()
	if err != nil {
		t.Skip("required environment variables not set, skipping end to end test")
	}
	mpfArgs.MPFMode = "terraform"

	var tfpath string
	if os.Getenv("MPF_TFPATH") == "" {
		t.Skip("Terraform Path TF_PATH not set, skipping end to end test")
	}
	tfpath = os.Getenv("MPF_TFPATH")

	_, filename, _, _ := runtime.Caller(0)
	curDir := path.Dir(filename)
	log.Infof("curDir: %s", curDir)
	wrkDir := path.Join(curDir, "../samples/terraform/aci")
	log.Infof("wrkDir: %s", wrkDir)
	varsFile := path.Join(curDir, "../samples/terraform/aci/dev.vars.tfvars")
	log.Infof("varsFile: %s", varsFile)

	ctx := t.Context()

	mpfConfig := getMPFConfig(mpfArgs)

	// azAPIClient := azureAPI.NewAzureAPIClients(mpfArgs.SubscriptionID)
	var rgManager usecase.ResourceGroupManager
	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
	rgManager = rgm.NewResourceGroupManager(mpfArgs.SubscriptionID)
	spRoleAssignmentManager = spram.NewSPRoleAssignmentManager(mpfArgs.SubscriptionID)

	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
	var mpfService *usecase.MPFService

	initialPermissionsToAdd := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
	permissionsToAddToResult := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
	deploymentAuthorizationCheckerCleaner = terraform.NewTerraformAuthorizationChecker(wrkDir, tfpath, varsFile, true, "")
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, false, true, false, false)

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		t.Error(err)
	}

	assert.NotEmpty(t, mpfResult.RequiredPermissions)
	assert.Equal(t, 8, len(mpfResult.RequiredPermissions[mpfConfig.SubscriptionID]))
}

func TestTerraformACINoTfvarsFile(t *testing.T) {

	mpfArgs, err := getTestingMPFArgs()
	if err != nil {
		t.Skip("required environment variables not set, skipping end to end test")
	}
	mpfArgs.MPFMode = "terraform"

	var tfpath string
	if os.Getenv("MPF_TFPATH") == "" {
		t.Skip("Terraform Path TF_PATH not set, skipping end to end test")
	}
	tfpath = os.Getenv("MPF_TFPATH")

	_, filename, _, _ := runtime.Caller(0)
	curDir := path.Dir(filename)
	log.Infof("curDir: %s", curDir)
	wrkDir := path.Join(curDir, "../samples/terraform/rg-no-tfvars")
	log.Infof("wrkDir: %s", wrkDir)
	ctx := t.Context()

	mpfConfig := getMPFConfig(mpfArgs)

	// azAPIClient := azureAPI.NewAzureAPIClients(mpfArgs.SubscriptionID)
	var rgManager usecase.ResourceGroupManager
	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
	rgManager = rgm.NewResourceGroupManager(mpfArgs.SubscriptionID)
	spRoleAssignmentManager = spram.NewSPRoleAssignmentManager(mpfArgs.SubscriptionID)

	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
	var mpfService *usecase.MPFService

	initialPermissionsToAdd := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
	permissionsToAddToResult := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
	deploymentAuthorizationCheckerCleaner = terraform.NewTerraformAuthorizationChecker(wrkDir, tfpath, "", true, "")
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, false, true, false, false)

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		t.Error(err)
	}

	assert.NotEmpty(t, mpfResult.RequiredPermissions)
	assert.Equal(t, 5, len(mpfResult.RequiredPermissions[mpfConfig.SubscriptionID]))
}

func TestTerraformModuleTest(t *testing.T) {

	mpfArgs, err := getTestingMPFArgs()
	if err != nil {
		t.Skip("required environment variables not set, skipping end to end test")
	}
	mpfArgs.MPFMode = "terraform"

	var tfpath string
	if os.Getenv("MPF_TFPATH") == "" {
		t.Skip("Terraform Path TF_PATH not set, skipping end to end test")
	}
	tfpath = os.Getenv("MPF_TFPATH")

	_, filename, _, _ := runtime.Caller(0)
	curDir := path.Dir(filename)
	log.Infof("curDir: %s", curDir)
	wrkDir := path.Join(curDir, "../samples/terraform/module-test")
	log.Infof("wrkDir: %s", wrkDir)
	ctx := t.Context()

	mpfConfig := getMPFConfig(mpfArgs)

	// azAPIClient := azureAPI.NewAzureAPIClients(mpfArgs.SubscriptionID)
	var rgManager usecase.ResourceGroupManager
	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
	rgManager = rgm.NewResourceGroupManager(mpfArgs.SubscriptionID)
	spRoleAssignmentManager = spram.NewSPRoleAssignmentManager(mpfArgs.SubscriptionID)

	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
	var mpfService *usecase.MPFService

	initialPermissionsToAdd := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
	permissionsToAddToResult := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
	deploymentAuthorizationCheckerCleaner = terraform.NewTerraformAuthorizationChecker(wrkDir, tfpath, "", true, "")
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, false, true, false, false)

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		t.Error(err)
	}

	assert.NotEmpty(t, mpfResult.RequiredPermissions)
	assert.Equal(t, 8, len(mpfResult.RequiredPermissions[mpfConfig.SubscriptionID]))
}

//
// Below test can take more than 10 mins to complete, hence commented
//

// func TestTerraformMultiResource(t *testing.T) {

// 	mpfArgs, err := getTestingMPFArgs()
// 	if err != nil {
// 		t.Skip("required environment variables not set, skipping end to end test")
// 	}
// 	mpfArgs.MPFMode = "terraform"

// 	var tfpath string
// 	if os.Getenv("MPF_TFPATH") == "" {
// 		t.Skip("Terraform Path TF_PATH not set, skipping end to end test")
// 	}
// 	tfpath = os.Getenv("MPF_TFPATH")

// 	_, filename, _, _ := runtime.Caller(0)
// 	curDir := path.Dir(filename)
// 	wrkDir := path.Join(curDir, "../samples/terraform/multi-resource")
// 	varsFile := path.Join(curDir, "../samples/terraform/multi-resource/dev.vars.tfvars")

// 	ctx := t.Context()

// 	mpfConfig := presentation.getMPFConfig(mpfArgs)

// 	// azAPIClient := azureAPI.NewAzureAPIClients(mpfArgs.SubscriptionID)
// 	var rgManager usecase.ResourceGroupManager
// 	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
// 	rgManager = rgm.NewResourceGroupManager(mpfArgs.SubscriptionID)
// 	spRoleAssignmentManager = spram.NewSPRoleAssignmentManager(mpfArgs.SubscriptionID)

// 	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
// 	var mpfService *usecase.MPFService

// 	initialPermissionsToAdd := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
// 	permissionsToAddToResult := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
// 	deploymentAuthorizationCheckerCleaner = terraform.NewTerraformAuthorizationChecker(wrkDir, tfpath, varsFile)
// 	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, false, true, false, false)

// 	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	// Microsoft.ContainerService/managedClusters/delete
// 	// Microsoft.ContainerService/managedClusters/listClusterUserCredential/action
// 	// Microsoft.ContainerService/managedClusters/read
// 	// Microsoft.ContainerService/managedClusters/write
// 	// Microsoft.Network/virtualNetworks/delete
// 	// Microsoft.Network/virtualNetworks/read
// 	// Microsoft.Network/virtualNetworks/subnets/delete
// 	// Microsoft.Network/virtualNetworks/subnets/join/action
// 	// Microsoft.Network/virtualNetworks/subnets/read
// 	// Microsoft.Network/virtualNetworks/subnets/write
// 	// Microsoft.Network/virtualNetworks/write
// 	// Microsoft.Resources/deployments/read
// 	// Microsoft.Resources/deployments/write
// 	// Microsoft.Resources/subscriptions/resourcegroups/delete
// 	// Microsoft.Resources/subscriptions/resourcegroups/read
// 	// Microsoft.Resources/subscriptions/resourcegroups/write
// 	//check if mpfResult.RequiredPermissions is not empty and has 54 permissions for scope ResourceGroupResourceID
// 	assert.NotEmpty(t, mpfResult.RequiredPermissions)
// 	assert.Equal(t, 16, len(mpfResult.RequiredPermissions[mpfConfig.SubscriptionID]))
// }
