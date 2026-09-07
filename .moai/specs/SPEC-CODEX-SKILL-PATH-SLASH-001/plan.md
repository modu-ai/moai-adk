# SPEC-CODEX-SKILL-PATH-SLASH-001 — Implementation Plan

Card t540 · worktree `.claude/worktrees/t540` · branch `WT-codex-path-escape` · base `9ce792637`
(dated anchor, not a range edge) · Tier M · Class C · cycle_type **tdd**.

Milestones are ordered by decision-reversibility: the decisions most likely to change come first, so
review attention lands where it is worth spending. M4 is mechanical and sits last deliberately.

## §A. Context

`spec.md §B` carries the measured facts and their addresses; they are not restated here. Two things
are worth repeating at the top of the plan:

1. **M0 is a gate, and nothing after it may start until it returns.** The whole direction rests on an
   unmeasured Codex behaviour.
2. **M1 (the separator seam) precedes every other code milestone.** `filepath.ToSlash`/`FromSlash`
   are the identity function on this host — measured (`spec.md` M6) — so without the seam M2 and M3
   have no RED to establish and cycle_type=tdd cannot be satisfied here.

## §B. Known Issues Entering the Card

- No Windows host exists in this worktree. M0 cannot be executed here.
- `filepath.ToSlash`/`FromSlash` are identity where `Separator == '/'`, so the naive formulation is
  untestable on this host. M1 is the answer (`spec.md §B.3.1`).
- `classifyCodexSkillPath` is not separator-injectable and this card does not make it so; its Windows
  branch stays inference I2 (`spec.md §G` gap 4).
- t502 **F4** (zero coverage on the `:245` guard) is inherited and closed by M2.
- t502 **F5** (skip returns rc=0) is inherited and only PARTIALLY dissolved — see `spec.md §D.3`.

## §C. Pre-flight

```bash
git rev-parse --show-toplevel        # must be the t540 worktree
git branch --show-current            # WT-codex-path-escape
git rev-parse --short HEAD
git fetch origin develop
git merge-base origin/develop HEAD   # the card's live left edge

# AC-CSPS-006 BASE count — taken ONCE here, before M1 lands, with the same command
go test ./internal/cli/... -timeout 1800s -v > .moai/reports/t540/ac-006-base.log 2>&1; echo "rc=$?"
/usr/bin/grep -c -- '--- PASS: ' .moai/reports/t540/ac-006-base.log
```

## §D. Constraints

Carried from `spec.md §E`; the ones that bind commands:

- **Test scope.** Touched packages only. `go test ./internal/cli/... -timeout 1800s` — the 1800s
  floor is not decorative; a lane measured 702.5s on this package on 2026-09-07. Never
  `go test ./...` locally.
- **Isolated `CODEX_HOME`.** Every experiment that writes a Codex config runs under a scratch
  `CODEX_HOME`. The real `~/.codex/config.toml` is never written.
- **Absence claims** use `/usr/bin/grep`, never the shell's `grep`.
- **Commit messages** all contain `t540`.
- **Evidence** under `.moai/reports/t540/`.

## §E. Milestones

### M0 — the gate: measure Codex's Windows slash resolution (Priority: CRITICAL, blocking)

Highest reversibility risk in the card: a negative result invalidates every milestone below it.

- Build the isolated-`CODEX_HOME` probe (the harness established by SPEC-CODEX-SKILLCONFIG-SHAPE-001).
- Run BOTH arms on a Windows host: slash-form `path` and native-backslash-form `path` (control).
- Re-stamp `codex --version` at measurement time.
- Write `.moai/reports/t540/ac-001-windows-slash.md` with the commands, verbatim output, and host OS.

**Exit:** AC-CSPS-001. On refutation (slash arm fails, control arm passes) the card STOPS and reports;
the direction reverts to option (b), which re-opens the t533 counting basis. On a both-arms-fail
result the verdict is "harness did not work", and M0 repeats — it is NOT a refutation.

### M1 — build the separator seam (Priority: CRITICAL, enables every RED below)

Landed first because without it M2 and M3 have no RED on this host, and cycle_type=tdd cannot be
satisfied. `filepath.ToSlash`/`FromSlash` are identity where `Separator == '/'` — measured on this
host, `go run .moai/reports/t540/lab/ts.go` (`spec.md` M6), not assumed.

