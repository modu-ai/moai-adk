# moai-adk-go Local Development Guide

> **Purpose**: Essential guide for local moai-adk-go development
> **Audience**: GOOS (local developer only)
> **Last Updated**: 2026-09-29

---

## Language rule for new artifacts
- New memory topic files and new backlog card text are written in English.
- User-facing chat and reports use the conversation language (conversation_language, from .moai/config/sections/language.yaml).
- Code, identifiers, commands, paths, and literal tokens keep their original form.

---

## 0. [HARD] Which copy of this file is canonical

Read this section before the rest. On 2026-09-07, an incorrect copy of this file was distributed to 12 lanes; the discriminator that prevents a repeat of that incident is here.

### 0.1 [HARD] The canonical copy is the copy in the tree that lanes branch from

**The discriminator is "the tree lanes branch from governs," and that tree is currently `main` (GitHub Flow transition on 2026-10-05 — §4.1).** Because card worktrees branch from `main`, the committed copy on `main` is the document lanes actually read, and that is the canonical copy.

**Do not use the date or "whichever side was delivered later" as the discriminator.** On 2026-09-07, the text delivered later was the wrong one; a recency rule would not have prevented that incident. The branch-point tree gives the same answer even if the push model changes.

### 0.2 [HARD] An uncommitted working copy cannot be cited as canonical

**An uncommitted working copy that has never been committed on any branch is not canonical, and it cannot be cited as canonical.** This is not a rule limited to this file; it is a general rule for citation — text that is absent from history gives no way to confirm that someone else read the same thing, and neither its author nor its point in time can be recovered.

The culprit on 2026-09-07 was not a stale branch but **a working copy in the primary checkout that git tracks but that was never committed on any branch**. Before citing, confirm which commit contains the text — if it cannot be read with `git show <ref>:<path>`, it is not canonical.

### 0.3 [HARD] The committed copy on `main` is a discarded third model

**The old `CLAUDE.local.md` committed to `main` before the transition is a retired model and is not a citation target.** That version describes a regime in which "`develop` is not pushed to the remote and each card gets a PR to `main`". That regime was superseded by the develop GitFlow chain of 2026-08-29, and the GitHub Flow transition of 2026-10-05 returned it to a `main` base (§4.1). The canonical copy now is the current committed copy on `main` after the transition (§0.1), not the pre-transition old commit. Taking the old version as canonical reproduces the 2026-09-07 failure.

### 0.4 [HARD] The outdated local file in primary `main` is not canonical

This migration passed through develop and reached `main` with the 2026-10-05 transition; after that, the `AGENTS.local.md` in the commit merged into `main` is canonical. The `M CLAUDE.local.md` visible in the primary checkout is an uncommitted shared working copy (§0.2). Do not cite that state as receipt evidence for a new worktree.

> **[HARD] This marker is not a cleanup target.**
>
> **Do not run `git restore CLAUDE.local.md` in the primary checkout** — that regression would revert the shared working copy to the discarded model of §0.3.
>
> When editing this file, edit it in a card worktree and merge into `main` (§4.1). The outdated file in the primary checkout has already gone through a separate transition (card t1317, 2026-09-29 — the preserved copy is in `.moai/state/retired/` in the primary checkout).

---

## 1. Quick Start

See: `.moai/docs/local-dev-guide.md` § 1 — work locations, the edit→`make build`→test→commit cycle, and the [CRITICAL] `moai` CLI vs `/moai` slash-command distinction table. The core in one line: edit templates in `internal/template/templates/` and regenerate them with `make build`; commit in a card worktree (the primary checkout is prohibited — AGENTS.md §2·§3).

---

## 2. File Synchronization

### Protected Directories (Never Modify During Template Sync)
```bash
# CRITICAL: These directories contain user data and must NEVER be deleted
.claude/        # Local Claude Code configuration
.moai/project/  # Project documentation (product.md, structure.md, tech.md)
.moai/specs/    # SPEC documents (active development files)
```

### Template Source (Single Source of Truth)
```bash
# All template changes MUST be made here
internal/template/templates/.claude/
internal/template/templates/.moai/
# (The .agency/ template root has been removed — legacy archive; [2026-08-27 audit correction])
internal/template/templates/CLAUDE.md
```

### [HARD] Template-First Rule

When adding new files to `.claude/`, `.moai/`, or `.agency/`:

1. **Add to template FIRST**: `internal/template/templates/<path>`
2. **Run `make build`** to regenerate embedded files
3. **Then sync to local**: `moai update` or manual copy

Never add files directly to the local project directories without also adding them to the template source. This includes:
- New agents (`.claude/agents/`)
- New skills (`.claude/skills/`)
- New commands (`.claude/commands/`)
- New rules (`.claude/rules/`)
- New config files (`.moai/config/`)
- New agency files (`.agency/`)

**Verification**: Before committing, check that every new file under `.claude/`, `.moai/`, or `.agency/` has a corresponding file in `internal/template/templates/`.

### [HARD] §2.0 There are three copies of the agent definitions, and only the agent source is edited by hand (C1 of the eight agents not yet generated is the exception until piece 2)

Agent definitions exist in three copies, rendered from one edit surface: the agent source layer `internal/template/agentemit/source/*.md`. **The agent source is the sole edit surface**; none of the three copies is a hand-edit target for the generated agents.

| Copy | Path | Nature |
|---|---|---|
| C1 | `.claude/agents/moai/*.md` | Local dogfooding; rendered from the agent source and byte-identical to C2 for the generated agents (builder-harness, e2e-tester, manager-design, super-advisor); hand-edited for the remaining agents until piece 2 absorbs or retires them |
| Source | `internal/template/agentemit/source/*.md` | **The only hand-edited layer for the generated agents** |
| C2 | `internal/template/templates/.claude/agents/moai/*.md` | Distribution mirror, **rendered from the agent source** |
| C3 | `internal/template/templates/.codex/agents/moai/*.toml` | **Machine-rendered from the agent source** (`internal/template/agentemit`) |

