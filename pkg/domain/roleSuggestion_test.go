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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActionMatchesPattern(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		action   string
		expected bool
	}{
		{"exact match", "Microsoft.Storage/storageAccounts/read", "Microsoft.Storage/storageAccounts/read", true},
		{"exact mismatch", "Microsoft.Storage/storageAccounts/read", "Microsoft.Storage/storageAccounts/write", false},
		{"case insensitive exact", "microsoft.storage/storageaccounts/read", "Microsoft.Storage/storageAccounts/read", true},
		{"global wildcard", "*", "Microsoft.Storage/storageAccounts/read", true},
		{"provider wildcard match", "Microsoft.Storage/*", "Microsoft.Storage/storageAccounts/read", true},
		{"provider wildcard mismatch", "Microsoft.Storage/*", "Microsoft.Compute/virtualMachines/read", false},
		{"suffix wildcard read match", "*/read", "Microsoft.Storage/storageAccounts/read", true},
		{"suffix wildcard read mismatch", "*/read", "Microsoft.Storage/storageAccounts/write", false},
		{"middle wildcard match", "Microsoft.Compute/*/read", "Microsoft.Compute/virtualMachines/read", true},
		{"middle wildcard mismatch action", "Microsoft.Compute/*/read", "Microsoft.Compute/virtualMachines/write", false},
		{"empty pattern", "", "Microsoft.Storage/storageAccounts/read", false},
		{"resource type wildcard", "Microsoft.Storage/storageAccounts/*", "Microsoft.Storage/storageAccounts/blobServices/read", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, actionMatchesPattern(tt.pattern, tt.action))
		})
	}
}

func TestRoleCoversPermission(t *testing.T) {
	contributorLike := BuiltInRole{
		RoleName:   "Contributor-like",
		Actions:    []string{"*"},
		NotActions: []string{"Microsoft.Authorization/*/write", "Microsoft.Authorization/*/delete"},
	}

	tests := []struct {
		name       string
		role       BuiltInRole
		permission string
		expected   bool
	}{
		{
			name:       "action covered no notactions",
			role:       BuiltInRole{Actions: []string{"Microsoft.Storage/*"}},
			permission: "Microsoft.Storage/storageAccounts/read",
			expected:   true,
		},
		{
			name:       "action not covered",
			role:       BuiltInRole{Actions: []string{"Microsoft.Storage/*"}},
			permission: "Microsoft.Compute/virtualMachines/read",
			expected:   false,
		},
		{
			name:       "wildcard action but excluded by notaction",
			role:       contributorLike,
			permission: "Microsoft.Authorization/roleAssignments/write",
			expected:   false,
		},
		{
			name:       "wildcard action not excluded",
			role:       contributorLike,
			permission: "Microsoft.Storage/storageAccounts/write",
			expected:   true,
		},
		{
			name:       "empty permission",
			role:       BuiltInRole{Actions: []string{"*"}},
			permission: "",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, roleCoversPermission(tt.role, tt.permission))
		})
	}
}

func TestSuggestBuiltInRoles_SingleRoleMatch(t *testing.T) {
	required := []string{
		"Microsoft.Storage/storageAccounts/read",
		"Microsoft.Storage/storageAccounts/write",
	}
	roles := []BuiltInRole{
		{RoleName: "Owner", Actions: []string{"*"}},
		{RoleName: "Storage Account Contributor", Actions: []string{"Microsoft.Storage/storageAccounts/*", "Microsoft.Insights/read"}},
		{RoleName: "Reader", Actions: []string{"*/read"}},
	}

	suggestion := SuggestBuiltInRoles(required, roles)

	// Owner and Storage Account Contributor both cover all; Reader does not (no write).
	assert.Len(t, suggestion.SingleRoleMatches, 2)
	// Storage Account Contributor is narrower (scoped wildcards) than Owner ("*"),
	// so it ranks first as the least-privilege option.
	assert.Equal(t, "Storage Account Contributor", suggestion.SingleRoleMatches[0].Role.RoleName)
	assert.Equal(t, "Owner", suggestion.SingleRoleMatches[1].Role.RoleName)
	assert.Empty(t, suggestion.UncoveredPermissions)
}

func TestSuggestBuiltInRoles_SpecificityOrdering(t *testing.T) {
	required := []string{"Microsoft.Storage/storageAccounts/read"}
	roles := []BuiltInRole{
		{RoleName: "Owner", Actions: []string{"*"}},
		{RoleName: "Broad", Actions: []string{"Microsoft.Storage/*", "a", "b", "c", "d"}},
		{RoleName: "Specific", Actions: []string{"Microsoft.Storage/storageAccounts/read"}},
	}

	suggestion := SuggestBuiltInRoles(required, roles)

	assert.Len(t, suggestion.SingleRoleMatches, 3)
	// Ordering is by breadth score (lower = narrower) first:
	// Specific (one exact action = 1), Broad (one scoped wildcard + 4 exact = 1004),
	// Owner (global "*" = 1,000,000).
	assert.Equal(t, "Specific", suggestion.SingleRoleMatches[0].Role.RoleName)
	assert.Equal(t, "Broad", suggestion.SingleRoleMatches[1].Role.RoleName)
	assert.Equal(t, "Owner", suggestion.SingleRoleMatches[2].Role.RoleName)
}

