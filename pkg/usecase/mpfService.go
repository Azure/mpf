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
	"strings"
	"time"

	"github.com/Azure/mpf/pkg/domain"
	log "github.com/sirupsen/logrus"
)

// RetryDeploymentResponseErrorMessage is the error message returned by a deployment authorization checker when it wants the deployment to be retried
const RetryDeploymentResponseErrorMessage = "RetryGetDeploymentAuthorizationErrors"

type MPFService struct {
	ctx                                 context.Context
	rgManager                           ResourceGroupManager
	spRoleAssignmentManager             ServicePrincipalRolemAssignmentManager
	deploymentAuthCheckerCleaner        DeploymentAuthorizationCheckerCleaner
	mpfConfig                           domain.MPFConfig
	initialPermissionsToAdd             []string
	permissionsToAddToResult            []string
	requiredPermissions                 map[string][]string
	autoAddReadPermissionForEachWrite   bool
	autoAddDeletePermissionForEachWrite bool
	autoCreateResourceGroup             bool
	iterationCount                      int
}

func NewMPFService(ctx context.Context, rgMgr ResourceGroupManager, spRoleAssgnMgr ServicePrincipalRolemAssignmentManager, deploymentAuthChkCln DeploymentAuthorizationCheckerCleaner, mpfConfig domain.MPFConfig, initialPermissionsToAdd []string, permissionsToAddToResult []string, autoAddReadPermissionForEachWrite bool, autoAddDeletePermissionForEachWrite bool, autoCreateResourceGroup bool) *MPFService {
	return &MPFService{
		ctx:                                 ctx,
		rgManager:                           rgMgr,
		spRoleAssignmentManager:             spRoleAssgnMgr,
		deploymentAuthCheckerCleaner:        deploymentAuthChkCln,
		mpfConfig:                           mpfConfig,
		initialPermissionsToAdd:             initialPermissionsToAdd,
		permissionsToAddToResult:            permissionsToAddToResult,
		requiredPermissions:                 make(map[string][]string),
		autoAddReadPermissionForEachWrite:   autoAddReadPermissionForEachWrite,
		autoAddDeletePermissionForEachWrite: autoAddDeletePermissionForEachWrite,
		autoCreateResourceGroup:             autoCreateResourceGroup,
	}
}

func (s *MPFService) returnMPFResult(err error) (domain.MPFResult, error) {
	mpfResult := domain.GetMPFResultWithIterationCount(s.requiredPermissions, s.iterationCount)

	if err != nil && len(mpfResult.RequiredPermissions) == 0 {
		return domain.MPFResult{}, err
	}

	if err != nil && len(mpfResult.RequiredPermissions) > 0 {
		return mpfResult, err
	}

	return mpfResult, nil
}

