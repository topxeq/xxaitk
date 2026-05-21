#!/bin/bash
# Example: Execute a single remote command via SSH
#
# Demonstrates: SSH cmd action

source "$(dirname "$0")/common.sh"

header "Example 1: Execute a single remote command"

echo "Command: hostname"
ssh_run_pretty '"action":"cmd","cmd":"hostname"'

echo ""
echo "Command: uname -a"
ssh_run_pretty '"action":"cmd","cmd":"uname -a"'

echo ""
echo "Command: df -h / (check disk usage)"
ssh_run_pretty '"action":"cmd","cmd":"df -h /"'
