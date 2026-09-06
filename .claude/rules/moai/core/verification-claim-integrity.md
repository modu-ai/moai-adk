# Verification-Claim Integrity

Doctrine establishing the **"no unobserved-verification-claim" invariant** for all MoAI actors. This rule is automatically loaded for the orchestrator and all agents. It is a policy-layer (codification) doctrine — it defines the norm; it does not itself run a runtime detector.

> Provenance: SPEC-EVIDENCE-CLAIM-INVARIANT-001. §2.2 from SPEC-BINLAG-INVOCATION-001.

## 1. The Invariant — no unobserved-claim (verification, defect, OR premise)

[ZONE:Evolvable] [HARD] An actor MUST NOT assert a verification, a completion, **a defect / debt / drift, OR the premise underlying a recommendation** it did not actually verify with the domain's mechanical tooling.

> **Evidence absent ≠ evidence of success — NOR of failure.**

The absence of a failure signal is not, by itself, evidence that a check passed. A claim of "tests pass", "coverage 87%", "lint clean", or "0 0 sync" is only valid when the actor actually ran the command and observed its output. An unran command, a skipped step, or a silent assumption is a gap — never a pass.

Symmetrically, inferring a defect, a technical-debt item, a lifecycle drift, or an anomalous state from frontmatter text, grep matches, or file absence alone — without running the domain's dedicated verification tool — is not evidence that the defect exists. A text-pattern inference is a hypothesis, never a verified defect. The invariant binds both directions: an actor may not claim success it did not observe, and may not claim a defect it did not verify with the appropriate tool.

The binding extends to the premise beneath a recommendation. A recommendation to KEEP, retain, or preserve something rests on a premise — that the thing is still live, still reachable, still depended upon. Observing that an artifact is *referenced* establishes only that a reference exists; it does not establish that the referenced capability is still live. **Reachability is not justification.** Before recommending retention, the actor MUST verify the referenced capability's lifecycle status — whether its producer still exists, and whether a completed retirement already covers it. An unverified premise dressed as a reason is an unobserved claim.

This direction is the more dangerous one, because its failure is silent. A wrong "remove it" claim is contradicted by the next build or test run; a wrong "keep it" claim preserves dead code and is never contradicted by any signal at all.

This is a policy-layer norm, not a mechanical guarantee. For the complementary mechanical-detection layer that surfaces one shape of this violation at runtime, cross-reference SPEC-STOP-EVIDENCE-GATE-001 (the cross-reference table lives in the detail companion, `verification-claim-integrity-detail.md`).

### 1.1 Binding scope — ALL FOUR surfaces

The invariant binds **all four** of the following surfaces. Each is named explicitly so none can claim exemption:

1. **Orchestrator self-report** — the orchestrator's own Completion Report and Verification Matrix banners, and its Trust-but-verify batches, as defined in `.claude/output-styles/moai/moai.md` §8 (Response Templates). When the orchestrator renders a Verification Matrix or Completion Report banner, every row it marks PASS MUST correspond to an actually-observed command output.

2. **Manager-agent completion report** — the `§E` self-verification (E1-E7) of `manager-develop` and `manager-docs`. When a manager agent reports an AC PASS/FAIL matrix (E1), cross-platform build result (E2), coverage (E3), subagent-boundary grep (E4), lint status (E5), or push state (E6), each reported result MUST be the verbatim output of a command the agent actually ran — not a summary, not an assumption, not a carry-over from a prior unrelated run.

3. **Defect / debt / drift identification claim** — any actor's assertion that a defect, technical-debt item, lifecycle drift, or anomalous state EXISTS and warrants action. A claim that "SPEC X is a close debt", "package Y has a coverage gap", or "N SPECs need Mx-close" is only valid when the actor ran the domain's dedicated verification tool (`moai spec audit`, `go test -cover`, `golangci-lint`, etc.) and observed its output. Inferring a defect from frontmatter text, grep matches, or file absence alone — without the dedicated tool — is an unobserved defect claim, and acting on it as if it were verified violates §2's attribution requirement. When a dedicated tool exists for a domain, text-only reasoning MUST NOT be the sole basis for a defect claim; the tool's output is the Evidence (§3.2).

4. **Recommendation-premise claim** — any actor's assertion of the REASON a proposed action should, or should NOT, be taken. A recommendation such as "removing this withdraws a live feature", "this is still in use", or "another consumer depends on it" is only valid when the actor verified the named premise — the producer's existence, the consumer's reachability, the owning task's lifecycle status — and observed the result. Two inferences are specifically forbidden as premise evidence: a reference existing is NOT evidence the referent is live (§1), and an originating task still reading as in-service is NOT evidence the feature it delivered survived, because a later task may have retired it. When an actor recommends AGAINST a user's stated instruction, the premise for that objection carries the same evidence burden as a defect claim (surface 3).

## 2. Baseline-Integrity Attribution / baseline 무결성 귀속

[ZONE:Evolvable] [HARD] Every verification claim MUST be attributed to an actually-measured baseline — the command that was run plus the output that was observed.

