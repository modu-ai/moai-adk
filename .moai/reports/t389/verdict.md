# t389 — docs-site 4로케일이 폐기된 증거 인용 규약을 가르치던 문제

카드: t389 · 브랜치 `WT-evidence-citation-docs` · 베이스 로컬 develop `b7462203a`(fast-forward)

## Claim

docs-site 4로케일에서 `.moai/state/verify/<session>/` 를 **감사 시점 인용 대상**으로 서술하던 3개 페이지 계열을,
SPEC-EVIDENCE-CITATION-CANON-001(t375, 착지)이 세운 정본 — 스크래치 포획 → 추적 경로 반출 → 파일 하나 인용 —
에 맞게 고쳤다. 카드가 지목한 5개 계열 중 2개는 위반이 아니어서 손대지 않았다.

## Evidence

측정 트리: 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t389`, HEAD `b7462203a`(편집 전).

**카드 전제 재측정** — `grep -rln 'state/verify' docs-site/content | wc -l` → `19`
(en 5 · ja 5 · zh 5 · ko 4). README 4로케일 `grep -rln 'state/verify' README*.md` → 0건. 카드 수치와 일치.

**전제 부분 반증 — 5계열 중 2계열은 REQ-ECC-006 기계 소비자 예외다.**

| 페이지 계열 | 서술 내용 | 판정 |
|---|---|---|
| `advanced/manager-lead` | fold 절차가 `.moai/state/verify/.../M2.*` 를 §E.2 evidence 값으로 지시 | **위반** — 수리 |
| `advanced/agent-guide` | 같은 fold 절차 + mermaid `S1` 노드 라벨 | **위반** — 수리 |
| `advanced/token-budget` | "gitignored 공간이지만 감사 시점에 그대로 열어볼 수 있다" (거짓 전제 그 자체) | **위반** — 수리 |
| `utility-commands/moai-gate` | `moai verify record` 가 쓰는 공유 스냅샷 스토어 | **예외** — 미변경 |
| `advanced/moai-web-console` | SSE 감시 대상 표의 `verify` 행 | **예외** — 미변경 |

근거: `spec.md` REQ-ECC-006 이 예외 대상 둘을 이름으로 열거한다 —
스냅샷 저장소(`internal/verify/store.go`, `moai verify record|check`)와 SSE 감시 소스(`internal/web/events.go`).
두 페이지가 서술하는 것이 정확히 그 둘이다.

**문구가 아니라 요구 성질로 재훑기** — `.moai/state/` 를 증거/인용 맥락에서 언급하는 다른 표현이 있는지
(`grep -rn '\.moai/state/' docs-site/content --include='*.md'` 에서 `state/verify` 제외 후
evidence/증거/証拠/证据/cite/인용/引用 교차) → 8건 전부 `state/goal`(goal 상태)·`state/`(loop 잔여 이슈)로
이 카드와 무관. 즉 이 카드에서는 문구 스윕과 요구 스윕의 결과가 일치한다.

**편집 후 잔여 인용** — `residual-citations.txt` (19행). 잔여 19건 전부 정당하다:
- `*/advanced/manager-lead.md` 각 2건 — fold step 1 의 스크래치 포획 경로 + 말미 `ls` 스크래치 확인 명령
- `en|ja|zh/advanced/agent-guide.md` 각 1건 — step 1 의 스크래치 포획 경로(ko 는 축약판이라 경로를 이름 붙이지 않음)
- `*/advanced/token-budget.md` 각 1건 — 스크래치 성격 서술
- `*/advanced/moai-web-console.md` · `*/utility-commands/moai-gate.md` 각 1건 — 위 예외

**4로케일 동시 수정** — `diffstat.txt`. 12파일 / +63 −49. `.moai/reports` 인용 수 로케일별 대조:

```
       manager-lead  agent-guide  token-budget
ko          6             1            2
en          6             3            2
ja          6             3            2
zh          6             3            2
```

en/ja/zh 완전 일치. ko `agent-guide` 만 1인 이유는 아래 Gaps 참조.

**빌드** — `hugo --quiet` exit 0, 출력 로그 0바이트(`hugo-build.log`), `sitemap.xml` 생성 확인.
mermaid 방향: 편집한 페이지 전부 `flowchart TD` 만, `LR`/`RL` 0건.

## Baseline-attribution

편집 전/후 모두 이 워크트리, 이 트리에서 잰 값이다. 카드 본문의 19(5·5·5·4)는 lane-8 이 2026-08-31 에 잰 값인데
이 트리에서 재측정해도 같은 값이 나왔다 — 즉 인계받은 수치가 아니라 자체 측정으로 확인한 값이다.
정본 문면은 이 트리의 `.claude/rules/moai/core/agent-common-protocol.md:268`,
`agent-common-protocol-reference.md:60,72-73`, `.claude/agents/moai/manager-lead.md:143-162` 에서 직접 읽었다.

## Gaps

- **ko `agent-guide.md` 는 다른 3로케일과 구조가 다르다** — 264행 vs 341행이고 헤딩 목록 자체가 다르다
  (en 의 `Context Folding in Three Steps` + `Peer Cross-Verification` 을 ko 는 `컨텍스트 접기와 교차 검증` 한 절로 합침).
  번역 누락이 아니라 선재하는 축약 재작성이다. 이 카드는 축약 구조를 유지한 채 같은 구분(스크래치 → 반출 → 파일 인용)만
  한 문장으로 실었고, 누락된 절을 복원하지 않았다 — 카드 범위 밖의 선재 드리프트다. **별도 카드 후보.**
- 헤딩 수 선재 차이도 그대로 둔다: `agent-guide` ko=15 / en·ja·zh=14, `token-budget` ko=9 / en·ja·zh=8.
  이 카드가 만든 차이가 아니고 이 카드로 좁히지도 않았다.
- **CI 판정 없음** — 브랜치가 미푸시라 이 변경에 대한 CI 판정이 존재하지 않는다. 위 빌드는 로컬 darwin 측정이다.
- `internal/template/templates/` 에 docs-site 미러가 없어 Template-First 대상이 아니다(미러 부재를 확인한 것이지
  미러가 필요 없다고 추론한 것이 아니다).

## Residual-risk

- 로케일 파생 3건 중 ja·zh 는 서브에이전트가 저술했고, 나는 diff 와 잔여 인용 수를 재측정해 확인했으나
  **일본어·중국어 문장의 자연스러움 자체는 재측정 대상이 아니다.** 두 에이전트 모두 페이지별 기존 플레이스홀더 표기
  (`<セッション>` / `<会话>` vs `<session>`)를 따랐다고 보고했고 diff 상으로 그렇게 보인다.
- zh 에이전트가 "`zh/utility-commands/moai-gate.md` 에는 `state/verify` 문자열이 아예 없다"고 보고했으나
  **재측정 결과 1건 있다**(`residual-citations.txt`). 파일을 건드리지 않은 결정 자체는 옳았으므로 산출물에는 영향이 없지만,
  에이전트 보고의 관측 하나가 틀렸다는 사실은 남긴다.
- 이 카드는 사용자 문서만 고친다. 규칙 트리는 t375 가 이미 고쳤고 이 카드에서 재확인만 했다.
