package script

import (
	"os"
	"testing"
)

func TestIsHexDigit(t *testing.T) {
	tests := []struct {
		ch   rune
		want bool
	}{
		{'0', true}, {'9', true}, {'a', true}, {'f', true},
		{'A', true}, {'F', true}, {'g', false}, {'z', false},
		{'G', false}, {' ', false}, {'/', false},
	}
	for _, tt := range tests {
		if got := isHexDigit(tt.ch); got != tt.want {
			t.Errorf("isHexDigit(%q) = %v, want %v", tt.ch, got, tt.want)
		}
	}
}

func TestFormatError(t *testing.T) {
	result := FormatError("test error", 5, 10)
	expected := "line 5, col 10: test error"
	if result != expected {
		t.Errorf("FormatError = %q, want %q", result, expected)
	}
}

func TestFormatErrorZero(t *testing.T) {
	result := FormatError("err", 0, 0)
	if result != "line 0, col 0: err" {
		t.Errorf("FormatError = %q", result)
	}
}

func TestOpCodeStringAll(t *testing.T) {
	codes := []OpCode{
		OpConstant, OpInt, OpFloat, OpNil, OpTrue, OpFalse,
		OpAdd, OpSub, OpMul, OpDiv, OpMod, OpPow,
		OpEq, OpNeq, OpLt, OpGt, OpLte, OpGte,
		OpAnd, OpOr, OpNot, OpNegate,
		OpIndex, OpDot,
		OpJump, OpJumpIfFalse, OpJumpIfTrue, OpLoop,
		OpGetLocal, OpSetLocal, OpGetGlobal, OpSetGlobal, OpDefineGlobal, OpDefineConst,
		OpCall, OpReturn, OpPop,
		OpBuildList, OpBuildMap,
		OpBreakpoint, OpNilReturn,
		OpGetFree, OpClosure,
		OpBreak, OpContinue,
	}
	for _, op := range codes {
		s := op.String()
		if s == "UNKNOWN" {
			t.Errorf("OpCode %d should have a name", op)
		}
	}
}

func TestOpCodeStringUnknown(t *testing.T) {
	op := OpCode(999)
	if op.String() != "UNKNOWN" {
		t.Errorf("unknown opcode = %q, want UNKNOWN", op.String())
	}
}

func TestExecShell(t *testing.T) {
	out, err := execShell("echo hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello\n" {
		t.Errorf("execShell = %q, want hello\\n", out)
	}
}

func TestExecShellFailed(t *testing.T) {
	_, err := execShell("false")
	if err == nil {
		t.Error("expected error for failed command")
	}
}

func TestObjectToJSONList(t *testing.T) {
	obj := ListObject{Elements: []Object{IntObject(1), StringObject("hello")}}
	result := objectToJSON(obj)
	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", result)
	}
	if len(arr) != 2 {
		t.Errorf("len = %d, want 2", len(arr))
	}
}

func TestObjectToJSONMap(t *testing.T) {
	obj := MapObject{Pairs: map[string]Object{"key": StringObject("val")}, Keys: []string{"key"}}
	result := objectToJSON(obj)
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if m["key"] != "val" {
		t.Errorf("m[key] = %v, want val", m["key"])
	}
}

func TestObjectToJSONNil(t *testing.T) {
	if objectToJSON(NilObject{}) != nil {
		t.Error("NilObject should map to nil")
	}
}

func TestObjectToJSONBool(t *testing.T) {
	if objectToJSON(BoolObject(true)) != true {
		t.Error("BoolObject(true) should map to true")
	}
}

func TestObjectToJSONFloat(t *testing.T) {
	f, ok := objectToJSON(FloatObject(3.14)).(float64)
	if !ok || f != 3.14 {
		t.Errorf("FloatObject(3.14) = %v", objectToJSON(FloatObject(3.14)))
	}
}

func TestObjectToJSONDefault(t *testing.T) {
	result := objectToJSON(BuiltinFn{Name: "test"})
	if result != nil {
		t.Errorf("default case should return nil, got %v", result)
	}
}

