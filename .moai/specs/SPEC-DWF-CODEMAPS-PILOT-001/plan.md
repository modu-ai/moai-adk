# Implementation Plan — SPEC-DWF-CODEMAPS-PILOT-001

> Tier S. Single dynamic-workflow pilot, read-only, no production Go change. Deliverable: validated pattern entry (PRIMARY) + conditional local workflow script + documented falsification verdict.

## A. Context

This plan implements the pilot defined in `spec.md`: a read-only per-package codemaps extraction fan-out using the Claude Code dynamic-workflow primitive, gated by a falsification test that decides whether per-package LLM synthesis adds value over `go list -deps -json` + `go doc`. The pilot's success is the learning + the verdict (either direction).

Key facts already verified at plan time (do not re-derive):
- `go list ./... | wc -l` = 97 packages.
- `go list -deps -json ./internal/spec/` emits structured `ImportPath` records — the deterministic baseline is confirmed available.
- No `.claude/workflows/` directory exists yet (will be created locally only if the verdict is "value proven").
- No existing `SPEC-DWF-*` SPEC — ID is unique.

## B. Known Issues / Constraints Inherited

- **Workflow agents cannot prompt the user mid-run** — all preferences collected at GATE-2 before launch (REQ-DCP-002, REQ-DCP-008).
- **Deterministic-script constraint** — no wall-clock / random in the script body; package list via `args`; timestamp stamped after the run returns (REQ-DCP-006).
- **Non-template placement** — pilot script in local `.claude/workflows/` only; never in `internal/template/templates/` (REQ-DCP-009).
- **Dynamic workflows are a research preview** — require Claude Code v2.1.154+, a paid plan, and per-`/config` enablement on Pro. If unavailable in the run environment, the run-phase must fall back to documenting the primitive mechanics from the rule doc + running the deterministic baseline for the falsification comparison (the verdict can still be reached without a live workflow run — see Risk R1).

## C. Pre-flight Checks (run-phase entry)

Priority High — perform before any workflow launch:

1. Confirm Claude Code version supports dynamic workflows (v2.1.154+) and the feature is enabled; if not, record the limitation and proceed via the fallback path (R1).
2. Confirm `go list -deps -json ./...` runs clean and capture a baseline sample for 3-5 representative packages (e.g. `internal/spec`, `internal/cli`, `internal/template`, `cmd/moai`, `pkg/version`).
3. Confirm `.claude/workflows/` does NOT exist in the template suite (`grep -r "codemaps-extract" internal/template/templates/` must return nothing).

## D. Constraints (HARD)

- No edit to `.claude/skills/moai/workflows/codemaps.md`.
- No Go source change under `internal/` or `pkg/`.
- The workflow covers extraction only; cohesive authoring + AskUserQuestion stay outside.
- Exactly one verdict (REQ-DCP-004 OR REQ-DCP-005) recorded with evidence.

## E. Self-Verification (plan author)

- [x] SPEC ID passes canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (self-check printed in the authoring turn → PASS).
- [x] 12-field frontmatter present and canonical (id/title/version/status/created/updated/author/priority/phase/module/lifecycle/tags), no snake_case aliases.
- [x] Exclusions section present with 6 entries.
- [x] All 3 binding guardrails mapped to MUST-PASS GEARS requirements (REQ-DCP-002/003/006/009).
- [x] No implementation detail (function names, schemas) leaked into spec.md.
- [x] Tier S — not inflated; design.md/research.md correctly omitted.

## F. Milestones (priority-ordered, no time estimates)

### M1 — Baseline capture (deterministic alternative) — Priority High

- Run `go list -deps -json ./...` and `go doc` on a sampled subset (3-5 packages spanning leaf, mid-fan-in, and high-fan-in profiles).
- Record the deterministic baseline output as the falsification-test control group.
- Output: a baseline sample artifact (transient scratch, cleaned at task end) capturing dependency edges + public surface per sampled package from the deterministic tools alone.

