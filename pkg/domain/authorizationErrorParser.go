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
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

func GetScopePermissionsFromAuthError(authErrMesg string) (map[string][]string, error) {
	if authErrMesg != "" && !strings.Contains(authErrMesg, "AuthorizationFailed") && !strings.Contains(authErrMesg, "Authorization failed") {
		log.Infoln("Non Authorization Error when creating deployment:", authErrMesg)
		return nil, errors.New("Could not parse deploment error, potentially due to a Non-Authorization error")
	}

	var resMap map[string][]string
	var err error

	switch {
	case strings.Count(authErrMesg, "LinkedAuthorizationFailed") >= 1:
		resMap, err = parseLinkedAuthorizationFailedErrors(authErrMesg)
	case strings.Count(authErrMesg, "AuthorizationFailed") >= 1:
		resMap, err = parseMultiAuthorizationFailedErrors(authErrMesg)
	case strings.Count(authErrMesg, "Authorization failed") >= 1:
		resMap, err = parseMultiAuthorizationErrors(authErrMesg)
	}

	if err != nil {
		return nil, err
	}

	// If map is empty, return error
	if len(resMap) == 0 {
		return nil, errors.New(fmt.Sprintf("Could not parse deployment error for scope/permissions: %s", authErrMesg))
	}

	return resMap, nil
}

// For 'AuthorizationFailed' errors
func parseMultiAuthorizationFailedErrors(authorizationFailedErrMsg string) (map[string][]string, error) {

	re := regexp.MustCompile(`The client '([^']+)' with object id '([^']+)' does not have authorization to perform action '([^']+)'.* over scope '([^']+)' or the scope is invalid\.`)

	matches := re.FindAllStringSubmatch(authorizationFailedErrMsg, -1)

	if len(matches) == 0 {
		return nil, errors.New("No matches found in 'AuthorizationFailed' error message")
	}

	scopePermissionsMap := make(map[string][]string)

	// Iterate through the matches and populate the map
	for _, match := range matches {
		if len(match) == 5 {
			// resourceType := match[1]
			action := match[3]
			scope := match[4]

			if _, ok := scopePermissionsMap[scope]; !ok {
				scopePermissionsMap[scope] = make([]string, 0)
			}
			scopePermissionsMap[scope] = append(scopePermissionsMap[scope], action)
		}
	}

	// if map is empty, return error
	if len(scopePermissionsMap) == 0 {
		return nil, errors.New("No scope/permissions found in Multi error message")
	}

	return scopePermissionsMap, nil

}

// For 'Authorization failed' errors
func parseMultiAuthorizationErrors(authorizationFailedErrMsg string) (map[string][]string, error) {

	// Regular expression to extract resource information
	re := regexp.MustCompile(`Authorization failed for template resource '([^']+)' of type '([^']+)'\. The client '([^']+)' with object id '([^']+)' does not have permission to perform action '([^']+)' at scope '([^']+)'\.`)

	// Find all matches in the error message
	matches := re.FindAllStringSubmatch(authorizationFailedErrMsg, -1)

	// If No Matches found return error
	if len(matches) == 0 {
		return nil, errors.New("No matches found in 'Authorization failed' error message")
	}

	// Create a map to store scope/permissions
	scopePermissionsMap := make(map[string][]string)

	// Iterate through the matches and populate the map
	for _, match := range matches {
		if len(match) == 7 {
			// resourceType := match[1]
			action := match[5]
			scope := match[6]

			if _, ok := scopePermissionsMap[scope]; !ok {
				scopePermissionsMap[scope] = make([]string, 0)
			}
			scopePermissionsMap[scope] = append(scopePermissionsMap[scope], action)
		}
	}

	// if map is empty, return error
	if len(scopePermissionsMap) == 0 {
		return nil, errors.New("No scope/permissions found in Multi error message")
	}

	return scopePermissionsMap, nil

}

// For 'LinkedAuthorizationFailed' errors
func parseLinkedAuthorizationFailedErrors(authorizationFailedErrMsg string) (map[string][]string, error) {

	// Regular expression to extract resource information
	// re := regexp.MustCompile(`Authorization failed for template resource '([^']+)' of type '([^']+)'\. The client '([^']+)' with object id '([^']+)' does not have permission to perform action '([^']+)' at scope '([^']+)'\.`)

	// Find regular expressions to pull action and scope from error message "does not have permission to perform action(s) 'Microsoft.Network/virtualNetworks/subnets/join/action' on the linked scope(s) '/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/az-mpf-tf-test-rg/providers/Microsoft.Network/virtualNetworks/vnet-32a70ccbb3247e2b/subnets/subnet-32a70ccbb3247e2b' (respectively) or the linked scope(s) are invalid".
	re := regexp.MustCompile(`does not have permission to perform action\(s\) '([^']+)' on the linked scope\(s\) '([^']+)' \(respectively\) or the linked scope\(s\) are invalid`)

	// Find all matches in the error message
	matches := re.FindAllStringSubmatch(authorizationFailedErrMsg, -1)

	// If No Matches found return error
	if len(matches) == 0 {
		return nil, errors.New("No matches found in 'Authorization failed' error message")
	}

	// Create a map to store scope/permissions
	scopePermissionsMap := make(map[string][]string)

	// Iterate through the matches and populate the map
	for _, match := range matches {
		if len(match) == 3 {
			// resourceType := match[1]
			action := match[1]
			scope := match[2]

			if _, ok := scopePermissionsMap[scope]; !ok {
				scopePermissionsMap[scope] = make([]string, 0)
			}
			scopePermissionsMap[scope] = append(scopePermissionsMap[scope], action)
		}
	}

	// if map is empty, return error
	if len(scopePermissionsMap) == 0 {
		return nil, errors.New("No scope/permissions found in Multi error message")
	}

	return scopePermissionsMap, nil

}
