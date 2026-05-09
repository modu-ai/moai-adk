---
id: SPEC-V3R2-SPC-003
document: plan
version: "0.1.0"
status: backfilled
created: 2026-05-10
updated: 2026-05-10
author: manager-spec (Batch 3 backfill)
related_spec: SPEC-V3R2-SPC-003
phase: plan
language: ko
---

# SPEC-V3R2-SPC-003 — Implementation Plan (구현 계획, BACKFILL)

> **BACKFILL NOTICE**: 본 plan.md는 PR #745 (Wave 5)로 이미 머지된 `moai spec lint` 구현의 milestone 구조를 사후 정리한 retroactive 문서이다. 코드는 FROZEN 상태이며 본 문서는 **실제 발생한 commit 흐름**을 M1-M4 milestone에 재구성한다. 새 요구사항은 추가하지 않으며, 이미 구현된 결과물의 traceability 보존이 목적이다.

---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-10 | manager-spec (Batch 3 backfill) | Retroactive plan reflecting as-merged Wave 5 work (PR #745) |

---

## 1. 전체 전략 (Strategy, retroactive)

본 SPEC은 **새 CLI subcommand 추가 + Go 패키지 신설** 작업이다. 기존 `internal/spec/` 패키지에 lint 엔진을 추가하고 `internal/cli/`에 cobra 서브커맨드를 등록했다. 핵심 원칙:

1. **TDD 순서**: RED 커밋(`44019ec4d`)에서 16개 AC 테스트를 먼저 작성 → GREEN 커밋(`4d5699f27`)에서 모든 테스트가 통과하도록 구현 → REFACTOR 커밋(`1587e6e3a`)에서 Karpathy simplicity 원칙 + MX 태그 적용.
2. **Rule interface 분리**: per-SPEC rule (8개) vs cross-SPEC rule (2개)를 별도 Go interface로 분리하여 향후 병렬화 여지 확보.
3. **외부 의존성 zero**: spec.md §7 제약 준수. `gopkg.in/yaml.v3` + `go-playground/validator/v10` 외 신규 라이브러리 없음.
4. **출력 3-tier**: table (default human-readable) → JSON (machine) → SARIF 2.1.0 (CI annotation).
5. **lint.skip 탈출구**: false-positive 위험을 frontmatter-scoped 메커니즘으로 완화 (line-scoped 주석은 가독성 저해).

---

## 2. 영향 범위 (Files Touched, 실측)

| 종류 | 파일 | LOC | 변경 형태 |
|------|------|-----|-----------|
| 신설 | `internal/spec/lint.go` | 816 | Linter struct, Rule interface, 10 rules |
| 신설 | `internal/spec/lint_test.go` | 572 | 16 AC 테이블 테스트 |
| 신설 | `internal/spec/dag.go` | 100 | Tarjan SCC iterative cycle detection |
| 신설 | `internal/spec/sarif.go` | 165 | SARIF 2.1.0 JSON writer |
| 신설 | `internal/cli/spec_lint.go` | 164 | cobra `newSpecLintCmd()` |
| 수정 | `internal/cli/spec.go` | +N | `newSpecLintCmd()` 등록 (one-line addition) |
| 신설 | `internal/spec/testdata/` | — | 11개 fixture SPEC 디렉터리 |

총 신설 sources: 1245 LOC. 실측은 `wc -l` 결과(2026-05-10 기준).

---

## 3. 마일스톤 (Milestones, retroactive)

총 **4개 마일스톤** (M1–M4). 각 마일스톤은 단일 commit에 매핑된다 (squash merge 후 단일 PR로 통합).

### M1 — RED phase (failing tests 작성)

**Priority**: P0 (TDD 사전 조건)
**Commit**: `44019ec4d test(spec): SPEC-V3R2-SPC-003 RED phase — failing lint tests`

산출물:
- `internal/spec/lint_test.go` 신설 — 16개 `TestLinter_AC{NN}_*` 테이블 테스트
- `internal/spec/testdata/` 11개 fixture SPEC 디렉터리 (각 AC별 시나리오)

검증:
- `go test ./internal/spec/...` 실행 시 모든 테스트 fail (RED 상태) — 미정의 심볼 (`Linter`, `LinterOptions`, `NewLinter`, `Report`, `Finding`)

이 단계에서 lint.go / dag.go / sarif.go / spec_lint.go는 아직 존재하지 않는다.

### M2 — GREEN phase (linter engine 구현)

**Priority**: P0
**Commit**: `4d5699f27 feat(spec): SPEC-V3R2-SPC-003 GREEN — moai spec lint linter engine`

산출물:
- `internal/spec/lint.go` 신설 — `Linter`, `Rule` interface, 10개 rule struct, `parseSPECDoc()`, `extractFrontmatter()`, `parseREQs()`
- `internal/spec/dag.go` 신설 — Tarjan SCC iterative
- `internal/spec/sarif.go` 신설 — `Report.ToSARIF()` writer
- `internal/cli/spec_lint.go` 신설 — `newSpecLintCmd()` cobra subcommand
- `internal/cli/spec.go` 수정 — `newSpecLintCmd()` 등록 한 줄

검증:
- `go test ./internal/spec/...` → 16/16 PASS
- `go test ./internal/cli/...` → 모두 PASS (회귀 없음)

### M3 — REFACTOR phase (Karpathy simplicity + MX 태그)

**Priority**: P1 (품질 게이트)
**Commit**: `1587e6e3a refactor(spec): SPEC-V3R2-SPC-003 REFACTOR — Karpathy simplicity + MX tags`

산출물:
- `lint.go` `collectLeafREQIDs()` 함수 제거 (사용처 없음)
- `errcheck` 위반 7건 수정 (`fmt.Fprintf` / `fmt.Fprintln` 반환값 처리)
- `sarif.go` `max()` → `positiveLineNum()` 이름 변경 (Go 1.21+ 내장 `max()` 충돌 회피)
- @MX 태그 주입: `@MX:NOTE` (의도 설명), `@MX:ANCHOR` (high-fan-in 함수)
- `progress.md` 갱신 (16 ACs / 86.6% coverage)

검증:
- `go vet ./...` → 0 issues
- `golangci-lint run ./...` → 0 issues
- `go test -race ./...` → 모든 테스트 PASS, 데이터 레이스 없음

### M4 — 머지 + Wave 5 squash

**Priority**: P0 (배포)
**Commit**: `03146d1ae feat(spec): SPEC-V3R2-SPC-003 — moai spec lint CLI (Wave 5) (#745)`

산출물:
- PR #745 squash merge → main
- M1+M2+M3의 3개 커밋이 단일 commit으로 압축됨
- main 진입 시점: 2026-04-30 17:51 KST

검증:
- main CI all-green (Lint / Test ubuntu/macos/windows / Build 5 / CodeQL)
- post-merge `moai spec lint .moai/specs/SPEC-V3R2-SPC-003/spec.md` 실행 가능

---

## 4. Traceability Matrix (REQ × AC × Test × Commit)

> 16개 AC는 모두 `internal/spec/lint_test.go`의 `TestLinter_AC{NN}_*` 함수와 1:1 대응한다. 각 테스트는 M1(RED)에서 작성되고 M2(GREEN)에서 통과되었다.

| AC | REQ 매핑 | Test 함수 | RED commit | GREEN 통과 |
|----|----------|----------|-----------|-----------|
| AC-V3R2-SPC-003-01 | REQ-SPC-003-001/002/005 | `TestLinter_AC01_HappyPath` | `44019ec4d` | `4d5699f27` |
| AC-V3R2-SPC-003-02 | REQ-SPC-003-005 | `TestLinter_AC02_CoverageIncomplete` | `44019ec4d` | `4d5699f27` |
| AC-V3R2-SPC-003-03 | REQ-SPC-003-003/050 | `TestLinter_AC03_ModalityMalformed` | `44019ec4d` | `4d5699f27` |
| AC-V3R2-SPC-003-04 | REQ-SPC-003-008 | `TestLinter_AC04_DependencyCycle` | `44019ec4d` | `4d5699f27` |
| AC-V3R2-SPC-003-05 | REQ-SPC-003-004 | `TestLinter_AC05_DuplicateREQID` | `44019ec4d` | `4d5699f27` |
| AC-V3R2-SPC-003-06 | REQ-SPC-003-009 | `TestLinter_AC06_MissingExclusions` | `44019ec4d` | `4d5699f27` |
| AC-V3R2-SPC-003-07 | REQ-SPC-003-007 | `TestLinter_AC07_MissingDependency` | `44019ec4d` | `4d5699f27` |
| AC-V3R2-SPC-003-08 | REQ-SPC-003-010 | `TestLinter_AC08_DanglingRuleReference` | `44019ec4d` | `4d5699f27` |
| AC-V3R2-SPC-003-09 | REQ-SPC-003-020 | `TestLinter_AC09_JSONOutput` | `44019ec4d` | `4d5699f27` |
| AC-V3R2-SPC-003-10 | REQ-SPC-003-021 | `TestLinter_AC10_SARIFOutput` | `44019ec4d` | `4d5699f27` |
| AC-V3R2-SPC-003-11 | REQ-SPC-003-030 | `TestLinter_AC11_StrictMode` | `44019ec4d` | `4d5699f27` |
| AC-V3R2-SPC-003-12 | REQ-SPC-003-031 | `TestLinter_AC12_DuplicateSPECID` | `44019ec4d` | `4d5699f27` |
| AC-V3R2-SPC-003-13 | REQ-SPC-003-040 | `TestLinter_AC13_LintSkip` | `44019ec4d` | `4d5699f27` |
| AC-V3R2-SPC-003-14 | REQ-SPC-003-052 | `TestLinter_AC14_BreakingChangeMissingID` | `44019ec4d` | `4d5699f27` |
| AC-V3R2-SPC-003-15 | REQ-SPC-003-022 | `TestLinter_AC15_ParseFailure` | `44019ec4d` | `4d5699f27` |
| AC-V3R2-SPC-003-16 | REQ-SPC-003-005 | `TestLinter_AC16_HierarchicalACCoverage` | `44019ec4d` | `4d5699f27` |

REQ coverage: 16/16 = 100%. 모든 spec.md §5의 REQ가 적어도 하나의 AC에 매핑된다.

---

## 5. 기술적 접근 (Technical Approach)

### 5.1 Go 패키지 레이아웃

```
internal/spec/
├── lint.go         — Linter, Rule interface, 10 rules, parseSPECDoc()
├── lint_test.go    — 16 TestLinter_AC* 테이블 테스트
├── dag.go          — Tarjan SCC (cycle detection)
├── sarif.go        — SARIF 2.1.0 writer
└── testdata/       — 11 fixture SPEC 디렉터리

internal/cli/
└── spec_lint.go    — cobra subcommand (newSpecLintCmd)
```

### 5.2 핵심 자료구조

```go
type Severity string  // "error" / "warning" / "info"
type Finding struct {
    File     string
    Line     int
    Severity Severity
    Code     string   // "CoverageIncomplete", "ModalityMalformed", ...
    Message  string
}
type Report struct {
    Findings []Finding
}
type Rule interface {
    Code() string
    Check(doc *SPECDoc, allDocs []*SPECDoc) []Finding
}
type crossSPECRule interface {
    Code() string
    CheckAll(allDocs []*SPECDoc) []Finding
}
```

### 5.3 Lint 흐름 (단일 진입점 `Linter.Lint(paths []string)`)

```
1. discoverSPECs(paths) → 모든 spec.md 경로 수집
2. for each path: parseSPECDoc() → SPECDoc{Frontmatter, REQs, ACs, Body}
3. for each doc: 8개 per-SPEC Rule.Check() 실행 → []Finding
4. 모든 docs 수집 후: 2개 cross-SPEC Rule.CheckAll() 실행 → []Finding
5. applylintSkip() → frontmatter `lint.skip` 코드 필터링
6. severity 정렬 → Report 반환
7. 출력 모드(table/json/sarif)에 따라 stdout 작성
8. exit code: errors > 0 → 1; --strict + warnings > 0 → 1; else → 0
```

---

## 6. 위험 및 완화 (Risks & Mitigations, retrospective)

| 위험 (spec.md §8) | 실현 여부 | 실제 대응 |
|--------------------|----------|----------|
| EARS modality regex false-positive | 부분 실현 (자연어 변형 일부) | `lint.skip: [ModalityMalformed]` 탈출구 + 보수적 패턴 유지 |
| Cycle detection 성능 저하 | 미실현 | Tarjan O(V+E) — 11 fixture 기준 < 1ms |
| Duplicate SPEC IDs | 미실현 (테스트에서 의도적 reproduction만) | `DuplicateSPECID` error로 surface |
| SARIF output 포맷 churn | 미실현 | 2.1.0 pin |
| Frontmatter schema drift | 미실현 | 단일 `SPECFrontmatter` struct + validator/v10 tag |

---

## 7. 외부 의존성

- `gopkg.in/yaml.v3` (이미 moai 모듈 그래프에 존재)
- `go-playground/validator/v10` (이미 moai 모듈 그래프에 존재)
- Go stdlib (`regexp`, `strings`, `path/filepath`, `encoding/json`, `os`, `io`, `fmt`)

신규 의존성 추가: 없음.

---

## 8. Wave 5 컨텍스트

본 SPEC은 Wave 5 batch에 포함되어 다른 SPEC들과 함께 일괄 머지되었다. Wave 5 batch:
- SPEC-V3R2-SPC-003 (본 SPEC, PR #745)
- 기타 Wave 5 SPECs (memory `project_wave5_complete.md` 참조)

머지 전략: PR #745는 squash merge되어 main에 단일 commit `03146d1ae`로 진입.

---

End of plan.md (backfill v0.1.0).
