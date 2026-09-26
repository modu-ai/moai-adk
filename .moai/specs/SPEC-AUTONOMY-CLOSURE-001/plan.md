# plan.md — SPEC-AUTONOMY-CLOSURE-001

## §A. Context

Card t1237 (AUTONOMY-A4). Tier **L**: three new commands, one new package, changes to the MCP
server and the PreToolUse hook, template and local mirrors of an agent and a skill, and a regenerated
Codex agent copy — well over ten files across more than three milestones, and the push stop is a new
gate on an irreversible action. Development mode per `.moai/config/sections/quality.yaml`; the pure
core (`internal/closure`) suits RED-GREEN-REFACTOR.

Ordering: run phase starts only after A1 (card t1234) and A2 (card t1235) have landed on `develop`.
A2b (card t1245) is not a hard prerequisite (design.md §C.1, §D fall back to A4's own classifier and
deny); A3 (card t1236) is independent.

## §B. Known Issues and Risks

| ID | Risk | Mitigation |
|---|---|---|
| R1 | A1 receipt fields change again after v0.5.1 | display decoder keyed on v0.5.1 fields; any unrecognized field listed under Not Performed, never dropped; validity from A1 verify only |
| R2 | A2 does not export a record reader or a read-only class-4 comparison | M0 pre-flight checks the exported surface; missing comparison → New APIs `not observed` plus a follow-up card, not a copy of A2 logic |
| R3 | Currency is path-based and can stop a candidate card's push because of another card's commit in the same range | fail-safe direction; `STALE` names the commit so the operator can re-run that card's second review and report. Candidacy is own-card (REQ-CLOSURE-015), so closed cards already on the remote are not re-evaluated; the accepted residual is that a code-only commit touching a closed card's governed paths is not attributed to it (spec.md §H) |
| R4 | A fail-closed PreToolUse check blocks a push on a transient git error | only under `mode: contract`, only for integration-branch pushes; bounded timeout; reason code names the cause |
| R5 | Evidence lives in gitignored per-worktree directories | one card evidence home for writers and readers (spec.md §C.7, design.md §C.3); undetermined → stop |
| R6 | Agents forge local records | accepted residual risk (spec.md §H); verdict path denied to agents; codex receipt shown as corroboration |
| R7 | Adding `card_id` to `audit_multi` changes tool behavior | argument optional; absent → byte-identical output (AC-CLOSURE-012) |
| R8 | Push classifier misses a push form | fail-closed: an unprovable destination is `push_check_undetermined` (REQ-CLOSURE-017) |

## §C. Pre-flight (run phase, before M1)

1. `git merge-base --is-ancestor <A1 landing commit> HEAD` and the same for A2 → both ancestors.
2. `go doc ./internal/contract Verify` and `go doc ./internal/contract LoadDir` → present.
3. `go doc ./internal/escalation` → record reader and class-4 comparison names recorded in
   progress.md §E.2 (R2).
4. Re-read A1's `design.md` § Kickoff Receipt and § Card Field at the landed commit and diff them
   against `65e0a9167`; record any delta in progress.md §E.2 (R1).
5. `grep -n 'checkBashCommand(input.ToolInput)' internal/hook/pre_tool.go` → anchor line re-measured.

## §D. Constraints

- No change to A1 or A2 files; no new contract field; no new configuration key.
- No write to `.moai/reports/<card-id>/verdict.md` by any code path.
- Under `guided`: hook output byte-identical, no new subprocess (AC-CLOSURE-023).
- Template content neutral (spec.md REQ-CLOSURE-025): no SPEC IDs, card IDs, dates, SHAs.
- Template-First: edit `internal/template/templates/**` first, mirror locally, `make build`; after
  editing a template agent run `make agents-emit` (Codex copy regenerated, never hand-edited).
- Tests: `t.TempDir()` fixtures only; run affected packages (`./internal/closure/...`,
  `./internal/cli/...` targeted `-run`, `./internal/hook/...`), never the full suite locally.

## §E. Self-Verification (run phase reports)

- E1 AC matrix: `TestAC_CLOSURE_<NNN>` pass check per acceptance.md §A, verbatim output.
- E2 `go build ./...` and `GOOS=windows go build ./...`.
- E3 `go test -cover ./internal/closure/...` ≥ 85%.
- E4 `golangci-lint run ./internal/closure/... ./internal/cli/... ./internal/hook/...` clean.
- E5 `make agents-emit-check` clean after the template agent edit.
- E6 `moai spec lint SPEC-AUTONOMY-CLOSURE-001` clean.

## §F. Milestones (ordered by decision reversibility — most likely to change first)

