---
id: SPEC-HOOK-FACTFORCE-ADVISORY-001
title: "GateGuard Fact-Force Advisory PreToolUse Hook (exit-0 rewrite)"
version: "0.1.0"
status: in-progress
created: 2026-07-08
updated: 2026-07-08
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/hooks/moai"
lifecycle: spec-anchored
tier: S
tags: "hook, pretooluse, fact-force, advisory, template-first"
---

# SPEC-HOOK-FACTFORCE-ADVISORY-001 — GateGuard Fact-Force Advisory PreToolUse Hook (exit-0 rewrite)

## HISTORY

| Date | Version | Change | Author |
|------|---------|--------|--------|
| 2026-07-08 | 0.1.0 | Initial plan-phase draft. Tier S. Supersedes SPEC-HOOK-PREEDIT-INVESTIGATE-001 (status: completed, version 0.1.1, sync_commit_sha 1c3e937df). Behavior change: the `gateguard-fact-force.sh` PreToolUse hook is rewritten from exit-2-blocking to exit-0-advisory — the FIRST Edit/Write/MultiEdit per (session, file path) now emits a one-time advisory `systemMessage` JSON on stdout and ALLOWS the operation (exit 0), instead of blocking with exit 2. Root cause: Claude Code renders exit-2 hooks unconditionally as a red "Error: PreToolUse:Write hook error" box, which non-developer users misread as a system crash. Two new NFRs added: jq-dependency prohibition (§C.5) and systemMessage JSON validity (§C.6). State-file logic, MOAI_FACT_FORCE=off opt-out, Read-as-investigation, self-loop prevention, fail-open, subagent boundary — all preserved unchanged from the predecessor. REQ-FA domain token distinguishes from predecessor's FF. | manager-spec |

## §A. Context and Intent

### §A.1 Why this SPEC supersedes the predecessor

The predecessor SPEC-HOOK-PREEDIT-INVESTIGATE-001 shipped a `gateguard-fact-force.sh` PreToolUse hook that **blocks** the first Edit/Write/MultiEdit per file per session with `exit 2` + a stderr guidance message. This SPEC changes it to **advise only**: emit a one-time advisory `systemMessage` JSON on stdout and `exit 0` (the edit is ALLOWED to proceed).

**The UX defect — recorded verbatim from the user directive:**

> When a PreToolUse hook exits 2, Claude Code's runtime renders it unconditionally as a red "Error: PreToolUse:Write hook error: [...]" box — even though it is an intentional advisory block. Non-developer users (the "모두의 클로드" / Claude-for-everyone audience) misread it as a system crash. Changing the stderr text does NOT remove the red error box (Claude Code rendering policy). Therefore `exit 2` itself must be removed. `exit 0` + stdout `systemMessage` is rendered by Claude Code as informational context, not a red error box.

This is a rendering-policy constraint, not a logic change. The investigation-guidance intent is preserved; only the delivery channel changes (exit-2 + stderr → exit-0 + stdout JSON systemMessage).

### §A.2 What is preserved unchanged from the predecessor

| Behavior | Predecessor REQ | This SPEC REQ | Changed? |
|----------|-----------------|---------------|----------|
| State-file keyed by (session_id, absolute_file_path) | REQ-FF-003 | REQ-FA-003 | No |
| Second-or-subsequent edit → no advisory, exit 0 | REQ-FF-002 | REQ-FA-002 | No (semantics identical) |
| MOAI_FACT_FORCE=off opt-out + skip log | REQ-FF-004 | REQ-FA-004 | No |
| Read-as-investigation (prior Read pre-populates state) | (in hook logic, not a separate REQ) | noted in §A.3 | No |
| Self-loop prevention on non-Edit/Write/MultiEdit payloads | REQ-FF-010 | REQ-FA-010 | No |
| Subagent context (state keyed by session_id, not agent_id) | REQ-FF-009 | REQ-FA-009 | No |
| Timeout safety (< 5s, O(1) shell-only) | REQ-FF-006 | REQ-FA-006 | No |
| Existing PreToolUse group preserved (additive-only) | REQ-FF-007 | REQ-FA-007 | No |
| No PostToolUse registration | REQ-FF-008 | REQ-FA-008 | No |
| Template-First mirror parity | REQ-FF-011 | REQ-FA-011 | No |
| State file hygiene (0o600, JSON-line, no active deletion) | REQ-FF-012 | REQ-FA-012 | No |
| Fail-open on any unexpected error | §C.2 NFR | §C.2 NFR | No |

