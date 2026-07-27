# SPEC-PRETOOL-GATE-MOVE-001 — Acceptance Criteria

> Tier M acceptance artifact. AC enumeration + Given-When-Then scenarios + quality gates.
> Cross-references spec.md (REQ mapping) and plan.md (M4 test mapping).

## §A AC Matrix

| AC ID         | Description                                              | Maps to REQ                       | Severity   | M-plan |
|---------------|----------------------------------------------------------|-----------------------------------|------------|--------|
| AC-PGM-001    | Reachability — bad commit blocked                        | REQ-PGM-001, -002, -003           | MUST-PASS  | M4     |
| AC-PGM-002    | Reachability — good commit passes                        | REQ-PGM-001, -003                 | MUST-PASS  | M4     |
| AC-PGM-003    | Budget independence — >5s gate completes                 | REQ-PGM-002, -003                 | MUST-PASS  | M4     |
| AC-PGM-004    | Bypass defense — `--no-verify` denied                    | REQ-PGM-006                       | MUST-PASS  | M3/M4  |
| AC-PGM-005    | Language neutrality — ≥3 non-Go languages gated          | REQ-PGM-007                       | MUST-PASS  | M4     |
| AC-PGM-006    | Fast PreToolUse preserved — ast-grep                     | REQ-PGM-005                       | MUST-PASS  | M4     |
| AC-PGM-007    | Fast PreToolUse preserved — frozen-zone                  | REQ-PGM-005                       | MUST-PASS  | M4     |
| AC-PGM-008    | Template neutrality — no internal tokens                 | REQ-PGM-009                       | MUST-PASS  | M4     |
| AC-PGM-009    | Template neutrality — POSIX shell, no Go/macOS bias      | REQ-PGM-007, -009                 | MUST-PASS  | M4     |
| AC-PGM-010    | Byte-identity — Go const matches template                | REQ-PGM-008                       | MUST-PASS  | M4     |
| AC-PGM-011    | Install path — `moai init`                               | REQ-PGM-004, -008                 | MUST-PASS  | M4     |
| AC-PGM-012    | Install path — `moai update`                             | REQ-PGM-004, -008                 | MUST-PASS  | M4     |
| AC-PGM-013    | Bypass env var — `SKIP_MOAI_PRECOMMIT=1`                 | REQ-PGM-010                       | MUST-PASS  | M4     |
| AC-PGM-014    | Error surfacing — Claude sees rejection                  | REQ-PGM-011                       | MUST-PASS  | M1/M4  |
| AC-PGM-015    | No PreToolUse regression — fast path intact              | REQ-PGM-005                       | MUST-PASS  | M4     |

**Totals**: 15 ACs, all MUST-PASS. Each REQ maps to ≥1 AC; each AC maps to ≥1 REQ.

## §B Severity

All 15 ACs are **MUST-PASS**. The defect is CRITICAL (census Priority 1, cross-evidence raises severity under `defaultMode: bypassPermissions`). No PASS-WITH-DEBT accepted for the relocation core (AC-PGM-001 through AC-PGM-005 — the deny-semantics, budget-independence, bypass-defense, and language-neutrality guarantees). The remaining ACs (006-015) MAY be PASS-WITH-DEBT only with explicit user approval and a recorded debt entry.

## §C Traceability

Each AC in §A maps to ≥1 REQ in spec.md §C. Cross-cutting REQs map to multiple ACs:
- REQ-PGM-005 (fast PreToolUse preservation) → AC-PGM-006, AC-PGM-007, AC-PGM-015.
- REQ-PGM-007 (16-language neutrality) → AC-PGM-005, AC-PGM-009.
- REQ-PGM-008 (template distribution) → AC-PGM-010, AC-PGM-011, AC-PGM-012.
- REQ-PGM-009 (template content neutrality) → AC-PGM-008, AC-PGM-009.

Each AC has a verification command family (Bash test command or grep assertion) enumerated in §D's Given-When-Then blocks. The M1 pre-flight baseline (plan.md §C) and the M4 test results (plan.md §F.M4) populate the Evidence section of each AC at run-phase.

## §D Given-When-Then Scenarios

### AC-PGM-001 — Reachability, bad commit blocked

