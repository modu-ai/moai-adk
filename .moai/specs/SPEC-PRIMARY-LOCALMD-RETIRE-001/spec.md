---
id: SPEC-PRIMARY-LOCALMD-RETIRE-001
title: "Retire the primary checkout's legacy CLAUDE.local.md — preserve under .moai/state, land as one policy commit"
version: "0.1.0"
status: draft
created: 2026-09-29
updated: 2026-09-29
author: manager-spec
priority: P1
phase: "v3.3.0 target"
module: "CLAUDE.local.md (primary checkout), AGENTS.local.md, .moai/state/retired"
lifecycle: spec-anchored
tags: "local-instructions, primary-checkout, retirement, token-budget, repo-local, one-policy-commit"
tier: M
related_specs: [SPEC-LOCAL-INSTR-RECEPTION-001, SPEC-SESSION-DOUBLELOAD-001, SPEC-INSTRUCTION-BUDGET-SCOPE-001]
---

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-09-29 | **Created by card t1317** (leader-issued 2026-09-29; t1303 [C]-axis follow-up ①; Tier M; class C). Retires the primary checkout's legacy `CLAUDE.local.md` — the perpetually-modified working copy on branch `main` — preserving it under `.moai/state/retired/` (NOT deleting outright) and landing the retirement as ONE policy commit. Card-inherent design choices take the recommended option and record reasoning in plan.md §A. Ordering gates: t1279 D-scope landed (verified: `status: completed` at 68e37864a) and t1290 reception gate passed (verified: `git merge-base --is-ancestor c13cee6d5 HEAD` → true) — both re-affirmed at run entry per REQ-PLR-001/002. |

---

## §A Context

The primary checkout of this repository sits on branch `main`, and `CLAUDE.local.md` is tracked
there. Its working copy holds the current shared maintainer doctrine while `main`'s committed
copy is a deprecated pre-gitflow model — the file's own §0.3 names that committed copy
"폐기된 제3의 모델" (the discarded third model), and its §0.4 marks the resulting permanent
` M CLAUDE.local.md` status line as intentional (never `git restore`-ed).

This geometry now costs tokens on three measured surfaces:

1. **Nested worktree sessions** — upward directory traversal loads the primary
   `CLAUDE.local.md` into every nested worktree session. Measured in card t1279 §①
   (probe 3 marker-file session: 41,354 first-turn input tokens vs probe 5 real-copy session:
   62,209 tokens = **+20,855 tokens/session**, `.moai/reports/t1279/verdict.md`).
2. **Primary launches** — the always-loaded set double-loads retired doctrine;
   t1303 reduced the always-loaded instruction set to 205,977 chars, and its follow-up ①
   attributes a further −15k residual to this file (`.moai/reports/t1303/verdict.md`
   follow-up ① draft, quoted in the card body).
3. **Confusion surface** — §0's canonical-copy discriminator exists precisely because the
   primary copy keeps threatening to be read as canonical. Retiring it closes the hazard
   instead of annotating it forever.

The successor content already lives on `develop` as `AGENTS.local.md` (card t1290,
`SPEC-LOCAL-INSTR-RECEPTION-001`: reception gate 5/5 measured, migration landed at
`c13cee6d5`, `AC-LIR-007` — develop tracks `AGENTS.local.md` exactly and does not track
`CLAUDE.local.md`). A develop checkout IS the end-state geometry: in this worktree
(based on develop tip `68e37864a`), `AGENTS.local.md` is present (35,897 B) and
`CLAUDE.local.md` is absent, with a clean status for that path — measured this session.

`AGENTS.local.md` §0.4 explicitly reserves this act: "primary의 구형 파일은 별도 전환 전까지
보존한다" (the primary's old file is preserved until the separate conversion). This SPEC IS
that separate conversion (별도 전환). After it, that sentence becomes obsolete and is revised
minimally on the develop side.

**The one act, two trees.** The retirement has exactly two write surfaces: (a) the PRIMARY
checkout's local `main` — one commit deleting `CLAUDE.local.md`, with the working copy first
preserved byte-identically under `.moai/state/retired/`; (b) THIS worktree's develop-side
branch — the minimal `AGENTS.local.md` §0.4 revision. Everything else is read-only.

---

## §B Goal

Remove the primary checkout's legacy `CLAUDE.local.md` load — the +20,855 tokens/session
nested-worktree burden and the primary-launch residual — by preserving its content under
`.moai/state/retired/` and landing the retirement as exactly ONE policy commit on local
`main`, with the develop-side `AGENTS.local.md` §0.4 doctrine updated to record the
conversion as complete.

---

## §C Requirements (GEARS)

> **Id convention.** `REQ-PLR-*` are this SPEC's own requirements
> (Primary-Local-md-Retire). All are new; none are moved.

### C.1 Ordering gates (re-affirmed at run entry)

