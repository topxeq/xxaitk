package handler

import (
	"encoding/json"
	"testing"
)

func TestPortSingleClosedPort(t *testing.T) {
	h := &PortHandler{}
	payload := PortPayload{Host: "localhost", Port: 9999, Timeout: 200}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*PortResult)
	if result.Count != 1 {
		t.Errorf("expected count=1, got: %d", result.Count)
	}
	if result.Results[0].Open {
		t.Error("expected port 9999 to be closed")
	}
}

func TestPortJSONPayload(t *testing.T) {
	h := &PortHandler{}
	resp := h.Handle(`{"host":"localhost","port":80,"timeout":200}`, "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*PortResult)
	if result.Host != "localhost" {
		t.Errorf("expected host=localhost, got: %s", result.Host)
	}
}

func TestPortHostPortFormat(t *testing.T) {
	h := &PortHandler{}
	resp := h.Handle("localhost:80", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*PortResult)
	if result.Host != "localhost" {
		t.Errorf("expected host=localhost, got: %s", result.Host)
	}
	if result.Count != 1 {
		t.Errorf("expected count=1, got: %d", result.Count)
	}
}

func TestPortDefaultHost(t *testing.T) {
	h := &PortHandler{}
	payload := PortPayload{Port: 80, Timeout: 200}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*PortResult)
	if result.Host != "localhost" {
		t.Errorf("expected default host=localhost, got: %s", result.Host)
	}
}

func TestPortScanRange(t *testing.T) {
	h := &PortHandler{}
	payload := PortPayload{Host: "localhost", From: 1, To: 5, Timeout: 100}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*PortResult)
	if result.Host != "localhost" {
		t.Errorf("expected host=localhost, got: %s", result.Host)
	}
}

func TestPortParsePayloadEmpty(t *testing.T) {
	h := &PortHandler{}
	payload := h.parsePayload("")
	if payload.Host != "" {
		t.Errorf("expected empty host, got: %s", payload.Host)
	}
}

func TestPortDefaultTimeout(t *testing.T) {
	h := &PortHandler{}
	resp := h.Handle(`{"host":"localhost","port":9999,"timeout":0}`, "test")
	if !resp.Ok {
		t.Fatalf("expected ok: %v", resp.Error)
	}
}

func TestPortScanSmallRange(t *testing.T) {
	h := &PortHandler{}
	resp := h.Handle(`{"host":"localhost","from":1,"to":5,"timeout":200}`, "test")
	if !resp.Ok {
		t.Fatalf("expected ok: %v", resp.Error)
	}
	result := resp.Data.(*PortResult)
	if result.Count < 0 {
		t.Errorf("count = %d", result.Count)
	}
}

func TestPortScanDefaultRange(t *testing.T) {
	h := &PortHandler{}
	resp := h.Handle(`{"host":"localhost","timeout":200}`, "test")
	if !resp.Ok {
		t.Fatalf("expected ok: %v", resp.Error)
	}
}

func TestPortProtocol(t *testing.T) {
	h := &PortHandler{}
	resp := h.Handle(`{"host":"localhost","port":22,"timeout":200,"protocol":"tcp"}`, "test")
	if !resp.Ok {
		t.Fatalf("expected ok: %v", resp.Error)
	}
}

func TestPortCheckCommonService(t *testing.T) {
	h := &PortHandler{}
	resp := h.Handle(`{"host":"localhost","port":80,"timeout":200}`, "test")
	if !resp.Ok {
		t.Fatalf("expected ok: %v", resp.Error)
	}
	result := resp.Data.(*PortResult)
	if len(result.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Results))
	}
	if result.Results[0].Port != 80 {
		t.Errorf("port = %d, want 80", result.Results[0].Port)
	}
}