```gherkin
Given a fresh git repo with the relocated gate installed (moai init)
And a staged Go file containing a `go vet` failure (e.g. unreachable code, shadowed err)
When `git commit -m "test"` runs
Then git exits non-zero
And the `go vet` failure output appears in stderr
And `git rev-parse HEAD` does not show a new commit (the commit object was never created)
```

### AC-PGM-002 — Reachability, good commit passes

```gherkin
Given a fresh git repo with the relocated gate installed
And a staged clean Go file (no vet/lint/test failures)
When `git commit -m "test"` runs
Then git exits 0
And `git rev-parse HEAD` shows the new commit
```

### AC-PGM-003 — Budget independence, the core fix

```gherkin
Given a repo whose full vet/lint/test suite takes >5s wall-clock
# (use a test-only slow step if no real suite crosses 5s in CI; e.g. a synthetic `time.sleep(8s)` test)
When Claude Code runs `git commit` via the Bash tool
Then the gate completes (does NOT truncate at 5s)
And the deny verdict reaches git (not silently dropped)
And the commit is rejected if any step fails, OR succeeds if all steps pass
```

### AC-PGM-004 — Bypass defense, `--no-verify` denied

```gherkin
Given the PreToolUse handler with the coding-standards.md §Bash Risk-Amplifier destructive-primitive extension
When Claude Code runs `git commit --no-verify -m "test"` via the Bash tool
Then the command is denied or surfaced for explicit user confirmation
And the defense-in-depth decision is logged
And the user is informed that `--no-verify` bypasses the relocated gate
```

### AC-PGM-005 — Language neutrality, ≥3 non-Go languages gated

```gherkin
Given three fixtures, each in a separate git repo:
  | language | marker file      | failing check                         |
  | Python   | pyproject.toml  | `ruff check .` exits non-zero         |
  | Rust     | Cargo.toml      | `cargo clippy -- -D warnings` fails   |
  | Node.js  | package.json    | `eslint .` reports errors             |
When `git commit` runs on each fixture
Then each fixture's toolchain check executes via detectToolchain
And each commit is rejected by the relocated gate
And no fixture passes despite the defect
```

### AC-PGM-006 — Fast PreToolUse preserved, ast-grep

```gherkin
Given a staged file matching an ast-grep domain rule in .moai/config/astgrep-rules
And the `sg` binary is available on PATH
When the PreToolUse hook fires on Write/Edit/Bash
Then the ast-grep scanner runs and reports findings (advisory mode by default per gate.yaml)
And this behavior is unchanged from SPEC-FALSE-ALLCLEAR-GUARD-001 (PR #1183)
And the unavailable-scanner sentinel (PR #1183) still fires when `sg` is missing
```

### AC-PGM-007 — Fast PreToolUse preserved, frozen-zone

```gherkin
Given a Write to a frozen-zone path (e.g. .claude/rules/moai/core/moai-constitution.md when frozen)
When the PreToolUse hook fires
Then the frozen-zone guard denies the write
And this behavior is unchanged from pre-this-SPEC state
```

### AC-PGM-008 — Template neutrality, no internal tokens

```gherkin
Given the distributed template at internal/template/templates/.git_hooks/pre-commit after `make build`
When the content is grepped for internal tokens:
  - SPEC IDs: grep -c 'SPEC-[A-Z]*-[A-Z0-9]*-[0-9]' → 0
  - REQ tokens: grep -c 'REQ-[A-Z]*-[A-Z0-9]*-[0-9]' → 0
  - Audit citations: grep -c 'Audit.*Finding\|census' → 0
  - Commit SHAs: grep -c '[0-9a-f]\{7,\}' → 0 (or only well-known generic SHAs)
Then all grep counts are 0
And the content passes the template-neutrality CI guard (.github/workflows/template-neutrality-check.yaml)
```

### AC-PGM-009 — Template neutrality, POSIX shell

