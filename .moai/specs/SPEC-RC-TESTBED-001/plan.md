# plan.md — SPEC-RC-TESTBED-001

Tier M (3 artifacts: spec.md + plan.md + acceptance.md; progress.md emitted, not counted).
Card: t281. Pre-work tree: `fa8ff89ba` (branch `WT-rc-testbed`, based on `origin/develop`).

## §A Context

**Goal.** Codify the local develop RC testbed: author the two missing rules (local rc.N
numbering policy; develop refresh procedure), wire pointer-only references from
`CLAUDE.local.md` §4.1, and verify the C3 doc-contradiction is resolved. Doc-only — zero code.

**Doc targets (run-phase write surfaces).**

| Target | Ownership | moai-update fate | Milestone |
|---|---|---|---|
| `.moai/docs/version-management.md` | tracked | safe (tracked) | M1 |
| `.claude/rules/local/gitflow-lane-protocol.md` | local-only, never mirrored | safe (non-moai rules dir) | M2 |
| `CLAUDE.local.md` §4.1 | local-only, never mirrored | safe (root, not a managed root) | M3 |
| `.moai/docs/git-workflow-doctrine.md` | tracked | safe | M4 verify-only (no edit planned) |

**D1 — RESOLVED (operator decision (b), executed by the lead before this card's run-phase,
2026-09-02).** The primary's formerly-uncommitted §4.1 chain (2026-08-29 formalization,
표준 체인 + [SUPERSEDED] markings, plus the 2026-09-01 [HARD] WT-branch-push ban) landed
separately onto develop as `9a161687a` (+ `6b03e1757` origin/main absorb); develop tip =
`6b03e1757`. The lead absorbed origin/develop into this branch at `a04afea53` (clean, no
conflicts; version-management.md and gitflow-lane-protocol.md untouched in
`fa8ff89ba..6b03e1757`). M3's §4.1 pointer work targets the LANDED canonical text — the 08-29
chain replaced the 08-27 version, and the 08-27 transition record lives in git history.

**Superseded card premises (lead corrections 2026-09-02, honored throughout).**
- "develop은 원격에 올리지 않는다 / 카드 브랜치는 main에서 분기 / PR은 카드별로 main" — SUPERSEDED:
  standing remote `develop`, card worktrees branch FROM develop, integration via
  `git merge --no-ff` in `.claude/worktrees/develop` through the integration window.
- Card scope item (1) "§4.1 HARD 3번 개정" — ALREADY LANDED; acknowledged, not redone.
- Card scope item (3) "재생성 기준을 origin/main으로 명시" — corrected to **`origin/develop`**.

## §B Known Issues

- **B1 — C3 (git-workflow-doctrine.md self-contradiction): RESOLVED BY MEASUREMENT.** The
  worktree/develop copy carries the 2026-08-27 git-flow revision;
  `grep -n "Gitflow 패턴\|develop 브랜치 생성" .moai/docs/git-workflow-doctrine.md` at
  `fa8ff89ba` → no output (exit 1). Research C3's "lines 8/45/344 forbid develop" was read from
  the primary checkout's STALE copy — a checkout-divergence artifact, not an in-tree defect.
  M4 therefore VERIFIES consistency (re-measures zero residual prohibition lines post-M1-M3)
  and does not edit the file. No AC is adopted on this (a criterion green at RED time is
  vacuous — verification-completeness §2); the verification is recorded in progress.md §E.2
  instead.
- **B2 — CLAUDE.local.md byte budget.** Over the 40,000-char heuristic (47,688 bytes on the
  post-absorb tree `a04afea53`, carrying the landed 08-29 chain). M3 additions are pointer
  lines only; byte delta measured before/after and stated in the M3 completion report
  (rule-authoring growth duty: <1,000 bytes single-edit → duty does not fire, but state the
  number anyway).
- **B3 — C2 (who created v3.1.0-rc.0/.1 tags) is surfaced, not adjudicated.** The tag messages
  ("Local-only release candidate … NOT pushed") are primary evidence that they were local
  annotated tags — a counter-precedent PREDATING the current ldflags-only rule. REQ-RC-002
  records them as exactly that; the doc must not silently rewrite the history.
