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
