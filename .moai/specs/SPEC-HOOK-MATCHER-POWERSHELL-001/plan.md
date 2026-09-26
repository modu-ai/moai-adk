# Plan — SPEC-HOOK-MATCHER-POWERSHELL-001

Card t1224. Tier M. Methodology: TDD (RED per converted branch first). Milestones are ordered by decision reversibility — the decisions most likely to change come first. Version 0.2.0 (plan-audit iter-1 FAIL resolved; see spec.md HISTORY).

## Open Decisions (resolve before Implementation Kickoff)

| ID | Decision | Options | Recommendation | Owner |
|---|---|---|---|---|
| D1 | Matcher form for the PreToolUse `handle-pre-tool.sh` registration | (A) add a separate block `"matcher": "PowerShell"` pointing at the same wrapper; (B) widen line 58 to `Write\|Edit\|Bash\|PowerShell` (the vendor doc form) and amend both the exact-equality assertion and the RED fixture in `agent_model_matcher_test.go` (:130-135, :145-150), which forbid widening that exact string | (A) — purely additive, keeps a tested invariant intact; a PowerShell call matches exactly one block, so the hook still runs once per call. (B) is equally valid mechanically but edits a guard written to stop a different widening | Operator |
| D2 | Unclassifiable-command policy for the branch guard and the integration lock (REQ-HMP-010) | (A) **unknown → allow and log**: allow, and append one line (`powershell-unclassified`, construct, command, session) per guard: **branch guard** → its existing audit log `.moai/logs/branch-guard-audit.log` (`branch_guard.go:47`); **integration lock** → it has no audit log today (`integration_lock_guard.go`, grep audit/.jsonl = 0), so (A) creates `.moai/logs/integration-lock-audit.log` (rooted at the handler's projectDir, like `branch-guard-audit.log`) in the same line format as the branch-guard log. (B) **unknown → deny**: deny with a reason naming the construct, on both guards. (C) allow silently | (A) — both guards are documented as fail-open, deny-on-positive-evidence (`main-checkout-branch-guard.md` § Mechanical Enforcement); the Bash path's own `eval "git switch x"` is not classified either, so (A) keeps parity and adds visibility. Cost of (A): one new log file for the integration lock. (B) would make PowerShell stricter than Bash and can block a legitimate `& git status` | Operator |
| D2-deny-list | Same question for the destructive deny/ask lists | — | Not a decision: those lists stay deny-on-positive-match on both tools; a quote-collapse misparse can under-match (residual risk), never newly deny text that does not match | Fixed |
| D2-encoded | Any argument pwsh resolves to `-EncodedCommand` (M0-measured set; `-enc` / `-ec` / `-e` are examples) | — | Not a decision: unclassifiable regardless of payload, never decoded, handled by D2 (spec REQ-HMP-010) | Fixed |
| D2-slot | Slot-lease guard under indirection | — | Not a decision: keeps its existing fail-open, match-on-scrubbed-text behaviour (spec §A.3 site 4) | Fixed |
| D3 | PostToolUse sites 5–6 (unreachable from the Claude template) | (A) convert to the predicate for parity, no matcher change; (B) leave untouched | (A) — one-token change, keeps the REQ-HMP-005/013 source guard closed-world; reachability via the Codex adapter is unmeasured | Lead |
| D4 | Sibling gaps outside this SPEC | Issue cards for (i) `permission/stack.go` `IsWriteOperation` PowerShell parity, (ii) hook runtime on Windows without Git Bash (`bash -c` wrapper) | Issue both as separate cards | Operator |
| D5 | LIVE arm C (branch-guard denial through the branch binary) | Run arm C with the provenance controls of M5 / skip arm C and rely on arms A–B + unit tests | Run it, best-effort; INCONCLUSIVE is an acceptable delivered outcome, FAIL is a blocker | Lead |
| D6 | Pre-tool wrapper's Bash Risk-Amplifier warning on PowerShell calls (REQ-HMP-002, spec §A.3 site 14). This changes distributed template behaviour | (A) **keep Bash-only**: the gate stays `"Bash"`; the comment states the matcher scope and why PowerShell is excluded (backtick = PowerShell escape/line continuation, `$(` = subexpression, so the count is not the Bash doctrine's subcommand count); the REQ-HMP-013 wrapper scan carries an exclusion entry with that reason. (B) **extend to PowerShell with the same counter**: gate on `Bash` or `PowerShell`; accept that backticks and `$(` inflate the count (warn-only — never blocks, but may false-warn on escaped or continued PowerShell lines). (C) extend with a PowerShell-specific counter — rejected: requires PowerShell grammar (spec §D) | (A) — the doctrine the warning enforces is written for POSIX shell; (B) produces a count whose meaning differs by tool. Either way no call is blocked | Operator |

