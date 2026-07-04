# Progress — SPEC-WEB-CONSOLE-011

Lifecycle tracking for the 3-phase plan → run → sync flow. §E carries the audit-ready signals.

## §E.1 Plan-phase Audit-Ready Signal

- **Phase**: plan (artifacts authored)
- **Tier**: L (5-artifact: spec.md + plan.md + acceptance.md + design.md + research.md)
- **Status**: `draft`
- **plan_complete_at**: 2026-07-03
- **plan_status**: audit-ready (plan-auditor Phase 0.5 pending)
- **Version**: 0.2.0 — 사용자 4결정 확정 (2026-07-03, AskUserQuestion 직접 선택): M2 10섹션 전면 확장 + M3 쓰기 전면 지원으로 개정. v0.1.0의 "선택 확장 / staged write"는 supersede. → **0.2.1** — plan-audit iter-1 D1-D6 + D8d/D9d 정정 (REQ 총계 42→49 오산 정정 포함).
- **Artifacts created (v0.2.0 amended)**:
  - `spec.md` — 12-field canonical frontmatter (+ tier: L, era: V3R6), HISTORY (0.1.0 + 0.2.0 + 0.2.1), 49 GEARS requirements (REQ-WC11-001..005, 010..019, 020..029, 030..034, 040..046, 050..053, 060..062, 070..074; v0.2.0 개정 ID: 001/002/022/025/026/050/053; v0.2.1 정정 ID: 016/032/033), Exclusions with 7 `### Out of Scope —` H3 sub-headings.
  - `plan.md` — §A Context (+§A.5 PRESERVE), §B Known Issues B1..B13 (B10 i18n 인플레이션, B11 파생 카운트, B12 섹션 키 미실측, B13 frontmatter 견고성), §C Pre-flight 11-command (섹션 키 열거 + 미러 재확인 포함), §D Constraints, §E Self-Verification pointer, §F Milestones M1 → M2a → {M2b, M3} → M4/M5/M6 (의존 그래프 명시), §G Anti-Patterns AP-1..11, §H Cross-References.
  - `acceptance.md` — 52 AC (AC-WC11-001..005, 010..019, 020..029, 030a/030b..034, 040..046, 050..054, 060..063, 070..074), 7 Given-When-Then, 8 edge cases, Definition of Done (agent body byte-무변경 + 미러 3종 parity 포함).
  - `design.md` — §A yaml.Node patch seam (+upsert 확장, §A.3 10섹션 라우팅 표, §A.4 정규화 리스크), §B effort opaque-node 결정 (Option A — 전면 쓰기 하 유지 확인), §C.1 frontmatter patch layer + template-mirror policy (live-only + 지속 경고), §C.2 workflow_agents typed 표면, §D 검증 배치, §E 보드 데이터 흐름, §F segment SSOT 테스트, §G risks.
  - `research.md` — §A-§H survey (wf_d19d522a-d39) + §I 결정 변경 기록 (사용자 직접 선택, AskUserQuestion, 2026-07-03) + plan-phase 추가 실측 (10섹션/미러/workflow_agents grep 0/미지명 섹션 4종).
