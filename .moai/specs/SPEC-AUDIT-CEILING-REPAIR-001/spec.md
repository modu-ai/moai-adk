---
id: SPEC-AUDIT-CEILING-REPAIR-001
title: "Audit-ceiling engine defect repair — legacy-family latest-verdict resolution and debt-inventory persistence"
version: "0.3.0"
status: draft
created: 2026-10-07
updated: 2026-10-07
author: "Goos Kim"
priority: P1
phase: "v3.2.0 target"
module: "internal/runtime"
lifecycle: spec-anchored
tags: "audit-ceiling, defect-repair, delta-round, debt-record, fail-closed"
tier: M
related_specs: [SPEC-AUDIT-CEILING-001, SPEC-AUDIT-CEILING-002]
---

# SPEC-AUDIT-CEILING-REPAIR-001 — Audit-ceiling engine defect repair

## §A Background and Motivation

Card t1560 (lane-2, run tmhxo0, in-lane plan→run→sync) repairs two defects in
the audit-ceiling outcome engine (`internal/runtime/audit_ceiling.go`, the
SPEC-AUDIT-CEILING-001 deliverable that reached main via card t1500's chain).
Both defects were measured by the dispatching lane and re-verified line-by-line
against source at `903ccd028` in this tree before this SPEC was authored.

**D1 — `previousAuditedSHA` early exit on a legacy-family latest verdict**
(`internal/runtime/audit_ceiling.go:274-277`). The function resolves the latest
round number through `iterationOf`, which parses ONLY the convention family
(`conventionFile`, `internal/runtime/audit_counter.go:46` — the
`plan-audit[-iter<N>].md` shapes). The round counter itself counts BOTH
families, and a legacy-family file (`<SpecID>-review-<N>.md`, the card-review
F4 stream) can be the selected `RoundEvidence.LatestPath`
(`audit_counter.go:89,115-116,145-148`). The previous-round scan inside
`previousAuditedSHA` already parses legacy numbers for non-latest sources
(`audit_ceiling.go:288-292`) — the latest file is the one name the dual-family
parsing never reaches. With a legacy-family latest file, `iterationOf` returns
0, the `latestN <= 0` guard aborts, and the previous audited SHA is never
found. The delta round's diff baseline is therefore lost:
`EvaluateCeiling` skips the diff-in-anchors and REQ/AC-set checks
(`audit_ceiling.go:103-108`), `DeltaEligible` sees `false, false`, and the
round becomes a final hit instead of a delta round. The existing test
`TestPreviousAuditedSHALegacyPriorRound` covers legacy as the PRIOR round
only; legacy as the LATEST round is the uncovered face of the same dual-family
contract.

**D2 — `persistOutcome` drops the debt inventory**
(`internal/runtime/audit_ceiling.go:461-471`). The record line persistOutcome
writes to the SPEC's `progress.md` §G Override and Refusal Record (the durable
carrier, SPEC-AUDIT-CEILING-001 REQ-ACE-007/012) and the line it appends to the
machine-local audit trail (`.moai/state/audit-enforcement.log`) serialize kind,
outcome, reasons (%q), and evidence — but not `oc.Debts`. The REQ-ACE-004
debt-admit rung constructs the outcome with `Debts: fields.Debts`
(`audit_ceiling.go:144`; `auditverdict.Debt{ID, DisposeIn, Description}`,
`internal/auditverdict/verdict.go:34-39`), so a debt-admit admission is
recorded durably WITHOUT the inventory it admitted — the durable carrier
records nothing about WHICH debts were accepted. This contradicts the record
posture SPEC-AUDIT-CEILING-001 REQ-ACE-004 itself states: the CLI records
"PASS-WITH-DEBT with the verdict's findings enumerated as debt lines (each
carrying `dispose_in`), written by the CLI into its outcome record". The
sibling JSON path (`EvaluatePlanAuditCeiling`/`RecordCeilingOutcome`,
SPEC-AUDIT-CEILING-002) already carries `DebtIDs` — the §G/trail path written
by `persistOutcome` is the inconsistent one.

