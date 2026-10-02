---
id: SPEC-CTX-TABLE1M-001
title: "Decision index — context-window 1M-default correction"
created: 2026-10-02
---

# Decision index — SPEC-CTX-TABLE1M-001

> Stateless by design (no `status:` field — the SPEC's lifecycle lives in
> spec.md alone). Required by `interview.decision_gate: on`
> (`.moai/config/sections/interview.yaml:6`). One row per decision surfaced
> during clarification or assembly that the operator has not settled. The two
> decisions settled at repair iteration 1 (conforming SPEC ID; disposition of
> the pre-rename directory) are operator-settled and carry no row — their
> record lives in spec.md HISTORY.

### Q1: Does an explicit `[1m]` suffix override `CLAUDE_CODE_DISABLE_1M_CONTEXT=1` where both are present?

Label: EVIDENCE-NEEDED
Authority anchor: upstream Claude Code CHANGELOG / env-vars documentation (not yet consulted on this precedence question)
Why unresolved: the two cited changelog bullets (2.1.285, 2.1.287) establish that the flag keeps 200K and that 1M is now default on the named surfaces, but not the precedence when the flag and an explicit suffix combine; plan P1's NEW text preserves the pre-edit line's claim without strengthening it (audit D5; acceptance.md Residual risks).
Operator verdict:

### Q2: Does explicit `sonnet[1m]` selection still function on a custom `ANTHROPIC_BASE_URL` gateway after CC 2.1.285, or is it merely unnecessary?

Label: EVIDENCE-NEEDED
Authority anchor: upstream Claude Code CHANGELOG (2.1.285 bullet) / model-config documentation
Why unresolved: 2.1.285 makes gateways use the native 1M window automatically for qualifying models, so the suffix is no longer needed there; whether explicit selection still resolves, and to what, is unmeasured. P1's wording ("selects the 1M window where the model exposes the suffix") deliberately stops short of claiming either way.
Operator verdict:
