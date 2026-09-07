# SPEC-CODEX-SKILL-PATH-READBACK-001 — Implementation Plan

Card t562 · worktree `.claude/worktrees/t562` · branch `WT-codex-read-inverse` · base `bce6d7e08`
(= origin/develop at entry, 2026-09-08; dated anchor, NOT an AC range edge) · Tier M · Class C ·
cycle_type **tdd** (`quality.yaml development_mode: tdd`, read this run).

Milestones are ordered by decision-reversibility: the conversion-site design (which branch converts)
is the card's most arguable decision, so M1 lands the tests that pin it before M2 writes the one-line
implementation. M3 is mechanical and sits last deliberately.

## §A. Context

`spec.md §B` carries the measured facts (M1-M7) and their addresses; they are not restated here.
Three things are worth repeating at the top of the plan:

1. **M1's absorb merge is a gate.** The seam does not exist in this tree until
   `git merge WT-codex-path-escape` lands; nothing after M1 compiles against it before that.
2. **The card's entire production diff is two one-line conversions** (one per reader, absolute
   branch only) — see `spec.md §D.2`. Everything else in this card is tests, guards, pins, and
   evidence. A run that finds itself writing more production code than that has left scope.
3. **The doctor half is structurally verified only** (`spec.md §D.3`). Do not manufacture symmetry;
   the seam at `doctor_codex.go` ~:857 is t563's, and adding it FAILS the DoD.

## §B. Known Issues Entering the Card

- `codexStaleSkillFinding` has ZERO existing tests (t540 measured: `grep codexStaleSkillFinding
  --include='*_test.go'` → rc=1). The regression guard (AC-CSRB-006) is a NEW test whose expected
  values are derived from pre-change behaviour and must not change.
- `doctor_codex.go` stats directly at two sites (~:459 `inspectSkillMirror` — mirror-walk results,
  not config declarations; ~:857 `codexStaleSkillFinding` — the config consumer). Only ~:857 is in
  scope; the scope-out rationale for ~:459 is `spec.md §F`.
- `classifyCodexSkillPath` is GOOS-fixed and not injectable; its Windows branch stays inference I-1.
- The mixed home-relative form residual (`spec.md §G` gap 5) is known, pre-existing, and out of scope.

## §C. Pre-flight

```bash
git rev-parse --show-toplevel        # must be the t562 worktree
git branch --show-current            # WT-codex-read-inverse
git rev-parse --short HEAD
git fetch origin develop
git merge-base origin/develop HEAD   # the card's live left edge (re-derived at every AC read)

# BASE test count for AC-CSRB-010's executed-test control — captured ONCE, BEFORE the absorb merge,
# with the same command the AFTER count will use:
go test ./internal/cli/... -timeout 1800s -v > .moai/reports/t562/ac-010-base.log 2>&1; echo "rc=$?"
/usr/bin/grep -c -- '--- PASS: ' .moai/reports/t562/ac-010-base.log
```

## §D. Constraints

Carried from `spec.md §E`; the ones that bind commands:

- **Test scope.** `go test ./internal/cli/... -timeout 1800s` only. Never `go test ./...` locally;
  the full-suite verdict is CI's, on the pushed head.
- **Isolated config.** Every fixture under `t.TempDir()`; `CODEX_HOME` and the `codexUserHomeDir`
  seam pointed at the temp tree. The real `~/.codex/config.toml` is never written, and no test
  performs a filesystem deletion.
- **Seam-override discipline.** Both `osStatFn` AND `configPathSeparator` are package-level vars:
  every override uses `t.Cleanup` restore and the test MUST NOT call `t.Parallel()` (the M5
  discipline, extended to the new seam).
- **Absence claims** use `/usr/bin/grep` with a positive control on a sibling file.
- **Commit messages** all contain `t562`. **Evidence** under `.moai/reports/t562/`.
- **Frozen branch is read-only**: `git show WT-codex-path-escape:<path>` only; the M1 merge is the
  single write-shaped interaction with it.

