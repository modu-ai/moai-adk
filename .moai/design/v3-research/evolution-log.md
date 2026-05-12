# Design Subsystem Evolution Log

Cross-reference of evolution events that affect the design pipeline (GAN loop, evaluator, Sprint Contract).
The **core** evolution log lives at `.moai/research/evolution-log.md`.
This file adds design-domain context (before/after snippets at the design config layer).

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-13 | SPEC-V3R2-HRN-002 (manager-spec) | Initial file creation — EVO-HRN-002 first entry. |

---

## EVO-HRN-002

```yaml
id: EVO-HRN-002
core_log_ref: ".moai/research/evolution-log.md#evo-hrn-002"
spec_id: SPEC-V3R2-HRN-002
timestamp: "2026-05-13"
zone: Frozen
domain: "GAN loop / evaluator memory"

before_snippet: |
  # design.yaml (before)
  design:
    gan_loop:
      max_iterations: 5
      # (no evaluator.memory_scope key)

after_snippet: |
  # design.yaml (after)
  design:
    evaluator:
      memory_scope: per_iteration  # FROZEN per §11.4.1
    gan_loop:
      max_iterations: 5

  # harness.yaml (after)
  harness:
    evaluator:
      memory_scope: per_iteration  # FROZEN per §11.4.1

canary_verdict: "CanaryUnavailable (0 of 3 required subjects; GAN loop not yet exercised in production)"
con_002_evidence_ref: ".moai/specs/SPEC-V3R2-HRN-002/con-002-amendment-evidence.md"

rationale_cite: "R1 §9 (Zhuge et al. 2024 arXiv:2410.10934) + Principle 4 (design-constitution §11.4.1)"
approver: "Goos Kim (bobby@afamily.kr) — approval recorded in landing PR description"

go_enforcement:
  - "internal/config/errors.go: ErrEvalMemoryFrozen sentinel"
  - "internal/config/loader.go: LoadHarnessConfig() rejects non-per_iteration values"
  - "internal/harness/evaluator_leak.go: DetectPriorJudgmentLeak() blocks prior-judgment context injection"

status: "landed"
version_before: "design-constitution 3.4.0"
version_after: "design-constitution 3.5.0"
```

---

_End of design evolution log._
