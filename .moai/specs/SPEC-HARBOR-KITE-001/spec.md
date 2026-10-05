---
id: SPEC-HARBOR-KITE-001
title: "런처는 CLAUDE_CODE_HARBOR_KITE를 주입하지 않는다 — 공유 플래그 슬롯 비주입 판정 기록"
version: "0.1.0"
status: completed
created: 2026-09-13
updated: 2026-09-13
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "cross-session-messaging, harbor-kite, feature-flag-slot, launcher, t691, issue-1682"
issue_number: 1682
tier: M
---

# SPEC-HARBOR-KITE-001 — 런처의 CLAUDE_CODE_HARBOR_KITE 비주입 판정 기록

## HISTORY

### 2026-09-13 — 최초 작성 (카드 t691, GH issue #1682 잔여)

카드 `t691`에서 열렸다. 발단은 GH issue #1682의 잔여 사안이다: 카드 `t400`(커밋 `25b341dcb`)이 피어 메시징의 공유 플래그 슬롯 기제를 문서화했지만(`.claude/rules/moai/workflow/cross-session-messaging-detail.md` § The shared flag slot), moai 런처는 의도적으로 `CLAUDE_CODE_HARBOR_KITE`를 설정하지 않는다. 그 결과 config dir의 `.claude.json`에 있는 기기 전역 슬롯 `cachedGrowthBookFeatures.tengu_harbor_kite`가 false인 기기에서는 GLM 세션(서드파티 백엔드, 슬롯을 기록하지 않음)이 교차 세션 메시징을 잃는다.

카드의 물음은 판정이다: 런처가 이 env를 주입하는 것이 안전한가(상류 사설 플래그의 부작용, 프로바이더 일관성, 사용자 설정 우선순위). 판정 결과는 **비주입(b)** 이며 그 근거 G1-G4와 새 행동 재현을 본 SPEC이 결정 기록으로 남긴다. 이 카드의 처분은 판정이므로 run 단계는 이 주기에서 승인되지 않았다.

작성 위치: 워크트리 `.claude/worktrees/t691`, 브랜치 `WT-harbor-kite-judgment`, base `e7b93c120`. 재현에 쓰인 두 프로파일 디렉터리(`~/.moai/claude-profiles/t691-true`, `t691-false`)는 이 레인이 아니라 오케스트레이터가 정리한다.

## §1 요구사항 (GEARS)

### REQ-001 — 런처 비주입 (Unwanted)

The launcher (`moai cc`, `moai glm`, `moai cg`) shall not set the environment variable `CLAUDE_CODE_HARBOR_KITE` in any session it spawns.

### REQ-002 — 판정 기록 의무 (Event-driven)

**When** a launcher-level decision about `CLAUDE_CODE_HARBOR_KITE` injection is made, the decision record shall state the disposition together with the safety grounds G1-G4 (§2), so a future reader can re-litigate the decision against its original reasons rather than rediscovering them.

### REQ-003 — 재현 기록 보존 (Event-driven)

**When** this card closes, the reproduction record (§3) shall preserve the two-cell behavioral comparison with the verbatim key outputs of both cells, so the tool-availability consequence is evidenced rather than asserted.

### REQ-004 — 유실의 성질 규정 (Ubiquitous)

The loss of cross-session messaging on a third-party-backend session whose shared flag slot reads false shall be treated as bounded degradation of a nudge layer — MoAI's delegation channel is the queue on disk, and the channel self-heals when any first-party session next writes the slot — not as a defect the launcher must compensate for.

### REQ-005 — 이슈 회신 초안 (Event-driven)

**When** the lead prepares the #1682 reply, the lead shall post from the Korean draft at `.moai/reports/t691/issue-reply-draft.md`, which carries the mechanism explanation, the two-cell reproduction, the G1-G4 rationale, the per-session manual escape hatch, and the doctor-check follow-up candidate.

## §2 판정 결정 기록 (Decision Record)

**처분: (b) 주입하지 않는다.** 런처는 `CLAUDE_CODE_HARBOR_KITE`를 자유로운 상태로 둔다. 근거:

