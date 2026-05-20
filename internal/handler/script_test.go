package handler

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/topxeq/xxaitk/internal/script"
)

func TestScriptSimple(t *testing.T) {
	h := &ScriptHandler{}
	resp := h.Handle(`print("hello")`, "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ScriptResult)
	found := false
	for _, s := range result.Output {
		if strings.Contains(s, "hello") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected output to contain 'hello', got: %v", result.Output)
	}
}

func TestScriptArithmetic(t *testing.T) {
	h := &ScriptHandler{}
	resp := h.Handle(`1 + 2`, "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ScriptResult)
	if result.Result != int64(3) {
		t.Errorf("expected result 3, got: %v", result.Result)
	}
}

func TestScriptLetAndPrint(t *testing.T) {
	h := &ScriptHandler{}
	resp := h.Handle("let x = 42\nprint(str_from_int(x))", "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ScriptResult)
	found := false
	for _, s := range result.Output {
		if strings.Contains(s, "42") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected output to contain '42', got: %v", result.Output)
	}
}

func TestScriptIfElse(t *testing.T) {
	h := &ScriptHandler{}
	code := `let x = 10
if x > 5 {
    print("big")
} else {
    print("small")
}`
	resp := h.Handle(code, "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ScriptResult)
	found := false
	for _, s := range result.Output {
		if strings.Contains(s, "big") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected output to contain 'big', got: %v", result.Output)
	}
}

func TestScriptFnCall(t *testing.T) {
	h := &ScriptHandler{}
	code := `fn add(a, b) { return a + b }
print(str_from_int(add(3, 4)))`
	resp := h.Handle(code, "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ScriptResult)
	found := false
	for _, s := range result.Output {
		if strings.Contains(s, "7") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected output to contain '7', got: %v", result.Output)
	}
}

func TestScriptJSONPayload(t *testing.T) {
	h := &ScriptHandler{}
	payload := ScriptPayload{Source: `print("json_test")`}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ScriptResult)
	found := false
	for _, s := range result.Output {
		if strings.Contains(s, "json_test") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected output to contain 'json_test', got: %v", result.Output)
	}
}

func TestScriptCompileError(t *testing.T) {
	h := &ScriptHandler{}
	resp := h.Handle(`fn {`, "")
	if resp.Ok {
		t.Fatal("expected failure for invalid syntax")
	}
	if resp.Error.Code != "SCRIPT_COMPILE_ERROR" {
		t.Errorf("expected error code SCRIPT_COMPILE_ERROR, got: %s", resp.Error.Code)
	}
}

func TestScriptRuntimeError(t *testing.T) {
	h := &ScriptHandler{}
	resp := h.Handle(`undefined_var_fn()`, "")
	if resp.Ok {
		t.Fatal("expected failure for runtime error")
	}
	if resp.Error.Code != "SCRIPT_RUNTIME_ERROR" {
		t.Errorf("expected error code SCRIPT_RUNTIME_ERROR, got: %s", resp.Error.Code)
	}
}

func TestScriptDebug(t *testing.T) {
	h := &ScriptHandler{}
	payload := ScriptPayload{Source: `1 + 1`, Debug: true}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*ScriptResult)
	if result.Debug == nil {
		t.Fatal("expected debug info to be present")
	}
	if result.Debug.InstructionsExecuted <= 0 {
		t.Errorf("expected instructions_executed > 0, got: %d", result.Debug.InstructionsExecuted)
	}
}

func TestScriptEmptySource(t *testing.T) {
	h := &ScriptHandler{}
	resp := h.Handle("", "test")
	if !resp.Ok {
		t.Fatalf("expected ok for empty source: %v", resp.Error)
	}
}

func TestScriptObjectToInterfaceNil(t *testing.T) {
	result := objectToInterface(nil)
	if result != nil {
		t.Errorf("nil object should map to nil, got %v", result)
	}
}

func TestScriptObjectToInterfaceBool(t *testing.T) {
	result := objectToInterface(script.BoolObject(true))
	if result != true {
		t.Errorf("BoolObject(true) should map to true, got %v", result)
	}
}

func TestScriptObjectToInterfaceInt(t *testing.T) {
	result := objectToInterface(script.IntObject(42))
	if result != int64(42) {
		t.Errorf("IntObject(42) should map to int64(42), got %v", result)
	}
}

func TestScriptObjectToInterfaceFloat(t *testing.T) {
	result := objectToInterface(script.FloatObject(3.14))
	if result != float64(3.14) {
		t.Errorf("FloatObject(3.14) should map to float64(3.14), got %v", result)
	}
}

func TestScriptObjectToInterfaceString(t *testing.T) {
	result := objectToInterface(script.StringObject("hello"))
	if result != "hello" {
		t.Errorf("StringObject should map to string, got %v", result)
	}
}

func TestScriptObjectToInterfaceList(t *testing.T) {
	list := script.ListObject{Elements: []script.Object{script.IntObject(1), script.StringObject("a")}}
	result := objectToInterface(list)
	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", result)
	}
	if len(arr) != 2 {
		t.Errorf("len = %d, want 2", len(arr))
	}
}

func TestScriptObjectToInterfaceMap(t *testing.T) {
	m := script.MapObject{Pairs: map[string]script.Object{"k": script.IntObject(1)}, Keys: []string{"k"}}
	result := objectToInterface(m)
	rm, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if rm["k"] != int64(1) {
		t.Errorf("map[k] = %v, want 1", rm["k"])
	}
}

func TestScriptObjectToInterfaceDefault(t *testing.T) {
	result := objectToInterface(script.BuiltinFn{Name: "test"})
	s, ok := result.(string)
	if !ok {
		t.Errorf("default should use Inspect(), got %T: %v", result, result)
	}
	_ = s
}

func TestScriptObjectToInterfaceNilObject(t *testing.T) {
	result := objectToInterface(script.NilObject{})
	if result != nil {
		t.Errorf("NilObject should map to nil, got %v", result)
	}
}
