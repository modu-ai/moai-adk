# plan — SPEC-SESSION-START-GUIDE-I18N-001

id: SPEC-SESSION-START-GUIDE-I18N-001
title: 팩토리 리더/레인 SessionStart 안내 문구 다국어 렌더
card: t1603
tier: M
created: 2026-10-08

## §A 맥락

운영자 지시(카드 t1603, 2026-10-08): 팩토리 리더/레인 SessionStart 안내 문구를 conversation_language별 다국어로 렌더, `/moai todo --auto-leader`·`--auto-lane` 2모드 설명 포함, https://adk.mo.ai.kr 온라인 문서 참고 안내 포함. 대상은 부트스트랩 주입 텍스트(standing spawn authority 포함)와 템플릿 미러 양측. 직렬 슬롯 임대는 t1588(레인-10 생 임대의 직렬 슬롯 정당 점유)로 거부됐고, 리더 배차 선례(t1498)로 본 레인이 카드를 완수한다.

## §B 조사 결과 (본 트리, base 81786284e, 2026-10-08 실측)

### B.1 런타임 방출기

- 메시지 테이블: `internal/hook/session_start_factory_i18n.go` — `factoryMessages` 구조체(:32-72, 20 필드), `factoryLocales` 4-로케일 테이블(:77-302, en/ko/ja/zh), `factoryMessagesFor` 영어 폴백(:304-312). 머리말 주석의 2 불변식: 산문만 테이블에 두고 프로토콜 토큰은 빌더에, 필드는 선행·후행 개행 없음.
- 빌더: `internal/hook/session_start_factory.go` — 리더/레인 안내 조립(:119, :173, :232), 레인 join + `laneSpawnAuthority` 결합(:243).
- spawn authority: `internal/hook/lane_spawn_authority.go:41` — 영어 단일 상수. :38-40 주석이 영어 고정의 근거(두-청중 규칙: additionalContext 는 agent-facing 이라 langEnglish 로 렌더)를 진술한다. 카드 지시가 이 규칙을 본 문구에 한해 뒤집는다.
- 호출점: `internal/hook/session_start.go:518` — additionalContext 에 `langEnglish` 하드코딩. :528 — systemMessage 에 `operatorLang(h.cfg)`. :503 주석은 agent-facing 사본이 `agent_prompt_language` 를 따른다고 서술하지만 코드는 langEnglish 상수다 — 서술과 코드가 이미 어긋난 지점이며 Q1 결정의 표적이다.
- 로케일 해석: `internal/hook/session_start_lang.go` — `operatorLang` 을 `c.Language.ConversationLanguage` 에서 읽고, nil/빈 값은 영어로 fail-open(:23-30).
- 설정: `.moai/config/sections/language.yaml` — conversation_language 옵션 ko/en/ja/zh/es/fr/de(현재 ko), `agent_prompt_language: en`(주석 "Always en"). 템플릿 원본은 `internal/template/templates/.moai/config/sections/language.yaml.tmpl`.

### B.2 템플릿 미러

- 안내 산문의 템플릿 리터럴 사본은 없다(훅은 Go 바이너리 내장, `internal/template/templates/.claude/hooks/moai/*.sh` 는 디스패처 래퍼).
- 미러 정합성의 실면: `internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md` § Lane spawn authority (standing) — 안내 문구와 같은 권한을 서술하는 배포 규칙; `factory-dispatch-detail.md:135` — "the SessionStart join notice carries the authority sentence verbatim" 서술(문구가 바뀌면 이 서술도 정확해야); `internal/template/templates/.claude/skills/moai/workflows/gtd.md` § --auto.
- 패리티 테스트 선례: `internal/cli/todo_auto_doc_test.go` — live+미러 쌍(`.claude/skills/moai/workflows/gtd.md` ↔ `internal/template/templates/.../gtd.md` 등)을 원문 리터럴(raw text)로 고정하고 비고정 문구는 공백 정규화 비교. 본 SPEC 의 REQ-008 이 이 선례를 확장한다.

