package handler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestArchivePackZip(t *testing.T) {
	dir, err := os.MkdirTemp("", "archive_pack_zip_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("hello zip"), 0644)

	target := filepath.Join(dir, "out.zip")
	h := &ArchiveHandler{}
	payload := ArchivePayload{Action: "pack", Format: "zip", Target: target, Files: []string{testFile}}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ArchiveResult)
	if result.Action != "pack" {
		t.Errorf("expected action pack, got: %s", result.Action)
	}
	if result.Format != "zip" {
		t.Errorf("expected format zip, got: %s", result.Format)
	}
	if result.Count < 1 {
		t.Errorf("expected at least 1 file, got: %d", result.Count)
	}
	if _, err := os.Stat(target); os.IsNotExist(err) {
		t.Error("expected zip file to exist")
	}
}

func TestArchivePackTar(t *testing.T) {
	dir, err := os.MkdirTemp("", "archive_pack_tar_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("hello tar"), 0644)

	target := filepath.Join(dir, "out.tar")
	h := &ArchiveHandler{}
	payload := ArchivePayload{Action: "pack", Format: "tar", Target: target, Files: []string{testFile}}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ArchiveResult)
	if result.Format != "tar" {
		t.Errorf("expected format tar, got: %s", result.Format)
	}
	if _, err := os.Stat(target); os.IsNotExist(err) {
		t.Error("expected tar file to exist")
	}
}

func TestArchivePackTarGz(t *testing.T) {
	dir, err := os.MkdirTemp("", "archive_pack_targz_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("hello tar.gz"), 0644)

	target := filepath.Join(dir, "out.tar.gz")
	h := &ArchiveHandler{}
	payload := ArchivePayload{Action: "pack", Format: "tar.gz", Target: target, Files: []string{testFile}}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ArchiveResult)
	if result.Format != "tar.gz" {
		t.Errorf("expected format tar.gz, got: %s", result.Format)
	}
	if _, err := os.Stat(target); os.IsNotExist(err) {
		t.Error("expected tar.gz file to exist")
	}
}

func TestArchiveListZip(t *testing.T) {
	dir, err := os.MkdirTemp("", "archive_list_zip_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("list me"), 0644)

	target := filepath.Join(dir, "out.zip")
	h := &ArchiveHandler{}
	payload := ArchivePayload{Action: "pack", Format: "zip", Target: target, Files: []string{testFile}}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("pack failed: %v", resp.Error)
	}

	listPayload := ArchivePayload{Action: "list", Source: target}
	listData, _ := json.Marshal(listPayload)
	resp = h.Handle(string(listData), "")
	if !resp.Ok {
		t.Fatalf("expected ok for list, got error: %v", resp.Error)
	}
	result := resp.Data.(*ArchiveResult)
	if result.Action != "list" {
		t.Errorf("expected action list, got: %s", result.Action)
	}
	if result.Count < 1 {
		t.Errorf("expected at least 1 file in listing, got: %d", result.Count)
	}
}

func TestArchiveUnpackZip(t *testing.T) {
	dir, err := os.MkdirTemp("", "archive_unpack_zip_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("unpack me"), 0644)

	target := filepath.Join(dir, "out.zip")
	h := &ArchiveHandler{}
	payload := ArchivePayload{Action: "pack", Format: "zip", Target: target, Files: []string{testFile}}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("pack failed: %v", resp.Error)
	}

	unpackDir := filepath.Join(dir, "unpacked")
	os.MkdirAll(unpackDir, 0755)
	unpackPayload := ArchivePayload{Action: "unpack", Source: target, Target: unpackDir}
	unpackData, _ := json.Marshal(unpackPayload)
	resp = h.Handle(string(unpackData), "")
	if !resp.Ok {
		t.Fatalf("expected ok for unpack, got error: %v", resp.Error)
	}
	result := resp.Data.(*ArchiveResult)
	if result.Action != "unpack" {
		t.Errorf("expected action unpack, got: %s", result.Action)
	}
	if result.Count < 1 {
		t.Errorf("expected at least 1 extracted file, got: %d", result.Count)
	}
}

func TestArchiveUnpackTarGz(t *testing.T) {
	dir, err := os.MkdirTemp("", "archive_unpack_targz_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("unpack targz"), 0644)

	target := filepath.Join(dir, "out.tar.gz")
	h := &ArchiveHandler{}
	payload := ArchivePayload{Action: "pack", Format: "tar.gz", Target: target, Files: []string{testFile}}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("pack failed: %v", resp.Error)
	}

	unpackDir := filepath.Join(dir, "unpacked")
	os.MkdirAll(unpackDir, 0755)
	unpackPayload := ArchivePayload{Action: "unpack", Source: target, Target: unpackDir}
	unpackData, _ := json.Marshal(unpackPayload)
	resp = h.Handle(string(unpackData), "")
	if !resp.Ok {
		t.Fatalf("expected ok for unpack, got error: %v", resp.Error)
	}
	result := resp.Data.(*ArchiveResult)
	if result.Count < 1 {
		t.Errorf("expected at least 1 extracted file, got: %d", result.Count)
	}
}

