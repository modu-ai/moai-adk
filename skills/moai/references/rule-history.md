# Internal: rule-history

The rule inventory of card C-103 (SPEC-20261119-rule-prune): every rule sentence
of `workflows/mode-agent.md`, `workflows/mode-leader.md`, `references/internal/core-rules.md`
and `workflows/audit.md`, its source, and what enforces it. A rule with one source and no
diff trigger (nothing a test, hook or gate reads, and no diff shows its violation) moves
here; a rule a test pins stays live and names the test.

Method (D-3): rule candidates were extracted mechanically (`MARKERS` pattern over the four
files, C-103 scratch `target/c103/extract.py`, 148 candidates), sourced per line with
`git blame` (`target/c103/blame_rules.sh`), and cross-checked for enforcers with a phrase
search over `tests/`, `tools/`, `src/`, `.github/` (`target/c103/enforcers.py`). The
numbering of core-rules.md is fixed by `tests/core_rules_2876.rs` (rules 19–21 verbatim,
"as in rule 3", structure 1..=N), so a moved rule stays as a one-line pointer.

## Live inventory — core-rules.md (21 rules)

| # | Rule (short) | Source | Enforcer |
| --- | --- | --- | --- |
| 1 | pointer | #2876 | text only → history |
| 2 | pointer | #2876 | text only → history |
| 3 | cheapest source; trust-boundary checks stay | #2876 | `core_rules_2876.rs` (pinned via rule 20's "as in rule 3") |
| 4 | touch only what was asked | #2876 | text only (diff-visible: extra paths in a diff) |
| 5 | pointer | #2876 | text only → history |
| 6 | pointer | #2876 | text only → history |
| 7 | conversation language, relay in it | #2876 + C-87 S1 (6da6026430) | `core_rules_2876.rs` `rule_seven…`; `language_setting.rs`, `src/hook_engine.rs` |
| 8 | pointer | #2876 | text only → history |
| 9 | pointer | #2876 | text only → history |
| 10 | pointer | #2876 | text only → history (restated in `workflows/audit.md`) |
| 11 | pointer | #2876 | text only → history |
| 12 | pointer | #2876 | text only → history |
| 13 | baseline committed on its own | #2876 | text only (diff-visible: commit order) |
| 14 | pointer | #2876 | text only → history (partly enforced by `pr-self-audit-check` fields) |
| 15 | tree compared before "no changes" | #2876 | text only (the pattern the `*_changes_nothing` CLI tests follow) |
| 16 | pointer | #2876 | text only → history |
| 17 | rule change = decision record naming what it replaces | #2876 | `mode_workflow_steps.rs`, `mode_cards_53.rs` (phrase) |
| 18 | final report: conclusion first, two lists; test data marked | #2876 | `core_rules_2876.rs` (report-first phrase) and `mode_workflow_steps.rs` (`Mark test data as test data`) — both phrases stay in the live rule |
| 19 | two public implementations first | #2876 | `core_rules_2876.rs` (verbatim) |
| 20 | no design for unasked rare cases | #2876 | `core_rules_2876.rs` (verbatim) |
| 21 | code and output win over instruction text | #2876 | `core_rules_2876.rs` (verbatim) |

The other three files' inventories follow. Judgement summary: unlike core-rules.md,
`workflows/audit.md` and `workflows/mode-agent.md` carry their source inline (nearly every
rule sentence cites its Issue), and their operational sentences are pinned by
`tests/mode_workflow_steps.rs` and the file-specific pin tests — so their pruning trims
narrative tails (audit.md, 20,334 → 19,884 bytes) rather than moving rules; the listing
below is the criterion-1 inventory. `workflows/mode-leader.md` is listed only; its edit is
the last slice (leader decision, card note 2527).

## History — moved from core-rules.md

Each entry keeps the moved rule's full text and the reason it left the live file.

