# t370 — 착수 전 t354 중복 판독 (조사 미착수)

읽은 트리: primary checkout `/Users/goos/MoAI/moai-adk-go`, HEAD `48239c7dc` (branch `main`).
`git fetch origin develop` 후 `origin/develop` = `1e5199b88`.

## Claim

t370 이 규명하라고 한 것 중 **1(최초 회귀 지점)·2(실패 테스트 동일성)·3(실패 성격)·5(수리 범위)** 는
t354 가 이미 규명했고 **수리까지 develop 에 착지**했다. 다만 그 수리는 **t370 이 인용한 다섯 head
전부보다 앞서 있고, 그 head 들이 여전히 붉다** — 즉 t354 의 수리가 실패를 닫지 못했다.

## Evidence

t354 는 `picked` 이고 워크트리가 살아 있다:

```
$ git worktree list | grep t354
.claude/worktrees/t354   21ee55c5a [WT-concurrency-stress]
$ git -C .claude/worktrees/t354 log --oneline -2
21ee55c5a docs(SPEC-BACKLOG-LOCK-BUDGET-001): sync-phase artifacts (t354)
a680ea6e8 fix(SPEC-BACKLOG-LOCK-BUDGET-001): break the queue lock retry lockstep and derive the wait budget (t354)
```

t354 산출물 3건(`verdict.md` · `measurements.md` · `ac-blb-006-lead-verdict.md`)이 primary 에 있고,
`verdict.md` 는 착지를 이렇게 적는다: SPEC-BACKLOG-LOCK-BUDGET-001, 병합 `728f91006`
(`77b2bcae6..728f91006`), 푸시 후 `0	0`.

t354 가 규명한 내용(재도출 불필요):

- 실패 테스트: `TestConcurrencyStress`, `internal/kanban/backlog_concurrency_test.go:56`,
  `2/48 adds failed under contention` / `kanban board lock held`
- 도입 커밋: `83a1d492a` (t306 M2) — lane-9 확정
- 성격: **data race 아님**. 불변식(lost update·id 충돌)은 유지됐고 **락 대기가 포기**한 것.
  `unix.Flock(LOCK_EX|LOCK_NB)` + 폴링이라 공정성이 없어 **기아(starvation)** 가 발생.
  CI 는 mutation 당 ~33 ms > 재시도 지연 25 ms, 로컬 darwin 은 ~14 ms < 25 ms — CI 전용인 이유.
- 예산: `boardLockRetryDelay=25ms`, `boardLockRetries=40` → 41 sleep ≈ 1.025 s
  (`internal/kanban/board_store.go:76-79`)
- t354 자신이 남긴 Gap: "로컬 증거는 무회귀+마진 개선만 세우고, **CI 실패가 닫혔음은 세우지 못한다**"

## Baseline-attribution — 이번 실행에서 잰 것

t370 이 인용한 다섯 head 각각에 대해 t354 수리 병합의 조상 관계를 측정:

```
$ for h in 1728136c7 d8a1a8e4e 1e5199b88 52c3fe590 d7010f86a; do
    git merge-base --is-ancestor 728f91006 "$h" && echo "$h AFTER" || echo "$h BEFORE"; done
1728136c7 AFTER    d8a1a8e4e AFTER    1e5199b88 AFTER    52c3fe590 AFTER    d7010f86a AFTER
$ git merge-base --is-ancestor 728f91006 origin/develop → yes
```

**다섯 head 전부가 수리를 포함한 채 실패했다.** 따라서 이 붉음은 스테일 재판독이 아니다.

## Gaps (관측하지 않은 것)

- 다섯 head 의 CI 로그를 **직접 판독하지 않았다** — 실패 테스트 동일성은 리드의 `--log-failed`
  관측(3 head)을 인용한 것이고, `52c3fe590`·`d7010f86a` 두 head 는 리드도 미판독이다
- 한 head 에서 `TestConcurrencyStress` 외 다른 테스트가 함께 실패했는지 미확인 (t370 질문 2)
- `WARNING: DATA RACE` 유무를 이번 실행에서 직접 보지 않았다 (t354 서술 인용)
- `21ee55c5a`(t354 sync 커밋)가 origin/develop 에 있는지 미확인 — 병합 `728f91006` 만 확인

## Residual-risk

t354 가 `picked` 상태로 살아 있어, 그 레인이 지금도 같은 실패를 다루고 있을 수 있다.
t370 을 그대로 착수하면 조사가 중복되거나 두 카드가 같은 파일을 만진다.
