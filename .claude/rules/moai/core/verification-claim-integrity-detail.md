---
description: "Detail companion for verification-claim-integrity.md — per-section elaboration of the 5-section report format, the cross-reference table, and the two worked-example incident records"
paths: "**/verification-claim-integrity*.md"
---

# Verification-Claim Integrity — Detail Companion

> Detail companion of `verification-claim-integrity.md` (the always-loaded stub). The stub owns the
> invariant, its four binding surfaces, the baseline-attribution rule, and the five section names.
> This file owns what each section contains, the cross-reference table, and the two incident
> records the doctrine was written from. Load it when composing an evidence-bearing report for the
> first time, or when tracing a clause back to the failure that produced it.

## What each of the five sections contains

### Claim (주장)

What is being asserted. The completion or verification statement, phrased as a discrete claim — one
row per assertion in a matrix, or one sentence per claim in prose.

### Evidence (증거)

The actual command that was run plus its verbatim output, rather than a summary. If the claim is
"tests pass", the Evidence section carries the literal command (`go test ./...`) and the literal
output block it produced. Summarized evidence ("all tests passed") does not serve as Evidence — the
verbatim output is the load-bearing artifact.

### Baseline-attribution (baseline 귀속)

The baseline the claim was measured against: the command plus the observed output, in this run,
against this tree. This section answers "measured against what?", and is what stops a claim from
silently borrowing a number from an unrelated prior measurement.

### Gaps (미검증)

What was explicitly not observed — the negative space, and the key defense of the whole format.
Enumerating what was not verified is what keeps an unobserved claim from passing silently as a
success. An empty Gaps section is itself a strong assertion — that nothing was left unobserved —
and carries the same burden as any other claim. When in doubt, name the gap.

### Residual-risk (잔여 위험)

Remaining uncertainty and deferred verification: the risk surviving even the observed evidence.
Distinct from Gaps, which records what was not observed — Residual-risk records what could still be
wrong despite what was. Flaky tests, environment-specific behavior, deferred acceptance criteria,
and time-of-check-to-time-of-use windows all belong here.

## Cross-references (each remains the single source of truth for its own subject)

- `.claude/rules/moai/core/agent-common-protocol.md` § Skeptical Evaluation Stance — the
  fresh-judgment auditor stance (treat claims as suspect until evidence is shown).
- `.claude/rules/moai/core/moai-constitution.md` § Agent Core Behaviors #6 "Verify, Don't Assume" —
  the cross-cutting behavior requiring evidence of completion.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § E (Self-Verification
  Deliverables, E1-E7) — the manager-agent self-verification matrix the five-section format
  generalizes.
- `.claude/rules/moai/workflow/verification-batch-pattern.md` — the orchestrator-side read-only
  verification batching pattern, the mechanism by which observed evidence is gathered efficiently.
- `.claude/output-styles/moai/moai.md` §8 — the Verification Matrix and Completion Report banners,
  the orchestrator self-report surface the invariant binds.

## Worked example — defect-claim hazard

A status report counted 29 SPECs with `status: implemented` and an absent `era:` frontmatter field.
From frontmatter text alone, the reporter inferred "these 29 are V3R6 SPECs with a missing close"
— the legacy Mx-phase-close inference, assuming a separate close commit was required — and proposed
batch-closing all 29.

This was an unobserved defect claim. The reporter had not run the domain's dedicated verification
tool. When `moai spec audit --json` was finally run, its mechanical era classification showed all 29
were grandfather era (`V3R2-R4` 28 + `V2.x` 1) — `era_final: true`, protected, outside the V3R6
3-phase close (plan→run→sync) — and the audit's must-fix drift count across the entire catalog was
zero. The inferred close debt did not exist; had the batch-close proceeded, 29 grandfather-protected
SPECs would have been touched for no reason.

Lesson codified: **a defect claim is a hypothesis until the domain's tool confirms it.** The
`era:`-absent plus `implemented` text pattern was compatible with two contradictory readings —
grandfather legacy, or modern close-debt — and only the dedicated tool could separate them. The
obligation this produced lives in the stub, §1.1 surface 3 and §2; the tools it names in practice
are `moai spec audit` for SPEC lifecycle, `go test -cover` for coverage gaps, and `golangci-lint`
for code defects.

## Worked example — retention-claim hazard

A user instructed that `.moai/brain` be removed. The orchestrator deleted the artifacts but held one
item back — a scan in the shipped `plan/context-discovery.md` that globbed
`.moai/brain/IDEA-*/proposal.md` — on the stated premise that removing it "would withdraw a live
feature from every distributed user", and recommended a separate retirement SPEC instead.

That premise was never checked. The orchestrator had verified the scan was *reachable*
(`plan.md`'s routing table points at it) and had read that `SPEC-V3R3-BRAIN-001` still carried
`status: implemented`, then treated both facts as evidence the feature was live. Neither establishes
that. When the producers were finally enumerated, every one was already gone: the `/moai brain`
command, `workflows/brain.md`, the `manager-brain` agent, the `moai brain` CLI, the
`/moai project --from-brain` flag, the `templates/.moai/brain/` scaffold, and the docs-site pages in
all four locales. `SPEC-SUBCOMMAND-RETIRE-001` (status: completed) had retired the feature from the
template source permanently, for all distributed users, and a later cleanup commit had swept the
orphans that retirement left behind. The scan simply survived both passes. With no producer and no
scaffold, the glob could only ever return zero on a user's machine.

Lesson codified: **reachability is not justification, and a SPEC still reading `status: implemented`
is not proof the feature it delivered is still live** — a later SPEC may have retired it. The
practice this produced lives in the stub, §1.1 surface 4: before recommending retention against an
instruction, enumerate the producers of the thing being retained and check for a completed
retirement SPEC. An objection whose premise was never verified is an unobserved claim.

---

Classification: Lazy companion — rationale, elaboration, cross-references, and incident records
only. Every obligation stays in `verification-claim-integrity.md`.
