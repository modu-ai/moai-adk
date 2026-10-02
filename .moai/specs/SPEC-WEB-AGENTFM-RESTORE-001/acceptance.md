# acceptance.md — SPEC-WEB-AGENTFM-RESTORE-001

Tier M 인수 계층. 각 AC는 Given-When-Then으로 기술되고 검증 명령 또는 증거 형태를 이름으로 지시한다. 모든 명령은 카드 워크트리 루트에서 실행한다.

## §D AC 매트릭스

| AC | 요구 | 제목 | 검증 |
|---|---|---|---|
| AC-AFR-001 | REQ-AFR-001 | 서브섹션 렌더 (12행·마커) | `go test ./internal/web/ -run 'TestAgentOverridesSubsection'` |
| AC-AFR-002 | REQ-AFR-003 | tier 선택 → llm.profile 저장 | `go test ./internal/web/ -run 'TestPerfTierSave'` |
| AC-AFR-003 | REQ-AFR-004/002 | override pin/clear + 상속 기본 | `go test ./internal/web/ -run 'TestAgentOverridesSave'` |
| AC-AFR-004 | REQ-AFR-006/007 | 집합 외 값 원자 거절 | `go test ./internal/web/ -run 'TestAgentOverridesSave'` (거절 케이스) |
| AC-AFR-005 | REQ-AFR-005 | 프론트매터 무접촉 | `go test ./internal/web/ -run 'TestAgentFrontmatterUntouched'` |
| AC-AFR-006 | REQ-AFR-008 | update strip 정합 | EV-AFR-006 레저 (4커맨드) |
| AC-AFR-007 | REQ-AFR-011 | i18n 4-locale + agentdesc 면제 | `go test ./internal/web/ -run 'TestI18n'` |
| AC-AFR-008 | REQ-AFR-013 | 13탭 계약 무변경 | EV-AFR-008 레저 (1커맨드) |
| AC-AFR-009 | REQ-AFR-009 | haiku effort 잠금 | `go test ./internal/web/ -run 'TestHaikuEffortLock'` |
| AC-AFR-010 | REQ-AFR-010 | GLM reasoning 표시 | `go test ./internal/web/ -run 'TestGLMReasoningColumn'` |
| AC-AFR-011 | REQ-AFR-012 | 저장 seam 주입 가능 | `go test ./internal/web/ -run 'TestAgentOverridesSeams'` |
| AC-AFR-012 | REQ-AFR-001/013 | 락 반전 완결 | grep 증거: 구 absent 어설션 0히트 + 신규 존재 어설션 목록 (§D.12) |
| AC-AFR-013 | REQ-AFR-014 | 수정안 링크 | grep 증거: AMI-001 HISTORY 행 + 본문 예외 문단 (§D.13 — 소유 마일스톤 M5b) |

## 검증 커맨드 레저 (evidence ledger)

표 셀은 셸 메타문자(`|`)를 망가뜨린다(verification-completeness §2.1 — 레저가 권장 운반체). 파이프 결합 선택자가 든 커맨드는 이 레저로 인용한다. 모든 커맨드는 카드 워크트리 루트에서 **단일 호출**로 실행한다(`&&` 체이닝 없음).

빈 스윕 금지 (verification-completeness §1.1): 각 실행의 증거 출력에서 `[no tests to run]`이 관측되면 **실패**로 판정한다. 실행 테스트 수 하한 — EV-AFR-006 ≥ 4 (행당 1건 이상), EV-AFR-008 ≥ 3 (선택자 3명 일치), EV-AFR-012는 0히트 기대(grep exit 1이 정상 패스, exit 2는 오류). `ok`와 함께 0이 아닌 테스트 카운트가 보여야 PASS로 읽는다.

```
EV-AFR-006 (AC-AFR-006 — REQ-AFR-008, 실행 테스트 ≥ 4):
go test ./internal/template/ -run 'RetiredModelKeys'
go test ./internal/config/  -run 'TestShippedConfigKeysHaveReaders'
go test ./internal/cli/     -run 'TestUpdateLLMYAML'
go test ./internal/cli/     -run 'TestStripRetiredModel'

EV-AFR-008 (AC-AFR-008 — REQ-AFR-013, 실행 테스트 ≥ 3):
go test ./internal/web/ -run 'TestConsoleTabsOrder|TestTabPanelRenderOrderMatchesTabs|TestSettingsPageHeaderKeepsAllThirteenTabs'

EV-AFR-012 (AC-AFR-012 — §D.12, 0히트 기대 = exit 1):
grep -rn 'TestAgentSettingsTab_IsNotRendered\|TestAgentSettingsFields_ArePostedWithoutEffect' internal/web/
grep -rn 'still renders' internal/web/*_test.go
```

