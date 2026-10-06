# design.md — SPEC-AUDIT-CEILING-001

System design for the narrowed work items (v0.4.0, per operator decision D9).
WHAT/WHY lives in spec.md; this file fixes the shape of the data model and
the enforcement seams so run-phase review focuses on the decisions most
likely to change. The v0.4.0 scope cut removed the doc-reconciliation design
(§9 of v0.3.0 — phase-execution.md / auto-semantics.md edits) and the
audited-state dedupe resolution (§1 of v0.3.0); both moves are recorded in
the sections below and in decision-index.md's disposition table.

## §1 Round-count evidence model (REQ-ACE-001)

Two iteration-file families exist on disk and both are evidence:

- Convention family: `.moai/reports/<card-id>/plan-audit-iter<N>.md` (and
  `.moai/reports/<SPEC-ID>/` for SPEC-scoped audits) — mandated one file per
  iteration by `.moai/docs/audit-artifact-convention.md` § Where.
- Legacy stream: `<reportDir>/<SPEC-ID>-review-<N>.md` — enumerated today by
  `internal/runtime/audit_review.go` `ResolveLatestPlanAudit`.

Count rule: the round count is the number of DISTINCT (SPEC id, iteration
number N) pairs across both families. Identity for dedupe: the pair (SPEC
id, N) — two evidence files naming the same SPEC and the same N are one
round regardless of which family or card directory recorded them; different
N are different rounds even when the verdict label, score, and audited hash
are identical (the repeated-audit-on-unchanged-artifacts case is exactly
what the ceiling exists to cap). A file missing its iteration number still
counts (fail-counted: an unidentifiable iteration is evidence of an
iteration, not of its absence).

Known accepted limitation (D9 scope cut — former v0.3.0 cross-card rule and
iter3 D33 deferred): because the identity is the bare (SPEC, N) pair, a
no-repair re-audit recorded under a NEW card directory at a reused
iteration number (same tree → same audited state, new card → iter1)
collapses into the earlier round and does not advance the count — the
ceiling is blind to audit-without-repair churn across cards. The v0.3.0
audited-state resolution that addressed one half of this was removed by the
scope cut because it could not resolve the same-SHA case without
contradicting the never-collapsed clause (iter3 D33); the refinement — a
run-distinguishing identity component, or collapse gated on explicit
double-export evidence — is follow-up card material. Within one card's own
audit stream (the case the ceiling policy governs), different N always
count, so the policy's own retry loop is capped exactly as configured.
The count saturates at the number of distinct (SPEC, N) pairs — equal to
max-N per SPEC only under contiguous numbering (non-contiguous N values,
e.g. {1, 3}, give a count below max-N; V4-O1).

SPEC attribution: a convention-family file belongs to the SPEC named in its
report header; the counter resolves a SPEC's evidence from
`.moai/reports/<SPEC-ID>/` plus every `.moai/reports/<card-id>/` directory
whose plan-audit iteration files name that SPEC. The daily run-history file
(`.moai/reports/plan-audit/…`) is never counted (plan.md §G).

Proposed API (internal/runtime):

```go
type RoundEvidence struct {
    Count    int      // distinct (SPEC, N) pairs
    Sources  []string // file paths consulted
    Latest   *auditverdict.Fields // parsed latest iteration, nil if none
}
func CountAuditRounds(specID string, reportDirs []string) (RoundEvidence, error)
```

## §2 Ceiling-policy outcome engine (REQ-ACE-003..007, REQ-ACE-013)

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

Ladder (fixed by spec.md §B; NOT config keys — research.md §4), evaluated
when a verdict is presented at a LIVE admission seam for a SPEC at any
ceiling state — the effective ceiling reached (tier + delta), or the tier
ceiling reached on a final hit with no eligible delta round (V4-D1):

0. `pass-through` (D31, REQ-ACE-013) — the verdict satisfies every
   admission check of the shared predicate (label, score, must-pass,
   blocking, hash binding, duplicate keys, and the §4 receipt checks). The
   CLI admits the round and records a ceiling-reached outcome naming that
   the ceiling was reached; no question is asked. The ceiling caps
   repetition, not a healthy result — a healthy resume presented at any
   ceiling state (both boundaries, per V4-D1's second face: the
   tier-ceiling final hit included) is never held by omission, which would
   force a mandatory operator question on exactly the population §A
   measures (35% of SPECs audited 3+ times) and contradict the SPEC's own
   no-question purpose. This rung evaluates FIRST, so the refusal rungs
   below see only verdicts that fail admission — REQ-ACE-003's refusal is
   scoped accordingly (a verdict satisfying every admission check follows
   REQ-ACE-013, never a refusal). The outcome record names the SPECs
   admitted this way so the population stays auditable.

For a verdict that FAILS admission, evaluated on the refusal of REQ-ACE-003:

