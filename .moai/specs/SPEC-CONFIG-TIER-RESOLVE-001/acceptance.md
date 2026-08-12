# SPEC-CONFIG-TIER-RESOLVE-001 — Acceptance Criteria

## §A Discipline

1. **Every AC states a command and its expected observable output.** A criterion phrased as a property with no command is not an AC.
2. **`go test -run <pattern>` exits 0 on zero matches.** Every `-run` AC therefore also requires a verbatim `--- PASS: <exact test name>` line in the output. A `-run` AC whose only assertion is `exit 0` is vacuous and is rejected.
3. **Baselines were recorded from the parent SPEC's tree** (code baseline `d5336214e`, parent branch `plan/epic-update-config-audit`, parent worktree `.claude/worktrees/epic-update-config-audit`) and re-validated against the current child baseline `865cd8aa2 = origin/main` at child authoring time. Each AC carries its observed pre-change baseline so a reviewer can tell a real change from a no-op.
3b. **An already-green criterion is labelled as such.** Where an AC's expected output is already produced before any implementation work, the AC is retained as a **regression guard** with its true measured baseline and an explicit `Class:` line. A regression guard is not evidence of progress; §D counts it separately.
4. **Every new guard needs a falsification** proving it FAILS against unfixed code. §C gives the runnable procedures.
5. **`git stash` is prohibited.** This checkout is shared with concurrent sessions. Falsification uses `go test -overlay` or a scratch `git worktree` driven by `go -C`.
6. **The probes that produced the parent §A defects are the regression fixtures below**; invented cases are not substituted for them.
7. **All fixtures use `t.TempDir()`** and touch no path outside it (NFR-CTR-001).

## §B Acceptance criteria

### M0 — falsey-value blast-radius enumeration

#### AC-CTR-004 — the falsey-value blast-radius enumeration exists

```bash
test -s .moai/specs/SPEC-CONFIG-TIER-RESOLVE-001/falsey-key-inventory.md && \
  grep -c '^| ' .moai/specs/SPEC-CONFIG-TIER-RESOLVE-001/falsey-key-inventory.md
```

Expected: exit 0 and a count greater than 1 (a header row plus at least one key row). The file tables every key in `internal/config/defaults.go` and the shipped `sections/*.yaml` whose effective value changes once falsey values win. R1 and R2 are mitigated by this table or not at all.

Baseline: re-measured at child authoring time — `test -s .moai/specs/SPEC-CONFIG-TIER-RESOLVE-001/falsey-key-inventory.md` exits **1**; the file does not exist. Fails today; passes only once M0 produces the table.

### M1 — Tier merge semantics

#### AC-CTR-001 — an explicit `false` from a higher tier wins

```bash
go test -run 'TestMergeAll_ExplicitFalseWins' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestMergeAll_ExplicitFalseWins` line. The test is the §A probe promoted to a fixture: it layers `SrcProject{"workflow.branch_guard.enabled": false, "x.count": 0, "x.name": ""}` over `SrcBuiltin{...: true, 42, "default"}` and asserts each merged value equals the project value with `Provenance.Source == SrcProject`.

Baseline: the test does not exist — `go test -run 'TestMergeAll_ExplicitFalseWins' ./internal/config/` prints `ok ... [no tests to run]` and exits 0, which is exactly the vacuity clause 2 excludes. The probe that produced the pre-fix behaviour observed:

```
workflow.branch_guard.enabled    => true ok=true source=builtin
x.count                          => 42 ok=true source=builtin
x.name                           => "default" ok=true source=builtin
```

#### AC-CTR-002 — absence and falsey-presence remain distinguishable

```bash
go test -run 'TestMergeAll_AbsentKeyOmittedFalseyKeyPresent' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestMergeAll_AbsentKeyOmittedFalseyKeyPresent` line, asserting that `Get("never.set")` returns `ok=false` while `Get("explicitly.false")` returns `ok=true` with value `false`. This pins REQ-CTR-004 so the fix does not over-correct into emitting every absent key.

Baseline: the test does not exist. Pre-fix, both keys are indistinguishable — the falsey key is skipped and never reaches `result.Set`, so both return `ok=false`.

#### AC-CTR-003 — strict-mode policy rejection is reachable for a falsey policy value

```bash
go test -run 'TestMergeAll_StrictModeRejectsOverrideOfFalseyPolicy' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestMergeAll_StrictModeRejectsOverrideOfFalseyPolicy` line, asserting `MergeAll` returns a `*PolicyOverrideRejected` when `strict_mode` is on, `SrcPolicy` supplies `false` for a policy-designated key, and `SrcProject` supplies `true`.

