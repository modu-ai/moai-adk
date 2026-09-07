# SPEC-CODEX-SKILL-PATH-SLASH-001 — Acceptance Criteria

Card t540. Every criterion below is binary-testable. Given-When-Then is the verification-layer
format; the GEARS requirement wording lives in `spec.md §C`.

## §D. AC Matrix

| AC | Covers | Kind | Blocking |
|---|---|---|---|
| AC-CSPS-001 | maps REQ-CSPS-001, REQ-CSPS-009 | measurement gate | YES — nothing lands before it |
| AC-CSPS-002 | maps REQ-CSPS-002 (+ closes t502 F4) | unit, separator-injected | yes |
| AC-CSPS-003 | maps REQ-CSPS-003, REQ-CSPS-010 | unit | yes |
| AC-CSPS-004 | maps REQ-CSPS-004, REQ-CSPS-005 | unit | yes |
| AC-CSPS-005 | maps REQ-CSPS-004 | integration (round-trip) | yes |
| AC-CSPS-006 | maps REQ-CSPS-006 | regression | yes |
| AC-CSPS-007 | maps REQ-CSPS-008 | unit | yes |
| AC-CSPS-008 | maps REQ-CSPS-007 | range-limited scope check | yes |

---

## AC-CSPS-001 — Codex on Windows resolves a forward-slash `path` (GATE)

**Given** a Windows host (physical, VM, or a Windows CI runner) with codex-cli installed, an
isolated `CODEX_HOME` pointing at a scratch directory, and a probe `SKILL.md` placed **under a skill
root** at `%CODEX_HOME%\skills\<probe>\SKILL.md` — placement under a root is what introduces a
skill; a `[[skills.config]]` entry does **not** introduce one, it only configures one already
introduced,
**When** the same fixture is observed across three arms of `codex debug prompt-input "hi"` — a
baseline arm with **no** `[[skills.config]]` entry at all, a slash arm declaring one entry whose
`path` is the **forward-slash** form of that SKILL.md with `enabled = false`, and a control arm
declaring the same entry in **native backslash** form with `enabled = false`,
**Then** the probe skill is observed **present** in the baseline arm and **disappears** in the slash
arm — the disappearance establishing that Codex matched the declared forward-slash path against the
on-disk native path, i.e. that it resolves the forward-slash form. Persistence of the probe skill in
the slash arm establishes that it did not.

- **Baseline arm is mandatory, not an aside.** The baseline arm MUST be run in the same run as the
  other two, and MUST observe the probe skill present. "Disappearance" is only an event if presence
  was established first; a run whose baseline arm does not show the skill present is "the harness
  did not work", not a result of any kind.
- **Slash arm** — the entry's `path` carries the forward-slash form (e.g.
  `C:/scratch/skills/probe/SKILL.md`) while the file on disk is at `C:\scratch\skills\probe\SKILL.md`.
  Disappearance = path matched = slash form resolved. Persistence = not resolved.
- **Control arm** — the identical entry in native backslash form, same observation. It is the
  attribution control: without it, a disappearance in the slash arm could not be attributed to path
  matching rather than to run-to-run variance in the render.
- **"Harness did not work" is distinguished from a refutation, in two shapes.** (i) A run where the
  baseline arm does not show the probe present is a broken harness. (ii) A run where the baseline
  arm shows the probe present but **both** the slash arm and the control arm leave it present is
  also a broken harness — the disabling mechanism itself did not fire — not a refutation of the
  slash direction.
- The codex-cli version is re-stamped at measurement time (`codex --version`), never carried over.
- Evidence: the command, its verbatim output, and the host's OS/version, written to
  `.moai/reports/t540/ac-001-windows-slash.md`.

**Fallback clause.** If the probe skill REMAINS PRESENT in the slash arm while it DISAPPEARS in the
control arm, the slash direction is refuted. The card then STOPS, reports the refutation, and the
direction reverts to option (b) — reader-side escape decoding — which re-opens the t533
counting-basis question (`spec.md §B.3` property 2). No publisher change lands on a refuted or
unmeasured gate.

