# moai-adk-go Local Development Guide

> **Purpose**: Essential guide for local moai-adk-go development
> **Audience**: GOOS (local developer only)
> **Last Updated**: 2026-10-20

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

See: `.moai/docs/local-dev-guide.md` § 1 — work locations, the edit→`make build`→test→commit cycle, and the [CRITICAL] `moai` CLI vs `/moai` slash-command distinction table. The core in one line: edit templates in `internal/template/templates/`, regenerate derived artifacts with the explicit emit targets (`make agents-emit`, `make commands-emit`, `make embed-manifest`), and use `make build` to check drift and embed the result; commit in a card worktree (the primary checkout is prohibited — AGENTS.md §2·§3).

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
internal/template/templates/AGENTS.md.tmpl
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

### [HARD] §2.0 Three copies of the agent definitions — full text moved

The three-copies table (C1 local / C2 template source / C3 machine-emitted), the `make agents-emit` obligation, the do-not-hand-edit-C3 rule, the three emission checkpoints, and the §2.1 template-content-neutrality paragraph moved to `.claude/rules/local/agents-local-detail.md` §C (paths:-scoped — it loads when `internal/template/**` or the other claimed paths are touched). Essentials, still [HARD]: edit C2 (`internal/template/templates/.claude/agents/moai/*.md`), run `make agents-emit` afterward — never hand-edit C3 (`templates/.codex/agents/moai/*.toml`, machine-emitted); the canonical neutrality catalogue is `.moai/docs/template-internal-isolation-doctrine.md` §25.1.

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

The full text is in `.moai/docs/local-dev-guide.md` § 2b·2c. The tracked per-file embed allowlist in `internal/template/embed_manifest_gen.go` carries the templates and catalog; `internal/template/embed.go` exposes that filesystem. Use the explicit `make agents-emit`, `make commands-emit`, and `make embed-manifest` targets when their inputs change, then let `make build` check drift and compile — `make build` does not regenerate agent or command outputs. Never hand-edit the generated agent or command artifacts; the command count is derived from the emitter inputs rather than maintained here.

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

### §4.0 Current card policy — draft PR, review gate, and safe handoff

This section is the operator-approved card-delivery policy (2026-10-19). It supersedes the local-main batch/card-PR model for card delivery — the superseded clauses are preserved as history in §4.0-H and `.claude/rules/local/agents-local-detail.md` §A·§B without execution authority over current card delivery. It states required behavior, not a claim that the installed tooling implements every step: where a step has no landed implementation it is named as an in-progress task or limitation, and the operative path is the manual command named here.

**[HARD] 1. Leader boundary.** A session assigned as leader or coordinator coordinates, delegates, reads evidence, and supervises. It never directly implements, repairs, builds, or runs tests in any worktree. There is no emergency, trivial-change, worker-absence, or explicit-request exception. Preserve an existing diff, its base, paths, and evidence before handing it to the officially assigned worker; never overwrite or appropriate another writer's work. If no worker is available, report the assignment blocker.

**[HARD] 2. Card worktree and draft PR.** Use a fresh, short-lived card worktree from the verified configured base, one writer, with the three traceability carriers (the dispatch `card:` field, the card id in every commit message, the evidence path). Finish the required scoped local checks, then open a draft PR for the exact verified head with `gh pr create --draft`. The manual draft-PR path is the operative delivery path: `moai factory complete` currently opens a non-draft PR and requests auto-merge, and the draft/auto-merge separation is an in-progress implementation task — until it lands, do not describe the tool as opening draft PRs. Keep required CI, E2E, typecheck, and security gates; external PR code must not receive secrets.

**[HARD] 3. Review gate (incident-derived; approved policy).** A merge is allowed only after confirming the **current** head SHA's `codex-review-state` two legs — `code_review` AND `security_review` — are both `completed` with zero unresolved findings. The state set includes `Failed` (execution failure): any state other than `completed` is not a pass and is a re-request target. **When** the head SHA changed after the review, the review shall be re-requested before merge; a completed review stays valid while the head SHA is unchanged (a same-SHA re-push neither refreshes nor invalidates it). An unprotected integration base is no exception to any of the above. The leader is coordination/supervision only — it does not implement or test — and the #1841 merge-before-review-completion violation is the concrete failure this rule prevents (merge 05:13:31Z preceded both review completions 05:14:11Z / 05:15:04Z; evidence: `inputs/REPORT.md` and `inputs/native-codex-rest-comment.json` under `.moai/specs/SPEC-AGENTS-CONTRACT-REVISE-001/`).

