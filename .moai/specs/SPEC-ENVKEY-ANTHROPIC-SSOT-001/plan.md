---
spec: SPEC-ENVKEY-ANTHROPIC-SSOT-001
phase: plan
tier: M
---

# Implementation Plan - SPEC-ENVKEY-ANTHROPIC-SSOT-001

Milestones are ordered by **decision reversibility**: the decisions most likely
to change (constant naming, prefix handling, guard architecture) lead; the
mechanical per-package transitions follow. Review attention belongs at M1 and M2.

---

## A. Context

Base commit: `76d9a8f3b`. Worktree branch: `plan/SPEC-ENVKEY-ANTHROPIC-SSOT-001`.

Scope confirmed by measurement, not assumption: 83 bare `ANTHROPIC_*` literals
across 10 production files, 9 distinct literal values, 6 existing constants, 2
names with no constant. See `spec.md` section A for the full per-file table and
the reproducing commands.

---

## B. Known Issues (discovered during plan-phase investigation)

### B.1 The existing guard is scope-blind (load-bearing)

`internal/cli/glm_env_parity_test.go:66` walks from `os.Getwd()`, which under
`go test` is the package directory `internal/cli`. 37 of the 83 occurrences live
outside that root and are structurally unreachable. Copying this pattern would
produce a guard covering 55 percent of the surface while reporting PASS.

Observed baseline (the guard currently passes, on a tree with 83 violations of
the invariant this SPEC introduces):

```
--- PASS: TestNoBareGLMEnvVarLiteralsInCLIProduction (0.03s)
```

That PASS is exactly the failure mode: it is green because it cannot see, not
because the tree is clean.

### B.2 Widening the walk root exposes the SSOT file

`internal/config/envkeys.go` contains 6 bare `"ANTHROPIC_*"` literals by design
(they are the constant definitions). A repo-root walk reaches them. The new guard
MUST exclude that file or it flags its own SSOT. This hazard did not exist for the
`internal/cli`-rooted guard.

### B.3 CI is PR-mandatory, so the RED window must not reach `main`

Branch protection is `enforce_admins: true`; all changes route through a PR with
4 required CI checks. The guard introduced at M2 is RED until the last package
transition lands at M5. Therefore **all milestones land in a single PR**, and the
RED state exists only inside this worktree. `main` never observes a failing guard.
See section D.3.

---

## C. Pre-flight

- [ ] Worktree is at `plan/SPEC-ENVKEY-ANTHROPIC-SSOT-001`, base `76d9a8f3b`
- [ ] `go test ./...` green before any edit (establish the behaviour-preservation baseline)
- [ ] `git status` clean apart from the SPEC directory
- [ ] Open questions OQ-1..OQ-4 resolved at the Implementation Kickoff Approval gate

### C.1 [HARD] Census re-measurement gate - run BEFORE M1

Every count in this SPEC is a snapshot at `76d9a8f3b` (spec.md section A.5). Run
these four commands against the run-phase HEAD and compare to the recorded values.
**If any differs, halt** - do not enter M1. Re-baseline the affected ACs, record
the new numbers plus the new base commit in the spec.md HISTORY table, then resume.

```bash
# 1. total production literals            expect 83
grep -rn '"ANTHROPIC_[A-Z_]*"' internal/ pkg/ cmd/ --include='*.go' \
  --exclude='*_test.go' | grep -v '^internal/config/envkeys.go' | wc -l
# 2. reachable by the OLD internal/cli walk root   expect 46
grep -rn '"ANTHROPIC_[A-Z_]*"' internal/cli/ --include='*.go' --exclude='*_test.go' | wc -l
# 3. unique literal values                 expect 9
grep -rho '"ANTHROPIC_[A-Z_]*"' internal/ pkg/ cmd/ --include='*.go' \
  --exclude='*_test.go' | sort -u | wc -l
# 4. constants already in the SSOT         expect 6
grep -c '"ANTHROPIC_[A-Z_]*"' internal/config/envkeys.go
# 5. internal/hook split (AC-EAS-009)      expect 29
grep -rn '"ANTHROPIC_[A-Z_]*"' internal/hook/ --include='*.go' --exclude='*_test.go' | wc -l
# 6. tmux+statusline+sandbox split (AC-EAS-010)   expect 8
grep -rn '"ANTHROPIC_[A-Z_]*"' internal/tmux/ internal/statusline/ internal/sandbox/ \
  --include='*.go' --exclude='*_test.go' | wc -l
```

