#!/bin/bash
# Example: Execute remote commands with timeout
#
# Demonstrates: SSH cmd action with cmd_timeout

source "$(dirname "$0")/common.sh"

header "Example 2: Command with timeout"

echo "Command that completes within timeout:"
ssh_run_pretty '"action":"cmd","cmd":"sleep 1 && echo done","cmd_timeout":"5s"'

echo ""
echo "Command that exceeds timeout (timed_out=true, exit_code=-1):"
ssh_run_pretty '"action":"cmd","cmd":"sleep 30","cmd_timeout":"2s"'
