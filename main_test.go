package main

import (
	"os/exec"
	"testing"

	"github.com/topxeq/xxaitk/internal/handler"
)

func buildAitk(t *testing.T) string {
	t.Helper()
	bin := t.TempDir() + "/aitk"
	cmd := exec.Command("go", "build", "-o", bin, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %s: %v", string(out), err)
	}
	return bin
}

func TestVersion(t *testing.T) {
	bin := buildAitk(t)
	out, err := exec.Command(bin, "--version").CombinedOutput()
	if err != nil {
		t.Fatalf("--version failed: %v", err)
	}
	s := string(out)
	if !containsStr(s, "v0.4.0") {
		t.Errorf("version output = %q, want v0.4.0", s)
	}
}

func TestVersionShort(t *testing.T) {
	bin := buildAitk(t)
	out, err := exec.Command(bin, "-v").CombinedOutput()
	if err != nil {
		t.Fatalf("-v failed: %v", err)
	}
	s := string(out)
	if !containsStr(s, "v0.4.0") {
		t.Errorf("version output = %q, want v0.4.0", s)
	}
}

func TestHelp(t *testing.T) {
	bin := buildAitk(t)
	out, err := exec.Command(bin, "--help").CombinedOutput()
	if err != nil {
		t.Fatalf("--help failed: %v", err)
	}
	s := string(out)
	if !containsStr(s, "SHELL") || !containsStr(s, "PREFIX") {
		t.Errorf("help output missing expected content: %q", s)
	}
}

func TestHelpShort(t *testing.T) {
	bin := buildAitk(t)
	out, err := exec.Command(bin, "-h").CombinedOutput()
	if err != nil {
		t.Fatalf("-h failed: %v", err)
	}
	if len(out) == 0 {
		t.Error("expected help output")
	}
}

func TestShellCommand(t *testing.T) {
	bin := buildAitk(t)
	out, err := exec.Command(bin, "SHELL_6563686f2068656c6c6f").CombinedOutput()
	if err != nil {
		t.Fatalf("shell command failed: %v", err)
	}
	s := string(out)
	if !containsStr(s, "hello") {
		t.Errorf("shell output = %q, want hello", s)
	}
	if !containsStr(s, `"ok":true`) {
		t.Errorf("shell output missing ok:true: %q", s)
	}
}

func TestDecodeCommand(t *testing.T) {
	bin := buildAitk(t)
	out, err := exec.Command(bin, "DECODE_68656c6c6f20776f726c64").CombinedOutput()
	if err != nil {
		t.Fatalf("decode command failed: %v", err)
	}
	s := string(out)
	if !containsStr(s, "hello world") {
		t.Errorf("decode output = %q, want hello world", s)
	}
}

func TestInvalidHex(t *testing.T) {
	bin := buildAitk(t)
	out, _ := exec.Command(bin, "SHELL_zzzz").CombinedOutput()
	s := string(out)
	if !containsStr(s, `"ok":false`) {
		t.Errorf("invalid hex should return ok:false, got: %q", s)
	}
}

func TestUnknownPrefix(t *testing.T) {
	bin := buildAitk(t)
	out, _ := exec.Command(bin, "FOOBAR_68656c6c6f").CombinedOutput()
	s := string(out)
	if !containsStr(s, "UNKNOWN_PREFIX") {
		t.Errorf("unknown prefix should return error, got: %q", s)
	}
}

func TestLibList(t *testing.T) {
	bin := buildAitk(t)
	out, err := exec.Command(bin, "lib", "list").CombinedOutput()
	if err != nil {
		t.Fatalf("lib list failed: %v", err)
	}
	_ = string(out)
}

func TestLibNoArgs(t *testing.T) {
	bin := buildAitk(t)
	out, err := exec.Command(bin, "lib").CombinedOutput()
	if err != nil {
		t.Fatalf("lib no args failed: %v", err)
	}
	s := string(out)
	if !containsStr(s, "Usage") {
		t.Errorf("lib no args should show usage, got: %q", s)
	}
}

func TestScriptCommand(t *testing.T) {
	bin := buildAitk(t)
	out, err := exec.Command(bin, "SCRIPT_7072696e74282268692229").CombinedOutput()
	if err != nil {
		t.Fatalf("script command failed: %v: %s", err, string(out))
	}
	s := string(out)
	if !containsStr(s, "hi") {
		t.Errorf("script output = %q, want hi", s)
	}
}

func TestDebugFlag(t *testing.T) {
	bin := buildAitk(t)
	out, err := exec.Command(bin, "--debug", "DECODE_68656c6c6f").CombinedOutput()
	if err != nil {
		t.Fatalf("debug decode failed: %v", err)
	}
	_ = string(out)
}