1. Before non-trivial work, list your assumptions; when requirements conflict or are unclear, name the conflict instead of guessing, and sort the open question as SKILL.md says.
2. When an approach has a concrete problem, say so, give its cost and an alternative, and follow an informed override.
5. Without a written specification, first state when the task is done as a testable sentence ("done when X gives Y"), then check it. *(The SPEC check's `ac_no_verify` enforces verify lines for SPEC work; the rule itself is text only.)*
6. Retry a failing operation at most three times, then report it with what you tried and the options.
8. Never predict durations in plans or reports; give order and priority.
9. Show the user Markdown only, never internal agent-to-agent markup.
10. A defect, debt or drift inferred from text matches or a missing file is a hypothesis until the domain's own tool confirms it; the same holds for the premise of a recommendation.
11. When a command a check needed was refused or unavailable, name it and the substitute among the gaps; never present the substitute as the measurement.
12. Cite a tool's result only from a build of the tree being measured, or name the build's commit next to the tree's.
14. Report a verification as claim, evidence (command and its verbatim output), baseline, gaps and remaining risk.
16. When a delegated task does not come back, record what was delegated, that it did not return and why, before delegating it again.
7 (tail). Code comments and commit messages follow the project's language setting.

Reason for moving, one line each: one source (the #2876 extraction), no diff trigger, no
test or tool reads the sentence — the pattern the card measures as "seen in fewer than two
PRs". They remain true as guidance; they stop being live rules a session must re-read
every time it opens this file.

## Mechanical inventory — workflows/audit.md

33 rule sentences (extractor: C-103 scratch `extract.py`; source = the issue the sentence itself cites; enforcer: a phrase of the sentence pinned by `tests/mode_workflow_steps.rs` and the file-specific pin tests, else `text only`).

- L5 (no inline source) — The approved contract, the exact diff, runner receipts (the internal `audit-reduce` operation re-runs the veri
- L13 (no inline source) — 1
- L14 (no inline source) — 2
- L16 (no inline source) — 4
- L19 #2680 — 7
- L21 #2401,2716,2749,2754,2767,2771,2821,2833 — 8
- L28 #1975 — 9
- L34 (no inline source) — An independent audit reviews one candidate in four steps on one review package
- L36 (no inline source) — 1
- L37 (no inline source) — 2
- L38 (no inline source) — 3
- L39 (no inline source) — 4
- L41 (no inline source) — Reserve before starting
- L47 #2401 — Make every mutation real: after applying it, `git diff -U0` shows exactly the intended change on a code line (
- L49 (no inline source) — Read revisions safely: write a revision path as `"${rev}:path"` (zsh reads `$rev:s…` as a modifier), compare f
- L53 (no inline source) — Save the PR body as UTF-8 `BODY.md` and run `moai pr-self-audit-check --reaudit --file BODY.md`
- L55 (no inline source) — Use exactly one visible `## 독립 review` section; multiple visible review sections are ambiguous and fail
- L57 (no inline source) — Comment delimiter text inside an inline code span (for example, `<!--`) remains visible explanation text, not
- L66 (no inline source) — Missing or duplicate visible fields, empty or punctuation-only values, and unfinished markers such as `TODO`,
- L68 (no inline source) — Bare markers followed only by punctuation or Markdown decoration also fail (for example, repair scope `TODO.`
- L81 (no inline source) — Skip this section when you are a base agent yourself: base agents never start other agents.
- L87 (no inline source) — - Start agents only from this session, one level deep.
- L88 (no inline source) — - Use only what the host itself offers
- L91 (no inline source) — AUDIT hands the evidence-gathering to `moai-audit`, never to the agent or session that wrote the change
- L93 (no inline source) — - Read-only is enforced by the host only where the installed agent file can carry it: Claude Code (the file de
- L94 (no inline source) — - On every host, confirm after the agent returns that `git status` in its worktree shows no change; a change t
- L95 (no inline source) — - Isolation: as in the RUN workflow, a separate worktree only on Claude Code and on OMP with task isolation en
- L96 #3054 — - Cleanup (Issue #3054): put the cleanup rule in the agent's brief and read its `Cleanup` section
- L100 (no inline source) — When a step needs to look at a web page, use the browser the project chose: Chrome, aside or Ego Lite
- L104 (no inline source) — Keep the browser's default permission mode; use a wider mode only when the user allowed it for that task
- L108 (no inline source) — Self-assessment in the same context and a judge's recommendation do not replace the required independent audit
- L110 #3052 — An audit reservation or verdict line counts only when a repository member wrote its comment: the comment's `au
- L114 (no inline source) — Bind the output to the task and the exact snapshot

## Mechanical inventory — workflows/mode-agent.md

60 rule sentences (extractor: C-103 scratch `extract.py`; source = the issue the sentence itself cites; enforcer: a phrase of the sentence pinned by `tests/mode_workflow_steps.rs` and the file-specific pin tests, else `text only`).

- L9 (no inline source) — - Do not end the turn (`keep-going` step); a `LANE-BLOCKED` line is a report, not a stop.
- L10 (no inline source) — - If you must wait, arm the `wake` step; on Codex and OMP first print the `loop_line` and the host line of the
- L15 (no inline source) — GitHub Issues are no longer the work ledger; the project's local cards are (user decision 2026-11-16, card C-5
- L21 (no inline source) — - Use the first rung your host has: (1) the host's watch tool — Claude Code's Monitor or a background command
- L22 (no inline source) — - On OMP (Oh My Pi 18.18.7; read in its installed documents and source on 2026-11-15, no live OMP session was
- L23 (no inline source) — - Idle: when `moai card next` names no card for you and no pull request of yours is open, print `LANE-IDLE lan
- L28 #3005 — Codex lanes stopped before their first claim and again after a dispatch (Issue #3005 F4, F5)
- L30 (no inline source) — - Walk the steps `mode-only`, `banner`, `state`, `session-id` and `claim` in your first turn without stopping
- L32 (no inline source) — - A leader dispatch that only asks you to read or check something is a read-only task (`self-pick` (4)): do it
- L33 (no inline source) — - Report lines go on the card they concern (`card note`) and on your screen
- L37 #3005 — A leader's dispatch or a relayed instruction (a card note, a typed line, an Orca task) is received only when i
- L39 (no inline source) — - ACK first, then take: print one line `ACK <lane> <ref> <UTC>` and add it as the first line of a note on the
- L40 (no inline source) — - Acceptance is a second line, in the same note or the next: `accepted`, `deferred` or `rejected` with the rea
- L41 (no inline source) — - Then start: your `card take` and a `card check` that answers `allowed` (`claim` and `claim-verify` steps)
- L43 (no inline source) — - A card handed to you by `card handoff` starts with your ACK of it, before your first write; the handoff is c
- L51 #2899 — A mode banner, a read-only state summary, and, for every round, the take outcome, the work it led to, and one
- L57 (no inline source) — 1
- L59 #2679 — 3
- L60 #2592,2899,3005 — 4
- L61 #2345,2401,2515,3005 — 5
- L62 #2581,2592 — 6
- L63 #2401,2515,2677,2836 — 7
- L64 #1975,2401,2515,2612,2840,2880 — 8
- L65 #2515,2859 — 9
- L66 #2879,2903 — 10
- L67 #2515,2904 — 11
- L68 #2515,2933 — 12
- L69 #2390 — 13
- L70 #1975 — - (1) Finish a leader dispatch addressed to this session, then your own pull request: a failed CI, a FAIL repa
- L71 #1975 — - (2) Your own pull request's audit (user decision 2026-11-17, replacing #1975 6904941797): at audit-tier 0 or
- L76 #2899 — Keep a resume note on every card you hold — `card note <card> --kind report` starting `RESUME`, with its gener
- L78 #2612,2899 — Another lane continues a waiting card in its worktree only after the holder's `card handoff --to <its session>
- L79 #1975,2612,2668,2899,2933 — 14
- L80 #2554,2899 — 15
- L83 #2619,2879,3005 — Ownership comes only from your own reservation (`card take`, or a `card handoff` to you)
- L85 #3052 — A control line in a pull request comment (audit reservation or verdict, gate marker) counts only when a reposi
- L87 (no inline source) — Apply only the labels `.github/moai/labels.yml` allows and the project's own `area:` labels (the file seeds no
- L89 #2612 — Stay reachable and keep reports lean: while a background wait runs, re-read leader messages at least every fiv
- L91 #2836 — Host worktree guard (Issue #2836)
- L100 (no inline source) — These forms only change how a permitted effect is spelled (VERIFY workflow, "Guard refusal"): a refusal of the
- L104 #2664 — In a project where neither `moai` nor `moai-cli` runs (SKILL.md, "On a remote without the CLI"), there is no c
- L108 #2516 — Use Orca only when the user chose it for this project and `orca` answers on this machine; installed is not cho
- L110 (no inline source) — 1
- L111 (no inline source) — 2
- L112 #2933,3005 — 3
- L113 (no inline source) — 4
- L114 (no inline source) — 5
- L115 #3005 — 6
- L116 #3005 — 7
- L117 #3005,3029 — 8
- L119 (no inline source) — Name every terminal by its handle on every command — without `--terminal` a command acts on the active termina
- L123 #3005 — Only a lane has this section (Issue #3005 P3b)
- L125 (no inline source) — 1
- L126 (no inline source) — 2
- L127 (no inline source) — 3
- L128 #3005 — 4
- L129 #3040 — 5
- L131 (no inline source) — Not measured: no `ask`, `gate-create` or `worker_done` from a lane was run on this machine (only the help text
- L135 (no inline source) — Not run yet (`not_run`): in a clean clone with a registered ledger and one `ready` card, enter with `/moai --m
- L139 (no inline source) — A hand-off completes a card, not the round (`keep-going` step)

## Mechanical inventory — workflows/mode-leader.md (listing only; its edit is the last slice)

41 rule sentences (extractor: C-103 scratch `extract.py`; source = the issue the sentence itself cites; enforcer: a phrase of the sentence pinned by `tests/mode_workflow_steps.rs` and the file-specific pin tests, else `text only`).

- L7 (no inline source) — GitHub Issues are no longer the work ledger; the project's local cards are (user decision 2026-11-16, card C-5
- L9 (no inline source) — - **Transition exception**: until C-53 S3 lands, `moai gate-check` asks for an Issue claim (`Moai-Claim`) that
- L10 (no inline source) — - **Stranded cards**: `card next` shows a `review` card that nobody holds as `review_no_holder`, and a card wh
- L14 #2899 — The `--autonomy` level (`full` or `semi`; `semi` when omitted), the project's cards (`moai card list`, `moai c
- L18 (no inline source) — A mode banner, a read-only state summary, one `/goal` line for the user to paste, a decision record for every
- L24 #2899,2905 — 1
- L25 (no inline source) — 2
- L26 #2928 — 3
- L27 #2679 — 4
- L28 #2899,3005 — 5
- L29 (no inline source) — 6
- L30 #1975,2515,2612,2879,2899,2902,2933,3005 — 7
- L31 #1975,2515,2612 — 8
- L32 #1975,2385,2388,2401,2515,2664,2665,2680,2681,2693,2708,2777,2906 — 9
- L33 #2668 — 10
- L34 #2619,3005 — 11
- L38 (no inline source) — Keep-set decisions go to the user at every autonomy level (SPEC §6)
- L40 #3052 — A control line in a pull request comment (audit reservation or verdict, gate marker) counts only when a reposi
- L42 (no inline source) — Label every Issue you open in the creating command: exactly one `type:`, exactly one `priority:P0..P3`, one or
- L44 #2401 — Before dispatching a card, read the receiving lane's cards (`moai card show`); before starting an audit agent,
- L48 #3005 — When the host summarises your context or you were cleared, the newest user message you can still see is what s
- L52 #3005 — A leader that runs on Codex CLI, OMP or Claude Code follows the same six stages for an instruction it receives
- L56 (no inline source) — 3
- L58 (no inline source) — 5
- L62 #2516 — Use Orca only when the user chose it for this project and `orca` answers on this machine; installed is not cho
- L64 (no inline source) — 1
- L65 (no inline source) — 2
- L66 #2933,3005 — 3
- L67 (no inline source) — 4
- L68 (no inline source) — 5
- L69 #3005 — 6
- L70 #3005 — 7
- L71 #3005,3029 — 8
- L73 (no inline source) — Name every terminal by its handle on every command — without `--terminal` a command acts on the active termina
- L77 #3015 — A leader with no timer of its own keeps its rounds going by typing the round instruction into its own terminal
- L81 #3005 — Only the leader creates Runs (Issue #3005 P3a); without Orca, or with no Run path that works for a lane, the `
- L83 #3005 — 1
- L84 #3005,3029 — 2
- L88 (no inline source) — Not run yet (`not_run`): in a clean clone with a registered ledger, open a new Claude Code session, enter with
- L92 (no inline source) — Judge completion only from evidence on the remote and in the ledger (merged pull requests, verdicts, `done` ca
- L94 #2678 — Every row of a status, close or decision table you post carries a source: the command whose output gave its nu
