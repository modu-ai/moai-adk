# plan.md — SPEC-WEB-AGENTFM-RESTORE-001

## §A Context

카드 t1411 (Class C, 운영자 지시 2026-10-02): `moai web` 서브 에이전트 설정 복원. 삭제 커밋 `384eb3460`(2026-09-27, card t1246)이 지운 콘솔 설정 표면 — 스키마·저장 경로·UI 노출 — 을 현재 트리의 정형 형태로 되살린다. 블루프린트: `.moai/reports/t1411/removed-commit-384eb3460.diff` (7,922행). 개별 슬라이스는 `git show 384eb3460 -- <path>` / `git show 384eb3460^:<path>` 로 도달한다.

개발 모드 TDD(`constitution.development_mode: tdd`). plan-auditor PASS 기준 Tier M 0.80.

### §A.1 블루프린트 지도 (측정 완료 — run-phase 재발굴 불필요)

| 제거물 | 블루프린트 위치 | 복원 형태 |
|---|---|---|
| `internal/settings/agentfm/agentfm.go` (218행) | `git show 384eb3460^:internal/settings/agentfm/agentfm.go` | `List`+`splitFrontmatter`+`AgentInfo`만 — `Patch`/`upsertScalar`/`deleteKey` 쓰기 레이어는 제외 (REQ-AFR-005) |
| `internal/web/agentfm.go` (561행) | `git show 384eb3460^:internal/web/agentfm.go` | parse/persist/렌더 도우미 재포트 — `agentGroupRank`, `agentFMGridRows`, `parseAgentFMForm`, `applyAgentOverrides`, `agentFMModelValues` 등. `agentTierBadge` 글리프는 로컬 재정의(v4manifest 헬퍼 소멸) |
| `internal/web/fieldsets.templ` fieldsetAgentFrontmatter (157행) | diff `-` 블록 | 서브섹션으로 재편 — 마커 `data-section="agent-overrides"`, llm 패널 앵커 |
| `internal/web/handlers.go` save 경로 (-71) | diff | `parsePerfTierForm`·`parseAgentFMForm`·`applyPerfTierEdits`·`applyAgentOverrides` 재배선 — alias 쓰기 제외 |
| `internal/web/app.go` seams (-22) | diff | `listAgentFMs`/`patchAgentFM`/`applyPerfTierEdits` seam 재도입 |
| `internal/web/schemaform.go` (-32) | diff | `consoleTabs`는 13탭 유지(탭 추가 없음), `applySchemaCurrent` 시딩만 복원 |
| `internal/web/assets/app.js` (-102) | diff | `wireProfileMatrix`·haiku lock 재포트 |
| `internal/web/assets/i18n.js` (-148) | diff | 키 38종 재수재 — `agentdesc` 집합은 현재 로스터로 교체 |
| `internal/template/profile_matrix.go` | `git show 3fa8bd2ab^:internal/template/profile_matrix.go` (삭제 커밋 `3fa8bd2ab`, AMI-001 M5) | resolver 기계 재포트 + **셀은 재유도**(아래 §B-1) |
| `internal/template/glm_effort_overlay.go`의 `ResolveGLMReasoningForModel` | `git show 3fa8bd2ab^:internal/template/glm_effort_overlay.go` | GLM reasoning 리졸버 **재포트** — 생존 기계가 아니다(HEAD 정의 0건 실측, plan-audit D3). REQ-AFR-010·AC-AFR-010의 단일-유도 앵커 |
| `config.LLMConfig` 필드군 | `git show 3fa8bd2ab^:internal/config/types.go` | `Profile`/`AgentOverrides`/`ModelEffort`/`EffectiveProfile`/`validateAgentOverrides` |
| 템플릿 `llm.yaml` 키·주석 | `git show 3fa8bd2ab^:internal/template/templates/.moai/config/sections/llm.yaml` | `profile`+`agent_overrides`만 재수재, 주석 블록 갱신 |

