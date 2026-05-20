package handler

import (
	"testing"
)

func TestEvalSimple(t *testing.T) {
	h := &EvalHandler{}
	resp := h.Handle("1 + 2", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ScriptResult)
	if result.Result != int64(3) {
		t.Errorf("expected result 3, got: %v", result.Result)
	}
}
