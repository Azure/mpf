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

package sproleassignmentmanager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v3"
	"github.com/Azure/mpf/pkg/domain"
	"github.com/Azure/mpf/pkg/infrastructure/azureAPI"
	"github.com/google/uuid"

	log "github.com/sirupsen/logrus"
)

type SPRoleAssignmentManager struct {
	azAPIClient *azureAPI.AzureAPIClients
}

func NewSPRoleAssignmentManager(subscriptionID string) *SPRoleAssignmentManager {
	azAPIClient := azureAPI.NewAzureAPIClients(subscriptionID)
	return &SPRoleAssignmentManager{
		azAPIClient: azAPIClient,
	}
}

// CreateUpdateCustomRole creates or updates a custom role in Azure
// It retries up to 5 times if it encounters an InvalidActionOrNotAction error
// It returns an error if it fails to create or update the role
// It returns a list of invalid actions that were removed from the role
func (r *SPRoleAssignmentManager) CreateUpdateCustomRole(subscription string, role domain.Role, permissions []string) (error, []string) {
	retryCount := 5
	permissionsToAdd := permissions
	var invalidActions []string

	for i := 0; i < retryCount; i++ {
		log.Debugf("Creating/Updating Role Definition: %s, Retry: %d", role.RoleDefinitionName, i+1)
		err := r.createUpdateCustomRole(subscription, role, permissionsToAdd)
		if err != nil && strings.Contains(err.Error(), "InvalidActionOrNotAction") {
			errMsg := err.Error()
			log.Warnf("InvalidActionOrNotAction error occurred. Attempting to remove invalid action...")
			actionsToRemove, err := domain.GetInvalidActionFromInvalidActionOrNotActionError(errMsg)
			if err != nil {
				log.Warnf("Could not get actions to remove from error: %s", err.Error())
				return err, []string{}
			}
			log.Debug("Filtering Invalid Actions: ", actionsToRemove)
			invalidActions = append(invalidActions, actionsToRemove...)
			permissionsToAdd = filterInvalidActions(permissionsToAdd, actionsToRemove)
			continue // retry
		}
		if err != nil { // not retrying for other errors
			log.Debugf("Error when updating role: %s", err.Error())
			return err, []string{}
		}
		log.Infof("Role definition created/updated successfully")
		break
	}
	return nil, invalidActions
}

func (r *SPRoleAssignmentManager) createUpdateCustomRole(subscription string, role domain.Role, permissions []string) error {

	// rgScope := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscription, resourceGroupName)
	subScope := fmt.Sprintf("/subscriptions/%s", subscription)

	data := map[string]interface{}{
		"assignableScopes": []string{
			// rgScope,
			subScope,
		},
		"description": role.RoleDefinitionName,
		"id":          role.RoleDefinitionResourceID,
		"name":        role.RoleDefinitionID,
		"permissions": []map[string]interface{}{
			{
				"actions":        permissions,
				"dataActions":    []string{},
				"notActions":     []string{},
				"notDataActions": []string{},
			},
		},
		"roleName": role.RoleDefinitionName,
		"roleType": "CustomRole",
		// "type":     "Microsoft.Authorization/roleDefinitions",
	}

	properties := map[string]interface{}{
		"properties": data,
	}
	// marshal data as json
	jsonData, err := json.Marshal(properties)
	if err != nil {
		return err
	}

	//convert to json string
	jsonString := string(jsonData)

	// log.Printf("jsonString: %s", jsonString)
	log.Debugf("jsonString: %s", jsonString)

	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/providers/Microsoft.Authorization/roleDefinitions/%s?api-version=2018-01-01-preview", subscription, role.RoleDefinitionID)

	client := &http.Client{}

	req, err := http.NewRequest("PUT", url, bytes.NewBufferString(jsonString))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Go HTTP Client")

	defaultApiBearerToken, err := r.azAPIClient.GetDefaultAPIBearerToken()
	if err != nil {
		return err
	}

	// add bearer token to header
	req.Header.Add("Authorization", "Bearer "+defaultApiBearerToken)

	// make request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Debugln(string(body))
	if strings.Contains(string(body), "InvalidActionOrNotAction") {
		return fmt.Errorf("InvalidActionOrNotAction: %s", string(body))
	}

	return nil
}

func (r *SPRoleAssignmentManager) AssignRoleToSP(subscription string, SPOBjectID string, role domain.Role) error {

	// scope := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscription, resourceGroupName)
	scope := fmt.Sprintf("/subscriptions/%s", subscription)
	url := fmt.Sprintf("https://management.azure.com/%s/providers/Microsoft.Authorization/roleAssignments/%s?api-version=2022-04-01", scope, uuid.New().String())

	data := map[string]interface{}{
		"principalId":      SPOBjectID,
		"principalType":    "ServicePrincipal",
		"roleDefinitionId": role.RoleDefinitionResourceID,
	}

	properties := map[string]interface{}{
		"properties": data,
	}

	// marshal data as json
	jsonData, err := json.Marshal(properties)
	if err != nil {
		return err
	}

	//convert to json string
	jsonString := string(jsonData)

	log.Debugf("jsonString: %s", jsonString)

	client := &http.Client{}

	req, err := http.NewRequest("PUT", url, bytes.NewBufferString(jsonString))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Go HTTP Client")

	defaultApiBearerToken, err := r.azAPIClient.GetDefaultAPIBearerToken()
	if err != nil {
		return err
	}

	// add bearer token to header
	req.Header.Add("Authorization", "Bearer "+defaultApiBearerToken)

	// make request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("failed to assign role to SP. Status code: %s", string(body))
	}

	// print response body
	log.Debugln(string(body))
	return nil
}

