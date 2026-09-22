# t181 — serialize release-tree integration by announcement, not by probe

Tier S doctrine edit to `kanban-dispatch.md` § Integration into the release branch is
self-served, plus its template mirror. Branch `WT-serial-integration`, based on
`origin/main` (independent of the v3.1.3 batch).

---

## Claim

1. The existing bullet let an empty `MERGE_HEAD` stand as proof that the release worktree
   was free. It is not proof, and the rule now says so.
2. Two [HARD] clauses replace the gap: serialize by announcement (lane announces → lead
   broadcasts the hold → no other session enters until the completion report), and re-read
   `HEAD` immediately before the commit and again before the push.
3. Local rule and template mirror stay byte-identical; the embedded copy was rebuilt.

## Why the old bullet was insufficient

It read: *"On entry, confirm no merge is in progress — `git rev-parse -q --verify MERGE_HEAD`
must print nothing; a lane that finds a merge in progress exits, waits, and retries."*

That is a probe of one state, offered as a test for another. `MERGE_HEAD` is absent while a
merge is in progress in every way that matters to a second lane: after `git merge --abort` and
before the retry, and before the resolving lane has staged anything. A lane reading the
silence as "free" enters, and neither side learns of the overlap until one commits.

The failure this repairs was observed in this project's own batch integration: one lane's
repair commit moved the release `HEAD` while another lane was mid-resolution. Nothing detected
it except the re-read of `HEAD` immediately before committing — which the standing contract
(`AGENTS.md` §2) already required everywhere, and which this section now names at the point
where it has actually paid.

## The change

```
- One integrating session at a time — and an empty MERGE_HEAD does not establish that.
    [HARD] Serialize by announcement, not by probe.
    [HARD] Re-read HEAD immediately before the commit and again before the push.
```

The probe is kept, demoted to a last check rather than the first. The `AGENTS.md` §2 clause is
cross-referenced, not restated, so the always-loaded surface carries the scoping sentence only.

## Evidence

| Command | Observed |
|---|---|
| `git diff --numstat` | `5 1` on both the rule and its mirror — same edit, no drift |
| `diff <local> <template>` | identical (they were identical before the edit too) |
| byte size, before → after | 25,915 → 27,002 (**+1,087**) |
| `make build` | exit 0; `catalog.yaml` unchanged (no new files) |
| `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -count=1` | `ok … 20.298s` (mirror parity + neutrality) |
| `go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$'` | PASS |

## Always-loaded cost statement (`rule-authoring.md` § The statement duty)

The edit grows an always-loaded file by **1,087 bytes**, over the 1,000-byte threshold, so the
duty fires.

- **(d) Scope first.** The content cannot move to a `paths:`-scoped companion. A lane learns it
  is integrating from a dispatch, not from a file it opened; a rule keyed to file paths would
  never load at the moment the decision is made. The same reasoning already keeps
  `kanban-dispatch.md` always-loaded.
- **(c) Non-invoking cost.** Every session pays these 1,087 bytes on every turn and again after
  every `/clear`, including sessions that never integrate anything — most of them. That is the
  real price. It is accepted because the failure it prevents is silent and expensive: two lanes
  building on divergent trees, discovered only at commit, in the batch's single serialization
  point. A lane cannot look this rule up after the fact; by then it has already entered.
- **Sizing.** Held to one bullet and two clauses. The incident record and this reasoning live
  here, not in the rule; the `AGENTS.md` §2 obligation is cross-referenced rather than copied.

## Baseline-attribution

Measured by me in `.claude/worktrees/t181` on `WT-serial-integration`, based on `origin/main`
at `76b2c4ece`, against the tree carrying the final diff.

## Gaps

- **No full-suite run.** Two markdown files and a rebuild; `internal/template` (which owns
  mirror parity and neutrality) and the budget guard are the packages that can observe this
  change, and both ran. CI on the PR is the full-suite measurement.
- **The always-loaded token figure was not captured on this branch.** The budget test on
  `origin/main` reports only on overflow — it emits no informational line — so the PASS is the
  evidence, and the byte delta above is the sized quantity. The token figure exists on the
  release branch's version of that test, which is not this base.
- **No runtime rehearsal.** The doctrine was not exercised by staging two concurrent lanes; it
  is a written obligation, and nothing here mechanically enforces it.

## Residual-risk

- **Announcement is a social protocol with no mechanical guard.** A lane that skips the
  announcement is not blocked by anything — the probe it would still run is exactly the
  insufficient one this card is about. A mechanical holder-lock in the release worktree would
  close that, and is not in scope here.
- **The hold depends on the nudge channel reaching every lane**, which is absent on native
  Windows and under some providers (`cross-session-messaging.md` § Availability constraints).
  Where it is absent, the serialization degrades to the probe it replaces.