## Milestones

### M0 — Absorb, re-measure, capture the payload (Priority High)

- Absorb local `develop` into `WT-hook-matcher-powershell`; record the pre-merge HEAD and the merge SHA. After absorption `BASE` for this card is `$(git merge-base develop HEAD)`, re-evaluated at read time and valid only until the card merges into develop.
- Re-run `grep -n '"Bash"'` over `internal/hook`, `internal/cli/hook.go`, `internal/codexadapter`, `internal/permission`; `grep -n 'tool_name' .claude/hooks/moai/*.sh internal/template/templates/.claude/hooks/moai/*.sh.tmpl`; `grep -n '"matcher"'` over the template. Record updated §A.3 line numbers in `progress.md` §E.2 (not in spec.md).
- Re-run `git diff --name-only develop...WT-dual-harness-parity-rebuild`; if it now touches `internal/hook/*`, `settings.json.tmpl`, `.claude/hooks/moai/*`, or `runHarnessObserve`, stop and return a blocker report naming the overlapping hunk.
- **Indirection classification (fixes the REQ-HMP-010 set).** Run the unchanged branch-guard and integration-lock matchers against `& git switch x`, `iex "git switch x"`, `Start-Process git -ArgumentList 'switch','x'`, `pwsh -EncodedCommand <b64>`, and the corresponding `git merge` forms; record which the existing parser already classifies. Classified constructs follow REQ-HMP-007/008; the rest are the REQ-HMP-010 set.
- **Encoded-command spellings (REQ-HMP-010).** Before M1, from a scratch directory outside the worktree session (the worktree-isolation guard refuses `pwsh` inside it), run `pwsh -NoProfile -NonInteractive <spelling> <base64 of 'Write-Output ok'>` for candidate spellings (at least `-e`, `-ec`, `-en`, `-enc`, `-enco`, `-encodedc`, `-EncodedCommand`, mixed case, and the `--` / `/` prefixed forms); record in `progress.md` §E.2 which spellings execute the payload. The accepted set pins the M4 encoded-command tests.
- **Bash controls for REQ-HMP-009.** Run the pre-tool and post-tool handlers on Bash payloads with `tool_input` = `"x"`, `[1]`, `null`, `{}`, `{"command": 42}`; record the outputs as the expected values for the PowerShell cases (measured 2026-09-26 by reading only: `extractBashCommand` / `checkBashCommand` return "" on a type mismatch, so allow is the expected Bash output — M0 confirms by running).
- **LIVE arms A and B (payload capture, REQ-HMP-016).** Run M5 arms A and B now: they need only a logger, not the branch binary. If the captured `tool_input` lacks a string `command`, or `tool_name` is not `PowerShell`, stop and return a blocker report before M1. If arm B is INCONCLUSIVE (no payload captured, or a validity condition fails), record that and proceed: M1–M3 run on REQ-HMP-009 fail-safe handling, with no live-path claim for REQ-HMP-006/011.
- Confirm `pwsh` and `claude` versions.

### M1 — Registration and wrapper (D1, D6) (Priority High)

