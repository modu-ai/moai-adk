# SPEC-V3R2-SPC-002 — @MX TAG v2 with hook JSON integration and sidecar index (Plan PR)

> Step 1 plan-in-main per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline.
> Branch: `plan/SPEC-V3R2-SPC-002` · Base: `origin/main` HEAD `fcb486c87` · Merge strategy: squash.

## Summary

Layer two additive capabilities on top of the FROZEN inline @MX TAG protocol (mx-tag-protocol.md, CONST-V3R2-003):

1. **Machine-readable JSON sidecar** at `.moai/state/mx-index.json` (`schema_version: 2`, atomic temp+rename writes) that materializes every @MX TAG as a structured Tag record (Kind/File/Line/Body/Reason/AnchorID/CreatedBy/LastSeenAt). Enables tool consumption (SPC-004 resolver, evaluator-active, codemaps) without re-scanning source on every query.

2. **PostToolUse hook integration** via SPEC-V3R2-RT-001's JSON dual-protocol — the new `post_tool_mx` handler emits `additionalContext` (human-readable summary) and `hookSpecificOutput.mxTags` (structured Tag array) into the model turn, AND atomically updates the sidecar.

The inline source-code convention `// @MX:KIND text` stays FROZEN. The sidecar is a **view** — never the source of truth.

This PR contains **plan artifacts only** (research / plan / acceptance / tasks / progress / issue-body). No code or hook changes ship in this PR — those land in the run PR (`feat/SPEC-V3R2-SPC-002`) after this plan PR squash-merges into `main`.

## Why this matters