// func (r *SPRoleAssignmentManager) AssignRoleToSP(scope string) error {

// 	// armauthorization.NewClassicAdministratorsClient()
// 	clientFactory, err := armauthorization.NewClientFactory(m.SubscriptionID, m.DefaultCred, nil)
// 	if err != nil {
// 		log.Fatalf("failed to create client: %v", err)
// 	}

// 	res, err := clientFactory.NewRoleAssignmentsClient().Create(m.Ctx, m.SubscriptionID, uuid.New().String(), armauthorization.RoleAssignmentCreateParameters{
// 		Properties: &armauthorization.RoleAssignmentProperties{
// 			PrincipalID:      &m.SPObjectID,
// 			PrincipalType:    to.StringPtr(armauthorization.PrincipalTypeServicePrincipal),
// 			RoleDefinitionID: &m.RoleDefinitionResourceID,
// 		},
// 	}, nil)

// 	// rac, err := armauthorization.NewRoleAssignmentsClient(m.SubscriptionID, m.DefaultCred, nil)
// 	// if err != nil {
// 	// 	return err
// 	// }

// 	// rao := armauthorization.RoleAssignmentsClientCreateOptions{

// 	// }
// 	// roleAssignmentParams := authorization.RoleAssignmentCreateParameters{
// 	// 	Properties: &authorization.RoleAssignmentProperties{
// 	// 		PrincipalID:      &m.SPObjectID,
// 	// 		RoleDefinitionID: &m.RoleDefinitionResourceID,
// 	// 	},
// 	// }

// 	// roleAssignmentParams := armauthorization.RoleAssignmentCreateParameters{
// 	// 	Properties: &armauthorization.RoleAssignmentProperties{
// 	// 		PrincipalID:      &m.SPObjectID,
// 	// 		RoleDefinitionID: &m.RoleDefinitionResourceID,
// 	// 	},
// 	// }

// 	_, err = rac.Create(m.Ctx, scope, uuid.New().String(), roleAssignmentParams, nil)
// 	// _, err = m.RoleAssignmentsClient.Create(m.Ctx, scope, uuid.New().String(), roleAssignmentParams)

// 	if err != nil {
// 		if strings.Contains(err.Error(), "RoleAssignmentExists") {
// 			log.Infoln("Role assignment already exists. Skipping...")
// 			return nil
// 		}
// 		return err
// 	}

// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }z

// DetachRolesFromSP detaches the specified role from the SP
func (r *SPRoleAssignmentManager) DetachRolesFromSP(ctx context.Context, subscription string, SPOBjectID string, role domain.Role) error {
	pager := r.azAPIClient.RoleAssignmentsClient.NewListForSubscriptionPager(&armauthorization.RoleAssignmentsClientListForSubscriptionOptions{
		Filter: to.Ptr(fmt.Sprintf("assignedTo('%s')", SPOBjectID)),
	})

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return err
		}

		for _, roleAssignment := range page.Value {
			if roleAssignment.Properties != nil && roleAssignment.Properties.RoleDefinitionID != nil && strings.EqualFold(*roleAssignment.Properties.RoleDefinitionID, role.RoleDefinitionResourceID) {
				_, err := r.azAPIClient.RoleAssignmentsDeletionClient.DeleteByID(ctx, string(*roleAssignment.ID), nil)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (r *SPRoleAssignmentManager) DeleteCustomRole(subscription string, role domain.Role) error {
	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/providers/Microsoft.Authorization/roleDefinitions/%s?api-version=2018-01-01-preview", subscription, role.RoleDefinitionID)

	client := &http.Client{}

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Go HTTP Client")

	defaultApiBearerToken, err := r.azAPIClient.GetDefaultAPIBearerToken()
	if err != nil {
		return err
	}

	// add bearer token to header
	req.Header.Add("Authorization", "Bearer "+defaultApiBearerToken)

	// make request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("Could not delete role definition: %s\n", err)
	}

	log.Debugln(string(body))
	log.Infoln("Role definition deleted successfully")

	return nil
}

func stringExistsInSlice(s string, sl []string) bool {
	for _, v := range sl {
		if v == s {
			return true
		}
	}
	return false
}

func filterInvalidActions(permissions []string, invalidActions []string) []string {
	var validPermissions []string
	for _, permission := range permissions {
		if !stringExistsInSlice(permission, invalidActions) {
			validPermissions = append(validPermissions, permission)
		}
	}
	return validPermissions
}
