package output

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestSuccess(t *testing.T) {
	start := time.Now()
	resp := Success("shell", "inline", "result data", start)
	if !resp.Ok {
		t.Error("Ok = false, want true")
	}
	if resp.Type != "shell" {
		t.Errorf("Type = %q, want %q", resp.Type, "shell")
	}
	if resp.Source != "inline" {
		t.Errorf("Source = %q, want %q", resp.Source, "inline")
	}
	if resp.Data != "result data" {
		t.Errorf("Data = %v, want %q", resp.Data, "result data")
	}
	if resp.Env == nil {
		t.Error("Env = nil, want non-nil when source is set")
	}
	if resp.Env.OS == "" {
		t.Error("Env.OS is empty")
	}
	if resp.Env.Arch == "" {
		t.Error("Env.Arch is empty")
	}
}

func TestFail(t *testing.T) {
	start := time.Now()
	resp := Fail("shell", "inline", "SHELL_ERROR", "command failed", "exit code 1", start)
	if resp.Ok {
		t.Error("Ok = true, want false")
	}
	if resp.Error == nil {
		t.Fatal("Error = nil, want non-nil")
	}
	if resp.Error.Code != "SHELL_ERROR" {
		t.Errorf("Error.Code = %q, want %q", resp.Error.Code, "SHELL_ERROR")
	}
	if resp.Error.Message != "command failed" {
		t.Errorf("Error.Message = %q, want %q", resp.Error.Message, "command failed")
	}
	if resp.Error.Detail != "exit code 1" {
		t.Errorf("Error.Detail = %q, want %q", resp.Error.Detail, "exit code 1")
	}
}

func TestFailEmptyDetail(t *testing.T) {
	start := time.Now()
	resp := Fail("shell", "inline", "ERR", "msg", "", start)
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	s := string(data)
	if strings.Contains(s, `"detail"`) {
		t.Errorf("JSON should omit empty detail, got: %s", s)
	}
}

func TestSuccessEmptySource(t *testing.T) {
	start := time.Now()
	resp := Success("decode", "", "data", start)
	if resp.Env != nil {
		t.Error("Env should be nil when source is empty")
	}
}

func TestFailEmptySource(t *testing.T) {
	start := time.Now()
	resp := Fail("shell", "", "ERR", "msg", "", start)
	if resp.Env != nil {
		t.Error("Env should be nil when source is empty")
	}
}

func TestSuccessJSON(t *testing.T) {
	start := time.Now()
	resp := Success("encode", "inline", map[string]string{"output": "68656c6c6f"}, start)
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if ok, _ := m["ok"].(bool); !ok {
		t.Error("JSON ok = false, want true")
	}
	if typ, _ := m["type"].(string); typ != "encode" {
		t.Errorf("JSON type = %q, want %q", typ, "encode")
	}
	if _, exists := m["env"]; !exists {
		t.Error("JSON missing env field")
	}
}

func TestFailJSON(t *testing.T) {
	start := time.Now()
	resp := Fail("file", "inline", "FILE_NOT_FOUND", "not found", "path: /tmp/x", start)
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if ok, _ := m["ok"].(bool); ok {
		t.Error("JSON ok = true, want false")
	}
	errObj, _ := m["error"].(map[string]interface{})
	if errObj == nil {
		t.Fatal("JSON error field is nil")
	}
	if code, _ := errObj["code"].(string); code != "FILE_NOT_FOUND" {
		t.Errorf("JSON error.code = %q, want %q", code, "FILE_NOT_FOUND")
	}
	if msg, _ := errObj["message"].(string); msg != "not found" {
		t.Errorf("JSON error.message = %q, want %q", msg, "not found")
	}
}

func TestDurationMs(t *testing.T) {
	start := time.Now()
	resp := Success("info", "inline", nil, start)
	if resp.DurMs < 0 {
		t.Errorf("DurMs = %d, want non-negative", resp.DurMs)
	}
}

func TestPrint(t *testing.T) {
	start := time.Now()
	resp := Success("ping", "inline", "pong", start)
	if err := resp.Print(); err != nil {
		t.Errorf("Print() error: %v", err)
	}
}

func TestPrintError(t *testing.T) {
	PrintError("TEST_ERROR", "something went wrong")
}

func TestPrintSuccess(t *testing.T) {
	PrintSuccess("decode", "result data")
}

func TestSuccessWithNilData(t *testing.T) {
	start := time.Now()
	resp := Success("info", "inline", nil, start)
	if resp.Data != nil {
		t.Errorf("Data = %v, want nil", resp.Data)
	}
}

func TestSuccessWithMapData(t *testing.T) {
	start := time.Now()
	data := map[string]interface{}{
		"key1": "value1",
		"key2": float64(42),
	}
	resp := Success("shell", "inline", data, start)
	m, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Data is not a map[string]interface{}")
	}
	if m["key1"] != "value1" {
		t.Errorf("Data[key1] = %v, want %q", m["key1"], "value1")
	}
	if m["key2"] != float64(42) {
		t.Errorf("Data[key2] = %v, want 42", m["key2"])
	}
}

func TestFailWithEmptyCode(t *testing.T) {
	start := time.Now()
	resp := Fail("shell", "inline", "", "message", "", start)
	if resp.Error == nil {
		t.Fatal("Error = nil, want non-nil")
	}
	if resp.Error.Code != "" {
		t.Errorf("Error.Code = %q, want empty", resp.Error.Code)
	}
	if resp.Error.Message != "message" {
		t.Errorf("Error.Message = %q, want %q", resp.Error.Message, "message")
	}
}
