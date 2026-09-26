---
id: SPEC-ROLE-NAMING-CODE-001
title: "Plan — role naming unification (code + CLI)"
version: "0.3.1"
created: 2026-09-26
---

# Plan — SPEC-ROLE-NAMING-CODE-001

## §A Context

Card t1256 (Tier L, class C), layer A of a two-card rename. Inputs: census `.moai/reports/t1256/census.md`, conflict map `.moai/reports/t1256/conflicts.md`, `research.md`, `design.md`, plan-audit iteration 1 `.moai/reports/t1256/plan-audit-iter1.md`. Development mode per `.moai/config/sections/quality.yaml`; the run phase uses TDD for the new rejection, stale-record, and legacy-retire behavior and characterization tests before touching the existing label parser, board write guard, SessionStart role writer, run-retire owner lookup, and broker slot matcher.

## §B Decisions for the Kickoff gate

Every item below is resolved. Sources: the operator's answers of 2026-09-26 relayed by the leader, and the operator's direct answers given in this lane window on 2026-09-26 (marked "operator, lane window"). Implementation Kickoff Approval is still required before any run-phase edit; it approves this plan, it no longer has to settle an open question.

- **O0 / Q1 — RESOLVED (operator).** `lane` is canonical. The `worker` and `agent` aliases are removed immediately: no compatibility aliases, no deprecation notice path. The reversal of the t1085 worker-canonical contract is approved, which removes the t1193 "unapproved reversal" ground. Consequence: REQ-RNC-003, -004, -005, -013, -020 are rejection requirements (non-zero exit, error naming the canonical form, nothing written). **Extension RESOLVED (operator, lane window):** `lead` / `lead-<suffix>` are rejected by the same rule (REQ-RNC-007).
- **O1 — RESOLVED (operator, lane window).** Environment variable names are kept with no second name (REQ-RNC-011, design D3): a single retained name is not an alias, and the dual-name window is excluded by O0. Values follow the new vocabulary for the leader-name and lane-label variables. The hard-rename reading, which would collide with card t1242 REQ-CFR-006/007/020 and the `env_vars` line in users' `.codex/config.toml`, is not taken.
- **O2 — RESOLVED (operator, lane window).** Persisted values: write-new, recognize-new-only, never rewrite (REQ-RNC-009, -010). Live legacy records are refused with retire-and-relaunch guidance: factory runs by REQ-RNC-022, legacy role declarations and the kanban registry by REQ-RNC-025, and the retire step itself made to work on a legacy-only run by REQ-RNC-024. Impact stated in design §2: a mid-run binary reinstall or a downgrade requires retiring and relaunching the run; old rows are not rewritten; card history keeps its recorded owner names (R4).
- **O3 — MOOT.** Legacy-spelling lifetime: the spellings are removed in this SPEC.
- **O4 — NOTED (leader act).** The leader edits card t1240's text in the queue to `-f lane` / `MOAI_FACTORY_ROLE=lane`. No SPEC action.
- **O5 — RESOLVED (operator/leader); mechanism fixed at v0.3.0.** t1245 lands first with value `worker`; this SPEC's run converts it to `lane` in bulk. The coupling mechanism is **one**: t1245's REQ-AP-013 equality assertion, which holds the marker value, the `-f` role token, and the lane-label prefix equal (REQ-RNC-012). This SPEC does not additionally require the three carriers to reference a shared definition — t1245 chose the assertion because `internal/config` cannot import `internal/cli`, and this SPEC keeps that choice. After conversion the guard recognizes `lane` only; `worker`/`agent` are not accepted. No fail-open window: the only stamper (t1240) lands after this SPEC.
- **O6 — RESOLVED (leader).** t1193 is removed from the ordering until the operator decides. REQ-RNC-014 no longer halts on it; the file overlap is a recorded risk (R3) and pre-flight re-checks t1193's status.
- **O7 — RESOLVED (operator, lane window).** Rename role-sense Go identifiers too, not only user strings (REQ-RNC-019, design D8, milestone M5). Env var name constants keep their names under O1 and are commented accordingly (REQ-RNC-019).
- **Q3 — RESOLVED (operator).** Lane self-dispatch is allowed up to promoting a queued card. Code layer: no enforcement change — the pick path (`moai todo next <n>`) carries no role guard today and gains none; the board write guard is about board columns and stays leader-only. The one code-layer change is the pick command's help text (REQ-RNC-021). The documented HARD promotion clause is handed to t1257.
- **Q4 — RESOLVED (operator).** `manager-lead` is kept (REQ-RNC-017, out of scope).
- **Q5 — RESOLVED (operator).** Homonyms of `leader` (CG-mode leader, Agent Teams lead) are disambiguated with qualifiers, not renamed (REQ-RNC-023).

