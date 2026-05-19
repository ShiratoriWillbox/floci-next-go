#!/bin/sh
set -e
export BACK_ADDR="${BACK_ADDR:-:8080}"
/app/server &
cd /app/front
exec node node_modules/next/dist/bin/next start -H 0.0.0.0 -p "${PORT:-3000}"
