---
id: SPEC-JEV-SKILL-SUGGESTION-001
title: "Jev skill-suggestion guidance skill — explicit opt-in surface, display-only, no call path"
version: "0.1.1"
status: in-progress
created: 2026-09-30
updated: 2026-09-30
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: ".claude/skills/moai-jev-skill-suggestion + internal/template + internal/cli"
lifecycle: spec-anchored
tags: "jev, skill-suggestion, opt-in, display-only, guidance-skill, t1340"
depends_on: [SPEC-JEV-GUARD-001]
tier: M
related_specs: [SPEC-JEV-CONSUMERS-001, SPEC-JEV-OPTIN-MEASURE-001, SPEC-JEV-CORE-001]
---

# SPEC-JEV-SKILL-SUGGESTION-001 — Jev 스킬 제안 가이던스 스킬 (명시 옵트인, display-only, 호출 경로 없음)

## HISTORY

| Date | Version | Change | Author |
|------|---------|--------|--------|
| 2026-09-30 | 0.1.0 | Initial plan-phase artifacts (card t1340, Class C). Input: 09-22 aitmpl 분석 보고서. 배경 §A 에서 dispatch 전제(consumer B 존재)가 SPEC-JEV-GUARD-001 에 의해 철수된 상태임을 관측·정정하고, 설계를 호출-경로-없는 가이던스 스킬로 확정. | manager-spec |
| 2026-09-30 | 0.1.1 | plan-audit iter-1 CONDITIONAL 0.875 반영 — D1(acceptance D.3 검출기 확폭: 다중 세그먼트 SPEC/REQ ID + 내부 날짜 클래스), D2(acceptance D.10 기준을 merge-base 재계산으로 재정의, lane protocol §8 인용), D3(research NEEDS-CLARIFICATION 어휘 정리), D4(RED-now 셀 런페이즈 완비 항목 명시), D5(§D.2 고려-제외 토큰 2종 기록). | manager-spec |

## A. 배경 (Background)

카드 t1340 (Class C, Tier M). 입력은 09-22 분석 보고서
(`.moai/reports/aitmpl-jev-skill-suggestion-20260922/aitmpl-jev-skill-suggestion-20260922.md`,
primary 체크아웃 — 읽기 전용 입력)이다. 보고서는 aitmpl 의 `jev-skill-suggestion`
mod 를 분석하고 (a) 목록 은닉 + SKILL.md 직접 주입 기제를 기각, (b) 운영 체크리스트 6건을
호출자 생길 때 채택, (c) 질문 설계 변이를 측정 축으로 이관한다.

**Dispatch 전제의 관측된 정정 (본 트리, 2026-09-30).** Dispatch 는 "consumer B
(`internal/cli/jev_skill_suggest.go`) 가 존재하고 스킬이 그것을 감싼다"고 전제했으나,
본 워크트리에서 관측한 결과:

- `internal/cli/jev_skill_suggest.go` 는 존재하지 않는다. `internal/cli/jev_question_design_skill_test.go:41`
  이 그 이유를 운반한다: "SPEC-JEV-GUARD-001 (t1083) withdrew jev_skill_suggest.go".
- `SPEC-JEV-GUARD-001` (completed, 2026-09-22) 는 측정 게이트 소유 요구
  (`SPEC-JEV-OPTIN-MEASURE-001` 의 `TestNoConsumerCallPathShips`, 내부 walk 로 비테스트
  Go 전체에서 소비자 마커 부재를 단언) 와 gate-unrun 출하 계약의 충돌을 소비자 **철수**로
  해소했고, 복원 계약(REQ-JEVG-006: 와이어 결함 수리 + 측정 통과 + baseline 격파 후
  후속 SPEC 이 복원)을 기록했다. **그 두 행위가 일어나기 전까지 트리는 소비자 호출 경로를
  가져서는 안 된다.**

따라서 본 카드가 전달하는 것은 **호출 경로가 없는 가이던스 스킬**이다 — `moai-ref-jev-question-design`
선례와 동일한 분할(스킬은 규칙을 가르치고, 도달성은 제품 카탈로그가 소유한다)을 따른다.
스킬은 (1) display-only 계약, (2) 기각된 기제의 기록, (3) 호출자가 생길 때의 운영
체크리스트, (4) 살아 있는 준비 표면(게이트·자격증명·doctor)을 운반한다. **어떤 호출
경로·명령 등록·함수 훅도 만들지 않는다** — 그것은 가드 계약 위반이고, 측정 축의 소관이다.

