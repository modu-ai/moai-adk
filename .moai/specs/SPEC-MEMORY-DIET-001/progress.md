# progress.md — SPEC-MEMORY-DIET-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-07-10
- tier: M
- artifacts: spec.md + plan.md + acceptance.md + progress.md
- ac_count: 17
- req_count: 17

## §E.2 Run-phase Evidence

### Baseline measurements (pre-flight, 2026-07-10)

| File | Lines | Bytes |
|------|-------|-------|
| `.claude/rules/moai/workflow/cadence-bridge.md` | 88 | 9,855 |
| `internal/template/templates/.claude/rules/moai/workflow/cadence-bridge.md` | 88 | 9,855 |
| `.claude/rules/moai/workflow/session-handoff.md` | 478 | 56,598 |
| `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md` | 478 | 56,598 |
| `~/.claude/projects/.../memory/MEMORY.md` | 91 | 17,109 |

MEMORY.md marker baseline: ✅=30, active-markers(🟢🟡🆕⏸️⚠️🔍)=18, load-bearing ✅(다음=|handoff|deferred|debt|후속|pending)=16

session-handoff doctrine integrity baseline: ✂(U+2702)=16, ─(U+2500)=17, Block headings=45, Pre-emit labels(paste-ready budget|localization render|session-handoff template completeness)=5

Build baseline: `go build ./...` exit=0. Branch: worktree-agent-a0c59bd587cb9a9ac at origin/main (0 0 divergence).

### M1 — REQ-1: cadence-bridge.md path-scope (AC-MD-001..005)

- AC-MD-001 PASS: frontmatter acquired (`---`, `description:`, `paths:` all present in first 10 lines)
- AC-MD-002 PASS: `paths:` glob committed + Loading-scope prose documents trigger-condition list + honest-fallback clause
- AC-MD-003 PASS: "Intentionally always-loaded" removed (0 matches); path-match declaration present
- AC-MD-004 PASS: `diff` local↔template = empty (byte-identical frontmatter)
- AC-MD-005 PASS: 8 H2 sections survive (body byte-preserved modulo Loading-scope prose rewrite)
- Post-edit: 97 lines / 11,130 bytes each (local + template byte-identical)

### M2 — REQ-2: session-handoff.md illustrative content extraction (AC-MD-006..011)

- AC-MD-006 PASS: 0 `### Example` H3 headings remain; 4 pointers to session-handoff-examples.md present
- AC-MD-007 PASS: en/ko columns retained inline; ja/zh pointer to examples file present (2 matches)
- AC-MD-008 PASS: ✂=10 (≥4), ─=11 (≥8), Block headings=45, Pre-emit labels=5, core H2 sections=5
- AC-MD-009 PASS: moai.md cross-references session-handoff (10 matches, baseline preserved)
- AC-MD-010 PASS: session-handoff.md + examples file both byte-identical local↔template
- AC-MD-011 PASS: examples file carries `paths: "**/session-handoff.md"` (NOT always-loaded)
- session-handoff.md: 431 lines / 54,593 bytes (was 478 / 56,598 — saved ~2,005 bytes)
- session-handoff-examples.md: 78 lines / 5,440 bytes (NEW, path-scoped)

### M3 — REQ-3: MEMORY.md archive pruning (AC-MD-012..015)

- AC-MD-012 PASS: ✅ count decreased 30→25 (5 stable entries moved to archive); MEMORY.md 91→86 lines
- AC-MD-013 PASS: active-marker count unchanged at 18 (all preserved)
- AC-MD-014 PASS: load-bearing ✅ count unchanged at 16 (all preserved)
- AC-MD-015 PASS: MEMORY.md 86 lines / 15,929 bytes (≤200 lines, ≤25,600 bytes); archive grew 974→986 lines
- MEMORY.md savings: 17,109 → 15,929 bytes (~1,180 bytes saved from always-loaded index)
- Archived entries: fix·loop 3-트랙, MOAI-SKILL-DOCTRINE-FIX, TOKEN-VERIFY-DIET, Sonnet 5 1M, HOOK-FACTFORCE-ADVISORY

### Token reduction measurement (post-run, 2026-07-10)

Combined always-loaded token reduction: **~3.0k tokens** (revised from the original plan-phase target of ~5.5k).

Per-file breakdown (orchestrator independent Trust-but-verify: wc + byte→token estimate at ~4 bytes/token for English markdown):

| File group | Before | After | Delta | ~Tokens |
|------------|--------|-------|-------|---------|
| cadence-bridge.md (path-match — leaves always-loaded set) | 88 lines / 9,855B (always-loaded) | 97 lines / 11,130B (path-matched, excluded) | file leaves set entirely | ~2.3k (dominant) |
| session-handoff.md (example extraction) | 56,598B | 54,650B | -1,948B | ~487 |
| MEMORY.md (archive pruning) | 17,109B | 16,266B | -843B | ~210 |
| **Combined** | | | | **~3.0k** |

