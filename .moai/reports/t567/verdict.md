# t567 verdict — a worktree vanished without a removal command: reproduction and condition-fixing

- Card: t567 (reproduction and condition-fixing only; the repair, if a cause is
  established, is a separate card)
- Worktree `.claude/worktrees/t567`, branch `WT-worktree-vanish`
- Base: `1d150a27d` (= `origin/develop` at entry). Local `develop` was
  `5ddccacc9`, 30 commits ahead of `origin/develop`; this branch has not
  absorbed it. The integration window absorb is still owed.
- Unpushed. The lane does not push.

## Claim

1. **The named hypothesis is falsified for all four combinations the card
   specifies.** `EnterWorktree(path:)` switching does NOT remove the tree it
   switches away from — not in combination A (the shape t528 was in), not in B
   (no `ExitWorktree` at all), not in C (tree created by raw
   `git worktree add`), and not in D (the no-switch control). Every subject
   survived on both axes, with its branch ref and its unpushed commit intact.

2. **A worktree nevertheless vanished during this run, unprompted.**
   `.claude/worktrees/agent-abbf2b88afd6960f2` was present in the opening
   inventory — `locked`, at `1d150a27d`, on branch
   `worktree-agent-abbf2b88afd6960f2` — and was afterwards absent from
   `git worktree list`, absent from disk, and its branch ref was gone from
   `git show-ref`. No session issued a removal command for it. The symptom
   matches the card's t528 report.

3. **The vanish signature is a full removal, not a pruned stale entry.**
   `git worktree prune` drops an administrative entry and leaves the branch;
   here the branch ref went too. Directory plus branch is the signature of a
   deliberate remove, which narrows the actor set.

4. **`moai worktree clean --json` removes no live worktree, but it is not the
   read-only reporter its help describes.** Its help states `--json` "Report
   every non-protected worktree and its state as JSON; removes nothing". It
   emitted no JSON at all — only the banner `✓ Cleaned stale worktree
   references` — and it performs a prune. A second invocation removed nothing,
   which is how claim 2's subject is excluded from being its victim: by then
   the directory was already gone and only the dangling entry remained to
   prune.

5. **Two MoAI sweeps can remove an L1 card worktree, and the doctrine does not
   model them.** `prMergeCleanup` (`internal/cli/session_worktree_prmerge.go`)
   runs at `moai session register` and `moai session list`, before the
   subcommand's own work. Both it and `moai worktree clean` enumerate through
   `git worktree list --porcelain`, which lists every worktree including the
   L1 `.claude/worktrees/*` trees that `moai worktree done` disclaims. Both are
   gated by `workflow.worktree.auto_cleanup`, which is `false` in this
   repository, in the shipped template, and in the compiled default.

6. **Their dirty guard cannot see unmerged commits.** `worktreeIsDirty`
   (`internal/cli/session_worktree.go:680`) is `git status --porcelain`
   non-empty — uncommitted changes only. A worktree holding committed,
   unpushed, unmerged work with a clean status reads "not dirty" and is
   eligible for removal. That is exactly the state the card's disposal rule
   ("an unpushed branch's worktree is the work's only instance") exists to
   protect.

7. **The shipped template states the opposite of the source about a
   removal-gating key.** `internal/template/templates/.moai/config/sections/workflow.yaml:53-55`
   says "auto_merge / auto_cleanup: declared but not read — no code path
   consumes them (reserved)". `internal/config/types.go:626-628` says
   AutoCleanup "is read by the two auto-cleanup paths … gating worktree
   removal". The source comment is the accurate one; the distributed file tells
   users a key that gates deletion is inert.

## Evidence

