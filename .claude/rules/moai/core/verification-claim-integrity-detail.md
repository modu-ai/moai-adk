---
description: "Detail companion for verification-claim-integrity.md — per-section elaboration of the 5-section report format, the cross-reference table, the two worked-example incident records, and the §2.1 moving-ref predicate procedure (four tests, grounded instances, detection limits)"
paths: "**/verification-claim-integrity*.md"
---

# Verification-Claim Integrity — Detail Companion

> Detail companion of `verification-claim-integrity.md` (the always-loaded stub). The stub owns the
> invariant, its four binding surfaces, the baseline-attribution rule, and the five section names.
> This file owns what each section contains, the cross-reference table, the two incident
> records the doctrine was written from, and the §2.1 moving-ref predicate procedure. Load it when
> composing an evidence-bearing report for the first time, when applying the moving-ref predicate,
> or when tracing a clause back to the failure that produced it.

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

## Moving-ref predicate — the four tests, the grounded instances, and the detection limits

> Relocated from the stub's §2.1 by card t492 (always-loaded surface diet). The stub retains the
> [HARD] clause itself, the classification-is-not-remedy rule, the four remediation branches
> (R1-R4) with their cost table, and the exemption-marker syntax. What lives here is the procedure
> for reaching a class, the adjudicated instances it was derived from, and what a mechanism
> enforcing it cannot see. Read this section before applying the predicate for the first time.

### The four tests

Applied in order. Every test is answerable by reading the sentence the ref appears in.

**Test 1 — Substitution.** Replace the ref token with the SHA it resolves to *right now*, then re-read the sentence **as a later reader will act on it — not as you read it at the moment of substitution.** The evaluation time is load-bearing: a Test 1 applied at the substitution instant returns ANCHOR for every live-state claim there is, because substituting today's tip into "the base you will start from" reads correctly today and is wrong for every reader after.

- Still says what it meant, *and still will when acted on later* → **ANCHOR** (an address at which a measurement was taken).
- Now says something different, narrower, or weaker — **including a meaning correct at the instant of substitution that decays afterwards** → **SUBJECT** (the claim is *about* the moving thing).

**Test 2 — Falsification source.** Conditional: it runs only when the claim currently reads false, and it returns an attribution rather than a class. Were the commits that flipped it authored by this work?

- Authored by this work → **true signal**. The claim is genuinely broken; fix the work, do not touch the ref. Pinning here hides a real defect, which is the worse of the two errors.
- Not authored by this work → **spurious red** from upstream drift; remediate per the branches below.

**Test 3 — Re-measurement expectation.** Re-run this claim next week with no work done in between. Is the same answer expected? Yes → ANCHOR. No, *and that variance is the point of the claim* → SUBJECT.

**Tie-break.** Tests 1 and 3 normally agree. Where they disagree, **run Test 4 — do not resolve to ANCHOR.** A disagreement is not noise to be settled by precedence; it is the signature of a live-state claim, in which Test 1 says "the value fits" while Test 3 says "the value must not be fixed". Treat a Test 1 / Test 3 disagreement as evidence *against* ANCHOR rather than for it.

**Test 4 — Read-time action.** Runs when Tests 1 and 3 both return SUBJECT, and whenever they disagree. Test 2 is deliberately outside this gate — it is conditional and returns no class, so a gate naming "Tests 1-3" would be unsatisfiable for any claim that reads true, which is most of them. Ask: must a later reader *act* on this claim by measuring something?

- **No — the claim is narrative.** It describes what mainline carries, quotes a command as text, or records a coordinate as the subject of a correction. Nothing is measured at read time. → **S1**.
- **Yes — the claim asserts the current state of a moving thing and a reader will act on it.** → **S2**.


### The five grounded instances

**Instance 1 — provenance narrative.** The claim's subject is what `origin/main` *currently carries*. Substituting a SHA converts "what mainline carries" into "what this one commit carried" — different and weaker. → **SUBJECT / S1 → R3.**

**Instance 2 — a line number as the subject of a correction** (`SPEC-GRAPH-FRESHNESS-CADENCE-001`, whose citation refresh deliberately left three source coordinates unrefreshed because they were the *subject* of an audit finding rather than addresses into the tree). Substituting the current coordinate destroys the record of the miscitation. → **SUBJECT / S1 → R3.** This instance establishes the predicate's generality beyond git refs.

**Instance 3 — a dispatch's base line.** "The tip of develop you will start from", written as a SHA. Test 1 read at read-time: the substitution decays immediately, because the sentence means "whatever the tip is when you enter". Test 3: re-measuring next week gives a different tip **and that variance is the point**. → **SUBJECT / S2 → R4.** This instance is why Test 4 exists. An earlier form of the predicate classified it ANCHOR → R2, which would have had the *dispatcher* freeze a value — the very shape that failed. The defect is not that the wrong SHA was chosen; it is that a value was stated where a command belonged.

