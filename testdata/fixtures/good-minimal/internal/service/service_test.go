package service

import "testing"

func TestService_Name(t *testing.T) {
	s := New("test")
	if s.Name() != "test" {
		t.Fatal("name mismatch")
	}
}
