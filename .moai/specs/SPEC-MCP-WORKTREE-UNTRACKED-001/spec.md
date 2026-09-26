---
id: SPEC-MCP-WORKTREE-UNTRACKED-001
title: "Accept a linked worktree as project_root when the repository keeps .moai/ out of git"
version: "0.1.0"
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
| 0.1.0 | 2026-09-26 | manager-spec (t1202) | Initial plan-phase draft. Compares design (a) validator-side acceptance against design (b) creation-side provisioning and recommends (a). Operator decides at Implementation Kickoff. |

## §1 Problem

SPEC-MCP-WORKTREE-ROOT-001 gave every tree-sensitive MCP tool an optional
`project_root` input so a worktree session can name its own tree. The input is
validated by `validateProjectRoot` (`internal/cli/mcp_project_root.go`, the
`.moai` check at the end of the function), and that validator accepts a path
only when `<path>/.moai` is a directory.

That test is correct for a repository that commits `.moai/`. It is wrong for a
repository that keeps `.moai/` untracked or gitignored — a common choice in user
projects. There, `git worktree add` (and every MoAI creation path, which all end
in `git worktree add`) produces a tree that contains the tracked files only, so
the new worktree has **no `.moai`** and the validator rejects it:

```
project_root ".../wt" has no .moai directory, so it is not a MoAI project root
```

The caller is left with two outcomes, both wrong:

1. Pass `project_root` → rejected. The audit, SPEC, verify, and graph tools are
   unusable from that worktree.
2. Omit `project_root` → resolution falls back to `CLAUDE_PROJECT_DIR`, which
   names the **primary checkout**. The call succeeds and silently audits the
   primary tree, missing every change made in the worktree. This is exactly the
   silent-wrong-tree defect the `project_root` input was introduced to remove.

Premise measured LIVE on base `df526c9a9` (evidence
`.moai/reports/t1202/verdict.md`): a temporary test under `t.TempDir()` built a
repository with `.moai/` gitignored, ran `git worktree add`, and observed the
primary accepted (positive control), the worktree's `.moai` absent, and the
worktree rejected with the message above.

Affected tools (the `project_root` family listed in
`.claude/rules/moai/core/moai-mcp-tools.md` § The `project_root` input):
`spec_progress`, `spec_audit`, `spec_drift`, `verify_snapshot`, `verify_trend`,
`codex_audit`, `glm_audit`, `claude_audit`, `audit_multi`, `graph_file_api`,
`graph_find_code`, `graph_shortest_path`, `graph_trace_calls`.

## §2 Design comparison

Two designs close the gap. Both keep the reject-never-fall-back contract.

### §2.1 Design (a) — the validator recognizes a linked worktree

When the named path has no `.moai`, the validator asks git whether the path is
the top level of a linked worktree registered in a repository whose **primary
checkout** has `.moai`. If so, the path is accepted and the call acts on the
worktree. An unrelated directory, a non-MoAI repository, or a subdirectory still
fails.

What it must also decide — which tree `.moai`-relative reads come from. The
worktree has no `.moai`, so a tool that reads MoAI state (SPEC catalogue,
verification snapshots, config sections) would find nothing under the worktree
and return an empty result that looks like success. That is a new silent-wrong
outcome, so design (a) splits the answer in two:

- **Tree root** — the worktree. Diff collection, code graph queries, and the
  working directory handed to review backends act here.
- **`.moai` root** — the worktree when it has its own `.moai`, otherwise the
  primary checkout. Every `.moai`-relative read the tool performs on the call's
  behalf resolves here, and the response names both roots.

Symlink canonicalization already precedes the `.moai` check in the validator;
its own comment anticipates a containment check against the git common dir.
Design (a) is that containment check: every path compared (the candidate, the
worktree list entries, the primary checkout) is compared in canonical form, so
a symlinked spelling cannot change the verdict.

