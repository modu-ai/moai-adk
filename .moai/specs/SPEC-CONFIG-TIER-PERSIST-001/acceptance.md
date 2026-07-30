# SPEC-CONFIG-TIER-PERSIST-001 — Acceptance Criteria

## §A Discipline

1. **Every AC states a command and its expected observable output.** A criterion phrased as a
   property with no command is not an AC.
2. **`go test -run <pattern>` exits 0 on zero matches.** Every `-run` AC therefore also requires a
   verbatim `--- PASS: <exact test name>` line in the output. A `-run` AC whose only assertion is
   `exit 0` is vacuous and is rejected.
3. **Baselines were recorded from this tree while authoring** (HEAD `d5336214e`, branch `plan/epic-update-config-audit` (merged with `origin/main`)).
   Each AC carries its observed pre-change baseline so a reviewer can tell a real change from a
   no-op.
4. **Every new guard needs a falsification** proving it FAILS against unfixed code. §C gives the
   runnable procedures.
5. **`git stash` is prohibited.** This checkout is shared with concurrent sessions — during the
   authoring of this SPEC a concurrent actor `chmod`-ed `.moai/config/sections/report.yaml` between
   two commands. `git stash push` refuses untracked files without `-u`, and with `-u` it is
   repository-global and can swallow a parallel session's work. Falsification uses
   `go test -overlay` or a scratch `git worktree` driven by `go -C`.
6. **F1 through F8 were each demonstrated by a runnable probe.** Those probes are the regression
   fixtures below; invented cases are not substituted for them.
7. **All fixtures use `t.TempDir()`** and touch no path outside it (NFR-CTP-001).

## §B Acceptance criteria

### M1 — Tier merge semantics

#### AC-CTP-001 — an explicit `false` from a higher tier wins

```bash
go test -run 'TestMergeAll_ExplicitFalseWins' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestMergeAll_ExplicitFalseWins` line. The test is the §A/F2 probe promoted
to a fixture: it layers `SrcProject{"workflow.branch_guard.enabled": false, "x.count": 0,
"x.name": ""}` over `SrcBuiltin{...: true, 42, "default"}` and asserts each merged value equals the
project value with `Provenance.Source == SrcProject`.

Baseline: the test does not exist —
`go test -run 'TestMergeAll_ExplicitFalseWins' ./internal/config/` prints
`ok github.com/modu-ai/moai-adk/internal/config 0.246s [no tests to run]` and exits 0, which is
exactly the vacuity clause 2 excludes. The probe that produced the pre-fix behaviour observed:

```
workflow.branch_guard.enabled    => true ok=true source=builtin
x.count                          => 42 ok=true source=builtin
x.name                           => "default" ok=true source=builtin
```

#### AC-CTP-002 — absence and falsey-presence remain distinguishable

```bash
go test -run 'TestMergeAll_AbsentKeyOmittedFalseyKeyPresent' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestMergeAll_AbsentKeyOmittedFalseyKeyPresent` line, asserting that
`Get("never.set")` returns `ok=false` while `Get("explicitly.false")` returns `ok=true` with value
`false`. This pins REQ-CTP-004 so the fix does not over-correct into emitting every absent key.

Baseline: the test does not exist. Pre-fix, both keys are indistinguishable — the falsey key is
skipped and never reaches `result.Set`, so both return `ok=false`.

#### AC-CTP-003 — strict-mode policy rejection is reachable for a falsey policy value

```bash
go test -run 'TestMergeAll_StrictModeRejectsOverrideOfFalseyPolicy' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestMergeAll_StrictModeRejectsOverrideOfFalseyPolicy` line, asserting
`MergeAll` returns a `*PolicyOverrideRejected` when `strict_mode` is on, `SrcPolicy` supplies
`false` for a policy-designated key, and `SrcProject` supplies `true`.

Baseline: pre-fix, the falsey policy value is skipped at `internal/config/merge.go:149-152`, so
`winningValue` stays `nil`, the rejection block at `:155-162` is never entered, and the merge
returns the project's `true` with no error. This AC is the security-relevant half of M1.

#### AC-CTP-004 — the falsey-value blast-radius enumeration exists

```bash
test -s .moai/specs/SPEC-CONFIG-TIER-PERSIST-001/falsey-key-inventory.md && \
  grep -c '^| ' .moai/specs/SPEC-CONFIG-TIER-PERSIST-001/falsey-key-inventory.md
```

