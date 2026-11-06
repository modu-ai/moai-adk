---
id: SPEC-LANDING-SURFACE-RULES-001
title: "Landing-surface guidance: main lives only in one dedicated integration worktree, the primary checkout is the human and leader workspace, and card worktrees branch from the named ref main"
version: "0.2.0"
status: completed
created: 2026-10-17
updated: 2026-10-18
author: manager-spec (card t1846)
priority: P1
phase: "v3.2.0 target"
module: ".claude/rules, .claude/agents, .claude/skills, .codex/agents, internal/template/templates, AGENTS.local.md"
lifecycle: spec-anchored
tags: "landing-surface, integration-worktree, primary-checkout, staleness, worktree-base, guidance, template-first"
tier: M
related_specs:
  - SPEC-LOCAL-MAIN-FLOW-001
  - SPEC-WORKTREE-BASEREF-001
---

# SPEC-LANDING-SURFACE-RULES-001 (card t1846)

Card: **t1846** (Class C, plan -> run -> sync; operator delegation 2026-10-17; split 1 of 2 of the abandoned card t1830). Sibling cards: **t1847** (split 2 of 2, the mechanical half, ordered after this card; its SPEC id SPEC-LANDING-SURFACE-MECH-001 is a **forward reference** — no such directory exists in this tree until t1847 plans) and **t1831** (the workflow name; the order against this card is requested here, not recorded on the t1831 card — see §A.5). Evidence ledger ids (G0-G11, L1-L7) are defined in acceptance.md §E; every figure below was produced by a command run in the card tree at HEAD `7faf75fdfbe27e442ea2c17a1a90caf5792d09e6` (ledger G0).

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.2.0 | 2026-10-17 | Repair after plan-audit iteration 1 (FAIL 0.75): D1 the emitter writes the template tree only, so the deployed `.codex` toml is brought in line by an explicit copy step (REQ-LSR-005); D2 `internal/template/catalog.yaml` joins the file set as a generated artifact; D3 the scope fence is an allow-list diff; D4 the primary-as-surface set is ten sentences, not eight; D5 the local Base-Ref sentence names the fallback limit; D7 and D8 coordination and forward-reference disclosures; D9 the template-form rule carries a scope clause. |
| 0.1.0 | 2026-10-17 | Initial plan-phase authoring for card t1846. An earlier unverified draft of this SPEC (written by a stopped author run, never audited) was re-verified command by command: its 129-file grep figure was replaced by the measured 47 (ledger G1), its projection edits were dropped after the projection mechanism was measured (G7), and every citation was re-read. Scope is the guidance half of abandoned card t1830: items (A), (B-guidance) and (E-guidance). The predecessor SPEC-PRIMARY-LANDING-ONLY-001 (an abandoned SPEC, untracked in tree `.moai/worktrees/t1830`; no such directory exists in this tree) ended its plan audit at the iteration ceiling with score 0.88; its blocking residue was wording-level (N7, an overclaiming "race-safe" sentence) plus carried debt O6 to O9. This SPEC reuses its scope for A, B-guidance and E-guidance, drops every Go mechanism (moved to t1847) and the flow name (moved to t1831). |

---

## §A — Context

### A.1 What is observed in this tree

