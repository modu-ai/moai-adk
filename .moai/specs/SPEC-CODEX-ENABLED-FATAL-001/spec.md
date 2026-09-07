---
id: SPEC-CODEX-ENABLED-FATAL-001
title: "Codex [[skills.config]].enabled fatal-shape detection in moai doctor"
version: 0.2.0
status: draft
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: internal/cli + internal/codexwiring
lifecycle: spec-anchored
tags: "codex, skills-config, enabled, doctor, fatal-finding, severity, t508"
era: V3R6
tier: M
related_specs: [SPEC-CODEX-SKILLCONFIG-SHAPE-001, SPEC-CODEX-SKILL-PATH-001, SPEC-CODEX-WIRING-001, SPEC-CODEX-MIRROR-DOCTOR-001, SPEC-CODEX-GHOST-SKILLS-PRUNE-001]
---

# SPEC-CODEX-ENABLED-FATAL-001 — `enabled` is required by codex and optional to us

## HISTORY

| Date | Change | Author |
|---|---|---|
| 2026-09-07 | Initial draft. Derived from the t508 measurement lab (`.moai/reports/t508/codex-enabled-lab.md`) and RED baseline (`.moai/reports/t508/red-baseline.md`); operator decisions on scope and severity recorded in §B. Declared `tier: S`. | manager-spec |
| 2026-09-07 | **Tier CORRECTION: S to M.** Not tier-shopping — a correction of a mis-declaration, and one that RAISES the bar rather than lowering it (Tier M's plan-audit PASS threshold is 0.80 against Tier S's 0.75). Ground: the work spans three layers (parser reading fidelity in `internal/codexwiring`, a new per-finding severity axis in `internal/cli`, and a detection surface built on both), it REVERSES a test-pinned decision across eight sites in three files (plan.md §B.3), and it changes `moai doctor`'s **process exit-code contract**, which hooks and CI wrappers read. A single-layer, non-reversing, non-contract-changing card is Tier S; this is none of those. Budget moves to 16 requirements / 16 AC. | manager-spec |
| 2026-09-07 | Plan-audit FAIL (0.69 against 0.75) repair. Citations restamped against `069795602` (post-develop-absorb HEAD). Added REQ-CEF-013 (reversal rationale) and REQ-CEF-014 (advisory grade is the severity zero value); rescoped REQ-CEF-010; corrected §B.1 to answer BOTH grounds of the leniency comment; folded in the t506 landing (`.moai/reports/t508/develop-t506-impact.md`). | manager-spec |

---

## §A Context — the disagreement, as measured

Codex REQUIRES every `[[skills.config]]` entry to declare `enabled` as a **bare TOML boolean**.
moai's read-only parser treats the key as OPTIONAL and additionally accepts quoted forms. The two
readers therefore reach different verdicts on the same bytes.

Measured on codex-cli 0.153.4 in an isolated `CODEX_HOME`, probe `codex mcp list` (a
config-loading verb; `codex --version` does not load the config and is not a usable probe). Full
lab: `.moai/reports/t508/codex-enabled-lab.md`.

| # | `enabled` line | codex rc | `ParseSkillEntries` reading | agreement |
|---|---|---|---|---|
| 1 | `enabled = true` | 0 | `SkillEnabledTrue` | agrees — **control** |
| 2 | *(line absent)* | 1 — ``missing field `enabled` in `skills.config` `` | `SkillEnabledUnspecified` | SILENT |
| 3 | `enabled = 1` | 1 — ``invalid type: integer `1`, expected a boolean`` | `SkillEnabledUnspecified` (regex does not match) | SILENT |
| 4 | `enabled = "true"` | 1 — ``invalid type: string "true", expected a boolean`` | **`SkillEnabledTrue`** | **ACTIVELY WRONG** |
| 5 | `enabled = 'false'` | 1 — ``invalid type: string "false", expected a boolean`` | **`SkillEnabledFalse`** | **ACTIVELY WRONG** |
| 6 | `enabled = true`, `path` line ABSENT | 0 | — | agrees — **scope control** |

Row 1 makes rows 2-5 attributable: the lab config itself is loadable, so the varied line is the
cause. **Row 6 is load-bearing**: `path` is OPTIONAL to codex where `enabled` is not, so this
defect class is scoped to `enabled` and MUST NOT widen to `path`.

