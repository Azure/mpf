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

type Role struct {
	RoleDefinitionID          string
	RoleDefinitionName        string
	RoleDefinitionDescription string
	RoleDefinitionResourceID  string
}

type ResourceGroup struct {
	ResourceGroupName       string
	ResourceGroupResourceID string
	Location                string
}

type ServicePrincipal struct {
	SPClientID     string
	SPObjectID     string
	SPClientSecret string
}

type MPFConfig struct {
	ResourceGroup  ResourceGroup
	SubscriptionID string
	TenantID       string
	SP             ServicePrincipal
	Role           Role
}

type MPFResult struct {
	// The map from which the minimum permissions will be calculated
	RequiredPermissions map[string][]string
	// IterationCount is the number of iterations MPF took to discover all permissions.
	// A value of 0 means all required permissions were provided upfront via initialPermissions.
	IterationCount int
}

func GetMPFResult(requiredPermissions map[string][]string) MPFResult {
	return MPFResult{
		RequiredPermissions: getMapWithUniqueValues(requiredPermissions),
	}
}

func GetMPFResultWithIterationCount(requiredPermissions map[string][]string, iterationCount int) MPFResult {
	return MPFResult{
		RequiredPermissions: getMapWithUniqueValues(requiredPermissions),
		IterationCount:      iterationCount,
	}
}
