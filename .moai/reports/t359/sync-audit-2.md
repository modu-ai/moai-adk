# Sync Audit #2 — delta-scoped confirming re-audit · SPEC-TODO-LANDING-EVIDENCE-001 (card t359)

- Auditor: sync-auditor. Fresh agent; the first verdict (`.moai/reports/t359/sync-audit.md`
  @ `d739cd051`) is on disk and was treated as a record to CHECK, not as a premise.
- Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359`
- Branch: `WT-landing-evidence`
- **HEAD re-read by THIS auditor at audit start: `5350d0e63`, working tree clean.** Matches the
  dispatched value; re-read rather than inherited.
- Tier L · PASS threshold 0.85 · profile `default` (`.moai/config/evaluator-profiles/default.md`)
- Mode: flat weighted-percentage (`harness.yaml` sets no `evaluator_mode: hierarchical`)
- Written in English to sit beside `sync-audit.md` and the card's other artifacts, which are English.

---

## Overall Verdict

**FAIL — 0.75 vs the Tier L threshold 0.85. Score-driven; the must-pass firewall did NOT fire.**

The FAIL from `d739cd051` **survives**, and that is the result, not a failure of this run. The two
routed fixes both hold, both were confirmed by measurement rather than by reading the diff, and
neither moves the arithmetic — because the score was never F1's to move. All four dimensions again
sit exactly one rubric anchor below their ceiling, for reasons broader than the two findings that
were fixed.

What changed, and what did not:

| | `d739cd051` | `5350d0e63` |
|---|---|---|
| F1 (blocking, Medium) | open, demonstrated | **CLOSED** — RED reproduced by this auditor at the pre-fix commit, GREEN at HEAD, end-to-end binary probe now stores `unknown` |
| F3 (Low) | open | **CLOSED** — the `#nosec` warrant now names what enforces it |
| F4 (Low) | open, 2 sites named | **CLOSED at 3 sites** — and my own prescription in it was WRONG; the correction is adopted (below) |
| F2, F5, F6, F7, F8, F9 | open | **still open** — all six re-measured at this HEAD, none went away by being disposed |
| new | — | **N1, N2, N3** introduced by the two fix commits |
| Weighted score | 0.75 | **0.75** |

Two of the three new findings are in the F4 commit — which is the dispatch's own prediction holding:
"two commits of security-shaped changes to a closed card is exactly where a new defect enters."

### Dimension Scores

| Dimension | Score | Verdict | Rubric anchor cited (default profile) | Evidence |
|---|---|---|---|---|
| Functionality (40%) | 0.75 | PASS (must-pass) | *"All primary acceptance criteria pass; minor edge cases missing"* | both suites rc=0 at this HEAD; F1's correctness half (REQ-TLE-005/010) now enforced; F2 remains a demonstrated unhandled edge case with an operator ruling not yet executed, and F5 leaves AC-TLE-010 unable to detect the F1 class |
| Security (25%) | 0.75 | PASS (must-pass) | *"No Critical/High findings; Medium findings documented with mitigations"* | the one Medium (F1) is CLOSED by measurement. 1.00 is unavailable: it requires *"No findings of any severity"* and Lows remain (write-side gap, N3's undeclared git floor, F2's availability shape) |
| Craft (20%) | 0.75 | **FAIL against the profile's hard threshold** (not must-pass ⇒ no firewall) | *"Coverage >= 80%, minor style issues, acceptable naming"* | `internal/kanban` **86.5%**, `internal/cli` **80.4%**. The profile states *"Coverage below 85% = Craft FAIL"*; `internal/cli` is below it. Plus N1 (a tautological assertion) and N2 (a misattached doc comment), both new |
| Consistency (15%) | 0.75 | PASS | *"Minor deviations from conventions; no structural inconsistencies"* | the deviation that drove the first 0.75 is GONE — `ReadPrimarySpecStatus` now guards exactly as its sibling does. It is replaced by N2, a localized Go doc-comment convention break introduced by the F4 commit |

Weighted total: `0.40(0.75) + 0.25(0.75) + 0.20(0.75) + 0.15(0.75)` = **0.75**.

Must-pass firewall: Functionality PASS, Security PASS. Neither forces the FAIL; the score does.

