# t581 — reproduction (t570 M-4, seam-bypass mutant survives the guards)

card: t581
tree: .claude/worktrees/t581
branch: WT-seam-bypass-mutant
HEAD: c8203fbf3f3443feec351085e6a7d6910dbdba5a (= local develop, re-read at dispatch)
measured: 2026-09-10, in ONE approved `internal/cli` slot

## Claim

Three claims, all established by the single run below.

1. **The defect is live, not stale.** With the seam bypass injected, all three t570 guards
   PASS. The bypass is what the card names verbatim: `strings.ReplaceAll(e.Path, "/", "\\")`
   in place of `fromConfigPath(e.Path, configPathSeparator)`.
2. **A `'/'`-pinned run REACHES the absolute arm.** This was the open risk — if the arm were
   unreachable under a `'/'` pin, the proposed repair would add a green that guards nothing.
   It is reachable, and that is measured rather than reasoned.
3. **The repair kills the mutant.** The `'/'`-pinned identity guard FAILS under the bypass,
   with the value difference named in the failure message.

## Evidence

Concurrency condition at the time of the run, recorded because the lead's slot was granted
under it: lane-6's ~60-minute full `internal/cli` run was executing concurrently.

```
$ uptime
14:55  up 2 days,  8:49, 12 users, load averages: 11.02 11.05 12.22
```

This is a VALUE measurement (which assertion fires, on what byte string), not a timing or
throughput measurement, so contention cannot flip the verdict. The condition is recorded so
the next reader can judge that for themselves rather than inheriting the judgement.

Mutant, applied at `internal/cli/doctor_codex.go:858`, the `codexPathAbsolute` arm:

```go
-			statPath = fromConfigPath(e.Path, configPathSeparator)
+			statPath = strings.ReplaceAll(e.Path, "/", "\\") // MUTANT t581 M-4
```

Run (output to file, exit code read unpiped, per the instrument contract):

```
$ go test ./internal/cli/ -count=1 -v -run 'TestCodexStaleSkillFinding_(AbsoluteStatTargetIsConvertedForm|ExtendedLengthStatTargetIsConvertedForm|ClassificationPrecedesConversion|SlashPinnedAbsoluteArmIsReached|SlashPinnedStatTargetIsUnconverted)' > t581-m4.txt 2>&1
$ echo "TEST_EXIT=$?"
TEST_EXIT=1
```

Verbatim output:

```
=== RUN   TestCodexStaleSkillFinding_AbsoluteStatTargetIsConvertedForm
--- PASS: TestCodexStaleSkillFinding_AbsoluteStatTargetIsConvertedForm (0.00s)
=== RUN   TestCodexStaleSkillFinding_ExtendedLengthStatTargetIsConvertedForm
--- PASS: TestCodexStaleSkillFinding_ExtendedLengthStatTargetIsConvertedForm (0.00s)
=== RUN   TestCodexStaleSkillFinding_ClassificationPrecedesConversion
--- PASS: TestCodexStaleSkillFinding_ClassificationPrecedesConversion (0.00s)
=== RUN   TestCodexStaleSkillFinding_SlashPinnedAbsoluteArmIsReached
--- PASS: TestCodexStaleSkillFinding_SlashPinnedAbsoluteArmIsReached (0.00s)
=== RUN   TestCodexStaleSkillFinding_SlashPinnedStatTargetIsUnconverted
    doctor_codex_seam_use_guard_test.go:95: stat target = "\\Users\\u\\skills\\probe\\SKILL.md", want the declared form "/Users/u/skills/probe/SKILL.md" unchanged — the conversion did not go through the fromConfigPath seam
--- FAIL: TestCodexStaleSkillFinding_SlashPinnedStatTargetIsUnconverted (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.981s
```

### Why the three existing guards cannot see the bypass

They pin the separator to `'\\'` and compare against a backslash literal
(`doctor_codex_path_guard_test.go:37-38`):

```go
const declared = "/Users/u/skills/probe/SKILL.md"
const wantStat = `\Users\u\skills\probe\SKILL.md`
```

and the seam under that pin is (`internal/cli/codex_config_path.go:61-66`):

```go
func fromConfigPath(p string, sep rune) string {
	if sep == '/' { return p }
	return strings.ReplaceAll(p, "/", string(sep))
}
```

With `sep` pinned to `'\\'`, the seam call and the bypass compute the same bytes. The guards
pin the converted VALUE; what needs pinning is that the seam was USED.

### Why the reachability test is separate from the identity test

They are split deliberately, and the split is what makes the FAIL interpretable.
`SlashPinnedAbsoluteArmIsReached` asserts only that the arm runs and hits the stat seam once
— it PASSES under the bypass, because the bypass changes the value the arm computes, not
whether the arm runs. So the identity test's FAIL is a value failure, and cannot be misread
as "the code path was never reached". A guard that fails because it is unreachable pins
nothing.

The classifier is what makes the arm reachable, and it does not consult the pinned
separator at all (`doctor_codex.go:685-687`):

