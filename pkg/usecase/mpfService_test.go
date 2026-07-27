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

package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newInvalidActionTestService() *MPFService {
	return &MPFService{
		requiredPermissions:  make(map[string][]string),
		autoAddedPermissions: make(map[string]bool),
		rejectedAutoAdded:    make(map[string]bool),
	}
}

func TestRecordInvalidActionsOnlyDropsAutoAddedPermissions(t *testing.T) {
	s := newInvalidActionTestService()
	s.requiredPermissions["sub"] = []string{
		"Microsoft.ContainerRegistry/registries/operationStatuses/read",
		"Microsoft.Insights/components/currentbillingfeatures/delete",
	}
	s.autoAddedPermissions["microsoft.containerregistry/registries/operationstatuses/read"] = true

	s.recordInvalidActions([]string{
		"Microsoft.ContainerRegistry/registries/operationStatuses/read",
		"Microsoft.Insights/components/currentbillingfeatures/delete",
	})

	assert.Equal(t, []string{"Microsoft.ContainerRegistry/registries/operationStatuses/read"}, s.rejectedAutoAddedList)
	// the auto added action is dropped, the deployment reported one is left untouched
	assert.Equal(t, []string{"Microsoft.Insights/components/currentbillingfeatures/delete"}, s.requiredPermissions["sub"])
}

func TestRecordInvalidActionsKeepsDeploymentReportedActions(t *testing.T) {
	s := newInvalidActionTestService()
	s.requiredPermissions["sub"] = []string{"Microsoft.Insights/components/currentbillingfeatures/delete"}

	s.recordInvalidActions([]string{"Microsoft.Insights/components/currentbillingfeatures/delete"})

	assert.Empty(t, s.rejectedAutoAddedList)
	assert.Equal(t, []string{"Microsoft.Insights/components/currentbillingfeatures/delete"}, s.requiredPermissions["sub"])
}

func TestRecordInvalidActionsDoesNotRecordTheSameRejectionTwice(t *testing.T) {
	s := newInvalidActionTestService()
	s.requiredPermissions["sub"] = []string{"Microsoft.Storage/storageAccounts/operationStatuses/read"}
	s.autoAddedPermissions["microsoft.storage/storageaccounts/operationstatuses/read"] = true

	s.recordInvalidActions([]string{"Microsoft.Storage/storageAccounts/operationStatuses/read"})
	s.recordInvalidActions([]string{"Microsoft.Storage/storageAccounts/operationStatuses/read"})

	assert.Equal(t, []string{"Microsoft.Storage/storageAccounts/operationStatuses/read"}, s.rejectedAutoAddedList)
	assert.Empty(t, s.requiredPermissions["sub"])
}

func TestRecordInvalidActionsIgnoresEmptyInput(t *testing.T) {
	s := newInvalidActionTestService()

	s.recordInvalidActions(nil)

	assert.Empty(t, s.rejectedAutoAddedList)
}

func TestReturnMPFResultExcludesRejectedAutoAddedPermissions(t *testing.T) {
	s := newInvalidActionTestService()
	s.requiredPermissions["sub"] = []string{
		"Microsoft.ContainerRegistry/registries/write",
		"Microsoft.ContainerRegistry/registries/operationStatuses/read",
		"Microsoft.Resources/subscriptions/resourceGroups/operationStatuses/read",
	}
	s.autoAddedPermissions["microsoft.resources/subscriptions/resourcegroups/operationstatuses/read"] = true
	s.recordInvalidActions([]string{"Microsoft.Resources/subscriptions/resourceGroups/operationStatuses/read"})

	result, err := s.returnMPFResult(nil)

	assert.NoError(t, err)
	assert.Contains(t, result.RequiredPermissions["sub"], "Microsoft.ContainerRegistry/registries/write")
	assert.Contains(t, result.RequiredPermissions["sub"], "Microsoft.ContainerRegistry/registries/operationStatuses/read")
	assert.NotContains(t, result.RequiredPermissions["sub"], "Microsoft.Resources/subscriptions/resourceGroups/operationStatuses/read")
}

func TestWithAutoAddOperationStatusesReadForWrite(t *testing.T) {
	s := &MPFService{}
	WithAutoAddOperationStatusesReadForWrite(true)(s)
	assert.True(t, s.autoAddOperationStatusesReadForWrite)

	WithAutoAddOperationStatusesReadForWrite(false)(s)
	assert.False(t, s.autoAddOperationStatusesReadForWrite)
}
