# plan.md — SPEC-AUDIT-CEILING-001

Tier L implementation plan (v0.4.0, narrowed scope per operator decision D9).
Milestones are ordered by decision-reversibility: the data-model decisions
(verdict receipt schema, admission predicate signature, config reader) come
first because they are the least reversible and the most likely to change
under review; mechanical mirror-sync and regression guarding come last. No
wall-clock estimates — priority labels only.

The D9 scope cut removed the former M4 (phase-execution.md /
auto-semantics.md doc reconciliation); the current M4 is the former M5.

## §A Context

- Card t1500 ([v3.2 1단계·감사], Class C). Branch `WT-audit-ceiling-counter`,
  worktree `.moai/worktrees/t1500`, base HEAD `2f492df19` (develop tip).
- SPEC directory: `.moai/specs/SPEC-AUDIT-CEILING-001/` (Tier L: spec.md,
  plan.md, acceptance.md, design.md, research.md).
- The measured defect chain: prose-only ceiling (405 audits / 160 SPECs, max
  11 vs configured 1/2/3) + a required-backend convergence `fail` that cannot
  influence admission (no receipt fields anywhere). Full evidence:
  `research.md` §1-§2.
- Operator decision D9 (card t1500, 2026-10-06): resume with narrowed scope —
  the counter, the single no-question policy-outcome path, and required-backend
  fail blocking run entry. The run-gate doc reconciliation (former work item
  3) and the cross-card re-audit counting refinement are follow-up card
  material.
- Infrastructure: `internal/runtime/audit_gate.go` (GateConfig.Invoke),
  `internal/runtime/audit_review.go` (iteration-stream reader),
  `internal/auditverdict/verdict.go` (the one admission predicate),
  `internal/contract/kickoff/decide.go`, `internal/homestate/card_*`
  (card-transition guard), `internal/cli/mcp_convergence.go`
  (ConvergenceResult), `internal/config/loader.go` (orphan keys).

## §B Known Issues (auto-injection, relevant subset)

- **B3/C-HRA-008 subagent boundary**: the CLI path gains no interactive
  prompt; static guard test required (AC-ACE-016).
- **B5 CI 3-tier**: spec-lint, golangci-lint, go test fail separately;
  baselines measured in §C to separate NEW defects from pre-existing.
- **B8 working-tree hygiene**: stage by explicit pathspec; no `.moai/state/`
  or `.moai/harness/` writes.
