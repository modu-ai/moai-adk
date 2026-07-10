# Progress — SPEC-WEB-CONSOLE-014

Lifecycle tracking for the 3-phase plan → run → sync flow. §E carries the audit-ready signals.

## §E.1 Plan-phase Audit-Ready Signal

- **Phase**: plan (artifacts authored)
- **Tier**: M (3-artifact: spec.md + plan.md + acceptance.md, + progress skeleton)
- **Status**: `draft`
- **plan_complete_at**: 2026-07-10
- **plan_status**: audit-ready (plan-audit iter-1: PASS-WITH-DEBT 0.86 → iter-2 D1-D6 정정 반영, v0.2.0)
- **Artifacts created (v0.2.0 amended)**:
  - `spec.md` — 12-field canonical frontmatter (+ tier: M, era: V3R6, depends_on: SPEC-WEB-CONSOLE-013), HISTORY (0.1.0 + 0.2.0), 16 GEARS requirements (REQ-WC14-001..003, 010..012, 020..021, 030..031, 040, 050..051, 060..062; iter-2 개정: 010/020/040/050 + F6/F9 재분류), Findings F1-F12, Exclusions 8× `### Out of Scope —` H3 (iter-2: sandbox/observability 배선 추가).
  - `plan.md` — §A Context (+§A.1 의존성 gate: 013 실존 draft, 직렬 gate 유지), §B Known Issues B1..B14 (실측 정정 4건: B1/B2 브리프 + B11/B12 iter-2), §C Pre-flight 7항 (C-3 반증 grep 확장), §D Constraints (+B14 ordering), §E Self-Verification, §F Milestones M0→M1→{M2,M3,M4}→M5→M6, §G Anti-Patterns AP-1..10, §H Cross-References.
  - `acceptance.md` — §A GWT 시나리오 4건, §B plan-phase 검증 증거 E-1..E-15 + Gaps 3건, §C AC Matrix 23 AC ([B] 22 / [N] 1; iter-2: 010c/040a/040b), §D DoD 5항.

## §E.2 Run-phase Evidence

Worktree: `.claude/worktrees/agent-a4322e53cb6b13381` (branch `worktree-agent-a4322e53cb6b13381`, base `f914d66f6` — SPEC-WEB-CONSOLE-013 close present, depends_on satisfied). All milestone SHAs below are **worktree SHAs**; the orchestrator rebases onto main before landing (expect an orchestrator SHA-correction commit).

### M0 — Pre-flight (§C 1-7)

- **C-1 의존성 gate**: `ls .moai/specs | grep WEB-CONSOLE-013` → present; `f914d66f6` is `docs(SPEC-WEB-CONSOLE-013): backfill sync_commit_sha` (013 run+sync landed → shared files handoff/cache seam registration present). Gate satisfied (orchestrator confirmed 013 status=completed).
- **C-1b**: 013's typed Model Policy surface = `internal/web/modelpolicy.{go,templ}` (FieldDef-less read-only view); no collision with `workflow.model_routing` flat prefix (denylist entry). Verified.
- **C-2 앵커 재실측** (content-token): `readTierThresholds`(internal/cli/hook.go), rate-limit compile const (internal/harness/safety enforcement uses `rateLimit*` names, not `MaxPerWeek/CooldownHours`), `validMergeMethods`/`ValidMergeMethods()`/`IsValidMergeMethod()`(internal/config/validation.go:271/290/283), `MergeMethod`(internal/config/types.go:125), `LoadDangerConfig`(internal/mx/danger_category.go:51), hook_metrics seam (schema_sections.go:241-242), permission scaffold (config grep empty).
- **C-3 반증 (B1/B2/B11/B12)** — all confirm plan scope (0 disproof):
  - B1 rate_limit: only `internal/cli/harness.go` reads `cfg.RateLimit.MaxPerWeek/CooldownHours` (display). `internal/constitution/rate_limiter.go` + `internal/evolution/` use own compile consts for OTHER domains (constitution amendment / evolution proposal) — NOT learning.rate_limit. → display-only confirmed.
  - B2 permission: `grep pre_allowlist|session_rules internal/config internal/permission` = ∅ → Go-unbound scaffold confirmed.
  - B11 sandbox: `grep Security.Sandbox|Sandbox.NetworkAllowlist|Sandbox.EnvScrubExtra internal/ (excl types.go)` = ∅; `grep EnvScrubExtra internal/sandbox/*.go` = ∅ → config→Options bridge absent (scaffold) confirmed.
  - B12 output_path: only `internal/settings/schema_sections.go` (self) + `internal/hook/post_tool_duration.go` `hookMetricsRelPath` const (the actual write path is fixed) → dead config confirmed.
- **C-4 렌더 파이프라인**: `fieldsetSchemaSection` renders SectionFields + ReadOnlyDisplayFields (per-section) + RawViewBlocks (per-section). Raw-only sections (0 editable fields) render via the same templ. mx rendered as a web-only `schemaSectionMeta` (NOT added to `settings.SchemaSectionIDs()` — that accessor requires ≥1 field per `TestSchemaSectionsRegistered`; no test couples metas↔SchemaSectionIDs).
- **C-5 병렬 세션 방어**: worktree isolated (branch `worktree-agent-a4322e53cb6b13381`); shared-main race observed (`4c0a2c64→325de988` SPEC-CLI-TUX-V3-001 landed between two reads) but does NOT touch internal/settings|web (verified `git diff --name-only f914..325de988 -- internal/settings internal/web` = ∅). Working in isolated branch.
- **C-6 템플릿 미러**: internal/settings + internal/web are Go source (not under internal/template/templates) → no mirror. i18n.js is `internal/web/assets/i18n.js` (embedded asset, not template tree). Confirmed.
- **C-7 i18n locale 앵커** (drift 전제 재실측): en:L20, ko:L395, ja:L770, zh:L1145 (`grep -n '^  en:|^  ko:|^  ja:|^  zh:'`).

### M1 — 노출 금지 가드 (P3 선행)

- New: `internal/settings/dormant_guard_test.go` (`TestDormantConfigNeverEditable` — table-driven denylist over `AllFields()` + allowlist pin for `slow_hook_threshold_ms`). B14: `learning.auto_apply`/`observability.hook_metrics.output_path` NOT preloaded (added in M2 with demotion — TDD RED→GREEN pair).
- Edit: `internal/settings/sectionroute_test.go` — explicit `sunset`/`mx` RouteExcluded pins (AC-051).
- Result: all M1 guards immediate GREEN (regression pins over current code — no dormant leak). Full affected suite (settings/web/cli) green.

_AC matrix + M2-M6 evidence appended per milestone below._

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
