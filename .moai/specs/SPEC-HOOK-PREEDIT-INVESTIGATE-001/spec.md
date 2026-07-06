---
id: SPEC-HOOK-PREEDIT-INVESTIGATE-001
title: "GateGuard Fact-Force PreToolUse Hook (ECC Adaptation)"
version: "0.1.1"
status: completed
created: 2026-07-07
updated: 2026-07-07
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/hooks/moai"
lifecycle: spec-anchored
tier: S
tags: "hook, pretooluse, fact-force, gateguard, template-first"
---

# SPEC-HOOK-PREEDIT-INVESTIGATE-001 — GateGuard Fact-Force PreToolUse Hook (ECC Adaptation)

## HISTORY

| Date | Version | Change | Author |
|------|---------|--------|--------|
| 2026-07-07 | 0.1.0 | Initial plan-phase draft. Tier S. Adapts ECC `pre:edit-write:gateguard-fact-force` hook pattern as a moai-adk PreToolUse shell gate. Orchestrator performed 4-surface read-only audit (settings.json:38-49, agent-common-protocol.md:209, internal/hook/pre_tool.go, PostToolUse "*" occupancy) — confirmed the broad PreToolUse `Edit|Write|MultiEdit` matcher slot is empty and the proposed gate is additive-only. User scoped this SPEC to F (GateGuard Fact-Force) ONLY; candidates B / G / H / A are explicitly Out of Scope. | manager-spec |
| 2026-07-07 | 0.1.1 | iter-2 defect fixes per plan-auditor verdict (PASS-WITH-DEBT 0.89). D1 (SHOULD-FIX): aligned hook-independence.md §3 catalogue-update phase across all 3 surfaces — spec.md §G changed from "sync phase" to "run-phase (M1 step 6)" to match plan.md §F.1 step 6 + acceptance.md §E. D2 (SHOULD-FIX): fixed pre-flight shasum/sha1sum operator in plan.md §C step 6 (`&&` → `||`) — macOS ships shasum only, so AND logic falsely failed. D3 (SHOULD-FIX): resolved session_id-absent contradiction — aligned acceptance.md E4 with spec.md §C.2 fail-open (removed sentinel "default" fallback; fail-open is the safer default). D4 (MINOR): weakened latency framing from "p99 < 100ms" to "each invocation < 100ms" in spec.md §C.1 + acceptance.md AC-FF-006 + §F (5 samples ≈ p80 not p99; per-invocation is the real concern for an O(1) shell hook). No REQ-FF or AC-FF-006 ID changes; GEARS intact; Out-of-Scope candidates B/G/H/A still excluded. | manager-spec |

## §A. Context and Intent

### §A.1 Why this hook

MoAI Rule 1 (Approach-First Development, CLAUDE.md §7) and the `agent-common-protocol.md` § File Operations Pattern "ALWAYS Read a file before using Edit" guideline both depend on agent self-discipline. There is no mechanical gate that forces the agent to investigate a file (importers, data schemas, user-instruction alignment) before its first edit in a session. The result is that an agent can produce an unobserved edit on a file it has not investigated, then claim the edit was correct — a `verification-claim-integrity.md` §1.1 surface-1 hazard (orchestrator self-report / agent completion report claiming success it did not observe).

This SPEC adapts a proven pattern from the external ECC repository (github.com/affaan-m/ECC, hook id `pre:edit-write:gateguard-fact-force`) to close that gap. The ECC source description is verbatim:

> "Fact-forcing gate: block first Edit/Write/MultiEdit per file and demand investigation (importers, data schemas, user instruction) before allowing"

Mechanism: for each file path, the FIRST Edit/Write/MultiEdit in a session is blocked with exit-code 2 and a guidance message demanding investigation; the SECOND attempt on the same file path is allowed. This converts the "Read-before-write" recommendation from a non-mechanical guideline into a deterministic per-file-per-session gate.

### §A.2 Why this is a real gap (audit-confirmed, file:line)

The orchestrator performed a 4-surface read-only audit before commissioning this SPEC. The findings relevant to the proposed hook:

