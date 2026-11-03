# t2025 progress record (lane-4, run tnwnit)

- 2026-11-03 intake: leased via `moai factory next --card t2025 --run tnwnit`; tree /Users/goos/.moai/worktrees/projects/moai-adk-go/t2025, branch WT-testfactoryagealarmcalle.
- base: merged release/main-batch 98d6257439 into the card tree -> HEAD 6b29bf9b80 (merge commit, card t2025).
- class: B (defect, cause unknown; no SPEC; related completed SPEC-CARD-AGE-ALARM-001). Stage order: run (manager-develop, effort medium) -> card-review -> merge-ready. STOP at merge-ready; no push/PR/paid audit/complete.
- scope (card text): reproduce first; if product code is the cause, do not fix and report; if the test fixture is the cause, repair only internal/cli/factory_age_alarm_test.go minimally, without weakening the cancellation check, without longer timeouts or skip. No full internal/cli run.
- lease check before resuming the writer: `moai factory renew t2025 --run tnwnit` -> "t2025 lease renewed until 2026-11-03T07:00:41.244869Z"; `git rev-parse --show-toplevel` -> the t2025 path; `git status --short` -> empty. Single writer: the first manager-develop spawn had stopped (no live writer) before it is resumed.

## Gaps (refusals, recorded per leader instruction)
- G1: the writer's first command was refused by the worktree-isolation guard before running. Verbatim refusal: "This session is isolated in the worktree /Users/goos/.moai/worktrees/projects/moai-adk-go/t2025, but this command names git in a form too complex to verify that it stays inside the worktree. Refusing to run it — a worktree-isolated session's git operations must target its own worktree. Split it into plain, separate commands and run them from /Users/goos/.moai/worktrees/projects/moai-adk-go/t2025."
  Refused command (writer's own quote): `cd <t2025 tree> && mkdir -p .moai/reports/t2025 && grep -n "factoryAgeAlarmGitTimeout" internal/cli/*.go | head; (uptime; timeout 400 go test ./internal/cli -run '^TestFactoryAgeAlarmCallerCancellation$' -count=20 > .moai/reports/t2025/before-unloaded.txt 2>&1; echo "exit=$?"; uptime) 2>&1; grep -c -- '--- FAIL' .moai/reports/t2025/before-unloaded.txt; grep -n "never started\|elapsed=" .moai/reports/t2025/before-unloaded.txt | head; tail -5 .moai/reports/t2025/before-unloaded.txt`
  The command contains no git invocation; the guard likely matched `Git` inside the identifier `factoryAgeAlarmGitTimeout` (a guess, unconfirmed). Nothing in that command ran, so it yields no evidence.
- Resume rule from the leader: continue with single plain commands (one `uptime`, one `go test`, one `grep` per call; no `cd`, `&&`, `;`, parentheses, variables; output to a file path or run_in_background; Write tool for stored text). A repeat refusal or an access/ownership refusal stops the card with the verbatim text.
- Observed before any test run (writer, read-only): `uptime` load averages 40.10 40.71 59.34. This is machine load at that moment, not a measured baseline of the test.

## Evidence
Before the fix (writer's observation, NOT re-observed by the lane; source: .moai/reports/t2025/repro.md, untracked scratch note):
- command: `go test ./internal/cli -run '^TestFactoryAgeAlarmCallerCancellation$' -count=20 -timeout 300s -json`; result 9 of 20 failed, each at `factory_age_alarm_test.go:862: in-flight control: git never started: stat .../started: no such file or directory` (~0.6s each); uptime load averages 13.65 24.31 46.38 before, 13.27 22.86 44.72 after.
- cause (writer's probe outside the tree): the first exec of a freshly built, renamed binary showed its marker after 290, 280, 282, 624, 282, 370, 284, 331 ms; the test's 300ms cancel timer is the same size, so the control assertion races the process start. Product code (`computeFactoryAgeAlarms` / `factoryAgeAlarmCommitAt`) starts git directly and is not the cause. The probe was 8 runs of a scratch program; the writer did not measure other work inside the old 300ms window.
Repair: commit 4b0d275fe6, only internal/cli/factory_age_alarm_test.go (24 insertions, 3 deletions). The lane read the diff: the 300ms timer is replaced by a cancelable context that a goroutine cancels once the marker file appears (10s poll bound); elapsed is measured from the cancel moment with the same `> 1s` bound; a missing marker still fails with `in-flight control: git never started`; the marker Stat check and the already-canceled-caller assertions are unchanged. No longer timeout, no skip.
After the fix (observed by the lane in this run, tree HEAD 4b0d275fe6, one plain command each):
- `uptime` -> load averages 10.17 17.39 38.26
- `go test ./internal/cli -run '^TestFactoryAgeAlarmCallerCancellation$' -count=20 -timeout 300s` -> `ok  github.com/modu-ai/moai-adk/internal/cli  16.272s`
- `uptime` -> load averages 10.51 16.89 37.47
Writer's own after-fix runs (not re-observed by the lane): same command 20 pass 0 fail (load 9.38 20.46 42.71 before, 11.70 19.86 41.48 after); `-run TestFactoryAgeAlarm` ok 7.689s; `-count=5 -race` ok 6.109s; `go vet ./internal/cli/` and `gofmt -l` on the file empty.
Gaps in this evidence: the before/after load levels differ (before ~13/24/46, lane's after ~10/17/38), so the comparison is same command and count but not an identical load; no extra load was added; the original full internal/cli failure (t2021) was not reproduced (full run forbidden); Windows/Linux first-exec latency not measured; lane did not run the writer's -race run.
Hooks note: the editor reported lint hints in factory_age_alarm_test.go at lines 271/358/415/421/497 (pre-existing, outside the changed range 851-880).

## Task state
- run DONE: repair committed as 4b0d275fe6 by the writer (one explicit-pathspec commit).
- card-review DONE: .moai/reports/t2025/card-review.md. Tool verdict `fail` (advisory, static): 2 findings (P1 internal/hook/branch_guard.go:463, P2 internal/cli/factory_release_route.go:252), both in Go code from the merged release/main-batch and outside this card's diff; not adopted, not recorded as a pass.
- state: merge-ready stop at the evidence commit on top of 4b0d275fe6; no push, no PR, no paid audit; `moai factory complete` NOT run.
