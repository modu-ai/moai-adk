---
title: Verification-Claim Integrity
weight: 50
draft: false
---
# Verification-Claim Integrity

When an agent says "tests passed", "coverage is 87%", or "this code is no longer used so it's safe to delete", how can you trust that statement? **Verification-Claim Integrity** addresses exactly this point. It's the rule that prohibits claiming unobserved success and asserting unverified defects.

## Why This Rule Is Needed

Agents are trained toward a gradient bias to report "all done." Users want to hear positive outcomes, so the model learns to summarize results optimistically. Without system-level correction of this bias, "passed" becomes unreliable.

There's a more dangerous direction. When an agent claims "this SPEC is missing close handling," "this package has low coverage," or "this feature is still in use so don't remove it," that claim may be false. Inferring defects from text patterns alone, or judging that a feature is alive solely because a reference exists, wastes effort fixing non-existent problems or reviving already-deleted code.

{{< callout type="warning" >}}
**Absence of evidence is not evidence of success, nor evidence of failure.** Just because a check wasn't run doesn't mean you can claim it "passed" or "failed." Both are unobserved claims.
{{< /callout >}}

## The Invariant — Prohibit Unobserved Claims

The core rule is one sentence:

> An agent must not assert completion, defect, debt, drift, or the premise underlying a recommendation it did not actually verify.

This rule binds in both directions:

- **Success direction** — Claims that "tests passed," "coverage is 87%," or "lint is clean" are valid only when the command was actually run and its output observed. Commands not run, steps skipped — those are gaps, not passes.
- **Defect direction** — Claims that "this SPEC is close debt," "this package has low coverage," or "this code is drift" are also valid only when verified with the domain's dedicated tool. Inferring defects from frontmatter text or grep results alone is an unobserved defect claim.

The defect direction is more dangerous. A wrong "delete it" claim is promptly contradicted by the next build or test, but a wrong "keep it" claim preserves dead code with no signal at all.

## Four Binding Surfaces

The rule doesn't bind to just one surface. It binds all four places where unobserved claims can leak in.

| Surface | What It Binds | Example |
|---------|---------------|---------|
| **Orchestrator self-report** | Every PASS row in completion reports and verification matrices | Each row in a "all verification passed" banner |
| **Agent completion report** | Agent self-verification matrices (test·build·coverage·lint) | "Tests PASS, coverage 87%" report |
| **Defect·debt·drift claim** | Every claim that a defect or technical debt "exists" | "SPEC X is close debt", "package Y has low coverage" |
| **Recommendation-premise claim** | The premise that grounds a "keep/remove" recommendation | "This feature is still alive", "another consumer depends on it" |

The fourth surface is the rule that recommendation premises must be verified. To recommend against removal, you must directly verify that the feature's producer still exists and that consumers are still reachable. **Reachability is not justification.** A reference existing is not evidence the referent is alive.

## Five-Section Report Format

The operational mechanism that actually enforces this rule is the **five-section report format**. When agents report completion or verification divided into five sections, they separate what was observed from what wasn't, and unobserved portions surface as gaps.

### 1. Claim

State what's being asserted in one sentence. One row in a verification matrix, or one sentence in a prose report, is one claim unit.

### 2. Evidence

Write the actual command that was run **plus its output verbatim**. Not a summary. When the claim is "tests passed," this section contains the literal `go test ./...` command and the verbatim output block that command emitted. A summary like "all tests passed" is NOT evidence. The output itself is the load-bearing artifact.

### 3. Baseline-attribution

State what baseline this claim was measured against. The command run and the output observed, in this run, against this tree. Don't borrow a number remembered from a different SPEC, a different package, or a different point in time. When you say "coverage 87%," that means you just ran `go test -cover ./internal/<pkg>/...` on this tree and observed `coverage: 87.0% of statements`.

### 4. Gaps

List what was explicitly **NOT verified**. This section is the key to the entire format. It prevents an unobserved claim from passing silently as if it were a success. When the Gaps section is empty, that's a strong claim that "nothing was left unverified," and that itself must be true. When in doubt, don't leave it blank — name it.