[HARD] **If you have edited the agent source `internal/template/agentemit/source/*.md`, run `make agents-emit`.** Hand-editing C2 turns `make build` red through `agents-emit-check`. Otherwise `internal/template/templates/.codex/agents/moai/*.toml` stays stale, and a binary built in that state runs with the old definitions embedded. C1 is byte-identical to C2 for the generated agents (builder-harness, e2e-tester, manager-design, super-advisor); the remaining agents join when piece 2 absorbs or retires them. The generation relationship runs in one direction only, source → C1, C2, C3 for the generated agents; C1 of the remaining agents is not rendered until piece 2.

[HARD] **Do not hand-edit C3.** A hand edit is silently overwritten at the next emission. If something needs fixing, fix the agent source and regenerate.

Where a miss gets caught — three checkpoints each look at a different axis:

| Check | What it examines | When |
|---|---|---|
| `make build` (preceded by `agents-emit-check`) | Source layer: committed `.toml` vs the output rendered from the agent source | Every local build. **Read-only — does not regenerate** |
| `go test ./internal/template/agentemit/...` | The same axis | Every CI run |
| `make embed-check` / `moai doctor --check "Agent Emit Embed"` | **Embed axis**: the bytes actually embedded in an already-built binary vs. the committed copy | Manual. `BIN=<path>` can also target the installed binary |

The reason the third checkpoint exists separately: `go test` recompiles the test binary every time, and `//go:embed` reads the committed copy as of that moment, so a source-to-source comparison **cannot, in principle, see a stale binary**. That is why `make embed-check` does not have `build` as a prerequisite — a freshly built binary matches by definition, so a check that runs only right after a build is a tautology. For the same reason, it is not attached to the CI build job either (CI builds from the commit it checks).

Regeneration happens **only through the explicit verb `make agents-emit`**. If the build silently overwrote, the very fact that a hand edit existed would disappear.

**§2.1 Template Content Neutrality — Acceptable Content Range for Templates**: When editing template source files in `internal/template/templates/`, ensure content adheres to the **acceptable** kept-classes (C1/C2/C4/C5/C6/C8) and excludes the FORBIDDEN content classes (SPEC IDs, REQ tokens, Audit citations, internal dates, commit SHAs, macOS-bias paths, CLAUDE.local references), enforced by CI guard (`.github/workflows/template-neutrality-check.yaml` trigger on path change). The canonical C1-C8 acceptable-vs-forbidden content-class catalogue lives in `.moai/docs/template-internal-isolation-doctrine.md §25.1` (cross-referenced by **§25 (Template Internal-Content Isolation)** of this file, now a stub). This keeps the template neutral across the 16 supported **programming languages** (§15) and free of moai-adk internal development state. Distinct axis: user-facing locales (ko/en/ja/zh) are governed by the §8 Localization Contract in the active output style — see §15's disambiguation note.

**Pre-PR Verification (template contributor-checklist)** — before opening a PR that touches `internal/template/templates/**`, run the canonical 5-item pre-commit self-check (the CI guard `template-neutrality-check.yaml` is the safety net). See `.moai/docs/template-internal-isolation-doctrine.md` §25.3 for the full 5-item checklist and §25.1 for the forbidden/allowed content-class catalogue (C1-C8). (C3 dates + C7 commit-hashes are owned by the sibling `internal_content_leak_test.go` per §25, not this neutrality checklist.)

### Local-Only Files (Never in Templates)

The full registry (30+ entries, including paths, reasons, and history): `.moai/docs/local-dev-guide.md` § 2a. The essentials: `settings.local.json` (runtime-managed) · `.claude/rules/local/*` (dev-only rules) · `scripts/{ci-watch,ci-autofix,jev,ac-baseline}/` · `.moai/{cache,logs,state,specs,plans,reports,docs}` contents · every dev-only document under `.moai/docs/*.md` — **all of these must have no template mirror, and they are kept either outside the managed roots of §2.3 or in locations that are not deleted even inside those roots.**

### [HARD] §2.3 Survival rules of moai update — code-enforced, plus residual cautions (card t750 → revised by card t1547)

**Summary — the full text has been moved to `.moai/docs/update-local-file-survival.md` (card t750; revised to code enforcement by card t1547 on 2026-10-07).** An update of an existing project no longer deletes the managed root wholesale. After classification (four classes: template-owned / user-modified / user-owned / stale), files are refreshed, merged, preserved, or archived and then deleted; local-only files and operator-set `git-strategy.yaml` values (such as `worktree_base_branch: main`) survive by code (SPEC-UPDATE-MIGRATION-001). The `Updated N files` summary also counts the managed files and lists every deletion. **Even so, [HARD] rules remain:** ① a new local-only file still defaults to being **outside** the managed root (preservation is a guarantee, not an exemption — the placement table by purpose is in the migration document); ② the post-update check (`git status --porcelain | grep '^ D'` — 0 is the expected result) stays in place; ③ only the **legacy fresh-install path** (config-incompatible) still retains wholesale removal, and even then the guard skips user-owned files and unresolved user-modified files; ④ conflicts are reported as `<path>.moai-new` sidecar files with no automatic resolution — when they pile up, the user cleans them up; ⑤ the `worktree_base_branch` value keeps the base `main` after the transition (if this key is left as develop, `moai worktree new` cuts card trees from a stale develop; 6 reverse-direction incidents on 2026-09-24, card t1159) — re-applying it each time is no longer necessary, but the confirmation grep is still useful. Commands, measured evidence, and the class table are in the migration document.

### [HARD] settings.local.json Separation

`settings.local.json` is **runtime-managed**. Never put it in templates.

- Modified by `moai glm`, `moai cc`, `moai cg` commands at runtime
- Modified by SessionStart hook (GLM credentials, teammateMode, CLAUDE_ENV_FILE)
- Contains per-machine values: tmux pane IDs, API tokens, absolute paths
- **Never** add effortLevel, teammateMode, or env tokens to the template

If you accidentally commit `settings.local.json`, run `git rm --cached .claude/settings.local.json`.

### [WARN] OpenTelemetry / OTEL in Tests

Do NOT use `t.Setenv` with OTEL environment variables (`OTEL_EXPORTER_*`, `OTEL_SERVICE_NAME`) in tests. Setting these in parallel tests causes data races because the OTEL SDK initializes global state from env vars on first use.

