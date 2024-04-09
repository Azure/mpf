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

package usecase

import (
	"context"

	"github.com/manisbindra/az-mpf/pkg/domain"
)

type ServicePrincipalAssignmentModifier interface {
	DetachRolesFromSP(ctx context.Context, subscription string, SPOBjectID string, role domain.Role) error
	AssignRoleToSP(subscription string, SPOBjectID string, role domain.Role) error
}

type CustomRoleCreatorModifier interface {
	CreateUpdateCustomRole(subscription string, role domain.Role, permissions []string) error
	DeleteCustomRole(subscription string, role domain.Role) error
}

type ServicePrincipalRolemAssignmentManager interface {
	ServicePrincipalAssignmentModifier
	CustomRoleCreatorModifier
}
