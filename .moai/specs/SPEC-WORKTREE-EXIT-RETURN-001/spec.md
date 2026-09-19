---
id: SPEC-WORKTREE-EXIT-RETURN-001
title: "ExitWorktree 복귀 지점 계약 명시 — 연쇄 진입·워크트리 출발 세션에서의 복귀 지점을 독트린에 못 박는다"
version: "0.1.0"
status: draft
created: 2026-09-19
updated: 2026-09-19
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
tier: S
tags: "worktree, exitworktree, enterworktree, doctrine, contract, return-point, documentation-gap"
---

# SPEC-WORKTREE-EXIT-RETURN-001 — ExitWorktree 복귀 지점 계약 명시

## HISTORY

| Version | Date       | Author       | Change |
|---------|------------|--------------|--------|
| 0.1.0   | 2026-09-19 | manager-spec | 최초 초안. 카드 t965. 근거는 `.moai/reports/t965/observations.md` 의 관측 1~4 이며, 이 SPEC 은 새 측정을 만들지 않는다. |

## 1. Overview

### 1.1 문제 — 무엇이 비어 있나

카드 t965 의 물음은 「복귀 지점이 틀렸는가」가 아니라 **「`A → EnterWorktree B → EnterWorktree C
→ ExitWorktree` 에서 복귀가 B 인가 A 인가」**였다. 조사(`.moai/reports/t965/observations.md`)의
답은 **A 쪽이되, 카드가 상정한 A 가 아니다**:

- 관측 1 (단일 진입, 대조군): A → Enter(C) → Exit ⇒ **A 로 복귀**. 문서 서술과 일치.
- 관측 2 (연쇄): A → Enter(B) → Enter(C) → Exit ⇒ **A 로 복귀. B 가 아니다.** A 와 B 는 서로
  다른 경로 문자열이므로 이 픽스처는 공허하지 않다.
- 관측 3: **한 응답 안에서** `Your work is preserved at …` 는 C 를, `Session is now back in …`
  은 A 를 지목한다. 읽는 사람이 복귀 지점을 오독하는 모양이 여기서 나온다.
- 관측 4 (세션 이력 1회): 워크트리 **안에서 출발한** 세션이 첫 `EnterWorktree` 뒤 Exit 했을 때
  복귀 지점은 출발 디렉터리가 아니라 **primary checkout** 이었다.

즉 관측 범위 안에서 복귀 지점은 **연쇄 깊이와도, 세션 출발 디렉터리와도 무관하게 primary
checkout 으로 고정**되는 것으로 보인다. 카드가 상정한 이분법(B 인가 A 인가)에는 없던 **세 번째
답**이다.

### 1.2 판단 — 런타임 결함이 아니라 계약(문서) 결함이다

관측 1~4 어디에도 「복귀 지점이 틀렸다」고 부를 동작은 없다. 복귀는 일관되고 예측 가능하다.
결함은 **그 동작을 서술하는 문장이 어느 트리를 가리키는지 말하지 않는다**는 데 있다:

- 상위 도구 설명 문장 (1): "return the session to the **original working directory**" — A 쪽을
  시사한다.
- 상위 도구 설명 문장 (2): "Restores the session's working directory to **where it was before
  EnterWorktree**" — **어느 `EnterWorktree`** 인지 말하지 않는다. 연쇄에서는 B 로도 읽히고,
  관측 4 의 경우(워크트리 출발 세션)에는 **거짓으로 읽힌다**.
