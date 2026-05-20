package script

import (
	"strings"
	"testing"
)

func runEngineTest(t *testing.T, code string) (Object, []string) {
	t.Helper()
	lexer := NewLexer(code)
	tokens := lexer.Tokenize()
	parser := NewParser(tokens)
	ast, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	compiler := NewCompiler()
	if err := compiler.Compile(ast); err != nil {
		t.Fatalf("compile error: %v", err)
	}
	builtins := GetBuiltins(true)
	vm := NewVMWithGlobals(builtins, true, compiler.GlobalNames())
	PrintCallback = func(s string) {
		vm.AddOutput(s)
	}
	result, err := vm.Run(compiler.Instructions(), compiler.Constants())
	if err != nil {
		t.Fatalf("runtime error: %v", err)
	}
	return result, vm.Outputs()
}

func runEngineExpectError(t *testing.T, code string) error {
	t.Helper()
	lexer := NewLexer(code)
	tokens := lexer.Tokenize()
	parser := NewParser(tokens)
	ast, err := parser.Parse()
	if err != nil {
		return err
	}
	compiler := NewCompiler()
	if err := compiler.Compile(ast); err != nil {
		return err
	}
	builtins := GetBuiltins(true)
	vm := NewVMWithGlobals(builtins, true, compiler.GlobalNames())
	PrintCallback = func(s string) {
		vm.AddOutput(s)
	}
	_, err = vm.Run(compiler.Instructions(), compiler.Constants())
	return err
}

func TestIfWithoutElse(t *testing.T) {
	_, outputs := runEngineTest(t, `let x = 10
if x > 5 {
    print("yes")
}`)
	if len(outputs) != 1 || outputs[0] != "yes" {
		t.Errorf("outputs = %v, want [yes]", outputs)
	}
}

func TestFnNoReturn(t *testing.T) {
	result, _ := runEngineTest(t, `fn nothing() {
    let x = 1
}
nothing()`)
	if _, ok := result.(NilObject); !ok {
		t.Errorf("expected NilObject, got %T", result)
	}
}

func TestFnMultipleReturns(t *testing.T) {
	_, outputs := runEngineTest(t, `fn check(n) {
    if n > 10 { return "big" }
    if n > 5 { return "medium" }
    return "small"
}
print(check(20))
print(check(7))
print(check(1))`)
	if outputs[0] != "big" || outputs[1] != "medium" || outputs[2] != "small" {
		t.Errorf("outputs = %v", outputs)
	}
}

func TestFnRecursive(t *testing.T) {
	result, _ := runEngineTest(t, `fn fact(n) {
    if n <= 1 { return 1 }
    return n * fact(n - 1)
}
fact(5)`)
	if i, ok := result.(IntObject); !ok || int64(i) != 120 {
		t.Errorf("fact(5) = %v, want 120", result)
	}
}

func TestHigherOrderFn(t *testing.T) {
	_, outputs := runEngineTest(t, `fn apply(f, x) {
    return f(x)
}
fn double(n) { return n * 2 }
print(str_from_int(apply(double, 5)))`)
	if len(outputs) != 1 || outputs[0] != "10" {
		t.Errorf("outputs = %v, want [10]", outputs)
	}
}

func TestListNested(t *testing.T) {
	_, outputs := runEngineTest(t, `let m = [[1,2],[3,4]]
print(str_from_int(list_get(list_get(m, 0), 1)))`)
	if len(outputs) != 1 || outputs[0] != "2" {
		t.Errorf("outputs = %v, want [2]", outputs)
	}
}

func TestListPushFunctional(t *testing.T) {
	_, outputs := runEngineTest(t, `let items = [1, 2]
let items2 = list_push(items, 3)
print(str_from_int(list_len(items2)))
print(str_from_int(list_get(items2, 2)))`)
	if outputs[0] != "3" || outputs[1] != "3" {
		t.Errorf("outputs = %v", outputs)
	}
}

func TestMapSetFunctional(t *testing.T) {
	_, outputs := runEngineTest(t, `let m = {"a": "1"}
let m2 = map_set(m, "b", "2")
print(str_from_int(map_len(m2)))
print(map_get(m2, "a"))
print(map_has(m2, "b"))`)
	if outputs[0] != "2" || outputs[1] != "1" || outputs[2] != "true" {
		t.Errorf("outputs = %v", outputs)
	}
}

func TestMapAllOps(t *testing.T) {
	_, outputs := runEngineTest(t, `let m = {"a": "1"}
let m2 = map_set(m, "b", "2")
print(str_from_int(map_len(m2)))
print(map_get(m2, "a"))
print(map_has(m2, "b"))
let keys = map_keys(m2)
print(str_from_int(list_len(keys)))
let m3 = map_del(m2, "a")
print(str_from_int(map_len(m3)))`)
	if outputs[0] != "2" || outputs[1] != "1" || outputs[2] != "true" || outputs[3] != "2" {
		t.Errorf("outputs = %v", outputs)
	}
}