## §B Known Issues

1. **5일 드리프트**: (a) 로스터 — 삭제당시 agentdesc 키엔 `mission-governor`가 있고 `manager-todo`가 없다. 현재 `.claude/agents/moai/` 실측 12종 기준으로 재작성. (b) 매트릭스 셀 — 9-27 데이터는 t1368 이전 모델 매트릭스. 셀은 현재 `config` 기본값(claude_models·GLM 티어)에서 **재유도**하고, 블루프린트 셀과의 불일치는 결함이 아니다.
2. **`agent_settings_test.go:51` 고아 주석**: "TestAgentFMWarnI18nParity는…" 한 조각이 본체 없이 남아 있다 — M2에서 정리.
3. **`crosssession_test.go:52-55`**: 노이즈 키 예시로 `agentfm.dev-a.effort`를 쓴다. 복원 후에도 agentfm.*은 crosssession 키가 아니므로 어설션은 유효 — 주석만 갱신.
4. **코멘트 정합**: `handlers.go:633`("duplicate of the agentfm performance tier"), `schemaform.go:181`(agentfm-gridnote 관례), `restyle_test.go:410`, `main_test.go:9` — M4에서 사실 관계로 갱신.
5. **mcp 감사 가드 2종은 별개 파일이다** (plan-audit D1 — 1차 초안이 형제 파일을 혼동): (a) `internal/cli/mcp_audit_test.go:145-151` — `AgentOverrides|agentfm` grep이 **mcp 패키지 파일만** 훑는다(실측) → 웹 레이어 복원과 무충돌, **keep**. (b) `internal/web/mcp_audit_surface_test.go`의 `TestWebConsole_NoPerAgentModelResolver` — `ResolveAgentModelEffort`·`ProfileMatrixAgents`·`EffectiveProfile`·`AgentOverrides` 4 센티널을 **비테스트 internal/web 전 파일**에 대해 검사한다(실측) → 복원된 agentfm 표면이 `applyAgentOverrides`만 가져도 적중한다. **amend** 필수 — 축소 계약은 §D.5.
6. **스키마 외부 live 키**: `llm.profile`/`llm.agent_overrides`는 제네릭 스키마 필드로 등록하지 않는다(구 설계와 동일 — "deliberately NOT part of the generic schema"). `SchemaCurrentValues`에 나타나지 않는 것은 기존 removed-keys 테스트(`settings/schema_sections_test.go:654-682`)와 정합.
7. **`profile_setup_schema_options_test.go:119` 고아 주석** (plan-audit D6): "TestModelPolicyLabels_AgreeWithProfileMatrix was removed…" — 파생 원천(템플릿 매트릭스) 소멸 사유 주석이 본체 없이 남아 있다. M1에서 매트릭스 기계가 재포트되면 이 주석의 서술을 사실 관계로 갱신한다 (§D.5 신규 행).

## §C Pre-flight

- [ ] 작업 트리 확인: `git rev-parse --show-toplevel` = 카드 워크트리. 분기 상태 재판독(`git branch --show-current`).
- [ ] 기준선 측정: `go build ./...` + 아래 스코프 테스트 전부 GREEN 확인 (baseline 기록 → progress.md §E.1):
  `go test ./internal/web/ ./internal/settings/... ./internal/template/ ./internal/cli/ -run 'AgentSettings|RetiredModelKey|UpdateLLMYAML|ConsoleTabs|PrimarySurface|I18n'`