1. `debt-admit` — the verdict fails admission on the LABEL ALONE: score at
   or above the tier threshold, `must_pass_failed == 0`, `blocking_count ==
   0`, plan-artifact hash binding, no duplicate keys, no REQ-ACE-009/010
   receipt refusal, and at least one finding to enumerate. The CLI writes
   the outcome record carrying PASS-WITH-DEBT and the verdict's findings as
   debts (`dispose_in` each); the consuming seam admits. The auditor's
   verdict file is never rewritten (that would duplicate decision keys). A
   hash-binding failure or a no-findings verdict is NEVER debt-admitted — it
   holds (D2: `Admit` refuses score and hash failures before any
   PASS-WITH-DEBT branch, verdict.go:208-233, so a rung admitted through them
   is unreachable; the label-only failure is the one shape the existing
   predicate genuinely refuses that the policy may convert, and downstream
   label-only consumers read `PASS-WITH-DEBT`, a passing label per
   `AdmitLabel`).

Consuming-seam admission semantics (D20): the seam re-runs the full `Admit`
— the §4 receipt checks included — with exactly one conversion: a debt-admit
outcome substitutes the outcome's PASS-WITH-DEBT for the raw verdict label
in the label check alone. Every other check (score, must-pass, blocking,
hash, duplicate keys, and the §4 receipt checks) evaluates the raw verdict
and keeps its refusing force. A verdict carrying a receipt refusal is
excluded from debt-admit eligibility (REQ-ACE-004), so work item 1 cannot
admit what work item 2 refuses: the combination (ceiling hit + label-only
failure + required-backend fail) holds (AC-ACE-004's negative arm,
`TestCeilingPolicyReceiptHold`). The pass-through rung needs no conversion:
a verdict that clears the full predicate is admitted by the seam's own
predicate evaluation; the engine's only act is recording the
ceiling-reached outcome (AC-ACE-013, `TestCeilingPolicyPassThrough`).

2. `split` — blocking findings exist and every one carries a scoped fix
   anchor: hold record + split proposal naming the anchor scope; blocked.
3. `hold` — otherwise (hash mismatch, no findings, unanchored blocking
   findings, missing fields): hold record; blocked.

Refusal point (D4): `GateConfig.Invoke` is production-dead — 0 non-test
references, 17 test-only (measured at 2f492df19 and re-measured at
69a085b2d, research.md §2) — so a refusal wired only there refuses nothing.
The engine is called at the two LIVE admission seams — the kickoff evaluator
(`internal/contract/kickoff/decide.go:373-376`) and the homestate card
transition (`internal/homestate/card_evidence_readers.go`
`admitVerdictFile`) — before `Admit`; `GateConfig.Invoke` gains the same
Step 0 as the library-level consumer for when a production caller exists. An
integration test proves a ceiling-hit round refuses at a production entry
point at both seams (AC-ACE-015, one arm per LIVE seam), so neither can
silently go dead the way `Invoke` did (verification-completeness §1.3
liveness). REQ-ACE-003's trigger is accordingly the admission-seam event —
the moment a produced verdict is presented for admission — not the
auditor's private spawn decision: the prose plan-audit loop self-governs
below the ceiling, and this engine adjudicates the machine seam (D23).

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
- Producer paths (D19 + D32 — both shapes REQ-ACE-008's trigger names):
  - Multi-model audit: the projection is mechanical from the convergence
    result the auditor already receives — `ConvergenceResult.OverallVerdict`
    and the `PerBackendVerdicts` entries whose `Gate == required`,
    projecting each entry's `Backend` and `Verdict` (mcp_convergence.go:92-99
    — the verdict vocabulary is already pass|fail|inconclusive). No new
    convergence semantics.
  - Single-model audit (D32): the deployed auditor contract runs its own
    review in single claude/glm/codex modes and calls no convergence tool,
    so a single-model producer fed only from `ConvergenceResult` would leave
    required-resolving trees with no writer and a permanent REQ-ACE-010
    block. The single-model path instead writes `convergence_overall` from
    the auditor's own verdict under the projection rule below and one
    `required_backend:` line for the backend it actually ran, sourced from
    the named field of its own review output. A required backend the audit
    did not cover stays absent from the receipt and refuses under
    REQ-ACE-010 — correct fail-closed, never weakened by this extension.
