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

### M2 — learning + observability 정직화

- schema_sections.go: seam FieldDef 철거 learning.auto_apply (F3) + observability.hook_metrics.output_path (F9) → ReadOnlyDisplayFields 강등(NoteKey honest label). slow_hook_threshold_ms editable 잔류(회귀 핀). RawViewBlocks +2: learning.tier_thresholds(generic) + learning.rate_limit(informational). NoteKey 필드 추가(ReadOnlyField/RawBlockRef).
- web fieldsets.templ: schemaReadOnlyRow/schemaRawBlock에 noteKey 파라미터(roNoteKey/rawNoteKey resolver, schema_label.go). templ regen(no-drift 확인). i18n.js: ro.note.governance/dead_config + raw.note.informational ×4.
- denylist +2 (auto_apply/output_path exact, B14 RED→GREEN 쌍). 全 GREEN(full `go test ./...` = 유일 pre-existing flake internal/hook TestHookWrapper_TempFileCleanup, isolation PASS → out-of-scope debt).

### M3 — merge_method editable select

- gitStrategyFields: 3-profile merge_method select(옵션 config.ValidMergeMethods() SSOT 정렬 파생, EmptyLabel "(project default)"). applyGitStrategyKey: merge_method enum 검증(config.IsValidMergeMethod — empty/out-of-enum 거부). i18n.js: 9 merge_method 키 ×4. schema_bridge parity green(typed→isWebOnlyKeyChipField scoped out).

### M4 — raw views (security + mx)

- SectionMx const(raw-only, SchemaSectionIDs 미포함). RawViewBlocks +4: security.sandbox.{network_allowlist,env_scrub_extra}(F6 informational) + mx.{danger_categories,test_paths}(generic). web schemaSectionMetas/consoleTabs +mx(alert-circle icon). i18n sec.mx.title/desc ×4. testdata/sections/mx.yaml fixture. TestMXRawViewRendered + TestRawBlockValues 확장.

### M5 — i18n 4-locale parity + bridge sweep

- TestSPEC014I18nKeysFourLocale: 14 신규 키 × 4-locale 존재 기계 검증(REQ-060). CLI 4 bridge tests green 무수정(REQ-061 both-surface).

### §E.2.1 AC Matrix (23 AC — verification-claim-integrity §3)

| AC | Sev | Status | Evidence |
|----|-----|--------|----------|
| AC-WC14-000 | [B] | PASS | `go build ./...` exit 0 + `go test ./internal/settings/... ./internal/web/... ./internal/cli/...` exit 0 (verify/e07c0351/m5-affected.log) |
| AC-WC14-001a | [B] | PASS | grep `func TestDormantConfigNeverEditable` = dormant_guard_test.go:80; denylist has `learning.auto_apply` exact; `go test -run TestDormantConfigNeverEditable` PASS |
| AC-WC14-001b | [B] | PASS | TestApplySchemaEditsRejectsUnknownAndReadOnly PASS (learning.auto_apply 거부); schema_sections.go auto_apply ReadOnly-reg=1, seam edit-gen=0 |
| AC-WC14-002 | [B] | PASS | grep tier_thresholds in schema_sections.go = 2 (≥1); TestRawBlockValues PASS |
| AC-WC14-003 | [B] | PASS | grep rate_limit in schema_sections.go = 2 (≥1); TestRawBlockValues PASS |
| AC-WC14-010a | [B] | PASS | grep merge_method in schema_sections.go = 5 (≥1); grep `func TestMergeMethodFieldsExposed` ≥1; test PASS (AllFields 정확 3건) |
| AC-WC14-010b | [B] | PASS | TestApplySchemaEditsGitStrategyTyped PASS (merge_method 라운드트립 + 주석 보존) |
| AC-WC14-010c | [B] | PASS | grep `func TestMergeMethodAbsentKeyDisplay` ≥1; test PASS (absent 빈값 읽기 성공 + empty 저장 거부) |
| AC-WC14-011 | [B] | PASS | grep ValidMergeMethods in schema_sections.go = 2 (≥1); enum 밖 값 거부 test PASS; grep `"squash"` = 0 (리터럴 재선언 금지) |
| AC-WC14-012 | [B] | PASS | denylist F5 4 substr(branch_creation.prompt_always/.auto_enabled/automation.auto_branch/.auto_pr); TestDormantConfigNeverEditable PASS |
| AC-WC14-020 | [B] | PASS | grep network_allowlist = 1, env_scrub_extra = 1 (each ≥1); TestRawBlockValues PASS; F2 informational 라벨 raw.note.informational ∈ AC-060 |
| AC-WC14-021 | [B] | PASS | grep pre_allowlist\|session_rules in schema_sections.go = 0; denylist substr 항목 포함; TestDormantConfigNeverEditable PASS |
| AC-WC14-030a | [B] | PASS | grep danger_categories = 1, test_paths = 1 (each ≥1); TestRawBlockValues PASS |
| AC-WC14-030b | [B] | PASS | grep `func TestMXRawViewRendered` = mx_rawview_test.go:17; `go test ./internal/web/ -run TestMXRawViewRendered` PASS (GET / mx 컨테이너 존재) |
| AC-WC14-031 | [B] | PASS | TestExcludedSectionsAllRejected PASS; denylist `mx.` prefix 포함 |
| AC-WC14-040a | [B] | PASS | TestDormantConfigNeverEditable allowlist 핀 assertEditablePresent(slow_hook_threshold_ms) PASS |
| AC-WC14-040b | [B] | PASS | output_path AllFields 부재(seam edit-gen=0) + ReadOnly-reg=1 + TestApplySchemaEditsRejectsUnknownAndReadOnly PASS + denylist exact |
| AC-WC14-050 | [B] | PASS | grep `func TestDormantConfigNeverEditable` ≥1; test PASS (sunset./model_upgrade_review/workflow.model_routing/mx./tool-policy prefix + 정확명) |
| AC-WC14-051 | [B] | PASS | TestExcludedSectionsAllRejected PASS; grep sunset\|tool-policy\|"mx" in sectionroute_test.go = 4 (≥3) |
| AC-WC14-060 | [B] | PASS | 14 신규 키 per-locale count = 4/4 (en/ko/ja/zh); TestSPEC014I18nKeysFourLocale + TestDataI18nKeysSubsetOfDictionary + TestI18nDictionaryEmbedded PASS |
| AC-WC14-061 | [B] | PASS | `go test ./internal/cli/ -run TestI18nKeySetParity\|TestI18nSegmentParity\|TestBridgeFieldDefResolver\|TestTUIRendersSchemaFieldSet` exit 0 (모든 SPEC-014 필드 seam/typed → isWebOnlyKeyChipField scoped out) |
| AC-WC14-062 | [B] | PASS | `git diff --name-only f914d66f6..HEAD -- internal/statusline/` = 0줄 |
| AC-WC14-063 | [N] | PASS | `golangci-lint run internal/settings/... internal/web/... internal/cli/...` = `0 issues.` (baseline 유지, NEW 0) |

