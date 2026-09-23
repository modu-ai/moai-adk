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

## §C Decision record (clarifications resolved 2026-09-23)

The iter-1 plan listed six open clarifications (Q1–Q6). All six are resolved; none remains open.

| # | Question | Decision | Source |
|---|---|---|---|
| Q1 | AC-POL-01 closure scope | Whole catalog. Every obligation is registered, including M1 standing-policy obligations. M1 rows are marked `blocked:M1` and reported honestly as FAIL until M1 lands (design.md §D4). | Lead via Jev, standing operator delegation 2026-09-23 (confidence 0.62) |
| Q2 | `needs_input` on Codex | Fail-closed deny, surfaced visibly: the deny reason reaches the model and the conversion is written to the adapter's discard record. No silent allow (design.md §D5). | Lead via Jev, standing operator delegation (confidence 0.78) |
| Q3 | Goal cancellation status | Add a new `cancelled` status; do not reuse `cleared`. Every non-test reader of goal status handles it explicitly; an unrecognised status surfaces a diagnostic (design.md §D6, AC-HPR-013). | Operator, directly (2026-09-23) |
| Q4 | Sync-gate port shape (shim over a Go entry vs parallel Go implementation) | Run-phase decision inside M2d, recorded in progress.md with the AC-HPR-002 goldens as the equivalence proof (design.md §D2). Not a pre-kickoff question (plan-audit iter-1 D9). | Coordinator instruction following plan-audit iter-1 D9 |
| Q5 | Live-run budget and credentials | No live runs in this SPEC. Live tests are built, opt-in, and unexecuted; every live leg records `NOT_RUN`, which is not PASS. The SPEC closes as partial (live-uncertified); live certification goes to a separate follow-up card (spec.md §E). | Operator, directly (2026-09-23) |
| Q6 | SPEC-CODEX-HOOK-ADAPTER-001 REQ-7 (`internal/hook` invariance) | Keep. Nothing under `internal/hook` changes (design.md §D7). | Lead via Jev, standing operator delegation (confidence 0.79) |

The Jev-sourced decisions (Q1, Q2, Q6) carry the delegation's confidence as recorded. They are
decisions, not measurements. The operator can reopen any of them at Implementation Kickoff
Approval.

## §D Constraints

- No push, no branch switch, and no edit of any other worktree. Integration follows the local develop chain via the lead.
- Local verification is scoped to changed packages (`internal/codexadapter`, `internal/codexwiring`, `internal/cli` with `-run` filters, `internal/goal`, `internal/verify`, registry package). CI gives the full-suite verdict.
- Template-first: any change under `internal/template/templates/` is followed by `make build`; agent definitions are not touched.
- No long check inside a hook (REQ-HPR-018).

## §E Self-verification (plan phase)

- SPEC ID regex: PASS (executed; see progress.md §E.1).
- 25 REQs (Tier L ceiling 25) and 21 ACs (ceiling 25) as list items (`- **REQ-HPR-NNN**`) so the lint collector sees them.
- Each REQ maps to ≥1 AC (acceptance.md §C). Each AC names a command and a mutation.

## §F Milestones

Ordered by decision reversibility. Data-model and interface decisions come first, mechanical work
last. Priorities are labels only.

