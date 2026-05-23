---
id: SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001-acceptance
title: "Acceptance criteria — Worktree --team contextual launch"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/cli/worktree, internal/tmux"
lifecycle: spec-anchored
tags: "acceptance, gwt, traceability, tier-m"
tier: M
---

# Acceptance — SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001

## 1. Acceptance Criteria (Given/When/Then)

Each AC is a binary verifiable criterion with a deterministic verification command (exit code 0 = PASS, non-zero = FAIL).

### AC-WTL-001 — P1 dispatch: tmux + CG → moai glm window

**Given** the user is inside a tmux session (`$TMUX` non-empty) AND `.claude/settings.local.json` has `teammateMode: "tmux"` AND tmux session env contains `ANTHROPIC_AUTH_TOKEN`,
**When** the user runs `moai worktree new SPEC-WTL-DEMO-001 --team`,
**Then** a new tmux window is spawned with cwd=worktree-path and command `moai glm`. The new window appears in the same tmux session.

**Verification**:
```bash
# Test invokes tmux fake; asserts captured argv contains "new-window" + "moai glm"
go test -run TestTeamLaunch_P1_TmuxCG ./internal/cli/worktree/
```

### AC-WTL-002 — P2 dispatch: tmux + CC → moai cc window

**Given** the user is inside a tmux session AND CG mode is NOT active (no `teammateMode: "tmux"` OR no GLM env),
**When** the user runs `moai worktree new SPEC-WTL-DEMO-002 --team`,
**Then** a new tmux window is spawned with cwd=worktree-path and command `moai cc`.

**Verification**:
```bash
go test -run TestTeamLaunch_P2_TmuxCC ./internal/cli/worktree/
```

### AC-WTL-003 — P3 dispatch: no tmux → syscall.Exec moai cc/glm

**Given** the user is NOT in a tmux session (`$TMUX` empty),
**When** the user runs `moai worktree new SPEC-WTL-DEMO-003 --team`,
**Then** `syscall.Exec` is invoked with argv `[moai, cc]` (or `[moai, glm]` if CG state file indicates GLM) and cwd switched to the worktree path. The orchestrator session terminates as `moai cc`/`moai glm` replaces the process.

**Verification**:
```bash
go test -run TestTeamLaunch_P3_NoTmux_SyscallExec ./internal/cli/worktree/
```

Test fakes `syscallExecFn` to capture argv + cwd without real exec.

### AC-WTL-004 — P4 dispatch: --team absent → stdout handoff guidance

**Given** the user runs `moai worktree new SPEC-WTL-DEMO-004` WITHOUT `--team`,
**When** the command completes,
**Then** stdout contains the literal string `cd ` and ` && moai` (a paste-ready command), no tmux window is spawned, no syscall.Exec invoked, exit code 0.

**Verification**:
```bash
# stdout assertion includes both literals
go test -run TestTeamLaunch_P4_NoFlag_Handoff ./internal/cli/worktree/
```

### AC-WTL-005 — CG mode detection returns correct boolean AND drift case emits stderr warning

**Given** the following 4 detection scenarios:
1. In tmux + teammateMode=tmux + ANTHROPIC_AUTH_TOKEN set → expect `true`, no warning emitted
2. In tmux + teammateMode=tmux + no GLM env → expect `false` AND stderr contains substring `GLM env vars are absent` (REQ-WTL-009 drift case; P2 fallback)
3. NOT in tmux + teammateMode=tmux + GLM env → expect `false` (tmux requirement)
4. In tmux + no teammateMode + no GLM env → expect `false`

**When** `tmux.IsCGMode(settingsPath, stderrSink)` is invoked for each scenario (stderr buffer captured in tests),
**Then** the boolean return matches the expected value AND scenario 2 produces a captured stderr line containing `GLM env vars are absent` (verifying REQ-WTL-009 warning behavior).

**Verification**:
```bash
# Boolean return values
go test -run TestIsCGMode ./internal/tmux/
# Drift case stderr warning (REQ-WTL-009)
go test -run TestIsCGMode_DriftWarning ./internal/tmux/ -v 2>&1 | grep -q "GLM env vars are absent"
```

### AC-WTL-006 — BODP HARD: TestNew_NoAskUserQuestion remains green