- [ ] LSP 기준선(quality.yaml plan: require_baseline).
- [ ] **원자성 실패-주입 프로브** (plan-audit D9 — codex 2차 소스 발견): 기존 저장 흐름의 persist 단계에 강제 실패를 주입한 뒤 `quality.yaml` 등 **무관 섹션 파일이 byte-unchanged인지 관측**한다. 비원자로 관측되면 M3에 staging/rollback 설계 단계를 추가하거나 REQ-AFR-007을 기존 흐름이 실증 보호하는 표면으로 축소한다. codex 1차 관측: 지속 실패 후 quality.yaml이 byte-identical이 아니었다 — 이 세션에서 미재현(2차 소스 상태). **이 프로브는 run-phase pre-flight(본 절)에서 실행**되며 plan-phase에서는 절차만 확정한다. 주입 차량: §D.5의 `recordingSeams` 하니스(partial_apply_repro_test.go 확장분 — plan-audit iter2 D15).
- [x] 이 SPEC의 디자인 결정 3건(Q1 템플릿 키 재수재 · Q2 스폰-소비 후속 카드 · Q3 llm 패널 배치) — **운영자 확인 완료 (2026-10-02, lane AskUserQuestion 라운드)**. 판정은 decision-index.md Operator verdict 행에 기록됨. Q1·Q3는 §C 기본 처분과 일치(설계 변경 없음), Q2는 후속 카드 발행으로 공백 경계 확정.

## §D Constraints (운영자 지시 5항목 상세 처분)

### §D.1 REQ-AMI-011 계열 단면별 처분 (결정 1)

| 단면 | 현재 상태 | 처분 |
|---|---|---|
| SPEC-AGENT-MODEL-INHERIT-001 REQ-AMI-011 본문 | completed, 콘솔 표면 금지 | 좁은 수정안: 콘솔 표면 예외 문단 추가. REQ-AFR-014 형식(Jev 선례: HISTORY 행, 삭제·재번호 없음, status 불변) |
| 동 SPEC REQ-AMI-013 (템플릿 키 금지) | profile/profiles/harness_agents/agent_overrides/performance_tier 전면 금지 | 좁히기: `performance_tier`·`harness_agents`·workflow 키만 금지로. `profile`/`agent_overrides`는 본 SPEC REQ-AFR-008로 재수재. `profiles`는 수재하지 않음(Go 기본 매트릭스가 SSOT, 사용자 확장 레이어는 선택) |
| 동 SPEC REQ-AMI-014 (update strip) | 전 키 strip | 본문 불변 — `ShippedRetiredModelKeys` seam이 수재 키를 자동 제외(코드 실측). strip 테스트 기대값만 갱신 |
| `internal/web/agent_settings_removed_test.go` | 부재 락 2개 테스트 | **invert** — §D.5 |
| `internal/web/agent_settings_test.go` | role_profiles·workflow_agents 락 | **keep** (agentfm과 무관) + 고아 주석 정리 |
| `internal/settings/schema.go:80`·`schema_sections.go:568` 주석 | "agentfm 제거됨" 서술 | 갱신 — 복원 서술로 |
| `internal/cli/update_model_key_strip.go` 사유 문자열 + `retired_model_keys.go` 문서 주석 | "no reader" 서술 | 갱신 — 콘솔 리더 부활 서술. 로직 불변(seam이 처리) |
| `internal/template/templates/.moai/config/sections/llm.yaml` 은퇴 주석 블록 | "profile matrix retired" | 갱신 — 키 재수재 + 상속 기본 서술 |
| `.claude/rules/moai/development/model-policy.md`·`tech.md:17` | 상속 전용 서술 | sync-phase 정산 (run 코드 확정 후). tech.md의 키 은퇴 목록에서 profile/agent_overrides 제거 |
| SPEC-AGENT-MODEL-INHERIT-DOCS-001 (completed) | 문서 페이지 정산 | 본 SPEC 범위外 — sync-phase에서 후속 카드 권고 |
| `internal/cli/mcp_audit_test.go:145-151` | mcp 패키지 순수성 가드 | **keep** — 웹 레이어와 무충돌(스코프 실측) |

수정안 링크 검증: `internal/contract/kickoff`의 `TestJevAmendmentLinkage`가 시사하듯 수정안은 본문·HISTORY·인용의 분열을 허용하지 않는다 — AC-AFR-013이 grep 계약으로 이를 확인한다.

