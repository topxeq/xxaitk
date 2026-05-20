package handler

import (
	"encoding/json"
	"testing"
	"time"
)

func cleanupProcess(id string) {
	procMu.Lock()
	if cmd, ok := procMap[id]; ok {
		cmd.Process.Kill()
		delete(procMap, id)
	}
	delete(procOutput, id)
	procMu.Unlock()
}

func TestProcessStart(t *testing.T) {
	h := &ProcessHandler{}
	payload := ProcessPayload{Action: "start", Command: "echo hello", ID: "test-start-1"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	defer cleanupProcess("test-start-1")

	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ProcessResult)
	if result.Status != "running" {
		t.Errorf("expected status 'running', got: %s", result.Status)
	}
	if result.ID != "test-start-1" {
		t.Errorf("expected id 'test-start-1', got: %s", result.ID)
	}
}

func TestProcessStartWithID(t *testing.T) {
	h := &ProcessHandler{}
	payload := ProcessPayload{Action: "start", Command: "sleep 5", ID: "test-custom-id"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	defer cleanupProcess("test-custom-id")

	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ProcessResult)
	if result.ID != "test-custom-id" {
		t.Errorf("expected id 'test-custom-id', got: %s", result.ID)
	}
}

func TestProcessStatus(t *testing.T) {
	h := &ProcessHandler{}
	startPayload := ProcessPayload{Action: "start", Command: "sleep 5", ID: "test-status-1"}
	startData, _ := json.Marshal(startPayload)
	h.Handle(string(startData), "")
	defer cleanupProcess("test-status-1")

	time.Sleep(100 * time.Millisecond)

	statusPayload := ProcessPayload{Action: "status", ID: "test-status-1"}
	statusData, _ := json.Marshal(statusPayload)
	resp := h.Handle(string(statusData), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ProcessResult)
	if result.ID != "test-status-1" {
		t.Errorf("expected id 'test-status-1', got: %s", result.ID)
	}
}

func TestProcessList(t *testing.T) {
	h := &ProcessHandler{}
	startPayload := ProcessPayload{Action: "start", Command: "sleep 5", ID: "test-list-1"}
	startData, _ := json.Marshal(startPayload)
	h.Handle(string(startData), "")
	defer cleanupProcess("test-list-1")

	time.Sleep(100 * time.Millisecond)

	listData, _ := json.Marshal(ProcessPayload{Action: "list"})
	resp := h.Handle(string(listData), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	dataMap := resp.Data.(map[string]interface{})
	count := dataMap["count"].(int)
	if count < 1 {
		t.Errorf("expected at least 1 process, got: %d", count)
	}
}

func TestProcessStop(t *testing.T) {
	h := &ProcessHandler{}
	startPayload := ProcessPayload{Action: "start", Command: "sleep 30", ID: "test-stop-1"}
	startData, _ := json.Marshal(startPayload)
	h.Handle(string(startData), "")

	time.Sleep(100 * time.Millisecond)

	stopPayload := ProcessPayload{Action: "stop", ID: "test-stop-1"}
	stopData, _ := json.Marshal(stopPayload)
	resp := h.Handle(string(stopData), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ProcessResult)
	if result.Status != "killed" {
		t.Errorf("expected status 'killed', got: %s", result.Status)
	}
}

func TestProcessNoAction(t *testing.T) {
	h := &ProcessHandler{}
	resp := h.Handle(`{}`, "")
	if resp.Ok {
		t.Fatal("expected failure for no action and no command")
	}
	if resp.Error.Code != "PROCESS_NO_ACTION" {
		t.Errorf("expected error code PROCESS_NO_ACTION, got: %s", resp.Error.Code)
	}
}

func TestProcessUnknownAction(t *testing.T) {
	h := &ProcessHandler{}
	payload := ProcessPayload{Action: "bad"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if resp.Ok {
		t.Fatal("expected failure for unknown action")
	}
	if resp.Error.Code != "PROCESS_UNKNOWN_ACTION" {
		t.Errorf("expected error code PROCESS_UNKNOWN_ACTION, got: %s", resp.Error.Code)
	}
}

func TestProcessStatusNotFound(t *testing.T) {
	h := &ProcessHandler{}
	payload := ProcessPayload{Action: "status", ID: "nonexistent-id"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if resp.Ok {
		t.Fatal("expected failure for unknown process id")
	}
	if resp.Error.Code != "PROCESS_NOT_FOUND" {
		t.Errorf("expected error code PROCESS_NOT_FOUND, got: %s", resp.Error.Code)
	}
}

func TestProcessStopNotFound(t *testing.T) {
	h := &ProcessHandler{}
	payload := ProcessPayload{Action: "stop", ID: "nonexistent-id"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if resp.Ok {
		t.Fatal("expected failure for unknown process id")
	}
	if resp.Error.Code != "PROCESS_NOT_FOUND" {
		t.Errorf("expected error code PROCESS_NOT_FOUND, got: %s", resp.Error.Code)
	}
}

func TestProcessNoCommand(t *testing.T) {
	h := &ProcessHandler{}
	payload := ProcessPayload{Action: "start"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if resp.Ok {
		t.Fatal("expected failure for start without command")
	}
	if resp.Error.Code != "PROCESS_NO_COMMAND" {
		t.Errorf("expected error code PROCESS_NO_COMMAND, got: %s", resp.Error.Code)
	}
}

func TestProcessStatusNoID(t *testing.T) {
	h := &ProcessHandler{}
	payload := ProcessPayload{Action: "status"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if resp.Ok {
		t.Fatal("expected failure for status without id")
	}
	if resp.Error.Code != "PROCESS_NO_ID" {
		t.Errorf("expected error code PROCESS_NO_ID, got: %s", resp.Error.Code)
	}
}

func TestProcessImplicitStart(t *testing.T) {
	h := &ProcessHandler{}
	resp := h.Handle(`{"command": "echo implicit", "id": "test-implicit-1"}`, "")
	defer cleanupProcess("test-implicit-1")

	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ProcessResult)
	if result.Action != "start" {
		t.Errorf("expected action 'start', got: %s", result.Action)
	}
	if result.ID != "test-implicit-1" {
		t.Errorf("expected id 'test-implicit-1', got: %s", result.ID)
	}
}
