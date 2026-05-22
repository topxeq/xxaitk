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

	if query == "" || query == "all" {
		return output.Success("capabilities", source, h.buildCapabilities(), start)
	}

	switch {
	case query == "version":
		return output.Success("capabilities", source, map[string]interface{}{
			"version": Version,
			"go":      runtime.Version(),
		}, start)

	case query == "prefixes":
		return output.Success("capabilities", source, map[string]interface{}{
			"prefixes": KnownPrefixesList(),
		}, start)

	case query == "builtins":
		return output.Success("capabilities", source, map[string]interface{}{
			"builtins": BuiltinCategories(),
		}, start)

	case query == "schema":
		return output.Success("capabilities", source, map[string]interface{}{
			"schema": PrefixSchemas(),
		}, start)

	case strings.HasPrefix(query, "schema:"):
		prefix := strings.ToUpper(strings.TrimSpace(query[7:]))
		schemas := PrefixSchemas()
		s, ok := schemas[prefix]
		if !ok {
			return output.Fail("capabilities", source, "CAPABILITIES_UNKNOWN_PREFIX",
				"no schema for prefix: "+prefix,
				"use 'schema' to see all prefixes", start)
		}
		return output.Success("capabilities", source, map[string]interface{}{
			"prefix": prefix,
			"schema": s,
		}, start)

	default:
		return output.Fail("capabilities", source, "CAPABILITIES_UNKNOWN_QUERY",
			"unknown query: "+query,
			"use: version, prefixes, builtins, schema, schema:<PREFIX>, all", start)
	}
}

func (h *CapabilitiesHandler) buildCapabilities() map[string]interface{} {
	return map[string]interface{}{
		"version":   Version,
		"go":        runtime.Version(),
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
		"prefixes":  KnownPrefixesList(),
		"builtins":  BuiltinCategories(),
		"schema":    PrefixSchemas(),
		"features":  buildFeatures(),
		"input_modes": []map[string]string{
			{"mode": "hex", "format": "<PREFIX>_<HEXDATA>", "example": "SHELL_7b22636d64223a226c73227d"},
			{"mode": "plaintext", "format": "<PREFIX>_<JSON>", "example": "SHELL_{\"cmd\":\"ls\"}"},
			{"mode": "stdin", "format": "echo '<JSON>' | aitk <PREFIX>", "example": "echo '{\"cmd\":\"ls\"}' | aitk SHELL"},
		},
		"source_modifiers": []map[string]string{
			{"modifier": "FILE_", "description": "Read command data from file path (hex-encoded)"},
			{"modifier": "URL_", "description": "Read command data from URL (hex-encoded)"},
		},
	}
}

func buildFeatures() map[string]bool {
	return map[string]bool{
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
		"auto_update":      true,
		"local_capture":    true,
		"ssh":              true,
	}
}

var Version = "dev"

func KnownPrefixesList() []string {
	return []string{
		"SHELL", "SCRIPT", "EVAL",
		"HTTPGET", "HTTPPOST", "HTTPPUT", "HTTPPATCH", "HTTPDELETE",
		"FILE", "READFILE", "WRITEFILE", "LISTDIR", "DELETE",
		"INFO", "DECODE", "ENCODE", "B64ENC", "B64DEC", "URLENC", "URLDEC",
		"PING", "HASH", "PROCESS", "DIFF", "ARCHIVE", "SQL", "GIT",
		"PORT", "NETDOWNLOAD", "SSH", "CAPABILITIES",
	}
}

