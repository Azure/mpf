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

package terraform

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Azure/mpf/pkg/domain"
	"github.com/hashicorp/terraform-exec/tfexec"
	log "github.com/sirupsen/logrus"
)

type terraformDeploymentConfig struct {
	ctx                            context.Context
	workingDir                     string
	execPath                       string
	varFilePath                    string
	importExistingResourcesToState bool
	targetModule                   string
	consecutiveRetryCount          int
	sleep                          func(time.Duration)
}

var inDestroyPhase bool

// This file is created once the destroy phase is entered
const (
	TFDestroyStateEnteredFileName = ".azmpfEnteredDestroyPhase.txt"
	TFExistingResourceErrorMsg    = "to be managed via Terraform this resource needs to be imported into the State"

	// Returning this response will trigger a retry
	RetryDeploymentResponseErrorMessage = "RetryGetDeploymentAuthorizationErrors"

	// Retryable errors
	BillingFeaturesPayloadError = "CurrentBillingFeatures is required in payload"
	WaitingForDataplaneError    = "waiting for the Data Plane"
)

var terraformRetryDelays = [...]time.Duration{
	10 * time.Second,
	20 * time.Second,
	30 * time.Second,
	30 * time.Second,
}

// These narrow contexts match the Storage reads observed with AzureRM 4.80.0.
var transientStorageNotFoundContexts = []string{
	"retrieving storage account",
	"retrieving share properties for storage account",
	"retrieving file properties for storage account",
	"retrieving queue properties for storage account",
	"retrieving blob properties for storage account",
	"retrieving static website properties for storage account",
	"waiting for the data plane for storage account",
	"waiting for the blob service to become available",
	"waiting for the file service to become available",
	"waiting for the queue service to become available",
	"waiting for the static website to become available",
}

func NewTerraformAuthorizationChecker(workDir string, execPath string, varFilePath string, importExistingResources bool, targetModule string) *terraformDeploymentConfig {
	err := deleteEnteredDestroyPhaseStateFile(workDir, TFDestroyStateEnteredFileName)
	if err != nil {
		log.Warnf("error deleting enteredDestroyPhaseStateFile: %s", err)
	}

	return &terraformDeploymentConfig{
		workingDir:                     workDir,
		execPath:                       filepath.Clean(execPath),
		ctx:                            context.Background(),
		varFilePath:                    varFilePath,
		importExistingResourcesToState: importExistingResources,
		targetModule:                   targetModule,
		sleep:                          time.Sleep,
	}
}

func (a *terraformDeploymentConfig) GetDeploymentAuthorizationErrors(mpfConfig domain.MPFConfig) (string, error) {
	return a.deployTerraform(mpfConfig)
}

func (a *terraformDeploymentConfig) CleanDeployment(mpfConfig domain.MPFConfig) error {

	err := deleteEnteredDestroyPhaseStateFile(a.workingDir, TFDestroyStateEnteredFileName)
	if err != nil {
		log.Warnf("error deleting enteredDestroyPhaseStateFile: %s", err)
	}

	tf, err := tfexec.NewTerraform(a.workingDir, a.execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}

	err = tf.Init(context.Background())
	if err != nil {
		log.Warnf("error running Init: %s", err)
		return err
	}

	switch {
	case a.varFilePath == "" && a.targetModule == "":
		err = tf.Destroy(a.ctx)
	case a.varFilePath != "" && a.targetModule == "":
		err = tf.Destroy(a.ctx, tfexec.VarFile(a.varFilePath))
	case a.varFilePath == "" && a.targetModule != "":
		err = tf.Destroy(a.ctx, tfexec.Target(a.targetModule))
	case a.varFilePath != "" && a.targetModule != "":
		err = tf.Destroy(a.ctx, tfexec.VarFile(a.varFilePath), tfexec.Target(a.targetModule))
	}
	if err != nil {
		log.Warnf("error running terraform destroy: %s", err)
	}
	return err
}