- Use a fake/no-op exporter instead of env-var configuration in tests
- If the test must set OTEL vars, make the parent test non-parallel and use `t.Setenv` only in non-parallel subtests

### Embedded Template System · Command-to-Skill Publication

The full text has been moved to `.moai/docs/local-dev-guide.md` § 2b·2c (card t1303). Summary: `internal/template/embed.go` compiles `//go:embed all:templates` + `catalog.yaml` into the binary, so `make build` is required after editing a template; the codex skill emission of the 16 `/moai` commands is regenerated only through the explicit verb `make commands-emit`, and the emitted copies (`.agents/skills/moai-*/SKILL.md`) are not hand-edited.

---

## 3. Code Standards

The language policy is defined in `.claude/rules/moai/development/coding-standards.md` (auto-loaded).

### Go-Specific (this project only)

- File naming: `snake_case.go`, `snake_case_test.go`
- Error wrapping: `fmt.Errorf("operation: %w", err)` (string concatenation prohibited)
- All code, comments, godoc in English

---

## 4. Git Workflow

### Before Commit
- [ ] Code in English
- [ ] Tests pass for the packages affected by the change (`go test -timeout 30m ./internal/<pkg>/...`; `-timeout 30m` = the D2 derivation — 1.61x headroom over the measured worst case of 1118.093s for a single local package; baseline `.moai/reports/t1253/measure-meta.txt`, SPEC-CLI-TEST-TIMEOUT-001) — **do not run the full suite (`go test ./...`) locally**. An incident occurred in which several lanes ran it at once, pushing load up to 413 and paralyzing other workspaces (2026-08-15). Verdicts for all packages belong to CI, which runs the PR head in a clean environment and is therefore stronger evidence. The exception is the integration verification of §4.1 — then, **one at a time, serially**
- [ ] Linting passing (`golangci-lint run`)
- [ ] Templates regenerated (`make build`)

### Before Push
- [ ] Branch rebased
- [ ] Commits organized
- [ ] Commit messages follow format (Conventional Commits)

### Commit Message Format

Conventional Commits — `<type>(<scope>): <description>` + an optional body/footer. Types: feat, fix, docs, style, refactor, perf, test, chore, revert. For examples and the full format, see `.moai/docs/local-dev-guide.md` § 4a.

### §4.0 Default development flow — agentic-trunk (local-main integration)

The flow is **agentic-trunk: lane worktree, serial landing into main, leader release batch per lead_push_threshold, tagged release.** Per-card delivery follows the target-holder routing: where only the primary checkout holds `main` (this repository — the 2026-10-21 rescope of SPEC-GIT-STRATEGY-AGENTIC-TRUNK-001 abolished the dedicated integration worktree), the completed card lands through its pull request, which the leader merges with no auto-merge request; where a dedicated integration worktree holds `main`, the card lands through the integration-window local merge of item 2. The landing surface — whichever tree holds `main` — performs landing and resync only, plus the operator-requested rc build (lane protocol §9): it is never used for card work, and it stays clean. Clauses marked in §4.1 are superseded by this section; clauses without a marker remain in force as written. (The live `git_strategy.manual.workflow` value converts at the leader's post-landing cutover; until then the temporary bridge `github-flow` is in force, and this section describes the post-cutover flow.)

**[HARD] 1. Branch base**

Card worktrees branch from local `main`. Because origin/main can fall behind local `main`, it is not used as the branch base. In this repository, this section replaces the template sentence "Create the fresh tree from the remote default branch." The distributed template does not change. A card worktree is created only through `moai factory next` or `moai worktree new`. When the configured base (`git_strategy.worktree_base_branch`, `main` here) resolves, it branches from the named ref `main` and not from HEAD; when the base is empty or does not resolve, creation falls back to the invoking tree's HEAD. Claude Code's default `--worktree` (remote default branch) is not used for card work. After creation, confirm that the new tree's HEAD equals `main`; that confirmation, not the creation path, is the check.

**[HARD] 2. Merge landing**

- Where a dedicated integration worktree holds `main`, the landing surface is that worktree and the tool refuses when no tree holds it. Where only the primary checkout holds `main` (this repository), the completed card takes the pull-request edge instead: the completion path (`moai factory complete` / merge-ready) opens the card's pull request, and the leader merges it — the merge request is never an auto-merge request (`--auto` was removed from the pr-open path by M5b of SPEC-GIT-STRATEGY-AGENTIC-TRUNK-001).
- Where the integration-window local merge applies, merging is performed only by `moai integration merge --card <id>`, which `moai factory complete` calls from lane sessions. Where the pull-request route applies, the lane never merges: the leader merges the card's pull request. On either route, an agent writing `git merge` directly in Bash is still refused — the designed exception covers only the tool path.
- Where the local-merge route applies, the tool (1) takes the integration window, (2) reads HEAD symbolically and confirms that it points at `main`, (3) applies the clean check of item 3, and (4) pins the tip commit of the card branch; then it refuses if any ignored or untracked file sits on the paths the merge will write or on their ancestor or descendant paths. The tool does not delete or stash files. It then (5) performs a no-fast-forward merge with autostash turned off, (6) compares the first parent, HEAD, and the status set against their pre-merge values, and (7) records the merge SHA and tree and releases the integration window (`moai integration release`).

**[HARD] 3. Clean-tree rule**

A landing tree must be clean: `git status --porcelain=v1 -z --untracked-files=all` must be empty, or the merge is refused. A separate integration worktree gets the overlap rule (REQ-LMF-005). Diagnose each untracked or modified file in the tree that holds `main` (commit it in its owning card worktree, move it there, or ignore it); never stash. Read `main` source through `git show main:<path>`, or — where a dedicated integration worktree holds it — in that worktree rather than the primary's working tree; `git show` does not cover a smudge or clean filter, end-of-line conversion, uncommitted changes, or a later move of `main`. (In this repository the primary checkout holds `main`, so its working tree is `main`'s tree.)

**[HARD] 4. Resynchronization**

