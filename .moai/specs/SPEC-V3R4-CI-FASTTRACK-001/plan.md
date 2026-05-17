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

# Plan — SPEC-V3R4-CI-FASTTRACK-001

## 1. Plan Summary

### 1.1 Strategy

**Single-Wave**. 8 tasks (T1..T8) 가 모두 단일 run-PR 안에서 atomic 하게 적용되어야 한다.
근거:

- **Half-applied state risk**: 5 review workflows 를 먼저 제거하고 paths-filter 를 나중에
  적용하면, 중간 PR 마다 docs-only PR 이 여전히 3-OS Go test 매트릭스를 wait. 반대로
  paths-filter 만 먼저 적용하면 review bot RED 가 continue 하여 신호 대 잡음 비율이 개선되지
  않음. 두 변경이 결합되어야 사용자가 실제로 80%+ wait 감소를 경험.
- **CI rollout atomicity**: branch protection (B) 가 이미 활성이므로, 본 PR 머지 후 첫 PR
  부터 즉시 새 게이트가 발효. 다단계 wave 는 결국 한 번에 적용되어야 하므로 wave-split 의
  실익이 없음.
- **LOC scale**: net delta ~ +300 / -300 (5 deletions + 2 new files + 1 large modify) —
  단일 PR review burden 으로 무리 없음.
- **Test scope**: Go code 0 LOC 수정. `go test ./...` 게이트는 sanity-check no-op.

### 1.2 Base branch

- `main` (HEAD `41b6f37dc` 기준; 본 plan-PR 머지 시점에 갱신).
- **Plan-PR branch**: `plan/SPEC-V3R4-CI-FASTTRACK-001` (현재 작업 중인 브랜치)
- **Run-PR branch**: `feat/SPEC-V3R4-CI-FASTTRACK-001-fasttrack-impl`
  (plan-PR 머지 후 main 에서 생성)
- **Sync-PR branch**: `sync/SPEC-V3R4-CI-FASTTRACK-001` (run-PR 머지 후 생성)

### 1.3 Working location

- **Main checkout** (`/Users/goos/MoAI/moai-adk-go`).
- 2026-05-17 user policy (`feedback_worktree_autonomous`) 에 따라 worktree 사용 안 함.
- 세 phase (plan / run / sync) 모두 main checkout 의 feature branch 위에서 실행.

### 1.4 PR sequence

1. **Plan-PR** (이 산출물): `plan/SPEC-V3R4-CI-FASTTRACK-001` → squash merge into main.
2. **Run-PR**: `feat/SPEC-V3R4-CI-FASTTRACK-001-fasttrack-impl` → squash merge into main.
3. **Sync-PR**: `sync/SPEC-V3R4-CI-FASTTRACK-001` → squash merge into main.

## 2. Wave Decomposition

**Single-Wave SPEC**. Wave decomposition 불필요. 근거는 §1.1 참조.

## 3. Tasks

8 sequential tasks (T1..T8). T1..T3 가 ci.yml + codeql.yml + review workflow consolidation
의 핵심 atomic 단위. T4 는 audit-only (코드 변경 결정 boundary). T5..T6 는 새 파일 추가.
T7..T8 은 doctrine + memory layer 동기화.

### T1 — ci.yml `dorny/paths-filter@v3` Conditional Matrix + Skip-Marker

**Deliverable**: `.github/workflows/ci.yml` 의 `jobs` 섹션을 다음과 같이 재구성:

1. 신규 `detect` job: `dorny/paths-filter@v3` 으로 docs-only 여부 판정. outputs:
   `go_code` (bool), `docs_only` (bool).
2. 기존 `test` job 에 `needs: detect` + `if: needs.detect.outputs.go_code == 'true'` 가드.
   matrix / strategy / steps 본문은 변경 없음 (race detector + fetch-depth: 0 + 기존
   AC-UTIL-001-04 zero-exec-Command assertion 모두 유지).
