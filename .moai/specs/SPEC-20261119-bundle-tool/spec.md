---
id: SPEC-20261119-bundle-tool
title: "moai bundle: assemble, verify and describe an M-Flow bundle from READY heads (plan)"
status: draft
revision: 2
created: 2026-11-19
---

# moai bundle: assemble, verify and describe an M-Flow bundle from READY heads (plan)

## Intent

Card C-119, slice 1: the plan and this SPEC, no code. Under M-Flow (user decision 2026-11-19) lanes push `spec/<date>-<slug>` branches, a head is audited, the lane writes one `LANE-READY-INTEGRATION` card note, and the leader integrates several READY heads into one pull request. That integration is done by hand today. This SPEC fixes what a `moai bundle` command takes over and what stays the leader's decision, so each later slice is a small SPEC of its own.

**The hand procedure, measured.** Read from the bodies of the merged bundle pull requests #3225 (bundle 2, merge `bdb4d3db8c`, 3 heads, 17 commits) and #3227 (bundle 4, merge `d25ecb1ba1`, 4 heads, 13 commits), `gh pr view <n> --json body`:

1. An integration branch is rebuilt from the tip with one merge commit per audited head; its SHA is pinned.
2. The pinned SHA is verified locally: `cargo build --locked --offline`, the whole `cargo test --locked --offline --lib`, a list of `--test` targets, `cargo fmt --check`, `cargo clippy --locked --offline --all-targets -- -D warnings`, `moai-cli trailer-check origin/v2026.11.0..HEAD`, `moai-cli spec check --all`, `python3 -m unittest` over a list of modules, and a check that every AC of each bundled SPEC has a line in its `## Result`.
3. One pull request is opened. Its body is: the line `audit-tier: N <paths>`, one `Moai-Spec:` line per SPEC, a table of heads (SPEC, card, head, audit evidence), an `Integration:` line (the integration branch, how it was built, the pinned SHA and the size of the diff; its parts differ a little between the two bodies), `## Self-Audit` (plan, red, mutation, host, probes, not done locally) and `## Result` (each command with its exit and count, then one line per SPEC).

**What goes wrong, measured.** None of the leader's scripts the card names (`spec_result_check.py`, `wait_pass_then_merge.sh`, `merge_when_green.sh`) is in the repository: `git grep -l -E 'spec_result_check|wait_pass_then_merge|merge_when_green' origin/v2026.11.0` finds 0 files, so every bundle is assembled again. The card C-119 body records bundle 5: 19 heads held for about two hours because six heads each failed one check and every failure stopped the whole bundle; rebuilding without the failing heads took minutes once done. The same body records two tool faults: a Python test listed as a cargo target, and the hand SPEC Result checker's two false positives that the product `spec check` does not give. One more fault is partly re-measured: the body of #3225 says its first bundle head failed at clippy on `tests/workflow_size_budget.rs:104`, and `gh run list --branch m-flow/bundle-20261119-2` shows `ci failure` on that head `eba078d26b` (the failing step was not read from the CI log). Three more are from the leader's own records and were not re-measured for this SPEC: a merge driver registered in the clone made a local `git merge-tree` clean where GitHub conflicted (#3162, 2026-11-17); `cargo test --lib <filter>` missed a repository-wide scan that failed on Linux (#3226, 2026-11-19); a head was merged on an audit note although its tip-built `trailer-check` had not passed (card C-60).

**The READY line, measured.** Counted through note 2931 (2026-11-19T08:47:51Z) over C-1 … C-121 (C-121 was the highest card at the snapshot), by `moai card show C-<n>`, taking the notes with `note_id` up to 2931 whose first line starts with `LANE-READY-INTEGRATION`: 31 notes; 29 read under the grammar of D-6 (`head=` required, `spec=` optional, text after a space allowed; 28 carry `spec=`); 2 do not (`… agent-1 C-33 S2a — …`, `… agent-1 C-60 branch … head <sha> …`). 19 cards have a marker line (17 have a readable one); 10 cards have two or more readable lines, each set with different heads: a repair moves the head, so the newest line must win. The ledger keeps growing, so a later count is larger (through note 2949, at 08:54Z, 33 notes and 31 readable); the figures above are those of the pinned snapshot.

