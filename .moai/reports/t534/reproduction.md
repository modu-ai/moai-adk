# t534 — the two-names-in-one-run defect, reproduced

card: t534 · worktree `.claude/worktrees/t534` · branch `WT-stale-msg-polarity`
tree base **`e0c904f58`** (= `origin/develop`, the tip carrying t508)

## Claim

The same `[[skills.config]]` entry is named two different ways inside a single `moai doctor`
run: correctly by the fatal finding, and as `unspecified` by the stale-path advisory line.

## Evidence — reproduced, not inferred

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

The `enabled = yes` entry is the `1 unspecified` in the first line and the subject of the second.
The control entry (`enabled = true`) is the `1 enabled`, correctly. So the split is right for three
states and wrong for the fourth.

## Mechanism — read from the landed source

`internal/codexwiring/skills.go:22-56` — the reading is now **four-state**:
`SkillEnabledUnspecified` (no key), `SkillEnabledTrue`, `SkillEnabledFalse`, and
`SkillEnabledNonBoolean` (t508's addition: a key IS declared, with a value that is not a bare
boolean).

`internal/cli/doctor_codex.go:864-871` — the stale-path finding buckets by three:

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
the surviving label (`unspecified`) asserts something false about the folded-in state — that no
`enabled` key was declared, when one was.

`internal/cli/doctor_codex.go:903-905` renders it:

```
"; %d with a path that no longer exists (%d enabled, %d disabled, %d unspecified) — remove the stale entries or restore the skill files"
```

## Scope — narrower than it first appears

The mislabel requires BOTH conditions:

1. the entry's `path` no longer exists (`errors.Is(serr, fs.ErrNotExist)`) — the only branch that
   reaches the bucketing switch at all; and
2. its `enabled` is `NonBoolean`.

A non-boolean entry whose path RESOLVES never reaches the counter and is reported only by the fatal
finding, correctly. So a live-path config — which is the normal case — shows no contradiction.

The absent-key case is NOT mislabeled: `SkillEnabledUnspecified` genuinely is unspecified, and
"unspecified" is the right word for it. **Only the `NonBoolean` state is misnamed.**

## Why the fourth bucket was not added at the time

`REQ-CEF-010` (SPEC-CODEX-ENABLED-FATAL-001) required the stale-path finding to keep its trigger
conditions, grade, and message **template**. t508 read that as forbidding a fourth `%d`, and
recorded the resulting imprecision rather than silently widening its own scope. The card exists
because the implementer flagged it instead of leaving it.

Worth noting for the judgment: t508's own SPEC already declared that the declared-split **counts**
change as an expected consequence of `REQ-CEF-001`, and its anti-pattern list says "do not fix the
counts back". So the requirement froze the template's SHAPE, never the numbers — the question this
card must settle is whether the shape itself should now change.

## Gaps

- Only this one fixture shape was run. Whether a config mixing absent-key and non-boolean entries
  renders both correctly in the fatal line was not separately probed here (the fatal finding does
  carry separate `absent` / `nonBoolean` clauses in source, `doctor_codex.go:762-768`).
- The real `~/.codex` was not involved: `CODEX_HOME` was pinned to the fixture for the whole run,
  and nothing was written.
- Frequency in the wild is unmeasured. On the reference machine all 49 entries declare a bare
  `enabled = false`, so none is in the mislabeled population there — but rarity is not accuracy.

---

## Judgment — add the fourth bucket (operator decision, 2026-09-07, lane-3 session)

Chosen: **(a)** revise the requirement and render
`(N enabled, N disabled, N unspecified, N non-boolean)`.

### Why

The advisory line's stated purpose is to split the missing-path entries **by their `enabled`
state**, so a reader can tell live breakage from stale bookkeeping. With four states and three
buckets, the split no longer does the thing it exists to do — and the surviving label does not
merely omit the fourth state, it *asserts something false about it*: `unspecified` says no key was
declared, when one was.

`REQ-CEF-010` froze the message **template's shape**, not its numbers — t508's own SPEC declared the
counts changing as an expected consequence and listed "do not fix the counts back" as an
anti-pattern. The fence existed so that t508's parser change could not silently break a neighbouring
feature, and it did its job: t508 landed with the stale finding's behaviour intact. Re-reading it
now as a permanent ban on ever widening the split would extend a scope fence into a design decision
nobody made.

### Why not the alternatives

- **Drop non-boolean entries from the advisory count** keeps the template byte-identical, but the
  leading `%d with a path that no longer exists` would then undercount the entries whose paths are
  genuinely absent. That trades a wrong label for a wrong number — a worse trade, because the number
  is what the "remove the stale entries" directive is sized against.
- **Append a correction clause after the template** preserves the template and adds information, but
  the same population would then be counted in three places in one run (inside `unspecified`, in the
  correction clause, and in the fatal finding). The defect being fixed is *two names in one run*;
  this makes it three.
- **Document the limit** leaves the inaccurate sentence in front of the user and puts the
  explanation somewhere they will not be reading. The card exists because the wording is wrong;
  explaining a wrong sentence does not make it right.

### What this obliges

The change is not free and the costs are accepted knowingly:

1. Tests asserting the exact rendered string break — `internal/cli/doctor_codex_test.go` around the
   `(1 enabled, 0 disabled, 1 unspecified)` assertion. That is an intended consequence, not a
   regression, and each updated assertion must say so.
2. `REQ-CEF-010` and `AC-CEF-011` of the completed `SPEC-CODEX-ENABLED-FATAL-001` are superseded on
   this specific point. The new SPEC must state the supersession explicitly and name what it does
   NOT change: the finding's trigger conditions and its advisory grade stay exactly as they are.
3. The absent-key population keeps the word `unspecified`, which remains correct for it. Only the
   fourth state moves.

## Correction — the card brief named the wrong assertion (orchestrator error)

The brief handed to `manager-spec` said the broken assertion was
`(1 enabled, 0 disabled, 1 unspecified)` in `internal/cli/doctor_codex_test.go`. Measured at
`e0c904f58`, that string appears at line 331 **inside a comment** — a historical note on the
pre-t508 counts — and nothing asserts it. The live assertions are three, and each carries a
different string:

```
$ grep -rn "disabled, [0-9]* unspecified" --include='*_test.go' internal/
internal/cli/doctor_codex_test.go:279:		"(1 enabled, 2 disabled, 0 unspecified)",
internal/cli/doctor_codex_test.go:331:// bucket, giving `(1 enabled, 0 disabled, 1 unspecified)`. Measured on
internal/cli/doctor_codex_test.go:335:// buckets, giving `(0 enabled, 0 disabled, 2 unspecified)`.
internal/cli/doctor_codex_test.go:349:	if !strings.Contains(codexDetailText(check), "(0 enabled, 0 disabled, 2 unspecified)") {
internal/cli/doctor_codex_test.go:943:	if !strings.Contains(detail, "(0 enabled, 1 disabled, 0 unspecified)") {
```

Found by `manager-spec`, which reported it rather than quietly working from the corrected list.
Re-measured here before adoption.

**Where the error came from.** The reference `doctor_codex_test.go:302` was carried over from
t508's plan-audit, which measured the tree **before** t508 landed. t508 then added tests to that
file and the lines moved; what sat at the old coordinates became a comment. A line reference was
relayed across a tree change without being re-measured — the same failure this session has been
correcting all day, this time in the direction of a *stale coordinate* rather than a mis-scoped
selector.

The practical consequence would have been real: a run-phase that edited "the assertion at :302"
would have edited a comment, left three live assertions failing, and had to diagnose that from a
red suite.
