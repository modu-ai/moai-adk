---
id: SPEC-WORKTREE-HOME-ROOT-001
title: "Unify MoAI worktree storage under a home-level project root (~/.moai/worktrees/projects/<project-key>/<name>) with a non-destructive migration of MoAI-made trees"
version: "0.1.0"
status: draft
created: 2026-10-19
updated: 2026-10-19
author: manager-spec (card t1640)
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/cli/worktree, internal/hook, internal/factory, internal/homestate, internal/config, internal/paths, internal/template/templates, AGENTS.md, .claude/rules/moai/workflow"
lifecycle: spec-anchored
tier: L
tags: "worktree, storage-location, home-root, migration, non-destructive, launcher, enterworktree, codex-cd, factory-lease, tier-by-class, card-t1640"
related_specs:
  - SPEC-WORKTREE-DONE-TIER-001
  - SPEC-WORKTREE-ENTRY-STRATEGY-001
  - SPEC-WORKTREE-BASEREF-001
  - SPEC-SESSION-WORKTREE-001
  - SPEC-FACTORY-SELF-DISPATCH-001
  - SPEC-SESSION-ANCHOR-ATTR-001
  - SPEC-DUAL-HARNESS-RECOVERY-001
  - SPEC-HANDOFF-NEUTRAL-001
  - SPEC-UPDATE-MIGRATION-FIX-001
---

# SPEC-WORKTREE-HOME-ROOT-001 — Home-level worktree root for MoAI-made trees

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-10-19 | manager-spec (card t1640) | Initial plan-phase emission, Tier L. 24 GEARS requirements, 24 Given-When-Then acceptance criteria, 9 operator decisions with recommended defaults, 6 pre-run measurements. All code citations measured in worktree `.moai/worktrees/t1640` at HEAD `74ad197380` (branch `WT-d2-moai-worktrees`). Absorbs cards t1611 (storage-location move design), t1610 (launcher/EnterWorktree guidance correction, Codex `/cd` facts) and t1501 (first cleaner, in the part this SPEC owns: classification and non-destructive retirement). Binding inputs: decision records `d-20261018T221252Z-cadf` (operator: lift deferral, target layout, non-destructive migration), `d-20261018T221922Z-dab5` (operator: Claude-native `agent-*` trees stay; no WorktreeCreate hook), `d-20261018T221839Z-ef12` (leader research note, non-binding defaults). |

## §A. Problem Statement

MoAI-made worktrees live inside the repository directory: `moai worktree new` and the factory card materializer create `<root>/.moai/worktrees/<name>` (`internal/cli/session_worktree.go:48,223`), while the `moai cc -w <name>` short-name path and the `-w <name> --branch <existing>` path land in `<root>/.claude/worktrees/<name>` (`internal/cli/worktree_branch_flag.go:109`; Claude Code creates the tree itself for a bare short name). Measured at HEAD `74ad197380` with `git worktree list`: 251 registered entries — 1 primary checkout, 192 under `<root>/.moai/worktrees`, 46 under `<root>/.claude/worktrees` (31 named `agent-<hex>`, 15 other), 12 under `~/.codex/worktrees` (Codex-app managed). Operator measurement 10-19 recorded 191 and 46 for the first two; the one-tree difference is this card's own tree (research.md §R1).

Consequences the operator and the cards name: the working copies share a directory subtree with the repository they belong to (scanners, ignore lists and path-prefix classifiers all special-case `worktrees`), 238 trees hold branches that may exist nowhere else, and the storage convention is split across two in-repo roots plus a third, never-used home root (`~/.moai/worktrees/<project>`, the old "L2" layout, which now holds only seven empty or near-empty directories). The operator decision (cadf, 10-19) lifts the 3.3 deferral: unify the location now.

Two hazards make this more than a path change, and the SPEC is built around them:

1. **Trees hold the only copy of unpushed work.** Any migration that is not strictly non-destructive can destroy the work of the 191-192 live trees (operator invariant, cadf).
2. **Many readers identify trees by path shape.** A tree moved to `~/.moai/worktrees/projects/<key>/<name>` is misread by at least five concrete readers unless they change (research.md §R3): the card-id-from-path derivations resolve the card id to the literal `projects`; the `todo --auto` owner-liveness probe looks only at two in-repo candidate paths and would judge a live owner dead; the `done`/`sweep` tier classifier would reclassify a moved session tree as an L2 registry tree (a disposal-eligible class); the factory card status bases and the launcher short-name resolvers do not know the new root; and the factory card record stores an absolute `worktree_path` that the record layer refuses to change.

