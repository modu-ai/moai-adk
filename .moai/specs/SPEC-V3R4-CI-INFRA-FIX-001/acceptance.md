# SPEC-V3R4-CI-INFRA-FIX-001 — Acceptance Criteria

> **Binary AC discipline**: 모든 AC는 PASS/FAIL 이분 결정. 부분 충족, partial credit, soft threshold 없음.
> **Verification environment**: main checkout, run-PR CI logs, `moai spec lint --strict` (cwd = repo root).

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-16 | manager-spec | 초기 draft. 4개 binary AC (Given-When-Then 형식) + Edge Cases + Definition of Done + Quality Gates. 모든 AC는 단일 verification command 로 PASS/FAIL 결정 가능. 각 sub-fix (A/B/C) 와 regression baseline (D) 매핑. |

---

## 1. Overview

본 acceptance.md 는 SPEC-V3R4-CI-INFRA-FIX-001 의 7개 EARS requirements 가 implementation 완료 시점에 검증 가능함을 정의한다. 각 AC 는 Given-When-Then 시나리오 + 구체적 verification command + 기대 출력을 포함한다.

### 1.1 AC ↔ REQ Mapping

| AC ID         | REQ ID(s)                              | Sub-fix     | Wave              | Verification scope |
|---------------|----------------------------------------|-------------|-------------------|--------------------|
| AC-CIIF-001   | REQ-CIIF-001                           | Fix A       | W1 (SIGPIPE)      | End-to-end: CI 3-run detect-language step success |
| AC-CIIF-002   | REQ-CIIF-002, REQ-CIIF-006             | Fix B       | W2 (403)          | E2E: spec-status-auto-sync workflow run success |
| AC-CIIF-003   | REQ-CIIF-003, REQ-CIIF-004, REQ-CIIF-005 | Fix C    | W3 (fetch-depth)  | Code grep + CI test execution (not skip) |
| AC-CIIF-004   | REQ-CIIF-007                           | Cross-cutting | Final            | Regression: `moai spec lint --strict` 0/0 |

### 1.2 Sub-fix Mapping (3-defect bundle)

| Sub-fix | Affected Files | REQ(s) | AC(s) |
|---------|----------------|--------|-------|
| **Fix A — detect-language SIGPIPE** | `.github/actions/detect-language/action.yml` | REQ-CIIF-001 | AC-CIIF-001 |
| **Fix B — spec-status-sync 403** | `.github/workflows/spec-status-auto-sync.yml` | REQ-CIIF-002, REQ-CIIF-006 | AC-CIIF-002 |
| **Fix C — checkout fetch-depth 0** | `.github/workflows/ci.yml` + `internal/spec/drift_specid_grep_test.go` | REQ-CIIF-003, REQ-CIIF-004, REQ-CIIF-005 | AC-CIIF-003 |
| **Cross-cutting regression** | (all of above + repo state) | REQ-CIIF-007 | AC-CIIF-004 |

---

## 2. Acceptance Scenarios

### AC-CIIF-001 — detect-language action SIGPIPE 0건 (Fix A)

**maps REQ-CIIF-001**

**Given**:
- run-PR (또는 sync-PR, or main 머지 후) commits 에 `.github/actions/detect-language/action.yml` SIGPIPE remediation (`head -1` → `awk 'NR==1{print;exit}'` per design.md D-1) 가 포함되어 있다.
- CI workflows 가 detect-language action 을 호출 (gemini-review.yml, glm-review.yml, codex-review.yml 등 — 본 SPEC 외 workflow).

**When**:
- 다음 명령으로 최근 3 CI runs 의 detect-language 단계 실패 수를 측정한다.
- (CI workflow 가 detect-language 를 사용하지 않는 경우 — ci.yml 만 트리거된 경우 — 본 AC 는 vacuous PASS 처리, edge-1 참조)

**Then**:
- 최근 3 consecutive CI runs 중 detect-language step 의 exit 141 (SIGPIPE) 발생 0건.
- 출력값이 정확히 `0` 이어야 한다.

