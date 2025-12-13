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

package ARMTemplateWhatIf

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	// "log"
	"net/http"
	URL "net/url"
	"strings"

	"github.com/Azure/mpf/pkg/domain"
	"github.com/Azure/mpf/pkg/infrastructure/ARMTemplateShared"
	"github.com/Azure/mpf/pkg/infrastructure/azureAPI"
	"github.com/Azure/mpf/pkg/infrastructure/mpfSharedUtils"

	log "github.com/sirupsen/logrus"
)

type armWhatIfConfig struct {
	// mpfConfig   domain.MPFConfig
	armConfig   ARMTemplateShared.ArmTemplateAdditionalConfig
	azAPIClient *azureAPI.AzureAPIClients
}

func NewARMTemplateWhatIfAuthorizationChecker(subscriptionID string, armConfig ARMTemplateShared.ArmTemplateAdditionalConfig) *armWhatIfConfig {
	azAPIClient := azureAPI.NewAzureAPIClients(subscriptionID)
	return &armWhatIfConfig{
		azAPIClient: azAPIClient,
		armConfig:   armConfig,
	}

}

func (a *armWhatIfConfig) GetDeploymentAuthorizationErrors(mpfConfig domain.MPFConfig) (string, error) {
	return a.GetARMWhatIfAuthorizationErrors(a.armConfig.DeploymentName, mpfConfig)
}

func (a *armWhatIfConfig) CleanDeployment(mpfConfig domain.MPFConfig) error {
	log.Infoln("No additional cleanup needed in WhatIf mode")
	log.Infoln("*************************")

	return nil
}

// Get parameters in standard format that is without the schema, contentVersion and parameters fields

func (a *armWhatIfConfig) CreateEmptyDeployment(client *http.Client, deploymentName string, bearerToken string, mpfConfig domain.MPFConfig) error {
	var deploymentUri string
	if a.armConfig.SubscriptionScoped {
		deploymentUri = fmt.Sprintf("https://management.azure.com/subscriptions/%s/providers/Microsoft.Resources/deployments/%s?api-version=2021-04-01", mpfConfig.SubscriptionID, deploymentName)
	} else {
		deploymentUri = fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourcegroups/%s/providers/Microsoft.Resources/deployments/%s?api-version=2020-10-01", mpfConfig.SubscriptionID, mpfConfig.ResourceGroup.ResourceGroupName, deploymentName)
	}

	log.Info("Creating empty deployment...")
	log.Debug(deploymentUri)

	emptyTempl, err := mpfSharedUtils.ReadJson("../samples/templates/empty.json")
	if err != nil {
		return err
	}

	emptyTemplStdFmtMap := map[string]interface{}{
		"properties": map[string]interface{}{
			"mode":       "Incremental",
			"template":   emptyTempl,
			"parameters": map[string]interface{}{},
		},
	}

	// convert bodyJSON to string
	emptyTemplJSONBytes, err := json.Marshal(emptyTemplStdFmtMap)
	if err != nil {
		return err
	}

	emptyDeploymentJSONString := string(emptyTemplJSONBytes)

	deploymentReq, err := http.NewRequest("PUT", deploymentUri, bytes.NewBufferString(emptyDeploymentJSONString))
	if err != nil {
		return err
	}
	deploymentReq.Header.Set("Content-Type", "application/json")
	deploymentReq.Header.Set("Accept", "application/json")
	deploymentReq.Header.Set("User-Agent", "Go HTTP Client")

	// add bearer token to header
	deploymentReq.Header.Add("Authorization", "Bearer "+bearerToken)

	log.Debugf("%v", deploymentReq)

	// make deploymentReq
	deploymentResp, err := client.Do(deploymentReq)
	if err != nil {
		return err
	}
	log.Debugf("%v", deploymentResp)
	defer deploymentResp.Body.Close() //nolint:errcheck

	return nil
}

