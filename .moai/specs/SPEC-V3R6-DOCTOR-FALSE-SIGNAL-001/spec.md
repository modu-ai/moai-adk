---
id: SPEC-V3R6-DOCTOR-FALSE-SIGNAL-001
title: "Repair moai doctor false-signal cluster (#1087 harness telemetry-presence FAIL, #1088 stale skills allowlist warning)"
version: "0.1.0"
status: in-progress
created: 2026-07-16
updated: 2026-07-16
author: manager-spec
priority: P1
phase: "v3.0.0-rc-stabilization"
module: "internal/cli"
lifecycle: spec-anchored
tags: "moai-doctor, false-signal, harness-5-layer, skills-allowlist, manifest-derivation, regression-repair"
tier: M
related_specs: [SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002, SPEC-DOCTOR-PROMOTION-001, SPEC-V3R3-HARNESS-001]
---

# SPEC-V3R6-DOCTOR-FALSE-SIGNAL-001

## §A Motivation

### §A.1 Shared Root Theme

`moai doctor` renders per-check PASS / WARN / FAIL signals to reassure users their project is healthy. Two OPEN regressions (v3.0.0-rc12, darwin/arm64) share one root theme: **the doctor conflates the mere on-disk PRESENCE of a runtime-generated artifact, or a hand-maintained STALE reference list, with a genuine "configured / known" state.** A merely-present telemetry file and a drifted hardcoded allowlist both produce false FAIL / warning signals on projects that are actually fine. This SPEC repairs both false-signal loci; it does NOT redesign the doctor architecture, add new checks, or change any check's PASS-path semantics for genuinely-healthy or genuinely-broken projects.

This SPEC is SPEC-2 of the v3.0.0-rc Stabilization Epic; the Epic entry SPEC (`SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002`) has landed independently. The two doctor defects are NOT gated on CLEAN-REINSTALL-002's implementation — they are independent diagnostic-layer false signals — but any fix that reaches the update or skills-sync code path MUST preserve that SPEC's user-asset preservation contract (see §A.4 and REQ-DFS-008).

### §A.2 Defect A — GitHub issue #1087 (OPEN, High): Harness 5-Layer FAILs on telemetry-only `.moai/harness/`

The Harness 5-Layer check (`internal/cli/doctor_harness.go` `runHarnessCheck`) decides "is a harness configured?" by a single `os.Stat(.moai/harness/)` directory-existence test (lines 23-28). When the directory is absent it reports `CheckOK` — `.moai/harness/ not present (no harness configured)`. When present it runs the L1-L6 layer battery.

The learning subsystem writes telemetry to `.moai/harness/usage-log.jsonl` (`harnessDefaultLogPath` in `internal/cli/harness.go`; the PostToolUse / Stop / SubagentStop observe hooks in `internal/cli/hook.go` append there). The FIRST recorded `tool_failure` observation CREATES the `.moai/harness/` directory. From that moment a project that never configured a v4 harness flips from PASS to FAIL:

```
✗ Harness 5-Layer  L1:PASS L2:FAIL L3:FAIL L4:FAIL L5:FAIL L6:PASS
  L5 missing: main.md, plan-extension.md, run-extension.md, ...
```

The observed cause: `checkLayer5Files` requires 7 baseline harness files (`main.md`, `plan-extension.md`, `run-extension.md`, `sync-extension.md`, `chaining-rules.yaml`, `interview-results.md`, `README.md`) — none of which exist in a telemetry-only directory. The directory-existence predicate treats "the learning hook wrote one log line" as equivalent to "the user configured a harness". Learning telemetry is a runtime byproduct, not a configuration act.

### §A.3 Defect B — GitHub issue #1088 (OPEN, High): Skills Allowlist reports 10 unknown skills after a fresh template install

The Skills Allowlist check (`internal/cli/doctor.go` `checkSkillsAllowlist` → `internal/cli/doctor_skills.go` `classifySkill`) validates each `.claude/skills/<dir>` name against `staticCoreAllowlist` — a hand-maintained 22-entry Go slice (the source comment claims "23"). Any `moai-`-prefixed directory NOT in that slice classifies as `WARN` (unknown skill).

The static slice has DRIFTED from the embedded template manifest. Right after `moai update --templates-only --yes` reinstalls the embedded templates (integrity PASSED), `moai doctor` reports:

```
! Skills Allowlist  10 unknown moai- skill(s) detected (run 'moai update' to sync)
```

