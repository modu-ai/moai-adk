---
spec: SPEC-ENVKEY-ANTHROPIC-SSOT-001
phase: plan
tier: M
---

# Acceptance Criteria - SPEC-ENVKEY-ANTHROPIC-SSOT-001

Every verdict command below that **can** run against the base tree
(`76d9a8f3b`, worktree `spec-envkey-anthropic`) was executed during plan-phase
authoring, and its observed pre-change output is recorded verbatim as the
baseline. An AC whose runnable command was never run is not an AC.

Commands that depend on artefacts M1-M5 create (the guard test, the new
constants, the post-transition state) **cannot** run at base. Those are marked
`not runnable at base` with the reason. That marking is the honest form; an
invented exit code would be a fabricated baseline
(`.claude/rules/moai/core/verification-claim-integrity.md` section 2).

## Standing verdict rules

1. **No vacuous `-run` selectors.** Any AC resting on `go test -run <pattern>`
   MUST use `-v -count=1` and read the verdict from an observed
   `--- PASS: <exact test name>` line. A zero-match selector exits 0 and would
   otherwise read as a false PASS.
2. **Presence is not reachability.** `grep -c` proves a token exists; it does not
   prove the code path bites. Guard ACs carry a falsification step.
3. **Scoped commands only.** Each grep states its directory roots, its
   `--include` filter, and its `_test.go` exclusion explicitly.
4. **`_test.go` exclusion via `--exclude`, never `grep -v` after `-h`.** `grep -h`
   suppresses filenames, so a downstream `grep -v '_test.go'` matches nothing and
   is inert - a latent trap that silently widens scope the moment a test file
   contributes a unique literal. Every grep in this file uses
   `--exclude='*_test.go'`.
5. **Exit codes are captured before a pipe, never after.** `cmd | tail -5; echo $?`
   yields `tail`'s status - `0` whether `cmd` succeeded, failed, or was absent.
   Verified: `bash -c 'false | tail -5; echo "exit=$?"'` prints `exit=0`. Every
   exit-code verdict here uses `cmd; rc=$?; echo "exit=$rc"`.
6. **A command that cannot run at base says so.** Some verdicts depend on artefacts
   that M1-M5 create (the guard test, the new constants). For those the baseline is
   recorded as *not runnable at base, and why* - never as an invented exit code. An
   honestly-absent baseline is evidence; a fabricated one is not.

---

## D. AC Matrix

| AC | REQ | Milestone | Subject |
|----|-----|-----------|---------|
| AC-EAS-001 | REQ-EAS-002 | M1 | Two missing constants added |
| AC-EAS-002 | REQ-EAS-003 | M1 | Prefix constant added and documented |
| AC-EAS-003 | REQ-EAS-001 | M1 | Constant set covers all 9 literal values |
| AC-EAS-004 | REQ-EAS-009 | M1 | Constant values byte-identical to literals |
| AC-EAS-005 | REQ-EAS-005 | M2 | Guard walk root reaches outside `internal/cli` |
| AC-EAS-006 | REQ-EAS-008 | M2 | Guard banned set covers all 9 names (runtime assertion) |
| AC-EAS-007 | REQ-EAS-006 | M2/M6 | SSOT exclusion is present **and narrow** (sibling file still scanned) |
| AC-EAS-008 | REQ-EAS-004 | M3 | `internal/cli` transitioned |
| AC-EAS-009 | REQ-EAS-004 | M4 | `internal/hook` transitioned |
| AC-EAS-010 | REQ-EAS-004 | M5 | Remaining packages transitioned |
| AC-EAS-011 | REQ-EAS-004 | M5 | Zero bare literals repo-wide; guard GREEN |
| AC-EAS-012 | REQ-EAS-007, REQ-EAS-005 | M6 | Guard falsifiable outside `internal/cli` **and** under `cmd/` |
| AC-EAS-013 | REQ-EAS-009 | M6 | Behaviour preserved: suite, vet, lint green |
| AC-EAS-014 | REQ-EAS-010 | M6 | Template surface untouched |
| AC-EAS-015 | REQ-EAS-011 | M6 | Test files untouched |

---

## AC-EAS-001 - Two missing constants added (M1)

**Given** `envkeys.go` defines 6 ANTHROPIC constants and two names in active use
(`ANTHROPIC_API_KEY`, `ANTHROPIC_DEFAULT_FABLE_MODEL`) have none,
**When** M1 lands,
**Then** both constants exist with the OQ-3 names.

```bash
grep -c 'EnvAnthropic[A-Za-z]* = "ANTHROPIC_' internal/config/envkeys.go
grep -n 'EnvAnthropicAPIKey = "ANTHROPIC_API_KEY"' internal/config/envkeys.go
grep -n 'EnvAnthropicDefaultFableModel = "ANTHROPIC_DEFAULT_FABLE_MODEL"' internal/config/envkeys.go
```