| Surface | Finding | Implication |
|---------|---------|-------------|
| `.claude/settings.json:38-49` | PreToolUse has ONE matcher group: `"matcher": "Write|Edit|Bash"` → `handle-pre-tool.sh` (security/secrets/reflective-write). | The broad PreToolUse `Edit|Write|MultiEdit` slot is empty; a NEW scoped matcher group can be added without touching the existing group. |
| `.claude/settings.json:50-77` | PostToolUse has TWO groups: `"matcher": "Write|Edit"` (handle-post-tool + status-transition-ownership) and a no-matcher group (`handle-harness-observe.sh`). | PostToolUse `"*"` slot is ALREADY OCCUPIED by `handle-harness-observe.sh`. The proposed hook MUST NOT add a PostToolUse `"*"` entry. |
| `agent-common-protocol.md` § File Operations Pattern | "ALWAYS Read a file before using Edit" is a GUIDELINE enforced by agent self-discipline, NOT a hook. | The proposed gate is a NEW mechanical layer, not a duplication of an existing gate. |
| `internal/hook/pre_tool.go` | `handle-pre-tool.sh` forwards to `moai hook pre-tool`. The existing pattern is shell-wrapper → Go subcommand. | If Go routing is later desired, the `moai hook fact-force` subcommand can be added following the same forwarding pattern. Tier S scope prefers shell-only. |

### §A.3 Doctrinal alignment

The hook is the preventive layer for three existing doctrines:

1. **CLAUDE.md §7 Rule 1 (Approach-First Development)** — the hook mechanically enforces "explain approach + which files change + why" at the run-phase layer by requiring the first edit to be preceded by an investigation step.
2. **CLAUDE.md §7 Rule 4 (Reproduction-First Bug Fixing)** — by demanding the agent articulate the user instruction and existing data schemas before editing, the hook reduces the probability of fixing a symptom rather than a root cause.
3. **`verification-claim-integrity.md` §1.1 surface 1 (orchestrator self-report) + surface 2 (manager-agent completion report)** — by blocking uninvestigated edits, the hook prevents unobserved claims from entering the completion-report pipeline.

The hook also adopts the ECC `observe.sh` self-loop-prevention pattern (entrypoint filter → profile → skip-env → agent_id → path exclusion) at a simplified level appropriate to the moai-adk hook topology, so it does not fire on its own tool calls or spurious subagent contexts.

### §A.4 Anti-conflict constraints (audit-verified)

| Constraint | Source | How the SPEC respects it |
|------------|--------|--------------------------|
| PostToolUse `"*"` slot is occupied by `handle-harness-observe.sh` | settings.json:67-76 | REQ-FF-008 prohibits PostToolUse registration. |
| Existing PreToolUse `Write\|Edit\|Bash` group behavior MUST be preserved | settings.json:38-49 | REQ-FF-007 requires additive-only — a SEPARATE matcher group, not modification of the existing one. |
| `status-transition-ownership.sh` (PostToolUse) is a separate concern | settings.json:59-65 | Out of scope — not touched. |
| Per-file state MUST be session-scoped to avoid cross-session contamination | `session-handoff.md` multi-session coordination policy | REQ-FF-003 + REQ-FF-009 require state file keyed by `session_id + absolute_file_path`. |
| Subagent boundary — hooks MUST NOT invoke AskUserQuestion | C-HRA-008, `agent-common-protocol.md` § User Interaction Boundary | REQ-FF-005 prohibits AskUserQuestion; the hook is a shell script and has no AskUserQuestion surface, but the requirement is stated explicitly for lint/audit traceability. |
| 5s hook timeout policy | `CLAUDE.local.md` §7, `hooks-system.md` § Timeout Configuration | REQ-FF-006 requires self-termination within 5s; the state-file check is an `O(1)` file-existence probe. |
| Template-First Rule | `CLAUDE.local.md` §2 [HARD] | REQ-FF-010 + plan.md §F.1 require the template tree edit to happen FIRST, then `make build`, then local mirror. |

## §B. Functional Requirements (GEARS)

### §B.1 First-edit blocking behavior

**REQ-FF-001** (Ubiquitous) — The GateGuard Fact-Force hook shall block the FIRST Edit, Write, or MultiEdit operation on each file path within a session by exiting with code 2 and emitting a guidance message that demands the agent investigate (a) who imports the file, (b) what data schemas / contracts the file touches, and (c) what user instruction justifies the edit, before allowing the operation to proceed on the next attempt.

### §B.2 Second-and-subsequent edit allowance

