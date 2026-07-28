# acceptance.md — SPEC-PREEDIT-PARALLEL-SESSION-GUARD-001

## A. Acceptance criteria

| ID | Requirement | Criterion (falsifiable) | Verify |
|----|-------------|------------------------|--------|
| AC-PES-001 | REQ-PES-001 | `agent-common-protocol.md` contains a `### Pre-Edit Sync Check` section that names the direct-edit trigger and reuses the Pre-Spawn command batch. | `grep -n "Pre-Edit Sync Check" .claude/rules/moai/core/agent-common-protocol.md` → ≥1 hit, section body present. |
| AC-PES-002 | REQ-PES-001 | The section specifies the trigger fires before non-trivial direct edits to shared paths (the §D path set). | grep for the path-set / "non-trivial direct edit" phrase in the section. |
| AC-PES-003 | REQ-PES-002 | On a foreign-session/race signal during direct-edit work, the procedure isolates (worktree) or surfaces AskUserQuestion (isolate/wait/abort). | grep for the interpretation-matrix reuse / AskUserQuestion isolate-wait-abort in the section. |
| AC-PES-004 | REQ-PES-002/003 | `worktree-integration.md` § auto-isolation trigger no longer requires "worktree entry chosen" as a hard conjunct for direct-edit work. | grep — the "worktree entry is chosen" phrase is qualified/relaxed for direct edits. |
| AC-PES-005 | REQ-PES-003 | The "direct edit bypasses the spawn gate" failure mode is named in doctrine (Pre-Spawn section + auto-isolation section). | grep for "bypass" + "spawn gate" (or equivalent) in both files. |
| AC-PES-006 | REQ-PES-004 | The PreToolUse-on-Edit hook is evaluated with a recorded cost finding + an explicit defer-or-implement decision. | acceptance.md §F records the finding + decision. |
| AC-PES-007 | REQ-PES-005 | An ambient foreign-session signal is specified (session-start note OR statusline segment). | grep in doctrine for the ambient-signal behavior. |
| AC-PES-008 | Template-First | Every `.claude/` rule edit is mirrored to `internal/template/templates/` byte-identical (sanitized-pair rules) and `make build` is clean. | `diff` local↔template per file; `make build` exit 0; `go test ./internal/template/... -run 'Neutrality\|Leak\|Mirror\|Parity'` PASS. |
| AC-PES-009 | Neutrality | No SPEC-ID / internal-date / commit-SHA leaks into the template mirrors. | `grep -rnoE 'SPEC-[A-Z0-9-]{6,}|REQ-[A-Z0-9-]{6,}|[0-9a-f]{7,40}' internal/template/templates/.claude/rules/moai/{core/agent-common-protocol,workflow/worktree-integration}.md` → 0 (date `updated:` frontmatter excepted per DC-1). |
| AC-PES-010 | Build | `go build ./...` exit 0; if a hook was added (M4 implemented), its unit test exists and passes. | `go build ./...` exit 0; `go test ./internal/hook/...` PASS. |

## B. Baseline (measured at run-phase entry)
- origin/main: `05117b6ba`
- `grep -c "Pre-Edit Sync Check" .claude/rules/moai/core/agent-common-protocol.md` = 0 (absent today)
- worktree-integration.md auto-isolation trigger currently requires "worktree entry is chosen".

## C. Falsifiability
- AC-PES-001/002/003/004/005/007 are presence-grep criteria; each is falsified by removing the named content → grep returns 0. Round-trip verified at run-phase (remove → 0, restore → ≥1).
- AC-PES-006 is a recorded-decision criterion; falsified by an empty §F (no decision recorded).
- AC-PES-008/009/010 are mechanical (diff/build/test/grep).

## D. Out of scope (deferred)
- A blocking PreToolUse-on-Edit hook if M4 defers it (recorded in §F).
- Auto-merge bot / branch-protection changes.
- `05117b6ba` history rewrite.

## E. Residual risk
- Procedural enforcement is not mechanical; the ambient signal + failure-mode naming raise compliance but cannot guarantee it. The hook (if implemented in M4) is the only mechanical guarantee — documented as the trade-off.

## F. Hook evaluation (REQ-PES-004) — to be filled at run-phase
- Cost finding: (registry-read `active-sessions.json` ≈ <1ms; per-edit `git fetch` ≈ 100ms–1s, unacceptable per-edit → gated to turn-first or omitted).
- Decision: [implement advisory-read-only / defer blocking hook] — recorded here at run-phase.
