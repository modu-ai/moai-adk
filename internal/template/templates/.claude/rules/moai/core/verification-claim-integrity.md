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

Version: 1.2.0
Classification: Canonical Reference (policy-layer codification) — do not duplicate cross-referenced content; cross-reference this file instead.