Expected: exit 0 and a count greater than 1 (a header row plus at least one key row). The file
tables every key in `internal/config/defaults.go` and the shipped `sections/*.yaml` whose effective
value changes once falsey values win. R1 and R2 are mitigated by this table or not at all.

Baseline: `test -s .../falsey-key-inventory.md` exits 1 — the file does not exist.

### M2 — Tier ordering and local-tier reachability

#### AC-CTP-005 — `SrcLocal` outranks `SrcProject`

```bash
go test -run 'TestSourceOrdering_LocalOutranksProject' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestSourceOrdering_LocalOutranksProject` line, asserting
`SrcLocal.Priority() < SrcProject.Priority()` and that `MergeAll` with
`SrcProject{"k":"from-project"}` and `SrcLocal{"k":"from-local"}` yields `from-local` with
`Provenance.Source == SrcLocal`.

Baseline: the probe against `d5336214e` observed
`k => from-project (source=project) [project=2 local=3]`.

#### AC-CTP-006 — the full tier ordering is pinned as an explicit sequence

```bash
go test -run 'TestSourceOrdering_FullSequencePinned' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestSourceOrdering_FullSequencePinned` line. The test compares
`AllSources()` against a literal, hand-written slice in priority order, so any future reorder fails
here rather than silently changing resolution. A test that derives the expected order from the enum
itself is tautological and does not satisfy this AC.

Baseline: `grep -rn "AllSources()" internal/config/source_test.go` — no test asserts the complete
sequence against a literal today.

#### AC-CTP-007 — permission resolution is correct under the new ordering

```bash
go test -run 'TestPermissionResolver' -count=1 -v ./internal/permission/
```

Expected: every `--- PASS:` line that the baseline produced, plus a
`--- PASS: TestPermissionResolver_LocalOutranksProject` line asserting a local-tier permission rule
wins over a project-tier rule. `internal/permission/resolver.go:230` iterates the same enum, so the
reorder reaches it.

Baseline: run `go test -run 'TestPermissionResolver' -count=1 -v ./internal/permission/` before M2
and record the exact set of `--- PASS:` lines. Every one must still appear afterwards.

#### AC-CTP-008 — no `Source` ordinal is persisted anywhere

```bash
grep -rn 'int(Src\|Source(.*[0-9]\|json:"source"' --include='*.go' internal/ | grep -v '_test.go'
```

Expected: every surviving match is a within-process comparison, not a serialisation. Any match that
writes a `Source` ordinal to disk or JSON is a blocker for M2 and must be converted to the string
form via `Source.String()` / `ParseSource` before the reorder lands.

Baseline: record the verbatim match set before M2 so the reviewer can compare.

#### AC-CTP-009 — the typed `Loader` reads `.moai/config/local/` and it overrides `sections/`

```bash
go test -run 'TestLoader_LocalTierOverridesSections' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestLoader_LocalTierOverridesSections` line. The fixture writes
`sections/workflow.yaml` with `branch_guard.enabled: false` and `local/workflow.yaml` with
`branch_guard.enabled: true`, calls `Loader.Load`, and asserts `cfg.Workflow.BranchGuard.Enabled`
is `true`.

Baseline: `grep -n "config/local\|localDir\|SrcLocal" internal/config/loader.go` returns **zero
matches** (exit 1). The audit probe with exactly this fixture observed
`ConfigManager path: BranchGuard.Enabled = false`.

#### AC-CTP-010 — the local tier can also turn a setting OFF

```bash
go test -run 'TestLoader_LocalTierCanDisable' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestLoader_LocalTierCanDisable` line, with `sections/workflow.yaml` setting
`branch_guard.enabled: true` and `local/workflow.yaml` setting `false`, asserting the loaded value
is `false`. This is the AC that enforces REQ-CTP-012's ordering — it cannot pass unless M1 landed
first.

Baseline: the test does not exist. Absent M1, this direction fails even with M2's wiring in place.

#### AC-CTP-011 — an absent `local/` directory changes nothing

```bash
go test -run 'TestLoader_NoLocalDirIsSilent' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestLoader_NoLocalDirIsSilent` line, asserting that a fixture with no
`local/` directory produces the same `*Config` as before and emits no warning about the missing
directory.

Baseline: the test does not exist; today the directory is never consulted, so the behaviour is
trivially satisfied and must remain so after M2.

