package handler

import (
	"encoding/json"
	"runtime"
	"testing"
)

func TestInfoOS(t *testing.T) {
	h := &InfoHandler{}
	resp := h.Handle("os", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(OSInfoResult)
	if result.Name != runtime.GOOS {
		t.Errorf("expected name %s, got: %s", runtime.GOOS, result.Name)
	}
	if result.Arch != runtime.GOARCH {
		t.Errorf("expected arch %s, got: %s", runtime.GOARCH, result.Arch)
	}
}

func TestInfoCPU(t *testing.T) {
	h := &InfoHandler{}
	resp := h.Handle("cpu", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(CPUInfoResult)
	if result.Count <= 0 {
		t.Errorf("expected cpu count > 0, got: %d", result.Count)
	}
	if result.Count != runtime.NumCPU() {
		t.Errorf("expected cpu count %d, got: %d", runtime.NumCPU(), result.Count)
	}
}

func TestInfoMem(t *testing.T) {
	h := &InfoHandler{}
	resp := h.Handle("mem", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(MemInfoResult)
	if result.AllocKB == 0 {
		t.Error("expected alloc_kb > 0")
	}
	if result.SysKB == 0 {
		t.Error("expected sys_kb > 0")
	}
}

func TestInfoEnv(t *testing.T) {
	h := &InfoHandler{}
	resp := h.Handle("env", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(EnvInfoResult)
	if result.Vars == nil {
		t.Error("expected vars map to not be nil")
	}
}

func TestInfoAll(t *testing.T) {
	h := &InfoHandler{}
	resp := h.Handle("all", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*AllInfoResult)
	if result.OS.Name != runtime.GOOS {
		t.Errorf("expected os name %s, got: %s", runtime.GOOS, result.OS.Name)
	}
	if result.CPU.Count != runtime.NumCPU() {
		t.Errorf("expected cpu count %d, got: %d", runtime.NumCPU(), result.CPU.Count)
	}
	if result.Mem.AllocKB == 0 {
		t.Error("expected alloc_kb > 0")
	}
}

func TestInfoUnknownQuery(t *testing.T) {
	h := &InfoHandler{}
	resp := h.Handle("bad_query", "")
	if resp.Ok {
		t.Fatal("expected failure for unknown query")
	}
	if resp.Error.Code != "INFO_UNKNOWN_QUERY" {
		t.Errorf("expected error code INFO_UNKNOWN_QUERY, got: %s", resp.Error.Code)
	}
}

func TestInfoEmptyQuery(t *testing.T) {
	h := &InfoHandler{}
	resp := h.Handle("", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*AllInfoResult)
	if result.OS.Name != runtime.GOOS {
		t.Errorf("expected os name %s, got: %s", runtime.GOOS, result.OS.Name)
	}
}

func TestInfoJSONPayload(t *testing.T) {
	h := &InfoHandler{}
	payload := InfoPayload{Query: "cpu"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(CPUInfoResult)
	if result.Count != runtime.NumCPU() {
		t.Errorf("expected cpu count %d, got: %d", runtime.NumCPU(), result.Count)
	}
}