Baseline: pre-fix, the falsey policy value is skipped at `internal/config/merge.go:149-152`, so `winningValue` stays `nil`, the rejection block at `:155-162` is never entered, and the merge returns the project's `true` with no error. This AC is the security-relevant half of M1.

#### AC-CTR-006 — `isZero`'s disposition is consistent with its caller count

> Covers REQ-CTR-005a/b. Either outcome (retain with a caller, or remove both) satisfies this AC.

```bash
CALLERS=$(grep -c 'isZero(' internal/config/merge.go)
DEFN=$(grep -c '^func isZero(' internal/config/merge.go)
echo "callers_incl_defn=$CALLERS defn=$DEFN"
test "$CALLERS" -ne 1 -o "$DEFN" -eq 0; echo "consistent=$?"
```

Expected: `consistent=0`. The assertion is that `isZero` is never left defined-but-unreferenced. A match count of exactly 1 means the only occurrence is the definition itself — an orphan — and the definition must therefore be absent (`DEFN=0`). Either outcome of REQ-CTR-005a/b satisfies it: keep the helper **with** a caller (`CALLERS >= 2`), or remove both (`CALLERS = 0`, `DEFN = 0`).

Why it can fail: leaving `func isZero` in place after M1 removes its only call site at `:149-152` yields `CALLERS=1, DEFN=1` → `consistent=1`. That is exactly the REQ-CTR-005b violation.

Baseline: measured on the baseline tree — `grep -n 'func isZero' internal/config/merge.go` → `200:func isZero(v any) bool {`, and the call site is `merge.go:149` (`if isZero(value) {`), so `CALLERS=2, DEFN=1` → `consistent=0`. The AC passes today and passes after a correct M1; it fails only on the specific orphan state REQ-CTR-005b forbids. Class: **decision guard** — it constrains the outcome of a choice rather than proving a behaviour change.

### M2 — Tier ordering

#### AC-CTR-005 — `SrcLocal` outranks `SrcProject`

```bash
go test -run 'TestSourceOrdering_LocalOutranksProject' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestSourceOrdering_LocalOutranksProject` line, asserting **both** of the following, in this order of importance:

1. `MergeAll` with `SrcProject{"k":"from-project"}` and `SrcLocal{"k":"from-local"}` yields `from-local` with `Provenance.Source == SrcLocal`. **This is the load-bearing half** — it observes the `AllSources()` literal that actually decides resolution.
2. `SrcLocal.Priority() < SrcProject.Priority()`. This observes only the `iota` block, which no production code reads; it is asserted so a partial reorder is caught, not because it proves the fix reached anything.

Ordering the assertions this way is deliberate: an AC that asserted only (2) would pass against an `iota`-only reorder that changes no merge outcome — the exact under-reach failure REQ-CTR-006 exists to prevent.

Baseline: the parent probe against `d5336214e` observed `k => from-project (source=project) [project=2 local=3]` — both halves fail today.

#### AC-CTR-006-guard — the existing ordering guard is updated, not duplicated or weakened

```bash
go test -run 'TestSourceOrdering$|TestAllSources$' -count=1 -v ./internal/config/
grep -n 'SrcPolicy, SrcUser, SrcLocal, SrcProject' internal/config/source_test.go
```

Expected: `--- PASS: TestSourceOrdering` **and** `--- PASS: TestAllSources`, plus a grep match showing `expectedOrder` now lists `SrcLocal` before `SrcProject`. A test that derives the expected order from the enum itself is tautological and does not satisfy this AC; neither does a newly authored second guard that leaves the existing literal stale.

Why it can fail: after REQ-CTR-006's three-site reorder, `TestSourceOrdering` fails until its `expectedOrder` literal is updated (it compares symbolic constants, so the reorder moves `AllSources()` out from under it). If a run-phase agent "fixes" that by deleting the test or by replacing the literal with `AllSources()` itself, the grep half of this AC fails.

Baseline: **the guard already exists and is green.** Re-measured (parent SPEC §B AC-CTP-006):

```
$ grep -rn 'AllSources()' internal/config/source_test.go
90:	sources := AllSources()
93:		t.Errorf("AllSources() returned %d sources, want 8", len(sources))
99:			t.Errorf("AllSources()[%d].Priority() = %d, want %d", i, sources[i].Priority(), i)
107:	sources := AllSources()
116:			t.Errorf("AllSources()[%d] = %v, want %v", i, sources[i], expected)
```

