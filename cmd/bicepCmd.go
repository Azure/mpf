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

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Azure/mpf/pkg/domain"
	"github.com/Azure/mpf/pkg/infrastructure/ARMTemplateShared"
	"github.com/Azure/mpf/pkg/infrastructure/authorizationCheckers/ARMTemplateDeployment"
	"github.com/Azure/mpf/pkg/infrastructure/mpfSharedUtils"
	resourceGroupManager "github.com/Azure/mpf/pkg/infrastructure/resourceGroupManager"
	sproleassignmentmanager "github.com/Azure/mpf/pkg/infrastructure/spRoleAssignmentManager"
	"github.com/Azure/mpf/pkg/usecase"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var flgBicepFilePath string
var flgBicepExecPath string

// armCmd represents the arm command

func NewBicepCommand() *cobra.Command {

	bicepCmd := &cobra.Command{
		Use:   "bicep",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
	and usage of using your command. For example:

	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,
		Run: getMPFBicep,
	}

	bicepCmd.Flags().StringVarP(&flgResourceGroupNamePfx, "resourceGroupNamePfx", "", "testdeployrg", "Resource Group Name Prefix")
	bicepCmd.Flags().StringVarP(&flgDeploymentNamePfx, "deploymentNamePfx", "", "testDeploy", "Deployment Name Prefix")

	bicepCmd.Flags().StringVarP(&flgBicepFilePath, "bicepFilePath", "", "", "Path to bicep File")
	err := bicepCmd.MarkFlagRequired("bicepFilePath")
	if err != nil {
		log.Errorf("Error marking flag required for Bicep file path: %v\n", err)
	}

	bicepCmd.Flags().StringVarP(&flgParametersFilePath, "parametersFilePath", "", "", "Path to bicep Parameters File")
	err = bicepCmd.MarkFlagRequired("parametersFilePath")
	if err != nil {
		log.Errorf("Error marking flag required for Bicep parameters file path: %v\n", err)
	}

	bicepCmd.Flags().StringVarP(&flgBicepExecPath, "bicepExecPath", "", "", "Bicep Executable Path")
	err = bicepCmd.MarkFlagRequired("bicepExecPath")
	if err != nil {
		log.Errorf("Error marking flag required for Bicep executable path: %v\n", err)
	}

	bicepCmd.Flags().StringVarP(&flgLocation, "location", "", "eastus2", "Location")

	// Note: fullDeployment flag removed - Full deployment mode is now the only supported mode for Bicep
	// Note: subscriptionScoped flag removed - Only resource group scoped deployments supported with Full deployment mode

	return bicepCmd
}

