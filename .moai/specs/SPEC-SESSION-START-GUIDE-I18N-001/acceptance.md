# acceptance — SPEC-SESSION-START-GUIDE-I18N-001

id: SPEC-SESSION-START-GUIDE-I18N-001
card: t1603
created: 2026-10-08

## §D AC 매트릭스

RED-now 증거 원장(evidence-ledger) — 아래 셀은 이 원장을 id 로 인용한다. 측정: 본 워크트리, tree SHA `81786284e6e5ee2fe5fbe7b498d649b27c3e0109`, 2026-10-08, 테스트 원본 `internal/hook/session_start_guide_i18n_red_test.go`.

- LEDGER-RED-AC001 — command: `go test ./internal/hook -run TestRedAgentChannelFollowsConversationLanguage -v -count=1` · exit code: `1` · verbatim stdout (결정 행):

  ```
  session_start_guide_i18n_red_test.go:48: additionalContext (agent-facing channel) must render ko locale prose, got:
          (이하 영어 리더 안내 전문 — "Factory Mode: run red001, leader session." 로 시작)
  --- FAIL: TestRedAgentChannelFollowsConversationLanguage (0.87s)
  FAIL
  FAIL	github.com/modu-ai/moai-adk/internal/hook	1.869s
  FAIL
  ```

  적색 사유: additionalContext 가 ko 설정에도 영어 안내를 렌더 — session_start.go:518 의 langEnglish 하드코딩이 표적 그 자체다(M2 가 뒤집는 이유와 일치).

