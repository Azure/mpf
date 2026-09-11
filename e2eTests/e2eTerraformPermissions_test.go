// MIT License
//
// Copyright (c) Microsoft Corporation.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package e2etests

import (
	"slices"
	"strings"
	"testing"
)

const terraformProviderReadPermission = "Microsoft.Resources/subscriptions/providers/read"

func getTerraformE2EBootstrapPermissions(additionalPermissions ...string) ([]string, []string) {
	// AzureRM 3.x/4.x can list subscription providers while initializing its
	// provider cache. Seed the permission so residual RBAC state cannot hide it.
	initialPermissions := append([]string{
		"Microsoft.Resources/deployments/read",
		"Microsoft.Resources/deployments/write",
		terraformProviderReadPermission,
	}, additionalPermissions...)

	return initialPermissions, slices.Clone(initialPermissions)
}

func getCaseInsensitivePermissionCount(permissions []string) int {
	uniquePermissions := make(map[string]struct{}, len(permissions))
	for _, permission := range permissions {
		uniquePermissions[strings.ToLower(permission)] = struct{}{}
	}

	return len(uniquePermissions)
}

func getCaseInsensitivePermissionSetDiff(expected, actual []string) ([]string, []string) {
	expectedPermissions := make(map[string]string, len(expected))
	actualPermissions := make(map[string]string, len(actual))

	for _, permission := range expected {
		expectedPermissions[strings.ToLower(permission)] = permission
	}
	for _, permission := range actual {
		actualPermissions[strings.ToLower(permission)] = permission
	}

	var missingPermissions []string
	for normalizedPermission, permission := range expectedPermissions {
		if _, found := actualPermissions[normalizedPermission]; !found {
			missingPermissions = append(missingPermissions, permission)
		}
	}

	var unexpectedPermissions []string
	for normalizedPermission, permission := range actualPermissions {
		if _, found := expectedPermissions[normalizedPermission]; !found {
			unexpectedPermissions = append(unexpectedPermissions, permission)
		}
	}

	slices.Sort(missingPermissions)
	slices.Sort(unexpectedPermissions)
	return missingPermissions, unexpectedPermissions
}

func TestE2ETerraformBootstrapPermissions(t *testing.T) {
	expected := []string{
		"Microsoft.Resources/deployments/read",
		"Microsoft.Resources/deployments/write",
		terraformProviderReadPermission,
		"Microsoft.Test/widgets/read",
	}

	initialPermissions, resultPermissions := getTerraformE2EBootstrapPermissions("Microsoft.Test/widgets/read")
	if !slices.Equal(initialPermissions, expected) {
		t.Fatalf("unexpected initial permissions: %v", initialPermissions)
	}
	if !slices.Equal(resultPermissions, expected) {
		t.Fatalf("unexpected result permissions: %v", resultPermissions)
	}

	initialPermissions[0] = "modified"
	if resultPermissions[0] != expected[0] {
		t.Fatal("initial and result permissions share a backing array")
	}

	nextInitialPermissions, _ := getTerraformE2EBootstrapPermissions("Microsoft.Test/widgets/read")
	if !slices.Equal(nextInitialPermissions, expected) {
		t.Fatalf("helper result was mutated across calls: %v", nextInitialPermissions)
	}

	t.Run("counts casing variants once", func(t *testing.T) {
		permissions := []string{
			"Microsoft.Resources/subscriptions/resourceGroups/read",
			"Microsoft.Resources/subscriptions/resourcegroups/read",
			"Microsoft.Resources/subscriptions/resourceGroups/write",
		}

		if count := getCaseInsensitivePermissionCount(permissions); count != 2 {
			t.Fatalf("expected 2 case-insensitive permissions, got %d", count)
		}
	})

	t.Run("compares permission sets case insensitively", func(t *testing.T) {
		expectedPermissions := []string{
			"Microsoft.Resources/subscriptions/resourceGroups/read",
			"Microsoft.Resources/subscriptions/resourceGroups/write",
		}
		actualPermissions := []string{
			"Microsoft.Resources/subscriptions/resourcegroups/read",
			"Microsoft.Resources/subscriptions/RESOURCEGROUPS/READ",
			"Microsoft.Resources/subscriptions/resourcegroups/delete",
		}

		missingPermissions, unexpectedPermissions := getCaseInsensitivePermissionSetDiff(expectedPermissions, actualPermissions)
		if !slices.Equal(missingPermissions, []string{"Microsoft.Resources/subscriptions/resourceGroups/write"}) {
			t.Fatalf("unexpected missing permissions: %v", missingPermissions)
		}
		if !slices.Equal(unexpectedPermissions, []string{"Microsoft.Resources/subscriptions/resourcegroups/delete"}) {
			t.Fatalf("unexpected additional permissions: %v", unexpectedPermissions)
		}
	})
}