- **Baseline observed:** `6`; both `grep -n` return no match (exit 1).
- **Expected after M1:** `9`; both `grep -n` return exactly one line each.

## AC-EAS-002 - Prefix constant added and documented (M1)

**Given** `"ANTHROPIC_"` is used as a prefix at
`internal/cli/worktree/tmux_integration.go:238`,
**When** M1 lands,
**Then** `EnvAnthropicPrefix` exists and its doc comment identifies it as a
namespace prefix rather than a variable name.

```bash
grep -n -B3 'EnvAnthropicPrefix = "ANTHROPIC_"' internal/config/envkeys.go
```

- **Baseline observed:** no match (exit 1).
- **Expected after M1:** one match, preceded by a doc comment containing the word
  `prefix`.

## AC-EAS-003 - Constant set covers all 9 literal values (M1)

**Given** 9 distinct `ANTHROPIC_*` literal values are in production use,
**When** M1 lands,
**Then** every one has a constant, so the transition has no orphan.

```bash
# left: literals in use (production).  right: literals defined in envkeys.go
# NOTE: --exclude (not `grep -v` after -h) -- see standing rule 4.
comm -23 \
  <(grep -rho '"ANTHROPIC_[A-Z_]*"' internal/ pkg/ cmd/ --include='*.go' \
      --exclude='*_test.go' | sort -u) \
  <(grep -o '"ANTHROPIC_[A-Z_]*"' internal/config/envkeys.go | sort -u)
```

- **Baseline observed (re-run with the corrected `--exclude` form):**
  unique-literal count `9`; `envkeys.go` defines `6`; `comm -23` emits exactly the
  3 uncovered values:

  ```
  "ANTHROPIC_"
  "ANTHROPIC_API_KEY"
  "ANTHROPIC_DEFAULT_FABLE_MODEL"
  ```
- **Expected after M1:** `comm -23` emits **nothing** (empty output).
- **Correction note (S5):** the prior form piped `grep -rho` into
  `grep -v '_test.go'`. `-h` suppresses filenames, so that filter matched nothing
  and was inert. Both forms happen to yield `9` today (verified: test files
  contribute no *unique* literal), so the defect had no present effect - but it
  would silently widen scope the first time a test file introduced a new value.

## AC-EAS-004 - Constant values byte-identical (M1)

**Given** this is a name-indirection refactor,
**When** M1 lands,
**Then** no constant's value differs from the literal it replaces.

```bash
go build ./...; build_rc=$?; echo "build_exit=$build_rc"

# Mechanical: every value ADDED to envkeys.go must already exist in the base-tree
# production inventory. `comm -13` emits lines present ONLY on the right (added
# values with no base-tree counterpart) -- i.e. an invented or typo'd value.
comm -13 \
  <(git grep -ho -E '"ANTHROPIC_[A-Z_]*"' 76d9a8f3b \
      -- 'internal/*.go' 'pkg/*.go' 'cmd/*.go' ':(exclude)*_test.go' | sort -u) \
  <(git diff 76d9a8f3b -- internal/config/envkeys.go \
      | grep '^+' | grep -oE '"ANTHROPIC_[A-Z_]*"' | sort -u)
```

- **Baseline observed:** `build_exit=0`. The `comm -13` produced **empty output**
  (the right-hand side is empty at base - no diff against `76d9a8f3b` yet - so the
  command is trivially satisfied; it becomes load-bearing only after M1). The
  base-tree left-hand inventory was confirmed to be exactly the 9 known values.
- **Expected after M1:** `build_exit=0`; `comm -13` still emits **nothing**. A
  non-empty line names a constant whose value does not correspond to any literal
  actually in use at base - a typo or an invented name.
- **Correction note (S6):** the prior verdict was "manual read of 3 added lines",
  a judgement call rather than a binary command. `git grep` at the base rev pins
  the left-hand inventory to `76d9a8f3b` so the check stays meaningful after M3-M5
  have driven the working-tree literal count to zero.

---

## AC-EAS-005 - Guard walk root reaches outside `internal/cli` (M2)

This is the AC that closes the load-bearing defect. It is deliberately verified
by **natural RED against the un-transitioned tree**, because at M2 the 37
out-of-CLI literals form a ready-made falsification corpus.

**Given** the existing guard walks only `internal/cli` and 37 occurrences live
outside it,
**When** the new repo-root guard runs at M2 against the still-untransitioned
tree,
**Then** it fails, and its offender list contains paths from `internal/hook/`,
`internal/tmux/`, `internal/statusline/`, and `internal/sandbox/`.