**Ordering (leader, 2026-09-26):** t1242 → t1245 → t1256 run → t1240 → t1257 body replacement. t1193 is excluded pending the operator's decision.

## §C Pre-flight (run-phase entry)

1. Kickoff approval recorded. §B carries no open item; the approval confirms the plan as written.
2. REQ-RNC-014 record: develop SHA; t1242 deletion landed (halt t1242-file milestones if not); t1245 constants present (selects the M4 branch); t1193 state — if t1193 has landed on develop or is about to, report to the leader before editing any of the six overlap files (R3).
3. Re-read the two sibling SPECs this plan cites from their landed copies on develop and confirm the cited REQ numbers and wording still hold: SPEC-AUTONOMY-PRECONDITION-001 REQ-AP-011/-012/-013 (plan time: branch `WT-push-serialize-sign` @ `e0a471d1e`, status draft) and SPEC-CODEX-FACTORY-RETIRE-001 REQ-CFR-006/-007/-015/-020 (plan time: branch `WT-codex-factory-retire` @ `d2fc5cd69`, status draft). A changed number or wording is a blocker report to the leader before M1.
4. Absorb local develop into the card worktree; re-measure the census (`python3 .moai/reports/t1256/census.py`, run from the worktree root) and record the new counts next to the plan-time counts.
5. Measure and record the CG-mode / Agent Teams user-facing string population for AC-RNC-023, with the exact command AC-RNC-023 fixes.
6. Characterization tests pass on the absorbed tree for the packages the milestones touch (`go test ./internal/kanban/... ./internal/cli/... ./internal/hook/... ./internal/factorymsg/... ./internal/homestate/...` — scoped, never `./...`).

## §D Constraints

- Schema byte-identical (REQ-RNC-008); no data migration, no row rewrite.
- Codex `env_vars` allowlist and env var names unchanged (REQ-RNC-011) — also card t1242 REQ-CFR-020.
- No edit under `internal/template/templates/**` (t1257).
- Unrelated senses untouched (REQ-RNC-016); `manager-lead` untouched (REQ-RNC-017); homonyms qualified, not renamed (REQ-RNC-023); kanban companion roles untouched (REQ-RNC-001).
- Lane-local verification only; the full suite runs in CI on the develop push.

## §E Self-verification

Each milestone closes with: scoped `go test` on the touched packages, `go vet` on them, `golangci-lint run` on them, and the AC rows for that milestone recorded in progress.md §E.2 with command and verbatim output.

**Remeasure scope after every develop absorption and before the merge-window request** — the merge tree, not the pre-absorb tree: `go test ./internal/kanban/... ./internal/cli/... ./internal/hook/... ./internal/factorymsg/... ./internal/homestate/... ./internal/web/... ./internal/config/... ./internal/codexwiring/... ./internal/spec/...`. `./internal/spec` is in scope because this card edits a depth-1 `acceptance.md`: the AC-count corpus test (`TestACCounterFullCorpusMatchesBaseline`) and the commit-time snapshot guard read it, and an absorbed develop can change the snapshot they compare against. Any acceptance.md edit during run (a D-NEW-1 re-delegation) regenerates the snapshot in the same commit if this SPEC is by then recorded in it.

## §F Milestones (ordered by change likelihood — highest first)

