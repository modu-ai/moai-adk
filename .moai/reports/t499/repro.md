# t499 — reproduction record (lane-5, plan phase)

Card: t499 — "반쪽 배선 상태를 doctor 가 못 잡는다"
Tree: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t499
Branch: WT-codex-partial-wiring @ ace1c5440 (== origin/develop)
Binary under test: ./bin/moai (make build in this tree, BuildID list-280-gace1c5440)

## Claim

A plain `moai init` (no --agent flag) leaves a project carrying the 11 Codex
agent TOMLs and NO wiring files, and `moai doctor` does not report that state.

## Evidence

### 1. Plain init deploys the 11 TOMLs and no wiring

```
$ ./bin/moai init /tmp/t499-repro/plain --non-interactive --name repro
rc=0

$ find /tmp/t499-repro/plain/.codex -type f | sort
/tmp/t499-repro/plain/.codex/agents/moai/builder-harness.toml
/tmp/t499-repro/plain/.codex/agents/moai/e2e-tester.toml
/tmp/t499-repro/plain/.codex/agents/moai/manager-design.toml
/tmp/t499-repro/plain/.codex/agents/moai/manager-develop.toml
/tmp/t499-repro/plain/.codex/agents/moai/manager-docs.toml
/tmp/t499-repro/plain/.codex/agents/moai/manager-git.toml
/tmp/t499-repro/plain/.codex/agents/moai/manager-lead.toml
/tmp/t499-repro/plain/.codex/agents/moai/manager-spec.toml
/tmp/t499-repro/plain/.codex/agents/moai/plan-auditor.toml
/tmp/t499-repro/plain/.codex/agents/moai/super-advisor.toml
/tmp/t499-repro/plain/.codex/agents/moai/sync-auditor.toml
(count: 11)

$ ls .codex/hooks.json .codex/config.toml .moai/state/codex-wiring.json
ls: .codex/config.toml: No such file or directory
ls: .codex/hooks.json: No such file or directory
ls: .moai/state/codex-wiring.json: No such file or directory
```

### 2. doctor verdict, codex ON PATH (/Users/goos/.local/bin/codex)

```
Codex Wiring   codex installed, project not wired — run moai init --agent codex;
               ~/.codex/config.toml: 49 stale skill entries
```

Directive is incidentally right; the WORDING is false — the project does carry
Codex agent definitions.

### 3. doctor verdict, codex ABSENT from PATH (PATH=/usr/bin:/bin)

```
Codex Wiring   not wired (claude-only project) — skipped
```

Full output: doctor-nocodex-path.txt (this directory). This is the silent hole:
a project holding 11 Codex agent TOMLs is asserted to be "claude-only".

## Baseline-attribution

All three measurements were taken in this run, in this tree, against the binary
built from ace1c5440 in this worktree. No figure is carried over.

## Root cause (read, not inferred)

internal/cli/doctor_codex.go:92
    wired := hooksErr == nil || cfgErr == nil

Only .codex/hooks.json and .codex/config.toml are read. `.codex/agents/` is
never inspected, so the half-wired state has no branch in the discriminant.

Existence-gate contract confirmed unchanged and in scope to PRESERVE:
internal/codexwiring/wire.go:51-56 — RefreshWiring returns Result{}, nil when
wiringFilesExist(projectRoot) is false; it creates nothing.

## Correction (2026-09-07, lane, post-audit)

The line coordinate above originally read `doctor_codex.go:105`. Measured
`grep -n 'wired := hooksErr' internal/cli/doctor_codex.go` → `92:`. The quoted
snippet was always correct; only the coordinate was wrong. It is corrected here
to `:92`. Recorded rather than silently overwritten because SPEC-CODEX-PARTIAL-
WIRING-001 `spec.md:49` inherited the wrong number FROM this file — the
propagation is the point, not the digit.

## Gaps (explicitly NOT observed)

- The interactive wizard path (SPEC-INIT-HARNESS-PROMPT-001) was not exercised;
  only --non-interactive was run. Whether the wizard's agent_wiring answer can
  also land the half state was not measured here.
- `moai update` on a half-wired project was not run.
- No Windows/Linux measurement; darwin only.

## Residual risk

The 11-TOML count is a template fact that a future template edit can change;
any detection built on the literal count would be brittle. Detection should key
on the agents directory being non-empty, not on the number 11.
