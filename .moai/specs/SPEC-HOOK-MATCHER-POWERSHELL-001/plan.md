# Plan — SPEC-HOOK-MATCHER-POWERSHELL-001

Card t1224. Tier M. Methodology: TDD (RED per converted branch first). Milestones are ordered by decision reversibility — the decisions most likely to change come first.

## Open Decisions (resolve before Implementation Kickoff)

| ID | Decision | Options | Recommendation | Owner |
|---|---|---|---|---|
| D1 | Matcher form for the PreToolUse `handle-pre-tool.sh` registration | (A) add a separate block `"matcher": "PowerShell"` pointing at the same wrapper; (B) widen line 58 to `Write\|Edit\|Bash\|PowerShell` (the vendor doc form) and amend `agent_model_matcher_test.go`, whose assertion forbids widening that exact string | (A) — purely additive, keeps a tested invariant intact; a PowerShell call matches exactly one block, so the hook still runs once per call. (B) is equally valid mechanically but edits a guard written to stop a different widening | Operator |
| D2 | Unclassifiable-command policy for the branch guard and integration lock (REQ-HMP-010) | (A) **unknown → allow and log**: allow, and append one audit line (`powershell-unclassified`, command, session) to the guard's existing audit log; (B) **unknown → deny**: deny with a reason naming the construct; (C) allow silently | (A) — both guards are documented as fail-open, deny-on-positive-evidence (`main-checkout-branch-guard.md` § Mechanical Enforcement); the Bash path's own `eval "git switch x"` is not classified either, so (A) keeps parity and adds visibility. (B) would make PowerShell stricter than Bash and can block legitimate `& git status` | Operator |
| D2-deny-list | Same question for the destructive deny/ask lists | — | Not a decision: those lists stay deny-on-positive-match on both tools; a quote-collapse misparse can under-match (residual risk), never newly deny text that does not match | Fixed |
| D3 | PostToolUse sites 5–6 (unreachable from the Claude template) | (A) convert to the predicate for parity, no matcher change; (B) leave untouched | (A) — one-token change, keeps REQ-HMP-005's source guard closed-world; reachability via the Codex adapter is unmeasured | Lead |
| D4 | Sibling gaps outside this SPEC | Issue cards for (i) `permission/stack.go` `IsWriteOperation` PowerShell parity, (ii) hook runtime on Windows without Git Bash (`bash -c` wrapper) | Issue both as separate cards | Operator |
| D5 | LIVE arm C (branch-guard denial) needs a binary built from this branch and a scratch MoAI project with `workflow.branch_guard.enabled: true` | Run arm C / skip arm C and rely on arms A–B + unit tests | Run it, best-effort; INCONCLUSIVE is an acceptable delivered outcome | Lead |

## Milestones

### M0 — Absorb and re-measure (Priority High)

- Absorb local `develop` into `WT-hook-matcher-powershell`; record BASE and merge SHAs.
- Re-run `grep -n '"Bash"'` over `internal/hook`, `internal/cli/hook.go`, `internal/codexadapter`, `internal/permission` and `grep -n '"matcher"'` over the template; update the §A.3 line numbers in `progress.md` §E.2 (not in spec.md).
- Re-run `git diff --name-only develop...WT-dual-harness-parity-rebuild` and `...WT-hook-stdin-failclosed`; if either now touches `pre_tool.go`, `post_tool.go`, `evidence_writer.go`, `settings.json.tmpl`, or `runHarnessObserve`, stop and return a blocker report naming the overlapping hunk.
- Confirm `pwsh` and `claude` versions for M5.

### M1 — Registration (D1) (Priority High)

- RED: template parity test (REQ-HMP-012) — parse the rendered settings, build (event, command) → matcher-name sets, assert every pair naming `Bash` has a `PowerShell` counterpart or an exclusion entry with reason; assert no stale exclusion. Run against the unchanged template → FAIL naming `PreToolUse handle-pre-tool.sh`. Mutant check: remove the PowerShell registration after GREEN → test FAILs again.
- GREEN: apply D1 to `settings.json.tmpl` and local `.claude/settings.json`; `make build`.
- If D1 = (B): amend `agent_model_matcher_test.go` to assert the new exact string, with a comment stating why the widening is by PowerShell and not Agent.

### M2 — Shared predicate + source guard (Priority High)

- New file in `internal/hook` exposing `IsShellTool(name string) bool` (closed set `Bash`, `PowerShell`), used by the CLI through the package export.
- RED: source guard test (REQ-HMP-013) — scan `internal/hook/*.go` (non-test) and `internal/cli/hook.go` for a tool-name comparison with `"Bash"` (`== "Bash"`, `!= "Bash"`, `case "Bash"`) outside the predicate file → FAIL on base (10 sites).

