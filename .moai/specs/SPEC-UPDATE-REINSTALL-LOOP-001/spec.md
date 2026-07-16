---
id: SPEC-UPDATE-REINSTALL-LOOP-001
title: "Break the moai update clean-reinstall loop + restore config preservation"
version: "0.1.1"
status: draft
created: 2026-07-17
updated: 2026-07-17
author: manager-spec
priority: Critical
phase: "v3.0.0"
module: "internal/defs, internal/cli, internal/template"
lifecycle: spec-anchored
tags: "update, clean-reinstall, deprecated-paths, settings-preservation, regression-guard"
issue_number: 1084
tier: M
---

# SPEC-UPDATE-REINSTALL-LOOP-001 — Break the `moai update` clean-reinstall loop

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-17 | manager-spec | Initial plan-phase authoring. Root cause verified: `.claude/rules/moai/design` is both a `DeprecatedPaths` entry and template-shipped, re-triggering clean-reinstall on every `moai update`. |
| 0.1.1 | 2026-07-17 | manager-spec | Plan-audit iter-1 clarification resolution: model-pin policy → merge-preserving (option b, R3 subsumes it); R2 (pre-flight hardening) SPLIT to follow-up `SPEC-UPDATE-PREFLIGHT-SAFETY-001` — this SPEC's scope is R1+R3 only; init system.yaml version-format → informational (scope not expanded); config-preservation extended to ALL `.moai/config/sections/*.yaml` (issue #1084 reports language.yaml/design.yaml loss). D3 fix: REQ-RIL-005 modality label corrected to (Event-driven). |

## §A Context and Problem

GitHub issue #1084 (Critical) reports that on a v3 project, `moai update` re-runs the
v2→v3 clean reinstall on **every** invocation, removing and recreating template-shipped
paths with zero net change, and — as a side effect of the clean-reinstall path — overwriting
user configuration (`settings.json`, `user.yaml`).

The orchestrator-verified root cause (see research.md §A for anchors) is a **single collision**:
the path `.claude/rules/moai/design` is simultaneously (1) enumerated in the `DeprecatedPaths`
registry and (2) shipped by the v3 embedded template. The v2-detection heuristic's
"deprecated path present" signal therefore fires perpetually; when the existing v3-version
override does not fire (it depends on the `system.yaml` version string being `v3.`-prefixed),
the clean-reinstall removes the path and the template redeploys it, re-arming the next update.

An existing partial fix (a v3-version negative-override) suppresses the loop **only** for
projects whose `moai.version` starts with the literal `v3.`. This SPEC delivers the
**version-format-independent, regression-proof** fix — eliminating the collision itself — plus
the config-preservation and pre-flight-safety hardening the clean-reinstall path currently lacks.

The reporter's file attribution (`harness-builder.md`) was a misdiagnosis — that file is not on
`DeprecatedPaths` — but the loop mechanism they observed is real.

## §B Requirements (GEARS)

### B.1 Break the loop (Critical)

- **REQ-RIL-001** (Ubiquitous): The `DeprecatedPaths` registry shall not contain any path
  that the embedded v3 template filesystem ships. Resolving the current collision means the
  stale `.claude/rules/moai/design` entry is removed (the template intentionally ships that
  directory now), NOT the template file.

- **REQ-RIL-002** (Event-driven): **When** `moai update` runs on a genuine v3 project that
  carries no true v2 residue, the update path **shall not** remove-and-reinstall any
  template-shipped path, so that consecutive `moai update` invocations each produce a
  zero-net-change working tree and no repeated clean-reinstall.

- **REQ-RIL-003** (Ubiquitous — regression guard): The test suite **shall** assert that the
  intersection of `defs.DeprecatedPaths` and the embedded template filesystem is empty, failing
  the build when any future entry (this one or a new one) collides with a shipped template path.

- **REQ-RIL-004** (State-driven): **While** the `DeprecatedPaths` slice is modified, the
  count and category-split test guards **shall** be updated in the same change, so the
  registry and its assertions never drift apart.

### B.2 Restore config preservation on the clean-reinstall path

- **REQ-RIL-007** (Event-driven): **When** the clean-reinstall reinstalls embedded templates,
  it **shall** preserve the user's `.claude/settings.json` customizations (including
  effortLevel, teammateMode, permissions, and non-MoAI hooks) rather than replacing them with
  template defaults.

- **REQ-RIL-008** (Ubiquitous): The clean-reinstall **shall** preserve EVERY user-populated
  `.moai/config/sections/*.yaml` file across the reinstall cycle — NOT only `user.yaml`. Issue
  #1084 explicitly reports loss of `language.yaml` and `design.yaml`; the invariant therefore
  covers the whole `sections/*.yaml` set (including the operator `name` in `user.yaml`, the
  conversation/documentation language settings in `language.yaml`, and the brand tokens in
  `design.yaml`), rather than blanking any of them to template defaults.

- **REQ-RIL-009** (Ubiquitous): The clean-reinstall's `settings.json` and `.moai/config/sections/*.yaml` (incl. `user.yaml`) handling
  **shall** use the same merge/preserve protection the normal `moai update` path already applies
  (the clean-reinstall must not be a lower-protection bypass of the normal path).