Command 1 minus command 2 is the out-of-CLI figure (`83 - 46 = 37`) that AC-EAS-005's
`>= 37` reachability threshold rests on. If commands 1 and 2 shift, that threshold
MUST be recomputed before M2, or the reachability assertion becomes unattributable.

Commands 5 and 6 exist because the totals alone are insufficient: a main-ward
commit that MOVES three literals from `internal/hook` to `internal/tmux` leaves
83 / 46 / 9 / 6 all unchanged, passes commands 1-4, and silently invalidates
AC-EAS-009's `29` and AC-EAS-010's `8`. The per-package splits must therefore be
re-measured too.

Numbers this gate does NOT re-measure, stated explicitly so their staleness is a
known and accepted risk rather than an unnoticed one: the `291` `_test.go`
occurrence count (AC-EAS-015), the `12` template-reference count (AC-EAS-014),
and the 10-file distribution table in spec.md A.1. All three are stability
assertions (their ACs require the number to stay unchanged), so a shift surfaces
as an AC failure rather than as a silent wrong baseline.

---

## D. Constraints

### D.1 Language and style

- All Go code, comments, and godoc in English (project rule).
- New constants follow the existing `EnvAnthropic<CamelCase>` pattern with Go
  initialism casing.

### D.2 Behaviour preservation

- Every constant's value is byte-identical to the literal it replaces. This is a
  name-indirection refactor only.
- Evidence: `go test ./...` green, plus `go vet ./...` clean.

### D.3 Landing strategy - single PR, no RED on main

The guard lands before the transitions (so it drives the work), which means it is
RED for the duration of M3-M5. To keep `main` green:

- All milestones M1-M6 land in **one PR** from this worktree.
- The RED window is intra-worktree only. CI observes the branch only at PR time,
  by which point M6 has driven the guard GREEN.
- A run-phase reader seeing the guard fail after M2 is observing the **expected
  RED window**, not a regression. This is stated here so it is not mis-triaged.

### D.4 Untouched surfaces

- `internal/template/templates/` (template neutrality) - no edits, no `make build`.
- `_test.go` files (hardcoding-allowed zone) - 291 occurrences stay as-is.
- `internal/cli/glm_env_parity_test.go` - the existing CLAUDE_CODE_* guard is not
  modified (OQ-4 recommendation).

---

## E. Self-Verification

Each milestone's exit is a command whose output is observed, not assumed. Verdict
commands and their pre-change baselines are enumerated in `acceptance.md`.

Standing rule for this SPEC: any AC whose verdict rests on a `go test -run`
selector MUST be run with `-v -count=1` and the verdict read from an observed
`--- PASS: <exact test name>` line. A zero-match `-run` selector exits 0 and would
otherwise read as a false PASS.

---

## F. Milestones

### M1 - SSOT constants (highest-reversibility decisions)

**Why first:** the naming choice and the prefix-vs-name decision are the two
judgements most likely to be revised on review. Everything downstream references
them, so settling them first avoids rework.

Work:

1. Add to the ANTHROPIC const block in `internal/config/envkeys.go`:
   - `EnvAnthropicAPIKey = "ANTHROPIC_API_KEY"`
   - `EnvAnthropicDefaultFableModel = "ANTHROPIC_DEFAULT_FABLE_MODEL"`
   - `EnvAnthropicPrefix = "ANTHROPIC_"` (doc comment must state: namespace
     prefix, not a variable name - resolves OQ-1)
2. Each constant carries an English godoc comment consistent with the existing 6.

Exit: constant count 6 -> 9; `go build ./...` clean. (AC-EAS-001, AC-EAS-002)

### M2 - repo-root guard, with reachability proven by natural RED

**Why second:** the guard's architecture (walk root, exclusion set, host package,
new-vs-extend) is the other high-reversibility decision. Landing it before the
transitions makes it the progress meter for M3-M5.

Work:

1. Create `internal/config/anthropic_env_ssot_test.go` with
   `TestNoBareAnthropicEnvVarLiteralsInProduction`.
