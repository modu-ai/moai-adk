# M4 evidence — the read surfaces (SPEC-TODO-LANDING-EVIDENCE-001, card t359)

Verbatim command + output pairs for Milestone M4. Closes with a known-losses
section (§9); material named there was NOT measured and MUST NOT be cited later
as a verdict basis.

Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359`
Branch: `WT-landing-evidence`

---

## 1. Working tree and baseline, read at M4 entry

```
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359
$ git rev-parse HEAD
5edba757cc0fb4b6e6ee9b8bbac53fa4e213fb21
$ git branch --show-current
WT-landing-evidence
$ git status --short
(empty)
$ git log --oneline -5
5edba757c docs(t359): M3 evidence and progress §E.2 — the recording verb
5c21cd96a feat(kanban): moai todo landed — the recording verb (t359, M3)
705838c2a docs(SPEC-TODO-LANDING-EVIDENCE-001): close M2's two RED gaps (t359)
751c2ab08 docs(SPEC-TODO-LANDING-EVIDENCE-001): three-value M2 baseline attribution (t359)
7769bbf91 feat(kanban): landing-evidence record type and the attribution boundary (t359)
```

My own reading of HEAD at M4 entry — `5edba757c` — matches the value the
dispatch quoted as a comparison. Tree clean; single writer confirmed.

## 2. Pre-edit baseline (attributable, this tree, this run)

```
$ go test ./internal/kanban/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/kanban	136.861s

$ go vet ./internal/kanban/... ./internal/cli/...
(no output)
vet exit=0

$ gofmt -l internal/kanban/ internal/cli/ | wc -l
      28
```

The 28 gofmt files are PRE-EXISTING and none is an M4 deliverable. Named so a
later reader cannot read the same count after M4 as an M4 regression:

```
internal/cli/chain_test.go               internal/cli/clean_home_test.go
internal/cli/doctor_disk_test.go         internal/cli/epic.go
internal/cli/glm_env_parity_test.go      internal/cli/hook_install_precommit_disclosure_test.go
internal/cli/init_audit_test.go          internal/cli/init_workflow_flags_test.go
internal/cli/launcher_worktree_l2_test.go internal/cli/preference/decay.go
internal/cli/preference/entry.go         internal/cli/preference/m4_crash_repro_test.go
internal/cli/schema_bridge.go            internal/cli/session_worktree_m7_test.go
internal/cli/session_worktree_test.go    internal/cli/tuxiu_characterization_test.go
internal/cli/uikit/banner_test.go        internal/cli/update/backup/snapshot_test.go
internal/cli/update/class_test.go        internal/cli/update/deploy/deploy_test.go
internal/cli/update/plan/plan_test.go    internal/cli/update/preview_test.go
internal/cli/update_characterization_test.go internal/cli/update_hygiene_characterization_test.go
internal/cli/update_namespace_protect.go internal/cli/update_version_test.go
internal/cli/web.go                      internal/cli/wizard/mcp_audit_test.go
```

The `internal/cli` suite was also green pre-edit (`go test ./internal/cli/...
-count=1 -timeout 600s`, all packages `ok`). See §9 for the limit on how that
particular pre-edit reading was inspected.

## 3. RED — the four behavioural criteria, against the pre-change render

The marker unit test was temporarily split out so the four criterion-bearing
tests would produce BEHAVIOURAL reds rather than a compile failure (the helper
and the constant they need did not exist yet). It was restored before GREEN.

```
$ go test ./internal/cli/ -count=1 -run 'TestTodoPR_SevenColumnsCardTextLast|TestTodoPR_AssertionVersusObservationSurvivesSHASubstitution|TestTodoPR_NoRecordRendersEmptyEvidenceCell|TestTodoPR_ProjectRootUnchangedWithEvidence' -timeout 600s
--- FAIL: TestTodoPR_SevenColumnsCardTextLast (0.28s)
    todo_pr_landing_test.go:139: row t1 has 6 fields, want exactly 7: "t1\tlinked\t#1614\tinferred\tqueued\tlinked card"
