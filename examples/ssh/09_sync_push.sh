#!/bin/bash
# Example: Sync local directory to remote (push)
#
# Demonstrates: SSH sync push action
# Shows: basic push, dry-run, include/exclude, delete

source "$(dirname "$0")/common.sh"

REMOTE_BASE="/tmp/aitk_examples/sync_push"

header "Example 9: Sync push (local -> remote)"

LOCAL_DIR="/tmp/aitk_sync_push_src"
rm -rf "$LOCAL_DIR"
mkdir -p "$LOCAL_DIR/css" "$LOCAL_DIR/js"
echo "Main page" > "$LOCAL_DIR/index.html"
echo "body { color: red; }" > "$LOCAL_DIR/css/style.css"
echo "console.log('app');" > "$LOCAL_DIR/js/app.js"
echo "//# sourceMappingURL=app.js.map" > "$LOCAL_DIR/js/app.js.map"
echo "Secret key: abc123" > "$LOCAL_DIR/secret.key"

echo "Local files:"
find "$LOCAL_DIR" -type f | sort
echo ""

echo "1. Push with include/exclude (only js files, exclude .map):"
ssh_run_pretty "\"action\":\"sync\",\"direction\":\"push\",\"local_path\":\"${LOCAL_DIR}\",\"remote_path\":\"${REMOTE_BASE}\",\"recursive\":true,\"include\":[\"js/**\"],\"exclude\":[\"**/*.map\"]"

echo ""
echo "Verify - only app.js should be on remote (no .map, no css, no html):"
ssh_run_pretty "\"action\":\"cmd\",\"cmd\":\"find ${REMOTE_BASE} -type f | sort\""

echo ""
echo "2. Push all files with delete (removes previously synced app.js):"
ssh_run_pretty "\"action\":\"sync\",\"direction\":\"push\",\"local_path\":\"${LOCAL_DIR}\",\"remote_path\":\"${REMOTE_BASE}\",\"recursive\":true,\"delete\":true"

echo ""
echo "Verify - should now match local:"
ssh_run_pretty "\"action\":\"cmd\",\"cmd\":\"find ${REMOTE_BASE} -type f | sort\""

echo ""
echo "3. Dry-run push (no actual changes):"
ssh_run_pretty "\"action\":\"sync\",\"direction\":\"push\",\"local_path\":\"${LOCAL_DIR}\",\"remote_path\":\"${REMOTE_BASE}\",\"recursive\":true,\"dry_run\":true"

rm -rf "$LOCAL_DIR"