### M1 — Record shapes and the report model (Priority High)

Data-model decisions first: second-review record (design.md §A.1), human verdict record (§A.2),
report JSON model and section order (§A.3), not-performed catalogue (§B). Decoders with strict
schema-version checks; the kickoff receipt display decoder (v0.5.1 fields).

Files: `internal/closure/model.go`, `internal/closure/records.go`, `internal/closure/receipt_view.go`,
tests. ACs: 002, 010, 013 (decoder half).

### M2 — Readiness rule and reason codes (Priority High)

Pure evaluator over injected inputs (range paths, contract facts, evidence files) returning the
closed nine-code set (design.md §C.4), the second-review selection (§D), and the currency rule
(spec.md §C.7).

Files: `internal/closure/readiness.go`, tests. ACs: 013, 016, 018.

### M3 — Report builder and renderer (Priority High)

Section builders (design.md §G), card evidence home (§C.3), Markdown renderer driven by the JSON
model only, determinism.

Files: `internal/closure/build.go`, `internal/closure/sections_*.go`, `internal/closure/render_md.go`,
`internal/closure/evidence.go`, `internal/closure/gitio/gitio.go`, tests. ACs: 002-009, 011, 014,
022.

### M4 — CLI surfaces (Priority Medium)

`moai contract report`, `moai contract verdict`, `moai contract push-check` on A1's `contract`
command; verdict human path reusing A1's marker set and TTY seam.

Files: `internal/cli/contract_report.go`, `internal/cli/contract_verdict.go`,
`internal/cli/contract_pushcheck.go`, tests. ACs: 001, 019, 020, 024.

### M5 — Second-review record from `audit_multi` (Priority Medium)

Optional `card_id` argument; record append (with target and reviewed scope) into the card evidence
directory after the existing persistence step; SPEC lookup via the queue store; contract card and
digest via A1 `Verify`; failure field `second_review_record_error`.

Files: `internal/cli/mcp_audit_multi.go`, `internal/cli/mcp_server.go`,
`internal/cli/mcp_convergence.go` (append hook point only), tests. AC: 012.

### M6 — Hook wiring (Priority Medium)

Verdict deny and push readiness in `internal/hook/pre_tool.go` after `checkBashCommand`
(design.md §F); mode check first; fail-closed push classifier (design.md §C.1).

Files: `internal/hook/pre_tool.go`, `internal/hook/closure_push.go`, tests. ACs: 015, 017, 021, 023.

### M7 — Auditor instructions and mirrors (Priority Low, mechanical)

Contract-mode second review passes `card_id` with target `baseBranch` after the last commit changing
the governed paths, between the `moai:closure-second-review` markers: template and local copies of
`.claude/skills/moai-ref-cross-model-audit/SKILL.md` and `.claude/agents/moai/sync-auditor.md`;
`make agents-emit`; `make build`.

Files: the four Markdown copies plus the regenerated `internal/template/templates/.codex/agents/moai/sync-auditor.toml`.
AC: 025.

## §G. Anti-Patterns to avoid

- Rendering an absent input as an empty list or a pass.
- Taking any state from an agent's message instead of a file.
- Re-implementing A1 verify or A2 record parsing.
- A conditional return above `checkBashCommand`.
- Writing `verdict.md`.

## §H. Resolved Decisions

- **OQ-1** — resolved 2026-09-26, lead decision. (a) A latest human verdict of `reject` or
  `amend-contract` blocks the push (`human_verdict_reject`, `human_verdict_amend_contract`); no recorded
  human verdict does not block, because in autonomous mode the human review is post-hoc and reversal
  runs through revoke or reject. (b) Under `second_review: required`, a second review that was performed
  and returned `fail` blocks the push with its own code `second_review_failed`, distinct from
  `second_review_not_performed`, and the report renders it distinctly (`FAILED`). Carried by
  REQ-CLOSURE-014, REQ-CLOSURE-016, AC-CLOSURE-014, AC-CLOSURE-016.
- **OQ-2** — resolved 2026-09-26, lead decision. The closure report is `closure-report.md` plus
  `closure-report.json`; `verdict.md` stays untouched as the lead's hand-written convention. Carried by
  spec.md §C.4 and REQ-CLOSURE-001.

## §I. Cross-References

- spec.md §D (requirements), design.md (records, readiness, layout), research.md (measured facts),
  acceptance.md (criteria).
- A1: SPEC-AUTONOMY-CONTRACT-001 v0.5.1 (branch `WT-contract-schema`, read at `65e0a9167`).
- A2: SPEC-AUTONOMY-ESCALATION-001 (read at `8c9ee29b7`), §I record format.