### M2 — Pilot workflow authoring (read-only fan-out) — Priority High

- Author the pilot script logic conceptually: input = package list injected as `args`; one read-only agent per package extracting dependency graph + public surface + any LLM-only architectural observation; aggregate in script variables; return structured per-package graphs.
- Enforce the deterministic-script constraint: no wall-clock, no random in the body.
- Do NOT persist the script to `.claude/workflows/` yet — persistence is deferred to M5 and conditional on the verdict.
- Constrain the fan-out to the SAME sampled subset as M1 for an apples-to-apples comparison (running all 97 is unnecessary for a falsification pilot and wastes tokens).

### M3 — Falsification test execution — Priority High (MUST-PASS gate)

- Run the per-package LLM synthesis on the sampled subset (or, if the workflow primitive is unavailable, simulate the per-package extraction via sequential read-only Explore agents on the same subset, noting the substitution).
- Place the LLM synthesis output side-by-side with the M1 deterministic baseline.
- Apply the value test: does the LLM output contain non-trivial architectural insight (layering, role, fan-in implication, domain boundary) that `go list -deps -json` + `go doc` do not mechanically produce? Distinguish genuine synthesis from restated mechanical facts.

### M4 — Verdict — Priority High (MUST-PASS gate)

- Record exactly one verdict with evidence:
  - **value proven** (REQ-DCP-004): cite the specific insights the LLM added beyond the baseline; recommend shipping the validated pattern → proceed to M5(a).
  - **value not proven** (REQ-DCP-005): cite that the LLM output is reducible to the deterministic baseline; record the negative verdict + how-to learning note; recommend the deterministic `go list -deps -json` path → skip M5(a), still complete M5(b).
- A defensible negative verdict is a PASS, not a failure.

### M5 — Deliverable persistence — Priority Medium

- **M5(b) PRIMARY (always)**: add a validated pattern entry to `.claude/rules/moai/workflow/dynamic-workflows.md` (a § pattern catalog entry for the read-only codemaps-extraction fan-out), capturing the primitive mechanics learned + the verdict reference. This fills the strategy-doc § 6.5 gap and lands regardless of verdict.
- **M5(a) CONDITIONAL (only if M4 = value proven)**: persist the runnable pilot script to local `.claude/workflows/codemaps-extract.js`. Confirm it is NOT under `internal/template/templates/`.

### M6 — Non-template + cleanup verification — Priority High

- `grep -r "codemaps-extract\|codemaps-pilot" internal/template/templates/` → must return nothing (REQ-DCP-009/010).
- Remove transient baseline scratch artifacts (temp-file hygiene).
- Confirm `.claude/skills/moai/workflows/codemaps.md` is byte-unchanged.

## G. Anti-Patterns to Avoid

- **AP-DCP-1**: Running the full 97-package fan-out when a sampled subset suffices for a falsification pilot (token waste, no incremental learning).
- **AP-DCP-2**: Recording "value proven" by counting LLM output that merely restates `go list` edges as mechanical facts (the falsification test must distinguish synthesis from restatement).
- **AP-DCP-3**: Persisting the workflow script before the verdict passes (over-commitment).
- **AP-DCP-4**: Pulling the 5-doc cohesive authoring or the next-steps AskUserQuestion into the workflow (boundary violation).
- **AP-DCP-5**: Adding the pilot script to the template suite (REQ-DCP-009 violation, CI neutrality guard would fail).
- **AP-DCP-6**: Treating a negative verdict as a run-phase failure — it is a PASS state.

## H. Cross-References

- `spec.md` — REQ-DCP-001..010, Exclusions, Design Decision § E.
- `acceptance.md` — Given-When-Then scenarios + MUST-PASS checks.
- `.claude/rules/moai/workflow/dynamic-workflows.md` — primitive + non-template statement.
- `.moai/docs/autonomous-workflow-strategy.md` § 6.5 / § 6.6 — codebase-sweep pattern target for M5(b).
