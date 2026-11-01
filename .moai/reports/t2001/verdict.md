# t2001 verdict — Monitor commands vs the shell guards (Class B, lane-3)

Policy rules applied: `verification-claim-integrity.md` §1.1 surface 3, §2, §3; `verification-completeness.md` §1.1 (observed failure), §2.1; `factory-dispatch.md` § Card classes (Class B). Tree: branch WT-cc-bash-monitor, HEAD 4f542d28c5 on origin/main 5c3f0ebf27. Claude Code 2.1.316.

## Claim 1 — the hypothesis is REFUTED for the shipped hook registration
Hypothesis: a Monitor call reaches the Bash-matcher hook with tool_name `Monitor`, IsShellTool is false for it, and the shell guards skip the command.

**Evidence (run 2, fresh `claude -p`, subscription auth, `ANTHROPIC_API_KEY` unset — `test -n "$ANTHROPIC_API_KEY"` printed `key-unset`):** temporary hook on matcher `Bash` ONLY wrote raw stdin to a new empty file.
```
hook file t2001-hook2.jsonl, raw stdin rows verbatim (2 rows; matcher Bash only):
{"session_id":"29926ea2-dd81-4ce8-9523-752652e1fb74","transcript_path":"/Users/goos/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go--moai-worktrees-t2001/29926ea2-dd81-4ce8-9523-752652e1fb74.jsonl","cwd":"/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2001","prompt_id":"187985df-531b-4e99-b2ab-3330ae15b687","permission_mode":"default","effort":{"level":"medium"},"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"echo ctl-bash","description":"Print ctl-bash"},"tool_use_id":"toolu_01KPsgibyhN2JiRhs4tGGk7z"}
{"session_id":"29926ea2-dd81-4ce8-9523-752652e1fb74","transcript_path":"/Users/goos/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go--moai-worktrees-t2001/29926ea2-dd81-4ce8-9523-752652e1fb74.jsonl","cwd":"/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2001","prompt_id":"187985df-531b-4e99-b2ab-3330ae15b687","permission_mode":"default","effort":{"level":"medium"},"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"echo ctl-monitor","description":"echo ctl-monitor"},"tool_use_id":"toolu_01YYi91YwFbXM72hjvaf3dg8"}
child transcript (session 29926ea2): tool_use toolu_01YYi91YwFbXM72hjvaf3dg8 name = Monitor ; toolu_01KPsgibyhN2JiRhs4tGGk7z name = Bash
child tool_use name tally (grep line tally, may count one id twice — not an exact call count): Bash 2, Monitor 1, Read 2, ToolSearch 2; the two hook rows' ids were each resolved to their tool_use name individually (above)
```
Positive control: the Bash call is recorded. Negative controls: the Read and ToolSearch calls (4 tool_uses) produced no hook rows. The Monitor tool_use (id toolu_01YYi…) produced the second row with tool_name `Bash` and the command under `tool_input.command`. Therefore the Monitor command reaches the guards as a Bash call. IsShellTool("Bash") is true; no guard bypass through the shipped matcher group `Write|Edit|Bash`. No code change to `IsShellTool` or the matcher.

Consistent corroboration: the shipped `preToolHandler` trace of run 1's child logged 2 PreToolUse events, both `tool":"Bash"`, for one Bash and one Monitor call, with exactly one matcher group containing `Bash` in `.claude/settings.json`.

