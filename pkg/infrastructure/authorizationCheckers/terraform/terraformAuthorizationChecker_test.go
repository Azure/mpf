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

package terraform

import (
	"errors"
	"testing"
	"time"
)

func TestIsTransientStorageNotFoundError(t *testing.T) {
	tests := []struct {
		name string
		err  string
		want bool
	}{
		{
			name: "storage account read",
			err:  "retrieving Storage Account: unexpected status 404 (404 ResourceNotFound)",
			want: true,
		},
		{
			name: "share properties read",
			err:  "retrieving share properties for Storage Account: 404 ResourceNotFound",
			want: true,
		},
		{
			name: "file properties read",
			err:  "retrieving file properties for Storage Account: 404 ResourceNotFound",
			want: true,
		},
		{
			name: "queue properties read",
			err:  "retrieving queue properties for Storage Account: 404 ResourceNotFound",
			want: true,
		},
		{
			name: "blob properties read",
			err:  "retrieving blob properties for Storage Account: 404 ResourceNotFound",
			want: true,
		},
		{
			name: "static website availability read",
			err:  "waiting for the static website to become available: 404 ResourceNotFound",
			want: true,
		},
		{
			name: "missing resource group",
			err:  "reading resource group: 404 ResourceNotFound",
			want: false,
		},
		{
			name: "unrelated resource",
			err:  "retrieving Key Vault properties: 404 ResourceNotFound",
			want: false,
		},
		{
			name: "bare resource not found",
			err:  "ResourceNotFound",
			want: false,
		},
		{
			name: "storage authorization error",
			err:  "retrieving share properties for Storage Account: 403 AuthorizationPermissionMismatch",
			want: false,
		},
		{
			name: "storage 404 without Azure error code",
			err:  "retrieving share properties for Storage Account: 404 The specified resource does not exist",
			want: false,
		},
		{
			name: "invalid Terraform configuration",
			err:  "Error: Invalid resource configuration",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTransientStorageNotFoundError(tt.err); got != tt.want {
				t.Fatalf("isTransientStorageNotFoundError() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestRetryOrReturnErrorIsBoundedAndReturnsOriginalError(t *testing.T) {
	var delays []time.Duration
	checker := &terraformDeploymentConfig{
		sleep: func(delay time.Duration) {
			delays = append(delays, delay)
		},
	}
	originalErr := errors.New("retrieving share properties for Storage Account: 404 ResourceNotFound")

	for i := range terraformRetryDelays {
		message, err := checker.retryOrReturnError(originalErr)
		if message != RetryDeploymentResponseErrorMessage || err != nil {
			t.Fatalf("retry %d = (%q, %v), want retry sentinel and nil error", i+1, message, err)
		}
	}

	if len(delays) != len(terraformRetryDelays) {
		t.Fatalf("got %d delays, want %d", len(delays), len(terraformRetryDelays))
	}
	for i, delay := range terraformRetryDelays {
		if delays[i] != delay {
			t.Errorf("delay %d = %s, want %s", i+1, delays[i], delay)
		}
	}

	message, err := checker.retryOrReturnError(originalErr)
	if message != originalErr.Error() {
		t.Errorf("exhausted retry message = %q, want original error %q", message, originalErr.Error())
	}
	if err != originalErr {
		t.Errorf("exhausted retry error = %v, want original error instance %v", err, originalErr)
	}
}

func TestRetryCountResetsAfterProgress(t *testing.T) {
	var delays []time.Duration
	checker := &terraformDeploymentConfig{
		sleep: func(delay time.Duration) {
			delays = append(delays, delay)
		},
	}
	originalErr := errors.New("temporary Storage error")

	for range terraformRetryDelays {
		if _, err := checker.retryOrReturnError(originalErr); err != nil {
			t.Fatalf("retryOrReturnError() error = %v", err)
		}
	}
	checker.resetRetryCountAfterProgress()

	if message, err := checker.retryOrReturnError(originalErr); message != RetryDeploymentResponseErrorMessage || err != nil {
		t.Fatalf("retry after progress = (%q, %v), want retry sentinel and nil error", message, err)
	}
	if got, want := delays[len(delays)-1], terraformRetryDelays[0]; got != want {
		t.Errorf("first delay after progress = %s, want %s", got, want)
	}
}

func TestRetryCountResetsAfterParseableAuthorizationError(t *testing.T) {
	checker := &terraformDeploymentConfig{consecutiveRetryCount: len(terraformRetryDelays)}
	authError := `{"error":{"code":"AuthorizationFailed","message":"The client 'client' with object id 'object' does not have authorization to perform action 'Microsoft.Resources/deployments/write' over scope '/subscriptions/sub/resourcegroups/rg/providers/Microsoft.Resources/deployments/deployment' or the scope is invalid."}}`

	checker.resetRetryCountForAuthorizationError(authError)
	if checker.consecutiveRetryCount != 0 {
		t.Fatalf("retry count after parseable authorization error = %d, want 0", checker.consecutiveRetryCount)
	}

	checker.consecutiveRetryCount = 2
	checker.resetRetryCountForAuthorizationError("unparseable Authorization error")
	if checker.consecutiveRetryCount != 2 {
		t.Fatalf("retry count after unparseable authorization error = %d, want 2", checker.consecutiveRetryCount)
	}
}

func TestParseableAuthorizationErrorTakesPriorityOverStorageNotFound(t *testing.T) {
	errorMsg := `retrieving share properties for Storage Account: 404 ResourceNotFound
{"error":{"code":"AuthorizationFailed","message":"The client 'client' with object id 'object' does not have authorization to perform action 'Microsoft.Resources/deployments/write' over scope '/subscriptions/sub/resourcegroups/rg/providers/Microsoft.Resources/deployments/deployment' or the scope is invalid."}}`
	applyErr := errors.New(errorMsg)
	slept := false
	checker := &terraformDeploymentConfig{
		consecutiveRetryCount: 2,
		sleep: func(time.Duration) {
			slept = true
		},
	}

	message, err := checker.handleTerraformApplyError(applyErr)
	if err != nil {
		t.Fatalf("handleTerraformApplyError() error = %v, want nil authorization result", err)
	}
	if message != errorMsg {
		t.Fatalf("handleTerraformApplyError() message = %q, want aggregated authorization error", message)
	}
	if checker.consecutiveRetryCount != 0 {
		t.Fatalf("retry count = %d, want reset after authorization progress", checker.consecutiveRetryCount)
	}
	if slept {
		t.Fatal("parseable authorization error must not trigger retry backoff")
	}
}

func TestTerraformApplyErrorRetryClassification(t *testing.T) {
	tests := []struct {
		name     string
		errorMsg string
	}{
		{
			name:     "plain transient Storage not found",
			errorMsg: "retrieving share properties for Storage Account: 404 ResourceNotFound",
		},
		{
			name:     "unparseable dataplane readiness authorization error",
			errorMsg: "waiting for the Data Plane for Storage Account: AuthorizationPermissionMismatch",
		},
		{
			name: "transient Storage not found with unparseable authorization diagnostic",
			errorMsg: `retrieving share properties for Storage Account: 404 ResourceNotFound
AuthorizationPermissionMismatch: incomplete provider diagnostic`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var delays []time.Duration
			checker := &terraformDeploymentConfig{
				sleep: func(delay time.Duration) {
					delays = append(delays, delay)
				},
			}

			message, err := checker.handleTerraformApplyError(errors.New(tt.errorMsg))
			if err != nil {
				t.Fatalf("handleTerraformApplyError() error = %v, want nil retry response", err)
			}
			if message != RetryDeploymentResponseErrorMessage {
				t.Fatalf("handleTerraformApplyError() message = %q, want retry sentinel", message)
			}
			if checker.consecutiveRetryCount != 1 {
				t.Fatalf("retry count = %d, want 1", checker.consecutiveRetryCount)
			}
			if len(delays) != 1 || delays[0] != terraformRetryDelays[0] {
				t.Fatalf("delays = %v, want [%s]", delays, terraformRetryDelays[0])
			}
		})
	}
}
