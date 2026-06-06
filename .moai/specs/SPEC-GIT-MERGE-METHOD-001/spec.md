---
id: SPEC-GIT-MERGE-METHOD-001
title: "Configurable sync-phase PR merge method (squash / merge / rebase)"
version: "0.1.0"
status: draft
created: 2026-06-06
updated: 2026-06-06
author: claude (from issue #1061 by michaelleone)
priority: P2
phase: "v3.0.0"
module: "internal/template/templates/.claude/skills/moai/workflows/sync + internal/template/templates/.claude/agents/moai/manager-git.md + internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md + internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl + internal/github"
lifecycle: spec-anchored
tags: "config, git-strategy, sync-phase, pr-merge, manager-git, frozen-rule-relaxation"
tier: S
issue_number: 1061
---

# SPEC-GIT-MERGE-METHOD-001 — Configurable Sync-phase PR Merge Method

## HISTORY

| Version | Date | Author | Notes |
|---------|------|--------|-------|
| 0.1.0 | 2026-06-06 | claude (issue #1061) | Initial draft authored from GitHub issue #1061 by michaelleone. Tier S (config field + thin template/agent edits + a one-cell relaxation of a FROZEN rule + optional wire-through to existing tested abstraction). AC inline §3. |

---

## 1. 개요 (Overview)

### 1.1 Mission Statement

The `/moai sync` auto-merge step is currently hardcoded to `gh pr merge --squash --delete-branch` in two workflow surfaces (the sync delivery skill body and the `manager-git` agent body) and is additionally pinned by a `[ZONE:Frozen] [HARD]` rule in `spec-workflow.md` whose lifecycle table fixes `squash` for all three phases. The MoAI binary already ships a complete, tested `merge`/`squash`/`rebase` abstraction (`internal/github/gh.go` `PRMerge` + `internal/github/pr_merger.go`) with **zero production callers** — only test code references it.

This SPEC introduces a single configuration key, `git_strategy.<mode>.merge_method`, in `.moai/config/sections/git-strategy.yaml`, defaulting to `squash` for backward compatibility, and threads it through the two hardcoded surfaces. As a non-blocking follow-up, the wire-through MAY route through the already-existing `PRMerger`. The FROZEN-rule cell that currently reads `squash` becomes `configurable (default: squash)` so the table remains the SSOT for phase strategy without pinning a single literal value.

The SPEC also documents the conscious choice between (a) a single global merge method per mode and (b) per-branch-type overrides (e.g. gitflow `release/*` → `merge` per MoAI's own `moai-ref-git-workflow/skill.md:133-134,171`). Option (a) is in scope and the default; option (b) is documented as a follow-up extension shape rather than implemented here, to keep this SPEC Tier S.

### 1.2 Background — verifiable evidence from issue #1061

The reporter audited `modu-ai/moai-adk@main` at commit `5714bae` and identified the following five concrete locations. All five were re-verified against the working tree before this SPEC was drafted:

1. **Hardcoded in sync skill**: `internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md:313` and `:325` (Step 3.4 Auto-Merge) — both literal `gh pr merge --squash --delete-branch`.
2. **Reinforced in manager-git agent body**: `internal/template/templates/.claude/agents/moai/manager-git.md:140,176,204` — all three references hardcode `--squash --delete-branch`. (Issue body cited :176/:204; the third location at :140 is documented here as part of the M-prime sweep set.)
3. **FROZEN rule pins squash for all 3 phases**: `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md:25` is `[ZONE:Frozen] [HARD]` and lines 29-31 (the lifecycle table) fix the `PR strategy` column to `squash` for plan/run/sync.
4. **No `merge_method` key in any config section**: `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` configures workflow type, automation toggles, branch prefixes, `main_branch`, `draft_pr`, `required_reviews`, `branch_protection`, hooks, and commit style — but no `merge_method` / `merge_strategy` key exists in any `.moai/config/sections/*.yaml`.
5. **Existing abstraction unused in production**: `internal/github/gh.go:49-55` defines all three method constants and `internal/github/gh.go:241` translates each to the right `gh` flag via `PRMerge(ctx, number, method, deleteBranch)`. `internal/github/pr_merger.go` is a complete conditional merger and its **default method is `merge`, not `squash`** (`pr_merger.go:134`). The reporter's grep count (30 test refs in `pr_merger_test.go`, 17 in `gh_client_test.go`, 5 in `integration_test.go`, 7 + 4 in the definitions themselves) is correct — there are **zero production callers**; every reference is test-only. The shipped `/moai sync` flow bypasses this entirely.

### 1.3 Inconsistency with project's own reference docs

`delivery.md:252` (Step 3.2) routes `gitflow` `release/*` and `hotfix/*` branches to PRs against `main`, and Step 3.4 then unconditionally squashes them. This contradicts MoAI's own reference documentation in `internal/template/templates/.claude/skills/moai-ref-git-workflow/skill.md:133-134`:

```
| Squash merge | Feature branches (clean history)  | gh pr merge --squash |
| Merge commit | Release branches (preserve history) | gh pr merge --merge  |
```

(`skill.md:171` further argues merge commits preserve narrative for release history.) In gitflow mode there is currently no way to honor this — release PRs get squashed regardless. This SPEC removes the global block by making the merge method configurable; per-branch-type overrides for the gitflow case are scoped as a follow-up §1.6.

### 1.4 Scope discrimination — not a bug

`internal/cli/worktree/sync.go:29`'s `--strategy merge|rebase` flag is for syncing a worktree against its base branch (a different operation; not the PR merge method). This SPEC does not touch worktree sync; the two surfaces happen to share the word "strategy" but are unrelated. The new `merge_method` key is named explicitly to avoid the collision.

### 1.5 In Scope

- Add a `merge_method` key (enum: `squash` | `merge` | `rebase`) to each of the three mode profiles in `git-strategy.yaml.tmpl` (`manual` / `personal` / `team`), defaulting to `squash` for backward compatibility.
- Add corresponding Go struct field `MergeMethod string` to the per-mode profile struct in `internal/config/types.go` (placement matches the symmetry layer landed by `SPEC-V3R5-GIT-STRATEGY-SCHEMA-001`).
- Thread the read at the two sync-phase hardcoded surfaces (`sync/delivery.md` Step 3.4 + `manager-git.md` Auto-Merge section). The two surfaces SHALL read `git_strategy.<mode>.merge_method` and substitute the literal `--squash` flag with the configured method's `gh pr merge` flag (`--squash` | `--merge` | `--rebase`).
- Amend the FROZEN lifecycle table in `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` so the `PR strategy` cell for the sync row reads `configurable (default: squash)` (with a one-line footnote referencing this SPEC). The plan and run rows MAY remain `squash` since they are out of scope for issue #1061 — but see §1.6 Q3 for the explicit decision request.
- Update `internal/template/templates/.claude/skills/moai-ref-git-workflow/skill.md` cross-reference to point at the new `merge_method` key when explaining the gitflow release-branch guidance, so the reference doc and the actual behavior reconverge.

### 1.6 Decision Questions surfaced for plan-auditor review

Three decisions are surfaced here rather than pre-resolved, since they affect the FROZEN-rule relaxation contract:

- **Q1 — Wire-through to existing `internal/github.PRMerger`**: thread the configured method through the existing tested `PRMerger` abstraction (preferred — eliminates the "complete unused abstraction" defect and routes through code that already has tests), OR keep the agent-level `gh pr merge` literal substitution (simpler patch, but the unused abstraction stays unused). Recommended: wire-through.
- **Q2 — Per-branch-type overrides (gitflow `release/*` → `merge`)**: implement in this SPEC (closes the contradiction with `moai-ref-git-workflow/skill.md` immediately), OR defer to a follow-up SPEC (keeps this Tier S). Recommended: defer; document the extension shape (`merge_method: {default: squash, overrides: {release: merge, hotfix: merge}}`) as a §B.4 Out-of-Scope note so the schema can be extended additively later.
- **Q3 — Scope of FROZEN-rule relaxation**: relax only the sync row (matches issue scope), OR relax all three rows (plan/run/sync) symmetrically. Recommended: relax only the sync row in this SPEC; if a future SPEC needs configurable plan/run merge methods, it should re-amend the same table.

### 1.7 Out of Scope (will NOT be done in this SPEC)

- Per-branch-type override implementation (Q2; deferred — additive extension shape documented but not implemented).
- Changing the `internal/cli/worktree/sync.go` `--strategy` flag (unrelated; §1.4).
- Adding new merge methods beyond the three the `internal/github` abstraction already supports (`squash` / `merge` / `rebase`).
- Relaxing the plan/run rows of the FROZEN lifecycle table (Q3 default; out of scope for this SPEC).
- Implementing CI guards that detect future hardcoded `--squash` reintroductions (would belong in a follow-up audit SPEC; documented as §B.5 follow-up below).

---

## 2. EARS Requirements

REQ count: 6 (5 functional + 1 documentation/SSOT).

### REQ-GMM-001 — Config field surface

[ZONE:Evolvable] [HARD] Where the user is editing `git_strategy.<mode>` in `.moai/config/sections/git-strategy.yaml`, the `merge_method` key SHALL be an optional enum string with the three accepted values `squash` | `merge` | `rebase`. When the key is absent, the loader SHALL default to `squash` for backward compatibility with all SPECs created before this SPEC merges.

### REQ-GMM-002 — Go struct symmetry

[ZONE:Evolvable] [HARD] When `internal/config/loader.go` `Loader.Load()` reads `git-strategy.yaml`, the per-mode profile struct (the substruct that holds `branch_prefix`, `main_branch`, etc. for each of `manual` / `personal` / `team`) SHALL include a `MergeMethod string` field with the YAML tag `yaml:"merge_method"`, populated from the file value when present and from the `"squash"` default when absent. The default SHALL be sourced from a single named constant `DefaultMergeMethod = "squash"` in `internal/config/defaults.go` (no string-literal duplication per CLAUDE.local.md §14 [HARD] hardcoding prevention).

### REQ-GMM-003 — Sync skill body wire-through

[ZONE:Evolvable] [HARD] When `sync/delivery.md` Step 3.4 Auto-Merge runs, the workflow SHALL read `git_strategy.<active-mode>.merge_method` from the loaded config and substitute the `--squash` literal in both `delivery.md:313` and `:325` with the configured method's `gh pr merge` flag (`--squash` | `--merge` | `--rebase`). The `--delete-branch` flag SHALL be preserved unchanged across all three methods.

### REQ-GMM-004 — manager-git agent body wire-through

[ZONE:Evolvable] [HARD] When `manager-git` performs the sync-PR auto-merge per its own body documentation, the agent body SHALL reference `git_strategy.<active-mode>.merge_method` rather than hardcode `--squash` at lines `:140`, `:176`, and `:204` of `internal/template/templates/.claude/agents/moai/manager-git.md`. The new literal at each site SHALL match the shape `gh pr merge <PR> --<method> --delete-branch` where `<method>` is one of `squash` | `merge` | `rebase`.

### REQ-GMM-005 — FROZEN-rule cell amendment

[ZONE:Evolvable] [HARD] When the maintainer edits `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` per this SPEC, the `PR strategy` cell of the sync row in the `## SPEC Phase Discipline` lifecycle table (currently `squash` at line 31) SHALL be replaced with `configurable (default: squash)` and a single inline `*See SPEC-GIT-MERGE-METHOD-001 for the config key.*` footnote below the table. The `[ZONE:Frozen] [HARD]` zone marker at line 25 SHALL remain — the table itself stays FROZEN; only the sync cell becomes a typed placeholder referencing a documented config key. The plan and run rows SHALL remain `squash` per §1.6 Q3 default decision.

### REQ-GMM-006 — Reference-doc reconvergence

[ZONE:Evolvable] [HARD] When `internal/template/templates/.claude/skills/moai-ref-git-workflow/skill.md` lines `:133-134` and `:171` describe the squash-vs-merge-commit distinction for feature vs release branches, the body SHALL add a cross-reference to `git_strategy.<mode>.merge_method` so a reader of the reference doc can locate the configuration surface that honors the guidance. (Per-branch-type override implementation is out of scope per §1.6 Q2 — only the cross-reference is added, not the override behavior.)

---

## 3. Acceptance Criteria (Tier S inline per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier)

Each AC is a binary PASS/FAIL evidence-driven check. No `[HARD]` is needed inside individual ACs (they inherit `[HARD]` from the corresponding REQ).

### AC-GMM-001 — `merge_method` key surface verifiable in template

**Bound to:** REQ-GMM-001.

**PASS criterion:**

```bash
grep -c "merge_method:" internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl
# Expected: 3 (one per mode profile: manual / personal / team)
```

Each of the three matches SHALL be inside the corresponding `manual:` / `personal:` / `team:` mode block, and the rendered default SHALL be `squash` (a comment line above each occurrence SHALL document the three accepted values).

### AC-GMM-002 — Go struct field present + default constant present

**Bound to:** REQ-GMM-002.

**PASS criteria (3 sub-checks):**

```bash
# 1. Struct field with correct YAML tag
grep -nE 'MergeMethod\s+string.*yaml:"merge_method"' internal/config/types.go
# Expected: 1 match (on the per-mode profile struct)

# 2. Named default constant
grep -n 'DefaultMergeMethod\s*=\s*"squash"' internal/config/defaults.go
# Expected: 1 match

# 3. Loader test passes after adding the field
go test ./internal/config/...
# Expected: PASS — including the existing struct-yaml-symmetry CI guard
# (audit_struct_yaml_symmetry_test.go:TestStructYAMLSymmetry_*)
```

The struct-yaml-symmetry test is the canonical regression guard per `.claude/rules/moai/core/settings-management.md` § CI Guards, and the new key + field MUST be added together in a single commit so the guard never sees a transient mismatch.

### AC-GMM-003 — Sync skill body no longer hardcodes `--squash`

**Bound to:** REQ-GMM-003.

**PASS criterion:**

```bash
grep -nE '^\s*gh pr merge .*--squash --delete-branch\b' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md
# Expected: 0 matches (both line 313 and line 325 replaced by config-driven substitution)
```

A successful PASS implies the new shape at each former site reads from `git_strategy.<mode>.merge_method` (verified by reading-back the file, since the workflow body is markdown-prose orchestration rather than executable Go).

### AC-GMM-004 — `manager-git.md` no longer hardcodes `--squash`

**Bound to:** REQ-GMM-004.

**PASS criterion:**

```bash
grep -nE 'gh pr merge.*--squash --delete-branch' internal/template/templates/.claude/agents/moai/manager-git.md
# Expected: 0 matches
```

(Pre-SPEC baseline: 3 matches at :140 / :176 / :204 per §1.2 evidence #2.)

### AC-GMM-005 — FROZEN-rule cell amendment landed

**Bound to:** REQ-GMM-005.

**PASS criteria (2 sub-checks):**

```bash
# 1. Sync row PR-strategy cell now reads the placeholder
grep -nE '^\|\s*3\s*\|.*sync/SPEC-XXX.*configurable \(default: squash\)' internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md
# Expected: 1 match (line 31 region)

# 2. SPEC cross-reference footnote present
grep -n "SPEC-GIT-MERGE-METHOD-001" internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md
# Expected: ≥1 match
```

The `[ZONE:Frozen] [HARD]` line 25 SHALL be unchanged (re-grep confirms it still says `Frozen`, `HARD`, `4-step lifecycle`).

### AC-GMM-006 — Reference-doc cross-reference added

**Bound to:** REQ-GMM-006.

**PASS criterion:**

```bash
grep -n "merge_method" internal/template/templates/.claude/skills/moai-ref-git-workflow/skill.md
# Expected: ≥1 match (near the squash-vs-merge-commit table at :133-134 OR near the release-history paragraph at :171)
```

### AC-GMM-007 — Mirror parity preserved between `internal/template/templates/...` and `.claude/...`

**Bound to:** Template-First Rule per `internal/template/CLAUDE.md` and CLAUDE.local.md §2 [HARD].

**PASS criterion:**

```bash
make build
# Expected: exit 0; embedded.go regenerated; no diff in tracked files except the edits above

go test ./internal/template/... -run TestEmbeddedMirror
# Expected: PASS — byte-parity invariant between source templates and embedded.go
```

### AC-GMM-008 — Backward compatibility — existing projects unchanged

**Bound to:** REQ-GMM-001 default behavior.

**PASS criterion:** an existing project initialized before this SPEC merges (i.e. its `git-strategy.yaml` does NOT contain a `merge_method` key) SHALL exhibit identical behavior to the pre-SPEC main branch — the sync auto-merge SHALL still resolve to `gh pr merge --squash --delete-branch` via the `DefaultMergeMethod = "squash"` constant. This is verified by a unit test in `internal/config/loader_test.go` (or wherever the existing git-strategy loader tests live per the Round 5 layout of `SPEC-V3R5-GIT-STRATEGY-SCHEMA-001`) that loads a fixture YAML with no `merge_method:` key and asserts the loaded struct value is `"squash"`.

---

## 4. Risk Register (Tier S — abbreviated)

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| R1: FROZEN-rule cell amendment is interpreted as relaxing the broader 4-step lifecycle contract | Medium | High | The amendment is one cell; the `[ZONE:Frozen] [HARD]` zone marker remains. AC-GMM-005 sub-check 1 verifies the cell text exactly; sub-check 2 verifies the SPEC cross-reference is present. The plan and run rows remain `squash` per §1.6 Q3 default. |
| R2: Struct-yaml symmetry CI guard fails mid-commit if the YAML key lands without the Go field (or vice versa) | Low | Medium | Per AC-GMM-002 sub-check 3, the key + field + default constant MUST be added in a single commit. The run-phase manager-develop delegation prompt MUST list both files in the same M-milestone so the symmetry guard never observes a transient state. |
| R3: Per-branch-type override (gitflow `release/*` → `merge`) deferred and never landed | Medium | Low | §B.4 of this SPEC's plan.md documents the additive extension shape verbatim; a follow-up SPEC can pick it up without breaking-change risk. The §1.6 Q2 decision and the reference-doc cross-reference (REQ-GMM-006) keep the question alive in code search. |
| R4: Users with non-default `merge_method: rebase` discover that GitHub branch-protection rules disallow rebase merges on their repo | Low | Low | Out of scope for this SPEC (a GitHub config concern, not a MoAI concern). Documented in the new template's inline YAML comment above each `merge_method:` key so a user editing the file sees the warning. |
| R5: The wire-through path bypasses the unused `internal/github.PRMerger` abstraction (§1.6 Q1 deferred) | Low | Medium | If Q1 is resolved as "defer the PRMerger wire-through", the unused-abstraction defect remains. Mitigation: this SPEC documents the choice in plan.md §B.4 so a follow-up SPEC can wire `PRMerger` in additively. The recommended Q1 resolution (wire through) eliminates the risk entirely. |

---

## 5. Edge Cases (Tier S — abbreviated)

- **E1**: Invalid `merge_method` value in user config (e.g. `merge_method: rebase-then-squash`). Loader SHALL reject with a typed error citing the three accepted values; AC-GMM-008's backward-compat test SHALL be paired with a malformed-fixture test asserting the loader error message.
- **E2**: User sets `merge_method: merge` but their repo's GitHub branch protection forbids merge commits on `main`. The `gh pr merge --merge` call will fail with a GitHub-side error; this SPEC does not catch that case (it surfaces naturally as a `gh pr merge` non-zero exit, which the existing sync-phase error handling already reports).
- **E3**: User has a pre-this-SPEC `git-strategy.yaml` that already contains a `merge_method:` typo at a different YAML path (e.g. nested under `automation:` rather than under the mode profile). The struct-yaml-symmetry CI guard SHALL flag this as an unknown key during their next `moai update`. (No special handling needed; the guard already exists.)
- **E4**: A user invokes `gh pr merge` manually outside the `/moai sync` flow with `--squash`. This SPEC does not touch manual user behavior — `merge_method` only governs the automated sync-phase auto-merge step.

---

## 6. Dependencies & Sequencing

- **D1 (hard)**: `SPEC-V3R5-GIT-STRATEGY-SCHEMA-001` (status: completed) — the Go struct symmetry layer this SPEC extends. The Round 5 layout of the per-mode profile substructs is the placement site for `MergeMethod string`. Without D1, REQ-GMM-002 cannot be satisfied symmetrically.
- **D2 (advisory)**: `CLAUDE.local.md §14` (hardcoding prevention) — the `DefaultMergeMethod = "squash"` constant placement follows the same single-source-of-truth pattern as `DefaultBranchPrefix` / `DefaultCommitStyle` (per `internal/config/defaults.go:225-226` cited by the SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 audit).
- **D3 (informational)**: `internal/github/gh.go` + `internal/github/pr_merger.go` — pre-existing tested abstraction. §1.6 Q1 recommends wiring through this code; if Q1 resolves as "defer", D3 becomes a §B.4 follow-up.

No blocking SPECs are in `status: in-progress`; this SPEC may proceed to plan-phase immediately upon issue triage approval.

---

## 7. References

- Issue #1061 (this SPEC's source) — `https://github.com/modu-ai/moai-adk/issues/1061`
- `internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md:313,325` — hardcoded sites #1
- `internal/template/templates/.claude/agents/moai/manager-git.md:140,176,204` — hardcoded sites #2
- `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md:25-31` — FROZEN rule + lifecycle table
- `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` — config schema to extend
- `internal/template/templates/.claude/skills/moai-ref-git-workflow/skill.md:133-134,171` — reference-doc squash/merge guidance
- `internal/github/gh.go:49-55,241` + `internal/github/pr_merger.go:134` — existing unused abstraction
- `SPEC-V3R5-GIT-STRATEGY-SCHEMA-001` — Go struct symmetry layer (dependency D1)
- `CLAUDE.local.md §14` — hardcoding prevention (default constant placement)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — frontmatter schema SSOT
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Tier S definition (2-artifact set; AC inline)
