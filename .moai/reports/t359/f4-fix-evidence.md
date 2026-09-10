# t359 — F4 fix evidence

Fixes F4 of `.moai/reports/t359/sync-audit.md` (Low, not blocking): user-supplied values reach git
subcommands with no end-of-options guard. F1 evidence is the sibling file `f1-fix-evidence.md`;
F2/F5-F9 are untouched.

- **Tree:** `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359`
- **Branch:** `WT-landing-evidence`
- **HEAD at start of F4 work:** `26e4e9200` (re-read in-tree)
- **Files:** `internal/cli/todo_landed.go`, `internal/cli/todo_landed_test.go`
- **git:** `git version 2.50.1 (Apple Git-155)`, darwin/arm64

---

## 1. The token: `--end-of-options`, not `--`

The audit words F4 as a missing "`--` end-of-options separator", which conflates two different git
tokens. Taking it literally breaks the verb. Measured in this tree:

```
$ git rev-parse --verify --quiet 'HEAD^{commit}'                    → rc=0  26e4e9200c979…
$ git rev-parse --verify --quiet -- 'HEAD^{commit}'                 → rc=1  (NO OUTPUT)
$ git rev-parse --verify --quiet --end-of-options 'HEAD^{commit}'   → rc=0  26e4e9200c979…
$ git merge-base --is-ancestor --end-of-options 'HEAD~1' 'HEAD'     → rc=0
$ git merge-base --is-ancestor -- 'HEAD~1' 'HEAD'                   → rc=0
```

In `rev-parse`, `--` separates revisions from PATHS, so a rev after it is read as a path and
resolves to nothing. `--end-of-options` stops option parsing while leaving the operand a revision.
`--` happens to work for `merge-base`, but using two different tokens across one seam would be its
own inconsistency; `--end-of-options` covers both.

Control proving the token actually changes parsing (not a no-op):

```
$ git rev-parse --verify --abbrev-ref HEAD                    → rc=0, "WT-landing-evidence"
$ git rev-parse --verify --end-of-options --abbrev-ref HEAD   → rc=128, "fatal: Needed a single revision"
```

**Version floor.** `--end-of-options` requires git 2.24+ (Nov 2019). The measurement labels are
`progress.md` §J.4's, mirrored rather than re-derived: **MEASURED** — CI runs on `ubuntu-latest` /
`windows-latest`, and no git version requirement was found in `README.md`, `.github/`,
`.moai/docs/`, or `docs-site/content/ko/`. That absence is **SCOPE-BOUNDED** — no exhaustive sweep
was run — and anything beyond those four locations is **UNMEASURED**. So the claim this change
supports is narrow: it introduces a 2.24 dependency that is undeclared *in the locations searched*,
NOT that the project states no git requirement anywhere. It is also the first use of the token in
the Go tree (`grep -rn -e "--end-of-options" --include="*.go"` → no hits outside this file).

The residual is left standing rather than resolved: an undeclared 2.24 floor now binds the verb
(sync-audit-2 N3). Declaring a floor somewhere would be a project-wide decision with no requirement
behind it, so it is recorded as a named residual and nothing more.

## 2. Three call sites, not two

The audit names `:161` and `:216`. The RED below found **three**: `validateSuppliedSHA`'s own
`rev-parse` on the user-supplied `--sha` is a third instance of the identical class, in the same
function the audit was already looking at. All three are guarded. Fixing two of three in one
function would leave an incoherent partial guard; this is closing the finding family, not widening
scope.

| Site | Command | User value |
|---|---|---|
| `buildLandingEvidence` | `rev-parse --verify --quiet <ref>^{commit}` | `--ref` |
| `validateSuppliedSHA` | `rev-parse --verify --quiet <sha>^{commit}` | `--sha` — **not in the audit** |
| `validateSuppliedSHA` | `merge-base --is-ancestor <resolved> <ref>` | `--ref`, passed RAW |

## 3. The RED — and what it is NOT

```
$ go test ./internal/cli/ -run "TestLandedGitCallsGuardEndOfOptions|TestLandedOptionShapedRefIsRefused" -count=1 -v
=== RUN   TestLandedGitCallsGuardEndOfOptions
    todo_landed_test.go:679: git [-C /var/…/001 rev-parse --verify --quiet HEAD^{commit}]: no --end-of-options guard; a user-supplied operand reaches git as a parsable option
    todo_landed_test.go:679: git [-C /var/…/001 rev-parse --verify --quiet f1f7fdd421a5…^{commit}]: no --end-of-options guard; a user-supplied operand reaches git as a parsable option
    todo_landed_test.go:679: git [-C /var/…/001 merge-base --is-ancestor f1f7fdd421a5… HEAD]: no --end-of-options guard; a user-supplied operand reaches git as a parsable option
--- FAIL: TestLandedGitCallsGuardEndOfOptions (0.64s)
=== RUN   TestLandedOptionShapedRefIsRefused
--- PASS: TestLandedOptionShapedRefIsRefused (0.47s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.109s
rc=1
```