--- FAIL: TestTodoPR_AssertionVersusObservationSurvivesSHASubstitution (1.00s)
    --- FAIL: TestTodoPR_AssertionVersusObservationSurvivesSHASubstitution/distinct_SHAs (0.51s)
        todo_pr_landing_test.go:242: asserted cell = "asserted card", want the "operator" marker
        todo_pr_landing_test.go:245: observed cell = "observed card", want the "ref-head" marker
        todo_pr_landing_test.go:268: observed card has no decodable landing record: [{"card_id":"t1","outcome":"landed"},{"card_id":"t2","outcome":"landed"}]
    --- FAIL: TestTodoPR_AssertionVersusObservationSurvivesSHASubstitution/ref_head_substituted_with_the_asserted_SHA (0.49s)
        todo_pr_landing_test.go:242: asserted cell = "asserted card", want the "operator" marker
        todo_pr_landing_test.go:245: observed cell = "observed card", want the "ref-head" marker
        todo_pr_landing_test.go:268: observed card has no decodable landing record: [{"card_id":"t1","outcome":"landed"},{"card_id":"t2","outcome":"landed"}]
--- FAIL: TestTodoPR_NoRecordRendersEmptyEvidenceCell (0.27s)
    todo_pr_landing_test.go:307: row t1 has 6 fields, want 7: "t1\tlanded\t\t\tqueued\tlanded but unrecorded"
--- FAIL: TestTodoPR_ProjectRootUnchangedWithEvidence (0.08s)
    todo_pr_landing_test.go:359: the hashed set does not contain the queue database under .moai/state/kanban/; the assertion is not covering the subject:
        .moai/state/todo/backlog.db d38e99d75d820b35dddcaa01027e4a586a2bbbaa49936365e48984838710f190
        .moai/state/todo/backlog.lock 2a2b22aceae5e678a2756e916b69f4dd62fbd8c01005dad93b865645fd195c0c
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.637s
FAIL
```

### 3.1 The AC-TLE-015 red doubles as the pre-change capture

AC-TLE-015's Given requires "the pre-change six-field rendering of the same
fixture captured as the expected prefix". The RED above IS that capture, and it
is stronger than a transcription would be: the row printed by the pre-change
code is

```
t1\tlinked\t#1614\tinferred\tqueued\tlinked card
```

— fields 1-5 exactly the values the test pins (`t1`, `linked`, `#1614`,
`inferred`, `queued`), while ONLY the field-count assertion fired. Fields 1-5
were therefore measured equal to their pre-change values against the pre-change
binary, rather than asserted against a number written down by hand.

### 3.2 Which assertion fired, and at which line

| Test | Line | Assertion that fired | What it establishes |
|---|---|---|---|
| `TestTodoPR_SevenColumnsCardTextLast` | 139 | field count 6 ≠ 7 | the column is genuinely absent pre-change; fields 1-5 already correct |
| `..._SurvivesSHASubstitution` (both subtests) | 242, 245 | cell = the CARD TEXT, carries no marker | index 5 was the text pre-change — the substitution property could not even be posed |
| `..._SurvivesSHASubstitution` (both subtests) | 268 | `--json` object has no `landing` key | the JSON conjunct of AC-TLE-013's render half is genuinely unbuilt |
| `TestTodoPR_NoRecordRendersEmptyEvidenceCell` | 307 | field count 6 ≠ 7 | AC-TLE-006's render conjunct is unbuilt |
| `TestTodoPR_ProjectRootUnchangedWithEvidence` | 359 | positive control: hashed set lacks the queue dir | see §4 — this red is a DEFECT IN THE CRITERION, not in the code |

Note the AC-TLE-014 red is at line 359, the **positive control**, NOT the
byte-identity assertion at 387. The equality clause was never reached in this
run, so this RED says nothing about whether `todo pr` writes. §6 is where that
is established.

## 4. FINDING — AC-TLE-014's containment clause names a directory that no longer exists

AC-TLE-014 (`acceptance.md:180`) requires the hashed set to contain "at least
the queue database under `.moai/state/kanban/`". The measured path is
`.moai/state/todo/backlog.db`.

`.moai/state/kanban/` is the **pre-rename legacy** name:

```
$ sed -n '36,43p' internal/kanban/state_dir.go
// stateDirName is the project-local state directory `moai todo` owns.
const stateDirName = "todo"

// legacyStateDirName is the name it carried before this SPEC. It appears in
// exactly two places — here and the fallback reader below — which is what
// makes the literal-cleanliness sweep (REQ-TOSQ-018) meaningful: any OTHER
// occurrence in production Go is a consumer that was missed.
const legacyStateDirName = "kanban"
```

