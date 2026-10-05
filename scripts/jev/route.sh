#!/bin/sh
# Route a lane question to a decision class. Dev-only.
# Usage: scripts/jev/route.sh < question.txt   |   echo "..." | scripts/jev/route.sh
set -eu
DIR="$(cd "$(dirname "$0")" && pwd)"
exec python3 "$DIR/route.py"