**REQ-FF-002** (Event-driven) — **When** the second or subsequent Edit, Write, or MultiEdit operation on the same file path is detected within the same session, the hook shall allow the operation to proceed without blocking (exit 0, no guidance message).

### §B.3 Per-file session-scoped state tracking

**REQ-FF-003** (Ubiquitous) — The hook shall track per-file already-investigated state in a session-scoped state file located under `${CLAUDE_PROJECT_DIR:-$PWD}/.moai/state/fact-force/`, keyed by a hash of `session_id + absolute_file_path`, so that state does not leak across sessions and does not collide across different files within the same session.

### §B.4 Advisory opt-out

**REQ-FF-004** (Capability gate) — **Where** the environment variable `MOAI_FACT_FORCE=off` is set, the hook shall skip the fact-force gate (exit 0 unconditionally) and append a one-line bypass record to `.moai/logs/fact-force-skip.log` noting the timestamp, session_id (if available), and the first-edit file path (if extractable from the payload) for audit trail purposes.

### §B.5 Subagent boundary

**REQ-FF-005** (Unwanted behavior) — The hook shall not invoke `AskUserQuestion` or any other user-prompting mechanism; the hook is a shell script and communicates only via stdout (guidance message), stderr (diagnostic log), and exit code (0 = allow, 2 = block).

### §B.6 Timeout safety

**REQ-FF-006** (Ubiquitous) — The hook shall self-terminate within the 5-second hook timeout registered in settings.json (per CLAUDE.local.md §7 hook timeout policy); the implementation MUST NOT perform any network I/O, LSP queries, or subprocess spawns beyond the `O(1)` state-file existence check and a single payload read.

### §B.7 Existing PreToolUse group preservation

**REQ-FF-007** (Ubiquitous) — The hook shall preserve the existing PreToolUse `"matcher": "Write|Edit|Bash"` group's behavior unmodified; the new hook is registered as a SEPARATE PreToolUse matcher group scoped to `Edit|Write|MultiEdit` (note: NOT including Bash, which remains the sole concern of the existing group), so the two groups' concerns are isolated.

### §B.8 PostToolUse wildcard prohibition

**REQ-FF-008** (Unwanted behavior) — The hook shall not register any PostToolUse matcher, including the `"*"` wildcard; the PostToolUse no-matcher slot is already occupied by `handle-harness-observe.sh` and the `"Write|Edit"` slot by `handle-post-tool.sh` + `status-transition-ownership.sh`.

### §B.9 Subagent context handling

**REQ-FF-009** (Event-driven) — **When** the hook is invoked from a subagent context (the payload contains an `agent_id` field per Claude Code v2.1.69+), the hook shall apply the same per-file-per-session gate with no subagent exemption, AND the state remains keyed by `session_id` (the parent session, not the agent_id), so that a file investigated by the orchestrator in the parent session is not re-gated when a subagent edits it later in the same session.

### §B.10 Self-loop prevention

**REQ-FF-010** (Capability gate) — **Where** the hook detects that its own payload's `tool_name` is not one of `Edit`, `Write`, or `MultiEdit` (e.g., it was accidentally invoked on its own Write to the state file), the hook shall exit 0 immediately without further processing, ensuring the hook cannot gate its own state-file writes or otherwise recurse.

### §B.11 Template-First mirror parity

**REQ-FF-011** (Ubiquitous) — The hook script and the settings.json edit SHALL be authored in the template tree (`internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh` + `internal/template/templates/.claude/settings.json.tmpl`) FIRST, then regenerated via `make build`, then mirrored to the local project (`.claude/hooks/moai/gateguard-fact-force.sh` + `.claude/settings.json`), per the CLAUDE.local.md §2 Template-First Rule.

### §B.12 State file hygiene

**REQ-FF-012** (State-driven) — **While** a session is active, the hook SHALL NOT delete or compact the state files; **when** a new session starts, the previous session's state files become orphaned but are not actively removed (cleanup is a follow-up concern owned by a separate SPEC, not this one). State files MUST be small (a single JSON line each) and MUST be created with `0o600` permissions to avoid leaking absolute file paths to other users on shared hosts.

## §C. Non-Functional Requirements (Constraints)

### §C.1 Performance

