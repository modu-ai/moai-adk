# SPEC-V3R4-LINT-SPECID-GREP-FIX-001 — Implementation Plan

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-16 | manager-spec | 초기 draft. 3-Wave TDD plan (RED → GREEN → REFACTOR). Approach 3종(A/B/C) 비교 후 plan-phase 잠정 채택 Approach B (2-pass post-filter, ExtractSPECIDs 재사용). 단일 파일 (`drift.go`) + 단일 test 파일 수정. BODP signals `¬a ¬b ¬c` → base `origin/main`. plan-in-main 원칙 (worktree 미사용). |

---

## 1. Overview

본 plan은 SPEC-V3R4-LINT-SPECID-GREP-FIX-001의 5개 EARS requirements (REQ-LSGF-001~005)를 3-Wave TDD 사이클로 분해하여 실행한다. 핵심은 `internal/spec/drift.go` `getGitImpliedStatus` walker에 word-boundary SPEC-ID 매칭 필터를 추가하는 외과적 변경 (surgical change).

### 1.1 Execution Mode

- **Plan-in-main**: BODP signals A=¬ (no depends_on) / B=¬ (no co-located worktree) / C=¬ (no open PR head) → main @ origin/main (per CLAUDE.local.md §18.12 BODP)
- **Worktree**: 사용 안 함 (project feedback `worktree_never_use` + 단일 source file 위주 수정)
- **Branch (plan-phase)**: `feat/SPEC-V3R4-LINT-SPECID-GREP-FIX-001-plan` (현재 브랜치)
- **Branch (run-phase)**: `feat/SPEC-V3R4-LINT-SPECID-GREP-FIX-001` (plan-PR 머지 후 fresh branch)
- **Lifecycle**: spec-anchored

### 1.2 Wave Breakdown

| Wave | Goal | Files | EARS coverage | Priority |
|------|------|-------|---------------|----------|
| Wave 1 (RED) | bug reproduction test 신설 + 회귀 가드 | `internal/spec/drift_test.go` (or new `drift_specid_grep_test.go`) | REQ-LSGF-001..004 | High |
| Wave 2 (GREEN) | drift.go walker에 word-boundary 필터 추가 | `internal/spec/drift.go` | REQ-LSGF-001..005 | High |
| Wave 3 (REFACTOR) | 코드 hygiene + @MX 태그 갱신 + 주석 + lint sweep 검증 | `internal/spec/drift.go` + lint sweep | (all) | Medium |

---

## 2. Approach Comparison (3-way analysis)

### Approach A — Regex word-boundary in git log --grep

**Idea**: `git log --grep="\\b<SPEC-ID>\\b"` with regex word boundary.

```go
cmd := exec.Command("git", "log", branch, "--oneline", "--no-merges",
    "--extended-regexp",
    "--grep=\\b"+regexp.QuoteMeta(specID)+"\\b",
    fmt.Sprintf("-%d", gitLogWindowSize))
```

**Pros**:
- Single-line change (drift.go:119-121).
- Filter pushed down to git (faster for huge histories).
- No Go-side regex matching needed.

**Cons**:
- `git log --grep` semantics: by default uses POSIX BRE (basic regex). `\b` word-boundary는 BRE에서 미정의. `--extended-regexp` 또는 `--perl-regexp` 플래그 필요.
- `--perl-regexp`는 git 빌드 시 PCRE 지원 필요 (CI runner는 OK 가능하지만 사용자 환경 호환성 우려).
- `--extended-regexp` 에서도 `\b`는 GNU grep 확장으로 표준 ERE에 없음 — 일부 OS의 git에서 동작 불일치 가능.
- 검증이 platform-dependent → CI에서만 검증하면 사용자 환경 false-negative 위험.

**Estimated risk**: Medium-High (POSIX 호환성).

### Approach B — Two-pass post-filter (Go-side word-boundary via ExtractSPECIDs)

**Idea**: 기존 `--grep=<SPEC-ID>` substring 매칭은 유지하되, walker scanner loop 안에서 각 commit title (그리고 필요시 commit body 추가 fetch) 에 대해 `ExtractSPECIDs(line)` 호출하여 정확한 SPEC-ID set을 추출. 그 set에 target specID가 포함되어 있을 때만 채택.

