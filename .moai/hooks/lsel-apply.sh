#!/usr/bin/env bash
# lsel-apply.sh — SPEC-LSEL-LOCAL-EVOLUTION-001 M3 playback-only APPLY consumer.
#
# Reads an already-approved decision.json (path passed as $1) and plays back the
# associated diff.patch onto the working tree. The script consumes a queue of
# approved decisions; it does not invent work. It:
#   1. Validates the target against the frozen allowlist (REQ-LSEL-001 / AC-001).
#   2. Mechanically refuses execution-meta targets that lack a synchronous-approval
#      marker in decision.json (REQ-LSEL-002 / AC-005 D3 self-amending-handcuffs).
#   3. Applies diff.patch via `git apply`, appends an apply-ledger.jsonl row, and
#      commits the change as a single lsel-<proposal-id>-tagged Conventional
#      Commit on the current (feature) branch (REQ-LSEL-004 / AC-013).
#
# No-arg invocation is a clean no-op (returns exit 0) so an empty approved-queue
# does not derail a loop pass.
#
# Exit codes: 0 = applied or no-op; 2 = frozen-path refuse; 3 = execution-meta
# refuse (no approval marker); 1 = other error.
set -euo pipefail

DECISION="${1:-}"

resolve_root() {
  # Prefer an explicit override, then $CLAUDE_PROJECT_DIR, then the git toplevel.
  if [[ -n "${LSEL_PROJECT_ROOT:-}" ]]; then echo "$LSEL_PROJECT_ROOT"; return; fi
  if [[ -n "${CLAUDE_PROJECT_DIR:-}" && -d "${CLAUDE_PROJECT_DIR:-}" ]]; then echo "$CLAUDE_PROJECT_DIR"; return; fi
  git rev-parse --show-toplevel 2>/dev/null || pwd
}
PROJECT_ROOT="$(resolve_root)"
ALLOWLIST="$PROJECT_ROOT/.claude/lsel/frozen-allowlist.json"
REJECT_LOG="$PROJECT_ROOT/.moai/logs/lsel-reject.log"
LEDGER="$PROJECT_ROOT/.moai/state/lsel/apply-ledger.jsonl"

# --- no-arg / empty-queue no-op ---
if [[ -z "$DECISION" ]]; then
  echo "lsel-apply: no decision.json argument — no-op (playback-only)" >&2
  exit 0
fi
if [[ ! -f "$DECISION" ]]; then
  echo "lsel-apply: decision not found: $DECISION" >&2
  exit 1
fi
if [[ ! -f "$ALLOWLIST" ]]; then
  echo "lsel-apply: frozen allowlist not found at $ALLOWLIST" >&2
  exit 1
fi

mkdir -p "$(dirname "$REJECT_LOG")" "$(dirname "$LEDGER")"

# --- portable field extraction (no jq dependency) ---
json_str_field() {
  # $1 key, $2 file → first string value
  grep -m1 "\"$1\"" "$2" 2>/dev/null \
    | sed 's/.*"'"$1"'"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' \
    | head -1
}
has_sync_approval() {
  # synchronous-approval marker = both the key and an approved decision present
  grep -q '"synchronous_approval"' "$1" && \
    grep -q '"decision"[[:space:]]*:[[:space:]]*"approved"' "$1"
}

PROPOSAL_ID="$(json_str_field proposal_id "$DECISION")"
TARGET="$(json_str_field target_surface "$DECISION")"
PROPOSAL_DIR="$(dirname "$DECISION")"
DIFF_PATCH="$PROPOSAL_DIR/diff.patch"
TS="$(date -u +%FT%TZ)"

if [[ -z "$PROPOSAL_ID" || -z "$TARGET" ]]; then
  echo "lsel-apply: decision.json missing proposal_id or target_surface" >&2
  exit 1
fi

