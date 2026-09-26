# Progress — SPEC-AGENT-MODEL-INHERIT-001

## §E.1 Plan-phase Audit-Ready Signal

- Card: t1246. Branch `WT-agent-model-inherit`, base develop `d6992e3a0`. Tier L.
- Artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, progress.md.
- Inventory measured 2026-09-26 (research.md). Budget baseline: 77539 / 77600 tokens, headroom 61.
- Run gate: card t1175 merged to develop (REQ-AMI-001); re-measure t1257 overlap at run entry (REQ-AMI-002).
- Open questions for the operator (answers to be recorded here before M1):
  - Q1 — web console extent: remove the whole agent-settings tab, or keep a read-only agent list without model/effort?
  - Q2 — leftover `llm.yaml` keys under `moai update`: strip with a report line (plan default) or leave in place and ignore?
  - Q3 — also remove `workflow_agents`, `model_routing`, `model_routing_profiles`, `performance_tier`, and the dynamic-workflow `agent()` model/effort literals (plan default: yes)?
  - Q4 — `init --profile` / `update --profile`: remove with an error, or keep as a deprecated no-op with a warning?
  - Q5 — harness v4 manifests: make specialist `model`/`effort` optional and ignored, and stop `/moai:harness` generating them?
  - Q6 — docs-site (48 pages, four locales): update in this SPEC (M8) or file a follow-up card?

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
