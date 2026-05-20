package handler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestListDirBasic(t *testing.T) {
	dir := t.TempDir()
	os.Create(filepath.Join(dir, "a.txt"))
	os.Create(filepath.Join(dir, "b.txt"))

	h := &ListDirHandler{}
	resp := h.Handle(dir, "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ListDirResult)
	if result.Count != 2 {
		t.Errorf("expected count 2, got: %d", result.Count)
	}
	if result.Path != dir {
		t.Errorf("expected path %s, got: %s", dir, result.Path)
	}
}

func TestListDirNotFound(t *testing.T) {
	h := &ListDirHandler{}
	resp := h.Handle("/nonexistent/dir/path", "")
	if resp.Ok {
		t.Fatal("expected failure for nonexistent directory")
	}
	if resp.Error.Code != "LISTDIR_NOT_FOUND" {
		t.Errorf("expected error code LISTDIR_NOT_FOUND, got: %s", resp.Error.Code)
	}
}

func TestListDirNotDir(t *testing.T) {
	f, err := os.CreateTemp("", "listdir_notdir_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	h := &ListDirHandler{}
	resp := h.Handle(f.Name(), "")
	if resp.Ok {
		t.Fatal("expected failure for file path")
	}
	if resp.Error.Code != "LISTDIR_NOT_DIR" {
		t.Errorf("expected error code LISTDIR_NOT_DIR, got: %s", resp.Error.Code)
	}
}

func TestListDirEmptyPath(t *testing.T) {
	h := &ListDirHandler{}
	resp := h.Handle("", "")
	if resp.Ok {
		t.Fatal("expected failure for empty path")
	}
	if resp.Error.Code != "LISTDIR_EMPTY_PATH" {
		t.Errorf("expected error code LISTDIR_EMPTY_PATH, got: %s", resp.Error.Code)
	}
}

func TestListDirRecursive(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	os.Mkdir(sub, 0755)
	os.Create(filepath.Join(dir, "top.txt"))
	os.Create(filepath.Join(sub, "nested.txt"))

	h := &ListDirHandler{}
	payload := ListDirPayload{Path: dir, Recursive: true}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ListDirResult)
	if result.Count != 3 {
		t.Errorf("expected count 3 (sub, top.txt, nested.txt), got: %d", result.Count)
	}
}

func TestListDirPattern(t *testing.T) {
	dir := t.TempDir()
	os.Create(filepath.Join(dir, "a.txt"))
	os.Create(filepath.Join(dir, "b.txt"))
	os.Create(filepath.Join(dir, "c.log"))

	h := &ListDirHandler{}
	payload := ListDirPayload{Path: dir, Pattern: "*.txt"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ListDirResult)
	if result.Count != 2 {
		t.Errorf("expected count 2 for *.txt, got: %d", result.Count)
	}
	for _, e := range result.Entries {
		if filepath.Ext(e.Name) != ".txt" {
			t.Errorf("unexpected entry: %s", e.Name)
		}
	}
}

func TestListDirShowHidden(t *testing.T) {
	dir := t.TempDir()
	os.Create(filepath.Join(dir, ".hidden"))
	os.Create(filepath.Join(dir, "visible.txt"))

	h := &ListDirHandler{}
	payload := ListDirPayload{Path: dir, ShowHidden: true}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ListDirResult)
	if result.Count != 2 {
		t.Errorf("expected count 2 with show_hidden, got: %d", result.Count)
	}
}

func TestListDirJSON(t *testing.T) {
	dir := t.TempDir()
	os.Create(filepath.Join(dir, "file.txt"))

	h := &ListDirHandler{}
	payload := ListDirPayload{Path: dir}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ListDirResult)
	if result.Count != 1 {
		t.Errorf("expected count 1, got: %d", result.Count)
	}
}

func TestListDirHiddenNotShown(t *testing.T) {
	dir := t.TempDir()
	os.Create(filepath.Join(dir, ".dotfile"))
	os.Create(filepath.Join(dir, "normal.txt"))

	h := &ListDirHandler{}
	resp := h.Handle(dir, "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ListDirResult)
	if result.Count != 1 {
		t.Errorf("expected count 1 (hidden files not shown), got: %d", result.Count)
	}
	if len(result.Entries) > 0 && result.Entries[0].Name == ".dotfile" {
		t.Error("hidden file should not appear by default")
	}
}