# --- (1) target normalization + frozen-allowlist hard-reject (REQ-LSEL-001 / AC-LSEL-001) ---
# Normalize $TARGET (resolve '..' traversal + relative segments) BEFORE matching.
# A target like `memory/../internal/template/templates/x` collapses to a frozen
# path and MUST be refused — without normalization the anchored `^...` patterns
# are bypassed because the traversal prefix does not start with the frozen stem.
TARGET_NORM="$(python3 -c 'import os,sys; print(os.path.normpath(sys.argv[1]))' "$TARGET")"

frozen_match() {
  # $1 = already-normalized path to test
  python3 - "$ALLOWLIST" "$1" <<'PY'
import json, re, sys
allow, target = sys.argv[1], sys.argv[2]
try:
    data = json.load(open(allow))
except Exception:
    sys.exit(2)
for pat in data.get("frozen_patterns", []):
    try:
        # Anchored match: frozen_patterns are written as ^... regexes. re.match
        # anchors at start; re.search is NOT used (it let a frozen substring slip).
        if re.match(pat, target):
            sys.exit(0)
    except re.error:
        continue
sys.exit(1)
PY
}
if frozen_match "$TARGET_NORM"; then
  echo "$TS proposal=$PROPOSAL_ID target=$TARGET category=frozen-path reason=allowlist-hard-reject" >> "$REJECT_LOG"
  echo "lsel-apply: REFUSED (frozen-path) target=$TARGET proposal=$PROPOSAL_ID" >&2
  exit 2
fi

# --- (2) execution-meta forced-gate (REQ-LSEL-002 / AC-LSEL-005 D3) ---
exec_meta_match() {
  # $1 = already-normalized path to test
  python3 - "$ALLOWLIST" "$1" <<'PY'
import json, os, sys
allow, target = sys.argv[1], sys.argv[2]
try:
    data = json.load(open(allow))
except Exception:
    sys.exit(2)
norm = os.path.normpath(target)
for pat in data.get("execution_meta", []):
    p = os.path.normpath(pat)
    # treat execution_meta entries as path prefixes / substring matches
    if norm == p or norm.startswith(p + os.sep) or norm.startswith(p) or p in norm:
        sys.exit(0)
sys.exit(1)
PY
}
if exec_meta_match "$TARGET_NORM"; then
  if ! has_sync_approval "$DECISION"; then
    echo "$TS proposal=$PROPOSAL_ID target=$TARGET category=execution-meta reason=no-synchronous-approval-marker" >> "$REJECT_LOG"
    echo "lsel-apply: REFUSED (execution-meta, no synchronous-approval marker) target=$TARGET proposal=$PROPOSAL_ID" >&2
    exit 3
  fi
  CATEGORY="execution-meta"
else
  CATEGORY="routine"
fi

# --- (3) PRE-APPLY patch-path validation gate (REQ-LSEL-001 F1 fix) ---
# The central safety claim: EVERY write the loop performs is confined to the 6
# evolvable surfaces. Validating only $TARGET is insufficient — the patch's own
# embedded destination paths (the +++ b/<path> lines, which is where git apply
# actually writes) were never checked, so a divergent patch could write a frozen
# path while the declared target_surface stayed evolvable. This gate extracts
# every destination path from diff.patch, normalizes it, and runs it through the
# SAME frozen + execution-meta matchers — BEFORE `git apply`, so a refusal
# performs NO write at all.
if [[ ! -f "$DIFF_PATCH" ]]; then
  echo "lsel-apply: no diff.patch alongside decision.json at $DIFF_PATCH" >&2
  exit 1
fi

