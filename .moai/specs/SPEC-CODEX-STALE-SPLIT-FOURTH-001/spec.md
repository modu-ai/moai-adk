---
id: SPEC-CODEX-STALE-SPLIT-FOURTH-001
title: "Stale-path advisory split gains a fourth bucket for the non-boolean `enabled` state"
version: 0.1.0
status: completed
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P3
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
tags: "codex, doctor, stale-finding, message-accuracy, t534"
era: V3R6
tier: S
related_specs: [SPEC-CODEX-ENABLED-FATAL-001, SPEC-CODEX-SKILL-PATH-001, SPEC-CODEX-WIRING-001]
---

# SPEC-CODEX-STALE-SPLIT-FOURTH-001 — one entry, two names, one run

## HISTORY

| Date | Change | Author |
|---|---|---|
| 2026-09-07 | Initial draft. Derived from `.moai/reports/t534/reproduction.md` (reproduction with control, mechanism read from source, scope analysis, judgment, rejected alternatives). Operator decision of 2026-09-07 (lane-3 session) recorded in §B. Declared `tier: S` — one switch arm and one format string in one function. | manager-spec |

---

## §A Context — the defect, as measured

`internal/cli/doctor_codex.go:864-871` buckets missing-path `[[skills.config]]` entries by their
`enabled` state into **three** counters, using a `default:` arm. Since SPEC-CODEX-ENABLED-FATAL-001
(t508) the parser at `internal/codexwiring/skills.go:22-56` reports **four** states, so that
`default:` arm now catches two of them: `SkillEnabledUnspecified` (no key declared) and
`SkillEnabledNonBoolean` (a key IS declared, with a value that is not a bare TOML boolean).

Both then render as `unspecified` in the advisory line at `doctor_codex.go:903-905`.

The consequence is that one entry is named two different ways inside a single `moai doctor` run.
Reproduced in this tree at `e0c904f58` (full fixture, control, and command in
`.moai/reports/t534/reproduction.md`); both lines below come from that one run:

```
with a path that no longer exists (1 enabled, 0 disabled, 1 unspecified)
declare `enabled` with a value that is not a bare TOML boolean
```

The `enabled = yes` entry is the `1 unspecified` of the first line and the subject of the second.
The control entry (`enabled = true`, path likewise absent) is the `1 enabled`, correctly.

The surviving label does not merely omit the fourth state — it **asserts something false about it**:
`unspecified` says no `enabled` key was declared, when one was.

**Scope is narrower than it first appears.** The mislabel requires BOTH that the entry's `path` no
longer exists (the only branch reaching the bucketing switch) AND that its `enabled` is
`NonBoolean`. A non-boolean entry whose path resolves never reaches the counter and is reported
only by the fatal finding, correctly. The absent-key population is **not** mislabeled —
`unspecified` remains the right word for it.

---

## §B Decision — add the fourth bucket

Operator decision, 2026-09-07, lane-3 session. The advisory line becomes:

```
(N enabled, N disabled, N unspecified, N non-boolean)
```

### Why

The advisory line's stated purpose is to split the missing-path entries **by their `enabled`
state**, so a reader can tell live breakage from stale bookkeeping. With four states and three
buckets the split no longer does the thing it exists to do, and the fold-in is not a silent omission
but a false assertion (§A).

`REQ-CEF-010` of SPEC-CODEX-ENABLED-FATAL-001 froze the message template's **shape**, not its
numbers: t508's own SPEC declared the counts changing as an expected consequence of `REQ-CEF-001`
and listed "do not fix the counts back" as an anti-pattern. That fence existed so t508's parser
change could not silently break a neighbouring feature, and it did its job — t508 landed with the
stale finding's behaviour intact. Re-reading it now as a permanent ban on ever widening the split
would extend a scope fence into a design decision nobody made.

### Why not the alternatives

- **Drop non-boolean entries from the advisory count.** Keeps the template byte-identical, but the
  leading `%d with a path that no longer exists` would then undercount entries whose paths are
  genuinely absent. That trades a wrong label for a wrong number — the worse trade, because the
  number is what the "remove the stale entries" directive is sized against.
- **Append a correction clause after the template.** Preserves the template and adds information,
  but the same population would then be counted in three places in one run (inside `unspecified`,
  in the correction clause, and in the fatal finding). The defect being fixed is *two names in one
  run*; this makes it three.
- **Document the limit.** Leaves the inaccurate sentence in front of the user and puts the
  explanation somewhere they will not be reading. Explaining a wrong sentence does not make it
  right.

### §B.1 The fourth counter renders unconditionally

**Decision: unconditional** — `0 non-boolean` prints when the bucket is empty, exactly as `0
disabled` prints today.

Ground: the parenthesis is a **partition** of the leading `%d` count, and a partition is read by
summing it. A conditionally-rendered member makes two different facts produce the same bytes — "this
run had no non-boolean entries" and "this build has no such bucket" — so a reader comparing two runs,
or a user comparing their output against a colleague's, cannot tell an empty bucket from an absent
feature. The three existing counters already render unconditionally (`0 disabled` appears in the §A
reproduction), so a conditional fourth would be the only asymmetric member of the group.

The cost is accepted and is bounded: unconditional rendering breaks **three** live assertions rather
than one, because `strings.Contains` on `"(1 enabled, 2 disabled, 0 unspecified)"` fails once the
closing paren moves. The three sites are enumerated in plan.md §B and are all in one file.

