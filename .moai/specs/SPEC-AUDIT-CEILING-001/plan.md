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
- **Pre-existing mirror drift**: `harness.yaml`, `phase-execution.md`, and
  `plan-auditor.md` mirrors already differ from deployed copies (research.md
  §3). Do not encode the drift as expected state; do not repair unrelated
  drift.

## §C Pre-flight (all measured on `2f492df19`, 2026-10-04)

```bash
git rev-parse --short HEAD          # 2f492df19
git branch --show-current           # WT-audit-ceiling-counter
go build ./...                      # green baseline
go test ./internal/runtime/... ./internal/auditverdict/... ./internal/config/...   # green baseline
go run ./cmd/moai spec lint SPEC-AUDIT-CEILING-001 --strict   # must be 0/0 before any commit
```

RED-now baselines (verbatim commands, this run, this tree, exit codes recorded):

| Baseline | Command | Observed |
|---|---|---|
| conflict text present | `grep -c "Override and proceed" .claude/skills/moai/workflows/run/phase-execution.md` | 1 (exit 0) |
| acknowledgement option present | `grep -c "Proceed with acknowledgement" .claude/skills/moai/workflows/run/phase-execution.md` | 1 (exit 0) |
| no counter | `grep -rn "audit.round\|AuditRound\|iteration.count" internal/runtime/*.go` | 0 hits (exit 1) |
| no receipt parsing | `grep -c "convergence\|receipt" internal/auditverdict/verdict.go` | 0 (exit 1) |
| §9 rows | sed -n '190,201p' auto-semantics.md, count `\|^| ` lines | 11 `\|^| ` lines = header + 10 disposition rows (the separator row does not match the pattern) |
| §9 named rows absent | the 11-row named grep of acceptance.md AC-ACE-014 | 0 (exit 1) |
| retired row present | `grep -c "plan-audit bypass flags" .claude/rules/moai/workflow/auto-semantics.md` | 1 (exit 0) |
| config orphan note | `grep -c "no Go reader" internal/config/loader.go` | 2 (exit 0) |
| bare symmetry selector absent | `grep -cE "func TestStructYAMLSymmetry\(" internal/config/audit_struct_yaml_symmetry_test.go` | 0 (exit 1; only `_`-suffixed variants exist) |
| GateConfig production-dead | `grep -rn "runtime\.GateConfig" internal/ cmd/ --include="*.go" \| grep -v _test` | 0 non-test matches; 17 test-only |
| mirror drift | `diff -q` deployed vs template | plan-auditor.md DIFF, phase-execution.md DIFF, harness.yaml DIFF; auto-semantics.md SAME, convention doc SAME |

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
  for every RB criterion whose RED is a new test (AC-ACE-001/003/004/006/
  007/008/009/010/011/015/017/018/019/020/021/022), and grep-class RED cells
  with recorded exit codes for AC-ACE-002/013/014 plus AC-ACE-008's
  export-path grep (baselines in §C).

## §F Milestones

### M1 (Priority High) — verdict receipt + admission predicate extension

Data-model first: the receipt schema is the least reversible decision.

- Extend `.moai/docs/audit-artifact-convention.md` § What with the receipt
  line format (design.md §3): `convergence_overall: <pass|fail>` and a
  repeatable `required_backend: <backend> <pass|fail|inconclusive>` line,
  one per required backend.
