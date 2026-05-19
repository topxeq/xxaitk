package handler

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/topxeq/xxaitk/internal/output"
	"github.com/topxeq/xxaitk/internal/script"
)

type ScriptHandler struct{}

type ScriptPayload struct {
	Source string `json:"source"`
	Unsafe bool   `json:"unsafe,omitempty"`
	Timeout int   `json:"timeout,omitempty"`
	Args   []interface{} `json:"args,omitempty"`
	Debug  bool   `json:"debug,omitempty"`
}

type ScriptResult struct {
	Result     interface{}  `json:"result"`
	Output     []string     `json:"output"`
	DurationMs int64        `json:"duration_ms"`
	Debug      *ScriptDebug `json:"debug,omitempty"`
	RawResult  script.Object `json:"-"`
}

type ScriptDebug struct {
	InstructionsExecuted int      `json:"instructions_executed"`
	StackDepth           int      `json:"max_stack_depth"`
	StackTrace           []string `json:"stack_trace,omitempty"`
}

func (h *ScriptHandler) Handle(data string, source string) *output.Response {
	start := time.Now()

	payload := h.parsePayload(data)
	sourceCode := payload.Source
	if sourceCode == "" {
		sourceCode = data
	}

	lexer := script.NewLexer(sourceCode)
	tokens := lexer.Tokenize()

	parser := script.NewParser(tokens)
	ast, err := parser.Parse()
	if err != nil {
		return output.Fail("script", source, "SCRIPT_COMPILE_ERROR",
			fmt.Sprintf("parse error: %s", err.Error()), "", start)
	}

	compiler := script.NewCompiler()
	if err := compiler.Compile(ast); err != nil {
		return output.Fail("script", source, "SCRIPT_COMPILE_ERROR",
			fmt.Sprintf("compile error: %s", err.Error()), "", start)
	}

	builtins := script.GetBuiltins(payload.Unsafe)
	vm := script.NewVMWithGlobals(builtins, payload.Unsafe, compiler.GlobalNames())
	script.PrintCallback = func(s string) {
		vm.AddOutput(s)
	}

	result, err := vm.Run(compiler.Instructions(), compiler.Constants())
	if err != nil {
		return output.Fail("script", source, "SCRIPT_RUNTIME_ERROR",
			fmt.Sprintf("runtime error: %s", err.Error()), "", start)
	}

	var debugInfo *ScriptDebug
	if payload.Debug {
		debugInfo = &ScriptDebug{
			InstructionsExecuted: vm.OpCount(),
			StackTrace:           []string{},
		}
	}

	return output.Success("script", source, &ScriptResult{
		Result:     objectToInterface(result),
		Output:     vm.Outputs(),
		DurationMs: time.Since(start).Milliseconds(),
		Debug:      debugInfo,
		RawResult:  result,
	}, start)
}

func (h *ScriptHandler) parsePayload(data string) *ScriptPayload {
	payload := &ScriptPayload{}
	trimmed := strings.TrimSpace(data)
	if len(trimmed) == 0 {
		return payload
	}
	if trimmed[0] == '{' {
		json.Unmarshal([]byte(trimmed), payload)
	} else {
		payload.Source = data
	}
	return payload
}

func objectToInterface(obj script.Object) interface{} {
	if obj == nil {
		return nil
	}
	switch v := obj.(type) {
	case script.NilObject:
		return nil
	case script.BoolObject:
		return bool(v)
	case script.IntObject:
		return int64(v)
	case script.FloatObject:
		return float64(v)
	case script.StringObject:
		return string(v)
	case script.ListObject:
		result := make([]interface{}, len(v.Elements))
		for i, e := range v.Elements {
			result[i] = objectToInterface(e)
		}
		return result
	case script.MapObject:
		result := make(map[string]interface{})
		for k, val := range v.Pairs {
			result[k] = objectToInterface(val)
		}
		return result
	default:
		return obj.Inspect()
	}
}
