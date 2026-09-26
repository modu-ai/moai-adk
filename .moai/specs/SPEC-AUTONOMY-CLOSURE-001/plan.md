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
| R1 | A1 receipt fields change again (v0.5.1, decider `llm` / `llm+jev`) | lenient display decoder, no decider value set hard-coded; validity from A1 verify only |
| R2 | A2 does not export a record reader or a read-only class-4 comparison | M0 pre-flight checks the exported surface; missing comparison → New APIs `not observed` plus a follow-up card, not a copy of A2 logic |
| R3 | Push-range staleness is path-based and can stop a push because of another card's commit | fail-safe direction; `push-check` names the commit so the operator can re-run the second review |
| R4 | A fail-closed PreToolUse check blocks a push on a transient git error | only under `mode: contract`, only for integration-branch pushes; bounded timeout; reason code names the cause |
| R5 | Evidence lives in gitignored per-worktree directories | evidence resolution order card worktree → primary checkout (design.md §C.3); undetermined → stop |
| R6 | Agents forge local records | accepted residual risk (spec.md §H); verdict path denied to agents; codex receipt shown as corroboration |
| R7 | Adding `card_id` to `audit_multi` changes tool behavior | argument optional; absent → byte-identical output (AC-CLOSURE-012) |

## §C. Pre-flight (run phase, before M1)

1. `git merge-base --is-ancestor <A1 landing commit> HEAD` and the same for A2 → both ancestors.
2. `go doc ./internal/contract Verify` and `go doc ./internal/contract LoadDir` → present.
3. `go doc ./internal/escalation` → record reader and class-4 comparison names recorded in
   progress.md §E.2 (R2).
4. Re-read A1's `design.md` § Kickoff Receipt at the landed commit; note field names (R1).
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
schema-version checks; the lenient kickoff receipt decoder.

Files: `internal/closure/model.go`, `internal/closure/records.go`, `internal/closure/receipt_view.go`,
tests. ACs: 002, 010, 013 (decoder half).

### M2 — Readiness rule and reason codes (Priority High)

Pure evaluator over injected inputs (range paths, contract facts, evidence files) returning the
closed code set (design.md §C.4). Includes OQ-1 outcome.

Files: `internal/closure/readiness.go`, tests. ACs: 013, 016, 018.

### M3 — Report builder and renderer (Priority High)

Section builders (design.md §F), evidence resolution (§C.3), Markdown renderer driven by the JSON
model only, determinism.

Files: `internal/closure/build.go`, `internal/closure/sections_*.go`, `internal/closure/render_md.go`,
`internal/closure/evidence.go`, `internal/closure/gitio/gitio.go`, tests. ACs: 003-009, 011, 014,
022, 024, 026.

### M4 — CLI surfaces (Priority Medium)

`moai contract report`, `moai contract verdict`, `moai contract push-check` on A1's `contract`
command; verdict human path reusing A1's marker set and TTY seam.

Files: `internal/cli/contract_report.go`, `internal/cli/contract_verdict.go`,
`internal/cli/contract_pushcheck.go`, tests. ACs: 001, 019, 020.

### M5 — Second-review record from `audit_multi` (Priority Medium)

Optional `card_id` argument; record append after the existing persistence step; SPEC lookup via the
queue store; contract digest via A1 `Verify`.

Files: `internal/cli/mcp_audit_multi.go`, `internal/cli/mcp_server.go`,
`internal/cli/mcp_convergence.go` (append hook point only), tests. AC: 012.

### M6 — Hook wiring (Priority Medium)

Verdict deny and push readiness in `internal/hook/pre_tool.go` after `checkBashCommand`
(design.md §E); mode check first.

Files: `internal/hook/pre_tool.go`, `internal/hook/closure_push.go`, tests. ACs: 015, 017, 021, 023.

### M7 — Auditor instructions and mirrors (Priority Low, mechanical)

Contract-mode second review passes `card_id` to `audit_multi`: template and local copies of
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

## §H. Open Questions

- **OQ-1** [NEEDS CLARIFICATION: Which recorded human verdicts block a push? Proposed default: a
  latest verdict of `reject` or `amend-contract` blocks (`human_verdict_reject`,
  `human_verdict_amend_contract`); no recorded verdict does not block, so autonomous runs are not
  forced to wait for a human. Alternative: require a current `accept` before any push. Also confirm
  that a performed second review with verdict `fail` blocks under `second_review: required`
  (`second_review_failed`); the card text names only "not performed".]
- **OQ-2** [NEEDS CLARIFICATION: Confirm the closure report file name `closure-report.md` (+ `.json`)
  instead of the design artifact's `verdict.md`, which is already the lead's hand-authored verdict
  file for 381 cards in this tree (research.md §B.6).]

## §I. Cross-References

- spec.md §D (requirements), design.md (records, readiness, layout), research.md (measured facts),
  acceptance.md (criteria).
- A1: SPEC-AUTONOMY-CONTRACT-001 (branch `WT-contract-schema`, read at `67a2f55cb`).
- A2: SPEC-AUTONOMY-ESCALATION-001 (read at `8c9ee29b7`), §I record format.
