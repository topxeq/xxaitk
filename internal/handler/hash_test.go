package handler

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

func TestHashDefaultSHA256(t *testing.T) {
	h := &HashHandler{}
	resp := h.Handle("hello", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*HashResult)
	if result.Algorithm != "sha256" {
		t.Errorf("expected algorithm sha256, got: %s", result.Algorithm)
	}
	expected := sha256.Sum256([]byte("hello"))
	expectedHex := hex.EncodeToString(expected[:])
	if result.Hash != expectedHex {
		t.Errorf("expected hash %s, got: %s", expectedHex, result.Hash)
	}
}

func TestHashMD5(t *testing.T) {
	h := &HashHandler{}
	payload := HashPayload{Data: "hello", Algo: "md5"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*HashResult)
	if result.Algorithm != "md5" {
		t.Errorf("expected algorithm md5, got: %s", result.Algorithm)
	}
	expected := md5.Sum([]byte("hello"))
	expectedHex := hex.EncodeToString(expected[:])
	if result.Hash != expectedHex {
		t.Errorf("expected hash %s, got: %s", expectedHex, result.Hash)
	}
}

func TestHashSHA1(t *testing.T) {
	h := &HashHandler{}
	payload := HashPayload{Data: "hello", Algo: "sha1"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*HashResult)
	if result.Algorithm != "sha1" {
		t.Errorf("expected algorithm sha1, got: %s", result.Algorithm)
	}
	expected := sha1.Sum([]byte("hello"))
	expectedHex := hex.EncodeToString(expected[:])
	if result.Hash != expectedHex {
		t.Errorf("expected hash %s, got: %s", expectedHex, result.Hash)
	}
}

func TestHashSHA512(t *testing.T) {
	h := &HashHandler{}
	payload := HashPayload{Data: "hello", Algo: "sha512"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*HashResult)
	if result.Algorithm != "sha512" {
		t.Errorf("expected algorithm sha512, got: %s", result.Algorithm)
	}
	expected := sha512.Sum512([]byte("hello"))
	expectedHex := hex.EncodeToString(expected[:])
	if result.Hash != expectedHex {
		t.Errorf("expected hash %s, got: %s", expectedHex, result.Hash)
	}
}

func TestHashUnknownAlgo(t *testing.T) {
	h := &HashHandler{}
	payload := HashPayload{Data: "hello", Algo: "bad"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if resp.Ok {
		t.Fatal("expected failure for unknown algo")
	}
	if resp.Error.Code != "HASH_UNKNOWN_ALGO" {
		t.Errorf("expected error code HASH_UNKNOWN_ALGO, got: %s", resp.Error.Code)
	}
}

func TestHashFile(t *testing.T) {
	f, err := os.CreateTemp("", "hash_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("file content")
	f.Close()

	h := &HashHandler{}
	payload := HashPayload{File: f.Name(), Algo: "sha256"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*HashResult)
	if result.Algorithm != "sha256" {
		t.Errorf("expected algorithm sha256, got: %s", result.Algorithm)
	}
	expected := sha256.Sum256([]byte("file content"))
	expectedHex := hex.EncodeToString(expected[:])
	if result.Hash != expectedHex {
		t.Errorf("expected hash %s, got: %s", expectedHex, result.Hash)
	}
	if result.Input != f.Name() {
		t.Errorf("expected input %s, got: %s", f.Name(), result.Input)
	}
}

func TestHashFileNotFound(t *testing.T) {
	h := &HashHandler{}
	payload := HashPayload{File: "/nonexistent/file.txt", Algo: "sha256"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if resp.Ok {
		t.Fatal("expected failure for nonexistent file")
	}
	if resp.Error.Code != "HASH_FILE_ERROR" {
		t.Errorf("expected error code HASH_FILE_ERROR, got: %s", resp.Error.Code)
	}
}

func TestHashEmptyData(t *testing.T) {
	h := &HashHandler{}
	resp := h.Handle("", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*HashResult)
	if result.Algorithm != "sha256" {
		t.Errorf("expected algorithm sha256, got: %s", result.Algorithm)
	}
	expected := sha256.Sum256([]byte(""))
	expectedHex := hex.EncodeToString(expected[:])
	if result.Hash != expectedHex {
		t.Errorf("expected hash %s, got: %s", expectedHex, result.Hash)
	}
}