**Firing assertion:** `todo_landed_test.go:679`, the `idx < 0` branch, three times — once per
unguarded call site. The test drives the real CLI verb against a real git fixture and records argv
through the `todoRunCommand` seam.

**[HARD] This is a SHAPE assertion, not a behavioural RED, and I will not dress it as one.**
The house rule is observe-the-RED-before-the-fix; for F4's *security* property that rule cannot be
satisfied, and the honest report of that is the deliverable.

- The auditor's write probe (`--ref '--output=/tmp/t359_pwn'`) gives rc=1 and no file. I re-ran it
  guarded and unguarded: **both** rc=1, **both** no file created. Indistinguishable.
- `TestLandedOptionShapedRefIsRefused` **passes on the unfixed code** — correctly. It is a
  characterization test pinning an existing invariant, not a defect demonstration, and it is
  labelled as such in the source.

**One correction to the framing I was given.** "Guarded and unguarded are indistinguishable by any
probe available today" is true of the *write* probe but not of option interpretation generally. At
the raw-ref site the injection **is** mechanically observable in isolation:

```
$ git merge-base --is-ancestor <sha> --independent                  → rc=129
    error: options '--independent' and '--is-ancestor' cannot be used together   ← consumed as an OPTION
$ git merge-base --is-ancestor --end-of-options <sha> --independent → rc=128
    fatal: Not a valid object name --independent                                 ← treated as an OPERAND
```

That is a real difference, and it is why the guard is not cosmetic. But it is **not reachable
through the CLI**, and the structural reason is worth recording because it is stronger than "no
exploit was found": the `^{commit}` gate in `buildLandingEvidence` runs first and refuses every
option-shaped ref (`--independent^{commit}` resolves to nothing), so the raw-ref site never receives
an option-shaped value. The two `rev-parse` sites are independently immune for the same
concatenation reason — `--default=HEAD^{commit}`, `--short=40^{commit}`, `--abbrev-ref^{commit}`,
`--disambiguate=26e4^{commit}` all give rc=1 identically guarded and unguarded.

So: **not exploitable today, by construction, at all three sites.** Never "safe" — the immunity
rests on an incidental property (string concatenation) of the current call shapes, which is exactly
the kind of thing a future edit removes silently. `TestLandedOptionShapedRefIsRefused` exists to
make that removal visible: if it ever fails, the latent guard has become live.

## 4. The fix

