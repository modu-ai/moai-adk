---
id: SPEC-SESSION-START-GUIDE-I18N-001
title: "팩토리 리더/레인 SessionStart 안내 문구 다국어 렌더 — 2모드 --auto 설명·온라인 문서 안내·템플릿 미러 패리티"
version: "0.1.0"
status: completed
created: 2026-10-08
updated: 2026-10-09
author: manager-spec (card t1603)
priority: P2
phase: "v3.2.0 target"
module: "internal/hook"
lifecycle: spec-anchored
tags: "i18n, session-start, factory-bootstrap, spawn-authority, template-parity, auto-mode"
tier: M
---

# SPEC — 팩토리 리더/레인 SessionStart 안내 문구 다국어 렌더

## HISTORY

- 2026-10-08: 카드 t1603(운영자 지시 2026-10-08, Class C)에서 plan-phase 작성. 카드 내용: 리더/레인 SessionStart 안내 문구를 conversation_language별 다국어로 렌더하고, `/moai todo --auto-leader`·`--auto-lane` 2모드 설명 안내를 포함하며, 상세는 https://adk.mo.ai.kr 온라인 문서 참고 안내를 덧붙인다. 대상은 부트스트랩 주입 텍스트(standing spawn authority 문구 포함)와 템플릿 미러 양측. 직렬 슬롯 임대 거부 환경에서 리더 배차 선례(t1498)로 레인이 진행.

## §1 배경과 문제 정의

팩토리 리더/레인 세션의 SessionStart 부트스트랩 안내 문구는 이미 4-로케일 메시지 테이블(`internal/hook/session_start_factory_i18n.go:77-302`, en/ko/ja/zh)을 갖추고 있고, 운영자 면 systemMessage 채널은 `operatorLang`(internal/hook/session_start_lang.go:23-30)으로 conversation_language를 따른다. 그러나 세 지점이 카드 요구와 어긋난다.

1. **agent-facing additionalContext 채널이 영어 고정이다.** `internal/hook/session_start.go:518` 이 `factoryBootstrapNoticeForSource(..., langEnglish)` 로 하드코딩한다. 레인이 실제로 읽는 채널이 바로 이 채널이고, standing spawn authority 문구도 여기에 실린다(`internal/hook/session_start_factory.go:243` 의 join + `laneSpawnAuthority` 결합). `laneSpawnAuthority` 상수는 두-청중 규칙에 근거해 영어 단일 문자열로 고정돼 있다(`internal/hook/lane_spawn_authority.go:38-41`). 결과적으로 한국어 운영자가 띄운 레인 세션도 부트스트랩 안내를 영어로 읽는다.
2. **`--auto-leader`·`--auto-lane` 2모드 설명이 없다.** 현 트리에는 통합 `--auto`(internal/cli/todo.go:323)와 레인 자기-디스패치 면(internal/cli/todo_auto_lane.go, t1554)만 있다. 2모드 분리는 형제 카드 t1600이 설계 중이며 본 base에 미병합이다. 안내 문구는 착지 전에는 설계 노면임을 명시해야 텍스트가 출하물에 대해 진실하게 유지된다.
3. **온라인 문서 안내가 없다.** 운영자 지시는 상세 문서 위치로 https://adk.mo.ai.kr 을 안내하도록 요구한다.

### 설계 결정 — 본 카드에 의한 규칙 개정 (Q1, 2026-10-08 판정)

본 카드는 두 기존 정책을 **SessionStart 부트스트랩 안내 면에 한해** 개정한다: (a) `lane_spawn_authority.go:38-40` 의 two-audience 규칙(additionalContext 를 langEnglish 로 렌더), (b) `.moai/config/sections/language.yaml` 의 `agent_prompt_language: en` 정책(주석 "Always en")의 이 면 적용. 개정 근거는 운영자 지시(카드 본문)의 우위다 — 카드 본문이 리더·레인 안내 전체를 conversation_language별 렌더로 명시하고 standing spawn authority 문구를 다국어 대상에 이름으로 포함한다. `agent_prompt_language` 는 그 외 모든 agent-facing 면(서브에이전트 프롬프트 등)에 계속 지배한다. 레인 판정 기록: decision-index.md Q1, `.moai/reports/t1603/`.

