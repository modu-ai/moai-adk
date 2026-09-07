#!/usr/bin/env bash
# t513 reproduction — GH #1692 (spec status --sync-git ignores --dry-run) + #1693
# (status locator matches body table rows / backticked prose).
#
# Rebuilds a disposable fixture from scratch under /tmp/t513-repro and observes:
#   1. spec status --list            -> #1693 read-side: "Notes" reported as status
#   2. spec drift                    -> #1693 drift-side: false DRIFT records
#   3. spec status --sync-git --dry-run --yes
#                                    -> #1692: must write NOTHING; defect writes 2 files
#   4. hashes + git status + git diff -> the write, byte-level
#
# Scope guard: this script only ever touches /tmp/t513-repro (a disposable git
# fixture) and reads the moai binary it is handed. It never touches the
# repository's own .moai/specs/. The in-script git commands operate on the
# fixture repo only — the worktree guard cannot statically verify that, which is
# why the scope is stated here and bounded to /tmp.
#
# Usage: bash repro.sh <path-to-moai-binary>
set -uo pipefail

BIN="${1:?usage: repro.sh <moai-binary>}"
ROOT="/tmp/t513-repro"
FIX="$ROOT/fixture"

rm -rf "$ROOT"
mkdir -p "$FIX/.moai/specs/SPEC-DEMO-001" "$FIX/.moai/specs/SPEC-PROSE-001" "$FIX/.moai/specs/SPEC-EMDASH-001"

# --- Fixture SPECs (reporter examples, placeholder IDs preserved) ---

cat > "$FIX/.moai/specs/SPEC-DEMO-001/spec.md" <<'EOF'
---
id: SPEC-DEMO-001
status: completed
---

## HISTORY

| Version | Date | Status | Notes |
|---|---|---|---|
| 0.1.0 | 2026-01-01 | draft | initial |
EOF

cat > "$FIX/.moai/specs/SPEC-PROSE-001/spec.md" <<'EOF'
---
id: SPEC-PROSE-001
status: draft
---

## HISTORY

- HISTORY is kept as a bullet list, not a `| Version | Date | Status | Notes |` table, to stay clear of the status-locator parse artifact.
EOF

cat > "$FIX/.moai/specs/SPEC-EMDASH-001/spec.md" <<'EOF'
---
id: SPEC-EMDASH-001
status: completed
---

## HISTORY

- 0.1.0 initial draft
- 0.2.0 implemented and closed

This SPEC's body intentionally avoids pipe tables so the status locator reads
the frontmatter cleanly, isolating the git-implied side of the drift probe.
EOF

# V3R6 progress markers so era classification reaches the git-implied walk
# (H-1 would otherwise exempt every SPEC as V2.x before git classification).
for id in SPEC-DEMO-001 SPEC-PROSE-001 SPEC-EMDASH-001; do
cat > "$FIX/.moai/specs/$id/progress.md" <<'EOF'
## §E.1 Plan-phase Audit-Ready Signal
plan_status: audit-ready

## §E.2 Run-phase Evidence
run evidence present

## §E.3 Run-phase Audit-Ready Signal
run audit-ready

## §E.4 Sync-phase Audit-Ready Signal
sync_commit_sha: "76e66d5592a8a326479e36515688e8bbcf743345"
EOF
done

# --- Fixture git history: SPEC-ID commits + em-dash close behind a --no-ff merge ---
git -C "$FIX" init -q -b main
git -C "$FIX" add -A
git -C "$FIX" -c user.name=t -c user.email=t@t.local commit -qm "chore: seed specs"
git -C "$FIX" -c user.name=t -c user.email=t@t.local commit -qm "feat(SPEC-DEMO-001): M1 implementation" --allow-empty
git -C "$FIX" -c user.name=t -c user.email=t@t.local commit -qm "feat(SPEC-PROSE-001): M1 implementation" --allow-empty
git -C "$FIX" checkout -qb side
git -C "$FIX" -c user.name=t -c user.email=t@t.local commit -qm "feat(SPEC-EMDASH-001): M1 implementation" --allow-empty
git -C "$FIX" -c user.name=t -c user.email=t@t.local commit -qm "docs(SPEC-EMDASH-001): sync-phase artifacts — 3-phase close" --allow-empty
git -C "$FIX" checkout -q main
git -C "$FIX" -c user.name=t -c user.email=t@t.local merge -q --no-ff side -m "Merge branch 'side'"

echo "=== fixture git log (main) ==="
git -C "$FIX" log --oneline main

echo
echo "=== [1] spec status --list  (#1693 read-side: expect frontmatter values, NOT Notes) ==="
(cd "$FIX" && "$BIN" spec status --list 2>&1 | grep -v 'level=WARN')

echo
echo "=== [2] spec drift  (#1693 drift-side: false DRIFT records) ==="
(cd "$FIX" && "$BIN" spec drift 2>&1 | grep -v 'level=WARN')
echo "--- drift-cache.json ---"
cat "$FIX/.moai/state/drift-cache.json"

echo
echo "=== [3] hashes BEFORE sync-git --dry-run ==="
shasum -a 256 "$FIX"/.moai/specs/*/spec.md
echo "--- git status (expect empty) ---"
git -C "$FIX" status --porcelain -- '.moai/specs/*/spec.md'

echo
echo "=== [4] spec status --sync-git --dry-run --yes  (#1692: must write NOTHING) ==="
(cd "$FIX" && "$BIN" spec status --sync-git --dry-run --yes 2>&1 | grep -v 'level=WARN')

echo
echo "=== [5] hashes AFTER (defect: DEMO/PROSE changed) ==="
shasum -a 256 "$FIX"/.moai/specs/*/spec.md
echo "--- git status (defect: two M lines) ---"
git -C "$FIX" status --porcelain -- '.moai/specs/*/spec.md'

echo
echo "=== [6] corruption shape ==="
git -C "$FIX" diff -- .moai/specs
