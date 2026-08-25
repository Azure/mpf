package domain

import (
	"reflect"
	"sort"
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
		name             string
		input            map[string][]string
		expected         map[string][]string
		expectedAppended []string
	}{
		{
			name: "no write permissions leaves map untouched",
			input: map[string][]string{
				scopeA: {"Microsoft.ContainerRegistry/registries/read"},
			},
			expected: map[string][]string{
				scopeA: {"Microsoft.ContainerRegistry/registries/read"},
			},
			expectedAppended: nil,
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
			expectedAppended: []string{"Microsoft.ContainerRegistry/registries/operationStatuses/read"},
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
			expectedAppended: []string{
				"Microsoft.ContainerRegistry/registries/operationStatuses/read",
				"Microsoft.Resources/subscriptions/resourcegroups/operationStatuses/read",
			},
		},
		{
			name: "polling permission already reported by the deployment is not appended",
			input: map[string][]string{
				scopeA: {
					"Microsoft.ContainerRegistry/registries/write",
					"Microsoft.ContainerRegistry/registries/operationStatuses/read",
				},
			},
			expected: map[string][]string{
				scopeA: {
					"Microsoft.ContainerRegistry/registries/write",
					"Microsoft.ContainerRegistry/registries/operationStatuses/read",
				},
			},
			expectedAppended: nil,
		},
		{
			name: "mixed case write action is mapped",
			input: map[string][]string{
				scopeA: {"Microsoft.ContainerRegistry/registries/Write"},
			},
			expected: map[string][]string{
				scopeA: {
					"Microsoft.ContainerRegistry/registries/Write",
					"Microsoft.ContainerRegistry/registries/operationStatuses/read",
				},
			},
			expectedAppended: []string{"Microsoft.ContainerRegistry/registries/operationStatuses/read"},
		},
		{
			name:             "empty map",
			input:            map[string][]string{},
			expected:         map[string][]string{},
			expectedAppended: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, appended := AppendOperationStatusesReadPermissions(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
			// Scope iteration order is not deterministic, so compare appended as a set.
			sort.Strings(appended)
			expectedAppended := append([]string(nil), tt.expectedAppended...)
			sort.Strings(expectedAppended)
			if len(appended) != len(expectedAppended) {
				t.Fatalf("expected appended %v, got %v", expectedAppended, appended)
			}
			for i := range appended {
				if appended[i] != expectedAppended[i] {
					t.Fatalf("expected appended %v, got %v", expectedAppended, appended)
				}
			}
		})
	}
}

func TestAppendOperationStatusesReadPermissionsDoesNotMutateInput(t *testing.T) {
	const scope = "/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS"
	input := map[string][]string{
		scope: {"Microsoft.ContainerRegistry/registries/write"},
	}

	_, _ = AppendOperationStatusesReadPermissions(input)

	if len(input[scope]) != 1 {
		t.Fatalf("input map was modified: %v", input[scope])
	}
}