func TestJsonToObjectNil(t *testing.T) {
	obj := jsonToObject(nil)
	if _, ok := obj.(NilObject); !ok {
		t.Errorf("expected NilObject, got %T", obj)
	}
}

func TestJsonToObjectBool(t *testing.T) {
	obj := jsonToObject(true)
	if b, ok := obj.(BoolObject); !ok || bool(b) != true {
		t.Errorf("expected BoolObject(true), got %T", obj)
	}
}

func TestJsonToObjectFloatInt(t *testing.T) {
	obj := jsonToObject(float64(42))
	if i, ok := obj.(IntObject); !ok || int64(i) != 42 {
		t.Errorf("expected IntObject(42), got %T", obj)
	}
}

func TestJsonToObjectFloatLarge(t *testing.T) {
	obj := jsonToObject(float64(1e16))
	if _, ok := obj.(FloatObject); !ok {
		t.Errorf("large float should stay FloatObject, got %T", obj)
	}
}

func TestJsonToObjectString(t *testing.T) {
	obj := jsonToObject("hello")
	if s, ok := obj.(StringObject); !ok || string(s) != "hello" {
		t.Errorf("expected StringObject, got %T", obj)
	}
}

func TestJsonToObjectList(t *testing.T) {
	obj := jsonToObject([]interface{}{1.0, "hello"})
	if l, ok := obj.(ListObject); !ok || len(l.Elements) != 2 {
		t.Errorf("expected ListObject with 2 elements, got %T", obj)
	}
}

func TestJsonToObjectMap(t *testing.T) {
	obj := jsonToObject(map[string]interface{}{"key": "val"})
	if m, ok := obj.(MapObject); !ok || m.Pairs["key"] != StringObject("val") {
		t.Errorf("expected MapObject, got %T", obj)
	}
}

func TestJsonToObjectDefault(t *testing.T) {
	obj := jsonToObject(42)
	if _, ok := obj.(NilObject); !ok {
		t.Errorf("expected NilObject for unknown type, got %T", obj)
	}
}

func TestCompareFloats(t *testing.T) {
	if compareFloats(1.0, 2.0) >= 0 {
		t.Error("1.0 < 2.0")
	}
	if compareFloats(2.0, 1.0) <= 0 {
		t.Error("2.0 > 1.0")
	}
	if compareFloats(1.0, 1.0) != 0 {
		t.Error("1.0 == 1.0")
	}
}

func TestCompareObjectsMixed(t *testing.T) {
	if compareObjects(IntObject(1), FloatObject(2.0)) >= 0 {
		t.Error("1 < 2.0")
	}
	if compareObjects(FloatObject(1.0), IntObject(2)) >= 0 {
		t.Error("1.0 < 2")
	}
	if compareObjects(StringObject("a"), IntObject(1)) >= 0 {
		t.Error("mixed types should be -1")
	}
}

func TestCompareObjectsString(t *testing.T) {
	if compareObjects(StringObject("abc"), StringObject("abd")) >= 0 {
		t.Error("abc < abd")
	}
	if compareObjects(StringObject("xyz"), StringObject("abc")) <= 0 {
		t.Error("xyz > abc")
	}
}

func TestCompareObjectsDefault(t *testing.T) {
	if compareObjects(NilObject{}, NilObject{}) != 0 {
		t.Error("nil == nil")
	}
}

func TestObjectEqualsFloat(t *testing.T) {
	if !objectEquals(FloatObject(3.14), FloatObject(3.14)) {
		t.Error("3.14 == 3.14")
	}
	if objectEquals(FloatObject(3.14), FloatObject(2.71)) {
		t.Error("3.14 != 2.71")
	}
}

func TestObjectEqualsDifferentTypes(t *testing.T) {
	if objectEquals(IntObject(1), StringObject("1")) {
		t.Error("different types should not be equal")
	}
}

func TestIsTruthyList(t *testing.T) {
	if !IsTruthy(ListObject{Elements: []Object{IntObject(1)}}) {
		t.Error("non-empty list should be truthy")
	}
	if IsTruthy(ListObject{Elements: []Object{}}) {
		t.Error("empty list should be falsy")
	}
}