**Given** the `--team` flag implementation adds `team_launch.go`, `handoff_guidance.go`, `swarm_registry.go`, and `team_launch_windows.go`,
**When** `TestNew_NoAskUserQuestion` is extended to scan all new files,
**Then** no file contains the substring `AskUserQuestion` outside test files or commented-out documentation.

**Verification**:
```bash
go test -run TestNew_NoAskUserQuestion ./internal/cli/worktree/
# Also raw grep:
grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/worktree/ | grep -v "_test.go" | grep -v "// "
# Expect: zero matches
```

### AC-WTL-007 — Pane spawn failure → P4 fallback with stderr notice

**Given** `--team` is set in P1/P2 path AND tmux `new-window` returns non-zero exit,
**When** the failure is handled,
**Then** the CLI falls back to printing handoff guidance on stdout, prints an error notice on stderr containing "tmux pane spawn failed", exit code is 0 (worktree was created successfully), AND no swarm registry file is written.

**Verification**:
```bash
go test -run TestTeamLaunch_PaneSpawnFailure_FallbackToP4 ./internal/cli/worktree/
```

### AC-WTL-008 — Swarm registry written with exact schema

**Given** `--team` succeeds via P1 (tmux + cg) for SPEC-WTL-DEMO-001,
**When** the swarm registry file is inspected,
**Then** `.moai/state/swarm/SPEC-WTL-DEMO-001.json` exists with exact fields per REQ-WTL-008 schema:

```bash
jq -e '.spec_id == "SPEC-WTL-DEMO-001"' .moai/state/swarm/SPEC-WTL-DEMO-001.json
jq -e '.worktree_path | length > 0' .moai/state/swarm/SPEC-WTL-DEMO-001.json
jq -e '.branch == "feature/SPEC-WTL-DEMO-001"' .moai/state/swarm/SPEC-WTL-DEMO-001.json
jq -e '.pane_id | length > 0' .moai/state/swarm/SPEC-WTL-DEMO-001.json
jq -e '.mode == "tmux-glm"' .moai/state/swarm/SPEC-WTL-DEMO-001.json
jq -e '.created_at | length > 0' .moai/state/swarm/SPEC-WTL-DEMO-001.json
jq -e '.created_by_pid > 0' .moai/state/swarm/SPEC-WTL-DEMO-001.json
# Permissions check
stat -c '%a' .moai/state/swarm/SPEC-WTL-DEMO-001.json | grep -q '^600$' || stat -f '%Lp' .moai/state/swarm/SPEC-WTL-DEMO-001.json | grep -q '^600$'
```

**Verification**:
```bash
go test -run TestSwarmRegistry_P1_Schema ./internal/cli/worktree/
```

### AC-WTL-009 — Cross-platform builds pass darwin/linux/windows

**Given** the `--team` flag is implemented with `team_launch.go` (POSIX) and `team_launch_windows.go` (Windows),
**When** cross-compilation is invoked for all three OS targets,
**Then** all three commands exit 0:

```bash
GOOS=darwin  GOARCH=amd64 go build ./...   # exit 0
GOOS=linux   GOARCH=amd64 go build ./...   # exit 0
GOOS=windows GOARCH=amd64 go build ./...   # exit 0
```

**Verification**: 3-command batch in M6 verification batch. Exit code captured per command.

### AC-WTL-010 — Invalid SPEC ID → no team launch, no registry

**Given** `moai worktree new <invalid-spec-id> --team` is invoked AND `WorktreeProvider.Add` returns an error,
**When** the command completes with non-zero exit,
**Then** no swarm registry file is written, no tmux window is spawned, no syscall.Exec invoked. The original error from `Add` is returned unchanged.

**Verification**:
```bash
go test -run TestTeamLaunch_WorktreeCreateFailure_NoLaunch ./internal/cli/worktree/
# Negative check: file does not exist
[ ! -f .moai/state/swarm/INVALID-SPEC.json ]
```

### AC-WTL-011 — Lint baseline: zero NEW violations

**Given** the SPEC implementation is complete,
**When** `golangci-lint run --timeout=2m` is compared against the baseline at `git merge-base HEAD origin/main`,
**Then** zero NEW errcheck / unused / staticcheck / ineffassign violations are introduced in the modified files.