| Milestone | Priority | Content | ACs | Depends on |
|---|---|---|---|---|
| **M2a — decision & state models** | High | Normalized decision type and per-event translation table (design.md §D1). New goal status `cancelled` (Q3) and its handling in every reader of goal status listed in design.md §D6: `schema.go` set, `evaluate.go:294` early return, the writers at `evaluate.go:306/325/340/404/435`, `launcher_blockcap_infinite.go:68`, `handoff.go:85`, `goal.go:1282/1331`, `dashboard.go:141`, `hook_stop_goal.go` emission; unknown-status diagnostic. Obligation registry schema and loader (design.md §D4, Q1) | AC-HPR-006 (table), AC-HPR-013 (status consumers), AC-HPR-018 (schema) | — |
| **M2b — receipt contract and Stop budget** | High | Extend the `internal/verify` snapshot record and the sync gate's outcome record with config digest, command, and tool version; comparison predicate; not-run on any mismatch (design.md §D3.6). Declare per-member internal budgets and the `chain_overhead` constant; `T_stop` stays 10 s (Q5) | AC-HPR-016, AC-HPR-017 | M2a |
| **M2c — decision hardening** | High | Replace the `ask`/`defer` drop with the fail-closed visible deny (Q2). Fault-injection tests for timeout, exit 1, corrupt output, exit 2. **Intentional characterization amendment:** `TestPreToolUseAskDropped` (`output_test.go:278`) is inverted, not kept green | AC-HPR-006, AC-HPR-008 (unit) | M2a |
| **M2d — Stop chain on Codex** | High | Inventory over both template renders (REQ-HPR-001). Go chain runner with the merge rule (design.md §D2). Lookup-only goal evaluation and receipt placement per design.md §D3.3. **Decide the sync-gate port shape here (Q4)** and record it in progress.md. Goldens for Claude-vs-Codex effect, including the `unmeasured` mapping | AC-HPR-001, 002, 003, 005 | M2a, M2b |
| **M2e — event adaptation** | High | Adapt PreCompact, PostCompact, PermissionRequest; add the Interrupt path in `internal/cli` (design.md §D7, Q6); update the `events.go` census comment. **Intentional characterization amendment:** `TestAdaptedRowCount` (`events_test.go:67`, `wantAdapted = 8`) is raised to the new count, not kept green | AC-HPR-009, 010, 011 (golden/unit legs) | M2a |
| **M2f — goal parity** | High | Codex Stop → existing evaluator in lookup-only mode; cancellation precedence through the real producers; budget termination; host override ≠ satisfied | AC-HPR-012 (golden), 013, 014, 015 | M2a, M2d, M2e |
| **M2g — live legs, built but not run** | Medium | `MOAI_PARITY_LIVE` axis declared in `codex_live_axis_declaration_test.go`; temp `CODEX_HOME`; scratch-project `claude -p`; `t.Skipf` discipline (acceptance.md §B rule 8); the Stop-timeout ceiling probe. **Not executed** (Q5): each live leg is run once without the switch and its `NOT_RUN` skip output is recorded | AC-HPR-004, 007, live legs of 008–012, 020, 021 — all `NOT_RUN` | M2c–M2f |
| **M2h — coverage & verdict aggregation** | Medium | Populate the registry with the whole catalog (Q1): M2 obligations, M1 rows `blocked:M1`, the Claude-interrupt row `UNSUPPORTED`, and the non-Stop chains `unverified`. Coverage check; aggregate verdict reading both the go-test action and the verdict record | AC-HPR-018, 019 | M2a–M2g |
| **M2i — mechanical tail** | Low | `RenderHooks` table update, `make build`, doc comments, codemaps refresh. The two amended tests above are intentional; every other existing codexwiring/codexadapter test stays green | regression of the unamended tests | M2d, M2e |

## §G Risks

| Risk | Mitigation |
|---|---|
| Codex cannot trigger compaction or permission requests non-interactively (as on 0.153.4) | REQ-HPR-013: record `NOT_RUN` with the trigger attempted; aggregate stays FAIL; no "does not fire" claim |
| Host resolves hook faults as allow (H2) | Record `UNSUPPORTED`; mitigate at pre-tool with fail-closed default where MoAI controls output; surface in verdict |
| Sync-gate port changes Claude behaviour | AC-HPR-002 goldens run both paths on the same inputs; Claude regression tests stay |
| Receipt digest misses an input the check reads | config_digest scope is declared per check; mutation per field in AC-HPR-017 |
| Live behaviour differs from the goldens | Q5: no live run here, so this risk is carried open; the SPEC closes as partial (live-uncertified) and the follow-up card runs the live legs |
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