1. **Guidance still licenses the primary checkout as the landing surface.** Ten sentences state it: `.claude/rules/local/gitflow-lane-protocol.md` lines 33 (the `## 2.` heading "Integration surface — the primary checkout"), 35, 43, 137, 178 (ledger L1) and `AGENTS.local.md` lines 184, 192, 194 (the tool-step description "(3) confirms that the primary checkout is clean (item 3)"), 198, 204 (ledger L2). The first draft anchored eight; the audit found lines 33 and 194 unanchored, and re-running the grep gives five per file. The anchor phrases the work adds (`landing and resync only`, `git show main:`, `planned in t1847`, `named ref`, `named local ref`) occur nowhere in the guidance surfaces today (ledger L3, exit 1, empty stdout).
2. **The operator's layout already is the target layout.** `git for-each-ref refs/heads/main` reports branch `main` held by `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/main` (ledger G2), while `git worktree list --porcelain` reports the primary checkout detached at `478285ab4`. A reader who takes the primary checkout for current source reads a tree that is not `main`.
3. **The tools do not yet match the layout.** `moai integration resync` is described as "Fast-forward the primary checkout's local main to origin/main" (`internal/cli/local_main_resync.go:91`) and refuses when another tree holds `main` ("The re-sync applies to the primary checkout only", `local_main_resync.go:361`); the merge resolver accepts the primary as holder only while `workflow.local_main_integration.enabled` is true (`internal/cli/integration_merge.go:145-157`), and that key is `true` in `.moai/config/sections/workflow.yaml:195` today (ledger G8). No verb moves a stale or detached primary checkout to the current local `main` tip (the verb list of `moai integration --help` has `resync`, not a follow step). Guidance written here therefore marks every such mechanism "planned in t1847".
4. **A card tree can start on the wrong base.** `internal/cli/session_worktree.go:236-237` reads `git_strategy.worktree_base_branch` and, when it is unresolvable, creation falls back to the no-operand form that cuts from the invoking tree's HEAD; Claude Code's own `--worktree` branches from `origin/HEAD` (`.claude/rules/moai/workflow/worktree-integration-ops.md:277`). `.moai/config/sections/git-strategy.yaml:5` sets `worktree_base_branch: main` in this repository.
5. **The distributed template is closer to the rule than the local files.** The template sources already say the integration branch is "checked out in exactly one worktree" (`sync/delivery.md:275`) and name "the dedicated integration worktree" (`manager-git.md:120`); what they lack is the statement that this tree performs landing and resync only, stays clean, and that the primary checkout does not hold the integration branch.

### A.2 The Integration-Worktree Rule (canonical wording)

Local form (repo-local files): "`main` is checked out in exactly one dedicated integration worktree, at a fixed path (currently `.claude/worktrees/main`). That worktree performs landing and resync only: it is never used to edit or experiment, and it stays clean. The primary checkout is the human and leader workspace and does not check out `main`."

Template form (distributed files): "Where the integration branch is held by a dedicated worktree (the git-flow `WT-*` integration route), it is checked out in exactly one dedicated worktree, at a fixed path. That worktree performs landing and resync only: it is never used to edit or experiment, and it stays clean. The primary checkout is the human and leader workspace and does not check out the integration branch. Routes that commit directly to the branch (Route A) or that run no dedicated integration worktree are outside this sentence."

The scope clause (`Where the integration branch is held by a dedicated worktree`) exists because `spec-workflow.md` also describes Route A, where manager-develop commits directly on `main` in the checkout (the default for Tier S and M), and `sync/delivery.md:324` tells non-`WT-*` git-flow branches to `git checkout main`; an unconditioned sentence would contradict both. The rule is placed only in the git-flow `WT-*` passages named in plan.md §B.1.

Anchor phrases for acceptance: `landing and resync only` (both forms), `is the human and leader workspace and does not check out` (both forms), `Where the integration branch is held by a dedicated worktree` (template form).

### A.3 The Base-Ref Clause (canonical wording)

Local form: "A card worktree is created only through `moai factory next` or `moai worktree new`. When the configured base (`git_strategy.worktree_base_branch`, `main` here) resolves, it branches from the named ref `main` and not from HEAD; when the base is empty or does not resolve, creation falls back to the invoking tree's HEAD (`internal/cli/session_worktree.go:236-237`). Claude Code's default `--worktree` (which branches from the remote default branch) is not used for card work. After creation, confirm that the new tree's HEAD equals `main`; that confirmation, not the creation path, is the check."

Template form: "A card or team worktree is created through the MoAI creation paths and branches from the named local ref of the integration branch, never from HEAD; `claude --worktree`, which branches from the remote default branch, is not used for it. This is a convention the creator confirms, not a check: creation can fall back to the invoking tree's HEAD when the configured base cannot be resolved, so confirm the new tree's HEAD after creation. The ref is read when the tree is created; the tree does not follow later moves of the branch."

Anchor phrases: `branches from the named ref` and `falls back to the invoking tree` (local, both required), `named local ref` (template).

### A.4 Staleness and non-guarantee wording (items B-guidance, N7)

