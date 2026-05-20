package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got: %s", r.Method)
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"deleted":true}`))
	}))
	defer srv.Close()

	h := &HTTPMethodHandler{Method: "DELETE"}
	payload := HTTPMethodPayload{URL: srv.URL}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*HTTPResult)
	if result.StatusCode != 200 {
		t.Errorf("expected status 200, got: %d", result.StatusCode)
	}
}

func TestHTTPPut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got: %s", r.Method)
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"updated":true}`))
	}))
	defer srv.Close()

	h := &HTTPMethodHandler{Method: "PUT"}
	payload := HTTPMethodPayload{URL: srv.URL, Body: `{"key":"val"}`, ContentType: "application/json"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*HTTPResult)
	if result.StatusCode != 200 {
		t.Errorf("expected status 200, got: %d", result.StatusCode)
	}
}

func TestHTTPPatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got: %s", r.Method)
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"patched":true}`))
	}))
	defer srv.Close()

	h := &HTTPMethodHandler{Method: "PATCH"}
	payload := HTTPMethodPayload{URL: srv.URL, Body: `{"key":"val"}`, ContentType: "application/json"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*HTTPResult)
	if result.StatusCode != 200 {
		t.Errorf("expected status 200, got: %d", result.StatusCode)
	}
}

func TestHTTPMethodEmptyURL(t *testing.T) {
	h := &HTTPMethodHandler{Method: "DELETE"}
	resp := h.Handle(`{}`, "")
	if resp.Ok {
		t.Fatal("expected failure for empty URL")
	}
	if resp.Error.Code != "HTTP_EMPTY_URL" {
		t.Errorf("expected HTTP_EMPTY_URL, got: %s", resp.Error.Code)
	}
}

func TestHTTPMethodJSONPayload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	h := &HTTPMethodHandler{Method: "PUT"}
	resp := h.Handle(`{"url":"`+srv.URL+`"}`, "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestHTTPMethodParsePayloadNonJSON(t *testing.T) {
	h := &HTTPMethodHandler{Method: "DELETE"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	resp := h.Handle(srv.URL, "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}
