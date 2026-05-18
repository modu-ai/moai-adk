# SPEC-V3R4-CI-INFRA-FIX-001 — Implementation Plan

> **Plan owner**: manager-spec (plan-phase) → manager-develop (run-phase, cycle_type=ddd) → manager-docs (sync-phase)
> **Strategy**: plan-in-main + run-in-main + sync-in-main (NO worktree per `feedback_worktree_never_use` + CI infra single-pane diff 보기 용이)
> **PR lifecycle**: 3 PRs (plan / run / sync) — 모두 `--squash` auto-merge (CLAUDE.local.md §18.3)
> **Run-phase methodology**: DDD (Domain-Driven Development) — CI workflow YAML 변경은 unit-testable TDD path 부재. evidence-driven verification (CI run × N=3) 으로 검증.

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-16 | manager-spec | 초기 draft. 3-Wave evidence-driven plan (W1: SIGPIPE / W2: 403 / W3: fetch-depth + skip-guard 제거). 단일 SPEC scope으로 3 sub-fix bundle. v2.20.0-rc1 release-readiness 최종 precondition. PR lifecycle: plan + run + sync = 3 PRs. |

---

## 1. Implementation Strategy

본 SPEC은 **CI infrastructure metadata-heavy** SPEC 이다. 3개 CI 결함을 동시 fix 하되 각각 독립 검증 가능한 3-Wave 로 분해. TDD/DDD cycle 중 DDD (ANALYZE-PRESERVE-IMPROVE) 모드 채택 — CI YAML 변경은 unit-test 작성이 부적절 (외부 의존 = GitHub Actions runner 환경).

```
Wave 1 (Fix A — SIGPIPE)         Wave 2 (Fix B — 403)         Wave 3 (Fix C — fetch-depth)
─────────────────────────         ──────────────────           ────────────────────────────
detect-language action.yml   →   spec-status-auto-sync   →    ci.yml 5x checkout +
shell pipeline replacement       permissions block 추가         drift_specid_grep_test.go
+ CI run 3x SIGPIPE 부재 확인     + trigger 정책 검증            skip guard 제거
                                  + merged PR test            + AC-CIIF-003 verification
                                  → AC-CIIF-002              → AC-CIIF-003
→ AC-CIIF-001                                                  → AC-CIIF-004 (regression)
```

### 1.1 Wave 1 — Fix A (detect-language SIGPIPE)

**Goal**: detect-language composite action 의 shell pipeline 에서 SIGPIPE (exit 141) 영구 제거.

**Approach**: design.md D-1 에 따라 다음 3 candidate 중 선택:
- (A1) `head -1` → `sed -n '1p'` (sed 는 stdin 전체 read 후 종료)
- (A2) `head -1` → `awk 'NR==1{print;exit}'` (awk explicit exit)
- (A3) `head -1` 유지 + `set +o pipefail` (run-phase 에서 only relevant snippet 에 한정)

**Plan-phase 잠정 결정**: A2 (`awk 'NR==1{print;exit}'`). 이유는 design.md D-1 참조. run-phase 에서 plan-auditor + manager-develop 검토 후 최종 확정.

**File changes**: `.github/actions/detect-language/action.yml` (1 file, 1 line modification 추정).

**Verification**:
1. PR open 후 CI 3 consecutive runs 모두 detect-language step 성공.
2. ubuntu-latest / macos-latest / windows-latest 3 platform 모두 PASS.

### 1.2 Wave 2 — Fix B (spec-status-auto-sync 403)

**Goal**: spec-status-auto-sync workflow 의 `git push origin main` 이 403 없이 성공.

**Approach**: design.md D-2 에 따라 다음 layer 추가:
- (B1) workflow file 최상단에 `permissions:` block 명시 (`contents: write` only)
- (B2) job-level `permissions:` 명시 (optional — workflow-level 로 충분 시 생략)

**Plan-phase 잠정 결정**: B1 (workflow-level `permissions: contents: write`). job 단위 over-scoping 회피.

**File changes**: `.github/workflows/spec-status-auto-sync.yml` (1 file, 3-5 line addition).

