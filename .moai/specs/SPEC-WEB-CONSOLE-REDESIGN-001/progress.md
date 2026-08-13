# SPEC-WEB-CONSOLE-REDESIGN-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-08-08
tier: L
artifacts: spec.md, plan.md, acceptance.md (Tier L 편차 — plan.md §G.1 참조)
open_clarifications: 0 (D3=(c) 제거, D4=(i) 2-option radio + __present 보존 확정)

## §E.2 Run-phase Evidence

M1-M3 완료. M4(서드파티 LLM 탭) / M5(프로필 UI) / M6(autonomy stub) / M7(횡단 마감)은
별도 위임 소관이라 이 구간에서 착수하지 않았다.

| AC | 상태 | 검증 명령 | 실측 |
|----|------|-----------|------|
| AC-WCR-001 | PASS | `go test ./internal/web/ -run TestDeadConfigAbsentFromSchema` | ok — `token_budget.{plan,run,sync}` 부재 |
| AC-WCR-002 | PASS | `go test ./internal/web/ -run TestDeadConfigAbsentFromSchema` | ok — `auto_clear.*` 4종 부재 |
| AC-WCR-003 | PASS | `go test ./internal/config/ -run TestWorkflowBackwardCompat` | ok — 두 블록 담은 yaml 로드 성공, 접근자 4종 디스크 값 반환 |
| AC-WCR-004 | PASS | `go test ./internal/web/ -run TestGitStrategyRendered` | ok — mode + 3× merge_method 컨트롤 렌더 |
| AC-WCR-010 | PASS | `go test ./internal/web/ -run 'TestConsoleTabsOrder\|TestTabPanelRenderOrderMatchesTabs'` | ok — 9탭 순서 + 패널 렌더 순서 일치 |
| AC-WCR-011 | PASS | `go test ./internal/web/ -run TestWorkflowTabRetainsProseConsumedFields` | ok — 산문 소비 6필드 잔류 |
| AC-WCR-012 | PASS | `go test ./internal/web/ -run TestGitWorktreeTabFields` | ok — git_strategy 4 + worktree 4 + branch_guard 1 |
| AC-WCR-013 | PASS | `go test ./internal/web/ -run TestAuditTabFields` | ok — 4필드 렌더 + Section/Persist 무변경 |
| AC-WCR-020 | PASS | `go test ./internal/web/ -run 'TestBoolFieldsRenderAsRadio\|TestCountCheckboxInputsIsFalsifiable'` | ok — checkbox 0개, bool마다 radio 2개. 검출기 falsifiability는 합성 fixture로 별도 증명 |
| AC-WCR-021 | PASS | `go test ./internal/web/ -run 'TestSchemaTogglePresentCompanion\|TestParseSchemaFormBoolSemanticsUnchanged'` | ok — `__present` 보존, 파서 3분기 동작 동일 |
| AC-WCR-022 | PASS | `go test ./internal/web/ -run 'TestClosedSetFieldsAreClosedWidgets\|TestClosedSetRejectsOutOfSetValue'` | ok — 14필드 select/radio + Validate, 집합 밖 값 거부 |
| AC-WCR-023 | PASS-WITH-DEBT | `go test ./internal/web/ -run 'TestFreeTextWhitelist\|TestM4PendingFreeTextStillPending'` | ok — 잔류 자유 텍스트는 화이트리스트 + M4 미착수 4종(`llm.glm.models.*`). 화이트리스트에 SPEC 열거 누락분 2종(`user_name`, `security.sandbox.docker_image`) 추가 — 둘 다 진정 열린 값 |
| AC-WCR-060 | PASS | `go test ./internal/web/ -run TestI18n -v` + `git diff --stat <base>..HEAD -- internal/web/i18n_untranslated_allowlist_test.go` | 14/14 PASS, allowlist diff 0바이트 (신규 예외 0건). 신규 키 44종/로케일 |
| AC-WCR-061 | 공백 통과 | `git diff --name-only <base>..HEAD -- internal/template/ .moai/config/sections/` | 출력 없음 — yaml/템플릿 편집 0건이므로 미러·`make build` 의무 미발생 |
| AC-WCR-062 | 공백 통과 | (AC-WCR-061과 동일 근거) | 템플릿 미러 미갱신 → 중립성 가드 대상 없음 |
| AC-WCR-063 | FAIL (M7 소관) | `go test -cover ./internal/web/...` | 65.4% (< 90.0%). 베이스라인 62.2% 대비 +3.2pp. 90% 달성은 M7 횡단 마감 항목 |

