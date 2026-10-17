---
description: "git-flow lane protocol (repo-local) — card worktrees branch from local main, lanes land through the integration verb on local main, origin/main is the CI verdict surface, batch releases reach origin/main only through a release PR"
paths: ".moai/specs/**,.claude/skills/moai/workflows/run.md,.claude/skills/moai/workflows/sync.md,.claude/rules/local/repo-local-pr-policy.md,AGENTS.local.md,internal/cli/**,internal/hook/**"
---

# git-flow Lane Protocol (moai-adk-go, local-only)

> **2026-08-27, operator instruction.** This repository switched from GitHub Flow to git-flow. It deliberately reverses the "local-only, disposable `develop`, no remote push" decision that backlog card t281 set on 2026-08-26. The model and its rationale are in `AGENTS.local.md` §4.1; this file is the **operating rule** that lanes and the factory leader follow as written.
>
> This is a local-only file (siblings: `ci-watch-protocol.md`, `repo-local-pr-policy.md`, `lifecycle-sync-gate.md`). Do not mirror it into `internal/template/templates/`.

---

## 1. Branch point — card worktrees branch from local `main`

[HARD] Create card worktrees **from local `main`**, not from `origin/main` (§4.0 §1).

[HARD] A card worktree is created only through `moai factory next` or `moai worktree new`. When the configured base (`git_strategy.worktree_base_branch`, `main` here) resolves, it branches from the named ref `main` and not from HEAD; when the base is empty or does not resolve, creation falls back to the invoking tree's HEAD (`internal/cli/session_worktree.go:236-237`). Claude Code's default `--worktree` (which branches from the remote default branch) is not used for card work. After creation, confirm that the new tree's HEAD equals `main`; that confirmation, not the creation path, is the check. A lease-time warning for the fallback is planned in t1847; it does not exist today.

[HARD] **Local `main` is both the card base and the landing surface** (§4.0 §1·§2). `origin/main` advances only through release batches, so it is normal for local `main` to be ahead. A `↓N` on the status line means a batch has reached `origin/main` but has not been resynchronized yet; in that case the resynchronization of §4.0 §4 applies.

[HARD] **Committing on local `main` is blocked by configuration (SPEC-MAIN-COMMIT-BAN-001, card t1337).** `workflow.branch_guard.deny_commits_on: [main]` rejects `git commit` / `git revert` / `git cherry-pick` in the primary checkout. Under the SPEC-LOCAL-MAIN-FLOW-001 Q1 decision (keep), that value is unchanged. Create card commits only in card worktrees branched from local `main`; only merges go through the tool path of §4.0 §2. The procedure for disposing of leftovers is historical, recorded in `.moai/docs/gitflow-integration-chain.md`.

- Enter through the path that fits your harness:
  - Claude Code lane: `moai cc -w <card-id>`, or `EnterWorktree(<card-id>)` in the current session.
  - Codex lane: when entering an existing tree with a new session, use `moai codex -w <card-id>`.
  - Codex `-l` lane: the supervising launcher picks the card worktree and starts the child session with `codex -C <worktree absolute path>`. The child session reads that tree's `AGENTS.local.md`, works there, and does not call `moai cc -w`. A session started by the `moai codex` launcher loads this file into `developer_instructions`.
  - If the tree does not exist, create it first with `moai worktree new <card-id>`. **Bare `git worktree add` is prohibited.** Git would know the tree while MoAI would not, so `done`/`clean`/`recover` would have no target to close.
