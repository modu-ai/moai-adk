---
id: SPEC-CONFIG-AUDIT-REPAIR-001
title: ".moai/config full-tree audit repair — all HIGH/MEDIUM/LOW findings in one Tier L effort"
version: "0.2.0"
status: completed
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: P1
phase: "v3.0.x target"
module: ".claude/rules + .claude/skills + internal/config + internal/hook + internal/template/templates"
lifecycle: spec-anchored
tags: "config, audit-repair, docs-drift, astgrep-gate-restore, template-parity, test-repair"
tier: L
related_specs: [SPEC-ASTGREP-DOGFOOD-CLEANUP-001]
---

# SPEC-CONFIG-AUDIT-REPAIR-001 — .moai/config 전수 감사 수리

## HISTORY

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 0.1.0 | 2026-07-25 | manager-spec | Initial draft — repairs ALL findings of `.moai/reports/moai-config-audit-20260725.md` (HIGH 4 / MEDIUM 7 incl. M7 / LOW 8; all CONFIRMED, Gaps 0). User-approved scope: 전체 일괄 (Tier L). |
| 0.2.0 | 2026-07-25 | manager-spec | Implementation Kickoff decisions recorded: DECISION-1 REVERSED to gate RESTORE (loader + loadable ruleset, default OFF) per user directive; DECISION-2/3 dev-only as recommended. Plan-auditor SHOULD-FIX folded: enable-path touch-point inventory (plan.md §F M-2), AC-CAR-002/005/007(c) tightened, LOW counting convention pinned. Added REQ-CAR-019..021; [NEEDS CLARIFICATION] markers removed. LOW counting convention: the audit's "LOW 8" = 6 defect bullets, with the registry/comment-drift bullet counted as 2 items (test comment + registry mislabel) and F3 promoted out of LOW to MEDIUM M7 by second-pass V2. |

## §A Context

