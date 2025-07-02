#!/bin/sh
set -e
echo "[start.sh] Starting chronyd daemon..."
chronyd -d -s &
# Allow chrony a moment to initialize and perform initial sync
sleep 5
echo "[start.sh] Starting ntp-proxy Go application..."
exec ntp-proxy