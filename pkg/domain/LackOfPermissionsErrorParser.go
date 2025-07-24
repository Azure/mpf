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
	"errors"
	"regexp"

	log "github.com/sirupsen/logrus"
)

func parseLackOfPermissionsError(authorizationFailedErrMsg string) (map[string][]string, error) {

	log.Debugf("Attempting to Parse LackOfPermissions Error: %s", authorizationFailedErrMsg)

	// "code": "LackOfPermissions",\r\n"target": "workspace.Kind",\r\n"message": "You do not have Azure RBAC permissions to create new AI hubs. To create an AI hub, you are required to have contributor permissions on your resource group, or more specifically the Microsoft.MachineLearningServices/workspaces/hubs/write permission.",\r\n"additionalInfo": "Please check your permissions and try again."

	re := regexp.MustCompile(`or more specifically the (Microsoft\.([^\.]+)) permission\.`)

	matches := re.FindAllStringSubmatch(authorizationFailedErrMsg, -1)

	if len(matches) == 0 {
		return nil, errors.New("no matches found in 'LackOfPermissions' error message")
	}

	scopePermissionsMap := make(map[string][]string)

	// Iterate through the matches and populate the map
	for _, match := range matches {
		if len(match) == 3 {
			// resourceType := match[1]
			action := match[1]
			scope := "ScopeCannotBeParsedFromLackOfPermissionsError"

			if _, ok := scopePermissionsMap[scope]; !ok {
				scopePermissionsMap[scope] = make([]string, 0)
			}
			scopePermissionsMap[scope] = append(scopePermissionsMap[scope], action)
		}
	}

	// if map is empty, return error
	if len(scopePermissionsMap) == 0 {
		return nil, errors.New("no scope/permissions found in LackOfPermissions message")
	}

	return scopePermissionsMap, nil

}
