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
	"reflect"
	"testing"
)

func TestParseDefaultDataActions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "Single action",
			input:    "Microsoft.Search/searchServices/indexes/documents/read",
			expected: []string{"Microsoft.Search/searchServices/indexes/documents/read"},
		},
		{
			name:     "Multiple actions with commas",
			input:    "Microsoft.Search/searchServices/indexes/documents/read,Microsoft.Search/searchServices/indexes/documents/write",
			expected: []string{"Microsoft.Search/searchServices/indexes/documents/read", "Microsoft.Search/searchServices/indexes/documents/write"},
		},
		{
			name:     "Multiple actions with spaces",
			input:    "Microsoft.Search/searchServices/indexes/documents/read, Microsoft.Search/searchServices/indexes/documents/write",
			expected: []string{"Microsoft.Search/searchServices/indexes/documents/read", "Microsoft.Search/searchServices/indexes/documents/write"},
		},
		{
			name:     "Multiple actions with extra spaces",
			input:    "  Microsoft.Search/searchServices/indexes/documents/read  ,  Microsoft.Search/searchServices/indexes/documents/write  ",
			expected: []string{"Microsoft.Search/searchServices/indexes/documents/read", "Microsoft.Search/searchServices/indexes/documents/write"},
		},
		{
			name:     "Actions with empty strings after split",
			input:    "Microsoft.Search/searchServices/indexes/documents/read,,Microsoft.Search/searchServices/indexes/documents/write",
			expected: []string{"Microsoft.Search/searchServices/indexes/documents/read", "Microsoft.Search/searchServices/indexes/documents/write"},
		},
		{
			name:     "Only commas and spaces",
			input:    ", , ,",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDefaultDataActions(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseDefaultDataActions(%q) = %v (len=%d), want %v (len=%d)", tt.input, result, len(result), tt.expected, len(tt.expected))
			}
		})
	}
}