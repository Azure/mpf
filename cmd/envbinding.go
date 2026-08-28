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
	"strings"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

// camelToSnakeCase converts a camelCase or PascalCase identifier to snake_case.
// Consecutive uppercase acronyms are handled so that "subscriptionID" becomes
// "subscription_id" and "spClientID" becomes "sp_client_id".
func camelToSnakeCase(s string) string {
	if s == "" {
		return s
	}

	runes := []rune(s)
	var b strings.Builder
	b.Grow(len(s) + 4)

	for i, r := range runes {
		if unicode.IsUpper(r) {
			// Insert underscore before this uppercase rune when it starts a new word:
			// - previous rune is lowercase (e.g. nID -> n_id), or
			// - previous is uppercase and the next is lowercase (end of acronym before a new word).
			if i > 0 {
				prev := runes[i-1]
				nextIsLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
				if unicode.IsLower(prev) || (unicode.IsUpper(prev) && nextIsLower) {
					b.WriteByte('_')
				}
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteRune(r)
	}

	return b.String()
}

// envNamesForFlag returns the full environment variable names accepted for a
// given cobra/viper flag name.
//
// Two forms are supported for backward compatibility:
//  1. Legacy concatenated form derived from the flag name as-is, e.g. subscriptionID -> MPF_SUBSCRIPTIONID
//  2. Snake_case form, e.g. subscriptionID -> MPF_SUBSCRIPTION_ID
//
// When both resolve to the same name (single-word flags like "verbose"), only one entry is returned.
// Order is legacy first, then snake_case, so existing deployments keep their current value if both are set.
func envNamesForFlag(flagName string) []string {
	legacy := envPrefix + "_" + strings.ToUpper(flagName)
	snake := envPrefix + "_" + strings.ToUpper(camelToSnakeCase(flagName))

	if legacy == snake {
		return []string{legacy}
	}
	return []string{legacy, snake}
}

// bindFlags applies viper config and environment values to cobra flags that
// were not set on the command line. Each flag accepts both the legacy
// concatenated env var name and the snake_case form (see envNamesForFlag).
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		configName := f.Name
		// If using camelCase in the config file, replace hyphens with a camelCased string.
		// Since viper does case-insensitive comparisons, we don't need to bother fixing the case, and only need to remove the hyphens.
		if replaceHyphenWithCamelCase {
			configName = strings.ReplaceAll(f.Name, "-", "")
		}

		// Explicit BindEnv is required so both legacy and snake_case env names are
		// recognized. When BindEnv is given explicit names they are used as-is
		// (prefix is not re-applied).
		envNames := envNamesForFlag(configName)
		bindArgs := append([]string{configName}, envNames...)
		if err := v.BindEnv(bindArgs...); err != nil {
			log.Errorf("Error binding env vars for flag %s: %v", f.Name, err)
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(configName) {
			val := v.Get(configName)
			err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
			if err != nil {
				log.Errorf("Error setting flag %s: %v", f.Name, err)
			}
		}
	})
}
