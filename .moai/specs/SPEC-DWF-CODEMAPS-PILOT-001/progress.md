# Progress — SPEC-DWF-CODEMAPS-PILOT-001

> Tier S, exploratory dynamic-workflow pilot. Run-phase executed by orchestrator-direct (Mode 6 workflow launched for M3). Deliverable = falsification verdict + pattern catalog entry (PRIMARY) + validated script (conditional).

## §E.1 — Phase 0.95 Mode Selection

**Input parameters**:
- tier: S (Simple) · scope: pilot (no production Go) · domain count: 2 (workflow script + rule doc) · file language mix: markdown + JS workflow script · concurrency benefit: research-heavy read-only fan-out · Agent Teams prereqs: not evaluated (small scale)

**Mode evaluation**:
- Mode 1 trivial — not selected (multi-step falsification)
- Mode 2 background — not selected (interactive verdict synthesis needed)
- Mode 3 agent-team — not selected (prereqs not met, small scale)
- Mode 4 parallel — viable alternative (Scenario 6 fallback path)
- Mode 5 sub-agent — viable alternative
- **Mode 6 workflow — SELECTED**

**Decision: workflow (Mode 6)**

**Justification**: The SPEC's explicit objective is to LEARN the dynamic-workflow primitive (not merely to reach a verdict). The sampled subset (4 packages) is below Mode 6's ≥30-file mechanical threshold, but primitive-learning is the stated purpose and the scale-vs-doctrine gap is itself falsification evidence. GATE-2 user approval obtained (run-phase entry approved + M3 workflow-launch opt-in explicitly confirmed via AskUserQuestion). All preferences collected before launch (no mid-run user input). Workflow launched: `wf_270163c4-a7d`, 4 read-only Explore agents, 265K tokens, 124s.

## §E.2 — Milestone Trace

| M | Status | Evidence |
|---|--------|----------|
| M1 baseline capture | DONE | `go list -deps -json` + `go doc` + fan-in for 4 sampled pkgs (leaf/mid/hub/entry profiles) |
| M2 pilot workflow authoring | DONE | script authored conceptually (no wall-clock/random; package list via `args`); deferred persistence to M5 |
| M3 falsification execution | DONE | Mode 6 workflow `wf_270163c4-a7d` launched; 4 per-package LLM syntheses returned |
| M4 verdict | DONE | value proven (REQ-DCP-004) with scoping caveats — see §E.3 |
| M5 deliverable persistence | DONE | M5(b) pattern catalog entry in dynamic-workflows.md; M5(a) script persisted to `.claude/workflows/codemaps-extract.js` |
| M6 non-template + cleanup | DONE | grep verified no template ref; codemaps.md byte-unchanged; no transient scratch left |

## §E.3 — M4 Falsification Verdict

**VERDICT: value proven (REQ-DCP-004) — with three honest scoping caveats that reshape the recommendation.**

### Sampled-subset baseline (M1, deterministic control)

| Package | direct internal imports | transitive deps | fan-in | go doc surface |
|---------|-------------------------|-----------------|--------|----------------|
| pkg/version | 0 (stdlib only) | 59 | 3 | `Version` + 4 getters |
| internal/spec | 1 (constitution) | 75 | 3 | 25+ funcs / 15+ types (parse / era / drift / lint / close) |
| internal/config | 2 (defs, models) | 148 | 14 | `Config`/`Loader`/`ConfigManager` + env-key constants |
| cmd/moai | 1 (cli) | 353 | 0 | `ExitCoder` interface only |

The deterministic baseline (`go list -deps -json` + `go doc`) fully and mechanically produces: import edges, transitive closure size, fan-in counts (reverse `go list`), and exported symbol surface. No LLM is required for any of these.

### Reduction test (D1 decision rule, acceptance.md §D.2)

Each LLM claim was tested per-sentence: RESTATEMENT (reducible to a go list edge or go doc symbol → zero value) vs SYNTHESIS (genuine layering/role/fan-in-implication/domain-boundary inference → added value).

