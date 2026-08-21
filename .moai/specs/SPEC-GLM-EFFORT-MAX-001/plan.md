# SPEC-GLM-EFFORT-MAX-001 — Implementation Plan (Tier S)

## §A Context

- Card t175. Worktree `.claude/worktrees/t175`, branch `WT-glm-effort-max`. Plan-phase artifacts only — no git operations in this delegation.
- SPEC: `.moai/specs/SPEC-GLM-EFFORT-MAX-001/spec.md` (6 REQ, 8 inline AC, Tier S ceilings respected).
- Ground truth: `.moai/reports/t175/measurements.md` (committed) + direct code reads listed in spec.md §1.1.
- Affected files (run-phase diff set, 8 files: 1 behavior file + 4 test flips + 2 comment-only + 1 template doc):
  1. `internal/template/glm_effort_overlay.go` — the ONLY behavior change: collapse branch (`:121-122` medium/high→`glmReasoningMax`), session default (`:197-199` → `glmReasoningMax`), dead-var removal (`:95`), comments (`:99-116` collapse doc, `:182-199` cost policy on the ratified grounds, `:194-196` UNVERIFIED→measured).
  2. `internal/template/glm_effort_overlay_test.go` — RED flips + AC-GET-003 re-anchor.
  3. `internal/cli/glm_reasoning_overlay_test.go` — session-default env flip (`:23-25`, literal `"high"` → `template.GLMReasoningEffortMax`).
  4. `internal/cli/glm_test.go` — `TestGLMReasoningEnvVarsForEffort` rows `:511-512,:515`.
  5. `internal/web/agentfm_glm_reasoning_test.go` — manager-spec chip `:98` + comment `:77-78`.
  6. `internal/cli/glm.go` — comment-only (UNVERIFIED→measured, REQ-GEM-005); NO delivery-site value change — both env writers inherit from the overlay (spec.md C-1).
  7. `internal/template/profile_matrix.go` — comment-only (`:219-227` grouping statement update, REQ-GEM-005 / audit D3).
  8. `internal/template/templates/.moai/config/sections/llm.yaml` — collapse doc block `:16-23` + `make build` (§D-2 ratified).
- NOT touched: `internal/web/glm_tier_test.go`, `internal/settings/schema_sections.go` (preserved stored-only surface; its post-change rationale-cross-reference staleness is recorded in spec.md §5, not edited — B-7), `buildTmuxClearVars`, any local `llm.yaml`.
- Route: repo-local Route B (PR) for all tiers — no direct `main` push (`.claude/rules/local/repo-local-pr-policy.md`).
- Cycle: TDD (`quality.yaml` `constitution.development_mode: tdd`) — RED (flip tests) → GREEN (overlay change) → doc surfaces → verify.

## §B Known Issues