기능 훅 결정 (dispatch 지시에 따라 명시적으로 기록): **채택하지 않는다.** CC 기능 훅
(조기 실험, CC ≥ 2.1.278 + `CLAUDE_CODE_ENABLE_FUNCTION_HOOKS=1` 전제)은 프롬프트마다
모델 호출을 상용화하고 템플릿 기반 파편화를 낳는다. 본 SPEC 은 그 전제까지를 "평가된
경계"로 기록하고, 채택 여부의 재검토는 측정 근거를 가진 후속 카드의 몫으로 남긴다.

## B. 요구사항 (GEARS)

**REQ-JSK-001** (Ubiquitous) The skill `moai-jev-skill-suggestion` shall ship as two byte-identical copies — the loaded skill (`.claude/skills/moai-jev-skill-suggestion/SKILL.md`) and its template mirror (`internal/template/templates/.claude/skills/moai-jev-skill-suggestion/SKILL.md`). Verified by `AC-JSK-001` and the M2 parity test.

**REQ-JSK-002** (Ubiquitous) Both copies of the skill body **shall** carry no call-path token, where the forbidden token set is `{internal/jev, mcp__moai__jev, jev_ask, moai jev}` — the same set the sibling question-design skill is guarded by. The skill teaches mechanism and contract; reachability lives in the product's MCP tool catalogue. Verified by `AC-JSK-002` (new guard test with positive controls, mirroring `internal/cli/jev_question_design_skill_test.go`).

**REQ-JSK-003** (Ubiquitous) Both copies of the skill body **shall** carry no internal-trace literal — no SPEC ID, no REQ token, no commit SHA, no internal date, no card id (template-neutrality internal-trace classes). Verified by `AC-JSK-003`.

**REQ-JSK-004** (Ubiquitous) The skill body **shall** state the display-only contract: the suggestion answer is a signal a person reads, it never selects a skill, never auto-loads skill content, and the existing selector's selection authority stays unchanged. Verified by `AC-JSK-004`.

**REQ-JSK-005** (Ubiquitous) The skill body **shall** record the rejected mechanisms — (a) concealing the skill listing from context and injecting a chosen SKILL.md body directly, (b) an ambient per-prompt function-hook wiring — as not adopted, with the early-experiment premise of (b) named so a future card can revisit it with measured evidence. Verified by `AC-JSK-005`.

**REQ-JSK-006** (Ubiquitous) The skill body **shall** carry the operational checklist for a future caller: per-decision logging with a one-time session announce; non-task origin filter; invocation-disabled skill filter; per-session roster/body cache; duplicate-injection guard; display-name ↔ directory-id canonicalization — each stated as guidance conditioned on a caller existing, not as a shipped behavior. Verified by `AC-JSK-006`.

**REQ-JSK-007** (Event-driven) **When** the capability gate is off (`workflow.jev.enabled: false`, the shipped default), the skill body **shall** instruct the caller to construct no request, make no network call, and proceed without a signal — an absent signal is NO SIGNAL, never a negative answer. Verified by `AC-JSK-007`.

**REQ-JSK-008** (Event-driven) **When** the template tree is rebuilt (`make build`), the catalog (`internal/template/catalog.yaml`) **shall** carry the new skill's entry (name, tier, path, hash, version) with a regenerated hash, and the build **shall** succeed on the host platform and on `GOOS=windows`, embedding both copies. Verified by `AC-JSK-008`, `AC-JSK-009`.

**REQ-JSK-009** (Unwanted) This SPEC **shall not** create or modify any non-test Go file, register no command, add no hook, and change no settings/config key; the consumer-marker walk (`TestNoConsumerCallPathShips`) **shall** stay green. Verified by `AC-JSK-010`, `AC-JSK-011`.

## C. 제약 (Constraints)

1. [HARD] 소비자 호출 경로 금지 — `SPEC-JEV-GUARD-001` 의 복원 계약(REQ-JEVG-006)이
   성립할 때까지 트리에 SkillSuggest 소비자 패밀리가 존재해서는 안 된다. 본 카드는 그
   복원을 수행하지 않고, 복원을 가리키는 계약 산문만 운반한다.