#### AC-CTP-012 — `.moai/config/local/` is gitignored

```bash
git check-ignore -v .moai/config/local/workflow.yaml
```

Expected: exit 0 and a line naming `.gitignore` and the matching pattern.

Baseline: `git check-ignore -v .moai/config/local/workflow.yaml` exits 1 with no output — the path
is not ignored, which is the `BLOCKER` recorded in `CLAUDE.local.md` §22.9.

#### AC-CTP-013 — the `CLAUDE.local.md` §22.9 BLOCKER note is cleared

```bash
grep -c 'BLOCKER (gitignore)' CLAUDE.local.md
```

Expected: `0`.

Baseline: `grep -c 'BLOCKER (gitignore)' CLAUDE.local.md` → `1`.

### M3 — Malformed-configuration contract

#### AC-CTP-014 — a malformed section is recorded as failed, not as loaded

```bash
go test -run 'TestLoader_MalformedSectionIsRecordedFailed' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestLoader_MalformedSectionIsRecordedFailed` line. The fixture writes a
`user.yaml` containing unparseable YAML, calls `Loader.Load`, and asserts the section appears in the
failed set and **not** in `LoadedSections()`.

Baseline: the test does not exist. Today `internal/config/loader.go:121-131` returns early after
`slog.Warn`, so the section is simply absent from `loadedSections` — indistinguishable from a
section whose file was never present.

#### AC-CTP-015 — an absent section stays silent

```bash
go test -run 'TestLoader_AbsentSectionIsSilent' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestLoader_AbsentSectionIsSilent` line, asserting that a fixture with no
`user.yaml` yields compiled defaults, records no failure, and returns no error. This pins
REQ-CTP-015 so M3 does not over-correct greenfield projects into errors.

Baseline: the test does not exist; the behaviour today is silent-default and must stay so.

#### AC-CTP-016 — a CLI command exits non-zero on a malformed section

```bash
go test -run 'TestCLI_MalformedSectionReturnsError' -count=1 -v ./internal/cli/
```

Expected: a `--- PASS: TestCLI_MalformedSectionReturnsError` line, asserting the returned error
names both the file (`user.yaml`) and the parse failure.

Baseline: the test does not exist. Today a malformed `user.yaml` produces a `slog.Warn` and the
command proceeds on compiled defaults with exit 0.

#### AC-CTP-017 — a hook stays fail-open but emits an operator-visible advisory

```bash
go test -run 'TestHook_MalformedSectionAdvisesButContinues' -count=1 -v ./internal/hook/
```

Expected: a `--- PASS: TestHook_MalformedSectionAdvisesButContinues` line, asserting the hook
returns its allow result **and** that the captured stderr writer contains the section filename.
Both halves are required: an assertion on continuation alone does not prove visibility, and an
assertion on the advisory alone does not prove fail-open.

Baseline: the test does not exist. Today the only signal is `slog.Warn`, which is not
operator-visible in a hook context.

#### AC-CTP-018 — `Save` refuses to persist a section that failed to load

```bash
go test -run 'TestConfigManager_SaveRefusesFailedSection' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestConfigManager_SaveRefusesFailedSection` line. The fixture writes a
malformed `user.yaml`, loads, calls `Save()`, asserts an error naming `user.yaml`, and asserts the
on-disk bytes of `user.yaml` are **byte-identical to before the `Save()`** — the compiled defaults
were not written over the user's file.

Baseline: the test does not exist. This is the data-loss path: pre-fix, `Save()` serialises the
compiled defaults it holds and overwrites the malformed-but-recoverable file. The byte-identity
assertion is the load-bearing half of this AC.

#### AC-CTP-019 — a clean section is still persisted alongside a failed one

```bash
go test -run 'TestConfigManager_SaveRefusalIsScopedToFailedSection' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestConfigManager_SaveRefusalIsScopedToFailedSection` line, asserting that
with `user.yaml` malformed and `language.yaml` clean, a `Save()` that errors on `user.yaml` still
wrote the updated `language.yaml`. Pins REQ-CTP-020 so one bad file does not block five good ones.

Baseline: the test does not exist.

### M4 — Atomic, mode-preserving writes

#### AC-CTP-020 — a `Save()` round trip preserves 0644

```bash
go test -run 'TestAtomicWrite_PreservesExistingMode' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestAtomicWrite_PreservesExistingMode` line. The fixture creates a target at
`0644`, runs the shared helper, and asserts `os.Stat(...).Mode().Perm() == 0o644`.