- **B-1 Cross-SPEC supersession (RESOLVED — lead-ruled 2026-08-22).** REQ-GEM-002 supersedes REQ-GER-004 of SPEC-GLM-EFFORT-REBALANCE-001, which the lead arbitrates as a stalled draft (v0.1.0; its v3.1.0 target already shipped without it; M1 763582247 unpushed; inactive since 2026-08-15) restating landed floor behavior, not living doctrine. The supersession sentence is recorded in spec.md §1.3; REBALANCE's retirement is Out of Scope (separate lead query at batch time). M1 is unblocked.
- **B-2 AC-GET-003 test re-anchor.** At Claude effort `high`, `ResolveGLMReasoning("builder-harness", …)` and the coding-max override now both yield `max` — the old discrimination ("not overridden") is unobservable at that input. Re-anchor the make-or-break assertion to `EffortLevelLow`: `builder-harness@low → low` (collapse) vs `manager-develop@low → max` (override). This narrows SPEC-GLM-EFFORT-TUNE-001 AC-GET-003's *witness*, not its guarantee (builder-harness is still not in the override set — `GLMCodingMaxOverrideAgents()` membership assertions are untouched).
- **B-3 Pre-existing template doc staleness.** `templates/.moai/config/sections/llm.yaml:16-23` still says "effort low → thinking off" — stale since the glm-5.3 floor rework (code moved to `low`, thinking enabled). REQ-GEM-005 fixes that line in the same hunk as the medium/high line. Mirror-parity discipline: edit template source, `make build`, mirror checks in the same commit.
- **B-4 Magic literal.** `glm_reasoning_overlay_test.go:23` asserts the raw string `"high"` (a §14 smell); the flip uses `template.GLMReasoningEffortMax`, consistent with `glm_test.go`'s constant usage.
- **B-5 Worktree-invisible config.** The local `llm.yaml` (primary checkout, gitignored) cannot be read from this worktree — all claims about its `glm.effort` block are attributed to the measurement record, not to direct observation. `llm.profiles` shadowing (REBALANCE Amendment 1) does not apply: the collapse and the hardcoded session default are not profile-derived (spec.md C-3).
- **B-6 Session env is frozen pre-change.** This live session already carries `ANTHROPIC_REASONING_EFFORT=high`; post-change verification is binary-output + fresh-session (AC-GEM-008), not this session's env.
- **B-7 Preserved-file staleness (recorded, not fixed — audit D6).** `internal/settings/schema_sections.go:175-180` grounds the stored tier-effort defaults on the same rationale as `SessionGLMReasoningState` (a rationale cross-reference). Post-change that cross-reference describes a superseded rationale. The file is a preserved stored-only surface (AC-GEM-006) and spec.md §5 records the staleness — run-phase must NOT "fix" it.

## §C Pre-flight

```bash
git branch --show-current && git rev-parse --short HEAD   # WT-glm-effort-max @ base
go test -run 'TestCollapseClaudeEffortToGLM|TestSessionGLMReasoningState|TestResolveGLMReasoning|TestGLMReasoningEnvVars' ./internal/template/ ./internal/cli/   # baseline GREEN (current behavior)
go build ./... && GOOS=windows GOARCH=amd64 go build ./...  # baseline build clean
golangci-lint run --timeout=2m ./internal/template/... ./internal/cli/... ./internal/web/... 2>&1 | tail -5  # lint baseline (NEW-vs-baseline classification)
```

## §D Constraints & Open Decisions (resolve before Implementation Kickoff Approval)

### §D-1 [RESOLVED 2026-08-22, lead-ratified] Session default = max; REQ-GER-004 supersession recorded

**Decision: candidate (a) — raise `SessionGLMReasoningState()` to `max`. APPROVED.**

Ratified grounds (the run-phase cost-policy comment rewrite carries exactly these, per REQ-GEM-005):

1. Under Branch-B delivery the session-global env var is the ONLY reasoning channel every spawn pays — a `high` default while the collapse maps high→max would reach only the prefs-driven path and silently withhold the operator's order from the default session.
2. t127 measured trivial-spawn cost ≈ 0 subagent tokens — the "paid by every spawn" premise of the old `high` floor is not where cost lives; it scales on large calls' reasoning tokens, which is where the operator wants `max`.
3. `max` is z.ai's own omit-default — the session default stops fighting the backend's native default.

Lead arbitration on the REBALANCE conflict: SPEC-GLM-EFFORT-REBALANCE-001 is a stalled draft (v0.1.0, v3.1.0 target already shipped without it, M1 763582247 unpushed, inactive since 2026-08-15); REQ-GER-004 restates landed floor behavior, not living doctrine. Recorded in spec.md §1.3: "REQ-GER-004 (SPEC-GLM-EFFORT-REBALANCE-001, stalled draft) is superseded by this SPEC." REBALANCE's retirement stays out of scope — its other REQs are unrelated; the stalled draft's disposition goes to a separate lead query at batch time.

### §D-2 [RESOLVED 2026-08-22, lead-ratified] Template-mirror doc-block hunk IN scope

The mirror hunk (`internal/template/templates/.moai/config/sections/llm.yaml:16-23`) + `make build` is APPROVED as in-scope (REQ-GEM-005, AC-GEM-007) — the delegation's "llm.yaml" exclusion is confirmed to mean the LOCAL config instance, which stays untouched (spec.md C-4, §5). One hunk, template source first, `make build` + mirror checks in the same commit.

