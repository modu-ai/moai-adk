# SPEC-CONFIG-TIER-PERSIST-001 — Acceptance Criteria

## §A Discipline

1. **Every AC states a command and its expected observable output.** A criterion phrased as a
   property with no command is not an AC.
2. **`go test -run <pattern>` exits 0 on zero matches.** Every `-run` AC therefore also requires a
   verbatim `--- PASS: <exact test name>` line in the output. A `-run` AC whose only assertion is
   `exit 0` is vacuous and is rejected.
3. **Baselines were recorded from this tree** (code baseline `d5336214e`, branch
   `plan/epic-update-config-audit`, worktree `.claude/worktrees/epic-update-config`) and re-measured
   during the plan-audit revision. Each AC carries its observed pre-change baseline so a reviewer can
   tell a real change from a no-op.

   **One documented exception.** F1's *field* evidence — four `.moai/config/sections/*.yaml` files at
   `-rw-------` — was observed in the **primary checkout** (`/Users/goos/MoAI/moai-adk-go`), not in
   this tree. Re-measured in this worktree, all 32 section files read `-rw-r--r--` and none is
   narrowed, because git records only the executable bit and a `git worktree add` materialises files
   at the checkout umask. The observation is real but does not belong to this tree's baseline. F1's
   *mechanism* is unaffected and reproduces anywhere: a probe observed a pre-existing `-rw-r--r--`
   target become `-rw-------` after one `atomicWrite` round trip. Consequently no AC asserts on this
   repository's own section-file modes — AC-CTP-020/022/023 use `t.TempDir()` fixtures, and widening
   the primary checkout's four files is a manual follow-up recorded in §D, not a criterion.

3b. **An already-green criterion is labelled as such.** Where the plan-audit found an AC whose
   expected output was already produced before any implementation work, the AC is retained as a
   **regression guard** with its true measured baseline and an explicit `Class:` line. A regression
   guard is not evidence of progress; §D counts it separately.
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

Baseline: re-measured during the plan-audit revision —
`test -s .moai/specs/SPEC-CONFIG-TIER-PERSIST-001/falsey-key-inventory.md` exits **1**; the file does
not exist. Fails today; passes only once M1 step 1 produces the table.

#### AC-CTP-037 — `isZero`'s disposition is consistent with its caller count

> Numbering is append-only: this AC covers REQ-CTP-005a/b, which replaced the self-contradictory
> REQ-CTP-005 of v0.1.0, and is appended rather than inserted so no existing AC-CTP-0NN is renumbered.

```bash
CALLERS=$(grep -c 'isZero(' internal/config/merge.go)
DEFN=$(grep -c '^func isZero(' internal/config/merge.go)
echo "callers_incl_defn=$CALLERS defn=$DEFN"
test "$CALLERS" -ne 1 -o "$DEFN" -eq 0; echo "consistent=$?"
```

Expected: `consistent=0`. The assertion is that `isZero` is never left defined-but-unreferenced. A
match count of exactly 1 means the only occurrence is the definition itself — an orphan — and the
definition must therefore be absent (`DEFN=0`). Either outcome of REQ-CTP-005a/b satisfies it: keep
the helper **with** a caller (`CALLERS >= 2`), or remove both (`CALLERS = 0`, `DEFN = 0`).

Why it can fail: leaving `func isZero` in place after M1 removes its only call site at `:149-152`
yields `CALLERS=1, DEFN=1` → `consistent=1`. That is exactly the REQ-CTP-005b violation.

Baseline: measured on the baseline tree — `grep -n 'func isZero' internal/config/merge.go` →
`200:func isZero(v any) bool {`, and the call site is `merge.go:149` (`if isZero(value) {`), so
`CALLERS=2, DEFN=1` → `consistent=0`. The AC passes today and passes after a correct M1; it fails
only on the specific orphan state REQ-CTP-005b forbids. Class: **decision guard** — it constrains the
outcome of a choice rather than proving a behaviour change.

