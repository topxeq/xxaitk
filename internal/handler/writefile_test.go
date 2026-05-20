package handler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileCreate(t *testing.T) {
	tmp := filepath.Join(os.TempDir(), "xxaitk_write_test_create.txt")
	defer os.Remove(tmp)

	h := &WriteFileHandler{}
	payload := WriteFilePayload{Path: tmp, Content: "new content", Mode: "create"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*WriteFileResult)
	if result.Mode != "create" {
		t.Errorf("expected mode 'create', got: %s", result.Mode)
	}

	readBytes, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if string(readBytes) != "new content" {
		t.Errorf("expected file content 'new content', got: %s", string(readBytes))
	}
}

func TestWriteFileCreateExists(t *testing.T) {
	f, err := os.CreateTemp("", "xxaitk_write_exists_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	h := &WriteFileHandler{}
	payload := WriteFilePayload{Path: f.Name(), Content: "content", Mode: "create"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if resp.Ok {
		t.Fatal("expected failure when file already exists in create mode")
	}
	if resp.Error.Code != "WRITEFILE_EXISTS" {
		t.Errorf("expected error code WRITEFILE_EXISTS, got: %s", resp.Error.Code)
	}
}

func TestWriteFileOverwrite(t *testing.T) {
	f, err := os.CreateTemp("", "xxaitk_write_overwrite_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("original")
	f.Close()

	h := &WriteFileHandler{}
	payload := WriteFilePayload{Path: f.Name(), Content: "overwritten", Mode: "overwrite"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*WriteFileResult)
	if result.Mode != "overwrite" {
		t.Errorf("expected mode 'overwrite', got: %s", result.Mode)
	}

	readBytes, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if string(readBytes) != "overwritten" {
		t.Errorf("expected file content 'overwritten', got: %s", string(readBytes))
	}
}

func TestWriteFileAppend(t *testing.T) {
	f, err := os.CreateTemp("", "xxaitk_write_append_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("original")
	f.Close()

	h := &WriteFileHandler{}
	payload := WriteFilePayload{Path: f.Name(), Content: " appended", Mode: "append"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*WriteFileResult)
	if result.Mode != "append" {
		t.Errorf("expected mode 'append', got: %s", result.Mode)
	}

	readBytes, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if string(readBytes) != "original appended" {
		t.Errorf("expected file content 'original appended', got: %s", string(readBytes))
	}
}

func TestWriteFileEmptyPath(t *testing.T) {
	h := &WriteFileHandler{}
	payload := WriteFilePayload{Path: "", Content: "content"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if resp.Ok {
		t.Fatal("expected failure for empty path")
	}
	if resp.Error.Code != "WRITEFILE_EMPTY_PATH" {
		t.Errorf("expected error code WRITEFILE_EMPTY_PATH, got: %s", resp.Error.Code)
	}
}

func TestWriteFileEmptyContent(t *testing.T) {
	tmp := filepath.Join(os.TempDir(), "xxaitk_write_empty_content.txt")
	defer os.Remove(tmp)

	h := &WriteFileHandler{}
	payload := WriteFilePayload{Path: tmp, Content: ""}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if resp.Ok {
		t.Fatal("expected failure for empty content")
	}
	if resp.Error.Code != "WRITEFILE_EMPTY_CONTENT" {
		t.Errorf("expected error code WRITEFILE_EMPTY_CONTENT, got: %s", resp.Error.Code)
	}
}

func TestWriteFileInvalidMode(t *testing.T) {
	tmp := filepath.Join(os.TempDir(), "xxaitk_write_invalid_mode.txt")
	defer os.Remove(tmp)

	h := &WriteFileHandler{}
	payload := WriteFilePayload{Path: tmp, Content: "content", Mode: "bad"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if resp.Ok {
		t.Fatal("expected failure for invalid mode")
	}
	if resp.Error.Code != "WRITEFILE_INVALID_MODE" {
		t.Errorf("expected error code WRITEFILE_INVALID_MODE, got: %s", resp.Error.Code)
	}
}

func TestWriteFileMkdir(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "xxaitk_mkdir_test_subdir")
	defer os.RemoveAll(dir)

	tmp := filepath.Join(dir, "nested_file.txt")

	h := &WriteFileHandler{}
	payload := WriteFilePayload{Path: tmp, Content: "nested content", Mode: "create", Mkdir: true}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}

	readBytes, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if string(readBytes) != "nested content" {
		t.Errorf("expected file content 'nested content', got: %s", string(readBytes))
	}
}

func TestWriteFileCreated(t *testing.T) {
	tmp := filepath.Join(os.TempDir(), "xxaitk_write_created_test.txt")
	defer os.Remove(tmp)

	h := &WriteFileHandler{}
	payload := WriteFilePayload{Path: tmp, Content: "new file", Mode: "create"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*WriteFileResult)
	if !result.Created {
		t.Errorf("expected Created=true for new file, got: %v", result.Created)
	}
}