- Each hook invocation's execution latency MUST be < 100ms on a warmed filesystem, measured from stdin payload read to exit. (Note: this is an O(1) shell hook where per-invocation latency is the real concern; the verification samples 5 invocations as a practical check, not a statistical percentile.) The implementation is shell-only: one payload read (head -c 1048576), one tool_name extraction (grep), one file_path extraction (grep), one state-file existence test (`[ -f ... ]`), at most one state-file write (`cat > ...`). No `moai` binary invocation. No network. No LSP.

### §C.2 Robustness

- The hook MUST fail-open on any unexpected error (missing CLAUDE_PROJECT_DIR, unwritable state dir, malformed JSON payload, absent session_id). A fail-open default of "allow the edit" is preferable to a fail-closed default that would lock the agent out of the session. The state-file write is wrapped in `mkdir -p 2>/dev/null || exit 0` so an unwritable state directory degrades to "no gate" rather than "session halt".

### §C.3 Observability

- The hook emits exactly two log artifacts: (a) the guidance message on stdout when blocking (exit 2), and (b) the bypass record on the `.moai/logs/fact-force-skip.log` append when MOAI_FACT_FORCE=off (REQ-FF-004). No other log surface.

### §C.4 Portability

- The hook is a `/bin/bash` script compatible with macOS (bash 3.2, the system default) and Linux (bash 4+). It uses only POSIX-adjacent utilities: `head`, `grep`, `sed`, `shasum` (macOS) / `sha1sum` (Linux) — the implementation MUST detect which is available and fall back gracefully, OR use a portable hashing alternative (e.g., a sanitized path-string filename with `/` → `_` substitution).

## §D. Acceptance Criteria Summary

Acceptance criteria are enumerated in `acceptance.md`. Cross-reference index:

| AC ID | Requirement | Severity |
|-------|-------------|----------|
| AC-FF-001 | First-edit block on Edit/Write/MultiEdit | MUST |
| AC-FF-002 | Second-edit allow on same file path | MUST |
| AC-FF-003 | State file keyed by session_id + absolute path | MUST |
| AC-FF-004 | MOAI_FACT_FORCE=off advisory opt-out + log | MUST |
| AC-FF-005 | No AskUserQuestion invocation | MUST |
| AC-FF-006 | Self-termination within 5s timeout | MUST |
| AC-FF-007 | Existing handle-pre-tool.sh behavior preserved | MUST |
| AC-FF-008 | No PostToolUse "*" registration | MUST |
| AC-FF-009 | Subagent context handled correctly | SHOULD |
| AC-FF-010 | Self-loop prevention on non-Edit/Write/MultiEdit payloads | MUST |
| AC-FF-011 | Template-First mirror parity (template + local) | MUST |
| AC-FF-012 | State file hygiene (0o600, JSON-line, no active deletion) | SHOULD |

## §E. Out of Scope

### Out of Scope — Candidate B (Search-First manager-spec body change)

- The ECC catalog contains a separate candidate (B) that would modify the `manager-spec` agent body to enforce "search the existing codebase before authoring a new SPEC". This is a SEPARATE Tier S SPEC and is explicitly NOT included in this SPEC's scope. The fact-force gate is a runtime mechanical hook; the Search-First change is an agent-authoring policy change. Mixing the two would violate scope discipline.

### Out of Scope — Candidate G (Config Protection)

- A separate candidate (G) for protecting `.moai/config/` from unauthorized modification was identified in the ECC audit. It is NOT included in this SPEC.

### Out of Scope — Candidate H (MCP Health Check)

- A separate candidate (H) for periodic MCP server health checking was identified. It is NOT included in this SPEC.

### Out of Scope — Candidate A (Instinct Learning)

- The ECC catalog contains an "instinct learning" hook that learns from past edits. moai-adk already has `internal/cli/preference/` and the harness-learner 4-tier subsystem covering the same concern. Candidate A is废弃 (abandoned) and is NOT included in this SPEC or any follow-up.

### Out of Scope — Confidence scoring / learning / instinct evolution

- This SPEC's hook is stateless beyond the per-file-per-session binary "already-investigated" flag. Confidence scoring, learning from past investigations, instinct evolution, and any form of ML-based gate decision-making are OUT OF SCOPE. The gate is deterministic: first attempt → block; subsequent attempts → allow; opt-out via env var.

### Out of Scope — State file cleanup / GC

