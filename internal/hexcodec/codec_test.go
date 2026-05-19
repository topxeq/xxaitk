package hexcodec

import "testing"

func TestEncodeString(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"hello", "68656c6c6f"},
		{"ls -la", "6c73202d6c61"},
		{"/etc/hosts", "2f6574632f686f737473"},
		{"", ""},
		{"ABC", "414243"},
		{"$`\\\"'", "24605c2227"},
	}
	for _, tt := range tests {
		got := EncodeString(tt.input)
		if got != tt.expect {
			t.Errorf("EncodeString(%q) = %q, want %q", tt.input, got, tt.expect)
		}
	}
}

func TestDecodeString(t *testing.T) {
	tests := []struct {
		input  string
		expect string
		hasErr bool
	}{
		{"68656c6c6f", "hello", false},
		{"6c73202d6c61", "ls -la", false},
		{"", "", false},
		{"414243", "ABC", false},
		{"zz", "", true},
		{"abc", "", true},
	}
	for _, tt := range tests {
		got, err := DecodeString(tt.input)
		if tt.hasErr && err == nil {
			t.Errorf("DecodeString(%q) expected error, got nil", tt.input)
		}
		if !tt.hasErr && err != nil {
			t.Errorf("DecodeString(%q) unexpected error: %v", tt.input, err)
		}
		if !tt.hasErr && got != tt.expect {
			t.Errorf("DecodeString(%q) = %q, want %q", tt.input, got, tt.expect)
		}
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	inputs := []string{
		"hello world",
		"ls -la /tmp",
		"print(\"hello\")",
		"$PATH | grep foo",
		"中文测试",
		"\n\t\r\000",
	}
	for _, s := range inputs {
		encoded := EncodeString(s)
		decoded, err := DecodeString(encoded)
		if err != nil {
			t.Errorf("round trip failed for %q: decode error: %v", s, err)
		}
		if decoded != s {
			t.Errorf("round trip failed: %q -> %q -> %q", s, encoded, decoded)
		}
	}
}

func TestIsValidHex(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"68656c6c6f", true},
		{"ABCDEF", true},
		{"0123456789abcdef", true},
		{"", false},
		{"g", false},
		{"abc", false},
		{"12345", false},
		{"zzzz", false},
	}
	for _, tt := range tests {
		got := IsValidHex(tt.input)
		if got != tt.want {
			t.Errorf("IsValidHex(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsKnownSource(t *testing.T) {
	if !IsKnownSource("FILE") {
		t.Error("FILE should be known source")
	}
	if !IsKnownSource("URL") {
		t.Error("URL should be known source")
	}
	if IsKnownSource("INLINE") {
		t.Error("INLINE should not be known source")
	}
}
