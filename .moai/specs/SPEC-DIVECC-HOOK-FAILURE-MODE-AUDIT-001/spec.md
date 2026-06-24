---
id: SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001
title: "Hook Defense-in-Depth Shared-Failure-Mode Audit + Independence Rule"
version: "0.1.0"
status: completed
created: 2026-06-22
updated: 2026-06-22
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai/development, .claude/hooks/moai"
lifecycle: spec-anchored
tags: "hooks, defense-in-depth, audit, doctrine, dogfooding, divecc"
era: V3R6
---

# SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001 — Hook Defense-in-Depth Shared-Failure-Mode Audit

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-22 | manager-spec | Initial plan-phase draft. Entry SPEC (N1) of Epic Dive-into-CC. Premise VERIFIED at authoring (evidence in research.md §A). |

---

## §A. Background

### A.1 Epic provenance (Dive-into-CC dogfooding)

This is the entry SPEC (candidate N1) of **Epic Dive-into-CC** (see `ROADMAP.md` in this directory). The Epic applies findings from an academic reverse-engineering analysis of Claude Code to moai-adk's own harness — a self-improvement (dogfooding) exercise. The source is one body of work on two surfaces: arXiv:2604.14228 "Dive into Claude Code: The Design Space of Today's and Future AI Agent Systems" (Liu, Zhao, Shang, Shen, 2026, cs.SE — a reverse-engineering of Claude Code v2.1.88), and its companion repository github.com/VILA-Lab/Dive-into-Claude-Code with the "Build Your Own AI Agent: A Design Guide".

The paper insight that motivates N1: **"Defense-in-depth fails when layers share failure modes."** A defense layered out of components that all collapse under one condition is not depth — it is one layer wearing many hats. moai-adk's hook layer (`.claude/hooks/moai/`) is exactly such a defense surface, so this insight is directly testable against moai-adk's own tree.

### A.2 Verified evidence (NOT hypothesis)

The shared-failure-mode premise for this SPEC was grounded by read-only inspection of the moai-adk tree at plan-phase and is recorded VERBATIM in `research.md` §A. It is recorded here as **observed fact**, not hypothesis, per `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 (a defect/risk claim is valid only when the domain's tooling — here, `grep`/`Read` over `.claude/hooks/moai/` — was actually run and its output observed). The grep/read commands and their verbatim output are in research.md; this section states the conclusions those observations support.

**Shared failure mode A — moai-binary resolution chain (the strongest shared mode).** All **31** `handle-*.sh` wrapper scripts share an identical 3-tier moai-binary resolution chain: (1) `command -v moai` on PATH, then (2) `$HOME/go/bin/moai`, then (3) `$HOME/.local/bin/moai`, falling through to a silent `exit 0` when none resolve. The grep `grep -l 'command -v moai' *.sh` returns exactly the 31 wrappers (and only the wrappers); no wrapper lacks the chain. The single point of correlated failure is therefore not "the `moai` binary in PATH" alone (the orchestrator's initial framing) but the **conjunction of all three resolution tiers failing** — when `moai` is absent from PATH AND from `$HOME/go/bin` AND from `$HOME/.local/bin`, all 31 wrappers degrade together to silent no-ops. This correlated degradation is corroborated by an already-documented real failure in `CLAUDE.md` §17 Troubleshooting: "`moai hook subagent-stop` fails | Binary not in PATH | Run `which moai`". The paper's exact warning, made concrete.

> **Honest refinement vs the orchestrator's initial evidence.** The orchestrator's grounding described mode A as a "single condition — the `moai` binary present in PATH". Reading the wrappers fully revealed a 3-tier chain, not a single `command -v moai` gate. The shared-failure-mode conclusion stands (all 31 share the *identical* chain and degrade together), but the trigger is the conjunction of all three tiers failing, not the first tier alone. This refinement is recorded because the verification-claim-integrity invariant binds in both directions: we may not over-state a risk any more than we may over-state a success.

**Shared failure mode B — `--skip-hook` bypass (by-design).** All **3** governance gates — `status-transition-ownership.sh`, `sync-phase-quality-gate.sh`, `team-ac-verify.sh` — share an identical first-argument bypass: `if [ "$1" = "--skip-hook" ]; then ...skipped... exit 0`. One flag disables all three. This is **by-design and audit-logged**: each gate appends a skip record to the shared `.moai/logs/hook-skip.log` before exiting per `.claude/rules/moai/core/agent-common-protocol.md` (the `--skip-hook` opt-out is a documented, logged override). The SPEC treats this as a documented shared-bypass surface to *evaluate and classify*, not as a presumed defect.

**Positive / independent signal — governance gates do NOT share mode A.** The 3 governance gates do **not** appear in the `command -v moai` grep — they are self-contained bash that depend on `git`, `jq`, `grep`, `awk` rather than on the `moai` binary. The governance-gate layer is therefore genuinely more independent of the wrapper layer than mode A alone would suggest: the wrapper layer's strongest shared failure (the moai-binary chain) does NOT propagate to the governance layer. This is the depth that mode-A correlation lacks, and the audit must document it as the contrasting positive case.

### A.3 What this SPEC is (and is not)

This SPEC delivers an **AUDIT + a new independence rule/doctrine**. It is **NOT** a rewrite of the hooks. The hook scripts are inspected and classified; they are not modified by this SPEC. Concretely the deliverable is:

1. An enumeration of every shared failure mode across the hook layer, each with its observed grep/read evidence.
2. A classification of each shared mode as **acceptable-by-design** vs **genuine-risk**.
3. A new rule (`hook-independence.md`) documenting the shared-failure-mode catalogue plus a "does this new hook introduce an independent failure mode?" authoring checklist.
4. **Optional** mitigation *recommendations* (not implementations) — e.g. whether the wrappers already carry a fallback branch (they do — the 3-tier chain), and whether any additional fallback for the all-three-absent case is warranted.

---

## §B. Requirements (GEARS)

> GEARS notation per `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Format. Subjects are generalized (the audit, the rule, the catalogue) rather than the hardcoded "the system".

