---
id: SPEC-CONFIG-AUDIT-REPAIR-001
title: "Acceptance criteria — config audit repair"
version: "0.2.0"
status: completed
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: P1
phase: "v3.0.x target"
module: ".claude/rules + .claude/skills + internal/config + internal/hook + internal/template/templates"
lifecycle: spec-anchored
tags: "config, audit-repair, acceptance"
tier: L
---

# Acceptance — SPEC-CONFIG-AUDIT-REPAIR-001

All ACs are machine-verifiable: each names the exact command and the expected observation. Every PASS claim must cite the executed command's verbatim output (verification-claim-integrity §2).

## §D AC Matrix

| AC | REQ | Command | Expected |
|----|-----|---------|----------|
| AC-CAR-001 | 001 | `grep -rn 'role_profiles' .claude/rules/moai/workflow/worktree-integration.md .claude/rules/moai/development/agent-authoring.md .claude/rules/moai/development/model-policy.md` | 0 matches (or only explicit "retired" historical wording, no live-config phrasing) |
| AC-CAR-002 | 002 | `grep -rn '\.moai/config/config\.yaml' .claude/skills/moai/SKILL.md .claude/agents/moai/manager-spec.md .claude/skills/moai/workflows/run/context-loading.md .claude/skills/moai/workflows/sync/quality-gates-context.md .claude/rules/moai/core/settings-management.md CLAUDE.local.md` | 0 raw matches, with exactly these enumerated allowed exceptions permitted if present: (1) a historical/legacy-path note explicitly marked as nonexistent-optional (`resolver.go loadProjectTier` explanation), (2) template `moai.version` aggregate mentions in `internal/template/templates/.moai/config/config.yaml` release-sync docs (CLAUDE.local.md §5 version-sync list). Any other match = FAIL |
| AC-CAR-003 | 003 | `grep -n 'config.yaml' .claude/skills/moai/workflows/harness.md` + read the abort-gate block | Gate checks `.moai/config/sections/` existence; no config.yaml gate |
| AC-CAR-004 | 004 | `grep -cn 'spec_git_workflow' .claude/skills/moai/workflows/sync/delivery.md` and `grep -n '[^_]git_workflow' .claude/skills/moai/workflows/sync/delivery.md` | ≥4 spec_git_workflow occurrences; 0 bare `github.git_workflow` reads |
| AC-CAR-005 | 005 | `grep -rn 'mcp-servers.yaml' --include='*.md' . \| grep -v '.moai/specs/SPEC-CONFIG-AUDIT-REPAIR-001/' \| grep -v '.moai/reports/'` | 0 matches (exclusions encoded in the command: this SPEC dir + reports) |
| AC-CAR-006 | 006 | `grep -rn 'phase_weights' .claude/rules/moai/design/` and `grep -rn 'quality\.yaml.*development_mode' .claude/ \| grep -v constitution.development_mode` | 0 phase_weights refs to design.yaml; 0 bare development_mode path citations |
| AC-CAR-007 | 007 | (a) `grep -n 'not yet runtime-consumed' internal/config/audit_loader_completeness_test.go` → 0 for db; (b) `grep -n 'constitution' internal/config/audit_registry.go` shows loader-present labeling; (c) `grep -n 'haiku' .claude/rules/moai/workflow/dynamic-workflows.md` → 0 in the workflow_agents recommendation enum sites (lines ~87/91/95 closed set + purpose rows) ONLY — the Go validator/default surface `internal/config/defaults.go:29,448` carries LIVE haiku and MUST remain untouched (`grep -c 'haiku' internal/config/defaults.go` unchanged from baseline); (d) quality-gates-context.md:~197 describes `db.auto_sync` as map (or site removed by DB track); (e) `grep -n 'coverage_threshold: 0' .moai/config/sections/quality.yaml` → 0 | Each sub-check as stated |
| AC-CAR-008 | 008 | For each edited `.claude/` file with mirror: `diff <local> internal/template/templates/<path>` on the corrected passage (allowing intentional local-only deltas) + `make build` exit 0 + `go test ./internal/template/... ` green | Parity on corrected content; build green; neutrality tests green |
| AC-CAR-009 | 009 | `go test -run TestAuditParity -v ./internal/config/` | Output contains `--- PASS: TestAuditParity` (or a real FAIL under investigation); MUST NOT contain `--- SKIP` |
| AC-CAR-010 | 010 | `go test ./internal/config/... -v 2>&1 \| grep -c 'SKIP'` | 0 skips; full package green |
| AC-CAR-011 | 011 | Go test proving config round-trip: a `gate.yaml` with `enabled: true` (schema per design.md) loads into `AstGrepGate.Enabled=true`, and a unit test on `hook/quality/gate.go` shows the :281 guard branch is reached when enabled | `go test -run <GateEnableTest> -v ./internal/...` → PASS; guard reachable via config |
| AC-CAR-019 | 019 | Characterization test: default config (no gate.yaml / no enabled key) yields `AstGrepGate.Enabled=false`; plus `go test ./internal/hook/...` green with default fixtures | PASS — default OFF invariant proven |
| AC-CAR-020 | 020 | (a) `grep -n 'utils' .moai/config/astgrep-rules/sgconfig.yml` → 0; (b) enabled-path test (or documented manual run) shows config-mode scan loads root `go-hardcoding.yml`/curated set without full-scan failure; (c) test with `sg` absent from PATH → gate skips with notice, exit non-error | All three sub-checks as stated |
| AC-CAR-021 | 021 | `grep -rn 'RunAstGrepGate\b' internal/ \| grep -v _test.go \| grep -v 'RunAstGrepGateV2'` | 0 matches (V1 deleted/folded per design.md); `go build ./...` green |
| AC-CAR-012 | 012 | `grep -n 'ASTGREP-DOGFOOD-CLEANUP-001' .moai/specs/SPEC-CONFIG-AUDIT-REPAIR-001/plan.md` | Absorption boundary recorded (≥1 match) naming what remains in the backlog (16-language productization, empty stubs, message unification, SPEC-ID stripping) |
| AC-CAR-013 | 013 | Dev-only route: `grep -n 'tool-policy' CLAUDE.local.md` shows §2 entry; `moai tool-policy list` in a clean `/tmp` project exits with a graceful one-line dev-only message (exit code non-panic, message contains no Go stack) | Per chosen option |
| AC-CAR-014 | 014 | `grep -rn 'mcp-matrix.yaml' internal/template/templates/` | 0 dangling references (reference removed/reworded) OR the yaml is present in template sections (distribute route) |
| AC-CAR-015 | 015 | `grep -rn 'auto_sync' internal/ cmd/ \| grep -v _test.go \| grep -vE 'debounce_seconds\|require_user_approval\|excluded_patterns'` unchanged from baseline; spec.md/plan.md record the DB-track dependency | No implementation of the 3 keys; dependency note present |
| AC-CAR-016 | 016 | `grep -rn 'tool-policy.yaml\|lsp.yaml' .claude/rules/ .moai/docs/ CLAUDE.local.md` | ≥1 SSOT doc reference each for tool-policy.yaml and lsp.yaml; acknowledged-orphan list present in exactly one location |
| AC-CAR-017 | 017 | `golangci-lint run 2>&1 \| tail -5` compared against captured baseline log | 0 new issues vs baseline |
| AC-CAR-018 | 018 | `go test ./internal/template/... -run 'Neutrality\|Leak' -v` (plus CI on PR) | All neutrality/leak guards PASS |