func (a *terraformDeploymentConfig) setTFConfig(mpfConfig domain.MPFConfig) (*tfexec.Terraform, error) {
	log.Infof("workingDir: %s", a.workingDir)
	log.Infof("varfilePath: %s", a.varFilePath)
	log.Infof("execPath: %s", a.execPath)

	tf, err := tfexec.NewTerraform(a.workingDir, a.execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}

	pathEnvVal := os.Getenv("PATH")
	var tfLogLevel string

	tfLogPathEnvVal := os.Getenv("TF_LOG_PATH")
	if tfLogPathEnvVal == "" {
		tfLogPathEnvVal = a.workingDir + "/terraform.log"
	}

	tfReattachProviders := os.Getenv("TF_REATTACH_PROVIDERS")

	switch log.GetLevel() {
	case log.InfoLevel:
		tfLogLevel = "INFO"
	case log.WarnLevel:
		tfLogLevel = "WARN"
	case log.DebugLevel:
		tfLogLevel = "DEBUG"
	case log.TraceLevel:
		tfLogLevel = "TRACE"
	default:
		tfLogLevel = "ERROR"
	}

	if tfLogLevel != "ERROR" {
		err := tf.SetLog(tfLogLevel)
		if err != nil {
			log.Warnf("error setting Terraform log level: %s", err)
		}
		err = tf.SetLogPath(tfLogPathEnvVal)
		if err != nil {
			log.Warnf("error setting Terraform log path: %s", err)
		}
		tf.SetStderr(os.Stderr)
		tf.SetStdout(os.Stdout)
	}

	envVars := map[string]string{
		"ARM_CLIENT_ID":       mpfConfig.SP.SPClientID,
		"ARM_CLIENT_SECRET":   mpfConfig.SP.SPClientSecret,
		"ARM_SUBSCRIPTION_ID": mpfConfig.SubscriptionID,
		"ARM_TENANT_ID":       mpfConfig.TenantID,
		"PATH":                pathEnvVal,
	}

	// Pass through temp directory env vars to avoid Terraform using system directories
	for _, k := range []string{"TMP", "TEMP", "TMPDIR", "USERPROFILE", "HOME", "HOMEDRIVE", "HOMEPATH"} {
		if v := os.Getenv(k); v != "" {
			envVars[k] = v
		}
	}

	if tfReattachProviders != "" {
		envVars["TF_REATTACH_PROVIDERS"] = tfReattachProviders
	}

	err = tf.SetEnv(envVars)
	if err != nil {
		log.Warnf("error setting Terraform env vars: %s", err)
	}
	return tf, nil
}

func (a *terraformDeploymentConfig) deployTerraform(mpfConfig domain.MPFConfig) (string, error) {
	tf, err := a.setTFConfig(mpfConfig)
	if err != nil {
		log.Fatalf("error setting Terraform start config: %s", err)
	}

	inDestroyPhase = doesEnteredDestroyPhaseStateFileExist(a.workingDir, TFDestroyStateEnteredFileName)
	if !inDestroyPhase {
		log.Infof("destroy phase file does not exist, in apply phase")
		msg, err := a.terraformApply(mpfConfig, tf)
		if err != nil || msg != "" {
			return msg, err
		}
	}

	return a.terraformDestroy(mpfConfig, tf)

}

func (a *terraformDeploymentConfig) terraformApply(mpfConfig domain.MPFConfig, tf *tfexec.Terraform) (string, error) {

	err := tf.Init(context.Background())
	if err != nil {
		log.Warnf("error running Init: %s", err)
		return "", err
	}

	log.Infoln("in apply phase")

	switch {
	case a.varFilePath == "" && a.targetModule == "":
		err = tf.Apply(a.ctx)
	case a.varFilePath != "" && a.targetModule == "":
		err = tf.Apply(a.ctx, tfexec.VarFile(a.varFilePath))
	case a.varFilePath == "" && a.targetModule != "":
		err = tf.Apply(a.ctx, tfexec.Target(a.targetModule))
	case a.varFilePath != "" && a.targetModule != "":
		err = tf.Apply(a.ctx, tfexec.VarFile(a.varFilePath), tfexec.Target(a.targetModule))
	}

	if err == nil {
		a.resetRetryCountAfterProgress()
		return "", nil
	}

	errorMsg := err.Error()
	log.Debugln("terraform apply error: ", errorMsg)

	// Temporary fix to workaround issue https://github.com/hashicorp/terraform-provider-azurerm/issues/27961
	// The checker bounds this retry as well, in case the provider error persists.
	if strings.Contains(errorMsg, BillingFeaturesPayloadError) {
		return a.retryOrReturnError(err)
	}

	// import errors can occur for some resources, when identity does not have all required permissions,
	// as described in https://github.com/hashicorp/terraform-provider-azurerm/issues/27961#issuecomment-2467392936
	if a.importExistingResourcesToState && strings.Contains(errorMsg, TFExistingResourceErrorMsg) {

		msg, err := a.terraformImport(tf, errorMsg)
		if err != nil || msg != "" {
			if strings.Contains(msg, "Authorization") {
				a.resetRetryCountForAuthorizationError(msg)
				return msg, nil
			}
			return msg, err
		}
		return a.terraformApply(mpfConfig, tf)
	}

	return a.handleTerraformApplyError(err)
}

