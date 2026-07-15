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

package presentation

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/Azure/mpf/pkg/domain"
)

// maxSingleRoleMatchesDisplayed caps how many single-role matches are shown in
// text output to keep the suggestion readable. The most specific roles are shown
// first, so broad roles such as Owner and Contributor appear last (or are hidden).
const maxSingleRoleMatchesDisplayed = 10

const roleSuggestionSeparator = "------------------------------------------------------------------------------------------------------------------------------------------"

// DisplayRoleSuggestion writes the built-in role suggestion to w, either as JSON
// (when jsonOutput is true) or as human readable text.
func DisplayRoleSuggestion(w io.Writer, suggestion domain.RoleSuggestion, jsonOutput bool) error {
	if jsonOutput {
		return displayRoleSuggestionJSON(w, suggestion)
	}
	return displayRoleSuggestionText(w, suggestion)
}

func displayRoleSuggestionJSON(w io.Writer, suggestion domain.RoleSuggestion) error {
	jsonBytes, err := json.MarshalIndent(suggestion, "", "  ")
	if err != nil {
		return err
	}
	_, err = w.Write(jsonBytes)
	return err
}

func displayRoleSuggestionText(w io.Writer, suggestion domain.RoleSuggestion) error {
	fmt.Fprintln(w, roleSuggestionSeparator)
	fmt.Fprintln(w, "Suggested Built-In Roles:")
	fmt.Fprintln(w, roleSuggestionSeparator)

	if len(suggestion.SingleRoleMatches) > 0 {
		fmt.Fprintln(w, "The following built-in role(s) each cover ALL required permissions (most specific first):")
		fmt.Fprintln(w)
		displayCount := len(suggestion.SingleRoleMatches)
		if displayCount > maxSingleRoleMatchesDisplayed {
			displayCount = maxSingleRoleMatchesDisplayed
		}
		for _, sr := range suggestion.SingleRoleMatches[:displayCount] {
			fmt.Fprintf(w, "  - %s (%s)\n", sr.Role.RoleName, sr.Role.RoleDefinitionID)
		}
		if len(suggestion.SingleRoleMatches) > displayCount {
			fmt.Fprintf(w, "  ... and %d more\n", len(suggestion.SingleRoleMatches)-displayCount)
		}
		fmt.Fprintln(w)
	} else {
		fmt.Fprintln(w, "No single built-in role covers all required permissions.")
		fmt.Fprintln(w)
	}

	if len(suggestion.MinimalCombination) > 0 {
		fmt.Fprintln(w, "Suggested minimal combination of built-in roles to cover the required permissions:")
		fmt.Fprintln(w)
		for _, sr := range suggestion.MinimalCombination {
			fmt.Fprintf(w, "  - %s (%s) covers %d permission(s):\n", sr.Role.RoleName, sr.Role.RoleDefinitionID, len(sr.CoveredPermissions))
			for _, perm := range sr.CoveredPermissions {
				fmt.Fprintf(w, "      %s\n", perm)
			}
		}
		fmt.Fprintln(w)
	}

	if len(suggestion.UncoveredPermissions) > 0 {
		fmt.Fprintln(w, "The following required permissions are NOT covered by any built-in role.")
		fmt.Fprintln(w, "A custom role is required to grant these:")
		for _, perm := range suggestion.UncoveredPermissions {
			fmt.Fprintf(w, "  %s\n", perm)
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w, roleSuggestionSeparator)
	return nil
}
