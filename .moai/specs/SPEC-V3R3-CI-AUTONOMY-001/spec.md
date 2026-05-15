---
id: SPEC-V3R3-CI-AUTONOMY-001
version: "0.2.0"
status: completed
created_at: 2026-05-05
updated_at: 2026-05-05
author: manager-spec
priority: P0
labels: [ci-cd, automation, worktree, branch-protection, quality-gate, ci-mirror, auxiliary-workflow, branch-origin, v3r3]
issue_number: null
breaking: false
bc_id: []
lifecycle: spec-anchored
depends_on: []
related_specs: [SPEC-V3R3-MX-INJECT-001, SPEC-CI-MULTI-LLM-001]
phase: "v3.0.0 R3 — CI/CD Autonomy"
module: "internal/bodp/, internal/cli/worktree/new.go, internal/cli/status.go, internal/template/templates/.git_hooks/, internal/template/templates/.github/workflows/, internal/template/templates/.claude/skills/moai-workflow-ci-watch/, .claude/skills/moai/workflows/plan.md, scripts/ci-mirror/, .claude/rules/moai/development/branch-origin-protocol.md"
tags: "ci-mirror, auto-fix-loop, worktree-state-guard, auxiliary-workflow-hygiene, branch-protection, branch-origin-decision-protocol, i18n-validator, v3r3"
related_theme: "Quality Pipeline Autonomy + GitHub Flow Hardening"
title: "Autonomous CI/CD Quality Pipeline + Worktree State Guard + Auxiliary Workflow Hygiene + Branch Origin Decision Protocol"
created: 2026-05-06
updated: 2026-05-13
---

