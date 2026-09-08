# t466 Plan Summary — SPEC-UPDATE-HOOK-DELIVERY-001

**Card**: t466 · **Tree**: `.claude/worktrees/t466` · **Branch**: `WT-update-hook-delivery` · **Baseline**: `d592b0551` · **Date**: 2026-09-03

## SPEC-ID and artifacts

- SPEC-ID: `SPEC-UPDATE-HOOK-DELIVERY-001` (Tier M; frontmatter 12 canonical fields + era/tier; ID regex PASS; no collision in `.moai/specs/`)
- `.moai/specs/SPEC-UPDATE-HOOK-DELIVERY-001/spec.md` — GEARS requirements: 6 invariant + 4 option-gated + 2 robustness REQs; 4 `### Out of Scope —` H3 blocks
- `.moai/specs/SPEC-UPDATE-HOOK-DELIVERY-001/plan.md` — M1-M5 ordered by decision-reversibility (identity scheme first, verification last)
- `.moai/specs/SPEC-UPDATE-HOOK-DELIVERY-001/acceptance.md` — 13 binary ACs (2 guards green-today, 1 core RED-now, 5 option-gated, robustness + PRESERVE + boundary)
- `.moai/specs/SPEC-UPDATE-HOOK-DELIVERY-001/design.md` — the OPEN decision axes
- `.moai/specs/SPEC-UPDATE-HOOK-DELIVERY-001/research.md` — verified code evidence
- `.moai/specs/SPEC-UPDATE-HOOK-DELIVERY-001/progress.md` — §E.1 populated; §E.2-§E.4 placeholders

## Verification performed (this run, this tree, d592b0551)

- `internal/cli/update/merge/base.go:112-131` — pruneToShared map-only recursion read verbatim: matches dispatch finding 1.
- `internal/merge/strategies.go:427-429` — "Only user changed" case read verbatim; `valuesEqual` at `:685-692` JSON-compares (arrays deep-equal) — consequence chain strict: matches finding 2.
- `internal/cli/doctor.go:768-783` — checkHooksConfig is os.Stat-only on `.claude/hooks/`, never opens settings.json: matches finding 4.
- `grep -c '^func Test' internal/template/settings_test.go` → `31`: matches finding 5.
- SPEC ID regex check → `PASS` (verbatim output above in transcript).

## The open decision (one paragraph)

The card's core question — what `moai update` should do when the template gains a hook entry inside an event key the user already carries — is left OPEN for the Implementation Kickoff Approval gate. Option A (deliver) closes the capability gap automatically but requires an entry-identity scheme to avoid resurrecting user-deleted entries; design.md frames three schemes (A1 stable identity key in-file, A2 in-file tombstones — durability under Claude Code's own rewrites unmeasured, A3 seen-registry sidecar — coupled to the `.moai/state/` wipe in `CleanMoaiManagedPaths`). Option B (detect + guide) makes doctor/update report missing entries read-only — near-zero risk, but drift keeps growing and headless update flows never read the report, so the capability stays inactive. Option C (explicit no-op + docs) converts the silent drop into a documented one at zero code risk but lets drift grow indefinitely and requires the doc statement to track every future template hook change. All options are bound by the invariant REQs: template-new event keys keep delivering, user deletions are never resurrected, user-modified entries survive, JSON validity and idempotence hold.

## Gaps (not verified)

- Claude Code's own settings.json rewrite durability (bears on Option A's in-file schemes) — unmeasured; flagged as a run-phase measurement if A is selected.
- golangci-lint / test baselines on this tree — deferred to run-phase M1 pre-flight per plan.md §C (no code changes at plan phase).
- The t216 investigation report was NOT accessible from this tree (another lane's worktree) and is not cited anywhere in the SPEC artifacts.
- The RED-now state of AC-UHD-003 is established from the code chain, not from an executed update run against a fixture project — the mechanical RED execution is M2's first deliverable per acceptance.md §1.

## Readiness

The SPEC is ready for plan-auditor: GEARS notation throughout, Out of Scope lint convention satisfied, frontmatter schema-conformant, artifact set complete for the declared Tier M + design-decision card.

## Delta fixes (post plan-audit PASS-WITH-DEBT 0.86, spec v0.2.0)

All six audit findings addressed in one pass — D1 (§D invariant range → 001..006, 011..012), D2 (AC-UHD-013 added for REQ-UHD-009; counts synced to 13 in spec.md §D, plan.md §E, this file), D3 (plan.md §E gained the E8 RED-evidence row), D4 (REQ-007 "not previously seen" disambiguated with a trailing clause), D5 (design.md §A quote corrected to a labeled paraphrase of REQ-004), D6 (AC-003 adoption gate naming the four mechanical RED elements). The open-decision structure (options A/B/C, gated at Implementation Kickoff Approval) is unchanged.
