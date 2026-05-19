package hexcodec

import (
	"encoding/hex"
	"strings"
)

func Encode(data []byte) string {
	return hex.EncodeToString(data)
}

func EncodeString(s string) string {
	return hex.EncodeToString([]byte(s))
}

func Decode(hexStr string) ([]byte, error) {
	return hex.DecodeString(hexStr)
}

func DecodeString(hexStr string) (string, error) {
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func IsValidHex(s string) bool {
	if len(s) == 0 {
		return false
	}
	if len(s)%2 != 0 {
		return false
	}
	for _, c := range s {
		if !isHexChar(c) {
			return false
		}
	}
	return true
}

func isHexChar(c rune) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func IsKnownSource(s string) bool {
	s = strings.ToUpper(s)
	return s == "FILE" || s == "URL"
}