### M3 — Per-site conversion, RED first (Priority High)

For each site 1–10 in spec §A.3: write a failing test with a `tool_name: "PowerShell"` payload that the Bash path already handles (deny, ask, branch-guard deny, integration-lock deny, slot-lease deny, evidence record, zero-execution advisory), observe it FAIL (allowed / no record), convert the condition to `IsShellTool`, observe PASS. Pair each with the Bash control asserting the unchanged Bash outcome, and the defect cases of REQ-HMP-009 (unparseable input, missing/non-string `command`).

- Branch guard / integration lock tests use a temp git repo as primary checkout, config enabled through the test ConfigProvider, `t.Setenv("MOAI_HOME", t.TempDir())`.
- Slot-lease tests: lease records under the per-test `MOAI_HOME` only (t1229).

### M4 — D2 unclassifiable policy (Priority Medium)

- RED then GREEN per D2 outcome for `& git switch x`, `iex "git switch x"`, `Start-Process git -ArgumentList 'switch','x'`, `pwsh -EncodedCommand …`. Under D2=(A): allow + exactly one audit line; under (B): deny with construct-naming reason. A Bash control (`eval "git switch x"`) keeps its current outcome.

### M5 — Isolated LIVE measurement (best-effort, Priority Medium)

**Declared caps (fixed before execution):** 3 arms, 1 run per arm, `--max-turns 3`, outer `timeout 180`, no retries, foreground only, no background processes. Each arm runs in its own fresh scratch directory created outside the repository and is launched from that directory.

Common invocation (per arm, run from the arm's scratch dir):

```bash
CLAUDE_CODE_USE_POWERSHELL_TOOL=1 timeout 180 claude -p "<prompt>" \
  --setting-sources project --strict-mcp-config \
  --tools PowerShell --permission-mode bypassPermissions \
  --max-turns 3 --output-format stream-json --verbose
```

`--safe-mode` is NOT used: it disables hooks, which are the subject.

| Arm | Scratch `.claude/settings.json` hook | Prompt | Expected observable |
|---|---|---|---|
| A (old) | PreToolUse matcher `Write\|Edit\|Bash` → a logger command that appends its stdin to `<arm>/hook.log` | Run `Get-Location` with the PowerShell tool | PowerShell call recorded in stream-json; `hook.log` absent/empty |
| B (new) | Same logger, matcher per D1 | Same | `hook.log` holds one payload with `tool_name` `PowerShell`; record its `tool_input` / `tool_response` keys (REQ-HMP-016) |
| C (guard) | Real wrapper → binary built from this branch; scratch dir is a git repo with `.moai/config/sections/workflow.yaml` `branch_guard.enabled: true` | Run `git switch -c probe` with the PowerShell tool | Denial carrying `BRANCH_GUARD_VIOLATION:`; `git branch --list probe` empty afterwards |

**Validity conditions (any failure → INCONCLUSIVE, deliver without LIVE claim):** `system/init` tools list is `PowerShell` alone; the recorded tool call is a PowerShell call; arm A shows the call executed; no arm exits 124 or hits `--max-turns`; `pwsh` 7+ on PATH. Raw outputs saved under `.moai/reports/t1224/live/`.

### M6 — Close (Priority Low)

- Scoped verification (env-scrubbed form in spec §C), `go vet ./internal/hook/... ./internal/cli/...`, `golangci-lint run ./internal/hook/... ./internal/template/...`, `GOOS=windows GOARCH=amd64 go build ./...`.
- Verdict at `.moai/reports/t1224/verdict.md` with the per-site table, D1–D5 outcomes, and LIVE outcome (PASS / INCONCLUSIVE).

## Risks

| Risk | Mitigation |
|---|---|
| PowerShell `tool_input` lacks `command` or names it differently | REQ-HMP-009 fail-safe; M5 arm B captures the real payload; if the key differs, stop and report before M3 claims |
| Quote-collapse misreads PowerShell backslash paths and hides a later segment | Recorded residual risk; deny lists never newly over-deny; D2 covers git indirection |
| Concurrent edits from t1099 / t1152 land in the same regions | M0 re-measure gate with blocker on overlap |
| A hook test touches the real lease DB (t1229) | REQ-HMP-014; per-test `MOAI_HOME` |
| D1=(B) silently drops the Agent-widening guard's intent | Amended assertion keeps a comment and the RED fixture for Agent widening |

## Anti-Patterns

- Widening the matcher without converting the Go branches (no effect, false sense of coverage).
- Scattered `ToolName == "Bash" || ToolName == "PowerShell"` pairs instead of the predicate.
- Claiming live hook coverage from unit tests alone.
- Running `go test ./...` locally.
