# Card t469 — Integration-window re-check (develop absorb + merge tree)

Branch: `WT-achwd-strip-exempt` · Window: acquired as lane-9 (`moai integration acquire --name lane-9`, rc=0)

## Pre-merge checks (lead-mandated)

- `git config core.fsmonitor` → (no output), rc=1 — unset
- `find .git -name index.lock` → 0 entries
- One `fatal: Unable to write index.` (rc=128, no path, fsmonitor unset, lock 0) — lead disposition form 3 (cause B): one retry approved; retry succeeded into the CHANGELOG conflict below.

## What this branch carries into develop

`git diff --name-only develop..HEAD` at absorb tip = **15 files, 0 `.go` files**:
`.moai/reports/t469/**` (8), `.moai/specs/{SPEC-ACHWD-STRIP-EXEMPT-001,SPEC-HOOK-WIRING-DRIFT-001}/**` (6), `CHANGELOG.md` (1).
No Go package is touched — Go package tests are therefore out of the re-check selector and are reported as NOT measured.

## CHANGELOG conflict resolution (absorb of develop `f942d1ea1`)

- Conflict in `CHANGELOG.md` `[Unreleased]` only (single unmerged file, `UU` count 1).
- Duplicate entries `SPEC-HOOK-WIRING-DRIFT-001` (t216) and the t456 statusline entry were **byte-identical** across the two sides (measured: `diff` on the paired lines, both silent) — develop-side single copy retained, own t469 entry appended after the six develop entries (t466/t462/t465/t216/t470/t456).
- Conflict markers after resolution: `grep -n '^<<<<<<<\|^=======$\|>>>>>>>'` → no match, rc=1.
- Absorb commit: `d0edf3c5b` (develop tip before absorb: `f942d1ea1`, confirmed via `git show-ref refs/heads/develop`).

## Round 1 — re-check at absorbed tip `d0edf3c5b`

- `gofmt -l .` → (no output), rc=0
- `perl .moai/reports/t469/sync-evidence/t469-mirror-check.pl .claude/rules/moai/development/hook-independence.md .claude/rules/moai/core/agent-common-protocol.md .claude/rules/moai/core/agent-common-protocol-reference.md` → (no output), rc=0 — strip-aware mirror identity holds after the develop absorb.

## Round 2 — deciding round at the evidence-commit tip

Round 2 runs the identical two commands at the tip this evidence commit creates (per the lane-4 lesson: the measured tree must be the tree being merged). Its verbatim result is recorded in the lane completion report; the merge commit's tree is verified byte-identical to that tip's tree via `git rev-parse <merge>^{tree}` before `moai integration release`.

## Not measured

- Go package tests (0 packages touched by this branch — nothing in the selector to run).
- The full suite (lane discipline — CI owns it after the lead's develop push).
