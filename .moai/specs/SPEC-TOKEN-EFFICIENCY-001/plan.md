# Implementation Plan — SPEC-TOKEN-EFFICIENCY-001

## §A Context

- Project root: `/Users/goos/MoAI/moai-adk-go`
- Tier: **M** (2 items / 2 Go domains / ~6–8 files / ~250–500 LOC)
- Artifact set: spec.md + plan.md + acceptance.md + progress.md
- plan-auditor PASS threshold: **0.80**
- Delegation template: Section A-E REQUIRED (Tier M) per
  `.claude/rules/moai/development/manager-develop-prompt-template.md`
- **Scope note (v0.2.0)**: the "~7×" reword (formerly P0-2) is carved out to
  SPEC-DIVECC-ATTRIBUTION-FIX-001. This SPEC now ships P0-1 (budget guard) +
  P0-2 (cache-hit-ratio statusline, formerly P0-3).

## §B Known Issues / pre-scanned risks (domain-filtered)

- **B1 Cross-platform build tags**: P0-1 and P0-2 are pure-Go, no `syscall`
  usage expected; `GOOS=windows go build ./...` must still pass.
- **B4 Frontmatter canonical schema**: this SPEC's own frontmatter uses the
  canonical 12 fields (`created`/`updated`/`tags`, not snake_case).
- **B6 spec-lint heading rule**: spec.md carries `### Out of Scope — <topic>`
  H3 sub-headings with `-` bullets (satisfies `MissingExclusions`).
- **B8 working-tree hygiene**: P0-2 must not touch runtime-managed files
  (`.moai/status_line.sh` is rendered; edit the Go package + `.sh.tmpl` only if
  the tmpl needs change — it does NOT for this SPEC).
- **B10 PRESERVE**: do not modify `internal/statusline` stdin schema
  (`types.go` `CurrentUsageInfo` stays as-is); do not modify the memory.go
  Priority-1/2/3 tokensUsed logic (P0-2 adds a SEPARATE derivation, not a
  replacement). Do NOT touch `.claude/output-styles/moai/moai.md` or its
  template mirror — the "~7×" reword is carved out to a separate SPEC.

## §C Item file targets

### P0-1 — always-loaded token-budget guard

- NEW: a guard test in the `internal/config` package (model:
  `internal/config/audit_test.go` `TestAuditParity` structure — repo-file scan +
  `t.Errorf` on violation; skip gracefully when run outside the repo tree).
- NEW: a named budget constant + token-estimate function (in `internal/config`
  or a small helper file alongside the test).
- Always-loaded surface enumerated dynamically: `CLAUDE.md` +
  `.claude/rules/moai/**/*.md` files with NO `paths:` frontmatter +
  `.claude/output-styles/moai/moai.md` + `MEMORY.md` head (200 lines / 25KB
  cap). Current baseline (measured 2026-07-02): 13 files, ~258KB total.

### P0-2 — cache-hit-ratio statusline

- EDIT/NEW: `internal/statusline/` — a derivation function computing the
  cache signal from `StdinData.ContextWindow.CurrentUsage.{CacheReadTokens,
  CacheCreationTokens}`, a new segment constant (parallel to
  `SegmentEffortThinking`), a renderer hook in `renderer.go`
  (`renderInfoLine` or `renderBarsInline`), and a config toggle entry.
- Graceful degradation: reuse the existing null-`CurrentUsage` handling shape
  in `memory.go` (Priority-1 `UsedPercentage` path already covers the null
  case) — when `CurrentUsage` is nil OR `CacheCreationTokens == 0`, omit the
  segment.
- Tests: table-driven `renderer_test.go` / a new `*_test.go` covering: normal
  ratio, null `current_usage`, zero cache-creation (divide-by-zero guard),
  segment-disabled.

## §D Plan decisions (resolve before RED)

### D-1 — P0-1 token-estimation method (DECISION REQUIRED)

Two candidate methods:

| Option | Method | Pro | Con |
|--------|--------|-----|-----|
| A (Recommended) | char/4 heuristic (`len(bytes)/4`) | zero dependency, deterministic, matches the common rule-of-thumb; simplicity ladder rung 6-7 | ±15% vs a real tokenizer |
| B | real tokenizer (tiktoken-go / anthropic tokenizer) | exact | new dependency, version drift, slower test |