Rows 2-3 degrade to a NEUTRAL reading (`Unspecified` "asserts nothing"). Rows 4-5 degrade to a
CONFIDENT HEALTHY reading: moai reports a live, enabled registration for a config on which codex
will not start at all.

### Structural facts (read from source at `069795602`, not inferred)

Every `file:line` below was re-read in this tree at commit `069795602` — the HEAD after the branch
absorbed develop. The pre-absorb draft's citations had drifted (research.md §9); an uncheckable
citation is indistinguishable from an unverified claim, so they were restamped rather than carried.

- `internal/codexwiring/skills.go:34` — `SkillEnabledUnspecified` is the first enum value,
  documented verbatim as:

  > `SkillEnabledUnspecified` is an entry declaring no `enabled` key, **or one whose value this
  > parser does not recognise**. It asserts nothing.

  The emphasised clause is the whole defect in one sentence: the absent key and the unrecognised
  value are collapsed onto ONE state, so the parser cannot tell "nothing was said" from "something
  was said that codex rejects". REQ-CEF-001 exists to separate them.
- `internal/codexwiring/skills.go:29` — `type SkillEnabled int`, whose doc comment argues the three
  states are "kept distinct deliberately". A fourth state extends that argument rather than
  contradicting it.
- `internal/codexwiring/skills.go:125` — the quoted-value leniency is DELIBERATE and its rationale
  is quoted in full at §B.1 below (two grounds, both answered there).
- `internal/codexwiring/skills.go:131` — `skillEnabledKeyRe`, whose three alternation branches
  (bare, double-quoted, single-quoted) are the mechanism of the leniency.
- `internal/codexwiring/skills_test.go:52-64` `TestParseSkillEntriesEnabledAbsentIsUnspecified` and
  `:66-89` `TestParseSkillEntriesEnabledQuoted` — the two tests that pin the current reading.
- `internal/cli/doctor_codex_test.go:302` — a LIVE ASSERTION (not a comment) that a quoted `"true"`
  counts in the `enabled` bucket of the declared split.
- `internal/cli/doctor_codex.go:654` `codexStaleSkillFinding` returns `ok=false` unless a path is
  MISSING or its SHAPE is unresolvable (`if missing == 0 && unresolvedShape == 0`, line 732). A
  live path with no `enabled` key satisfies neither, so nothing is emitted at all.
- `internal/cli/doctor_codex.go:114` `codexFinding{summary, detail}` carries NO severity field;
  every problem lands on `uikit.CheckWarn` (line 283). **There is no ERROR grade on this check —
  the severity axis must be BUILT, not selected.**
- `internal/cli/uikit/types.go:17` — `uikit.CheckFail` exists.
- `internal/cli/doctor.go:142` `doctorExitStatus` — any `CheckFail` makes `moai doctor` exit 1;
  `CheckWarn` stays exit 0, deliberately.

### Priority rationale — no longer latent: a writing surface has landed

Read-only census of the real `~/.codex` on this machine: 49 `[[skills.config]]` entries, 49/49
declaring a bare `enabled = false`. **Nothing on this machine is currently in the fatal set.**

The original draft called the hazard "latent until a WRITING surface lands (cards t502, t506)".
**One of them has landed.** Card t506 (`SPEC-CODEX-GHOST-SKILLS-PRUNE-001`) shipped
`internal/cli/codex_skills_prune.go`, which writes the user's `~/.codex/config.toml` (`os.WriteFile`
at lines 205 and 211 of that file on develop). Measured impact:
`.moai/reports/t508/develop-t506-impact.md`.

Two findings from that measurement, both of which bear on priority and neither of which widens
scope:

1. **The prune verb CANNOT create the fatal shape.** Removal is whole-entry over a precomputed
   line range, so it never strips an `enabled` line while leaving its header standing; and
   `judgeCodexSkillEntry` skips any entry with `FirstUnrecognizedLine >= 0`, so an `enabled = 1`
   entry is PRESERVED rather than pruned. The two cards do not compose into a new defect.
2. **The prune verb WILL happily rewrite a config codex cannot load.** A user in the fatal set runs
   `moai clean --codex-skills`, sees moai report success, and still has a codex that exits 1 — with
   nothing anywhere telling them why. The finding this SPEC adds is the only surface that would.

So the priority argument moves from *latent* to *live*: not because a new defect exists, but
because moai now has a verb that reports success on a machine it cannot see is dead. **P1 stands,
and its ground is stronger than it was.** Scope does NOT widen to t506's code; see §D.

