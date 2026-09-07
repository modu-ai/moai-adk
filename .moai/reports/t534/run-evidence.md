# t534 — run-phase evidence (SPEC-CODEX-STALE-SPLIT-FOURTH-001)

Worktree `.claude/worktrees/t534`, branch `WT-stale-msg-polarity`.
Run-phase entry HEAD **`9575e8843`**; final HEAD **`b0cb9318a`**.

---

## Claim

1. The stale-path advisory now partitions missing-path entries into four buckets — the format
   string reads `(%d enabled, %d disabled, %d unspecified, %d non-boolean)` — with the fourth
   member rendered unconditionally (REQ-SSF-001, 002, 003, 004).
2. No `SkillEnabled*` state reaches a `default:` arm that increments a named bucket; the retained
   `default:` counts into a separate unknown-state counter reported under its own clause
   (REQ-SSF-001).
3. The finding's trigger, advisory grade, leading missing-count VALUE, and remove directive are
   unchanged, and `unspecified` keeps exactly the absent-key population (REQ-SSF-005).
4. `codexEnabledShapeFinding` is untouched (REQ-SSF-006).
5. All three moved assertions carry a why-comment naming this SPEC, and the t508 "deliberately
   does NOT grow a fourth bucket" paragraph is rewritten with its reversal rationale, not deleted
   (REQ-SSF-008).
6. AC-SSF-001's RED was observed BEFORE the M1 render change and is recorded verbatim
   (acceptance.md §D.4).
7. All 8 acceptance criteria PASS in one run; `go test ./internal/cli/ -count=1` green; `go vet`
   clean; exactly two Go files changed.

## Evidence

### RED, captured before the production change (commit `a9c9a275d`, guard test only)

```
$ go test ./internal/cli/ -run 'TestCheckCodexWiring_NonBooleanEnabledCountedSeparatelyInStaleSplit' -count=1 -v -timeout 1800s
exit=1

--- FAIL: TestCheckCodexWiring_NonBooleanEnabledCountedSeparatelyInStaleSplit (0.01s)
    --- FAIL: .../non-boolean_leaves_unspecified_empty,_bare_booleans_unmoved (0.00s)
    --- FAIL: .../absent_key_stays_unspecified_alongside_a_non-boolean (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.873s
```

The first failure printed the whole check Detail, and both defect halves are in it, from the one
run:

```
… 4 with a path that no longer exists (1 enabled, 2 disabled, 1 unspecified) …
… 1 declare `enabled` with a value that is not a bare TOML boolean …
```

At that commit the production file carried no such token at all:

```
$ grep -c 'non-boolean' internal/cli/doctor_codex.go
0
exit=1
```

### GREEN and the controls (all at `b0cb9318a`)

```
$ go test ./internal/cli/ -run 'TestCheckCodexWiring_NonBooleanEnabledCountedSeparatelyInStaleSplit' -count=1 -v -timeout 1800s
exit=0
--- PASS: TestCheckCodexWiring_NonBooleanEnabledCountedSeparatelyInStaleSplit (0.07s)
    --- PASS: .../non-boolean_leaves_unspecified_empty,_bare_booleans_unmoved (0.03s)
    --- PASS: .../absent_key_stays_unspecified_alongside_a_non-boolean (0.04s)

$ go test ./internal/cli/ -run 'TestCheckCodexWiring_UnspecifiedEnabledReportedSeparately' -count=1 -v -timeout 1800s
exit=0
--- PASS: TestCheckCodexWiring_UnspecifiedEnabledReportedSeparately (0.00s)

$ go test ./internal/cli/ -run 'TestCheckCodexWiring_StaleHomeSkillsReported|TestCodexSkillPath_AbsoluteExistingAndMissing' -count=1 -v -timeout 1800s
exit=0
--- PASS: TestCheckCodexWiring_StaleHomeSkillsReported (0.00s)
--- PASS: TestCodexSkillPath_AbsoluteExistingAndMissing (0.00s)

$ go test ./internal/cli/ -run 'TestCheckCodexWiring_StaleHomeSkillsReported' -count=1 -v -timeout 1800s
exit=0
--- PASS: TestCheckCodexWiring_StaleHomeSkillsReported (0.00s)
```