| Property | Assessment |
|---|---|
| Covers trees made by hand (`git worktree add`, the reporter's case) | Yes |
| Covers trees that already exist today | Yes — no re-creation needed |
| Covers trees Claude Code creates outside MoAI's control | Yes — the check reads git, not the creator |
| Introduces a second copy of `.moai` | No |
| Change surface | One validator, the provenance block, the `.moai`-reading handlers, two doc copies |
| New failure mode | A git subprocess per call on the no-`.moai` path only; fail-closed on error |

### §2.2 Design (b) — worktree creation provisions `.moai`

Every creation path copies or links `.moai` into the new tree, so the existing
validator passes unchanged.

| Property | Assessment |
|---|---|
| Creation paths to change | At least four: `moai worktree new` (shared materializer), the launcher `-w --branch` materializer, the session-worktree materializer, and the `WorktreeCreate` hook serving `EnterWorktree` / agent isolation — plus the L2 tree path |
| Trees made by hand with `git worktree add` | **Not covered** — this is the reporter's own path |
| Trees that already exist | **Not covered** — each needs manual provisioning |
| Copy variant | Two diverging copies: config edited in one is stale in the other; state written by one tree (snapshots, SPEC status, goal state) is invisible to the other |
| Link variant | The symlink workaround the issue reporter deliberately avoided; it also routes worktree-local writes into the primary tree through a path that does not look shared |
| Contract with the no-auto-restore stance | A copy of runtime-written files (which may hold tokens or absolute paths) is the same class of act the project refuses for `.claude/settings.json` |

### §2.3 Recommendation — design (a) alone

Recommend **design (a)**, and do **not** implement (b) alongside it.

1. **(a) covers every tree; (b) covers only trees MoAI creates from now on.** The
   reporter's tree was made by hand, and existing trees would stay broken under (b).
2. **(a) adds no second copy of state.** (b)'s copy variant creates the drift
   problem it was meant to avoid; its link variant is the workaround the reporter
   rejected.
3. **(a) is one boundary.** The validator is the single choke point every tool
   already shares; (b) is four-plus creation paths that must stay in lockstep.
4. **Both is strictly worse than (a).** Once (a) accepts the worktree and routes
   `.moai` reads to the primary, a provisioned copy adds nothing but a second,
   diverging `.moai` root that (a) would then prefer (a worktree's own `.moai`
   wins), re-introducing the drift.

The operator decides at Implementation Kickoff; open questions are in `plan.md`
§C.

## §3 Requirements (GEARS)

- **REQ-MWU-001 (Ubiquitous, unchanged path).** The validator shall accept a
  `project_root` whose canonical path contains a `.moai` directory, and shall
  return that canonical path, exactly as it does today.

- **REQ-MWU-002 (Event-driven, linked-worktree acceptance).** When a
  `project_root` names a directory with no `.moai` directory, and that directory
  is the canonical top level of a linked git worktree registered in its
  repository, and the repository's primary checkout contains a `.moai`
  directory, the validator shall accept it and return the canonical worktree
  path as the tree root.

- **REQ-MWU-003 (Event-detected, rejection preserved).** When a `project_root`
  with no `.moai` directory fails any condition of REQ-MWU-002 — not a git
  directory, not a worktree top level (including a subdirectory of one), the
  primary checkout itself, a worktree not registered in the repository, or a
  repository whose primary checkout has no `.moai` — the validator shall reject
  it with an error that names the failed condition, and shall never substitute
  a default root.

- **REQ-MWU-004 (Event-detected, fail-closed).** When a git inspection required
  by REQ-MWU-002 cannot complete (git unavailable, a non-zero exit, or output of
  an unexpected shape), the validator shall reject the path rather than accept
  it or fall back.

- **REQ-MWU-005 (Ubiquitous, canonical containment).** The validator shall
  compare the candidate, the registered worktree paths, and the primary checkout
  path in symlink-canonical form, so that the verdict does not depend on the
  spelling through which a path reached it.

- **REQ-MWU-006 (Ubiquitous, environment isolation).** The git inspection shall
  be unaffected by repository-redirecting variables inherited from the server
  environment (`GIT_DIR`, `GIT_WORK_TREE`, `GIT_COMMON_DIR`, `GIT_INDEX_FILE`).

- **REQ-MWU-007 (State-driven, tree/`.moai` root split).** While a call's
  `project_root` was accepted through REQ-MWU-002, the tool shall act on the
  worktree for tree operations (diff collection, code graph queries, the working
  directory given to review backends), and shall resolve every `.moai`-relative
  read it performs on the call's behalf against the primary checkout's `.moai`.

- **REQ-MWU-008 (State-driven, provenance).** While a call's `project_root` was
  accepted through REQ-MWU-002, every response that carries the `_root`
  provenance block shall name both the tree root and the `.moai` root it used.

- **REQ-MWU-009 (Ubiquitous, absent parameter unchanged).** The resolution of an
  absent or empty `project_root` shall remain unchanged for both the fallback
  and the pass-through variants.

- **REQ-MWU-010 (Ubiquitous, documentation).** The `project_root` input
  descriptions and `.claude/rules/moai/core/moai-mcp-tools.md` § The
  `project_root` input shall state the linked-worktree acceptance and the
  tree/`.moai` split; the rule file and its template mirror under
  `internal/template/templates/` shall stay byte-identical, and the template copy
  shall carry no SPEC ID, card id, or date.

## §4 Constraints

- The reject-never-fall-back contract of the `project_root` input is binding: no
  path in this SPEC may turn a rejection into a silent default.
- The git subprocess runs only on the no-`.moai` path; the existing `.moai`
  path gains no subprocess.
- Template text stays neutral across the 16 supported programming languages and
  free of internal development state (no SPEC IDs, card ids, dates, commit SHAs).
- No creation path writes, copies, or links `.moai` as part of this SPEC.

## §5 Exclusions (What NOT to Build)

This section records what is out of scope for this SPEC.

### Out of Scope — design (b) creation-side provisioning

- Copying or linking `.moai` into worktrees from `moai worktree new`, the
  launcher materializers, the `WorktreeCreate` hook, or the L2 tree path.
- Any repair or re-provisioning of existing worktrees.

### Out of Scope — adjacent resolution defects

- Changing `resolveProjectDir()` or the absent-parameter fallback (owned by
  SPEC-MCP-WORKTREE-ROOT-001; its other consumers stay undecided).
- Merging a worktree's own `.moai` with the primary's when both exist — a
  worktree with its own `.moai` keeps today's behavior (REQ-MWU-001).
- Bare-repository layouts (a bare primary has no working tree, hence no `.moai`;
  such worktrees stay rejected).
- The base branch `moai worktree new` selects (observed on this card, tracked
  separately).

### Out of Scope — tools outside the `project_root` family

- MCP tools that take no `project_root` input, and the CLI verbs, which already
  read the tree they run in.
