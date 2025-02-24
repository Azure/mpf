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

package resourcegroupmanager

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	azureAPI "github.com/Azure/mpf/pkg/infrastructure/azureAPI"
)

type RGManager struct {
	rgAPIClient *armresources.ResourceGroupsClient
}

func NewResourceGroupManager(subscriptionID string) *RGManager {
	azAPIClient := azureAPI.NewAzureAPIClients(subscriptionID)
	return &RGManager{
		rgAPIClient: azAPIClient.ResourceGroupsClient,
	}
}

func (r *RGManager) DeleteResourceGroup(ctx context.Context, rgName string) error {

	_, err := r.rgAPIClient.BeginDelete(ctx, rgName, nil)
	if err != nil {
		return err
	}

	return nil
}

// method to create resource group
func (r *RGManager) CreateResourceGroup(ctx context.Context, rgName string, location string) error {

	rgParams := armresources.ResourceGroup{
		Location: &location,
		Name:     &rgName,
	}

	// create resource group
	_, err := r.rgAPIClient.CreateOrUpdate(ctx, rgName, rgParams, nil)
	if err != nil {
		return err
	}

	return nil
}