- **B4 — C4 (--merged applicability).** Develop-lane merges are mandated `--no-ff`
  (ancestry-preserving), so "squash makes --merged empty" is NOT asserted as fact for this lane.
  REQ-RC-004 phrases the criterion as merge-shape-agnostic and cites
  SPEC-WORKTREE-SQUASH-MERGE-001 only for why `--merged` is unreliable in general.
- **B5 — C5 (sentinel route reliability).** Incident memory undermines the operator-terminal
  sentinel route; REQ-RC-005 names worktree entry as the only dependably sanctioned route and
  the doc must not recommend the sentinel route to subagents.
- **B6 — Checkout divergence generally.** Primary vs worktree copies of CLAUDE.local.md and
  git-workflow-doctrine.md differ; all M-target greps are re-measured on the card worktree tree
  at run-phase start, not carried over from research.

## §C Pre-flight (run-phase start)

```bash
git rev-parse --short HEAD && git branch --show-current     # confirm WT-rc-testbed lineage
grep -c "Local RC Numbering" .moai/docs/version-management.md        # expect 0 (RED intact)
grep -c "develop 갱신" .claude/rules/local/gitflow-lane-protocol.md  # expect 0 (RED intact)
wc -c CLAUDE.local.md                                                # byte baseline for M3
```

If a RED probe returns ≥1, another actor landed overlapping content — re-baseline before editing
(blocker report to lead, not silent absorption).

## §D Constraints (DO NOT VIOLATE)

- PRESERVE: `gitflow-lane-protocol.md` §9 runbook body (single source — cross-reference only);
  `version-management.md` existing release-tag progression sections; the four other §4.1 HARD
  disciplines; `scripts/release.sh` and all release-harness surface.
- PRESERVE: exact English evidence-anchor tokens (below) — the ACs grep them verbatim.
- FORBIDDEN: creating git tags; pushing the WT branch (`git push origin WT-*` — operator
  directive 2026-09-01); edits to `internal/`, `cmd/`, `.github/`; re-landing or rewriting the
  landed §4.1 chain (D1-resolved, §A — the chain arrived via `9a161687a`; M3 adds pointers
  only); duplicating the §9 runbook into any other file.
- Commit discipline: explicit pathspec only; re-read HEAD + branch immediately before commit.

**Evidence-anchor tokens (prescribed literals — run-phase MUST emit them verbatim):**

| Token | Target file | AC |
|---|---|---|
| `Local RC Numbering` (section heading) | version-management.md; CLAUDE.local.md | AC-001, AC-007 |
| `counter-precedent` | version-management.md | AC-002 |
| `BUILD_ID` | version-management.md | AC-003 |
| `develop 갱신` (section heading) | gitflow-lane-protocol.md; CLAUDE.local.md | AC-004, AC-007 |
| `SPEC-WORKTREE-SQUASH-MERGE-001` (citation) | gitflow-lane-protocol.md | AC-005 |
| `BranchGuard` (route note) | gitflow-lane-protocol.md | AC-006 |
| `rc_version_format` | version-management.md | AC-008 |

## §E Self-Verification (run-phase completion)

Re-run all 8 AC commands from acceptance.md on the final tree; each must print ≥1 and exit 0.
Record per the 5-section evidence format: command, verbatim stdout, exit code, tree SHA.
Additionally: (a) `wc -c CLAUDE.local.md` before/after M3; (b) the B1 verification re-measurement
(zero residual prohibition lines in git-workflow-doctrine.md); (c) `moai spec lint` on the SPEC
directory (expect no error-severity findings).

## §F Milestones (priority order; no time estimates)

