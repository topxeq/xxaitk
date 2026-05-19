package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/topxeq/xxaitk/internal/output"
)

type NetDownloadHandler struct{}

type NetDownloadPayload struct {
	URL       string            `json:"url"`
	Path      string            `json:"path"`
	Insecure  bool              `json:"insecure,omitempty"`
	Timeout   int               `json:"timeout,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	Resume    bool              `json:"resume,omitempty"`
	Verify    string            `json:"verify,omitempty"`
	Mkdir     bool              `json:"mkdir,omitempty"`
	Overwrite bool              `json:"overwrite,omitempty"`
}

type NetDownloadResult struct {
	URL         string `json:"url"`
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	Status      string `json:"status"`
	ContentType string `json:"content_type,omitempty"`
	Sha256      string `json:"sha256,omitempty"`
	Resumed     bool   `json:"resumed,omitempty"`
}

func (h *NetDownloadHandler) Handle(data string, source string) *output.Response {
	start := time.Now()

	payload := h.parsePayload(data)
	if payload.URL == "" {
		return output.Fail("netdownload", source, "NETDOWNLOAD_NO_URL", "no URL specified", "", start)
	}
	if payload.Path == "" {
		return output.Fail("netdownload", source, "NETDOWNLOAD_NO_PATH", "no target path specified", "", start)
	}

	timeout := time.Duration(payload.Timeout) * time.Second
	if payload.Timeout <= 0 {
		timeout = 120 * time.Second
	}

	if payload.Mkdir {
		idx := strings.LastIndex(payload.Path, "/")
		if idx > 0 {
			os.MkdirAll(payload.Path[:idx], 0755)
		}
	}

	if !payload.Overwrite && !payload.Resume {
		if _, err := os.Stat(payload.Path); err == nil {
			return output.Fail("netdownload", source, "NETDOWNLOAD_EXISTS",
				fmt.Sprintf("file exists: %s (use overwrite:true or resume:true)", payload.Path), "", start)
		}
	}

	client := &http.Client{Timeout: timeout}
	if payload.Insecure {
		client.Transport = &http.Transport{}
	}

	req, err := http.NewRequest("GET", payload.URL, nil)
	if err != nil {
		return output.Fail("netdownload", source, "NETDOWNLOAD_REQUEST_ERROR", err.Error(), "", start)
	}

	resumed := false
	if payload.Resume {
		if info, err := os.Stat(payload.Path); err == nil {
			req.Header.Set("Range", "bytes="+strconv.FormatInt(info.Size(), 10)+"-")
			resumed = true
		}
	}

	for k, v := range payload.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return output.Fail("netdownload", source, "NETDOWNLOAD_ERROR", err.Error(), "", start)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return output.Fail("netdownload", source, "NETDOWNLOAD_HTTP_ERROR",
			fmt.Sprintf("HTTP %d", resp.StatusCode), "", start)
	}

	var out *os.File
	if resumed && resp.StatusCode == 206 {
		out, err = os.OpenFile(payload.Path, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			out, _ = os.Create(payload.Path)
			resumed = false
		}
	} else {
		out, err = os.Create(payload.Path)
	}
	if err != nil {
		return output.Fail("netdownload", source, "NETDOWNLOAD_CREATE_ERROR", err.Error(), "", start)
	}
	defer out.Close()

	hasher := sha256.New()
	writers := io.MultiWriter(out, hasher)

	written, err := io.Copy(writers, resp.Body)
	if err != nil {
		return output.Fail("netdownload", source, "NETDOWNLOAD_WRITE_ERROR", err.Error(), "", start)
	}

	sha256Hash := hex.EncodeToString(hasher.Sum(nil))

	contentType := resp.Header.Get("Content-Type")

	return output.Success("netdownload", source, &NetDownloadResult{
		URL:         payload.URL,
		Path:        payload.Path,
		Size:        written,
		Status:      "downloaded",
		ContentType: contentType,
		Sha256:      sha256Hash,
		Resumed:     resumed,
	}, start)
}

func (h *NetDownloadHandler) parsePayload(data string) *NetDownloadPayload {
	payload := &NetDownloadPayload{}
	trimmed := strings.TrimSpace(data)
	if len(trimmed) == 0 {
		return payload
	}
	if trimmed[0] == '{' {
		json.Unmarshal([]byte(trimmed), payload)
	}
	return payload
}
