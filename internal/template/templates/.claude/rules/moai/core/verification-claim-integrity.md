# Verification-Claim Integrity

Doctrine establishing the **"no unobserved-verification-claim" invariant** for all MoAI actors. This rule is automatically loaded for the orchestrator and all agents. It is a policy-layer doctrine — it defines the norm; it does not itself run a runtime detector.

> The motivating defect class is general: an actor claiming a verification or completion it did not actually observe. A complementary runtime layer (advisory, warn-first, fail-open) may detect one shape of this violation; this doctrine codifies the policy norm that binds every actor regardless of whether such a runtime layer is present.

## 1. The Invariant — no unobserved-claim (verification, defect, OR premise)

[ZONE:Evolvable] [HARD] An actor MUST NOT assert a verification, a completion, **a defect / debt / drift, OR the premise underlying a recommendation** it did not actually verify with the domain's mechanical tooling.

> **Evidence absent ≠ evidence of success — NOR of failure.**

The absence of a failure signal is not, by itself, evidence that a check passed. A claim of "tests pass", "coverage met", "lint clean", or "remote in sync" is only valid when the actor actually ran the command and observed its output. An unran command, a skipped step, or a silent assumption is a gap — never a pass.

Symmetrically, inferring a defect, a technical-debt item, a drift, or an anomalous state from text patterns, grep matches, or file absence alone — without running the domain's dedicated verification tool — is not evidence that the defect exists. A text-pattern inference is a hypothesis, never a verified defect. The invariant binds both directions: an actor may not claim success it did not observe, and may not claim a defect it did not verify with the appropriate tool.

The binding extends to the premise beneath a recommendation. A recommendation to KEEP, retain, or preserve something rests on a premise — that the thing is still live, still reachable, still depended upon. Observing that an artifact is *referenced* establishes only that a reference exists; it does not establish that the referenced capability is still live. **Reachability is not justification.** Before recommending retention, the actor MUST verify the referenced capability's lifecycle status — whether its producer still exists, and whether a completed retirement already covers it. An unverified premise dressed as a reason is an unobserved claim.

This direction is the more dangerous one, because its failure is silent. A wrong "remove it" claim is contradicted by the next build or test run; a wrong "keep it" claim preserves dead code and is never contradicted by any signal at all.

This is a policy-layer norm, not a mechanical guarantee. A complementary mechanical-detection layer may surface one shape of this violation at runtime, but the norm binds every actor independently of that layer.

### 1.1 Binding scope — ALL FOUR surfaces

The invariant binds **all four** of the following surfaces. Each is named explicitly so none can claim exemption:

1. **Orchestrator self-report** — the orchestrator's own Completion Report and Verification Matrix banners, and its trust-but-verify batches, as defined in `.claude/output-styles/moai/moai.md` (Response Templates). When the orchestrator renders a Verification Matrix or Completion Report banner, every row it marks PASS MUST correspond to an actually-observed command output.

2. **Manager-agent completion report** — the self-verification deliverables of `manager-develop` and `manager-docs`. When a manager agent reports an acceptance-criteria PASS/FAIL matrix, a build result, coverage, a boundary grep, lint status, or push state, each reported result MUST be the verbatim output of a command the agent actually ran — not a summary, not an assumption, not a carry-over from a prior unrelated run.