선택자 근거 (plan-audit D4/D13): `TestShippedConfigKeysHaveReaders` = `internal/config/shipped_key_reader_test.go:74` — 1차 초안의 `TestShippedKeyReader`는 존재하지 않는 이름이었고 패키지도 틀렸다(internal/cli → internal/config). `TestUpdateLLMYAML*` = `internal/cli/update_llm_preserve_test.go:229+` — 1차 초안의 `TestUpdateLlmPreserve`는 대소문자 불일치로 0매치. 템플릿 strip = `TestStripRetiredModelKeys_*`(internal/template), CLI strip = `TestStripRetiredModelConfig_*`(internal/cli). 1차 초안 커맨드의 중복 `./internal/template/` 패키지 인자는 제거되었다.

## §D.1 AC-AFR-001 — 서브섹션 렌더

- **Given** 12종 moai 에이전트 정의가 `.claude/agents/moai/`에 있고 파싱 가능한 **When** `GET /settings` (llm 탭) **Then** 응답 본문에 `data-section="agent-overrides"` 블록이 있고, 각 에이전트마다 `data-agent-row="<name>"` 행과 `name="agentfm.<name>.model"` / `name="agentfm.<name>.effort"` select가 렌더되며, `sec.agentfm.title` 라벨이 존재한다.
- **Given** frontmatter 파싱이 실패하는 에이전트 파일 **When** 렌더 **Then** 해당 행은 `agentfm.unavailable`로 강등되고 select이 없다 (전체 페이지 실패 금지).
- **Given** `.claude/agents/harness/`의 전문가 파일 **When** 렌더 **Then** 행이 렌더되지 않는다 (스캔은 하되 미렌더).

## §D.2 AC-AFR-002 — tier 저장

- **Given** llm.yaml에 `profile` 키가 없는 상태 **When** `POST /save`가 `performance_tier=max`를 제출 **Then** llm.yaml에 `llm.profile: "max"`가 기록되고 `llm.performance_tier` 키는 생성되지 않는다.
- **Given** 동일 상태 **When** `performance_tier=custom` 또는 생략 **Then** llm.yaml은 byte-identical이다 (preserve).

## §D.3 AC-AFR-003 — override pin/clear

- **Given** 기본 프로필 셀이 (sonnet, high)인 에이전트 **When** `agentfm.<name>.model=opus`·`agentfm.<name>.effort=xhigh` 제출 **Then** `llm.agent_overrides`에 `{<name>: {model: opus, effort: xhigh}}`가 생긴다.
- **Given** 동일 에이전트에 override가 있는 상태 **When** 프로필 기본값과 동일한 쌍을 제출 **Then** 해당 override가 제거되고 다른 에이전트의 override는 보존된다.
- **Given** 제출에 변화가 없는 상태 **When** 저장 **Then** llm.yaml은 byte-identical이다 (no-op 스킵).
- **Given** 매트릭스 비멤버 이름 제출 **When** 저장 **Then** override가 생기지 않는다.
- **Given** `llm.agent_overrides`에 항목이 없는 에이전트 **When** 행 해상 **Then** 렌더는 프로필-매트릭스 셀 값으로 해상한다(REQ-AFR-001) — pin/clear 대비 기준도 동일한 프로필-매트릭스 해상이다(REQ-AFR-004, no-op byte-identical 보존). REQ-AFR-002의 세션 model/effort 해상은 스폰 경로 한정이며(런타임 소비는 Out of Scope — decision-index Q2, 후속 카드 t1421) 콘솔 표시와 혼동하지 않는다.
- **Given** 복원된 저장 경로 **When** `go test ./internal/cli/ -run 'TestCodexResolution_IgnoresPerAgentLLMCells'` (`internal/cli/model_backend_default_test.go:28`) **Then** 스폰-경로 무시 계약이 green으로 유지된다 — 본 SPEC이 스폰 경로를 바꾸지 않음을 관측 (REQ-AFR-002, regression-guard; `[no tests to run]`은 실패).

## §D.4 AC-AFR-004 — 원자 거절

