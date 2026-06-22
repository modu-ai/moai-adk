# Acceptance Criteria — SPEC-V3R6-CG-MODE-HARDENING-001

Every criterion is mechanically verifiable (a `go test`, a `grep`, a `GOOS=windows go build`, or a file/coverage assertion). Each AC group maps 1:1 to a requirement in spec.md §B.

## §D. AC Matrix

| AC ID | Requirement | Verification mechanism | Pass condition |
|-------|-------------|------------------------|----------------|
| AC-CGH-001 | REQ-CGH-001 | `GOOS=windows GOARCH=amd64 go build ./...` + grep | Windows build succeeds; `launcher.go` (or `launch_windows.go`) guards `syscall.Exec` behind POSIX-only path |
| AC-CGH-002 | REQ-CGH-002 | `go test` (ordering) | Test proves leader GLM-cred strip executes before tmux injection; injection failure leaves no stale creds |
| AC-CGH-003 | REQ-CGH-003 | `go test` (single-RMW) | After `applyCGMode`, `teammateMode == "tmux"` in one write; no intermediate state with `teammateMode` absent |
| AC-CGH-004 | REQ-CGH-004 | `grep` + (conditional) `make build` | `CLAUDE.local.md §22.3` shows `"tmux"`/`""`; template synced + rebuilt if root `CLAUDE.md §15` edited |
| AC-CGH-005 | REQ-CGH-005 | `go test` (atomicity + locking) | Settings writes go through one locked + atomic helper; concurrent-write test produces no truncated/clobbered file; user keys preserved |
| AC-CGH-006 | REQ-CGH-006 | `go test` (detector SSOT) | `IsCGMode` true from `llm.yaml team_mode=cg` / tmux session-env even with CLEAN process env; sibling `REQ-WTL-009` drift warning reconciled |
| AC-CGH-007 | REQ-CGH-007 | `go test` (validation) | Malformed / non-https / off-allowlist `base_url` rejected with clear error; `DefaultGLMBaseURL` passes |
| AC-CGH-008 | REQ-CGH-008 | `go test` (precondition) | tmux-present-but-unavailable yields "tmux not installed" message, not "restart your tmux session" |
| AC-CGH-009 | REQ-CGH-009 | `go test -cover` + parity test | Production credential-routing path exercised; leader-strip + teammate-inject asserted; inject↔clear key parity (minus `ANTHROPIC_AUTH_TOKEN`) asserted |
| AC-CGH-010 | REQ-CGH-010 | full suite + `golangci-lint` | `go test ./internal/cli/... ./internal/tmux/... ./internal/config/...` pass; zero new lint findings |

## §D.1 Given-When-Then scenarios

### AC-CGH-001 — Windows launch safety (REQ-CGH-001)

**Scenario 1a — Windows cross-compile**
- GIVEN the `launchClaudeDefault` shared launch path
- WHEN `GOOS=windows GOARCH=amd64 go build ./...` is run
- THEN the build succeeds (the unguarded `syscall.Exec` no longer breaks the Windows compile path)
- VERIFY: `GOOS=windows GOARCH=amd64 go build ./... && echo OK`

**Scenario 1b — POSIX path preserved**
- GIVEN a POSIX host
- WHEN the launcher executes the launch path
- THEN it still uses `syscall.Exec` to replace the current process
- VERIFY: `grep -n "syscall.Exec" internal/cli/launch*.go` shows the call gated behind a POSIX build tag OR a `runtime.GOOS != "windows"` branch

### AC-CGH-002 — Cleanup-before-injection ordering (REQ-CGH-002)

**Scenario 2a — injection failure leaves no stale creds**
- GIVEN a `settings.local.json` carrying stale GLM creds from a prior `moai glm` run
- AND a tmux session-env injection that is forced to fail (injected fake)
- WHEN `applyCGMode` runs
- THEN the leader GLM credentials have ALREADY been stripped before the injection failure returns
- VERIFY: `go test -run TestApplyCGMode_StripsCredsBeforeInjection ./internal/cli/`

### AC-CGH-003 — Single atomic teammateMode write (REQ-CGH-003)

**Scenario 3a — no teammateMode-absent window**
- GIVEN a fresh CG launch
- WHEN `applyCGMode` mutates `settings.local.json`
- THEN `teammateMode == "tmux"` is established in a single read-modify-write (no separate clear-then-set sequence)
- VERIFY: `go test -run TestApplyCGMode_SingleTeammateModeWrite ./internal/cli/` (the named test asserts the single-RMW invariant: `teammateMode == "tmux"` after one write, with no separate clear-then-set sequence on the CG path)