**The verdict line on cards, measured.** Through note 2931, 81 card notes have a first line that starts with `AUDIT` and contains ` head `; 63 of them read under `gate::parse_verdict` (a 40-hex head, a 40-hex base, a verdict word without a space) and 26 start with `AUDIT #`. A card note is written `AUDIT C-<n> … head <sha> base <sha> — verdict: PASS`, so `gate::verdict_line` (which needs `AUDIT #`) accepts none of the verdicts that name a READY head. Of the 29 readable READY lines, 18 have a `PASS` note on the same card naming exactly that head (read by `parse_verdict` alone) and 11 have none (mostly repair heads; of the 7 heads of #3225 and #3227, `e7427dd756`, `6313f800e3` and `5cc7759186` have none).

**Parts that exist and are reused, not written again.** `gate::tier_of` and `gate::parse_body_tier` (`src/gate.rs:306`, `:280`: the audit tier from the changed paths, the body line rule); `gate::parse_verdict` (`src/gate.rs:346`: the verdict first line; `gate::verdict_line`, `:341`, is not reused, see D-8); `trailers::check` (`src/trailers.rs:361`, behind `trailer-check`); `spec_doc` (`SpecDoc.result`, `Outcome`, the `closure` rule in `src/spec_doc.rs`: the status rule is part of `closure`, so a Result-covers-AC check without it is a small new function over the same parse); the Self-Audit checker (`self_audit::check`, `src/self_audit.rs:29`, reached from `pr-self-audit-check`, `src/main.rs:2513`, and from the gate's `body_check`, `src/gate_collect.rs:685`); `moai heavy-slot-acquire` (`src/heavy_slot.rs`); `identity::scrub_git_env` (`src/identity.rs:44`, `pub(crate)`: usable by a new module of this crate).

## Non-goals

- Opening, editing or merging a pull request, pushing a branch, changing a ruleset or the protected branch: they are outward-facing and stay the leader's (Decisions D-1). CI, the independent audit and the ruleset are untouched; no step of `moai bundle` replaces or skips one.
- Choosing which READY heads go into a bundle, or accepting a head's audit: the caller names the heads; the tool reports evidence found or missing.
- Writing a card note or any other ledger row: the tool reads the ledger.
- Merging more than the heads it was given, resolving a conflict, or editing a head: a conflict names the head and the files and the head stays out.
- A configuration file for the check list in this slice: the list is data in the code with a test that pins it (D-4); a file is a later card if the leader wants to change it without a build.

## Requirements

These are the contract every slice keeps. The detail (the READY grammar, the stable code of each per-head finding, the check list) is stated again, as requirements with tests, in the SPEC of the slice that implements it (Plan).

- REQ-1: When `moai bundle ready` is run, `moai bundle ready` shall list, for each card whose newest `LANE-READY-INTEGRATION` note names a head, the card, the note number, the head and the SPEC ids named by the `Moai-Spec` trailers of that head's commits, and write nothing.
- REQ-2 (REQ-1): IF the first line of a `LANE-READY-INTEGRATION` note does not read under the grammar of D-6, THEN `moai bundle ready` shall list the note as `unreadable` with its card and note number and use no head from it.
- REQ-3: When `moai bundle check` is given heads, `moai bundle check` shall judge every head on every per-head check and report all findings together, each naming the head, its card and the check.
- REQ-4 (REQ-3): IF a head fails a per-head check (not reachable from `origin`, already in the tip, no `PASS` verdict line (D-8) naming exactly that head when its changed paths set an audit floor of 1 or more (D-9), trailers, a SPEC error, an AC with no outcome line in `## Result`, or a conflict with no merge driver running), THEN `moai bundle check` shall report the head with a stable code and, for a conflict, the files.
- REQ-5: When `moai bundle build` is given heads in an order, `moai bundle build` shall create a new branch and a new worktree from the tip, merge each head into it with one merge commit, and print the pinned SHA.
- REQ-6 (REQ-5): IF a head's merge conflicts, THEN `moai bundle build` shall leave the branch without that head, name the head and the files, and go on with the next head.
- REQ-7: `moai bundle build` shall never move, reset or delete an existing branch or worktree, so a rebuild without a head is a new branch name.
- REQ-8: When `moai bundle verify` is given a bundle branch, `moai bundle verify` shall run every command of the check list on the pinned SHA under one heavy slot, going on after a failed command, and record each command, its exit status, its counts and what it did not run.
- REQ-9 (REQ-8): IF a check fails, THEN `moai bundle verify` shall name the head whose diff against the tip touches the failing file, or `unattributed`.
- REQ-10: When `moai bundle body` is given heads, a bundle branch and a run record, `moai bundle body` shall write a pull request body that passes the Self-Audit checker from the heads' facts and the run record only, with the `audit-tier:` line from `gate::tier_of(Some(0), <paths of tip...branch>)`.
- REQ-11 (REQ-10): IF a fact the body needs has no record (a head not in the branch, a head without the audit evidence of REQ-4, a check of the list that neither ran nor is named not run, a failed command, a Self-Audit line), THEN `moai bundle body` shall refuse, name every missing fact, and write no placeholder.
- REQ-12: `moai bundle` shall not push a branch, open or edit or merge a pull request, write a card note, or change a repository setting.

## Acceptance

- AC-1 (REQ-1, REQ-2): the Intent's numbers are reproducible: each is followed by the command or the pull request that gives it, and the ledger counts are those of the pinned snapshot (through note 2931: 31 marker notes, 29 readable, 28 with `spec=`, 2 unreadable, 19 cards with a marker line, 10 cards with two or more readable lines; 81 `AUDIT … head` notes, 63 readable, 26 with `AUDIT #`; 18 of 29 READY lines with a `PASS` note naming the head).
  manual: re-run the ledger count described in the Intent with the note ceiling 2931, `gh pr view 3225 --json body` and `gh pr view 3227 --json body`; the steps and the body parts of Intent point 1 to 3 are the ones those bodies show.
- AC-2 (REQ-12): the Plan's boundary table puts every step of the hand procedure on the CLI side or on the leader's side, and every leader step says why it is not the CLI's.
  manual: read the table; every step of Intent points 1 to 3 appears in it once.
- AC-3 (REQ-1 … REQ-11): the Plan's slice table gives, for each of slices 2 to 4, its actions, the REQ numbers it implements, what it reads and writes, and the names of its first tests; each later slice is its own SPEC with `parent: SPEC-20261119-bundle-tool`.
  manual: read the slice table; every REQ from REQ-1 to REQ-11 is in exactly one slice, and REQ-12 holds in all of them. Slice 2 carries `body` (REQ-10, REQ-11) on a run record file; slice 4 writes that file.
- AC-4 (REQ-1 … REQ-12): this SPEC passes the format check with no error.
  verify: `moai-cli spec check SPEC-20261119-bundle-tool --project .`

## Scope

- write (slice 1, a plan; the SPEC of each later slice states its own scope): `.moai/specs/SPEC-20261119-bundle-tool/spec.md` only.
- read-only: `src/gate.rs`, `src/trailers.rs`, `src/spec_doc.rs`, `src/audit_gate.rs`, `src/gate_collect.rs`, `src/self_audit.rs`, `src/identity.rs`, `src/heavy_slot.rs`, `src/main.rs`, `src/ops/trailer_check.rs`, `.github/pull_request_template.md`, the real ledger (read through `moai card show`), the bodies of #3225 and #3227.
- forbidden: any change under `src/`, `tests/`, `skills/`, `docs/`, the real HOME and ledger as a write target, any push or pull request.

## Plan

**Boundary table: every step of the hand procedure, and who owns it.**

| Step | Owner | Why |
| --- | --- | --- |
| list the READY heads (newest line per card, unreadable ones named) | CLI `ready` | a read of the ledger; same answer every time |
| choose which heads go into this bundle, and in what order | leader | judgement: priority, size, risk, what is waiting on what |
| per-head checks (origin, already merged, audit evidence, trailers, spec, result coverage, conflict without the driver) | CLI `check` | each is a fact of the head; all findings at once is what bundle 5 lacked |
| accept or refuse a head's audit, decide the audit tier a head needs | leader (audit agent for the verdict) | independent audit is never delegated; the tool only reports a verdict line found or missing |
| build the integration branch (one merge per head, pinned SHA) | CLI `build` | mechanical; a new branch each time, so no reset |
| drop a failing head and rebuild | leader chooses, CLI `build` does | one decision per head; the tool names the head and the files and offers the same list minus it |
| run the check list on the pinned SHA under a heavy slot | CLI `verify` | repeatable, records the exits; one slot hold for the whole list (not one wait per command) |
| write the pull request body | CLI `body` from records | cannot say a command ran that did not |
| push the bundle branch, open the pull request, watch `ci`, merge | leader | outward-facing; the audit gate and the ruleset decide, no bypass |
| write the card notes that follow (verdict, PR link, release) | leader | the ledger is the lane's and leader's record; the tool never writes it |

**Slices.** Each is a small SPEC of its own. Slice 1 and slice 2 share the branch `spec/20261119-bundle-tool` (leader decision, card C-119); a later slice gets its own branch.

| Slice | Actions | REQ | Reads / writes | First tests (planned) |
| --- | --- | --- | --- | --- |
| 2 | `bundle ready`, `bundle check`, `bundle body` (`SPEC-20261119-bundle-ready-check`) | REQ-1 … REQ-4, REQ-10, REQ-11 | reads the ledger, the repository and the local `origin` refs (the caller fetched first, D-11) and a run record file; writes nothing, or the one new file `body --out` names | `bundle_ready_cli`, `bundle_check_cli`, `bundle_body_cli` in a temp repository with a bare `origin`, a temp ledger |
| 3 | `bundle build` | REQ-5 … REQ-7 | writes a new branch and worktree only | `bundle_build_cli`: two heads, a conflicting head, a rebuild without it, an existing branch name refused |
| 4 | `bundle verify` | REQ-8, REQ-9 | runs the list, writes the run record file that `body` reads, under the worktree's `target/` | `bundle_verify_cli` with a stub `cargo`, `python3` and `moai heavy-slot-acquire`; a `.py` test file is run by unittest, never as a cargo target |

**Order and reasons.** Slice 2 first: it is read-only, it gives the leader the one-screen view of every READY head and every failing check, which is the part that cost two hours. `body` moved into slice 2 on the leader's direction: until `verify` exists the leader writes the run record file by hand, and the body is still generated from facts. Slice 3 next, so the rebuild-without-a-head step is one command. Slice 4 needs 3 (a branch to run on) and writes the run record. REQ-12 holds in every slice; each slice's help says first whether it writes (`Reads only.` / `Writes the ledger.` convention of `card`).

**What a check run records.** Per command: the argument list, the exit status, the counts the output gives (`N passed`), the start and end clock, and the pinned SHA it ran on. The body is generated from this record and from the heads' facts; a command that was not run is `not_run` with its reason, never absent. The file is JSON (`commands`, `not_run`, `self_audit`; the exact schema is in `SPEC-20261119-bundle-ready-check`); a check of the list that neither ran nor is named in `not_run` with a reason is a missing fact for `body`, so the standing not-run items (the unfiltered whole `cargo test`, for which `ci` is cited; Linux) are named by whoever writes the record, never assumed.

**The check list (data, pinned by a test, D-4).** The union of the two merged bodies plus the lesson of bundle 3: `cargo build --locked --offline --bin moai-cli` (both bodies); the whole `--lib` (the body of #3227 only; #3225 ran a list of targets); `moai-cli spec check --all` (#3227 only); `--test home_isolation_scan` and `--test product_cli` always (both lists carry them); every `tests/<name>.rs` the heads add or change; every `--test` named in a bundled SPEC's `verify:` lines; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`; `trailer-check`; the Result-covers-AC check; `python3 -m unittest discover -s tests -p '*_test.py'` (the same discovery CI runs, so no module list to get wrong).

**Git details a slice must keep.** Run git through `identity::scrub_git_env`. The conflict check uses `git -c merge.moai-manifest.driver='git merge-file -L ours -L base -L theirs %A %O %B' merge-tree --write-tree`, reads the exit status from the process (not from a pipe), and requires one output line (a tree id) for a clean merge, because GitHub never runs the registered driver. The caller fetches the tip and the heads before any check (D-11); the command says when `origin` was last fetched and refuses a head that is not present locally.

## Risks

- A bundle tool that trusts the card notes trusts what lanes wrote: a READY line can name a head that was never audited. REQ-4 asks for the verdict line naming exactly that head; the tool cannot tell whether the verdict was written by an independent auditor, nor at which tier: a head with audit floor 2 is satisfied by any `PASS` naming it, tier-1 lane audit included. Its output says so.
- Repair heads often have no verdict naming exactly them (3 of the 7 heads in #3225 and #3227): the tool reports `audit_missing` with the nearest ancestor PASS and the leader still decides (D-10).
- `unattributed` failures (an interaction between two heads) are not solved here: the leader still decides, and the rebuild-without command makes each trial a single command.
- The check list is the current practice. If the leader or user changes it, the code and its pinning test change together; until a configuration file exists that is a build.
- This slice decides a boundary the leader owns. The SPEC stays `draft` until the leader accepts the table; nothing is built before.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | No push, no pull request, no merge, no card note from the tool | outward-facing and ledger writes are the leader's; the card says "no bypass of CI, audit or the ruleset" | a later flag, after the leader asks for it |
| D-2 | Heads and their order come from the caller; `ready` lists but never selects | which heads go together is a judgement; the card body names head selection, push and pull request as the leader's decisions | a default selection rule |
| D-3 | A rebuild is a new branch and a new worktree; no existing branch is moved | the lane rule against `reset --hard`; a pinned SHA of an earlier build stays valid | reuse a name |
| D-4 | The check list is data in the code, pinned by a test; a test file is classified by its extension (`.rs` is a cargo `--test` target, `_test.py` runs by `python3 -m unittest`, never both) | the list is the union of the two merged bodies (the whole `--lib` and `spec check --all` are in #3227 only), and a Python test was once listed as a cargo target (card C-119 body); a config file is another decision | a config file |
| D-5 | All findings at once for the per-head checks; `verify` runs the remaining checks after a failure | bundle 5: every single failure stopped the bundle for two hours | stop at the first failure (a flag) |
| D-6 | The READY grammar is the first line matching `^LANE-READY-INTEGRATION (C-[0-9]+) head=([0-9a-f]{40})(?: spec=(SPEC-[0-9]{8}-[a-z0-9-]+))?(?:$| )` with the card equal to the card the note is on; the SPEC ids come from the `Moai-Spec` trailers of the head | 29 of 31 real lines read (28 with `spec=`); the trailers are what `trailer-check` already validates | require `spec=` |
| D-7 | The action names are `bundle ready`, `check`, `build`, `verify`, `body` under one `moai bundle` | one word per action, as `card` | one command with flags |
| D-8 | A verdict on a card is a note whose first line starts with `AUDIT ` and reads under `gate::parse_verdict` unchanged (a 40-hex head, a 40-hex base, a verdict word without a space); a note that starts with `AUDIT ` and does not read is listed by `check` as `unreadable_verdict` (informational, on the head's own card; it does not by itself fail the check), never guessed | 63 of 81 card verdict notes read, and `gate::verdict_line` would accept none of those that name a READY head | accept short bases (`parse_verdict_short_base`) or a card-specific grammar |
| D-9 | The audit evidence a head needs is set by its changed paths: the floor is `gate::tier_of(Some(0), <paths of tip...head>)` (three dots: what the head adds since its merge base with the tip; two dots overcount for a head behind the tip); 0 needs no verdict, 1 or more needs a `PASS` naming exactly the head. A higher tier the lane asked for is a note the tool does not read | `tier_of` takes a declared tier and paths and a bare head has neither; a declared 0 makes the paths the only input, and `None` would make every head tier 2 | read `감사 요청 tier N` notes |
| D-10 | A head with no verdict naming it exactly is reported as `audit_missing`, with the nearest PASS on an ancestor of the head when there is one (`ancestor_pass`); the leader decides whether a repair delta is accepted | 11 of 29 READY heads have no exact verdict, mostly repair heads; the gate carries a PASS over a mechanical delta only, which the tool does not judge | treat an ancestor PASS as enough |
| D-11 | The commands never fetch; the caller fetches first. They read local refs, print the tip's SHA and commit time (git keeps no fetch time that every worktree shares: `FETCH_HEAD` is per worktree), and report a head that is not present locally as `head_missing`, "fetch first" | REQ-1 and REQ-12 say the tool writes no ref; a check run on a stale tip is the quiet failure | the tool fetches the named branches |
| D-12 | The `Moai-Spec` trailers and `trailer-check` are read over `tip..head` (merges skipped, as `trailer-check` does) | the commits a head adds to the tip are the ones a bundle brings | the whole branch history |
| D-13 | The `build` worktree is a new folder under `~/.moai/worktrees/`, named `<project>-bundle-<n>`, and the branch `m-flow/bundle-<yyyymmdd>-<n>`, where `<n>` is the lowest number that no local or `origin` branch of that date uses and no worktree folder uses; an existing name is refused | `git worktree list` shows this project's worktrees in that folder; bundles 2 to 5 used `m-flow/bundle-20261119-<n>` | another folder |

## Result

- AC-1: met — the Intent's numbers, recounted after the first audit at the pinned snapshot (note ceiling 2931): `moai card show C-n` over C-1 to C-121 gives 31 notes with the marker, 29 under the grammar (28 with `spec=`), 2 unreadable, 19 cards with a marker line, 10 cards with two or more readable lines; 81 `AUDIT … head` notes, 63 read by `parse_verdict`, 26 start with `AUDIT #`; 18 of 29 READY lines have a PASS naming the head, 0 of those 18 start with `AUDIT #`; `gh pr view 3225` and `gh pr view 3227` give the steps and body parts; `git grep -l -E 'spec_result_check|wait_pass_then_merge|merge_when_green' origin/v2026.11.0` gives 0 files (tip d25ecb1ba1); `gh run list --branch m-flow/bundle-20261119-2` shows `ci failure` on `eba078d26b`.
- AC-2: met — the boundary table lists the ten steps; the Intent's points 1 to 3 and the card body's failure classes map to them (rebuild: `build`; per-head: `check`; run: `verify`; body: `body`; push, PR, CI, merge, notes: leader).
- AC-3: met — the slice table covers REQ-1 to REQ-11 once each (2: REQ-1…4 and REQ-10, 11 with `body` on a run record file; 3: REQ-5…7; 4: REQ-8 and 9), and REQ-12 holds in every slice; revision 2 moved `body` from slice 5 to slice 2 on the leader's direction and dropped slice 5.
- AC-4: met — `moai-cli spec check SPEC-20261119-bundle-tool --project .` (a binary built from tip d25ecb1ba1 plus the C-94 and C-108 branches, not rebuilt from this exact head; the SPEC check does not depend on those branches) → `errors 0`, `warnings 0` (the SPEC has 12 requirements, the count at which `size_requirements` stops warning).