2. [HARD] Template-First — 템플릿 사본(`internal/template/templates/...`)이 원본이고,
   로컬 사본(`.claude/skills/...`)과 바이트 동일을 유지한다. 편집 후 `make build`
   (카탈로그 해시 재생성 포함) 실행.
3. [HARD] 템플릿 중립성 — SKILL.md 양 사본에 내부 흔적 클래스(SPEC ID·REQ 토큰·커밋
   SHA·내부 날짜·카드 id) 금지; 영어 본문(자연어 정본형 중립성).
4. [HARD] 자격증명 — 키 재질은 `~/.moai/.env.typesafe` 로만 참조된다. 어느 파일에도
   인라인 금지, settings/config 템플릿 편입 금지 (기존 jevcred 계약 유지).
5. [HARD] 전체 스위트(`go test ./...`) 로컬 실행 금지 — lane-local 검증만. 전 판정은 CI 몫.
6. [HARD] `internal/jevmeasure/gate_demo_test.go`, `internal/cli/jev_question_design_skill_test.go`
   를 포함한 기존 가드 테스트 zero edits — 새 가드 테스트는 **새 파일**로 추가한다.

## D. 설계 결정 (Design decisions — 변경 가능성 순)

### D.1 스킬의 정체 — 호출 경로 없는 가이던스 (최고 변경 가능)

질문-디자인 스킬 선례의 분할을 채택한다: 스킬은 제안 **기제와 계약**을 가르치고, 실제
호출은 제품 표면(MCP 도구 카탈로그가 도달성을 소유)의 몫이다. dispatch 가 상정한
"기존 consumer B 를 감싼다"는 전제는 철수 사실로 무효 — 스킬이 호출 경로를 대신하는
것도, 호출을 가르치는 것도 아니다. 스킬 본문 목차: 게이트 우선(기본 꺼짐, NO SIGNAL
원칙, 준비 표면 3종) → 제안이란 무엇인가(신호 4 부정) → 기각된 기제 기록 2건 → 운영
체크리스트 6건 → 질문 설계 위임 → 이 스킬이 아닌 것(§29 경계 산문).

### D.2 금지 토큰 집합 — 형제 선례 상속

새 가드 테스트의 금지 토큰은 형제(`jev_question_design_skill_test.go`)와 동일한 4토큰:
`internal/jev`, `mcp__moai__jev`, `jev_ask`, `moai jev`. 양성 대조 2건도 동일
(`mcp_jev.go` → `internal/jev`, `moai-mcp-tools-catalogue.md` → `jev_ask`). 토큰 집합을
늘리지 않는다 — 철수된 명령명 부재 검사는 일회성 AC grep(AC-JSK-003 별행)이 담당하고,
영구 가드는 형제와 같은 모양을 유지한다.

고려하고 제외한 토큰 2종 (plan-audit D5): `scripts/jev` — 운영자 로컬 전용 경로로 사용자
프로젝트에 배포되지 않고 제품 호출 경로가 아니며, 형제 본문도 0회 참조한다 (관측:
plan-audit verified-tree 3번 항). `SkillSuggest` — Go 측 walk 가드
(`TestNoConsumerCallPathShips`)가 소유하는 마커 계열로, 스킬 본문 위생이 아니라 가드의
소관이고(해당 walk 는 비테스트 `.go` 만 훑는다), 본문 검사 토큰으로 편입하면 소유자가
다른 두 검사가 한 집합에서 갈라진다.

### D.3 스킬 명명 — dispatch literal 유지

`moai-jev-skill-suggestion`. `moai*` 글롭이 템플릿 관리 대상이라(CLAUDE.local.md §2.3)
기계적으로 안전하고, catalog 엔트리 tier 는 형제 선례대로 `core`. 형제 축 명명
(`moai-ref-jev-*`) 과의 정합은 이점이지만 dispatch 지정명을 존존한다.

### D.4 기능 훅 결정 기록 (dispatch 의무)

채택 안 함. 평가된 경계: CC 기능 훅(조기 실험, CC ≥ 2.1.278 +
`CLAUDE_CODE_ENABLE_FUNCTION_HOOKS=1`). 기각 사유: 프롬프트마다 모델 호출 상용화(미측정
행위의 상용화), 템플릿 배포물과의 파편화, early-experiment 표면 변동. 재검토 조건:
후속 카드가 측정 근거(상수 baseline 격파)를 가질 때. 이 결정은 스킬 본문에는 중립 산문으로,
정확한 전제(버전·환경변수)는 본 절과 research.md 에만 기록한다.

