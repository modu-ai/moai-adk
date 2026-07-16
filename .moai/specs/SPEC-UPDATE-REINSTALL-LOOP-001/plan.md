# Implementation Plan — SPEC-UPDATE-REINSTALL-LOOP-001

> Milestones are ordered by **decision-reversibility** — the decisions most likely to change
> (product/UX/data-flow) lead so human review focuses there; the mechanical, low-change-likelihood
> loop-break is last. This is a REVIEW ordering, not a strict execution order — see §F dependencies.

## §A Context

`moai update` re-runs the v2→v3 clean reinstall on every invocation because one path,
`.claude/rules/moai/design`, is both a `DeprecatedPaths` entry and a v3-template-shipped
directory (research.md §A). The existing v3-version override only suppresses the loop for
`v3.`-prefixed version strings; the collision itself is the durable fuel. As a side effect,
the clean-reinstall path overwrites user `settings.json` and `user.yaml` (research.md §C.1)
because it bypasses the normal path's 3-way merge.

Subsystems touched: `internal/defs` (registry + guard), `internal/cli` (clean-reinstall path),
`internal/template` (settings template pin). No net-new architecture → no design.md.

## §B Known Issues / Resolved Clarifications

All three plan-audit iter-1 open questions were resolved via orchestrator AskUserQuestion on
2026-07-17. The resolutions are recorded below (the former clarification markers are
retired).

- **RESOLVED — model pin policy → option (b) merge-preserving.** `settings.json.tmpl` pins
  `"model": "sonnet"` (research.md §C.3); on the clean-reinstall force-deploy this silently
  downgrades an Opus/higher-tier project. **Decision:** KEEP the template pin, but R3 makes the
  clean-reinstall's `settings.json` handling merge-preserving so an existing user's `model` always
  wins; the pin applies only to genuinely new projects. No separate template-pin removal — R3's
  merge-preservation subsumes the concern (the former model-pin milestone is folded into M1 config-preservation; see §F). Reflected in spec.md
  REQ-RIL-010 (§B.4, now "RESOLVED — merge-preserving").

- **RESOLVED — init system.yaml version format → informational only.** The `moai init`
  generator's `moai.version` string format (research.md §D) was NOT traced and the scope is NOT
  expanded to trace it. R1's collision removal is version-format-independent, so the loop is broken
  regardless of the init version format. Recorded as an out-of-scope informational note in spec.md
  §C; no acceptance criterion depends on it.

- **RESOLVED — R2 (pre-flight hardening) → SPLIT to follow-up SPEC-UPDATE-PREFLIGHT-SAFETY-001.**
  research.md §C.2 shows the current 7-step order already fails closed before destruction for both
  triggers (symlink / nested-`.git`), so R2's residual value is hardening (explicit actionable
  error + rollback invariant), not a Critical defect. **Decision:** SPLIT R2 (M3, REQ-RIL-005/006,
  AC-RIL-008/009) to the follow-up **`SPEC-UPDATE-PREFLIGHT-SAFETY-001`**. This SPEC's scope is
  R1 (collision removal) + R3 (config preservation) only. Reflected in spec.md §B.3 + §C Exclusions.

- **EXTENDED — config-preservation scope (R3) → ALL `.moai/config/sections/*.yaml`.** Issue #1084
  explicitly reports loss of `language.yaml` and `design.yaml`, not only `user.yaml`. REQ-RIL-008
  and AC-RIL-006 are broadened to cover the whole `sections/*.yaml` set (see spec.md §B.2 REQ-RIL-008
  and acceptance.md AC-RIL-006).

## §C Pre-flight (before run-phase)

- (All three clarifications resolved 2026-07-17 — see §B; no open markers remain.)
- Confirm `development_mode` (DDD vs TDD) from quality.yaml for the run-phase cycle.
- Re-run the collision intersection at run-start to confirm still exactly 1 collision (guards
  against drift between plan and run).

## §D Constraints

- **Template-First (CLAUDE.local.md §2)**: with the model-pin resolved as merge-preserving
  (option b), NO `internal/template/templates/` edit is expected — the pin stays; the user model
  survives via M1's settings merge. Should any template edit nonetheless arise, it goes to the
  template source first, then `make build`.
- **Atomic count update (dirs.go @MX:REASON)**: the `DeprecatedPaths` slice edit MUST update
  `internal/defs/dirs_test.go` (`TestDeprecatedPathsTotalCount` 41→40,
  `TestDeprecatedPathsCategorySplit` B 29→28) AND the count recited in the `dirs.go` @MX:REASON
  comment and the `v2_detection.go` header comment — all in the same change.
- **No behavior change to the normal update path** — only close the clean-reinstall bypass.
- **16-language template neutrality** — the settings template edit must stay language-neutral.

## §E Self-Verification (deliverables — run-phase)

