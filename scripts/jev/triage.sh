#!/bin/sh
# Local-only card premise triage. Dev-only; not distributed.
# Usage: scripts/jev/triage.sh t784 [t787 ...]
set -eu
DIR="$(cd "$(dirname "$0")" && pwd)"
exec python3 "$DIR/triage.py" "$@"
