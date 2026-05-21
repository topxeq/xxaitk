#!/bin/bash
# Example: Download a file from remote server
#
# Demonstrates: SSH download action, download with custom filename

source "$(dirname "$0")/common.sh"

REMOTE_BASE="/tmp/aitk_examples"

header "Example 5: Download file from remote server"

echo "Download to specific local path:"
ssh_run_pretty "\"action\":\"download\",\"local_path\":\"/tmp/aitk_example_dl.txt\",\"remote_path\":\"${REMOTE_BASE}/aitk_example_upload.txt\""

echo ""
echo "Downloaded content:"
cat /tmp/aitk_example_dl.txt 2>/dev/null || echo "(file not found)"

echo ""
echo "Download with custom filename:"
mkdir -p /tmp/aitk_example_dl_dir
ssh_run_pretty "\"action\":\"download\",\"local_path\":\"/tmp/aitk_example_dl_dir/\",\"remote_path\":\"${REMOTE_BASE}/aitk_example_upload.txt\",\"file_name\":\"downloaded.txt\""

echo ""
echo "Downloaded with custom name:"
cat /tmp/aitk_example_dl_dir/downloaded.txt 2>/dev/null || echo "(file not found)"

rm -f /tmp/aitk_example_dl.txt
rm -rf /tmp/aitk_example_dl_dir