### B.1 Audit requirements

- **REQ-DIVECC-001 (Ubiquitous)**: The audit **shall** enumerate every shared failure mode present across the scripts in `.claude/hooks/moai/`, where a "shared failure mode" is any single dependency, condition, flag, or convention whose failure correlates the degradation of two or more hook scripts.

- **REQ-DIVECC-002 (Ubiquitous)**: For each enumerated shared failure mode, the audit **shall** cite the observed `grep`/`Read` evidence (the command run and the result observed) that establishes the mode, per `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3.

- **REQ-DIVECC-003 (Ubiquitous)**: The audit **shall** classify each enumerated shared failure mode as exactly one of `acceptable-by-design` or `genuine-risk`, with a stated rationale for the classification.

- **REQ-DIVECC-004 (When)**: **When** the audit inspects a governance gate (`status-transition-ownership.sh`, `sync-phase-quality-gate.sh`, `team-ac-verify.sh`), it **shall** record, for that gate, whether it shares each of: (a) the moai-PATH resolution chain, (b) the `--skip-hook` bypass, (c) the `--skip-hook` shared log file (`.moai/logs/hook-skip.log`), (d) the `jq` dependency, (e) the `${CLAUDE_PROJECT_DIR:-$PWD}` root-fallback convention, (f) the `set -e` convention, (g) the configured timeout (5s / 10s).

- **REQ-DIVECC-005 (When)**: **When** the audit inspects a `handle-*.sh` wrapper, it **shall** record whether the wrapper carries a fallback branch beyond the first `command -v moai` tier (the `$HOME/go/bin` and `$HOME/.local/bin` tiers).

- **REQ-DIVECC-006 (Ubiquitous)**: The audit **shall** record the contrasting positive signal that the governance gates do NOT share shared-failure-mode A (the moai-binary resolution chain), establishing that the wrapper layer's strongest shared failure does not propagate to the governance layer.

### B.2 Rule (doctrine) requirements

- **REQ-DIVECC-007 (Ubiquitous)**: A new rule file `hook-independence.md` **shall** exist under `.claude/rules/moai/development/` documenting the shared-failure-mode catalogue produced by the audit.

- **REQ-DIVECC-008 (Ubiquitous)**: The `hook-independence.md` rule **shall** contain a "does this new hook introduce an independent failure mode?" authoring checklist that a hook author can apply before adding a new hook.

- **REQ-DIVECC-009 (Where)**: **Where** a shared failure mode is classified `acceptable-by-design`, the rule **shall** state the design justification (so a future reader does not misread an intentional shared bypass as a defect — Chesterton's Fence).

- **REQ-DIVECC-010 (Ubiquitous)**: The rule **shall** cross-reference, rather than duplicate, the canonical hook surfaces (`hooks-system.md`, `agent-common-protocol.md` § Hook Invocation Surface, `runtime-recovery-doctrine.md`) per the SSOT-no-duplication convention.

### B.3 Mitigation-recommendation requirements (optional content, mandatory placeholder)

- **REQ-DIVECC-011 (Where)**: **Where** the audit classifies a shared failure mode as `genuine-risk`, the SPEC **shall** record a mitigation *recommendation* (not an implementation) for that mode, OR explicitly record that no mitigation is recommended with a rationale.

- **REQ-DIVECC-012 (Ubiquitous)**: The SPEC **shall not** modify any script in `.claude/hooks/moai/` — the deliverable is an audit + doctrine, and any hook-script change is deferred to a separate follow-up SPEC.

### B.4 Run-phase constraint requirements (template-first, neutrality)

- **REQ-DIVECC-013 (When)**: **When** the new rule `hook-independence.md` is created at run-phase, it **shall** first be authored in the template source `internal/template/templates/.claude/rules/moai/development/hook-independence.md`, then propagated via `make build`, per `CLAUDE.local.md` §2 Template-First Rule. (Run-phase note — not acted on at plan-phase.)

- **REQ-DIVECC-014 (Where)**: **Where** the template copy of `hook-independence.md` is authored, it **shall not** leak internal SPEC IDs, internal dates, or commit SHAs, per `CLAUDE.local.md` §25 Template Internal-Content Isolation. (Run-phase note — the local `.claude/` copy may reference this SPEC; the template mirror must be neutral.)

---

## §C. Out of Scope

The exclusions below are out of scope for this SPEC. Each is expressed as an `### Out of Scope — <topic>` H3 sub-heading with bullet items, satisfying the `OutOfScopeRule` (`MissingExclusions`) lint.

