# Verification-Claim Integrity

Doctrine establishing the **"no unobserved-verification-claim" invariant** for all MoAI actors. This rule is automatically loaded for the orchestrator and all agents. It is a policy-layer (codification) doctrine — it defines the norm; it does not itself run a runtime detector.

> Provenance: SPEC-EVIDENCE-CLAIM-INVARIANT-001 (IMP-06 of the fable-ish 13-agent "Verify, Don't Assume" analysis roadmap). The complementary mechanical advisory detection layer for one shape of this invariant's violation (code-session false-success) lives in SPEC-STOP-EVIDENCE-GATE-001 (IMP-02/03) — runtime, advisory, warn-first, fail-open. The two layers are complementary: this doctrine codifies the policy; that runtime gate detects one shape of its violation.

## 1. The Invariant — no unobserved-claim (verification OR defect)

[ZONE:Evolvable] [HARD] An actor MUST NOT assert a verification, a completion, **OR a defect / debt / drift** it did not actually verify with the domain's mechanical tooling.

> **Evidence absent ≠ evidence of success — NOR of failure.**

The absence of a failure signal is not, by itself, evidence that a check passed. A claim of "tests pass", "coverage 87%", "lint clean", or "0 0 sync" is only valid when the actor actually ran the command and observed its output. An unran command, a skipped step, or a silent assumption is a gap — never a pass.

Symmetrically, inferring a defect, a technical-debt item, a lifecycle drift, or an anomalous state from frontmatter text, grep matches, or file absence alone — without running the domain's dedicated verification tool — is not evidence that the defect exists. A text-pattern inference is a hypothesis, never a verified defect. The invariant binds both directions: an actor may not claim success it did not observe, and may not claim a defect it did not verify with the appropriate tool.

This is a policy-layer norm, not a mechanical guarantee. For the complementary mechanical-detection layer that surfaces one shape of this violation at runtime, cross-reference SPEC-STOP-EVIDENCE-GATE-001 (see Cross-References below).

### 1.1 Binding scope — ALL THREE surfaces

The invariant binds **all three** of the following surfaces. Each is named explicitly so none can claim exemption:

1. **Orchestrator self-report** — the orchestrator's own Completion Report and Verification Matrix banners, and its Trust-but-verify batches, as defined in `.claude/output-styles/moai/moai.md` §8 (Response Templates). When the orchestrator renders a Verification Matrix or Completion Report banner, every row it marks PASS MUST correspond to an actually-observed command output.

2. **Manager-agent completion report** — the `§E` self-verification (E1-E7) of `manager-develop` and `manager-docs`. When a manager agent reports an AC PASS/FAIL matrix (E1), cross-platform build result (E2), coverage (E3), subagent-boundary grep (E4), lint status (E5), or push state (E6), each reported result MUST be the verbatim output of a command the agent actually ran — not a summary, not an assumption, not a carry-over from a prior unrelated run.

3. **Defect / debt / drift identification claim** — any actor's assertion that a defect, technical-debt item, lifecycle drift, or anomalous state EXISTS and warrants action. A claim that "SPEC X is a close debt", "package Y has a coverage gap", or "N SPECs need Mx-close" is only valid when the actor ran the domain's dedicated verification tool (`moai spec audit`, `go test -cover`, `golangci-lint`, etc.) and observed its output. Inferring a defect from frontmatter text, grep matches, or file absence alone — without the dedicated tool — is an unobserved defect claim, and acting on it as if it were verified violates §2's attribution requirement. When a dedicated tool exists for a domain, text-only reasoning MUST NOT be the sole basis for a defect claim; the tool's output is the Evidence (§3.2).

## 2. Baseline-Integrity Attribution / baseline 무결성 귀속

[ZONE:Evolvable] [HARD] Every verification claim MUST be attributed to an actually-measured baseline — the command that was run plus the output that was observed.

A claim MUST NOT be assumed, and MUST NOT be carried over from a prior unrelated measurement. "Coverage is 87%" attributed to a baseline means: the actor ran `go test -cover ./internal/<pkg>/...` (the command) and observed `coverage: 87.0% of statements` (the output) in this run, against this tree. A number remembered from a different SPEC, a different package, or a different point in time is NOT a baseline — it is a carry-over, and using it as if it were a fresh measurement violates this attribution requirement.

Concretely, an attributed claim names:

- **The command** — the exact invocation that produced the evidence.
- **The observed output** — the verbatim result of that invocation in this run.