- **Pre-existing mirror drift**: `plan-auditor.md` mirrors already differ
  from its deployed copy (research.md §3; re-measured DIFF at 69a085b2d).
  Do not encode the drift as expected state; do not repair unrelated drift.
  (The D9 scope cut removed phase-execution.md from this SPEC's edit targets;
  its drift is the follow-up card's material.)

## §C Pre-flight (re-measured on `69a085b2d`, 2026-10-06 — the v0.4.0 re-plan HEAD; originally measured on `2f492df19`, 2026-10-04)

```bash
git rev-parse --short HEAD          # 69a085b2d (v0.4.0 work start)
git branch --show-current           # WT-audit-ceiling-counter
go build ./...                      # green baseline
go test ./internal/runtime/... ./internal/auditverdict/... ./internal/config/...   # green baseline
go run ./cmd/moai spec lint SPEC-AUDIT-CEILING-001 --strict   # must be 0/0 before any commit
```

RED-now baselines (verbatim commands, this run, this tree, exit codes recorded; every value re-executed at 69a085b2d and matching the pre-implementation measurement at 2f492df19):

| Baseline | Command | Observed @69a085b2d |
|---|---|---|
| no counter | `grep -rn "audit.round\|AuditRound\|iteration.count" internal/runtime/*.go` (non-test) | 0 hits (exit 1) |
| no receipt parsing | `grep -c "convergence\|receipt" internal/auditverdict/verdict.go` | 0 (exit 1) |
| config orphan note | `grep -c "no Go reader" internal/config/loader.go` | 2 (exit 0) |
| bare symmetry selector absent | `grep -cE "func TestStructYAMLSymmetry\(" internal/config/audit_struct_yaml_symmetry_test.go` | 0 (exit 1; only `_`-suffixed variants exist) |
| package-wide selector listing | `go test -list '^TestStructYAMLSymmetry$' ./internal/config` | `ok github.com/modu-ai/moai-adk/internal/config 0.352s`, exit 0, no test listed |
| GateConfig production-dead | `grep -rn "runtime\.GateConfig" internal/ cmd/ --include="*.go" \| grep -v _test` | 0 non-test matches |
| receipt export instruction absent (multi-model arm) | `grep -c "convergence_overall" .claude/agents/moai/plan-auditor.md` | 0 (exit 1) |
| receipt export instruction absent (single-model arm) | `grep -c "backend it actually ran" .claude/agents/moai/plan-auditor.md` | 0 (exit 1; measured at a4c5b9594 — the V4-D3 per-arm split needs each arm's own baseline) |
| receipt export instruction absent (multi-model arm) | `grep -c "PerBackendVerdicts" .claude/agents/moai/plan-auditor.md` | 0 (exit 1; measured at 23fe75465 — the V4-D4 third count; the convergence-result field name is multi-model-only, absent from the single-model instruction by design §3's phrasing pins. The verdict's example phrase `convergence result` was measured 1 / exit 0 at 23fe75465 — pre-existing prose at plan-auditor.md:251 — so it cannot discriminate arms and was not adopted) |
| mirror drift | `diff -q` deployed vs template | plan-auditor.md DIFF; audit-artifact-convention.md SAME (phase-execution.md DIFF measured but no longer an edit target — Out of Scope) |
| Tier ceilings verified | `sed -n '75,84p' .moai/config/sections/harness.yaml` | S:1 M:2 L:3; policy auto_delta_rounds=1, on_final_hit=hold-and-split |
| resolver fail-open confirmed | `sed -n '123,131p' internal/cli/mcp_worktree_root.go` | `if err != nil { return config.AuditGates{}, "" }` — the fail-open path REQ-ACE-010 corrects |

## §D Constraints

- Template-First: every deployed edit lands with its
  `internal/template/templates/` mirror in the same commit (REQ-ACE-014).
- No interactive prompt anywhere in the CLI path (C3; guard test).
- Admission thresholds and auditor behavior untouched (§C5, Out of Scope).
- No new config keys (research.md §4): the outcome ladder is REQ-encoded
  behavior; config keeps `auto_delta_rounds` + `on_final_hit` and gains only
  a Go reader.
- D9 scope cut: no edit to `.claude/skills/moai/workflows/run/phase-execution.md`,
  `.claude/rules/moai/workflow/auto-semantics.md`, or their mirrors; the
  run-gate doc reconciliation is follow-up card material (spec.md §E).
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
  007/008/009/010/011/013/014/015/017/018/019/020/021/022), and grep-class
  RED cells with recorded exit codes for AC-ACE-002 plus AC-ACE-008's
  three export-path greps plus AC-ACE-022's production-path range-read
  (baselines in §C).

## §F Milestones

### M1 (Priority High) — verdict receipt + admission predicate extension

Data-model first: the receipt schema is the least reversible decision.

- Extend `.moai/docs/audit-artifact-convention.md` § What with the receipt
  line format (design.md §3): `convergence_overall: <pass|fail>` and a
  repeatable `required_backend: <backend> <pass|fail|inconclusive>` line,
  one per required backend.
- Receipt producer (D19 + D32): the plan-auditor agent body is the writer —
  its export step (`.claude/agents/moai/plan-auditor.md` § Output Format)
  appends the receipt lines to the exported verdict file per the convention
  § What, covering BOTH audit shapes REQ-ACE-008's trigger names:
  - multi-model audit: project the `audit_multi` convergence result it
    already receives (`ConvergenceResult.OverallVerdict` +
    `PerBackendVerdicts`, design.md §3) — the instruction carries the
    `PerBackendVerdicts` field name, the arm-distinctive phrase AC-ACE-008's
    third asserted count runs on (V4-D4);
  - single-model audit: write `convergence_overall` from its own verdict
    under the §3 projection rule ({PASS, PASS-WITH-DEBT} → pass;
    {FAIL, FAIL_WARNED} → fail; {INCONCLUSIVE} → inconclusive on the backend
    line, fail on convergence_overall — no silent upgrade; V4-D2) and one
    `required_backend:` line for the backend it actually ran, sourced
    from the named field of its own review output; a required backend the
    audit did not cover stays absent from the receipt and refuses under
    REQ-ACE-010 (correct fail-closed).
  The step lands in the deployed agent body AND its template mirror in the
  same change (research.md §3 edit target; the pre-existing whole-file drift
  keeps it on AC-ACE-014's known-FAIL carve-out list for the untouched
  regions).
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
- Update the TWO admission call sites (`decide.go` `planAuditCheck`,
  `homestate/card_evidence_readers.go` `admitVerdictFile`) to resolve the
  tree's configured gate set — the error-vs-empty contract of REQ-ACE-010 —
  and pass it to `Admit` (the LIVE seams of design.md §5; AC-ACE-022's
  per-seam arms — D2).
  `contract/rules.go:160` is deliberately NOT updated (D5 disposition): it
  is a label-only consumer (`auditverdict.AdmitLabel`; its own comment
  records that the field-level plan predicate runs at the kickoff evaluator
  and at T7), not an admission seam — a gate-set check cannot apply at a
  site that reads only the label, so including it in the call-site list
  would over-include and imply a bypass that does not exist.
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
  `plan-auditor.md` mirror drift stays untouched (Out of Scope).

### M3 (Priority High) — counter + ceiling-policy engine + enforcement

- `internal/runtime`: round-count derivation from iteration evidence
  (both families, one-iteration-once identity = the (SPEC id, iteration
  number) pair, unconditional — the D9-narrowed identity of REQ-ACE-001),
  the delta-eligibility check (fix_scope anchors + REQ/AC id sets + STOP;
  REQ-ACE-003), and the policy outcome engine (pass-through FIRST for a
  fully-admission-clean verdict at any ceiling state (REQ-ACE-013, D31;
  the tier-ceiling final-hit boundary included — V4-D1), then
  debt-admit / scope-split / hold-record for verdicts that fail admission;
  REQ-ACE-004..006).
- Enforcement wiring at the LIVE admission seams (the iter1 D4 finding —
  `GateConfig.Invoke` has no production caller, measured 0 non-test
  references): the kickoff evaluator (`decide.go`) and the homestate card
  transition call the engine before admission and surface the same refusal +
  outcome (exit nonzero / refusal record); `GateConfig.Invoke` gains the same
  Step-0 call as the library-level consumer for when a caller exists. An
  integration test proves a ceiling-hit round refuses at a production entry
  point at BOTH seams — AC-ACE-015 carries one arm per LIVE seam, the card
  transition and the kickoff evaluator (D23).
- Refusal output: structured (JSON or parseable lines) carrying outcome,
  reasons, evidence paths; persist to `progress.md`; audit-trail log append
  (REQ-ACE-007, REQ-ACE-012). Design the AuditResult extension per
  design.md §7 (separate outcome field, not a new Verdict enum value that
  the default branch would fold into INCONCLUSIVE); the outcome vocabulary
  is `pass-through | debt-admit | split | hold` (design.md §2).
- Override input for required-backend refusals (REQ-ACE-011) — explicit
  flag + note + logging only; the CLI writes the ack to `progress.md` §G
  Override and Refusal Record (outside the plan-artifact hash subject set)
  and the REQ-ACE-012 trail.
- Depends on: decision-index Q2/Q3/Q5 verdicts (blocker report if
  unresolved — the SPEC's embedded defaults stand while the questions stay
  open, each recorded kickoff-amendable).

### M4 (Priority Medium) — regression guard + mirror verification

- Template audit tests extended: region-scoped deployed-vs-mirror equality
  for every file this SPEC edits; the one file carrying pre-existing
  whole-file drift among this SPEC's edit targets (plan-auditor.md —
  measured DIFF at f2f815008 and re-measured DIFF at 69a085b2d) is named
  known-FAIL until repaired, never expected-pass (AC-ACE-014's carve-out).
- Static guard: no interactive prompt in the new CLI surface.
- Full lint + spec lint --strict 0/0; §E self-verification report.

## §G Anti-Patterns

- Do not edit `phase-execution.md` or `auto-semantics.md` — the D9 scope cut
  removed them; the run-gate doc reconciliation is follow-up card material.
- Do not reintroduce the audited-state dedupe resolution or a cross-card
  never-collapsed clause into REQ-ACE-001 — the D9-narrowed identity is the
  plain (SPEC id, iteration number) pair; the refinement is follow-up
  material.
- Do not count the daily run-history file (`.moai/reports/plan-audit/…`) as
  round evidence — iteration stream only.
- Do not add a second prose statement of the ceiling policy; C1 keeps one
  policy with the CLI as its machine consumer.
- Do not weaken `Admit`'s existing checks while adding refusal causes.
- Do not "fix" the pre-existing plan-auditor.md mirror drift inside this
  SPEC's commits beyond the keys/sections this SPEC edits.

## §H Cross-References

- spec.md §B (REQ-ACE-001..015), §C constraints, §D.1 id-history note
- acceptance.md §D (AC-ACE-001..022), §C edge cases
- design.md §1-§10 (counter model, receipt schema, enforcement, open points)
- research.md §1-§5 (source verification, Go surfaces, mirrors, gaps)
- decision-index.md Q0-Q6 (the D9 decision record, the disposition table,
  and the unresolved operator decisions)
