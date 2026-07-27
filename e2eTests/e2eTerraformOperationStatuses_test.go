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
	"strings"
	"testing"

	"github.com/Azure/mpf/pkg/infrastructure/authorizationCheckers/terraform"
	rgm "github.com/Azure/mpf/pkg/infrastructure/resourceGroupManager"
	spram "github.com/Azure/mpf/pkg/infrastructure/spRoleAssignmentManager"
	"github.com/Azure/mpf/pkg/usecase"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestTerraformACRWithOperationStatusesReadPermissions verifies that the LRO polling permission is
// reported for a resource type which exposes an operationStatuses action.
//
// Microsoft.ContainerRegistry/registries is created through a long running operation, so the
// azurerm provider polls the operationStatuses URL after the PUT returns. The registry therefore
// needs Microsoft.ContainerRegistry/registries/operationStatuses/read on top of the write
// permission.
// See https://github.com/Azure/mpf/issues/62
func TestTerraformACRWithOperationStatusesReadPermissions(t *testing.T) {
	mpfArgs, err := getTestingMPFArgs()
	if err != nil {
		t.Skip("required environment variables not set, skipping end to end test")
	}
	mpfArgs.MPFMode = "terraform"

	if os.Getenv("MPF_TFPATH") == "" {
		t.Skip("Terraform Path MPF_TFPATH not set, skipping end to end test")
	}
	tfpath := os.Getenv("MPF_TFPATH")

	_, filename, _, _ := runtime.Caller(0)
	curDir := path.Dir(filename)
	wrkDir := path.Join(curDir, "../samples/terraform/acr")
	log.Infof("wrkDir: %s", wrkDir)

	cleanTerraformWorkingDir(t, wrkDir)
	t.Cleanup(func() { cleanTerraformWorkingDir(t, wrkDir) })

	varsFile := path.Join(wrkDir, "dev.vars.tfvars")

	ctx := t.Context()
	mpfConfig := getMPFConfig(mpfArgs)

	var rgManager usecase.ResourceGroupManager = rgm.NewResourceGroupManager(mpfArgs.SubscriptionID)
	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager = spram.NewSPRoleAssignmentManager(mpfArgs.SubscriptionID)

	initialPermissionsToAdd := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
	permissionsToAddToResult := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}

	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner = terraform.NewTerraformAuthorizationChecker(wrkDir, tfpath, varsFile, true, "")
	mpfService := usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, false, true, false,
		usecase.WithAutoAddOperationStatusesReadForWrite(true))

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		t.Error(err)
	}

	assert.NotEmpty(t, mpfResult.RequiredPermissions)
	perms := mpfResult.RequiredPermissions[mpfConfig.SubscriptionID]
	log.Infof("Found %d permissions: %v", len(perms), perms)

	// The registry write permission and its LRO polling permission must both be reported
	assert.Contains(t, perms, "Microsoft.ContainerRegistry/registries/write")
	assert.Contains(t, perms, "Microsoft.ContainerRegistry/registries/operationStatuses/read")

	// Resource types without an operationStatuses action must not pollute the result, the
	// candidates that Azure rejects as invalid actions are filtered out again
	assert.NotContains(t, perms, "Microsoft.Resources/subscriptions/resourcegroups/operationStatuses/read")
	for _, perm := range perms {
		if strings.Contains(perm, "/operationStatuses/") {
			assert.True(t, strings.HasPrefix(perm, "Microsoft.ContainerRegistry/registries/"),
				"unexpected operationStatuses permission in the result: %s", perm)
		}
	}
}
