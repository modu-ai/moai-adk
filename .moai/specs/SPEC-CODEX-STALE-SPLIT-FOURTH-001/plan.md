# SPEC-CODEX-STALE-SPLIT-FOURTH-001 — implementation plan

Tier S. Tree pin **`e0c904f58`**, worktree `.claude/worktrees/t534`, branch `WT-stale-msg-polarity`.

---

## §A Context

One switch arm and one format string in one function. The parser is already four-state; the defect
is entirely in a single consuming `switch` at `internal/cli/doctor_codex.go:864-871` and the format
string it feeds at `:903-905`. Full mechanism: research.md §2.

The predecessor SPEC (`SPEC-CODEX-ENABLED-FATAL-001`, t508) was correctly Tier M because it spanned
three layers and changed `moai doctor`'s exit-code contract. This card spans one function and changes
no contract.

---

## §B Known issues — the assertions that break

Measured at `e0c904f58` (research.md §4). The card brief cited
`(1 enabled, 0 disabled, 1 unspecified)` as the broken assertion; that string lives **only in a
comment** (`doctor_codex_test.go:331`, historical narrative) and is not asserted. The three live
sites are:

| Line | Enclosing test | Current asserted string | Becomes |
|---|---|---|---|
| 279 | `TestCheckCodexWiring_StaleHomeSkillsReported` | `(1 enabled, 2 disabled, 0 unspecified)` | `(1 enabled, 2 disabled, 0 unspecified, 0 non-boolean)` |
| 349 | `TestCheckCodexWiring_UnspecifiedEnabledReportedSeparately` | `(0 enabled, 0 disabled, 2 unspecified)` | `(0 enabled, 0 disabled, 1 unspecified, 1 non-boolean)` |
| 943 | `TestCodexSkillPath_AbsoluteExistingAndMissing` | `(0 enabled, 1 disabled, 0 unspecified)` | `(0 enabled, 1 disabled, 0 unspecified, 0 non-boolean)` |

All three break because the chosen render is unconditional (spec.md §B.1): `strings.Contains` on the
old three-member parenthesis fails once the closing paren moves. **This is an intended consequence
of this SPEC, not a regression** (REQ-SSF-008 / AC-SSF-007), and each edit carries a comment saying
so.

Line 335 (comment) restates the line-349 assertion and moves with it. Line 331 is a pre-t508
historical note and stays as history.

The line-349 test's doc block currently ends with a paragraph stating the split "deliberately does
NOT grow a fourth bucket … (REQ-CEF-010)". That is the recorded prior decision: it is **rewritten
with its reversal rationale, never deleted** — the discipline t508's own `REQ-CEF-013` applied to
the sites it reversed.

---

## §C Pre-flight

- [ ] `git rev-parse --short HEAD` reads `e0c904f58` (or the branch tip after an absorb — restamp
      every baseline citation if it moved).
- [ ] `go test ./internal/cli/ -run 'TestCheckCodexWiring_StaleHomeSkillsReported|TestCheckCodexWiring_UnspecifiedEnabledReportedSeparately|TestCodexSkillPath_AbsoluteExistingAndMissing' -count=1 -v`
      shows **three** `--- PASS:` lines. Fewer is a zero-match sweep (§D.0), not a green.
- [ ] `grep -rn "with a path that no longer exists" --include='*.go' . | grep -v '_test.go'` returns
      exactly one line. A second production site means the change has two homes.

---

## §D Constraints

- Only two files may change: `internal/cli/doctor_codex.go` and `internal/cli/doctor_codex_test.go`.
- `internal/codexwiring/` is untouched (spec.md §D).
- `codexEnabledShapeFinding` is untouched (AC-SSF-006).
- The advisory's grade, trigger, leading count, and remove directive are preserved (REQ-SSF-005).
- No fixture reads or writes the real `~/.codex`; no `t.Setenv("HOME", ...)` (AC-SSF-008).
- Lane-local verification is scoped to `./internal/cli/...`. The full suite is CI's verdict.

---

## §E Self-verification

Before declaring the run phase closed:

1. Every MUST-PASS criterion of acceptance.md §D.2 observed green **in one run**, with its §D.0
   swept-count satisfied (a `--- PASS:` line per named test, `no tests to run` is a FAIL).
2. `go test ./internal/cli/ -count=1` green; `go vet ./internal/cli/...` clean.
3. `git diff --name-only e0c904f58..HEAD -- '*.go'` lists exactly the two permitted files.
4. Every claim in the completion report carries the command run and its verbatim output. A figure
   without an attributed command is a Gap, not a Claim.

---

## §F Milestones

Ordered by decision-reversibility: the user-facing message shape first (the decision most likely to
be revised on review), the mechanical guard work after it.