**D3 — `appendProgressRecord` unlocked read-modify-write**
(`internal/runtime/audit_ceiling.go:479-496`, the write at `:495`; leader
scope extension, 2026-10-07). The §G record helper reads progress.md,
rebuilds the content in memory, and writes it back in place with no lock and
no atomic replace. Concurrent `persistOutcome` calls on one SPEC overwrite
each other with stale content and can destroy pre-existing progress content:
the turn-end codex review gate measured 24 parallel record calls on
`903ccd028` surviving as 7, with the pre-existing §G heading lost.

**D4 — `CountAuditRounds` collapses an unparseable iteration number into
round 1** (`internal/runtime/audit_counter.go:107-114`, the Atoi at `:111`;
leader ruling #2 fold-in, 2026-10-07). A convention-family file whose numeric
suffix overflows the integer range (e.g. `plan-audit-iter99999999999999999999.md`)
fails `strconv.Atoi`, and `n` silently keeps its initialized value 1 — the
same round number as a bare `plan-audit.md`. The two files dedupe into ONE
counted round (`audit_counter.go:144`), undercounting the ceiling input. The
counter's own fail-counted convention (the `conventionUnnumbered` branch,
`audit_counter.go:49,117-119`) already treats a non-numeric suffix as its own
round; the numeric-but-overflowing suffix is the one spelling that collapses.

## §B Requirements (GEARS)

- **REQ-ACR-001** (When the latest audit-round evidence file the round counter
  selects — `RoundEvidence.LatestPath` — carries a legacy-family file name
  (`<SpecID>-review-<N>.md`), the ceiling engine shall resolve the latest
  round number through the same dual-family parsing the previous-round scan
  applies — the convention family `plan-audit[-iter<N>].md` OR the legacy
  family): The engine shall resolve the previous audited SHA across both
  evidence families, and shall not report no-prior-round solely because the
  latest evidence file's name belongs to the legacy family. The genuinely
  unparseable-name semantics of REQ-ACR-002 are outside this requirement's
  subject.

- **REQ-ACR-002** (When the latest evidence file's name parses under neither
  family, or its number does not parse as a positive integer, the ceiling
  engine shall keep the no-prior-round result): An unparseable or non-positive
  latest round number never yields a guessed, defaulted, or sibling-derived
  diff baseline — the `latestN <= 0` fail-closed semantics are preserved, and
  the round stays a final hit.

- **REQ-ACR-003** (When a ceiling evaluation in a delta-eligible state
  presents a legacy-family latest verdict while every eligibility condition
  the prose policy names holds — fix_scope anchors present, no STOP signal,
  the diff between the two audited SHAs inside the anchors, and the REQ/AC id
  set unchanged): The engine shall grant the delta round on the same
  conditions it would grant for a convention-family latest verdict. Delta
  eligibility shall not depend on the latest evidence file's name family; with
  the defect, this evaluation produced a final-hit outcome instead.

- **REQ-ACR-004** (When the ceiling engine persists a ceiling outcome whose
  debt inventory is non-empty — the REQ-ACE-004 debt-admit outcome constructed
  with the verdict's enumerated debts): The record lines the persistence call
  writes — the durable `progress.md` §G record and the machine-local
  audit-trail entry — shall each carry the admitted debt inventory, enumerating
  every admitted debt's ID and `dispose_in` (with its description), as one
  additive field in the record's existing machine-parseable one-line
  `key=value` grammar. The durable carrier states WHICH debts were admitted
  (the record posture SPEC-AUDIT-CEILING-001 REQ-ACE-004 already requires of
  the outcome record).

- **REQ-ACR-005** (When an admitted debt's fields carry characters significant
  to the record grammar — double quotes, the field and entry separators, or a
  newline — the ceiling engine shall escape them per the record's existing
  quoting discipline): The record line remains a single line and stays
  machine-parseable; a debt description never injects record lines into §G or
  the audit trail, and every debt field remains recoverable after escaping.

- **REQ-ACR-006** (While a persisted ceiling outcome carries no debts, the
  record lines the persistence call writes shall keep the current grammar
  unchanged): No empty debt field is emitted for pass-through, split, or hold
  outcomes, and the required-backend refusal and operator-override record
  shapes are untouched.

- **REQ-ACR-007** (Ubiquitous): The ceiling engine shall keep outcome
  persistence best-effort — a record-write failure warns on stderr and never
  flips the admission decision the outcome already decided. The repair must
  not weaken this invariant: a debt-inventory formatting or serialization
  failure degrades to the same best-effort warning, never to an admission
  change or an engine error.

- **REQ-ACR-008** (When multiple ceiling outcomes are persisted to one SPEC's
  progress.md concurrently — parallel persistence calls racing on the same
  §G record — the ceiling engine shall append each record atomically and
  completely): Every record line written survives exactly once (no lost
  updates — the measured defect lost 17 of 24 parallel records), and
  pre-existing progress.md content — the §G heading included — is never
  destroyed, truncated, or duplicated by a racing write.

- **REQ-ACR-009** (When a convention-family evidence file carries an
  iteration suffix that does not parse as an in-range integer — an Atoi range
  error — the round counter shall count it as its own round): An unparseable
  iteration number is never collapsed into another round's number (round 1
  included); it counts via the counter's existing fail-counted path, never
  becomes the latest verdict, and never merges with a same-numbered file
  into one round. Semantics note (leader-carried, one line): round-counting
  semantics change — unparseable iteration numbers count as their own round
  rather than collapsing to 1.

