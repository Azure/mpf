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

package main

import (
	"context"
	"os"

	"github.com/Azure/mpf/pkg/domain"
	roledefinitionmanager "github.com/Azure/mpf/pkg/infrastructure/roleDefinitionManager"
	"github.com/Azure/mpf/pkg/presentation"

	log "github.com/sirupsen/logrus"
)

// displayResultAndRoleSuggestion writes the permissions result and, when
// --suggestRoles is set, the suggested built-in role(s). With --jsonOutput both are
// emitted as a single JSON document so that stdout stays valid JSON.
func displayResultAndRoleSuggestion(ctx context.Context, subscriptionID string, mpfResult domain.MPFResult, displayOptions presentation.DisplayOptions) {
	suggestion, ok := getRoleSuggestion(ctx, subscriptionID, mpfResult)

	if ok && displayOptions.JSONOutput {
		if err := presentation.DisplayCombinedJSON(os.Stdout, mpfResult, suggestion); err != nil {
			log.Errorf("Error displaying result: %v", err)
		}
		return
	}

	displayResult(mpfResult, displayOptions)

	if !ok {
		return
	}

	if err := presentation.DisplayRoleSuggestion(os.Stdout, suggestion, displayOptions.JSONOutput); err != nil {
		log.Errorf("Error displaying role suggestion: %v", err)
	}
}

// getRoleSuggestion fetches the Azure built-in role definitions and matches them
// against the required permissions discovered by MPF. The second return value is
// false when the suggestion was not requested or could not be produced, in which
// case failures are logged but do not abort the command.
func getRoleSuggestion(ctx context.Context, subscriptionID string, mpfResult domain.MPFResult) (domain.RoleSuggestion, bool) {
	if !flgSuggestRoles {
		return domain.RoleSuggestion{}, false
	}

	requiredPermissions := flattenRequiredPermissions(mpfResult.RequiredPermissions)
	if len(requiredPermissions) == 0 {
		log.Warnln("No permissions available to suggest built-in roles for")
		return domain.RoleSuggestion{}, false
	}

	log.Infoln("Fetching Azure built-in role definitions to suggest matching roles...")
	roleProvider := roledefinitionmanager.NewRoleDefinitionManager(subscriptionID)
	builtInRoles, err := roleProvider.GetBuiltInRoles(ctx, subscriptionID)
	if err != nil {
		log.Errorf("Error fetching built-in roles for role suggestion: %v", err)
		return domain.RoleSuggestion{}, false
	}

	return domain.SuggestBuiltInRoles(requiredPermissions, builtInRoles), true
}

// flattenRequiredPermissions collects the unique permissions across all scopes in
// the MPF result, since a role assignment grants the union of these permissions.
func flattenRequiredPermissions(requiredPermissions map[string][]string) []string {
	seen := make(map[string]bool)
	var all []string
	for _, perms := range requiredPermissions {
		for _, perm := range perms {
			if perm == "" || seen[perm] {
				continue
			}
			seen[perm] = true
			all = append(all, perm)
		}
	}
	return all
}
