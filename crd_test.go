package kgetset

import (
	"testing"
)

func TestCRD(t *testing.T) {
	mock := newCRD()
	err := mock.test()
	if err != nil {
		t.Fatalf("crd test failed: %+v", err)
	}
}
