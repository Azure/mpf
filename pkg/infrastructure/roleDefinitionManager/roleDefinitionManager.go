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

package roledefinitionmanager

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v3"
	"github.com/Azure/mpf/pkg/domain"
	"github.com/Azure/mpf/pkg/infrastructure/azureAPI"

	log "github.com/sirupsen/logrus"
)

// RoleDefinitionManager retrieves Azure built-in role definitions so that MPF can
// suggest which built-in role(s) cover the required permissions.
type RoleDefinitionManager struct {
	azAPIClient *azureAPI.AzureAPIClients
}

func NewRoleDefinitionManager(subscriptionID string) *RoleDefinitionManager {
	azAPIClient := azureAPI.NewAzureAPIClients(subscriptionID)
	return &RoleDefinitionManager{
		azAPIClient: azAPIClient,
	}
}

// GetBuiltInRoles lists all built-in role definitions at the subscription scope
// and returns them with their control-plane actions and notActions.
func (r *RoleDefinitionManager) GetBuiltInRoles(ctx context.Context, subscriptionID string) ([]domain.BuiltInRole, error) {
	scope := fmt.Sprintf("/subscriptions/%s", subscriptionID)

	pager := r.azAPIClient.RoleDefinitionsClient.NewListPager(scope, &armauthorization.RoleDefinitionsClientListOptions{
		Filter: to("type eq 'BuiltInRole'"),
	})

	var roles []domain.BuiltInRole
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list built-in role definitions: %w", err)
		}

		for _, roleDef := range page.Value {
			if roleDef == nil || roleDef.Properties == nil {
				continue
			}

			role := domain.BuiltInRole{
				RoleName:         derefString(roleDef.Properties.RoleName),
				RoleDefinitionID: derefString(roleDef.Name),
			}

			for _, perm := range roleDef.Properties.Permissions {
				if perm == nil {
					continue
				}
				role.Actions = append(role.Actions, derefStringSlice(perm.Actions)...)
				role.NotActions = append(role.NotActions, derefStringSlice(perm.NotActions)...)
			}

			roles = append(roles, role)
		}
	}

	log.Debugf("Retrieved %d built-in role definitions", len(roles))
	return roles, nil
}

func to(s string) *string {
	return &s
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefStringSlice(s []*string) []string {
	result := make([]string, 0, len(s))
	for _, item := range s {
		if item != nil {
			result = append(result, *item)
		}
	}
	return result
}
