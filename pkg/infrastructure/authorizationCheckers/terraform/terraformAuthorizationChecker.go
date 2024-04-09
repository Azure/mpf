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
	"strings"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/manisbindra/az-mpf/pkg/domain"
	log "github.com/sirupsen/logrus"
)

type terraformDeploymentConfig struct {
	ctx         context.Context
	workingDir  string
	execPath    string
	varFilePath string
}

func NewTerraformAuthorizationChecker(workDir string, execPath string, varFilePath string) *terraformDeploymentConfig {
	return &terraformDeploymentConfig{
		workingDir:  workDir,
		execPath:    execPath,
		ctx:         context.Background(),
		varFilePath: varFilePath,
	}
}

func (a *terraformDeploymentConfig) GetDeploymentAuthorizationErrors(mpfConfig domain.MPFConfig) (string, error) {
	return a.deployTerraform(mpfConfig)
}

func (a *terraformDeploymentConfig) CleanDeployment(mpfConfig domain.MPFConfig) error {
	tf, err := tfexec.NewTerraform(a.workingDir, a.execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}

	err = tf.Init(context.Background())

	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}

	err = tf.Destroy(a.ctx, tfexec.VarFile(a.varFilePath))
	if err != nil {
		log.Warnf("error running terraform destroy: %s", err)
	}
	return err
}

func (a *terraformDeploymentConfig) deployTerraform(mpfConfig domain.MPFConfig) (string, error) {

	log.Infof("workingDir: %s", a.workingDir)
	log.Infof("varfilePath: %s", a.varFilePath)
	log.Infof("execPath: %s", a.execPath)

	tf, err := tfexec.NewTerraform(a.workingDir, a.execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}

	envVars := map[string]string{
		"ARM_CLIENT_ID":       mpfConfig.SP.SPClientID,
		"ARM_CLIENT_SECRET":   mpfConfig.SP.SPClientSecret,
		"ARM_SUBSCRIPTION_ID": mpfConfig.SubscriptionID,
		"ARM_TENANT_ID":       mpfConfig.TenantID,
	}

	tf.SetEnv(envVars)

	err = tf.Init(context.Background())

	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}

	log.Infoln("in apply phase")
	err = tf.Apply(a.ctx, tfexec.VarFile(a.varFilePath))

	if err != nil {
		errorMsg := err.Error()
		log.Debugln(errorMsg)

		if strings.Contains(errorMsg, "Authorization") {
			return errorMsg, nil
		}

		log.Warnf("terraform apply: non authorizaton error occured: %s", errorMsg)

	}

	log.Infoln("in destroy phase")
	err = tf.Destroy(a.ctx, tfexec.VarFile(a.varFilePath))

	if err != nil {
		errorMsg := err.Error()
		log.Debugln(errorMsg)
		if strings.Contains(errorMsg, "Authorization") {
			return errorMsg, nil
		}

		log.Warnf("terraform destroy: non authorizaton error occured: %s", errorMsg)

	}

	return "", nil

}
