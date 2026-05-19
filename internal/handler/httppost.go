package handler

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/topxeq/xxaitk/internal/datasource"
	"github.com/topxeq/xxaitk/internal/output"
)

type HTTPPostHandler struct{}

type HTTPPostPayload struct {
	URL             string            `json:"url"`
	Body            string            `json:"body,omitempty"`
	ContentType     string            `json:"content_type,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	Insecure        bool              `json:"insecure,omitempty"`
	Timeout         int               `json:"timeout,omitempty"`
	FollowRedirects bool              `json:"follow_redirects,omitempty"`
	Source          string            `json:"source,omitempty"`
	SourcePath      string            `json:"source_path,omitempty"`
	SourceURL       string            `json:"source_url,omitempty"`
}

func (h *HTTPPostHandler) Handle(data string, source string) *output.Response {
	start := time.Now()

	payload := h.parsePayload(data)
	if payload.URL == "" {
		return output.Fail("httppost", source, "HTTP_EMPTY_URL", "empty URL", "", start)
	}

	body := payload.Body
	if payload.Source != "" {
		resolved, err := datasource.ResolveJSONSource(payload.Source, payload.SourcePath, payload.SourceURL)
		if err != nil {
			return output.Fail("httppost", source, "HTTP_SOURCE_ERROR", err.Error(), "", start)
		}
		body = resolved
	}

	timeout := time.Duration(payload.Timeout) * time.Second
	if payload.Timeout <= 0 {
		timeout = 30 * time.Second
	}

	transport := &http.Transport{}
	if payload.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
	if !payload.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	contentType := payload.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	req, err := http.NewRequest("POST", payload.URL, strings.NewReader(body))
	if err != nil {
		return output.Fail("httppost", source, "HTTP_REQUEST_ERROR", err.Error(), "", start)
	}
	req.Header.Set("Content-Type", contentType)

	for k, v := range payload.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return output.Fail("httppost", source, "HTTP_ERROR", err.Error(), "", start)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return output.Fail("httppost", source, "HTTP_READ_ERROR", err.Error(), "", start)
	}

	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	return output.Success("httppost", source, &HTTPResult{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       string(respBody),
		Size:       len(respBody),
		URL:        payload.URL,
	}, start)
}

func (h *HTTPPostHandler) parsePayload(data string) *HTTPPostPayload {
	payload := &HTTPPostPayload{FollowRedirects: true}
	trimmed := strings.TrimSpace(data)
	if len(trimmed) == 0 {
		return payload
	}
	if trimmed[0] == '{' {
		json.Unmarshal([]byte(trimmed), payload)
	}
	return payload
}