### Out of Scope — Hook script rewrites

- Modifying, refactoring, or rewriting any script in `.claude/hooks/moai/` (the 31 wrappers or the 3 governance gates). This SPEC inspects and classifies; it does not change hook behavior. Any hook-script change is a separate follow-up SPEC.
- Adding a new fallback tier to the `handle-*.sh` wrappers for the all-three-absent case. If the audit recommends this, the recommendation is recorded — but the implementation is out of scope.

### Out of Scope — Mechanical enforcement of the independence checklist

- Building a lint rule, hook, or CI check that mechanically enforces the "independent failure mode?" authoring checklist. This SPEC delivers the checklist as authoring doctrine; mechanical enforcement (e.g. a meta-hook that scans new hook scripts) is a future SPEC.
- Wiring the `runtime-recovery-doctrine.md` §4 recovery-signal carve-out to parse `stopReason` (already deferred to a future runtime-layer SPEC — not in N1 scope).

### Out of Scope — Other Epic Dive-into-CC candidates

- Candidates N2–N7 of Epic Dive-into-CC (extension context-cost ladder, delegation token-cost signal, observability loop, compaction naming cross-ref, unified inventory, paper archival). Each is its own SPEC; N1 covers only the hook shared-failure-mode audit + independence rule.

### Out of Scope — Claude Code runtime internals

- Reimplementing or modifying any Claude Code runtime behavior (the hook dispatcher, compaction layers, the query loop). moai-adk is a harness ON TOP of Claude Code; this SPEC audits moai-adk's own hook scripts, not the runtime that invokes them.

---

## §D. Acceptance Criteria (summary)

Full Given-When-Then acceptance criteria are in `acceptance.md`. Summary of the binding gates:

- AC-DIVECC-001: every shared failure mode in the audit cites observed grep/read evidence (no unobserved claim).
- AC-DIVECC-002: each shared failure mode is classified acceptable-by-design vs genuine-risk with rationale.
- AC-DIVECC-003: `hook-independence.md` exists with the catalogue + the authoring checklist.
- AC-DIVECC-004: the SPEC modifies zero hook scripts (deliverable is audit + doctrine only).
- AC-DIVECC-005: the positive signal (governance gates do not share mode A) is documented.

---

## §E. Cross-References

- `ROADMAP.md` (this directory) — Epic Dive-into-CC, candidate N1.
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 — the defect/risk-claim grounding invariant that binds §A.2.
- `.claude/rules/moai/core/hooks-system.md` — canonical hook event + execution-type reference (the rule cross-references, does not duplicate).
- `.claude/rules/moai/core/agent-common-protocol.md` § Hook Invocation Surface — the 3 governance gates' owning REQs + exit-code semantics + the `--skip-hook` audit-log convention.
- `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` §4 — the recovery-signal carve-out (a related but separate hook concern; N1 cross-references it for completeness).
- `CLAUDE.md` §17 Troubleshooting — the documented "`moai hook ... fails | Binary not in PATH`" real-world corroboration of shared failure mode A.
- `CLAUDE.local.md` §2 (Template-First) + §25 (Template Internal-Content Isolation) — run-phase constraints REQ-DIVECC-013/014.