```go
for scanner.Scan() {
    line := scanner.Text()
    parts := strings.SplitN(line, " ", 2)
    if len(parts) < 2 {
        continue
    }
    commitTitle := parts[1]

    // Existing chore-skip filter (LSCSK-001)
    if shouldSkipCommitTitle(commitTitle) {
        continue
    }

    // NEW: word-boundary SPEC-ID filter (LSGF-001)
    extractedIDs := ExtractSPECIDs(commitTitle)
    if !containsString(extractedIDs, specID) {
        continue
    }

    _, status, err := ClassifyPRTitle(commitTitle)
    // ... rest unchanged
}
```

`ExtractSPECIDs` 정규식 (`SPEC-[A-Z0-9-]+-[0-9]+`, transitions.go:97) 은 정확한 SPEC-ID 토큰만 추출하므로 substring collision 자동 차단.

**Pros**:
- Platform-independent (Go regex, BRE/ERE 무관).
- 기존 검증된 `ExtractSPECIDs` 재사용 (외부 의존성 0).
- `git log --grep` 은 candidate set 축소용으로 유지 → 성능 무손실 (N=50 commit 가져온 후 Go-side filter).
- Defense-in-depth: chore-skip + word-boundary 두 필터 합성.

**Cons**:
- `git log --oneline` 은 commit subject만 반환. body에 SPEC-ID가 있고 subject에 없는 경우는 매칭되지 않음 → 이건 의도된 동작 (subject가 commit 의 primary scope).
- 만약 body 매칭이 필요해지면 `git log --format=%H %s%n%b` 형식 추가 fetch 필요 (현재 SPEC에서는 불필요).

**Estimated risk**: Low.

### Approach C — Combine A + B

**Idea**: `--grep="\\b...\\b"` (Approach A) + Go-side `ExtractSPECIDs` filter (Approach B).

**Pros**: Maximum robustness, defense-in-depth.
**Cons**: A의 POSIX 호환성 리스크가 그대로 잔존. B만으로 충분히 안전하므로 C의 추가 이득 낮음.

**Estimated risk**: A의 리스크 상속.

### Decision (plan-phase 잠정)

**채택**: **Approach B** (Two-pass post-filter).

**근거**:
1. **Platform-safe**: 모든 OS/git 빌드에서 Go regex로 동일 동작 보장. CI/local 환경 차이 무위험.
2. **No new dependency**: 기존 `ExtractSPECIDs` (transitions.go:97-116) 재사용. 외부 라이브러리 0.
3. **Performance**: N=50 commit × 1회 regex match ≈ <1ms. base latency ~30ms 대비 무시 가능.
4. **Composability**: `shouldSkipCommitTitle` (LSCSK-001) 와 자연 합성. chore-skip → word-boundary 순서 명확.
5. **Reversibility**: 단일 파일 한 함수 내 5-10줄 추가. rollback 용이.

**OQ (Open Questions, run-phase 결정)**:

- **OQ1**: plan-auditor 검토 결과 Approach A/C 가 더 적합하다면 재고. (default: B 유지)
- **OQ2**: `containsString` helper 신설 vs `slices.Contains` (Go 1.21+) 사용. Go 1.23+ 프로젝트이므로 `slices.Contains` 권장. (default: `slices.Contains`)
- **OQ3**: body 매칭 필요성. 본 SPEC scope에서는 subject-only 충분 — body까지 확장은 별도 follow-up SPEC. (default: subject-only)

---

## 3. Wave 1 — RED (bug reproduction + regression guard tests)

### 3.1 Objectives

- 현재 walker가 false-positive를 발생시키는 시나리오를 unit test로 명확하게 캡처 (RED).
- 회귀 가드 (chore-skip 무변경 확인) 베이스라인 확보.

### 3.2 Tasks

#### T1-001: 새 테스트 파일 신설 또는 기존 test 파일 확장

- **What**: `internal/spec/drift_specid_grep_test.go` 신설 (또는 `drift_test.go` 확장).
- **Tests**:
  1. `TestGetGitImpliedStatus_HARNESS001Resolution` — 실제 repo HEAD 기준 SPEC-V3R4-HARNESS-001 walker 호출. 현재 (pre-fix) `"planned"` 반환 → RED. fix 후 `"completed"` 반환 → GREEN.
  2. `TestGetGitImpliedStatus_SPECIDWordBoundary` — table-driven test. 5 cases (acceptance.md AC-LSGF-004 참조). pre-fix에서는 C2/C4/C5가 substring 매칭으로 잘못 통과 → RED.
