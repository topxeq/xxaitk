package scriptlib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestLibDir(t *testing.T) {
	dir := LibDir()
	if dir == "" {
		t.Error("LibDir should not be empty")
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("LibDir = %q, should be absolute", dir)
	}
}

func TestListInstalledEmpty(t *testing.T) {
	names, err := ListInstalled()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = names
}

func TestListInstalledWithFiles(t *testing.T) {
	dir := LibDir()
	os.MkdirAll(dir, 0755)
	f, err := os.CreateTemp(dir, "test_*.xxs")
	if err != nil {
		t.Fatal(err)
	}
	name := filepath.Base(f.Name())
	f.Close()
	defer os.Remove(f.Name())

	names, err := ListInstalled()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, n := range names {
		bn := name[:len(name)-4]
		if n == bn {
			found = true
		}
	}
	if !found {
		t.Errorf("expected to find installed lib, got %v", names)
	}
}

func TestListInstalledNonXxsIgnored(t *testing.T) {
	dir := LibDir()
	os.MkdirAll(dir, 0755)
	f, err := os.CreateTemp(dir, "test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	names, err := ListInstalled()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, n := range names {
		if n == "test_" {
			t.Error(".txt files should not be listed")
		}
	}
}

func TestListInstalledDirIgnored(t *testing.T) {
	dir := LibDir()
	subdir := filepath.Join(dir, "subdir_test")
	os.MkdirAll(subdir, 0755)
	defer os.RemoveAll(subdir)

	names, err := ListInstalled()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, n := range names {
		if n == "subdir_test" {
			t.Error("directories should not be listed")
		}
	}
}

func TestRemoveNotInstalled(t *testing.T) {
	err := Remove("nonexistent_lib_xyz")
	if err == nil {
		t.Error("expected error for non-installed lib")
	}
}

func TestLoadNotInstalled(t *testing.T) {
	_, err := Load("nonexistent_lib_xyz")
	if err == nil {
		t.Error("expected error for non-installed lib")
	}
}

func TestLoadAndRemove(t *testing.T) {
	dir := LibDir()
	os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, "_test_load.xxs")
	os.WriteFile(path, []byte("print(\"hello\")"), 0644)
	defer os.Remove(path)

	content, err := Load("_test_load")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "print(\"hello\")" {
		t.Errorf("content = %q, want print hello", content)
	}

	err = Remove("_test_load")
	if err != nil {
		t.Fatalf("unexpected error removing: %v", err)
	}
}

func TestRegistryURL(t *testing.T) {
	if RegistryURL == "" {
		t.Error("RegistryURL should not be empty")
	}
}

func TestLibraryJSON(t *testing.T) {
	lib := Library{Name: "test", Version: "1.0", Author: "me", Description: "desc", URL: "http://example.com/lib.xxs", Category: "test"}
	data, err := json.Marshal(lib)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var decoded Library
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if decoded.Name != "test" || decoded.Version != "1.0" {
		t.Errorf("decoded = %+v", decoded)
	}
}

func TestRegistryJSON(t *testing.T) {
	reg := Registry{Libraries: []Library{
		{Name: "test", Version: "1.0", Author: "me", Description: "desc", URL: "http://example.com/lib.xxs", Category: "test"},
	}}
	data, err := json.Marshal(reg)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var decoded Registry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(decoded.Libraries) != 1 || decoded.Libraries[0].Name != "test" {
		t.Errorf("unexpected registry: %+v", decoded)
	}
}

func TestListRemoteWithMockServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reg := Registry{Libraries: []Library{
			{Name: "mock-lib", Version: "1.0", Author: "test", Description: "mock lib", URL: "http://example.com/lib.xxs", Category: "test"},
		}}
		data, _ := json.Marshal(reg)
		w.Write(data)
	}))
	defer srv.Close()

	oldURL := RegistryURL
	RegistryURL = srv.URL
	defer func() { RegistryURL = oldURL }()

	reg, err := ListRemote()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reg.Libraries) != 1 {
		t.Fatalf("expected 1 library, got %d", len(reg.Libraries))
	}
	if reg.Libraries[0].Name != "mock-lib" {
		t.Errorf("name = %q, want mock-lib", reg.Libraries[0].Name)
	}
}

