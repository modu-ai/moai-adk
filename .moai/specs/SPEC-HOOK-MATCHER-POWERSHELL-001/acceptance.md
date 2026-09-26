# Acceptance — SPEC-HOOK-MATCHER-POWERSHELL-001

Verification layer. Each criterion is binary. Test commands run in the env-scrubbed, scoped form of spec §C with an isolated `MOAI_HOME`.

## AC Matrix

| AC | Requirements | Evidence |
|---|---|---|
| AC-HMP-001 | REQ-HMP-001, REQ-HMP-002 | Rendered template and local settings both route PowerShell PreToolUse to `handle-pre-tool.sh`; `make build` exit 0 |
| AC-HMP-002 | REQ-HMP-003 | `git diff BASE..HEAD` shows no change to the PostToolUse `handle-post-tool.sh` matcher nor to any `permissions` rule |
| AC-HMP-003 | REQ-HMP-012 | Parity test RED on base, GREEN after M1, RED again under the remove-registration mutant |
| AC-HMP-004 | REQ-HMP-004, REQ-HMP-005, REQ-HMP-013 | Source guard RED on base (10 sites), GREEN after M3 |
| AC-HMP-005 | REQ-HMP-006 | PowerShell deny/ask/`--no-verify` tests RED then GREEN; Bash controls unchanged |
| AC-HMP-006 | REQ-HMP-007 | Branch-guard PowerShell deny test RED then GREEN |
| AC-HMP-007 | REQ-HMP-008 | Integration-lock and slot-lease PowerShell deny tests RED then GREEN |
| AC-HMP-008 | REQ-HMP-009 | Defect-input tests (unparseable, missing `command`, non-string `command`) return the Bash-equivalent output, no panic |
| AC-HMP-009 | REQ-HMP-010 | Indirection tests pass under the D2 outcome; destructive deny list unaffected by D2 |
| AC-HMP-010 | REQ-HMP-011 | Evidence record kind matches Bash for a PowerShell `go test` payload; unrecognized response shape records no pass |
| AC-HMP-011 | REQ-HMP-014 | Every new/modified hook test sets `MOAI_HOME` to `t.TempDir()`; real MoAI home lease DB mtime unchanged across the run |
| AC-HMP-012 | REQ-HMP-015, REQ-HMP-016 | `.moai/reports/t1224/live/` holds 3 arm outputs and the verdict records PASS or INCONCLUSIVE with the failed validity condition |

## Scenarios

### AC-HMP-001 — PowerShell reaches the pre-tool hook

Given the template and local settings after M1
When the settings are parsed and the PreToolUse registrations for `handle-pre-tool.sh` are collected
Then at least one registration's matcher names `PowerShell` in both files, and the two files agree

### AC-HMP-003 — Parity test discriminates

Given the base template where `Bash` is registered for `handle-pre-tool.sh` and `PowerShell` is not
When the parity test runs
Then it fails naming `PreToolUse` and `handle-pre-tool.sh`
And after M1 it passes
And when the PowerShell registration is removed again it fails

### AC-HMP-004 — No scattered Bash literal

Given `internal/hook` non-test files and `internal/cli/hook.go`
When the source guard scans for `"Bash"` tool-name comparisons
Then the only match is inside the shared predicate file

### AC-HMP-005 — Deny list applies to PowerShell

Given a pre-tool payload with `tool_name` `PowerShell` and `tool_input.command` `terraform destroy`
When the pre-tool handler runs
Then the decision is deny with the same reason as the Bash payload carrying the same command

### AC-HMP-006 — Branch guard applies to PowerShell

Given a temporary primary-checkout git repo, branch guard enabled, a non-exempt agent, and `MOAI_HOME` set to a per-test temp dir
When a PowerShell payload carries `git switch -c probe`
Then the decision is deny and the reason starts with `BRANCH_GUARD_VIOLATION:`
And the same payload in a linked worktree is allowed

### AC-HMP-008 — Defective PowerShell input is fail-safe

Given a PowerShell payload whose `tool_input` is `{}` or `{"command": 42}` or invalid JSON
When the pre-tool and post-tool handlers run
Then neither panics and each output equals the output for the Bash payload with the same defect

### AC-HMP-009 — Unclassifiable indirection follows D2

Given branch guard enabled in a primary checkout and a PowerShell payload `iex "git switch probe"`
When the pre-tool handler runs
Then under D2=(A) the call is allowed and exactly one `powershell-unclassified` audit line is appended
Or under D2=(B) the call is denied with a reason naming `Invoke-Expression`/`iex`

### AC-HMP-010 — Evidence parity

Given a post-tool payload with `tool_name` `PowerShell`, command `go test ./x/...`, and a Bash-shaped passing `tool_response`
When the evidence path records it
Then the recorded kind equals the Bash record for the same input
And given a `tool_response` whose shape is not recognized, no pass is recorded

### AC-HMP-012 — LIVE is best-effort with an explicit inconclusive branch

Given the caps declared in plan M5 and three fresh scratch directories outside the repository
When arms A, B, C run once each in the foreground
Then either arm A shows no hook line, arm B shows one PowerShell payload, and arm C shows a `BRANCH_GUARD_VIOLATION:` denial with no `probe` branch created — recorded as PASS
Or a validity condition fails and the verdict records INCONCLUSIVE naming that condition, with no LIVE claim made

## Edge Cases

- `tool_name` casing: the predicate is exact-match (`PowerShell`), matching the vendor tool name; `powershell` is not a shell tool.
- PowerShell `2>&1` / pipeline to `Select-Object` after `go test`: evidence classification keys on the command prefix; a suffix does not change the kind.
- Compound PowerShell `git fetch; git switch main`: `;` is a separator in the existing splitter → branch guard sees `git switch`.

## Quality Gates

- Scoped tests green in the §C command form; `go vet` and `golangci-lint` clean on touched packages; `GOOS=windows` build exit 0; CI full suite green on the develop push.

## Definition of Done

- AC-HMP-001..011 PASS with verbatim command output in `progress.md` §E.2; AC-HMP-012 PASS or INCONCLUSIVE with reason; D1–D5 outcomes recorded; verdict at `.moai/reports/t1224/verdict.md`.
