package domain

import (
	"reflect"
	"testing"
)

func TestGetOperationStatusesReadPermission(t *testing.T) {
	tests := []struct {
		name       string
		permission string
		expected   string
		expectedOk bool
	}{
		{
			name:       "resource type write permission",
			permission: "Microsoft.ContainerRegistry/registries/write",
			expected:   "Microsoft.ContainerRegistry/registries/operationStatuses/read",
			expectedOk: true,
		},
		{
			name:       "nested resource type write permission",
			permission: "Microsoft.OperationalInsights/workspaces/dataSources/write",
			expected:   "Microsoft.OperationalInsights/workspaces/dataSources/operationStatuses/read",
			expectedOk: true,
		},
		{
			name:       "surrounding whitespace is trimmed",
			permission: "  Microsoft.ContainerRegistry/registries/write  ",
			expected:   "Microsoft.ContainerRegistry/registries/operationStatuses/read",
			expectedOk: true,
		},
		{
			name:       "read permission is ignored",
			permission: "Microsoft.ContainerRegistry/registries/read",
			expected:   "",
			expectedOk: false,
		},
		{
			name:       "delete permission is ignored",
			permission: "Microsoft.ContainerRegistry/registries/delete",
			expected:   "",
			expectedOk: false,
		},
		{
			name:       "action permission is ignored",
			permission: "Microsoft.OperationalInsights/workspaces/sharedKeys/action",
			expected:   "",
			expectedOk: false,
		},
		{
			name:       "wildcard permission is ignored",
			permission: "Microsoft.Storage/*/write",
			expected:   "",
			expectedOk: false,
		},
		{
			name:       "namespace only permission is ignored",
			permission: "write",
			expected:   "",
			expectedOk: false,
		},
		{
			name:       "operationStatuses write permission is ignored",
			permission: "Microsoft.ContainerRegistry/registries/operationStatuses/write",
			expected:   "",
			expectedOk: false,
		},
		{
			name:       "empty permission is ignored",
			permission: "",
			expected:   "",
			expectedOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := GetOperationStatusesReadPermission(tt.permission)
			if ok != tt.expectedOk {
				t.Fatalf("expected ok=%v, got ok=%v", tt.expectedOk, ok)
			}
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestAppendOperationStatusesReadPermissions(t *testing.T) {
	const scopeA = "/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X44"
	const scopeB = "/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X45"

	tests := []struct {
		name     string
		input    map[string][]string
		expected map[string][]string
	}{
		{
			name: "no write permissions leaves map untouched",
			input: map[string][]string{
				scopeA: {"Microsoft.ContainerRegistry/registries/read"},
			},
			expected: map[string][]string{
				scopeA: {"Microsoft.ContainerRegistry/registries/read"},
			},
		},
		{
			name: "appends polling permission for write permission",
			input: map[string][]string{
				scopeA: {"Microsoft.ContainerRegistry/registries/read", "Microsoft.ContainerRegistry/registries/write"},
			},
			expected: map[string][]string{
				scopeA: {
					"Microsoft.ContainerRegistry/registries/read",
					"Microsoft.ContainerRegistry/registries/write",
					"Microsoft.ContainerRegistry/registries/operationStatuses/read",
				},
			},
		},
		{
			name: "appends polling permissions across multiple scopes",
			input: map[string][]string{
				scopeA: {"Microsoft.ContainerRegistry/registries/write"},
				scopeB: {"Microsoft.Resources/subscriptions/resourcegroups/write"},
			},
			expected: map[string][]string{
				scopeA: {
					"Microsoft.ContainerRegistry/registries/write",
					"Microsoft.ContainerRegistry/registries/operationStatuses/read",
				},
				scopeB: {
					"Microsoft.Resources/subscriptions/resourcegroups/write",
					"Microsoft.Resources/subscriptions/resourcegroups/operationStatuses/read",
				},
			},
		},
		{
			name:     "empty map",
			input:    map[string][]string{},
			expected: map[string][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AppendOperationStatusesReadPermissions(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}