### §A.3 Doctrinal alignment

The doctrinal alignment from the predecessor (§A.3) is preserved: the hook remains the preventive layer for CLAUDE.md §7 Rule 1 (Approach-First Development), Rule 4 (Reproduction-First Bug Fixing), and `verification-claim-integrity.md` §1.1 surface 1 + 2 (orchestrator self-report / manager-agent completion report). The change from exit-2-block to exit-0-advisory **weakens the mechanical enforcement** (the agent can now proceed without investigating) but preserves the **informational nudge** (the advisory still recommends investigation). This tradeoff is accepted because the UX defect (red error box misread as crash) caused more harm than the weakened enforcement — a red error box that non-developer users interpret as a system crash is a worse user experience than an informational advisory that the agent can ignore.

**Read-as-investigation** behavior is preserved: when the hook receives a `tool_name=Read` payload, it pre-populates the state file for that (session_id, file_path) tuple and exits 0 with no advisory. The next Edit/Write/MultiEdit on the same path skips the advisory because the state file already exists. This rewards investigation (Read) by suppressing the advisory on the subsequent edit.

### §A.4 Anti-conflict constraints (audit-verified, unchanged from predecessor)

All anti-conflict constraints from the predecessor (§A.4) remain in force: PostToolUse `"*"` slot occupied, existing PreToolUse `Write|Edit|Bash` group preserved, state session-scoped, subagent boundary enforced, 5s timeout, Template-First Rule. The ONLY constraint that changes is the hook's exit-code semantics (exit 2 removed; exit 0 on every path).

## §B. Functional Requirements (GEARS)

### §B.1 First-edit advisory behavior (CHANGED from predecessor REQ-FF-001)

**REQ-FA-001** (Ubiquitous) — The GateGuard Fact-Force hook shall emit a one-time ADVISORY notice via stdout `systemMessage` JSON on the FIRST Edit, Write, or MultiEdit operation on each file path within a session, and ALLOW the operation to proceed (exit 0). The notice shall recommend the agent investigate (a) who imports the file, (b) what data schemas / contracts the file touches, and (c) what user instruction justifies the edit. The hook shall NEVER exit with code 2 (it never blocks); every code path shall exit 0.

### §B.2 Second-and-subsequent edit allowance (preserved from REQ-FF-002)

**REQ-FA-002** (Event-driven) — **When** the second or subsequent Edit, Write, or MultiEdit operation on the same file path is detected within the same session, the hook shall allow the operation to proceed without emitting an advisory (exit 0, no systemMessage on stdout), and shall not modify the state file's `first_seen` timestamp.

### §B.3 Per-file session-scoped state tracking (preserved from REQ-FF-003)

**REQ-FA-003** (Ubiquitous) — The hook shall track per-file already-investigated state in a session-scoped state file located under `${CLAUDE_PROJECT_DIR:-$PWD}/.moai/state/fact-force/`, keyed by a hash of `session_id + absolute_file_path`, so that state does not leak across sessions and does not collide across different files within the same session.

### §B.4 Advisory opt-out (preserved from REQ-FF-004)

**REQ-FA-004** (Capability gate) — **Where** the environment variable `MOAI_FACT_FORCE=off` is set, the hook shall skip the advisory entirely (exit 0 unconditionally, no systemMessage) and append a one-line bypass record to `.moai/logs/fact-force-skip.log` noting the timestamp, session_id (if available), and the first-edit file path (if extractable from the payload) for audit trail purposes.

### §B.5 Hook communication discipline (CHANGED from predecessor REQ-FF-005)

**REQ-FA-005** (Unwanted behavior) — The GateGuard Fact-Force hook shall not invoke `AskUserQuestion` or any other user-prompting mechanism, shall not exit with code 2 (the hook never blocks), and shall not communicate via any channel other than stdout (advisory `systemMessage` JSON on first edit), stderr (diagnostic log), and exit code 0 (allow — every path exits 0).

### §B.6 Timeout safety (preserved from REQ-FF-006)

