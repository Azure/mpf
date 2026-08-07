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

package main

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestCamelToSnakeCase(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"verbose", "verbose"},
		{"debug", "debug"},
		{"location", "location"},
		{"subscriptionID", "subscription_id"},
		{"tenantID", "tenant_id"},
		{"spClientID", "sp_client_id"},
		{"spObjectID", "sp_object_id"},
		{"spClientSecret", "sp_client_secret"},
		{"showDetailedOutput", "show_detailed_output"},
		{"jsonOutput", "json_output"},
		{"initialPermissions", "initial_permissions"},
		{"templateFilePath", "template_file_path"},
		{"parametersFilePath", "parameters_file_path"},
		{"resourceGroupNamePfx", "resource_group_name_pfx"},
		{"deploymentNamePfx", "deployment_name_pfx"},
		{"bicepFilePath", "bicep_file_path"},
		{"bicepExecPath", "bicep_exec_path"},
		{"tfPath", "tf_path"},
		{"workingDir", "working_dir"},
		{"varFilePath", "var_file_path"},
		{"importExistingResourcesToState", "import_existing_resources_to_state"},
		{"targetModule", "target_module"},
		// Acronym edge cases
		{"JSONOutput", "json_output"},
		{"parseURLToHTML", "parse_url_to_html"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := camelToSnakeCase(tt.in)
			if got != tt.want {
				t.Errorf("camelToSnakeCase(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestEnvNamesForFlag(t *testing.T) {
	tests := []struct {
		flag string
		want []string
	}{
		{"verbose", []string{"MPF_VERBOSE"}},
		{"subscriptionID", []string{"MPF_SUBSCRIPTIONID", "MPF_SUBSCRIPTION_ID"}},
		{"spClientID", []string{"MPF_SPCLIENTID", "MPF_SP_CLIENT_ID"}},
		{"showDetailedOutput", []string{"MPF_SHOWDETAILEDOUTPUT", "MPF_SHOW_DETAILED_OUTPUT"}},
		{"importExistingResourcesToState", []string{"MPF_IMPORTEXISTINGRESOURCESTOSTATE", "MPF_IMPORT_EXISTING_RESOURCES_TO_STATE"}},
		{"templateFilePath", []string{"MPF_TEMPLATEFILEPATH", "MPF_TEMPLATE_FILE_PATH"}},
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			got := envNamesForFlag(tt.flag)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("envNamesForFlag(%q) = %v, want %v", tt.flag, got, tt.want)
			}
		})
	}
}

