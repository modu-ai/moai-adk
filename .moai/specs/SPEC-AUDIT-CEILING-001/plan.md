# plan.md — SPEC-AUDIT-CEILING-001

Tier L implementation plan. Milestones are ordered by decision-reversibility:
the data-model decisions (verdict receipt schema, admission predicate
signature, config reader) come first because they are the least reversible and
the most likely to change under review; mechanical mirror-sync and regression
guarding come last. No wall-clock estimates — priority labels only.

## §A Context

- Card t1500 ([v3.2 1단계·감사], Class C). Branch `WT-audit-ceiling-counter`,
  worktree `.moai/worktrees/t1500`, base HEAD `2f492df19` (develop tip).
- SPEC directory: `.moai/specs/SPEC-AUDIT-CEILING-001/` (Tier L: spec.md,
  plan.md, acceptance.md, design.md, research.md).
- The measured defect chain: prose-only ceiling (405 audits / 160 SPECs, max
  11 vs configured 1/2/3) + a required-backend convergence `fail` that cannot
  influence admission (no receipt fields anywhere) + two governing texts that
  disagree on the negative-verdict path. Full evidence:
  `research.md` §1-§2.
- Infrastructure: `internal/runtime/audit_gate.go` (GateConfig.Invoke),
  `internal/runtime/audit_review.go` (iteration-stream reader),
  `internal/auditverdict/verdict.go` (the one admission predicate),
  `internal/contract/kickoff/decide.go`, `internal/homestate/card_*`
  (card-transition guard), `internal/cli/mcp_convergence.go`
  (ConvergenceResult), `internal/config/loader.go` (orphan keys).

## §B Known Issues (auto-injection, relevant subset)

- **B2 cross-SPEC policy conflict**: `phase-execution.md` Step 4c/4d vs
  `auto-semantics.md` §7/§9 — the conflict this SPEC resolves (REQ-ACE-013);
  do not "fix" one side without the other.
- **B3/C-HRA-008 subagent boundary**: the CLI path gains no interactive
  prompt; static guard test required (AC-ACE-016).
- **B5 CI 3-tier**: spec-lint, golangci-lint, go test fail separately;
  baselines measured in §C to separate NEW defects from pre-existing.
- **B8 working-tree hygiene**: stage by explicit pathspec; no `.moai/state/`
  or `.moai/harness/` writes.
- **Pre-existing mirror drift**: `harness.yaml` and `phase-execution.md`
  mirrors already differ from deployed copies (research.md §3). Do not encode
  the drift as expected state; do not repair unrelated drift.

## §C Pre-flight (all measured on `2f492df19`, 2026-10-04)

```bash
git rev-parse --short HEAD          # 2f492df19
git branch --show-current           # WT-audit-ceiling-counter
go build ./...                      # green baseline
go test ./internal/runtime/... ./internal/auditverdict/... ./internal/config/...   # green baseline
go run ./cmd/moai spec lint SPEC-AUDIT-CEILING-001 --strict   # must be 0/0 before any commit
```

RED-now baselines (verbatim commands, this run, this tree):

| Baseline | Command | Observed |
|---|---|---|
| conflict text present | `grep -c "Override and proceed" .claude/skills/moai/workflows/run/phase-execution.md` | 1 |
| no counter | `grep -rn "audit.round\|AuditRound\|iteration.count" internal/runtime/*.go` | 0 hits |
| no receipt parsing | `grep -c "convergence\|receipt" internal/auditverdict/verdict.go` | 0 |
| §9 rows | sed -n '190,201p' auto-semantics.md, count `\|^| ` lines | 11 lines = header + separator + 10 disposition rows |
| config orphan note | loader.go:345-346 | "no Go reader" on both ceiling keys |
| mirror drift | `diff -q` deployed vs template | harness.yaml DIFF, phase-execution.md DIFF, auto-semantics.md SAME, convention doc SAME |

## §D Constraints

- Template-First: every deployed edit lands with its
  `internal/template/templates/` mirror in the same commit (REQ-ACE-015).
- No interactive prompt anywhere in the CLI path (C3; guard test).
- Admission thresholds and auditor behavior untouched (§C5, Out of Scope).
- No new config keys (research.md §4): the three-outcome ladder is
  REQ-encoded behavior; config keeps `auto_delta_rounds` +
  `on_final_hit` and gains only a Go reader.
- `phase:`/`status:` frontmatter discipline per the schema SSOT; artifacts
  other than spec.md carry no `status:` field.
- progress.md §E.2/§E.3 belong to manager-develop and §E.4 to manager-docs —
  this plan never writes them.

## §E Self-Verification

Per-milestone, reported in the 5-section evidence-bearing format:

- E1 AC binary matrix (acceptance.md §D) with command + verbatim output +
  HEAD attribution per row.
