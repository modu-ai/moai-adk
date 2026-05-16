# SPEC-V3R4-CI-INFRA-FIX-001 — Design Rationale

> **Scope**: 본 design 문서는 3개 CI infrastructure sub-fix 의 **architectural decisions** 와 **trade-off analysis** 만 다룬다. 본 SPEC 은 CI workflow YAML + composite action shell + Go test skip guard 의 3-layer 변경이며, design.md 는 각 layer 의 선택 근거를 명시한다.

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-16 | manager-spec | 초기 draft. D-1 (SIGPIPE remediation) + D-2 (permissions scope) + D-3 (fetch-depth: 0 scope) + D-4 (backward compatibility) 4개 architectural decision. run-phase 적용 전 plan-auditor 검토 대상. |

---

## 1. Why This SPEC Exists

### 1.1 v3.0.0-rc1 Release Gate 의 Final Precondition

`SPEC-V3R4-LINT-SPECID-GREP-FIX-001` (LSGF-001) + `SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002` (FOLLOWUP-002) 두 SPEC closure 후 main HEAD `9be0ef03b` 에서 `moai spec lint --strict` 가 `✓ No findings (0/0)` 를 달성했다. 이는 lint correctness 의 release readiness 확보.

그러나 CI infrastructure 자체의 신뢰성은 별개:
- detect-language action 의 SIGPIPE 4건 incident (NAMESPACE-001 PR #944)
- spec-status-auto-sync workflow 의 403 push failure (NAMESPACE-001 force-push 후)
- LSGF-001 PR #948 의 `GITHUB_ACTIONS` env workaround (drift_specid_grep_test.go skip guard)

본 SPEC 은 이 3 결함을 **단일 SPEC scope** 으로 묶어 영구 해소. v3.0.0-rc1 태그 발행의 마지막 precondition 충족.

### 1.2 왜 1개 SPEC 으로 묶는가 (vs 3 SPEC 분리)

**3 SPEC 분리 (가설)**:
- SPEC-CIIF-A: SIGPIPE 만
- SPEC-CIIF-B: 403 만
- SPEC-CIIF-C: fetch-depth 만

**Pros of 분리**: per-defect 독립 verification, plan-auditor 부하 분산.

**Cons of 분리**: 3 plan-PR + 3 run-PR + 3 sync-PR = 9 PRs. 각각 lint baseline 회귀 위험 + branch base 관리 부담. 3 SPEC 모두 v3.0.0-rc1 precondition 이므로 release tag 까지 9 PR 머지 대기.

**1 SPEC 통합**: 3 plan + 3 run-Wave + 3 sync = 3 PRs. 단일 lint baseline 회귀 검증. plan-auditor 단일 review.

**Decision**: 1 SPEC 통합. cohesion 높음 (모두 CI infrastructure layer), 의존성 적음 (sub-fix 간 mutually independent), v3.0.0-rc1 gate 동일 목적.

### 1.3 LSGF-001 / FOLLOWUP-002 와의 lineage

```
LSGF-001 (walker fix)          →  17 drift exposed  →  FOLLOWUP-002 (17→0 binary)
                               ↓
                               LSGF-001 PR #948 의 GITHUB_ACTIONS skip workaround
                               ↓
                               CI-INFRA-FIX-001 (본 SPEC) — workaround 영구 제거
                               ↓
                               v3.0.0-rc1 release gate 충족
```

본 SPEC 은 LSGF-001 의 직접 후속 (workaround 제거) + FOLLOWUP-002 와 병렬 release-gate sibling.

---

## 2. Decision D-1: SIGPIPE Remediation Strategy

### 2.1 Context

`.github/actions/detect-language/action.yml:14-17` 의 shell pipeline 이 SIGPIPE (exit 141) 를 generate. 이는 `head -1` 이 첫 줄 후 stdin 을 close 하지만 upstream `find` 가 계속 write 하면서 broken pipe signal 받는 race condition.

`set -o pipefail` (GitHub Actions bash default) 가 활성화된 경우 SIGPIPE → exit 141 → step 실패. NAMESPACE-001 PR #944 에서 4건 incident 발생.

### 2.2 Candidate Solutions

#### A1 — sed 첫 줄 추출

```yaml
LANG_COUNT=$(find ... | sed -n '1p' | sed 's/.*\.//' | sort | uniq -c | sort -rn | sed -n '1p' | awk '{print $2}')
```

**Pros**:
- sed 는 stdin 전체 read 후 종료 (SIGPIPE 없음).
- POSIX standard.

**Cons**:
- 두 `head -1` 위치 모두 sed 로 교체 필요.
- `sed -n '1p'` 는 첫 줄 출력하지만 stdin 끝까지 읽어 input pipe 가 길면 latency 증가 (현 case 무시 가능).
- find 가 large tree 에서 모든 결과를 generate 후에야 sed 가 출력 → 의미는 동일하지만 성능 약간 다름.

#### A2 — awk NR==1 + exit

```yaml
LANG_COUNT=$(find ... | awk 'NR==1{print;exit}' | sed 's/.*\.//' | sort | uniq -c | sort -rn | awk 'NR==1{print $2;exit}')
```

**Pros**:
- awk 명시적 `exit` → 첫 매칭 후 즉시 종료. upstream stdin close 와 동기화.
- 명시적 의도 표현 (`NR==1{print;exit}`).
- POSIX awk standard.

**Cons**:
- 의미는 head -1 과 동일하지만 awk 도 explicit exit 시점에 upstream 이 still writing 이면 broken pipe 가능 — 단, awk 가 head 보다 stdin 처리 buffer 가 다름 (POSIX line buffering) → broken pipe race 차이는 본 case 에서 충분히 작음.
- 두 위치 모두 변경 필요.

#### A3 — set +o pipefail 비활성화

```yaml
- id: detect
  shell: bash
  run: |
    set +o pipefail   # SIGPIPE 가 step 실패시키지 않도록
    LANG_COUNT=$(find ... | head -1 | sed ... | head -1 | awk ...)
```

**Pros**:
- 변경 최소 (한 줄 추가).
- pipeline 의도 보존.

**Cons**:
- pipefail 비활성화는 *모든* pipeline error 를 mask. e.g., find 가 permission denied 로 fail 해도 step 성공.
- explicit error-handling 손실. CI 신뢰성 측면에서 회귀.
- best practice 위배.

### 2.3 Plan-phase Decision: A2 (awk)

**선택 이유**:
1. **명시적 의도**: `NR==1{print;exit}` 는 "첫 줄 처리 후 종료" 를 코드 자체로 표현. 다음 reader 에게 명확.
2. **SIGPIPE safety**: awk explicit exit 가 head implicit close 보다 race window 좁음.
3. **Pipefail preservation**: A3 와 달리 다른 error 는 여전히 detect.
4. **POSIX compliance**: ubuntu/macos/windows-bash 3 platform 모두 동일 동작.

**Trade-off acknowledged**: A1 (sed) 도 valid. 차이는 cosmetic (`'1p'` vs `'NR==1{...}'`). run-phase 에서 manager-develop + plan-auditor 가 두 candidate 코드 diff 비교 후 최종 선택. A1 도 acceptable.

**Rejected**: A3. pipefail-off 는 안티패턴.

### 2.4 Verification Strategy (D-1)

- run-phase W1-T2: 16 supported languages 각각 fixture 디렉토리 생성 (e.g., `tmp/fixture-go/main.go`, `tmp/fixture-python/main.py` 등). action 실행 후 `LANG_COUNT` 출력값이 기대 언어 코드와 일치 확인.
- run-phase W1-T3: ci.yml CI 3 consecutive runs 모두 detect-language step 성공 확인 (AC-CIIF-001).

---

## 3. Decision D-2: spec-status-auto-sync Permission Scope

### 3.1 Context

`.github/workflows/spec-status-auto-sync.yml` 은 `pull_request: types: [closed]` trigger 로 활성화. PR 머지 후:
1. PR title 에서 SPEC-ID 추출
2. SPEC frontmatter status 업데이트
3. `git push origin main`

현재 file 에 `permissions:` block 부재 → GITHUB_TOKEN 의 workflow default scope 사용 (`contents: read`). `git push origin main` 은 403 으로 실패.

추가로 line 95-99 에 `gh issue create` fallback step 존재 — 이도 추가 permission 필요.

### 3.2 Candidate Solutions

#### B1 — workflow-level minimum permissions

```yaml
on:
  pull_request:
    types: [closed]

permissions:
  contents: write    # git push origin main
  issues: write      # gh issue create fallback
```

**Pros**:
- minimum scope (least-privilege).
- workflow-level 적용 → 모든 job 에 일관성.
- 2 step 모두 cover.

**Cons**:
- workflow 전체에 `contents: write` 부여. 본 workflow 의 다른 step 도 잠재적 write 가능 (단, 현재 step list 검토 시 위험 없음).

#### B2 — job-level permissions (granular)

```yaml
jobs:
  spec-status-sync:
    runs-on: ubuntu-latest
    if: github.event.pull_request.merged == true
    permissions:
      contents: write
      issues: write
    steps: ...
```

**Pros**:
- job-level scope. workflow 에 다른 job 추가 시 영향 격리.

**Cons**:
- 본 workflow 는 single job → workflow-level 과 동일.
- 추후 job 추가 시점에 reconsider 필요 — 본 SPEC scope 외.

#### B3 — pull_request_target trigger (security risk)

```yaml
on:
  pull_request_target:
    types: [closed]

permissions:
  contents: write
  issues: write
```

**Pros**:
- fork PR 도 base repo permissions 으로 실행 → fork 머지 후 status sync 가능.

**Cons**:
- **Security risk**: untrusted code (fork PR) 가 base repo GITHUB_TOKEN 으로 실행 가능 → repo compromise 위험.
- 본 SPEC scope 외 — fork PR security model 변경 필요.

### 3.3 Plan-phase Decision: B1 (workflow-level minimum permissions)

**선택 이유**:
1. **Least-privilege**: `contents: write` + `issues: write` 만. 다른 scope (actions/id-token/pull-requests) 추가 거부.
2. **Single job**: B1 = B2 effective. workflow-level 이 redundant 명시 회피.
3. **Internal PR only**: fork PR 은 별도 problem (OQ-2 → out of scope). 본 workflow 는 internal PR 머지 시점만 다룸.

**Rejected**: B3 (pull_request_target). security risk 우선.

### 3.4 Trigger 정책 검토 (out of scope, deferred)

현재 trigger `on: pull_request: types: [closed]` 는 fork PR 도 발동. fork PR 은 read-only GITHUB_TOKEN → permissions 추가해도 여전히 403.

**Plan-phase deferral**: fork PR 시나리오는 본 SPEC scope 외. 별도 follow-up SPEC `SPEC-V3R4-CI-FORK-PR-POLICY-001` 후보.

**Current state mitigation**: workflow 의 첫 step 이 already `if: github.event.pull_request.merged == true` 로 머지된 PR 만 처리. fork PR 도 가끔 머지되므로 (admin override) 403 가능성 잔존. 본 SPEC 은 internal PR 머지 (대부분 케이스) 만 fix.

### 3.5 Verification Strategy (D-2)

- run-phase W2-T2: workflow_dispatch 시뮬레이션 (또는 dummy SPEC 변경 PR 머지) 후 workflow run 결과 success.
- run-phase W2-T3: workflow run log grep `403` → 0 매칭 (AC-CIIF-002).

---

## 4. Decision D-3: checkout fetch-depth: 0 Scope

### 4.1 Context

`actions/checkout@v4` / `@v5` 의 default `fetch-depth: 1` (shallow clone) 은 SPEC commits 가 부재한 환경 생성. `internal/spec/drift.go` 의 walker 는 `git log --grep=<SPEC-ID>` 로 full history 필요.

LSGF-001 PR #948 의 workaround: `GITHUB_ACTIONS=true` env 일 때 `TestGetGitImpliedStatus_HARNESS001Resolution` skip. 본 SPEC 은 이 workaround 영구 제거.

### 4.2 Candidate Solutions: Workflow Scope

#### C-Scope-1 — ci.yml only (5 checkout step)

ci.yml 의 5 checkout step (test, test-integration, lint, build, constitution-check) 모두 fetch-depth: 0.

**Pros**:
- 본 SPEC 범위 명확. REQ-CIIF-003 정확 일치.
- drift walker 가 실제 사용되는 workflow (test job) 만 변경.
- 후속 audit 가능.

**Cons**:
- 다른 workflow (codeql.yml 등) 의 잠재적 추후 회귀 위험. 단, 현재 git log 사용 0건.

#### C-Scope-2 — all-workflows comprehensive

ci.yml + codeql.yml + test-install.yml + auto-merge.yml 등 모든 actions/checkout step.

**Pros**:
- Consistency. 미래 추가 workflow 도 패턴 표준화.

**Cons**:
- 본 SPEC scope creep (REQ-CIIF-003 명시 ci.yml 만).
- 각 workflow 의 fetch-depth 필요성 audit 필요 → run-phase 부담 증가.
- 후속 SPEC `SPEC-V3R4-CI-WORKFLOWS-AUDIT-001` 으로 분리하는 것이 적절.

#### C-Scope-3 — drift-test-affected only

ci.yml 의 test job (drift walker test 가 실행되는 step) 만.

**Pros**:
- minimum change. 다른 step 의 cache hit 영향 0.

**Cons**:
- ci.yml 내 일관성 위배 (4 step 만 shallow, 1 step 만 full).
- test-integration job 도 향후 drift test 추가 가능 → preemptive coverage 우선.

### 4.3 Plan-phase Decision: C-Scope-1 (ci.yml all 5 steps)

**선택 이유**:
1. **REQ alignment**: REQ-CIIF-003 명시 "All `actions/checkout` steps in `.github/workflows/ci.yml`".
2. **Internal consistency**: ci.yml 5 step 모두 동일 패턴 → 가독성/유지보수성.
3. **Preemptive coverage**: 미래 ci.yml 내 새 test job 도 drift walker 사용 가능 → 모든 step 일관.
4. **Performance impact 무시 가능**: moai-adk-go ~1000 commits → full clone +2-3s. cache key 는 fetch-depth 무관 → cache miss 없음.

**Rejected**: C-Scope-2 (다른 workflow 포함) — scope creep + audit 부담. follow-up SPEC 후보.
**Rejected**: C-Scope-3 (drift-only) — consistency 위배.

### 4.4 Candidate Solutions: Skip Guard 처리

#### Skip-A — 완전 제거

```go
func TestGetGitImpliedStatus_HARNESS001Resolution(t *testing.T) {
    // skip guard 모두 제거. CI 환경에서 정상 실행 가정.
    status, err := getGitImpliedStatus("SPEC-V3R4-HARNESS-001")
    ...
}
```

**Pros**: 가장 깨끗.

**Cons**: fork/shallow clone 사용자 환경 (PR-from-fork) 에서 fail. test 가 robust 하지 않음.

#### Skip-B — GITHUB_ACTIONS env probe 만 제거, local probe 보존

```go
func TestGetGitImpliedStatus_HARNESS001Resolution(t *testing.T) {
    // CI env probe 제거 (fetch-depth: 0 적용 후 불필요).
    // local probe 보존 (fork/shallow clone 사용자 케이스).
    probe := exec.Command("git", "log", "main", "--oneline", "--grep=SPEC-V3R4-HARNESS-001", "-1")
    if out, err := probe.Output(); err != nil || len(out) == 0 {
        t.Skip("SPEC-V3R4-HARNESS-001 commits not available in local git history (fork/shallow clone).")
    }
    ...
}
```

**Pros**:
- CI 정상 실행 (fetch-depth: 0 → main commits 존재 → probe PASS).
- fork/shallow 사용자 환경 보호.
- robust.

**Cons**: 약간의 코드 잔존.

#### Skip-C — env probe 제거 + local probe 강화

local probe 를 추가 검증 (e.g., commit count 비교) 으로 강화.

**Pros**: 더 robust.
**Cons**: scope creep. 본 SPEC 본질과 무관.

### 4.5 Plan-phase Decision: Skip-B (env probe 제거, local probe 보존)

**선택 이유**:
1. **CI primary use case**: CI 환경에서 fetch-depth: 0 → main commits 존재 → probe PASS → test 정상 실행. env probe 불필요.
2. **Fork/shallow safety**: local probe 가 fork PR 또는 shallow clone 사용자 환경에서 graceful skip 보장.
3. **Minimum change**: env probe 만 제거. local probe 코드 보존 + 주석 갱신.

**Rejected**: Skip-A (모두 제거) — fork PR 안전성 손실.
**Rejected**: Skip-C (강화) — scope creep.

### 4.6 Verification Strategy (D-3)

- run-phase W3-T2: ci.yml 5 step 모두 `with: fetch-depth: 0` grep 확인.
- run-phase W3-T4: CI run 후 `gh run view <id> --log | grep TestGetGitImpliedStatus_HARNESS001Resolution` 결과 `--- PASS` (not `--- SKIP`).
- run-phase W3-T5: local 환경에서 `go test ./internal/spec/... -run TestGetGitImpliedStatus_HARNESS001Resolution -v` PASS.

---

## 5. Decision D-4: Backward Compatibility

### 5.1 Context

본 SPEC 은 CI infrastructure 전용. 사용자 코드 영향 0. 하지만 다음 측면 검토:

### 5.2 Compatibility Analysis

#### Fix A (SIGPIPE)
- **External impact**: 0건. detect-language action 의 output (`language=<lang>`) semantics 보존.
- **Other workflows depending on detect-language**: gemini-review.yml, glm-review.yml. 모두 output 형식만 사용 → 영향 0.
- **Runner compatibility**: ubuntu/macos/windows-bash 모두 awk 표준 → 호환.

#### Fix B (permissions)
- **External impact**: 0건. permissions 명시화는 implicit scope 보다 narrow 또는 동일.
- **Internal impact**: workflow 의 step 들이 사용하는 scope 모두 명시적 cover (`contents: write` + `issues: write`).
- **Breaking change**: 없음. branch protection rule (CLAUDE.local.md §18.7) 와 동일 권한 모델.

#### Fix C (fetch-depth)
- **CI runtime**: +2-3s per checkout step. 5 step × 3 platform = 15 step → +30-45s aggregate.
- **CI cache**: actions/cache key 는 go.sum + go-version 기반 → fetch-depth 영향 0. cache hit-rate 유지.
- **Storage**: GitHub-hosted runner 에서 free. self-hosted 일 경우 storage 영향 (현재 N/A).
- **Other workflows unchanged**: codeql.yml, test-install.yml 등은 fetch-depth: 1 유지 (D-3 결정).

#### Skip guard 제거
- **CI behavior**: post-fix 에서 test 정상 실행 (skip 아님). regression 없음.
- **Local user (full clone)**: 변화 없음. test 정상 실행.
- **Local user (fork/shallow clone)**: local probe 가 graceful skip 처리. 변화 없음.

### 5.3 Plan-phase Decision: All Changes are Non-Breaking

**Conclusion**:
- 본 SPEC 은 **non-breaking** infrastructure change.
- `breaking: false` (spec.md frontmatter).
- `bc_id: []` (no backward-compatibility tracking ID).
- CHANGELOG entry 분류: `### Fixed` (CI infra defect remediation).

### 5.4 SemVer Bump

본 SPEC 자체는 SemVer 영향 없음 (CI infra only, no API/CLI change). 단, 본 SPEC closure 가 v3.0.0-rc1 release tag 의 precondition.

---

## 6. Cross-cutting Design Decisions

### 6.1 Sub-fix Independence

3 sub-fix 모두 mutually independent:
- Fix A 적용 ≠ Fix B 회귀 (별 파일).
- Fix B 적용 ≠ Fix C 회귀 (별 파일).
- Fix C 적용 ≠ Fix A 회귀 (별 파일).

따라서 Wave 1/2/3 sequential 적용 가능 + selective rollback 시 영향 격리.

### 6.2 PR Commit Sequence

run-PR 의 commit sequence: Wave 1 (Fix A) → Wave 2 (Fix B) → Wave 3 (Fix C ci.yml) → Wave 3 (skip guard 제거).

Squash merge 후 main 에는 단일 commit 으로 압축. 단, PR 자체의 CI 는 각 commit 시점에 trigger → cumulative 검증.

### 6.3 Plan-PR vs Run-PR Separation

- Plan-PR: 본 5 파일 (spec/plan/design/acceptance/tasks) 만. 무코드.
- Run-PR: 3 sub-fix YAML/shell + drift test skip guard 제거. CI 검증.
- Sync-PR: 본 SPEC spec.md frontmatter `status: draft → completed` + CHANGELOG + version bump.

---

## 7. Open Design Questions (deferred to run-phase)

### 7.1 awk vs sed 최종 선택

§2.3 D-1 plan-phase 잠정 결정 A2 (awk). run-phase W1 에서 plan-auditor 가 A1 (sed) 와 비교 후 최종 선택. 두 안 모두 acceptable.

### 7.2 permissions block 위치

§3.3 D-2 잠정 workflow-level. run-phase W2 에서 job-level vs workflow-level 코드 diff 비교 후 선택.

### 7.3 fetch-depth 다른 workflow 확장

§4.3 D-3 잠정 ci.yml only. 후속 SPEC `SPEC-V3R4-CI-WORKFLOWS-AUDIT-001` 후보. 본 SPEC out of scope.

### 7.4 detect-language windows-latest 검증

§2.4 windows-bash 호환성 검증은 ci.yml 의 detect-language 사용 부재 → CI 자체로 검증 불가. fixture 기반 local 검증 + 호출 workflow (gemini-review.yml 등) 에서 windows runner 사용 여부 확인.

**Plan-phase 결정**: gemini-review.yml 등 호출 workflow 는 ubuntu-latest 만 (검토 결과). windows 영향 0 → 별도 검증 task 불요.

---

## 8. Anti-patterns (피해야 할 design)

### 8.1 pipefail-off 광범위 적용

A3 (set +o pipefail) 는 SIGPIPE remediation 의 빠른 fix 처럼 보이지만, 다른 error 도 silent ignore 하므로 CI 신뢰성 회귀. 본 SPEC 거부.

### 8.2 광범위 permissions 부여

spec-status-auto-sync.yml 에 `permissions: write-all` 또는 `actions: write` 등 불필요한 scope 추가는 least-privilege 위배. 본 SPEC 은 `contents: write` + `issues: write` 만.

### 8.3 fetch-depth: 0 일괄 전역 적용

모든 workflow 의 fetch-depth: 0 일괄 변경은 unrelated workflow (codeql, test-install) 의 cache + runtime 영향. drift walker 가 실제 필요한 ci.yml 만 변경 (D-3 선택).

### 8.4 walker code 변경

drift.go walker 에 shallow-clone fallback (e.g., `if !fullClone { return early }`) 추가는 walker contract 위배 (full history 전제). CI 환경 fix 가 올바른 layer.

### 8.5 skip guard 완전 제거 without local probe

Skip-A 는 fork PR 사용자 환경 break. 본 SPEC 거부 (Skip-B 채택).

---

## 9. References (cross-file)

- `spec.md` — 3 sub-fix root cause + REQ-CIIF-001~007
- `plan.md` — Wave 1/2/3 implementation strategy
- `acceptance.md` — AC-CIIF-001~004 binary verification + Given/When/Then
- `tasks.md` — task-level breakdown (W1-T1 ~ W3-T5)
- `.moai/specs/SPEC-V3R4-LINT-SPECID-GREP-FIX-001/design.md` — walker word-boundary design precedent
- `.moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002/design.md` — Category A/B decision framework precedent
- GitHub Actions docs: `permissions:` keyword reference
- POSIX awk + sed semantics reference
- actions/checkout README: fetch-depth: 0 rationale

---

## 10. Summary Decision Matrix

| Decision | Plan-phase Choice | Alternatives Considered | Run-phase Verification |
|----------|-------------------|--------------------------|------------------------|
| D-1 SIGPIPE | A2 (awk NR==1{print;exit}) | A1 (sed), A3 (pipefail off) | 16-language fixture + CI 3-run |
| D-2 Permissions | B1 (workflow-level minimum) | B2 (job-level), B3 (pull_request_target) | dummy PR merge + 403 grep |
| D-3 fetch-depth scope | C-Scope-1 (ci.yml 5 steps) | C-Scope-2 (all-workflows), C-Scope-3 (drift-only) | grep 5 steps + drift test PASS |
| D-3 skip guard | Skip-B (env probe 제거, local probe 보존) | Skip-A (complete remove), Skip-C (strengthen) | local + CI run 양쪽 PASS |
| D-4 Compatibility | non-breaking (breaking: false) | n/a | regression baseline 0/0 |

---

본 design.md 는 run-phase 시작 전 plan-auditor 검토 대상. plan-auditor PASS ≥ 0.85 가 run-phase 진입 gate.
