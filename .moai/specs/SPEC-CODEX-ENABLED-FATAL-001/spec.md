---
id: SPEC-CODEX-ENABLED-FATAL-001
title: "Codex [[skills.config]].enabled fatal-shape detection in moai doctor"
version: 0.1.0
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
tier: S
related_specs: [SPEC-CODEX-SKILLCONFIG-SHAPE-001, SPEC-CODEX-SKILL-PATH-001, SPEC-CODEX-WIRING-001, SPEC-CODEX-MIRROR-DOCTOR-001]
---

# SPEC-CODEX-ENABLED-FATAL-001 — `enabled` is required by codex and optional to us

## HISTORY

| Date | Change | Author |
|---|---|---|
| 2026-09-07 | Initial draft. Derived from the t508 measurement lab (`.moai/reports/t508/codex-enabled-lab.md`) and RED baseline (`.moai/reports/t508/red-baseline.md`); operator decisions on scope and severity recorded in §B. | manager-spec |

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

### Structural facts (read from source, not inferred)

- `internal/codexwiring/skills.go:32` — `SkillEnabledUnspecified` is the first enum value,
  documented as "an entry declaring no `enabled` key … It asserts nothing." The fatal shape is
  modelled as an ordinary tri-state reading.
- `internal/codexwiring/skills.go` ~line 60 (`skillEnabledKeyRe` comment) — the quoted-value
  leniency is DELIBERATE, with a stated rationale: reading `enabled = "true"` as false "silently
  DEMOTES a live registration to stale bookkeeping, which is the more damaging misreading."
- `internal/codexwiring/skills_test.go:66` `TestParseSkillEntriesEnabledQuoted` asserts
  `'true'` → `SkillEnabledTrue`.
- `internal/cli/doctor_codex_test.go:297` — fixture comment "quoted string, still true".
- `internal/cli/doctor_codex.go:654` `codexStaleSkillFinding` returns `ok=false` unless a path is
  MISSING or its SHAPE is unresolvable (`if missing == 0 && unresolvedShape == 0`, ~line 732). A
  live path with no `enabled` key satisfies neither, so nothing is emitted at all.
- `internal/cli/doctor_codex.go:114` `codexFinding{summary, detail}` carries NO severity field;
  every problem lands on `uikit.CheckWarn` (~line 286). **There is no ERROR grade on this check —
  the severity axis must be BUILT, not selected.**
- `internal/cli/uikit/types.go:17` — `uikit.CheckFail` exists.
- `internal/cli/doctor.go:137` `doctorExitStatus` — any `CheckFail` makes `moai doctor` exit 1;
  `CheckWarn` stays exit 0, deliberately.

### Priority rationale — latent, not theoretical

Read-only census of the real `~/.codex` on this machine: 49 `[[skills.config]]` entries, 49/49
declaring a bare `enabled = false`. **Nothing on this machine is currently in the fatal set.** The
hazard is latent until a WRITING surface lands (cards t502, t506) and emits an entry in a shape
codex rejects — at which point the user's codex is dead and `moai doctor` says "wired and
consistent". P1 is for the ordering: the detector should exist before the writer.

---

## §B Operator decisions (recorded, not re-opened)

Both were answered by the operator on 2026-09-07 in the lane-3 session. They are inputs to this
SPEC, not open questions.

### B.1 Scope — both fatal shapes

Missing key AND non-boolean value (integer, double-quoted string, single-quoted string) are all
reported at the fatal grade.

**This REVERSES the documented, tested quoted-value leniency** cited in §A. The reversal is stated
here explicitly because the leniency was a deliberate, commented, test-pinned decision — reversing
one silently would be indistinguishable from not having noticed it.

The ground for reversing it: **the leniency's own stated rationale is falsified by measurement.**
It exists to avoid "demoting a LIVE registration to stale bookkeeping" — but row 4 shows codex
exits 1 before loading anything, so on that input there is no live registration to demote. The
choice is not between two readings of a working config; it is between saying nothing and saying
the config is fatal. A decision whose stated ground turned out to be false is not a decision to
preserve.

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

---

## §C Requirements (GEARS)

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
the check's reported status and `moai doctor`'s exit code shall be unchanged from the current
behaviour.

### Preservation

**REQ-CEF-010** (Ubiquitous) — The existing stale-path finding shall retain its current trigger
conditions, grade, and wording.

**REQ-CEF-011** (Where — capability gate) — Where a fatal finding's text characterises what codex
accepts, the text shall attribute that behaviour to the measured codex version rather than
asserting it holds for every codex release.

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

### Out of Scope — the writing surfaces

- Changing how any moai code path WRITES `[[skills.config]]` entries.
- Cards t502 and t506 own the writing surfaces. This SPEC builds the detector that will observe
  them; it does not constrain them.

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

Acceptance criteria are enumerated in `acceptance.md`. This is a Tier S SPEC; a separate
`acceptance.md` is emitted rather than inlining the criteria here, because the matrix covers six
measured fixture rows plus a process-level exit-code criterion and does not compress usefully.

---

## §F Cross-references

- `.moai/reports/t508/codex-enabled-lab.md` — the six-row acceptance lab (factual basis for §A).
- `.moai/reports/t508/red-baseline.md` — the RED baseline and structural findings.
- SPEC-CODEX-SKILLCONFIG-SHAPE-001, SPEC-CODEX-SKILL-PATH-001 — the `path`-axis siblings this SPEC
  deliberately does not widen into.
- SPEC-CODEX-WIRING-001 — the Codex Wiring check this SPEC extends.
- SPEC-CODEX-MIRROR-DOCTOR-001 — a sibling doctor-surface card.