func TestStringSplitJoin(t *testing.T) {
	_, outputs := runEngineTest(t, `let parts = str_split("a,b,c", ",")
print(str_from_int(list_len(parts)))
let joined = str_join(parts, "-")
print(joined)`)
	if outputs[0] != "3" || outputs[1] != "a-b-c" {
		t.Errorf("outputs = %v", outputs)
	}
}

func TestStringSubRepeatReverse(t *testing.T) {
	result, _ := runEngineTest(t, `str_sub("hello", 1, 3)`)
	if result.Inspect() != "el" {
		t.Errorf("str_sub = %q, want el", result.Inspect())
	}

	result, _ = runEngineTest(t, `str_repeat("ab", 3)`)
	if result.Inspect() != "ababab" {
		t.Errorf("str_repeat = %q, want ababab", result.Inspect())
	}

	result, _ = runEngineTest(t, `str_reverse("abc")`)
	if result.Inspect() != "cba" {
		t.Errorf("str_reverse = %q, want cba", result.Inspect())
	}
}

func TestListSlice(t *testing.T) {
	_, outputs := runEngineTest(t, `let items = [0, 1, 2, 3, 4]
let sub = list_slice(items, 1, 3)
print(str_from_int(list_len(sub)))
print(str_from_int(list_get(sub, 0)))`)
	if outputs[0] != "2" || outputs[1] != "1" {
		t.Errorf("outputs = %v", outputs)
	}
}

func TestListSort(t *testing.T) {
	_, outputs := runEngineTest(t, `let items = [3, 1, 2]
let sorted = list_sort(items)
print(str_from_int(list_get(sorted, 0)))
print(str_from_int(list_get(sorted, 2)))`)
	if outputs[0] != "1" || outputs[1] != "3" {
		t.Errorf("outputs = %v", outputs)
	}
}

func TestConvToIntFromInt(t *testing.T) {
	result, _ := runEngineTest(t, `conv_to_int(42)`)
	if result.Inspect() != "42" {
		t.Errorf("conv_to_int(42) = %q, want 42", result.Inspect())
	}
}

func TestConvToFloatFromFloat(t *testing.T) {
	result, _ := runEngineTest(t, `conv_to_float(3.14)`)
	if result.Inspect() != "3.14" {
		t.Errorf("conv_to_float(3.14) = %q, want 3.14", result.Inspect())
	}
}

func TestConvToString(t *testing.T) {
	result, _ := runEngineTest(t, `conv_to_string(42)`)
	if result.Inspect() != "42" {
		t.Errorf("conv_to_string(42) = %q, want 42", result.Inspect())
	}
}

func TestConvToBool(t *testing.T) {
	tests := []struct {
		code   string
		expect string
	}{
		{`conv_to_bool(1)`, "true"},
		{`conv_to_bool(0)`, "false"},
		{`conv_to_bool("")`, "false"},
	}
	for _, tt := range tests {
		result, _ := runEngineTest(t, tt.code)
		got := result.Inspect()
		if got != tt.expect {
			t.Errorf("%s = %q, want %q", tt.code, got, tt.expect)
		}
	}
}

func TestConvHexBuiltins(t *testing.T) {
	_, outputs := runEngineTest(t, `let h = conv_hex_encode("hello")
print(h)
let s = conv_hex_decode(h)
print(s)`)
	if outputs[0] != "68656c6c6f" || outputs[1] != "hello" {
		t.Errorf("outputs = %v", outputs)
	}
}

func TestConvB64Builtins(t *testing.T) {
	_, outputs := runEngineTest(t, `let b = conv_b64_encode("hello")
print(b)
let s = conv_b64_decode(b)
print(s)`)
	if outputs[0] != "aGVsbG8=" || outputs[1] != "hello" {
		t.Errorf("outputs = %v", outputs)
	}
}

func TestTimeBuiltin(t *testing.T) {
	result, _ := runEngineTest(t, `time_now_unix()`)
	if _, ok := result.(IntObject); !ok {
		t.Errorf("expected IntObject, got %T", result)
	}
}

func TestJSONEncode(t *testing.T) {
	result, _ := runEngineTest(t, `json_encode({"key": "value"})`)
	s := result.Inspect()
	if !strings.Contains(s, "key") {
		t.Errorf("json_encode result = %q, expected to contain 'key'", s)
	}
}

