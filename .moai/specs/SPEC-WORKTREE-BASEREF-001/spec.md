---
id: SPEC-WORKTREE-BASEREF-001
title: "Configurable card-worktree base branch — one stored setting, two consumers (origin/HEAD alignment at SessionStart, git worktree add base operand), surfaced by moai doctor and the web console"
version: "0.3.1"
status: completed
created: 2026-08-27
updated: 2026-08-27
author: manager-spec (card t313)
priority: P1
phase: "v3.2.0"
module: "internal/config, internal/hook, internal/cli, internal/settings, internal/web, internal/template/templates/.moai/config/sections"
lifecycle: spec-anchored
tags: "worktree, base-branch, git-metadata, session-start, doctor, web-console, template-neutrality, backward-compat"
tier: M
related_specs:
  - SPEC-SYNC-STRATEGY-KEY-001
  - SPEC-WORKTREE-BRANCH-GUARD-001
---

# SPEC-WORKTREE-BASEREF-001 — Configurable card-worktree base branch

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-08-27 | manager-spec (card t313) | Initial plan-phase emission. Tier M. 15 GEARS REQs / 14 binary ACs. All citations measured in worktree `.claude/worktrees/t313` at HEAD `48eb945df`, branch `WT-worktree-baseref`. Carries one unresolved design conflict (plan.md §B D2 — web widget shape) returned to the orchestrator as a blocker. |
| 0.3.0 | 2026-08-27 | manager-spec (card t313) | Plan-audit iteration 1 repair (verdict `.moai/reports/t313/plan-audit-iter1.md`, FAIL 0.78 harmonic, 7/7 must-pass clear). Design unchanged; five blocking specification defects repaired and six debt items folded in. Repairs land WITHIN the Tier M 16/16 ceiling (`spec-workflow.md:152`) by extending existing requirements rather than adding new ids: REQ-WBR-004 becomes compound and carries the D4 concurrency narrowing (consumer 1 fires only from the primary checkout); REQ-WBR-013 absorbs the typed round-trip preservation promoted from the orphaned AC-WBR-014; REQ-WBR-011 is bound to REQ-WBR-009's predicate as the sole resolvability authority for both consumers; REQ-WBR-012 fixes the exact `DiagnosticCheck.Name`; REQ-WBR-008's citation corrected to `:176-181`. One AC added — AC-WBR-016, REQ-WBR-004's firing-point criterion, carrying both the primary-checkout and the linked-worktree halves. Now 16 GEARS REQs / 16 binary ACs, exactly at the ceiling. Citations re-measured in worktree `.claude/worktrees/t313` at HEAD `48eb945df`. |
| 0.2.0 | 2026-08-27 | manager-spec (card t313) | D2 blocker CLOSED by operator ruling: `TypeText` with `main` / `develop` named in the description; the new-combo-widget alternative rejected, the `FieldType` set at `internal/settings/schema.go:105-113` unchanged. Consequences: REQ-WBR-014 restated as free text, AC-WBR-011 made unconditional, and free text's load-bearing consequence added as REQ-WBR-009 (pre-write resolvability check) + AC-WBR-015, with REQ-WBR-012 extended to report the unresolvable-value case as its own non-OK state. Old REQ-WBR-009..015 renumbered to 010..016. Now 16 GEARS REQs / 15 binary ACs (Tier M ceiling 16/16). Citations re-measured in the same worktree at HEAD `48eb945df`. |
| 0.3.1 | 2026-08-27 | manager-spec (card t313) | Plan-audit iteration 2 debt repair (verdict `.moai/reports/t313/plan-audit-iter2.md`, PASS-WITH-DEBT 0.92 harmonic, 7/7 must-pass clear, zero blocking defects). Two criterion-layer repairs, no design change and no id change (still 16 GEARS REQs / 16 binary ACs). **N1** — AC-WBR-013 check (1) probed only a same-named plain-file template counterpart, so a correct implementation of this SPEC's own plan §D write list emitted a false `NO-TEMPLATE-COUNTERPART` for `.moai/config/sections/git-strategy.yaml`, whose counterpart ships only as `git-strategy.yaml.tmpl` (plan.md §B G5); the probe now accepts either form and reports only when neither was changed in the diff. **N2** — AC-WBR-016's "read seam" carried two denotations that disagreed on the empty-value branch its own `Given` admits; the seam is now pinned in the criterion as the alignment-entry (configured-value) read, with the reasoning recorded at plan.md §A D3.2. Remaining audit debt (N3, N4, N5, the seam budget, the folding traceability, and inherited G2/G4/G6) is knowingly carried into run-phase. |

