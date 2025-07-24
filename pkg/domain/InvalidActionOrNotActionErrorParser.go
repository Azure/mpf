package domain

import (
	"errors"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

func GetDeleteActionFromInvalidActionOrNotActionError(invalidActionErrMesg string) ([]string, error) {

	// parse error message to get the delete action
	// {"error":{"code":"InvalidActionOrNotAction","message":"'Microsoft.Insights/components/currentbillingfeatures/delete' does not match any of the actions supported by the providers."}}

	invalidActions := []string{}
	if invalidActionErrMesg != "" && !strings.Contains(invalidActionErrMesg, "InvalidActionOrNotAction") {
		log.Infoln("Non InvalidActionOrNotAction Error when creating deployment:", invalidActionErrMesg)
		return invalidActions, errors.New("Could not parse deploment error, potentially due to a Non-InvalidActionOrNotAction error")
	}

	re := regexp.MustCompile(`"message":"\'([^']+delete)\' does not match any of the actions supported by the providers."`)
	matches := re.FindAllStringSubmatch(invalidActionErrMesg, -1)

	if len(matches) == 0 {
		return invalidActions, errors.New("No matches found in 'invalidActionErrorMessage' error message")
	}

	for _, match := range matches {
		if len(match) == 2 {
			invalidActions = append(invalidActions, match[1])
		}
	}

	return invalidActions, nil

}

func GetInvalidActionFromInvalidActionOrNotActionError(invalidActionErrMesg string) ([]string, error) {
	// parse error message to get the action
	// {"error":{"code":"InvalidActionOrNotAction","message":"'Microsoft.Insights/components/currentbillingfeatures/delete' does not match any of the actions supported by the providers."}}
	// {"error":{"code":"InvalidActionOrNotAction","message":"'Microsoft.OperationalInsights/workspaces/PrivateEndpointConnectionsApproval/action' does not match any of the actions supported by the providers."}}
	log.Debugf("Attempting to Parse InvalidActionOrNotAction Error: %s", invalidActionErrMesg)
	invalidActions := []string{}
	if invalidActionErrMesg != "" && !strings.Contains(invalidActionErrMesg, "InvalidActionOrNotAction") {
		log.Infoln("Non InvalidActionOrNotAction Error when creating deployment:", invalidActionErrMesg)
		return invalidActions, errors.New("could not parse deploment error, potentially due to a Non-InvalidActionOrNotAction error")
	}

	re := regexp.MustCompile(`"message":"\'([^']+)\' does not match any of the actions supported by the providers."`)
	matches := re.FindAllStringSubmatch(invalidActionErrMesg, -1)

	if len(matches) == 0 {
		return invalidActions, errors.New("no matches found in 'invalidActionErrorMessage' error message")
	}

	for _, match := range matches {
		if len(match) == 2 {
			invalidActions = append(invalidActions, match[1])
		}
	}

	return invalidActions, nil
}
