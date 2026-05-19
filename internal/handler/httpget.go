package handler

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/topxeq/xxaitk/internal/output"
)

type HTTPGetHandler struct{}

type HTTPGetPayload struct {
	URL             string            `json:"url"`
	Headers         map[string]string `json:"headers,omitempty"`
	Insecure        bool              `json:"insecure,omitempty"`
	Timeout         int               `json:"timeout,omitempty"`
	FollowRedirects bool              `json:"follow_redirects,omitempty"`
}

type HTTPResult struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body"`
	Size       int               `json:"size"`
	URL        string            `json:"url"`
}

func (h *HTTPGetHandler) Handle(data string, source string) *output.Response {
	start := time.Now()

	payload := h.parsePayload(data)
	if payload.URL == "" {
		return output.Fail("httpget", source, "HTTP_EMPTY_URL", "empty URL", "", start)
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

	req, err := http.NewRequest("GET", payload.URL, nil)
	if err != nil {
		return output.Fail("httpget", source, "HTTP_REQUEST_ERROR", err.Error(), "", start)
	}

	for k, v := range payload.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return output.Fail("httpget", source, "HTTP_ERROR", err.Error(), "", start)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return output.Fail("httpget", source, "HTTP_READ_ERROR", err.Error(), "", start)
	}

	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	return output.Success("httpget", source, &HTTPResult{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       string(body),
		Size:       len(body),
		URL:        payload.URL,
	}, start)
}

func (h *HTTPGetHandler) parsePayload(data string) *HTTPGetPayload {
	payload := &HTTPGetPayload{FollowRedirects: true}
	trimmed := strings.TrimSpace(data)
	if len(trimmed) == 0 {
		return payload
	}
	if trimmed[0] == '{' {
		json.Unmarshal([]byte(trimmed), payload)
	} else {
		payload.URL = data
	}
	return payload
}
