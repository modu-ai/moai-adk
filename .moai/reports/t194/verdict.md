# t194 — a mechanical holder lock for the release-integration window

Tier M. Branch `WT-integration-lock`, based on `origin/main` (independent of the
v3.1.3 batch). Closes the gap card t181 named in its own completion report.

---

## Claim

1. The release-integration window now has a **record** — holder session, pid, branch,
   worktree, timestamp — in the primary checkout's `.moai/state/integration-lock.json`,
   visible from every linked worktree.
2. `moai integration acquire | status | release` is the lane-facing surface. A second
   live lane is refused by name; `--force` takes a window over and reports what it
   displaced.
3. A PreToolUse guard refuses a non-holder's `git merge` while a live foreign holder is
   recorded. Opt-in (`workflow.integration_lock.enabled`, default false), fail-open on
   every uncertainty, deny sentinel `INTEGRATION_LOCK_VIOLATION:`.
4. The doctrine names the record as the **first** check, keeping t181's `MERGE_HEAD` and
   pre-commit `HEAD` re-read as the later ones.
5. What this does **not** do is stated in the code and repeated below: it is a
   coordination signal, not a capability boundary.

## Design decisions worth reading

**The board lock next door could not be reused, and the reason is lifetime.**
`internal/kanban/board_lock.go` is a mature cross-platform lock with identity recording
and a bounded stale clear — but it is an flock spanning one process's read-modify-write,
so it dies with the process that took it. An integration window spans many CLI
invocations and minutes of human-paced work; an fd cannot represent it. So the window is
a record whose validity is decided by the recorded holder's liveness. The liveness probe
itself IS reused (`kanban.FactoryProcessAlive`, which already has unix and windows twins).

**Indeterminate liveness reads as LIVE.** Treating an unprobeable holder as dead would
clear a window that may be in use. A false "stale" costs two lanes merging at once — the
failure this card exists to prevent; a false "live" costs one operator asking the holder
to release. The asymmetry is not close.

**An unreadable record is a hard error in the CLI and an allow in the guard.** Same fact,
deliberately different answers. The CLI is a lane asking "may I enter?", where refusing to
answer is the safe reply. The guard sits on a hot tool path, where the same refusal would
deny every `git merge` in the repository until someone hand-repaired a JSON file. A guard
that wedges the batch it protects is worse than the overlap it prevents.

**Only `git merge` is guarded.** Widening to commit or push would deny ordinary work in
every worktree for as long as any lane held the window. A lane that has already merged is
past the contended step.

---

## Evidence

### The over-match the first cut shipped, and its repair

The first guard matched `\bgit\s+merge\b` on the raw command string, which denied this:

```
echo 'git merge is mentioned in this string only'
```

Measured, not hypothesised — the test logged the deny. Repaired by reusing the branch
guard's existing `substituteQuotedArguments` scrub rather than writing a second one; the
test row is now a plain assertion instead of a logged exception.

### The test-isolation defect the CLI tests caught in themselves

Two CLI tests failed with a holder they never created:

```
error does not tell the caller how to proceed: release integration window held by
another session: sess-lane5 (pid 38947) since ... on WT-integration-lock
```

`integrationLockRoot()` had resolved through `git rev-parse --git-common-dir` to the
**real repository**, so the earlier "passing" tests had been writing a lock record into
the developer's own `.moai/state`. Confirmed by hand
(`ls /Users/goos/MoAI/moai-adk-go/.moai/state/integration-lock.json` → present), removed,
and re-checked after every subsequent run (absent).

Fixed by consulting `CLAUDE_PROJECT_DIR` first — which is this package's B7 convention
anyway, and is exactly right here: Claude Code sets it to the PROJECT root even inside a
worktree, the measured behaviour `SPEC-MCP-WORKTREE-ROOT-001` had to work around and the
property this lock wants.

### A file I destroyed and restored

My CLI tests were first written to `internal/cli/integration_test.go` — a name already
taken by a 210-line suite covering `Execute` and DI wiring. The Write overwrote it, and
**the full suite went green anyway**, because deleted tests do not fail; they simply stop
existing. `git status` showing that path as `M` rather than `??` is what surfaced it.

Restored from `HEAD` (210 lines before and after, verified), and my tests moved to
`internal/cli/integration_lock_cli_test.go`. Both sets now run:
`TestExecute_InitsDeps` PASS alongside the six new ones, and the full package suite
re-run on the corrected tree is `ok … 329.592s` with zero failures.

