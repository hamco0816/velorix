//go:build integration

package tlsfingerprint

import (
	"os"
	"testing"
)

const runNetworkTLSFingerprintTestsEnv = "TLSFINGERPRINT_RUN_NETWORK_TESTS"

func skipUnlessNetworkTLSFingerprintTestsEnabled(t *testing.T) {
	t.Helper()

	if os.Getenv(runNetworkTLSFingerprintTestsEnv) != "1" {
		t.Skipf("skipping external TLS fingerprint test; set %s=1 to run manually", runNetworkTLSFingerprintTestsEnv)
	}
}