### §D-3 Binding constraints

- PRESERVE (no edit): `internal/web/glm_tier_test.go`, `internal/settings/schema_sections.go` (stale rationale cross-reference recorded, B-7), `internal/cli/glm.go` code — its edit is comment-only (REQ-GEM-005), no delivery-site value change — `buildTmuxClearVars`, any local `llm.yaml`. (`profile_matrix.go` was moved OUT of this list by audit D3: its `:219-227` grouping comment goes stale post-change and is updated comment-only per REQ-GEM-005.)
- No magic literals (§14): use `glmReasoningMax` / `GLMReasoningEffortMax`, never a fresh `"max"` string in code.
- Affected-packages-only local tests (C-7); full-matrix verdict is CI's.
- Conventional commits `feat(SPEC-GLM-EFFORT-MAX-001): M<N> …`; Route B PR (C-6); never `--no-verify`.

## §E Self-Verification (run-phase §E evidence, attributable per VCI §2)

- **E1** AC matrix (AC-GEM-001..008) — command + verbatim output per row; AC-GEM-008 attributed to lead/operator with the observed command.
- **E2** `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` — exit 0 both.
- **E3** Coverage: `go test -cover ./internal/template/...` (affected; 85% target).
- **E4** Boundary grep n/a (no hook/harness code touched) — record the n/a with reason, do not omit silently.
- **E5** `golangci-lint run` on the three affected packages — NEW-vs-baseline split; must include the no-unused-`glmReasoningHigh` observation (AC-GEM-003).
- **E6** Commit list + push/PR state (Route B; per mission, git stays with the lead unless re-delegated).
- **E7** Blocker reports (none anticipated beyond §D riders).
- **E8** RED evidence: verbatim failing output of the flipped tests BEFORE the GREEN edit (TDD falsifiability).

## §F Milestones (priority order, no time estimates)

- **M1 — RED (decision-heavy first).** Flip the four runtime test files (template overlay test incl. session-default + AC-GET-003 low-anchor; cli session-default env; cli ForEffort rows; web agentfm chip). Capture verbatim RED output (E8). *Encodes the §D-1 decision — ratified 2026-08-22; unblocked.*
- **M2 — GREEN.** `glm_effort_overlay.go`: collapse branch medium/high→`glmReasoningMax`; session default→`glmReasoningMax`; delete `glmReasoningHigh`; rewrite cost-policy comment on the ratified grounds + UNVERIFIED comments; glm.go + profile_matrix.go comment updates (REQ-GEM-005). Affected-package tests green.
- **M3 — Doc surface.** Template-mirror doc block (per §D-2, ratified) + `make build` + mirror checks, same commit.
- **M4 — Verify & evidence.** E1-E8 sweep, cross-platform build, AC matrix, `progress.md` §E.2 population (incl. AC-GEM-008 fresh-session check handoff to lead).

## §G Anti-Patterns

- Do NOT "fix" `glm_tier_test.go` or `glmDefaultTierEffort` — stored-only surface; their `high` defaults are the recorded divergence, not a defect (spec.md §5).
- Do NOT delete `GLMStateHigh`/`GLMReasoningEffortHigh` — only the `glmReasoningHigh` *value* dies; the name/wire constants stay live (REQ-GEM-003).
- Do NOT run the local full suite; do NOT push to `main`.
- Do NOT widen scope to per-agent wire channels, profile cells, or REBALANCE's artifacts.
- Do NOT claim the fresh-session env check from inside this worktree session — it needs a rebuilt binary and a new session (B-6); hand it to the lead with the exact command.

## §H Cross-References

- spec.md §1.3 (REBALANCE conflict), §4 (C-1..C-7), §5 (exclusions)
- `.claude/rules/local/repo-local-pr-policy.md` — Route B mandate
- `manager-develop-prompt-template.md` § Applicability — Tier S minimal delegation form
- `verification-claim-integrity.md` §2-3 — E-item attribution + Gaps discipline
