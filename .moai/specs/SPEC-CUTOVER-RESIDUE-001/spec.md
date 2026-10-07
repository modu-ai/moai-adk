---
id: SPEC-CUTOVER-RESIDUE-001
title: "GitHub Flow cutover residue cleanup: sweep base fallback, git-strategy config completion, workflow branch filters"
version: "0.1.0"
status: in-progress
created: 2026-10-07
updated: 2026-10-07
author: t1564 card worker
priority: P1
phase: "v3.3.0 target"
module: "internal/cli/worktree"
lifecycle: spec-anchored
tags: "cutover, github-flow, worktree-sweep, git-strategy, ci-workflows, residue"
tier: M
related_specs: [SPEC-WORKTREE-SWEEP-001, SPEC-GITSTRAT-WORKFLOW-READER-001, SPEC-WORKTREE-BASEREF-001]
---

# SPEC-CUTOVER-RESIDUE-001

## §A — History

- **2026-10-07** — plan-phase v0.1.0 authored from card t1564 (factory run tmhxo0). The 2026-10-05 GitHub Flow cutover moved the integration base to `main` and deleted `origin/develop`, but three measured surfaces still carry the git-flow-era shape. Card class C (design change across subsystems), Tier M, cycle tdd for the code component.
- **2026-10-07, run-phase M2 amendment** — the plan's REQ-CR-005 (flip the local config to github-flow, drop the git-flow-era keys) collided at verification with a designed tripwire: `TestGitHubFlowSweepGuard` (SPEC-GITHUB-FLOW-DEFAULT-001, card t1453 M4) arms ONLY when the local git_strategy workflow resolves github-flow, and its armed assertions cover exactly the develop-era live text this card's §E.1 defers to the residue-sweep card. Flipping the value without that sweep turned the guard red (observed: `--- FAIL: TestGitHubFlowSweepGuard` on the flipped tree). The value's flip therefore belongs to that card's own M5; REQ-CR-005 is revised to keep the operator keys + a branch-neutral counting comment and leave the workflow untouched. The sweep is value-independent through the M1 fallback (measured GREEN live with the git-flow config before the flip was attempted).

## §B — Problem

Three measured residues, each observed in this run against tree `cb2a011d0`:

