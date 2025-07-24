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

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompletePartialActionResourceGroup(t *testing.T) {
	action := "/read"
	scope := "/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/e2eTest-zSUDZd0"

	completedAction, err := completePartialAction(action, scope)

	assert.Nil(t, err)
	assert.Equal(t, "Microsoft.Resources/subscriptions/resourceGroups/read", completedAction)
}

func TestCompletePartialActionSubscription(t *testing.T) {
	action := "/read"
	scope := "/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS"

	completedAction, err := completePartialAction(action, scope)

	assert.Nil(t, err)
	assert.Equal(t, "Microsoft.Resources/subscriptions/read", completedAction)
}

func TestCompletePartialActionUnrecognizedScope(t *testing.T) {
	action := "/read"
	scope := "/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/test/providers/Microsoft.Storage/storageAccounts/mystorageaccount"

	completedAction, err := completePartialAction(action, scope)

	assert.NotNil(t, err)
	// assert that error occurs and not specific error message
	assert.Contains(t, err.Error(), "unable to complete partial action '/read' for unrecognized scope pattern")
	assert.Equal(t, "/read", completedAction) // Should return original action even with error
}

func TestCompletePartialActionCompleteAction(t *testing.T) {
	action := "Microsoft.Storage/storageAccounts/read"
	scope := "/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/test/providers/Microsoft.Storage/storageAccounts/mystorageaccount"

	completedAction, err := completePartialAction(action, scope)

	assert.Nil(t, err)
	assert.Equal(t, "Microsoft.Storage/storageAccounts/read", completedAction) // Should return original action unchanged
}

func TestCompletePartialActionDifferentPartialAction(t *testing.T) {
	action := "/write"
	scope := "/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/e2eTest-zSUDZd0"

	completedAction, err := completePartialAction(action, scope)

	assert.Nil(t, err)
	assert.Equal(t, "/write", completedAction) // Should return original action unchanged since only /read is handled
}

func TestCompletePartialActionResourceGroupWithCasing(t *testing.T) {
	action := "/read"
	scope := "/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourcegroups/e2eTest-zSUDZd0" // lowercase 'resourcegroups'

	completedAction, err := completePartialAction(action, scope)

	assert.Nil(t, err)
	assert.Equal(t, "Microsoft.Resources/subscriptions/resourceGroups/read", completedAction) // Should now match due to case insensitive regex
}

func TestCompletePartialActionResourceGroupUpperCase(t *testing.T) {
	action := "/read"
	scope := "/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/RESOURCEGROUPS/e2eTest-zSUDZd0" // uppercase 'RESOURCEGROUPS'

	completedAction, err := completePartialAction(action, scope)

	assert.Nil(t, err)
	assert.Equal(t, "Microsoft.Resources/subscriptions/resourceGroups/read", completedAction) // Should match due to case insensitive regex
}

func TestCompletePartialActionResourceGroupVariousCases(t *testing.T) {
	action := "/read"
	scope := "/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/ResourceGroups/e2eTest-zSUDZd0" // mixed case 'ResourceGroups'

	completedAction, err := completePartialAction(action, scope)

	assert.Nil(t, err)
	assert.Equal(t, "Microsoft.Resources/subscriptions/resourceGroups/read", completedAction) // Should match due to case insensitive regex
}

func TestCompletePartialActionResourceGroupMixedCase(t *testing.T) {
	// Test the actual case from the test data
	action := "/read"
	scope := "/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/e2eTest-zSUDZd0"

	completedAction, err := completePartialAction(action, scope)

	assert.Nil(t, err)
	assert.Equal(t, "Microsoft.Resources/subscriptions/resourceGroups/read", completedAction)
}

func TestCompletePartialActionEmptyAction(t *testing.T) {
	action := ""
	scope := "/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/e2eTest-zSUDZd0"

	completedAction, err := completePartialAction(action, scope)

	assert.NotNil(t, err)
	assert.Equal(t, "action cannot be empty", err.Error())
	assert.Equal(t, "", completedAction) // Should return empty string when action is empty
}

func TestCompletePartialActionEmptyScope(t *testing.T) {
	action := "/read"
	scope := ""

	completedAction, err := completePartialAction(action, scope)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "unable to complete partial action '/read' for unrecognized scope pattern")
	assert.Equal(t, "/read", completedAction) // Should return original action with error
}

func TestCompletePartialActionInvalidScopeFormat(t *testing.T) {
	action := "/read"
	scope := "invalid-scope-format"

	completedAction, err := completePartialAction(action, scope)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "unable to complete partial action '/read' for unrecognized scope pattern")
	assert.Equal(t, "/read", completedAction) // Should return original action with error
}