`source_test.go:104-119 TestSourceOrdering` already compares `AllSources()` element-by-element against the literal `[]Source{SrcPolicy, SrcUser, SrcProject, SrcLocal, SrcPlugin, SrcSkill, SrcSession, SrcBuiltin}`. Both `TestSourceOrdering` and `TestAllSources` pass today; the grep half fails today (the literal still reads `SrcProject, SrcLocal`).

#### AC-CTR-007 — permission resolution is correct under the new ordering

```bash
go test -run 'TestPermissionResolver' -count=1 -v ./internal/permission/
```

Expected: every `--- PASS:` line in the baseline set below, plus a `--- PASS: TestPermissionResolver_LocalOutranksProject` line asserting a local-tier permission rule wins over a project-tier rule.

**Mechanism (load-bearing).** `internal/permission/resolver.go` does **not** iterate the `Source` enum. It builds its own literal slice at `:225-234`:

```go
tiers := []config.Source{
    config.SrcPolicy, config.SrcUser, config.SrcProject, config.SrcLocal,
    config.SrcPlugin, config.SrcSkill, config.SrcSession, config.SrcBuiltin,
}
```

The reorder reaches permission resolution only because that slice is named as one of REQ-CTR-006's three edit sites. `resolver.go:201` — the `config.Source(999)` hook sentinel — is unrelated to tier order. This AC is therefore load-bearing: it fails if M2 reorders `source.go` but leaves `resolver.go` alone.

Baseline: measured on the baseline tree — `go test -run 'TestPermissionResolver' -count=1 -v ./internal/permission/` → `ok … 1.607s` with 14 `--- PASS:` lines:

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

All 14 must still appear afterwards; `TestPermissionResolver_LocalOutranksProject` does not exist today and is the new assertion.

#### AC-CTR-008 — no `Source` ordinal is cast to an integer outside tests

```bash
grep -rnE 'int\(Src[A-Za-z]+\)' --include='*.go' internal/ | grep -v '_test.go'; echo "exit=$?"
```

Expected: `exit=1` (zero matches). An ordinal cast of a named `Source` constant is the one shape that would let REQ-CTR-006's reorder change the meaning of already-written data; the binary assertion is that no such cast exists in non-test code.

Baseline: **`exit=1` — the AC passes today.** Class: **regression guard** (§A clause 3b).

The broader discovery scan (six safe matches: four `string`-typed serialisation fields, one comment, one in-memory `config.Source(999)` hook sentinel — all unaffected by a swap within the tier range) is retained as an M2 step-1 task rather than an AC, because its output requires human classification. M2 step 1 re-runs this scan and confirms the set has not grown; a **new** match outside the recorded six is the blocker.

### M3 — Local-tier reachability

#### AC-CTR-009 — the typed `Loader` reads `.moai/config/local/` and it overrides `sections/`

```bash
go test -run 'TestLoader_LocalTierOverridesSections' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestLoader_LocalTierOverridesSections` line. The fixture writes `sections/workflow.yaml` with `branch_guard.enabled: false` and `local/workflow.yaml` with `branch_guard.enabled: true`, calls `Loader.Load`, and asserts `cfg.Workflow.BranchGuard.Enabled` is `true`.

Baseline: `grep -n "config/local\|localDir\|SrcLocal" internal/config/loader.go` returns **zero matches** (exit 1). The audit probe with exactly this fixture observed `ConfigManager path: BranchGuard.Enabled = false`.

#### AC-CTR-010 — the local tier can also turn a setting OFF

```bash
go test -run 'TestLoader_LocalTierCanDisable' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestLoader_LocalTierCanDisable` line, with `sections/workflow.yaml` setting `branch_guard.enabled: true` and `local/workflow.yaml` setting `false`, asserting the loaded value is `false`. This is the AC that enforces REQ-CTR-012's ordering — it cannot pass unless M1 landed first.

Baseline: the test does not exist. Absent M1, this direction fails even with M3's wiring in place.

#### AC-CTR-011 — an absent `local/` directory changes nothing

```bash
go test -run 'TestLoader_NoLocalDirIsSilent' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestLoader_NoLocalDirIsSilent` line, asserting that a fixture with no `local/` directory produces the same `*Config` as before and emits no warning about the missing directory.

Baseline: the test does not exist; today the directory is never consulted, so the behaviour is trivially satisfied and must remain so after M3.

### Regression guards (REQ-CTR-013/014) — AC-only, no milestone work item

#### AC-CTR-013 — `.moai/config/local/` remains gitignored

