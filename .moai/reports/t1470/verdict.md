# t1470 verdict — 발견 줄의 drop/edit 제안이 엉뚱한 카드를 가리키던 결함 (GitHub #1732)

- card: t1470 · class B (SPEC 없음) · cycle: tdd
- branch: `WT-todo-finding-subject`
- base: local `develop` `b5815ca80` (fast-forward로 흡수)
- RED commit: `4e21c7fec` · fix commit: `069503329`

## Claim

1. `todoFindingLine`(internal/cli/todo_analysis.go)은 drop/edit 제안을 `cardID` — 그 줄이 **렌더링되는 행** — 로 채웠다. near-duplicate 발견은 쌍의 양쪽 행 아래에 모두 찍히므로, **원본 카드 아래에서는 원본을 drop하라고** 안내했다. `todo why`도 같은 렌더러를 재사용해 같은 결함을 가졌다.
2. 수리 후 제안은 발견의 subject(`f.SubjectID` — 그 발견이 설명하는, 새로 들어온 카드)를 가리킨다. subject 자신의 행 아래에서는 출력이 바이트 단위로 이전과 같다.
3. 이슈의 「행이 보이지 않는다」 절반은 다루지 않았다 — 리더 조사상 list limit / queued 위치 효과로 추정되며 t1313이 이미 완화했다.

## Evidence

RED (수리 전, 커밋 `4e21c7fec`에 테스트만 단독 커밋):

```
$ go test -count=1 ./internal/cli/ -run TestTodoFindingLineSuggestsDroppingTheSubjectNotTheRow
--- FAIL: TestTodoFindingLineSuggestsDroppingTheSubjectNotTheRow (9.58s)
    todo_finding_subject_test.go:50: list: finding line under t1 suggests "moai todo drop t1" — that acts on the original, not the near-duplicate:
        	↳ near-duplicate t2 (mechanical, score 0.83, machine-only) — moai todo drop t1 | moai todo edit t1 "<text>"
    todo_finding_subject_test.go:50: list: finding line under t1 suggests "moai todo edit t1" — that acts on the original, not the near-duplicate:
        	↳ near-duplicate t2 (mechanical, score 0.83, machine-only) — moai todo drop t1 | moai todo edit t1 "<text>"
    todo_finding_subject_test.go:50: list: finding line does not name the subject t2 in its drop suggestion:
        	↳ near-duplicate t2 (mechanical, score 0.83, machine-only) — moai todo drop t1 | moai todo edit t1 "<text>"
    todo_finding_subject_test.go:63: why: finding line under t1 suggests "moai todo drop t1" — that acts on the original, not the near-duplicate:
        1 ↳ near-duplicate t2 (mechanical, score 0.83, machine-only) — moai todo drop t1 | moai todo edit t1 "<text>"
    todo_finding_subject_test.go:63: why: finding line under t1 suggests "moai todo edit t1" — that acts on the original, not the near-duplicate:
        1 ↳ near-duplicate t2 (mechanical, score 0.83, machine-only) — moai todo drop t1 | moai todo edit t1 "<text>"
    todo_finding_subject_test.go:63: why: finding line does not name the subject t2 in its drop suggestion:
        1 ↳ near-duplicate t2 (mechanical, score 0.83, machine-only) — moai todo drop t1 | moai todo edit t1 "<text>"
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	11.118s
```

GREEN (수리 후, 대상 + 인접 테스트):

```
$ go test -count=1 -timeout 30m ./internal/cli/ -run 'TestTodoFindingLineSuggestsDroppingTheSubjectNotTheRow|TestTodoListShowsFindingAndNextStep|TestTodoWhy' -v
--- PASS: TestTodoFindingLineSuggestsDroppingTheSubjectNotTheRow (22.22s)
--- PASS: TestTodoListShowsFindingAndNextStep (7.91s)
--- PASS: TestTodoWhySaysNothingFound (7.09s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	40.795s
```

GREEN (발견 렌더링을 건드리는 테스트 파일 전체 — relate / jev_finding / audit_regression / analysis / analysis_add / finding_subject):

```
$ go test -count=1 -timeout 30m -v -run 'TestJev|TestTodoDoneReclaimsFindings|TestTodoLegacyRecord|TestTodoRelate|TestSemanticRelations|TestTodoList|TestTodoWhy|TestMachineOnly|TestTodoNewVerbs|TestRelateAndUnrelate|TestTodoAudit|TestTodoFindingLine|TestTodoAdd|TestTodoAnaly|TestTodoExactRefusal' ./internal/cli/
--- PASS 82건 · --- FAIL 0건 · --- SKIP 0건
ok  	github.com/modu-ai/moai-adk/internal/cli	162.304s
```

정적 검사:

```
$ go vet ./internal/cli/        → exit 0, 출력 없음
$ gofmt -l internal/cli/todo_analysis.go internal/cli/todo_finding_subject_test.go  → 출력 없음
```

## Baseline-attribution

- 측정 트리: 이 워크트리(`.claude/worktrees/agent-ab43bc24861258df2`), 브랜치 `WT-todo-finding-subject`. RED는 `4e21c7fec`의 트리(local develop `b5815ca80` + 테스트 파일), GREEN은 `069503329` 커밋 직전 워킹 트리(= 커밋된 트리; 그 사이 수정 없음).
- 순서 증인은 커밋 그래프다: RED 테스트 단독 커밋 `4e21c7fec` → 수리 커밋 `069503329`.
- 테스트 바이너리는 매 실행 `go test`가 이 트리에서 새로 컴파일했다(설치된 `moai` 빌드를 쓰지 않았다).

## Gaps

- **`-run 'Todo'` 전체 선택은 판정이 나지 않았다.** 머신 부하(load average 160~210) 때문에 1차는 기본 10분, 2차는 `-timeout 30m`에서 `panic: test timed out`으로 끝났다. 두 실행 모두 타임아웃 이전에 `--- FAIL`은 0건이었지만, 끝까지 돌지 못한 테스트들은 관측되지 않았다. 대신 `todoFindingLine`을 호출·인용하는 테스트 파일 범위로 좁혀 82건을 끝까지 돌렸다.
- `golangci-lint`는 돌리지 않았다.
- 전체 패키지(`./internal/cli/...`)와 전 저장소 스위트는 돌리지 않았다 — 저장소 전체 판정은 local develop 병합 후 리더의 develop push가 일으키는 CI 몫이며, **이 보고 시점에 그 판정은 PENDING**이다.

## Residual-risk

- `todo relate`로 운영자가 직접 만든 관계(`depends`/`blocks`/`relates`)의 subject는 운영자가 지정한 카드다. 이제 제안은 그 subject를 가리키는데, 관계 종류에 따라 「subject를 drop」이 언제나 맞는 행동이라는 보장은 없다 — 다만 수리 전에도 subject 행 아래서는 똑같이 subject를 가리켰으므로 새로 생긴 위험은 아니다.
- analyze 재스윕은 나중 카드(인덱스 j)를 subject로 기록한다. 그래서 near-duplicate 쌍에서는 늘 「새 카드를 drop」이 제안되며, 운영자가 사실은 옛 카드를 버리고 싶은 경우엔 직접 판단해야 한다.
- 이슈의 「행이 보이지 않는다」 절반은 손대지 않았다(Claim 3).