Swept-count (§D.0) satisfied on every selector: one `--- PASS:` line per named test, two where the
criterion names two. No run printed `no tests to run`.

### Suite, vet, build, scope

```
$ go test ./internal/cli/ -count=1 -timeout 1800s
exit=0
ok  	github.com/modu-ai/moai-adk/internal/cli	479.657s

$ go test ./internal/codexwiring/ -count=1 -timeout 600s
exit=0
ok  	github.com/modu-ai/moai-adk/internal/codexwiring	0.656s

$ go vet ./internal/cli/...
exit=0   (no output)

$ gofmt -l internal/cli/doctor_codex.go internal/cli/doctor_codex_test.go
(no output)

$ go build ./...
exit=0   (no output)

$ GOOS=windows GOARCH=amd64 go build ./...
exit=0   (no output)

$ git diff --name-only e0c904f58..HEAD -- '*.go'
internal/cli/doctor_codex.go
internal/cli/doctor_codex_test.go
```

### Diff criteria

`git diff --unified=0 e0c904f58..HEAD -- internal/cli/doctor_codex.go` — every hunk header names
`func codexStaleSkillFinding()`, first hunk at line 821. `codexEnabledShapeFinding` (lines
755-790) appears in no hunk (AC-SSF-006). The switch reads:

```go
case codexwiring.SkillEnabledTrue:      missingEnabled++
case codexwiring.SkillEnabledFalse:     missingDisabled++
case codexwiring.SkillEnabledUnspecified: missingUnspecified++
case codexwiring.SkillEnabledNonBoolean:  missingNonBoolean++
default:                                missingUnknownState++
```

`git diff e0c904f58..HEAD -- internal/cli/doctor_codex_test.go` — three edited assertions, each
with a preceding comment naming `SPEC-CODEX-STALE-SPLIT-FOURTH-001`; one rewritten (not removed)
rationale paragraph under the heading `PRIOR DECISION, REVERSED` (AC-SSF-007). Grep for an added
`t.Setenv("HOME"` over the same diff returned exit 1, no output (AC-SSF-008).

## Baseline-attribution

Every measurement above was taken in this worktree, in this run.

- The **pre-change** baselines (`grep -c 'non-boolean'` → 0; the three green preflight tests
  asserting the old three-member strings) were measured at HEAD `9575e8843`.
- The **post-change** measurements were taken at HEAD `b0cb9318a`.
- acceptance.md pins its cells to `e0c904f58`. That pin carries over legitimately, and this is
  measured rather than assumed: `git diff --stat e0c904f58..9575e8843 -- internal/cli/doctor_codex.go
  internal/cli/doctor_codex_test.go` produced NO output (exit 0) — the two files under change are
  byte-identical across the interval, and `git merge-base --is-ancestor e0c904f58 HEAD` succeeded.
  The single intervening commit `9575e8843` is the plan-phase artifact commit.
- Preflight (plan.md §C) was satisfied before any edit: three `--- PASS:` lines from the three
  named tests, and `grep -rn "with a path that no longer exists" --include='*.go' . | grep -v
  '_test.go'` returned exactly one production line (`doctor_codex.go:904`).

## Gaps — explicitly NOT observed

1. **Windows and Linux test EXECUTION.** Only a `GOOS=windows GOARCH=amd64 go build ./...` compile
   was run (exit 0). No Windows or Linux test binary was executed; the cross-platform test verdict
   is CI's, not this run's.
2. **The full repository test suite.** Only `./internal/cli/` and `./internal/codexwiring/` were
   run, per the lane-local verification rule. `go test ./...` was deliberately NOT run locally.
3. **`golangci-lint`.** Not run this session; only `go vet ./internal/cli/...` and `gofmt -l`. No
   lint baseline was measured, so no claim is made about lint delta beyond vet and gofmt.
4. **The real `moai doctor` CLI surface.** No end-to-end `moai doctor` invocation was made against
   a real config; the four-member render is observed through the Go test fixtures only. (The
   original CLI-level reproduction lives in `.moai/reports/t534/reproduction.md` and predates this
   run — it is cited as context, not re-measured here.)
