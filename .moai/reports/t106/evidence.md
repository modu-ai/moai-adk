# t106 — todo 큐 경로 해석 근본 수정: 증거

Date: 2026-08-17 · Card: t106 (lead dispatch + 3-item rework order) · Branch: WT-t106 · Base: 5c3141372
Commits: `6ba8ef90e` (initial fix) + rework commit (fallback rename / adopt-not-shadow / stale-lock sweep)

## Card

> todo 큐 경로 해석 근본 수정: internal/cli/todo.go:30이 root(워크트리 자신) 기준으로
> backlog.json을 잡아 분열. (a)본안 git rev-parse --git-common-dir 부모=primary 체크아웃을
> root로(internal/hook/branch_guard.go primary 판별 선례 재사용, 파일 이동 없음) +
> (b)폴백 primary 미발견 시 ~/.moai/todo/<project-key>/
> [HARD] 검증 2종: ① 카드 워크트리에서 moai todo list가 primary와 동일 항목 수
> ② 워크트리에서 moai todo add한 항목이 primary 조회에 출현
> + 리뷰 전 재작업 3건 (운영자 긴급 정정, lead msg): 폴백 명명 ~/.moai/todo/ ·
> 무손실 채택(adopt-not-shadow) + [HARD] 검증 3 · 잔재 락 처리 + 세션 레코드 미접촉 명시

## Claim

`moai todo` resolves the backlog queue against the PRIMARY checkout of the
repository from any linked worktree; a launch context git cannot answer falls
back to `~/.moai/todo/<project-key>/` (named for the owning command — NOT
"kanban", which would falsely read as moving internal/kanban's session
records). The fallback's first run ADOPTS an existing project-local queue
(same items, same states) instead of shadowing it behind an empty queue. A
superseded legacy lock artifact (`backlog.json.lock`) is swept on store
construction. The delegation-channel invariant ("one queue per repository")
holds from inside a card worktree.

## Defect reproduction (before the fix)

Measured on this card's worktree, pre-fix binary (installed v3.1.0,
`~/go/bin/moai`), CLAUDE_PROJECT_DIR = the worktree:

```
$ ls .moai/state/kanban/backlog.json   # in the worktree
MISSING in t106 tree

$ ~/go/bin/moai todo list              # from the worktree
queue is empty                          # ← the split: primary held 30 cards
```

Root cause: `resolveProjectDir()` returns `$CLAUDE_PROJECT_DIR` or CWD, and
`todoBacklogPath` joined `.moai/state/kanban/backlog.json` under it — so a
worktree context addressed a worktree-local queue file the primary never
sees. `.moai/state` is untracked, so a fresh worktree reads as an empty queue
while an `add` from the worktree creates a second, orphaned queue.

## Change

- `internal/cli/todo.go`
  - `resolveTodoQueueRoot()` — resolves the launch context through
    `gitcore.ResolveGitDirs` (internal/core/git; the same discrimination
    branch_guard.go applies — no file moved) and returns
    `filepath.Dir(CommonDir)` = the primary checkout root, identical from any
    worktree and from the primary itself. Fail-open to the fallback when git
    cannot answer.
  - `fallbackTodoQueueRoot()` — `~/.moai/todo/<project-key>/` via the
    package's existing `userHomeDirFn` seam. Named for `moai todo` (no
    `moai kanban` command exists); a "kanban" name would read as moving
    internal/kanban's per-session records — a scope this queue never touches.
    Unresolvable home degrades to the project-local path rather than erroring.
  - `adoptLocalTodoQueue()` — adopt-not-shadow: before the fallback's queue
    is created, a pre-existing project-local `backlog.json` is carried over —
    atomic same-volume rename, cross-volume copy that KEEPS the original
    (deletion is the one outcome the lossless requirement forbids; a leftover
    original is inert because the populated fallback wins on later runs).
    Best-effort: a failure leaves the local queue untouched so a later run
    can adopt again.
  - `todoQueueProjectKey()` — `<base>-<sha256(abs)[:4] hex]>` (8 hex chars).
  - `newTodoStore()` — single constructor; the five verb call sites share one
    root resolution per invocation. Help text states the residence rule.
- `internal/kanban/backlog_store.go` — `NewBacklogStore` sweeps the
  superseded `backlog.json.lock` artifact (0 B, dated three days older than
  the live `backlog.lock` on the primary) best-effort; the live lock name is
  untouched.
- `internal/cli/todo_test.go` — `todoFixture` git-inits its temp dir (a real
  primary checkout). Existing test bodies untouched.
- `internal/cli/todo_queue_root_test.go` (new) — worktree→primary convergence,
  primary self, subdirectory→repo-root, no-git fallback (+key shape,
  uniqueness), the [HARD] ①② pair in code form, and the adoption test
  ([HARD] ③: 2 queued + 1 picked + last_seq survive the cutover, and a
  re-run neither duplicates nor re-adopts).
- `internal/kanban/backlog_store_test.go` — legacy-lock sweep test (legacy
  removed, live lock untouched).
- `.claude/skills/moai/workflows/todo.md` + template mirror — residence rule
  and the adopt-not-shadow first-run behavior documented (mirror-identical;
  `make build` re-ran).

