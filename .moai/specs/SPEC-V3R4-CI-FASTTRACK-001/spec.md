---
id: SPEC-V3R4-CI-FASTTRACK-001
title: "CI/CD Fast Track for Single-Developer Workflow (Path-Filter + Review Bot Consolidation)"
version: "0.1.0"
status: draft
created: 2026-05-17
updated: 2026-05-17
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".github/workflows"
lifecycle: spec-anchored
tags: "ci, cd, github-actions, paths-filter, review-bot, single-developer, productivity"
---

# CI/CD Fast Track for Single-Developer Workflow (Path-Filter + Review Bot Consolidation)

## HISTORY

| Version | Date       | Author       | Change                                                                                            |
|---------|------------|--------------|---------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-17 | manager-spec | Initial draft. Captures (A) workflow-level optimizations + (B) branch-protection 6→4 already applied. |

## 1. Overview

### 1.1 Goal

본 SPEC은 moai-adk-go 1인 개발자(macOS) 환경의 CI/CD 회전 시간(per-PR wait)을 5-6분+ 에서
80%+ 단축하는 것을 목표로 한다. 두 축의 변경을 결합한다:

- **축 A — Workflow-level optimization** (이 SPEC의 RUN-PHASE 산출물):
  - `dorny/paths-filter@v3` 기반 docs-only PR fast-path
  - `codeql.yml` paths-ignore 적용
  - 5개 review bot workflow 제거 (codex / gemini / glm / llm-panel / claude-code-review.optional)
  - `lefthook.yml` + Makefile `preflight` 타깃으로 로컬 pre-push 게이트
  - `nightly-full-matrix.yml` 일일 03:00 UTC + tag-push full 3-OS matrix
- **축 B — Branch protection rule** (오케스트레이터가 PRE-PLAN 시점에 이미 적용 완료, 본 SPEC의 baseline):
  - Required status checks: `Lint` / `Test (ubuntu-latest)` / `Build (linux/amd64)` / `CodeQL` (4 items)
  - 제거됨: `Test (macos-latest)` / `Test (windows-latest)` (2 items)
  - 적용 시점: main HEAD `680232c82` 기준 (현재 main HEAD `41b6f37dc`에서도 활성)

축 B는 본 SPEC의 RUN-PHASE 산출물이 아니라 PRE-PLAN orchestrator immediate action으로 이미
활성화되어 있다. 본 SPEC은 (a) doctrine layer (`CLAUDE.local.md` §18.7)를 축 B 결과로
동기화하고, (b) 축 A 6개 task로 PR-level wait time을 추가 단축한다.

### 1.2 Why now

triggering problem (사용자 보고, 2026-05-17):

- 매 PR마다 5-6분+ CI 대기. windows-latest cold-start (~3-4분) + race-instrumented test
  (`go test -race`) 가 본질적 hot-loop.
- 19개 workflow 가 모든 PR에 매번 트리거. docs-only PR (예: README, CHANGELOG, `.moai/specs/`
  markdown만 수정) 도 3-OS Go test 매트릭스를 wait.
- Review bot 4건 (`codex` / `gemini` / `glm` / `private-guard`) 이 매 PR 마다
  `command not found` FAIL — 이중 `private-guard` 는 codex-review.yml + llm-panel.yml의
  job 이름. 실제 CLI 가 어느 머신에도 설치돼 있지 않아 영구 RED.
- 1인 개발 ROI 대비 over-engineered. macOS / Windows 매트릭스의 의미 있는 회귀 탐지율
  대비 PR-level wait time 비용이 비대칭.

WHY 본 SPEC 이 지금 필요한가:

- 1인 개발자 한 명의 흐름(flow state)이 CI 대기로 끊기는 비용은 시간 / 일 단위로 누적된다.
- 축 B (branch-protection reduce)는 이미 활성 — doctrine layer (`CLAUDE.local.md` §18.7,
  현재 6 required check 로 명시)와의 drift 가 발생한 상태이며 즉시 동기화 필요.
