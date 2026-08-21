# AGENTS.md — standing contract for agents in this repository

Every clause here binds a turn regardless of which agent harness drives it. The file is
**self-sufficient**: it assumes no other instruction file is loaded, and no nested `AGENTS.md`
exists anywhere in this repository.

**Budget warning.** A personal `~/.codex/AGENTS.md` joins the same merged chain and is consumed
**before** this file, narrowing what the project's contract can carry. Overflow is dropped from the
**tail**, silently — no warning, no stderr, exit 0. Clauses below are ordered most-critical-first
for that reason.

Obligations are carried from `.claude/rules/moai/**` and `CLAUDE.md`, which remain the source of
truth; compression removed rationale and incident records, never an obligation. Claude-only
mechanisms (the question channel, subagent spawning, skills, session handoff) stay there.

---

## 1. Evidence and verification claims

**No unobserved claim.** An actor MUST NOT assert a verification, a completion, **a defect / debt /
drift, OR the premise underlying a recommendation** it did not actually verify with the domain's
mechanical tooling. Evidence absent is not evidence of success — nor of failure. The absence of a
failure signal never establishes that a check passed; a text-pattern inference is a hypothesis, not
a verified defect; a reference existing does not establish that the referenced capability is still
live. Reachability is not justification.

**Baseline-integrity attribution.** Every verification claim MUST be attributed to an
actually-measured baseline — the command that was run plus the output observed, in this run,
against this tree. A figure carried over from another package, tree, or point in time is not a
baseline; using it as a fresh measurement violates this. Anything unattributed is a Gap, not a
Claim.

**Evidence-bearing report format.** Verification and completion reports SHOULD carry five sections,
on every report and not only the first: **Claim** (what is asserted); **Evidence** (the command run
plus its verbatim output — a summary is not evidence); **Baseline-attribution** (what it was
measured against, in this run); **Gaps** (what was explicitly NOT observed — an empty Gaps section
asserts nothing was left unobserved, which must itself be true); **Residual-risk** (what could
still be wrong despite what was observed).

---

## 2. Git, branches, and the shared checkout

The primary checkout is shared — several sessions may work in it at once, and branch state there is
global.