```bash
go test ./internal/config/ -run 'TestNoBareAnthropicEnvVarLiteralsInProduction' \
  -v -count=1 > /tmp/m2-guard.log 2>&1; echo "guard_exit=$?"
grep -E '^(--- (PASS|FAIL)|FAIL)' /tmp/m2-guard.log
# reachability assertion: offenders outside internal/cli must be present
grep -c 'internal/hook/\|internal/tmux/\|internal/statusline/\|internal/sandbox/' /tmp/m2-guard.log
```

- **Baseline observed (base tree, guard absent):** the test does not exist; a
  `-run` against it would match nothing and exit 0 — precisely the vacuous-PASS
  trap standing rule 1 forbids. The reference point instead is the **existing**
  guard, observed at base:

  ```
  === RUN   TestNoBareGLMEnvVarLiteralsInCLIProduction
  --- PASS: TestNoBareGLMEnvVarLiteralsInCLIProduction (0.05s)
  ok  	github.com/modu-ai/moai-adk/internal/cli	3.776s
  ```

  That PASS is green on a tree holding 83 violations — the defect made visible.
- **Expected at M2:** `guard_exit` non-zero, an observed
  `--- FAIL: TestNoBareAnthropicEnvVarLiteralsInProduction` line, and the
  reachability grep returning **>= 37**.
- **Fails the AC if:** the guard reports FAIL but every offender path starts with
  `internal/cli/` — that is the old scope bug reproduced.
- **Threshold provenance (S-census).** The `>= 37` figure is
  `83 (total) - 46 (internal/cli)`, both measured at `76d9a8f3b`. Per-name
  out-of-CLI counts, verified individually: `ANTHROPIC_AUTH_TOKEN` 12,
  `ANTHROPIC_BASE_URL` 8, `ANTHROPIC_DEFAULT_OPUS_MODEL` 6,
  `ANTHROPIC_DEFAULT_HAIKU_MODEL` 5, `ANTHROPIC_DEFAULT_SONNET_MODEL` 5,
  `ANTHROPIC_API_KEY` 1, and `ANTHROPIC_` / `ANTHROPIC_DEFAULT_FABLE_MODEL` /
  `ANTHROPIC_REASONING_EFFORT` 0 each — sum 37. If the plan.md section C.1 census
  gate reports different totals at run-phase entry, this threshold MUST be
  recomputed before M2 or the assertion is unattributable.
- **Redirect note:** `tee ... | tail -40` was replaced by a plain redirect plus a
  scoped grep. `tail -40` can crop the `--- FAIL:` line below a long offender list,
  and its exit status is `tail`'s, not the test's (standing rule 5).

## AC-EAS-006 - Guard banned set covers all 9 names, asserted at runtime (M2)

**Given** REQ-EAS-008 requires the banned set to track `envkeys.go`, and the
banned set is **derived from the constants** (plan.md F/M2 item 3) so the guard
file contains no bare literals to grep,
**When** M2 lands,
**Then** a **top-level test function named exactly
`TestAnthropicBannedSetCoversAllNames`** asserts the set has exactly 9 members
and that each of the 9 expected names is present.

Implementation shape (plan.md F/M2 item 4) - the guard exposes
`bannedAnthropicEnvNames`, and a table-driven **top-level test function** in the
same package asserts the following. It MUST NOT be a `t.Run` sub-test: the
verdict command below selects it with a bare top-level `-run` pattern, which
cannot select a sub-test, so a sub-test implementation would report
`[no tests to run]` and trip this AC's own fail-condition.

```go
if len(bannedAnthropicEnvNames) != 9 { t.Fatalf("banned set size = %d, want 9", ...) }
for _, want := range []string{
    EnvAnthropicPrefix, EnvAnthropicAPIKey, EnvAnthropicAuthToken,
    EnvAnthropicBaseURL, EnvAnthropicDefaultFableModel,
    EnvAnthropicDefaultHaikuModel, EnvAnthropicDefaultOpusModel,
    EnvAnthropicDefaultSonnetModel, EnvAnthropicReasoningEffort,
} { /* assert membership; t.Errorf naming the missing name */ }
```

```bash
go test ./internal/config/ -run 'TestAnthropicBannedSetCoversAllNames' \
  -v -count=1 2>&1 | grep -E '^(=== RUN|--- (PASS|FAIL)|ok|FAIL)'
```

- **Baseline: not runnable at base.** The guard file
  `internal/config/anthropic_env_ssot_test.go` does not exist at `76d9a8f3b`, so
  that test function cannot run. Per standing rule 1 the `-run` selector would match
  nothing and exit 0 - **observed verbatim** on this tree with an absent selector:

  ```
  testing: warning: no tests to run
  PASS
  ok  	github.com/modu-ai/moai-adk/internal/config	0.359s [no tests to run]
  ```

  That is the vacuous-PASS trap, not a baseline. No exit code is claimed here.
- **Expected after M2:** an observed
  `--- PASS: TestAnthropicBannedSetCoversAllNames` line.