품질 게이트:

- `go build ./...` → exit 0
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- `go vet ./...` → exit 0
- `golangci-lint run --timeout=5m` → `0 issues.`
- `go test ./...` → `internal/template` 3건 실패(`TestLateBranchTemplateMirror` /
  `TestRuleTemplateMirrorDrift` / `TestSanitizedPairParity`). 전부 PRE-EXISTING:
  변경분을 `git stash` 한 베이스라인에서 동일하게 실패함을 실측 확인. 신규 실패 0건
- `templ generate` 재실행 후 `*_templ.go` 해시 무변경 (생성 산출물 동기화 완료)

### §E.2.1 `workflow.execution_mode` 닫힌 집합 정정 (M3 잔여 항목 해소)

M3는 `config.ValidExecutionModes()`를 `{auto, solo, team}`으로 선언했으나 이 집합은
**추론**이었고 SSOT 근거가 없었다. M4-M6 구간에서 실측한 결과 `cg`가 누락된 4번째
값임이 확인됐다.

| 근거 | 관측 |
|------|------|
| `.moai/config/sections/harness.yaml` `mode_defaults` 키 | `solo` / `team` / `cg` (3개) |
| `internal/config/types.go:868` `ModeDefaults` 주석 | "default harness level map per execution mode (solo/team/cg)" |
| `SPEC-V3R2-HRN-001` REQ-HRN-001-014 | "**While** `workflow.yaml execution_mode` is `auto`, the router **shall** consult `cfg.ModeDefaults.solo\|team\|cg`" |
| Go 런타임 reader | 0건 (`ExecutionMode` struct 멤버 + 기본값만; 소비는 산문) |

M3가 이 필드를 닫힌 위젯으로 전환하면서 집합 밖 값의 저장이 거부되므로, 누락은
표시 문제가 아니라 **문서화된 모드 하나를 사용자가 설정할 수 없게 만드는 회귀**였다.

처분: `ValidExecutionModePins()` = `{solo, team, cg}`를 신설하고
`ValidExecutionModes()` = `auto` + pins로 구성했다. `harness.mode_defaults.*`
FieldDef 목록도 같은 접근자에서 파생하도록 바꿔 두 집합이 구조적으로 드리프트할 수
없게 했다 (`modeDefaultFields()`). `internal/config/execution_modes_test.go`가 배포되는
`harness.yaml`의 `mode_defaults` 키를 실제로 파싱해 pin 집합과 대조한다 — 추론이 아닌
산출물 대조다.

### §E.2.2 M4-M6 실행 증거

