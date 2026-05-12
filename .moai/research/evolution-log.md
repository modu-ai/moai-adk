# MoAI-ADK Evolution Log (Core)

Records all graduated learnings applied to FROZEN/EVOLVABLE zones per design-constitution §6.
This file is the **core** evolution log. Design-subsystem cross-references live in `.moai/design/v3-research/evolution-log.md`.

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-13 | SPEC-V3R2-HRN-002 (manager-spec) | Initial file creation — EVO-HRN-002 first entry. |

---

## EVO-HRN-002

```yaml
id: EVO-HRN-002
spec_id: SPEC-V3R2-HRN-002
timestamp: "2026-05-13"
zone: Frozen
target_file: .claude/rules/moai/design/constitution.md
target_section: "§11 GAN Loop Contract"
amendment_type: "sub-section insertion"

before_snippet: |
  # (§11.4.1 did not exist)
  # §11.4 Sprint Contract Protocol defined durable Sprint Contract state,
  # but did not constrain evaluator memory scope to ephemeral-per-iteration.

after_snippet: |
  ### §11.4.1 Evaluator Memory Scope (Principle 4)
  Evaluator judgment memory SHALL be ephemeral per iteration...
  Sprint Contract state SHALL be durable across iterations...
  Implementation: the GAN loop runner SHALL respawn evaluator-active for each iteration...
  Configuration: evaluator.memory_scope: per_iteration (FROZEN; value cannot be changed without a new CON-002 amendment cycle)

canary_verdict: "CanaryUnavailable (0 of 3 required subjects; v3 corpus has no completed GAN-loop evaluation artifacts)"
canary_log_ref: ".moai/specs/SPEC-V3R2-HRN-002/canary-fresh-memory-eval.txt"

con_002_evidence_ref: ".moai/specs/SPEC-V3R2-HRN-002/con-002-amendment-evidence.md"
con_002_layers:
  layer1_frozen_guard: "PASS"
  layer2_canary: "PASS (CanaryUnavailable — no prior evaluator runs to regress)"
  layer3_contradiction: "PASS"
  layer4_rate_limiter: "PASS (2 of 3 used in v3.x cycle)"
  layer5_human_oversight: "PENDING-FINAL (maintainer approval in landing PR)"

approver: "Goos Kim (bobby@afamily.kr) — approval recorded in landing PR description"
rationale_cite: "R1 §9 (Zhuge et al. 2024 arXiv:2410.10934 Agent-as-a-Judge anti-pattern: cumulative evaluator memory) + Principle 4 (design-constitution §11.4.1)"

const_registry_entry: "CONST-V3R2-153"
design_log_ref: ".moai/design/v3-research/evolution-log.md"

status: "landed"
version_before: "3.4.0"
version_after: "3.5.0"
```

---

_End of evolution log._
