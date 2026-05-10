# SPEC-V3R2-SPC-004 — @MX anchor resolver (query by SPEC ID, fan_in, danger category) (Plan PR)

> Step 1 plan-in-main per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline.
> Branch: `plan/SPEC-V3R2-SPC-004` · Base: `origin/main` HEAD `73742e3ee` · Merge strategy: squash.

## Summary

Expose a structured query API over the @MX TAG sidecar index (SPEC-V3R2-SPC-002) as an ACI-shaped tool: `moai mx query`. Callers can filter tags by SPEC ID, by Kind (NOTE/WARN/ANCHOR/TODO/LEGACY), by fan_in threshold (high-callsite anchors for invariant enforcement), by danger category (WARN-tagged items whose REASON sub-line matches safety-critical patterns), and by file path prefix.

Output format is JSON (primary) for tool consumption and a human-readable table (secondary) for terminal inspection. This same API is consumed by codemaps generation, evaluator-active scoring hints, and the /moai review phase.

This PR contains **plan artifacts only** (research / plan / acceptance / tasks / progress / issue-body). No code or hook changes ship in this PR — those land in the run PR (`feat/SPEC-V3R2-SPC-004`) after this plan PR squash-merges into `main`.

## Why this matters

- **Resolver completes the T-1 ACI priority-1 pattern**: pattern-library.md §T-1 names `moai_locate_mx_anchor` as the strongest single-leverage ACI command. Together with SPC-002 (sidecar) and SPC-001 (`/moai mx index --json`), this SPEC forms the 6-command suite per design-principles.md §P2 "Interface Design Over Tool Count".
- **Downstream consumers depend on it**: codemaps generation needs per-module anchor count; evaluator-active (HRN-003) needs danger_category distribution as a harness signal; `/moai review` needs structured query into the @MX layer.
- **90%+ of the code already exists**: SPC-004 PR #746 (`68795dbe3`, 2026-04-30) merged `internal/mx/{resolver_query,fanin,danger_category,spec_association}.go` with all 10 CLI flags, 11-field JSON schema, 3 sentinel errors, JSON/table/markdown formatters, AND-composed multi-filter, and 15 AC fixtures. **Run-phase scope is gap-closure (LSP integration + user config wire-up + 16-language sweep), not from-scratch build** — see plan §1.2.

## Plan-phase deliverables (this PR)

- `spec.md` (already on main; no change)
- `research.md` — 45 evidence anchors, 9 OQ resolutions, cross-SPEC boundary survey for SPC-002 / LSP-CORE-002 / HRN-003 / RT-001, gap inventory G-01..G-08, danger_categories default mapping survey, 16-lang LSP availability matrix.
- `plan.md` — 6 milestones (M1..M6), 20 tasks, 18 REQ → 15 AC → 20 Task traceability matrix; 6 @MX tags planned; risk mitigation cross-referenced to spec §8.
- `acceptance.md` — 15 ACs in Given/When/Then form with verification commands, test fixture references (11 of 15 already in `internal/mx/resolver_query_test.go` and `internal/cli/mx_query_test.go`), and Definition of Done.
- `tasks.md` — 20 tasks with REQ/AC traceback, dependency graph, SPC-002 reuse plan; no wave-split required (under 30-task threshold).
- `progress.md` — milestone tracker shell + AC status + risk watch.
- `issue-body.md` — this file.

## Run-phase scope (next PR after this plan PR merges)

After plan PR squash-merge:

1. `moai worktree new SPEC-V3R2-SPC-004 --base origin/main` (Step 2 spec-workflow).
2. `/moai run SPEC-V3R2-SPC-004` invokes Phase 0.5 plan-audit gate.
3. **M1 — LSP `find-references` integration (G-01) (P0)**:
   - New `internal/mx/fanin_lsp.go` (~150 LOC) with `LSPFanInCounter` (powernap-backed).
   - 4 RED tests (`internal/mx/fanin_lsp_test.go` ~250 LOC).
   - `Resolver.Resolve` strictMode 분기 강화 (LSP availability detect; replaces current unconditional `LSPRequired` return).
4. **M2 — `mx.yaml` `danger_categories:` user wire-up (G-02) (P1)**:
   - `LoadDangerConfig()` helper in `internal/mx/danger_category.go`.
   - `validateQuery()` extended to validate `--danger` value against known categories.
   - CLI exit code 2 fix on `*InvalidQueryError` (currently exits 1).
5. **M3 — `.moai/specs/*/spec.md` `module:` 자동 로드 (G-03) (P1)**:
   - New `internal/mx/spec_loader.go` (~80 LOC) — yaml frontmatter parse, supports both string + array module formats.
   - SpecAssociator path-based association activated (currently inactive due to empty map).