2. **Walk root - reuse the same-package helper; do NOT copy a colliding symbol.**
   The guard file is `package config` (verified: `internal/config/required_checks_test.go:1`
   declares `package config`), so it can call the existing helper directly:

   | Symbol | Location | Signature | Usable here? |
   |--------|----------|-----------|--------------|
   | `findProjectRoot` | `internal/config/required_checks_test.go:162` | `func findProjectRoot(t *testing.T) string` | **YES - reuse this.** Same package, test-scoped, directly callable, no import needed. |
   | `findRepoRoot` | `internal/config/token_budget_guard.go:45` | `func findRepoRoot(start string) (string, bool)` | Already declared in this package (a **production** file). See hazard below. |
   | `findRepoRoot` | `internal/spec/drift_doctrine_test.go:13` | `func findRepoRoot(t *testing.T) string` | **NO** - unexported in a different package; not importable. |

   > **[HAZARD] `findRepoRoot` redeclaration.** `internal/config` already declares
   > `findRepoRoot` at `internal/config/token_budget_guard.go:45`. Copying the
   > `internal/spec` helper of that name into `internal/config` produces
   > `findRepoRoot redeclared in this block`, which breaks compilation of the
   > **entire** `internal/config` package - and with it every AC that runs
   > `go test ./internal/config/`. Call `findProjectRoot(t)` instead. Do not
   > copy, do not rename-and-copy: a second root-finder in one package is
   > duplicated logic with no benefit.

3. **Banned set - derived from the constants, exposed for runtime assertion.**
   Declare the banned set in terms of the `envkeys.go` identifiers, not bare
   literals:

   ```go
   // bannedAnthropicEnvNames is derived from the envkeys.go constants, so a new
   // constant that is not added here is a visible omission (REQ-EAS-008).
   var bannedAnthropicEnvNames = []string{
       EnvAnthropicPrefix, EnvAnthropicAPIKey, EnvAnthropicAuthToken,
       EnvAnthropicBaseURL, EnvAnthropicDefaultFableModel,
       EnvAnthropicDefaultHaikuModel, EnvAnthropicDefaultOpusModel,
       EnvAnthropicDefaultSonnetModel, EnvAnthropicReasoningEffort,
   }
   ```

   Consequence to note: a correctly-derived guard therefore contains **zero** bare
   `"ANTHROPIC_*"` literals. Any AC that tries to verify banned-set completeness by
   grepping this file for bare literals is structurally unsatisfiable - it can only
   pass if the derivation intent is abandoned. Completeness is verified **at
   runtime** instead (next item, and AC-EAS-006).

4. **Runtime banned-set assertion** (satisfies REQ-EAS-008 and closes the
   banned-set-subset hole): a **top-level test function named exactly
   `TestAnthropicBannedSetCoversAllNames`** asserts
   `len(bannedAnthropicEnvNames) == 9` and, table-driven, that each of the 9
   expected names is a member. It MUST be a top-level `func Test...`, NOT a
   `t.Run` sub-test: AC-EAS-006 selects it with a bare top-level `-run`
   pattern, which cannot select a sub-test, so a sub-test implementation
   would make a correct guard fail the AC. It lives in the same file as the
   scan guard but as a separate top-level function, so it can be GREEN while
   the scan guard is RED (exit signal 2 below). A guard that silently omits `ANTHROPIC_`,
   `ANTHROPIC_DEFAULT_FABLE_MODEL`, or `ANTHROPIC_REASONING_EFFORT` - the three
   names with **zero** out-of-CLI offenders, hence zero natural falsification
   coverage - fails here and only here.
5. Exclusions: `_test.go` suffix, and the SSOT file `internal/config/envkeys.go`
   (per B.2). The `envkeys.go` exclusion MUST be an exact-path match, not a
   substring or package-level match: excluding `internal/config` or `envkeys`
   broadly would exempt the whole SSOT package. AC-EAS-007 asserts the exclusion
   is narrow by planting a literal in a **sibling** `internal/config` file and
   observing RED.
6. Scan limited to `internal/`, `pkg/`, `cmd/`. These are three **enumerated**
   roots: a typo'd or omitted entry fails independently of the others, which is
   why AC-EAS-012 plants a probe under `cmd/` as well as under `internal/`
   (a probe in one root cannot prove another root is scanned).
7. **Failure output format**: on failure the guard MUST emit **one offender per
   line**, each line containing the offending file's path. AC-EAS-005 counts
   those lines (`grep -c ... >= 37`), so a guard that instead emits a summary
   count plus a truncated sample would satisfy REQ-EAS-007 yet fail AC-EAS-005.
   Reuse the existing `strings.Join(offenders, "\n  ")` shape from
   `internal/cli/glm_env_parity_test.go`.

