package handler

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/topxeq/xxaitk/internal/output"
)

type PingHandler struct{}

type PingPayload struct {
	Host    string `json:"host"`
	Timeout int    `json:"timeout,omitempty"`
	Port    int    `json:"port,omitempty"`
}

type PingResult struct {
	Host    string `json:"host"`
	Reachable bool `json:"reachable"`
	IP      string `json:"ip,omitempty"`
	LatencyMs int64 `json:"latency_ms,omitempty"`
	Port    int    `json:"port,omitempty"`
}

func (h *PingHandler) Handle(data string, source string) *output.Response {
	start := time.Now()

	payload := h.parsePayload(data)
	if payload.Host == "" {
		return output.Fail("ping", source, "PING_EMPTY_HOST", "empty host", "", start)
	}

	timeout := time.Duration(payload.Timeout) * time.Second
	if payload.Timeout <= 0 {
		timeout = 10 * time.Second
	}

	port := payload.Port
	if port <= 0 {
		port = 80
	}

	resolveStart := time.Now()
	ipAddrs, err := net.LookupHost(payload.Host)
	if err != nil {
		return output.Success("ping", source, &PingResult{
			Host:      payload.Host,
			Reachable: false,
		}, start)
	}

	ip := ipAddrs[0]
	resolveTime := time.Since(resolveStart)

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), timeout)
	if err != nil {
		return output.Success("ping", source, &PingResult{
			Host:      payload.Host,
			Reachable: false,
			IP:        ip,
			Port:      port,
		}, start)
	}
	conn.Close()

	return output.Success("ping", source, &PingResult{
		Host:      payload.Host,
		Reachable: true,
		IP:        ip,
		LatencyMs: resolveTime.Milliseconds(),
		Port:      port,
	}, start)
}

func (h *PingHandler) parsePayload(data string) *PingPayload {
	payload := &PingPayload{}
	trimmed := strings.TrimSpace(data)
	if len(trimmed) == 0 {
		return payload
	}
	if trimmed[0] == '{' {
		json.Unmarshal([]byte(trimmed), payload)
	} else {
		payload.Host = data
	}
	return payload
}
