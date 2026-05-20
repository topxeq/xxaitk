package handler

import (
	"testing"
)

func TestCapabilitiesAll(t *testing.T) {
	h := &CapabilitiesHandler{}
	resp := h.Handle("", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	caps, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("expected map[string]interface{}")
	}
	if _, exists := caps["version"]; !exists {
		t.Error("expected 'version' key in capabilities")
	}
	if _, exists := caps["prefixes"]; !exists {
		t.Error("expected 'prefixes' key in capabilities")
	}
	if _, exists := caps["builtins"]; !exists {
		t.Error("expected 'builtins' key in capabilities")
	}
	if _, exists := caps["features"]; !exists {
		t.Error("expected 'features' key in capabilities")
	}
}

func TestCapabilitiesVersion(t *testing.T) {
	h := &CapabilitiesHandler{}
	resp := h.Handle("version", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("expected map[string]interface{}")
	}
	version, exists := data["version"]
	if !exists {
		t.Error("expected 'version' key")
	}
	if version != Version {
		t.Errorf("expected version %s, got: %v", Version, version)
	}
}

func TestCapabilitiesPrefixes(t *testing.T) {
	h := &CapabilitiesHandler{}
	resp := h.Handle("prefixes", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("expected map[string]interface{}")
	}
	if _, exists := data["prefixes"]; !exists {
		t.Error("expected 'prefixes' key")
	}
}

func TestCapabilitiesBuiltins(t *testing.T) {
	h := &CapabilitiesHandler{}
	resp := h.Handle("builtins", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("expected map[string]interface{}")
	}
	if _, exists := data["builtins"]; !exists {
		t.Error("expected 'builtins' key")
	}
}

func TestCapabilitiesUnknown(t *testing.T) {
	h := &CapabilitiesHandler{}
	resp := h.Handle("bad", "")
	if resp.Ok {
		t.Fatal("expected failure for unknown query")
	}
	if resp.Error.Code != "CAPABILITIES_UNKNOWN_QUERY" {
		t.Errorf("expected error code CAPABILITIES_UNKNOWN_QUERY, got: %s", resp.Error.Code)
	}
}