func TestSuggestBuiltInRoles_MinimalCombination(t *testing.T) {
	required := []string{
		"Microsoft.Storage/storageAccounts/read",
		"Microsoft.Compute/virtualMachines/read",
	}
	roles := []BuiltInRole{
		{RoleName: "Storage Reader", Actions: []string{"Microsoft.Storage/*"}},
		{RoleName: "Compute Reader", Actions: []string{"Microsoft.Compute/*"}},
	}

	suggestion := SuggestBuiltInRoles(required, roles)

	assert.Empty(t, suggestion.SingleRoleMatches)
	assert.Len(t, suggestion.MinimalCombination, 2)
	assert.Empty(t, suggestion.UncoveredPermissions)

	// Both required permissions are covered exactly once across the combination.
	var covered []string
	for _, sr := range suggestion.MinimalCombination {
		covered = append(covered, sr.CoveredPermissions...)
	}
	assert.ElementsMatch(t, required, covered)
}

func TestSuggestBuiltInRoles_PrefersSingleBroadRoleOverTwo(t *testing.T) {
	required := []string{
		"Microsoft.Storage/storageAccounts/read",
		"Microsoft.Compute/virtualMachines/read",
	}
	roles := []BuiltInRole{
		{RoleName: "Storage Reader", Actions: []string{"Microsoft.Storage/*"}},
		{RoleName: "Compute Reader", Actions: []string{"Microsoft.Compute/*"}},
		{RoleName: "Owner", Actions: []string{"*"}},
	}

	suggestion := SuggestBuiltInRoles(required, roles)

	// Owner covers everything so it is a single-role match.
	assert.Len(t, suggestion.SingleRoleMatches, 1)
	assert.Equal(t, "Owner", suggestion.SingleRoleMatches[0].Role.RoleName)
	// Greedy combination picks the single role covering the most (Owner covers both).
	assert.Len(t, suggestion.MinimalCombination, 1)
	assert.Equal(t, "Owner", suggestion.MinimalCombination[0].Role.RoleName)
}

func TestSuggestBuiltInRoles_UncoveredPermissions(t *testing.T) {
	required := []string{
		"Microsoft.Storage/storageAccounts/read",
		"Microsoft.CustomProvider/customResource/read",
	}
	roles := []BuiltInRole{
		{RoleName: "Storage Reader", Actions: []string{"Microsoft.Storage/*"}},
	}

	suggestion := SuggestBuiltInRoles(required, roles)

	assert.Empty(t, suggestion.SingleRoleMatches)
	assert.Len(t, suggestion.MinimalCombination, 1)
	assert.Equal(t, []string{"Microsoft.CustomProvider/customResource/read"}, suggestion.UncoveredPermissions)
}

func TestSuggestBuiltInRoles_EmptyInputs(t *testing.T) {
	// No required permissions.
	suggestion := SuggestBuiltInRoles([]string{}, []BuiltInRole{{RoleName: "Owner", Actions: []string{"*"}}})
	assert.Empty(t, suggestion.SingleRoleMatches)
	assert.Empty(t, suggestion.MinimalCombination)
	assert.Empty(t, suggestion.UncoveredPermissions)

	// No roles.
	suggestion = SuggestBuiltInRoles([]string{"Microsoft.Storage/storageAccounts/read"}, nil)
	assert.Empty(t, suggestion.SingleRoleMatches)
	assert.Empty(t, suggestion.MinimalCombination)
	assert.Equal(t, []string{"Microsoft.Storage/storageAccounts/read"}, suggestion.UncoveredPermissions)
}

func TestSuggestBuiltInRoles_DeduplicatesAndTrims(t *testing.T) {
	required := []string{
		"Microsoft.Storage/storageAccounts/read",
		"  Microsoft.Storage/storageAccounts/read  ",
		"",
	}
	roles := []BuiltInRole{
		{RoleName: "Storage Reader", Actions: []string{"Microsoft.Storage/*"}},
	}

	suggestion := SuggestBuiltInRoles(required, roles)

	assert.Len(t, suggestion.SingleRoleMatches, 1)
	assert.Len(t, suggestion.SingleRoleMatches[0].CoveredPermissions, 1)
	assert.Empty(t, suggestion.UncoveredPermissions)
}