`SPEC-TODO-SQLITE-001` REQ-TOSQ-015 renamed the directory; the file's own header
records that `LegacyStateDirForRoot` is "the fallback reader's subject and the
relocation's source; **nothing writes through it**".

**This is the same shape M3 found in AC-TLE-012's containment clause** — a
criterion probing a value the implementation cannot produce — but in the
opposite, louder direction: M3's was vacuous-PASSING (the clause could never
fail), this one is vacuous-FAILING (the clause could never pass), which is why
it surfaced on the first run instead of hiding inside a green.

**Repair, applied:** the clause derives the directory from
`kanban.StateDirForRoot(root)` rather than transcribing a literal, so it stays
correct across the next rename too. `acceptance.md` itself was NOT edited — that
file is manager-spec's, not this milestone's (see §8).

Re-observed after the repair, with the criterion still unbuilt in the code: the
positive control passes and the run proceeds to the byte-identity assertion —
see §6, where the planted mutant reaches line 387.

## 5. GREEN

```
$ go test ./internal/cli/ -count=1 -run 'TestTodoPR_SevenColumnsCardTextLast|TestTodoPR_AssertionVersusObservationSurvivesSHASubstitution|TestTodoPR_NoRecordRendersEmptyEvidenceCell|TestTodoPR_ProjectRootUnchangedWithEvidence|TestFormatLandingEvidence_MarkersAreDisjoint' -timeout 600s -v
=== RUN   TestTodoPR_SevenColumnsCardTextLast
--- PASS: TestTodoPR_SevenColumnsCardTextLast (0.49s)
=== RUN   TestTodoPR_AssertionVersusObservationSurvivesSHASubstitution
=== RUN   TestTodoPR_AssertionVersusObservationSurvivesSHASubstitution/distinct_SHAs
=== RUN   TestTodoPR_AssertionVersusObservationSurvivesSHASubstitution/ref_head_substituted_with_the_asserted_SHA
--- PASS: TestTodoPR_AssertionVersusObservationSurvivesSHASubstitution (1.00s)
=== RUN   TestTodoPR_NoRecordRendersEmptyEvidenceCell
--- PASS: TestTodoPR_NoRecordRendersEmptyEvidenceCell (0.49s)
=== RUN   TestTodoPR_ProjectRootUnchangedWithEvidence
--- PASS: TestTodoPR_ProjectRootUnchangedWithEvidence (0.49s)
=== RUN   TestFormatLandingEvidence_MarkersAreDisjoint
--- PASS: TestFormatLandingEvidence_MarkersAreDisjoint (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	3.474s
```

### 5.1 The emitted row, verbatim

```
$ go test ./internal/cli/ -count=1 -run 'TestTodoPR_SevenColumnsCardTextLast' -timeout 600s -v
=== RUN   TestTodoPR_SevenColumnsCardTextLast
    todo_pr_landing_test.go:121: rendered rows (tabs shown as \t):
        t1\tlinked\t#1614\tinferred\tqueued\t\tlinked card
        t2\tlanded\t\t\tpicked\tlanded@origin/develop:c9f7122(operator)\tlanded card
        t3\tno-link\t\t\tqueued\t\tuntouched card
--- PASS: TestTodoPR_SevenColumnsCardTextLast (0.48s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.390s
```

Seven fields; card text last; the evidence cell empty for the two cards with no
record; `landed@origin/develop:c9f7122(operator)` byte-identical to the form
`design.md` §4 specifies.

## 6. AC-TLE-014 — the planted mutant

### 6.1 Hashes before

```
$ shasum internal/cli/todo_pr.go internal/cli/todo_pr_landing_test.go
dba5b4b8fb2f80471a0e3c1ee030fadf07bfb87e  internal/cli/todo_pr.go
3d358b00bf69e82d947ece3ed13ffd27ba4af81a  internal/cli/todo_pr_landing_test.go
```

### 6.2 The mutant applied

Inserted into `runTodoPR`, immediately after the queue load and before
`fetchOpenPRs`, plus the `os` / `path/filepath` imports it needs:

