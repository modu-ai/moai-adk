# Implementation Plan — SPEC-MERGE-METHOD-CONFIG-001

> Plan-phase artifact. Tier M. Do NOT implement — GATE-2 (plan→run human gate) is a separate control point.

## §A. Context

GitHub Issue #1061: the sync-phase PR merge method is hardcoded to `gh pr merge --squash` in agent/template prose and is not configurable, despite a complete-but-unused `merge`/`squash`/`rebase` abstraction in `internal/github` and a FROZEN rule that fixes squash for all 3 SPEC phases.

This plan adds a per-mode `merge_method` config field (default `squash`, preserving current behavior), wires the Go READ path (mirroring the established `git_strategy.<mode>.hooks.pre_push` pattern from the PREPUSH dead-config line), threads the config value into the sync-agent prose so `gh pr merge` honors it, and reconciles the FROZEN lifecycle table from a hardcoded `squash` literal to a configured-default formulation.

## §B. Known Issues / Constraints discovered during planning

1. **FROZEN-zone edit** — `spec-workflow.md` lifecycle table is `[ZONE:Frozen] [HARD]`. The amendment is non-destructive (default stays squash) but still requires explicit GATE-2 acknowledgment. This is the single highest-risk item; surfaced prominently to the orchestrator. (spec.md §C.1)
2. **SSOT mirror parity** — `spec-workflow.md` exists in both `.claude/rules/.../` and `internal/template/templates/.claude/rules/.../`. `embedded_mirror_test.go` enforces byte-identity; BOTH copies MUST be edited in the same commit. (internal/template/CLAUDE.md § Mirror parity checks)
3. **Consumer is agent prose** — the `gh pr merge` consumer is the sync agent, not Go code. REQ-MMC-007/008/009 are verified by grep/structural assertions on rendered templates, not Go behavior tests. (spec.md §C.2)
4. **Template neutrality** — all `internal/template/templates/` edits go through CI guard `template-neutrality-check.yaml`; no SPEC IDs/dates/SHAs in template content. (CLAUDE.local.md §15/§25)
5. **`make build` required** — template source changes must be followed by `make build` to regenerate `go:embed` assets (or confirm `go:embed all:templates` makes regeneration unnecessary — verify at run-phase per recent embedded.go retirement noted in memory).

## §C. Pre-flight verification (run at start of run-phase)

```bash
# 1. git_strategy loader is wired (READ path active)
grep -n 'loadGitStrategySection' internal/config/loader.go
# 2. ModeProfile has Hooks but NOT yet MergeMethod
grep -n 'MergeMethod\|Hooks ' internal/config/types.go
# 3. zero production callers of PRMerger (confirm EX-1 boundary holds)
grep -rln 'PRMerger\|PRMerge\|MergeOptions' internal/ | grep -v '_test.go' | grep -v 'internal/github/'
# 4. hardcoded --squash sites (confirm 6 prose sites unchanged at run-phase start)
grep -rn 'gh pr merge --squash' internal/template/templates/.claude/
# 5. FROZEN lifecycle table squash cells
grep -n 'squash' internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md
```

## §D. Constraints

- Default merge method MUST be `squash` (REQ-MMC-002, EX-4) — zero behavior change without opt-in.
- `squash`-resolved rendered command MUST be byte-equivalent to current (`gh pr merge --squash --delete-branch`, REQ-MMC-009).
- Go-side change mirrors the existing `HooksConfig` / `loadGitStrategySection` pattern (no new loader machinery, just a new field on `ModeProfile`).
- No FROZEN rationale removal (EX-7).
- Tier M: per-mode field only; no per-branch-type override (EX-2), no PRMerger wiring (EX-1).

## §E. Self-Verification (plan-phase author checklist)

- [x] All 12 frontmatter fields present, canonical names (`created`/`updated`/`tags`, not snake_case aliases)
- [x] `id` matches `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (decomposition printed: SPEC ✓ | MERGE ✓ | METHOD ✓ | CONFIG ✓ | 001 ✓ → PASS)
- [x] `status: draft` initial
- [x] Exclusions section present (7 entries)
- [x] No implementation code in spec.md (WHAT/WHY only)
- [x] GEARS requirements (Ubiquitous/Event-driven/State-driven/Event-detected forms used)
- [x] FROZEN-rule risk surfaced prominently (spec.md §C.1, this plan §B.1)

## §F. Milestones

Tier M, sequential. Run-phase only (post-GATE-2).

### M1 — Go config field + default (internal/config)
- Add `MergeMethod string \`yaml:"merge_method"\`` to `ModeProfile` (types.go), with a godoc note mirroring the `HooksConfig` forward-compat style.
- Add `MergeMethod: "squash"` to all 3 mode profiles in `NewDefaultGitStrategyConfig()` (defaults.go).
- Covers: REQ-MMC-001, REQ-MMC-002.
- Tests: extend `types_test.go` / `defaults_test.go` to assert the field + default.

