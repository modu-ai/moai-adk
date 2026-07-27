# SPEC-HARNESS-LOOP-REPAIR-001 — Acceptance Criteria

> **This file is the SSOT for acceptance criteria.** `spec.md` §E carries an index only.
> Version 0.2.0 · 2026-07-27 · manager-spec

---

## §A Standing rules for every criterion in this file

These rules bind every AC below. They exist because §A.4 of `spec.md` attributes the four-SPEC recurrence to acceptance criteria that verified component *existence* rather than *reachability*.

1. **Falsifiability is mandatory.** Every AC states a falsification condition: a specific revert that makes the AC fail. An AC whose falsification cannot be named is not admissible in this SPEC.
2. **Token presence alone is rejected.** An AC MUST NOT be satisfied by grepping a hand-authored document for a string. Where an AC checks a token, the token must appear in an artifact the change *generates* from an input, and changing the input must change the token (a round trip).
3. **Evidence is observed, not asserted.** A PASS row cites the command that was run and the output that was seen, in the run that produced it. A remembered figure is not evidence (`.claude/rules/moai/core/verification-claim-integrity.md` §2).
4. **Fixtures are self-built.** `.moai/harness/proposals/` is gitignored, so the worktree has zero drafts while the primary checkout has 52. Unit fixtures MUST construct their own `proposals/<ID>/proposal.json` under `t.TempDir()`. Any live-tree check MUST pass `--project-root` explicitly and MUST use a binary built from the tree under test.

---

## §B Status summary

| AC | M | Status | Evidence |
|---|---|---|---|
| AC-HLR-001 | M1 | **PASS** | §C.1 |
| AC-HLR-002 | M1 | **PASS** | §C.2 |
| AC-HLR-003 | M1 | **PASS** | §C.3 |
| AC-HLR-006 | M1 | **PASS** | §C.4 |
| AC-HLR-004 | M2 | open | §D.1 |
| AC-HLR-005 | M2 | open | §D.2 |
| AC-HLR-014 | M2 | open | §D.3 |
| AC-HLR-015 | M2 | open | §D.4 |
| AC-HLR-016 | M2 | open | §D.5 |
| AC-HLR-007 | M3 | **PASS** | §E.1 |
| AC-HLR-008 | M4 | open | §F.1 |
| AC-HLR-017 | M4 | open | §F.2 |
| AC-HLR-009 | M5 | open | §G.1 |
| AC-HLR-010 | M5 | open | §G.2 |
| AC-HLR-011 | M6 | open | §H.1 |
| AC-HLR-012 | M6 | open | §H.2 |
| AC-HLR-013 | M6 | open | §H.3 |

4 of 17 PASS. All four PASS rows belong to M1 and were verified at commit `c996eb294`.

---

## §C M1 — layout contract (COMPLETE)

M1 repaired the producer/consumer layout mismatch (`spec.md` §A.3.1) by introducing a shared accessor at `internal/harness/proposalgen/layout.go` and routing all three consumers through it.

### §C.1 AC-HLR-001 — pending count matches disk · **PASS**

- **Given** N draft proposal directories under `.moai/harness/proposals/`, each carrying `proposal.json`
- **When** `moai harness status` runs against that project root
- **Then** it reports `pending proposals: N items`, with N > 0
- **Falsification** — inverting the directory predicate in `ListDraftIDs` restores `0 items` against the same fixture

**Evidence.** `moai harness status --project-root <primary checkout>`, using a binary built from this worktree, reported `pending proposals: 52 items`. Baseline before the change: `0 items`. The falsification was executed, not assumed: inverting the predicate made `TestCountProposals_NestedLayout` report `countProposals = 0, want 2`; the edit was reverted and the suite re-confirmed green.

### §C.2 AC-HLR-002 — apply returns a payload · **PASS**

- **Given** at least one draft proposal on disk
- **When** `moai harness apply` runs
- **Then** stdout carries that draft's payload, not `No pending proposals.`
- **Falsification** — reverting the accessor restores the `No pending proposals.` branch

**Evidence.** `moai harness apply --project-root <primary checkout>` emitted the `PROPOSAL-20260617-dc05149f` payload. Falsification verified via `TestHarnessApply_NestedLayout`.

### §C.3 AC-HLR-003 — execute resolves by ID · **PASS**