# Validate every patch destination path through BOTH matchers, deterministically.
# Exits 0 (all paths clear) or non-zero (at least one path refused). A refusal
# is reported via a structured verdict string on stdout for the caller to log.
validate_patch_paths() {
  python3 - "$ALLOWLIST" "$DIFF_PATCH" "$DECISION" <<'PY'
import json, os, re, sys
allow, patch, decision = sys.argv[1], sys.argv[2], sys.argv[3]
try:
    data = json.load(open(allow))
except Exception:
    sys.exit(2)  # allowlist unreadable → fail-stop (not a refuse-with-write)
frozen = data.get("frozen_patterns", [])
exec_meta = [os.path.normpath(p) for p in data.get("execution_meta", [])]

# synchronous-approval marker: decision.json carries a synchronous_approval
# object whose "decision" == "approved". Present → exec-meta targets allowed.
has_sync = False
try:
    d = json.load(open(decision))
    sa = d.get("synchronous_approval")
    if isinstance(sa, dict) and sa.get("decision") == "approved":
        has_sync = True
except Exception:
    has_sync = False

# Extract destination paths from BOTH the `diff --git a/X b/Y` headers AND the
# `+++ b/Y` lines. `+++` is the authoritative destination (covers rename/copy
# patches and new-file patches alike); the header line is parsed too for
# robustness, but `+++` is canonical.
dests = []
seen = set()
try:
    with open(patch, errors="replace") as f:
        for line in f:
            if line.startswith("+++ b/"):
                p = line[len("+++ b/"):].rstrip("\n")
            elif line.startswith("diff --git "):
                m = re.match(r"diff --git a/(.*) b/(.*)$", line.rstrip("\n"))
                if m:
                    p = m.group(2)
                else:
                    continue
            else:
                continue
            norm = os.path.normpath(p)
            if norm and norm not in seen:
                seen.add(norm)
                dests.append(norm)
except FileNotFoundError:
    print("refuse:patch-unreadable")
    sys.exit(3)

if not dests:
    print("refuse:patch-no-paths")
    sys.exit(4)

for norm in dests:
    # frozen check — anchored match against the regex patterns
    for pat in frozen:
        try:
            if re.match(pat, norm):
                print(f"refuse:frozen-path:{norm}")
                sys.exit(2)
        except re.error:
            continue
    # execution-meta check — without sync marker, any exec-meta path is refused
    if not has_sync:
        for p in exec_meta:
            if norm == p or norm.startswith(p + os.sep) or norm.startswith(p) or p in norm:
                print(f"refuse:execution-meta:{norm}")
                sys.exit(5)

sys.exit(0)
PY
}

# Capture the gate verdict WITHOUT tripping `set -e` — the python deliberately
# exits non-zero on a refusal, and a bare `GATE_OUT="$(validate_patch_paths)"`
# would abort the whole script at that line before GATE_RC is recorded, so the
# reject-log row would never be written (the refusal fired but left no audit
# trail). The `|| true` list keeps the assignment's exit status at 0 while the
# real rc is re-read from the command's own non-zero status via the temporary
# unset, then restored.
set +e
GATE_OUT="$(validate_patch_paths)"
GATE_RC=$?
set -e
if [[ $GATE_RC -ne 0 ]]; then
  # Parse the refusal verdict into a category + offending path for the reject-log.
  case "$GATE_OUT" in
    refuse:frozen-path:*)
      OFFEND="${GATE_OUT#refuse:frozen-path:}"
      echo "$TS proposal=$PROPOSAL_ID target=$TARGET category=frozen-path reason=patch-path-diverges-from-target path=$OFFEND" >> "$REJECT_LOG"
      echo "lsel-apply: REFUSED (frozen-path in diff.patch) proposal=$PROPOSAL_ID target=$TARGET path=$OFFEND" >&2
      exit 2 ;;
    refuse:execution-meta:*)
      OFFEND="${GATE_OUT#refuse:execution-meta:}"
      echo "$TS proposal=$PROPOSAL_ID target=$TARGET category=execution-meta reason=patch-path-is-execution-meta-without-marker path=$OFFEND" >> "$REJECT_LOG"
      echo "lsel-apply: REFUSED (execution-meta in diff.patch, no synchronous-approval marker) proposal=$PROPOSAL_ID target=$TARGET path=$OFFEND" >&2
      exit 3 ;;
    refuse:patch-unreadable)
      echo "lsel-apply: diff.patch unreadable at $DIFF_PATCH" >&2
      exit 1 ;;
    refuse:patch-no-paths)
      echo "lsel-apply: diff.patch declared no destination paths" >&2
      exit 1 ;;
    *)
      echo "lsel-apply: patch-path validation failed (rc=$GATE_RC out=$GATE_OUT)" >&2
      exit 1 ;;
  esac
fi