| AC | 상태 | 검증 명령 | 실측 |
|----|------|-----------|------|
| AC-WCR-030 | PASS-WITH-DEBT | `go test ./internal/web/ -run TestGLMModelSelectOptions` | 4필드 전부 `TypeSelect` + 옵션 `{glm-5.2, glm-5.1, glm-4.7, glm-4.5-air}` + `Validate` 집합 밖 거부. **debt**: AC의 `grep glm-4.5-air \| wc -l` 기대치 1은 착수 전부터 성립 불가였다(베이스라인 7건: `defaults.go` 상수 2 + `types.go` 주석 1 + `statusline/memory.go` 2 + `cli/glm.go` 2). 본 밀스톤은 리터럴을 **0건 추가**했고(`ValidGLMModels()`는 `DefaultGLM*` 상수 파생) 커밋 전후 실측 7 → 7로 동일. AC 의도("스키마 파일에서 리터럴 재선언 금지")는 충족, 숫자 기대치는 미충족 |
| AC-WCR-031 | PASS | `go test ./internal/web/ -run TestGLMEffortTierDefaults` | 옵션 3종(`reasoning-max`/`reasoning-high`/`thinking-off` — `template.GLMReasoningStateNames()` 파생) + 기본값 high·fable=Max, medium=High, low=None. 렌더 preselect까지 단언 |
| AC-WCR-032 | PASS | `go test ./internal/web/ -run TestGLMTierLabelVsKey` | 폼 키 `high`/`medium`/`low`/`fable` 잔류, i18n title은 Opus/Sonnet/Haiku/Fable |
| AC-WCR-033 | PASS | `go test ./internal/web/ -run TestGLMEffortScopeBadge` | `sec.llm.effortnote` 정확히 1회(패널 헤더) + 4로케일 + `effort_level` 원천 명시 + 티어별 `data-store-only` 마커 4개. "티어가 적용된다" 취지 문구 부재 단언 포함 |
| AC-WCR-034 | PASS | `go test ./internal/web/ -run 'TestGLMKeyNeverEchoedByDefault\|TestGLMKeyReveal'` | 기본 렌더 평문 부재 + `value=""` 유지; POST reveal 루프백 200 평문, 비루프백/cross-site/GET 전부 거부, 미설정 시 404 |
| AC-WCR-040 | PASS | `go test ./internal/web/ -run TestProfileManagerCardAbsent` | `profilemgr` / `profile-manager` 마커 0건, 프로필 바 잔존 |
| AC-WCR-041 | PASS | `go test ./internal/web/ -run TestProfileFormsNotNested` | HTML 파서로 메인 form 서브트리 내 `<form>` 0개. 중첩 fixture 음성 대조 포함 |
| AC-WCR-042 | PASS | `go test ./internal/web/ -run TestProfileRename` | 8서브테스트: ok / default / current / conflict / traversal(대상·신규 양쪽) / 부재 / GET. 거부 케이스는 디렉터리 무변경까지 단언 |
| AC-WCR-050 | PASS | `go test ./internal/web/ -run TestAutonomyStubResolved` | 제거안(plan.md D3 (c)): 렌더 HTML에 `/autonomy/tiers` 부재 + 라우트 미등록. `internal/config/autonomy_tiers.go`와 init 경로는 무접촉 |
| AC-WCR-060 | PASS | `go test ./internal/web/ -run TestI18n` | 신규 키 en/ko/ja/zh 전부, allowlist 예외 0건 추가 |
| AC-WCR-061 | 공백 통과 | `git diff 36f9bfea4..HEAD --name-only -- .moai/config/sections/ internal/template/templates/` | 출력 없음 — yaml·템플릿 편집 0건이므로 미러·`make build` 의무 미발생 (M4의 `llm.glm.effort.*`는 스키마 필드 신설일 뿐 배포 yaml 편집이 아니며, 저장 시에만 키가 생긴다) |
| AC-WCR-062 | 공백 통과 | (AC-WCR-061과 동일 근거) | 템플릿 미러 미갱신 → 중립성 가드 대상 없음 |
| AC-WCR-063 | **FAIL** | `go test -cover ./internal/web/` | **73.5%** (< 90.0%). M1-M3 종료 시점 65.4% → 73.5% (+8.1pp). 아래 §E.2.3 참조 |

### §E.2.3 AC-WCR-063 커버리지 — 90% 미달 사유 (실측)

증거: `.moai/state/verify/web-redesign/coverage-analysis.txt`

```
generated (templ): 3022/4323 = 69.9%
handwritten:        806/885  = 91.1%
TOTAL:             3828/5208 = 73.5%
uncovered generated: templ error/ctx boilerplate=927 (17.8% of package), real branches=374
ceiling with EVERY real branch covered: 80.7%
```

이 패키지 구문의 **83%(4323/5208)가 `templ generate` 산출물**이다. 손으로 쓴 코드는
이미 91.1%로 목표를 넘었고, 총계를 끌어내리는 것은 생성 코드다.

생성 코드의 미커버 1301구문 중 **927구문(패키지 전체의 17.8%)은 templ이 모든 write
뒤에 기계적으로 붙이는 `if templ_err != nil { return templ_err }` / `ctx.Err()` /
`NopComponent` 보일러플레이트**다. `strings.Builder`나 `httptest` 버퍼로 렌더하는 한
이 분기는 도달 불가하며, 도달시키려면 일부러 실패하는 `io.Writer`를 주입해야 한다.
그 테스트는 **어떤 동작도 단언하지 않는다** — 위임 브리프가 금지한 "라인만 실행하는
패딩"에 정확히 해당한다.