When a release batch lands on `origin/main`, local `main` follows it. The tool performs this as a fast-forward-only update inside the integration window; it is not done from the operator's terminal. `moai integration resync` refuses when another tree holds `main` and no moai verb moves the primary checkout to the current local `main` tip; both are planned in t1847.

- (1) Take the window, and confirm that HEAD points at `main` and that the tree being updated is clean.
- (2) Run `git fetch origin main` and check its exit status. Pin the `origin/main` value obtained as the baseline, and make every later remote comparison against this value.
- (3) Read the ancestry relation between HEAD and the baseline. If HEAD is behind the baseline, check for ignored-file overwrites, then fast-forward. If the baseline is an ancestor of HEAD, or the two values are the same, report that no update is needed and leave HEAD where it is. If neither is an ancestor of the other, the histories have diverged: refuse, and HEAD does not move.
- (4) After the move, re-read HEAD to confirm that it equals the baseline and still points at `main`; then report the old and new SHAs and release the window.

The handling of the diverged case is set by operator decision Q2 (`.moai/specs/SPEC-LOCAL-MAIN-FLOW-001/decision-index.md`).

**[HARD] 5. Closing and worktree disposal**

- Closing a card merged on the local `main` surface does not push, and does not wait for a release batch.
- Disposing of a card worktree waits until the batch carrying that merge has landed on origin. Until then the card branch has no remote copy, so the worktree is the only copy of the work.
- The closing conditions are as follows. The merge commit must have two parents, its tree must equal the tree of the second parent, and it must be reachable from the integration branch. A re-measurement record keyed by that tree must pass verification. The decider is a human, and the leader's approval receipt must be verified. An evidence file produced by a lane cannot substitute for the re-measurement record.
- Implementation of this rule is owned by the follow-up card t1621. Until t1621 opens the done transition, merged cards cannot reach done.

**[HARD] 6. Batch release procedure**

`origin/main` advances only through release PRs and is never pushed to directly.

- (1) Push local `main` to the `release/main-batch-YYYYMMDD` branch on origin. `YYYYMMDD` is the date of the batch. The command is `git push origin main:refs/heads/release/main-batch-YYYYMMDD`.
- (2) Open a pull request from that branch to `main`.
- (3) Merge it with a merge commit.
- (4) When the merge is complete, fast-forward local `main` to `origin/main` through the resynchronization of item 4.
- The merge method does not use squash (no squash).

The pre-push hook's verdict on pushes of `main:release/*` has not yet been measured (OQ-10).

**[HARD] 7. Batch-closing rule**

- The timing of closing a batch is set by the count trigger. When `git rev-list --count origin/main..main` reaches `git_strategy.manual.lead_push_threshold`, the leader closes the batch and starts the procedure of item 6. The source of the threshold is `.moai/config/sections/git-strategy.yaml`.
- The trigger is evaluated only after the integration window has closed.
- The green condition takes precedence over the count trigger. Every card merge must pass the merge-tree re-measurement in the integration window, and while the `origin/main` CI of the most recently landed batch is red, the next batch is held. If that CI verdict does not yet exist, it is not a reason to hold.
- If the threshold is 0 or the key is absent, the trigger is off, and the leader decides when to close a batch. This is not an error condition.

Basis: SPEC-LOCAL-MAIN-FLOW-001 (spec REQ-LMF-010·REQ-LMF-011·§G; plan §B2·§B3·§B3a·§B4·§B7·§B9·§G.1). Rescope 0.4.0 of SPEC-GIT-STRATEGY-AGENTIC-TRUNK-001 (2026-10-21) renamed the model `agentic-trunk` and adopted the target-holder routing (card pull request merged by the leader with no auto-merge request where only the primary holds `main`; integration-window local merge where a dedicated worktree holds it); the landing and resync tool paths are unchanged.

### §4.1 Integration chain (main)

This closes the gap where each card is verified and merged on its own, but **no one checks the combined state**. On 2026-08-15, 12 PRs each entered main with a green check, and only after they converged did it come to light that `moai update` deletes local-only files. With the 2026-10-05 GitHub Flow transition (M2), the integration base is `main` — the chain below is the post-transition standard, and the develop chain remains as history, marked [SUPERSEDED] below.

**[SUPERSEDED by §4.0]** **[HARD] Standard chain — GitHub Flow transition (operator instruction, M2 landed 2026-10-05)**

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

1. **[SUPERSEDED by §4.0]** **Card branches are cut from `main`.** When a card finishes, it is applied to `main` through a card PR (base `main`). `develop` is legacy — do not use it as the base of a new card.
2. **[SUPERSEDED by §4.0]** **`main` is the remote default branch.** The CI on its head produces the integration verdict. Merges and pushes are performed **in bulk by the leader** (2026-09-02 — lanes do not push; the only exception is the candidate path (`ci/<card-id>`), which the 'candidate-path exception' below substitutes for this prohibition).
3. **The integration window is serial.** `moai integration acquire --card <card-id>` → merge → `release` — the push is outside the window (leader bulk, 2026-09-02). A lock is a device for serializing merges, not for serializing repairs, so if repairs remain, take the window only once they are ready.
4. **The verdict is CI.** A local pass is only an early signal — it is neither a clean environment nor the darwin/windows matrix. Do not reuse pre-merge verification as post-merge evidence: either measure again on the merge tree, or show that `git rev-parse <merge>^{tree}` of the merge commit equals the tree that was re-measured.
5. **[SUPERSEDED by §4.0]** **Local `main` is commit-dead (SPEC-MAIN-COMMIT-BAN-001, card t1337).** No session commits inside `main` in the primary checkout — `git commit` / `git revert` / `git cherry-pick` are refused by the BranchGuard (`workflow.branch_guard.deny_commits_on: [main]`), and commits are made only in card worktrees branched from `main`. The leader's push to `main` is closed by the batch trigger — when `git rev-list --count origin/main..main` reaches `git_strategy.manual.lead_push_threshold`, the batch is closed (the count comment in the configuration file still points at `origin/develop..develop` — drift). The source of the current value is the configuration file, and this rule only names the key and the count command. Pushes follow the green condition — per-card re-measurement of the merge tree in the integration window is the pre-gate for each card merge, and while the `origin/main` CI of the last push is red, the next push is held. Details: `.claude/rules/local/gitflow-lane-protocol.md` §4 (that rule still describes develop — drift).