- **Where**: `internal/spec/drift_specid_grep_test.go` (new file)
- **Acceptance**: 두 테스트 모두 pre-fix에서 FAIL 확인.

#### T1-002: chore-skip regression guard 식별

- **What**: 기존 `TestClassifyPRTitle_*` 또는 `TestShouldSkipCommitTitle_*` 테스트가 LSCSK-001 회귀를 가드하는지 확인. 없다면 명시적 회귀 가드 추가.
- **Where**: `internal/spec/transitions_test.go`, `internal/spec/drift_test.go`
- **Acceptance**: chore-skip 기능 가드 테스트 존재 + 본 SPEC fix 적용 후에도 GREEN.

### 3.3 Wave 1 Exit Criteria

- 2개 신규 테스트 PASS expectation 명확.
- pre-fix `go test ./internal/spec/... -run TestGetGitImpliedStatus_HARNESS001Resolution` → FAIL (RED 확인).
- pre-fix `go test ./internal/spec/... -run TestGetGitImpliedStatus_SPECIDWordBoundary` → FAIL (3+ sub-cases).

---

## 4. Wave 2 — GREEN (implementation)

### 4.1 Objectives

- `internal/spec/drift.go` `getGitImpliedStatus` walker 안에 word-boundary 필터 추가.
- Wave 1의 모든 테스트 GREEN.
- HARNESS-001/002/003 lint WARNING 0건.

### 4.2 Tasks

#### T2-001: word-boundary 필터 함수 신설

- **What**: helper 함수 신설 (예: `commitMatchesSPECID(commitTitle, specID string) bool`).
- **Logic**:
  ```go
  // commitMatchesSPECID는 commit title에 정확한 SPEC-ID 토큰이 포함되는지 확인한다.
  // ExtractSPECIDs를 사용해 substring collision (예: HARNESS-001 vs HARNESS-NAMESPACE-001) 차단.
  //
  // @MX:NOTE: [AUTO] commitMatchesSPECID — word-boundary SPEC-ID 필터
  // @MX:REASON: SPEC-V3R4-LINT-SPECID-GREP-FIX-001 — git log --grep substring 매칭이 NAMESPACE supersede commit을 walker first로 채택하던 결함 차단
  func commitMatchesSPECID(commitTitle, specID string) bool {
      extracted := ExtractSPECIDs(commitTitle)
      return slices.Contains(extracted, specID)
  }
  ```
- **Where**: `internal/spec/drift.go` (`shouldSkipCommitTitle` 직후 위치)
- **Acceptance**: helper unit test PASS.

#### T2-002: walker loop에 필터 호출 추가

- **What**: `getGitImpliedStatus` scanner loop에서 `shouldSkipCommitTitle` 직후 `commitMatchesSPECID(commitTitle, specID)` 호출. false면 continue.
- **Order rationale**: chore-skip이 cheaper (prefix check), word-boundary는 regex. cheaper-first 적용.
- **Where**: `internal/spec/drift.go:131-164` (scanner loop)
- **Code diff (예상)**:
  ```go
  for scanner.Scan() {
      line := scanner.Text()
      parts := strings.SplitN(line, " ", 2)
      if len(parts) < 2 {
          continue
      }
      commitTitle := parts[1]

      if shouldSkipCommitTitle(commitTitle) {
          continue
      }

      // NEW (LSGF-001): word-boundary SPEC-ID 필터 — substring collision 차단
      if !commitMatchesSPECID(commitTitle, specID) {
          continue
      }

      _, status, err := ClassifyPRTitle(commitTitle)
      // ... unchanged
  }
  ```
- **Acceptance**: Wave 1 RED tests → GREEN.

#### T2-003: AC-LSGF-001 End-to-End 검증

- **What**: `moai spec lint --strict` 실행. 3 WARNING → 0 WARNING 확인.
- **Where**: 사용자 환경 (CI + local)
- **Acceptance**: `0 error(s), 0 warning(s)`.

### 4.3 Wave 2 Exit Criteria

- AC-LSGF-001..005 모두 PASS.
- `go test ./internal/spec/... -race` GREEN.
- `golangci-lint run ./internal/spec/...` clean.

---