**Verification**:
1. workflow_dispatch 또는 dummy merged PR 시뮬레이션 → workflow run 결과 success.
2. workflow run log 에 `403` 부재 grep 확인.

### 1.3 Wave 3 — Fix C (checkout fetch-depth: 0)

**Goal**: ci.yml 5 checkout step 모두 `fetch-depth: 0` 명시 + `drift_specid_grep_test.go` skip guard 제거.

**Approach**:
- (C1) ci.yml 의 5 `actions/checkout@v5` step 각각에 `with: fetch-depth: 0` 추가
- (C2) `internal/spec/drift_specid_grep_test.go:21-34` 의 `GITHUB_ACTIONS` env probe + 보조 local probe 제거 (또는 simplification — design.md D-3 에서 결정).

**Plan-phase 잠정 결정**: C1 + C2 모두 적용. 단, run-phase 에서 W3-T3 (skip guard 제거) 가 PASS 임을 verify 후 commit.

**File changes**:
- `.github/workflows/ci.yml` (5 checkout step + `with: fetch-depth: 0` 추가)
- `internal/spec/drift_specid_grep_test.go` (skip guard 제거 + 주석 정리 + `@MX:NOTE` 갱신)

**Verification**:
1. CI run 후 `TestGetGitImpliedStatus_HARNESS001Resolution` 가 PASS (not SKIP).
2. ci.yml diff 검토: 5 checkout step 모두 fetch-depth: 0 명시.

---

## 2. Milestones (Priority-Ordered, No Time Estimates)

| Priority | Milestone                                                    | Output | Phase |
|----------|--------------------------------------------------------------|--------|-------|
| P1-1     | Wave 1 (Fix A — SIGPIPE) action.yml 수정 + commit            | 1 commit, 1 file | run-phase |
| P1-2     | Wave 2 (Fix B — 403) workflow.yml permissions 추가 + commit   | 1 commit, 1 file | run-phase |
| P1-3     | Wave 3 (Fix C — fetch-depth) ci.yml 5x checkout 수정 + commit | 1 commit, 1 file | run-phase |
| P1-4     | Wave 3 follow-up (skip guard 제거) drift test 정리 + commit    | 1 commit, 1 file | run-phase |
| P2-1     | run-PR open with `--squash` auto-merge                       | PR #N+1 (base main) | run-phase |
| P2-2     | run-PR CI all-green 확인 + AC-CIIF-001~003 PASS               | CI logs | run-phase |
| P2-3     | sync-phase: spec.md frontmatter status `draft → completed`   | 1 commit | sync-phase |
| P2-4     | sync-PR open with `--squash` auto-merge                      | PR #N+2 (base main) | sync-phase |
| P2-5     | sync-PR CI all-green + AC-CIIF-004 regression PASS           | CI logs | sync-phase |

**Sequential execution**: P1-1 → P1-2 → P1-3 → P1-4 → P2-1. P2-1 merged 후 P2-3 → P2-4. P1-1~P1-3 는 mutually independent 하지만 main checkout 에서 sequential commit (commit history 명확성).

---

## 3. Technical Approach (per Wave)

### 3.1 Wave 1 — detect-language SIGPIPE remediation

**Current state** (`.github/actions/detect-language/action.yml:14-17`):

```yaml
LANG_COUNT=$(find . -name "*.go" -o -name "*.py" -o -name "*.ts" -o -name "*.js" | \
  grep -v node_modules | grep -v vendor | grep -v ".git" | \
  head -1 | sed 's/.*\.//' | sort | uniq -c | sort -rn | head -1 | \
  awk '{print $2}')
```

**Plan-phase 잠정 fix**:

```yaml
LANG_COUNT=$(find . -name "*.go" -o -name "*.py" -o -name "*.ts" -o -name "*.js" | \
  grep -v node_modules | grep -v vendor | grep -v ".git" | \
  awk 'NR==1{print;exit}' | sed 's/.*\.//' | sort | uniq -c | sort -rn | \
  awk 'NR==1{print $2;exit}')
```

