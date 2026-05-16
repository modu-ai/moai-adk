# Acceptance — SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-16 | manager-spec | 초기 acceptance. 8개 AC (AC-SDF-001~008) + 5개 Given/When/Then 시나리오 (primary, Pattern A bulk, Pattern D-G exemption, idempotency, lint.skip ban) + 4개 edge case 시나리오 + Definition of Done 체크리스트 + Quality Gate criteria. |

---

## 1. Acceptance Criteria Summary

| AC ID | 검증 대상 | Primary REQ |
|-------|----------|-------------|
| AC-SDF-001 | Binary lint outcome (StatusGitConsistency WARN = 0) | REQ-SDF-007 |
| AC-SDF-002 | Pattern A bulk downgrade (50 SPECs) | REQ-SDF-004 |
| AC-SDF-003 | Pattern B+C per-SPEC verification (10 SPECs) | REQ-SDF-005, REQ-SDF-006 |
| AC-SDF-004 | Pattern D/E/F/G detector exemption (4 SPECs + 4 tests) | REQ-SDF-001, REQ-SDF-008 |
| AC-SDF-005 | Pattern H recursive cleanup (4 SPECs sync-phase) | REQ-SDF-009 |
| AC-SDF-006 | lint.skip ban preservation (0 new entries) | REQ-SDF-010 |
| AC-SDF-007 | Out-of-scope files untouched | REQ-SDF-011, 012, 013, 014 |
| AC-SDF-008 | Bulk script idempotency | REQ-SDF-015 |

---

## 2. Given/When/Then Scenarios

### Scenario 1 — Primary Success (AC-SDF-001 Binary Outcome)