**Recommendation: Option A (char/4).** The guard is a *regression tripwire*, not
an accounting ledger — a deterministic, dependency-free estimate that tracks
relative growth is sufficient and honors the simplicity ladder (avoid a new
dependency for a guard). Record the method as a named function
(`estimateTokens(b []byte) int { return len(b) / 4 }`) with a comment stating it
is an approximation. Final choice deferred to run-phase implementer + user
confirmation via the orchestrator if they prefer B.

### D-2 — P0-1 budget value (DECISION REQUIRED)

Current measured baseline (2026-07-02): always-loaded surface ≈ 258KB ≈ ~64,500
tokens (char/4). Proposal: set the budget at the **current baseline + ~15%
headroom**, rounded to a clean constant (e.g. `AlwaysLoadedTokenBudget = 75000`),
so the guard fires on meaningful growth but not on today's tree. The exact
number is a run-phase decision recorded as a commented constant; the AC
(AC-TEF-002) asserts the guard PASSES on the current tree and FAILS on a
synthetic over-budget fixture — it does NOT hardcode the number in the AC.

### D-3 — P0-2 display form (DECISION REQUIRED)

Options: (a) cache-hit % = `cache_read / (cache_read + cache_creation) * 100`;
(b) raw ratio `cache_read : cache_creation`. **Recommendation: cache-hit %**
(single number, directly comparable to CC's own "cache-hit rate" SEV metric,
easier to color-threshold). Zero/null → omit segment. Final form deferred to
run-phase; AC-TEF-005/006 assert behavior, not the exact glyph.

## §E Self-Verification (plan-phase audit-ready signal)

- [ ] spec.md carries 12 canonical frontmatter fields (verified: yes)
- [ ] SPEC ID passes `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (self-check PASS printed)
- [ ] `### Out of Scope — <topic>` H3 sub-headings present with `-` bullets
- [ ] GEARS REQs for both items (REQ-TEF-001…007)
- [ ] acceptance.md AC bijection with REQs (7 REQ ↔ 7 AC)
- [ ] P0-2 data-availability VERIFIED against official docs + Go struct
- [ ] "~7×" reword carved out to SPEC-DIVECC-ATTRIBUTION-FIX-001 (not in scope)
- [ ] Over-engineering guard: CC-native compaction/caching explicitly out of scope

## §F Milestones (priority-ordered, no time estimates)

- **M1 — P0-1 budget guard.** Add `estimateTokens` + budget constant + guard
  test; verify PASS on current tree + FAIL on synthetic over-budget fixture.
- **M2 — P0-2 statusline cache signal.** Add derivation + segment + renderer
  hook + config toggle + table-driven tests incl. null/zero degradation.
- **M3 — cross-item verification.** `go test ./...`, `go vet`, `golangci-lint`,
  `GOOS=windows go build ./...`.

Milestones are independently landable (D-2 in spec.md): M1/M2 can merge in
any order; M3 is the aggregate gate.

## §G Anti-Patterns to avoid

- Reimplementing CC compaction/caching (AP-RR-001) — out of scope.
- Adding a tokenizer dependency for a tripwire guard (violates simplicity
  ladder) — prefer char/4 unless the user requests exactness.
- Touching `.claude/output-styles/moai/moai.md` or its template mirror for the
  "~7×" reword (scope discipline — the reword is carved out to
  SPEC-DIVECC-ATTRIBUTION-FIX-001; P0-1 only *reads* moai.md as a measurement
  input, never edits it).
- Editing `.moai/status_line.sh` (rendered artifact) or the `.sh.tmpl` wrapper
  for P0-2 (the wrapper only forwards stdin).
- Hardcoding the always-loaded file list in P0-1 (enumerate dynamically so it
  self-updates as rules are scoped/added).

## §H Cross-References

- `.claude/rules/moai/core/verification-claim-integrity.md` — every claim cited
  in this SPEC (§A.1, §A.3) is attributed to a verified source.
- `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` AP-RR-001 —
  CC-native mechanism reimplementation prohibition (over-engineering guard).
- `.claude/rules/moai/workflow/context-window-management.md` — graduated
  compaction layer names (consumed, not implemented) + MEMORY.md 200-line/25KB
  loader cap (P0-1 head definition).
- `internal/statusline/types.go` / `memory.go` — existing cache-field parse +
  consumption (P0-2 builds on these).
- `internal/config/audit_test.go` — guard-test structural model for P0-1.
- SPEC-DIVECC-ATTRIBUTION-FIX-001 — the carved-out "~7×" reword follow-up SPEC.
- Official statusline schema: https://code.claude.com/docs/en/statusline
