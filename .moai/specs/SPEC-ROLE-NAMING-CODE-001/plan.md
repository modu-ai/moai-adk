---
id: SPEC-ROLE-NAMING-CODE-001
title: "Plan — role naming unification (code + CLI)"
version: "0.2.0"
created: 2026-09-26
---

# Plan — SPEC-ROLE-NAMING-CODE-001

## §A Context

Card t1256 (Tier L, class C), layer A of a two-card rename. Inputs: census `.moai/reports/t1256/census.md`, conflict map `.moai/reports/t1256/conflicts.md`, `research.md`, `design.md`. Development mode per `.moai/config/sections/quality.yaml`; the run phase uses TDD for the new rejection and stale-record behavior and characterization tests before touching the existing label parser, board write guard, and broker slot matcher.

## §B Decisions for the Kickoff gate

Operator answers were given directly on 2026-09-26 and relayed by the leader; leader decisions are marked as such. Items derived here from those answers are marked "derived" and are to be confirmed at Implementation Kickoff Approval.

- **O0 / Q1 — RESOLVED (operator).** `lane` is canonical. The `worker` and `agent` aliases are removed immediately: no compatibility aliases, no deprecation notice path. The reversal of the t1085 worker-canonical contract is approved, which removes the t1193 "unapproved reversal" ground. Consequence: REQ-RNC-003, -004, -005, -013, -020 rewritten from alias to rejection (non-zero exit, error naming the canonical form, nothing written). **Derived:** `lead` / `lead-<suffix>` are rejected by the same rule (REQ-RNC-007) — the operator's words named `worker`/`agent`; applying "no compatibility aliases" to `lead` is this SPEC's reading, to be confirmed at Kickoff.
- **O1 — RESOLVED (derived from O0).** Environment variable names are kept with no second name (REQ-RNC-011, design D3). No collision with O0: a single retained name is not an alias, and the dual-name window (the old alternative) is excluded by O0. Values follow the new vocabulary. To confirm at Kickoff: if the operator reads "no aliases" as requiring the names themselves to change, the only consistent form is a hard rename, which **collides** with card t1242 REQ-CFR-006/007/020 and the `env_vars` line in users' `.codex/config.toml` — that reading would be a blocker requiring a t1242 amendment first.
- **O2 — RESOLVED (derived from O0).** Persisted values: write-new, recognize-new-only, never rewrite (REQ-RNC-009, -010), plus the run-boundary rule for live legacy records (REQ-RNC-022). Impact stated in design §2: a mid-run binary reinstall or a downgrade requires retiring and relaunching the run; old rows are not rewritten; card history keeps its recorded owner names. To confirm at Kickoff, because it changes an operating habit of this repository (R4).
- **O3 — MOOT.** Legacy-spelling lifetime: the spellings are removed in this SPEC.
- **O4 — NOTED (leader act).** The leader edits card t1240's text in the queue to `-f lane` / `MOAI_FACTORY_ROLE=lane`. No SPEC action.
- **O5 — RESOLVED (operator/leader).** t1245 lands first with value `worker`; this SPEC's run converts it to `lane` in bulk. REQ-RNC-012 keeps the one-definition coupling (marker value = `-f` token = lane-label prefix, via t1245's REQ-AP-013 equality assertion) and, after conversion, the guard recognizes `lane` only; `worker`/`agent` are not accepted. No fail-open window: the only stamper (t1240) lands after this SPEC.
- **O6 — RESOLVED (leader).** t1193 is removed from the ordering until the operator decides. REQ-RNC-014 no longer halts on it; the file overlap is a recorded risk (R3) and pre-flight re-checks t1193's status.
- **Q3 — RESOLVED (operator).** Lane self-dispatch is allowed up to promoting a queued card. Code layer: no enforcement change — the pick path (`moai todo next <n>`) carries no role guard today and gains none; the board write guard is about board columns and stays leader-only. The one code-layer change is the pick command's help text (REQ-RNC-021). The documented HARD promotion clause is handed to t1257.
- **Q4 — RESOLVED (operator).** `manager-lead` is kept (REQ-RNC-017, out of scope).
- **Q5 — RESOLVED (operator).** Homonyms of `leader` (CG-mode leader, Agent Teams lead) are disambiguated with qualifiers, not renamed (REQ-RNC-023).
- **O7** [NEEDS CLARIFICATION: identifier rename scope] — not covered by the relayed answers. Recommended: rename role-sense Go identifiers in the touched packages (REQ-RNC-019, design D8). Alternative: strings only, identifiers later (amends REQ-RNC-019 and AC-RNC-019).