func (a *armWhatIfConfig) GetARMWhatIfAuthorizationErrors(deploymentName string, mpfConfig domain.MPFConfig) (string, error) {

	bearerToken, err := a.azAPIClient.GetSPBearerToken(mpfConfig.TenantID, mpfConfig.SP.SPClientID, mpfConfig.SP.SPClientSecret)
	if err != nil {
		// wrap error and return
		return "", fmt.Errorf("error getting bearer token: %w", err)
	}

	// read template and parameters
	template, err := mpfSharedUtils.ReadJson(a.armConfig.TemplateFilePath)
	if err != nil {
		// log.Errorf("Error reading template file: %v", err)
		return "", fmt.Errorf("%w: %w", ARMTemplateShared.ErrInvalidTemplate, fmt.Errorf("error reading template file: %w", err))
	}

	parameters, err := mpfSharedUtils.ReadJson(a.armConfig.ParametersFilePath)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ARMTemplateShared.ErrInvalidTemplate, fmt.Errorf("error reading parameters file: %w", err))
	}

	// convert parameters to standard format
	parameters = ARMTemplateShared.GetParametersInStandardFormat(parameters)

	// if Subscription scoped, we need to specify deployment location

	fullTemplate := map[string]interface{}{
		"properties": map[string]interface{}{
			"mode":       "Incremental",
			"template":   template,
			"parameters": parameters,
		},
	}

	if a.armConfig.SubscriptionScoped {
		fullTemplate["location"] = a.armConfig.Location
	}

	// convert bodyJSON to string
	fullTemplateJSONBytes, err := json.Marshal(fullTemplate)
	if err != nil {
		return "", fmt.Errorf("%w, %w", ARMTemplateShared.ErrInvalidTemplate, fmt.Errorf("error marshalling fullTemplateJSON: %w", err))
	}

	fullTemplateJSONString := string(fullTemplateJSONBytes)

	log.Debugln()
	log.Debugln(fullTemplateJSONString)
	log.Debugln()
	// create JSON body with template and parameters

	client := &http.Client{}

	var url string
	if a.armConfig.SubscriptionScoped {
		url = fmt.Sprintf("https://management.azure.com/subscriptions/%s/providers/Microsoft.Resources/deployments/%s/whatIf?api-version=2022-09-01", mpfConfig.SubscriptionID, deploymentName)
	} else {
		url = fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourcegroups/%s/providers/Microsoft.Resources/deployments/%s/whatIf?api-version=2021-04-01", mpfConfig.SubscriptionID, mpfConfig.ResourceGroup.ResourceGroupName, deploymentName)
	}

	reqMethod := "POST"

	req, err := http.NewRequest(reqMethod, url, bytes.NewBufferString(fullTemplateJSONString))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Go HTTP Client")

	// add bearer token to header
	req.Header.Add("Authorization", "Bearer "+bearerToken)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close() //nolint:errcheck

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	log.Debugf("Response Body: %s", string(bodyBytes))

	// if response status code is 400, indicates invalid template
	if resp.StatusCode == 400 {
		// log.Errorf("InvalidTemplate error occured: %s. Please check the Template and Parameters", resp.Status)
		// return "", errors.New("InvalidTemplate")
		// return "", fmt.Error("InvalidTemplate: Please check the Template and Parameters: %w", resp.Status)
		log.Warnf("Response Body: %s", string(bodyBytes))
		return "", fmt.Errorf("%w, %w", ARMTemplateShared.ErrInvalidTemplate, errors.New("please check the template and parameter files"))
	}

	whatIfRespLoc := resp.Header.Get("Location")
	log.Debugf("What if response Location: %s \n", whatIfRespLoc)

	_, err = URL.ParseRequestURI(whatIfRespLoc)
	if err != nil {
		// return "", err
		log.Warnf("Response Body: %s", string(bodyBytes))
		return "", fmt.Errorf("error parsing what if response location: %w", err)
	}

	respBody, err := a.GetWhatIfResp(whatIfRespLoc, bearerToken)
	if err != nil {
		log.Infof("Could not fetch what if response: %v \n", err)
		// return "", err
		return "", fmt.Errorf("could not fetch what if response: %w", err)
	}

	log.Debugln(respBody)

	switch {
	case strings.Contains(respBody, "LackOfPermissions"):
		log.Warnf("LackOfPermissions error occurred: %s", respBody)
		return respBody, nil
	case strings.Contains(respBody, "InvalidTemplate") && !strings.Contains(respBody, "InvalidTemplateDeployment"):
		// This indicates the ARM Template or Bicep File has issues.
		// Sample
		// {"status":"Failed","error":{"code":"InvalidTemplate","message":"Deployment template validation failed: 'The template parameters 'aksClusterName, virtualNetworkName' in the parameters file are not valid; they are not present in the original template and can therefore not be provided at deployment time. The only supported parameters for this template are 'clusterName, location, subnetName, vnetName'. Please see https://aka.ms/arm-pass-parameter-values for usage details.'.","additionalInfo":[{"type":"TemplateViolation","info":{"lineNumber":0,"linePosition":0,"path":""}}]}}
		log.Warnf("Response Body: %s", string(bodyBytes))
		return "", fmt.Errorf("%w: please check the template and parameters file: %s", ARMTemplateShared.ErrInvalidTemplate, respBody)
	case strings.Contains(respBody, "Authorization") && !strings.Contains(respBody, "{\"status\":\"Succeeded\""):
		// This indicates Authorization errors occured
		return respBody, nil
	case strings.Contains(respBody, "InvalidTemplateDeployment"):
		// This indicates all Authorization errors are fixed
		// Sample error [{\"code\":\"PodIdentityAddonFeatureFlagNotEnabled\",\"message\":\"Provisioning of resource(s) for container service aks-24xalwx7i2ueg in resource group testdeployrg-Y2jsRAG failed. Message: PodIdentity addon is not allowed since feature 'Microsoft.ContainerService/EnablePodIdentityPreview' is not enabled.
		// Hence ok to proceed, and not return error in this condition
		log.Warnf("Post Authorizaton error occured: %s", respBody)

	}

	return "", nil

}

func (a *armWhatIfConfig) GetWhatIfResp(whatIfRespLoc string, bearerToken string) (string, error) {

	client := &http.Client{}

	req, err := http.NewRequest("GET", whatIfRespLoc, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Go HTTP Client")

	// add bearer token to header
	req.Header.Add("Authorization", "Bearer "+bearerToken)

	var respBody string
	maxRetries := 50
	retryCount := 0
	for {
		// make request
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close() //nolint:errcheck

		// read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		respBody = string(body)

		// If response body is not empty, break out of loop
		if respBody != "" {
			log.Infoln("Whatif Results Response Received..")
			break
		}

		retryCount++
		if retryCount == maxRetries {
			log.Warnf("Whatif Results Response Body is empty after %d retries, returning empty response body", maxRetries)
			return "", fmt.Errorf("whatif Results Response Body is empty after %d retries", maxRetries)
		}

		log.Infoln("Whatif Results Response Body is empty, retrying in a bit...")
		// Sleep for 500 milli seconds and try again
		time.Sleep(500 * time.Millisecond)

	}

	log.Debugln("Whatif Results Response Body:")
	log.Debugln(respBody)

	return respBody, nil

}
