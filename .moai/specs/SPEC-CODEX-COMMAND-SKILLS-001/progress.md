# progress.md — SPEC-CODEX-COMMAND-SKILLS-001

Card: t503 · Branch `WT-codex-command-skills` · Base `ace1c5440` (`origin/develop`) · Tier M

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-07
plan_phase_artifacts: spec.md, plan.md, acceptance.md (Tier M set; progress.md carried per canonical skeleton)
measured_baseline: 16 command sources / 34 canonical skills / 0 name collisions on both bare and moai- prefixed derivations (this tree, 2026-09-07)

## §E.2 Run-phase Evidence

Milestones M1-M4 complete. Runs in worktree `.claude/worktrees/t503`, branch `WT-codex-command-skills`.

- M1 (commit e7d2a1658): emitter package `internal/template/commandemit` + 16 committed golden artifacts under `templates/.agents/skills/` + gitignore re-inclusion layers (root `.gitignore` re-includes the template `.agents/` subtree per the `.codex/` precedent; `templates/.gitignore` re-includes the 16 published names — the `.agents/skills/moai*` mirror rule would otherwise ignore them, which the plan did not anticipate). AC-005 RED observed before the collision guard landed (rc=1, `Emit succeeded over a colliding name; want refusal diagnostic`).
- M2 (commit 6ae337e60): `make commands-emit` / `commands-emit-check`, check wired ahead of `build`. AC-010 RED observed on a hand-mutated artifact (rc non-zero + drift diagnostic), GREEN after `make commands-emit`; check-only run leaves the tree unchanged. Measured nuance: the recipe exits 1; `make` itself surfaces rc 2 on target failure — identical to the shipped `agents-emit-check` behavior.
- M3 (commit 082c2c04f): `deployer.go` provenance check survives `forceUpdate` for the exact published-name set; skips reported via `DeployResult.ProtectedSkips`; template-managed published entries still overwrite in update mode. `TestSkillMirror_SlimSetEqualsCanonicalAndIsSmaller` re-expressed for the new namespace reality (mirror partition vs published occupants), invariant preserved.
- M4: boundary flag in every emission report entry; `CLAUDE.local.md` §2 pointer (D5); neutrality audit green; AC-013 both builds exit 0.

Key decisions / deviations (each attributable to the SPEC envelope):
- Emitter package named `commandemit`; env switch `COMMAND_EMIT_UPDATE`; make targets `commands-emit` / `commands-emit-check` (names were M2-run-phase choices per plan.md D4).
- Generated provenance header placed inside the YAML frontmatter as a `#` comment (plan.md D3).
- `skill_mirror.go` received ONE additive edit: the `ProtectedSkips` field on `DeployResult` (struct defined there; zero behavior change to mirror logic). Recorded deviation from the §A.5 PRESERVE listing — the field is the R-011 "report the skip" carrier.
- gitignore two-layer change (root + `templates/.gitignore`): unanticipated mechanical blocker — committed golden artifacts (R-007) are otherwise unaddable. Guarded by `TestGitignoreCarriesEveryPublishedName` (emitter side) and the exact-name `publishedSkillNames` set + `TestPublishedSkillsNamesMatchTree` (deploy side).
- t497 boundary: every published body retains the verbatim `Use Skill("moai")` line (16/16 grep). Nothing repaired, nothing annotated in bodies.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-09-07
run_commit_sha: 082c2c04f (M3; M4 docs/progress commit follows)
run_status: complete
ac_pass_count: 13
ac_fail_count: 0
preserve_list_post_run_count: 0 (no byte changed under §A.5: command sources, codexwiring, agentemit, `.claude/commands/moai` local copy — verified via `git status --porcelain` empty on the source dirs after emission runs)
l44_pre_commit_fetch: n/a (worktree branch, no push by lane)
l44_post_push_fetch: n/a (push is lead's batch action)
new_warnings_or_lints_introduced: 0 (`golangci-lint run ./internal/template/...` → 0 issues)
cross_platform_build: host exit 0; GOOS=windows GOARCH=amd64 exit 0
total_run_phase_files: 27 (6 emitter pkg + 16 artifacts + 2 gitignore + Makefile + deployer.go + skill_mirror.go + published_skills.go + tests)
m1_to_mn_commit_strategy: per-milestone commits (M1 e7d2a1658, M2 6ae337e60, M3 082c2c04f; M4 docs+progress commit follows)

## §E.4 Sync-phase Audit-Ready Signal

sync_status: complete
sync_commit_sha: 9da07ac65 (backfilled — sync commit SHA)
b12_self_test_a: SPEC-ID pre-emission grep count = 0 (safe to emit; rc=1, no prior entry)
b12_self_test_b: AC count match — acceptance.md distinct AC identifiers = 13, CHANGELOG entry references the 13-AC close (run report: ac_pass_count 13 / ac_fail_count 0, read from .moai/reports/t503/run-evidence.md)
b12_self_test_c: file-path verification — CHANGELOG-named paths (`internal/template/commandemit`, `.agents/skills/moai-<command>/SKILL.md`, `templates/.gitignore`, root `.gitignore`) confirmed present on this tree
changelog_entry_position: [Unreleased] ### Added, first bullet
frontmatter_status_transitions: spec.md in-progress → completed (merged 3-phase close; updated: 2026-09-07 unchanged); plan.md / acceptance.md stateless on the status axis per spec-frontmatter-schema.md § Artifact Statelessness — no status field to transition
canary_compliance_check: n/a — this SPEC defines no forward-looking policy its own sync tests
mx_tag_validation: 16/16 published bodies retain the verbatim `Use Skill("moai")` boundary line (run-phase verified); no @MX tag repair needed on the sync surface
user_facing_docs_note: DEFERRED — broader user-guide / docs-site documentation of the codex skill publication surface is out of this phase's scope; the SPEC's own doc requirement (R-007 / D5, CLAUDE.local.md §2 pointer) landed in M4. docs-site 4-locale + hugo-build discipline is a separate surface (per card t503 dispatch, docs-site/ not touched in this phase)
