# SPEC-CODEX-ENABLED-FATAL-001 — Implementation Plan

Tier **M** (corrected from S — spec.md HISTORY). Milestones are ordered by
**decision-reversibility**: the decisions most likely to be revised on review come first, and the
mechanical work sits at the bottom. M1 and M2 are the two places where a reviewer's disagreement
changes what gets built; M4 and M5 are consequences of them.

All `file:line` citations in this file were re-read in this tree at commit **`069795602`**, the
HEAD after the branch absorbed develop.

---

## §A Context

`enabled` is REQUIRED by codex and OPTIONAL to moai's parser, so `moai doctor` reports "wired and
consistent" for configs on which codex exits 1. Two divergence classes: silent (absent key, integer
value) and actively wrong (quoted values read as a healthy registration). Full context and
measurement: spec.md §A, research.md.

Operator decisions B.1 (both fatal shapes in scope, reversing the quoted-value leniency) and B.2
(`uikit.CheckFail`, doctor exits 1) are settled inputs.

**What changed since the first draft.** Card t506 landed on develop, shipping
`internal/cli/codex_skills_prune.go` — a surface that WRITES the user's `~/.codex/config.toml`. It
cannot create the fatal shape (measured: `.moai/reports/t508/develop-t506-impact.md`), but it will
rewrite a config codex cannot load and report success. The hazard is live rather than latent; the
scope is unchanged. t506 also grew `internal/codexwiring/skills.go` by ~144 lines and added a
second consumer of the parser — see §B.4.

---

## §B Known Issues in the current tree

### B.1 — the severity axis does not exist

`codexFinding` (`internal/cli/doctor_codex.go:114`) is `{summary, detail}` with no severity field;
every problem lands on `uikit.CheckWarn` (`:283`). **The severity axis must be built, not
selected.**

### B.2 — the `enabled` axis has no emission path

`codexStaleSkillFinding` (`:654`) returns early unless a path is missing or shape-unresolvable
(`if missing == 0 && unresolvedShape == 0`, `:732`). A live path with no `enabled` key satisfies
neither, so nothing is emitted at all. This is not a wording fix.

### B.3 — the reversal's blast radius: EIGHT sites in three files

The first draft said "two tests and two comments … all four". **Measured at `069795602`, it is
eight**, and three of them are assertion changes rather than comment changes:

| # | Site | Kind | What changes |
|---|---|---|---|
| 1 | `internal/codexwiring/skills.go:29` | production comment | `type SkillEnabled` doc comment — "the tri-state reading"; a fourth state makes "tri-state" wrong |
| 2 | `internal/codexwiring/skills.go:34` | production comment | `SkillEnabledUnspecified` doc comment says "or one whose value this parser does not recognise" — that clause becomes FALSE once a non-boolean value gets its own state |
| 3 | `internal/codexwiring/skills.go:125` | production comment | the leniency rationale; must be ANSWERED on both grounds (AC-CEF-013, spec.md §B.1), not deleted |
| 4 | `internal/codexwiring/skills.go:131` | production code | `skillEnabledKeyRe` itself — the three alternation branches are the leniency's mechanism |
| 5 | `internal/codexwiring/skills_test.go:52-65` | **assertion + comment** | `TestParseSkillEntriesEnabledAbsentIsUnspecified` — asserts an absent key reads UNSPECIFIED, with a comment saying Codex's default is unobserved. Measurement now settles that: codex does not default, it refuses. Both the assertion and its comment move |
| 6 | `internal/codexwiring/skills_test.go:74-78` | **assertion, FIVE rows** | `TestParseSkillEntriesEnabledQuoted`'s table has five rows and **all five change**: `"true"`→True, `"false"`→False, `'true'`→True, `'false'`→False all flip to the new non-boolean state; and the fifth row `enabled = yes` → `SkillEnabledUnspecified` **also changes**, because `yes` is a DECLARED non-boolean value and belongs in the new state rather than in "unspecified" |
| 7 | `internal/cli/doctor_codex_test.go:297` | fixture comment | "quoted string, still true" — no longer true |
| 8 | `internal/cli/doctor_codex_test.go:302` | **LIVE ASSERTION** | `strings.Contains(codexDetailText(check), "(1 enabled, 0 disabled, 1 unspecified)")` — a quoted `"true"` currently satisfies this and will not after M2. This is the second surface REQ-CEF-010's rescoping names |