— even though every `.claude/skills/moai-*` directory present was JUST installed by the same binary's embedded templates. Measured drift (this SPEC, plan-phase evidence): the embedded template ships 27 `moai-*` skill directories; the static allowlist knows 22, of which 5 (`moai-workflow-design-context`, `moai-workflow-design-import`, `moai-workflow-gan-loop`, `moai-domain-brand-design`, `moai-domain-copywriting`) no longer exist in the template. The exact 10 template skills MISSING from the allowlist — reproducing the "10 unknown" report byte-for-byte — are:

```
moai-domain-backend      moai-domain-database    moai-domain-frontend
moai-domain-html-report  moai-domain-humanize    moai-harness-learner
moai-ref-llm-security     moai-ref-secops          moai-ref-supply-chain
moai-workflow-ci-loop
```

The remediation hint (`run 'moai update' to sync`) is misleading: the skills WERE just synced by `moai update`; the drift is in the doctor's own hardcoded list, which `moai update` cannot touch. The allowlist must be DERIVED from the authoritative embedded template manifest so a template-fresh project reports zero unknown skills by construction.

### §A.4 Preservation Constraint (CLEAN-REINSTALL-002)

`SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002` owns the `moai update` clean-reinstall contract: version-aware backup-remove-reinstall, v2 FLAT layout handling, selective user-asset preservation (PRESERVE inventory + SHA-256 hash diff), and namespace protection (`hns-*` / `harness-*` user-owned skills never deleted). Neither doctor defect's fix requires touching `internal/cli/update.go` or `internal/harness/applier.go`. Both fixes are DIAGNOSTIC-side (read-only): Defect A adjusts a detection predicate in `doctor_harness.go`; Defect B changes the allowlist SOURCE in `doctor_skills.go`. This SPEC MUST NOT weaken the CLEAN-REINSTALL-002 preservation contract, and its verification includes a guard that the update/skills-sync paths remain behaviorally unchanged (REQ-DFS-008, AC-DFS-008).

### §A.5 User Story

As a moai user on a v3 project that has never configured a v4 harness, when the learning hook records a routine `tool_failure` observation, `moai doctor` MUST continue to report the harness as "not configured" (OK) — it MUST NOT flip to a red FAIL that implies my project is broken. As a user who just ran `moai update` to reinstall templates, `moai doctor` MUST report zero unknown `moai-` skills — the skills the same binary just installed MUST NOT be reported as strangers.

## §B Requirements

> **Notation:** GEARS (current canonical). `When` event-driven, `While` state-driven, `Where` capability-gate, `shall not` for unwanted behavior. REQ token prefix `REQ-DFS-NNN` (Doctor-False-Signal). AC token prefix `AC-DFS-NNN`.

### §B.1 Shared Root Invariant

**REQ-DFS-001 (Artifact-presence ≠ configured/known state).** The `moai doctor` diagnostic **shall not** treat the mere on-disk presence of a runtime-generated artifact (learning telemetry file) or a hand-maintained stale reference list as evidence of a "configured" or "known" state. A "configured" signal **shall** be derived from a genuine configuration marker (a baseline harness file), and a "known" signal **shall** be derived from the authoritative embedded template manifest — not from directory existence alone, and not from a static list that can silently drift from the manifest.

### §B.2 Defect A — Harness Telemetry-Presence False FAIL (#1087)

**REQ-DFS-002 (Telemetry exclusion from harness-configured detection).** **When** the Harness 5-Layer check (`runHarnessCheck`, **defined by SPEC-V3R3-PROJECT-HARNESS-001**) determines whether a harness is configured, it **shall** exclude runtime learning-telemetry artifacts — specifically `.moai/harness/usage-log.jsonl` and any file that is not a baseline harness file — from the determination. **When** the `.moai/harness/` directory contains ONLY learning telemetry (no baseline harness file), the check **shall** report the `CheckOK` "no harness configured" status, NOT run the L1-L6 battery.

**REQ-DFS-003 (Genuine-harness evaluation preserved).** **While** the `.moai/harness/` directory contains at least one baseline harness file (one of the 7 files enumerated by `checkLayer5Files`: `main.md` / `plan-extension.md` / `run-extension.md` / `sync-extension.md` / `chaining-rules.yaml` / `interview-results.md` / `README.md`), the Harness 5-Layer check **shall** proceed with the full L1-L6 evaluation exactly as before. The telemetry-exclusion predicate **shall not** suppress the layer battery for a genuinely-configured (or genuinely-misconfigured) harness.

**REQ-DFS-004 (No FAIL from a telemetry observation).** The Harness 5-Layer check **shall not** report a `CheckFail` (or any non-`CheckOK` status) on a project that has never configured a v4 harness solely because the learning subsystem recorded a `tool_failure` (or any) observation to `.moai/harness/usage-log.jsonl`.

