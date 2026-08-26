# Acceptance Criteria — SPEC-TEAMMATE-REVIVAL-GUARD-001

> Verification layer. Every entry is `AC-TRG-NNN` labeled Given/When/Then and binary-testable.
> Two-cell adoption discipline (`verification-completeness.md` §2): every AC carries a RED-now cell pinned to tree `c9eed8ac6` (measured 2026-08-26) and a green path naming the flipping milestone.

## §A AC Design Notes

1. **RED baseline provenance** — RED cells cite the §B measurements in plan.md (B1: no `SendMessage|TaskStop` matcher in `.claude/settings.json`; B2: `grep -rn "TaskStop\|SendMessage" internal/ --include="*.go" -l` → 0 files). Both measured on `c9eed8ac6` before any implementation.
2. **RED reason** — every AC is red for the same stated reason: the guard, its wiring, and its registry do not exist. No AC is red due to unrelated pre-existing files, and each green path runs through code this SPEC adds (no "someone fixes unrelated files" paths).
3. **Mutant guard** — AC-TRG-003/004 pin the fail-open direction and its mirror (a guard that denies nothing satisfies nothing; a guard that denies live teammates fails AC-TRG-004b). The human-discipline mutant (broadcast instead of mechanism) satisfies no AC — there is nothing to run.
4. **Live-vs-unit split** — unit ACs (001–008 and 011) run on synthetic `HookInput` fixtures with `t.TempDir()` registries; AC-TRG-010 is the live-session gate (E-P1) because mid-session probing was inconclusive (plan.md §C.1-E7).

## §B AC Matrix

### AC-TRG-001 — TaskStop completion records a stop (REQ-TRG-001, REQ-TRG-002) · P1 · M1

- **Given** a synthetic PostToolUse input `{tool_name: "TaskStop", tool_input: {…target…}, session_id: S}` naming teammate X,
- **When** the stop-guard recorder runs,
- **Then** `.moai/state/agent-stops/S.json` contains `{name: X, agent_id, stopped_at}` AND `.moai/logs/agent-stop-audit.jsonl` gains exactly one `stop_recorded` row.
- **RED-now** (`c9eed8ac6`): no handler exists (B2) → neither file is produced.
- **Green path**: M1 — unit test constructs the fixture, runs the handler against a `t.TempDir()` root, asserts both artifacts.

### AC-TRG-002 — Send to a stopped name is denied with sentinel + audit (REQ-TRG-003) · P1 · M2

- **Given** the registry holds a live entry for name X in session S and `workflow.agent_stop_guard.enabled = true`,
- **When** a PreToolUse input `{tool_name: "SendMessage", tool_input: {to: X, …}, session_id: S}` reaches the guard,
- **Then** the decision is deny, the reason starts with `STOPPED_TEAMMATE_VIOLATION:`, names X, and names the orchestrator route; a `send_denied` audit row is appended.
- **RED-now**: no matcher routes SendMessage to any handler (B1) → no deny exists anywhere in the tree.
- **Green path**: M2 unit test with registry fixture + gate enabled.

### AC-TRG-003 — Uncertain evidence never denies (REQ-TRG-004) · P1 · M2

- **Given** the registry holds a live entry for X, and a PreToolUse SendMessage input that is (a) malformed JSON, (b) missing `to`, or (c) paired with an unreadable registry file,
- **When** the guard evaluates each variant,
- **Then** every variant allows (no deny), and each appends an observe-only audit row.
- **RED-now**: vacuous-true guard absent — with no guard, nothing denies, but nothing observes either; the AC's observable (audit rows) does not exist (B2).
- **Green path**: M2 table-driven tests over the three variants.

### AC-TRG-004 — Live teammate sends unaffected (mirror mutant) (REQ-TRG-003, REQ-TRG-004) · P1 · M2

- **Given** the registry holds an entry for X only, and the gate is enabled,
- **When** SendMessage inputs addressed to any other name Y (live teammate), to `main`, or carrying `notify_when_idle` to Y,
- **Then** all allow; audit rows record `send_observed`.
- **RED-now**: no guard → no audit rows exist (B2).
- **Green path**: M2 negative tests — pins the over-blocking mutation.

### AC-TRG-005 — Fresh spawn with the same name clears the entry (REQ-TRG-005) · P1 · M2

- **Given** the registry holds a live entry for name X,
- **When** a PreToolUse `{tool_name: "Task", tool_input: {subagent_type: …, name: X, …}}` reaches the guard,
- **Then** the entry for X is removed BEFORE the spawn proceeds, and a `respawn_cleared` audit row is appended; a subsequent SendMessage to X allows.
- **RED-now**: `extractAgentSpawn` (`agent_model_guard.go:87-102`) reads only `subagent_type`/`model` — no name-based clearing exists (B2/B4).
- **Green path**: M2 — extend spawn parsing with `Name`; unit test asserts registry state + follow-up send.

### AC-TRG-006 — Session end clears the session's entries (REQ-TRG-006) · P2 · M2

- **Given** `.moai/state/agent-stops/S.json` exists with ≥1 entry,
- **When** the SessionEnd handler for S completes,
- **Then** the file is removed (or emptied) and a `session_cleared` audit row is appended.
- **RED-now**: no such cleanup path exists (B2).
- **Green path**: M2 unit test invoking the session-end hook path with a `t.TempDir()` root.

### AC-TRG-007 — Gate off ⇒ observe + advise, never deny (REQ-TRG-007) · P1 · M2