### §D.2 rosterguard·v4manifest (결정 2) — OUT

블루프린트 임포트 실측: 웹 저장 경로가 v4manifest에서 가져온 것은 closed-set 상수(`ModelInherit`/`EffortLow` 등 — 전부 생존, `schema.go:44-63`)과 표시 헬퍼(`ModelColor`/`ModelColorRank` — 소멸). 상수는 재사용, 헬퍼는 `agentfm.go` 내 로컬 글리프 맵으로 대체. rosterguard 행 5종+exemption 1은 콘솔과 무관 — 미복원. 단 M5에서 복원 파일(`internal/web/agentfm*.go`, `internal/settings/agentfm/`)의 numeral-scope 리스트(`rosterguard/numeral_test.go:241` in-list)와 registry 등록 필요성을 재판정한다 — 등록이 필요해지면 그 행의 근거 노트를 새로 쓴다(구 노트 인용 금지, 파일 실측).

### §D.3 티어 패널 공존 (결정 3)

- 배치: **llm (3rd Party LLM) 패널** 서브섹션. 근거: 저장 대상 파일 정합성(llm.yaml), 13탭 계약 유지, `jevkey.go`가 증명하는 "패널 seam과 다른 persistence를 지닌 커스텀 서브섹션" 패턴.
- 기계 공유: 마커 상수(`agentTierSectionMarker` 대응물), `.subsection` 블록, `partition*` 분리 대신 전용 `agentOverridesSectionFields()` 없음 — 행은 런타임 스캔이라 스키마 FieldDef 정적 목록이 없다. jev credential 블록이 증명하는 커스텀 마크업 경로를 따른다.
- 티어 차트(`workflow` 패널, `agenttierpanel.go`)는 클래스 단위 3티어 할당(workflow.yaml), 본 서브섹션은 에이전트 단위 오버라이드(llm.yaml) — 상보. 패널 상단 help가 서로를 한 줄 참조한다(키 1종 추가).
- 기각안: (a) 14번째 최상위 탭 — `wantTabOrder`·`>13<`·chrome·i18n 락 전면 재협상, 구 tab 기계(`agentFMFieldPrefix` 등) 부활 강요. (b) workflow 패널 배치 — llm 데이터가 workflow 패널에 섞임(파일 비정합).

### §D.4 Re-port 결정 목록 (결정 4)

1. agentdesc 키: 현재 로스터 12종 (manager-todo 포함, mission-governor 제외).
2. 매트릭스 셀: 현재 모델 매트릭스에서 재유도. `profileMatrixData()`는 Go 기본 매트릭스에서 파생(구현과 동일 — `config.LLMConfig{Profile: tier}` 해상).
3. `performance_tier` 레거시 alias 미부활: `applyPerfTierEdits`는 `llm.profile`만 쓴다 (`template.ApplyPerformanceTier` 호출 제거 — 소비자 소멸).
4. 프론트매터 쓰기 없음 — `internal/settings/agentfm`은 `List`만.
5. 탭 없음 — 서브섹션. `settings_shell.go`의 tab desc/effect/field-names 복원 불요, llm 패널 필드 카운트에 렌더 행수 합산만 추가.
6. 포트 선례 t1385/t1388: 대상 트리의 규약 형태로 재작성 — 구현 직전 primary 최신본 재판독.
7. `template.ResolveGLMReasoningForModel`은 **재포트 대상**이다(생존 기계 아님 — HEAD 정의 0건 실측, plan-audit D3). 원천: `git show 3fa8bd2ab^:internal/template/glm_effort_overlay.go`. REQ-AFR-010의 "no second derivation"은 재포트된 이 함수에 대한 단일-유도 단언으로 유지한다.

### §D.5 락 테스트 처분표 (결정 5)