## §E. Milestones

### M1 — Absorb t540 and establish the RED (Priority: High, blocks everything)

- `git merge WT-codex-path-escape` into `WT-codex-read-inverse` (the run-phase absorb; NO push).
- Verify the seam arrived and the tree builds: AC-CSRB-001 (`go build ./...` rc=0; the three seam
  symbols present in `internal/cli/codex_config_path.go`; the publisher conversions present at
  `codex_skills_disable.go` — locate by symbol, not line).
- Write the prune-side tests in `internal/cli/codex_skills_prune_test.go` (or a sibling `_test.go`
  in the same package — either is fine; keep the seam discipline header comment):
  - **AC-CSRB-002 (the behavioral RED)**: `configPathSeparator = '\\'` override; `osStatFn`
    replaced by a recorder capturing its argument; entry declares a HOST-absolute slash path
    (`filepath.Join(t.TempDir(), "gone", "SKILL.md")`) so `classifyCodexSkillPath` returns
    `codexPathAbsolute` on darwin. Assert the recorder observed
    `fromConfigPath(declared, '\\')` (i.e. the backslash-converted form).
  - **AC-CSRB-003 (ordering guard)**: entry declares `C:/Users/u/SKILL.md` with the separator
    override; assert verdict SkipReason is the RELATIVE skip and the recorder counted 0 stat calls
    (classification ran on the declared form; the forbidden convert-first order would skip as
    oddly-formed instead — the two reasons are distinct strings, so the ordering is observable).
  - **AC-CSRB-004 (no-double-conversion guard)**: entry declares `~/x/SKILL.md`; `stubHome(t, tmp)`;
    separator override; assert the recorder observed `filepath.Join(tmp, "x/SKILL.md")` EXACTLY.
  - **AC-CSRB-005 (eligibility contract pins)**: absolute declaration + recorder returns
    `fs.ErrNotExist` → `Eligible: true`; recorder returns nil → skip "the path resolves"; backslash
    declaration `C:\Users\u\SKILL.md` on PRODUCTION separators → `codexPathOddlyFormed` skip,
    stat count 0.
- Observe the RED: AC-CSRB-002 FAILS before the conversion exists (recorder sees the declared form).
  AC-CSRB-003/004/005 are expected GREEN before (guards; they claim no RED) — their discriminating
  power is established by the M3 mutants instead.
- **AC-CSRB-006 baseline (guard authored and captured HERE, in the M1 window)**: author the doctor
  regression-guard test NOW — `t.Setenv("CODEX_HOME", tmp)` + a config fixture declaring an absolute
  path to an EXISTING file — and record its passing execution as the baseline (expected counter
  values derived from this PRE-CHANGE run) under `.moai/reports/t562/`. M2 changes the doctor stat
  target; after M2 the pre-change behaviour no longer exists to measure, so this baseline cannot be
  captured later than this window.

**Exit:** AC-CSRB-001, AC-CSRB-002's RED-now cell, **AC-CSRB-006 baseline captured**.

### M2 — Implement the conversion (Priority: High, the production diff)

- `internal/cli/codex_skills_prune.go`, `judgeCodexSkillEntry`, `codexPathAbsolute` case:
  `statPath = fromConfigPath(e.Path, configPathSeparator)` (was `statPath = e.Path`).
- `internal/cli/doctor_codex.go`, `codexStaleSkillFinding`, `codexPathAbsolute` case: the same
  one-line shape. NOTHING else in that function or file changes; `inspectSkillMirror` untouched.
- The `codexPathHomeRelative` case in BOTH readers: untouched (`spec.md` REQ-CSRB-002).
- Observe GREEN on all M1 tests.

**Exit:** AC-CSRB-002 GREEN, AC-CSRB-003/004/005 still GREEN.

### M3 — Doctor guard, mutants, pins, verification (Priority: Medium, mechanical)

