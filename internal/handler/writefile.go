package handler

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/topxeq/xxaitk/internal/datasource"
	"github.com/topxeq/xxaitk/internal/output"
)

type WriteFileHandler struct{}

type WriteFilePayload struct {
	Path       string `json:"path"`
	Content    string `json:"content,omitempty"`
	Source     string `json:"source,omitempty"`
	SourcePath string `json:"source_path,omitempty"`
	SourceURL  string `json:"source_url,omitempty"`
	Mode       string `json:"mode,omitempty"`
	Encoding   string `json:"encoding,omitempty"`
	Mkdir      bool   `json:"mkdir,omitempty"`
}

type WriteFileResult struct {
	Path    string `json:"path"`
	Size    int    `json:"size"`
	Mode    string `json:"mode"`
	Created bool   `json:"created"`
}

func (h *WriteFileHandler) Handle(data string, source string) *output.Response {
	start := time.Now()

	payload := h.parsePayload(data)
	if payload.Path == "" {
		return output.Fail("writefile", source, "WRITEFILE_EMPTY_PATH", "empty file path", "", start)
	}

	content := payload.Content
	if payload.Source != "" {
		resolved, err := datasource.ResolveJSONSource(payload.Source, payload.SourcePath, payload.SourceURL)
		if err != nil {
			return output.Fail("writefile", source, "WRITEFILE_SOURCE_ERROR", err.Error(), "", start)
		}
		content = resolved
	}

	if content == "" {
		return output.Fail("writefile", source, "WRITEFILE_EMPTY_CONTENT", "no content to write", "", start)
	}

	mode := payload.Mode
	if mode == "" {
		mode = "create"
	}

	if payload.Mkdir {
		idx := strings.LastIndex(payload.Path, "/")
		if idx > 0 {
			dir := payload.Path[:idx]
			os.MkdirAll(dir, 0755)
		}
	}

	exists := false
	if _, err := os.Stat(payload.Path); err == nil {
		exists = true
	}

	var writeErr error
	switch mode {
	case "append":
		f, err := os.OpenFile(payload.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return output.Fail("writefile", source, "WRITEFILE_OPEN_ERROR", err.Error(), "", start)
		}
		defer f.Close()
		_, writeErr = f.WriteString(content)
	case "overwrite":
		writeErr = os.WriteFile(payload.Path, []byte(content), 0644)
	case "create":
		if exists {
			return output.Fail("writefile", source, "WRITEFILE_EXISTS",
				fmt.Sprintf("file already exists: %s (use mode 'overwrite' or 'append')", payload.Path), "", start)
		}
		writeErr = os.WriteFile(payload.Path, []byte(content), 0644)
	default:
		return output.Fail("writefile", source, "WRITEFILE_INVALID_MODE",
			fmt.Sprintf("invalid mode: %s (use create, overwrite, or append)", mode), "", start)
	}

	if writeErr != nil {
		return output.Fail("writefile", source, "WRITEFILE_WRITE_ERROR", writeErr.Error(), "", start)
	}

	return output.Success("writefile", source, &WriteFileResult{
		Path:    payload.Path,
		Size:    len(content),
		Mode:    mode,
		Created: !exists,
	}, start)
}

func (h *WriteFileHandler) parsePayload(data string) *WriteFilePayload {
	payload := &WriteFilePayload{}
	trimmed := strings.TrimSpace(data)
	if len(trimmed) == 0 {
		return payload
	}
	if trimmed[0] == '{' {
		json.Unmarshal([]byte(trimmed), payload)
	}
	return payload
}
