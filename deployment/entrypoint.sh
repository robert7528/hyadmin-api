#!/bin/sh
set -e

echo "=== [migrate] Applying DB migrations ==="
./hyadmin-migrate admin

if [ "${RUN_SEED:-false}" = "true" ]; then
  echo "=== [seed] Running initial seed ==="
  ./hyadmin-seed
fi

echo "=== [serve] Starting API server ==="
exec ./hyadmin-api serve
