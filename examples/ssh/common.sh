#!/bin/bash
# Common helper functions for aitk SSH examples
# Source this file: source "$(dirname "$0")/common.sh"

AITK="${AITK:-aitk}"

# Default SSH connection parameters - edit these for your environment
SSH_HOST="${SSH_HOST:-YOUR_SERVER_IP}"
SSH_PORT="${SSH_PORT:-22}"
SSH_USER="${SSH_USER:-root}"
SSH_PASS="${SSH_PASS:-YOUR_PASSWORD}"
# SSH_KEY="${SSH_KEY:-~/.ssh/id_rsa"

# Run an SSH action via aitk
# Usage:
#   ssh_run '"action":"cmd","cmd":"hostname"'
#   (JSON fragment - auto-wrapped with connection params and braces)
#
#   ssh_run '{"host":"1.2.3.4","user":"root","password":"x","action":"cmd","cmd":"hostname"}'
#   (complete JSON object - used as-is)
ssh_run() {
    local payload="$1"
    local json

    if [[ "$payload" == "{"* ]]; then
        json="$payload"
    else
        json="{\"host\":\"${SSH_HOST}\",\"port\":${SSH_PORT},\"user\":\"${SSH_USER}\",\"password\":\"${SSH_PASS}\",${payload}}"
    fi

    local hex
    hex=$(echo -n "$json" | xxd -p | tr -d '\n')
    $AITK "SSH_$hex"
}

# Pretty-print JSON output
ssh_run_pretty() {
    local result
    result=$(ssh_run "$@")
    echo "$result" | python3 -m json.tool 2>/dev/null || echo "$result"
}

# Create a temporary local test file
make_test_file() {
    local path="${1:-/tmp/aitk_example_test.txt}"
    local content="${2:-Hello from aitk SSH example - $(date)}"
    echo "$content" > "$path"
    echo "$path"
}

# Print section header
header() {
    echo ""
    echo "============================================================"
    echo "  $1"
    echo "============================================================"
}