Run-phase §E matrix (manager-develop owns E1-E7 population): AC-RIL matrix pass/fail;
cross-platform build (`GOOS=windows` included per internal/cli CLAUDE.md); coverage on
`internal/defs` + `internal/cli`; `golangci-lint`; the collision guard as the headline evidence.
The plan-phase audit-ready signal is emitted to progress.md §E.1 by the orchestrator after audit.

## §F Milestones (ordered by decision-reversibility — highest-change-likelihood first)

> Post-clarification the SPEC has **two** milestones. The former M1 (model-pin policy) collapses
> into M1 below — the resolved merge-preserving policy (option b) requires NO separate template
> change; the model survives via the same settings merge. The former M3 (R2 pre-flight hardening)
> is SPLIT to `SPEC-UPDATE-PREFLIGHT-SAFETY-001` and removed from this plan (REQ-RIL-005/006 +
> AC-RIL-008/009 carry no DoD weight here).

### M1 — Config preservation on clean-reinstall (R3, incl. model-pin merge-preservation) — data-flow decision
- Decision: EITHER (i) add `.claude/settings.json` + ALL `.moai/config/sections/*.yaml` to the
  clean-reinstall PRESERVE inventory (extend `preserveInventoryRoots` / a dedicated config-preserve
  step), OR (ii) route the clean-reinstall's settings/config write through the same 3-way merge the
  normal path uses (`collectMergeableFiles` / `restoreMoaiConfig`) instead of the forceUpdate deploy.
  Recommended: (ii) — reuses the proven merge, satisfies REQ-RIL-009 by construction, and subsumes
  the model-pin concern (the resolved option-b merge-preservation: a preserved user `model` in
  `settings.json` always wins over the template `"model": "sonnet"` pin).
- Scope note (D2/D4): the config-preservation set is ALL `sections/*.yaml` (user.yaml, language.yaml,
  design.yaml, and every other section file) — issue #1084 reports language.yaml/design.yaml loss,
  not only user.yaml. Whichever approach is chosen MUST cover the whole `sections/*.yaml` set.
- Implement + tests for AC-RIL-005/006/007 and the AC-RIL-010 model-survives check (which is now a
  by-product of the settings merge, not a separate template edit).
- Priority: High. Moderate reversibility (touches the clean-reinstall data flow).

### M2 — Break the loop + regression guard (R1) — LOWEST change likelihood, CRITICAL priority
- Remove the `.claude/rules/moai/design` entry from `defs.DeprecatedPaths` (Category B).
- Update `dirs_test.go` count/category guards atomically (total 41→40, B 29→28) + the two
  source-comment counts.
- Add the collision-guard test: assert `defs.DeprecatedPaths ∩ embedded-template-FS = ∅`
  (walk the embedded FS or `fs.Stat` each entry), with a negative-path fixture proving it FAILS
  when a colliding entry is present (REQ-RIL-003, AC-RIL-001/002/004).
- Add the behavioral zero-net-change test: two consecutive update runs on a v3 fixture remove 0
  deprecated paths and leave the tree stable (AC-RIL-003).
- Priority: **Critical**. Mechanical + surgical; executes independently of M1 and could ship
  first as a hotfix even though it is reviewed last.

> **Execution vs review order**: M2 is the Critical hotfix and may execute first; it is placed last
> here only because it is the most mechanical / least-likely-to-change decision, per the
> reversibility review-ordering rule. M1 carries the data-flow decision that warrants reviewer attention.

## §G Anti-Patterns to avoid

- Tightening the v3-version override instead of removing the collision (band-aid, not the fix).
- Editing the template `constitution.md` or deleting `.claude/rules/moai/design/` — the template
  is SUPPOSED to ship it; the DeprecatedPaths ENTRY is the stale artifact.
- Changing the `DeprecatedPaths` slice without atomically updating the count guards (CI fail).
- Satisfying AC-RIL-002/003 with a token-presence grep instead of a real executed test / behavioral run.
- Blanket-adding settings.json to PRESERVE without confirming it does not double-write against the
  deploy (verify against the normal-path merge outcome — AC-RIL-007).

## §H Cross-References

- research.md — verified anchors, intersection result, R2 refutation.
- `SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001/-002` — clean-reinstall path + existing v3 override (do not regress).
  Note: `SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002` frontmatter is `status: draft`, yet its REQ-CRR-001
  v3-version negative-override code IS live in the tree (`internal/cli/v2_detection.go:~199`,
  `strings.HasPrefix(v, "v3.")` branch). Treat the override as present/functional behavior (rely on
  the code, not the SPEC status) when reasoning about the loop; do not assume it is absent because
  the SPEC is unclosed.
- `SPEC-DEPRECATEDPATHS-RECONCILE-001` — the 41-entry count contract this SPEC decrements.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — status/ownership matrix.
