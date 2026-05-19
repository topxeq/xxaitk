package handler

import (
	"encoding/json"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/topxeq/xxaitk/internal/output"
)

type PortHandler struct{}

type PortPayload struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	From     int    `json:"from"`
	To       int    `json:"to"`
	Timeout  int    `json:"timeout,omitempty"`
	Protocol string `json:"protocol,omitempty"`
}

type PortResult struct {
	Host    string      `json:"host"`
	Results []PortEntry `json:"results"`
	Count   int         `json:"count"`
}

type PortEntry struct {
	Port    int    `json:"port"`
	Open    bool   `json:"open"`
	Service string `json:"service,omitempty"`
}

var commonPorts = map[int]string{
	21: "ftp", 22: "ssh", 23: "telnet", 25: "smtp", 53: "dns",
	80: "http", 110: "pop3", 143: "imap", 443: "https", 993: "imaps",
	995: "pop3s", 3306: "mysql", 5432: "postgresql", 6379: "redis",
	8080: "http-alt", 8443: "https-alt", 27017: "mongodb", 1521: "oracle",
	1433: "mssql", 636: "ldaps", 3389: "rdp", 5900: "vnc",
	11211: "memcached", 5672: "amqp", 9200: "elasticsearch",
}

func (h *PortHandler) Handle(data string, source string) *output.Response {
	start := time.Now()

	payload := h.parsePayload(data)
	host := payload.Host
	if host == "" {
		host = "localhost"
	}

	timeout := time.Duration(payload.Timeout) * time.Millisecond
	if payload.Timeout <= 0 {
		timeout = 2000 * time.Millisecond
	}

	protocol := payload.Protocol
	if protocol == "" {
		protocol = "tcp"
	}

	if payload.Port > 0 {
		entry := h.checkPort(host, payload.Port, timeout, protocol)
		return output.Success("port", source, &PortResult{
			Host:    host,
			Results: []PortEntry{entry},
			Count:   1,
		}, start)
	}

	fromPort := payload.From
	toPort := payload.To
	if fromPort <= 0 {
		fromPort = 1
	}
	if toPort <= 0 {
		toPort = 1024
	}
	if toPort > 65535 {
		toPort = 65535
	}

	results := h.scanPorts(host, fromPort, toPort, timeout, protocol)

	return output.Success("port", source, &PortResult{
		Host:    host,
		Results: results,
		Count:   len(results),
	}, start)
}

func (h *PortHandler) checkPort(host string, port int, timeout time.Duration, protocol string) PortEntry {
	target := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout(protocol, target, timeout)
	if err != nil {
		return PortEntry{Port: port, Open: false}
	}
	conn.Close()
	service := commonPorts[port]
	return PortEntry{Port: port, Open: true, Service: service}
}

func (h *PortHandler) scanPorts(host string, from, to int, timeout time.Duration, protocol string) []PortEntry {
	var results []PortEntry
	var mu sync.Mutex
	var wg sync.WaitGroup

	sem := make(chan struct{}, 100)

	for port := from; port <= to; port++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(p int) {
			defer wg.Done()
			defer func() { <-sem }()
			entry := h.checkPort(host, p, timeout, protocol)
			if entry.Open {
				mu.Lock()
				results = append(results, entry)
				mu.Unlock()
			}
		}(port)
	}
	wg.Wait()

	return results
}

func (h *PortHandler) parsePayload(data string) *PortPayload {
	payload := &PortPayload{}
	trimmed := strings.TrimSpace(data)
	if len(trimmed) == 0 {
		return payload
	}
	if trimmed[0] == '{' {
		json.Unmarshal([]byte(trimmed), payload)
	} else {
		if strings.Contains(data, ":") {
			parts := strings.SplitN(data, ":", 2)
			payload.Host = parts[0]
			port, _ := strconv.Atoi(parts[1])
			payload.Port = port
		}
	}
	return payload
}