- **REQ-PLR-001** — **When** the retirement is scheduled for execution, the executing agent
  shall re-affirm at the then-current develop tip that `SPEC-SESSION-DOUBLELOAD-001`'s
  `spec.md` frontmatter reads `status: completed` before any primary-side act; **while** that
  field reads any other value, the retirement shall not be executed. Baseline evidence:
  `grep '^status:' .moai/specs/SPEC-SESSION-DOUBLELOAD-001/spec.md` → `status: completed`
  (observed this session at 68e37864a). This executes the card's [HARD] ordering
  "t1279 D-scope 착지 후 은퇴".

- **REQ-PLR-002** — **When** the retirement is scheduled for execution, the executing agent
  shall re-affirm the develop-side reception end-state — `git ls-files` on the develop tree
  lists `AGENTS.local.md` and does not list `CLAUDE.local.md` — before any primary-side act;
  **while** that check fails, the retirement shall not be executed. Baseline evidence:
  this worktree at 68e37864a carries `AGENTS.local.md` (35,897 B) and no `CLAUDE.local.md`;
  `git merge-base --is-ancestor c13cee6d5 HEAD` → true (t1290 migration an ancestor of this
  branch point). Note: the primary checkout (`main`) does NOT carry `AGENTS.local.md`; the
  gate check runs against the develop tree only.

### C.2 Preservation (before any deletion)

- **REQ-PLR-003** — **When** the executing agent is about to delete the primary working
  copy, it shall first write a byte-identical copy of it to
  `.moai/state/retired/CLAUDE.local.md` in the primary
  checkout, and shall record in this SPEC's `progress.md` the copy's sha256, the copy date,
  the card id (t1317), and a source description ("primary checkout working copy, branch
  `main`, retired per card t1317"). `.moai/state/` is gitignored machine-local scratch
  (verified: `git check-ignore .moai/state/retired/x` → ignored at 68e37864a); nothing
  canonical is lost — the doctrine's canonical home is develop's `AGENTS.local.md`, and
  `main`'s deprecated committed copy remains recoverable from git history.

- **REQ-PLR-004** — The preserved copy shall be byte-identical with the pre-deletion working
  copy, proven by `cmp` (or equal `shasum -a 256`) between the preserved file and a capture
  taken before deletion; **when** the comparison fails or the capture is missing, the
  deletion shall not proceed.

### C.3 The one policy commit (primary side)

- **REQ-PLR-005** — The primary-side retirement shall land as exactly ONE commit on the
  primary checkout's already-checked-out branch (`main`), deleting `CLAUDE.local.md` and
  nothing else: staged by explicit pathspec (`git rm CLAUDE.local.md` / `git add` of that
  path only), with `git status --short` re-read immediately before staging to exclude any
  foreign session's files, and a commit message naming card t1317 and this SPEC id.
  Committing to the already-checked-out branch is permitted by the standing contract
  (`AGENTS.md` §2); the forbidden list there covers branch switches, resets, and stashes —
  not commits. Consequence, documented: local `main` diverges from `origin/main` by at
  least this one commit until the lead batch-pushes; exact counts are measured and reported
  at completion (AC-PLR-008), currently expected `0 1` from the audited `0 0` pre-state
  (plan-audit iter1) — the `0 N` divergence-matrix row reads "proceed normally".

- **REQ-PLR-006** — **When** the executing session performs the primary-side act, it shall
  do so through the ExitWorktree → primary act → EnterWorktree round-trip, and shall
  immediately before the commit re-read `git rev-parse --short HEAD` and
  `git branch --show-current` (the staleness rule, `AGENTS.md` §2); **when** the branch
  reads anything other than `main` or HEAD has moved from the value assumed, it shall stop
  and report the divergence instead of proceeding.

- **REQ-PLR-007** — The executing agent shall not run `git restore CLAUDE.local.md` (or any
  working-copy-discarding command — checkout/reset/stash per `AGENTS.md` §2) against the
  primary copy, at any point of the retirement: the content transition happens through
  preservation (REQ-PLR-003) plus the deletion commit (REQ-PLR-005), never through restoring
  the deprecated committed model. This perpetuates the file's own §0.4 [HARD] and the
  2026-09-07 incident it was written from (a `git restore` would silently revert the working
  copy to the §0.3 discarded model).

- **REQ-PLR-008** — The executing lane shall not push: `origin/main` (and `origin/develop`)
  updates remain the lead's batched responsibility (git-flow lane protocol, 2026-09-02);
  **when** the lane's report is issued, it shall state the local commit SHA and the
  unpushed-commit count instead of a push result.

### C.4 Post-state (verification)

- **REQ-PLR-009** — **When** the policy commit has landed, the primary checkout shall
  satisfy all of:
  (a) `git -C <primary> status --porcelain -- CLAUDE.local.md` → empty; (b) the file absent
  from the primary working tree (`test ! -e <primary>/CLAUDE.local.md`); (c) the permanent
  ` M CLAUDE.local.md` status line gone. Baseline (pre-state, delegation-measured this
  session at 68e37864a): the primary shows ` M CLAUDE.local.md` and the file is tracked.

- **REQ-PLR-010** — The upward-traversal load shall be eliminated: the mechanical absence
  check (REQ-PLR-009(b)) is the required evidence, and a session probe — one nested session
  started inside this worktree whose loaded-instructions list no longer contains the primary
  `CLAUDE.local.md`, mirroring t1279 probe 3's methodology — is the optional confirming
  evidence. The SPEC accepts mechanical absence + documented probe-skip when a session spawn
  is impractical in the run phase; the choice and its reason shall be recorded in
  `progress.md`.

### C.5 Develop-side doctrine refresh (this worktree)

- **REQ-PLR-011** — The executing agent shall revise `AGENTS.local.md` §0.4 minimally on
  this worktree's branch: the sentence "primary의 구형 파일은 별도 전환 전까지 보존한다"
  (line 37, measured this session) becomes obsolete after the conversion and shall be
  revised to record that the 별도 전환 completed (card t1317, 2026-09-29), while §0.1–§0.3's
  canonical-copy discriminators stay intact and the edit matches the file's existing Korean
  register. This is the only content edit to tracked files in this card.

- **REQ-PLR-012** — **When** any step of the retirement cannot be completed as specified
  (ordering gate fails, preservation mismatch, staleness-rule divergence, unexpected
  pre-state), the executing agent shall stop and return a structured blocker report to the
  orchestrator instead of improvising.

---

## §D Out of Scope

This SPEC deliberately does not build the following.

### Out of Scope — citation re-pointing (sibling sweep)

- Re-pointing the 10 `.moai/docs/*.md` files that cite `CLAUDE.local.md` §-anchors, and the
  `internal/config/{team_mode,envkeys,types}.go` comment anchors citing "§14/§25".
  Dependency sweep (carried from this session): no runtime code requires the file's PRESENCE
  (`internal/contract/frozen.go` `FrozenInstructionFiles` is a NAME list with self-contained
  synthetic-repo tests; `internal/hook/instructions_loaded.go` has zero CLAUDE.local
  references). After retirement the citations remain resolvable via git history plus the
  `AGENTS.local.md` successor numbering. Follow-up card candidate; residual-risk recorded in
  plan.md.

### Out of Scope — remote deletion and push

- Deleting anything from `origin/main` — push is lead-batched (REQ-PLR-008). `main`'s
  deprecated committed copy becomes unreachable only after the lead's batch push; until
  then the retirement exists on local `main` alone (documented `0 N` divergence).

### Out of Scope — sibling SPEC bookkeeping

- Repairing `SPEC-LOCAL-INSTR-RECEPTION-001`'s frontmatter (still `status: draft` although
  its `progress.md` closes with AC-LIR-001~009 all passing, audit-read 2026-09-28) — a
  status-transition bookkeeping gap recorded as observation, not this card's scope.

