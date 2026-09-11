# t605 — moai goal 미등록 동사가 goal 무장으로 흘러감: 재현 기록

- 카드: t605 (Class B, run 단계 — 원인 확립 기록)
- 워크트리: `.claude/worktrees/t605` · 브랜치 `WT-goal-verb-arm` · HEAD `e7b553617`
- 재현 테스트 커밋: `f93bfcd78` (수리보다 먼저 들어간 단독 커밋)
- 적용 규칙: `verification-claim-integrity.md` §2.2(도구 출처 — 설치 바이너리 대신 트리 컴파일), `verification-completeness.md` §1.1(대조군으로 빈 통과 배제)

## 1. 결론

- `moai goal <미등록 단어>`는 cobra 단계에서 막히지 않는다. `goal` 명령이 `Args: cobra.ArbitraryArgs`이고 부모가 있어서 cobra v1.10.2 `legacyArgs`의 unknown-command 분기(루트 명령 전용)가 적용되지 않는다. 인자는 그대로 `runGoalArm`의 조건 문자열이 된다.
- 카드 예시 `statuss`는 **이미 거절된다.** 막는 쪽은 t556의 문장 판정(5단어 하한)이 아니라 t436의 실행 가능성 게이트(`command -v statuss` 실패)다.
- **실제 구멍:** 셸에서 명령으로 해석되는 한 단어 — `stat`, `ls`, `reset`, `done`, `cancel`, `rm`, `help` — 는 두 게이트를 모두 통과해 오류 없이 무장된다.

## 2. 측정 (원인 증거)

명령: `go test ./internal/cli -run TestGoalVerb -count=1 -v` (t605 트리 `e7b553617` 컴파일, 리드 지명 슬롯)

- 결과: exit 0, `=== RUN` 17 / `--- PASS` 17 / `--- FAIL` 0 / `--- SKIP` 0, `ok github.com/modu-ai/moai-adk/internal/cli 1.062s`
- 실행 직전 `ps`로 다른 `go test` 0건 확인(레인 밖 `internal/gateway` 테스트가 끝날 때까지 기다린 뒤 시작).

대조군 6칸 (`TestGoalVerbControls`, 모두 PASS):

| 칸 | 관측 |
|---|---|
| 등록 동사 `status` | 무장 없음, "no armed goal" |
| 등록 동사 `clear` | 무장 없음 |
| 미등록 동사 `statuss` | 거절 — `its first word "statuss" resolves to no command` |
| 실제 조건 `go test ./... exits 0` | 무장됨 |
| 인자 없음 | 도움말 출력, 무장 없음 |
| 빈 문자열 `""` | 거절 |

한 단어 측정 (`TestGoalVerbResolvingTokens`, 단정 없이 로그만):

| 단어 | 무장 여부 |
|---|---|
| `stat`, `ls`, `reset`, `done`, `cancel`, `rm`, `help` | **armed=true** |
| `list`, `show` | armed=false (t436 게이트: 명령으로 해석 안 됨) |

## 3. 격리 확인

- 테스트 헬퍼는 `CLAUDE_PROJECT_DIR=t.TempDir()`로 두고, `goalProjectRoot()`가 그 경로와 다르면 `Fatalf`로 멈춘다(fail-closed).
- 실행 전후 `.moai/state/goal/` 목록(primary·develop 워크트리·t605 워크트리, 파일별 mtime·크기·sha256 앞 16자): 23줄 → 23줄, `diff` exit 0. 실제 goal 상태 파일은 바뀌지 않았다.

## 4. 수리 방향 후보 (결정 대기)

조건은 자유 텍스트라 "동사처럼 생긴 한 단어"를 모두 거절하면 `true`, `make` 같은 정당한 조건을 버린다. 후보:

- **A. 등록 동사의 오타만 거절:** 인자가 한 단어이고 등록 동사(`arm`, `status`, `clear`, `render`)와 편집 거리 2 이하이거나 동사의 접두어일 때 거절하고 맞는 동사를 제안한다. cobra `SuggestionsFor`를 쓰되 `SuggestionsMinimumDistance`를 2로 명시해야 한다(직접 호출하면 기본값이 0). 이 규칙으로 잡히는 것: `stat`(status 접두어), `rm`(arm과 거리 1). 잡히지 않는 것: `ls`, `reset`, `done`, `cancel`, `help`.
- **B. 동사로 오해하기 쉬운 단어를 명시적으로 지정:** cobra `SuggestFor` 필드로 `cancel`·`reset`·`stop`·`done` → `clear`, `show`·`list`·`info` → `status`처럼 사용자가 동사로 착각할 만한 단어를 등록 동사에 연결하고, 한 단어 인자가 여기에 해당하면 거절한다. `help`는 도움말로 보내는 것이 자연스럽다.
- A와 B를 함께 쓰는 것이 가능하다. 둘 다 여러 단어로 된 조건이나 `cmd:`/`model:` 접두어가 붙은 조건에는 적용하지 않아야 과잉 거절을 피한다.

## 5. Gaps

- 설치된 `~/go/bin/moai`(rev `2213871af`, dirty)는 t436·t556을 포함하지 않아 재현에 쓰지 않았다. 실제 CLI 바이너리로 `moai goal stat`을 치는 종단 재현은 하지 않았고, cobra 라우팅 결론은 cobra 소스 판독과 테스트 시접(`newGoalCmd`를 루트에 붙인 테스트 루트)에 근거한다.
- 슬롯 한 번에 한 번만 돌렸다.
