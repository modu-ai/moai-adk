# t156 verdict — CC 2.1.236 Tier 2 docs sync (GD-1 / GD-3 / GD-4)

Card: t156 (Class B — plan skipped, run → sync in-lane)
Base: release/v3.1.1 @ 18e0de51f (local integration commit; origin is still at ca7e15fa2)
Worktree: `.claude/worktrees/t156`, branch `WT-cc236-docs`
Source report: `.moai/research/cc-update-2.1.235-to-2.1.236.md` (gitignored — read from the primary
checkout, never copied or committed)

## Claim

Three Tier 2 documentation gaps from the CC 2.1.236 delta are closed across three always-loaded /
lazy-companion rule files plus their three template mirrors. Tier 1 was 0. No Go source changed.

## Evidence

### Base anchoring

The worktree was created from `origin/main` (the `worktree.baseRef` default is `fresh`; no
`baseRef` key is set in `.claude/settings.json` or `settings.local.json` — measured), so it did NOT
start on the card's base. Corrected by fast-forward, which was legal because the created HEAD was an
ancestor of the base:

    $ git merge-base --is-ancestor HEAD 18e0de51f   -> true
    $ git rev-list --count HEAD..18e0de51f          -> 81
    $ git rev-list --count 18e0de51f..HEAD          -> 0
    $ git merge --ff-only 18e0de51f                 -> exit 0
    $ git rev-parse --short HEAD                    -> 18e0de51f
    $ git branch -m WT-cc236-docs                   -> WT-cc236-docs

`git reset --hard` was attempted first and is DENIED by this session's permission settings. The
fast-forward reaches the identical tree without it. No peer was asked to run the denied command.

### Mirror parity BEFORE editing (the report left this unverified — gap #3)

The source report verified the three mirrors *exist* but explicitly did not diff their bodies. All
three were byte-identical to their live counterparts before any edit:

    $ diff .claude/rules/moai/workflow/cross-session-messaging.md   internal/template/.../cross-session-messaging.md   -> PARITY-OK
    $ diff .claude/rules/moai/workflow/kanban-dispatch.md           internal/template/.../kanban-dispatch.md           -> PARITY-OK
    $ diff .claude/rules/moai/workflow/goal-directive-detail.md     internal/template/.../goal-directive-detail.md     -> PARITY-OK

Because parity was measured as byte-identical, the mirrors were produced by `cp` of the edited live
file. That is provably the same edit, not a blind copy — the general "never `cp` a mirror" caution
exists because templates can carry deliberate neutralization, which these three demonstrably did not.

### docs-site 4-locale sweep (the report left this unverified — gap #2)

`docs-site/content` was grepped for each of the three changed statements. None of them appear there:

- Tool-surface / send options: only 4 hits, all one summary line per locale in
  `multi-llm/kanban-mode.md` naming `ListAgents` / `SendMessage` and `crossSessionInbound`. It does
  not enumerate send parameters, so `notify_when_idle` has nothing to update.
- Stop-paths: docs-site mentions `crossSessionInbound` only; the four-way stop-path table that GD-3
  extends exists solely in the rule file.
- Native `/goal` check-in: `grep 'GOAL_CHECKIN|30분|30 minutes|30分|check-in|체크인'` over
  `docs-site/content` returns only unrelated matches (onboarding "30 minutes", `moai session purge`
  heartbeat, scheduled-task jitter, a SPEC example string). Zero hits on the goal check-in.

Conclusion: no same-PR locale sync is owed. Scope of this claim: `docs-site/content` only.

### The edits

GD-1 — `notify_when_idle` (was 0 matches repo-wide; now 4 live + 4 mirror):

- `cross-session-messaging.md:9` — the two-tool surface sentence now names the opt-in request and
  points at the new section.
- `cross-session-messaging.md:23` — version floor `v2.1.236+` appended to the existing Versions
  bullet, alongside the existing 2.1.224 / .225 / .232 floors. The macOS/Linux constraint needed no
  new caveat: it already matches the OS bullet.
- `cross-session-messaging.md` § An idle notice is a scheduling hint (new H2, line 84-88) — mechanics
  (opt-in per send, one-shot, replaces polling) plus the [HARD] guard.
- `kanban-dispatch.md:59` — one paragraph appended to § The delegation channel is the queue.

  [HARD] guard, present in BOTH files (verified by grep): a session goes idle when it finishes, when
  it stops at a permission prompt, and when it dies; the notice cannot separate those, so it says
  *when to read the evidence* and nothing about what the evidence says. Adopting it as a completion
  signal would convert the [HARD] read-don't-trust rule into an unobserved completion claim
  (`verification-claim-integrity.md` §1.1 surface 1).

GD-3 — burst refusal:

