---
id: SPEC-MCP-WORKTREE-UNTRACKED-001
title: "Accept a linked worktree as project_root when the repository keeps .moai/ out of git"
version: "0.2.0"
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
| 0.2.0 | 2026-09-26 | manager-spec (t1202) | Revision for plan-audit iter-1 (FAIL 0.62, defects D1–D16). Adds the measured per-tool `.moai` read/write inventory (§3), splits REQ-007 into four access classes, defines the `.moai` source-root predicate and the primary-checkout predicate, corrects the design (b) premise against issue #1716, and rewrites the requirements under the recommended defaults of the operator decisions listed in `plan.md` §C. |

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

1. Pass `project_root` → rejected. The audit, SPEC, verify, and graph tools are
   unusable from that worktree.
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

When the named path has no MoAI configuration of its own, the validator asks git
whether the path is the top level of a linked worktree **registered** in a
repository whose **primary checkout** (predicate in REQ-MWU-004) carries a MoAI
configuration. If so, the path is accepted. An unrelated directory, a non-MoAI
repository, a subdirectory, or an ambiguous repository layout still fails.

The check is a registration check (the candidate must be listed by
`git worktree list --porcelain`, not flagged prunable) plus a common-dir check
(the candidate's git common dir must be the primary's). It is not a filesystem
containment check. Every path compared is compared in symlink-canonical form.

What design (a) must also decide is **where each `.moai` access goes** once the
worktree itself has no `.moai`. The measured inventory in §3 shows four access
classes with different correct answers; a single "read everything from the
primary" rule would return the primary tree's code graph as the worktree's
answer, and a single "use the worktree" rule would return empty catalogues as
success. §4 states the rule per class. The guiding principle: a worktree of an
untracked-`.moai` repository behaves as a worktree of a tracked-`.moai`
repository would — **configuration and catalogue come from the shared project
copy; state and derived artifacts stay with the tree that produced them**.

