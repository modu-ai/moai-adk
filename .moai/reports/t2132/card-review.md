# t2132 card review (self-review)

Self-review only; no external review was run (Codex not approved for this card, audit tools not used).

Checked: every change is in a `_test.go` file and only adds a `runtime.GOOS != "windows"` condition to a mode-bit comparison (or skips/returns on the read-only-directory step). No assertion was removed, loosened, or altered on POSIX; all other assertions in each test still run on Windows. One production-code mode comparison (`internal/graph/gtd_private.go:117`) is recorded as a finding in progress.md Gaps and left untouched.