## §C Success Criteria

Acceptance criteria live in `acceptance.md` (AC-ACR-001 … AC-ACR-014, Tier
M). The release-blocking RED-first criteria — AC-ACR-001 (D1), AC-ACR-004
(D1 end-to-end), AC-ACR-005 (D2), AC-ACR-013 (D3), AC-ACR-014 (D4) — each
carry the family-convention declaration **"RED is a new test (E8 evidence
required)"** (SPEC-AUDIT-CEILING-001 acceptance.md §A:11-18): the repro
tests are M1 deliverables authored against the still-pristine `903ccd028`
code, and each RED (verbatim stdout + exit code) is recorded in
`progress.md` §E.2 at M1. AC-ACR-003 is a preserve-behavior check (passes on
unmodified main and must keep passing) — deliberately not a RED observation.

## §D Non-Functional Constraints and Security

- **Scope discipline**: the defect sites across the two engine files only
  (`audit_ceiling.go` D1-D3, `audit_counter.go` D4) and their tests. No
  engine refactor, no renaming, no drive-by cleanup of adjacent code.
- **C5 posture inherited**: admission thresholds, the admission predicate
  (`internal/auditverdict`), and the outcome ladder's rung order keep their
  meaning; this SPEC changes no admission decision — only what the previous-
  round baseline resolves from, and what the record carries.
- **Security lens (card lens: --security --deep)**: evidence file names are
  untrusted input — the legacy parse stays anchored
  (`^<QuoteMeta(SpecID)>-review-([0-9]+)\.md$`, full-line anchors, digits
  only) and never interpolates the SPEC id unquoted into a pattern. Debt
  fields are verdict-supplied free text — the escaping duty of REQ-ACR-005
  applies to all three fields. Persistence stays best-effort (REQ-ACR-007) —
  a record failure is a warning, never an admission flip and never a new
  error path out of `EvaluateCeiling`.
- **One-line record grammar**: §G and trail records remain one physical line
  each; the debt field is additive and appears only when the inventory is
  non-empty.

## Out of Scope

The exclusions below bound this repair SPEC; anything named here is out of
scope for card t1560 and must ride its own card or SPEC.

### Out of Scope — engine surface beyond the two defects

- No refactor of `audit_ceiling.go` / `audit_counter.go` beyond the named
  repair sites (D1-D4) — no restructuring of the outcome ladder, no signature
  or naming changes beyond the minimum the fixes need.
- No change to `DeltaEligible`'s predicate, the diff-in-anchors check, the
  REQ/AC-set comparison, or the tier-ceiling / delta-round configuration
  semantics.
- No change to the admission predicate package (`internal/auditverdict`), the
  verdict file format, or the `Debt` struct.
- No migration, backfill, or re-parse layer for existing §G or trail lines —
  the record is append-only; historical lines stay exactly as written.

### Out of Scope — adjacent carriers and known debt

- No change to `RecordRequiredBackendRefusal` / `AcknowledgeRequiredBackend`
  record shapes or error semantics (neither carries debts).
