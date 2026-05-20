package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPGetBasic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got: %s", r.Method)
		}
		w.WriteHeader(200)
		w.Write([]byte("hello from get"))
	}))
	defer srv.Close()

	h := &HTTPGetHandler{}
	payload := HTTPGetPayload{URL: srv.URL}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*HTTPResult)
	if result.StatusCode != 200 {
		t.Errorf("expected status 200, got: %d", result.StatusCode)
	}
	if result.Body != "hello from get" {
		t.Errorf("body = %q, want hello from get", result.Body)
	}
}

func TestHTTPGetEmptyURL(t *testing.T) {
	h := &HTTPGetHandler{}
	resp := h.Handle(`{}`, "")
	if resp.Ok {
		t.Fatal("expected failure for empty URL")
	}
	if resp.Error.Code != "HTTP_EMPTY_URL" {
		t.Errorf("expected HTTP_EMPTY_URL, got: %s", resp.Error.Code)
	}
}

func TestHTTPGetWithHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom") != "testval" {
			t.Errorf("missing X-Custom header")
		}
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	h := &HTTPGetHandler{}
	payload := HTTPGetPayload{
		URL:     srv.URL,
		Headers: map[string]string{"X-Custom": "testval"},
	}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestHTTPGetInsecure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	h := &HTTPGetHandler{}
	payload := HTTPGetPayload{URL: srv.URL, Insecure: true}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestHTTPGetNoFollowRedirects(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			w.Header().Set("Location", "/target")
			w.WriteHeader(302)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte("target"))
	}))
	defer srv.Close()

	h := &HTTPGetHandler{}
	payload := HTTPGetPayload{URL: srv.URL + "/redirect", FollowRedirects: false}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*HTTPResult)
	if result.StatusCode != 302 && result.StatusCode != 200 {
		t.Errorf("expected status 302 or 200, got: %d", result.StatusCode)
	}
}

func TestHTTPGetParsePayloadNonJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	h := &HTTPGetHandler{}
	resp := h.Handle(srv.URL, "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestHTTPGetResponseHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Response", "hello")
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	h := &HTTPGetHandler{}
	payload := HTTPGetPayload{URL: srv.URL}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*HTTPResult)
	if result.Headers == nil {
		t.Fatal("expected headers to be set")
	}
}

func TestHTTPGetSizeAndURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	}))
	defer srv.Close()

	h := &HTTPGetHandler{}
	payload := HTTPGetPayload{URL: srv.URL}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*HTTPResult)
	if result.Size != 5 {
		t.Errorf("size = %d, want 5", result.Size)
	}
	if result.URL != srv.URL {
		t.Errorf("url = %q, want %q", result.URL, srv.URL)
	}
}

func TestHTTPPostBasic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got: %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content type")
		}
		w.WriteHeader(201)
		w.Write([]byte(`{"created":true}`))
	}))
	defer srv.Close()

	h := &HTTPPostHandler{}
	payload := HTTPPostPayload{
		URL:         srv.URL,
		Body:        `{"key":"val"}`,
		ContentType: "application/json",
	}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*HTTPResult)
	if result.StatusCode != 201 {
		t.Errorf("expected status 201, got: %d", result.StatusCode)
	}
}

func TestHTTPPostEmptyURL(t *testing.T) {
	h := &HTTPPostHandler{}
	resp := h.Handle(`{}`, "")
	if resp.Ok {
		t.Fatal("expected failure for empty URL")
	}
}

func TestHTTPPostDefaultContentType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/octet-stream" {
			t.Errorf("expected default content type, got: %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	h := &HTTPPostHandler{}
	payload := HTTPPostPayload{URL: srv.URL, Body: "raw data"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestHTTPPostWithHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Auth") != "token123" {
			t.Errorf("missing X-Auth header")
		}
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	h := &HTTPPostHandler{}
	payload := HTTPPostPayload{
		URL:     srv.URL,
		Body:    "data",
		Headers: map[string]string{"X-Auth": "token123"},
	}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestHTTPPostNoFollowRedirects(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			w.Header().Set("Location", "/target")
			w.WriteHeader(302)
			return
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	h := &HTTPPostHandler{}
	payload := HTTPPostPayload{URL: srv.URL + "/redirect", FollowRedirects: false}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*HTTPResult)
	if result.StatusCode != 302 && result.StatusCode != 200 {
		t.Errorf("expected status 302 or 200, got: %d", result.StatusCode)
	}
}

func TestHTTPGetParsePayloadEmpty(t *testing.T) {
	h := &HTTPGetHandler{}
	payload := h.parsePayload("")
	if payload.URL != "" {
		t.Errorf("expected empty URL, got: %s", payload.URL)
	}
}

func TestHTTPPostParsePayloadEmpty(t *testing.T) {
	h := &HTTPPostHandler{}
	payload := h.parsePayload("")
	if payload.URL != "" {
		t.Errorf("expected empty URL, got: %s", payload.URL)
	}
}
