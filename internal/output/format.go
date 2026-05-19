package output

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"
)

type Response struct {
	Ok      bool        `json:"ok"`
	Type    string      `json:"type"`
	Source  string      `json:"source,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	DurMs   int64       `json:"duration_ms,omitempty"`
	Env     *EnvInfo    `json:"env,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

type EnvInfo struct {
	OS    string `json:"os"`
	Arch  string `json:"arch"`
	Shell string `json:"shell,omitempty"`
}

func detectShell() string {
	if runtime.GOOS == "windows" {
		return "cmd.exe"
	}
	return "/bin/sh"
}

func Success(opType string, source string, data interface{}, startTime time.Time) *Response {
	resp := &Response{
		Ok:     true,
		Type:   opType,
		Source: source,
		Data:   data,
		DurMs:  time.Since(startTime).Milliseconds(),
	}
	if source != "" {
		resp.Env = &EnvInfo{
			OS:    runtime.GOOS,
			Arch:  runtime.GOARCH,
			Shell: detectShell(),
		}
	}
	return resp
}

func Fail(opType string, source string, code string, message string, detail string, startTime time.Time) *Response {
	resp := &Response{
		Ok:     false,
		Type:   opType,
		Source: source,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Detail:  detail,
		},
		DurMs: time.Since(startTime).Milliseconds(),
	}
	if source != "" {
		resp.Env = &EnvInfo{
			OS:    runtime.GOOS,
			Arch:  runtime.GOARCH,
			Shell: detectShell(),
		}
	}
	return resp
}

func (r *Response) Print() error {
	data, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func PrintError(code string, message string) {
	r := &Response{
		Ok:   false,
		Type: "unknown",
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}
	data, _ := json.Marshal(r)
	fmt.Fprintln(os.Stderr, string(data))
}

func PrintSuccess(opType string, data interface{}) {
	r := &Response{
		Ok:   true,
		Type: opType,
		Data: data,
	}
	data_bytes, _ := json.Marshal(r)
	fmt.Println(string(data_bytes))
}