3. 신규 `test-skip-marker` job: docs-only PR 일 때 (`if: needs.detect.outputs.docs_only ==
   'true'`) 트리거. matrix: `[ubuntu-latest, macos-latest, windows-latest]`. 단일 step
   `echo "Docs-only PR — Go test matrix skipped per paths-filter"`. **Critical**:
   `name:` 필드를 `Test (${{ matrix.os }})` 로 EXACT 매칭 (branch protection 의 required
   check name 충족 위해).
4. Build / Lint / CodeQL job 은 paths-filter 영향 받지 않음 (Lint / Build 는 branch
   protection required 이므로 paths-filter 적용 불가; 별도 fast 동작).

paths-filter 패턴 (REQ-CIFT-001 와 동기화):

```yaml
docs_only:
  - '**.md'
  - '.moai/specs/**'
  - '.moai/docs/**'
  - '.moai/reports/**'
  - '.moai/research/**'
  - '.claude/rules/**'
  - '.claude/skills/**'
  - 'docs-site/**'
  - 'CHANGELOG.md'
  - 'LICENSE'
  - 'NOTICE'
go_code:
  - '!**.md'
  - '!.moai/**'
  - '!.claude/**'
  - '!docs-site/**'
  - '**/*.go'
  - 'go.mod'
  - 'go.sum'
  - 'Makefile'
  - '.github/workflows/ci.yml'
  - '.github/workflows/codeql.yml'
```

**Verification**:

- `yamllint .github/workflows/ci.yml` 통과.
- 빈 docs-only PR 시뮬레이션 (브랜치에 markdown 1줄만 변경 후 PR 생성 → CI 트리거 →
  `gh pr checks` 출력에서 `Test (ubuntu-latest)` 가 PASS (skip-marker) + 실제 test job
  은 skipped 상태).
- 빈 Go PR 시뮬레이션 (`internal/cli/main.go` 등에 무해한 주석 1줄 변경) 에서 full matrix
  trigger 확인.
- branch protection 의 `Test (ubuntu-latest)` required check 가 skip-marker pass 로
  satisfy 되는지 `gh pr view <PR> --json statusCheckRollup` 으로 확인.

**Implements**: REQ-CIFT-001.

### T2 — codeql.yml paths-ignore Stanza

**Deliverable**: `.github/workflows/codeql.yml` 의 `on.pull_request` 섹션에 `paths-ignore`
stanza 추가 (REQ-CIFT-002 의 정확한 패턴 적용). `on.push` / `on.schedule` 는 변경 없음.

```yaml
on:
  pull_request:
    branches: [main]
    paths-ignore:
      - '**.md'
      - '.moai/specs/**'
      - '.moai/docs/**'
      - '.moai/reports/**'
      - '.moai/research/**'
      - '.claude/rules/**'
      - 'docs-site/**'
  push:
    branches: [main]
  schedule:
    - cron: '<existing>'
```

**Verification**:

- `yamllint .github/workflows/codeql.yml` 통과.
- 빈 docs-only PR 에서 CodeQL workflow run 이 skip / not-applicable 상태 (CodeQL 의
  required check 는 branch protection 에 포함이므로, skip 시 어떻게 동작하는지 사전 확인
  필요 — 보통 paths-ignore 시 workflow 자체가 트리거되지 않고 required check 는 success
  로 간주되는 GitHub 동작에 의존; T1 의 skip-marker 패턴과 동일한 보호가 필요할 수 있음.
  설계는 design.md AD-002 에서 확정).
- 빈 Go PR 에서 CodeQL 정상 실행.

**Implements**: REQ-CIFT-002.

### T3 — Delete 5 Review Workflows

**Deliverable**: 다음 5개 파일을 `git rm` 으로 제거:

