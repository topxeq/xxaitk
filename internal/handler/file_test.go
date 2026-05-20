package handler

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"
)

func TestFileRead(t *testing.T) {
	f, err := os.CreateTemp("", "file_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("hello file content")
	f.Close()

	h := &FileHandler{}
	resp := h.Handle(f.Name(), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*FileResult)
	if result.Content != "hello file content" {
		t.Errorf("expected content 'hello file content', got: %s", result.Content)
	}
	if result.Encoding != "utf8" {
		t.Errorf("expected encoding 'utf8', got: %s", result.Encoding)
	}
}

func TestFileReadJSON(t *testing.T) {
	f, err := os.CreateTemp("", "file_json_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("json content")
	f.Close()

	h := &FileHandler{}
	payload := FilePayload{Path: f.Name()}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*FileResult)
	if result.Content != "json content" {
		t.Errorf("expected content 'json content', got: %s", result.Content)
	}
}

func TestFileNotFound(t *testing.T) {
	h := &FileHandler{}
	resp := h.Handle("/nonexistent/file.txt", "")
	if resp.Ok {
		t.Fatal("expected failure for nonexistent file")
	}
	if resp.Error.Code != "FILE_NOT_FOUND" {
		t.Errorf("expected error code FILE_NOT_FOUND, got: %s", resp.Error.Code)
	}
}

func TestFileEmptyPath(t *testing.T) {
	h := &FileHandler{}
	resp := h.Handle("", "")
	if resp.Ok {
		t.Fatal("expected failure for empty path")
	}
	if resp.Error.Code != "FILE_EMPTY_PATH" {
		t.Errorf("expected error code FILE_EMPTY_PATH, got: %s", resp.Error.Code)
	}
}

func TestFileIsDir(t *testing.T) {
	h := &FileHandler{}
	dir := os.TempDir()
	resp := h.Handle(dir, "")
	if resp.Ok {
		t.Fatal("expected failure for directory path")
	}
	if resp.Error.Code != "FILE_IS_DIR" {
		t.Errorf("expected error code FILE_IS_DIR, got: %s", resp.Error.Code)
	}
}

func TestFileReadBinary(t *testing.T) {
	f, err := os.CreateTemp("", "file_binary_test_*.bin")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	raw := []byte{0x00, 0x01, 0x02, 0xFF}
	f.Write(raw)
	f.Close()

	h := &FileHandler{}
	payload := FilePayload{Path: f.Name(), Encoding: "binary"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*FileResult)
	expectedB64 := base64.StdEncoding.EncodeToString(raw)
	if result.ContentB64 != expectedB64 {
		t.Errorf("expected ContentB64 %s, got: %s", expectedB64, result.ContentB64)
	}
	if result.Encoding != "binary" {
		t.Errorf("expected encoding 'binary', got: %s", result.Encoding)
	}
}

func TestFileReadWithOffset(t *testing.T) {
	f, err := os.CreateTemp("", "file_offset_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("ABCDEFGHIJ")
	f.Close()

	h := &FileHandler{}
	payload := FilePayload{Path: f.Name(), Offset: 5, Limit: 5}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*FileResult)
	if result.Content != "FGHIJ" {
		t.Errorf("expected content 'FGHIJ', got: %s", result.Content)
	}
}

func TestFileReadWithLimit(t *testing.T) {
	f, err := os.CreateTemp("", "file_limit_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("ABCDEFGHIJ")
	f.Close()

	h := &FileHandler{}
	payload := FilePayload{Path: f.Name(), Limit: 10}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*FileResult)
	if result.Content != "ABCDEFGHIJ" {
		t.Errorf("expected content 'ABCDEFGHIJ', got: %s", result.Content)
	}
	if result.Size != 10 {
		t.Errorf("expected size 10, got: %d", result.Size)
	}
}

func TestFileReadOffsetAndLimit(t *testing.T) {
	f, err := os.CreateTemp("", "file_offsetlimit_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("ABCDEFGHIJ")
	f.Close()

	h := &FileHandler{}
	payload := FilePayload{Path: f.Name(), Offset: 5, Limit: 3}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*FileResult)
	if result.Content != "FGH" {
		t.Errorf("expected content 'FGH', got: %s", result.Content)
	}
}

func TestFileDefaultEncoding(t *testing.T) {
	f, err := os.CreateTemp("", "file_encoding_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("test")
	f.Close()

	h := &FileHandler{}
	resp := h.Handle(f.Name(), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*FileResult)
	if result.Encoding != "utf8" {
		t.Errorf("expected default encoding 'utf8', got: %s", result.Encoding)
	}
}