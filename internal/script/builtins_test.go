package script

import (
	"os"
	"strings"
	"testing"
)

func runBuiltinTest(t *testing.T, code string) (Object, []string) {
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

func TestIOReadFile(t *testing.T) {
	f, err := os.CreateTemp("", "read_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("hello world")
	f.Close()

	result, _ := runBuiltinTest(t, `io_read_file("`+f.Name()+`")`)
	s, ok := result.(StringObject)
	if !ok {
		t.Fatalf("expected StringObject, got %T", result)
	}
	if !strings.Contains(string(s), "hello world") {
		t.Errorf("expected content 'hello world', got %q", string(s))
	}
}

func TestIOWriteFile(t *testing.T) {
	path := os.TempDir() + "/write_test_" + strings.ReplaceAll(strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(""), "\n")), " ", "_")
	defer os.Remove(path)

	result, _ := runBuiltinTest(t, `io_write_file("`+path+`", "written")`)
	b, ok := result.(BoolObject)
	if !ok || !bool(b) {
		t.Fatalf("expected true, got %v", result)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "written" {
		t.Errorf("expected 'written', got %q", string(data))
	}
}

func TestIOAppendFile(t *testing.T) {
	f, err := os.CreateTemp("", "append_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("first")
	f.Close()

	result, _ := runBuiltinTest(t, `io_append_file("`+f.Name()+`", "second")`)
	b, ok := result.(BoolObject)
	if !ok || !bool(b) {
		t.Fatalf("expected true, got %v", result)
	}
	data, _ := os.ReadFile(f.Name())
	if string(data) != "firstsecond" {
		t.Errorf("expected 'firstsecond', got %q", string(data))
	}
}

func TestIOExists(t *testing.T) {
	f, err := os.CreateTemp("", "exists_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	result, _ := runBuiltinTest(t, `io_exists("`+f.Name()+`")`)
	b, ok := result.(BoolObject)
	if !ok || !bool(b) {
		t.Errorf("expected true for existing file, got %v", result)
	}

	result, _ = runBuiltinTest(t, `io_exists("/nonexistent_path_12345")`)
	b, ok = result.(BoolObject)
	if ok && bool(b) {
		t.Errorf("expected false for nonexistent file, got true")
	}
}

func TestIOIsDir(t *testing.T) {
	dir := os.TempDir()
	result, _ := runBuiltinTest(t, `io_is_dir("`+dir+`")`)
	b, ok := result.(BoolObject)
	if !ok || !bool(b) {
		t.Errorf("expected true for temp dir, got %v", result)
	}
}

func TestIOIsFile(t *testing.T) {
	f, err := os.CreateTemp("", "isfile_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	result, _ := runBuiltinTest(t, `io_is_file("`+f.Name()+`")`)
	b, ok := result.(BoolObject)
	if !ok || !bool(b) {
		t.Errorf("expected true for file, got %v", result)
	}
}

func TestIOSize(t *testing.T) {
	f, err := os.CreateTemp("", "size_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("12345")
	f.Close()

	result, _ := runBuiltinTest(t, `io_size("`+f.Name()+`")`)
	i, ok := result.(IntObject)
	if !ok {
		t.Fatalf("expected IntObject, got %T", result)
	}
	if int64(i) != 5 {
		t.Errorf("expected size 5, got %d", int64(i))
	}
}

func TestIOMkdir(t *testing.T) {
	dir := os.TempDir() + "/mkdir_test_xyz"
	defer os.RemoveAll(dir)

	result, _ := runBuiltinTest(t, `io_mkdir("`+dir+`")`)
	b, ok := result.(BoolObject)
	if !ok || !bool(b) {
		t.Fatalf("expected true, got %v", result)
	}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		t.Errorf("expected directory to exist")
	}
}

func TestIOCopy(t *testing.T) {
	f, err := os.CreateTemp("", "copy_src_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("copy me")
	f.Close()

	dst := f.Name() + ".copy"
	defer os.Remove(dst)

	result, _ := runBuiltinTest(t, `io_copy("`+f.Name()+`", "`+dst+`")`)
	b, ok := result.(BoolObject)
	if !ok || !bool(b) {
		t.Fatalf("expected true, got %v", result)
	}
	data, _ := os.ReadFile(dst)
	if string(data) != "copy me" {
		t.Errorf("expected 'copy me', got %q", string(data))
	}
}

func TestIOMove(t *testing.T) {
	f, err := os.CreateTemp("", "move_src_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	src := f.Name()
	f.WriteString("move me")
	f.Close()
	defer os.Remove(src)

	dst := src + ".moved"
	defer os.Remove(dst)

	result, _ := runBuiltinTest(t, `io_move("`+src+`", "`+dst+`")`)
	b, ok := result.(BoolObject)
	if !ok || !bool(b) {
		t.Fatalf("expected true, got %v", result)
	}
	data, _ := os.ReadFile(dst)
	if string(data) != "move me" {
		t.Errorf("expected 'move me', got %q", string(data))
	}
}

func TestIORemove(t *testing.T) {
	f, err := os.CreateTemp("", "remove_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.Close()

	result, _ := runBuiltinTest(t, `io_remove("`+path+`")`)
	b, ok := result.(BoolObject)
	if !ok || !bool(b) {
		t.Fatalf("expected true, got %v", result)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected file to be removed")
	}
}

func TestIOTempDir(t *testing.T) {
	result, _ := runBuiltinTest(t, `io_temp_dir()`)
	s, ok := result.(StringObject)
	if !ok {
		t.Fatalf("expected StringObject, got %T", result)
	}
	if string(s) == "" {
		t.Error("expected non-empty temp dir")
	}
}

func TestIOAbsPath(t *testing.T) {
	result, _ := runBuiltinTest(t, `io_abs_path(".")`)
	s, ok := result.(StringObject)
	if !ok {
		t.Fatalf("expected StringObject, got %T", result)
	}
	if string(s) == "" {
		t.Error("expected non-empty absolute path")
	}
}

func TestIOReadFileNotFound(t *testing.T) {
	result, _ := runBuiltinTest(t, `io_read_file("/nonexistent_file_xyz_12345")`)
	_, ok := result.(ErrorObject)
	if !ok {
		t.Errorf("expected ErrorObject, got %T", result)
	}
}

func TestOSExec(t *testing.T) {
	result, _ := runBuiltinTest(t, `os_exec("echo hello")`)
	s, ok := result.(StringObject)
	if !ok {
		t.Fatalf("expected StringObject, got %T", result)
	}
	if !strings.Contains(string(s), "hello") {
		t.Errorf("expected output containing 'hello', got %q", string(s))
	}
}

func TestOSEnv(t *testing.T) {
	result, _ := runBuiltinTest(t, `os_env()`)
	_, ok := result.(MapObject)
	if !ok {
		t.Fatalf("expected MapObject, got %T", result)
	}
}

func TestOSGetenv(t *testing.T) {
	os.Setenv("XXAITK_TEST_KEY", "testval")
	defer os.Unsetenv("XXAITK_TEST_KEY")

	result, _ := runBuiltinTest(t, `os_getenv("XXAITK_TEST_KEY")`)
	s, ok := result.(StringObject)
	if !ok {
		t.Fatalf("expected StringObject, got %T", result)
	}
	if string(s) != "testval" {
		t.Errorf("expected 'testval', got %q", string(s))
	}
}

func TestOSCwd(t *testing.T) {
	result, _ := runBuiltinTest(t, `os_cwd()`)
	s, ok := result.(StringObject)
	if !ok {
		t.Fatalf("expected StringObject, got %T", result)
	}
	if string(s) == "" {
		t.Error("expected non-empty cwd")
	}
}

func TestOSHostname(t *testing.T) {
	result, _ := runBuiltinTest(t, `os_hostname()`)
	s, ok := result.(StringObject)
	if !ok {
		t.Fatalf("expected StringObject, got %T", result)
	}
	if string(s) == "" {
		t.Error("expected non-empty hostname")
	}
}

func TestTryBuiltinWithBuiltinFn(t *testing.T) {
	result, _ := runBuiltinTest(t, `try(io_read_file, "/nonexistent_xyz_12345")`)
	list, ok := result.(ListObject)
	if !ok {
		t.Fatalf("expected ListObject, got %T", result)
	}
	if len(list.Elements) < 2 {
		t.Fatalf("expected at least 2 elements, got %d", len(list.Elements))
	}
	b, ok := list.Elements[0].(BoolObject)
	if !ok || bool(b) {
		t.Errorf("expected first element false (error), got %v", list.Elements[0])
	}
	_, ok = list.Elements[1].(ErrorObject)
	if !ok {
		t.Errorf("expected second element ErrorObject, got %T", list.Elements[1])
	}
}

func TestCatchBuiltinWithNoError(t *testing.T) {
	result, _ := runBuiltinTest(t, `catch("not an error")`)
	s, ok := result.(StringObject)
	if !ok {
		t.Fatalf("expected StringObject, got %T", result)
	}
	if string(s) != "" {
		t.Errorf("expected empty string for non-error, got %q", string(s))
	}
}

func TestIsErrorOnRegularValue(t *testing.T) {
	result, _ := runBuiltinTest(t, `is_error("hello")`)
	b, ok := result.(BoolObject)
	if !ok {
		t.Fatalf("expected BoolObject, got %T", result)
	}
	if bool(b) {
		t.Error("expected false for non-error value")
	}
}

func TestNetHTTPGet(t *testing.T) {
	result, _ := runBuiltinTest(t, `net_http_get("https://httpbin.org/get")`)
	if _, ok := result.(NilObject); ok {
		t.Skip("network unavailable, skipping")
	}
	s, ok := result.(StringObject)
	if !ok {
		t.Fatalf("expected StringObject, got %T", result)
	}
	if string(s) == "" {
		t.Error("expected non-empty response")
	}
}

func TestNetDNSLookup(t *testing.T) {
	result, _ := runBuiltinTest(t, `net_dns_lookup("localhost")`)
	list, ok := result.(ListObject)
	if !ok {
		if _, ok2 := result.(ListObject); !ok2 {
			t.Skip("network unavailable, skipping")
		}
		t.Fatalf("expected ListObject, got %T", result)
	}
	if len(list.Elements) == 0 {
		t.Error("expected at least one address for localhost")
	}
}

func TestLogInfo(t *testing.T) {
	result, _ := runBuiltinTest(t, `log_info("test info")`)
	_, ok := result.(NilObject)
	if !ok {
		t.Errorf("expected NilObject, got %T", result)
	}
}

func TestLogWarn(t *testing.T) {
	result, _ := runBuiltinTest(t, `log_warn("test warn")`)
	_, ok := result.(NilObject)
	if !ok {
		t.Errorf("expected NilObject, got %T", result)
	}
}

func TestLogError(t *testing.T) {
	result, _ := runBuiltinTest(t, `log_error("test error")`)
	_, ok := result.(NilObject)
	if !ok {
		t.Errorf("expected NilObject, got %T", result)
	}
}

func TestLogDebug(t *testing.T) {
	result, _ := runBuiltinTest(t, `log_debug("test debug")`)
	_, ok := result.(NilObject)
	if !ok {
		t.Errorf("expected NilObject, got %T", result)
	}
}

func TestTimeNow(t *testing.T) {
	result, _ := runBuiltinTest(t, `time_now()`)
	s, ok := result.(StringObject)
	if !ok {
		t.Fatalf("expected StringObject, got %T", result)
	}
	if string(s) == "" {
		t.Error("expected non-empty time string")
	}
}

func TestTimeFormat(t *testing.T) {
	result, _ := runBuiltinTest(t, `time_format("2024-01-15T10:30:00Z", "2006-01-02")`)
	s, ok := result.(StringObject)
	if !ok {
		t.Fatalf("expected StringObject, got %T", result)
	}
	if !strings.Contains(string(s), "2024") {
		t.Errorf("expected formatted date containing '2024', got %q", string(s))
	}
}

func TestTimeSleep(t *testing.T) {
	result, _ := runBuiltinTest(t, `time_sleep(1)`)
	_, ok := result.(NilObject)
	if !ok {
		t.Errorf("expected NilObject, got %T", result)
	}
}

func TestTimeDuration(t *testing.T) {
	result, _ := runBuiltinTest(t, `time_duration("2024-01-01T00:00:00Z", "2024-01-01T00:00:01Z")`)
	i, ok := result.(IntObject)
	if !ok {
		t.Fatalf("expected IntObject, got %T", result)
	}
	if int64(i) != 1000 {
		t.Errorf("expected 1000 ms, got %d", int64(i))
	}
}

func TestTimeParse(t *testing.T) {
	result, _ := runBuiltinTest(t, `time_parse("2006-01-02", "2024-06-15")`)
	s, ok := result.(StringObject)
	if !ok {
		t.Fatalf("expected StringObject, got %T", result)
	}
	if !strings.Contains(string(s), "2024") {
		t.Errorf("expected parsed time containing '2024', got %q", string(s))
	}
}

func TestJSONSet(t *testing.T) {
	result, _ := runBuiltinTest(t, `json_set({"a": 1}, "b", 2)`)
	m, ok := result.(MapObject)
	if !ok {
		t.Fatalf("expected MapObject, got %T", result)
	}
	if _, exists := m.Pairs["b"]; !exists {
		t.Error("expected key 'b' to exist")
	}
}

func TestJSONHas(t *testing.T) {
	result, _ := runBuiltinTest(t, `json_has({"a": 1}, "a")`)
	b, ok := result.(BoolObject)
	if !ok || !bool(b) {
		t.Errorf("expected true, got %v", result)
	}

	result, _ = runBuiltinTest(t, `json_has({"a": 1}, "b")`)
	b, ok = result.(BoolObject)
	if ok && bool(b) {
		t.Errorf("expected false for missing key")
	}
}

func TestJSONDecodeList(t *testing.T) {
	result, _ := runBuiltinTest(t, `json_decode("[1, 2, 3]")`)
	list, ok := result.(ListObject)
	if !ok {
		t.Fatalf("expected ListObject, got %T", result)
	}
	if len(list.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(list.Elements))
	}
}

func TestJSONDecodeInvalid(t *testing.T) {
	result, _ := runBuiltinTest(t, `json_decode("not json")`)
	_, ok := result.(NilObject)
	if !ok {
		t.Errorf("expected NilObject for invalid JSON, got %T", result)
	}
}

func TestMathRand(t *testing.T) {
	result, _ := runBuiltinTest(t, `math_rand()`)
	f, ok := result.(FloatObject)
	if !ok {
		t.Fatalf("expected FloatObject, got %T", result)
	}
	if float64(f) < 0 || float64(f) >= 1 {
		t.Errorf("expected value in [0,1), got %v", float64(f))
	}
}

func TestMathRandInt(t *testing.T) {
	result, _ := runBuiltinTest(t, `math_rand_int(10)`)
	i, ok := result.(IntObject)
	if !ok {
		t.Fatalf("expected IntObject, got %T", result)
	}
	if int64(i) < 0 || int64(i) >= 10 {
		t.Errorf("expected value in [0,10), got %d", int64(i))
	}
}

func TestMathLog(t *testing.T) {
	result, _ := runBuiltinTest(t, `math_log(2.718281828)`)
	f, ok := result.(FloatObject)
	if !ok {
		t.Fatalf("expected FloatObject, got %T", result)
	}
	if float64(f) < 0.9 || float64(f) > 1.1 {
		t.Errorf("expected log(e) ~= 1.0, got %v", float64(f))
	}
}

func TestMathExp(t *testing.T) {
	result, _ := runBuiltinTest(t, `math_exp(1)`)
	f, ok := result.(FloatObject)
	if !ok {
		t.Fatalf("expected FloatObject, got %T", result)
	}
	if float64(f) < 2.7 || float64(f) > 2.8 {
		t.Errorf("expected exp(1) ~= 2.718, got %v", float64(f))
	}
}

func TestMathSin(t *testing.T) {
	result, _ := runBuiltinTest(t, `math_sin(0)`)
	f, ok := result.(FloatObject)
	if !ok {
		t.Fatalf("expected FloatObject, got %T", result)
	}
	if float64(f) != 0 {
		t.Errorf("expected sin(0) = 0, got %v", float64(f))
	}
}

func TestMathCos(t *testing.T) {
	result, _ := runBuiltinTest(t, `math_cos(0)`)
	f, ok := result.(FloatObject)
	if !ok {
		t.Fatalf("expected FloatObject, got %T", result)
	}
	if float64(f) != 1 {
		t.Errorf("expected cos(0) = 1, got %v", float64(f))
	}
}

func TestStrPadLeft(t *testing.T) {
	result, _ := runBuiltinTest(t, `str_pad_left("hi", 5, "0")`)
	s, ok := result.(StringObject)
	if !ok {
		t.Fatalf("expected StringObject, got %T", result)
	}
	if string(s) != "000hi" {
		t.Errorf("expected '000hi', got %q", string(s))
	}
}

func TestStrPadRight(t *testing.T) {
	result, _ := runBuiltinTest(t, `str_pad_right("hi", 5, "0")`)
	s, ok := result.(StringObject)
	if !ok {
		t.Fatalf("expected StringObject, got %T", result)
	}
	if string(s) != "hi000" {
		t.Errorf("expected 'hi000', got %q", string(s))
	}
}

func TestStrIndex(t *testing.T) {
	result, _ := runBuiltinTest(t, `str_index("hello", "ll")`)
	i, ok := result.(IntObject)
	if !ok {
		t.Fatalf("expected IntObject, got %T", result)
	}
	if int64(i) != 2 {
		t.Errorf("expected 2, got %d", int64(i))
	}
}

func TestStrFromFloat(t *testing.T) {
	result, _ := runBuiltinTest(t, `str_from_float(3.14)`)
	s, ok := result.(StringObject)
	if !ok {
		t.Fatalf("expected StringObject, got %T", result)
	}
	if !strings.Contains(string(s), "3.14") {
		t.Errorf("expected string containing '3.14', got %q", string(s))
	}
}

func TestStrToInt(t *testing.T) {
	result, _ := runBuiltinTest(t, `str_to_int("42")`)
	i, ok := result.(IntObject)
	if !ok {
		t.Fatalf("expected IntObject, got %T", result)
	}
	if int64(i) != 42 {
		t.Errorf("expected 42, got %d", int64(i))
	}
}

func TestStrToFloat(t *testing.T) {
	result, _ := runBuiltinTest(t, `str_to_float("3.14")`)
	f, ok := result.(FloatObject)
	if !ok {
		t.Fatalf("expected FloatObject, got %T", result)
	}
	if float64(f) < 3.13 || float64(f) > 3.15 {
		t.Errorf("expected ~3.14, got %v", float64(f))
	}
}

func TestStrRepeat(t *testing.T) {
	result, _ := runBuiltinTest(t, `str_repeat("ab", 3)`)
	s, ok := result.(StringObject)
	if !ok {
		t.Fatalf("expected StringObject, got %T", result)
	}
	if string(s) != "ababab" {
		t.Errorf("expected 'ababab', got %q", string(s))
	}
}

func TestListShift(t *testing.T) {
	result, _ := runBuiltinTest(t, `list_shift([10, 20, 30])`)
	i, ok := result.(IntObject)
	if !ok {
		t.Fatalf("expected IntObject, got %T", result)
	}
	if int64(i) != 10 {
		t.Errorf("expected 10, got %d", int64(i))
	}
}

func TestListSet(t *testing.T) {
	result, _ := runBuiltinTest(t, `list_set([1, 2, 3], 1, 99)`)
	list, ok := result.(ListObject)
	if !ok {
		t.Fatalf("expected ListObject, got %T", result)
	}
	elem, ok := list.Elements[1].(IntObject)
	if !ok || int64(elem) != 99 {
		t.Errorf("expected element at index 1 to be 99, got %v", list.Elements[1])
	}
}

func TestListIndex(t *testing.T) {
	result, _ := runBuiltinTest(t, `list_index([10, 20, 30], 20)`)
	i, ok := result.(IntObject)
	if !ok {
		t.Fatalf("expected IntObject, got %T", result)
	}
	if int64(i) != 1 {
		t.Errorf("expected 1, got %d", int64(i))
	}
}

func TestListJoin(t *testing.T) {
	result, _ := runBuiltinTest(t, `list_join(["a", "b", "c"], "-")`)
	s, ok := result.(StringObject)
	if !ok {
		t.Fatalf("expected StringObject, got %T", result)
	}
	if string(s) != "a-b-c" {
		t.Errorf("expected 'a-b-c', got %q", string(s))
	}
}

func TestListFlat(t *testing.T) {
	result, _ := runBuiltinTest(t, `list_flat([[1, 2], [3, 4]])`)
	list, ok := result.(ListObject)
	if !ok {
		t.Fatalf("expected ListObject, got %T", result)
	}
	if len(list.Elements) != 4 {
		t.Errorf("expected 4 elements after flat, got %d", len(list.Elements))
	}
}
