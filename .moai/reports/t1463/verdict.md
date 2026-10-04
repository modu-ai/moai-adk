# t1463 — FIFO at harness retention state path hangs the prune pre-check

Branch: WT-harness-fifo-hang (base develop 30ce3a02d). Lane: lane-22. Class B (no SPEC).

## Claim
A FIFO at `<log>.prune-state` no longer blocks `PruneStaleEntries`: the lock-free stamp read opens with `O_NONBLOCK` on unix.

## Evidence
- RED on base (test added, fix absent):
  `go test ./internal/harness/ -run '^TestPruneStateFIFODoesNotHang$' -count=1 -v -timeout 60s`
  -> `retention_fifo_unix_test.go:43: PruneStaleEntries blocked on a FIFO at the state path`, `--- FAIL (3.00s)`, exit 1.
- GREEN after fix, same command -> `--- PASS: TestPruneStateFIFODoesNotHang (0.00s)`, `ok`.
- `go test ./internal/harness/ -count=1 -timeout 600s` -> `ok ... 3.019s`.
- `go test ./internal/harness/ -run 'Prune|Retention' -race -count=5` -> `ok ... 40.546s`.
- `golangci-lint run ./internal/harness/...` -> `0 issues.`
- `GOOS=windows go vet ./internal/harness/` -> empty, exit 0.
- Origin of the premise: `.moai/worktrees/t1432/.moai/reports/t1432/plan-audit-iter2.md` B3 (probe P-FIFO, `retention.go:94`, `:187-189`).

## Change
- `internal/harness/retention.go`: `readStampFile` calls `openStampReadOnly`.
- `internal/harness/retention_open_unix.go` (new): `os.OpenFile(path, O_RDONLY|syscall.O_NONBLOCK, 0)`.
- `internal/harness/retention_open_windows.go` (new): `os.Open`.
- `internal/harness/retention_fifo_unix_test.go` (new): reproduction test.

## Baseline-attribution
All commands run in this run against worktree HEAD 30ce3a02d plus the uncommitted/committed change above. Tool: `golangci-lint` from PATH; its build/version was NOT checked against CI's v2.1.6.

## Gaps
- `golangci-lint` version not verified as v2.1.6 (CI version).
- Windows only vetted, not executed.
- Linux not run; behaviour of `O_NONBLOCK` read on a writer-less FIFO measured on darwin only.
- Only the pre-check read was repaired. The locked path (`openStateFile` default branch, `O_RDWR|O_CREATE` on a FIFO) was not probed for a hang beyond the test completing within 3 s.
- Independent card-review not run (lane self-review only).

## REQ-HRH-008 judgment (one line)
REQ-HRH-008 (t1432 spec.md:196) bars adding "no lock, no file open and no system call beyond those present at base"; `O_NONBLOCK` changes the flags of the one existing open and adds none, so the wording does not forbid it, but AC-009's function-scoped diff over `readStampFile` will now show a change and t1432 spec.md:64 calls the FIFO hang "not repaired here (REQ-HRH-008 forbids a call added)" — that sentence needs a follow-up edit when t1432 lands (read, not run against t1432's tree).

## Residual-risk
- SPEC-HARNESS-RETENTION-HARDEN-001 (t1432) REQ-HRH-008 forbids adding a system call to the pre-lock path; this change adds a flag, not a call, but t1432's wording/AC-009 should be reconciled when that SPEC is revisited.
- A FIFO with an active writer would deliver its bytes to the stamp read (bounded by `maxStampBytes`).
