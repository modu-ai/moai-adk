# Acceptance — SPEC-HOOK-MATCHER-POWERSHELL-001

Verification layer. Each criterion is binary. Test commands run in the env-scrubbed, scoped form of spec §C. `BASE` below always means `$(git merge-base develop HEAD)`, evaluated at read time after the M0 absorption and only before this card merges into develop (after the merge the range is empty and the check is vacuous).

## AC Matrix

| AC | Requirements | Evidence |
|---|---|---|
| AC-HMP-001 | REQ-HMP-001 | Rendered template and local settings both route PowerShell PreToolUse to `handle-pre-tool.sh`; `make build` exit 0 |
| AC-HMP-002 | REQ-HMP-003 | `git diff BASE..HEAD` shows no change to the PostToolUse `handle-post-tool.sh` matcher nor to any `permissions` rule; non-empty control |
| AC-HMP-003 | REQ-HMP-012 | Parity test RED on base, GREEN after M1, RED under both mutants (removed registration; PowerShell on a different script) |
| AC-HMP-004 | REQ-HMP-004, REQ-HMP-005, REQ-HMP-013 | Source guard RED on base (≥ 10 Go hits + the wrapper hit), GREEN after M3; positive controls flagged |
| AC-HMP-005 | REQ-HMP-006 | PowerShell deny/ask/`--no-verify` tests RED then GREEN; Bash controls unchanged |
| AC-HMP-006 | REQ-HMP-007 | Branch-guard PowerShell deny test RED then GREEN |
| AC-HMP-007 | REQ-HMP-008 | Integration-lock and slot-lease PowerShell deny tests RED then GREEN |
| AC-HMP-008 | REQ-HMP-009 | Wrongly-typed `tool_input` cases equal the M0-recorded Bash outputs, no panic; unparseable stdin keeps the t1152 fail-closed deny for both tool names |
| AC-HMP-009 | REQ-HMP-010 | Indirection tests pass under the D2 outcome on both guards; encoded command not decoded; destructive deny list unaffected by D2 |
| AC-HMP-010 | REQ-HMP-011 | Evidence record kind matches Bash for a PowerShell `go test` payload; unrecognized response shape records no pass |
| AC-HMP-011 | REQ-HMP-014 | Static check: every `TestHMP*` test and every listed modified test calls `t.Setenv("MOAI_HOME", …)` and no `t.Parallel`; positive control fails the check; swept count ≥ 1 |
| AC-HMP-012 | REQ-HMP-015, REQ-HMP-016 | `.moai/reports/t1224/live/` holds 3 arm outputs; payload keys recorded before M1; verdict records PASS, FAIL (blocker), or INCONCLUSIVE with the failed condition |
| AC-HMP-013 | REQ-HMP-002 | Wrapper test: PowerShell payload above the soft cap produces the warning per D6; Bash control unchanged; comment names the delivered matcher |

## Scenarios

### AC-HMP-001 — PowerShell reaches the pre-tool hook

Given the template and local settings after M1
When the settings are parsed and the PreToolUse registrations whose hook script is `.claude/hooks/moai/handle-pre-tool.sh` are collected
Then at least one registration's matcher names `PowerShell` in both files, and the two files agree
And `make build` exits 0

### AC-HMP-002 — No PostToolUse or permission change

Given the card branch after M6, before it merges into develop
When `git diff --name-only BASE..HEAD | wc -l` is run
Then it prints a number ≥ 1 (non-empty control; 0 means the range is unmeasured, not clean)
And `git diff BASE..HEAD -- internal/template/templates/.claude/settings.json.tmpl .claude/settings.json` contains no changed line inside the PostToolUse block registering `handle-post-tool.sh` and no changed line inside a `permissions` object

### AC-HMP-003 — Parity test discriminates

Given the base template where `Bash` is registered for `handle-pre-tool.sh` and `PowerShell` is not
When the parity test runs
Then it fails naming `PreToolUse` and `.claude/hooks/moai/handle-pre-tool.sh`
And after M1 it passes
And when the PowerShell registration is removed again it fails
And when `PowerShell` is instead added to a different PreToolUse script's matcher while `handle-pre-tool.sh` lacks it, it fails

### AC-HMP-004 — No scattered Bash literal

Given `internal/hook` non-test Go files, `internal/cli/hook.go`, and the template and local pre-tool wrappers
When the source guard runs on base
Then it reports at least the 10 Go comparisons of spec §A.3 and the wrapper's `tool_name` match (a lower count on base blocks the milestone as a blind scan)
And after M3 the only Go hit is inside the shared predicate file and every remaining hit is on the exclusion list with a reason
And fixtures `const x = "Bash"` and `strings.EqualFold(name, "bash")` are each flagged

### AC-HMP-005 — Deny list applies to PowerShell

Given a pre-tool payload with `tool_name` `PowerShell` and `tool_input.command` `terraform destroy`
When the pre-tool handler runs
Then the decision is deny with the same reason as the Bash payload carrying the same command

### AC-HMP-006 — Branch guard applies to PowerShell

Given a temporary primary-checkout git repo, branch guard enabled, a non-exempt agent, and `MOAI_HOME` set to a per-test temp dir
When a PowerShell payload carries `git switch -c probe`
Then the decision is deny and the reason starts with `BRANCH_GUARD_VIOLATION:`
And the same payload in a linked worktree is allowed

### AC-HMP-007 — Integration lock and slot lease apply to PowerShell

