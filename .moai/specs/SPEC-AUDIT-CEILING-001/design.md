# design.md — SPEC-AUDIT-CEILING-001

System design for the three work items. WHAT/WHY lives in spec.md; this file
fixes the shape of the data model and the enforcement seams so run-phase
review focuses on the decisions most likely to change.

## §1 Round-count evidence model (REQ-ACE-001)

Two iteration-file families exist on disk and both are evidence:

- Convention family: `.moai/reports/<card-id>/plan-audit-iter<N>.md` (and
  `.moai/reports/<SPEC-ID>/` for SPEC-scoped audits) — mandated one file per
  iteration by `.moai/docs/audit-artifact-convention.md` § Where.
- Legacy stream: `<reportDir>/<SPEC-ID>-review-<N>.md` — enumerated today by
  `internal/runtime/audit_review.go` `ResolveLatestPlanAudit`.

Count rule: the round count is the number of DISTINCT iterations across both
families. Identity for dedupe: (verdict label, overall score, audited
commit/plan-artifact hash) — the three fields every iteration file carries or
that `Parse` already reads. A file missing any identity field still counts
(fail-counted: an unidentifiable iteration is evidence of an iteration, not
of its absence). The daily run-history file (`.moai/reports/plan-audit/…`)
is never counted (plan.md §G).

Proposed API (internal/runtime):

```go
type RoundEvidence struct {
    Count    int      // distinct iterations
    Sources  []string // file paths consulted
    Latest   *auditverdict.Fields // parsed latest iteration, nil if none
}
func CountAuditRounds(specID string, reportDirs []string) (RoundEvidence, error)
```

## §2 Ceiling-policy outcome engine (REQ-ACE-003..007)

Inputs: `RoundEvidence`, the tier ceiling (M2 structs), the policy values
(`auto_delta_rounds`, `on_final_hit`), and the latest verdict's
`auditverdict.Fields` plus its blocking findings' anchor annotations
(`fix_scope` presence — already authored by plan-auditor.md § Retry Loop
Contract for delta eligibility).

Ladder (fixed by spec.md §B; NOT config keys — research.md §4):

1. `debt-admit` — latest non-admitted verdict has `must_pass_failed == 0 &&
   blocking_count == 0` (only the score threshold or hash binding failed):
   record PASS-WITH-DEBT (debts enumerated from the verdict's findings with
   `dispose_in`), admit.
2. `split` — blocking findings exist and every one carries a scoped fix
   anchor: hold record + split proposal naming the anchor scope; blocked.
3. `hold` — otherwise: hold record; blocked.

Refusal point: `GateConfig.Invoke` gains a Step 0 before the bypass path's
auditor spawn decision and before Step 3; at or above the effective ceiling
(tier ceiling + auto_delta_rounds) it returns the outcome instead of invoking
the auditor. The same evaluation is callable from `internal/contract/kickoff`
and `internal/homestate` so all three enforcement seams answer identically
(one engine, three callers — mirroring the one-admission-predicate pattern of
`internal/auditverdict`).

## §3 Verdict receipt schema (REQ-ACE-008)

Two machine-readable line shapes, appended to the exported plan-audit verdict
file per the convention's existing "machine lines" pattern:

```
convergence_overall: <pass|fail>
required_backend_fail: <backend-name>     # repeatable, one per failed required backend
```

- Names follow the existing parser conventions: case-insensitive keys,
  `key: value`, single occurrence of `convergence_overall` (a second,
  differing occurrence is a duplicate-key inadmissibility under the existing
  rule).
- The projection is mechanical: `ConvergenceResult.OverallVerdict` and the
  `PerBackendVerdicts` entries whose `Gate == required` and `Verdict == fail`
  (mcp_convergence.go). No new convergence semantics.
- The convention doc's export mandate extends to these lines when the audit
  fanned out to multiple backends; a single-backend audit omits them (and
  with no required backends configured, absence refuses nothing).

## §4 Admission predicate extension (REQ-ACE-009, REQ-ACE-010)

`auditverdict.Admit` gains a required-backend parameter (the tree's
configured gate set, resolved by each call site from `workflow.audit.gates` /
`audit.model` resolution — the same resolution `resolveAuditGates` in
`internal/cli/mcp_worktree_root.go` already performs). Evaluation order after
the existing checks:

1. Receipt declares a required backend fail → refuse, reason names the
   backend (label-independent — REQ-ACE-009).
2. Required backend(s) configured and no receipt present → refuse,
   fail-closed default (REQ-ACE-010; decision-index Q4).
3. Malformed or duplicate receipt keys → refuse (extends the existing
   duplicate-key rule to the new keys).