## §A Context

### A.1 The reported problem

Card worktrees created with the Claude Code `EnterWorktree` tool branched from `origin/main`, while two [HARD] protocol rules require card worktrees to be cut from `develop`:

- `.claude/rules/local/gitflow-lane-protocol.md` §1 rule 1
- `CLAUDE.local.md` §4.1.1 rule 1

Three lanes independently reported a file missing that exists only on develop-descended refs. The missing file was the symptom; the base ref was the cause.

### A.2 The emergency fix and what it proved

On 2026-08-27 the lead ran `git remote set-head origin develop`, moving `refs/remotes/origin/HEAD` from `origin/main` to `origin/develop`.

Measured in this worktree at HEAD `48eb945df`:

| Observation | Command | Output |
|---|---|---|
| origin/HEAD now names develop | `git symbolic-ref refs/remotes/origin/HEAD` | `refs/remotes/origin/develop` |
| This tree was cut from develop | worktree reflog (lead-measured, cited in the dispatch) | `branch: Created from origin/develop` |

So `EnterWorktree`'s `fresh` mode does read `refs/remotes/origin/HEAD`. The fix works — but it is a hand-applied mutation of local repository metadata, not a reproducible configuration.

### A.3 The decisive constraint

`worktree.baseRef` **cannot name a branch**. `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md:171` states verbatim: *"accepts only `"fresh"` or `"head"`, not arbitrary refs"* (the local mirror `.claude/rules/moai/workflow/worktree-integration.md:171` carries the identical sentence).

Therefore the reproducible path **cannot** be a Claude Code settings pass-through key. The only handle on `EnterWorktree`'s base is `refs/remotes/origin/HEAD`, which is local repository metadata and does not survive a fresh clone. The reproducible path must be a **consumer that sets that handle from a stored setting**.

### A.4 The second, independent defect

moai's own worktree creation path passes no base at all. `internal/cli/session_worktree.go:217-219` (`gitWorktreeAddReal`) runs:

```go
cmd := exec.Command("git", "worktree", "add", "-b", branch, destDir)
```

With no base operand, `git worktree add` branches from the **current HEAD** of the invoking tree. The primary checkout currently sits on `main`, so `moai cc -w <name>` is also main-based. This is a second defect on the same axis, in a path this repository owns end to end.

## §B Requirements (GEARS)

### B.1 The stored setting

**REQ-WBR-001** (Ubiquitous) — The configuration schema shall carry a single key `git_strategy.worktree_base_branch`, a string naming the branch that card worktrees are cut from, stored in `.moai/config/sections/git-strategy.yaml` at the `git_strategy` root level (peer of `mode` / `provider` / `github_username`, per the shape measured at `.moai/config/sections/git-strategy.yaml:1-6`).

**REQ-WBR-002** (Ubiquitous) — The key's default value shall be the empty string, and the empty value shall mean "take no action", reproducing pre-SPEC behavior exactly.

**REQ-WBR-003** (Unwanted) — The shipped template default shall not name `develop`, `main`, or any other repository-specific branch; `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` shall ship the key with an empty value only.

### B.2 Consumer 1 — origin/HEAD alignment (EnterWorktree's handle)

**REQ-WBR-004** (Event-driven, compound) — **When** a session starts **while** its working directory is the primary checkout, the SessionStart handler shall read the configured `git_strategy.worktree_base_branch` exactly once and compare it against the branch currently named by `refs/remotes/origin/HEAD`; and **while** its working directory is inside a linked worktree, it shall perform no read of either, no write, and shall emit no output. The primary-checkout test shall be the discriminant already implemented by `inGitWorktreeReal` (`internal/cli/session_worktree.go:234-241`): the primary checkout is where `git rev-parse --git-dir` and `git rev-parse --git-common-dir` resolve to the same path; they differ inside a linked worktree.