### AC-CGH-004 — Doc correctness + template sync (REQ-CGH-004)

**Scenario 4a — CLAUDE.local.md §22.3 matches code**
- GIVEN the code only ever writes `teammateMode` as `"tmux"` or `""`
- WHEN `CLAUDE.local.md §22.3` is read
- THEN it documents `"tmux"`/`""` (NOT `"glm"`/`"claude"`) and disambiguates from `llm.yaml team_mode` (`cg`/`glm`/`""`)
- VERIFY: `grep -n 'teammateMode' CLAUDE.local.md` shows the corrected values

**Scenario 4b — template sync when root CLAUDE.md §15 edited (conditional)**
- GIVEN the user-facing root `CLAUDE.md §15` CG description was edited
- WHEN the change is finalized
- THEN `internal/template/templates/CLAUDE.md §15` carries the identical change AND `make build` was run
- VERIFY: `diff <(grep -A30 'CG Mode' CLAUDE.md) <(grep -A30 'CG Mode' internal/template/templates/CLAUDE.md)` shows the §15 description in sync; `git status internal/template/embedded.go` shows regeneration. (Vacuously satisfied if §15 was not edited.)

### AC-CGH-005 — Locked + atomic settings write (REQ-CGH-005)

**Scenario 5a — atomic write, no truncation**
- GIVEN two concurrent settings mutations
- WHEN both run through the new helper
- THEN the resulting file is always valid JSON (no truncation / partial write) and user-only keys (`defaultMode`, `env.PATH`) survive
- VERIFY: `go test -race -run TestSettingsLocal_ConcurrentAtomicWrite ./internal/cli/`

### AC-CGH-006 — Detector SSOT (REQ-CGH-006) [HEADLINE]

**Scenario 6a — CG detected from clean leader env**
- GIVEN `llm.yaml team_mode == "cg"` (and/or tmux session-env carries GLM markers)
- AND a CLEAN leader PROCESS env (no `ANTHROPIC_AUTH_TOKEN`, no `z.ai` base URL)
- WHEN `IsCGMode` is evaluated
- THEN it returns true (teammates will route to GLM)
- VERIFY: `go test -run TestIsCGMode_AuthoritativeOnTeamMode ./internal/tmux/`