## §D.1 Given-When-Then Scenarios

1. **M7 test revival** — Given a plain `go test ./internal/config/...` invocation from any cwd, When TestAuditParity runs, Then it resolves the repo root via `runtime.Caller`, scans the real `.moai/config/sections` tree, and reports PASS/FAIL (never SKIP), with every real-tree section either loader-covered or a registered exception.
2. **Distributed-user tool-policy UX** (dev-only route) — Given a fresh `moai init` project in `/tmp` without `tool-policy.yaml`, When the user runs `moai tool-policy list`, Then the CLI prints a single graceful message stating the feature is maintainer-only/not distributed, and does not emit a raw file-not-found hard error.
3. **Harness abort gate** — Given a normally-initialized project (sections present, no root config.yaml — the universal state per audit V1), When the harness workflow evaluates its abort gate, Then it proceeds (gate passes on `sections/` existence) instead of always aborting.
4. **ast-grep gate opt-in enable** — Given a project with `gate.yaml` setting the gate enabled and the repaired ruleset present, When a hook-gated tool call reaches `pre_tool.go`, Then `RunAstGrepGateV2` executes against the loaded ruleset; and Given the default config (no enable), Then behavior is byte-identical to pre-SPEC (gate OFF).

## §D.2 Edge Cases

- DB-removal track lands first and deletes `db.yaml`: AC-CAR-007(d) and AC-CAR-015 re-evaluate against the removed surface (absence satisfies both).
- Template mirror does not exist for a corrected local doc (e.g. CLAUDE.local.md, dev-only rules): parity AC-CAR-008 skips that file by design — record the skip list.
- Gate restore + `sg` not installed on CI: AC-CAR-020(c) graceful-degrade path must be the CI-observed default (no CI dependency on the `sg` binary).
- V1 `RunAstGrepGate` deletion exposes tests referencing it: adjust within the same milestone; full suite green is the gate.

## §D.3 Quality Gate / Definition of Done

- All 21 ACs evaluated with executed commands and verbatim output citations.
- `go test ./...` green; `go vet ./...` clean; lint delta 0 vs baseline; `make build` green.
- All 3 design decisions RESOLVED and recorded in design.md (Implementation Kickoff 2026-07-25): D1 gate restore (default OFF), D2 tool-policy dev-only, D3 mcp-matrix reword.
- Frontmatter transitions: draft → in-progress (manager-develop) → implemented → completed (manager-docs).