**Gap acknowledged.** No Windows host is available in this worktree. This AC is not satisfiable here
and must be executed elsewhere; recording it as "assumed to pass" is prohibited.

### Amendment record — polarity inverted from appearance to disappearance

**What was measured.** Three harness-shape probe cells were run on this host (codex-cli `0.153.4`,
darwin, isolated `CODEX_HOME`), with an identical fixture
(`$CODEX_HOME/skills/t540probe/SKILL.md`) and an identical command
(`CODEX_HOME=<scratch> timeout 90 codex debug prompt-input "hi" < /dev/null`). Raw figures, cited to
`.moai/reports/t540/probe-harness-shape.log` (full report:
`.moai/reports/t540/m0-harness-probes.md`):

| cell | `[[skills.config]]` entry | rc | stderr bytes | `grep -c 't540probe'` |
|---|---|---|---|---|
| p1 | none at all | 0 | 0 | **1** |
| p2 | entry, `enabled = false` | 0 | 0 | **0** |
| p2b | entry, `enabled = true` (attribution control) | 0 | 0 | **1** |

**Why the prior Then clause was vacuous.** It read:

> **Then** the probe skill appears in the rendered prompt-input skill roots — establishing that
> Codex resolves the forward-slash form.

Cell p1 is decisive: with **no entry at all** the probe skill still appears, so appearance is the
**default state**. A failure by Codex to resolve the forward-slash form would merely leave the
declared entry inert, and the skill would appear anyway by virtue of its placement under the root.
The criterion was therefore satisfied by the null state — the vacuous-green shape named in
`.claude/rules/moai/development/verification-completeness.md` §1.1. The Fallback clause rested on
the same false premise (absence-in-slash-arm as the refutation signal) and so was rewritten in the
same pass rather than left behind. Cell p2b is the attribution control for the measurement itself:
without it, p2's zero could not be attributed to the `enabled` flag rather than to run-to-run
variance.

**AC-CSPS-001 is OPEN both before and after this amendment.** On darwin the forward-slash form and
the native form are the same string, so no verdict reachable on this host can close it. This
amendment converts a vacuous check into a meaningful one while leaving the criterion unsatisfied; it
grants no progress and cannot be self-serving.

**Gaps in the probes this amendment rests on.**

- `r1` in the rendered skill roots resolved to the real `/Users/goos/.agents/skills` even under an
  isolated `CODEX_HOME`, so the probes are **not hermetic** on a developer machine (a distinctive
  probe name was used so a collision would be visible). That this does not arise on a clean CI
  runner is an **INFERENCE**, not a measurement — no CI cell was run.
- The placement finding ("placement under a root introduces a skill; a config entry does not") rests
  on an arbitrary-path cell whose hit count was zero. Some property of that cell **other than** its
  position outside a skill root might explain the zero; no probe separated the two explanations, so
  the placement finding is cited only at that strength.

---

## AC-CSPS-002 — a backslash-shaped path is published, not skipped

**Given** `upsertCodexSkillDisable`, an empty (or entry-free) config content, and the package
separator seam set to `'\\'` (`configPathSeparator` overridden with `t.Cleanup` restore, test
non-parallel),
**When** it is called with a Windows-shaped path literal such as `C:\Users\u\.codex\skills\probe\SKILL.md`,
**Then** the returned verdict `Action` is `codexSkillDisableAppended` (not `codexSkillDisableSkipped`),
and the emitted content contains the line `path = "C:/Users/u/.codex/skills/probe/SKILL.md"`.

- **The separator injection is load-bearing, not decoration.** `filepath.ToSlash` is the identity
  function on a `/`-separator host — measured, not assumed (`spec.md` M6,
  `go run .moai/reports/t540/lab/ts.go`). Without the seam this AC is **unsatisfiable on darwin**:
  the backslash survives and `:245` refuses. The claim "this measurement is platform-independent" is
  true **only** on top of `toConfigPath(p, sep)` (`spec.md §B.3.1`).