**Never change branch state in the primary checkout.** Forbidden there: `git checkout <branch>` /
`git switch` (relocates every concurrent session's tree); `git checkout -b` / `git switch -c` /
`git branch` (same, plus an unexpected branch); `git reset --hard` / `git checkout -- <path>`
(discards work of unknown provenance); `git stash` (repository-global — it silently absorbs another
session's uncommitted changes); `git rebase` / `git merge` onto the checked-out branch (rewrites or
advances shared history mid-operation). Read-only inspection, `git fetch`, commits to the
already-checked-out branch, and pushing it are permitted.

**Re-read branch and commit state immediately before any commit or push** — never a value read
earlier in the turn, never the branch reported at session start:

```bash
git rev-parse --short HEAD
git branch --show-current
```

A difference from what the turn assumed means another actor is writing the same tree: stop and
report the divergence instead of proceeding.

**Never sweep-stage.** In the primary checkout, never `git add -A`, `git add .`, or
`git commit -a`. Stage by explicit pathspec and re-read `git status --short` immediately before
staging, so another session's files are visible and excluded. This binds **even when no foreign
session was detected** — one can arrive after the check, and the sweep is what turns its presence
into lost work.

**Detect parallel sessions before a non-trivial direct edit** to a shared path (`.claude/`,
`.moai/`, `internal/`, `pkg/`, `cmd/`, repo-root config), and surface any divergence:

```bash
git fetch origin main 2>&1
git rev-list --count --left-right origin/main...HEAD
```

`0 0` or `0 N` proceeds; `N 0` or `N M` means resolve before editing. Where another live session
shares the checkout, isolate into a worktree rather than editing in the shared tree. The check
decays — re-run it before any commit and after a long pause.

---

## 3. Worktrees

**Work inside a worktree, entered through the launcher** (`moai cc -w <name>`,
`moai cc -w <name> --spawn` for a new window, `EnterWorktree(<path>)` to re-enter); never create one
with a bare `git worktree add`. Leave with `ExitWorktree`. Drive a worktree with `git -C <path>`,
not `cd`.

**`moai worktree done` closes L2 trees only.** A tree under `.claude/worktrees/` is L1, is absent
from the registry, and is disposed by the session-end prompt or by `git worktree unlock` +
`git worktree remove`.

**A card's branch is unpushed, so its worktree holds the only copy of the work.** Dispose of no
worktree — L1 or L2 — until the branch is integrated and the remote merge has landed.

**Start a new card in a new worktree.** Exit any previous worktree back to the primary checkout
first, or the new card's work lands on the old card's branch. Create the fresh tree from the remote
default branch; never reuse the previous card's tree. Where the new card depends on a prior card's
unmerged code, merge that branch inside the new worktree.

**Card worktree branches carry the `WT-` prefix and a descriptive slug, never the card id.** Rename
in place immediately after creating the tree: `git branch -m WT-<slug>`. Re-entry resolves by tree
name, so the rename is safe; the worktree directory keeps the card id.

**Three traceability carriers are then mandatory**, because the branch name no longer identifies
the card: the dispatch's `card:` field, the card id in every commit message on the branch, and the
card id in the evidence path (`.moai/reports/<card-id>/verdict.md`).

---

## 4. How verification is run

**Scope verification to the change**: run the tests the change can affect, then push and let CI run
the full suite. A full-suite run on a loaded developer machine measures the machine, not the code.

**Never spawn background load.** Where a verification needs contention, the load must be
cleanup-guaranteed — kills registered with the test framework's cleanup hook, or a `timeout`
wrapper bounding the process from outside. A trailing `kill` is not cleanup.

**Scrub the environment in one compound invocation.** Inside a worktree, an environment-scrubbed
verification runs as a single `unset <VARS> && <command>` call; a separate `unset` does not carry
into the next command, because each invocation is a fresh process.

**Batch independent read-only verifications rather than serializing them** across turns. Serialize
only for a genuine dependency: one command's output feeding another, writes to the same path, or
shared-state mutation.

**A CodeRabbit row in `gh pr checks` is not evidence that a review ran** — the status reads
`success` and prints `pass` identically whether or not one did. Count the row only when BOTH hold:
(1) `gh api "repos/$REPO/commits/$HEAD_SHA/status"` reports the `CodeRabbit` context with
`state == "success"` and description `Review completed`; (2) a `Merge Risk:` line exists whose
commit prefix matches the current `headRefOid`. Anything else is a gap, not a pass; `Review rate
limited` means the review never started.

---

## 5. Core behaviors

Six behaviors bind every turn, whatever the task.

**1. Surface assumptions.** Before implementing anything non-trivial, list assumptions explicitly
and wait for confirmation — silent assumptions are the most dangerous misunderstanding. State them
as a short list and invite correction. Anti-pattern: silently picking one reading of an ambiguous
requirement and running with it.

**2. Manage confusion actively.** On an inconsistency, a conflicting requirement, or an unclear
specification: STOP — do not proceed on a guess; name the specific confusion; present the tradeoff
or clarifying question; wait for resolution. Anti-pattern: "the spec says X but the code does Y",
then silently choosing Y because it is easier.

**3. Push back when warranted.** Say so directly when an approach has a concrete downside,
contradicts an established convention without justification, or breaks a tested invariant. State
the issue, quantify the downside ("adds ~200 ms latency", not "might be slower"), propose an
alternative, and accept an override once the user has full information. Sycophancy is a failure
mode. Anti-pattern: "Of course!" followed by a known-bad implementation.

**4. Enforce simplicity.** Actively resist overcomplexity; generation tends toward
over-engineering. Before completing, ask: fewer lines without losing clarity? are these
abstractions earning their complexity? would a staff engineer ask "why didn't you just…"? Apply the
ladder in order, cheapest capability first: (1) does this need building at all? (2) does a helper,
type, or pattern already exist here — reuse it; (3) does the standard library do it; (4) a native
platform feature; (5) an already-installed dependency; (6) can it be one line; (7) only then, the
minimum code that works. The ladder is language-neutral. **Never simplify away safety**: it MUST
NOT be used to drop input validation at trust boundaries, error handling that prevents data loss,
security measures, accessibility, or one runnable check behind non-trivial logic. If an
implementation exceeds 3× the estimated minimum viable line count, stop and simplify first.

**5. Maintain scope discipline.** Touch only what you were asked to touch; drive-by refactors
create noise and risk regressions. Do NOT remove comments you do not understand, clean up code
orthogonal to the task, refactor adjacent systems as a side effect, delete seemingly-unused code
without explicit approval, or add unrequested features because they seem useful. Match the existing
style of the file being modified — naming, error handling, import organization; consistency within
a file outranks personal preference. Anti-pattern: "while I was in this file I noticed…".

**6. Verify, don't assume.** Every task requires evidence of completion; "seems right" is never
sufficient. Tests passing means showing the test output; a build succeeding, the build output; a
file created, reading it back; behavior correct, the runtime evidence. For ad-hoc work without a
spec, define the goal as a testable assertion first — "done when X produces Y" — then verify it.
Anti-pattern: claiming tests pass without running them.

---

## 6. Output, language, and format

**Respond in the user's configured `conversation_language`.** Code, identifiers, paths, commands,
and flags stay in their original form.

**Non-English output must be native idiom, not English mapped word-for-word.** When
`conversation_language ≠ en`, every user-facing surface — chat, reports, README, docs, generated
sites, question text — MUST read as natural native prose. Translation-style calques (carry-over of
English syntax, metaphor, and figurative stock) are prohibited; native idiom is required. Chat uses
the colloquial native register, artifacts the clean native written register. Deliberately-coined
brand terms and established loanwords are not calques.

**Write non-ASCII payloads as native UTF-8.** Every tool-call payload carrying
`conversation_language` text — command strings, file content, question text — MUST be native UTF-8;
hand-authored `\uXXXX` escapes are PROHIBITED, because a malformed escape corrupts the payload into
a validation error and tends to be copied forward. On such a failure, re-author the text from its
intended meaning rather than repairing the escape.

**User-facing output is Markdown**; never display XML tags to users.

**XML is reserved for agent-to-agent data transfer.** Use semantic XML sections for structured data
exchange between agents; never surface XML structure in user-facing output.

**Never use time predictions in plans or reports.** Use priority labels (High / Medium / Low) and
phase ordering ("complete A, then start B"). Prohibited: "2-3 days", "1 week", "as soon as
possible".

---

## 7. Tools and command output

**Follow tool usage patterns optimized for accuracy and efficiency.** Read a file before editing
it. Locate before reading — find the file by pattern, find the line by content, then read that
region rather than the whole file. Use absolute paths and verify a path exists rather than
constructing it from an assumption. Prefer a targeted edit over rewriting a file, and a dedicated
tool over a shell equivalent. Retry safety is asymmetric: read-only and idempotent calls may be
retried, but a side-effecting one (write, commit, push, PR, deploy, external mutation) that fails
ambiguously requires observing current state first and retrying only when the effect is confirmed
absent — no success signal is not evidence the effect did not land. After three failures on one
operation, report the blocker.

**Maintain effectiveness without MCP servers.** Where one is unavailable, fall back to web search
and fetch for library documentation and established patterns, then continue — analysis quality must
not depend on MCP availability.

**Keep command output bounded**: quiet flags, targeted queries, or redirect-to-file with the exit
code and a bounded tail. A runtime output limit is a backstop, not the target.

**Prefer the quiet form of routine commands** — `--no-progress`, `-q`, machine-readable output plus
a targeted filter — not forms emitting spinners, banners, tables, or color noise. The same decision
bytes at a fraction of the context cost.

**Weigh session length as a cost axis.** One long session is cheaper than several short ones for
the same work: a fresh session re-pays the always-loaded prefix at write price where a continuing
one reads it from cache — provided it stays warm, since a long idle gap or an edit to the loaded
prefix reverts it to write price. Splitting a session is a cost to justify, not a default.