- **Given** a draft proposal with identifier `<ID>`
- **When** `moai harness execute --id <ID>` runs
- **Then** it does not fail with `proposal not found`
- **Falsification** — restoring `draftID + ".json"` in `ProposalPath` reintroduces `proposal not found`

**Evidence.** `moai harness execute --id PROPOSAL-20260617-dc05149f` against an isolated copy reached the file and failed at the schema layer instead (exit 1). Falsification verified via `TestResolveProposalPath_NestedLayout` and `TestLoadProposalByID_NestedLayout`.

> **Scope note.** This AC asserts *path resolution only*. The subsequent schema failure is `spec.md` §A.3.2 / §A.7 and is addressed by M2 — it is not a defect in M1's deliverable.

### §C.4 AC-HLR-006 — single accessor · **PASS**

- **Given** the repaired code
- **Then** C1 (`countProposals`), C2 (the apply selector) and C3 (`resolveProposalPath`) resolve proposals through one shared function, and no call site re-derives `id + ".json"`
- **Falsification** — a regression test fails if any site reintroduces an independent path derivation

**Evidence.** `go test -run TestNoFlatProposalPathDerivation ./internal/cli/` passed; the guard scans call sites for independent `id + ".json"` derivation.

**Full-suite context at `c996eb294`.** `go test ./...` exit 0 (105 packages), `go vet ./...` exit 0, `golangci-lint run` exit 0, `GOOS=windows` and `GOOS=linux` builds exit 0.

---

## §D M2 — route drafts to their designed consumer

M2 acts on `spec.md` §A.3.2 / §A.4: a `proposalgen` draft is a discovery report whose consumer is manager-spec SPEC authoring, not `Applier.Apply()`.

### §D.1 AC-HLR-004 — a draft reaches its designed consumer

> Rewritten in v0.2.0. Quoted verbatim below because it replaces the SPEC's former central success signal.

- **Given** a draft proposal directory `<DRAFT-ID>` under `.moai/harness/proposals/`, carrying both `spec.md` and `proposal.json`
- **When** the promotion path for `<DRAFT-ID>` runs
- **Then** all three of the following hold:
  1. a SPEC directory exists under `.moai/specs/` that did not exist before the run;
  2. that SPEC's `spec.md` records `<DRAFT-ID>` as its provenance, and the recorded value equals the input draft ID exactly — promoting a *different* draft records a *different* value;
  3. `moai harness status` no longer counts `<DRAFT-ID>` among pending drafts, and the count of the remaining drafts is unchanged.
- **Falsification** — three independent reverts, each of which fails this AC on its own:
  1. remove the promotion path → no SPEC directory is created;
  2. hard-code the provenance value instead of deriving it from the input → the AC fails when run twice with two different draft IDs;
  3. leave the draft in the pending queue after promotion → the pending count is unchanged.
- **Explicitly NOT required** — an `apply_outcome` record. The former formulation (`grep -c apply_outcome … ≥ 1`) assumed drafts feed `Applier.Apply()`; `spec.md` §A.4 establishes they cannot. `apply_outcome` remains the success signal of the Applier path, which this SPEC leaves unfed.

**Traces to** REQ-HLR-004.

**Why clause 2 is a round trip, not a token check.** The provenance value is derived from the command's input and compared against it; running the AC with a second draft ID must produce a second, different value. A hard-coded string passes a single-run grep but fails the two-run comparison. This satisfies §A rule 2.

### §D.2 AC-HLR-005 — promotion is auditable

> Rewritten in v0.2.0. The former criterion required `.moai/harness/learning-history/applied/` to materialise; that directory belongs to the Applier path, which this SPEC does not feed.

- **Given** one completed promotion of `<DRAFT-ID>` to `<SPEC-ID>`
- **Then** exactly one durable record exists linking `<DRAFT-ID>` → `<SPEC-ID>` with a timestamp, and a second promotion of a different draft appends exactly one further record (not zero, not two)
- **Falsification** — removing the record write leaves the count unchanged across a promotion
- **Explicitly NOT required** — `.moai/harness/learning-history/applied/`. Its absence is the correct state for this SPEC and MUST NOT be reported as a defect.

**Traces to** REQ-HLR-004.

### §D.3 AC-HLR-014 — `execute` no longer accepts a proposalgen draft, and says why

