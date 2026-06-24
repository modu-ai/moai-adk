# Verification-Claim Integrity

Doctrine establishing the **"no unobserved-verification-claim" invariant** for all MoAI actors. This rule is automatically loaded for the orchestrator and all agents. It is a policy-layer doctrine — it defines the norm; it does not itself run a runtime detector.

> The motivating defect class is general: an actor claiming a verification or completion it did not actually observe. A complementary runtime layer (advisory, warn-first, fail-open) may detect one shape of this violation; this doctrine codifies the policy norm that binds every actor regardless of whether such a runtime layer is present.

## 1. The Invariant — no unobserved-claim (verification OR defect)

[ZONE:Evolvable] [HARD] An actor MUST NOT assert a verification, a completion, **OR a defect / debt / drift** it did not actually verify with the domain's mechanical tooling.

> **Evidence absent ≠ evidence of success — NOR of failure.**

The absence of a failure signal is not, by itself, evidence that a check passed. A claim of "tests pass", "coverage met", "lint clean", or "remote in sync" is only valid when the actor actually ran the command and observed its output. An unran command, a skipped step, or a silent assumption is a gap — never a pass.

Symmetrically, inferring a defect, a technical-debt item, a drift, or an anomalous state from text patterns, grep matches, or file absence alone — without running the domain's dedicated verification tool — is not evidence that the defect exists. A text-pattern inference is a hypothesis, never a verified defect. The invariant binds both directions: an actor may not claim success it did not observe, and may not claim a defect it did not verify with the appropriate tool.

This is a policy-layer norm, not a mechanical guarantee. A complementary mechanical-detection layer may surface one shape of this violation at runtime, but the norm binds every actor independently of that layer.

### 1.1 Binding scope — ALL THREE surfaces

The invariant binds **all three** of the following surfaces. Each is named explicitly so none can claim exemption:

1. **Orchestrator self-report** — the orchestrator's own Completion Report and Verification Matrix banners, and its trust-but-verify batches, as defined in `.claude/output-styles/moai/moai.md` (Response Templates). When the orchestrator renders a Verification Matrix or Completion Report banner, every row it marks PASS MUST correspond to an actually-observed command output.

2. **Manager-agent completion report** — the self-verification deliverables of `manager-develop` and `manager-docs`. When a manager agent reports an acceptance-criteria PASS/FAIL matrix, a build result, coverage, a boundary grep, lint status, or push state, each reported result MUST be the verbatim output of a command the agent actually ran — not a summary, not an assumption, not a carry-over from a prior unrelated run.