**Instance 4 — a quoted command string** (`AC-COORD-016`, which asserts a literal command is preserved verbatim in a document). The ref token sits inside quoted subject matter and no reader measures anything on the strength of it. Test 4: no read-time action. → **SUBJECT / S1 → R3.** This is the class a shape-reading detector most often flags wrongly, and it is what the marker is for.

**Instance 5 — the same case as remedy.** Instance 3 rewritten as a standing dispatch format: *measure at entry with `git fetch origin develop`, dispatch-time reference value `<sha>`*. It is the only one of the five that shows the remedy rather than the defect.

**The classification skew, in both readings.** All five adjudicated instances are SUBJECT-class; **none is ANCHOR**. Positively, this is the strongest argument for shipping the predicate rather than the warning alone: the instances that get *noticed* and escalated are overwhelmingly the ones where pinning would have been wrong, because anchor-class defects are quietly correct to pin and generate no incident, while subject-class ones destroy information. A guard tuned only on the noticed cases is tuned entirely on the exemption class. Negatively — and the negative reading is not optional — the SUBJECT branch has five adjudicated instances and **the ANCHOR branch has zero**, resting instead on corpus lines classified by reading alone, none independently escalated or disputed. The ANCHOR branch is the unvalidated half of this predicate, and the one defect found in it by an auditor (Test 1 over-returning ANCHOR for want of a stated evaluation time) is exactly the failure an unvalidated branch would be expected to have.

### Detection limits

A mechanism enforcing this clause reads shape, never subject. These limits are stated here rather than discovered later, and nothing in this clause implies coverage of any of them.

- **L1 — Refs expressed without an `origin/` token are invisible.** `git diff --stat main`, `git diff @{u}`, `git diff HEAD~10`, and the prose form "compared against mainline" all carry the same hazard and none of them match. Stated first because it is the limit most likely to be quietly omitted: saying it weakens the apparent value of the deliverable.
- **L2 — Shape is not subject.** No detector can apply the predicate above. Every finding is a question put to a human, never a verdict, and the finding message must say so.
- **L3 — Detection is line-scoped.** A claim whose command and whose invariant word sit on different lines — a wrapped table row, a fenced block with its assertion in prose above — is missed.
- **L4 — A documented command is indistinguishable from an asserted claim.** Doctrine that *quotes* `git reset --hard origin/main` will be flagged. This is why the exemption is a marker rather than a cleverer regex.
- **L5 — Carriers outside the artifact tree are not covered.** Dispatch messages, commit messages, PR bodies, and generated reports carry the same defect and are not scanned. Instance 3 (the occurrence that motivated R4) sits on exactly such a carrier and would not have been caught.
- **L6 — A rotted reference value is indistinguishable from a live one.** An R4-form exemption reads the *shape* of the demotion, not the freshness of the number, so a years-stale reference value passes exactly as a current one does. This limit is **created by** the R4 exemption rather than pre-existing it, and is accepted deliberately: the alternative is flagging the recommended form, which teaches readers to avoid it. The residual is bounded by R4's own ordering — the command is stated first, so a reader who follows the line re-measures regardless of what the value says. The same shape-blindness carries a second residual, on incentives: a detector cannot distinguish an author who applied the predicate and reached R4 from one who rephrased into R4 shape to stop being flagged. That price is enforced by **review, not by the detector**.
- **L7 — the ANCHOR branch of the predicate is unvalidated.** Not a limit of any detector but of the doctrine it enforces, stated among the limits because a reader applying the predicate needs it at the same moment they need the others. The SUBJECT branch has five adjudicated instances; the ANCHOR branch has zero, and the one defect found in it was found by an auditor rather than by its author. Consequences for a reader: weight an ANCHOR verdict less confidently than a SUBJECT one, and treat a Test 1 / Test 3 disagreement as evidence against ANCHOR (per the tie-break above).

### The divergence figure, split by carrier

A divergence figure measured once — `git rev-list --count --left-right` returning `0 0` — and then re-cited without re-measurement is **the same defect on a different carrier**, not a neighbouring rule: a measurement whose validity expired, re-served as current. Its remedy is correspondingly **R4, not R1**. One does not pin `0 0`; one writes the re-measuring command as the criterion and demotes `0 0` to a dated reference.

- **A requirement for the document carrier.** A progress or plan record citing a divergence figure with no accompanying SHA or timestamp is in scope and mechanically detectable in the same place as the ref form.
- **Guidance only for the dispatch carrier.** A dispatch message is not a file in the tree; no detector sees it. It is doctrine and nothing more — stated so, because the two carriers look identical in prose and are not identical to any mechanism.


---

Classification: Lazy companion — rationale, elaboration, cross-references, incident records, and
the §2.1 predicate procedure. Every obligation stays in `verification-claim-integrity.md`: the
predicate procedure here is HOW a class is reached, never WHETHER it must be applied.
