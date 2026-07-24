---
id: SPEC-CONFIG-AUDIT-REPAIR-001
title: "Implementation plan — config audit repair"
version: "0.2.0"
status: completed
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: P1
phase: "v3.0.x target"
module: ".claude/rules + .claude/skills + internal/config + internal/template/templates"
lifecycle: spec-anchored
tags: "config, audit-repair, plan"
tier: L
---

# Plan — SPEC-CONFIG-AUDIT-REPAIR-001

Milestones are ordered by decision-reversibility: the two open design decisions (hardest to reverse, need human approval) lead; mechanical doc sweeps trail.

## §A Context

Source of truth: `.moai/reports/moai-config-audit-20260725.md` (HIGH 4 / MEDIUM 7 / LOW 8, all CONFIRMED, Gaps 0). User approved 전체 일괄 repair. Working repo: /Users/goos/MoAI/moai-adk-go. Current branch at authoring time: feat/SPEC-CLI-TUX-INIT-UPDATE-001 (run phase must branch off a fresh baseline — see §C).

## §B Known Issues / Constraints

- Parallel DB-removal track approved (project_db_retire_removal_approved.md) — M1 is dependency-scoped, not implemented here (REQ-CAR-015).
- SPEC-ASTGREP-DOGFOOD-CLEANUP-001 backlog owns the local sgconfig repair (REQ-CAR-012 boundary).
- Template-First rule: every `.claude/` doc with a `internal/template/templates/` mirror must be fixed in BOTH trees (template first) + `make build`. Template content must stay §25-neutral.
- CI guards: template-neutrality-check.yaml, internal_content_leak_test.go, agents_frontmatter tests.

## §C Pre-flight

1. `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` — resolve divergence before spawn.
2. Capture baselines: `golangci-lint run > baseline-lint.log`, `go vet ./...`, `go test ./internal/config/... -v | grep -E 'SKIP|FAIL'` (expect TestAuditParity SKIP pre-fix — the M7 repro).
3. For each doc-fix target, verify a template mirror exists: `test -f internal/template/templates/<path>` — build the two-tree edit list before editing.

## §D Constraints

- No commits by manager-spec (plan-phase artifacts committed by orchestrator flow); run-phase commits pathspec-scoped.
- No new config features beyond the resolved design decisions (design.md — all 3 RESOLVED at Implementation Kickoff 2026-07-25: D1 gate RESTORE default-OFF, D2 tool-policy dev-only + graceful CLI error, D3 mcp-matrix reword/dev-only).
- Cross-SPEC absorption boundary with SPEC-ASTGREP-DOGFOOD-CLEANUP-001: this SPEC absorbs ONLY the two enable-path blockers (local `sgconfig.yml:24` phantom `utils` ruleDir removal; config-mode loading of root `go-hardcoding.yml`/curated set). REMAINING in that backlog: 16-language ruleset productization, 10/17 empty language stubs, ruleset message-language unification, distribution SPEC-ID stripping.

## §E Self-Verification

Run-phase completion requires: AC matrix in acceptance.md fully evaluated with executed commands; `go test ./...` green, zero SKIP in internal/config; lint delta vs baseline = 0; template/local parity greps clean; neutrality guards green.

## §F Milestones

### M-1 — Design decisions — RESOLVED (Implementation Kickoff 2026-07-25)
- DECISION-1: H3 gate RESTORE (user directive — "재설계를 통해 ast-grep을 제대로 사용 가능하도록 게이트 복구"), default OFF. DECISION-2: tool-policy dev-only + graceful CLI. DECISION-3: mcp-matrix reword, dev-only. Recorded in design.md; no open markers remain.

