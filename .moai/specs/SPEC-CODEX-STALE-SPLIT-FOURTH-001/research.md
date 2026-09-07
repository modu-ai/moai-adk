# SPEC-CODEX-STALE-SPLIT-FOURTH-001 — research

Derived from `.moai/reports/t534/reproduction.md`. Tree pin: **`e0c904f58`** (= `origin/develop`,
the tip carrying t508), worktree `.claude/worktrees/t534`, branch `WT-stale-msg-polarity`.

Every measurement below was made in this tree. A cell citing no SHA is not a baseline
(verification-claim-integrity §2).

---

## §1 The reproduction — observed, not inferred

Fixture (`/tmp/t534lab/home/config.toml`), two entries, both with paths that do not exist:

```toml
[[skills.config]]
path = "/tmp/t534lab/definitely-absent/SKILL.md"
enabled = yes          # declared, non-boolean

[[skills.config]]
path = "/tmp/t534lab/also-absent/SKILL.md"
enabled = true         # control — bare boolean
```

```
$ CODEX_HOME=/tmp/t534lab/home go run ./cmd/moai doctor --verbose
doctor rc=1
```

Two lines from that one run (box-drawing stripped, whitespace collapsed):

```
with a path that no longer exists (1 enabled, 0 disabled, 1 unspecified)
declare `enabled` with a value that is not a bare TOML boolean
```

The control is load-bearing: the `enabled = true` entry lands in `1 enabled` correctly, which
establishes that the split works for three states and fails for the fourth — rather than being
broken generally.

`rc=1` is attributable to the **fatal** `enabled`-shape finding, not to the advisory line. The
advisory's grade is unchanged by this SPEC (REQ-SSF-005).

---

## §2 Mechanism — read from the landed source

### §2.1 The parser is four-state

`internal/codexwiring/skills.go:22-56` reports `SkillEnabledUnspecified` (no key),
`SkillEnabledTrue`, `SkillEnabledFalse`, and `SkillEnabledNonBoolean` (t508's addition: a key IS
declared, with a value that is not a bare boolean).

### §2.2 The consumer is three-bucket

`internal/cli/doctor_codex.go:864-871`:

```go
switch e.Enabled {
case codexwiring.SkillEnabledTrue:
    missingEnabled++
case codexwiring.SkillEnabledFalse:
    missingDisabled++
default:
    missingUnspecified++      // catches Unspecified AND NonBoolean
}
```

The `default:` arm is the defect: it folds a four-state reading into a three-bucket partition, and
the surviving label asserts something false about the folded-in state.

### §2.3 The render site

`internal/cli/doctor_codex.go:903-905` — one format string, one production site (verified: the
literal appears in exactly one non-test file):

```
$ grep -rn "with a path that no longer exists" --include='*.go' . | grep -v '_test.go'
internal/cli/doctor_codex.go:904:			"; %d with a path that no longer exists (%d enabled, %d disabled, %d unspecified) — remove the stale entries or restore the skill files",
```

### §2.4 The fatal finding is already correct

`internal/cli/doctor_codex.go:762-768` counts `absent` and `nonBoolean` into separate variables and
sums them only for the `unusable` total. It is out of scope (spec.md §D) precisely because it is
already right.

---

## §3 Scope analysis

The mislabel requires BOTH conditions:

1. the entry's `path` no longer exists (`errors.Is(serr, fs.ErrNotExist)`) — the only branch that
   reaches the bucketing switch at all; and
2. its `enabled` is `NonBoolean`.

A non-boolean entry whose path RESOLVES never reaches the counter and is reported only by the fatal
finding, correctly. So a live-path config — the normal case — shows no contradiction.

The absent-key case is NOT mislabeled: `SkillEnabledUnspecified` genuinely is unspecified.
**Only the `NonBoolean` state is misnamed.**

---

## §4 The affected assertions — measured, and NOT as the card brief stated

The card brief cited `(1 enabled, 0 disabled, 1 unspecified)` as the broken assertion. Measured, that
string appears **only inside a comment** (`doctor_codex_test.go:331`, a historical narrative of what
the counts used to be before t508) — it is not a live assertion and nothing checks it.

The live assertion sites are three, all in `internal/cli/doctor_codex_test.go`:

```
$ grep -rn "disabled, [0-9]* unspecified" --include='*_test.go' .
internal/cli/doctor_codex_test.go:279:		"(1 enabled, 2 disabled, 0 unspecified)",
internal/cli/doctor_codex_test.go:331:// bucket, giving `(1 enabled, 0 disabled, 1 unspecified)`. Measured on
internal/cli/doctor_codex_test.go:335:// buckets, giving `(0 enabled, 0 disabled, 2 unspecified)`.
internal/cli/doctor_codex_test.go:349:	if !strings.Contains(codexDetailText(check), "(0 enabled, 0 disabled, 2 unspecified)") {
```

