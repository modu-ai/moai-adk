---
description: "AGENTS.local.md detail companion (2026-10-19 size diet, card t1900) — moved procedure-and-reference mass: historical local-main batch integration, historical integration chain (main), the three agent-definition copies, and the consolidated references index. Conditional (paths-scoped) — never joins the always-loaded surface."
paths: "AGENTS.local.md,.claude/rules/local/*.md,.moai/specs/**,.moai/docs/git-workflow-doctrine.md,.moai/docs/git-local-workflow-doctrine.md,.moai/docs/gitflow-integration-chain.md,.claude/skills/moai/workflows/run.md,.claude/skills/moai/workflows/sync.md,internal/cli/**,internal/hook/**,internal/template/**,.github/workflows/**"
---

# AGENTS.local.md — moved detail (local-only)

> **2026-10-19 size diet (card t1900, SPEC-AGENTS-CONTRACT-REVISE-001).** The four largest
> AGENTS.local.md chunks moved here under the leader-corrected always-loaded model (the
> decision-index measurement table): §4.0's superseded integration items → §A, §4.1 → §B,
> §2.0/§2.1 → §C, References §18-27 → §D. Every moved [HARD] clause is preserved in this file
> and leaves a one-line pointer or essence at its original location. This is a local-only file —
> no template mirror.
>
> **Transition scope.** §A and §B are historical records — superseded for current card
> delivery by `AGENTS.local.md` §4.0 (the per-card draft-PR policy). They carry no execution
> authority over current card delivery, and their superseded clauses (no-card-PR, local-only
> closure, leader bulk push, wait-until-batch dispatch) are not procedures. Preservation,
> one-writer, ownership, and resource-lease safeguards remain binding regardless of section;
> do not replay historical commands to bypass a current refusal.

## §A Historical local-main batch integration (former §4.0 items 1-5)

> **[HISTORICAL — superseded for card delivery by `AGENTS.local.md` §4.0 (2026-10-19); no
> execution authority over current card delivery.]** The batch release procedure and
> batch-closing rule (former items 6-7) are NOT here — they remain current as
> `AGENTS.local.md` §4.0-R until issue #1830 processing.

`main` is checked out in exactly one dedicated integration worktree, at a fixed path (currently `.claude/worktrees/main`). That worktree performs landing and resync only, plus the operator-requested rc build (lane protocol §9): it is never used to edit or experiment, and it stays clean. The primary checkout is the human and leader workspace and does not check out `main`. Clauses marked in the former §4.1 were superseded by this section; clauses without a marker remained in force as written.

**[HARD] 1. Branch base** *(historical)*

Card worktrees branch from local `main`. Because origin/main can fall behind local `main`, it is not used as the branch base. In this repository, this section replaced the template sentence "Create the fresh tree from the remote default branch." The distributed template does not change. A card worktree is created only through `moai factory next` or `moai worktree new`. When the configured base (`git_strategy.worktree_base_branch`, `main` here) resolves, it branches from the named ref `main` and not from HEAD; when the base is empty or does not resolve, creation falls back to the invoking tree's HEAD. Claude Code's default `--worktree` (remote default branch) is not used for card work. After creation, confirm that the new tree's HEAD equals `main`; that confirmation, not the creation path, is the check.

**[HARD] 2. Merge landing** *(historical — superseded for card delivery; the verb and window remain the mechanism of the retained release-path resynchronization)*

- The landing surface is the integration worktree that holds `main`; the tool refuses when no tree holds it. `workflow.local_main_integration.enabled` is `true` today, so the tool accepts the primary as holder; changing it is planned in t1847.
- Merging is performed only by `moai integration merge --card <id>`. `moai factory complete` calls this command from lane sessions. Agents writing `git merge` directly in Bash is still refused. This path is a designed exception.
- The tool (1) takes the integration window, (2) reads HEAD symbolically and confirms that it points at `main`, (3) applies the clean check of item 3, and (4) pins the tip commit of the card branch; then it refuses if any ignored or untracked file sits on the paths the merge will write or on their ancestor or descendant paths. The tool does not delete or stash files. It then (5) performs a no-fast-forward merge with autostash turned off, (6) compares the first parent, HEAD, and the status set against their pre-merge values, and (7) records the merge SHA and tree and releases the integration window (`moai integration release`).

