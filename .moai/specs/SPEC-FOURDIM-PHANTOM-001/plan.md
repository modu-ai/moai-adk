# plan.md — SPEC-FOURDIM-PHANTOM-001

## §A Context

`sync-audit-4dim.js` is the binding sync-phase verdict owner on the happy PASS path (SPEC-AUDIT-SNAPSHOT-001 A3 PROMOTED). The predecessor SPEC-SYNC-AUDIT-FALSIFICATION-001 (merged PR #1344) hardened the agent-side counterpart (IMP-1/3/6) but deferred IMP-5 — the workflow-JS phantom-mechanism guard — to this SPEC.

The hazard: a SPEC declares a defensive / structural mechanism in its plan or acceptance artifacts; the four dimension judges — reading the declaration in the Context payload — score on the assumption that the declared mechanism is real. If the declaration is phantom (the mechanism was never actually written, or was written to a different path), the judges return honest but mis-informed scores, and the harmonic mean can issue a falsified PASS. Because a PASS bypasses the cold `sync-auditor` subagent entirely, nothing downstream catches the phantom.

The fix is a deterministic verdict-phase guard: probe each declared mechanism against the actual write surface; any mechanism with zero on-disk evidence → hard FAIL naming the phantom. The guard composes with the existing null-judge, zero-score, and harmonic-mean verdicts; it does not replace any of them and does not add a 5th dimension.

### Applied lessons

- `feedback_ac_stated_mechanism_can_be_false` — a SPEC's declared mechanism is a claim, not evidence.
- `feedback_lsel_f1_actual_write_surface` — probe MUST hit the diff.patch / produced-files surface, not the declared `target_surface` intent.
- `verification-claim-integrity.md` §1.1 surface 3 + §2 — probe command IS the command; `actual_matches: 0` IS the observed output.

## §B Known Issues

1. **Script body has no FS / shell access.** Per `dynamic-workflows.md` § How a Workflow Runs, the workflow script coordinates agents — it does not run shell commands itself. So the probe CANNOT be a raw JS grep in the verdict phase. See §C Design Tension for the resolution.
2. **Context agent already has a job.** The Context agent currently extracts `spec_id / acceptance_criteria / changed_files / test_command`. Adding probe execution to its responsibilities increases its payload size and runtime cost. Mitigation: probes are SPEC-declared and finite (typically 1–5 per SPEC); the cost is bounded and linear in the declared set.
3. **Probe substring matching is literal.** `expected_match_substring` is a literal string match, not regex. This keeps the verdict deterministic and auditable (REQ-FP-005) but means a mechanism that exists under a slightly different name will register zero matches. The SPEC author owns the probe's faithfulness — the workflow owns the deterministic execution.

## §C Pre-flight (design decisions — DECIDED, do not re-derive)

The four design decisions are resolved. They are recorded here verbatim so run-phase implementation can proceed without re-deriving them, and so plan-auditor can weigh in on the trade-offs.

### Decision 1 — Capture: optional `claimed_mechanisms[]` on `CONTEXT_SCHEMA`

Add an optional field `claimed_mechanisms: [{name, probe_command, expected_match_substring}]` to the existing `CONTEXT_SCHEMA` (NOT a new schema, NOT judge-reported). The Context payload is the single source of truth — it already feeds every Judge, and it will now also carry the declared probes and (after execution) the probe results.

### Decision 2 — Probe: SPEC-declared, executed against the actual write surface

The SPEC under audit DECLARES each `probe_command` + `expected_match_substring` in its plan / acceptance artifacts. This is a deterministic specification authored in the SPEC, not LLM-inferred at audit time. The probe is EXECUTED against the actual write surface — the diff.patch paths plus the produced files in the working tree — NOT the declared `target_surface` intent alone (lessons `feedback_lsel_f1_actual_write_surface` + `feedback_ac_stated_mechanism_can_be_false`). The probe counts `expected_match_substring` occurrences in the combined stdout + file-contents corpus.

### Decision 3 — Verdict precedence: phantom guard sits AFTER zero-score, BEFORE harmonic mean

Total order:
1. null-judge guard → `INCOMPLETE`
2. zero-score guard → `FAIL`
3. **phantom-mechanism guard → `FAIL`** (NEW)
4. harmonic mean → `PASS` / `FAIL`

Phantom sits AFTER zero-score (structural / cheaper) and BEFORE harmonic mean. It is a hard FAIL, never absorbed into a softer verdict. The phantom-BEFORE-zero-score alternative was rejected because zero-score is structurally cheaper (pure JS, no probe execution) AND more fundamental (a 0 dimension makes the harmonic mean mathematically undefined — `1/0` diverges); both orders return FAIL, only the payload array differs (`zero_scored[]` vs `phantom_mechanisms[]`).

### Decision 4 — FAIL payload: per-mechanism baseline attribution

On phantom-FAIL the verdict carries `phantom_mechanisms: [{name, probe_command, expected_matches, actual_matches: 0}]` alongside the existing `findings` / `evidence_gaps` arrays. This satisfies `verification-claim-integrity.md` §2 baseline-integrity attribution — the `probe_command` IS the command that was run, `actual_matches: 0` IS the observed output, named per-mechanism so the FAIL is auditable and falsifiable.

## §D Design Tension (FLAGGED for plan-auditor)

The workflow SCRIPT BODY has no shell / filesystem access (per `dynamic-workflows.md` § How a Workflow Runs — the script coordinates; agents do I/O). So the probe CANNOT be a raw JS grep in the verdict phase.

### Chosen resolution — Context agent executes probe_commands at Context-extraction time

The **Context agent executes the literal `probe_commands` at Context-extraction time** and returns `probe_results[]` in the Context payload. The verdict-phase JS then applies the deterministic `actual_matches == 0 → FAIL` rule (no LLM decides FAIL). This is consistent with the existing architecture — judges are agents that report scores; JS aggregates — and the phantom guard has a TIGHTER trust boundary than the harmonic mean (a literal command + deterministic count vs a subjective 0–10 score).

### Alternative considered — dedicated 6th read-only probe agent

A dedicated 6th read-only probe agent would execute the probes and return its own payload. REJECTED: it adds an agent call, conflates nothing structurally but costs more (one extra `Explore` spawn per audit), and the Context agent already has the audit surface in scope to execute the probes against. Revisit only if Context-agent probe fidelity becomes a measured concern.

### Trust-boundary note for plan-auditor

The phantom-mechanism guard has a TIGHTER trust boundary than the harmonic mean:
- **Harmonic mean** — aggregates four subjective 0–1 scores, each a verifier's judgment with cited evidence but a judgment nonetheless. Trust rests on the judge's skeptical stance + the command+verbatim-output evidence requirement.
- **Phantom guard** — a literal `probe_command` (SPEC-declared, deterministic) + a deterministic count (`actual_matches == 0`). Trust rests on the probe being executed faithfully by the Context agent and the count being a JS integer comparison.

The phantom guard's FAIL is MORE auditable than a low dimension score: a reviewer can re-run the exact `probe_command` against the diff and observe `actual_matches` directly. The harmonic mean's FAIL is harder to falsify (it depends on which finding dragged which score).

The residual trust question for plan-auditor: is delegating probe EXECUTION to the Context agent (an `Explore` agent, read-only, schema-forced) sufficient fidelity, or does the 6th-probe-agent alternative merit a forward-looking debt marker? The plan author judges it sufficient at this Tier (S) on the grounds that the Context agent already has read access to the entire audit surface and the probe is a literal command — but flags this as the single design decision most worth a second opinion.

## §E Self-Verification (run-phase manager-develop)

- Verdict logic is pure JS, no wall-clock / random (deterministic — resume-cache safe).
- Live `.claude/workflows/sync-audit-4dim.js` and mirror `internal/template/templates/.claude/workflows/sync-audit-4dim.js` are byte-identical after edit.
- `make build` regenerated the embedded catalog.
- Template-neutrality §25 guard: distributed JS contains NO SPEC IDs / REQ tokens / IMP-5 / dates / SHAs (grep-verified).
- New tests cover: phantom-FAIL, happy-path passthrough, precedence (null-judge + zero-score still fire first).

## §F Milestones

Order is by decision-reversibility (most likely to change first), NOT chronological. The schema capture field (M1) is the highest-change-likelihood decision because it reshapes the Context payload that every Judge consumes; the tests + mirror (M4) are the most mechanical and sit last.

### M1 — Schema + capture field (DECISION REVERSIBILITY: HIGHEST)

Add optional `claimed_mechanisms: [{name, probe_command, expected_match_substring}]` to `CONTEXT_SCHEMA`. Update the `CONTEXT_PROMPT` to instruct the Context agent to extract the declared probes from the SPEC's plan / acceptance artifacts. No verdict-phase change in this milestone — the field is captured but not yet consumed.

Priority: High. Reversibility: a schema-field addition ripples into every Judge's prompt context; get this shape right before wiring execution.

### M2 — Context-agent probe execution + `probe_results[]` payload (DECISION REVERSIBILITY: HIGH)

Extend the Context agent's responsibilities: execute each declared `probe_command` against the actual write surface (diff.patch paths + produced files in the working tree), count `expected_match_substring` occurrences in the combined stdout + file-contents corpus, and return `probe_results[]: [{name, probe_command, expected_match_substring, actual_matches}]` in the Context payload. Update `CONTEXT_SCHEMA` to include `probe_results`.

Priority: High. Reversibility: the probe-execution contract — what corpus is searched, how matches are counted — is the second-highest-change-likelihood decision. Document the literal-substring semantics (no regex) in the prompt.

### M3 — Verdict-precedence integration (DECISION REVERSIBILITY: MEDIUM)

Wire the phantom-mechanism guard into the verdict pipeline AFTER the zero-score guard and BEFORE the harmonic mean. On any `actual_matches == 0`, return verdict `FAIL` carrying `phantom_mechanisms[]` with per-mechanism `probe_command` + `actual_matches: 0` + `expected_match_substring`. On all-non-zero, fall through to the harmonic mean unchanged. Add a comment block explaining the precedence and the fail-honest semantics.

Priority: High. Reversibility: the precedence order is fixed by REQ-FP-003/004/006; the implementation is mechanical given M1+M2.

### M4 — Tests + template mirror + `make build` + §25 neutrality guard (DECISION REVERSIBILITY: LOWEST — MECHANICAL)

Author tests for: (a) phantom-FAIL with the phantom mechanism named in the payload; (b) happy-path passthrough; (c) precedence (null-judge INCOMPLETE + zero-score FAIL fire before phantom). Mirror the live JS to the template source byte-identically. Run `make build`. Grep the distributed JS for SPEC IDs / REQ tokens / IMP-5 / dates / SHAs — expect zero matches.

Priority: Medium. Reversibility: mechanical; the decisions live in M1–M3.

## §G Anti-Patterns

- **AP-FP-001 — LLM-decided FAIL**: implementing the phantom-FAIL as a 5th Judge that "scores truthfulness" — this would smooth the deterministic FAIL into a subjective score, defeating the guard's purpose. The phantom guard is JS, not an LLM.
- **AP-FP-002 — Declared-intent probe**: probing the `target_surface` declared in the SPEC's plan instead of the diff.patch / produced-files surface — this is the `feedback_lsel_f1_actual_write_surface` failure mode. The probe MUST hit the actual write surface.
- **AP-FP-003 — 5th dimension**: adding "Truthfulness" to the DIMENSIONS enum — this breaks REQ-FP-008 and the FROZEN-4-dimension contract.
- **AP-FP-004 — Absorbed verdict**: downgrading the phantom-FAIL to a warning or folding it into the harmonic mean — this breaks the fail-honest invariant (REQ-FP-004).
- **AP-FP-005 — Template leakage**: writing SPEC-FOURDIM-PHANTOM-001 / REQ-FP-XXX / IMP-5 into the distributed JS comments — this breaks CLAUDE.local.md §25. SPEC artifacts under `.moai/specs/` are dev-only and NOT mirrored; the distributed JS uses the generic name "phantom-mechanism guard".
- **AP-FP-006 — Mirror drift**: editing only the live JS and forgetting the template mirror + `make build` — the next `moai update` overwrites the local copy and the guard is lost.

## §H Cross-References

- spec.md §B (REQ-FP-001..010), §C (AC summary), §E (scope / out-of-scope).
- acceptance.md §D (Given-When-Then AC matrix).
- progress.md §F.1 (plan-done signal).
- Predecessor: `.moai/specs/SPEC-SYNC-AUDIT-FALSIFICATION-001/`.
- Template-First: `CLAUDE.local.md` §2 + §25.
- Dynamic-workflow architecture: `.claude/rules/moai/workflow/dynamic-workflows.md` § How a Workflow Runs + § Purpose-driven model+effort selection (Context agent is `read-only-extract`, `effort: medium`; the phantom guard is JS, not an agent).