**Rationale**:
- `head -1` 두 위치 모두 `awk 'NR==1{print;exit}'` / `awk 'NR==1{print $2;exit}'` 로 교체.
- awk 는 첫 줄 출력 후 `exit` 로 명시적 종료 → upstream pipe 와 broken-pipe race 회피.
- shell idiom 보존 (find + grep + awk + sed pipeline 형태 유지).

**Alternative (run-phase 검토)**: D-1 에서 (A1) sed, (A3) pipefail off 의 trade-off 비교.

### 3.2 Wave 2 — spec-status-auto-sync permissions

**Current state** (`.github/workflows/spec-status-auto-sync.yml:1-15`):

```yaml
name: SPEC Status Auto-Sync

on:
  pull_request:
    types: [closed]

jobs:
  spec-status-sync:
    runs-on: ubuntu-latest
    if: github.event.pull_request.merged == true
```

**Plan-phase 잠정 fix** (workflow-level `permissions:` 추가):

```yaml
name: SPEC Status Auto-Sync

on:
  pull_request:
    types: [closed]

permissions:
  contents: write    # git push origin main 필요
  issues: write      # gh issue create fallback (line 95-99) 필요

jobs:
  spec-status-sync:
    runs-on: ubuntu-latest
    if: github.event.pull_request.merged == true
```

**Rationale**:
- `contents: write` — `git push origin main` (line 107)
- `issues: write` — `gh issue create` fallback (line 95-99) 발견됨. minimal-privilege 원칙 + 누락 시 fallback fail.
- `pull-requests:` 등 다른 scope 는 미사용 → 미추가.

**Run-phase 검증 task**: workflow_dispatch 시뮬레이션 (또는 dummy SPEC 변경 PR 머지) → push 성공 + 403 부재 확인.

### 3.3 Wave 3-A — ci.yml fetch-depth: 0

**Current state** — `ci.yml` 의 5개 `actions/checkout@v5` step (line 27-28, 95-96, 119-120, 156-157, 195-196):

```yaml
      - name: Checkout code
        uses: actions/checkout@v5
```

**Plan-phase 잠정 fix** (5 step 모두 동일 패턴):

```yaml
      - name: Checkout code
        uses: actions/checkout@v5
        with:
          fetch-depth: 0
```

**Rationale**:
- 5 step 모두 일관성. constitution-check job 은 git 미사용이지만 일관성 위해 동일 적용.
- design.md D-3 의 "all-or-none" 결정 따라 5 step 모두 적용. partial 적용은 inconsistency 생성.

### 3.4 Wave 3-B — drift_specid_grep_test.go skip guard 제거

**Current state** (`internal/spec/drift_specid_grep_test.go:21-34`):

```go
func TestGetGitImpliedStatus_HARNESS001Resolution(t *testing.T) {
    // CI 환경 자동 skip — shallow clone으로 SPEC commits 부재
    if os.Getenv("GITHUB_ACTIONS") == "true" {
        t.Skip("requires full git history; CI uses actions/checkout@v4 default fetch-depth: 1 (shallow). " +
            "Word-boundary logic is fully covered by TestGetGitImpliedStatus_SPECIDWordBoundary 5 sub-cases. " +
            "Follow-up: SPEC-V3R4-CI-INFRA-FIX-001 to set fetch-depth: 0 for full-history tests.")
    }

    // Probe: target SPEC commits이 local git에 존재하는지 확인 (non-CI 환경에서도 fork/shallow 대응)
    probe := exec.Command("git", "log", "main", "--oneline", "--grep=SPEC-V3R4-HARNESS-001", "-1")
    if out, err := probe.Output(); err != nil || len(out) == 0 {
        t.Skip("SPEC-V3R4-HARNESS-001 commits not available in local git history (fork/shallow clone). " +
            "WordBoundary helper test (5 sub-cases) covers the logic.")
    }
```

**Plan-phase 잠정 fix**:
- `GITHUB_ACTIONS` env probe (line 23-27) 완전 제거.
- local probe (line 30-34) **보존** — fork/shallow clone 사용자 환경 대응 (CI 외).
- 주변 주석에 `@MX:NOTE` 갱신: `SPEC-V3R4-CI-INFRA-FIX-001 적용 — fetch-depth: 0 으로 CI shallow-clone 한계 해소. local probe 만 보존 (fork/shallow 사용자 케이스).`

