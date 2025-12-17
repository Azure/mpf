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

package ARMTemplateDeployment

import (
	"context"
	"errors"
	"fmt"
	"time"

	// "log"
	"strings"

	"github.com/Azure/mpf/pkg/infrastructure/ARMTemplateShared"
	"github.com/Azure/mpf/pkg/infrastructure/azureAPI"
	"github.com/Azure/mpf/pkg/infrastructure/mpfSharedUtils"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/mpf/pkg/domain"
	log "github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

type armDeploymentConfig struct {
	ctx         context.Context
	armConfig   ARMTemplateShared.ArmTemplateAdditionalConfig
	azAPIClient *azureAPI.AzureAPIClients
}

func NewARMTemplateDeploymentAuthorizationChecker(subscriptionID string, armConfig ARMTemplateShared.ArmTemplateAdditionalConfig) *armDeploymentConfig {
	azAPIClient := azureAPI.NewAzureAPIClients(subscriptionID)
	return &armDeploymentConfig{
		azAPIClient: azAPIClient,
		armConfig:   armConfig,
		ctx:         context.Background(),
	}

}

func (a *armDeploymentConfig) GetDeploymentAuthorizationErrors(mpfConfig domain.MPFConfig) (string, error) {
	// return a.deployARMTemplate(a.armConfig.DeploymentName, mpfConfig)
	return a.deployARMTemplatev2(a.armConfig.DeploymentName, mpfConfig)
}

func (a *armDeploymentConfig) CleanDeployment(mpfConfig domain.MPFConfig) error {
	log.Infoln("Cleaning up resources...")
	log.Infoln("*************************")

	// Cancel deployment. Even if cancelling deployment fails attempt to delete other resources
	_ = a.cancelDeployment(a.ctx, a.armConfig.DeploymentName, mpfConfig)

	return nil
}

func (a *armDeploymentConfig) deployARMTemplatev2(deploymentName string, mpfConfig domain.MPFConfig) (string, error) {

	cred, err := azidentity.NewClientSecretCredential(mpfConfig.TenantID, mpfConfig.SP.SPClientID, mpfConfig.SP.SPClientSecret, nil)
	if err != nil {
		return "", fmt.Errorf("error creating client secret credential: %w", err)
	}

	defer func() {
		if cred != nil {
			// cred.Close()
			cred = nil
		}
	}()

	ctx := context.Background()
	clientFactory, err := armresources.NewClientFactory(mpfConfig.SubscriptionID, cred, nil)
	if err != nil {
		return "", fmt.Errorf("error creating client factory: %w", err)
	}

	// read template and parameters
	template, err := mpfSharedUtils.ReadJson(a.armConfig.TemplateFilePath)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ARMTemplateShared.ErrInvalidTemplate, fmt.Errorf("error reading template file: %w", err))
	}

	parameters, err := mpfSharedUtils.ReadJson(a.armConfig.ParametersFilePath)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ARMTemplateShared.ErrInvalidTemplate, fmt.Errorf("error reading parameters file: %w", err))
	}

	// convert parameters to standard format
	parameters = ARMTemplateShared.GetParametersInStandardFormat(parameters)

	// fullTemplate := map[string]interface{}{
	// 	"properties": map[string]interface{}{
	// 		"mode":       "Incremental",
	// 		"template":   template,
	// 		"parameters": parameters,
	// 	},
	// }

	// // convert bodyJSON to string
	// fullTemplateJSONBytes, err := json.Marshal(fullTemplate)
	// if err != nil {
	// 	return "", fmt.Errorf("%w, %w", ARMTemplateShared.ErrInvalidTemplate, fmt.Errorf("error marshalling fullTemplateJSON: %w", err))
	// }

	// fullTemplateJSONString := string(fullTemplateJSONBytes)

	// log.Debugln()
	// log.Debugln(fullTemplateJSONString)
	// log.Debugln()

	poller, err := clientFactory.NewDeploymentsClient().BeginCreateOrUpdate(ctx, mpfConfig.ResourceGroup.ResourceGroupName, deploymentName, armresources.Deployment{
		Properties: &armresources.DeploymentProperties{
			Mode:       to.Ptr(armresources.DeploymentModeIncremental),
			Parameters: parameters,
			Template:   template,
		},
	}, nil)

	if err != nil {
		errMesg := err.Error()

		// Check for different types of errors and handle accordingly
		switch {
		case strings.Contains(errMesg, "AuthorizationFailed") ||
			strings.Contains(errMesg, "Authorization failed") ||
			strings.Contains(errMesg, "AuthorizationPermissionMismatch") ||
			strings.Contains(errMesg, "LinkedAccessCheckFailed"):
			return errMesg, nil

		case strings.Contains(errMesg, "InvalidTemplate") && !strings.Contains(errMesg, "InvalidTemplateDeployment"):
			log.Warnf("Error message: %s", errMesg)
			return "", fmt.Errorf("%w: please check the template and parameters file: %s", ARMTemplateShared.ErrInvalidTemplate, errMesg)

		case strings.Contains(errMesg, "InvalidRequestContent") && !strings.Contains(errMesg, "InvalidTemplateDeployment"):
			log.Warnf("Error message: %s", errMesg)
			return "", fmt.Errorf("%w: please check the template and parameters file: %s", ARMTemplateShared.ErrInvalidTemplate, errMesg)

		case strings.Contains(errMesg, "InvalidTemplateDeployment"):
			// This indicates all Authorization errors are fixed
			// Sample error [{\"code\":\"PodIdentityAddonFeatureFlagNotEnabled\",\"message\":\"Provisioning of resource(s) for container service aks-24xalwx7i2ueg in resource group testdeployrg-Y2jsRAG failed. Message: PodIdentity addon is not allowed since feature 'Microsoft.ContainerService/EnablePodIdentityPreview' is not enabled.
			// Hence ok to proceed, and not return error in this condition
			log.Warnf("Post Authorization error occurred: %s", errMesg)
			return "", nil
		}
	}

	if poller == nil {
		return "", nil
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		errMesg := err.Error()
		log.Debugf("Error message: %s", errMesg)

		// Check for different types of errors and handle accordingly
		switch {
		case strings.Contains(errMesg, "AuthorizationFailed") ||
			strings.Contains(errMesg, "Authorization failed") ||
			strings.Contains(errMesg, "AuthorizationPermissionMismatch") ||
			strings.Contains(errMesg, "LinkedAccessCheckFailed"):
			return errMesg, nil

		case strings.Contains(errMesg, "LackOfPermissions"):
			log.Warnf("LackOfPermissions error occurred: %s", errMesg)
			return errMesg, nil

		case strings.Contains(errMesg, "InvalidTemplateDeployment"):
			// This indicates all Authorization errors are fixed
			// Sample error [{\"code\":\"PodIdentityAddonFeatureFlagNotEnabled\",\"message\":\"Provisioning of resource(s) for container service aks-24xalwx7i2ueg in resource group testdeployrg-Y2jsRAG failed. Message: PodIdentity addon is not allowed since feature 'Microsoft.ContainerService/EnablePodIdentityPreview' is not enabled.
			// Hence ok to proceed, and not return error in this condition
			log.Warnf("Post Authorization error occurred: %s", errMesg)
		}
	}
	return "", nil

}

