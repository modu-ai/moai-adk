# manual-check.md — AC-MT-015 (operator-held, NOT run in CI)

Status: **operator confirmation pending, no terminal designated.** Nothing below has been run. This file is the fill-in form for the procedure of `acceptance.md` §4; the recorded result goes to `.moai/reports/t1408/manual-attach-check.md` (local report, gitignored).

Preconditions: a real terminal, an authenticated codex-cli 0.160.0, a factory run with the stamps, `MOAI_FACTORY_MANAGED=1`, launch `moai codex`. Record the codex version in the result file.

| Step | What to do | Premise it observes | Observed result (fill in) |
|---|---|---|---|
| 1 | The TUI appears after the priming turn; the launcher printed the log-file path first. | attach order, REQ-MT-006 | not run |
| 2 | Type a message: the model answers in the TUI. | P9 (resume by thread id works) | not run |
| 3 | From the leader send a broker message while the thread is idle: a launcher-started turn runs, is visible in the TUI, and the receipt is recorded. | P9 | not run |
| 4 | Start a long operator turn, then send a broker message: record whether the batch waited or steered into the turn. | P6 | not run |
| 5 | Trigger a command approval in an operator turn: it appears in the TUI and the launcher did not decline it; record the session log lines. | P7 | not run |
| 6 | `/quit` in the TUI: the shell prompt returns, exit 0, `pgrep -f 'codex app-server'` finds none, the terminal is usable. | REQ-MT-010, P9 | not run |
| 7 | Kill the App Server from another shell: the TUI stops, exit 1; record the terminal state. | REQ-MT-011, known debt 5 | not run |
| 8 | `MOAI_FACTORY_MANAGED_TUI=off`: headless behavior with one notice. | REQ-MT-003(c) | not run |

Failure disposition: any failed step becomes a named known debt or a follow-up card. Step 2 or 3 failing means the opt-out should become opt-in. Until steps 1, 2, 3 and 6 are recorded as observed the documentation says "addressed, unverified", never "resolved" (REQ-MT-013).