### M2 — Loader READ-path coverage (internal/config)
- No new loader function needed — `loadGitStrategySection` already unmarshals the whole `ModeProfile`; the new field rides the existing partial-override contract.
- Covers: REQ-MMC-003, REQ-MMC-004.
- Tests: extend `git_strategy_loader_test.go` with: (a) file omits `merge_method` → default `squash` retained; (b) file sets `merge_method: merge` → file value populated; (c) partial override (one mode set, siblings keep default).

### M3 — Validation enum (internal/config)
- Add `merge_method` enum validation (`{squash, merge, rebase}`) in validation.go, alongside the existing `git_strategy.<mode>.hooks.*` checks. Empty → treated as default (no error).
- Covers: REQ-MMC-005, REQ-MMC-006.
- Tests: validation test for invalid value (error names field path) + empty value (no error).

### M4 — Template config key (internal/template/templates)
- Add `merge_method: squash` under each mode's profile in `git-strategy.yaml.tmpl` (manual/personal/team), with a neutral inline comment (`# squash, merge, rebase`).
- Covers: REQ-MMC-002 (template default surface), REQ-MMC-012 (neutrality).
- Verify: `go test ./internal/template/... -run TestTemplateNeutralityAudit`.

### M5 — Sync-agent prose rewiring (internal/template/templates)
- `sync/delivery.md` (lines ~313, ~325): replace hardcoded `gh pr merge --squash --delete-branch` with method-selection prose driven by the active mode's `merge_method` (default → `--squash`).
- `manager-git.md` (lines ~140, ~176, ~204): same.
- `moai-ref-git-workflow/skill.md`: align the squash/merge/rebase guidance table to reference the configurable field (the rows already document all 3 methods; add a pointer that the method is config-driven).
- Covers: REQ-MMC-007, REQ-MMC-008, REQ-MMC-009, REQ-MMC-012.
- Verify: grep that zero `gh pr merge --squash` literals remain UNCONDITIONALLY hardcoded (each must be inside a default/conditional formulation); `squash`-default rendered command preserved byte-equivalent.

### M6 — FROZEN lifecycle table amendment (SSOT + mirror, LAST)
- `spec-workflow.md` lifecycle table: change the `Merge`/`PR strategy` column from literal `squash` to "configured `merge_method` (default `squash`)" for all 3 phase rows, preserving the FROZEN rationale prose.
- Edit BOTH `.claude/rules/.../spec-workflow.md` AND `internal/template/templates/.claude/rules/.../spec-workflow.md` in the same commit (mirror parity).
- Covers: REQ-MMC-010, REQ-MMC-011.
- Verify: `embedded_mirror_test.go` passes (byte-identity); FROZEN rationale text still present.
- [HARD] This milestone is the FROZEN-zone edit — it proceeds ONLY after GATE-2 approval already obtained at run-phase entry.

## §G. Anti-Patterns to avoid

- AP-1: Changing the default to `merge` (PRMerger's default) — violates EX-4; default MUST stay `squash`.
- AP-2: Wiring PRMerger to a real caller "while we're here" — out of scope (EX-1, scope discipline).
- AP-3: Adding per-branch-type overrides — out of scope (EX-2).
- AP-4: Editing only one copy of the mirrored `spec-workflow.md` — breaks `embedded_mirror_test.go`.
- AP-5: Leaking SPEC ID / date / SHA into template content — fails neutrality CI guard.
- AP-6: Removing the FROZEN squash rationale instead of widening it (EX-7).
- AP-7: Asserting REQ-MMC-007/008/009 via Go behavior tests — those are prose requirements verified by grep/structural assertion (spec.md §C.2).

## §H. Cross-References

- GitHub Issue #1061
- spec.md §A.2 (verified ground-truth table), §C.1 (FROZEN risk)
- PREPUSH dead-config line precedent: `SPEC-PREPUSH-WIRING-001`..`-SAVE-WIRING-001` (same git_strategy section, hooks.pre_push field — the established wiring pattern)
- `internal/config/CLAUDE.md` (section-file layout, loader/validation/defaults conventions)
- `internal/template/CLAUDE.md` (mirror parity, neutrality, `make build`)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` (Status Transition Ownership Matrix — `(none) → draft` owned by manager-spec)