func BuiltinCategories() map[string][]string {
	return map[string][]string{
		"str": {
			"str_len", "str_concat", "str_split", "str_join", "str_sub",
			"str_trim", "str_upper", "str_lower", "str_replace",
			"str_has_prefix", "str_has_suffix", "str_contains", "str_index",
			"str_from_int", "str_from_float", "str_to_int", "str_to_float",
			"str_repeat", "str_reverse", "str_interp",
		},
		"math": {
			"math_abs", "math_max", "math_min", "math_floor", "math_ceil",
			"math_round", "math_sqrt", "math_pow", "math_mod",
			"math_rand", "math_rand_int", "math_log", "math_exp",
			"math_sin", "math_cos",
		},
		"list": {
			"list_len", "list_push", "list_pop", "list_shift",
			"list_get", "list_set", "list_contains", "list_index",
			"list_join", "list_map", "list_filter", "list_sort",
			"list_reverse", "list_slice", "list_flat",
			"list_reduce", "list_find",
		},
		"map": {
			"map_get", "map_set", "map_has", "map_keys",
			"map_values", "map_del", "map_len", "map_merge",
		},
		"json": {
			"json_encode", "json_decode", "json_get", "json_set", "json_has",
		},
		"io": {
			"io_read_file", "io_write_file", "io_append_file",
			"io_exists", "io_is_dir", "io_is_file", "io_list_dir",
			"io_size", "io_mkdir", "io_copy", "io_move",
			"io_remove", "io_temp_dir", "io_abs_path",
		},
		"net": {
			"net_http_get", "net_dns_lookup", "net_tcp_connect",
		},
		"os": {
			"os_exec", "os_env", "os_getenv", "os_cwd",
			"os_hostname", "os_platform", "os_arch",
		},
		"time": {
			"time_now", "time_now_unix", "time_format",
			"time_parse", "time_sleep", "time_duration",
		},
		"type": {
			"type_of", "type_is_nil", "type_is_bool", "type_is_int",
			"type_is_float", "type_is_string", "type_is_list",
			"type_is_map", "type_is_fn",
		},
		"conv": {
			"conv_to_int", "conv_to_float", "conv_to_string", "conv_to_bool",
			"conv_hex_encode", "conv_hex_decode", "conv_b64_encode", "conv_b64_decode",
		},
		"log": {
			"log_info", "log_warn", "log_error", "log_debug",
		},
		"error": {
			"try", "error",
		},
	}
}

type FieldDef struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Desc     string `json:"desc"`
}

type PrefixDef struct {
	Fields       []FieldDef          `json:"fields"`
	Actions      []string            `json:"actions,omitempty"`
	Modes        []string            `json:"modes,omitempty"`
	Algorithms   []string            `json:"algorithms,omitempty"`
	Drivers      []string            `json:"drivers,omitempty"`
	Queries      []string            `json:"queries,omitempty"`
	Directions   []string            `json:"directions,omitempty"`
	Conflicts    []string            `json:"conflicts,omitempty"`
	AuthMethods  []string            `json:"auth_methods,omitempty"`
	Formats      []string            `json:"formats,omitempty"`
	StepTypes    []string            `json:"step_types,omitempty"`
	InputModes   []map[string]string `json:"input_modes,omitempty"`
	Notes        []string            `json:"notes,omitempty"`
}