Given the integration lock enabled, a release worktree, and a lock record held by a live session other than the caller (fixture holder with a live pid and a different session id), with `MOAI_HOME` per-test
When a PowerShell payload carries `git merge --no-ff WT-x` from inside the release worktree
Then the decision is deny with the same reason text the Bash payload receives for the same command and fixture
And given the slot lease enabled with a resource whose command regex matches `go test ./internal/cli/...` and a live holder in another session under the per-test `MOAI_HOME`
When a PowerShell payload carries that command
Then the decision is deny with the same reason text the Bash payload receives
And both PowerShell tests fail (allowed) before the M3 conversion

### AC-HMP-008 — Defective PowerShell input is fail-safe

Given PowerShell payloads with a parseable envelope and `tool_input` equal to `"x"`, `[1]`, `null`, `{}`, and `{"command": 42}`
When the pre-tool and post-tool handlers run
Then neither panics and each output equals the M0-recorded output for the Bash payload with the identical `tool_input`
And given a truncated PreToolUse stdin whose visible prefix names `PowerShell`, and the same truncation naming `Bash`
When `moai hook pre-tool` reads each
Then both receive the same SPEC-HOOK-STDIN-FAILCLOSED-001 fail-closed deny

### AC-HMP-009 — Unclassifiable indirection follows D2

Given branch guard enabled in a primary checkout and a PowerShell payload using a construct M0 found unclassified (for example `iex "git switch probe"`)
When the pre-tool handler runs
Then under D2=(A) the call is allowed and exactly one `powershell-unclassified` line is appended to `.moai/logs/branch-guard-audit.log`
Or under D2=(B) the call is denied with a reason naming the construct
And given the integration lock held by another live session and a PowerShell payload `iex "git merge --no-ff WT-x"` in the release worktree
Then under D2=(A) the call is allowed and exactly one `powershell-unclassified` line is appended to `.moai/logs/integration-lock-audit.log`
Or under D2=(B) the call is denied with a reason naming the construct
And a payload `pwsh -EncodedCommand <base64 of "git status">` takes the same D2 path (the payload is not decoded)
And a PowerShell payload `terraform destroy` behind the same construct is still denied whenever the deny list matches its text

### AC-HMP-010 — Evidence parity

Given a post-tool payload with `tool_name` `PowerShell`, command `go test ./x/...`, and a Bash-shaped passing `tool_response`
When the evidence path records it
Then the recorded kind equals the Bash record for the same input
And given a `tool_response` whose shape is not recognized, no pass is recorded

### AC-HMP-011 — Hook tests isolate the MoAI home (static)

Given the test files touched by this card and the list of pre-existing tests the card modified (recorded in `progress.md` §E.2)
When the static check (a Go test using `go/ast`) collects every function named `TestHMP*` plus the listed modified tests
Then the collected count is ≥ 1 (an empty sweep is a failure, not a pass)
And every collected function calls `t.Setenv("MOAI_HOME", …)` with a `t.TempDir()`-derived value, directly or through a helper on the check's named helper list
And no collected function calls `t.Parallel`
And a positive-control fixture source containing a `TestHMP` function without that call makes the check report a violation

### AC-HMP-012 — LIVE has PASS, FAIL, and INCONCLUSIVE outcomes

Given the caps declared in plan M5 and three fresh scratch directories outside the repository
When arms A and B run once each in M0 and arm C runs once in M5, all in the foreground
Then the arm B payload's `tool_name`, `tool_input` keys, and `tool_response` keys are recorded in the evidence file before M1 begins
And if every validity condition holds (including arm C's `binary-version.txt` SHA equal to the branch HEAD the binary was built from) and arm A shows no hook line, arm B shows ≥ 1 PowerShell payload, and arm C shows a `BRANCH_GUARD_VIOLATION:` denial with no `probe` branch — the verdict records PASS
Or if every validity condition holds and any of those observables is absent or contradicted — the verdict records FAIL naming the arm and observable, and a blocker report is returned
Or if a validity condition fails — the verdict records INCONCLUSIVE naming that condition, with no LIVE claim made

### AC-HMP-013 — Wrapper follows D6

Given the rendered pre-tool wrapper, a stub `moai` first on `PATH` that exits 0, and `MOAI_HOOK_STDERR_LOG` under a per-test temp project's `.moai/logs/`
When a PowerShell payload whose command contains 6 metacharacters (`a | b | c | d | e | f | g`) is piped to it
Then under D6=(A) no `[moai:bash-risk]` warning is written, and under D6=(B) the warning is written
And a Bash payload with the same command still produces the warning
And the wrapper's matcher-scope comment names the matcher delivered by REQ-HMP-001 in both the template and the local copy

## Edge Cases

- `tool_name` casing: the predicate is exact-match (`PowerShell`), matching the vendor tool name; `powershell` is not a shell tool.
- PowerShell `2>&1` / pipeline to `Select-Object` after `go test`: evidence classification keys on the command prefix; a suffix does not change the kind.
- Compound PowerShell `git fetch; git switch main`: `;` is a separator in the existing splitter → branch guard sees `git switch`.
- `-e` short form: matched only as an argument of a `pwsh` / `powershell` invocation, never as a bare token elsewhere.

## Quality Gates

- Scoped tests green in the §C command form; `go vet` and `golangci-lint` clean on touched packages; `GOOS=windows` build exit 0; CI full suite green on the develop push.

## Definition of Done

- AC-HMP-001..011 and AC-HMP-013 PASS with verbatim command output in `progress.md` §E.2; AC-HMP-012 PASS or INCONCLUSIVE with reason (FAIL blocks delivery); D1–D6 outcomes recorded; verdict at `.moai/reports/t1224/verdict.md`.