### D.5 인보크이션 표면

`user-invocable: true`(명시), model-invocation 기본(가이던스 자동 로드 무해 — 본문 자체가
게이트 확인을 첫 절에 요구). `disable-model-invocation` 필터는 체크리스트 항목(제안
대상 스킬에 대한 것)이지 이 스킬에 대한 것이 아니다.

## E. 제외 (Exclusions)

### Out of Scope — 철수된 제품 소비자의 복원

- `moai jev-suggest` 명령 재등록, `internal/cli/jev_skill_suggest.go` 계열 코드 재착지 —
  `SPEC-JEV-GUARD-001` REQ-JEVG-006 의 후속 SPEC 소관 (와이어 결함 수리 + 측정 격파 선행).
- 측정 게이트의 실행·판정, baseline 격파 여부 — `SPEC-JEV-OPTIN-MEASURE-001` 소관.

### Out of Scope — 측정 행렬

- 질문 설계 변이 2종(요청 수준 게이트 noul 셋, choice 1 + criteria 맵)의 등록·실험 —
  09-22 분석이 이관한 측정 축(OPERATOR 승인 축). 본 카드는 참조만 한다.
- 와이어 프로브 변이(F2 판별 후보 3축) 설계·집행.

### Out of Scope — 상시 배선과 설정

- 함수 훅, `hooks.json`, settings.json / skillOverrides / workflow.yaml 기본값 변경 — 없음.
- 게이트 기본값 플립, 자격증명 경로·스키마 변경 — 기존 jevcred 계약 유지.

### Out of Scope — Go 제품 코드

- 비테스트 Go 파일의 신설·수정 전부. 새 Go 파일은 `internal/cli/` 의 **테스트 파일 하나**뿐이다.
- 기존 가드 테스트·질문-디자인 스킬·MCP 래퍼(`mcp_jev.go`)·doctor 체크의 수정.

## F. 추적 사슬 (Traceability chain)

```
aitmpl-jev-skill-suggestion-20260922 분석 보고서 (입력 — 기각/채택/이관 판정의 출처)
  ← SPEC-JEV-CORE-001 (게이트·와이어·자격증명 — 스킬이 참조하는 준비 표면의 소유자)
  ← SPEC-JEV-OPTIN-MEASURE-001 (측정 게이트 소유자 — TestNoConsumerCallPathShips)
  ← SPEC-JEV-CONSUMERS-001 (소비자 소유자 — REQ-JEVN-016 three-state 규정)
  ← SPEC-JEV-GUARD-001 (철수 + 복원 계약 REQ-JEVG-006 — 본 SPEC 의 최상위 구속 조건)
  → SPEC-JEV-SKILL-SUGGESTION-001 (본 SPEC: 호출-경로-없는 가이던스 스킬)
  → card t1340 (Class C, Tier M)
```

## G. @MX 태그 보고 (계획)

- 신설 Go 파일은 테스트 파일 — 태그 실지 않음 (공개 함수 없음).
- 스킬/카탈로그는 마크다운/YAML — @MX 대상 아님.
추가·갱신·제거 태그 0건 예상. 상세는 run-phase 보고의 `## @MX Tag Report` 참조.

## H. 교차 참조

- `.moai/reports/aitmpl-jev-skill-suggestion-20260922/aitmpl-jev-skill-suggestion-20260922.md` — 입력 분석 (primary 체크아웃, 읽기 전용)
- `SPEC-JEV-GUARD-001` — 복원 계약(REQ-JEVG-006)과 가드(`TestNoConsumerCallPathShips`)
- `SPEC-JEV-CONSUMERS-001` — 소비자 3상태 규정(REQ-JEVN-015/016), display-only 축
- `SPEC-JEV-CORE-001` — 게이트·자격증명·doctor 표면의 소유자
- `.claude/skills/moai-ref-jev-question-design/SKILL.md` — 분할 선례(규칙만, 호출 경로 없음) + 가드 테스트 선례
- `CLAUDE.local.md` §29 — Jev 자율 등급(운영자 위임)과 §29 경계; §2.1 — 템플릿 중립성
- `.moai/docs/template-internal-isolation-doctrine.md` §25 — C1-C8 내부 흔적 클래스