```go
	// MUTANT (AC-TLE-014): a cache written from the todo pr path, planted
	// OUTSIDE the queue directory and outside .git/ — the exact placement a
	// queue-scoped assertion would miss.
	{
		cacheDir := filepath.Join(resolveTodoQueueRoot(), ".moai", "cache")
		_ = os.MkdirAll(cacheDir, 0o755)
		_ = os.WriteFile(filepath.Join(cacheDir, "todo-pr.cache"), []byte("planted"), 0o644)
	}
```

### 6.3 The RED, verbatim

```
$ go test ./internal/cli/ -count=1 -run 'TestTodoPR_ProjectRootUnchangedWithEvidence' -timeout 600s
--- FAIL: TestTodoPR_ProjectRootUnchangedWithEvidence (0.56s)
    todo_pr_landing_test.go:387: the project root changed across `todo pr`
        before:
        .moai/state/todo/backlog.db abe7b0d64c53a953e3ffe3bdd5f909a00ab46b6bcb277b190596db47c08f80b5
        .moai/state/todo/backlog.lock 40a138d216de0ee81bda0cce4101455628f33033c4125c924f0fce020606f738
        after:
        .moai/cache/todo-pr.cache 372eb3774802a8d97badd0f3afdabbf7a17ef8a1f900b394db99a460b893406c
        .moai/state/todo/backlog.db abe7b0d64c53a953e3ffe3bdd5f909a00ab46b6bcb277b190596db47c08f80b5
        .moai/state/todo/backlog.lock 40a138d216de0ee81bda0cce4101455628f33033c4125c924f0fce020606f738
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.467s
FAIL
```

**Which assertion fired, at which line: the byte-identity assertion at
`todo_pr_landing_test.go:387`** — not the positive control at 359, which passed
this time and let the run reach the clause that matters. The diff names the
planted path explicitly: `.moai/cache/todo-pr.cache`, which is outside the queue
directory (`.moai/state/todo/`) and outside `.git/`. The mutant the criterion
names is the mutant that was planted, and the assertion caught it at the
position the criterion specifies.

### 6.4 What this assertion would still miss

Asked per the M2/M3 reading discipline. Two things:

1. **A write outside the fixture root** — `$HOME`, `/tmp`, a global cache. This
   is an APPROVED RESIDUAL inherited from half A and stated in the criterion's
   own scope note: the assertion is "the verb's reach within the project", NOT
   "this verb writes nowhere". Unchanged by M4, not closed here.
2. **A write inside `.git/`.** Carved out deliberately (a real flake source),
   and the carve-out is exactly the subtree `todo pr`'s `git log` subprocess is
   entitled to touch. A mutant planting there would survive; the positive
   control is what stops the carve-out being widened into vacuity.

Neither is a gap the plant could have closed — both are boundaries the criterion
draws on purpose.

### 6.5 Revert, and the proof no mutant survived

```
$ shasum internal/cli/todo_pr.go internal/cli/todo_pr_landing_test.go
dba5b4b8fb2f80471a0e3c1ee030fadf07bfb87e  internal/cli/todo_pr.go
3d358b00bf69e82d947ece3ed13ffd27ba4af81a  internal/cli/todo_pr_landing_test.go

$ grep -c "MUTANT\|todo-pr.cache" internal/cli/todo_pr.go
0
```

Byte-identical to §6.1 on both files, and zero textual residue. No mutant
survived.

## 7. FINDING — an inherited six-column guard, and a truncated read that hid it

`TestTodoPR_RowCarriesQueueState` (`internal/cli/todo_landing_test.go:101`) is
half A's own AC-TLS-010 criterion and pins the SIX-column contract exactly. The
seventh column breaks it by construction — this is the contract change
AC-TLE-015 mandates, landing on the previous contract change's guard.

```
$ go test ./internal/cli/ -count=1 -run 'TestTodoPR_RowCarriesQueueState' -timeout 600s
--- FAIL: TestTodoPR_RowCarriesQueueState (0.26s)
    todo_landing_test.go:125: row [t1 no-link   picked  picked but no commits] has 7 columns, want 6 (the state column was added)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.251s
FAIL
```

**Repair:** the expected count is bumped 6 → 7 and the text index 5 → 6, as a
visible act in the same change that adds the column — the discipline M1 used for
the schema-freeze guard. It was NOT loosened to a `>=` lower bound: this guard's
value is that it fails on ANY count change, and a lower bound would stop
catching the next one. The test's substance is untouched — state still asserted
at index 4, text still asserted as the last field.