표면-결합 테스트 파일 26종의 명시적 열거(spec.md §C-5 — 전수 서술이 아닌 열거; 실측 스캔 기준, 신규 파일은 run-phase 착수 시 재판정).

| 파일 | 처분 | 근거 |
|---|---|---|
| `agent_settings_removed_test.go` | **invert** — "존재 안 함" 5 마커 → 존재+효과 테스트(REQ-AFR-001/003/004), 위조 POST 테스트는 "frontmatter 무접촉"만 유지하고 llm.yaml 불변 절반은 합법 제출 시 효과 검증으로 대체 | RED 기선 — 복원의 관측-RED |
| `agent_settings_test.go` | **keep** (role_profiles·workflow_agents) + :51 고아 주석 제거 | 대상 표면이 agentfm이 아님 |
| `console_ux_fix_test.go` | **keep** — llm 탭 개칭 어설션은 복원과 무충돌. G2-2 코멘트의 "single-panel agentfm" 서술만 갱신 | 실측: sec.llm.title 어설션 |
| `tab_layout_test.go` | **keep-and-why** — 13탭 시퀀스 유지가 본 설계의 전제. `wantTabOrder` 무변경 통과가 AC-AFR-008 | 서브섹션 선택의 직접 귀결 |
| `primary_surface_test.go` | **keep-and-why** — `>13<` 헤더 계약 유지 | 상동 |
| `i18n_untranslated_allowlist_test.go` | **replace** — `i18nEnExemptPrefixes` 빈 등록소에 `agentdesc.` 재등록(정당화 문구 필수, `i18nMaxAllowlistEntries=30` 별도) | 구成员 복원, 등록 행위는 리뷰 대상(주석 규약) |
| `i18n_governance_test.go` | **extend** — 신규 키 4-locale 정합 카운트 갱신 | 거버넌스 규약 |
| `fieldsets_states_test.go` | **extend** — 서브섹션 상태(파싱 실패 행·빈 목록) 추가 | 삭제당시 -49행 분의 재포트 |
| `crosssession_test.go` | **keep** + 주석 갱신 | §B-3 |
| `internal/cli/mcp_audit_test.go` | **keep** | mcp 패키지 한정 가드(실측 :145-151) — 웹 레이어 복원과 무충돌 |
| `internal/web/mcp_audit_surface_test.go` | **amend** — `TestWebConsole_NoPerAgentModelResolver` 센티널 계약 축소 | 4 센티널(`ResolveAgentModelEffort`·`ProfileMatrixAgents`·`EffectiveProfile`·`AgentOverrides`)은 블루프린트 실측 센티널-보유 표면 파일(`agentfm.go`·`app.go`·`handlers.go`·`schemaform.go`)에서 **허용**하고, 그 외 비테스트 internal/web 전 파일(센티널 0히트인 생성물 `fieldsets_templ.go` 포함)에서는 금지를 **유지**한다. 허용 집합 근거(레저 diff): `app.go` :3100 `patchAgentFM` 배선·`handlers.go` :6488 `template.ResolveAgentModelEffort` 호출·`schemaform.go` :7098/:7115 `EffectiveProfile()`·`len(cfg.LLM.AgentOverrides)` — `fieldsets_templ.go`(diff :3916-6222)는 센티널 0히트라 허용 불요. plan-audit iter2 D14. 보존되는 축소 계약: (1) internal/web은 자체 유도 함수를 정의하지 않는다(`func ResolveAgentModelEffort` 정의 전 파일 금지 — 단일-유도, REQ-AFR-010과 정합), (2) 제외 표면 집합이 비면 가드 실패(표면 삭제로 가드를 통과하는 것 금지). 블루프린트 실측: 구 web/agentfm.go도 `template.ResolveAgentModelEffort`를 호출했으므로(레저 diff :287 이하) 계약은 "호출 허용·정의 금지"가 정확하다 |
| `update_llm_preserve_test.go`·`update_model_key_strip_test.go`·`retired_model_keys_test.go`·`shipped_key_reader_test.go` | **extend** — 수재 키 생존 + `performance_tier`/`harness_agents`/workflow 키 계속 strip 기대 | REQ-AFR-008 |
| `settings/schema_sections_test.go`·`web/schema_sections_test.go` | **keep** — removed-key 무시 계약은 schema-external live 키와 정합 | §B-6 |
| `webux_followup_test.go`·`webux_haiku_effort_test.go` | **replace** — 소멸. 해당 AC(haiku lock 등)는 신규 테스트로 흡수 | 통째 재생성 금지(Out of Scope) |
| `internal/web/m4_description_test.go` | **keep** | :68/:84 — `performance_tier`/`claude_models`가 baked/hidden으로 유지된다는 서술은 복원 후에도 참(서브섹션은 스키마 필드가 아님 — §B-6). plan-audit D6 |
| `internal/cli/update_retained_advisory_test.go` | **keep** | :33 — 렌더 형태만 검사; 수재 키 집합 변화는 `ShippedRetiredModelKeys` seam으로 자동 적응. plan-audit D6 |
| `internal/cli/update_user_keys_survey_test.go` | **keep** | :81 — seam-adaptive: 수재 키는 strip 제외되므로 스킵 경로가 자동 적응. plan-audit D6 |
| `internal/cli/update_clean_install_config_preserve_test.go` | **keep** | :106/:231 — 클린 설치 보존 경로; 재수재 키는 strip 대상이 아니므로 계약 불변. plan-audit D6 |
| `internal/cli/update/backup/merge_useradd_test.go` | **keep** | :9-30 — 합성 픽스처 기반 백업 병합 단위 테스트, 표면 토큰과 무관. plan-audit D6 |
| `internal/cli/model_backend_default_test.go` | **keep** | :13-51 — 스폰-경로 무시 계약(`TestCodexResolution_IgnoresPerAgentLLMCells` :28). 런타임 소비가 Out of Scope인 한 계약 생존 — Q2 후속 카드가 소비를 도입하면 그 SPEC이 재판정. plan-audit D6 |
| `internal/cli/profile_setup_schema_options_test.go` | **keep** + :119 고아 주석 갱신 | 매트릭스 재포트(M1) 후 주석 서술을 사실 관계로 갱신 — §B-7. plan-audit D6 |
| `partial_apply_repro_test.go` | **amend/extend** — saveStep 상수·`injectableSteps`·`recordingSeams`·양성 대조 `want`(:130-132)에 제거된 7/8 단계(`patchAgentFM`·`applyPerfTierEdits`) 재등록 | M3가 seam 호출을 재도입하면 `TestPartialApplyOrderPositiveControl`이 기계적으로 깨짐 — 레저 diff :6807-6876 제거분의 역방향. 현재 단계 종단 `:59-64 stepApplySchema` 실측. plan-audit iter2 D15 |
| `save_observability_test.go` | **amend/extend** — `seamTable`·seam 목록을 같은 확장에 동행 | :28 스스로 종속 형제를 자칭 — 모든 저장 seam을 `recordingSeams`로 구동. plan-audit iter2 D15 |