### Out of Scope — hook metric work and content migration

- Card t1318's hook-metric work, and any further `CLAUDE.local.md` → `AGENTS.local.md`
  content migration — the migration itself is complete (t1290); this SPEC only retires the
  primary residue and updates the §0.4 doctrine sentence.

---

## §E Cross-references

- `SPEC-SESSION-DOUBLELOAD-001` — the +20,855 tokens/session measurement this retirement
  eliminates; `status: completed` is REQ-PLR-001's ordering gate.
- `SPEC-LOCAL-INSTR-RECEPTION-001` — the migration this SPEC retires the residue of;
  `AC-LIR-007` (single tracked local instruction file on develop) is REQ-PLR-002's gate.
- `SPEC-INSTRUCTION-BUDGET-SCOPE-001` (t1303) — the [C]-axis instruction-budget follow-up
  this card belongs to; verdict `.moai/reports/t1303/verdict.md` follow-up ① draft is the
  procedure SSOT (quoted in the card body; file not resolvable from this worktree at
  68e37864a — carried-quote citation).
- `AGENTS.local.md` §0 (lines 9-39, this tree) — the canonical-copy discriminator and the
  §0.4 보존 clause this SPEC executes and then minimally revises (REQ-PLR-011).
- `.moai/reports/t1279/verdict.md` §① — the probe-3/probe-5 methodology REQ-PLR-010's
  optional probe mirrors.
- `AGENTS.md` §2 — the shared-checkout contract that permits commits to the
  already-checked-out branch and forbids restore/reset/stash (REQ-PLR-005/006/007).
- `.claude/rules/local/gitflow-lane-protocol.md` §4/§6 — lanes do not push; worktree
  disposal waits for remote landing (REQ-PLR-008).
- `.claude/rules/local/gitflow-lane-protocol.md` §8 — merge-base-based scope measurement
  for the WT-branch side (used by acceptance.md AC-PLR-006).
