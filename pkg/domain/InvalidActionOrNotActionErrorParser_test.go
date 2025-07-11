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

func TestGetDeleteActionFromInvalidActionOrNotActionError(t *testing.T) {
	tests := []struct {
		name                string
		input               string
		expectedActions     []string
		expectedError       bool
		expectedErrorString string
	}{
		{
			name:            "Valid delete action",
			input:           `{"error":{"code":"InvalidActionOrNotAction","message":"'Microsoft.Insights/components/currentbillingfeatures/delete' does not match any of the actions supported by the providers."}}`,
			expectedActions: []string{"Microsoft.Insights/components/currentbillingfeatures/delete"},
			expectedError:   false,
		},
		{
			name:                "Valid action (non-delete)",
			input:               `{"error":{"code":"InvalidActionOrNotAction","message":"'Microsoft.OperationalInsights/workspaces/PrivateEndpointConnectionsApproval/action' does not match any of the actions supported by the providers."}}`,
			expectedActions:     []string{},
			expectedError:       true,
			expectedErrorString: "No matches found in 'invalidActionErrorMessage' error message",
		},
		{
			name:                "Valid read action",
			input:               `{"error":{"code":"InvalidActionOrNotAction","message":"'Microsoft.Resources/subscriptions/resourceGroups/read' does not match any of the actions supported by the providers."}}`,
			expectedActions:     []string{},
			expectedError:       true,
			expectedErrorString: "No matches found in 'invalidActionErrorMessage' error message",
		},
		{
			name:                "Non InvalidActionOrNotAction error",
			input:               `{"error":{"code":"SomeOtherError","message":"'Microsoft.Insights/components/currentbillingfeatures/delete' does not match any of the actions supported by the providers."}}`,
			expectedActions:     []string{},
			expectedError:       true,
			expectedErrorString: "Could not parse deploment error, potentially due to a Non-InvalidActionOrNotAction error",
		},
		{
			name:                "No matches found",
			input:               `{"error":{"code":"InvalidActionOrNotAction","message":"No delete action here"}}`,
			expectedActions:     []string{},
			expectedError:       true,
			expectedErrorString: "No matches found in 'invalidActionErrorMessage' error message",
		},
		{
			name:                "Empty input",
			input:               "",
			expectedActions:     []string{},
			expectedError:       true,
			expectedErrorString: "No matches found in 'invalidActionErrorMessage' error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actions, err := GetDeleteActionFromInvalidActionOrNotActionError(tt.input)
			assert.Equal(t, tt.expectedActions, actions)
			if tt.expectedError {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErrorString)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetInvalidActionFromInvalidActionOrNotActionError(t *testing.T) {
	tests := []struct {
		name                string
		input               string
		expectedActions     []string
		expectedError       bool
		expectedErrorString string
	}{
		{
			name:            "Valid delete action",
			input:           `{"error":{"code":"InvalidActionOrNotAction","message":"'Microsoft.Insights/components/currentbillingfeatures/delete' does not match any of the actions supported by the providers."}}`,
			expectedActions: []string{"Microsoft.Insights/components/currentbillingfeatures/delete"},
			expectedError:   false,
		},
		{
			name:            "Valid action (non-delete)",
			input:           `{"error":{"code":"InvalidActionOrNotAction","message":"'Microsoft.OperationalInsights/workspaces/PrivateEndpointConnectionsApproval/action' does not match any of the actions supported by the providers."}}`,
			expectedActions: []string{"Microsoft.OperationalInsights/workspaces/PrivateEndpointConnectionsApproval/action"},
			expectedError:   false,
		},
		{
			name:            "Valid read action",
			input:           `{"error":{"code":"InvalidActionOrNotAction","message":"'Microsoft.Resources/subscriptions/resourceGroups/read' does not match any of the actions supported by the providers."}}`,
			expectedActions: []string{"Microsoft.Resources/subscriptions/resourceGroups/read"},
			expectedError:   false,
		},
		{
			name:                "Non InvalidActionOrNotAction error",
			input:               `{"error":{"code":"SomeOtherError","message":"'Microsoft.Insights/components/currentbillingfeatures/delete' does not match any of the actions supported by the providers."}}`,
			expectedActions:     []string{},
			expectedError:       true,
			expectedErrorString: "could not parse deploment error, potentially due to a Non-InvalidActionOrNotAction error",
		},
		{
			name:                "No matches found",
			input:               `{"error":{"code":"InvalidActionOrNotAction","message":"No delete action here"}}`,
			expectedActions:     []string{},
			expectedError:       true,
			expectedErrorString: "no matches found in 'invalidActionErrorMessage' error message",
		},
		{
			name:                "Empty input",
			input:               "",
			expectedActions:     []string{},
			expectedError:       true,
			expectedErrorString: "no matches found in 'invalidActionErrorMessage' error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actions, err := GetInvalidActionFromInvalidActionOrNotActionError(tt.input)
			assert.Equal(t, tt.expectedActions, actions)
			if tt.expectedError {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErrorString)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
