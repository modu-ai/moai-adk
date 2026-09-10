# t359 — F1 fix evidence

Fixes the ONE blocking finding of `.moai/reports/t359/sync-audit.md` (F1: `ReadPrimarySpecStatus`
joins an unvalidated `spec_id`). F2-F9 are non-blocking and are NOT touched here.

- **Tree:** `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359`
- **Branch:** `WT-landing-evidence`
- **HEAD at start of this work:** `d739cd051` (re-read in-tree; matches the dispatched value)
- **Files:** `internal/kanban/status_read.go`, `internal/kanban/status_read_test.go`

---

## 1. The RED, observed against the UNFIXED code

The test drives the traversal end-to-end through `ReadPrimarySpecStatus` with a file planted
outside `primaryRoot` — it does not assert against `ValidateSpecID`, because testing the sanitizer
tests the sanitizer while the defect is that the reader never called it.

Two fixture preconditions run before the assertion, so a later refusal cannot be mistaken for a
fixture that simply pointed at nothing: the escaped path must `Stat` successfully, and it must not
be prefixed by `root + separator`.

```
$ go test ./internal/kanban/ -run TestReadPrimarySpecStatus_RefusesTraversingSpecID -count=1 -v
=== RUN   TestReadPrimarySpecStatus_RefusesTraversingSpecID
=== PAUSE TestReadPrimarySpecStatus_RefusesTraversingSpecID
=== CONT  TestReadPrimarySpecStatus_RefusesTraversingSpecID
    status_read_test.go:419: ReadPrimarySpecStatus(root, "../../../outside") = ("PWNED-TRAVERSAL", true) — a traversing spec id read a document outside the project root; want ('', false) so the caller maps it onto the unknown marker
--- FAIL: TestReadPrimarySpecStatus_RefusesTraversingSpecID (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	0.487s
FAIL
rc=1
```

**Firing assertion:** `status_read_test.go:419` — the `if status, ok := ReadPrimarySpecStatus(root,
traversing); ok` branch. The observed return was `("PWNED-TRAVERSAL", true)`: the read escaped the
project root and returned a foreign document's status.

The test carries a second leg asserting a legitimate id (`SPEC-NAV-X`) still returns
`(StatusCompleted, true)`. Without it, a guard that refused every id would satisfy the traversal leg
alone — the vacuous-pass shape.

## 2. The fix

`specid.ValidateSpecID(specID)` before the join, returning the existing `("", false)` not-read
signal on rejection, so `internal/cli/todo_landed.go:300` maps it onto `LandingSpecStatusUnknown` —
the marker that already means "asked, unanswerable". No new marker was invented.

Shape and comment density match the sibling guard in the same file (`ReadCardStatus`,
`status_read.go:106-111`), which guards the identical value class; a second validation idiom in one
file is how the next divergence starts.

**`#nosec G304` — kept, justification rewritten.** gosec still flags the call (the path is still
built from a variable), so removing the annotation would reintroduce a finding. What was wrong was
not its presence but its claim: `project-local SPEC path` asserted an invariant the code did not
enforce. It now reads `specID validated above; path is project-local`, which is true after the
guard and names what makes it true. Verified by the `golangci-lint` run in §3 (rc=0, 0 issues).

The function's doc comment enumerates the causes of `ok=false`; a fourth cause now exists, so it is
listed there — otherwise the doc would have become false.

```
 	if strings.TrimSpace(primaryRoot) == "" || strings.TrimSpace(specID) == "" {
 		return "", false
 	}
+	// The spec identifier is interpolated into the primary-checkout spec.md
+	// path, and `moai todo next --spec` records the operator's value verbatim
+	// by design — so a traversal-shaped value (`..`, separator, absolute) must
+	// be refused here rather than reach the join, exactly as ReadCardStatus
+	// refuses it above. A rejected id is the ok=false outcome, not a plausible
+	// default: an unreadable identifier is unanswerable, which is what the
+	// caller's LandingSpecStatusUnknown marker already means.
+	if err := specid.ValidateSpecID(specID); err != nil {
+		return "", false
+	}
-	raw, err := os.ReadFile(filepath.Join(...)) // #nosec G304 -- project-local SPEC path
+	raw, err := os.ReadFile(filepath.Join(...)) // #nosec G304 -- specID validated above; path is project-local
```

## 3. The GREEN and the gates

Targeted, same command that produced the RED:

```
$ go test ./internal/kanban/ -run TestReadPrimarySpecStatus_RefusesTraversingSpecID -count=1 -v
--- PASS: TestReadPrimarySpecStatus_RefusesTraversingSpecID (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.434s
rc=0
```

Every verdict below is the unfiltered `$?` of the command, not a filtered view. `internal/cli` was
captured to a file and then scanned for `FAIL` across the whole output rather than a bounded head —
on this card a `head`-bounded window once showed green while a `--- FAIL:` line sat below the cut.

