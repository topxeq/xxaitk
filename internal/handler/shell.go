package handler

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/topxeq/xxaitk/internal/output"
)

type ShellHandler struct{}

type ShellPayload struct {
	Cmd     string            `json:"cmd"`
	Shell   string            `json:"shell,omitempty"`
	Timeout int               `json:"timeout,omitempty"`
	Cwd     string            `json:"cwd,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Stdin   string            `json:"stdin,omitempty"`
}

type ShellResult struct {
	ExitCode   int    `json:"exitcode"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	DurationMs int64  `json:"duration_ms"`
	Shell      string `json:"shell"`
	OS         string `json:"os"`
}

func (h *ShellHandler) Handle(data string, source string) *output.Response {
	start := time.Now()

	payload := h.parsePayload(data)
	if payload.Cmd == "" {
		return output.Fail("shell", source, "SHELL_EMPTY_CMD", "empty command", "", start)
	}

	shell := payload.Shell
	if shell == "" {
		shell = detectShell()
	}

	timeout := time.Duration(payload.Timeout) * time.Second
	if payload.Timeout <= 0 {
		timeout = 30 * time.Second
	}

	cmd := h.buildCommand(payload.Cmd, shell)
	if payload.Cwd != "" {
		cmd.Dir = payload.Cwd
	}

	if len(payload.Env) > 0 {
		cmd.Env = append(cmd.Environ(), envSlice(payload.Env)...)
	}

	if payload.Stdin != "" {
		cmd.Stdin = strings.NewReader(payload.Stdin)
	}

	resultCh := make(chan shellExecResult, 1)
	go func() {
		var stdout, stderr strings.Builder
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		resultCh <- shellExecResult{
			exitCode: h.exitCode(err),
			stdout:   stdout.String(),
			stderr:   stderr.String(),
			err:      err,
		}
	}()

	select {
	case r := <-resultCh:
		if r.exitCode != 0 {
			return output.Fail("shell", source, "SHELL_EXIT_NONZERO",
				fmt.Sprintf("command exited with code %d", r.exitCode),
				r.stderr, start)
		}
		return output.Success("shell", source, &ShellResult{
			ExitCode:   r.exitCode,
			Stdout:     r.stdout,
			Stderr:     r.stderr,
			DurationMs: time.Since(start).Milliseconds(),
			Shell:      shell,
			OS:         runtime.GOOS,
		}, start)
	case <-time.After(timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return output.Fail("shell", source, "SHELL_TIMEOUT",
			fmt.Sprintf("command timed out after %ds", int(timeout.Seconds())), "", start)
	}
}

func (h *ShellHandler) parsePayload(data string) *ShellPayload {
	payload := &ShellPayload{}

	trimmed := strings.TrimSpace(data)
	if len(trimmed) == 0 {
		return payload
	}

	if trimmed[0] == '{' {
		if err := json.Unmarshal([]byte(trimmed), payload); err != nil {
			payload.Cmd = data
		}
	} else {
		payload.Cmd = data
	}

	return payload
}

func (h *ShellHandler) buildCommand(cmdStr string, shell string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command(shell, "/c", cmdStr)
	}
	return exec.Command(shell, "-c", cmdStr)
}

func (h *ShellHandler) exitCode(err error) int {
	if err == nil {
		return 0
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode()
	}
	return -1
}

func detectShell() string {
	if runtime.GOOS == "windows" {
		return "cmd.exe"
	}
	return "/bin/sh"
}

func envSlice(env map[string]string) []string {
	s := make([]string, 0, len(env))
	for k, v := range env {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	return s
}

type shellExecResult struct {
	exitCode int
	stdout   string
	stderr   string
	err      error
}