**Reachability proof - the natural falsification corpus.** At M2 the tree still
holds all 83 literals, 37 of them outside `internal/cli/`. The guard's first run
must therefore report RED with **at least 37 offenders whose paths are outside
`internal/cli/`**. That observation is the direct evidence that the walk root
reaches past `internal/cli` - the defect in B.1 made mechanically visible. A guard
that goes RED with only `internal/cli` paths has the old scope bug and MUST be
fixed before M3. (AC-EAS-005)

Exit, three separable signals:

1. **The scan test is RED** (expected - the tree still holds 83 literals): its
   offender list contains paths from `internal/hook/`, `internal/tmux/`,
   `internal/statusline/`, and `internal/sandbox/`, and `envkeys.go` appears in
   **no** offender path. (AC-EAS-005)
2. **The runtime banned-set test `TestAnthropicBannedSetCoversAllNames`
   (top-level function, not a sub-test) is GREEN** - it asserts set membership, not
   tree cleanliness, so it does not participate in the RED window and must pass
   from the moment it lands. A RED here at M2 means the derived set is wrong.
   (AC-EAS-006)
3. `go build ./...` and the rest of `internal/config`'s suite still compile - the
   F/M2 `findRepoRoot` hazard did not fire. (AC-EAS-007 step (b) is deferred to M6,
   since it requires a clean tree to distinguish the planted offender from the
   83 pre-existing ones.)

### M3 - transition `internal/cli` (46 occurrences)

- `internal/cli/glm.go` (32)
- `internal/cli/launcher.go` (7)
- `internal/cli/settings.go` (6)
- `internal/cli/worktree/tmux_integration.go` (1, the prefix usage from OQ-1)

Exit: 0 bare literals under `internal/cli/`; guard still RED (37 remain);
`go test ./internal/cli/... ./internal/cli/worktree/...` green. (AC-EAS-008)

### M4 - transition `internal/hook` (29 occurrences)

- `internal/hook/session_end.go` (12)
- `internal/hook/glm_tmux.go` (9)
- `internal/hook/session_start.go` (8)

Exit: 0 bare literals under `internal/hook/`; guard still RED (8 remain);
`go test ./internal/hook/...` green. (AC-EAS-009)

### M5 - transition the remaining packages (8 occurrences)

- `internal/tmux/cg_detect.go` (4)
- `internal/statusline/metrics.go` (3)
- `internal/sandbox/env.go` (1)

Exit: guard turns **GREEN** - the RED window from D.3 closes here.
(AC-EAS-010, AC-EAS-011)

### M6 - post-green falsifiability proof and full verification

**Why last, and why not merged into M2:** M2's natural RED proves the guard
*reaches* the wider surface while violations still exist. It does not prove the
guard *still bites* once the tree is clean - a guard that is green because its
banned set silently emptied would look identical. M6 supplies the second,
independent proof.

Work:

1. **Deliberate re-introduction, `internal/` (verified target).** After M5,
   `internal/sandbox/env.go:32` holds `config.EnvAnthropicAPIKey` as an element of
   `defaultDenyList`. Revert that one element to the bare `"ANTHROPIC_API_KEY"`.
   Run the guard; observe FAIL naming that file. Revert; re-run; observe PASS.
   Both directions recorded verbatim. (AC-EAS-012 steps a-c)

   > Target verified at base by grep: `"ANTHROPIC_API_KEY"` has **exactly one**
   > production occurrence repo-wide, `internal/sandbox/env.go:32`. The earlier
   > draft named `internal/hook/session_start.go`, which contains **no**
   > `ANTHROPIC_API_KEY` reference at all (its ANTHROPIC references are
   > `_AUTH_TOKEN`, `_BASE_URL`, and the three `_DEFAULT_*_MODEL` names) - there
   > would have been nothing to replace, forcing the implementer to *insert* an
   > unused literal and risk a lint failure colliding with AC-EAS-013.
   > `internal/sandbox/env.go` is outside `internal/cli/`, so it still exercises
   > the widened walk root. The literal sits inside an existing string slice, so
   > reverting it introduces no unused identifier.

2. **Deliberate re-introduction, `cmd/` (REQ-EAS-005 coverage proof).** `pkg/` and
   `cmd/` hold zero literals (spec.md A.1), so nothing there is proven by the
   natural corpus. Plant one bare `"ANTHROPIC_API_KEY"` in `cmd/moai/main.go`,
   observe FAIL naming that file, revert, observe PASS. Without this step two
   thirds of REQ-EAS-005's claimed surface is asserted and never exercised.
   (AC-EAS-012 step d)