func PrefixSchemas() map[string]PrefixDef {
	return map[string]PrefixDef{
		"SHELL": {
			Fields: []FieldDef{
				{Name: "cmd", Type: "string", Required: true, Desc: "shell command to execute"},
				{Name: "shell", Type: "string", Required: false, Desc: "shell binary path (default: /bin/sh)"},
				{Name: "timeout", Type: "int", Required: false, Desc: "timeout in seconds (default: 30)"},
				{Name: "cwd", Type: "string", Required: false, Desc: "working directory"},
				{Name: "env", Type: "map[string]string", Required: false, Desc: "environment variables"},
				{Name: "stdin", Type: "string", Required: false, Desc: "stdin input for the command"},
			},
		},
		"SCRIPT": {
			Fields: []FieldDef{
				{Name: "source", Type: "string", Required: true, Desc: "script source code"},
				{Name: "unsafe", Type: "bool", Required: false, Desc: "enable io_*/net_*/os_* functions (default: false)"},
				{Name: "timeout", Type: "int", Required: false, Desc: "execution timeout in seconds"},
				{Name: "args", Type: "[]interface{}", Required: false, Desc: "arguments accessible as args[]"},
				{Name: "debug", Type: "bool", Required: false, Desc: "enable debug output (default: false)"},
			},
			Notes: []string{
				"Multi-statement scripts use newline separation (no semicolons)",
				"Safe mode by default; set unsafe:true to unlock io_*/net_*/os_*",
				"Built-in functions prefixed by category: str_*, math_*, list_*, map_*, json_*, conv_*, type_*, time_*, log_*",
				"Error handling: try(fn) returns [bool, result], error(msg) throws",
				"Control flow: break, continue in loops",
				"Closures capture variables by reference",
			},
		},
		"EVAL": {
			Fields: []FieldDef{
				{Name: "source", Type: "string", Required: true, Desc: "expression to evaluate (delegates to SCRIPT)"},
				{Name: "unsafe", Type: "bool", Required: false, Desc: "enable io_*/net_*/os_* functions"},
				{Name: "timeout", Type: "int", Required: false, Desc: "timeout in seconds"},
			},
			Notes: []string{"Same as SCRIPT, semantically for single expressions"},
		},
		"HTTPGET": {
			Fields: []FieldDef{
				{Name: "url", Type: "string", Required: true, Desc: "request URL"},
				{Name: "headers", Type: "map[string]string", Required: false, Desc: "custom headers"},
				{Name: "insecure", Type: "bool", Required: false, Desc: "skip TLS verification (default: false)"},
				{Name: "timeout", Type: "int", Required: false, Desc: "timeout in seconds (default: 30)"},
				{Name: "follow_redirects", Type: "bool", Required: false, Desc: "follow HTTP redirects (default: true)"},
			},
		},
		"HTTPPOST": {
			Fields: []FieldDef{
				{Name: "url", Type: "string", Required: true, Desc: "request URL"},
				{Name: "body", Type: "string", Required: false, Desc: "request body"},
				{Name: "content_type", Type: "string", Required: false, Desc: "Content-Type header"},
				{Name: "headers", Type: "map[string]string", Required: false, Desc: "custom headers"},
				{Name: "insecure", Type: "bool", Required: false, Desc: "skip TLS verification"},
				{Name: "timeout", Type: "int", Required: false, Desc: "timeout in seconds"},
				{Name: "follow_redirects", Type: "bool", Required: false, Desc: "follow redirects"},
				{Name: "source", Type: "string", Required: false, Desc: "body source: 'file' or 'url'"},
				{Name: "source_path", Type: "string", Required: false, Desc: "local file path when source=file"},
				{Name: "source_url", Type: "string", Required: false, Desc: "URL to fetch body from when source=url"},
			},
		},
		"HTTPPUT": {
			Fields: []FieldDef{
				{Name: "url", Type: "string", Required: true, Desc: "request URL"},
				{Name: "body", Type: "string", Required: false, Desc: "request body"},
				{Name: "content_type", Type: "string", Required: false, Desc: "Content-Type header"},
				{Name: "headers", Type: "map[string]string", Required: false, Desc: "custom headers"},
				{Name: "insecure", Type: "bool", Required: false, Desc: "skip TLS verification"},
				{Name: "timeout", Type: "int", Required: false, Desc: "timeout in seconds"},
				{Name: "source", Type: "string", Required: false, Desc: "body source: 'file' or 'url'"},
				{Name: "source_path", Type: "string", Required: false, Desc: "local file path when source=file"},
				{Name: "source_url", Type: "string", Required: false, Desc: "URL when source=url"},
			},
		},
		"HTTPPATCH": {
			Fields: []FieldDef{
				{Name: "url", Type: "string", Required: true, Desc: "request URL"},
				{Name: "body", Type: "string", Required: false, Desc: "request body"},
				{Name: "content_type", Type: "string", Required: false, Desc: "Content-Type header"},
				{Name: "headers", Type: "map[string]string", Required: false, Desc: "custom headers"},
				{Name: "insecure", Type: "bool", Required: false, Desc: "skip TLS verification"},
				{Name: "timeout", Type: "int", Required: false, Desc: "timeout in seconds"},
			},
		},
		"HTTPDELETE": {
			Fields: []FieldDef{
				{Name: "url", Type: "string", Required: true, Desc: "request URL"},
				{Name: "body", Type: "string", Required: false, Desc: "request body"},
				{Name: "headers", Type: "map[string]string", Required: false, Desc: "custom headers"},
				{Name: "insecure", Type: "bool", Required: false, Desc: "skip TLS verification"},
				{Name: "timeout", Type: "int", Required: false, Desc: "timeout in seconds"},
			},
		},
		"FILE": {
			Fields: []FieldDef{
				{Name: "path", Type: "string", Required: true, Desc: "file path to read"},
				{Name: "encoding", Type: "string", Required: false, Desc: "'utf8' (default) or 'binary' (returns base64)"},
				{Name: "offset", Type: "int64", Required: false, Desc: "read offset in bytes (default: 0)"},
				{Name: "limit", Type: "int64", Required: false, Desc: "max bytes to read"},
			},
			Notes: []string{"READFILE is an alias for FILE"},
		},
		"WRITEFILE": {
			Fields: []FieldDef{
				{Name: "path", Type: "string", Required: true, Desc: "file path to write"},
				{Name: "content", Type: "string", Required: false, Desc: "content to write"},
				{Name: "mode", Type: "string", Required: false, Desc: "write mode"},
				{Name: "encoding", Type: "string", Required: false, Desc: "'utf8' (default) or 'binary'"},
				{Name: "mkdir", Type: "bool", Required: false, Desc: "create parent directories (default: false)"},
				{Name: "source", Type: "string", Required: false, Desc: "content source: 'file' or 'url'"},
				{Name: "source_path", Type: "string", Required: false, Desc: "local file path when source=file"},
				{Name: "source_url", Type: "string", Required: false, Desc: "URL when source=url"},
			},
			Modes: []string{"create", "overwrite", "append"},
			Notes: []string{"Default mode is 'create' (fails if file exists)"},
		},
		"LISTDIR": {
			Fields: []FieldDef{
				{Name: "path", Type: "string", Required: true, Desc: "directory path"},
				{Name: "recursive", Type: "bool", Required: false, Desc: "list recursively (default: false)"},
				{Name: "pattern", Type: "string", Required: false, Desc: "glob filter pattern (e.g. '*.go')"},
				{Name: "show_hidden", Type: "bool", Required: false, Desc: "include hidden files (default: false)"},
			},
		},
		"DELETE": {
			Fields: []FieldDef{
				{Name: "path", Type: "string", Required: true, Desc: "path to delete"},
				{Name: "recursive", Type: "bool", Required: false, Desc: "delete directory recursively (default: false)"},
			},
		},
		"INFO": {
			Fields: []FieldDef{
				{Name: "query", Type: "string", Required: true, Desc: "info category to query"},
			},
			Queries: []string{"os", "cpu", "mem", "env", "all"},
		},
		"DECODE": {
			InputModes: []map[string]string{
				{"mode": "hex", "desc": "input is hex-encoded data, decoded to plaintext"},
			},
			Notes: []string{"Takes raw data (not JSON). Input must be hex-encoded in CLI mode."},
		},
		"ENCODE": {
			InputModes: []map[string]string{
				{"mode": "hex", "desc": "input is hex-decoded first, then re-encoded as hex"},
			},
			Notes: []string{"Takes raw data (not JSON). Input must be hex-encoded in CLI mode."},
		},
		"B64ENC": {
			Notes: []string{"Takes raw data, outputs base64 string. Input must be hex-encoded in CLI mode."},
		},
		"B64DEC": {
			Notes: []string{"Takes base64 string, outputs hex string. Input must be hex-encoded in CLI mode."},
		},
		"URLENC": {
			Notes: []string{"Takes raw data, outputs URL-encoded string. Input must be hex-encoded in CLI mode."},
		},
		"URLDEC": {
			Notes: []string{"Takes URL-encoded string, outputs hex string. Input must be hex-encoded in CLI mode."},
		},
		"PING": {
			Fields: []FieldDef{
				{Name: "host", Type: "string", Required: true, Desc: "hostname or IP"},
				{Name: "timeout", Type: "int", Required: false, Desc: "timeout in seconds (default: 5)"},
				{Name: "port", Type: "int", Required: false, Desc: "port to test (default: 0 = ICMP)"},
			},
		},
		"NETDOWNLOAD": {
			Fields: []FieldDef{
				{Name: "url", Type: "string", Required: true, Desc: "download URL"},
				{Name: "path", Type: "string", Required: true, Desc: "local file path to save"},
				{Name: "insecure", Type: "bool", Required: false, Desc: "skip TLS verification"},
				{Name: "timeout", Type: "int", Required: false, Desc: "timeout in seconds (default: 120)"},
				{Name: "headers", Type: "map[string]string", Required: false, Desc: "custom HTTP headers"},
				{Name: "resume", Type: "bool", Required: false, Desc: "resume partial download"},
				{Name: "verify", Type: "string", Required: false, Desc: "SHA-256 checksum to verify"},
				{Name: "mkdir", Type: "bool", Required: false, Desc: "create parent directories"},
				{Name: "overwrite", Type: "bool", Required: false, Desc: "overwrite existing file"},
			},
		},
		"HASH": {
			Fields: []FieldDef{
				{Name: "data", Type: "string", Required: false, Desc: "data to hash (mutually exclusive with file)"},
				{Name: "file", Type: "string", Required: false, Desc: "file path to hash (mutually exclusive with data)"},
				{Name: "algo", Type: "string", Required: false, Desc: "hash algorithm (default: sha256)"},
			},
			Algorithms: []string{"md5", "sha1", "sha256", "sha512"},
		},
		"PROCESS": {
			Fields: []FieldDef{
				{Name: "action", Type: "string", Required: true, Desc: "process action"},
				{Name: "command", Type: "string", Required: false, Desc: "command to start (for action=start)"},
				{Name: "id", Type: "string", Required: false, Desc: "process ID (auto-generated if empty)"},
				{Name: "shell", Type: "string", Required: false, Desc: "shell binary path"},
				{Name: "timeout", Type: "int", Required: false, Desc: "timeout in seconds"},
				{Name: "work_dir", Type: "string", Required: false, Desc: "working directory"},
				{Name: "env", Type: "map[string]string", Required: false, Desc: "environment variables"},
			},
			Actions: []string{"start", "status", "stop", "list"},
			Notes:   []string{"Process state is in-memory. Use REPL mode for cross-command management."},
		},
		"DIFF": {
			Fields: []FieldDef{
				{Name: "content_a", Type: "string", Required: false, Desc: "first content (or use file_a)"},
				{Name: "content_b", Type: "string", Required: false, Desc: "second content (or use file_b)"},
				{Name: "file_a", Type: "string", Required: false, Desc: "first file path"},
				{Name: "file_b", Type: "string", Required: false, Desc: "second file path"},
				{Name: "context", Type: "int", Required: false, Desc: "context lines (default: 3)"},
			},
		},
		"ARCHIVE": {
			Fields: []FieldDef{
				{Name: "action", Type: "string", Required: true, Desc: "archive action"},
				{Name: "format", Type: "string", Required: false, Desc: "archive format (default: zip)"},
				{Name: "source", Type: "string", Required: false, Desc: "source archive path (for unpack/list)"},
				{Name: "target", Type: "string", Required: false, Desc: "target archive path (for pack) or extract dir"},
				{Name: "dir", Type: "string", Required: false, Desc: "directory to pack"},
				{Name: "files", Type: "[]string", Required: false, Desc: "specific files to pack"},
				{Name: "overwrite", Type: "bool", Required: false, Desc: "overwrite existing archive"},
			},
			Actions: []string{"pack", "unpack", "list"},
			Formats: []string{"zip", "tar", "tar.gz", "tgz"},
			Notes:   []string{"Pack requires either 'dir' or 'files' field"},
		},
		"SQL": {
			Fields: []FieldDef{
				{Name: "driver", Type: "string", Required: true, Desc: "database driver name"},
				{Name: "dsn", Type: "string", Required: true, Desc: "data source name / connection string"},
				{Name: "query", Type: "string", Required: true, Desc: "SQL query or statement"},
				{Name: "args", Type: "[]interface{}", Required: false, Desc: "parameterized query arguments"},
				{Name: "max_rows", Type: "int", Required: false, Desc: "max result rows (default: 1000)"},
			},
			Drivers: []string{"sqlite", "sqlite3", "mysql", "postgres", "postgresql", "pg", "mssql", "sqlserver", "oracle", "ora"},
			Notes:   []string{"All drivers are pure Go (no CGO)", "Use args field for parameterized queries to prevent SQL injection"},
		},
		"GIT": {
			Fields: []FieldDef{
				{Name: "action", Type: "string", Required: true, Desc: "git action"},
				{Name: "repo", Type: "string", Required: false, Desc: "repository path (default: current dir)"},
				{Name: "url", Type: "string", Required: false, Desc: "remote URL (for clone)"},
				{Name: "files", Type: "[]string", Required: false, Desc: "files to add"},
				{Name: "message", Type: "string", Required: false, Desc: "commit message"},
				{Name: "branch", Type: "string", Required: false, Desc: "branch name"},
				{Name: "remote", Type: "string", Required: false, Desc: "remote name"},
				{Name: "tag", Type: "string", Required: false, Desc: "tag name"},
				{Name: "n", Type: "int", Required: false, Desc: "number of log entries"},
				{Name: "file", Type: "string", Required: false, Desc: "single file path"},
				{Name: "author", Type: "string", Required: false, Desc: "author filter (log) or override (commit)"},
				{Name: "since", Type: "string", Required: false, Desc: "since date filter (log)"},
				{Name: "until", Type: "string", Required: false, Desc: "until date filter (log)"},
				{Name: "depth", Type: "int", Required: false, Desc: "clone depth"},
				{Name: "force", Type: "bool", Required: false, Desc: "force operation"},
				{Name: "all", Type: "bool", Required: false, Desc: "all branches/remotes"},
				{Name: "staged", Type: "bool", Required: false, Desc: "staged changes only (diff)"},
				{Name: "args", Type: "[]string", Required: false, Desc: "extra arguments (revparse, config)"},
			},
			Actions: []string{"status", "log", "diff", "add", "commit", "branch", "checkout", "pull", "push", "fetch", "tag", "stash", "merge", "rebase", "remote", "clone", "init", "show", "blame", "revparse", "reset", "config", "clean"},
		},
		"PORT": {
			Fields: []FieldDef{
				{Name: "host", Type: "string", Required: true, Desc: "hostname or IP"},
				{Name: "port", Type: "int", Required: false, Desc: "single port to check"},
				{Name: "from", Type: "int", Required: false, Desc: "port range start"},
				{Name: "to", Type: "int", Required: false, Desc: "port range end"},
				{Name: "timeout", Type: "int", Required: false, Desc: "timeout in seconds (default: 2)"},
				{Name: "protocol", Type: "string", Required: false, Desc: "'tcp' (default) or 'udp'"},
			},
		},
		"SSH": {
			Fields: []FieldDef{
				{Name: "host", Type: "string", Required: true, Desc: "SSH server hostname or IP"},
				{Name: "port", Type: "int", Required: false, Desc: "SSH port (default: 22)"},
				{Name: "user", Type: "string", Required: true, Desc: "SSH username"},
				{Name: "password", Type: "string", Required: false, Desc: "password for auth (use password OR key)"},
				{Name: "key", Type: "string", Required: false, Desc: "private key file path (use password OR key)"},
				{Name: "key_passphrase", Type: "string", Required: false, Desc: "passphrase for encrypted key"},
				{Name: "strict_host_key", Type: "bool", Required: false, Desc: "strict host key checking (default: false)"},
				{Name: "known_hosts", Type: "string", Required: false, Desc: "known_hosts file path"},
				{Name: "timeout", Type: "string", Required: false, Desc: "connection timeout (e.g. '10s')"},
				{Name: "action", Type: "string", Required: true, Desc: "SSH action to perform"},
				{Name: "cmd", Type: "string", Required: false, Desc: "remote command (action=cmd)"},
				{Name: "cmd_file", Type: "string", Required: false, Desc: "local file with commands (action=cmd)"},
				{Name: "cmd_timeout", Type: "string", Required: false, Desc: "command timeout (e.g. '30s')"},
				{Name: "local_path", Type: "string", Required: false, Desc: "local file/dir path"},
				{Name: "remote_path", Type: "string", Required: false, Desc: "remote file/dir path"},
				{Name: "source", Type: "string", Required: false, Desc: "source path (action=move)"},
				{Name: "target_path", Type: "string", Required: false, Desc: "target path (action=move)"},
				{Name: "temp_path", Type: "string", Required: false, Desc: "temp file path (action=upload_atomic)"},
				{Name: "mode", Type: "string", Required: false, Desc: "file permission mode (action=chmod, e.g. '0755')"},
				{Name: "plan", Type: "string", Required: false, Desc: "deploy plan file path (action=deploy)"},
				{Name: "plan_json", Type: "string", Required: false, Desc: "deploy plan JSON string (action=deploy)"},
				{Name: "direction", Type: "string", Required: false, Desc: "sync direction (action=sync)"},
				{Name: "recursive", Type: "bool", Required: false, Desc: "recursive operation"},
				{Name: "delete", Type: "bool", Required: false, Desc: "delete extraneous files (sync)"},
				{Name: "dry_run", Type: "bool", Required: false, Desc: "preview only, no changes (sync)"},
				{Name: "conflict", Type: "string", Required: false, Desc: "conflict resolution strategy (sync)"},
				{Name: "include", Type: "[]string", Required: false, Desc: "include glob patterns (sync)"},
				{Name: "exclude", Type: "[]string", Required: false, Desc: "exclude glob patterns (sync)"},
			},
			Actions:     []string{"cmd", "upload", "download", "upload_atomic", "mkdir", "remove", "chmod", "move", "deploy", "sync"},
			AuthMethods: []string{"password", "key"},
			Directions:  []string{"push", "pull", "bidirectional"},
			Conflicts:   []string{"fail_on_conflict", "newer_wins", "local_wins", "remote_wins"},
			StepTypes:   []string{"cmd", "upload", "upload_atomic", "download", "mkdir", "remove", "chmod", "move", "sync"},
			Notes: []string{
				"Use password OR key for authentication (not both)",
				"deploy: plan_json is a JSON string with {steps:[...], continue_on_error:bool}",
				"Each deploy step has: name, type (see step_types), plus type-specific fields",
			},
		},
		"CAPABILITIES": {
			Fields: []FieldDef{
				{Name: "query", Type: "string", Required: false, Desc: "capabilities query (default: 'all')"},
			},
			Queries: []string{"version", "prefixes", "builtins", "schema", "schema:<PREFIX>", "all"},
			Notes: []string{
				"schema returns JSON payload definitions for all prefixes",
				"schema:<PREFIX> returns definition for one prefix (e.g. schema:SHELL)",
				"AI agents should call CAPABILITIES_616c6c on first connection",
			},
		},
	}
}
