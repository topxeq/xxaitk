package handler

import (
	"github.com/topxeq/xxaitk/internal/output"
)

type EvalHandler struct{}

func (h *EvalHandler) Handle(data string, source string) *output.Response {
	sh := &ScriptHandler{}
	return sh.Handle(data, source)
}
