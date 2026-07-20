# SPEC-CLIFIX-HYGIENE-001 — Implementation Plan

## §A Context

P4 row of the CLI audit roadmap: reduce maintenance cost and recurrence risk. Tier L — widest file surface of the five CLIFIX SPECs, but lowest urgency; strictly last in the series so it rebases on all prior fixes. Methodology: DDD (ANALYZE-PRESERVE-IMPROVE) for the decomposition and dead-code milestones; Reproduction-First TDD for the genuine behavior defects (merge key drop, perms, PAT echo, YAML misparse, rune truncation).

## §B Known Issues (findings inventory)

| # | Anchor (re-verify) | Defect | Fix direction |
|---|---|---|---|
| 1 | update.go (3,181 LOC) | single-file structure debt; analysis run twice (920-963); global verbose flag | split into update_sync/merge/wizard/settings seams |
| 2 | update_cleanup.go (~300 LOC post-P0), update.go:618-717, 1596-1624, 1069-1073, 730-737; buildGLMEnvVars; ttyConfirmer; worktree_validation/init_layout | ~500 lines dead code | delete after caller-graph verification |
| 3 | glm.go:194-209 + 6 functions; glm_tools.go:368-377; constitution.go:21,149; update.go:459; CI=="1" checks | env names inlined; GLM inject/clear drift | envkeys.go constants + shared glmEnvVarSet() |
| 4 | tier thresholds [1,3,5,10] ×2; hook.go:214,345 30s×2; 10MB/30s/120s/5min/circuit-3 | threshold literals duplicated | defaults.go single source |
| 5 | doctor.go, migration.go, clean.go, web_port* | Korean UI strings vs error_messages:en | English literals |
| 6 | constitution, tool_policy, github.go (+1 audited site) | byte-slice truncation breaks CJK | one rune-boundary truncate helper |
| 7 | update.go:2396-2452 | deepMerge3Way drops user-only keys | preserve old-only keys |
| 8 | update.go:2641-2651 | tokens written 0644 | 0600 |
| 9 | wizard/questions.go:119-154; wizard/wizard.go:99,126 | PAT plaintext echo; stepper total hardcoded 6 vs 7-9 shown | EchoModePassword; dynamic total |
| 10 | worktree/tmux_integration.go:114,124,160; worktree/new.go:481 | strings.Contains YAML parse (comments match); unquoted tmux paths | yaml.v3 parse; %q quoting |

## §C Pre-flight

1. Confirm P0-P3 merged; refresh every anchor (this SPEC's anchors drift the most — update.go will already differ from the audit tree after P0 lock wiring).
2. ANALYZE: caller graph for each dead-code symbol (`go vet`, `grep -rn`, unused linter) — anything with a live caller leaves the inventory.
3. PRESERVE: characterization tests over update flows (dry-run plan, merge behaviors, settings sync) before any split; capture goldens.
4. Inventory Hangul-bearing strings: `grep -rlP '[\x{AC00}-\x{D7A3}]' internal/cli --include='*.go'` (rg fallback: `rg -l '[가-힣]' internal/cli -g '*.go'`) — expected set = the six audited files; diff against actual before editing.

## §D Constraints

- Behavior preservation is the default; the ten behavior-correcting fixes each need a RED reproduction test first.
- Decomposition commits are mechanical-move-only (no logic edits in the same commit) to keep review and git blame tractable.
- Dead-code deletion must not remove: the P0-wired lock path, any symbol registered as a hook/live wrapper (cf. HOOK-DEADCODE-001 lesson: wrappers may be live via agent registration), or exported symbols with external callers.
- §15 template-language-neutrality untouched: this SPEC edits Go sources only, no template files.

## §E Self-Verification

- E1: AC matrix PASS/FAIL against acceptance.md (10 ACs).
- E2: `go build ./... && go test ./internal/cli/... ./internal/cli/wizard/... ./internal/cli/worktree/... -count=1` verbatim.
- E3: coverage of update cluster ≥ baseline (characterization suite counted).
- E4: dead-code grep audit — each deleted symbol returns 0 matches.
- E5: `golangci-lint run` no new findings; `wc -l` report for the update cluster files.

## §F Milestones (priority order)

- M1 — PRESERVE net: characterization tests + goldens for update flows; Hangul/threshold/env-literal inventories frozen as fixtures.
- M2 — Correctness fixes (repro-first): deepMerge3Way old-key preservation; token 0600; rune-truncate helper + 4 sites; worktree yaml.v3 + tmux quoting; wizard PAT masking + dynamic stepper.
- M3 — Hardcoding sweep: envkeys.go constants + glmEnvVarSet(); defaults.go threshold extraction; CI env check correction fold-in where audited.
- M4 — English UI strings: six files converted; goldens updated deliberately.
- M5 — Structure: update.go decomposition (mechanical moves); dead-code deletion per verified inventory; final full-suite + lint + §E self-verification.

## §G Anti-Patterns and Risks

- Execution order: P0→P1→P2→P3→P4 — this SPEC is last; starting it earlier guarantees merge conflicts on update.go/glm.go/launcher.go/hook.go with P0-P2 work.
- Shared-file overlap (rebase surface): update.go+update_cleanup.go (CRITICAL-001 e), glm.go/launcher.go/glm_tools.go (CRITICAL-001 a, CONCURRENCY-001 1-3), hook.go (CRITICAL-001 h, CONTRACT-001), doctor_skills-adjacent doctor.go (LINTER-STALE-001 4).
- Anti-pattern: mixing mechanical moves with logic fixes in one commit — forbidden by §D.
- Anti-pattern: deleting "dead" code by grep absence alone — reflection/registration/tag-gated callers must be checked (DDD ANALYZE names every caller).
- Anti-pattern: translating Korean strings by blind sed — each message reviewed so diagnostics stay accurate (mirrors docs-site no-blind-sed rule).
- Risk: characterization goldens over-fitting current bugs — where a golden captures a defect fixed in M2, regenerate the golden in the fixing commit with rationale.
- Risk: stepper dynamic total changes rendered output consumed by wizard tests/screenshots — update fixtures deliberately.

## §H Cross-References

- Findings SSOT: `.moai/reports/cli-improvement-audit-20260710.html` §3 clusters 1/2/4/5, §4 rows 6-9, §5 P4.
- Depends on: all four prior CLIFIX SPECs (P0-P3).
- CLAUDE.local.md §14 (hardcoding policy), §3/§6 (Go standards, test isolation), language.yaml `error_messages: en`.
- moai-workflow-ddd skill (ANALYZE-PRESERVE-IMPROVE governs M1/M5).
