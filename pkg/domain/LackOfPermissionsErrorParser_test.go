package domain

import (
	"testing"
)

func TestParseLackOfPermissionsError(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  map[string][]string
		expectErr bool
	}{
		{
			name:  "Valid error message",
			input: "GET https://management.azure.com/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourcegroups/e2eTest-FZhv7Yf/providers/Microsoft.Resources/deployments/e2eTest-MQJBRka/operationStatuses/08584561178049431871\n--------------------------------------------------------------------------------\nRESPONSE 200: 200 OK\nERROR CODE: DeploymentFailed\n--------------------------------------------------------------------------------\n{\n  \"status\": \"Failed\",\n  \"error\": {\n    \"code\": \"DeploymentFailed\",\n    \"message\": \"At least one resource deployment operation failed. Please list deployment operations for details. Please see https://aka.ms/arm-deployment-operations for usage details.\",\n    \"details\": [\n      {\n        \"code\": \"Conflict\",\n        \"message\": \"{\\r\\n  \\\"status\\\": \\\"Failed\\\",\\r\\n  \\\"error\\\": {\\r\\n    \\\"code\\\": \\\"ResourceDeploymentFailure\\\",\\r\\n    \\\"message\\\": \\\"The resource write operation failed to complete successfully, because it reached terminal provisioning state 'Failed'.\\\",\\r\\n    \\\"details\\\": [\\r\\n      {\\r\\n        \\\"code\\\": \\\"DeploymentFailed\\\",\\r\\n        \\\"target\\\": \\\"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/e2eTest-FZhv7Yf/providers/Microsoft.Resources/deployments/ai-gen-unique-9-lvul-deployment\\\",\\r\\n        \\\"message\\\": \\\"At least one resource deployment operation failed. Please list deployment operations for details. Please see https://aka.ms/arm-deployment-operations for usage details.\\\",\\r\\n        \\\"details\\\": [\\r\\n          {\\r\\n            \\\"code\\\": \\\"ValidationError\\\",\\r\\n            \\\"target\\\": \\\"workspace.Kind\\\",\\r\\n            \\\"message\\\": \\\"You do not have Azure RBAC permissions to create new AI hubs. To create an AI hub, you are required to have contributor permissions on your resource group, or more specifically the Microsoft.MachineLearningServices/workspaces/hubs/write permission.\\\",\\r\\n            \\\"details\\\": [\\r\\n              {\\r\\n                \\\"code\\\": \\\"LackOfPermissions\\\",\\r\\n                \\\"target\\\": \\\"workspace.Kind\\\",\\r\\n                \\\"message\\\": \\\"You do not have Azure RBAC permissions to create new AI hubs. To create an AI hub, you are required to have contributor permissions on your resource group, or more specifically the Microsoft.MachineLearningServices/workspaces/hubs/write permission.\\\",\\r\\n                \\\"details\\\": []\\r\\n              }\\r\\n            ]\\r\\n          }\\r\\n        ]\\r\\n      }\\r\\n    ]\\r\\n  }\\r\\n}\"\n      }\n    ]\n  }\n}\n--------------------------------------------------------------------------------\n",
			expected: map[string][]string{
				"ScopeCannotBeParsedFromLackOfPermissionsError": {"Microsoft.MachineLearningServices/workspaces/hubs/write", "Microsoft.MachineLearningServices/workspaces/hubs/write"}},
			expectErr: false,
		},
		{
			name:      "No matches",
			input:     "This is an unrelated error message.",
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseLackOfPermissionsError(tt.input)
			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, got: %v", tt.expectErr, err)
			}
			if !tt.expectErr && !mapsEqual(result, tt.expected) {
				t.Errorf("Expected: %v, got: %v", tt.expected, result)
			}
		})
	}
}

func mapsEqual(a, b map[string][]string) bool {
	if len(a) != len(b) {
		return false
	}
	for key, valA := range a {
		valB, ok := b[key]
		if !ok || len(valA) != len(valB) {
			return false
		}
		for i := range valA {
			if valA[i] != valB[i] {
				return false
			}
		}
	}
	return true
}