- E2 `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- E3 `go test -cover ./internal/runtime/... ./internal/auditverdict/...`
  at or above the 85% package threshold for changed packages.
- E4 subagent-boundary grep over changed CLI packages (0 matches).
- E5 `golangci-lint run --timeout=2m` — NEW issues named separately from the
  measured baseline.
- E6 RED evidence per TDD AC (verbatim pre-GREEN failure output) — required
  for AC-ACE-003/004/005/006/009/010 whose RED is a new test, and
  grep-class RED for AC-ACE-002/013/014 (baselines in §C).

## §F Milestones

### M1 (Priority High) — verdict receipt + admission predicate extension

Data-model first: the receipt schema is the least reversible decision.

- Extend `.moai/docs/audit-artifact-convention.md` § What with the receipt
  line format (design.md §3): `convergence_overall: <pass|fail>` and a
  repeatable `required_backend_fail: <backend>` line.
- Extend `internal/auditverdict`: `Parse` reads the receipt keys;
  `Admit` (PhasePlan) gains the tree's required-backend set as input and
  refuses on (a) any required backend fail regardless of label (REQ-ACE-009),
  (b) required backend configured + receipt absent (REQ-ACE-010, Q4
  default refuse), (c) explicit override path feeding REQ-ACE-011.
- Update the three call sites (`decide.go`, `contract/rules.go`,
  `homestate/card_evidence_readers.go`) to pass the configured gate set.
- Mirrors: convention doc mirror in the same change.
- Depends on: decision-index Q4 verdict (blocker report if unresolved — the
  refuse default stands while the question stays open).

### M2 (Priority High) — config Go reader

- Add `PlanAuditTierCeilingsConfig` and `PlanAuditCeilingPolicyConfig`
  structs (`internal/config/types.go`), wire into the harness config load,
  defaults, and `audit_struct_yaml_symmetry_test.go`; remove the two
  "no Go reader" orphan entries (`loader.go:345-346`).
- Mirrors: `harness.yaml` template mirror — note the measured pre-existing
  drift; sync only the keys this SPEC touches and name the residue.

### M3 (Priority High) — counter + ceiling-policy engine + enforcement

- `internal/runtime`: round-count derivation from iteration evidence
  (both families, one-iteration-once identity; REQ-ACE-001), the policy
  outcome engine (debt-admit / scope-split / hold-record; REQ-ACE-004..006),
  and the ceiling check placed before the auditor spawn in
  `GateConfig.Invoke` (REQ-ACE-003).
- Enforcement wiring: `kickoff decide` and `homestate` card transition
  surface the same refusal + outcome (exit nonzero / refusal record).
- Refusal output: structured (JSON or parseable lines) carrying outcome,
  reasons, evidence paths; persist to `progress.md`; audit-trail log append
  (REQ-ACE-007, REQ-ACE-012). Design the AuditResult extension per
  design.md §7 (separate outcome field, not a new Verdict enum value that
  the default branch would fold into INCONCLUSIVE).
- Override input for required-backend refusals (REQ-ACE-011) — explicit
  flag + note + logging only.

### M4 (Priority Medium) — doc reconciliation

- `phase-execution.md` Step 4c/4d: rewrite to the fail-closed path — the
  ceiling-policy outcome is the only non-block exit; no AskUserQuestion
  branch, no override-and-proceed, no BYPASSED recording (REQ-ACE-013).
- `auto-semantics.md` §9: add the 11 rows of spec.md §D.3 with dispositions
  from the existing vocabulary, each citing file + section (REQ-ACE-014).
  Retire/adjust the `plan-audit bypass flags` row wording to name the
  machine enforcement.
- Cross-check `run.md` § Run-phase Autonomy and the operator-form gate text
  for residual references to the removed override branch.
- Mirrors in the same change.

### M5 (Priority Low) — regression guard + mirror verification

- Template audit tests extended: deployed-vs-mirror byte equality for every
  file this SPEC edits (names the pre-existing drift files as known-FAIL
  until repaired — never as expected-pass).
- Static guard: no interactive prompt in the new CLI surface.
- Full lint + spec lint --strict 0/0; §E self-verification report.

## §G Anti-Patterns

- Do not import the research note's C2 tier taxonomy into §9 (spec.md Out of
  Scope).
- Do not count the daily run-history file (`.moai/reports/plan-audit/…`) as
  round evidence — iteration stream only.
- Do not add a second prose statement of the ceiling policy; C1 keeps one
  policy with the CLI as its machine consumer.
- Do not weaken `Admit`'s existing checks while adding refusal causes.
- Do not "fix" the pre-existing harness.yaml / phase-execution.md mirror
  drift inside this SPEC's commits beyond the keys/sections this SPEC edits.

## §H Cross-References

- spec.md §B (REQ-ACE-001..016), §C constraints, §D.3 row list
- acceptance.md §D (AC-ACE-001..016), §C edge cases
- design.md §1-§10 (counter model, receipt schema, enforcement, open points)
- research.md §1-§5 (source verification, Go surfaces, mirrors, gaps)
- decision-index.md Q1-Q6 (unresolved operator decisions)
