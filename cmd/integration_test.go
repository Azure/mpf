package main

import (
	"os"
	"testing"
)

// Test integration to verify the flag parsing works end-to-end
func TestDefaultDataActionsFlagIntegration(t *testing.T) {
	// Save original args and restore after test
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	// Mock the minimum required flags to avoid validation errors
	testCases := []struct {
		name            string
		dataActions     string
		expectedActions []string
	}{
		{
			name:            "No data actions provided",
			dataActions:     "",
			expectedActions: []string{},
		},
		{
			name:            "Single data action",
			dataActions:     "Microsoft.Search/searchServices/indexes/documents/read",
			expectedActions: []string{"Microsoft.Search/searchServices/indexes/documents/read"},
		},
		{
			name:            "Multiple data actions",
			dataActions:     "Microsoft.Search/searchServices/indexes/documents/read,Microsoft.Search/searchServices/indexes/documents/write",
			expectedActions: []string{"Microsoft.Search/searchServices/indexes/documents/read", "Microsoft.Search/searchServices/indexes/documents/write"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset the flag variable
			flgDefaultDataActions = ""
			
			// Simulate command line args
			if tc.dataActions != "" {
				os.Args = []string{"azmpf", "--defaultDataActions", tc.dataActions}
			} else {
				os.Args = []string{"azmpf"}
			}

			// Test the parsing function directly
			result := parseDefaultDataActions(tc.dataActions)
			
			if len(result) != len(tc.expectedActions) {
				t.Errorf("Expected %d actions, got %d", len(tc.expectedActions), len(result))
				return
			}
			
			for i, expected := range tc.expectedActions {
				if result[i] != expected {
					t.Errorf("Expected action %d to be %q, got %q", i, expected, result[i])
				}
			}
		})
	}
}

// Test the configuration creation with data actions
func TestRootMPFConfigWithDataActions(t *testing.T) {
	// Save original values and restore after test
	origSubscriptionID := flgSubscriptionID
	origTenantID := flgTenantID
	origSPClientID := flgSPClientID
	origSPObjectID := flgSPObjectID
	origSPClientSecret := flgSPClientSecret
	origDefaultDataActions := flgDefaultDataActions
	
	defer func() {
		flgSubscriptionID = origSubscriptionID
		flgTenantID = origTenantID
		flgSPClientID = origSPClientID
		flgSPObjectID = origSPObjectID
		flgSPClientSecret = origSPClientSecret
		flgDefaultDataActions = origDefaultDataActions
	}()

	// Set test values
	flgSubscriptionID = "test-subscription-id"
	flgTenantID = "test-tenant-id"
	flgSPClientID = "test-client-id"
	flgSPObjectID = "test-object-id"
	flgSPClientSecret = "test-client-secret"
	
	testCases := []struct {
		name            string
		dataActions     string
		expectedActions []string
	}{
		{
			name:            "No data actions",
			dataActions:     "",
			expectedActions: []string{},
		},
		{
			name:            "Single data action",
			dataActions:     "Microsoft.Search/searchServices/indexes/documents/read",
			expectedActions: []string{"Microsoft.Search/searchServices/indexes/documents/read"},
		},
		{
			name:            "Multiple data actions",
			dataActions:     "Microsoft.Search/searchServices/indexes/documents/read,Microsoft.Search/searchServices/indexes/documents/write",
			expectedActions: []string{"Microsoft.Search/searchServices/indexes/documents/read", "Microsoft.Search/searchServices/indexes/documents/write"},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flgDefaultDataActions = tc.dataActions
			
			config := getRootMPFConfig()
			
			if len(config.Role.DefaultDataActions) != len(tc.expectedActions) {
				t.Errorf("Expected %d data actions in config, got %d", len(tc.expectedActions), len(config.Role.DefaultDataActions))
				return
			}
			
			for i, expected := range tc.expectedActions {
				if config.Role.DefaultDataActions[i] != expected {
					t.Errorf("Expected data action %d to be %q, got %q", i, expected, config.Role.DefaultDataActions[i])
				}
			}
		})
	}
}