func TestTypeCheckBuiltins(t *testing.T) {
	tests := []struct {
		code   string
		expect string
	}{
		{`type_is_nil(nil)`, "true"},
		{`type_is_bool(true)`, "true"},
		{`type_is_int(42)`, "true"},
		{`type_is_string("hi")`, "true"},
		{`type_is_list([1])`, "true"},
		{`type_is_map({"a":"b"})`, "true"},
		{`type_is_int("no")`, "false"},
		{`type_is_fn(fn(){})`, "true"},
	}
	for _, tt := range tests {
		result, _ := runEngineTest(t, tt.code)
		got := result.Inspect()
		if got != tt.expect {
			t.Errorf("%s = %q, want %q", tt.code, got, tt.expect)
		}
	}
}

func TestIsErrorBuiltin(t *testing.T) {
	result, _ := runEngineTest(t, `is_error(nil)`)
	if result.Inspect() != "false" {
		t.Errorf("is_error(nil) = %q, want false", result.Inspect())
	}
}

func TestEmptyBlock(t *testing.T) {
	result, _ := runEngineTest(t, `if true {}`)
	if _, ok := result.(NilObject); !ok {
		t.Errorf("expected NilObject for empty block, got %T", result)
	}
}

func TestNestedFunctionCall(t *testing.T) {
	_, outputs := runEngineTest(t, `fn double(n) { return n * 2 }
fn inc(n) { return n + 1 }
print(str_from_int(double(inc(3))))`)
	if len(outputs) != 1 || outputs[0] != "8" {
		t.Errorf("outputs = %v, want [8]", outputs)
	}
}

func TestFloatResult(t *testing.T) {
	result, _ := runEngineTest(t, `math_sqrt(2)`)
	if _, ok := result.(FloatObject); !ok {
		t.Errorf("math_sqrt(2) should be FloatObject, got %T", result)
	}
}

func TestNegateFloat(t *testing.T) {
	result, _ := runEngineTest(t, `-3.14`)
	if f, ok := result.(FloatObject); !ok || float64(f) >= 0 {
		t.Errorf("expected negative float, got %v", result)
	}
}

func TestComparisonInts(t *testing.T) {
	tests := []struct {
		code   string
		expect bool
	}{
		{"5 < 10", true},
		{"10 < 5", false},
		{"5 > 10", false},
		{"10 > 5", true},
		{"5 <= 5", true},
		{"5 >= 5", true},
		{"5 == 5", true},
		{"5 != 5", false},
	}
	for _, tt := range tests {
		result, _ := runEngineTest(t, tt.code)
		if b, ok := result.(BoolObject); !ok || bool(b) != tt.expect {
			t.Errorf("%s = %v, want %v", tt.code, result, tt.expect)
		}
	}
}

func TestDivisionByZeroReturnsNil(t *testing.T) {
	result, _ := runEngineTest(t, `10 / 0`)
	if _, ok := result.(NilObject); !ok {
		t.Errorf("10/0 should return NilObject, got %T", result)
	}
}

func TestTypeMismatchReturnsNil(t *testing.T) {
	result, _ := runEngineTest(t, `"hello" + 5`)
	if _, ok := result.(NilObject); !ok {
		t.Errorf("type mismatch should return NilObject, got %T", result)
	}
}

func TestUndefinedVarReturnsNil(t *testing.T) {
	result, _ := runEngineTest(t, `undefined_var`)
	if _, ok := result.(NilObject); !ok {
		t.Errorf("undefined var should return NilObject, got %T", result)
	}
}

func TestCallNonFunctionErrors(t *testing.T) {
	err := runEngineExpectError(t, `let x = 42
x()`)
	if err == nil {
		t.Error("expected error for calling non-function")
	}
}

func TestParseErrorBadSyntax(t *testing.T) {
	err := runEngineExpectError(t, `fn (`)
	if err == nil {
		t.Error("expected parse error for bad syntax")
	}
}

func TestMultipleLet(t *testing.T) {
	_, outputs := runEngineTest(t, `let a = 1
let b = 2
let c = a + b
print(str_from_int(c))`)
	if len(outputs) != 1 || outputs[0] != "3" {
		t.Errorf("outputs = %v, want [3]", outputs)
	}
}

func TestNestedIf(t *testing.T) {
	_, outputs := runEngineTest(t, `let x = 15
if x > 10 {
    if x > 20 {
        print("very big")
    } else {
        print("medium")
    }
}`)
	if len(outputs) != 1 || outputs[0] != "medium" {
		t.Errorf("outputs = %v, want [medium]", outputs)
	}
}