- **Given** a draft proposal directory `<DRAFT-ID>` written by `proposalgen`
- **When** the apply/execute path is invoked against `<DRAFT-ID>`
- **Then** it fails with a diagnostic that names the reason — that a discovery draft carries no `target_path` / `field_key` / `new_value` and is not an apply input — and it does NOT fail with a raw JSON unmarshal error
- **And** no snapshot directory is created by the attempt
- **Falsification** — reverting the de-wiring restores the raw `parse proposal … cannot unmarshal string into Go struct field` diagnostic

**Traces to** REQ-HLR-004b.

**Baseline.** Today the verb fails at `internal/cli/harness/execute.go:223` with a raw unmarshal error, because `proposal.json` carries `"tier": "auto_update"` (string) while `harness.Proposal.Tier` is numeric. That message describes a symptom, not the reason.

### §D.4 AC-HLR-015 — applicability guard precedes snapshot

- **Given** a proposal that decodes successfully but carries an empty `target_path`, `field_key`, or `new_value`
- **When** it is submitted to the apply path
- **Then** it is rejected with a diagnostic naming the missing field, **and** the snapshot base directory contains no new entry as a result of the attempt
- **Falsification** — removing the guard lets the proposal reach `createSnapshot`, which creates a dated directory (`internal/harness/applier.go:648`) before failing on `os.ReadFile("")` (line 653); the new directory is then observable

**Traces to** REQ-HLR-004c.

**Why this AC is load-bearing.** `spec.md` §A.7 verifies that all five safety layers pass a content-free proposal: L1 explicitly returns false for an empty path (`frozen_guard.go:42-44`), L2 scores it as an unchanged baseline (`canary.go:56-59`), L3 finds no frozen rule (`frozen_rules.go:66-68`), L4 is content-independent, and L5 is auto-approved (`execute.go` `AutoApply=true`). The pipeline checks whether an edit is *safe*, never whether an edit was *specified*. The `unsupported fieldKey` branch at `applier.go:476` cannot serve as this guard — `createSnapshot` fails first, so that branch is unreachable for a content-free proposal.

### §D.5 AC-HLR-016 — tier-split disposition recorded, tripwire retired with rationale

- **Given** the de-wiring in AC-HLR-014 has landed
- **Then** all of:
  1. no `MarshalJSON` / `UnmarshalJSON` method is added to `harness.Tier`;
  2. `TestLoadProposalByID_ProducerSchemaMismatch` (`internal/cli/harness/layout_repro_test.go:141`) is retired, with its removal accompanied by a recorded rationale stating the mismatch was dissolved by de-wiring rather than repaired;
  3. `go test ./...` is green after the retirement.
- **Falsification** — deleting the test without the recorded rationale, or adding a `Tier` JSON codec, each fails this AC

**Traces to** REQ-HLR-012.

**Why no codec is needed.** The `tier` string/numeric split has exactly one consumer: `loadProposalByID` in the execute path (`internal/cli/harness/execute.go:222-225`). The other two consumers never decode into `harness.Proposal` — `countProposals` only counts directories (`internal/cli/harness.go:194-200`), and the default `apply` path reads raw bytes and echoes them to stdout (`internal/cli/harness.go:278-285`). Removing the execute→draft wiring removes the only consumer, so the split is dissolved rather than fixed. Adding a codec would be work with no caller.

> **Conditional fallback.** Should M2 design instead retain a typed reader for `proposal.json`, the minimal repair is a `Tier.UnmarshalJSON` accepting both the string vocabulary (the inverse of the existing `String()` SSOT at `internal/harness/types.go:245`) and a bare number. That is compile-neutral: a method addition changes no existing struct literal, so the 57 `Tier:` literal sites across `internal/**/*_test.go` are unaffected. This fallback is recorded so the option is not re-derived; it is NOT the selected path.

---

## §E M3 — dispatch observation

### §E.1 AC-HLR-007 — dispatch recorded

- **Given** the routing observation opt-in is enabled
- **When** a `/moai` subcommand dispatches
- **Then** the routing-ledger line count increases by exactly one
- **Falsification** — removing the record call leaves the count unchanged across a dispatch

**Traces to** REQ-HLR-005.

