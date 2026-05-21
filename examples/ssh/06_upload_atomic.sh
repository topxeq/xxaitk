#!/bin/bash
# Example: Atomic file upload (upload to temp path, then rename)
#
# Demonstrates: SSH upload_atomic action
# Use case: Deploying a binary without downtime - the file appears
# atomically at the target path only after the upload completes.

source "$(dirname "$0")/common.sh"

REMOTE_BASE="/tmp/aitk_examples"

header "Example 6: Atomic upload"

LOCAL_FILE=$(make_test_file "/tmp/aitk_example_atomic.bin" "Atomic binary content - $(date)")

echo "Upload atomically (temp path -> final path):"
ssh_run_pretty "\"action\":\"upload_atomic\",\"local_path\":\"${LOCAL_FILE}\",\"remote_path\":\"${REMOTE_BASE}/app.bin\",\"temp_path\":\"${REMOTE_BASE}/app.bin.tmp\""

echo ""
echo "Verify the file exists at final path (not temp path):"
ssh_run_pretty "\"action\":\"cmd\",\"cmd\":\"ls -la ${REMOTE_BASE}/app.bin\""

rm -f "$LOCAL_FILE"