func TestBoolTruthiness(t *testing.T) {
	tests := []struct {
		code   string
		expect bool
	}{
		{"if 0 { true } else { false }", false},
		{"if 1 { true } else { false }", true},
		{`if "" { true } else { false }`, false},
		{`if "x" { true } else { false }`, true},
		{"if nil { true } else { false }", false},
	}
	for _, tt := range tests {
		result, _ := runEngineTest(t, tt.code)
		if b, ok := result.(BoolObject); !ok || bool(b) != tt.expect {
			t.Errorf("%s = %v, want %v", tt.code, result, tt.expect)
		}
	}
}

func TestObjectTypes(t *testing.T) {
	tests := []struct {
		obj   Object
		type_ ObjectType
	}{
		{NilObject{}, ObjNil},
		{BoolObject(true), ObjBool},
		{IntObject(42), ObjInt},
		{FloatObject(3.14), ObjFloat},
		{StringObject("hi"), ObjString},
		{ListObject{Elements: []Object{}}, ObjList},
		{MapObject{Pairs: map[string]Object{}}, ObjMap},
	}
	for _, tt := range tests {
		if tt.obj.Type() != tt.type_ {
			t.Errorf("Type() = %q, want %q", tt.obj.Type(), tt.type_)
		}
	}
}

func TestIsTruthy(t *testing.T) {
	tests := []struct {
		obj    Object
		expect bool
	}{
		{NilObject{}, false},
		{BoolObject(true), true},
		{BoolObject(false), false},
		{IntObject(1), true},
		{IntObject(0), false},
		{StringObject(""), false},
		{StringObject("x"), true},
	}
	for _, tt := range tests {
		if IsTruthy(tt.obj) != tt.expect {
			t.Errorf("IsTruthy(%v) = %v, want %v", tt.obj, !tt.expect, tt.expect)
		}
	}
}

func TestOpCodeString(t *testing.T) {
	names := map[OpCode]string{
		OpConstant: "CONSTANT", OpInt: "INT", OpNil: "NIL", OpTrue: "TRUE",
		OpAdd: "ADD", OpSub: "SUB", OpMul: "MUL", OpDiv: "DIV",
		OpEq: "EQ", OpJump: "JUMP", OpReturn: "RETURN", OpCall: "CALL",
		OpGetFree: "GET_FREE", OpClosure: "CLOSURE",
	}
	for op, want := range names {
		if op.String() != want {
			t.Errorf("OpCode(%d).String() = %q, want %q", op, op.String(), want)
		}
	}
}

func TestMapMerge(t *testing.T) {
	_, outputs := runEngineTest(t, `let a = {"x": "1"}
let b = {"y": "2"}
let c = map_merge(a, b)
print(str_from_int(map_len(c)))`)
	if len(outputs) != 1 || outputs[0] != "2" {
		t.Errorf("outputs = %v, want [2]", outputs)
	}
}

func TestStrInterp(t *testing.T) {
	_, outputs := runEngineTest(t, `let name = "world"
print(str_interp("Hello ${name}", {"name": name}))`)
	if len(outputs) != 1 || outputs[0] != "Hello world" {
		t.Errorf("outputs = %v, want [Hello world]", outputs)
	}
}

func TestMathBuiltinsExtra(t *testing.T) {
	tests := []struct {
		code   string
		expect string
	}{
		{"math_sqrt(16)", "4"},
		{"math_pow(2, 3)", "8"},
		{"math_round(3)", "3"},
		{"math_mod(10, 3)", "1"},
	}
	for _, tt := range tests {
		result, _ := runEngineTest(t, tt.code)
		got := result.Inspect()
		if got != tt.expect {
			t.Errorf("%s = %q, want %q", tt.code, got, tt.expect)
		}
	}
}

func TestSafeMode(t *testing.T) {
	builtins := GetBuiltins(false)
	if len(builtins) == 0 {
		t.Error("safe mode should have some builtins")
	}
	if _, ok := builtins["os_exec"]; ok {
		t.Error("safe mode should not have os_exec")
	}
}

func TestUnsafeMode(t *testing.T) {
	builtins := GetBuiltins(true)
	if _, ok := builtins["os_exec"]; !ok {
		t.Error("unsafe mode should have os_exec")
	}
}

func TestFibonacci(t *testing.T) {
	_, outputs := runEngineTest(t, `let a = 0
let b = 1
let i = 0
while i < 10 {
    let t = a + b
    a = b
    b = t
    i = i + 1
}
print(str_from_int(a))`)
	if len(outputs) != 1 || outputs[0] != "55" {
		t.Errorf("outputs = %v, want [55]", outputs)
	}
}

func TestIODirBuiltin(t *testing.T) {
	_, _ = runEngineTest(t, `io_exists("/tmp")`)
}

