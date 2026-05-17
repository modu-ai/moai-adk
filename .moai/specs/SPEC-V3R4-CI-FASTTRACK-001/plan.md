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

**Two-Wave SPEC** (revised post-plan-audit iteration 1):

- **Wave 0 — Skip-Marker Proof-of-Concept** (sandbox isolation, throw-away PR).
  Validates that GitHub Actions의 "two jobs with identical `name`, mutually exclusive
  `if:` guards" 패턴이 docs-only PR 에서 branch-protection 의 required check 를 satisfy
  하는지 empirically 확인. plan-auditor 의 P0 defect D4 (skip-marker correctness
  unverified) 직접 대응. Wave 0 PASS 이후에만 Wave 1 진입.
- **Wave 1 — Main Implementation** (T1..T8). 단일 run-PR atomic 적용.

Wave 0 의 sandbox PR 은 main 으로 머지되지 않음 (close + delete-branch). 산출물은
Wave 0 의 verification report (design.md AD-002 에 PASS/FAIL 기록) 만 main 으로
귀결됨. Wave 1 run-PR 본문에 Wave 0 PR link 를 참조.

## 3. Tasks

9 sequential tasks (T0..T8). T0 는 Wave 0 의 PoC sandbox. T1..T3 가 ci.yml + codeql.yml +
review workflow consolidation 의 핵심 atomic 단위. T4 는 audit-only (코드 변경 결정 boundary).
T5..T6 는 새 파일 추가. T7..T8 은 doctrine + memory layer 동기화.

### T0 — Wave 0 Skip-Marker Proof-of-Concept (Wave 0)

**Deliverable**: 별도 throw-away PR (`test/skip-marker-poc` branch) 에서 skip-marker
pattern 의 GitHub Actions 동작을 isolated 환경에서 검증.

**Procedure**:

1. **CodeQL check-run name 사전 확인**:

   ```bash
   gh api repos/modu-ai/moai-adk/commits/main/check-runs --jq '.check_runs[].name' \
     | grep -i codeql
   ```

   2026-05-17 main HEAD 기준 expected output: `Analyze (Go) (go)` (workflow `CodeQL`
   + job name `Analyze (Go)` + matrix `language: go`). Branch protection 의 bare
   `CodeQL` 은 legacy workflow-name match — Wave 0 가 어느 name 이 실제 required check
   를 satisfy 하는지 binary 확인.

2. **Sandbox workflow 작성**: `.github/workflows/skip-marker-poc.yml` (throw-away,
   T0 branch 한정):

   ```yaml
   name: PoC Skip-Marker
   on:
     pull_request:
       branches: [main]
   jobs:
     detect:
       runs-on: ubuntu-latest
       outputs:
         docs_only: ${{ steps.f.outputs.docs_only }}
       steps:
         - uses: actions/checkout@v5
         - uses: dorny/paths-filter@v3
           id: f
           with:
             filters: |
               docs_only:
                 - '**.md'
     test:
       needs: detect
       if: needs.detect.outputs.docs_only != 'true'
       runs-on: ubuntu-latest
       name: PoC Test
       steps:
         - run: echo "real test"
     test-skip-marker:
       needs: detect
       if: needs.detect.outputs.docs_only == 'true'
       runs-on: ubuntu-latest
       name: PoC Test
       steps:
         - run: echo "skipped (paths-filter)"
   ```

3. **Open PR with markdown-only diff** (e.g., README.md 1-line edit). Wait for CI.

4. **Verify**:
   - `gh pr checks <T0-PR> --json name,state -q '.[] | select(.name=="PoC Test") | .state'`
     → expected: `SUCCESS` (skip-marker job).
   - Workflow run history: `test` job in `skipped` state, `test-skip-marker` in `success`.
   - Branch protection 의 sandbox PR 에 대한 영향 0 (PoC 는 required check 가 아님).

5. **Record result in design.md AD-002**: PASS / FAIL + observed check-run name(s).
   Reference: https://github.com/orgs/community/discussions/13690 (canonical community
   discussion on the skip-marker pattern).

6. **Cleanup**: `gh pr close <T0-PR> --delete-branch` + `git push origin --delete
   test/skip-marker-poc`.

**Decision matrix**:

| Wave 0 result | Wave 1 action |
|---------------|---------------|
| PoC Test = SUCCESS | Proceed to Wave 1 (T1..T8) with skip-marker pattern confirmed |
| PoC Test = PENDING-forever | Abort Wave 1, escalate (skip-marker pattern unsupported on this runner config — fall back to alternative design, e.g., GitHub Actions `workflow_call` or `composite-action`-based gate) |
| PoC Test = FAILURE | Inspect logs, fix PoC, re-run T0 |

**Implements**: design.md AD-002 PoC verification gate.

**Out of Wave 1 scope**: T0 's sandbox workflow file is NOT merged into main. Only its
recorded result in AD-002 + reference link in run-PR description matters.

### T1 — ci.yml `dorny/paths-filter@v3` Conditional Matrix + Skip-Marker (Wave 1)

**Pre-condition**: Wave 0 (T0) PASS 가 design.md AD-002 에 기록됨.

