---
id: SPEC-GITSTRAT-WORKFLOW-READER-001
title: "Validated multi-flow interpreter for git_strategy.<mode>.workflow — 4-value validation, per-flow integration-target interpretation, characterization-first extension"
version: "0.1.0"
status: in-progress
created: 2026-09-14
updated: 2026-09-14
author: GOOS행님
priority: P2
phase: "v3.2.0 target"
module: "internal/config"
lifecycle: spec-anchored
tags: "git-strategy, workflow, config-reader, validation, integration-target, tdd, t656"
tier: M
related_specs: [SPEC-V3R5-GIT-STRATEGY-SCHEMA-001, SPEC-WORKTREE-BASEREF-001]
---

# SPEC-GITSTRAT-WORKFLOW-READER-001 — Validated multi-flow interpreter for `git_strategy.<mode>.workflow`

## A. Intent / Provenance (WHAT & WHY)

### A.1 Discarded premise (card t656 re-scope)

The original card premise — "`git_strategy.<mode>.workflow` has 0 production readers" — is **DISCARDED as stale**. Cards t449 and t637 already landed a reader: `internal/config/loader_integration_branch.go` reads the active mode profile's `Workflow` field through `LoadGitFlowIntegrationConfig` (provenance and measured evidence in `research.md` §2). The reader is production-live with two consumer sites (`moai integration acquire`, the session-exit auto-merge gate). `internal/config/loader_slot_lease.go` is NOT a consumer — its line-6 comment states it is *modelled on* `LoadGitFlowDevelopBranch` (design precedent) and it reads an entirely different key (`workflow.slot_lease.default_max_duration`); the repo has zero non-test callers of the reader from slot-lease code (plan-audit iter1 D5, `git log -S` verified).

### A.2 Re-scoped goal

The existing reader is a **single-value discriminator**: it answers exactly one question — is the workflow `git-flow` — and folds every other value (`github-flow`, `gitlab-flow`, `release-flow`, typos, and unsupported flows) into one indistinguishable "not git-flow" bucket. A user who typos `git-flwo` gets the same silent fallback as a user who deliberately picked `github-flow`.

This SPEC plans the extension of that reader into a **validated, multi-flow interpreter**:

1. `Workflow` values are validated against exactly 4 allowed values: `github-flow`, `git-flow`, `gitlab-flow`, `release-flow`. `trunk-based` is deliberately EXCLUDED (operator decision, card t656) — it is an unsupported value and must be rejected/diagnosed, not silently tolerated.
2. Each allowed flow carries an **integration-target interpretation** (§C REQ-GWS-004). Its in-SPEC production consumer is the new `moai doctor` check (REQ-GWS-009): the check validates the configured flow and names its standing branches and integration target. The acquire/automerge consumers are NOT rewired to per-flow targets in this SPEC — REQ-GWS-008 keeps their observable behavior unchanged; wiring them is a follow-up card candidate.
3. Existing git-flow behavior MUST be bit-identical: characterization tests are written FIRST (before any production extension) pinning today's `LoadGitFlowIntegrationConfig` / `IsGitFlow` / `LoadGitFlowDevelopBranch` behavior.

### A.3 Constitution alignment

Development mode: **tdd** (`.moai/config/sections/quality.yaml` `constitution.development_mode`), coverage target **85%** (`test_coverage_target`). Template default stays `github-flow` (verified in-tree: `git-strategy.yaml.tmpl` lines 18/50/86 and `internal/config/defaults.go:763,775,788`).

## B. Requirements (GEARS)

### B.1 Validation (REQ-GWS-001..003)

- REQ-GWS-001 (Ubiquitous): The workflow reader shall validate `git_strategy.<mode>.workflow` against exactly the allowed set {`github-flow`, `git-flow`, `gitlab-flow`, `release-flow`} and shall report a three-way disposition — `git-flow` / `valid-non-git-flow` / `invalid` — instead of the current two-way (git-flow / everything-else) answer.

- REQ-GWS-002 (Event-driven): **When** the active mode profile's `Workflow` value is not a member of the allowed set, the reader shall report the value as `invalid` while carrying the offending raw string, and `IsGitFlow()` shall return false.

- REQ-GWS-003 (Event-driven): **When** a consumer (acquire, auto-merge gate) encounters an `invalid` workflow disposition, the consumer shall emit an explicit diagnostic naming the offending value and the allowed set, and shall fall back to the caller-provided target exactly as the current non-git-flow path does. The reader itself shall not fail — the failure surface is diagnosis, not load refusal (the fail-open contract of `loader_integration_branch.go` is preserved).

