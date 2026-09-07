---
id: SPEC-CODEX-MIRROR-DOCTOR-001
title: "Acceptance criteria — .agents/skills mirror-state diagnostic"
version: "0.2.1"
created: 2026-09-07
updated: 2026-09-07
---

# Acceptance Criteria — SPEC-CODEX-MIRROR-DOCTOR-001

Every criterion names the command that decides it and the REQ it judges. All Go tests live in
`internal/cli/doctor_codex_test.go` and pin PATH and home through the existing `codexWiringLookPath`
and `codexUserHomeDir` seams — no test may read the developer's real `$HOME`, real
`~/.codex/config.toml`, or real PATH (REQ-CMD-011, judged by AC-CMD-015).

## The swept-count gate (applies to every `Decided by` below)

Every test name this file names is **new** — none exists on the pre-implementation tree. `go test
-run <absent-name>` exits 0 and prints `ok ... [no tests to run]`, which is indistinguishable from a
pass. So no `Decided by` here is satisfied by an exit code alone:

> **Swept-count gate.** The named command runs with `-v`, and its output must contain
> `--- PASS: <name>` **exactly once per named test**. Output containing `[no tests to run]`, or
> containing zero `--- PASS:` lines for a named test, is **RED** — never green — whatever the exit
> code says (`.claude/rules/moai/development/verification-completeness.md` §1.1, "A pass whose swept
> set is empty asserts nothing").

Measured on the pre-implementation tree (tree SHA `dcb3ba0c7`, worktree `.claude/worktrees/t498`,
branch `WT-codex-mirror-doctor`), demonstrating the gate is not decorative:

```
$ go test ./internal/cli/ -run TestCheckCodexWiring_ClaudeOnlyMachineNoMirrorRow -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.690s [no tests to run]
EXIT=0
$ grep -c -- '--- PASS: TestCheckCodexWiring_ClaudeOnlyMachineNoMirrorRow'
0
```

Exit 0 with a swept count of 0: the exit code alone would have read this as a pass.

## AC-CMD-001 — the un-nagging invariant (the regression this change most plausibly causes)

REQ: REQ-CMD-003 (and the §6 no-new-row constraint on the claude-only path).

**Given** a project root from `t.TempDir()` carrying no `.codex/` wiring files and no
`.agents/skills` directory, and a stubbed PATH reporting `codex` NOT found,
**When** `checkCodexWiring(root, true)` is called (verbose on, so Detail is populated),
**Then** `check.Status == uikit.CheckOK`, the Message is the existing informational-skip text, and
neither `check.Message` nor `check.Detail` contains `.agents` or `mirror` (case-insensitive).

Decided by: `go test ./internal/cli/ -run TestCheckCodexWiring_ClaudeOnlyMachineNoMirrorRow -count=1 -v`
— swept-count gate applies: exactly one `--- PASS: TestCheckCodexWiring_ClaudeOnlyMachineNoMirrorRow`.

**Relationship to the existing test.** This AC **supplements**, and does not replace,
`TestCheckCodexWiring_ClaudeOnlyMachineStaysSilent` (`internal/cli/doctor_codex_test.go:547`). That
test guards the same invariant against a different intruder — the home-layer `skills.config`
sub-check — and stays in the file unmodified. The new test guards it against the mirror
sub-check. Deleting or rewriting the existing test is out of scope.

### RED-now cell (§2 two-cell adoption discipline)

- **Command**: `go test ./internal/cli/ -run TestCheckCodexWiring_ClaudeOnlyMachineNoMirrorRow -count=1 -v`
- **Verbatim output**: `testing: warning: no tests to run` / `PASS` /
  `ok  github.com/modu-ai/moai-adk/internal/cli 0.690s [no tests to run]`
- **Exit code**: `0`
- **Tree SHA**: `dcb3ba0c7` (worktree `.claude/worktrees/t498`, branch `WT-codex-mirror-doctor`)
- **Verdict under this AC**: **RED** — swept count is 0 (`grep -c -- '--- PASS: <name>'` → `0`).
- **Why it is red — the stated reason**: the judging test does not exist on this tree. This is the
  *vacuous* direction, not the *impossible* one: the assertion targets `check.Message` /
  `check.Detail` of a function that already exists and already returns `CheckOK` on this input
  (`TestCheckCodexWiring_ClaudeOnlyMachineStaysSilent` passes today), so no pre-existing unrelated
  file stands between this work and green.

### Green path cell

- **Milestone that flips it**: M4 (Tests), which authors the test; M2 (wiring) is what keeps its
  body true — the inspector is called after the `!wired && !codexInstalled` early return.
- **Passing output becomes**: `--- PASS: TestCheckCodexWiring_ClaudeOnlyMachineNoMirrorRow (0.00s)`
  present exactly once, with `[no tests to run]` absent.

## AC-CMD-002 — claude-only render is byte-unchanged

REQ: REQ-CMD-003 (§6 no-new-row constraint, render side).

**Given** the committed doctor golden fixture,
**When** the doctor panel is rendered on a project with no Codex wiring,
**Then** `internal/cli/testdata/doctor-nocolor.golden` requires no update — the diff is empty.

Decided by: `go test ./internal/cli/ -run 'TestDoctor.*Golden|TestDoctor_' -count=1 -v` followed by
`git diff --name-only ace1c5440 -- internal/cli/testdata/` printing nothing.

This selector matches tests that already exist (`TestDoctorGolden_{Light,Dark,NoColor}`,
`TestDoctor_CheckCount`), so the swept-count gate is satisfied by its ordinary output; the second
clause is pinned to the base SHA `ace1c5440` rather than the moving working tree.

## AC-CMD-003 — absent mirror on a wired project is a finding carrying the directive

REQ: REQ-CMD-004.

**Given** a wired project root (`wireProjectForDoctor`) with `moai` and `codex` stubbed found, and
no `.agents/skills` directory,
**When** `checkCodexWiring(root, false)` is called,
**Then** `check.Status == uikit.CheckWarn` and `check.Message` contains both `.agents/skills` and
the exact substring `moai update --templates-only --force --yes`.

Decided by: `go test ./internal/cli/ -run TestCheckCodexWiring_MirrorAbsentAdvisesRedeploy -count=1 -v`
— swept-count gate applies: exactly one `--- PASS: TestCheckCodexWiring_MirrorAbsentAdvisesRedeploy`.

## AC-CMD-004 — dangling mirror entries are a finding naming the count

REQ: REQ-CMD-005.

**Given** a wired project root whose `.agents/skills` holds three entries, two of them relative
symlinks to `../../.claude/skills/<name>` targets that do not exist,
**When** `checkCodexWiring(root, false)` is called,
**Then** `check.Status == uikit.CheckWarn` and `check.Message` contains `2` together with the
re-deploy directive substring.

Decided by: `go test ./internal/cli/ -run TestCheckCodexWiring_DanglingMirrorEntriesCounted -count=1 -v`
— swept-count gate applies: exactly one `--- PASS: TestCheckCodexWiring_DanglingMirrorEntriesCounted`.

## AC-CMD-005 — copy-mode entries are Detail-only, never a Message finding

REQ: REQ-CMD-006.

**Given** a wired project root whose `.agents/skills` holds one valid symlink and one real
directory, with `.claude/skills` carrying both canonical skills,
**When** `checkCodexWiring(root, true)` is called,
**Then** the copy-mode count appears in `check.Detail`, `check.Message` contains no copy-mode text,
and — with no other problem present — `check.Status == uikit.CheckOK`.

Decided by: `go test ./internal/cli/ -run TestCheckCodexWiring_CopyModeDetailOnly -count=1 -v`
— swept-count gate applies: exactly one `--- PASS: TestCheckCodexWiring_CopyModeDetailOnly`.

## AC-CMD-006 — unmirrored skills are Detail-only, never a Message finding

REQ: REQ-CMD-007.

**Given** a wired project root whose `.claude/skills` holds three skills and whose `.agents/skills`
mirrors only one of them, all mirror entries valid,
**When** `checkCodexWiring(root, true)` is called,
**Then** the unmirrored count `2` appears in `check.Detail`, `check.Message` contains no unmirrored
text, and `check.Status == uikit.CheckOK`.

Decided by: `go test ./internal/cli/ -run TestCheckCodexWiring_UnmirroredSkillsDetailOnly -count=1 -v`
— swept-count gate applies: exactly one `--- PASS: TestCheckCodexWiring_UnmirroredSkillsDetailOnly`.

## AC-CMD-007 — an unwired project with codex installed gains no mirror output at all

REQ: REQ-CMD-008, and the unwired clause of REQ-CMD-006/007.

**Given** a `t.TempDir()` project root with no `.codex/` files and a stubbed PATH reporting `codex`
FOUND, exercised in **two** sub-cases: (a) no `.agents/skills` directory, and (b) an `.agents/skills`
directory holding one valid symlink, one real directory, and one dangling symlink,
**When** `checkCodexWiring(root, true)` is called in each sub-case,
**Then** in both: `check.Message` contains the existing `run moai init --agent codex` directive and
contains no `.agents/skills` text, and `check.Detail` carries no mirror line — no finding, and no
copy-mode / unmirrored / indeterminate count. Sub-case (b) is what pins REQ-CMD-006/007's unwired
clause: a mirror in a *reportable* state still produces nothing, because the inspector is never
called outside the `wired` branch.

Decided by: `go test ./internal/cli/ -run TestCheckCodexWiring_UnwiredNoMirrorNag -count=1 -v`
— swept-count gate applies: exactly one `--- PASS: TestCheckCodexWiring_UnwiredNoMirrorNag`, plus
one `--- PASS:` line per sub-test if the test is written table-driven.

## AC-CMD-008 — an unreadable mirror is indeterminate, never a finding

REQ: REQ-CMD-010.

**Given** a wired project root where `.agents/skills` is made unreadable using this file's existing
indeterminate-fixture idiom — a **symlink loop**, as in
`TestCheckCodexWiring_IndeterminateStatNotMissing` (`internal/cli/doctor_codex_test.go:335-352`):
`.agents/skills` is a symlink to a sibling path which is itself a symlink back to `.agents/skills`,
**When** `checkCodexWiring(root, true)` is called,
**Then** `check.Detail` records the condition as not checked, `check.Message` carries no mirror
finding, and the mirror contributes no `CheckWarn`.

Fixture obligations, carried verbatim from the neighbouring function so the test is portable and
leaves no cleanup hazard:

- Skip on Windows (`if runtime.GOOS == "windows" { t.Skip("symlink creation needs privileges on windows") }`).
- `t.Skipf` if either `os.Symlink` call fails — the host does not support the fixture.
- Control assertion before the subject call: `os.ReadDir` on the loop path must return an error that
  is **non-nil and not `fs.ErrNotExist`**; otherwise `t.Skipf`. Without it the AC could pass on a
  host where the loop resolved, asserting nothing.

**No `chmod 0o000` fixture.** It was measured to break the test it is written into: a `t.TempDir()`
mode-0 directory holding entries fails `TempDir RemoveAll cleanup` with `permission denied`, which
Go reports as a test failure (plan-audit.md D1). If a future author reintroduces a chmod fixture,
the AC must state whether the directory holds entries **and** must restore the mode via
`t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })` — the symlink loop needs neither.

Decided by: `go test ./internal/cli/ -run TestCheckCodexWiring_MirrorUnreadableIndeterminate -count=1 -v`
— swept-count gate applies: exactly one `--- PASS: TestCheckCodexWiring_MirrorUnreadableIndeterminate`.
A `--- SKIP:` line is **not** a pass: on a host that skips, this AC is unmet and must be re-run on a
symlink-capable host (linux/darwin), which is where the run-phase evidence is taken.

## AC-CMD-009 — every mirror summary stays inside the width bound standalone

REQ: REQ-CMD-009, clause (a).

**Given** each mirror summary the implementation can emit,
**When** its rune length is measured with `utf8.RuneCountInString`,
**Then** each is `<= codexMessageWidthCeiling` (113) standalone, and the rendered panel width test
already in the file still passes.

Decided by:
`go test ./internal/cli/ -run 'TestCheckCodexWiring_MirrorSummaryWidth|TestCheckCodexWiring_RenderedPanelStaysInBand' -count=1 -v`
— swept-count gate applies to **both** names: exactly one `--- PASS:` line each. The second test
already exists (`doctor_codex_test.go:437`); the first is new.

## AC-CMD-010 — the check writes nothing

REQ: REQ-CMD-002.

**Given** a wired project root whose `.agents/skills` and `.claude/skills` trees are recorded as a
sorted `path + mode + symlink-target` listing before the call,
**When** `checkCodexWiring(root, true)` is called,
**Then** the same listing taken after the call is byte-identical.

Decided by: `go test ./internal/cli/ -run TestCheckCodexWiring_MirrorCheckIsReadOnly -count=1 -v`
— swept-count gate applies: exactly one `--- PASS: TestCheckCodexWiring_MirrorCheckIsReadOnly`. This
gate matters more here than anywhere else: this AC is the *only* mechanical guard on the read-only
boundary, so a vacuous green here would silently remove the boundary's enforcement.

## AC-CMD-011 — package gates stay green

REQ: — (§6 Constraints; no single REQ).

**Given** the completed change,
**When** the package-scoped gates run,
**Then** all three exit 0: `go test ./internal/cli/... -count=1`,
`go vet ./internal/cli/...`, and `golangci-lint run ./internal/cli/...`.

Baseline attribution: `go vet` and `golangci-lint` were both measured green on tree `26b77973e`
before this change (plan-audit.md, "베이스라인 게이트"). `go test ./internal/cli/... -count=1` was
**not** measured at plan time (deliberately, per CLAUDE.local.md §4 load discipline) — the run phase
takes its own baseline before the first edit, so a pre-existing red is attributable.

## AC-CMD-012 — cross-platform build

REQ: — (§6 Constraints, cross-platform clause).

**Given** the completed change,
**When** `GOOS=windows GOARCH=amd64 go build ./...` runs,
**Then** it exits 0. (Note: a cross-build does not compile tests — the Windows copy-mode path stays
unexercised, consistent with root-cause.md Gaps and spec.md §6.)

## AC-CMD-013 — a mirror finding riding alongside an existing finding stays inside the bound

REQ: REQ-CMD-009, clause (b) — participation in `joinCodexSummaries` tail-drop, unchanged.

**Given** a wired project root arranged so that the check produces **both** an existing non-mirror
finding (a stale home-config skill entry, built with `writeCodexHomeConfig` +
`absentSkillPath` as in the existing stale-skill tests) **and** a mirror finding (absent
`.agents/skills`), with the two summaries chosen so their joined length exceeds
`codexMessageWidthCeiling` (113),
**When** `checkCodexWiring(root, false)` is called,
**Then** all three hold:
1. `utf8.RuneCountInString(check.Message) <= codexMessageWidthCeiling`;
2. `check.Message` ends with the overflow marker `" (+N more, see --verbose)"` naming the exact
   number of dropped summaries;
3. `check.Detail` still carries the **full** text of every finding including the dropped one — the
   directive survives in Detail even when Message drops it.

**And** a second sub-case pinning the deliberate lead-summary exception in `joinCodexSummaries`
(`internal/cli/doctor_codex.go:228-248`): where the mirror finding is the **only** finding and its
summary alone exceeds the ceiling, it is emitted whole rather than truncated, and no overflow marker
appears. (AC-CMD-009 asserts every mirror summary is under the ceiling standalone, so this branch
should be unreachable for mirror summaries alone — this sub-case exists to make that
unreachability *observed* rather than assumed.)

Decided by: `go test ./internal/cli/ -run TestCheckCodexWiring_MirrorFindingParticipatesInTailDrop -count=1 -v`
— swept-count gate applies: exactly one `--- PASS: TestCheckCodexWiring_MirrorFindingParticipatesInTailDrop`.

## AC-CMD-014 — the observation reports through the existing row in the existing two-register shape

REQ: REQ-CMD-001.

**Given** a wired project root with an absent `.agents/skills` mirror,
**When** `checkCodexWiring(root, true)` is called,
**Then** all three hold:
1. `check.Name == "Codex Wiring"` — the mirror observation rides the existing row rather than
   renaming or splitting it. (The single-return shape is fixed by the signature
   `func checkCodexWiring(root string, verbose bool) DiagnosticCheck`, `doctor_codex.go:86`, and is
   not asserted here — an assertion the compiler already guarantees would be decorative.)
2. the mirror observation appears in **both** registers per the file's `codexFinding` convention:
   its `summary` in `check.Message` and its fuller `detail` text in `check.Detail`;
3. rendering that check through `renderDoctorGroups(w, groups, false, th)` (`doctor_render.go:114`)
   produces output carrying no mirror detail text — Detail is a `--verbose`-only *display*
   register, gated at the render layer (`doctor_render.go:134-137`). The `check.Detail` **field**
   itself is populated regardless of `verbose` (`doctor_codex.go:196, 200`), so the assertion is on
   the rendered output, never on the field.

Plus the registration-layer half, decided without a new test: `go test ./internal/cli/ -run
TestDoctor_CheckCount -count=1 -v` passes unchanged (exactly one `--- PASS: TestDoctor_CheckCount`),
confirming no new top-level row was added.

Decided by: `go test ./internal/cli/ -run 'TestCheckCodexWiring_MirrorUsesExistingRowTwoRegisters|TestDoctor_CheckCount' -count=1 -v`
— swept-count gate applies to both names.

## AC-CMD-015 — the new tests are hermetic

REQ: REQ-CMD-011.

**Given** the test file after this change,
**When** the seam usage is swept,
**Then** both hold:
1. every new test that needs PATH or home goes through `stubCodexLookup` / `stubMoaiLookup` /
   `stubCodexHome` / `wireProjectForDoctor`, and every project root is a `t.TempDir()`;
2. the swept file contains **no** direct `os.UserHomeDir(`, `exec.LookPath(`, or `os.Getenv("HOME")`
   call — decided by
   `grep -nE 'os\.UserHomeDir\(|exec\.LookPath\(|os\.Getenv\("HOME"\)' internal/cli/doctor_codex_test.go`
   printing nothing.

The grep's non-vacuity needs a control, because a broken pattern also prints nothing. The control is
**`internal/cli/doctor.go`**, measured on tree `dcb3ba0c7`:

```
$ grep -cE 'os\.UserHomeDir\(|exec\.LookPath\(|os\.Getenv\("HOME"\)' internal/cli/doctor.go
4
```

Four hits — so the pattern does match real call sites when they exist. `internal/cli/doctor_codex.go`
is deliberately **not** the control: it already prints zero on this tree, because its own PATH access
goes through the `codexWiringLookPath = exec.LookPath` seam assignment (an assignment, not a call),
which is exactly the hermetic shape this AC requires of the tests. A run where the doctor.go control
prints `0` means the pattern broke; the AC is then unmet, not passed.

## Definition of Done

- AC-CMD-001 through AC-CMD-015 all pass, each with the named command's output cited **and** its
  swept count shown — an exit code alone satisfies none of them.
- No file under `.agents/` is created, modified, or committed by this change.
- `internal/cli/update.go` and `internal/template/skill_mirror.go` are unmodified.
  `git diff --name-only ace1c5440` shows only `internal/cli/doctor_codex.go`,
  `internal/cli/doctor_codex_test.go`, and files under
  `.moai/specs/SPEC-CODEX-MIRROR-DOCTOR-001/` (which includes the run-phase `progress.md` §E.2/§E.3
  updates — those are expected, not a scope breach).