- **Given** `agentfm.<name>.effort=absurd` (집합 외) **When** 저장 **Then** 응답은 200 재렌더 + `agentfm.<name>.effort` 필드 오류를 달고, llm.yaml과 agent frontmatter는 byte-identical이다.
- **Given** `performance_tier=bogus` **When** 저장 **Then** `performance_tier` 필드 오류 + 전체 원자 거절.

## §D.5 AC-AFR-005 — 프론트매터 무접촉

- **Given** `.claude/agents/moai/manager-develop.md` 원본 bytes **When** 합법 제출(tier 변경 + override pin 포함) **Then** 파일은 byte-identical이다 (REQ-MPM-040 승계 — 콘솔은 frontmatter를 쓰지 않는다).

## §D.6 AC-AFR-006 — update strip 정합

검증 커맨드: EV-AFR-006 레저 (4커맨드, 빈 스윕 금지 — 실행 테스트 ≥ 4).

- **Given** 수재된 템플릿 llm.yaml이 `profile`·`agent_overrides`를 운반 **When** `template.ShippedRetiredModelKeys()` **Then** 반환 집합에 두 키가 **포함**된다(멤버십 = strip 제외 — `retired_model_keys.go:87-98`은 템플릿에서 발견된 키의 집합을 반환한다; plan-audit iter2 D16 극성 교정) **그리고** strip 실행 결과에서 두 키는 제거되지 않는다. 음(역)측 절반은 다음 bullet: 미수재 은퇴 키는 반환 집합에 있으면서 strip 대상이다.
- **Given** `llm.performance_tier: high`만 지닌 사용자 llm.yaml **When** strip 단계 실행 **Then** 해당 키는 제거되고 백업에 원본이 남는다 (여전히 은퇴 키).
- **Given** `agent_overrides`에 사용자 값 **When** (미래의) strip이 실행되는 경우 **Then** `UserValues` 플래그 동작은 기존 계약 유지.

## §D.7 AC-AFR-007 — i18n

- **Given** 4-locale 사전 **When** 거버넌스 테스트 **Then** 신규 `agentfm.*`·`sec.agentfm.*`·`fieldDesc.agentfm.*` 키가 4 locale 모두 존재하고, `agentdesc.<name>`은 ko/ja/zh에만 존재하며 en은 `data-i18n-baseline`(frontmatter description)으로 복원된다.
- **When** `i18nEnExemptPrefixes` 검사 **Then** `agentdesc.` 행이 정당화 문구와 함께 등록되어 있다.

## §D.8 AC-AFR-008 — 13탭 계약

- **When** EV-AFR-008 레저 커맨드(`TestConsoleTabsOrder`·`TestTabPanelRenderOrderMatchesTabs`·`TestSettingsPageHeaderKeepsAllThirteenTabs` — 실행 테스트 ≥ 3) **Then** `wantTabOrder`(13)·`>13<` 어설션이 **무수정으로** 통과한다 — 어설션 파일의 변경 자체가 실패 증거다.

## §D.9 AC-AFR-009 — haiku 잠금

- **Given** 해상 모델이 haiku인 행 **When** 렌더 **Then** effort select이 `disabled`로 방출되고 `data-haiku-hint` 노트가 노출된다.
- **When** 클라이언트에서 모델 select 변경 **Then** 같은 행(`[data-agent-row]` 스코프)의 effort 잠금이 재적용된다 (다른 행 오염 금지 — 구 `applyHaikuEffortLock` 계약).

## §D.10 AC-AFR-010 — GLM reasoning

- **Given** `team_mode: glm` **When** 렌더 **Then** 각 행에 `data-glm-reasoning` 표시가 붙고 그 값은 `template.ResolveGLMReasoningForModel`과 동일하다 (2차 유도 금지 — 동일 함수 호출 단언).
- **Given** Claude 백엔드 **When** 렌더 **Then** 표시가 없다.

## §D.11 AC-AFR-011 — seam

- **When** 테스트가 `listAgentFMs`/`patchAgentFM`(persist) seam을 대체 **Then** 기본 배선 (`agentfm.List`·`applyAgentOverrides`)은 무변경이고 대체된 seam만 관측된다.

## §D.12 AC-AFR-012 — 락 반전 완결 (증거 형태 — 반전 후 상태 핀, plan-audit D5)