**Scenario 6b — sibling drift warning reconciled, not deleted**
- GIVEN the `REQ-WTL-009` drift condition
- WHEN `IsCGMode` evaluates it against the new source of truth
- THEN a meaningful drift warning is still emitted (semantics preserved)
- VERIFY: `go test ./internal/tmux/...` (sibling SPEC's `IsCGMode` tests still pass) AND `grep -n 'REQ-WTL-009' internal/tmux/cg_detect.go` still present

### AC-CGH-007 — GLM base_url validation (REQ-CGH-007, SECURITY)

**Scenario 7a — malicious/typo URL rejected**
- GIVEN an `llm.yaml` with `base_url` set to a non-https or off-allowlist host
- WHEN config validation runs
- THEN validation fails with a clear error and the token is NOT injected toward that endpoint
- VERIFY: `go test -run TestValidateLLM_RejectsUnsafeGLMBaseURL ./internal/config/`

**Scenario 7b — canonical default passes**
- GIVEN `base_url == DefaultGLMBaseURL` (`https://api.z.ai/api/anthropic`)
- WHEN validation runs
- THEN it passes
- VERIFY: same test, positive case asserts `DefaultGLMBaseURL` is accepted

### AC-CGH-008 — tmux availability precondition (REQ-CGH-008)

**Scenario 8a — clear error when binary absent**
- GIVEN `InTmuxSession()` true but `Detector.IsAvailable()` false (injected fake detector)
- WHEN `applyCGMode` runs
- THEN it fails with a "tmux not installed / not executable" message including install guidance, NOT "restart your tmux session"
- VERIFY: `go test -run TestApplyCGMode_TmuxUnavailableMessage ./internal/cli/`

### AC-CGH-009 — Credential-routing coverage (REQ-CGH-009)

**Scenario 9a — production path leader-strip + teammate-inject**
- GIVEN the production `applyCGMode` path (NOT the `isTestEnvironment()` early-return)
- WHEN exercised with a recording-fake session manager
- THEN the leader settings env is stripped of GLM creds AND the teammate injection records the GLM credential set
- VERIFY: `go test -run TestApplyCGMode_CredentialRoutingInvariant ./internal/cli/`

**Scenario 9b — inject↔clear key parity**
- GIVEN the key set written by `injectTmuxSessionEnv`
- WHEN compared against `clearTmuxSessionEnv`'s removal list
- THEN every injected key except the intentionally-retained `ANTHROPIC_AUTH_TOKEN` appears in the removal list
- VERIFY: `go test -run TestTmuxEnv_InjectClearParity ./internal/cli/`

### AC-CGH-010 — Regression safety (REQ-CGH-010)

**Scenario 10a — full suites + lint green**
- GIVEN all changes applied
- WHEN the verification batch runs
- THEN `go test ./internal/cli/... ./internal/tmux/... ./internal/config/...` pass and `golangci-lint run` reports zero new findings
- VERIFY: the read-only verification batch (below)

## §D.2 Edge Cases

- EC-1: `settings.local.json` absent on CG launch → helper creates it with `teammateMode="tmux"`, no panic.
- EC-2: `settings.local.json` empty (zero-byte) → treated as no-settings, not a parse error (matches current `len(data) == 0` guards).
- EC-3: `llm.yaml` missing `team_mode` (default `""`) → `IsCGMode` returns false (not CG); no false positive.
- EC-4: tmux session-env reader returns error (e.g., not in tmux) → `IsCGMode` falls back gracefully to false, no crash.
- EC-5: `base_url` empty in `llm.yaml` → `loadGLMConfig` falls back to `DefaultGLMBaseURL` (validation passes); empty is not a rejection trigger by itself on the load path.
- EC-6: Windows + `--team` already routes to a stub (sibling SPEC) → REQ-CGH-001 only fixes the shared `launchClaudeDefault` path; CG's tmux dependency keeps it POSIX-only at runtime.

## §D.3 Definition of Done

- [ ] All 10 AC groups pass their named verification.
- [ ] `GOOS=windows GOARCH=amd64 go build ./...` succeeds.
- [ ] `go test ./internal/cli/... ./internal/tmux/... ./internal/config/...` all pass.
- [ ] `go test -cover ./internal/cli/` shows `injectTmuxSessionEnv` / `clearTmuxSessionEnv` credential-routing logic exercised on the production path (no longer solely behind `isTestEnvironment()` early-return).
- [ ] `golangci-lint run` zero new findings.
- [ ] Sibling SPEC `SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001` tests remain green (no `REQ-WTL-009` regression).
- [ ] `CLAUDE.local.md §22.3` corrected; template `CLAUDE.md` synced + `make build` run IF root `CLAUDE.md §15` was edited.
- [ ] No `AskUserQuestion` / `mcp__askuser` in CLI code (C-1 static guard).
- [ ] No inline `os.Getenv("ANTHROPIC_*")` raw strings (C-2).

## §D.4 Read-only Verification Batch (run-phase completion)

```bash
# 1. Cross-platform build
GOOS=windows GOARCH=amd64 go build ./...

# 2. Full affected test suites
go test ./internal/cli/... ./internal/tmux/... ./internal/config/...

# 3. Credential-routing coverage
go test -coverprofile=/tmp/cgh.out ./internal/cli/ && go tool cover -func=/tmp/cgh.out | grep -iE "injectTmuxSessionEnv|clearTmuxSessionEnv"

# 4. Subagent-boundary grep (C-1)
grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/ | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//"

# 5. envkeys constant discipline (C-2) — NEW production refs only (binds new code, not baseline)
# Baseline TODAY (before any change): 10 matches = 8 in *_test.go + 2 in cg_detect.go:80,83
# (hasGLMEnv process-env reads). The 2 cg_detect.go matches are REMOVED by REQ-CGH-006
# (the detector no longer reads process env), so the post-change PRODUCTION count must be 0.
grep -rn 'os.Getenv("ANTHROPIC_' internal/cli/ internal/tmux/ | grep -v envkeys.go | grep -v '_test.go'
# Expected post-change: 0 production matches (the 2 cg_detect.go process-env reads removed by REQ-CGH-006).

# 6. Sibling-SPEC drift warning preserved (C-7, REQ-CGH-006)
grep -n 'REQ-WTL-009' internal/tmux/cg_detect.go

# 7. Lint baseline
golangci-lint run --timeout=2m
```