> Why the second clause exists. `.moai/config/sections/git-strategy.yaml` is a **tracked** file (measured: `git ls-files --error-unmatch .moai/config/sections/git-strategy.yaml` → rc 0), so each active card worktree carries its own working-tree copy whose content follows its own branch, while `refs/remotes/origin/HEAD` is a **single repository-global** handle. Unnarrowed, every active lane is a writer of one shared handle, and two lanes whose branches carry different values each observe a difference the other just created — the analysis is plan.md §E R1. One writer suffices, because the handle is repository-global. The narrowing binds **consumer 1 only**: consumer 2 (REQ-WBR-010 / REQ-WBR-011) is unaffected and shall honour the configured base from any working tree, linked worktrees included.

**REQ-WBR-005** (State-driven) — **While** the configured value is empty or the key is absent, the SessionStart handler shall perform no git-metadata read-modify-write and emit no output.

**REQ-WBR-006** (State-driven) — **While** the configured value equals the branch already named by `refs/remotes/origin/HEAD`, the SessionStart handler shall perform no write and emit no output.

**REQ-WBR-007** (Event-driven) — **When** the configured value differs from the branch named by `refs/remotes/origin/HEAD` **and** that value resolves to an existing remote-tracking branch (REQ-WBR-009), the SessionStart handler shall run `git remote set-head origin <configured-value>` and shall emit exactly one line naming the previous branch, the new branch, and the setting that caused the change.

**REQ-WBR-008** (Unwanted) — The SessionStart handler shall not block, fail, or abort session start when the alignment step errors; every failure path shall be fail-open, matching the best-effort contract already stated for the synchronous step group at `internal/hook/session_start.go:176-181` (measured: the contract sentence *"Handle never returns a non-nil error from these steps"* begins at `:176`).

**REQ-WBR-009** (Event-driven) — **When** the alignment step is about to perform a write, it shall first verify that the configured value resolves to an existing remote-tracking branch by running `git show-ref --verify refs/remotes/origin/<configured-value>`; and **while** that verification does not succeed, it shall perform no `git remote set-head`, shall emit exactly one diagnostic line naming the unresolvable value, and shall let session start proceed normally (fail-open, REQ-WBR-008). `refs/remotes/origin/HEAD` shall never be left naming a ref that does not exist.

> The check shall use the plumbing form `git show-ref --verify refs/remotes/origin/<value>` and shall not use `git branch --list` or `git branch -vv`: BranchGuard's `\bgit\s+branch\b` pattern does not distinguish read-only branch queries from branch-state mutation, so it refuses those two invocations at the tool layer (CLAUDE.local.md §4.1.4, "알려진 마찰"). A run-phase implementation reaching for the porcelain form would be blocked rather than merely slower.

### B.3 Consumer 2 — moai's own worktree creation

**REQ-WBR-010** (Event-driven) — **When** `materializeSessionWorktree` (`internal/cli/session_worktree.go:181`) creates a session worktree and `git_strategy.worktree_base_branch` carries a non-empty value, the worktree-add call shall pass that value as the base operand, so the new tree is cut from the configured branch rather than from the invoking tree's HEAD.

**REQ-WBR-011** (State-driven) — **While** the configured value is empty, absent, or unresolvable as a git ref, the worktree-add call shall invoke `git worktree add -b <branch> <dest>` with no base operand, byte-identically to the invocation measured at `internal/cli/session_worktree.go:218`. "Unresolvable" shall be decided by the predicate REQ-WBR-009 specifies, which shall be implemented **once** as a single shared helper and shall be the **sole** resolvability authority for both consumers; consumer 2 shall not carry a second resolvability rule of its own (`git rev-parse --verify`, a `git branch --list` scrape, or a local-branch check are each a violation of this requirement even when their runtime behaviour agrees). The two consumers cannot be permitted to disagree about whether a configured value is usable — see plan.md §A D4.

### B.4 Diagnostic and web surfaces

**REQ-WBR-012** (Event-driven) — **When** `moai doctor` runs, it shall report a diagnostic item comparing the configured base branch against `refs/remotes/origin/HEAD`, distinguishing four states: OK on an empty setting, OK on match, a non-OK status naming the repair command on a plain mismatch, and a **separate** non-OK status naming the unresolvable value when the configured value does not resolve to an existing remote-tracking branch (REQ-WBR-009). The two non-OK states shall not be collapsed into one, because their repairs differ — a mismatch is repaired by running the alignment (start a session, or run the repair command), an unresolvable value is repaired by correcting the setting. The item's `DiagnosticCheck.Name` shall be exactly the string `Worktree Base Branch` — `moai doctor --check` filters by exact name equality (`internal/cli/doctor.go:232`: `if filterCheck != "" && c.name != filterCheck`), so the name is part of the contract, not an implementation detail. The item shall follow the existing `DiagnosticCheck` shape (`internal/cli/doctor.go:30-35`), use the `uikit.CheckStatus` enum (`ok` / `warn` / `fail`, `internal/cli/uikit/types.go:11-18`), and register in the same group as `Worktree State` (`internal/cli/doctor.go:220`).