6. **M4 — `Resolver.ResolveAnchorCallsites()` API parity (G-04) (P1)**:
   - New `internal/mx/callsite.go` with Callsite struct.
   - Additive `Resolver.ResolveAnchorCallsites(ctx, anchorID, projectRoot, includeTests) ([]Callsite, error)`.
   - Existing `ResolveAnchor(anchorID) (Tag, error)` preserved (backward-compat).
7. **M5 — test_paths glob (G-05) + stderr verify (G-06) (P1)**:
   - `isTestFile` extended with user `mx.yaml test_paths:` glob patterns.
   - Explicit `TestSidecarUnavailable_StderrFormat` fixture verifies both substrings present.
8. **M6 — Verification (P0)**:
   - CLI wire-up (LoadDangerConfig + LoadSpecModules + LSP detect).
   - 16-language sweep (G-08) verifying resolver layer across all 16 supported languages.
   - Performance benchmark fixtures (G-07; advisory).
   - Full test suite + lint + build + CHANGELOG + 6 @MX tags + manual smoke test.

Estimated artifacts: 7 new Go source files + extensions to 4 existing test files + 5 modified Go files + CHANGELOG = **~1,400 LOC delta** (production ~480 + test ~920).

## Acknowledged discrepancies (from plan §1.2.1)

1. **90%+ of code already exists.** SPC-004 PR #746 (`68795dbe3`) merged the resolver query API + CLI subcommand + 4 of 5 fan_in / danger / spec_association / pagination subsystems. Run-phase task count reflects gap-closure not from-scratch build. Sync-phase HISTORY in spec.md should reconcile §1 "Expose ... API" → "complete the API to spec parity (LSP integration + user config wire-up + 16-language coverage)".

2. **`Resolver.ResolveAnchor(anchorID) []Callsite` signature gap.** Spec verbatim is `[]Callsite`; current implementation returns `(Tag, error)`. Two APIs have different semantics — plan-phase decision: existing `ResolveAnchor` preserved + new `ResolveAnchorCallsites()` added (additive, backward-compat). G-04 closes the gap.

3. **JSON output envelope.** Spec §5.3 REQ-021 "TruncationNotice in the output header" interpretable as stdout envelope, but current CLI emits stdout=`[]TagResult` slice + stderr=TruncationNotice. Plan-phase decision (research §9 OQ-5): preserve existing format for backward-compat. Go API exposes envelope (`QueryResult`); CLI keeps slice-only stdout.

4. **`--danger` enum validation gap.** Current `validateQuery()` only enum-checks `Kind` (research [E-15]); `--danger` accepts arbitrary strings (silent empty result). Plan-phase decision (research §9 OQ-4): extend `validateQuery` with `DangerCategoryMatcher.ValidateCategory()` call → `InvalidQueryError` on unknown category. Empty `--danger` (no filter) skips validation.

5. **LSP availability detection.** Current strictMode 분기 (research [E-21]) returns unconditional `LSPRequired` (todo "(no LSP client)"). G-01 replaces with powernap server discovery result. Backward-compat: strictMode unset → no behavior change.

6. **Performance budget verification format.** Spec §7 of <100ms / <2s is machine-dependent; CI does not enforce. Plan-phase decision (research §9 OQ-7): benchmark fixtures (`go test -bench`) for regression detection; absolute value assertion is advisory.

## SPC-002 schema invariant guarantee

This SPEC is a read-only consumer of the SPC-002 sidecar (`.moai/state/mx-index.json`). All changes:

- Read sidecar via `mx.Manager.Load()` (existing API).
- Iterate `sidecar.Tags` (read-only).
- Apply filters in-memory.
- Compute fan_in via separate counter (no sidecar mutation).

Zero changes to:
- `internal/mx/sidecar.go`
- `internal/mx/scanner.go`
- `internal/mx/tag.go`
- `internal/mx/comment_prefixes.go`
- `.claude/rules/moai/workflow/mx-tag-protocol.md` (FROZEN per CONST-V3R2-003)

`SchemaVersion = 2` invariant preserved per SPC-002 plan §1.4.

## File-level scope (run-phase)

| Layer | Files |
|---|---|
| LSP integration | `internal/mx/fanin_lsp.go` (new), `internal/mx/fanin_lsp_test.go` (new) |
| Config loaders | `internal/mx/danger_category.go` (extended), `internal/mx/spec_loader.go` (new) |
| API parity | `internal/mx/callsite.go` (new), `internal/mx/resolver.go` (extended) |
| Validation | `internal/mx/resolver_query.go` (extended) |
| Test exclusion glob | `internal/mx/fanin.go` (extended) |
| CLI wire-up | `internal/cli/mx_query.go` (extended) |
| Test fixtures | 3 new test files (fanin_lsp_test.go, spec_loader_test.go, resolver_callsites_test.go) + extensions to 4 existing test files |
| Performance | `internal/mx/resolver_query_bench_test.go` (new, advisory) |
| Build | `make build` (regenerates `internal/template/embedded.go` defensively; no template diff expected) |
| Trackability | `CHANGELOG.md` Unreleased entry, 6 @MX tags |

