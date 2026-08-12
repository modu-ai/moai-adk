# SPEC-CONFIG-SECTION-INTEGRITY-001 — Acceptance Criteria

## §A Discipline

1. **Every AC states a command and its expected observable output.** A criterion phrased as a property with no command is not an AC.
2. **`go test -run <pattern>` exits 0 on zero matches.** Every `-run` AC therefore also requires a verbatim `--- PASS: <exact test name>` line in the output. A `-run` AC whose only assertion is `exit 0` is vacuous and is rejected.
3. **Baselines were recorded from the baseline tree** (code baseline `865cd8aa2` = `origin/main`, main checkout). Each AC carries its observed pre-change baseline so a reviewer can tell a real change from a no-op. The parent SPEC's baseline tree (`d5336214e`, branch `plan/epic-update-config-audit`, worktree `.claude/worktrees/epic-update-config`) produced the original probe evidence this child re-attributes; the child's own ACs re-measure against `865cd8aa2`.
3b. **An already-green criterion is labelled as such.** Where an AC's expected output was already produced before any implementation work (e.g. a pre-existing characterization test), the AC is retained as a **regression guard** with its true measured baseline and an explicit `Class:` line. A regression guard is not evidence of progress; §D counts it separately.
4. **Every new guard needs a falsification** proving it FAILS against unfixed code. §C gives the runnable procedures.
5. **`git stash` is prohibited.** This checkout is shared with concurrent sessions. `git stash push` refuses untracked files without `-u`, and with `-u` it is repository-global and can swallow a parallel session's work. Falsification uses `go test -overlay` or a scratch `git worktree` driven by `go -C`.
6. **The §A probes were each demonstrated by a runnable probe against the parent baseline** (`d5336214e`) and are re-attributed here. Those probes are the regression fixtures below; invented cases are not substituted for them.
7. **All fixtures use `t.TempDir()`** and touch no path outside them (NFR-CSI-001).

## §B Acceptance criteria

### M1 — Malformed-configuration contract

#### AC-CSI-001 — a malformed section is recorded as failed, not as loaded

```bash
go test -run 'TestLoader_MalformedSectionIsRecordedFailed' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestLoader_MalformedSectionIsRecordedFailed` line. The fixture writes a `user.yaml` containing unparseable YAML, calls `Loader.Load`, and asserts the section appears in the failed set and **not** in `LoadedSections()`.

Baseline (re-measured against `865cd8aa2`): the test does not exist. Today `internal/config/loader.go:121-131` (and thirteen siblings) returns early after `slog.Warn`, so the section is simply absent from `loadedSections` — indistinguishable from a section whose file was never present.

#### AC-CSI-002 — an absent section stays silent

```bash
go test -run 'TestLoader_AbsentSectionIsSilent' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestLoader_AbsentSectionIsSilent` line, asserting that a fixture with no `user.yaml` yields compiled defaults, records no failure, and returns no error. This pins REQ-CSI-001 so M1 does not over-correct greenfield projects into errors.

Baseline: the test does not exist; the behaviour today is silent-default and MUST stay so.

#### AC-CSI-003 — a CLI command exits non-zero on a malformed section

```bash
go test -run 'TestCLI_MalformedSectionReturnsError' -count=1 -v ./internal/cli/
```

Expected: a `--- PASS: TestCLI_MalformedSectionReturnsError` line, asserting the returned error names both the file (`user.yaml`) and the parse failure.

Baseline: the test does not exist. Today a malformed `user.yaml` produces a `slog.Warn` and the command proceeds on compiled defaults with exit 0.

#### AC-CSI-004 — a hook stays fail-open but emits an operator-visible advisory

```bash
go test -run 'TestHook_MalformedSectionAdvisesButContinues' -count=1 -v ./internal/hook/
```

Expected: a `--- PASS: TestHook_MalformedSectionAdvisesButContinues` line, asserting the hook returns its allow result **and** that the captured stderr writer contains the section filename. Both halves are required: an assertion on continuation alone does not prove visibility, and an assertion on the advisory alone does not prove fail-open.

Baseline: the test does not exist. Today the only signal is `slog.Warn`, which is not operator-visible in a hook context.

#### AC-CSI-005 — `Save` refuses to persist a section that failed to load

```bash
go test -run 'TestConfigManager_SaveRefusesFailedSection' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestConfigManager_SaveRefusesFailedSection` line. The fixture writes a malformed `user.yaml`, loads, calls `Save()`, asserts an error naming `user.yaml`, and asserts the on-disk bytes of `user.yaml` are **byte-identical to before the `Save()`** — the compiled defaults were not written over the user's file. The byte-identity assertion is the load-bearing half of this AC.

