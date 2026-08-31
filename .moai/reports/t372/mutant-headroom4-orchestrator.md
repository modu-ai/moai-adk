# t372 — AC-SIV-008 뮤턴트 독립 재현 (오케스트레이터)

manager-develop 의 run-phase 증거와 **별개로**, 오케스트레이터가 직접 심고 관측한 기록.

- 트리: `.claude/worktrees/t372`, 브랜치 `WT-stress-invariant-guard`
- 측정 시점 HEAD: `0fa8606fe`
- 원본 verbose 로그는 `*.log` 가 `.gitignore:106` 에 걸려 착지하지 못한다(`git check-ignore -v` 로 확인). 판정에 쓰인 줄을 여기에 그대로 옮긴다.

## 1) 가드 census — 심기 **전에** 열거

```
$ grep -rln 'boardLockHeadroom\|boardLockWaitBudget' --include='*_test.go' internal/
internal/kanban/integration_lock_cross_test.go
internal/kanban/backlog_concurrency_test.go
internal/kanban/board_lock_wait_test.go
```

세 파일이 이 지점을 덮는다. 그래서 셀렉터 부분집합이 아니라 **전 패키지**를 돌렸다 —
셋 중 하나만 잡으면 나머지 둘이 잡았을 가능성을 배제하지 못한다.

## 2) 뮤턴트 (상수 축)

`internal/kanban/board_store.go`: `boardLockHeadroom = 5` → `4`
(예산 1.65s → 1.32s, floor 48 × 33ms = 1.584s 미달)

cost 축 뮤턴트는 심지 않았다 — `boardLockCIMutationCost` 는 부등식 양변에서 소거되므로
어떤 크기로 바꿔도 이 가드를 발화시킬 수 없다(REQ-SIV-009).

## 3) 관측

```
$ go test -race -count=1 -v ./internal/kanban/
exit=1
$ grep -c '^=== RUN' <log>
389
$ grep '^--- FAIL' <log>
--- FAIL: TestBoardLockWaitBudgetCoversSerializedMutations (0.00s)
$ grep 'BoardLockWaitBudgetDerivedFromNamedInputs' <log>
--- PASS: TestBoardLockWaitBudgetDerivedFromNamedInputs (0.00s)
```

389 개가 실행됐고(셀렉터 0매치 아님), 붉은 것은 **정확히 하나**이며 그것이 새 가드다.
기존 가드는 **같은 실행에서 초록으로 남았다** — 이것이 RED 의 귀속 근거다.
census 의 나머지 두 파일도 발화하지 않았다.

실패 메시지 전문:

```
board_lock_wait_test.go:103: constant coherence broken: the lock policy budgets 10 supported
writers x 4 headroom = 40 serialized mutations, while the stress test serializes 8 x 6 = 48
(1.32s budget < 1.584s floor). Lowering either policy constant, or raising either stress
constant past that product, fails this guard. The per-mutation cost cancels on both sides, so
the relation is cost-independent and asserts nothing about the wait any real machine needs —
the CI -race per-mutation cost observed by t370 was 42-105ms against the declared 33ms.
```

## 4) 복원

```
$ sed -i '' 's/boardLockHeadroom = 4/boardLockHeadroom = 5/' internal/kanban/board_store.go
$ git status --short internal/
(빈 출력 — 추적 트리 원상복구)
$ go test -race -count=1 -run 'TestBoardLockWaitBudget' -v ./internal/kanban/
--- PASS: TestBoardLockWaitBudgetDerivedFromNamedInputs (0.00s)
--- PASS: TestBoardLockWaitBudgetCoversSerializedMutations (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/kanban	1.359s
```

## 5) 뮤턴트 이전 기준 실행

```
$ go test -race -count=1 -v -run 'TestConcurrencyStress|TestBoardLockWaitBudget|TestStress' ./internal/kanban/
=== RUN 5 개, 전부 PASS
backlog_concurrency_test.go:235: ... 48 attempts; 48 succeeded, 0 starved (tolerated),
0 hard failures; 48 distinct ids, 48 stored items, last_seq 48;
back-derived per-mutation cost 14.801705ms (elapsed 710.481875ms / 48 successful mutations)
```

## 6) 이 관측이 세우지 못하는 것

- **로컬 초록은 수리의 근거가 아니다.** 실측 cost 14.8ms 는 임계 34.4ms 아래로,
  t370 의 로컬 darwin 17.5ms 와 같은 대역이다 — 이 머신은 원래 통과하던 쪽이다.
  CI 가 붉었던 42~105ms 대역은 재현되지 않았고, 재현을 시도하지도 않았다(부하 생성 금지).
- **AC-SIV-009(불변식 방향)는 여기서 재현하지 않았다.** 그 증거는 manager-develop 의
  run-phase 기록(`run-evidence.md`)이며, 이 문서의 관측 범위 밖이다.
- CI 는 읽지도 촉발하지도 않았다.