따라서 실제 분기를 **전부** 덮어도 상한은 **80.7%**이고, 90%는 구조적으로 도달
불가하다. 남은 374구문의 실제 분기는 주로 `agentFMRow`(42, 에이전트 파일 픽스처
필요), `fieldsetIdentity`(28), `schemaSelectRow`/`schemaRadioRow` 잔여 조합,
`boardPage` 배너 분기다.

본 구간에서 추가한 것은 전부 동작 단언 테스트다: 위젯 표시 상태(오류 상태·빈 옵션
시맨틱·stacked 레이아웃 opt-in·타입→위젯 디스패치), 미도달 위젯 3종의 계약(제출 컨트롤
부재·이스케이핑), 라벨 fallback의 구별성, 뷰모델 저하 경로(읽기 실패 시 빈 맵·거부된
저장의 제출값 에코), `newApp` seam 배선 누락, 에이전트 그리드의 전수 순위 매핑.

**권고**: AC-WCR-063의 90% 기준은 생성 코드를 분모에서 제외하거나
(`-coverpkg` 필터 / `//go:generate` 산출물 제외), 손으로 쓴 코드 기준 90%로
재정의하는 편이 정직하다. 현재 손으로 쓴 코드는 91.1%로 이미 그 기준을 충족한다.
이 재정의는 SPEC 본문 변경이므로 manager-spec 소관이다.

## §E.3 Run-phase Audit-Ready Signal