func TestListContains(t *testing.T) {
	_, outputs := runEngineTest(t, `let items = [10, 20, 30]
print(list_contains(items, 20))
print(list_contains(items, 99))`)
	if outputs[0] != "true" || outputs[1] != "false" {
		t.Errorf("outputs = %v", outputs)
	}
}

func TestListReverse(t *testing.T) {
	_, outputs := runEngineTest(t, `let items = [1, 2, 3]
let rev = list_reverse(items)
print(str_from_int(list_get(rev, 0)))`)
	if outputs[0] != "3" {
		t.Errorf("outputs = %v, want [3]", outputs)
	}
}

func TestStringComparisonEqual(t *testing.T) {
	result, _ := runEngineTest(t, `"abc" == "abc"`)
	if b, ok := result.(BoolObject); !ok || !bool(b) {
		t.Errorf("expected true, got %v", result)
	}
}

func TestModulo(t *testing.T) {
	result, _ := runEngineTest(t, `10 % 3`)
	if i, ok := result.(IntObject); !ok || int64(i) != 1 {
		t.Errorf("10 %% 3 = %v, want 1", result)
	}
}

func TestPowOperator(t *testing.T) {
	result, _ := runEngineTest(t, `math_pow(2, 10)`)
	if _, ok := result.(FloatObject); !ok {
		t.Errorf("math_pow(2,10) = %v, want FloatObject", result)
	}
}

func TestStringIndexAccess(t *testing.T) {
	result, _ := runEngineTest(t, `let s = "hello"
s[1]`)
	if s, ok := result.(StringObject); !ok || string(s) != "e" {
		t.Errorf("s[1] = %v, want 'e'", result)
	}
}

func TestNestedFnCalls(t *testing.T) {
	_, outputs := runEngineTest(t, `fn add(a, b) { return a + b }
fn mul(a, b) { return a * b }
print(str_from_int(mul(add(2, 3), add(1, 1))))`)
	if len(outputs) != 1 || outputs[0] != "10" {
		t.Errorf("outputs = %v, want [10]", outputs)
	}
}

func TestGCD(t *testing.T) {
	_, outputs := runEngineTest(t, `let a = 48
let b = 18
while b != 0 {
    let t = b
    b = a % b
    a = t
}
print(str_from_int(a))`)
	if len(outputs) != 1 || outputs[0] != "6" {
		t.Errorf("outputs = %v, want [6]", outputs)
	}
}

func TestErrorObject(t *testing.T) {
	err := ErrorObject{Message: "test error"}
	if err.Type() != ObjError {
		t.Errorf("ErrorObject.Type() = %q, want %q", err.Type(), ObjError)
	}
	if !strings.Contains(err.Inspect(), "test error") {
		t.Errorf("ErrorObject.Inspect() = %q, should contain message", err.Inspect())
	}
}

func TestBuiltinFnType(t *testing.T) {
	b := BuiltinFn{Name: "test"}
	if b.Type() != ObjBuiltin {
		t.Errorf("BuiltinFn.Type() = %q, want %q", b.Type(), ObjBuiltin)
	}
}

func TestFnObjectType(t *testing.T) {
	f := FnObject{Name: "myfn"}
	if f.Type() != ObjFn {
		t.Errorf("FnObject.Type() = %q, want %q", f.Type(), ObjFn)
	}
	if f.Inspect() != "fn myfn" {
		t.Errorf("FnObject.Inspect() = %q, want 'fn myfn'", f.Inspect())
	}
}

func TestListObjectInspect(t *testing.T) {
	l := ListObject{Elements: []Object{IntObject(1), StringObject("hello")}}
	s := l.Inspect()
	if !strings.Contains(s, "1") || !strings.Contains(s, "hello") {
		t.Errorf("ListObject.Inspect() = %q", s)
	}
}

func TestMapObjectInspect(t *testing.T) {
	m := MapObject{
		Pairs: map[string]Object{"key": StringObject("val")},
		Keys:  []string{"key"},
	}
	s := m.Inspect()
	if !strings.Contains(s, "key") || !strings.Contains(s, "val") {
		t.Errorf("MapObject.Inspect() = %q", s)
	}
}

func TestVMOpCount(t *testing.T) {
	lexer := NewLexer(`1 + 2`)
	tokens := lexer.Tokenize()
	parser := NewParser(tokens)
	ast, _ := parser.Parse()
	compiler := NewCompiler()
	compiler.Compile(ast)
	builtins := GetBuiltins(true)
	vm := NewVMWithGlobals(builtins, true, compiler.GlobalNames())
	PrintCallback = func(s string) { vm.AddOutput(s) }
	vm.Run(compiler.Instructions(), compiler.Constants())
	if vm.OpCount() == 0 {
		t.Error("OpCount should be > 0 after execution")
	}
}

