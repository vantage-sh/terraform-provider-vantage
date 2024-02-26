package acctest

import (
	"os"
	"testing"
)

var VantageEnvVars = []string{
	"VANTAGE_API_TOKEN",
}

func PreCheck(t *testing.T) {
	t.Helper()

	// Ensure environment is properly configured for acceptance tests.
	for _, envVar := range VantageEnvVars {
		if _, ok := os.LookupEnv(envVar); !ok {
			t.Fatalf("environment variable %s must be set", envVar)
		}
	}
}
