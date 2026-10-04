# t1463 card-review (codex_review, scope=card, advisory)

## Round 1 — HEAD cc351b5a1, base 30ce3a02d, backend codex
Verdict: fail. One finding.

- [P1] internal/harness/retention_fifo_unix_test.go:39-42 — On Linux the pre-check passes, then `openStateFile` opens the FIFO `O_RDWR` (does not block) and `pruneLocked`'s `readStamp` waits forever; the test's recovery open does not release the read, so `<-done` waits unbounded and the test fails Ubuntu CI. Reviewer reproduced with a Linux/arm64 test binary: `panic: test timed out after 12s`, exit 2 (confidence 1.00). Asked: reject the FIFO in the lock path too, and bound the test's recovery wait.

Disposition: FIXED in the next commit.
- `retention.go` `openStateFile`: a named pipe, socket or device entry returns `prune state entry ... is not a regular file; prune skipped` before any open.
- `retention_fifo_unix_test.go`: the unbounded `<-done` wait and the recovery open were removed; a timeout now fails the test immediately.

## Round 2
See the section below once recorded.
