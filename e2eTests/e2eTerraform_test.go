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
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Azure/mpf/pkg/infrastructure/authorizationCheckers/terraform"
	rgm "github.com/Azure/mpf/pkg/infrastructure/resourceGroupManager"
	spram "github.com/Azure/mpf/pkg/infrastructure/spRoleAssignmentManager"
	"github.com/Azure/mpf/pkg/usecase"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// cleanTerraformWorkingDir removes Terraform state files and cache directories
// from the working directory to ensure each test starts with a clean state.
// This prevents test pollution from leftover state files between test runs.
func cleanTerraformWorkingDir(t *testing.T, workingDir string) {
	t.Helper()

	// Files and directories to remove
	artifactsToRemove := []string{
		".terraform",
		".terraform.lock.hcl",
		"terraform.tfstate",
		"terraform.tfstate.backup",
		".terraform.tfstate.lock.info",
		"terraform.log",
		".azmpfEnteredDestroyPhase.txt",
	}

	for _, artifact := range artifactsToRemove {
		artifactPath := filepath.Join(workingDir, artifact)
		if err := os.RemoveAll(artifactPath); err != nil {
			// Log warning but don't fail - the artifact may not exist
			log.Debugf("could not remove %s: %v", artifactPath, err)
		}
	}

	log.Infof("cleaned Terraform artifacts from: %s", workingDir)
}

func TestTerraformACI(t *testing.T) {
	mpfArgs, err := getTestingMPFArgs()
	if err != nil {
		t.Skip("required environment variables not set, skipping end to end test")
	}
	mpfArgs.MPFMode = "terraform"

	var tfpath string
	if os.Getenv("MPF_TFPATH") == "" {
		t.Skip("Terraform Path MPF_TFPATH not set, skipping end to end test")
	}
	tfpath = os.Getenv("MPF_TFPATH")

	_, filename, _, _ := runtime.Caller(0)
	curDir := path.Dir(filename)
	log.Infof("curDir: %s", curDir)
	wrkDir := path.Join(curDir, "../samples/terraform/aci")
	log.Infof("wrkDir: %s", wrkDir)

	// Clean up Terraform artifacts before and after test
	cleanTerraformWorkingDir(t, wrkDir)
	t.Cleanup(func() { cleanTerraformWorkingDir(t, wrkDir) })

	varsFile := path.Join(wrkDir, "dev.vars.tfvars")
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
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, false, true, false)

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
		t.Skip("Terraform Path MPF_TFPATH not set, skipping end to end test")
	}
	tfpath = os.Getenv("MPF_TFPATH")

	_, filename, _, _ := runtime.Caller(0)
	curDir := path.Dir(filename)
	log.Infof("curDir: %s", curDir)
	wrkDir := path.Join(curDir, "../samples/terraform/rg-no-tfvars")
	log.Infof("wrkDir: %s", wrkDir)

	// Clean up Terraform artifacts before and after test
	cleanTerraformWorkingDir(t, wrkDir)
	t.Cleanup(func() { cleanTerraformWorkingDir(t, wrkDir) })

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
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, false, true, false)

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
		t.Skip("Terraform Path MPF_TFPATH not set, skipping end to end test")
	}
	tfpath = os.Getenv("MPF_TFPATH")

	_, filename, _, _ := runtime.Caller(0)
	curDir := path.Dir(filename)
	log.Infof("curDir: %s", curDir)
	wrkDir := path.Join(curDir, "../samples/terraform/module-test")
	log.Infof("wrkDir: %s", wrkDir)

	// Clean up Terraform artifacts before and after test
	cleanTerraformWorkingDir(t, wrkDir)
	t.Cleanup(func() { cleanTerraformWorkingDir(t, wrkDir) })

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
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, false, true, false)

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		t.Error(err)
	}

	// Microsoft.OperationalInsights/workspaces/delete
	// Microsoft.OperationalInsights/workspaces/read
	// Microsoft.OperationalInsights/workspaces/write
	// Microsoft.Resources/deployments/read
	// Microsoft.Resources/deployments/write
	// Microsoft.Resources/subscriptions/resourcegroups/delete
	// Microsoft.Resources/subscriptions/resourcegroups/read
	// Microsoft.Resources/subscriptions/resourcegroups/write

	assert.NotEmpty(t, mpfResult.RequiredPermissions)
	perms := mpfResult.RequiredPermissions[mpfConfig.SubscriptionID]
	log.Infof("Found %d permissions: %v", len(perms), perms)
	assert.Equal(t, 8, len(perms))
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
// 		t.Skip("Terraform Path MPF_TFPATH not set, skipping end to end test")
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
// 	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, false, true, false)

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
