#!/bin/bash
# Example: Remote directory and file operations
#
# Demonstrates: SSH mkdir, chmod, move, remove actions

source "$(dirname "$0")/common.sh"

REMOTE_BASE="/tmp/aitk_examples"

header "Example 7: Remote directory and file operations"

echo "1. Create nested directory:"
ssh_run_pretty "\"action\":\"mkdir\",\"remote_path\":\"${REMOTE_BASE}/data/logs\""

echo ""
echo "2. Upload a file for operations:"
make_test_file "/tmp/aitk_ex7.txt" "test file for ops"
ssh_run_pretty "\"action\":\"upload\",\"local_path\":\"/tmp/aitk_ex7.txt\",\"remote_path\":\"${REMOTE_BASE}/data/script.sh\""

echo ""
echo "3. Chmod the file to 0755:"
ssh_run_pretty "\"action\":\"chmod\",\"remote_path\":\"${REMOTE_BASE}/data/script.sh\",\"mode\":\"0755\""

echo ""
echo "4. Verify chmod:"
ssh_run_pretty "\"action\":\"cmd\",\"cmd\":\"ls -la ${REMOTE_BASE}/data/script.sh\""

echo ""
echo "5. Move (rename) the file:"
ssh_run_pretty "\"action\":\"move\",\"remote_path\":\"${REMOTE_BASE}/data/script.sh\",\"target_path\":\"${REMOTE_BASE}/data/app.sh\""

echo ""
echo "6. Verify move:"
ssh_run_pretty "\"action\":\"cmd\",\"cmd\":\"ls -la ${REMOTE_BASE}/data/\""

echo ""
echo "7. Remove a file:"
ssh_run_pretty "\"action\":\"remove\",\"remote_path\":\"${REMOTE_BASE}/data/app.sh\""

echo ""
echo "8. Remove a directory recursively:"
ssh_run_pretty "\"action\":\"remove\",\"remote_path\":\"${REMOTE_BASE}/data\""

echo ""
echo "9. Remove non-existent path (graceful - no error):"
ssh_run_pretty "\"action\":\"remove\",\"remote_path\":\"${REMOTE_BASE}/nonexistent_xyz\""

rm -f /tmp/aitk_ex7.txt
