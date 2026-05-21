# aitk SSH Examples

This directory contains working examples for the `SSH` prefix in aitk, demonstrating full [sshrun](https://github.com/topxeq/sshrun) feature parity.

## Setup

1. Edit `common.sh` and set your server details:

```bash
SSH_HOST="your.server.ip"
SSH_PORT="22"
SSH_USER="root"
SSH_PASS="your_password"
# Or use key auth:
# SSH_KEY="~/.ssh/id_rsa"
```

2. Make sure `aitk` is in your PATH:

```bash
export AITK=/path/to/aitk
# or
# ln -s /path/to/aitk /usr/local/bin/aitk
```

3. Run any example:

```bash
bash 01_cmd_basic.sh
```

## Examples

| # | File | Description |
|---|------|-------------|
| 01 | `01_cmd_basic.sh` | Execute remote commands (hostname, uname, df) |
| 02 | `02_cmd_timeout.sh` | Commands with timeout, timeout exceeded |
| 03 | `03_cmd_file.sh` | Execute multiple commands from a file |
| 04 | `04_upload.sh` | Upload file, upload with custom filename |
| 05 | `05_download.sh` | Download file, download with custom filename |
| 06 | `06_upload_atomic.sh` | Atomic upload (temp file + rename) |
| 07 | `07_mkdir_chmod_move_remove.sh` | Directory/file operations |
| 08 | `08_deploy.sh` | Multi-step deployment plan, continue_on_error |
| 09 | `09_sync_push.sh` | Sync push with include/exclude, delete, dry-run |
| 10 | `10_sync_pull.sh` | Sync pull, pull with delete |
| 11 | `11_sync_bidirectional.sh` | Bidirectional sync with conflict resolution |
| 12 | `12_sync_single_file.sh` | Single file sync (push and pull) |
| 13 | `13_key_auth_and_errors.sh` | Key-based auth, error handling examples |

## How It Works

All SSH operations use the `SSH_` prefix with hex-encoded JSON payloads:

```bash
# Build JSON payload
json='{"host":"1.2.3.4","port":22,"user":"root","password":"secret","action":"cmd","cmd":"hostname"}'

# Hex-encode it
hex=$(echo -n "$json" | xxd -p | tr -d '\n')

# Run via aitk
aitk "SSH_${hex}"
```

The `common.sh` helper wraps this into a simple `ssh_run` function:

```bash
source common.sh
ssh_run '"action":"cmd","cmd":"hostname"'   # auto-prepends connection params
```

## SSH Actions Reference

| Action | Required Fields | Description |
|--------|----------------|-------------|
| `cmd` | `cmd` or `cmd_file` | Execute remote command(s) |
| `upload` | `local_path`, `remote_path` | Upload file |
| `download` | `local_path`, `remote_path` | Download file |
| `upload_atomic` | `local_path`, `remote_path` | Upload via temp + atomic rename |
| `mkdir` | `remote_path` | Create remote directory |
| `remove` | `remote_path` | Remove remote file/directory |
| `chmod` | `remote_path`, `mode` | Change file mode (e.g. "0755") |
| `move` | `remote_path`, `target_path` | Move/rename remote file |
| `deploy` | `plan` or `plan_json` | Multi-step deployment |
| `sync` | `local_path`, `remote_path`, `direction` | File/directory synchronization |

## Deploy Plan Format

```json
{
  "steps": [
    {"name": "stop service", "type": "cmd", "cmd": "pkill app || true", "timeout": "10s"},
    {"name": "upload binary", "type": "upload_atomic", "local_path": "./app", "remote_path": "/opt/app/app", "temp_path": "/opt/app/app.tmp"},
    {"name": "make executable", "type": "chmod", "remote_path": "/opt/app/app", "mode": "0755"},
    {"name": "start service", "type": "cmd", "cmd": "cd /opt/app && ./app", "timeout": "15s"},
    {"name": "sync configs", "type": "sync", "local_path": "./config", "remote_path": "/opt/app/config", "direction": "push", "recursive": true},
    {"name": "optional step", "type": "cmd", "cmd": "echo done", "continue_on_error": true}
  ]
}
```

## Sync Conflict Policies

For bidirectional sync (`"direction": "bidirectional"`):

| Policy | Behavior |
|--------|----------|
| `fail_on_conflict` | Report conflict, skip file (default) |
| `newer_wins` | Upload or download whichever is newer |
| `local_wins` | Always upload (overwrite remote) |
| `remote_wins` | Always download (overwrite local) |
