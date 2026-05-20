package handler

import (
	"testing"
)

func TestRegisterAndGet(t *testing.T) {
	prefix := "TESTREG"
	h := &PingHandler{}
	Register(prefix, h)
	got, ok := Get(prefix)
	if !ok {
		t.Fatal("expected to find registered handler")
	}
	if got != h {
		t.Error("expected to get the same handler back")
	}
	delete(registry, prefix)
}

func TestGetUnknown(t *testing.T) {
	_, ok := Get("UNKNOWN_PREFIX_XYZ")
	if ok {
		t.Error("expected not to find unregistered prefix")
	}
}

func TestKnownPrefixes(t *testing.T) {
	Register("TESTKP", &PingHandler{})
	prefixes := KnownPrefixes()
	if len(prefixes) == 0 {
		t.Error("expected non-empty known prefixes list")
	}
	delete(registry, "TESTKP")
}

func TestIsKnownPrefix(t *testing.T) {
	if !IsKnownPrefix("SHELL") {
		t.Error("expected SHELL to be a known prefix")
	}
	if !IsKnownPrefix("SCRIPT") {
		t.Error("expected SCRIPT to be a known prefix")
	}
	if IsKnownPrefix("NONEXISTENT") {
		t.Error("expected NONEXISTENT to not be a known prefix")
	}
}