### M1 — Persisted vocabulary and the run boundary (Priority High)
The decisions with the widest blast radius (O2 run boundary, legacy retire, the SessionStart writer) land first, as tests before code.
- Writers emit `leader` / `lane` / `lane-<n>` for role, slot, label, registry key, and card owner (REQ-RNC-010).
- Readers recognize only the new vocabulary (REQ-RNC-009): board write guard, SessionStart role reader, run-retire owner lookup, broker slot matcher, factory message hook, web view model, `todo` card-owner reader.
- **SessionStart role-declaration writer** (`internal/hook/session_start_record.go`: `kanbanRoleFromEnv` re-derives the role from the environment on every fire, and `kanban.WriteBestEffort` writes it): when the session's launch label or existing session record carries a legacy value, leave the record byte-identical and emit the stale-run notice instead (REQ-RNC-025). The session record (`internal/kanban/record.go`) is not the board role declaration (`internal/kanban/role.go`) the board write guard reads; both are covered. Without this, the first SessionStart after a reinstall silently adopts an old leader session as `leader`.
- Factory run boundary: live legacy records of the same run refuse the launch or join with the retire-and-relaunch message; dead ones are stale; history displays verbatim (REQ-RNC-022). Run membership is read from the run id in the factory database and the run's own broker database.
- Kanban registry and legacy declarations: a live `lead` entry in `leads.json` produces one notice and the launch proceeds under `leader`; a legacy `lead` board role declaration loses board write access (REQ-RNC-025).
- Legacy-only run retire: the run-retire owner lookup's legacy-row fallback (`internal/factorymsg/factory_run_retire.go` `LeadPeerIdentity`, consulted from `internal/homestate/factory_run_retire.go:181-194` when `lead_pid = 0`) reads a `lead` peer's process identity as identity evidence, so the retire step the refusal names works on the runs it names (REQ-RNC-024).
- Broker delivery for `leader` / `lane-<n>`; legacy `to_slot` refused (REQ-RNC-013).
- Schema-freeze guard (REQ-RNC-008) and env-name/allowlist freeze guard with new values (REQ-RNC-011).
- ACs: AC-RNC-006, -008, -009, -010, -011, -022, -024, -025.

### M2 — CLI surface: tokens, labels, help, rejection (Priority High)
- `lane-<n>` produced; `worker-<n>`/`agent-<n>` refused on every input path, any letter case (REQ-RNC-004, -005).
- `-f lane` canonical; `-f worker`/`-f agent` refused (REQ-RNC-002, -003); `lead`/`lead-<suffix>` refused (REQ-RNC-007).
- Usage error, `Use`/help strings for `cc`/`glm` (and `codex` if still present after t1242), collision messages, `moai tokens --role` example, `gtd answer` short text, board write-guard error text (REQ-RNC-001).
- `todo next` help text for lane self-dispatch (REQ-RNC-021).
- ACs: AC-RNC-001, -002, -003, -004, -005, -007, -016, -021.

### M3 — Notices, locales, dashboard, homonyms (Priority Medium)
- Factory and kanban SessionStart notices in en/ko/ja/zh per design §3 (REQ-RNC-015), including the stale-run notice.
- Doctor, statusline, and web display labels read `leader`/`lane`; legacy chains render `legacy run: relaunch required` (REQ-RNC-001).
- CG-mode / Agent Teams strings carry their qualifier (REQ-RNC-023).
- ACs: AC-RNC-012, -013, -023.

### M4 — Factory role marker value (Priority Medium, conditional)
- Where t1245 has landed (expected under the ordering): flip the value constant to `lane`, keep the guard on the constant so it recognizes `lane` only, confirm t1245's REQ-AP-013 equality assertion passes with marker value, `-f` token, and lane-label prefix all `lane`, and record a red run of the assertion under a one-carrier mutation (REQ-RNC-012).
- Where t1245 has not landed: record the value `lane` in progress.md for the leader to hand to t1245; no edit.
- ACs: AC-RNC-014.

### M5 — Mechanical rename and guards (Priority Low)
- Role-sense Go identifiers and comments renamed (REQ-RNC-019, operator O7). Retained env name constants commented "name kept under REQ-RNC-011; value follows leader/lane"; allowlisted legacy-value literals commented refuse/detect-only.
- Test fixtures to the new vocabulary, with one rejection test per legacy spelling (REQ-RNC-020).
- Vocabulary guard test with word-boundary `lead` matching, `MOAI_*` token exclusion, and file:line-bound allowlist entries (REQ-RNC-018); unrelated-sense diff check (REQ-RNC-016, -017).
- ACs: AC-RNC-015, -017, -018, -019, -020.

### M6 — Sync-phase lifecycle records (Priority Low)
- Record the partial supersession on the reversed contract: add `partially_superseded_by: [SPEC-ROLE-NAMING-CODE-001]` to SPEC-FACTORY-WORKER-NAMING-001's frontmatter; its `status: completed` stays. This is a non-transition frontmatter correction owned by manager-spec, reached by orchestrator re-delegation during sync (manager-docs may not edit that SPEC's frontmatter beyond `status`/`updated`).
- Report to the leader that t1245's REQ-AP-012 text still names the value `worker`; amending t1245's SPEC is t1245's, not this card's.
- No AC (lifecycle bookkeeping); recorded in progress.md §E.4.

