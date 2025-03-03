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
	"reflect"
	"testing"
)

func TestParseAuthorizationPermissionMismatchError(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string][]string
		wantErr bool
	}{
		{
			name: "Valid queue error message",
			input: `retrieving queue properties for Storage Account (Subscription: "SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS"
Resource Group Name: "rg-nygst"
Storage Account Name: "sanygst"): executing request: unexpected status 403 (403 This request is not authorized to perform this operation using this permission.) with AuthorizationPermissionMismatch: This request is not authorized to perform this operation using this permission.`,
			want: map[string][]string{
				"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/rg-nygst/providers/Microsoft.Storage/storageAccounts/sanygst": {"Microsoft.Storage/storageAccounts/queueServices/read"},
			},
			wantErr: false,
		},
		{
			name: "Valid file error message",
			input: `retrieving file properties for Storage Account (Subscription: "SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS"
Resource Group Name: "rg-nygst"
Storage Account Name: "sanygst"): executing request: unexpected status 403 (403 This request is not authorized to perform this operation using this permission.) with AuthorizationPermissionMismatch: This request is not authorized to perform this operation using this permission.`,
			want: map[string][]string{
				"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/rg-nygst/providers/Microsoft.Storage/storageAccounts/sanygst": {"Microsoft.Storage/storageAccounts/fileServices/read"},
			},
			wantErr: false,
		},
		{
			name: "Valid blob error message",
			input: `retrieving blob properties for Storage Account (Subscription: "SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS"
Resource Group Name: "rg-nygst"
Storage Account Name: "sanygst"): executing request: unexpected status 403 (403 This request is not authorized to perform this operation using this permission.) with AuthorizationPermissionMismatch: This request is not authorized to perform this operation using this permission.`,
			want: map[string][]string{
				"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/rg-nygst/providers/Microsoft.Storage/storageAccounts/sanygst": {"Microsoft.Storage/storageAccounts/blobServices/read"},
			},
			wantErr: false,
		},
		{
			name:    "Invalid error message",
			input:   `some invalid error message`,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Empty error message",
			input:   ``,
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid Format Error Message",
			input: `retrieving blob properties for Storage Account (Subscription: "SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS"
Resource Group Name: "rg-nygst" WRONG WRONG WRONG
Storage Account Name: "sanygst"): executing request: unexpected status 403 (403 This request is not authorized to perform this operation using this permission.) with AuthorizationPermissionMismatch: This request is not authorized to perform this operation using this permission.`,
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAuthorizationPermissionMismatchError(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAuthorizationPermissionMismatchError() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseAuthorizationPermissionMismatchError() = %v, want %v", got, tt.want)
			}
		})
	}
}
