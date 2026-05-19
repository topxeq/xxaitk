package handler

import (
	"time"

	"github.com/topxeq/xxaitk/internal/output"
)

type Handler interface {
	Handle(data string, source string) *output.Response
}

type Context struct {
	Debug   bool
	Timeout time.Duration
}

var registry = map[string]Handler{}

func Register(prefix string, h Handler) {
	registry[prefix] = h
}

func Get(prefix string) (Handler, bool) {
	h, ok := registry[prefix]
	return h, ok
}

func KnownPrefixes() []string {
	prefixes := make([]string, 0, len(registry))
	for k := range registry {
		prefixes = append(prefixes, k)
	}
	return prefixes
}

var knownPrefixes = map[string]bool{
	"SHELL":     true,
	"SCRIPT":    true,
	"EVAL":      true,
	"HTTPGET":   true,
	"HTTPPOST":  true,
	"HTTPPUT":   true,
	"HTTPPATCH": true,
	"HTTPDELETE": true,
	"FILE":      true,
	"READFILE":  true,
	"WRITEFILE": true,
	"LISTDIR":   true,
	"DELETE":    true,
	"INFO":      true,
	"DECODE":    true,
	"ENCODE":    true,
	"B64ENC":    true,
	"B64DEC":    true,
	"URLENC":    true,
	"URLDEC":    true,
	"PING":      true,
	"HASH":      true,
	"PROCESS":   true,
	"DIFF":      true,
	"ARCHIVE":   true,
	"SQL":       true,
	"GIT":       true,
	"PORT":      true,
	"NETDOWNLOAD": true,
	"CAPABILITIES": true,
}

func IsKnownPrefix(prefix string) bool {
	return knownPrefixes[prefix]
}
