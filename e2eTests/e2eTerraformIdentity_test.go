// MIT License
//
// Copyright (c) Microsoft Corporation.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package e2etests

import (
	"errors"
	"os"
	"sync/atomic"
	"testing"

	log "github.com/sirupsen/logrus"
)

const (
	terraformAltSPClientIDEnv     = "MPF_TERRAFORM_ALT_SPCLIENTID"
	terraformAltSPClientSecretEnv = "MPF_TERRAFORM_ALT_SPCLIENTSECRET"
	terraformAltSPObjectIDEnv     = "MPF_TERRAFORM_ALT_SPOBJECTID"
)

var terraformIdentitySequence atomic.Uint64
var errMissingAlternateTerraformCredentials = errors.New("required alternate Terraform service principal environment variables not set")

func getTestingTerraformMPFArgs(t *testing.T) (MpfCLIArgs, error) {
	t.Helper()

	mpfArgs, err := getTestingMPFArgs()
	if err != nil {
		return MpfCLIArgs{}, err
	}

	slot := terraformIdentitySequence.Add(1)
	mpfArgs, identity, err := selectTerraformIdentity(mpfArgs, slot)
	if err != nil {
		t.Fatalf("selecting Terraform E2E identity: %v", err)
	}
	log.Infof("Terraform E2E identity for %s: %s", t.Name(), identity)

	return mpfArgs, nil
}

func selectTerraformIdentity(mpfArgs MpfCLIArgs, slot uint64) (MpfCLIArgs, string, error) {
	if slot%2 == 0 {
		return mpfArgs, "primary", nil
	}

	clientID := os.Getenv(terraformAltSPClientIDEnv)
	clientSecret := os.Getenv(terraformAltSPClientSecretEnv)
	objectID := os.Getenv(terraformAltSPObjectIDEnv)
	if clientID == "" && clientSecret == "" && objectID == "" {
		return mpfArgs, "primary (alternate not configured)", nil
	}
	if clientID == "" || clientSecret == "" || objectID == "" {
		return MpfCLIArgs{}, "", errMissingAlternateTerraformCredentials
	}

	mpfArgs.SPClientID = clientID
	mpfArgs.SPClientSecret = clientSecret
	mpfArgs.SPObjectID = objectID

	return mpfArgs, "alternate", nil
}

func TestE2EIdentitySelectorAlternates(t *testing.T) {
	t.Setenv(terraformAltSPClientIDEnv, "alternate-client")
	t.Setenv(terraformAltSPClientSecretEnv, "alternate-secret")
	t.Setenv(terraformAltSPObjectIDEnv, "alternate-object")
	primary := MpfCLIArgs{
		SPClientID:     "primary-client",
		SPClientSecret: "primary-secret",
		SPObjectID:     "primary-object",
	}

	alternate, identity, err := selectTerraformIdentity(primary, 1)
	if err != nil {
		t.Fatalf("selecting alternate identity: %v", err)
	}
	if identity != "alternate" ||
		alternate.SPClientID != "alternate-client" ||
		alternate.SPClientSecret != "alternate-secret" ||
		alternate.SPObjectID != "alternate-object" {
		t.Fatalf("unexpected alternate identity: slot=%q client=%q object=%q", identity, alternate.SPClientID, alternate.SPObjectID)
	}

	selectedPrimary, identity, err := selectTerraformIdentity(primary, 2)
	if err != nil {
		t.Fatalf("selecting primary identity: %v", err)
	}
	if identity != "primary" ||
		selectedPrimary.SPClientID != "primary-client" ||
		selectedPrimary.SPClientSecret != "primary-secret" ||
		selectedPrimary.SPObjectID != "primary-object" {
		t.Fatalf("unexpected primary identity: slot=%q client=%q object=%q", identity, selectedPrimary.SPClientID, selectedPrimary.SPObjectID)
	}

	alternateAgain, identity, err := selectTerraformIdentity(primary, 3)
	if err != nil {
		t.Fatalf("selecting alternate identity again: %v", err)
	}
	if identity != "alternate" || alternateAgain.SPClientID != "alternate-client" {
		t.Fatalf("expected selector to return to alternate identity, got slot=%q client=%q", identity, alternateAgain.SPClientID)
	}
}

func TestE2EIdentitySelectorRejectsIncompleteAlternateCredentials(t *testing.T) {
	t.Setenv(terraformAltSPClientIDEnv, "alternate-client")
	t.Setenv(terraformAltSPClientSecretEnv, "")
	t.Setenv(terraformAltSPObjectIDEnv, "")

	_, _, err := selectTerraformIdentity(MpfCLIArgs{}, 1)
	if !errors.Is(err, errMissingAlternateTerraformCredentials) {
		t.Fatalf("expected missing alternate credentials error, got %v", err)
	}
}

func TestE2EIdentitySelectorFallsBackWhenAlternateIsNotConfigured(t *testing.T) {
	t.Setenv(terraformAltSPClientIDEnv, "")
	t.Setenv(terraformAltSPClientSecretEnv, "")
	t.Setenv(terraformAltSPObjectIDEnv, "")
	primary := MpfCLIArgs{SPClientID: "primary-client"}

	mpfArgs, identity, err := selectTerraformIdentity(primary, 1)
	if err != nil {
		t.Fatalf("selecting fallback identity: %v", err)
	}
	if identity != "primary (alternate not configured)" || mpfArgs.SPClientID != "primary-client" {
		t.Fatalf("expected primary fallback identity, got slot=%q client=%q", identity, mpfArgs.SPClientID)
	}
}
