---
id: SPEC-V3R6-MEMORY-CONFIG-CLEANUP-001
title: "Memory subsystem config theater removal + Gap C/B decision records"
version: "0.1.0"
status: completed
created: 2026-06-03
updated: 2026-06-03
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/config"
lifecycle: spec-anchored
tags: "config, memory, cleanup, dead-code, honesty"
era: V3R6
tier: S
---

# SPEC-V3R6-MEMORY-CONFIG-CLEANUP-001 — Memory subsystem config theater removal

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-03 | manager-spec | Initial plan-phase authoring. Tier S. Encodes 3 ground-truth verdicts from the memory-system audit (prior memory: project_memory_system_audit_complete.md, commit 2a658e873) + adversarial re-scope: Gap A REMOVE (real defect), Gap B DEFER (refuted — fabricated requirement + already-documented intentional deferral), Gap C DEFER/DOCUMENT (real but internally consistent with Claude Code per-cwd model). |

## §A. Context and Motivation

A memory-system audit found three Go-code gaps in the memory subsystem that were too worktree-sensitive for ad-hoc edits and were deferred to a SPEC. An orchestrator-run ground-truth investigation with adversarial verification (3 parallel investigators + per-gap skeptical refutation) re-scoped the original "Tier M 3-part code plan": only **1 of the 3** gaps is a real fix-worthy code defect; the other two are resolved by a documentation/deferral decision. The user has confirmed all three decisions.

This SPEC is a **honesty-and-cleanup** effort, not a feature. Its WHY:

1. **Gap A (config theater)**: A 4-field config schema (`config.MemoryConfig`) is bound by `yaml.Unmarshal` then discarded — it has zero production consumers. Editing `workflow.yaml memory.*` has no runtime effect. This is misleading config surface ("config theater") and should be removed so the deployed schema is honest about what is actually wired.
2. **Gap B (fabricated requirement + latent checks)**: Recorded as an explicit evaluated-and-deferred decision so future maintainers do not re-discover the same non-issue and waste effort wiring latent checks or adding a fabricated byte-cap.
3. **Gap C (per-cwd memory divergence)**: Recorded + documented so future maintainers understand the per-cwd behavior is intentional and aligned with Claude Code's own per-cwd memory model — a git-root "fix" would cause a regression, not a fix.

This is a follow-up to the memory-system audit; it does NOT re-expand scope to the original 3-part code plan.

## §B. Ground-Truth Verdicts (the three gaps)

### B.1 Gap A — `config.MemoryConfig` is inert (config theater) → REMOVE

Verdict: **REAL** (high confidence, not refuted).

Evidence:
- The struct `MemoryConfig` (4 fields: `AuditEnabled`, `IndexLineCap`, `StaleAggregateThreshold`, `StalenessThresholdHours`) is defined in `internal/config/types.go`, held as field `Memory MemoryConfig` inside parent struct `WorkflowConfig`, defaulted in `internal/config/defaults.go`, and mirrors the YAML block in both the template and the local deployed `workflow.yaml`.
- The 4 struct fields have **zero production consumers**. They are bound by `yaml.Unmarshal` then discarded.
- The real, working memory-audit subsystem is driven entirely by **package-level constants** (`internal/config/defaults.go`: `DefaultMemoryStalenessHours=24`, `DefaultMemoryIndexLineCap=200`, `DefaultMemoryStaleAggregateThreshold=10`) plus the `MOAI_MEMORY_AUDIT` env var — NOT the struct. (The PostToolUse hook reads the env var; SessionStart and the audit engine use the constants.)
- The `internal/statusline` package field `data.Memory` is a different struct (statusline render data), NOT a consumer of `config.MemoryConfig`.

### B.2 Gap B — AuditIndex 25KB byte-cap + dead-wiring → DEFER (no code change)

Verdict: **REFUTED** (high confidence). Two independent grounds:

- (b1) The 25KB byte-cap is a **fabricated requirement** — it is the NATIVE Claude Code loader's limit (described in prose in the memory doctrine), NOT something any MoAI spec ever mandated. The governing memory-taxonomy SPEC (status: completed) specs only a 200-line cap. The PostToolUse hook explicitly does NOT observe the auto-memory `MEMORY.md` index — it only watches `.claude/agent-memory/`.
- (b2) The "dead-wiring" of `AuditIndex` / `AuditDuplicates` (in `internal/hook/memo/taxonomy/audit.go` — zero production callers; the memory-audit hook path calls only `AuditFile`) is an **already-documented intentional deferral**. The memory doctrine has an "Available checks, NOT yet wired into the PostToolUse hook" section listing `MEMORY_INDEX_OVERFLOW` + `MEMORY_DUPLICATE` with a user-facing warning not to rely on them firing. Library-level acceptance criteria of the original taxonomy SPEC pass.

