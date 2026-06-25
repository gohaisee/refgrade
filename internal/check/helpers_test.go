package check

import "testing"

func TestShouldSkipIntegrationE2E(t *testing.T) {
	t.Parallel()
	if !shouldSkipIntegrationE2E(nil) {
		t.Fatal("expected skip without tags")
	}
	if shouldSkipIntegrationE2E([]string{"integration"}) {
		t.Fatal("integration tag should disable skip")
	}
}
