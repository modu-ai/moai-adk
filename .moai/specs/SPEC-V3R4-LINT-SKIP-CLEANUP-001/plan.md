# Plan — SPEC-V3R4-LINT-SKIP-CLEANUP-001

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.1   | 2026-05-16 | plan-audit remediation | plan-audit 0.904 PASS 후 P2 4건 remediation: (1) AC-LSKC-002 placeholder → plan.md §5.2 cross-ref, (2) HISTORY date harmonize 2026-05-16, (3) REQ-005↔AC-002 매핑 rationale 명시, (4) design.md §2.4 redundancy 정리. |
| 0.1.0   | 2026-05-16 | manager-spec | 초기 plan. 55개 SPEC frontmatter의 `lint.skip: [StatusGitConsistency]` 일괄 제거 전략 + 4개 milestone (BASELINE / BULK EDIT / VERIFICATION / RUN-PR). bulk edit 도구 선택지 3가지 비교 + idempotent 검증 절차. plan-in-main 표준 적용 — worktree 없음. |

---

## 1. Implementation Strategy

### 1.1 핵심 접근

본 cleanup은 **frontmatter-only bulk edit** 이다. 55개 SPEC의 `spec.md` 파일에서 YAML frontmatter 영역만 정확히 수정하고 본문은 byte-level로 보존해야 한다.

전략 단계:

1. **Baseline capture**: 55개 SPEC의 현재 상태 (body sha256, lint --strict 출력) 를 capture
2. **Sequential edit**: 55개 SPEC에 대해 sequential하게 다음 4가지를 수행
   - frontmatter `lint:` 블록 전체 제거 (현재 모든 SPEC이 단일 엔트리 케이스)
   - frontmatter `version` patch bump
   - frontmatter `updated: 2026-05-15`
   - HISTORY 표에 새 row 1줄 추가
3. **Post-edit verification**: body sha256 비교 + `moai spec lint --strict` WARN 0 검증
4. **PR commit**: 단일 commit (55개 SPEC modified) + PR push

### 1.2 도구 선택지 (run-phase 결정)

| 옵션 | 장점 | 단점 | 권장 |
|------|------|------|------|
| (A) 순수 Edit tool calls (55 × 4 = ~220 edit ops) | 가장 단순, line-level 정밀 | edit op 수 많음, Edit tool overhead | △ |
| (B) `sed`/`awk` shell script | 빠름, idempotent 보장 어려움 | YAML 들여쓰기 깨질 위험, line ending OS 의존 | △ |
| (C) Go 헬퍼 스크립트 (in `.moai/scripts/lint-skip-cleanup.go`, `go run`) | YAML 라이브러리로 안전, idempotent 자연스럽게 보장, 재사용성 | LOC 100~150 새 코드, 검토 필요 | **○ 권장** |

권장: **옵션 C (Go script)** — `gopkg.in/yaml.v3`로 YAML 부분만 파싱/수정하고 body는 byte-for-byte 보존. script 자체는 PR에 포함하지 않고 cleanup commit only push 후 폐기, 또는 `.moai/scripts/lint-skip-cleanup.go`에 보존하여 향후 유사 cleanup에 재사용 (OQ2 참조).

대안 (옵션 A polyfill): Go script 작성 부담을 피하고 싶으면 Edit tool로 55개 SPEC 순회 — manager-develop이 매 SPEC마다 `Read` → 4개 `Edit` op 수행 (개당 ~5 op × 55 = 275 op). 시간은 더 걸리지만 review 가시성 높음.

### 1.3 Sequential vs Parallel

[HARD] **Sequential only**. 이유:

- `moai spec lint --strict` baseline 캡처 시 git working tree 상태 의존
- 55개 SPEC을 병렬 수정하면 lint 검증을 partial state에서 수행하게 됨
- Single PR이 목표이므로 commit ordering 일관성 필요

---

## 2. Milestones

