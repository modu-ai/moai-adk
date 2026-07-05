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

### M4 — Profile CRUD (repro-test-first 보안 수정 + CRUD UI) — 커밋 133a3c9d5 / b66580ffc / d436aa80c

> **범위**: 본 후속 위임은 **M4만** 커버한다 (M2b=002446611, M3=ab555742f/4f5b3fbbb 는 앞선 위임에서 landing). base = 4f5b3fbbb (M3 tip). cycle_type=tdd, Reproduction-First.

**Step 1 — RED (AC-WC11-030a, 커밋 133a3c9d5)**: 가설(웹 쓰기 경로가 `__profile`/`?profile=` 를 `isValidProfileName` 없이 `WritePreferences`→`MkdirAll` 로 흘려 path traversal 로 profile store 밖 디렉터리 생성)을 repro test 로 기계 검증. 수정 前 트리에서 **FAIL 관측 (verbatim)**:

```
--- FAIL: TestProfileNameTraversal (internal/web)
    profile_traversal_test.go:75: POST /save with __profile="../../moai-repro-escaped" status = 200, want 4xx
    profile_traversal_test.go:83: traversal created a directory outside the profile store: .../001/sub1/moai-repro-escaped
--- FAIL: TestProfileNameTraversal (internal/profile)
    traversal_test.go:42: WritePreferences("../../moai-repro-escaped") = nil, want error (invalid profile name)
    traversal_test.go:47: WritePreferences created a directory outside the profile store: .../001/sub1/moai-repro-escaped
```

가설은 **반증되지 않고 확정**됨 (verification-claim-integrity §1.1 surface 3). RED 커밋(test-only, failing)이 fix 커밋에 선행 — `git log` 순서 증거: 133a3c9d5(RED) → b66580ffc(GREEN).

**Step 2 — GREEN (AC-WC11-030b/031/034, 커밋 b66580ffc)**: design.md §D.1 defense-in-depth 2중 배치. (1차) 웹 경계 `handleSave`: `selected` 해석 직후 `profile.IsValidProfileName` 검증→400, write seam 도달 전 차단. (2차) `WritePreferences` 내부: `os.MkdirAll` 이전 `isValidProfileName` 가드→오류. `profile.IsValidProfileName` exported wrapper 신설(REQ-WC11-031, 재선언 금지). `""`/`"default"` 특수명 통과 확인(회귀 0). repro test GREEN 전환: `__profile=../../x` → 400 + escaped dir 미생성.

**Step 3 — CRUD UI (AC-WC11-032/033, 커밋 d436aa80c)**: 2개 신규 POST 라우트(`/profile/create`, `/profile/delete`) + `profileManager` Templ 컴포넌트 + 4-locale i18n(6키 ×4). create=`GetProfileDir`+`os.MkdirAll`(env 무접촉 — EnsureDir 의 `CLAUDE_CONFIG_DIR` os.Setenv 부작용 회피, 장수 서버+테스트 격리 근거 profile_crud.go 주석). delete guards: default(기존 guard) + active(cfg.ProfileName/GetCurrentName — 웹 경계 NEW 4xx guard, live 는 stderr 경고 후 진행). switch=기존 GET `/?profile=<name>` 재사용. `handleIndex`→`buildIndexView` 추출(전체 페이지 재렌더 재사용). PRESERVE 준수: app.go @MX:NOTE(REQ-WC-009) CSRF 블록 무접촉(routes() 라우트만 추가), statusline/agent body 무접촉.

**M4 AC 매트릭스 (전 [B] PASS)**:

| AC | 검증 | 결과 |
|----|------|------|
| AC-WC11-030a | RED 커밋(133a3c9d5) → fix 커밋(b66580ffc) 순서 + RED verbatim 출력 | PASS — repro FAIL 관측 후 커밋 순서 증거 |
| AC-WC11-030b | `go test -run TestProfileNameTraversal ./internal/web/ ./internal/profile/` | PASS (양 패키지 ok) — `../../x` → 4xx + 디렉터리 미생성 |
| AC-WC11-031 | `grep isValidProfileName internal/web/ internal/profile/preferences.go` | PASS — 실호출 preferences.go:159 + web 경계 `profile.IsValidProfileName` |
| AC-WC11-032 | `TestProfileCRUDFlow` (create→list→switch→delete) + i18n 4-locale | PASS — 전 흐름 + `TestProfileCRUDI18nKeys` 6키×4 |
| AC-WC11-033 | `TestProfileDeleteGuards` (default + active keepme 삭제) | PASS — 양쪽 4xx + 디렉터리 잔존 |
| AC-WC11-034 | `TestProfileCreateInvalidName` + `TestProfileDeleteInvalidName` (빈/예약/traversal) | PASS — 4xx + MkdirAll side effect 0 |

