---
id: SPEC-20261119-bundle-tool
title: "moai bundle: assemble, verify and describe an M-Flow bundle from READY heads (plan)"
status: draft
revision: 1
created: 2026-11-19
---

# moai bundle: assemble, verify and describe an M-Flow bundle from READY heads (plan)

## Intent

Card C-119, slice 1: the plan and this SPEC, no code. Under M-Flow (user decision 2026-11-19) lanes push `spec/<date>-<slug>` branches, a head is audited, the lane writes one `LANE-READY-INTEGRATION` card note, and the leader integrates several READY heads into one pull request. That integration is done by hand today. This SPEC fixes what a `moai bundle` command takes over and what stays the leader's decision, so each later slice is a small SPEC of its own.

**The hand procedure, measured.** Read from the bodies of the merged bundle pull requests #3225 (bundle 2, merge `bdb4d3db8c`, 3 heads, 17 commits) and #3227 (bundle 4, merge `d25ecb1ba1`, 4 heads, 13 commits), `gh pr view <n> --json body`:

1. An integration branch is rebuilt from the tip with one merge commit per audited head; its SHA is pinned.
2. The pinned SHA is verified locally: `cargo build --locked --offline`, the whole `cargo test --locked --offline --lib`, a list of `--test` targets, `cargo fmt --check`, `cargo clippy --locked --offline --all-targets -- -D warnings`, `moai-cli trailer-check origin/v2026.11.0..HEAD`, `moai-cli spec check --all`, `python3 -m unittest` over a list of modules, and a check that every AC of each bundled SPEC has a line in its `## Result`.
3. One pull request is opened. Its body is: the line `audit-tier: N <paths>`, one `Moai-Spec:` line per SPEC, a table of heads (SPEC, card, head, audit evidence), an `Integration:` line (tip, pinned SHA, commit and file counts), `## Self-Audit` (plan, red, mutation, host, probes, not done locally) and `## Result` (each command with its exit and count, then one line per SPEC).

