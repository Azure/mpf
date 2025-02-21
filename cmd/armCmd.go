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

	"github.com/manisbindra/az-mpf/pkg/domain"
	"github.com/manisbindra/az-mpf/pkg/infrastructure/ARMTemplateShared"
	"github.com/manisbindra/az-mpf/pkg/infrastructure/authorizationCheckers/ARMTemplateWhatIf"
	"github.com/manisbindra/az-mpf/pkg/infrastructure/mpfSharedUtils"
	resourceGroupManager "github.com/manisbindra/az-mpf/pkg/infrastructure/resourceGroupManager"
	sproleassignmentmanager "github.com/manisbindra/az-mpf/pkg/infrastructure/spRoleAssignmentManager"
	"github.com/manisbindra/az-mpf/pkg/presentation"
	"github.com/manisbindra/az-mpf/pkg/usecase"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var flgResourceGroupNamePfx string
var flgDeploymentNamePfx string
var flgLocation string
var flgTemplateFilePath string
var flgParametersFilePath string

// var flgFullDeployment bool

// armCmd represents the arm command

func NewARMCommand() *cobra.Command {

	armCmd := &cobra.Command{
		Use:   "arm",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
	and usage of using your command. For example:
	
	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,
		Run: getMPFARM,
	}

	armCmd.Flags().StringVarP(&flgResourceGroupNamePfx, "resourceGroupNamePfx", "", "testdeployrg", "Resource Group Name Prefix")
	armCmd.Flags().StringVarP(&flgDeploymentNamePfx, "deploymentNamePfx", "", "testDeploy", "Deployment Name Prefix")

	armCmd.Flags().StringVarP(&flgTemplateFilePath, "templateFilePath", "", "", "Path to ARM Template File")
	err := armCmd.MarkFlagRequired("templateFilePath")
	if err != nil {
		log.Errorf("Error marking flag required for ARM template file path: %v\n", err)
	}

	armCmd.Flags().StringVarP(&flgParametersFilePath, "parametersFilePath", "", "", "Path to Template Parameters File")
	err = armCmd.MarkFlagRequired("parametersFilePath")
	if err != nil {
		log.Errorf("Error marking flag required for ARM template parameters file path: %v\n", err)
	}

	armCmd.Flags().StringVarP(&flgLocation, "location", "", "eastus", "Location")

	// armCmd.Flags().BoolVarP(&flgFullDeployment, "fullDeployment", "", false, "Full Deployment")

	return armCmd
}

func getMPFARM(cmd *cobra.Command, args []string) {
	setLogLevel()

	log.Info("Executing MPF for ARM")

	log.Debugf("ResourceGroupNamePfx: %s\n", flgResourceGroupNamePfx)
	log.Debugf("DeploymentNamePfx: %s\n", flgDeploymentNamePfx)
	log.Infof("TemplateFilePath: %s\n", flgTemplateFilePath)
	log.Infof("ParametersFilePath: %s\n", flgParametersFilePath)

	// validate if template and parameters files exists
	if _, err := os.Stat(flgTemplateFilePath); os.IsNotExist(err) {
		log.Fatal("Template File does not exist")
	}

	flgTemplateFilePath, err := getAbsolutePath(flgTemplateFilePath)
	if err != nil {
		log.Errorf("Error getting absolute path for ARM template file: %v\n", err)
	}

	if _, err := os.Stat(flgParametersFilePath); os.IsNotExist(err) {
		log.Fatal("Parameters File does not exist")
	}

	flgParametersFilePath, err := getAbsolutePath(flgParametersFilePath)
	if err != nil {
		log.Errorf("Error getting absolute path for ARM template parameters file: %v\n", err)
	}

	ctx := context.Background()

	mpfConfig := getRootMPFConfig()
	mpfRG := domain.ResourceGroup{}
	mpfRG.ResourceGroupName = fmt.Sprintf("%s-%s", flgResourceGroupNamePfx, mpfSharedUtils.GenerateRandomString(7))
	mpfRG.ResourceGroupResourceID = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", flgSubscriptionID, mpfRG.ResourceGroupName)
	mpfRG.Location = flgLocation
	mpfConfig.ResourceGroup = mpfRG
	deploymentName := fmt.Sprintf("%s-%s", flgDeploymentNamePfx, mpfSharedUtils.GenerateRandomString(7))
	armConfig := &ARMTemplateShared.ArmTemplateAdditionalConfig{
		TemplateFilePath:   flgTemplateFilePath,
		ParametersFilePath: flgParametersFilePath,
		DeploymentName:     deploymentName,
	}

	var rgManager usecase.ResourceGroupManager
	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
	rgManager = resourceGroupManager.NewResourceGroupManager(flgSubscriptionID)
	spRoleAssignmentManager = sproleassignmentmanager.NewSPRoleAssignmentManager(flgSubscriptionID)

	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
	var mpfService *usecase.MPFService
	var initialPermissionsToAdd []string
	var permissionsToAddToResult []string

	deploymentAuthorizationCheckerCleaner = ARMTemplateWhatIf.NewARMTemplateWhatIfAuthorizationChecker(flgSubscriptionID, *armConfig)
	initialPermissionsToAdd = []string{"Microsoft.Resources/deployments/*", "Microsoft.Resources/subscriptions/operationresults/read"}
	permissionsToAddToResult = []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}

	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, true, false, true)

	displayOptions := getDislayOptions(flgShowDetailedOutput, flgJSONOutput, mpfConfig.ResourceGroup.ResourceGroupResourceID)

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		if len(mpfResult.RequiredPermissions) > 0 {
			fmt.Println("Error occurred while getting minimum permissions required. However, some permissions were identified prior to the error.")
			displayResult(mpfResult, displayOptions)
		}
		log.Fatal(err)
	}

	displayResult(mpfResult, displayOptions)
}

func getDislayOptions(flgShowDetailedOutput bool, flgJSONOutput bool, rgResourceId string) presentation.DisplayOptions {
	return presentation.DisplayOptions{
		ShowDetailedOutput:             flgShowDetailedOutput,
		JSONOutput:                     flgJSONOutput,
		DefaultResourceGroupResourceID: rgResourceId,
	}
}

func displayResult(mpfResult domain.MPFResult, displayOptions presentation.DisplayOptions) {
	resultDisplayer := presentation.NewMPFResultDisplayer(mpfResult, displayOptions)
	err := resultDisplayer.DisplayResult(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}