```gherkin
Given the distributed pre-commit template content
When inspected
Then it starts with `#!/bin/sh` (POSIX, not bash-specific)
And it uses only POSIX-compatible primitives (no `[[ ]]`, no bash arrays)
And it contains no Go-specific assumptions beyond `command -v go` / `command -v gofmt` guards
And it contains no macOS-specific paths (no /Users/goos/, no /usr/local/Homebrew)
And non-Go projects pass silently when their toolchain binaries are absent
```

### AC-PGM-010 — Byte-identity, Go const matches template

```gherkin
Given the extended preCommitHookContent Go const in internal/cli/hook_install.go
And the template file at internal/template/templates/.git_hooks/pre-commit
When the byte-identity test runs (mirror of TestPrePushTemplateMatchesConstant)
Then the const and the template file are byte-identical
And the test walks to the go.mod root, reads the template, compares with the const
And the test gracefully skips when the template is absent
```

### AC-PGM-011 — Install path, moai init

```gherkin
Given a fresh empty repo (git init) and `moai init` invoked
When init completes
Then .git/hooks/pre-commit exists
And it has mode 0755
And its first 3 lines contain the `# MoAI-ADK pre-commit hook` marker
And its body contains the heavy-gate invocation token (e.g. `moai gate` or equivalent)
And its body contains the SKIP_MOAI_PRECOMMIT bypass token
```

### AC-PGM-012 — Install path, moai update

```gherkin
Given an existing repo whose .git/hooks/pre-commit carries the MoAI marker
When `moai update` runs
Then the hook is overwritten with the latest content (marker-based safe overwrite)