| Property | Assessment |
|---|---|
| Covers trees made by `moai worktree new` (the issue's reproduction path) | Yes |
| Covers trees made by hand with `git worktree add` | Yes |
| Covers trees that already exist today | Yes — no re-creation needed |
| Introduces a second copy of `.moai` configuration | No |
| New cost | A scrubbed git subprocess on the no-configuration path only; fail-closed on error |

### §2.2 Design (b) — worktree creation provisions `.moai`

Every creation path copies or links `.moai` into the new tree, so the existing
validator passes unchanged. **Issue #1716 names this as an acceptable expected
behavior** ("provide the needed project settings safely when the worktree is
created"), and its reproduction starts from `moai worktree new` — a MoAI
creation path that (b) would cover. (b) is therefore a legitimate option, not a
misreading of the report.

Measured facts about (b) (evidence commands in `plan.md` §B):

| Property | Measured assessment |
|---|---|
| `git worktree add` execution sites in the tree | 2 — `internal/core/git/worktree.go` (the `Add` argument builder) and `internal/cli/session_worktree.go` (the session materializer) |
| Creation entry points reaching them | 3 — `moai worktree new` (via the session materializer), the launcher `-w --branch` path, and the `WorktreeCreate` hook serving `EnterWorktree` / agent isolation |
| Trees made by hand with `git worktree add` | Not covered — no MoAI code runs |
| Trees that already exist | Not covered — each needs separate provisioning |
| Copy variant | Two copies of configuration and SPECs; an edit in one is not seen by the other |
| Link variant | Writes through the link land in the primary tree; the issue reporter deliberately avoided this workaround |

Unmeasured and therefore not used as a reason: whether copied runtime files
would carry credentials or machine-specific paths.

### §2.3 Recommendation — design (a) alone

Recommend **design (a)**, and do **not** implement (b) alongside it. Reasons,
each resting on a measured fact above:

1. **Coverage.** (a) covers every tree, including existing and hand-made ones;
   (b) covers only trees a MoAI entry point creates from now on.
2. **One copy.** (a) introduces no second configuration copy; (b)'s copy variant
   does, and its link variant is the workaround the reporter avoided.
3. **One boundary.** (a) changes the validator every `project_root` tool already
   shares plus the per-class routing in §4; (b) changes three entry points and
   still needs (a)'s routing for existing trees.
4. **Both is worse than (a).** A provisioned configuration copy would satisfy the
   §4 source-root predicate and become the source root, re-introducing drift
   between the worktree copy and the project copy.

Issue #1716 lists (b) as acceptable; the operator decides at Implementation
Kickoff (`plan.md` §C, decision 6).

## §3 Measured `.moai` access inventory (`project_root` family)

Measured on base `df526c9a9` by reading each handler and the package function it
calls (commands in `plan.md` §B). Class key: **T** tree operation, **C**
configuration or catalogue read, **G** derived graph artifact read, **S** state
write (and the read that pairs with it).

| Tool | Access | Path under `.moai` | Class | Code site |
|---|---|---|---|---|
| `spec_progress` | read | `specs/SPEC-*/spec.md` | C | `mcp_server.go` → `spec.ListDocs` (`internal/spec/listdocs.go`) |
| `spec_audit`, `spec_drift` | read | `specs/SPEC-*/…` | C | `mcp_server.go` → `spec.Audit` (`internal/spec/audit.go`) |
| `verify_snapshot` (load), `verify_trend` | read | `state/verify/snapshots/<key>.json` | S | `internal/verify/store.go` `SnapshotDir` |
| `verify_snapshot` (record) | **write** | `state/verify/snapshots/<key>.json` | S | `verify.RecordCheck` |
| `graph_file_api`, `graph_find_code`, `graph_shortest_path`, `graph_trace_calls` | read | `project/graph/edges.jsonl` + meta sidecar | G | `internal/graph/codequery.go` `edgesArtifactPath`, `AnswerProvenance` |
| `codex_audit` | review cwd | (git tree) | T | `mcp_codex.go` handler `params["cwd"] = root` |
| `codex_audit` | read | `config/sections/llm.yaml`, `config/sections/workflow.yaml` (audit pin, gates) | C | `codexSSOTModelEffort`, `workflowAuditPins`, `applyGateUnmet` |
| `codex_audit` | **write** | `state/audit-receipts/…` | S | `recordAuditReceipt` → `auditreceipt.WriteReceipt` |
| `codex_audit` | read | `config/sections/workflow.yaml` (gate-required check for the receipt id) | C | `auditreceipt.CodexGateRequired(root)` |
| `claude_audit` | diff collection | (git tree) | T | `performClaudeAuditWith` → `collectReviewDiff(root, target)` |
| `claude_audit` | read | `config/sections/workflow.yaml` (audit pin) | C | `workflowAuditPins(root).Claude` |
| `glm_audit` | diff collection | (git tree) | T | `mcp_glm.go` handler, root from `resolveToolProjectRoot` |
| `glm_audit` | read | `config/sections/workflow.yaml` (audit pin) | C | `workflowAuditPins(root).GLM` |
| `glm_audit` | read | `config/sections/llm.yaml` | C (server root) | `mcp_glm.go` — reads `projectDirResolver()`, not `project_root` (pre-existing) |
| `audit_multi` | review cwd / diff per backend | (git tree) | T | `mcp_audit_multi.go` → `mcp_convergence.go` |
| `audit_multi` | read | `config/sections/workflow.yaml` (gates) | C | `workflowAuditGates(root)` |
| `audit_multi` | **write** | `state/audit-multi/<session>.json` | S | `persistConvergenceResult` → `convergenceStateDirFor(root)` |

Readers of the S-class writes key on the **worktree**, not on the primary: the
auditor receipt guard resolves its tree with `auditreceipt.TreeRootFromCWD`
(the caller's git top level), and `persistConvergenceResult`'s own comment
records a prior defect where a worktree run that wrote into the primary's
`.moai/state` left its verdict where the worktree's gate never looked.

The inventory is a lower bound for code reached through package functions:
`spec.Audit` and `verify` bodies were read at their `.moai` join sites, not
traced line by line (`plan.md` §B, Gaps).

## §4 Requirements (GEARS)

Requirements are written under the recommended default of every operator
decision in `plan.md` §C; each decision names the requirements it would change.

Definitions used below:

- **MoAI configuration** of a directory `D`: `D/.moai/config/sections` exists as
  a directory.
- **Tree root**: the canonical path the caller named.
- **Source root**: the tree root when it has MoAI configuration; otherwise the
  primary checkout (REQ-MWU-004).

### §4.1 Acceptance

- **REQ-MWU-001 (Ubiquitous, unchanged path).** The validator shall accept a
  `project_root` whose canonical path has MoAI configuration, and shall return
  that canonical path as both tree root and source root, exactly as it accepts
  such a path today.

- **REQ-MWU-002 (Event-driven, linked-worktree acceptance).** When a
  `project_root` names a directory without MoAI configuration, and that
  directory is the canonical top level of a linked git worktree listed by
  `git worktree list --porcelain` without a `prunable` mark, and the primary
  checkout of its repository has MoAI configuration, the validator shall accept
  it with the worktree as tree root and the primary checkout as source root.

- **REQ-MWU-003 (Event-driven, rejection preserved).** When a `project_root`
  without MoAI configuration fails any condition of REQ-MWU-002 — not inside a
  git repository, not a worktree top level (including a subdirectory of one),
  the primary checkout itself, a worktree that is not listed or is marked
  prunable, or a repository whose primary checkout has no MoAI configuration —
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
  The validator shall compare the candidate, every listed worktree path, and the
  primary path in symlink-canonical form, and shall run every git inspection
  with `GIT_DIR`, `GIT_WORK_TREE`, `GIT_COMMON_DIR`, `GIT_INDEX_FILE`, and
  `GIT_CEILING_DIRECTORIES` removed from the child environment.

### §4.2 Per-class routing (while accepted through REQ-MWU-002)

- **REQ-MWU-007 (State-driven, tree operations).** While a call's `project_root`
  was accepted through REQ-MWU-002, every class-T operation in §3 shall act on
  the tree root.

- **REQ-MWU-008 (State-driven, configuration and catalogue).** While a call's
  `project_root` was accepted through REQ-MWU-002, every class-C read in §3
  whose input is the call's root shall resolve against the source root. (The
  `glm_audit` `llm.yaml` read, which ignores `project_root` today, is unchanged.)

- **REQ-MWU-009 (State-driven, derived graph).** While a call's `project_root`
  was accepted through REQ-MWU-002, every class-G read shall resolve against the
  tree root only; where the tree root holds no graph artifact the tool shall
  return its existing "graph layer absent" error, and shall never answer from
  the primary checkout's graph.

- **REQ-MWU-010 (State-driven, state writes).** While a call's `project_root`
  was accepted through REQ-MWU-002, every class-S write and its paired read
  shall resolve against the tree root, so the readers that key on the worktree
  find it.

- **REQ-MWU-011 (Ubiquitous, no resolution flip).** The source-root predicate
  shall depend only on MoAI configuration, so that state or graph directories
  created under the tree root by REQ-MWU-009/010 shall not change the source
  root of later calls.

- **REQ-MWU-012 (State-driven, provenance).** While a call's `project_root` was
  accepted through REQ-MWU-002, every response carrying the `_root` provenance
  block shall name both the tree root and the source root.

### §4.3 Compatibility and documentation

- **REQ-MWU-013 (Ubiquitous, absent parameter unchanged).** The resolution of an
  absent or empty `project_root` shall remain unchanged for both the fallback
  and the pass-through variants.

- **REQ-MWU-014 (Ubiquitous, documentation).** The shared input description text
  (`projectRootDescCommon`) and `.claude/rules/moai/core/moai-mcp-tools.md`
  § The `project_root` input shall state the linked-worktree acceptance and the
  per-class routing; the rule file and its template mirror under
  `internal/template/templates/` shall stay byte-identical, and the template
  copy shall carry no SPEC ID, card id, or date.

## §5 Constraints

- The reject-never-fall-back contract of the `project_root` input is binding: no
  path in this SPEC turns a rejection into a silent default.
- The git subprocess runs only on the no-configuration path; the existing path
  gains no subprocess.
- Template text stays neutral across the 16 supported programming languages and
  free of internal development state (no SPEC IDs, card ids, dates, commit SHAs).
- No creation path writes, copies, or links `.moai` as part of this SPEC.

## §6 Exclusions (What NOT to Build)

This section records what is out of scope for this SPEC.

### Out of Scope — design (b) creation-side provisioning

- Copying or linking `.moai` into worktrees from `moai worktree new`, the
  launcher `--branch` path, or the `WorktreeCreate` hook.
- Any repair or re-provisioning of existing worktrees.

### Out of Scope — ambiguous repository layouts

- `--separate-git-dir` repositories, submodule-internal git dirs, and bare
  repositories: rejected by REQ-MWU-004, not supported.

### Out of Scope — adjacent resolution behavior

- Changing `resolveProjectDir()` or the absent-parameter fallback (owned by
  SPEC-MCP-WORKTREE-ROOT-001).
- Merging a worktree-local untracked `.moai/specs` into the catalogue read from
  the source root (`plan.md` §C decision 5).
- Building the code graph automatically for a worktree (`plan.md` §C decision 4).
- The `glm_audit` `llm.yaml` read that ignores `project_root` (pre-existing).
- The base branch `moai worktree new` selects (observed on this card, tracked
  separately).

### Out of Scope — tools outside the `project_root` family

- MCP tools that take no `project_root` input (including `codex_role_audit`),
  and the CLI verbs, which read the tree they run in.