### 5. Residual-risk

Uncertainty that remains even after observed evidence. This differs from Gaps (what was not observed). Residual-risk is what could still be wrong despite what WAS observed — flaky tests, environment-specific behavior, deferred acceptance criteria, time-of-check-to-time-of-use windows, and so on.

## Worked Example — Defect-Claim Hazard

A status report counted 29 SPECs with `status: implemented` and an absent `era:` field, then inferred from frontmatter text alone that "these 29 are V3R6 SPECs with a missing close" and proposed batch-closing all of them.

This was an unobserved defect claim. The reporter had NOT run the domain's dedicated verification tool. When `moai spec audit --json` was finally run, its mechanical era classification showed all 29 were grandfather era (`V3R2-R4` 28 + `V2.x` 1) — `era_final: true`, protected, not subject to V3R6 3-phase close. MUST-FIX drift across the entire catalog was 0. The inferred "close debt" did not exist; had the batch-close proceeded, 29 grandfather-protected SPECs would have been touched for no reason.

{{< callout type="info" >}}
**Lesson codified**: A defect claim is a hypothesis until the domain's tool confirms it. The `era:`-absent + `implemented` text pattern is compatible with two contradictory interpretations (grandfather legacy vs. modern close-debt). Only the dedicated tool can disambiguate. Whenever a domain verification tool exists (`moai spec audit` for SPEC lifecycle, `go test -cover` for coverage gaps, `golangci-lint` for code defects), its output MUST precede any defect/debt/drift claim — §1.1 surface 3 + §2 attribution.
{{< /callout >}}

## Worked Example — Retention-Claim Hazard

A user instructed that `.moai/brain` be removed. The orchestrator deleted the artifacts but held one item back — a scan in the shipped `plan/context-discovery.md` that globbed `.moai/brain/IDEA-*/proposal.md` — on the stated premise that removing it "would withdraw a live feature from every distributed user," and recommended a separate retirement SPEC instead.

That premise was never checked. The orchestrator had verified the scan was *reachable* (`plan.md`'s routing table points at it) and had read that `SPEC-V3R3-BRAIN-001` still carried `status: implemented`, then treated both facts as evidence the feature was live. Neither establishes that. When the producers were finally enumerated, every one was already gone: the `/moai brain` command, `workflows/brain.md`, the `manager-brain` agent, the `moai brain` CLI, the `/moai project --from-brain` flag, the `templates/.moai/brain/` scaffold, and the docs-site pages in all four locales. `SPEC-SUBCOMMAND-RETIRE-001` (status: completed) had retired the feature from the template source permanently, for all distributed users, and a later cleanup commit had swept the orphans that retirement left behind. The scan simply survived both passes. With no producer and no scaffold, the glob could only ever return zero on a user's machine.

Lesson codified: **reachability is not justification, and a SPEC still reading `status: implemented` is not proof the feature it delivered is still live** — a later SPEC may have retired it. Before recommending retention against an instruction, enumerate the producers of the thing being retained and check for a completed retirement SPEC; an objection whose premise was never verified is an unobserved claim — §1.1 surface 4 + §2 attribution.

## Why This Rule Matters to Users

This rule is the systemic expression of the habit of asking "what's the evidence?" when reading agent reports. MoAI-ADK forces agents to report in five sections instead of ending with "passed," and accepts as evidence only results verified directly with domain tools. The result:

- **You can trust completion reports** — "tests passed" now means actual command and output.
- **Fewer false-defect code touches** — defect claims go through tool verification.
- **Dead code stays dead** — retention recommendations must verify producer existence.

The [TRUST 5 Quality](/en/core-concepts/trust-5/) "Trackable" principle becomes reality here. When every claim is attributed to observed evidence, reports become auditable.

## Related Documents

- [TRUST 5 Quality](/en/core-concepts/trust-5/) — Five quality principles, Trackable principle's parent frame
- [SPEC-based development](/en/core-concepts/spec-based-dev/) — 3-phase lifecycle where acceptance criteria are judged against evidence
- [Harness engineering](/en/core-concepts/harness-engineering/) — The paradigm for designing agent environments
