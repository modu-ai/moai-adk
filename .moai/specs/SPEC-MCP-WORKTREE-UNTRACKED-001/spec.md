---
id: SPEC-MCP-WORKTREE-UNTRACKED-001
title: "Accept a linked worktree as project_root when the repository keeps .moai/ out of git"
version: "0.3.0"
status: draft
created: 2026-09-26
updated: 2026-09-26
author: manager-spec (card t1202)
priority: P1
phase: "v3.1.4 target"
module: "internal/cli"
lifecycle: spec-anchored
tier: M
issue_number: 1716
related_specs: [SPEC-MCP-WORKTREE-ROOT-001]
tags: "mcp, worktree, project-root, untracked-moai, audit, validator"
---

# SPEC-MCP-WORKTREE-UNTRACKED-001

## HISTORY

| Version | Date | Author | Change |
|---|---|---|---|
| 0.1.0 | 2026-09-26 | manager-spec (t1202) | Initial plan-phase draft. Compares design (a) validator-side acceptance against design (b) creation-side provisioning and recommends (a). |
| 0.2.0 | 2026-09-26 | manager-spec (t1202) | Revision for plan-audit iter-1 (FAIL 0.62, D1–D16): per-tool `.moai` inventory, four access classes, source-root and primary-checkout predicates, design (b) premise corrected. |
| 0.3.0 | 2026-09-26 | manager-spec (t1202) | Scope reduction after plan-audit iter-2 (FAIL 0.78), lead decision option B. Kept: validator acceptance of a registered linked worktree whose primary checkout is a MoAI root, and tree operations resolving to the worktree. Moved to follow-up card t1213: configuration/catalogue source-root routing and state-write destinations. Fixed D17 (today's acceptance path is untouched), D19, D21, D23, D24. |

## §1 Problem

SPEC-MCP-WORKTREE-ROOT-001 gave every tree-sensitive MCP tool an optional
`project_root` input so a worktree session can name its own tree. The input is
validated by `validateProjectRoot` (`internal/cli/mcp_project_root.go`, the
`.moai` check at the end of the function), and that validator accepts a path
only when `<path>/.moai` is a directory.

That test is correct for a repository that commits `.moai/`. It is wrong for a
repository that keeps `.moai/` untracked or gitignored — a common choice in user
projects. There, every tree created with `git worktree add` — including the trees
MoAI's own creation paths make — contains tracked files only, so the new
worktree has **no `.moai`** and the validator rejects it:

```
project_root ".../wt" has no .moai directory, so it is not a MoAI project root
```

The caller is left with two outcomes, both wrong:

1. Pass `project_root` → rejected. The audit and graph tools cannot review the
   worktree's own changes.
2. Omit `project_root` → resolution falls back to `CLAUDE_PROJECT_DIR`, which
   names the **primary checkout**. The call succeeds and silently audits the
   primary tree, missing every change made in the worktree — the silent
   wrong-tree defect the `project_root` input was introduced to remove.

Premise measured LIVE on base `df526c9a9` (evidence
`.moai/reports/t1202/verdict.md`): a temporary test under `t.TempDir()` built a
repository with `.moai/` gitignored, ran `git worktree add`, and observed the
primary accepted (positive control), the worktree's `.moai` absent, and the
worktree rejected with the message above. Issue #1716 reproduces the same
rejection starting from `moai worktree new`.

## §2 Design comparison

Both designs keep the reject-never-fall-back contract.

### §2.1 Design (a) — the validator recognizes a linked worktree

The validator keeps today's test as its first branch: a path whose `.moai` is a
directory is accepted exactly as now. Only when that test fails does a second,
new branch run: the validator asks git whether the path is the top level of a
linked worktree **registered** in a repository whose **primary checkout**
(REQ-MWU-004) has a `.moai` directory. If so, the path is accepted and tree
operations act on the worktree. An unrelated directory, a non-MoAI repository, a
subdirectory, or an ambiguous repository layout still fails.

The new branch is a registration check (the candidate must be listed by
`git worktree list --porcelain` and not flagged prunable) plus a common-dir
check, not a filesystem containment check. Paths are compared in
symlink-canonical form.

| Property | Assessment |
|---|---|
| Changes today's accepted set | No — the existing branch is untouched; the new branch only adds paths that are rejected today |
| Covers trees made by `moai worktree new` (the issue's reproduction path) | Yes |
| Covers trees made by hand with `git worktree add`, and existing trees | Yes |
| Introduces a second copy of `.moai` | No |
| New cost | A scrubbed git subprocess on the no-`.moai` path only; fail-closed on error |

### §2.2 Design (b) — worktree creation provisions `.moai`

Every creation path copies or links `.moai` into the new tree, so the existing
validator passes unchanged. **Issue #1716 names this as an acceptable expected
behavior** ("provide the needed project settings safely when the worktree is
created"), and its reproduction starts from `moai worktree new` — a MoAI
creation path that (b) would cover. (b) is a legitimate option.

Measured facts about (b) (commands in `plan.md` §B.1):

| Property | Measured assessment |
|---|---|
| `git worktree add` execution sites | 2 — `internal/core/git/worktree.go` (the `Add` argument builder) and `internal/cli/session_worktree.go` (the session materializer) |
| Creation entry points reaching them | 6 — session entry (`session_worktree.go`, `enterSessionWorktree`), `moai worktree new` (`root.go` wires the materializer), factory lane handoff (`factory_lane_handoff.go`), its recovery path (`factory_lane_handoff_recover.go`), the launcher `-w --branch` path (`worktree_branch_flag.go`), and the `WorktreeCreate` hook (`internal/hook/worktree_create.go`) |
| Trees made by hand with `git worktree add` | Not covered — no MoAI code runs |
| Trees that already exist | Not covered — each needs separate provisioning |
| Copy variant | Two copies of configuration and SPECs; an edit in one is not seen by the other |
| Link variant | Writes through the link land in the primary tree; the issue reporter deliberately avoided this workaround |

### §2.3 Recommendation — design (a) alone

Recommend **design (a)**, and do **not** implement (b) alongside it:

1. **Coverage.** (a) covers every tree, including existing and hand-made ones;
   (b) covers only trees one of six entry points creates from now on.
2. **One copy.** (a) introduces no second `.moai`; (b)'s copy variant does, and
   its link variant is the workaround the reporter avoided.
3. **One boundary.** (a) adds one branch to the validator every `project_root`
   tool already shares; (b) changes six entry points and still leaves existing
   trees broken.
4. **Both is worse than (a).** A provisioned `.moai` in the worktree would send
   it down the existing branch with a diverging copy.

Issue #1716 lists (b) as acceptable; the operator decides at Implementation
Kickoff (`plan.md` §C, decision 2).

## §3 In-scope access inventory (tree operations)

Measured on base `df526c9a9`. These accesses take the call's root and are tree
operations; under this SPEC they act on the accepted worktree.

| Tool | Access | Code site |
|---|---|---|
| `codex_audit` | review `cwd` | `mcp_codex.go` handler, `params["cwd"] = root` |
| `claude_audit` | diff collection | `performClaudeAuditWith` → `collectReviewDiff(root, target)` (`mcp_review_material.go`) |
| `glm_audit` | diff collection | `mcp_glm.go` handler, root from `resolveToolProjectRoot` |
| `audit_multi` | per-backend review tree | `mcp_audit_multi.go` → `mcp_convergence.go` |
| `codex_audit`, `claude_audit`, `glm_audit`, `audit_multi` | build identity (git comparison) | `auditBuildIdentity(ctx, root)` (`mcp_build_identity.go`) |
| `graph_file_api`, `graph_find_code`, `graph_shortest_path`, `graph_trace_calls` | derived graph artifact `<root>/.moai/project/graph/edges.jsonl` | `internal/graph/codequery.go` `edgesArtifactPath`, `AnswerProvenance` |

The graph artifact lives under `.moai` but is derived from the tree's own code,
so it is read from the worktree only: a worktree without one gets the existing
"graph layer absent" error, never the primary's graph.

The configuration, catalogue, and state accesses of the same tools (SPEC
catalogue, `config/sections/*.yaml`, `state/verify`, `state/audit-receipts`,
`state/audit-multi`) are deferred to card t1213 — see §6.

## §4 Requirements (GEARS)

### §4.1 Acceptance

- **REQ-MWU-001 (Ubiquitous, existing branch unchanged).** The validator shall
  accept a `project_root` whose canonical path has a `.moai` directory, and
  shall return that canonical path, exactly as it does today and without running
  any git subprocess.

- **REQ-MWU-002 (Event-driven, linked-worktree branch).** When a `project_root`
  has no `.moai` directory, and its canonical path is the top level of a linked
  git worktree that `git worktree list --porcelain` lists without a `prunable`
  mark, and the repository's primary checkout (REQ-MWU-004) has a `.moai`
  directory, the validator shall accept it and return the canonical worktree
  path.

- **REQ-MWU-003 (Event-driven, rejection preserved).** When a `project_root`
  with no `.moai` directory fails any condition of REQ-MWU-002 — not inside a
  git repository, not a worktree top level (including a subdirectory of one),
  the primary checkout itself, a worktree that is not listed or is marked
  prunable, or a repository whose primary checkout has no `.moai` directory —
  the validator shall reject it with an error naming the failed condition, and
  shall never substitute a default root.

- **REQ-MWU-004 (Ubiquitous, primary-checkout predicate).** The validator shall
  identify the primary checkout as the parent directory of the repository's git
  common dir, and shall treat the layout as ambiguous — and reject — unless the
  common dir's final path element is `.git`, that parent's own `.git` resolves
  to the same common dir, and the parent is the first `worktree` entry of
  `git worktree list --porcelain`. (This rejects `--separate-git-dir`
  repositories, submodule-internal git dirs, and bare repositories.)

- **REQ-MWU-005 (Event-driven, fail-closed).** When a git inspection required by
  REQ-MWU-002 or REQ-MWU-004 cannot complete — git unavailable, a non-zero exit,
  or output of an unexpected shape — the validator shall reject the path, and
  shall not retry through an alternative git invocation that could change the
  answer.

- **REQ-MWU-006 (Ubiquitous, canonical comparison and environment isolation).**
  The validator shall compare the candidate, the listed worktree paths it
  matches against, and the primary path in symlink-canonical form, and shall run
  every git inspection with `GIT_DIR`, `GIT_WORK_TREE`, `GIT_COMMON_DIR`,
  `GIT_INDEX_FILE`, and `GIT_CEILING_DIRECTORIES` removed from the child
  environment.

- **REQ-MWU-007 (Ubiquitous, sibling entries).** The validator shall decide
  REQ-MWU-002 on the listed entry that matches the candidate only; any other
  listed entry that is marked prunable or cannot be canonicalized shall be left
  out of the comparison and shall not be a reason to reject the candidate.

### §4.2 Tree operations

- **REQ-MWU-008 (State-driven, tree operations).** While a call's `project_root`
  was accepted through REQ-MWU-002, every access in §3 shall act on the accepted
  worktree path, and the graph tools shall never answer from the primary
  checkout's graph artifact.

### §4.3 Compatibility and documentation

- **REQ-MWU-009 (Ubiquitous, absent parameter unchanged).** The resolution of an
  absent or empty `project_root` shall remain unchanged for both the fallback
  and the pass-through variants.

- **REQ-MWU-010 (Ubiquitous, documentation).** The shared input description text
  (`projectRootDescCommon`) and `.claude/rules/moai/core/moai-mcp-tools.md`
  § The `project_root` input shall state the linked-worktree acceptance and that
  configuration, catalogue, and state are still read from the accepted tree; the
  rule file and its template mirror under `internal/template/templates/` shall
  stay byte-identical, and the template copy shall carry no SPEC ID, card id, or
  date.

## §5 Constraints

- The reject-never-fall-back contract of the `project_root` input is binding.
- The git subprocess runs only on the no-`.moai` branch; the existing branch
  gains no subprocess.
- Template text stays neutral across the 16 supported programming languages and
  free of internal development state (no SPEC IDs, card ids, dates, commit SHAs).
- No creation path writes, copies, or links `.moai` as part of this SPEC.

## §6 Exclusions (What NOT to Build)

This section records what is out of scope for this SPEC.

### Out of Scope — configuration, catalogue, and state routing (deferred to card t1213)

- Routing configuration and catalogue reads (`.moai/specs`,
  `.moai/config/sections/*.yaml` — audit pins, audit gates, `llm.yaml`, the
  gate-required check used for receipt ids) of a linked-worktree root to the
  primary checkout's `.moai`.
- Choosing the destination of state writes and their paired reads
  (`verify_snapshot` record/load, `verify_trend`, audit receipts written by
  `codex_audit` and `audit_multi`, `audit_multi` convergence state), including
  measuring which root the consuming hooks read (receipt guard, multi-review
  gate).
- Handling a worktree-local untracked `.moai/specs` versus the primary
  catalogue.
- Interim behavior until t1213 lands, stated so it is not mistaken for intent:
  on a root accepted through REQ-MWU-002, catalogue tools return the worktree's
  (empty) catalogue; configuration reads see no worktree config and use
  defaults, so a primary-declared `required` audit gate is not seen; and a state
  write creates `.moai` under the worktree, after which later calls take the
  REQ-MWU-001 branch. `plan.md` §C decision 1 asks the operator to accept this
  interim or pull a warning signal forward.

### Out of Scope — design (b) creation-side provisioning

- Copying or linking `.moai` into worktrees from any of the six entry points in
  §2.2, and any re-provisioning of existing worktrees.

### Out of Scope — ambiguous repository layouts

- `--separate-git-dir` repositories, submodule-internal git dirs, and bare
  repositories: rejected by REQ-MWU-004, not supported.

### Out of Scope — adjacent behavior

- Changing `resolveProjectDir()` or the absent-parameter fallback (owned by
  SPEC-MCP-WORKTREE-ROOT-001).
- Building the code graph automatically for a worktree.
- The base branch `moai worktree new` selects (observed on this card, tracked
  separately).
- MCP tools that take no `project_root` input (including `codex_role_audit`),
  and the CLI verbs.
