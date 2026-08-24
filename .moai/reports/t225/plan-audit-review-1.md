# Plan-Audit Review 1 — SPEC-V3R6-AUDIT-MODEL-PIN-001 (card t225)

- Auditor: t225-planaudit (plan-auditor, opus/high), adversarial read-only
- Target: `.moai/specs/SPEC-V3R6-AUDIT-MODEL-PIN-001/` @ commit e95305563 (branch WT-audit-model-pin)
- Verdict: **FAIL** (iteration 1/2, Tier M ceiling 2)
- Score: 0.875 — Clarity .75 / Completeness 1.0 / Testability .75 / Traceability 1.0
- Milestone firewall MP-1..MP-7: all PASS (FAIL is defect-driven, not firewall-driven)

## Blocking findings

| # | Severity | Finding |
|---|----------|---------|
| MF1 | critical | plan.md M5 mandates committing `.moai/config/sections/llm.yaml`, but `.gitignore:192` ignores that exact path — commit impossible as written. Additionally `moai update` cleans `.moai/config` wholesale (deploy.go:187) and re-deploys template llm.yaml (empty audit) — the pin is non-durable and contradicts spec.md §B's goal. |
| MF2 | major | plan M2 locus `codexSSOTModelEffort` is shared: codex_task.go:226 → openCodexSessionOn → resolveCodexModelEffort (mcp_codex.go:565, also :840) — the pin leaks into codex_task. REQ-AMP-008 isolation exists only for the GLM side; codex_task must be scoped explicitly. |
| MF3 | major | spec.md §E + plan M1 + AC-AMP-001 claim CONFIG_STRUCT_YAML_MISMATCH guards llm.audit — false: `audit_struct_yaml_symmetry_test.go` symmetryCases ("4 MIG-003 sections plus StatuslineConfig", 7 cases) excludes LLMConfig; the named `-run` is vacuous for this SPEC. |
| MF4 | major | `audit.glm.effort=high` is dual-interpretation: GLM state name (pass-through) vs Claude vocab (CollapseClaudeEffortToGLM → max, glm_effort_overlay.go:129) — plan M3 must state precedence. |
| MF5 | minor | AC-AMP-006 "materially higher" lacks a numeric decision threshold. |
| MF6 | minor | Declare the no-credential path for M5 live gates (skip semantics when GLM_API_KEY absent). |

All other anchors verified TRUE, including `~/.codex/config.toml:2-3` exact content and template baseline 0 hits for `gpt-5.6-sol`.

## Lane-side verification of auditor claims (2026-08-24, this tree)

- MF1: `git check-ignore -v .moai/config/sections/llm.yaml` → `.gitignore:192` match, rc=0. Sibling section files ARE tracked (archive.yaml, cache.yaml, …) — llm.yaml is the sole per-machine section by design (gitignore comment: runtime-expanded by `moai glm/cg`). **CONFIRMED**.
- MF2: `resolveCodexModelEffort` call sites mcp_codex.go:565 (thread/start) + :840 (turn/start); both sit downstream of `codexSSOTModelEffort`. **CONFIRMED**.
- MF3: `symmetryCases` comment "the 4 MIG-003 sections plus StatuslineConfig"; zero LLMConfig occurrences in the file. **CONFIRMED**.

## Disposition

- MF1 escalated to the lead (card mandate "llm.yaml 값 설정" vs repo mechanics — genuine either/or; see lead message). Revision blocked on that ruling.
- MF2–MF6: revision brief to manager-spec.

---
Generated from the auditor's delivered verdict message (its session directive barred report-file writes; content preserved verbatim-in-substance by the lane).
