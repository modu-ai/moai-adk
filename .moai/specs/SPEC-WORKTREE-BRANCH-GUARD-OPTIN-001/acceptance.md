# Acceptance — SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001

> One falsifiable AC per REQ. Each AC names a command and an expected observable outcome (deny reason or allow). The implementer MUST run the command and observe the literal output.

## §D AC Matrix

| AC ID | REQ | Severity | Summary |
|---|---|---|---|
| AC-REQ-1a | REQ-1 | MUST-PASS | Default-off → mutating command NOT denied |
| AC-REQ-1b | REQ-1 | MUST-PASS | Config flag explicitly false → mutating command NOT denied |
| AC-REQ-2a | REQ-2 | MUST-PASS | Enabled + `git stash list` → allow |
| AC-REQ-2b | REQ-2 | MUST-PASS | Enabled + `git stash show -p` → allow |
| AC-REQ-2c | REQ-2 | MUST-PASS | Enabled + `git merge-base` → allow |
| AC-REQ-2d | REQ-2 | MUST-PASS | Enabled + bare `git stash` → deny (dangerous form preserved) |
| AC-REQ-2e | REQ-2 | MUST-PASS | Enabled + `git merge feature` → deny (dangerous form preserved) |
| AC-REQ-3 | REQ-3 | MUST-PASS | Handler reads `BranchGuardConfig.Enabled` from ConfigProvider |
| AC-REQ-4 | REQ-4 | MUST-PASS | Template default `enabled: false`; no `enabled: true` in templates |
| AC-REQ-5 | REQ-5 | MUST-PASS | Rule mirror updated on both sides; parity tests green |
| AC-REQ-6a | REQ-6 | MUST-PASS | Enabled + `MOAI_BRANCH_GUARD_EXEMPT=1` → allow regardless |
| AC-REQ-6b | REQ-6 | MUST-PASS | Enabled + `AgentType == "manager-git"` → allow regardless |
| AC-REQ-6c | REQ-6 | MUST-PASS | Enabled + git-context uncertainty → allow + audit log (fail-open preserved) |
| AC-REQ-7 | REQ-4 | MUST-PASS | CHANGELOG entry under this SPEC's release section notes the default-off behavior change for existing users |

## §D.1 Severity / Traceability

- All ACs are MUST-PASS — the SPEC is Tier M, behavior-affecting, and ships through the template.
- Traceability: each AC maps 1:1 to a REQ in `spec.md` §C.

## §D.2 Given-When-Then Scenarios

### AC-REQ-1a — Default-off → mutating command NOT denied

```
GIVEN the default WorkflowConfig (BranchGuard.Enabled == false)
  AND a Bash HookInput with command "git reset --hard HEAD~1"
  AND a primary-checkout projectDir
  AND no exemption env / agent type
WHEN the preToolHandler.Handle is invoked
THEN the hook returns an allow decision
  AND no "BRANCH_GUARD_VIOLATION" string appears in the output
```

**Falsifiable command** (run in `internal/hook` test):
```bash
go test ./internal/hook/ -run 'TestBranchGuard_DefaultOff_DoesNotDeny' -v
```
**Observed exit / output**: `PASS`; test asserts the handler returns allow and `reason == ""`.

### AC-REQ-1b — Explicit false → mutating command NOT denied

```
GIVEN WorkflowConfig.BranchGuard.Enabled == false (explicit in workflow.yaml)
  AND a Bash HookInput with command "git switch feature/x"
  ...
WHEN Handle is invoked
THEN allow (no deny)
```

### AC-REQ-2a — Enabled + `git stash list` → allow

```
GIVEN WorkflowConfig.BranchGuard.Enabled == true
  AND a primary-checkout projectDir
  AND a Bash HookInput with command "git stash list"
WHEN Handle is invoked
THEN allow (no deny)
  AND no "BRANCH_GUARD_VIOLATION" in output
```
**Falsifiable command**: `go test ./internal/hook/ -run 'TestBranchState_StashList_Allowed' -v`

### AC-REQ-2b — Enabled + `git stash show -p` → allow

Same shape as AC-REQ-2a with command `git stash show -p stash@{0}`.

### AC-REQ-2c — Enabled + `git merge-base` → allow

Same shape with command `git merge-base HEAD origin/main`.

### AC-REQ-2d — Enabled + bare `git stash` → DENY (dangerous form preserved)

```
GIVEN Enabled == true, primary checkout, no exemption
  AND command "git stash"  (bare, mutating — pushes the working tree)
WHEN Handle is invoked
THEN the hook returns DecisionDeny
  AND the reason starts with "BRANCH_GUARD_VIOLATION: git stash"
```
Also covers `git stash push`, `git stash pop`, `git stash apply`, `git stash drop`.

### AC-REQ-2e — Enabled + `git merge feature` → DENY (dangerous form preserved)

```
GIVEN Enabled == true, primary checkout, no exemption
  AND command "git merge feature/x"
WHEN Handle is invoked
THEN DecisionDeny
  AND reason starts with "BRANCH_GUARD_VIOLATION: git merge"
```

### AC-REQ-3 — Handler reads config from ConfigProvider