### B.2 Per-flow integration-target interpretation (REQ-GWS-004..006)

- REQ-GWS-004 (Ubiquitous): The reader shall expose a per-flow integration-target interpretation mapping each allowed flow to its target-branch source: `github-flow` → the fixed default `main`; `git-flow` → the `develop_branch` key (current behavior); `gitlab-flow` → the profile's `environment` key; `release-flow` → the profile's `release_branch_prefix` key.

- REQ-GWS-005 (State-driven): **While** the resolved flow is `git-flow`, the reader shall continue to gate `develop_branch` adoption on `Manual && GitFlowWorkflow` exactly as today (types.go contract "manual mode, git-flow only" preserved), and `develop_branch` shall remain ignored under every other flow.

- REQ-GWS-006 (Event-driven): **When** the resolved flow is `gitlab-flow` or `release-flow` and its target key (`environment` / `release_branch_prefix`) is empty, the reader shall yield the empty-string neutral value — sending the consumer to its caller fallback with the same warning semantics as the git-flow empty-`develop_branch` case (t637 precedent).

### B.3 Behavior preservation (REQ-GWS-007..008)

- REQ-GWS-007 (Ubiquitous): The characterization suite shall pin, before any production change lands, the current behavior of `LoadGitFlowIntegrationConfig`, `IsGitFlow()`, and `LoadGitFlowDevelopBranch` across all failure paths (missing file, unparseable file, non-manual mode, non-git-flow workflow, no active profile, empty/whitespace develop_branch) and success paths.

- REQ-GWS-008 (Ubiquitous): The template default shall remain `github-flow` in all three mode profiles of `git-strategy.yaml.tmpl` and in `internal/config/defaults.go`; a `github-flow` project's observable behavior shall be unchanged by this SPEC (valid-non-git-flow → identical consumer behavior to today).

- REQ-GWS-009 (Ubiquitous): The `moai doctor` diagnostic surface shall consume the per-flow interpretation table: it shall report the configured flow's validity, the flow's standing-branch interpretation, and the resolved integration target (or the empty-key fallback), so the table carries at least one production consumer. The acquire/automerge consumers shall not be rewired to per-flow targets by this SPEC.

## C. Design decisions (plan-phase, owned by this SPEC)

### C.1 D1 — Rejection semantics (operator item 1)

Three-way disposition, not error-throwing:

| Value class | Reader disposition | Consumer behavior | Diagnosability |
|---|---|---|---|
| `git-flow` | flow = git-flow | develop_branch adopted (unchanged) | — |
| `github-flow` / `gitlab-flow` / `release-flow` | valid-non-git-flow | caller fallback (identical to today) | no warning — deliberate choice |
| anything else (typos, `trunk-based`, empty) | invalid | caller fallback + explicit warning naming the value and the 4 allowed entries | new `moai doctor` check |

Rationale: the reader runs outside the Loader lifecycle in consumers that must never hard-fail (t449 fail-open contract); "explicitly rejected" is therefore implemented as *machine-detectable invalidity* (structured disposition + diagnostic), not as load error. `trunk-based` being excluded means a user who sets it gets the invalid-value diagnostic — the deliberate exclusion is enforced by omission from the allowed set, and the diagnostic is where they discover it.

Doctor precedent: `internal/cli/doctor_worktree_base.go` (four-state check with distinct repair next-steps). The new check reports the invalid value, the allowed set, and leaves repair to the user (edit the YAML) — no auto-repair. Per D1(b), the doctor check is ALSO the production consumer of the interpretation table (REQ-GWS-009): for a valid flow it reports the flow's standing-branch interpretation and resolved integration target (e.g. gitlab-flow → environment value, or the empty-key fallback), giving scope item 2 a real in-SPEC consumer without rewiring acquire/automerge.

### C.2 D2 — Integration-target precedence vs `develop_branch` (operator item 2)

**Precedence rule: flow-scoped keys, never competing.** There is no precedence conflict to resolve because each flow reads exactly one target source:

```
github-flow  → "main"                       (fixed default, no key read)
git-flow     → develop_branch               (existing; empty → caller fallback + warning)
gitlab-flow  → environment                  (empty → caller fallback + warning)
release-flow → release_branch_prefix        (empty → caller fallback + warning)
```

