package handler

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/topxeq/xxaitk/internal/output"
)

type HashHandler struct{}

type HashPayload struct {
	Data   string `json:"data"`
	File   string `json:"file"`
	Algo   string `json:"algo"`
	Encode string `json:"encode"`
}

type HashResult struct {
	Algorithm string `json:"algorithm"`
	Input     string `json:"input,omitempty"`
	Hash      string `json:"hash"`
}

func (h *HashHandler) Handle(data string, source string) *output.Response {
	start := time.Now()

	payload := h.parsePayload(data)

	var input []byte
	if payload.File != "" {
		fileData, err := os.ReadFile(payload.File)
		if err != nil {
			return output.Fail("hash", source, "HASH_FILE_ERROR",
				fmt.Sprintf("failed to read file: %s", err.Error()), "", start)
		}
		input = fileData
	} else if payload.Data != "" {
		input = []byte(payload.Data)
	} else {
		input = []byte(data)
	}

	algo := payload.Algo
	if algo == "" {
		algo = "sha256"
	}

	var hashStr string
	switch strings.ToLower(algo) {
	case "md5":
		h := md5.Sum(input)
		hashStr = hex.EncodeToString(h[:])
	case "sha1":
		h := sha1.Sum(input)
		hashStr = hex.EncodeToString(h[:])
	case "sha256":
		h := sha256.Sum256(input)
		hashStr = hex.EncodeToString(h[:])
	case "sha512":
		h := sha512.Sum512(input)
		hashStr = hex.EncodeToString(h[:])
	default:
		return output.Fail("hash", source, "HASH_UNKNOWN_ALGO",
			fmt.Sprintf("unknown algorithm: %s (use md5, sha1, sha256, sha512)", algo), "", start)
	}

	inputDesc := "inline"
	if payload.File != "" {
		inputDesc = payload.File
	}

	return output.Success("hash", source, &HashResult{
		Algorithm: strings.ToLower(algo),
		Input:     inputDesc,
		Hash:      hashStr,
	}, start)
}

func (h *HashHandler) parsePayload(data string) *HashPayload {
	payload := &HashPayload{}
	trimmed := strings.TrimSpace(data)
	if len(trimmed) == 0 {
		return payload
	}
	if trimmed[0] == '{' {
		json.Unmarshal([]byte(trimmed), payload)
	} else {
		payload.Data = data
	}
	return payload
}