- RED: template parity test (REQ-HMP-012) — parse the rendered settings, key each registration by (event, hook script) where hook script = final `args` element with `${CLAUDE_PROJECT_DIR}/` stripped (or `command` without `args`), collect matcher names, and assert every pair naming `Bash` has a `PowerShell` counterpart for the same script or an exclusion entry with reason; assert no stale exclusion. Run against the unchanged template → FAIL naming `PreToolUse` and `.claude/hooks/moai/handle-pre-tool.sh`.
- GREEN: apply D1 to `settings.json.tmpl` and local `.claude/settings.json`; `make build`.
- Mutants after GREEN, each must turn the test red: (m1) remove the PowerShell registration; (m2) on a synthetic settings fixture carrying a second PreToolUse script (the real template has none — every PreToolUse entry points to `handle-pre-tool.sh`), register `PowerShell` on that second script while `handle-pre-tool.sh` lacks it.
- If D1 = (B): amend `agent_model_matcher_test.go` assertion and RED fixture to the new exact string, with a comment stating the widening is by PowerShell, not Agent.
- Wrapper (REQ-HMP-002): RED test drives the rendered wrapper with a stub `moai` on `PATH` and a PowerShell payload whose command has more than 5 metacharacters; assert the warning is absent (D6=A) or present (D6=B), plus a Bash control asserting the unchanged warning. Update the template `.sh.tmpl`, the local `.sh`, and the matcher-scope comment together.

### M2 — Shared predicate + source guard (Priority High)

- New file in `internal/hook` exposing `IsShellTool(name string) bool` (closed set `Bash`, `PowerShell`), used by the CLI through the package export.
- RED: source guard test (REQ-HMP-013) — (i) parse non-test Go files of `internal/hook` and `internal/cli/hook.go` with `go/ast`, flag every `BasicLit` string whose value equals `bash` case-insensitively outside the predicate file (comments are not literals, so they are excluded by construction); (ii) scan the template and local pre-tool wrappers for a `tool_name` match naming `Bash` without `PowerShell`. Exclusion list with reasons; stale entries fail. Record the base hit count (expected ≥ 10 Go hits plus the wrapper hit) as the RED-now cell; a count of 0 on base means the scan is blind and blocks M2.
- Positive control: a fixture declaring `const x = "Bash"` compared later, and a fixture using `strings.EqualFold(name, "bash")`, are both flagged.

### M3 — Per-site conversion, RED first (Priority High)

Precondition: M0 payload capture recorded (REQ-HMP-016), or arm B recorded INCONCLUSIVE — in which case M1–M3 proceed on REQ-HMP-009 fail-safe handling with no live-path claim for REQ-HMP-006/011. For each Go site 1–10 in spec §A.3: write a failing test with a `tool_name: "PowerShell"` payload that the Bash path already handles (deny, ask, branch-guard deny, integration-lock deny, slot-lease deny, evidence record, zero-execution advisory), observe it FAIL (allowed / no record), convert the condition to `IsShellTool`, observe PASS. Pair each with the Bash control asserting the unchanged Bash outcome, and the REQ-HMP-009 defect cases against the M0-recorded Bash controls.

- Every SPEC test is named with the prefix `TestHMP`, sets `t.Setenv("MOAI_HOME", t.TempDir())`, and does not call `t.Parallel` (`t.Setenv` panics under it).
- Branch guard / integration lock tests use a temp git repo as primary checkout and config enabled through the test ConfigProvider.
- Slot-lease tests: lease records under the per-test `MOAI_HOME` only (t1229).
- CLI-level tests go in a new `internal/cli/*_test.go` file (t1099 edits `hook_test.go`).

### M4 — D2 unclassifiable policy (Priority Medium)

- RED then GREEN per D2 outcome for each construct M0 found unclassified, on both the branch guard and the integration lock. Under D2=(A): allow + exactly one line in the guard's named log (branch guard → `branch-guard-audit.log`; integration lock → `integration-lock-audit.log`). Under (B): deny with a construct-naming reason. `pwsh -EncodedCommand` with a payload that decodes to a harmless command still takes the D2 path (proves no decoding). A Bash control (`eval "git switch x"`) keeps its current outcome.

### M5 — Isolated LIVE measurement arm C (best-effort, Priority Medium)