```go
func TestGetGitImpliedStatus_HARNESS001Resolution(t *testing.T) {
    // @MX:NOTE: [AUTO] CI shallow-clone skip guard 제거 — SPEC-V3R4-CI-INFRA-FIX-001 적용 후
    // ci.yml 5 checkout step 모두 fetch-depth: 0. 본 test 는 CI 에서 정상 실행.
    // @MX:REASON: LSGF-001 PR #948 의 GITHUB_ACTIONS env workaround 영구 제거.

    // Probe: target SPEC commits이 local git에 존재하는지 확인 (fork/shallow clone 사용자 환경 대응)
    probe := exec.Command("git", "log", "main", "--oneline", "--grep=SPEC-V3R4-HARNESS-001", "-1")
    if out, err := probe.Output(); err != nil || len(out) == 0 {
        t.Skip("SPEC-V3R4-HARNESS-001 commits not available in local git history (fork/shallow clone). " +
            "WordBoundary helper test (5 sub-cases) covers the logic.")
    }
    // ... 기존 본문 유지
}
```

**Test verification**:
- 로컬 (full clone): test 정상 실행 + PASS.
- CI (post-fix): test 정상 실행 + PASS (이전엔 skip).

### 3.5 Wave 3-C — Verification

```bash
# AC-CIIF-001 verification
gh run list --workflow=ci.yml --limit=3 --json conclusion,databaseId,event \
  | jq '[.[] | select(.conclusion == "failure")] | length'
# expected: 0

# AC-CIIF-002 verification (manual or dummy PR)
gh run view <spec-status-auto-sync-run-id> --log | grep -c "403"
# expected: 0

# AC-CIIF-003 verification (after PR merge)
gh run view <ci-run-id> --log | grep "TestGetGitImpliedStatus_HARNESS001Resolution"
# expected: "--- PASS" (not "--- SKIP")

# AC-CIIF-004 verification (regression baseline)
moai spec lint --strict 2>&1 | tail -3
# expected: "0 error(s), 0 warning(s)"
```

---

## 4. Risks and Mitigation

### Risk-1: awk replacement 가 detection semantics 변경

**상황**: `head -1` → `awk 'NR==1{print;exit}'` 가 첫 줄 추출 의도와 동일한지 의문.

**Mitigation**:
- awk `NR==1{print;exit}` 는 정확히 첫 줄 (`NR==1`) 출력 후 종료 (`exit`). semantics 동일.
- run-phase W1-T2 (fixture 검증) 에서 16 supported languages 각각 fixture (`*.go`, `*.py` etc.) 로 LANG_COUNT 값 비교.

### Risk-2: permissions 변경이 다른 implicit scope 사용을 break

**상황**: spec-status-auto-sync.yml 의 step (gh issue create 등) 이 implicit scope 에 의존.

**Mitigation**:
- 본 plan §3.2 에서 명시한 `contents: write` + `issues: write` 가 모든 step requirement 를 cover 한다고 audit.
- run-phase W2-T2 에서 workflow run log 검토 시 추가 권한 필요한 step 발견되면 plan §3.2 갱신.

### Risk-3: fetch-depth: 0 가 CI cache miss 유발

**상황**: shallow clone 에서 full clone 으로 변경 시 actions/cache hit-rate 감소.

**Mitigation**:
- ci.yml 의 cache key 는 `go-version` + `go.sum` 기반 (`actions/setup-go@v6` cache: true). clone depth 와 무관.
- 우려 무근. run-phase 에서 첫 CI run 의 cache hit 확인.

### Risk-4: drift_specid_grep_test.go 가 fork/PR-from-fork 에서 fail

**상황**: external PR fork 환경에서 `SPEC-V3R4-HARNESS-001` commits 부재 시 test 실패 (skip 아님).

**Mitigation**:
- local probe (line 30-34) 보존. fork/shallow 사용자 환경에서는 여전히 skip.
- main repo CI 에서는 fetch-depth: 0 + main branch commits 정상 존재 → PASS.