| Claim | Command and observation | Path |
|---|---|---|
| Opening inventory | `git worktree list` → 241 rows, one matching `t567` | `repro/00-inventory-before.txt` |
| D (control: create via `EnterWorktree(name:)`, commit, `ExitWorktree keep`, no switch) | subject `t567-d`, commit `05a4dcb14`, `git rev-list --count develop..HEAD` → 1. Probe after exit: `axis1: PRESENT`, `axis2: PRESENT` | inline probes, transcript |
| A (t528's shape: create via `name:`, commit, switch via `path:`, `ExitWorktree keep`) | subject `t567-a`, commit `aeab4bac1`. Probe immediately after the switch: PRESENT/PRESENT. Probe after the exit: PRESENT/PRESENT. `git show-ref --verify refs/heads/worktree-t567-a` → `aeab4bac157f90ebb4c8d0ec521d69617a30619d` | inline probes, transcript |
| B (same, but `ExitWorktree` never called) | subject `t567-b`, commit `7ae0f48b7`. Probe after the switch: PRESENT/PRESENT | inline probes, transcript |
| C (created by raw `git worktree add`, then switched away and exited) | subject `t567-c`, commit `99b0c4063`. Final probe: PRESENT/PRESENT | inline probes, transcript |
| All four subjects at the end | six-path loop over `t567`, `t567-a`, `t567-b`, `t567-c`, `t567-d`, `t567-target` → every row `axis1=PRESENT axis2=PRESENT` | transcript; `repro/10-inventory-after-combinations.txt` |
| The vanish | `comm -23` of the sorted opening and closing path lists → exactly one path missing: `.claude/worktrees/agent-abbf2b88afd6960f2`; five added, all mine | `repro/11-disappeared.txt`, `repro/12-appeared.txt` |
| The vanish is total | `ls -d …/agent-abbf2b88afd6960f2` → `No such file or directory`; `git show-ref \| grep abbf2b88` → no ref | transcript |
| Its opening state | `1d150a27d [worktree-agent-abbf2b88afd6960f2] locked` | `repro/00-inventory-before.txt` |
| `clean --json` emits no JSON | `moai worktree clean --json` → exit 0, 296 bytes, banner `✓ Cleaned stale worktree references`, stderr 0 bytes | `repro/20-clean-json-run1.txt` |
| `clean --json` removes no live tree | second invocation, inventory diffed immediately before and after → zero paths disappeared | `repro/21-clean-json-run2.txt`, `repro/22-clean-run2-disappeared.txt` |
| Sweep enumeration and gate | `internal/cli/session_worktree_prmerge.go:6-7,141-151`; `internal/cli/worktree/clean.go:538` — both read `git worktree list --porcelain`; `prMergeCleanup` returns early unless `cfg.Workflow.Worktree.AutoCleanup` | source read |
| Gate is off here | `.moai/config/sections/workflow.yaml:149` `auto_cleanup: false`; template `…/workflow.yaml:59` `auto_cleanup: false`; `internal/config/defaults.go:874` `AutoCleanup: false` | source read |
| Dirty guard scope | `internal/cli/session_worktree.go:680-686` — `git status --porcelain` non-empty | source read |
| Template/source contradiction | template `workflow.yaml:53-55` vs `internal/config/types.go:626-628` | source read |

## Baseline-attribution

- Every probe ran in this tree in this run, against this machine's live
  worktree set. Tools: `claude 2.1.269`, `git 2.50.1 (Apple Git-155)`.
- Both existence axes were measured per subject: `git worktree list
  --porcelain` matched with `grep -x "worktree <abs-path>"` (exact-line, so
  `t567` cannot be satisfied by `t567-a`), and `ls -d`/`[ -d ]` on the absolute
  path. Neither axis alone was accepted.
- The inventory diffs are `comm` over sorted first-column path lists taken from
  the same command (`git worktree list`) at the two points compared.
- Source claims are read at this tree's HEAD base `1d150a27d`. No installed
  binary was cited for them.
- `moai worktree clean --json` was invoked as the installed `moai` on PATH; its
  build was not pinned to this tree, so claim 4 is attributed to that installed
  build, not to the source at `1d150a27d`.

## Gaps

- **The actor behind claim 2 is not established.** Three candidates are named
  and none is observed acting: (i) the Claude Code runtime's own cleanup of an
  `Agent(isolation: "worktree")` tree — the Agent tool's documentation states
  such a worktree is "auto-cleaned if unchanged", and this subject was an
  `agent-*` tree sitting unchanged at its base, which fits; (ii) the session-end
  keep/remove prompt of whichever session owned it; (iii) another live session.
  The subject was already gone when first observed missing, so nothing was
  watched in the act.
- **The subject class differs from t528's.** What vanished here was an
  `agent-*` runtime worktree; t528's was a named card worktree. The shapes
  match, the classes do not, so this run does not establish that the same actor
  took t528's tree.
- **Session end was not observed.** Combination B's subject was left with no
  `ExitWorktree` deliberately, but this session has not ended, so what the
  exit-time prompt does to it is unmeasured. That is the one branch of the
  card's matrix that a single in-session run cannot close.
- **The sweeps were read, not run.** `prMergeCleanup` was never executed with
  `auto_cleanup: true`; claims 5 and 6 rest on reading the code, not on
  observing a removal. Whether a merged L1 card tree would in fact be removed
  under that toggle is untested — deliberately, since the test destroys trees
  belonging to live cards.
- **`auto_cleanup` history was checked only in this repository's committed
  config.** `git log -S` puts the `true → false` flip at `8ff3e0823`
  (2026-08-24), before the t528 event. Whether the working copy actually held
  `false` on 2026-09-08 was not measured — `.moai/config` is wiped and
  redeployed by `moai update`, so the committed value is not proof of the
  value in force that day.
- **No timeline was reconstructed for the vanished tree.** Its owning session,
  its lock reason, and the moment of removal were not recovered; transcripts
  were not scanned.

## Residual-risk

- **The disposal rule still rests on an unverified premise.** "A worktree
  disappears only when I remove it" was contradicted once by t528 and once
  again in this run. Nothing in this card repairs that; the cheap guard below
  converts the premise into a measurement, which is not the same as removing
  the actor.
- **A removal that takes the branch too is unrecoverable by the usual reflex.**
  Where only the directory goes, the branch still holds the commits; here the
  ref went with it. Had that tree carried unmerged work, `git worktree list`
  would have shown nothing to recover and the branch name would have resolved
  to nothing.
- **The `--json` defect hides its own blast radius.** A flag documented to
  remove nothing that prints a "Cleaned" banner invites exactly the reflex this
  card is about: running it to *look*, and mutating state instead. This run
  measured that it removed no live tree; it did not establish that it never
  can.
- **The sweeps' eligibility test is merge state, and merge state is not
  safety.** A branch merged into the base can still be the only copy of
  evidence, untracked artifacts excepted by the ignored-content check.

## Cheap guard (the card's free by-product)

Tree existence is currently assumed at two points where it is cheap to measure.
Both axes, absolute path, exact-line match:

```
git worktree list --porcelain | grep -x "worktree <abs-path>"   # registry
[ -d "<abs-path>" ]                                             # disk
```

Take the reading before entering an integration window and again before sending
a completion report. A disagreement between the two axes is itself the finding:
registry-present with disk-absent is a dangling entry, and disk-present with
registry-absent is a tree git no longer manages.

## Disposition of the by-products (lead ruling, 2026-09-12)

- Claim 4 filed as **modu-ai/moai-adk#1704**; body kept verbatim at
  `issue-clean-json.md`.
- Claim 7 filed as **modu-ai/moai-adk#1705**; body kept verbatim at
  `issue-autocleanup-comment.md`.
- Claims 5 and 6 (the sweeps' dirty guard cannot see unmerged commits) are a
  separate card, **t673**, deliberately kept out of both issues.
- Experiment trees `t567-a`, `t567-c`, `t567-d`, `t567-target` disposed here by
  `git worktree remove` (exit 0 each, absence confirmed on disk). Their branches
  were NOT deleted — `git worktree remove` leaves them, and nothing asked for
  their removal. **`t567-b` is preserved to batch close**: it is the subject
  left with no `ExitWorktree` on purpose, and whether it survives this session's
  end is the one datum a single in-session run cannot otherwise produce.

## Next candidates (the card asks that they be named, not that they be chased)

1. The Claude Code runtime's automatic cleanup of an unchanged
   `Agent(isolation: "worktree")` tree — the documented behaviour, and the best
   fit for what was observed here.
2. The session-end keep/remove prompt, answered `remove`, in a session other
   than the tree's user.
3. `prMergeCleanup` at `moai session register` / `moai session list`, in any
   window where `workflow.worktree.auto_cleanup` was true.
4. `moai worktree clean` invoked by another lane, which is ungated by that
   toggle.

🗿 MoAI

---

## Continuation — 2026-09-13 (lane-3, resuming from this evidence)

### Dispositions of the findings above (all verified today, against origin/develop `6916f9e83`)

| Finding | Disposition | Verified by |
|---|---|---|
| 4 — `clean --json` emits no JSON and prunes | REPAIRED: card t681, merged `ae980ef2d` — `--json` routes to the inventory in every flag combination; the reporting path no longer prunes; regression tests pin it | `git merge-base --is-ancestor ae980ef2d origin/develop` → true; t681 verdict `.moai/reports/t681/verdict.md` |
| 5 + 6 — the two auto-cleanup sweeps reach L1 trees; dirty guard is porcelain-only | REPAIRED: card t673, merged `6916f9e83` — shared unpushed predicate in both paths; committed-unpushed and detached-HEAD trees observed preserved (real-git RED→GREEN), pushed control still removable | `git merge-base --is-ancestor 6916f9e83 origin/develop` → true; t673 verdict `.moai/reports/t673/verdict.md` |
| 7 — template says auto_cleanup is a reserved unread key | REPAIRED twice over: prose fixed on develop (`390d71753`, t655 family — per lead's re-measurement) and drift guard landed card t682, merged `d84458994` (`internal/config/workflow_key_honesty_test.go`) | `git merge-base --is-ancestor d84458994 origin/develop` → true; t682 verdict `.moai/reports/t682/verdict.md` |
| 1, 2, 3 — hypothesis falsified; live vanish observed; full-removal signature | STAND (this card's own conclusion) | unchanged |

### Continuation probes (2026-09-13, this session)

The four experiment subjects and the vanished tree, re-measured one day after the reproduction run:

| Subject | Branch ref | Directory |
|---|---|---|
| `worktree-t567-a` | PRESENT `aeab4bac1` | ABSENT |
| `worktree-t567-b` | PRESENT `7ae0f48b7` | PRESENT |
| `worktree-t567-c` | PRESENT `99b0c4063` | ABSENT |
| `worktree-t567-d` | PRESENT `05a4dcb14` | ABSENT |
| `agent-abbf2b88afd6960f2` (claim 2's subject) | ABSENT | ABSENT |

### New narrowing: the t528/agent-abbf actor is distinguishable by signature

Between the reproduction run (2026-09-12) and this probe (2026-09-13), the directories of subjects a, c, and d disappeared while their branch refs survived. This matches the known disposal signature of a session-end keep/remove disposal and a plain `git worktree remove` — the directory goes, the branch ref is left behind (recorded in project memory: "폐기가 브랜치까지 안 지움"). The card's subject — agent-abbf2b88, and before it t528 — lost DIRECTORY AND BRANCH TOGETHER.

The two signatures differ, so the actors differ:

- the disposer that took a/c/d leaves refs — NOT the actor this card hunts;
- the t528/agent-abbf actor removes the ref as well — i.e. it performs a branch deletion on top of (or instead of) a worktree removal. Remaining candidates: `moai worktree done` / explicit branch delete, the Claude Code runtime's `Agent(isolation: "worktree")` auto-clean (documented "auto-cleaned if unchanged" — the subject was an unchanged agent-* tree at its base), or an external actor. The session-end disposer drops out of the candidate set.

This is the continuation's one new fact; the actor is still not observed in the act. The residual gap below stands.

### Residual gap (unchanged)

The actor behind claim 2 is not established. What is new: the session-end disposal path is now excluded by signature, leaving `moai worktree done`-class full disposal, the runtime's Agent-isolation auto-clean, and external actors. A next experiment, if this card is ever reopened: park an unchanged dummy `agent-`-prefixed tree and a changed one, and diff which disappears.