func TestArchiveNoTarget(t *testing.T) {
	h := &ArchiveHandler{}
	payload := ArchivePayload{Action: "pack", Format: "zip", Files: []string{"/tmp/fake.txt"}}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if resp.Ok {
		t.Fatal("expected failure for pack without target")
	}
	if resp.Error.Code != "ARCHIVE_NO_TARGET" {
		t.Errorf("expected error code ARCHIVE_NO_TARGET, got: %s", resp.Error.Code)
	}
}

func TestArchiveNoSource(t *testing.T) {
	h := &ArchiveHandler{}
	payload := ArchivePayload{Action: "unpack", Format: "zip"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if resp.Ok {
		t.Fatal("expected failure for unpack without source")
	}
	if resp.Error.Code != "ARCHIVE_NO_SOURCE" {
		t.Errorf("expected error code ARCHIVE_NO_SOURCE, got: %s", resp.Error.Code)
	}
}

func TestArchiveUnknownAction(t *testing.T) {
	h := &ArchiveHandler{}
	payload := ArchivePayload{Action: "bad"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if resp.Ok {
		t.Fatal("expected failure for unknown action")
	}
	if resp.Error.Code != "ARCHIVE_UNKNOWN_ACTION" {
		t.Errorf("expected error code ARCHIVE_UNKNOWN_ACTION, got: %s", resp.Error.Code)
	}
}

func TestArchiveUnsupportedFormat(t *testing.T) {
	dir, err := os.MkdirTemp("", "archive_unsupported_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	target := filepath.Join(dir, "out.rar")
	h := &ArchiveHandler{}
	payload := ArchivePayload{Action: "pack", Format: "rar", Target: target, Files: []string{"/tmp/fake.txt"}}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if resp.Ok {
		t.Fatal("expected failure for unsupported format")
	}
	if resp.Error.Code != "ARCHIVE_UNSUPPORTED_FORMAT" {
		t.Errorf("expected error code ARCHIVE_UNSUPPORTED_FORMAT, got: %s", resp.Error.Code)
	}
}

func TestArchiveListTar(t *testing.T) {
	dir, err := os.MkdirTemp("", "archive_list_tar_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("list tar"), 0644)

	target := filepath.Join(dir, "out.tar")
	h := &ArchiveHandler{}
	payload := ArchivePayload{Action: "pack", Format: "tar", Target: target, Files: []string{testFile}}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("pack failed: %v", resp.Error)
	}

	listPayload := ArchivePayload{Action: "list", Source: target, Format: "tar"}
	listData, _ := json.Marshal(listPayload)
	resp = h.Handle(string(listData), "")
	if !resp.Ok {
		t.Fatalf("expected ok for list tar, got error: %v", resp.Error)
	}
	result := resp.Data.(*ArchiveResult)
	if result.Count < 1 {
		t.Errorf("expected at least 1 file in tar listing, got: %d", result.Count)
	}
}

func TestArchiveListTarGz(t *testing.T) {
	dir, err := os.MkdirTemp("", "archive_list_targz_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("list tar.gz"), 0644)

	target := filepath.Join(dir, "out.tar.gz")
	h := &ArchiveHandler{}
	payload := ArchivePayload{Action: "pack", Format: "tar.gz", Target: target, Files: []string{testFile}}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("pack failed: %v", resp.Error)
	}

	listPayload := ArchivePayload{Action: "list", Source: target, Format: "tar.gz"}
	listData, _ := json.Marshal(listPayload)
	resp = h.Handle(string(listData), "")
	if !resp.Ok {
		t.Fatalf("expected ok for list tar.gz, got error: %v", resp.Error)
	}
	result := resp.Data.(*ArchiveResult)
	if result.Count < 1 {
		t.Errorf("expected at least 1 file in tar.gz listing, got: %d", result.Count)
	}
}

func TestArchiveUnpackTar(t *testing.T) {
	dir, err := os.MkdirTemp("", "archive_unpack_tar_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("unpack tar"), 0644)

	target := filepath.Join(dir, "out.tar")
	h := &ArchiveHandler{}
	payload := ArchivePayload{Action: "pack", Format: "tar", Target: target, Files: []string{testFile}}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("pack failed: %v", resp.Error)
	}

	unpackDir := filepath.Join(dir, "unpacked")
	os.MkdirAll(unpackDir, 0755)
	unpackPayload := ArchivePayload{Action: "unpack", Source: target, Target: unpackDir, Format: "tar"}
	unpackData, _ := json.Marshal(unpackPayload)
	resp = h.Handle(string(unpackData), "")
	if !resp.Ok {
		t.Fatalf("expected ok for unpack tar, got error: %v", resp.Error)
	}
	result := resp.Data.(*ArchiveResult)
	if result.Count < 1 {
		t.Errorf("expected at least 1 extracted file, got: %d", result.Count)
	}
}

func TestArchivePackWithDir(t *testing.T) {
	dir, err := os.MkdirTemp("", "archive_packdir_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	subdir := filepath.Join(dir, "subdir")
	os.MkdirAll(subdir, 0755)
	os.WriteFile(filepath.Join(subdir, "a.txt"), []byte("aaa"), 0644)
	os.WriteFile(filepath.Join(subdir, "b.txt"), []byte("bbb"), 0644)

	target := filepath.Join(dir, "out.zip")
	h := &ArchiveHandler{}
	payload := ArchivePayload{Action: "pack", Format: "zip", Target: target, Dir: subdir}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ArchiveResult)
	if result.Count < 2 {
		t.Errorf("expected at least 2 files, got: %d", result.Count)
	}
}

func TestArchiveAutoDetectFormat(t *testing.T) {
	h := &ArchiveHandler{}
	payload := h.parsePayload(`{"action":"list","source":"/tmp/test.zip"}`)
	if payload.Format != "" {
		_ = payload
	}
}

func TestArchivePackTgz(t *testing.T) {
	dir, err := os.MkdirTemp("", "archive_pack_tgz_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("hello tgz"), 0644)

	target := filepath.Join(dir, "out.tgz")
	h := &ArchiveHandler{}
	payload := ArchivePayload{Action: "pack", Format: "tgz", Target: target, Files: []string{testFile}}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestArchiveListNoSource(t *testing.T) {
	h := &ArchiveHandler{}
	resp := h.Handle(`{"action":"list"}`, "")
	if resp.Ok {
		t.Fatal("expected failure for list without source")
	}
}

func TestArchivePackNoFiles(t *testing.T) {
	h := &ArchiveHandler{}
	resp := h.Handle(`{"action":"pack","format":"zip","target":"/tmp/out.zip"}`, "")
	if resp.Ok {
		t.Fatal("expected failure for pack without files")
	}
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"test.zip", "zip"},
		{"test.tar.gz", "tar.gz"},
		{"test.tgz", "tar.gz"},
		{"test.tar", "tar"},
		{"test.rar", "zip"},
		{"", "zip"},
	}
	for _, tt := range tests {
		got := detectFormat(tt.path)
		if got != tt.want {
			t.Errorf("detectFormat(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestArchiveUnpackUnsupportedFormat(t *testing.T) {
	dir, _ := os.MkdirTemp("", "archive_unpack_unsupported_*")
	defer os.RemoveAll(dir)

	f, _ := os.CreateTemp(dir, "test.txt")
	f.WriteString("test")
	f.Close()

	target := filepath.Join(dir, "out.zip")
	h := &ArchiveHandler{}
	payload := ArchivePayload{Action: "pack", Format: "zip", Target: target, Files: []string{f.Name()}}
	data, _ := json.Marshal(payload)
	h.Handle(string(data), "")

	unpackPayload := ArchivePayload{Action: "unpack", Source: target, Format: "rar", Target: filepath.Join(dir, "unpacked")}
	unpackData, _ := json.Marshal(unpackPayload)
	resp := h.Handle(string(unpackData), "")
	if resp.Ok {
		t.Fatal("expected failure for unpack with unsupported format")
	}
}

func TestArchiveListNoFormatAutoDetect(t *testing.T) {
	dir, _ := os.MkdirTemp("", "archive_list_autodetect_*")
	defer os.RemoveAll(dir)

	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("content"), 0644)

	target := filepath.Join(dir, "data.zip")
	h := &ArchiveHandler{}
	payload := ArchivePayload{Action: "pack", Format: "zip", Target: target, Files: []string{testFile}}
	data, _ := json.Marshal(payload)
	h.Handle(string(data), "")

	listPayload := ArchivePayload{Action: "list", Source: target}
	listData, _ := json.Marshal(listPayload)
	resp := h.Handle(string(listData), "")
	if !resp.Ok {
		t.Fatalf("expected ok for auto-detect format: %v", resp.Error)
	}
}