**[HARD] 3. Clean-tree rule** *(historical origin; the safeguard itself remains binding)*

A landing tree must be clean: `git status --porcelain=v1 -z --untracked-files=all` must be empty, or the merge is refused. A separate integration worktree gets the overlap rule (REQ-LMF-005). Diagnose each untracked or modified file in the primary checkout (commit it in its owning card worktree, move it there, or ignore it); never stash. Read `main` source through `git show main:<path>` or in the integration worktree, not from the primary's working tree; it does not cover a smudge or clean filter, end-of-line conversion, uncommitted changes, or a later move of `main`.

**[HARD] 4. Resynchronization** *(historical origin; retained under `AGENTS.local.md` §4.0-R item 4)*

When a release batch lands on `origin/main`, local `main` follows it. The tool performs this as a fast-forward-only update inside the integration window; it is not done from the operator's terminal. `moai integration resync` refuses when another tree holds `main` and no moai verb moves the primary checkout to the current local `main` tip; both are planned in t1847.

- (1) Take the window, and confirm that HEAD points at `main` and that the tree being updated is clean.
- (2) Run `git fetch origin main` and check its exit status. Pin the `origin/main` value obtained as the baseline, and make every later remote comparison against this value.
- (3) Read the ancestry relation between HEAD and the baseline. If HEAD is behind the baseline, check for ignored-file overwrites, then fast-forward. If the baseline is an ancestor of HEAD, or the two values are the same, report that no update is needed and leave HEAD where it is. If neither is an ancestor of the other, the histories have diverged: refuse, and HEAD does not move.
- (4) After the move, re-read HEAD to confirm that it equals the baseline and still points at `main`; then report the old and new SHAs and release the window.

The handling of the diverged case is set by operator decision Q2 (`.moai/specs/SPEC-LOCAL-MAIN-FLOW-001/decision-index.md`).

