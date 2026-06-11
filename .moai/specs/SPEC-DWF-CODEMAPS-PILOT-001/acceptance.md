# Acceptance Criteria — SPEC-DWF-CODEMAPS-PILOT-001

> A run is a PASS when a defensible verdict is recorded with evidence — value proven OR value not proven. Both are PASS states. The only FAIL is the absence of an evidence-backed verdict, a boundary violation, a determinism violation, or a template-suite landing.

## D. Acceptance Criteria Matrix

| AC | Requirement | Severity | Verifiable check |
|----|-------------|----------|------------------|
| AC-DCP-001 | REQ-DCP-001 | SHOULD | Pilot covers extraction stage only; no cohesive-authoring logic in the workflow. |
| AC-DCP-002 | REQ-DCP-002 | MUST-PASS | 5-doc authoring + Phase 5 AskUserQuestion are outside the workflow. |
| AC-DCP-003 | REQ-DCP-003 | MUST-PASS | Falsification test executed with side-by-side evidence on a sampled subset. |
| AC-DCP-004 | REQ-DCP-004 | MUST-PASS (PASS state) | "value proven" verdict recorded with cited LLM-only insights. |
| AC-DCP-005 | REQ-DCP-005 | MUST-PASS (PASS state) | "value not proven" verdict recorded with how-to note + deterministic recommendation. |
| AC-DCP-006 | REQ-DCP-006 | MUST-PASS | Script body has no wall-clock/random; package list via `args`; timestamp post-run. |
| AC-DCP-007 | REQ-DCP-007 | SHOULD | Missing-input case returns a structured blocker report, not a user prompt. |
| AC-DCP-008 | REQ-DCP-008 | MUST | GATE-2 passed + preferences collected before launch (GATE-2 is a HARD score-independent human gate per project doctrine; REQ-DCP-008 uses "shall"). |
| AC-DCP-009 | REQ-DCP-009 | MUST-PASS | No pilot artifact under `internal/template/templates/`. |
| AC-DCP-010 | REQ-DCP-010 | MUST | Non-template verified by grep returning nothing. |

> AC-DCP-004 and AC-DCP-005 are mutually exclusive — exactly one fires per run. Both are PASS states. The MUST-PASS gate for the verdict is satisfied when EITHER fires with evidence.

## D.1 Given-When-Then Scenarios

### Scenario 1 — Falsification test, value-proven outcome (PASS)

```
GIVEN the deterministic baseline (go list -deps -json + go doc) has been captured for a sampled subset of packages
  AND the pilot read-only fan-out has produced per-package LLM synthesis for the same subset
WHEN the two outputs are placed side-by-side and the value test is applied
  AND the LLM synthesis contains non-trivial architectural insight (layering, role, fan-in
      implication, domain boundary) NOT mechanically derivable from go list/go doc
THEN a "value proven" verdict is recorded (REQ-DCP-004) citing the specific added insights
  AND the pilot recommends shipping the validated pattern
  AND M5(a) persists the runnable script to local .claude/workflows/codemaps-extract.js
  AND the run is a PASS.
```

### Scenario 2 — Falsification test, value-not-proven outcome (PASS)

```
GIVEN the deterministic baseline has been captured for a sampled subset of packages
  AND the pilot read-only fan-out has produced per-package LLM synthesis for the same subset
WHEN the two outputs are placed side-by-side and the value test is applied
  AND the LLM synthesis is reducible to a restatement of the go list/go doc mechanical facts
THEN a "value not proven" negative verdict is recorded (REQ-DCP-005) citing the reducibility
  AND the how-to learning note (primitive mechanics) is retained
  AND the pilot recommends the deterministic go list -deps -json path for this operation
  AND M5(a) is skipped (no script persisted)
  AND the run is a PASS — a defensible negative verdict is success, not failure.
```

### Scenario 3 — Extraction/authoring boundary holds (MUST-PASS)

```
GIVEN the pilot workflow has run and returned structured per-package graphs to the orchestrator
WHEN the workflow contents are inspected
THEN the workflow contains ONLY read-only per-package extraction logic
  AND the 5-doc cohesive authoring (overview/modules/dependencies/entry-points/data-flow) is
      NOT performed inside the workflow
  AND no AskUserQuestion / next-steps interaction occurs inside the workflow
  AND any cohesive authoring or next-steps round happens in the orchestrator turn AFTER the run returns.
```

### Scenario 4 — Determinism constraint (MUST-PASS)