| Gate | Command | Result |
|---|---|---|
| kanban tests | `go test ./internal/kanban/... -count=1` | `ok github.com/modu-ai/moai-adk/internal/kanban 136.703s` — **rc=0** |
| cli tests | `go test ./internal/cli/... -count=1 -timeout 600s` | **rc=0**; 17 `ok`/`no test files` lines; full-output `FAIL` scan returned zero matches |
| vet | `go vet ./internal/kanban/... ./internal/cli/...` | no output — **rc=0** |
| gofmt | `gofmt -l internal/kanban/` | no output (no files listed) |
| lint | `golangci-lint run ./internal/kanban/...` | `0 issues.` — **rc=0** |

`internal/cli` is in scope because it exercises `ReadPrimarySpecStatus` through the real path
(`internal/cli/todo_landed.go:300`); the cross-package attribution is why the function can look
uncovered when it is not.

## 4. Package sweep — other unvalidated path constructions

Asked as a report item, not a fix. Scanned `internal/kanban` for `filepath.Join` / `os.ReadFile` /
`os.Stat` / `exec.Command` reached from a caller-supplied identifier.

**Inside `internal/kanban`: none remaining.** Both identifier classes are guarded at every
filesystem boundary:

- `specID` → `specid.ValidateSpecID` at `status_read.go:109` (`ReadCardStatus`),
  `status_read.go:293` (this fix), `board_store.go:268`, `board_store.go:360`.
- `sessionID` → `validateSessionID` at `record.go:202` (`Write`), `record.go:237` (read),
  `role.go:74`, `role.go:115`. It rejects `.`, `..`, and both separators; a value with no separator
  cannot traverse a single `filepath.Join`, so the different shape is adequate rather than a gap.

**One unguarded call site, outside the package:** `internal/hook/session_start_record.go:79` passes
hook-stdin `input.SessionID` to `kanban.RecordPath` with no prior `validateSessionID`. It reaches
only `os.Stat` — an existence probe, no read and no disclosure — so this is a note, not a second
instance of F1. Reported, not fixed: it is outside this card's declared scope.

## 5. Known losses — NOT measured

Named here so none of it is later cited as a verdict basis.

- **No cross-platform verification.** Every run above is darwin/arm64. The traversal fixture uses
  `filepath.Join`, so it is separator-portable in principle, but no `GOOS=windows` build or run was
  performed, and the test carries no `runtimeIsWindows()` skip (unlike its git-plumbing neighbours
  in the same file) because it touches no git — untested rather than known-good.
- **No full-suite run.** Only `./internal/kanban/...` and `./internal/cli/...`. Per the repo rule a
  local `go test ./...` is not run; the full-suite verdict is CI's on the pushed head, and no push
  has happened from this worker.
- **No coverage measurement.** Neither before nor after; the `-cover` figure for this function is
  unknown to me and the auditor's 0.0% attribution note was not re-derived.
- **No lint on `internal/cli`.** `golangci-lint` was run on `./internal/kanban/...` only — the
  package I edited. `internal/cli` was tested and vetted but not linted in this run.
- **The auditor's `d739cd051` lint baseline was not re-verified as a baseline;** I ran lint fresh on
  the fixed tree instead, which is the stronger evidence for this change but says nothing about
  whether the earlier figure was accurate.
- **No end-to-end `moai todo landed` run.** The fix is verified at the function boundary and through
  the package test suites, not by driving the CLI verb against a real queue with a traversing
  `--spec`. The auditor did that against the unfixed tree; I did not repeat it against the fixed one.

## 6. Residual risk

- **The admission boundary is still open by design.** `moai todo next --spec` continues to record
  any string verbatim (`internal/cli/todo.go:584`), so a traversal-shaped `spec_id` can still be
  *stored* in the queue — it just no longer *reads* a foreign file. Any other consumer that joins
  `spec_id` into a path without validating would reintroduce the same class. Validating at
  admission has a wider blast radius and is the operator's call, per the audit.
- **`LandingEvidence.Validate` still does not constrain `SpecStatus`** (`landing_evidence.go:146`).
  This fix removes the foreign-file source of a bad value; it does not add a field invariant, so a
  bad value arriving by some other route would still pass validation.
- **The guard is exact-match on shape, not on existence.** A syntactically clean id naming a SPEC
  that does not exist still returns `("", false)` — correct, but it means the unknown marker now
  covers four distinct causes that the caller cannot tell apart.
- **`ValidateSpecID` accepts embedded `..`-free oddities** (e.g. a leading `.`), which cannot
  traverse but can name a hidden directory inside `.moai/specs/`. In-root and therefore not an
  escape, but not a canonical-form check either.