# --- (3a) scope-consistency guard (SHOULD — defense-in-depth) ---
# Warn (not refuse) if any normalized patch path is not within the declared
# target_surface. The MUST is the allowlist validation above; this checks that
# declared intent matches actual write scope. A divergence that reaches an
# evolvable-but-out-of-target path is suspicious but not frozen — surfaced as a
# stderr warning so it is visible without blocking a legitimate multi-file
# evolvable edit.
SCOPE_WARN="$(python3 - "$DIFF_PATCH" "$TARGET_NORM" <<'PY'
import os, re, sys
patch, declared = sys.argv[1], sys.argv[2]
decl_norm = os.path.normpath(declared)
dests = []
seen = set()
try:
    with open(patch, errors="replace") as f:
        for line in f:
            if line.startswith("+++ b/"):
                p = line[len("+++ b/"):].rstrip("\n")
            elif line.startswith("diff --git "):
                m = re.match(r"diff --git a/(.*) b/(.*)$", line.rstrip("\n"))
                p = m.group(2) if m else None
            else:
                continue
            if p is None:
                continue
            norm = os.path.normpath(p)
            if norm and norm not in seen:
                seen.add(norm)
                dests.append(norm)
except Exception:
    pass
# A single-file target_surface scopes each patch path to itself or its subtree.
# Divergence = any patch path outside the declared target (neither equal, nor
# nested under it as a directory).
for norm in dests:
    if norm == decl_norm:
        continue
    if decl_norm and norm.startswith(decl_norm + os.sep):
        continue
    print(norm)
    break
PY
)"
if [[ -n "$SCOPE_WARN" ]]; then
  echo "lsel-apply: WARN scope-divergence (patch path '$SCOPE_WARN' outside declared target '$TARGET'; evolvable — proceeding)" >&2
fi

# --- (4) apply the diff.patch (playback of an already-approved decision) ---
cd "$PROJECT_ROOT"
if ! git apply --whitespace=nowarn "$DIFF_PATCH"; then
  echo "lsel-apply: git apply failed for $DIFF_PATCH" >&2
  exit 1
fi

# Stage ONLY the paths the patch touches (working-tree hygiene: never git add -A).
PATCH_PATHS="$(grep -E '^\+\+\+ b/' "$DIFF_PATCH" | sed 's|^+++ b/||')"
if [[ -z "$PATCH_PATHS" ]]; then
  echo "lsel-apply: diff.patch declared no target paths" >&2
  exit 1
fi
# shellcheck disable=SC2086
git add $PATCH_PATHS

# --- (4) ledger row (AC-LSEL-013) — placeholder SHA backfilled after commit ---
printf '{"proposal_id":"%s","target_surface":"%s","ts":"%s","result":"applied","commit_sha":"pending-backfill","category":"%s"}\n' \
  "$PROPOSAL_ID" "$TARGET" "$TS" "$CATEGORY" >> "$LEDGER"

# --- (5) commit as one lsel-<proposal-id>-tagged Conventional Commit (REQ-LSEL-004) ---
if ! git commit -q -m "feat(lsel-${PROPOSAL_ID}): LSEL apply — ${TARGET}

lsel-${PROPOSAL_ID}
Playback via lsel-apply.sh (decision.json + diff.patch).

🗿 MoAI"; then
  echo "lsel-apply: git commit failed for proposal=$PROPOSAL_ID" >&2
  exit 1
fi
COMMIT_SHA="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"

# Backfill the real commit SHA into the last ledger row.
python3 - "$LEDGER" "$COMMIT_SHA" <<'PY'
import sys
ledger, sha = sys.argv[1], sys.argv[2]
with open(ledger) as f:
    lines = f.read().splitlines()
if lines:
    lines[-1] = lines[-1].replace('"commit_sha":"pending-backfill"', f'"commit_sha":"{sha}"')
with open(ledger, 'w') as f:
    f.write('\n'.join(lines) + '\n')
PY

echo "lsel-apply: applied proposal=$PROPOSAL_ID target=$TARGET commit=$COMMIT_SHA category=$CATEGORY"
