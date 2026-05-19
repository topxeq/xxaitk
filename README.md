# xxAiToolkit (aitk)

**A one-stop CLI toolkit for AI agents** — covering 80%+ of common AI tool-use scenarios, so agents no longer need to hunt for various tools across the system.

AI agents today face two critical problems when interacting with the host system:

1. **Shell chaos**: Special characters (`$`, `` ` ``, `\`, `"`, `'`, `|`, `&`, `<`, `>`, spaces) behave inconsistently across bash/zsh/fish/PowerShell, causing repeated failures and even accidental operations.
2. **Tool fragmentation**: To accomplish common tasks — run a command, read/write a file, make an HTTP request, process JSON, encode/decode data, check system info — an AI agent must locate and learn `sh`, `cat`, `curl`, `jq`, `base64`, `xxd`, `uname`, `df`, and many more, each with its own syntax, edge cases, and output format.

**aitk solves both**: all arguments are hex-encoded (eliminating shell interpretation issues entirely), and all results are structured JSON (eliminating parsing ambiguity). One tool, one interface, one output format — covering shell execution, file I/O, HTTP requests, encoding/decoding, system introspection, and a built-in scripting language for complex logic.

```
aitk SHELL_6c73202d6c61        # execute shell command
aitk FILE_2f6574632f686f737473  # read a file
aitk HTTPGET_687474703a2f2f...  # make HTTP request
aitk INFO_616c6c               # get system info
aitk SCRIPT_7072696e742822...   # run complex logic
```

Every operation returns consistent JSON. No more guessing output formats. No more shell escaping hell.

## Install

```bash
go install github.com/topxeq/xxaitk@latest
```

Or build from source:

```bash
git clone https://github.com/topxeq/xxaitk.git
cd xxaitk
go build -o aitk .
```

## Usage

### Single Argument Mode

```
aitk <OPERATION>[_<SOURCE>]_<HEXDATA>
```

- **OPERATION**: What to do (e.g. `SHELL`, `FILE`, `SCRIPT`)
- **SOURCE** (optional): `FILE` or `URL` — read command data from a file or URL instead of inline
- **HEXDATA**: Hex-encoded payload

### REPL Mode

```
aitk                    # Enter interactive REPL
```

### Flags

| Flag | Description |
|------|-------------|
| `--version`, `-v` | Print version |
| `--help`, `-h` | Print help |
| `--debug` | Enable debug output |

## Operation Prefixes

### Execution

| Prefix | Description | Hex Decoded Format |
|--------|-------------|-------------------|
| `SHELL` | Execute shell command | String or `{"cmd":"...","shell":"bash","timeout":30,"cwd":"/tmp"}` |
| `SCRIPT` | Execute built-in script | Script source or `{"source":"...","unsafe":false,"debug":false}` |
| `EVAL` | Evaluate expression (single-line SCRIPT) | Expression string |

### Network

| Prefix | Description | Hex Decoded Format |
|--------|-------------|-------------------|
| `HTTPGET` | HTTP GET request | URL string or `{"url":"...","headers":{},"insecure":false}` |
| `HTTPPOST` | HTTP POST request | `{"url":"...","body":"...","content_type":"application/json"}` |
| `PING` | Network connectivity test | Host string or `{"host":"...","port":80}` |

### File System

| Prefix | Description | Hex Decoded Format |
|--------|-------------|-------------------|
| `FILE` / `READFILE` | Read file | Path string or `{"path":"...","encoding":"utf8","range":{"offset":0,"limit":1024}}` |
| `WRITEFILE` | Write file | `{"path":"...","content":"...","mode":"create"}` |
| `LISTDIR` | List directory | Path string or `{"path":"...","recursive":false,"pattern":"*.go"}` |
| `DELETE` | Delete file/directory | Path string or `{"path":"...","recursive":true}` |

### Encoding

| Prefix | Description | Input → Output |
|--------|-------------|----------------|
| `DECODE` | Hex decode | Hex → Plaintext (in JSON) |
| `ENCODE` | Hex encode | Plaintext → Hex string |
| `B64ENC` | Base64 encode | Plaintext → Base64 string |
| `B64DEC` | Base64 decode | Base64 string → Hex string |
| `URLENC` | URL encode | Plaintext → URL-encoded string |
| `URLDEC` | URL decode | URL-encoded → Hex string |

### System

| Prefix | Description |
|--------|-------------|
| `INFO` | System info (`os`, `cpu`, `mem`, `env`, `all`) |

## Data Source Modifiers

When the payload is too large for a command-line argument, use `FILE_` or `URL_` to read data from an external source:

```bash
# Read shell commands from a file
aitk SHELL_FILE_2f746d702f636d642e7368     # hex("/tmp/cmd.sh")

# Read script source from a URL
aitk SCRIPT_URL_68747470733a2f2f6578616d706c65  # hex("https://example")
```

WRITEFILE_ can also reference external sources via JSON:

```json
{
  "path": "/tmp/output.bin",
  "source": "file",
  "source_path": "/tmp/input.bin"
}
```

## Output Format

All output is JSON, easy for AI agents to parse:

```json
{
  "ok": true,
  "type": "shell",
  "source": "inline",
  "data": {
    "exitcode": 0,
    "stdout": "file1\nfile2\n",
    "stderr": "",
    "duration_ms": 23,
    "shell": "/bin/sh",
    "os": "linux"
  },
  "duration_ms": 23,
  "env": {"os": "linux", "arch": "amd64", "shell": "/bin/sh"}
}
```

Error response:

```json
{
  "ok": false,
  "type": "shell",
  "source": "inline",
  "error": {
    "code": "SHELL_TIMEOUT",
    "message": "command timed out after 30s",
    "detail": ""
  }
}
```

## Built-in Script Language

The `SCRIPT_` and `EVAL_` prefixes execute a purpose-built scripting language designed for AI agents — no module system, all functions built-in, categorized by prefix.

### Data Types

| Type | Literal | Example |
|------|---------|---------|
| nil | `nil` | |
| bool | `true` / `false` | |
| int | integer | `42`, `-1`, `0xFF` |
| float | floating point | `3.14`, `1e10` |
| string | double-quoted | `"hello\nworld"` |
| list | square brackets | `[1, 2, "three"]` |
| map | curly braces | `{"key": "value"}` |

### Syntax

```javascript
// Variables
let x = 10
const PI = 3.14
x = x + 1                // reassignment

// Conditionals
if x > 5 {
    print("big")
} elif x > 2 {
    print("medium")
} else {
    print("small")
}

// Loops
while i < 10 {
    i = i + 1
}

for item in [1, 2, 3] {
    print(item)
}

// Functions
fn add(a, b) {
    return a + b
}
print(str_from_int(add(3, 7)))   // 10
```

### Built-in Functions

| Category | Functions |
|----------|-----------|
| `str_*` | `str_len`, `str_concat`, `str_split`, `str_join`, `str_sub`, `str_trim`, `str_upper`, `str_lower`, `str_replace`, `str_has_prefix`, `str_has_suffix`, `str_contains`, `str_index`, `str_from_int`, `str_from_float`, `str_to_int`, `str_to_float`, `str_repeat`, `str_reverse` |
| `math_*` | `math_abs`, `math_max`, `math_min`, `math_floor`, `math_ceil`, `math_round`, `math_sqrt`, `math_pow`, `math_mod`, `math_rand`, `math_rand_int`, `math_log`, `math_exp`, `math_sin`, `math_cos` |
| `list_*` | `list_len`, `list_push`, `list_pop`, `list_shift`, `list_get`, `list_set`, `list_contains`, `list_index`, `list_join`, `list_map`, `list_filter`, `list_sort`, `list_reverse`, `list_slice`, `list_flat`, `list_reduce`, `list_find` |
| `map_*` | `map_get`, `map_set`, `map_has`, `map_keys`, `map_values`, `map_del`, `map_len`, `map_merge` |
| `json_*` | `json_encode`, `json_decode`, `json_get`, `json_set`, `json_has` |
| `io_*` | `io_read_file`, `io_write_file`, `io_append_file`, `io_exists`, `io_is_dir`, `io_is_file`, `io_list_dir`, `io_size` |
| `os_*` | `os_exec`, `os_env`, `os_getenv`, `os_cwd`, `os_hostname`, `os_platform`, `os_arch` (require `unsafe: true`) |
| `time_*` | `time_now`, `time_now_unix`, `time_format`, `time_parse`, `time_sleep`, `time_duration` |
| `log_*` | `log_info`, `log_warn`, `log_error`, `log_debug` |
| `type_*` | `type_of`, `type_is_nil`, `type_is_bool`, `type_is_int`, `type_is_float`, `type_is_string`, `type_is_list`, `type_is_map`, `type_is_fn` |
| `conv_*` | `conv_to_int`, `conv_to_float`, `conv_to_string`, `conv_to_bool`, `conv_hex_encode`, `conv_hex_decode`, `conv_b64_encode`, `conv_b64_decode` |

### Sandbox Mode

By default, scripts run in **safe mode** — dangerous functions like `os_exec`, `io_write_file`, `os_setenv` are disabled. Pass `"unsafe": true` in the JSON payload to unlock all functions.

```bash
# Safe mode (default)
aitk SCRIPT_7072696e74282268656c6c6f2229           # print("hello")

# Unsafe mode — hex({"source":"os_exec(\"rm -rf /tmp/test\")","unsafe":true})
aitk SCRIPT_7b22736f75726365223a226f735f65786563285c22726d202d7266202f746d702f746573745c2229222c22756e73616665223a747275657d
```

## REPL

```
$ aitk
aitk v0.1.0 | type .help for help

aitk> let x = 10
aitk> print(math_abs(-5))
5

aitk> .help
  .help        Show help
  .prefixes    List all operation prefixes
  .builtins    List all script builtins
  .debug on    Enable debug mode
  .quit        Exit REPL

aitk> :shell ls -la /tmp           # prefix shortcut (no hex needed)
aitk> :readfile /etc/hosts
aitk> .quit
```

## Error Codes

| Code | Description |
|------|-------------|
| `PARSE_ERROR` | Invalid argument format |
| `UNKNOWN_PREFIX` | Unknown operation prefix |
| `SOURCE_RESOLVE_ERROR` | Failed to resolve FILE/URL source |
| `HEX_DECODE_FAIL` | Invalid hex data |
| `SHELL_EXIT_NONZERO` | Shell command exited non-zero |
| `SHELL_TIMEOUT` | Shell command timed out |
| `SCRIPT_COMPILE_ERROR` | Script parse/compile error |
| `SCRIPT_RUNTIME_ERROR` | Script runtime error |
| `FILE_NOT_FOUND` | File not found |
| `FILE_PERMISSION` | Permission denied |
| `FILE_IS_DIR` | Path is a directory (expected file) |
| `WRITEFILE_EXISTS` | File exists (use mode `overwrite` or `append`) |
| `HTTP_ERROR` | HTTP request failed |
| `HTTP_TIMEOUT` | HTTP request timed out |

## Examples

```bash
# Execute "ls -la /tmp"
aitk SHELL_6c73202d6c61202f746d70

# Read /etc/hosts
aitk FILE_2f6574632f686f737473

# System info
aitk INFO_6f73                    # "os"

# Evaluate expression
aitk EVAL_31202b2032202b2033      # "1 + 2 + 3" → 6

# Script: print hello
aitk SCRIPT_7072696e74282268656c6c6f2229

# Write file
aitk WRITEFILE_7b2270617468223a222f746d702f746573742e747874222c22636f6e74656e74223a2248656c6c6f227d

# HTTP GET
aitk HTTPGET_687474703a2f2f6874747062696e2e6f72672f676574

# Hex decode
aitk DECODE_68656c6c6f20776f726c64  # → "hello world"

# Base64 decode
aitk B64DEC_614756736247383d        # "aGVsbG8=" → hex of "hello"
```

## License

[MIT](LICENSE)
