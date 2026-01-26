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

package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseInitialPermissions(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        []string
		wantErr     bool
		setupFile   bool
		fileContent string
	}{
		{
			name:    "empty string returns nil",
			input:   "",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "single permission",
			input:   "Microsoft.Storage/storageAccounts/read",
			want:    []string{"Microsoft.Storage/storageAccounts/read"},
			wantErr: false,
		},
		{
			name:  "comma-separated permissions",
			input: "Microsoft.Storage/storageAccounts/read,Microsoft.Storage/storageAccounts/write",
			want: []string{
				"Microsoft.Storage/storageAccounts/read",
				"Microsoft.Storage/storageAccounts/write",
			},
			wantErr: false,
		},
		{
			name:  "comma-separated permissions with spaces",
			input: "Microsoft.Storage/storageAccounts/read, Microsoft.Storage/storageAccounts/write , Microsoft.Storage/storageAccounts/listKeys/action",
			want: []string{
				"Microsoft.Storage/storageAccounts/read",
				"Microsoft.Storage/storageAccounts/write",
				"Microsoft.Storage/storageAccounts/listKeys/action",
			},
			wantErr: false,
		},
		{
			name:      "file reference with valid JSON",
			input:     "@testperms.json",
			setupFile: true,
			fileContent: `{
				"RequiredPermissions": {
					"": [
						"Microsoft.Storage/storageAccounts/read",
						"Microsoft.Storage/storageAccounts/write"
					]
				}
			}`,
			want: []string{
				"Microsoft.Storage/storageAccounts/read",
				"Microsoft.Storage/storageAccounts/write",
			},
			wantErr: false,
		},
		{
			name:    "file reference to non-existent file",
			input:   "@nonexistent.json",
			want:    nil,
			wantErr: true,
		},
		{
			name:        "file reference with invalid JSON",
			input:       "@invalid.json",
			setupFile:   true,
			fileContent: `not valid json`,
			want:        nil,
			wantErr:     true,
		},
		{
			name:      "file reference with empty permissions array",
			input:     "@empty.json",
			setupFile: true,
			fileContent: `{
				"RequiredPermissions": {
					"": []
				}
			}`,
			want:    []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup file if needed
			if tt.setupFile {
				filename := tt.input[1:] // Remove @ prefix
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, filename)
				err := os.WriteFile(filePath, []byte(tt.fileContent), 0644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				// Update input to use absolute path
				tt.input = "@" + filePath
			}

			got, err := parseInitialPermissions(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInitialPermissions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseInitialPermissions() = %v, want %v", got, tt.want)
			}
		})
	}
}