**Ordering (leader, 2026-09-26):** t1242 → t1245 → t1256 run → t1240 → t1257 body replacement. t1193 is excluded pending the operator's decision.

## §C Pre-flight (run-phase entry)

1. Kickoff approval recorded, including confirmation of the derived items in §B (O0 `lead` extension, O1, O2) and an answer to O7.
2. REQ-RNC-014 record: develop SHA; t1242 deletion landed (halt t1242-file milestones if not); t1245 constants present (selects the M4 branch); t1193 state — if t1193 has landed on develop or is about to, report to the leader before editing any of the six overlap files (R3).
3. Absorb local develop into the card worktree; re-measure the census (`python3 .moai/reports/t1256/census.py`) and record the new counts next to the plan-time counts.
4. Measure and record the CG-mode / Agent Teams user-facing string list for AC-RNC-023.
5. Characterization tests pass on the absorbed tree for the packages the milestones touch (`go test ./internal/kanban/... ./internal/cli/... ./internal/hook/... ./internal/factorymsg/...` — scoped, never `./...`).

## §D Constraints

- Schema byte-identical (REQ-RNC-008); no data migration, no row rewrite.
- Codex `env_vars` allowlist and env var names unchanged (REQ-RNC-011) — also card t1242 REQ-CFR-020.
- No edit under `internal/template/templates/**` (t1257).
- Unrelated senses untouched (REQ-RNC-016); `manager-lead` untouched (REQ-RNC-017); homonyms qualified, not renamed (REQ-RNC-023).
- Lane-local verification only; the full suite runs in CI on the develop push.

## §E Self-verification

Each milestone closes with: scoped `go test` on the touched packages, `go vet` on them, `golangci-lint run` on them, and the AC rows for that milestone recorded in progress.md §E.2 with command and verbatim output.

## §F Milestones (ordered by change likelihood — highest first)

### M1 — Persisted vocabulary and the run boundary (Priority High)
The decisions most likely to be revised at Kickoff (O2, the derived `lead` rejection) land first, as tests before code.
- Writers emit `leader` / `lane` / `lane-<n>` for role, slot, label, registry key, and card owner (REQ-RNC-010).
- Readers recognize only the new vocabulary (REQ-RNC-009); live legacy records in the same run trigger the refusal message, dead ones are stale, history displays verbatim (REQ-RNC-022).
- Broker delivery for `leader` / `lane-<n>`; legacy `to_slot` refused (REQ-RNC-013).
- Schema-freeze guard (REQ-RNC-008) and env-name/allowlist freeze guard with new values (REQ-RNC-011).
- ACs: AC-RNC-006, -008, -009, -010, -011, -022.

### M2 — CLI surface: tokens, labels, help, rejection (Priority High)
- `lane-<n>` produced; `worker-<n>`/`agent-<n>` refused on every input path (REQ-RNC-004, -005).
- `-f lane` canonical; `-f worker`/`-f agent` refused (REQ-RNC-002, -003); `lead`/`lead-<suffix>` refused (REQ-RNC-007).
- Usage error, `Use`/help strings for `cc`/`glm` (and `codex` if still present after t1242), collision messages, `moai tokens --role` example, `gtd answer` short text, board write-guard error text (REQ-RNC-001).
- `todo next` help text for lane self-dispatch (REQ-RNC-021).
- ACs: AC-RNC-001, -002, -003, -004, -005, -007, -016, -021.