**REQ-FA-006** (Ubiquitous) — The hook shall self-terminate within the 5-second hook timeout registered in settings.json; the implementation MUST NOT perform any network I/O, LSP queries, or subprocess spawns beyond the `O(1)` state-file existence check, a single payload read, and the advisory JSON emission.

### §B.7 Existing PreToolUse group preservation (preserved from REQ-FF-007)

**REQ-FA-007** (Ubiquitous) — The hook shall preserve the existing PreToolUse `"matcher": "Write|Edit|Bash"` group's behavior unmodified; the hook is registered as a SEPARATE PreToolUse matcher group scoped to `Edit|Write|MultiEdit` (NOT including Bash), so the two groups' concerns are isolated.

### §B.8 PostToolUse wildcard prohibition (preserved from REQ-FF-008)

**REQ-FA-008** (Unwanted behavior) — The hook shall not register any PostToolUse matcher, including the `"*"` wildcard; the PostToolUse no-matcher slot is already occupied by `handle-harness-observe.sh` and the `"Write|Edit"` slot by `handle-post-tool.sh` + `status-transition-ownership.sh`.

### §B.9 Subagent context handling (preserved from REQ-FF-009)

**REQ-FA-009** (Event-driven) — **When** the hook is invoked from a subagent context (the payload contains an `agent_id` field per Claude Code v2.1.69+), the hook shall apply the same per-file-per-session advisory with no subagent exemption, AND the state remains keyed by `session_id` (the parent session, not the agent_id), so that a file investigated by the orchestrator in the parent session does not produce a new advisory when a subagent edits it later in the same session.

### §B.10 Self-loop prevention (preserved from REQ-FF-010)

**REQ-FA-010** (Capability gate) — **Where** the hook detects that its own payload's `tool_name` is not one of `Edit`, `Write`, `MultiEdit`, or `Read` (e.g., it was accidentally invoked on its own Write to the state file, or on a `Bash` tool call), the hook shall exit 0 immediately without further processing, ensuring the hook cannot emit advisories on its own state-file writes or otherwise recurse.

### §B.11 Template-First mirror parity (preserved from REQ-FF-011)

**REQ-FA-011** (Ubiquitous) — The hook script and the settings.json edit SHALL be authored in the template tree (`internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh` + `internal/template/templates/.claude/settings.json.tmpl`) FIRST, then regenerated via `make build`, then mirrored to the local project (`.claude/hooks/moai/gateguard-fact-force.sh` + `.claude/settings.json`), per the CLAUDE.local.md §2 Template-First Rule.

### §B.12 State file hygiene (preserved from REQ-FF-012)

**REQ-FA-012** (State-driven) — **While** a session is active, the hook SHALL NOT delete or compact the state files; **when** a new session starts, the previous session's state files become orphaned but are not actively removed (cleanup is a follow-up concern owned by a separate SPEC, not this one). State files MUST be small (a single JSON line each) and MUST be created with `0o600` permissions to avoid leaking absolute file paths to other users on shared hosts.

## §C. Non-Functional Requirements (Constraints)

### §C.1 Performance (unchanged from predecessor)

- Each hook invocation's execution latency MUST be < 100ms on a warmed filesystem, measured from stdin payload read to exit. The implementation is shell-only: one payload read (`head -c 1048576`), one `tool_name` extraction (`grep`), one `file_path` extraction (`grep`), one state-file existence test (`[ -f ... ]`), at most one state-file write (`cat > ...`), and at most one advisory JSON emission (`printf` on stdout). No `moai` binary invocation. No network. No LSP. No `jq`.

### §C.2 Robustness (unchanged from predecessor)

- The hook MUST fail-open on any unexpected error (missing `CLAUDE_PROJECT_DIR`, unwritable state dir, malformed JSON payload, absent `session_id`). A fail-open default of "allow the edit" (exit 0, no advisory) is preferable to a fail-closed default that would lock the agent out of the session. The state-file write is wrapped in `mkdir -p 2>/dev/null || exit 0` so an unwritable state directory degrades to "no advisory" rather than "session halt".

### §C.3 Observability (CHANGED — advisory channel moves from stderr to stdout)

