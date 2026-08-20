# release/v3.1.1 통합 시 알려진 충돌 + 해소본

리드가 실측으로 확인한 충돌만 기록한다. 통합(no-ff 머지) 담당자가 읽고 그대로 적용한다.

## C1 — `.claude/rules/moai/workflow/kanban-dispatch-detail.md:86` (t133 × t137)

두 커밋이 같은 문장의 **다른 부분**을 고쳤다. 실측:

- base(e7aeec088): `… no run id travels in companion names (one run per machine; the lead keeps the id). \`ListAgents\` reports the live set; …`
- t133(c326eb4e0): 앞부분을 고침 — run id 서술
- t137(afff28d6a): 뒷부분을 고침 — ListAgents 완전성 주장

해소본(두 수정의 합집합. lane-3 제안, 리드가 세 판본을 직접 대조해 확인):

```
Dispatch is addressed by session name. Companions are named by their bare role; a name held by a live session is bumped to the next free number, and no run id travels in any session name, the lead's included (one run per machine; the id lives in `MOAI_KANBAN_ID` and the lead socket path). `ListAgents` lists live sessions and says when it could not check them all; send with `SendMessage({to: "<name>", message: "…"})`, using the short reference the listing prints when a bare name is ambiguous.
```

템플릿 미러(`internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-detail.md`)에도 **같은 문장**을 적용할 것 — 두 파일은 바이트 동일이어야 한다.

## C2 — `kanban-dispatch.md` 예산 상쇄 2줄 (t137 × t131) — 예상, 미실측

t137 이 always-loaded 예산(+30B)을 맞추려고 스텁 안 중복 2건을 제거했다:
① `env -u` 가드 문단의 "guard cannot statically track" 중복 ② 워크트리 조항의 `EnterWorktree(<card-id>)` 중복.
t131(항상 로드 예산 구조 정리)이 같은 파일을 구조 정리 중이라 겹칠 수 있다. t131 착지 후 이 두 지점을 대조할 것.

## 예산 실측 (통합 판정용)

- `kanban-dispatch.md`: 23,728 → **23,758** (+30B ≈ +8토큰). 여유 8토큰 안에 들어감 — t131 착지 전이라면 이 여유가 이미 소진된 상태다.

## C2 갱신 — 예상 → **실재** (t131 × t137, `kanban-dispatch.md`)

리드 실측 2026-08-19:
- t137(`afff28d6a`): 23,728 → 23,758 (+30B, ListAgents 완전성 한정 + 중복 2건 제거)
- t131(`3f31135c3`): 23,728 → 23,421 (-307B, 설명 산문 미세 압축; [HARD] 문장은 바이트 그대로 보존)

둘 다 같은 파일을 손댔으므로 통합 시 충돌한다. **해소 원칙**: t137 의 의미 변경(ListAgents 한정 + fault 조항 한정)은 **반드시 살린다** — 그것이 카드의 본체다. t131 의 압축은 바이트 절감이므로, 충돌 구간에서는 t137 본문을 채택하고 나머지 구간의 압축을 유지한다. 해소 후 `kanban-dispatch.md` 최종 바이트를 기록하고 `go test ./internal/config/ -run TestAlwaysLoadedTokenBudget` 로 예산을 재확인할 것.

## 예산 (t131 착지 후, 리드 실측)

4개 파일 합계 157,256 → 135,442 B = **-21,814 B ≈ -5,453 토큰**. lane-5 보고(-5,454)와 일치.
(단, lane-5 보고의 `moai.md` 개별 수치 63,406 은 실측 61,474 와 어긋난다 — 합계는 맞으므로 전사 오류로 보이나, 통합 후 개별 수치를 다시 읽을 것.)
여유: 8토큰 → 약 5,462토큰. 상수 `AlwaysLoadedTokenBudget = 76000` 무변경 확인.