func TestIsTruthyMap(t *testing.T) {
	if !IsTruthy(MapObject{Pairs: map[string]Object{"k": IntObject(1)}}) {
		t.Error("non-empty map should be truthy")
	}
	if IsTruthy(MapObject{Pairs: map[string]Object{}}) {
		t.Error("empty map should be falsy")
	}
}

func TestIsTruthyDefault(t *testing.T) {
	if !IsTruthy(BuiltinFn{Name: "test"}) {
		t.Error("builtin fn should be truthy")
	}
}

func TestInspectValueString(t *testing.T) {
	s := inspectValue(StringObject("hello"))
	if s != "\"hello\"" {
		t.Errorf("inspectValue(StringObject) = %q", s)
	}
}

func TestInspectValueNil(t *testing.T) {
	s := inspectValue(nil)
	if s != "nil" {
		t.Errorf("inspectValue(nil) = %q", s)
	}
}

func TestReadStringEscapeNewline(t *testing.T) {
	lexer := NewLexer(`"line1\nline2"`)
	tokens := lexer.Tokenize()
	if tokens[0].Literal != "line1\nline2" {
		t.Errorf("got %q", tokens[0].Literal)
	}
}

func TestReadStringEscapeTab(t *testing.T) {
	lexer := NewLexer(`"col1\tcol2"`)
	tokens := lexer.Tokenize()
	if tokens[0].Literal != "col1\tcol2" {
		t.Errorf("got %q", tokens[0].Literal)
	}
}

func TestReadStringEscapeQuote(t *testing.T) {
	lexer := NewLexer(`"say \"hi\""`)
	tokens := lexer.Tokenize()
	if tokens[0].Literal != `say "hi"` {
		t.Errorf("got %q", tokens[0].Literal)
	}
}

func TestReadStringEscapeBackslash(t *testing.T) {
	lexer := NewLexer(`"path\\name"`)
	tokens := lexer.Tokenize()
	if tokens[0].Literal != `path\name` {
		t.Errorf("got %q", tokens[0].Literal)
	}
}

func TestParsePower(t *testing.T) {
	result, _ := runEngineTest(t, `math_pow(2, 3)`)
	if result.Inspect() != "8" {
		t.Errorf("2**3 = %s, want 8", result.Inspect())
	}
}

func TestParseExprStatementAssignment(t *testing.T) {
	_, outputs := runEngineTest(t, `let x = 10
x = 20
print(str_from_int(x))`)
	if len(outputs) != 1 || outputs[0] != "20" {
		t.Errorf("outputs = %v", outputs)
	}
}

func TestParseExprStatementFnCall(t *testing.T) {
	_, outputs := runEngineTest(t, `print("test")`)
	if len(outputs) != 1 || outputs[0] != "test" {
		t.Errorf("outputs = %v", outputs)
	}
}

func TestIOBuiltinsEdgeCases(t *testing.T) {
	_, _ = runEngineTest(t, `io_read_file()`)
	_, _ = runEngineTest(t, `io_write_file()`)
	_, _ = runEngineTest(t, `io_append_file()`)
	_, _ = runEngineTest(t, `io_exists()`)
	_, _ = runEngineTest(t, `io_is_dir()`)
	_, _ = runEngineTest(t, `io_is_file()`)
	_, _ = runEngineTest(t, `io_list_dir()`)
	_, _ = runEngineTest(t, `io_size()`)
	_, _ = runEngineTest(t, `io_remove()`)
	_, _ = runEngineTest(t, `io_abs_path()`)
}

func TestIOCopyFile(t *testing.T) {
	f1, _ := os.CreateTemp("", "io_copy_src_*.txt")
	f1.WriteString("copy content")
	f1.Close()
	defer os.Remove(f1.Name())

	f2, _ := os.CreateTemp("", "io_copy_dst_*.txt")
	f2.Close()
	defer os.Remove(f2.Name())

	_, _ = runEngineTest(t, `io_copy("`+f1.Name()+`", "`+f2.Name()+`")`)
}

func TestNetBuiltinsNoArgs(t *testing.T) {
	_, _ = runEngineTest(t, `net_http_get()`)
	_, _ = runEngineTest(t, `net_dns_lookup()`)
	_, _ = runEngineTest(t, `net_tcp_connect()`)
}