# SPEC-V3R3-CI-AUTONOMY-001: Autonomous CI/CD Quality Pipeline + Worktree State Guard + Auxiliary Workflow Hygiene + Branch Origin Decision Protocol

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-05 | manager-spec | Initial draft. Root-cause analysis after 5-PR sweep on 2026-05-05 (PR #783/#744/#739/#747/#743) where each PR had distinct CI failures requiring manual debug+fix+push cycles. Pattern inventory P1-P10, root causes R1-R7, scope T1-T8 (user approved). 7-Wave implementation plan to avoid Anthropic SSE stream stalls (~1.5KB prompt per wave). |

---

## 1. Goal (목적)

CI/CD 품질 파이프라인을 사용자 개입 없이 자율적으로 실행·감시·복구하도록 만들고, GitHub Worktree/브랜치 상태가 세션 간에 유실되지 않도록 보호한다. 5-PR sweep (2026-05-05) 사례에서 드러난 7가지 근본 원인(R1-R7)을 8개 Tier(T1-T8)로 해결한다. 각 Tier는 독립적으로 가치를 제공하면서, 결합되었을 때 "사용자가 PR을 열기만 하면 머지 직전까지 자동으로 흐른다"는 목표 상태를 달성한다.

### 1.1 배경: 5-PR Sweep 사건 (2026-05-05)

오케스트레이터가 5개 PR(#783 chore translation batch B, #744 SPEC-CACHE-ORDER-001, #739 SPEC-CC2122-MCP-001, #747 SPEC-CC2122-HOOK-001, #743 SPEC-CC2122-STATUSLINE-001)을 순차 처리하는 과정에서 모든 PR이 서로 다른 CI 실패 패턴을 보였고, 각각 수동 debug → fix → push 사이클을 강요했다. 사용자 명시 지시:

> "전체 github 관련 CI/CD 및 지침들을 모두 분석해서 근본적인 문제 해결과 worktree 및 branche 사용시 적용이 안되어거나 유실 되는 경우가 허다하다 근본적인 문제를 해결하자."

### 1.2 Pattern Inventory (관찰된 실패 P1-P10)

| ID | 패턴 | 사례 |
|----|------|------|
| P1 | i18n 번역으로 인한 테스트 문자열 리터럴 손상 | #783: `"유효한 YAML 문서가 아닙니다"` → `"Not a valid YAML document"` (ST1005 violation + `mockReleaseData` map 데이터 삭제) |
| P2 | Local `go vet` PASS / CI `golangci-lint` FAIL | #739/#744 errcheck, unused, ST1005, QF1003 — 로컬에서 미검출 |
| P3 | 로컬 빌드는 Windows 미실행 | #743 Windows-only 실패 (`syscall.Flock` 직접 사용, agent_lint errcheck) |
| P4 | TestRuleSeed가 5개 언어 astgrep YAML 파일 부재로 실패 | #744 |
| P5 | Linux + race detector ETXTBSY in `TestLauncher_Launch_HappyPath` | #747 |
| P6 | main CI workflow가 retrigger empty-commit 후에도 실행 거부 | #744/#739 — auxiliary만 실행됨 |
| P7 | `Agent(isolation: "worktree")` returned `worktreePath: {}` | sub-agent 4/4 모두 main workspace 변경 |
| P8 | Auxiliary workflow 영구 실패 | claude-review (org quota), docs-i18n-check, llm-panel, Release Drafter |
| P9 | PR 생성 후 `gh pr checks` 수동 폴링 필요 | auto-watch + auto-fix loop 부재 |
| P10 | Sub-agent가 plan 파일을 untracked 상태로 남김 | cross-session file loss; 새 브랜치가 main이 아닌 feature branch에서 분기됨 |

### 1.3 Root Causes (R1-R7)

| ID | 근본 원인 | 해결 Tier |
|----|----------|----------|
| R1 | Pre-push 검증 hook 부재 (`.git/hooks` 비어있음, lefthook/husky/pre-commit 미사용). 로컬 quality gate < CI gate. | T1 |
| R2 | CI Mirror 로직(`.claude/skills/moai/workflows/sync.md` Step 3.1.5)이 `/moai sync`에서만 실행, 직접 push 시 미실행. | T2 (mirror를 reusable script로 분리) |
| R3 | `/moai sync`는 one-shot — push + PR만 생성, CI 상태 watch 미실행, auto-fix loop 부재. | T2 watch + T3 fix |
| R4 | Claude Code `Agent(isolation: "worktree")` 회귀/버그. Sub-agent가 main workspace 변경. | T6 |
| R5 | Auxiliary workflow noise가 CI signal을 침몰시킴. | T4 |
| R6 | CLAUDE.local.md §18.7가 branch protection 문서화하지만 적용된 적 없음; auto-merge 미활성화. | T5 |
| R7 | 새 브랜치가 main이 아닌 현재 브랜치 base로 default. Working tree contamination. | **T8 BODP** |

### 1.4 Tier 정의 (사용자 승인 범위 T1-T8)

| Tier | 해결 | 노력 | 핵심 산출물 |
|------|------|------|-------------|
| **T1** | R1: Pre-push hook + `make ci-local` Makefile target | 1d | `internal/template/templates/.git_hooks/pre-push` + Makefile target |
| **T2** | R2, R3 (read): CI watch loop | 3-5d | `moai-workflow-ci-watch` skill OR /moai sync 확장; `gh pr checks --watch` 사용 |
| **T3** | R3 (act): CI fail 시 auto-fix loop | 3-5d | CI watch → `expert-debug` (max 3 iterations) + AskUserQuestion 에스컬레이션 wiring |
| **T4** | R5: Auxiliary workflow cleanup | 1d | claude-code-review.yml/llm-panel.yml disable/fix; Release Drafter 80+ stale 정리; docs-i18n-check non-blocking |
| **T5** | R6: Branch protection + auto-merge | 0.5d | `gh api -X PUT .../branches/main/protection` 적용 (CLAUDE.local.md §18.7); GitHub auto-merge 활성화 |
| **T6** | R4: Worktree state guard + investigation | 2-3d | Pre/post Agent() state assertion + auto-restore; Anthropic upstream investigation을 claude-code-guide에 위임 |
| **T7** | P1: i18n validator | 1-2d | Static check: 테스트 assertion에 사용되는 error/string literal은 번역 금지; ci-local에 통합 |
| **T8** | R7, P10: Branch Origin Decision Protocol | 2d | 신규 CLI command `moai branch new <name>` + relatedness check + AskUserQuestion gate; CLAUDE.local.md §18 update |

---

## 2. Scope (범위)

### In Scope (이 SPEC이 다루는 것)

- Pre-push validation hook (16-language neutral) + `make ci-local` Makefile target (T1)
- `gh pr checks --watch` 기반 CI watch loop를 skill 또는 /moai sync 확장으로 구현 (T2)
- CI fail 시 `expert-debug` 자동 위임 + max 3 iterations + AskUserQuestion 에스컬레이션 (T3)
- Auxiliary workflow 정리: claude-code-review/llm-panel disable, Release Drafter stale 정리, docs-i18n-check non-blocking (T4)
- Branch protection rule 적용 (`main`, `release/*`) + GitHub auto-merge 활성화 (T5)
- `Agent(isolation: "worktree")` 호출 전후 working tree state assertion + auto-restore + claude-code-guide 위임 (T6)
- i18n validator: 테스트 참조 string/error literal 번역 금지 정적 검사 (T7)
- `moai branch new <name>` CLI command + 3-axis relatedness check + AskUserQuestion gate + audit trail (T8)
- CLAUDE.local.md §18 Enhanced GitHub Flow에 BODP 규칙 추가 (T8)

### Out of Scope (이 SPEC이 다루지 않는 것)

- Release tag/GoReleaser 자동 실행 (per `feedback_release_no_autoexec.md`: T5는 release 자동화 절대 금지)
- 사용자 코드 자동 수정 (auto-fix loop는 lint/format/test fix만 시도; semantic bug는 사용자 검토 필수)
- Claude Code `Agent(isolation:)` upstream 버그 수정 (claude-code-guide에 보고만, 우리 측은 guard layer로 우회)
- 16개 언어 전체 i18n validator AST 파서 (T7는 Go 코드만, 다른 언어는 follow-up SPEC)
- BODP의 monorepo path-based stacked PR 자동 분해 (단순 dependency detection만)
- `.moai/specs/SPEC-V3R3-MX-INJECT-001/` plan 파일 처리 (별도 SPEC scope)

### Exclusions (What NOT to Build)

- **❌ Auto-merge to main without CI green**: T5 GitHub auto-merge는 반드시 모든 required check 통과 후에만 트리거되어야 함 (branch protection이 강제)
- **❌ Auto release/tag creation**: feedback_release_no_autoexec.md per user directive — `/moai sync` 또는 본 SPEC의 어떤 Tier도 `git tag`, `gh release create`, `goreleaser release` 자동 실행 금지
- **❌ Hard-block on auxiliary workflow failure**: T4 후 docs-i18n-check, llm-panel, claude-code-review 실패는 PR merge를 차단하지 않음 (informational only)
- **❌ Sub-agent direct AskUserQuestion**: T3 auto-fix loop의 사용자 에스컬레이션은 오케스트레이터 경유 (subagent는 blocker report 반환만)
- **❌ Force-push to main**: T5 branch protection으로 영구 차단; 본 SPEC도 어떤 우회 메커니즘도 제공 금지
- **❌ Implicit branch deletion**: T8 BODP는 새 브랜치 생성만; 기존 브랜치 삭제 결정은 사용자가 직접
- **❌ Cross-platform native pre-push hook**: T1은 bash/sh 기반 단일 스크립트; PowerShell 변형 별도 (Windows 개발자는 git-bash 가정)
- **❌ Auto-fix for semantic test failures**: T3는 lint/format/missing-import/typo만 수정 시도; assertion failure, race, deadlock 등 semantic은 사용자 에스컬레이션

---

## 3. Requirements (EARS 요구사항)

### 3.1 T1 — Pre-Push Hook + ci-local

**REQ-CIAUT-001 (Ubiquitous)**: The MoAI-ADK template **shall** provide a pre-push git hook at `internal/template/templates/.git_hooks/pre-push` that runs `make ci-local` before allowing `git push` to succeed.

**REQ-CIAUT-002 (Event-Driven)**: **When** the user runs `moai init` or `moai update`, the system **shall** install the pre-push hook into `.git/hooks/pre-push` (with executable permission) unless the user opts out via `--no-hooks` flag.

**REQ-CIAUT-003 (Ubiquitous)**: The Makefile **shall** define a `ci-local` target that runs the same lint/test toolchain that GitHub Actions runs (`golangci-lint run`, `go vet ./...`, `go test -race ./...`, `go build ./...`), with cross-compilation for linux/darwin/windows × amd64/arm64 (Windows test execution skipped, build-only).

**REQ-CIAUT-004 (State-Driven)**: **While** `make ci-local` is running, the hook **shall** stream progress to stderr so the user sees which check is active.

**REQ-CIAUT-005 (Unwanted Behavior)**: **If** `make ci-local` exits non-zero, **then** the pre-push hook **shall** abort the push with a non-zero exit code and print a remediation hint (e.g., "Run `make fmt && make lint && make test` to fix").

**REQ-CIAUT-006 (Optional)**: **Where** the user explicitly bypasses (`git push --no-verify`), the system **shall** allow the push but log a warning to `.moai/logs/prepush-bypass.log` for audit.

**REQ-CIAUT-007 (Ubiquitous)**: The pre-push hook **shall** be 16-language neutral: detect project language from `internal/config/language_markers.go` and run the appropriate toolchain (Go: `go test`, Python: `pytest`, Node: `npm test`, etc.), skipping when no marker found.

### 3.2 T2 — CI Watch Loop

**REQ-CIAUT-008 (Event-Driven)**: **When** `/moai sync` finishes pushing a branch and creating/updating a PR, the orchestrator **shall** invoke `moai-workflow-ci-watch` skill (or extended sync sub-phase) to monitor `gh pr checks --watch` until all required checks complete or fail.

**REQ-CIAUT-009 (Ubiquitous)**: The CI watch implementation **shall** distinguish required checks (Lint, Test ubuntu/macos/windows, Build matrix, CodeQL) from auxiliary checks (claude-code-review, llm-panel, docs-i18n-check) per `.github/workflows/required-checks.yml` configuration.

**REQ-CIAUT-010 (State-Driven)**: **While** the CI watch loop is active, the orchestrator **shall** display the latest check state (queued/in_progress/completed) every 30 seconds via natural-language status update.

**REQ-CIAUT-011 (Event-Driven)**: **When** all required checks complete with success status, the orchestrator **shall** report PR ready-to-merge state and (if T5 branch protection allows) propose enabling GitHub auto-merge via AskUserQuestion.

**REQ-CIAUT-012 (Unwanted Behavior)**: **If** any required check fails (failure, cancelled, action_required, timed_out), **then** the orchestrator **shall** transition to T3 auto-fix loop with the failure metadata captured.

**REQ-CIAUT-013 (Ubiquitous)**: CI mirror logic currently embedded in `.claude/skills/moai/workflows/sync.md` Step 3.1.5 **shall** be extracted into a reusable shell script at `scripts/ci-mirror/run.sh` so it can be called from both `make ci-local` (T1) and the watch loop (T2).

### 3.3 T3 — Auto-Fix Loop on CI Fail

**REQ-CIAUT-014 (Event-Driven)**: **When** T2 watch detects a required check failure, the orchestrator **shall** download the failed check's log via `gh run view <run-id> --log-failed` and pass the log + PR diff context to `expert-debug` subagent for diagnosis.

**REQ-CIAUT-015 (Ubiquitous)**: The auto-fix loop **shall** attempt at most **3 iterations** of (debug → propose patch → user-confirm via AskUserQuestion → apply → push → re-watch) before mandatory escalation.

**REQ-CIAUT-016 (State-Driven)**: **While** the auto-fix loop iteration count is below 3, the orchestrator **shall** route mechanical failures (lint/format/missing-import/typo) through `expert-debug` for automated patch proposal.

**REQ-CIAUT-017 (Unwanted Behavior)**: **If** the failure is semantic (test assertion failure, race condition, deadlock, panic), **then** the orchestrator **shall** skip auto-patch and immediately escalate to user via AskUserQuestion with the diagnosis report.

**REQ-CIAUT-018 (Event-Driven)**: **When** the iteration count reaches 3 without green CI, the orchestrator **shall** halt the loop and present the user with options: (a) continue manually, (b) revise SPEC, (c) abandon PR, via AskUserQuestion.

**REQ-CIAUT-019 (Ubiquitous)**: All auto-fix iterations **shall** be logged to `.moai/reports/ci-autofix/<PR-NNN>-<YYYY-MM-DD>.md` with: iteration number, failure type, patch diff, result, escalation reason if any.

### 3.4 T4 — Auxiliary Workflow Hygiene

**REQ-CIAUT-020 (Ubiquitous)**: Auxiliary workflows that consistently fail due to external constraints (org quota, third-party API limits) **shall** be moved to `.github/workflows/optional/` and removed from required-checks list.

**REQ-CIAUT-021 (Event-Driven)**: **When** a Release Drafter draft is older than 30 days and not associated with an active release branch, the system **shall** automatically close it via scheduled cleanup workflow.

**REQ-CIAUT-022 (State-Driven)**: **While** auxiliary workflows run, their failure **shall not** block PR merge (advisory only).

**REQ-CIAUT-023 (Ubiquitous)**: `docs-i18n-check` workflow **shall** continue to run on PRs touching `docs-site/**` but with `continue-on-error: true` and explicit "advisory" badge in PR comment.

**REQ-CIAUT-024 (Optional)**: **Where** the user wants to permanently disable a flaky auxiliary workflow, the system **shall** provide `make ci-disable WORKFLOW=<name>` that comments out the trigger and commits with `chore(ci): disable <name>` message.

### 3.5 T5 — Branch Protection + Auto-Merge

**REQ-CIAUT-025 (Ubiquitous)**: The MoAI-ADK installer (`moai github init` or `moai update`) **shall** prompt the user to apply branch protection rules to `main` and `release/*` patterns, using the JSON payload defined in CLAUDE.local.md §18.7.

**REQ-CIAUT-026 (Event-Driven)**: **When** the user confirms protection application via AskUserQuestion, the system **shall** invoke `gh api -X PUT /repos/<owner>/<repo>/branches/main/protection` with the canonical JSON.

**REQ-CIAUT-027 (Ubiquitous)**: The required_status_checks contexts list **shall** be sourced from a single SSoT at `.github/required-checks.yml`, not hardcoded in multiple places.

**REQ-CIAUT-028 (Unwanted Behavior)**: **If** `gh` CLI is not authenticated or lacks admin permission, **then** the system **shall** display the exact `gh api` command for the user to run manually and exit gracefully without false success claim.

**REQ-CIAUT-029 (Optional)**: **Where** the user opts in via `moai pr enable-auto-merge <PR>`, the system **shall** invoke `gh pr merge <PR> --auto --squash` (or `--merge` for release branches per CLAUDE.local.md §18.3) so GitHub auto-merges once branch protection conditions are met.

**REQ-CIAUT-030 (Ubiquitous)**: The system **shall NOT** automatically create release tags, GitHub Releases, or trigger GoReleaser as part of T5 (per `feedback_release_no_autoexec.md`).

### 3.6 T6 — Worktree State Guard

**REQ-CIAUT-031 (Event-Driven)**: **When** the orchestrator is about to invoke `Agent(isolation: "worktree", ...)`, it **shall** capture the current working tree state snapshot: `git status --porcelain`, `git rev-parse HEAD`, `git rev-parse --abbrev-ref HEAD`, list of untracked files under `.moai/specs/`.

**REQ-CIAUT-032 (Event-Driven)**: **When** the `Agent()` call returns, the orchestrator **shall** verify the post-call state matches the captured pre-call state and that the agent's response includes a non-empty `worktreePath` field.

**REQ-CIAUT-033 (Unwanted Behavior)**: **If** post-call state diverges from pre-call state (HEAD changed, untracked files lost, branch changed), **then** the orchestrator **shall** halt subsequent agent invocations, log the divergence to `.moai/reports/worktree-guard/<YYYY-MM-DD>.md`, and surface the issue to the user via AskUserQuestion (options: restore from snapshot, accept divergence, abort).

**REQ-CIAUT-034 (Unwanted Behavior)**: **If** the agent response has empty `worktreePath: {}` despite `isolation: "worktree"` request, **then** the orchestrator **shall** treat the result as suspect, validate file targets, and warn the user before any subsequent push.

**REQ-CIAUT-035 (Ubiquitous)**: The orchestrator **shall** delegate Anthropic upstream investigation of the `Agent(isolation:)` regression to the `claude-code-guide` subagent, producing a structured bug report at `.moai/reports/upstream/agent-isolation-regression.md`.

**REQ-CIAUT-036 (Optional)**: **Where** the worktree state diverges and the user opts to restore, the system **shall** apply `git restore --source=<snapshot-sha> --staged --worktree :/` and reattach untracked file paths from the snapshot.

### 3.7 T7 — i18n Validator

**REQ-CIAUT-037 (Ubiquitous)**: A static analyzer at `scripts/i18n-validator/` **shall** scan Go source for string literals participating in test assertions (`require.Equal`, `assert.Contains`, etc.) and flag them as "translation-locked".

**REQ-CIAUT-038 (Event-Driven)**: **When** a translation PR modifies a translation-locked string, the validator **shall** exit non-zero and report the file:line and the test that depends on it.

**REQ-CIAUT-039 (Ubiquitous)**: The validator **shall** be integrated into `make ci-local` (T1) and run in CI as a required check.

**REQ-CIAUT-040 (State-Driven)**: **While** the validator processes a file, it **shall** report progress and not exceed 30s wall-clock for the full repo (acceptable for pre-push integration).

**REQ-CIAUT-041 (Optional)**: **Where** a string literal is intentionally translatable (user-facing CLI message), it **shall** be marked with a `// i18n:translatable` magic comment that exempts it from the lock check.

### 3.8 T8 — Branch Origin Decision Protocol (BODP)

**Design principle (resolves user critique 2026-05-05)**: BODP is implemented as **behavior inside existing branch-creation entry points**, NOT as new slash commands or CLI subcommands. User-facing surface remains unchanged; BODP is invisible to users who already use `/moai plan --branch`, `/moai plan --worktree`, and `moai worktree new`. Goal: minimize cognitive load + avoid command proliferation.

**Existing entry points (reused without breaking changes)**:

| Entry Point | 위치 | 기존 동작 | BODP 추가 |
|------------|------|----------|----------|
| `/moai plan --branch [name]` | `.claude/skills/moai/workflows/plan.md` Phase 3 Branch Path | `feature/SPEC-{ID}-{desc}` 생성 (현재 HEAD에서) | 생성 전 BODP 검사 + AskUserQuestion으로 base 선택 |
| `/moai plan --worktree` | `.claude/skills/moai/workflows/plan.md` Phase 3 Worktree Path | `moai worktree new` 호출 | worktree 호출 전 BODP 검사 + base 결정 |
| `moai worktree new <SPEC-ID>` | `internal/cli/worktree/new.go` | 현재 HEAD에서 worktree 생성 | default = origin/main; `--from-current` flag로 opt-out |
| raw `git checkout -b ...` | git 자체 | MoAI 무관 | `moai status`가 audit trail 부재 감지 → reminder |

**REQ-CIAUT-042 (Event-Driven)**: **When** `/moai plan --branch` is invoked, the orchestrator **shall** run the BODP relatedness check (REQ-044) before delegating branch creation to `manager-git`. The orchestrator **shall** present the user with a base-branch choice via AskUserQuestion when relatedness signals are present, and pass the chosen base to `manager-git` as a parameter.

**REQ-CIAUT-043 (Event-Driven)**: **When** `/moai plan --worktree` is invoked, the orchestrator **shall** run the same BODP check before invoking `moai worktree new`. The chosen base branch **shall** be propagated via a new flag `--base <branch>` on `moai worktree new`.

**REQ-CIAUT-043b (Event-Driven)**: **When** `moai worktree new <SPEC-ID>` CLI is invoked directly (bypassing `/moai plan`), it **shall** default to `origin/main` as the base. A new flag `--from-current` **shall** allow opt-out (preserve old default of current HEAD). The CLI **shall NOT** invoke AskUserQuestion (orchestrator-only HARD per `.claude/rules/moai/core/askuser-protocol.md`); decisions in this path are flag-based only.

**REQ-CIAUT-044 (Event-Driven)**: BODP relatedness check **shall** evaluate three signals on every invocation:
- (a) Does the new SPEC depend on the current branch's commits? (signal: SPEC `depends_on:` field referencing an in-progress SPEC on this branch; file-path overlap with current branch's diff vs main)
- (b) Are working tree untracked files related to the new SPEC? (signal: SPEC-ID directory match in `git status --porcelain` untracked entries)
- (c) Is there an open PR with current branch as head? (signal: `gh pr list --head <current> --state open` returns ≥1 PR)

**REQ-CIAUT-045 (Event-Driven)**: **When** all 3 signals are negative, the orchestrator (slash command path) **shall** recommend "main에서 분기" (branch from `origin/main`) as the first AskUserQuestion option marked `(권장)`. The CLI path uses `origin/main` automatically without prompting.

**REQ-CIAUT-046 (Event-Driven)**: **When** signal (a) is positive, the orchestrator **shall** recommend "현재 브랜치에서 분기 (stacked PR)" as the first option with rationale citing the detected dependency.

**REQ-CIAUT-047 (Event-Driven)**: **When** signal (b) is positive, the orchestrator **shall** recommend "현재 브랜치에 계속 작업" as the first option (no new branch creation; user proceeds on current).

**REQ-CIAUT-047b (Event-Driven)**: **When** signal (c) is positive, the orchestrator **shall** recommend "stacked PR (base=현재 브랜치)" as the first option and warn about parent-merge gotcha (CLAUDE.local.md §18.11 Case Study).

**REQ-CIAUT-048 (Ubiquitous)**: After confirmation (slash command AskUserQuestion or CLI flag default), the existing branch-creation handler **shall** execute the appropriate git commands with the chosen base:
- main 분기: `git fetch origin main && git checkout -B <new> origin/main`
- stacked: `git checkout -B <new>` (from current HEAD)
- continue: no-op, just record decision

**REQ-CIAUT-049 (Ubiquitous)**: Every BODP decision **shall** be recorded at `.moai/branches/decisions/<branch-name>.md` with: timestamp, invocation entry point (`plan-branch`, `plan-worktree`, `worktree-cli`), current branch, relatedness signals (a/b/c with evidence), user choice, executed command — for audit trail.

**REQ-CIAUT-050 (Unwanted Behavior)**: **If** the user invokes raw `git checkout -b` (bypassing all MoAI entry points), **then** the system **shall NOT** intervene (BODP is opt-in via the existing entry points), but the next `moai status` invocation **shall** detect the off-protocol branch (absence of audit trail file) and emit a friendly reminder pointing to `/moai plan --branch` or `moai worktree new` as the BODP-aware paths.

**REQ-CIAUT-051 (Ubiquitous)**: CLAUDE.local.md §18 (Enhanced GitHub Flow) **shall** be amended with a new subsection §18.12 (Branch Origin Decision Protocol) that documents the algorithm, the three existing entry points where BODP is enforced, and the raw-git-checkout reminder. **No new slash command or CLI subcommand is introduced.**

---

## 4. Constraints (제약조건)

### 4.1 Hard Constraints

- [HARD] **Template-First**: 모든 `.claude/`, `.moai/`, `.github/` 변경은 `internal/template/templates/`에 먼저 추가 후 `make build` 실행 (CLAUDE.local.md §2)
- [HARD] **16-language neutrality**: T1 pre-push hook은 16개 언어(go/python/typescript/javascript/rust/java/kotlin/csharp/ruby/php/elixir/cpp/scala/r/flutter/swift) 모두에서 동작 (CLAUDE.local.md §15)
- [HARD] **No hardcoding**: URL, model, env keys, threshold는 const 추출 (CLAUDE.local.md §14)
- [HARD] **CI mirror cross-compile only on Windows**: 로컬에서는 Windows 테스트 실행 불가, build-only 검증
- [HARD] **AskUserQuestion is sole user channel**: T3 auto-fix loop의 모든 사용자 결정은 오케스트레이터 경유 (subagent는 blocker report만)
- [HARD] **No release/tag automation** (`feedback_release_no_autoexec.md`): T5 branch protection은 release 자동 생성 금지
- [HARD] **Conventional Commits + 🗿 MoAI co-author trailer**: 본 SPEC이 만드는 commit/PR도 동일 규칙
- [HARD] **BODP defaults to "main에서 분기"**: relatedness check가 positive signal 없이 모두 negative이면 main이 권장 (사용자 명시 직접 지시: "main에서 분기 + T8 추가 (권장)")

### 4.2 Soft Constraints

- 토큰 예산: 본 SPEC 전체 구현 시 plan phase 30K + run phase 180K + sync phase 40K = 250K (Wave 분할 시 wave당 ~30K)
- Wave 당 prompt 크기 ≤1.5KB (`feedback_large_spec_wave_split.md` 기준)
- 각 Tier 독립 가치: T1만 적용해도 R1 해결, T5만 적용해도 R6 해결 — 부분 채택 가능

---

## 5. Acceptance Criteria 요약

상세는 `acceptance.md` 참조. 25개 AC 매핑 (canonical IDs):

- **T1 (Pre-Push)**: AC-CIAUT-001 (lint block), AC-CIAUT-002 (16-language detect), AC-CIAUT-003 (`--no-verify` log)
- **T2 (CI Watch)**: AC-CIAUT-004 (auto-invoke), AC-CIAUT-005 (required vs auxiliary)
- **T3 (Auto-Fix)**: AC-CIAUT-006 (mechanical resolve), AC-CIAUT-007 (semantic escalate), AC-CIAUT-008 (iteration cap 3)
- **T4 (Aux Cleanup)**: AC-CIAUT-009 (non-blocking), AC-CIAUT-010 (Release Drafter cleanup)
- **T5 (Branch Protection)**: AC-CIAUT-011 (force-push block), AC-CIAUT-012 (no required → no merge), AC-CIAUT-013 (no release automation), AC-CIAUT-021 (required-checks SSoT), AC-CIAUT-022 (gh auth fail graceful)
- **T6 (Worktree Guard)**: AC-CIAUT-014 (divergence detect), AC-CIAUT-015 (empty worktreePath suspect)
- **T7 (i18n Validator)**: AC-CIAUT-016 (mockReleaseData block), AC-CIAUT-017 (magic comment exempt), AC-CIAUT-023 (30s budget)
- **T8 (BODP)**: AC-CIAUT-018 (current session replay), AC-CIAUT-019 (stacked PR), AC-CIAUT-024 (off-protocol reminder), AC-CIAUT-025 (CLAUDE.local.md §18.12)
- **Cross-Tier**: AC-CIAUT-020 (5-PR sweep replay — manual validation)

---

## 6. Cross-References

- **Related SPECs**:
  - `SPEC-V3R3-MX-INJECT-001`: 다른 in-flight SPEC, working tree에 untracked plan 파일 존재 (별도 scope, 본 SPEC 영향 받지 않음)
  - `SPEC-CI-MULTI-LLM-001`: 이전 CI 작업 (이미 머지됨), 본 SPEC이 그 다음 단계
- **Memory / Feedback**:
  - `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/project_ci_failures_fix_session.md` (5-PR sweep 사건 상세)
  - `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_large_spec_wave_split.md` (Wave 분할 근거)
  - `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_release_no_autoexec.md` (release 자동화 금지)
  - `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/audit_sweep_patterns.md` (검증 패턴)
- **Documentation Updates**:
  - CLAUDE.local.md §18 → §18.12 BODP 추가, §18.7 적용 명령 강조
  - `.claude/rules/moai/development/branch-origin-protocol.md` 신규 규칙
  - `.claude/skills/moai/workflows/sync.md` Step 3.1.5 mirror 로직을 `scripts/ci-mirror/run.sh` 참조로 변경

---

## 7. Resolved Decisions (was Open Questions)

본 섹션은 사용자 AskUserQuestion 라운드 (2026-05-05)로 plan-auditor finding F-002에 따라 모두 해결됨. 각 결정은 plan.md Wave task로 즉시 반영되었으며, /moai run 진입 전 모든 OQ가 closed 상태임.

1. **OQ1 (T2 placement)** → **RESOLVED: 신규 skill `moai-workflow-ci-watch` 분리**
   - 결정 근거: 관심사 분리 (sync는 one-shot 유지, watch는 opt-in long-running), 독립적 테스트/비활성화 가능
   - 반영: plan.md Wave 2 W2-T01이 "OQ1 결정" task에서 "skill SKILL.md 작성" task로 교체됨

2. **OQ2 (T3 user-confirm cadence)** → **RESOLVED: 첫 iteration은 항상 confirm + iteration 2-3은 trivial fix만 silent**
   - 결정 근거: 안전성(첫 patch는 검토)과 개발자 흐름(trivial 후속은 자동) 균형
   - "Trivial" 정의: whitespace, gofmt/goimports, import order만 — 그 외는 confirm 강제
   - 반영: plan.md Wave 3 W3-T01이 "OQ2 결정" task에서 "iteration cadence wiring" task로 교체됨

3. **OQ3 (T5 protection scope)** → **RESOLVED: `main` + `release/*` 적용 (feat/SPEC-* 미포함)**
   - 결정 근거: 핵심 브랜치 보호 + 실험/탐색 자유도 보존 (CLAUDE.local.md §18.7 제시안)
   - 반영: REQ-CIAUT-025의 protection scope를 두 패턴으로 명시

4. **OQ4 (T8 BODP scope creep)** → **RESOLVED: opt-in CLI 유지 (raw `git checkout -b` 가로채기 안 함)**
   - 결정 근거: intrusion 방지 + 비-MoAI 워크플로우 호환 + audit-trail 이점은 `moai status` reminder (REQ-CIAUT-050)로 보존
   - 반영: REQ-CIAUT-050 wording 그대로 유지

---

## 8. Risks

- **R-CIAUT-1**: Pre-push hook이 느려 개발자 unproductive (mitigation: progress streaming + `--no-verify` escape hatch)
- **R-CIAUT-2**: CI watch loop가 long-running으로 토큰 소비 (mitigation: 30s polling + max 30분 timeout + 사용자 abort 옵션)
- **R-CIAUT-3**: Auto-fix loop가 잘못된 patch로 main 오염 (mitigation: 매 patch는 force-push가 아닌 새 commit + 사용자 confirm + max 3 iteration)
- **R-CIAUT-4**: T5 branch protection 적용 후 admin override 어려움 (mitigation: `enforce_admins: false`로 시작, 점진적 강화)
- **R-CIAUT-5**: T6 worktree state diff에 false positive (mitigation: `.gitignore` 상에 명시된 경로는 비교 제외)
- **R-CIAUT-6**: T7 i18n validator AST가 false positive로 정당 번역 차단 (mitigation: `// i18n:translatable` magic comment escape)
- **R-CIAUT-7**: T8 BODP relatedness check가 잘못된 권장 (mitigation: 항상 사용자 AskUserQuestion 최종 확정, "Other" 옵션 자동 포함)

---

Version: 0.1.0
Status: draft
Last Updated: 2026-05-05
