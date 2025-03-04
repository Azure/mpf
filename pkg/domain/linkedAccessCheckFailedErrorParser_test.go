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

func TestParseLinkedAccessCheckFailedError(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string][]string
		wantErr bool
	}{
		{
			name:  "Valid error message",
			input: "Virtual Network Gateway Name: \"vpn-bhlqb\"): performing CreateOrUpdate: unexpected status 400 (400 Bad Request) with error: LinkedAccessCheckFailed: The client with object id 'ddfcf162-2cf2-40cf-bd4a-49a63e248436' does not have authorization to perform action 'Microsoft.Network/publicIPAddresses/join/action' over scope '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/rg-bhlqb/providers/Microsoft.Network/publicIPAddresses/vpn-pip-bhlqb' or the scope is invalid. For details on the required permissions, please visit 'https://aka.ms/vngwroles'. If access was recently granted, please refresh your credentials.",
			want: map[string][]string{
				"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/rg-bhlqb/providers/Microsoft.Network/publicIPAddresses/vpn-pip-bhlqb": {"Microsoft.Network/publicIPAddresses/join/action"},
			},
			wantErr: false,
		},
		{
			name:    "Invalid error message",
			input:   "Invalid error message",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Edge case with empty input",
			input:   "",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLinkedAccessCheckFailedError(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLinkedAccessCheckFailedError() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseLinkedAccessCheckFailedError() = %v, want %v", got, tt.want)
			}
		})
	}
}