**Verification**:
```bash
# Baseline measure
git stash
golangci-lint run --timeout=2m 2>&1 | grep -E "internal/(cli/worktree|tmux)/" | wc -l > /tmp/lint-baseline.txt
git stash pop
# Post-change measure
golangci-lint run --timeout=2m 2>&1 | grep -E "internal/(cli/worktree|tmux)/" | wc -l > /tmp/lint-post.txt
# Compare
[ "$(cat /tmp/lint-post.txt)" -le "$(cat /tmp/lint-baseline.txt)" ]
```

### AC-WTL-012 — Test coverage ≥85% for new files

**Given** the four NEW source files `team_launch.go`, `handoff_guidance.go`, `swarm_registry.go`, `cg_detect.go`,
**When** `go test -coverprofile=cover.out` is run with the standard package threshold,
**Then** coverage for each new file is ≥85%, with documented uncoverable lines (syscall.Exec replacement) explicitly marked.

**Verification**:
```bash
go test -coverprofile=cover.out ./internal/cli/worktree/ ./internal/tmux/
go tool cover -func=cover.out | grep -E "team_launch|handoff_guidance|swarm_registry|cg_detect" | \
    awk '$NF+0 < 85.0 { print "FAIL:", $0; exit 1 }'
```

### AC-WTL-013 — Skill body update + template mirror byte-identical

**Given** `.claude/skills/moai-workflow-worktree/SKILL.md` is updated with §`--team` Flag matrix,
**When** the template mirror is verified,
**Then** the local skill file and the template mirror are byte-identical:

```bash
diff -r .claude/skills/moai-workflow-worktree/ internal/template/templates/.claude/skills/moai-workflow-worktree/
# Expect: zero output, exit 0
diff .claude/rules/moai/workflow/worktree-integration.md internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md
# Expect: zero output, exit 0
```

Skill body MUST contain P1/P2/P3/P4 decision matrix (verifiable by grep):

```bash
grep -c "P1.*tmux.*cg\|P2.*tmux.*cc\|P3.*syscall\|P4.*handoff" .claude/skills/moai-workflow-worktree/SKILL.md
# Expect: ≥4 matches (one per pattern)
```

### AC-WTL-014 — BODP audit trail preserved after team launch

**Given** `--team` succeeds via any pattern (P1/P2/P3),
**When** `.moai/branches/decisions/<normalized-branch>.md` is inspected,
**Then** the audit trail file exists and was written BEFORE the team launch step (verifiable by file mtime ordering vs swarm registry mtime).

**Verification**:
```bash
go test -run TestTeamLaunch_BODPAuditTrailPreserved ./internal/cli/worktree/
# The test verifies bodp.WriteDecision was called before team launch dispatch
```

## 2. Edge Cases

| Case | Expected Behavior |
|------|-------------------|
| `--team` with `--tmux` | Cobra reports mutually exclusive flags error; exit non-zero before any worktree creation |
| `--team` with `--from-current` | Both flags compose normally (from-current sets base, --team adds launch); P1-P4 dispatch as usual |
| `--team` with `--path /custom/dir` | Worktree created at `/custom/dir`; team launch dispatches with cwd=`/custom/dir` |
| `--team` when claude binary not in PATH (P3) | syscall.Exec LookPath fails; print error on stderr; exit non-zero (different from pane spawn failure — this is a setup error, not a runtime convenience failure) |
| `--team` when moai binary not in PATH (P1/P2 internal `moai cc`/`moai glm`) | tmux window spawns but the shell inside fails. Captured in pane via `moai cc` exit, not visible to parent. P4 fallback NOT triggered (parent considers spawn succeeded once new-window returned 0) |
| `--team` invoked twice for same SPEC-ID | Worktree creation fails (path exists) → REQ-WTL-010 → no launch, no registry write, original error returned. If user pre-removed worktree and re-runs, registry is overwritten (resolved OQ-4) |
| settings.local.json corrupt JSON | `IsCGMode` returns `(false, error)`; cobra surfaces the error context; treat as CC mode fallback (P2/P3 cc), warn on stderr |
| Worktree path contains spaces | Tmux `new-window -c '<path with spaces>'` MUST quote the cwd. Verified in TestTeamLaunch_PathWithSpaces |

## 3. Traceability Matrix (REQ ↔ AC)

