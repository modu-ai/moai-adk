---
id: SPEC-JEV-SKILL-SUGGESTION-001
title: "research — Jev skill-suggestion guidance skill"
version: "0.1.0"
created: 2026-09-30
author: manager-spec
---

# research.md — SPEC-JEV-SKILL-SUGGESTION-001

입력 보고서: `/Users/goos/MoAI/moai-adk-go/.moai/reports/aitmpl-jev-skill-suggestion-20260922/aitmpl-jev-skill-suggestion-20260922.md`
(2026-09-22, primary 체크아웃 — 읽기 전용 입력). 본 문서는 그것의 증류와, 본 워크트리에서
재검증한 저장소 관측을 운반한다.

## 1. 입력 분석의 판정 (전수 인용 — 본 카드의 설계를 규정)

| 판정 | 항목 | 본 카드에서의 처분 |
|---|---|---|
| 채택(근거) | F2 와이어 변이 후보 3축 (questions 형태·state 형태·필드명) | 측정 축 — 본 카드 OUT (참조만) |
| 채택(측정 게이트로) | 요청 수준 게이트 noul 셋 (주제 독립 3문항, 1 역방향) | 측정 축 — OUT |
| 채택(측정 게이트로) | choice 1 + criteria 맵 와이어 재설계 | 측정 축 — OUT |
| 채택(호출자 생길 때) | 운영 체크리스트 6건 | **본 카드의 핵심 산문** (REQ-JSK-006) |
| 기각 | 목록 은닉 + SKILL.md 직접 주입 (display-only 체인 위배) | **REQ-JSK-005 기록 산문** |
| 보류 | 함수 훅 (CC ≥ 2.1.278 조기 실험) | **REQ-JSK-005 — 미채택 결정 + 전제 기록** (dispatch 의무) |

모드의 기제(관측): `prompt.attachment`(skill_listing) 은닉 → `prompt.submit` 2요청(wide
랭크 + rerank) → 당첨 1개 `<skill_re>` 첨부(`inject:"content"` = SKILL.md 전문). 운영
특성 6건: 결정별 로그 + 세션 1회 announce / NOT_A_TASK origin 필터 /
disable-model-invocation 필터 / SKILL.md 세션 캐시 / 중복 주입 방지 / 표시명↔디렉터리 id
캐논화. 전 경로 fail-open, 목록 훅 항상 동일 답변(프롬프트 캐시 유지).

저자 측정 수치(잘못된 로드 16.8→7.3% 등)는 **자기 주장·미재측정** — 근거 인용 금지.
지표 나눔(오로드 vs 불필요 로드 분리)만 측정 설계에 베낄 수 있다 (측정 축 기록).

## 2. 저장소 재검증 (본 워크트리, 2026-09-30 관측)

| 관측 | 명령/근거 | 결과 |
|---|---|---|
| consumer B 부재 | `find . -iname "*jev*"` + `grep -rn "skill_suggest\|SkillSuggest" internal/ cmd/` | `jev_skill_suggest.go` 없음; 3개 테스트 주석이 "SPEC-JEV-GUARD-001 (t1083) withdrew" 운반 |
| 가드 walk 스코프 | `internal/jevmeasure/gate_demo_test.go:109-132` | `filepath.WalkDir(internal)` + `.go` 비테스트만; 마커 `NearDuplicateMark/LaneQuestionRoute/SkillSuggest`. **마크다운은 스코프 밖 — 그러나 REQ-JEVN-016(iii)(가용성 presenting 금지)이 규범 경계** |
| 살아 있는 Jev 표면 | `internal/cli/mcp_server.go:651-653` | `jev_ask` MCP 래퍼 — `workflow.jev.enabled`(기본 false) 뒤 게이트됨 |
| doctor 표면 | `internal/cli/doctor_jev.go:67-121` | `moai doctor --check "Jev"` — TCP-only 준비성 프로브(판단 요청 안 보냄), 게이트·자격증명·도달성 3행 보고, fail-open |
| 자격증명 SSOT | `internal/jevcred/jevcred.go` | `~/.moai/.env.typesafe` 단일 구현(REQ-JEVC-018), settings.AllFields 의도적 제외(REQ-JEVC-019), 4자 공개 하한(REQ-JEVC-020) |
| 와이어 | `internal/jev/jev.go` | `ModelID = "jev-1.13.0"`(핀), `Questions []Question`(배열), state 문자열 — 모드(객체/객체)와 이형, 모드 와이어를 따르지 않는다는 dispatch 지시의 근거 |
| 스킬 선례 | `.claude/skills/moai-ref-jev-question-design/SKILL.md` + `internal/cli/jev_question_design_skill_test.go` | 분할(스킬=규칙, 카탈로그=도달성), 금지 토큰 4종, 양성 대조 2건, 양 사본 바이트 동일 테스트 — **본 카드의 구조적 템플릿**. `diff -q` 로 양 사본 IDENTICAL 관측 |
| 등록 절차 | `internal/template/catalog.yaml:51-55` + `Makefile:34-36` | 스킬 엔트리 = name/tier(`core`)/path/hash/version; `make build` 레시피가 `gen-catalog-hashes.go --all` 내장 — 손 해시 불필요 |
| 모드 정리(glob) | CLAUDE.local.md §2.3 | `moai update` 가 `.claude/skills/moai*` 글롭 통째 재배포 → `moai-jev-*` 명명도 템플릿 관리 대상 (기계 안전) |
| 중립성 클래스 | `.moai/docs/template-internal-isolation-doctrine.md` §25 + CLAUDE.local.md §2.1 | 금지: SPEC ID·REQ 토큰·감사 인용·내부 날짜·커밋 SHA·macOS 편향 경로·CLAUDE.local 참조 (+C9 자연어 정본형) |

