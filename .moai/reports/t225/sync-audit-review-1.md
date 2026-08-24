# Sync-Audit Review 1 — SPEC-V3R6-AUDIT-MODEL-PIN-001 (card t225)

- Auditor: t225-syncaudit (sync-auditor, opus/high), read-only, tree @ 638737651 (base f84904cd9)
- Verdict: **FAIL** — Functionality must-pass firewall (harmonic mean ≈ 71.2)
- Scores: Functionality 55 (FAIL) / Security 92 (PASS) / Craft 65 (FAIL) / Consistency 85 (PASS)
- AC 8/8 individually verified green (pin: 16 tests re-run `ok 2.435s`; live evidence figures all match). FAIL is from 2 deterministic regressions in the unfiltered package suite.

## MUST-FIX (blocking)

### F1 — first-TEXT-block parser regression
`internal/cli/mcp_glm.go:313-321` gates the scan on `Type == "text"`, skipping text blocks with no type field. `internal/cli/mcp_glm_parse_test.go:31-47` (production-envelope fixture) fails deterministically: `z.ai response carried no text content`. Base f84904cd9 read `Content[0].Text` unconditionally and passed — regression introduced by this branch (M5 parseGLMReview fix). REQ-AMP-004 violation. Fix: accept `Type == "" || Type == "text"`.
Lane reproduction (2026-08-24): `go test ./internal/cli/ -run 'TestParseGLMReview_StripsMarkdownFence|TestRunInit_WizardAuditSelectionPersists' -count=1` → FAIL 1.190s, identical message.

### F2 — template audit block breaks init wizard persistence
New template `workflow.yaml:65` audit block (pin sub-keys only, no model/gates leaves) × `internal/core/project/initializer_audit.go:179-204`: `patchAuditLeaves` silently drops values when the leaf is absent (`ok=false`, no insertion fallback). Base had no template block → `insertAuditBlockUnderWorkflow` wrote the whole block; now `workflowHasAuditBlock`=true → patch path → wizard's `workflow.audit.model` never lands. `internal/cli/init_audit_wiring_test.go` deterministic FAIL. User-facing: new `moai init` audit selection silently lost (SPEC-INIT-WIZARD-REPAIR-001 AC-009 broken). Fix: insertion fallback in patchAuditLeaves (preferred) or template leaf skeleton.

## Other findings

- F3 [OBS] The "unreproduced flake" (7.5s) is SEPARATE from F1/F2 — the sweep regex didn't match either failing test name even as substrings (that's why the sweep was green). Flake itself green on 6 re-runs (lane 4 + auditor 2); unexplained residual.
- F4 [NICE] `mcp_convergence.go:394-400` comment claims legacy `runCodexReviewRPC` reuse; actual :425 seam is pin-aware `runCodexAuditReviewRPC` — reads opposite to this SPEC's audit-vs-legacy distinction.
- F5 [OBS] codex effort has no vocabulary validation (asymmetric vs GLM's `glmAuditReasoningEffort`) — not an AC violation (REQ-AMP-006 governs GLM only); hand-edited invalid codex effort transmits verbatim.
- F6 [OBS] Coverage: config 80.6% (below 85% package bar; incremental code covered), settings 90.2%; internal/cli instrumented run exceeded 600s default timeout (602.07s — matches the documented 600s floor behavior).
- F7 [OBS] progress.md duplicate `## §E.3` placeholder heading (self-disclosed in §E.4 gaps; era-classification harmless). iter-2 plan-audit PASS report file not persisted (progress.md record only; review-1 is tracked).

## Trap-check results (all 4)

1. Isolation PASS — structural: `codexReviewRPC` production callers exactly 2 (both audit); codex_task → nil resolver → legacy; review gate uses `runCodexReviewRPC` directly. GLM: 2 audit callers; glm_task → `resolveGLMModelForAgent` directly. Negative-arm test drives the real codex_task handler and POSITIVELY asserts the SSOT pair (gpt-5-codex/medium).
2. Web closed set PASS — UI via SSOT accessor `template.GLMReasoningStateNames()` = {max,high,low}; server-side enforcement at request builder `glmAuditReasoningEffort` (invalid → directive dropped). Hand-edited yaml neutralized at the builder.
3. Evidence preservation PASS — `.moai/state/verify/t225/` 5 files present; load-bearing figures preserved in tracked progress.md + acceptance.md and ALL MATCH (2456/1662=1.48≥1.25; null 11533/11288=1.02<1.1; 513/367=1.40; AC-AMP-007 wire NDJSON shows `"model":"gpt-5.6-sol"` thread/start + turn/start, `"effort":"high"` turn/start). Amendment is evidence-backed.
4. Flake — see F3; the real find was 2 deterministic failures the filtered sweep could not see.

## Companion patch (ee600ccbd) — out-of-SPEC, sanity-checked

Self-consistent: tests rewritten to assert RETIRED-marker absence, catalog hash updated, 26.3s green in this tree. No issue.

## Disposition

F1+F2+F4 fix delegation to manager-develop (2026-08-24); required new §E evidence: UNFILTERED `go test ./internal/cli/ -count=1 -timeout 900s` + unfiltered config/web/core runs. Re-audit scoped to the fix delta after.

---
Preserved from the auditor's delivered report by the lane (auditor session barred from report-file writes). Raw suite output: /tmp/t225-cli-full.txt lines 354, 1522.