### M2 — Tier ordering and local-tier reachability

#### AC-CTP-005 — `SrcLocal` outranks `SrcProject`

```bash
go test -run 'TestSourceOrdering_LocalOutranksProject' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestSourceOrdering_LocalOutranksProject` line, asserting **both** of the
following, in this order of importance:

1. `MergeAll` with `SrcProject{"k":"from-project"}` and `SrcLocal{"k":"from-local"}` yields
   `from-local` with `Provenance.Source == SrcLocal`. **This is the load-bearing half** — it observes
   the `AllSources()` literal that actually decides resolution.
2. `SrcLocal.Priority() < SrcProject.Priority()`. This observes only the `iota` block, which no
   production code reads; it is asserted so a partial reorder is caught, not because it proves the
   fix reached anything.

Ordering the assertions this way is deliberate: an AC that asserted only (2) would pass against an
`iota`-only reorder that changes no merge outcome — the exact under-reach failure REQ-CTP-006 exists
to prevent.

Baseline: the probe against `d5336214e` observed
`k => from-project (source=project) [project=2 local=3]` — both halves fail today.

#### AC-CTP-006 — the existing ordering guard is updated, not duplicated or weakened

```bash
go test -run 'TestSourceOrdering$|TestAllSources$' -count=1 -v ./internal/config/
grep -n 'SrcPolicy, SrcUser, SrcLocal, SrcProject' internal/config/source_test.go
```

Expected: `--- PASS: TestSourceOrdering` **and** `--- PASS: TestAllSources`, plus a grep match
showing `expectedOrder` now lists `SrcLocal` before `SrcProject`. A test that derives the expected
order from the enum itself is tautological and does not satisfy this AC; neither does a newly
authored second guard that leaves the existing literal stale.

Why it can fail: after REQ-CTP-006's three-site reorder, `TestSourceOrdering` fails until its
`expectedOrder` literal is updated (it compares symbolic constants, so the reorder moves
`AllSources()` out from under it). If a run-phase agent "fixes" that by deleting the test or by
replacing the literal with `AllSources()` itself, the grep half of this AC fails.

Baseline: **the guard already exists and is green — the v0.1.0 baseline for this AC was false.**
Re-measured:

```
$ grep -rn 'AllSources()' internal/config/source_test.go
90:	sources := AllSources()
93:		t.Errorf("AllSources() returned %d sources, want 8", len(sources))
99:			t.Errorf("AllSources()[%d].Priority() = %d, want %d", i, sources[i].Priority(), i)
107:	sources := AllSources()
116:			t.Errorf("AllSources()[%d] = %v, want %v", i, sources[i], expected)
```

`source_test.go:104-119 TestSourceOrdering` already compares `AllSources()` element-by-element
against the literal `[]Source{SrcPolicy, SrcUser, SrcProject, SrcLocal, SrcPlugin, SrcSkill,
SrcSession, SrcBuiltin}` — precisely the "literal, hand-written slice" v0.1.0 asked someone to
author. `source_test.go:89-102 TestAllSources` separately pins `AllSources()[i].Priority() == i`.
Both pass today; the grep half fails today (the literal still reads `SrcProject, SrcLocal`).

#### AC-CTP-007 — permission resolution is correct under the new ordering

```bash
go test -run 'TestPermissionResolver' -count=1 -v ./internal/permission/
```

Expected: every `--- PASS:` line in the baseline set below, plus a
`--- PASS: TestPermissionResolver_LocalOutranksProject` line asserting a local-tier permission rule
wins over a project-tier rule.

**Mechanism (corrected).** `internal/permission/resolver.go` does **not** iterate the `Source` enum.
It builds its own literal slice at `:225-234`:

```go
tiers := []config.Source{
    config.SrcPolicy, config.SrcUser, config.SrcProject, config.SrcLocal,
    config.SrcPlugin, config.SrcSkill, config.SrcSession, config.SrcBuiltin,
}
```