### B.3 Pre-flight safety — SPLIT to follow-up SPEC-UPDATE-PREFLIGHT-SAFETY-001

The pre-flight hardening requirements (fail-closed before destruction; no half-migrated end
state) are **NOT in this SPEC's scope**. Per the plan-audit iter-1 clarification decision they
are split to the follow-up **`SPEC-UPDATE-PREFLIGHT-SAFETY-001`**. This SPEC's scope is R1
(collision removal, §B.1) + R3 (config preservation, §B.2) only. The retired requirements are
retained here (labelled `[SPLIT → SPEC-UPDATE-PREFLIGHT-SAFETY-001]`) so the follow-up SPEC can
lift them verbatim, but they carry NO acceptance criteria in this SPEC's Definition of Done.

- **REQ-RIL-005** (Event-driven) `[SPLIT → SPEC-UPDATE-PREFLIGHT-SAFETY-001]`: **When** the
  clean-reinstall detects a symlink or a nested `.git` directory inside a PRESERVE scan root,
  the clean-reinstall **shall** abort before performing any destructive removal or template
  redeploy, and **shall** emit an actionable error that names the offending path.

- **REQ-RIL-006** (Ubiquitous) `[SPLIT → SPEC-UPDATE-PREFLIGHT-SAFETY-001]`: The clean-reinstall
  **shall not** leave a project in a half-migrated state — i.e. deprecated paths removed and
  templates reinstalled but the PRESERVE inventory not restored.

### B.4 Model pin policy (RESOLVED — merge-preserving)

- **REQ-RIL-010** (Where — capability gate): **Where** the deployed template pins a default
  Claude model, the pin **shall not** silently downgrade a project that is already configured
  for a higher-capability model. **Resolved policy (plan-audit iter-1):** option (b) —
  KEEP the template pin, but R3 (§B.2) makes the clean-reinstall's `settings.json` handling
  merge-preserving, so an existing user's `model` setting always wins; the pin applies ONLY to
  genuinely new projects (fresh `moai init`). No separate template-pin removal is required —
  REQ-RIL-009's merge-preservation subsumes the model-pin concern by construction.

## §C Exclusions

### Out of Scope — normal (non-clean-reinstall) update path
- The normal `moai update` path already 3-way merges `settings.json` and `.moai/config/sections/*.yaml`. This SPEC does not change that path; it only closes the clean-reinstall bypass so the clean path matches the normal path's protection.

### Out of Scope — pre-flight safety hardening (R2, split to a follow-up SPEC)
- Pre-flight validation (fail-closed symlink / nested-`.git` abort with a named-path error) and the no-half-migration invariant (REQ-RIL-005/006) are SPLIT to the follow-up **`SPEC-UPDATE-PREFLIGHT-SAFETY-001`** per the plan-audit iter-1 clarification decision. This SPEC delivers R1 (collision removal) + R3 (config preservation) only. Rationale: research.md §C.2 shows the current 7-step order already fails closed before destruction for both triggers, so R2's residual value is hardening (an explicit actionable error + rollback invariant), not a Critical defect fix — it does not belong on this Critical-path SPEC.

### Out of Scope — the v3-version negative-override redesign
- The existing version-string override (REQ-CRR-001, `SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002`) is retained as-is. This SPEC does not replace or re-engineer the version-string heuristic; the loop fix is deliberately version-format-independent (it removes the collision rather than tightening the detector).

### Out of Scope — init system.yaml version-string format (informational only)
- Whether the `moai init` generator always writes a `v3.`-prefixed `moai.version` (which governs how often the existing override silently fails in the wild) is treated as **informational only** per the plan-audit iter-1 clarification decision. It does NOT block or expand this SPEC's scope: R1's collision removal is version-format-independent, so the loop is broken regardless of the init version format. Tracing the init generator's format is deferred; it does not gate this SPEC.

### Out of Scope — general DeprecatedPaths content audit
- Only the template-collision invariant (REQ-RIL-001/003) is enforced. Auditing which paths *should* or *should not* be on the deprecated list, beyond the one confirmed collision, is a separate concern.

### Out of Scope — outputStyle default pin
- The `outputStyle: MoAI-Easy` template pin is an intentional product default (CLAUDE.local.md §22.6) and is not touched by REQ-RIL-010.

### Out of Scope — genuine v2→v3 migration behavior
- The v2→v3 clean-reinstall for real v2 projects (its 7-step canonical order, `.agency/` migration) remains functional and is not redesigned; this SPEC only prevents it from firing on already-migrated v3 projects and hardens its config-preservation.

## §D Success Definition

Consecutive `moai update` runs on a v3 project (with `.claude/rules/moai/design/` present)
perform zero deprecated-path removals and leave the working tree unchanged; user `settings.json`
and `user.yaml` customizations survive any clean-reinstall; a build-time guard makes a future
DeprecatedPaths↔template collision impossible to merge unnoticed.

## §E References

- research.md (this directory) — verified source anchors and the collision intersection result.
- Issue #1084 — original bug report.
- `SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001` / `-002` — the clean-reinstall path and the existing v3 override.
- `SPEC-DEPRECATEDPATHS-RECONCILE-001` — the 41-entry count governance.
