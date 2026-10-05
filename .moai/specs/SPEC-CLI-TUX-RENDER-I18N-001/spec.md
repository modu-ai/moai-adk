---
id: SPEC-CLI-TUX-RENDER-I18N-001
title: "init/update TUX render residual repair and i18n unification — card t756 follow-up over SPEC-INIT-TUX-I18N-001"
version: "0.1.1"
status: completed
created: 2026-09-14
updated: 2026-09-14
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/cli, internal/cli/wizard"
lifecycle: spec-anchored
tags: "cli, tux, wizard, huh-v2, render, i18n, layout, pty"
tier: M
related_specs: [SPEC-INIT-TUX-I18N-001, SPEC-CLI-WIZARD-RESTRUCTURE-001, SPEC-CLI-TUX-INIT-UPDATE-001, SPEC-I18N-GOVERNANCE-001]
---

# SPEC-CLI-TUX-RENDER-I18N-001 — init/update TUX 렌더 잔여 수리와 i18n 통일

> 카드: **t756** (Factory, Tier M, Class C). 발행 2026-09-09 · 운영자 승인. 재현 근거: 운영자 스크린샷 2건, `.moai/reports/init-tui-audit-20260909.html` §1 (primary 체크아웃, 읽기전용 인용).

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-14 | manager-spec | 최초 작성(카드 t756 plan 단계). 카드의 F10/F11 을 현재 트리(a404132e7)에 재고정한 결과, 선행 SPEC-INIT-TUX-I18N-001(t586, completed)이 이미 흡수·정렬·간격·폭·현지화 대부분을 전달했음을 확인하고, 본 SPEC 을 **잔여 결함 수리 + 재발 방지 가드 + 캡처 기반 재검증**으로 범위 확정했다. |
| 0.1.1 | 2026-09-14 | manager-spec | plan-audit 1차(FAIL 0.925) 결함 D1~D3 반영. REQ-TRI-003 의 기본 수리 경로를 로컬 확인 필드 래퍼로 확정하고 라이브러리 인라인 모드를 구조적 부적합(huh `field_confirm.go:254`·`:293`)으로 M1 측정 참고로 격하했다. huh 고정 빈 줄 인용을 `:261-263` 으로 재고정했고, 실존 검사 기호(`TestLayout_NoBlankBetweenFields`, `TestEmitAcceptEditsConfirmationAnchor`)로 판정 명령을 바로잡았으며, 앵커 토큰 계약의 소유 SPEC(SPEC-V3R6-CLI-CONFIG-INTEGRITY-001)을 명명했다. |

## §A 배경

### A.1 카드 판정문과 현재 트리의 관계

카드 t756의 F10(영어 고정 표면)/F11(렌더 결함)은 2026-09-09 감사 보고서의 판정문을 따른다. 그러나 같은 감사에서 파생된 선행 SPEC-INIT-TUX-I18N-001(카드 t586)이 2026-09-13 completed 로 닫혔고 그 결과 코드가 본 카드의 베이스 커밋(a404132e7)에 포함돼 있다. plan 단계 재고정(re-anchor) 결과:

| 카드 판정 | 현재 트리 상태 | 본 SPEC 처리 |
|---|---|---|
| F11-(1) 예/아니오 버튼 중앙정렬 | 수리됨 — 확인 필드 2곳 모두 좌측 정렬 지정 + 버튼 좌패딩 제거(t586 REQ-ITI-014/AC-ITI-015) | 재발 방지 가드(REQ-TRI-002) + 캡처 재검증 |
| F11-(2) 질문 사이 빈 공간 3겹 | 부분 수리 — 필드 사이 빈 줄 0/빈 카드 줄 0(t586 REQ-ITI-015/AC-ITI-016). **잔여**: huh 라이브러리 확인 필드가 제목과 버튼 사이에 고정 빈 줄 2개를 내부에서 출력(라이브러리 소스 `field_confirm.go:261-263`, 테마로 제거 불가) | 잔여 간격 수리(REQ-TRI-003/004) |
| F11-(3) 선택 항목 폭 | 수리됨 — 표시 폭 기준 설명 열 정렬 + 폭 상한(t586 REQ-ITI-016/AC-ITI-017, 캡처·뮤턴트 대조군 포함) | 캡처 재검증 + 회귀 금지(REQ-TRI-005) |
| F10-(1) init 프로필 확인 프롬프트·구형 위저드 영어 고정 | 수리됨 — init 은 프로필 확인 질문을 아예 묻지 않고(t586 REQ-ITI-001), 구형 v1 프로필 위저드는 v2 로 흡수·현지화됨 | i18n 잔여 훑기(REQ-TRI-006)로 미번역 표면이 더 있는지 전수 확인 |
| F10-(2) 버튼 Yes/No 고정 | 수리됨 — 확인 버튼 라벨이 로케일 문자열 표에서 해석됨, 키맵도 로케일별 제공 | 캡처 재검증(ko 프레임) |
| F10-(3) v1 위저드 처분 | 결정 완료 — v2 흡수. 트리에 huh v1 잔존 import 없음 | 비-회귀 가드(REQ-TRI-007) |