run_status: partial (M1-M6 완료, M7 부분 — AC-WCR-063 미충족)
run_complete_at: 2026-08-08
run_commit_sha: c9406abf5 (M1-M3) / a9995c80b (M6) / 62629d7b0 (M4) / e5efa49da (M5) / ced853c17 (execution_mode)
ac_pass_count: 23 (M1-M3 13 + M4-M6 10)
ac_fail_count: 1 (AC-WCR-063 커버리지 73.5% — §E.2.3 구조적 상한 80.7%)
ac_pass_with_debt_count: 2 (AC-WCR-023 M1-M3 / AC-WCR-030 grep 기대치)
ac_vacuous_count: 2 (AC-WCR-061 / AC-WCR-062 — yaml·템플릿 편집 0건)
preserve_list_post_run_count: 5/6 (plan.md §A.3). 6번 항목("미렌더 섹션 중
  git_strategy 외 나머지의 FieldDef 무접촉")은 의도적 편차다: REQ-WCR-022가
  `harness.*`를 닫힌 집합 전환 대상으로 명시하므로 harness FieldDef 8종의
  Type/Options/Validate를 변경했다. 이름·섹션·영속화 경로는 무변경이고 harness
  섹션은 여전히 미렌더다. 나머지 5항목(workflow.yaml 키, config struct+접근자,
  `parseSchemaForm` bool 분기, glmkey.go 계약, board.*)은 diff 0라인으로 실측 확인.
new_warnings_or_lints_introduced: 0 (`golangci-lint run --timeout=5m` → `0 issues.`, `go vet ./...` → exit 0)
cross_platform_build: darwin/arm64 PASS, windows/amd64 PASS
total_run_phase_files: 15 (M1-M3) + 23 (M4-M6 + execution_mode + 커버리지)
m1_to_mN_commit_strategy: 밀스톤당 1커밋 (M1 / M2 / M3 / M6 / M4 / M5 + execution_mode 수정 + 커버리지), push 없음
full_suite: `go test ./...` → `internal/template` 3건 실패
  (`TestLateBranchTemplateMirror/spec-assembly.md`,
  `TestRuleTemplateMirrorDrift/spec-workflow.md`,
  `TestSanitizedPairParity/main-checkout-branch-guard.md`). 전부 PRE-EXISTING이며
  신규 실패 0건. 귀속 근거: 세 테스트의 입력 파일 3종과 그 템플릿 미러 3종 모두
  `git diff 36f9bfea4..HEAD` 0라인 + 워킹트리 무변경 — 즉 M1-M3 베이스라인과
  바이트 동일하므로 판정이 달라질 수 없다

## §E.4 Sync-phase Audit-Ready Signal

sync_status: ready
sync_complete_at: 2026-08-13
sync_commit_sha: pending-backfill-sync
  (self-referential-hazard workaround per spec-frontmatter-schema.md § Status Transition
  Ownership Matrix D3 — a commit cannot know its own SHA until after it lands. manager-git
  backfills the real SHA in the Late-Branch follow-up commit after the sync PR merges.)
changelog_entry_position: top of `## [Unreleased] ### Added` in CHANGELOG.md
  (SPEC-WEB-CONSOLE-REDESIGN-001, immediately above SPEC-CHAIN-CORE-001)

frontmatter_status_transitions:
  spec.md: in-progress → completed (single sync-commit transition; the
    `implemented` intermediate is merged into the same sync commit per the
    3-phase close contract). `updated:` refreshed 2026-08-08 → 2026-08-13.
    This is the ONLY YAML-frontmatter artifact — acceptance.md / plan.md use
    the markdown-header convention (no YAML frontmatter).
  acceptance.md: no frontmatter (markdown-header convention). Body change
    LIMITED to the AC-WCR-063 PASS-WITH-DEBT marker + §E.2.3 rationale pointer
    (no other body content touched, per Status Transition Ownership Matrix
    forbidden-ownership-crossings).
  plan.md: no frontmatter (markdown-header convention). Body UNCHANGED.

canary_compliance_check:
  b12_self_test_a_pre_emission_grep: PASS — `grep -c 'SPEC-WEB-CONSOLE-REDESIGN-001' CHANGELOG.md`
    returned 0 before emission (no duplicate entry from parallel BATCH-SYNC sessions).
  b12_self_test_b_ac_count_match: N/A — acceptance.md uses markdown-header AC markers
    (no `### AC-XXX` headings / no `| AC-XXX |` table cells); the AC set is referenced by
    the explicit enumeration in §E.3 below. The CHANGELOG entry cites 28 AC identifiers
    (25 numbered, 23 PASS / 1 FAIL→PASS-WITH-DEBT / 2 PASS-WITH-DEBT / 2 vacuous).
  b12_self_test_c_file_path_verification: PASS — every path claimed in the CHANGELOG entry
    exists (`ls .moai/specs/SPEC-WEB-CONSOLE-REDESIGN-001/{spec,plan,acceptance,progress}.md`
    + `ls docs-site/content/{en,ko,ja,zh}/{advanced/moai-web-console,cli-reference/web}.md`
    + `ls CHANGELOG.md` + `ls internal/web/...` tree).

sync_artifacts_authored:
  - CHANGELOG.md — SPEC entry added under `## [Unreleased] ### Added`, en per
    `git_commit_messages: en`. Covers M1-M6 + execution-mode correction + AC-WCR-063
    PASS-WITH-DEBT + docs-site 4-locale refresh.
  - docs-site/content/{en,ko,ja,zh}/advanced/moai-web-console.md — 6-tab → 9-tab
    structure, profile-card → profile-bar dedup, widget policy + GLM honesty badge
    summarized (4-locale same-PR obligation per docs-site i18n rules).
  - docs-site/content/{en,ko,ja,zh}/cli-reference/web.md — 6-tab → 9-tab refresh in the
    Console Screens section (4-locale).
  - spec.md frontmatter — `status: in-progress → completed`, `updated: 2026-08-13`.
  - acceptance.md body — AC-WCR-063 PASS-WITH-DEBT marker + §E.2.3 pointer (sole body
    change; everything else forbidden per ownership matrix).
  - progress.md §E.4 — this section.

ac_disposition_sync:
  ac_pass_count: 23 (unchanged from §E.3)
  ac_pass_with_debt_count: 3 (AC-WCR-023 + AC-WCR-030 from §E.3 + AC-WCR-063 newly
    marked PASS-WITH-DEBT in this sync per user approval 2026-08-13)
  ac_fail_count: 0 (AC-WCR-063 moved from FAIL → PASS-WITH-DEBT)
  ac_vacuous_count: 2 (AC-WCR-061 / AC-WCR-062 — unchanged)

evidence_quality_gate:
  source: orchestrator trust-but-verify (this tree, this run), NOT re-executed by
    manager-docs — the orchestrator already ran the commands; sync-phase consumes the
    §E.3 attribution per the attributable diff-check doctrinal switch
    (SPEC-SYNC-PARALLEL-DOCS-001 A9 — snapshot-key attribution, no re-execution).
  go_build: `go build ./internal/web/... ./internal/cli/...` → exit 0 (orchestrator).
  go_vet: `go vet ./internal/web/...` → exit 0 (orchestrator).
  go_test_internal_web: `go test ./internal/web/...` → `ok 1.728s` (orchestrator).
  go_test_repo_wide: `go test ./...` → 3 PRE-EXISTING failures in `internal/template`
    (`TestLateBranchTemplateMirror/spec-assembly.md`,
    `TestRuleTemplateMirrorDrift/spec-workflow.md`,
    `TestSanitizedPairParity/main-checkout-branch-guard.md`). All three are
    PRE-EXISTING on the baseline and unrelated to this SPEC — the same three fail
    on the pre-PR-#1410 baseline (`git diff 36f9bfea4..HEAD` 0 lines on those
    fixtures + their template mirrors). NOT a SPEC regression; attributed, not
    re-executed.
  coverage: AC-WCR-063 PASS-WITH-DEBT — 73.5% measured, structural ceiling 80.7%,
    hand-written 91.1%. See §E.2.3 for the full coverage-analysis evidence.