## §B. Scope

In scope (components):

1. **Location resolver** — one function family that maps (repository, name) to the home-root path, owns the project-key derivation, the collision guard and the legacy-root candidates (REQ-WHR-001..004, 022, 023).
2. **Creation paths** — every MoAI creation path places new trees at the home root; none creates under a legacy root (REQ-WHR-005, 006).
3. **Readers** — card-id derivation, owner liveness, status bases, launcher/anchor/hook-seed short-name resolution, tier classification, `moai worktree list` (REQ-WHR-007..010).
4. **Stored-path migration** — card record rebind and the other path-bearing state (REQ-WHR-011, 012).
5. **Migration engine** — `moai worktree migrate` (dry-run default, apply, verify, rollback) and the adoption of the same classifier by the `moai update` consolidation (REQ-WHR-013..019, 024).
6. **Retirement** — legacy in-repo roots stop being written; the empty MoAI legacy directory is removed; the Claude-owned directory is never touched (REQ-WHR-020).
7. **Guidance (t1610 absorbed part)** — AGENTS.md, rules, skills, agents and template mirrors carry the launcher-vs-in-session distinction and the new location, with measured entry behavior on both harnesses (REQ-WHR-021, 022 — see §C numbering). The wider EnterWorktree/in-session-refusal guidance rewrite belongs to card t1879, which is ordered after this card so the docs are rewritten once.
8. **Pre-run measurement gate** — EnterWorktree and Codex `/cd` entry into a home-root tree, measured before the run phase starts (REQ-WHR-021 gate clause; §F).

## §C. Requirements (GEARS notation)

The `<subject>` is generalized per the GEARS guide. "Home root" means `<MoaiHome>/worktrees/projects/<project-key>`; "legacy roots" means `<primary>/.moai/worktrees` and `<primary>/.claude/worktrees`; "moai-made tree" is defined by REQ-WHR-013.

### Location and key

**REQ-WHR-001** (Ubiquitous — home root). The location resolver shall return, for a repository and a tree name, the path `<MoaiHome>/worktrees/projects/<project-key>/<name>`, where `<MoaiHome>` is obtained only through `paths.MoaiHome()` so that `MOAI_HOME` relocates the whole tree and tests stay hermetic.

**REQ-WHR-002** (Ubiquitous — project key). The location resolver shall derive `<project-key>` from the repository's primary checkout, never from the invoking worktree: the key is the sanitized basename of the primary checkout root (characters outside `[A-Za-z0-9._-]` become `-`), disambiguated per REQ-WHR-003 when another repository already owns that name, and recorded in a marker file `<MoaiHome>/worktrees/projects/<project-key>/.moai-project.json` holding the schema version, the primary checkout root, and the creation time.

**REQ-WHR-003** (Event-detected — collision guard). When the marker of the candidate key records a primary checkout root that is not the same directory as the requesting repository's root (compared per REQ-WHR-023), the location resolver shall select the deterministic disambiguated key `<project-name>-<first 8 hex of SHA-256 of the case-folded canonical primary root>`, reuse an existing marker that already names this repository when one exists, and shall not place a tree under the other repository's key; when a marker exists but cannot be read or parsed, creation shall refuse with a message naming the marker path.

**REQ-WHR-004** (Where — operator pin). Where `git_strategy.worktree_project_key` is set to a non-empty value, the location resolver shall use that value (validated as a single safe path segment) as `<project-key>` after the same marker guard, so that a renamed or moved repository keeps its trees under one key. The default is empty, which reproduces REQ-WHR-002 exactly.

### Creation

**REQ-WHR-005** (Event-driven — creation paths). When any MoAI creation path creates a worktree — `moai worktree new`, the factory card materializer (`factory next` and its three other callers), the session-worktree auto-entry, and the launcher `-w <name> --branch <existing>` path — the system shall create it at the home root through the one shared materializer, preserving that materializer's existing guarantees (configured base branch, `WT-` branch convention, global `safe.directory` registration, worktree-scoped identity, Codex hook seeding, existing-destination refusal). No creation path shall create a tree under a legacy root.