- This is what turns inference I1 ("Windows is broken") into a measured fact for the publisher half.
  The Codex-side half stays with AC-CSPS-001.
- Closes t502 **F4**: the guard at `internal/cli/codex_skills_disable.go:245` acquires its first test.
- The emitted line contains no `\` and no escape sequence.
- RED before the change: with the seam set to `'\\'` and no `toConfigPath` in place, the call returns
  `codexSkillDisableSkipped`. That is a real RED on this host.

---

## AC-CSPS-003 — genuinely unrepresentable characters are still refused, and `\` on a slash host still is

**Given** `upsertCodexSkillDisable`,
**When** it is called with a path containing `"`, LF, or CR (one case each),
**Then** each returns `codexSkillDisableSkipped` with the "cannot carry verbatim" reason, and the
content is returned byte-unchanged.

**Second arm (REQ-CSPS-010) — Given** the separator seam left at its production value on a
`/`-separator host, **When** `upsertCodexSkillDisable` is called with a path containing `\`,
**Then** it still returns `codexSkillDisableSkipped`.

- The second arm is what forbids an unconditional `strings.ReplaceAll(p, "\\", "/")`. A unix filename
  may legally contain `\`; converting it would turn a refusal into the publication of a wrong path.
  A mutation replacing `toConfigPath` with the unconditional form must fail this arm.
- This is the other half of the F4 closure: the refuse branch keeps a test, so a future change that
  widens the guard away entirely fails RED rather than passing silently.
- No emitted output may contain `\"`, `\\`, `\n`, or `\r` as an escape.

---

## AC-CSPS-004 — the readers convert before stat, and a resolving entry is not deleted

**Arm A (prune conversion wiring — this is where the RED lives, and it is the PRUNE reader only).
Given** the separator seam set to `'\\'` and `osStatFn` replaced by a recorder that captures the path
it receives, **and** a config entry declaring a host-absolute slash path (e.g. `/tmp/x/SKILL.md`,
which `classifyCodexSkillPath` accepts as `codexPathAbsolute` on this host), **When**
`judgeCodexSkillEntry` evaluates it, **Then** the recorder observes `\tmp\x\SKILL.md` — the
`fromConfigPath` form.

- RED before the change: the recorder observes `/tmp/x/SKILL.md`. A genuine RED on a `/`-separator
  host, which the un-seamed `filepath.FromSlash` formulation could not produce (`spec.md` M6).
- **Scoped to prune deliberately.** `codex_skills_prune.go:96` routes through the `osStatFn` seam;
  `doctor_codex.go:857` calls `os.Stat` **directly**, so a recorder cannot be placed there
  (`spec.md §G` gap 5). Introducing that seam is production-code change and is out of scope
  (`spec.md §F`).
- The declared path is host-absolute, not `C:/…`: on this host `filepath.IsAbs("C:/…")` is false
  (M6), so a `C:`-prefixed fixture classifies `codexPathRelative` and never reaches stat.

**Arm A' (doctor conversion — regression guard, claims NO RED). Given** the `doctor` stale-skill
scan after the change, **When** it resolves a slash-declared entry naming a file that exists,
**Then** it does not count that entry as missing.

- This arm asserts observable doctor OUTPUT, not the stat argument: the argument is unobservable
  here (`spec.md §G` gap 5), and `codexStaleSkillFinding` carries zero existing tests
  (`/usr/bin/grep -rn 'codexStaleSkillFinding' --include='*_test.go' internal/cli/` → no output,
  rc=1).
- **Regression guard, not a RED source.** On a `/`-separator host the conversion is identity, so this
  arm is green before and after. Reporting it as TDD evidence would be false; the M3 RED is carried
  by arm A (prune) and arm B.

**Arm B (pure function). Given** `fromConfigPath`, **When** called as
`fromConfigPath("C:/Users/u/SKILL.md", '\\')`, **Then** it returns `C:\Users\u\SKILL.md`; and
`fromConfigPath(p, '/')` returns `p` unchanged for every input.

- RED before the change: the function does not exist. A stub returning its input unchanged must fail
  the first assertion — that stub is the required mutation arm.

