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

// errWriter wraps an io.Writer and remembers the first write error so callers
// can check it once at the end instead of after every write.
type errWriter struct {
	w   io.Writer
	err error
}

func (ew *errWriter) println(a ...any) {
	if ew.err != nil {
		return
	}
	_, ew.err = fmt.Fprintln(ew.w, a...)
}

func (ew *errWriter) printf(format string, a ...any) {
	if ew.err != nil {
		return
	}
	_, ew.err = fmt.Fprintf(ew.w, format, a...)
}

func displayRoleSuggestionText(w io.Writer, suggestion domain.RoleSuggestion) error {
	ew := &errWriter{w: w}

	ew.println(roleSuggestionSeparator)
	ew.println("Suggested Built-In Roles:")
	ew.println(roleSuggestionSeparator)

	if len(suggestion.SingleRoleMatches) > 0 {
		ew.println("The following built-in role(s) each cover ALL required permissions (most specific first):")
		ew.println()
		displayCount := len(suggestion.SingleRoleMatches)
		if displayCount > maxSingleRoleMatchesDisplayed {
			displayCount = maxSingleRoleMatchesDisplayed
		}
		for _, sr := range suggestion.SingleRoleMatches[:displayCount] {
			ew.printf("  - %s (%s)\n", sr.Role.RoleName, sr.Role.RoleDefinitionID)
		}
		if len(suggestion.SingleRoleMatches) > displayCount {
			ew.printf("  ... and %d more\n", len(suggestion.SingleRoleMatches)-displayCount)
		}
		ew.println()
	} else {
		ew.println("No single built-in role covers all required permissions.")
		ew.println()
	}

	if len(suggestion.MinimalCombination) > 0 {
		ew.println("Suggested minimal combination of built-in roles to cover the required permissions:")
		ew.println()
		for _, sr := range suggestion.MinimalCombination {
			ew.printf("  - %s (%s) covers %d permission(s):\n", sr.Role.RoleName, sr.Role.RoleDefinitionID, len(sr.CoveredPermissions))
			for _, perm := range sr.CoveredPermissions {
				ew.printf("      %s\n", perm)
			}
		}
		ew.println()
	}

	if len(suggestion.UncoveredPermissions) > 0 {
		ew.println("The following required permissions are NOT covered by any built-in role.")
		ew.println("A custom role is required to grant these:")
		for _, perm := range suggestion.UncoveredPermissions {
			ew.printf("  %s\n", perm)
		}
		ew.println()
	}

	ew.println(roleSuggestionSeparator)
	return ew.err
}