```bash
git rm .github/workflows/codex-review.yml
git rm .github/workflows/gemini-review.yml
git rm .github/workflows/glm-review.yml
git rm .github/workflows/llm-panel.yml
git rm .github/workflows/claude-code-review.optional.yml
```

**Verification**:

- `ls .github/workflows/ | wc -l` 이전 18 → 13 (T6 nightly 추가 전 시점).
- `gh api repos/modu-ai/moai-adk/actions/workflows --jq '.workflows[].path'` 출력에서
  위 5개 path 부재 확인.
- 신규 PR 생성 시 5개 review workflow run 트리거 없음 (review 영역은 `claude-code-review.yml`
  단독 트리거).

**Implements**: REQ-CIFT-003.

### T4 — Audit `claude.yml` + `review-quality-gate.yml` + private-guard

**Deliverable**: 다음 audit 를 수행 후 design.md AD-004 의 decision matrix 를 확정:

1. `grep -nE 'codex|gemini|glm|sg [^a-z]' .github/workflows/claude.yml` →
   결과 0 → **PRESERVE** (codex 비의존, `@claude` issue/comment trigger 만).
2. `grep -nE 'codex|gemini|glm' .github/workflows/review-quality-gate.yml` →
   결과 0 → **PRESERVE** (Claude Code Review check_run severity parser 만).
3. `grep -rln 'private-guard\|private_guard' .github/workflows/` →
   결과: `codex-review.yml` + `llm-panel.yml` (둘 다 T3 에서 DELETE).
   결론: private-guard 는 자동 소멸. 별도 처리 불필요.

design.md AD-004 의 decision template 에 위 3개 결과를 명시. 추가 workflow 가 발견되면
같은 절차로 분류 (DELETE / PRESERVE / GUARD 중 선택).

**Verification**:

- design.md AD-004 가 PRESERVE 2건 + 자동소멸 1건 (private-guard) 으로 갱신됨.
- run-PR diff 에서 `claude.yml` / `review-quality-gate.yml` 의 변경 없음 (no-op).

**Implements**: REQ-CIFT-004.

### T5 — lefthook.yml + Makefile `preflight` Target

**Deliverable**:

1. `lefthook.yml` (NEW, repo root) — pre-push hook 정의:

   ```yaml
   pre-push:
     commands:
       preflight:
         run: make preflight
   ```

2. `Makefile` (MODIFY) — `preflight` 타깃 추가:

   ```makefile
   .PHONY: preflight lint-fast test-race-short

   preflight: lint-fast test-race-short build
   	@echo "preflight: OK"

   lint-fast:
   	@golangci-lint run --fast ./... || (echo "preflight: lint-fast FAIL"; exit 1)

   test-race-short:
   	@go test -race -short ./... || (echo "preflight: test-race-short FAIL"; exit 1)
   ```

   `build` 타깃은 기존 Makefile 에 이미 존재.

3. `README.md` (또는 `CONTRIBUTING.md`) onboarding 섹션에 1-line install 안내:
   `brew install lefthook && lefthook install` (Run-phase 가 결정; 최소한 `README.md`
   상단 또는 `CLAUDE.local.md` §6 Testing Guidelines 직후에 추가).

**Verification**:

- `lefthook --version` 통과 (시스템에 설치된 경우).
- `lefthook install` 후 `.git/hooks/pre-push` 가 lefthook wrapper 로 갱신됨.
- `make preflight` (clean main checkout 에서) 0 종료.
- `lefthook run pre-push` 가 `make preflight` 트리거.
- `LEFTHOOK=0 git push` (드라이런) 시 hook 우회 가능.

**Implements**: REQ-CIFT-005.

### T6 — nightly-full-matrix.yml Workflow

**Deliverable**: `.github/workflows/nightly-full-matrix.yml` (NEW):