**[SUPERSEDED by the chain above — 2026-10-05]** The develop GitFlow chain of 2026-08-29 to 2026-10-05 (cards branched from develop and merged back to develop with `--no-ff`, with only release PRs to main), and the regime before it (develop not pushed to the remote · a main PR per card · a disposable develop), were all discarded. The record of the discard and its reasons exists so that two documents do not issue contradictory instructions; the full text of the develop chain is in `.moai/docs/gitflow-integration-chain.md`.

**[HARD] Post-transition drift — the procedure document set.** `.moai/docs/git-workflow-doctrine.md`·`.moai/docs/git-local-workflow-doctrine.md`·`.moai/docs/gitflow-integration-chain.md`·`.claude/rules/local/gitflow-lane-protocol.md` still describe the develop chain. The canonical source of post-transition norms is this §4.1 together with the transition rules (`.claude/rules/moai/**`, card t1522), and when you open those documents, read them under this notice.

**Related document pointers (SPEC-RC-TESTBED-001)** — the procedure text is not kept here (as with the delivery.md delegation in rule 4 above, two copies diverge the moment they exist):

- The `rc.N` numbering policy, the untagged principle, and `BUILD_ID` build identification for local rc builds: `.moai/docs/version-management.md` — the **Local RC Numbering** section
- Updating local `main` after a merge (criterion and the BranchGuard-safe path): `.claude/rules/local/gitflow-lane-protocol.md` — the **develop update** section (that section still describes develop — drift; it cross-references the §9 rc runbook)

**[HARD] Lane obligations in -f / -l mode**

In Factory mode (leader `moai cc -f` · lane `moai cc -l`), when a lane finishes its card work it **must request that the leader apply the change to `main`.** A lane does not take the merge window on its own.

- **Self-dispatch lane exception — merge window.** A lane of a Claude self-dispatch factory run does not make the request above — through the integration procedure of `moai factory complete`, which calls `moai integration merge --card`, it takes the integration window itself and applies its own card to `main` (OD-2). Codex lanes are not covered by this exception: they stop at merge-ready (REQ-SD-025). This exception also does not change the prohibitions above on other queue changes (`add`, `drop`, `done`, `edit`, etc.) and on `moai contract sign` (only card leasing with `moai factory next` is allowed — OD-1; `.claude/rules/local/gitflow-lane-protocol.md` §6).
- Contents of the completion report: card id · branch and HEAD · local merge SHA · number of unpushed commits · evidence path (whether it was moved out of the primary) · re-measurement scope
- **[SUPERSEDED by §4.0]** `moai integration status` showing `free` is **not an approval.** Under the open policy, a lane promoted to the front of the queue that holds the window merges without a leader nomination.
- **Lanes do not git-merge into `main` by hand; they merge only with moai integration merge --card, or with moai factory complete, which calls it.** The queue schedules the window, and the leader sets only the policy with `moai integration policy open|hold` (to receive the window, the lane queues via acquire --wait — re-measurement happens before waiting; inside the window, only the merge).
- **[SUPERSEDED by §4.0]** When the window is granted: `moai integration acquire --name <lane> --card <card-id>` → absorb `git merge main` in your own worktree (the target is the **local** `main` — not the remote; before absorbing, first check whether that local main is up to date — the criterion and update path are in `.claude/rules/local/gitflow-lane-protocol.md` §11, which still describes develop — drift) → **re-measure on the merge tree** → enter the integration worktree → apply to `main` (merge the card PR with base `main`, or a `--no-ff` merge — the migration document sets the detailed form) → `moai integration release` → `ExitWorktree keep` → completion report (report the applied SHA to the leader — the push and the merge are done in bulk by the leader)
- **[SUPERSEDED by §4.0]** **[HARD] The card PR (base `main`) is the only public path (revised with the 2026-10-05 transition).** When a card closes, the lane pushes the card branch and opens a PR (base `main`), but **does not merge it directly** — the merge and push are performed in bulk by the leader, and the lane is not the party that does so. A lane also does not directly request or re-request CI, such as `gh run rerun` or `workflow dispatch` — it leaves the CI verdict to the run triggered by the application to `main`, and reading the verdict is the leader's job. (The former prohibition — operator instruction 2026-09-01, the WT-branch push ban and the develop batch-push regime — became history together with the develop chain; this clause was added in response to a precedent: lane-2 pushed `WT-version-stamp-predicate` to origin that same day.)
- **[HARD] Candidate-path exception — operator approval 2026-10-03 (SPEC-CANDIDATE-CI-001, REQ-CCI-016).** `moai integration candidate --card <card-id>` is the path that pushes a candidate commit to `ci/<card-id>` to receive a CI verdict before landing. For this one path only, it replaces `WT push · CI request prohibited`. The landing gate requires a green verdict for that candidate. Pushing the integration branch and merging remain a bulk action by the leader, and the candidate verb does not take the integration window. The same intent is recorded in `.claude/rules/local/gitflow-lane-protocol.md` §2·§4 and in three documents under `.moai/docs/`, and the develop descriptions and drift notices in those documents remain as they are.
- **[HARD] `acquire` verifies the caller's tree before it records the window.** It checks whether the working copy of the tracked `.claude/settings.json` has been modified, using `git --no-optional-locks status --porcelain -- .claude/settings.json`. On a hit, it preserves that copy under `.moai/state/settings-drift/` in the primary checkout, appends one line to `ledger.jsonl` in the same place, and prints the preserved path and sha256. **Detection, preservation, and the ledger run every time regardless of configuration** (the problem was that nobody looked for 9 days, not that nothing blocked it). Only refusal is opt-in (`workflow.settings_drift_gate.enabled`; this repository turns it on), and the bypass is `--allow-settings-drift` — `--force` belongs to a different axis, taking the window away from a live holder, so it is not used as a bypass. To check by hand regardless of the window, use `moai integration preflight [path]`. **Never restore automatically under any circumstances** — the file is written by the runtime and can contain tokens, absolute paths, and tmux pane ids, so automatic restoration would itself destroy data. When a hit is reported, the leader decides the disposition.
- Do not dispose of a worktree until its remote merge has been confirmed. The worktree of an unpushed branch is the only copy of its work.
- sync is finished **before** merging, inside the worktree. If you close only run and merge, the SPEC reaches `main` as `in-progress`, and the window must be taken again (a real case on 2026-08-29, t342).