### §B.3 Defect B — Stale Skills Allowlist False Warning (#1088)

**REQ-DFS-005 (Manifest-derived allowlist).** The Skills Allowlist check (`classifySkill` / `checkSkillsAllowlist`, **defined by SPEC-V3R3-HARNESS-001**) **shall** determine the set of known `moai-*` core skills by DERIVING it from the embedded template skills manifest — the same authoritative embedded source `moai update` installs (`internal/template/catalog.yaml` skill entries, or the embedded `templates/.claude/skills/moai-*` directory set) — **shall not** hard-code the set as a hand-maintained static slice that can drift from the manifest.

**REQ-DFS-006 (Zero unknowns on a template-fresh project).** **When** `moai doctor` runs on a project whose `.claude/skills/moai-*` directories were installed by the running binary's embedded templates, the Skills Allowlist check **shall** report zero unknown `moai-` skills (`CheckOK`).

**REQ-DFS-007 (Genuine-unknown detection + actionable hint).** **Where** a `moai-*` skill directory is present that is genuinely absent from the embedded manifest (a stale or third-party skill), the Skills Allowlist check **shall** still classify it as unknown (`CheckWarn`) — the manifest derivation **shall not** disable the check. **When** an unknown is reported, the remediation hint **shall** be actionable and **shall not** direct the user to `moai update` for a drift that `moai update` cannot resolve.

### §B.4 Preservation Guard

**REQ-DFS-008 (CLEAN-REINSTALL-002 contract preserved).** **Where** any fix in this SPEC would touch the `moai update` or skills-sync code paths, the fix **shall not** weaken the `SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002` preservation contract (version-aware backup-remove-reinstall, v2 FLAT layout, selective user-asset preservation, namespace protection). Both doctor fixes **shall** be read-only diagnostics confined to `internal/cli/doctor_harness.go` and `internal/cli/doctor_skills.go` (plus a manifest-read helper and tests); they **shall not** modify update-phase behavior, and the existing CLEAN-REINSTALL preservation tests **shall** continue to pass unchanged.

## §C Out of Scope

This SPEC is a false-signal regression repair, NOT a doctor redesign.

### Out of Scope — Doctor architecture and new checks
- Adding new diagnostic checks, renaming existing checks, or altering the doctor output format / status vocabulary (`CheckOK` / `CheckWarn` / `CheckFail`).
- Changing any check's PASS-path or genuine-FAIL-path semantics beyond the two false-signal loci named in §B.

### Out of Scope — Harness telemetry data path
- Relocating the `usage-log.jsonl` write path out of `.moai/harness/` (the alternative fix candidate for Defect A). The DEFAULT fix is the doctor-side detection predicate (REQ-DFS-002); a write-path relocation carries migration + backward-compat cost and is deferred. If run-phase discovers the doctor-side fix is infeasible, that pivot is a blocker-report event, not a silent scope expansion.
- Redesigning the learning subsystem, the observe hooks, or the `AggregatePatterns` pipeline.

### Out of Scope — Update / clean-reinstall behavior
- Any modification to `internal/cli/update.go`, `internal/harness/applier.go`, or the CLEAN-REINSTALL-002 clean-reinstall code path. This SPEC only READS the embedded manifest; it does not change how templates are installed.
- Namespace-protection logic, PRESERVE inventory semantics, or the SHA-256 hash-diff detection (all owned by CLEAN-REINSTALL-002 and its predecessors).

### Out of Scope — Manifest maintenance tooling
- Building a CI guard or generator that keeps `catalog.yaml` regenerated on skill add/remove (that manifest is already generated). This SPEC consumes the manifest; it does not own the manifest's own freshness pipeline.

## §D Acceptance Criteria

The full Given-When-Then acceptance matrix (AC-DFS-001 .. AC-DFS-009) is enumerated in `acceptance.md`. Each AC is reproduction-test-shaped (Go test asserting observable check status), NOT a token-presence grep, because both defects are false-signal regressions with deterministic reproduction steps — the reproduction IS the RED test.

## §E History

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-16 | manager-spec | Initial plan-phase authoring. Two OPEN v3.0.0-rc12 regressions (#1087 harness telemetry-presence FAIL, #1088 stale skills allowlist warning) sharing the artifact-presence-≠-configured-state root theme. Tier M. Drift quantified at plan-phase: 27 embedded moai-* skills, 10 missing from the 22-entry static allowlist (exact match to issue #1088's "10 unknown"). |
