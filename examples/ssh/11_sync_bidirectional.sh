#!/bin/bash
# Example: Bidirectional sync with conflict resolution
#
# Demonstrates: SSH sync bidirectional with different conflict policies
# Shows: local_wins, newer_wins, fail_on_conflict

source "$(dirname "$0")/common.sh"

REMOTE_BASE="/tmp/aitk_examples/sync_bidi"

header "Example 11: Bidirectional sync with conflict resolution"

# --- Scenario 1: local_wins ---
echo "=== Scenario 1: local_wins ==="
LOCAL_DIR="/tmp/aitk_bidi_local1"
rm -rf "$LOCAL_DIR"
mkdir -p "$LOCAL_DIR"
echo "local content" > "$LOCAL_DIR/both_changed.txt"
echo "local only" > "$LOCAL_DIR/local_only.txt"

ssh_run "\"action\":\"cmd\",\"cmd\":\"rm -rf ${REMOTE_BASE} && mkdir -p ${REMOTE_BASE} && echo remote_content > ${REMOTE_BASE}/both_changed.txt && echo remote_only > ${REMOTE_BASE}/remote_only.txt\""

echo "Before sync:"
echo "  Local:  both_changed.txt='local content', local_only.txt='local only'"
echo "  Remote: both_changed.txt='remote content', remote_only.txt='remote only'"
echo ""

ssh_run_pretty "\"action\":\"sync\",\"direction\":\"bidirectional\",\"local_path\":\"${LOCAL_DIR}\",\"remote_path\":\"${REMOTE_BASE}\",\"recursive\":true,\"conflict\":\"local_wins\""

echo "After sync (local_wins means local overwrites remote for conflicts):"
echo "  Local both_changed.txt: $(cat "$LOCAL_DIR/both_changed.txt")"
echo "  Remote both_changed.txt:"
ssh_run "\"action\":\"cmd\",\"cmd\":\"cat ${REMOTE_BASE}/both_changed.txt\"" | python3 -c "import sys,json; d=json.load(sys.stdin); print('  ' + d['data']['results'][0]['output'].strip()) if d.get('ok') else print('  ERROR')" 2>/dev/null

echo ""

# --- Scenario 2: fail_on_conflict ---
echo "=== Scenario 2: fail_on_conflict ==="
LOCAL_DIR="/tmp/aitk_bidi_local2"
rm -rf "$LOCAL_DIR"
mkdir -p "$LOCAL_DIR"
echo "local version" > "$LOCAL_DIR/conflict.txt"

ssh_run "\"action\":\"cmd\",\"cmd\":\"rm -rf ${REMOTE_BASE} && mkdir -p ${REMOTE_BASE} && echo remote_version > ${REMOTE_BASE}/conflict.txt\""

ssh_run_pretty "\"action\":\"sync\",\"direction\":\"bidirectional\",\"local_path\":\"${LOCAL_DIR}\",\"remote_path\":\"${REMOTE_BASE}\",\"recursive\":true,\"conflict\":\"fail_on_conflict\""

echo ""

# --- Scenario 3: newer_wins ---
echo "=== Scenario 3: newer_wins ==="
LOCAL_DIR="/tmp/aitk_bidi_local3"
rm -rf "$LOCAL_DIR"
mkdir -p "$LOCAL_DIR"
echo "newer local" > "$LOCAL_DIR/timed.txt"
touch -t 202701010000 "$LOCAL_DIR/timed.txt"

ssh_run "\"action\":\"cmd\",\"cmd\":\"rm -rf ${REMOTE_BASE} && mkdir -p ${REMOTE_BASE} && echo older_remote > ${REMOTE_BASE}/timed.txt\""

echo "Local mtime set to 2027-01-01 (future), remote is current time."
echo "With newer_wins, local should win (it's newer):"
echo ""

ssh_run_pretty "\"action\":\"sync\",\"direction\":\"bidirectional\",\"local_path\":\"${LOCAL_DIR}\",\"remote_path\":\"${REMOTE_BASE}\",\"recursive\":true,\"conflict\":\"newer_wins\""

echo ""
echo "Remote content after sync (should be 'newer local'):"
ssh_run "\"action\":\"cmd\",\"cmd\":\"cat ${REMOTE_BASE}/timed.txt\"" | python3 -c "import sys,json; d=json.load(sys.stdin); print('  ' + d['data']['results'][0]['output'].strip()) if d.get('ok') else print('  ERROR')" 2>/dev/null

rm -rf /tmp/aitk_bidi_local1 /tmp/aitk_bidi_local2 /tmp/aitk_bidi_local3
