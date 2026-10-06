# t1471 verdict — Codex audit launcher reads roles from the embedded templates

Card: t1471 · GitHub #1735 · class B (no SPEC) · cycle_type: tdd
Branch: `WT-codex-role-embedded` · base: local develop `b5815ca80`
Commits: `842b11cc5` (RED test only) → `14c7af8c8` (fix)

## Claim

1. `codexAuditLoadRole` (internal/cli/codex_audit_launch.go) now reads the role file from
   `codexAuditRoleFS`, which defaults to `template.EmbeddedTemplates` — the same source
   `codexAuditLaunchableRoles` derives eligibility from. The project's `.codex/agents/moai/`
   copy is no longer read.
2. A stale project copy declaring `sandbox_mode = "workspace-write"` and a missing project copy
   both launch the embedded read-only `plan-auditor`, and the exec argv carries the embedded
   `developer_instructions`.
3. The read-only assertion still applies to the embedded content, and error messages still name
   the role. `TestCodexAuditLaunchRoleEligibility` still passes (contract-derived set; a tampered
   writer role is still refused).

## Evidence

RED, at `b5815ca80` plus the new test (commit `842b11cc5`):

```
$ go test -count=1 ./internal/cli/ -run 'TestCodexAuditLaunchRoleFromEmbedded|TestCodexAuditLaunchRoleEligibility' -v
--- PASS: TestCodexAuditLaunchRoleEligibility (11.48s)
=== RUN   TestCodexAuditLaunchRoleFromEmbedded/stale_workspace-write_project_copy
    codex_audit_launch_test.go:685: launch refused: code=1 stderr="codex audit plan-auditor: role \"plan-auditor\" file declares sandbox_mode \"workspace-write\", not read-only\n"
=== RUN   TestCodexAuditLaunchRoleFromEmbedded/no_project_copy
    codex_audit_launch_test.go:696: launch refused: code=1 stderr="codex audit plan-auditor: role \"plan-auditor\" has no emitted role file: open /private/var/folders/.../A1/.codex/agents/moai/plan-auditor.toml: no such file or directory\n"
--- FAIL: TestCodexAuditLaunchRoleFromEmbedded (8.43s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	21.757s
```

GREEN, with the fix (working tree identical to commit `14c7af8c8`):

```
$ go test -count=1 ./internal/cli/ -run 'CodexAudit|CodexRole' -v
--- PASS: TestCodexAuditLaunchArgv (8.51s)
--- PASS: TestCodexAuditLaunchRoleEligibility (19.53s)
--- PASS: TestCodexAuditLaunchRoleFromEmbedded (16.91s)
    --- PASS: TestCodexAuditLaunchRoleFromEmbedded/stale_workspace-write_project_copy (8.70s)
    --- PASS: TestCodexAuditLaunchRoleFromEmbedded/no_project_copy (8.20s)
--- PASS: TestCodexAuditLaunchVerbatimWrite (10.56s)
...
$ go test -count=1 ./internal/cli/ -run 'CodexAudit|CodexRole'
--- FAIL: TestCodexAuditMCPTool (29.15s)
    codex_audit_mcp_test.go:195: start blocked for 2.25053425s — it must return before the audit ends
FAIL	github.com/modu-ai/moai-adk/internal/cli	482.982s
```

The run had exactly one failure: a 900 ms wall-clock bound in `TestCodexAuditMCPTool`.
Control run at the base code (base files `b5815ca80` checked out into the tree, then restored):

```
$ go test -count=2 ./internal/cli/ -run '^TestCodexAuditMCPTool$'      # base b5815ca80 code
--- FAIL: TestCodexAuditMCPTool (21.84s)
    codex_audit_mcp_test.go:195: start blocked for 1.154173958s — it must return before the audit ends
FAIL
$ uptime
14:26  load averages: 204.96 226.80 208.12   (during the GREEN run)
14:32  load averages: 88.93 137.96 171.55    (after the control)
```

The same bound fails on the base code too, so the failure was already there; machine load
explains it, not this change. With the fix, the test failed 2 of 2 alone (948 ms, 1.32 s). At
base it failed 1 of 2.

Static checks on the fix tree:

```
$ go vet ./internal/cli/                       → (no output, exit 0)
$ GOOS=windows go vet ./internal/cli/          → (no output, exit 0)
$ GOOS=windows go build ./...                  → (no output, exit 0)
$ gofmt -l internal/cli/codex_audit_launch.go internal/cli/codex_audit_launch_test.go internal/cli/codex_audit_live_test.go
                                               → (no output)
```

## Baseline-attribution

All runs were measured in this session, in worktree
`.claude/worktrees/agent-abd5b18d4cb27db5d` on branch `WT-codex-role-embedded`. RED ran against
`b5815ca80` plus the test only. GREEN and the static checks ran against the tree of `14c7af8c8`.
The control ran with the three changed files reverted to `b5815ca80`. Tests were compiled from
source with `go test`; no installed `moai` binary was involved.

## Gaps

- No whole-repo suite was run, because the machine was heavily loaded. CI on the develop push
  gives the repository-wide verdict, and that verdict is PENDING.
- `TestCodexAuditMCPTool`'s async-start timing assertion was not observed passing under the
  current load, either on this tree or on the base. Its pass or fail under clean conditions is
  left to CI.
- The gated LIVE test `codex_audit_live_test.go` (MOAI_CODEX_ROLE_LIVE=1) was not run. Its
  nonce injection now goes through `overrideAuditRole` instead of the project copy. This change
  compiles, but it has not been exercised against a real codex.
- No end-to-end reproduction ran in a real Claude-profile project (`llm.yaml harness: claude`)
  with a stale `.codex/`. That scenario is covered only by the unit test fixture.

## Residual-risk

- **Project-local customisation of audit role instructions is no longer honoured.** A user who
  edited `.codex/agents/moai/plan-auditor.toml` (or any read-only role) to change the audit
  instructions will find that the launcher ignores the edit and uses the instructions embedded
  in the installed binary.
- The role content now follows the binary. An old installed `moai` launches old embedded
  instructions even after the project's templates move forward.
- `codexAuditRoleFS` is a package-level test seam. Tests that override it must not run with
  `t.Parallel()`. No test in `codex_audit_launch_test.go` uses `t.Parallel` today.
