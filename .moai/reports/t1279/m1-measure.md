# M1 measurement evidence — SPEC-SESSION-DOUBLELOAD-001 (card t1279)

Measured 2026-09-29, this tree (`WT-prose-residuals` @ `13aac1598`, probes launched from the
/tmp/t1279-m1 scratchpad only), Claude Code 2.1.284, tmux 3.6a. Caps: `.moai/reports/t1279/m1-caps.md`
(committed before any probe output). Raw evidence: `probes/commands.txt` (verbatim launch lines),
`probes/extracts.txt` (verbatim extractor outputs), `probes/transcripts/` (session transcript copies),
`probes/*.jsonl` + `probes/*.debug` (headless stream output + debug logs).

probes_run: 5
probe_launches: 6
skill_dup_tokens: gap
skill_dup_method: first-turn-input-delta
clear_restores_skill_set: gap
clear_method: none
extractor_skills: jq -r 'select(.subtype=="init") | .slash_commands[]' <transcript> | grep -c '^fixture-'
extractor_instructions: grep -oE 'Contents of [^"\]+(CLAUDE\.md|CLAUDE\.local\.md)' <transcript> | sort -u
probe_session_id: e1fe10e4-8606-41c0-a370-1a2f301efb67 (probe 1, first run — superseded by the re-instrumented run)
probe_session_id: d2e8b483-8021-494b-a50f-64d9213d1952 (probe 1, re-instrumented run)
probe_session_id: f533a0a0-5425-4acf-a007-f39770990550 (probe 2)
probe_session_id: cc035f13-6e85-407f-a69d-6969cae356b7 (probe 3, attempt 1)
probe_session_id: 765d59a3-5d01-45b8-9fd4-1427d8e3ff96 (probe 3, spare-probe retry)
probe_session_id: 696b29d1-2fcf-4bb9-a203-7a69d13b2198 (probe 4)
pgid=18002 (probe 1, re-instrumented run; invocation process group)
pgid=39342 (probe 2; invocation process group)
pgid=78638 (probe 3; tmux pane process group, retry session)
pgid=80346 (probe 4; invocation process group)
pgid=25674 (probe 5 positive control; extractor invocation process group)
primary_branch_before: main
primary_branch_after: main
positive_control: probes/transcripts/probe5-fixture-two-sources.jsonl (two skill sources); re-running the recorded extractor_skills on it prints 4 (fixture-wt 2, fixture-outer 2), i.e. >= 2

Launch accounting: the five matrix probes consumed 6 launches — probe 1 ran twice (first run
predates the process-group capture segment and is superseded by the identical-geometry
re-instrumented run; its session id is retained above for completeness), probe 3 ran twice (the
declared spare retry after attempt 1 hit its in-session timeout). This is exactly the 6-probe cap
committed in m1-caps.md; probe 5 is the extractor positive control (no claude process).

## Claim

1. The mid-session-move skill-list duplication did not occur in any probe of the declared mode
   (headless or interactive, CC 2.1.284, this environment): no probe's move attached a second
   skill list, so the REQ-SDL-002 token cost is unobservable in this mode and is recorded as a Gap.
2. A real `/clear` could not be executed within the declared caps by the tmux-driven interactive
   probe (TUI slash-menu fill swallows the Enter), so the REQ-SDL-003 restoration question is
   recorded as a Gap with `clear_method: none`.
3. The verdict §① Claim 2 instruction-file shape IS transcription-reproduced: a session launched
   inside the nested worktree loads the worktree's two instruction files PLUS the primary
   checkout's `CLAUDE.local.md`, while the primary `CLAUDE.md` is skipped.
4. Launcher-start skill single-source (t1219 Evidence 1, previously debug-log based) is now
   confirmed transcription-based: the in-worktree launcher-start session lists exactly one
   project skill source (40 worktree fixtures, 0 primary fixtures).

## Evidence

- Probe 1 (launcher start inside w1; transcripts `p1-launcher-start-w1.jsonl`,
  `p1b-launcher-start-w1-reinstrumented.jsonl`): first-turn input sums 28,790
  (28150 + 0 + 640) and 28,791 (28727 + 0 + 64) — the two runs agree within 1 token.
  Skill set (extractor_skills): 40 fixture-wt, 0 fixture-outer.
  Instruction set (extractor_instructions): w1/CLAUDE.local.md, w1/CLAUDE.md, outer/CLAUDE.local.md.