### B.3 --auto 2모드 현황

- `--auto-leader`·`--auto-lane` 플래그는 본 base에 존재하지 않는다(전트리 grep 0적중 — 플래그 등록 `internal/cli/todo.go:323` 을 통합 `--auto` 하나만 갖는다).
- 현 면: 통합 `--auto` = 운영자 일괄 승인 직렬 소비 사이클(`internal/cli/todo_auto.go`, SPEC-MANAGER-TODO-001), 레인 자기-디스패치 면 = `moai factory next` 임대 소비(`internal/cli/todo_auto_lane.go`, t1554).
- 2모드 분리는 형제 카드 t1600(워크트리 `.moai/worktrees/auto-mode-split`, 미병합)이 설계 중. 본 SPEC 은 안내 텍스트를 설계 의미론(리더 모드 = 일괄 승인 수용·배차 / 레인 모드 = 임대 소비)에 대해 정의하고 REQ-005의 설계-노면 명시로 진실성을 유지한다.

## §C 사전 비행 확인

- [x] SPEC ID 정규식 Bash 실행 — `PASS` (SPEC-SESSION-START-GUIDE-I18N-001)
- [x] 프런트매터 12 정식 필드 검증(스키마 SSOT 대조)
- [x] ID 중복 없음(`.moai/specs/` 에 SESSION-START-GUIDE-I18N 0건)
- [x] GEARS 표기 요구사항(spec.md §2)
- [x] Out of Scope H3 + 불릿(3개 주제)
- [x] decision_gate: on(interview.yaml:6) → decision-index.md 동반 작성

## §D 제약

- Go 변경은 문자열 테이블 + 호출점 1곳 + 테스트 확장에 한정한다. 새 메시지 필드 추가 시 `factoryMessages` 구조체와 4-로케일 항목을 함께 채운다(한 필드라도 빈 로케일 항목은 REQ-006 위반이다).
- 문서-패리티 테스트는 원문 리터럴 핀 원칙(todo_auto_doc_test.go 머리말)을 따른다 — 줄바꿈에 걸친 문구는 리터럴 grep 이 깨지므로 고정 문구는 단락 내 연속 텍스트로 쓴다.
- [RESOLVED 2026-10-08: agent-facing additionalContext 채널의 로케일 키 → conversation_language. 카드 본문(운영자 지시)의 우위로 two-audience 규칙(lane_spawn_authority.go:38-40)과 `agent_prompt_language: en` 정책을 본 안내 면에 한해 개정한다; `agent_prompt_language`는 그 외 agent-facing 면에 계속 지배. 판정: decision-index.md Q1, 레인 결정 기록 `.moai/reports/t1603/`.]
- RED 테스트 파일 커밋 시점 고정(plan-audit iter 7 D10): `internal/hook/session_start_guide_i18n_red_test.go` 는 **M2 전환 커밋과 함께** 커밋한다 — TestRed 3종이 모두 녹색이 되는 첫 커밋이다. plan-phase 커밋에도 M1 커밋에도 넣지 않는다. 사유: pre-commit 게이트가 변경 패키지 테스트 통과를 요구하는데, AC-001·AC-005는 M2까지 적색이므로 그 이전 커밋은 게이트를 위반하고(또는 CI 적색) 적색 테스트를 트리에 남긴다. M1 작업 동안 이 파일은 워크트리에 미커밋 상태로 존재한다.

## §E 자기 검증

- 본 plan 의 모든 file:line 인용은 본 워크트리(base 81786284e)에서 읽은 값이다.
- 플래그 부재 확인은 grep 실측이다(§B.3). t1600 트리는 읽지 않았다 — 설계 의미론은 카드 지시문이 준 두 줄 요약(리더 = 일괄 승인 수용·배차 / 레인 = 자기-디스패치 임대 소비)에 근거하며, t1600 착지 시 안내 문구와 실제 플래그 동작의 대조를 M3 패리티 단계에서 재확인한다.