func TestConvBuiltinsEdgeCases(t *testing.T) {
	_, _ = runEngineTest(t, `conv_to_int()`)
	_, _ = runEngineTest(t, `conv_to_float()`)
	_, _ = runEngineTest(t, `conv_to_string()`)
	_, _ = runEngineTest(t, `conv_to_bool()`)
	_, _ = runEngineTest(t, `conv_hex_encode()`)
	_, _ = runEngineTest(t, `conv_hex_decode()`)
	_, _ = runEngineTest(t, `conv_b64_encode()`)
	_, _ = runEngineTest(t, `conv_b64_decode()`)
}

func TestConvB64DecodeInvalid(t *testing.T) {
	result, _ := runEngineTest(t, `conv_b64_decode("!!!invalid!!!")`)
	if result.Inspect() != "" {
		t.Errorf("invalid base64 should return empty string, got %q", result.Inspect())
	}
}

func TestConvHexDecodeInvalid(t *testing.T) {
	result, _ := runEngineTest(t, `conv_hex_decode("zzzz")`)
	if result.Inspect() != "" {
		t.Errorf("invalid hex should return empty string, got %q", result.Inspect())
	}
}

func TestTypeIsFloat(t *testing.T) {
	result, _ := runEngineTest(t, `type_is_float(3.14)`)
	if result.Inspect() != "true" {
		t.Errorf("type_is_float(3.14) = %s", result.Inspect())
	}
}

func TestTypeOfNoArgs(t *testing.T) {
	result, _ := runEngineTest(t, `type_of()`)
	if result.Inspect() != "nil" {
		t.Errorf("type_of() = %s", result.Inspect())
	}
}

func TestContainsStr(t *testing.T) {
	if !containsStr([]string{"hello", "world"}, "world") {
		t.Error("should contain 'world'")
	}
	if containsStr([]string{"hello"}, "world") {
		t.Error("should not contain 'world'")
	}
}

func TestHexEncodeDecode(t *testing.T) {
	encoded := hexEncode("hello")
	if encoded != "68656c6c6f" {
		t.Errorf("hexEncode = %q", encoded)
	}
	decoded, err := hexDecode("68656c6c6f")
	if err != nil || decoded != "hello" {
		t.Errorf("hexDecode = %q, err = %v", decoded, err)
	}
}

func TestHexDecodeInvalid(t *testing.T) {
	_, err := hexDecode("zzzz")
	if err == nil {
		t.Error("expected error for invalid hex")
	}
}

func TestJoinParams(t *testing.T) {
	s := joinParams([]string{"a", "b", "c"})
	if s != "a,b,c" {
		t.Errorf("joinParams = %q", s)
	}
}

func TestJoinParamsEmpty(t *testing.T) {
	s := joinParams(nil)
	if s != "" {
		t.Errorf("joinParams(nil) = %q", s)
	}
}

func TestFilterNewlines(t *testing.T) {
	tokens := []Token{
		{Type: TokIdent, Literal: "hello", Line: 1},
		{Type: TokEOF, Literal: "", Line: 1},
	}
	filtered := filterNewlines(tokens)
	if len(filtered) != 2 {
		t.Errorf("filterNewlines len = %d, want 2", len(filtered))
	}
}

func TestGetBuiltinsCount(t *testing.T) {
	safe := GetBuiltins(false)
	unsafe := GetBuiltins(true)
	if len(unsafe) <= len(safe) {
		t.Error("unsafe should have more builtins than safe")
	}
}

func TestTokenStringWithLiteral(t *testing.T) {
	tok := Token{Type: TokIdent, Literal: "myvar", Line: 1}
	if tok.String() != "myvar" {
		t.Errorf("Token.String() = %q, want myvar", tok.String())
	}
}

func TestTokenStringWithoutLiteral(t *testing.T) {
	tok := Token{Type: TokEOF, Literal: "", Line: 1}
	s := tok.String()
	if s != "EOF" {
		t.Errorf("Token.String() = %q, want EOF", s)
	}
}

