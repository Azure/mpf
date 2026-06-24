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

// Package bicepUtils contains helpers for working with Bicep input files,
// such as compiling .bicepparam files to ARM JSON parameter files.
package bicepUtils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// IsBicepParamFile reports whether the given path refers to a .bicepparam file.
func IsBicepParamFile(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".bicepparam")
}

// CompileBicepParamsToTempFile compiles the given .bicepparam file to an ARM
// JSON parameters file using `bicep build-params`. The output is written to a
// freshly-created temp file, so any pre-existing "<name>.parameters.json" next
// to the source file is left untouched.
//
// The caller is responsible for removing the returned path when no longer
// needed (typically via defer / t.Cleanup). If an error occurs, the temp file
// is removed before returning.
func CompileBicepParamsToTempFile(bicepExecPath, paramsFilePath string) (string, error) {
	tempFile, err := os.CreateTemp("", "mpf-bicepparam-*.parameters.json")
	if err != nil {
		return "", fmt.Errorf("creating temporary ARM parameters file: %w", err)
	}
	compiledPath := tempFile.Name()
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(compiledPath)
		return "", fmt.Errorf("closing temporary ARM parameters file: %w", err)
	}

	cmd := exec.Command(bicepExecPath, "build-params", paramsFilePath, "--outfile", compiledPath)
	cmd.Dir = filepath.Dir(paramsFilePath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		_ = os.Remove(compiledPath)
		return "", fmt.Errorf("running bicep build-params: %w\n%s", err, string(output))
	}

	return compiledPath, nil
}