func (s *MPFService) GetMinimumPermissionsRequired() (domain.MPFResult, error) {

	if s.autoCreateResourceGroup {
		// Create Resource Group
		log.Infof("Creating Resource Group: %s \n", s.mpfConfig.ResourceGroup.ResourceGroupName)
		err := s.rgManager.CreateResourceGroup(s.ctx, s.mpfConfig.ResourceGroup.ResourceGroupName, s.mpfConfig.ResourceGroup.Location)
		if err != nil {
			// Avoid terminating the entire process (log.Fatal calls os.Exit).
			// Bubble the error up so callers/tests can handle it.
			log.Warnf("failed to create resource group %q: %v", s.mpfConfig.ResourceGroup.ResourceGroupName, err)
			return s.returnMPFResult(err)
		}
		log.Infof("Resource Group: %s created successfully \n", s.mpfConfig.ResourceGroup.ResourceGroupName)
		// defer s.deploymentAuthCheckerCleaner.CleanDeployment(s.mpfConfig)
	}

	defer s.CleanUpResources()

	// Delete all existing role assignments for the service principal
	// Pass empty role to delete ALL role assignments (not just the specific custom role)
	err := s.spRoleAssignmentManager.DetachRolesFromSP(s.ctx, s.mpfConfig.SubscriptionID, s.mpfConfig.SP.SPObjectID, domain.Role{})
	if err != nil {
		log.Warnf("Unable to delete Role Assignments: %v\n", err)
		return s.returnMPFResult(err)
	}
	log.Info("Deleted all existing role assignments for service principal \n")

	// Wait for Azure RBAC propagation after deleting role assignments
	// This ensures that any previous permissions are fully revoked before starting the new test
	log.Infoln("Waiting for Azure RBAC propagation after deleting role assignments...")
	time.Sleep(90 * time.Second)

	// Initialize new custom role
	log.Infoln("Initializing Custom Role")
	// err = mpf.CreateUpdateCustomRole([]string{})

	err, invalidActions := s.spRoleAssignmentManager.CreateUpdateCustomRole(s.mpfConfig.SubscriptionID, s.mpfConfig.Role, s.initialPermissionsToAdd)
	if err != nil {
		log.Warn(err)
		return s.returnMPFResult(err)
	}
	if len(invalidActions) > 0 {
		log.Warnf("The following invalid actions were removed from the role: %v", invalidActions)
	}
	log.Infoln("Custom role initialized successfully")

	// Assign new custom role to service principal
	log.Infoln("Assigning new custom role to service principal")
	// err = mpf.AssignRoleToSP()
	err = s.spRoleAssignmentManager.AssignRoleToSP(s.mpfConfig.SubscriptionID, s.mpfConfig.SP.SPObjectID, s.mpfConfig.Role)
	if err != nil {
		log.Warn(err)
		return s.returnMPFResult(err)
	}
	log.Infoln("New Custom Role assigned to service principal successfully")

	// Wait for Azure RBAC propagation after initial role assignment
	// Azure role assignments can take a few seconds to propagate across all authorization endpoints
	log.Infoln("Waiting for Azure RBAC propagation after initial role assignment...")
	time.Sleep(5 * time.Second)

	// Add initial permissions to requiredPermissions map
	log.Infoln("Adding initial permissions to requiredPermissions map")
	s.requiredPermissions[s.mpfConfig.SubscriptionID] = append(s.requiredPermissions[s.mpfConfig.SubscriptionID], s.permissionsToAddToResult...)

	maxIterations := 50
	for {
		authErrMesg, err := s.deploymentAuthCheckerCleaner.GetDeploymentAuthorizationErrors(s.mpfConfig)

		log.Infof("Iteration Number: %d \n", s.iterationCount)

		if authErrMesg == "" && err == nil {
			log.Infoln("Authorization Successful")
			break
		}

		log.Debugln("authErrMesg: ", authErrMesg)

		if err == nil && strings.Contains(authErrMesg, RetryDeploymentResponseErrorMessage) {
			log.Warnf("received retry request from authorization checker, retrying deployment.... \n")
			continue
		}

		if err != nil {
			log.Warnf("Non Authorization error received: %v \n", err)
			return s.returnMPFResult(err)
		}

		log.Debugln("Deployment Authorization Error:", authErrMesg)

		scpMp, err := domain.GetScopePermissionsFromAuthError(authErrMesg)
		if err != nil {
			log.Warnf("Could Not Parse Deployment Authorization Error: %v \n", err)
			return s.returnMPFResult(err)
		}

		log.Infoln("Successfully Parsed Deployment Authorization Error")
		log.Debugln("scope permissions found from deployment error:", scpMp)

		// auto add read and delete permissions as per configuration
		for scope, permissions := range scpMp {
			for _, permission := range permissions {
				if s.autoAddReadPermissionForEachWrite && strings.HasSuffix(permission, "/write") {
					readPermission := strings.Replace(permission, "/write", "/read", 1)
					scpMp[scope] = append(scpMp[scope], readPermission)
				}
				if s.autoAddDeletePermissionForEachWrite && strings.HasSuffix(permission, "/write") {
					deletePermission := strings.Replace(permission, "/write", "/delete", 1)
					scpMp[scope] = append(scpMp[scope], deletePermission)
				}
			}
		}

		log.Infoln("Adding mising scopes/permissions to final result map...")
		for k, v := range scpMp {
			s.requiredPermissions[k] = append(s.requiredPermissions[k], v...)
			s.requiredPermissions[s.mpfConfig.SubscriptionID] = append(s.requiredPermissions[s.mpfConfig.SubscriptionID], v...)
		}

		// assign permission to role
		log.Infoln("Adding permission/scope to role...........")
		log.Debugln("Number of Permissions added to role:", len(s.requiredPermissions[s.mpfConfig.SubscriptionID]))

		permissionsIncludingInitialPermissions := append(s.initialPermissionsToAdd, s.requiredPermissions[s.mpfConfig.SubscriptionID]...)
		err, invalidActions := s.spRoleAssignmentManager.CreateUpdateCustomRole(s.mpfConfig.SubscriptionID, s.mpfConfig.Role, permissionsIncludingInitialPermissions)

		// err = s.spRoleAssignmentManager.CreateUpdateCustomRole(s.mpfConfig.SubscriptionID, s.mpfConfig.ResourceGroup.ResourceGroupName, s.mpfConfig.Role, s.requiredPermissions[s.mpfConfig.ResourceGroup.ResourceGroupResourceID])

		if err != nil {
			log.Infoln("Error when adding permission/scope to role: \n", err)
			log.Warn(err)
			return s.returnMPFResult(err)
		}
		if len(invalidActions) > 0 {
			log.Warnf("The following invalid actions were removed from the role during iteration: %v", invalidActions)
		}
		log.Infoln("Permission/scope added to role successfully")

		// Wait for Azure RBAC propagation before retrying deployment
		// Azure role definition updates can take a few seconds to propagate across all authorization endpoints
		log.Infoln("Waiting for Azure RBAC propagation...")
		time.Sleep(5 * time.Second)

		s.iterationCount++
		if s.iterationCount == maxIterations {
			log.Warnln("max iterations for fetching authorization errors reached, exiting...")
			return s.returnMPFResult(err)
		}
	}

	return s.returnMPFResult(nil)

}

func (s *MPFService) CleanUpResources() {
	log.Infoln("Cleaning up resources...")
	log.Infoln("*************************")

	// Cancel deployment. Even if cancelling deployment fails attempt to delete other resources
	// _ = m.CancelDeployment(deploymentName)

	err := s.deploymentAuthCheckerCleaner.CleanDeployment(s.mpfConfig)
	if err != nil {
		log.Warnln("Cleaning up deployment returned an error, attempting to clean rest of the resources")
	}

	// Detach Roles from SP
	err = s.spRoleAssignmentManager.DetachRolesFromSP(s.ctx, s.mpfConfig.SubscriptionID, s.mpfConfig.SP.SPObjectID, s.mpfConfig.Role)
	if err != nil {
		log.Warnf("Could not detach roles from SP: %s\n", err)
	}

	// Delete Custom Role
	err = s.spRoleAssignmentManager.DeleteCustomRole(s.mpfConfig.SubscriptionID, s.mpfConfig.Role)
	if err != nil {
		log.Warnf("Could not delete custom role: %s\n", err)
	}

	// Delete Resource Group
	if s.autoCreateResourceGroup {
		err = s.rgManager.DeleteResourceGroup(s.ctx, s.mpfConfig.ResourceGroup.ResourceGroupName)
		if err != nil {
			log.Warnf("Error when deleting resource group: %s \n", err)
		}
		log.Infoln("Resource group deletion initiated successfully...")
	}

}