The completed `.moai/config` full-tree audit (`.moai/reports/moai-config-audit-20260725.md`, 2026-07-25) confirmed 19 findings across four lenses (loader coverage / template drift / doc gap / asset consumers), all with cited evidence and zero verification gaps. This SPEC repairs every finding in a single Tier L effort. Evidence citations below reference the audit report table rows (H1-H4, M1-M7, LOW bullets); per `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3, each claimed defect in this SPEC carries the audit's evidence citation — no defect is asserted beyond what the audit's dedicated verification (grep / ls / `go test -run` runs recorded in the report) observed.

Tier L rationale: 6 milestone groups spanning rule/doc files (10+), skill files, Go test code (`internal/config`), a design decision with two viable options (H3), a distribution-status codification decision (H4+M3), a cross-SPEC dependency (M1 → DB-removal track), and mandatory template-mirror parity (`make build`) with CI neutrality guards. File touch count exceeds 20 across doc + template + Go surfaces.

## §B Requirements (GEARS)

### Group 1 — Mechanical doc/rule corrections (H1, H2, M2, M4, M6, LOW)

- **REQ-CAR-001** (H1 — audit row H1, evidence `worktree-integration.md:93,98,178-179,193,239,249`, `agent-authoring.md:168`, `model-policy.md:166`; `grep '^team:' workflow.yaml` → no match): The repair shall remove or reword every reference to the retired `team.role_profiles` config (including the `ORC_WORKTREE_REQUIRED` lint clause tied to it) in `.claude/rules/moai/workflow/worktree-integration.md`, `.claude/rules/moai/development/agent-authoring.md`, and `.claude/rules/moai/development/model-policy.md`, aligning them with the Agent Teams retirement (CLAUDE.md §4/§15).
- **REQ-CAR-002** (H2 — audit rows H2+V1, evidence `ls .moai/config/config.yaml` → No such file on both local and template; `resolver.go:349` optional-read only, no writer): The repair shall replace every reference that presents `.moai/config/config.yaml` as an existing file with sections-SSOT wording (the `.moai/config/sections/*.yaml` family), at these sites: `.claude/skills/moai/SKILL.md:19,335`, `.claude/agents/moai/manager-spec.md:102`, `.claude/skills/moai/workflows/run/context-loading.md:130`, `.claude/skills/moai/workflows/sync/quality-gates-context.md:108`, `.claude/rules/moai/core/settings-management.md`, and `CLAUDE.local.md` §5/§9.
- **REQ-CAR-003** (H2 abort gate — evidence `harness.md:129` uses config.yaml existence as abort gate; per V1 this gate always fails): **When** the harness workflow evaluates its project-initialized abort gate, the gate shall check `.moai/config/sections/` directory existence instead of the nonexistent `.moai/config/config.yaml` file.
- **REQ-CAR-004** (M2 — evidence `delivery.md:17,215,417` read `github.git_workflow`; actual key is `spec_git_workflow` per `system.yaml:44`; `delivery.md:27` already correct): The repair shall correct all 3 `git_workflow` key references in `.claude/skills/moai/workflows/sync/delivery.md` to `spec_git_workflow`.
- **REQ-CAR-005** (M4 — evidence `MCP_OAUTH_SETUP.md:163` references nonexistent `.moai/config/mcp-servers.yaml`): The repair shall replace the `mcp-servers.yaml` reference with the actual surfaces (`.mcp.json` / `mcp-matrix.yaml`).
- **REQ-CAR-006** (M6 — evidence `design/constitution.md:60` cites nonexistent `design.yaml adaptation.phase_weights`; 5 docs incl. `run.md:74`, `moai.md:152` cite `quality.yaml development_mode` bare instead of `constitution.development_mode`): The repair shall correct the phase_weights reference and rewrite all 5 bare `development_mode` path citations to `constitution.development_mode`.
- **REQ-CAR-007** (LOW batch — evidence audit LOW bullets): The repair shall fix each of: (a) `audit_loader_completeness_test.go:16` stale comment "db not yet runtime-consumed" (false — hook consumes `migration_patterns` via `cli/hook.go:513-520`); (b) `audit_registry.go` mislabeling constitution as "no loader" (loader_constitution.go exists); (c) `dynamic-workflows.md:91` retired `haiku` in the workflow_agents model enum; (d) `quality-gates-context.md:197` describing `db.auto_sync` (a map) as a scalar; (e) local `quality.yaml` dead leaf `coverage_threshold: 0` removal (dead per `schema_sections.go:205`; template already clean).
- **REQ-CAR-008** (Template-First — CLAUDE.local.md §2): **Where** a corrected file under `.claude/` has a mirror in `internal/template/templates/`, the repair shall apply the fix to the template source first and to the local copy, and shall run `make build` so the embedded FS is regenerated; template content shall stay §25-neutral (no SPEC IDs, REQ tokens, audit citations, internal dates, commit SHAs).

### Group 2 — M7 test repair + orphan reconciliation

- **REQ-CAR-009** (M7 — audit V2, evidence `go test -run TestAuditParity -v` → `--- SKIP` because cwd is the package dir and relative path `.moai/config/sections` is not found): The repair shall make `TestAuditParity` in `internal/config/audit_test.go` resolve the repo root via `runtime.Caller` (same pattern as `audit_loader_completeness_test.go`) so the test executes (PASS or FAIL, never SKIP) under plain `go test ./internal/config/...`.
- **REQ-CAR-010** (M7 follow-on): **When** the repaired `TestAuditParity` exposes previously-unguarded orphan sections (mcp-matrix, tool-policy, cache, db, feedback, observability, report), the repair shall reconcile each by registering it in `audit_registry.go` (registry entry or documented exception) such that the full test suite is green with zero skips in `internal/config`.

### Group 3 — H3+M5 ast-grep gate RESTORE (DECISION-1 RESOLVED: restore, default OFF)

- **REQ-CAR-011** (H3 — evidence `defaults.go:286-296` never sets `AstGrepGate.Enabled`, `loader.go:47-98` has no gate entry, guard at `internal/hook/quality/gate.go:281`, `pre_tool.go:566-575` unreachable): The repair shall add a config loader path for the existing `gate` section (GateConfig / `gate` registry key already present in `audit_registry.go` as defaults-only) such that `AstGrepGate.Enabled` is configurable from a `gate.yaml` section file, making the guard at `gate.go:281` and the mapping at `pre_tool.go:566-575` reachable.
- **REQ-CAR-019** (default posture): **Where** no `gate.yaml` (or no `enabled: true`) is present, the ast-grep pre-tool gate shall remain OFF (opt-in enable); deployed users shall observe no behavior change from this SPEC unless they explicitly enable the gate.
- **REQ-CAR-020** (M5 ruleset loadability — evidence local `sgconfig.yml:24` phantom `utils` ruleDir; root `go-hardcoding.yml` not loaded in config-mode per `scanner.go:250-281`): **While** the gate is enabled, the scan path shall load an actually-existing ruleset: the phantom `utils` ruleDir line shall be removed from local `sgconfig.yml`, and the root `go-hardcoding.yml` (or the template's curated go/security set) shall be loaded in config-mode. **When** the `sg` binary is absent, the enabled gate shall degrade gracefully (skip with a notice, never a hard failure).
- **REQ-CAR-021** (legacy path disposition): The repair shall dispose of the V1 `RunAstGrepGate` path per design.md DECISION-1 record (keep V2 as the sole gate entry; delete or fold V1 if caller-less) so that no dead sibling branch remains after the restore.
- **REQ-CAR-012** (cross-SPEC boundary, revised): The repair shall absorb from backlog SPEC-ASTGREP-DOGFOOD-CLEANUP-001 exactly the items REQ-CAR-020 names (sgconfig `utils` line + config-mode root-rule loading) and shall record in plan.md what REMAINS in that backlog: 16-language ruleset productization, empty language stubs (10/17), message-language unification, and SPEC-ID stripping for distribution.

### Group 4 — H4+M3 distribution-status codification

- **REQ-CAR-013** (H4 — evidence `toolpolicy/loader.go:19` hard error when file absent; 47KB `tool-policy.yaml` not in template sections; user-facing docs 0; dev-only intent recorded only in `audit_loader_completeness_test.go:39` comment): The repair shall codify `tool-policy.yaml` distribution status per design.md DECISION-2 — either declare dev-only (register in CLAUDE.local.md §2 local-only list AND replace the CLI hard error with a graceful "not distributed / dev-only" message in `moai tool-policy list|build`) or neutralize+distribute via the template.
- **REQ-CAR-014** (M3 — evidence shipped template skill `project/doc-generation.md` references `mcp-matrix.yaml` which the template does not distribute): The repair shall codify `mcp-matrix.yaml` status per design.md DECISION-3 — either remove/reword the dangling template-skill reference (dev-only declaration) or neutralize+distribute the yaml.

### Group 5 — M1 db.yaml dead keys (cross-SPEC)

- **REQ-CAR-015** (M1 — evidence `db.yaml:15-23` `auto_sync.{debounce_seconds, require_user_approval, excluded_patterns}` have zero Go consumers, cross-confirmed by `.moai/reports/db-scaffold-sync-logic-20260725.md`): **While** the parallel-approved DB subsystem removal track (memory: project_db_retire_removal_approved.md) is active, this SPEC shall NOT implement the 3 dead keys and shall mark M1 as resolved-by-dependency on that track; **When** the removal track is confirmed stalled or abandoned at sync time, the fallback shall delete the 3 dead keys from `db.yaml` (local + template).

### Group 6 — Orphan-section documentation

- **REQ-CAR-016** (LOW orphans — evidence audit LOW bullet: 10 sections with zero doc references; priority tool-policy, lsp): The repair shall add SSOT documentation references for at least `tool-policy.yaml` (per DECISION-2 outcome) and `lsp.yaml` (named as the LSP-gate SSOT where CLAUDE.md §6 discusses LSP gates), and shall list the remaining orphan sections (security, observability, report, sunset, archive, cache, feedback, project) as acknowledged-orphans in one documented location.

### Group 7 — Quality gates

- **REQ-CAR-017**: **When** the run phase begins, the implementation shall capture an LSP/lint baseline (`golangci-lint run`, `go vet ./...`) before any edit, and the sync-phase state shall show zero new lint errors relative to that baseline.
- **REQ-CAR-018**: The repair shall keep the template-neutrality CI guards green (`.github/workflows/template-neutrality-check.yaml`, `internal_content_leak_test.go`) for every template-mirror edit.

## Out of Scope (exclusions)

### Out of Scope — DB subsystem implementation or removal
- Implementing `auto_sync` dead keys, or executing the DB subsystem removal itself — owned by the parallel DB-removal track (see REQ-CAR-015; this SPEC only carries the dependency note + fallback).

### Out of Scope — ast-grep ruleset productization
- 16-language ruleset distribution, the 10/17 empty language stubs, ruleset message-language unification, and distribution SPEC-ID stripping — these remain owned by backlog SPEC-ASTGREP-DOGFOOD-CLEANUP-001. This SPEC absorbs ONLY the two enable-path blockers (sgconfig phantom `utils` line, config-mode root-rule loading) per REQ-CAR-012/020.

### Out of Scope — new config features
- Any new config section, schema field, or loader capability beyond what a chosen H3/H4 design option strictly requires; healthy/intentional items the audit cleared (sunset DORMANT, evaluator-profiles, gate/runtime defaults-only structs) are untouched.

### Out of Scope — other SPEC directories and unrelated code
- No modification to any other `.moai/specs/` directory; no behavior changes outside the audited finding sites.

## §H Risk Register

| Risk | Impact | Mitigation |
|------|--------|------------|
| Template mirror parity miss (fix local only) | Distributed users keep the defect; CI may not catch doc-only drift | REQ-CAR-008 template-first order + AC parity greps on both trees |
| Template neutrality violation (SPEC/REQ tokens leaking into templates) | `template-neutrality-check.yaml` / `internal_content_leak_test.go` CI FAIL | Neutral rewording in templates; REQ-CAR-018 gate |
| Cross-SPEC race with DB-removal track (shared `db.yaml`, `cli/hook.go`) | Duplicate/conflicting edits | REQ-CAR-015 dependency-not-implement scoping; pre-spawn sync check |
| Cross-branch race with active TUX branch (feat/SPEC-CLI-TUX-INIT-UPDATE-001) | Merge conflicts on shared docs | Rebase before run-phase; pathspec-scoped commits |
| REQ-CAR-010 orphan reconciliation may reveal genuinely-missing loaders | Scope growth | Register as documented exceptions, not new loaders (except the `gate` loader mandated by REQ-CAR-011) |
| Gate restore ships a surprising new enforcement to deployed users | User-facing regression | REQ-CAR-019 default-OFF invariant + characterization test that default config yields Enabled=false |
| Enabled gate runs against a broken ruleset (M5) | Gate fires with full-scan failure | REQ-CAR-020 ruleset loadability is a hard precondition of the enable path; graceful degrade when `sg` absent |