**One divergence from the first verdict, named rather than hidden.** `sync-audit.md` marked Craft
**PASS** at the same 80.4%. The profile's Hard Thresholds section says *"Coverage below 85% = Craft
FAIL"*, so the verdict token there was wrong even though its score anchor was right. Craft is not a
must-pass dimension, so the correction changes no arithmetic and no firewall — it changes only what
the row honestly says. I am correcting my predecessor's token, not its number.

---

## Findings

### Delta dispositions (F1-F9, each re-measured at `5350d0e63`)

#### F1 — CLOSED by measurement
The fix (`bedb269a3`) calls `specid.ValidateSpecID(specID)` before the join in
`ReadPrimarySpecStatus` (`internal/kanban/status_read.go:290-292`) and returns the existing
`("", false)` not-read signal, which the caller maps onto `LandingSpecStatusUnknown`.

I did not accept this from the diff. Three independent measurements:

1. **RED, reproduced by me at the pre-fix commit.** `git archive d739cd051` into the scratchpad
   (never this tree), the post-fix TEST copied in, and the suite run there:

   ```
   $ /usr/bin/grep -n 'ValidateSpecID' <redtree>/internal/kanban/status_read.go
   109:  if err := specid.ValidateSpecID(specID); err != nil {     ← ReadCardStatus (the sibling)
   (no hit in ReadPrimarySpecStatus; :285 carries only `#nosec G304 -- project-local SPEC path`)

   $ go test ./internal/kanban/ -run TestReadPrimarySpecStatus_RefusesTraversingSpecID -count=1 -v
       status_read_test.go:419: ReadPrimarySpecStatus(root, "../../../outside") = ("PWNED-TRAVERSAL", true)
   --- FAIL: TestReadPrimarySpecStatus_RefusesTraversingSpecID (0.00s)
   RED rc=1
   ```
2. **GREEN at HEAD**, same command, same test: `--- PASS … ok … 0.262s`, rc=0.
3. **End-to-end through the shipped verb**, a binary built from THIS HEAD, an isolated git project,
   `spec.md` planted outside the project root — with the fixture precondition made non-vacuous
   first, so a refusal cannot be a fixture pointing at nothing:

   ```
   $ ls -l <probe>/project/.moai/specs/../../../outside/spec.md    → exists, rc=0
   $ python3 … realpath comparison                                  → inside_root= False
   $ moai todo add … ; moai todo next 1 --spec '../../../outside' ; moai todo landed t1 --ref HEAD
   $ strings .moai/state/todo/backlog.db | grep -o '{"ref".*}'
   {"ref":"HEAD","ref_head":"e23ac75f…","observed_at":"…","spec_status":"unknown"}
   ```

   Where the first audit read `"spec_status":"PWNED-TRAVERSAL"` out of the same database, this HEAD
   stores `unknown`. The traversal is refused **through the reader**, not merely by the validator in
   isolation.

Side effect worth recording: `ReadPrimarySpecStatus` went from `0.0%` (the first audit's
cross-package attribution artifact) to **75.0%** in the `internal/kanban` profile, because the new
test exercises it in-package. The artifact is now moot rather than merely explained.

#### F3 — CLOSED, subsumed as predicted
`// #nosec G304 -- specID validated above; path is project-local`. gosec still flags the variable
path, so the annotation is correctly kept; what changed is that its warrant now names the thing that
makes it true instead of asserting the unenforced invariant. `golangci-lint` rc=0.

#### F4 — CLOSED at three sites, and my own prescription in it was wrong
`gitEndOfOptions` now precedes the user-controlled operand at `todo_landed.go:161`, `:205`, `:216`.
I confirmed the call-site list is complete for this verb: `grep -n 'todoGitOutput(' internal/cli/todo_landed.go`
returns exactly those three plus the definition at `:260`.

**I adopt the correction, having reproduced it myself** on `git version 2.50.1 (Apple Git-155)`:

```
$ git rev-parse --verify --quiet 'HEAD^{commit}'                    → rc=0  <sha>
$ git rev-parse --verify --quiet -- 'HEAD^{commit}'                 → rc=1  NO OUTPUT   ← my F4 text would have broken the verb
$ git rev-parse --verify --quiet --end-of-options 'HEAD^{commit}'   → rc=0  <sha>
$ git merge-base --is-ancestor <sha> --independent                  → rc=129  "options … cannot be used together"   (consumed as an OPTION)
$ git merge-base --is-ancestor --end-of-options <sha> --independent → rc=128  "Not a valid object name --independent" (an OPERAND)
```

