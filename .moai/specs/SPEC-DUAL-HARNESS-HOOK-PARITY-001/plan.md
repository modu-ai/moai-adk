# Plan — SPEC-DUAL-HARNESS-HOOK-PARITY-001

## §A Context

Card t1099, Class C, Tier L. M2 of the dual-harness design: hook chain, approval-decision
preservation, goal continuation, and obligation coverage. Sibling card t1100 owns M3/M4/M5 and MCP.
Worktree: `.claude/worktrees/dual-harness-parity-rebuild`, branch `WT-dual-harness-parity-rebuild`.

## §B Known issues at plan time (measured, research.md §R1)

- Codex Stop runs one handler that returns allow except for factory continuation (R1.3).
- 4 of 12 Codex events unadapted (R1.1).
- PreToolUse `ask` dropped to no-opinion on Codex (R1.4). Whether that resolves to allow is unmeasured.
- Codex handler timeout 10 s vs goal-condition budget 90 s (R1.5).
- No obligation registry (R1.6); installed codex-cli 0.155.1 differs from the 0.153.4 basis (R1.7).

## §C Clarifications required before Implementation Kickoff

- [NEEDS CLARIFICATION: Q1 AC-POL-01 closure scope] Does AC-POL-01 close over the full required-obligation catalog, including M1 standing-policy delivery? If yes, AC-POL-01 stays FAIL until M1 lands. The alternative is to close over only the M2 obligations this SPEC seeds. Proposed: full catalog registered, M1 rows marked `blocked:M1`, aggregate reported honestly as FAIL for them.
- [NEEDS CLARIFICATION: Q2 needs_input on Codex] Fail-closed deny with a reason (design.md §D5, proposed), or keep the host approval flow wherever measurement shows Codex prompts under every supported policy?
- [NEEDS CLARIFICATION: Q3 cancellation status] Add a new goal status `cancelled` (proposed, design.md §D6), or reuse `cleared`?
- [NEEDS CLARIFICATION: Q4 sync-gate on Codex] Port the sync gate's decision core to a Go entry that both harnesses call (proposed), or run a parallel Go implementation with the shell script kept for Claude? The first touches a 778-line distributed script.
- [NEEDS CLARIFICATION: Q5 live-run authorization] Live AC-HPR-004/007/008/009/010/011/020 spend real model turns in Claude Code and Codex. What is the approved per-run turn budget and credential source? Live runs are never executed without this answer; until then those ACs record NOT_RUN.
- [NEEDS CLARIFICATION: Q6 REQ-7 amendment] Keep SPEC-CODEX-HOOK-ADAPTER-001 REQ-7 (no `internal/hook` change; proposed), or allow extracting shared Stop-member functions into `internal/hook`?

## §D Constraints

- No push, no branch switch, and no edit of any other worktree. Integration follows the local develop chain via the lead.
- Local verification is scoped to changed packages (`internal/codexadapter`, `internal/codexwiring`, `internal/cli` with `-run` filters, `internal/goal`, `internal/verify`, registry package). CI gives the full-suite verdict.
- Template-first: any change under `internal/template/templates/` is followed by `make build`; agent definitions are not touched.
- No long check inside a hook (REQ-HPR-018).

## §E Self-verification (plan phase)

- SPEC ID regex: PASS (executed; see progress.md §E.1).
- 25 REQs (Tier L ceiling 25) and 20 ACs as list items (`- **REQ-HPR-NNN**`) so the lint collector sees them.
- Each REQ maps to ≥1 AC (acceptance.md §C). Each AC names a command and a mutation.

## §F Milestones

Ordered by decision reversibility. Data-model and interface decisions come first, mechanical work
last. Priorities are labels only.

