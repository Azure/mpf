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
	"log"
	"regexp"
)

type PermissionErrorDetails struct {
	MissingPermissions string
	Scope              string
}

func parseAuthorizationPermissionMismatchError(authorizationFailedErrMsg string) (map[string][]string, error) {

	log.Printf("Attempting to Parse AuthorizationPermissionMismatch Error: @@%s@@", authorizationFailedErrMsg)
	re := regexp.MustCompile(`retrieving (queue|file|blob) properties for Storage Account \(Subscription: \"([^"]+)\"\nResource Group Name: \"([^"]+)\"\nStorage Account Name: \"([^"]+)\"\): executing request: unexpected status 403 \(403 This request is not authorized to perform this operation using this permission.\) with AuthorizationPermissionMismatch: This request is not authorized to perform this operation using this permission.`)

	matches := re.FindAllStringSubmatch(authorizationFailedErrMsg, -1)

	if len(matches) == 0 {
		return nil, errors.New("no matches found in 'AuthorizationPermissionMismatch' error message")
	}

	scopePermissionsMap := make(map[string][]string)

	// Iterate through the matches and populate the map
	for _, match := range matches {
		if len(match) == 5 {
			var action string
			switch match[1] {
			case "queue":
				action = "Microsoft.Storage/storageAccounts/queueServices/read"
			case "file":
				action = "Microsoft.Storage/storageAccounts/fileServices/read"
			case "blob":
				action = "Microsoft.Storage/storageAccounts/blobServices/read"
			}

			scope := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Storage/storageAccounts/%s", match[2], match[3], match[4])

			if _, ok := scopePermissionsMap[scope]; !ok {
				scopePermissionsMap[scope] = make([]string, 0)
			}
			scopePermissionsMap[scope] = append(scopePermissionsMap[scope], action)
		}
	}

	// if map is empty, return error
	if len(scopePermissionsMap) == 0 {
		return nil, errors.New("no scope/permissions found in AuthorizationPermissionMismatch error message")
	}

	return scopePermissionsMap, nil

}