**Given**:
- main HEAD가 `758341089` (PR #937 merge) 또는 그 이후
- `moai spec lint --strict 2>&1 | grep -c "StatusGitConsistency"` 가 64 (또는 plan-phase 추정 ±변동) 출력
- 본 SPEC 의 run-phase 진입 직전

**When**:
- 본 SPEC 의 모든 5 Wave (BASELINE → Wave 1 Pattern A → Wave 2 Pattern B+C → Wave 3 Pattern D-G detector → Wave 5 Pattern H sync-phase) 가 완료되어 main 에 머지됨

**Then**:
- main HEAD 에서 `moai spec lint --strict 2>&1 | grep -c "StatusGitConsistency"` 가 **정확히 0** 을 출력
- 다른 lint rule (예: OrphanBCID, RequirementCoverage) 의 WARN/ERROR 는 본 검증 대상 외 (out of scope per §6)

---

### Scenario 2 — Pattern A Bulk Downgrade (AC-SDF-002)

**Given**:
- BASELINE milestone에서 `affected-list-pattern-A.txt` 가 산출됨 (50 ± 변동 SPEC IDs)
- 각 SPEC의 frontmatter `status: completed` 이고 git-implied = `implemented` 상태
- bulk script `.moai/scripts/status-drift-cleanup.go` 작성 완료

**When**:
- `go run .moai/scripts/status-drift-cleanup.go --pattern A` 실행

**Then**:
- 50 Pattern A SPECs 모두에서 다음 변경:
  - `awk '/^status:/' .moai/specs/<SPEC-ID>/spec.md` → `status: implemented` (50/50 hit)
  - frontmatter `version` 이 patch +1 (예: `"0.3.0" → "0.3.1"`)
  - frontmatter `updated_at: "2026-05-16"` (또는 실행 시점 날짜)
  - HISTORY 표 마지막에 1 row 추가 (description에 `Pattern A` + `SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001` 포함)
- `git status .moai/specs/` 결과 50 SPECs `spec.md` 만 modified (다른 파일 0건)
- `git diff .moai/specs/` 의 변경 라인 수: per SPEC ~4 lines (status, version, updated_at, HISTORY row)
- `moai spec lint --strict 2>&1 | grep "StatusGitConsistency" | grep -E "$(cat affected-list-pattern-A.txt | sed 's|.moai/specs/||;s|/spec.md||' | tr '\n' '|')" | wc -l` → 0

---

### Scenario 3 — Pattern D Detector Exemption (AC-SDF-004 case D)

**Given**:
- `t.TempDir()` 안에 git 저장소 초기화 (CLAUDE.local.md §6 Test Isolation)
- 다음 git fixture commits (oldest → newest):
  - Commit 1: `feat(SPEC-X-001): initial implementation`
  - Commit 2: `docs(sync): SPEC-X-001 sync close`
- 임시 SPEC 파일 `<tmpDir>/.moai/specs/SPEC-X-001/spec.md` 작성, frontmatter `status: superseded`, `version: "1.0.0"`
- `internal/spec/lint.go::StatusGitConsistencyRule` 의 Wave 3 적용된 코드 (terminalStatusEnum 추가됨)

**When**:
- `rule := &StatusGitConsistencyRule{}` 생성
- `findings := rule.Check(testContext(tmpDir, "SPEC-X-001"))` 호출

**Then**:
- `len(findings) == 0` (terminal state exemption 발동, finding 미발생)
- 4 SPEC (SPEC-LSP-001, SPEC-V3R3-HARNESS-001, SPEC-I18N-001-ARCHIVED, SPEC-V3R3-WEB-001) 의 frontmatter `status` 변경 0건
- `go test ./internal/spec/... -run TestStatusGitConsistency_TerminalState_Superseded_Completed` PASS

---

### Scenario 4 — Bulk Script Idempotency (AC-SDF-008)

**Given**:
- Wave 1 적용 완료 후 (모든 Pattern A SPECs 처리됨)
- main에 미머지 상태 (working tree 에 변경 사항 유지)

**When**:
- `go run .moai/scripts/status-drift-cleanup.go --pattern A` 두 번 연속 실행

**Then**:
- 1차 실행: script 출력 `OK — applied=50 skipped=0` (또는 동등)
- 2차 실행: script 출력 `OK — applied=0 skipped=50 (idempotent)` (모든 SPECs `currentStatus == ToStatus` 분기로 no-op)
- `git diff --stat` 가 1차 실행 후와 2차 실행 후 동일 (추가 변경 0건)
- script 자체 exit code 0 (성공)

---

### Scenario 5 — lint.skip Ban Preservation (AC-SDF-006)

**Given**:
- 본 SPEC plan-phase 시점에 `grep -lR "lint.skip" .moai/specs/ | wc -l` → 0 (LSKC-001 PR #937 머지 후 baseline)
- 본 SPEC 의 모든 Wave 진행 중

**When**:
- 본 SPEC PR이 main 에 머지됨

**Then**:
- main HEAD 에서 `grep -lR "lint.skip" .moai/specs/ | wc -l` → 여전히 0 (새 lint.skip 도입 0건)
- `git diff main..HEAD .moai/specs/ | grep "+.*lint\.skip"` → 0 hit (새 추가 0줄)
- 본 SPEC 자체 frontmatter 에 `lint.skip` 필드 부재

---

## 3. Edge Case Scenarios

### Edge 1 — BASELINE 측정 결과가 64에서 ±20 변동

**Given**:
- plan-phase 시점 측정: 64 WARN
- run-phase BASELINE milestone 측정: 75 WARN (예: 새 SPEC 7건 + 기존 변경 4건 추가됨)

**When**:
- BASELINE milestone에서 affected-list 재산출

**Then**:
- `affected-list-pattern-{A,B,C,H}.txt` 가 plan-phase 추정 size와 다르더라도 정상 진행
- Pattern 재분류 결과를 progress.md 에 기록
- spec.md §1.3 의 count는 plan-phase 추정으로 명시 (research.md §10 Note 참조)
- 큰 변동 (±20 이상) 시: 사용자 확인 (orchestrator AskUserQuestion via plan-auditor)

---

### Edge 2 — Pattern B+C 분기 (b) keep 이 4건 이상 발생

**Given**:
- Wave 2 verification 진행 중
- 10 SPECs 개별 검토 결과 분기 (b) keep 4건 식별 (frontmatter status 가 사실에 부합하나 git history 가 stale)

**When**:
- M3 milestone에서 카운트 추적

**Then**:
- 4건 이상 발견 시 orchestrator AskUserQuestion 발동:
  - 옵션 (i): 분기 (b) keep 유지 (4건 lint --strict WARN 잔존 + run-verification.md 기록) → AC-SDF-001 partial fail
  - 옵션 (ii): 분기 (b) → 분기 (a) downgrade 강제 (정책 약화) → AC-SDF-001 GREEN
  - 옵션 (iii): 4건 별도 SPEC 분리 (본 SPEC scope 축소) → AC-SDF-001 GREEN (본 SPEC), 후속 SPEC 발급
- 기본 권장: 옵션 (iii) — scope 명확성 보존

---

### Edge 3 — Detector Exemption이 기존 lint test 회귀

**Given**:
- Wave 3 코드 변경 적용 (`terminalStatusEnum` + Check 분기 추가)
- `go test ./internal/spec/...` 실행

**When**:
- 기존 테스트 일부 fail 발생 (예: 기존 `TestStatusGitConsistency_SupersededMismatch` 가 WARN 발생을 기대)

**Then**:
- fail 한 테스트의 expected output 갱신:
  - 기존 `len(findings) == 1` → 새 `len(findings) == 0` (terminal state exemption 발동)
- 또는 기존 테스트의 frontmatter status 를 non-terminal 로 변경하여 의미 보존
- 갱신 commit에 `@MX:NOTE` 부착 + SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 reference
- `go test ./internal/spec/...` 전체 PASS 회복

---

### Edge 4 — Pattern H sync-phase 재귀가 2-stage 발생

**Given**:
- Wave 5 sync-phase 적용 (4 cleanup-chain SPECs 정합화)
- sync PR squash merge 후 새 `chore(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 sync` commit 발생
- 본 SPEC 자체가 또다시 H 패턴 entry 생성 가능

**When**:
- 다음 `moai spec lint --strict` 실행

**Then**:
- 본 SPEC frontmatter `status: completed` + git-implied (post-sync) 비교:
  - 만약 git-implied 가 여전히 mismatch (예: `implemented`) → WARN 재발
  - 만약 일치 → 정상
- WARN 재발 시: 별도 cleanup SPEC 발급 (본 SPEC scope 외, 드물게 발생 < 1%)
- 또는 본 SPEC sync-phase 에 추가 sub-PR 으로 2-stage 흡수

---

## 4. Definition of Done Checklist

### 4.1 Functional Acceptance

- [ ] AC-SDF-001: `moai spec lint --strict | grep -c "StatusGitConsistency"` → 0
- [ ] AC-SDF-002: Pattern A 50 SPECs frontmatter `status: implemented`, version patch bump, HISTORY row
- [ ] AC-SDF-003: Pattern B+C 10 SPECs verification 완료, run-verification.md 10 row 기록
- [ ] AC-SDF-004: terminal state exemption 코드 + 4 unit tests PASS
- [ ] AC-SDF-005: Pattern H 4 cleanup-chain SPECs sync-phase 정합화
- [ ] AC-SDF-006: 새 `lint.skip` entry 도입 0건 (REQ-SDF-010)
- [ ] AC-SDF-007: 64-affected 외 파일 변경 0건 (REQ-SDF-011~014)
- [ ] AC-SDF-008: bulk script idempotency 검증 (2nd run no-op)

### 4.2 Code Quality

- [ ] `go test ./...` 전체 PASS (4 새 cases + 기존 회귀 0)
- [ ] `go test -race ./internal/spec/` PASS
- [ ] `go vet ./...` 0 issue
- [ ] `golangci-lint run` 0 warning
- [ ] @MX:NOTE / @MX:ANCHOR 부착 (mx-tag-protocol.md 준수, code_comments=ko)
- [ ] `internal/spec/drift.go` 변경 0줄 (REQ-SDF-011 enforcement)
- [ ] `internal/spec/transitions.go` 변경 0줄
- [ ] 다른 lint rule 파일 변경 0줄

### 4.3 Documentation

- [ ] spec.md / plan.md / design.md / research.md / acceptance.md frontmatter 9-field 모두 (id, version, status, created_at, updated_at, author, priority, labels, issue_number)
- [ ] spec-compact.md 생성 (REQ + AC + Exclusions 추출)
- [ ] progress.md 에 plan_complete_at + plan_status: audit-ready 기록
- [ ] CHANGELOG `[Unreleased]` row 1줄 (sync-phase 책임 — manager-docs)
- [ ] HISTORY row 각 단계 (plan / run / sync) 시점에 추가

### 4.4 Workflow Compliance

- [ ] BODP 평가 기록 (HISTORY 0.1.0 + plan.md §10) — A=¬ B=¬ C=¬ → main
- [ ] plan-in-main 정책 준수 (worktree 미사용 — feedback_worktree_never_use)
- [ ] PR squash merge (Enhanced GitHub Flow §18.3)
- [ ] branch: `feat/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001`
- [ ] PR title 컨벤션 준수 (feat(spec):)
- [ ] Conventional Commits commit messages
- [ ] CI Lint + Test + Build all green
- [ ] plan-auditor PASS (≥ 0.85, 목표 iter 1)

### 4.5 Status Transition

- [ ] `draft` (plan-phase 종료) → `in-progress` (run-phase 시작) → `implemented` (run-phase 종료) → `completed` (sync-phase 종료)
- [ ] 각 transition 시점에 적절한 commit prefix (`plan(spec):` → `feat(SPEC-XXX):` → `docs(sync):`)

---

## 5. Quality Gate Criteria

### 5.1 Lint Quality Gate

| 검증 | Threshold | 검증 명령 |
|------|-----------|----------|
| StatusGitConsistency WARN | 0 (binary) | `moai spec lint --strict 2>&1 \| grep -c "StatusGitConsistency"` |
| Total lint ERROR | 0 | `moai spec lint --strict 2>&1 \| grep -c "ERROR"` |
| Other lint WARN | (baseline 유지) | `moai spec lint --strict 2>&1 \| grep -c "WARN"` (baseline 대비 증가 0건) |

### 5.2 Test Quality Gate

| 검증 | Threshold | 검증 명령 |
|------|-----------|----------|
| Test PASS rate | 100% | `go test ./internal/spec/... -v` |
| Race condition | 0 | `go test -race ./internal/spec/` |
| Coverage (internal/spec/lint package) | ≥ 85% (기존 baseline 유지) | `go test ./internal/spec/lint/... -cover` |

### 5.3 Code Quality Gate

| 검증 | Threshold | 검증 명령 |
|------|-----------|----------|
| `go vet` issues | 0 | `go vet ./...` |
| `golangci-lint` warnings | 0 | `golangci-lint run` |
| @MX tag coverage (Wave 3 코드) | terminalStatusEnum + Check 분기에 @MX:NOTE/ANCHOR | `grep -A2 "@MX:" internal/spec/lint.go` |

### 5.4 Scope Compliance Gate

| 검증 | Threshold | 검증 명령 |
|------|-----------|----------|
| 변경 파일 set | 63 SPECs + 1 detector code + 1 detector test + 1 bulk script + 본 SPEC artifacts | `git diff main..HEAD --name-only \| sort -u` |
| `internal/spec/drift.go` 변경 | 0 lines | `git diff main..HEAD internal/spec/drift.go \| wc -l` |
| `internal/spec/transitions.go` 변경 | 0 lines | `git diff main..HEAD internal/spec/transitions.go \| wc -l` |
| 새 `lint.skip` 추가 | 0 hits | `git diff main..HEAD .moai/specs/ \| grep "+.*lint\.skip" \| wc -l` |
| 64-외 SPEC 변경 | 0 files | (run-verification.md 의 affected-list와 비교) |

---

## 6. Out-of-Scope Verification (Negative Tests)

본 SPEC 의 out-of-scope 항목들이 변경되지 않았는지 검증:

### 6.1 Walker Filter Unchanged (REQ-SDF-011)

```bash
git diff main..HEAD internal/spec/drift.go
# expected: empty (0 lines)
```

### 6.2 StatusGitConsistencyRule Not Deprecated (REQ-SDF-012)

```bash
grep -c "StatusGitConsistencyRule" internal/spec/lint.go
# expected: 동일 또는 증가 (deprecation 없음, exemption만 추가)
```

### 6.3 No New lint.skip (REQ-SDF-010)

```bash
grep -lR "lint.skip" .moai/specs/ | wc -l
# expected: 0 (baseline 동일, LSKC-001 PR #937 머지 후 상태)
```

### 6.4 No Body Modification on 64 SPECs (REQ-SDF-013)

```bash
# 각 SPEC 의 body sha256 비교 (HISTORY row 1줄 제외)
for spec in $(cat affected-list-pattern-*.txt); do
  awk '/^---$/{f++} f>=2' "$spec" | sed '/^| 0\.[0-9]*\.[0-9]* | 2026-05-16/d' | sha256sum
done > /tmp/post-sha.txt
# compare with baseline captured at M1
diff /tmp/baseline-sha.txt /tmp/post-sha.txt
# expected: empty (no drift)
```

### 6.5 No Out-of-Scope File Changes (REQ-SDF-014)

```bash
git diff main..HEAD --name-only | sort -u > /tmp/changed-files.txt
# expected set:
# - 63 SPECs spec.md (Pattern A + B+C 분기 a + H)
# - internal/spec/lint.go OR internal/spec/lint/checks/status_git_consistency.go
# - internal/spec/lint_test.go OR equivalent test file
# - .moai/scripts/status-drift-cleanup.go (Optional)
# - .moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001/* (본 SPEC artifacts)
# - 외 다른 파일 0건 (CI workflow, docs-site, README, CHANGELOG, etc.)
```

---

## 7. Test Plan Coverage Summary

| Test Type | Count | Location | Verifies AC |
|-----------|-------|----------|-------------|
| Unit Test (terminal state exemption) | 4 | `internal/spec/lint_test.go` 또는 동등 | AC-SDF-004 |
| Integration Test (bulk script idempotency) | 1 | shell script (M7 verification) | AC-SDF-008 |
| Integration Test (lint --strict baseline) | 1 | M7 verification | AC-SDF-001 |
| Manual Verification (Pattern B+C decision) | 10 | run-verification.md | AC-SDF-003 |
| Integration Test (frontmatter diff scope) | 1 | M7 git diff verification | AC-SDF-007 |
| Integration Test (lint.skip ban) | 1 | M7 grep verification | AC-SDF-006 |
| Recursive Cleanup Test (Pattern H) | 1 | Wave 5 sync-phase | AC-SDF-005 |
| Bulk Apply Test (Pattern A) | 1 | M2 verification | AC-SDF-002 |

**총 19 test cases / verifications** 으로 8 AC 전체 커버. 추가 테스트 도입 절제 (REQ-SDF-014 정신 + 부담 최소화).
