package handler

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/topxeq/xxaitk/internal/hexcodec"
	"github.com/topxeq/xxaitk/internal/output"
)

type DecodeHandler struct{}
type EncodeHandler struct{}
type B64EncHandler struct{}
type B64DecHandler struct{}
type URLEncHandler struct{}
type URLDecHandler struct{}

type EncodeResult struct {
	Input  string `json:"input,omitempty"`
	Output string `json:"output"`
	Format string `json:"format"`
}

func (h *DecodeHandler) Handle(data string, source string) *output.Response {
	start := time.Now()
	return output.Success("decode", source, &EncodeResult{
		Output: data,
		Format: "plaintext",
	}, start)
}

func (h *EncodeHandler) Handle(data string, source string) *output.Response {
	start := time.Now()
	result := hexcodec.EncodeString(data)
	return output.Success("encode", source, &EncodeResult{
		Input:  data,
		Output: result,
		Format: "hex",
	}, start)
}

func (h *B64EncHandler) Handle(data string, source string) *output.Response {
	start := time.Now()
	result := base64.StdEncoding.EncodeToString([]byte(data))
	return output.Success("b64enc", source, &EncodeResult{
		Input:  data,
		Output: result,
		Format: "base64",
	}, start)
}

func (h *B64DecHandler) Handle(data string, source string) *output.Response {
	start := time.Now()
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(data))
	if err != nil {
		return output.Fail("b64dec", source, "B64DEC_ERROR",
			fmt.Sprintf("base64 decode error: %s", err.Error()), "", start)
	}
	result := hexcodec.Encode(decoded)
	return output.Success("b64dec", source, &EncodeResult{
		Input:  data,
		Output: result,
		Format: "hex",
	}, start)
}

func (h *URLEncHandler) Handle(data string, source string) *output.Response {
	start := time.Now()
	result := url.QueryEscape(data)
	return output.Success("urlenc", source, &EncodeResult{
		Input:  data,
		Output: result,
		Format: "url_encoded",
	}, start)
}

func (h *URLDecHandler) Handle(data string, source string) *output.Response {
	start := time.Now()
	decoded, err := url.QueryUnescape(strings.TrimSpace(data))
	if err != nil {
		return output.Fail("urldec", source, "URLDEC_ERROR",
			fmt.Sprintf("URL decode error: %s", err.Error()), "", start)
	}
	result := hexcodec.EncodeString(decoded)
	return output.Success("urldec", source, &EncodeResult{
		Input:  data,
		Output: result,
		Format: "hex",
	}, start)
}
