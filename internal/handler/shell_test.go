package handler

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestShellEcho(t *testing.T) {
	h := &ShellHandler{}
	resp := h.Handle("echo hello", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ShellResult)
	if !strings.Contains(result.Stdout, "hello") {
		t.Errorf("expected stdout to contain 'hello', got: %s", result.Stdout)
	}
}

func TestShellJSON(t *testing.T) {
	h := &ShellHandler{}
	payload := ShellPayload{Cmd: "echo json_test"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ShellResult)
	if !strings.Contains(result.Stdout, "json_test") {
		t.Errorf("expected stdout to contain 'json_test', got: %s", result.Stdout)
	}
}

func TestShellFailed(t *testing.T) {
	h := &ShellHandler{}
	resp := h.Handle("false", "")
	if resp.Ok {
		t.Fatal("expected failure for 'false' command")
	}
	if resp.Error.Code != "SHELL_EXIT_NONZERO" {
		t.Errorf("expected error code SHELL_EXIT_NONZERO, got: %s", resp.Error.Code)
	}
}

func TestShellEmptyCmd(t *testing.T) {
	h := &ShellHandler{}
	resp := h.Handle("", "")
	if resp.Ok {
		t.Fatal("expected failure for empty command")
	}
	if resp.Error.Code != "SHELL_EMPTY_CMD" {
		t.Errorf("expected error code SHELL_EMPTY_CMD, got: %s", resp.Error.Code)
	}
}

func TestShellCwd(t *testing.T) {
	h := &ShellHandler{}
	payload := ShellPayload{Cmd: "pwd", Cwd: "/tmp"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ShellResult)
	if !strings.Contains(result.Stdout, "/tmp") {
		t.Errorf("expected stdout to contain '/tmp', got: %s", result.Stdout)
	}
}

func TestShellEnv(t *testing.T) {
	h := &ShellHandler{}
	payload := ShellPayload{
		Cmd: "echo $XXAITK_TEST_VAR",
		Env: map[string]string{"XXAITK_TEST_VAR": "env_value"},
	}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ShellResult)
	if !strings.Contains(result.Stdout, "env_value") {
		t.Errorf("expected stdout to contain 'env_value', got: %s", result.Stdout)
	}
}

func TestShellStdin(t *testing.T) {
	h := &ShellHandler{}
	payload := ShellPayload{Cmd: "cat", Stdin: "piped content"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ShellResult)
	if !strings.Contains(result.Stdout, "piped content") {
		t.Errorf("expected stdout to contain 'piped content', got: %s", result.Stdout)
	}
}

func TestShellMultiLine(t *testing.T) {
	h := &ShellHandler{}
	resp := h.Handle("echo line1; echo line2", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ShellResult)
	if !strings.Contains(result.Stdout, "line1") || !strings.Contains(result.Stdout, "line2") {
		t.Errorf("expected stdout to contain 'line1' and 'line2', got: %s", result.Stdout)
	}
}

func TestShellInvalidJSON(t *testing.T) {
	h := &ShellHandler{}
	resp := h.Handle("{invalid", "")
	if resp.Ok {
		t.Fatalf("expected failure, invalid JSON treated as cmd should fail execution")
	}
}

func TestShellSpecialChars(t *testing.T) {
	h := &ShellHandler{}
	resp := h.Handle("echo 'hello world'", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ShellResult)
	if !strings.Contains(result.Stdout, "hello world") {
		t.Errorf("expected stdout to contain 'hello world', got: %s", result.Stdout)
	}
}