// func (a *armDeploymentConfig) getARMDeployment(deploymentName string) error {

// 	bearerToken, err := a.azAPIClient.GetSPBearerToken(a.mpfCfg.TenantID, a.mpfCfg.SPClientID, a.mpfCfg.SPClientSecret)
// 	if err != nil {
// 		return err
// 	}

// 	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourcegroups/%s/providers/Microsoft.Resources/deployments/%s?api-version=2020-10-01", a.mpfCfg.SubscriptionID, a.mpfCfg.ResourceGroup.ResourceGroupName, deploymentName)

// 	client := &http.Client{}

// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		return err
// 	}

// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Accept", "application/json")
// 	req.Header.Set("User-Agent", "Go HTTP Client")

// 	// add bearer token to header
// 	req.Header.Add("Authorization", "Bearer "+bearerToken)

// 	// make request
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	// read response body
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return err
// 	}

// 	respBody := string(body)

// 	// fmt.Println(respBody)
// 	log.Debugln(respBody)
// 	// print response body
// 	if strings.Contains(respBody, "Authorization") {
// 		return errors.New(respBody)
// 	}

// 	return nil
// }

// Delete ARM deployment
func (a *armDeploymentConfig) cancelDeployment(ctx context.Context, deploymentName string, mpfConfig domain.MPFConfig) error {

	// Get deployments status. If status is "Running", cancel deployment, then delete deployment
	getResp, err := a.azAPIClient.DeploymentsClient.Get(ctx, mpfConfig.ResourceGroup.ResourceGroupName, deploymentName, nil)
	if err != nil {
		// Error indicates deployment does not exist, so cancelling deployment not needed
		if strings.Contains(err.Error(), "DeploymentNotFound") {
			log.Infof("Could not get deployment %s: ,Error :%s \n", deploymentName, err)
			return nil
		}
		// For any other error, log and return without trying to access potentially nil response
		log.Infof("Could not get deployment %s: ,Error :%s \n", deploymentName, err)
		return nil
	}

	// Additional safety checks for nil pointers
	if getResp.Properties == nil || getResp.Properties.ProvisioningState == nil {
		log.Infof("Deployment %s has incomplete response, skipping cleanup", deploymentName)
		return nil
	}

	log.Infof("Deployment status: %s\n", *getResp.Properties.ProvisioningState)

	if *getResp.Properties.ProvisioningState != armresources.ProvisioningStateRunning {
		return nil
	}

	const maxRetries = 24
	for retryCount := 0; retryCount < maxRetries; retryCount++ {
		_, err := a.azAPIClient.DeploymentsClient.Cancel(ctx, mpfConfig.ResourceGroup.ResourceGroupName, deploymentName, nil)
		if err == nil {
			log.Infof("Cancelled deployment %s", deploymentName)
			return nil
		}

		// cancel deployment
		log.Warnf("Could not cancel deployment %s: %s, retrying in a bit", deploymentName, err)
		time.Sleep(5 * time.Second)
	}

	log.Warnf("Could not cancel deployment %s after %d retries, giving up", deploymentName, maxRetries)
	return errors.New("could not cancel deployment")

}
