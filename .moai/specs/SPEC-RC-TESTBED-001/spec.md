---
id: SPEC-RC-TESTBED-001
title: "Local develop RC testbed — rc.N numbering policy, develop refresh procedure, clean-reinstall procedure codification"
version: "0.1.0"
status: draft
created: 2026-09-02
updated: 2026-09-02
author: manager-spec
priority: P2
phase: "v3.1.5"
module: "CLAUDE.local.md; .moai/docs/version-management.md; .claude/rules/local/gitflow-lane-protocol.md"
lifecycle: spec-anchored
tags: "release-candidate, develop, gitflow, versioning, docs"
tier: M
---

# SPEC-RC-TESTBED-001 — Local develop RC testbed

## §1 Background & Problem

Card t281 (operator decision 2026-08-26, "C안") promoted local `develop` to a standing
integration + RC testbed. The git-flow reversal has since landed in the tree: `develop` is a
standing REMOTE branch (`origin/develop` exists; CI triggers on `[main, develop]`), card
worktrees branch FROM develop, and card integration is `git merge --no-ff` inside the single
integration worktree `.claude/worktrees/develop` through the `moai integration acquire/release`
window (canonical rules: `.claude/rules/local/gitflow-lane-protocol.md`; model:
`CLAUDE.local.md` §4.1, reversal provenance 2026-08-27). This SPEC acknowledges that reversal as
landed and does NOT redo it.

What remains missing — confirmed by the 4-lens research (research.md §2-§4: "NONE found" on
every lens that searched) — is exactly two authored rules plus the pointer wiring that makes
them reachable from the testbed's declaration site:

1. **rc.N increment/reset policy for LOCAL (untagged, ldflags-injected) rc builds cut from
   develop.** The release-tag progression (`rc.0 … rc.N → vX.Y.Z`, SemVer, no leading zero)
   already exists in `.moai/docs/version-management.md`; the LOCAL-build policy — when N
   increments, when it resets to 0, relation to release tags — exists nowhere.
2. **develop refresh/regeneration procedure after card merges.** §4.1 and the lane protocol
   cover merging INTO develop; refreshing local develop from `origin/develop` — by what
   criterion and through which BranchGuard-safe route — is documented nowhere.

