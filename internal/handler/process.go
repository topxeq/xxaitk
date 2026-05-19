package handler

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/topxeq/xxaitk/internal/output"
)

type ProcessHandler struct{}

type ProcessPayload struct {
	Action   string            `json:"action"`
	Command  string            `json:"command"`
	Shell    string            `json:"shell,omitempty"`
	ID       string            `json:"id,omitempty"`
	Timeout  int               `json:"timeout,omitempty"`
	WorkDir  string            `json:"work_dir,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
	Signal   string            `json:"signal,omitempty"`
}

type ProcessResult struct {
	Action  string `json:"action"`
	ID      string `json:"id,omitempty"`
	PID     int    `json:"pid,omitempty"`
	Status  string `json:"status,omitempty"`
	Stdout  string `json:"stdout,omitempty"`
	Stderr  string `json:"stderr,omitempty"`
	ExitCode int   `json:"exitcode,omitempty"`
}

var (
	procMu     sync.Mutex
	procMap    = map[string]*exec.Cmd{}
	procOutput = map[string]*processOutput{}
)

type processOutput struct {
	stdout string
	stderr string
	done   bool
	code   int
}

func (h *ProcessHandler) Handle(data string, source string) *output.Response {
	start := time.Now()

	payload := h.parsePayload(data)
	action := strings.ToLower(payload.Action)
	if action == "" {
		if payload.Command != "" {
			action = "start"
		} else {
			return output.Fail("process", source, "PROCESS_NO_ACTION",
				"no action specified (start, status, stop, list)", "", start)
		}
	}

	switch action {
	case "start":
		return h.startProcess(payload, source, start)
	case "status":
		return h.statusProcess(payload, source, start)
	case "stop":
		return h.stopProcess(payload, source, start)
	case "list":
		return h.listProcesses(source, start)
	default:
		return output.Fail("process", source, "PROCESS_UNKNOWN_ACTION",
			fmt.Sprintf("unknown action: %s (use start, status, stop, list)", action), "", start)
	}
}

func (h *ProcessHandler) startProcess(payload *ProcessPayload, source string, start time.Time) *output.Response {
	if payload.Command == "" {
		return output.Fail("process", source, "PROCESS_NO_COMMAND", "no command specified", "", start)
	}

	shell := payload.Shell
	if shell == "" {
		if runtime.GOOS == "windows" {
			shell = "cmd.exe"
		} else {
			shell = "/bin/sh"
		}
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command(shell, "/c", payload.Command)
	} else {
		cmd = exec.Command(shell, "-c", payload.Command)
	}

	if payload.WorkDir != "" {
		cmd.Dir = payload.WorkDir
	}
	if len(payload.Env) > 0 {
		cmd.Env = append(cmd.Environ(), envSlice(payload.Env)...)
	}

	id := payload.ID
	if id == "" {
		id = fmt.Sprintf("p%d", time.Now().UnixNano())
	}

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return output.Fail("process", source, "PROCESS_START_ERROR", err.Error(), "", start)
	}

	procMu.Lock()
	procMap[id] = cmd
	procOutput[id] = &processOutput{code: -1}
	procMu.Unlock()

	go func() {
		err := cmd.Wait()
		procMu.Lock()
		if po, ok := procOutput[id]; ok {
			po.stdout = stdout.String()
			po.stderr = stderr.String()
			po.done = true
			if err != nil {
				po.code = -1
			} else {
				po.code = 0
			}
		}
		procMu.Unlock()
	}()

	return output.Success("process", source, &ProcessResult{
		Action: "start",
		ID:     id,
		PID:    cmd.Process.Pid,
		Status: "running",
	}, start)
}

func (h *ProcessHandler) statusProcess(payload *ProcessPayload, source string, start time.Time) *output.Response {
	id := payload.ID
	if id == "" {
		return output.Fail("process", source, "PROCESS_NO_ID", "no process id specified", "", start)
	}

	procMu.Lock()
	cmd, cmdOk := procMap[id]
	po, poOk := procOutput[id]
	procMu.Unlock()

	if !cmdOk {
		return output.Fail("process", source, "PROCESS_NOT_FOUND",
			fmt.Sprintf("process %s not found", id), "", start)
	}

	result := &ProcessResult{
		Action: "status",
		ID:     id,
		PID:    cmd.Process.Pid,
	}

	if poOk && po.done {
		result.Status = "exited"
		result.ExitCode = po.code
		result.Stdout = po.stdout
		result.Stderr = po.stderr
	} else {
		result.Status = "running"
	}

	return output.Success("process", source, result, start)
}

func (h *ProcessHandler) stopProcess(payload *ProcessPayload, source string, start time.Time) *output.Response {
	id := payload.ID
	if id == "" {
		return output.Fail("process", source, "PROCESS_NO_ID", "no process id specified", "", start)
	}

	procMu.Lock()
	cmd, ok := procMap[id]
	procMu.Unlock()

	if !ok {
		return output.Fail("process", source, "PROCESS_NOT_FOUND",
			fmt.Sprintf("process %s not found", id), "", start)
	}

	if err := cmd.Process.Kill(); err != nil {
		return output.Fail("process", source, "PROCESS_KILL_ERROR", err.Error(), "", start)
	}

	procMu.Lock()
	delete(procMap, id)
	procMu.Unlock()

	return output.Success("process", source, &ProcessResult{
		Action: "stop",
		ID:     id,
		Status: "killed",
	}, start)
}

func (h *ProcessHandler) listProcesses(source string, start time.Time) *output.Response {
	procMu.Lock()
	defer procMu.Unlock()

	type procInfo struct {
		ID     string `json:"id"`
		PID    int    `json:"pid"`
		Status string `json:"status"`
	}

	var list []procInfo
	for id, cmd := range procMap {
		status := "running"
		if po, ok := procOutput[id]; ok && po.done {
			status = "exited"
		}
		list = append(list, procInfo{ID: id, PID: cmd.Process.Pid, Status: status})
	}

	return output.Success("process", source, map[string]interface{}{
		"processes": list,
		"count":     len(list),
	}, start)
}

func (h *ProcessHandler) parsePayload(data string) *ProcessPayload {
	payload := &ProcessPayload{}
	trimmed := strings.TrimSpace(data)
	if len(trimmed) == 0 {
		return payload
	}
	if trimmed[0] == '{' {
		json.Unmarshal([]byte(trimmed), payload)
	} else {
		payload.Command = data
		payload.Action = "start"
	}
	return payload
}
