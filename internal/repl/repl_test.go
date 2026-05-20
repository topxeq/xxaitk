package repl

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/topxeq/xxaitk/internal/handler"
)

func init() {
	handler.Register("DECODE", &handler.DecodeHandler{})
	handler.Register("ENCODE", &handler.EncodeHandler{})
	handler.Register("INFO", &handler.InfoHandler{})
	handler.Register("SHELL", &handler.ShellHandler{})
	handler.Register("SCRIPT", &handler.ScriptHandler{})
}

func TestNew(t *testing.T) {
	r := New(false)
	if r == nil {
		t.Error("expected non-nil REPL")
	}
	if r.debug {
		t.Error("expected debug=false")
	}
	r = New(true)
	if !r.debug {
		t.Error("expected debug=true")
	}
}

func TestDotHelp(t *testing.T) {
	r := New(false)
	if r.handleDotCommand(".help") {
		t.Error(".help should not quit")
	}
}

func TestDotPrefixes(t *testing.T) {
	r := New(false)
	if r.handleDotCommand(".prefixes") {
		t.Error(".prefixes should not quit")
	}
}

func TestDotBuiltins(t *testing.T) {
	r := New(false)
	if r.handleDotCommand(".builtins") {
		t.Error(".builtins should not quit")
	}
}

func TestDotDebugOn(t *testing.T) {
	r := New(false)
	if r.handleDotCommand(".debug on") {
		t.Error("should not quit")
	}
	if !r.debug {
		t.Error("expected debug=true")
	}
}

func TestDotDebugOff(t *testing.T) {
	r := New(true)
	if r.handleDotCommand(".debug off") {
		t.Error("should not quit")
	}
	if r.debug {
		t.Error("expected debug=false")
	}
}

func TestDotHistory(t *testing.T) {
	r := New(false)
	r.history = []string{"first", "second"}
	if r.handleDotCommand(".history") {
		t.Error("should not quit")
	}
}

func TestDotQuit(t *testing.T) {
	r := New(false)
	if !r.handleDotCommand(".quit") {
		t.Error(".quit should return true")
	}
}

func TestDotUnknown(t *testing.T) {
	r := New(false)
	if r.handleDotCommand(".unknown") {
		t.Error("unknown should not quit")
	}
}

func TestColonCommandKnown(t *testing.T) {
	New(false).handleColonCommand(":decode hello world")
}

func TestColonCommandUnknown(t *testing.T) {
	New(false).handleColonCommand(":foobar test")
}

func TestColonCommandNoArgs(t *testing.T) {
	New(false).handleColonCommand(":decode")
}

func TestColonCommandEmpty(t *testing.T) {
	New(false).handleColonCommand(":")
}

func TestExecuteScriptInt(t *testing.T)    { New(false).executeScript("1 + 2") }
func TestExecuteScriptString(t *testing.T) { New(false).executeScript(`"hello"`) }
func TestExecuteScriptBool(t *testing.T)   { New(false).executeScript("true") }
func TestExecuteScriptFloat(t *testing.T)  { New(false).executeScript("3.14") }
func TestExecuteScriptList(t *testing.T)   { New(false).executeScript("[1, 2, 3]") }
func TestExecuteScriptMap(t *testing.T)    { New(false).executeScript(`{"key": "val"}`) }
func TestExecuteScriptNil(t *testing.T)    { New(false).executeScript("nil") }
func TestExecuteScriptError(t *testing.T)  { New(false).executeScript("fn(") }
func TestExecuteScriptFn(t *testing.T)     { New(false).executeScript("fn add(a, b) { return a + b }") }

func TestColonCommandWithArgs(t *testing.T) {
	New(false).handleColonCommand(":info os")
}

func TestHistoryPreserved(t *testing.T) {
	r := New(false)
	r.history = append(r.history, "cmd1")
	r.handleDotCommand(".history")
	if len(r.history) != 1 || r.history[0] != "cmd1" {
		t.Errorf("history should be preserved: %v", r.history)
	}
}

func TestDotCommandTrimSpace(t *testing.T) {
	r := New(false)
	if r.handleDotCommand("  .quit  ") {
	}
}

func TestColonCommandWithWhitespace(t *testing.T) {
	New(false).handleColonCommand(":  decode  hello")
}

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestDotHelpOutput(t *testing.T) {
	out := captureOutput(func() { New(false).handleDotCommand(".help") })
	if !strings.Contains(out, ".help") || !strings.Contains(out, ".quit") {
		t.Errorf("help output incomplete: %q", out)
	}
}

func TestDotPrefixesOutput(t *testing.T) {
	out := captureOutput(func() { New(false).handleDotCommand(".prefixes") })
	if !strings.Contains(out, "SHELL") {
		t.Errorf("prefixes output missing SHELL: %q", out)
	}
}

func TestDotHistoryOutput(t *testing.T) {
	r := New(false)
	r.history = []string{"cmd1", "cmd2"}
	out := captureOutput(func() { r.handleDotCommand(".history") })
	if !strings.Contains(out, "cmd1") {
		t.Errorf("history output: %q", out)
	}
}

func TestDotDebugOnOutput(t *testing.T) {
	out := captureOutput(func() { New(false).handleDotCommand(".debug on") })
	if !strings.Contains(out, "enabled") {
		t.Errorf("debug on output: %q", out)
	}
}

func TestDotUnknownOutput(t *testing.T) {
	out := captureOutput(func() { New(false).handleDotCommand(".xyz") })
	if !strings.Contains(out, "Unknown") {
		t.Errorf("unknown cmd output: %q", out)
	}
}

func TestDotBuiltinsOutput(t *testing.T) {
	out := captureOutput(func() { New(false).handleDotCommand(".builtins") })
	if !strings.Contains(out, "str_*") {
		t.Errorf("builtins output: %q", out)
	}
}

func TestDotDebugOffOutput(t *testing.T) {
	out := captureOutput(func() { New(true).handleDotCommand(".debug off") })
	if !strings.Contains(out, "disabled") {
		t.Errorf("debug off output: %q", out)
	}
}