**Deliverable**: `.github/workflows/ci.yml` 의 `jobs` 섹션을 다음과 같이 재구성. ci.yml
은 실제로 **5개 job** 으로 구성되어 있으며, 각 job 의 conditional 적용 정책을 명시한다.

**Pre-existing ci.yml 5-job inventory** (2026-05-17 main HEAD `41b6f37dc`):

| Job ID | Name field | Required check? | LOC scope |
|--------|-----------|-----------------|-----------|
| `test` | `Test (${{ matrix.os }})` | Yes (only `Test (ubuntu-latest)`) | L18-86 |
| `test-integration` | `Integration Tests (${{ matrix.os }})` | No | L87-114 |
| `lint` | `Lint` | Yes | L118-139 |
| `build` | `Build (${{ matrix.goos }}/${{ matrix.goarch }})` | Yes (only `Build (linux/amd64)`) | L143-196 |
| `constitution-check` | (workflow default) | No (`continue-on-error: true`) | L197+ |

**ci.yml 5-job behavior table** (post-T1, docs-only PR 시 동작):

| Job | docs-only PR 동작 | go-code PR 동작 | Rationale |
|-----|-------------------|------------------|-----------|
| `test` | SKIP (via `if: needs.detect.outputs.go_code == 'true'`) — skip-marker job 이 `Test (ubuntu-latest)` SUCCESS 발행 | Full 3-OS matrix 실행 (변경 없음) | branch-protection required check name match via skip-marker |
| `test-integration` | SKIP (자동, `needs: test` 의 GitHub Actions semantics — `test` job 이 if-skip 되면 dependent 도 skip) | 실행 (변경 없음) | NOT in branch protection (advisory). 별도 skip-marker 불필요 — `gh pr checks` 에 `Integration Tests (...)` 가 skipped 로 표시되어도 mergeable. |
| `lint` | ALWAYS RUN (변경 없음) | ALWAYS RUN | Required check `Lint`. lint 자체가 fast (~1분) — 가드할 가치 없음. |
| `build` | ALWAYS RUN — `Build (linux/amd64)` required check 보장. macOS/Windows 등 다른 build matrix slot 은 advisory 이므로 conditional gate 추가 안 함 (over-engineering 회피). | ALWAYS RUN | Required check `Build (linux/amd64)` satisfy 필수. 추가 4 platform 은 GoReleaser cross-compile sanity, paths-filter 적용 시 cost-benefit marginal. |
| `constitution-check` | ALWAYS RUN (`continue-on-error: true` 유지) | ALWAYS RUN | NOT required check. fast (~30초). best-effort. |

**Job-level changes**:

1. 신규 `detect` job: `dorny/paths-filter@v3` 으로 docs-only 여부 판정. outputs:
   `go_code` (bool), `docs_only` (bool). All other jobs `needs: detect` 또는 자연 dependency
   chain 으로 의존.
2. 기존 `test` job 에 `needs: detect` + `if: needs.detect.outputs.go_code == 'true'` 가드.
   matrix / strategy / steps 본문은 변경 없음 (race detector + fetch-depth: 0 + 기존
   AC-UTIL-001-04 zero-exec-Command assertion 모두 유지).
3. 신규 `test-skip-marker` job: docs-only PR 일 때 (`if: needs.detect.outputs.docs_only ==
   'true'`) 트리거. matrix: `[ubuntu-latest, macos-latest, windows-latest]`. 단일 step
   `echo "Docs-only PR — Go test matrix skipped per paths-filter"`. **Critical**:
   `name:` 필드를 `Test (${{ matrix.os }})` 로 EXACT 매칭 (branch protection 의 required
   check name 충족 위해).
4. `test-integration`, `lint`, `build`, `constitution-check` 의 `needs:` 와 step 본문은
   변경 없음. `test-integration` 의 `needs: test` 는 그대로 유지 (test skip 시 dependent
   도 skip 되는 GitHub Actions native behavior 의존).

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

### T2 — codeql.yml Skip-Marker Pattern (mirrors T1)

**Pre-condition**: Wave 0 (T0) 결과로 design.md AD-002 에 CodeQL required-check 의 실제
satisfying name 이 binary 기록됨.

**Pre-T2 subtask — T2.0 Verify CodeQL check-run name** (REQ-CIFT-002b):

```bash
gh api repos/modu-ai/moai-adk/commits/main/check-runs --jq '.check_runs[].name' \
  | grep -i codeql
```

2026-05-17 main HEAD 기준 expected output: `Analyze (Go) (go)`. Branch protection
`contexts` 에는 bare `CodeQL` 이 있음 — GitHub 의 legacy workflow-name 매칭에 의존.
T2.0 가 어느 name 이 required check 를 satisfy 하는지 (workflow name `CodeQL` 또는
check-run name `Analyze (Go) (go)`) binary 확인 후 T2 deliverable 의 skip-marker
`name:` 필드를 그에 맞춰 설정.

**Deliverable**: `.github/workflows/codeql.yml` 을 다음과 같이 재구성 (T1 와 동일한
detect-job + skip-marker 패턴 적용. bare `paths-ignore` 는 사용 금지 — branch protection
의 `expected, never reported` block 위험 회피):