**Honest caveat / root cause of the original 5.5k over-estimate.** The plan-phase Discovery report conflated BYTES saved with TOKENS saved for the two smaller items: session-handoff (~2k bytes was misread as ~2k tokens) and MEMORY.md (~1.5k bytes was misread as ~1.5k tokens). The cadence-bridge estimate (~2.3k tokens) was always a true token estimate of the file leaving the always-loaded set. The honest combined figure is ~3.0k tokens, dominated by the cadence-bridge path-match conversion.

**Authoritative measurement.** The `/context` command reports the always-loaded "Memory files" total; the authoritative before/after delta is measured by re-running `/context` after a `/clear` (so the path-matched cadence-bridge.md is no longer in the Memory files set). The wc byte deltas above are the file-level evidence; the `/context` delta is the user-visible outcome. Per verification-claim-integrity §3, the wc measurements are the Evidence (directly observed), and the ~4 bytes/token conversion is the Baseline-attribution (standard English-markdown estimate); the `/context` re-measurement is the authoritative confirmation pending the next `/clear`.

## §E.3 Run-phase Audit-Ready Signal

- run_status: audit-ready
- run_complete_at: 2026-07-10
- ac_pass_count: 17
- ac_fail_count: 0
- preserve_list_post_run_count: 0
- cross_platform_build: go build ./... exit 0 (no Go source changes)
- lint_status: moai spec lint + golangci-lint baseline preserved (no Go changes)
- neutrality_ci: go test ./internal/template/... pass (template-neutrality guards intact)

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: audit-ready
- sync_complete_at: 2026-07-10
- sync_commit_sha: d8299d64a00a66004bfcb09b595406b76b2062e6
- ac_final_count: 17
- ac_pass_count: 17
- ac_fail_count: 0
- close_type: 3-phase (plan→run→sync merged into single sync commit per SPEC-V3R6-LIFECYCLE-DESIGN-001)
- tier: M
- frontmatter_status_transition: in-progress → implemented → completed (merged close)
- lint_status: moai spec lint baseline preserved (no Go changes)
- neutrality_ci: template-neutrality CI pass (REQ-1/REQ-2 template mirrors intact, no SPEC IDs leaked)
- CHANGELOG_entry_position: [Unreleased] section populated
- canary_compliance_check:
    - status_git_consistency: resolved (status=completed aligns with git-implied; verify via moai spec lint post-backfill)
    - mx_tags: N/A (no @MX tags added; out of scope for this SPEC)
    - era_classification: V3R6 (§E.2 + §E.3 + §E.4 present, sync_commit_sha populated)

## §F Phase 0.95 Mode Selection

Input parameters:
- tier: M
- scope: ~5 files (cadence-bridge.md local+template, session-handoff.md local+template, session-handoff-examples.md local+template new, MEMORY.md auto-memory)
- domain count: 2 (rules/doctrine + auto-memory)
- file language mix: 100% markdown
- concurrency benefit: LOW (sequential document edits with ordering dependencies + AC-gated verification; not research-fanout)
- Agent Teams prereqs: NOT met (workflow.team.enabled default false, env var not set)

Mode evaluation:
- Mode 1 (trivial): not selected — multi-file doctrine edit with regression risk
- Mode 2 (background): not selected — write tasks (file edits), not read-only
- Mode 3 (agent-team): not selected — prereqs not met, scope < 3 domains / < 10 files
- Mode 4 (parallel): not selected — editing-heavy work per Anthropic coding-task parallelism caveat; sequential safer (template mirror must match local; doctrine byte-identity verified before/after)
- Mode 5 (sub-agent): selected — single implementation agent, AC-gated, Tier M
- Mode 6 (workflow): not selected — < 30 files, not mechanical-uniform

Decision: sub-agent

Justification: Tier M document-diet SPEC touching always-loaded doctrine files + template mirrors + auto-memory. Per Anthropic's coding-task parallelism caveat, the sequential sub-agent path is the safe default for editing-heavy work where file edits carry ordering dependencies. Mode 4 parallel offers no benefit (one coherent edit set, not independent research lenses). Agent Teams disabled by default (Sonnet 5 / Opus 4.8 era). skip-eligible: NO (plan-auditor iter-2 score 0.89 < 0.90) — Phase 0.5 plan-audit was performed as iter-1/iter-2 independent adversarial review; Implementation Kickoff Approval obtained separately via AskUserQuestion.
