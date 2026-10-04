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
families. Identity for dedupe: the pair (SPEC id, iteration number N) — two
evidence files naming the same SPEC and the same N are one round; different
N are different rounds even when the verdict label, score, and audited hash
are identical (the repeated-audit-on-unchanged-artifacts case is exactly what
the ceiling exists to cap — a label+score+hash identity would collapse it to
one count and the ceiling would never fire). A file missing its iteration
number still counts (fail-counted: an unidentifiable iteration is evidence of
an iteration, not of its absence). SPEC attribution: a convention-family file
belongs to the SPEC named in its report header; the counter resolves a
SPEC's evidence from `.moai/reports/<SPEC-ID>/` plus every
`.moai/reports/<card-id>/` directory whose plan-audit iteration files name
that SPEC. The daily run-history file (`.moai/reports/plan-audit/…`) is
never counted (plan.md §G).

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

Delta eligibility (encoded — the prose policy's own condition, D6): at the
tier ceiling with a non-admitted verdict, the auto delta round is granted
only when the verdict lists `fix_scope` anchors, the diff between the two
audited SHAs stays inside them (plus `progress.md` and `.moai/reports/**`),
and the REQ/AC id sets are unchanged. An ineligible delta or a STOP signal is
a final hit: the outcomes below fire immediately at the tier ceiling.

Ladder (fixed by spec.md §B; NOT config keys — research.md §4), evaluated on
the refusal of REQ-ACE-003:

1. `debt-admit` — the verdict fails admission on the LABEL ALONE: score at
   or above the tier threshold, `must_pass_failed == 0`, `blocking_count ==
   0`, plan-artifact hash binding, no duplicate keys, and at least one
   finding to enumerate. The CLI writes the outcome record carrying
   PASS-WITH-DEBT and the verdict's findings as debts (`dispose_in` each);
   the consuming seam admits. The auditor's verdict file is never rewritten
   (that would duplicate decision keys). A hash-binding failure or a
   no-findings verdict is NEVER debt-admitted — it holds (D2: `Admit`
   refuses score and hash failures before any PASS-WITH-DEBT branch,
   verdict.go:208-233, so a rung admitted through them is unreachable; the
   label-only failure is the one shape the existing predicate genuinely
   refuses that the policy may convert, and downstream label-only consumers
   read `PASS-WITH-DEBT`, a passing label per `AdmitLabel`).
2. `split` — blocking findings exist and every one carries a scoped fix
   anchor: hold record + split proposal naming the anchor scope; blocked.
3. `hold` — otherwise (hash mismatch, no findings, unanchored blocking
   findings, missing fields): hold record; blocked.

Refusal point (D4): `GateConfig.Invoke` is production-dead — 0 non-test
references, 17 test-only (measured, research.md §2) — so a refusal wired only
there refuses nothing. The engine is called at the two LIVE admission seams —
the kickoff evaluator (`internal/contract/kickoff/decide.go:373-376`) and the
homestate card transition (`internal/homestate/card_evidence_readers.go`
`admitVerdictFile`) — before `Admit`; `GateConfig.Invoke` gains the same
Step 0 as the library-level consumer for when a production caller exists. An
integration test proves a ceiling-hit round refuses at a production entry
point (AC-ACE-022), so the seam cannot silently go dead the way `Invoke` did
(verification-completeness §1.3 liveness).

## §3 Verdict receipt schema (REQ-ACE-008)

Two machine-readable line shapes, appended to the exported plan-audit verdict
file per the convention's existing "machine lines" pattern:

```
convergence_overall: <pass|fail>
required_backend: <backend-name> <pass|fail|inconclusive>   # repeatable, one line per required backend
```

- Names follow the existing parser conventions: case-insensitive keys,
  `key: value`.
- Per-key repeat rule (D18): `convergence_overall` is single-occurrence (a
  second, differing occurrence is a duplicate-key inadmissibility under the
  existing rule); `required_backend` is repeatable BY DESIGN — one line per
  required backend. A second line for an already-recorded backend is a
  duplicate when the recorded verdict differs and an admissible identical
  repeat when it does not (the report-header-plus-machine-line pattern the
  existing duplicate rule already admits). `Admit` refuses a CONFIGURED
  required backend with no line at all (D3 — receipts record every required
  backend's state, so an absent or inconclusive required backend is
  distinguishable from a passing one).
- The projection is mechanical: `ConvergenceResult.OverallVerdict` and the
  `PerBackendVerdicts` entries whose `Gate == required`, projecting each
  entry's `Backend` and `Verdict` (mcp_convergence.go:92-99 — the verdict
  vocabulary is already pass|fail|inconclusive). No new convergence
  semantics.
- The convention doc's export mandate extends to these lines whenever the
  tree resolves a non-empty required-backend set (REQ-ACE-008's trigger — a
  single-required-backend tree emits receipts too, which is what keeps
  REQ-ACE-010 from permanently blocking it); a tree resolving no required
  backend omits them and admits without them (C4).

