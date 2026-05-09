---
id: SPEC-V3R2-SPC-003
document: acceptance
version: "0.1.0"
status: backfilled
created: 2026-05-10
updated: 2026-05-10
author: manager-spec (Batch 3 backfill)
related_spec: SPEC-V3R2-SPC-003
phase: plan
language: ko
total_acceptance_criteria: 16
---

# SPEC-V3R2-SPC-003 — Acceptance Criteria (수용 기준, BACKFILL)

> **BACKFILL NOTICE**: 본 acceptance.md는 spec.md §6의 16개 flat AC를 SPC-001 hierarchical schema에 맞춰 Given-When-Then 형식으로 재작성한 retroactive 문서이다. 각 AC는 `internal/spec/lint_test.go`의 `TestLinter_AC{NN}_*` 테스트 함수와 1:1 대응한다. 16/16 자동 검증 가능.

---

## REQ ↔ AC Traceability Matrix

| REQ ID | AC ID(s) | Test Function |
|--------|----------|---------------|
| REQ-SPC-003-001 (Ubiquitous: zero-or-more args) | AC-V3R2-SPC-003-01 | `TestLinter_AC01_HappyPath` |
| REQ-SPC-003-002 (Ubiquitous: exit 0 iff no errors) | AC-V3R2-SPC-003-01 | `TestLinter_AC01_HappyPath` |
| REQ-SPC-003-003 (Ubiquitous: EARS modality forms) | AC-V3R2-SPC-003-03 | `TestLinter_AC03_ModalityMalformed` |
| REQ-SPC-003-004 (Ubiquitous: REQ ID regex + uniqueness) | AC-V3R2-SPC-003-05 | `TestLinter_AC05_DuplicateREQID` |
| REQ-SPC-003-005 (Ubiquitous: AC→REQ coverage ≥100%) | AC-V3R2-SPC-003-01, AC-V3R2-SPC-003-02, AC-V3R2-SPC-003-16 | `TestLinter_AC01/02/16_*` |
| REQ-SPC-003-006 (Ubiquitous: frontmatter schema) | AC-V3R2-SPC-003-15 (parse failure subset) | `TestLinter_AC15_ParseFailure` |
| REQ-SPC-003-007 (Ubiquitous: dependency exists) | AC-V3R2-SPC-003-07 | `TestLinter_AC07_MissingDependency` |
| REQ-SPC-003-008 (Ubiquitous: DAG cycle) | AC-V3R2-SPC-003-04 | `TestLinter_AC04_DependencyCycle` |
| REQ-SPC-003-009 (Ubiquitous: Out-of-Scope present) | AC-V3R2-SPC-003-06 | `TestLinter_AC06_MissingExclusions` |
| REQ-SPC-003-010 (Ubiquitous: zone registry xref) | AC-V3R2-SPC-003-08 | `TestLinter_AC08_DanglingRuleReference` |
| REQ-SPC-003-020 (Event-driven: --json) | AC-V3R2-SPC-003-09 | `TestLinter_AC09_JSONOutput` |
| REQ-SPC-003-021 (Event-driven: --sarif) | AC-V3R2-SPC-003-10 | `TestLinter_AC10_SARIFOutput` |
| REQ-SPC-003-022 (Event-driven: ParseFailure continue) | AC-V3R2-SPC-003-15 | `TestLinter_AC15_ParseFailure` |
| REQ-SPC-003-030 (State-driven: --strict promote warn→err) | AC-V3R2-SPC-003-11 | `TestLinter_AC11_StrictMode` |
| REQ-SPC-003-031 (State-driven: DuplicateSPECID) | AC-V3R2-SPC-003-12 | `TestLinter_AC12_DuplicateSPECID` |
| REQ-SPC-003-040 (Optional: lint.skip) | AC-V3R2-SPC-003-13 | `TestLinter_AC13_LintSkip` |
| REQ-SPC-003-050 (Complex: WHEN-without-SHALL) | AC-V3R2-SPC-003-03 | `TestLinter_AC03_ModalityMalformed` |
| REQ-SPC-003-052 (Complex: breaking + bc_id consistency) | AC-V3R2-SPC-003-14 | `TestLinter_AC14_BreakingChangeMissingID` |

REQ coverage: 16/16 REQ → 16 AC, 모든 REQ가 적어도 하나의 AC에 매핑됨.

---

## AC-V3R2-SPC-003-01 — Happy path: 0 findings, exit 0

**REQ 매핑**: REQ-SPC-003-001 (Ubiquitous), REQ-SPC-003-002 (Ubiquitous), REQ-SPC-003-005 (Ubiquitous)

### Given

