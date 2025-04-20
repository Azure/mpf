package e2etests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Azure/mpf/pkg/infrastructure/ARMTemplateShared"
	"github.com/Azure/mpf/pkg/infrastructure/authorizationCheckers/ARMTemplateWhatIf"
	mpfSharedUtils "github.com/Azure/mpf/pkg/infrastructure/mpfSharedUtils"
	spram "github.com/Azure/mpf/pkg/infrastructure/spRoleAssignmentManager"
	"github.com/Azure/mpf/pkg/usecase"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestBicepSubscriptionScoped(t *testing.T) {

	mpfArgs, err := getTestingMPFArgs()
	if err != nil {
		t.Skip("required environment variables not set, skipping end to end test")
	}

	if checkBicepTestEnvVars() {
		t.Skip("required environment variables not set, skipping end to end test")
	}

	bicepExecPath := os.Getenv("MPF_BICEPEXECPATH")
	bicepFilePath := "../samples/bicep/subscription-scope-create-rg.bicep"
	parametersFilePath := "../samples/bicep/subscription-scope-create-rg-params.json"

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

	ctx := t.Context()

	mpfConfig := getMPFConfig(mpfArgs)

	deploymentName := fmt.Sprintf("%s-%s", mpfArgs.DeploymentNamePfx, mpfSharedUtils.GenerateRandomString(7))
	armConfig := &ARMTemplateShared.ArmTemplateAdditionalConfig{
		TemplateFilePath:   armTemplatePath,
		ParametersFilePath: parametersFilePath,
		DeploymentName:     deploymentName,
		SubscriptionScoped: true,
		Location:           "eastus2",
	}

	spRoleAssignmentManager := spram.NewSPRoleAssignmentManager(mpfArgs.SubscriptionID)

	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
	var mpfService *usecase.MPFService

	deploymentAuthorizationCheckerCleaner = ARMTemplateWhatIf.NewARMTemplateWhatIfAuthorizationChecker(mpfArgs.SubscriptionID, *armConfig)
	initialPermissionsToAdd := []string{"Microsoft.Resources/deployments/*", "Microsoft.Resources/subscriptions/operationresults/read"}
	permissionsToAddToResult := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
	mpfService = usecase.NewMPFService(ctx, nil, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, true, false, false)

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		t.Error(err)
	}

	// Check if mpfResult.RequiredPermissions is not empty and has 2 permissions for scope SubscriptionID
	// Microsoft.Resources/deployments/read
	// Microsoft.Resources/deployments/write
	// Microsoft.Resources/subscriptions/resourceGroups/read
	// Microsoft.Resources/subscriptions/resourceGroups/write
	assert.NotEmpty(t, mpfResult.RequiredPermissions)
	assert.Equal(t, 4, len(mpfResult.RequiredPermissions[mpfConfig.SubscriptionID]))
}