3. **SSOT-exclusion scope proof - both directions.** Confirm the guard stays GREEN
   with `envkeys.go`'s 6 literals present (the exclusion works), **and** that a
   bare literal planted in a *sibling* `internal/config` production file
   (`internal/config/log.go`) still goes RED (the exclusion is narrow, not a
   package-wide exemption). A one-directional check cannot distinguish a correct
   exact-path exclusion from an over-broad one that exempts the whole SSOT package.
   (AC-EAS-007)
4. Full suite: `go test ./...`, `go vet ./...`, `golangci-lint run`.
   (AC-EAS-013)
5. Untouched-surface confirmation: template refs still 12, `_test.go` occurrences
   still 291. (AC-EAS-014, AC-EAS-015)

Exit: all ACs PASS with observed evidence.

---

## G. Anti-Patterns

- **Copying the existing guard's walk root.** `os.Getwd()` under `go test` is the
  package dir. Copying it reproduces the exact defect this SPEC exists to fix.
- **Reading a `-run` selector's exit 0 as PASS.** A zero-match selector exits 0.
  Always `-v -count=1` and read the `--- PASS: <name>` line.
- **Treating the M2-M5 RED window as a regression.** It is the designed progress
  meter (D.3).
- **Declaring the guard proven by presence alone.** `grep -c` on the banned list
  shows tokens exist, not that the code path bites. The M6 re-introduction is the
  proof.
- **Grepping the guard file for bare `"ANTHROPIC_*"` literals.** The banned set is
  *derived from the constants* (F/M2 item 3), so a correct guard contains none. A
  source-grep completeness check is unsatisfiable-by-construction: it can only pass
  if the implementer duplicates all 9 bare literals in the guard, abandoning the
  derivation. Verify completeness at runtime (AC-EAS-006).
- **Copying `findRepoRoot` into `internal/config`.** The package already declares
  it at `internal/config/token_budget_guard.go:45`; the copy fails to compile and
  takes the whole package - and every `go test ./internal/config/` AC - with it.
  Reuse `findProjectRoot(t)` at `internal/config/required_checks_test.go:162`.
- **Reading `$?` after a pipeline.** `cmd | tail -5; echo $?` reports `tail`'s
  status, never `cmd`'s - it is `0` whether the command succeeded, failed, or was
  missing. Capture into a variable before the pipe:
  `cmd; rc=$?; echo "exit=$rc"`.
- **Excluding `_test.go` with `grep -v` after `grep -h`.** `-h` suppresses
  filenames, so the `-v` filter has nothing to match and is inert. Use
  `--exclude='*_test.go'`.
- **Excluding the SSOT by package or substring.** An exclusion on
  `internal/config` or `envkeys` exempts far more than the one SSOT file. Match the
  exact path; AC-EAS-007 plants in a sibling `internal/config` file to prove the
  exclusion is narrow.
- **Entering M1 on a stale census.** The counts are a `76d9a8f3b` snapshot. Run the
  section C.1 gate first.
- **Migrating `_test.go` literals.** Out of scope by record (spec.md section D);
  a well-meaning cleanup here contradicts CLAUDE.local.md section 14.
- **Touching `internal/template/templates/`.** Template neutrality; also would
  entail `make build`, which this SPEC explicitly does not.
- **Renaming or widening `TestNoBareGLMEnvVarLiteralsInCLIProduction`.** OQ-4
  recommends leaving it intact; its name encodes its scope.

---

## H. Cross-References

- `spec.md` sections A.2, A.3 - the guard defect and the SSOT-exclusion hazard
- `acceptance.md` - AC verdict commands with observed pre-change baselines
- `internal/cli/glm_env_parity_test.go:66` - the defective guard (unmodified)
- `internal/config/required_checks_test.go:162` - `findProjectRoot(t *testing.T) string`,
  the same-package helper the new guard reuses (verified by grep)
- `internal/config/token_budget_guard.go:45` - `findRepoRoot(start string) (string, bool)`,
  the pre-existing same-package symbol a copied helper would collide with (F/M2 hazard)
- `internal/spec/drift_doctrine_test.go:13` - `findRepoRoot(t)`; a *different-package*
  unexported helper, cited as design precedent only, NOT importable or copyable here
- `internal/sandbox/env.go:32` - the sole production `"ANTHROPIC_API_KEY"` site;
  the M6 falsification target
- `cmd/moai/main.go` - the M6 `cmd/` plant target for REQ-EAS-005 coverage
- CLAUDE.local.md section 14 (hardcoding policy), sections 15 and 25 (template
  neutrality), section 23 (PR-mandatory git policy)