- Immediately after creation, confirm that the card tree's `HEAD` and local `main` point to the same commit. If they differ, do not start work; correct the branch base first.
- Immediately after creation, rename the branch in place: `git branch -m WT-<slug>`. Take the slug from **what the card does** (lowercase `a-z0-9-`, at most 3 tokens, at most 24 characters), and **do not put the card id in it**. The worktree directory keeps the card id (`.claude/worktrees/<card-id>`). The derivation is now automated and honors the same shape (metadata prefixes skipped, label tokens dropped, SPEC-id fallback); branches created before it carry legacy numeric slugs (`WT-3-2-2` shape) — rename them in place the same way inside the card tree: renaming the checked-out branch is safe (the tree, its lock, and the session anchoring are unaffected) and card tooling keys on the worktree path, not the branch name (`resolveCardBranch` resolves through `card.WorktreePath`), so the card record and evidence paths carry unchanged.
- A new card gets a new worktree. If the session is anchored to the previous card's tree, return to the primary checkout with `ExitWorktree` before creating it; otherwise the new card's work stacks on top of the old card's branch.
- **The three traceability carriers stay in force**: the dispatch's `card:` field, the card id in **every** commit message on the branch, and the evidence path (`.moai/reports/<card-id>/…`). Branch names no longer identify the card, so do not omit any of the three.
- [HARD] A lane session starts inside its card worktree and stays there. Between cards a Claude Code lane moves only by the in-session transition (`ExitWorktree(keep)`, `moai factory next` from the primary checkout, `EnterWorktree` of the leased tree, verify the toplevel); no operator clear is involved. The canonical source for this transition is the Lane card transition and Isolation sections of the distributed rule `.claude/rules/moai/workflow/factory-dispatch.md`, and it is not restated here. Returning to the primary checkout (`ExitWorktree`, §6) is outside the scope of this prohibition.

## 2. Integration surface — the dedicated integration worktree (`main`)

[HARD] `main` is checked out in exactly one dedicated integration worktree, at a fixed path (currently `.claude/worktrees/main`). That worktree performs landing and resync only, plus the operator-requested rc build of §9: it is never used to edit or experiment, and it stays clean. The primary checkout is the human and leader workspace and does not check out `main`. That worktree is the landing surface (§4.0 §2). A lane does not check out `main` in its own card tree (git checks out a given branch in only one worktree).

[HARD] Read `main` source through `git show main:<path>` or inside the integration worktree, never from the primary checkout's working tree, which is not `main`. `git show` reads the committed blob only: it does not reproduce a smudge or clean filter or end-of-line conversion that a working-tree copy carries, it shows no uncommitted change, and it says nothing about a later move of `main`. Keep the primary checkout clean: diagnose each untracked or modified file there one by one (commit it in the card worktree that owns it, move it to that worktree, or cover it with an ignore rule) rather than leaving it in place. Today no moai verb moves the primary checkout to the current local `main` tip; a primary-follow step is planned in t1847.

[HARD] **The canonical merge procedure is not here.** The canonical source for merging into local `main` is `AGENTS.local.md` §4.0 §2 (the tool path). The `WT-*` path in Step 3.2 of the distributed skill `.claude/skills/moai/workflows/sync/delivery.md` does not apply to this repository wherever it conflicts with that canonical source. The procedure is not restated here, because two copies diverge as soon as they exist.

[HARD] **Repo-local exception — Step 6 of delivery.md Step 3.2 (push of the integration branch) is EXCLUDED in this repository (operator instruction, 2026-09-02).** Even if the distributed skill's WT-* procedure describes the push step, lanes in this repository do not perform that step; batch releases are a single bulk action by the leader (§4, §7). The distributed skill file is left untouched, and for lanes this repository's local answer takes precedence.

Repo-specific matters:

- The tool does not provision the integration worktree: `moai integration merge` refuses when no tree holds `main`, and the leader provisions the tree (`internal/cli/integration_merge.go:145-157`). There is no provisioning procedure here. The statusline landing-surface display and a configuration key that designates the integration worktree path are planned in t1847 and do not exist today. `.claude/worktrees/develop` and `moai cc -w develop --branch develop` (2026-09-02) from the develop era are history.
- The integration branch is `main`. Batches land on `origin/main` via `release/main-batch-YYYYMMDD` (§4.0 §6), and the CI of origin/main is the verdict surface (§4).
- After the merge, if work remains on your own card, re-enter with `EnterWorktree(<card-id>)`. `ExitWorktree` returns you to the primary checkout, not to your own tree.

[HARD] Merging from within your own worktree with `git -C <primary path> merge …` **is refused by the worktree-session guard**. Merges into local `main` go only through the tool path (`moai integration merge --card`) (§4.0 §2).