Baseline: the test does not exist. The audit probe observed
`os.CreateTemp mode = -rw-------`, `pre-existing target = -rw-r--r--`, `after rename = -rw-------`.
The repository itself carries four surviving witnesses:

```
$ ls -la .moai/config/sections/ | awk '$1 !~ /rw-r--r--/'
-rw-------@ 1 goos staff  248 .moai/config/sections/git-convention.yaml
-rw-------@ 1 goos staff 2491 .moai/config/sections/git-strategy.yaml
-rw-------@ 1 goos staff  200 .moai/config/sections/language.yaml
-rw-------@ 1 goos staff   35 .moai/config/sections/user.yaml
$ git ls-files -s .moai/config/sections/user.yaml
100644 0bb3f84df5a69061a68e6c1f3bb2f5a2730edc2d 0	.moai/config/sections/user.yaml
```

All four are `Save()`'s `saveSection` targets.

#### AC-CTP-021 — a newly created file gets `defs.FilePerm`, not 0600

```bash
go test -run 'TestAtomicWrite_NewFileUsesFilePerm' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestAtomicWrite_NewFileUsesFilePerm` line, asserting a write to a
non-existent path produces mode `0644`. This pins REQ-CTP-022 — `yamlpatch.atomicWrite` errors on a
missing target, which is correct for a patch but wrong for the shared helper.

Baseline: the test does not exist.

#### AC-CTP-022 — the migration widens the already-narrowed files

```bash
go test -run 'TestConfigModeMigration_WidensNarrowedFiles' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestConfigModeMigration_WidensNarrowedFiles` line, asserting that a fixture
`sections/` containing files at `0600` reads `0644` afterwards, and that a file already at `0644` is
untouched.

Baseline: the test does not exist. On the real tree the four files above remain at `-rw-------`
after any number of `Save()` calls, because a `Stat`-based preservation fix preserves what it finds.

#### AC-CTP-023 — the migration never narrows and never leaves `.moai/config/`

```bash
go test -run 'TestConfigModeMigration_NeverNarrowsAndStaysScoped' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestConfigModeMigration_NeverNarrowsAndStaysScoped` line, asserting that a
file at `0600` **outside** `.moai/config/` is untouched, and that a file inside at `0640` is widened
to `0644` while a file at `0666` is left alone. Pins REQ-CTP-026 and mitigates R7.

Baseline: the test does not exist.

#### AC-CTP-024 — no bare `os.WriteFile` reaches `.moai/config/**`

```bash
go test -run 'TestNoBareWriteFileIntoMoaiConfig' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestNoBareWriteFileIntoMoaiConfig` line. The guard walks the non-test Go
sources under `internal/config/`, `internal/cli/update/`, and `internal/core/project/` and asserts
no `os.WriteFile` call site targets a `.moai/config` path.

Baseline: the guard does not exist. The current offenders, all confirmed against `d5336214e`:

```
internal/cli/update/backup/restore.go:105,128,142,145
internal/cli/update/merge/merge.go:84
internal/core/project/initializer.go:371,388,399,411,424,435,450
internal/cli/harness.go:390
```

#### AC-CTP-025 — the hardcoded permission literal is gone

```bash
grep -n '0o644' internal/cli/harness.go
```

Expected: no output, exit 1.

Baseline: `grep -n '0o644' internal/cli/harness.go` →
`390:	if err := os.WriteFile(configPath, newData, 0o644); err != nil {`.

### M5 — `.gitignore` merge correctness

#### AC-CTP-026 — a template-dropped pattern is not migrated into the user section

```bash
go test -run 'TestMergeGitignoreFile_TemplateDroppedPatternNotMigrated' -count=1 -v ./internal/cli/update/merge/
```

Expected: a `--- PASS: TestMergeGitignoreFile_TemplateDroppedPatternNotMigrated` line. The fixture
is the F7 probe: template v1 contains `DS_Store`, the user adds `mysecret.env` and `build/`,
template v2 drops `DS_Store`, and the test asserts `DS_Store` does not appear below the
`# User Custom Patterns` header after the second merge.

Baseline: the probe against `d5336214e` produced, after the second merge:

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

#### AC-CTP-027 — duplicates are collapsed

```bash
go test -run 'TestMergeGitignoreFile_DeduplicatesUserPatterns' -count=1 -v ./internal/cli/update/merge/
```