The reorder reaches permission resolution only because that slice is named as one of REQ-CTP-006's
three edit sites. `resolver.go:201` — the `config.Source(999)` hook sentinel that v0.1.0 cited as
`:230` "iterating the enum" — is unrelated to tier order. This AC is therefore load-bearing: it fails
if step 2 of M2 reorders `source.go` but leaves `resolver.go` alone.

Baseline: re-measured on the baseline tree —
`go test -run 'TestPermissionResolver' -count=1 -v ./internal/permission/` → `ok … 1.607s` with 14
`--- PASS:` lines:

```
TestPermissionResolver_Resolve_PreAllowlist
TestPermissionResolver_Resolve_ProjectDeny
TestPermissionResolver_Resolve_PolicyDenyWins
TestPermissionResolver_Resolve_PlanModeDeniesWrites
TestPermissionResolver_Resolve_BypassPermissions
TestPermissionResolver_Resolve_BypassPermissionsInFork
TestPermissionResolver_Resolve_BubbleMode
TestPermissionResolver_Resolve_BubbleModeParentUnavailable
TestPermissionResolver_Resolve_NonInteractive
TestPermissionResolver_Resolve_HookOverride
TestPermissionResolver_Resolve_HookUpdatedInput
TestPermissionResolver_Resolve_ForkDepthExceedsLimit
TestPermissionResolver_ValidateMode
TestPermissionResolver_Resolve_TraceGeneration
```

All 14 must still appear afterwards; `TestPermissionResolver_LocalOutranksProject` does not exist
today and is the new assertion.

#### AC-CTP-008 — no `Source` ordinal is cast to an integer outside tests

```bash
grep -rnE 'int\(Src[A-Za-z]+\)' --include='*.go' internal/ | grep -v '_test.go'; echo "exit=$?"
```

Expected: `exit=1` (zero matches). An ordinal cast of a named `Source` constant is the one shape that
would let REQ-CTP-006's reorder change the meaning of already-written data; the binary assertion is
that no such cast exists in non-test code.

Why it can fail: adding `int(SrcProject)` anywhere under `internal/` outside a `_test.go` file
flips the grep to exit 0. The falsification in §C.4 does exactly that.

Baseline: **`exit=1` — the AC passes today.** Class: **regression guard** (per §A clause 3b), not a
progress criterion.

The broader discovery scan that v0.1.0 conflated with this AC is retained as an M2 step-1 task rather
than an AC, because its output requires human classification and cannot be made binary. Its verbatim
baseline, re-measured on the baseline tree:

```
$ grep -rn 'int(Src\|Source(.*[0-9]\|json:"source"' --include='*.go' internal/ | grep -v '_test.go'
internal/config/merge.go:321:		Source     string `json:"source"`
internal/config/provenance.go:81:	Source        string   `json:"source"`
internal/cli/doctor_permission.go:122:		// T-RT002-28: handle the hook tier sentinel (config.Source(999)) emitted by result.ExportTrace().
internal/permission/resolver.go:201:			Tier:    config.Source(999), // Hook tier is above SrcPolicy
internal/hook/failure_observer.go:121:	Source    string `json:"source"`
internal/session/state.go:22:	Source string    `json:"source"` // "user", "project", "local", "session", "hook"
```

Six matches, all safe: four are `string`-typed serialisation fields (the wire format is already the
name, not the ordinal), one is a comment, and one is the in-memory `config.Source(999)` hook
sentinel, which is above every real tier and unaffected by a swap within the tier range. M2 step 1
re-runs this scan and confirms the set has not grown; a **new** match outside these six is the
blocker.

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

#### AC-CTP-012 — `.moai/config/local/` remains gitignored

```bash
git check-ignore -v .moai/config/local/workflow.yaml; echo "exit=$?"
```

Expected: `exit=0` and a line naming `.gitignore` and the matching pattern.