- 12개 REQ + 12개 AC가 1:1 매핑된 valid SPEC 1개가 fixture에 존재한다.
- frontmatter 14필드 모두 정상, EARS modality 모두 정상, dependency cycle 없음, Out-of-Scope 섹션 존재.

### When

- `Linter.Lint([]string{fixturePath})`를 호출한다.

### Then

- `Report.Findings`가 비어 있다 (`len(findings) == 0`).
- `Report.HasErrors()` returns `false`.
- exit code 0에 해당하는 상태가 반환된다.

### 검증 방식

- **자동**: `go test -run TestLinter_AC01_HappyPath ./internal/spec/...`
- 테스트 위치: `internal/spec/lint_test.go:55`

---

## AC-V3R2-SPC-003-02 — CoverageIncomplete: REQ에 매핑된 AC 없음

**REQ 매핑**: REQ-SPC-003-005 (Ubiquitous)

### Given

- `REQ-X-001-007`이 §5에 선언되어 있으나 §6의 어떤 AC tail도 `(maps REQ-X-001-007)`을 포함하지 않는 fixture SPEC.

### When

- `Linter.Lint(...)`를 실행한다.

### Then

- `Report.Findings`에 `Code == "CoverageIncomplete"` finding이 1개 이상 존재한다.
- finding의 `Message`가 `"REQ-X-001-007"`을 포함한다.
- `Severity == "error"`.

### 검증 방식

- **자동**: `go test -run TestLinter_AC02_CoverageIncomplete ./internal/spec/...`
- 테스트 위치: `internal/spec/lint_test.go:75`

---

## AC-V3R2-SPC-003-03 — ModalityMalformed: WHEN에 SHALL 누락

**REQ 매핑**: REQ-SPC-003-003 (Ubiquitous), REQ-SPC-003-050 (Complex)

### Given

- 요구사항 텍스트가 `"WHEN the user logs in, the system creates a session"`인 fixture SPEC (WHEN trigger 있으나 SHALL 응답절 없음).

### When

- `Linter.Lint(...)`를 실행한다.

### Then

- `Report.Findings`에 `Code == "ModalityMalformed"` finding이 존재한다.
- `Severity == "error"`.

### 검증 방식

- **자동**: `go test -run TestLinter_AC03_ModalityMalformed ./internal/spec/...`
- 테스트 위치: `internal/spec/lint_test.go:107`

---

## AC-V3R2-SPC-003-04 — DependencyCycle: A→B→A

**REQ 매핑**: REQ-SPC-003-008 (Ubiquitous)

### Given

- SPEC A의 `dependencies: [B]`, SPEC B의 `dependencies: [A]`인 두 fixture SPEC.

### When

- `Linter.Lint([A, B])`를 실행한다 (cross-SPEC mode).

### Then

- `Report.Findings`에 `Code == "DependencyCycle"` finding이 존재한다.
- finding의 `Message`가 두 SPEC ID를 모두 언급한다.
- `Severity == "error"`.

### 검증 방식

- **자동**: `go test -run TestLinter_AC04_DependencyCycle ./internal/spec/...`
- 테스트 위치: `internal/spec/lint_test.go:126`

---

## AC-V3R2-SPC-003-05 — DuplicateREQID: 단일 SPEC 내 중복

**REQ 매핑**: REQ-SPC-003-004 (Ubiquitous)

### Given

- `REQ-X-001-005`가 §5 내에서 두 번 선언된 fixture SPEC.

### When

- `Linter.Lint(...)`를 실행한다.

### Then

- `Report.Findings`에 `Code == "DuplicateREQID"` finding이 존재한다.
- `Severity == "error"`.

### 검증 방식

- **자동**: `go test -run TestLinter_AC05_DuplicateREQID ./internal/spec/...`
- 테스트 위치: `internal/spec/lint_test.go:149`

---

## AC-V3R2-SPC-003-06 — MissingExclusions: Out-of-Scope 섹션 부재

**REQ 매핑**: REQ-SPC-003-009 (Ubiquitous)

### Given

- §2.2 Out-of-Scope 서브섹션이 누락된 fixture SPEC.

### When

- `Linter.Lint(...)`를 실행한다.

### Then

- `Report.Findings`에 `Code == "MissingExclusions"` finding이 존재한다.
- `Severity == "error"`.

### 검증 방식

- **자동**: `go test -run TestLinter_AC06_MissingExclusions ./internal/spec/...`
- 테스트 위치: `internal/spec/lint_test.go:168`

---

## AC-V3R2-SPC-003-07 — MissingDependency: 존재하지 않는 SPEC 참조

**REQ 매핑**: REQ-SPC-003-007 (Ubiquitous)

### Given

- `dependencies: [SPEC-NONEXISTENT-001]`인 fixture SPEC. `SPEC-NONEXISTENT-001` 디렉터리는 디스크에 부재.

### When

