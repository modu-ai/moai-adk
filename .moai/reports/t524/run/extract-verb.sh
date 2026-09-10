#!/usr/bin/env bash
# Print the first ```bash block under the "### Group 4:" heading of an agent doc (card t524).
# Mirrors the extraction the Go test performs (auditVerb), so the evidence runs
# the committed verb rather than a pasted copy.
# usage: extract-verb.sh <plan-auditor.md>
awk '
  /^### Group 4:/ { insec = 1; next }
  insec && !inblk && $0 == "```bash" { inblk = 1; next }
  inblk && $0 == "```" { exit }
  inblk { print }
' "$1"
