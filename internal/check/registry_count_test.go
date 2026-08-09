package check

import "testing"

func TestRegistry_hasAtLeast150Checks(t *testing.T) {
	t.Parallel()
	if n := len(AllCheckIDs()); n < 150 {
		t.Fatalf("registry has %d checks, want at least 150", n)
	}
}