- **G1 — 상류 사설 플래그.** 문서화되지 않은 내부 플래그로, 이름과 의미가 예고 없이 바뀔 수 있다. 런처에 박아 넣은 기본값은 moai를 불안정한 계약에 결합시키고, 플래그가 나중에 메시징 외의 것을 게이트하기 시작하면 세션 동작이 조용히 바뀐다.
- **G2 — 사용자 전역 설정 침해.** 슬롯은 사용자의 기기 전역 협상 상태다. `=1` 강제는 슬롯이 false인 이유가 무엇이든(비필수 트래픽 제한 계열 플래그로 플래그 평가 자체를 끈 경우 포함) 그 이유를 덮어쓰며, moai 서드파티 세션 전부에 적용된다.
- **G3 — 프로바이더 일관성.** 게이트웨이 GPT 세션(`ANTHROPIC_BASE_URL=127.0.0.1`)도 같은 슬롯 비기록자다. 유출은 glm 고유가 아니라 moai 전체의 서드파티 클래스 속성이므로, glm에만 주입하면 불일치고 전부에 주입하면 G2를 증폭시킨다.
- **G4 — 피해 유계.** 이 채널은 디스크 큐 delegation 위의 nudge 층이다. 부재는 편의를 깎고 설계상 조용히 진행되며, 어떤 1st-party 세션이 다음에 슬롯을 기록하면 자가 치유된다.

**건설적 대안 (이 카드가 아니라 신규 카드로 리드에게 제안):** `moai doctor` 검사 — 서드파티 백엔드 세션이면서 슬롯이 false인 조합을 감지해, 문서화된 세션별 탈출구를 안내한다. 강제 없는 관측 가능성이다.

## §3 행동 재현 기록 (2026-09-13, 이 레인 측정)

t400은 소켓 존재를 쟀다. 이 재현은 행동 수준(도구 가용성)까지 닫는다. 방법: `/tmp` 스크래치에서 `claude -p` 원샷 2회, `CLAUDE_CONFIG_DIR`을 조립된 프로파일 디렉터리로 지정(격리된 config — 실제 `~/.claude.json`의 슬롯은 수정하지 않았고 현재 true로 읽힌다). 두 세션은 동일 기기/바이너리/환경, glm-5.3-flash (z.ai base URL — `moai glm` 세션 클래스), `.claude.json` 슬롯 값만 다르다.

| Cell | 슬롯 `tengu_harbor_kite` | ListAgents 도구 | 관측된 핵심 출력 (그대로) |
|------|--------------------------|-----------------|---------------------------|
| t691-true | true | 존재 | "peer messaging itself is available" |
| t691-false | false | 툴셋에서 완전 제거 | "There is no tool named ListAgents in my available toolset" |

슬롯 값 하나가 도구 가용성을 좌우한다는 인과가 격리된 config로 확인됐다.

## §4 Out of Scope

이 SPEC은 판정 기록이다. 다음은 이 카드의 범위 밖이다.

### Out of Scope — 런처 주입 구현

- 처분이 (b)이므로 `CLAUDE_CODE_HARBOR_KITE` 주입 코드는 어떤 런처 경로(`moai cc`/`moai glm`/`moai cg`)에도 작성하지 않는다.
- 이 카드에서 코드 변경은 일절 없다.

### Out of Scope — doctor 검사 구현

- §2의 `moai doctor` 슬롯 감지 검사는 신규 카드 후보로 리드에게 제안되는 것이지 이 카드의 산출물이 아니다.
- 이 카드는 그 검사의 요구사항이나 구현을 정의하지 않는다.

### Out of Scope — 이슈 회신 게시와 프로파일 정리

- #1682에 회신을 게시하는 행위는 리드 소관이며, 이 레인은 초안(`.moai/reports/t691/issue-reply-draft.md`)만 남긴다.
- 재현에 쓰인 `~/.moai/claude-profiles/t691-true` / `t691-false` 정리는 오케스트레이터 소관이다.

## §5 교차 참조

- 카드 `t400` — 공유 플래그 슬롯 기제의 최초 측정: 커밋 `25b341dcb`, `.claude/rules/moai/workflow/cross-session-messaging-detail.md` § The shared flag slot
- `.claude/rules/moai/workflow/cross-session-messaging.md` — 채널 가용성 제약과 nudge/delegation 위계
- `.moai/reports/t691/issue-reply-draft.md` — #1682 한국어 회신 초안 (D2)
- GH issue #1682 — 최초 보고