- **Fails the AC if:** the run reports `[no tests to run]`. Exit 0 with no
  `--- PASS:` line is a vacuous pass, not a verdict.
- **Correction note (M2, two defects in the prior form):**
  1. *Fabricated baseline.* The prior text claimed "guard file does not exist;
     command errors (exit 2)". Executed on this tree, the `comm` pipeline in fact
     **exits 0 and prints 6 lines** - `comm` treats a missing right-hand file as
     the empty set, so all 6 defined names are reported as unbanned. Observed:

     ```
     "ANTHROPIC_AUTH_TOKEN"
     "ANTHROPIC_BASE_URL"
     "ANTHROPIC_DEFAULT_HAIKU_MODEL"
     "ANTHROPIC_DEFAULT_OPUS_MODEL"
     "ANTHROPIC_DEFAULT_SONNET_MODEL"
     "ANTHROPIC_REASONING_EFFORT"
     exit=0
     ```

     The recorded baseline did not reproduce.
  2. *Design contradiction - a correct implementation failed the AC.* plan.md
     requires the banned set be derived from the `envkeys.go` constants. A derived
     guard references `EnvAnthropic*` identifiers and therefore contains **zero**
     bare `"ANTHROPIC_*"` literals, so the prior `comm -23` would emit all 9 names
     and FAIL. The AC could only pass if the implementer duplicated 9 bare literals
     into the guard, abandoning the derivation. The runtime assertion above removes
     the contradiction: it verifies the property REQ-EAS-008 actually asserts
     (set completeness) rather than a source-text proxy that contradicts the design.
- **Coverage note (S2).** This assertion is the *only* check covering
  `ANTHROPIC_`, `ANTHROPIC_DEFAULT_FABLE_MODEL`, and `ANTHROPIC_REASONING_EFFORT`
  - the three names with **zero** out-of-CLI offenders (verified per-name counts:
  0, 0, 0; the other six total 37). Without it, a guard silently omitting all
  three still yields exactly 37 out-of-CLI offenders and passes AC-EAS-005,
  AC-EAS-011, and AC-EAS-012 alike.

## AC-EAS-007 - SSOT exclusion is present AND narrow (M2, re-checked M6)

**Given** a repo-root walk reaches `internal/config/envkeys.go` (6 bare literals
by design) and 291 `_test.go` occurrences, and an *over-broad* exclusion
(`strings.Contains(path, "internal/config")` or `"envkeys"`) would exempt the
entire SSOT package while looking identical from the outside,
**When** the guard runs on a clean tree and then on a tree with a literal planted
in a **sibling** `internal/config` production file,
**Then** the clean run is GREEN (exclusion present) and the planted run is RED
naming the sibling file (exclusion narrow).

```bash
# (a) exclusion present -- guard GREEN with the excluded surfaces populated
go test ./internal/config/ -run 'TestNoBareAnthropicEnvVarLiteralsInProduction' \
  -v -count=1 2>&1 | grep -E '^(=== RUN|--- (PASS|FAIL)|ok|FAIL)'
grep -c '"ANTHROPIC_[A-Z_]*"' internal/config/envkeys.go        # excluded surface still populated
grep -rn '"ANTHROPIC_[A-Z_]*"' internal/ pkg/ cmd/ --include='*_test.go' | wc -l

# (b) exclusion NARROW -- plant one bare literal in a SIBLING internal/config
#     production file and require RED naming that file.
#     Target: internal/config/log.go (verified present, non-envkeys, package config)
printf '\n// scope-probe (AC-EAS-007b): revert immediately\nvar _ = "ANTHROPIC_API_KEY"\n' \
  >> internal/config/log.go
go test ./internal/config/ -run 'TestNoBareAnthropicEnvVarLiteralsInProduction' \
  -v -count=1 2>&1 | tee /tmp/ac007b.log | grep -E '^(--- (PASS|FAIL)|FAIL)|log\.go'
# Two SEPARATE observations are required, in this order, so that a build failure
# in package config (whose error lines also name log.go) cannot masquerade as a
# guard detection:
grep -c -- '--- FAIL: TestNoBareAnthropicEnvVarLiteralsInProduction' /tmp/ac007b.log  # MUST be 1
grep -c 'anthropic_env_ssot_test.go' /tmp/ac007b.log   # MUST be >= 1 (the guard's own
                                                       # t.Fatalf site -- present only
                                                       # when the guard actually ran)

# (c) revert and prove the tree is byte-clean again
git checkout -- internal/config/log.go
git status --porcelain internal/config/log.go        # MUST be empty
go test ./internal/config/ -run 'TestNoBareAnthropicEnvVarLiteralsInProduction' \
  -v -count=1 2>&1 | grep -E '^(--- (PASS|FAIL)|ok|FAIL)'
```

