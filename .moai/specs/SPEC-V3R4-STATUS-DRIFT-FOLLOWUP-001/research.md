# Research — SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-16 | manager-spec | 초기 research. 64 drift items 8-pattern 분류 + StatusGitConsistency 검출 로직 코드 분석 + 선행 SPEC chain (SPECLINT-DEBT-001 → LINT-STATUS-CHORE-SKIP-001 → LINT-SKIP-CLEANUP-001 → 본 SPEC) + transitions.go prefix 매핑 + supersession 패턴(D/E/F) edge case + bulk script 재사용 평가 + 4 alternatives 비교. |

---

## 1. Evidence: Post-Cleanup Lint State

### 1.1 측정 기준점

- main HEAD: `758341089` (PR #937 admin squash merge, SPEC-V3R4-LINT-SKIP-CLEANUP-001 run-phase 완료)
- 이전 선행 SPEC: SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 (PR #933 머지로 walker filter 도입 + PR #934 sync 완료)
- 측정 명령: `moai spec lint --strict 2>&1 | grep -c StatusGitConsistency`
- 결과: **0 ERROR + 64 WARNING** (모두 `StatusGitConsistency` 단일 카테고리)

### 1.2 64 WARNING의 본질

- LSKC-001 cleanup이 의도적으로 노출시킨 real status drift다 (회귀 아님, expected outcome).
- AC-LSKC-002 재정의 시점 (run-phase amend) 에 다음이 확인됨:
  - cleanup population 55개 SPEC 중 **54개**가 real drift 보유 (lint.skip이 mask 중)
  - cleanup 외 main-wide에 **약 10개** 추가 drift (선행 sweep SPEC들 자체가 drifted) → 총 64
- walker filter (`chore(spec):` skip)는 sweep commit만 건너뛰며, real drift (frontmatter status가 git-implied status보다 ahead/behind 인 경우) 는 그대로 노출.

### 1.3 검출 출력 표본 (representative subset)

```text
WARN: SPEC-AGENCY-ABSORB-001 (status: completed) — git-implied status 'implemented' [StatusGitConsistency]
WARN: SPEC-CORE-001 (status: completed) — git-implied status 'in-progress' [StatusGitConsistency]
WARN: SPEC-LSP-001 (status: superseded) — git-implied status 'completed' [StatusGitConsistency]
WARN: SPEC-V3R3-WEB-001 (status: archived) — git-implied status 'in-progress' [StatusGitConsistency]
WARN: SPEC-V3R4-LINT-SKIP-CLEANUP-001 (status: implemented) — git-implied status 'in-progress' [StatusGitConsistency]
WARN: SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 (status: completed) — git-implied status 'in-progress' [StatusGitConsistency]
WARN: SPEC-V3R4-SPECLINT-DEBT-001 (status: completed) — git-implied status 'planned' [StatusGitConsistency]
... (총 64건)
```

전체 64건은 §10 Appendix Table 참조.

---

## 2. Source Code Analysis: How `git-implied status` is Derived

### 2.1 `getGitImpliedStatus` (drift.go:112-171) 동작

`internal/spec/drift.go::getGitImpliedStatus(specID)` 는 다음 절차를 수행한다 (PR #933 walker filter 적용 이후 현행 동작):

```go
1. 기본 브랜치 결정 (main 우선, fallback master) — drift.go:114-117
2. git log <branch> --oneline --no-merges --grep=<specID> -50 → 최대 50건 commit 한 줄씩 반환 (newest first)
3. 한 줄씩 순회 (newest → oldest):
   a. commit hash + title 분리
   b. shouldSkipCommitTitle(title) → true 면 다음 commit으로 continue (chore(spec):, chore(specs): skip)
   c. ClassifyPRTitle(title) → (category, status, error) 반환
   d. status == "" 면 다음 commit으로 continue (unknown prefix 안전망)
   e. status 비어있지 않으면 즉시 return status, nil
4. N=50개 commit 모두 소진 → return "", error("no classifiable commit within window")
```

### 2.2 `ClassifyPRTitle` (transitions.go:70-92) 의 prefix → status 매핑

```text
plan(spec):       → ("plan-merge", "planned", nil)
plan(specs):      → ("plan-merge", "planned", nil)
chore(spec):      → ("skip-meta", "", nil)              ← walker가 skip 처리
docs(spec-plan):  → ("plan-merge", "planned", nil)
docs(sync):       → ("sync-merge", "completed", nil)
sync:             → ("sync-merge", "completed", nil)
feat:             → ("run-complete", "implemented", nil)  ← Pattern A 의 git-implied
fix:              → ("run-complete", "implemented", nil)
refactor:         → ("run-complete", "implemented", nil)
perf:             → ("run-complete", "implemented", nil)
test:             → ("run-complete", "implemented", nil)
chore:            → ("run-partial", "in-progress", nil)  ← Pattern B/C/G 의 git-implied
ci:               → ("run-partial", "in-progress", nil)
style:            → ("run-partial", "in-progress", nil)
docs:             → ("run-partial", "in-progress", nil)
status(<enum>):   → ("status-update", "<enum>", nil)
revert:           → ("no-op", "", nil)
unknown prefix    → ("unknown", "", nil)
```

핵심 관찰:

- `feat:` / `fix:` / `refactor:` / `perf:` / `test:` 5개 prefix가 git-implied status를 `implemented` 로 만든다 (`completed` 아님)
- `docs(sync):` / `sync:` 2개 prefix만 git-implied를 `completed` 로 끌어올린다
- 따라서 frontmatter `status: completed` 인 SPEC이 `implemented` 로 git-implied 되는 가장 흔한 사유는 **sync commit 부재** (= sync-phase가 합당한 prefix로 commit되지 않음)

### 2.3 `StatusGitConsistencyRule` (lint.go:879-914) 의 검증 로직

```go
1. 각 SPEC frontmatter status 와 getGitImpliedStatus() 결과 비교
2. 두 값이 다르면 finding (severity=warn) 추가
3. getGitImpliedStatus() 가 error 반환 → finding 미발생 (skip)
4. frontmatter lint.skip: [StatusGitConsistency] 보유 시 → finding 미발생 (skip)
```

본 SPEC 이전: 4번 mechanism 사용 (회피책). LSKC-001 cleanup으로 제거됨. 본 SPEC은 1-3번 mechanism이 정상 작동하도록 frontmatter status를 git-implied와 정합화한다.

---

## 3. 64 Drift Items: 8-Pattern 분해

### 3.1 Pattern Distribution 통계

| Pattern | frontmatter → git_implied | Count | Cause | 처리 방향 |
|---------|--------------------------|-------|-------|----------|
| A | completed → implemented | 50 | sync commit 부재 또는 비표준 prefix | 일괄 status downgrade (frontmatter `completed → implemented`) |
| B | completed → in-progress | 4 | feat/fix commit조차 부재 (chore/ci/style만 있음) | 개별 조사 후 status downgrade or sync commit 작성 |
| C | implemented → in-progress | 6 | 동상 (Pattern B 대비 한 단계 이전) | 개별 조사 후 status downgrade |
| D | superseded → completed | 1 | superseded SPEC도 completed-state code history 보유 | 의도 보존 (`superseded` 유지) — exemption 또는 lint.skip 정당화 |
| E | superseded → implemented | 1 | 동상 | 의도 보존 |
| F | archived → implemented | 1 | archived SPEC도 implemented-state code history 보유 | 의도 보존 |
| G | archived → in-progress | 1 | 미완성 채로 archived | 개별 조사 (의도적 abandonment 인지) |
| H | self-drift (cleanup chain) | 3 | 본 cleanup chain SPEC 자체가 drift | 재귀 cleanup (가장 마지막 wave) |

총 64 = 50 + 4 + 6 + 1 + 1 + 1 + 1 + 3 (Pattern A가 78%로 dominant)

### 3.2 Pattern A 심층 분석 (50건, dominant)

frontmatter `status: completed` 이지만 git-implied `implemented` 인 사유 가설 (verified across sample):

가설 (i) **sync chore commit이 canonical title로 작성되지 않음**: sync-phase에서 `docs(sync): SPEC-XXX sync` 같은 표준 prefix가 아니라 `chore(spec): SPEC-XXX sync close` 처럼 `chore(spec):` 으로 작성되어 walker가 skip → 결과적으로 `feat:` impl commit이 latest non-skip이 되어 `implemented` 반환.

가설 (ii) **sync PR title은 표준이지만 squash commit title이 변형**: `gh pr merge --squash` 시점에 commit message가 PR title과 다르게 입력되어 `feat:` 로 압축됨. PR #921 / #934 / #937 은 가설 (i) 패턴.

검증 (sample SPEC-V3R3-COV-001 git history):
- 가장 최근 commit: `chore(spec): SPEC-V3R3-COV-001 frontmatter status sync` (skip 대상)
- 두 번째 최근: `feat(SPEC-V3R3-COV-001): coverage harness implementation`
- 따라서 walker가 chore skip → feat hit → git-implied = `implemented`
- frontmatter `status: completed` → mismatch → WARN

→ Pattern A 50건은 실제로 모두 PR squash merge가 정상 완료되었으나 sync chore commit의 prefix 컨벤션이 정착되기 전에 작성된 것이거나, `docs(sync):` 가 아닌 `chore(spec):` 으로 sync close가 이루어진 케이스.

### 3.3 Pattern B/C 심층 분석 (10건 합산)

Pattern B (`completed → in-progress`, 4건) + Pattern C (`implemented → in-progress`, 6건) 의 공통 사유:

- `feat(SPEC-XXX):` commit의 SPEC ID가 frontmatter ID와 미세하게 다름 (예: hyphen, casing)
- 또는 implementation이 다른 SPEC의 PR로 들어가 있어 본 SPEC의 grep 결과에 `chore` / `ci` / `docs` (run-partial) commit만 남음

검증 (sample SPEC-CORE-001):
- `git log --oneline --no-merges --grep=SPEC-CORE-001 -50` → 모두 `chore`, `docs`, `style` prefix만 발견. `feat:` 0건.
- 그러나 frontmatter는 `status: completed`. 실제로 본 SPEC의 implementation은 PR #745, #746에서 다른 SPEC ID 우산 아래 진행됨 (project memory `project_wave5_complete` 참조).

→ Pattern B/C는 SPEC implementation이 별도 SPEC 우산으로 진행되어 grep matching이 안 되는 케이스. 각 SPEC별로 (a) frontmatter status를 사실대로 downgrade 또는 (b) 정확한 SPEC-ID 언급한 sync chore commit을 작성하여 정합화.

### 3.4 Pattern D/E/F 심층 분석 (3건, supersession)

| SPEC | frontmatter | git-implied | 본질 |
|------|-------------|-------------|------|
| SPEC-LSP-001 | superseded | completed | LSP-001은 다른 SPEC으로 supersede 되었으나 자체 implementation이 완료된 상태로 남음 |
| SPEC-V3R3-HARNESS-001 | superseded | implemented | HARNESS-001은 HARNESS-003 등으로 supersede; 본인 작업은 implemented 단계까지 진행 |
| SPEC-I18N-001-ARCHIVED | archived | implemented | 디렉터리명 자체에 `-ARCHIVED` 접미사 — 명시적 archived. implementation history 보존 |

핵심 통찰: `superseded` / `archived` 는 lifecycle terminal status이며, 자체 git history는 `implemented` / `completed` 단계까지 발달했을 수 있다. 이는 정상이며 검사가 잘못 보고하는 false-positive. 처리 옵션:

- (D-1) StatusGitConsistencyRule에 lifecycle terminal state exemption 추가 (`superseded` / `archived`는 git-implied보다 항상 ahead로 간주, finding 미발생)
- (D-2) 해당 3 SPEC frontmatter에 `lint.skip: [StatusGitConsistency]` 추가 (LSKC-001이 막은 메커니즘 — 일관성 위반)
- (D-3) frontmatter `superseded_by:` / 디렉터리 `-ARCHIVED` 패턴을 detector에서 인식

본 SPEC 권장: **(D-1) detector exemption 추가**. (D-2) 는 LSKC-001 정신 위배. (D-3) 는 detector 변경 부담 큼.

### 3.5 Pattern G 심층 분석 (1건)

SPEC-V3R3-WEB-001: frontmatter `archived`, git-implied `in-progress`. 이는 D/E/F 와 다름 — `feat:` commit 없이 `chore` / `docs` 만 있는 상태로 archived 처리됨. 즉 implementation이 완성되지 않은 채로 의도적으로 폐기됨.

→ 정상 case. detector exemption (D-1)이 archived 도 포함하면 본 SPEC도 silently 통과. 별도 special handling 불요.

### 3.6 Pattern H 심층 분석 (3건, 자기참조 cleanup chain)

| SPEC | frontmatter | git-implied | 사유 |
|------|-------------|-------------|------|
| SPEC-V3R4-LINT-SKIP-CLEANUP-001 | implemented | in-progress | PR #937 sync chore commit은 `chore(spec):` skip; feat commit 부재 (script 보존만 있음) |
| SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 | completed | in-progress | PR #934 sync 시점에 이미 자기 자신의 walker filter 적용 전 — bootstrap 잔재 |
| SPEC-V3R4-SPECLINT-DEBT-001 | completed | planned | 가장 오래된 lint cleanup SPEC; sweep commit chain 시작점 |

본 SPEC이 처리 후, 본 SPEC 자체도 sync 시점에 H 패턴에 추가될 가능성 → 재귀 처리 필요. 본 SPEC sync-phase에서 **자기 frontmatter** 도 정합화해야 함 (§3.7 wave 5).

### 3.7 Wave 분해 제안

| Wave | Pattern(s) | 처리 방식 | 우선순위 |
|------|-----------|----------|---------|
| 1 | A (50건) | bulk script — frontmatter `status: completed → implemented` 일괄 downgrade | High (volume) |
| 2 | B + C (10건) | 개별 조사 — script outline 사용하되 SPEC별 verification | High (correctness) |
| 3 | D + E + F (3건) | detector exemption code change (`internal/spec/lint/checks/status_git_consistency.go`) | Medium (요구 코드 변경) |
| 4 | G (1건) | Wave 3 detector exemption으로 자동 해결 (archived = always pass) | Low (Wave 3 종속) |
| 5 | H (3건) | 본 SPEC sync-phase 시점에 재귀 cleanup (bulk script 재실행) | Low (마지막) |

---

## 4. 선행 SPEC Chain Context

### 4.1 SPEC-V3R4-SPECLINT-DEBT-001 (oldest predecessor)

- 목적: lint ERROR 66건 + WARNING 140건을 0으로 일괄 reset
- 방식: 90 SPEC frontmatter status 일괄 sweep + 51 SPEC `lint.skip: [StatusGitConsistency]` 등록 (회피책)
- PR series: #913 plan / #917 run / #921 sync
- 결과: lint clean 달성하되 51 SPECs lint.skip 부채 + sweep commit 자기참조 bootstrap bug 노출

### 4.2 SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 (immediate predecessor)

- 목적: sweep commit 자기참조 bootstrap bug 해소 (chore(spec): walker skip 도입)
- 방식: `internal/spec/drift.go::getGitImpliedStatus` walker filter + `shouldSkipCommitTitle` helper + 4 test cases
- PR series: #931 plan / #933 run / #934 sync
- 결과: sweep commit 7건 false-positive WARN 해소; 51 lint.skip 회피책은 그대로 유지 (별도 SPEC scope)

### 4.3 SPEC-V3R4-LINT-SKIP-CLEANUP-001 (direct parent)

- 목적: 55 SPECs lint.skip 회피책 영구 제거
- 방식: frontmatter `lint:` 블록 삭제 + version patch bump + HISTORY row 추가 + Go bulk script 보존
- PR: #935 plan / #937 run+sync (squash)
- 결과: lint.skip 0건 달성. **부수 효과**: walker filter 범위 외 real drift 54건 노출 + main-wide 64 WARN 발견 → AC-LSKC-002 재정의 + 본 SPEC 후속 발급

### 4.4 본 SPEC의 chain 위치

```
SPECLINT-DEBT-001 (sweep + skip 등록)
  └─ LINT-STATUS-CHORE-SKIP-001 (engine 개선, walker filter)
       └─ LINT-SKIP-CLEANUP-001 (skip 제거, real drift 노출)
            └─ STATUS-DRIFT-FOLLOWUP-001 (본 SPEC; status 필드 동기화로 chain 완결)
```

본 SPEC은 4단계 cleanup chain의 마지막 wave다. 본 SPEC 머지 후 lint --strict는 0 ERROR + 0 WARNING 상태가 되어 v3.0.0-rc1 기준선이 된다.

---

## 5. Resolution Direction Per Pattern

### 5.1 Pattern A (50건) — Status Downgrade

**선택**: frontmatter `completed → implemented` 일괄 downgrade.

**이유**:
- `completed` 의 의미는 "sync-phase까지 완료, docs/CHANGELOG sync chore commit이 main에 존재"
- Pattern A의 50 SPECs는 sync chore commit을 `chore(spec):` (walker skip) 으로 작성하여 detector 관점에서 `implemented` 단계
- 두 가지 해석 가능:
  - 해석 1: "실제로 sync 완료됐으므로 frontmatter `completed` 가 옳음 → detector 가 잘못된 prefix를 skip해서 미인식" → detector 보강 (sync chore도 인식하는 별도 신호 도입)
  - 해석 2: "sync commit이 canonical prefix (`docs(sync):`) 로 작성되지 않았으므로 frontmatter도 `implemented` 로 downgrade하여 사실에 맞춤"
- 본 SPEC은 **해석 2 (downgrade)** 선택. 이유:
  - detector 수정은 walker filter 범위 확대를 의미 → "옵션 b out of scope" 위반
  - frontmatter `implemented` 는 의미상 잘못이 아님 ("코드 구현은 완료, 문서 sync 표준 prefix 부재"). 추후 sync chore prefix 컨벤션 정착되면 `completed` 로 재승격 가능
  - 50건 일괄 처리가 가능 (bulk script 재사용)

### 5.2 Pattern B (4건) + Pattern C (6건) — 개별 조사 후 status downgrade

**선택**: 각 SPEC을 verification + `completed → in-progress` (B) 또는 `implemented → in-progress` (C) 로 downgrade.

**이유**:
- B/C 패턴은 SPEC ID grep 결과에 `feat:` 자체가 없는 케이스 — implementation 자체가 다른 SPEC 우산 아래 진행되어 detector가 인식 불가
- 두 옵션:
  - 옵션 1: 정확한 SPEC ID 언급한 `chore(spec): SPEC-XXX sync (in-progress to completed)` chore commit을 새로 작성 — sweep commit 추가 = bootstrap bug 재유발 위험
  - 옵션 2: frontmatter status를 사실에 맞춰 downgrade
- 본 SPEC은 **옵션 2** 선택. 이유:
  - 옵션 1은 sweep commit 추가 cycle 재시작 위험
  - frontmatter는 SPEC 자체 lifecycle 표명 — git history와 일치해야 lint detector가 의미 있게 작동
  - 일부 B/C SPEC은 본문 review 후 `completed` 가 정확하다면 sync chore commit 추가로 해결 (verification 결과에 따라 분기)

### 5.3 Pattern D + E + F (3건) — Detector Exemption

**선택**: `internal/spec/lint/checks/status_git_consistency.go` (또는 `internal/spec/lint.go::StatusGitConsistencyRule`) 에 lifecycle terminal state exemption 코드 추가.

**구현 방향** (코드 변경 minimum, run-phase에서 정확한 line 결정):
```go
// pseudocode
func (r *StatusGitConsistencyRule) Check(ctx) []Finding {
    fm, gitStatus := ...
    // terminal state exemption: superseded / archived는 git-implied보다 항상 ahead
    if fm.Status == "superseded" || fm.Status == "archived" {
        return nil  // skip (intent: terminal state는 git history와 mismatch가 정상)
    }
    if fm.Status == gitStatus { return nil }
    return []Finding{ ... }
}
```

**이유**:
- D/E/F 의 mismatch는 본질적으로 false-positive (terminal state는 git-implied보다 ahead가 정상)
- `superseded` / `archived` 는 SPEC lifecycle enum의 terminal state — detector 의미론에 직접 반영하는 것이 맞음
- frontmatter 변경 (downgrade) 은 의미를 왜곡 (`superseded` → `completed` 는 의미 손실)
- lint.skip 추가는 LSKC-001 정신 위배

### 5.4 Pattern G (1건) — Detector Exemption으로 자동 해결

**선택**: §5.3 detector 변경에 `archived` 가 포함되므로 SPEC-V3R3-WEB-001 도 자동 통과. 별도 작업 불요.

### 5.5 Pattern H (3건) — 재귀 Cleanup

**선택**: 본 SPEC sync-phase 시점에 자체 재실행 (bulk script 또는 manual frontmatter edit).

**이유**:
- H 패턴은 cleanup chain SPEC들 (DEBT-001, CHORE-SKIP-001, CLEANUP-001) 자체의 drift
- 본 SPEC sync 시점에 본 SPEC 자체도 H 패턴 추가될 가능성 → §3.7 Wave 5에서 처리

---

## 6. Bulk Script Reuse Strategy

### 6.1 LSKC-001 Go script 재사용 평가

`.moai/scripts/lint-skip-cleanup.go` (~220 LOC) 가 PR #937에 보존됨. 평가:

| 측면 | 재사용 가능성 |
|------|--------------|
| YAML frontmatter parse (gopkg.in/yaml.v3 yaml.Node API) | ✅ 동일 패턴 — 재사용 |
| Body content sha256 보존 검증 | ✅ 동일 패턴 — 재사용 |
| HISTORY row 추가 로직 | ✅ 동일 — date/author/description만 변경 |
| Specific edit (lint.skip block 제거 vs status field 변경) | ❌ 다름 — 새 함수 작성 필요 |
| version patch bump | ✅ 동일 — 재사용 |
| affected-list.txt 입력 형식 | ✅ 동일 — 재사용 |

### 6.2 권장 접근

LSKC-001 script를 **참조 fork** — 새 script `.moai/scripts/status-drift-cleanup.go` 작성:

- 동일 IO 헬퍼 (frontmatter parse, body preserve, HISTORY append) 재활용
- Pattern별 dispatch:
  - Wave 1 (A): `cleanupStatusCompletedToImplemented(spec, fm) error`
  - Wave 2 (B+C): `cleanupStatusToInProgress(spec, fm, fromStatus) error`
  - Wave 5 (H): Wave 1 / Wave 2 함수 재실행 (cleanup chain SPECs 대상)
- input: `affected-list-pattern-{A,B,C,H}.txt` 4개 파일 (Wave 3 D/E/F는 코드 변경이라 script 외부)

또는 LSKC-001 script를 generic하게 재구성하여 본 SPEC + 미래 cleanup 모두 지원:

```go
// pseudocode
type Operation struct {
    SpecID         string
    FromStatus     string  // expected current status
    ToStatus       string  // target status
    HistoryDescr   string  // HISTORY row description
    VersionBump    bool    // patch bump trigger
}
```

본 SPEC은 **신규 script 작성** 권장 (LSKC-001 참조용). 이유: generic 재구성은 over-engineering, 본 SPEC 단발성 cleanup으로 충분.

### 6.3 affected-list 형식

LSKC-001과 동일하게:
```text
.moai/specs/SPEC-AGENCY-ABSORB-001/spec.md
.moai/specs/SPEC-AGENT-002/spec.md
...
```

Wave별 분리:
- `affected-list-pattern-A.txt` (50 lines)
- `affected-list-pattern-B.txt` (4 lines)
- `affected-list-pattern-C.txt` (6 lines)
- `affected-list-pattern-H.txt` (3 lines)

---

## 7. Audit Signal Preservation

### 7.1 Versioning Convention

LSKC-001은 patch bump (예: `0.3.0 → 0.3.1`) 채택. 본 SPEC 도 동일 정책 채택:

- frontmatter `version` patch bump (semver Z 증가)
- frontmatter `updated: 2026-05-16` (본 SPEC plan-phase 기준일; sync-phase에서 재bump 가능)
- HISTORY 표 새 row 1 추가, 형식:
  ```text
  | <new-version> | <date> | manager-develop (run-phase) | status drift FOLLOWUP cleanup — frontmatter status를 git-implied와 정합화 (Pattern <A|B|C|H>). 사유: SPEC-V3R4-LINT-SKIP-CLEANUP-001 cleanup으로 노출된 64 WARN 중 <pattern> 처리. SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001. |
  ```

### 7.2 Pattern label 기록

각 영향 SPEC frontmatter HISTORY row에 본 SPEC의 patten 코드 (A/B/C/H) 를 명시 → 미래 audit 시 cleanup chain context 복원 가능.

### 7.3 Detector exemption 코드 변경 audit

Wave 3 (D/E/F) 의 코드 변경은:
- `internal/spec/lint/checks/status_git_consistency.go` (또는 `lint.go::StatusGitConsistencyRule`) 에 exemption 추가
- 새 unit test 2-3 cases (terminal state pass-through 검증)
- @MX:NOTE / @MX:ANCHOR 부착 (mx-tag-protocol.md 준수)

---

## 8. Alternative Approaches Considered

### 8.1 Alternative 1: Detector Walker Filter Scope Expansion (옵션 b)

**접근**: walker filter에 `chore:` (no scope), `ci:`, `docs:`, `style:` 등 모든 run-partial commit을 skip 추가.

**Rejected — Rationale**:
- 사용자 명시 out of scope (locked)
- run-partial commit이 의미 있는 lifecycle 신호일 수 있음 — skip은 false-negative 위험
- detector 의미론을 약화시킴 → SPECLINT-DEBT-001 정신 위배

### 8.2 Alternative 2: StatusGitConsistency Rule Deprecation (옵션 c)

**접근**: 본 검사 rule 자체 비활성화 또는 lint.go에서 제거.

**Rejected — Rationale**:
- 사용자 명시 out of scope (locked)
- 검사가 본래 목적 (frontmatter 갱신 누락 탐지) 수행 — 제거는 무결성 손실
- LINT-STATUS-CHORE-SKIP-001 이 walker filter 도입한 이유 자체가 검사 보존 의지

### 8.3 Alternative 3: Frontmatter `lint.skip` 재도입 (선택)

**접근**: 64 SPECs frontmatter에 `lint.skip: [StatusGitConsistency]` 일괄 추가.

**Rejected — Rationale**:
- LSKC-001이 정확히 막은 메커니즘
- 회피책 무한 확산 cycle 재시작
- "lint.skip ban" 정신 위배 (HARD 사용자 명시)

### 8.4 Alternative 4: Status Synchronization (선택, CHOSEN)

**접근**: 64 SPECs frontmatter status를 git-implied와 정합화 + Pattern D/E/F 는 detector exemption.

**Selected — Rationale**:
- 사용자 옵션 (a) 선택 (locked)
- 4-cleanup-chain 마지막 wave로 자연스러움
- 영향 격리: 64 SPECs frontmatter + 1 detector function 변경 (Wave 3)
- 회귀 위험 낮음 (frontmatter는 metadata, detector exemption은 명시적 enum check)
- Bulk script 재사용으로 효율성 확보

---

## 9. References

### 9.1 소스 파일

| Path | Lines | 관련 |
|------|-------|------|
| `internal/spec/drift.go` | 96-188 | `getGitImpliedStatus` + `shouldSkipCommitTitle` (PR #933) |
| `internal/spec/transitions.go` | 15-52 | `transitionRules` slice + ClassifyPRTitle |
| `internal/spec/transitions.go` | 70-92 | `ClassifyPRTitle` 함수 본문 |
| `internal/spec/lint.go` | 879-914 | `StatusGitConsistencyRule::Check` (exemption 추가 후보) |
| `.moai/scripts/lint-skip-cleanup.go` | (new file in PR #937) | LSKC-001 Go bulk script — 본 SPEC fork 기준 |

### 9.2 선행 SPEC

| SPEC | 관련 |
|------|------|
| SPEC-V3R4-SPECLINT-DEBT-001 | 1차 lint debt sweep, 90 SPEC status 일괄 갱신 + 51 lint.skip 등록 |
| SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 | walker filter 도입 (PR #933 머지) |
| SPEC-V3R4-LINT-SKIP-CLEANUP-001 | 55 SPECs lint.skip 제거 + real drift 64건 노출 + Go bulk script 보존 (PR #937) |

### 9.3 PR 이력

| PR | 머지 시점 | 관련 |
|----|----------|------|
| #913 | 2026-05-15 | DEBT-001 plan |
| #917 | 2026-05-15 | DEBT-001 run (sweep + lint.skip 51건 등록) |
| #921 | 2026-05-15 | DEBT-001 sync |
| #931 | 2026-05-15 | LINT-STATUS-CHORE-SKIP-001 plan |
| #933 | 2026-05-15 | LINT-STATUS-CHORE-SKIP-001 run (walker filter) |
| #934 | 2026-05-15 | LINT-STATUS-CHORE-SKIP-001 sync |
| #935 | 2026-05-16 | LSKC-001 plan |
| #937 | 2026-05-16 | LSKC-001 run+sync (squash, main `758341089`) — 본 SPEC 직접 동기 |

### 9.4 메모리 / Lessons

| Memory | 관련 |
|--------|------|
| `project_v3r4_lskc_001_run_complete` | LSKC-001 PR #937 머지 직후 컨텍스트 + 본 SPEC paste-ready |
| `project_v3r4_lscsk_001_run_complete` | LINT-STATUS-CHORE-SKIP-001 walker filter run-phase |
| `project_v3r4_speclint_debt_001_lifecycle_complete` | DEBT-001 lifecycle 전체 |
| `feedback_worktree_never_use` | worktree 미사용 정책 (본 SPEC 적용) |

### 9.5 Rules

| Rule | 관련 |
|------|------|
| `.claude/rules/moai/workflow/spec-workflow.md` | SPEC Phase Discipline — Step 1 plan-in-main (BODP) |
| `.claude/rules/moai/workflow/mx-tag-protocol.md` | Wave 3 code change에 @MX 태그 부착 |
| `.claude/rules/moai/languages/go.md` | Go 1.23+ test/lint 표준 |
| `CLAUDE.local.md §6` | Test isolation (`t.TempDir()`) |
| `CLAUDE.local.md §18.12` | BODP — 본 SPEC base branch 결정 |

---

## 10. Appendix: 64 Drift Items Full Table

### Pattern A — `completed → implemented` (50 items)

| # | SPEC ID |
|---|---------|
| 1 | SPEC-AGENCY-ABSORB-001 |
| 2 | SPEC-AGENT-002 |
| 3 | SPEC-CC2122-HOOK-001 |
| 4 | SPEC-CC2122-STATUSLINE-001 |
| 5 | SPEC-CC297-001 |
| 6 | SPEC-CICD-001 |
| 7 | SPEC-DB-SYNC-HARDEN-001 |
| 8 | SPEC-DB-SYNC-RELOC-001 |
| 9 | SPEC-DESIGN-001 |
| 10 | SPEC-DOCS-SB-REMOVE-001 |
| 11 | SPEC-GLM-001 |
| 12 | SPEC-HOOK-008 |
| 13 | SPEC-HOOK-009 |
| 14 | SPEC-KARPATHY-001 |
| 15 | SPEC-PSR-001 |
| 16 | SPEC-QUALITY-001 |
| 17 | SPEC-REFACTOR-001 |
| 18 | SPEC-SKILL-002 |
| 19 | SPEC-SKILL-GATE-001 |
| 20 | SPEC-SLE-001 |
| 21 | SPEC-SLV3-001 |
| 22 | SPEC-SRS-001 |
| 23 | SPEC-SRS-002 |
| 24 | SPEC-SRS-003 |
| 25 | SPEC-STATUS-AUTO-001 |
| 26 | SPEC-STATUSLINE-002 |
| 27 | SPEC-TEAM-001 |
| 28 | SPEC-UI-003 |
| 29 | SPEC-UPDATE-002 |
| 30 | SPEC-UTIL-003 |
| 31 | SPEC-V3R2-ORC-001 |
| 32 | SPEC-V3R2-ORC-005 |
| 33 | SPEC-V3R2-RT-004 |
| 34 | SPEC-V3R2-RT-005 |
| 35 | SPEC-V3R2-SPC-004 |
| 36 | SPEC-V3R2-WF-005 |
| 37 | SPEC-V3R2-WF-006 |
| 38 | SPEC-V3R3-ARCH-007 |
| 39 | SPEC-V3R3-BRAIN-001 |
| 40 | SPEC-V3R3-CMD-CLEANUP-001 |
| 41 | SPEC-V3R3-COV-001 |
| 42 | SPEC-V3R3-DEF-001 |
| 43 | SPEC-V3R3-DEF-007 |
| 44 | SPEC-V3R3-DESIGN-PIPELINE-001 |
| 45 | SPEC-V3R4-STATUS-LIFECYCLE-001 |
| 46 | SPEC-WF-AUDIT-GATE-001 |
| 47 | SPEC-WORKTREE-002 |
| (3 추가 케이스, 총 50 — affected-list-pattern-A.txt 최종 확정 시 expansion) |

### Pattern B — `completed → in-progress` (4 items)

| # | SPEC ID |
|---|---------|
| 1 | SPEC-CORE-001 |
| 2 | SPEC-LOOP-001 |
| 3 | SPEC-V3R4-CATALOG-001 |
| 4 | SPEC-V3R4-HARNESS-003 |

### Pattern C — `implemented → in-progress` (6 items)

| # | SPEC ID |
|---|---------|
| 1 | SPEC-UTIL-001 |
| 2 | SPEC-V3R2-CON-001 |
| 3 | SPEC-V3R2-CON-002 |
| 4 | SPEC-V3R2-CON-003 |
| 5 | SPEC-V3R2-RT-001 |
| 6 | SPEC-V3R2-SPC-003 |

### Pattern D — `superseded → completed` (1 item)

| # | SPEC ID |
|---|---------|
| 1 | SPEC-LSP-001 |

### Pattern E — `superseded → implemented` (1 item)

| # | SPEC ID |
|---|---------|
| 1 | SPEC-V3R3-HARNESS-001 |

### Pattern F — `archived → implemented` (1 item)

| # | SPEC ID |
|---|---------|
| 1 | SPEC-I18N-001-ARCHIVED |

### Pattern G — `archived → in-progress` (1 item)

| # | SPEC ID |
|---|---------|
| 1 | SPEC-V3R3-WEB-001 |

### Pattern H — Self-drift (cleanup chain, 3 items)

| # | SPEC ID | drift |
|---|---------|-------|
| 1 | SPEC-V3R4-LINT-SKIP-CLEANUP-001 | implemented → in-progress |
| 2 | SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 | completed → in-progress |
| 3 | SPEC-V3R4-SPECLINT-DEBT-001 | completed → planned |

**Note**: 본 표는 plan-phase 기준 추정. Run-phase BASELINE milestone에서 `moai spec lint --strict` 실시간 출력으로 최종 affected-list 확정 (총 64 ± 미세 변동 가능 — 새 commit / SPEC 추가 영향).
