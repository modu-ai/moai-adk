# research.md — SPEC-PRIMARY-LOCALMD-RETIRE-001

> Stateless artifact (no `status:` frontmatter). Phase 2/6 research was COMPLETE before
> plan-phase entry — the delegation prompt carried the full evidence set and the dependency
> sweep results; this file records and structures them. Skip rationale is in
> `progress.md` § Phase 1 SKIP Rationale.

## §A Problem evidence

| Finding | Magnitude | Source |
|---------|-----------|--------|
| Upward traversal loads primary `CLAUDE.local.md` into every nested worktree session | **+20,855 tokens/session** (probe 3 marker file: 41,354 first-turn input tokens vs probe 5 real 52,280 B copy: 62,209) | `.moai/reports/t1279/verdict.md` §① (carried citation) |
| Primary-launch always-loaded set double-loads retired doctrine | always-loaded set measured 205,977 chars in t1303; **−15k residual** attributed to this file | card t1317 body quoting `.moai/reports/t1303/verdict.md` follow-up ① draft (file not resolvable from this worktree at 68e37864a — carried quote) |
| Confusion hazard | primary's committed `main` copy is the deprecated "폐기된 제3의 모델" (its own §0.3); the permanent ` M` marker exists by design (§0.4) | `CLAUDE.local.md` §0.3/§0.4 (working copy content, carried) |

## §B Verified preconditions (this run, tree 68e37864a)

| # | Claim | Command | Observed output |
|---|-------|---------|-----------------|
| 1 | t1279 D-scope ordering gate satisfied | `grep '^status:' .moai/specs/SPEC-SESSION-DOUBLELOAD-001/spec.md` | `status: completed` |
| 2 | t1290 merged (reception migration is an ancestor) | `git merge-base --is-ancestor c13cee6d5 HEAD` | exit 0 → true |
| 3 | Worktree geometry = end-state | `ls AGENTS.local.md CLAUDE.local.md` | `AGENTS.local.md` present (35,897 B); `CLAUDE.local.md: No such file or directory` |
| 4 | `.moai/state/` is the right preserve target | `git check-ignore -q .moai/state/retired/x` | exit 0 → ignored (gitignored machine-local scratch) |
| 5 | §0.4 revision target located | `grep -n "별도 전환" AGENTS.local.md` | line 37: `… primary의 구형 파일은 별도 전환 전까지 보존한다.` |
| 6 | Preserve target absent today | `ls -d /Users/goos/MoAI/moai-adk-go/.moai/state/retired` | absent |
| 7 | Primary pre-state (` M` marker, tracked, branch `main`, HEAD) | (intended `git -C <primary> …`) | **NOT re-measured** — the worktree-session guard refused cross-tree `git -C`; carried delegation-verified (this session, 68e37864a): ` M CLAUDE.local.md`, tracked, branch `main`. Re-observe at M1 (AC-PLR-007 discipline). |

Reception-gate closure detail (carried, `SPEC-LOCAL-INSTR-RECEPTION-001/progress.md`
closing entry, 2026-09-28, audit-read 646ae8302): AC-LIR-001~009 all pass, including
AC-IFU-007 `git show HEAD:AGENTS.local.md | wc -m` → 37,061 (≤ 39,999); AC-LIR-007
`git ls-files` → `AGENTS.local.md` as the only tracked local instruction file and
`git show HEAD:CLAUDE.local.md` → exit 128 on develop. Observation recorded: that SPEC's
frontmatter still reads `status: draft` — bookkeeping gap, out of scope (spec.md §D).

## §C Dependency sweep (carried — results quoted, not re-run)

`grep -rn "CLAUDE.local" internal/ cmd/ scripts/ .moai/config/` hits classify as:

- **(a) Name-list only** — `internal/contract/frozen.go`
  `FrozenInstructionFiles = []string{"CLAUDE.md", "CLAUDE.local.md", "AGENTS.md", "AGENTS.local.md"}`
  plus its self-contained synthetic-repo tests (`derived_test.go`,
  `frozen_compat_test.go`). A NAME list: retiring the primary's copy does not affect any
  runtime check. No change needed.
- **(b) Already models the rename** — `internal/contract/kickoff/activation_test.go`
  (line 257 renames the file in a test repo; linkage markers checked against both paths).
  No change needed.
- **(c) Provenance comment anchors** — `internal/config/{team_mode,envkeys,types}.go`
  comments citing "CLAUDE.local.md §14/§25". After retirement the citations remain
  resolvable via git history; re-pointing is the sibling sweep (out of scope).
- **(d) Zero references** — `internal/hook/instructions_loaded.go` has ZERO CLAUDE.local
  references (checked in the sweep).

**Conclusion: no runtime code requires the file's PRESENCE.** The retirement is safe at the
code level; the only tracked-content edit in this card is the §0.4 sentence (REQ-PLR-011).

`.moai/docs` side: 10 files cite CLAUDE.local.md §-anchors (carried sweep count). Residual
risk: readers following those anchors post-retirement must use git history or the
`AGENTS.local.md` successor numbering. Follow-up card candidate.

## §D The geometry insight (design-exploited fact)

A develop checkout IS the end-state: this worktree proves it (§B row 3). The primary-side
retirement is therefore not a migration — it is the removal of the last legacy copy, with
the successor already landed and reception-proven. The only reason the primary still loads
the old file is that `main` tracks it and the working copy sits there; ONE deletion commit
removes both the load and the deprecated committed copy from reach.

## §E Alternatives considered (rejected)

- **Delete outright, no preservation** — violates the card body ("삭제 아님") and §0.4's
  보존 clause intent; rejected.
- **Commit the working copy's current content to `main` first, then a second commit
  deletes it** — two commits where ONE suffices, and it promotes the deprecated model's
  history forward; the git history already preserves every prior version. Rejected.
- **Preserve under `.moai/reports/` or a template mirror** — `.moai/state/` is the
  designated gitignored machine-local scratch; reports are for evidence artifacts; a
  template mirror would distribute a retired file. Rejected.
- **Leave primary untouched, only update doctrine** — keeps the +20,855 tokens/session
  load; defeats the card's purpose. Rejected.