**Scope guard — session records untouched**: this change never reads, moves,
or writes internal/kanban's per-session records (`.moai/state/kanban/<uuid>.json`,
35 files on the primary, owned by `internal/kanban/record.go`). The diff
touches only `backlog.json` resolution/adoption and the stale
`backlog.json.lock` artifact. Verified by file list (6 + 1 files above); no
`<uuid>.json` path appears anywhere in the diff.

## Evidence (verbatim commands + outputs)

Fix verification ran with the rebuilt binary (`./bin/moai`, this branch) from
INSIDE the t106 worktree (session CLAUDE_PROJECT_DIR = the worktree):

[HARD] ① — same item count from the worktree as the primary. Measured twice
(the queue moved underneath us — the lead processed a card — and the worktree
view tracked it exactly):

```
$ ./bin/moai todo list | wc -l
      35
$ python3 (direct read of primary backlog.json)
primary items: 35

  ... lead processed one card ...

$ ./bin/moai todo list | wc -l          # after the rework changes
      34
$ python3 (direct read of primary backlog.json)
primary items: 34
t106 state: picked
```

[HARD] ② — an add from the worktree lands in the primary queue (measured on
the first fix revision; unchanged by the rework, which touches only the
no-git fallback path):

```
$ ./bin/moai todo add "t106 검증용 임시 카드 — 실측 후 즉시 제거"
t112 21
$ python3 (primary backlog.json)
found in PRIMARY: True | state: queued
$ ./bin/moai todo done t112
done t112
$ python3 (primary backlog.json)
t112 remaining in PRIMARY: False | total items: 35
```

[HARD] ③ — fallback first run adopts the existing local queue (same item
count, same states). The manual /tmp reproduction could not be executed from
this session: the worktree-isolation guard (correctly) refuses Bash commands
targeting paths outside the worktree, and routing around it via a script file
is itself a named violation — so the proof is the dedicated test, which runs
the REAL resolution path (`resolveTodoQueueRoot` → rename → `Load`) against a
real seeded file (2 queued + 1 picked, last_seq 7, spec_id attached), plus a
re-run proving no duplicate adoption:

```
$ go test ./internal/cli/ -run 'TestTodoQueue_FallbackAdoptsExistingLocalQueue' -count=1 -v
=== RUN   TestTodoQueue_FallbackAdoptsExistingLocalQueue
--- PASS: TestTodoQueue_FallbackAdoptsExistingLocalQueue (0.15s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	6.035s
```

Stale-lock sweep — measured live on the primary checkout. Before any new-binary
run the artifact was present; one `todo list` from the worktree (whose root
resolution addresses the PRIMARY queue — the store construction sweeps beside
it) removed it, leaving the live lock:

```
$ ls /Users/goos/MoAI/moai-adk-go/.moai/state/kanban/backlog.json.lock
-rw-r--r--  1 goos  staff  0 Aug 14 20:44 ... backlog.json.lock
STALE LOCK STILL PRESENT

$ ./bin/moai todo list | wc -l
      34
STALE LOCK SWEPT
```

Test suite (this branch, serial, lane-local per the load discipline):

```
$ go build ./internal/cli/ ./internal/kanban/ && go vet ./internal/cli/ ./internal/kanban/
BUILD+VET OK

$ go test ./internal/kanban/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/kanban	19.592s

$ go test ./internal/cli/ -run 'Todo' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	8.692s

$ go test ./internal/cli/ -count=1          # first fix revision
ok  	github.com/modu-ai/moai-adk/internal/cli	277.844s

$ golangci-lint run ./internal/cli/ ./internal/kanban/
0 issues.
```

## Baseline attribution

All measurements above were produced on this branch's tree (WT-t106 @ the
rework commit) against the live primary checkout's queue file at
/Users/goos/MoAI/moai-adk-go/.moai/state/kanban/backlog.json, same morning.
The pre-fix reproduction used the installed v3.1.0 binary on the same
machine.

## Gaps

- `internal/hook/session_start_kanban.go:202` reads the queue for the lead's
  bootstrap notice under `input.ProjectDir` — the same split class when a
  worktree session starts and Claude Code reports the worktree as ProjectDir.
  Not measured (what ProjectDir a worktree session actually receives is
  unverified); recommend a follow-up card rather than widening this one.
- The [HARD] ③ proof is the test run, not a manual out-of-tree reproduction —
  the session's worktree guard forbids out-of-tree Bash (noted above with the
  reason). The test exercises the identical production code path.
- Bare-repository launch contexts are not handled (Dir(CommonDir) would be
  the bare repo's parent); treated as unsupported, no test.
- Full-suite / cross-platform verdict deferred to CI per the lane-local
  verification discipline (darwin local only).

## Residual risk

- A session whose `CLAUDE_PROJECT_DIR` points at a DIFFERENT repository than
  the CWD resolves that repository's primary — the session context wins by
  design (pre-existing precedence), but operators should know the queue
  follows the session's project, not the shell's directory.
- Inside a git submodule, `CommonDir` belongs to the submodule, so the queue
  roots at the submodule's checkout — reasonable but unmeasured.
- Adoption uses rename on the same volume; on cross-volume setups the copied
  original stays behind as an inert downgrade-era snapshot (deliberate — no
  deletion). A later downgrade to pre-fallback binaries would read that
  stale local copy, not the live fallback queue.