## §E Self-Verification

- [ ] 모든 AC가 §acceptance.md 매트릭스에서 검증 명령 또는 증거 형태를 이름으로 지시한다.
- [ ] REQ 14 ≤ 16, AC 13 ≤ 16 (Tier M 상한).
- [x] decision-index FOUNDER 행 3건(Q1·Q2·Q3) 운영자 판정 기록 완료 (2026-10-02) — plan.md의 미해결-질문 마커 3건은 해소·제거됨 (plan-audit D2).
- [ ] 블루프린트 인용은 전부 실측 커밋 해시·경로로 검증됨 (`384eb3460`, `3fa8bd2ab`, `238219302`).

## §F Milestones (단일 구현 에이전트 기준, TDD 사이클)

### M1 (High) — 데이터 모델·해상 기계 복원 [결정 역전 가능성 최상 — 최우선]

- `internal/config/types.go`: `LLMConfig`에 `Profile`/`AgentOverrides` 재도입 + `ModelEffort`·`EffectiveProfile()`·`validateAgentOverrides` (블루프린트: `3fa8bd2ab^`).
- `internal/template/profile_matrix.go` 재포트: `ResolveAgentModelEffort`·`ProfileMatrixAgents`·`AgentGroup`·`ValidPerformanceTiers`·`IsValidPerformanceTier`. **셀은 현재 모델 매트릭스에서 재유도**하고 셀 테스트는 `config` 현재 기본값을 단일 원천으로 단언.
- 템플릿 `llm.yaml`: `profile: ""` + `agent_overrides: {}` 재수재 + 주석 블록 갱신(상속 기본·오버라이드 설명).
- **shipped-keys 리더 인벤토리 등록** (plan-audit D10): 재수재 키 `llm.profile`·`llm.agent_overrides`를 `internal/config/testdata/shipped_key_inventory.yaml`에 리더(R) 등록 — `TestShippedConfigKeysHaveReaders`(`internal/config/shipped_key_reader_test.go:74`) 계약. 미등록 시 update가 콘솔이 쓰는 키를 인벤토리 미일치로 적린다.
- **`ResolveGLMReasoningForModel` 재포트** (§D.4-7, plan-audit D3): 원천 `git show 3fa8bd2ab^:internal/template/glm_effort_overlay.go`. 모델 집합 상수 원천 — v4manifest closed-set 4종(`schema.go:62-73`) + config `Fable`(`internal/config/types.go:367`); v4manifest에 fable 상수는 없다.
- strip/advisory 정합: 4개 테스트 기대 갱신(§D.5) — `performance_tier`·`harness_agents`·workflow 키는 계속 strip됨을 확인.
- RED: 셀 해상·strip 생존 테스트 선작성.