```
GIVEN a ConfigProvider whose Get() returns Workflow.BranchGuard.Enabled == false
  AND command "git reset --hard"
WHEN Handle is invoked
THEN allow
GIVEN the SAME provider with Enabled == true (everything else identical)
  AND the same command
WHEN Handle is invoked
THEN DecisionDeny
```
**Falsifiable command**: `go test ./internal/hook/ -run 'TestPreTool_BranchGuard_ConfigGate' -v`
**Evidence**: test asserts the ONLY variable that flips the decision is the config flag — proving the handler reads the provider.

### AC-REQ-4 — Template default false; no `enabled: true` in templates

```
GIVEN the template tree internal/template/templates/
WHEN grepping for the branch-guard enabled flag
THEN no file under internal/template/templates/ sets branch_guard.enabled: true
  AND the Go default NewDefaultWorkflowConfig().BranchGuard.Enabled == false
```
**Falsifiable commands**:
```bash
rg -n 'branch_guard' internal/template/templates/ | grep -i 'enabled.*true'   # MUST return 0 matches
go test ./internal/config/ -run 'TestNewDefaultWorkflowConfig_BranchGuardDefaultFalse' -v
```

### AC-REQ-5 — Rule mirror updated on both sides; parity tests green

```
GIVEN both rule files have been updated:
  .claude/rules/moai/workflow/main-checkout-branch-guard.md
  internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard.md
WHEN running the template mirror + sanitized-pair tests
THEN go test ./internal/template/... exits 0
  AND both files contain the opt-in default documentation
  AND the template-side file contains NO SPEC-ID / REQ tokens (§25 sanitized)
```
**Falsifiable command**: `go test ./internal/template/ -run 'TestSanitizedPairParity|TestTemplateNoInternalContentLeak|TestRuleTemplateMirror' -v`

### AC-REQ-6a — Enabled + `MOAI_BRANCH_GUARD_EXEMPT=1` → allow regardless

```
GIVEN Enabled == true, primary checkout
  AND env MOAI_BRANCH_GUARD_EXEMPT=1
  AND command "git reset --hard"
WHEN Handle is invoked
THEN allow (exemption bypasses regardless of the config flag)
```

### AC-REQ-6b — Enabled + `manager-git` agent → allow regardless

```
GIVEN Enabled == true, primary checkout
  AND HookInput.AgentType == "manager-git"
  AND command "git switch feature/x"
WHEN Handle is invoked
THEN allow
```

### AC-REQ-6c — Enabled + git-context uncertainty → allow + audit log (fail-open preserved)

```
GIVEN Enabled == true
  AND projectDir is a NON-git directory (rev-parse exits non-zero)
  AND command "git reset --hard"
WHEN Handle is invoked
THEN allow (fail-open)
  AND stderr contains a fail-open advisory line
  AND .moai/logs/branch-guard-audit.log has a new structured entry
```
**Falsifiable command**: `go test ./internal/hook/ -run 'TestBranchGuard_FailOpen_Uncertainty' -v`

## §D.3 Indirect Verification

- **Pattern test isolation**: the M2 deny-origin test (AC-WBG-010 from the parent SPEC) MUST still pass — i.e., with `branchStatePatterns` set to nil, a deny must NOT come from `checkBashCommand`. This proves the deny path still originates in `checkBranchState`. Run: `go test ./internal/hook/ -run 'TestBranchGuard_CheckBranchStateOrigin' -v`.
- **Mirror parity**: `go test ./internal/template/...` MUST be green (the `main-checkout-branch-guard.md` pair is NOT in the byte-parity allowlist but IS in the §25 sanitized-pair registry — both invariants must hold).

## §D.4 Closure Gate (Definition of Done)

All of:
- [ ] Every MUST-PASS AC in §D.2 has been run and the observed output matches the expected.
- [ ] `go test ./...` green; `go test -race ./internal/hook/... ./internal/config/...` green.
- [ ] `make build` green (template embed recompiles).
- [ ] `golangci-lint run --timeout=2m` clean.
- [ ] `rg -n 'branch_guard.*enabled.*true' internal/template/templates/` returns 0 matches (template neutrality, §25).
- [ ] Both rule mirror files updated in the same commit.
- [ ] CLAUDE.local.md §22-family subsection documents the maintainer opt-in.

## §D.5 Forward-Looking Checks (Post-Merge)

- `moai hook pre-tool` smoke test on a maintainer machine with `enabled: true` — verify `git stash list` is allowed AND `git reset --hard` is denied in the primary checkout. (Out of CI scope; manual.)
- On the next `moai update` cycle for a non-maintainer project, verify the distributed default leaves the guard inert (no `BRANCH_GUARD_VIOLATION` on routine git commands).

## §D.6 Quality Gate Criteria (TRUST 5)

- **Tested**: new unit tests in `internal/hook` (M2/M3) + `internal/config` (M1/M4) — coverage of the `branch_guard.go` patterns and the new gate path ≥ 85%.
- **Readable**: comment at `branch_guard.go:58-60` updated to reflect the new regex behavior (the old comment was stale — see spec.md §D Evidence).
- **Unified**: `gofmt` + repo linter clean; regex patterns follow the existing `(?i)` prefix convention.
- **Secured**: the guard still prevents the dangerous git-state mutations when enabled; the default-off does NOT weaken the maintainer's protection when they opt in.
- **Trackable**: single conventional-commit `feat(hook): ...` referencing SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001, OR per-milestone commits with the SPEC ID in the footer.