5. **The `missingUnknownState` clause's rendered text.** It is unreachable while `SkillEnabled`
   has four states, so no test exercises it and its wording has never been printed by a running
   check. Its correctness rests on reading, not observation.
6. **The M3 forward-pointer HISTORY row** in `SPEC-CODEX-ENABLED-FATAL-001/spec.md` was not
   written (skip recorded in progress.md §E.2 deviation 3), so no evidence exists for it — it was
   not attempted.

## Residual-risk

1. **The unknown-state escape hatch is a judgment call, not a plan mandate.** plan.md §F M1 offered
   "panic or record an explicit unknown state"; this run chose the latter because
   `internal/cli/CLAUDE.md` forbids `panic()` and `.golangci.yml` enables no `exhaustive` linter.
   A reviewer could reasonably hold that a fifth counter is more machinery than a Tier S card
   should add, and that a four-case switch with no `default:` (letting an unknown state fall out
   of every bucket) is simpler. That alternative silently drops the entry from the leading count,
   which is why it was rejected — but the trade is a real one and is flagged rather than hidden.
2. **A future fifth state is reported, not prevented.** The mechanism makes it loud at runtime; it
   does not make it break the build, because this repository has no exhaustiveness linter. The
   plan's stated intent ("a future fifth state must break the build rather than silently join a
   bucket") is therefore only partially met — the "silently" half is closed, the "break the build"
   half is not achievable at this layer.
3. **Message-width regression is unmeasured.** The advisory line grew by roughly 16 characters.
   The panel sizes itself to its widest row (per the `codexEnabledShapeFinding` comment), and no
   test asserts a width bound, so a narrow terminal's rendering of the longer line was not
   observed.
4. **Downstream consumers of the three-member phrase.** The preflight grep covered `*.go`. A
   non-Go consumer — a docs page, a shell script, a fixture outside Go — that matches the old
   three-member parenthesis would not have been found by that scan.

---

## Orchestrator review of deviation 1 — the fifth counter is necessary, not surplus

The implementer flagged `default: missingUnknownState++` as possibly "more machinery than a Tier S
card should add" and asked for a reviewer's judgment. Reviewed against the code, in this tree:

```
$ grep -n 'missing :=' internal/cli/doctor_codex.go
901:	missing := missingEnabled + missingDisabled + missingUnspecified + missingNonBoolean + missingUnknownState
```

`missing` **is** the sum of the arms, and it is what the leading `%d with a path that no longer
exists` renders. So the alternative — omitting `default:` entirely — is not the simpler option it
looks like: a state added later would increment nothing, drop out of the sum, and make the leading
count **understate the number of entries whose paths are genuinely absent**. That is the failure
this card's own judgment section rejected when it turned down "drop non-boolean entries from the
advisory count": *it trades a wrong label for a wrong number, and the number is what the "remove the
stale entries" directive is sized against.*

Omitting the arm would reproduce that rejected shape one state later. Keeping it preserves the
invariant the parenthesis depends on — the members sum to the leading count — under a change nobody
has made yet.

**The concern was that it is untested (gap 5), and that part stands**: the clause is unreachable at
four states, so its text has never been printed and its correctness rests on reading. That is a real
gap, correctly recorded. But it is the cost of a guard against a future edit, not evidence that the
guard is surplus — and roughly three lines is a proportionate price for keeping a rendered count
honest.

**Do not strip it in review.** A reviewer seeing an unreachable branch and deleting it would be
removing the only thing standing between a fifth state and a silently wrong leading count.

## Orchestrator note — the SPEC's milestone order and its Definition of Done disagreed

`plan.md` §F orders M1 → M2 → M3, while `acceptance.md` §D.4 requires AC-SSF-001's RED captured
BEFORE the M1 render change. Those cannot both be followed. The implementer took §D.4 — correctly:
a RED reconstructed after the change is not an observation — and reported the tension rather than
silently reordering.

This is a defect in the SPEC as authored, not in the implementation. Recorded here because the
same shape (a plan and a gate that disagree about order) is cheap to reproduce and was not caught
by plan-audit.
