---
id: SPEC-DEAD-CONFIG-001
title: "Remove dead runtime.yaml config section and CI-guard allowlist rows"
version: "1.0.0"
status: completed
created: 2026-07-02
updated: 2026-07-02
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/config"
lifecycle: spec-anchored
tier: S
era: V3R6
tags: "config, dead-code, cleanup, ci-guard, yaml"
related_specs: [SPEC-V3R3-ARCH-007, SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001, SPEC-CI-MULTI-LLM-001]
---

# SPEC-DEAD-CONFIG-001 — Remove dead runtime.yaml config section and CI-guard allowlist rows

## HISTORY

- 2026-07-02 — v0.2.0 — draft — manager-spec — **Scope narrowed to `runtime.yaml` ONLY.** After
  plan-auditor PASS-WITH-DEBT (0.76), the "github-actions.yaml is dead" premise did NOT survive
  independent verification: it is the live hand-maintained config of `SPEC-CI-MULTI-LLM-001`
  (status: implemented, verified via frontmatter), referenced by docs-site across 4 locales, and already
  scheduled for authoritative removal via `DeprecatedPaths` at v3.0.0. github-actions.yaml is removed from
  scope entirely; its allowlist row is retained. REQs/ACs renumbered sequentially. The false
  "grep found none — Verified" risk claim (a Go-only grep that never searched the docs surface) is removed.
- 2026-07-02 — v0.1.0 — draft — manager-spec — Initial plan-phase authoring (github-actions.yaml +
  runtime.yaml). Superseded by v0.2.0 scope narrowing.

## §A. Context and Motivation

A prior configuration audit ("미사용 설정 제거" deliverable) sought to remove genuinely-dead
`.moai/config/sections/*.yaml` files and dead rows in the `internal/config/audit_loader_completeness_test.go`
allowlist. Independent plan-audit narrowed the verified-solid half to **`runtime.yaml` only**.

This SPEC removes ONLY the on-disk `runtime.yaml` (both trees) and three dead allowlist rows. It is a
minimal dead-code removal, not a refactor. No production Go code changes.

### Verified findings (plan-phase, confirmed against the tree — not assumed)

| Finding | Evidence |
|---------|----------|
| `runtime.yaml` is dead on-disk | Local (`.moai/config/sections/runtime.yaml`) AND template (`internal/template/templates/.moai/config/sections/runtime.yaml`) both exist. `LoadRuntime` (`internal/runtime/config.go:81`) is acknowledged in `acknowledgedDedicatedLoaders`, but is NEVER called in production — the only `LoadRuntime` callers are `internal/runtime/budget_test.go`, and `internal/runtime` has **zero production importers**. Production `NewTracker` consumers pass `DefaultRuntimeConfig()`. Content is stale (references archived agents such as `expert-backend`). |
| `LoadRuntime` removal is SAFE (no test fix) | `TestLoadRuntimeFromFile` (`budget_test.go:368`) writes its OWN inline YAML to a `t.TempDir()` path and reads that; `TestLoadRuntimeMissingFile` uses a nonexistent path. Neither reads the real `runtime.yaml`. Removing the on-disk YAML from both trees breaks NO test. |
| `gate` / `memo` allowlist rows are dead | `gate.yaml` and `memo.yaml` are ABSENT from BOTH the local and template sections directories. `TestAuditLoaderCompleteness` iterates the TEMPLATE directory, so these rows are never consulted. |
| CI guard reads the TEMPLATE dir | `audit_loader_completeness_test.go:67` builds `templateSectionsDir` and iterates it. `runtime` is currently LIVE (template `runtime.yaml` present) and becomes dead after the template file is removed. |
| `DeprecatedPaths` manifest is a DIFFERENT mechanism | `internal/defs/dirs.go` `DeprecatedPaths` + `dirs_test.go` `TestDeprecatedPathsCategoryBExpectedEntries` drive user-project v2→v3 migration cleanup (SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 §A.4), NOT the loader-completeness guard. Pinned by test; MUST NOT be edited by this SPEC. |
| `github-actions.yaml` is NOT dead — excluded | Independent verification REVERSED the earlier "dead" premise. It is the live config of `SPEC-CI-MULTI-LLM-001` (`status: implemented`). See §D "Out of Scope — github-actions.yaml" for the full rationale. |

## §B. Baseline (green, this tree — plan-phase observation)

- `go test ./internal/config/ -run TestAuditLoaderCompleteness` → `ok`
- `go test ./internal/runtime/ -run 'TestLoadRuntime'` → `PASS` (both LoadRuntime tests)
- `go test ./internal/defs/ -run TestDeprecatedPaths` → `ok`

## §C. Requirements (GEARS)

### REQ-DC-001 — Remove dead on-disk `runtime.yaml` (Unwanted behavior)

The moai-adk repository **shall not** retain `.moai/config/sections/runtime.yaml` in either the local tree
or the template tree (`internal/template/templates/.moai/config/sections/runtime.yaml`), because production
code resolves runtime defaults via `DefaultRuntimeConfig()` and never invokes `LoadRuntime` on the on-disk
file.

### REQ-DC-002 — Preserve the runtime Go surface (Ubiquitous)

The runtime Go surface — the `RuntimeConfig` type, `DefaultRuntimeConfig()`, and `LoadRuntime` in
`internal/runtime/config.go` — **shall** remain unchanged. Only the on-disk YAML is removed; the Go API
(still exercised by `budget_test.go` self-contained fixtures) stays intact.

### REQ-DC-003 — Remove dead `acknowledgedUnloadedSections` rows; retain `github-actions` (Unwanted behavior)