Expected: a `--- PASS: TestMergeGitignoreFile_DeduplicatesUserPatterns` line, asserting exactly one
`build/` line in the output when the backup contained three.

Baseline: the same probe output above shows three `build/` lines surviving.

#### AC-CTP-028 — the merge is idempotent

```bash
go test -run 'TestMergeGitignoreFile_IsIdempotent' -count=1 -v ./internal/cli/update/merge/
```

Expected: a `--- PASS: TestMergeGitignoreFile_IsIdempotent` line, asserting that merging a file
with its own previous output produces byte-identical content.

Baseline: the test does not exist.

#### AC-CTP-029 — a pre-header backup keeps its user patterns

```bash
go test -run 'TestMergeGitignoreFile_PreHeaderBackupFallsBackToHeuristic' -count=1 -v ./internal/cli/update/merge/
```

Expected: a `--- PASS: TestMergeGitignoreFile_PreHeaderBackupFallsBackToHeuristic` line, asserting
that a backup containing **no** `# User Custom Patterns` header still preserves `mysecret.env`.
This is the AC that prevents the fix from becoming a data-loss regression (R8).

Baseline: the test does not exist. Today the heuristic is the only path, so pre-header backups are
handled by construction; after M5 the fallback must be explicit.

#### AC-CTP-030 — the existing characterization test still passes

```bash
go test -count=1 -v ./internal/cli/update/merge/ 2>&1 | grep -E '^(--- )?(PASS|FAIL|ok)'
```

Expected: no `FAIL` line, and the `--- PASS:` lines from
`gitignore_merge_characterization_test.go` still present. Where a characterization test pins the
pre-fix behaviour this SPEC deliberately changes, it is updated in the same commit with a comment
naming this SPEC — never deleted.

Baseline: `go test -count=1 ./internal/cli/update/merge/` → `ok` at `d5336214e`; record the exact
`--- PASS:` set before M5.

### M6 — Fallback visibility

#### AC-CTP-031 — the first 3-way failure is announced

```bash
go test -run 'TestRestore_ThreeWayFailureAdvisesOnFirstOccurrence' -count=1 -v ./internal/cli/update/backup/
```

Expected: a `--- PASS: TestRestore_ThreeWayFailureAdvisesOnFirstOccurrence` line, asserting that a
single 3-way merge failure writes an advisory naming the file to the captured stderr writer.

Baseline: the test does not exist. At `d5336214e`, `internal/cli/update/backup/restore.go:131-134`
calls `recordFallback(projectRoot, relPath, false, os.Stderr)` and falls through silently; the
noise-suppression ledger stays quiet until three consecutive failures, while the 2-way path at
`:139-141` prints `Warning: merge failed for %s, restoring backup` immediately.

#### AC-CTP-032 — an absent base is distinguishable from a merge failure

```bash
go test -run 'TestRestore_AbsentBaseAdvisoryIsDistinct' -count=1 -v ./internal/cli/update/backup/
```

Expected: a `--- PASS: TestRestore_AbsentBaseAdvisoryIsDistinct` line, asserting that the advisory
emitted when `os.ReadFile(basePath)` fails differs in text from the merge-failure advisory of
AC-CTP-031.

Baseline: the test does not exist. Today this path — `restore.go:118-120`, the `err == nil` guard
failing — emits neither a `recordFallback` call nor a warning. It is the quietest of the three.

#### AC-CTP-033 — the ledger still suppresses repetition

```bash
go test -run 'TestRestore_LedgerStillSuppressesRepetition' -count=1 -v ./internal/cli/update/backup/
```

Expected: a `--- PASS: TestRestore_LedgerStillSuppressesRepetition` line, asserting that repeated
failures for the same file do not produce an unbounded advisory stream. Pins REQ-CTP-035 so M6 does
not defeat the noise-suppression mechanism while making the first occurrence visible.

Baseline: the test does not exist.

### Cross-cutting

#### AC-CTP-034 — full suite and lint are clean

```bash
go test ./... -count=1 2>&1 | grep -E '^(FAIL|ok +github)' | grep -c '^FAIL'
golangci-lint run --timeout=2m; echo "lint exit=$?"
```

Expected: `0` failing packages, and `lint exit=0`.

Baseline: record both before M1. `go test ./internal/config/... ./internal/cli/update/... -count=1`
was green at `d5336214e`.

#### AC-CTP-035 — no template tree was touched