3. **Defect / debt / drift identification claim** — any actor's assertion that a defect, technical-debt item, drift, or anomalous state EXISTS and warrants action. A claim that "module X is broken", "package Y has a coverage gap", or "N items are stale and need cleanup" is only valid when the actor ran the domain's dedicated verification tool (the project's audit / lint / type-check / coverage command) and observed its output. Inferring a defect from text patterns, grep matches, or file absence alone — without the dedicated tool — is an unobserved defect claim, and acting on it as if it were verified violates §2's attribution requirement. When a dedicated tool exists for a domain, text-only reasoning MUST NOT be the sole basis for a defect claim; the tool's output is the Evidence (§3.2).

4. **Recommendation-premise claim** — any actor's assertion of the REASON a proposed action should, or should NOT, be taken. A recommendation such as "removing this withdraws a live feature", "this is still in use", or "another consumer depends on it" is only valid when the actor verified the named premise — the producer's existence, the consumer's reachability, the owning task's lifecycle status — and observed the result. Two inferences are specifically forbidden as premise evidence: a reference existing is NOT evidence the referent is live (§1), and an originating task still reading as in-service is NOT evidence the feature it delivered survived, because a later task may have retired it. When an actor recommends AGAINST a user's stated instruction, the premise for that objection carries the same evidence burden as a defect claim (surface 3).

## 2. Baseline-Integrity Attribution / baseline 무결성 귀속

[ZONE:Evolvable] [HARD] Every verification claim MUST be attributed to an actually-measured baseline — the command that was run plus the output that was observed.

A claim MUST NOT be assumed, and MUST NOT be carried over from a prior unrelated measurement. "Coverage is at threshold" attributed to a baseline means: the actor ran the coverage command and observed the coverage figure in this run, against this tree. A number remembered from a different task, a different package, or a different point in time is NOT a baseline — it is a carry-over, and using it as if it were a fresh measurement violates this attribution requirement.

Concretely, an attributed claim names:

- **The command** — the exact invocation that produced the evidence.
- **The observed output** — the verbatim result of that invocation in this run.

Anything else (an inferred value, a stale figure, a "should be" estimate) is unattributed and MUST be reported as a Gap (§3.4), not as a Claim.

### 2.1 Moving-ref attribution — the anchor-or-subject predicate

[ZONE:Evolvable] [HARD] A claim decided against a **moving ref** — `origin/main`, `origin/develop`, `origin/HEAD`, or any other name that resolves to a different commit as work lands — carries no baseline in the sense §2 requires. The ref is an address that moves; the sentence containing it does not. What was measured against the tip on Monday is re-served as current on Friday, unchanged in text and false in fact. The same hazard rides any moving coordinate, a source line number included, so the predicate below is written for coordinates generally and merely detected on the git-ref form.

The corrective is **not** "pin every ref". Some claims are *about* the moving thing — what mainline currently carries, which tip a reader is to start from, a coordinate that is itself the subject of a correction — and pinning those destroys exactly the information they exist to record. Indiscriminate pinning is therefore the dominant failure mode of this clause, not its compliant outcome. The predicate decides, per claim, which case is at hand.

#### The four tests

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

#### Classification and remedy are two separate steps

[HARD] The tests return a **class**; the class does not name the remedy. There are **two classes and four remedies** — ANCHOR selects between R1 and R2, SUBJECT between R3 and R4. Collapsing the two steps is the second, subtler route to indiscriminate pinning: a reader who believes the class *is* the remedy has only as many remedies as there are classes, and reaches for the first one that fits.

#### The four remediation branches

| | Branch | Class | When | Form |
|---|---|---|---|---|
| **R1** | Pin the literal SHA | ANCHOR | the anchor value is already known at authoring time | replace the ref with the resolved 40-hex SHA, recorded with the tree and date it was resolved in |
| **R2** | Freeze at pre-flight *(the anchor-class default)* | ANCHOR | the value is not knowable when the criterion is written — the usual case for a run-phase PRESERVE criterion | `BASELINE_SHA=$(git rev-parse origin/main)` captured before the first run-phase commit; criteria decided against `$BASELINE_SHA`, resolved value recorded in the progress record |
| **R3** | Keep the moving ref, declare the exemption | SUBJECT / S1 | narrative — nothing is measured at read time | leave the ref; add the inline marker with a stated reason |
| **R4** | State the measuring command; demote the value to a dated reference | SUBJECT / S2 | the claim asserts the current state of a moving thing and a reader will act on it | lead with the command that must be run at read time; any value follows it, parenthesized, dated, and explicitly labelled a reference |

R2 is preferred over R1 for run-phase criteria: it removes R1's authoring-time knowledge requirement while giving the same fixed-value guarantee.

**R4's ordering is load-bearing, not stylistic.** A value written first reads as the criterion and demotes re-measurement to a confirmation step. Command first, value second and marked as a reference, so a reader who only skims still sees an instruction to measure rather than a number to trust.

**Every remedy costs the author something, and the count is what does the work.** With one remedy on offer the author pins; with four, none of them free, choosing requires applying the predicate.

| Remedy | What it costs the author |
|---|---|
| R1 | resolving the SHA and recording the tree and date it was resolved in |
| R2 | capturing the baseline before the first run-phase commit, and recording the resolved value |
| R3 | writing a non-empty reason a reviewer can disagree with |
| R4 | naming the deciding command, which a later reader will run |

R4's cost is its own definition made binding: the command it names must be the one that actually decides the claim. Left unpriced, R4 would be the cheapest available silencer — rephrasing into a shape is always cheaper than writing a justification — and that is bulk suppression reached by another road. A wrong or vague command is visible to the next reader who runs it, which is what makes the price real.

#### The exemption marker

The tests are judgments about meaning. No regex decides them, so the exemption is **author-declared**, written after applying the predicate:

```
<!-- moving-ref-ok: <reason> -->
```

- **Scope**: the flagged line, or the line immediately above it. Nothing wider — a per-claim judgment does not get document granularity.
- **Form**: an HTML comment, invisible in the rendered artifact. The marker is an author-to-linter annotation, not content for a reader of the rendered document.
- **The reason is mandatory and non-empty.** A bare marker would make "silence the warning" cheaper than "pin the SHA", inverting the incentive this clause sets. With a reason required, declaring and pinning cost about the same and the author picks on the merits.
- **An empty or whitespace-only reason does not suppress.** It produces a finding reporting the marker as *incomplete* — the one outcome that keeps the reason from becoming a formality.

A document-wide lint skip is not the exemption path: it silences a whole file, which is the wrong granularity for a per-claim judgment.

#### The five grounded instances

**Instance 1 — provenance narrative.** The claim's subject is what `origin/main` *currently carries*. Substituting a SHA converts "what mainline carries" into "what this one commit carried" — different and weaker. → **SUBJECT / S1 → R3.**

**Instance 2 — a line number as the subject of a correction** (a cadence SPEC whose citation refresh deliberately left three source coordinates unrefreshed because they were the *subject* of an audit finding rather than addresses into the tree). Substituting the current coordinate destroys the record of the miscitation. → **SUBJECT / S1 → R3.** This instance establishes the predicate's generality beyond git refs.

**Instance 3 — a dispatch's base line.** "The tip of develop you will start from", written as a SHA. Test 1 read at read-time: the substitution decays immediately, because the sentence means "whatever the tip is when you enter". Test 3: re-measuring next week gives a different tip **and that variance is the point**. → **SUBJECT / S2 → R4.** This instance is why Test 4 exists. An earlier form of the predicate classified it ANCHOR → R2, which would have had the *dispatcher* freeze a value — the very shape that failed. The defect is not that the wrong SHA was chosen; it is that a value was stated where a command belonged.

**Instance 4 — a quoted command string** (an acceptance criterion asserting that a literal command is preserved verbatim in a document). The ref token sits inside quoted subject matter and no reader measures anything on the strength of it. Test 4: no read-time action. → **SUBJECT / S1 → R3.** This is the class a shape-reading detector most often flags wrongly, and it is what the marker is for.

**Instance 5 — the same case as remedy.** Instance 3 rewritten as a standing dispatch format: *measure at entry with `git fetch origin develop`, dispatch-time reference value `<sha>`*. It is the only one of the five that shows the remedy rather than the defect.

**The classification skew, in both readings.** All five adjudicated instances are SUBJECT-class; **none is ANCHOR**. Positively, this is the strongest argument for shipping the predicate rather than the warning alone: the instances that get *noticed* and escalated are overwhelmingly the ones where pinning would have been wrong, because anchor-class defects are quietly correct to pin and generate no incident, while subject-class ones destroy information. A guard tuned only on the noticed cases is tuned entirely on the exemption class. Negatively — and the negative reading is not optional — the SUBJECT branch has five adjudicated instances and **the ANCHOR branch has zero**, resting instead on corpus lines classified by reading alone, none independently escalated or disputed. The ANCHOR branch is the unvalidated half of this predicate, and the one defect found in it by an auditor (Test 1 over-returning ANCHOR for want of a stated evaluation time) is exactly the failure an unvalidated branch would be expected to have.

#### Detection limits

A mechanism enforcing this clause reads shape, never subject. These limits are stated here rather than discovered later, and nothing in this clause implies coverage of any of them.

- **L1 — Refs expressed without an `origin/` token are invisible.** `git diff --stat main`, `git diff @{u}`, `git diff HEAD~10`, and the prose form "compared against mainline" all carry the same hazard and none of them match. Stated first because it is the limit most likely to be quietly omitted: saying it weakens the apparent value of the deliverable.
- **L2 — Shape is not subject.** No detector can apply the predicate above. Every finding is a question put to a human, never a verdict, and the finding message must say so.
- **L3 — Detection is line-scoped.** A claim whose command and whose invariant word sit on different lines — a wrapped table row, a fenced block with its assertion in prose above — is missed.
- **L4 — A documented command is indistinguishable from an asserted claim.** Doctrine that *quotes* `git reset --hard origin/main` will be flagged. This is why the exemption is a marker rather than a cleverer regex.
- **L5 — Carriers outside the artifact tree are not covered.** Dispatch messages, commit messages, PR bodies, and generated reports carry the same defect and are not scanned. Instance 3 (the occurrence that motivated R4) sits on exactly such a carrier and would not have been caught.
- **L6 — A rotted reference value is indistinguishable from a live one.** An R4-form exemption reads the *shape* of the demotion, not the freshness of the number, so a years-stale reference value passes exactly as a current one does. This limit is **created by** the R4 exemption rather than pre-existing it, and is accepted deliberately: the alternative is flagging the recommended form, which teaches readers to avoid it. The residual is bounded by R4's own ordering — the command is stated first, so a reader who follows the line re-measures regardless of what the value says. The same shape-blindness carries a second residual, on incentives: a detector cannot distinguish an author who applied the predicate and reached R4 from one who rephrased into R4 shape to stop being flagged. That price is enforced by **review, not by the detector**.
- **L7 — the ANCHOR branch of the predicate is unvalidated.** Not a limit of any detector but of the doctrine it enforces, stated among the limits because a reader applying the predicate needs it at the same moment they need the others. The SUBJECT branch has five adjudicated instances; the ANCHOR branch has zero, and the one defect found in it was found by an auditor rather than by its author. Consequences for a reader: weight an ANCHOR verdict less confidently than a SUBJECT one, and treat a Test 1 / Test 3 disagreement as evidence against ANCHOR (per the tie-break above).

#### The divergence figure, split by carrier

A divergence figure measured once — `git rev-list --count --left-right` returning `0 0` — and then re-cited without re-measurement is **the same defect on a different carrier**, not a neighbouring rule: a measurement whose validity expired, re-served as current. Its remedy is correspondingly **R4, not R1**. One does not pin `0 0`; one writes the re-measuring command as the criterion and demotes `0 0` to a dated reference.

- **A requirement for the document carrier.** A progress or plan record citing a divergence figure with no accompanying SHA or timestamp is in scope and mechanically detectable in the same place as the ref form.
- **Guidance only for the dispatch carrier.** A dispatch message is not a file in the tree; no detector sees it. It is doctrine and nothing more — stated so, because the two carriers look identical in prose and are not identical to any mechanism.

### 2.2 Tool-provenance attribution — which build judged the tree

[ZONE:Evolvable] [HARD] A measurement produced by the project's own tooling is attributed to **two** coordinates, not one: the tree it read, and the build that judged it. §2 binds the first. This clause binds the second, because a tool invoked through a shell path resolves to an *installed* build, which need not be the build the tree describes.

The failure is silent by construction, and its silence is **symmetric**. Where the judging build is behind the tree, rules that landed after that build simply do not run: the output is a clean pass, the exit status is zero, the error stream is empty. Where the judging build matches the tree, the output is *also* a clean pass, the exit status is *also* zero, the error stream is *also* empty. Nothing in either result says which case occurred — so a green result is not evidence that the checks passed, only that whatever checks the invoked build happens to carry reported nothing.

**The obligation.** A tool measurement cited as evidence MUST have been produced by a build made from the tree under measurement. Concretely, either:

- build the tool from the tree and invoke that build **by its path**, rather than letting a shell path resolve to an installed one; or
- verify — and state alongside the citation — that the installed build's commit is not a strict ancestor of the tree's HEAD.

**What the citation carries.** A cited tool measurement names the judging build's commit next to the tree's HEAD. A measurement citing only the tree is unattributed under §2: a Gap, not a Claim.

**Where it does not bind.** A build with no repository to compare against — a released artifact inside a user's project, a checkout without history — has no lag to state, and this clause requires nothing of it. A missing second coordinate is a defect only where the coordinate exists.

**Not a substitute for the tooling's own verdict.** Where the tooling already computes a freshness verdict, that verdict is the mechanism; this clause governs the **citation**, and holds whether or not the invoked build is one that reports it. A build old enough to predate the freshness check is exactly the build that cannot warn you about itself.

## 3. The 5-Section Evidence-Bearing Report Format

[ZONE:Evolvable] [HARD] Verification and completion reports — on either binding surface (§1.1) — SHOULD be structured as the following five sections. The format is the operational mechanism that enforces §1 and §2: it forces the actor to separate what is claimed from what was observed, and to make the unobserved explicit. Apply the format to every report, not only the first.

The five sections, in order:

| Section | Carries |
|---|---|
| **Claim** (주장) | what is being asserted — one discrete claim per row or sentence |
| **Evidence** (증거) | the command that was run **plus its verbatim output**; a summary is not evidence |
| **Baseline-attribution** (baseline 귀속) | what it was measured against, per §2 — command + observed output, in this run, against this tree |
| **Gaps** (미검증) | what was explicitly **NOT** observed; an empty Gaps section asserts nothing was left unobserved, which must itself be true |
| **Residual-risk** (잔여 위험) | what could still be wrong *despite* what was observed — distinct from Gaps, which is what was not observed |

What each section contains in full, the cross-reference table, and the two worked-example incident
records (the defect-claim hazard and the retention-claim hazard the §1 clauses were written from)
live in the detail companion `verification-claim-integrity-detail.md`. Load it when composing an
evidence-bearing report for the first time, or when tracing a clause back to its originating
failure.

---

Version: 1.3.0
Classification: Canonical Reference (policy-layer codification) — do not duplicate cross-referenced content; cross-reference this file instead.