Sites 5, 6 and 8 are **assertion changes**; the rest are comment changes. Site 6's fifth row is the
one most easily missed — it is not a quoted value, so a reader scanning for "the quoted rows" walks
past it.

### B.4 — the parser has a second consumer now (t506)

`internal/codexwiring/skills.go` gained `StartLine`, `EndLine` and `FirstUnrecognizedLine` on
`SkillEntry`, and `judgeCodexSkillEntry` in `internal/cli/codex_skills_prune.go:59` reads BOTH
`Enabled` and `FirstUnrecognizedLine` to decide whether to DELETE a user's registration. That
predicate carries an `@MX:WARN` deletion guard.

[HARD] **Any change to what those fields mean is checked against the prune predicate, not only
against the doctor.** Specifically: if M2's new state is implemented by making
`skillEnabledKeyRe` fail to match a quoted value, the quoted line becomes an *unrecognised* line,
`FirstUnrecognizedLine` goes non-negative, and the prune verb's disposition for that entry changes
from eligible to preserved. That direction is safe (it preserves rather than deletes), but it is a
behaviour change in another card's code and must be **stated in the run-phase evidence**, not
discovered later.

`internal/codexwiring/skills_extent_test.go` (new with t506, 221 lines) pins extent behaviour. M4's
new tests must not collide with it, and it must stay green (acceptance.md §D.5).

---

## §C Pre-flight

- `internal/cli/doctor_codex_enabled_test.go` is present and committed at `069795602`. Confirm the
  mutant (`:27`) FAILS and the control (`:49`) PASSES before any production edit — that pairing is
  the run's RED attribution. Both were measured at plan-time; re-measure, do not carry.
- Confirm the real `~/.codex` hash before starting; re-confirm at close (AC-CEF-012).
- No template change expected (research.md §7). Re-check if any edit strays outside `internal/cli/`
  and `internal/codexwiring/`.
- Read `internal/cli/codex_skills_prune.go:51-90` before touching `SkillEntry` semantics (§B.4).

---

## §D Constraints

- **Diagnosis only.** No repair affordance, no config rewrite (REQ-CEF-012).
- **Parser stays read-only.** Reading fidelity changes; posture does not (REQ-CEF-002).
- **The real `~/.codex` is untouchable** — another card's observation subject. All fixtures under
  `t.TempDir()`, `CODEX_HOME` pinned per fixture. Never `t.Setenv("HOME", ...)`.
- **`path` stays out of scope** (lab row 6).
- **t506's code is not modified** — only read (§B.4).
- Verification scoped to touched packages; CI owns the full-suite verdict.

---

## §F Milestones

### M1 — The severity axis (highest reversibility: a new type shape)

The decision most open to review: **how per-finding severity is represented.**

- Add a severity to `codexFinding` (`internal/cli/doctor_codex.go:114`).
- Replace the constant `check.Status = uikit.CheckWarn` (`:283`) with a fold over the findings'
  severities: any fatal ⇒ `CheckFail`; otherwise the current behaviour, byte-identical.
- Advisory findings must still surface their text in a run that also carries a fatal finding —
  this is what distinguishes a built axis from a flipped check (AC-CEF-010).

#### [HARD] The advisory grade MUST be the enum's zero value (REQ-CEF-014)

There are **13** `problems = append(` construction sites in `internal/cli/doctor_codex.go`:

```
194, 198, 201, 209, 223, 236, 242, 247, 249, 265, 273, 456, 466
```

Enumerated at `069795602` by `grep -n 'problems = append(' internal/cli/doctor_codex.go`; note that
223 and the two sites at 456 and 466 sit outside `checkCodexWiring` proper, so a read scoped to that
function alone counts ten and misses three.

Every one of them constructs a finding **without naming a severity**, so every one takes the Go zero
value. Were the enum declared fatal-first — the `iota` ordering its neighbour `SkillEnabled` uses at
`skills.go:34`, which is the ordering a copy of the adjacent style would produce — all 13 would
silently re-grade to fatal and `moai doctor` would exit 1 on every advisory machine in existence.

Declare advisory first. AC-CEF-016 pins it; AC-CEF-014 catches the observable consequence.

#### The design question, answered here rather than deferred

*Why a per-finding severity field rather than a second findings slice, or a severity returned
alongside the finding?*