[HARD] **Candidate-path exception — operator approval 2026-10-03 (REQ-CCI-016).** `moai integration candidate --card <card-id>` is the path that pushes a candidate commit to `ci/<card-id>`. It neither takes nor touches the integration window (`moai integration acquire`/`release`). Merges and integration-branch pushes remain a single bulk action by the leader, as in §4.

## 3. Serialization — one lane at a time in the merge window

[HARD] Use the existing mechanism as it is. Do not build a new one.

> **[HARD] Whether the window is alive is tied to the session process — card t298.** The pid that `acquire` records belongs not to the short-lived CLI process that ran the command, but to **the session that ran it**. So the window keeps reading as held even after `acquire` returns, and there are only three ways it is released: the holder session dies, the holder itself runs `release`, or another lane takes it with a recorded `--force`.
> A window taken without identifying its owner is recorded with pid 0 and reads as **alive**. When uncertain, emptying the window would lead to two lanes merging together, so the judgment always leans toward "alive".
> **A window taken before the fix still reads as available for takeover** — old records carry no session anchor. A lane that held the window at upgrade time re-acquires it by running `moai integration acquire --name <lane> --card <card-id>` once more.
> This record does not mechanically separate the lanes. `acquire` itself does not make its read-modify-write sequence atomic, so if two lanes try to acquire at the same moment, both may believe they hold the window. This is a coordination signal, not a permission boundary; this record is the mechanical layer. Under the open policy, a lane promoted to the front of the queue and holding the window merges without a leader nomination.

```bash
moai integration acquire --name <lane> --card <card-id>   # before entering the integration worktree
moai integration status                  # who holds it
moai integration release                 # after sending the completion report
```

Taking the window from a live holder requires `--force`, and what was displaced is recorded. It must be deliberate and must not be silent.

[HARD] **`MERGE_HEAD` being empty is a necessary condition, never a sufficient one.** A state where `git rev-parse -q --verify MERGE_HEAD` prints nothing looks exactly the same while another lane is resolving a conflict — between `git merge --abort` and a retry, or at a point where nothing has been staged yet. Reading this silence as "the tree is empty" is precisely the path on which two lanes overlap. The overlap stays invisible until one side commits. The probe is the **last** check, not the first.

[HARD] **Re-read `HEAD` immediately before the merge commit and immediately before any push.** Do not use a value read earlier in the turn, or the branch reported at session start.

```bash
git rev-parse --short HEAD
git branch --show-current
```

If either differs from what you assumed, do not proceed; report the divergence.

## 4. Push — the leader does it in bulk; lanes never push

**Lanes do not push `main`.** This rule is the current form of the 2026-09-02 operator instruction (at that time it covered develop). A lane's responsibility ends at the local merge: it passes the merge commit SHA (including the card id) to the leader in its completion report, and nothing further. The leader performs batch releases in **bulk**: it collects the merge SHAs from completion reports, closes the batch, and pushes to `release/main-batch-YYYYMMDD` by the procedure in §4.0 §6.

**Candidate-path exception — operator approval 2026-10-03 (REQ-CCI-016).** For this one path only, this replaces `WT push · CI request prohibited`. A lane may push a candidate commit made by `moai integration candidate --card <card-id>` to `ci/<card-id>` to receive a CI verdict before landing, and the landing gate requires a green verdict for that candidate. Pushing the integration branch remains, as before, a bulk action by the leader. The remaining rules of this section (landing verification, batch trigger, green condition) do not change, and this section's develop descriptions are to be read under the transition notice (drift) in AGENTS.local.md §4.1.

**The leader verifies the remote landing.** After the batch PR is merged, the leader confirms that the remote has moved, using `git fetch origin main` followed by `git rev-parse origin/main`, and only then approves worktree disposal. A card's done does not wait for the batch (§4.0 §5; implemented by t1621).

**The remote CI (`origin/main`) is the authority for the integration verdict.** A local pass is only an early signal — it is neither a clean environment nor the darwin/windows matrix.