**Baseline (corrected M3).** `.moai/state/routing-ledger.jsonl` was absent until the audit wrote one row by hand. `moai harness ledger record` exits 0 and creates a **pending** row (`.moai/state/routing-pending-<session>.json`) — it does not append to the ledger directly; the ledger line appears only at terminal-evidence finalization on Stop. The original "one row written" phrasing conflated the pending row with a finalized ledger line; the ledger was empty because no dispatch was ever driven to a terminal-evidence Stop, not because the writer was broken.

**Evidence (M3, executed 2026-07-28, worktree binary, isolated `/tmp` root).** Both opt-ins ON (`hook.opt_in.enabled: true` in `system.yaml`; `learning.enabled` default-ON via absent `harness.yaml`):

| Step | Command | Observed |
|---|---|---|
| dispatch | `echo "/moai run SPEC-X" \| moai harness ledger record --subcommand run --session m3a --tier M` | 1 pending row created; ledger still 0 |
| terminal evidence | `moai harness ledger evidence --session m3a --kind gate_exit --value 0 --terminal --ref "go test ./..."` | terminal `gate_exit` appended to the pending row |
| finalize | `moai hook harness-observe-stop` (stdin `{"session_id":"m3a",...}`) | ledger 0→**1**; pending deleted; row `outcome:"success"` derived from the terminal evidence |

The finalized ledger row carries `request_digest: sha256:e0e7dba9545a` and **not** the verbatim request — `grep -c 'SPEC-X-DISPATCH-123' routing-ledger.jsonl` = 0 (privacy preserved).

**Falsification executed.** With the `record` call removed (only `harness-observe-stop` run for a session `m3b` with no pending row), the ledger count stayed at 1 — `FinalizeOnStop` is a self-gated no-op when no pending row exists (`pending.go:148-150`). Removing the record call leaves the count unchanged across a dispatch, as required. The opt-in gate was also confirmed: with `hook.opt_in.enabled: false`, `record`+`evidence`+`Stop` left the ledger unchanged (`finalizeRoutingLedgerOnStop` gate 0, `hook.go:737-738`).

**Lifecycle clarification (operational reading of "routing-ledger line count +1").** The dispatch-time `record` writes a *pending* row; the *ledger* line appears only at terminal-evidence finalization on Stop (`internal/harness/routing/pending.go` `FinalizeOnStop` → `DeriveOutcome`). The AC's "+1 routing-ledger line per dispatch" is realized across the dispatch→finalize lifecycle, not atomically at the `record` call. This is the designed pending→finalize architecture, not a defect.

**Residual risk (not closed by M3).** The recording obligation lives in doctrine (`.claude/skills/moai/SKILL.md` router section; `.claude/skills/moai/workflows/run.md:196`) and is LLM-obeyed — no mechanical backstop guarantees the orchestrator calls `record` at dispatch. The doctrine shipped 2026-07-12 (`SPEC-HARNESS-EVOLVE-001` M3, commit `1c54cd9c6`); the primary-checkout ledger nonetheless holds only 1 row (audit-hand-written). M3 verifies the mechanics are correct and the falsification passes; closing the obedience gap is a candidate follow-up (a mechanical dispatch-time trigger), not part of this milestone.

---

## §F M4 — generator quality

### §F.1 AC-HLR-008 — promotion excludes bare tool names

- **Given** the narrowed promotion rule
- **When** the generator runs against the existing usage log
- **Then** no newly generated draft carries a `pattern_key` whose subject is a bare tool name
- **Falsification** — reverting the narrowing reproduces an `agent_invocation:<Tool>` draft from the same log

**Traces to** REQ-HLR-009.

**Baseline.** 29 of 52 drafts (56%) are `agent_invocation` records naming a bare tool; 28 are unique (`spec.md` §A.5).

### §F.2 AC-HLR-017 — one pattern yields one draft across dates

- **Given** a fixture with two `Promotion` records sharing one `pattern_key` and carrying timestamps on two different dates
- **When** the generator runs
- **Then** exactly one draft directory exists for that `pattern_key`
- **Falsification** — reverting the identity change reproduces two directories differing only in their date prefix

**Traces to** REQ-HLR-011.

**Baseline.** 52 drafts collapse to 45 unique `pattern_key` values; the 7-draft difference is 7 duplicate pairs, enumerated in `spec.md` §A.6. `buildDraftID` (`internal/harness/proposalgen/mapper.go:130-135`) takes the date from the promotion timestamp, so the same pattern promoted on another day yields a new ID.

