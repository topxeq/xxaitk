package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func createTestTarGz(t *testing.T, binaryName, content string) []byte {
	t.Helper()
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.tar.gz")

	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	gzw := gzip.NewWriter(f)
	tw := tar.NewWriter(gzw)

	hdr := &tar.Header{
		Name: binaryName,
		Mode: 0755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	gzw.Close()
	f.Close()

	data, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func createTestZip(t *testing.T, binaryName, content string) []byte {
	t.Helper()
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.zip")

	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create(binaryName)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	zw.Close()
	f.Close()

	data, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestCheckUpdateHasUpdate(t *testing.T) {
	redirectTarget := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/releases/latest" {
			w.Header().Set("Location", redirectTarget)
			w.WriteHeader(http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	redirectTarget = server.URL + "/releases/tag/v0.9.0"

	u := New("0.6.0")
	u.BaseURL = server.URL + "/releases"

	latest, hasUpdate, err := u.CheckUpdate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasUpdate {
		t.Error("expected hasUpdate=true")
	}
	if latest != "0.9.0" {
		t.Errorf("latest = %q, want 0.9.0", latest)
	}
}

func TestCheckUpdateNoUpdate(t *testing.T) {
	redirectTarget := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/releases/latest" {
			w.Header().Set("Location", redirectTarget)
			w.WriteHeader(http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	redirectTarget = server.URL + "/releases/tag/v0.6.0"

	u := New("0.6.0")
	u.BaseURL = server.URL + "/releases"

	_, hasUpdate, err := u.CheckUpdate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasUpdate {
		t.Error("expected hasUpdate=false")
	}
}

func TestCheckUpdateNonRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	u := New("0.6.0")
	u.BaseURL = server.URL + "/releases"

	_, _, err := u.CheckUpdate()
	if err == nil {
		t.Error("expected error for non-redirect response")
	}
}

func TestExtractFromTarGz(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping tar.gz test on windows")
	}

	content := "#!/bin/sh\necho hello"
	archiveData := createTestTarGz(t, "aitk", content)

	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "xxaitk_Linux_x86_64.tar.gz")
	if err := os.WriteFile(archivePath, archiveData, 0644); err != nil {
		t.Fatal(err)
	}

	u := New("0.6.0")
	binPath, err := u.extractFromTarGz(archivePath, "aitk")
	if err != nil {
		t.Fatalf("extract error: %v", err)
	}

	data, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatalf("read extracted: %v", err)
	}
	if string(data) != content {
		t.Errorf("extracted content = %q, want %q", string(data), content)
	}
}

func TestExtractFromZip(t *testing.T) {
	content := "echo hello"
	archiveData := createTestZip(t, "aitk.exe", content)

	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "xxaitk_Windows_x86_64.zip")
	if err := os.WriteFile(archivePath, archiveData, 0644); err != nil {
		t.Fatal(err)
	}

	u := New("0.6.0")
	binPath, err := u.extractFromZip(archivePath, "aitk.exe")
	if err != nil {
		t.Fatalf("extract error: %v", err)
	}

	data, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatalf("read extracted: %v", err)
	}
	if string(data) != content {
		t.Errorf("extracted content = %q, want %q", string(data), content)
	}
}

func TestExtractBinaryNotFound(t *testing.T) {
	archiveData := createTestTarGz(t, "other_file", "hello")

	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.tar.gz")
	if err := os.WriteFile(archivePath, archiveData, 0644); err != nil {
		t.Fatal(err)
	}

	u := New("0.6.0")
	_, err := u.extractFromTarGz(archivePath, "aitk")
	if err == nil {
		t.Error("expected error when binary not in archive")
	}
}

func TestDownloadArchive(t *testing.T) {
	content := "#!/bin/sh\necho hello"
	archiveData := createTestTarGz(t, "aitk", content)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/gzip")
		w.Write(archiveData)
	}))
	defer server.Close()

	u := New("0.6.0")
	u.Client = server.Client()

	archivePath, err := u.downloadArchive(server.URL + "/download/v0.9.0/xxaitk_Linux_x86_64.tar.gz")
	if err != nil {
		t.Fatalf("download error: %v", err)
	}
	defer os.RemoveAll(filepath.Dir(archivePath))

	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Error("archive file not created")
	}
}