Recorded because the near-miss is the instructive part: a green suite is not evidence that
coverage survived a write, and the only signal here was a two-character difference in
`git status`.

### Verification batch (this worktree, this tree)

| Command | Observed |
|---|---|
| `go test ./internal/kanban/ -run IntegrationLock -count=1 -v` | 9/9 PASS, **0 skipped** (checked: the staleness path did run) |
| `go test ./internal/hook/ -run IntegrationLock -count=1 -v` | 9/9 PASS incl. 6 fail-open subtests |
| `go test ./internal/cli/ -run 'TestIntegration\|TestExecute_InitsDeps' -count=1 -v` | 7/7 PASS (6 new + the restored file's suite) |
| `go test ./internal/hook/ -count=1` | `ok … 21.342s` |
| `go test ./internal/cli/ -count=1 -timeout 900s` (corrected tree) | `ok … 329.592s`, `grep -c '^--- FAIL'` → `0` |
| `go test ./internal/kanban/ ./internal/config/ -count=1` | `ok … 11.366s` / `ok … 1.704s` |
| `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -count=1` | `ok … 21.103s` (mirror parity + neutrality) |
| `golangci-lint run` (kanban, hook, cli, config) | `0 issues.` (5 errcheck findings fixed first) |
| `go vet` (same four) | exit 0 |
| `GOOS=windows GOARCH=amd64 go build ./...` + `go vet` | exit 0 |
| `make build` | exit 0; `catalog.yaml` unchanged |
| doctrine byte size, before → after | 25,915 → 26,587 (**+672**, under the 1,000-byte statement threshold) |

## Baseline-attribution

Measured by me in `.claude/worktrees/t194` on `WT-integration-lock`, based on
`origin/main` at `76b2c4ece`, against the tree carrying the final diff.

No template config entry was added, and that is deliberate rather than an omission:
`branch_guard` is likewise absent from `internal/template/templates/` (verified by grep —
the only hit is prose in a rule file), so the Go defaults carry both. Template neutrality
forbids `enabled: true` anywhere under the template tree.

---

## Gaps — what was NOT observed

- **No two-session runtime rehearsal.** The guard was exercised through
  `checkIntegrationLock` with seeded records, not by running two real Claude Code sessions
  against one release worktree. The unit level proves the decision function; it does not
  prove the wiring fires under a live PreToolUse event.
- **The deny path was never observed with the config flag ON in a real session.** The
  call-site gate is asserted by reading the code and by the existing branch-guard pattern
  it copies, not by a test that flips `workflow.integration_lock.enabled` and watches a
  real merge get refused.
- **Windows was compiled, not run.** `GOOS=windows go vet` proves compilation and nothing
  about behaviour — the standing lesson. `FactoryProcessAlive`'s windows twin carries its
  own conservative default, and that path is unexercised here. CI's windows job is the
  measurement.
- **Staleness is probed by pid only.** A recycled pid — the OS reusing a dead holder's
  number — would read as live and hold the window until `--force`. Not measured; judged
  acceptable because the failure direction is the safe one (over-holding, not overlap).
- **No concurrent-acquire race test.** Two processes calling acquire at the same instant
  could both read a free window before either writes. The write is atomic
  (`os.Rename`), so the record never tears, but last-writer-wins is possible in that
  window. The announcement protocol makes simultaneous acquire unlikely rather than
  impossible; a compare-and-swap would close it and is not built.

## Residual-risk

- **This is a coordination signal, not a capability boundary.** A determined caller can
  delete the record, pass `--force`, or leave the flag off. The value is that skipping the
  announcement now takes a deliberate, recorded act instead of an honest mistake — that is
  the whole claim, and it should not be read as enforcement.
- **The doctrine bullet was written against `origin/main`, which does not yet carry
  t181's edit** (PR #1603, open). It was deliberately added as a NEW bullet after the
  contested one rather than rewriting it, so the two compose in either merge order — but
  if git cannot isolate the hunks, whoever merges second resolves by keeping both.
- **A lane whose session id is unresolvable cannot acquire at all** (acquire refuses
  rather than inventing an identity). In an environment where the SessionStart side-channel
  is absent, every lane must pass `--session` explicitly or the lock is unusable.
