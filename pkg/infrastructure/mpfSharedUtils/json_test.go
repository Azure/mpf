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

package mpfSharedUtils

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadJson(t *testing.T) {
	// Create a temporary JSON file for testing
	tempFile, err := os.CreateTemp("", "test.json")
	assert.Nil(t, err)
	defer os.Remove(tempFile.Name())

	// Define a sample JSON content
	jsonContent := `{"name": "John Doe", "age": 30}`

	// Write the JSON content to the temporary file
	err = os.WriteFile(tempFile.Name(), []byte(jsonContent), 0644)
	assert.Nil(t, err)

	// Call the ReadJson function with the temporary file path
	result, err := ReadJson(tempFile.Name())

	// Assert that there is no error
	assert.Nil(t, err)

	// Assert that the result matches the expected JSON content
	expected := make(map[string]interface{})
	err = json.Unmarshal([]byte(jsonContent), &expected)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}