### M3 — Notices, locales, dashboard, homonyms (Priority Medium)
- Factory and kanban SessionStart notices in en/ko/ja/zh per design §3 (REQ-RNC-015), including the stale-run notice.
- Doctor, statusline, and web display labels read `leader`/`lane`; legacy chains shown as legacy runs (REQ-RNC-001).
- CG-mode / Agent Teams strings carry their qualifier (REQ-RNC-023).
- ACs: AC-RNC-012, -013, -023.

### M4 — Factory role marker value (Priority Medium, conditional)
- Where t1245 has landed (expected under the ordering): flip the value constant to `lane`, keep the guard on the one constant so it recognizes `lane` only, and confirm t1245's equality assertion binds token, prefix, and marker (REQ-RNC-012).
- Where t1245 has not landed: record the value `lane` in progress.md for the leader to hand to t1245; no edit.
- ACs: AC-RNC-014.

### M5 — Mechanical rename and guards (Priority Low)
- Role-sense Go identifiers and comments (REQ-RNC-019), with legacy detection literals commented as refuse/detect-only.
- Test fixtures to the new vocabulary, with one rejection test per legacy spelling (REQ-RNC-020).
- Vocabulary guard test (REQ-RNC-018) and unrelated-sense diff check (REQ-RNC-016, -017).
- ACs: AC-RNC-015, -017, -018, -019, -020.

## §G Risks

| # | Risk | Mitigation |
|---|---|---|
| R1 | The leader role value moves from `lead` to `leader` in the board write guard, the role reader, and the run-retire owner lookup at once; a missed reader leaves the new leader unable to write the board (sole-writer refusal on every write) | M1 first; characterization tests on the board guard and role reader before the change; AC-RNC-008 exercises a `leader` declaration through all three readers |
| R2 | Stamp/guard value drift on the role marker → guard silently denies nothing | REQ-RNC-012 one-definition rule plus t1245's equality assertion; AC-RNC-014 asserts the `lane` deny and names every stamp/compare site |
| R3 | t1193 overlaps six files (`internal/cli/factory.go`, `internal/factorymsg/store.go`, `internal/hook/factory_messages.go`, `internal/hook/session_start_factory.go`, `internal/cli/codex_launcher.go`, `internal/codexwiring/configtoml.go`) and adds 49 lines of worker/lead vocabulary; it is out of the ordering pending the operator's decision | Pre-flight re-checks t1193's status (§C.2, AC-RNC-015); if it lands first, this run absorbs and converts its vocabulary; if it is live and unlanded, report to the leader before editing the overlap files |
| R4 | Mid-run binary reinstall — a routine act in this repository — now breaks a running factory or kanban run until it is relaunched (REQ-RNC-022, O2) | The refusal names the run and the relaunch step; the leader schedules the develop push that carries this change between runs; O2 confirmed at Kickoff |
| R5 | Guard test allowlist grows into a loophole | Allowlist limited to refuse/detect literals and their error text (REQ-RNC-018); AC-RNC-018 counts allowlist entries |
| R6 | Heuristic census undercounts multi-line user-facing strings | Pre-flight re-measure plus the guard test, which reads compiled string constants rather than lines |
| R7 | Operator retraining churn (third change in five days) with no hint path | Operator chose rejection knowingly (O0); every rejection names the canonical form |
| R8 | Card t1240 starts before its card text is amended (O4) and ships `-f agent` | O4 is the leader's queue edit; the ordering places t1240 after this run, whose rejection path would refuse `-f agent` at t1240's first test |

## §H Cross-references

- `.moai/reports/t1256/census.md`, `.moai/reports/t1256/conflicts.md`
- SPEC-FACTORY-WORKER-NAMING-001 (the contract this reverses), SPEC-CODEX-FACTORY-RETIRE-001 (t1242), SPEC-AUTONOMY-PRECONDITION-001 (t1245)
- Sibling card t1257 consumes `design.md` §3 and owns the promotion-clause amendment.
