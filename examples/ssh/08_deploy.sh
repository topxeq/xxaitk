#!/bin/bash
# Example: Multi-step deployment plan
#
# Demonstrates: SSH deploy action with plan file and plan_json
# Shows: cmd, mkdir, upload, upload_atomic, chmod steps
# Also demonstrates continue_on_error behavior

source "$(dirname "$0")/common.sh"

REMOTE_BASE="/tmp/aitk_examples/deploy_demo"

header "Example 8: Deployment plan"

PLANFILE="/tmp/aitk_deploy_plan.json"
cat > "$PLANFILE" << PLANEOF
{
  "steps": [
    {"name": "prepare directories", "type": "mkdir", "remote_path": "${REMOTE_BASE}/bin"},
    {"name": "check system", "type": "cmd", "cmd": "uname -s && echo System ready"},
    {"name": "upload binary atomically", "type": "upload_atomic", "local_path": "/tmp/aitk_deploy_demo_app.bin", "remote_path": "${REMOTE_BASE}/bin/app", "temp_path": "${REMOTE_BASE}/bin/app.tmp"},
    {"name": "make executable", "type": "chmod", "remote_path": "${REMOTE_BASE}/bin/app", "mode": "0755"},
    {"name": "verify deployment", "type": "cmd", "cmd": "ls -la ${REMOTE_BASE}/bin/"}
  ]
}
PLANEOF

make_test_file "/tmp/aitk_deploy_demo_app.bin" "#!/bin/bash\necho 'My App v1.0'"

echo "Deploy plan:"
cat "$PLANFILE" | python3 -m json.tool 2>/dev/null || cat "$PLANFILE"
echo ""

echo "Running deployment from plan file:"
ssh_run_pretty "\"action\":\"deploy\",\"plan\":\"${PLANFILE}\""

echo ""
echo "--- Inline plan_json with continue_on_error ---"
echo ""

PLAN_JSON='{"steps":[{"name":"inline step","type":"cmd","cmd":"echo Hello from inline deploy"},{"name":"continue on error demo","type":"cmd","cmd":"ls /nonexistent_path","continue_on_error":true},{"name":"runs despite error","type":"cmd","cmd":"echo This step ran because continue_on_error was true"}]}'
PLAN_ESCAPED=$(echo "$PLAN_JSON" | python3 -c "import sys,json; print(json.dumps(sys.stdin.read().strip()))")

echo "Running inline deploy plan with continue_on_error:"
ssh_run "{\"host\":\"${SSH_HOST}\",\"port\":${SSH_PORT},\"user\":\"${SSH_USER}\",\"password\":\"${SSH_PASS}\",\"action\":\"deploy\",\"plan_json\":${PLAN_ESCAPED}}"

rm -f "$PLANFILE" /tmp/aitk_deploy_demo_app.bin