**REQ-WHR-006** (Event-driven — launcher short name). When `moai cc -w <name>` or `moai glm -w <name>` receives a short name that resolves to no existing tree in the home root or a legacy root, the launcher shall materialize the tree at the home root through the shared materializer and hand the backend the absolute path, instead of letting Claude Code create `.claude/worktrees/<name>`; a short name that resolves to an existing tree enters that tree. This requirement is conditional on pre-run measurement PM-3 and operator decision D-5 (§E); if either fails, the launcher short-name path is reported as the single documented exception to "no new tree outside the root" and the SPEC is amended before the run phase.

### Readers

**REQ-WHR-007** (Ubiquitous — path-independent identity). Every component that derives a card or tree identity from a directory path, or looks a tree up by card id, shall resolve it through the shared location resolver (home root first, then the legacy roots while the transition lasts), with the recorded card `worktree_path` taking precedence over any derived candidate. The components are: the SessionStart card-id derivations (`cardIDFromRegisteredWorktree`, `cardIDFromPath`), the `todo --auto` owner-liveness probe (`autoOwnerWorktrees`), the factory card status bases (`DefaultWorktreeBases`), the launcher short-name resolvers (`launch_chain`, `factory_anchor_guard`, `codex_hooks_seed`, `ccWorktreeWriterPrecheck`), `resolveCodexWorktreeDir`, and `cleanupMoaiWorktrees`.

**REQ-WHR-008** (Unwanted — misidentification). A tree whose parent segment is the project key shall not resolve to a card id equal to `projects` or to the project key; a card whose recorded tree is outside every legacy root shall not be judged "owner not alive" because no in-repo candidate path exists.

**REQ-WHR-009** (Ubiquitous — tier by creation class). The system shall classify a worktree as an L1 session tree or an L2 registry tree by how it was created, never by whether its path lies under a fixed in-repo prefix: every moai-made tree at the home root is an L1 tree, `moai worktree done` shall refuse it with or without `--force`, and `moai worktree sweep` shall report it as tier L1 and dispose of it only on its existing landing predicates. The legacy in-repo roots remain L1 for as long as trees exist there.

**REQ-WHR-010** (Ubiquitous — `moai worktree list`). The system shall provide `moai worktree list [--json]`, read-only, that enumerates every worktree registered with the repository and reports for each: name, absolute path, root kind (home / legacy-moai / legacy-claude / codex-managed / other), class (moai-made / claude-native / codex-managed / integration / primary / unclassified), branch, short HEAD, dirty flag, commits not present on any remote, and whether a live session is anchored. Two spellings of one directory shall be reported as one tree (REQ-WHR-023).

### Stored paths

**REQ-WHR-011** (Event-driven — card record rebind). When migration moves a tree that a factory card record names, the system shall rebind the card's recorded `worktree_path` from the old path to the new path through a dedicated record operation that (a) accepts only the transition `old → new` for the card whose current value equals `old`, (b) verifies that `new` is a registered worktree of the same repository at the same HEAD and branch, (c) updates every run row of that card that carries `old`, so that `PreviousCardWorktree` cannot carry a stale path into a replacement run, and (d) appends one `card.fields` event recording both paths. The existing refusal to move a card onto a different tree through `RecordCardWorktree` shall remain unchanged.

**REQ-WHR-012** (Event-driven — other path state). When migration moves a tree, the system shall re-key the entry of that tree in the project worktree state file (`.moai/state/worktrees.json`), leave the session registry and the origin-trail chain ledger entries that name the old path unmodified (live sessions are skipped by REQ-WHR-016; the ledger is history), and record in the journal (REQ-WHR-017) every path-bearing store it did or did not touch.

### Migration