```yaml
name: Nightly Full Matrix

on:
  schedule:
    - cron: "0 3 * * *"
  workflow_dispatch:
  push:
    tags: ['v*']

concurrency:
  group: nightly-${{ github.ref }}
  cancel-in-progress: false

jobs:
  test-full-matrix:
    name: Test full matrix (${{ matrix.os }})
    runs-on: ${{ matrix.os }}
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    steps:
      - uses: actions/checkout@v5
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v6
        with:
          go-version: "1.26"
          cache: true
      - run: go mod download
      - run: go vet ./...
      - run: go test -race ./...

  notify-on-failure:
    needs: test-full-matrix
    if: failure() && github.event_name == 'schedule'
    runs-on: ubuntu-latest
    permissions:
      issues: write
    steps:
      - uses: actions/github-script@v7
        with:
          script: |
            const title_prefix = `Nightly matrix failure`;
            // Dedup: search OPEN issues with matching prefix within 24h
            const since = new Date(Date.now() - 24*60*60*1000).toISOString();
            const search = await github.rest.search.issuesAndPullRequests({
              q: `repo:${context.repo.owner}/${context.repo.repo} is:open is:issue ` +
                 `in:title "${title_prefix}" created:>${since}`,
            });
            const body = `Nightly matrix run failed. Logs: ${context.serverUrl}/${context.repo.owner}/${context.repo.repo}/actions/runs/${context.runId}`;
            if (search.data.total_count > 0) {
              await github.rest.issues.createComment({
                owner: context.repo.owner, repo: context.repo.repo,
                issue_number: search.data.items[0].number,
                body: `Repeat failure: ${body}`,
              });
            } else {
              await github.rest.issues.create({
                owner: context.repo.owner, repo: context.repo.repo,
                title: `${title_prefix} @ ${context.sha.substring(0,7)}`,
                body: body,
                labels: ['type:ci', 'priority:P1', 'area:ci'],
              });
            }
```

**Verification**:

- `yamllint .github/workflows/nightly-full-matrix.yml` 통과.
- `gh workflow list` 출력에 `Nightly Full Matrix` 포함.
- `gh workflow run nightly-full-matrix.yml` 으로 수동 트리거 → 3-OS matrix 정상 실행.
- 인위적 실패 주입 후 (별도 PR 에서 검증) issue 생성 + dedup 동작 확인 (run-PR scope 외,
  AC-CIFT-006 에 deferred).

**Implements**: REQ-CIFT-006.

### T7 — CLAUDE.local.md §18.7 Doctrine Sync

**Deliverable**: `CLAUDE.local.md` §18.7 Branch Protection Rule (GitHub) 섹션 정정:

1. **Required status checks 목록** 6 → 4 항목으로 갱신:
   - `Lint`
   - `Test (ubuntu-latest)`
   - `Build (linux/amd64)`
   - `CodeQL`
2. **(B) 결정 rationale** 1단락 추가:
   - 2026-05-17 사용자 정책: 1인 개발, macOS 환경, 5-6분+ wait 비현실적.
   - `Test (macos-latest)` / `Test (windows-latest)` 제거. nightly matrix 로 이전.
3. **3-tier CI philosophy** 명문화 (REQ-CIFT-007 본문):
   - Tier 1 (per-PR fast / required): ubuntu Go test + Lint + Build + CodeQL
   - Tier 2 (per-PR opt-in / informational): macOS / Windows
   - Tier 3 (nightly + tag): full 3-OS matrix
4. **Cross-reference**: `(SPEC: SPEC-V3R4-CI-FASTTRACK-001)` 표기.

**Verification**:

- `grep -nE '"contexts": \[' CLAUDE.local.md` 결과 (`gh api ... required_status_checks`
  명령 안의 JSON) 에서 contexts 가 정확히 4개 항목으로 명시.
- `grep -n "3-tier\|Tier 1\|Tier 2\|Tier 3" CLAUDE.local.md` 가 §18.7 부근에서 최소 4
  매치.