| Milestone | Priority | Content | ACs | Depends on |
|---|---|---|---|---|
| **M2a — decision & state models** | High | Normalized decision type and per-event translation table (§D1); goal `cancelled` status (§D6, Q3); obligation registry schema + loader (§D4, Q1) | AC-HPR-006 (table), AC-HPR-018 (schema) | Q1, Q3 |
| **M2b — receipt contract** | High | Extend the `internal/verify` snapshot record with config digest, command, and tool version; comparison predicate; not-run on any mismatch (§D3) | AC-HPR-016, AC-HPR-017 | M2a |
| **M2c — decision hardening** | High | Replace the `ask`/`defer` drop with the §D5 translation (Q2); fault-injection tests for timeout, exit 1, corrupt output, exit 2 | AC-HPR-006, AC-HPR-008 (unit) | M2a, Q2 |
| **M2d — Stop chain on Codex** | High | Inventory (REQ-001); Go chain runner with merge rule (§D2); sync-gate decision core reachable without `.claude/` (Q4); goldens for Claude vs Codex effect | AC-HPR-001, 002, 003, 005 | M2a, M2b, Q4 |
| **M2e — event adaptation** | High | Adapt PreCompact, PostCompact, PermissionRequest; add the Interrupt path in `internal/cli` (§D7); update `events.go` census comment | AC-HPR-009, 010, 011 (golden) | M2a, Q6 |
| **M2f — goal parity** | High | Codex Stop → existing evaluator; cancellation precedence; budget termination; host override ≠ satisfied | AC-HPR-012, 013, 014, 015 | M2a, M2d, M2e |
| **M2g — live campaign** | Medium | `MOAI_PARITY_LIVE` axis; temp `CODEX_HOME`; `claude -p` scratch project; needs_input policy sweep; fault outcomes; compaction/permission/interrupt triggers | AC-HPR-004, 007, 008 (live), 009–011 (live), 020 | M2c–M2f, Q5 |
| **M2h — coverage & verdict aggregation** | Medium | Populate the registry with M2 obligations (and M1 rows per Q1); coverage check; aggregate verdict with attribution; rule-P reader | AC-HPR-018, 019 | M2a–M2g |
| **M2i — mechanical tail** | Low | `RenderHooks` table update, `make build`, doc comments, codemaps refresh | regression of existing codexwiring/codexadapter tests | M2d, M2e |

## §G Risks

| Risk | Mitigation |
|---|---|
| Codex cannot trigger compaction or permission requests non-interactively (as on 0.153.4) | REQ-HPR-013: record `NOT_RUN` with the trigger attempted; aggregate stays FAIL; no "does not fire" claim |
| Host resolves hook faults as allow (H2) | Record `UNSUPPORTED`; mitigate at pre-tool with fail-closed default where MoAI controls output; surface in verdict |
| Sync-gate port changes Claude behaviour | AC-HPR-002 goldens run both paths on the same inputs; Claude regression tests stay |
| Receipt digest misses an input the check reads | config_digest scope is declared per check; mutation per field in AC-HPR-017 |
| Live runs cost real turns | Q5 budget; the live axis is opt-in; unit/golden ACs never depend on live runs |
| Scope creep into t1100 (permissions, kanban, rollback) | spec.md §F exclusions; plan audit checks the diff against them |

## §H Anti-patterns to avoid

- Counting `ok` / exit 0 as PASS when the named test skipped.
- Declaring an event "does not fire" because a trigger was not achieved.
- Running a long check inside a Codex hook and raising the timeout to fit it.
- Claiming full parity from this SPEC while t1100 and M1/M5 remain open.

## §I Cross-references

- `reports/moai-dual-harness-full-design-20260922.md` §07–§09, §17–§19
- `reports/moai-dual-harness-implementation-status-20260923.md`
- `reports/moai-dual-harness-handoff-20260923.md`
- SPEC-CODEX-HOOK-ADAPTER-001, SPEC-CODEX-EVENT-COVERAGE-001, SPEC-CODEX-DUAL-AGENTS-001, SPEC-CODEX-WIRING-001
