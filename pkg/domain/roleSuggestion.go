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
	"regexp"
	"sort"
	"strings"
	"sync"
)

// BuiltInRole represents an Azure built-in role definition and the control-plane
// actions (and notActions) it grants. It is used to suggest which built-in
// role(s) cover the minimum permissions discovered by MPF.
type BuiltInRole struct {
	// RoleName is the human readable role name, e.g. "Storage Account Contributor".
	RoleName string
	// RoleDefinitionID is the role definition GUID.
	RoleDefinitionID string
	// Actions are the control-plane actions granted by the role.
	Actions []string
	// NotActions are the control-plane actions explicitly excluded from the role.
	NotActions []string
}

// SuggestedRole is a built-in role paired with the subset of the required
// permissions that it covers.
type SuggestedRole struct {
	Role BuiltInRole
	// CoveredPermissions are the required permissions covered by this role, sorted.
	CoveredPermissions []string
}

// RoleSuggestion is the result of matching a set of required permissions against
// the available Azure built-in roles.
type RoleSuggestion struct {
	// SingleRoleMatches are built-in roles that individually cover every required
	// permission. They are ordered from most specific (fewest granted actions) to
	// least specific, so least-privilege options appear first.
	SingleRoleMatches []SuggestedRole
	// MinimalCombination is a small set of built-in roles that together cover as
	// many of the required permissions as possible, computed with a greedy
	// set-cover heuristic. When a single role covers everything it contains that
	// one role.
	MinimalCombination []SuggestedRole
	// UncoveredPermissions are required permissions not covered by any built-in
	// role. When non-empty, a custom role is required for full coverage.
	UncoveredPermissions []string
}

// SuggestBuiltInRoles matches the given required permissions against the provided
// built-in roles and returns single-role matches, a minimal combination, and any
// permissions that no built-in role covers.
func SuggestBuiltInRoles(requiredPermissions []string, builtInRoles []BuiltInRole) RoleSuggestion {
	required := getUniqueSlice(normalizePermissions(requiredPermissions))
	sort.Strings(required)

	suggestion := RoleSuggestion{}

	if len(required) == 0 {
		return suggestion
	}

	// Precompute the set of required permissions each role covers.
	roleCoverage := make(map[int][]string, len(builtInRoles))
	for i, role := range builtInRoles {
		var covered []string
		for _, perm := range required {
			if roleCoversPermission(role, perm) {
				covered = append(covered, perm)
			}
		}
		if len(covered) > 0 {
			sort.Strings(covered)
			roleCoverage[i] = covered
		}
	}

	// Single-role matches: roles that cover every required permission.
	for i, covered := range roleCoverage {
		if len(covered) == len(required) {
			suggestion.SingleRoleMatches = append(suggestion.SingleRoleMatches, SuggestedRole{
				Role:               builtInRoles[i],
				CoveredPermissions: covered,
			})
		}
	}
	sortSuggestedRolesBySpecificity(suggestion.SingleRoleMatches)

	suggestion.MinimalCombination, suggestion.UncoveredPermissions = greedyRoleCombination(required, builtInRoles, roleCoverage)

	return suggestion
}

// greedyRoleCombination computes a small set of roles that together cover as many
// required permissions as possible. At each step it selects the role covering the
// most currently-uncovered permissions, breaking ties in favor of the more
// specific role (fewest total granted actions) and then by name for determinism.
func greedyRoleCombination(required []string, builtInRoles []BuiltInRole, roleCoverage map[int][]string) ([]SuggestedRole, []string) {
	remaining := make(map[string]bool, len(required))
	for _, perm := range required {
		remaining[perm] = true
	}

	var combination []SuggestedRole
	usedRoles := make(map[int]bool)

	for len(remaining) > 0 {
		bestIdx := -1
		var bestNewlyCovered []string

		for i := range builtInRoles {
			if usedRoles[i] {
				continue
			}
			covered, ok := roleCoverage[i]
			if !ok {
				continue
			}

			var newlyCovered []string
			for _, perm := range covered {
				if remaining[perm] {
					newlyCovered = append(newlyCovered, perm)
				}
			}
			if len(newlyCovered) == 0 {
				continue
			}

			if bestIdx == -1 || isBetterGreedyChoice(builtInRoles[i], len(newlyCovered), builtInRoles[bestIdx], len(bestNewlyCovered)) {
				bestIdx = i
				bestNewlyCovered = newlyCovered
			}
		}

		if bestIdx == -1 {
			// No remaining role covers any of the leftover permissions.
			break
		}

		sort.Strings(bestNewlyCovered)
		combination = append(combination, SuggestedRole{
			Role:               builtInRoles[bestIdx],
			CoveredPermissions: bestNewlyCovered,
		})
		usedRoles[bestIdx] = true
		for _, perm := range bestNewlyCovered {
			delete(remaining, perm)
		}
	}

	uncovered := make([]string, 0, len(remaining))
	for perm := range remaining {
		uncovered = append(uncovered, perm)
	}
	sort.Strings(uncovered)

	return combination, uncovered
}