This gate is approved policy plus an in-progress implementation task (card t1868, in progress): no sentence here presents the gate as already-implemented tooling. Observed triggers and re-review behavior (PR-level observations, 2026-10-19): the ready transition started the review (#1841); an `@codex review` comment started the review (#1831, #1832, #1834 — #1832 and #1834 have recorded results), so manual triggering is an observed trigger as well and manual triggering from draft is not asserted impossible; the head changed after ready with no automatic re-review observed (#1836, #1842). Official Codex review triggers, per the Codex summary comment on PR #1844 (observed 2026-10-19T06:05Z): (i) opening a PR for review, (ii) changing a draft to ready, (iii) an `@codex review` comment, (iv) an `@codex security review` comment; an in-progress review shows 👀 and finishing without findings shows 👍.

**[HARD] 4. Auto-merge conditions and mechanism accuracy.** The repository's auto-merge mechanism (`.github/workflows/auto-merge.yml`) is measured as: a `workflow_run` trigger on CI completion (L7-10); a job condition of `event == 'pull_request' && conclusion == 'success'` (L23-26) — no draft filter, no configuration gate; the merge call carries `--match-head-commit` and `--delete-branch` (L355). Conditions (approved policy): conditional auto-merge applies only to a non-fork, non-bot PR authored by the verified connected GitHub user — identity looked up from the actual connected account, never a hardcoded product login. Other-author, fork, and bot PRs must not have auto-merge enabled and require BOTH the user's test verification confirmed AND explicit approval — both together, never one alone. Eligible own PRs still keep latest-SHA required CI, real Codex review, resolved findings, protection rules, and any third-party approvals. The repository stays public and automatic branch deletion is forbidden, own PRs included — the active `--delete-branch` behavior mismatches this policy and is recorded as a limitation. The conditional own-PR setting is a local operation choice, not a change to the product-wide manual default. Stranded-arm limitation: once auto-merge is requested, the arm persists on the PR (`gh pr view --json autoMergeRequest`); before retrying a gated pr-open PR, check that stranded `auto_merge:on` state and cancel it (`gh pr merge --disable-auto`) or account for it.

**[HARD] 5. Safe handoff and receiver confirmation.** A handoff records: the card id, base and head SHA, the PR, scoped verification evidence, the review/repair owner and next action, and the owned lease/resource disposition (integration window, test slots). The worker takes the next independent card in a fresh worktree only after receiver-side confirmation of that handoff — the receiver (leader or designated reviewer) has read the recorded evidence. Opening the PR is not card completion: closure requires the confirmed remote merge and the required completion evidence, and the original branch/worktree stays recoverable until that boundary. A preservation push of a card branch by its assigned worker is in-scope delivery, not a protocol violation.

**[HARD] 6. Review-fix return path.** When review findings arrive on an open PR, the recorded repair owner returns to the work: re-enter the card worktree (`ExitWorktree` to the primary checkout, then `EnterWorktree(<card-tree>)` — never a second lease), push fixes as new commits to the same PR branch (never force-push), and re-request the review — the new head SHA obliges the re-request. If the card holds the integration window or a merge-ready slot when findings arrive, release it first (`moai integration release`; verify with `moai integration status`) — a pr-open card holds no window; only the merge verb takes one. Release owned test/resource slots at their documented boundary (`moai slot release --resource <name>`) before stepping away. The card record's state (merge-ready / pr-open / failed / abandoned) and the leases it still holds are read from the record — a holder string or PR existence alone is not evidence that leases were returned.

**7. Wait only on real dependencies.** Only actual code or acceptance dependencies block a successor; a failed card blocks its affected dependants, not the entire backlog. Isolate failures; end active work before releasing owned leases.

**[HARD] 8. Primary-main preservation.** Recovery of the shared primary checkout requires inventory → preservation → an explicit handoff plan → delegation to the officially assigned worker. Do not run `git checkout` / `git reset` / `git stash` / force-checkout to align the primary with `main`. A clean dedicated main tree is not proof the shared primary is safe to move.

**9. Policy vs implementation.** Deployment policy: manual-by-default with staged independent opt-in for issuance, dispatch/execution, push/PR creation, review repair, merge, and release/deploy — enabling one never enables another; automatic PR creation does not authorize automatic merge or deployment. Code-side default values for this policy are a separate implementation task (in progress); this text states the policy, not the current shipped state. Preserve explicit operator choices and the local factory override during update and migration. The conditional own-PR auto-merge setting approved for this repository is a local operation choice — not a change to the product-wide default. This policy does not authorize implementing or switching to Kanban mode, and a Kanban simulation is never treated as implementation approval.

**10. Command forms.** The automated card-dispatch forms are: the CLI shell command `moai todo --auto`; the slash-skill commands `/moai todo --auto-lane` (lane session) and `/moai todo --auto-leader` (leader session). `moai todo --auto-lane` is never a shell form. Code-side occurrences of these strings (SessionStart hook, codex launcher) are card t1695's delegated scope.

### §4.0-R Release-path procedure — retained until #1830 processing

**[HARD] Batch release procedure.** The batch release procedure remains the current release-path procedure until issue #1830 is processed. `origin/main` advances only through release PRs and is never pushed to directly.

- (1) Push local `main` to the `release/main-batch-YYYYMMDD` branch on origin. `YYYYMMDD` is the date of the batch. The command is `git push origin main:refs/heads/release/main-batch-YYYYMMDD`.
- (2) Open a pull request from that branch to `main`.
- (3) Merge it with a merge commit.
- (4) When the merge is complete, fast-forward local `main` to `origin/main` through the resynchronization procedure (`.claude/rules/local/agents-local-detail.md` §A item 4 — the criterion is the `origin/main` rev-list comparison; diverged histories refuse).
- Merge-flag rule (every merge-flag statement matches this): a `release/*` OR `hotfix/*` branch merges with `--merge`; every other branch merges with `--squash` (measured: `.github/workflows/auto-merge.yml:341-345`). The batch release PR itself (a `release/*` branch) therefore merges with a merge commit.

The pre-push hook's verdict on pushes of `main:release/*` has not yet been measured (OQ-10).

**[HARD] Batch-closing rule.**

- The timing of closing a batch is set by the count trigger. When `git rev-list --count origin/main..main` reaches `git_strategy.manual.lead_push_threshold`, the leader closes the batch and starts the §4.0-R procedure. The source of the threshold is `.moai/config/sections/git-strategy.yaml`.
- The trigger is evaluated only after the integration window has closed.
- The green condition takes precedence over the count trigger. Every card merge must pass the merge-tree re-measurement in the integration window, and while the `origin/main` CI of the most recently landed batch is red, the next batch is held. If that CI verdict does not yet exist, it is not a reason to hold.
- If the threshold is 0 or the key is absent, the trigger is off, and the leader decides when to close a batch. This is not an error condition.

### §4.0-H Historical local-main batch integration — not the current card procedure

> **Transition-scoped record (superseded for card delivery, 2026-10-19).** The clauses below
> were the card-delivery procedure before the per-card draft-PR policy (§4.0): card worktrees
> branching from local `main` as both card base AND landing surface, tool-path merge landing
> through `moai integration merge --card`, and card closing gated on the batch landing. They
> are preserved as history and carry NO execution authority over current card delivery. The
> full historical text is `.claude/rules/local/agents-local-detail.md` §A. Preservation,
> one-writer, ownership, and resource-lease safeguards remain binding.

- [HARD] (historical) Branch base: card worktrees branched from local `main` — the base discipline survives in §4.0 item 2 (fresh worktree from the verified configured base); landing on local `main` is superseded by the draft-PR path.
- [HARD] (historical) Merge landing through `moai integration merge --card` in the integration window — superseded for card delivery; the verb and window remain the mechanism behind the retained §4.0-R resynchronization.
- [HARD] (historical) Clean-tree rule — STILL BINDING as a safeguard: a landing or merge tree must be clean (`git status --porcelain` empty); never stash to clean it.
- [HARD] (historical) Resynchronization after a batch lands — retained under §4.0-R item 4 (fast-forward-only, inside the integration window, never from the operator's terminal).
- [HARD] (historical) Card closing waited for the batch landing — superseded: closure requires the confirmed remote merge of the card's PR plus the completion evidence (§4.0 item 5).

Basis: SPEC-LOCAL-MAIN-FLOW-001 (spec REQ-LMF-010·REQ-LMF-011·§G; plan §B2·§B3·§B3a·§B4·§B7·§B9·§G.1) — historical basis of the superseded model.

### §4.1 Integration chain (main) — historical

> **Transition-scoped record (superseded for card delivery, 2026-10-19).** The integration
> chain that stood here — the 2026-10-05 GitHub Flow chain, the develop GitFlow chain before
> it, and the lane obligations around the local-landing window — is retained for provenance in
> `.claude/rules/local/agents-local-detail.md` §B. It is not authority to bypass the current
> draft-PR, review, role, or preservation gates (§4.0). Two [HARD] items remain binding as
> ordering and safety discipline outside their superseded delivery model:
>
> - [HARD] Complete run AND sync before the PR merges — a card merged with run only leaves its SPEC `in-progress` on `main`, and the landing has to be taken again (a real case on 2026-08-29, t342).
> - [HARD] `acquire` verifies the caller's tree before it records the window — detection, preservation, and the ledger run every time; automatic restore of `.claude/settings.json` is prohibited under any circumstances (full text: companion §B).

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
2. Does that specialist agent exist in the catalogue? (`.moai/config/sections/delegation.yaml` and the current agent definitions under `.claude/agents/moai/`)
3. Is delegation more advantageous than direct execution for quality, independence, and bias prevention?
4. Can part of this task be decomposed into a parallel spawn of read-only sub-agents? (Anthropic "Exploration-First Pattern": analyze with WebFetch + Explore subagent → synthesize in the main agent)

**If three or more answers are YES → direct execution is prohibited**. If the fourth question is also YES, apply the Exploration-First Pattern (parallel spawn of read-only sub-agents) first. Confirm the delegation method with AskUserQuestion before executing.

### Volume-based triggers

- Creating **5+** files of the same kind → delegation to `manager-develop` or `builder-harness` is mandatory
- New Go code of **500+ LOC** → `manager-develop` (cycle_type=tdd, domain context: backend) is mandatory — the former `expert-backend` is archived per SPEC-V3R6-AGENT-TEAM-REBUILD-001
- Creating **3+** agents/skills → `builder-harness` is mandatory

### Leader boundary and assigned-worker execution

A session assigned as leader or coordinator may coordinate, delegate, read evidence, and prepare handoff reports. It shall not implement, repair, build, or run tests directly in any worktree — and it never directly implements under any framing of the work. Trivial edits, urgency, explicit requests for completion, and missing workers do not create an exception. Preserve existing changes and their provenance and delegate them to the officially assigned worker; if none is available, report the assignment blocker. The volume-based triggers above never relax this role boundary. An assigned worker performs the bounded work within its own delegated scope.

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

The consolidated §18-27 document index moved to `.claude/rules/local/agents-local-detail.md` §D (paths:-scoped — loads for the SPEC, local-rules, docs, workflow, CLI/hook/template, and CI-workflow paths it claims). The pointed-to documents remain authoritative at their `.moai/docs/` locations; §19's canonical SSOTs remain `.claude/rules/moai/core/askuser-protocol.md` and `.claude/rules/moai/workflow/auto-semantics.md` §9.1.

---

## 28. LSEL drain operations (local self-evolution loop)

[HARD] The SessionStart wiring in `.claude/settings.local.json` runs `session_drain.sh` and `backlog_check.sh`. Do not call `drain.sh` directly; use only the preservation wrapper `session_drain.sh --inbox .moai/lessons-inbox.jsonl --state-dir .moai/state/lsel`. Candidate proposals are read from the preserved copies in `clusters-history/`, not from the volatile live file. The inbox is limited to tool and test-failure stubs. The wiring, verification, and learning boundaries are in `.moai/docs/lsel-drain-operations.md` and `.moai/docs/learning-channel-scope.md`.

---

## 29. Jev (TypeSafe System One) — local only

[HARD] `scripts/jev/` is a local tool of this repository and is not wired into the product. **Jev is decision advisory within a card (card t1542, revised 2026-10-07):** when `workflow.jev.enabled` is set, a lane calls `jev_ask` directly, and otherwise it calls the local scripts (such as `scripts/jev/route.sh`) directly, without a round trip to the leader. The output is **자문 입력** (advisory input) that the lane weighs, not a verdict; the lane therefore decides on its own from that input and records the decision record (auto-semantics.md §10) in its progress record. **Completion, merges, queue changes, and the operator keep-set gates still prohibit a Jev-only verdict (단독 판정 금지)** — that boundary has been redefined as advisory input plus the lane's judgment, not removed; the keep-set three categories keep the operator's direct answer. Keep the key only in `~/.moai/.env.typesafe`, and check the card's secrets and customer data before any external transmission. If there is no key or no network, judge from the doctrine and the evidence you have read directly. The 0.50 confidence threshold is provisional. The grades, commands, and measurement limits are in `.moai/docs/jev-local-operations.md`.

## 30. Premise check before dispatch (cards that are several days old)

[HARD] Before dispatch, measure the premises of an old card in order: symbol cross-check with `git log --all -S`, attribution with `git merge-base --is-ancestor`, and a positive control on commits from before the repair. Output that is empty without a positive control is not an absence but an unmeasured result. Do not automatically cancel a card based on a model confidence score, and do not confirm that a premise has lapsed on that basis. Record the grounds for lapse in the verdict document, and follow the operator's ruling for queue changes. The measurement method and the rejected experiments are in `.moai/docs/stale-card-premise-check.md`.

## 31. Kickoff autonomy and delegation of decisions (operator policy)

[HARD] Kickoff proceeds autonomously unless the card explicitly specifies an operator gate — it follows the gate inventory plus the keep-set model (`.claude/rules/moai/workflow/auto-semantics.md` §9: the default autonomous transition is audit-cross evidence plus a decision record, and only the three categories environment-impossible / operator-held / irreversible external-shared keep the operator's direct answer). Do not override a mode that the operator has already chosen in the lane window. A lane does not ask the operator again about a choice; it escalates to the leader. Even when the leader uses Jev, it does not judge completion, merges, queue changes, or operator gates by a model's answer. The basis and scope of application are in `.moai/docs/kickoff-autonomy.md`.