```
GIVEN the pilot workflow script body
WHEN the script body is inspected
THEN it contains no wall-clock read and no random-number draw
  AND the package list is received via the args global input
  AND any timestamp in the final result is stamped by the orchestrator AFTER the run returns,
      never generated inside the script body.
```

### Scenario 5 — Non-template placement (MUST-PASS)

```
GIVEN the pilot has produced its deliverables
WHEN `grep -r "codemaps-extract\|codemaps-pilot" internal/template/templates/` is run
THEN the grep returns nothing
  AND any persisted workflow script lives only under the local user-owned .claude/workflows/ directory
  AND .claude/skills/moai/workflows/codemaps.md is byte-unchanged.
```

### Scenario 6 — Workflow primitive unavailable (graceful fallback)

```
GIVEN the run environment does not support the dynamic-workflow primitive (Claude Code < v2.1.154,
      feature disabled, or unpaid plan)
WHEN the run-phase reaches the falsification test
THEN the per-package extraction is performed via sequential read-only Explore agents on the same
     sampled subset (substitution noted in the verdict)
  AND the falsification verdict is still reached with side-by-side evidence
  AND the primitive-unavailability limitation is documented in the how-to note
  AND the run is a PASS (the verdict, not a live workflow run, is the MUST-PASS deliverable).
```

## D.2 Edge Cases

- **Sampled subset vs full 97**: the falsification pilot uses a representative subset (leaf / mid-fan-in / high-fan-in packages). Running all 97 is NOT required and is an anti-pattern (token waste). Coverage of the subset across fan-in profiles is the quality bar, not raw count.
- **Mechanical restatement masquerading as synthesis**: if LLM output merely re-emits dependency edges that `go list` already prints, that counts as NO added value (drives Scenario 2, not Scenario 1).
- **Partial verdict**: a run that produces extraction output but records NO evidence-backed verdict is a FAIL (the verdict is the deliverable).
- **Synthesis discrimination — mechanical-reduction test (decision rule for AC-DCP-003 / M3)**: apply per-sentence to each LLM synthesis claim. A claim is RESTATEMENT (zero value) when it is mechanically derivable from the deterministic baseline — i.e., it only re-emits a dependency edge already present in `go list -deps -json` `ImportPath` records OR re-lists an exported symbol already present in `go doc`. A claim is SYNTHESIS (added value) only when it asserts something NOT mechanically derivable: a layering/tier judgment, a role/responsibility naming, a fan-in implication, or a domain-boundary inference. Verdict rule: count the synthesis claims that survive the reduction. ≥ 1 surviving synthesis claim per sampled package (averaged across the subset) → value-proven candidate (Scenario 1); 0 surviving synthesis claims → value-not-proven (Scenario 2). The surviving-claim count AND the per-claim reduction (which baseline record each restatement reduces to) MUST appear in the side-by-side evidence so a second reviewer re-derives the same verdict — this is what makes the reproducibility clause (D.4) operative rather than asserted.

## D.3 Definition of Done

- [ ] Falsification test executed with side-by-side deterministic-baseline vs LLM-synthesis evidence on a sampled subset (AC-DCP-003).
- [ ] Exactly one verdict recorded with evidence — value proven OR value not proven (AC-DCP-004 / AC-DCP-005).
- [ ] Extraction/authoring boundary verified: no cohesive authoring, no AskUserQuestion inside the workflow (AC-DCP-002).
- [ ] Determinism verified: no wall-clock/random in body, package list via `args`, post-run timestamp (AC-DCP-006).
- [ ] Non-template placement verified by grep (AC-DCP-009 / AC-DCP-010).
- [ ] PRIMARY deliverable landed: validated pattern entry in `dynamic-workflows.md` § pattern catalog (M5(b)).
- [ ] CONDITIONAL deliverable: local `.claude/workflows/codemaps-extract.js` persisted ONLY if verdict = value proven (M5(a)).
- [ ] `.claude/skills/moai/workflows/codemaps.md` byte-unchanged.
- [ ] Transient baseline scratch artifacts removed (temp-file hygiene).

## D.4 Quality Gate Criteria

- No production Go change (the pilot is workflow-script + docs only) — verifiable via `git diff --stat` showing no `internal/**.go` or `pkg/**.go` modifications.
- No template-suite landing — CI template-neutrality guard remains green.
- The verdict is reproducible from the recorded side-by-side evidence (a reviewer can re-derive the same conclusion from the cited baseline vs synthesis comparison).