One `const gitEndOfOptions = "--end-of-options"` (per the repo's no-hardcoding rule) inserted before
each user-controlled operand at all three sites. The constant's doc comment records why it is not
`--`, the 2.24 floor, and the latency reasoning — so the next reader who "simplifies" it to `--`
finds the measurement first.

## 5. A regression I introduced and repaired

The first version of that comment contained the literal token `merge-base`. `internal/cli` then
failed:

```
mcp_build_identity_test.go:638: sweep: NEW ancestry hit todo_landed.go:250 — a second comparison outside binlag.Evaluate (REQ-ABI-006 violation)
```

`TestAuditLagUsesBinlagSeam` greps every non-test `.go` file in the package for `merge-base` /
`is-ancestor` and compares against a fixed coordinate baseline; a mention in a *comment* registers
as a phantom coordinate. A pre-existing comment 40 lines below in the same file warns about exactly
this, and I walked into it anyway. Repaired by naming the command indirectly ("the raw-ref
reachability call in validateSuppliedSHA") and adding a pointer to the sibling warning.

Recorded rather than quietly fixed because it is the same failure class this card keeps finding: it
would have been invisible to a bounded `head` view of the suite output, and was caught only by
scanning the full output for `FAIL`.

## 6. Gates

Every verdict is the unfiltered `$?`; `internal/cli` was captured to a file and scanned across the
whole output for both `FAIL` and `test timed out`.

| Gate | Command | Result |
|---|---|---|
| F4 targeted | `go test ./internal/cli/ -run "TestLandedGitCallsGuardEndOfOptions\|TestLandedOptionShapedRefIsRefused" -count=1 -v` | both PASS — **rc=0** |
| sweep regression | `go test ./internal/cli/ -run TestAuditLagUsesBinlagSeam -count=1 -v` | PASS — **rc=0** |
| cli suite | `go test ./internal/cli/... -count=1 -timeout 1800s` | `ok … 440.472s` — **rc=0**; zero `FAIL`, zero `timed out` |
| kanban suite (F1 regression) | `go test ./internal/kanban/... -count=1 -timeout 900s` | `ok … 142.625s` — **rc=0** |
| vet | `go vet ./internal/kanban/... ./internal/cli/...` | no output — **rc=0** |
| lint | `golangci-lint run ./internal/cli/... ./internal/kanban/...` | `0 issues.` — **rc=0** |
| gofmt (my files) | `gofmt -l` on the 4 touched files | no files listed |

**Timeout note.** An intermediate run at `-timeout 600s` failed with `panic: test timed out after
10m0s` and **zero** `--- FAIL` lines — a cumulative-duration timeout, not an assertion failure (the
test in flight, `TestAuditVerdictCarriesBuildCommit`, had run 0s). The suite measured 508s, 601s,
and 440s across three runs, so it sits at the 600s boundary and the verdict depends on machine load.
The rc=0 above is from the 1800s run. Anyone re-deriving this should not use 600s.

**Pre-existing gofmt deviations (NOT mine, not fixed).** `gofmt -l internal/cli/` lists 28 files
(`epic.go`, `web.go`, `schema_bridge.go`, 25 others). All are unmodified at HEAD — `git diff HEAD --
internal/cli/epic.go` is empty — so they predate this card. My four touched files are clean.
Reported, not fixed: out of scope, and a 28-file reformat would bury this card's diff.

## 7. Reported, not fixed

- **The write-side gap.** `specid.ValidateSpecID` is called from `spec_view.go`, `spec_status.go`,
  `spec_close.go`, `github.go`, `board_store.go`, and `status_read.go` — but **not** from
  `internal/cli/todo.go`, so `moai todo next --spec` still admits an unvalidated value into the
  queue. F1 blocks it at the READ boundary, which is the right layer and matches the sibling.
  Closing the write side would change what `todo next` accepts — a behavior change with no criterion
  behind it. Follow-up card candidate.
- **F3 is closed by F1's fix, not separately open.** The audit lists it at `:105` as subsumed. The
  `#nosec G304` justification now reads `specID validated above; path is project-local`, pointing at
  where the invariant is enforced rather than asserting it. Recording this so F3 is not later
  re-raised as open.
- **28 pre-existing gofmt deviations** in `internal/cli` (§6).

## 8. Known losses — NOT measured

- **No cross-platform verification.** darwin/arm64 only. `--end-of-options` is a git feature, not an
  OS one, but no Windows or Linux run was performed, and the git floor is unverified on any other
  platform's shipped git.
- **The git-version floor is stated from documentation, not measured.** I did not test against a
  git older than 2.50.1, so "2.24+" is git's documented introduction version, not something I
  observed failing below.
- **No full-suite run** (`go test ./...`); per repo rule that verdict is CI's on the pushed head, and
  I pushed nothing.
- **No coverage measurement** for either package, before or after.
- **The option-injection probes in §3 were run as raw `git` commands**, not through the verb. That is
  the point — the verb refuses first — but it means the guarded/unguarded difference is evidenced at
  the git level and the unreachability at the CLI level, from two different measurements rather than
  one.
- **`TestLandedGitCallsGuardEndOfOptions` asserts only the two subcommands it watches.** A future git
  call added to this verb with a different subcommand name is not covered and would pass silently.

## 9. Residual risk

- **The guard is latent, so nothing exercises it.** No test can fail if `--end-of-options` stops
  working, because no reachable input reaches the guarded parse. The shape assertion is the only
  thing holding it in place, and a shape assertion can be satisfied by a wrong-but-present token.
- **Immunity rests on string concatenation.** If a future edit drops `^{commit}` from either
  `rev-parse` call, option-shaped refs become reachable at the raw-ref site immediately. That is the
  scenario the guard is for, and `TestLandedOptionShapedRefIsRefused` is the only alarm.
- **The 2.24 floor is newly introduced and undeclared.** On a host with git < 2.24 all three calls
  now fail where they previously succeeded — the verb would refuse every record rather than
  mis-record, so it fails safe, but it fails.
- **`TestAuditLagUsesBinlagSeam` remains coordinate-brittle.** Its baseline is line-keyed, so any
  future edit above the tracked lines in `todo_landed.go` re-breaks it. My change did not move the
  tracked `:216` coordinate, but that was luck of placement, not design.

---

# Delta re-audit follow-up — N1, N2, N3 (card t359)

Same finding family, same file: the re-audit (`sync-audit-2.md`) found two new defects **inside the
F4 fix recorded above**, plus one deferred wording correction. HEAD at the start of this work:
`96826161f958f6aad6ede939f1a67d17e03fb31b`, tree clean, branch `WT-landing-evidence`, tree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359`.

## N1 — the position assertion was a tautology

### What was wrong

`TestLandedGitCallsGuardEndOfOptions` carried two assertions. The presence half is sound and has an
observed RED (§3 above). The position half was:

```go
if idx != len(c.argv)-1-countOperandsAfter(c.argv, idx) { … }
func countOperandsAfter(argv []string, idx int) int { return len(argv) - idx - 1 }
```

which reduces to `idx != idx` — always false.

**Why it survived review, stated plainly.** The expression *looked* like a position check because
both sides computed the same index by different routes: one counting forward from `idx`, the other
backward from the end. Two different-looking derivations of one value read as a comparison and were
an identity. The lesson is not "check the arithmetic" but "a shape assertion is not verified until a
mutant in the shape it guards has been seen to reject it".

### What the corrected assertion asserts, in words

**No operand may appear BEFORE the `--end-of-options` token.** An operand is any argument that is
not a `-`-prefixed flag, not the `-C <dir>` pair `todoGitOutput` prepends, and not the git
subcommand itself. An operand placed ahead of the token is still parsed as an option by git, which
is exactly the defect the token exists to prevent.

The assertion is deliberately **not** stated on the arguments *after* the token, and that is not an
oversight: the token's entire purpose is to make everything after it an operand whatever its shape,
so a `--`-prefixed argument there is a legitimate operand and is indistinguishable from a flag by
inspection. Any assertion attempting to classify post-token arguments by shape would be asserting
the opposite of what the token guarantees.

### RED, observed on the case the assertion exists for

The mutant is a real source change, not a test-side reorder: the guard was moved to AFTER both
operands at one call site (`validateSuppliedSHA`'s `merge-base` call in `internal/cli/todo_landed.go`).
Both git forms were probed first to confirm the mutant does not break the command — a broken command
would fail the test at the fixture line and produce a RED for the wrong reason:

```
$ git rev-parse --verify --quiet 'HEAD^{commit}' --end-of-options   → rc=0, 96826161f958f6…
$ git merge-base --is-ancestor HEAD HEAD --end-of-options            → rc=0
```

Mutant applied:

```go
-	if _, err := todoGitOutput("merge-base", "--is-ancestor", gitEndOfOptions, resolved, ref); err != nil {
+	if _, err := todoGitOutput("merge-base", "--is-ancestor", resolved, ref, gitEndOfOptions); err != nil {
```

```
$ go test ./internal/cli/ -run 'TestLandedGitCallsGuardEndOfOptions' -count=1 -v
=== RUN   TestLandedGitCallsGuardEndOfOptions
    todo_landed_test.go:692: git [-C /private/var/folders/…/001 merge-base --is-ancestor
      88bc3d06101a3d07b126f7995a58b52dc83af59b HEAD --end-of-options]: --end-of-options at 6 is
      preceded by operand "88bc3d06101a3d07b126f7995a58b52dc83af59b", which git still parses as an option
--- FAIL: TestLandedGitCallsGuardEndOfOptions (0.64s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.655s
```

**Firing assertion: `todo_landed_test.go:692`** — the position assertion, on the merge-base call.
The verb itself still succeeded (no fixture `t.Fatalf`), so the RED is attributable to the position
change and to nothing else.

The old assertion was then run over that same observed argv, verbatim, in a standalone program:

```
old-assertion-fires=false
```

So on the exact input the corrected assertion rejects, the tautology stayed silent. That is the
comparison that makes this a repair rather than a rewrite.

Mutant reverted; the call site is back to `gitEndOfOptions, resolved, ref`.

## N2 — the const swallowed `todoGitOutput`'s doc comment

`go doc` before:

```
$ go doc -u -all ./internal/cli | grep -A3 'const gitEndOfOptions'
const gitEndOfOptions = "--end-of-options"
    todoGitOutput runs one git command against the queue's own repository
    through the shared process seam, and returns its trimmed stdout.

    `-C <root>` is explicit: the queue resolves against the PRIMARY checkout,
```

`go doc` after (the const's own paragraph now leads; the function has its doc back):

```
$ go doc -u -all ./internal/cli | grep -A3 'const gitEndOfOptions'
const gitEndOfOptions = "--end-of-options"
    gitEndOfOptions stops git's option parsing so a user-supplied operand
    (`--ref`, `--sha`) cannot be read as a flag, whatever its shape.

$ go doc -u -all ./internal/cli | grep -A4 'func todoGitOutput'
func todoGitOutput(args ...string) (string, error)
    todoGitOutput runs one git command against the queue's own repository
    through the shared process seam, and returns its trimmed stdout.

    `-C <root>` is explicit: the queue resolves against the PRIMARY checkout,
```

Verified through `go doc`, not by reading the source — the source reads plausibly in both states.

## N3 — the git-floor sentence, corrected

The **Version floor** paragraph in §1 above previously read that this change introduces a 2.24
dependency *"where none was previously stated"*, which upgrades a scope-bounded absence into a
project-wide one. It now mirrors `progress.md` §J.4's labels verbatim (MEASURED / SCOPE-BOUNDED /
UNMEASURED) rather than re-deriving them, and states the narrow claim the search actually supports.

N3 itself is **left open as a named residual**. Declaring a git floor somewhere would be a
project-wide decision with no requirement behind it, so it is recorded, not resolved.

## Verification

Both suites run **sequentially**, not concurrently. `internal/cli` at `-timeout 1800s`: the
re-auditor measured 615s under concurrent load and a `-timeout 600s` run would produce a red about
the clock rather than the code. Every verdict is the unfiltered `$?`, and the `internal/cli` output
was captured whole (17 lines) and scanned in full — no `head`-bounded window.

```
$ go test ./internal/cli/... -count=1 -timeout 1800s   ; echo "cli rc=$?"
cli rc=0
ok  github.com/modu-ai/moai-adk/internal/cli  497.289s        ok  …/cli/agentlint  0.479s
ok  …/cli/harness  9.127s      ok  …/cli/pr  0.779s           ok  …/cli/preference  3.812s
ok  …/cli/printer  2.098s      ok  …/cli/specid  2.150s       ok  …/cli/taskledger  5.881s
ok  …/cli/uikit  1.177s        ok  …/cli/update  3.189s       ok  …/cli/update/backup  4.455s
ok  …/cli/update/deploy  3.953s   ok  …/cli/update/merge  5.589s   ok  …/cli/update/plan  6.198s
ok  …/cli/update/report  3.507s   ok  …/cli/wizard  5.415s        ok  …/cli/worktree  10.867s
$ /usr/bin/grep -n -- '--- FAIL\|^FAIL\|timed out\|panic' cli.txt   → no output, grep rc=1 (17 lines total)

$ go test ./internal/kanban/... -count=1                ; echo "kanban rc=$?"
kanban rc=0
ok  github.com/modu-ai/moai-adk/internal/kanban  137.198s

$ go vet ./internal/kanban/... ./internal/cli/...       ; echo "vet rc=$?"
vet rc=0   (0 bytes of output)

$ golangci-lint run ./internal/cli/...                  ; echo "lint rc=$?"
0 issues.
lint rc=0

$ gofmt -l internal/cli/ | wc -l                        → 28
$ gofmt -l internal/cli/ | /usr/bin/grep -c 'todo_landed' → 0   (grep rc=1: neither touched file listed)
```

Note the `internal/cli` figure: **497s here vs the re-auditor's 615s**. Same suite, same tree; the
difference is machine load, which is precisely why 615s is not a clean baseline and why the timeout
floor is set well above both.

The 28 `gofmt` deviations are pre-existing and unmodified — the audit recorded the same count, and
neither file this work touched (`todo_landed.go`, `todo_landed_test.go`) appears in the list.

## Known losses — what this work did NOT observe

- **The corrected assertion does not catch a guard placed too EARLY.** `[… rev-parse
  --end-of-options --verify --quiet X]` passes it: no operand precedes the token, so the stated
  invariant holds. That shape is a real defect of a different kind (git would read `--verify` as an
  operand), but it breaks the command outright and is caught loudly by every behavioural test in the
  file rather than silently. Recorded so the guard is not read as wider than it is.
- **Operand-ness before the token is decided by SHAPE**, not by knowing git's grammar. A future call
  site whose flag takes a separate non-`-` value (`git -c key=value`, `--sort <field>`) would have
  that value misread as an operand and produce a false RED. No such call site exists on this seam
  today; all three are flag-only before the token.
- **One mutant, one call site.** The RED was established on `merge-base` only. The two `rev-parse`
  sites were not independently mutated, so the assertion is verified against one of the three shapes
  it covers.
- **N3 is unresolved by design**, and no git below 2.24 was tested here either — the floor remains
  documentary, exactly as `sync-audit-2.md` recorded it.
- **Nothing outside `internal/cli` and `internal/kanban` was measured**, no `-race`, no
  cross-platform build. CI on the pushed head is the full-suite verdict.
- **The other seven open findings (F2, F5-F9) were not touched** and were not re-measured here.
