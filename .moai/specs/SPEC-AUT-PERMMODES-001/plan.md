# SPEC-AUT-PERMMODES-001 — Implementation Plan

> Card **t584**. Tier M, 3 milestones, priority-ordered (no time estimates). Worktree: `.claude/worktrees/t584`, branch `WT-autonomy-perm-modes`.

## §A. Context

The init wizard's autonomy question (post-t586 "Agents & Autonomy" page, `internal/cli/wizard/questions.go`) currently speaks MoAI tier vocabulary and maps `semi-auto` → `defaultMode: "default"`. The operator redefined the question as Claude Code's real permission modes with **acceptEdits as the default** (card source: `.moai/reports/init-tui-audit-20260909.md`, card C2). Plan-phase research settled the token question: `auto` IS a valid `permissions.defaultMode` value in current CC (six-value enum; see spec.md §A Evidence) — the repo's bundled IAM reference claiming "exactly four values" is stale and is refreshed in scope.

## §B. Known Issues

1. **Stale bundled reference** — `internal/template/templates/.claude/skills/moai-foundation-cc/reference/claude-code-iam-official.md:30-40` asserts a four-value enum and explicitly denies `auto`. This is the root cause of the card's token doubt; REQ-007 refreshes it.
2. **Zero-delta tests now assert the wrong invariant** — `TestApplyAutonomyTierBundle_SemiAutoIsZeroDelta` (autonomy_bundle_test.go:89), `TestApplyAutonomyTierBundle_EmptyIsZeroDelta` (:115), `TestRunInit_SemiAutoAndEmptyAreZeroDelta` (init_autonomy_wiring_test.go:193) all assert "no file written". Under the re-scoped REQ-004 they must assert "only USER defaultMode=acceptEdits written, everything else byte-identical".
3. **`auto` availability floor unverified** — current docs confirm the token but the fetched pages do not state which CC version introduced `defaultMode: "auto"`. M1 pins this before M2 lands the mapping.
4. **Doc-comment drift** — `autonomy_bundle.go` header and `TierDefaultMode` godoc describe the old mapping and old REQ-007 wording.

## §C. Pre-flight (run-phase entry checks)

- Confirm t586's wizard structure is the base being edited: autonomy_tier question lives in `internal/cli/wizard/questions.go` (Group "Agents & Autonomy"), translation keys `autonomy_tier` at translations.go:98/184/270 (ko/en/ja/zh blocks).
- Confirm the web tier-toggle surface state on this tree: `internal/web/handlers.go:203` and `internal/web/app.go:193` reference `ApplyAutonomyTierBundle`; verify whether the web surface is retired (app.go comment says "only the web surface is gone") before assuming any web edit. Expected outcome: NO web edit (Out of Scope), but the check prevents a surprise compile break.
- Verify `go test ./internal/cli/wizard/... ./internal/config/... ./internal/core/project/... ./internal/cli/...` (affected packages only, per repo test discipline) is green BEFORE the first edit, so regressions are attributable.

## §D. Constraints

- Do NOT re-expand the question count (t583/t586 structure is the shipped baseline; this card reshapes ONE question's options + its bundle mapping).
- Persisted tier tokens (`semi-auto` / `automatic` / `fully-autonomous`) MUST NOT change — no config migration, `ValidateAutonomyTierSelection` closed set unchanged.
- The downgrade path and its regression tests MUST be preserved (card HARD requirement).
- Template edits go to `internal/template/templates/` first (Template-First rule); `make build` before local verification.
- Code comments English; wizard user-facing strings localized across ko/en/ja/zh.
- No `git add -A`; stage explicit paths. No push from this card.

## §E. Self-Verification (run-phase §E evidence hooks)

- E1: affected-package tests green (commands + verbatim output).
- E2: `make build` succeeds after template edits (IAM reference refresh).
- E3: translation completeness test green (`translations_completeness_test.go`).
- E4: grep guard — no residual `"Semi-auto (Recommended)"` / `"Prompt before each non-trivial action"` strings in wizard sources; no residual `→ "default"` mapping comment for semi-auto.
- E5: downgrade advisory still lands in `.moai/logs/autonomy-downgrade.log` (wiring test observed output).

## §F. Milestones

### M1 — Token verification + version floor (Priority: High)

1. Pin the CC version that introduced `defaultMode: "auto"` (official docs / release notes sweep) and record it in progress.md; verify the local CC build reads it (settings round-trip in a temp project).
2. Confirm `permissions.disableAutoMode` semantics from official docs; record as Out-of-Scope context (no wiring this card).
3. Re-read the bundled IAM reference and enumerate exactly which lines REQ-007 must change.

**Exit**: version floor recorded; reference-doc edit list written. M2 MUST NOT start before M1 exits — if the floor research contradicts the six-value enum (e.g., `auto` is gated behind a plan tier), STOP and return a blocker report.

### M2 — Question options + bundle mapping (Priority: High)

1. `internal/cli/wizard/questions.go` — autonomy_tier options re-labeled per REQ-001/REQ-002 (labels + descriptions; values and `Default: semi-auto` unchanged).
2. `internal/config/autonomy_tiers.go` — `TierDefaultMode`: `semi-auto` → `"acceptEdits"`; unknown → `"default"` retained; godoc updated (REQ-003, REQ-010).
3. `internal/core/project/autonomy_bundle.go` — behavior table + REQ-004 re-scoped wording in godoc; the semi-auto early-return replaced by the bounded write (REQ-004).
4. `internal/cli/wizard/translations.go` — 4-locale labels/descs (REQ-008).
5. `internal/cli/init.go:126` — flag help text (REQ-009).
6. `internal/template/templates/.claude/skills/moai-foundation-cc/reference/claude-code-iam-official.md` — six-value enum + kill-switch pair + version notes (REQ-007); `make build`.

### M3 — Tests + guards (Priority: Medium)

1. Adapt the three zero-delta tests to the bounded-delta invariant (REQ-004): assert USER file carries exactly `defaultMode: "acceptEdits"` and PROJECT file byte-identical.
2. Preserve the downgrade regression set verbatim in intent (REQ-006); adjust only assertions that name the old mapping.
3. Keep `TestAutonomyTierQuestion_FullyAutonomousNotRecommended` green — extend it to also assert the Recommended label sits on the acceptEdits option.
4. Run affected packages; record §E evidence.

## §G. Anti-Patterns

- Silently deleting the zero-delta tests instead of re-scoping them (loses the REQ-004 guard).
- Changing tier tokens or adding a settings migration (scope creep).
- Hand-editing the `.claude/` mirror of the IAM reference instead of the template source.
- Mapping unknown tiers to `acceptEdits` (fail-safe must stay at the MOST restrictive mode).

## §H. Cross-References

- SPEC body: `./spec.md` · AC: `./acceptance.md` · Owning SPEC of amended invariants: `.moai/specs/SPEC-AUTONOMY-TIERS-001/spec.md`
- Card dispatch + verdict home: lead session, card t584, evidence path `.moai/specs/SPEC-AUT-PERMMODES-001/progress.md`