- Own-verdict projection rule (V4-D2 — defined for EVERY label the
  auditor's own verdict enum can carry, so no projection is undefined):
  the auditor's own verdict vocabulary is {PASS, PASS-WITH-DEBT, FAIL,
  FAIL_WARNED, INCONCLUSIVE} (BYPASSED never reaches the export step — a
  bypassed audit exports no verdict file, hence no receipt). The projection
  onto the receipt vocabulary:
  - `PASS` and `PASS-WITH-DEBT` → backend line `pass`, and
    `convergence_overall: pass` — PASS-WITH-DEBT is a passing label per the
    shared predicate's `AdmitLabel` (design.md §2), so mapping it to `pass`
    is not an upgrade; the raw own label stays readable in the verdict
    body, which the receipt never rewrites.
  - `FAIL` and `FAIL_WARNED` → backend line `fail`,
    `convergence_overall: fail`.
  - `INCONCLUSIVE` → backend line `inconclusive`,
    `convergence_overall: fail` (an unresolved audit is not a pass — the
    same fail-closed posture §7 records; the resulting receipt refuses
    under REQ-ACE-009/010 either way).
  A receipt line carrying a value outside `pass|fail|inconclusive` (e.g. a
  hand-written or future label written verbatim) stays malformed and
  refuses under the existing malformed-line rule — the projection exists so
  the producer never emits one, not so the parser learns to accept one.
  AC-ACE-008's `TestParseReceiptPassWithDebtProjection` covers the
  PASS-WITH-DEBT single-model receipt end to end.
- The convention doc's export mandate extends to these lines whenever the
  tree resolves a non-empty required-backend set (REQ-ACE-008's trigger —
  a single-required-backend tree emits receipts under either producer path,
  which is what keeps REQ-ACE-010 from permanently blocking it); a tree
  resolving no required backend omits them and admits without them (C4).

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

Config-error disposition (D21): `resolveAuditGates` today returns an empty
gate set when `ResolveAuditPlan` errors (mcp_worktree_root.go:127-130 —
re-read and confirmed at 69a085b2d, plan.md §C) — fail-open against §7's
posture. The M1 call sites resolve with the opposite disposition: the
resolution result distinguishes an error from a genuinely-empty
configuration, and an audit section that exists but cannot be read or parsed
refuses (REQ-ACE-010's third trigger arm). Config absent → C4 admission
without a receipt; config error → refuse. The two paths are distinct values
end to end — the error is never folded into the empty set on an admission
path.

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
    Ceiling *CeilingOutcome `json:"ceiling,omitempty"` // set on ceiling evaluation
}
type CeilingOutcome struct {
    Outcome   string   `json:"outcome"`   // pass-through | debt-admit | split | hold
    Reasons   []string `json:"reasons"`
    Evidence  []string `json:"evidence"`  // iteration file paths
    Blocked   bool     `json:"blocked"`   // run-entry stays closed
}
```

`Invoke` returns the result with `Verdict` set to the honest underlying
state (FAIL/INCONCLUSIVE as measured) and `Ceiling` carrying the policy
result; refusal is observable via `Ceiling != nil && Ceiling.Blocked`. A
`pass-through` outcome carries `Blocked == false` and records that the
ceiling was reached with an admission-clean verdict (REQ-ACE-013).

## §8 Override surface (REQ-ACE-011)

One explicit input, two spellings (flag preferred, env for non-argv paths):

- `--ack-required-backend=<backend> --ack-note="<note>"`
- `MOAI_ACK_REQUIRED_BACKEND=<backend>`, `MOAI_ACK_NOTE=<note>`

Empty note → refuse the override (edge case §C.8). Every acceptance appends
the trail line and the ack text lands in progress.md §G Override and Refusal
Record + the REQ-ACE-012 trail (D10 — never `decision-index.md`).
No equivalent for the ceiling refusal: the ceiling has no override — its
outcomes ARE the decisions (the card's 질문 없음 requirement). A hold is
never terminal-silent: the hold record names its release path (REQ-ACE-006)
— the split/new-SPEC route of REQ-ACE-005, or an operator decision recorded
in progress.md §G (D30).

## §9 Doc reconciliation — REMOVED from scope (D9)

The v0.3.0 §9 (phase-execution.md Step 4 rewrite + auto-semantics.md §9
11-row inventory expansion, former REQ-ACE-013/014) was removed by operator
decision D9: this SPEC no longer edits either file or their mirrors. The
run-gate doc reconciliation is follow-up card material; its design notes are
preserved in the v0.3.0 tree (`git show 9dd4d5c74:.moai/specs/SPEC-AUDIT-CEILING-001/design.md`)
for the follow-up author. No AC of this SPEC reads either file.

## §10 Open points → decision-index

| Design point | Row |
|---|---|
| Effective-ceiling arithmetic (tier + delta) | Q2 |
| Round-count evidence source | Q3 |
| Receipt-absent disposition | Q4 |
| Debt-admit eligibility mapping | Q5 |
| Pass-through posture at the ceiling | settled in-scope per D9's no-question mandate (REQ-ACE-013; recorded in the disposition table) |
| Delta-round count (1) | Q1 (settled — POLICY-COVERED) |
| §9 row set | removed with the D9 scope cut (former Q6) |