### A.2 결정 항목 D1 — v1 위저드 처분 (Kickoff 게이트 확인용)

**권고: t586 이 내린 "v2 흡수"를 유지한다.** 근거: v1 구현과 v1 전용 테스트가 트리에서 제거돼 있고(모듈 캐시 의존 v1 흔적 없음, `go.mod`에 huh v1 부재), 흡수 결과물에 대한 캡처 기반 검증(AC-ITI-015~017)이 이미 통과 상태다. "최소화" 대안은 이미 삭제된 코드를 되살리는 비용만 남기므로 기각한다. 본 항목은 run 진입 시 Kickoff 게이트에서 운영자에게 최종 확인을 받는다(본 SPEC이 새 결정을 만들지 않는다).

### A.3 판정 방식 — 실제 TTY 렌더 캡처 [HARD]

카드의 경구대로, 모든 렌더/i18n 수리 판정은 실제 TTY 렌더 캡처로 한다. 마크업·내부 구조 단위의 단언만으로는 공허한 초록이 된다. 재사용 기반: `internal/cli/ptycaptest` 하니스와 `internal/cli/wizard/ptycap_test.go`·`layout_alignment_test.go` 패턴(tmux 세션 + 80열 프레임 캡처 + ANSI 제거 + 표시 폭 계산 + golden 내보내기). 마크업 단위 검사는 프레임 도달성의 **전제 단언**(기준 문자열 존재 확인)으로만 쓰고, 합격 판정은 캡처 프레임에서 한다.

## §B 범위

- **표면**: `moai init` / `moai update` / `moai profile setup` 이 띄우는 모든 대화형 확인(confirm)·선택(select) 화면 — 위저드 본체, 다운그레이드 확인창, 프로필 위저드 그룹, 그 밖에 M1 조사에서 밝혀지는 대화형 표면 전부.
- **로케일**: 기존 번역 표가 지원하는 로케일(en/ko/ja/zh) 범위 안에서 검증한다. 캡처 기본 로케일은 en·ko 2종이다.
- **지오메트리**: 80열 표준 폭(t586 캡처 관례와 동일).

## §C 요구사항 (GEARS)

