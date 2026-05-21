#!/bin/bash
# Example: Private key authentication and error handling
#
# Demonstrates: SSH with key-based auth, various error scenarios

source "$(dirname "$0")/common.sh"

header "Example 13: Key-based auth and error handling"

echo "1. Key-based authentication example:"
echo "   To use key auth instead of password, pass key instead of password:"
echo ""
echo '   ssh_run '\''{"host":"1.2.3.4","port":22,"user":"root","key":"~/.ssh/id_rsa","action":"cmd","cmd":"hostname"}'\'''
echo ""
echo "   With passphrase:"
echo ""
echo '   ssh_run '\''{"host":"1.2.3.4","port":22,"user":"root","key":"~/.ssh/id_rsa","key_passphrase":"mypass","action":"cmd","cmd":"hostname"}'\'''
echo ""

echo "2. Error: missing host (complete JSON)"
ssh_run_pretty '{"user":"root","password":"x","action":"cmd","cmd":"echo hi"}'

echo ""
echo "3. Error: missing auth (no password or key)"
ssh_run_pretty '{"host":"1.2.3.4","user":"root","action":"cmd","cmd":"echo hi"}'

echo ""
echo "4. Error: unknown action"
ssh_run_pretty '{"host":"1.2.3.4","user":"root","password":"x","action":"invalid_action"}'

echo ""
echo "5. Error: bidirectional sync with delete (not allowed)"
ssh_run_pretty '{"host":"1.2.3.4","user":"root","password":"x","action":"sync","direction":"bidirectional","local_path":"/tmp/x","remote_path":"/tmp/x","recursive":true,"delete":true}'

echo ""
echo "6. Command exits with non-zero code (captured in result, not an error)"
ssh_run_pretty '"action":"cmd","cmd":"ls /nonexistent_directory_xyz"'
