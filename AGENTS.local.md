# moai-adk local development rules (AGENTS.local.md)

Rules for developing MoAI ADK itself in this checkout (Rust `moai-cli`, RC `2026.11.0-rc.1`). `AGENTS.md` holds the generic MoAI workflow and rules; this file holds what is specific to this repository, this machine and the current RC. Tracked by the user's decision; keep secrets and private paths out.

## Ledgers and sources

- Issues on `modu-ai/moai-adk`, milestone `2026.11.0-rc.1`: root #1975, nine areas #1976-#1984, leaves #1985-#2016 plus #1973, #1974 and #2020 (00-E). Korean bodies are the detailed designs; shared-interface owners and HOLDs are in the bodies.
- `docs/decisions-2026-11-08.md`: every user decision of the day with time (KST), scope and affected Issues; also the leader's recorded operational decisions (marked as such).
- `.moai-cli/manifests/github-issues.json` and `.moai-cli/reports/issue-bodies/`: leader ledger mirror (key↔number map, link/dependency receipts, body hashes per design round, dispatch journal). Kept as is until 07-A/07-B deliver the product-grade forge adapter and outbox (user decision 2026-11-08).
- Design Pages (verified copies only): the approved RC1 structure and the feature contracts listed in the Issue root body. Reading rule: confirm a real read (logged-in browser or a user-provided full-text copy with path, URL, time and checksum).
- Memory: Claude auto-memory for this project (`feedback_*.md`, `project_*.md`, indexed by `MEMORY.md`); re-read the index at session start.

## Current scope (RC1, decided 2026-11-08)