**The operating procedures below have been moved to `.moai/docs/gitflow-integration-chain.md` (card t750) — the bash that executes the window (entering the integration worktree, absorbing, re-measuring, merging), the leader's bulk push procedure, the record of rejecting local CI, the friction from over-matching BranchGuard queries, and a caution about the docs-site Vercel binding. The procedure text in that document still describes the develop chain — when executing, read it under the transition notice above.** As long as the [HARD] rules above and the lane obligations stay unchanged, this summary is sufficient; when you execute a procedure, open the migration document.

---

## 5. Version Management

See: `.moai/docs/version-management.md` — SemVer 2.0.0 pre-release form (`-rc.N`), build version injection via ldflags, files requiring version sync, release process under the PR-mandatory regime.

---

## 6. Testing Guidelines

### ⚠️ IMPORTANT: Prevent Accidental File Modifications

When running tests, **always check if they modify project files**.

### Test Isolation

**[HARD] All test temp directories MUST be created under `/tmp` and cleaned up automatically** — `t.TempDir()` for every temporary directory (auto-cleanup registered).

**[HARD] Pitfall — `filepath.Join(cwd, absPath)` does not strip the leading `/` of the second argument**: `filepath.Join("/a/b", "/var/folders/x") = "/a/b/var/folders/x"` (wrong behavior). To resolve user-supplied paths, use `filepath.Abs()`. For measured examples and the full code, see `.moai/docs/local-dev-guide.md` § 6a.

### Coverage Targets

The full text has been moved to `.moai/docs/local-dev-guide.md` § 6a — 85% per package (`quality.yaml` `test_coverage_target`) and 90%+ for critical packages (cli·template·hook) are targets, and the only mechanical basis is the global gate of the strict profile.

### Go Test Execution Rules

- [HARD] After fixing ANY test, run the AFFECTED packages only and let CI give the full-suite verdict — the command, its `-timeout` derivation, and the no-local-full-suite rule live in §4 Before Commit (single source, not restated here)
- Do not declare success after fixing only the initially failing tests
- Run `go test -count=1 ./internal/<pkg>/...` to disable test caching when debugging flaky tests
- Run `go test -race ./internal/<pkg>/...` for concurrency safety on any code touching goroutines or channels
- Run `go vet ./internal/<pkg>/...` on the changed packages before committing

---

## 7. Hook Development Guidelines

See: `.moai/docs/hook-development.md` — shell-script-only hook pattern, hook wrapper template, settings.json format, `$CLAUDE_PROJECT_DIR` quoting rules, platform differences, timeout policy.

---

## 8. Template Variable Strategy

The full text has been moved to `.moai/docs/local-dev-guide.md` § 8 (card t1303). Summary: templates (`*.tmpl`) use Go template variables such as `{{.GoBinPath}}` (rendered at `moai init` time), while local settings use `$CLAUDE_PROJECT_DIR` (resolved by Claude Code at run time) — the reasons both path-variable strategies are needed, and the `TemplateContext` and `Deploy` signatures, are in the migration document. **[HARD] `.HomeDir` is prohibited in the `.sh.tmpl` fallback — use `$HOME` (same axis as §14).**

---

## 9. Configuration System

