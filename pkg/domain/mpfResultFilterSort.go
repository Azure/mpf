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
	"sort"
	"strings"
)

// FilterOutPermissions removes the supplied permissions from every scope of the
// scope/permission map. Comparison is case-insensitive because Azure action names
// are case-insensitive. The supplied map is not modified.
func FilterOutPermissions(scpPerms map[string][]string, permissionsToRemove []string) map[string][]string {
	if len(permissionsToRemove) == 0 {
		return scpPerms
	}

	removeSet := make(map[string]bool, len(permissionsToRemove))
	for _, perm := range permissionsToRemove {
		removeSet[strings.ToLower(strings.TrimSpace(perm))] = true
	}

	filtered := make(map[string][]string, len(scpPerms))
	for scope, perms := range scpPerms {
		kept := make([]string, 0, len(perms))
		for _, perm := range perms {
			if removeSet[strings.ToLower(strings.TrimSpace(perm))] {
				continue
			}
			kept = append(kept, perm)
		}
		filtered[scope] = kept
	}
	return filtered
}

func getMapWithUniqueValues(m map[string][]string) map[string][]string {
	sm := make(map[string][]string)
	for key, vals := range m {
		vals = getUniqueSlice(vals)
		sort.Strings(vals)
		sm[key] = vals
	}
	return sm
}

func getUniqueSlice(s []string) []string {
	uniqueSlice := make([]string, 0, len(s))
	m := make(map[string]bool)
	for _, val := range s {
		if _, ok := m[val]; !ok {
			m[val] = true
			uniqueSlice = append(uniqueSlice, val)
		}
	}
	return uniqueSlice
}