Override (REQ-ACE-011) is NOT an `Admit` input — admission stays pure. The
override lives at the CLI seam: an explicit flag/env carrying
`backend + note` marks the refusal acknowledged, the ack is persisted
(progress.md + decision record + REQ-ACE-012 trail) and the call site then
proceeds with the recorded exception. `Admit` never silently passes a
required-backend fail.

## §5 Enforcement seams and the audit trail (REQ-ACE-012)

| Seam | File | Behavior at ceiling / required-backend fail |
|---|---|---|
| Run gate | `internal/runtime/audit_gate.go` | Step 0 refusal + outcome; no auditor spawn |
| Contract kickoff | `internal/contract/kickoff/decide.go` | admission refusal carries the receipt/ceiling reason; exit nonzero via kickoff-check |
| Card transition | `internal/homestate/card_audit_kickoff.go` + `card_evidence_readers.go` | same refusal reasons through the shared predicate |
| Audit trail | new `.moai/state/audit-enforcement.log` (append-only) | one line per refusal/override: timestamp, SPEC, kind, outcome/reason |

The trail is machine-local state (`.moai/state/`), gitignored, never a
citation target — the durable decision record lives in progress.md and the
report, per the evidence-export doctrine.

## §6 Config reader (REQ-ACE-002)

```go
type PlanAuditTierCeilingsConfig struct {
    S int `yaml:"S"`
    M int `yaml:"M"`
    L int `yaml:"L"`
}
type PlanAuditCeilingPolicyConfig struct {
    AutoDeltaRounds int    `yaml:"auto_delta_rounds"`
    OnFinalHit      string `yaml:"on_final_hit"`
}
```

Wired into the harness config load path (`LoadHarnessConfig`), defaults in
`defaults.go`, symmetry cases in `audit_struct_yaml_symmetry_test.go`; the
two orphan entries removed from `internal/config/loader.go`'s acknowledged
list. harness.yaml content is unchanged (no new keys).

## §7 AuditResult extension (no new Verdict value)

Adding a `VerdictCeilingBlocked` enum value is rejected: consumers switch on
known values and the default branch folds unknowns into INCONCLUSIVE
(audit_gate.go Step 4) — a refusal that silently downgrades. Instead:

```go
type AuditResult struct {
    // ... existing fields ...
    Ceiling *CeilingOutcome `json:"ceiling,omitempty"` // set only on refusal
}
type CeilingOutcome struct {
    Outcome   string   `json:"outcome"`   // debt-admit | split | hold
    Reasons   []string `json:"reasons"`
    Evidence  []string `json:"evidence"`  // iteration file paths
    Blocked   bool     `json:"blocked"`   // run-entry stays closed
}
```

`Invoke` returns the result with `Verdict` set to the honest underlying
state (FAIL/INCONCLUSIVE as measured) and `Ceiling` carrying the policy
result; refusal is observable via `Ceiling != nil && Ceiling.Blocked`.

## §8 Override surface (REQ-ACE-011)

One explicit input, two spellings (flag preferred, env for non-argv paths):

- `--ack-required-backend=<backend> --ack-note="<note>"`
- `MOAI_ACK_REQUIRED_BACKEND=<backend>`, `MOAI_ACK_NOTE=<note>`

Empty note → refuse the override (edge case §C.8). Every acceptance appends
the trail line and the ack text lands in progress.md + the decision record.
No equivalent for the ceiling refusal: the ceiling has no override — its
outcomes ARE the decisions (the card's 질문 없음 requirement).

## §9 Doc changes (REQ-ACE-013, REQ-ACE-014)

- phase-execution.md Step 4c: FAIL after grace → the gate blocks; the text
  names the ceiling-policy path as the only non-block exit and the
  fail-closed rule of auto-semantics §7 as the owner. Step 4d: INCONCLUSIVE
  → fail-closed, record + escalate per §7's ladder; "max 3 retries total"
  (a second prose ceiling) is replaced by the machine counter reference.
- auto-semantics.md §9: the 11 rows of spec.md §D.3, each
  `<gate> | <disposition from the existing vocabulary> — classified from
  <file §section>`. The existing "plan-audit bypass flags | RETIRED" row is
  updated to name the machine enforcement (counter + admission refusal).

## §10 Open points → decision-index

| Design point | Row |
|---|---|
| Effective-ceiling arithmetic (tier + delta) | Q2 |
| Round-count evidence source | Q3 |
| Receipt-absent disposition | Q4 |
| Debt-admit eligibility mapping | Q5 |
| §9 row set | Q6 |
| Delta-round count (1) | Q1 (settled — POLICY-COVERED) |
