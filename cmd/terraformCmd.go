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
	"os"

	"github.com/manisbindra/az-mpf/pkg/infrastructure/authorizationCheckers/terraform"
	resourceGroupManager "github.com/manisbindra/az-mpf/pkg/infrastructure/resourceGroupManager"
	sproleassignmentmanager "github.com/manisbindra/az-mpf/pkg/infrastructure/spRoleAssignmentManager"
	"github.com/manisbindra/az-mpf/pkg/presentation"
	"github.com/manisbindra/az-mpf/pkg/usecase"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	flgTFPath      string
	flgWorkingDir  string
	flgVarFilePath string
)

// terraformCmd represents the terraform command
// var

func NewTerraformCommand() *cobra.Command {
	terraformCmd := &cobra.Command{
		Use:   "terraform",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
	and usage of using your command. For example:
	
	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,
		Run: getMPFTerraform,
	}

	terraformCmd.Flags().StringVarP(&flgTFPath, "tfPath", "", "", "Path to Terraform Executable")
	// viper.BindEnv("executablePath", "TF_PATH")
	terraformCmd.MarkFlagRequired("tfPath")

	terraformCmd.Flags().StringVarP(&flgWorkingDir, "workingDir", "", "", "Path to Terraform Working Directory")
	terraformCmd.MarkFlagRequired("workingDir")

	terraformCmd.Flags().StringVarP(&flgVarFilePath, "varFilePath", "", "", "Path to Terraform Variable File")
	terraformCmd.MarkFlagRequired("varFilePath")

	return terraformCmd
}

func getMPFTerraform(cmd *cobra.Command, args []string) {
	setLogLevel()

	log.Info("Executin MPF for Terraform")
	log.Infof("TFPath: %s\n", flgTFPath)
	log.Infof("WorkingDir: %s\n", flgWorkingDir)
	log.Infof("VarFilePath: %s\n", flgVarFilePath)

	// validate if working directory exists
	if _, err := os.Stat(flgWorkingDir); os.IsNotExist(err) {
		log.Fatalf("Working Directory does not exist: %s\n", flgWorkingDir)
	}

	flgWorkingDir, err := getAbsolutePath(flgWorkingDir)
	if err != nil {
		log.Errorf("Error getting absolute path for terraform working directory: %v\n", err)
	}

	// validate if tfPath exists
	if _, err := os.Stat(flgTFPath); os.IsNotExist(err) {
		log.Fatalf("Terraform Executable does not exist: %s\n", flgTFPath)
	}

	flgTFPath, err := getAbsolutePath(flgTFPath)
	if err != nil {
		log.Errorf("Error getting absolute path for terraform executable: %v\n", err)
	}

	if _, err := os.Stat(flgVarFilePath); os.IsNotExist(err) {
		log.Fatalf("Terraform Variable File does not exist: %s\n", flgVarFilePath)
	}

	flgVarFilePath, err := getAbsolutePath(flgVarFilePath)
	if err != nil {
		log.Errorf("Error getting absolute path for terraform variable file: %v\n", err)
	}

	ctx := context.Background()

	mpfConfig := getRootMPFConfig()

	var rgManager usecase.ResourceGroupManager
	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
	rgManager = resourceGroupManager.NewResourceGroupManager(flgSubscriptionID)
	spRoleAssignmentManager = sproleassignmentmanager.NewSPRoleAssignmentManager(flgSubscriptionID)

	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
	var mpfService *usecase.MPFService

	initialPermissionsToAdd := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
	permissionsToAddToResult := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
	deploymentAuthorizationCheckerCleaner = terraform.NewTerraformAuthorizationChecker(flgWorkingDir, flgTFPath, flgVarFilePath)
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, false, true, false)

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		log.Fatal(err)
	}

	displayOptions := presentation.DisplayOptions{
		ShowDetailedOutput:             flgShowDetailedOutput,
		JSONOutput:                     flgJSONOutput,
		DefaultResourceGroupResourceID: mpfConfig.ResourceGroup.ResourceGroupResourceID,
	}

	resultDisplayer := presentation.NewMPFResultDisplayer(mpfResult, displayOptions)
	err = resultDisplayer.DisplayResult(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

}
