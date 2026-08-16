# t106 — todo 큐 경로 해석 근본 수정: 증거

Date: 2026-08-17 · Card: t106 (lead dispatch) · Branch: WT-t106 · Base: 5c3141372

## Card

> todo 큐 경로 해석 근본 수정: internal/cli/todo.go:30이 root(워크트리 자신) 기준으로
> backlog.json을 잡아 분열. (a)본안 git rev-parse --git-common-dir 부모=primary 체크아웃을
> root로(internal/hook/branch_guard.go primary 판별 선례 재사용, 파일 이동 없음) +
> (b)폴백 primary 미발견 시 ~/.moai/kanban/<project-key>/
> [HARD] 검증 2종: ① 카드 워크트리에서 moai todo list가 primary와 동일 항목 수
> ② 워크트리에서 moai todo add한 항목이 primary 조회에 출현

## Claim

`moai todo` now resolves the backlog queue against the PRIMARY checkout of the
repository from any linked worktree, and against a home-based fallback
(`~/.moai/kanban/<project-key>/`) when git cannot resolve a primary. The
delegation-channel invariant ("the queue is one per repository") holds from
inside a card worktree.

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
worktree context addressed a worktree-local queue file that the primary (and
the lead reading it) never sees. `.moai/state` is untracked, so a fresh
worktree has no file and reads as an empty queue while an `add` from the
worktree creates a second, orphaned queue.

## Change

- `internal/cli/todo.go`
  - `resolveTodoQueueRoot()` — resolves the launch context through
    `gitcore.ResolveGitDirs` (internal/core/git, the same resolution
    branch_guard.go discriminates with; no file moved) and returns
    `filepath.Dir(CommonDir)` = the primary checkout root. Fail-open to
    `fallbackTodoQueueRoot()` when git cannot answer.
  - `fallbackTodoQueueRoot()` — `~/.moai/kanban/<project-key>/` via the
    package's existing `userHomeDirFn` seam; unresolvable home degrades to the
    project-local path rather than erroring.
  - `todoQueueProjectKey()` — `<base>-<sha256(abs)[:4] hex]>` (8 hex chars):
    readable and collision-safe.
  - `newTodoStore()` — single constructor; the five verb call sites now share
    one root resolution per invocation.
  - Command help text states the residence rule.
- `internal/cli/todo_test.go` — `todoFixture` now `git init`s its temp dir so
  the fixture is a real primary checkout (queue-root resolution goes through
    git now). Existing test bodies untouched.
- `internal/cli/todo_queue_root_test.go` (new) — worktree→primary convergence,
  primary self, subdirectory→repo-root, no-git fallback (+key shape, key
  uniqueness), and the [HARD] acceptance pair in code form.
- `.claude/skills/moai/workflows/todo.md` +
  `internal/template/templates/.../todo.md` — queue-residence rule documented
  (mirror-identical; `make build` re-ran, catalog.yaml hashes updated).

## Evidence (verbatim commands + outputs)

Fix verification ran with the rebuilt binary (`./bin/moai`, this branch) from
INSIDE the t106 worktree (session CLAUDE_PROJECT_DIR = the worktree):

[HARD] ① — same item count from the worktree as the primary:

```
$ ./bin/moai todo list | wc -l
      35

$ python3 (direct read of /Users/goos/MoAI/moai-adk-go/.moai/state/kanban/backlog.json)
primary items: 35
last 3 ids: ['t109', 't110', 't111']
```

35 == 35 — and the queue had grown from 30 to 35 since the defect
reproduction (the lead registered t105..t111), which the worktree view tracks:
the split is gone.

[HARD] ② — an add from the worktree lands in the primary queue:

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

Test suite (this branch, serial, lane-local per the load discipline):

```
$ go test ./internal/cli/ -run 'Todo' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	5.274s

$ go test ./internal/cli/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	277.844s

$ go build ./internal/cli/ && go vet ./internal/cli/
(exit 0)

$ golangci-lint run ./internal/cli/
0 issues. (exit 0)
```

## Baseline attribution

All measurements above were produced on this branch's tree (WT-t106 @ the
commit staged with this evidence) against the live primary checkout's queue
file at /Users/goos/MoAI/moai-adk-go/.moai/state/kanban/backlog.json. The
pre-fix reproduction used the installed v3.1.0 binary on the same machine the
same morning.

## Gaps

- `internal/hook/session_start_kanban.go:202` reads the queue for the lead's
  bootstrap notice under `input.ProjectDir` — the same split class when a
  worktree session starts and Claude Code reports the worktree as ProjectDir.
  Not measured (what ProjectDir a worktree session actually receives is
  unverified); recommend a follow-up card rather than widening this one.
- Non-git project migration: an existing queue at
  `<project>/.moai/state/kanban/backlog.json` in a project WITHOUT git
  metadata is not adopted into the home fallback — after this change such a
  project's CLI reads the (new) home queue while the old file stays behind.
  Git projects (the overwhelmingly common case, and the case the card's [HARD]
  checks exercise) are unaffected: the file does not move.
- Bare-repository launch contexts are not handled (Dir(CommonDir) would be the
  bare repo's parent); treated as unsupported, no test.
- Full-suite / cross-platform verdict deferred to CI per the lane-local
  verification discipline (darwin local only).

## Residual risk

- A session whose `CLAUDE_PROJECT_DIR` points at a DIFFERENT repository than
  the CWD resolves that repository's primary — the session context wins by
  design (pre-existing precedence), but operators should know the queue
  follows the session's project, not the shell's directory.
- Inside a git submodule, `CommonDir` belongs to the submodule, so the queue
  roots at the submodule's checkout — reasonable but unmeasured.
