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
	"errors"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

func GetAddressAndResourceIDFromExistingResourceError(existingResourceErr string) (map[string]string, error) {
	if existingResourceErr != "" && !strings.Contains(existingResourceErr, TFExistingResourceErrorMsg) {
		log.Infoln("Non existing resource error :", existingResourceErr)
		return nil, errors.New("Non existing resource error")
	}

	var resMap map[string]string
	resMap = make(map[string]string)
	// var err error
	re := regexp.MustCompile(`Error: A resource with the ID "([^"]+)" already exists - to be managed via Terraform this resource needs to be imported into the State. Please see the resource documentation for "([^"]+)" for more information.\n\n  with ([^,]+),`)
	// re := regexp.MustCompile(`Error: A resource with the ID "([^']+)" already exists - to be managed via Terraform this resource needs to be imported into the State. Please see the resource documentation for "([^']+)" for more information\.(.*)  with ([^,]+),`)

	matches := re.FindAllStringSubmatch(existingResourceErr, -1)

	if len(matches) == 0 {
		return nil, errors.New("No matches found in 'existingResourceErrorMessage' error message")
	}

	for _, match := range matches {
		if len(match) == 4 {
			// resourceType := match[1]
			resID := match[1]
			addr := match[3]
			resMap[addr] = resID
		}
	}

	return resMap, nil
}
