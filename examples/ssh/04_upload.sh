#!/bin/bash
# Example: Upload a file to remote server
#
# Demonstrates: SSH upload action, upload with custom filename

source "$(dirname "$0")/common.sh"

REMOTE_BASE="/tmp/aitk_examples"

header "Example 4: Upload file to remote server"

ssh_run "\"action\":\"mkdir\",\"remote_path\":\"${REMOTE_BASE}\""

LOCAL_FILE=$(make_test_file "/tmp/aitk_example_upload.txt" "Upload test - $(date)")

echo "Upload to directory (auto-named):"
ssh_run_pretty "\"action\":\"upload\",\"local_path\":\"${LOCAL_FILE}\",\"remote_path\":\"${REMOTE_BASE}/\""

echo ""
echo "Upload with custom filename:"
ssh_run_pretty "\"action\":\"upload\",\"local_path\":\"${LOCAL_FILE}\",\"remote_path\":\"${REMOTE_BASE}/\",\"file_name\":\"custom_name.txt\""

echo ""
echo "Verify uploaded files:"
ssh_run_pretty "\"action\":\"cmd\",\"cmd\":\"ls -la ${REMOTE_BASE}\""

rm -f "$LOCAL_FILE"