- **Baseline (the runnable legs, observed at base):** `envkeys.go` holds `6`
  literals; `_test.go` files hold `291`.
- **Baseline (the guard legs): not runnable at base** - the guard does not exist
  at `76d9a8f3b`; a `-run` against it yields `[no tests to run]` (verbatim output
  recorded under AC-EAS-006), which is a vacuous pass, not a baseline.
- **Expected at M6:**
  - (a) `--- PASS: TestNoBareAnthropicEnvVarLiteralsInProduction`; counts still
    `6` and `291`
  - (b) `--- FAIL: ...` **and** an offender line containing `internal/config/log.go`
  - (c) `git status --porcelain` empty; `--- PASS: ...` again
- **Fails the AC if:** step (b) passes. A guard that ignores a bare literal in
  `internal/config/log.go` has an exclusion scoped to the package or to the
  substring `envkeys`, not to the single SSOT file - it would exempt the entire
  SSOT package while every other AC still passed.
- **Correction note (S1).** The prior verdict grepped `go test -v` output for
  `envkeys.go\|_test.go` expecting `0`. On the GREEN path `go test -v` emits only
  `=== RUN` / `--- PASS` / `ok <pkg>` lines - no filenames - so it returns `0`
  unconditionally, whether the exclusion is correct, over-broad, or absent
  entirely. Worse, its risk was inverted: on a FAIL the runner prints
  `anthropic_env_ssot_test.go:NN:`, which *contains* `_test.go`, producing a false
  exclusion violation. No AC constrained the exclusion's scope, so an
  implementation exempting all of `internal/config` passed all 15 ACs. Step (b) is
  the constraint that was missing.

---

## AC-EAS-008 - `internal/cli` transitioned (M3)

**Given** `internal/cli` holds 46 of the 83 bare literals,
**When** M3 lands,
**Then** none remain there and the package's tests still pass.

```bash
grep -rn '"ANTHROPIC_[A-Z_]*"' internal/cli/ --include='*.go' --exclude='*_test.go' | wc -l
go test ./internal/cli/... -count=1 2>&1 | grep -c '^FAIL'
```

- **Baseline observed:** literal count `46`; `FAIL` count `0` (executed at base:
  every `internal/cli/...` package reported `ok`).
- **Expected after M3:** `0`; `FAIL` count still `0`.
- **Correction note (S4/S5):** the test leg had no recorded baseline and used
  `tail -5`, which truncates the package list and can hide a `FAIL` line above the
  window. `grep -c '^FAIL'` is a deterministic verdict over the whole output.

## AC-EAS-009 - `internal/hook` transitioned (M4)

**Given** `internal/hook` holds 29 bare literals,
**When** M4 lands,
**Then** none remain there and the package's tests still pass.

```bash
grep -rn '"ANTHROPIC_[A-Z_]*"' internal/hook/ --include='*.go' --exclude='*_test.go' | wc -l
go test ./internal/hook/... -count=1 2>&1 | grep -c '^FAIL'
```

- **Baseline observed:** literal count `29`; `FAIL` count `0` (executed at base).
- **Expected after M4:** `0`; `FAIL` count still `0`.

## AC-EAS-010 - Remaining packages transitioned (M5)

**Given** `internal/tmux`, `internal/statusline`, and `internal/sandbox` hold the
last 8 bare literals,
**When** M5 lands,
**Then** none remain and those packages' tests still pass.

```bash
grep -rn '"ANTHROPIC_[A-Z_]*"' internal/tmux/ internal/statusline/ internal/sandbox/ \
  --include='*.go' --exclude='*_test.go' | wc -l
go test ./internal/tmux/... ./internal/statusline/... ./internal/sandbox/... \
  -count=1 2>&1 | grep -c '^FAIL'
```

- **Baseline observed:** literal count `8`; `FAIL` count `0` (executed at base).
- **Expected after M5:** `0`; `FAIL` count still `0`.

## AC-EAS-011 - Zero bare literals repo-wide; guard GREEN (M5)

**Given** M3-M5 have transitioned all 10 files,
**When** the guard runs,
**Then** the RED window from plan.md D.3 closes.

```bash
grep -rn '"ANTHROPIC_[A-Z_]*"' internal/ pkg/ cmd/ --include='*.go' --exclude='*_test.go' \
  | grep -v '^internal/config/envkeys.go' | wc -l
go test ./internal/config/ -run 'TestNoBareAnthropicEnvVarLiteralsInProduction' \
  -v -count=1 2>&1 | grep -E '^(=== RUN|--- (PASS|FAIL)|ok|FAIL)'
```