### M2 (High) — RED 기선: 락 반전

- `agent_settings_removed_test.go` → `agent_settings_test.go`(파일명 재사용 or 신규) 존재+효과 테스트로 반전 (§D.5 1행).
- `mcp_audit_surface_test.go` RED 관측 (§D.5 amend 행, plan-audit D1): 복원 agentfm 표면 파일 유입 시 `TestWebConsole_NoPerAgentModelResolver`가 4 센티널 중 1건 이상 적중으로 RED가 됨을 관측한 뒤, 축소 계약(표면 파일 제외 + 정의-금지 + 제외집합 비-공 가드)으로 개정해 GREEN 전환한다. RED 출력은 progress.md §E.2에 전사.
- 저장 경로 테스트 골격: tier 저장·override pin/clear·원자 거절·frontmatter 무접촉.
- 이 시점 전부 RED를 관측하고 progress.md에 기록.

### M3 (High) — 저장 경로

- `internal/settings/agentfm` 복원(List 전용) + 단위 테스트(파싱 실패 행·빈 디렉터 저하).
- `handlers.go`: `parsePerfTierForm`·`parseAgentFMForm`·persist 단계 재배선 — alias 쓰기 없음, `llm.profile`만.
- `app.go` seams, `schemaform.go` `applySchemaCurrent` 시딩.
- GREEN: M2 테스트 전환 관측.

### M4 (Medium) — UI 노출

- `fieldsets.templ`: llm 패널 서브섹션(마커 `data-section="agent-overrides"`) + `agentFMRow` 재포트 + `fieldsets_templ.go` 재생성(templ v0.3.1020).
- `settings_shell.go`: llm 패널 필드 카운트에 렌더 행수 합산.
- `fieldsets.templ`: `templ.JSONScript("moai-profile-matrix", profileMatrixData())` island 재포트 — 클라이언트 tier 재설정(`wireProfileMatrix`)의 데이터 원천, 부재 락이던 `id="moai-profile-matrix"` 마커의 반전 대상.
- `app.js`: `wireProfileMatrix` + haiku lock 재포트. `console.css` 최소 추가.
- 코멘트 정합(§B-4). GREEN: 렌더 마커 테스트 전환.

