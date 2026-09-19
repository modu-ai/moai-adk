# SPEC-WORKTREE-EXIT-RETURN-001 — 인수조건

> 각 항목은 Given-When-Then 이며 이진 판정 가능하다. 문서 층 항목의 판별식은 **대상 파일에서
> 실행 가능한 문자열 검사**로 적는다. 판별식이 인용하는 문자열은 M2 개정이 실제로 심는 문구와
> 일치해야 한다.

## AC-WXR-001 — 연쇄 케이스의 복귀 지점이 문서에 못 박혀 있다

**Given** 개정된 `.claude/rules/moai/workflow/worktree-integration.md` 의
§ `EnterWorktree` / `ExitWorktree` Tools 절이 있고,
**When** 독자가 `A → EnterWorktree B → EnterWorktree C → ExitWorktree` 의 복귀 지점을 찾을 때,
**Then** 그 절은 복귀 지점을 **primary checkout** 으로 지목하고 **B 가 아님을 명시적으로
부정**하는 문장을 담는다. 판별식: 해당 절 안에서 「primary checkout」과 「not the previous
worktree」(또는 같은 뜻의 한국어 부정 문구)가 각각 1회 이상 적중한다.

## AC-WXR-002 — 한 응답이 두 트리를 지목한다는 경고가 있다

**Given** 같은 절이 있고,
**When** 독자가 `ExitWorktree` 응답의 읽는 법을 찾을 때,
**Then** 그 절은 `Your work is preserved at …` 가 **방금 떠난 트리**를 가리키고
`Session is now back in …` 이 **복귀 지점**을 가리킨다는 구분을 담으며, 앞 문장을 복귀 지점으로
읽는 것이 오독임을 명시한다. 판별식: 두 응답 문구가 각각 1회 이상 인용돼 있고, 그 사이에 구분을
선언하는 문장이 있다.

## AC-WXR-003 — 「originating checkout」 문장이 한 가지로만 읽힌다

**Given** 개정 전 문장 "`ExitWorktree` returns to the originating checkout" 이 있고,
**When** 세션이 **워크트리 안에서 출발한** 경우를 대입할 때,
**Then** 개정된 문장은 복귀 지점이 세션 출발 디렉터리가 **아니라** primary checkout 이었다는
관측을 담아, "originating" 이 출발 디렉터리로 읽히는 경로를 남기지 않는다. 판별식: 개정 후 그
절에 무자격 「originating checkout」 단독 문장이 남아 있지 않다(적중 0).

## AC-WXR-004 — 재측정: 응답 문자열이 아니라 작업 디렉터리를 직접 읽는다 [증거의 약한 고리]

**Given** 이 카드가 소유한 버리는 워크트리 두 개(B, C)와 primary checkout A 가 있고,
**When** 한 세션이 `A → EnterWorktree(B) → EnterWorktree(C) → ExitWorktree` 를 수행한 뒤
**Exit 직후 작업 디렉터리를 직접 판독**하고, 같은 회차에 **런타임 버전을 기록**할 때,
**Then** 다음 세 가지가 증거 기록에 남는다:

1. Exit 직후 판독된 작업 디렉터리 경로 **한 줄의 원문 출력** (도구 응답 문자열이 아니라 판독값).
2. 측정 시점의 **런타임 버전 문자열**.
3. 그 판독값이 A / B / C 중 어느 것과 일치하는지의 판정.

판독값이 관측 1~4 와 어긋나면 **어긋남 자체가 기록**되며, 두 회차를 합쳐 「고정」이라고 쓰지
않는다. 이 AC 는 판독값이 A 일 것을 요구하지 않는다 — **판독과 버전 기록이 존재할 것**을
요구한다.

## AC-WXR-005 — 서술이 관측에 귀속돼 있다

**Given** 개정된 절이 있고,
**When** 독자가 복귀 지점 서술의 근거를 물을 때,
**Then** 그 절은 (a) 관측 기록 경로 `.moai/reports/t965/observations.md`, (b) 측정 날짜,
(c) AC-WXR-004 가 기록한 런타임 버전을 함께 담으며, 보장 문언(`always` / `guaranteed` / 「항상」)
으로 쓰지 않는다. 판별식: 세 항목이 모두 적중하고, 복귀 지점 문장에 보장 문언이 적중하지 않는다.

## AC-WXR-006 — 미재현 불일치가 열린 채로 기록돼 있다

**Given** 원 제보(다른 워크트리로 복귀)와 이 조사의 관측(primary checkout 복귀)이 다르고,
**When** 독자가 그 불일치의 상태를 찾을 때,
**Then** 그 절(또는 그 절이 가리키는 기록)은 불일치를 **재현되지 않음 / 원인 미확립**으로
적으며, 원인을 부여하는 문장을 담지 않는다. 판별식: 「재현되지 않았다」류 문구가 적중하고,
원인 단정 문구(구버전·측정 오류 등)가 적중하지 않는다.

## AC-WXR-007 — 템플릿 미러가 같은 커밋에서 같은 내용을 담는다

**Given** 로컬 독트린과 `internal/template/templates/` 미러가 있고,
**When** M2 의 개정이 커밋될 때,
**Then** 두 사본의 해당 절이 같은 복귀 지점 서술을 담으며, 같은 커밋에 포함된다. 판별식:
AC-WXR-001 · AC-WXR-005 의 판별 문자열이 두 사본 모두에서 적중한다.

## Definition of Done

- AC-WXR-001 ~ AC-WXR-007 전부 PASS.
- 코드 변경 0 (`internal/`, `pkg/`, `cmd/` 무변경).
- 재측정 기록이 `.moai/reports/t965/` 아래에 남아 있고, AC-WXR-005 의 인용이 그 파일을 가리킨다.
- Gaps 격상 없음 — 특히 원 제보 불일치에 원인이 부여되지 않았다.