`develop_branch` keeps its flow gate (REQ-GWS-005) — a github-flow project carrying a stale `develop_branch` is unaffected, exactly the t449 rationale generalized to all flows. The interpretation table lives in `internal/config` next to the reader so the doctor check (REQ-GWS-009) and any future consumer share one table and none re-derives it. Acquire/automerge continue to resolve targets exactly as today (REQ-GWS-008); adopting the table there is a follow-up card candidate, not this SPEC.

### C.3 D3 — Wizard workflow choices (operator item 4a): DEFERRED

The init/update wizard currently writes only `mode` and `provider` into git-strategy.yaml (measured: `internal/cli/wizard_config_test.go` — no workflow write path exists; the shipped template default `github-flow` survives the wizard untouched). Adding a workflow-picker question expands the interview surface but is not needed for reader correctness: users of gitlab-flow/release-flow edit the YAML directly today and continue to. **Deferred** with rationale: the UX surface should follow once per-flow behavior is differentiated; bundling it here would put an interview change inside a reader-validation SPEC. Recorded as a follow-up card candidate in the completion report.

### C.4 D4 — shipped_key_inventory.yaml (operator item 4b): INCLUDED

`git_strategy.{manual,personal,team}.workflow` entries (lines 500/557/623 of `internal/config/testdata/shipped_key_inventory.yaml`) carry `class: W / evidence: reader`. After this SPEC the evidence is updated to name the validated reader surface so the inventory stays honest about what reads the key. Small, mechanical, same-repo file — included in M4.

## D. Acceptance criteria

See `acceptance.md` §D AC matrix (AC-GWS-001..013, Given-When-Then).

## E. Non-functional constraints

- Coverage: ≥ 85% on `internal/config` package for the new/extended code paths; the t637 baseline for `loader_integration_branch.go` is already 100% and shall not regress.
- TDD: RED→GREEN per milestone; characterization tests (M1) land as their own commit BEFORE the M2 production commit (verification-claim-integrity §2.3 ordering attribution — baseline in its own, earlier commit).
- Fail-open: the reader never returns an error to a consumer; invalidity is a structured disposition.
- No template-default change; no wizard interview change (D3 deferred).

## F. Out of Scope

### Out of Scope — wizard interview surface
- No init/update wizard workflow-picker question (D3 deferred, see C.3).

### Out of Scope — per-flow automation behavior
- No new per-flow behavior beyond target-branch interpretation: gitlab-flow environment-branch creation, release-flow release-branch versioning, and flow-specific standing-branch ENFORCEMENT (e.g. refusing merges onto main under git-flow) are not built. This SPEC interprets the value; it does not automate the flows.

### Out of Scope — trunk-based support
- `trunk-based` is deliberately not in the allowed set (operator decision); no tolerance path, alias, or migration shim is provided.

### Out of Scope — schema changes
- No new YAML keys; `environment` and `release_branch_prefix` already exist in ModeProfile (types.go:108,130). No git-strategy.yaml schema version bump.

## G. Risks

- R1: A project deliberately carrying a non-set value (e.g. hand-written `trunk`) today falls silently into "not git-flow"; after M3 it gains a doctor warning. Mitigation: warning is diagnostic-only, consumer behavior identical — no workflow breaks.
- R2: `environment`/`release_branch_prefix` are pass-through fields with round-trip-only contracts today (types.go comment); wiring them as target sources promotes them to read keys. Mitigation: REQ-GWS-006 empty-key fallback keeps absence safe; characterization tests cover both. Additional hazard (plan-audit iter1 D6): the shipped default `environment` values ("local" for manual, "github" for personal/team, `defaults.go:763-788`) are environment LABELS, not branch names — a gitlab-flow adopter who leaves the default gets a non-branch string as the interpreted target. REQ-GWS-006 covers only the empty case, so this is a configuration-quality risk, not a load failure; the doctor check MAY warn on known non-branch defaults (implementation discretion, run phase).
- R3: Consumers other than the three measured sites may branch on `Workflow` raw values in future; the interpretation table being in `internal/config` (single source) mitigates re-derivation drift.

## H. Cross-references

- `plan.md` — milestones M1-M4 (TDD), pre-flight, constraints.
- `acceptance.md` — AC-GWS-001..012 matrix, characterization-first gate.
- `research.md` — in-tree verified evidence, t449/t637/t655 provenance, discarded-premise record.
- `progress.md` — §E plan-phase audit-ready signal.