F4's *finding* stands; F4's *prescription* named `--` and `--end-of-options` as one token and was
wrong. The lane's correction is right and is adopted here. The second correction owed to the finding
is likewise confirmed: the injection IS observable at the raw-ref site in isolation (129 vs 128), so
my "indistinguishable by any probe available today" was true of the write probe only.

#### F2 — STILL OPEN, re-confirmed live at this HEAD
The operator ruled option 3 (`progress.md` §J.2); the ruling was deliberately not executed on this
card, with reasons I find sound (the audit window had one writer; the SPEC is `completed`; the change
needs its own RED). Re-measured against the binary built from THIS HEAD, one cell corrupted with
`sqlite3`:

```
$ sqlite3 .moai/state/todo/backlog.db "UPDATE items SET landing='not-json' WHERE id='t1';"
$ moai todo               → "invalid character 'o' in literal null (expecting 'u')."   rc=1
$ moai todo landed t1 --clear → same error                                              rc=1   ← the designed escape hatch cannot run
$ moai todo pr            → same error                                                  rc=1
```

Unchanged from the first audit. It is disposed, not closed, and it is the reason Functionality does
not reach its 1.00 anchor: this is a demonstrated edge case, not a hypothetical one.

#### F5 — STILL OPEN (measured, not inferred)
`acceptance.md:126-135` is byte-unchanged in the delta. Both of AC-TLE-010's cases use well-formed
identifiers, so the criterion remains structurally incapable of failing on the F1 class. The only
thing holding F1's fix in place is `TestReadPrimarySpecStatus_RefusesTraversingSpecID`, which is not
an AC. §J.3 states this risk in the same terms and defers it; the deferral is honest and the risk is
real, so the finding stays open.

#### F6 — STILL OPEN, re-measured
```
$ for l in ko en ja zh; do grep -c 'todo landed' docs-site/content/$l/utility-commands/moai-todo.md; done
0  0  0  0
```
Out of the `module:` list; correctly a follow-up card.

#### F7 — STILL OPEN as a guard gap
`diff` between the two `kanban-dispatch.md` surfaces returns rc=0 — byte-identical, stronger than
AC-TLE-021 requires. The gap is unchanged: AC-TLE-021 compares two extracted rows, so nothing would
CATCH future drift on the `[HARD]` paragraph.

#### F8 — STILL OPEN, INHERITED, re-confirmed on both legs
```
$ make agents-emit-check
    golden_test.go:109: .codex/agents/moai/sync-auditor.toml: committed artifact differs from emission (sha256 mismatch)
rc=2
$ git diff --name-only e50964ad3 HEAD | grep -c 'agents/'   →  0
```
Zero agent files in this branch's whole diff against the merge-base, so the drift is not this card's.
The integration lane must still re-measure `make build` on the MERGED tree.

#### F9 — STILL OPEN; the attribution gap is NOT closed
Re-measured at this HEAD: `internal/cli` **80.4%** (identical to the first audit's figure, from my own
run), `internal/kanban` **86.5%** (up from 86.2%). The card's own files remain at or above the bar:

```
todo.go              stmts=225  covered=215  95.6%
todo_landed.go       stmts=85   covered=73   85.9%
todo_pr.go           stmts=83   covered=76   91.6%
landing_evidence.go  stmts=36   covered=34   94.4%
backlog_sqlite.go    stmts=92   covered=79   85.9%
status_read.go       stmts=86   covered=73   84.9%   (was 79.8% — the F1 test)
backlog_migrate.go   stmts=311  covered=244  78.5%   (majority pre-existing)
```

**I did not measure the merge-base baseline**, so the 80.4% is still unattributed — see Gaps. This is
the same gap the first audit left, deliberately not papered over by an argument.

### New findings introduced by the delta

#### N1 — the F4 test's positional assertion is a tautology: it can never fire
- **Severity: Low · Confidence: HIGH (mechanically demonstrated) · blocking: no · Craft**
- **Location:** `internal/cli/todo_landed_test.go`, `TestLandedGitCallsGuardEndOfOptions`