## §F 마일스톤 (결정 가역성 순 — 바뀔 가능성이 큰 결정부터)

- M1 (Priority High) — 내용 결정: 로케일 키 정책 확정(Q1 RESOLVED — additionalContext 채널도 conversation_language, two-audience 규칙은 본 면 개정) + `laneSpawnAuthority`를 메시지 테이블 필드(`laneSpawnAuthority`, 4-로케일)로 이항 + 2모드 설명 필드(`autoModeGuide`, 리더/레인 변형) + 문서 안내 필드(`docsPointer`) 신설, en/ko/ja/zh 전 항목 작성. REQ-001~004, REQ-007. ko/ja/zh 산문은 자연 원어로 작성하고 프로토콜 토큰(URL, 매트릭스 경로, 플래그명, cron 식)은 그대로 둔다. **설계-노면 마커 리터럴 고정(D12-b): 2모드 설명 블록은 로케일 무관 리터럴 `designed surface — t1600, not yet shipped` 을 블록 안에 그대로 포함해야 한다(REQ-001 프로토콜 토큰 규율 — 번역하지 않는다). 온전한 리터럴 없이 맨 카드 번호 `t1600`만 두는 변형은 AC-005 시험이 거부한다.** **[상태: 완료 2026-10-08 — 필드 4종 신설·4-로케일 전 항목 착지·AC-009 윤문 게이트 grade A; 커밋은 D10 고정에 따라 레인이 M1 경계에서 스테이지 (progress.md §E.2)]**
- M2 (Priority High) — 배선: `session_start.go:518` 호출점의 로케일 인자를 conversation_language로 교체(Q1 RESOLVED — additionalContext 채널), spawn authority 결합점(`session_start_factory.go:243`)을 테이블 참조로 전환. REQ-001, REQ-002, REQ-006. **커밋 고정(§D 제약): RED 테스트 파일 `internal/hook/session_start_guide_i18n_red_test.go` 는 본 M2 전환 커밋에 포함한다** — TestRed 3종이 모두 녹색이 되는 첫 커밋이므로 pre-commit 게이트(변경 패키지 테스트 통과)를 위반하지 않는다. plan-phase·M1 커밋에는 포함하지 않는다. **[상태: 배선 완료 2026-10-08 — session_start.go :518 conversation_language 교체·session_start_factory.go :243 테이블 참조 전환 착지, TestRed 3종 녹색 확인 (progress.md §E.2); 커밋(RED 테스트 파일 포함)은 레인의 M2 전환 커밋 소관]**
- M3 (Priority Medium) — 미러·패리티: 배포 규칙 문단 갱신(factory-dispatch.md § Lane spawn authority (standing), factory-dispatch-detail.md:135 서술, gtd.md § --auto의 필요 문단) + `todo_auto_doc_test.go` 선례의 doc-parity 테스트를 본 SPEC 문단 쌍으로 확장. REQ-008. t1600 착지 시점에 따라 2모드 문구의 설계-노면 표기를 재확인한다.
- M4 (Priority Medium) — 검증 고정: 렌더 테스트(4-로케일 + 폴백 + 빈 설정), 레이아웃 불변식 테스트 확장(새 필드의 개행·혼합 언어), 원문 리터럴 핀(플래그명+설계 노면 마커, URL). REQ-005, REQ-006, REQ-009. 레인-로컬 영향 계열만 실행 — 패키지 분할 형태로, 본 카드의 RED 테스트(TestRed*)를 반드시 포함한다: `go test ./internal/hook -run 'TestRed|Factory'` 와 `go test ./internal/cli -run 'TestRed|AutoRank'`. (plan-audit iter 2 D5 수리: 기존 결합 2-패키지 필터 `-run 'Factory|AutoDoc'`는 TestRed*를 0건 선택 — 감사자 `-list` 실측 47선택/0TestRed — 하며 결합 실행은 로컬 120초 초과. TestRed를 셀렉터에 포함하고 패키지별로 나눠 실행한다. plan-audit iter 3 D6 수리: cli 측 `AutoDoc`는 실제 테스트 계열 이름이 아니어서 0건을 골랐다 — doc-parity 계열의 실측 이름은 `TestAutoRank*`(internal/cli/todo_auto_doc_test.go, `-list 'AutoRank'` 실측 16건: TestAutoRankDoctrineAmendment·TestAutoRankMirrorParity·TestAutoRankMarkerDisclosure 등)다. 규약에 의존해 아직 그 규약을 채택한 테스트가 없는 이름을 검증 명령에 쓰면 D5가 닫은 것과 같은 빈-스위프 위험이다 — 셀렉터는 `-list`로 존재를 실측한 이름만 쓴다. cli 측 TestRed는 현 트리 0건(cli 패키지에는 RED 테스트가 아직 없음)이며, AC-007 확장 테스트가 M3에서 착지할 때 같은 셀렉터로 흡수된다.) **[상태: 완료 2026-10-08 — M4 셀렉터 갱신(턴 종료 리뷰 P1 수리 동반): hook 측 명령을 `go test ./internal/hook -run 'TestRed|Factory|LeaderNotice'` 로 확장한다 — 추가 가족: TestAutoRankSessionStartGuideAuthorityParity(cli, 'TestRed|AutoRank' 셀렉터 흡수), TestFactoryGuideNewFieldsCompleteInEveryLocale·TestFactoryGuideAutoModeMarkerBlocksEveryLocale·TestFactoryGuideDocsPointerInBothNotices·TestFactoryGuideFallbackCompleteNotice·TestFactoryGuideNewFieldsLayoutInvariants(hook, 'Factory|Guide' 매치), 그리고 리더 additionalContext 로케일 정책을 고정하던 TestLeaderNoticeBatchGatePointer 가족(t1480 계열 수리 — session_start.go :518→conversation_language 배선이 뒤집은 영어-고정 기대를 로케일-적합성 단정으로 갱신, 'LeaderNotice' 셀렉터 추가). 레인-로컬 실행은 갱신된 두 명령으로 수행했다 (progress.md §E.2)]**

