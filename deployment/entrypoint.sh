#!/bin/sh
set -e

echo "=== [migrate] Applying DB migrations ==="
./hyadmin-migrate admin

echo "=== [serve] Starting API server ==="
exec ./hyadmin-api serve
