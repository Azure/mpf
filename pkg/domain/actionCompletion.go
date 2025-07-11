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
	"fmt"
	"regexp"

	log "github.com/sirupsen/logrus"
)

// completePartialAction converts incomplete actions to proper Azure resource provider format
// based on the scope pattern. Returns the completed action and any error encountered.
func completePartialAction(action, scope string) (string, error) {
	// Handle incomplete actions that need to be converted to proper Azure resource provider format
	// if action is blank or "" return error
	if action == "" {
		return "", fmt.Errorf("action cannot be empty")
	}

	if action == "/read" {
		// Compile regex patterns for scope type detection
		resourceGroupScopeRe := regexp.MustCompile(`(?i)^/subscriptions/[^/]+/resourceGroups/[^/]+$`)
		subscriptionScopeRe := regexp.MustCompile(`^/subscriptions/[^/]+$`)

		// Use compiled regex patterns to determine the correct action based on the scope pattern
		if resourceGroupScopeRe.MatchString(scope) {
			// This is a resource group scope pattern: /subscriptions/{subscription-id}/resourceGroups/{resource-group-name}
			completedAction := "Microsoft.Resources/subscriptions/resourceGroups/read"
			log.Debugf("Completed partial action '%s' to '%s' for resource group scope: %s", action, completedAction, scope)
			return completedAction, nil
		}

		if subscriptionScopeRe.MatchString(scope) {
			// This is a subscription scope pattern: /subscriptions/{subscription-id}
			completedAction := "Microsoft.Resources/subscriptions/read"
			log.Debugf("Completed partial action '%s' to '%s' for subscription scope: %s", action, completedAction, scope)
			return completedAction, nil
		}

		// Add more scope type mappings as needed
		log.Warnf("Unable to complete partial action '%s' for unrecognized scope pattern: %s", action, scope)
		return action, fmt.Errorf("unable to complete partial action '%s' for unrecognized scope pattern: %s", action, scope)
	}

	// Return the original action if no completion is needed
	return action, nil
}
