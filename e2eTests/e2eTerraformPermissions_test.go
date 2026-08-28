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
}