**Batch-close trigger — the threshold is carried by configuration (SPEC-LEAD-AUTOPUSH-001).** The moment to close a batch is set by a count trigger, not by the leader's discretion: when `git rev-list --count origin/main..main` reaches `git_strategy.manual.lead_push_threshold`, the leader closes the batch and starts the procedure in §4.0 §6. The source of the value is `.moai/config/sections/git-strategy.yaml`, and this document does not restate the number; it names only the key and the count command. The trigger is evaluated only after the merge window has closed — do not push inside the window (`moai integration acquire`/`release`).

**Green condition — a batch reopens only when green.** There are two layers. (i) Per-card pre-gate: a merge that fails the merge-tree re-measurement in the integration window is not put into a batch (all three recent reds were caught by this re-measurement before the push). (ii) Per-batch hold: if the `origin/main` CI of the most recently landed batch is red, the next batch is held; the trigger becomes valid again only after the batch has been repaired and returned to green. The green condition takes precedence over the count trigger. If that CI verdict does not yet exist (for example, for the first batch), it is not red and is not a reason to hold.

**A threshold of 0 (or a missing key) means disabled, not an error.** If `lead_push_threshold` is 0, the trigger is treated as off, and the timing of batch closure reverts to the leader's judgment; do not treat it as an error condition.

## 5. Conflicts — the lane that owns the change handles them

The lane performing a merge resolves the conflicts that merge causes. A conflict that cannot be resolved (a semantic conflict with changes another lane has already merged) is a **blocker report to the leader**. It is not a forced merge.

## 6. After the merge — a lane does not pick its own card

- After completing the local merge and reporting the merge SHA to the leader, return to the primary checkout with `ExitWorktree` and **wait until the leader dispatches the next card.** A lane does not take a card from the queue. This wait is not an open-ended wait but an **explicit wait** — a wait that leaves a wait record on disk (reason, target, and re-check point). On an awaken turn, before card work resumes, the lane runs the lane-stall watchdog first: it invokes Skill("moai-lane-watchdog") and follows it. The watchdog re-measures three progress channels (HEAD SHA, evidence mtime, integration-window state), classifies the cause, and when a judgment is needed, resolves it without waiting for a reply through the decision ladder (disk evidence → decision board → audit cross-check → jev_ask → lead chat as a last resort). Do not read the absence of a reply as the absence of progress. The canonical source for this wait and ladder doctrine is `.claude/rules/moai/workflow/auto-semantics.md`. The watchdog's queue-readonly boundary holds in combination with SPEC-RELATION-PICKUP-FILTER-001, which owns the queue-level pickup filter (a lane still does not pick cards — the same boundary as the prohibition above).
- **Self-dispatch lane exception — card lease.** A lane of a self-dispatch factory run may lease the next waiting card with `moai factory next` (the only promotion a lane performs). Apart from this exception, queue changes (`add`, `drop`, `done`, `edit`, etc.) and `moai contract sign` are prohibited for lanes.
- **Self-dispatch lane exception — merge window.** A Claude self-dispatch lane does not request the window from the leader. Through the integration procedure of `moai factory complete`, which calls `moai integration merge --card`, it takes the integration window itself and merges its own card into local `main` (this lane replaces "ask the leader to merge" in the first item above). A lane never runs a hand-made git merge into main; it merges only with moai integration merge --card, or with moai factory complete, which calls it. Codex lanes are not covered by this exception: they stop at merge-ready (REQ-SD-025). Neither exception changes the prohibition above (other queue changes plus `moai contract sign`).
- [HARD] **A card worktree is disposed of only after the batch carrying its merge has landed on `origin/main` (§4.0 §5).** Until then, that tree is the only copy of the work. The remote landing is produced by the leader's batch release (§4, §7). An L1 tree (`.claude/worktrees/…`) is not a target of `moai worktree done`; close it through the keep/remove prompt at session end, or with `git worktree unlock` followed by `git worktree remove`.

## 7. Leader — judge by reading the evidence