M2의 반전은 파일명 재사용 또는 신규 파일 둘 다 허용되므로, 증거는 **반전 후 상태**에 핀한다. 구 파일 경로(`internal/web/agent_settings_removed_test.go`)를 대상으로 한 grep은 파일이 삭제·개명되면 0히트가 아니라 오류(exit 2)가 되므로 증거로 쓰지 않는다 — 이 지점이 1차 초안의 결함이었다.

- 구 absent-모드 테스트명 소멸 (삭제·개명·전파 전부에서 유효한 0히트): EV-AFR-012 레저 1행 → 0히트 (grep exit 1 = 정상 패스, exit 2 = 오류).
- 구 absent 모드의 5종 gone 마커(`sec.agentfm.title`, `name="agentfm.`, `name="performance_tier"`, `id="moai-profile-matrix"`, `data-agent-row=`)가 **부재-단정 문맥**으로 잔존하지 않음: EV-AFR-012 레저 2행(`still renders` 오류 문구 결합) → 0히트.
- 신규 존재 어설션 목록 (파일+행+극성)을 progress.md §E.2에 전사 — `data-agent-row` 등 5종 마커의 극성 반전 대응.
- 반전 테스트 실행 수 ≥ 2 기록 (구 absent 2테스트의 대체분) — `[no tests to run]` 금지.
- `agent_settings_test.go:51` 고아 주석 제거 확인.

## §D.13 AC-AFR-013 — 수정안 링크 (증거 형태)

소유 마일스톤: **M5b** (plan.md §F — manager-spec 재위임 단계; run-phase 에이전트의 SPEC 본문 수정 금지로 인한 명시적 재위임, plan-audit D8). 직렬 순서 M5 → M5b → M6 — M6 DoD 이전 완료.

- `grep -n "AGENTFM-RESTORE" .moai/specs/SPEC-AGENT-MODEL-INHERIT-001/spec.md` → HISTORY 신규 행 1히트 이상.
- 동 SPEC 본문 REQ-AMI-011에 본 SPEC을 향한 예외 문단 존재, REQ-AMI-013의 축소 문구 존재.
- 불변 확인: 동 SPEC `status: completed`·`sync_commit_sha` 미수정 (`git diff` 빈 출력).
- `TestJevAmendmentLinkage` 스타일 정합: 본문·HISTORY·인용 3자 분열 없음.

## 간접 검증 (indirect)

- `go test ./internal/web/ ./internal/settings/... ./internal/template/ ./internal/config/` 전량 GREEN — 특히 `settings/schema_sections_test.go` removed-key 계약(§B-6)과 `crosssession_test.go` 무변경 통과.
- `golangci-lint run internal/web/... internal/settings/agentfm/...` + `gofmt` (Unified).
- 커버리지: `go test -cover ./internal/settings/agentfm/ ./internal/template/` ≥ 85%.

## 에지 케이스

- agents 디렉터 부재 (greenfield) → 빈 목록 저하, 페이지 200.
- frontmatter 구분선 없는 .md → ParseOK=false 행 강등.
- llm.yaml 로드 실패 → 렌더는 zero-config 저하(구 design §C.1), 저장은 원자 거절.
- effort 미제출 (disabled select) → resolved 값 backfill.
- 템플릿에 `agent_overrides: {}`가 있어도 strip `UserValues` 비교는 빈 맵=사용자 값 아님 유지.

## 품질 게이트 (TRUST 5)

- **Tested**: 위 AC 전부 + 커밋당 80% (quality.yaml tdd_settings).
- **Readable/Unified**: 구현 전 primary 최신본 재판독(t1385/t1388 선례) — 네이밍·주석 언어(code_comments: en) 일치.
- **Secured**: 폼 이름 경로 순회 가지(`agentfm.<name>`에 `/`·`..` 거부 — 구 구현 계약 승계), frontmatter 읽기 전용.
- **Trackable**: Conventional 커밋 + SPEC ID 참조.

## Definition of Done

1. AC 매트릭스 13/13 PASS (증거 명령 출력 전사 → progress.md §E.2).
2. decision-index Q1-Q3 운영자 확인 **완료** (2026-10-02, decision-index Operator verdict 기록 — Q1 승인 · Q2 후속 카드 발행 · Q3 llm 패널 확정; 전부 §C 기본 처분과 일치).
3. Out of Scope 3개 항목 미침수 — `git diff --stat` 이 스코프 파일만.
4. LSP 게이트 run 기준(오류 0) 통과.
