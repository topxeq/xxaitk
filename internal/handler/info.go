package handler

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/topxeq/xxaitk/internal/output"
)

type InfoHandler struct{}

type InfoPayload struct {
	Query string `json:"query"`
}

type OSInfoResult struct {
	Name    string `json:"name"`
	Arch    string `json:"arch"`
	Version string `json:"version"`
}

type CPUInfoResult struct {
	Count int `json:"count"`
}

type MemInfoResult struct {
	AllocKB      uint64 `json:"alloc_kb"`
	TotalAllocKB uint64 `json:"total_alloc_kb"`
	SysKB        uint64 `json:"sys_kb"`
	GCCount      uint32 `json:"gc_count"`
}

type EnvInfoResult struct {
	Vars map[string]string `json:"vars,omitempty"`
}

type AllInfoResult struct {
	OS  OSInfoResult  `json:"os"`
	CPU CPUInfoResult `json:"cpu"`
	Mem MemInfoResult `json:"mem"`
	Env EnvInfoResult `json:"env"`
}

func (h *InfoHandler) Handle(data string, source string) *output.Response {
	start := time.Now()

	payload := h.parsePayload(data)
	query := strings.ToLower(payload.Query)
	if query == "" {
		query = "all"
	}

	switch query {
	case "os":
		return output.Success("info", source, h.getOSInfo(), start)
	case "cpu":
		return output.Success("info", source, h.getCPUInfo(), start)
	case "mem":
		return output.Success("info", source, h.getMemInfo(), start)
	case "env":
		return output.Success("info", source, h.getEnvInfo(), start)
	case "all":
		return output.Success("info", source, &AllInfoResult{
			OS:  h.getOSInfo(),
			CPU: h.getCPUInfo(),
			Mem: h.getMemInfo(),
			Env: h.getEnvInfo(),
		}, start)
	default:
		return output.Fail("info", source, "INFO_UNKNOWN_QUERY",
			fmt.Sprintf("unknown query: %s (use os, cpu, mem, env, all)", query), "", start)
	}
}

func (h *InfoHandler) getOSInfo() OSInfoResult {
	return OSInfoResult{
		Name:    runtime.GOOS,
		Arch:    runtime.GOARCH,
		Version: runtime.Version(),
	}
}

func (h *InfoHandler) getCPUInfo() CPUInfoResult {
	return CPUInfoResult{Count: runtime.NumCPU()}
}

func (h *InfoHandler) getMemInfo() MemInfoResult {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return MemInfoResult{
		AllocKB:      m.Alloc / 1024,
		TotalAllocKB: m.TotalAlloc / 1024,
		SysKB:        m.Sys / 1024,
		GCCount:      m.NumGC,
	}
}

func (h *InfoHandler) getEnvInfo() EnvInfoResult {
	vars := map[string]string{}
	for _, key := range []string{"HOME", "PATH", "USER", "SHELL", "LANG", "PWD", "TERM"} {
		val := os.Getenv(key)
		if val != "" {
			vars[key] = val
		}
	}
	return EnvInfoResult{Vars: vars}
}

func (h *InfoHandler) parsePayload(data string) *InfoPayload {
	payload := &InfoPayload{}
	trimmed := strings.TrimSpace(data)
	if len(trimmed) == 0 {
		return payload
	}
	if trimmed[0] == '{' {
		json.Unmarshal([]byte(trimmed), payload)
	} else {
		payload.Query = data
	}
	return payload
}