## 5. Wave 3 — REFACTOR (hygiene + MX tags + docs)

### 5.1 Objectives

- 코드 주석 / godoc / @MX 태그 정비.
- 회귀 방지를 위한 long-term anchor 설치.
- spec.md HISTORY 갱신 (run-phase 진입 시 v0.2.0 bump).

### 5.2 Tasks

#### T3-001: @MX 태그 갱신

- **What**: `getGitImpliedStatus` 함수 헤더 주석 갱신 — chore-skip + word-boundary 두 필터 명시.
  ```go
  // getGitImpliedStatus는 SPEC-ID에 대한 git log를 분석하여 lifecycle status를 추론한다.
  //
  // walker는 두 필터를 순차 적용한다:
  //   1. chore-skip 필터 (LSCSK-001): chore(spec): sweep commit 제외
  //   2. word-boundary 필터 (LSGF-001): substring collision (예: HARNESS-001 vs HARNESS-NAMESPACE-001) 차단
  //
  // @MX:ANCHOR: [AUTO] getGitImpliedStatus — git-implied status 추론 진입점
  // @MX:REASON: StatusGitConsistencyRule.Check + DetectDrift 두 곳에서 호출 (fan_in=2);
  //   LSCSK-001 + LSGF-001 두 결함 fix가 적용된 core walker
  ```
- **Where**: `internal/spec/drift.go:100-112`

#### T3-002: gitLogWindowSize MX 주석 보강

- **What**: 기존 `gitLogWindowSize` 주석에 LSGF-001 영향 추가.
  ```go
  // @MX:NOTE: [AUTO] N=50 결정 근거: SPEC당 평균 git log 매칭 commit 5-10건 + word-boundary 필터 후 expected match 1-3건.
  // @MX:REASON: SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 OQ1 (원본 결정) + SPEC-V3R4-LINT-SPECID-GREP-FIX-001 (word-boundary 필터 영향 평가).
  ```
- **Where**: `internal/spec/drift.go:96-99`

#### T3-003: 회귀 방지 anchor 테스트

- **What**: `TestGetGitImpliedStatus_SPECIDWordBoundary` 의 5 cases 외에 anchor case 1건 추가 — 본 SPEC SPEC-ID (`SPEC-V3R4-LINT-SPECID-GREP-FIX-001`) 의 walker 호출이 적절히 동작하는지.
- **Where**: `internal/spec/drift_specid_grep_test.go`

#### T3-004: lint sweep 검증

- **What**: `moai spec lint --strict` 재실행 + main 전체 SPECs scope. 0 WARNING 확인.
- **Acceptance**: AC-LSGF-001 final pass.

### 5.3 Wave 3 Exit Criteria

- @MX 태그 LSCSK-001 + LSGF-001 cross-reference 명시.
- lint sweep 0 WARNING.
- AC-LSGF-001~005 모두 final PASS.

---

## 6. Rollback Plan

본 변경이 main 머지 후 회귀를 유발할 경우:

1. **Immediate revert**: `git revert <run-PR squash commit>` → main에 직접 push (admin).
2. **Side effects**: walker가 substring 매칭으로 되돌아가 HARNESS-001/002/003 WARNING 3건 재발. 영향: lint output 노이즈만 (functional impact 0).
3. **No data loss**: 본 SPEC은 코드 fix만이며 frontmatter/SPEC 파일은 미수정. revert는 무손실.
4. **Re-plan**: revert 후 본 SPEC plan.md 의 OQ1/OQ2/OQ3 재검토 → Approach A/C 재고.

---

## 7. Open Questions (Run-phase 결정)

- **OQ1** (Approach 재고): plan-auditor 검토 결과 Approach A (regex grep) 가 더 적합하다고 판단되면 재선택. 단, POSIX 호환성 리스크 mitigation 명시 필요.
  - **Default**: Approach B 유지.

- **OQ2** (helper 형식): `slices.Contains` (Go 1.21+) vs custom `containsString` helper.
  - **Default**: `slices.Contains` (Go 1.23+ 프로젝트 기준).

- **OQ3** (subject vs body 매칭): body 매칭 확장 필요성. 현재 SPEC scope는 subject-only.
  - **Default**: subject-only 유지. body 매칭은 별도 follow-up SPEC 후보.