**Gap (carried, per verification-claim-integrity §1):** the prune verb's behaviour was **READ, not
executed**. Both claims above rest on reading `pruneCodexSkillEntries` and `judgeCodexSkillEntry`;
no fixture was run against the prune verb. A run-phase fixture asserting the prune does not
manufacture a fatal shape is cheap and SHOULD be added rather than inheriting this reading
(acceptance.md §D.4).

---

## §B Operator decisions (recorded, not re-opened)

Both were answered by the operator on 2026-09-07 in the lane-3 session. They are inputs to this
SPEC, not open questions.

### B.1 Scope — both fatal shapes

Missing key AND non-boolean value (integer, double-quoted string, single-quoted string) are all
reported at the fatal grade.

**This REVERSES the documented, tested quoted-value leniency.** The reversal is stated here
explicitly because the leniency was a deliberate, commented, test-pinned decision — reversing one
silently would be indistinguishable from not having noticed it.

The comment being reversed, quoted in FULL from `internal/codexwiring/skills.go:123-127` (no
elision — the previous draft dropped its final sentence, which carries one of its two grounds):

> The enabled matcher accepts a quoted value as well as a bare one. A quoted `"true"` is a TOML
> string rather than a boolean, so it is arguably malformed; reading it as false, however, silently
> DEMOTES a live registration to stale bookkeeping, which is the more damaging misreading. **The
> declared intent is unambiguous, so it is taken at face value and reported as declared.**

The comment makes **two** distinct arguments, and both must be answered:

**Ground 1 — "reading it as false demotes a LIVE registration."** *Falsified by measurement.* The
argument weighs two readings of a live registration, but lab row 4 establishes there is no live
registration on that input: codex exits 1 before loading anything, so the demotion the comment
guards against cannot occur. The real choice is not between two readings of a working config; it is
between saying nothing and saying the config is fatal.

**Ground 2 — "the declared intent is unambiguous, so it is taken at face value."** *True as stated,
and beside the point.* The author's intent may well be unambiguous — but **codex never reads the
intent.** It reads the bytes, finds a string where a boolean is required, and refuses to start.
Reporting a registration codex refuses to load is a confident-and-wrong reading regardless of what
the author meant. Face-value reporting of a declared intent is the right posture only where that
intent is what the consuming tool acts on; here it is not.

So the ground for reversing is not "a later author preferred a different trade-off". It is that one
ground was measured false and the other, though true, does not reach the consuming tool.
REQ-CEF-013 requires both answers to be written into the code at the site being reversed, so the
next reader finds a reply rather than a bare flipped regex.

### B.2 Severity — `uikit.CheckFail`, `moai doctor` exits 1

The finding is emitted at `uikit.CheckFail`, which by `doctorExitStatus` makes the process exit 1.

**This is not "the tool got noisier."** A user whose config carries such an entry ALREADY has a
codex that exits 1 on every invocation. Their tooling is already broken; the only thing this
change alters is whether `moai doctor` tells them so. Today doctor reports "wired and consistent"
about a machine where codex cannot start — the exit-code change makes doctor agree with reality
rather than contradict it.

The behavioural change is real and is scoped honestly: `moai doctor`'s exit code is read by hooks
and CI wrappers, and those wrappers will now fail on an affected machine. They will fail because
the machine is broken.

**Blast-radius consequence, made explicit.** Adding a severity field means every EXISTING
construction site gets the Go zero value. There are **13** `problems = append(` sites in
`internal/cli/doctor_codex.go` (lines 194, 198, 201, 209, 223, 236, 242, 247, 249, 265, 273, 456,
466 — enumerated at `069795602`; three of them — 223, 456, 466 — sit outside the window a narrower
read would have covered). Were the severity enum declared fatal-first (the `iota` ordering its
neighbour `SkillEnabled` uses at `skills.go:34`), all 13 would silently re-grade to fatal and
`moai doctor` would exit 1 on every one of them. REQ-CEF-014 forbids that ordering.

---

## §C Requirements (GEARS)

14 requirements against the Tier M ceiling of 16.

### Parser reading fidelity

**REQ-CEF-001** (Ubiquitous) — The skill-entry parser shall report an entry's `enabled`
declaration in a form that distinguishes three cases from one another: an absent key, a bare TOML
boolean, and a key declared with a value that is not a bare TOML boolean.