- **A second slice (`problems` + `fatals`)** requires every one of the 13 sites to choose a slice,
  so the zero-value protection above does not exist: a new site appended to the wrong slice is a
  silent mis-grade with no default to fall back on. It also duplicates the join/width logic
  (`joinCodexSummaries`, `joinCodexDetails`, the `codexMessageWidthCeiling` tail-drop), which is
  where the rendering invariants live — two slices means two chances to get the tail-drop wrong.
- **A severity returned alongside** (`(codexFinding, severity, bool)`) puts the grade on the
  *call site* rather than on the finding, so the grade is lost the moment a finding is stored,
  appended, or passed on. The fold in `checkCodexWiring` needs the grade at status-computation
  time, which is after every producer has returned.
- **A field on `codexFinding`** keeps the grade attached to the thing it grades, survives every
  append and join unchanged, and — decisively — has a zero value that can be made safe. That last
  property is the tiebreaker, not aesthetics.

**Cost, stated:** every future construction site silently inherits advisory. That is the intended
default, and AC-CEF-016 is what stops it drifting.

Satisfies REQ-CEF-007, REQ-CEF-008, REQ-CEF-009, REQ-CEF-014. Verified by AC-CEF-010 (mixed
finding), AC-CEF-014 (advisory-only run), AC-CEF-016 (zero value).

### M2 — Parser reading fidelity (second-highest: reverses a documented decision)

The decision a reviewer is most likely to challenge on grounds of precedent.

- Widen the `enabled` reading so a **declared-but-non-boolean** value is distinguishable from an
  absent key. Today both collapse onto `SkillEnabledUnspecified` — `skills.go:34` says so in its own
  doc comment — which is why the integer case is as silent as the absent case.
- Reverse the quoted-value leniency: `enabled = "true"` and `enabled = 'false'` become non-boolean
  declarations rather than True/False readings.
- **Rewrite the `skills.go:125` comment, answering BOTH of its grounds** (spec.md §B.1: ground 1
  falsified by lab row 4; ground 2 true but never reaching codex). The existing argument must be
  answered in place, not deleted — a future reader who finds only a flipped regex and no reasoning
  will read this as a regression (REQ-CEF-013, AC-CEF-013).
- Work through the **eight** sites of §B.3, including the `enabled = yes` row that is not a quoted
  value and the live assertion at `doctor_codex_test.go:302`.
- Check the change against `judgeCodexSkillEntry` (§B.4) and record the prune-side consequence.

Satisfies REQ-CEF-001, REQ-CEF-002, REQ-CEF-013. Note the ordering dependency: M2 gives M3 something
to detect, but M1 gives it somewhere to report to.

### M3 — Detection and finding text

- Emit a fatal finding when any entry declares no `enabled` key, or declares it non-boolean.
- The finding text names the `enabled` key (every detection AC asserts this token).
- Text names the measured codex version **0.153.4**, not codex in general (REQ-CEF-011,
  AC-CEF-015) — the version history is unmeasured and the wording must not outrun the evidence.
- Respect the `codexMessageWidthCeiling` (113) convention already in the file: short summary, full
  enumeration in Detail.
- `path`-absent must NOT reach this path (REQ-CEF-006).
- **The declared-split counts at `doctor_codex.go:757` will move** as a consequence of M2 — a quoted
  `"true"` leaves the `enabled` bucket. That is expected (REQ-CEF-010 as rescoped), and it is the
  reason `doctor_codex_test.go:302` must be restated rather than treated as a regression. Do not
  "fix" the counts back.

Satisfies REQ-CEF-003, REQ-CEF-004, REQ-CEF-005, REQ-CEF-006, REQ-CEF-010, REQ-CEF-011.

### M4 — Test suite (mechanical, follows from M1-M3)

Ordered so the release-blocking gap identified in acceptance.md §D.2 closes first.

1. **[HARD] Author the process-level exit-code test FIRST**, before M1's severity lands, so its RED
   is observed against the pre-change behaviour and AC-CEF-008 can be promoted from regression guard
   to release-blocking within the run. The test BUILDS a binary and reads a **real process exit
   status**; asserting `doctorExitStatus(1)` returns a non-nil `exitCodeError` does NOT satisfy it —
   `internal/cli/exitcode_contract_test.go:30,38` and `internal/cli/binary_lag_test.go:102` already
   do exactly that in-package, which is the form AC-CEF-008 forbids.