**AC 결과**: 23/23 PASS ([B] 22 + [N] 1). AC-WC14-063은 [N].

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-11
run_commit_sha: pending-backfill-M6   # worktree HEAD; orchestrator rebase 후 backfill
run_status: PASS
ac_pass_count: 23
ac_fail_count: 0
preserve_list_post_run_count: 0        # internal/statusline + internal/template/templates 무접촉 (git diff 0)
l44_pre_commit_fetch: n/a-worktree     # isolated worktree branch, shared-main race 회피
l44_post_push_fetch: n/a-no-push       # 지시대로 push 안 함
new_warnings_or_lints_introduced: 0    # golangci-lint 0 issues, go vet clean
cross_platform_build:
  linux_darwin: pass                   # go build ./... exit 0
  windows: pass                        # GOOS=windows GOARCH=amd64 go build ./... exit 0
total_run_phase_files: 17              # SPEC artifacts 4 + internal/settings 7 + internal/web 6
m1_to_mN_commit_strategy: per-milestone-worktree-commit  # M1-M6 각 1 commit, no push
coverage:
  internal_settings: 89.4%             # ≥85% ✓
  internal_web: 70.8%                  # pre-existing package baseline (신규 코드는 렌더 테스트로 커버; 회귀 아님)
templ_drift: none                      # templ generate 재실행 no-op (CI guard green)
make_build: pass                       # bin/moai 재컴파일 exit 0, working tree clean
worktree_shas:
  M1: 76ececa733b4920d714a82d3a45aa93cda54a0fb   # test: dormant guards + draft→in-progress
  M2: b3eeabfeb49319b0250ee3f97af6ab7d709a8394   # feat: learning/observability honesty
  M3: 5d5297a8c28c4d233975e9e35eed0635c2703a34   # feat: merge_method select
  M4: faa808d88a0512a5dc2ead25aeadbacaeff54fea   # feat: raw views security+mx
  M5: 898c8df5a38466cd1bd0c12cb2373d8e4e5f5c4b   # test: i18n 4-locale parity guard
out_of_scope_debt:
  - internal/hook TestHookWrapper_TempFileCleanup (load-sensitive flake, isolation PASS; do-not-chase)
  - internal/web package coverage 70.8% < 85% (pre-existing render-path baseline, not introduced here)
```

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
