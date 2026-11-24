# Leader rules

The detail behind the "Leader rules" section of `CLAUDE.md`. The leader is the session the user talks to: it supervises the lanes, relays what a lane cannot post, lands audited heads and keeps the decisions log. The division of work, the autonomy levels `full` and `semi` with the keep-set and the paste-ready `/goal` line are defined once in `docs/design-review/autonomy-modes-spec.md` §2, §5, §6 and §7 (written for the Issue ledger; where it says claim comment or Issue, read the card reservation and the card). Where a rule moved here from the earlier `CLAUDE.md`, `docs/claude-md-rule-provenance.md` names the line it came from.

## What a round reads

Evidence, not status labels:

- The cards: `moai card list`, `moai card next`, `moai card inbox` (the lane notes that wait for the leader) and `moai card show` for the reservation, notes and criteria.
- `ListAgents`: who is live, busy or idle, and duplicate names.
- Each lane worktree's `git status --short`, and the lane's last reported command and exit code.
- Stage changes come from the ledger and git (reservation, branch, landed head), never from a lane message or an idle notice alone.

Ownership is the card reservation, never a local file or a message. The earlier `assignments.json` cache (`docs/design-review/assignments-schema.md`) is not written by any code today, so a round reads the ledger.

## When a round runs

On the interval (`leader.pr_check_interval_secs` of the profile, default 300, a user-set value and not a measured one), and also after every landing, every lane report or hand-off, every restart, whenever the leader would otherwise wait, and when a lane reports idle with an empty list. The user's decision of 2026-11-08 is that the leader always keeps checking lane progress. A later reading that the fixed polling loops are off in favour of waking on ledger events (`moai card wait`, `moai card inbox`) is not recorded in the decisions log; its basis is not confirmed (card note 7366, finding N7), so the interval above stands until the leader records it.

## What a round does

1. Reservation mismatch: a lane that writes without a live reservation of its card (`moai card check` does not answer `allowed: true`) is told to stop, and what differed is recorded as a note on that card.
2. Repair: write it as a card note with `assignee-hint: agent-N` addressed to a free lane that is not the author. The hint grants nothing: ownership comes from that lane's own reservation, and one writer per branch holds.
3. Stalled lane: a lane idle without a reason, or busy past its reported step with no report, is asked for its actual last command and exit code; a block is recorded and raised to the user.
4. Refill: keep a next card ready for every lane, in this order: (1) a ready card whose blockers are all closed; (2) a card whose only open blocker is a head already in the leader's landing queue; (3) a recorded follow-up turned into work: an audit's non-blocking finding is added as a criterion of the SPEC's open follow-up card with `moai card criterion-add <card> --ac <text>`, and a card is opened with `moai card add --spec <SPEC-ID> --follow-up` only when the SPEC has none open (a SPEC keeps one open follow-up card); defect candidates with the RED test first and closed as not-a-defect when it does not reproduce and the ledger's follow-up candidates are turned into work the same way; (4) when none exists, tell the waiting lane why (which landing or HOLD it waits on) and which card is reserved for it.
5. Each round and its outcome go into the dispatch journal.

## What the leader never does

- Invents work outside the ledger and its recorded backlog, splits one card between two writers, creates a new or duplicate writer, or lifts a HOLD to make work. A card that needs a HOLD answered waits for the user while the lane takes another.
- Uses a duplicate session found after a restart: it freezes until the operator assigns it.
- Publishes new cards from a report before the user confirms the report; each such card's body is mirrored with a "사용자 결정" section, and a card that edits a design page takes "sha256 of the corrected copy plus the leader's read-back" as its acceptance. Lanes get only `.md` paths in tracked locations, so a report moves to a tracked path such as `docs/design-review/` before its cards are issued.

## Naming a lane branch

The leader makes the lane worktree and its branch before the lane has picked its SPEC name, and the SPEC-first hook reads the branch name as the SPEC id (`spec/<slug>-<yyyymmdd>` for `SPEC-<yyyymmdd>-<slug>`). So the leader names the branch `spec/<slug>-<yyyymmdd>` after the card's own words, with the date of the day it is made, and tells the lane that name in the dispatch. A lane never renames the branch. When the SPEC the lane writes has another slug or date, the lane adds one line `branch: spec/<slug>-<yyyymmdd>` (the branch it is on, written exactly) to the SPEC's frontmatter, best while the SPEC is still a draft so the approval covers it; the hook reads that SPEC for the branch while its status is `draft` or `approved` (the folder name equal to the `id:`, the line inside the frontmatter's closing `---`, one level under `.moai/specs/`, no link followed; two SPECs naming one branch are refused). It looks for the line only when no reading of the branch name has a SPEC file of its own. A branch with no SPEC and no such line stays refused: code waits for a SPEC (`SPEC-20261124-lane-branch-after-spec`).

## Bundles, clearing and the whole suite

- Bundles stay small, 5 to 7 heads; a head that fails verification is dropped and the rest ships. (Bundle 5 shipped 13 heads and left out three after local verification failed on them; its audit took 41 minutes.)
- Clear after every card: a READY, REPAIRED or MERGE-READY note and a `LANE-DONE` line hand a card back and are never done. The leader reads the ledger and the lane screen and sends `/clear` through `orca terminal send` to the exact existing window only when the ledger shows the lane's card done or released and nothing is open (no card in review, no live reservation, no run of the lane's own, no unsaved change, no prompt on the screen). After `/clear` it confirms the new session is ready, sends the next line once and reads the new reservation back (the ACK); with no ACK it inspects and does not send again. Stages: `skills/moai/references/flows/next-flow.md`.
- Every five landings the leader runs every Rust test on the integration tip once; a red test stops landings until it is fixed.
- Lease: a card reservation lasts four hours unless `--lease-secs` says otherwise (14400 seconds by default); expiry is displayed only, and `moai card reclaim` takes over a card whose reservation has ended.

## Landing and the release pull request

- Landing a card is `moai card land <card> --head <40hex> --by leader`, after the leader has read the verdict note on the card. The verdict is written by a session that wrote none of the change; a verdict by the author does not count, and a posted FAIL stops the landing at every tier.
- A bundle verdict is written by a lane that wrote none of the bundle's commits, and its note names that lane (`agent-N`) and its `Moai-Session` value (`moai trailers --spec <SPEC-id>`, run in the auditor's own worktree). `audit-gate` counts a PASS only when it names an auditor lane or `Moai-Session` that is not an author of the commits (rule `non-author-pass`, `src/audit_gate.rs`), so a verdict by the leader's own subagent, which names no auditor lane, does not count.
- The release pull request is the only pull request. Procedure: read the verdict's first line; verify the head equals the audited SHA and the merge-base equals the base tip (else `gh pr update-branch` and audit again); after the verdict re-run the `audit-gate` run (`gh run rerun <run id>`); post `moai/audit=success` with the verdict comment as `target_url` only when that job succeeded on the same head, never on the strength of a green `ci` and a PASS comment alone; then `gh pr merge --auto --merge`.
- After every `update-branch`, compare the changed lines with the audited delta before asking for a delta audit: when both sides changed the same line the same way (for example a fixed array length), git merges silently and the build can break; registries such as `OPERATIONS` are length-free slices for that reason.
- An audit dispatch names its caps: no whole `cargo test`, no `python3 -m unittest discover` (about 38 minutes), named test modules only.
- Both autonomy levels keep the ruleset, the independent audit, no secrets in logs, no hook or permission bypass and the preservation boundaries of `CLAUDE.md` unchanged.
- Known gap: all sessions share one GitHub account, so the status source cannot be pinned (an audit-only GitHub App is a later HOLD).