func TestListRemoteError(t *testing.T) {
	oldURL := RegistryURL
	RegistryURL = "http://127.0.0.1:1/fail"
	defer func() { RegistryURL = oldURL }()

	_, err := ListRemote()
	if err == nil {
		t.Error("expected error for unreachable URL")
	}
}

func TestListRemoteInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	oldURL := RegistryURL
	RegistryURL = srv.URL
	defer func() { RegistryURL = oldURL }()

	_, err := ListRemote()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestListRemoteEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reg := Registry{Libraries: []Library{}}
		data, _ := json.Marshal(reg)
		w.Write(data)
	}))
	defer srv.Close()

	oldURL := RegistryURL
	RegistryURL = srv.URL
	defer func() { RegistryURL = oldURL }()

	reg, err := ListRemote()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reg.Libraries) != 0 {
		t.Errorf("expected 0 libraries, got %d", len(reg.Libraries))
	}
}

func TestGetWithMockServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/registry" {
			reg := Registry{Libraries: []Library{
				{Name: "test-get-lib", Version: "1.0", Author: "test", Description: "test", URL: "http://" + r.Host + "/download", Category: "test"},
			}}
			data, _ := json.Marshal(reg)
			w.Write(data)
		} else {
			w.Write([]byte("print(\"from remote\")"))
		}
	}))
	defer srv.Close()

	oldURL := RegistryURL
	RegistryURL = srv.URL + "/registry"
	defer func() { RegistryURL = oldURL }()

	err := Get("test-get-lib")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer Remove("test-get-lib")

	content, err := Load("test-get-lib")
	if err != nil {
		t.Fatalf("unexpected error loading: %v", err)
	}
	if content != "print(\"from remote\")" {
		t.Errorf("content = %q", content)
	}
}

func TestGetNotFoundInRegistry(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reg := Registry{Libraries: []Library{}}
		data, _ := json.Marshal(reg)
		w.Write(data)
	}))
	defer srv.Close()

	oldURL := RegistryURL
	RegistryURL = srv.URL
	defer func() { RegistryURL = oldURL }()

	err := Get("nonexistent")
	if err == nil {
		t.Error("expected error for library not found in registry")
	}
}

func TestGetRegistryUnreachable(t *testing.T) {
	oldURL := RegistryURL
	RegistryURL = "http://127.0.0.1:1/fail"
	defer func() { RegistryURL = oldURL }()

	err := Get("anything")
	if err == nil {
		t.Error("expected error for unreachable registry")
	}
}

func TestGetDownloadFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/registry" {
			reg := Registry{Libraries: []Library{
				{Name: "test-dl-fail", Version: "1.0", Author: "test", Description: "test", URL: "http://127.0.0.1:1/fail.xxs", Category: "test"},
			}}
			data, _ := json.Marshal(reg)
			w.Write(data)
		}
	}))
	defer srv.Close()

	oldURL := RegistryURL
	RegistryURL = srv.URL + "/registry"
	defer func() { RegistryURL = oldURL }()

	err := Get("test-dl-fail")
	if err == nil {
		t.Error("expected error for download failure")
	}
}

func TestLoadWriteAndRemove(t *testing.T) {
	dir := LibDir()
	os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, "_test_cycle.xxs")
	os.WriteFile(path, []byte("// cycle test"), 0644)

	content, err := Load("_test_cycle")
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if content != "// cycle test" {
		t.Errorf("content = %q", content)
	}

	err = Remove("_test_cycle")
	if err != nil {
		t.Fatalf("remove error: %v", err)
	}

	_, err = Load("_test_cycle")
	if err == nil {
		t.Error("expected error after removal")
	}
}
