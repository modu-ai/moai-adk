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

# plan.md — SPEC-HARBOR-KITE-001

## §A Context

GH issue #1682의 잔여 사안을 카드 t691로 처리한다. t400(커밋 `25b341dcb`)이 문서화한 공유 플래그 슬롯(`cachedGrowthBookFeatures.tengu_harbor_kite`)은 기기 전역 last-writer-wins 슬롯이고, `ANTHROPIC_BASE_URL` 호스트가 `api.anthropic.com`인 세션만 fetch/기록한다. moai 런처는 `CLAUDE_CODE_HARBOR_KITE`를 설정하지 않으므로, 슬롯이 false인 기기의 GLM/게이트웨이 세션은 피어 메시징 도구(ListAgents/SendMessage)를 잃는다. 카드의 물음은 "주입이 안전한가"였고, 판정은 **비주입(b)** 이다. 본 카드의 산출물은 판정 기록 SPEC + 재현 기록 + 이슈 회신 초안 3가지이며, run 단계(구현)는 없다.

## §B Known Issues

- 상류 플래그 `CLAUDE_CODE_HARBOR_KITE`는 사설·비문서 — 계약 불안정(G1).
- 슬롯은 사용자의 기기 전역 상태로, 런처 강제는 사용자 의사를 덮어쓴다(G2).
- 유출은 glm 고유가 아니라 서드파티 클래스 전체의 속성이다(G3) — 부분 주입은 불일치, 전면 주입은 G2 증폭.
- 유실이 조용하다(quiet failure) — nudge 층 부재는 설계상 눈에 띄지 않는다(G4). 이것이 doctor 검사 후보의 존재 이유다.

## §C Pre-flight

- [x] 워크트리 격리 확인: `.claude/worktrees/t691`, 브랜치 `WT-harbor-kite-judgment`, base `e7b93c120`
- [x] SPEC ID 선행 검사: `SPEC-HARBOR-KITE-001` → regex `PASS` (Bash 실행, 2026-09-13)
- [x] ID 중복 확인: `.moai/specs/`에 HARBOR/KITE/Messaging 계열 SPEC 없음
- [x] 재현 프로파일(`~/.moai/claude-profiles/t691-true`/`t691-false`) 미수정 확인 — 이 레인은 읽지도 쓰지도 않음

## §D Constraints

- 이 카드에서 구현/코드 변경 금지 — 처분이 판정(b)이므로 run 단계 없음.
- 커밋 금지 — 산출물은 오케스트레이터의 검토 커밋을 위해 미커밋으로 남긴다.
- 두 프로파일 디렉터리는 오케스트레이터 정리 대상 — 손대지 않는다.
- SPEC 본문은 WHAT/WHY만 — 주입 코드 형태, doctor 검사의 세부 설계는 다루지 않는다.

## §E Self-Verification

- [x] spec.md §2에 처분(b) + G1-G4 기록 (REQ-002)
- [x] spec.md §3에 두 셀 재현 + 그대로의 핵심 출력 기록 (REQ-003)
- [x] `.moai/reports/t691/issue-reply-draft.md` 존재, 5개 필수 요소 포함 (REQ-005)
- [x] REQ-001(비주입)이 grep으로 검증 가능한 형태 — acceptance.md AC-004 참조

## §F Milestones

우선순위 순이며 시간 추정은 두지 않는다. **이 카드는 M1-M2로 닫힌다 — 구현 마일스톤은 존재하지 않는다(처분 (b)).**

| # | 내용 | 우선순위 | 상태 |
|---|------|----------|------|
| M1 | 판정 결정 기록 SPEC 작성 — 처분(b) + G1-G4 + 재현 기록을 `.moai/specs/SPEC-HARBOR-KITE-001/`의 Tier M 세트(spec.md/plan.md/acceptance.md/progress.md)로 남긴다. 이 카드의 최상위 변경-가능성 결정(판정 자체)이므로 선두. | High | 완료 (plan 단계에서 작성) |
| M2 | #1682 한국어 회신 초안 작성 — `.moai/reports/t691/issue-reply-draft.md`. 기제 설명 + 신규 행동 재현 + G1-G4 완곡어 + 세션별 탈출구 + doctor 후속 후보. | High | 완료 (plan 단계에서 작성) |
| M3 (해당 없음) | 구현 단계 — 처분이 비주입이므로 승인되지 않음. 이 행은 "구현이 없음"을 명시하기 위해 존재한다. | — | N/A |
| 후속 (신규 카드 제안) | `moai doctor` 서드파티+slot=false 감지 검사 — 리드가 신규 카드로 접수할 것. | Medium | 제안됨 |

## §G Anti-Patterns

- "편의를 위해 그냥 =1 박자" — G2(사용자 전역 설정 덮어쓰기)를 재발시키는 발상.
- glm에만 주입 — G3(프로바이더 불일치).
- 플래그 부재를 결함으로 보고 런처 보상 구현 착수 — G4(유계 피해) 위반, run 단계 무단 진입에 해당.
- 슬롯 값을 이 카드에서 직접 수정 — 사용자 전역 상태이며 t400 측정 때부터 금지.

## §H Cross-References

- `.moai/specs/SPEC-HARBOR-KITE-001/spec.md` — 결정 기록 본문
- `.moai/specs/SPEC-HARBOR-KITE-001/acceptance.md` — AC 행렬
- `.moai/reports/t691/issue-reply-draft.md` — 회신 초안 (D2)
- 카드 t400 커밋 `25b341dcb` — 슬롯 기제 문서화
- `.claude/rules/moai/workflow/cross-session-messaging-detail.md` § The shared flag slot — 기제 SSOT