## 3. dispatch 전제 정정 (연구 결과의 핵심)

Dispatch 전제 "consumer B 가 존재하고 스킬이 감싼다" → **관측 결과 부정**. 소비자는
2026-09-22 `SPEC-JEV-GUARD-001` 이 철수했고, 복원은 REQ-JEVG-006 의 세 전제(와이어 결함
수리 → 측정 실행 → baseline 격파) 뒤 후속 SPEC 만 가능하다. 이 정정이 설계를 바꾼다:

- "스킬이 consumer B 를 호출/감싼다" → "스킬이 제안 기제·계약·체크리스트를 가르친다"
- Go 변경 0은 지시서의 관대함이 아니라 **가드 계약의 요구**다 (비테스트 Go 에 소비자
  마커 재진입 = `TestNoConsumerCallPathShips` RED).
- "결과물을 읽는다"의 대상은 consumer B 가 아니라 살아 있는 표면 3종 — 게이트
  (`workflow.jev.enabled`), 자격증명 파일, doctor 준비성 체크. 판단 래퍼의 도달성은
  카탈로그 소유이므로 스킬은 이름조차 대지 않는다(형제 분할 상속).

이 정정은 운영자 확인 항목(NEEDS-CLARIFICATION 마커)을 만들지 않는다 — 철수 사실이 관측되고, dispatch 자체가
"NO product Go code changes" 를 경계로 정해 뒀으며, 복원 경로는 기존 계약(REQ-JEVG-006)이
이미 소유한다. 운영자 결정 신규 필요 없음.

## 4. 설계 결정의 근거

- **명명 `moai-jev-skill-suggestion`**: dispatch literal. §2 glob 관측으로 기계 안전 확인.
  `moai-ref-jev-*` 축 명명과 불일치하는 비용은 이름 검색 시 한 항목의 불규칙 — dispatch
  지정을 이길 이유 아님. catalog tier 는 형제와 같은 `core`.
- **금지 토큰 4종 상속**: §2 선례와 동일 집합. `jev-suggest`(철수 명령명)는 영구 토큰에
  넣지 않고 일회성 AC grep(D.3) 이 담당 — 토큰 집합 확장은 양성 대조 유지보수 비용을
  키우고, 형제와의 대칭을 깬다.
- **user-invocable: true + 모델 자동 로드 허용**: 본문 첫 절이 게이트 확인을 강제하므로
  자동 로드 무해. "기본 꺼짐"은 기능 게이트가 이미 집행(REQ-JEVC-016)하고 스킬은 그
  행동 양식을 산문으로 반복한다.
- **§29 경계 산문**: 스킬 말미의 "What this skill is not" 이 CLAUDE.local.md §29 의
  3등급(완결 판정·병합 승인·큐 변경 무위임)을 제품 언어로 재서술 — 내부 문서 참조 없이.

## 5. 스킬 본문 목차 (M2 가 채울 구조)

1. The gate comes first — 기본 꺼짐, NO SIGNAL, 준비 3표면(게이트 키/자격증명 파일/doctor 체크), 카탈로그 위임 선언
2. What a suggestion is — 신호 정의 + 4 부정 불릿(선택 아님/자동로드 아님/판정 아님/코드 경로 입력 아님)
3. The loading mechanisms that are rejected here — 기각 2건 (은닉+주입 / 상시 훅) + 재검토 조건
4. Operational checklist — 6항목 (호출자 존재 조건부)
5. Question design is a separate discipline — 형제 스킬 위임 + 측정 게이트 상위 소유 선언
6. What this skill is not — §29 경계 산문

sentinel grep 표는 acceptance.md §D.4-§D.7 이 소유한다.

## 6. 측정 축 경계 (본 카드가 하지 않는 것)

질문 변이 등록, 와이어 프로브(운영자 승인 축), 로스터 토큰 재측정, baseline 격파 판정 —
전부 `SPEC-JEV-OPTIN-MEASURE-001` 축. 본 카드의 산문이 "측정이 상위 소유"라고 선언하는
것까지가 경계다.