## §G Risks

| # | Risk | Mitigation |
|---|---|---|
| R1 | The leader role value moves from `lead` to `leader` in the board write guard, the role reader, the SessionStart role writer, and the run-retire owner lookup at once; a missed SessionStart reader or writer silently adopts or mislabels old sessions, and a missed run-retire lookup leaves a run unretirable. The board write guard is unwired in production today (measured at `5102a69e9`, production `.go` files only: `DeclareRole` 0 callers; its reader `ResolveDeclaredRole` is called only at `internal/kanban/board_store.go:191` inside `requireLeadRole`; `WriteBoardState` is called only at `internal/kanban/board_store.go:363` inside `TransitionIntoRunOpts`, which is called only at `internal/kanban/board_store.go:344` inside `TransitionIntoRun`, which has 0 callers; `RecoverBoard` 0 callers), so a missed board-guard reader has no production symptom and is covered at package level | M1 first; characterization tests on the board guard, role reader, and SessionStart writer before the change; AC-RNC-008 exercises a test-written `leader` declaration through the board guard and the other readers, AC-RNC-025 proves the SessionStart writer leaves a legacy session record byte-identical |
| R2 | Stamp/guard value drift on the role marker → guard silently denies nothing | REQ-RNC-012 binds the value through t1245's equality assertion plus constant-only stamp/compare sites; AC-RNC-014 asserts the `lane` deny, a red run of the assertion, and names every stamp/compare site |
| R3 | t1193 overlaps six files (`internal/cli/factory.go`, `internal/factorymsg/store.go`, `internal/hook/factory_messages.go`, `internal/hook/session_start_factory.go`, `internal/cli/codex_launcher.go`, `internal/codexwiring/configtoml.go`) and adds 49 lines of worker/lead vocabulary; it is out of the ordering pending the operator's decision | Pre-flight re-checks t1193's status (§C.2, AC-RNC-015); if it lands first, this run absorbs and converts its vocabulary; if it is live and unlanded, report to the leader before editing the overlap files |
| R4 | Mid-run binary reinstall — a routine act in this repository — now breaks a running factory or kanban run until it is relaunched (REQ-RNC-022, -025, O2) | The refusal names the run and the relaunch step; REQ-RNC-024 makes the named retire step work on a legacy-only run; the leader schedules the develop push that carries this change between runs |
| R5 | Guard test allowlist grows into a loophole | Allowlist entries limited to refuse/detect literals, their error text, and non-role senses, each bound to a file:line the guard re-checks (REQ-RNC-018); AC-RNC-018 records the count and fails on a stale entry |
| R6 | Heuristic census undercounts multi-line user-facing strings | Pre-flight re-measure plus the guard test, which reads compiled string constants rather than lines |
| R7 | Operator retraining churn (third change in five days) with no hint path | Operator chose rejection knowingly (O0); every rejection names the canonical form |
| R8 | Card t1240 starts before its card text is amended (O4) and ships `-f agent` | O4 is the leader's queue edit; the ordering places t1240 after this run, whose rejection path would refuse `-f agent` at t1240's first test |
| R9 | A legacy-only run whose owner cannot be identified stays unretirable, so the refusal message points at a step that fails | REQ-RNC-024 reads the legacy peer identity for classification only; AC-RNC-024 proves both the dead-owner retire and the live-owner refusal |

## §H Cross-references

- `.moai/reports/t1256/census.md`, `.moai/reports/t1256/conflicts.md`, `.moai/reports/t1256/plan-audit-iter1.md`
- SPEC-FACTORY-WORKER-NAMING-001 (the contract this partially supersedes; M6 records it)
- SPEC-CODEX-FACTORY-RETIRE-001 (t1242) — at plan time only on branch `WT-codex-factory-retire` @ `d2fc5cd69` (draft); re-checked at pre-flight §C.3
- SPEC-AUTONOMY-PRECONDITION-001 (t1245) — at plan time only on branch `WT-push-serialize-sign` @ `e0a471d1e` (draft); re-checked at pre-flight §C.3
- Sibling card t1257 consumes `design.md` §3 and owns the promotion-clause amendment.