### §B.2 Supersession — explicit and partial

`REQ-CEF-010` and `AC-CEF-011` of the **completed** `SPEC-CODEX-ENABLED-FATAL-001` are superseded
**on one point only**: the message template's literal text may now carry a fourth `%d` member.

Not superseded, and preserved verbatim by this SPEC:

- the stale-path finding's **trigger conditions** — when it fires is unchanged;
- its **advisory grade** — it does not become fatal, and does not change `moai doctor`'s exit code;
- the leading `%d with a path that no longer exists` count and the trailing `— remove the stale
  entries or restore the skill files` directive;
- the word **`unspecified`**, which remains correct for the absent-key population and keeps that
  population.

Only the fourth state moves.

---

## §C Requirements (GEARS) — 8 of the Tier S ceiling of 8

### The split

**REQ-SSF-001** (Ubiquitous) — The stale-path advisory line shall partition the missing-path
entries into four buckets, one per `codexwiring.SkillEnabled*` state, with no state reached by a
`default:` arm.

**REQ-SSF-002** (Event-driven) — When the Codex Wiring check encounters an entry whose `path` is
absent and whose `enabled` reads as `SkillEnabledNonBoolean`, the advisory line shall count that
entry in the `non-boolean` bucket and shall not count it in the `unspecified` bucket.

**REQ-SSF-003** (Ubiquitous) — The `unspecified` bucket shall count exactly the entries whose
`enabled` reads as `SkillEnabledUnspecified`.

**REQ-SSF-004** (Ubiquitous) — The `non-boolean` counter shall render on every firing of the
advisory line, including when its value is zero (§B.1).

### Preservation

**REQ-SSF-005** (Unwanted) — The stale-path finding shall not change its trigger conditions, its
advisory grade, its leading missing-path count, or its trailing remove directive.

**REQ-SSF-006** (Unwanted) — The fatal `enabled`-shape finding (`codexEnabledShapeFinding`,
`internal/cli/doctor_codex.go:762-768`) shall not be modified; it already reports `absent` and
`nonBoolean` separately and correctly.

### Traceability

**REQ-SSF-007** (Ubiquitous) — This SPEC shall record the partial supersession of `REQ-CEF-010` and
`AC-CEF-011`, naming both what is superseded and what is preserved (§B.2).

**REQ-SSF-008** (Ubiquitous) — Each updated string assertion in
`internal/cli/doctor_codex_test.go` shall carry a comment naming this SPEC as the cause of the
changed string, so a later reader does not read the edit as a silent regression fix.

---

## §D Exclusions — what this SPEC does NOT build

This section is the SPEC's out-of-scope boundary. Each item below is deliberately excluded.

### Out of Scope — the fatal `enabled`-shape finding

- Any change to `codexEnabledShapeFinding`, its trigger, its grade, or its wording.
- Rationale: it already separates `absent` from `nonBoolean` (`doctor_codex.go:762-768`). Changing
  it would recreate the duplicate-naming defect this card removes.

### Out of Scope — the parser

- Any change to `internal/codexwiring/skills.go` or to the `SkillEnabled*` state set.
- Rationale: the parser is already four-state as of t508. The defect is entirely downstream of it,
  in one consuming switch.

### Out of Scope — `path` handling

- Widening, narrowing, or reclassifying the `path` axis: the missing / relative / oddly-formed /
  indeterminate classification and the `os.Stat` branch structure are untouched.
- Rationale: the bucketing switch is reached only from the `fs.ErrNotExist` branch, and this SPEC
  changes only what happens inside it.

### Out of Scope — the severity axis

- Promoting the stale-path finding from advisory to fatal, or otherwise changing `moai doctor`'s
  process exit-code contract.
- Rationale: the defect is a wrong word, not a wrong grade. The grade is preserved by REQ-SSF-005.

### Out of Scope — automatic repair of a user's config

- Removing, rewriting, or normalizing stale or non-boolean `[[skills.config]]` entries on the
  user's behalf.
- Rationale: `moai doctor` reports; it does not edit the user's configuration.

### Out of Scope — the real `~/.codex`

- Any read of or write to the machine's real `~/.codex/config.toml` by tests, fixtures, or tooling.
- Rationale: every fixture is created under `t.TempDir()` with `CODEX_HOME` pinned per fixture
  (AC-SSF-008). `t.Setenv("HOME", ...)` is prohibited — it pollutes parallel tests.

### Out of Scope — frequency measurement in the wild

- Establishing how often the mislabeled population occurs across real user configs.
- Rationale: rarity is not accuracy. On the reference machine all 49 entries declare a bare
  `enabled = false`, so none is in the mislabeled population there — and the sentence is still
  wrong when it fires.

---

## §E Cross-references

- `.moai/reports/t534/reproduction.md` — the reproduction, control, mechanism, scope analysis,
  judgment, and rejected alternatives this SPEC is derived from.
- `SPEC-CODEX-ENABLED-FATAL-001` — the predecessor; `REQ-CEF-010` / `AC-CEF-011` partially
  superseded per §B.2, and `§D.0` (the swept-count obligation) reused verbatim in acceptance.md.
- `internal/cli/doctor_codex.go:864-871` — the three-bucket switch.
- `internal/cli/doctor_codex.go:903-905` — the render site.
