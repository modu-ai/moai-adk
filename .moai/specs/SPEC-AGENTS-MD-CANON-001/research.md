# SPEC-AGENTS-MD-CANON-001 — Research

## Sources

- `.moai/reports/t82/codex-probe.md` — measured codex loading behavior via `codex debug
  prompt-input` (zero model calls) against a git-initialised 3-level fixture with a byte-ruler root
  document. Primary source for `spec.md` §A.2. Settles all three premises the first draft carried
  as blocking preconditions.
- `.moai/reports/t82/measurement.md` — measured always-loaded surface (bytes, token estimate).
  Primary source for `spec.md` §A.1.
- `internal/config/token_budget_guard.go` — the surface definition (`alwaysLoadedSurface()`), the
  budget constant, and the comment recording the 75,000 → 76,000 temporary raise.
- `CLAUDE.md` §9 — live evidence that `@`-imports resolve in this repo.
- `.moai/specs/SPEC-ALWAYS-LOADED-DIET-001/` — closed; source of the inherited stub +
  lazy-companion pattern and the budget guard.

## What the probe settled

| Premise | Status | Effect on the design |
|---|---|---|
| `project_doc_max_bytes` default | 32,768 B, confirmed | The ceiling is derived against this figure, not an assumed one |
| Nested `AGENTS.md` merge scope | git-root → CWD path only; unloaded at repo-root invocation; shares the budget root-first | **Overturned the card's nested-document leg.** Single root contract (Option A) is the design |
| Truncation visibility | Silent — stderr 0 B, exit 0 | The CI byte guard is the only detector, so it is blocking rather than advisory |
| P4 — cap-raise scope (`codex-probe-p4.md`) | Project scope works, and beats the user value — **only** under `trust_level = "trusted"`; ignored silently otherwise | Corrected the "cannot ship" premise, and strengthened the conclusion: the untrusted first session at 32,768 B is the binding case (REQ-AMC-018) |

Two further findings the probe produced that the card did not anticipate: truncation takes the
**tail** (the head survives), and outside a git repository only the CWD's own document loads.

**Fixture reproducibility.** `.moai/reports/t82/probe-fixture.sh` rebuilds the three-level fixture
from scratch and prints each recorded run with its expected result, so `AC-AMC-002` is executable
by someone other than the original author. The earlier scratchpad path was session-local and did
not survive.

## What was measured for this SPEC

The contract-layer sizing in `spec.md` §A.4 and `design.md` §1 is original to this SPEC. Commands
and outputs are recorded in `progress.md` §E.1.

Two headline figures:

- `[HARD]` line proxy across the always-loaded rules and `CLAUDE.md`: **32,543 B** — 99.3 % of the
  confirmed budget, leaving 225 B. A numeric fit and a practical failure.
- Upper bound on the Claude-only exclusion (six most Claude-mechanism-bound files):
  **14,360 B across 38 lines** — an upper bound only; the per-clause split is M1's deliverable.

**Both are line-level proxies and are labelled as such** (`spec.md` §A.4). Cross-checked error in
both directions: 15 of 93 rule markers are prose mentions rather than obligations (overcount), and
16 end in `:` and lead into uncounted bodies (undercount, unbounded). The consequence for the
design is that no requirement or criterion may depend on 32,543 B being exact — M1 works on clause
blocks and re-derives the ceiling, and the 8,192 B reserve is what absorbs the difference.

## What remains unmeasured

- `AGENTS.override.md` precedence and `project_doc_fallback_filenames` behavior — symbols observed,
  behavior not probed. The design depends on neither, so both are out of scope.
- A user's global `~/.codex/AGENTS.md` was not placed in the fixture. Chain-merge semantics imply it
  is consumed before the project document; `spec.md` §D.3 decides to warn about it in shipped
  documentation rather than treat it as measured.
- Whether trust registration is offered or recorded automatically on a first interactive run was
  not observed (`codex debug prompt-input` raises no trust prompt). The real acquisition path for a
  distributed user is M4/t88's to establish.
- `trust_level` values other than `"trusted"` were not measured.
- The probe ran on macOS with `codex-cli` 0.147.0 only. A different default elsewhere is not
  excluded (`spec.md` §D.9).

## Prior-art note

`SPEC-ALWAYS-LOADED-DIET-001` demonstrated the pattern this SPEC scales up: `goal-directive.md`
went to a 6,531 B stub with a 17,334 B lazy companion — a 72 % always-loaded reduction with no
obligation moved off the always-loaded surface. That precedent is why REQ-AMC-002 is a constraint
rather than an aspiration, and it bounds the plausibility of `design.md` §1.1's condensation target
(a 24.5 % trim in the pessimistic case).