## Test plan

- [ ] M1 RED tests: LSPFanInCounter 4 sub-tests + strictMode 강화 → all FAIL initially
- [ ] M1 GREEN: LSPFanInCounter implementation + Resolve strictMode 분기 → all GREEN
- [ ] M2 RED: LoadDangerConfig 4 sub-tests + validateQuery danger 분기 → all FAIL
- [ ] M2 GREEN: helper + validation extension → all GREEN
- [ ] M3 RED: spec_loader 5 sub-tests + SpecAssociator integration → all FAIL
- [ ] M3 GREEN: helper land → all GREEN
- [ ] M4 RED: Callsite struct + ResolveAnchorCallsites 4 sub-tests → all FAIL
- [ ] M4 GREEN: additive method lands → all GREEN; existing ResolveAnchor unchanged
- [ ] M5 RED: isTestFile glob 3 sub-tests + stderr fixture → all FAIL
- [ ] M5 GREEN: helper + integration → all GREEN
- [ ] M6 final: 16-lang sweep PASS, benchmarks compile, `go test -race -count=1 ./...` PASS, `golangci-lint run` clean, `make build` exit 0, end-to-end manual verification (real gopls)

## Dependencies

| SPEC | Status | Role |
|---|---|---|
| SPEC-V3R2-SPC-002 (sidecar contract) | merged (plan PR #836 at `73742e3ee`; sidecar code at PR #741 commit `3f0933550`) | Blocks (consumed); read-only sidecar reader; schema_version: 2 invariant preserved |
| SPEC-LSP-CORE-002 (LSP client via powernap) | merged (assumed; `internal/lsp/core/client.go` present) | Blocks (consumed); G-01 LSP-backed counter dependency |
| SPEC-V3R2-HRN-003 (evaluator MX scoring) | in-flight | Downstream consumer (danger_category distribution as harness signal) |
| codemaps generation tools | in-flight | Downstream consumer (per-module anchor count via Resolver.Resolve) |
| pattern-library.md §T-1 priority 1 | design pattern source | Establishes "ACI 6-command suite" including this SPEC |

## Plan-auditor target

- **PASS verdict** ≥ 0.85 first iteration.
- 18 REQs (REQ-SPC-004-001..007, 010..013, 020..021, 030..031, 040..042) traced to ≥1 AC and ≥1 task per plan §1.5 traceability matrix.
- 15 ACs each with Given/When/Then + verification command + REQ traceback.
- 45 evidence anchors in research.md (counted [E-01]..[E-45]).
- 9 OQ resolutions in research §9.
- §1.2.1 explicitly addresses 6 known plan-vs-spec discrepancies (90%+ already implemented; ResolveAnchor signature; JSON envelope; danger 검증 gap; LSP detect; benchmark format).
- 6 @MX tags planned (1 ANCHOR + 2 WARN + 3 NOTE) — covers all 3 of {ANCHOR, WARN, NOTE} types.
- Worktree-base alignment per Step 2 (`moai worktree new --base origin/main`) called out (plan §10).
- Parallel SPEC isolation: only `internal/mx/`, `internal/cli/`, `CHANGELOG.md`. No `.claude/` tree changes expected.
- SPC-002 schema_version: 2 invariant preservation 명시 (plan §1.2.1 + plan §3.5).

## References

- `.claude/rules/moai/workflow/mx-tag-protocol.md` (FROZEN inline syntax — read-only reference)
- `.claude/rules/moai/core/zone-registry.md` `CONST-V3R2-003`
- `internal/mx/{tag,scanner,sidecar,resolver,resolver_query,fanin,danger_category,spec_association,comment_prefixes}.go` (existing implementation; PR #741 + PR #746)
- `internal/mx/{resolver_query_test,fanin_test,danger_category_test,spec_association_test}.go` (existing test fixtures with verbatim AC-SPC-004-NN annotations)
- `internal/cli/{mx,mx_query,mx_query_test}.go` (existing CLI surface)
- `internal/lsp/core/{client,capabilities,document}.go` (powernap LSP client — G-01 dependency)
- `.moai/config/sections/mx.yaml` (config target for danger_categories + test_paths)
- spec.md §1 (Goal), §2 (Scope), §3 (Environment), §4 (Assumptions), §5 (REQ), §6 (AC), §7 (Constraints), §8 (Risks), §9 (Dependencies), §10 (Traceability)
- design-principles.md §P2 (Interface Design Over Tool Count); pattern-library.md §T-1 priority 1
- master-v3 §7.3 (ACI command list including `moai_locate_mx_anchor`)

---

🗿 MoAI <email@mo.ai.kr>