**REQ-WHR-013** (Ubiquitous — moai-made classification). The system shall classify a registered tree as moai-made when it is neither the primary checkout, nor an integration worktree (the tree that holds the configured integration branch), nor a Claude-native agent-isolation tree (directory name matching `agent-[0-9a-f]+`), nor a Codex-managed tree, and either its directory lies under `<primary>/.moai/worktrees` or its branch begins with `WT-`. A tree that fits no class is `unclassified`; it shall be listed with that label, never moved, and never counted as moai-made. The same classifier shall serve `moai worktree list`, `moai worktree migrate` and the `moai update` consolidation (REQ-WHR-024).

**REQ-WHR-014** (Ubiquitous — non-destructive invariant). The migration shall be non-destructive. The only filesystem mutations it may make to a tree are `git worktree move <source> <destination>` and, for rollback, the same command in the opposite direction; it shall never run `git worktree remove`, `git worktree prune`, `git branch -D`, `git reset`, `git clean`, `git checkout`, `git stash`, `rm`, or a recursive delete, and it shall never dispose of a tree. Disposal remains with the existing landing-gated verbs and requires integration plus remote landing of the tree's branch.

**REQ-WHR-015** (Event-driven — dry-run default). When `moai worktree migrate` runs without `--apply`, it shall perform no mutation of any kind and shall print, for every registered tree, its class, source, destination and skip reason; `--apply` shall be required to move trees; `--verify` shall re-read the journal and the registered state and report each recorded move as intact, drifted or rolled back.

**REQ-WHR-016** (State-driven — skip conditions). While a tree is locked, prunable, the current working directory of any running process, anchored by a live session, missing, or blocked by an existing destination, or while its destination lies on a different filesystem than its source, or while the process working directories cannot be inspected, the migration shall skip the tree, leave it at its old path, and print the reason. A skipped tree shall remain fully usable at its old path.