- **M1 — version-management.md: author the Local RC Numbering section** (REQ-RC-001/002/003/008;
  flips AC-001, 002, 003, 008). Highest-change-likelihood content: the two authored rules live
  here. Section MUST contain: (1) increment/reset policy (N starts 0 per target vX.Y.Z, +1 per
  candidate, resets only on target change); (2) no-tags rule + `v3.1.0-rc.0/.1` counter-precedent
  record; (3) BUILD_ID ordering warning citing SPEC-BINARY-LAG-VISIBILITY-001 (v3.1.2 newer than
  v3.1.3-rc.5); (4) moai-update re-application note for `git-strategy.yaml` keys
  (`rc_version_format`). Cross-reference — do not duplicate — the release-tag progression and
  gitflow-lane-protocol.md §9 runbook.
- **M2 — gitflow-lane-protocol.md: author the develop 갱신 section** (REQ-RC-004/005; flips
  AC-004, 005, 006). New numbered section between §10 and Cross-references. Content: refresh
  criterion keyed to `origin/develop`, merge-shape-agnostic, `--merged` not sufficient (cite
  SPEC-WORKTREE-SQUASH-MERGE-001); BranchGuard-safe route = launcher entry into
  `.claude/worktrees/develop` (`moai cc -w develop` / `EnterWorktree`), never primary-checkout
  `git branch`, never `git -C` (worktree-session guard); cross-reference §3 window mechanics and
  §9 rc runbook; note the sentinel route is not dependable for subagents (C5).
- **M3 — CLAUDE.local.md §4.1: pointer-only wiring** (REQ-RC-006/007; flips AC-007). At most 2
  pointer lines: one to the Local RC Numbering section, one to the develop 갱신 section of the
  lane protocol (which itself cross-references §9 — §4.1's existing discipline 4 already
  delegates to §9, so no third pointer). Both pointers quote the section names verbatim — they
  are the AC-007 grep anchors. Measure byte delta; report it. Targets the LANDED canonical
  §4.1 (08-29 chain, absorbed at `a04afea53` — D1 resolution §A).
- **M4 — verification sweep + B1 re-measurement** (no doc edits planned). Re-run AC greps, byte
  measurement, doctrine-consistency probe, spec-lint; populate progress.md §E.2 evidence;
  report per the 5-section format.

## §G Anti-Patterns

- **AP-1 (the card's representative mutant):** declare develop standing without documenting the
  rc build + clean reinstall path → AC-007 stays red. The pointer from the declaration site
  (§4.1) to the numbering policy is the mutant-killer.
- **AP-2: duplicating the §9 runbook** into CLAUDE.local.md or version-management.md — two copies
  diverge ("두 벌이 되는 순간 갈라진다").
- **AP-3: procedure bodies in §4.1** — byte budget already exceeded; pointers only.
- **AP-4: asserting "--merged is empty" for the develop lane** as established fact (C4) — the
  lane mandates `--no-ff`; state the criterion merge-shape-agnostically.
- **AP-5: recommending the sentinel route** to subagents (C5) — both exemptions are unreachable
  from tool-spawned subagents.
- **AP-6: re-landing or rewriting the landed §4.1 chain** — D1 resolved via `9a161687a`
  (option b); M3 adds pointers onto the landed text and touches nothing else in the chain.
- **AP-7: ordering builds by version string** — BUILD_ID is the identity (REQ-RC-003).

## §H Cross-References

- research.md (same directory) — 4-lens synthesis, contradictions C1-C5, gaps.
- `.moai/docs/version-management.md` — release-tag progression, ldflags doctrine (cross-ref target).
- `.claude/rules/local/gitflow-lane-protocol.md` §3 (window), §9 (rc runbook).
- `CLAUDE.local.md` §4.1 (model), §2.3 (moai-update wipe), §11 (exit-137 known issue).
- `.claude/rules/moai/development/rule-authoring.md` — growth duty for M3.
- `.claude/rules/moai/workflow/main-checkout-branch-guard.md` — BranchGuard exemptions unreachability.
- SPEC-WORKTREE-SQUASH-MERGE-001 — `--merged` reachability blindness.
- SPEC-BINARY-LAG-VISIBILITY-001 — version-string ordering incident.
- SPEC-WORKTREE-ENTRY-STRATEGY-001 — worktree-entry canonical forms.
