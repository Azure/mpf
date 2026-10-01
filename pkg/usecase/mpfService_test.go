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
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Azure/mpf/pkg/domain"
)

type deploymentResponse struct {
	message string
	err     error
}

type scriptedDeploymentChecker struct {
	responses []deploymentResponse
	calls     int
}

func (c *scriptedDeploymentChecker) GetDeploymentAuthorizationErrors(domain.MPFConfig) (string, error) {
	c.calls++
	if c.calls > len(c.responses) {
		return "", nil
	}
	response := c.responses[c.calls-1]
	return response.message, response.err
}

func (*scriptedDeploymentChecker) CleanDeployment(domain.MPFConfig) error {
	return nil
}

type noOpRoleManager struct{}

func (noOpRoleManager) DetachRolesFromSP(context.Context, string, string, domain.Role) error {
	return nil
}

func (noOpRoleManager) AssignRoleToSP(string, string, domain.Role) error {
	return nil
}

func (noOpRoleManager) CreateUpdateCustomRole(string, domain.Role, []string) (error, []string) { //nolint:staticcheck
	return nil, nil
}

func (noOpRoleManager) DeleteCustomRole(string, domain.Role) error {
	return nil
}

func newTestMPFService(checker *scriptedDeploymentChecker) *MPFService {
	service := NewMPFService(
		context.Background(),
		nil,
		noOpRoleManager{},
		checker,
		domain.MPFConfig{SubscriptionID: "subscription"},
		nil,
		nil,
		false,
		false,
		false,
	)
	service.sleep = func(time.Duration) {}
	return service
}

func TestMPFServiceStopsAfterConsecutiveRetrySentinels(t *testing.T) {
	responses := make([]deploymentResponse, maxConsecutiveRetryResponses)
	for i := range responses {
		responses[i].message = RetryDeploymentResponseErrorMessage
	}
	checker := &scriptedDeploymentChecker{responses: responses}
	service := newTestMPFService(checker)

	result, err := service.GetMinimumPermissionsRequired()
	if err == nil {
		t.Fatal("GetMinimumPermissionsRequired() error = nil, want retry cap error")
	}
	if !strings.Contains(err.Error(), "requested retry") {
		t.Errorf("error = %q, want descriptive retry-cap error", err)
	}
	if checker.calls != maxConsecutiveRetryResponses {
		t.Errorf("checker calls = %d, want %d", checker.calls, maxConsecutiveRetryResponses)
	}
	if result.IterationCount != 0 {
		t.Errorf("permission iteration count = %d, want 0", result.IterationCount)
	}
}

func TestMPFServiceResetsRetryCapAfterAuthorizationProgress(t *testing.T) {
	responses := make([]deploymentResponse, 0, 2*maxConsecutiveRetryResponses)
	for range maxConsecutiveRetryResponses - 1 {
		responses = append(responses, deploymentResponse{message: RetryDeploymentResponseErrorMessage})
	}
	responses = append(responses, deploymentResponse{message: `{"error":{"code":"AuthorizationFailed","message":"The client 'client' with object id 'object' does not have authorization to perform action 'Microsoft.Resources/deployments/write' over scope '/subscriptions/sub/resourcegroups/rg/providers/Microsoft.Resources/deployments/deployment' or the scope is invalid."}}`})
	for range maxConsecutiveRetryResponses - 1 {
		responses = append(responses, deploymentResponse{message: RetryDeploymentResponseErrorMessage})
	}
	checker := &scriptedDeploymentChecker{responses: responses}
	service := newTestMPFService(checker)

	result, err := service.GetMinimumPermissionsRequired()
	if err != nil {
		t.Fatalf("GetMinimumPermissionsRequired() error = %v, want nil", err)
	}
	if checker.calls != len(responses)+1 {
		t.Errorf("checker calls = %d, want %d including successful completion", checker.calls, len(responses)+1)
	}
	if result.IterationCount != 1 {
		t.Errorf("permission iteration count = %d, want 1", result.IterationCount)
	}
}