| Package | surviving synthesis claims | example surviving claim |
|---------|----------------------------|-------------------------|
| pkg/version | 4 | "version string-pattern `dirty/dev/none` is a hidden dev-build-detection contract duplicated across `deps.go` + `update.go`" — not derivable from the import edge |
| internal/spec | 5 | "`Close()` violates pure-function semantics (query + command) — a critical architectural boundary vs the read-only Audit/Lint/Drift siblings" |
| internal/config | 5 | "`MergeAll` is a policy-enforcement mechanism (8-tier provenance + strict_mode override rejection), not just a data structure" |
| cmd/moai | 4 | "no subagent-boundary protection at the entry point — a negative-space gap (absence) that mechanical tools cannot surface" |

All 4 packages cleared ≥1 surviving claim (averaged ≥4). No AP-DCP-2 restatement-padding detected — the syntheses are genuine inference, not re-emitted edges. **By the D1 reduction rule, this is a value-proven outcome (Scenario 1).**

### Three scoping caveats (the honest part of the verdict)

1. **Augmentation, not extraction.** The surviving value is architecture-REVIEW insight (coupling risk, latent contracts, layering judgments, negative-space gaps). codemaps *extraction* proper — dependency graph + public surface — is mechanically complete from `go list`/`go doc`; the LLM adds zero to extraction. The LLM value sits as an *augmentation layer on top of* the deterministic extraction, not as a replacement for it.

2. **Not primitive-specific.** The synthesis value is NOT specific to the dynamic-workflow fan-out primitive. An identical synthesis is obtainable from a sequential sub-agent or a single Explore agent. The fan-out's only marginal benefit is parallel wall-clock speed, which is material ONLY at high package counts (≈full 97); at the sampled 4-pkg scale the speed gain is negligible and the token cost (265K) is not offset.

3. **High-count justification only.** Combining (1)+(2): the dynamic-workflow primitive earns its overhead for codemaps-augmentation ONLY when (a) architecture insight beyond extraction is actually wanted, AND (b) the package count is high enough (≈97) that parallel fan-out speed offsets the per-agent token cost. For pure extraction, or at small scale, the deterministic `go list`/`go doc` path (optionally + a single sub-agent for insight) is cheaper and sufficient.

### Recommendation

Ship the *pattern* (M5(b) catalog entry) + the *validated script* (M5(a)), but scoped honestly per the three caveats: the script is an **architecture-insight augmentation tool for high-count codemaps**, NOT a replacement for deterministic dep-graph/surface extraction. This aligns with the prior 16-agent adoption analysis (fit 0.52 CONDITIONAL): the primitive has narrow, conditional value here — exactly what the pilot set out to falsify.

## §E.4 — Acceptance Criteria

| AC | Status | Note |
|----|--------|------|
| AC-DCP-001 (extraction-only) | PASS | workflow agents did read-only extraction; no cohesive authoring inside |
| AC-DCP-002 (authoring outside) | PASS | 5-doc authoring + AskUserQuestion stayed in orchestrator; never in workflow |
| AC-DCP-003 (falsification w/ evidence) | PASS | side-by-side baseline vs synthesis on 4-pkg subset, reduction test applied |
| AC-DCP-004 (value proven) | PASS | fired — material value beyond baseline, with scoping caveats |
| AC-DCP-005 (value not proven) | n/a | mutually exclusive with 004; 004 fired |
| AC-DCP-006 (determinism) | PASS | script body: no wall-clock/random; package list via `args`; timestamp post-run |
| AC-DCP-007 (blocker not prompt) | PASS | no missing-input case arose; agents are read-only Explore |
| AC-DCP-008 (GATE-2 passed) — MUST | PASS | GATE-2 approved + M3 launch opt-in confirmed before launch |
| AC-DCP-009 (non-template) | PASS | script in `.claude/workflows/` only; grep clean (M6) |
| AC-DCP-010 (grep verified) | PASS | `grep -r "codemaps-extract\|codemaps-pilot" internal/template/templates/` → nothing |

**MUST-PASS gate**: AC-DCP-002/003/004/006/008/009 all PASS. A value-proven verdict with evidence satisfies the verdict gate. Run is a **PASS**.