## §G 안티 패턴

- 플래그가 없는 빌드에서 살아 있는 명령처럼 `--auto-leader`를 서술한다 (REQ-005 위반).
- 테이블에 en 항목만 채우고 ko/ja/zh 는 영어 사본으로 둔다 (REQ-001 위반 — 폴백이 가리지 못한다).
- 산문만 고치고 `factory-dispatch-detail.md:135` 의 "verbatim" 서술을 정확하지 않게 남겨둔다 (REQ-008 위반 — 미러 갈라짐).
- 문단을 줄바꿈에 걸쳐 쓰고 리터럴 핀을 채택한다 (todo_auto_doc_test.go 머리말이 경고한 결함형).
- ko 번역을 영어 어순 직역으로 쓴다 (REQ-007 위반 — "선행 개행"류 캘크).

## §H 교차 참조

- 선례 SPEC: SPEC-TODO-AUTO-PRIORITY-001 (doc-parity 테스트 계열의 원천), SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-019 (레인 규칙의 로케일 규율 원천), SPEC-LAUNCHER-ENTRY-FLAGS-001 (두-청중 채널 규율의 원천).
- 외부 의존: 카드 t1600 (`--auto` 2모드 분리, 미병합) — M1 문구의 설계 노면 표기와 M3 재확인의 대상. SPEC ID 가 아니므로 `depends_on` 이 아니라 본 절과 REQ-005로 운반한다.
- 규칙: `.claude/rules/moai/core/native-idiom-and-register.md`, `.claude/rules/moai/workflow/factory-dispatch.md`.