- **REQ-TRI-001** (baseline capture gate) **When** run phase 가 시작되면, the run-phase executor shall 수리 편집에 앞서 §B 의 모든 대화형 표면을 고정 로케일(en·ko)·고정 지오메트리(80열)로 PTY 캡처해 기준 프레임을 만들고, 각 프레임을 카드 6개 판정(F10-(1)(2)(3), F11-(1)(2)(3))에 대응시켜 잔여 결함 표를 산출한다. 캡처 프레임과 잔여 표는 SPEC 디렉터리 아래 증거로 남긴다.
- **REQ-TRI-002** (alignment sweep guard) The wizard package shall 모든 `huh.NewConfirm` 생성 지점이 버튼 좌측 정렬을 명시하도록 유지하며, 정렬 지정이 없는 새 확인 필드 생성 지점이 생기면 실패하는 소스 스윕 가드 검사를 둔다. (huh 기본값은 중앙 정렬이므로 무방비 상태에서는 결함이 재발한다.)
- **REQ-TRI-003** (confirm internal gap) The wizard confirm surfaces shall 확인 필드의 제목(묻는 말)과 버튼 줄 사이의 빈 행이 캡처 프레임에서 1개 이하가 되게 한다. 라이브러리가 확인 필드 내부에 고정 빈 줄 2개를 출력하는 것은 테마로 제거할 수 없으므로, **기본 경로는 로컬 확인 필드 래퍼**(huh.Field 를 구현한 우리 쪽 타입이 자체 View 조합으로 제목·설명과 버튼 줄 사이 간격을 제어)로 달성한다. 라이브러리의 인라인 모드는 구조적으로 부적합하다 — 제목과 설명 사이 개행을 함께 제거하고 버튼 줄을 앞 줄과 합쳐 한 줄로 만들기 때문에 간격 축소 수단으로 쓸 수 없으며, M1 측정 참고용으로만 관찰한다.
- **REQ-TRI-004** (gap non-regression) While consecutive fields render, the wizard shall 필드 사이 빈 줄 0·선택 필드 아래 빈 카드 줄 0 상태를 유지한다(선행 SPEC의 기준을 감쇠시키지 않는다).
- **REQ-TRI-005** (selection width) The select option rows shall 같은 선택 목록 안에서 모든 옵션 줄의 설명 시작 표시 열(동아시아 전각 2칸 기준)이 같도록 유지한다.
- **REQ-TRI-006** (i18n sweep) The init/update/profile interactive surfaces shall 모든 사용자 대상 문자열을 로케일 계층으로 해석하며, ko 로케일 캡처 프레임에는 번역 표에 ko 항목이 존재하는 문자열의 영어 원문이 나오지 않게 한다. M1 조사에서 영어 고정 잔여가 발견되면 번역 표에 넣어 해석한다. 표준 출력 고지문 중 grep 고정 앵커 토큰 계약(예: acceptEdits 정규화 고지의 `acceptEdits`·`settings.local.json` 토큰)이 있는 문자열은 현지화하더라도 앵커 토큰을 번역문 안에 보존해 기존 앵커 검사를 깨지 않게 한다.
- **REQ-TRI-007** (v1 non-regression) The init/update/profile surfaces shall huh v1 을 다시 들이지 않으며, 패키지 의존에 huh v1 이 나타나면 실패하는 가드 검사를 둔다.
- **REQ-TRI-008** (verification-only closure) When 기준 캡처가 어떤 카드 판정 항목의 적합 상태를 보여주면, the run-phase executor shall 그 항목을 수리 편집 없이 "검증만으로 종결"로 기록하고 캡처 프레임을 증거로 남긴다. 적합한 코드를 고쳐서 결함을 만들지 않는다.

## §D 제약조건

- huh 라이브러리 자체는 수정하지 않는다(의존 고정 `charm.land/huh/v2 v2.0.3`, upstream 패치 제외). 우리 쪽 설정·테마·래퍼로만 해결한다.
- 기존 캡처 하니스(`internal/cli/ptycaptest`)와 기존 골든 관례를 재사용한다. 새 캡처 프레임워크를 만들지 않는다.
- 캡처 검사는 tmux 게이트(`MOAI_PTY_CAPTURE=1`) 아래에서만 실행되고, 게이트 없이는 건너뜀이 관측 가능해야 한다(선행 SPEC의 AC-ITI-019 계약 승계).
- 질문 수·페이지 구조 개편, 번역 신규 로케일 추가는 본 SPEC 밖이다(§F).

## §E 성공 기준 (요약)

모든 REQ 에 대응하는 기계 판정 가능한 AC 가 acceptance.md 에 있고, 각 렌더/i18n AC 의 합격 증거가 캡처 프레임이다. 잔여 표의 모든 행이 "수리 후 재검증 PASS" 또는 "검증만으로 종결" 둘 중 하나로 닫힌다.

## §F Out of Scope

### Out of Scope — 질문 수·페이지 구조 개편

- 감사 보고서 §2 이하의 "질문 빼기"/moai web 이관 등 질문 재설계는 본 SPEC 의 대상이 아니다(별도 카드/다음 release 후보). 본 SPEC 은 렌더와 i18n 만 다룬다.

### Out of Scope — huh 라이브러리 upstream 수정

- huh v2.0.3 소스 자체의 패치·fork·replace 지시는 하지 않는다. 확인 필드 내부 빈 줄은 라이브러리가 제공하는 설정 지점으로만 제어한다.

### Out of Scope — 신규 로케일 콘텐츠 추가

- 기존(en/ko/ja/zh) 번역 표의 빈칸 메우기는 범위 안이지만, 새 로케일 추가나 번역 체계 개편은 하지 않는다.