**REQ-WHR-017** (Event-driven — manifest, verification, automatic rollback). When the migration moves a tree, it shall first record a manifest (HEAD, branch, the list of commits not reachable from any remote, a hash of `git status --porcelain=v1 --untracked-files=all`, the count of ignored files, and the file count of the tree's `.moai/reports`) and a journal entry, and after the move it shall recompute the same manifest at the destination. When any field differs, or the card rebind (REQ-WHR-011) fails, the migration shall reverse the move and report the tree as not migrated; when the reversal also fails, it shall leave the tree at the destination, record that fact, and exit non-zero.

**REQ-WHR-018** (Event-driven — explicit rollback). When `moai worktree migrate --rollback <journal>` runs, the system shall move each recorded tree back to its recorded source if that source path is free and the tree's manifest still matches, rebind the card record in the opposite direction, and report each tree restored or refused with the reason.

**REQ-WHR-019** (Event-driven — `moai update` consolidation). When `moai update` runs in the primary checkout and legacy-rooted moai-made trees exist, the `moai update` consolidation shall use the classifier of REQ-WHR-013, shall not select a Claude-native agent-isolation tree, and shall behave as decision D-6 (§E) settles: by default it prints the dry-run summary and the explicit `moai worktree migrate --apply` command and moves nothing.

### Retirement and transition

**REQ-WHR-020** (Event-driven — retirement). When no registered tree and no entry remain under `<primary>/.moai/worktrees`, the system shall remove that directory only if it is empty, using a non-recursive removal; it shall never remove, move or modify `<primary>/.claude/worktrees` or any tree in it that is not moai-made; and no creation path shall name either legacy root after the run phase completes, except as a read candidate for the transition (REQ-WHR-022).

**REQ-WHR-021** (State-driven — pre-run gate and guidance). While the pre-run measurement record (§F, PM-1 and PM-2) is absent from `.moai/reports/t1640/`, the run phase shall not begin. Where the record exists, the guidance surfaces in scope shall state the new location, shall separate a terminal launch (`moai cc -w <absolute-path>`, `codex -C <absolute-path>`) from an in-session move (`EnterWorktree` / `ExitWorktree` for Claude Code, `/cd <absolute-path>` for Codex), shall assert only entry behavior that the record shows, and shall carry zero sentences that tell a session to run `moai cc -w` from inside Claude Code. AGENTS.md shall not grow beyond its measured pre-change size of 23,780 bytes.

**REQ-WHR-022** (Ubiquitous — transition resolver). For as long as legacy-rooted trees exist, every lookup of a tree by name or card id shall find a tree at the home root or at a legacy root, home root first, and a name present in both shall be reported as ambiguous with an instruction to use an absolute path; the transition shall end by a later release removing the legacy candidates, not by elapsed time.

**REQ-WHR-023** (Ubiquitous — spelling-safe path equality). Every path equality or containment test introduced or changed by this SPEC shall treat two spellings of one directory as equal on a case-insensitive volume (compared after symlink resolution and, on darwin and windows, case folding) and shall never fold case on a case-sensitive volume.

**REQ-WHR-024** (Unwanted — scope creep into Claude-owned state). The migration and the resolver shall not create, move, rename, prune or remove any tree under `<primary>/.claude/worktrees` that is classified claude-native, any tree under `~/.codex/worktrees`, or the integration worktree, and shall not register a `WorktreeCreate` or `WorktreeRemove` hook.

## §D. Constraints and invariants

- **I-1 Non-destructive (operator invariant, cadf).** REQ-WHR-014 is the invariant; every other requirement is subordinate to it. A requirement that can only be met by disposing of a tree is not met — the migration skips instead.
- **I-2 Evidence of non-loss.** "Nothing was lost" is proven per tree by the manifest comparison of REQ-WHR-017 and by `moai worktree migrate --verify`, never by the absence of an error message.
- **I-3 Claude and Codex own their trees.** Per dab5, Claude Code owns its `agent-*` trees; per the Codex worktree facts (research.md §R6) Codex owns `~/.codex/worktrees`. MoAI's home root holds only MoAI-made trees. This is also how the t1611 requirement "do not mix with Codex-app-managed trees; split storage and cleanup ownership" is met.
- **I-4 No time-based transitions.** Neither the transition window nor retirement is tied to elapsed time; both end on observed state.
- **I-5 Always-loaded budget.** AGENTS.md is 23,780 bytes against a 24,576-byte Codex budget (796 bytes of headroom); every guidance edit there is net-neutral or net-negative in bytes. Any other always-loaded rule edit states its byte size and non-invoking-session cost per `rule-authoring.md`.
- **I-6 Template-First.** Every change to a deployed surface lands in `internal/template/templates/` and is mirrored to the local tree; the run phase regenerates embedded assets with the repository's own build step.
- **I-7 GEARS/AC budget.** Tier L ceilings are 25 requirements and 25 acceptance criteria; this SPEC carries 24 and 24.

## §E. Operator decisions (recommended defaults; none is applied silently)

Each decision lists the recommended default and the reason. The run phase may start only when these are either confirmed or left at their default by an explicit record. Full alternatives and rejected options are in design.md.

| # | Decision | Recommended default | Reason (one line) | If the operator chooses otherwise |
|---|----------|---------------------|-------------------|-----------------------------------|
| D-1 | Exact path form and key | `<MoaiHome>/worktrees/projects/<project-key>/<name>`; key = sanitized project name, hash-suffixed only on collision, recorded in a marker | Keeps the operator's literal wording (`projects/<project-name>/<name>`) in the common case, and a deterministic suffix removes the silent-merge hazard | Always-suffixed key (`<name>-<8hex>`): REQ-WHR-002/003 simplify, paths lose readability |
| D-2 | Which trees are "moai-made" | Classifier of REQ-WHR-013 (name pattern + `WT-` or in-repo MoAI root; unclassified are listed, never moved) | Satisfies dab5 literally and fails closed on the two real outliers | Broaden/narrow the predicate; 2 trees (`audit-gate-async-worktree-doc`, `main`) are the measured outliers |
| D-3 | Tier by class, not path | Adopt REQ-WHR-009 | A moved session tree must not become `done`-eligible | None viable that keeps `done` safe |
| D-4 | Transition mechanism | Dual-read resolver (REQ-WHR-022), no symlinks at old paths | Symlinks break the tier classifier, re-enter the repository tree and need privilege on Windows | Symlink compatibility: see design.md D-4 for the cost list |
| D-5 | Launcher short name for a new tree | Materialize at the home root, pass the absolute path (REQ-WHR-006) | Only way the completion criterion "no new tree outside the root" can hold for `moai cc -w <new-name>` without a WorktreeCreate hook | Leave Claude-native creation: then `moai cc -w <new-name>` stays an exception and loses nothing of `.worktreeinclude` |
| D-6 | Migration trigger | Explicit `moai worktree migrate --apply`; `moai update` previews and prints the command | Migration now has a non-git side effect (card rebind) and a scale of ~200 trees; the leader should choose a quiet moment | `moai update` applies (the ef12 literal reading); one-line change in the update caller |
| D-7 | Live apply timing | A leader-operated step recorded in sync evidence, after landing; the lane delivers tool, rehearsal and dry-run only | The lane's own tree and other lanes' trees are live; applying from inside a lane risks the `current process is inside this worktree` skips and one-writer violations | Lane applies at landing: needs a quiet-machine precondition |
| D-8 | Config surface | `MOAI_HOME` relocates the root; add only `git_strategy.worktree_project_key`; leave the dead `worktree_root` key unread | The dead key is documented deprecated and may hold stale values that would silently activate if revived | Revive `worktree_root` as an absolute-root override with validation |
| D-9 | Integration worktree (`.claude/worktrees/main`) | Out of scope; stays at its fixed path | Docs and the lane protocol address it by fixed path; relocating it is a separate decision | Include it in a later SPEC |

## §F. Open questions and pre-run measurements

Every item below depends on a fact this plan phase did not measure. Each is stated with its pass/fail meaning in plan.md §B; none is assumed.

| Id | Unmeasured fact | Why it matters | Blocks |
|----|-----------------|----------------|--------|
| PM-1 | `EnterWorktree` into a home-root tree on the installed Claude Code (2.1.303 measured by `claude --version` in this session): from the primary checkout and from inside another worktree session, in default permission mode and in bypassPermissions | The CC changelog line 4231 says entry outside `.claude/worktrees/` asks for confirmation; t1611 measured no prompt from the primary and a refusal from a worktree session; bypass behavior is unmeasured. Lanes run in bypass and hop between cards in-session | Run-phase start (REQ-WHR-021); guidance text |
| PM-2 | Codex `/cd <abs>` into a home-root tree on the installed Codex (0.163.0): trust requirement of the destination, AGENTS.md/hooks reload, and write-root behavior | The memory note (10-10, on 0.161.0) says `/cd` refuses an untrusted destination; whether a home-root tree is trusted is unmeasured; `codex -C` is the fallback | Run-phase start; Codex guidance |
| PM-3 | `claude --worktree <absolute path outside .claude/worktrees>` creation/entry semantics, `.worktreeinclude`, and that a session started in a home-root tree gets hooks, statusline, trust handling and a coherent project-memory key | REQ-WHR-006 passes an absolute path to a backend that has so far been shown (launcher comment only) to accept one | REQ-WHR-006, M2 |
| PM-4 | `git worktree move` across filesystems (source repository and `~/.moai` on different volumes) | Measured same device (16777230) for the repository, `~/.moai/worktrees` and `.git` on this machine; the cross-volume behavior is unmeasured, so REQ-WHR-016 refuses instead of copying | M6 safety |
| PM-5 | Codex workspace-write behavior for a home-root tree: write to `<repo>/.git/worktrees/<id>` and commit | Memory note: commit refused in the sandbox in both layouts (10-10); not re-measured for the new location | Codex lane guidance |
| PM-6 | Whether moving a tree changes the Claude project-memory/transcript key (the encoded cwd directory under the Claude profile `projects/`) so `claude --continue` in the new path sees none of the old history | The encoded directory name visible in this session's own tool-result path is derived from the cwd; the effect on a moved tree is inferred, not measured | Residual-risk disclosure in migration output |

Open questions that are not measurements: OQ-1 release target (the SPEC states `v3.2.0 target`; the card's "3.3 deferral" lift implies now, but the leader may re-target); OQ-2 whether the optional config key of REQ-WHR-004 is wanted at all (YAGNI alternative: drop REQ-WHR-004, keep the deterministic suffix); OQ-3 treatment of the 31 `agent-*` trees that carry unpushed `WT-` branches (dab5 leaves them; the SPEC leaves them and lists them in `moai worktree list` so the reconcile question is visible); OQ-4 whether the Windows case needs a separate measurement (paths use the `paths.Home()` resolver; no Windows measurement taken); OQ-5 whether the docs-site (46 files naming `moai cc -w`, four locales under the same-PR rule) and the five README/CHANGELOG files are updated by this card or by t1879 — recommended default: t1879, because it rewrites that guidance once against the new location, and this card's REQ-WHR-021 covers the agent-facing surfaces only; OQ-6 the design documents named by the card (`.moai design §8.5`, the master upgrade plan HTML §2 stage 2) could not be located in this tree (research.md §R7), so the plan rests on the card text and the three decision records.

## §G. Out of Scope

### Out of Scope — Claude-native agent-isolation trees
- The 31 `.claude/worktrees/agent-*` trees and any future ones. Claude Code owns their lifecycle (dab5). Not moved, not cleaned, not reclassified.

### Out of Scope — WorktreeCreate / WorktreeRemove hooks
- Registering the shipped but unregistered `handle-worktree-create.sh`. A hook replaces native creation for every isolation spawn and loses `.worktreeinclude` and `baseRef` processing (ef12; changelog snapshot lines 7215, 5540). Rejected by dab5.

### Out of Scope — wider guidance rewrite (card t1879) and in-session refusal guard
- The EnterWorktree-versus-`moai cc -w` guidance rewrite beyond the t1610 part, the lane in-session move guidance (card t1753, contained in t1879), and any change to the Claude Code runtime's worktree-isolation guard, which is runtime state and not repository code.

### Out of Scope — name resolution defect GH #1813 (card t1791)
- Lands before this card if its audit passes; this SPEC's launcher change then absorbs its behavior instead of re-implementing it.

### Out of Scope — Codex-managed worktree pool
- `~/.codex/worktrees` (12 trees observed) and the Codex app's Worktree root setting. MoAI lists them in `moai worktree list` and never touches them.

### Out of Scope — integration worktree and its fixed path
- `.claude/worktrees/main` and the `/.claude/worktrees/develop` suffix check at `internal/mission/integration.go:41` (decision D-9).

### Out of Scope — disposal and the first cleaner's deletion predicates
- Card t1501's cleaner extensions that would delete trees (sweep predicates: card state done, own-commit landing, ignored-content allow-list). This SPEC absorbs only the non-deleting part of t1501: classification and listing. The disposal predicates stay as they are.

### Out of Scope — `core.hooksPath=/dev/null` and other git config
- The leader candidate (10-10) is observed already effective in the shared `.git/config` (`git config --show-origin --get-all core.hooksPath` prints `file:/Users/goos/MoAI/moai-adk-go/.git/config /dev/null` in this tree). Git hooks are resolved from the common directory, independent of worktree location, so no location-specific hook work is needed and no code change is made for it.

### Out of Scope — public documentation (docs-site and README locales)
- `docs-site/content/**` (46 files naming `moai cc -w`) and `README*.md` / `CHANGELOG.md`. Recommended default per OQ-5: rewritten once by card t1879 after this card lands. This SPEC's guidance requirement binds the agent-facing surfaces (AGENTS.md, rules, skills, agents, template mirrors, help strings).

### Out of Scope — web console surfacing of the new key
- Adding `worktree_project_key` to the settings web console schema.

### Out of Scope — live migration executed by the lane
- Applying the migration to the real repository's ~200 trees (decision D-7).

## §H. Cross-references

- Card bodies: t1640 (this card), t1611, t1610, t1501 (absorbed); t1879 and t1753 (follow-on, guidance); t1791 (GH #1813, lands first).
- Decision board: `d-20261018T221252Z-cadf`, `d-20261018T221922Z-dab5`, `d-20261018T221839Z-ef12` (`moai decision read --scope card:t1640`).
- `.claude/rules/moai/workflow/worktree-integration.md` § Terminology Glossary (L1/L2) and `worktree-integration-ops.md` § WorktreeCreate and WorktreeRemove Hooks (Not Registered by Default).
- `AGENTS.md` §3 Worktrees; `AGENTS.local.md` §4 (WT- branch and traceability rules).
- Auto-memory reference `reference_codex_in_session_worktree` (Codex `/cd`, managed pool, sandbox facts; measured 2026-10-10 on 0.161.0).
- Evidence and measurement record: research.md; architecture and rejected alternatives: design.md; milestones: plan.md; criteria: acceptance.md.
