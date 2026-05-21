#!/bin/bash
# Example: Execute multiple remote commands from a file
#
# Demonstrates: SSH cmd action with cmd_file

source "$(dirname "$0")/common.sh"

header "Example 3: Execute commands from a file"

CMDFILE="/tmp/aitk_example_commands.txt"
cat > "$CMDFILE" << 'EOF'
echo "=== System Info ==="
uname -s
echo "=== Memory ==="
free -h | head -2
echo "=== Uptime ==="
uptime
EOF

echo "Command file contents:"
cat "$CMDFILE"
echo ""

ssh_run_pretty "\"action\":\"cmd\",\"cmd_file\":\"${CMDFILE}\""

rm -f "$CMDFILE"
