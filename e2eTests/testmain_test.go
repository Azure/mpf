package e2etests

import (
	"os"
	"testing"
)

func ensureAzureEnvForE2E(t *testing.T) {
	t.Helper()

	// The product code uses DefaultAzureCredential in some paths.
	// E2E tests authenticate via MPF_* env vars (service principal creds), so we
	// map those to AZURE_* env vars (if unset) to enable EnvironmentCredential.
	setFromIfUnset(t, "AZURE_TENANT_ID", "MPF_TENANTID")
	setFromIfUnset(t, "AZURE_CLIENT_ID", "MPF_SPCLIENTID")
	setFromIfUnset(t, "AZURE_CLIENT_SECRET", "MPF_SPCLIENTSECRET")
	setFromIfUnset(t, "AZURE_SUBSCRIPTION_ID", "MPF_SUBSCRIPTIONID")
}

func setFromIfUnset(t *testing.T, dst, src string) {
	t.Helper()

	if os.Getenv(dst) != "" {
		return
	}
	if v := os.Getenv(src); v != "" {
		t.Setenv(dst, v)
	}
}