템플릿 미러 측: 안내 산문 자체는 Go 런처 바이너리에 컴파일되어 템플릿 트리에 리터럴 사본이 없다. 미러 정합성의 실제 면은 (a) 같은 의미를 서술하는 배포 규칙 문서들(`internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md` § Lane spawn authority (standing), `factory-dispatch-detail.md:135` "the SessionStart join notice carries the authority sentence verbatim", `gtd.md` § --auto)과 (b) live·미러 쌍을 원문 리터럴로 고정하는 문서-패리티 테스트 계열(`internal/cli/todo_auto_doc_test.go` 선례 — AC-TAP-011/-013/-014)이다. 안내 문구가 바뀌면 이 양측이 같은 변경으로 움직여야 갈라지지 않는다.

## §2 GEARS 요구사항

- REQ-001 (Ubiquitous): The SessionStart 부트스트랩 안내(리더 변형·레인 변형 모두, additionalContext 채널과 systemMessage 채널 모두) shall 자신의 산문을 세션의 conversation_language 로케일 메시지로 렌더한다 — agent-facing additionalContext 채널을 포함한다(§1 설계 결정의 카드 개정). 명령·run id·소켓 경로·레인 라벨·MCP 도구 이름·cron 식·규칙 파일 경로는 프로토콜 토큰으로 모든 로케일에서 그대로 둔다(기존 factoryMessages 규율, session_start_factory_i18n.go 머리말 주석의 2 불변식을 계승).

- REQ-002 (Ubiquitous): The standing spawn authority 문구 shall 메시지 테이블의 로케일 키 항목이 되어 지원 로케일마다 하나의 번역을 갖는다. 번역은 기존 상수의 3 설계 결정을 의미 수준에서 보존한다 — (1) 전문가 매핑의 위임(포인터가 매핑의 집이다), (2) depth-1 리프 봉인, (3) 부트스트랩 상시 권한(동료 메시지가 주거나 빼앗을 수 없음). 매트릭스 포인터 경로 `.claude/rules/moai/development/spec-frontmatter-schema.md` 는 프로토콜 토큰으로 모든 로케일에서 그대로다.

- REQ-003 (Ubiquitous): The 리더·레인 안내는 each 상세 참조 문장을 shall 갖는다 — https://adk.mo.ai.kr 온라인 문서에서 전체 내용을 볼 수 있다는 안내. URL 자체는 모든 로케일에서 그대로며, 주변 산문만 로케일화한다.

- REQ-004 (Where): **Where** 안내 문구가 `/moai todo --auto-leader`·`--auto-lane` 2모드 분리를 설명하면, the 문구 shall 두 모드를 구분해 설명한다 — 리더 모드 = 운영자 일괄 승인 하의 큐 수용·레인 배차, 레인 모드 = `moai factory next` 임대 소비의 자기-디스패치 — and 각 변형(리더 안내에는 리더 모드, 레인 안내에는 레인 모드 중심)에 맞는 쪽을 앞세운다.

- REQ-005 (When): **When** 안내 문구가 출하 빌드가 파싱하지 못하는 플래그나 명령을 이름으로 언급하려 하면, the 문구 shall 그 표면을 설계/착지 예정 노면으로 명시한다(형제 카드 t1600 귀속 표기 포함) — 살아 있는 명령처럼 서술하지 않는다. 텍스트는 출하물에 대해 진실해야 한다.

- REQ-006 (When): **When** conversation_language 가 메시지 테이블에 없는 값(예: language.yaml 이 광고하는 es/fr/de)이거나 빈 값·읽을 수 없는 설정이면, the resolver shall 영어 폴백으로 완전한 안내를 렌더한다 — 빈 안내나 한 블록 안의 언어 혼합은 없다. 폴백 계약은 기존 factoryMessagesFor(session_start_factory_i18n.go:304-312)와 operatorLang(session_start_lang.go:23-30)의 fail-open 규율을 계승한다.

