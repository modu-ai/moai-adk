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
| AC-PES-009 | Neutrality | No SPEC-ID / internal-date / commit-SHA leaks into the template mirrors. | Neutrality/leak grep on the two template mirrors returns 0 matches (date `updated:` frontmatter excepted per DC-1); verified by the CI Template Neutrality Audit check. |
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

## F. Hook evaluation (REQ-PES-004) — recorded at run-phase (commit 87938efa3)

- **Decision: implement the advisory-read-only hook** (M4). A blocking PreToolUse-on-Edit hook is **deferred** to a follow-up SPEC.
- **Cost finding (measured against the canonical in-process reader):**
  - Registry read (`session.NewRegistry(path, nil).Query("")`) is an in-process `os.ReadFile` + `json.Unmarshal` of `.moai/state/active-sessions.json`, completing in **<1 ms** per edit. `Query` reaches `readAllUnlocked()` only and does NOT touch the registry `Clock`, so a `nil` clock is safe; no subprocess is spawned.
  - A per-edit `git fetch` (~100 ms–1 s network round-trip) is **omitted** — unacceptable per-edit latency. The advisory relies on the registry's heartbeat-driven staleness (`session.DefaultStaleMinutes = 30`) instead of a live fetch; a dead-PID entry with a fresh heartbeat is counted as foreign (conservative over-count — the fail-open advisory tolerates it).
  - Because the advisory never blocks (returns the empty allow), even the <1 ms read is paid for awareness only, never for gating.
- **REQ-PES-005 (ambient signal) — already satisfied:** `internal/hook/session_start.go` Step 3 already calls `session.QueryActiveWork` → `session.FormatStderrReminder(input.SessionID, entries, now)`, emitting a `<system-reminder>` that lists foreign active sessions to stderr at session start. No additional code is needed for M5; this §F records that REQ-PES-005 is met by the existing SessionStart hook. The blocking hook remains the only deferred mechanical backstop (follow-up SPEC).