**How this was nearly missed, recorded because the near-miss is the reusable
part.** My first post-implementation `internal/cli` run was piped through
`grep -v "^ok" | head -20`. The package emits a large volume of `INFO`/`WARN`
log lines on stdout, so the 20-line window filled with log noise and the
`--- FAIL` line sat below it. I read that truncated window as green. The failure
was caught only when a later run captured the exit code:

```
$ go test ./internal/cli/... -count=1 -timeout 600s >/tmp/t359-cli.log 2>&1; echo "cli exit=$?"
cli exit=1
```

A `head`-bounded window over a noisy suite is not a verdict — the exit code is.
The §8 run below uses the exit-code form.

## 8. Post-repair verification (the M4 verdict basis)

```
$ go test ./internal/cli/... -count=1 -timeout 600s >/tmp/t359-cli2.log 2>&1; echo "cli exit=$?"
cli exit=0
$ grep -E "^--- FAIL|^FAIL" /tmp/t359-cli2.log
(no output)
$ grep -c '^ok' /tmp/t359-cli2.log
17
$ grep "^ok.*internal/cli\s" /tmp/t359-cli2.log
ok  	github.com/modu-ai/moai-adk/internal/cli	412.816s

$ go test ./internal/kanban/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/kanban	137.054s

$ go vet ./internal/kanban/... ./internal/cli/...
(no output)
vet exit=0

$ gofmt -l internal/cli/todo_pr.go internal/cli/todo_pr_landing_test.go internal/cli/todo_landing_test.go
(no output)

$ gofmt -l internal/kanban/ internal/cli/ | wc -l
      28
```

The whole-tree gofmt count is 28 both before (§2) and after — unchanged, and the
28 are the pre-existing files enumerated in §2. No M4 file is among them.

The `internal/kanban` run is included although M4 edits nothing there: the render
consumes `LandingEvidence.Marker()` and `Validate()`, so a change in this
milestone's reading of that seam would surface in the owning package's suite.

## 9. Known losses — NOT a verdict basis

Named per the evidence-bearing report contract. Nothing below was measured; none
of it may be cited later as establishing anything.

1. **The `internal/cli` PRE-EDIT green (§2) was read through a filtered,
   `head`-bounded window**, the same shape §7 shows to be unreliable. Its exit
   code was not captured. The pre-edit `internal/cli` state is therefore a
   known loss, not a measured baseline. It does not weaken the M4 verdict — §8
   is exit-code-based and green — but it means "the suite was green before M4"
   is unattributed for that package. `internal/kanban`'s pre-edit green (§2) IS
   attributable: it printed a single `ok` line with no filtering.
2. **The malformed-marker branch was never exercised through a queue fixture.**
   It is asserted only at the render helper
   (`TestFormatLandingEvidence_MarkersAreDisjoint`), because the store's read
   path (`backlog_migrate.go:87-92`) refuses an undecodable stored value before
   any render runs. That the marker renders correctly *from the store* is
   UNMEASURED. It is reachable only if M3's readRecord-errors decision is
   overturned; that decision is under separate review and is not this
   milestone's.
3. **No cross-platform build was run.** `GOOS=windows` / `GOOS=linux` builds of
   the touched packages were not attempted. M4 introduces one platform-sensitive
   construct — `filepath.Separator` in the test's queue-directory prefix — and
   it is unverified off darwin.
4. **`golangci-lint` was not run.** Only `go vet` and `gofmt`. The project
   linter's verdict on the M4 files is unobserved.
5. **The evidence cell was never rendered with a real, non-fixture record.** All
   evidence in this file comes from fixtures whose `Ref` / `RefHead` /
   `ObservedAt` were hand-authored, never produced by `moai todo landed` against
   a live repository. That the two verbs agree end-to-end — the ref M3 resolves
   being the ref M4 renders — is unmeasured, and is neither AC's clause.
6. **No coverage measurement.** `go test -cover` was not run on either package,
   so M4's effect on the coverage figure is unknown.
7. **The ast-grep tooling noise the dispatch warned of did not recur.** No
   `sql-injection` or `go-error-ignored-blank` finding was produced against M4's
   files, because that tooling was not invoked in this milestone. Absence of a
   finding here is absence of a run, not a clean result.
