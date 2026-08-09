package svc

import "testing"

func HelperExport() string {
	return "ok"
}

func TestUsesHelper(t *testing.T) {
	if HelperExport() != "ok" {
		t.Fatal("unexpected")
	}
}