- The hook emits exactly two observable artifacts: (a) the advisory `systemMessage` JSON on **stdout** when emitting the first-edit advisory (exit 0), and (b) the bypass record appended to `.moai/logs/fact-force-skip.log` when `MOAI_FACT_FORCE=off` (REQ-FA-004). The predecessor emitted the guidance on **stderr** (because exit-2 semantics required the block reason on stderr per CC 2.1.202); the new version emits on **stdout** because exit-0 + stdout `systemMessage` is the Claude Code informational-context channel (NOT a red error box). No other log surface.

### §C.4 Portability (unchanged from predecessor)

- The hook is a `/bin/bash` script compatible with macOS (bash 3.2, the system default) and Linux (bash 4+). It uses only POSIX-adjacent utilities: `head`, `grep`, `sed`, `awk`, `printf`, `date`, `mkdir`, `shasum` (macOS) / `sha1sum` (Linux). The implementation MUST detect which hash utility is available and fall back gracefully, OR use a portable hashing alternative (e.g., a sanitized path-string filename with `/` → `_` substitution).

### §C.5 jq dependency prohibition (NEW — not in predecessor)

- The hook MUST NOT depend on `jq`. The hook is classified `self-contained / dependency-free` in `hook-independence.md` §3 mode G (acceptable-by-design). Adding `jq` as a runtime dependency would break that classification and introduce a shared-failure-mode correlation not present today. The multiline `systemMessage` JSON escaping MUST use `awk` only (backslash → `\\`, double-quote → `\"`, inter-line newline → literal `\n`). The hook's output MUST be valid JSON parseable by `jq` on the verification side, but the hook itself MUST NOT invoke `jq` to produce that output.

### §C.6 systemMessage JSON validity (NEW — not in predecessor)

- The emitted stdout JSON MUST be valid such that `jq -e '.systemMessage | type'` returns `"string"`. Raw newlines are PROHIBITED inside the JSON string value — all inter-line newlines in the advisory text MUST be escaped as literal `\n` sequences (backslash + `n`) before emission. The JSON MUST be a single-line `{"systemMessage":"..."}` object on stdout, terminated by a newline.

## §D. Acceptance Criteria Summary

Acceptance criteria are enumerated in `acceptance.md`. Cross-reference index:

| AC ID | Requirement | Severity |
|-------|-------------|----------|
| AC-FA-001 | First-edit advisory on Edit/Write/MultiEdit (exit 0 + systemMessage) | MUST |
| AC-FA-002 | Second-edit no-advisory on same file path (exit 0, no output) | MUST |
| AC-FA-003 | State file keyed by session_id + absolute path (advisory fires per-session) | MUST |
| AC-FA-004 | MOAI_FACT_FORCE=off advisory opt-out + skip log | MUST |
| AC-FA-005 | Zero exit-2 paths + no AskUserQuestion (hook discipline grep) | MUST |
| AC-FA-006 | Self-termination within 5s timeout | MUST |
| AC-FA-007 | Existing handle-pre-tool.sh behavior preserved | MUST |
| AC-FA-008 | No PostToolUse "*" registration | MUST |
| AC-FA-009 | Subagent context handled correctly | SHOULD |
| AC-FA-010 | Self-loop prevention on non-Edit/Write/MultiEdit/Read payloads | MUST |
| AC-FA-011 | Template-First mirror parity (template + local byte-identical) | MUST |
| AC-FA-012 | State file hygiene (0o600, JSON-line, no active deletion) | SHOULD |
| AC-FA-013 | Read-as-investigation (prior Read → next Write skips advisory) | MUST |
| AC-FA-014 | systemMessage JSON validity (parses via jq) | MUST |
| AC-FA-015 | Fail-open on missing session_id (exit 0, no advisory) | MUST |

## §E. Out of Scope

### Out of Scope — Predecessor's exit-2 block behavior

- The exit-2 block semantics are NOT preserved. The hook no longer blocks any edit under any condition. If a future use case requires a hard block (not just an advisory), a SEPARATE PreToolUse hook with a different name and a different matcher group MUST be created. This SPEC's hook is advisory-only — `exit 0` on every path, by design.

### Out of Scope — Confidence scoring / learning / instinct evolution

- This SPEC's hook is stateless beyond the per-file-per-session binary "already-advised" flag. Confidence scoring, learning from past investigations, instinct evolution, and any form of ML-based gate decision-making are OUT OF SCOPE. The advisory is deterministic: first attempt → advise; subsequent attempts → no advise; opt-out via env var.

