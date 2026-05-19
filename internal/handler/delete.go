package handler

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/topxeq/xxaitk/internal/output"
)

type DeleteHandler struct{}

type DeletePayload struct {
	Path      string `json:"path"`
	Recursive bool   `json:"recursive,omitempty"`
	Force     bool   `json:"force,omitempty"`
}

type DeleteResult struct {
	Path    string `json:"path"`
	Deleted bool   `json:"deleted"`
	WasDir  bool   `json:"was_dir"`
}

func (h *DeleteHandler) Handle(data string, source string) *output.Response {
	start := time.Now()

	payload := h.parsePayload(data)
	if payload.Path == "" {
		return output.Fail("delete", source, "DELETE_EMPTY_PATH", "empty path", "", start)
	}

	info, err := os.Stat(payload.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return output.Fail("delete", source, "DELETE_NOT_FOUND",
				fmt.Sprintf("path not found: %s", payload.Path), "", start)
		}
		return output.Fail("delete", source, "DELETE_STAT_ERROR", err.Error(), "", start)
	}

	wasDir := info.IsDir()

	if wasDir {
		if payload.Recursive {
			err = os.RemoveAll(payload.Path)
		} else {
			err = os.Remove(payload.Path)
		}
	} else {
		err = os.Remove(payload.Path)
	}

	if err != nil {
		return output.Fail("delete", source, "DELETE_ERROR",
			fmt.Sprintf("failed to delete %s: %s", payload.Path, err.Error()), "", start)
	}

	return output.Success("delete", source, &DeleteResult{
		Path:    payload.Path,
		Deleted: true,
		WasDir:  wasDir,
	}, start)
}

func (h *DeleteHandler) parsePayload(data string) *DeletePayload {
	payload := &DeletePayload{}
	trimmed := strings.TrimSpace(data)
	if len(trimmed) == 0 {
		return payload
	}
	if trimmed[0] == '{' {
		json.Unmarshal([]byte(trimmed), payload)
	} else {
		payload.Path = data
	}
	return payload
}
