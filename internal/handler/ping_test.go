package handler

import (
	"encoding/json"
	"testing"
)

func TestPingLocalhost(t *testing.T) {
	h := &PingHandler{}
	resp := h.Handle("localhost", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*PingResult)
	if result.Host != "localhost" {
		t.Errorf("expected host 'localhost', got: %s", result.Host)
	}
}

func TestPingEmptyHost(t *testing.T) {
	h := &PingHandler{}
	resp := h.Handle("", "")
	if resp.Ok {
		t.Fatal("expected failure for empty host")
	}
	if resp.Error.Code != "PING_EMPTY_HOST" {
		t.Errorf("expected error code PING_EMPTY_HOST, got: %s", resp.Error.Code)
	}
}

func TestPingJSONPayload(t *testing.T) {
	h := &PingHandler{}
	payload := PingPayload{Host: "localhost"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*PingResult)
	if result.Host != "localhost" {
		t.Errorf("expected host 'localhost', got: %s", result.Host)
	}
}

func TestPingUnreachableHost(t *testing.T) {
	h := &PingHandler{}
	resp := h.Handle("this.host.does.not.exist.invalid", "")
	if !resp.Ok {
		t.Fatalf("expected ok (reachable=false, not error), got error: %v", resp.Error)
	}
	result := resp.Data.(*PingResult)
	if result.Reachable {
		t.Error("expected reachable=false for unresolvable host")
	}
}