// isBetterGreedyChoice reports whether candidate is a better greedy pick than the
// current best. More newly-covered permissions wins; ties prefer the more
// specific (least-privilege) role by breadth score, then the lexicographically
// smaller name.
func isBetterGreedyChoice(candidate BuiltInRole, candidateNew int, best BuiltInRole, bestNew int) bool {
	if candidateNew != bestNew {
		return candidateNew > bestNew
	}
	candidateBreadth := roleBreadthScore(candidate)
	bestBreadth := roleBreadthScore(best)
	if candidateBreadth != bestBreadth {
		return candidateBreadth < bestBreadth
	}
	return candidate.RoleName < best.RoleName
}

// sortSuggestedRolesBySpecificity orders roles from most specific (narrowest
// breadth score) to least specific, so least-privilege suggestions appear first
// and broad roles such as Contributor and Owner appear last.
func sortSuggestedRolesBySpecificity(roles []SuggestedRole) {
	sort.SliceStable(roles, func(i, j int) bool {
		bi := roleBreadthScore(roles[i].Role)
		bj := roleBreadthScore(roles[j].Role)
		if bi != bj {
			return bi < bj
		}
		return roles[i].Role.RoleName < roles[j].Role.RoleName
	})
}

const (
	// globalWildcardWeight is the breadth contribution of the "*" action, which
	// grants every control-plane operation (as in Owner/Contributor).
	globalWildcardWeight int64 = 1_000_000
	// wildcardWeight is the breadth contribution of a scoped wildcard action such
	// as "Microsoft.Storage/*" or "*/read".
	wildcardWeight int64 = 1_000
	// exactActionWeight is the breadth contribution of a single, fully qualified
	// action with no wildcards.
	exactActionWeight int64 = 1
)

// roleBreadthScore estimates how broad a role's granted permissions are. Lower
// scores indicate more specific (least-privilege) roles. It intentionally ranks
// a role that grants "*" (a single but all-encompassing action) as far broader
// than a role that lists many narrowly scoped actions.
func roleBreadthScore(role BuiltInRole) int64 {
	var score int64
	for _, action := range role.Actions {
		score += actionBreadthWeight(action)
	}
	if score == 0 {
		// A role with no usable actions should not be treated as the most specific.
		return globalWildcardWeight
	}
	return score
}

func actionBreadthWeight(action string) int64 {
	action = normalizePermission(action)
	switch {
	case action == "":
		return 0
	case action == "*":
		return globalWildcardWeight
	case strings.Contains(action, "*"):
		return wildcardWeight
	default:
		return exactActionWeight
	}
}

// roleCoversPermission reports whether the role grants the given permission: the
// permission must match at least one Action pattern and must not match any
// NotAction pattern.
func roleCoversPermission(role BuiltInRole, permission string) bool {
	permission = normalizePermission(permission)
	if permission == "" {
		return false
	}

	matched := false
	for _, action := range role.Actions {
		if actionMatchesPattern(normalizePermission(action), permission) {
			matched = true
			break
		}
	}
	if !matched {
		return false
	}

	for _, notAction := range role.NotActions {
		if actionMatchesPattern(normalizePermission(notAction), permission) {
			return false
		}
	}
	return true
}

// compiledPatternCache memoises the regexp built for each wildcard action pattern.
// SuggestBuiltInRoles matches every role action against every required permission,
// so the same handful of patterns would otherwise be recompiled many times.
var compiledPatternCache sync.Map

// actionMatchesPattern reports whether an Azure action pattern (which may contain
// '*' wildcards matching any sequence of characters) matches the given action.
// Matching is case-insensitive, consistent with Azure RBAC evaluation.
func actionMatchesPattern(pattern string, action string) bool {
	if pattern == "" {
		return false
	}
	if !strings.Contains(pattern, "*") {
		return strings.EqualFold(pattern, action)
	}
	if pattern == "*" {
		return true
	}
	// "Microsoft.Storage/*" style prefixes are the most common wildcard form.
	if strings.Count(pattern, "*") == 1 && strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return len(action) >= len(prefix) && strings.EqualFold(action[:len(prefix)], prefix)
	}

	re, err := compilePattern(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(action)
}

func compilePattern(pattern string) (*regexp.Regexp, error) {
	if cached, ok := compiledPatternCache.Load(pattern); ok {
		return cached.(*regexp.Regexp), nil
	}

	var sb strings.Builder
	sb.WriteString("(?i)^")
	for _, segment := range strings.Split(pattern, "*") {
		sb.WriteString(regexp.QuoteMeta(segment))
		sb.WriteString(".*")
	}
	// Remove the trailing ".*" added after the final segment and anchor the end.
	regexStr := strings.TrimSuffix(sb.String(), ".*") + "$"

	re, err := regexp.Compile(regexStr)
	if err != nil {
		return nil, err
	}
	cached, _ := compiledPatternCache.LoadOrStore(pattern, re)
	return cached.(*regexp.Regexp), nil
}

func normalizePermissions(permissions []string) []string {
	normalized := make([]string, 0, len(permissions))
	for _, perm := range permissions {
		perm = normalizePermission(perm)
		if perm != "" {
			normalized = append(normalized, perm)
		}
	}
	return normalized
}

func normalizePermission(permission string) string {
	return strings.TrimSpace(permission)
}