Why it can fail: deleting the `.moai/config/local/` line from `.gitignore` flips it to `exit=1` with
no output. §C.4 runs that mutation.

Baseline: **already satisfied — the v0.1.0 baseline for this AC was false.** Re-measured:

```
$ git check-ignore -v .moai/config/local/workflow.yaml
.gitignore:183:.moai/config/local/	.moai/config/local/workflow.yaml      (exit 0)

$ git merge-base --is-ancestor b9fc75016 d5336214e; echo $?
0
```

The entry landed at `b9fc75016`, an ancestor of the code baseline, so the work REQ-CTP-013 described
was complete before this SPEC was drafted. Class: **regression guard** (§A clause 3b). It is not
evidence of M2 progress and §D counts it separately.

#### AC-CTP-013 — the `CLAUDE.local.md` §22.9 BLOCKER note stays cleared

```bash
grep -c 'BLOCKER (gitignore)' CLAUDE.local.md
```

Expected: `0`.

Why it can fail: reintroducing the note anywhere in `CLAUDE.local.md` yields `1`. §C.4 runs that
mutation.

Baseline: **already satisfied — the v0.1.0 baseline for this AC was false.** Re-measured:
`grep -c 'BLOCKER (gitignore)' CLAUDE.local.md` → **`0`**, not `1`. §22.9 already reads
"**gitignore (해결됨)**: `.moai/config/local/` 디렉터리는 이제 `.gitignore`에 등록되었다". Class:
**regression guard** (§A clause 3b).

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

Baseline: the test does not exist. The mechanism is confirmed by probe against the code baseline —
`os.CreateTemp mode = -rw-------`, `pre-existing target = -rw-r--r--`, `after rename = -rw-------` —
and this probe reproduces in any tree, so it is the fixture this AC promotes.

The field witnesses (four `Save()` `saveSection` targets at `-rw-------`, against a git index
recording `100644`) exist in the **primary checkout only** and are *not* part of this AC's baseline;
§A clause 3 records why they cannot reproduce in a worktree. Re-measured here:

```
$ ls -la .moai/config/sections/*.yaml | awk '$1 !~ /^-rw-r--r--/' | wc -l
0        # this worktree — no narrowed file exists to observe

$ ls -la /Users/goos/MoAI/moai-adk-go/.moai/config/sections/*.yaml | awk '$1 !~ /^-rw-r--r--/'
-rw-------@ .../git-convention.yaml
-rw-------@ .../git-strategy.yaml
-rw-------@ .../language.yaml
-rw-------@ .../user.yaml
```

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

Baseline: the test does not exist. The fixture is `t.TempDir()`-scoped by construction (NFR-CTP-001)
and does not read this repository's own files — §A clause 3 records why an assertion against them
would be vacuous in a worktree. The motivating observation stands in the primary checkout: the four
files there remain at `-rw-------` after any number of `Save()` calls, because a `Stat`-based
preservation fix preserves what it finds.

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

Baseline: re-measured — `grep -n '0o644' internal/cli/harness.go` →
`390:	if err := os.WriteFile(configPath, newData, 0o644); err != nil {` (exit 0). Fails today.

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
with its own previous output produces byte-identical content **after the M5 header-parse and dedupe
changes**.

Baseline: the named test does not exist, but **the property is already guarded and green** — a
finding the plan-audit did not surface. Re-measured:

```
$ go test -count=1 -v ./internal/cli/update/merge/ 2>&1 | grep -i idempot
--- PASS: TestMergeGitignore_DoubleReMerge_Idempotent (0.01s)
```

`TestMergeGitignore_DoubleReMerge_Idempotent` (in
`internal/cli/update/merge/gitignore_merge_characterization_test.go`) already asserts double-re-merge
idempotence and passes on the baseline tree. AC-CTP-028 is therefore **not** a new-behaviour
criterion: its value is that M5's header parse and dedupe must not *break* an invariant that already
holds. Two consequences follow. First, the new test must exercise a case the existing one does not —
specifically a backup carrying a `# User Custom Patterns` header **and** duplicate user lines, so
that dedupe and re-merge interact. Second, the pre-existing test must still pass afterwards
(AC-CTP-030 covers that). Class: **regression guard extended into a new case**, not a fresh
behavioural assertion.

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