- 축 A 의 paths-filter / review consolidation 변경 사항이 클수록 (B) 의 benefit이 실현된다.
  (B) 만으로는 docs-only PR 의 Go test 매트릭스 자체가 사라지지 않기 때문이다 (branch
  protection은 required check 만 줄일 뿐, workflow trigger 자체는 그대로 동작).

### 1.3 Impact

- **영향받는 파일**: 7개 워크플로우 (1 modify ci.yml / 1 modify codeql.yml / 1 preserve claude-code-review.yml /
  5 delete review workflows / 1 new nightly-full-matrix.yml) + 2개 root-level
  (`lefthook.yml` NEW / `Makefile` MODIFY) + 1개 doctrine (`CLAUDE.local.md` §18.7) +
  auto-memory `lessons.md` (1 entry append).
- **추정 LOC delta**:
  - ci.yml: +60 / -10 (detect job + skip-marker job 추가, 매트릭스 conditional 적용)
  - codeql.yml: +6 / -0 (paths-ignore stanza 추가)
  - claude-code-review.yml: 변경 없음 (PRESERVE — 사용자 신뢰 + Anthropic API 인증 기존 적용)
  - claude.yml: 변경 없음 (PRESERVE — `@claude` mention trigger, issue/comment, codex 비의존)
  - review-quality-gate.yml: 변경 없음 (PRESERVE — Claude Code Review check_run 의존, codex 비의존)
  - 5 deleted workflows: 총 -약 300 LOC (codex-review / gemini-review / glm-review / llm-panel /
    claude-code-review.optional)
  - nightly-full-matrix.yml: +120 LOC (NEW)
  - lefthook.yml: +30 LOC (NEW)
  - Makefile: +20 LOC (preflight 타깃 + dependency stub)
  - CLAUDE.local.md §18.7: ~15 LOC (required check 4개로 갱신 + 3-tier philosophy 명문화)
  - lessons.md: +40 LOC (entry #18)
- **Workflow count**: 18 → 14 (delete 5 + add 1 nightly = net -4)
- **Branch protection required checks**: 6 → 4 (이미 적용된 축 B baseline)
- **Backward compatibility**: 기존 GoReleaser / Dependabot / Release Drafter / docs-i18n /
  community / label-sync / auto-merge workflow 모두 변경 없음. `make build` /
  `go test ./...` API 그대로 유지.

## 2. Goals

1. **Docs-only PR fast-path**: `dorny/paths-filter@v3` 기반 conditional matrix 으로 markdown /
   spec / docs-site / .moai/ 전용 PR 에서 Go test 매트릭스를 SKIP 한다. branch protection
   compatibility 를 위해 skip-marker job 으로 동일한 check name (`Test (ubuntu-latest)`)을
   pass 신호로 제공한다.
2. **CodeQL paths-ignore**: `codeql.yml` 에서 markdown / spec / rule / docs-site 경로 변경
   PR 의 CodeQL run 을 자동 skip 한다.
3. **Review bot consolidation**: 5개 review workflow (codex / gemini / glm / llm-panel /
   claude-code-review.optional) 를 제거하고 단 1개 `claude-code-review.yml` 만 유지한다.
   `claude.yml` (issue/comment `@claude` trigger) 와 `review-quality-gate.yml` (Claude Code
   Review check_run severity parser) 는 codex 비의존이므로 PRESERVE.
4. **Local pre-flight gate**: `lefthook.yml` + Makefile `preflight` 타깃으로 개발자 머신에서
   pre-push 시점에 `lint --fast` + `test -race -short` + `build` 를 실행해 PR 제출 전
   회귀를 차단한다.
5. **Nightly full matrix**: 1인 개발 trade-off 의 안전망으로 일일 03:00 UTC + workflow_dispatch +
   release tag push 시점에 3-OS full 매트릭스를 실행. 실패 시 자동 GitHub issue 생성.
6. **Doctrine sync**: `CLAUDE.local.md` §18.7 의 required status checks 목록을 6 항목에서
   4 항목으로 정정하고 (B) 결정의 rationale + 3-tier CI philosophy (Tier 1 fast / Tier 2
   opt-in / Tier 3 nightly) 를 명문화한다.
7. **Lessons capture**: `lessons.md` #18 항목에 1인 개발 CI 3-tier pattern 의 선택 근거와
   미래 SPEC 에서 동일 패턴을 적용할 때의 의사결정 휴리스틱을 기록한다.

## 3. Non-Goals

### 3.1 Out of Scope

다음 항목은 본 SPEC scope 외이며 향후 별도 SPEC 으로 분리되어야 한다. Listed explicitly
to prevent scope creep:

- **Self-hosted runner setup**: 1인 개발 비용 구조상 prohibitive (월별 인프라 + 유지보수
  부담 > PR wait 비용 절감 효과).
- **Test sharding**: premature optimization. Go test 매트릭스의 ubuntu-latest 평균 wait
  time 이 paths-filter 이후 1-2분 수준이라면 sharding ROI 가 음수.
- **Main matrix 에서 모든 OS 제거**: 본 SPEC 은 ubuntu-latest 를 required check 로 유지하며,
  macOS/Windows 는 nightly 매트릭스로 이전. 모든 OS 제거 (예: linux-only build/test) 는
  cross-platform binary distribution (GoReleaser 5 플랫폼) 안전망을 잃으므로 거부.
- **`claude-code-review.yml` 제거**: 사용자가 명시적으로 유지를 승인한 path. 본 SPEC 은
  이 workflow 의 trigger 조건 / 인증 / payload 를 변경하지 않는다.
- **임의 review bot 재활성화 (e.g., codex / gemini 부활)**: 본 SPEC 머지 후 별도 사용자 의사결정
  없이 자동 재활성화 금지. 사용자가 마음을 바꾸면 별도 SPEC 으로 root cause (CLI 설치 +
  credential 분배 + ROI 재평가) 를 먼저 다뤄야 한다.
- **macOS-latest 의 required check 재추가**: 축 B 에서 사용자가 승인한 제거. 본 SPEC 은
  nightly matrix 로 회귀 탐지를 보장하므로 required check 재추가는 의도적으로 거부.
- **GitHub Actions → Jenkins / CircleCI / 외부 CI provider 이전**: orthogonal scope.
  별도 사용자 의사결정 필요.
- **v3.0.0-rc1 release tag 발행 actions**: 본 SPEC 의 nightly matrix 가 tag push 트리거를
  포함하지만, release tag 발행 자체는 별도 SPEC (release readiness) 의 scope.
- **`claude.yml` / `review-quality-gate.yml` 자동 제거**: 본 SPEC 의 RUN-PHASE 에서
  graceful-guard 또는 PRESERVE 결정 중 선택. 사전 audit (T4) 에서 codex 비의존성 확인 후
  PRESERVE 가 기본값. PRE-PLAN 사전 정찰 결과 (Overview §1.3 참조) 도 PRESERVE 를 시사함.
- **CI cache 전략 변경 (Go module cache TTL 등)**: orthogonal. Go test wait time 단축의
  지배 변수는 paths-filter 가 아니라 cold-start 와 race detector 자체이므로 cache 튜닝은
  marginal.

## 4. EARS Requirements

### REQ-CIFT-001 — Docs-Only PR Fast Path (Ubiquitous)

The system SHALL route docs-only PRs to skip the Go test matrix. A docs-only PR is
defined as any PR whose changed file set falls ENTIRELY under one or more of:

- `**.md` (any markdown)
- `.moai/specs/**`
- `.moai/docs/**`
- `.moai/reports/**`
- `.moai/research/**`
- `.claude/rules/**`
- `.claude/skills/**` (markdown bodies of skills)
- `docs-site/**`
- `CHANGELOG.md` / `LICENSE` / `NOTICE`

For docs-only PRs, the `Test (ubuntu-latest)` / `Test (macos-latest)` /
`Test (windows-latest)` jobs SHALL emit an immediate-success skip-marker job whose
check-run name matches the matrix job name EXACTLY. This satisfies branch protection
required-check expectations without executing the actual test suite.

For non-docs-only PRs, the full matrix SHALL execute as before.

Implementation: `dorny/paths-filter@v3` detect job + conditional matrix gate + skip-marker
job pattern (per design.md AD-001/AD-002).

### REQ-CIFT-002 — CodeQL Paths-Ignore (Ubiquitous)

The `.github/workflows/codeql.yml` workflow SHALL apply `paths-ignore` filter to skip
CodeQL analysis on PRs whose changes match ONLY documentation-class paths:

```yaml
on:
  pull_request:
    paths-ignore:
      - '**.md'
      - '.moai/specs/**'
      - '.moai/docs/**'
      - '.moai/reports/**'
      - '.moai/research/**'
      - '.claude/rules/**'
      - 'docs-site/**'
```

Other CodeQL triggers (push to main, schedule) SHALL remain unchanged.

### REQ-CIFT-003 — Review Workflow Consolidation (Event-driven)

WHEN a pull request is opened, reopened, or synchronized, THEN EXACTLY ONE review-class
workflow SHALL trigger: `claude-code-review.yml`. The following 5 review workflow files
SHALL be removed from `.github/workflows/`:

- `codex-review.yml`
- `gemini-review.yml`
- `glm-review.yml`
- `llm-panel.yml`
- `claude-code-review.optional.yml`

The `claude-code-review.yml` (Claude API based, user-approved), `claude.yml`
(`@claude` mention trigger on issues/comments), and `review-quality-gate.yml` (parses
Claude Code Review check-run severity) workflows SHALL be preserved unchanged.

### REQ-CIFT-004 — private-guard Dependency Audit (Event-driven)

WHEN run-phase task T4 audits the `private-guard` job identifier and the
`review-quality-gate.yml` + `claude.yml` workflows, THEN one of the following decisions
SHALL be recorded in design.md AD-004 for each workflow:

- **DELETE** if the workflow depends on a codex / gemini / glm CLI that is not installed
  on the runner (verified by `grep` on workflow body for `codex`/`gemini`/`glm` CLI
  invocations).
- **PRESERVE** if the workflow is orthogonal to the deleted review bots (e.g., depends
  only on `gh` CLI + GitHub native APIs + Claude API).
- **GUARD** if the workflow has ambiguous dependency: wrap the dependent step with
  `if: command -v <cli> >/dev/null 2>&1` and add `continue-on-error: true` to prevent
  required-check failure.

Pre-plan recon (Overview §1.3) indicates `private-guard` is a job name inside
`codex-review.yml` + `llm-panel.yml` (both DELETE per REQ-CIFT-003), and that
`review-quality-gate.yml` + `claude.yml` are codex-independent (PRESERVE candidate).
T4 SHALL confirm via grep.

### REQ-CIFT-005 — Local Pre-Flight Gate (Ubiquitous)

A `lefthook.yml` file at the repository root SHALL enforce a `pre-push` hook that
invokes `make preflight`. The `Makefile` SHALL expose a `preflight` target that runs
THREE checks in order:

1. `golangci-lint run --fast` (or `make lint-fast`)
2. `go test -race -short ./...`
3. `go build ./...`

Any non-zero exit from these three checks SHALL block the push. The developer onboarding
flow SHALL include a 1-command install: `brew install lefthook && lefthook install`.

The `lefthook.yml` SHALL allow bypass via environment variable `LEFTHOOK=0` for
emergency or known-flaky cases.

### REQ-CIFT-006 — Nightly Full-Matrix Workflow (Ubiquitous)

A new `.github/workflows/nightly-full-matrix.yml` SHALL execute the full 3-OS test
matrix on the following triggers:

- `schedule: cron "0 3 * * *"` (daily 03:00 UTC)
- `workflow_dispatch` (manual)
- `push: tags: ['v*']` (release tags)

The workflow SHALL run `go test -race ./...` on `ubuntu-latest`, `macos-latest`, and
`windows-latest`. On any matrix-job failure, the workflow SHALL open a GitHub issue
using `actions/github-script` with:

- Title: `Nightly matrix failure: <os> @ <commit-short>`
- Body: link to the failing run + matrix job logs URL
- Labels: `type:ci` / `priority:P1` / `area:ci`
- Deduplication: SHALL search for an OPEN issue with identical title prefix
  (`Nightly matrix failure: <os>`) created within the prior 24h and append a comment
  instead of opening a duplicate.

### REQ-CIFT-007 — CLAUDE.local.md §18.7 Doctrine Sync (Ubiquitous)

The `CLAUDE.local.md` §18.7 section SHALL reflect the post-(B) required status checks
list:

```
Required status checks (4 items):
- Lint
- Test (ubuntu-latest)
- Build (linux/amd64)
- CodeQL
```

The section SHALL also document:

- The (B) decision rationale (2026-05-17 user policy: 1인 개발, macOS, 5-6분+ wait 단축)
- The 3-tier CI philosophy:
  - **Tier 1 (per-PR fast)**: ubuntu-latest Go test + Lint + Build + CodeQL — required
  - **Tier 2 (per-PR opt-in)**: macos-latest / windows-latest — informational only,
    NOT required
  - **Tier 3 (nightly + tag)**: full 3-OS matrix via nightly-full-matrix.yml — safety net
- Cross-reference to this SPEC ID (`SPEC-V3R4-CI-FASTTRACK-001`) for traceability.

### REQ-CIFT-008 — Lessons Capture (Ubiquitous)

The `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/lessons.md` file SHALL
gain a new entry `#18 — 1인 개발 CI 3-tier pattern (2026-05-17)` following the
MoAI Lessons Protocol format:

```
## #18 — 1인 개발 CI 3-tier pattern (2026-05-17)

**Category**: workflow

**Incorrect approach**: 3-OS Go test 매트릭스를 모든 PR 의 required check 로 강제하면 1인
개발자의 PR-level wait 비용이 5-6분+ 으로 누적되어 flow state 가 끊긴다. Over-engineered
review bot 4개 (codex/gemini/glm/private-guard) 가 매 PR 마다 RED 로 떨어지면서 신호 대 잡음 비율이
저하된다.

**Correct approach**: 1인 개발자는 CI 비용을 3-tier 로 분할한다.
- Tier 1 (per-PR fast): ubuntu-latest 만 required. paths-filter 로 docs-only PR 은 skip.
- Tier 2 (per-PR opt-in): macOS/Windows 는 nightly 로 이전 (per-PR 비강제).
- Tier 3 (nightly + tag): full 3-OS matrix 는 일일 03:00 UTC + tag push 시점에만.
Review bot 은 단 1개 (claude-code-review) 만 신뢰. 다른 LLM panel 은 본질적 redundancy.

**Why**: PR-level wait 1분 절감 × 일일 10 PR × 250 영업일 = 41시간/년. 회귀 탐지 latency 24h
지연은 1인 개발자에게 acceptable (즉시 production 배포 cadence 가 아님).

**How to apply**: 미래 SPEC 에서 CI workflow 추가/수정 시 (a) required check 인가 (b)
informational 인가를 명시적으로 분리할 것. paths-filter + skip-marker 패턴은 branch
protection compatibility 의 표준 패턴. nightly safety-net 없이 required check 만 줄이면
회귀 탐지가 영구 손실됨.
```

## 5. Affected Files

### MODIFY (existing files)

| File | Estimated Δ | Change Summary |
|------|-------------|----------------|
| `.github/workflows/ci.yml` | +60 / -10 LOC | Add `detect` job (paths-filter) + conditional matrix + skip-marker job. Keep race detector + fetch-depth: 0 + zero-exec-Command assertion. |
| `.github/workflows/codeql.yml` | +6 / -0 LOC | Add `paths-ignore` stanza in `on.pull_request`. Other triggers unchanged. |
| `Makefile` | +20 / -0 LOC | Add `preflight` target chaining `lint-fast` + `test-race-short` + `build`. |
| `CLAUDE.local.md` (§18.7) | ~15 LOC | Update required checks list 6→4 + 3-tier philosophy + (B) rationale. |
| `~/.claude/projects/.../memory/lessons.md` | +40 LOC | New entry #18 per REQ-CIFT-008. |
| `~/.claude/projects/.../memory/MEMORY.md` | +1 LOC | New index entry pointing to lessons #18 (or project memory if entry created). |

### DELETE (5 review workflows)

| File | LOC removed (approx) | Reason |
|------|----------------------|--------|
| `.github/workflows/codex-review.yml` | ~80 | codex CLI not installed; permanent RED; contains `private-guard` job. |
| `.github/workflows/gemini-review.yml` | ~70 | gemini CLI not installed; permanent RED. |
| `.github/workflows/glm-review.yml` | ~70 | glm CLI not configured for review on this repo; redundancy. |
| `.github/workflows/llm-panel.yml` | ~50 | Aggregator over the 3 deleted bots; orphan after their removal. Also contains `private-guard`. |
| `.github/workflows/claude-code-review.optional.yml` | ~30 | Duplicate of `claude-code-review.yml`; orphan. |

### NEW

| File | Estimated Δ | Purpose |
|------|-------------|---------|
| `.github/workflows/nightly-full-matrix.yml` | +120 LOC | Daily + tag-push full 3-OS matrix safety net (REQ-CIFT-006). |
| `lefthook.yml` | +30 LOC | Repository-root pre-push hook (REQ-CIFT-005). |

### PRESERVE (unchanged, audit-only)

| File | Audit decision | Reason |
|------|----------------|--------|
| `.github/workflows/claude-code-review.yml` | PRESERVE | User-approved review path. Claude API + Anthropic OAuth already configured. |
| `.github/workflows/claude.yml` | PRESERVE | `@claude` mention trigger on issues/comments. codex-independent. |
| `.github/workflows/review-quality-gate.yml` | PRESERVE | Parses Claude Code Review check-run severity. codex-independent. |
| `.github/workflows/auto-merge.yml` | PRESERVE | Dependabot auto-merge. Orthogonal scope. |
| `.github/workflows/community.yml` | PRESERVE | Community guidelines. Orthogonal scope. |
| `.github/workflows/docs-i18n-check.yml` | PRESERVE | docs-site i18n sync gate. Required by §17. |
| `.github/workflows/label-sync.yml` | PRESERVE | `.github/labels.yml` sync. Orthogonal scope. |
| `.github/workflows/release-drafter.yml` | PRESERVE | Release notes drafter. Required by §18.9. |
| `.github/workflows/release-drafter-cleanup.yml` | PRESERVE | Release drafter cleanup helper. |
| `.github/workflows/release.yml` | PRESERVE | GoReleaser tag-push handler. Required by §18.9. |

## 6. References

- **Predecessor session**: SPEC-WORKTREE-DOCS-001 (sibling doctrine SPEC, 2026-05-17),
  `feedback_worktree_autonomous` memory.
- **Triggering memory**: 2026-05-17 user report — 1인 개발 CI 비용 비대칭, review bot
  영구 RED.
- **Doctrine target**: `CLAUDE.local.md` §18.7 (current 6-item baseline → 4-item).
- **Branch protection (축 B baseline)**: applied via
  `gh api -X PATCH /repos/modu-ai/moai-adk/branches/main/protection/required_status_checks`
  at main HEAD `680232c82`. Current main HEAD: `41b6f37dc`.
- **Workflow files**: `.github/workflows/ci.yml` (226 LOC, T1 target),
  `.github/workflows/codeql.yml` (52 LOC, T2 target),
  `.github/workflows/claude-code-review.yml` (preserve), `.github/workflows/claude.yml`
  (64 LOC, audit), `.github/workflows/review-quality-gate.yml` (134 LOC, audit).
- **External tooling**: `dorny/paths-filter@v3`, `evilmartians/lefthook`,
  `actions/github-script@v7`, GoReleaser, Release Drafter.
- **Cross-SPEC**: SPEC-V3R4-WORKFLOW-SPLIT-001 (Bundle F, in-progress) — orthogonal,
  no expected conflict (different file scope: `.claude/skills/moai/workflows/*.md` vs
  `.github/workflows/*.yml`).
- **Anti-pattern reference**: Karpathy / forrestchang anti-pattern catalog — this SPEC
  is the *correct* scope for "right-sized CI" and explicitly NOT premature optimization,
  because the triggering report is empirical (5-6분+ measured wait) and the trade-off
  is documented (24h regression latency acceptable).
