---
id: SPEC-AGENT-MEMORY-DRAIN-001
title: "Worktree agent-memory drain: write-time mirror to the primary store plus one-shot backfill, so no agent memory dies with its tree"
version: "0.2.1"
status: draft
created: 2026-09-02
updated: 2026-09-02
author: manager-spec
priority: P1
phase: "v3.1.4"
module: "internal/cli"
lifecycle: spec-anchored
tags: "agent-memory,worktree,hook,mirror,drain,reaper,post-tool"
tier: M
---

## §A — History

- **2026-09-02** — v0.2.1 non-transition correction after the Implementation Kickoff
  approval (operator, relayed via the lead): design (c) adopted, AUTONOMOUS progression.
  `phase` corrected `"v3.1.5 target"` → `"v3.1.4"` — the operator decided this card joins
  the v3.1.4 close (PR #1685), retiring the v0.2.0 assumption that v3.1.4 would already
  have closed. `status` untouched (stays `draft`; non-transition). Stale-target mentions
  aligned in plan.md §E and progress.md §E.1. Requirements and criteria unchanged
  (10 / 9).
- **2026-09-02** — plan-phase v0.2.0 after plan-audit iteration 1 returned **PASS-WITH-DEBT 0.94** (Tier M threshold met; all load-bearing measurements reproduced by the auditor). Fixes folded in: **D1** — REQ-AM-001's purpose clause qualified to the design's actual coverage (mirror-mediated Write/Edit writes + drain backfill), naming the Bash-write residue as an accepted, doctor-detectable residual instead of asserting "no loss through any path". **D2** — the MEMORY.md index count corrected (30 at the audit snapshot, not 29; the v0.1.0 four-agent enumeration summed to 28 because the classification listing was truncated at `head -20`, dropping `manager-lead` ×1, `manager-docs` ×1, and undercounting `plan-auditor`); propagations fixed in §B.1/§B.4 and plan.md §A.1. A drift note added: a re-count minutes later reads 31 indices + 42 topics = 73 files — the population is live and drifting, so the figure is a snapshot, not a census. **D3** — RED-cell carrier forms fixed per verification-completeness §2.1 (AC-AM-002 split to single-invocation form; AC-AM-005 exit code recorded; AC-AM-006's design-only disposition replaced with the absent-mechanism grep probe). **D5** — the t209 death-path claim qualified as inference (the actual disposal path is unobservable post hoc). **D7** — M2's mirror trigger anchored to the `.claude/agent-memory/` path segment, not the audit's unanchored `agent-memory/` substring predicate. **D8** — M3's docs-decision branch now names the README 4-locale parity obligation. Requirements and criteria counts unchanged (10 / 9).
- **2026-09-02** — plan-phase v0.1.0 authored from card t223 ("워크트리 안에서 만든 agent-memory 가 primary 에 도달하지 못하고 트리와 함께 죽는다", Class C design decision, Tier S~M judged **M** — see §C.3). Predecessor: SPEC-WORKTREE-REAPER-001 REQ-WR-025 records that its P2 keep-classification is a stopgap and drain-then-dispose is the correct fix; this SPEC is that fix. All scale figures re-measured on this tree (`c0c36c421`) — the card's 5-of-156 figure from 2026-08-24 is stale by an order of magnitude.

## §B — Problem

### B.1 — The defect, re-measured

`.claude/agent-memory/` is gitignored (`.gitignore:194`, template `.gitignore:179`) and
per-project — and a worktree is its own project root. Memory a subagent writes inside a
worktree therefore never reaches the primary checkout through git, and dies with the tree.

Measured on 2026-09-02 from the t223 worktree (tree `c0c36c421`):

- `git worktree list | wc -l` → **186** worktrees.
- Worktrees carrying a `.claude/agent-memory/` directory → **88** (find under
  `.claude/worktrees`, maxdepth 3).
- File-bearing trees → **26**, holding **70 files** at the plan-phase snapshot
  (2026-09-02): **30 per-agent `MEMORY.md` indices + 40 topic files**. Index breakdown by
  agent: `manager-develop` ×12, `manager-spec` ×9, `plan-auditor` ×6,
  `sync-auditor` ×1, `manager-lead` ×1, `manager-docs` ×1 (= 30). The 40 topic files are
  mostly `feedback_*`/`project_*` lessons. The other 62 trees hold skeleton-only empty
  agent directories (auto-created at session start). **Drift note**: the population is
  live — a re-count in the same session minutes later reads 31 indices + 42 topics = 73
  files (one new `plan-auditor/MEMORY.md` among them). The figure is a snapshot, not a
  census; the defect's scale is order-of-magnitude.
- Overlap with primary: of the 40 worktree topic files, **0** exist in the primary store
  under the same `agent/filename` (`comm -12 /tmp/t223-wt-topics.txt
  /tmp/t223-primary-topics.txt` → empty output, exit 0, against the primary's 203 topic
  files).
- The card's concrete instance is already lost: `t209` no longer appears in
  `git worktree list` (0 matches), and its orphaned lesson
  `feedback_go_test_run_nonexistent_passes.md` is absent from the primary store
  (`find … -name '*nonexistent_passes*'` → no hit; the one `*falsifiab*` hit is an
  unrelated sync-auditor file). P2's preserve-forever did not hold in practice — the
  tree is gone while no moai-owned sweep removed it. **The disposal path itself is
  inference, not observed fact**: which path killed the tree is unobservable post hoc;
  the observed facts are the tree's absence and that the reaper's guards never fired on
  it, which is consistent with a manual `git worktree remove` or session-end-prompt
  disposal — exactly the class a dispose-time-only drain cannot hook (§B.3).

### B.2 — Surface scoping: which memory store the defect reaches

Two distinct surfaces (`.claude/rules/moai/workflow/moai-memory.md` § Two Memory Locations):

1. **Repo-local `.claude/agent-memory/<agent>/`** — per-project by construction; the
   worktree is its own project root, so this surface **is** fragmented per-tree and **is**
   the defect's reach. Verified: 26 file-bearing worktree copies, 0 mirrored.
2. **Native auto-memory `~/.claude/projects/<derivation>/memory/`** — derives `<project>`
   from the git repository root, so worktrees share the primary's store. Verified on this
   machine: under the active profile's `projects/` directory there are 181 per-worktree
   *session-transcript* dirs but **0** per-worktree `memory/` dirs
   (`find … -maxdepth 2 -type d -name memory | grep -c worktree` → 0). The native store
   is **not** affected by this defect and is out of scope.

### B.3 — Design directions investigated

- **(a) Drain at dispose time** (hook a drain into every moai-owned removal point —
  `prMergeCleanup` `internal/cli/session_worktree_prmerge.go`, `worktree clean --merged-only`
  and `--stale` `internal/cli/worktree/clean.go`, `worktree done` and `remove`
  `internal/cli/worktree/{done,remove}.go`). **Rejected as the primary mechanism**: the
  enumerated moai-owned paths are 5, but manual `git worktree remove` and the session-end
  keep/remove prompt are outside moai code entirely — and the t209 tree is gone without
  any moai-owned sweep having removed it, consistent with death through one of those
  unhookable paths (§B.1; inference, not observed fact). A drain that covers only the
  reaper's paths is a partial fix measured against that loss.
- **(b) Wire worktree sessions to write agent-memory to the primary path directly.**
  **Rejected as infeasible short-term**: the repo-local agent-memory path is injected by the
  Claude Code runtime from the session's project directory (worktree root); no override
  surface exists for it (the `autoMemoryDirectory` setting overrides only the *native*
   store, and no env/setting redirects the repo-local agent-memory path). This direction
  requires an upstream Claude Code change.
- **(c) Write-time mirror through the existing PostToolUse memory hook — RECOMMENDED.**
  `internal/hook/post_tool.go` `runMemoryAudit` already intercepts every Write/Edit
  targeting a `.md` under `agent-memory/`, in worktree sessions too (settings.json is
  tracked, so the hook wiring is present in every tree). Extending that interception to
  copy the just-written file into the primary store makes memory reach primary
  **continuously, at write time** — before any disposal path, moai-owned or manual, can
  destroy it. Plus a **one-shot backfill** (`moai memory drain`) to harvest the existing
  26 trees / 70 files so the backlog does not wait for re-writes.

### B.4 — The index problem, solved together

The `MEMORY.md` index is per-agent per-tree (30 of the 70 snapshot files are indices). A
naive drain that copies everything would clobber the primary's index; a drain that copies
only topic files strands them (a topic file with no index line is stored and never
recalled — the exact failure `moai memory doctor` exists to report). And `[[name]]` links
are name-based, not path-based, so they survive the move — the t209 orphaned link
(`[[card-premise-needs-investigation]]`, referencing a memory living outside the tree)
actually *resolves better* once its file lives in the primary store. The reconciliation
rules this SPEC adopts: copy topic files (never the index file), never overwrite a primary
file (tree-qualified suffix on collision), and append a missing index line per copied
topic (see REQ-AM-003/004).

## §C — Scope

### C.1 — In scope

- A write-time mirror of agent-memory `.md` writes from worktree sessions into the primary
  checkout's `.claude/agent-memory/` store, implemented in the existing PostToolUse hook
  path, fail-open.
- A `moai memory drain` subcommand: enumerate registered worktrees, report agent-memory
  content, preview by default, copy file-bearing content into the primary store under the
  same reconciliation rules (backfill).
- Reconciliation rules: collision policy, index-line append, `_archive/` handling.
- Primary-checkout resolution from a worktree (`git rev-parse --git-common-dir` parent).
- Unit tests with fixtures for all of the above.

### C.2 — Constraints

- **Fail-open is non-negotiable** (inherited from the hook's contract,
  `internal/hook/post_tool.go`: observation-only, exit 0 always): a mirror failure must
  never block or error the originating Write/Edit.
- **Drain never deletes** — copies only. Worktree disposal remains the reaper's and the
  operator's act; SPEC-WORKTREE-REAPER-001's P2 guard (REQ-WR-024) is untouched and stays
  as the safety net for writes the mirror cannot see.
- **Known blind spot, accepted**: the hook sees Write/Edit tool calls only. Memory written
  via Bash (`cp`, heredoc) bypasses the mirror — the same blind class the existing memory
  audit has. The backfill and an operator-run `moai memory drain` cover the residue.
- **Concurrency**: N worktree sessions mirroring simultaneously write distinct topic files
  almost always (measured frequency: 70 files across ~5 months); the index append is
  read-modify-write and may race — tolerated, because `moai memory doctor` (existing)
  detects orphans/dangling index lines and the index is rebuildable from topic files. The
  never-overwrite rule (REQ-AM-003) makes file-level races non-destructive.
- **Shared-checkout discipline**: mirrored writes land only in the primary's *gitignored*
  agent-memory store — no branch state, no tracked-file mutation, no sweep staging.
- **Template-First**: this design adds **no template files** (Go code + CLI surface only).
  If run-phase documentation work later edits `.claude/rules/moai/workflow/moai-memory.md`,
  that edit must be mirrored to `internal/template/templates/.claude/rules/moai/workflow/moai-memory.md`
  (CLAUDE.local.md §2 HARD) — flagged here, executed only if needed.

### C.3 — Tier judgment

**Tier M.** Rationale: a single cohesive feature, but it crosses two packages
(`internal/cli` — new `memory drain` subverb; `internal/hook` — mirror extension) plus a
shared reconciliation helper, needs fixture-based tests in both, and carries a real design
decision (mirror-vs-dispose-drain) the operator confirms at kickoff. Not Tier S (≥3
milestones, multi-package surface, non-trivial collision/index semantics); not Tier L (no
independent research.md needed — the research embedded here is already conclusive; no
≥10-file surface; single-actor implementation).

## §D — Requirements (GEARS)

### D.1 — Write-time mirror

- **REQ-AM-001** (Ubiquitous) — The drain mechanism shall copy every worktree-local
  `.claude/agent-memory/<agent>/` topic file into the primary checkout's
  `.claude/agent-memory/<agent>/` store, such that memory written through the mirrored
  surface (Write/Edit tool calls, REQ-AM-002) or harvested by the backfill (REQ-AM-007)
  is no longer lost to worktree disposal through any path, moai-owned or manual. Memory
  written by other means (e.g. a Bash `cp`/heredoc the PostToolUse mirror cannot see) is
  an accepted residual — undrained until an operator-run `moai memory drain` harvests it,
  and detectable by `moai memory doctor` (spec.md §C.2, §G).
- **REQ-AM-002** (Event-driven) — **When** a Write or Edit tool call targets a `.md` file
  under `.claude/agent-memory/` inside a worktree, the memory mirror shall copy the written
  file to the same agent-relative path under the primary checkout's agent-memory store.
- **REQ-AM-003** (Unwanted) — The mirror and the drain shall not overwrite an existing
  topic file in the primary store; a name collision shall be resolved by writing a
  tree-qualified copy (suffix `.wt-<worktree-name>` before `.md`).
- **REQ-AM-004** (Event-driven) — **When** a mirrored or drained topic file has no index
  line in the primary agent's `MEMORY.md`, the drain mechanism shall append exactly one
  index line for it, derived from the worktree's index line or the file's frontmatter
  description.
- **REQ-AM-005** (State-driven) — **While** the primary checkout cannot be resolved or a
  copy fails, the mirror shall fail open: the originating Write/Edit proceeds unblocked,
  the hook exits 0, and the failure is emitted as a non-blocking notice.
- **REQ-AM-006** (State-driven) — **While** the session's project root is the primary
  checkout itself (not a worktree), the mirror shall be a no-op.

### D.2 — Backfill command

- **REQ-AM-007** (Ubiquitous) — `moai memory drain` shall enumerate every registered
  worktree, report per-tree agent-memory content (agent names, file counts, index
  presence), and with the apply flag copy file-bearing content into the primary store
  under the REQ-AM-001..004 rules.
- **REQ-AM-008** (Where — capability gate) — **Where** `moai memory drain` runs without
  the apply flag, the command shall preview the copy set and write nothing.
- **REQ-AM-009** (Unwanted) — `moai memory drain` shall not delete or modify any file
  inside a worktree and shall not remove any worktree.
- **REQ-AM-010** (Ubiquitous) — `moai memory drain --json` shall emit machine-readable
  per-tree records carrying at minimum path, agent, file counts, copied, collided, and
  skipped counts.

## §E — Acceptance-criteria pointer

All criteria, their two-cell adoption records (RED-now measured on this tree
`c0c36c421`, 2026-09-02, plus green-path milestone mapping), and severity/traceability
matrices live in `acceptance.md` (§D). Run-phase verification follows
`.claude/rules/moai/development/verification-completeness.md` §2.

## §F — Dependencies

- **SPEC-WORKTREE-REAPER-001** (completed) — REQ-WR-024 (P2 keep-classification) stays in
  force as the safety net; REQ-WR-025 names this SPEC's mechanism as the correct fix. This
  SPEC does not modify the reaper's guard.
- **moai-memory.md** taxonomy — mirrored copies must preserve frontmatter (4-type schema);
  the existing PostToolUse taxonomy audit applies to the source write as today.

## §G — Out of Scope

### Out of Scope — relaxing the reaper's P2 guard

- SPEC-WORKTREE-REAPER-001 REQ-WR-024 keeps `.claude/agent-memory/` in the irreplaceable
  ignored-content class. Once the mirror has been deployed and the backfill has run, that
  category empties on its own and a follow-up card may relax P2 — this SPEC changes
  nothing about the guard, so an undrained tree remains protected in the meantime.

### Out of Scope — the native auto-memory store

- Verified unaffected (§B.2): worktree sessions already share the primary's native store
  (`~/.claude/projects/<derivation>/memory/`, derivation from the git repository root;
  0 per-worktree memory dirs observed). No change to it, and no `autoMemoryDirectory`
  work.

### Out of Scope — upstream Claude Code path override (direction b)

- Redirecting the repo-local agent-memory write path itself requires an upstream Claude
  Code override surface that does not exist today. Not attempted; the mirror (direction c)
  achieves the effect without it.

### Out of Scope — SessionEnd-hook drain and Bash-write coverage

- A drain wired to SessionEnd (to catch memory written via Bash, which the PostToolUse
  mirror cannot see) was considered and deferred: SessionEnd does not fire on crash/kill,
  so it is a weaker guarantee than the write-time mirror, and the residue is covered by
  operator-run `moai memory drain`. Revisit only if measured residue proves significant.

### Out of Scope — cross-worktree memory deduplication/merging

- Two trees writing the same lesson produce a tree-qualified duplicate copy by design
  (REQ-AM-003). Semantic deduplication (recognizing two files as the same lesson and
  merging) is a judgement, not a mechanism — `moai memory archive` (existing) retires
  duplicates on operator naming, which is the right surface for it.
