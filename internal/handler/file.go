package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/topxeq/xxaitk/internal/output"
)

type FileHandler struct{}

type FilePayload struct {
	Path     string `json:"path"`
	Encoding string `json:"encoding,omitempty"`
	Offset   int64  `json:"offset,omitempty"`
	Limit    int64  `json:"limit,omitempty"`
}

type FileResult struct {
	Path       string `json:"path"`
	Content    string `json:"content,omitempty"`
	ContentB64 string `json:"content_b64,omitempty"`
	Size       int64  `json:"size"`
	Encoding   string `json:"encoding"`
}

func (h *FileHandler) Handle(data string, source string) *output.Response {
	start := time.Now()

	payload := h.parsePayload(data)
	if payload.Path == "" {
		return output.Fail("file", source, "FILE_EMPTY_PATH", "empty file path", "", start)
	}

	info, err := os.Stat(payload.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return output.Fail("file", source, "FILE_NOT_FOUND",
				fmt.Sprintf("file not found: %s", payload.Path), "", start)
		}
		return output.Fail("file", source, "FILE_STAT_ERROR", err.Error(), "", start)
	}

	if info.IsDir() {
		return output.Fail("file", source, "FILE_IS_DIR",
			fmt.Sprintf("path is a directory: %s", payload.Path), "", start)
	}

	encoding := payload.Encoding
	if encoding == "" {
		encoding = "utf8"
	}

	f, err := os.Open(payload.Path)
	if err != nil {
		return output.Fail("file", source, "FILE_PERMISSION",
			fmt.Sprintf("permission denied: %s", payload.Path), err.Error(), start)
	}
	defer f.Close()

	if payload.Offset > 0 {
		_, err = f.Seek(payload.Offset, 0)
		if err != nil {
			return output.Fail("file", source, "FILE_SEEK_ERROR", err.Error(), "", start)
		}
	}

	var bytesRead []byte
	if payload.Limit > 0 {
		buf := make([]byte, payload.Limit)
		n, err := f.Read(buf)
		if err != nil && n == 0 {
			return output.Fail("file", source, "FILE_READ_ERROR", err.Error(), "", start)
		}
		bytesRead = buf[:n]
	} else {
		bytesRead, err = os.ReadFile(payload.Path)
		if err != nil {
			return output.Fail("file", source, "FILE_READ_ERROR", err.Error(), "", start)
		}
	}

	result := &FileResult{
		Path:     payload.Path,
		Size:     int64(len(bytesRead)),
		Encoding: encoding,
	}

	if encoding == "binary" {
		result.ContentB64 = base64.StdEncoding.EncodeToString(bytesRead)
	} else {
		result.Content = string(bytesRead)
	}

	return output.Success("file", source, result, start)
}

func (h *FileHandler) parsePayload(data string) *FilePayload {
	payload := &FilePayload{Encoding: "utf8"}
	trimmed := strings.TrimSpace(data)
	if len(trimmed) == 0 {
		return payload
	}
	if trimmed[0] == '{' {
		if err := json.Unmarshal([]byte(trimmed), payload); err != nil {
			payload.Path = data
		}
	} else {
		payload.Path = data
	}
	return payload
}