- `grep -n "SPEC-V3R4-CI-FASTTRACK-001" CLAUDE.local.md` 최소 1 매치.

**Implements**: REQ-CIFT-007.

### T8 — lessons.md #18 Capture

**Deliverable**: `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/lessons.md` 에
새 entry `#18 — 1인 개발 CI 3-tier pattern (2026-05-17)` 추가. 본문은 REQ-CIFT-008 본문
verbatim.

또한 `MEMORY.md` 에 (필요 시) entry 추가 — 단, lessons.md 본문이 분리 파일이 아닌 단일
`lessons.md` 의 #18 항목이라면 별도 MEMORY.md 인덱싱은 lessons.md 자체로 충분 (auto-memory
규약상 lessons.md 는 단일 파일 chronicle 패턴).

**Verification**:

- `grep -n "^## #18" lessons.md` 1 매치.
- entry 본문이 Category / Incorrect / Correct / Why / How-to-apply 5개 구조를 포함.
- `head -3 lessons.md | grep -c "^---"` ≥ 2 (frontmatter 보존).

**Implements**: REQ-CIFT-008.

## 4. Risk Analysis

### R1 — paths-filter False-Negative (Go change misclassified as docs-only)

**Risk**: paths-filter 패턴이 어떤 Go 변경 (예: `.github/workflows/ci.yml` 자체 변경,
혹은 `go.mod` 만 단독 변경) 을 docs-only 로 잘못 분류하면 회귀가 production 까지 새 어 들어갈
가능성.

**Mitigation**:

- paths-filter 패턴에 `**/*.go`, `go.mod`, `go.sum`, `Makefile`, `.github/workflows/ci.yml`,
  `.github/workflows/codeql.yml` 을 명시적으로 `go_code` 카테고리 포함.
- Nightly full matrix (T6) 가 24h 안전망. 회귀 발생 시 다음 nightly run 에서 issue 자동
  생성.
- design.md AD-001 의 conservative filter pattern 명시.

### R2 — Skip-Marker Job Name Collision with Real Test Job

**Risk**: `Test (ubuntu-latest)` 라는 동일한 check name 을 두 job 이 둘 다 가지면 branch
protection 의 required check matching 이 혼란 (어느 한쪽만 매칭되거나, 둘 다 매칭되어 둘
중 하나라도 fail 시 block).

**Mitigation**:

- T1 에서 두 job 은 mutually exclusive `if:` 가드. 동일 PR 에서 둘 다 트리거되지 않음.
- 동일 check name 은 의도된 설계 — GitHub Actions 의 conditional matrix 패턴이
  branch protection compatibility 의 표준 우회법 (design.md AD-002).
- 사전 검증: branch protection api response 의 `required_status_checks.contexts` 가
  `Test (ubuntu-latest)` 단일 entry 인지 확인. 한 PR 의 두 job 중 정확히 한 개만 run
  하면 OK.

### R3 — lefthook Installation Friction

**Risk**: `lefthook` 이 머신에 사전 설치되어 있지 않으면 `make preflight` 자체는 동작하지만
hook 이 동작하지 않아 RED PR 이 push 됨.

**Mitigation**:

- README onboarding 섹션에 `brew install lefthook && lefthook install` 명시 (T5
  deliverable 3).
- `Makefile preflight` 타깃이 lefthook 의존 없이 단독 호출 가능 (사용자가 수동으로 `make
  preflight && git push` 가능).
- lefthook 미설치 시 `git push` 가 정상 동작 (warning 만, blocking 없음) — 안전한 graceful
  degradation.

### R4 — Nightly Full-Matrix Failure Noise (intermittent windows flake)

**Risk**: Windows runner 의 간헐적 flake (ETXTBSY race per CLAUDE.local.md §18.11 등) 가
nightly 마다 issue 를 생성하면 issue queue 가 오염됨.

**Mitigation**:

- T6 의 dedup 로직 (24h window 내 동일 prefix issue 발견 시 댓글 append). 같은 flake 가
  연속 발생해도 issue 는 1개만 유지.
- issue title 에 commit SHA (short) 포함 → tag-push / 동일 commit 재실행 dedup 도 자연
  발생.
- Run-phase 에서 GitHub Actions retry mechanism 추가 검토 (별도 SPEC 필요, 본 SPEC scope 외).

### R5 — `review-quality-gate` / `claude.yml` 우발 삭제

**Risk**: T3 batch deletion 의 부주의로 audit 대상 (PRESERVE) 인 workflow 가 함께 삭제됨.

**Mitigation**:

- T3 의 `git rm` 명령은 5개 파일 이름을 *명시적*으로 나열. wildcard / glob 사용 금지.
- T4 audit 가 T3 *이전* 에 수행 — design.md AD-004 에 PRESERVE 결정이 기록된 후 T3 실행.
- Acceptance gate AC-CIFT-003 + AC-CIFT-004 가 정확히 5개 삭제 + 3개 (claude-code-review,
  claude, review-quality-gate) 보존을 binary 검증.
- run-PR diff 가 자동으로 변경 파일 목록을 노출 — 사용자 검토 가능.

## 5. Mitigation Summary

5개 risk 모두 명시적 mitigation 보유. R1 (false-negative) 과 R5 (우발 삭제) 가 highest-impact;
둘 다 acceptance gate 의 binary 검증 + nightly safety net 으로 차단. R2 (skip-marker
collision) 는 설계 의도이므로 risk 가 아니라 trade-off; design.md AD-002 가 정당화.

## 6. Implementation Sequence

```
Plan-PR (이 산출물):
  branch: plan/SPEC-V3R4-CI-FASTTRACK-001
  files: 5 markdown artifacts in .moai/specs/SPEC-V3R4-CI-FASTTRACK-001/
  merge: squash → main
  status: draft → planned (post-merge via spec-status-sync)

Run-PR:
  branch: feat/SPEC-V3R4-CI-FASTTRACK-001-fasttrack-impl
  files:
    - MODIFY: .github/workflows/ci.yml (T1)
    - MODIFY: .github/workflows/codeql.yml (T2)
    - DELETE: .github/workflows/{codex-review,gemini-review,glm-review,llm-panel,claude-code-review.optional}.yml (T3)
    - AUDIT: .github/workflows/{claude,review-quality-gate}.yml (T4, no-edit)
    - NEW: lefthook.yml + MODIFY: Makefile (T5)
    - NEW: .github/workflows/nightly-full-matrix.yml (T6)
    - MODIFY: CLAUDE.local.md §18.7 (T7)
    - MODIFY: ~/.claude/projects/.../memory/lessons.md (T8)
  task order: T4 (audit) → T1 → T2 → T3 → T5 → T6 → T7 → T8
  gates: AC-CIFT-001..008 + PG-1..2 + EC-1..4 ALL PASS
  merge: squash → main

Sync-PR:
  branch: sync/SPEC-V3R4-CI-FASTTRACK-001
  files:
    - spec.md frontmatter status: planned → completed
    - CHANGELOG.md (top-of-file entry)
    - lessons.md entry verification (T8 보강 / consistency check)
  merge: squash → main
```

## 7. Out of Scope

spec.md §3 (Non-Goals) 의 mirror. 명시적 DROP 목록:

- Self-hosted runner setup
- Test sharding
- Main matrix 의 모든 OS 제거
- `claude-code-review.yml` 제거
- 임의 review bot 재활성화 (codex / gemini / glm 부활)
- macOS-latest 의 required check 재추가
- GitHub Actions → 외부 CI provider 이전
- v3.0.0-rc1 release tag 발행 actions
- `claude.yml` / `review-quality-gate.yml` 자동 제거 (T4 audit 결과로 PRESERVE)
- CI cache 전략 변경