1. **Sweep landing determination fails repo-wide.** `moai worktree sweep` derives its default base from `LoadGitFlowIntegrationConfig(root).IntegrationTarget`; the repo-local tracked `git-strategy.yaml` still declares `manual.workflow: git-flow` + `develop_branch: develop`, so the derived base is `origin/develop`. `git ls-remote --exit-code origin refs/heads/develop` exits **2** — the ref does not exist on the remote (deleted at cutover) — so `sweepFetchBase` runs `git fetch origin develop`, exits 128, and every tree is preserved with `cause=fetch-failed` (measured 2026-10-07 tick 3, `.moai/reports/tmf011-mission-20261006.md`: 187 worktrees). The remote default branch is `main` (`git ls-remote --symref origin HEAD` → `ref: refs/heads/main`, exit 0), which the sweep never consults.
2. **Config + re-apply doctrine still declare git-flow.** The tracked `.moai/config/sections/git-strategy.yaml` carries `workflow: git-flow`, `develop_branch: develop`, and a `lead_push_threshold` comment counting `origin/develop..develop`; `.moai/docs/update-local-file-survival.md` §2.3 re-apply still instructs `git restore --source=develop` and lists `worktree_base_branch: develop` — both now-wrong guidance (the cutover doctrine in `AGENTS.local.md` §4.1/§2.3 names `main` as base and `main`'s committed copy as the re-apply source; the survival doc was not swept in the cutover). The template's `lead_push_threshold` comment likewise names a flow-specific branch.
3. **Workflow branch filters reference the deleted ref.** 8 files under `.github/workflows/` carry `branches: [main, develop]` filters (test-install.yml:5, codeql.yml:5, ci.yml:18, graph-freshness.yml:9,15,17, judgment-first-consistency.yaml:34, workflow-parse-guard.yaml:14, lsel-leak-guard.yaml:11, template-neutrality-check.yaml:31), and docs-i18n-check.yml still lists `develop` in its trigger with a git-flow-era comment. The filters are harmless for existing pushes but mis-describe the branch model and will confuse trigger reasoning.

## §C — Goal

The sweep resolves its landing base resiliently: when the derived default ref does not exist on the remote, fall back to the remote's default branch (derived from the remote itself — no hardcoded branch name) with a visible notice; develop-based projects keep their configured base unchanged. The repo-local config, its re-apply doctrine, and the template comment state the cutover shape (github-flow, `worktree_base_branch: main`). Workflow branch filters name only branches that exist.

## §D — Requirements (GEARS)

8 requirements. `<subject>` is the affected surface named per requirement.

### D.1 — M1: Sweep base fallback

- **REQ-CR-001** (Ubiquitous) — The sweep shall, when deriving its default base from config, probe the configured remote ref's existence via `git ls-remote --exit-code <remote> refs/heads/<ref>`; the probe result classifies the derived ref as present (exit 0), absent (exit 2), or undeterminable (any other exit).
- **REQ-CR-002** (State-driven) — While the derived ref is present, the sweep shall use the derived base unchanged (develop-based projects are unaffected).
- **REQ-CR-003** (Event-driven) — When the derived ref is absent, the sweep shall resolve the remote's HEAD default branch via `git ls-remote --symref <remote> HEAD`, use `<remote>/<head-branch>` as the effective base, and emit a one-line notice naming both bases; when existence or HEAD resolution is undeterminable, the sweep shall keep the derived base and let the existing three-way landing contract (fetch failure → PRESERVE, SPEC-WORKTREE-SWEEP-001 REQ-WS-003) apply — never switch to an unverified base silently.
- **REQ-CR-004** (Unwanted) — An explicit `--base` flag shall bypass the fallback entirely: the operator's word is the base, and an absent explicit ref keeps the existing fetch-failed PRESERVE semantics.

### D.2 — M2: Config and doctrine residue

- **REQ-CR-005** (Unwanted, revised at run-phase M2 — see §A amendment note) — The repo-local tracked `.moai/config/sections/git-strategy.yaml` shall keep its operator keys correct (`worktree_base_branch: main`, `lead_push_threshold` at its operator-given value) and its `lead_push_threshold` counting comment shall name no flow-specific branch. `manual.workflow` shall REMAIN `git-flow` until SPEC-GITHUB-FLOW-DEFAULT-001 M5 flips it together with its residue sweep: the workflow value is that guard's arming switch (`TestGitHubFlowSweepGuard` asserts live develop-base text only when it resolves github-flow), and this repository still carries the deferred develop-era rules/docs (§E.1). The sweep does not depend on the value either way (REQ-CR-003's fallback).
- **REQ-CR-006** (Ubiquitous) — The re-apply procedure in `.moai/docs/update-local-file-survival.md` §2.3 shall name `main`'s committed copy as the re-apply source and verify the post-apply state with greps for `workflow: github-flow` and `worktree_base_branch: main`; the develop-era `--source=develop` instruction and its HEAD-vs-develop rationale shall be superseded with the cutover answer.
- **REQ-CR-007** (Ubiquitous) — The template `git-strategy.yaml.tmpl` comment on `lead_push_threshold` shall describe the counting command generically (`origin/<integration branch>..<integration branch>`), carrying no flow-specific branch name (template neutrality §25.1).

### D.3 — M3: Workflow branch filters

- **REQ-CR-008** (Unwanted) — No workflow branch filter under `.github/workflows/` shall list `develop`; each affected filter becomes `branches: [main]` (or the push/pull_request equivalent), and docs-i18n-check.yml's develop-specific trigger entry and its git-flow-era comment block are aligned with the cutover.

## §E — Out of Scope

### E.1 Deferred
- Sweeping every historical comment that mentions a dead "develop run" ID in ci.yml / release-pr-multi-os.yml (cosmetic; comments do not alter behavior).
- Migrating `.claude/rules/local/gitflow-lane-protocol.md` and `.moai/docs/gitflow-integration-chain.md` (both already carry the cutover drift notice in `AGENTS.local.md` §4.1 — a separate card's scope if the operator wants the bodies rewritten).
- Changing the primary checkout's deployed (untracked) `.moai/config/sections/git-strategy.yaml` — runtime-managed, shared-tree; the re-apply doctrine lands here and the operator/next `moai update` cycle applies it.

### E.2 Explicitly excluded
- Any change to `develop_branch`'s loader semantics (`LoadGitFlowIntegrationConfig` gating) — the gating is correct; only the local config's declared value is the residue.
- Branch deletion on the remote or in local trees.

## §F — Design Decisions

- **Fallback is derived-only, never explicit**: explicit `--base` means what it says (REQ-CR-004). The fallback exists so config drift cannot again silently disable the sweep, not to second-guess operators.
- **Probe before fetch, never parse fetch errors**: `ls-remote --exit-code` exit 2 is a structured "no such ref" answer; classifying by fetch stderr text would be message-fragile across git versions.
- **The fallback resolves the remote's own HEAD** — a generic answer from the remote, the same source `git clone` uses — so no branch name is hardcoded anywhere in the change.
- **Notice on stderr** keeps stdout machine-readable (`--json` inventory untouched).