- **AC-CSRB-006 (doctor regression guard — RE-RUN against the M1 baseline)**: the guard test was
  authored and its passing execution recorded as the baseline during M1, before any production
  change. Here: re-run the guard post-M2 and diff against that baseline — the existing-file entry
  stays not-counted-missing and the relative/oddly-formed/indeterminate counters are unchanged. Any
  diff between the post-M2 run and the M1 baseline is a FAIL, never a new expectation.
- **Mutants (record caught/missed per mutant in `.moai/reports/t562/`)**:
  - bypass mutant: remove the prune conversion call → AC-CSRB-002 must FAIL;
  - blanket-wrap mutant: `osStatFn(fromConfigPath(statPath, configPathSeparator))` at the stat
    site instead of the branch conversion → AC-CSRB-004 must FAIL;
  - reorder mutant: convert before `classifyCodexSkillPath` → AC-CSRB-003 must FAIL;
  - seam mutant: add `osStatFn` to `doctor_codex.go` → AC-CSRB-007 must FAIL (run, record, revert).
- **AC-CSRB-007**: `/usr/bin/grep -c 'osStatFn' internal/cli/doctor_codex.go` → 0, WITH the positive
  control (same pattern against `internal/cli/codex_skills_prune.go` → ≥ 1).
- **AC-CSRB-008**: re-derive `CARD_BASE=$(git merge-base origin/develop HEAD)`; changed-file set ⊆
  the allowlist in §G; `internal/codexwiring/skills.go` probe empty with a non-zero control.
- **AC-CSRB-009**: `go build ./...` rc=0 AND `GOOS=windows GOARCH=amd64 go build ./...` rc=0.
- **AC-CSRB-010**: scoped package suite with the executed-test control (predicate in §G).

**Exit:** AC-CSRB-006 through AC-CSRB-010.

## §F. Technical Approach

TDD, per milestone: RED observed before the implementation lands; guards carry their discriminating
power as named mutants rather than RED-now.

Two properties make the guards non-vacuous:

- **AC-CSRB-003/004/005 are green-by-construction guards** — their failure modes are future
  regressions (a reorder, a blanket wrap, a gating change), and each carries a named mutant that was
  actually executed and actually failed the test (§1.1 observed-failure: the mutant IS the known
  failing input).
- **AC-CSRB-007/008 are absence-shaped** — each zero carries a positive control, and each carries a
  mutant (seam insertion; a token probe) demonstrating the check can fail.

## §G. Verification

```bash
# scoped package suite — 1800s floor measured by t540's lane (702.5s, 2026-09-07)
go test ./internal/cli/... -timeout 1800s -v > .moai/reports/t562/ac-010.log 2>&1; echo "rc=$?"
# executed-test control — trailing space, NOT a $ anchor (go appends " (0.06s)")
AFTER=$(/usr/bin/grep -c -- '--- PASS: ' .moai/reports/t562/ac-010.log)
BEFORE=$(/usr/bin/grep -c -- '--- PASS: ' .moai/reports/t562/ac-010-base.log)
# PASS iff AFTER >= BEFORE AND AFTER > 0; anything else is "not measurable".
# Expected delta: t540's absorbed tests + this card's new tests (account for both, not just ours).

# builds
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# seam-out pin (zero WITH a positive control)
/usr/bin/grep -c 'osStatFn' internal/cli/doctor_codex.go        # expect 0
/usr/bin/grep -c 'osStatFn' internal/cli/codex_skills_prune.go  # control: expect >= 1

# scope pin (re-derived at read time; never a pinned SHA as the left edge)
git fetch origin develop
CARD_BASE=$(git merge-base origin/develop HEAD)
git diff --name-only "$CARD_BASE"..HEAD
git diff --name-only "$CARD_BASE"..HEAD -- internal/codexwiring/skills.go   # probe: expect empty
```