```bash
git check-ignore -v .moai/config/local/workflow.yaml; echo "exit=$?"
```

Expected: `exit=0` and a line naming `.gitignore` and the matching pattern.

Why it can fail: deleting the `.moai/config/local/` line from `.gitignore` flips it to `exit=1` with no output.

Baseline: **already satisfied.** Re-measured (parent SPEC §B AC-CTP-012):

```
$ git check-ignore -v .moai/config/local/workflow.yaml
.gitignore:183:.moai/config/local/	.moai/config/local/workflow.yaml      (exit 0)

$ git merge-base --is-ancestor b9fc75016 d5336214e; echo $?
0
```

The entry landed at `b9fc75016`, an ancestor of the code baseline. Class: **regression guard** (§A clause 3b).

#### AC-CTR-014 — the `CLAUDE.local.md` §22.9 BLOCKER note stays cleared

```bash
grep -c 'BLOCKER (gitignore)' CLAUDE.local.md
```

Expected: `0`.

Why it can fail: reintroducing the note anywhere in `CLAUDE.local.md` yields `1`.

Baseline: **already satisfied.** Re-measured: `grep -c 'BLOCKER (gitignore)' CLAUDE.local.md` → **`0`**. §22.9 already reads "**gitignore (해결됨)**: `.moai/config/local/` 디렉터리는 이제 `.gitignore`에 등록되었다". Class: **regression guard** (§A clause 3b).

### Cross-cutting

#### AC-CTR-012x — full suite and lint are clean

```bash
go test ./... -count=1 2>&1 | grep -E '^(FAIL|ok +github)' | grep -c '^FAIL'
golangci-lint run --timeout=2m; echo "lint exit=$?"
```

Expected: `0` failing packages, and `lint exit=0`.

Baseline: record both before M1. The slice-(a)-affected packages (`internal/config/`, `internal/permission/`) were green at `865cd8aa2`.

#### AC-CTR-015x — no template tree was touched

```bash
git diff --stat origin/main -- internal/template/templates/ | wc -l
```

Expected: `0`. NFR-CTR-004.

Baseline: re-measured — `0`. Class: **regression guard** (§A clause 3b).

## §C Falsification procedures

Each new guard must be shown to FAIL against unfixed code. `git stash` is prohibited (§A clause 5).

### C.1 — `go test -overlay` for a single-file mutation

Write the mutated source to a scratch path, point an overlay JSON at it, and run the guard.

```bash
SCRATCH=$(mktemp -d)
cp internal/config/merge.go "$SCRATCH/merge_mutated.go"
# re-insert the zero-skip that M1 removed (the `if isZero(value) { continue }` after the valueExists check)
$EDITOR "$SCRATCH/merge_mutated.go"
cat > "$SCRATCH/overlay.json" <<EOF
{"Replace": {"$(pwd)/internal/config/merge.go": "$SCRATCH/merge_mutated.go"}}
EOF
go test -overlay "$SCRATCH/overlay.json" \
  -run 'TestMergeAll_ExplicitFalseWins' -count=1 -v ./internal/config/
rm -rf "$SCRATCH"
```

Expected: a `--- FAIL: TestMergeAll_ExplicitFalseWins` line. A `--- PASS` here means the guard does not actually observe the fix and is non-falsifiable.