### Risk-5: 본 SPEC plan-PR 머지 후 fetch-depth: 0 가 본 PR run-phase CI 에서 즉시 적용 안 됨

**상황**: plan-PR 머지 (squash) 후 run-PR base 는 새 main. fetch-depth: 0 변경은 run-PR commits 에 포함 → run-PR 자체 CI 에서 적용.

**Mitigation**:
- run-PR commits 에 ci.yml fetch-depth 변경 포함 → run-PR 의 CI run 부터 적용 (GitHub Actions workflow는 PR head 의 .github/workflows/* 를 사용).
- 단, 머지 후 main 에 적용되어 후속 PR 부터 effective. 본 SPEC run-PR 의 CI run 은 self-referential 검증.

### Risk-6: skip guard 제거 commit 이 fetch-depth 추가 commit 보다 먼저 머지 시 CI fail

**상황**: run-PR 내부에서 commit 순서가 (1) skip guard 제거 → (2) fetch-depth 추가 인 경우, (1) 단독 시점에는 CI 가 fetch-depth: 1 환경 + skip guard 없음 → test fail.

**Mitigation**:
- run-PR 의 commit sequence: P1-1 (Fix A) → P1-2 (Fix B) → P1-3 (Fix C ci.yml) → P1-4 (Fix C skip guard 제거).
- squash merge 이므로 main 머지 시점에는 단일 commit (sequence 무관).
- 단, PR 자체의 CI 는 각 commit 별로 trigger 되므로 P1-4 단독 commit 의 CI run 은 P1-3 도 포함된 상태 (cumulative) → 안전. design.md D-3 참조.

---

## 5. Rollback Plan

### 5.1 Wave 1-3 rollback (run-phase 중)

만약 Wave N 적용 후 CI 결과 악화 시:

```bash
git reset --hard HEAD~1  # 마지막 Wave commit revert
gh run rerun <previous-failing-run-id>
```

run-PR 미머지 상태 → main 영향 0건.

### 5.2 Sync-phase rollback (sync-PR open 후 lint regression 노출 시)

```bash
gh pr close <sync-PR-number>
gh pr close <run-PR-number>
# run-phase commits 모두 main 미머지 상태 → revert 불필요
```

### 5.3 Post-merge rollback (main 머지 후 회귀 발견)

```bash
# 새 revert PR
git revert <merge-commit-sha> --no-commit
# 추가 fix 후 새 PR 발급
```

run/sync PR 모두 `--squash` → 단일 commit revert 로 깔끔 되돌리기.

### 5.4 Selective rollback (3 sub-fix 중 1건만 회귀 시)

**상황**: e.g., Fix A 가 unexpected detection 회귀 유발 → Fix A 만 rollback, Fix B/C 유지.

**Approach**:
- run-PR squash merge 후 단일 commit 이므로 selective revert 불가능 → 새 fix-up PR 로 Fix A 만 정정.
- Plan-auditor 참고: 본 risk 대응 위해 run-PR 내부 commit 분할 보존 (squash 전) 권장. 단, branch protection (CLAUDE.local.md §18.7) 와 squash 정책 (§18.3) 충돌 → squash 우선.

---

## 6. Telemetry & Observability

### 6.1 Pre-fix baseline (main `9be0ef03b`)

```bash
moai spec lint --strict 2>&1 | tail -1
# expected: 0 error(s), 0 warning(s)
```

NAMESPACE-001 (PR #944) baseline data:
- detect-language SIGPIPE: 4건 incident (project_v3r4_namespace_001_plan_merged.md memory)
- spec-status-auto-sync 403: 1건 incident (force-push 후)
- drift test skip on CI: 1건 (PR #948)

### 6.2 Post-fix target

```bash
# CI 3 consecutive runs 모두 success
gh run list --workflow=ci.yml --limit=3 --json conclusion \
  | jq -r '.[].conclusion' \
  | sort -u
# expected: "success"

# spec-status-auto-sync run success
gh run list --workflow=spec-status-auto-sync.yml --limit=1 --json conclusion \
  | jq -r '.[0].conclusion'
# expected: "success"

# drift test 정상 실행
gh run view <ci-run-id> --log \
  | grep "TestGetGitImpliedStatus_HARNESS001Resolution" \
  | grep -E "PASS|SKIP"
# expected: "--- PASS"
```

### 6.3 Run-phase progress.md 형식

run-phase에서 `.moai/specs/SPEC-V3R4-CI-INFRA-FIX-001/progress.md` 생성:

```markdown
# Progress — SPEC-V3R4-CI-INFRA-FIX-001

## Wave 1 (Fix A — SIGPIPE) — <timestamp>
- action.yml diff: <details>
- commit: <sha>
- CI run (post-commit): <run-id>, conclusion: success/failure

## Wave 2 (Fix B — 403) — <timestamp>
- workflow.yml diff: <details>
- commit: <sha>
- workflow run (dummy/dispatch): <run-id>, conclusion: success/failure, 403 grep: 0

## Wave 3 (Fix C — fetch-depth + skip guard) — <timestamp>
- ci.yml diff: 5 checkout step fetch-depth: 0
- drift_specid_grep_test.go diff: skip guard 제거
- commit: <sha-ci>, <sha-skip>
- CI run (post-commit): <run-id>, drift test result: PASS

## Verification (post run-PR open) — <timestamp>
- AC-CIIF-001: PASS/FAIL
- AC-CIIF-002: PASS/FAIL
- AC-CIIF-003: PASS/FAIL
- AC-CIIF-004: PASS/FAIL (lint regression baseline)

## PR Status
- run-PR #<N+1>: <status>
- sync-PR #<N+2>: <status>
```

---

## 7. Open Questions (OQ block)

### OQ-1: detect-language action SIGPIPE 솔루션 — awk vs sed vs pipefail-off

**Question**: A1 (sed -n '1p'), A2 (awk NR==1exit), A3 (set +o pipefail) 중 어느 것이 가장 robust 한가?

**Status**: plan-phase 잠정 A2 채택. run-phase 에서 plan-auditor 검토 + 16 language fixture 테스트로 확정.

**Resolution Path**:
- A1 (sed): `sed -n '1p'` 는 sed 가 stdin 전체 read 후 종료. SIGPIPE 안전. 그러나 sed BRE regex syntax 와 첫 줄 추출이 idiomatic 한지 검토 필요.
- A2 (awk): `awk 'NR==1{print;exit}'`. 명시적 exit. 가장 portable + clear.
- A3 (pipefail-off): `set +o pipefail` 로 SIGPIPE 무시. 그러나 step 전체에서 pipefail off 는 다른 errror 도 mask 가능.

**Plan-phase 권장**: A2 (awk). run-phase 에서 sed 와 비교 후 최종 결정.

### OQ-2: spec-status-auto-sync 의 trigger 정책 검토

**Question**: 현재 `on: pull_request: types: [closed]` trigger 는 fork-from-external PR 에도 발동. fork PR 은 GITHUB_TOKEN scope 제한 (read-only) → 403 의 또 다른 원인.

**Status**: plan-phase 미해결. run-phase 검증.

**Resolution Path**:
- Option (a): trigger 를 `pull_request_target` 로 변경 — fork PR 도 base repo 권한으로 실행. 단, security risk (untrusted code 가 main repo token 사용) — 신중 검토 필요.
- Option (b): trigger 유지 + `if: github.event.pull_request.head.repo.full_name == github.repository` 추가 — fork PR 은 skip.
- Option (c): permissions block 만 추가하고 fork PR 검토는 별도 follow-up.

**Plan-phase 권장**: Option (c). 본 SPEC scope 외 fork PR security model 변경 회피.

### OQ-3: ci.yml 외 다른 workflow 의 fetch-depth 검토 scope

**Question**: codeql.yml, test-install.yml, auto-merge.yml 등 다른 workflow 도 fetch-depth: 0 필요한가?

**Status**: 미해결.

**Resolution Path**:
- design.md D-3 결정: 본 SPEC은 ci.yml 만 (REQ-CIIF-003 명시). 다른 workflow 는 git log 사용 없으므로 fetch-depth: 1 적정.
- 단, 후속 audit SPEC `SPEC-V3R4-CI-WORKFLOWS-AUDIT-001` 후보 (out of scope).

### OQ-4: skip guard 제거 후 local non-CI 환경에서 회귀 위험

**Question**: skip guard 제거 시 local 개발자가 `go test ./internal/spec/... -race` 실행 시 fork/shallow clone 환경에서 fail.

**Status**: 해결됨.

**Resolution**: drift_specid_grep_test.go 의 local probe (line 30-34) 보존. fork/shallow 사용자 환경에서는 여전히 skip. CI 환경 (GitHub Actions, fetch-depth: 0) 에서만 실행.

### OQ-5: detect-language action 의 windows-latest 호환성

**Question**: action.yml 의 shell 은 bash. windows-latest 에서 git bash 사용. find/grep/awk/sed POSIX command 호환?

**Status**: plan-phase 미검증.

**Resolution Path**: run-phase W1-T3 에서 ci.yml 의 windows-latest runner CI run 결과 검토. 현재까지 detect-language action 은 ci.yml 에서 직접 호출되지 않음 (gemini-review.yml, glm-review.yml 등에서 사용) → windows 영향 0 일 가능성. 단, 안전상 검증 task 포함.

### OQ-6: spec-status-auto-sync 의 branch protection 우회 메커니즘

**Question**: CLAUDE.local.md §18.7 에 따라 main 은 branch protection 보호 (`enforce_admins: false`, `required_approving_review_count: 1`, `allow_force_pushes: false`). bot 의 `git push origin main` 직접 push 가 가능?

**Status**: plan-phase 미해결.

**Resolution Path**: GitHub Actions 의 `github-actions[bot]` 는 branch protection 의 "Restrict who can push to matching branches" 가 비활성화된 경우 push 가능. 현재 CLAUDE.local.md §18.7 의 protection rule 은 `restrictions: null` → bot push 허용. permissions: contents: write 만 추가하면 충분 가능.

**Plan-phase 권장**: run-phase W2-T3 에서 dummy push 시뮬레이션 으로 검증.

---

## 8. References (cross-file)

- `spec.md` — REQ ↔ AC mapping + 3 sub-fix root causes
- `acceptance.md` — Given-When-Then scenarios + edge cases + DoD
- `design.md` — D-1/D-2/D-3/D-4 architectural decisions (sub-fix solutions)
- `tasks.md` — Wave 1/2/3 task breakdown (10 tasks)
- `.moai/specs/SPEC-V3R4-LINT-SPECID-GREP-FIX-001/spec.md` — walker word-boundary fix (직접 trigger)
- `.moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002/plan.md` — 3-Wave evidence-driven plan precedent
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — 12-field canonical schema SSOT
- `CLAUDE.local.md §18.3` — merge strategy: feature/* → squash, release/* → merge commit
- `CLAUDE.local.md §18.7` — branch protection rule (read-only reference, 본 SPEC 미변경)
- `CLAUDE.local.md §18.8` — release process (post-merge `./scripts/release.sh v2.20.0-rc1`)
- `internal/spec/drift.go:120-180` — walker (참조 only, 무수정)
- `internal/spec/drift_specid_grep_test.go:21-34` — skip guard (Wave 3-B 제거 대상)

---

## 9. Merge Strategy

본 SPEC은 3 PRs (plan / run / sync) 모두 **`--squash` auto-merge** (CLAUDE.local.md §18.3 — feature/fix/chore/docs → main).

- plan-PR: `feat/SPEC-V3R4-CI-INFRA-FIX-001-plan` → main, `--squash --auto`
- run-PR: `feat/SPEC-V3R4-CI-INFRA-FIX-001` (plan 머지 후 fresh branch) → main, `--squash --auto`
- sync-PR: `sync/SPEC-V3R4-CI-INFRA-FIX-001` (run 머지 후 fresh branch) → main, `--squash --auto`

각 PR 머지 후 `--delete-branch=true` (CLAUDE.local.md §18.3 공통 규칙). 모든 단계 plan-auditor + branch protection required check 6/6 PASS 후 머지.