func (a *terraformDeploymentConfig) handleTerraformApplyError(err error) (string, error) {
	errorMsg := err.Error()
	if strings.Contains(errorMsg, "Authorization") || strings.Contains(errorMsg, "LinkedAccessCheckFailed") {
		if hasParseableAuthorizationError(errorMsg) {
			a.resetRetryCountAfterProgress()
			log.Debug("terraform apply: parseable authorization error occurred")
			return errorMsg, nil
		}
		if strings.Contains(errorMsg, WaitingForDataplaneError) {
			log.Warnln("terraform apply: waiting for dataplane error occurred, requesting bounded retry")
			return a.retryOrReturnError(err)
		}
		if isTransientStorageNotFoundError(errorMsg) {
			log.Warnln("terraform apply: transient Storage ResourceNotFound occurred with unparseable authorization diagnostics, requesting bounded retry")
			return a.retryOrReturnError(err)
		}
		log.Debug("terraform apply: authorization error occured")
		return errorMsg, nil
	}

	if isTransientStorageNotFoundError(errorMsg) {
		log.Warnln("terraform apply: transient Storage ResourceNotFound error occurred, requesting bounded retry")
		return a.retryOrReturnError(err)
	}

	log.Warnf("terraform apply: non authorizaton error occured: %s", errorMsg)
	return errorMsg, err
}

func isTransientStorageNotFoundError(errorMsg string) bool {
	lowerErrorMsg := strings.ToLower(errorMsg)
	if !strings.Contains(lowerErrorMsg, "404") || !strings.Contains(lowerErrorMsg, "resourcenotfound") {
		return false
	}

	for _, context := range transientStorageNotFoundContexts {
		if strings.Contains(lowerErrorMsg, context) {
			return true
		}
	}
	return false
}

func (a *terraformDeploymentConfig) retryOrReturnError(err error) (string, error) {
	if a.consecutiveRetryCount >= len(terraformRetryDelays) {
		a.consecutiveRetryCount = 0
		return err.Error(), err
	}

	delay := terraformRetryDelays[a.consecutiveRetryCount]
	a.consecutiveRetryCount++
	sleep := a.sleep
	if sleep == nil {
		sleep = time.Sleep
	}
	sleep(delay)
	return RetryDeploymentResponseErrorMessage, nil
}

func (a *terraformDeploymentConfig) resetRetryCountForAuthorizationError(errorMsg string) {
	if hasParseableAuthorizationError(errorMsg) {
		a.resetRetryCountAfterProgress()
	}
}

func hasParseableAuthorizationError(errorMsg string) bool {
	_, err := domain.GetScopePermissionsFromAuthError(errorMsg)
	return err == nil
}

func (a *terraformDeploymentConfig) resetRetryCountAfterProgress() {
	a.consecutiveRetryCount = 0
}

func (a *terraformDeploymentConfig) terraformImport(tf *tfexec.Terraform, existingResErrMesg string) (string, error) {
	log.Warnf("terraform apply: existing resource error occured:|| %s ||\n\n", existingResErrMesg)
	log.Warn("importing existing resources to state")

	exstResAddrAndResIDs, err := GetAddressAndResourceIDFromExistingResourceError(existingResErrMesg)
	if err != nil {
		log.Warnf("error getting existing resource address and resource ID: %s \n", err)
		return existingResErrMesg, err
	}

	for addr, resID := range exstResAddrAndResIDs {
		log.Warnf("importing existing resource: %s, %s ||\n", addr, resID)
		if a.varFilePath != "" {
			err = tf.Import(a.ctx, addr, resID, tfexec.VarFile(a.varFilePath))
		} else {
			err = tf.Import(a.ctx, addr, resID)
		}

		if err != nil {
			log.Warnf("error importing existing resource: %s \n", err)
			return err.Error(), err
		}
		log.Warnf("imported existing resource: %s, %s \n", addr, resID)
	}
	log.Warnf("imported all existing resources to state, triggering deployTerrafom again \n")
	return "", nil
}

func (a *terraformDeploymentConfig) terraformDestroy(mpfConfig domain.MPFConfig, tf *tfexec.Terraform) (string, error) {
	var err error
	log.Infoln("in destroy phase")
	if !inDestroyPhase {
		err = createEnteredDestroyPhaseStateFile(a.workingDir, TFDestroyStateEnteredFileName)
		if err != nil {
			log.Warnf("error creating enteredDestroyPhaseStateFile: %s", err)
		}
	}

	switch {
	case a.varFilePath == "" && a.targetModule == "":
		err = tf.Destroy(a.ctx)
	case a.varFilePath != "" && a.targetModule == "":
		err = tf.Destroy(a.ctx, tfexec.VarFile(a.varFilePath))
	case a.varFilePath == "" && a.targetModule != "":
		err = tf.Destroy(a.ctx, tfexec.Target(a.targetModule))
	case a.varFilePath != "" && a.targetModule != "":
		err = tf.Destroy(a.ctx, tfexec.VarFile(a.varFilePath), tfexec.Target(a.targetModule))
	}

	if err != nil {
		errorMsg := err.Error()
		log.Debugln(errorMsg)
		if strings.Contains(errorMsg, "Authorization") {
			return errorMsg, nil
		}
		log.Warnf("terraform destroy: non authorizaton error occured: %s", errorMsg)
		return errorMsg, err
	}
	return "", nil
}