## Claim 2 — a matcher that NAMES Monitor delivers tool_name `Monitor` (not shipped)
Run 1 (fresh `claude -p`, session 20fbd1a2, matcher `Bash|Monitor`), the child's two raw stdin rows verbatim (rows 2 and 3 of t2001-hook.jsonl; row 1 and later rows are the lane session's own calls and are not part of this evidence):
{"session_id":"20fbd1a2-b790-41ed-ae94-9205cbb555db","transcript_path":"/Users/goos/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go--moai-worktrees-t2001/20fbd1a2-b790-41ed-ae94-9205cbb555db.jsonl","cwd":"/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2001","prompt_id":"26077a03-6e6c-4182-b1d9-f1c990a64789","permission_mode":"default","effort":{"level":"medium"},"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"echo ctl-bash","description":"Print ctl-bash"},"tool_use_id":"toolu_01DHqjzQ3iesicRhcuNMEAjU"}
{"session_id":"20fbd1a2-b790-41ed-ae94-9205cbb555db","transcript_path":"/Users/goos/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go--moai-worktrees-t2001/20fbd1a2-b790-41ed-ae94-9205cbb555db.jsonl","cwd":"/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2001","prompt_id":"26077a03-6e6c-4182-b1d9-f1c990a64789","permission_mode":"default","effort":{"level":"medium"},"hook_event_name":"PreToolUse","tool_name":"Monitor","tool_input":{"description":"ctl-monitor echo","timeout_ms":30000,"command":"echo ctl-monitor"},"tool_use_id":"toolu_01LAdz1gcW6fcGR8ErQ7GbNQ"}
That Monitor row (tool_name "Monitor", fields description/timeout_ms/command) is the matcher-names-Monitor case. Only matters if a project registers a Monitor matcher; `IsShellTool("Monitor")` stays false by design and the set was not widened.

## Claim 3 — regression test pinned and shown able to fail
`internal/hook/monitor_wireformat_test.go` (`TestMonitorCommandUnderBashMatcherIsGuarded`): observed payload → real decoder → IsShellTool → command read → real PreToolUse handler → `terraform destroy` denied.
```
go test ./internal/hook -run 'TestMonitorCommandUnderBashMatcherIsGuarded|TestHMPLiveWireFormat' -count=1 -v
--- PASS: TestMonitorCommandUnderBashMatcherIsGuarded (0.01s)   exit 0
mutant (fixture tool_name Bash -> Monitor):
--- FAIL: TestMonitorCommandUnderBashMatcherIsGuarded  monitor_wireformat_test.go:44: decoded tool_name "Monitor" is not a shell tool   exit 1
restored; go test ./internal/hook -run 'TestHMPSourceGuard|TestHMPTestIsolation|TestHMPLiveWireFormat|TestMonitorCommandUnderBashMatcherIsGuarded' -count=1 -v → 6 PASS
scoped gate: gofmt -l internal/hook/ (empty); go vet ./internal/hook/ (ok); golangci-lint run ./internal/hook/ → 0 issues;
go test ./internal/hook -run 'TestHMP|TestMonitor|TestIsShellTool|TestPreTool|TestHandlePreTool' -count=1 -timeout 240s → ok 10.042s
```
Swept set is non-empty (`-v` showed the named tests ran).

## Gaps (not observed)
- Upstream constraint "pre-write block only, rewrite not honoured" for Monitor was NOT tested (no third `claude -p`; cap 2 reached, 2 used). A deny on the Monitor-as-Bash path is shown only at handler level by the test, not through a live Monitor call.
- Run-1 in-session probes (3) and the negative control in the lane session were uninterpretable (hook not reloaded; negative control failed) and are NOT evidence.
- Refusals: the worktree-isolation guard refused 3 Bash invocations (compound heredoc write; `env -u … claude -p`; `unset … && claude -p` with redirect) and later 3 analysis commands (awk -v, sed $(…), `git grep`-style combos); the substitutions were plain commands/Read/Write. No result here rests on a refused command.
- Run 2's child made Bash twice (ids counted by grep; the second row is the Monitor id, verified by id); the first claude -p child's trace claim is corroboration only.
- Single OS/version: macOS, Claude Code 2.1.316; behaviour on other versions unmeasured.

### PR #1955 CI repair (lane-9, 2026-11-02) — scope and Gaps
Scope: the PR's own change (Claims 1-3 above) is unchanged. Two test-only commits were added because the PR's CI was red on tests this PR does not touch: `68b379bd5a` (reports-archive fixture shard; to be reverted before absorbing origin/main because #1956 carries the same fix more broadly) and `eb7c12b822` (inject `activeProcessCWDProbe` in the migrate rehearsal world; production code unchanged).
- Repair evidence (migrate): RED `env PATH=/opt/homebrew/bin:/usr/bin:/bin go test ./internal/cli/worktree -run TestMigrate_ApplyMovesEligibleAndSkipsProtected -count=1` -> `moved 0, skipped 6`, FAIL migrate_command_test.go:225 (macOS without lsof; same symptom as the Windows CI log). GREEN: same PATH with `-run 'TestMigrate|TestRollback|TestJournal|TestUpdateConsolidation'` -> ok; normal PATH -> ok. NOT run on Windows: the 14 Windows tests are verified only by the PR CI after push.
- NOT repaired, NOT PASS: `internal/hook` `TestWorktreeRegistry_TwoWritersAcrossProcessesLoseNeitherEntry` and `TestWorktreeRegistry_HeldLockPastLoweredBoundLapses` fail on the Windows runner ("a writer's update was lost" / "write succeeded while the lock was held past the bound"); cause not isolated (leader ruling: leave, record as Gap).
- No test pins the lsof-probe-failure skip ("cannot inspect active process working directories"): grep over internal/cli and internal/cli/worktree tests found none. That conservative-skip behaviour is untested; production code left unchanged per leader.
- Whether main itself is red on these Windows tests is unobserved (main has no Windows push CI).
- Guard refusal 2026-11-01T18:45:38Z: a `sed -n 845,870p` over the downloaded CI log was refused by the worktree-isolation guard (runtime-computed variable in an option position); not run. The Read tool read the same log lines instead. A later compound `command -v` + git form was also refused and re-issued as plain separate commands.

## Residual-risk
- The Monitor-as-Bash relabel is Claude Code behaviour, not MoAI's; a later runtime may deliver `Monitor`. The test pins today's shape and will not detect the runtime change by itself (it replays a fixture); the guard-reach claim is re-measured only by a live run.
- A project that adds a `Monitor` matcher would bypass shell guards (IsShellTool false). Not shipped; widening needs an operator decision (shell_tool.go comment).

## Card review
`.moai/reports/t2001/card-review.md` — codex pass, 0 findings (advisory).