The full text has been moved to `.moai/docs/local-dev-guide.md` § 9 (card t1303). Summary: there is no single `config.yaml` — settings exist only as 32 files under `.moai/config/sections/*.yaml`. Precedence: env vars (`MOAI_DEVELOPMENT_MODE`·`MOAI_LOG_LEVEL`·`MOAI_LOG_FORMAT`·`MOAI_NO_COLOR` — the constants are in `internal/config/envkeys.go`) > sections/*.yaml > template defaults. `MOAI_USER_NAME`/`MOAI_CONVERSATION_LANG` are not implemented.

---

## 10. Build and Development Commands

Build, test, and lint targets are defined in the `Makefile` with `##` help text — list them with `make help`. For the order template edit → `make build` → test → commit, see §2 [HARD] Template-First Rule.

---

## 11. Frequent Issues and Solutions

Four frequently occurring issues — one-line summaries and details are in `.moai/docs/local-dev-guide.md` § 11a: templates not updated (`make build` required) · tests contaminating the real configuration (isolate with `t.TempDir()`) · hook timeouts (`{"timeout": 60}`) · **exit 137 leftovers from a reinstall** (you must do a clean reinstall with `rm -f ~/go/bin/moai && cp bin/moai ~/go/bin/moai` plus verify with `sh scripts/verify-local-install.sh`; bare `go install` is prohibited — a missing `LDFLAGS` disables provenance verification).

### [HARD] Bug or improvement found during use → `/moai:feedback` immediately

Whenever you find a bug or an area needing improvement while using MoAI-ADK, submit feedback immediately with `/moai:feedback` — do not batch it up for after the session ends. Scope: `moai` CLI behavior, hooks, templates, skills, agents, and factory operations defects in general. Record the reproduction command together with the observed output. Distinction: items to report to the maintainers go to `/moai:feedback`; items to schedule as work go to `/moai todo add`.

---

## 12. YAML frontmatter quick reference

The general format rules are defined in `skill-authoring.md` and `agent-authoring.md` under `.claude/rules/moai/development/`. The local development checklist (CSV prohibited · the `skills:` array exception · quoting of `metadata.*` · `make build` · local sync) and the location of detection scripts have been moved to `.moai/docs/local-dev-guide.md` § 12 (card t1303).

---

## 13. GLM Integration Testing

### [HARD] Do not run GLM integration tests in the dev project

The `moai cc`/`moai glm` command flows modify the real settings files, so never run them in the dev project.

- Unit tests: in the dev project, only the changed packages (`go test ./internal/<pkg>/...` — no local full suite, §4), and only files inside `t.TempDir()`
- Integration tests: run with `claude -p` in `/tmp/test-project`
- Auth token: `loadGLMKey()` (reads `~/.moai/.env.glm`); if absent, `t.Skip()`
- Prohibited: `t.Setenv("HOME", tmpDir)` (contaminates parallel tests), and hardcoded fake keys

---

## 14. Preventing hardcoding

### [HARD] No hardcoding in Go code (internal/, pkg/)

- Extract URLs, model names, organization names, and API headers into `const`
- Define environment variable names as constants in `internal/config/envkeys.go` and reference them from there
- Define thresholds once, as a single source, in `config/defaults.go`; duplication is prohibited
- For cross-platform code, prefer environment variables such as `$HOME` and `HOMEBREW_PREFIX`

### [HARD] `.HomeDir` is prohibited in .sh.tmpl fallback paths

`.HomeDir`/`.GoBinPath` get frozen as absolute paths at `moai init` time. Use `$HOME` for the fallback:
- Primary: `{{posixPath .GoBinPath}}/moai` (OK, init-time)
- Fallback: `$HOME/go/bin/moai` (MUST use `$HOME`)
- `renderer.go`: `$HOME` is already registered in `claudeCodePassthroughTokens`

### Areas where hardcoding is allowed

`AGENTS.local.md`, `settings.local.json`, and `_test.go` (within t.TempDir()).

---

## 15. Programming-language neutrality of templates

> **[HARD] Terminology — "language" refers to two axes. Do not use it without a qualifier.**
>
> | Axis | Values | Requirement | Governing rule |
> |---|---|---|---|
> | **Programming language** (16) | go, python, typescript, … swift | **Neutral** — no single language is treated as PRIMARY | this §15 |
> | **User conversation locale** (4) | ko, en, ja, zh | **Translation** — natural native prose for each locale | §8 Localization Contract of the active output style |
>
> The two are independent of each other. Templates are neutral across the 16 programming languages and, at the same time, translated into 4 locales.
> Because reading "16 languages" as the number of distribution locales is a misreading that has been observed repeatedly, this document always qualifies the term as
> **"programming language"** or **"locale"** when writing it.

### [HARD] Everything under `internal/template/templates/` treats the 16 programming languages equally

The tool's implementation language (Go) and the programming language of a user's project are separate matters. Templates are for all users.

- Programming-language bias is allowed in: `AGENTS.local.md`, `settings.local.json`, and the local `.moai/config/`
- Programming-language bias is prohibited in: all of `internal/template/templates/**`

### The 16 supported programming languages (all equal)

```
go, python, typescript, javascript, rust, java, kotlin, csharp,
ruby, php, elixir, cpp, scala, r, flutter, swift
```

Canonical name for Dart/Flutter: **"flutter"** (not "dart").

### Checklist (when editing templates)

- [ ] Did you avoid placing any specific language as "PRIMARY"?
- [ ] Are the 16 programming languages listed at equal level?
- [ ] Did you avoid downgrading the rest to "planned" while only one specific language is "enabled"?
- [ ] Is project_markers-based automatic detection logic included?
- [ ] A difference between the local config and the template is normal (identical contents are, if anything, suspicious)

Detailed lessons: `lessons.md` #5 in the global project memory — how to read it: that store (`~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/`) is **dormant** and is not loaded by sessions. Open and read it directly; to check both stores, use `moai memory doctor`.

---

## 16. Orchestrator self-check

### [HARD] Four self-check questions (required before starting a complex task)

1. Is this task the exclusive domain of a specialist agent?
2. Does that specialist agent exist in the catalogue? (CLAUDE.md Section 4)
3. Is delegation more advantageous than direct execution for quality, independence, and bias prevention?
4. Can part of this task be decomposed into a parallel spawn of read-only sub-agents? (Anthropic "Exploration-First Pattern": analyze with WebFetch + Explore subagent → synthesize in the main agent)

**If three or more answers are YES → direct execution is prohibited**. If the fourth question is also YES, apply the Exploration-First Pattern (parallel spawn of read-only sub-agents) first. Confirm the delegation method with AskUserQuestion before executing.

### Volume-based triggers

- Creating **5+** files of the same kind → delegation to `manager-develop` or `builder-harness` is mandatory
- New Go code of **500+ LOC** → `manager-develop` (cycle_type=tdd, domain context: backend) is mandatory — the former `expert-backend` is archived per SPEC-V3R6-AGENT-TEAM-REBUILD-001
- Creating **3+** agents/skills → `builder-harness` is mandatory

### Permitted direct execution

Typo/format fixes, editing one config setting, explicit user requests, absence of a delegation target, orchestration itself, git operations, and `/tmp` work.

### Order: Rule 5 → §16 → Rule 1

Rule 5 (WHAT) → §16 (WHO) → Rule 1 (HOW) → execute

Detailed lessons and 5 Whys: `lessons.md` #4 in the global project memory — a **dormant store** that sessions do not load (the same reading method as §15: open and read it directly, and check both stores with `moai memory doctor`).

---

## 17. Four-language document synchronization rules for the docs-site

The docs-site is the official user documentation at `adk.mo.ai.kr`. The full doctrine — URL standards, the 4-locale synchronization obligation, Mermaid TD-only, the Vercel project binding, and the build/deploy checklist — is in an external file.

See: `.moai/docs/docs-site-i18n-rules.md`

### §17.1 Design components and icon conventions → See `.moai/docs/docs-site-design-components.md`

The details of the docs-site Claude Warm Editorial [HARD] rules (icon shortcode · code blocks · Mermaid · footer · CSS cache busting · single light theme) have been moved to an external file. The essentials: decorative emoji in body text is prohibited → use the `{{</* icon */>}}` shortcode; the sidebar `icon:` value requires a matching `menu.html` SVG case; if a CSS change does not appear in dev, restart hugo.

---

## References

Sections §18-27 were consolidated into external `.moai/docs/` files to reduce launch-time context. Each entry below is the authoritative location for its domain.

- **§5 Version Management** (SemVer pre-release, ldflags injection, release process): `.moai/docs/version-management.md`
- **§7 Hook Development** (shell-script-only pattern, settings.json format, quoting rules): `.moai/docs/hook-development.md`
- **§18 Git Workflow** (the Enhanced GitHub Flow body, plus the parent-model note for the [2026-08-27] git-flow transition — the canonical source is §4.1 (`main`-based, 2026-10-05 transition); this document still describes develop — drift; branch protection `enforce_admins: true`; Hybrid Trunk RETIRED): `.moai/docs/git-workflow-doctrine.md`
- **§19 AskUserQuestion Enforcement + §19.1 Implementation Kickoff Approval** (REQ-ATR-015, revised to the default autonomous transition): canonical SSOT at `.claude/rules/moai/core/askuser-protocol.md` + `.claude/rules/moai/workflow/auto-semantics.md` §9.1 (the default form of plan→run entry is audit-cross evidence — plan-auditor PASS + audit-ready + hash integrity + no blocking issues — plus a decision record; the three keep-set categories [environment-impossible / operator-held / irreversible external-shared] keep the operator's direct answer)
- **§20 Vercel Build Cost Guard** [HARD]: all Vercel projects MUST use Elastic build machine ($0.0035/CPU min vs Turbo $0.126/min); check Build Machine setting first on cost anomalies
- **§21 Dev-Only Commands Isolation** (split harnesses, `SPLIT_HARNESS_NAMESPACE_LEAK` sentinel): `.moai/docs/dev-only-commands-isolation.md`
- **§22 Dev Settings Intent** (settings.json key semantics): `.moai/docs/local-dev-settings-intent.md`
- **§23 Local Git Workflows** (PR-mandatory, 1-person OSS, all tiers via PR — the release-PR path; [2026-08-27] transition to git-flow `develop`, then [2026-10-05] GitHub Flow transition returning to a `main` base — the canonical source is §4.1; this document still describes develop · drift): `.moai/docs/git-local-workflow-doctrine.md`
- **§24 Harness Namespace** (template-managed vs user-owned separation): `.moai/docs/harness-namespace-doctrine.md`
- **§25 Template Internal-Content Isolation** (neutrality catalogue, CI guard): `.moai/docs/template-internal-isolation-doctrine.md`
- **§26 Linear integration** (local-only): `.moai/docs/local-linear-integration.md`
- **§27 Agent-Skill Architecture** [HARD]: every agent gets ≥1 skill set (4 elements: workflow skill + knowhow reference + scripts + trigger); all skill bodies in English; `/moai:<sub>` slash-wrapping maintained; `/moai:harness` meta-harness (v4 Builder); Analyze-First execution plan before any work

---

## 28. LSEL drain operations (local self-evolution loop)

[HARD] The SessionStart wiring in `.claude/settings.local.json` runs `session_drain.sh` and `backlog_check.sh`. Do not call `drain.sh` directly; use only the preservation wrapper `session_drain.sh --inbox .moai/lessons-inbox.jsonl --state-dir .moai/state/lsel`. Candidate proposals are read from the preserved copies in `clusters-history/`, not from the volatile live file. The inbox is limited to tool and test-failure stubs. The wiring, verification, and learning boundaries are in `.moai/docs/lsel-drain-operations.md` and `.moai/docs/learning-channel-scope.md`.

---

## 29. Jev (TypeSafe System One) — local only

[HARD] `scripts/jev/` is a local tool of this repository and is not wired into the product. **Jev is decision advisory within a card (card t1542, revised 2026-10-07):** when `workflow.jev.enabled` is set, a lane calls `jev_ask` directly, and otherwise it calls the local scripts (such as `scripts/jev/route.sh`) directly, without a round trip to the leader. The output is **자문 입력** (advisory input) that the lane weighs, not a verdict; the lane therefore decides on its own from that input and records the decision record (auto-semantics.md §10) in its progress record. **Completion, merges, queue changes, and the operator keep-set gates still prohibit a Jev-only verdict (단독 판정 금지)** — that boundary has been redefined as advisory input plus the lane's judgment, not removed; the keep-set three categories keep the operator's direct answer. 한 가지 예외만 있다 — contract 모드 Kickoff 의 `llm+jev` 교차 확인에서는 `moai contract decide` 가 Jev 를 두 번째 신호로 직접 부를 수 있다. Jev 답은 LLM 의 승인을 확인하거나 사람에게 보낼 뿐 혼자서 판정을 시작시키지 못하며, 그 밖의 단독 판정 금지는 그대로다. Keep the key only in `~/.moai/.env.typesafe`, and check the card's secrets and customer data before any external transmission. If there is no key or no network, judge from the doctrine and the evidence you have read directly. The 0.50 confidence threshold is provisional. The grades, commands, and measurement limits are in `.moai/docs/jev-local-operations.md`.

## 30. Premise check before dispatch (cards that are several days old)

[HARD] Before dispatch, measure the premises of an old card in order: symbol cross-check with `git log --all -S`, attribution with `git merge-base --is-ancestor`, and a positive control on commits from before the repair. Output that is empty without a positive control is not an absence but an unmeasured result. Do not automatically cancel a card based on a model confidence score, and do not confirm that a premise has lapsed on that basis. Record the grounds for lapse in the verdict document, and follow the operator's ruling for queue changes. The measurement method and the rejected experiments are in `.moai/docs/stale-card-premise-check.md`.

## 31. Kickoff autonomy and delegation of decisions (operator policy)

[HARD] Kickoff proceeds autonomously unless the card explicitly specifies an operator gate — it follows the gate inventory plus the keep-set model (`.claude/rules/moai/workflow/auto-semantics.md` §9: the default autonomous transition is audit-cross evidence plus a decision record, and only the three categories environment-impossible / operator-held / irreversible external-shared keep the operator's direct answer). Do not override a mode that the operator has already chosen in the lane window. A lane does not ask the operator again about a choice; it escalates to the leader. Even when the leader uses Jev, it does not judge completion, merges, queue changes, or operator gates by a model's answer. The basis and scope of application are in `.moai/docs/kickoff-autonomy.md`.