```go
func classifyCodexSkillPath(p string) codexSkillPathShape {
	if filepath.IsAbs(p) { return codexPathAbsolute }
```

`filepath.IsAbs` is host-dependent, not `configPathSeparator`-dependent. That reading
predicted reachability; the run is what established it.

## Baseline-attribution

Measured in this run, on this tree, at HEAD c8203fbf3 with exactly one production-source
mutation applied and then reverted. Nothing is carried over from the t570 measurement — the
`m4-missed.log` in `.moai/reports/t570/` records the same shape at that card's tree, but the
run above is this tree's own.

Mutant revert verified byte-exact:

```
$ git diff -- internal/cli/doctor_codex.go
(no output)
```

## Clean-tree confirmation (second approved slot)

The gap the first run left open — the new guard had never run on unmutated code — is closed
here. A guard that fails only under the mutant proves nothing unless it also passes on clean
code; otherwise it could be failing for an unrelated reason in both states.

Slot condition, checked immediately before starting, with the lead's prescribed census
(filtered by executable name, so the awk process cannot match itself):

```
$ ps -eo comm,args | awk '$1=="go" && /internal\/cli/' > t581-ps-precheck.txt 2>&1
$ echo "PS_AWK_EXIT=$?"      → PS_AWK_EXIT=0
$ date '+%H:%M:%S'           → 15:31:16
$ cat t581-ps-precheck.txt
go               go test ./internal/cli/ -count=1 -timeout 60m
```

Exactly one `go` process on `internal/cli` — lane-6's full suite — and no other lane's scoped
compile, so this run did not stack a second scoped burst. The census sees only the `go` driver
process; a test binary whose driver had already exited would not appear in it.

Concurrency condition: lane-6's full `internal/cli` suite ran alongside this run.

```
$ uptime
15:31  up 2 days,  9:24, 13 users, load averages: 9.27 11.15 10.69
```

As before this is a value measurement — which assertion fires, on which bytes — so contention
cannot flip the verdict; the condition is recorded so a later reader can judge that too.

Run, with the mutant's absence re-verified in the same invocation:

```
$ git diff --quiet -- internal/cli/doctor_codex.go ; echo "PROD_CLEAN_EXIT=$?"
PROD_CLEAN_EXIT=0
$ git rev-parse --short HEAD
c8203fbf3
$ go test ./internal/cli/ -count=1 -v -run 'TestCodexStaleSkillFinding_(AbsoluteStatTargetIsConvertedForm|ExtendedLengthStatTargetIsConvertedForm|ClassificationPrecedesConversion|SlashPinnedAbsoluteArmIsReached|SlashPinnedStatTargetIsUnconverted)' > t581-clean.txt 2>&1
$ echo "TEST_EXIT=$?"
TEST_EXIT=0
```

`TEST_EXIT=0` alone would not have been evidence — a `-run` selection matching zero tests also
exits 0. The file was read, and all five tests ran by name:

```
--- PASS: TestCodexStaleSkillFinding_AbsoluteStatTargetIsConvertedForm (0.00s)
--- PASS: TestCodexStaleSkillFinding_ExtendedLengthStatTargetIsConvertedForm (0.00s)
--- PASS: TestCodexStaleSkillFinding_ClassificationPrecedesConversion (0.00s)
--- PASS: TestCodexStaleSkillFinding_SlashPinnedAbsoluteArmIsReached (0.00s)
--- PASS: TestCodexStaleSkillFinding_SlashPinnedStatTargetIsUnconverted (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.945s
```

Both states side by side — the guard is discriminating:

| tree | reachability guard | identity guard | three t570 guards |
|---|---|---|---|
| seam bypass injected | PASS | **FAIL** — names the value difference | PASS (the defect) |
| clean | PASS | **PASS** | PASS |

## Gaps — not observed

- ~~The new guard has NOT been run against unmutated code.~~ Closed by the clean-tree
  confirmation above.
- The full `internal/cli` package verdict. The lead is serializing that package; the slot
  granted here covered one `-run`-scoped compile only.
- The four known-red baseline tests were not exercised — the `-run` selection does not
  match their names.
- Other platforms; darwin only. `filepath.IsAbs` is host-dependent, so the reachability
  result above is a darwin result.

## Residual-risk

The new guard pins the seam's identity behaviour under `'/'`. It does not pin that
`fromConfigPath` is called — only that the value matches what calling it would produce.
A hypothetical bypass that itself branched on the separator (`if sep == '/' { return p }`)
would still pass both pins. Closing that would need a call-observing seam rather than a
value comparison, which is a larger change than this card carries.

## Tier / Class

`git diff --stat` on production source: EMPTY — the card's delivery is one new test file
(`internal/cli/doctor_codex_seam_use_guard_test.go`), zero production-source lines.
Class A's second condition (CI green on the head that will merge) cannot be established
before push, so per the lead's rule the card stays **Class B**: run → sync, no plan phase.