- `cross-session-messaging.md` § Configuration surface — a paragraph after the four-row table naming
  inbox burst capacity as the fifth thing that stops a message, and the first that is not a setting.
  Framed on Factory fan-out (`moai cc -f <N>`), since a lead nudging N lanes in one turn is exactly
  the burst shape. Records that the refusal is up-front and visible (an improvement over a silent
  drop), and that it costs the board nothing because delegation rides the queue.
- Also folded into the `kanban-dispatch.md:59` paragraph above, where the dispatcher will meet it.
- The § Availability trap paragraph was deliberately NOT touched: it enumerates causes of a message
  *silently not arriving*, and a burst refusal is visible, so it does not belong in that list.

GD-4 — native `/goal` check-in backoff:

- `goal-directive-detail.md:81` — "waiting 30+ minutes … checks in" replaced with the 30m → 1h → 2h
  backoff plus the refined trigger framing ("an idle session whose goal is parked behind
  long-running background work … rather than waiting for the user to come back"), attributed to
  2.1.236 as a refinement of the original single check.
- SCOPE GUARD HELD. The diff for this file is exactly one line (`git diff --numstat` = 1 added,
  1 removed). Line 83's no-transposition prohibition is untouched, and a repo-wide grep for the
  check-in wording returns hits in the native paragraph only — live and mirror, nowhere else. No
  `/moai goal` surface (`goal-directive.md`, `.claude/skills/moai/workflows/goal.md`) was modified.

### Verification

    $ git status --porcelain            -> exactly 6 modified files (3 live + 3 mirror), nothing else
    $ git diff --numstat -- .claude/rules/
        12  2  cross-session-messaging.md
         1  1  goal-directive-detail.md
         2  0  kanban-dispatch.md

    $ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/... -run 'Leak|Neutrality|Mirror|Parity'
        ok  github.com/modu-ai/moai-adk/internal/template  1.801s

    $ make build                         -> exit 0 (templates re-embedded)
    $ git diff --stat -- internal/template/catalog.yaml
        (empty — rules are not catalog entries, so no catalog mirror-commit is owed)

    $ go test ./internal/template/... ./internal/config/... -count=1
        ok  internal/template            51.387s
        ok  internal/config              18.175s
        ok  internal/config/atomicfile    0.847s
        ok  internal/config/toolpolicy    1.853s

    $ go vet ./internal/template/... ./internal/config/...   -> exit 0

Always-loaded budget (two of the three files are always-loaded, so this is the load-bearing check):

    $ go test ./internal/config/ -run TestAlwaysLoadedTokenBudget   -> PASS
    measured after the edits: total=71207  budget=76000  headroom=4793  entries=17

The headroom figure came from a scratch probe (`zz_t156_headroom_test.go`) calling the guard's own
`measureAlwaysLoaded`, because the shipped guard passes silently without printing the total. The
probe was deleted after measuring; `git status` above confirms it left no residue.

## Baseline-attribution

Every number above is from a command run in this worktree at `18e0de51f` + these six edits, in this
session. The pre-edit parity result is from the same worktree before the first Edit call.

## Gaps (NOT observed)

- `notify_when_idle` was not exercised. Its semantics are taken from the changelog bullet, which the
  source report flagged as changelog-only (absent from the fetched docs pages). The [HARD] guard does
  not depend on those semantics — it derives from MoAI's own read-don't-trust rule — but the
  mechanics sentence (opt-in, one-shot, no polling) is changelog-sourced and untested.
- The full suite was not run (card prohibition, and CLAUDE.local.md §4). Only `internal/template`
  and `internal/config` were run. Cross-cutting breakage outside those two packages would not have
  been seen here; CI on the eventual PR is the verdict.
- GD-2 (`ANTHROPIC_DEFAULT_MODEL` absent from `internal/config/envkeys.go`) is NOT in this card and
  was not touched. It changes Go source and needs its own run-phase card, per the source report.
- `SPEC-GOAL-STOPFAILURE-CLEAR-001` was not opened. The source report flagged a possible scope
  overlap with GD-4 and left it unresolved; this card did not resolve it either. If that SPEC also
  edits the native paragraph, these two edits will collide on `goal-directive-detail.md:81`.
- The docs-site sweep covered `docs-site/content` only. Layouts, data files, and shortcodes were not
  grepped (no rule text is expected there, but it was not measured).

## Residual-risk

1. Always-loaded headroom is now 4,793 tokens against a 76,000 cap. This card spent part of a budget
   that a separate diet effort has been actively defending. The guard passes, but the two files this
   card grew are always-loaded, so the cost is paid on every session of every user.
2. The base is a LOCAL commit (`18e0de51f`); origin is still at `ca7e15fa2`. Anything downstream that
   resolves the base from origin will not find it.
3. `notify_when_idle` is changelog-only. If the documented shape differs once the docs page catches
   up, the mechanics sentence needs revision (the [HARD] guard does not).

## Scope compliance

- No push. Nothing written to the primary checkout. Full suite not run.
- Six files changed, all of them the intended three rules and their three mirrors.