Given an existing repo whose .git/hooks/pre-commit does NOT carry the MoAI marker
When `moai update` runs
Then the foreign hook is preserved unchanged
And ErrUserHookExists is surfaced as a non-fatal warning
```

### AC-PGM-013 — Bypass env var, SKIP_MOAI_PRECOMMIT

```gherkin
Given SKIP_MOAI_PRECOMMIT=1 in the environment
And a repo with the relocated gate installed
And a staged Go file containing a `go vet` failure
When `git commit` runs
Then the hook prints a bypass notice to stderr (e.g. "SKIP_MOAI_PRECOMMIT=1 — skipping MoAI pre-commit gate")
And exits 0 without running the heavy gate
And the commit proceeds despite the defect (consistent with SPEC-PRECOMMIT-001 REQ-PC-012)
```

### AC-PGM-014 — Error surfacing, Claude sees rejection

```gherkin
Given a git pre-commit rejection with a specific marker string on stderr (P1B_REJECT_MARKER_<random>)
When Claude Code runs `git commit` via the Bash tool
Then the Bash tool result includes the marker string in its stderr output
And Claude can read and relay the rejection reason to the user
And no additional plumbing is required (git's native stderr is the surfacing path)
```

### AC-PGM-015 — No PreToolUse regression

```gherkin
Given the relocation is complete (M2 + M3 shipped)
When the PreToolUse fast-check path runs (ast-grep + frozen-zone)
Then both checks still function (regression-tested against M1 baseline)
And IsGitCommit regex still detects `git commit` commands (gate.go:621)
And detectToolchain still resolves all 16 language toolchains (gate.go:88-225)
And the PreToolUse hook timeout in settings.json remains at 5s (unchanged by this SPEC)
```

## §E Edge Cases

- **Empty staged set**: the relocated gate MUST exit 0 when no files are staged (mirror SPEC-PRECOMMIT-001 REQ-PC-009 — fast no-op for non-Go or docs-only commits).
- **Merge commit (`git merge`** with no explicit `-m` from the user): git's own semantics govern whether pre-commit fires; verify at M1.
- **Worktree-resident repos**: the relocated gate runs with cwd inside the worktree; `$CLAUDE_PROJECT_DIR` resolution priority (CLAUDE.local.md B7) is preserved.
- **Non-moai downstream project**: the distributed hook runs in projects that may not have `moai` on PATH; `command -v moai` guard skips the heavy gate with a stderr notice (no failure — REQ-PGM-007/009).
- **Detached HEAD**: pre-commit runs under detached HEAD identically to a branch; verify git semantics.
- **Hooks disabled globally**: if `moai hook` is unavailable or the moai binary is absent, the hook body MUST degrade gracefully (exit 0 with a notice, not fail the commit).
- **Large repo (slow `git diff --cached`)**: the relocated gate's staged-file enumeration is bounded by git's own performance; no additional scan is added beyond what `detectToolchain` + the existing `stagedFiles` helper already do.
- **`git commit --amend`**: the relocated gate fires on `--amend` identically to a fresh commit (git's own semantics). SPEC-GATE-001 REQ-GATE-011 no-bypass intent is preserved (the prior defect was that the gate timed out before completing; once relocated, it completes).
- **Hooks chain (multiple pre-commit hooks)**: if a downstream user has additional pre-commit hooks (e.g. husky, pre-commit.com framework), the MoAI hook composes via the standard marker-based coexistence; verify no orphaning.
- **CI environment (headless)**: git pre-commit runs in CI's shell identically to local; the gate fires on `git commit` whether local or CI. The push-tier (pre-push) and CI-pipeline-tier remain unchanged.

## §F Quality Gate Criteria

The run-phase MUST verify the following at M4 completion (per plan.md §F.M4):

- **All 15 ACs MUST-PASS** with verifiable evidence (Claim + Evidence + Baseline-attribution + Gaps + Residual-risk per verification-claim-integrity.md §3).
- **Lint**: `golangci-lint run --timeout=2m` exits 0 with zero NEW findings (pre-existing baseline allowed; baseline captured at M1 §C pre-flight).
- **Tests**: `go test ./internal/cli/... ./internal/hook/...` exits 0 with all new + existing tests passing.
- **Cross-platform build**: `GOOS=windows GOARCH=amd64 go build ./...` exits 0.
- **Template-neutrality CI guard**: `.github/workflows/template-neutrality-check.yaml` passes (REQ-PGM-009 / AC-PGM-008).
- **Byte-identity**: extended `TestPrePushTemplateMatchesConstant` mirror passes (AC-PGM-010).
- **Coverage**: per-package coverage ≥85% on touched packages (`internal/cli/hook_install.go`, `internal/hook/quality/`).
- **spec-lint**: `moai spec lint .moai/specs/SPEC-PRETOOL-GATE-MOVE-001/` reports zero MUST-FIX findings on the SPEC artifacts.

## §G Definition of Done

- All 15 ACs PASS with verifiable evidence (each carries a command + observed output in the §E.2 run-phase evidence).
- No PRESERVE-listed surface touched (plan.md §A.5) — verified via `git diff --stat` against the base `3e6c92ef7` showing only EXTEND-listed files modified.
- Distributed template content is internal-content-clean (AC-PGM-008 grep returns 0).
- 3-phase close: plan → run → sync; the `completed` status transition rides the single sync commit per the Status Transition Ownership Matrix.
- progress.md §E.2 (run-phase evidence) populated by manager-develop; §E.4 (`sync_commit_sha`) populated by manager-docs (with D3 backfill if needed).
- CHANGELOG entry recorded by manager-docs (not duplicated — `grep -c 'SPEC-PRETOOL-GATE-MOVE-001' CHANGELOG.md` returns exactly 1).

## §H Forward-Looking Checks

- **Will this fix survive the next Claude Code runtime version?** Yes — git pre-commit is git's mechanism, not Claude Code's. Claude Code upgrades do not affect the relocation surface. The fix is runtime-version-independent.
- **Will this fix handle a determined `--no-verify` attacker?** Only with the M3 bypass defense (PreToolUse-side). The M3 extension is REQUIRED, not optional — without it, an agent (or user) can bypass the relocated gate via `--no-verify`, recreating the silent-drop defect under a new shape.
- **Will this fix handle the 17th language?** Yes — `detectToolchain()` table extension (gate.go:88-225) is the canonical path; the relocated gate calls the same detection. Adding a 17th language is a one-row addition to `toolchains`, not a hook-body change.
- **Will this fix handle CI environments?** Yes — git pre-commit runs in CI's shell identically to local; the gate fires on `git commit` whether local or CI. CI is unaffected by the relocation (the push-tier and CI-pipeline-tier remain unchanged).
- **Will this fix survive a `moai update` re-install?** Yes — the marker-based safe-overwrite (REQ-PGM-004 / SPEC-PRECOMMIT-001 REQ-PC-003) ensures `moai update` refreshes the hook body; a foreign hook without the marker is preserved.