- **OQ4** (필터 순서): chore-skip → word-boundary vs word-boundary → chore-skip.
  - **Default**: chore-skip 먼저 (cheaper-first).

- **OQ5** (anchor test): 본 SPEC SPEC-ID 자체에 대한 walker 호출이 plan-phase에서 동작하는지 — plan-PR 머지 후 RED일 가능성 (status: planned ≠ frontmatter: draft).
  - **Default**: anchor test는 sync-PR 머지 후에만 적용 (lint scope에서 본 SPEC은 일시적 WARNING 허용).

---

## 8. Telemetry Expectations

| Metric | Pre-fix baseline | Post-fix target |
|--------|------------------|-----------------|
| `moai spec lint --strict` WARNING 수 (HARNESS-{001,002,003} scope) | 3 | 0 |
| `moai spec lint --strict` 전체 main scope WARNING 수 | 3 | 0 (가정: 다른 substring collision 없음) |
| walker call latency (single SPEC, N=50) | ~30ms | ~31ms (+1ms regex overhead, 무시 가능) |
| `go test ./internal/spec/... -race` 실행 시간 | 변동 | +~50ms (2 신규 테스트) |
| diff size | n/a | drift.go +10-15 LOC + test file +50-80 LOC |
| Files modified | n/a | 2 (drift.go + drift_specid_grep_test.go) |

---

## 9. Risk Register

| ID | Risk | Likelihood | Impact | Mitigation |
|----|------|-----------|--------|-----------|
| R1 | `ExtractSPECIDs` 정규식이 예상치 못한 SPEC-ID 형식 (legacy) 을 누락 | Low | Medium | transitions.go:97 정규식은 historical SPEC ID도 매칭하도록 의도적 permissive. existing 197 SPECs sample test로 검증. |
| R2 | git log subject만 검사 → body에 genuine signal이 있는 경우 false-negative | Low | Low | 현재 본 SPEC 시나리오에서 body-only signal commit은 발견되지 않음. follow-up SPEC 후보로 기록. |
| R3 | Approach B의 walker가 word-boundary 필터로 모든 N=50 commit을 reject → LSCSK-001 fail-safe path로 빠짐 | Low | Low | EC-003 처리 동일. lint rule이 skip 처리하여 false-positive 0 보장. |
| R4 | plan-auditor가 본 plan을 0.85 미만으로 평가 | Medium | Medium | OQ1~OQ5 자율 재검토. 필요시 iteration 2 revise. |
| R5 | 본 SPEC fix 후 다른 SPEC들에서 새 WARNING 발생 (가려져 있던 진짜 drift 노출) | Low | Medium | lint sweep 0 가정. 만약 발생 시 별도 follow-up SPEC. |

---

## 10. Plan-Run-Sync Cadence

- **Plan PR** (current): `feat/SPEC-V3R4-LINT-SPECID-GREP-FIX-001-plan` → main, squash, auto-merge.
  - Title: `plan(spec): SPEC-V3R4-LINT-SPECID-GREP-FIX-001 — walker SPEC-ID grep precision fix`
- **Run PR** (post plan-merge): `feat/SPEC-V3R4-LINT-SPECID-GREP-FIX-001` → main, squash, auto-merge.
  - Title: `feat(spec): SPEC-V3R4-LINT-SPECID-GREP-FIX-001 — drift.go word-boundary filter (LSGF-001 GREEN)`
- **Sync PR**: `sync/SPEC-V3R4-LINT-SPECID-GREP-FIX-001` → main, squash, auto-merge.
  - Title: `sync(SPEC-V3R4-LINT-SPECID-GREP-FIX-001): status transition draft → completed + HISTORY 0.3.0`

---

## 11. References

- `internal/spec/drift.go:96-197` — walker code (LSCSK-001 baseline)
- `internal/spec/transitions.go:97-116` — `ExtractSPECIDs` (재사용 대상)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — SSOT (frontmatter 정렬용)
- `.claude/rules/moai/workflow/spec-workflow.md` — SPEC phase discipline
- CLAUDE.local.md §18.12 — BODP (plan-in-main 결정 근거)
- SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 — chore-skip 필터 원본 SPEC
- SPEC-V3R4-SPECLINT-DEBT-002 — 12-field canonical SSOT 원본 SPEC
- Lessons #16 (sync-prefix-correction) — 본 SPEC fix의 학습 체인 종착점