### Out of Scope — State file cleanup / GC

- The hook creates state files but does not delete them. Active cleanup (e.g., SessionEnd sweep of stale state files, age-based GC, cross-session deduplication) is a separate concern owned by a follow-up SPEC. This SPEC only requires that state files are small (single JSON line) and session-scoped (keyed by session_id), so the storage footprint is bounded by the number of unique (session_id, file_path) pairs in the project's history.

### Out of Scope — Go-side `internal/hook/fact_force.go`

- A Go-side handler (`moai hook fact-force` subcommand) following the `internal/hook/pre_tool.go` pattern is OPTIONAL and is NOT implemented in this SPEC. The shell-only implementation is sufficient for the Tier S scope.

### Out of Scope — Team-mode / TaskCompleted integration

- The hook is a PreToolUse gate, not a TeammateIdle or TaskCompleted gate. Team-mode AC verification is owned by `team-ac-verify.sh` (currently dormant) and is NOT touched.

### Out of Scope — settings.json edit

- The predecessor SPEC already added the PreToolUse matcher group for `Edit|Write|MultiEdit` → `gateguard-fact-force.sh` to `.claude/settings.json` and its template. This SPEC does NOT change the settings.json registration — only the hook script's behavior changes (exit 2 → exit 0). The matcher group, timeout, and type remain identical.

### Out of Scope — hook-independence.md body changes beyond the Mode G row wording

- The ONLY change to `hook-independence.md` is the Mode G row rationale column: the phrase `"no gate"` is replaced with `"no advisory notice"` to reflect that the hook no longer gates (blocks) but advises. No other row in the catalogue is touched; no new row is added; the classification (`acceptable-by-design`) is unchanged.

## §F. Risks and Open Questions

### §F.1 Risk: weakened enforcement (advisory, not block)

Changing from exit-2-block to exit-0-advisory weakens the mechanical enforcement. An agent CAN now proceed with an edit without investigating — the advisory is informational, not a gate. This tradeoff is accepted because the UX defect (red error box misread as crash) was the root cause for superseding the predecessor. The `MOAI_FACT_FORCE` env var name is preserved for backward compatibility, but its semantics change: `MOAI_FACT_FORCE=off` now skips the advisory (not the block); there is no block to skip.

### §F.2 Risk: systemMessage rendering varies by Claude Code version

The `systemMessage` stdout JSON field is documented in `hooks-system.md` as a standard Claude Code hook output. Its rendering as informational context (not a red error box) is the Claude Code runtime's responsibility. If a future Claude Code version changes how `systemMessage` is rendered, the advisory may become more or less prominent. This is outside the hook's control.

### §F.3 Risk: awk escaping edge cases

The awk-based JSON escape (backslash → `\\`, double-quote → `\"`, newline → `\n`) covers the three character classes that appear in the advisory text. If the advisory text is later extended to include tab characters, control characters, or non-ASCII Unicode that requires special JSON handling, the awk escape MUST be extended accordingly. The current advisory text (§B.1) contains only printable ASCII + newlines, so the 3-escape awk is sufficient.

## §G. Cross-References

- **Predecessor SPEC**: `.moai/specs/SPEC-HOOK-PREEDIT-INVESTIGATE-001/` — the completed SPEC being superseded (status transition: `completed → superseded`, owned by manager-spec per the Status Transition Ownership Matrix)
- **CLAUDE.md §7** — Rule 1 (Approach-First) + Rule 4 (Reproduction-First) doctrinal anchors (weakened from mechanical enforcement to informational nudge, per §A.3)
- **`verification-claim-integrity.md` §1.1** — surface 1 + 2 preventive layer (preserved as informational nudge)
- **`hooks-system.md` § Hook Execution Types** — Command hook stdout JSON `systemMessage` field semantics
- **`hook-independence.md` § 3** — Mode G row; the rationale wording changes from `"no gate"` to `"no advisory notice"` (run-phase M1 step)
- **CLAUDE.local.md §2** — Template-First Rule
- **CLAUDE.local.md §7** — 5s hook timeout policy
- **C-HRA-008** — subagent boundary grep acceptance criterion