func TestRegisterHandlers(t *testing.T) {
	registerHandlers()
	prefixes := []string{
		"SHELL", "SCRIPT", "EVAL", "HTTPGET", "HTTPPOST", "HTTPPUT",
		"HTTPPATCH", "HTTPDELETE", "FILE", "READFILE", "WRITEFILE",
		"LISTDIR", "DELETE", "INFO", "DECODE", "ENCODE", "B64ENC",
		"B64DEC", "URLENC", "URLDEC", "PING", "HASH", "PROCESS",
		"DIFF", "ARCHIVE", "SQL", "GIT", "PORT", "NETDOWNLOAD", "CAPABILITIES",
	}
	for _, p := range prefixes {
		if _, ok := handler.Get(p); !ok {
			t.Errorf("handler %s not registered", p)
		}
	}
}

func TestRegisterHandlersAllTypes(t *testing.T) {
	registerHandlers()
	h, ok := handler.Get("HTTPPUT")
	if !ok {
		t.Fatal("HTTPPUT not registered")
	}
	_ = h
	h, ok = handler.Get("HTTPPATCH")
	if !ok {
		t.Fatal("HTTPPATCH not registered")
	}
	_ = h
}

func TestEncodeCommand(t *testing.T) {
	bin := buildAitk(t)
	out, err := exec.Command(bin, "ENCODE_68656c6c6f").CombinedOutput()
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	s := string(out)
	if !containsStr(s, "hex") {
		t.Errorf("encode output missing 'hex': %q", s)
	}
}

func TestInfoCommand(t *testing.T) {
	bin := buildAitk(t)
	out, err := exec.Command(bin, "INFO_6f73").CombinedOutput()
	if err != nil {
		t.Fatalf("info failed: %v", err)
	}
	s := string(out)
	if !containsStr(s, "linux") && !containsStr(s, "darwin") {
		t.Errorf("info output missing OS: %q", s)
	}
}

func TestLibRemoveNoArg(t *testing.T) {
	bin := buildAitk(t)
	out, _ := exec.Command(bin, "lib", "remove").CombinedOutput()
	s := string(out)
	if !containsStr(s, "Usage") && !containsStr(s, "remove") {
		t.Errorf("lib remove no arg: %q", s)
	}
}

func TestLibGetNoArg(t *testing.T) {
	bin := buildAitk(t)
	out, _ := exec.Command(bin, "lib", "get").CombinedOutput()
	s := string(out)
	if !containsStr(s, "Usage") && !containsStr(s, "get") {
		t.Errorf("lib get no arg: %q", s)
	}
}

func TestLibUnknownSubcommand(t *testing.T) {
	bin := buildAitk(t)
	out, err := exec.Command(bin, "lib", "badsubcmd").CombinedOutput()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)
	if !containsStr(s, "Usage") {
		t.Errorf("unknown lib subcommand should show usage: %q", s)
	}
}

func TestHandleLibCommandEmpty(t *testing.T) {
	handleLibCommand(nil)
}

func TestHandleLibCommandList(t *testing.T) {
	handleLibCommand([]string{"list"})
}

func TestHandleLibCommandLs(t *testing.T) {
	handleLibCommand([]string{"ls"})
}

func TestPrintHelp(t *testing.T) {
	printHelp()
}

func TestPrintLibHelp(t *testing.T) {
	printLibHelp()
}

func TestDebugWithShell(t *testing.T) {
	bin := buildAitk(t)
	out, err := exec.Command(bin, "--debug", "SHELL_6563686f2074657374").CombinedOutput()
	if err != nil {
		t.Fatalf("debug shell failed: %v", err)
	}
	_ = string(out)
}

func TestB64EncCommand(t *testing.T) {
	bin := buildAitk(t)
	out, err := exec.Command(bin, "B64ENC_68656c6c6f").CombinedOutput()
	if err != nil {
		t.Fatalf("b64enc failed: %v", err)
	}
	s := string(out)
	if !containsStr(s, "aGVsbG8=") {
		t.Errorf("b64enc output: %q", s)
	}
}

func TestHashCommand(t *testing.T) {
	bin := buildAitk(t)
	out, err := exec.Command(bin, "HASH_68656c6c6f").CombinedOutput()
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}
	s := string(out)
	if !containsStr(s, "sha256") {
		t.Errorf("hash output missing sha256: %q", s)
	}
}

func TestEvalCommand(t *testing.T) {
	bin := buildAitk(t)
	out, err := exec.Command(bin, "EVAL_31202b2032").CombinedOutput()
	if err != nil {
		t.Fatalf("eval failed: %v", err)
	}
	s := string(out)
	if !containsStr(s, "3") {
		t.Errorf("eval output: %q", s)
	}
}

func TestCapabilitiesCommand(t *testing.T) {
	bin := buildAitk(t)
	out, err := exec.Command(bin, "CAPABILITIES_616c6c").CombinedOutput()
	if err != nil {
		t.Fatalf("capabilities failed: %v", err)
	}
	s := string(out)
	if !containsStr(s, "prefixes") {
		t.Errorf("capabilities output missing 'prefixes': %q", s)
	}
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
