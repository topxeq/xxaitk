package argparser

import (
	"fmt"
	"strings"

	"github.com/topxeq/xxaitk/internal/hexcodec"
)

type ParsedArg struct {
	Operation string
	Source    string
	HexData   string
}

var knownSources = map[string]bool{
	"FILE": true,
	"URL":  true,
}

func Parse(arg string) (*ParsedArg, error) {
	if arg == "" {
		return nil, fmt.Errorf("empty argument")
	}

	parts := strings.SplitN(arg, "_", 3)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid argument format: missing underscore separator")
	}

	operation := strings.ToUpper(parts[0])
	if len(operation) == 0 {
		return nil, fmt.Errorf("empty operation prefix")
	}

	if len(parts) >= 3 && knownSources[strings.ToUpper(parts[1])] {
		source := strings.ToUpper(parts[1])
		hexData := parts[2]
		if hexcodec.IsValidHex(hexData) {
			return &ParsedArg{
				Operation: operation,
				Source:    source,
				HexData:   hexData,
			}, nil
		}
		if isPlaintextJSON(hexData) {
			trimmed := strings.TrimSpace(hexData)
			return &ParsedArg{
				Operation: operation,
				Source:    source,
				HexData:   hexcodec.EncodeString(trimmed),
			}, nil
		}
		return nil, fmt.Errorf("invalid hex data after %s_%s_", operation, source)
	}

	hexData := parts[1]
	if len(parts) == 3 {
		hexData = parts[1] + "_" + parts[2]
		if hexcodec.IsValidHex(hexData) {
			return &ParsedArg{
				Operation: operation,
				Source:    "INLINE",
				HexData:   hexData,
			}, nil
		}
	}

	if hexcodec.IsValidHex(hexData) {
		return &ParsedArg{
			Operation: operation,
			Source:    "INLINE",
			HexData:   hexData,
		}, nil
	}

	rawData := arg[len(parts[0])+1:]
	if isPlaintextJSON(rawData) {
		trimmed := strings.TrimSpace(rawData)
		return &ParsedArg{
			Operation: operation,
			Source:    "INLINE",
			HexData:   hexcodec.EncodeString(trimmed),
		}, nil
	}

	return nil, fmt.Errorf("invalid hex data for operation %s", operation)
}

func isPlaintextJSON(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[")
}

func (p *ParsedArg) DecodeData() (string, error) {
	return hexcodec.DecodeString(p.HexData)
}
