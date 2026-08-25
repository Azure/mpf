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
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/Azure/mpf/pkg/domain"
	"github.com/stretchr/testify/assert"
)

func TestDisplayRoleSuggestion_Text_SingleMatch(t *testing.T) {
	suggestion := domain.RoleSuggestion{
		SingleRoleMatches: []domain.SuggestedRole{
			{
				Role:               domain.BuiltInRole{RoleName: "Storage Account Contributor", RoleDefinitionID: "17d1049b-9a84-46fb-8f53-869881c3d3ab"},
				CoveredPermissions: []string{"Microsoft.Storage/storageAccounts/read"},
			},
		},
	}

	var buf bytes.Buffer
	err := DisplayRoleSuggestion(&buf, suggestion, false)
	assert.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "Suggested Built-In Roles:")
	assert.Contains(t, out, "each cover ALL required permissions")
	assert.Contains(t, out, "Storage Account Contributor")
}

func TestDisplayRoleSuggestion_Text_CombinationAndUncovered(t *testing.T) {
	suggestion := domain.RoleSuggestion{
		MinimalCombination: []domain.SuggestedRole{
			{
				Role:               domain.BuiltInRole{RoleName: "Storage Reader", RoleDefinitionID: "id-1"},
				CoveredPermissions: []string{"Microsoft.Storage/storageAccounts/read"},
			},
		},
		UncoveredPermissions: []string{"Microsoft.CustomProvider/customResource/read"},
	}

	var buf bytes.Buffer
	err := DisplayRoleSuggestion(&buf, suggestion, false)
	assert.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "No single built-in role covers all required permissions.")
	assert.Contains(t, out, "Suggested minimal combination")
	assert.Contains(t, out, "Storage Reader")
	assert.Contains(t, out, "NOT covered by any built-in role")
	assert.Contains(t, out, "Microsoft.CustomProvider/customResource/read")
}

func TestDisplayRoleSuggestion_JSON(t *testing.T) {
	suggestion := domain.RoleSuggestion{
		SingleRoleMatches: []domain.SuggestedRole{
			{
				Role:               domain.BuiltInRole{RoleName: "Owner", RoleDefinitionID: "id-owner"},
				CoveredPermissions: []string{"Microsoft.Storage/storageAccounts/read"},
			},
		},
	}

	var buf bytes.Buffer
	err := DisplayRoleSuggestion(&buf, suggestion, true)
	assert.NoError(t, err)

	var decoded domain.RoleSuggestion
	err = json.Unmarshal(buf.Bytes(), &decoded)
	assert.NoError(t, err)
	assert.Len(t, decoded.SingleRoleMatches, 1)
	assert.Equal(t, "Owner", decoded.SingleRoleMatches[0].Role.RoleName)
}

func TestDisplayRoleSuggestion_Text_CapsSingleMatches(t *testing.T) {
	var matches []domain.SuggestedRole
	for i := 0; i < maxSingleRoleMatchesDisplayed+5; i++ {
		matches = append(matches, domain.SuggestedRole{
			Role: domain.BuiltInRole{RoleName: "Role", RoleDefinitionID: "id"},
		})
	}
	suggestion := domain.RoleSuggestion{SingleRoleMatches: matches}

	var buf bytes.Buffer
	err := DisplayRoleSuggestion(&buf, suggestion, false)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "and 5 more")
}

func TestDisplayCombinedJSON_IsSingleDocument(t *testing.T) {
	result := domain.MPFResult{
		RequiredPermissions: map[string][]string{
			"/subscriptions/sub-id": {"Microsoft.Storage/storageAccounts/read"},
		},
	}
	suggestion := domain.RoleSuggestion{
		SingleRoleMatches: []domain.SuggestedRole{
			{Role: domain.BuiltInRole{RoleName: "Owner", RoleDefinitionID: "id-owner"}},
		},
	}

	var buf bytes.Buffer
	err := DisplayCombinedJSON(&buf, result, suggestion)
	assert.NoError(t, err)

	decoder := json.NewDecoder(&buf)
	var decoded CombinedJSONResult
	assert.NoError(t, decoder.Decode(&decoded))
	// Anything after the first document would make stdout invalid JSON.
	assert.Equal(t, io.EOF, decoder.Decode(new(json.RawMessage)))

	assert.Equal(t, []string{"Microsoft.Storage/storageAccounts/read"}, decoded.RequiredPermissions["/subscriptions/sub-id"])
	assert.Len(t, decoded.RoleSuggestion.SingleRoleMatches, 1)
	assert.Equal(t, "Owner", decoded.RoleSuggestion.SingleRoleMatches[0].Role.RoleName)
}
