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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	// "log"
	"net/http"
	"strings"

	"github.com/Azure/mpf/pkg/infrastructure/ARMTemplateShared"
	"github.com/Azure/mpf/pkg/infrastructure/azureAPI"
	"github.com/Azure/mpf/pkg/infrastructure/mpfSharedUtils"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/mpf/pkg/domain"
	log "github.com/sirupsen/logrus"
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
	return a.deployARMTemplate(a.armConfig.DeploymentName, mpfConfig)
}

func (a *armDeploymentConfig) CleanDeployment(mpfConfig domain.MPFConfig) error {
	log.Infoln("Cleaning up resources...")
	log.Infoln("*************************")

	// Cancel deployment. Even if cancelling deployment fails attempt to delete other resources
	_ = a.cancelDeployment(a.ctx, a.armConfig.DeploymentName, mpfConfig)

	return nil
}

func (a *armDeploymentConfig) deployARMTemplate(deploymentName string, mpfConfig domain.MPFConfig) (string, error) {

	// jsonData, err := json.Marshal(properties)
	// spCred, err := azidentity.NewClientSecretCredential(a.mpfCfg.Args.TenantID, a.mpfCfg.Args.SPClientID, a.mpfCfg.Args.SPClientSecret, nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	bearerToken, err := a.azAPIClient.GetSPBearerToken(mpfConfig.TenantID, mpfConfig.SP.SPClientID, mpfConfig.SP.SPClientSecret)
	if err != nil {
		return "", err
	}

	// read template and parameters
	template, err := mpfSharedUtils.ReadJson(a.armConfig.TemplateFilePath)
	if err != nil {
		return "", err
	}

	parameters, err := mpfSharedUtils.ReadJson(a.armConfig.ParametersFilePath)
	if err != nil {
		return "", err
	}

	// convert parameters to standard format
	parameters = ARMTemplateShared.GetParametersInStandardFormat(parameters)

	fullTemplate := map[string]interface{}{
		"properties": map[string]interface{}{
			"mode":       "Incremental",
			"template":   template,
			"parameters": parameters,
		},
	}

	// convert bodyJSON to string
	fullTemplateJSONBytes, err := json.Marshal(fullTemplate)
	if err != nil {
		return "", err
	}

	fullTemplateJSONString := string(fullTemplateJSONBytes)

	log.Debugln()
	log.Debugln(fullTemplateJSONString)
	log.Debugln()
	// create JSON body with template and parameters

	client := &http.Client{}

	log.Info("MPF mode is fullDeployment, Proceeding to create resources....")
	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourcegroups/%s/providers/Microsoft.Resources/deployments/%s?api-version=2020-10-01", mpfConfig.SubscriptionID, mpfConfig.ResourceGroup.ResourceGroupName, deploymentName)
	reqMethod := "PUT"

	req, err := http.NewRequest(reqMethod, url, bytes.NewBufferString(fullTemplateJSONString))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Go HTTP Client")

	// add bearer token to header
	req.Header.Add("Authorization", "Bearer "+bearerToken)

	// make request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var respBody string

	// read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	respBody = string(body)

	// fmt.Println(respBody)
	log.Debugln(respBody)
	// print response body
	if strings.Contains(respBody, "Authorization") {
		return respBody, nil
	}

	if strings.Contains(respBody, "InvalidTemplateDeployment") {
		// This indicates all Authorization errors are fixed
		// Sample error [{\"code\":\"PodIdentityAddonFeatureFlagNotEnabled\",\"message\":\"Provisioning of resource(s) for container service aks-24xalwx7i2ueg in resource group testdeployrg-Y2jsRAG failed. Message: PodIdentity addon is not allowed since feature 'Microsoft.ContainerService/EnablePodIdentityPreview' is not enabled.
		// Hence ok to proceed, and not return error in this condition
		log.Warnf("Non Authorizaton error occured: %s", respBody)
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
	}

	log.Infof("Deployment status: %s\n", *getResp.DeploymentExtended.Properties.ProvisioningState)

	if *getResp.DeploymentExtended.Properties.ProvisioningState == armresources.ProvisioningStateRunning {

		retryCount := 0
		for _, err := a.azAPIClient.DeploymentsClient.Cancel(ctx, mpfConfig.ResourceGroup.ResourceGroupName, deploymentName, nil); err != nil; {
			// cancel deployment
			if err != nil {
				// return err
				log.Warnf("Could not cancel deployment %s: %s, retrying in a bit", deploymentName, err)
				time.Sleep(5 * time.Second)
				retryCount++
				if retryCount >= 24 {
					log.Warnf("Could not cancel deployment %s: %s, giving up", deploymentName, err)
					return errors.New("could not cancel deployment")
				}

			}

		}
		log.Infof("Cancelled deployment %s", deploymentName)
		return nil
	}

	return nil
}
