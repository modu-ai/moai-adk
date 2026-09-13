---
id: SPEC-HARBOR-KITE-001
title: "런처는 CLAUDE_CODE_HARBOR_KITE를 주입하지 않는다 — 공유 플래그 슬롯 비주입 판정 기록"
version: "0.1.0"
created: 2026-09-13
author: manager-spec
priority: P2
module: "internal/cli"
lifecycle: spec-anchored
tier: M
---

# acceptance.md — SPEC-HARBOR-KITE-001

이 SPEC의 인수 조건은 **기록의 존재와 검증 가능성**을 재며 — 카드는 판정 문서화, 재현 기록, 회신 초안 산출 3가지가 갖춰지면 닫힌다. 구현이 수반되지 않으므로 코드 커버리지형 기준은 없다.

## §D AC Matrix

| AC-ID | 요구 연결 | 검증 방법 | 심각도 |
|-------|-----------|-----------|--------|
| AC-001 | REQ-002 | spec.md §2에 처분 (b)과 G1-G4 4항목 존재 — Read로 확인 | Must |
| AC-002 | REQ-003 | spec.md §3 재현 표에 두 셀과 각 셀의 그대로의 핵심 출력 존재 — Read로 확인 | Must |
| AC-003 | REQ-005 | `.moai/reports/t691/issue-reply-draft.md` 존재 + 5개 필수 요소(기제/재현/G1-G4/탈출구/doctor 후보) 포함 — Read로 확인 | Must |
| AC-004 | REQ-001 | `git grep -n "CLAUDE_CODE_HARBOR_KITE" -- '*.go' ':!internal/template'` → 0히트 | Must |
| AC-005 | REQ-004 | spec.md §2의 G4에 nudge 층 규정 + 디스크 큐 위임 채널 인용 + 자가 치유 조건 존재 — Read로 확인 | Must |

### AC-001 — 판정 문서화

**Given** plan-auditor가 `.moai/specs/SPEC-HARBOR-KITE-001/spec.md`를 연다
**When** §2 판정 결정 기록을 읽는다
**Then** 처분이 "(b) 주입하지 않는다"로 명시되어 있고, G1(상류 사설 플래그), G2(사용자 전역 설정 침해), G3(프로바이더 일관성), G4(피해 유계) 4개 근거가 각각 서술되어 있으며, 신규 카드로 넘길 doctor 검사 제안이 포함되어 있다.

### AC-002 — 재현 기록

**Given** t400의 소켓 수준 측정이 문서화된 상태에서
**When** spec.md §3의 재현 표를 읽는다
**Then** (1) 두 셀(t691-true / t691-false)이 슬롯 값만 다른 동일 조건 세션으로 기술되어 있고, (2) true 셀에는 "peer messaging itself is available", false 셀에는 "There is no tool named ListAgents in my available toolset"이 각 셀의 관측 출력으로 실려 있으며, (3) 실제 `~/.claude.json` 슬롯을 수정하지 않았다는 격리 조건이 기록되어 있다.

### AC-003 — 회신 초안 산출

**Given** 리드가 #1682에 회신을 준비한다
**When** `.moai/reports/t691/issue-reply-draft.md`를 읽는다
**Then** 한국어 초안이 (1) 슬롯 기제 설명, (2) 두 셀 행동 재현, (3) G1-G4의 완곡한 비주입 근거, (4) 세션별 수동 탈출구(`export CLAUDE_CODE_HARBOR_KITE=1`), (5) doctor 검사 후속 후보 — 5개를 모두 담고 있고, 보고자에 대한 존중 어조로 쓰여 있다.

### AC-004 — 비주입 코드 부재

**Given** 레포의 구현 트리(Go 소스 전체)
**When** `git grep -n "CLAUDE_CODE_HARBOR_KITE" -- '*.go' ':!internal/template'`를 실행한다
**Then** 0히트다 — 런처는 플래그를 자유로운 상태로 둔다는 REQ-001의 기계적 증거다. `internal/template`은 배포 문서 미러(룰 마크다운)를 담고 있어 문서 언급이 적중하지 않도록 pathspec에서 제외하며, `.claude/`·`.moai/` 문서의 언급도 애초에 이 검사 범위 밖이다.

### AC-005 — 유실의 성질 규정 (REQ-004)

**Given** 서드파티 백엔드 세션에서 슬롯이 false여 메시징이 보이지 않는 상황을 검토한다
**When** spec.md §2의 G4 근거를 읽는다
**Then** G4가 이 유실을 "디스크 큐 delegation 위의 nudge 층의 유계된 성능 저하"로 규정하고, 큐-온-디스크 위임 채널을 명시적으로 인용하며, 자가 치유 조건(1st-party 세션이 다음에 슬롯을 기록)을 기술한다 — 유실을 런처가 보상해야 할 결함으로 규정하는 문구는 없다.

## §D.1 경계 사례

- 슬롯이 true로 자가 치유된 기기 — AC-004는 여전히 성립해야 한다 (비주입은 슬롯 값과 무관한 불변식).
- 상류가 플래그 이름을 바꾸는 경우 — 이 SPEC의 판정 근거 G1이 그 변화를 예상한 것이므로, 기록은 유효하고 재판정은 REQ-002의 절차(새 근거와 함께 재기록)를 따른다.
- 사용자가 셸 프로파일에서 스스로 `CLAUDE_CODE_HARBOR_KITE=1`을 export — 허용된다. REQ-001은 런처의 자동 주입만 금지한다.

## §D.2 품질 게이트

- plan-auditor가 이 아티팩트 세트를 감사해 PASS — 본 카드의 게이트다.
- 산출물 5종(spec 4종 + 회신 초안)이 오케스트레이터 검토 커밋 전에 모두 존재.
- 불변식: 어떤 마일스톤도 `internal/`, `cmd/`, `pkg/`의 코드 변경을 포함하지 않는다.

## §D.3 Definition of Done

1. AC-001~AC-005 전부 충족 (Must — 하나라도 미충족이면 카드 미닫힘)
2. progress.md §E.1에 plan 단계 감사 준비 신호 기록
3. `.moai/reports/t691/issue-reply-draft.md`가 리드가 그대로 게시할 수 있는 완성형 한국어 문안
4. 미커밋 상태로 오케스트레이터에 보고 — SPEC-ID, 파일 경로, 요약 동봉
