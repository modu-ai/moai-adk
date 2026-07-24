# Research — SPEC-CONFIG-AUDIT-REPAIR-001

Digest of the source audit and the evidence base this SPEC binds to. SSOT: `.moai/reports/moai-config-audit-20260725.md` (all findings CONFIRMED, Gaps 0; two second-pass verifications V1/V2 closed the last unknowns).

## A. Evidence map (finding → mechanically-observed evidence)

| Finding | Evidence (as recorded in the audit) |
|---------|-------------------------------------|
| H1 | `worktree-integration.md:93,98,178-179,193,239,249` (2 HARD clauses + `ORC_WORKTREE_REQUIRED` lint), `agent-authoring.md:168`, `model-policy.md:166`; `grep '^team:' workflow.yaml` → no match (Agent Teams retired per CLAUDE.md §4/§15) |
| H2 + V1 | `ls .moai/config/config.yaml` → No such file (local AND template); repo-wide `WriteFile/Create` grep → no code writes it; `resolver.go:349 loadProjectTier` optional-read only. Referencing sites: SKILL.md:19,335, manager-spec.md:102, harness.md:129 (abort gate — always fails), run/context-loading.md:130, sync/quality-gates-context.md:108, settings-management.md, CLAUDE.local.md §5/§9 |
| H3 | `defaults.go:286-296` (AstGrepGate never enabled), `loader.go:47-98` (no gate entry), `hook/quality/gate.go:281` guard, `pre_tool.go:566-575` unreachable → `RunAstGrepGateV2` dead |
| H4 | `toolpolicy/loader.go:19` hard error on absent file; 47KB yaml not in template sections; user-facing doc refs 0; dev-only intent only in `audit_loader_completeness_test.go:39` allowlist comment |
| M1 | `db.yaml:15-23` — `auto_sync.{debounce_seconds,require_user_approval,excluded_patterns}` zero Go consumers (repo-wide grep no-hit); hook `IsExcluded` uses built-in defaults; only `migration_patterns` consumed via fragile line-scan `cli/hook.go:513-520`. Cross-confirmed by `.moai/reports/db-scaffold-sync-logic-20260725.md` |
| M2 | `delivery.md:17,215,417` read nonexistent `github.git_workflow`; actual key `spec_git_workflow` (`system.yaml:44`); `delivery.md:27` correct → internal contradiction |
| M3 | Shipped template skill `project/doc-generation.md` references `mcp-matrix.yaml`; template does not distribute it; Go consumers 0 |
| M4 | `MCP_OAUTH_SETUP.md:163` → nonexistent `.moai/config/mcp-servers.yaml` (real surfaces: `.mcp.json`, `mcp-matrix.yaml`) |
| M5 | Local `sgconfig.yml:24` phantom `utils` ruleDir (config-mode `sg scan` full-failure risk); root `go-hardcoding.yml` not loaded in config-mode (`scanner.go:250-281`) — backlog SPEC-ASTGREP-DOGFOOD-CLEANUP-001 |
| M6 | `design/constitution.md:60` cites nonexistent `design.yaml adaptation.phase_weights`; 5 docs (`run.md:74`, `moai.md:152`, …) cite `quality.yaml development_mode` bare; real path `constitution.development_mode` |
| M7 + V2 | `go test -run TestAuditParity -v` → `--- SKIP` (cwd = package dir; relative `.moai/config/sections` not found). Only `TestAuditLoaderCompleteness` (template-dir scan) is alive → local-only files (mcp-matrix etc.) escape all guards. Fix pattern: `runtime.Caller` repo-root resolution as in `audit_loader_completeness_test.go` |
| LOW | Orphan sections (doc refs 0): tool-policy(47KB), lsp(8.2KB), security, observability, report, sunset, archive, cache, feedback, project. Comment/registry drift: completeness test :16 db comment false; audit_registry constitution "no loader" false. `coverage_threshold: 0` dead local leaf (`schema_sections.go:205`). `dynamic-workflows.md:91` retired haiku enum. `quality-gates-context.md:197` db.auto_sync scalar misdescription |

## B. Healthy items (explicitly NOT defects — do not touch)

sunset DORMANT (by design); astgrep empty stubs 10/17 (CLAUDE.local §2.2); gate/runtime defaults-only structs (in use); evaluator-profiles all 4 reachable; toolpolicy schema ↔ yaml fully consistent; handoff/llm/workflow/harness/interview keys spot-checked consistent; template astgrep subset clean (0 internal-content leaks).

## C. Cross-SPEC landscape

- **DB-removal track** (parallel session, APPROVED; evidence `.moai/reports/db-scaffold-sync-logic-20260725.md`, memory project_db_retire_removal_approved.md): owns db.yaml/auto_sync fate → M1 scoped as dependency.
- **SPEC-ASTGREP-DOGFOOD-CLEANUP-001** (backlog, draft): owns local sgconfig/ruleset repair → M5 delegated; H3 decision here must not conflict (both options preserve the explicit `moai ast-grep` CLI path).
- **feat/SPEC-CLI-TUX-INIT-UPDATE-001** (active branch at authoring time): shares no finding sites identified, but rebase-before-run is mandated in plan.md §C.

## D. Template-mirror inventory (to be re-verified at run pre-flight)

Candidate mirrored files among fix targets: `.claude/skills/moai/SKILL.md`, `.claude/agents/moai/manager-spec.md`, `.claude/skills/moai/workflows/{harness.md, run/context-loading.md, sync/quality-gates-context.md, sync/delivery.md}`, `.claude/rules/moai/**` (worktree-integration, agent-authoring, model-policy, settings-management, dynamic-workflows, design/constitution). Non-mirrored (local-only): CLAUDE.local.md, repo-local rules per §2 list. The exact two-tree edit list is a run pre-flight deliverable (`test -f internal/template/templates/<path>` per file) — line numbers and mirror status drift, so this inventory is indicative, not binding.
