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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Azure/mpf/pkg/domain"
	"github.com/Azure/mpf/pkg/infrastructure/mpfSharedUtils"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

var (
	defaultConfigFilename      = "stingoftheviper"
	envPrefix                  = "MPF"
	replaceHyphenWithCamelCase = false

	flgSubscriptionID     string
	flgTenantID           string
	flgSPClientID         string
	flgSPObjectID         string
	flgSPClientSecret     string
	flgShowDetailedOutput bool
	flgJSONOutput         bool
	flgVerbose            bool
	flgDebug              bool
	flgInitialPermissions string
	// RootCmd            *cobra.Command
)

func NewRootCommand() *cobra.Command {

	rootCmd := &cobra.Command{
		Use:   "azmpf",
		Short: "Find minimum permissions required for Azure deployments",
		Long: `Find minimum permissions required for Azure deployments including ARM and Terraform based deployments. For example:

		This CLI allows you to find the minimum permissions required for Azure deployments including ARM and Terraform based deployments.
		A Service Principal is required to run this CLI. All permissions associated with the Service principal are initially wiped by this command:`,
		Example: `azmpf arm --subscriptionID <subscriptionID> --tenantID <tenantID> --spClientID <spClientID> --spObjectID <spObjectID> --spClientSecret <spClientSecret>
		azmpf terraform --subscriptionID <subscriptionID> --tenantID <tenantID> --spClientID <spClientID> --spObjectID <spObjectID> --spClientSecret <spClientSecret> --tfPath <executablePath> --workingDir <workingDir> --varFilePath <varFilePath>
		`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initializeConfig(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	// Define cobra flags, the default value has the lowest (least significant) precedence
	rootCmd.PersistentFlags().StringVarP(&flgSubscriptionID, "subscriptionID", "s", "", "Azure Subscription ID")
	rootCmd.PersistentFlags().StringVarP(&flgTenantID, "tenantID", "", "", "Azure Tenant ID")
	rootCmd.PersistentFlags().StringVarP(&flgSPClientID, "spClientID", "", "", "Service Principal Client ID")
	rootCmd.PersistentFlags().StringVarP(&flgSPObjectID, "spObjectID", "", "", "Service Principal Object ID")
	rootCmd.PersistentFlags().StringVarP(&flgSPClientSecret, "spClientSecret", "", "", "Service Principal Client Secret")
	rootCmd.PersistentFlags().BoolVarP(&flgShowDetailedOutput, "showDetailedOutput", "", false, "Show detailed output")
	rootCmd.PersistentFlags().BoolVarP(&flgJSONOutput, "jsonOutput", "", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVarP(&flgVerbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&flgDebug, "debug", "d", false, "debug output")
	rootCmd.PersistentFlags().StringVarP(&flgInitialPermissions, "initialPermissions", "", "", "Initial permissions to add to the custom role before starting MPF analysis. Can be a comma-separated list (e.g., 'perm1,perm2') or @path/to/file.json to load from a JSON file with format: {\"RequiredPermissions\":{\"\":[\"perm1\",\"perm2\"]}}.")

	err := rootCmd.MarkPersistentFlagRequired("subscriptionID")
	if err != nil {
		log.Errorf("Error marking flag required for subscription ID: %v\n", err)
	}

	err = rootCmd.MarkPersistentFlagRequired("tenantID")
	if err != nil {
		log.Errorf("Error marking flag required for tenant ID: %v\n", err)
	}

	err = rootCmd.MarkPersistentFlagRequired("spClientID")
	if err != nil {
		log.Errorf("Error marking flag required for SP client ID: %v\n", err)
	}

	err = rootCmd.MarkPersistentFlagRequired("spObjectID")
	if err != nil {
		log.Errorf("Error marking flag required for SP object ID: %v\n", err)
	}

	err = rootCmd.MarkPersistentFlagRequired("spClientSecret")
	if err != nil {
		log.Errorf("Error marking flag required for SP client secret: %v\n", err)
	}

	rootCmd.MarkFlagsMutuallyExclusive("showDetailedOutput", "jsonOutput")

	// Add subcommands
	rootCmd.AddCommand(NewARMCommand())
	rootCmd.AddCommand(NewBicepCommand())
	rootCmd.AddCommand(NewTerraformCommand())

	return rootCmd
}

func initializeConfig(cmd *cobra.Command) error {
	v := viper.New()

	v.SetConfigName(defaultConfigFilename)

	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		// It's okay if there isn't a config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	bindFlags(cmd, v)

	return nil
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Determine the naming convention of the flags when represented in the config file
		configName := f.Name
		// If using camelCase in the config file, replace hyphens with a camelCased string.
		// Since viper does case-insensitive comparisons, we don't need to bother fixing the case, and only need to remove the hyphens.
		if replaceHyphenWithCamelCase {
			configName = strings.ReplaceAll(f.Name, "-", "")
		}

		// Bind the flag to its environment variables. In addition to the legacy
		// concatenated form (e.g. MPF_SUBSCRIPTIONID), also accept a snake_case form
		// (e.g. MPF_SUBSCRIPTION_ID) so environment variables are easier to read.
		bindEnvVars(v, configName)

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(configName) {
			val := v.Get(configName)
			err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
			if err != nil {
				log.Errorf("Error setting flag %s: %v\n", f.Name, err)
			}
		}
	})
}

// bindEnvVars binds a viper config key to its supported environment variables. Both the
// legacy concatenated form (e.g. MPF_SUBSCRIPTIONID) and a snake_case form
// (e.g. MPF_SUBSCRIPTION_ID) are accepted so that existing setups keep working while
// newer, more readable variable names are also supported.
func bindEnvVars(v *viper.Viper, configName string) {
	legacyEnv := envPrefix + "_" + strings.ToUpper(strings.ReplaceAll(configName, "-", ""))
	snakeEnv := envPrefix + "_" + camelCaseToSnakeUpper(configName)

	var err error
	if snakeEnv != legacyEnv {
		err = v.BindEnv(configName, legacyEnv, snakeEnv)
	} else {
		err = v.BindEnv(configName, legacyEnv)
	}
	if err != nil {
		log.Errorf("Error binding environment variables for %s: %v\n", configName, err)
	}
}

// camelCaseToSnakeUpper converts a camelCase identifier to an UPPER_SNAKE_CASE string.
// For example, "subscriptionID" becomes "SUBSCRIPTION_ID" and "spClientID" becomes "SP_CLIENT_ID".
func camelCaseToSnakeUpper(s string) string {
	var b strings.Builder
	runes := []rune(strings.ReplaceAll(s, "-", "_"))
	for i, r := range runes {
		if i > 0 && r >= 'A' && r <= 'Z' {
			prev := runes[i-1]
			prevIsLowerOrDigit := (prev >= 'a' && prev <= 'z') || (prev >= '0' && prev <= '9')
			nextIsLower := i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z'
			// Insert an underscore at a lowercase->uppercase boundary (e.g. "subscriptionID" -> "subscription_ID")
			// or at an acronym->word boundary (e.g. "APIVersion" -> "API_Version"), but not between
			// consecutive acronym letters (e.g. the "ID" in "subscriptionID" stays together).
			if prevIsLowerOrDigit || (nextIsLower && prev != '_') {
				b.WriteRune('_')
			}
		}
		b.WriteRune(r)
	}
	return strings.ToUpper(b.String())
}

func setLogLevel() {
	if flgVerbose {
		log.SetLevel(log.InfoLevel)
	}
	if flgDebug {
		log.SetLevel(log.DebugLevel)
	}
}

func getRootMPFConfig() domain.MPFConfig {
	mpfRole := domain.Role{}

	roleDefUUID, _ := uuid.NewRandom()
	mpfRole.RoleDefinitionID = roleDefUUID.String()
	mpfRole.RoleDefinitionName = fmt.Sprintf("tmp-rol-%s", mpfSharedUtils.GenerateRandomString(7))
	mpfRole.RoleDefinitionResourceID = fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Authorization/roleDefinitions/%s", flgSubscriptionID, mpfRole.RoleDefinitionID)
	log.Infoln("roleDefinitionResourceID:", mpfRole.RoleDefinitionResourceID)

	return domain.MPFConfig{
		SubscriptionID: flgSubscriptionID,
		TenantID:       flgTenantID,
		SP: domain.ServicePrincipal{
			SPClientID:     flgSPClientID,
			SPObjectID:     flgSPObjectID,
			SPClientSecret: flgSPClientSecret,
		},
		Role: mpfRole,
	}
}

func getAbsolutePath(path string) (string, error) {
	absPath := path
	if !filepath.IsAbs(path) {

		absWorkingDir, err := os.Getwd()
		if err != nil {
			return "", err
		}
		absPath = absWorkingDir + "/" + absPath
	}
	return absPath, nil
}

// parseInitialPermissions parses the initial permissions from either a comma-separated string
// or from a JSON file (if the value starts with @).
// The JSON file should have the same format as .permissionsFromFailedRun.json:
// {"RequiredPermissions":{"":["perm1","perm2"]}}
func parseInitialPermissions(value string) ([]string, error) {
	if value == "" {
		return nil, nil
	}

	// Check if it's a file reference (starts with @)
	if after, ok := strings.CutPrefix(value, "@"); ok {
		filePath := after
		absPath, err := getAbsolutePath(filePath)
		if err != nil {
			return nil, fmt.Errorf("error getting absolute path for permissions file: %w", err)
		}

		file, err := os.Open(absPath)
		if err != nil {
			return nil, fmt.Errorf("error opening permissions file %s: %w", absPath, err)
		}
		defer file.Close() //nolint:errcheck

		var result domain.MPFResult
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&result); err != nil {
			return nil, fmt.Errorf("error parsing permissions file %s: %w", absPath, err)
		}

		permissions := result.RequiredPermissions[""]
		if len(permissions) == 0 {
			log.Warnf("No permissions found in file %s under the empty string key", absPath)
		}
		return permissions, nil
	}

	// Parse as comma-separated string
	permissions := strings.Split(value, ",")
	for i := range permissions {
		permissions[i] = strings.TrimSpace(permissions[i])
	}
	return permissions, nil
}

// appendUserInitialPermissions parses the --initialPermissions flag and appends
// the permissions to both slices. This is a helper to reduce code duplication
// across arm, bicep, and terraform commands.
func appendUserInitialPermissions(initialPermissionsToAdd, permissionsToAddToResult []string) ([]string, []string) {
	if flgInitialPermissions == "" {
		return initialPermissionsToAdd, permissionsToAddToResult
	}

	userPermissions, err := parseInitialPermissions(flgInitialPermissions)
	if err != nil {
		log.Fatalf("Error parsing initial permissions: %v\n", err)
	}

	if len(userPermissions) > 0 {
		log.Infof("Adding user-specified initial permissions: %v\n", userPermissions)
		initialPermissionsToAdd = append(initialPermissionsToAdd, userPermissions...)
		permissionsToAddToResult = append(permissionsToAddToResult, userPermissions...)
	}

	return initialPermissionsToAdd, permissionsToAddToResult
}