Baseline: the test does not exist. This is the data-loss path: pre-fix, `Save()` serialises the compiled defaults it holds and overwrites the malformed-but-recoverable file.

#### AC-CSI-006 — a clean section is still persisted alongside a failed one

```bash
go test -run 'TestConfigManager_SaveRefusalIsScopedToFailedSection' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestConfigManager_SaveRefusalIsScopedToFailedSection` line, asserting that with `user.yaml` malformed and `language.yaml` clean, a `Save()` that errors on `user.yaml` still wrote the updated `language.yaml`. Pins REQ-CSI-006 so one bad file does not block five good ones.

Baseline: the test does not exist.

### M2 — `.gitignore` merge correctness

#### AC-CSI-007 — a template-dropped pattern is not migrated into the user section

```bash
go test -run 'TestMergeGitignoreFile_TemplateDroppedPatternNotMigrated' -count=1 -v ./internal/cli/update/merge/
```

Expected: a `--- PASS: TestMergeGitignoreFile_TemplateDroppedPatternNotMigrated` line. The fixture is the §A probe: template v1 contains `DS_Store`, the user adds `mysecret.env` and `build/`, template v2 drops `DS_Store`, and the test asserts `DS_Store` does not appear below the `# User Custom Patterns` header after the second merge.

Baseline (re-measured against `865cd8aa2`): the parent probe (against `d5336214e`) produced, after the second merge:

```
node_modules/

# User Custom Patterns (preserved by moai update)
DS_Store
mysecret.env
build/
build/
build/
```

`DS_Store` is present in the user section — that is the failure this AC pins.

#### AC-CSI-008 — duplicates are collapsed

```bash
go test -run 'TestMergeGitignoreFile_DeduplicatesUserPatterns' -count=1 -v ./internal/cli/update/merge/
```

Expected: a `--- PASS: TestMergeGitignoreFile_DeduplicatesUserPatterns` line, asserting exactly one `build/` line in the output when the backup contained three.

Baseline: the same probe output above shows three `build/` lines surviving.

#### AC-CSI-009 — the merge is idempotent

```bash
go test -run 'TestMergeGitignoreFile_IsIdempotent' -count=1 -v ./internal/cli/update/merge/
```

Expected: a `--- PASS: TestMergeGitignoreFile_IsIdempotent` line, asserting that merging a file with its own previous output produces byte-identical content **after the M2 header-parse and dedupe changes**. The new test MUST exercise a case the existing `TestMergeGitignore_DoubleReMerge_Idempotent` does not — specifically a backup carrying a `# User Custom Patterns` header **and** duplicate user lines, so that dedupe and re-merge interact.

Baseline: the named test does not exist, but **the property is already guarded and green** — re-measured against `865cd8aa2`:

```
$ go test -count=1 -v ./internal/cli/update/merge/ 2>&1 | grep -i idempot
--- PASS: TestMergeGitignore_DoubleReMerge_Idempotent (0.01s)
```

`TestMergeGitignore_DoubleReMerge_Idempotent` already asserts double-re-merge idempotence and passes on the baseline tree. AC-CSI-009 is therefore **not** a new-behaviour criterion: its value is that M2's header parse and dedupe must not *break* an invariant that already holds. Class: **regression guard extended into a new case**, not a fresh behavioural assertion.

#### AC-CSI-010 — a pre-header backup keeps its user patterns

```bash
go test -run 'TestMergeGitignoreFile_PreHeaderBackupFallsBackToHeuristic' -count=1 -v ./internal/cli/update/merge/
```

Expected: a `--- PASS: TestMergeGitignoreFile_PreHeaderBackupFallsBackToHeuristic` line, asserting that a backup containing **no** `# User Custom Patterns` header still preserves `mysecret.env`. This is the AC that prevents the fix from becoming a data-loss regression (R8).

Baseline: the test does not exist. Today the heuristic is the only path, so pre-header backups are handled by construction; after M2 the fallback MUST be explicit.

#### AC-CSI-011 — the existing characterization test still passes

```bash
go test -count=1 -v ./internal/cli/update/merge/ 2>&1 | grep -E '^(--- )?(PASS|FAIL|ok)'
```

Expected: no `FAIL` line, and the `--- PASS:` lines from `gitignore_merge_characterization_test.go` still present. Where a characterization test pins the pre-fix behaviour this SPEC deliberately changes, it is updated in the same commit with a comment naming this SPEC — never deleted.