Applies to: AC-CTR-001, AC-CTR-002, AC-CTR-003 (re-insert the zero-skip); AC-CTR-006 (delete the zero-skip but keep `func isZero`); AC-CTR-005, AC-CTR-006-guard (swap `SrcLocal` and `SrcProject` back in **both** the `iota` block and the `AllSources()` literal — reverting only one leaves the enum and the slice inconsistent and fails `TestAllSources` for the wrong reason, which is not a valid falsification); AC-CTR-007 (swap them back in `internal/permission/resolver.go`'s `tiers` slice **only**, leaving `source.go` fixed — this is the falsification that proves the permission assertion is load-bearing rather than confirmatory).

### C.2 — a scratch worktree driven by `go -C` for behavioural or multi-file mutations

```bash
WT=$(mktemp -d)/wt
git worktree add --detach "$WT" HEAD
# mutate freely inside $WT — it is not the shared checkout
go -C "$WT" test -run 'TestLoader_LocalTierOverridesSections' -count=1 -v ./internal/config/
git worktree remove --force "$WT"
```

Applies to: AC-CTR-009, AC-CTR-010 (remove the local-tier read from `loader.go`); AC-CTR-011 (remove the absent-directory silent path).

### C.3 — non-Go ACs whose falsification IS the baseline

| AC | Command | Observed now | Expected after |
|---|---|---|---|
| AC-CTR-004 | `test -s .../falsey-key-inventory.md` | exit **1** | exit 0 + row count > 1 |

### C.4 — falsification for already-satisfied regression guards

A regression guard cannot be falsified by its baseline. Each is falsified by mutating the state it protects and confirming the command flips. All mutations are reverted immediately; none uses `git stash` (§A clause 5).

**AC-CTR-013** — remove the ignore entry from a `.gitignore` placed in a **scratch git repository**, and run the check inside that repository (the parent SPEC's C.4 procedure verbatim — `core.excludesFile` cannot override a repository-tracked `.gitignore`):

```bash
SCRATCH=$(mktemp -d); trap 'rm -rf "$SCRATCH"' EXIT
git init -q "$SCRATCH"
grep -c '^\.moai/config/local/$' .gitignore                      # mutation precondition: must print 1
grep -v '^\.moai/config/local/$' .gitignore > "$SCRATCH/.gitignore"
mkdir -p "$SCRATCH/.moai/config/local"
git -C "$SCRATCH" check-ignore -v --no-index .moai/config/local/workflow.yaml; echo "exit=$?"
```

Expected: the precondition grep prints `1`, then `exit=1` with no matched-pattern line.

**AC-CTR-014** — append the note to a scratch copy and grep that copy:

```bash
SCRATCH=$(mktemp -d); trap 'rm -rf "$SCRATCH"' EXIT
{ cat CLAUDE.local.md; echo '- **BLOCKER (gitignore)**: mutation probe'; } > "$SCRATCH/CLAUDE.local.md"
grep -c 'BLOCKER (gitignore)' "$SCRATCH/CLAUDE.local.md"
```

Expected: `1`.

**AC-CTR-008** — same shape, using an ordinal cast:

```bash
P=internal/config/zzz_ordinal_probe.go
printf 'package config\n\nvar zzzProbe = int(SrcProject)\n' > "$P"
grep -rnE 'int\(Src[A-Za-z]+\)' --include='*.go' internal/ | grep -v '_test.go'; echo "exit=$?"  # expect: exit=0
rm -f "$P"
grep -rnE 'int\(Src[A-Za-z]+\)' --include='*.go' internal/ | grep -v '_test.go'; echo "exit=$?"  # expect: exit=1
```

Expected: `exit=0` while the probe file exists, `exit=1` after removal. **Remove the probe file before committing** — it is a falsification artifact, not a fixture.

**AC-CTR-006** is falsified by the §C.1 overlay: delete the `if isZero(value) { continue }` block from the mutated `merge.go` while leaving `func isZero` in place, then run the AC's shell snippet against `$SCRATCH/merge_mutated.go`. Expected: `consistent=1`.

**AC-CTR-015x** — touch one byte in the template tree, observe, revert:

```bash
F=$(git ls-files internal/template/templates/ | head -1)
printf '\n' >> "$F"
git diff --stat origin/main -- internal/template/templates/ | wc -l   # expect: non-zero
git checkout -- "$F"
git diff --stat origin/main -- internal/template/templates/ | wc -l   # expect: 0
```

## §D Definition of Done

- Every AC in §B ran, and every `-run` AC's output carried its `--- PASS: <exact name>` line.
- Every falsification in §C.1 and §C.2 ran and produced its expected `--- FAIL`; every falsification in §C.4 ran and produced its expected flip, with the mutation reverted afterwards.
- **The four regression guards (AC-CTR-008, AC-CTR-013, AC-CTR-014, AC-CTR-015x) are counted separately from the behavioural criteria and are not reported as evidence of progress.** They were already green before any work started; their §C.4 falsification is what makes them meaningful.
- The falsey-key inventory (AC-CTR-004) exists and its table was reviewed before M1 landed.
- The ordinal scan (AC-CTR-008 + its M2 step-1 discovery scan) was run and its result compared against the recorded six-match baseline before M2 landed.
- `isZero`'s disposition was decided and recorded (REQ-CTR-005a/b, AC-CTR-006).
- M1 landed before M3 (REQ-CTR-012). REQ-CTR-006's three ordering sites were reordered in **one** commit together with `TestSourceOrdering`'s literal update (REQ-CTR-008).
- `go test ./... -count=1` green; `golangci-lint run` clean.
- No diff under `internal/template/templates/`.
- No falsification probe file (`zzz_*_probe*.go`) remains in the tree.
- `progress.md` §E.2 and §E.3 populated by `manager-develop` with commit SHAs and verbatim command output.
