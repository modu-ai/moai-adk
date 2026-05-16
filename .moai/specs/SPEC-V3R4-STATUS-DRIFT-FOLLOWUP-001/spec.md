---
id: SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001
version: "0.1.0"
status: in-progress
created_at: 2026-05-16
updated_at: 2026-05-16
author: GOOS행님
priority: P1
labels: [v3r4, lint, spec-frontmatter, status-drift, plan-in-main]
issue_number: null
---

# SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 — 64 SPEC frontmatter status 일괄 동기화 (LSKC-001 cleanup으로 노출된 real drift 해소)

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-16 | GOOS행님 / manager-spec | 초기 draft. SPEC-V3R4-LINT-SKIP-CLEANUP-001 (PR #937, main `758341089`) 머지 직후 lint.skip 회피책이 mask하던 real status drift 64건이 노출됨. walker filter (PR #933)는 sweep commit (`chore(spec):`) 만 skip하므로 frontmatter status가 git-implied status보다 ahead/behind인 real mismatch 64건 잔존. 사용자 명시 옵션 (a) status field bulk synchronization 단일 scope. 옵션 (b) walker filter 확장 / (c) rule deprecation 모두 OUT OF SCOPE. 8 패턴 (A: 50건, B: 4건, C: 6건, D: 1건, E: 1건, F: 1건, G: 1건, H: 3건) 으로 분류 + 5 wave 처리 전략. plan-in-main + worktree 미사용 정책 (per `feedback_worktree_never_use`) 적용. BODP 평가: A=¬ B=¬ C=¬ → main @ origin/main. |

---

## 1. Goal

`SPEC-V3R4-LINT-SKIP-CLEANUP-001` (PR #937, `758341089` main 머지) 의 실측 결과 lint.skip 회피책 제거가 walker filter 범위를 벗어나는 real status drift를 노출시켜 main에서 `moai spec lint --strict | grep -c StatusGitConsistency` 가 **64 WARNING** 을 보고한다. 본 SPEC은 64건 모두를 frontmatter `status` 필드 일괄 동기화로 해소하여 `0 ERROR + 0 WARNING` 상태를 달성한다.

### 1.1 Binary Success Criterion (AC-SDF-001)

본 SPEC 머지 후 main에서 다음 명령이 정확히 0을 출력해야 한다:

```bash
moai spec lint --strict 2>&1 | grep -c "StatusGitConsistency"
# expected output: 0
```

### 1.2 Cleanup Chain 종결

본 SPEC은 4-단계 lint cleanup chain의 마지막 wave다:

```
SPECLINT-DEBT-001 (sweep + skip 51건 등록)
  └─ LINT-STATUS-CHORE-SKIP-001 (engine 개선, walker filter, 7건 false-positive 해소)
       └─ LINT-SKIP-CLEANUP-001 (51건 skip 제거, real drift 64건 노출)
            └─ STATUS-DRIFT-FOLLOWUP-001 (본 SPEC; 64건 status 동기화로 chain 완결)
```

### 1.3 8-Pattern 분포 요약

| Pattern | frontmatter → git_implied | Count | 처리 방향 |
|---------|--------------------------|-------|----------|
| A | completed → implemented | 50 | bulk script: status downgrade (`completed → implemented`) |
| B | completed → in-progress | 4 | 개별 verification + status downgrade |
| C | implemented → in-progress | 6 | 개별 verification + status downgrade |
| D | superseded → completed | 1 | detector exemption (`superseded`/`archived` terminal state pass-through) |
| E | superseded → implemented | 1 | 동상 (D와 동일 detector exemption) |
| F | archived → implemented | 1 | 동상 |
| G | archived → in-progress | 1 | 동상 (D detector exemption으로 자동 해결) |
| H | self-drift (cleanup chain) | 3 | 본 SPEC sync-phase 시점에 재귀 cleanup |

**합계: 64건** (50 + 4 + 6 + 1 + 1 + 1 + 1 + 3). 패턴별 자세한 SPEC ID 표는 research.md §10 Appendix 참조.

---

## 2. Background & Problem Statement

### 2.1 Cleanup Chain 컨텍스트

- 2026-05-15 SPEC-V3R4-SPECLINT-DEBT-001 sync (PR #921) 가 lint ERROR 66 + WARNING 140을 0으로 일괄 reset하면서 90 SPEC frontmatter status를 sweep + 51 SPEC에 `lint.skip: [StatusGitConsistency]` 회피책 등록.
- 2026-05-15 SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 (PR #933 run + #934 sync) 이 sweep commit 자기참조 bootstrap bug를 walker filter (`chore(spec):` skip) 로 해소.
- 2026-05-16 SPEC-V3R4-LINT-SKIP-CLEANUP-001 (PR #937 squash) 이 55 SPECs lint.skip 회피책을 영구 제거.

### 2.2 Real Drift 노출

LSKC-001 run-phase 실측에서 lint.skip suppression이 mask하던 진짜 mismatch가 노출됨:

- cleanup population 55 SPEC 중 **54건이 real drift 보유** (lint.skip이 회피)
- main-wide 추가 **약 10건** real drift 발견 → 총 64 WARNING
- 이는 회귀 아닌 expected outcome (LSKC-001 AC-LSKC-002 재정의 시점에 이미 follow-up SPEC 발급 합의됨)

### 2.3 walker filter 범위 한계

`internal/spec/drift.go::shouldSkipCommitTitle` (PR #933) 는 `chore(spec):` / `chore(specs):` prefix 만 skip한다. 그러나 64 WARN의 검출 사유는 다음과 같이 walker 의 정상 동작 범위:

- **Pattern A (50건)**: latest non-skip commit이 `feat(SPEC-XXX):` → git-implied `implemented`. frontmatter `completed` 와 mismatch. (sync chore commit이 `chore(spec):` 으로 작성되어 walker skip → impl commit이 다음 hit)
- **Pattern B/C (10건)**: SPEC ID grep 매칭 commit에 `feat:` 자체가 없음 (다른 SPEC 우산 아래 implementation). 결과적으로 `chore`/`ci`/`docs` 등 run-partial prefix만 hit → git-implied `in-progress`.
- **Pattern D/E/F/G (4건)**: SPEC이 `superseded` 또는 `archived` terminal state이지만 자체 git history는 `implemented`/`completed` 단계까지 진행됨. detector가 terminal state 의미를 모름 → false-positive.
- **Pattern H (3건)**: cleanup chain SPEC 자체의 sync chore commit이 `chore(spec):` skip 처리 + impl commit 부재 → git-implied 가 잘못된 단계.

### 2.4 사용자 옵션 결정 (locked)

사용자 명시: 옵션 (a) **Status field bulk synchronization**. 다른 두 옵션은 explicitly out of scope:

- ❌ 옵션 (b): walker filter scope 확장 (`chore:`, `ci:`, `docs:` 추가 skip) — false-negative 위험, lint detector 의미론 약화
- ❌ 옵션 (c): StatusGitConsistency rule deprecation — 검사 무력화, SPECLINT-DEBT-001 정신 위배

본 SPEC 은 옵션 (a) 단일 경로. 단, **Pattern D/E/F/G** 의 4건은 frontmatter 변경이 의미 손실을 야기 (`superseded → completed` 는 lifecycle 정보 파괴) 하므로 detector exemption 코드 변경으로 처리한다 — 이는 옵션 (a) 의 `bulk synchronization` 정의 안에서 "synchronization 방향이 detector 측인 케이스"로 해석.

### 2.5 LSKC-001 Bulk Script 자산 활용

LSKC-001 PR #937 에 `.moai/scripts/lint-skip-cleanup.go` (~220 LOC, `gopkg.in/yaml.v3` Node API 기반) 가 보존됨. 본 SPEC은 이 script를 **fork**하여 `.moai/scripts/status-drift-cleanup.go` 작성:

- 동일 IO 헬퍼 (frontmatter parse, body sha256 보존, HISTORY append) 재활용
- Wave별 dispatch (Pattern A bulk + Pattern B/C verification + Pattern H 재귀)
- idempotent 보장 (LSKC-001 정신 계승)

---

## 3. Requirements (EARS)

### 3.1 Ubiquitous Requirements

**REQ-SDF-001** (Ubiquitous): The system shall preserve all `superseded` and `archived` SPEC frontmatter `status` values without modification, even when they mismatch git-implied status (terminal lifecycle state intent must be preserved).

**REQ-SDF-002** (Ubiquitous): The system shall preserve every affected SPEC's body content (everything below the closing `---` of YAML frontmatter) byte-for-byte, except for one new HISTORY table row recording the synchronization action.

**REQ-SDF-003** (Ubiquitous): The system shall bump each affected SPEC's `version` field by a patch increment (semver Z+1) and update `updated_at` to the synchronization commit date when applying any frontmatter change.

### 3.2 Event-Driven Requirements

**REQ-SDF-004** (Event): When a SPEC matches Pattern A (`frontmatter completed` + `git-implied implemented`), the system shall downgrade frontmatter `status` from `completed` to `implemented` via the bulk script.

**REQ-SDF-005** (Event): When a SPEC matches Pattern B (`frontmatter completed` + `git-implied in-progress`), the system shall first conduct per-SPEC verification (read SPEC body, search related PR history) and then either (a) downgrade frontmatter `status` to `in-progress` OR (b) document why the SPEC remains `completed` despite git evidence.

**REQ-SDF-006** (Event): When a SPEC matches Pattern C (`frontmatter implemented` + `git-implied in-progress`), the system shall apply the same per-SPEC verification flow as REQ-SDF-005 with target downgrade `implemented → in-progress`.

**REQ-SDF-007** (Event): When `moai spec lint --strict` is invoked after the synchronization is complete, the system shall report exactly 0 `StatusGitConsistency` warnings against the post-cleanup main HEAD.

### 3.3 State-Driven Requirements

**REQ-SDF-008** (State): While the affected SPEC's frontmatter `status` is `superseded` or `archived` (Pattern D/E/F/G), the system shall apply a detector-side exemption in `internal/spec/lint.go::StatusGitConsistencyRule::Check` (or `internal/spec/lint/checks/status_git_consistency.go`) that returns no finding when the frontmatter `status` is in the terminal lifecycle enum set.

**REQ-SDF-009** (State): While Wave 5 (Pattern H — self-drift cleanup) is being applied during the sync-phase, the system shall recursively re-run the bulk script against the cleanup-chain SPECs (SPEC-V3R4-LINT-SKIP-CLEANUP-001, SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001, SPEC-V3R4-SPECLINT-DEBT-001, and the present SPEC SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001) to absorb any newly introduced self-drift.

### 3.4 Unwanted Behavior Requirements

**REQ-SDF-010** (Unwanted): The system shall not introduce any new `lint.skip` entry of any kind into any SPEC frontmatter (the `lint.skip` ban established by SPEC-V3R4-LINT-SKIP-CLEANUP-001 is permanent).

**REQ-SDF-011** (Unwanted): The system shall not modify the walker filter scope in `internal/spec/drift.go::shouldSkipCommitTitle` (no addition of `chore:`, `ci:`, `docs:`, `style:` or any other prefix beyond the existing `chore(spec):` / `chore(specs):` set).

**REQ-SDF-012** (Unwanted): The system shall not deprecate, disable, or weaken the `StatusGitConsistencyRule` itself (the rule remains active and reporting; only Pattern D/E/F/G receives a narrow terminal-state exemption inside the check).

**REQ-SDF-013** (Unwanted): The system shall not modify any SPEC body (REQ/AC/HISTORY beyond a single new row, design.md, plan.md, acceptance.md) for the 64 affected SPECs.

**REQ-SDF-014** (Unwanted): The system shall not modify any SPEC outside the 64-affected list during the synchronization (other SPECs and the cleanup-chain SPECs not listed in Pattern H must remain untouched).

### 3.5 Optional Requirements

**REQ-SDF-015** (Optional): Where the run-phase agent chooses to externalize the bulk synchronization logic as a reusable Go script (`.moai/scripts/status-drift-cleanup.go`), the system shall record the script's idempotency property — running it twice produces an identical result on the second invocation (no-op).

**REQ-SDF-016** (Optional): Where additional cleanup waves are needed (e.g., new SPECs created between plan-phase and run-phase introduce additional drift), the system shall extend the affected-list files (`affected-list-pattern-A.txt`, etc.) without renumbering existing entries.

---

## 4. Acceptance Criteria

### AC-SDF-001 — Binary Lint Outcome (Primary Success)

After the synchronization is merged into main, `moai spec lint --strict 2>&1 | grep -c "StatusGitConsistency"` MUST output exactly `0`.

검증:
- main HEAD에서 `moai spec lint --strict` 실행 → `StatusGitConsistency` 카테고리 WARN 0건
- 다른 lint rule WARN/ERROR 는 본 SPEC 의 검증 대상 밖 (out of scope per §6)

### AC-SDF-002 — Pattern A Bulk Downgrade

Pattern A 의 50개 SPEC 모두에서:
- frontmatter `status` 가 `completed → implemented` 로 downgrade
- frontmatter `version` 이 patch +1 (예: `0.3.0 → 0.3.1`)
- frontmatter `updated_at` 이 synchronization commit date 로 갱신
- HISTORY 표에 본 SPEC 식별 row 1줄 추가

검증:
- `awk '/^status: implemented$/' .moai/specs/<pattern-A SPEC>/spec.md` 결과 50건 모두 hit
- 각 SPEC `git diff` 의 `+` 라인이 frontmatter + HISTORY 1줄로 한정

### AC-SDF-003 — Pattern B+C Per-SPEC Verification

Pattern B (4건) + Pattern C (6건) 의 10개 SPEC 모두에 대해:
- 개별 verification 결과 (frontmatter downgrade vs sync commit 추가) 가 `.moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001/run-verification.md` 에 SPEC ID 별 row로 기록됨
- frontmatter status 가 git-implied 와 일치하거나, 일치하지 않는 경우 verification.md 에 사유 명시

검증:
- run-verification.md 에 10개 row 모두 존재
- 각 row에 `decision: downgrade` 또는 `decision: keep + sync-commit-pending` 명시

### AC-SDF-004 — Pattern D/E/F/G Detector Exemption

`internal/spec/lint.go::StatusGitConsistencyRule::Check` (또는 동등 파일/함수) 에 terminal-state exemption 추가:
- frontmatter `status` 가 `superseded` 또는 `archived` 인 경우 finding 미발생
- 새 unit test 최소 4 cases 추가 (Pattern D/E/F/G 각 1):
  - case (D): frontmatter `superseded`, git-implied `completed` → no finding
  - case (E): frontmatter `superseded`, git-implied `implemented` → no finding
  - case (F): frontmatter `archived`, git-implied `implemented` → no finding
  - case (G): frontmatter `archived`, git-implied `in-progress` → no finding

검증:
- `go test ./internal/spec/lint/checks/... -run TestStatusGitConsistency_TerminalState` PASS (4 cases)
- 4 SPEC (LSP-001, V3R3-HARNESS-001, I18N-001-ARCHIVED, V3R3-WEB-001) 의 frontmatter `status` 필드 unchanged

### AC-SDF-005 — Pattern H Recursive Cleanup

본 SPEC sync-phase 시점에 cleanup-chain self-drift 4건 (SPECLINT-DEBT-001, LINT-STATUS-CHORE-SKIP-001, LINT-SKIP-CLEANUP-001, STATUS-DRIFT-FOLLOWUP-001 본인) 모두 정합화:

검증:
- 4 SPECs frontmatter `status` 가 git-implied 와 일치하거나 terminal state exemption 으로 통과
- `moai spec lint --strict` 의 4 SPEC 대상 WARN 0건 (AC-SDF-001 의 sub-condition)

### AC-SDF-006 — `lint.skip` Ban Preservation

본 SPEC 머지 후 어떤 SPEC frontmatter에도 `lint.skip` 엔트리가 새로 추가되지 않아야 한다:

검증:
- `grep -lR "lint.skip" .moai/specs/ | wc -l` 결과가 plan-phase baseline (예: 0건 — LSKC-001 PR #937 머지 후 상태) 과 동일

### AC-SDF-007 — Out-of-Scope Files Untouched

64-affected-list + Pattern D/E/F/G 4 SPECs (frontmatter 변경 없음) 외 어떤 SPEC, Go source file, config file, CI workflow, docs-site 도 수정 없음:

검증:
- `git diff --name-only main..HEAD` 결과가 다음 set으로 한정:
  - 50 + 10 + 3 = 63 SPECs `spec.md` (Pattern A + B + C + H frontmatter 변경)
  - 1 detector code file (`internal/spec/lint.go` 또는 `internal/spec/lint/checks/status_git_consistency.go`)
  - 1 detector test file (`*_test.go`)
  - 1 bulk script (`.moai/scripts/status-drift-cleanup.go`) — Optional
  - 본 SPEC 자체 5 artifacts + progress.md + run-verification.md + affected-list-*.txt

### AC-SDF-008 — Idempotency

`status-drift-cleanup.go` 또는 동등 bulk edit 절차를 두 번 실행했을 때 두 번째 실행은 no-op (git diff 0):

검증:
- 1차 실행 → `git diff --stat` 캡처
- 2차 실행 → `git diff --stat` 캡처 (1차 대비 추가 변경 0)
- script 자체 출력에 "already cleaned (no changes)" 등의 idempotent 신호 출력

---

## 5. REQ ↔ AC Coverage Matrix

| REQ | Mapped ACs | Notes |
|-----|-----------|-------|
| REQ-SDF-001 | AC-SDF-004, AC-SDF-007 | Terminal state preservation (D/E/F/G frontmatter unchanged + detector exemption 통해 lint pass) |
| REQ-SDF-002 | AC-SDF-002, AC-SDF-003, AC-SDF-005, AC-SDF-007 | body byte preservation (HISTORY 1줄 제외) |
| REQ-SDF-003 | AC-SDF-002 | version patch bump + updated_at + HISTORY |
| REQ-SDF-004 | AC-SDF-002 | Pattern A bulk downgrade |
| REQ-SDF-005 | AC-SDF-003 | Pattern B verification |
| REQ-SDF-006 | AC-SDF-003 | Pattern C verification (REQ-005 와 동일 verification 절차) |
| REQ-SDF-007 | AC-SDF-001 | Binary lint outcome (primary) |
| REQ-SDF-008 | AC-SDF-004 | Detector exemption code change |
| REQ-SDF-009 | AC-SDF-005 | Pattern H 재귀 cleanup |
| REQ-SDF-010 | AC-SDF-006 | lint.skip ban |
| REQ-SDF-011 | AC-SDF-007 | walker filter unchanged (Go diff에 drift.go 변경 없음) |
| REQ-SDF-012 | AC-SDF-007 | StatusGitConsistencyRule 자체 deprecation 없음 (Check 함수 내 narrow exemption만) |
| REQ-SDF-013 | AC-SDF-002, AC-SDF-003, AC-SDF-005 | body 변경 0줄 |
| REQ-SDF-014 | AC-SDF-007 | 64 외 SPEC 미수정 |
| REQ-SDF-015 | AC-SDF-008 | (Optional) script idempotency |
| REQ-SDF-016 | (Optional) | run-phase 자유 — 검증 대상 아님 |

각 must-have REQ (001-014) 는 최소 1개 AC와 매핑되며, 각 AC는 최소 1개 REQ를 검증한다.

---

## 6. Scope

### In-Scope

- 64 SPEC frontmatter `status` 필드 동기화 (Pattern A: 50, B: 4, C: 6, H: 3 — 총 63 SPEC; D/E/F/G: 1 SPEC씩 = 4 SPECs frontmatter 변경 없음, detector exemption 으로 처리)
- 각 변경 SPEC의 frontmatter `version` patch bump + `updated_at` 갱신 + HISTORY row 1줄 추가
- `internal/spec/lint.go::StatusGitConsistencyRule::Check` (또는 `internal/spec/lint/checks/status_git_consistency.go`) 에 terminal-state exemption 코드 추가 + 4 unit tests
- `.moai/scripts/status-drift-cleanup.go` (또는 LSKC-001 script extension) 작성 또는 보존
- `affected-list-pattern-{A,B,C,H}.txt` 4 파일 작성 (run-phase BASELINE 산출)
- `run-verification.md` 작성 (Pattern B/C 개별 verification 기록)
- `moai spec lint --strict` 사후 검증 (`StatusGitConsistency` WARN 0건)

### Out of Scope (HARD — locked per user directive)

- **옵션 (b) walker filter scope expansion**: `internal/spec/drift.go::shouldSkipCommitTitle` 변경 일체 금지. 새 prefix (`chore:`, `ci:`, `docs:`, `style:` 등) 추가 절대 금지.
- **옵션 (c) StatusGitConsistencyRule deprecation**: rule 자체 비활성화 또는 lint.go 에서 제거 금지. exemption 은 `superseded` / `archived` terminal state 전용 narrow scope.
- **새 lint.skip 도입**: 어떤 SPEC frontmatter에도 `lint.skip` 추가 절대 금지 (LSKC-001 정신 영구 보존).
- **walker filter 외 detector 의미 변경**: `getGitImpliedStatus` 동작 변경 금지. `shouldSkipCommitTitle` skip pattern 확장 금지.
- **다른 lint rule 변경**: `OrphanBCID`, `MissingExclusions`, `RequirementCoverage`, `EARSCompliance` 등 다른 lint rule 변경 금지.
- **64-affected SPEC body 수정**: REQ/AC/HISTORY 본문 수정 금지 (HISTORY 1줄 추가만 허용).
- **64 외 SPEC 변경**: 다른 SPEC frontmatter 또는 body 수정 금지.
- **CI workflow 변경**: `.github/workflows/spec-lint.yml` 등 lint 관련 workflow 수정 금지.
- **docs-site 4-locale sync**: sync phase concern (별도 PR로 분리 또는 추후 SPEC).
- **새 테스트 추가**: AC-SDF-004 의 4 cases (terminal state exemption) 외 추가 테스트 금지 (테스트 부담 최소화).
- **SPEC ID convention / lifecycle status enum 변경**: SPEC-ID naming, status enum (draft/planned/in-progress/implemented/completed/superseded/archived/rejected) 변경 금지.
- **`moai spec lint` CLI flag 변경**: 새 플래그 (예: `--no-status-check`) 도입 금지.
- **CHANGELOG entry**: sync phase 책임 (run-phase에서는 작성하지 않음).

---

## 7. Constraints

- **C1 (HARD)**: predecessor SPEC `SPEC-V3R4-LINT-SKIP-CLEANUP-001` (PR #937, main `758341089`) 가 main에 머지되어 있어야 한다. 그 이전 시점에서는 lint.skip 회피책이 mask 중이므로 본 cleanup의 baseline이 다름.
- **C2 (HARD)**: 본 SPEC 자체의 frontmatter에 `lint.skip` 추가 금지 (REQ-SDF-010 본 SPEC에도 적용).
- **C3 (HARD)**: 모든 수정은 main checkout에서 수행 (worktree 미사용 정책 per `feedback_worktree_never_use`).
- **C4 (HARD)**: 64 SPEC frontmatter 수정은 sequential 수행 (병렬 수정 금지 — lint baseline 검증 충돌 회피).
- **C5 (HARD)**: `version` bump는 patch 단위 (semver Z+1). metadata-only 변경이므로 minor/major bump 부적절.
- **C6 (HARD)**: detector exemption 코드 변경은 narrow scope (`if status == "superseded" || status == "archived" { return nil }` 형태) — 광범위 리팩토링 금지.
- **C7 (HARD)**: Go 1.23+ (per `.claude/rules/moai/languages/go.md`). 코드 변경은 `internal/spec/` 패키지에 한정.
- **C8 (HARD)**: 코드 주석 한국어 (`code_comments: ko` per language.yaml). identifier (변수/함수/타입) 는 English.
- **C9 (HARD)**: `go test ./...` 전체 통과 + `go vet ./...` + `golangci-lint run` 0 warning 유지.
- **C10 (HARD)**: detector exemption 코드 변경에 @MX:NOTE 또는 @MX:ANCHOR 부착 (mx-tag-protocol.md 준수, fan_in 검토 후).

---

## 8. Out of Scope (Explicit Restatement)

본 섹션은 §6 의 explicit list 강조 — run-phase scope creep 방지:

1. **walker filter scope expansion** (옵션 b) — `internal/spec/drift.go` 변경 0건
2. **StatusGitConsistency rule deprecation** (옵션 c) — `lint.go::StatusGitConsistencyRule` 자체 보존
3. **새 `lint.skip` entry 도입** — 어떤 SPEC에도 추가 금지
4. **64 외 SPEC frontmatter 수정** — 격리 보장
5. **64 SPEC body 수정** — REQ/AC/HISTORY 본문 보존 (HISTORY 1줄 추가만)
6. **CI workflow 변경** — `.github/workflows/spec-lint.yml` 등 미수정
7. **docs-site 4-locale documentation sync** — sync-phase 또는 별도 SPEC
8. **추가 lint rule 도입 또는 기존 rule severity 변경**
9. **SPEC-ID naming / lifecycle status enum 변경**
10. **`moai spec lint` CLI flag 추가** (예: `--no-status-check`)
11. **새 테스트 추가** — AC-SDF-004 의 4 cases (terminal state exemption) 외 절제
12. **frontmatter ordering / formatting 변경** — baseline key order 보존 (yaml.Node API 사용 강제)

---

## 9. Glossary

- **status drift**: SPEC frontmatter `status` 필드 값과 git history에서 추론한 lifecycle status 가 불일치하는 상태. `StatusGitConsistencyRule` 이 검출.
- **git-implied status**: `internal/spec/drift.go::getGitImpliedStatus(specID)` 가 `git log --grep=<specID> -50` 결과를 walker filter (chore(spec): skip) 적용 후 가장 최근 의미 있는 commit의 prefix를 `ClassifyPRTitle`로 분류하여 산출한 lifecycle status.
- **walker filter**: PR #933 에서 도입된 `shouldSkipCommitTitle` helper. `chore(spec):` / `chore(specs):` prefix commit을 skip하여 sweep commit이 git-implied status를 오염시키는 bootstrap bug 방지.
- **terminal state**: lifecycle status enum 중 `superseded` / `archived` 두 상태. SPEC이 더 이상 진행되지 않는 상태로 git history와 mismatch 가 정상.
- **bulk script**: `.moai/scripts/status-drift-cleanup.go` (계획). LSKC-001의 `lint-skip-cleanup.go` 를 fork한 Go script. YAML node API로 frontmatter 안전 수정 + body byte preservation + idempotent.
- **affected-list**: `affected-list-pattern-{A,B,C,H}.txt` 4 파일. 각 패턴별로 처리 대상 SPEC `spec.md` 절대 경로 list.
- **Wave**: 본 SPEC의 milestone-equivalent 단위. 5 waves (BASELINE / Pattern A bulk / Pattern B+C verification / Pattern D-G detector / Pattern H 재귀).
- **cleanup chain**: SPECLINT-DEBT-001 → LINT-STATUS-CHORE-SKIP-001 → LINT-SKIP-CLEANUP-001 → STATUS-DRIFT-FOLLOWUP-001 4단계 lint debt 처리 SPEC 시리즈. 본 SPEC이 마지막 wave.

---

## 10. Exclusions (What NOT to Build)

[HARD] 본 SPEC 은 다음 항목을 명시적으로 만들지 않는다:

1. **walker filter expansion** — `chore:`, `ci:`, `docs:`, `style:`, `refactor:`, `perf:`, `test:` 어떤 prefix 도 `shouldSkipCommitTitle` 에 추가하지 않음. (옵션 b 거부)
2. **StatusGitConsistencyRule 비활성화 / deprecation** — rule 자체는 영구 active. (옵션 c 거부)
3. **새 `lint.skip` entry** — 본 SPEC 자체 frontmatter 포함 어떤 SPEC에도 `lint.skip` 추가 0건. (LSKC-001 정신 영구 보존)
4. **detector 의미 광범위 변경** — `getGitImpliedStatus` 동작, `ClassifyPRTitle` prefix 매핑, lifecycle status enum 어느 것도 변경 없음. exemption 만 narrow scope.
5. **64 SPEC body content 변경** — REQ/AC/HISTORY 본문 수정 0줄 (HISTORY 1줄 추가만 허용).
6. **64 외 SPEC 수정** — 다른 SPEC frontmatter / body 0건 변경.
7. **CI workflow 변경** — `.github/workflows/spec-lint.yml` 미수정.
8. **새 lint rule** — `StatusGitConsistencyRule` 외 rule 추가 / 변경 0건.
9. **새 CLI flag** — `moai spec lint` 명령에 새 옵션 추가 0건.
10. **docs-site 4-locale 동기화** — sync-phase 또는 별도 SPEC scope.
11. **AC-SDF-004 외 새 테스트** — terminal state exemption 4 cases 외 추가 테스트 0건 (절제).
12. **frontmatter format / ordering 변경** — baseline key order, 들여쓰기 스타일, quote style 보존 (yaml.Node API 사용 강제).
13. **CHANGELOG entry** — run-phase 미작성, sync-phase 책임.

---

## REQ Coverage

REQ-SDF-001, REQ-SDF-002, REQ-SDF-003, REQ-SDF-004, REQ-SDF-005, REQ-SDF-006, REQ-SDF-007, REQ-SDF-008, REQ-SDF-009, REQ-SDF-010, REQ-SDF-011, REQ-SDF-012, REQ-SDF-013, REQ-SDF-014, REQ-SDF-015, REQ-SDF-016