func TestPlatformArchiveName(t *testing.T) {
	name, err := PlatformArchiveName()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedOS := osMap[runtime.GOOS]
	expectedArch := archMap[runtime.GOARCH]
	expected := fmt.Sprintf("xxaitk_%s_%s", expectedOS, expectedArch)
	if name != expected {
		t.Errorf("PlatformArchiveName() = %q, want %q", name, expected)
	}
}

func TestArchiveExt(t *testing.T) {
	ext := ArchiveExt()
	if runtime.GOOS == "windows" {
		if ext != ".zip" {
			t.Errorf("ArchiveExt() = %q, want .zip on windows", ext)
		}
	} else {
		if ext != ".tar.gz" {
			t.Errorf("ArchiveExt() = %q, want .tar.gz on non-windows", ext)
		}
	}
}

func TestBinaryName(t *testing.T) {
	name := BinaryName()
	if runtime.GOOS == "windows" {
		if name != "aitk.exe" {
			t.Errorf("BinaryName() = %q, want aitk.exe on windows", name)
		}
	} else {
		if name != "aitk" {
			t.Errorf("BinaryName() = %q, want aitk on non-windows", name)
		}
	}
}

func TestFullUpdateFlow(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping full flow on windows (tar.gz)")
	}

	content := "#!/bin/sh\necho updated-binary"
	archiveData := createTestTarGz(t, "aitk", content)

	baseURL := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases/latest":
			w.Header().Set("Location", baseURL+"/releases/tag/v0.9.0")
			w.WriteHeader(http.StatusFound)
		case "/releases/download/v0.9.0/xxaitk_Linux_x86_64.tar.gz":
			w.Header().Set("Content-Type", "application/gzip")
			w.Write(archiveData)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	baseURL = server.URL

	u := New("0.6.0")
	u.BaseURL = server.URL + "/releases"
	u.Client = server.Client()

	latest, hasUpdate, err := u.CheckUpdate()
	if err != nil {
		t.Fatalf("check error: %v", err)
	}
	if !hasUpdate {
		t.Fatal("expected update available")
	}
	if latest != "0.9.0" {
		t.Errorf("latest = %q, want 0.9.0", latest)
	}

	archiveName, _ := PlatformArchiveName()
	downloadURL := fmt.Sprintf("%s/download/v%s/%s%s", u.BaseURL, latest, archiveName, ArchiveExt())

	archivePath, err := u.downloadArchive(downloadURL)
	if err != nil {
		t.Fatalf("download error: %v", err)
	}
	defer os.RemoveAll(filepath.Dir(archivePath))

	binPath, err := u.extractBinary(archivePath)
	if err != nil {
		t.Fatalf("extract error: %v", err)
	}

	data, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatalf("read extracted: %v", err)
	}
	if string(data) != content {
		t.Errorf("extracted = %q, want %q", string(data), content)
	}
}

func TestReplaceBinary(t *testing.T) {
	tmpDir := t.TempDir()

	fakeExe := filepath.Join(tmpDir, "aitk")
	newBin := filepath.Join(tmpDir, "aitk.new")

	if err := os.WriteFile(fakeExe, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newBin, []byte("new"), 0755); err != nil {
		t.Fatal(err)
	}

	u := New("0.6.0")
	u.ExePath = fakeExe

	if err := u.replaceBinary(newBin); err != nil {
		t.Fatalf("replace error: %v", err)
	}

	data, err := os.ReadFile(fakeExe)
	if err != nil {
		t.Fatalf("read replaced: %v", err)
	}
	if string(data) != "new" {
		t.Errorf("replaced content = %q, want 'new'", string(data))
	}
}

func TestCleanupOldBinaryNoop(t *testing.T) {
	if runtime.GOOS != "windows" {
		CleanupOldBinary()
	}
}