- **Baseline observed:** literal count `83` (re-run with the corrected
  `--exclude` form; identical to the prior form's value). Guard leg **not runnable
  at base** - the test does not exist at `76d9a8f3b`.
- **Expected after M5:** `0`, and an observed
  `--- PASS: TestNoBareAnthropicEnvVarLiteralsInProduction` line.
- **Fails the AC if:** the guard leg prints `[no tests to run]` - vacuous pass
  (standing rule 1). `tail -6` was replaced because it can crop the `--- PASS:`
  line out of view when the runner emits build or cache chatter.

---

## AC-EAS-012 - Guard falsifiable outside `internal/cli` and under `cmd/` (M6)

The decisive AC. A green guard proves nothing unless it can be made to fail on
demand, in every region REQ-EAS-005 claims.

> **[HARD] Precondition - M1 through M5 MUST be committed before M6 runs.**
> Steps (b)-(e) revert their probes with `git checkout -- <file>`, which
> restores the file to its last **commit**, not to its pre-probe working-tree
> state. If any M1-M5 transition of `internal/sandbox/env.go` (or
> `cmd/moai/main.go`) is still uncommitted when M6 runs, the checkout silently
> discards that transition and permanently reintroduces the bare literal --
> and `git status --porcelain` then reports **empty**, so the AC's own
> clean-revert proof would affirm the corruption. Verify before starting M6:
> `git status --porcelain internal/sandbox/env.go cmd/moai/main.go` MUST be
> empty, i.e. every transition is already committed.

**Given** the tree is clean, all M1-M5 work is committed, and the guard is GREEN,
**When** a bare literal is deliberately re-introduced (i) in `internal/sandbox`
(outside `internal/cli`, where the old guard was blind) and (ii) in `cmd/` (a root
with zero natural offenders), and each is then reverted,
**Then** the guard fails naming the planted file each time, passes again after
each revert, and the working tree is byte-identical to its pre-probe state.

```bash
# (a) GREEN before
go test ./internal/config/ -run 'TestNoBareAnthropicEnvVarLiteralsInProduction' \
  -v -count=1 2>&1 | grep -E '^(--- (PASS|FAIL)|ok|FAIL)'

# (b) plant OUTSIDE internal/cli. Target verified by grep: "ANTHROPIC_API_KEY"
#     has exactly ONE production occurrence repo-wide --
#     internal/sandbox/env.go:32, inside the defaultDenyList string slice.
#     After M5 that element reads `config.EnvAnthropicAPIKey`. The probe below
#     appends an equivalent bare literal to the same file rather than editing
#     the slice element in place: appending is executable verbatim and reverts
#     cleanly, whereas an in-place edit cannot be expressed as a command.
printf '\n// reachability probe (AC-EAS-012b): revert immediately\nvar _ = "ANTHROPIC_API_KEY"\n' \
  >> internal/sandbox/env.go
go test ./internal/config/ -run 'TestNoBareAnthropicEnvVarLiteralsInProduction' \
  -v -count=1 2>&1 | grep -E '^(--- (PASS|FAIL)|FAIL)|sandbox/env\.go'

# (c) revert; prove no residual diff; re-confirm GREEN
git checkout -- internal/sandbox/env.go
git status --porcelain internal/sandbox/env.go            # MUST be empty
git diff --stat internal/sandbox/env.go                   # MUST be empty
go test ./internal/config/ -run 'TestNoBareAnthropicEnvVarLiteralsInProduction' \
  -v -count=1 2>&1 | grep -E '^(--- (PASS|FAIL)|ok|FAIL)'

# (d) plant under cmd/ -- REQ-EAS-005 coverage proof (pkg/ + cmd/ hold 0 literals,
#     so nothing there is exercised by the natural corpus).
printf '\n// reachability probe (AC-EAS-012d): revert immediately\nvar _ = "ANTHROPIC_API_KEY"\n' \
  >> cmd/moai/main.go
go test ./internal/config/ -run 'TestNoBareAnthropicEnvVarLiteralsInProduction' \
  -v -count=1 2>&1 | grep -E '^(--- (PASS|FAIL)|FAIL)|cmd/moai/main\.go'
git checkout -- cmd/moai/main.go
git status --porcelain cmd/moai/main.go                   # MUST be empty
go test ./internal/config/ -run 'TestNoBareAnthropicEnvVarLiteralsInProduction' \
  -v -count=1 2>&1 | grep -E '^(--- (PASS|FAIL)|ok|FAIL)'

# (f) plant under pkg/ -- the third enumerated root. plan.md F/M2 item 6 states
#     these are three ENUMERATED roots that fail independently, so step (d)'s
#     cmd/ plant proves nothing about pkg/. Target verified: pkg/version/version.go
#     (package version, no build constraint, ends at package scope).
printf '\n// reachability probe (AC-EAS-012f): revert immediately\nvar _ = "ANTHROPIC_API_KEY"\n' \
  >> pkg/version/version.go
go test ./internal/config/ -run 'TestNoBareAnthropicEnvVarLiteralsInProduction' \
  -v -count=1 2>&1 | grep -E '^(--- (PASS|FAIL)|FAIL)|pkg/version/version\.go'
git checkout -- pkg/version/version.go
git status --porcelain pkg/version/version.go             # MUST be empty
go test ./internal/config/ -run 'TestNoBareAnthropicEnvVarLiteralsInProduction' \
  -v -count=1 2>&1 | grep -E '^(--- (PASS|FAIL)|ok|FAIL)'

# (e) per-name falsification table -- every one of the 9 banned names, in turn.
#     Three names (ANTHROPIC_, ANTHROPIC_DEFAULT_FABLE_MODEL,
#     ANTHROPIC_REASONING_EFFORT) have ZERO out-of-CLI offenders (verified counts
#     0/0/0), so no other AC exercises them.
for NAME in ANTHROPIC_ ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN ANTHROPIC_BASE_URL \
            ANTHROPIC_DEFAULT_FABLE_MODEL ANTHROPIC_DEFAULT_HAIKU_MODEL \
            ANTHROPIC_DEFAULT_OPUS_MODEL ANTHROPIC_DEFAULT_SONNET_MODEL \
            ANTHROPIC_REASONING_EFFORT; do
  printf '\nvar _ = "%s"\n' "$NAME" >> internal/sandbox/env.go
  go test ./internal/config/ -run 'TestNoBareAnthropicEnvVarLiteralsInProduction' \
    -v -count=1 >/dev/null 2>&1; rc=$?
  git checkout -- internal/sandbox/env.go
  echo "$NAME guard_exit=$rc"      # every line MUST show a NON-ZERO exit
done
git status --porcelain internal/sandbox/env.go            # MUST be empty
```

- **Baseline: not runnable at base.** The guard does not exist at `76d9a8f3b`, and
  after M5 the plant targets hold constants rather than literals. No exit code is
  claimed for any leg (standing rule 6).
- **Expected at M6:**
  - (a) `--- PASS: TestNoBareAnthropicEnvVarLiteralsInProduction`
  - (b) `--- FAIL: ...` **and** an offender line containing
    `internal/sandbox/env.go`
  - (c) both git checks empty; `--- PASS: ...`
  - (d) `--- FAIL: ...` **and** an offender line containing `cmd/moai/main.go`;
    then git check empty and `--- PASS: ...`
  - (e) all 9 lines show a non-zero `guard_exit`; final git check empty
- **Fails the AC if:** any of (b), (d), or any row of (e) passes. A guard that
  cannot be made to fail in `internal/sandbox/`, under `cmd/`, or on any single
  banned name has not covered the surface REQ-EAS-005 and REQ-EAS-008 claim,
  regardless of what step (a) showed.
- **Evidence requirement:** all outputs recorded verbatim in `progress.md`
  section E.2, including the `git status --porcelain` empties. A summary
  ("falsification confirmed") is not evidence.
- **Correction note (M4).** The prior text targeted
  `internal/hook/session_start.go`, instructing the implementer to "replace one
  `config.EnvAnthropicAPIKey` reference with the bare literal". Verified by grep:
  that file contains **no** `ANTHROPIC_API_KEY` reference at all - its ANTHROPIC
  references are `ANTHROPIC_AUTH_TOKEN` (lines 365, 380), `ANTHROPIC_BASE_URL`
  (381-382), and the three `ANTHROPIC_DEFAULT_*_MODEL` names (350-352, 435). There
  would have been nothing to replace; the implementer would have had to *insert* an
  unused literal, risking a lint failure that collides with AC-EAS-013. The retarget
  to `internal/sandbox/env.go:32` uses the sole real `ANTHROPIC_API_KEY` site,
  which is also outside `internal/cli/` and sits inside an existing string slice
  (no unused identifier on revert).
- **Correction note (S2).** Step (e) closes the banned-set-subset hole: the three
  zero-offender names previously had no falsification coverage anywhere. It
  composes with AC-EAS-006's runtime membership assertion - (e) proves the guard
  *acts* on each name, AC-EAS-006 proves the set *contains* each name.
- **Correction note (S3/N3).** Steps (d) and (f) are the evidence for the `cmd/`
  and `pkg/` roots of REQ-EAS-005's claimed surface -- one plant per root, because
  plan.md F/M2 item 6 enumerates the three roots as independently-failing list
  elements, so a plant in one root proves nothing about another. The `git status --porcelain` /
  `git diff --stat` checks in (c), (d), and (e) are the residual-diff proof that
  every deliberately-planted literal was actually reverted - previously asserted
  by the `git checkout` line alone, with nothing verifying it took effect.

## AC-EAS-013 - Behaviour preserved (M6)

**Given** this is a name-indirection refactor with no intended behaviour change,
**When** M6 runs the full toolchain,
**Then** the suite, `go vet`, and `golangci-lint` are all clean.

```bash
go test ./... -count=1 2>&1 | grep -c '^FAIL'
go vet ./... 2>&1; vet_rc=$?; echo "vet_exit=$vet_rc"
golangci-lint run --timeout=3m; lint_rc=$?; echo "lint_exit=$lint_rc"
```

- **Baseline observed:**
  - `go vet ./...` produced no output, `vet_exit=0`.
  - `golangci-lint run --timeout=3m` printed `0 issues.`, **`lint_exit=0`**
    (executed at base with the corrected capture form; `golangci-lint` resolved to
    `/opt/homebrew/bin/golangci-lint`).
  - The **full** `go test ./...` leg was **not run at base**. It was deliberately
    skipped: the repository's package-level suites are individually slow (several
    `internal/cli/...` packages exceed 5 s each) and a whole-repo run risks a
    tool timeout that would leave no observed output at all. In its place the four
    packages this SPEC actually touches were run and their `^FAIL` counts observed:
    `internal/cli/...` `0`, `internal/hook/...` `0`,
    `internal/tmux/... internal/statusline/... internal/sandbox/...` `0`,
    `internal/config/...` `0`. This is stated rather than fabricated
    (standing rule 6); the whole-repo leg's first real observation is at M6.