func TestVMOutputs(t *testing.T) {
	lexer := NewLexer(`print("hello")`)
	tokens := lexer.Tokenize()
	parser := NewParser(tokens)
	ast, _ := parser.Parse()
	compiler := NewCompiler()
	compiler.Compile(ast)
	builtins := GetBuiltins(true)
	vm := NewVMWithGlobals(builtins, true, compiler.GlobalNames())
	PrintCallback = func(s string) { vm.AddOutput(s) }
	vm.Run(compiler.Instructions(), compiler.Constants())
	outputs := vm.Outputs()
	if len(outputs) != 1 || outputs[0] != "hello" {
		t.Errorf("Outputs = %v, want [hello]", outputs)
	}
}

func TestBreakInWhile(t *testing.T) {
	_, outputs := runEngineTest(t, `
let sum = 0
let i = 0
while i < 10 {
    if i == 5 {
        break
    }
    sum = sum + i
    i = i + 1
}
print(str_from_int(sum))
`)
	want := "10"
	if len(outputs) < 1 || outputs[len(outputs)-1] != want {
		t.Errorf("break while: got %v, want %s", outputs, want)
	}
}

func TestBreakInFor(t *testing.T) {
	_, outputs := runEngineTest(t, `
let sum = 0
for x in [1, 2, 3, 4, 5] {
    if x == 4 {
        break
    }
    sum = sum + x
}
print(str_from_int(sum))
`)
	want := "6"
	if len(outputs) < 1 || outputs[len(outputs)-1] != want {
		t.Errorf("break for: got %v, want %s", outputs, want)
	}
}

func TestContinueInWhile(t *testing.T) {
	_, outputs := runEngineTest(t, `
let sum = 0
let i = 0
while i < 5 {
    i = i + 1
    if i == 3 {
        continue
    }
    sum = sum + i
}
print(str_from_int(sum))
`)
	want := "12"
	if len(outputs) < 1 || outputs[len(outputs)-1] != want {
		t.Errorf("continue while: got %v, want %s", outputs, want)
	}
}

func TestContinueInFor(t *testing.T) {
	_, outputs := runEngineTest(t, `
let result = ""
for x in [1, 2, 3, 4, 5] {
    if x == 3 {
        continue
    }
    result = str_concat(result, str_from_int(x))
}
print(result)
`)
	want := "1245"
	if len(outputs) < 1 || outputs[len(outputs)-1] != want {
		t.Errorf("continue for: got %v, want %s", outputs, want)
	}
}

func TestBreakContinueCombined(t *testing.T) {
	_, outputs := runEngineTest(t, `
let result = ""
let i = 0
while i < 10 {
    i = i + 1
    if i == 3 {
        continue
    }
    if i == 7 {
        break
    }
    result = str_concat(result, str_from_int(i))
}
print(result)
`)
	want := "12456"
	if len(outputs) < 1 || outputs[len(outputs)-1] != want {
		t.Errorf("break+continue: got %v, want %s", outputs, want)
	}
}

func TestBreakOutsideLoopError(t *testing.T) {
	err := runEngineExpectError(t, `break`)
	if err == nil {
		t.Error("break outside loop should error")
	}
}

func TestContinueOutsideLoopError(t *testing.T) {
	err := runEngineExpectError(t, `continue`)
	if err == nil {
		t.Error("continue outside loop should error")
	}
}

func TestFloatArithmeticPreserved(t *testing.T) {
	result, _ := runEngineTest(t, `3.0 + 1.0`)
	if _, ok := result.(FloatObject); !ok {
		t.Errorf("3.0+1.0 should be FloatObject, got %T: %s", result, result.Inspect())
	}
	if result.Inspect() != "4" {
		t.Errorf("3.0+1.0 = %s, want 4", result.Inspect())
	}
}

func TestFloatIntMixedArithmetic(t *testing.T) {
	result, _ := runEngineTest(t, `3 + 1.5`)
	if _, ok := result.(FloatObject); !ok {
		t.Errorf("3+1.5 should be FloatObject, got %T: %s", result, result.Inspect())
	}
}

func TestListMapWithClosure(t *testing.T) {
	result, _ := runEngineTest(t, `
let items = [1, 2, 3]
fn double(x) { return x * 2 }
let result = list_map(items, double)
result`)
	if result.Inspect() != "[2, 4, 6]" {
		t.Errorf("list_map with closure = %s", result.Inspect())
	}
}

func TestListFilterWithClosure(t *testing.T) {
	result, _ := runEngineTest(t, `
let items = [1, 2, 3, 4, 5]
fn isEven(x) { return x % 2 == 0 }
let result = list_filter(items, isEven)
result`)
	if result.Inspect() != "[2, 4]" {
		t.Errorf("list_filter with closure = %s", result.Inspect())
	}
}

