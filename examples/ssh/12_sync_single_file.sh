#!/bin/bash
# Example: Single file sync (not a directory)
#
# Demonstrates: SSH sync with single files (push and pull)

source "$(dirname "$0")/common.sh"

REMOTE_BASE="/tmp/aitk_examples/single_file_sync"

header "Example 12: Single file sync"

ssh_run "\"action\":\"mkdir\",\"remote_path\":\"${REMOTE_BASE}\""

echo "1. Push single file to remote:"
make_test_file "/tmp/aitk_single_src.txt" "Single file sync content - $(date)"
ssh_run_pretty "\"action\":\"sync\",\"direction\":\"push\",\"local_path\":\"/tmp/aitk_single_src.txt\",\"remote_path\":\"${REMOTE_BASE}/synced_file.txt\""

echo ""
echo "Verify on remote:"
ssh_run_pretty "\"action\":\"cmd\",\"cmd\":\"cat ${REMOTE_BASE}/synced_file.txt\""

echo ""
echo "2. Pull single file from remote:"
rm -f /tmp/aitk_single_dst.txt
ssh_run_pretty "\"action\":\"sync\",\"direction\":\"pull\",\"local_path\":\"/tmp/aitk_single_dst.txt\",\"remote_path\":\"${REMOTE_BASE}/synced_file.txt\""

echo "Pulled content:"
cat /tmp/aitk_single_dst.txt 2>/dev/null || echo "(not found)"

rm -f /tmp/aitk_single_src.txt /tmp/aitk_single_dst.txt