- **Sidecar enables tooling at scale**: SPC-004 resolver query (already merged, PR #746) reads the sidecar. Without atomic write semantics + schema_version pinning, resolver consumers face partial-read crashes. This SPEC formalizes the contract.
- **PostToolUse integration closes the agent feedback loop**: Today, Claude edits a file with `@MX:WARN goroutine leak` and the warning vanishes into the file system. With this SPEC, the next model turn sees the WARN injected into context AND consumers (evaluator-active per HRN-003) see the structured tag.
- **80% of the code already exists**: Wave 3 PR #741 (`3f0933550`) merged `internal/mx/` package skeleton (tag.go, scanner.go, sidecar.go, comment_prefixes.go, resolver.go). SPC-004 PR #746 (`68795dbe3`) added resolver_query.go, fanin.go, danger_category.go, spec_association.go. **Run-phase scope is gap-closure, not from-scratch build** — see plan §1.2.

## Plan-phase deliverables (this PR)

- `spec.md` (already on main; no change)
- `research.md` — 30 evidence anchors, 9 OQ resolutions, cross-SPEC boundary survey for CON-001 / RT-001 / SPC-004 / HRN-003 / WF-005, gap inventory G-01..G-08.
- `plan.md` — 6 milestones (M1..M6), 22 tasks, 22 REQ → 15 AC → 22 Task traceability matrix; 6 @MX tags planned; risk mitigation cross-referenced to spec §8.
- `acceptance.md` — 15 ACs in Given/When/Then form with verification commands, test fixtures (Go), and Definition of Done.
- `tasks.md` — 22 tasks with REQ/AC traceback, dependency graph, SPC-004 reuse plan.
- `progress.md` — milestone tracker shell + AC status + risk watch.
- `issue-body.md` — this file.

## Run-phase scope (next PR after this plan PR merges)

After plan PR squash-merge:

1. `moai worktree new SPEC-V3R2-SPC-002 --base origin/main` (Step 2 spec-workflow).
2. `/moai run SPEC-V3R2-SPC-002` invokes Phase 0.5 plan-audit gate.
3. **M1 — PostToolUse handler + types (P0)**:
   - New `internal/hook/post_tool_mx.go` (~150 LOC) handler emitting `mxTags` + `additionalContext`.
   - Add `MxTags []mx.Tag` field to `HookSpecificOutput` in `internal/hook/types.go`.
   - 16-language regression fixture in `internal/mx/scanner_test.go`.
4. **M2 — `/moai mx` flag dispatcher (P0)**:
   - `--full` (rescan + rebuild + archive sweep)
   - `--index-only` (silent — CI 친화)
   - `--json` (dump current sidecar to stdout)
   - `--anchor-audit` (low fan_in anchor reporter)
5. **M3 — silent env + mx.yaml ignore (P1)**:
   - `MOAI_MX_HOOK_SILENT=1` empties additionalContext but preserves MxTags + sidecar
   - mx.yaml `ignore:` patterns wire-up to Scanner
6. **M4 — Scanner correctness fixtures (P1)**:
   - MissingReasonForWarn 3-line lookahead
   - DuplicateAnchorID refuse-write
   - Corrupt sidecar repair suggestion
   - HookSpecificOutput mismatch validator
7. **M5 — Archive sweep + atomic verify (P1)**:
   - 7-day stale preservation
   - 8-day archive sweep
   - Atomic write race fixture (`-race -count=10`)
8. **M6 — Verification (P0)**:
   - Full test suite + lint + build + CHANGELOG + @MX tags + smoke test

Estimated artifacts: 5 new Go source files + 5 new test files + 4 modified Go files + CHANGELOG = **~1,250 LOC delta** (production ~480 + test ~770).

## Acknowledged discrepancies (from plan §1.2.1)

1. **80% of code already exists.** Wave 3 PR #741 (`3f0933550`) merged the `internal/mx/` skeleton; SPC-004 PR #746 (`68795dbe3`) added resolver/fanin/danger. Run-phase task count reflects gap-closure not from-scratch build. Sync-phase HISTORY in spec.md should reconcile §1 "introduces" → "completes".

2. **REQ-006 vs REQ-040 unification.** Both REQs describe the same code path (Scanner WARN → 3-line lookahead for REASON). tasks.md treats as single task (T-SPC002-09).

3. **MxTags field vs additionalContext embed.** OQ-1 (research §7) chose new `HookSpecificOutput.MxTags` field (omitempty) over JSON embed in additionalContext. Rationale: protocol-level structured emission consistent with RT-001 dual-protocol intent.

4. **schema_version: 2 preservation.** SPC-004 already consumes the sidecar schema. New fields MUST be omitempty; no field semantic changes.

5. **`internal/hook/file_changed.go:supportedExtensions` drift with `comment_prefixes.go`** is acknowledged advisory; integration deferred to WF-005 16-language enum SoT SPEC.

## FROZEN preservation guarantee

The inline @MX TAG syntax defined in `.claude/rules/moai/workflow/mx-tag-protocol.md` is FROZEN per zone-registry.md `CONST-V3R2-003`. This SPEC adds NO new tag kinds, NO sub-line modifications, NO syntax variants. Every change lives in:

- `internal/hook/post_tool_mx.go` (new handler)
- `internal/hook/types.go` (new MxTags field)
- `internal/cli/mx.go` (new flags)
- `internal/mx/{config,integration,anchor_audit}.go` (new helpers)
- `internal/mx/scanner_test.go`, `internal/mx/sidecar_test.go` (new fixtures)

Zero changes to `.claude/rules/moai/workflow/mx-tag-protocol.md`. Zero changes to existing inline syntax in agent files.

## Plan-auditor target

- **PASS verdict** ≥ 0.85 first iteration.
- 22 REQs (REQ-SPC-002-001..008, 010..014, 020..022, 030..032, 040..042) traced to ≥1 AC and ≥1 task per plan §1.5 traceability matrix.
- 15 ACs each with Given/When/Then + verification command + REQ traceback.
- 30 evidence anchors in research.md (counted [E-01]..[E-30]).
- 9 OQ resolutions in research §7.
- §1.2.1 explicitly addresses 5 known plan-vs-spec discrepancies.
- 6 @MX tags planned (1 ANCHOR + 2 WARN + 3 NOTE) — covers all 3 of {ANCHOR, WARN, NOTE} types.
- Worktree-base alignment per Step 2 (`moai worktree new --base origin/main`) called out (plan §10).
- Parallel SPEC isolation: only `internal/hook/`, `internal/cli/`, `internal/mx/`, `CHANGELOG.md`. No `.claude/` tree changes expected.

## File-level scope (run-phase)

| Layer | Files |
|---|---|
| Hook handler | `internal/hook/post_tool_mx.go` (new), `internal/hook/types.go` (modified) |
| CLI surface | `internal/cli/mx.go` (extended) |
| MX package | `internal/mx/config.go` (new), `internal/mx/integration.go` (new), `internal/mx/anchor_audit.go` (new), `internal/mx/sidecar.go` (extended) |
| Test fixtures | 5 new test files + extensions to existing scanner_test.go, sidecar_test.go |
| Build | `make build` (regenerates `internal/template/embedded.go` defensively; no template diff expected) |
| Trackability | `CHANGELOG.md` Unreleased entry, 6 @MX tags |

## Test plan

- [ ] M1 RED tests: PostToolUse handler emission (T-SPC002-01..03), MxTags field marshalling (T-SPC002-02), 16-lang fixture (T-SPC002-15) — all FAIL initially
- [ ] M1 GREEN: handler implementation lands → all 4 GREEN
- [ ] M2 RED: `--full` / `--json` / `--anchor-audit` CLI tests FAIL initially
- [ ] M2 GREEN: flag dispatcher + helpers land → all GREEN
- [ ] M3 RED: silent env + mx.yaml ignore tests FAIL
- [ ] M3 GREEN: env + config loader land → GREEN
- [ ] M4 RED: 4 correctness fixtures (MissingReason/Duplicate/Corrupt/Mismatch) FAIL where logic is missing
- [ ] M4 GREEN: refinements + sentinel errors land → GREEN
- [ ] M5 RED: atomic + stale + archive fixtures FAIL initially (especially `-race -count=10`)
- [ ] M5 GREEN: archive sweep + race-safe writes land → GREEN
- [ ] M6 final: `go test -race -count=1 ./...` PASS, `golangci-lint run` clean, `make build` exit 0, end-to-end smoke test verified

## Dependencies

| SPEC | Status | Role |
|---|---|---|
| SPEC-V3R2-CON-001 (FROZEN/EVOLVABLE) | merged (assumed) | Blocks (consumed); CONST-V3R2-003 enforces FROZEN inline syntax |
| SPEC-V3R2-RT-001 (JSON hook protocol) | merged (assumed; types.go implementation in main) | Blocks (consumed); HookSpecificOutput dual-protocol substrate |
| SPEC-V3R2-SPC-004 (resolver query API) | **merged** (PR #746, `68795dbe3`) | Co-resident; this SPEC must preserve schema_version: 2 backward-compat |
| SPEC-V3R2-HRN-003 (evaluator MX scoring) | in-flight | Downstream consumer (may use sidecar at scoring time) |
| SPEC-V3R2-WF-005 (16-language enum SoT) | in-flight | Downstream (advisory drift detection only) |

## References

- `.claude/rules/moai/workflow/mx-tag-protocol.md` (FROZEN inline syntax — read-only reference)
- `.claude/rules/moai/core/zone-registry.md` `CONST-V3R2-003`
- `internal/mx/{tag,scanner,sidecar,resolver,comment_prefixes,fanin,danger_category,spec_association,resolver_query}.go` (existing implementation)
- `internal/hook/types.go` (RT-001 dual-protocol surface)
- `internal/hook/file_changed.go` (existing FileChanged MX integration)
- `internal/cli/mx.go` + `internal/cli/mx_query.go` (existing CLI surface)
- `.moai/config/sections/mx.yaml` (config target for ignore patterns + hook budget)
- spec.md §1 (Goal), §2 (Scope), §5 (REQ), §6 (AC), §7 (Constraints), §8 (Risks)
- design-principles.md §P8 (Hook JSON Protocol); pattern-library.md §T-1 / §T-5 (ACI + dual-protocol patterns)

---

🗿 MoAI <email@mo.ai.kr>