- REQ-007 (Ubiquitous): The ko·ja·zh 산문 shall 영어의 직역 계어(캘크)가 아니라 자연 원어 문장이다. 레지스터는 깨끗한 기술 문서체다(채팅 구어체 아님). 근거 규칙: `.claude/rules/moai/core/native-idiom-and-register.md`.

- REQ-008 (When): **When** 안내 산문·spawn authority 문구·2모드 설명·문서 안내 문장이 바뀌면, the 변경 shall 런타임 방출기(메시지 테이블)와 템플릿 미러 doctr 面(같은 의미를 서술하는 배포 규칙 문단 — factory-dispatch.md § Lane spawn authority (standing), gtd.md § --auto, 필요 시 auto-semantics.md 인용 문단) 양쪽에 같은 변경으로 착지한다, and the 문서-패리티 테스트 계열은 todo_auto_doc_test.go 선례를 따라 수정된 문단을 live+미러 쌍 원문 리터럴로 고정한다.

- REQ-009 (Unwanted): The 안내 문구 shall not 하나의 메시지 블록 안에 두 로케일의 문장을 섞지 않는다. 블록 내 줄 결합·블록 간 공백 분리의 기존 레이아웃 불변식(선행·후행 개행 없음)은 새 필드에도 유지된다.

## §3 제약

- 하네스: standard. 본 SPEC 은 Go 문자열 테이블 + 훅 호출점 1곳 + 문서-패리티 테스트 확장이므로 Tier M.
- 아티팩트 언어: 한국어(카드 운영자 지시 언어, conversation_language: ko 일치).
- 지원 로케일 집합은 기존 4-로케일(en/ko/ja/zh) + 영어 폴백을 유지한다. es/fr/de 는 language.yaml 이 conversation_language 로 광고하지만 안내 테이블에는 없으며, 이 SPEC 은 테이블을 확장하지 않는다(폴백이 처리).
- `--auto-leader`/`--auto-lane` 플래그 구현은 본 SPEC 범위 밖이다(형제 카드 t1600 소관). 본 SPEC 은 안내 텍스트만 운반하며 착지 순서 의존을 명시한다.
- URL 화이트리스트: 안내에 들어가는 문서 도메인은 https://adk.mo.ai.kr 하나다.
- 레인-로컬 검증만 수행한다(`go test ./internal/hook ./internal/cli` 영향 계열). 전체 스위트는 금지(레인 프로토콜 §8).
- 새 내보내기 함수에는 @MX:NOTE 이상, 런타임-템플릿 문단 쌍에는 원문 리터럴 핀을 둔다.

## §4 수용 기준 요약

상세 Given-When-Then 은 acceptance.md (AC-001 ~ AC-008). 요지: (1) 4-로케일 렌더 + 영어 폴백, (2) spawn authority 로케일 항목과 3 결정 보존, (3) 2모드 설명 + 설계 노면 명시, (4) adk.mo.ai.kr 안내, (5) live·템플릿 미러 문단 패리티 테스트, (6) 블록 레이아웃 불변식 유지.

## §5 Out of Scope

### Out of Scope — `--auto-leader`/`--auto-lane` 플래그 자체

- 플래그 파싱·2모드 엔진 분리·큐 수용 반변경은 t1600 소관이다. 본 SPEC 은 안내 문구와 그 진실성 표기만 운반하며, 플래그가 없는 빌드에서 안내가 깨지지 않게 하는 것(설계 노면 명시)까지가 책임의 끝이다.

### Out of Scope — 온라인 문서(adk.mo.ai.kr) 콘텐츠

- docs-site 페이지의 작성·갱신·배포는 본 SPEC 밖이다. 안내 문구는 기존 페이지를 가리킬 뿐, 가리키는 대상의 내용 정합성은 oss-docs 하네스 소관이다.

### Out of Scope — 안내 문구 밖의 SessionStart 표면

- stale-run 통지, handoff 주입, chain banner, instructions-loaded 등 SessionStart 의 다른 표면은 대상이 아니다. 본 SPEC 의 대상은 팩토리 부트스트랩 안내(factoryMessages 전 필드)와 laneSpawnAuthority 문장뿐이다.
