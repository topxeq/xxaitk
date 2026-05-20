package handler

import (
	"runtime"
	"strings"
	"time"

	"github.com/topxeq/xxaitk/internal/output"
)

type CapabilitiesHandler struct{}

func (h *CapabilitiesHandler) Handle(data string, source string) *output.Response {
	start := time.Now()

	query := strings.ToLower(strings.TrimSpace(data))

	caps := h.buildCapabilities()

	if query != "" && query != "all" {
		switch query {
		case "version":
			return output.Success("capabilities", source, map[string]interface{}{
				"version": Version,
				"go":      runtime.Version(),
			}, start)
		case "prefixes":
			return output.Success("capabilities", source, map[string]interface{}{
				"prefixes": KnownPrefixesList(),
			}, start)
		case "builtins":
			return output.Success("capabilities", source, map[string]interface{}{
				"builtins": BuiltinCategories(),
			}, start)
		default:
			return output.Fail("capabilities", source, "CAPABILITIES_UNKNOWN_QUERY",
				"unknown query: "+query, "use: version, prefixes, builtins, all", start)
		}
	}

	return output.Success("capabilities", source, caps, start)
}

func (h *CapabilitiesHandler) buildCapabilities() map[string]interface{} {
	return map[string]interface{}{
		"version": Version,
		"go":      runtime.Version(),
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
		"prefixes": KnownPrefixesList(),
		"builtins": BuiltinCategories(),
		"features": map[string]bool{
			"hex_codec":         true,
			"file_source":       true,
			"url_source":       true,
			"script_engine":     true,
			"closure":           true,
			"mutable_closures":  true,
			"sandbox":           true,
			"sql_drivers":       true,
			"process_mgmt":      true,
			"archive":          true,
			"diff":              true,
			"hash":              true,
			"break_continue":    true,
			"closure_iterators": true,
			"compound_assign":   true,
			"string_iteration":  true,
			"unicode_safe":      true,
			"negative_indexing": true,
			"const_enforcement": true,
			"error_builtin":     true,
			"local_capture":     true,
		},
	}
}

var Version = "0.6.0"

func KnownPrefixesList() []string {
	return []string{
		"SHELL", "SCRIPT", "EVAL",
		"HTTPGET", "HTTPPOST", "HTTPPUT", "HTTPPATCH", "HTTPDELETE",
		"FILE", "READFILE", "WRITEFILE", "LISTDIR", "DELETE",
		"INFO", "DECODE", "ENCODE", "B64ENC", "B64DEC", "URLENC", "URLDEC",
		"PING", "HASH", "PROCESS", "DIFF", "ARCHIVE", "SQL", "GIT",
		"PORT", "NETDOWNLOAD", "CAPABILITIES",
	}
}

func BuiltinCategories() map[string][]string {
	return map[string][]string{
		"str":  {"str_len", "str_concat", "str_split", "str_join", "str_sub", "str_trim", "str_upper", "str_lower", "str_replace", "str_interp", "..."},
		"math": {"math_abs", "math_max", "math_min", "math_floor", "math_ceil", "math_round", "math_sqrt", "math_pow", "..."},
		"list": {"list_len", "list_push", "list_pop", "list_get", "list_set", "list_map", "list_filter", "list_sort", "..."},
		"map":  {"map_get", "map_set", "map_has", "map_keys", "map_values", "map_del", "map_len", "map_merge"},
		"json": {"json_encode", "json_decode", "json_get", "json_set", "json_has"},
		"io":   {"io_read_file", "io_write_file", "io_append_file", "io_exists", "io_mkdir", "io_copy", "io_move", "..."},
		"net":  {"net_http_get", "net_dns_lookup", "net_tcp_connect"},
		"os":   {"os_exec", "os_env", "os_getenv", "os_cwd", "os_hostname", "..."},
		"time": {"time_now", "time_now_unix", "time_format", "time_parse", "time_sleep"},
		"type": {"type_of", "type_is_nil", "type_is_bool", "type_is_int", "type_is_string", "..."},
		"conv": {"conv_to_int", "conv_to_float", "conv_to_string", "conv_hex_encode", "conv_b64_encode", "..."},
		"error": {"try", "catch", "is_error"},
	}
}
