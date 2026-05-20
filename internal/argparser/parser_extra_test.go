package argparser

import "testing"

func TestParseURLSourceHTTPGet(t *testing.T) {
	p, err := Parse("HTTPGET_URL_687474703a2f2f6578616d706c65")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Operation != "HTTPGET" {
		t.Errorf("Operation = %q, want HTTPGET", p.Operation)
	}
	if p.Source != "URL" {
		t.Errorf("Source = %q, want URL", p.Source)
	}
}

func TestParseLowercaseSource(t *testing.T) {
	p, err := Parse("SHELL_file_2f746d70")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Source != "FILE" {
		t.Errorf("Source = %q, want FILE (should uppercase)", p.Source)
	}
}

func TestParseLowercaseOperation(t *testing.T) {
	p, err := Parse("shell_68656c6c6f")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Operation != "SHELL" {
		t.Errorf("Operation = %q, want SHELL (should uppercase)", p.Operation)
	}
}

func TestParseLongHexData(t *testing.T) {
	longHex := ""
	for i := 0; i < 1000; i++ {
		longHex += "41"
	}
	p, err := Parse("DECODE_" + longHex)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Source != "INLINE" {
		t.Errorf("Source = %q, want INLINE", p.Source)
	}
	decoded, err := p.DecodeData()
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(decoded) != 1000 {
		t.Errorf("decoded length = %d, want 1000", len(decoded))
	}
}

func TestParseUnderscoreInHex(t *testing.T) {
	p, err := Parse("SHELL_5f5f5f")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	decoded, err := p.DecodeData()
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if decoded != "___" {
		t.Errorf("decoded = %q, want ___", decoded)
	}
}

func TestParseMultipleUnderscores(t *testing.T) {
	p, err := Parse("DECODE_FILE_68656c6c6f")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Operation != "DECODE" {
		t.Errorf("Operation = %q, want DECODE", p.Operation)
	}
	if p.Source != "FILE" {
		t.Errorf("Source = %q, want FILE", p.Source)
	}
	if p.HexData != "68656c6c6f" {
		t.Errorf("HexData = %q, want 68656c6c6f", p.HexData)
	}
}

func TestParseOnlyUnderscore(t *testing.T) {
	_, err := Parse("_")
	if err == nil {
		t.Error("expected error for single underscore")
	}
}

func TestParseDoubleUnderscore(t *testing.T) {
	_, err := Parse("__68656c6c6f")
	if err == nil {
		t.Error("expected error for empty operation")
	}
}

func TestParseEmptyOperationWithSource(t *testing.T) {
	_, err := Parse("_FILE_68656c6c6f")
	if err == nil {
		t.Error("expected error for empty operation")
	}
}

func TestDecodeDataInvalidHex(t *testing.T) {
	p := &ParsedArg{Operation: "SHELL", Source: "INLINE", HexData: "zzzz"}
	_, err := p.DecodeData()
	if err == nil {
		t.Error("expected error for invalid hex in DecodeData")
	}
}

func TestParseAllLongPrefixes(t *testing.T) {
	prefixes := []string{
		"HTTPGET", "HTTPPOST", "HTTPPUT", "HTTPPATCH", "HTTPDELETE",
		"WRITEFILE", "READFILE", "LISTDIR", "CAPABILITIES", "NETDOWNLOAD",
	}
	hexData := "68656c6c6f"
	for _, prefix := range prefixes {
		arg := prefix + "_" + hexData
		p, err := Parse(arg)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", arg, err)
			continue
		}
		if p.Operation != prefix {
			t.Errorf("Operation = %q, want %q", p.Operation, prefix)
		}
	}
}

func TestParseHTTPWithSource(t *testing.T) {
	p, err := Parse("HTTPGET_FILE_6874747073")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Operation != "HTTPGET" {
		t.Errorf("Operation = %q, want HTTPGET", p.Operation)
	}
	if p.Source != "FILE" {
		t.Errorf("Source = %q, want FILE", p.Source)
	}
}