Baseline: re-measured — `go test -count=1 -v ./internal/cli/update/merge/` → `ok … 1.262s`, **35
`--- PASS:` lines, zero `FAIL`**. The eight characterization tests that M5 most directly risks are:

```
--- PASS: TestMergeGitignoreFile_NoUserAdditions
--- PASS: TestMergeGitignoreFile_WithUserAdditions
--- PASS: TestMergeGitignore_UserLineCollidingWithTemplate_Deduplicated
--- PASS: TestMergeGitignore_CRLFBackup_PatternsSurvive
--- PASS: TestMergeGitignore_UserCommentAnnotationsNotCarried
--- PASS: TestMergeGitignore_ReMergeWithExistingHeader_PatternsSurvive
--- PASS: TestMergeGitignore_NegationLinesPreserved
--- PASS: TestMergeGitignore_DoubleReMerge_Idempotent
```

`TestMergeGitignore_ReMergeWithExistingHeader_PatternsSurvive` is the one M5 must read before
editing: it already asserts that pattern lines under a *previous* run's header are re-collected under
the new header. M5's header parse changes the classification path those lines take, so this test is
the most likely to need a same-commit update with a comment naming this SPEC — never a deletion.

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

Baseline: re-measured — `0`. Class: **regression guard** (§A clause 3b). It already holds and its
purpose is to fail if a run-phase commit touches the template tree; its falsification is in §C.4, not
in the baseline.

#### AC-CTP-036 — no test sets an OTEL environment variable

```bash
grep -rn 't.Setenv("OTEL' --include='*_test.go' internal/ | wc -l
```

Expected: `0`. NFR-CTP-002.

Baseline: re-measured — `grep -rn 't.Setenv("OTEL' --include='*_test.go' internal/ | wc -l` → `0`.
Class: **regression guard** (§A clause 3b); falsification in §C.4.

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

Applies to: AC-CTP-001, AC-CTP-002, AC-CTP-003 (re-insert the zero-skip); AC-CTP-037 (delete the
zero-skip but keep `func isZero`); AC-CTP-005, AC-CTP-006 (swap `SrcLocal` and `SrcProject` back in
**both** the `iota` block and the `AllSources()` literal — reverting only one leaves the enum and the
slice inconsistent and fails `TestAllSources` for the wrong reason, which is not a valid
falsification); AC-CTP-007 (swap them back in `internal/permission/resolver.go`'s `tiers` slice
**only**, leaving `source.go` fixed — this is the falsification that proves the permission assertion
is load-bearing rather than confirmatory); AC-CTP-018 (remove the `Save` refusal);
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

### C.3 — non-Go ACs whose falsification IS the baseline

Two ACs currently produce the opposite of their expected output on this tree, so their baseline is
their falsification, verbatim:

| AC | Command | Observed now | Expected after |
|---|---|---|---|
| AC-CTP-004 | `test -s .../falsey-key-inventory.md` | exit **1** | exit 0 + row count > 1 |
| AC-CTP-025 | `grep -n '0o644' internal/cli/harness.go` | exit **0**, one match at `:390` | exit 1, no output |

**v0.1.0 listed four more ACs here and was wrong about all four.** AC-CTP-012 (exit 0), AC-CTP-013
(`0`), AC-CTP-035 (`0`), and AC-CTP-036 (`0`) each produce **exactly their expected output** on this
tree, so the "baseline is the falsification" claim does not hold for them — passing proves nothing
about the implementation. They are reclassified as regression guards and given real, runnable
falsifications in §C.4.

### C.4 — falsification for already-satisfied regression guards

