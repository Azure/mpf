package domain

import (
	log "github.com/sirupsen/logrus"
)

var toAppendSpecialCasePermissions = map[string][]string{

	// Below permissions are required for Microsoft.Insights/components due to the following issue:
	// https://github.com/hashicorp/terraform-provider-azurerm/issues/27961#issuecomment-2467392936

	"Microsoft.Insights/components/read": {
		"Microsoft.Insights/components/currentbillingfeatures/read",
		"Microsoft.AlertsManagement/smartDetectorAlertRules/read",
	},
	"Microsoft.Insights/components/write": {
		"Microsoft.Insights/components/currentbillingfeatures/write",
		"Microsoft.AlertsManagement/smartDetectorAlertRules/write",
	},
}

func appendPermissionsForSpecialCases(scpPerms map[string][]string) map[string][]string {
	for scp, perms := range scpPerms {
		for _, perm := range perms {
			if toAppend, ok := toAppendSpecialCasePermissions[perm]; ok {
				scpPerms[scp] = append(scpPerms[scp], toAppend...)
				log.Infof("Appended special case permissions for scope %s: %v", scp, toAppend)
			}
		}
	}
	return scpPerms
}
