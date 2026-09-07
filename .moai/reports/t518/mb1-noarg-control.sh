#!/bin/bash
# t518 M-B1 — REQ-SLI-001's SHALL NOT: a no-argument invocation must be
# unchanged, because it runs on detectBaseDir(cwd) and the resolver never
# touches it.
#
# "The early return makes it unchanged" is an argument, not a measurement, so
# two binaries are built — one from the pre-change sources at HEAD, one from
# the working tree — and the same corpus scan is run through both. Exit codes
# are read on their own line; no pipe.
set -u
WT=/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t518
cd "$WT" || exit 1
OUT="$WT/.moai/reports/t518/mb1-noarg-control.txt"
BK=/tmp/t518-noarg-bk
mkdir -p "$BK"

cp internal/cli/spec_lint.go "$BK/after-spec_lint.go"
cp internal/cli/specid/specid.go "$BK/after-specid.go"

git show HEAD:internal/cli/spec_lint.go > "$BK/before-spec_lint.go"
git show HEAD:internal/cli/specid/specid.go > "$BK/before-specid.go"

{
  echo "tree: $WT"
  echo "HEAD (pre-change sources): $(git rev-parse HEAD)"
  echo
} > "$OUT"

# BEFORE
cp "$BK/before-spec_lint.go" internal/cli/spec_lint.go
cp "$BK/before-specid.go" internal/cli/specid/specid.go
go build -o /tmp/t518-moai-before ./cmd/moai
echo "before build rc=$?" >> "$OUT"

# AFTER (restore the working tree immediately, so nothing is left swapped)
cp "$BK/after-spec_lint.go" internal/cli/spec_lint.go
cp "$BK/after-specid.go" internal/cli/specid/specid.go
go build -o /tmp/t518-moai-after ./cmd/moai
echo "after build rc=$?" >> "$OUT"

/tmp/t518-moai-before spec lint > /tmp/t518-noarg-before.txt 2>&1
echo "BEFORE: moai spec lint (no argument) rc=$?" >> "$OUT"
/tmp/t518-moai-after spec lint > /tmp/t518-noarg-after.txt 2>&1
echo "AFTER:  moai spec lint (no argument) rc=$?" >> "$OUT"

{
  echo
  echo "--- BEFORE summary line ---"
  tail -2 /tmp/t518-noarg-before.txt
  echo "--- AFTER summary line ---"
  tail -2 /tmp/t518-noarg-after.txt
  echo
  echo "--- byte-level diff of the two full outputs ---"
  diff /tmp/t518-noarg-before.txt /tmp/t518-noarg-after.txt
  echo "diff rc=$?  (0 = identical)"
  echo
  echo "--- line counts (a non-empty diff would be visible above; these are a second reading) ---"
  wc -l /tmp/t518-noarg-before.txt /tmp/t518-noarg-after.txt
} >> "$OUT"

echo DONE