A claim MUST NOT be assumed, and MUST NOT be carried over from a prior unrelated measurement. "Coverage is 87%" attributed to a baseline means: the actor ran `go test -cover ./internal/<pkg>/...` (the command) and observed `coverage: 87.0% of statements` (the output) in this run, against this tree. A number remembered from a different SPEC, a different package, or a different point in time is NOT a baseline — it is a carry-over, and using it as if it were a fresh measurement violates this attribution requirement.

Concretely, an attributed claim names:

- **The command** — the exact invocation that produced the evidence.
- **The observed output** — the verbatim result of that invocation in this run.

Anything else (an inferred value, a stale figure, a "should be" estimate) is unattributed and MUST be reported as a Gap (§3.4), not as a Claim.

### 2.1 Moving-ref attribution — the anchor-or-subject predicate

[ZONE:Evolvable] [HARD] A claim decided against a **moving ref** — `origin/main`, `origin/develop`, `origin/HEAD`, or any other name that resolves to a different commit as work lands — carries no baseline in the sense §2 requires. The ref is an address that moves; the sentence containing it does not. What was measured against the tip on Monday is re-served as current on Friday, unchanged in text and false in fact. The same hazard rides any moving coordinate, a source line number included, so the predicate is written for coordinates generally and merely detected on the git-ref form.

The corrective is **not** "pin every ref". Some claims are *about* the moving thing — what mainline currently carries, which tip a reader is to start from, a coordinate that is itself the subject of a correction — and pinning those destroys exactly the information they exist to record. Indiscriminate pinning is therefore the dominant failure mode of this clause, not its compliant outcome. The predicate decides, per claim, which case is at hand.

[HARD] The predicate is applied, not recalled. Before remediating any moving-ref or moving-coordinate claim, read `verification-claim-integrity-detail.md` § Moving-ref predicate and run its four tests in order; they return one of two classes — **ANCHOR** (an address at which a measurement was taken) or **SUBJECT** (the claim is *about* the moving thing) — and that companion also carries the five adjudicated instances and the detection limits L1-L7 a mechanism enforcing this clause cannot see. Reaching a remedy below without having run the tests is indiscriminate pinning by another name.

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

### 2.2 Tool-provenance attribution — which build judged the tree

[ZONE:Evolvable] [HARD] A measurement produced by the project's own tooling is attributed to **two** coordinates, not one: the tree it read, and the build that judged it. §2 binds the first. This clause binds the second, because a tool invoked through a shell path resolves to an *installed* build, which need not be the build the tree describes.

The failure is silent by construction, and its silence is **symmetric**. Where the judging build is behind the tree, rules that landed after that build simply do not run: the output is a clean pass, the exit status is zero, the error stream is empty. Where the judging build matches the tree, the output is *also* a clean pass, the exit status is *also* zero, the error stream is *also* empty. Nothing in either result says which case occurred — so a green result is not evidence that the checks passed, only that whatever checks the invoked build happens to carry reported nothing.

**The obligation.** A tool measurement cited as evidence MUST have been produced by a build made from the tree under measurement. Concretely, either:

- build the tool from the tree and invoke that build **by its path**, rather than letting a shell path resolve to an installed one; or
- verify — and state alongside the citation — that the installed build's commit is not a strict ancestor of the tree's HEAD.

**What the citation carries.** A cited tool measurement names the judging build's commit next to the tree's HEAD. A measurement citing only the tree is unattributed under §2: a Gap, not a Claim.

**Where it does not bind.** A build with no repository to compare against — a released artifact inside a user's project, a checkout without history — has no lag to state, and this clause requires nothing of it. A missing second coordinate is a defect only where the coordinate exists.

**Not a substitute for the tooling's own verdict.** Where the tooling already computes a freshness verdict, that verdict is the mechanism; this clause governs the **citation**, and holds whether or not the invoked build is one that reports it. A build old enough to predate the freshness check is exactly the build that cannot warn you about itself.

### 2.3 Ordering attribution — the commit graph is the only sequencing witness

[ZONE:Evolvable] [HARD] When a claim's validity depends on a measurement having been taken BEFORE the change it measures (a baseline-first acceptance criterion), the baseline artifact MUST land in its own commit that precedes the change's commit. Git snapshots the tree per commit and cannot witness authoring order inside one commit, so a baseline committed together with the implementation it measured leaves the ordering claim permanently unverifiable — however truthfully the commit message asserts the sequence. A commit message and a session record ASSERT ordering; only the commit graph witnesses it. Where committing the baseline ahead is impossible (measurement and change are inherently one atomic act), the acceptance criterion's ordering clause is rewritten to what the commit graph can verify — never silently left to rest on a same-commit pair. (Motivating instance, recorded as a permanent deviation: SPEC-V3R6-GRAPH-FRESHNESS-001 AC-GF-022 — its baseline artifact `.moai/reports/t250/m5-baseline.md` shared commit `7f2e9e77d` with the implementation it measured, and the audit could not re-witness the asserted ordering from git history; recurrence prevention is this clause, tracked as backlog card t300.)

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