- No change to the JSON ceiling-outcome path (`EvaluatePlanAuditCeiling` /
  `RecordCeilingOutcome`, SPEC-AUDIT-CEILING-002) — it already carries
  `DebtIDs`.
- The t1500 card-review re-review findings (including the template-neutrality
  P1 in `.moai/reports/t1500/card-review.md`) stay with their owning cards.
- t1538's remaining gate inventory (mirror-path P1, `^TestReview` overlay
  reproductions — factory dispatch-store and review-overlay territory) is
  untouched by this repair.

### Out of Scope — concurrency hardening beyond the §G append

- Other writers of progress.md (the audit reporter's append path in
  `audit_report.go`, and the agents' own §E progress-section edits) keep
  their current behavior — no lock or atomic-replace is added to them.
- Cross-process file locking (flock) stays out unless the lane finds it
  required for THIS fix — the measured defect is concurrent calls within one
  process.
- The audit-trail append (`appendAuditTrail`) already writes via a single
  O_APPEND write per line and is untouched.

## HISTORY

- v0.3.0 (2026-10-07): plan-audit-1 repair revision (verdict FAIL 0.81,
  iteration 1/2, receipt rcpt-136dc402f67a6b1d93bf69bf,
  `.moai/reports/t1560/plan-audit-1.md`) + leader ruling #2. MP-8 closed by
  the family-convention per-criterion declaration "RED is a new test (E8
  evidence required)" (acceptance.md §A rewritten; command lines added to
  the AC-ACR-004/013 RED cells; AC-ACR-012's no-RED no-regression character
  stated). AC-ACR-004 fixture corrected to tier ceiling 2 (the ceiling-1
  form could not grant a delta even after the D1 fix — `audit_ceiling.go:94/:110`
  arithmetic, codex-measured). AC-ACR-003 marked a preserve-behavior check;
  M1's RED obligation scoped to the RED-first tests. AC-ACR-010 re-anchored
  to the SPEC-local verbatim excerpt `references/t1500-seal-excerpt.md`
  (finding D4: the cross-tree path does not resolve from the card tree).
  D4 folded in per leader ruling #2 (주제당 한 장 basis, decision-index Q5):
  new REQ-ACR-009 + AC-ACR-014 (`audit_counter.go:111` Atoi-error collapse;
  M2 extended; one-line semantics note carried in the AC). Optionals folded:
  §D.2 non-REQ rows (D7), `-race -count=1` bound (D8), DoD #5 rephrased
  (D9), t1538 remote-provenance note (D10). Frontmatter version 0.3.0.
- v0.2.0 (2026-10-07): scope extension per leader ruling (card t1560,
  adopted on the 주제당 한 장 basis — decision-index Q4) — third defect D3
  added: `appendProgressRecord` (`audit_ceiling.go:479-496`, write at
  `:495`) performs an unlocked read-modify-write on progress.md; the
  turn-end codex review gate measured 24 parallel §G record calls surviving
  as 7 with the pre-existing heading lost (overlay repro on `903ccd028`).
  New REQ-ACR-008 + AC-ACR-013; folded into the persistence milestone (M3);
  broader concurrency hardening explicitly excluded.
- v0.1.0 (2026-10-07): initial plan-phase draft (card t1560, lane-2, run
  tmhxo0). Both defects verified against source at `903ccd028` in this tree
  (D1: `audit_ceiling.go:274-277` + `audit_counter.go:46,89,115-116,145-148`;
  D2: `audit_ceiling.go:461-471` + `:144` + `auditverdict/verdict.go:34-39`).
  Post-repair consistency sources read at plan phase: t1500 seal record
  (`.moai/worktrees/t1500/.moai/reports/t1500/lane28-wait-claude-gate.md`
  §SEAL/§PUSH — engine work sealed at `3d7215b72`, pushed; tip NOT an ancestor
  of origin/main, engine reached main by another landing) and t1538 sealed
  resume point (`origin/WT-t1538-factory-recovery`
  `.moai/specs/SPEC-FACTORY-COMPLETION-RECOVERY-001/progress.md` §봉인 기록 —
  remaining gates are factory mirror-path and review-overlay items, disjoint
  from `audit_ceiling.go`). Tier M (3 artifacts + progress.md); development
  mode tdd — reproduction-first.