- `Linter.Lint(...)`를 실행한다.

### Then

- `Report.Findings`에 `Code == "MissingDependency"` finding이 존재한다.
- finding의 `Message`가 `"SPEC-NONEXISTENT-001"`을 포함한다.

### 검증 방식

- **자동**: `go test -run TestLinter_AC07_MissingDependency ./internal/spec/...`
- 테스트 위치: `internal/spec/lint_test.go:187`

---

## AC-V3R2-SPC-003-08 — DanglingRuleReference: zone registry 부재 ID

**REQ 매핑**: REQ-SPC-003-010 (Ubiquitous)

### Given

- `related_rule: [CONST-V3R2-999]`인 fixture SPEC. `CONST-V3R2-999`는 `.claude/rules/moai/core/zone-registry.md`에 없는 ID.

### When

- `Linter.Lint(...)`를 실행한다 (zone registry 적재 상태).

### Then

- `Report.Findings`에 `Code == "DanglingRuleReference"` finding이 존재한다.
- `Severity == "warning"` (NOT error — zone registry 미적재 시 graceful degradation).

### 검증 방식

- **자동**: `go test -run TestLinter_AC08_DanglingRuleReference ./internal/spec/...`
- 테스트 위치: `internal/spec/lint_test.go:218`

---

## AC-V3R2-SPC-003-09 — JSON output 포맷

**REQ 매핑**: REQ-SPC-003-020 (Event-driven)

### Given

- 임의의 fixture SPEC + `LinterOptions{OutputFormat: "json"}`.

### When

- `Linter.Lint(...)` 후 `Report.ToJSON()`을 호출한다.

### Then

- 반환된 byte slice가 valid JSON array이다 (`json.Unmarshal` 성공).
- 각 finding 객체가 `file`, `line`, `severity`, `code`, `message` 필드를 포함한다.

### 검증 방식

- **자동**: `go test -run TestLinter_AC09_JSONOutput ./internal/spec/...`
- 테스트 위치: `internal/spec/lint_test.go:245`

---

## AC-V3R2-SPC-003-10 — SARIF 2.1.0 output 포맷

**REQ 매핑**: REQ-SPC-003-021 (Event-driven)

### Given

- 임의의 fixture SPEC + `LinterOptions{OutputFormat: "sarif"}`.

### When

- `Linter.Lint(...)` 후 `Report.ToSARIF()`를 호출한다.

### Then

- 반환된 byte slice가 valid JSON이며 SARIF 2.1.0 schema에 부합한다.
- 최상위 객체가 `version: "2.1.0"`을 갖는다.
- `runs[0].tool.driver.name == "moai spec lint"`.
- 각 finding이 `runs[0].results[]`에 매핑되며 `ruleId`, `level`, `message.text`, `locations[]` 필드를 갖는다.

### 검증 방식

- **자동**: `go test -run TestLinter_AC10_SARIFOutput ./internal/spec/...`
- 테스트 위치: `internal/spec/lint_test.go:283`

---

## AC-V3R2-SPC-003-11 — Strict mode: warning → error 승격

**REQ 매핑**: REQ-SPC-003-030 (State-driven)

### Given

- 0 errors + 2 warnings를 산출하는 fixture SPEC + `LinterOptions{Strict: true}`.

### When

- `Linter.Lint(...)`를 실행한다.

### Then

- `Report.HasErrors()` returns `true` (warnings가 errors로 promote됨).
- exit code 1에 해당하는 상태가 반환된다.

### 검증 방식

- **자동**: `go test -run TestLinter_AC11_StrictMode ./internal/spec/...`
- 테스트 위치: `internal/spec/lint_test.go:324`

---

## AC-V3R2-SPC-003-12 — DuplicateSPECID: 두 SPEC이 동일 id

**REQ 매핑**: REQ-SPC-003-031 (State-driven)

### Given

- 두 fixture SPEC 디렉터리가 모두 frontmatter `id: SPEC-V3R2-X-001`을 선언.

### When

- `Linter.Lint([dir1, dir2])` cross-SPEC 모드를 실행한다.

### Then

- `Report.Findings`에 `Code == "DuplicateSPECID"` finding이 존재한다.
- finding의 `Message`가 두 file path를 모두 언급한다.
- `Severity == "error"`.

### 검증 방식

- **자동**: `go test -run TestLinter_AC12_DuplicateSPECID ./internal/spec/...`
- 테스트 위치: `internal/spec/lint_test.go:346`

---

## AC-V3R2-SPC-003-13 — lint.skip: per-SPEC 코드 suppress

**REQ 매핑**: REQ-SPC-003-040 (Optional)

### Given

- frontmatter `lint.skip: [DanglingRuleReference]`를 가진 fixture SPEC + 본래 dangling rule reference를 포함.

