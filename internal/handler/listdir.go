package handler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/topxeq/xxaitk/internal/output"
)

type ListDirHandler struct{}

type ListDirPayload struct {
	Path       string `json:"path"`
	Recursive  bool   `json:"recursive,omitempty"`
	Pattern    string `json:"pattern,omitempty"`
	ShowHidden bool   `json:"show_hidden,omitempty"`
}

type FileEntry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
}

type ListDirResult struct {
	Path    string      `json:"path"`
	Entries []FileEntry `json:"entries"`
	Count   int         `json:"count"`
}

func (h *ListDirHandler) Handle(data string, source string) *output.Response {
	start := time.Now()

	payload := h.parsePayload(data)
	if payload.Path == "" {
		return output.Fail("listdir", source, "LISTDIR_EMPTY_PATH", "empty directory path", "", start)
	}

	info, err := os.Stat(payload.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return output.Fail("listdir", source, "LISTDIR_NOT_FOUND",
				fmt.Sprintf("directory not found: %s", payload.Path), "", start)
		}
		return output.Fail("listdir", source, "LISTDIR_STAT_ERROR", err.Error(), "", start)
	}

	if !info.IsDir() {
		return output.Fail("listdir", source, "LISTDIR_NOT_DIR",
			fmt.Sprintf("path is not a directory: %s", payload.Path), "", start)
	}

	entries := []FileEntry{}
	err = h.walkDir(payload.Path, payload, &entries)
	if err != nil {
		return output.Fail("listdir", source, "LISTDIR_WALK_ERROR", err.Error(), "", start)
	}

	return output.Success("listdir", source, &ListDirResult{
		Path:    payload.Path,
		Entries: entries,
		Count:   len(entries),
	}, start)
}

func (h *ListDirHandler) walkDir(path string, payload *ListDirPayload, entries *[]FileEntry) error {
	entries_read, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries_read {
		name := entry.Name()
		if !payload.ShowHidden && strings.HasPrefix(name, ".") {
			continue
		}

		if payload.Pattern != "" {
			matched, _ := filepath.Match(payload.Pattern, name)
			if !matched {
				continue
			}
		}

		fullPath := filepath.Join(path, name)
		info, err := entry.Info()
		isDir := entry.IsDir()
		var size int64
		if err == nil {
			size = info.Size()
		}

		*entries = append(*entries, FileEntry{
			Name:  name,
			Path:  fullPath,
			IsDir: isDir,
			Size:  size,
		})

		if payload.Recursive && isDir {
			h.walkDir(fullPath, payload, entries)
		}
	}

	return nil
}

func (h *ListDirHandler) parsePayload(data string) *ListDirPayload {
	payload := &ListDirPayload{}
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