1. 신규 `detect` job: `dorny/paths-filter@v3` (T1 와 동일한 path inventory — `docs_only`
   + `go_code`). outputs propagation.
2. 기존 `analyze` job 에 `needs: detect` + `if: needs.detect.outputs.go_code == 'true'`
   가드. matrix `language: [go]` + steps 본문은 변경 없음.
3. 신규 `analyze-skip-marker` job: docs-only PR 시 (`if: needs.detect.outputs.docs_only
   == 'true'`) 트리거. 단일 step `echo "Docs-only PR — CodeQL skipped via skip-marker"`.
   **Critical**: `name:` 필드를 T2.0 에서 확인된 satisfying name 으로 EXACT 설정 (예:
   `Analyze (Go)` + matrix `language: go` → emitted name `Analyze (Go) (go)`, 또는
   workflow name `CodeQL` 매칭이라면 별도 단일 job).
4. `on.push: branches: [main]` / `on.schedule` 는 변경 없음.

Pseudocode (정확한 indentation 은 run-phase 가 확정):

```yaml
name: CodeQL
on:
  pull_request:
    branches: [main]
  push:
    branches: [main]
  schedule:
    - cron: '<existing>'
jobs:
  detect:
    runs-on: ubuntu-latest
    outputs:
      go_code: ${{ steps.f.outputs.go_code }}
      docs_only: ${{ steps.f.outputs.docs_only }}
    steps:
      - uses: actions/checkout@v5
      - uses: dorny/paths-filter@v3
        id: f
        with:
          filters: |
            docs_only:
              - '**.md'
              - '.moai/specs/**'
              # (T1 과 동일한 docs_only 인벤토리)
            go_code:
              - '**/*.go'
              - 'go.mod'
              - 'go.sum'
              - 'Makefile'
              - '.github/workflows/ci.yml'
              - '.github/workflows/codeql.yml'
  analyze:
    needs: detect
    if: needs.detect.outputs.go_code == 'true'
    # (기존 analyze 본문 그대로)
  analyze-skip-marker:
    needs: detect
    if: needs.detect.outputs.docs_only == 'true'
    runs-on: ubuntu-latest
    name: Analyze (Go)  # T2.0 에서 확정된 name 으로 교체
    strategy:
      matrix:
        language: [go]  # T2.0 결과에 따라 추가/제거
    steps:
      - run: echo "Docs-only PR — CodeQL skipped"
```

**Verification**:

- `yamllint .github/workflows/codeql.yml` 통과.
- 빈 docs-only PR 시뮬레이션 (Wave 0 의 PoC 와 유사):
  - `gh pr checks <PR>` 에서 CodeQL required check 가 SUCCESS (skip-marker 경유).
  - 실제 `analyze` job 은 skipped 상태.
- 빈 Go PR 에서 CodeQL `analyze` 정상 실행.
- Branch protection 의 `CodeQL` required check 가 skip-marker pass 로 satisfy 되는지
  `gh pr view <PR> --json statusCheckRollup` 으로 확인.

**Implements**: REQ-CIFT-002a, REQ-CIFT-002b.

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

- `find .github/workflows -name '*.yml' | wc -l` 이전 20 → 15 (T6 nightly 추가 전 시점).
  T6 nightly 추가 후 최종 16. Net delta = -4 (delete 5 + add 1).
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

Wave 0 — Skip-Marker PoC (sandbox, NOT merged):
  branch: test/skip-marker-poc (throw-away)
  files:
    - NEW (sandbox): .github/workflows/skip-marker-poc.yml
    - 1-line markdown edit (e.g., README.md)
  task order: T0
  gates: T0 verification PASS (skip-marker job emits SUCCESS for docs-only PR)
  outcome: PR opened → checks observed → PR closed + branch deleted
  side-effect: design.md AD-002 가 PoC 결과 (PASS / observed name) 로 update.
               Wave 1 run-PR 본문에 sandbox PR link 인용.

Wave 1 — Run-PR (main implementation):
  pre-condition: Wave 0 PASS recorded in AD-002
  branch: feat/SPEC-V3R4-CI-FASTTRACK-001-fasttrack-impl
  files:
    - MODIFY: .github/workflows/ci.yml (T1, 5-job behavior table per plan.md)
    - MODIFY: .github/workflows/codeql.yml (T2, skip-marker pattern mirror)
    - DELETE: .github/workflows/{codex-review,gemini-review,glm-review,llm-panel,claude-code-review.optional}.yml (T3)
    - AUDIT: .github/workflows/{claude,review-quality-gate}.yml (T4, no-edit)
    - NEW: lefthook.yml + MODIFY: Makefile (T5)
    - NEW: .github/workflows/nightly-full-matrix.yml (T6)
    - MODIFY: CLAUDE.local.md §18.7 (T7)
    - MODIFY: ~/.claude/projects/.../memory/lessons.md (T8)
  task order: T4 (audit) → T1 → T2 → T3 → T5 → T6 → T7 → T8
  gates: AC-CIFT-001..009 + PG-1..2 + EC-1..4 ALL PASS
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