- LEDGER-RED-AC002 — command: `go test ./internal/hook -run TestRedLaneJoinAuthorityLocalized -v -count=1` · exit code: `1` · verbatim stdout (D11-a 강화판 — 부재 단정이 접두 + 영어 본문 단편 둘 다를 검사):

  ```
  session_start_guide_i18n_red_test.go:90: ko lane join must carry the localized authority sentence, English prefix or body still present
  session_start_guide_i18n_red_test.go:90: ja lane join must carry the localized authority sentence, English prefix or body still present
  session_start_guide_i18n_red_test.go:90: zh lane join must carry the localized authority sentence, English prefix or body still present
  --- FAIL: TestRedLaneJoinAuthorityLocalized (0.00s)
  FAIL
  FAIL	github.com/modu-ai/moai-adk/internal/hook	1.269s
  FAIL
  ```

  적색 사유: laneSpawnAuthority 상수(lane_spawn_authority.go:41)가 영어 단일 문자열 — M1 테이블 이항이 뒤집는다. 매트릭스 경로 긍정 단정 4건은 오늘 통과한다(현 트리의 ko/ja/zh join도 같은 영어 문장을 덧붙여 경로를 갖는다) — 이 단정들은 M1 이후 변이(문장 삭제·빈 로케일 항목)를 잡는 포획기다. **검증 범위 배정(D12-a): 본 AC-002 테스트는 현지화 존재(접두+본문 부재, 매트릭스 경로)만 검증한다. 현지화된 authority 문장의 하위 내용 품질(위임 매핑·depth-1 봉인·동료 불변식의 번역 충실도)은 AC-009의 release-blocking 등급 게이트가 소관이다 — 경로만 옮긴 텍스트(직역 계어)는 AC-009에서 등급 F로 떨어져 종결을 막는다. plan 단계에서 M1 미작성 텍스트에 긍정적 하위 내용 단정을 박아 넣지 않는 것이 의도다.** 자기-변이 전수 스윕(iter 9 갱신, `/tmp/t1603-d11/sweep.go`, 테스트와 동일 술어 — **13행 전부 DETECTED, 0 undetected**): A1 영어 렌더·A2 매트릭스 경로×문장삭제/빈필드·A3 en 정준 문장 삭제·A4 접두 존재·A5 재명명 접두+영어 본문 잔존·A6/A7 플래그명 제거·A8 마커 없음/별도 문단/재명명 t9999/**D12-b 맨 카드 번호(온전한 리터럴 없는 t1600)**.

- LEDGER-RED-AC005 — command: `go test ./internal/hook -run TestRedAutoModeGuidancePresent -v -count=1` · exit code: `1` · verbatim stdout (D11-b 강화판 — 플래그명 2종 + 블록 범위 마커 단정: `--auto-leader` 블록 안의 `t1600`):

  ```
  session_start_guide_i18n_red_test.go:111: leader notice must explain the 2-mode auto surface, missing "--auto-leader"
  session_start_guide_i18n_red_test.go:111: leader notice must explain the 2-mode auto surface, missing "--auto-lane"
  session_start_guide_i18n_red_test.go:125: leader notice has no --auto-leader guidance block
  --- FAIL: TestRedAutoModeGuidancePresent (0.00s)
  FAIL
  FAIL	github.com/modu-ai/moai-adk/internal/hook	1.269s
  FAIL
  ```

  적색 사유: 2모드 설명 문구가 테이블에 없음 — M1 신설이 뒤집는다(오늘은 블록 자체가 없어 세 번째 단정이 블록 부재로 적색; M1 이후에는 블록 내 리터럴 부재가 같은 단정을 적색으로 만든다). 블록 범위 + 온전한 리터럴 단정(`designed surface — t1600, not yet shipped`, D12-b)은 REQ-005의 진실성 보장 — 별도 문단 출처 표기(D11-b 반례)와 맨 카드 번호 변형(D12-b 반례), 플래그를 살아 있는 명령처럼 서술하는 변이를 모두 잡는다. 전수 자기-변이 스윕 결과(LEDGER-RED-AC002 셀과 같은 스윕, 13행 전부 DETECTED·0 undetected)가 이 셀의 변이 근거를 대신 인용한다.

- AC-001 (REQ-001) — 4-로케일 렌더: **Given** `.moai/config/sections/language.yaml` 의 conversation_language 가 `ko` 이고 **When** 부트스트랩 소스로 SessionStart 가 발화하면 **Then** 렌더된 안내의 산문 줄이 ko 테이블 항목과 일치한다(예: 리더 헤더가 `팩토리 모드: run ` 접두를 갖는다) — and 같은 시험을 `en` 설정으로 돌리면 영어 항목과 일치한다. 판정: `go test ./internal/hook -run TestRedAgentChannelFollowsConversationLanguage` (M2 후 녹색 전환). RED-now: LEDGER-RED-AC001 (4요소 충족 — command·stdout·exit code·tree SHA).
- AC-002 (REQ-002) — spawn authority 로케일 항목: **Given** 메시지 테이블의 spawn authority 필드가 **When** 4-로케일 각각에 대해 읽히면 **Then** 네 항목 모두 비어 있지 않고, ko/ja/zh 항목은 영어 접두 `Standing spawn authority:` 와 영어 본문 단편 `you are the lane session and therefore the orchestrator for your card` 를 모두 갖지 않으며(접두만 바꾼 재명명 변형도 잡는다 — D11-a), 네 항목 모두 매트릭스 포인터 경로 `.claude/rules/moai/development/spec-frontmatter-schema.md` 를 그대로 갖는다. 판정: `go test ./internal/hook -run TestRedLaneJoinAuthorityLocalized` (M1 후 녹색 전환) + 테이블 전수 테스트. RED-now: LEDGER-RED-AC002 (4요소 충족).
- AC-003 (REQ-003) — 문서 안내: **Given** 리더·레인 안내 각각이 **When** 임의 로케일로 렌더되면 **Then** 렌더 결과가 `https://adk.mo.ai.kr` 을 1회 이상 그대로 포함한다. 판정: 렌더 테스트의 URL 리터럴 포함 검사.
- AC-004 (REQ-004) — 2모드 설명: **Given** 안내가 `--auto-leader`·`--auto-lane` 을 설명하면 **When** 리더 변형과 레인 변형을 각각 읽으면 **Then** 리더 변형이 리더 모드(운영자 일괄 승인 수용·배차)를, 레인 변형이 레인 모드(`moai factory next` 임대 소비)를 중심으로 설명하고, 두 변형 모두 반대 모드의 존재를 한 문장으로 언급한다. 판정: 원문 리터럴 핀 — 각 변형이 모드 식별 문구를 포함(단락 내 연속 텍스트로 핀).
- AC-005 (REQ-005) — 설계 노면 명시: **Given** 출하 빌드에 `--auto-leader`/`--auto-lane` 플래그가 없으면(t1600 미병합) **When** 안내가 두 플래그를 이름으로 언급하면 **Then** 언급 블록(안내의 확립된 블록 구분자 `\n\n` 기준)이 로케일 무관 리터럴 `designed surface — t1600, not yet shipped` 을 같은 블록 안에 그대로 갖는다(REQ-001 프로토콜 토큰 규율 — 번역 대상 아님). 맨 카드 번호 `t1600`만 있는 변형이나 별도 문단의 출처 표기는 충분하지 않다(블록 범위 + 온전한 리터럴 검사 — D11-b·D12-b). 판정: `go test ./internal/hook -run TestRedAutoModeGuidancePresent` (M1 후 녹색 전환). RED-now: LEDGER-RED-AC005 (4요소 충족).
- AC-006 (REQ-006) — 폴백: **Given** conversation_language 가 `fr`(테이블 밖 광고값)·빈 문자열·nil 설정 각각일 때 **When** 안내를 렌더하면 **Then** 세 경우 모두 영어 테이블 항목과 동일한 완전한 안내다(빈 문자열 아님). 판정: `go test ./internal/hook -run TestFactoryMessagesFor` 확장. RED-now 없음 — regression-guard로 강등: 영어 폴백 기제(`factoryMessagesFor`, session_start_factory_i18n.go:304-312와 operatorLang, session_start_lang.go:23-30)가 현 트리에 이미 존재하고 동작하므로, 본 AC는 M1~M2 어느 변경으로도 뒤집을 수 없는 보존-행동 커버리지다(구현 착수 전에 적색이 재현 불가 — 셀 사유 기재, verification-completeness §2의 강등 규정 적용).
- AC-007 (REQ-008) — 미러 패리티: **Given** spawn authority·2모드·문서 안내의 서술 문단이 **When** 런타임 테이블과 배포 규칙 문단(factory-dispatch.md § Lane spawn authority (standing), gtd.md § --auto, factory-dispatch-detail.md:135의 verbatim 서술)에서 각각 읽히면 **Then** doc-parity 테스트가 live+미러 쌍에서 같은 고정 문구를 확인한다. 판정: `go test ./internal/cli -run 'TestAutoRank'` — todo_auto_doc_test.go 계열의 실측 가족 이름은 `TestAutoRank*`다(`go test ./internal/cli -list 'AutoRank'` 실측 16건 — TestAutoRankDoctrineAmendment·TestAutoRankMirrorParity·TestAutoRankMarkerDisclosure 등; 대조 실측 `-list 'AutoDoc'` → 0건. plan-audit iter 3 D6: 규약 기반 이름 'AutoDoc'은 실제 계열이 아니어서 빈 스위프를 냈다). 본 SPEC의 미러 문단 확장 테스트는 이 `TestAutoRank*` 계열에 추가되며 같은 셀렉터로 흡수된다.
- AC-008 (REQ-009) — 레이아웃 불변식: **Given** 새 메시지 필드 전체가 **When** 빌더가 블록을 조립하면 **Then** 어떤 필드도 선행·후행 개행을 갖지 않고 한 블록 안에 두 로케일 문장이 섞이지 않는다. 판정: 기존 레이아웃 불변식 테스트를 새 필드로 확장.
- AC-009 (REQ-007) — 원어 산문 품질 게이트: **Given** M1 이 작성한 ko·ja·zh 신규 산문(spawn authority 번역·2모드 설명·문서 안내 문장)이 **When** 항목별 원어 윤문 점검(moai-domain-humanize 자동 점검 또는 동등한 수동 검토 — 영어 직역 계어·AI 흔적 S1 유형)을 각 항목에 실행하면 **Then** 로케일별 판정 행이 `AC-009-verdict: <locale> grade <A|B> S1 <0-3>` 형식으로 run 증거(progress.md §E.2)에 로케일당 정확히 1행씩 기록된다. 판정 명령(3 grep — 행 앵커형, 각각 단일 호출, 셋 모두 관측값 1이어야 녹색): `grep -cE '^AC-009-verdict: ko grade (A|B) S1 [0-3]$' <progress.md>` = 1 · 같은 패턴의 `ja` 행 = 1 · 같은 패턴의 `zh` 행 = 1. (`<progress.md>` = `.moai/specs/SPEC-SESSION-START-GUIDE-I18N-001/progress.md`. 행 앵커 `^…$`가 산문 중간의 마커 언급 계수를 차단하고, 등급 술어 `(A|B) S1 [0-3]` + 앵커가 `grade F S1 9`·`grade A S1 30`류 미달 판정을 잡는다. 계보: iter 4 D7 — 단일 계수 `≥ 3`은 한 로케일 3반복과 grade F를 통과; iter 5 D8 — 비앵커 술어는 `S1 30`을 통과. 반례 완전 실측은 아래 RED-now 원장.) RED-now (4요소, 본 워크트리 tree SHA `81786284e6e5ee2fe5fbe7b498d649b27c3e0109`, 2026-10-08 — 교정 셀렉터를 기록 전에 세 위반 변형+실트리에 실행해 관측, 각 실행이 =1 요구를 올바르게 위반):
  - 변형 1 — ko 판정 행 3회 반복(`AC-009-verdict: ko grade A S1 0`·`ko grade B S1 1`·`ko grade A S1 2`): ko grep → stdout `3`, exit 0 — 계수 3 ≠ 1 요구 위반으로 판정 실패; ja·zh grep → `0`, exit 1.
  - 변형 2 — 전 로케일 `grade F S1 9`: ko·ja·zh grep 모두 → stdout `0`, exit 1.
  - 변형 3 — 전 로케일 `grade A S1 30`: ko·ja·zh grep 모두 → stdout `0`, exit 1 — 앵커가 비앵커 형태가 통과시켰던 `S1 30`을 거부.
  - 실제 progress.md(판정 행 부재): ko·ja·zh grep 모두 → stdout `0`, exit 1.
  green path: M1이 산문을 작성하고 §E.2에 로케일당 1행의 합격 판정 행을 남기면 세 grep이 각각 1이 된다. 적색 사유: 판정 행이 아직 존재하지 않아 0적중 — 이 AC가 검증하는 바로 그 산출물의 부재다.

## §D.1 심각도

- release-blocking: AC-001, AC-002, AC-005, AC-009 (카드 요구의 핵심, 텍스트 진실성, 안내 품질).
- regression-guard: AC-003, AC-004, AC-006, AC-007, AC-008 (고정·패리티·보존 행동·불변식 — 고장이 나도 즉시 사용자 불능은 아니다). AC-006의 강등 사유는 본인 셀에 기재했다.

## §D.2 간접 검증

- AC-004·AC-005의 모드 설명 정확성은 t1600 착지 전까지 설계 의미론 대비 텍스트 대조로만 검증 가능하다 — 실제 플래그 동작 대조는 t1600 착지 후 M3 재확인에서 이행한다.

## §D.3 종결 게이트

- 전 AC 녹색 + 레인-로컬 영향 계열 녹색 + 미러 문단 커밋 포함 + decision-index Q1 운영자 판정 기록.

## §D.4 품질 게이트 기준 (TRUST 5)

- Tested: 신설·확장 테스트가 전 AC를 운반하고 신규 Go 코드 커버리지 85% 이상.
- Readable: 메시지 필드 주석이 기존 factoryMessages 스타일(필드 당 역할·핀 근거 주석)을 따른다.
- Unified: gofmt 통과.
- Secured: 해당 없음(문자열 테이블 변경 — 신뢰 경계 없음).
- Trackable: 커밋에 카드 id t1603 + SPEC id 포함.

## §D.5 에지 케이스

- es/fr/de 등 광고-only 로케일 → 영어 폴백(AC-006).
- language.yaml 부재·파싱 불가 → fail-open 영어(operatorLang 계승).
- 증분 레인 join(`laneJoinNoCount` 형태)과 owned-card(Codex) 규칙에도 새 필드 결합이 깨지지 않는다 — 빌더 결합점 시험에 두 변형 포함.
- 한 로케일 항목만 빈 문자열로 두면 폴백이 아니라 빈 안내가 된다 — AC-002의 전수 비공백 검사가 잡는다.

## §D.6 정의 완료 (Definition of Done)

- REQ-001~009 전항목에 대응 AC 가 존재하고 판정 명령이 적혀 있다 — 매핑: REQ-001→AC-001, REQ-002→AC-002, REQ-003→AC-003, REQ-004→AC-004, REQ-005→AC-005, REQ-006→AC-006, REQ-007→AC-009, REQ-008→AC-007, REQ-009→AC-008.
- 런타임 방출기와 템플릿 미러 문단이 같은 커밋 집합 안에서 함께 바뀌었다(커밋 열거로 확인).
- decision-index.md 의 미해결 행(Q1)이 운영자 판정으로 닫혔다.

## §D.7 향후 확인

- t1600 착지 후: 안내 문구의 플래그 동작 서술과 실제 파싱·모드 동작의 대조 재확인(M3 재확인 항목), 설계-노면 마커 제거 여부 판정.