Baseline (re-measured against `865cd8aa2`): `go test -count=1 -v ./internal/cli/update/merge/` → `ok …` with the characterization suite green. The two tests M2 most directly risks are:

```
--- PASS: TestMergeGitignore_ReMergeWithExistingHeader_PatternsSurvive
--- PASS: TestMergeGitignore_DoubleReMerge_Idempotent
```

`TestMergeGitignore_ReMergeWithExistingHeader_PatternsSurvive` is the one M2 must read before editing: it already asserts that pattern lines under a *previous* run's header are re-collected under the new header. M2's header parse changes the classification path those lines take, so this test is the most likely to need a same-commit update with a comment naming this SPEC — never a deletion. Class: **regression guard** (§A clause 3b); falsified by deletion or by silent weakening of the assertion.

### M3 — Fallback visibility

#### AC-CSI-012 — the first 3-way failure is announced

```bash
go test -run 'TestRestore_ThreeWayFailureAdvisesOnFirstOccurrence' -count=1 -v ./internal/cli/update/backup/
```

Expected: a `--- PASS: TestRestore_ThreeWayFailureAdvisesOnFirstOccurrence` line, asserting that a single 3-way merge failure writes an advisory naming the file to the captured stderr writer. **This AC implements REQ-CSI-012, which declares a supersession of `SPEC-V3R6-UPDATE-NOISE-001` `REQ-UN-007`'s silent-first-occurrence clause (see `spec.md` §K).** Landing this AC without the §K reconciliation in place in `spec.md` is prohibited.

Baseline: the test does not exist. At `865cd8aa2`, `internal/cli/update/backup/restore.go:131-134` calls `recordFallback(projectRoot, relPath, false, os.Stderr)` and falls through silently; the noise-suppression ledger stays quiet until three consecutive failures, while the 2-way path at `:139-141` prints `Warning: merge failed for %s, restoring backup` immediately.

#### AC-CSI-013 — an absent base is distinguishable from a merge failure

```bash
go test -run 'TestRestore_AbsentBaseAdvisoryIsDistinct' -count=1 -v ./internal/cli/update/backup/
```

Expected: a `--- PASS: TestRestore_AbsentBaseAdvisoryIsDistinct` line, asserting that the advisory emitted when `os.ReadFile(basePath)` fails differs in text from the merge-failure advisory of AC-CSI-012.

Baseline: the test does not exist. Today this path — `restore.go:118-120`, the `err == nil` guard failing — emits neither a `recordFallback` call nor a warning. It is the quietest of the three paths. Unlike AC-CSI-012, this AC conflicts with nothing: if the §K reversal is rejected at Implementation Kickoff Approval, AC-CSI-012 is withdrawn and this AC still lands (it adds a signal rather than reversing a suppression).

#### AC-CSI-014 — the ledger still suppresses repetition

```bash
go test -run 'TestRestore_LedgerStillSuppressesRepetition' -count=1 -v ./internal/cli/update/backup/
```

Expected: a `--- PASS: TestRestore_LedgerStillSuppressesRepetition` line, asserting that repeated failures for the same file do not produce an unbounded advisory stream. Pins REQ-CSI-014 so M3 does not defeat the noise-suppression mechanism while making the first occurrence visible. The 3-strike threshold advisory (`REQ-UN-008`, which survives the §K supersession per `spec.md` §K.2) still fires on sustained failure.

Baseline: the test does not exist.

## §C Falsification procedures

Each new guard must be shown to FAIL against unfixed code. `git stash` is prohibited (§A clause 5). Two mechanisms are used, both borrowed from `SPEC-UPDATE-REINSTALL-LOOP-002` and `SPEC-UPDATE-DATA-SURVIVAL-001`.

### C.1 — `go test -overlay` for a single-file mutation

Write the mutated source to a scratch path, point an overlay JSON at it, and run the guard.

```bash
SCRATCH=$(mktemp -d)
# 1. copy the file under test and re-introduce the defect
cp internal/config/loader.go "$SCRATCH/loader_mutated.go"
# 2. restore the silent-swallow that M1 removed (re-insert the early return after slog.Warn
#    without recording the failed section)
$EDITOR "$SCRATCH/loader_mutated.go"
# 3. point an overlay at it
cat > "$SCRATCH/overlay.json" <<EOF
{"Replace": {"$(pwd)/internal/config/loader.go": "$SCRATCH/loader_mutated.go"}}
EOF
# 4. the guard must now FAIL
go test -overlay "$SCRATCH/overlay.json" \
  -run 'TestLoader_MalformedSectionIsRecordedFailed' -count=1 -v ./internal/config/
rm -rf "$SCRATCH"
```