Decision: **defer, no code change**. Record this as an explicit evaluated-and-deferred decision (traceability). Do NOT add a byte-cap. Do NOT wire the latent checks. Activation would be a future deliberate enhancement, not a defect fix.

### B.3 Gap C — `resolveMemoryDir` worktree divergence → DEFER / DOCUMENT (no logic change)

Verdict: **REAL** (high confidence) but the current behavior is internally consistent with Claude Code's own per-cwd memory model.

Evidence:
- `resolveMemoryDir` (in `internal/hook/session_end.go`) slugifies the raw session `CWD` (via `projectSlug`) with NO git-root resolution, so a git-worktree session resolves a different memory directory than the main repo.
- The existing test (`internal/hook/resolve_memory_dir_test.go`) ENCODES this divergence as the intended contract.
- Empirically, `~/.claude/projects/` contains per-worktree hash directories — Claude Code itself is per-cwd. A git-root "fix" would desync MoAI writes (main-root) from Claude Code's native auto-load (cwd), causing a regression. The investigated `CLAUDE_PROJECT_DIR` alternative was verified to NOT fix divergence on its own.

Decision: **defer / document only (no logic change)**:
1. Add an explanatory comment in `internal/hook/session_end.go` at `resolveMemoryDir` stating the per-cwd behavior is intentional and aligned with Claude Code's per-cwd memory model.
2. Improve the warn-message wording in `internal/hook/handoff/persist.go` (the `os.Stat` miss → "memory directory unavailable; skipping persistence") to hint at worktree divergence.
3. Add a short doctrine note documenting that SessionEnd handoff persistence is per-cwd by design and that L3 `--worktree` resume relies on Block 0 re-anchoring cwd.

## §C. Requirements (GEARS)

### REQ-MCC-001 — Remove inert MemoryConfig schema (Gap A)
The build system **shall not** retain the `MemoryConfig` struct, the `Memory MemoryConfig` field of `WorkflowConfig`, or the `Memory: MemoryConfig{...}` default block after this SPEC is implemented. (Ubiquitous — observable behavior: the symbol `MemoryConfig` is absent from non-test Go source.)

### REQ-MCC-002 — Atomic lockstep removal across all binding sites (Gap A)
**While** strict-mode YAML unmarshalling is in effect, the configuration loader **shall** continue to load `workflow.yaml` without error after the `memory:` block is removed. (State-driven — strict yaml.v3 fails loud on unknown keys, so the struct field and the YAML keys MUST be removed atomically. Removing only one side is a defect.)