- **SPEC ID self-check**: `decomposition: SPEC ✓ | WEB ✓ | CONSOLE ✓ | 011 ✓ → PASS` (canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`).
- **Scope SSOT**: 사용자 승인 4결정 (2026-07-03) — spec.md §1.1.
- **Plan-auditor gate**: iter-1 **PASS-WITH-DEBT 0.85** (6 SHOULD-FIX + 4 MINOR, 전부 mechanical) → D1-D6 + D8d/D9d 정정 완료 (v0.2.1); D10d(house-convention orphan AC)는 noticed-only 잔존.
- **Implementation Kickoff Approval**: pending (plan-to-implement human gate — run-phase 진입 전 필수).

## §E.2 Run-phase Evidence

> **범위 주의**: 본 run-phase 위임은 **M1 + M2a만** 커버한다 (orchestrator 위임 명시). M2b/M3/M4/M5/M6은 후속 위임 소관 — 아래 evidence는 M1/M2a 바인딩 AC(AC-WC11-001..005, 017)에 국한된다. 작업 트리 = L1 runtime worktree (`worktree-agent-a20044e6263f21afb` 브랜치, base 2510c2775 = plan-phase 커밋; main과의 차분은 무관 SPEC-HANDOFF-CTXGUIDE-001 plan 커밋 1개뿐).

### Pre-flight 결과 (plan.md §C subset — 1, 3, 4, 5, 7, 8)

| # | 항목 | 명령 | 실측 결과 |
|---|------|------|----------|
| C-1 | baseline | `git branch --show-current && git rev-parse HEAD` | worktree 브랜치 base `2510c2775` (plan-phase 커밋); main checkout HEAD는 `b303d9916` (병렬 세션 미push 커밋 — 무접촉) |
| C-3 | cross-platform build | `go build ./... && GOOS=windows GOARCH=amd64 go build ./...` | `BUILD_OK` + `WIN_BUILD_OK` (worktree 기준 재측정 포함) |
| C-4 | lint baseline | `golangci-lint run --timeout=3m` | `0 issues.` (baseline 0) |
| C-5 | 앵커 재검증 | content-token grep | `REQ-WC-012` @ server.go:10, `REQ-WC3-007` @ projectconfig.go:159(드리프트 +1행), `isValidProfileName` @ profile.go:118/126, `REQ-WC-009` @ app.go — 전부 실존 |
| C-7 | 10섹션 키 열거 | `cat .moai/config/sections/{8섹션}.yaml` | 8파일 전부 실존. harness.yaml은 최상위 키 2개(`harness:` + `learning:`) — sectionRootKeys에 반영. db.yaml 8키 = system 5(enabled, dir, auto_sync, migration_patterns, engine) + interview 3(orm, multi_tenant, migration_tool) |
| C-8 | no-Save-path 전제 기계 검증 | `internal/config/manager.go` `Save()` 실독 (L166-228) | **전제 성립**: Save()는 정확히 6파일만 기록 — user/language/quality/git-convention/git-strategy(dirty-flag)/llm. 8개 seam 섹션(workflow, harness, ralph, research, feedback, observability, security, db)은 Save 경로 부재 (typed 읽기 struct는 일부 존재하나 쓰기 경로 없음 — B12 해소, 불일치 0건) |

### M1 — Foundation (커밋 81f8ee11e)

- **Scope contract supersede**: server.go 패키지 doc(구 REQ-WC-012) + projectconfig.go @MX:REASON(구 REQ-WC3-007)을 10섹션 계약으로 재작성. `grep -rn "SPEC-WEB-CONSOLE-011" internal/web/server.go internal/web/projectconfig.go` = **2 매치** (AC-WC11-001 충족).
- **라우팅 SSOT 신설**: `internal/settings/sectionroute.go` — RouteForSection (typed 6 / seam 8 / statusline / 미등재 기본 거부 zero-value) + SeamSections() + ExcludedSections() (12종).
- **Guard test 신설**: `internal/web/scope_contract_test.go` — TestScopeContractTenSections + TestScopeContractExclusions (제외군 12종 + 임의 미등재 4종 거부).
- **yamlpatch seam (TDD RED→GREEN)**: `internal/settings/yamlpatch/` — KeyEdit/PatchFile (스칼라 교체 + 명시 경로 upsert + Style/Tag 보존 + indent 검출(4 기본/2 db형) + temp+rename 원자 기록, 삭제 미지원). **RED 관측**: stub 상태에서 6개 테스트 FAIL (`yamlpatch: not implemented`) → 구현 → GREEN. 실제 workflow.yaml fixture round-trip: 1-line diff + 주석/`team.patterns`/effort 전량 보존 + upsert additive-only.
- frontmatter 전환: spec/plan/acceptance `status: draft → in-progress` (M1 커밋 동승).

### M2a — 8섹션 seam persistence 인프라

- **per-section 쓰기 라우팅**: `internal/settings/sectionwrite.go` — WriteSectionViaSeam (RouteSeam 아닌 섹션 전부 오류 거부 + 파일 무접촉, sectionRootKeys로 섹션 파일 밖 최상위 키 주입 차단, db는 validateDBEdit로 REQ-WC11-019 3/5 분리 강제).
- **golden fixture ×8**: `internal/settings/testdata/sections/*.yaml` (실파일 사본 8개) + TestYAMLPatchGoldenSections — 섹션별 스칼라 1개 편집 후 (i) 1-line diff (ii) 주석 전량 보존 (iii) 키 순서 보존 검증. `go test -run TestYAMLPatchGolden ./...` **8/8 PASS** (AC-WC11-017).
- **yaml.v3 공백-only 정규화 실측 + golden 고정 (design.md §A.4 사전 승인 경로)**: 재인코딩이 매핑 항목-사이 빈 줄을 제거함을 관측 — 8섹션 중 항목-사이 빈 줄을 가진 **security.yaml / db.yaml 2개에서만** 발생. 주석/키/값/순서는 전량 보존(공백-only) → blocker 아님. golden 테스트에 `blankNormalized` 플래그로 범위 고정: 빈 줄 제거 외 라인 변동 0 + 라인 추가 0 + 비-빈 줄 1-line diff. 나머지 6섹션(workflow, harness, ralph, research, feedback, observability)은 strict byte-level 1-line diff 유지.
- **db 3/5 분리 (data-level)**: DBEditableKeys()={orm, multi_tenant, migration_tool} / DBSystemKeys()={enabled, dir, auto_sync, migration_patterns, engine} — 실측 db.yaml 헤더 주석("5 system-fixed + 3 interview-input")과 정합, 드리프트 가드 테스트(TestWriteSectionViaSeamKeySplitConstantsMatchFixture) 포함. system 키/미지명 키/중첩 경로 write 시도 → 오류 + 파일 불변 (AC-WC11-019 데이터 계층 준비).
- harness `learning:` 최상위 키 seam 기록 검증 (TestYAMLPatchGoldenHarnessLearningRoot).
- 커버리지 보강: yamlpatch 오류 경로 직접 테스트 (atomicWrite stat/read-only-dir, encode unknown-kind, 매핑/시퀀스 대상 거부) → 77.9% → **86.3%**.

### M2b — 10섹션 fieldsets + i18n (worktree 브랜치, rebased base 5d2d18f3b 위)

- **스키마 확장 (REQ-WC11-010/011/012/016/019)**: `internal/settings/schema_sections.go` — 신규 편집 필드 **163개** (git-strategy 57 [3 mode-profile 루프 생성 + top-level 4 + 조건부 8], llm 안전 키 10 [performance_tier select + claude_models×3 + glm.models×6; mode/team_mode 제외], quality 잔여 typed 16, workflow 21, harness 14, ralph 19, research 12, feedback 1, observability 7, security 3, db 3). PersistKind 2종 신설: `PersistTypedSection`(git_strategy/llm/quality — config 매니저 LoadRaw→apply→SetSection→Save) + `PersistSeam`(8섹션 — WriteSectionViaSeam 전용). 렌더 섹션 11개 신설(SectionQualityExtras 포함 — 기존 Project fieldset 9-필드 구성과 충돌 방지 분리).
- **영속화 디스패처**: `sectionapply.go` ApplySchemaEdits — typed 3섹션 per-key applier switch(전 분기 전수 round-trip 테스트로 커버) + seam 파일별 단일 PatchFile. git-strategy dirty-flag 격리 검증(비-git-strategy 편집 시 git-strategy.yaml byte 불변 — TestApplySchemaEditsGitStrategyDirtyFlagIsolation).
- **제네릭 읽기 seam**: `sectionvalues.go` SchemaCurrentValues(전 확장 필드 + read-only 표시 키) + RawBlockValues(REQ-WC11-062 raw view 블록 표시용 재직렬화).
- **웹 배선**: `schemaform.go` 스키마 주도 제네릭 파서(hand-wiring 0 — FieldDef가 SSOT; bool `__present` companion, int/float/select 검증 → atomic reject 합류) + handlers.go handleSave/handleIndex/projectView 배선 + app.go injectable seam 3종.
- **제네릭 fieldset 렌더**: fieldsets.templ `fieldsetSchemaSection` + 위젯 5종(text/number/toggle/select/read-only) + `schemaRawBlock`(collapsed `<details>`, input 컨트롤 0). **필드 라벨은 key-chip 기술 식별자로 렌더, data-i18n 미방출** — i18n.js 기존 계약("Field identifiers stay in English as code chips and are NOT translated")과 정합. root.templ에 11 fieldset 조립. templ generate로 _templ.go 재생성 (v0.3.1020 = go.mod 일치).
- **파생 카운트 라벨 (REQ-WC11-053 전략 선적용, B11)**: 기존 5개 fieldset의 하드코딩 "N fields" 리터럴 전부 제거 → `sectionCount(len(settings.SectionFields(...)))` 파생 + `count.fields` 단위 접미 i18n 키. `grep -nE '"[0-9]+ fields?"' fieldsets.templ` 코드 리터럴 0 (잔여 1건은 갱신된 주석에서도 제거 완료).
- **i18n ×4 (REQ-WC11-015/061)**: 신규 data-i18n 키 25종(sec.<11섹션>.title/.desc ×22 + count.fields + ro.note + raw.note)을 en/ko/ja/zh 4-locale 전부에 추가 (수작업 번역, blind sed 없음 — B10). TestDataI18nKeysSubsetOfDictionary(렌더된 키 ⊆ 사전) + TestI18nKeySetParity 전부 GREEN.
- **기존 테스트 계약 갱신 3건**: (i) settings TestSchemaFieldNameSet/TestSchemaSixSections — 34-총계 pin 제거, 기존 34필드 잔존 floor로 전환(B11); (ii) web TestWebRendersSchemaFieldSet — 34 pin 제거, 전 필드 집합 검증으로 강화; (iii) TestI18nKeySetParity — M2b key-chip 필드(PersistSeam/PersistTypedSection) 제외 스코프 명시(사유 주석 포함); (iv) projectnested_parse_test count.project → count.fields 마커.
- **AC-WC11-004 행동 완결**: TestSaveWorkflowRoutesThroughSeam — POST /save의 workflow 스칼라 편집이 seam 경유로 정확히 1라인만 변경 + 주석/team.patterns 보존을 end-to-end 검증.
- **신규 웹 AC 테스트**: TestSchemaSectionsRenderSmoke(016 GET + 014 placeholder + 018 렌더 + 063 rawview input-0), TestSaveLLMModeReadOnlyIgnored(013), TestSaveInvalidSchemaValueRejected(012 oneof 4xx + EC-2), TestSaveDBKeySplitWebLayer(019), TestSaveExcludedSectionForgedPost(EC-8/018), TestSaveSchemaSmokeAllSections(016 저장 10/10 + git-strategy/llm 포함 11필드).
- **커버리지**: settings 75.4%→**90.9%**(전수 round-trip 테스트 TestApplySchemaEditsAllFieldsRoundTrip — 163필드 단일 호출 왕복), yamlpatch 86.3% 유지, web 70.9%(baseline 73.0% 대비 −2.1pp — templ generate 산출 코드가 분모 팽창; 신규 비생성 코드는 신규 테스트가 커버).
- lint: `golangci-lint run` **0 issues** (unused parseFloatValue 제거 후).

### M3 — Agent Settings 4표면 전면 쓰기 (rebased base 002446611 위)

- **(b) role_profiles 28필드 + (d) workflow_agents 14필드** (REQ-WC11-020/022/024/070..073): `schema_sections.go` agentSettingsFields — SectionAgentSettings 신설, 전부 PersistSeam(workflow.yaml). 옵션은 **v4manifest exported 상수 재사용**(EffortLow../ModelInherit../IsolationNone.. — 리터럴 재선언 0, REQ-WC11-024/072). role_profiles effort는 seam opaque-node 패치 — `RoleProfileEntry`에 Effort 필드 미추가(TestRoleProfileEntryHasNoEffortField reflection 가드, AC-WC11-023; 제네릭 파서/렌더가 그대로 처리해 웹 신규 코드 0줄로 두 표면 커버).
- **workflow_agents typed 읽기** (REQ-WC11-070/071): `internal/config/types.go` WorkflowAgentEntry{Model, Effort} + WorkflowConfig.WorkflowAgents map — 블록 부재 시 nil 무오류(TestWorkflowAgentsAbsentBlockZeroValue), 7-purpose 파싱(TestWorkflowAgentsTypedLoad). 쓰기는 M1 seam upsert 전용(AP-11 — TestWorkflowAgentsUpsertGolden: 블록 부재 fixture 최초 기록 additive-only + 주석/patterns/role_profile_keys 보존, GWT-7).
- **(c) frontmatter patch layer** (REQ-WC11-025/027..029): 신규 패키지 `internal/settings/agentfm` — 첫 두 `---` 구분선으로 frontmatter만 yaml.Node 패치, **body는 파싱조차 없이 원본 bytes 재조립**(구조적 byte 보존). 연산: model/effort 교체 + upsert + **effort 키 삭제**(EC-7 "(absent)" 복귀 — seam에 없는 유일한 삭제 연산). Idempotency 2회 patch byte-identical + body bytes.Equal 기계 검증(TestPatchIdempotencyAndBodyPreserved, AC-WC11-027). live 파일 전용 — template dual-write 0 (`grep "internal/template/templates"` in agentfm/web 경로 = 0, AC-WC11-029 iii).
- **웹 배선**: `internal/web/agentfm.go` — v4manifest **직접 import 검증**(AC-WC11-024 grep ≥1 실참조), no-change 제출 필터(불필요 재직렬화 방지), agent 이름 경로 조작 가드. fieldsetAgentFrontmatter templ 카드: 지속 경고(agentfm.warn, REQ-WC11-028) + llm 교차 참조 + dynamic-workflows.md taxonomy 참조(REQ-WC11-026) + "(keep current)"/"(absent)" select. i18n 신규 10키 × 4 locale(TestAgentFMWarnI18nParity).
- **AC 테스트 (전부 GREEN 1-pass)**: 020(4표면 렌더 — llm/role_profiles 7/frontmatter/purposes 7), 022(role_profile 편집 1-line diff + 주석/patterns 보존), 023, 024+071(superhigh/gpt5/ultra → 400 + 파일 불변), 025(GWT-6 round-trip + 재렌더 ✓ 반영), 029(i 거부/ii effort 키 미주입), 072(upsert golden), 027(agentfm 패키지).
- **REQ-WC11-074 (rule + mirror work item)**: live `.moai/config/sections/workflow.yaml` + template mirror에 workflow_agents 기본 블록(7 purposes — taxonomy 권장값) 추가; dynamic-workflows.md live+mirror에 "Config surface" SSOT 문단(config 블록 = 기본값 SSOT, per-script 리터럴 = override) 추가. **§25 neutrality**: 미러 추가분에 SPEC ID/내부 날짜 0 (기존 `{SPEC-ID}` 플레이스홀더 변수 1건은 선재 허용 클래스); `make build` **exit 0** (재임베드 + catalog 재생성 — 터치한 config/rule은 catalog 해시 대상 아님이라 catalog.yaml 무변경); `go test ./internal/template/` PASS (neutrality/leak 가드).
- **스코프 가드**: 광역 gofmt가 선재 드리프트 파일(projectnested_error_test.go)을 포맷한 것을 감지 → `git checkout`으로 원복 (Scope Discipline — 해당 파일 무접촉 유지).
- 커버리지: agentfm 85.3%(오류 경로 보강 후), settings 91.1%, config 80.2%(선재 대형 패키지 — 본 위임 추가 실행문 0, struct 필드 + 테스트만).

### 발견 사항 / 특이 기록

1. **yaml.v3 blank-line 정규화** (위 M2a 항 — design.md §A.4 예상 리스크의 실측 확정, 허용 범위 내 golden 고정).
2. **선재 실패 (내 변경 무관)**: `go test ./...`에서 `internal/cli` `TestRunHookEvent_ReadInputError` FAIL (nil pointer panic @ coverage_test.go:77). **base 시점 선재** — 본 worktree에서 internal/cli는 무변경(`git status --porcelain internal/cli/` = empty)이고 M1/M2a 커밋은 해당 패키지 무접촉. main checkout에는 병렬 세션(SPEC-CLI-SUBPKG-SPLIT-001)의 internal/cli/coverage_test.go 미커밋 수정이 존재 — 해당 세션 소관. PRESERVE 제약상 무접촉 유지.
3. **선재 gofmt 드리프트 (내 변경 무관)**: `gofmt -l`이 internal/web/handlers.go + internal/web/projectnested_error_test.go 지적 — 양쪽 모두 git 무변경 base 파일. 내 신규/수정 파일은 전부 gofmt clean.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-04
run_commit_sha: "af4bdf245 (M1) + 5d2d18f3b (M2a) + 002446611 (M2b — rebased SHA) + <M3 — 본 progress.md 동승 커밋>"
run_status: in-progress — M1 + M2a + M2b + M3 완료; M4/M5/M6 후속 위임 대기
ac_pass_count: 28  # M1/M2a/M2b 15 + M3: 020, 021(llm tier round-trip — M2b 스모크 포함), 022, 023, 024, 025, 026, 027, 028, 029, 070, 071, 072, 073, 074 중 검증 완료분 (021은 TestSaveSchemaSmokeAllSections의 llm 필드로 커버)
ac_fail_count: 0
ac_partial_notes: "AC-WC11-011은 '노출 ∪ 명시 제외 목록' 파티션 테스트(TestQualityKeyPartition)로 충족 — lsp_quality_gates/lsp_integration/principles/cycle_type_routing은 명시 제외(form UI 부적합 대형 정책 블록). AC-WC11-015는 신규 data-i18n 키 25종 ×4 locale로 충족 — M2b 필드 라벨은 key-chip 기술 식별자(비번역 계약)라 per-field 키 비대상. AC-WC11-005 allowlist 불변(신규 NewConfigManager/.Save( 0 — 신규 typed 경로는 internal/settings 소재)"
preserve_list_post_run_count: 0   # PRESERVE 위반 0 — statusline/app.go@MX:NOTE/병렬 세션 산출물/agent 파일 전부 무접촉
l44_pre_commit_fetch: "n/a — L1 runtime worktree 격리 실행 (전용 브랜치, push 금지 지시); landing은 orchestrator gate 검증 후 수행"
l44_post_push_fetch: "n/a — push 미수행 (지시 사항)"
new_warnings_or_lints_introduced: 0   # golangci-lint baseline 0 → 이후에도 "0 issues."
cross_platform_build:
  darwin_arm64: PASS (go build ./... → BUILD_OK)
  windows_amd64: PASS (GOOS=windows GOARCH=amd64 go build ./... → WIN_BUILD_OK)
  go_vet: PASS (VET_OK)
coverage:
  internal_settings: "90.9% (M2b 확장 후 — 163필드 전수 round-trip 테스트 포함)"
  internal_settings_yamlpatch: "86.3%"
  internal_web: "70.9% — baseline 73.0% 대비 −2.1pp (templ generate 산출 코드 분모 팽창; 신규 비생성 코드는 신규 AC 테스트가 커버)"
total_run_phase_files: 30   # M1(11) + M2a(12) + M2b: settings 4신규+2수정+testdata 3, web 2신규(go)+1신규(test)+3수정(go)+2 templ+2 _templ 재생성+3 테스트 수정+i18n.js
m1_to_mN_commit_strategy: "마일스톤별 path-limited 커밋 (M1: af4bdf245 / M2a: 5d2d18f3b — rebased SHA / M2b: 후속 커밋), 전용 worktree 브랜치, push/landing은 orchestrator 소관"
known_preexisting_failures: "internal/cli TestRunHookEvent_ReadInputError (base 선재, 병렬 세션 SPEC-CLI-SUBPKG-SPLIT-001 도메인, 무접촉); gofmt 드리프트 2파일 (base 선재, 무접촉)"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
