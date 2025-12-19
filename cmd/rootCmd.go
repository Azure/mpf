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
		azmpm terraform --subscriptionID <subscriptionID> --tenantID <tenantID> --spClientID <spClientID> --spObjectID <spObjectID> --spClientSecret <spClientSecret> --tfPath <executablePath> --workingDir <workingDir> --varFilePath <varFilePath>
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
