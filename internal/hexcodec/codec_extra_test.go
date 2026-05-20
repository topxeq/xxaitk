package hexcodec

import "testing"

func TestEncodeBytes(t *testing.T) {
	data := []byte{0x00, 0x01, 0xff}
	got := Encode(data)
	if got != "0001ff" {
		t.Errorf("Encode(%v) = %q, want 0001ff", data, got)
	}
}

func TestDecode(t *testing.T) {
	got, err := Decode("0001ff")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 || got[0] != 0x00 || got[1] != 0x01 || got[2] != 0xff {
		t.Errorf("Decode = %v, want [0 1 255]", got)
	}
}

func TestDecodeInvalid(t *testing.T) {
	_, err := Decode("zz")
	if err == nil {
		t.Error("expected error for invalid hex")
	}
}

func TestEncodeDecodeBytesRoundTrip(t *testing.T) {
	inputs := [][]byte{
		{},
		{0x00},
		{0xff, 0xfe, 0xfd},
		make([]byte, 256),
	}
	for _, inp := range inputs {
		encoded := Encode(inp)
		decoded, err := Decode(encoded)
		if err != nil {
			t.Errorf("round trip error: %v", err)
			continue
		}
		if len(decoded) != len(inp) {
			t.Errorf("length mismatch: %d vs %d", len(decoded), len(inp))
		}
		for i, b := range decoded {
			if b != inp[i] {
				t.Errorf("byte mismatch at %d: %d vs %d", i, b, inp[i])
			}
		}
	}
}

func TestIsValidHexEdgeCases(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"", false},
		{"a", false},
		{"AB", true},
		{"00", true},
		{"FF", true},
		{"0123456789abcdef", true},
		{"0123456789ABCDEF", true},
		{"g0", false},
		{"0g", false},
		{"  ", false},
		{"!!", false},
	}
	for _, tt := range tests {
		got := IsValidHex(tt.input)
		if got != tt.want {
			t.Errorf("IsValidHex(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsKnownSourceAll(t *testing.T) {
	if !IsKnownSource("FILE") {
		t.Error("FILE should be known")
	}
	if !IsKnownSource("file") {
		t.Error("file (lowercase) should be known via ToUpper")
	}
	if !IsKnownSource("URL") {
		t.Error("URL should be known")
	}
	if !IsKnownSource("url") {
		t.Error("url (lowercase) should be known")
	}
	if IsKnownSource("INLINE") {
		t.Error("INLINE should not be known")
	}
	if IsKnownSource("FTP") {
		t.Error("FTP should not be known")
	}
}

func TestEncodeStringUnicode(t *testing.T) {
	input := "中文"
	got := EncodeString(input)
	decoded, err := DecodeString(got)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if decoded != input {
		t.Errorf("round trip failed: %q -> %q -> %q", input, got, decoded)
	}
}

func TestDecodeStringEmpty(t *testing.T) {
	got, err := DecodeString("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("DecodeString('') = %q, want empty", got)
	}
}
