package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestNetDownloadBasic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("downloaded content"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	target := filepath.Join(dir, "out.txt")

	h := &NetDownloadHandler{}
	payload := NetDownloadPayload{URL: srv.URL, Path: target}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*NetDownloadResult)
	if result.Status != "downloaded" {
		t.Errorf("status = %q, want downloaded", result.Status)
	}
	if result.Size != int64(len("downloaded content")) {
		t.Errorf("size = %d, want %d", result.Size, len("downloaded content"))
	}
	if result.Sha256 == "" {
		t.Error("expected sha256 hash")
	}

	content, _ := os.ReadFile(target)
	if string(content) != "downloaded content" {
		t.Errorf("file content = %q", string(content))
	}
}

func TestNetDownloadNoURL(t *testing.T) {
	h := &NetDownloadHandler{}
	resp := h.Handle(`{}`, "")
	if resp.Ok {
		t.Error("expected failure for no URL")
	}
	if resp.Error.Code != "NETDOWNLOAD_NO_URL" {
		t.Errorf("code = %q", resp.Error.Code)
	}
}

func TestNetDownloadNoPath(t *testing.T) {
	h := &NetDownloadHandler{}
	resp := h.Handle(`{"url":"http://example.com"}`, "")
	if resp.Ok {
		t.Error("expected failure for no path")
	}
	if resp.Error.Code != "NETDOWNLOAD_NO_PATH" {
		t.Errorf("code = %q", resp.Error.Code)
	}
}

func TestNetDownloadMkdir(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("data"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	target := filepath.Join(dir, "subdir", "out.txt")

	h := &NetDownloadHandler{}
	payload := NetDownloadPayload{URL: srv.URL, Path: target, Mkdir: true}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got: %v", resp.Error)
	}
}

func TestNetDownloadOverwrite(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("new"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	target := filepath.Join(dir, "out.txt")
	os.WriteFile(target, []byte("old"), 0644)

	h := &NetDownloadHandler{}
	payload := NetDownloadPayload{URL: srv.URL, Path: target, Overwrite: true}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got: %v", resp.Error)
	}
}

func TestNetDownloadExistsNoOverwrite(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.txt")
	os.WriteFile(target, []byte("old"), 0644)

	h := &NetDownloadHandler{}
	payload := NetDownloadPayload{URL: "http://example.com", Path: target}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if resp.Ok {
		t.Error("expected failure for existing file without overwrite")
	}
	if resp.Error.Code != "NETDOWNLOAD_EXISTS" {
		t.Errorf("code = %q", resp.Error.Code)
	}
}

func TestNetDownloadWithHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Auth") != "token" {
			w.WriteHeader(403)
			return
		}
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	target := filepath.Join(dir, "out.txt")

	h := &NetDownloadHandler{}
	payload := NetDownloadPayload{URL: srv.URL, Path: target, Headers: map[string]string{"X-Auth": "token"}}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got: %v", resp.Error)
	}
}

func TestNetDownloadHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("error"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	target := filepath.Join(dir, "out.txt")

	h := &NetDownloadHandler{}
	payload := NetDownloadPayload{URL: srv.URL, Path: target, Overwrite: true}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if resp.Ok {
		t.Error("expected failure for 500")
	}
	if resp.Error.Code != "NETDOWNLOAD_HTTP_ERROR" {
		t.Errorf("code = %q", resp.Error.Code)
	}
}

func TestNetDownloadInvalidURL(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.txt")

	h := &NetDownloadHandler{}
	payload := NetDownloadPayload{URL: "http://127.0.0.1:1/fail", Path: target}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if resp.Ok {
		t.Error("expected failure for invalid URL")
	}
}

func TestNetDownloadResume(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeHeader := r.Header.Get("Range")
		if rangeHeader != "" {
			w.WriteHeader(206)
			w.Write([]byte(" resumed"))
		} else {
			w.Write([]byte("initial"))
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	target := filepath.Join(dir, "out.txt")
	os.WriteFile(target, []byte("initial"), 0644)

	h := &NetDownloadHandler{}
	payload := NetDownloadPayload{URL: srv.URL, Path: target, Resume: true}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got: %v", resp.Error)
	}
	result := resp.Data.(*NetDownloadResult)
	if !result.Resumed {
		t.Error("expected resumed=true")
	}
}

func TestNetDownloadInsecure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	target := filepath.Join(dir, "out.txt")

	h := &NetDownloadHandler{}
	payload := NetDownloadPayload{URL: srv.URL, Path: target, Insecure: true}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got: %v", resp.Error)
	}
}

func TestNetDownloadVerifySHA256(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("verify test"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	target := filepath.Join(dir, "out.txt")

	h := &NetDownloadHandler{}
	payload := NetDownloadPayload{URL: srv.URL, Path: target, Verify: "fakehash"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got: %v", resp.Error)
	}
	result := resp.Data.(*NetDownloadResult)
	if result.Sha256 == "" {
		t.Error("expected sha256 hash")
	}
}

func TestNetDownloadParsePayloadEmpty(t *testing.T) {
	h := &NetDownloadHandler{}
	payload := h.parsePayload("")
	if payload.URL != "" {
		t.Errorf("expected empty URL, got %q", payload.URL)
	}
}

func TestNetDownloadParsePayloadNonJSON(t *testing.T) {
	h := &NetDownloadHandler{}
	payload := h.parsePayload("not json")
	if payload.URL != "" {
		t.Errorf("non-JSON should not set URL, got %q", payload.URL)
	}
}