### M5 (Medium) — i18n·가드 재등록 [기계적 — 후순위]

- i18n 4-locale: `agentfm.*`(8)·`sec.agentfm.*`(2)·`fieldDesc.agentfm.*`(13)·`agentdesc.<12 로스터>`(ko/ja/zh, en은 baseline).
- `i18nEnExemptPrefixes` 재등록 + 거버넌스 카운트 갱신.
- rosterguard/numeral 재판정(§D.2) + 구 선례 인용 없이 새 근거 노트.

### M5b (High) — 상위 SPEC 수정안 적용 [소유자: manager-spec 재위임 — plan-audit D8]

- **소유자 경계**: run-phase 에이전트는 SPEC 본문 수정이 금지되어 있다(spec-frontmatter-schema § Forbidden ownership crossings). 오케스트레이터가 D-NEW-1 인라인-수정 패턴으로 **manager-spec에 재위임**하여 본 단계를 수행한다 — manager-develop가 M5 완료 보고 시 블로커로 반환하고, 오케스트레이터가 재위임 → 재복귀 시퀀스를 진행한다. **명시적 직렬 순서: M5 → M5b → M6** — M6의 DoD 확인 이전에 완료되어야 한다.
- 수정 대상: `SPEC-AGENT-MODEL-INHERIT-001/spec.md` — (1) HISTORY 신규 행, (2) REQ-AMI-011에 본 SPEC을 향한 콘솔-표면 예외 문단 추가, (3) REQ-AMI-013을 잔여 금지 키 집합(`performance_tier`·`harness_agents`·workflow 키)으로 축소. REQ-AMI-013 수정안은 decision-index Q1 운영자 승인(2026-10-02)이 동반한다.
- 형식: SPEC-JEV-CORE-001 HISTORY v0.2.0/v0.3.0 선례 — 요구사항 삭제·재번호 없음, `status`·`sync_commit_sha` 불변. 검증은 AC-AFR-013(acceptance.md §D.13)이 담당.

### M6 (High) — 전체 검증·LSP

- 스코프 테스트 전량 + LSP 게이트(run: 오류 0) + §acceptance.md DoD 체크리스트.
- 커버리지: 신규 패키지 85%+.

## §G Anti-Patterns

- 블루프린트 diff를 패치로 그대로 적용 금지 (re-port 원칙, §D.4).
- `performance_tier` alias·frontmatter 쓰기·14번째 탭 되살리기 금지.
- `wantTabOrder`·`>13<` 등 탭 계약 어설션 수정 금지 — 통과가 설계의 증명.
- `internal/settings/agentfm`의 `Patch` 쓰기 레이어 부활 금지 (죽은 코드).
- 스키마 필드 등록으로 `llm.profile`을 넣지 마라 — schema-external live 키 유지(§B-6).
- 가드 등록 시 구 rosterguard 노트 인용 금지 — 새 파일 실측으로 재작성.

## §H Cross-References

- SPEC 본문: `spec.md` §C (5 결정 요약), §Out of Scope.
- decision-index.md: Q1(템플릿 키 재수재 비준)·Q2(스폰 시점 소비)·Q3(패널 배치) — FOUNDER.
- 블루프린트: `.moai/reports/t1411/removed-commit-384eb3460.diff`.
- 선례: SPEC-JEV-CORE-001 HISTORY v0.2.0/v0.3.0 (completed-SPEC 좁은 수정안 형식), `internal/contract/kickoff/activation_test.go` `TestJevAmendmentLinkage`.