**커버리지 (M4 신규 함수)**: `createProfileDir` 100% / `activeProfileName` 100% / `renderProfileSuccess` 100% / `renderProfileResult` 100% / `buildIndexView` 93.8% / `handleProfileDelete` 88.2% / `handleProfileCreate` 85.7% (잔여 gap = ParseForm 오류 방어 분기, 저가치). 패키지 커버리지: internal/web 71.1%(M3 base ~70.6% → 신규 코드 잘 커버, 미세 상승), internal/profile 80.2%. 패키지 총계 <85% 는 대형 UI 패키지의 선재 baseline(M4 도입 아님) — 신규 코드는 전부 85%+.

**PRESERVE 검증**: `git diff --name-only 4f5b3fbbb..HEAD` = internal/web/{app.go, handlers.go, root.templ, root_templ.go, profile_crud.go, profile_crud_test.go, profile_traversal_test.go, assets/{i18n.js, console.css}} + internal/profile/{profile.go, preferences.go, traversal_test.go} + progress.md. renderer.go/cache_hit_test.go/app.go:90-92 CSRF/agent 파일/병렬 세션 산출물 전부 무접촉.

**선재 실패 (무접촉 유지)**: `go test ./...` = 91 pkg ok, 유일 FAIL = `internal/cli TestRunHookEvent_ReadInputError` (base 선재, 병렬 세션 SPEC-CLI-SUBPKG-SPLIT-001 도메인 — finding #2 동일). internal/cli 무변경(`git status --porcelain internal/cli/` empty).

### M5 — SPEC READ-ONLY Board (커밋 <M5-spec> + <M5-web>)

> **범위**: 본 후속 위임은 **M5만** 커버한다. base = 3e6471f25 (M4 tip). cycle_type=tdd. L1 runtime worktree (`worktree-agent-acd2613c5f8ce51b9` 브랜치). M6은 후속 위임 소관.

- **spec.ListDocs export (REQ-WC11-041)**: `internal/spec/listdocs.go` — `DocRecord{Path, Frontmatter, ParseError}` + `ListDocs(baseDir)` exported wrapper (unexported discoverSPECs/parseSPECDoc 경유). baseDir=프로젝트 루트(Audit 규약과 동일하게 `.moai/specs` 내부 append), 결측 dir→빈 slice 무오류, per-record ParseError 표면(malformed 1건이 전체 스캔 중단 방지), path 정렬 결정성. **TDD RED**(stub 부재 → `undefined: ListDocs` 컴파일 실패) → GREEN. 커버리지 93.3%.
- **Tier frontmatter field (REQ-WC11-042)**: `SPECFrontmatter`에 `Tier string yaml:"tier,omitempty"` 추가 — optional(12 필수 무영향, TestFrontmatterSchemaRule GREEN 유지, Era/HarnessLevel 선례와 동일 패턴). 부재→빈 문자열(오류 아님).
- **보드 핸들러 (REQ-WC11-040)**: `internal/web/board.go` — GET `/specs` `handleBoard`(app.go routes()에 라우트 추가, GET-only 405 gate). `buildBoardView` = `spec.ListDocs(ProjectRoot)` + `spec.Audit(BaseDir=ProjectRoot)` **순수 FS 스캔만**. status 분포(canonical 8-value enum 순서 + out-of-enum/`(unknown)` 버킷) + close-debt(status==implemented) + MUST-FIX(Severity==MUST-FIX) 뷰모델 조립. boardSpecID는 frontmatter id 부재 시 디렉터리명 폴백.
- **Templ (`board.templ` → `board_templ.go`)**: standalone `boardPage` — 자체 `<html>` shell, 패키지 const `brandBadgeSVG`/`foucHeadScript` + `@icon` 재사용, `<form>`/write 컨트롤 0. status badge + close-debt 열(tier optional badge — 있으면 `board-tier`, 없으면 생략) + MUST-FIX badge + copyable remediation(`<code class="board-remedy">` + `data-copy` 버튼). `templ generate -path ./internal/web`(Makefile/CI 정본) → **기존 _templ.go 무변경(드리프트 0)**, board_templ.go만 신규(bare `board.templ` FileName 규약 일치).
- **copyable remediation (REQ-WC11-043/044)**: `assets/app.js` document-레벨 위임 click 리스너를 IIFE 최상위에서 **1회만** 등록(initConsole 내부 아님 — htmx afterSettle 중복등록 회피) — `data-copy` 값 클립보드 복사(navigator.clipboard + textarea/execCommand 폴백, 복사 ✓ 플래시). **서버측 명령 실행 0**.
- **제외 (REQ-WC11-045/046)**: `DetectDrift` 동기 호출 0(`grep -rn DetectDrift internal/web/` = 0 실측), status 쓰기 경로 0, 명령 실행 0. 가드: `TestBoard_NoWritePathSourceScan`(board.go 소스 스캔 — exec.Command/os-exec/UpdateStatus/WriteFile/PatchFile/git-drift 토큰 0) + `TestBoard_GETOnly`(POST/PUT/DELETE/PATCH → 405, loopback host로 hostCheck 통과 후 handleBoard 자체 405 격리).
- **i18n ×4 (REQ-WC11-015/061)**: 신규 `board.*` 키 **13종**을 en/ko/ja/zh 4-locale 전부에 추가(수작업 번역, native UTF-8, blind sed 0). status/SPEC-ID/tier/finding-type/remediation 명령은 code-chip(비번역 계약 — i18n.js 헤더 정합) 유지. `node --check` + 4-locale parity 스크립트 PASS.

**M5 AC 매트릭스 (전 [B] PASS, [N] 포함)**:

| AC | Sev | 검증 | 결과 |
|----|-----|------|------|
| AC-WC11-040 | [B] | `go test -run TestBoard_Render ./internal/web/` (fixture 4-SPEC) | PASS — status 분포(implemented/completed) + close-debt(DEBT-001/DRIFT-002, DONE-003 제외) + MUST-FIX badge 렌더 |
| AC-WC11-041 | [B] | `go doc ./internal/spec ListDocs` + `go test -run TestListDocs ./internal/spec/` | PASS — exported `func ListDocs(baseDir string) ([]DocRecord, error)` 문서화 + 4 unit test ok |
| AC-WC11-042 | [N] | `go test -run 'TestBoard_TierBadge\|TestListDocs' ./...` | PASS — tier: L → `board-tier` badge 렌더; tier 부재 → badge 생략 + 200(오류 0) |
| AC-WC11-043 | [B] | `go test -run TestBoard_RemediationCopyable ./internal/web/` | PASS — `moai spec close SPEC-BOARD-DRIFT-002 --backfill-only` 이 `<code class="board-remedy">` + `data-copy="..."` 버튼으로 렌더 |
| AC-WC11-044 | [B] | `go test -run TestBoard_NoWritePathSourceScan ./internal/web/` (board.go 소스 스캔) | PASS — exec/os-exec/status-transition/write/patch 토큰 0 (board 핸들러 신규 exec 0) |
| AC-WC11-045 | [B] | `grep -rn "DetectDrift" internal/web/` + `go test -run TestBoard_NoGitDriftPathInWebPackage` | PASS — 비-테스트 web 소스 0 매치 (동기 렌더는 pure-FS Audit만) |
| AC-WC11-046 | [B] | `go test -run TestBoard_GETOnly ./internal/web/` | PASS — GET 200, POST/PUT/DELETE/PATCH 405 (쓰기 경로 부재) |
| AC-WC11-060 | [B] | `grep -rn "csrf\|CSRF\|xsrf" internal/web/ \| grep -v _test.go \| grep -v REQ-WC-009 \| grep -vE ':[0-9]+:[[:space:]]*//'` | PASS — 비-주석 CSRF 매치 0 (app.go @MX:NOTE 보존, routes()에 라우트만 추가) |
| AC-WC11-061 | [B] | `go test -run TestBoard_I18nParity ./internal/web/` | PASS — 렌더된 board.* 키 각각 i18n.js 4회(en/ko/ja/zh) 출현 |
| AC-WC11-062 | [B] | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/web/board.go internal/web/board.templ internal/spec/listdocs.go \| grep -v // ` | PASS — 0 매치 (subagent boundary) |

**라이브 리포 성능 재확인(GWT-3 전제, spec.Audit 순수 FS)**: `Audit(BaseDir=repo root)` = 414 SPECs / 330 drift findings(0 MUST-FIX) / **112ms**; `ListDocs(repo root)` = 414 records / **37ms**. 합계 ~150ms — 동기 렌더 허용(git 호출 0, DetectDrift 7.9s 경로 미접촉). 라이브 catalog MUST-FIX 0.

**PRESERVE 검증**: `git status --porcelain` = internal/spec/{lint.go, listdocs.go, listdocs_test.go} + internal/web/{app.go, board.go, board.templ, board_templ.go, board_test.go, assets/{i18n.js, app.js}} + progress.md 만. renderer.go/cache_hit_test.go(병렬 세션)/app.go:90-92 CSRF @MX:NOTE/agent 파일/병렬 세션 산출물 전부 무접촉. internal/cli 무변경.

**선재 실패 (무접촉 유지)**: `go test ./...` = 유일 FAIL `internal/cli TestRunHookEvent_ReadInputError`(nil pointer panic, base 선재, 병렬 세션 SPEC-CLI-SUBPKG-SPLIT-001 도메인 — finding #2/M4 동일). `git status --porcelain internal/cli/` empty. 선재 gofmt 드리프트(lint.go EOF 후행 빈 줄 등)는 무접촉(내 Tier 편집 구간 gofmt clean).

### M6 — Statusline cache_hit delta + segment-list SSOT (커밋 <M6>)

> **범위**: 본 후속 위임은 **M6만** 커버한다 (마지막 마일스톤 — run-phase 완료). base = 62e3077c9 (M5 tip). cycle_type=tdd. L1 runtime worktree (`worktree-agent-a12d03f291ea54d9e` 브랜치). renderer는 committed 3e30fef48(SPEC-TOKEN-EFFICIENCY-001)에 이미 존재 — M6은 렌더링 미구현, **노출 fan-out만**.

- **cache_hit 노출 fan-out (REQ-WC11-050)**: renderer(무접촉)에만 존재하던 `SegmentCacheHit`("cache_hit")를 6개 노출 표면 전부에 추가 — SSOT `statusline.CanonicalSegments`(preset.go) + `settings.statuslineSegmentKeys`(schema.go) + profile `defaultStatuslineSegments`(sync.go) + TUI `statuslineAllSegments`(profile_setup.go) + live/template statusline.yaml + i18n.js. Go 표면은 magic string 대신 SSOT 상수 `statusline.SegmentCacheHit` 참조(§14 하드코딩 회피). **segment 카운트 15→16, section 필드 카운트(theme+segments) 16→17.**
- **TUI 완전 배선**: statuslineAllSegments만 추가하면 MultiSelect Options(15)와 불일치하는 half-done TUI가 되므로 — `profileSetupText.SegmentCacheHit` struct 필드 + 4-locale 번역(en "Cache hit ratio" / ko "캐시 적중률" / ja "キャッシュヒット率" / zh "缓存命中率") + MultiSelect `huh.NewOption` + `schemaSegmentBridge` `seg.cache_hit` 전부 추가.
- **segment-list SSOT set-equality (REQ-WC11-051, design.md §F)**: `TestSegmentListSSOT` 3-패키지 white-box(settings/profile/cli) — 각 로컬 목록을 SSOT `statusline.CanonicalSegments`와 집합 비교(순서 무시, 중복/누락/초과 실패). cli→settings import 사이클 때문에 단일 중앙 테스트 불가 → 패키지별 분산(동일 SSOT 앵커로 이행적 4-목록 집합-동일; 신규 public API 0). `go test -run TestSegmentListSSOT ./...` = 3 PASS. drift 시 컴파일·CI 차단(3-way orphan 재발 방지).
- **파생 카운트 (REQ-WC11-053)**: fieldsets.templ 카운트 라벨은 base(committed)에서 이미 `@sectionCount(len(settings.SectionFields(SectionStatusline)))` 파생 — cache_hit 추가로 17 자동 반영. `grep -nE '"[0-9]+ fields"'`=0. stale "11-segment" 주석(profile_setup.go) 정정.
- **Template-First (REQ-WC11-052)**: live + template statusline.yaml 양쪽 `cache_hit: true` + `make build` 재임베드(exit 0; catalog.yaml 무변 — config yaml은 skill/agent 해시셋 밖). §25 neutrality: 추가 라인 SPEC/REQ/date 토큰 0 (line 28 `<SPEC-ID>`는 base 선재 플레이스홀더, 무접촉).

**M6 AC 매트릭스 (전 [B] PASS, [N] 포함)**:

| AC | Sev | 검증 | 결과 |
|----|-----|------|------|
| AC-WC11-050 | [B] | 6표면 노출: profile_setup.go=2, i18n.js=4, live/template yaml=1/1, Go 표면 SSOT 상수 `SegmentCacheHit`=schema.go/preset.go/sync.go 각 1 + `TestSegmentListSSOT` 집합 증명 | PASS — 전 표면 노출, 집합-동일 |
| AC-WC11-051 | [B] | `go test -run TestSegmentListSSOT ./...` (settings/profile/cli 3-pkg) | PASS — 3/3 ok; drift-on-add 실패 보장 |
| AC-WC11-052 | [B] | live vs template `grep cache_hit` 양쪽 1; template §25 (추가 라인 SPEC/REQ 토큰 0) | PASS — 미러 동일 키 + neutrality clean |
| AC-WC11-053 | [B] | `grep -nE '"[0-9]+ fields"' internal/web/fieldsets.templ`=0 + @sectionCount 파생 + `go test -run TestStatusline ./internal/web/` | PASS — 파생 라벨 + 하드코딩 총계 0 + segment 16 |
| AC-WC11-054 | [N] | `grep -n "11-segment" internal/cli/profile_setup.go` | PASS — 0 매치 (stale 주석 정정) |
| AC-WC11-060 | [B] | (재확인) CSRF 비-주석 grep | PASS — 0 (M6 web 편집=i18n.js/handlers.go 주석/statusline_test/schema_render_test) |
| AC-WC11-061 | [B] | seg.cache_hit 4-locale + TUI SegmentCacheHit 4-locale(`TestProfileSetupTranslations_PresetSegments`) | PASS — parity 4/4 |
| AC-WC11-062 | [B] | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/web/ internal/spec/ \| grep -v _test.go \| grep -v //` | PASS — 0 매치 |

**PRESERVE 검증**: `git status --porcelain` = statusline/{preset.go,types.go} + settings/{schema.go,accessors.go,accessors_test.go,schema_test.go,+segment_ssot_test.go} + profile/{sync.go,statusline_segments_test.go,+segment_ssot_test.go} + cli/{profile_setup.go,profile_setup_translations.go,profile_setup_translations_test.go,profile_setup_test.go,schema_bridge.go,+segment_ssot_test.go} + web/{assets/i18n.js,handlers.go,statusline_test.go,schema_render_test.go} + live/template statusline.yaml + progress.md. **renderer.go/cache_hit_test.go/stdinfields_test.go(병렬 세션) 무접촉**(git diff --name-only 부재), app.go:90-92 CSRF @MX:NOTE/agent 파일/병렬 세션 산출물 무접촉. internal/spec 무변경.

**선재 실패 (무접촉 유지)**: `go test ./...` = 유일 FAIL 패키지 `internal/cli`. git stash로 clean HEAD 대비 검증한 결과 (a) `coverage_test.go` `TestRunHookEvent_ReadInputError`/`TestRunAgentHook_ReadInputError` nil-pointer panic (병렬 세션 SPEC-CLI-SUBPKG-SPLIT-001), (b) `schema_bridge_test.go` `TestBridgeFieldDefResolver`/`TestI18nKeySetParity`/`TestTUIRendersSchemaFieldSet` — M3 `workflow.workflow_agents.*` TUI-bridge 미배선 — **전부 base 선재**(clean HEAD에서 동일 FAIL, 해당 파일 무접촉). 내 M6 편집분은 신규 FAIL 0 (수정 대상 테스트 TestSectionFieldsOrder/TestStatuslineAllSegments_CardinalityAndOrder/TestSchemaSixSections 등 count assertion 전부 갱신 후 PASS).

### M2b-regression 후속 정정 — TUI 브리지 테스트 범위 조정 (REQ-WC11-015/061 대칭성 완결)

> **범위**: base = 22220186c (M6 tip). cycle_type=tdd. 위 M6 "선재 실패 (b)"로 기록된 3개 `internal/cli/schema_bridge_test.go` 테스트를 정식 해소한다. M2b(002446611)가 WEB측 i18n parity 테스트(`internal/web/schema_label_test.go` `TestI18nKeySetParity`)에는 PersistSeam/PersistTypedSection 제외 스코프를 적용했으나, 대칭인 TUI측 3개 브리지 테스트에는 누락하여 회귀가 남아 있었다 — 본 정정으로 교차표면 대칭성을 완결한다.

- **근본 원인 검증(verify, don't assume)**: TUI 위저드(`profile_setup.go`)가 `settings.AllFields()`를 제네릭 순회하지 않고 수제 위젯 셋(기존 34필드 + statusline 세그먼트)만 바인딩함을 실측 확인 — `PersistSeam`/`PersistTypedSection` 섹션명 참조 0건, 인라인 `huh.New*` 92건. 따라서 M2b/M3 확장 필드(웹 콘솔 전용 key-chip, TUI 위저드 부재)는 TUI 브리지 항목이 설계상 부재이며, 3개 테스트의 `settings.AllFields()` 전수 순회가 이들을 잘못 요구하고 있었다.
- **정정(테스트 범위 조정, 동작 변경 아님)**: `schema_bridge_test.go`에 `isWebOnlyKeyChipField(f) = f.Persist.Kind==PersistSeam || PersistTypedSection` 술어를 추가하고 3개 테스트 루프(`TestI18nKeySetParity`/`TestBridgeFieldDefResolver`/`TestTUIRendersSchemaFieldSet`)에서 skip — WEB측(`schema_label_test.go`) 스코핑의 대칭 미러. persist-kind 기준 제외(blanket skip 아님)라 TUI-렌더 셋 신규 추가(profile-store/project-config, M6 cache_hit 세그먼트 포함)는 여전히 검증됨(가드 주석 명시).
- **RED→GREEN 증거**: 정정 前 3개 전부 FAIL(누락 키 verbatim: `f.quality.coverage_threshold` 등 M2b/M3 seam/typed 필드) → 정정 後 3개 전부 PASS. 파티션 실측(ephemeral test): `AllFields()=240` = asserted(TUI-렌더) 35(profile-store 26 + project-config 9, `seg.cache_hit` 포함) + excluded(web-only key-chip) 205, 교차오염 0.
- **AC 매트릭스 (전 PASS)**:

| AC | 검증 | 결과 |
|----|------|------|
| REQ-WC11-015/061 대칭 | `go test -run 'TestI18nKeySetParity\|TestBridgeFieldDefResolver\|TestTUIRendersSchemaFieldSet' ./internal/cli/` | PASS — before=3 FAIL / after=3 PASS (verbatim) |
| 34+세그먼트 셋 보존 | ephemeral 파티션 test — asserted=35(profile-store 26 + project-config 9) + `seg.cache_hit` asserted=true | PASS — 제외군은 Seam/Typed만(205), asserted에 Seam/Typed 0 |
| 잔여 FAIL 격리 | `go test ./internal/cli/` = coverage_test.go의 2개 nil-pointer panic만(병렬 SPEC-CLI-SUBPKG-SPLIT-001) | PASS — skip 3개 검증 시 패키지 `ok` |

- **PRESERVE 검증**: `git diff --name-only` = `internal/cli/schema_bridge_test.go` + `progress.md` 2파일만. renderer.go/cache_hit_test.go(병렬 세션)·coverage_test.go(무관 nil-pointer)·spec/plan/acceptance body·프로덕션 코드 무접촉. 빌드 darwin/windows/linux 전부 exit 0, `go vet` OK, `golangci-lint run ./internal/cli/` = 0 issues.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-05
run_commit_sha: "af4bdf245 (M1) + 5d2d18f3b (M2a) + 002446611 (M2b) + ab555742f/4f5b3fbbb (M3) + 133a3c9d5/b66580ffc/d436aa80c (M4) + <M5-spec 커밋>/<M5-web 커밋> + <M6 커밋 — 커밋 후 backfill>"
run_status: in-progress — M1 + M2a + M2b + M3 + M4 + M5 + M6 완료; run-phase 완료 (sync 대기)
ac_pass_count: 45  # M1/M2a/M2b 15 + M3 13 + M4 6 + M5 6 + M6: 050, 051, 052, 053 (4 [B]) + 054 ([N]) = 5
ac_fail_count: 0
ac_partial_notes: "M5: AC-WC11-042는 [N] tier optional badge PASS(present/absent 양 케이스). AC-WC11-060/061/062는 M5 보드 스코프에서 재확인 PASS(CSRF 비-주석 0, board.* 키 4-locale parity, subagent boundary 0). 이전: AC-WC11-011 파티션 테스트(TestQualityKeyPartition) 충족; AC-WC11-005 allowlist 불변; AC-WC11-015 key-chip 비번역 계약."
preserve_list_post_run_count: 0   # PRESERVE 위반 0 — statusline/app.go@MX:NOTE/병렬 세션 산출물/agent 파일/internal/cli 전부 무접촉
l44_pre_commit_fetch: "n/a — L1 runtime worktree 격리 실행 (전용 브랜치, push 금지 지시); landing은 orchestrator gate 검증 후 수행"
l44_post_push_fetch: "n/a — push 미수행 (지시 사항)"
new_warnings_or_lints_introduced: 0   # M6: golangci-lint run ./internal/statusline/... ./internal/settings/... ./internal/profile/... ./internal/cli/... ./internal/web/... → "0 issues."
cross_platform_build:
  darwin_arm64: PASS (go build ./... → BUILD_OK)
  windows_amd64: PASS (GOOS=windows GOARCH=amd64 go build ./... → WIN_OK)
  go_vet: PASS (go vet ./internal/web/ ./internal/spec/ → VET_OK)
coverage:
  internal_spec: "87.8% (M5 후 — listdocs.go ListDocs 93.3%)"
  internal_settings: "90.9%"
  internal_settings_yamlpatch: "86.3%"
  internal_web: "70.4% (M5 후 go test -cover ./internal/web/); 신규 board.go 88.9%(48/54 stmts — boardSpecID/orderedStatusCounts/boardCount 100%, buildBoardView 88.9%, handleBoard/renderBoard 잔여 gap=방어적 err 분기). 패키지 총계 <85%는 대형 UI 패키지 선재 baseline — M5 도입 아님, 신규 코드는 85%+"
  m6_touched_packages: "statusline 84.7% / settings 91.1% / web 70.6% / profile 80.2% (go test -cover, M6 후). 패키지 총계 <85%(web/profile/statusline)는 선재 baseline — M6은 리스트 상수/맵 항목 추가 + 주석 위주 additive 변경, 수정 라인은 TestSegmentListSSOT/PresetSegments/CardinalityAndOrder로 커버. 패키지-전체 커버리지 상향은 M6 스코프 아님."
total_run_phase_files: 22   # M6: 신규 3(settings/profile/cli segment_ssot_test.go) + 수정 18(statusline/{preset,types}.go, settings/{schema,accessors,accessors_test,schema_test}.go, profile/{sync,statusline_segments_test}.go, cli/{profile_setup,profile_setup_translations,profile_setup_translations_test,profile_setup_test,schema_bridge}.go, web/{assets/i18n.js,handlers,statusline_test,schema_render_test}.go, live+template statusline.yaml) + progress.md
m1_to_mN_commit_strategy: "마일스톤별 path-limited 커밋; M5는 spec-export 커밋 → web-board+progress 커밋 2분할(데이터 소스 먼저), 전용 worktree 브랜치, push/landing은 orchestrator 소관"
known_preexisting_failures: "internal/cli TestRunHookEvent_ReadInputError (base 선재, 병렬 세션 SPEC-CLI-SUBPKG-SPLIT-001 도메인, 무접촉); gofmt 드리프트(lint.go EOF 후행 빈 줄 등, base 선재, 무접촉 — 내 Tier 편집 구간은 gofmt clean)"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
