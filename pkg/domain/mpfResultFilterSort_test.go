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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUniqueSlice(t *testing.T) {

	t.Parallel()

	tests := []struct {
		name     string
		slice    []string
		expected []string
	}{
		{
			name:     "Empty slice",
			slice:    []string{},
			expected: []string{},
		},
		{
			name:     "Slice with duplicate values",
			slice:    []string{"apple", "banana", "apple", "orange", "banana"},
			expected: []string{"apple", "banana", "orange"},
		},
		{
			name:     "Slice with unique values",
			slice:    []string{"apple", "banana", "orange"},
			expected: []string{"apple", "banana", "orange"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getUniqueSlice(tt.slice)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMapWithUniqueValues(t *testing.T) {

	t.Parallel()

	tests := []struct {
		name     string
		input    map[string][]string
		expected map[string][]string
	}{
		{
			name:     "Empty map",
			input:    map[string][]string{},
			expected: map[string][]string{},
		},
		{
			name: "Map with empty slices",
			input: map[string][]string{
				"key1": {},
				"key2": {},
			},
			expected: map[string][]string{
				"key1": {},
				"key2": {},
			},
		},
		{
			name: "Map with duplicate values",
			input: map[string][]string{
				"key1": {"apple", "banana", "apple", "orange", "banana"},
				"key2": {"apple", "banana", "apple", "orange", "banana"},
			},
			expected: map[string][]string{
				"key1": {"apple", "banana", "orange"},
				"key2": {"apple", "banana", "orange"},
			},
		},
		{
			name: "Map with unique values",
			input: map[string][]string{
				"key1": {"apple", "banana", "orange"},
				"key2": {"apple", "banana", "orange"},
			},
			expected: map[string][]string{
				"key1": {"apple", "banana", "orange"},
				"key2": {"apple", "banana", "orange"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMapWithUniqueValues(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}

}
