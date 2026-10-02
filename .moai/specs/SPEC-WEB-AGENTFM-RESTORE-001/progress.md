# progress.md — SPEC-WEB-AGENTFM-RESTORE-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_complete_at: 2026-10-02 (plan-audit PASS-delta — iter3 0.90 + 델타 확인, run 진입 GO)
- plan_status: audit-ready (감사 계보: iter1 FAIL 0.69 D1-D13 → 수리 → iter2 FAIL 0.875 D14-D17 → 수리 → iter3 FAIL 0.90 D18 잔여 → 리드 승인 ② 1절 수리+델타 확인 → **PASS-delta GO**. 판정 파일: plan-audit.md·plan-audit-iter2.md·plan-audit-iter3.md·plan-audit-iter3-delta.md — 전부 .moai/reports/t1411/. 최종 아티팩트 기준 HEAD ccac1f555)
- pre-flight baseline: 미측정 (§C Pre-flight 체크리스트 — run-phase 착수 시 최우선 기록)

## §E.2 Run-phase Evidence

### Pre-flight (2026-10-02, HEAD a48216da1)

- branch/HEAD: `WT-web-subagent-config` / `a48216da1` — 기대값과 일치.
- baseline build: `go build ./...` exit 0 · `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- baseline lint: `golangci-lint run --timeout=2m` → `0 issues.` (exit 0) — /tmp/t1411-lint-baseline.txt.
- baseline scope tests: `go test ./internal/web/ ./internal/settings/... ./internal/template/ ./internal/config/ ./internal/cli/ -run 'AgentSettings|RetiredModelKey|UpdateLLMYAML|ConsoleTabs|PrimarySurface|I18n|ShippedConfigKeys'` — GREEN (M1 커밋 전 재측정 기록은 아래 각 마일스톤 행).
- **원자성 실패-주입 프로브 (plan §C / plan-audit D9)**: 관측 결과 **원자(atomic)** — 실패 주입점 3곳(stepApplySchema·stepGlmcredSave·stepJevcredSave)에서 실제 섹션 쓰기 seam(writeProjectConfig·writeProjectNestedConfig·applySchemaEdits)을 실구동한 뒤 관측: 무관 섹션 파일(quality.yaml, no-op 제출)은 3케이스 전부 byte-identical, 대상 파일(feedback.yaml)은 제어군 대로 실패 이전 단계에서만 기록. codex 1차 관측(quality.yaml 비원자)은 본 트리에서 미재현 — SPEC-WEB-SAVE-LOSSLESS-001 라인-스플라이스 seam(no-op 게이트 포함)이 그 결함 기제를 제거한 뒤다. **M3 조건부 staging/rollback 설계 단계 불요 — REQ-AFR-007은 기존 흐름이 실증 보호.** 주입 차량: recordingSeams 확장 임시 프루브(커밋 대상 아님 — 관측 후 삭제, §D.5의 partial_apply 확장은 M3가 정식으로 수행).

### M1 (2026-10-02) — 데이터 모델·해상 기계 복원

- RED (구현 전 관측, HEAD a48216da1 트리): 신규 테스트 4파일 컴파일 실패 — `internal/config` (`undefined: DefaultProfile / unknown field Profile / EffectiveProfile undefined / IsValidProfile...`), `internal/template` (`undefined: ProfileHigh / ProfileMatrixAgents / DefaultProfileMatrix / IsGLMCodingMaxOverrideAgent / GLMCodingMaxOverrideAgents / ResolveGLMReasoning / ResolveGLMReasoningForModel`).
- GREEN: config 신규 테스트 5종 + template 신규 테스트 10종 전부 PASS (TestEffectiveProfile·TestProfileClosedSet·TestValidateProfileRule·TestValidateAgentOverridesRule·TestValidateWiresProfileRules / TestDefaultProfileMatrix_CellsAreCurrentConfigDefaults·TestDefaultProfileMatrix_RowMonotonicityAndNoSentinels·TestProfileMatrixAgents_CurrentRoster·TestResolveAgentModelEffort_Precedence·TestValidPerformanceTiers_SelectorVocabulary·TestAgentGroup_CurrentMembership·TestIsGLMCodingMaxOverrideAgent·TestResolveGLMReasoning·TestResolveGLMReasoningForModel·TestShippedRetiredModelKeys_IncludesReshippedConsoleKeys).
- **셀 재유도 설계 기록 (plan §B-1(b)/§F M1)**: 1차 시도(config.DefaultClaudeTier* 페어를 셀로)는 REQ-AFR-004 계약과 충돌을 확인하고 기각 — 티어 페어의 model 값은 전체 세대 id("sonnet-5-5")라 override 폐쇄집합 {inherit, haiku, sonnet, opus, fable} 밖이므로, "제출=기본값 → clear" 비교가 UI 경유로는 성립 불가(행 select도 일치 옵션 없음). 최종 채택: model 축=config claude_models 기본값({high: opus, medium: sonnet, low: haiku})의 High/Medium 열(No-Haiku 정책 승계 — Low는 셀에 진입 안 함), effort 축=구 매트릭스의 judgment-weighted 정책 레벨(EffortLevel* 상수). 셀 테스트가 이 단일 원천을 단언(TestDefaultProfileMatrix_CellsAreCurrentConfigDefaults).
- 수재 키 seam 실측: 템플릿 llm.yaml에 profile: ""·agent_overrides: {} 재수재 → `ShippedRetiredModelKeys()`가 두 키를 집합에 포함(멤버십=strip 제외) 실측 PASS · performance_tier/profiles/harness_agents는 계속 strip 대상 (TestStripRetiredModelConfig_ReshippedConsoleKeysSurvive — 신설, 실제 임베디드 템플릿 대상 실측).
- 인벤토리 등록: llm.agent_overrides·llm.profile → internal/config/testdata/shipped_key_inventory.yaml (R) — TestShippedConfigKeysHaveReaders GREEN.
- 유출 가드 수정: 템플릿 llm.yaml 주석에서 내부 SPEC ID 제거 (TestTemplateNoInternalContentLeak C1-spec-id-prefix 적중 → 재기술 후 GREEN).

### M2 — RED 기선: 락 반전 (M1 커밋 5fbb0ab58 기점 관측)

- **관측-RED 1차 (락 반전, /tmp/t1411-m2-red.txt verbatim)**: `internal/web/agent_overrides_test.go` 신설(구 agent_settings_removed_test.go 반전분) — 첫 실행 exit 1:
  - `--- FAIL: TestAgentOverridesSubsection` — `GET /settings lacks "data-section=\"agent-overrides\""`, `lacks "sec.agentfm.title"`, `lacks "data-agent-row=\"manager-develop\""`, `lacks "name=\"agentfm.manager-develop.model\""`, (manager-todo 동일 3종), `the parse-failed agent row must render downgraded as agentfm.unavailable` — 렌더가 아직 없음(M4 전).
  - `--- FAIL: TestPerfTierSave` — `llm.profile = <nil>, want "max" (the selector wire value persists verbatim)` — 저장 경로가 아직 없음(M3 전). custom/empty 보존 서브테스트는 통과(작성자가 없어 자명 — GREEN 시 유의 어설션).
  - `--- FAIL: TestAgentOverridesSave` — `override equal to the profile default must be cleared: map[manager-develop:...]`, `backfilled override = map[], want {model: sonnet, effort: medium}` — 핀/clear/백필 경로 부재. 원자 거절 서브테스트는 통과(현재 무시되는 제출이 우연히 조건 충족 — M3 GREEN 시 필드 오류 어설션이 유의성을 전환).
  - `--- PASS: TestAgentFrontmatterUntouched` — §D.5 1행 설계대로 frontmatter 무접촉 절반만 유지(이미 참). 효과 절반은 TestPerfTierSave/TestAgentOverridesSave가 보유.
- **관측-RED 2차 (mcp_audit_surface 가드, plan §F M2)**: 센티널 운반 프루브 파일(zz_sentinel_probe.go — template.ResolveAgentModelEffort·ProfileMatrixAgents·config LLMConfig.EffectiveProfile 참조)을 internal/web에 두자 기존 `TestWebConsole_NoPerAgentModelResolver`가 RED: `internal/web/zz_sentinel_probe.go references ResolveAgentModelEffort/ProfileMatrixAgents/EffectiveProfile — the web console assigns no per-agent model or effort` — 프루브 삭제 후 축소 계약으로 개정: (1) 정의 금지(`func ResolveAgentModelEffort` 전 파일 금지 — 단일 유도), (2) 4 센티널 참조는 표면 파일(agentfm.go·app.go·handlers.go·schemaform.go)에서만 허용 — M2 시점 실존 3파일만 등록, agentfm.go는 M3 착지 시 등록 확장, (3) 제외 집합 비-공 가드 + 파일 존재 검사. 개정 후 GREEN 실측.
- 구 absent 테스트 삭제: agent_settings_removed_test.go 파일째 삭제 (TestAgentSettingsTab_IsNotRendered·TestAgentSettingsFields_ArePostedWithoutEffect 소멸 — AC-AFR-012 EV-AFR-012 전제).
- agent_settings_test.go :51 고아 주석 제거 (§D.5 2행).
- 코드 주석 정합 (§D.1 cli/template 행): update_model_key_strip.go 사유 문자열("no reader"→"seam 자동 제외" 서술), retired_model_keys.go 헤더, profile_setup_schema_options_test.go :119 고아 주신 갱신. settings/schema.go:80·schema_sections.go:568의 agentfm 서술 주석은 M3(표면 착지 시점)에 갱신 — 복원 전 갱신은 허위 서술이 되므로 순서 유지.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

- 입력: tier=M · scope=약 15-20파일(internal/template·config·cli·web + assets) · 도메인=Go/templ/JS/i18n 4종 · 언어 혼합=코드+생성물+마크다운 · 동시성 이점=LOW(coding-heavy) · agent-team 전제=미충족(명시 요청 없음)
- 모드 평가: direct=미선정(단순 수정 아님) · fanout=미선정(coding-heavy — Anthropic 병렬화 주의) · sweep=미선정(기계적 균일 변환 아님) · agent-team=미선정(명시 요청 없음)
- Decision: serial
- 근거: 마일스톤 M1→M2→M3→M4→M5→M5b→M6가 순차 의존(데이터 모델→락 반전→저장→UI→i18n→개정→검증)하는 coding-heavy 재포트 — 단일 구현 에이전트 순차 위임이 기본 선택(Anthropic coding-task 병렬화 경고). M5b는 소유권 매트릭스상 별도 manager-spec 재위임으로 직렬 삽입.
- Kickoff decision record (2026-10-02, 자율형 — auto-semantics §9.1): plan→run 진입 승인. 증거: 감사 교차(상기 판정 파일 4종, 최종 PASS-delta) + 증거 기준(점수 0.90 ≥ Tier M 0.80·아티팩트 해시 ccac1f555 기준 불변 — iter3-delta 직접 기록) + 운영자 판정 3건 기록(decision-index Q1·Q2·Q3) + 차단 결함 0건. keep-set 해당 없음(환경 불가·운영자 보유·외부 공유 시스템 조작 없음 — 후속 카드 t1421은 리드 큐 발행 완료). 구현 배차는 일반 타입 에이전트(manager-develop 타입 스폰 자체 L1 격리 회피 — t1318 교훈).
- 역방향 완결: agent_settings_removed_test.go 삭제(TestAgentSettingsTab_IsNotRendered·TestAgentSettingsFields_ArePostedWithoutEffect 소멸), agent_settings_test.go :51 고아 주석 제거, crosssession_test.go 노이즈 키 예시를 현 로스터 이름으로 갱신(어설션 불변 — agentfm.*은 crosssession 키가 아님).

### M3 — 저장 경로 (M1 커밋 5fbb0ab58 · M2 커밋 315958ff7 이후)

- 구현: internal/settings/agentfm 패키지 복원(List·splitFrontmatter·AgentInfo 전용 — Patch 쓰기 레이어는 REQ-AFR-005대로 미부활, 패키지 헤더에 계약 명기) + 단위 테스트 4종(파싱 실패 행 강등·빈/부재 디렉터 저하·비-md 제외). internal/settings/llmoverrides.go 신설 — WriteLLMProfile(seam 스플라이스 + 동치 무기록 게이트)·WriteLLMAgentOverrides(블록 라인-스플라이스 — yamlpatch는 스칼라 전용·삭제 미지원이므로 agent_overrides 맵의 upsert/clear를 블록 단위 재작성으로 해결, 블록 밖 바이트 전부 보존, 결과 byte-identical 시 무기록). 구 SetSection("llm")→Save 전체 재마샬은 부활하지 않음(GitHub issue #1731 결함 기제).
- web: agentfm.go 재포트(M3 저장 경로 절반 — parsePerfTierForm·parseAgentFMForm·applyPerfTierEdits·applyAgentOverrides·listAllAgentFMs·해상 접근자. 렌더 절반은 M4에 동일 파일에 추가), app.go seam 3종(listAgentFMs·patchAgentFM·applyPerfTierEdits) 재도입, handlers.go 저장 배선(perfTier→agentfm 파싱을 validator 병합 전에, applyPerfTierEdits(7)→patchAgentFM(8)를 applySchemaEdits(6) 뒤에 — 동일 요청의 override가 새 티어 기준으로 비교되도록 순서 유지), schemaform.go applySchemaCurrent 시딩 복원(PerfTier/PerfTierIsEmpty/AgentFMs/LLM/PerfTierCustom — performance_tier 별칭 열 제거 적응), pageView 필드 재도입.
- 하니스 확장(§D.5 amend/extend 2행): partial_apply_repro_test.go — stepApplyPerfTier(7)·stepPatchAgentFM(8) 재등록(상수·injectableSteps·recordingSeams·양성 대조 want — 레저 diff :6807-6876 제거분의 역방향), save_observability_test.go — seamTable 2행 동행. TestPartialApplyOrderPositiveControl 9스텝 GREEN 실측.
- mcp_audit_surface_test.go: 허용 집합에 agentfm.go 추가(M2 등록 확장) — 가드 GREEN.
- settings schema.go/schema_sections.go의 agentfm 서술 주석 복원 서술로 갱신(§D.1 — M3 착지 시점에 맞춰 갱신).
- 신규 단위 테스트: TestWriteLLMProfileSplicesAndGates·TestWriteLLMAgentOverridesBlockSplice·TestWriteLLMAgentOverridesAbsentKeyAndFile(주석·미모델링 키 보존, clear·전체 clear→{}·무기록 게이트 실측) + agentfm 4종.
- 관측: TestPerfTierSave GREEN 전환(M2 RED→GREEN — llm.profile: "max" 기록·performance_tier 미생성·custom/empty byte-identical 실측). TestAgentOverridesSave의 pin/clear/no-op/non-matrix/backfill 서브테스트 GREEN — 원자 거절 2서브테스트는 렌더 의존(M4 전환 대상으로 RED 유지). TestAgentOverridesSubsection RED 유지(렌더 — M4). web 패키지 나머지 전량 GREEN(신규 회귀 0).

### M4 — UI 노출 (M3 커밋 514a48f34 이후)

- fieldsets.templ: fieldsetAgentFM(llm 패널 서브섹션 — data-section="agent-overrides" 마커·sec.agentfm.title·성능 티어 라디오[키 칩은 llm.profile — 은퇴 alias 아님]·moai-profile-matrix JSON island·agentFMGridRows 그룹 헤더 그리드) + agentFMRow(배지·설명 data-i18n-baseline·model/effort select·haiku disabled+data-haiku-hint·data-glm-reasoning·행별 fieldErr) 재포트. 14탭 아님 — `meta.PanelID == "llm"` 앵커(§D.3). 구 warn 배너("update가 덮어쓴다")는 저장 대상이 llm.yaml이라 성립하지 않아 미재포트. fieldsets_templ.go 재생성(templ v0.3.1020 — make templ-generate).
- 컴포넌트명 설계 기록: 서브섹션 컴포넌트를 fieldsetAgentFM으로 명명 — fieldsetAgentOverrides라 지으면 생성물(fieldsets_templ.go)이 함수명 문자소로 "AgentOverrides" 센티널을 포함해 mcp_audit_surface 가드(생성물 0히트 계약 — §D.5 보존 계약)를 적중한다. 실측으로 확인 후 개명.
- app.js: wireProfileMatrix(G3-3 — 티어 라디오→셀 재설정, dirty 보존, Custom 전환) + applyHaikuEffortLock/wireHaikuEffortLock/reapplyHaikuLocks(haiku→effort 잠금) 재포트 + initConsole 배선. console.css: .tr--afm 3열 그리드 최소 추가.
- settings_shell.go: settingsTabFieldCount("llm") = 스키마 필드 + agentFMRenderCount — 레일-패널 수 일치 불변 유지.
- 생성/운반 표면 정합 3건(실측): (a)TestDataI18nKeysSubsetOfDictionary — 렌더 data-i18n 키의 사전 존재 의무(R6)로 섹션 UI 키 8종을 4-locale 사전에 선수재(M5 잔여 = agentdesc.<12> 번역 + 면제 접두사 + 거버넌스 카운트); (b)TestAppJsFireManifestInventoryCount — change 그룹 3종 증가(10→13)를 appjs_fire_probe.py에 사유 exclusion 3행으로 등록(GLM 잠금과 동형 form-state pairing 근거); (c)mcp_audit_surface 가드는 (위 개명으로) 생성물 0히트 유지 GREEN.
- 테스트: TestHaikuEffortLock·TestGLMReasoningColumn 신설(AC-AFR-009/010 — GLM 열은 template.ResolveGLMReasoningForModel 동일 호출 단언, Claude 백엔드 미렌더 실측), fieldsets_states_test.go 확장(TestAgentFMRowStates 재포트 + TestAgentOverridesSubsectionStates — 파싱 실패 행·빈 목록·Custom 선선택), TestAgentOverridesSubsection의 파싱-실패 행 어설션을 구 표준 형태에 정렬(비편집 행은 data-agent-row 마커 없음 — 테스트 결함 수정).
- §B-4 코멘트 정합: handlers.go model_policy 각주("agent-overrides 프로필 셀렉터의 중복"으로), schemaform.go agentfm-gridnote 관례 각주, restyle_test.go warn 배너 갈라냄 각주. main_test.go:9는 복원 후에도 참이어서 무변경.
- 관측: web 패키지 전량 GREEN(M2 역반전 테스트 전부 GREEN 전환 완료 — TestAgentOverridesSubsection·TestPerfTierSave·TestAgentOverridesSave·TestAgentFrontmatterUntouched). lint 0 issues. go build + GOOS=windows 빌드 exit 0.

### M5 — i18n·가드 재등록 (M4 커밋 49719635e 이후)

- i18n 잔여분: fieldDesc.agentfm.* 11종(model.{inherit,haiku,sonnet,opus,fable}+custom+effort.{low,medium,high,xhigh,max}) 4-locale 재수재 — 구 색상 키(red/orange/blue/lightblue)는 배지가 모델 유도로 재포트되어 참조하지 않아 미수재(계획 수치 13은 plan-phase 산정, run 재판정 11 — §D.5 신규 재판정 원칙). agentdesc.<12 로스터> ko/ja/zh 재수재(manager-todo 신규 작성, mission-governor 키 소멸, 11종은 현역 서술과 일치하는 구 번역 재사용). 섹션 UI 키 8종은 M4에서 선수재 완료.
- i18nEnExemptPrefixes 재등록: agentdesc. — 정당화 문구(en baseline은 frontmatter SSOT를 data-i18n-baseline으로 서버 렌더, en 사전 키는 SSOT 중복) 동반. 거버넌스 전 테스트(TestI18nKeyCoverageForward/Reverse·Parity·Allowlist) GREEN.
- rosterguard/numeral 재판정(§D.2 — 구 노트 미인용, 파일 실측): (a)중복 리터럴 제거 — profileMatrixAgentOrder를 RetainedAgents() 유도(Explore 필터)로 재작성(retained_agents.go 앵커의 "두 번째 수제 로스터 금지" 계약 준수, 테스트도 유도 계약 단언으로 전환); (b)신규 등록 4행 — internal/config/profile.go(retainedAgentNames, retained-roster membership), internal/template/profile_matrix.go(agentGroupMembership, retained-roster membership), internal/web/agentfm.go(agentGroupRank, subset-by-design — 버킷 랭킹은 부분 나열이며 카탈로그 정책 아님), internal/web/assets/i18n.js(agentdesc+agent_tiers 키의 합집합 = 정준 로스터, retained-roster membership — 카탈로그 변경이 번역 키 변경을 강제하는 결합이 콘솔 행이 필요로 하는 것); (3)카운트 산문 정합 — "12-agent catalog/roster" 문구를 정준 계보 서술로 재작성(수치 축 미적중 실측). rosterguard 전 패키지 GREEN.
- 측정 기록: registry ClaimCount 시도 → "32 reachable Count paths but 30 discharged" 산술 불일치 실측(등록된 카운트 행이 도달하는 수치 히트가 없음) → 재산정(prose가 수치 클레임이 아님) → membership 전용 행으로 확정. 신규 파일의 착수 재판정(§D.5)이 실제로 두 번 작동한 사례.

### §E.3 Run-phase Audit-Ready Signal

- run_complete_at: 2026-10-02 (M1→M5 완료, 커밋 5종 — §E.2 마일스톤 행 각각)
- run_status: audit-ready (커버리지·경계 grep·양방향 빌드·lint·web 전체 스위트 실측 — 아래 E-항목 행)
- M5b 미착수(소유자: manager-spec 재위임 — plan §F M5b). M6 전체 검증은 오케스트레이터 몫.

### M5 후행 (AC-AFR-011 보강 + 증거 수집)

- TestAgentOverridesSeams 신설(AC-AFR-011 — §D.11): listAgentFMs/patchAgentFM 치환 시 대체된 seam만 관측(치환된 persist가 소관을 가져 실 llm.yaml 무기록 실측) + 신규 앱의 기본 배선 3종(listAgentFMs·patchAgentFM·applyPerfTierEdits) non-nil 실측. GREEN.
- AC-AFR-012 증거 정합: agent_overrides_test.go 머리 주석이 구 absent 테스트 식별자를 인용 → EV-AFR-012 0-hit grep(1히트 실측) 위반으로 주석을 비-인용 형태로 재작성 → 0히트 실측. `still renders` 문구 결합 마커 5종: agentfm 표면 결합 0히트 실측(나머지 11히트는 타 표면의 무관 부재 테스트).
- 역방향 회귀 관측: TestCodexResolution_IgnoresPerAgentLLMCells GREEN(스폰-경로 무시 계약 생존 — REQ-AFR-002).

- 커버리지 실측(이 트리, 이 런): internal/settings/agentfm 94.4% · internal/settings 87.0% · internal/config 82.8% · internal/template 84.4% · internal/web 73.5%. acceptance.md 간접 검증이 이름하는 2패키지 중 agentfm ≥85% 충족; template은 84.4%로 0.6pt 미달 — 신규 파일 기여분과 기존 레거시 기여분의 구분은 -coverprofile 산출 후 보고(아래 행), 본 SPEC 범위 밖 레거시 코드의 기여분이 있으면 sync-auditor 판정 자료로 남긴다.
- coverprofile 산출(이 트리, 이 런): 신규·재포트 함수 전부 100% — profile_matrix.go(ValidPerformanceTiers·IsValidPerformanceTier·AgentGroup·ProfileMatrixAgents·DefaultProfileMatrix·ResolveAgentModelEffort)와 glm_effort_overlay.go 재포트 4종(ResolveGLMReasoning·ResolveGLMReasoningForModel·IsGLMCodingMaxOverrideAgent·GLMCodingMaxOverrideAgents) 포함. 유일 0%는 GLMReasoningStateNames — 본 SPEC이 건드리지 않은 기존 함수(스키마 위젯 소비). template 패키지 84.4%의 미달분은 레거시 코드 기여로 확인 — 신규 코드 커버리지는 충족.
