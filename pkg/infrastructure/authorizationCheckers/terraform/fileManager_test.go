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
	"bytes"
	"strings"
	"testing"

	"github.com/manisbindra/az-mpf/pkg/domain"
)

func TestSaveResultAsJSON(t *testing.T) {
	tests := []struct {
		name       string
		mpfResult  domain.MPFResult
		wantErr    bool
		wantOutput string
	}{
		{
			name: "valid MPFResult",
			mpfResult: domain.MPFResult{
				RequiredPermissions: map[string][]string{
					"": []string{
						"Microsoft.Authorization/roleAssignments/delete",
						"Microsoft.Authorization/roleAssignments/read",
						"Microsoft.Authorization/roleAssignments/write",
					},
				},
			},
			wantErr:    false,
			wantOutput: `{"RequiredPermissions":{"":["Microsoft.Authorization/roleAssignments/delete","Microsoft.Authorization/roleAssignments/read","Microsoft.Authorization/roleAssignments/write"]}}`,
		},
		{
			name: "valid MPFResult with detailed permissions",
			mpfResult: domain.MPFResult{
				RequiredPermissions: map[string][]string{
					"": []string{
						"Microsoft.ContainerService/managedClusters/read",
						"Microsoft.ContainerService/managedClusters/write",
						"Microsoft.Network/virtualNetworks/read",
						"Microsoft.Network/virtualNetworks/subnets/read",
						"Microsoft.Network/virtualNetworks/subnets/write",
						"Microsoft.Network/virtualNetworks/write",
						"Microsoft.Resources/deployments/read",
						"Microsoft.Resources/deployments/write",
					},
					"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X44/providers/Microsoft.ContainerService/managedClusters/azmpfakstestcluster": []string{
						"Microsoft.ContainerService/managedClusters/read",
						"Microsoft.ContainerService/managedClusters/write",
					},
					"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X44/providers/Microsoft.Network/virtualNetworks/azmpfakstestvnet": []string{
						"Microsoft.Network/virtualNetworks/read",
						"Microsoft.Network/virtualNetworks/write",
					},
					"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X44/providers/Microsoft.Network/virtualNetworks/azmpfakstestvnet/subnets/azmpfakstestsubnet": []string{
						"Microsoft.Network/virtualNetworks/subnets/read",
						"Microsoft.Network/virtualNetworks/subnets/write",
					},
				},
			},
			wantErr:    false,
			wantOutput: `{"RequiredPermissions":{"":["Microsoft.ContainerService/managedClusters/read","Microsoft.ContainerService/managedClusters/write","Microsoft.Network/virtualNetworks/read","Microsoft.Network/virtualNetworks/subnets/read","Microsoft.Network/virtualNetworks/subnets/write","Microsoft.Network/virtualNetworks/write","Microsoft.Resources/deployments/read","Microsoft.Resources/deployments/write"],"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X44/providers/Microsoft.ContainerService/managedClusters/azmpfakstestcluster":["Microsoft.ContainerService/managedClusters/read","Microsoft.ContainerService/managedClusters/write"],"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X44/providers/Microsoft.Network/virtualNetworks/azmpfakstestvnet":["Microsoft.Network/virtualNetworks/read","Microsoft.Network/virtualNetworks/write"],"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X44/providers/Microsoft.Network/virtualNetworks/azmpfakstestvnet/subnets/azmpfakstestsubnet":["Microsoft.Network/virtualNetworks/subnets/read","Microsoft.Network/virtualNetworks/subnets/write"]}}`,
		},
		{
			name: "empty MPFResult",
			mpfResult: domain.MPFResult{
				RequiredPermissions: map[string][]string{},
			},
			wantErr:    false,
			wantOutput: `{"RequiredPermissions":{}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := saveResultAsJSON(&buf, tt.mpfResult)
			if (err != nil) != tt.wantErr {
				t.Errorf("saveResultAsJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOutput := buf.String(); strings.TrimSpace(gotOutput) != strings.TrimSpace(tt.wantOutput) {
				t.Errorf("saveResultAsJSON() = %v, want %v", gotOutput, tt.wantOutput)
			}
		})
	}
}