- **Given** the registry holds a live entry for X and `workflow.agent_stop_guard.enabled = false` (the resolved shipped default — plan.md §C.3 decision: false; value unchanged),
- **When** a SendMessage to X is evaluated,
- **Then** no deny is emitted; an advisory surfaces; the audit row records the would-have-denied state (`send_observed` + advisory flag).
- **RED-now**: no gate, no guard, no rows (B1/B2).
- **Green path**: M2 test with the config default; config-default value asserted in `internal/config/defaults.go` test.

### AC-TRG-008 — Audit row shape is complete and exactly-once (REQ-TRG-001) · P1 · M1

- **Given** any guard-observed event (stop, send, deny, respawn-clear, session-clear),
- **When** the audit append runs,
- **Then** each produces exactly one JSONL row carrying all of `{timestamp(UTC RFC3339), session_id, kind, name, agent_id, decision}`; repeated TaskStop of the same name upserts the registry without duplicating audit confusion (registry idempotent; audit appends one `stop_recorded` per observed completion).
- **RED-now**: `.moai/logs/agent-stop-audit.jsonl` does not exist; no producer (B2).
- **Green path**: M1 shape tests over all event kinds.

### AC-TRG-009 — Wiring + template mirror + neutrality (REQ-TRG-008) · P2 · M1 (matcher twins), M2 (config)

- **Given** the implemented tree,
- **When** checking `.claude/settings.json` and `internal/template/templates/.claude/settings.json.tmpl` (the template twin is the `.tmpl` source; no plain settings.json exists there — auditor-measured 2026-08-26),
- **Then** both carry a PreToolUse matcher entry `SendMessage|TaskStop` routed to the existing pre-tool wrapper; the config default exists in `internal/config/defaults.go`; template twins contain no SPEC IDs / REQ tokens / commit SHAs / internal dates (neutrality greps 0 hits); `make build` regenerates without diff noise; `internal_content_leak_test.go` + `template_neutrality_audit_test.go` pass.
- **RED-now**: no such matcher in either tree (B1 — local tree measured; template twin assumed identical pending M1 mirror verification, which is itself part of this AC's green path).
- **Green path**: M1/M2 mirror + `make build` + CI neutrality guards.

### AC-TRG-010 — Live-session firing gate (all REQs, E4 parity) · P1 · M2

- **Given** a fresh session started after the wiring ships (hooks active from launch, not mid-session),
- **When** the E-P1 recipe runs (spawn → TaskStop → registry check → send deny → teammate-issued send deny → same-name respawn → clear → send allow),
- **Then** every step's expected observation holds (registry entry, sentinel deny from both lead- and teammate-issued sends, respawn clear).
- **RED-now**: cannot be red-measured pre-implementation (requires the wiring); its pre-state is plan.md §C.1-E7's inconclusive probe — the AC exists to force the live verification, substituting for the probe that could not run mid-session.
- **Green path**: M2 — full-recipe execution recorded in progress.md §E.2 (M3 re-runs it as sustained dogfood for the default-flip trigger).

### AC-TRG-011 — `name [ref]` addressing matched both directions (REQ-TRG-003, REQ-TRG-004) · P1 · M2

- **Given** the registry holds a live entry for stopped name X (gate enabled) and Y is a live teammate,
- **When** SendMessage inputs arrive addressed `X [ref]` and `Y [ref]` (the sanctioned disambiguated addressing form),
- **Then** the `X [ref]` send is DENIED (the optional `[ref]` suffix parsed and stripped, base name matched against the registry) and the `Y [ref]` send is ALLOWED with a `send_observed` audit row; unparseable suffix forms stay fail-open (AC-TRG-003).
- **RED-now** (`c9eed8ac6`): no guard parses any recipient (B1/B2).
- **Green path**: M2 table-driven tests over both directions.

## §C Severity classification

- **P1 (blocker)**: AC-TRG-001, 002, 003, 004, 005, 007, 008, 010, 011 — the deny mechanism, its fail-open boundary (both addressing forms), and live firing.
- **P2 (required for close)**: AC-TRG-006, 009 — lifecycle hygiene and distribution hygiene.

## §D Traceability

See spec.md §3 (REQ ↔ AC ↔ milestone matrix). Every REQ-TRG-001..008 maps to ≥1 AC; no AC floats free of a REQ.

## §E Indirect verification

- Neutrality/mirror guards (`internal_content_leak_test.go`, `template_neutrality_audit_test.go`, CI `template-neutrality-check.yaml`) cover REQ-TRG-008 without a bespoke test.
- Existing hook test suites (`agent_model_guard_test.go`, `branch_guard_test.go`) demonstrate the fixture style M1/M2 tests follow.

## §F Closure gates (Definition of Done)

1. All P1 ACs green with cited test output; P2 green or explicitly deferred with reason.
2. E-P1 live recipe executed once with results recorded in progress.md §E.2 (AC-TRG-010).
3. Affected-package suites green (`go test ./internal/hook/... ./internal/config/...`), pushed, CI full-suite green.
4. Template mirror + `make build` + neutrality 0/0/0.
5. Enforcement default recorded as resolved — **false** (operator decision 2026-08-26, orchestrator AskUserQuestion round) — in plan.md §C.3 and progress.md §E.1; the M3 default-flip trigger remains open as an evidence-gated follow-up, not an open clarification.

## §G Forward-looking checks

- After M3 dogfood: record the default-flip verdict (template default false → true?) against the zero-false-positive evidence, as a named upgrade trigger.
- Revisit the deferred write-path quarantine (spec.md §Out of Scope) once `agent_id` stability across resume is observed in the M3 record.
- Watch upstream: if the runtime ever adds a stop-that-reclaims or a settings knob for auto-resume, this guard's deny layer becomes redundant — prefer the runtime knob and retire the guard (upstream-preference precedent).