- Ship the first version fast on the existing plan. The communication hub (`moai -l`/`-a` real-time channel) is deferred to the next version; its decisions are recorded for later and must not leak into RC1 requirements.
- Hosts: Claude Code and Codex; ZCode headless only; ACP, remote options, Orca/Hide interop later.
- Install (04-A): GitHub Releases + cargo-dist installer, artifact attestation only, tags `moai-cli-v<version>`, binary `~/.moai-cli/bin/moai-cli`, Windows self-replace. Go migration (04-D #2016): Rust `moai-cli migrate --from-go`; automatic removal limited to the Go binary and Go-only caches after self-verification with backup; `~/.moai` data preserved, credentials never copied.
- Paths: project state `<project>/.moai-cli/` (tracked), home install `~/.moai-cli/`. Go's `.moai/`, `~/.moai/` and this folder's `.claude/` are live Go MoAI state: never delete, overwrite or re-apply into them. `.claude/`, `.moai/` (any depth) and `AGENTS.md` are gitignored here.
- Shared-interface owners (12 groups): tree digest → 00-A #1985 (delivered `tree_digest(root, excludes) -> Result<String, String>`); GitHub read adapter → 00-B #1986; `.moai-cli` sub-paths and `src/agents.rs` refactor → 01-B #1989; exit code/error contract → 01-C #1990; host config merge/toml_edit → 02-B #1992; support matrices → 02-C #1993; redaction module → 03-B #1996; PR body fields, write_scope check and `docs/pr-contract.md` → 06-C #2007; outbox/operation marker → 07-B #2009; release workflows and the `rc-ci` required check → 08-D #2015; first `rc-ci.yml`, `label-sync.yml` and `PULL_REQUEST_TEMPLATE.md` → 00-E #2020.

## Sessions and roles

- Leader session (label `leader`): owns `src/`, `tests/`, `docs/`, `.github/`, `AGENTS.md`, `AGENTS.local.md`, `.gitignore`, `.moai-cli/`; holds the user question channel; dispatches Issues; relays lane reports; integrates PRs; keeps the decisions log. Leader commits also travel through a PR under the integration gate.
- Lane sessions (`agent-N`): one Issue at a time in their own worktree. `agent-2` is the read-only independent auditor (never writes to the repo or GitHub). `agent-1` owns `skills/moai/` only when assigned. New sessions are opened only by the operator; no session spawns another.
- Start sessions with `moai-cli session --role leader|agent --name <label> [--advisor fable] [--dry-run]`: main model `opus` at effort `medium`, advisor `fable` for hard reasoning (user decision 2026-11-08; configurable later).
- Each session uses its own `CARGO_TARGET_DIR=/tmp/moai-rc-target-<label>`; `CARGO_HOME=/tmp/moai-rc-product-cargo` is shared. A shared target dir let one session's build overwrite another's test binary (observed 2026-11-08).

## Lane loop specifics for this repository

- Dispatch is a fixed-field block (issue, branch, base, worktree, claim, cmd, evidence, rules), never prose. Branch `issue/<number>-<key>` from the latest `origin/v2026.11.0`; worktree at `<primary>/.moai-cli/worktrees/<number>-<key>` for every host (the Go `moai` PreToolUse hook rejects Edit/Write outside the primary checkout; `.claude/worktrees/` would write into preserved state and pollute the 00-A inventory; `/.moai-cli/worktrees/` is gitignored).
- The same hook blocks `gh` comments and Writes from lane worktrees, so lanes send their owned-report text to the leader, who posts it marked "레인 agent-N 보고 중계"; audit verdicts are relayed the same way. The hook itself is Go MoAI state and is not edited.
- Lane cycle (user decision 2026-11-08, ADK_LANE_LOOP / ADK_LOCAL_GUIDANCE_20261108_01): accept one ready card from the leader → enter that card's worktree → implement and run the full check set → open the PR → send the leader a durable report (exact head and base SHAs, each check command with its exit code, every AC as passed/failed/not_run, remaining responsibilities and HOLDs) → stop writing and confirm nothing is lost (`git status --short` empty and `git ls-remote origin <branch>` equals local HEAD) → leave the worktree without deleting it → from the primary location take the next non-conflicting ready card at once. A lane never waits for the merge, and its handoff is not PR completion.
- Worktree entry and exit (user decision 2026-11-08): a lane starts a card with `EnterWorktree(<primary>/.moai-cli/worktrees/<n>-<key>)`, and after the handoff leaves with `ExitWorktree(keep)` — the worktree is never deleted. Never move into or out of a worktree with a shell `cd` (nor `cd <worktree> &&` command prefixes). A worktree first entered by shell `cd` before this rule is finished as is; `ExitWorktree` is a no-op there, and the next card starts with `EnterWorktree`.
- Leader integration queue: the leader watches every existing lane (idle state, owner, recent command) and owns PR review, independent audit, `update-branch`, repair dispatch and merge to the end. A repair goes to the branch's last writer when that lane is free, otherwise to one free lane — one writer per branch. One card per lane that has actually handed off; the leader never creates a new or duplicate writer, and a duplicate session found after a restart freezes until the operator assigns it.
- Leader lane-check routine (user decision 2026-11-08): the leader always keeps checking lane progress so that no lane sits stopped by accident. It runs the check after every merge, every lane report or handoff, every restart, and whenever it would otherwise wait. Each check reads evidence, not status labels: `ListAgents` (who is live, busy or idle, duplicate names), the open PRs with head, base and merge state, each lane worktree's `git status --short`, and the lane's last reported command and exit code. Then it acts: a lane that has handed off gets the next non-conflicting ready card or a repair for its own branch; a lane idle without a reason, or busy past its reported step with no report, is asked for its actual last command and exit code; a block is recorded and raised to the user. Each check and its outcome go into the dispatch journal.
- No idle lanes (user decision 2026-11-08: "a lane waiting with nothing to do is the bottleneck"): the moment a lane opens its PR or pushes a repair, it reports to the leader and takes its next card — it never waits for the audit or the merge. The leader keeps a next card ready for every lane, picked in this order: (1) a ready card whose blockers are all closed; (2) a card whose only open blocker is a PR already in the leader queue — the lane starts from that PR's head, says so in its own PR, and re-merges onto `v2026.11.0` after the blocker merges; (3) a recorded follow-up turned into an Issue — audit non-blocking notes, defect candidates (RED test first, closed as not-a-defect when it does not reproduce), the ledger's `followup_candidates`; (4) when none of the three exists, the leader tells the lane why (which open PR or HOLD it waits on) and which card is reserved for it. The leader never invents work outside the RC1 Issues and that recorded backlog, never splits one card between two writers, and never lifts a HOLD to make work; a card that needs a HOLD answered waits for the user while the lane takes another.
- Blocks and limits: report a real block (permission or hook denial, missing decision, failing gate) with the exact command and stop. Never route around an unapproved permission, spend, HOLD or security gate; stop after three failures of the same step. No forced deletion or process kill, and no branch switch, reset or stash in the shared checkout. This cycle does not lift any HOLD.
- Loading: hosts do not load this file by themselves (Claude Code loads `AGENTS.md` only, and only without a `CLAUDE.md` variant; Codex reads `AGENTS.md` chains). Every dispatch tells the lane to read this file, and the leader restates the cycle in the dispatch message.
- Integration gate: ruleset `RC1 integration gate (v2026.11.0)` (id 26791615) requires a PR, the status `moai/audit`, an up-to-date branch (strict) and forbids force-push and deletion, with no bypass. Procedure: read the auditor's verdict first line, verify PR head == audited SHA and merge-base == base tip (else `gh pr update-branch` and re-audit), post `moai/audit=success` with the verdict comment as target_url, then `gh pr merge --auto --merge`. After every `update-branch`, compare the changed lines with the audited delta before asking for a delta audit: when both sides changed the same line the same way (for example a fixed array length), git merges silently and the build can break (observed on PR #2044, 2026-11-08) — registries such as `OPERATIONS` are length-free slices for that reason. Known gap: all sessions share one GitHub account, so the status source cannot be pinned (audit-only GitHub App is a later HOLD).
- Check set, run from the repository root before every PR and re-run after the final edit:

```sh
CARGO_HOME=/tmp/moai-rc-product-cargo CARGO_TARGET_DIR=/tmp/moai-rc-target-<label> cargo test --offline --locked
CARGO_HOME=/tmp/moai-rc-product-cargo CARGO_TARGET_DIR=/tmp/moai-rc-target-<label> cargo clippy --offline --locked --all-targets -- -D warnings
cargo fmt --check
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tests -p '*_test.py'
git diff --check
git diff --cached --check
```

Test in temporary projects or fixtures, never against installed user state. Do not claim Windows/Linux results from Mac fixtures.

## Assignment record and supervision

- Record: `.moai-cli/manifests/assignments.json`, schema in `docs/design-review/assignments-schema.md` (#2139, user decision 2026-11-09 00:3x). `unassigned` is a snapshot recomputed from the remote (open Issues of the active milestones minus those with a claim comment); `assigned` gets an entry only when a claim comment id exists; a claim that lost the lowest-id race stays as `stage: lost`.
- **The local file is not a lock; ownership is the claim comment.** The file is a same-host cache and write-ordering aid. A name in it grants no right to write that Issue, and when it disagrees with the remote, the remote wins.
- Leader supervision, on every 5-minute tick (and after every merge, lane report or restart, together with the lane-check routine above):
  1. Read PR state: open PRs with head, base, merge state and the `moai/audit` status on the current head.
  2. Compare the claim comments with `assignments.json`. On a mismatch, correct the file from the remote and record what differed as a comment on that Issue; the file keeps no history.
  3. When a PR needs a repair and its author lane is not free, send an `assignee-hint` to a different free lane (one writer per branch still holds).
  4. Refill: recompute `unassigned` from the remote and keep a next card ready for every lane.
- Stage changes are read from the remote (branch, PR, status, merge), never from a lane message or an idle notice alone. An expired lease never triggers automatic takeover; the operator decides.

## Repository and preservation boundaries

- Branches: integration target `v2026.11.0` (first commit `f6808390`; design commits `970e81f8`, `a6c67ba2`, `334ed74d`; first lane merges `1a152900`, `110a9377`, `cd6a9441`). `2026.11.0-rc.1` and `archive/main-based-2026.11.0-rc.1` are preserved refs; remote `main` is untouched (its Go-generation workflows still run on RC PRs via `pull_request_target` — cleanup is a user HOLD). Re-read `HEAD` and the branch immediately before any commit or push; never reset, rebase, stash or switch branches in the shared checkout.
- Stage by explicit pathspec after re-reading `git status --short`; never `git add -A`, `git add .` or `commit -a`. Scan staged bytes for secrets and personal paths; `.claude/`, `.moai/`, `AGENTS.md`, `target/`, `.mcp.json` and raw `references/` are never staged.
- The sibling `moai-adk-go` repository is read-only reference: preserve its dirty files, binary, credentials and remote main; never invoke the installed Go `moai` for RC tests. The Go-generation `.github` files were removed from this checkout on the user's instruction (23 paths, byte-identical to the Go templates, listed in `.github/moai/`); do not recreate them.
- No release, deployment or Go removal under the current scope.

## Implementation map

- `Cargo.toml`, `src/main.rs`, `src/lib.rs`, `src/agents.rs`, `src/session.rs`, `src/baseline.rs`: `project-plan/apply`, `package-inspect/resolve`, `agents-plan/apply`, `session` (dev launcher), `baseline-inventory` + `tree_digest`. Contract: `docs/rust-product-path.md`.
- `skills/moai/`: host-neutral common skill (not yet projected into any host).
- `.github/`: RC1 Issue forms (`moai-{feature,bugfix,investigation}.yml`), `.github/moai/` (labels record, README), and the 00-E workflows and PR template once merged. Issue forms render only from the default branch.
- `docs/`: `rust-product-path.md`, `lane-loop-contract.md`, `decisions-2026-11-08.md`, `session-hub-research-2026-11-08.md` (deferred hub), `github-plan-proposal.md`, `github-inventory-rc1.md`, `baseline-rc1.md`.
- `tests/product_cli.rs` (Rust fixtures), `tests/*_test.py` (Python fixtures, GitHub asset checks, inventory collector), `tests/fixtures/github/` (00-B recorded inventory).
- `adapters/claude/` (provisional Python, not called by Rust) and `experiments/` are not integrated product features.

Still open for M1: interview orchestration, common skill/host projection, lossless profile/AGENTS updates, all-host preflight and safe recovery.

## Machine notes

- Dev Mac; GitHub auth through `gh` as the operator; local Claude Code 2.1.319 (above the AGENTS.md v2.1.277 and Fable v2.1.257 thresholds). Tools available to sessions: Hide, ego-browser (logged-in Page reads), `gh`. None grants publication or deletion permission.
- Response language: Korean (colloquial in chat, clean written register in reports); code, identifiers, commands and paths verbatim.
