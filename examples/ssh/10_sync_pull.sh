#!/bin/bash
# Example: Sync remote directory to local (pull)
#
# Demonstrates: SSH sync pull action
# Shows: basic pull, pull with delete

source "$(dirname "$0")/common.sh"

REMOTE_BASE="/tmp/aitk_examples/sync_push"

header "Example 10: Sync pull (remote -> local)"

LOCAL_DIR="/tmp/aitk_sync_pull_dst"
rm -rf "$LOCAL_DIR"
mkdir -p "$LOCAL_DIR"
echo "stale local file" > "$LOCAL_DIR/stale_file.txt"

echo "Local files before pull:"
find "$LOCAL_DIR" -type f | sort
echo ""

echo "1. Pull from remote:"
ssh_run_pretty "\"action\":\"sync\",\"direction\":\"pull\",\"local_path\":\"${LOCAL_DIR}\",\"remote_path\":\"${REMOTE_BASE}\",\"recursive\":true"

echo ""
echo "Local files after pull (stale_file.txt still exists - no delete):"
find "$LOCAL_DIR" -type f | sort
echo ""

echo "2. Pull with delete (removes stale_file.txt):"
ssh_run_pretty "\"action\":\"sync\",\"direction\":\"pull\",\"local_path\":\"${LOCAL_DIR}\",\"remote_path\":\"${REMOTE_BASE}\",\"recursive\":true,\"delete\":true"

echo ""
echo "Local files after pull+delete:"
find "$LOCAL_DIR" -type f | sort

rm -rf "$LOCAL_DIR"
