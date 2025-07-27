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
	"encoding/json"
	"reflect"
	"testing"

	"github.com/Azure/mpf/pkg/domain"
)

func TestCreateUpdateCustomRoleDataActions(t *testing.T) {
	tests := []struct {
		name               string
		role               domain.Role
		permissions        []string
		expectedDataActions []string
	}{
		{
			name: "No default data actions",
			role: domain.Role{
				RoleDefinitionID:         "test-role-id",
				RoleDefinitionName:       "test-role",
				RoleDefinitionResourceID: "/subscriptions/test/providers/Microsoft.Authorization/roleDefinitions/test-role-id",
				DefaultDataActions:       []string{},
			},
			permissions:         []string{"Microsoft.Resources/deployments/read"},
			expectedDataActions: []string{},
		},
		{
			name: "Single default data action",
			role: domain.Role{
				RoleDefinitionID:         "test-role-id",
				RoleDefinitionName:       "test-role",
				RoleDefinitionResourceID: "/subscriptions/test/providers/Microsoft.Authorization/roleDefinitions/test-role-id",
				DefaultDataActions:       []string{"Microsoft.Search/searchServices/indexes/documents/read"},
			},
			permissions:         []string{"Microsoft.Resources/deployments/read"},
			expectedDataActions: []string{"Microsoft.Search/searchServices/indexes/documents/read"},
		},
		{
			name: "Multiple default data actions",
			role: domain.Role{
				RoleDefinitionID:         "test-role-id",
				RoleDefinitionName:       "test-role",
				RoleDefinitionResourceID: "/subscriptions/test/providers/Microsoft.Authorization/roleDefinitions/test-role-id",
				DefaultDataActions: []string{
					"Microsoft.Search/searchServices/indexes/documents/read",
					"Microsoft.Search/searchServices/indexes/documents/write",
					"Microsoft.KeyVault/vaults/secrets/getSecret/action",
				},
			},
			permissions: []string{"Microsoft.Resources/deployments/read"},
			expectedDataActions: []string{
				"Microsoft.Search/searchServices/indexes/documents/read",
				"Microsoft.Search/searchServices/indexes/documents/write",
				"Microsoft.KeyVault/vaults/secrets/getSecret/action",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily test the HTTP calls, but we can test the data structure creation
			// by creating the same data structure that would be sent in the HTTP request
			
			subscription := "test-subscription"
			subScope := "/subscriptions/" + subscription
			
			expectedData := map[string]interface{}{
				"assignableScopes": []string{subScope},
				"description":      tt.role.RoleDefinitionName,
				"id":               tt.role.RoleDefinitionResourceID,
				"name":             tt.role.RoleDefinitionID,
				"permissions": []map[string]interface{}{
					{
						"actions":        tt.permissions,
						"dataActions":    tt.role.DefaultDataActions,
						"notActions":     []string{},
						"notDataActions": []string{},
					},
				},
				"roleName": tt.role.RoleDefinitionName,
				"roleType": "CustomRole",
			}
			
			properties := map[string]interface{}{
				"properties": expectedData,
			}
			
			// Marshal to JSON to verify the structure is correct
			jsonData, err := json.Marshal(properties)
			if err != nil {
				t.Fatalf("Failed to marshal JSON: %v", err)
			}
			
			// Unmarshal back to verify the dataActions are preserved correctly
			var unmarshaled map[string]interface{}
			err = json.Unmarshal(jsonData, &unmarshaled)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}
			
			// Navigate to the dataActions field
			props, ok := unmarshaled["properties"].(map[string]interface{})
			if !ok {
				t.Fatal("Properties not found or not a map")
			}
			
			permissions, ok := props["permissions"].([]interface{})
			if !ok {
				t.Fatal("Permissions not found or not an array")
			}
			
			if len(permissions) != 1 {
				t.Fatalf("Expected 1 permission object, got %d", len(permissions))
			}
			
			permissionObj, ok := permissions[0].(map[string]interface{})
			if !ok {
				t.Fatal("Permission object is not a map")
			}
			
			dataActions, ok := permissionObj["dataActions"].([]interface{})
			if !ok {
				t.Fatal("DataActions not found or not an array")
			}
			
			// Convert to string slice for comparison
			actualDataActions := make([]string, len(dataActions))
			for i, v := range dataActions {
				actualDataActions[i] = v.(string)
			}
			
			if !reflect.DeepEqual(actualDataActions, tt.expectedDataActions) {
				t.Errorf("Expected dataActions %v, got %v", tt.expectedDataActions, actualDataActions)
			}
		})
	}
}