- [HARD] The leader advances a card by **reading the lane's evidence, not its reply**. A reply is a claim, not an observation, and routing is not guaranteed either.
- The merge verdict (what entered local `main`) belongs to the leader. Do not let the lane that produced the work judge its own result.
- If the evidence file is absent, unreadable, or stale, it is a **gap** — leave the card where it is and report the reason.
- Once a merge is confirmed, dispatch the next card to **a lane that is free right now**.
- **Batch release is the leader's bulk responsibility.** Collect the card ids and local merge SHAs from the lanes' completion reports → the timing of batch closure is set by the §4 count trigger and the green condition (not by the leader's discretion alone) → push to `release/main-batch-YYYYMMDD` and merge the PR by the procedure in §4.0 §6 → verify the remote landing with `git fetch` and `git rev-parse origin/main` → only then approve worktree disposal.

## 8. Verification is lane-local

**[SUPERSEDED by §4.0]** [HARD] Run only the tests your own change can affect, and after the push let the `origin/develop` CI run the full suite.

- **Do not run `go test ./...` locally.** There was an incident in which several lanes ran it at once, pushing load up to 413 and paralyzing the machine (2026-08-15).
- **Do not create background load.** If a verification needs contention, the load must come with a cleanup guarantee: either a kill registered in the test framework's cleanup hook, or a `timeout` wrapper that bounds the process from outside. A `kill` appended afterward is not cleanup (it is a line that cannot be reached).
- **Where heavy runs may overlap, take a resource lease first.** `moai slot acquire --resource <name> --max-duration <limit>` → run → `moai slot release --resource <name>`. Read the holder with `moai slot status`. The integration window (`moai integration`) and its records, locks, and config keys are **kept separate** — so that a merge wait and a heavy-run wait do not block each other. The limit is the holder's own declaration, so if the holder forgets to release, the lease is bound only up to that limit (after which another lane takes it over without `--force`). The PreToolUse guard that refuses matching commands is **optional and off by default** (`workflow.slot_lease.enabled`); its three behaviors run regardless of that value. For the full surface, see `.claude/rules/moai/workflow/resource-slot-lease.md`.
  - Refusals come as exit codes: while a live holder is within its limit, another session's acquisition is silently refused with **exit 3**, and no one is replaced (observed 2026-09-14, `t774-repro-demo`). Instead of asking the leader, check who holds the current turn with a single `moai slot status --resource <name>`.
  - Examples of "heavy runs" — at a minimum, the full `internal/cli` suite (2026-09-10 intrusion: three lanes running it at once flipped verdicts at load 8–21; this is the control record showing that intrusion happens in practice when no surface exists, and it is not re-staged by re-running the suites concurrently), plus minute-scale package suites such as `internal/factory` and `internal/hook`.
  - The reason this discipline must live on disk rather than in the leader's memory is precisely those two intrusions: the first was a missing notice; the second was that even when the notice arrived, the lane had no way to look up the turn. Declaring, notifying, and having a queryable procedure are each different things.
- **[SUPERSEDED by §4.0]** **[HARD] "What did this card change?" is measured from the merge-base with the absorbed ref, not from a literal base SHA.** The left end of a scope judgment such as "no Go changes", "no template changes", or "this path only" is recomputed at read time with `CARD_BASE=$(git merge-base develop HEAD)`, and the value is not pinned. The control is `git diff --name-only "$CARD_BASE"..HEAD | wc -l` (must be 1 or more); the probe is the same range with a pathspec attached. If the control is 0, report "not measurable", not "no changes".
  - **[SUPERSEDED by §4.0]** Reason: at the moment of absorption, commits from other cards enter the literal pinned range. If local develop is ahead of the remote, the merge-base against `origin/develop` also stays at the pre-absorption fork point and produces the same false positive. Measured (2026-09-10, `.moai/reports/t543/verdict.md`): after absorbing local develop, both the literal pin and the `origin/develop` base reported 51 Go files, while only the `develop` base reported the card's own contribution (4 files, 0 Go). This reproduction is the control for this discipline.
  - **[SUPERSEDED by §4.0]** The principle is "the very ref that was absorbed." The absorption target in this repository's procedure is local `develop` (§11, `AGENTS.local.md` §4.1), so the default is `develop`. For a procedure that absorbs the remote develop, the ref becomes `origin/develop`. Even if develop moves further ahead after absorption, the merge-base stays at the last absorbed develop commit and remains correct.
  - **[SUPERSEDED by §4.0]** Limitation — **it cannot be used after the merge.** Once the card merges into develop, the merge-base becomes the card's own tip, the range empties, and the judgment passes vacuously (observed: empty output on an already-merged card branch, `.moai/reports/t543/repro/limit-a-post-merge-all.txt`). The scope judgment is for evaluation before the merge only. After the merge, the evidence is instead the identity of the merged tree and the card branch tree.
- **[HARD] "Whose is this?" is answered only by enumerating commits — never by comparing trees.** There is one canonical form: `git log <my HEAD>..<counterpart ref> -- <path>`. And **the second operand is a positive control path**: run the same form once more on **a path where another party's commits are certain to exist**, and check that it fires. If that second run is empty, the zero rows of the first are not an absence but **not measured**. This control is not kept as a discipline to be remembered separately, because when it was left that way, it did not catch anything (on 2026-09-20 three cards all mentioned the control and all were wrong — `t950` mentioned it nine times).
  - `git diff` is a tree comparison **in every spelling**, so it does not answer this question. Observed: with the same operand pair, `git diff A..B` and `git diff A B` produce **byte-identical output**, and the same `A..B` gives the opposite answer under `git log` (log: 0 rows / diff: 1 row). Only the three-dot form `git diff A...B` agrees with `log A..B`.
  - The left end is **`HEAD`** — it means "everything I already have." If you use a merge-base or a pinned SHA instead of HEAD as the left end, **commits from the earlier card come back as contact from the other side** (observed twice; in both cases `git merge-base --is-ancestor <commit> <my tip>` exit 0 = mine).
  - The question differs from the discipline just above: that one asks "What did this card change?" (its own contribution), while this one asks "Whose is this?" (attribution). Both are for evaluation before the merge only.
  - Measurement source: `.moai/reports/t1014/verdict.md` (accompanied by 13 reproductions).

## 9. rc build — on operator request, from the local main tree

```bash
# Inside the integration worktree (the tree whose HEAD is local main)
# If macOS's /usr/bin dev tools exit 69 because of the Xcode license, prefer the installed CLT.
if [ -x /Library/Developer/CommandLineTools/usr/bin/make ]; then
  PATH="/Library/Developer/CommandLineTools/usr/bin:$PATH"
  export PATH
fi
make build VERSION=vX.Y.Z-rc.N
git status --short                      # After the build, there must be no tracked-file changes (AGENTS.local.md §4.0 §3)
rm -f ~/go/bin/moai && cp bin/moai ~/go/bin/moai
sh scripts/verify-local-install.sh       # byte identity + installed-binary version SHA / exit 0
```

- The build runs in the integration worktree because that is the tree whose HEAD is local `main`; the build is expected to write ignored outputs, but it can regenerate a tracked file such as the template catalog. Run the `git status --short` check above afterwards; if a tracked file changed, stop and report it instead of committing from this tree.
- [HARD] A bare `cp` overwrite that omits `rm -f` has a precedent of producing **exit 137 (SIGKILL)** on the next invocation — the reinstall must be a clean one that replaces the inode. (`make install` is equivalent.)
- [HARD] **Bare `go install ./cmd/moai` is prohibited.** It does not carry the Makefile's `LDFLAGS`, so the version, commit, and date get baked in as compile-time defaults, and then the binary-lag verification itself becomes impossible.
- [HARD] Do not use macOS `strings` to verify the installed binary. Whether the Xcode license has been accepted would halt the verdict on the product binary. `scripts/verify-local-install.sh` uses `cmp` to confirm that `bin/moai` and the installed binary are byte-identical, and it runs the installed binary's `version` directly to observe, together, the short SHA of HEAD at measurement time and exit 0. If the default `git` is unavailable, it automatically retries with the `git` from macOS Command Line Tools; if neither works, it must receive the expected short SHA as the third argument. `make verify-local-install` is a convenience alias that calls the same script.
- [HARD] If `/usr/bin/make` or `/usr/bin/git` on macOS exits 69 because of the Xcode license, put the installed Command Line Tools path at the front of PATH using the preprocessing above. Do not run `sudo xcodebuild -license` automatically, and do not hide a verification failure with `|| true`.
- If you get 137, run `rm -f` and `cp` again.

## 10. Release — outside the lane's scope

**[SUPERSEDED — pending OQ-7]** Once the local rc test passes, branch `release/vX.Y.Z` **from `develop`**, and the `release` harness takes it from there. `origin/main` is updated only through release PRs — including batch releases (`release/main-batch-YYYYMMDD`) (§4.0 §6; branch protection `enforce_admins: true`). Lanes neither create release branches nor open PRs to main.

## 11. Updating main — resynchronizing local main after a batch lands (SPEC-RC-TESTBED-001)

When a release batch lands on `origin/main`, local `main` can fall behind (§4.0 §4). This section sets the **criterion** for deciding when to update and the **path** for updating. The rc.N numbering policy (which candidate this is) is owned by the Local RC Numbering section of `.moai/docs/version-management.md`.

[HARD] **The criterion is a ref comparison with `origin/main`, and it assumes nothing about the shape of the merges.**

```bash
git fetch origin main
git rev-list --count --left-right origin/main...main
# "0 0" = identical. "0 M" = landed commits only in local (before the batch) → not an update target
# "N 0" (N≥1) = commits only on the origin/main side → fast-forward target (§4.0 §4)
# "N M" (both ≥1) = diverged → the tool refuses (operator decision Q2 (i))
```

- `git branch --merged` is **not a sufficient criterion.** `--merged` reports only whether a branch head is reachable, and changes brought in by a squash merge leave no head in the commit graph, so it misses them — the reachability-blindness case is `SPEC-WORKTREE-SQUASH-MERGE-001`. Landing on local main in this repository is specified as `--no-ff` (the tool path of §4.0 §2), but for that reason this lane also does not assert *"--merged is empty"* as an established fact — because the judgment must be independent of the merge shape, use the rev-list comparison above.
- The criterion is always `origin/main`. Neither the local `main` being updated nor the branch state of a card worktree is the criterion.

[HARD] **The update path is one tool path only — a fast-forward-only merge inside the integration window (§4.0 §4).**

- Do not do this from the operator's terminal. The tool takes the integration window, confirms that HEAD points at `main` and that the tree being updated is clean, and then performs a fast-forward-only update (§4.0 §4). Today `moai integration resync` applies to the primary checkout only and refuses when another tree holds `main` (`internal/cli/local_main_resync.go:361`); running it in the dedicated integration worktree is planned in t1847, and this file gives no hand-run `git` substitute.
- Do not update with `git branch` / `git checkout` in the primary checkout. The BranchGuard (`workflow.yaml` `branch_guard.enabled`, local opt-in) blocks those patterns, and the two exemption paths (direct `claude --agent` execution and `MOAI_BRANCH_GUARD_EXEMPT=1`) are **unreachable from subagents spawned through the tool.** Do not recommend the operator-terminal sentinel path for subagents — a procedure that relies on an unreachable exemption is not a procedure that can be depended on.
- Operating with `git -C <primary path> …` from within your own worktree is refused by the worktree-session guard, for the same reason as §2. **The tool path is the only authorized path.**
- Coordinate the update window using the §3 mechanism as it is (so the merge window and the update do not overlap). The rc build and clean reinstall after the update are owned by the §9 runbook — not restated here.

---

## Cross-references

- `.claude/skills/moai/workflows/sync/delivery.md` Step 3.2 — canonical source of the `WT-*` integration procedure (owned by the distributed skill; this document covers only repo-specific matters)
- `AGENTS.local.md` §4.1 — the rationale for this model and its transition record (reversal of t281; the current status of the rejection reasons of 2026-08-14)
- `.claude/rules/local/repo-local-pr-policy.md` — the `main` branch protection and the mandatory-PR rule
- `.moai/docs/git-workflow-doctrine.md` (§18) / `.moai/docs/git-local-workflow-doctrine.md` (§23) — parent doctrines (some parts superseded; see the notes at the top)
- **[SUPERSEDED by §4.0]** `.moai/config/sections/git-strategy.yaml` — `git_strategy.manual.workflow: git-flow` plus the develop/release keys (must be re-applied after `moai update`; `AGENTS.local.md` §2.3)

---

Classification: Local-only operational rule (dev-only, never mirrored to the distributed template).