func TestTokenTypeNameAll(t *testing.T) {
	tests := map[TokenType]string{
		TokEOF: "EOF", TokNewline: "newline", TokIdent: "ident",
		TokInt: "int", TokFloat: "float", TokString: "string",
		TokLet: "let", TokConst: "const", TokFn: "fn",
		TokReturn: "return", TokIf: "if", TokElif: "elif",
		TokElse: "else", TokWhile: "while", TokFor: "for",
		TokIn: "in", TokTrue: "true", TokFalse: "false",
		TokNil: "nil", TokBreak: "break", TokContinue: "continue",
		TokBreakpoint: "breakpoint",
	}
	for tt, want := range tests {
		got := tokenTypeName(tt)
		if got != want {
			t.Errorf("tokenTypeName(%d) = %q, want %q", tt, got, want)
		}
	}
}

func TestTokenTypeNameDefault(t *testing.T) {
	got := tokenTypeName(TokenType(9999))
	if got != "unknown" {
		t.Errorf("tokenTypeName(9999) = %q, want unknown", got)
	}
}

func TestNilObjectInspect(t *testing.T) {
	var n NilObject
	if n.Inspect() != "nil" {
		t.Errorf("NilObject.Inspect() = %q", n.Inspect())
	}
	if n.Type() != ObjNil {
		t.Errorf("NilObject.Type() = %q", n.Type())
	}
}

func TestBuiltinFnInspect(t *testing.T) {
	b := BuiltinFn{Name: "myfunc"}
	if b.Inspect() != "builtin:myfunc" {
		t.Errorf("BuiltinFn.Inspect() = %q", b.Inspect())
	}
	if b.Type() != ObjBuiltin {
		t.Errorf("BuiltinFn.Type() = %q", b.Type())
	}
}

func TestErrorObjectType(t *testing.T) {
	e := ErrorObject{Message: "test"}
	if e.Type() != ObjError {
		t.Errorf("ErrorObject.Type() = %q", e.Type())
	}
	if e.Inspect() != "error: test" {
		t.Errorf("ErrorObject.Inspect() = %q", e.Inspect())
	}
}

func TestFnObjectInspect(t *testing.T) {
	f := FnObject{Name: "add"}
	if f.Inspect() != "fn add" {
		t.Errorf("FnObject.Inspect() = %q", f.Inspect())
	}
	if f.Type() != ObjFn {
		t.Errorf("FnObject.Type() = %q", f.Type())
	}
}

func TestApplyArithmeticIntAllOps(t *testing.T) {
	vm := NewVM(GetBuiltins(true), true)
	tests := []struct{ a, b Object; op string; want string }{
		{IntObject(10), IntObject(3), "+", "13"},
		{IntObject(10), IntObject(3), "-", "7"},
		{IntObject(10), IntObject(3), "*", "30"},
		{IntObject(10), IntObject(3), "/", "3"},
	}
	for _, tt := range tests {
		result := vm.applyArithmetic(tt.a, tt.b, tt.op)
		if result.Inspect() != tt.want {
			t.Errorf("applyArithmetic(%v, %v, %q) = %s, want %s", tt.a, tt.b, tt.op, result.Inspect(), tt.want)
		}
	}
}

func TestApplyArithmeticStringConcat(t *testing.T) {
	vm := NewVM(GetBuiltins(true), true)
	result := vm.applyArithmetic(StringObject("a"), StringObject("b"), "+")
	if result.Inspect() != "ab" {
		t.Errorf("string concat = %q", result.Inspect())
	}
}

func TestApplyArithmeticDivZero(t *testing.T) {
	vm := NewVM(GetBuiltins(true), true)
	result := vm.applyArithmetic(IntObject(1), IntObject(0), "/")
	if _, ok := result.(NilObject); !ok {
		t.Errorf("div by zero should return NilObject, got %T", result)
	}
}

func TestApplyArithmeticFloatDivZero(t *testing.T) {
	vm := NewVM(GetBuiltins(true), true)
	result := vm.applyArithmetic(FloatObject(1.0), FloatObject(0.0), "/")
	if _, ok := result.(NilObject); !ok {
		t.Errorf("float div by zero should return NilObject, got %T", result)
	}
}