- Add the two pure functions in `internal/cli`, following the package's existing seam convention
  (`osStatFn`, `codexUserHomeDir` — no new abstraction):

  ```go
  func toConfigPath(p string, sep rune) string   // sep == '/' → identity; else sep → '/'
  func fromConfigPath(p string, sep rune) string // sep == '/' → identity; else '/' → sep
  ```

- Add the injection point: `var configPathSeparator = filepath.Separator`. Production call sites pass
  it; tests override it with `t.Cleanup` restore and stay **non-parallel** (a package-level var is
  shared state — the same discipline `osStatFn` already requires).
- **[HARD] Separator-aware, never `strings.ReplaceAll(p, "\\", "/")`.** A unix filename may legally
  contain `\`, and such a path is today REFUSED by `:245`. An unconditional replacement converts that
  refusal into publication of a wrong path — strictly worse than the defect being repaired. When
  `sep == '/'`, `toConfigPath` returns its input untouched.

**RED (run in this order):**

```bash
# 1. the functions do not exist → the new test file does not compile. That IS the first RED.
go test ./internal/cli/... -run 'TestConfigPath' -timeout 1800s

# 2. implement as an identity stub for BOTH separators, re-run: the sep='\\' assertions fail.
#    This stub is the required mutation arm — record that it was rejected.
# 3. implement the separator-aware bodies → GREEN.
```

**Exit:** AC-CSPS-004 arm B.

### M2 — publisher conversion and comparison (Priority: High)

The design decision most likely to be argued with, so it is settled early.

- Convert before the `ContainsAny` guard at `internal/cli/codex_skills_disable.go:245`:
  `skillPath = toConfigPath(skillPath, configPathSeparator)`.
- The guard's character set keeps `"`, LF, CR. On a non-`/` host the separator no longer reaches it;
  on a `/` host a literal `\` still does, and is still refused (REQ-CSPS-010).
- Compare on the converted form at `:252`: `toConfigPath(e.Path, configPathSeparator) == skillPath`.
  Per `spec.md` M4 this is the verb's ONLY comparison — `:257` counts the resulting slice and
  performs no comparison of its own, so there is no second site to keep in step.
- The update branch continues to rewrite ONLY `enabled`; the stored `path` line is left in whatever
  form it already had. Rewriting a declaration this verb did not author is out of scope.
- No escape sequence is ever emitted. The comment block above the guard is rewritten to describe what
  the guard now protects — it currently asserts a reach the code will no longer have, and a comment
  asserting a property the code lacks is worse than none.

**RED:**

```bash
# with configPathSeparator overridden to '\\' in the test, the backslash literal is still
# refused before the change → Action == codexSkillDisableSkipped. Real RED on this host.
go test ./internal/cli/... -run 'TestUpsertCodexSkillDisable' -timeout 1800s
```

**Exit:** AC-CSPS-002, AC-CSPS-003 (both arms), AC-CSPS-007 (both arms). Closes t502 F4.

### M3 — reader-side conversion (Priority: High)

- `internal/cli/codex_skills_prune.go` (`judgeCodexSkillEntry`, consumption at `:71-94`, stat via the
  `osStatFn` seam at `:96`) and `internal/cli/doctor_codex.go` (`codexStaleSkillFinding`, `:824-857`
  — classification `:831-856`, stat `:857`) apply `fromConfigPath(statPath, configPathSeparator)` to
  the resolved stat target.
- **[HARD] The two readers are NOT symmetric for testing, and this milestone must not pretend they
  are.** `doctor_codex.go:857` calls `os.Stat` **directly**, not `osStatFn` — measured:
  `/usr/bin/grep -n 'os\.Stat\|osStatFn' internal/cli/doctor_codex.go` → `459: os.Stat`,
  `857: os.Stat`, zero `osStatFn`. `codexStaleSkillFinding` also has zero existing tests
  (`/usr/bin/grep -rn 'codexStaleSkillFinding' --include='*_test.go' internal/cli/` → rc=1). So the
  recorder-based RED exists on the **prune** side only; the doctor side is covered by an
  output-level regression guard. **Do NOT add a stat seam at `:857` to make the two symmetric** —
  that is production-code change, out of scope (`spec.md §F`), and doing it quietly is exactly the
  failure this note prevents.
- **Ordering constraint (`spec.md §D.2`):** classification runs on the DECLARED form.
  `classifyCodexSkillPath` (`doctor_codex.go:667`) runs `filepath.IsAbs` first and treats any `\` as
  oddly-formed; the conversion is applied to the stat target AFTER classification, never before it.
  Inverting this reclassifies legitimate Windows-absolute declarations.
- The home-relative branch (`expandCodexHomeRelativePath`, `:686`) is a `filepath.Join` product and
  already carries native separators — confirm it is not double-converted.
- `classifyCodexSkillPath` itself is NOT made separator-injectable (`spec.md §F`). Its Windows
  behaviour stays inference I2 and is not claimed as measured.

**RED — at command level (this is what F-2 asked for):**

```bash
# PRUNE arm (AC-CSPS-004 arm A) — the ONLY recorder-based RED available in this card.
# Fixture: configPathSeparator = '\\'; osStatFn replaced by a recorder capturing its argument;
# entry declares a HOST-ABSOLUTE slash path (/tmp/x/SKILL.md) so classifyCodexSkillPath returns
# codexPathAbsolute on this host. A "C:/..." fixture does NOT work here — IsAbs is false (M6),
# the entry classifies codexPathRelative, and stat is never reached.
go test ./internal/cli/... -run 'TestJudgeCodexSkillEntry_SeparatorConversion' -timeout 1800s