// unsetEnvForTest removes key for the duration of the test and restores it afterward.
// Empty-string values are not used because LookupEnv treats "" as set, which would
// make the legacy env name shadow snake_case values (legacy is checked first).
func unsetEnvForTest(t *testing.T, key string) {
	t.Helper()
	orig, had := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unsetenv %s: %v", key, err)
	}
	t.Cleanup(func() {
		if had {
			_ = os.Setenv(key, orig)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}

// newTestViper mirrors initializeConfig's viper setup for unit tests.
func newTestViper() *viper.Viper {
	v := viper.New()
	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()
	return v
}

func TestBindFlags_SnakeCaseEnvVars(t *testing.T) {
	// Ensure legacy names do not leak from the developer environment.
	for _, name := range []string{
		"MPF_SUBSCRIPTIONID",
		"MPF_TENANTID",
		"MPF_SPCLIENTID",
		"MPF_SPOBJECTID",
		"MPF_SPCLIENTSECRET",
		"MPF_SHOWDETAILEDOUTPUT",
		"MPF_TEMPLATEFILEPATH",
	} {
		unsetEnvForTest(t, name)
	}

	t.Setenv("MPF_SUBSCRIPTION_ID", "snake-sub-id")
	t.Setenv("MPF_TENANT_ID", "snake-tenant-id")
	t.Setenv("MPF_SP_CLIENT_ID", "snake-sp-client-id")
	t.Setenv("MPF_SP_OBJECT_ID", "snake-sp-object-id")
	t.Setenv("MPF_SP_CLIENT_SECRET", "snake-sp-secret")
	t.Setenv("MPF_SHOW_DETAILED_OUTPUT", "true")
	t.Setenv("MPF_TEMPLATE_FILE_PATH", "/tmp/template.json")
	t.Setenv("MPF_VERBOSE", "true")

	var (
		subscriptionID     string
		tenantID           string
		spClientID         string
		spObjectID         string
		spClientSecret     string
		showDetailedOutput bool
		templateFilePath   string
		verbose            bool
	)

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().StringVar(&subscriptionID, "subscriptionID", "", "")
	cmd.Flags().StringVar(&tenantID, "tenantID", "", "")
	cmd.Flags().StringVar(&spClientID, "spClientID", "", "")
	cmd.Flags().StringVar(&spObjectID, "spObjectID", "", "")
	cmd.Flags().StringVar(&spClientSecret, "spClientSecret", "", "")
	cmd.Flags().BoolVar(&showDetailedOutput, "showDetailedOutput", false, "")
	cmd.Flags().StringVar(&templateFilePath, "templateFilePath", "", "")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "")

	bindFlags(cmd, newTestViper())

	assertEq := func(name, got, want string) {
		t.Helper()
		if got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
	assertEq("subscriptionID", subscriptionID, "snake-sub-id")
	assertEq("tenantID", tenantID, "snake-tenant-id")
	assertEq("spClientID", spClientID, "snake-sp-client-id")
	assertEq("spObjectID", spObjectID, "snake-sp-object-id")
	assertEq("spClientSecret", spClientSecret, "snake-sp-secret")
	assertEq("templateFilePath", templateFilePath, "/tmp/template.json")
	if !showDetailedOutput {
		t.Errorf("showDetailedOutput = false, want true")
	}
	if !verbose {
		t.Errorf("verbose = false, want true")
	}
}

func TestBindFlags_LegacyEnvVarsStillWork(t *testing.T) {
	// Clear snake_case variants so only legacy names are used.
	for _, name := range []string{
		"MPF_SUBSCRIPTION_ID",
		"MPF_TENANT_ID",
		"MPF_SP_CLIENT_ID",
		"MPF_TEMPLATE_FILE_PATH",
	} {
		unsetEnvForTest(t, name)
	}

	t.Setenv("MPF_SUBSCRIPTIONID", "legacy-sub-id")
	t.Setenv("MPF_TENANTID", "legacy-tenant-id")
	t.Setenv("MPF_SPCLIENTID", "legacy-sp-client-id")
	t.Setenv("MPF_TEMPLATEFILEPATH", "/legacy/template.json")

	var (
		subscriptionID   string
		tenantID         string
		spClientID       string
		templateFilePath string
	)

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().StringVar(&subscriptionID, "subscriptionID", "", "")
	cmd.Flags().StringVar(&tenantID, "tenantID", "", "")
	cmd.Flags().StringVar(&spClientID, "spClientID", "", "")
	cmd.Flags().StringVar(&templateFilePath, "templateFilePath", "", "")

	bindFlags(cmd, newTestViper())

	if subscriptionID != "legacy-sub-id" {
		t.Errorf("subscriptionID = %q, want %q", subscriptionID, "legacy-sub-id")
	}
	if tenantID != "legacy-tenant-id" {
		t.Errorf("tenantID = %q, want %q", tenantID, "legacy-tenant-id")
	}
	if spClientID != "legacy-sp-client-id" {
		t.Errorf("spClientID = %q, want %q", spClientID, "legacy-sp-client-id")
	}
	if templateFilePath != "/legacy/template.json" {
		t.Errorf("templateFilePath = %q, want %q", templateFilePath, "/legacy/template.json")
	}
}

func TestBindFlags_LegacyTakesPrecedenceWhenBothSet(t *testing.T) {
	// Legacy is checked first so existing deployments are not surprised if both forms are present.
	t.Setenv("MPF_SUBSCRIPTIONID", "from-legacy")
	t.Setenv("MPF_SUBSCRIPTION_ID", "from-snake")

	var subscriptionID string
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().StringVar(&subscriptionID, "subscriptionID", "", "")

	bindFlags(cmd, newTestViper())

	if subscriptionID != "from-legacy" {
		t.Errorf("subscriptionID = %q, want legacy value %q when both env forms are set", subscriptionID, "from-legacy")
	}
}

func TestBindFlags_CLIFlagTakesPrecedenceOverEnv(t *testing.T) {
	t.Setenv("MPF_SUBSCRIPTION_ID", "from-env")
	t.Setenv("MPF_SUBSCRIPTIONID", "from-legacy-env")

	var subscriptionID string
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().StringVar(&subscriptionID, "subscriptionID", "", "")

	// Simulate the flag being set on the CLI before bindFlags runs.
	if err := cmd.Flags().Set("subscriptionID", "from-cli"); err != nil {
		t.Fatalf("failed to set CLI flag: %v", err)
	}

	bindFlags(cmd, newTestViper())

	if subscriptionID != "from-cli" {
		t.Errorf("subscriptionID = %q, want CLI value %q", subscriptionID, "from-cli")
	}
}

func TestBindFlags_TerraformProviderFlags(t *testing.T) {
	for _, name := range []string{
		"MPF_TFPATH",
		"MPF_WORKINGDIR",
		"MPF_VARFILEPATH",
		"MPF_IMPORTEXISTINGRESOURCESTOSTATE",
		"MPF_TARGETMODULE",
	} {
		unsetEnvForTest(t, name)
	}

	t.Setenv("MPF_TF_PATH", "/usr/bin/terraform")
	t.Setenv("MPF_WORKING_DIR", "/work/tf")
	t.Setenv("MPF_VAR_FILE_PATH", "/work/tf/dev.tfvars")
	t.Setenv("MPF_IMPORT_EXISTING_RESOURCES_TO_STATE", "false")
	t.Setenv("MPF_TARGET_MODULE", "module.law")

	var (
		tfPath                         string
		workingDir                     string
		varFilePath                    string
		importExistingResourcesToState bool
		targetModule                   string
	)

	cmd := &cobra.Command{Use: "terraform"}
	cmd.Flags().StringVar(&tfPath, "tfPath", "", "")
	cmd.Flags().StringVar(&workingDir, "workingDir", "", "")
	cmd.Flags().StringVar(&varFilePath, "varFilePath", "", "")
	// Default true mirrors the real terraform command flag.
	cmd.Flags().BoolVar(&importExistingResourcesToState, "importExistingResourcesToState", true, "")
	cmd.Flags().StringVar(&targetModule, "targetModule", "", "")

	bindFlags(cmd, newTestViper())

	if tfPath != "/usr/bin/terraform" {
		t.Errorf("tfPath = %q, want %q", tfPath, "/usr/bin/terraform")
	}
	if workingDir != "/work/tf" {
		t.Errorf("workingDir = %q, want %q", workingDir, "/work/tf")
	}
	if varFilePath != "/work/tf/dev.tfvars" {
		t.Errorf("varFilePath = %q, want %q", varFilePath, "/work/tf/dev.tfvars")
	}
	if importExistingResourcesToState {
		t.Errorf("importExistingResourcesToState = true, want false")
	}
	if targetModule != "module.law" {
		t.Errorf("targetModule = %q, want %q", targetModule, "module.law")
	}
}

func TestInitializeConfig_BindsRootAndSubcommandSnakeCaseEnv(t *testing.T) {
	// Clear legacy forms that may be present in the environment.
	for _, name := range []string{
		"MPF_SUBSCRIPTIONID", "MPF_TENANTID", "MPF_SPCLIENTID",
		"MPF_SPOBJECTID", "MPF_SPCLIENTSECRET", "MPF_TEMPLATEFILEPATH",
		"MPF_PARAMETERSFILEPATH",
	} {
		unsetEnvForTest(t, name)
	}

	t.Setenv("MPF_SUBSCRIPTION_ID", "init-sub")
	t.Setenv("MPF_TENANT_ID", "init-tenant")
	t.Setenv("MPF_SP_CLIENT_ID", "init-client")
	t.Setenv("MPF_SP_OBJECT_ID", "init-object")
	t.Setenv("MPF_SP_CLIENT_SECRET", "init-secret")
	t.Setenv("MPF_TEMPLATE_FILE_PATH", "/arm/template.json")
	t.Setenv("MPF_PARAMETERS_FILE_PATH", "/arm/params.json")

	// Reset package-level flag vars that NewRootCommand binds to.
	flgSubscriptionID = ""
	flgTenantID = ""
	flgSPClientID = ""
	flgSPObjectID = ""
	flgSPClientSecret = ""
	flgTemplateFilePath = ""
	flgParametersFilePath = ""

	root := NewRootCommand()
	// Locate the arm subcommand (includes inherited persistent flags after merge).
	armCmd, _, err := root.Find([]string{"arm"})
	if err != nil {
		t.Fatalf("Find arm: %v", err)
	}

	// Merge inherited persistent flags the same way cobra does before PersistentPreRun.
	armCmd.Flags().AddFlagSet(root.PersistentFlags())

	if err := initializeConfig(armCmd); err != nil {
		t.Fatalf("initializeConfig: %v", err)
	}

	if flgSubscriptionID != "init-sub" {
		t.Errorf("flgSubscriptionID = %q, want %q", flgSubscriptionID, "init-sub")
	}
	if flgTenantID != "init-tenant" {
		t.Errorf("flgTenantID = %q, want %q", flgTenantID, "init-tenant")
	}
	if flgSPClientID != "init-client" {
		t.Errorf("flgSPClientID = %q, want %q", flgSPClientID, "init-client")
	}
	if flgSPObjectID != "init-object" {
		t.Errorf("flgSPObjectID = %q, want %q", flgSPObjectID, "init-object")
	}
	if flgSPClientSecret != "init-secret" {
		t.Errorf("flgSPClientSecret = %q, want %q", flgSPClientSecret, "init-secret")
	}
	if flgTemplateFilePath != "/arm/template.json" {
		t.Errorf("flgTemplateFilePath = %q, want %q", flgTemplateFilePath, "/arm/template.json")
	}
	if flgParametersFilePath != "/arm/params.json" {
		t.Errorf("flgParametersFilePath = %q, want %q", flgParametersFilePath, "/arm/params.json")
	}
}