**Declared caps (fixed before execution, whole LIVE budget):** 3 arms total (A and B run in M0, C here), 1 run per arm, `--max-turns 3`, outer `timeout 180`, no retries, foreground only, no background processes. Each arm runs in its own fresh scratch directory created outside the repository and is launched from that directory.

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
| A (old, M0) | PreToolUse matcher `Write\|Edit\|Bash` → a logger command that appends its stdin to `<arm>/hook.log` | Run `Get-Location` with the PowerShell tool | PowerShell call recorded in stream-json; `hook.log` absent or empty |
| B (new, M0) | Same logger, matcher per D1 | Same | `hook.log` holds ≥ 1 payload with `tool_name` `PowerShell`; record its `tool_input` / `tool_response` keys (REQ-HMP-016) |
| C (guard, M5) | Hook command = `bash -c 'moai version > <arm>/binary-version.txt; exec bash <arm>/.claude/hooks/moai/handle-pre-tool.sh'` (wrapper rendered from this branch). Launch with `PATH="<branch bin dir>:$PATH"`, `MOAI_HOME=<arm>/moai-home`, `MOAI_HOOK_STDERR_LOG=<arm>/.moai/logs/hook-stderr.log` (inside the wrapper's allowlisted `$CLAUDE_PROJECT_DIR/.moai/logs/` prefix). Scratch dir is a git repo with `.moai/config/sections/workflow.yaml` `branch_guard.enabled: true` | Run `git switch -c probe` with the PowerShell tool | Denial carrying `BRANCH_GUARD_VIOLATION:`; `git branch --list probe` empty afterwards |

**Validity conditions (any failure → INCONCLUSIVE, deliver without LIVE claim):** `system/init` tools list is `PowerShell` alone; the recorded tool call is a PowerShell call; arm A shows the call executed; no arm exits 124 or hits `--max-turns`; `pwsh` 7+ on PATH; `MOAI_BRANCH_GUARD_EXEMPT` unset in the launch environment; for arm C, `binary-version.txt` exists and its `-g<sha>` suffix equals the short SHA of the branch HEAD the binary was built from.

**Outcomes:** PASS — all validity conditions hold and every expected observable is present. FAIL — all validity conditions hold and any expected observable is absent or contradicted (for example arm A `hook.log` non-empty, arm B no PowerShell payload, arm C no denial or `probe` created); recorded with arm and observable, and returned as a blocker. INCONCLUSIVE — a validity condition fails; recorded naming it. Raw outputs saved under `.moai/reports/t1224/live/`.

### M6 — Close (Priority Low)

- Scoped verification (env-scrubbed form in spec §C), `go vet ./internal/hook/... ./internal/cli/...`, `golangci-lint run ./internal/hook/... ./internal/template/... ./internal/cli/...`, `GOOS=windows GOARCH=amd64 go build ./...`.
- Verdict at `.moai/reports/t1224/verdict.md` with the per-site table, D1–D6 outcomes, and LIVE outcome (PASS / FAIL / INCONCLUSIVE).

## Risks

| Risk | Mitigation |
|---|---|
| PowerShell hook `tool_input` lacks `command` or names it differently | Transcript-level t1211 evidence says `{command, description}` (spec §A.1); M0 arm B captures the real hook payload and stops before M1 if the key differs; REQ-HMP-009 fail-safe covers the rest |
| Quote-collapse misreads PowerShell backslash paths and hides a later segment | Recorded residual risk; deny lists never newly over-deny; D2 covers git indirection |
| Concurrent edits from t1099 land in the same regions | M0 re-measure gate with blocker on overlap; new CLI tests in a new file |
| A hook test touches the real lease DB (t1229) | REQ-HMP-014 + AC-HMP-011 static check with positive control |
| D1=(B) silently drops the Agent-widening guard's intent | Amended assertion keeps a comment and the RED fixture for Agent widening |
| D6=(B) false-warns on escaped PowerShell lines | Warn-only path; recorded as known limitation if chosen |
| LIVE arm C runs the installed binary instead of the branch build | PATH pin + `binary-version.txt` SHA validity condition |

## Anti-Patterns

- Widening the matcher without converting the Go branches or the wrapper (no effect, false sense of coverage).
- Scattered `ToolName == "Bash" || ToolName == "PowerShell"` pairs instead of the predicate.
- Claiming live hook coverage from unit tests alone, or reading a LIVE non-denial as PASS without the binary-provenance condition.
- Running `go test ./...` locally.
