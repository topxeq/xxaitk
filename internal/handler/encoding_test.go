package handler

import (
	"encoding/base64"
	"net/url"
	"strings"
	"testing"

	"github.com/topxeq/xxaitk/internal/hexcodec"
)

func TestDecode(t *testing.T) {
	h := &DecodeHandler{}
	resp := h.Handle("hello world", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*EncodeResult)
	if result.Output != "hello world" {
		t.Errorf("expected output 'hello world', got: %s", result.Output)
	}
	if result.Format != "plaintext" {
		t.Errorf("expected format 'plaintext', got: %s", result.Format)
	}
}

func TestEncode(t *testing.T) {
	h := &EncodeHandler{}
	resp := h.Handle("hello", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*EncodeResult)
	expected := hexcodec.EncodeString("hello")
	if result.Output != expected {
		t.Errorf("expected output %s, got: %s", expected, result.Output)
	}
	if result.Format != "hex" {
		t.Errorf("expected format 'hex', got: %s", result.Format)
	}
	if !strings.Contains(result.Output, "68656c6c6f") {
		t.Errorf("expected hex output to contain '68656c6c6f', got: %s", result.Output)
	}
}

func TestB64Enc(t *testing.T) {
	h := &B64EncHandler{}
	resp := h.Handle("hello", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*EncodeResult)
	expected := base64.StdEncoding.EncodeToString([]byte("hello"))
	if result.Output != expected {
		t.Errorf("expected output %s, got: %s", expected, result.Output)
	}
	if result.Format != "base64" {
		t.Errorf("expected format 'base64', got: %s", result.Format)
	}
}

func TestB64Dec(t *testing.T) {
	h := &B64DecHandler{}
	encoded := base64.StdEncoding.EncodeToString([]byte("hello"))
	resp := h.Handle(encoded, "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*EncodeResult)
	expected := hexcodec.EncodeString("hello")
	if result.Output != expected {
		t.Errorf("expected output %s, got: %s", expected, result.Output)
	}
	if result.Format != "hex" {
		t.Errorf("expected format 'hex', got: %s", result.Format)
	}
}

func TestB64DecInvalid(t *testing.T) {
	h := &B64DecHandler{}
	resp := h.Handle("!!!invalid-base64!!!", "")
	if resp.Ok {
		t.Fatal("expected failure for invalid base64")
	}
	if resp.Error.Code != "B64DEC_ERROR" {
		t.Errorf("expected error code B64DEC_ERROR, got: %s", resp.Error.Code)
	}
}

func TestURLEnc(t *testing.T) {
	h := &URLEncHandler{}
	resp := h.Handle("hello world", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*EncodeResult)
	expected := url.QueryEscape("hello world")
	if result.Output != expected {
		t.Errorf("expected output %s, got: %s", expected, result.Output)
	}
	if result.Format != "url_encoded" {
		t.Errorf("expected format 'url_encoded', got: %s", result.Format)
	}
}

func TestURLDec(t *testing.T) {
	h := &URLDecHandler{}
	encoded := url.QueryEscape("hello world")
	resp := h.Handle(encoded, "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*EncodeResult)
	expected := hexcodec.EncodeString("hello world")
	if result.Output != expected {
		t.Errorf("expected output %s, got: %s", expected, result.Output)
	}
	if result.Format != "hex" {
		t.Errorf("expected format 'hex', got: %s", result.Format)
	}
}

func TestURLDecInvalid(t *testing.T) {
	h := &URLDecHandler{}
	resp := h.Handle("%ZZ", "")
	if resp.Ok {
		t.Fatal("expected failure for invalid URL encoding")
	}
	if resp.Error.Code != "URLDEC_ERROR" {
		t.Errorf("expected error code URLDEC_ERROR, got: %s", resp.Error.Code)
	}
}

func TestEncodeEmpty(t *testing.T) {
	h := &EncodeHandler{}
	resp := h.Handle("", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*EncodeResult)
	if result.Output != "" {
		t.Errorf("expected empty output for empty input, got: %s", result.Output)
	}
}

func TestDecodeEmpty(t *testing.T) {
	h := &DecodeHandler{}
	resp := h.Handle("", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*EncodeResult)
	if result.Output != "" {
		t.Errorf("expected empty output for empty input, got: %s", result.Output)
	}
}