**Arm C (behaviour). Given** a config entry whose `path` is the slash form of a file that EXISTS
under `t.TempDir()`, **When** prune evaluates it, **Then** it returns `Eligible: false` with the
"the path resolves" skip reason.

- Negative arm required: an entry naming a file that does NOT exist still yields `Eligible: true`.
  Without it, a mutation making every entry unconditionally "resolve" passes.
- Arm C exercises the identity path on this host and is therefore a **regression guard**, not a RED
  source. Arms A and B carry the TDD obligation for M3.

---

## AC-CSPS-005 — publish → parse → stat round-trips to the same file

**Given** a real SKILL.md created under `t.TempDir()`,
**When** its path is published by `upsertCodexSkillDisable`, the resulting content is parsed by
`codexwiring.ParseSkillEntries`, and the parsed `Path` is resolved through
`classifyCodexSkillPath` + `fromConfigPath` and stat'd,
**Then** the stat succeeds and names the same file that was created.

- The assertion is on file identity (`os.SameFile`), not merely on "stat returned nil".
- On a `/`-separator host this is an identity round-trip and is a regression guard; its RED is
  carried by AC-CSPS-004 arms A and B, and this AC does not claim a RED of its own.

---

## AC-CSPS-006 — unix behaviour is byte-identical to base

**Given** the existing `internal/cli` test suite for the disable, prune, and doctor codex surfaces at
base `9ce792637`,
**When** the suite is run against the changed tree on a `/`-separator host,
**Then** every pre-existing test passes unchanged, and no pre-existing test's expectations were
edited to accommodate the change.

- With the seam at its production value the separator is `/`, so `toConfigPath` / `fromConfigPath`
  are identity (M6) and any required edit to an existing expectation is evidence the change does more
  than normalize.
- **Control required — a green suite is not evidence a suite ran.** Cite the number of tests actually
  executed alongside the pass verdict, and compare it against the same count at base:

  ```bash
  go test ./internal/cli/... -timeout 1800s -v > .moai/reports/t540/ac-006.log 2>&1; echo "rc=$?"
  /usr/bin/grep -c -- '--- PASS: ' .moai/reports/t540/ac-006.log
  ```

  The **base count** is measured ONCE, with the identical command, on the tree **before M1 lands**,
  and written to `.moai/reports/t540/ac-006-base.log`:

  ```bash
  # run at pre-flight, BEFORE any M1 change
  go test ./internal/cli/... -timeout 1800s -v > .moai/reports/t540/ac-006-base.log 2>&1; echo "rc=$?"
  /usr/bin/grep -c -- '--- PASS: ' .moai/reports/t540/ac-006-base.log   # → BEFORE
  ```

  Verdict predicate, stated mechanically rather than left to judgment:
  **PASS iff `AFTER >= BEFORE` AND `AFTER > 0`.** Anything else — `AFTER == 0`, `AFTER < BEFORE`, or
  a missing base log — is reported as **"not measurable"**, never as "everything passed". A
  zero-match selector and a package that failed to compile both present as a quiet green.
- [HARD] Count with `'--- PASS: '` (trailing space), not a `$`-anchored pattern: Go appends ` (0.06s)`
  to the line, so a `$` anchor matches nothing and reads as a permanent 0.
- Evidence: the command, `rc`, the PASS count, and
  `git diff --name-only "$CARD_BASE"..HEAD -- 'internal/cli/*_test.go'` showing which test files
  changed, with each change justified as an ADDITION.

---

## AC-CSPS-007 — an existing backslash entry is updated, not duplicated

**Given** the separator seam set to `'\\'`, and a config already containing one `[[skills.config]]`
entry whose `path` is the native backslash form of a skill
(`path = "C:\Users\u\.codex\skills\probe\SKILL.md"`, verbatim, no escape) with `enabled = true`,
**When** `upsertCodexSkillDisable` is called for that same skill (whose path converts to the slash
form),
**Then** the verdict `Action` is `codexSkillDisableUpdated`, the entry's `enabled` line becomes
`false`, and the emitted content contains exactly ONE `[[skills.config]]` header — no second entry is
appended.