A regression guard cannot be falsified by its baseline. Each is falsified by mutating the state it
protects and confirming the command flips. All four mutations are reverted immediately; none uses
`git stash` (§A clause 5), and each is confined to a scratch copy or an immediately-undone edit.

**AC-CTP-012** — remove the ignore entry from a `.gitignore` placed in a **scratch git repository**,
and run the check inside that repository:

```bash
SCRATCH=$(mktemp -d); trap 'rm -rf "$SCRATCH"' EXIT
git init -q "$SCRATCH"
grep -c '^\.moai/config/local/$' .gitignore                      # mutation precondition: must print 1
grep -v '^\.moai/config/local/$' .gitignore > "$SCRATCH/.gitignore"
mkdir -p "$SCRATCH/.moai/config/local"
git -C "$SCRATCH" check-ignore -v --no-index .moai/config/local/workflow.yaml; echo "exit=$?"
```

Expected: the precondition grep prints `1` (the entry exists to be removed — a `0` means the
mutation is a no-op and the falsification proves nothing), then `exit=1` with no matched-pattern
line. Observed, executed at authoring time:

```
$ grep -c '^\.moai/config/local/$' .gitignore
1
$ git -C "$SCRATCH" check-ignore -v --no-index .moai/config/local/workflow.yaml; echo "exit=$?"
exit=1
```

Control — the same scratch repository with the **unmutated** `.gitignore` copied in, confirming the
flip is caused by the removed entry and not by the scratch environment:

```
$ cp .gitignore "$SCRATCH/.gitignore"
$ git -C "$SCRATCH" check-ignore -v --no-index .moai/config/local/workflow.yaml; echo "exit=$?"
.gitignore:183:.moai/config/local/	.moai/config/local/workflow.yaml
exit=0
```

If the mutated run still exits 0, the scratch repository is picking up a `.gitignore` other than the
one written into it (for example a parent-directory ignore file or a global excludes file) — inspect
the matched-pattern line, which names the file and line number of whatever actually matched.

> **Why `core.excludesFile` was abandoned.** An earlier form of this procedure ran
> `git -c core.excludesFile="$SCRATCH/gitignore.mutated" check-ignore … ` against **this** repository.
> That is inert: `core.excludesFile` replaces only the *user/global* excludes file and cannot disable
> a repository-tracked `.gitignore`, so the original `.gitignore:183` kept matching and the procedure
> returned `exit=0` where it predicted `exit=1`. Observed:
>
> ```
> $ git -c core.excludesFile="$SCRATCH/gitignore.mutated" check-ignore -v --no-index .moai/config/local/workflow.yaml
> .gitignore:183:.moai/config/local/	.moai/config/local/workflow.yaml
> exit=0
> ```
>
> Its diagnostic text made this worse rather than surfacing it: it read "another pattern is matching
> and AC-CTP-012 is not observing the entry it claims to observe", when in fact the **same** pattern
> matched and the mutation had simply never taken effect. A reader following that text would have
> hunted for a phantom second pattern instead of the real fault. The mutation must be applied where
> git will actually read it — hence the scratch repository above.

**AC-CTP-013** — append the note to a scratch copy and grep that copy:

```bash
SCRATCH=$(mktemp -d); trap 'rm -rf "$SCRATCH"' EXIT
{ cat CLAUDE.local.md; echo '- **BLOCKER (gitignore)**: mutation probe'; } > "$SCRATCH/CLAUDE.local.md"
grep -c 'BLOCKER (gitignore)' "$SCRATCH/CLAUDE.local.md"
```

Expected: `1`.

**AC-CTP-035** — touch one byte in the template tree, observe, revert:

```bash
F=$(git ls-files internal/template/templates/ | head -1)
printf '\n' >> "$F"
git diff --stat origin/main -- internal/template/templates/ | wc -l   # expect: non-zero
git checkout -- "$F"
git diff --stat origin/main -- internal/template/templates/ | wc -l   # expect: 0
```

