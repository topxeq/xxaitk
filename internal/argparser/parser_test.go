package argparser

import "testing"

func TestParseInline(t *testing.T) {
	p, err := Parse("SHELL_6c73202d6c61")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Operation != "SHELL" {
		t.Errorf("Operation = %q, want SHELL", p.Operation)
	}
	if p.Source != "INLINE" {
		t.Errorf("Source = %q, want INLINE", p.Source)
	}
	if p.HexData != "6c73202d6c61" {
		t.Errorf("HexData = %q, want 6c73202d6c61", p.HexData)
	}
}

func TestParseFileSource(t *testing.T) {
	p, err := Parse("SHELL_FILE_2f746d702f636d642e7368")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Operation != "SHELL" {
		t.Errorf("Operation = %q, want SHELL", p.Operation)
	}
	if p.Source != "FILE" {
		t.Errorf("Source = %q, want FILE", p.Source)
	}
	if p.HexData != "2f746d702f636d642e7368" {
		t.Errorf("HexData = %q, want 2f746d702f636d642e7368", p.HexData)
	}
}

func TestParseURLSource(t *testing.T) {
	p, err := Parse("SCRIPT_URL_68747470733a2f2f6578616d706c65")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Operation != "SCRIPT" {
		t.Errorf("Operation = %q, want SCRIPT", p.Operation)
	}
	if p.Source != "URL" {
		t.Errorf("Source = %q, want URL", p.Source)
	}
}

func TestParseInvalidHex(t *testing.T) {
	_, err := Parse("SHELL_zzzz")
	if err == nil {
		t.Error("expected error for invalid hex, got nil")
	}
}

func TestParseNoUnderscore(t *testing.T) {
	_, err := Parse("SHELL")
	if err == nil {
		t.Error("expected error for no underscore, got nil")
	}
}

func TestParseEmpty(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Error("expected error for empty arg, got nil")
	}
}

func TestParseDecodeData(t *testing.T) {
	p, err := Parse("FILE_2f6574632f686f737473")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	decoded, err := p.DecodeData()
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if decoded != "/etc/hosts" {
		t.Errorf("DecodeData() = %q, want /etc/hosts", decoded)
	}
}

func TestParseAllPrefixes(t *testing.T) {
	prefixes := []string{"SHELL", "SCRIPT", "EVAL", "HTTPGET", "HTTPPOST", "FILE", "READFILE", "WRITEFILE", "LISTDIR", "DELETE", "INFO", "DECODE", "ENCODE", "B64ENC", "B64DEC", "URLENC", "URLDEC", "PING"}
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