**What goes wrong, measured.** None of the leader's scripts the card names (`spec_result_check.py`, `wait_pass_then_merge.sh`, `merge_when_green.sh`) is in the repository: `git grep -l -E 'spec_result_check|wait_pass_then_merge|merge_when_green' origin/v2026.11.0` finds 0 files, so every bundle is assembled again. The card C-119 body records bundle 5: 19 heads held for about two hours because six heads each failed one check and every failure stopped the whole bundle; rebuilding without the failing heads took minutes once done. The same body records two tool faults: a Python test listed as a cargo target, and the hand SPEC Result checker's two false positives that the product `spec check` does not give. One more fault was re-measured: clippy was not run before the bundle 2 pull request, and the `ci` run on its first head `eba078d26b` failed (`gh run list --branch m-flow/bundle-20261119-2` shows `ci failure`; the body of #3225 names `tests/workflow_size_budget.rs:104`). Three more are from the leader's own records and were not re-measured for this SPEC: a merge driver registered in the clone made a local `git merge-tree` clean where GitHub conflicted (#3162, 2026-11-17); `cargo test --lib <filter>` missed a repository-wide scan that failed on Linux (#3226, 2026-11-19); a head was merged on an audit note although its tip-built `trailer-check` had not passed (card C-60).

**The READY line, measured.** In the real ledger (`moai card show C-1 … C-134`, read only) 31 card notes begin with `LANE-READY-INTEGRATION`. 29 read under the grammar of D-6 (`head=` required, `spec=` optional, text after a space allowed; 28 carry `spec=`), 2 do not (`… agent-1 C-33 S2a — …`, `… agent-1 C-60 branch … head <sha> …`). 6 of the 15 cards that have a line have two or three, with different heads: a repair moves the head, so the newest line must win.

**Parts that exist and are reused, not written again.** `gate::tier_of` and `gate::parse_body_tier` (`src/gate.rs`: the effective audit tier from the changed paths, the body line rule); `gate::verdict_line` and `gate::parse_verdict` (`src/gate.rs`: the verdict first line); `trailers::check` (`src/trailers.rs:361`, behind `trailer-check`); `spec_doc` (`SpecDoc.result`, `Outcome`, the `closure` rule in `src/spec_doc.rs`: the status rule is part of `closure`, so a Result-covers-AC check without it is a small new function over the same parse); the Self-Audit checker (`pr-self-audit-check`, `src/gate_collect.rs:681`); `moai heavy-slot-acquire`; `identity::scrub_git_env`.

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
- REQ-4 (REQ-3): IF a head fails a per-head check (not reachable from `origin`, already in the tip, no `PASS` verdict naming exactly that head at its effective audit tier, trailers, a SPEC error, an AC with no outcome line in `## Result`, or a conflict with no merge driver running), THEN `moai bundle check` shall report the head with a stable code and, for a conflict, the files.
- REQ-5: When `moai bundle build` is given heads in an order, `moai bundle build` shall create a new branch and a new worktree from the tip, merge each head into it with one merge commit, and print the pinned SHA.
- REQ-6 (REQ-5): IF a head's merge conflicts, THEN `moai bundle build` shall leave the branch without that head, name the head and the files, and go on with the next head.
- REQ-7: `moai bundle build` shall never move, reset or delete an existing branch or worktree, so a rebuild without a head is a new branch name.
- REQ-8: When `moai bundle verify` is given a bundle branch, `moai bundle verify` shall run every command of the check list on the pinned SHA under one heavy slot, going on after a failed command, and record each command, its exit status, its counts and what it did not run.
- REQ-9 (REQ-8): IF a check fails, THEN `moai bundle verify` shall name the head whose diff against the tip touches the failing file, or `unattributed`.
- REQ-10: When `moai bundle body` is given a verified bundle, `moai bundle body` shall write a pull request body that passes the Self-Audit checker from the heads' facts and the run records only, with the `audit-tier:` line from `gate::tier_of` over the union of the changed paths.
- REQ-11 (REQ-10): IF a fact the body needs has no record (a verdict line, a plan note, a command that was not run and is not on the list of standing not-run items), THEN `moai bundle body` shall refuse, name the fact, and write no placeholder.
- REQ-12: `moai bundle` shall not push a branch, open or edit or merge a pull request, write a card note, or change a repository setting.

## Acceptance

- AC-1 (REQ-1, REQ-2): the Intent's numbers are reproducible: each is followed by the command or the pull request that gives it, and the READY-line counts (31, 29, 28, 2, 6 of 15) are those of the grammar in D-6.
  manual: re-run the ledger count of the Result section and `gh pr view 3225 --json body` and `gh pr view 3227 --json body`; the steps and the body parts of Intent point 1 to 3 are the ones those bodies show.
- AC-2 (REQ-12): the Plan's boundary table puts every step of the hand procedure on the CLI side or on the leader's side, and every leader step says why it is not the CLI's.
  manual: read the table; every step of Intent points 1 to 3 appears in it once.
- AC-3 (REQ-1 … REQ-11): the Plan's slice table gives, for each of slices 2 to 5, its actions, the REQ numbers it implements, what it reads and writes, and the names of its first tests; each later slice is its own SPEC with `parent: SPEC-20261119-bundle-tool`.
  manual: read the slice table; every REQ from REQ-1 to REQ-11 is in exactly one slice, and REQ-12 holds in all of them.
- AC-4 (REQ-1 … REQ-12): this SPEC passes the format check with no error.
  verify: `moai-cli spec check SPEC-20261119-bundle-tool --project .`

## Scope

- write: `.moai/specs/SPEC-20261119-bundle-tool/spec.md` only (slice 1 is a plan).
- read-only: `src/gate.rs`, `src/trailers.rs`, `src/spec_doc.rs`, `src/audit_gate.rs`, `src/gate_collect.rs`, `src/ops/trailer_check.rs`, `.github/pull_request_template.md`, the real ledger (read through `moai card show`), the bodies of #3225 and #3227.
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

**Slices.** Each is a small SPEC of its own, pushed on its own branch.

| Slice | Actions | REQ | Reads / writes | First tests (planned) |
| --- | --- | --- | --- | --- |
| 2 | `bundle ready`, `bundle check` | REQ-1 … REQ-4 | reads the ledger, the repository (after a fetch the caller made), `origin` refs; writes nothing | `bundle_ready_cli`, `bundle_check_cli` in a temp repository with a bare `origin`, a temp ledger |
| 3 | `bundle build` | REQ-5 … REQ-7 | writes a new branch and worktree only | `bundle_build_cli`: two heads, a conflicting head, a rebuild without it, an existing branch name refused |
| 4 | `bundle verify` | REQ-8, REQ-9 | runs the list, writes a run record under the worktree's `target/` | `bundle_verify_cli` with a stub `cargo`, `python3` and `moai heavy-slot-acquire`; a `.py` test file is run by unittest, never as a cargo target |
| 5 | `bundle body` | REQ-10, REQ-11 | reads the run records, writes the body file | `bundle_body_cli`: a body that passes `pr-self-audit-check`, a missing fact refused |

**Order and reasons.** Slice 2 first: it is read-only, it gives the leader the one-screen view of every READY head and every failing check, which is the part that cost two hours. Slice 3 next, so the rebuild-without-a-head step is one command. Slice 4 needs 3 (a branch to run on). Slice 5 needs 4 (records). REQ-12 holds in every slice; each slice's help says first whether it writes (`Reads only.` / `Writes the ledger.` convention of `card`).

**What a check run records.** Per command: the argument list, the exit status, the counts the output gives (`N passed`), the start and end clock, and the pinned SHA it ran on. The body is generated from this record and from the heads' facts; a command that was not run is `not_run` with its reason, never absent. Standing not-run items (the unfiltered whole `cargo test`: `ci` is cited instead; Linux) are on a list the SPEC of slice 4 fixes.

**The check list (data, pinned by a test, D-4).** From the two merged bodies: build; `--lib` whole; `--test home_isolation_scan` and `--test product_cli` always (both lists carry them); every `tests/<name>.rs` the heads add or change; every `--test` named in a bundled SPEC's `verify:` lines; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`; `trailer-check`; `spec check --all`; the Result-covers-AC check; `python3 -m unittest discover -s tests -p '*_test.py'` (the same discovery CI runs, so no module list to get wrong).

**Git details a slice must keep.** Run git through `identity::scrub_git_env`. The conflict check uses `git -c merge.moai-manifest.driver='git merge-file -L ours -L base -L theirs %A %O %B' merge-tree --write-tree`, reads the exit status from the process (not from a pipe), and requires one output line (a tree id) for a clean merge, because GitHub never runs the registered driver. The tip is fetched before any check, and the command says when it was last fetched.

## Risks

- A bundle tool that trusts the card notes trusts what lanes wrote: a READY line can name a head that was never audited. REQ-5 asks for the verdict line naming exactly that head; the tool cannot tell whether the verdict was written by an independent auditor, and says so in its output.
- `unattributed` failures (an interaction between two heads) are not solved here: the leader still decides, and the rebuild-without command makes each trial a single command.
- The check list is the current practice. If the leader or user changes it, the code and its pinning test change together; until a configuration file exists that is a build.
- This slice decides a boundary the leader owns. The SPEC stays `draft` until the leader accepts the table; nothing is built before.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | No push, no pull request, no merge, no card note from the tool | outward-facing and ledger writes are the leader's; the card says "no bypass of CI, audit or the ruleset" | a later flag, after the leader asks for it |
| D-2 | Heads and their order come from the caller; `ready` lists but never selects | which heads go together is a judgement; the card says "head selection stays the leader's" | a default selection rule |
| D-3 | A rebuild is a new branch and a new worktree; no existing branch is moved | the lane rule against `reset --hard`; a pinned SHA of an earlier build stays valid | reuse a name |
| D-4 | The check list is data in the code, pinned by a test; a test file is classified by its extension (`.rs` is a cargo `--test` target, `_test.py` runs by `python3 -m unittest`, never both) | the list is the two merged bodies' common part, and a Python test was once listed as a cargo target (card C-119 body); a config file is another decision | a config file |
| D-5 | All findings at once for the per-head checks; `verify` runs the remaining checks after a failure | bundle 5: every single failure stopped the bundle for two hours | stop at the first failure (a flag) |
| D-6 | The READY grammar is the first line matching `^LANE-READY-INTEGRATION (C-[0-9]+) head=([0-9a-f]{40})(?: spec=(SPEC-[0-9]{8}-[a-z0-9-]+))?(?:$| )` with the card equal to the card the note is on; the SPEC ids come from the `Moai-Spec` trailers of the head | 29 of 31 real lines read (28 with `spec=`); the trailers are what `trailer-check` already validates | require `spec=` |
| D-7 | The action names are `bundle ready`, `check`, `build`, `verify`, `body` under one `moai bundle` | one word per action, as `card` | one command with flags |

## Result

- AC-1: met — the Intent's numbers: `moai card show C-n` over C-1 to C-134 gives 31 notes with the marker, 29 under the grammar (28 with `spec=`), 2 unreadable, 6 of 15 cards with more than one line; `gh pr view 3225` and `gh pr view 3227` give the steps and body parts; `git grep -l -E 'spec_result_check|wait_pass_then_merge|merge_when_green' origin/v2026.11.0` gives 0 files (tip d25ecb1ba1); `gh run list --branch m-flow/bundle-20261119-2` shows `ci failure` on `eba078d26b`.
- AC-2: met — the boundary table lists the ten steps; the Intent's points 1 to 3 and the card body's failure classes map to them (rebuild: `build`; per-head: `check`; run: `verify`; body: `body`; push, PR, CI, merge, notes: leader).
- AC-3: met — the slice table covers REQ-1 to REQ-11 once each (2: REQ-1…4, 3: REQ-5…7, 4: REQ-8 and 9, 5: REQ-10 and 11), and REQ-12 holds in every slice.
- AC-4: met — `moai-cli spec check SPEC-20261119-bundle-tool --project .` (binary built from tip d25ecb1ba1 plus C-108's branch) → `errors 0`, `warnings 0` (the first draft had 17 requirements and drew `size_requirements`; it was cut to 12 and the detail moved to the slice SPECs).
