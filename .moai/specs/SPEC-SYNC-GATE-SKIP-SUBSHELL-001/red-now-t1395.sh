#!/bin/bash
# red-now.sh — observational RED for card t1395 (sync gate skipped-tool pipe-subshell loss).
# Single invocation against the PRE-implementation tree. Observes, does not fix:
#   cell 1: gate stdout is EMPTY when the csharp checker (dotnet) is absent
#   cell 2: audit log line carries `skipped_tools=` with an empty value
#   cell 3: RECORD_FILE records `pass` (the pass is cached for later turns)
# Positive controls (guard the zero-hits from reading as measurement failure):
#   pc1: `command -v dotnet` FAILS inside the scrubbed env (the absent-tool branch is reachable)
#   pc2: the gate reached the csharp case (audit carries `language=csharp`)
#   pc3: the audit line DOES carry the literal token `skipped_tools=` (so the empty
#        value is a real observation, not a missing line)
set -u
GATE="$1"                       # absolute path to the gate script under test
EV="$2"                         # evidence dir
FIX="$(mktemp -d /tmp/t1395-red.XXXXXX)"
SCRUB_LOG="$FIX/scrub.log"

# --- scrubbed env: /usr/bin:/bin only (no /opt/homebrew/bin => no dotnet) ---
PATH="/usr/bin:/bin"
export PATH

for t in git bash find head grep awk sed stat date shasum sort tr cut cat mkdir rm tail cksum; do
    command -v "$t" >/dev/null 2>&1 || { echo "MISSING TOOL IN SCRUB: $t" | tee "$SCRUB_LOG"; exit 3; }
done
if command -v dotnet >/dev/null 2>&1; then
    echo "SCRUB FAILED: dotnet still resolvable: $(command -v dotnet)" | tee "$SCRUB_LOG"; exit 3
fi
printf 'pc1 PASS: dotnet unresolvable in scrubbed env\n' | tee "$SCRUB_LOG"

# --- C# fixture (mirrors internal/template/hook_gate_exit_status_test.go:54) ---
printf '<Project/>\n' > "$FIX/f.csproj"
printf 'class A {}\n' > "$FIX/a.cs"
git -C "$FIX" init --quiet
git -C "$FIX" -c user.email=red@example.invalid -c user.name=red add .
git -C "$FIX" -c user.email=red@example.invalid -c user.name=red commit --quiet -m "docs: sync-phase fixture"

# --- run the gate once, dotnet absent, blocking mode, stdin closed ---
mkdir -p "$EV"
OUT=$(cd "$FIX" && CLAUDE_PROJECT_DIR="$FIX" MOAI_SYNC_GATE_BLOCKING=1 MOAI_AUTONOMY_TIER= \
    bash "$GATE" </dev/null 2>"$FIX/stderr")
RC=$?
printf '%s' "$OUT" > "$EV/red-stdout.txt"
cp "$FIX/stderr" "$EV/red-stderr.txt"
AUDIT="$FIX/.moai/logs/sync-quality-gate.log"
[ -f "$AUDIT" ] && cp "$AUDIT" "$EV/red-audit.log"
RECORD="$FIX/.moai/state/sync-quality-gate.last"
[ -f "$RECORD" ] && cp "$RECORD" "$EV/red-record.txt"

# --- the three defect cells ---
echo "=== gate exit code: $RC ==="
if [ -z "$OUT" ]; then echo "CELL1 CONFIRMED: stdout EMPTY"; else echo "CELL1 NOT OBSERVED: stdout=$(head -c 200 <<<"$OUT")"; fi
if grep -q 'skipped_tools= ' "$EV/red-audit.log" 2>/dev/null || grep -q 'skipped_tools=$' "$EV/red-audit.log" 2>/dev/null; then
    echo "CELL2 CONFIRMED: audit skipped_tools empty:"; grep -o 'skipped_tools=[^ ]*' "$EV/red-audit.log" | head -1
else
    echo "CELL2 NOT OBSERVED"
fi
if [ -f "$EV/red-record.txt" ] && grep -q ' pass ' "$EV/red-record.txt"; then
    echo "CELL3 CONFIRMED: pass record written: $(cat "$EV/red-record.txt")"
else
    echo "CELL3 NOT OBSERVED: record=$(cat "$EV/red-record.txt" 2>/dev/null || echo MISSING)"
fi

# --- positive controls ---
if grep -q 'language=csharp' "$EV/red-audit.log" 2>/dev/null; then
    echo "PC2 PASS: gate reached the csharp case"
else
    echo "PC2 FAIL: gate never reached csharp — measurement invalid"
fi
if grep -q 'skipped_tools=' "$EV/red-audit.log" 2>/dev/null; then
    echo "PC3 PASS: audit line carries the skipped_tools= token"
else
    echo "PC3 FAIL: no skipped_tools token at all — measurement invalid"
fi
rm -rf "$FIX"