```bash
git diff --stat origin/main -- internal/template/templates/ | wc -l
```

Expected: `0`. NFR-CTP-004.

Baseline: `0` at branch point.

#### AC-CTP-036 — no test sets an OTEL environment variable

```bash
grep -rn 't.Setenv("OTEL' --include='*_test.go' internal/ | wc -l
```

Expected: `0`. NFR-CTP-002.

Baseline: record before M1.

## §C Falsification procedures

Each new guard must be shown to FAIL against unfixed code. `git stash` is prohibited (§A clause 5).
Two mechanisms are used, both borrowed from `SPEC-UPDATE-REINSTALL-LOOP-002` and
`SPEC-UPDATE-DATA-SURVIVAL-001`.

### C.1 — `go test -overlay` for a single-file mutation

Write the mutated source to a scratch path, point an overlay JSON at it, and run the guard.

```bash
SCRATCH=$(mktemp -d)
# 1. copy the file under test and re-introduce the defect
cp internal/config/merge.go "$SCRATCH/merge_mutated.go"
# 2. restore the zero-skip that M1 removed
#    (re-insert `if isZero(value) { continue }` after the valueExists check)
$EDITOR "$SCRATCH/merge_mutated.go"
# 3. point an overlay at it
cat > "$SCRATCH/overlay.json" <<EOF
{"Replace": {"$(pwd)/internal/config/merge.go": "$SCRATCH/merge_mutated.go"}}
EOF
# 4. the guard must now FAIL
go test -overlay "$SCRATCH/overlay.json" \
  -run 'TestMergeAll_ExplicitFalseWins' -count=1 -v ./internal/config/
rm -rf "$SCRATCH"
```

Expected: a `--- FAIL: TestMergeAll_ExplicitFalseWins` line. A `--- PASS` here means the guard does
not actually observe the fix and is non-falsifiable.

Applies to: AC-CTP-001, AC-CTP-002, AC-CTP-003 (re-insert the zero-skip); AC-CTP-005, AC-CTP-006
(swap `SrcLocal` and `SrcProject` back in `source.go`); AC-CTP-018 (remove the `Save` refusal);
AC-CTP-020, AC-CTP-021 (remove the `os.Chmod` from the shared helper); AC-CTP-026, AC-CTP-027
(restore the not-in-template heuristic as the only path); AC-CTP-031, AC-CTP-032 (remove the
advisory emissions).

### C.2 — a scratch worktree driven by `go -C` for behavioural or multi-file mutations

```bash
WT=$(mktemp -d)/wt
git worktree add --detach "$WT" HEAD
# mutate freely inside $WT — it is not the shared checkout
go -C "$WT" test -run 'TestLoader_LocalTierOverridesSections' -count=1 -v ./internal/config/
git worktree remove --force "$WT"
```

Applies to: AC-CTP-009, AC-CTP-010 (remove the local-tier read from `loader.go`); AC-CTP-022,
AC-CTP-023 (remove the migration); AC-CTP-024 (re-introduce a bare `os.WriteFile` into
`.moai/config` and confirm the guard fails).

### C.3 — non-Go ACs

AC-CTP-004, AC-CTP-012, AC-CTP-013, AC-CTP-025, AC-CTP-035, AC-CTP-036 are file-state or grep
assertions whose baselines are already recorded above as failing. Their falsification is the
baseline itself: each currently produces the opposite of its expected output on this tree, verbatim.

AC-CTP-008 has no falsification because it is a discovery step, not a guard — its purpose is to
surface a blocker before M2's reorder, and its output is compared by a human.

## §D Definition of Done

- Every AC in §B ran, and every `-run` AC's output carried its `--- PASS: <exact name>` line.
- Every falsification in §C ran and produced its expected `--- FAIL`.
- The falsey-key inventory (AC-CTP-004) exists and its table was reviewed before M1 landed.
- The persisted-ordinal grep (AC-CTP-008) was run and its result recorded before M2 landed.
- M1 landed before M2 (REQ-CTP-012); M2's gitignore entry landed in the same commit as its loader
  wiring (REQ-CTP-013).
- `go test ./... -count=1` green; `golangci-lint run` clean.
- No diff under `internal/template/templates/`.
- The four narrowed files in this repository read `-rw-r--r--`.
- `progress.md` §E.2 and §E.3 populated by `manager-develop` with commit SHAs and verbatim command
  output.