**REQ-CEF-002** (Ubiquitous, unwanted) — The skill-entry parser shall not write to any file. Its
reading fidelity changes; its read-only posture does not.

### Detection

**REQ-CEF-003** (event-driven) — When the Codex Wiring check reads a user-layer `config.toml`
containing a `[[skills.config]]` entry that declares no `enabled` key, the check shall emit a
finding at fatal grade whose text names the `enabled` key.

**REQ-CEF-004** (event-driven) — When the Codex Wiring check reads a `[[skills.config]]` entry
declaring `enabled` with a value that is not a bare TOML boolean, the check shall emit a finding
at fatal grade whose text names the `enabled` key.

> **Induction, stated rather than assumed.** REQ-CEF-004 generalises "not a bare TOML boolean" from
> THREE measured shapes: an integer (`1`), a double-quoted string (`"true"`), and a single-quoted
> string (`'false'`). Codex's error text is uniform across them (``invalid type: <type> …, expected
> a boolean``), which is the ground for generalising — but the generalisation IS an induction, not
> a measurement. Unmeasured members of the class (a float, an array, an inline table, a bareword
> such as `yes`) are expected to behave the same way and were not probed (acceptance.md §D.4).

**REQ-CEF-005** (state-driven) — While every `[[skills.config]]` entry in the read config declares
`enabled` as a bare TOML boolean, the Codex Wiring check shall not emit any fatal finding on the
`enabled` axis.

**REQ-CEF-006** (Ubiquitous, unwanted) — The Codex Wiring check shall not treat an absent `path`
key as a fatal shape. Codex accepts a `path`-less entry (§A row 6); this defect class is scoped to
`enabled`.

### Severity surface

**REQ-CEF-007** (Ubiquitous) — The Codex Wiring check shall carry a per-finding severity, so that
a fatal finding and an advisory finding can be produced by the same check run.

**REQ-CEF-008** (state-driven) — While the Codex Wiring check has produced at least one fatal
finding, the check's reported status shall be the fatal status, and `moai doctor` shall exit
non-zero.

**REQ-CEF-009** (state-driven) — While the Codex Wiring check has produced only advisory findings,
the check's reported status and `moai doctor`'s exit code shall be unchanged from the behaviour of
this tree at commit `069795602` — that is, `uikit.CheckWarn` and process exit 0.

**REQ-CEF-014** (Ubiquitous) — The advisory grade shall be the severity type's zero value, so that
a finding constructed without naming a severity is advisory. Codex Wiring has 13 existing
construction sites (§B.2); a fatal-first ordering would silently re-grade all of them.

### Preservation

**REQ-CEF-010** (Ubiquitous) — The existing stale-path finding shall retain its current trigger
conditions, grade, and **message template**.

> **Rescoped (was: "trigger conditions, grade, and wording").** The previous wording forbade what
> M2 requires. The declared-split text at `internal/cli/doctor_codex.go:757` —
> `"; %d with a path that no longer exists (%d enabled, %d disabled, %d unspecified) — remove the
> stale entries or restore the skill files"` — takes its counts from the very `enabled` reading
> REQ-CEF-001 changes. A quoted `"true"` entry moves OUT of the `enabled` bucket, so the RENDERED
> string necessarily changes. That change is a **consequence of REQ-CEF-001, not a regression**,
> and it is expected at exactly two surfaces: the format string at `doctor_codex.go:757` and the
> live assertion at `doctor_codex_test.go:302`. Everything else about the finding — when it fires,
> at what grade, and the template's literal text — is preserved.

**REQ-CEF-011** (Where — capability gate) — Where a fatal finding's text characterises what codex
accepts, the text shall name the measured codex version (0.153.4) rather than asserting the
behaviour holds for every codex release.

**REQ-CEF-013** (Ubiquitous) — Where this SPEC reverses the quoted-value leniency, the code site
being reversed shall carry a replacement rationale answering BOTH grounds of the comment it
replaces (§B.1), rather than having the previous rationale deleted.

### Boundaries

**REQ-CEF-012** (Ubiquitous, unwanted) — Nothing introduced by this SPEC shall modify, rewrite, or
repair the user's `config.toml`. The check reads and reports.

---

## §D Out of Scope

### Out of Scope — automatic repair

