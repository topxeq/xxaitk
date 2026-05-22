# xxAiToolkit (aitk)

**A one-stop CLI toolkit for AI agents** — covering 80%+ of common AI tool-use scenarios, so agents no longer need to hunt for various tools across the system.

AI agents today face two critical problems when interacting with the host system:

1. **Shell chaos**: Special characters (`$`, `` ` ``, `\`, `"`, `'`, `|`, `&`, `<`, `>`, spaces) behave inconsistently across bash/zsh/fish/PowerShell, causing repeated failures and even accidental operations.
2. **Tool fragmentation**: To accomplish common tasks — run a command, read/write a file, make an HTTP request, process JSON, encode/decode data, check system info — an AI agent must locate and learn `sh`, `cat`, `curl`, `jq`, `base64`, `xxd`, `uname`, `df`, and many more, each with its own syntax, edge cases, and output format.

**aitk solves both**: all arguments are hex-encoded (eliminating shell interpretation issues entirely), and all results are structured JSON (eliminating parsing ambiguity). One tool, one interface, one output format — covering shell execution, file I/O, HTTP requests, encoding/decoding, system introspection, SSH operations, and a built-in scripting language for complex logic.

## Install

**One-click install (Linux / macOS / Windows):**

```bash
curl -fsSL https://raw.githubusercontent.com/topxeq/xxaitk/main/install.sh | bash
```

> Custom install directory: `INSTALL_DIR=~/.local/bin curl -fsSL https://raw.githubusercontent.com/topxeq/xxaitk/main/install.sh | bash`

**Or with Go:**

```bash
go install github.com/topxeq/xxaitk@latest
```

**Or download from GitHub Releases:**

[https://github.com/topxeq/xxaitk/releases/latest](https://github.com/topxeq/xxaitk/releases/latest)

Available platforms: Linux (x86_64 / aarch64 / armv7 / armv6 / i386), macOS (x86_64 / aarch64), Windows (x86_64 / i386)

**Or build from source:**

```bash
git clone https://github.com/topxeq/xxaitk.git
cd xxaitk
go build -o aitk .
```

**Self-update:**

```bash
aitk update
```

## Usage

### Three Input Modes

aitk supports three ways to pass data — all produce the same JSON output:

**1. Hex mode** — the classic, unambiguous format for AI agents:

```bash
aitk SHELL_7b22636d64223a226c73227d
```

**2. Plaintext JSON mode** — for humans, no hex conversion needed:

```bash
aitk 'SHELL_{"cmd":"ls"}'
```

**3. Stdin mode** — pipe data in, great for composing with other tools:

```bash
echo '{"cmd":"ls"}' | aitk SHELL
```

### Single Argument Format

```
aitk <OPERATION>[_<SOURCE>]_<HEXDATA>       # hex mode
aitk '<OPERATION>_<JSON>'                    # plaintext JSON mode
echo '<JSON>' | aitk <OPERATION>             # stdin mode
```

- **OPERATION**: What to do (e.g. `SHELL`, `FILE`, `SCRIPT`)
- **SOURCE** (optional): `FILE` or `URL` — read command data from a file or URL instead of inline
- **HEXDATA**: Hex-encoded payload (for hex mode)
- **JSON**: Plaintext JSON object (for plaintext mode, must start with `{` or `[`)

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

### Library Commands

```
aitk lib list              List installed script libraries
aitk lib search            Search remote library registry
aitk lib get <name>        Download and install a library
aitk lib remove <name>     Remove an installed library
```

## Operation Prefixes

### Execution

| Prefix | Description | JSON Payload |
|--------|-------------|--------------|
| `SHELL` | Execute shell command | `{"cmd":"...","shell":"/bin/sh","timeout":30,"cwd":"/tmp","env":{},"stdin":"..."}` |
| `SCRIPT` | Execute built-in script | `{"source":"...","unsafe":false,"timeout":0,"args":[],"debug":false}` |
| `EVAL` | Evaluate expression (delegates to SCRIPT) | `{"source":"1+2"}` |

### Network

| Prefix | Description | JSON Payload |
|--------|-------------|--------------|
| `HTTPGET` | HTTP GET request | `{"url":"...","headers":{},"insecure":false,"timeout":30,"follow_redirects":true}` |
| `HTTPPOST` | HTTP POST request | `{"url":"...","body":"...","content_type":"...","headers":{},"insecure":false,"source":"file","source_path":"..."}` |
| `HTTPPUT` | HTTP PUT request | Same as HTTPPOST |
| `HTTPPATCH` | HTTP PATCH request | Same as HTTPPOST |
| `HTTPDELETE` | HTTP DELETE request | `{"url":"...","headers":{},"insecure":false}` |
| `PING` | Network connectivity test | `{"host":"...","timeout":5,"port":80}` |
| `NETDOWNLOAD` | Download file with verification | `{"url":"...","path":"...","insecure":false,"timeout":120,"headers":{},"resume":false,"verify":"...","mkdir":false,"overwrite":false}` |
| `PORT` | Check port or scan range | `{"host":"...","port":80}` or `{"host":"...","from":1,"to":1024,"timeout":2,"protocol":"tcp"}` |

### File System

| Prefix | Description | JSON Payload |
|--------|-------------|--------------|
| `FILE` / `READFILE` | Read file | `{"path":"...","encoding":"utf8","offset":0,"limit":1024}` |
| `WRITEFILE` | Write file | `{"path":"...","content":"...","mode":"create","encoding":"...","mkdir":false,"source":"file","source_path":"..."}` |
| `LISTDIR` | List directory | `{"path":"...","recursive":false,"pattern":"*.go","show_hidden":false}` |
| `DELETE` | Delete file/directory | `{"path":"...","recursive":true}` |

WRITEFILE modes: `create` (default, fails if exists), `overwrite`, `append`

### Encoding

| Prefix | Description | Input → Output |
|--------|-------------|----------------|
| `DECODE` | Hex decode | Hex → Plaintext |
| `ENCODE` | Hex encode | Plaintext → Hex |
| `B64ENC` | Base64 encode | Plaintext → Base64 |
| `B64DEC` | Base64 decode | Base64 → Hex |
| `URLENC` | URL encode | Plaintext → URL-encoded |
| `URLDEC` | URL decode | URL-encoded → Hex |

### System

| Prefix | Description |
|--------|-------------|
| `INFO` | System info — query: `os`, `cpu`, `mem`, `env`, `all` |
| `CAPABILITIES` | Query aitk capabilities — query: `version`, `prefixes`, `builtins`, `all` |

AI agents can call `aitk CAPABILITIES_616c6c` on first connection to discover available features.

### Crypto

| Prefix | Description | JSON Payload |
|--------|-------------|--------------|
| `HASH` | Hash data or file | `{"data":"hello","algo":"sha256"}` or `{"file":"/path","algo":"md5"}` |

Algorithms: `md5`, `sha1`, `sha256`, `sha512`

### Process Management

| Prefix | Description | JSON Payload |
|--------|-------------|--------------|
| `PROCESS` | Start/stop/list background processes | `{"action":"start","command":"...","id":"...","work_dir":"...","env":{},"shell":"...","timeout":0}` |

Actions: `start`, `status`, `stop`, `list`

> Note: Process state is in-memory. Use REPL mode (`:PROCESS ...`) to manage processes across commands.

### Version Control

| Prefix | Description | JSON Payload |
|--------|-------------|--------------|
| `GIT` | Git operations | `{"action":"status","repo":"/path"}` |

Actions: `status`, `log`, `diff`, `add`, `commit`, `branch`, `checkout`, `pull`, `push`, `fetch`, `tag`, `stash`, `merge`, `rebase`, `remote`, `clone`, `init`, `show`, `blame`, `revparse`, `reset`, `config`, `clean`

### Diff

| Prefix | Description | JSON Payload |
|--------|-------------|--------------|
| `DIFF` | Compare files or strings | `{"file_a":"...","file_b":"..."}` or `{"content_a":"...","content_b":"...","context":3}` |

### Archive

| Prefix | Description | JSON Payload |
|--------|-------------|--------------|
| `ARCHIVE` | Pack/unpack/list archives | `{"action":"pack","format":"tar.gz","dir":"/src","target":"/out.tar.gz"}` or `{"action":"pack","files":["..."],"target":"..."}` |

Actions: `pack`, `unpack`, `list`
Formats: `zip`, `tar`, `tar.gz` / `tgz`

### Database

| Prefix | Description | JSON Payload |
|--------|-------------|--------------|
| `SQL` | Execute SQL queries | `{"driver":"sqlite","dsn":"/path/db.db","query":"SELECT ?","args":[1],"max_rows":1000}` |

Supported drivers (all pure Go, no CGO):
- `sqlite` / `sqlite3` — [modernc.org/sqlite](https://modernc.org/sqlite)
- `mysql` — [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)
- `postgres` / `postgresql` / `pg` — [lib/pq](https://github.com/lib/pq)
- `mssql` / `sqlserver` — [go-mssqldb](https://github.com/microsoft/go-mssqldb)
- `oracle` / `ora` — [go-ora](https://github.com/sijms/go-ora)

### SSH

| Prefix | Description | JSON Payload |
|--------|-------------|--------------|
| `SSH` | Remote operations via SSH | `{"host":"...","port":22,"user":"root","password":"...","action":"cmd","cmd":"..."}` |

Auth: `password` or `key` (+ optional `key_passphrase`)

Actions:

| Action | Key Fields | Description |
|--------|-----------|-------------|
| `cmd` | `cmd` or `cmd_file` | Execute remote command |
| `upload` | `local_path`, `remote_path`, `recursive` | Upload file/directory |
| `download` | `local_path`, `remote_path`, `recursive` | Download file/directory |
| `upload_atomic` | `local_path`, `remote_path`, `temp_path` | Atomic upload via temp file |
| `mkdir` | `remote_path` | Create remote directory |
| `remove` | `remote_path`, `recursive` | Delete remote file/directory |
| `chmod` | `remote_path`, `mode` | Change file permissions |
| `move` | `source`, `target` | Move/rename remote file |
| `deploy` | `plan` or `plan_json` | Multi-step deployment |
| `sync` | `local_path`, `remote_path`, `direction` | File synchronization |

Sync directions: `push`, `pull`, `bidirectional`
Sync conflict strategies: `fail_on_conflict`, `newer_wins`, `local_wins`, `remote_wins`

Deploy step types: `cmd`, `upload`, `upload_atomic`, `download`, `mkdir`, `remove`, `chmod`, `move`, `sync`

## Data Source Modifiers

Any prefix supports `_FILE_` and `_URL_` to read data from external sources:

```bash
# Read command payload from a file (hex mode)
aitk SHELL_FILE_2f746d702f636d642e6a736f6e    # hex("/tmp/cmd.json")

# Read script source from a URL (hex mode)
aitk SCRIPT_URL_68747470733a2f2f6578616d706c65  # hex("https://example")

# Decode content from a file
aitk DECODE_FILE_2f746d702f686578646174612e747874

# Encode content from a URL
aitk ENCODE_URL_68747470733a2f2f6874747062696e2e6f72672f676574
```

WRITEFILE / HTTPPOST also support source modifiers via JSON fields:

```json
{"path":"/tmp/out.bin","source":"file","source_path":"/tmp/input.bin"}
{"url":"https://httpbin.org/post","source":"url","source_url":"https://example/data","content_type":"application/json"}
```

## Output Format

All output is JSON, easy for AI agents to parse:

**Success:**

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

**Error:**

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
x += 1                  // compound: +=, -=, *=, /=

// Negative indexing
let a = [10, 20, 30]
print(a[-1])            // 30

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
    i += 1
}

for item in [1, 2, 3] {
    print(item)
}

// break / continue
for i in [0, 1, 2, 3, 4] {
    if i == 3 { break }
    if i == 1 { continue }
    print(i)
}

// Functions and closures
fn add(a, b) {
    return a + b
}
print(add(3, 7))        // 10

fn make_counter() {
    let count = 0
    return fn() {
        count += 1
        return count
    }
}
let c = make_counter()
print(c())              // 1
print(c())              // 2

// Error handling
let r = try(fn() {
    return 1 / 0
})
// r = [false, "error message"]

let r2 = try(fn() {
    return 42
})
// r2 = [true, 42]
```

### Built-in Functions

| Category | Functions |
|----------|-----------|
| `str_*` | `str_len`, `str_concat`, `str_split`, `str_join`, `str_sub`, `str_trim`, `str_upper`, `str_lower`, `str_replace`, `str_has_prefix`, `str_has_suffix`, `str_contains`, `str_index`, `str_from_int`, `str_from_float`, `str_to_int`, `str_to_float`, `str_repeat`, `str_reverse` |
| `math_*` | `math_abs`, `math_max`, `math_min`, `math_floor`, `math_ceil`, `math_round`, `math_sqrt`, `math_pow`, `math_mod`, `math_rand`, `math_rand_int`, `math_log`, `math_exp`, `math_sin`, `math_cos` |
| `list_*` | `list_len`, `list_push`, `list_pop`, `list_shift`, `list_get`, `list_set`, `list_contains`, `list_index`, `list_join`, `list_map`, `list_filter`, `list_sort`, `list_reverse`, `list_slice`, `list_flat`, `list_reduce`, `list_find` |
| `map_*` | `map_get`, `map_set`, `map_has`, `map_keys`, `map_values`, `map_del`, `map_len`, `map_merge` |
| `json_*` | `json_encode`, `json_decode`, `json_get`, `json_set`, `json_has` |
| `io_*` | `io_read_file`, `io_write_file`, `io_append_file`, `io_exists`, `io_is_dir`, `io_is_file`, `io_list_dir`, `io_size`, `io_mkdir`, `io_copy`, `io_move`, `io_remove`, `io_temp_dir`, `io_abs_path` |
| `net_*` | `net_http_get`, `net_dns_lookup`, `net_tcp_connect` |
| `os_*` | `os_exec`, `os_env`, `os_getenv`, `os_cwd`, `os_hostname`, `os_platform`, `os_arch` (require `unsafe: true`) |
| `time_*` | `time_now`, `time_now_unix`, `time_format`, `time_parse`, `time_sleep`, `time_duration` |
| `log_*` | `log_info`, `log_warn`, `log_error`, `log_debug` |
| `type_*` | `type_of`, `type_is_nil`, `type_is_bool`, `type_is_int`, `type_is_float`, `type_is_string`, `type_is_list`, `type_is_map`, `type_is_fn` |
| `conv_*` | `conv_to_int`, `conv_to_float`, `conv_to_string`, `conv_to_bool`, `conv_hex_encode`, `conv_hex_decode`, `conv_b64_encode`, `conv_b64_decode` |
| `try` / `error` | `try(fn)` returns `[ok, result]`, `error("msg")` throws |

### Sandbox Mode

By default, scripts run in **safe mode** — dangerous functions like `os_exec`, `io_write_file`, `os_setenv` are disabled. Pass `"unsafe": true` in the JSON payload to unlock all functions.

```bash
# Safe mode (default) — only safe builtins available
aitk 'SCRIPT_{"source":"print(math_sqrt(144))"}'

# Unsafe mode — io_*, net_*, os_* unlocked
aitk 'SCRIPT_{"source":"print(os_hostname())","unsafe":true}'
```

## REPL

```
$ aitk
aitk v0.8.0 | type .help for help

aitk> let x = 10
aitk> print(math_abs(-5))
5

aitk> .help
  .help        Show help
  .prefixes    List all operation prefixes
  .builtins    List all script builtins
  .debug on    Enable debug mode
  .history     Show command history
  .quit        Exit REPL

  :<prefix> <args>   Execute prefix command (no hex encoding)
  <script>           Execute script statement

aitk> :shell ls -la /tmp
aitk> :readfile /etc/hosts
aitk> .quit
```

## Examples

### Three input modes — all produce identical output

```bash
# Hex mode:
aitk SHELL_7b22636d64223a226c73227d

# Plaintext JSON mode:
aitk 'SHELL_{"cmd":"ls"}'

# Stdin mode:
echo '{"cmd":"ls"}' | aitk SHELL
```

### Quick examples

```bash
# Execute shell command
aitk 'SHELL_{"cmd":"hostname"}'

# Read a file
aitk 'FILE_{"path":"/etc/hostname"}'

# Write a file
aitk 'WRITEFILE_{"path":"/tmp/test.txt","content":"hello"}'

# Append to file
aitk 'WRITEFILE_{"path":"/tmp/test.txt","content":"world","mode":"append"}'

# System info
aitk 'INFO_{"query":"os"}'

# Evaluate expression
aitk 'EVAL_{"source":"1 + 2 * 3"}'

# Run a script
aitk 'SCRIPT_{"source":"for i in [1, 2, 3] { print(i * i) }"}'

# HTTP GET
aitk 'HTTPGET_{"url":"https://httpbin.org/get"}'

# HTTP POST with JSON
aitk 'HTTPPOST_{"url":"https://httpbin.org/post","body":"{\"name\":\"aitk\"}","content_type":"application/json"}'

# SHA-256 hash
aitk 'HASH_{"data":"hello","algo":"sha256"}'

# SQLite query (in-memory)
aitk 'SQL_{"driver":"sqlite","dsn":":memory:","query":"SELECT 1 AS value"}'

# Parameterized query (prevents SQL injection)
aitk 'SQL_{"driver":"sqlite","dsn":"/tmp/test.db","query":"SELECT * FROM users WHERE age > ?","args":[25]}'

# Ping
aitk 'PING_{"host":"127.0.0.1"}'

# Port check
aitk 'PORT_{"host":"127.0.0.1","port":22}'

# Download file
aitk 'NETDOWNLOAD_{"url":"https://httpbin.org/json","path":"/tmp/data.json"}'

# Hex decode
aitk DECODE_68656c6c6f           # → "hello"

# Base64 encode
aitk B64ENC_68656c6c6f           # → "aGVsbG8="

# Diff two strings
aitk 'DIFF_{"content_a":"hello","content_b":"world"}'

# Archive a directory
aitk 'ARCHIVE_{"action":"pack","format":"tar.gz","dir":"/tmp/src","target":"/tmp/src.tar.gz"}'

# Git log
aitk 'GIT_{"action":"log","n":5}'

# Capabilities discovery
aitk CAPABILITIES_616c6c

# SSH execute remote command
aitk 'SSH_{"host":"1.2.3.4","port":22,"user":"root","password":"secret","action":"cmd","cmd":"hostname"}'

# SSH upload file
aitk 'SSH_{"host":"1.2.3.4","port":22,"user":"root","password":"secret","action":"upload","local_path":"/tmp/file.txt","remote_path":"/tmp/file.txt"}'

# SSH key authentication
aitk 'SSH_{"host":"1.2.3.4","port":22,"user":"root","key":"/root/.ssh/id_rsa","action":"cmd","cmd":"hostname"}'
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
| `SQL_EXEC_ERROR` | SQL execution error |
| `SSH_CONNECT_ERROR` | SSH connection failed |
| `SSH_AUTH_ERROR` | SSH authentication failed |
| `SSH_CMD_ERROR` | Remote command failed |
| `ARCHIVE_NO_FILES` | No files specified for packing |
| `ARCHIVE_NO_TARGET` | No target path specified |

## License

[MIT](LICENSE)