- Rationale (`spec.md §D.4`): without normalized comparison at `:252` the match fails,
  `appendDisableEntry` runs, and the card manufactures the duplicate debt `moai clean --codex-skills`
  (t506) exists to clear.
- The stored `path` line stays in its original backslash form — this verb rewrites only `enabled`.
  Asserting the path line is UNCHANGED is part of this AC.

**Second arm — duplicate detection still holds under the new comparison. Given** the same seam and a
config holding TWO entries for the same skill, one backslash-shaped and one slash-shaped, **When**
`upsertCodexSkillDisable` is called for it, **Then** it returns the `len(matches) > 1` skip
(`:257`) with count 2, and the content is returned byte-unchanged.

- **Why this arm exists.** It is NOT guarding against two comparison sites disagreeing — per
  `spec.md` M4 there is exactly one comparison, at `:252`, and `:257` performs no comparison at all.
  That earlier premise was retracted (`spec.md §D.4`). The arm's actual purpose: normalization
  changes **which entries land in `matches`**, so the duplicate branch's input distribution changes,
  and this pins that it still reaches the skip rather than silently updating one entry or appending a
  third.

---

## AC-CSPS-008 — the parser file is unmodified

**Given** this card's branch,
**When** the card's diff range is measured with a re-derived left edge,
**Then** `internal/codexwiring/skills.go` appears in ZERO changed-file rows.

```bash
git fetch origin develop
CARD_BASE=$(git merge-base origin/develop HEAD)
git diff --name-only "$CARD_BASE"..HEAD | wc -l                                  # control: must be >= 1
git diff --name-only "$CARD_BASE"..HEAD -- 'internal/codexwiring/skills.go'      # probe: must be empty
```

- [HARD] The left edge is re-derived at read time. A literal base SHA (`9ce792637`) is valid only as
  a dated anchor recording where a measurement was taken — never as the range's left edge.
- A control of `0` is reported as **"not measurable"**, never as "no changes". An empty probe under a
  zero control asserts nothing.

---

## §D.1 Severity

MUST-PASS: AC-CSPS-001 (gate), 002, 003, 004, 007, 008.
MUST-PASS with justified exception path: AC-CSPS-005, AC-CSPS-006.

## §D.2 Definition of Done

- AC-CSPS-001 measured on a real Windows host with both arms, evidence written; or the card STOPPED
  and the refutation reported.
- All remaining ACs GREEN, each with the command, its verbatim output, and its exit code cited.
- Every test added is demonstrated RED before the change (cycle_type tdd), **with the RED source
  named per AC**: AC-CSPS-002 and AC-CSPS-007 RED under the seam set to `'\\'`; AC-CSPS-004 arm A
  (**prune only**) and arm B carry M3's RED. AC-CSPS-004 arm A' (doctor), arm C, AC-CSPS-005, and
  AC-CSPS-006 are regression guards on a `/`-separator host and do NOT claim a RED — saying so
  explicitly is what stops a regression guard being reported as TDD evidence.
- The doctor reader's stat argument is NOT claimed as observed (`spec.md §G` gap 5). A completion
  report asserting a doctor-side stat measurement, or silently adding a seam at
  `doctor_codex.go:857` to obtain one, fails this DoD.
- An absence-shaped assertion (AC-CSPS-003, AC-CSPS-008) is satisfied vacuously by a broken harness,
  so a mutation arm is required for each: the identity-stub mutant (must fail AC-CSPS-004 arm B and
  AC-CSPS-002), the unconditional-`ReplaceAll` mutant (must fail AC-CSPS-003 arm 2), and a
  parser-touch mutant (must fail AC-CSPS-008). Mutants NOT caught are recorded too — they draw the
  guard's boundary.
- AC-CSPS-006 cites its executed-test count, not only its pass verdict.
- `internal/codexwiring/skills.go` unmodified (AC-CSPS-008).
- No experiment wrote the real `~/.codex/config.toml`.
- Every commit message contains `t540`.