- Rewriting a user's `[[skills.config]]` entry to add or correct an `enabled` key.
- Offering a `--fix` flag, an interactive prompt, or any other repair affordance on this finding.
- Rationale: editing someone else's home config is a separate decision this card does not take.
  Diagnosis only.

### Out of Scope — the writing surfaces, t506 included

- Changing any moai code path that WRITES `[[skills.config]]` entries — including
  `internal/cli/codex_skills_prune.go`, which landed on develop as card t506 while this SPEC was in
  plan-phase.
- The t506 measurement (§A) establishes the two cards do not compose into a new defect. It
  strengthens this SPEC's priority argument and changes nothing about its scope.
- Card t502 (the other named writer) — whether it has landed was NOT checked.

### Out of Scope — the `path` key

- Widening fatal-grade detection to a missing, relative, or oddly-formed `path`.
- Rationale: §A row 6 measured `path` as OPTIONAL to codex. The existing stale-path finding
  continues to own that axis at advisory grade (REQ-CEF-010).

### Out of Scope — the real `~/.codex`

- Any modification of the machine's real `~/.codex/config.toml` by tests, fixtures, or tooling.
- It is another card's observation subject. Every fixture is created under `t.TempDir()` with
  `CODEX_HOME` pinned per fixture.

### Out of Scope — codex version compatibility research

- Determining in which codex release `enabled` became required, or whether older releases tolerate
  its absence. Only codex-cli 0.153.4 was measured. REQ-CEF-011 constrains the wording so the
  finding does not overclaim; establishing the version history is separate work.

### Out of Scope — multi-entry reporting semantics

- Deciding whether codex reports the first offending entry or all of them. Unmeasured; it affects
  finding wording only, never the verdict.

### Out of Scope — template changes

- No file under `internal/template/templates/` is expected to change. Measured: `grep -rl
  'skills\.config' internal/template/templates/` exits 1 with zero matches, against a non-vacuous
  control (`grep -rl 'moai'` matches 396 files), and the tree contains no `.go` files. Should
  run-phase discover a template touch is required after all, the Template-First mirror obligation
  and `make build` apply in full.

---

## §E Acceptance

Acceptance criteria are enumerated in `acceptance.md`: 16 criteria against the Tier M ceiling of
16. This SPEC emits a separate `acceptance.md` rather than inlining the criteria, because the
matrix covers six measured fixture rows, a process-level exit-code pair, a zero-value severity
pin, and a per-criterion RED-now attribution table that does not compress usefully.

**Every criterion carries a swept-count obligation** (acceptance.md §D.0). Six criteria in the
previous draft named tests that do not exist, and a Go `-run` selector matching nothing exits 0
while printing `PASS` — measured in this tree at `069795602`:

```
$ go test ./internal/cli/ -run 'TestCheckCodexWiring_NonBooleanEnabled' -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.673s [no tests to run]
$ echo $?
0
```

A criterion satisfiable by a run that executed nothing asserts nothing. §D.0 is the clause that
closes it.

---

## §F Cross-references

- `.moai/reports/t508/codex-enabled-lab.md` — the six-row acceptance lab (factual basis for §A).
- `.moai/reports/t508/red-baseline.md` — the RED baseline and structural findings.
- `.moai/reports/t508/develop-t506-impact.md` — the measured t506 impact, the citation-drift table,
  and the new-surface note.
- SPEC-CODEX-GHOST-SKILLS-PRUNE-001 (card t506) — the landed writing surface. Out of scope; its
  landing strengthens this SPEC's priority argument.
- SPEC-CODEX-SKILLCONFIG-SHAPE-001, SPEC-CODEX-SKILL-PATH-001 — the `path`-axis siblings this SPEC
  deliberately does not widen into.
- SPEC-CODEX-WIRING-001 — the Codex Wiring check this SPEC extends.
- SPEC-CODEX-MIRROR-DOCTOR-001 — a sibling doctor-surface card.

> **Traceability note for readers (`moai spec lint` cannot check this).** The `CoverageIncomplete`
> lint rule counts REQ-to-AC *references*; it cannot tell whether the AC a requirement points at is
> about the same subject. The previous draft mapped REQ-CEF-011 (finding text names the measured
> version) to AC-CEF-013 (reversed tests carry a rewritten rationale) — different subjects, so the
> requirement was uncovered and the AC orphaned while the coverage summary read "complete". The
> mapping in acceptance.md §D.3 is a reader-level obligation; check subjects, not counts.
