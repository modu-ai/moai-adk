---
description: "Gate-measurement companion for factory-dispatch.md — the native Codex artifact contract, review lens selection table, the pre-merge settings-drift assertion, and the verification-load incident record"
paths: "**/factory-dispatch*.md,**/.claude/agents/moai/manager-lead.md,**/.claude/skills/moai/workflows/gtd.md"
---

# Factory Dispatch — Sync-Gate and Integration Measurements

> Companion of `factory-dispatch.md` (the always-loaded stub), split from `factory-dispatch-detail.md` by its per-file budget. The stub keeps every [HARD] rule; this file owns the gate measurements behind the stub's review and integration clauses: how native Codex artifacts bind the current head, how the leader picks review lenses, what `acquire` asserts about the caller's tree, and the verification-load incident record. Load when choosing review lenses for a dispatch, running the settings-drift assertion, or reviewing a lane's verification recipe. Siblings: `factory-dispatch-detail.md` (dispatch cycle and coordination) and `factory-dispatch-cards.md` (card lifecycle and traceability).

## Native Codex review evidence

**Native Codex review evidence.** Read the trusted Codex bot’s current-head review artifacts, not a successful check row. A completed code-review summary must bind the full head SHA, report no findings or unresolved findings, and agree with its clean receipt. Missing, failed, running, stale, malformed, or contradictory evidence withholds integration. Re-read the head and review artifacts before merge; preserve the separate functional and security checks.

## Review lens selection

`review` is not one thing. The leader picks the lenses from what the card actually changed, and states the choice in the dispatch so the sync gate runs that review rather than re-deriving it:

| Card touched | Lenses to instruct |
|---|---|
| Auth, session handling, input parsing, external calls, secrets, file/path handling | `--security` (add `--deep` when the surface is reachable by untrusted input) |
| Non-trivial logic across several files | `--deep` (adversarially verified multi-phase scan) |
| Whole-tree sweep rather than a diff | add `--repo` |
| Suspected over-engineering | `--lean` (advisory only; applies no fixes) |
| UI or design-system surface | `--design`, and `--critique` after the build |
| Small, local, low-risk diff | no flag — the default 4-perspective pass is enough |

## The pre-merge settings-drift assertion

The stub's `acquire` clause has a narrow subject: the **tracked** `.claude/settings.json` in the tree the lane is standing in when it takes the window. Twice, a card worktree has been found with that file modified in the working copy and no author anyone could name. Neither was caught by the merge window — the first surfaced because a merge happened to be refused, the second in a full sweep of every worktree nine days later. Authorship was closed as unattributable, which leaves detection as the only end that can be fixed: an unnoticed modification cannot be traced, but it can be caught before it rides a merge.

`acquire` is the right point because of cwd. The procedure has a lane run it from its own card worktree and only then enter the release worktree, so the tree standing under `acquire` is the tree about to be merged. Code does not enforce that ordering, so the report always names the tree it measured: measuring a different tree is acceptable, measuring one quietly is not.

The predicate:

```
git --no-optional-locks status --porcelain -- .claude/settings.json
```

Three properties of it are load-bearing.

- **`--no-optional-locks` is mandatory.** A plain `git status` takes an index WRITE lock for tens of milliseconds. A check that runs immediately before a merge, on a machine with several lanes on it, would otherwise manufacture the contention it exists to protect against.
- **The verdict is the match count, never an exit code.** Zero lines is a pass, one or more is drift.
- **A failed measurement is its own state.** `clean` / `drift` / `undetermined` are three states, not two with a fallback: an unmeasured tree that reports `clean` is precisely the nine-day silence, and the absence of a signal is not evidence of cleanliness.

On a hit: the working copy is preserved under the primary checkout's state directory (visible to the leader, who is not in the lane's tree), one row goes into `ledger.jsonl` beside it, and the path plus its sha256 and size are reported. Contents are never printed and never committed — the file is runtime-written and can carry tokens, absolute paths, and pane ids; promoting it into history needs a human secret-scan first.

**Nothing is ever restored, reverted, or deleted.** For the same reason: an automatic restore of a file holding machine-specific values is itself data destruction. Disposal is a human decision.

Two surfaces, and neither is redundant. Without the `acquire` precondition, running the check would be a social protocol — the same gap the announcement rule named about itself. Without `moai integration preflight [path]`, there would be no way to ask the question without taking a window.

Refusal is the only part that is opt-in (`workflow.settings_drift_gate.enabled`, default off, the same posture as the repository's other guards). Detection, preservation, the ledger row and the report run on every `acquire` regardless — the observed failure was nine days of nobody looking, not nine days of nothing being blocked, so gating the observation would remove the very thing the default-off posture rests on. Where refusal is on, `--allow-settings-drift` records the window anyway and stamps the bypass into the lock record; `--force` does not and must not, because it answers a different question ("take the window from a live holder") and one flag carrying both leaves the record unable to say which was meant. An `undetermined` verdict does not refuse — it is reported loudly and the window is recorded, matching every other guard's fail-open posture on uncertainty.

## Verification load incident record

Sessions share one machine as surely as they share one checkout, and verification is where that sharing goes wrong. Measured on a day when it did: load average reached 413, a neighbouring workspace's build took two and a half minutes, and its browser tests timed out — not because anything was wrong with them, but because four lanes were each running the full test suite at once and a full suite there takes five to ten minutes.

The never-spawn-background-load rule (normative statement lives in the stub) comes from the same incident's second cause: a verification recipe started eight spin loops to test behaviour under CPU contention and placed its kill line *after* the long test command. The agent finished before reaching it, and twelve spinners ran orphaned for thirty-seven minutes.

The same day supplied the reason contention and flakiness feed each other: a failing test left an unbounded spin-loop goroutine running, which burned a core for the remainder of that package's run and slowed every test after it. Load makes tests fail; failing tests can generate load.
