# SPEC-CODEX-ENABLED-FATAL-001 — Implementation Plan

Milestones are ordered by **decision-reversibility**: the decisions most likely to be revised on
review come first, and the mechanical work sits at the bottom. M1 and M2 are the two places where a
reviewer's disagreement changes what gets built; M4 and M5 are consequences of them.

---

## §A Context

`enabled` is REQUIRED by codex and OPTIONAL to moai's parser, so `moai doctor` reports "wired and
consistent" for configs on which codex exits 1. Two divergence classes: silent (absent key,
integer value) and actively wrong (quoted values read as a healthy registration). Full context and
measurement: spec.md §A, research.md.

Operator decisions B.1 (both fatal shapes in scope, reversing the quoted-value leniency) and B.2
(`uikit.CheckFail`, doctor exits 1) are settled inputs.

---

## §B Known Issues in the current tree

1. `codexFinding` has no severity field; every problem lands on `uikit.CheckWarn`. **The severity
   axis must be built, not selected.**
2. `codexStaleSkillFinding` returns early unless a path is missing or shape-unresolvable, so the
   `enabled` axis has no emission path at all — this is not a wording fix.
3. The quoted-value leniency is pinned by two tests and two comments. Reversing it touches all
   four, and the comments must be rewritten rather than deleted (AC-CEF-013).

---

## §C Pre-flight

- `internal/cli/doctor_codex_enabled_test.go` exists uncommitted; confirm mutant FAILS and control
  PASSES before any production edit. That pairing is the run's RED attribution.
- Confirm the real `~/.codex` hash before starting; re-confirm at close (AC-CEF-012).
- No template change expected (research.md §7). Re-check if any edit strays outside
  `internal/cli/` and `internal/codexwiring/`.

---

## §D Constraints

- **Diagnosis only.** No repair affordance, no config rewrite (REQ-CEF-012).
- **Parser stays read-only.** Reading fidelity changes; posture does not (REQ-CEF-002).
- **The real `~/.codex` is untouchable** — another card's observation subject. All fixtures under
  `t.TempDir()`, `CODEX_HOME` pinned per fixture. Never `t.Setenv("HOME", ...)`.
- **`path` stays out of scope** (lab row 6).
- Verification scoped to touched packages; CI owns the full-suite verdict.

---

## §F Milestones

### M1 — The severity axis (highest reversibility: a new type shape)

The decision most open to review: **how per-finding severity is represented.**

- Add a severity to `codexFinding` (`internal/cli/doctor_codex.go:114`).
- Replace the constant `check.Status = uikit.CheckWarn` (~line 286) with a fold over the findings'
  severities: any fatal ⇒ `CheckFail`; otherwise the current behaviour, byte-identical.
- Advisory findings must still surface their text in a run that also carries a fatal finding —
  this is what distinguishes a built axis from a flipped check.

Reviewer's question this milestone must answer: *why a per-finding field rather than a
second findings slice, or a severity returned alongside?* State the choice and its cost in the
commit message.

Satisfies REQ-CEF-007, REQ-CEF-008, REQ-CEF-009. Verified by AC-CEF-010 (mixed-finding case).

### M2 — Parser reading fidelity (second-highest: reverses a documented decision)

The decision a reviewer is most likely to challenge on grounds of precedent.

- Widen the `enabled` reading so a **declared-but-non-boolean** value is distinguishable from an
  absent key. Today both collapse onto `SkillEnabledUnspecified`, which is why the integer case is
  as silent as the absent case.
- Reverse the quoted-value leniency: `enabled = "true"` and `enabled = 'false'` become
  non-boolean declarations rather than True/False readings.
- **Rewrite the `skillEnabledKeyRe` comment.** The existing argument must be answered in place —
  its premise was falsified by lab row 4 — not deleted. A future reader who finds only a flipped
  regex and no reasoning will read this as a regression.

Satisfies REQ-CEF-001, REQ-CEF-002. Note the ordering dependency: M2 gives M3 something to detect,
but M1 gives it somewhere to report to.

### M3 — Detection and finding text

- Emit a fatal finding when any entry declares no `enabled` key, or declares it non-boolean.
- The finding text names the `enabled` key (every detection AC asserts this token).
- Text is attributed to the measured codex version, not to codex in general (REQ-CEF-011) — the
  version history is unmeasured and the wording must not outrun the evidence.
- Respect the `codexMessageWidthCeiling` (113) convention already in the file: short summary,
  full enumeration in Detail.
- `path`-absent must NOT reach this path (REQ-CEF-006).

Satisfies REQ-CEF-003, REQ-CEF-004, REQ-CEF-005, REQ-CEF-006, REQ-CEF-011.

### M4 — Test suite (mechanical, follows from M1-M3)

- Extend the existing RED guard's fixture family to the four fatal shapes and three controls
  (AC-CEF-001..007).
- Add the **executed** process-level exit-code pair, AC-CEF-008 / AC-CEF-009. [HARD] The exit path
  was read from `doctorExitStatus` during plan-phase and never run; run-phase must execute it. A
  source-read is not an observation.
- Add the read-only hash assertion (AC-CEF-012).
- Update the two reversed tests (`skills_test.go:66`, `doctor_codex_test.go:297`) with rewritten
  rationale (AC-CEF-013).
- Confirm the stale-path finding and its golden fixtures are unchanged (AC-CEF-011).

### M5 — Verification and close (mechanical)

- `go test ./internal/cli/... ./internal/codexwiring/... -count=1`; `go vet` on both.
- Cite the executed doctor exit codes verbatim.
- Cite the real `~/.codex` hash, before and after.
- Carry acceptance.md §D.4 gaps into the run-phase evidence verbatim.

---

## §G Anti-Patterns to avoid

- **Flipping the whole check to `CheckFail`.** Satisfies the mutant, breaks every advisory finding.
  AC-CEF-010's mixed-finding case exists to catch exactly this.
- **Deleting the leniency comment instead of answering it.** Leaves the old argument standing in
  the history with no reply.
- **Reading the exit code from source and reporting it as verified.** AC-CEF-008 is an execution
  criterion; a source-read is a hypothesis.
- **Making bare `false` fatal.** It is the shape all 49 real entries use — it would break the only
  existing population while passing the `true` control.
- **Widening to `path`.** Lab row 6 measured `path` as optional; widening is a scope breach that
  none of AC-CEF-001..006 would detect.
- **Touching the real `~/.codex`.** Another card's observation subject.

---

## §H Cross-References

- spec.md §B — the operator decisions and their arguments
- acceptance.md — the 13 criteria, controls, and the §D.4 gaps
- research.md — the measurement and structural read this plan rests on
- `.moai/reports/t508/{codex-enabled-lab,red-baseline}.md` — the evidence files