- **Expected at M6:** `FAIL` count `0`; `vet_exit=0`; `lint_exit=0`.
- **Correction note (M3).** The prior lint leg read
  `golangci-lint run --timeout=3m 2>&1 | tail -5; echo "lint_exit=$?"`. `$?` after
  a pipeline is **`tail`'s** status, never `golangci-lint`'s, so the AC passed
  whether lint succeeded, failed, or the binary was missing. Proof observed on this
  machine: `bash -c 'false | tail -5; echo "lint_exit=$?"'` prints `lint_exit=0`,
  while `bash -c 'false; lint_exit=$?; echo "lint_exit=$lint_exit"'` prints
  `lint_exit=1`. The capture-before-pipe form is now standing rule 5. The
  `go test` leg's redundant intermediate `grep -E` was also dropped - `grep -c
  '^FAIL'` alone is the verdict.

## AC-EAS-014 - Template surface untouched (M6)

**Given** REQ-EAS-010 forbids any edit under `internal/template/templates/`,
**When** M6 checks the template surface,
**Then** both the reference count and the changed-file count are unchanged.

```bash
grep -rn 'ANTHROPIC_' internal/template/templates/ | wc -l
git diff --name-only 76d9a8f3b -- internal/template/templates/ | wc -l
```

- **Baseline observed:** `12` references; changed-file count `0` (both legs
  executed at base - the second leg previously had no recorded baseline).
- **Expected at M6:** still `12`; changed-file count still `0`.

## AC-EAS-015 - Test files untouched (M6)

**Given** `_test.go` is a declared hardcoding-allowed zone and REQ-EAS-011 forbids
migrating its literals,
**When** M6 counts them,
**Then** the count is unchanged.

```bash
grep -rn '"ANTHROPIC_[A-Z_]*"' internal/ pkg/ cmd/ --include='*_test.go' | wc -l
```

- **Baseline observed:** `291`.
- **Expected at M6:** `291` (unchanged). A **lower** number means the
  out-of-scope boundary in spec.md section D was crossed and the run must be
  corrected, not accepted.

---

## Definition of Done

- [ ] plan.md section C.1 census gate run BEFORE M1; all four counts matched the
      recorded values, or the SPEC was re-baselined and the HISTORY table updated
- [ ] All 15 ACs PASS with observed, verbatim-recorded evidence
- [ ] AC-EAS-005 (natural RED reachability) and AC-EAS-012 (deliberate
      re-introduction) both recorded — the guard is proven to see the wider
      surface **and** to still bite once clean
- [ ] AC-EAS-006 runtime banned-set assertion and AC-EAS-012 step (e) per-name
      table both recorded — the three zero-offender names
      (`ANTHROPIC_`, `ANTHROPIC_DEFAULT_FABLE_MODEL`, `ANTHROPIC_REASONING_EFFORT`)
      are covered
- [ ] AC-EAS-007 step (b) recorded — the SSOT exclusion is proven narrow, not a
      package-wide exemption
- [ ] AC-EAS-012 step (d) recorded — `cmd/` reachability proven by plant, closing
      REQ-EAS-005's otherwise-unexercised `pkg/`+`cmd/` claim
- [ ] Every deliberate plant reverted with an empty `git status --porcelain`
      recorded as proof
- [ ] `go test ./...`, `go vet ./...`, `golangci-lint run` all green
- [ ] Untouched surfaces confirmed: template refs `12`, `_test.go` refs `291`
- [ ] All milestones landed in a single PR; `main` never observed the RED window
- [ ] OQ-1..OQ-4 resolutions reflected in the implementation as approved