Anything else (an inferred value, a stale figure, a "should be" estimate) is unattributed and MUST be reported as a Gap (§3.4), not as a Claim.

## 3. The 5-Section Evidence-Bearing Report Format

[ZONE:Evolvable] [HARD] Verification and completion reports — on either binding surface (§1.1) — SHOULD be structured as the following five sections. The format is the operational mechanism that enforces §1 and §2: it forces the actor to separate what is claimed from what was observed, and to make the unobserved explicit. Apply the format to every report, not only the first.

### 3.1 Claim (주장)

What is being asserted. The completion or verification statement, phrased as a discrete claim (one row per assertion in a matrix, or one sentence per claim in prose).

### 3.2 Evidence (증거)

The actual command that was run **plus its verbatim output** — not a summary. If the claim in §3.1 is "tests pass", the Evidence section contains the literal command (`go test ./...`) and the literal output block it produced. Summarized evidence ("all tests passed") is NOT acceptable as Evidence — the verbatim output is the load-bearing artifact.

### 3.3 Baseline-attribution (baseline 귀속)

The baseline against which the claim was measured (per §2): the command + the observed output, in this run, against this tree. This section answers "measured against what?" and prevents a claim from silently borrowing a number from an unrelated prior measurement.

### 3.4 Gaps (미검증)

What was explicitly **NOT** observed — the negative space. This is the key defense of the entire format. By forcing the actor to enumerate what it did not verify, this section prevents an unobserved claim from passing silently as if it were a success. A report with an empty Gaps section is making the strong assertion that nothing was left unobserved — which itself must be true. When in doubt, name the gap.

### 3.5 Residual-risk (잔여 위험)

Remaining uncertainty and deferred verification — the risk that survives even after the observed evidence. Distinct from Gaps (§3.4, what was not observed): Residual-risk is what could still be wrong despite what WAS observed (flaky tests, environment-specific behavior, deferred AC, time-of-check-to-time-of-use windows, etc.).

## 4. Cross-References (SSOT — cross-reference only, do not duplicate)

This doctrine cross-references the following canonical surfaces. It does NOT copy their content — each remains the single source of truth for its own subject:

- `.claude/rules/moai/core/agent-common-protocol.md` § Skeptical Evaluation Stance — the fresh-judgment auditor stance (treat claims as suspect until evidence is shown).
- `.claude/rules/moai/core/moai-constitution.md` § Agent Core Behaviors #6 "Verify, Don't Assume" — the cross-cutting HARD behavior requiring evidence of completion.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § E (Self-Verification Deliverables, E1-E7) — the manager-agent §E self-verification matrix that the 5-section format generalizes and relates to.
- `.claude/rules/moai/workflow/verification-batch-pattern.md` — the orchestrator-side read-only verification batching pattern (the mechanism by which observed evidence is gathered efficiently).
- `.claude/output-styles/moai/moai.md` §8 — the Verification Matrix and Completion Report banners (the orchestrator self-report surface bound by §1.1).

## 5. Worked Example — Defect-Claim Hazard (2026-06-17)

A status report counted 29 SPECs with `status: implemented` and an absent `era:` frontmatter field. From frontmatter text alone, the reporter inferred "these 29 are V3R6 SPECs with a missing Mx-phase close — a close debt" and proposed batch-closing all 29.

This was an unobserved defect claim. The reporter had NOT run the domain's dedicated verification tool. When `moai spec audit --json` was finally run, its mechanical era classification showed all 29 were grandfather era (`V3R2-R4` 28 + `V2.x` 1) — `era_final: true`, protected, not subject to V3R6 4-phase close — and MUST-FIX drift across the entire catalog was 0. The inferred "close debt" did not exist; had the batch-close proceeded, 29 grandfather-protected SPECs would have been touched for no reason.

Lesson codified: **a defect claim is a hypothesis until the domain's tool confirms it.** The `era:`-absent + `implemented` text pattern was compatible with two contradictory interpretations (grandfather legacy vs. modern close-debt); only the dedicated tool could disambiguate. Whenever a domain verification tool exists (`moai spec audit` for SPEC lifecycle, `go test -cover` for coverage gaps, `golangci-lint` for code defects), its output MUST precede any defect/debt/drift claim — §1.1 surface 3 + §2 attribution.

---

Version: 1.1.0
Classification: Canonical Reference (policy-layer codification) — do not duplicate cross-referenced content; cross-reference this file instead.