2. Add AC-CEF-014's advisory-only fixture (a non-`enabled` warn finding: hooks-whitelist missing, or
   the sidecar mismatch) and AC-CEF-016's zero-value pin.
3. Extend the RED guard's fixture family to the four fatal shapes and three controls
   (AC-CEF-001..007), giving the controls their **positive-read clause** — the detail text must
   mention the fixture's entry, so a fixture degrading to the codex-not-in-play informational skip
   cannot pass a "not CheckFail" assertion (AC-CEF-005, 006, 007).
4. **Convert `TestCheckCodexWiring_DeclaredEnabledKeyStaysQuiet` into a table** with `bare_true` and
   `bare_false` cases — at `069795602` it is a single-case test, so AC-CEF-006's "table-driven row"
   has nothing to run (acceptance.md §D.1 selector note).
5. Add the read-only hash assertion (AC-CEF-012).
6. Update all eight reversal sites (§B.3) with rewritten rationale (AC-CEF-013).
7. Confirm the stale-path finding and its golden fixtures are unchanged apart from the two expected
   surfaces (AC-CEF-011).
8. Optional but cheap (acceptance.md §D.4): a prune-verb fixture asserting `moai clean
   --codex-skills` does not manufacture a fatal shape, replacing the READ-not-EXECUTED claim
   inherited from `.moai/reports/t508/develop-t506-impact.md`.

[HARD] Every new selector is checked against the **swept-count obligation** (acceptance.md §D.0): a
`-run` that matches nothing exits 0 and prints `PASS`. Cite `--- PASS:` lines, never `ok … [no tests
to run]`.

### M5 — Verification and close (mechanical)

- `go test ./internal/cli/... ./internal/codexwiring/... -count=1`; `go vet` on both.
- `go test ./internal/codexwiring/ -run 'TestSkillEntryExtent' -count=1 -v` (or the package run) —
  t506's extent pins must stay green.
- Cite the executed doctor exit codes verbatim, with the tree SHA.
- Cite the real `~/.codex` hash, before and after.
- Carry acceptance.md §D.4 gaps into the run-phase evidence verbatim, including the two that are
  new: the REQ-CEF-004 induction, and the prune verb having been read rather than executed.
- Record the prune-side consequence of the parser change (§B.4).

---

## §G Anti-Patterns to avoid

- **Declaring the severity enum fatal-first.** All 13 construction sites silently re-grade and
  `moai doctor` exits 1 everywhere. AC-CEF-016 exists for this; the `iota` ordering next door at
  `skills.go:34` is exactly the wrong thing to copy.
- **Flipping the whole check to `CheckFail`.** Satisfies the mutant, breaks every advisory finding.
  AC-CEF-010's mixed-finding case and AC-CEF-014's advisory-only run exist to catch this.
- **Satisfying AC-CEF-008 with an in-package `doctorExitStatus` assertion.** No process, no exit
  status, no evidence. The forbidden form already exists in the tree to copy from.
- **Citing an `ok … [no tests to run]` line as a pass.** It exits 0 and prints `PASS`; it ran
  nothing (acceptance.md §D.0).
- **Deleting the leniency comment instead of answering it** — and answering only ground 1. The
  comment makes two arguments; a reply to one leaves the other standing.
- **Missing the `enabled = yes` row** at `skills_test.go:78`. It is not quoted, so it survives a
  scan for "the quoted rows", and it changes.
- **Treating the moved declared-split counts as a regression.** They move because REQ-CEF-001 moved
  the buckets; REQ-CEF-010 is rescoped to say so.
- **Making bare `false` fatal.** It is the shape all 49 real entries use — it would break the only
  existing population while passing the `true` control.
- **Widening to `path`.** Lab row 6 measured `path` as optional; widening is a scope breach that
  none of AC-CEF-001..006 would detect.
- **Editing `internal/cli/codex_skills_prune.go`.** t506's code is read, never modified.
- **Touching the real `~/.codex`.** Another card's observation subject.

---

## §H Cross-References

- spec.md §B — the operator decisions and their arguments
- acceptance.md — the 16 criteria, §D.0 swept-count, §D.2 severity classification, §D.4 gaps
- research.md — the measurement and structural read this plan rests on
- `.moai/reports/t508/{codex-enabled-lab,red-baseline,develop-t506-impact}.md` — the evidence files