- The hook creates state files but does not delete them. Active cleanup (e.g., SessionEnd sweep of stale state files, age-based GC, cross-session deduplication) is a separate concern owned by a follow-up SPEC. This SPEC only requires that state files are small (single JSON line) and session-scoped (keyed by session_id), so the storage footprint is bounded by the number of unique (session_id, file_path) pairs in the project's history.

### Out of Scope — Go-side `internal/hook/fact_force.go`

- A Go-side handler (`moai hook fact-force` subcommand) following the `internal/hook/pre_tool.go` pattern is OPTIONAL and is NOT implemented in this SPEC. The shell-only implementation is sufficient for the Tier S scope. The Go handler is deferred to a follow-up SPEC IF measurement shows the shell implementation's latency or capability is insufficient.

### Out of Scope — Team-mode / TaskCompleted integration

- The hook is a PreToolUse gate, not a TeammateIdle or TaskCompleted gate. Team-mode AC verification is owned by `team-ac-verify.sh` (currently dormant) and is NOT touched.

## §F. Risks and Open Questions

### §F.1 Risk: false-positive friction on bulk-edit scenarios

The hook will block the first edit on EVERY file in a bulk-edit session (e.g., a 10-file codemod). This is by design (each file deserves at least one investigation), but it can be perceived as friction. The `MOAI_FACT_FORCE=off` opt-out (REQ-FF-004) is the documented escape hatch. The orchestrator MAY recommend setting this env var for the duration of a known bulk-edit task and unsetting it afterwards.

### §F.2 Risk: state-file path collision on shared hosts

If two developers share a checkout (unlikely but possible in pair-programming), their state files could collide if they happen to use the same session_id prefix. The `0o600` permission (REQ-FF-012) prevents cross-user read/write, and the session_id keying prevents cross-session contamination within one user, but the collision risk remains if two sessions happen to produce identical session_id+path hashes. Mitigation: the hash uses `shasum` (SHA-1, 40-char output), so the collision probability is negligible.

### §F.3 Open question: should the hook gate subagent-initiated edits?

REQ-FF-009 says yes — the gate applies in subagent context too. This is a deliberate choice: a subagent's edit is still an edit on a file the parent session may not have investigated. The state-keying by `session_id` (not `agent_id`) means the parent session's investigation counts for the subagent. Open question for the plan-auditor: is there a use case where a subagent should be exempted (e.g., a builder-harness agent generating template files that don't need investigation)? If yes, an `agent_type` exemption list could be added in a follow-up SPEC.

### §F.4 Open question: `shasum` vs `sha1sum` portability

macOS ships `shasum` by default; Linux typically ships `sha1sum`. The implementation MUST detect which is available. Alternative: substitute `/` → `_` in the absolute path and use the result as the filename (no hashing). This is simpler but produces longer filenames. The plan-auditor SHOULD evaluate which is more maintainable. The implementing agent SHOULD prefer whichever is shorter and requires no external dependency beyond coreutils.

## §G. Cross-References

- **ECC source pattern**: `https://raw.githubusercontent.com/affaan-m/ECC/main/hooks/hooks.json` (hook id `pre:edit-write:gateguard-fact-force`)
- **ECC observe.sh self-loop pattern**: `https://raw.githubusercontent.com/affaan-m/ECC/main/skills/continuous-learning-v2/hooks/observe.sh`
- **CLAUDE.md §7** — Rule 1 (Approach-First) + Rule 4 (Reproduction-First) doctrinal anchors
- **`agent-common-protocol.md` § File Operations Pattern** — "ALWAYS Read a file before using Edit" guideline that this hook mechanically enforces
- **`verification-claim-integrity.md` §1.1** — surface 1 (orchestrator self-report) + surface 2 (manager-agent completion report) preventive layer
- **`hooks-system.md` § Hook Configuration** + **§ Timeout Configuration** — settings.json hook JSON structure + 5s synchronous-default timeout policy
- **`hook-independence.md` § 3** — shared-failure-mode catalogue; this hook introduces a NEW state-file dependency (mode B candidate). The § 3 catalogue MUST be updated in the run phase (M1 step 6 of plan.md §F.1) of this SPEC to record the new shared condition (acceptance.md §E lists this as a run-phase exit criterion).
- **CLAUDE.local.md §2** — Template-First Rule
- **CLAUDE.local.md §7** — 5s hook timeout policy
- **C-HRA-008** — subagent boundary grep acceptance criterion
- **`session-handoff.md`** multi-session coordination policy — per-session state keying rationale
