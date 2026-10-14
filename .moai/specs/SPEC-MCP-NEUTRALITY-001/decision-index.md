# SPEC-MCP-NEUTRALITY-001 — Decision Index

Stateless on the status axis: the SPEC lifecycle lives in `spec.md` alone. One row per decision surfaced during assembly that the operator has not settled. Labels: DECIDED / POLICY-COVERED / EVIDENCE-NEEDED / FOUNDER.

### Q1: What ceilings should the GLM audit path use (currently 120 s / 4,096 tokens)?

Label: EVIDENCE-NEEDED

Class: implementation-level

Authority anchor: none — the current values live in `internal/cli/mcp_glm.go`; no committed artifact settles replacement values, and the truncation-rate measurement that would justify new ones does not exist yet.

Why unresolved: lens-3 #22 flags the ceilings as truncation suspects, but no measured truncation rate or latency distribution exists in any committed artifact; picking new numbers now would be an unmeasured claim.

Default: keep the current values as configurable defaults (rule: preserves current behavior) and add the configuration seam so measured values can land without a code change.

Alternate: raise the ceilings immediately to unmeasured headroom values.

Operator verdict: DEFAULT-APPLIED 2026-10-14T01:39:31Z manager-spec (implementation-level row; the run phase keeps current values as defaults plus the configuration seam, and records the truncation-rate measurement as evidence for any future change).

### Q2: What constitutes the positive operator mark for operator-only verbs?

Label: FOUNDER

Class: implementation-level

Authority anchor: none — the card mandates "a positive operator mark" but no committed artifact names the mark; `.moai/reports/` design docs are untracked and cannot anchor.

Why unresolved: the mark's vocabulary is a contract every Codex role file and launcher stamps must agree on; guessing a new vocabulary risks breaking the existing launcher stamps.

Default: extend the existing launcher-stamped `MOAI_FACTORY_ROLE` vocabulary (an `operator` value stamped where the launcher already stamps lane roles) rather than introducing a separate marker variable (rule: smaller user-visible surface — no new env vocabulary).

Alternate: a dedicated `MOAI_OPERATOR=1` marker independent of the role vocabulary.

Operator verdict: DEFAULT-APPLIED 2026-10-14T01:39:31Z manager-spec (implementation-level row; the run phase extends the existing role vocabulary per the default).
