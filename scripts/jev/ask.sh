#!/bin/sh
# Raw one-question probe against arbitrary text on stdin. Dev-only.
# Usage: echo "<state>" | scripts/jev/ask.sh noul "<instructions>"
set -eu
DIR="$(cd "$(dirname "$0")" && pwd)"
. "$DIR/lib.sh"
jev_load_key || { echo "jev: no key at $JEV_ENV_FILE (measurement-only mode)" >&2; exit 0; }
TYPE="${1:?usage: ask.sh <noul|choice|score> <instructions>}"
INSTR="${2:?missing instructions}"
STATE="$(cat)"
python3 - "$TYPE" "$INSTR" "$STATE" <<'PY'
import json
import os
import sys
import urllib.request

qtype, instr, state = sys.argv[1], sys.argv[2], sys.argv[3]
body = json.dumps({
    "model": os.environ.get("JEV_MODEL", "jev-latest"),
    "state": state,
    "questions": {"q": {"type": qtype, "instructions": instr}},
}).encode()
req = urllib.request.Request(
    os.environ.get("JEV_API", "https://api.typesafe.ai/v1/systemone"),
    data=body,
    headers={
        "Authorization": "Bearer " + os.environ["TYPESAFE_API_KEY"],
        "Content-Type": "application/json",
        "User-Agent": "moai-adk-jev/1.0",
    },
)
with urllib.request.urlopen(req, timeout=60) as resp:
    print(json.dumps(json.loads(resp.read()), ensure_ascii=False, indent=1))
PY