**[HARD] 5. Closing and worktree disposal** *(historical — superseded: closure requires the confirmed remote merge of the card's PR and the completion evidence, per `AGENTS.local.md` §4.0 item 5)*

- Closing a card merged on the local `main` surface does not push, and does not wait for a release batch.
- Disposing of a card worktree waits until the batch carrying that merge has landed on origin. Until then the card branch has no remote copy, so the worktree is the only copy of the work.
- The closing conditions are as follows. The merge commit must have two parents, its tree must equal the tree of the second parent, and it must be reachable from the integration branch. A re-measurement record keyed by that tree must pass verification. The decider is a human, and the leader's approval receipt must be verified. An evidence file produced by a lane cannot substitute for the re-measurement record.
- Implementation of this rule is owned by the follow-up card t1621. Until t1621 opens the done transition, merged cards cannot reach done.

Basis: SPEC-LOCAL-MAIN-FLOW-001 (spec REQ-LMF-010·REQ-LMF-011·§G; plan §B2·§B3·§B3a·§B4·§B7·§B9·§G.1).

## §B Historical integration chain (main) (former §4.1)

> **[HISTORICAL — superseded for card delivery by `AGENTS.local.md` §4.0 (2026-10-19); no
> execution authority over current card delivery.]**

This closes the gap where each card is verified and merged on its own, but **no one checks the combined state**. On 2026-08-15, 12 PRs each entered main with a green check, and only after they converged did it come to light that `moai update` deletes local-only files. With the 2026-10-05 GitHub Flow transition (M2), the integration base is `main` — the chain below is the post-transition standard, and the develop chain remains as history, marked [SUPERSEDED] below.

**[SUPERSEDED by the former §4.0]** **[HARD] Standard chain — GitHub Flow transition (operator instruction, M2 landed 2026-10-05)**

```
origin/main
   ↓ fetch
local/main  ← default branch. Base of card branches
   ↓ branch
worktree (per card, WT-<slug>)
   ↓ PR (base: main)  ← on card completion — through the integration window (serial); the leader merges
origin/main  ← CI issues the integration verdict
   ↓ fetch
local/main
```

**[HARD] Rules**

1. **[SUPERSEDED by the former §4.0]** **Card branches are cut from `main`.** When a card finishes, it is applied to `main` through a card PR (base `main`). `develop` is legacy — do not use it as the base of a new card.
2. **[SUPERSEDED by the former §4.0]** **`main` is the remote default branch.** The CI on its head produces the integration verdict. Merges and pushes are performed **in bulk by the leader** (2026-09-02 — lanes do not push; the only exception is the candidate path (`ci/<card-id>`), which the 'candidate-path exception' below substitutes for this prohibition).
3. **The integration window is serial.** `moai integration acquire --card <card-id>` → merge → `release` — the push is outside the window (leader bulk, 2026-09-02). A lock is a device for serializing merges, not for serializing repairs, so if repairs remain, take the window only once they are ready.
4. **The verdict is CI.** A local pass is only an early signal — it is neither a clean environment nor the darwin/windows matrix. Do not reuse pre-merge verification as post-merge evidence: either measure again on the merge tree, or show that `git rev-parse <merge>^{tree}` of the merge commit equals the tree that was re-measured.
5. **[SUPERSEDED by the former §4.0]** **Local `main` is commit-dead (SPEC-MAIN-COMMIT-BAN-001, card t1337).** No session commits inside `main` in the primary checkout — `git commit` / `git revert` / `git cherry-pick` are refused by the BranchGuard (`workflow.branch_guard.deny_commits_on: [main]`), and commits are made only in card worktrees branched from `main`. The leader's push to `main` is closed by the batch trigger — when `git rev-list --count origin/main..main` reaches `git_strategy.manual.lead_push_threshold`, the batch is closed (the count comment in the configuration file still points at `origin/develop..develop` — drift). The source of the current value is the configuration file, and this rule only names the key and the count command. Pushes follow the green condition — per-card re-measurement of the merge tree in the integration window is the pre-gate for each card merge, and while the `origin/main` CI of the last push is red, the next push is held. Details: `.claude/rules/local/gitflow-lane-protocol.md` §4 (that rule still describes develop — drift).

**[SUPERSEDED by the chain above — 2026-10-05]** The develop GitFlow chain of 2026-08-29 to 2026-10-05 (cards branched from develop and merged back to develop with `--no-ff`, with only release PRs to main), and the regime before it (develop not pushed to the remote · a main PR per card · a disposable develop), were all discarded. The record of the discard and its reasons exists so that two documents do not issue contradictory instructions; the full text of the develop chain is in `.moai/docs/gitflow-integration-chain.md`.

**[HARD] Post-transition drift — the procedure document set.** `.moai/docs/git-workflow-doctrine.md`·`.moai/docs/git-local-workflow-doctrine.md`·`.moai/docs/gitflow-integration-chain.md`·`.claude/rules/local/gitflow-lane-protocol.md` still describe the develop chain. The canonical source of post-transition norms was the former §4.1 together with the transition rules (`.claude/rules/moai/**`, card t1522) — since 2026-10-19 it is `AGENTS.local.md` §4.0 — and when you open those documents, read them under this notice.

**Related document pointers (SPEC-RC-TESTBED-001)** — the procedure text is not kept here (as with the delivery.md delegation in rule 4 above, two copies diverge the moment they exist):

- The `rc.N` numbering policy, the untagged principle, and `BUILD_ID` build identification for local rc builds: `.moai/docs/version-management.md` — the **Local RC Numbering** section
- Updating local `main` after a merge (criterion and the BranchGuard-safe path): `.claude/rules/local/gitflow-lane-protocol.md` §11 (that section was written against develop; the criterion is the `origin/main` rev-list comparison)

**[HARD] Lane obligations in -f / -l mode**

In Factory mode (leader `moai cc -f` · lane `moai cc -l`), when a lane finishes its card work it **must request that the leader apply the change to `main`.** A lane does not take the merge window on its own.

- **Self-dispatch lane exception — merge window.** A lane of a Claude self-dispatch factory run does not make the request above — through the integration procedure of `moai factory complete`, which calls `moai integration merge --card`, it takes the integration window itself and applies its own card to `main` (OD-2). Codex lanes are not covered by this exception: they stop at merge-ready (REQ-SD-025). This exception also does not change the prohibitions above on other queue changes (`add`, `drop`, `done`, `edit`, etc.) and on `moai contract sign` (only card leasing with `moai factory next` is allowed — OD-1; `.claude/rules/local/gitflow-lane-protocol.md` §6).
- Contents of the completion report: card id · branch and HEAD · local merge SHA · number of unpushed commits · evidence path (whether it was moved out of the primary) · re-measurement scope
- **[SUPERSEDED by the former §4.0]** `moai integration status` showing `free` is **not an approval.** Under the open policy, a lane promoted to the front of the queue that holds the window merges without a leader nomination.
- **Lanes do not git-merge into `main` by hand; they merge only with moai integration merge --card, or with moai factory complete, which calls it.** The queue schedules the window, and the leader sets only the policy with `moai integration policy open|hold` (to receive the window, the lane queues via acquire --wait — re-measurement happens before waiting; inside the window, only the merge).
- **[SUPERSEDED by the former §4.0]** When the window is granted: `moai integration acquire --name <lane> --card <card-id>` → absorb `git merge main` in your own worktree (the target is the **local** `main` — not the remote; before absorbing, first check whether that local main is up to date — the criterion and update path are in `.claude/rules/local/gitflow-lane-protocol.md` §11) → **re-measure on the merge tree** → enter the integration worktree → apply to `main` (merge the card PR with base `main`, or a `--no-ff` merge — the migration document sets the detailed form) → `moai integration release` → `ExitWorktree keep` → completion report (report the applied SHA to the leader — the push and the merge are done in bulk by the leader)
- **[SUPERSEDED by the former §4.0]** **[HARD] The card PR (base `main`) is the only public path (revised with the 2026-10-05 transition).** When a card closes, the lane pushes the card branch and opens a PR (base `main`), but **does not merge it directly** — the merge and push are performed in bulk by the leader, and the lane is not the party that does so. A lane also does not directly request or re-request CI, such as `gh run rerun` or `workflow dispatch` — it leaves the CI verdict to the run triggered by the application to `main`, and reading the verdict is the leader's job. (The former prohibition — operator instruction 2026-09-01, the WT-branch push ban and the develop batch-push regime — became history together with the develop chain; this clause was added in response to a precedent: lane-2 pushed `WT-version-stamp-predicate` to origin that same day.)
- **[HARD] Candidate-path exception — operator approval 2026-10-03 (SPEC-CANDIDATE-CI-001, REQ-CCI-016).** `moai integration candidate --card <card-id>` is the path that pushes a candidate commit to `ci/<card-id>` to receive a CI verdict before landing. For this one path only, it replaces `WT push · CI request prohibited`. The landing gate requires a green verdict for that candidate. Pushing the integration branch and merging remain a bulk action by the leader, and the candidate verb does not take the integration window. The same intent is recorded in `.claude/rules/local/gitflow-lane-protocol.md` §2·§4 and in three documents under `.moai/docs/`, and the develop descriptions and drift notices in those documents remain as they are.
- **[HARD] `acquire` verifies the caller's tree before it records the window.** It checks whether the working copy of the tracked `.claude/settings.json` has been modified, using `git --no-optional-locks status --porcelain -- .claude/settings.json`. On a hit, it preserves that copy under `.moai/state/settings-drift/` in the primary checkout, appends one line to `ledger.jsonl` in the same place, and prints the preserved path and sha256. **Detection, preservation, and the ledger run every time regardless of configuration** (the problem was that nobody looked for 9 days, not that nothing blocked it). Only refusal is opt-in (`workflow.settings_drift_gate.enabled`; this repository turns it on), and the bypass is `--allow-settings-drift` — `--force` belongs to a different axis, taking the window away from a live holder, so it is not used as a bypass. To check by hand regardless of the window, use `moai integration preflight [path]`. **Never restore automatically under any circumstances** — the file is written by the runtime and can contain tokens, absolute paths, and tmux pane ids, so automatic restoration would itself destroy data. When a hit is reported, the leader decides the disposition.
- Do not dispose of a worktree until its remote merge has been confirmed. The worktree of an unpushed branch is the only copy of its work.
- sync is finished **before** merging, inside the worktree. If you close only run and merge, the SPEC reaches `main` as `in-progress`, and the window must be taken again (a real case on 2026-08-29, t342).

**The operating procedures below were moved to `.moai/docs/gitflow-integration-chain.md` (card t750) — the bash that executes the window (entering the integration worktree, absorbing, re-measuring, merging), the leader's bulk push procedure, the record of rejecting local CI, the friction from over-matching BranchGuard queries, and a caution about the docs-site Vercel binding. The procedure text in that document still describes the develop chain — when executing, read it under the transition notice above.** As long as the [HARD] rules above and the lane obligations stay unchanged, this summary is sufficient; when you execute a procedure, open the migration document.

## §C The three copies of the agent definitions (former §2.0/§2.1)

### [HARD] There are three copies of the agent definitions, and the third is not edited by hand

Agent definitions exist in three copies, but **only one of them is not a hand-edit target.**

| Copy | Path | Nature |
|---|---|---|
| C1 | `.claude/agents/moai/*.md` | Local dogfooding, hand-edited |
| C2 | `internal/template/templates/.claude/agents/moai/*.md` | Distribution mirror, hand-edited — **the neutral source layer** |
| C3 | `internal/template/templates/.codex/agents/moai/*.toml` | **Machine-emitted from C2** (`internal/template/agentemit`) |

[HARD] **If you have edited `internal/template/templates/.claude/agents/moai/*.md`, run `make agents-emit`.** Otherwise `internal/template/templates/.codex/agents/moai/*.toml` stays stale, and a binary built in that state runs with the old definitions embedded. C1 ↔ C2 is **not** a byte-identical relationship (an intentional divergence); the generation relationship runs in one direction only, C2 → C3.

[HARD] **Do not hand-edit C3.** A hand edit is silently overwritten at the next emission. If something needs fixing, fix C2 and regenerate.

Where a miss gets caught — three checkpoints each look at a different axis:

| Check | What it examines | When |
|---|---|---|
| `make build` (preceded by `agents-emit-check`) | Source layer: committed `.toml` vs the output emitted from `.md` | Every local build. **Read-only — does not regenerate** |
| `go test ./internal/template/agentemit/...` | The same axis | Every CI run |
| `make embed-check` / `moai doctor --check "Agent Emit Embed"` | **Embed axis**: the bytes actually embedded in an already-built binary vs. the committed copy | Manual. `BIN=<path>` can also target the installed binary |

The reason the third checkpoint exists separately: `go test` recompiles the test binary every time, and `//go:embed` reads the committed copy as of that moment, so a source-to-source comparison **cannot, in principle, see a stale binary**. That is why `make embed-check` does not have `build` as a prerequisite — a freshly built binary matches by definition, so a check that runs only right after a build is a tautology. For the same reason, it is not attached to the CI build job either (CI builds from the commit it checks).

Regeneration happens **only through the explicit verb `make agents-emit`**. If the build silently overwrote, the very fact that a hand edit existed would disappear.

### Template content neutrality — acceptable content range for templates

**§2.1 Template Content Neutrality — Acceptable Content Range for Templates**: When editing template source files in `internal/template/templates/`, ensure content adheres to the **acceptable** kept-classes (C1/C2/C4/C5/C6/C8) and excludes the FORBIDDEN content classes (SPEC IDs, REQ tokens, Audit citations, internal dates, commit SHAs, macOS-bias paths, CLAUDE.local references), enforced by CI guard (`.github/workflows/template-neutrality-check.yaml` trigger on path change). The canonical C1-C8 acceptable-vs-forbidden content-class catalogue lives in `.moai/docs/template-internal-isolation-doctrine.md §25.1` (cross-referenced by **§25 (Template Internal-Content Isolation)** of `AGENTS.local.md`, now a stub). This keeps the template neutral across the 16 supported **programming languages** and free of moai-adk internal development state. Distinct axis: user-facing locales (ko/en/ja/zh) are governed by the §8 Localization Contract in the active output style.

**Pre-PR Verification (template contributor-checklist)** — before opening a PR that touches `internal/template/templates/**`, run the canonical 5-item pre-commit self-check (the CI guard `template-neutrality-check.yaml` is the safety net). See `.moai/docs/template-internal-isolation-doctrine.md` §25.3 for the full 5-item checklist and §25.1 for the forbidden/allowed content-class catalogue (C1-C8). (C3 dates + C7 commit-hashes are owned by the sibling `internal_content_leak_test.go` per §25, not this neutrality checklist.)

## §D References index (former §18-27)

Sections §18-27 were consolidated into external `.moai/docs/` files to reduce launch-time context. Each entry below is the authoritative location for its domain.

- **§5 Version Management** (SemVer pre-release, ldflags injection, release process): `.moai/docs/version-management.md`
- **§7 Hook Development** (shell-script-only pattern, settings.json format, quoting rules): `.moai/docs/hook-development.md`
- **§18 Git Workflow** (the Enhanced GitHub Flow body, plus the parent-model note for the [2026-08-27] git-flow transition — the canonical source is now `AGENTS.local.md` §4.0 (per-card draft-PR policy, 2026-10-19); this document still describes develop — drift; branch protection `enforce_admins: true`; Hybrid Trunk RETIRED): `.moai/docs/git-workflow-doctrine.md`
- **§19 AskUserQuestion Enforcement + §19.1 Implementation Kickoff Approval** (REQ-ATR-015, revised to the default autonomous transition): canonical SSOT at `.claude/rules/moai/core/askuser-protocol.md` + `.claude/rules/moai/workflow/auto-semantics.md` §9.1 (the default form of plan→run entry is audit-cross evidence — plan-auditor PASS + audit-ready + hash integrity + no blocking issues — plus a decision record; the three keep-set categories [environment-impossible / operator-held / irreversible external-shared] keep the operator's direct answer)
- **§20 Vercel Build Cost Guard** [HARD]: all Vercel projects MUST use Elastic build machine ($0.0035/CPU min vs Turbo $0.126/min); check Build Machine setting first on cost anomalies
- **§21 Dev-Only Commands Isolation** (split harnesses, `SPLIT_HARNESS_NAMESPACE_LEAK` sentinel): `.moai/docs/dev-only-commands-isolation.md`
- **§22 Dev Settings Intent** (settings.json key semantics): `.moai/docs/local-dev-settings-intent.md`
- **§23 Local Git Workflows** (PR-mandatory, 1-person OSS, all tiers via PR — the release-PR path; [2026-08-27] transition to git-flow `develop`, then [2026-10-05] GitHub Flow transition returning to a `main` base — the canonical source is now `AGENTS.local.md` §4.0; this document still describes develop · drift): `.moai/docs/git-local-workflow-doctrine.md`
- **§24 Harness Namespace** (template-managed vs user-owned separation): `.moai/docs/harness-namespace-doctrine.md`
- **§25 Template Internal-Content Isolation** (neutrality catalogue, CI guard): `.moai/docs/template-internal-isolation-doctrine.md`
- **§26 Linear integration** (local-only): `.moai/docs/local-linear-integration.md`
- **§27 Agent-Skill Architecture** [HARD]: every agent gets ≥1 skill set (4 elements: workflow skill + knowhow reference + scripts + trigger); all skill bodies in English; `/moai:<sub>` slash-wrapping maintained; `/moai:harness` meta-harness (v4 Builder); Analyze-First execution plan before any work
