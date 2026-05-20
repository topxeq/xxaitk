package dispatcher

import (
	"encoding/hex"
	"os"
	"testing"

	"github.com/topxeq/xxaitk/internal/handler"
)

func init() {
	handler.Register("SHELL", &handler.ShellHandler{})
	handler.Register("FILE", &handler.FileHandler{})
	handler.Register("DECODE", &handler.DecodeHandler{})
	handler.Register("ENCODE", &handler.EncodeHandler{})
}

func TestDispatchShell(t *testing.T) {
	d := New(false)
	hexData := hex.EncodeToString([]byte("echo hello"))
	d.Dispatch("SHELL_" + hexData)
}

func TestDispatchFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "dispatch_test_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := "dispatch file test content"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	d := New(false)
	hexPath := hex.EncodeToString([]byte(tmpFile.Name()))
	d.Dispatch("FILE_" + hexPath)
}

func TestDispatchDecode(t *testing.T) {
	d := New(false)
	hexData := hex.EncodeToString([]byte("hello world"))
	d.Dispatch("DECODE_" + hexData)
}

func TestDispatchUnknownPrefix(t *testing.T) {
	d := New(false)
	hexData := hex.EncodeToString([]byte("hello"))
	d.Dispatch("FOOBAR_" + hexData)
}

func TestDispatchInvalidHex(t *testing.T) {
	d := New(false)
	d.Dispatch("SHELL_zzzz")
}

func TestDispatchEmptyArg(t *testing.T) {
	d := New(false)
	d.Dispatch("")
}

func TestDispatchNoUnderscore(t *testing.T) {
	d := New(false)
	d.Dispatch("SHELL")
}

func TestDispatchDebug(t *testing.T) {
	d := New(true)
	hexData := hex.EncodeToString([]byte("echo debug"))
	d.Dispatch("SHELL_" + hexData)
}

func TestDispatchURLSource(t *testing.T) {
	d := New(false)
	hexURL := hex.EncodeToString([]byte("http://httpbin.org/get"))
	d.Dispatch("DECODE_URL_" + hexURL)
}

func TestDispatchHandlerNotRegistered(t *testing.T) {
	d := New(false)
	hexData := hex.EncodeToString([]byte("test"))
	d.Dispatch("INFO_" + hexData)
}

func TestDispatchFileSource(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "dispatch_file_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("file source test")
	tmpFile.Close()

	d := New(false)
	hexPath := hex.EncodeToString([]byte(tmpFile.Name()))
	d.Dispatch("FILE_FILE_" + hexPath)
}

func TestDispatchSourceResolveError(t *testing.T) {
	d := New(false)
	hexPath := hex.EncodeToString([]byte("/nonexistent/path/file.txt"))
	d.Dispatch("FILE_FILE_" + hexPath)
}