# BEFORE the change the recorder observes "/tmp/x/SKILL.md"  → assertion fails (RED)
# AFTER  the change the recorder observes "\\tmp\\x\\SKILL.md" → GREEN

# DOCTOR arm (AC-CSPS-004 arm A') — output-level REGRESSION GUARD, claims NO RED.
# doctor_codex.go:857 calls os.Stat directly, so no recorder can observe its argument.
# This asserts that a slash-declared entry naming an existing file is not counted missing.
go test ./internal/cli/... -run 'TestCodexStaleSkillFinding_SlashDeclaredEntry' -timeout 1800s
```

**Do not report the doctor arm as RED evidence.** It is green before and after on a `/`-separator
host. Arm A' of AC-CSPS-004, arm C, and all of AC-CSPS-005 are regression guards; the TDD obligation
for M3 is discharged by the prune recorder arm above plus AC-CSPS-004 arm B.

**Exit:** AC-CSPS-004 (arms A and C), AC-CSPS-005.

### M4 — scope pinning and regression (Priority: Medium, mechanical)

- Confirm `internal/codexwiring/skills.go` untouched via the re-derived-left-edge recipe
  (AC-CSPS-008). Control first; a zero control is "not measurable".
- Run the full `internal/cli` package suite with the **executed-test control** (AC-CSPS-006): cite the
  `--- PASS: ` count, not only the pass verdict. A count of 0, or one materially below the base
  count, is reported as "not measurable" — a zero-match selector and a package that failed to
  compile both present as a quiet green.
- Confirm no pre-existing expectation was edited.

**Exit:** AC-CSPS-006, AC-CSPS-008.

## §F. Technical Approach

TDD, per milestone: write the failing test, observe RED, implement, observe GREEN.

Two AC shapes are vacuously satisfiable and get a mutation arm rather than a bare green:

- **AC-CSPS-003** (refusal still fires) and **AC-CSPS-008** (a file is absent from a diff) are
  absence-shaped. A broken harness satisfies both. Establish RED by mutation — delete the guard,
  touch the parser — and record the mutant that was NOT caught as well as the one that was.
- **AC-CSPS-004** needs its negative arm (a slash path naming a nonexistent file still yields
  `Eligible: true`), or a mutation making everything "resolve" passes.
- **The identity-stub mutant is mandatory** for both `toConfigPath` and `fromConfigPath`: a stub
  returning its input for every separator must FAIL AC-CSPS-004 arm B and AC-CSPS-002. Without it
  the seam could be wired to a no-op and every AC would still read green on this host.
- **The unconditional-replacement mutant is mandatory**: replacing `toConfigPath` with
  `strings.ReplaceAll(p, "\\", "/")` must FAIL AC-CSPS-003 arm 2.
- **AC-CSPS-006 is vacuous without its executed-test control** — see M4.

## §G. Verification

```bash
# scoped package suite — the 1800s floor is measured, not guessed
go test ./internal/cli/... -timeout 1800s -v > .moai/reports/t540/ac-006.log 2>&1; echo "rc=$?"
# executed-test control — trailing space, NOT a $ anchor (go appends " (0.06s)")
AFTER=$(/usr/bin/grep -c -- '--- PASS: ' .moai/reports/t540/ac-006.log)
BEFORE=$(/usr/bin/grep -c -- '--- PASS: ' .moai/reports/t540/ac-006-base.log)
# PASS iff AFTER >= BEFORE AND AFTER > 0; anything else is "not measurable"
echo "before=$BEFORE after=$AFTER"

