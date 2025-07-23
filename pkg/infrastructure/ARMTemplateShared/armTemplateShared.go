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

package ARMTemplateShared

import "errors"

var ErrInvalidTemplate = errors.New("InvalidTemplate")

type ArmTemplateAdditionalConfig struct {
	TemplateFilePath   string
	ParametersFilePath string
	DeploymentName     string
	Location           string
	SubscriptionScoped bool
}

// Get parameters in standard format that is without the schema, contentVersion and parameters fields
func GetParametersInStandardFormat(parameters map[string]interface{}) map[string]interface{} {
	// convert from
	// {
	// 	"$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentParameters.json#",
	// 	"contentVersion": "1.0.0.0",
	// 	"parameters": {
	// 	  "adminUsername": {
	// 		"value": "GEN-UNIQUE"
	// 	  },
	// 	  "adminPasswordOrKey": {
	// 		"value": "GEN-PASSWORD"
	// 	  },
	// 	  "dnsLabelPrefix": {
	// 		"value": "GEN-UNIQUE"
	// 	  }
	// 	}
	//   }

	// convert to
	// {
	// 	  "adminUsername": {
	// 		"value": "GEN-UNIQUE"
	// 	  },
	// 	  "adminPasswordOrKey": {
	// 		"value": "GEN-PASSWORD"
	// 	  },
	// 	  "dnsLabelPrefix": {
	// 		"value": "GEN-UNIQUE"
	// 	  }
	// 	}
	if parameters["$schema"] != nil {

		return parameters["parameters"].(map[string]interface{})

	}
	return parameters
}