The `acknowledgedUnloadedSections` allowlist in `internal/config/audit_loader_completeness_test.go`
**shall not** retain the `gate` or `memo` rows, each of which references a YAML file absent from BOTH trees.
The `github-actions` row **shall** be retained — `github-actions.yaml` is a live, out-of-scope file (see
§D), and its row remains a valid out-of-scope acknowledgement.

### REQ-DC-004 — Remove the dead `runtime` dedicated-loader row (Event-driven)

**When** `internal/template/templates/.moai/config/sections/runtime.yaml` is removed, the
`acknowledgedDedicatedLoaders` allowlist **shall not** retain the `runtime` row, so that no dead allowlist
entry remains after the file deletion.

### REQ-DC-005 — Order the runtime row removal after the file deletion (State-driven)

**While** the template `runtime.yaml` still exists on disk, the `runtime` allowlist row **shall** be
retained. The row removal **shall** be sequenced after the file deletion (Milestone ordering M1 → M2) so
that `TestAuditLoaderCompleteness` never observes an uncovered `runtime` section in any intermediate commit.

### REQ-DC-006 — Regenerate the embedded binary (Where — capability gate)

**Where** the template tree changes (template `runtime.yaml` removed), the embedded template FS **shall**
be regenerated via `make build` so the compiled binary reflects the deletion.

### REQ-DC-007 — Leave the migration manifest intact (Ubiquitous)

The `internal/defs` `DeprecatedPaths` migration manifest and its `dirs_test.go` pins **shall** remain
unchanged. They serve user-project v2→v3 cleanup — a distinct mechanism from the loader-completeness
CI-guard allowlist — and are out of scope for this SPEC.

## §D. Exclusions (out of scope)

This SPEC is deliberately narrow. The following are explicitly out of scope. Each item below is out of
scope for the reason stated; do NOT remove or edit these.

### Out of Scope — github-actions.yaml

- `.moai/config/sections/github-actions.yaml` is **NOT dead** and is removed from this SPEC's scope
  entirely. Independent verification reversed the earlier "dead" premise:
  - It is the live hand-maintained config of `SPEC-CI-MULTI-LLM-001` (`status: implemented`, verified via
    the SPEC's frontmatter), which references it across its spec/plan/acceptance/tasks artifacts.
  - `.moai/research/config-audit-2026-05-22.md` §2.2 explicitly REVERSED its own v1 "dead" verdict with a
    v2 correction ("scaffolding for SPEC-CI-MULTI-LLM-001").
  - It is referenced by docs-site across 4 locales (`docs-site/content/{en,ja,zh,ko}/guides/multi-llm-ci.md`,
    one reference each — verified by grep).
  - Its removal is already scheduled authoritatively by `DeprecatedPaths` (`internal/defs/dirs.go:245`,
    Category B, `RemovalSchedule: v3.0.0`) — so it self-removes at the v3.0.0 release without this SPEC.
- Because the file stays, its `github-actions` row in `acknowledgedUnloadedSections` is retained (REQ-DC-003).
- Deliberately excluded to avoid a cross-reference reconciliation burden (docs-site ×4 + the owning SPEC).

### Out of Scope — parked / governance config YAMLs

- `constitution.yaml`, `research.yaml`, `context.yaml` — Go `cfg.Constitution` / `cfg.Research` /
  `cfg.Context` have zero consumers, but these are conceptually parked/governance artifacts that may be
  agent-consumed or planned. Deferred — needs separate confirmation before any removal.
- `sunset.yaml` — intentional DORMANT scaffolding (REQ-MIG003-006). Do NOT touch.
- `design.yaml`, `feedback.yaml`, `interview.yaml` — agent/orchestrator-consumed via skills/rules that
  read the YAML directly even though Go ignores them. Do NOT touch.

### Out of Scope — dead Go API surface (separate concern)

- The 4 MIG-003 dead public loaders (`LoadConstitutionConfig`, `LoadContextConfig`,
  `LoadInterviewConfig`, `LoadDesignConfig`) — zero production callers, but this is dead Go API surface,
  a different concern from dead YAML files. Possible follow-up SPEC.
- The `internal/runtime` package having zero production importers (dead package) — a larger structural
  concern than the on-disk YAML. This SPEC keeps the Go code per REQ-DC-002; package retirement is a
  possible follow-up SPEC.
- The `lsp` row in `acknowledgedUnloadedSections` — also effectively dead (the template ships only
  `lsp.yaml.tmpl`, which the audit test's `.yaml.tmpl` filter excludes). Not in this SPEC's scope;
  possible follow-up.

### Out of Scope — the migration manifest

- `internal/defs/dirs.go` `DeprecatedPaths` and `internal/defs/dirs_test.go` — the user-project v2→v3
  cleanup manifest. It legitimately lists `github-actions.yaml`, `gate.yaml`, `memo.yaml` (and others) as
  files to remove FROM USER PROJECTS, and is pinned by `dirs_test.go` to CLEAN-REINSTALL-001 §A.4.
  Editing it would break that test and regress user-project cleanup. Leave intact (REQ-DC-007).

## §E. Cross-references

- `internal/config/audit_loader_completeness_test.go` — the loader-completeness CI guard (the allowlist edited here).
- `internal/runtime/config.go` — the runtime Go surface preserved by REQ-DC-002.
- `internal/runtime/budget_test.go` — the self-contained LoadRuntime tests confirming safe removal.
- `internal/defs/dirs.go` + `dirs_test.go` — the migration manifest left intact (REQ-DC-007).
- `SPEC-CI-MULTI-LLM-001` — owner of `github-actions.yaml` (excluded, §D).
- CLAUDE.local.md §2 (Template-First rule) — governs the template `runtime.yaml` removal + `make build`.
