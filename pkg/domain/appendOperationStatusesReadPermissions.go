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

import "strings"

// OperationStatusesReadSuffix is the action suffix that the Azure Resource Manager
// long running operation (LRO) polling endpoint is protected by.
const OperationStatusesReadSuffix = "/operationStatuses/read"

// GetOperationStatusesReadPermission returns the LRO polling permission that corresponds to
// the supplied write permission, for example:
//
//	Microsoft.ContainerRegistry/registries/write -> Microsoft.ContainerRegistry/registries/operationStatuses/read
//
// The second return value is false when the supplied permission is not a resource type write
// action that could require LRO polling.
//
// Background: azurerm provider resources created via an LRO flavoured API (CreateThenPoll,
// CreateOrUpdateThenPoll, ...) issue a PUT that returns 201/202 and then poll the returned
// operationStatuses URL. Without the polling permission Terraform reports a failure even though
// the resource was created, which leaves the resource outside of the Terraform state.
// See https://github.com/hashicorp/terraform-provider-azurerm/issues/29047
func GetOperationStatusesReadPermission(permission string) (string, bool) {
	permission = strings.TrimSpace(permission)

	if !strings.HasSuffix(permission, "/write") {
		return "", false
	}

	resourceType := strings.TrimSuffix(permission, "/write")

	// A resource type action is at minimum <Namespace>/<resourceType>, so it needs to contain
	// a namespace segment and at least one resource type segment.
	if strings.Count(resourceType, "/") < 1 {
		return "", false
	}

	// Wildcard actions such as Microsoft.Storage/*/write cannot be mapped to a concrete
	// operationStatuses action.
	if strings.Contains(resourceType, "*") {
		return "", false
	}

	// Guard against double-appending for actions that already target operationStatuses.
	if strings.HasSuffix(strings.ToLower(resourceType), "/operationstatuses") {
		return "", false
	}

	return resourceType + OperationStatusesReadSuffix, true
}

// AppendOperationStatusesReadPermissions appends the LRO polling read permission for every
// resource type write permission found in the supplied scope/permission map.
//
// Permissions that Azure does not recognise are removed later on, when the custom role update
// reports them as InvalidActionOrNotAction.
func AppendOperationStatusesReadPermissions(scpPerms map[string][]string) map[string][]string {
	for scope, perms := range scpPerms {
		var toAppend []string
		for _, perm := range perms {
			if opStatusPerm, ok := GetOperationStatusesReadPermission(perm); ok {
				toAppend = append(toAppend, opStatusPerm)
			}
		}
		if len(toAppend) > 0 {
			scpPerms[scope] = append(scpPerms[scope], toAppend...)
		}
	}
	return scpPerms
}