- 같은 도구군의 `EnterWorktree` 설명은 **정리 대상**에 대해서만 명시적이다("only the new one is
  tracked for exit-time cleanup"). 정리 대상은 한 문장으로 못 박아 두고 **복귀 지점은 같은
  수준으로 못 박지 않았다** — 이것이 공백의 정확한 위치다.
- 이 저장소 자신의 독트린(`.claude/rules/moai/workflow/worktree-integration.md`
  § `EnterWorktree` / `ExitWorktree` Tools)도 같은 공백을 물려받았다: "`ExitWorktree` returns to
  the **originating checkout**" 한 문장뿐이고, 「originating」이 세션 출발 디렉터리를 뜻하는지
  primary checkout 을 뜻하는지 구분하지 않는다. 관측 4 는 이 두 해석을 **갈라 놓는다**.

### 1.3 우리가 고칠 수 있는 표면은 하나뿐이다

[HARD] **상위 도구 설명은 이 저장소의 산출물이 아니다.** `EnterWorktree` / `ExitWorktree` 는
Claude Code 런타임 도구이고 그 설명 문자열은 우리가 편집할 수 없다. 따라서 이 SPEC 은 **우리가
바꿀 수 없는 표면에 대고 요구사항을 쓰지 않는다.** 실제 인도물은:

1. `.claude/rules/moai/workflow/worktree-integration.md` — 이 저장소의 워크트리 독트린 (+ §2.0
   Template-First 규율에 따른 템플릿 미러).
2. 그 서술을 관측에 귀속시키는 인수조건, 그리고 증거의 가장 약한 고리를 닫는 재측정 인수조건.

상위 도구 설명의 모호성 자체는 **관측으로 기록하고 제보 후보로 남길 수는 있으나**, 이 SPEC 의
완료 조건이 될 수 없다(우리 손 밖이다).

### 1.4 관측의 귀속과 한계 (그대로 운반한다 — 격상하지 않는다)

아래는 조사 파일의 Gaps 를 **그대로** 옮긴 것이다. 이 SPEC 은 이것들을 findings 로 격상하지
않으며, 격상하는 요구사항을 쓰지도 않는다.

- **원 사례는 재현되지 않았다.** 다른 레인이 다른 **워크트리**로 복귀했다는 최초 제보와, 이
  조사의 모든 관측(primary checkout 복귀)이 서로 다르다. 그 세션의 출발 디렉터리는 복구
  불가능하므로 불일치는 **열린 채로 남는다**.
- **관측 4 는 세션 이력에서 얻은 1회 관측**이며 의도적으로 재실행하지 않았다.
- `EnterWorktree(name:)` 가 워크트리 안에서 거부되는지는 **시험하지 않았다**.
- **런타임 버전이 기록되지 않았다** — 관측은 버전에 귀속되지 않는다.
- 관측은 **도구 응답 문자열과 환경 갱신 보고**에 의존하며, Exit 직후의 `pwd` 판독이 아니다.

## 2. Requirements (GEARS)

`<subject>` 는 일반화된 주어를 쓴다. 대상 문서는 `.claude/rules/moai/workflow/worktree-integration.md`
(및 그 템플릿 미러)이며, 아래에서 **the doctrine** 으로 부른다.

- **REQ-WXR-001** (Ubiquitous) — The doctrine shall state the observed `ExitWorktree` return point
  as the **primary checkout**, and shall state that the observed return point is independent of the
  `EnterWorktree` chain depth and of the session's launch directory.

- **REQ-WXR-002** (Event-driven) — **When** a reader consults the doctrine's `EnterWorktree` /
  `ExitWorktree` section for the chained case (`A → Enter B → Enter C → Exit`), the doctrine shall
  name the return point unambiguously as A (the primary checkout) and shall state explicitly that
  it is **not** B.

- **REQ-WXR-003** (Event-driven) — **When** the doctrine describes the `ExitWorktree` response, it
  shall state that the response names **two different trees in one message** — the tree just left
  (`Your work is preserved at …`) and the return point (`Session is now back in …`) — and that the
  first is not the return point.

- **REQ-WXR-004** (Ubiquitous) — The doctrine shall reconcile its own sentence "`ExitWorktree`
  returns to the originating checkout" with the observation that a session **launched inside a
  worktree** returned to the primary checkout rather than to its launch directory, so that
  "originating" carries exactly one reading.

- **REQ-WXR-005** (Where — capability gate) — **Where** a return-point statement rests on the
  upstream tool description that this repository cannot edit, the doctrine shall present the
  statement as a **measured observation with its attribution** (observation-record path, measurement
  date, runtime version) rather than as a quotation of a tool contract.

- **REQ-WXR-006** (Unwanted) — The doctrine shall not assert that the originating cross-worktree
  return report was reproduced, and shall not assert a cause for the discrepancy; it shall record
  the discrepancy as open and unreproduced.

- **REQ-WXR-007** (While — state-driven) — **While** the re-measurement of AC-WXR-004 has not been
  performed, the doctrine's return-point statement shall carry the residual-risk note that the
  observations rest on tool-response strings and environment-update notices, not on a working-
  directory read taken after exit.

- **REQ-WXR-008** (Ubiquitous) — The template mirror of the doctrine file shall carry the same
  return-point statement in the same commit as the local copy, per the repository's Template-First
  rule.

## 3. Acceptance Criteria

Tier S 이지만 카드가 세 파일을 명시했으므로 인수조건 본문은 `acceptance.md` 에 둔다. 요지:
AC-WXR-001~003 은 문서 층의 서술을 고정하고, **AC-WXR-004 는 증거의 가장 약한 고리**(응답
문자열 의존)를 닫는 재측정이다 — Exit 직후 작업 디렉터리를 **직접 판독**하고 **런타임 버전을
기록**한다.

## 4. Exclusions

이 SPEC 이 만들지 않는 것 — out of scope 목록이다.

### Out of Scope — 상위 런타임 도구 표면

- `EnterWorktree` / `ExitWorktree` 도구 설명 문자열의 수정. 우리 저장소의 산출물이 아니며,
  요구사항을 걸 수 없다.
- 복귀 지점 자체의 동작 변경(어느 트리로 복귀할지). 관측된 동작은 일관되며, 이 카드는 그것을
  바꾸자고 하지 않는다.

### Out of Scope — 미측정 영역

- 원 제보(다른 워크트리로의 복귀)의 원인 규명. 그 세션의 출발 디렉터리가 복구 불가능하므로 이
  SPEC 에서 닫지 않는다.
- `EnterWorktree(name:)` 가 워크트리 안에서 거부되는지의 판정. 조사에서 시험되지 않았고, 이
  SPEC 은 그에 대한 서술을 만들지 않는다.
- 복귀 지점의 **버전별 귀속**(어느 런타임 버전부터 이 동작인가). AC-WXR-004 는 측정 시점 버전
  **한 점**만 기록하며, 버전 구간을 주장하지 않는다.

### Out of Scope — 코드 층

- Go 코드, 훅, CLI 변경 없음. 이 카드는 계약 명시 카드이며 기능 카드가 아니다.
- 워크트리 생성·폐기·잠금 수명 규율의 변경. 독트린의 다른 절은 건드리지 않는다.