The test carries two assertions. The first — *is `--end-of-options` present?* — is genuine and has an
observed RED (three firings, recorded in `f4-fix-evidence.md` §3). The second is the refinement that
makes presence meaningful:

```go
// Present is not enough — it must precede the operands it guards.
if idx != len(c.argv)-1-countOperandsAfter(c.argv, idx) { … }

func countOperandsAfter(argv []string, idx int) int { return len(argv) - idx - 1 }
```

Substituting: `len-1-(len-idx-1)` = `idx`, so the condition is `idx != idx` — **always false**.
Demonstrated with a standalone program running the exact expression over four argv shapes, including
the one case the assertion exists to catch:

```
argv=[-C /r rev-parse --verify --quiet --end-of-options X^{commit}] idx=5  assertion-fires=false
argv=[-C /r rev-parse --end-of-options --verify --quiet X^{commit}] idx=3  assertion-fires=false
argv=[--end-of-options a b c]                                       idx=0  assertion-fires=false
argv=[a b --end-of-options]                                         idx=2  assertion-fires=false   ← guard AFTER every operand
```

The comment above it states an invariant the code does not check. The consequence is exactly the risk
`f4-fix-evidence.md` §9 names in the abstract — *"a shape assertion can be satisfied by a
wrong-but-present token"* — except that here it is also satisfied by a correctly-spelled token in the
wrong POSITION, which is the same defect the guard exists to prevent.

**Required fix:** assert the position directly, e.g. that every argument after `idx` is an operand the
user supplied, or simply that `idx` is the last flag-position (`idx == len(argv)-1-<expected operand
count>` with the count stated as a literal per subcommand). Then re-establish RED by moving the token.

#### N2 — the new const swallowed `todoGitOutput`'s doc comment
- **Severity: Low · Confidence: HIGH (mechanically demonstrated) · blocking: no · Consistency/Craft**
- **Location:** `internal/cli/todo_landed.go:232-258`

`const gitEndOfOptions` was inserted between `todoGitOutput`'s existing doc comment and the function,
with no blank line, so the two comment blocks merged. `go doc` reports the result:

```
$ go doc -u -all ./internal/cli | grep -A4 'const gitEndOfOptions'
const gitEndOfOptions = "--end-of-options"
    todoGitOutput runs one git command against the queue's own repository
    through the shared process seam, and returns its trimmed stdout.
    `-C <root>` is explicit: the queue resolves against the PRIMARY checkout, …

$ go doc -u -all ./internal/cli | grep -A1 'func todoGitOutput'
func todoGitOutput(args ...string) (string, error)
func todoItemIndex(…)                                    ← no doc comment at all
```

The const's godoc now opens with two paragraphs about a different identifier, and `todoGitOutput` — a
function whose `-C <root>` behaviour the card's own design notes treat as load-bearing — is left
undocumented. This is a Go documentation-convention break, localized, and it is what keeps Consistency
off 1.00 now that F1's sibling-deviation is gone.

**Required fix:** move `const gitEndOfOptions` and its comment above `todoGitOutput`'s doc block (or
below the function), separated by a blank line.

#### N3 — an undeclared git 2.24+ floor now binds the verb
- **Severity: Low · Confidence: HIGH · blocking: no · disclosed by the implementer, not concealed**

`--end-of-options` needs git 2.24 (Nov 2019); the repo declares no git floor anywhere the implementer
searched (`README.md`, `.github/`, `.moai/docs/`, `docs-site/content/ko/` — scope-bounded, not an
exhaustive sweep, and I did not extend it). On a host below 2.24 all three calls now fail where they
previously succeeded. It fails SAFE — the verb refuses to record rather than mis-records — but it
fails, and the floor is invisible to a reader of the project's stated requirements. I record it as a
finding rather than as a note because it is a NEW constraint on the shipped product, introduced after
the SPEC closed and covered by no criterion.

### Non-findings, recorded so they are not re-raised

- **The `internal/cli` suite now takes 615s.** `f4-fix-evidence.md` warned it sits at the 600s
  boundary; my run measured `615.499s` under my own concurrent load, i.e. a `-timeout 600s` invocation
  would have failed with `panic: test timed out` and **zero** `--- FAIL` lines. That is a measurement
  hazard, not a defect: use `-timeout 900s` or higher. It is why my gate used 900s.
