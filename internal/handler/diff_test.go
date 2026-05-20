package handler

import (
	"encoding/json"
	"os"
	"testing"
)

func TestDiffSameContent(t *testing.T) {
	h := &DiffHandler{}
	payload := DiffPayload{ContentA: "line1\nline2", ContentB: "line1\nline2"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*DiffResult)
	if !result.Same {
		t.Error("expected Same=true for identical content")
	}
}

func TestDiffDifferentContent(t *testing.T) {
	h := &DiffHandler{}
	payload := DiffPayload{ContentA: "line1\nline2", ContentB: "line1\nline3"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*DiffResult)
	if result.Same {
		t.Error("expected Same=false for different content")
	}
	if result.Lines == nil {
		t.Error("expected diff lines")
	}
}

func TestDiffFiles(t *testing.T) {
	fA, err := os.CreateTemp("", "diff_test_a_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(fA.Name())
	fA.WriteString("aaa\nbbb")
	fA.Close()

	fB, err := os.CreateTemp("", "diff_test_b_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(fB.Name())
	fB.WriteString("aaa\nccc")
	fB.Close()

	h := &DiffHandler{}
	payload := DiffPayload{FileA: fA.Name(), FileB: fB.Name()}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*DiffResult)
	if result.Same {
		t.Error("expected Same=false for different files")
	}
}

func TestDiffFileNotFound(t *testing.T) {
	h := &DiffHandler{}
	payload := DiffPayload{FileA: "/nonexistent/file_a.txt", ContentB: "hello"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if resp.Ok {
		t.Fatal("expected failure for nonexistent file")
	}
	if resp.Error.Code != "DIFF_INPUT_ERROR" {
		t.Errorf("expected error code DIFF_INPUT_ERROR, got: %s", resp.Error.Code)
	}
}

func TestDiffEmptyPayload(t *testing.T) {
	h := &DiffHandler{}
	resp := h.Handle("{}", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*DiffResult)
	if !result.Same {
		t.Error("expected Same=true when both inputs are empty")
	}
}

func TestDiffAddOnly(t *testing.T) {
	h := &DiffHandler{}
	payload := DiffPayload{ContentA: "", ContentB: "line1\nline2"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*DiffResult)
	if result.Same {
		t.Error("expected Same=false")
	}
	if result.Adds == 0 {
		t.Error("expected adds > 0 when contentB has lines and contentA is empty")
	}
}

func TestDiffDeleteOnly(t *testing.T) {
	h := &DiffHandler{}
	payload := DiffPayload{ContentA: "line1\nline2", ContentB: ""}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*DiffResult)
	if result.Same {
		t.Error("expected Same=false")
	}
	if result.Dels == 0 {
		t.Error("expected dels > 0 when contentA has lines and contentB is empty")
	}
}

func TestDiffMixedChanges(t *testing.T) {
	h := &DiffHandler{}
	payload := DiffPayload{
		ContentA: "alpha\nbeta\ngamma",
		ContentB: "alpha\ndelta\ngamma",
	}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*DiffResult)
	if result.Same {
		t.Error("expected Same=false")
	}
	if result.Adds == 0 && result.Dels == 0 {
		t.Error("expected at least one add or del")
	}
}