Expected: a `--- FAIL: TestLoader_MalformedSectionIsRecordedFailed` line. A `--- PASS` here means the guard does not actually observe the fix and is non-falsifiable.

Applies to: AC-CSI-001, AC-CSI-002 (remove the failed-section record AND break the absent path — the two must stay distinguishable); AC-CSI-003, AC-CSI-004 (remove the caller-context escalation); AC-CSI-005, AC-CSI-006 (remove the `Save` refusal, or make it blanket instead of scoped); AC-CSI-007, AC-CSI-008 (restore the not-in-template heuristic as the only path); AC-CSI-010 (remove the heuristic fallback so a pre-header backup discards patterns); AC-CSI-012, AC-CSI-013 (remove the advisory emissions); AC-CSI-014 (remove the ledger's repetition governance).

### C.2 — a scratch worktree driven by `go -C` for behavioural or multi-file mutations

```bash
WT=$(mktemp -d)/wt
git worktree add --detach "$WT" HEAD
# mutate freely inside $WT — it is not the shared checkout
go -C "$WT" test -run 'TestConfigManager_SaveRefusesFailedSection' -count=1 -v ./internal/config/
git worktree remove --force "$WT"
```

Applies to: AC-CSI-005, AC-CSI-006 (the `Save` refusal interacts with both the loader's failed-section record and the manager's persistence loop — a multi-file mutation); AC-CSI-009 (remove the dedupe + header-parse together and confirm the idempotence guard fails on the new case the pre-existing test does not exercise).

### C.3 — non-Go ACs whose falsification IS the baseline

Two ACs currently produce the opposite of their expected output on this tree, so their baseline is their falsification, verbatim:

| AC | Command | Observed now | Expected after |
|---|---|---|---|
| AC-CSI-007 | `go test -run 'TestMergeGitignoreFile_TemplateDroppedPatternNotMigrated' …` | test does not exist; the §A probe shows `DS_Store` in the user section | `--- PASS:` line, `DS_Store` absent |

**v0.1.0 of the parent listed more ACs in this category than actually qualified.** AC-CSI-009 (idempotence) and AC-CSI-011 (characterization suite) each produce **exactly their expected output** on this tree, so the "baseline is the falsification" claim does not hold for them — passing proves nothing about the implementation. They are reclassified as regression guards and given real, runnable falsifications in §C.1/§C.2.

### C.4 — falsification for already-satisfied regression guards

**AC-CSI-009** is falsified by the §C.2 overlay: remove the dedupe + header-parse together and confirm the new idempotence test (the one exercising header + duplicates together) fails on the mutated tree while the pre-existing `TestMergeGitignore_DoubleReMerge_Idempotent` still passes. Expected: `--- FAIL: TestMergeGitignoreFile_IsIdempotent` on the mutated tree.

**AC-CSI-011** is falsified by deleting (or silently weakening) `TestMergeGitignore_ReMergeWithExistingHeader_PatternsSurvive` and observing the grep half of AC-CSI-011 drop its `--- PASS:` line. The mutation is confined to a scratch worktree (§C.2) and reverted immediately; none uses `git stash`.

## §D Definition of Done

- Every AC in §B ran, and every `-run` AC's output carried its `--- PASS: <exact name>` line.
- Every falsification in §C.1 and §C.2 ran and produced its expected `--- FAIL`; every falsification in §C.4 ran and produced its expected flip, with the mutation reverted afterwards.
- **The two regression guards (AC-CSI-009, AC-CSI-011) are counted separately from the behavioural criteria and are not reported as evidence of progress.** Their §C.4 falsification is what makes them meaningful.
- M1 landed as one revertable commit (NFR-CSI-005); M2 landed as one commit with REQ-CSI-007 (header boundary) AND REQ-CSI-011 (heuristic fallback) together (the R8 mitigation); M3 landed as one commit.
- `spec.md` §K is present and the REQ-CSI-012 reversal was accepted at Implementation Kickoff Approval, **or** REQ-CSI-012/AC-CSI-012 were withdrawn together and M3 landed REQ-CSI-013/014 only.
- The M3 commit body states the `REQ-UN-007` supersession and references `spec.md` §K (NFR-CSI-006).
- `go test ./... -count=1` green; `golangci-lint run` clean.
- No diff under `internal/template/templates/`.
- No falsification probe file (`zzz_*_probe*.go`) remains in the tree.
- `progress.md` §E.2 and §E.3 populated by `manager-develop` with commit SHAs and verbatim command output.