### M1 — the render and the bucket (Priority High)

The reversible decision. `internal/cli/doctor_codex.go`:

- Replace the `default:` arm at `:864-871` with an explicit
  `case codexwiring.SkillEnabledNonBoolean:` incrementing a new `missingNonBoolean`, and keep an
  explicit arm for `SkillEnabledUnspecified`. **No state may be reached by a `default:` arm**
  (REQ-SSF-001) — a future fifth state must break the build rather than silently join a bucket.
  Where the Go type system does not force exhaustiveness, retain a `default:` that panics or
  records an explicit "unknown state" rather than folding into a named bucket.
- Add `missingNonBoolean` to the `missing` sum, so the leading count is unchanged in value
  (REQ-SSF-005).
- Extend the format string at `:903-905` to
  `(%d enabled, %d disabled, %d unspecified, %d non-boolean)`, rendering the fourth member
  unconditionally (REQ-SSF-004).
- Update the block comment above the render site: it currently explains why an absent key is counted
  as unspecified rather than folded into either side. That reasoning still holds for the absent-key
  population and must be extended, not replaced, to say why the non-boolean population is now its
  own member.

Then update the three assertions of §B, each with a why-comment naming this SPEC (REQ-SSF-008), and
rewrite the line-349 "deliberately does NOT grow a fourth bucket" paragraph with its reversal
rationale.

Exit: AC-SSF-002, 003, 004, 005, 007 green.

### M2 — the guard and its controls (Priority High)

New test in `internal/cli/doctor_codex_test.go`:
`TestCheckCodexWiring_NonBooleanEnabledCountedSeparatelyInStaleSplit`.

Fixture (via the existing `writeCodexHomeConfig` / `stubCodexHome` helpers, `t.TempDir()`-backed):
one entry with an absent path and a non-boolean `enabled`, plus the absent-key and bare-boolean
controls needed to make the assertion discriminating (acceptance.md §D.2 control obligation).

Assert the full parenthesis as an ordered phrase — not the `non-boolean` token alone. A bare token
assertion passes if every bucket collapses into `non-boolean`.

**Establish RED before GREEN.** The test is authored against `e0c904f58` behaviour first and
observed failing, so the guard is known to have a live edge; a guard first seen green proves
nothing. Record the RED output in progress.md §E.2.

Exit: AC-SSF-001 green, with its RED recorded.

### M3 — supersession and closure (Priority Medium)

Mechanical, and deliberately last.

- Confirm spec.md §B.2 records the partial supersession of `REQ-CEF-010` / `AC-CEF-011` and names
  what is preserved (REQ-SSF-007).
- Add one HISTORY row to `.moai/specs/SPEC-CODEX-ENABLED-FATAL-001/spec.md` pointing forward to this
  SPEC. **This is a documentation-only edit to a completed SPEC** — a HISTORY append, never a
  rewrite of its requirements or its status. If the orchestrator judges a completed SPEC immutable,
  skip it and record the skip: the supersession is already load-bearing in this SPEC's §B.2, and
  the forward pointer is a convenience.
- Run the §E self-verification batch and write §E.2 / §E.3 of progress.md.

Exit: AC-SSF-006, 008 confirmed; all MUST-PASS green in one run.

---

## §G Anti-patterns

- **Fixing the counts back.** If a test asserting the old three-member string is made to pass by
  narrowing the render instead of updating the assertion, the card has been reversed rather than
  implemented.
- **A bare string edit with no comment.** AC-SSF-007 exists because a later reader who sees three
  assertion strings changed with no explanation will read it as a silent regression fix.
- **Deleting the prior rationale.** The line-349 paragraph recording "deliberately does NOT grow a
  fourth bucket" is rewritten with its reversal reason. Deleting it erases the fact that a decision
  was reversed.
- **Widening to the fatal finding.** It is already correct (research.md §2.4). Touching it recreates
  the duplicate-naming defect.
- **A `default:` arm that names a bucket.** The defect being fixed is exactly this shape. A new
  `default: missingNonBoolean++` reproduces it one state later.
- **Conditional rendering as a shortcut to fewer test edits.** Two assertion edits saved is not a
  design argument; spec.md §B.1 settled this and AC-SSF-004 pins it.
- **Running the full local suite.** Lane-local verification is `./internal/cli/...`; CI owns the
  full-suite verdict.

---

## §H Cross-references

- `.moai/reports/t534/reproduction.md` — reproduction, control, judgment, rejected alternatives.
- research.md §4 — the measured assertion inventory (which corrects the card brief's citation).
- `SPEC-CODEX-ENABLED-FATAL-001` §D.0 — the swept-count obligation reused in acceptance.md.