**REQ-WBR-013** (Where — capability gate) — **Where** the web console renders the `git-worktree` panel (`internal/web/schemaform.go:226-230`), the panel shall carry an editable control for `git_strategy.worktree_base_branch`, registered as a `FieldDef` in `gitStrategyFields()` (`internal/settings/schema_sections.go:160-177`) and routed to the typed struct by `applyGitStrategyKey` (`internal/settings/sectionapply.go:170-204`); and that write shall preserve the keys present in the file but unmodelled by `GitStrategyConfig` — `manual.develop_branch`, `manual.release_branch_prefix`, and `manual.rc_version_format`, all three set at `.moai/config/sections/git-strategy.yaml:15-17` and absent from `ModeProfile` (`internal/config/types.go:106-132`). Preservation is a correctness property of the write path this SPEC introduces: the SPEC does not repair the `ModeProfile` schema gap (§C), but it shall not newly expose that gap by silently dropping keys this repository depends on.

**REQ-WBR-014** (Ubiquitous) — The web control shall be a free-text field (`TypeText`, `internal/settings/schema.go:109`) whose description names `main` and `develop` as the two branch names in common use; it shall accept any other branch name, and shall accept the empty value as the neutral "take no action" state. It shall not be a closed option set (`TypeSelect` / `TypeRadio`) — see plan.md §A D2 for the ruling and its template-neutrality reason.

### B.5 Cross-cutting

**REQ-WBR-015** (Ubiquitous) — Every stored key introduced by this SPEC shall have an observable runtime consumer proven by an executing test, and a regression guard modelled on `internal/web/dead_config_guard_test.go` shall pin that the key is present in `settings.AllFields()`, present in the rendered console HTML, and reaches a consumer — a grep over source text shall not be accepted as the proof.

**REQ-WBR-016** (Event-driven) — **When** a file under `.claude/`, `.moai/`, or shipped config is changed, the change shall be made in `internal/template/templates/` first, `make build` shall be run, and the local mirror shall be synchronized; and **when** a hook wrapper `.sh` file is edited, its `.sh.tmpl` twin shall be edited identically in the same change (CLAUDE.local.md §2, §2.3).

## §C Exclusions

### Out of Scope — tab_schema.json

- Neither `.claude/skills/moai-workflow-project/schemas/tab_schema.json` nor its template mirror is modified by this SPEC. Card t316 owns those files and its worktree is active. Any implication for the interview schema is recorded as a note in plan.md, never as a work item here.

### Out of Scope — the `automation.auto_branch` spelling defect

- The canonical spelling of the neighbouring automation key carries the `.automation` level (`.moai/config/sections/git-strategy.yaml:22-26`, `:49-53`, `:76-80`). `tab_schema.json` omits that level. That is card t316's defect and is not addressed here.

### Out of Scope — the `moai update` config-revert hazard

- `.moai/config` sits inside the `moai update` wipe root (CLAUDE.local.md §2.3), so a locally-set `worktree_base_branch` shares the existing hazard that update reverts `git-strategy.yaml` to template defaults. This SPEC inherits the hazard; it does not fix it.

### Out of Scope — changing `worktree.baseRef` semantics

- `worktree.baseRef` accepts only `"fresh"` or `"head"` (§A.3). This SPEC does not attempt to extend, wrap, or pass a branch through it.

### Out of Scope — ModeProfile schema gaps

- `ModeProfile` (`internal/config/types.go:106-132`) carries no `develop_branch`, `release_branch_prefix`, or `rc_version_format` field, although `.moai/config/sections/git-strategy.yaml:15-17` sets all three under `manual`. That divergence is observed and recorded (plan.md §B Gaps) but is not repaired here.

### Out of Scope — retroactive repair of existing worktrees

- Existing card worktrees already cut from the wrong base are not rebased, reset, or re-created by this SPEC. The doctor item reports the metadata state; it does not repair trees.
