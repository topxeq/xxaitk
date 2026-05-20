package datasource

import (
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"testing"
)

func TestResolveInline(t *testing.T) {
	hexData := hex.EncodeToString([]byte("hello world"))
	rd := Resolve("INLINE", hexData)
	if rd.Err != nil {
		t.Fatalf("unexpected error: %v", rd.Err)
	}
	if rd.Content != "hello world" {
		t.Errorf("Content = %q, want %q", rd.Content, "hello world")
	}
	if rd.Source != "inline" {
		t.Errorf("Source = %q, want %q", rd.Source, "inline")
	}
}

func TestResolveInlineInvalidHex(t *testing.T) {
	rd := Resolve("INLINE", "zzzz")
	if rd.Err == nil {
		t.Fatal("expected error for invalid hex, got nil")
	}
}

func TestResolveInlineEmptyHex(t *testing.T) {
	rd := Resolve("INLINE", "")
	if rd.Err != nil {
		t.Fatalf("unexpected error: %v", rd.Err)
	}
	if rd.Content != "" {
		t.Errorf("Content = %q, want empty string", rd.Content)
	}
}

func TestResolveFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "resolver_test_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := "test file content"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	hexPath := hex.EncodeToString([]byte(tmpFile.Name()))
	rd := Resolve("FILE", hexPath)
	if rd.Err != nil {
		t.Fatalf("unexpected error: %v", rd.Err)
	}
	if rd.Content != content {
		t.Errorf("Content = %q, want %q", rd.Content, content)
	}
	if rd.Source != "file" {
		t.Errorf("Source = %q, want %q", rd.Source, "file")
	}
}

func TestResolveFileNotFound(t *testing.T) {
	hexPath := hex.EncodeToString([]byte("/tmp/nonexistent_file_xyz_12345.dat"))
	rd := Resolve("FILE", hexPath)
	if rd.Err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
}

func TestResolveFileInvalidHex(t *testing.T) {
	rd := Resolve("FILE", "zzzz")
	if rd.Err == nil {
		t.Fatal("expected error for invalid hex, got nil")
	}
}

func TestResolveURL(t *testing.T) {
	hexURL := hex.EncodeToString([]byte("https://httpbin.org/get"))
	rd := Resolve("URL", hexURL)
	if rd.Err != nil {
		if _, ok := rd.Err.(interface{ Temporary() bool }); ok || isNetworkError(rd.Err) {
			t.Skipf("skipping: network error: %v", rd.Err)
		}
		t.Fatalf("unexpected error: %v", rd.Err)
	}
	if rd.Source != "url" {
		t.Errorf("Source = %q, want %q", rd.Source, "url")
	}
	if rd.Content == "" {
		t.Error("Content is empty, expected body from httpbin.org")
	}
}

func TestResolveURLInvalidHex(t *testing.T) {
	rd := Resolve("URL", "zzzz")
	if rd.Err == nil {
		t.Fatal("expected error for invalid hex, got nil")
	}
}

func TestResolveUnknownSource(t *testing.T) {
	rd := Resolve("UNKNOWN", "68656c6c6f")
	if rd.Err == nil {
		t.Fatal("expected error for unknown source type, got nil")
	}
}

func TestResolveJSONSourceFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "resolver_json_test_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := "json source file content"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	result, err := ResolveJSONSource("file", tmpFile.Name(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != content {
		t.Errorf("result = %q, want %q", result, content)
	}
}

func TestResolveJSONSourceFileNotFound(t *testing.T) {
	_, err := ResolveJSONSource("file", "/tmp/nonexistent_file_xyz_99999.dat", "")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestResolveJSONSourceURL(t *testing.T) {
	result, err := ResolveJSONSource("url", "", "https://httpbin.org/get")
	if err != nil {
		t.Skipf("skipping: network error: %v", err)
	}
	if result == "" {
		t.Error("result is empty, expected body from httpbin.org")
	}
}

func TestResolveJSONSourceUnknown(t *testing.T) {
	_, err := ResolveJSONSource("ftp", "", "")
	if err == nil {
		t.Fatal("expected error for unknown source type, got nil")
	}
}

func isNetworkError(err error) bool {
	if _, ok := err.(net.Error); ok {
		return true
	}
	msg := fmt.Sprintf("%v", err)
	return len(msg) > 0 && (contains(msg, "connection") || contains(msg, "timeout") || contains(msg, "DNS") || contains(msg, "lookup"))
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