### REQ-MCC-003 — Remove the memory block from both YAML copies + regenerate embedded (Gap A)
The template source `workflow.yaml`, the local deployed `workflow.yaml`, and the regenerated `embedded.go` **shall** all agree that no `memory:` block exists. (Ubiquitous — the dev project's own `config.Load()` reads the local copy under strict yaml; the embedded copy ships to users.)

### REQ-MCC-004 — Preserve the working const + env audit subsystem unchanged
The memory-taxonomy audit subsystem **shall** behave identically before and after this SPEC. The package-level constant block in `internal/config/defaults.go` (`DefaultMemoryStalenessHours`, `DefaultMemoryIndexLineCap`, `DefaultMemoryStaleAggregateThreshold`) and the `MOAI_MEMORY_AUDIT` env-var path **shall** remain unmodified. (Ubiquitous — these are the real, wired functionality; only the dead struct is removed.)

### REQ-MCC-005 — Update mirror/assertion tests to match the removed schema (Gap A)
The configuration test suite **shall** pass after the `Memory.*` field reflection assertions, default-value assertions, and nested-config assertions referencing the removed struct are updated or removed. (Ubiquitous — `internal/config/{workflow_nested_test,types_test,defaults_test}.go` reference `Memory.*`.)

### REQ-MCC-006 — Preserve template neutrality and mirror-drift invariants
**Where** template neutrality and mirror-drift CI guards are active, this SPEC's edits **shall** keep both guards green: the embedded copy, the template `workflow.yaml`, and the local `workflow.yaml` agree (memory block absent everywhere), and no internal SPEC-ID / date / commit-SHA leaks into template-bound files. (Capability gate — per CLAUDE.local.md §15/§25 + `internal/template/rule_template_mirror_test.go`.)

### REQ-MCC-007 — Document Gap C per-cwd behavior without logic change
The codebase **shall** carry an explanatory comment at `resolveMemoryDir` stating the per-cwd memory resolution is intentional and aligned with Claude Code's per-cwd model; the `persist.go` warn message **shall** hint at worktree divergence; and a doctrine note **shall** document the per-cwd-by-design behavior. The `resolveMemoryDir` and `projectSlug` resolution logic **shall not** change. (Ubiquitous + Unwanted — no behavioral change; `resolve_memory_dir_test.go` assertions remain unchanged.)

### REQ-MCC-008 — Record the Gap B deferral decision (no code change)
This SPEC **shall** record Gap B (the fabricated 25KB byte-cap requirement and the latent `AuditIndex`/`AuditDuplicates` dead-wiring) as an explicit evaluated-and-deferred decision, referencing the existing canonical disclosure in the memory doctrine. **When** a future maintainer revisits the latent checks, the recorded decision **shall** make the prior evaluation discoverable. (Event-detected — no code change to `audit.go` / the hook audit path.)

### REQ-MCC-009 — No new byte-cap, no latent-check wiring (Gap B unwanted behavior)
This SPEC **shall not** add a byte-cap to any audit check and **shall not** wire `AuditIndex` / `AuditDuplicates` into the PostToolUse hook. (Unwanted behavior — explicitly out of scope; activation is a future deliberate enhancement.)

## §D. @MX Targets

- `config.Load()` path: removing the `Memory MemoryConfig` field touches the strict-yaml unmarshal target struct. `config.Load` has high fan_in (every CLI subcommand reads from it) — the run-phase agent SHOULD verify no `@MX:ANCHOR` invariant on `WorkflowConfig` is violated and add `@MX:NOTE` at the removal site explaining the const+env subsystem is the real wiring (Gap A rationale). No new `@MX:WARN` expected (removal reduces surface).

## §E. Exclusions (What NOT to Build)

- **No byte-cap.** Do NOT add a 25KB (or any) byte-cap to any audit check (Gap B refuted — fabricated requirement). [EXCL-MCC-001]
- **No latent-check wiring.** Do NOT wire `AuditIndex` / `AuditDuplicates` into the PostToolUse hook or any production caller (Gap B intentional deferral). [EXCL-MCC-002]
- **No `resolveMemoryDir` / `projectSlug` logic change.** Gap C is document-only; the resolution behavior and its encoding test stay unchanged. Do NOT introduce git-root resolution (would cause a regression vs Claude Code's per-cwd model). [EXCL-MCC-003]
- **No change to the const+env audit subsystem.** Do NOT remove or modify `DefaultMemoryStalenessHours` / `DefaultMemoryIndexLineCap` / `DefaultMemoryStaleAggregateThreshold` or the `MOAI_MEMORY_AUDIT` env path. Only the dead `MemoryConfig` struct is removed. [EXCL-MCC-004]
- **No new MemoryConfig consumer.** Do NOT "fix" the inert struct by wiring it to a consumer; the decision is REMOVE, not adopt. [EXCL-MCC-005]
- **No re-expansion to the original Tier M 3-part code plan.** Only Gap A is a code change; Gap B/C are decision-record + doc-only. [EXCL-MCC-006]
- **No `CLAUDE_PROJECT_DIR`-based resolution change** for Gap C (verified not to fix divergence on its own). [EXCL-MCC-007]

## §F. Cross-References

- Prior audit memory: `project_memory_system_audit_complete.md` (commit 2a658e873).
- Memory doctrine (Gap B canonical disclosure — "Available checks, NOT yet wired"): `.claude/rules/moai/workflow/moai-memory.md`.
- Frontmatter schema SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md`.
- Config module conventions: `internal/config/CLAUDE.md` (strict yaml, struct↔YAML symmetry guards).
- Template neutrality / mirror-drift: CLAUDE.local.md §15/§25, `internal/template/CLAUDE.md`, `internal/template/rule_template_mirror_test.go`.
- Hook module conventions: `internal/hook/CLAUDE.md`.