- **`internal/kanban` importing `internal/cli/specid`** is the pre-existing arrangement the sibling
  already relies on (the package documents itself as a leaf precisely to permit this). Not a new
  layering violation.
- **28 pre-existing `gofmt` deviations in `internal/cli`** are unmodified at HEAD and predate the card.

---

## Independent Verification

### Claim
At `5350d0e63`, in this tree: both changed packages' suites, `go vet`, and `golangci-lint` are green;
coverage is `internal/kanban` 86.5% and `internal/cli` 80.4%; F1 is closed through the reader (RED at
the pre-fix commit, GREEN at HEAD, end-to-end binary probe); F4 is guarded at all three call sites;
F2 and F5-F9 remain open; and three new findings exist, two of them introduced by the F4 commit.

### Evidence
Every verdict from the unfiltered `$?`. The `internal/cli` suite was captured to a file and the WHOLE
file scanned (17 lines, quoted in full below — nothing sits under a display cut).

```
$ go test ./internal/kanban/... -count=1 -timeout 600s      ; echo rc=$?
ok  	github.com/modu-ai/moai-adk/internal/kanban	143.312s
kanban rc=0
$ /usr/bin/grep -c -- '--- FAIL\|^FAIL' kanban.txt          → 0

$ go test ./internal/cli/... -count=1 -timeout 900s         ; echo rc=$?
cli rc=0
ok  …/internal/cli 615.499s          ok …/cli/agentlint 2.831s     ok …/cli/harness 10.170s
ok  …/cli/pr 2.419s                  ok …/cli/preference 2.946s    ok …/cli/printer 1.645s
ok  …/cli/specid 5.879s              ok …/cli/taskledger 1.713s    ok …/cli/uikit 6.731s
ok  …/cli/update 4.712s              ok …/cli/update/backup 4.217s ok …/cli/update/deploy 3.626s
ok  …/cli/update/merge 5.216s        ok …/cli/update/plan 3.110s   ok …/cli/update/report 5.542s
ok  …/cli/wizard 9.439s              ok …/cli/worktree 10.591s
$ /usr/bin/grep -n -- '--- FAIL\|^FAIL\|timed out\|panic' cli.txt   → (no output; 17 lines total)

$ go vet ./internal/kanban/... ./internal/cli/...           ; echo rc=$?
vet rc=0        (0 bytes of output)

$ golangci-lint run ./internal/kanban/... ./internal/cli/... ; echo rc=$?
0 issues.
lint rc=0

$ go test ./internal/kanban/... -count=1 -cover -coverprofile=…   → coverage: 86.5%   rc=0
$ go test ./internal/cli/     -count=1 -cover -coverprofile=…     → coverage: 80.4%   rc=0
$ go tool cover -func=kanban.cov | grep ReadPrimarySpecStatus     → 75.0%   (was 0.0%)
```

The F1 RED/GREEN pair, the end-to-end traversal probe, the F2 corruption probe, the F4 git-token
measurements, the N1 tautology program, and the N2 `go doc` output are quoted verbatim in their
findings above.

### Baseline-attribution
Every figure was produced in THIS run, in this tree, at HEAD `5350d0e63`, working tree clean at audit
start and carrying only this auditor's own report commits thereafter. The RED leg was produced in a
throwaway `git archive` export of `d739cd051` under the session scratchpad — never in this tree, and
never by mutating a shared file. The probe binary was built from this HEAD (`go build ./cmd/moai`,
rc=0). Nothing is carried from `sync-audit.md`, from `f1-fix-evidence.md` / `f4-fix-evidence.md`, or
from the dispatch: the dispatch's HEAD was re-read and matched, and every figure those documents offer
was treated as a claim, with each one I rely on re-measured here.

### Gaps — what was NOT observed
- **The merge-base coverage baseline is still unmeasured.** `internal/cli` 80.4% is therefore
  attributed to neither this card nor its predecessors. I chose not to spend a second ~550s run plus a
  full pre-fix build on it; F9 states the gap rather than resolving it. Anyone who wants Craft
  attributed must run it at `e50964ad3`.
- **The two suites ran CONCURRENTLY** in this audit (kanban and cli started in one turn). Both were
  rc=0, so no flake was masked, but a timing-sensitive test's margin was smaller than it would be
  serially, and the 615s `internal/cli` figure is a loaded-machine figure.