priority-based ordering (시간 추정 금지 per `agent-common-protocol.md` §Time Estimation):

### M1 — BASELINE

**Priority: High** (모든 후속 단계의 기준선)

- 55개 영향 SPEC list 확정 (research.md §2의 list 검증 — frontmatter strict scan)
- 각 SPEC body sha256 capture → `.moai/state/lint-skip-cleanup-baseline.json` 산출 (스크립트 사용 시)
- 현재 `moai spec lint --strict` 출력에서 `StatusGitConsistency` count 캡처 (예상: PR #933 머지 후이므로 이미 0건 — walker filter 작동 확인용)
- 각 SPEC 현재 `version` 필드 캡처

**완료 조건:**
- baseline.json 산출
- lint --strict StatusGitConsistency 카운트 기록

### M2 — BULK EDIT

**Priority: High** (메인 작업)

55개 SPEC 각각에 대해 sequential 수행:

1. frontmatter `lint:` 블록 전체 제거 (`lint:`, `  skip:`, `    - StatusGitConsistency` 3줄 — 모든 케이스가 단일 엔트리이므로 블록 전체 제거)
2. frontmatter `version`: 현재값 patch +1 (예: `"0.3.0"` → `"0.3.1"`)
3. frontmatter `updated`: 현재값 → `2026-05-15`
4. HISTORY 표에 새 row 1개 prepend (또는 append — baseline ordering 보존)
   - 표준 row 형식: `| <new-version> | 2026-05-15 | manager-develop (run-phase) | lint.skip StatusGitConsistency 회피책 제거 — SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 walker filter 머지로 불필요해짐. |`

**완료 조건:**
- `git status` 결과 55개 SPEC `spec.md` modified 표시
- 각 SPEC `lint:` 블록 부재 확인 (`grep -L "^lint:" $files` 결과 모든 55 SPEC 포함)

### M3 — VERIFICATION

**Priority: High** (회귀 방지)

1. Body sha256 비교 — 55개 SPEC 모두 baseline과 일치 (HISTORY row 1줄 추가는 frontmatter 인접 영역으로 간주, 본문 sha256 산정에서 제외)
2. `moai spec lint --strict` 실행 — `StatusGitConsistency` WARN 0 확인
3. `git diff .moai/specs/` 결과 55개 SPEC 외 파일 0개
4. 각 SPEC frontmatter parse 검증 (`moai spec list` 정상 동작 — frontmatter 깨짐 회귀 방지)

**완료 조건:**
- lint --strict 통과
- diff stats: 55 files modified, 본문 변경 라인 0 (HISTORY 1줄 + frontmatter 변경만)

### M4 — RUN-PR

**Priority: Medium**

1. commit message: `chore(spec): SPEC-V3R4-LINT-SKIP-CLEANUP-001 — 55 SPECs lint.skip cleanup`
2. push to feature branch `feat/SPEC-V3R4-LINT-SKIP-CLEANUP-001` (Enhanced GitHub Flow §18.2)
3. PR 생성: title `feat(spec): SPEC-V3R4-LINT-SKIP-CLEANUP-001 — 55 SPECs lint.skip workaround removed`
4. PR body: SPEC link + AC summary + lint --strict before/after 캡처
5. Wait for CI (Lint + Test + Build all green)
6. admin squash merge 후 worktree cleanup 불필요 (worktree 미사용)

**완료 조건:**
- PR MERGED to main
- main HEAD `moai spec lint --strict` WARN 0 (StatusGitConsistency 카테고리)

---

## 3. Files to Modify

### 3.1 변경 대상 (frontmatter only)

55개 SPEC `spec.md` 파일. 각 파일 평균 변경 ~5-8 lines (`lint:` 블록 3줄 제거 + version 1줄 + updated 1줄 + HISTORY 1줄 추가).

총 변경 추정:
- 제거: 3 lines × 55 = 165 lines
- 추가: 1 line × 55 = 55 lines (HISTORY row)
- 수정: 2 lines × 55 = 110 lines (version, updated)
- Net diff: ~330 lines across 55 files

(research.md §2의 55-SPEC list 참조)

### 3.2 변경 없음 보장 (out-of-scope safety net)

- `internal/spec/*.go` — 0 lines
- `.github/workflows/*.yml` — 0 lines
- `docs-site/**` — 0 files
- `README.md` / `README.ko.md` / `CHANGELOG.md` — 0 lines (CHANGELOG는 sync-phase 범위)
- 55개 외 `.moai/specs/*/spec.md` — 0 lines

### 3.3 (선택) bulk edit script

`.moai/scripts/lint-skip-cleanup.go` (또는 `.sh`) — OQ2 결정에 따라 PR에 포함 / 폐기 결정. PR 포함 시 별도 1 file added.

---

## 4. Technical Approach

### 4.1 frontmatter parse 전략 (옵션 C 권장)

```go
// pseudocode — run-phase가 구체화
package main

import (
    "gopkg.in/yaml.v3"
    "os"
    "strings"
)

func cleanup(specPath string) error {
    data, _ := os.ReadFile(specPath)
    parts := strings.SplitN(string(data), "\n---\n", 2)
    if len(parts) < 2 {
        return fmt.Errorf("no frontmatter: %s", specPath)
    }
    fm := parts[0]  // "---\n<yaml>"
    body := "---\n" + parts[1]  // body 보존
    
    // YAML parse
    var node yaml.Node
    yaml.Unmarshal([]byte(strings.TrimPrefix(fm, "---\n")), &node)
    
    // lint.skip에서 StatusGitConsistency 제거
    // lint.skip이 빈 배열이 되면 lint: 블록 전체 제거
    // version patch bump
    // updated = "2026-05-15"
    
    // serialize 시 key ordering 보존 (yaml.v3 Node API)
    // HISTORY 표는 frontmatter 외부이므로 별도 markdown 처리
    
    return os.WriteFile(specPath, ...)
}
```

핵심 challenges:
1. **Key ordering 보존**: `gopkg.in/yaml.v3`의 `yaml.Node` API 사용 필수 (`Unmarshal → map → Marshal` 사용 시 key ordering 손실)
2. **Multi-line string preservation**: title, related_theme 등 따옴표 / unquoted 스타일 보존
3. **HISTORY 표 편집**: markdown table 행 추가 — frontmatter 외부 처리

### 4.2 frontmatter parse 전략 (옵션 A polyfill — Edit tool)

옵션 C 부담이 크면 Edit tool 직접 사용:

```
For each SPEC:
  1. Read spec.md (frontmatter 영역)
  2. Edit: 'lint:\n  skip:\n    - StatusGitConsistency\n' → '' (3-line block 제거)
  3. Edit: 'version: <old>' → 'version: <new>' (patch bump)
  4. Edit: 'updated: <old>' → 'updated: 2026-05-15'
  5. Edit: HISTORY 표 첫 row 위 또는 마지막 row 아래에 새 row 1줄 insert
```

이 경우 idempotency는 manager-develop이 매번 baseline 비교로 수동 보장.

### 4.3 idempotency 검증

bulk script 또는 manual edit 모두:
- 1차 실행 후 `git diff` 캡처 → 220+ lines
- 2차 실행 (script re-run 또는 manual re-walk) 후 `git diff` 캡처 → 0 lines
- 2차 실행이 no-op임을 확인

---

## 5. Testing Strategy

본 SPEC은 frontmatter metadata cleanup이므로 **Go unit test 작성 불요**. 검증은 다음 외부 도구로 수행:

### 5.1 Pre-edit baseline test

```bash
# Body content sha256 capture
for spec in $(cat affected-list.txt); do
  awk '/^---$/{f++} f>=2' "$spec" | sha256sum > baselines/"$(basename $(dirname $spec)).sha256"
done

# lint --strict StatusGitConsistency count
moai spec lint --strict 2>&1 | grep -c "StatusGitConsistency" > baseline-warns.txt
```

### 5.2 Post-edit verification

```bash
# Body sha256 unchanged
for spec in $(cat affected-list.txt); do
  current=$(awk '/^---$/{f++} f>=2' "$spec" | sha256sum)
  baseline=$(cat baselines/"$(basename $(dirname $spec)).sha256")
  [ "$current" = "$baseline" ] || echo "DRIFT: $spec"
done

# lint --strict — population-scoped (55 cleanup-target SPECs)
# cleanup 외 SPECs의 기존 drift WARN (2026-05-15 실측 8건)은 본 SPEC scope 외
moai spec lint --strict 2>&1 | grep "StatusGitConsistency" | \
  grep -E "$(cat affected-list.txt | tr '\n' '|' | sed 's/|$//')" | \
  wc -l > /tmp/post-warns.txt
diff /tmp/post-warns.txt baseline-warns.txt && echo "PASS: delta=0" || echo "FAIL"

# Affected files only
git diff --name-only .moai/specs/ | wc -l  # expect 55 (or 55 + 1 if script committed)
```

### 5.3 Idempotency test (script-only)

```bash
go run .moai/scripts/lint-skip-cleanup.go  # first run
git diff --stat  # capture
go run .moai/scripts/lint-skip-cleanup.go  # second run
git diff --stat  # expect identical to first run (no additional changes)
```

### 5.4 Regression guard

`SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001`의 run-phase에서 추가된 unit tests (`internal/spec/drift_test.go` 의 `TestGetGitImpliedStatus_SkipSweepCommit*`) 는 본 SPEC과 무관 — walker filter 자체가 변경 대상이 아니므로 별도 regression test 추가 없음.

---

## 6. Definition of Done

- [ ] **AC-LSKC-001 ~ 005 모두 GREEN** (M3 VERIFICATION 통과)
- [ ] **PR MERGED** to main (Enhanced GitHub Flow §18.3 squash merge)
- [ ] **`moai spec lint --strict`** WARN 0 (StatusGitConsistency 카테고리, main HEAD 기준)
- [ ] **HISTORY new row 1줄** 각 55 SPEC에 추가됨
- [ ] **version patch bump** 각 55 SPEC에 적용됨
- [ ] **updated: 2026-05-15** 각 55 SPEC에 적용됨
- [ ] **CHANGELOG `[Unreleased]` row 1줄** 추가 (sync-phase 범위 — manager-docs가 처리)
- [ ] **plan-auditor PASS** (≥ 0.85)
- [ ] **MX tags**: `@MX:NOTE` (이력 + 회피책 제거 의도) — 코드 수정 없으므로 `@MX:ANCHOR` / `@MX:WARN` 불필요
- [ ] **frontmatter status 전이**: `draft → in-progress → implemented → completed` (각 phase entry 시점에 manager-* agent가 갱신)
- [ ] **(Optional) bulk edit script 보존 결정**: OQ2 결정 따라 `.moai/scripts/lint-skip-cleanup.{go,sh}` 보존 또는 폐기

---

## 7. Open Questions

### OQ1 — lint.skip의 다른 rule entries 보존

**Question**: 본 SPEC scoping 시점에 검증한 결과 55 SPEC 모두 `lint.skip`이 `StatusGitConsistency` 단일 엔트리 케이스다. 그러나 향후 다른 lint rule entries 추가될 경우 정책이 필요한가?

**Default Decision**: REQ-LSKC-002, REQ-LSKC-006 으로 정책 명시 — 다른 entries 보존, `StatusGitConsistency`만 element-level 제거. 현재 검증 대상 SPEC 없음 (forward-compat).

**Action Required**: 없음 — design.md §6 Edge Cases (a), (b) 에서 처리.

### OQ2 — bulk edit 도구 선택 (Go script vs Edit tool vs shell)

**Question**: §1.2의 3개 옵션 중 어느 것을 run-phase가 채택할 것인가?

**Default Decision**: 옵션 C (Go script) 권장. run-phase가 다음 기준으로 최종 결정:
- 옵션 C: 재사용 가능성 (`.moai/scripts/` 보존 가치) + idempotency 자연스러움
- 옵션 A: 시간 절약 (스크립트 작성 LOC 0) + Edit op 가시성
- 옵션 B (shell): 거부 — YAML 들여쓰기 안정성 부족

**Action Required**: run-phase 시작 시 manager-develop이 결정. 결정 결과 design.md §4에 기록.

### OQ3 — version bump 정책 (patch vs minor)

**Question**: `version` 필드를 patch (Z+1) vs minor (Y+1) 중 어느 단위로 bump?

**Default Decision**: patch (Z+1). 근거:
- metadata-only 변경 (semver "non-functional")
- 본문 0줄 수정 — minor bump 부적절
- 영향 범위 좁음 — lint engine 자체의 동작은 변경 없음

**Action Required**: 없음 — REQ-LSKC-003 + AC-LSKC-005 + 정책 §1.2 (C5) 으로 확정.

---

## 8. Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| R1: 55 SPEC 중 일부에 frontmatter parser drift (예: title escape) → bulk script crash | Low | 옵션 C 채택 시 `yaml.Node` API 로 안전. crash 시 sequential mode로 fallback. |
| R2: HISTORY 표 형식 SPEC마다 미세 차이 (열 개수, 들여쓰기 4가지 변형) | Low | sample 5-10개 SPEC pre-check → manager-develop이 분기 처리 |
| R3: idempotency 깨짐 (script 2회 실행 시 추가 HISTORY row 발생) | Medium | OQ2 옵션 C 선택 시 "현재 HISTORY top row가 본 cleanup row인지" 사전 검사 후 skip |
| R4: lint --strict 검증이 walker filter 미적용 binary로 수행됨 (`moai` PATH 우선순위 문제) | Medium | M1에서 `which moai` 캡처 + `moai version` 확인. walker filter는 main `9e394e51b` 이후 빌드여야 함 |
| R5: PR CI 실패 (Test, Build, Lint 중 1+) | Low | 본 cleanup은 Go 코드 무변경이므로 Lint/Test/Build 전부 unaffected. spec-lint CI job만 검증 대상. |
| R6: 55개 SPEC 외 의도하지 않은 파일 수정 | Low | M3 verification 단계의 `git diff --name-only` 검증 + plan-auditor도 cross-check |

---

## 9. Dependencies

- **Predecessor (HARD)**: `SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001` (PR #933 머지 + PR #934 sync 완료) — walker filter가 main에 있어야 함. 현재 main `9e394e51b` 이미 충족.
- **None (Successor)**: 본 SPEC 머지 후 후속 SPEC 없음 — 일회성 cleanup.

---

## 10. Branch & PR Strategy

- **Branch name**: `feat/SPEC-V3R4-LINT-SKIP-CLEANUP-001` (Enhanced GitHub Flow §18.2 — feat prefix; cleanup은 chore도 가능하나 SPEC 기반 변경이므로 feat 표준)
- **Base**: `origin/main` (BODP A=¬ B=¬ C=¬ → main 결정)
- **Merge strategy**: squash (Enhanced GitHub Flow §18.3 — feat → main은 squash)
- **Worktree**: 미사용 (per `feedback_worktree_never_use` 정책)
- **PR title**: `feat(spec): SPEC-V3R4-LINT-SKIP-CLEANUP-001 — 55 SPECs lint.skip workaround removed`
- **Commit message**: `chore(spec): SPEC-V3R4-LINT-SKIP-CLEANUP-001 — 55 SPECs lint.skip cleanup` (Conventional Commits)
- **Reviewers**: 자동 — Release Drafter autolabeler가 `type:chore` + `area:spec-lint` 부착