3. **Defect / debt / drift identification claim** — any actor's assertion that a defect, technical-debt item, drift, or anomalous state EXISTS and warrants action. A claim that "module X is broken", "package Y has a coverage gap", or "N items are stale and need cleanup" is only valid when the actor ran the domain's dedicated verification tool (the project's audit / lint / type-check / coverage command) and observed its output. Inferring a defect from text patterns, grep matches, or file absence alone — without the dedicated tool — is an unobserved defect claim, and acting on it as if it were verified violates §2's attribution requirement. When a dedicated tool exists for a domain, text-only reasoning MUST NOT be the sole basis for a defect claim; the tool's output is the Evidence (§3.2).

## 2. Baseline-Integrity Attribution / baseline 무결성 귀속

[ZONE:Evolvable] [HARD] Every verification claim MUST be attributed to an actually-measured baseline — the command that was run plus the output that was observed.

A claim MUST NOT be assumed, and MUST NOT be carried over from a prior unrelated measurement. "Coverage is at threshold" attributed to a baseline means: the actor ran the coverage command and observed the coverage figure in this run, against this tree. A number remembered from a different task, a different package, or a different point in time is NOT a baseline — it is a carry-over, and using it as if it were a fresh measurement violates this attribution requirement.

Concretely, an attributed claim names:

- **The command** — the exact invocation that produced the evidence.
- **The observed output** — the verbatim result of that invocation in this run.

Anything else (an inferred value, a stale figure, a "should be" estimate) is unattributed and MUST be reported as a Gap (§3.4), not as a Claim.

## 3. The 5-Section Evidence-Bearing Report Format

[ZONE:Evolvable] [HARD] Verification and completion reports — on either binding surface (§1.1) — SHOULD be structured as the following five sections. The format is the operational mechanism that enforces §1 and §2: it forces the actor to separate what is claimed from what was observed, and to make the unobserved explicit. Apply the format to every report, not only the first.

### 3.1 Claim (주장)

What is being asserted. The completion or verification statement, phrased as a discrete claim (one row per assertion in a matrix, or one sentence per claim in prose).

### 3.2 Evidence (증거)

The actual command that was run **plus its verbatim output** — not a summary. If the claim in §3.1 is "tests pass", the Evidence section contains the literal command and the literal output block it produced. Summarized evidence ("all tests passed") is NOT acceptable as Evidence — the verbatim output is the load-bearing artifact.

### 3.3 Baseline-attribution (baseline 귀속)

The baseline against which the claim was measured (per §2): the command + the observed output, in this run, against this tree. This section answers "measured against what?" and prevents a claim from silently borrowing a number from an unrelated prior measurement.

### 3.4 Gaps (미검증)

What was explicitly **NOT** observed — the negative space. This is the key defense of the entire format. By forcing the actor to enumerate what it did not verify, this section prevents an unobserved claim from passing silently as if it were a success. A report with an empty Gaps section is making the strong assertion that nothing was left unobserved — which itself must be true. When in doubt, name the gap.

### 3.5 Residual-risk (잔여 위험)

Remaining uncertainty and deferred verification — the risk that survives even after the observed evidence. Distinct from Gaps (§3.4, what was not observed): Residual-risk is what could still be wrong despite what WAS observed (flaky tests, environment-specific behavior, deferred criteria, time-of-check-to-time-of-use windows, etc.).

## 4. Cross-References (SSOT — cross-reference only, do not duplicate)

This doctrine cross-references the following canonical surfaces. It does NOT copy their content — each remains the single source of truth for its own subject:

- `.claude/rules/moai/core/agent-common-protocol.md` § Skeptical Evaluation Stance — the fresh-judgment auditor stance (treat claims as suspect until evidence is shown).
- `.claude/rules/moai/core/moai-constitution.md` § Agent Core Behaviors "Verify, Don't Assume" — the cross-cutting HARD behavior requiring evidence of completion.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § E (Self-Verification Deliverables) — the manager-agent self-verification matrix that the 5-section format generalizes and relates to.
- `.claude/rules/moai/workflow/verification-batch-pattern.md` — the orchestrator-side read-only verification batching pattern (the mechanism by which observed evidence is gathered efficiently).
- `.claude/output-styles/moai/moai.md` — the Verification Matrix and Completion Report banners (the orchestrator self-report surface bound by §1.1).

## 5. Worked Example — Defect-Claim Hazard

A status report counted N items matching a text pattern (for example, a metadata field absent from N files) and inferred "these N items are debt requiring action" — then proposed batch-modifying all N.

This was an unobserved defect claim: the domain had a dedicated verification tool, and it had not been run. The text pattern was compatible with two contradictory interpretations (items legitimately in a protected or legacy state versus items with a genuinely missing step); only the dedicated tool could disambiguate. When the tool was finally run, the inferred debt did not exist — the items were in their correct state — and had the batch modification proceeded, N items would have been touched for no reason.

Lesson codified: **a defect claim is a hypothesis until the domain's tool confirms it.** Whenever a domain verification tool exists (an audit command, a type checker, a linter, a coverage tool), its output MUST precede any defect / debt / drift claim — §1.1 surface 3 + §2 attribution. Text-pattern matching alone produces a candidate defect, never a verified one.

---

Version: 1.1.0
Classification: Canonical Reference (policy-layer codification) — do not duplicate cross-referenced content; cross-reference this file instead.