- **No mutants planted in this tree.** Per the dispatch, implementation and test files were not
  modified. N1's vacuity was demonstrated by an external program running the identical expression, not
  by mutating the test; the F1 RED was obtained in an exported copy rather than by reverting the fix
  here.
- **N1's fix is unverified**, since I did not write one — I demonstrated only that the current
  assertion cannot fire.
- **Cross-platform: none.** darwin/arm64 only. `--end-of-options` behaviour was measured on
  `git 2.50.1 (Apple Git-155)` only, and no git below 2.24 was tested, so N3's floor is documentary,
  not observed failing.
- **No `go test -race`**, no concurrency probe, no full local suite (`CLAUDE.local.md` §4). Packages
  outside `internal/kanban` and `internal/cli` are unmeasured at this HEAD; CI on the pushed head is
  the full-suite verdict.
- **The merged tree is unmeasured.** Every green here is this branch's tip in isolation. `make build`
  is still blocked by the inherited `agents-emit` drift (F8) and must be re-measured after the absorb;
  `catalog.yaml` should still be expected to conflict.
- **F1's blast radius beyond `spec_status`** was not swept, and the WRITE side (`todo next --spec`
  still admits an unvalidated value, `progress.md` §J.4) was verified only as a reported fact, not
  re-derived across every consumer.
- **The `sync-verify/` captures and the two fix-evidence files were not re-derived line by line.** I
  re-ran the gates myself instead, so those files are neither corroborated nor contradicted here,
  except where a figure of theirs is repeated above as my own measurement.

### Residual-risk
- **The verdict is one anchor wide, and the rubric cannot express the improvement.** The delta closed
  the blocking finding, closed two more, and raised two coverage figures — and the score is byte-identical
  at 0.75, because every dimension was already at the same anchor for reasons the fixes did not touch.
  A reader comparing only the two verdict tokens will conclude nothing changed; a great deal did. The
  distance to PASS remains smaller than "FAIL" suggests, and closing it requires Functionality and
  Security to reach 1.00 anchors that demand, respectively, verified edge cases (F2) and zero findings
  of any severity.
- **N1 means the F4 guard's positional property is currently unguarded.** The token is in the right
  place today, by construction of the three edits; nothing would catch a future call site that places
  it wrongly, and the test's comment asserts otherwise. A reader trusting the comment is worse off than
  one who never read it.
- **The `--spec` write side is still open**, so F1's fix is a single-layer defense at the read
  boundary. That is the correct layer and matches the sibling, but the queue can still be made to carry
  a traversal-shaped identifier that every future consumer must independently refuse.
- **Symlink and Windows shapes were not probed.** `ValidateSpecID` is a denylist (absolute, `..`,
  separators). A single-component identifier naming a symlink planted INSIDE `.moai/specs/` would still
  be followed, and Windows-specific shapes (drive-relative, alternate data streams) were not tested.
  Both are inherited from the sibling guard rather than introduced here, and both are outside what this
  delta changed — recorded so the fix is not read as more complete than it is.
- **Absence of further findings is not evidence of their absence.** I audited the delta and
  re-measured the nine prior findings; a defect in the parts of this card the delta did not touch would
  plausibly have escaped, since I did not repeat the first audit's full AC re-verification.

---

## Recommendations

1. **Fix N1** — the tautological assertion — and re-establish its RED by moving the token, since the
   assertion exists precisely to catch a misplaced one. Smallest item on this list and the only one
   that silently weakens a guard the card just added.
2. **Fix N2** — move the const above or below `todoGitOutput`'s doc block. One blank line.
3. **Declare or accept N3** — either state a git floor where the project states its requirements, or
   record the decision that an undeclared 2.24 floor is acceptable because the failure is safe.
4. **F2 and F5 remain the lead's routed follow-ups**, unchanged: `--clear` without a full decode, and
   an AC that puts `spec_id` SHAPE on its test axis (proved non-vacuous by reverting F1 and observing
   AC-TLE-010 actually fail).
5. **F6 (docs-site, four locales) and the write-side `--spec` validation** remain follow-up cards.
6. **Re-measure on the merged tree** — `make build` (F8 is inherited but still blocking there), and the
   `catalog.yaml` conflict resolved by regenerating.
7. **Do not run this suite at `-timeout 600s`.** It measured 615s here.