docs_site_audit:
  moai_web_console_md: 4-locale refresh applied (en/ko/ja/zh). Pre-sync state described
    "six tabs: Identity · Language · LLM · 3rd Party LLM · Agents · Report" + a standalone
    Profiles card. Both surfaces were STALE: the redesign shipped a 9-tab structure
    (Identity · Language · LLM · 3rd Party LLM · Workflow · Git & Worktree · Audit ·
    Agents · Report) and moved add/rename/delete into the profile bar (REQ-WCR-040).
    Each locale's native register preserved (no calques); code identifiers verbatim.
  cli_reference_web_md: 4-locale refresh applied — "six tabs" → "nine tabs" in the
    Console Screens section, profile-bar dedup noted.
  nav_config_untouched: _meta.yaml / main.yaml / menu.html NOT modified (no page moved;
    the page paths and weights are unchanged).
  images: NOT refreshed in this sync — the existing screenshots (`web-console-overview.png`,
    `web-console-switch.png`, `web-console-llm-tab.png`) still show the pre-redesign
    6-tab layout. This is recorded as adjacent docs debt for a follow-up screenshot
    refresh (image production is out of scope for a markdown-only sync commit).

debt_recorded:
  - AC-WCR-063 denominator redefinition — SPEC-body change, manager-spec's domain.
    Recorded here and in the CHANGELOG entry; the AC marker in acceptance.md carries
    the §E.2.3 pointer. NOT resolved in this sync.
  - docs-site screenshots — pre-redesign 6-tab images still shipped. Adjacent docs
    debt for a follow-up image refresh (out of scope for sync-phase).

repo_local_pr_policy: Route B (PR-mandatory, this repo's `enforce_admins: true`
  branch-protection override). manager-git creates the branch + opens the PR;
  this sync-phase authored only the working-tree artifacts above (uncommitted).

working_tree_hygiene_preserved:
  - `.moai/config/sections/llm.yaml` — locally MODIFIED (`team_mode: glm`, a
    factory-mode runtime value). PR #1499 (origin/main, `54d748ddf`) already
    untracks + gitignores it. NOT staged, NOT committed, NOT modified by this
    sync-phase.
  - `.moai/reports/diagram-i18n-demo.html` and `.moai/reports/diagram-styles-gallery-v2.html`
    — untracked report artifacts. NOT staged, NOT committed.
  - The orchestrator's behind-resolution (manager-git) handles the PR #1499
    llm.yaml untracking; manager-docs operates on the doc surface only.
