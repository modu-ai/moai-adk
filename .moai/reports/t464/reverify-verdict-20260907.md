# t464 재검증 판정 — 카드는 이미 착지했다

**카드**: t464 · **레인**: lane-6 · **워크트리**: `.claude/worktrees/t464` (`WT-worktree-flag-race`)
**재검증 트리**: `615d18c1f` (= `origin/develop`, 자체 커밋 0) · **일자**: 2026-09-07

---

## Claim (주장)

1. **t464 는 이미 완결·착지한 카드다.** 수리 커밋 `76f2165a1` 이 `origin/develop` 의 조상이고,
   `SPEC-CLI-WORKTREE-FLAG-RACE-001` 의 frontmatter 는 `status: completed` 다.
2. **배차문의 재현 명령은 현재 트리에서 RED 를 세우지 못한다** — 결함이 이미 수리됐기 때문이다.
3. **착지한 수리는 하중을 받고 있다** — 되돌리면 DATA RACE 가 되살아난다(뮤턴트 실증).
4. **배차문이 서술한 기전은 현재 트리에 대해 거짓이다** — 네 형제는 `t.Parallel()` 을 부르지 않는다.
   그 서술은 수리 **이전** 트리(lane-11 이 잰 `7594698d2` 이전)에 대해서만 참이었다.
5. **t464 가 넘긴 이월 부채는 그 사이 해소됐다** — `TestBinaryLag_DoctorCheckNameSetIsUnchanged` 통과.

## Evidence (증거)

**착지 판정**
```
$ git merge-base --is-ancestor 76f2165a1 origin/develop && echo YES
YES
$ git log -1 --format='%H %ad' 76f2165a1
76f2165a13e1a1742a9476a30410286d18020d59 Fri Sep 4 04:10:42 2026 +0900
$ grep -m3 -E '^(id|status|updated):' .moai/specs/SPEC-CLI-WORKTREE-FLAG-RACE-001/spec.md
id: SPEC-CLI-WORKTREE-FLAG-RACE-001
status: completed
updated: 2026-09-04
```
`git log --oneline 7594698d2..HEAD -- internal/cli/worktree_branch_flag_test.go`
→ `76f2165a1 fix(card t464): remove t.Parallel() from the four TestResolveWorktreeExistingBranch siblings`
즉 **lane-11 이 잰 `7594698d2` 는 수리 커밋보다 앞선다.** 그 측정은 그 시점에 옳았고, 지금은 낡았다.

**4회 측정 — GREEN → 뮤턴트 RED → GREEN**

| # | 트리 상태 | 명령 | rc | `WARNING: DATA RACE` |
|---|---|---|---|---|
| 1-3 | 원본(수리 착지분) | `go test ./internal/cli/ -run TestResolveWorktreeExistingBranch -count=20 -race` | **0** ×3회 | **0** ×3회 |
| 4 | **뮤턴트**(수리 역적용) | 동일 | **1** | **24** |
| 5 | 복원 | 동일 | **0** | **0** |

뮤턴트 적용 방법(수리를 정확히 역적용, 손 편집 아님):
```
$ git show 76f2165a1 -- internal/cli/worktree_branch_flag_test.go | git apply -R -
$ grep -n 't.Parallel()' internal/cli/worktree_branch_flag_test.go
25: 44: 66: 96: 126: 161:
```
66/96/126/161 — 수리 커밋 메시지가 스스로 적은 네 줄과 정확히 일치한다.
뮤턴트 레이스 좌표에 `worktree_branch_flag_test.go:162` 읽기가 포함돼, **배차문이 인용한 lane-11 좌표
(`:130` 쓰기 ↔ `:162` 읽기)를 재현한다** — 그 좌표가 수리 전 트리의 것이었음을 확인해 준다.
복원 후 `git status --porcelain <file>` 빈 출력(원본과 바이트 동일).

**배차문 기전의 현재 트리 반증**
```
$ grep -n 't.Parallel()' internal/cli/worktree_branch_flag_test.go
25:	t.Parallel()      ← TestSplitWorktreeBranchFlag (전역 미접촉)
44:			t.Parallel()  ← 그 서브테스트 (전역 미접촉)
```
`TestResolveWorktreeExistingBranch_*` 네 형제에는 `t.Parallel()` 이 **없다**.

**이월 부채 재측정** (t464 §E.2 가 t466 소관으로 넘긴 건)
```
$ go test ./internal/cli/ -run TestBinaryLag_DoctorCheckNameSetIsUnchanged -count=1 -v
=== RUN   TestBinaryLag_DoctorCheckNameSetIsUnchanged
--- PASS: TestBinaryLag_DoctorCheckNameSetIsUnchanged (0.05s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.743s
```
`=== RUN` 행 수 **1** — 셀렉터 0매치로 인한 공허한 초록이 아님을 명시적으로 확인했다.

## Baseline-attribution (baseline 귀속)

위 전부 이 워크트리에서, 이번 실행에, `615d18c1f` 트리에 대해 측정했다.
`615d18c1f` = `origin/develop` (`git rev-list --count --left-right origin/develop...HEAD` → `0 0`).
lane-11 / lane-12 의 수치는 **재사용하지 않았다** — 인용한 것은 그들의 좌표뿐이고,
그 좌표는 내 뮤턴트 실행에서 독립적으로 재현됐다.

## Gaps (미검증)

- **패키지 전수 `go test ./internal/cli/... -race` 를 재실행하지 않았다.** 카드가 이미 닫혀 있어
  착지한 작업의 재수행에 해당하고, 직전 레인이 1800s 실행으로 이미 남겼다
  (`package-race-1800s-76f2165a1.txt`: 레이스 0, 16/16 서브패키지 ok). 따라서
  **현재 develop 트리의 패키지 전수 건강도는 내가 관측하지 않았다.**
- 이월 부채를 누가·어느 커밋으로 닫았는지 귀속하지 않았다 — 통과 사실만 측정했다.
- linux/amd64 에서의 레이스 거동은 미측정(원 카드 §F2 와 동일하게 남는다).
- CI 판정은 읽지 않았다 — 리드 소관.

## Residual-risk (잔여 위험)

- `-count=20` 은 확률적이다. 원본 GREEN 3회·복원 GREEN 1회, 총 4회 통과가 근거이고,
  이것이 "레이스 부재"의 증명은 아니다. 다만 **뮤턴트가 24건을 즉시 되살렸다**는 사실이
  이 명령의 검출력 자체는 살아 있음을 보인다 — 공허한 초록이 아니다.
- 수리가 Option A(구조적 주입이 아닌 `t.Parallel()` 제거)라, 나중에 누군가 `t.Parallel()` 을
  다시 넣으면 결함이 재발한다. 원 카드가 §E.2 에서 이미 인지하고 상시 회귀 명령으로 방어한 항목이다.

---

## 레인의 처분 요청

**나는 이 카드를 철회하지 않는다** — 배차 전 교차확인은 보고하되 거부권은 없다는 규율
(`kanban-dispatch.md` § 교차확인은 보고하지 거부하지 않는다)에 따라, 리드에게 올리고
**운영자의 확인 또는 철회**를 기다린다.

권고: **철회.** 수리·SPEC·증거가 전부 develop 에 있고, 이 트리에서 재현이 서지 않는다.