# the separator probe behind M6 (re-run to re-stamp, never cite from memory)
go run .moai/reports/t540/lab/ts.go

# absence claims
/usr/bin/grep -rn 'cannot carry verbatim' --include='*.go' internal/

# scope pin (control first)
git fetch origin develop
CARD_BASE=$(git merge-base origin/develop HEAD)
git diff --name-only "$CARD_BASE"..HEAD | wc -l
git diff --name-only "$CARD_BASE"..HEAD -- 'internal/codexwiring/skills.go'
```

Full-suite judgment belongs to CI on the pushed head, not to this machine.

## §H. Risks

| Risk | Shape | Mitigation |
|---|---|---|
| M0 refutes the direction | The whole card reverts to option (b) | M0 is the gate; nothing lands before it |
| M0 harness fails on both arms | Reads as refutation if the control is omitted | Control arm is mandatory (AC-CSPS-001) |
| Seam skipped, `filepath.ToSlash` used directly | Identity on this host (M6) → AC-CSPS-002/007 unsatisfiable, "platform-independent" claim false | M1 lands the seam first; identity-stub mutant is mandatory |
| Seam implemented as unconditional `ReplaceAll` | A unix filename containing `\` gets published as a wrong path — worse than today's refusal | REQ-CSPS-010, AC-CSPS-003 arm 2, dedicated mutant |
| Conversion applied before classification | Windows-absolute declarations reclassify as oddly-formed | §D.2 ordering constraint, pinned in M3 |
| Comparison normalization omitted | Duplicate entries; manufactures t506's debt | M2 pins `:252`; AC-CSPS-007 both arms |
| `configPathSeparator` override leaks across tests | A parallel test reads another's separator | Override with `t.Cleanup`, tests non-parallel (the `osStatFn` discipline) |
| AC-CSPS-006 green on a suite that never ran | Vacuous pass | Executed-test control in M4 |
| Parser touched "just to make it cleaner" | t533's counting basis moves | AC-CSPS-008, mutation-verified |
| An experiment writes the real `~/.codex` | Hard to reverse | Isolated `CODEX_HOME`, REQ-CSPS-009 |

## §I. Anti-Patterns

- Treating AC-CSPS-001 as passed because forward slashes "generally work on Windows". That is the
  premise under test.
- Deleting the `:245` guard instead of narrowing what reaches it. The guard protects a real
  divergence (`spec.md §B.1`).
- Emitting `path = "C:\\Users\\..."`. Any escape is a defect, not a workaround.
- Editing a pre-existing test expectation to make AC-CSPS-006 green.
- Reading a green AC-CSPS-006 without its `--- PASS: ` count as evidence the suite ran.
- Claiming AC-CSPS-002 is a "platform-independent" measurement without the M1 seam in place — on this
  host that claim is false (M6).
- Refactoring `classifyCodexSkillPath` to chase the last unmeasurable branch. It is out of scope, and
  the gap is recorded (`spec.md §G` gap 4) rather than closed.
- Pinning `9ce792637` as an AC's diff-range left edge (card t543 discipline).

## §J. Cross-References

- `spec.md` §B.1 (measured vs inferred), §B.3.1 (the separator seam), §D.2 (ordering), §D.3 (F4/F5), §D.4 (reader site 3 + the retracted clause), §G (gaps 1-5)
- `acceptance.md` — the eight ACs
- `.moai/reports/t540/reader-census.md` — the reader census and its 2026-09-08 corrections
- `.moai/reports/t540/lab/ts.go` — the separator probe producing M6
- `.moai/reports/t540/plan-audit.md` — plan-audit iteration 1/2 (FAIL)
- t533 — its prune-execution ACs wait on this card; t506 — the duplicate-collapse surface