func TestListReduceWithClosure(t *testing.T) {
	result, _ := runEngineTest(t, `
let items = [1, 2, 3, 4]
fn add(a, b) { return a + b }
let result = list_reduce(items, add, 0)
result`)
	if result.Inspect() != "10" {
		t.Errorf("list_reduce with closure = %s", result.Inspect())
	}
}

func TestListFindWithClosure(t *testing.T) {
	result, _ := runEngineTest(t, `
let items = [1, 2, 3, 4]
fn isThree(x) { return x == 3 }
let result = list_find(items, isThree)
result`)
	if result.Inspect() != "3" {
		t.Errorf("list_find with closure = %s", result.Inspect())
	}
}

func TestCompoundAddAssign(t *testing.T) {
	result, _ := runEngineTest(t, `let x = 10
x += 5
x`)
	if result.Inspect() != "15" {
		t.Errorf("x += 5 = %s, want 15", result.Inspect())
	}
}

func TestCompoundSubAssign(t *testing.T) {
	result, _ := runEngineTest(t, `let x = 10
x -= 3
x`)
	if result.Inspect() != "7" {
		t.Errorf("x -= 3 = %s, want 7", result.Inspect())
	}
}

func TestCompoundMulAssign(t *testing.T) {
	result, _ := runEngineTest(t, `let x = 10
x *= 3
x`)
	if result.Inspect() != "30" {
		t.Errorf("x *= 3 = %s, want 30", result.Inspect())
	}
}

func TestCompoundDivAssign(t *testing.T) {
	result, _ := runEngineTest(t, `let x = 10
x /= 2
x`)
	if result.Inspect() != "5" {
		t.Errorf("x /= 2 = %s, want 5", result.Inspect())
	}
}

func TestCompoundAssignInLoop(t *testing.T) {
	_, outputs := runEngineTest(t, `
let sum = 0
for x in [1, 2, 3, 4, 5] {
    sum += x
}
print(str_from_int(sum))
`)
	if len(outputs) < 1 || outputs[len(outputs)-1] != "15" {
		t.Errorf("sum += in for = %v, want 15", outputs)
	}
}

func TestForInString(t *testing.T) {
	_, outputs := runEngineTest(t, `
let result = ""
for ch in "abc" {
    result = str_concat(result, ch)
}
print(result)
`)
	if len(outputs) < 1 || outputs[len(outputs)-1] != "abc" {
		t.Errorf("for in string = %v, want abc", outputs)
	}
}

func TestForInUnicodeString(t *testing.T) {
	_, outputs := runEngineTest(t, `
let result = ""
for ch in "你好" {
    result = str_concat(result, ch)
}
print(result)
`)
	if len(outputs) < 1 || outputs[len(outputs)-1] != "你好" {
		t.Errorf("for in unicode = %v, want 你好", outputs)
	}
}

func TestMutableClosure(t *testing.T) {
	result, _ := runEngineTest(t, `
let count = 0
fn increment() {
    count += 1
    return count
}
let a = increment()
let b = increment()
let c = increment()
str_concat(str_from_int(a), str_concat(str_from_int(b), str_from_int(c)))
`)
	if result.Inspect() != "123" {
		t.Errorf("mutable closure = %s, want 123", result.Inspect())
	}
}

func TestStrLen(t *testing.T) {
	result, _ := runEngineTest(t, `str_len("hello")`)
	if result.Inspect() != "5" {
		t.Errorf("str_len(hello) = %s, want 5", result.Inspect())
	}
}

func TestStrLenUnicode(t *testing.T) {
	result, _ := runEngineTest(t, `str_len("你好")`)
	if result.Inspect() != "2" {
		t.Errorf("str_len(你好) = %s, want 2", result.Inspect())
	}
}

func TestStrLenNonString(t *testing.T) {
	result, _ := runEngineTest(t, `str_len(42)`)
	if result.Inspect() != "0" {
		t.Errorf("str_len(42) = %s, want 0", result.Inspect())
	}
}

func TestStrSubUnicode(t *testing.T) {
	result, _ := runEngineTest(t, `str_sub("你好世界", 1, 3)`)
	if result.Inspect() != "好世" {
		t.Errorf("str_sub unicode = %s, want 好世", result.Inspect())
	}
}

func TestTryWithScriptFn(t *testing.T) {
	result, _ := runEngineTest(t, `
fn safeDiv(a, b) {
    if b == 0 {
        return error("division by zero")
    }
    return a / b
}
let r = try(safeDiv, 10, 0)
r`)
	if result.Type() != ObjList {
		t.Errorf("try result type = %s, want list", result.Type())
	}
}

