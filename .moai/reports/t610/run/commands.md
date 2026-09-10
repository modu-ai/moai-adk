# t610 run-phase (M1–M3) — exact commands

Tree: worktree `.claude/worktrees/t610`, branch `WT-go-1266`.
Measured on the M1 commit: HEAD `41f445fa54f3a4949491bbe9fe06c68a97cf283e`,
tree `e6ec2991e4ce028b83d2be0658844ddb3813a8dd` (go.mod:3 `go 1.26.8`).
Shell environment: no `GOTOOLCHAIN` / `GOFLAGS` set (`env | grep -E '^(GOTOOLCHAIN|GOFLAGS|...)='` printed nothing).
Run by manager-develop on 2026-09-10. cwd = worktree root for every command.

Exit-code form for every judged command: `<cmd> > <file> 2>&1` (or `>>`), then
`echo "<name>_exit=$?" > <file>.exit` in the same shell, so `$?` is the command's own
exit status, never a pipeline's.

## Files that carry their command inline

Each of these starts with a header naming the command, HEAD, and tree id:

| File | Check |
|------|-------|
| `ac001-toolchain.txt` (+ `.exit`) | AC-GTS2-001 `go -C <root> version`, `go -C <root> env GOTOOLCHAIN GOMOD` |
| `ac002-directive.txt` (+ `.exit`) | AC-GTS2-002 `sed -n 3p go.mod`, `grep -c '^toolchain' go.mod`, control `grep -c '^go ' go.mod` |
| `ac004-govulncheck-env.txt` (+ `.exit`) | AC-GTS2-004 env record for the judged scan |
| `ac004-ids.txt` (+ `.exit`) | AC-GTS2-004 per-ID `grep -c` on the scan log, plus the two text controls |
| `ac004-ids-control.txt` (+ `.exit`) | instrument control: the same grep form on the go1.26.4 baseline log |
| `ac005-make-build.log` (+ `.exit`) | AC-GTS2-005 `make build` |
| `ac005-go-version-m.txt` (+ `.exit`) | AC-GTS2-005 `go version -m bin/moai` |
| `ac006-go-vet.log` (+ `.exit`) | AC-GTS2-006 `go vet ./...` (wrapped in `timeout 580`) |
| `ac006-vet-pkgset.txt` (+ `.exit`, `.list`) | vet swept-set control: `go list ./...` count |
| `ac006-go-test.log` (+ `.exit`) | AC-GTS2-006 `go test -count=1 -timeout 600s ./internal/web/... ./internal/update/... ./internal/goal/... ./pkg/...` (wrapped in `timeout 590`) |
| `ac006-swept.txt` (+ `.exit`) | swept-set counts, selected package list, before/after status `cmp` |

## Raw files (git commands; header not inline)

The worktree guard refuses a git command combined with other shell statements, so each
git capture is a single git command redirected to a `.raw` file, with its exit in a
sibling `.exit`. The command for each:

| File | Command |
|------|---------|
| `ac003-numstat.raw` | `git diff --numstat develop...HEAD -- go.mod` |
| `ac003-diff.raw` | `git diff develop...HEAD -- go.mod` |
| `ac007-complement-m1.raw` | `git diff --name-only develop...HEAD -- . ':!go.mod' ':!.moai/specs/SPEC-GO-TOOLCHAIN-SEC-002' ':!.moai/reports/t610'` (on the M1 commit) |
| `ac005-status-before-build.raw` | `git status --porcelain --untracked-files=no` (before `make build`) |
| `ac005-status-after-build.raw` | `git status --porcelain --untracked-files=no` (after `make build`) |
| `ac004-govulncheck.log` | `GOMAXPROCS=2 timeout 590 govulncheck ./...` (header is in `ac004-govulncheck-env.txt`) |

The AC-007 re-check after the evidence commit is recorded as `ac007-complement-m3.raw`
with the same command, taken on the evidence-commit HEAD.

## Deviations and instrument corrections

- govulncheck wrapper `timeout 590` instead of the `timeout 900` in the prompt: the Bash
  tool caps a foreground call at 600 s. The scan exited 0 inside that bound, so the lower
  ceiling did not cut it short.
- `ac002-directive.txt` was captured twice. The first capture passed `--` to BSD `sed`,
  which took it as a filename (`sed: --: No such file or directory`, `sed_exit=1`). The
  file was overwritten by a capture without `--`; the recorded file is the second one.
- One earlier AC-002 attempt using a shell variable in the `sed` argument was refused by
  the worktree guard before running and wrote nothing.
- `internal/cli` tests were not run (lead slot rule). `go test ./...` was not run.
