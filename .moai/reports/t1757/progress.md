# t1757 progress (lane-39)

## t1808 live-acceptance first check
- Session model: claude-sonnet-5-5 (stated by the runtime system prompt; not independently measured).
- Session effort: NOT observed. CLAUDE_CODE_EFFORT_LEVEL is unset in the env and no tool reports it; "medium" is only the documented default. Gap.
- First spawn (manager-develop), row in the primary checkout's `.moai/logs/agent-model-audit.jsonl` (the log is not in this worktree), verbatim:
  {"timestamp":"2026-10-15T21:29:30Z","session_id":"ea140374-cc3f-4c19-b515-5c5570c06ba1","agent":"manager-develop","declared_model":"","resolved_model":"inherit","verdict":"inherit","override_consumption":"off"}

- Gap (leader clarification): the audit row's `declared_model` records only the Agent call argument, not the agent frontmatter `model`. `""` / `inherit` is therefore not a defect; this log cannot decide which model the spawn actually ran on.

## Run phase
- Evidence: CC 2.1.299 changelog line 229 (cc-changelog-snapshot-2.1.300.md): hooks that set no shell now block when Git for Windows is absent. MoAI hooks are exec-form `bash` (40 entries in settings.json.tmpl).
- Windows effect: NOT measurable on this Mac (darwin) -> Gap. Judged from docs and code; the doctor warning is advisory.
- Delegated to manager-develop: Windows-only `moai doctor` check (bash on PATH vs bash exec-form hooks). Result: committed 754752a1d.

## t1808 discriminator (first sync-auditor spawn, no model arg)
- Subagent transcript (task output a9069a34015d7c25e) `"model"` values: claude-sonnet-5-5 x65; one stray `"model":"claude-opus-5-5"` occurrence whose context I could not extract (Gap, probably quoted text).
- Hook row, verbatim: {"timestamp":"2026-10-15T21:55:44Z","session_id":"ea140374-cc3f-4c19-b515-5c5570c06ba1","agent":"sync-auditor","declared_model":"","resolved_model":"inherit","verdict":"unknown","source":"subagent_stop","agent_id":"a9069a34015d7c25e","served_models":["claude-sonnet-5-5"],"self_reported_model":"claude-sonnet-5-5","cause":"no declared model — the subagent inherits the session model"}
- Reading: served model sonnet-5-5, not opus-5-5. The session model is also sonnet-5-5, so plain session inheritance gives the same value; this cannot by itself prove which definition file was loaded.

## Sync audit round 1
- FAIL (card-review.md, commit 754752a1d): B1 doctor goldens break on Windows (runtime.GOOS vs the override seam); B2 the warning over-claims for exec-form hooks. Fix round delegated to manager-develop.
- Round 2 (cumulative 754752a1d + 226e99de0): PASS-WITH-DEBT, codex gate pass, receipts rcpt-354314c19db771c13207c77d / rcpt-2502460cc153b79fa357194b. B1/B2 closed. Debt: O1 (LookPath uses doctor PATH, not hook PATH), O2 (counts all `bash` commands regardless of event/args/shell), O3 (no docs-site page for the new check). Gap: real Windows behaviour not measured (darwin host).
- Stage record: run v7 and sync v8 recorded with 754752a1d; sync-audit needs `<sha>:<artifact>` evidence, to be recorded once card-review.md is committed.
