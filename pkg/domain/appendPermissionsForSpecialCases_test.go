package domain

import (
	"reflect"
	"testing"
)

func TestAppendPermissiosnForSpecialCases(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string][]string
		expected map[string][]string
	}{
		{
			name: "No additional permissions appended",
			input: map[string][]string{
				"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X44": {"Microsoft.ContainerService/managedClusters/read", "Microsoft.ContainerService/managedClusters/write"},
			},
			expected: map[string][]string{
				"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X44": {"Microsoft.ContainerService/managedClusters/read", "Microsoft.ContainerService/managedClusters/write"},
			},
		},
		{
			name: "No additional permissions appended for multiple scopes",
			input: map[string][]string{
				"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X47": {"Microsoft.ContainerService/managedClusters/read", "Microsoft.ContainerService/managedClusters/write"},
				"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X48": {"Microsoft.ContainerService/managedClusters/read", "Microsoft.ContainerService/managedClusters/write"},
			},
			expected: map[string][]string{
				"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X47": {"Microsoft.ContainerService/managedClusters/read", "Microsoft.ContainerService/managedClusters/write"},
				"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X48": {"Microsoft.ContainerService/managedClusters/read", "Microsoft.ContainerService/managedClusters/write"},
			},
		},
		{
			name: "Append for multiple scopes and permissions",
			input: map[string][]string{
				"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X47": {"Microsoft.ContainerService/managedClusters/read", "Microsoft.ContainerService/managedClusters/write", "Microsoft.Insights/components/read"},
				"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X48": {"Microsoft.ContainerService/managedClusters/read", "Microsoft.ContainerService/managedClusters/write", "Microsoft.Insights/components/write"},
			},
			expected: map[string][]string{
				"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X47": {
					"Microsoft.ContainerService/managedClusters/read",
					"Microsoft.ContainerService/managedClusters/write",
					"Microsoft.Insights/components/read",
					"Microsoft.Insights/components/currentbillingfeatures/read",
					"Microsoft.AlertsManagement/smartDetectorAlertRules/read",
				},
				"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X48": {
					"Microsoft.ContainerService/managedClusters/read",
					"Microsoft.ContainerService/managedClusters/write",
					"Microsoft.Insights/components/write",
					"Microsoft.Insights/components/currentbillingfeatures/write",
					"Microsoft.AlertsManagement/smartDetectorAlertRules/write",
				},
			},
		},
		{
			name: "Additional permissions appended",
			input: map[string][]string{
				"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X44": {"Microsoft.Insights/components/read"},
			},
			expected: map[string][]string{
				"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X44": {
					"Microsoft.Insights/components/read",
					"Microsoft.Insights/components/currentbillingfeatures/read",
					"Microsoft.AlertsManagement/smartDetectorAlertRules/read",
				},
			},
		},
		{
			name: "Additional permissions appended for write",
			input: map[string][]string{
				"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X44": {"Microsoft.Insights/components/write"},
			},
			expected: map[string][]string{
				"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X44": {
					"Microsoft.Insights/components/write",
					"Microsoft.Insights/components/currentbillingfeatures/write",
					"Microsoft.AlertsManagement/smartDetectorAlertRules/write",
				},
			},
		},
		{
			name: "Additional permissions appended for multiple scopes",
			input: map[string][]string{
				"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X47": {"Microsoft.Insights/components/read"},
				"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X48": {"Microsoft.Insights/components/write"},
			},
			expected: map[string][]string{
				"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X47": {
					"Microsoft.Insights/components/read",
					"Microsoft.Insights/components/currentbillingfeatures/read",
					"Microsoft.AlertsManagement/smartDetectorAlertRules/read",
				},
				"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X48": {
					"Microsoft.Insights/components/write",
					"Microsoft.Insights/components/currentbillingfeatures/write",
					"Microsoft.AlertsManagement/smartDetectorAlertRules/write",
				},
			},
		},
		{
			name:     "Empty input",
			input:    map[string][]string{},
			expected: map[string][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := appendPermissionsForSpecialCases(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