plus, from the same sweep:

```
internal/cli/doctor_codex_test.go:943:	if !strings.Contains(detail, "(0 enabled, 1 disabled, 0 unspecified)") {
```

| Line | Enclosing test | Current asserted string |
|---|---|---|
| 279 | `TestCheckCodexWiring_StaleHomeSkillsReported` | `(1 enabled, 2 disabled, 0 unspecified)` |
| 349 | `TestCheckCodexWiring_UnspecifiedEnabledReportedSeparately` | `(0 enabled, 0 disabled, 2 unspecified)` |
| 943 | `TestCodexSkillPath_AbsoluteExistingAndMissing` | `(0 enabled, 1 disabled, 0 unspecified)` |

Lines 331 and 335 are comment text inside the doc block of the line-349 test; 335 restates the
live assertion and must be re-stated with it, 331 is a historical note about the pre-t508 counts and
stays as history.

Baseline (line-349 test, green before the change):

```
$ go test ./internal/cli/ -run 'TestCheckCodexWiring_UnspecifiedEnabledReportedSeparately' -count=1 -v
=== RUN   TestCheckCodexWiring_UnspecifiedEnabledReportedSeparately
--- PASS: TestCheckCodexWiring_UnspecifiedEnabledReportedSeparately (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.946s
```

### §4.1 Why the count of broken assertions depends on §B.1

Under **unconditional** rendering all three break, because `strings.Contains` on
`"(1 enabled, 2 disabled, 0 unspecified)"` fails once the closing paren moves to
`"(1 enabled, 2 disabled, 0 unspecified, 0 non-boolean)"`.

Under **conditional** rendering (fourth member printed only when non-zero) only line 349 breaks.

That asymmetry is the one measured argument for the conditional form, and spec.md §B.1 rejects it:
three assertions in one file is a bounded cost, and a partition whose members appear and disappear
cannot be read by summing.

The line-349 test's doc comment currently ends:

> The declared split deliberately does NOT grow a fourth bucket: this finding's message template is
> preserved (REQ-CEF-010), and the non-boolean shape gets its own fatal finding rather than a wider
> advisory count.

That paragraph is the recorded prior decision, and it must be **rewritten with its reversal
rationale, not deleted** — the same discipline t508's `REQ-CEF-013` applied to the sites it reversed.

---

## §5 Why the fourth bucket was not added at the time

`REQ-CEF-010` required the stale-path finding to keep its trigger conditions, grade, and message
**template**. t508 read that as forbidding a fourth `%d`, and recorded the resulting imprecision
rather than silently widening its own scope. This card exists because the implementer flagged it
instead of leaving it.

t508's own SPEC had already declared that the declared-split **counts** change as an expected
consequence of `REQ-CEF-001`, and its anti-pattern list says "do not fix the counts back". So the
requirement froze the template's SHAPE, never the numbers. The question this card settles is
whether the shape itself should now change — decided yes, spec.md §B.

---

## §6 Gaps — what was NOT observed

- Only one fixture shape was run in the reproduction. Whether a config mixing absent-key and
  non-boolean entries renders both correctly in the **fatal** line was not separately probed; source
  reading (`doctor_codex.go:762-768`) says it does, but that is a reading, not a measurement.
- The real `~/.codex` was not involved: `CODEX_HOME` was pinned to the fixture for the whole run,
  and nothing was written.
- Frequency in the wild is unmeasured. On the reference machine all 49 entries declare a bare
  `enabled = false`, so none is in the mislabeled population there.
- No RED baseline exists yet for the new guard test (AC-SSF-001), because that test is authored in
  run-phase M2. Per `verification-completeness.md` §2.1 a criterion with no observed RED is
  classified a regression guard, not MUST-PASS — see acceptance.md §D.2.
- Whether any consumer outside this repository parses the advisory line's parenthesis was not
  investigated. It is a human-facing `moai doctor` detail string, not a documented machine
  interface.

## §7 Residual risk

- A downstream reader (a script, a support macro) that greps the advisory line for the exact
  three-member parenthesis will stop matching. Unmeasured (§6), and accepted: the string is a
  human-facing diagnostic.
- The word `non-boolean` is chosen to match the fatal finding's existing vocabulary ("a value that
  is not a bare TOML boolean"). If a later card renames that vocabulary, the two surfaces drift
  apart again.