- Probe 2 (outer start, EnterWorktree into w1, tool result "Entered worktree at .../w1 on branch
  wt1"; transcript `p2-primary-start-moved.jsonl`): init skill set 12 fixture-outer / 0 fixture-wt;
  fixture-wt mentions in the whole post-move transcript: 0. Instruction set: outer/CLAUDE.local.md,
  outer/CLAUDE.md — the move added no worktree instruction files. Final-request input sum 54,883
  (53155 + 0 + 1728).
- Plan-formula delta (skill_dup_method first-turn-input-delta): probe 2 final-request 54,883 minus
  probe 1 first-turn 28,790 = 26,093. This figure is recorded but NOT attributed to skill
  duplication: the duplicated phenomenon (a second skill list) is absent from the transcript
  record (fixture-wt count 0), and the delta's composition is not decomposable from observable
  events (the largest transcript event is the 5,160-byte init; no post-move skill or instruction
  attachment appears anywhere). See Gaps.
- Probe 4 (control, outer start, no move; `p4-control-primary.jsonl`): first-turn input sum 27,238
  (27174 + 0 + 64); skill set 12 fixture-outer / 0 fixture-wt.
- Probe 3 (interactive, tmux): attempt 1 (`p3a-interactive-attempt1.jsonl`) — turn A marker
  P3-MOVED arrived; the `/clear` send was concatenated with the following prompt by the TUI
  slash-menu fill mechanism (single Enter inserts the command text without executing); the
  in-session `timeout 300` expired during the declared marker wait. Spare retry
  (`p3b-interactive-retry.jsonl`) — turn A marker arrived again; an 8 s settle still left `/clear`
  unexecuted (user message verbatim: "/clearReply with exactly P3-CLEARED and nothing else."), the
  model answered the mangled single prompt, no second transcript file opened, and subsequent pane
  sends returned API Error 400 from the serving backend. Probe slots then stood at the committed
  cap; REQ-SDL-005 applies.
- Positive control (probe 5): recorded extractor on the two-source fixture printed 4 (>= 2);
  both sources represented (2 + 2).
- Isolation: every recorded process group is dead — `pgrep -g` over all five recorded groups
  returns 0 lines. All six probe session ids appear in no `.moai/logs` or `.moai/state` file of
  either the t1279 worktree or the primary checkout (grepped, 0 hits each). The primary checkout
  branch was `main` before and after M1. The tmux sessions ended via
  `tmux kill-session -t <probe-unique-name>`; no process was killed by name.
- Environment note: the probes' `--model haiku` flag was declared per caps; the environment's
  auth source served model glm-5.3-flash for every probe identically (model-id line recorded in
  each stream file). The interactive probes ran with default model settings (same resolved model).
  Token figures rest on the serving backend's reported usage.

## Baseline-attribution

All probe launches, extractions, and sweeps in this file were executed in this run, against this
tree at HEAD `13aac1598` (branch `WT-prose-residuals`, t1279 worktree), with the scratchpad
`/tmp/t1279-m1` as the only probe subject location. Commands: `probes/commands.txt` (verbatim).
Measured environment: Claude Code 2.1.284, tmux 3.6a, macOS. Caps baseline: `m1-caps.md` @
`13aac1598`. The pre-flight baseline recorded in this run: hard clause count T=33 / L=33,
`cmp T L` byte-identical, t1175 (`7fe658815`) ancestor-of-HEAD exit 0, ANCHOR paragraph at
T:140 / L:140, `exactly once` count 0.

## Gaps

- `skill_dup_tokens: gap` — the REQ-SDL-002 phenomenon did not occur in the declared probe mode:
  no probe's mid-session move attached a second skill list (transcription-based fixture-wt count
  is 0 in every moved-session transcript, headless and interactive alike), so the duplication's
  token cost is unobservable here. The first-turn delta of 26,093 is recorded as Evidence but is
  not attributable to skill duplication. Per plan §E this does not block the doctrine milestone:
  the prohibition is ordered by the lead (verdict §④) and the measurement informs its note, not
  its existence. The verdict §⑩ byte figures (43,566 B listing, <= 20,829 tokens) remain the
  baseline from the t1279 session research; the 2,406 B frontmatter-first-line sum stays rejected
  as evidence of the token cost and is cited nowhere outside Gaps and this line's context.
- `clear_restores_skill_set: gap` with `clear_method: none` — probe 3 could not complete within
  the declared caps: attempt 1 expired its in-session timeout during the declared marker wait, and
  the spare retry exposed the TUI slash-menu fill mechanism (single Enter inserts `/clear` without
  executing; it concatenated with the following prompt in both attempts); a third attempt would
  exceed the committed 6-probe cap. The declared gap route (plan §E.1) applies verbatim: this is
  an expected outcome, AC-SDL-003 passes on the gap branch, and M2 proceeds carrying this Gap.
- Process-group record coverage: probe 1's first run and probe 3's attempt 1 predate the capture
  segment (headless) / were driven without a debug file (tmux), so no group number exists for
  those two launches; both invocations exited 0 (probe 1's first-run claude PID 16788 appears in
  its debug file and the process is gone). The pgid lines above record one group per matrix probe
  from its recorded run, and the live-group sweep returns 0.
- Model serving: the declared `--model haiku` was resolved by the environment's auth source to
  glm-5.3-flash for every probe; no haiku-served run exists in this environment, so no
  model-sensitivity check was possible within caps.

## Residual-risk

- Single environment (one CC version, one serving backend, one machine): load rules differ across
  versions and providers; the absence of move-attached skill duplication is a fact about
  CC 2.1.284 in this configuration, not a universal. The field observation (t1219 Evidence 4,
  interactive, real repo) and this fixture result now disagree in outcome; both are recorded.
- The token figures rest on the serving backend's usage report (cache_creation reported 0 on
  every probe; cache_read small); a backend that reports differently would shift the sums.
- The instruction-file Claim-2 reproduction uses one-line marker fixtures; the +20,855-token cost
  figure from the verdict's real-size copy was not re-measured (its geometry belongs to the
  t1243/t1259 work, out of scope here).
- The `/clear` restoration question remains open (Gap above); if a later probe executes `/clear`
  correctly (second Enter after the menu fill), the set comparison is ready to rerun within the
  same scratchpad geometry.
