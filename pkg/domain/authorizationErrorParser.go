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
	log.Debugf("Attempting to Parse Authorization Error: %s", authErrMesg)
	if authErrMesg != "" && !strings.Contains(authErrMesg, "AuthorizationFailed") && !strings.Contains(authErrMesg, "Authorization failed") && !strings.Contains(authErrMesg, "AuthorizationPermissionMismatch") && !strings.Contains(authErrMesg, "LinkedAccessCheckFailed") && !strings.Contains(authErrMesg, "LackOfPermissions") {
		log.Warnln("Non Authorization Error when creating deployment:", authErrMesg)
		return nil, errors.New("Could not parse deploment error, potentially due to a Non-Authorization error")
	}

	resMap := make(map[string][]string)
	var tempMaps []map[string][]string

	// Process LinkedAuthorizationFailed errors
	if strings.Count(authErrMesg, "LinkedAuthorizationFailed") >= 1 {
		log.Debug("Parsing LinkedAuthorizationFailed Error")
		tempMap, err := parseLinkedAuthorizationFailedErrors(authErrMesg)
		if err != nil {
			return nil, err
		}
		tempMaps = append(tempMaps, tempMap)
	}

	// Process LackOfPermissions errors
	if strings.Count(authErrMesg, "LackOfPermissions") >= 1 {
		log.Debug("Parsing LackOfPermissions Error")
		tempMap, err := parseLackOfPermissionsError(authErrMesg)
		if err != nil {
			return nil, err
		}
		tempMaps = append(tempMaps, tempMap)
	}

	// Process AuthorizationFailed errors
	if strings.Count(authErrMesg, "\"AuthorizationFailed") >= 1 || strings.Count(authErrMesg, "AuthorizationFailed:") >= 1 {
		log.Debug("Parsing AuthorizationFailed Error")
		tempMap, err := parseMultiAuthorizationFailedErrors(authErrMesg)
		if err != nil {
			return nil, err
		}
		tempMaps = append(tempMaps, tempMap)
	}

	// Process Authorization failed errors
	if strings.Count(authErrMesg, "Authorization failed") >= 1 {
		log.Debug("Parsing Authorization failed Error")
		tempMap, err := parseMultiAuthorizationErrors(authErrMesg)
		if err != nil {
			return nil, err
		}
		tempMaps = append(tempMaps, tempMap)
	}

	// Process AuthorizationPermissionMismatch errors
	if strings.Count(authErrMesg, "AuthorizationPermissionMismatch") >= 1 {
		log.Debug("Parsing AuthorizationPermissionMismatch Error")
		tempMap, err := parseAuthorizationPermissionMismatchError(authErrMesg)
		if err != nil {
			return nil, err
		}
		tempMaps = append(tempMaps, tempMap)
	}

	// Process LinkedAccessCheckFailed errors
	if strings.Count(authErrMesg, "LinkedAccessCheckFailed") >= 1 {
		log.Debug("Parsing LinkedAccessCheckFailed Error")
		tempMap, err := parseLinkedAccessCheckFailedError(authErrMesg)
		if err != nil {
			return nil, err
		}
		tempMaps = append(tempMaps, tempMap)
	}

	// Merge all temp maps into the result map
	for _, tempMap := range tempMaps {
		mergePermissionMaps(resMap, tempMap)
	}

	// If map is empty, return error
	if len(resMap) == 0 {
		return nil, fmt.Errorf("Could not parse deployment error for scope/permissions: %s", authErrMesg)
	}

	return appendPermissionsForSpecialCases(resMap), nil
}

// mergePermissionMaps merges the source map into the destination map,
// ensuring unique permissions for each scope
func mergePermissionMaps(dest, src map[string][]string) {
	for scope, permissions := range src {
		if _, exists := dest[scope]; !exists {
			dest[scope] = make([]string, 0)
		}

		// Add permissions, avoiding duplicates
		for _, permission := range permissions {
			found := false
			for _, existing := range dest[scope] {
				if existing == permission {
					found = true
					break
				}
			}
			if !found {
				dest[scope] = append(dest[scope], permission)
			}
		}
	}
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
