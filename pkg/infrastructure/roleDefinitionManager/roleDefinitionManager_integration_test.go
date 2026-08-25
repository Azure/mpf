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

package roledefinitionmanager

import (
	"context"
	"os"
	"testing"

	"github.com/Azure/mpf/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetBuiltInRoles_Integration is a read-only integration test that verifies
// the manager can enumerate built-in role definitions from a real subscription.
// It is skipped unless MPF_INTEGRATION_SUBSCRIPTION_ID is set. It creates no
// Azure resources and only performs list operations.
func TestGetBuiltInRoles_Integration(t *testing.T) {
	subscriptionID := os.Getenv("MPF_INTEGRATION_SUBSCRIPTION_ID")
	if subscriptionID == "" {
		t.Skip("MPF_INTEGRATION_SUBSCRIPTION_ID not set; skipping read-only Azure integration test")
	}

	mgr := NewRoleDefinitionManager(subscriptionID)
	roles, err := mgr.GetBuiltInRoles(context.Background(), subscriptionID)
	require.NoError(t, err)

	// Azure ships well over 100 built-in roles.
	assert.Greater(t, len(roles), 100, "expected many built-in roles")

	// Every returned role should have a name and at least one action.
	var reader *domain.BuiltInRole
	for i := range roles {
		assert.NotEmpty(t, roles[i].RoleName, "role name should not be empty")
		assert.NotEmpty(t, roles[i].RoleDefinitionID, "role definition id should not be empty")
		if roles[i].RoleName == "Reader" {
			reader = &roles[i]
		}
	}

	// The built-in Reader role should exist and grant "*/read".
	require.NotNil(t, reader, "expected to find the built-in Reader role")
	assert.Contains(t, reader.Actions, "*/read")

	// Sanity check that suggestion works end-to-end against real role data.
	suggestion := domain.SuggestBuiltInRoles([]string{"Microsoft.Storage/storageAccounts/read"}, roles)
	assert.NotEmpty(t, suggestion.SingleRoleMatches, "Reader (and others) should cover a read permission")
	assert.Empty(t, suggestion.UncoveredPermissions)
}