func TestTryWithScriptFnSuccess(t *testing.T) {
	result, _ := runEngineTest(t, `
fn safeDiv(a, b) {
    if b == 0 {
        return error("division by zero")
    }
    return a / b
}
let r = try(safeDiv, 10, 2)
r`)
	l := result.(ListObject)
	if len(l.Elements) < 1 {
		t.Fatal("expected list with 2 elements")
	}
	if l.Elements[0].Inspect() != "true" {
		t.Errorf("try success = %s, want true", l.Elements[0].Inspect())
	}
}

func TestStackLeakInLoop(t *testing.T) {
	_, outputs := runEngineTest(t, `
let sum = 0
let i = 0
while i < 100 {
    print(str_from_int(i))
    i += 1
}
print("done")
`)
	lastOutput := outputs[len(outputs)-1]
	if lastOutput != "done" {
		t.Errorf("stack leak test: last output = %q, want done", lastOutput)
	}
}

func TestClosureCapturesLocal(t *testing.T) {
	_, outputs := runEngineTest(t, `
fn makeCounter() {
    let count = 0
    fn inc() {
        count += 1
        return count
    }
    return inc
}
let counter = makeCounter()
print(str_from_int(counter()))
print(str_from_int(counter()))
print(str_from_int(counter()))
`)
	if len(outputs) != 3 || outputs[0] != "1" || outputs[1] != "2" || outputs[2] != "3" {
		t.Errorf("closure captures local = %v, want [1 2 3]", outputs)
	}
}

func TestFactoryPattern(t *testing.T) {
	result, _ := runEngineTest(t, `
fn makeAdder(x) {
    fn add(y) { return x + y }
    return add
}
let add5 = makeAdder(5)
add5(3)
`)
	if result.Inspect() != "8" {
		t.Errorf("factory pattern = %s, want 8", result.Inspect())
	}
}

func TestErrorBuiltin(t *testing.T) {
	result, _ := runEngineTest(t, `error("something went wrong")`)
	if result.Type() != ObjError {
		t.Errorf("error() type = %s, want error", result.Type())
	}
	if result.Inspect() != "error: something went wrong" {
		t.Errorf("error() = %s", result.Inspect())
	}
}

func TestErrorBuiltinNoArgs(t *testing.T) {
	result, _ := runEngineTest(t, `error()`)
	if result.Type() != ObjError {
		t.Errorf("error() no args type = %s, want error", result.Type())
	}
}

func TestTryWithErrorBuiltin(t *testing.T) {
	err := runEngineExpectError(t, `
fn mayFail(x) {
    if x < 0 {
        return error("negative")
    }
    return x * 2
}
let r = try(mayFail, -1)
r`)
	if err != nil {
		return
	}
}

func TestNegativeListIndex(t *testing.T) {
	result, _ := runEngineTest(t, `[1, 2, 3][-1]`)
	if result.Inspect() != "3" {
		t.Errorf("list[-1] = %s, want 3", result.Inspect())
	}
}

func TestNegativeStringIndex(t *testing.T) {
	result, _ := runEngineTest(t, `"hello"[-1]`)
	if result.Inspect() != "o" {
		t.Errorf("string[-1] = %s, want o", result.Inspect())
	}
}

func TestNegativeListIndexOutOfRange(t *testing.T) {
	result, _ := runEngineTest(t, `[1, 2, 3][-10]`)
	if _, ok := result.(NilObject); !ok {
		t.Errorf("list[-10] should be nil, got %T", result)
	}
}

func TestConstReassignment(t *testing.T) {
	err := runEngineExpectError(t, `const X = 10
X = 20`)
	if err == nil {
		t.Error("const reassignment should error")
	}
}

func TestStrUpperNonString(t *testing.T) {
	result, _ := runEngineTest(t, `str_upper(42)`)
	if result.Inspect() != "" {
		t.Errorf("str_upper(42) = %s, want empty string", result.Inspect())
	}
}

func TestStrTrimNonString(t *testing.T) {
	result, _ := runEngineTest(t, `str_trim(42)`)
	if result.Inspect() != "" {
		t.Errorf("str_trim(42) = %s, want empty string", result.Inspect())
	}
}

func TestStrContainsNonString(t *testing.T) {
	result, _ := runEngineTest(t, `str_contains(42, "4")`)
	if result.Inspect() != "false" {
		t.Errorf("str_contains(42, 4) = %s, want false", result.Inspect())
	}
}

func TestStrReverseUnicode(t *testing.T) {
	result, _ := runEngineTest(t, `str_reverse("你好")`)
	if result.Inspect() != "好你" {
		t.Errorf("str_reverse unicode = %s, want 好你", result.Inspect())
	}
}