**Verification Command**:
```bash
# 본 SPEC run-PR head SHA 기준 최근 3 runs
PR_SHA=$(gh pr view <run-PR-number> --json headRefOid -q .headRefOid)

# detect-language 호출 workflow 중 본 PR commit 의 runs
gh run list --commit="$PR_SHA" --json conclusion,workflowName \
  | jq '[.[] | select(.conclusion == "failure" and (.workflowName | test("gemini-review|glm-review|codex-review")))] | length'
# Expected: 0

# 보조 검증 (run log 에 exit 141 부재)
gh run view <recent-run-id> --log \
  | grep -c "exit 141\|SIGPIPE"
# Expected: 0
```

**Pre-fix baseline** (NAMESPACE-001 PR #944): 4 SIGPIPE incidents (4 CI runs 에서 detect-language step 실패).
**Post-fix target**: 0 SIGPIPE incidents (3 consecutive runs).

---

### AC-CIIF-002 — spec-status-auto-sync 403 0건 (Fix B)

**maps REQ-CIIF-002, REQ-CIIF-006**

**Given**:
- run-PR commits 에 `.github/workflows/spec-status-auto-sync.yml` `permissions:` block 추가 (`contents: write`, `issues: write` per design.md D-2) 가 포함되어 있다.
- 본 SPEC 의 run-PR 또는 sync-PR 머지 후, workflow 가 `pull_request: types: [closed]` trigger 로 발동된다.

**When**:
- 본 SPEC 의 run-PR 또는 sync-PR 머지 직후 spec-status-auto-sync workflow run 결과를 검증한다.

**Then**:
- 해당 workflow run 의 conclusion 이 `success` (실패 0건).
- run log 에 HTTP `403` 부재.

**Verification Command**:
```bash
# 본 SPEC sync-PR 머지 trigger 된 workflow run
SYNC_PR=<sync-PR-number>
RUN_ID=$(gh run list --workflow=spec-status-auto-sync.yml --limit=5 --json databaseId,event,headSha \
  | jq -r --arg sha "$(gh pr view $SYNC_PR --json mergeCommit -q .mergeCommit.oid)" '.[] | select(.headSha == $sha) | .databaseId' \
  | head -1)

# Conclusion 검증
gh run view "$RUN_ID" --json conclusion -q .conclusion
# Expected: "success"

# 403 grep
gh run view "$RUN_ID" --log 2>&1 | grep -c "403"
# Expected: 0
```

**Pre-fix baseline** (NAMESPACE-001 force-push 후): 403 push failure 1건.
**Post-fix target**: 403 부재 (workflow run success).

**Note**: 본 AC 검증은 본 SPEC 의 sync-PR 머지 후에만 가능 — sync-PR 자체가 SPEC frontmatter 변경 → spec-status-auto-sync 발동의 자연 trigger. 추가 dummy PR 불요.

---

### AC-CIIF-003 — fetch-depth: 0 + drift test 정상 실행 (Fix C)

**maps REQ-CIIF-003, REQ-CIIF-004, REQ-CIIF-005**

**Given**:
- run-PR commits 에 다음 변경이 포함되어 있다:
  - `.github/workflows/ci.yml` 의 5 `actions/checkout@v5` step 모두 `with: fetch-depth: 0` 명시 (design.md D-3 C-Scope-1)
  - `internal/spec/drift_specid_grep_test.go:21-27` 의 `GITHUB_ACTIONS` env probe 제거 (design.md D-3 Skip-B)
  - `internal/spec/drift_specid_grep_test.go:30-34` 의 local probe 보존
- 본 SPEC run-PR CI run 이 main HEAD 에서 변경된 ci.yml 사용 (PR 자체 workflow 적용).

**When**:
- run-PR CI run 에서 `TestGetGitImpliedStatus_HARNESS001Resolution` 실행 결과를 확인한다.
- ci.yml 의 5 checkout step 의 fetch-depth: 0 명시 여부를 grep 한다.

**Then**:
- ci.yml 5 step 모두 `with: fetch-depth: 0` 명시 (grep count = 5).
- CI run log 에 test 결과 `--- PASS` (not `--- SKIP`).
- exit code 0 (test passes).

**Verification Commands**:
```bash
# 1) ci.yml 의 fetch-depth: 0 5 step 확인
grep -c "fetch-depth: 0" .github/workflows/ci.yml
# Expected: 5

# 2) CI run 에서 drift test PASS
RUN_ID=$(gh pr view <run-PR-number> --json statusCheckRollup \
  | jq -r '.statusCheckRollup[] | select(.name | startswith("Test (ubuntu")) | .detailsUrl' \
  | head -1 \
  | grep -oE '[0-9]+$')

gh run view "$RUN_ID" --log 2>&1 \
  | grep "TestGetGitImpliedStatus_HARNESS001Resolution" \
  | grep -E "PASS|SKIP" \
  | head -1
# Expected: contains "PASS", NOT "SKIP"

# 3) Skip guard 제거 확인 (소스 검증)
grep -c 'os.Getenv("GITHUB_ACTIONS")' internal/spec/drift_specid_grep_test.go
# Expected: 0 (skip guard 완전 제거)

# 4) local probe 보존 확인
grep -c "SPEC-V3R4-HARNESS-001 commits not available in local git history" internal/spec/drift_specid_grep_test.go
# Expected: 1 (fork/shallow user case skip 보존)
```

**Pre-fix baseline**: drift test skip 1건 (CI `GITHUB_ACTIONS=true` 환경).
**Post-fix target**: drift test PASS 1건 (CI 정상 실행).

---

### AC-CIIF-004 — `moai spec lint --strict` 0/0 regression baseline (Cross-cutting)

**maps REQ-CIIF-007**

**Given**:
- 본 SPEC run-PR 또는 sync-PR 머지 후 main HEAD.
- LSGF-001 + FOLLOWUP-002 lifecycle COMPLETE 후 main `9be0ef03b` 의 lint baseline `0 error(s), 0 warning(s)`.

**When**:
- `moai spec lint --strict 2>&1 | tail -3` 을 main 머지 후 실행한다.

**Then**:
- 출력의 마지막 의미 있는 줄이 정확히 `0 error(s), 0 warning(s)` 이어야 한다 (또는 `✓ No findings (0/0)`).
- exit code = 0 이어야 한다.

**Verification Command**:
```bash
# main 머지 후
git checkout main && git pull
moai spec lint --strict 2>&1 | tail -3
# Expected: 출력 마지막 줄에 "0 error(s), 0 warning(s)" 또는 "✓ No findings (0/0)"

echo $?
# Expected: 0
```

**Pre-fix baseline** (main `9be0ef03b`): `0 error(s), 0 warning(s)` (FOLLOWUP-002 closeout 후).
**Post-fix target**: 동일 baseline 유지 (regression 0).

**Note**: 본 AC 는 본 SPEC 의 CI infra 변경이 lint walker (drift.go) 의 동작을 우회적으로 변경하지 않음을 보장. CI infrastructure-only change 가 사용자 영향 0건임을 확인.

---

## 3. Edge Cases

### EC-001 — detect-language action 이 본 SPEC run-PR 의 CI 에서 트리거되지 않음

**Scenario**: ci.yml 은 detect-language 미사용. detect-language 는 gemini-review.yml 등에서만 사용. 본 SPEC run-PR CI 가 ci.yml 만 trigger 시 detect-language step 자체 부재.

**Expected behavior**:
- AC-CIIF-001 의 verification command 가 `0` 반환 (vacuous PASS).
- detect-language SIGPIPE remediation 의 verification 은 별도 manual trigger 필요 (e.g., main 머지 후 gemini-review.yml 발동 PR 작성).
- AC 충족 기준: action.yml 의 코드 diff 가 design.md D-1 채택 솔루션 (awk) 와 일치 + CI 자체 fail 부재.

**Mitigation**:
- 추가 verification: `bash .github/actions/detect-language/action.yml` (action 의 shell snippet 만 추출하여 local 실행) — fixture directory 에서 detection 성공 확인.
- 또는 fixture-driven test (manager-develop 가 run-phase 에서 추가).

### EC-002 — fork PR 사용자 환경 (PR-from-fork) 에서 drift test 거동

**Scenario**: external contributor 가 fork 에서 PR 작성. fork repo 의 git history 에 main repo 의 SPEC commits 부재.

**Expected behavior**:
- drift_specid_grep_test.go 의 local probe (line 30-34) 가 발동 → test skip (graceful).
- AC-CIIF-003 검증은 본 SPEC 의 run-PR (internal, full clone) 에서만 검증 → fork PR edge case 영향 0.

**Existing behavior preserved**: Skip-B (env probe 제거, local probe 보존) 결정으로 fork/shallow 사용자 환경에서 변화 없음.

### EC-003 — spec-status-auto-sync workflow 가 본 SPEC sync-PR 의 PR title 인식 실패

**Scenario**: sync-PR title `sync(spec): SPEC-V3R4-CI-INFRA-FIX-001 lifecycle COMPLETE — ...` 의 SPEC-ID extraction (workflow line 25: `grep -oE 'SPEC-[A-Z0-9-]+-[0-9]+'`) 이 매칭 실패.

**Expected behavior**:
- workflow line 26-28: `if [ -z "$SPEC_IDS" ]; then echo "No SPEC-IDs found in PR title"; exit 0; fi` → graceful exit, workflow conclusion `success`.
- AC-CIIF-002 검증 통과 (workflow run success).

**Likely scenarios**:
- SPEC-V3R4-CI-INFRA-FIX-001 은 정상 매칭 (regex `SPEC-[A-Z0-9-]+-[0-9]+` 충족).
- 따라서 EC-003 발동 가능성 낮음. fallback 안전성 보장.

### EC-004 — fetch-depth: 0 적용 후 다른 ci.yml step 의 cache 회귀

**Scenario**: actions/setup-go@v6 의 cache key 가 fetch-depth 에 의존.

**Expected behavior**:
- actions/setup-go cache key 는 `go.sum` + `go-version` 기반 (action 내부 구현). fetch-depth 무관.
- cache hit-rate 유지.

**Mitigation**: run-PR 첫 CI run 의 cache hit 확인 (gh run view log 에 "Cache hit" 또는 "Cache restored from key" grep).

### EC-005 — Wave 1-3 적용 중 lint baseline 회귀

**Scenario**: Wave 적용 중간 commit 에서 `moai spec lint --strict` 가 새 ERROR/WARNING 노출.

**Expected behavior**:
- 본 SPEC 의 코드 변경 (action.yml, workflow.yml, ci.yml, drift_specid_grep_test.go) 은 모두 lint walker (drift.go) 영향 0건 (CI infra 만, SPEC frontmatter 미변경).
- lint regression 가능성 0.

**Mitigation**: 각 Wave commit 후 progress.md 에 lint 결과 기록. 회귀 발생 시 즉시 rollback (plan.md §5).

### EC-006 — windows-latest 에서 awk 호환성 이슈

**Scenario**: detect-language action 의 awk 솔루션 (design.md D-1 A2) 가 windows-bash (git bash) 에서 syntax error.

**Expected behavior**:
- POSIX awk 표준 `NR==1{print;exit}` 는 git bash 의 mawk/gawk 모두 호환.
- design.md §7.4 에서 windows runner 영향 0 확인 (detect-language 호출 workflow 는 모두 ubuntu-latest).

**Mitigation**: run-phase W1-T3 에서 manual local test on windows (optional, 영향 0 확인 시 skip).

### EC-007 — permissions 추가 후 GITHUB_TOKEN scope 의 다른 workflow 영향

**Scenario**: spec-status-auto-sync.yml 의 `permissions: contents: write` 가 다른 workflow (auto-merge.yml, release.yml) 의 권한 모델과 충돌.

**Expected behavior**:
- GitHub Actions 의 `permissions:` 는 workflow-scoped — 다른 workflow 영향 0.
- 각 workflow 는 자기 자신의 `permissions:` block 에 따라 독립적으로 scope 결정.

**Mitigation**: 본 SPEC 은 spec-status-auto-sync.yml 만 수정. 다른 workflow 검토 불필요.

---

## 4. Definition of Done

본 SPEC 의 implementation 이 "완료" 로 간주되려면 다음 모든 조건이 충족되어야 한다:

- [ ] **AC-CIIF-001** PASS: detect-language SIGPIPE 0건 (CI 3 consecutive runs).
- [ ] **AC-CIIF-002** PASS: spec-status-auto-sync workflow run conclusion=success + 403 grep=0.
- [ ] **AC-CIIF-003** PASS: ci.yml 5 step `fetch-depth: 0` 명시 (grep=5) + drift test `--- PASS` (not SKIP) + skip guard 제거 검증.
- [ ] **AC-CIIF-004** PASS: `moai spec lint --strict` 출력 `0 error(s), 0 warning(s)`.
- [ ] plan-PR `--squash` auto-merge 완료 + plan-auditor PASS ≥ 0.85.
- [ ] run-PR `--squash` auto-merge 완료 + CI all-green (Lint / Test (ubuntu/macos/windows) / Build × 5 / CodeQL 6/6 PASS per CLAUDE.local.md §18.7).
- [ ] sync-PR `--squash` auto-merge 완료 + CI all-green.
- [ ] spec.md frontmatter `status: draft → completed`, `version: "0.1.0" → "0.2.0"`, `updated: 2026-05-16`.
- [ ] HISTORY entry 0.2.0 추가 (sync-phase 완료 기록).
- [ ] CHANGELOG.md Unreleased 섹션에 CI-INFRA-FIX-001 entry 추가 (ko + en, `### Fixed` 분류).
- [ ] drift_specid_grep_test.go 의 `@MX:NOTE` / `@MX:REASON` 갱신 (skip guard 제거 reason 명시).
- [ ] LSGF-001 + FOLLOWUP-002 lint baseline 회귀 0건 검증.
- [ ] v3.0.0-rc1 release tag 발행 준비 완료 (별도 작업, 본 SPEC scope 외).

---

## 5. Quality Gate Criteria

본 SPEC 은 CI infrastructure 변경 + 소규모 Go test 정리. 표준 TRUST-5 quality gate 의 각 항목 적용:

| TRUST 5 Pillar | 적용 여부 | Verification |
|----------------|-----------|--------------|
| Tested         | 부분 적용 | drift_specid_grep_test.go skip guard 제거 후 PASS 검증 (`go test ./internal/spec/... -race -count=1`) |
| Readable       | 적용     | action.yml shell snippet 명확성 (awk explicit exit), permissions block 명시성, fetch-depth: 0 일관성 |
| Unified        | 적용     | 12-field canonical SPEC frontmatter schema 준수 + 5 checkout step 모두 일관 |
| Secured        | 적용     | permissions least-privilege (`contents: write` + `issues: write` 만) |
| Trackable      | 적용     | HISTORY entry + commit message + PR description 모두 SPEC-ID 명시 + 3 sub-fix 분류 명시 |

### QG-1 (plan-phase, this phase)
- plan-auditor PASS ≥ 0.85.
- 본 acceptance.md 의 4개 AC + 7개 Edge Cases 가 모두 binary/testable.
- design.md D-1 ~ D-4 4개 architectural decision 명시.

### QG-2 (run-phase, before merge)
- `go vet ./...` clean.
- `golangci-lint run ./...` clean.
- `go test ./internal/spec/... -race -count=1` GREEN (drift test 포함).
- AC-CIIF-001 ~ 003 모두 PASS.
- CI required checks 6/6 PASS (Lint / Test ubuntu/macos/windows / Build × 5 / CodeQL per CLAUDE.local.md §18.7).

### QG-3 (sync-phase, after merge)
- main HEAD 에서 `moai spec lint --strict` `0 error(s), 0 warning(s)` (AC-CIIF-004).
- 본 SPEC 자체의 frontmatter status `completed` 로 sync 됨 (sync-PR 머지 후).
- spec-status-auto-sync workflow 가 본 SPEC sync-PR 머지 시 success (AC-CIIF-002 자연 검증).
- LSGF-001 회귀 부재 (HARNESS-001/002/003 lint clean 유지).
- v3.0.0-rc1 release tag 발행 가능 상태 (precondition 충족).

---

## 6. Out-of-AC (의도적 비검증)

- detect-language action 의 16 supported languages 각각 detection 정확도 (별도 fixture test).
- spec-status-auto-sync workflow 의 fork PR 처리 (OQ-2 deferred).
- ci.yml 외 다른 workflow (codeql.yml 등) 의 fetch-depth 적정성 (D-3 결정 — 별도 follow-up SPEC).
- v3.0.0-rc1 release process 자체 (CLAUDE.local.md §18.8 별도 작업).
- detect-language action 의 windows runner 직접 호환성 검증 (호출 workflow 모두 ubuntu — EC-006 무영향).
- branch protection rule (CLAUDE.local.md §18.7) 변경 검증.
- moai-adk-go 사용자 프로젝트의 CI workflow 영향 (CI infra change 는 dev project only — `internal/template/templates/` 미변경).

---

## 7. AC ↔ REQ Mapping (verification matrix — concise)

| AC ID         | REQ ID(s)                              | Test Command (verification)                                                                       |
|---------------|----------------------------------------|---------------------------------------------------------------------------------------------------|
| AC-CIIF-001   | REQ-CIIF-001                           | `gh run list --commit=<sha> --json conclusion \| jq '[.[] \| select(.conclusion == "failure" and (.workflowName \| test("gemini\|glm\|codex")))] \| length'` → 0 |
| AC-CIIF-002   | REQ-CIIF-002, REQ-CIIF-006             | `gh run view <run-id> --json conclusion -q .conclusion` → "success" + `gh run view <id> --log \| grep -c "403"` → 0 |
| AC-CIIF-003   | REQ-CIIF-003, REQ-CIIF-004, REQ-CIIF-005 | `grep -c "fetch-depth: 0" .github/workflows/ci.yml` → 5 + CI log `TestGetGitImpliedStatus_HARNESS001Resolution` PASS |
| AC-CIIF-004   | REQ-CIIF-007                           | `moai spec lint --strict 2>&1 \| tail -3` → contains "0 error(s), 0 warning(s)" |

---

## 8. Risk-Acceptance Trade-offs (acceptance-level)

| Risk | Plan-phase Mitigation | Acceptance-level Tolerance |
|------|------------------------|----------------------------|
| AC-CIIF-001 의 vacuous PASS (EC-001) | fixture-based local test 추가 | run-phase 에서 manager-develop fixture verification 의무화 |
| AC-CIIF-002 가 sync-PR 머지 후에만 verify 가능 | sync-PR open 단계에서 dummy trigger 불요, 자연 trigger | sync-PR CI run 결과로 검증 — 즉시 확인 |
| AC-CIIF-003 의 5-step grep 이 fetch-depth: 0 일관성 검증만 | 추가: `git log -p` 로 5 step 모두 동일 시점 변경 확인 | run-PR CI run 자체가 effective verification |
| AC-CIIF-004 의 lint baseline 회귀 0 검증이 본 SPEC scope 와 무관 | non-breaking change 보장 (design.md D-4) | sync-phase 자동 검증 |