### M-2 — ast-grep gate restore (LARGEST implementation milestone; REQ-CAR-011/019/020/021)
Enable-path touch-point inventory (plan-auditor SHOULD-FIX 1, inverted from removal to restore):
- `internal/config/types.go:569-570` — `AstGrepGate` / GateConfig field surface (extend if schema needs enable/ruleset keys).
- `internal/config/defaults.go` (~:286-296 vicinity) — gate defaults: explicit `Enabled: false` default (REQ-CAR-019 invariant).
- `internal/config/loader.go:47-98` — add `gate` section entry so `gate.yaml` loads (registry key `gate` already exists in `audit_registry.go` with defaults-only; loader completes the pair).
- `internal/hook/quality/gate.go:281` — guard branch becomes reachable when enabled; unit test proves both branches.
- `internal/hook/.../pre_tool.go:566-575` — enabled-path mapping exercised by test.
- V1 `RunAstGrepGate` disposition per design.md: keep V2 as sole entry; delete/fold V1 (REQ-CAR-021).
- Ruleset loadability (REQ-CAR-020): remove `sgconfig.yml:24` phantom `utils` line; ensure config-mode loads root `go-hardcoding.yml`/curated set (`scanner.go:250-281`); graceful degrade when `sg` absent (CI has no `sg` — degrade path is the CI default).
- Full suite + characterization: default config behavior byte-identical (gate OFF).

### M-3 — H4+M3 distribution codification (per resolved D2/D3)
- tool-policy: dev-only route = CLAUDE.local.md §2 entry + graceful error in `toolpolicy/loader.go:19` call path (`moai tool-policy` prints a "dev-only, not distributed" message instead of hard error). mcp-matrix: reword template `project/doc-generation.md` (both trees) or distribute.

### M-4 — M7 test repair + orphan reconciliation
- `audit_test.go` TestAuditParity: `runtime.Caller` repo-root resolution (mirror the `audit_loader_completeness_test.go` pattern). Then run it, enumerate newly-exposed orphans (mcp-matrix, tool-policy, cache, db, feedback, observability, report), register each in `audit_registry.go` registry/exceptions with a one-line rationale. Also fix REQ-CAR-007(a)(b) comment/registry drift here (same files). Sequencing note (design.md interaction): M-2's new `gate` loader must land BEFORE this milestone so the registry/completeness expectations are updated once, against the post-loader surface.

### M-5 — Mechanical doc/rule sweep (lowest risk, most files)
- H1 role_profiles removal (3 rule files), H2 config.yaml → sections-SSOT (7 sites incl. harness.md:129 gate → `sections/` check), M2 delivery.md ×3, M4 MCP_OAUTH_SETUP.md:163, M6 phase_weights + development_mode ×5, LOW (c)(d)(e). LOW (c) scope guard: edit ONLY the dynamic-workflows.md workflow_agents recommendation enum (~:87/:91/:95) — `internal/config/defaults.go:29,448` carry LIVE haiku (validator/default surface, verified 2026-07-25) and stay untouched. Template-first for every mirrored file; `make build`; parity grep both trees.

### M-6 — M1 dependency note + orphan documentation + closeout
- Record M1 resolved-by-DB-removal dependency (fallback: delete 3 keys if stalled). Add tool-policy/lsp SSOT doc references + acknowledged-orphan list (single location, e.g. settings-management.md or a config README). Final verification batch (acceptance.md §D), lint delta check, neutrality guards.

## §G Anti-Patterns

- Fixing local `.claude/` docs without the template mirror (breaks REQ-CAR-008).
- Grep-only "fixed" claims — every AC must be an executed command with observed output (verification-claim-integrity §1.1).
- Implementing db auto_sync keys (violates REQ-CAR-015 cross-SPEC constraint).
- Blind sed across doc files — per-site Edit with content anchors (line numbers drift).

## §H Cross-References

- Audit SSOT: `.moai/reports/moai-config-audit-20260725.md`
- DB evidence: `.moai/reports/db-scaffold-sync-logic-20260725.md`
- Backlog: SPEC-ASTGREP-DOGFOOD-CLEANUP-001
- Rules: CLAUDE.local.md §2 (Template-First, local-only list), §25 (template neutrality)