**AC-CSRB-008 changed-file allowlist** (the only files permitted in the diff): `internal/cli/
codex_skills_prune.go`, `internal/cli/codex_skills_prune_test.go` (or the new test file's actual
name), `internal/cli/doctor_codex.go`, a new doctor test file if the guard lands in its own file,
plus the four absorbed-from-t540 files (`codex_config_path.go`, `codex_config_path_test.go`,
`codex_skills_disable.go`, `codex_skills_disable_path_test.go`) and `.moai/` artifacts. Anything
outside the allowlist is a FAIL, and the allowlist is re-checked against the actual diff at
verification time rather than assumed.

Full-suite judgment belongs to CI on the pushed head, not to this machine.

## §H. Risks

| Risk | Shape | Mitigation |
|---|---|---|
| Conversion wrapped at the stat site instead of the branch | Home-relative Join products routed through `fromConfigPath` — the forbidden double conversion | `spec.md §D.2` pins the branch site; AC-CSRB-004 + the blanket-wrap mutant enforce it on this host |
| Conversion applied before classification | Windows-absolute declarations reclassify; skip reasons change | AC-CSRB-003's SkipReason discriminator + the reorder mutant |
| A doctor seam added "for testability" | Production-code change beyond scope; t563's design decision pre-empted | AC-CSRB-007 zero-pin with positive control; seam mutant executed and reverted in M3 |
| Regression guard's expected values derived post-change | The guard asserts the NEW behaviour, not the unchanged one | M3 order: capture pre-change output BEFORE M2 lands (the M1 window), derive expected values from it |
| `configPathSeparator` override leaks across tests | A parallel test reads another test's separator | `t.Cleanup` restore; tests non-parallel (M5 discipline, extended to the new seam) |
| Absorb merge conflicts | t540 touched `codex_skills_disable.go` + new files; origin/develop moved since | Merge conflict is a blocker report to the lead, not a forced resolution |
| AC-CSRB-010 green on a suite that never ran | Vacuous pass | Executed-test control: `--- PASS:` count predicate, BASE captured pre-flight |
| Parser touched "just to make it cleaner" | t533's counting basis moves | AC-CSRB-008 probe with non-zero control |
| An experiment writes the real `~/.codex` | Hard to reverse | Isolated fixtures only (REQ-CSRB-008) |
| A test performs a real deletion | The card's boundary is crossed | No test fires the destructive branch against a filesystem; eligibility is asserted on the verdict object (REQ-CSRB-007) |

## §I. Anti-Patterns

- Writing `fromConfigPath` logic inline (a local `strings.ReplaceAll`) instead of calling the seam.
  The seam is the contract; a local copy is a re-creation of it.
- Wrapping the stat call instead of converting in the absolute branch (see §H row 1).
- Reporting AC-CSRB-003/004/006 as RED evidence. They are guards; their strength is recorded as
  caught mutants, and claiming RED for them is a false claim.
- Claiming any Windows runtime observation. Build-level + structural only.
- Adding the doctor stat seam to "finish the symmetry". It is t563's scope and its presence FAILS
  the DoD.
- Pinning `bce6d7e08` as an AC's diff-range left edge (the t543 discipline — re-derive
  `CARD_BASE` at read time; the dated anchor is context, not a range edge).
- Editing pre-existing test expectations to make AC-CSRB-010 green.

## §J. Cross-References

- `spec.md` §B.5 (conversion-site derivation), §D.2 (the decision + its mutant), §D.3 (test
  asymmetry), §D.5 (predecessor-AC mapping), §F (exclusions), §G (gaps)
- `acceptance.md` — the ten ACs with their Given-When-Then bodies and two-cell records
- t540 (`WT-codex-path-escape`, frozen) — seam owner; its `progress.md §E.3` open-M3 record
- t563 — doctor stat seam owner (explicitly excluded here)
- t533 (`SPEC-CODEX-GHOST-SKILLS-MEASURE-001`) — layer-2 unblock target (§B.4)