Expected: a non-zero count while mutated, `0` after revert. A count that stays `0` while mutated
means the AC's `git diff --stat origin/main` comparison is not observing the working tree, and the
guard is inert.

**AC-CTP-036** — add the forbidden call to a scratch file inside the scanned tree, observe, remove:

```bash
P=internal/config/zzz_otel_probe_test.go
printf 'package config\n\nimport "testing"\n\nfunc TestZZZOtelProbe(t *testing.T) { t.Setenv("OTEL_SERVICE_NAME", "x") }\n' > "$P"
grep -rn 't.Setenv("OTEL' --include='*_test.go' internal/ | wc -l   # expect: 1
rm -f "$P"
grep -rn 't.Setenv("OTEL' --include='*_test.go' internal/ | wc -l   # expect: 0
```

Expected: `1` while the probe file exists, `0` after removal. **Remove the probe file before
committing** — it is a falsification artifact, not a fixture.

**AC-CTP-008** — same shape, using an ordinal cast:

```bash
P=internal/config/zzz_ordinal_probe.go
printf 'package config\n\nvar zzzProbe = int(SrcProject)\n' > "$P"
grep -rnE 'int\(Src[A-Za-z]+\)' --include='*.go' internal/ | grep -v '_test.go'; echo "exit=$?"  # expect: exit=0
rm -f "$P"
grep -rnE 'int\(Src[A-Za-z]+\)' --include='*.go' internal/ | grep -v '_test.go'; echo "exit=$?"  # expect: exit=1
```

**AC-CTP-037** is falsified by the §C.1 overlay: delete the `if isZero(value) { continue }` block
from the mutated `merge.go` while leaving `func isZero` in place, then run the AC's shell snippet
against `$SCRATCH/merge_mutated.go`. Expected: `consistent=1`.

## §D Definition of Done

- Every AC in §B ran, and every `-run` AC's output carried its `--- PASS: <exact name>` line.
- Every falsification in §C.1 and §C.2 ran and produced its expected `--- FAIL`; every falsification
  in §C.4 ran and produced its expected flip, with the mutation reverted afterwards.
- **The five regression guards (AC-CTP-008, AC-CTP-012, AC-CTP-013, AC-CTP-035, AC-CTP-036) are
  counted separately from the behavioural criteria and are not reported as evidence of progress.**
  They were already green before any work started; their §C.4 falsification is what makes them
  meaningful. Reporting "36 of 37 AC pass" without this split overstates completion by five.
- The falsey-key inventory (AC-CTP-004) exists and its table was reviewed before M1 landed.
- The ordinal scan (AC-CTP-008 + its M2 step-1 discovery scan) was run and its result compared
  against the recorded six-match baseline before M2 landed.
- `isZero`'s disposition was decided and recorded (REQ-CTP-005a/b, AC-CTP-037).
- M1 landed before M2 (REQ-CTP-012). REQ-CTP-006's three ordering sites were reordered in **one**
  commit together with `TestSourceOrdering`'s literal update.
- `spec.md` §K is present and the REQ-CTP-033 reversal was accepted at Implementation Kickoff
  Approval, **or** REQ-CTP-033/AC-CTP-031 were withdrawn together and M6 landed steps 2-3 only.
- `go test ./... -count=1` green; `golangci-lint run` clean.
- No diff under `internal/template/templates/`.
- No falsification probe file (`zzz_*_probe*.go`) remains in the tree.
- **Manual follow-up, not a criterion:** the four narrowed files in the primary checkout
  (`git-convention.yaml`, `git-strategy.yaml`, `language.yaml`, `user.yaml`) are widened to `0644` by
  running the REQ-CTP-025 migration there. This is not an AC because the files read `0644` in every
  worktree and an assertion on them passes vacuously (§A clause 3).
- `progress.md` §E.2 and §E.3 populated by `manager-develop` with commit SHAs and verbatim command
  output.