The local guidance tells readers to read current `main` source through `git show main:<path>` or inside the integration worktree, never from the primary checkout's working tree, and to keep the primary clean by diagnosing each untracked or modified file there. Every sentence about what `git show` or a checkout protects names a limit: it reads the committed blob only, so it does not reproduce a smudge or clean filter or end-of-line conversion a working-tree copy carries, shows no uncommitted change, and says nothing about a later move of `main`. Anchor phrases for acceptance: `git show main:`, `smudge`, `each untracked or modified file`, `no moai verb moves the primary checkout to the current local`, `planned in t1847`.

### A.5 Coordination with t1831 and t1847

Both t1846 and t1831 edit `AGENTS.local.md` and `.claude/rules/local/gitflow-lane-protocol.md`, and the overlap is wider than those two files: `moai todo show t1831` (read in this session) also names "배포 규칙·스킬·docs의 git-flow/GitHub Flow 운영 서술" under Template-First, which includes the template sources this card edits. The card text of t1831 records **no ordering** against t1846; both cards are `picked`. This SPEC therefore **requests** the order (t1846 first) and asks the leader to record it on t1831 (progress.md §E.1 carries the note). Under that order, t1831 renames flow wording in the shared files by rebasing onto this card's text and keeping both. This card rewrites the landing-surface sentences (`AGENTS.local.md` §4.0 intro and items 1 to 4; protocol §1, §2, §9, §11), adds the (B) and (E) guidance, and edits the template sources of plan.md §B.1. It writes no flow name and touches no `git_strategy.manual.workflow` text. t1847 is ordered after this card and implements the mechanisms the local guidance marks "planned in t1847"; when it lands it replaces those markers.

---

## §B — Requirements (GEARS)

12 requirements (Tier M ceiling 16). Each REQ line keeps `shall` on its first physical line. Acceptance criteria live in acceptance.md and map by `maps REQ-LSR-NNN`.

