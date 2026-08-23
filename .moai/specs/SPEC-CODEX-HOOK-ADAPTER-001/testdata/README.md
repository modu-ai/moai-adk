# testdata — golden hook payloads

`hook-payloads/` holds Codex hook stdin captures, vendored here so REQ-6's golden tests are
executable from the branch and in CI.

**Origin**: captured during the M0 premise measurement and preserved at
`.moai/reports/t91/hook-payloads/` in the primary checkout. That directory is **untracked**, so it
resolves only on the machine that produced it — not in a worktree, not on a branch, not in CI. The
copies here are byte-identical and are the location the tests read.

**Provenance of each file** — `codex-cli 0.147.0`, captured by a hook that dumped stdin and exited
0, under an isolated `CODEX_HOME`. Paths and one encrypted blob were masked at capture time.

| File | Event | Notes |
|---|---|---|
| `PreToolUse.json` | PreToolUse | independently re-observed in round 2; key set matched exactly |
| `PostToolUse.json` | PostToolUse | |
| `SessionStart.json` | SessionStart | carries `source: startup` |
| `SessionEnd.json` | SessionEnd | no `model` / `permission_mode` fields |
| `Stop.json` | Stop | independently re-observed in round 3; `stop_hook_active` present |
| `UserPromptSubmit.json` | UserPromptSubmit | independently re-observed in round 2 |
| `subagent-PreToolUse.json` | PreToolUse | delegation path; `tool_name` begins `collaboration` |
| `subagent-PostToolUse.json` | PostToolUse | delegation path |

The two `subagent-*` files are why the `SubagentStop` mapping was retired: delegation surfaces as
ordinary tool calls, and `SubagentStart` / `SubagentStop` never fire in this build.

Three of these events were re-captured independently in rounds 2 and 3; those second captures live
in `.moai/reports/t83/probe/*.observed.json` and matched these key sets. The remaining five have a
single capture each.