- Receipt producer (D19): the plan-auditor agent body is the writer — its
  export step (`.claude/agents/moai/plan-auditor.md` § Output Format)
  appends the receipt lines from the `audit_multi` convergence result it
  already receives (`ConvergenceResult.OverallVerdict` + `PerBackendVerdicts`,
  design.md §3) to the exported verdict file per the convention § What; the
  step lands in the deployed agent body AND its template mirror in the same
  change (research.md §3 edit target; the pre-existing whole-file drift keeps
  it on AC-ACE-015's known-FAIL carve-out list for the untouched regions).
- Config-error disposition (D21): the M1 call sites resolve the gate set
  with the opposite of today's `resolveAuditGates` fail-open path — the
  resolution result distinguishes an error from a genuinely-empty
  configuration, and an unreadable/unparseable audit section refuses
  (REQ-ACE-010's third trigger arm; design.md §4).
- Extend `internal/auditverdict`: `Parse` reads the receipt keys;
  `Admit` (PhasePlan) gains the tree's required-backend set as input and
  refuses on (a) any required backend fail or inconclusive regardless of
  label (REQ-ACE-009), (b) required backend configured + receipt absent or
  missing that backend's line (REQ-ACE-010, Q4 default refuse). The override
  is NOT an `Admit` input — it lives at the CLI seam (M3, design.md §4).
- Update the three call sites (`decide.go`, `contract/rules.go`,
  `homestate/card_evidence_readers.go`) to pass the configured gate set.
- Mirrors: convention doc mirror in the same change.
- Depends on: decision-index Q4 verdict (blocker report if unresolved — the
  refuse default stands while the question stays open).

### M2 (Priority High) — config Go reader

- Add `PlanAuditTierCeilingsConfig` and `PlanAuditCeilingPolicyConfig`
  structs (`internal/config/types.go`), wire into the harness config load,
  defaults, and the symmetry audit; validate `on_final_hit` (accepted value
  `hold-and-split`, anything else a config error); remove the two
  "no Go reader" orphan entries (`loader.go:345-346`).
- Add a bare `func TestStructYAMLSymmetry` harness — the exact name
  AC-ACE-002's selector runs (only `_`-suffixed variants exist today;
  §C baseline 0 matches) — covering `harness.yaml` with a symmetry case for
  the new structs.
- No `harness.yaml` edit: the config content is unchanged (design.md §6), so
  no mirror sync belongs to this milestone; the measured pre-existing
  `harness.yaml` mirror drift stays untouched (Out of Scope).

### M3 (Priority High) — counter + ceiling-policy engine + enforcement

- `internal/runtime`: round-count derivation from iteration evidence
  (both families, one-iteration-once identity = SPEC id + iteration number;
  REQ-ACE-001), the delta-eligibility check (fix_scope anchors + REQ/AC id
  sets + STOP; REQ-ACE-003), and the policy outcome engine (debt-admit /
  scope-split / hold-record; REQ-ACE-004..006).
- Enforcement wiring at the LIVE admission seams (the iter1 D4 finding —
  `GateConfig.Invoke` has no production caller, measured 0 non-test
  references): the kickoff evaluator (`decide.go`) and the homestate card
  transition call the engine before admission and surface the same refusal +
  outcome (exit nonzero / refusal record); `GateConfig.Invoke` gains the same
  Step-0 call as the library-level consumer for when a caller exists. An
  integration test proves a ceiling-hit round refuses at a production entry
  point at BOTH seams — AC-ACE-022 carries one arm per LIVE seam, the card
  transition and the kickoff evaluator (D23).
- Refusal output: structured (JSON or parseable lines) carrying outcome,
  reasons, evidence paths; persist to `progress.md`; audit-trail log append
  (REQ-ACE-007, REQ-ACE-012). Design the AuditResult extension per
  design.md §7 (separate outcome field, not a new Verdict enum value that
  the default branch would fold into INCONCLUSIVE).
- Override input for required-backend refusals (REQ-ACE-011) — explicit
  flag + note + logging only; the CLI writes the ack to `progress.md` §G
  Override and Refusal Record (outside the plan-artifact hash subject set)
  and the REQ-ACE-012 trail.
- Depends on: decision-index Q2/Q3/Q5 verdicts (blocker report if
  unresolved — the SPEC's embedded defaults stand while the questions stay
  open, each recorded kickoff-amendable).

### M4 (Priority Medium) — doc reconciliation

- `phase-execution.md` Step 4c/4d: rewrite to the fail-closed path — the
  ceiling-policy outcome is the only non-block exit the question branches
  offer; no AskUserQuestion branch, no override-and-proceed, no BYPASSED
  recording (REQ-ACE-013).
- `auto-semantics.md` §9: add the 11 rows of spec.md §D.2 with dispositions
  from the existing vocabulary, each citing file + section (REQ-ACE-014).
  Row 2 REPLACES the `plan-audit bypass flags` row: 10 existing rows − 1 + 11
  = 20 disposition rows after M4.
- Cross-check `run.md` § Run-phase Autonomy and the operator-form gate text
  for residual references to the removed override branch.
- Mirrors in the same change.

### M5 (Priority Low) — regression guard + mirror verification

- Template audit tests extended: region-scoped deployed-vs-mirror equality
  for every file this SPEC edits; the three files carrying pre-existing
  whole-file drift (phase-execution.md, plan-auditor.md, harness.yaml —
  each measured DIFF at f2f815008) are named known-FAIL until repaired,
  never expected-pass (AC-ACE-015's carve-out list).
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

- spec.md §B (REQ-ACE-001..016), §C constraints, §D.2 row list
- acceptance.md §D (AC-ACE-001..022), §C edge cases
- design.md §1-§10 (counter model, receipt schema, enforcement, open points)
- research.md §1-§5 (source verification, Go surfaces, mirrors, gaps)
- decision-index.md Q1-Q6 (unresolved operator decisions)