- REQ-LSR-001: The repo-local protocol file `.claude/rules/local/gitflow-lane-protocol.md` shall carry the Integration-Worktree Rule (local form) in its §2 and shall no longer name the primary checkout as the landing surface in the §2 heading and body or as the rc build tree in §9, nor describe the primary checkout as the place the resync tool checks in §11 (the five sentences of ledger L1, including the `## 2.` heading at line 33).
- REQ-LSR-002: The repo-local file `AGENTS.local.md` shall carry the Integration-Worktree Rule (local form) in §4.0 and shall no longer state that card work lands in the primary checkout, that the landing condition is a primary HEAD on `main`, that the primary checkout is the place that must be clean to land (items 2 and 3, including the tool-step description at line 194, which for a separate integration worktree is replaced by the overlap decision of `internal/factory/integration_merge_step.go:309-310`), or that the resync tool checks the primary checkout in item 4 (the five sentences of ledger L2).
- REQ-LSR-003: Each of the five template sources (`manager-git.md`, `delivery-policy.md`, `spec-workflow.md`, `worktree-integration.md`, `sync/delivery.md`) shall carry the Integration-Worktree Rule (template form), phrased for "the integration branch" and opened by the scope clause "Where the integration branch is held by a dedicated worktree", placed only in the git-flow `WT-*` passages of plan.md §B.1, and shall name no card id, SPEC id, date, or mechanism that does not exist in this tree.
- REQ-LSR-004: When a template source is edited, its deployed copy under `.claude/` shall carry the identical change, so that every pair that is byte-identical at HEAD stays byte-identical and the `worktree-integration-ops.md` pair keeps exactly its recorded difference (numstat `9 9`, ledger G5).
- REQ-LSR-005: When `internal/template/templates/.claude/agents/moai/manager-git.md` or any other template source changes, the template copy `internal/template/templates/.codex/agents/moai/manager-git.toml` shall be regenerated through `make agents-emit` (the emitter reads and writes the template tree only, ledger G10), the deployed copy `.codex/agents/moai/manager-git.toml` shall be brought in line by a byte copy of that regenerated file, `make build` shall regenerate `internal/template/catalog.yaml` (whose entries at lines 8 and 386 hash the edited skill directory and agent file), and `make agents-emit-check`, `make embed-manifest-check` and the six named template tests of acceptance.md AC-LSR-008 shall pass.
- REQ-LSR-006: The repo-local guidance (`AGENTS.local.md` and the protocol file) shall tell readers to read current `main` source through `git show main:<path>` or inside the integration worktree and never from the primary checkout's working tree, and shall state what that read does not cover: smudge or clean filters, end-of-line conversion, uncommitted changes, and later moves of `main`.
- REQ-LSR-007: The repo-local guidance shall state that the primary checkout is kept clean, with each untracked or modified file there diagnosed (committed in the card worktree that owns it, moved to that worktree, or covered by an ignore rule) rather than left in place, and shall state that no moai verb moves the primary checkout to the current local `main` tip, marking a primary-follow step "planned in t1847".
- REQ-LSR-008: The repo-local guidance (`AGENTS.local.md` §4.0 item 1 and the protocol file's §1) shall carry the Base-Ref Clause (local form), including the limit that creation falls back to the invoking tree's HEAD when the configured base does not resolve, and shall keep the existing check that a new card tree's HEAD and local `main` point to the same commit.
- REQ-LSR-009: The template sources `worktree-integration-ops.md` (the `worktree.baseRef` section) and `worktree-integration.md` (the `claude --worktree` behavior list), and their deployed copies, shall carry the Base-Ref Clause (template form) without removing the existing statement that `claude --worktree` branches from the default remote branch.
- REQ-LSR-010: Where the repo-local guidance names a behavior that no moai verb, setting or hook provides in this tree (the primary-follow step, the statusline landing-surface display, the designated-path resolution of the integration worktree, the lease-time base warning, a resync run in the dedicated worktree), it shall mark the sentence "planned in t1847" and shall not claim enforcement; the template sources and their deployed copies shall name no such behavior.
- REQ-LSR-011: The guidance edits shall contain no guarantee wording ("race-safe", "guaranteed", "always safe"), and every sentence that states what a git operation protects shall name a limit the operation does not cover.
- REQ-LSR-012: While this SPEC is in its run phase, the set of files changed between the plan pin and HEAD shall be a subset of the 17 named files of plan.md §B.7 plus the SPEC directory and `.moai/reports/t1846/` (so no Go file, configuration file, hook, settings file, statusline source, `cmd/` file, or file under `.moai/policies/` or `.moai/workflows/` changes), and the workflow name `agentic-trunk` shall not appear in an edited file.

---

## §C — Out of Scope

### Out of Scope — mechanical enforcement (card t1847, SPEC-LANDING-SURFACE-MECH-001)
- Any Go change; the verb that moves the primary checkout to the `main` tip; the statusline landing-surface display; the configuration key that designates the integration worktree path; refusals and warnings in `moai integration` and `factory complete`; the lease-time base warning; `worktree.baseRef` and `worktree_base_branch` handling; changing `workflow.local_main_integration.enabled`.
- Guidance written here describes those mechanisms only as "planned in t1847" in the repo-local files, never as existing and never as enforced.

### Out of Scope — the workflow name (card t1831)
- Adding the `agentic-trunk` value, changing `git_strategy.manual.workflow`, and renaming flow wording in guidance. This card adds no flow name to any file.

### Out of Scope — derived projections
- `.moai/policies/**` and `.moai/workflows/**` are written by the deployer from `.claude/rules/moai` and `.claude/skills/moai/workflows` (`internal/cli/update_reconcile.go:71`, `internal/cli/migrate_render.go:36`; ledger G7). They are refreshed by `moai update`, not hand-edited by this card; plan.md OQ-1 records the alternative.

### Out of Scope — adjacent surfaces, classified in plan.md §B.5
- `parallel-advanced.md` and `parallel-workflows.md` skill modules (their "integration worktree" phrases are a flag-test scratch tree and a team contract-integration tree), `.claude/agents/harness/hns-release-specialist.md` (develop-era, marked superseded), `.moai/docs/**` doctrine documents, `CHANGELOG.md`, Go sources and test fixtures, `reports/` data files, and historical SPEC and report bodies.

### Out of Scope — the operator's own terminal work
- Creating, moving or removing the integration worktree, branch-state changes in the primary checkout or the integration worktree, and operations runs and card cleanup.