**Scope boundary.** Forward-looking only. The 7 existing duplicate pairs are grandfathered per `spec.md` §B.2; this AC MUST NOT be read as requiring their removal.

---

## §G M5 — lesson channel

### §G.1 AC-HLR-009 — one designated lesson store

- **Given** the reconciled constitution
- **Then** the Lessons Protocol names the practiced store, and no rule names a store that has not been written to in 30 days
- **Falsification** — reverting the reconciliation restores a designated store whose mtime is older than the practiced one

**Traces to** REQ-HLR-006.

**Baseline.** `lessons.md` (the constitution-designated store) was last modified 2026-06-17 — 40 days stale — while live traffic goes to 102 `feedback_*.md` topic files.

### §G.2 AC-HLR-010 — inbox drain named

- **Given** the reconciled doctrine
- **Then** the drain actor and the drain trigger are both named, and a drain run reduces the undrained entry count
- **Falsification** — a drain run that leaves the count unchanged fails this AC

**Traces to** REQ-HLR-007.

**Baseline.** `.moai/lessons-inbox.jsonl` held 845 lines, appended to on the day of the audit, never drained.

---

## §H M6 — falsifiability and CLI reporting

### §H.1 AC-HLR-011 — falsifiability recorded

- **Given** a harness edit made under this SPEC
- **Then** its lesson entry carries `prediction:` at edit time and `verified: true|false` after observation
- **Falsification** — an entry carrying neither field fails this AC

**Traces to** REQ-HLR-008.

**Baseline.** Zero occurrences of `prediction:` or `verified:` across 102 feedback files plus `lessons.md`, despite the constitution requiring both (§ Lessons Protocol, Harness Edit Discipline). This SPEC is itself the first test of the obligation.

### §H.2 AC-HLR-012 — help enumerates every verb

- **Given** `moai harness --help`
- **Then** its description lists every verb present in the command table
- **Falsification** — removing a verb from the description while it remains registered fails this AC

**Traces to** REQ-HLR-010.

**Baseline.** 6 verbs omitted from the help text: `clusters`, `propose`, `install`, `execute`, `doctor`, `ledger`.

### §H.3 AC-HLR-013 — list and doctor agree

- **Given** a command-only thin harness
- **Then** `list` does not describe it in defect-suggesting terms while `doctor` classifies the same state as expected
- **Falsification** — reverting restores the disagreement on the same fixture

**Traces to** REQ-HLR-010.

---

## §I Traceability matrix

| REQ | Covered by |
|---|---|
| REQ-HLR-001 | AC-HLR-006 |
| REQ-HLR-002 | AC-HLR-001 |
| REQ-HLR-003 | AC-HLR-002 |
| REQ-HLR-004 | AC-HLR-004, AC-HLR-005 |
| REQ-HLR-004b | AC-HLR-014 |
| REQ-HLR-004c | AC-HLR-015 |
| REQ-HLR-005 | AC-HLR-007 |
| REQ-HLR-006 | AC-HLR-009 |
| REQ-HLR-007 | AC-HLR-010 |
| REQ-HLR-008 | AC-HLR-011 |
| REQ-HLR-009 | AC-HLR-008 |
| REQ-HLR-010 | AC-HLR-012, AC-HLR-013 |
| REQ-HLR-011 | AC-HLR-017 |
| REQ-HLR-012 | AC-HLR-016 |

Every REQ has at least one AC. AC-HLR-003's coverage of REQ-HLR-003 is partial by design: REQ-HLR-003 concerns apply reachability, and the payload's *applicability* is governed separately by REQ-HLR-004b/004c.

---

## §J Definition of Done

The SPEC is done when all of:

1. Every AC in §B is PASS, or is explicitly deferred with a recorded rationale and a successor SPEC ID.
2. `go test ./...`, `go vet ./...`, and `golangci-lint run` all exit 0.
3. `GOOS=windows` and `GOOS=linux` builds exit 0.
4. No `AskUserQuestion` / `mcp__askuser` invocation exists in `internal/harness/` or `internal/cli/harness/` (prose comments and flag descriptions are permitted).
5. Every falsification named in this file has been **executed** at least once — the revert applied, the failure observed, the revert undone, the suite re-confirmed green. An unexecuted falsification is a §A rule 3 gap, not a PASS.
6. `spec.md` §G open questions 2, 3, 4 and 5 are each either resolved or explicitly carried forward with an owner.