func getMPFBicep(cmd *cobra.Command, args []string) {
	setLogLevel()

	log.Info("Executing MPF for Bicep")

	log.Debugf("ResourceGroupNamePfx: %s\n", flgResourceGroupNamePfx)
	log.Debugf("DeploymentNamePfx: %s\n", flgDeploymentNamePfx)
	log.Infof("BicepFilePath: %s\n", flgBicepFilePath)
	log.Infof("ParametersFilePath: %s\n", flgParametersFilePath)
	log.Infof("BicepExecPath: %s\n", flgBicepExecPath)
	log.Infof("Location: %s\n", flgLocation)

	// validate if template and parameters files exists
	if _, err := os.Stat(flgBicepFilePath); os.IsNotExist(err) {
		log.Fatal("Bicep File does not exist")
	}

	if _, err := os.Stat(flgBicepExecPath); os.IsNotExist(err) {
		log.Fatal("Bicep Executable does not exist")
	}

	if _, err := os.Stat(flgParametersFilePath); os.IsNotExist(err) {
		log.Fatal("Parameters File does not exist")
	}

	flgBicepExecPath, err := getAbsolutePath(flgBicepExecPath)
	if err != nil {
		log.Errorf("Error getting absolute path for bicep executable: %v\n", err)
	}

	flgBicepFilePath, err := getAbsolutePath(flgBicepFilePath)
	if err != nil {
		log.Errorf("Error getting absolute path for bicep file: %v\n", err)
	}

	flgParametersFilePath, err := getAbsolutePath(flgParametersFilePath)
	if err != nil {
		log.Errorf("Error getting absolute path for parameters file: %v\n", err)
	}

	armTemplatePath := strings.TrimSuffix(flgBicepFilePath, ".bicep") + ".json"
	bicepCmd := exec.Command(flgBicepExecPath, "build", flgBicepFilePath, "--outfile", armTemplatePath)
	bicepCmd.Dir = filepath.Dir(flgBicepFilePath)

	_, err = bicepCmd.CombinedOutput()
	if err != nil {
		log.Errorf("error running bicep build: %s", err)
	}
	log.Infoln("Bicep build successful, ARM Template created at:", armTemplatePath)

	ctx := context.Background()

	mpfConfig := getRootMPFConfig()
	mpfRG := domain.ResourceGroup{}
	mpfRG.ResourceGroupName = fmt.Sprintf("%s-%s", flgResourceGroupNamePfx, mpfSharedUtils.GenerateRandomString(7))
	mpfRG.ResourceGroupResourceID = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", flgSubscriptionID, mpfRG.ResourceGroupName)
	mpfRG.Location = flgLocation
	mpfConfig.ResourceGroup = mpfRG
	deploymentName := fmt.Sprintf("%s-%s", flgDeploymentNamePfx, mpfSharedUtils.GenerateRandomString(7))
	armConfig := &ARMTemplateShared.ArmTemplateAdditionalConfig{
		TemplateFilePath:   armTemplatePath,
		ParametersFilePath: flgParametersFilePath,
		DeploymentName:     deploymentName,
		Location:           flgLocation,
	}

	var rgManager usecase.ResourceGroupManager
	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
	rgManager = resourceGroupManager.NewResourceGroupManager(flgSubscriptionID)
	spRoleAssignmentManager = sproleassignmentmanager.NewSPRoleAssignmentManager(flgSubscriptionID)

	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
	var mpfService *usecase.MPFService
	var initialPermissionsToAdd []string
	var permissionsToAddToResult []string

	// Always use Full Deployment mode - whatif mode has been disabled
	deploymentAuthorizationCheckerCleaner = ARMTemplateDeployment.NewARMTemplateDeploymentAuthorizationChecker(flgSubscriptionID, *armConfig)

	initialPermissionsToAdd = []string{"Microsoft.Resources/deployments/*", "Microsoft.Resources/subscriptions/operationresults/read"}
	permissionsToAddToResult = []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}

	// Add initial permissions from flag if provided (supports comma-separated string or @file.json)
	initialPermissionsToAdd, permissionsToAddToResult = appendUserInitialPermissions(initialPermissionsToAdd, permissionsToAddToResult)

	// Always auto-create resource group since only resource group scoped deployments are supported
	var autoCreateResourceGroup = true
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, true, false, autoCreateResourceGroup)

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		log.Fatal(err)
	}

	log.Infoln("Deleting Generated ARM Template file...")
	// delete generated ARM template file
	err = os.Remove(armTemplatePath)
	if err != nil {
		log.Errorf("Error deleting Generated ARM template file: %v\n", err)
	}

	// log.Infof("Displaying MPF Result: %v\n", mpfResult)
	log.Infof("Show Detailed Output: %t\n", flgShowDetailedOutput)
	log.Infof("JSON Output: %t\n", flgJSONOutput)
	log.Infof("Subscription ID: %s\n", mpfConfig.SubscriptionID)

	displayOptions := getDislayOptions(flgShowDetailedOutput, flgJSONOutput, mpfConfig.SubscriptionID)

	if err != nil {
		if len(mpfResult.RequiredPermissions) > 0 {
			fmt.Println("Error occurred while getting minimum permissions required. However, some permissions were identified prior to the error.")
			displayResult(mpfResult, displayOptions)
		}
		log.Fatal(err)
	}

	displayResult(mpfResult, displayOptions)

}