## §4 Admission predicate extension (REQ-ACE-009, REQ-ACE-010)

`auditverdict.Admit` gains a required-backend parameter (the tree's
configured gate set, resolved by each call site from `workflow.audit.gates` /
`audit.model` resolution — the same resolution `resolveAuditGates` in
`internal/cli/mcp_worktree_root.go` already performs). Evaluation order after
the existing checks:

1. Receipt declares a required backend fail OR inconclusive → refuse, reason
   names the backend (label-independent — REQ-ACE-009).
2. Required backend(s) configured and no receipt present, or a receipt
   missing a configured required backend's line → refuse, fail-closed
   default (REQ-ACE-010; decision-index Q4). No required backend configured
   + no receipt → admit (C4).
3. Malformed or duplicate receipt keys → refuse (extends the existing
   duplicate-key rule to the new keys; design.md §3 states the per-key
   repeat rule).

Override (REQ-ACE-011) is NOT an `Admit` input — admission stays pure. The
override lives at the CLI seam: an explicit flag/env carrying
`backend + note` marks the refusal acknowledged, the CLI writes the ack to
`progress.md` §G Override and Refusal Record (a fresh section letter outside
the plan-artifact hash subject set — never `decision-index.md`, whose edit
would break the hash binding of every later verdict; D10) and to the
REQ-ACE-012 trail, and the call site then proceeds with the recorded
exception. `Admit` never silently passes a required-backend fail.

## §5 Enforcement seams and the audit trail (REQ-ACE-012)

| Seam | File | Behavior at ceiling / required-backend fail |
|---|---|---|
| Kickoff evaluator (LIVE) | `internal/contract/kickoff/decide.go:373-376` | engine before `Admit`; refusal carries the receipt/ceiling reason; exit nonzero via kickoff-check |
| Card transition (LIVE) | `internal/homestate/card_audit_kickoff.go` + `card_evidence_readers.go` | engine before `Admit`; same refusal reasons through the shared predicate |
| Run gate (library) | `internal/runtime/audit_gate.go` | the same Step 0 for when a caller exists — `GateConfig.Invoke` has no production caller today (0 non-test references, 17 test-only; research.md §2), which is why the LIVE rows above carry the ACs |
| Audit trail | new `.moai/state/audit-enforcement.log` (append-only) | one line per refusal/override: timestamp, SPEC, kind, outcome/reason |

The trail is machine-local state (`.moai/state/`), gitignored, never a
citation target — the durable record lives in progress.md §G Override and
Refusal Record and the exported report, per the evidence-export doctrine.

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
`defaults.go`, symmetry cases in a new bare `TestStructYAMLSymmetry` harness
covering `harness.yaml` (the exact selector AC-ACE-002 runs; only
`_`-suffixed variants exist today); the two orphan entries removed from
`internal/config/loader.go`'s acknowledged list. `OnFinalHit` is validated on
load — the accepted value is `hold-and-split`, the only value the prose
policy and this engine implement; any other value is a config error, because
a reader that silently accepts a policy name it does not enforce would read
the key while ignoring its meaning (D17 — the reader exists to validate the
policy the CLI enforces). harness.yaml content is unchanged (no new keys),
so M2 performs no harness.yaml mirror sync.

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
the trail line and the ack text lands in progress.md §G Override and Refusal
Record + the REQ-ACE-012 trail (D10 — never `decision-index.md`).
No equivalent for the ceiling refusal: the ceiling has no override — its
outcomes ARE the decisions (the card's 질문 없음 requirement).

## §9 Doc changes (REQ-ACE-013, REQ-ACE-014)

- phase-execution.md Step 4c: FAIL after grace → the gate blocks; the text
  names the ceiling-policy path as the only non-block exit and the
  fail-closed rule of auto-semantics §7 as the owner. Step 4d: INCONCLUSIVE
  → fail-closed, record + escalate per §7's ladder; "max 3 retries total"
  (a second prose ceiling) is replaced by the machine counter reference.
- auto-semantics.md §9: the 11 rows of spec.md §D.2, each
  `<gate> | <disposition from the existing vocabulary> — classified from
  <file §section>`. Row 2 REPLACES the existing "plan-audit bypass flags |
  RETIRED" row with the reconciled wording naming the machine enforcement
  (counter + admission refusal): 10 existing rows − 1 + 11 = 20.

## §10 Open points → decision-index

| Design point | Row |
|---|---|
| Effective-ceiling arithmetic (tier + delta) | Q2 |
| Round-count evidence source | Q3 |
| Receipt-absent disposition | Q4 |
| Debt-admit eligibility mapping | Q5 |
| §9 row set | Q6 |
| Delta-round count (1) | Q1 (settled — POLICY-COVERED) |
