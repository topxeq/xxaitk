package script

import "testing"

func runScript(t *testing.T, code string) (Object, []string) {
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

func TestArithmetic(t *testing.T) {
	tests := []struct {
		code   string
		expect int64
	}{
		{"1 + 2", 3},
		{"10 - 3", 7},
		{"4 * 5", 20},
		{"20 / 4", 5},
		{"10 % 3", 1},
	}
	for _, tt := range tests {
		result, _ := runScript(t, tt.code)
		if i, ok := result.(IntObject); ok {
			if int64(i) != tt.expect {
				t.Errorf("%s = %d, want %d", tt.code, int64(i), tt.expect)
			}
		} else {
			t.Errorf("%s = %v, want IntObject(%d)", tt.code, result, tt.expect)
		}
	}
}

func TestStringConcat(t *testing.T) {
	result, _ := runScript(t, `"hello" + " " + "world"`)
	if s, ok := result.(StringObject); ok {
		if string(s) != "hello world" {
			t.Errorf("got %q, want %q", string(s), "hello world")
		}
	}
}

func TestLetAndVariable(t *testing.T) {
	_, outputs := runScript(t, `let x = 42
print(str_from_int(x))`)
	if len(outputs) != 1 || outputs[0] != "42" {
		t.Errorf("outputs = %v, want [42]", outputs)
	}
}

func TestIfElse(t *testing.T) {
	_, outputs := runScript(t, `let x = 10
if x > 5 {
    print("big")
} else {
    print("small")
}`)
	if len(outputs) != 1 || outputs[0] != "big" {
		t.Errorf("outputs = %v, want [big]", outputs)
	}
}

func TestIfElifElse(t *testing.T) {
	_, outputs := runScript(t, `let x = 3
if x > 5 {
    print("big")
} elif x > 2 {
    print("medium")
} else {
    print("small")
}`)
	if len(outputs) != 1 || outputs[0] != "medium" {
		t.Errorf("outputs = %v, want [medium]", outputs)
	}
}

func TestWhile(t *testing.T) {
	_, outputs := runScript(t, `let i = 0
while i < 3 {
    print(str_from_int(i))
    i = i + 1
}`)
	if len(outputs) != 3 || outputs[0] != "0" || outputs[1] != "1" || outputs[2] != "2" {
		t.Errorf("outputs = %v, want [0 1 2]", outputs)
	}
}

func TestForIn(t *testing.T) {
	_, outputs := runScript(t, `let sum = 0
for i in [1, 2, 3, 4, 5] {
    sum = sum + i
}
print(str_from_int(sum))`)
	if len(outputs) != 1 || outputs[0] != "15" {
		t.Errorf("outputs = %v, want [15]", outputs)
	}
}

func TestFnDecl(t *testing.T) {
	_, outputs := runScript(t, `fn add(a, b) {
    return a + b
}
print(str_from_int(add(3, 7)))`)
	if len(outputs) != 1 || outputs[0] != "10" {
		t.Errorf("outputs = %v, want [10]", outputs)
	}
}

func TestBoolOps(t *testing.T) {
	tests := []struct {
		code   string
		expect bool
	}{
		{"true && true", true},
		{"true && false", false},
		{"false || true", true},
		{"false || false", false},
		{"!true", false},
		{"!false", true},
		{"1 == 1", true},
		{"1 != 2", true},
		{"1 < 2", true},
		{"2 > 1", true},
		{"1 <= 1", true},
		{"1 >= 2", false},
	}
	for _, tt := range tests {
		result, _ := runScript(t, tt.code)
		if b, ok := result.(BoolObject); ok {
			if bool(b) != tt.expect {
				t.Errorf("%s = %v, want %v", tt.code, bool(b), tt.expect)
			}
		} else {
			t.Errorf("%s = %v, want BoolObject(%v)", tt.code, result, tt.expect)
		}
	}
}

func TestListOps(t *testing.T) {
	_, outputs := runScript(t, `let items = [10, 20, 30]
print(str_from_int(list_len(items)))
print(str_from_int(list_get(items, 1)))
print(str_from_int(list_index(items, 30)))`)
	if outputs[0] != "3" || outputs[1] != "20" || outputs[2] != "2" {
		t.Errorf("outputs = %v", outputs)
	}
}

func TestMapOps(t *testing.T) {
	_, outputs := runScript(t, `let m = {"name": "alice", "age": "30"}
print(map_get(m, "name"))
print(str_from_int(map_len(m)))`)
	if outputs[0] != "alice" || outputs[1] != "2" {
		t.Errorf("outputs = %v", outputs)
	}
}

func TestJsonOps(t *testing.T) {
	_, outputs := runScript(t, `let data = json_decode("{\"key\": \"value\"}")
print(json_get(data, "key"))`)
	if len(outputs) != 1 || outputs[0] != "value" {
		t.Errorf("outputs = %v, want [value]", outputs)
	}
}

func TestStrBuiltins(t *testing.T) {
	tests := []struct {
		code   string
		expect string
	}{
		{`str_len("hello")`, "5"},
		{`str_upper("hello")`, "HELLO"},
		{`str_lower("HELLO")`, "hello"},
		{`str_trim("  hi  ")`, "hi"},
		{`str_has_prefix("hello", "he")`, "true"},
		{`str_has_suffix("hello", "lo")`, "true"},
		{`str_contains("hello", "ell")`, "true"},
		{`str_replace("hello", "l", "r", 1)`, "herlo"},
	}
	for _, tt := range tests {
		result, _ := runScript(t, tt.code)
		got := result.Inspect()
		if got != tt.expect {
			t.Errorf("%s = %q, want %q", tt.code, got, tt.expect)
		}
	}
}

func TestMathBuiltins(t *testing.T) {
	tests := []struct {
		code   string
		expect string
	}{
		{"math_abs(-42)", "42"},
		{"math_max(10, 20)", "20"},
		{"math_min(10, 20)", "10"},
		{"math_floor(3.7)", "3"},
		{"math_ceil(3.2)", "4"},
	}
	for _, tt := range tests {
		result, _ := runScript(t, tt.code)
		got := result.Inspect()
		if got != tt.expect {
			t.Errorf("%s = %q, want %q", tt.code, got, tt.expect)
		}
	}
}

func TestTypeBuiltins(t *testing.T) {
	tests := []struct {
		code   string
		expect string
	}{
		{`type_of(42)`, "int"},
		{`type_of("hello")`, "string"},
		{`type_of(true)`, "bool"},
		{`type_of(nil)`, "nil"},
		{`type_of([1,2])`, "list"},
		{`type_of({"a":"b"})`, "map"},
	}
	for _, tt := range tests {
		result, _ := runScript(t, tt.code)
		got := result.Inspect()
		if got != tt.expect {
			t.Errorf("%s = %q, want %q", tt.code, got, tt.expect)
		}
	}
}

func TestNegate(t *testing.T) {
	result, _ := runScript(t, "-5")
	if i, ok := result.(IntObject); ok {
		if int64(i) != -5 {
			t.Errorf("got %d, want -5", int64(i))
		}
	}
}

func TestIndexExpr(t *testing.T) {
	_, outputs := runScript(t, `let items = [10, 20, 30]
print(str_from_int(items[1]))`)
	if len(outputs) != 1 || outputs[0] != "20" {
		t.Errorf("outputs = %v, want [20]", outputs)
	}
}

func TestMapLiteral(t *testing.T) {
	_, outputs := runScript(t, `let m = {"name": "bob"}
print(m["name"])`)
	if len(outputs) != 1 || outputs[0] != "bob" {
		t.Errorf("outputs = %v, want [bob]", outputs)
	}
}

func TestReturnNil(t *testing.T) {
	result, _ := runScript(t, `fn nothing() { return }
nothing()`)
	if _, ok := result.(NilObject); !ok {
		t.Errorf("expected NilObject, got %T", result)
	}
}

func TestNestedCall(t *testing.T) {
	_, outputs := runScript(t, `print(str_concat("a", "b", "c"))`)
	if len(outputs) != 1 || outputs[0] != "abc" {
		t.Errorf("outputs = %v, want [abc]", outputs)
	}
}

func TestConstDecl(t *testing.T) {
	_, outputs := runScript(t, `const PI = 314
print(str_from_int(PI))`)
	if outputs[0] != "314" {
		t.Errorf("outputs = %v, want [314]", outputs)
	}
}

func TestStringComparison(t *testing.T) {
	result, _ := runScript(t, `"abc" == "abc"`)
	if b, ok := result.(BoolObject); !ok || bool(b) != true {
		t.Errorf("expected true, got %v", result)
	}
}

func TestNilComparison(t *testing.T) {
	result, _ := runScript(t, `nil == nil`)
	if b, ok := result.(BoolObject); !ok || bool(b) != true {
		t.Errorf("expected true, got %v", result)
	}
}

func TestClosure(t *testing.T) {
	_, outputs := runScript(t, `let make_adder = fn(x) {
    return fn(y) { return x + y }
}
let add5 = make_adder(5)
print(str_from_int(add5(3)))
print(str_from_int(add5(10)))`)
	if outputs[0] != "8" || outputs[1] != "15" {
		t.Errorf("outputs = %v, want [8 15]", outputs)
	}
}

func TestClosureMultipleCaptures(t *testing.T) {
	_, outputs := runScript(t, `let x = 10
let y = 20
let getter = fn() { return x + y }
print(str_from_int(getter()))`)
	if outputs[0] != "30" {
		t.Errorf("outputs = %v, want [30]", outputs)
	}
}

func TestClosureCounter(t *testing.T) {
	_, outputs := runScript(t, `let make_counter = fn(start) {
    return fn() {
        return start
    }
}
let c = make_counter(42)
print(str_from_int(c()))`)
	if outputs[0] != "42" {
		t.Errorf("outputs = %v, want [42]", outputs)
	}
}