### When

- `Linter.Lint(...)`를 실행한다.

### Then

- `Report.Findings`에 `Code == "DanglingRuleReference"` finding이 **존재하지 않는다**.
- 다른 종류 finding은 정상적으로 reporting된다 (예: 다른 SPEC의 동일 코드는 영향 없음).

### 검증 방식

- **자동**: `go test -run TestLinter_AC13_LintSkip ./internal/spec/...`
- 테스트 위치: `internal/spec/lint_test.go:367`

---

## AC-V3R2-SPC-003-14 — BreakingChangeMissingID: breaking + bc_id 일관성

**REQ 매핑**: REQ-SPC-003-052 (Complex)

### Given

- frontmatter `breaking: true` + `bc_id: []`인 fixture SPEC.

### When

- `Linter.Lint(...)`를 실행한다.

### Then

- `Report.Findings`에 `Code == "BreakingChangeMissingID"` finding이 존재한다.
- `Severity == "error"`.

### 검증 방식

- **자동**: `go test -run TestLinter_AC14_BreakingChangeMissingID ./internal/spec/...`
- 테스트 위치: `internal/spec/lint_test.go:389`

---

## AC-V3R2-SPC-003-15 — ParseFailure: malformed YAML 후 continue

**REQ 매핑**: REQ-SPC-003-022 (Event-driven), REQ-SPC-003-006 (Ubiquitous: frontmatter schema)

### Given

- 두 fixture SPEC 디렉터리 — 첫 번째는 broken YAML frontmatter, 두 번째는 valid SPEC.

### When

- `Linter.Lint([broken, valid])`를 실행한다.

### Then

- `Report.Findings`에 `Code == "ParseFailure"` finding이 broken file에 대해 존재한다.
- valid file에 대한 정상 lint가 계속 수행된다 (linter는 첫 번째 실패 후 abort하지 않음).
- 전체 exit code는 1 (errors > 0).

### 검증 방식

- **자동**: `go test -run TestLinter_AC15_ParseFailure ./internal/spec/...`
- 테스트 위치: `internal/spec/lint_test.go:408`

---

## AC-V3R2-SPC-003-16 — Hierarchical AC coverage 계산

**REQ 매핑**: REQ-SPC-003-005 (Ubiquitous)

### Given

- SPC-001 hierarchical AC 트리를 가진 fixture SPEC: `AC-X-01`이 3개 자식(`AC-X-01-a`, `AC-X-01-b`, `AC-X-01-c`)을 갖고, 그 중 적어도 하나의 leaf가 `(maps REQ-X-001-001)` tail을 갖는다.

### When

- `Linter.Lint(...)`를 실행한다.

### Then

- `REQ-X-001-001`은 coverage 계산에서 *covered*로 간주된다.
- `Report.Findings`에 해당 REQ에 대한 `CoverageIncomplete` finding이 존재하지 **않는다**.
- 즉, 하나의 leaf가 매핑하면 parent AC 전체가 그 REQ를 covering한 것으로 본다.

### 검증 방식

- **자동**: `go test -run TestLinter_AC16_HierarchicalACCoverage ./internal/spec/...`
- 테스트 위치: `internal/spec/lint_test.go:448`

---

## Edge Cases (검증 완료 항목)

다음 edge case들은 위 16개 AC에 묵시적으로 포함되어 있으며 추가 별도 AC는 발행하지 않는다 (FROZEN scope 준수):

1. **Self-loop dependency** (A → A): AC-04의 일반화 케이스. `DependencyCycle` finding 발화.
2. **Multi-cycle 동시 검출** (A→B→A and C→D→C): Tarjan SCC 특성상 모두 검출됨. `internal/spec/dag.go` 구현 검증.
3. **빈 `dependencies: []`**: AC-01 happy path의 일부. finding 없음.
4. **Frontmatter 일부 필드 결손** (e.g., `tags` 누락): AC-15 ParseFailure의 일반화 케이스 또는 `FrontmatterInvalid` finding.
5. **EARS Ubiquitous form ("The system shall ...")**: AC-01 happy path에서 검증됨. `ModalityMalformed` 발화 안 함.

---

## Definition of Done (DoD)

- [x] 16/16 ACs 자동 테스트 PASS (`go test ./internal/spec/...`)
- [x] `internal/spec` 패키지 coverage ≥ 85% (실측 86.6%)
- [x] `go vet ./...` 0 issues
- [x] `golangci-lint run ./...` 0 issues
- [x] `go test -race ./...` 데이터 레이스 없음
- [x] PR #745 main 머지 (commit `03146d1ae`)

본 AC 16/16은 PR #745 머지 전 시점에 모두 통과되었음 (progress.md 기록).

---

End of acceptance.md (backfill v0.1.0).