The representative failure this SPEC exists to prevent (card's own mutant): an implementation
that declares develop standing WITHOUT documenting the rc build + clean reinstall procedure —
forcing every round to re-derive the build command and the clean-reinstall order, and re-hitting
the cp-over exit-137 failure.

## §2 Scope

**In scope (doc targets — run-phase writes; this SPEC specifies):**

- `.moai/docs/version-management.md` — new "Local RC Numbering" section (REQ-RC-001, 002, 003, 008).
- `.claude/rules/local/gitflow-lane-protocol.md` — new develop-refresh section (REQ-RC-004, 005).
- `CLAUDE.local.md` §4.1 — pointer-only additions (REQ-RC-006, 007).

**Verified-consistent, no edit planned (measured at tree `fa8ff89ba`):**
`.moai/docs/git-workflow-doctrine.md` carries the 2026-08-27 git-flow revision with zero residual
develop-prohibition lines (research C3 was a checkout-divergence artifact; see plan.md §B).

## §3 Requirements (GEARS)

**REQ-RC-001** (Ubiquitous) — The local RC numbering policy in `.moai/docs/version-management.md`
shall define `rc.N` for local (untagged, ldflags-injected) builds cut from develop as: `N`
starts at `0` for each new target `vX.Y.Z`, increments by exactly 1 per candidate build cut for
that target, and resets to `0` only when the target `X.Y.Z` changes — the next release line's `N`
starts fresh regardless of how far the previous line climbed.

**REQ-RC-002** (Ubiquitous) — The local RC numbering policy shall state that local rc builds
create no git tags — tagging remains a remote-facing act performed only by the release harness —
and shall record the `v3.1.0-rc.0`/`-rc.1` local-annotated-tag episode as a counter-precedent
predating the current rule (the v3.1.3 line ran entirely tagless).

**REQ-RC-003** (Ubiquitous) — The local RC documentation shall warn that the version string does
not order builds: the monotone build identity is `BUILD_ID`, and an explicit rc VERSION reads
HIGHER than a later default build (incident SPEC-BINARY-LAG-VISIBILITY-001 — an installed
`v3.1.2` binary was actually newer than the prior `v3.1.3-rc.5`).

**REQ-RC-004** (Event-driven) — **When** a card merge lands on `origin/develop`, the develop
refresh procedure in `.claude/rules/local/gitflow-lane-protocol.md` shall key the local-develop
regeneration criterion to `origin/develop`, stated merge-shape-agnostically, and shall not
present `git branch --merged` as a sufficient criterion (reachability blindness; cross-reference
SPEC-WORKTREE-SQUASH-MERGE-001 for why `--merged` is unreliable, without asserting "--merged is
empty" as an established fact for the develop lane, whose merges are mandated `--no-ff`).

**REQ-RC-005** (Capability gate) — **Where** BranchGuard is enabled for the primary checkout
(`workflow.yaml branch_guard.enabled: true`, local opt-in), the develop refresh procedure shall
route regeneration through launcher entry into the single integration worktree
`.claude/worktrees/develop` (`moai cc -w develop` / `EnterWorktree`), and shall not route it
through primary-checkout `git branch` / `git checkout` (both guard exemptions are unreachable
from tool-spawned subagents) nor through cross-tree `git -C` (denied by the worktree-session
guard — entering is the only sanctioned path).

**REQ-RC-006** (Ubiquitous, no-duplication) — The documentation set shall not duplicate the rc
build + clean-reinstall runbook: `CLAUDE.local.md` §4.1 shall carry pointer lines to (i) the
Local RC Numbering section of `.moai/docs/version-management.md` (numbering policy) and (ii) the
develop 갱신 (refresh) section of `.claude/rules/local/gitflow-lane-protocol.md` — which itself
cross-references §9, whose runbook body shall remain the single source ("두 벌이 되는 순간
갈라진다"). A separate §4.1 → §9 pointer is NOT additionally required: §4.1's existing
discipline 4 already delegates to §9, so a third pointer would be redundant indirection.

**REQ-RC-007** (State-driven) — **While** `CLAUDE.local.md` exceeds its 40,000-character
heuristic budget (47,820 bytes in the worktree copy at research time), §4.1 additions made for
this SPEC shall be pointer lines only, with the byte growth measured and stated per the
rule-authoring growth duty (`.claude/rules/moai/development/rule-authoring.md`).

**REQ-RC-008** (Ubiquitous) — The local RC documentation shall note that
`.moai/config/sections/git-strategy.yaml` keys (`rc_version_format: vX.Y.Z-rc.N` among them) are
reset by `moai update` to template defaults and must be re-applied after every update
(CLAUDE.local.md §2.3).

## §4 Constraints

- **Template-First Rule does not apply to two of three targets**: `CLAUDE.local.md` and
  `.claude/rules/local/gitflow-lane-protocol.md` are deliberate local-only files, never mirrored
  to `internal/template/templates/` (no distribution impact — card note "로컬 전용"). Only
  `.moai/docs/version-management.md` is tracked, moai-update-safe, and already distributed-facing.
- **Pointer-not-body discipline**: research §7 — gitflow-lane-protocol.md §9 already owns the
  rc-build runbook verbatim; cross-reference, never duplicate.
- **Language**: doc-target prose follows each file's existing convention (mixed ko/en);
  grep-able canonical tokens prescribed in plan.md §F are English literals and MUST appear
  verbatim (they are the AC evidence anchors).
- **No implementation in plan-phase**: doc targets are run-phase write surfaces.

## §5 Out of Scope

### Out of Scope — §4.1 HARD-3 revision (already landed)

- The develop-standing revision of CLAUDE.local.md §4.1 HARD 3 is already committed on develop
  (reversal provenance 2026-08-27); this SPEC does not re-word, re-litigate, or redo it.
- The four remaining §4.1 HARD disciplines are untouched.

### Out of Scope — release harness and tagging mechanics

- Release-tag progression (rc.0 → rc.N → vX.Y.Z), `scripts/release.sh`, release-branch creation,
  and the release PR flow — owned by `.moai/docs/version-management.md`'s existing sections and
  the release harness; this SPEC only cross-references them.
- No new tags, no tag tooling, no release-branch work.

### Out of Scope — landing the primary checkout's uncommitted §4.1 chain

- The primary checkout's working copy carries an UNCOMMITTED 2026-08-29 §4.1 chain formalization
  that exists on no branch; absorbing it into this SPEC's run-phase is the lead's call
  ([NEEDS CLARIFICATION] marker in plan.md §A), never a silent side-effect.

### Out of Scope — docs-site Vercel production-branch binding

- The unresolved Vercel preview/production reaction to develop-side docs-site changes is carried
  as residual risk by the existing model (§4.1.3); not tested or decided here.

### Out of Scope — Go code, CI workflows, hooks

- Zero code changes; no workflow/hook edits; no `internal/` or `cmd/` surface.

## §6 HISTORY

| Date | Author | Change |
|------|--------|--------|
| 2026-09-02 | manager-spec | Plan-phase creation (card t281, Tier M). Research: 4-lens synthesis persisted as research.md. RED-now baselines measured at tree `fa8ff89ba`. |
| 2026-09-02 | manager-spec | Plan-audit iter-1 fixes: D2 — REQ-RC-006 relaxed to the (numbering, develop-갱신) pointer pair via the refresh section's §9 cross-reference (option a; §4.1 discipline 4 already delegates to §9), AC-RC-007 now verifies both anchors; D3 — AC-RC-005 mutant-probe wording corrected to sync-audit disclosure form. |