| REQ | AC(s) | Rationale |
|-----|-------|-----------|
| REQ-WTL-000 (Pre-existing State Survey) | (verified during M6 pre-flight) | Grep-based pre-flight, not a runtime AC |
| REQ-WTL-001 (P1) | AC-WTL-001 | Direct pattern dispatch test |
| REQ-WTL-002 (P2) | AC-WTL-002 | Direct pattern dispatch test |
| REQ-WTL-003 (P3) | AC-WTL-003 | syscall.Exec injectable test |
| REQ-WTL-004 (P4) | AC-WTL-004 | Stdout assertion |
| REQ-WTL-005 (CG state-driven LLM) | AC-WTL-001, AC-WTL-005 | P1 path + IsCGMode detection |
| REQ-WTL-006 (CG mode detection definition) | AC-WTL-005 | Verifies IsCGMode boolean returns iff: InTmuxSession AND teammateMode=tmux AND (GLM token OR base URL) per spec.md L80 |
| REQ-WTL-007 (Pane spawn failure → P4) | AC-WTL-007 | Fault injection test |
| REQ-WTL-008 (Swarm registry schema) | AC-WTL-008 | jq schema check |
| REQ-WTL-009 (teammateMode=tmux but no GLM env — drift case) | AC-WTL-005 scenario 2 | Boolean false + stderr warning emission |
| REQ-WTL-010 (Worktree creation failure → no launch) | AC-WTL-010 | Negative test, file absence check |
| REQ-WTL-011 (Template-First mirror) | AC-WTL-013 | byte-identical diff |
| REQ-WTL-012 (Windows P3 simulation) | AC-WTL-009 | Cross-platform build |
| REQ-WTL-013 (No AskUserQuestion HARD) | AC-WTL-006 | Static guard test — `TestNew_NoAskUserQuestion` extended over team_launch.go, handoff_guidance.go, swarm_registry.go |
| (Constraint) BODP audit trail preserved | AC-WTL-014 | mtime ordering + Section A spec.md L107 unchanged guarantee |
| (Constraint) settings.local.json NOT mutated | AC-WTL-001 + AC-WTL-002 sub-assertion | byte-identical pre/post check inside each P1/P2 test |
| (Constraint) coverage ≥85% | AC-WTL-012 | go tool cover threshold |
| (Constraint) lint baseline 0 NEW | AC-WTL-011 | baseline diff compare |

**Coverage check**: 13 unique REQ IDs (REQ-WTL-001 through REQ-WTL-013) all mapped to ACs. REQ-WTL-000 is a pre-flight gate, not a runtime AC. Every binary REQ has at least one AC; every AC verifies a specific REQ or HARD constraint. Traceability is complete with no orphan REQs.

## 4. Quality Gate Criteria

All criteria below MUST PASS before the SPEC is marked `implemented`:

- [ ] All 14 ACs PASS (binary, verified via single-command verification)
- [ ] `go test ./internal/cli/worktree/... ./internal/tmux/...` → exit 0
- [ ] `go vet ./...` → exit 0
- [ ] `golangci-lint run --timeout=2m` → 0 NEW violations vs baseline
- [ ] `GOOS=darwin GOARCH=amd64 go build ./...` → exit 0
- [ ] `GOOS=linux GOARCH=amd64 go build ./...` → exit 0
- [ ] `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- [ ] Coverage ≥85% per new file (cg_detect.go, team_launch.go, handoff_guidance.go, swarm_registry.go)
- [ ] Template mirror diff = empty (skill + rule)
- [ ] Pre-existing `TestNew_NoAskUserQuestion` extended and PASS
- [ ] Conventional Commit message + `🗿 MoAI` trailer present in feat-branch commit

## 5. Definition of Done

The SPEC is `implemented` when:

1. All 14 ACs PASS via the verification commands above
2. All 6 milestones M1-M6 are complete (per `progress.md` Evidence Tracker)
3. PR is created with title `feat(SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001): contextual --team flag for worktree new`
4. PR CI passes all required checks (per `.github/required-checks.yml`)
5. PR is merged into main (squash)
6. `/moai sync SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001` is invoked to update CHANGELOG + codemap + docs-site if applicable
7. SPEC `status:` field transitions from `draft` → `planned` (post-plan-merge) → `in-progress` (post-run-start) → `implemented` (post-run-merge) → `completed` (post-sync-merge) per the canonical 8-status enum
