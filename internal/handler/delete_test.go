package handler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDeleteFile(t *testing.T) {
	f, err := os.CreateTemp("", "delete_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("delete me")
	f.Close()

	h := &DeleteHandler{}
	resp := h.Handle(f.Name(), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*DeleteResult)
	if !result.Deleted {
		t.Error("expected deleted=true")
	}
	if result.WasDir {
		t.Error("expected was_dir=false for file")
	}
	if _, err := os.Stat(f.Name()); !os.IsNotExist(err) {
		t.Error("expected file to be deleted")
	}
}

func TestDeleteDir(t *testing.T) {
	dir, err := os.MkdirTemp("", "delete_dir_test_*")
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(dir, "nested.txt"), []byte("data"), 0644)

	h := &DeleteHandler{}
	payload := DeletePayload{Path: dir, Recursive: true}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*DeleteResult)
	if !result.Deleted {
		t.Error("expected deleted=true")
	}
	if !result.WasDir {
		t.Error("expected was_dir=true for directory")
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("expected directory to be deleted")
	}
}

func TestDeleteNotFound(t *testing.T) {
	h := &DeleteHandler{}
	resp := h.Handle("/nonexistent/path", "")
	if resp.Ok {
		t.Fatal("expected failure for nonexistent path")
	}
	if resp.Error.Code != "DELETE_NOT_FOUND" {
		t.Errorf("expected error code DELETE_NOT_FOUND, got: %s", resp.Error.Code)
	}
}

func TestDeleteEmptyPath(t *testing.T) {
	h := &DeleteHandler{}
	resp := h.Handle("", "")
	if resp.Ok {
		t.Fatal("expected failure for empty path")
	}
	if resp.Error.Code != "DELETE_EMPTY_PATH" {
		t.Errorf("expected error code DELETE_EMPTY_PATH, got: %s", resp.Error.Code)
	}
}

func TestDeleteDirNotRecursive(t *testing.T) {
	dir, err := os.MkdirTemp("", "delete_nonrecursive_test_*")
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(dir, "nested.txt"), []byte("data"), 0644)
	defer os.RemoveAll(dir)

	h := &DeleteHandler{}
	payload := DeletePayload{Path: dir, Recursive: false}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if resp.Ok {
		t.Fatal("expected failure when deleting non-empty dir without recursive")
	}
	if resp.Error.Code != "DELETE_ERROR" {
		t.Errorf("expected error code DELETE_ERROR, got: %s", resp.Error.Code)
	}
}