func TestApplyArithmeticTypeMismatch(t *testing.T) {
	vm := NewVM(GetBuiltins(true), true)
	result := vm.applyArithmetic(IntObject(1), StringObject("x"), "+")
	if _, ok := result.(NilObject); !ok {
		t.Errorf("type mismatch should return NilObject, got %T", result)
	}
}

func TestApplyArithmeticStringNonConcat(t *testing.T) {
	vm := NewVM(GetBuiltins(true), true)
	result := vm.applyArithmetic(StringObject("a"), IntObject(1), "+")
	if _, ok := result.(NilObject); !ok {
		t.Errorf("string+int should return NilObject, got %T", result)
	}
}

func TestApplyArithmeticStringSub(t *testing.T) {
	vm := NewVM(GetBuiltins(true), true)
	result := vm.applyArithmetic(StringObject("a"), StringObject("b"), "-")
	if _, ok := result.(NilObject); !ok {
		t.Errorf("string-string should return NilObject, got %T", result)
	}
}

func TestApplyArithmeticFloatOps(t *testing.T) {
	vm := NewVM(GetBuiltins(true), true)
	result := vm.applyArithmetic(FloatObject(3.14), FloatObject(1.0), "+")
	switch result.(type) {
	case FloatObject, IntObject:
	default:
		t.Errorf("3.14+1.0 should be numeric, got %T", result)
	}
}

func TestApplyArithmeticFloatPreserved(t *testing.T) {
	vm := NewVM(GetBuiltins(true), true)
	result := vm.applyArithmetic(FloatObject(3.0), FloatObject(1.0), "+")
	if _, ok := result.(FloatObject); !ok {
		t.Errorf("FloatObject(3.0)+FloatObject(1.0) should be FloatObject, got %T: %s", result, result.Inspect())
	}
}

func TestApplyArithmeticMixedIntFloat(t *testing.T) {
	vm := NewVM(GetBuiltins(true), true)
	result := vm.applyArithmetic(IntObject(3), FloatObject(1.5), "+")
	if _, ok := result.(FloatObject); !ok {
		t.Errorf("IntObject(3)+FloatObject(1.5) should be FloatObject, got %T: %s", result, result.Inspect())
	}
}

func TestParsePowerExpression(t *testing.T) {
	_, outputs := runEngineTest(t, `print(str_from_int(math_pow(2, 8)))`)
	if len(outputs) != 1 || outputs[0] != "256" {
		t.Errorf("outputs = %v", outputs)
	}
}

func TestParseExprStatementLet(t *testing.T) {
	_, outputs := runEngineTest(t, `let x = 100
print(str_from_int(x))`)
	if outputs[0] != "100" {
		t.Errorf("outputs = %v", outputs)
	}
}

func TestObjectToJSONNested(t *testing.T) {
	inner := MapObject{Pairs: map[string]Object{"a": IntObject(1)}, Keys: []string{"a"}}
	outer := MapObject{Pairs: map[string]Object{"inner": inner}, Keys: []string{"inner"}}
	result := objectToJSON(outer)
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if _, ok := m["inner"]; !ok {
		t.Error("expected 'inner' key")
	}
}

func TestObjectToJSONNestedList(t *testing.T) {
	inner := ListObject{Elements: []Object{IntObject(1), IntObject(2)}}
	outer := ListObject{Elements: []Object{inner, StringObject("x")}}
	result := objectToJSON(outer)
	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", result)
	}
	if len(arr) != 2 {
		t.Errorf("len = %d, want 2", len(arr))
	}
}

func TestReadStringWithEscapeSequences(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"hello\nworld"`, "hello\nworld"},
		{`"tab\there"`, "tab\there"},
		{`"quote\"inside\""`, "quote\"inside\""},
		{`"backslash\\path"`, "backslash\\path"},
	}
	for _, tt := range tests {
		lexer := NewLexer(tt.input)
		tokens := lexer.Tokenize()
		if tokens[0].Literal != tt.want {
			t.Errorf("input=%q got=%q want=%q", tt.input, tokens[0].Literal, tt.want)
		}
	}
}

func strContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
